package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"
	"github.com/yalp/jsonpath"
	"gopkg.in/yaml.v2"
)

// Version represents the version of Kubernetes
// for which we should load the schema
var Version string

// OpenShift represents whether to test against
// upstream Kubernetes of the OpenShift schemas
var OpenShift bool

// ValidFormat is a type for quickly forcing
// new formats on the gojsonschema loader
type ValidFormat struct{}

// IsFormat always retusn true and meets the
// gojsonschema.FormatChecker interface
func (f ValidFormat) IsFormat(input string) bool {
	return true
}

func info(message ...interface{}) {
	fmt.Println(message...)
}

func success(message ...interface{}) {
	green := color.New(color.FgGreen)
	green.Println(message...)
}

func warn(message ...interface{}) {
	yellow := color.New(color.FgYellow)
	yellow.Println(message...)
}

func error(message ...interface{}) {
	red := color.New(color.FgRed)
	red.Println(message...)
}

func stringInSlice(a string, list []interface{}) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func isJsonNumber(what interface{}) bool {
	switch what.(type) {
	case json.Number:
		return true
	}
	return false
}

// Based on https://stackoverflow.com/questions/40737122/convert-yaml-to-json-without-struct-golang
// We unmarshal yaml into a value of type interface{},
// go through the result recursively, and convert each encountered
// map[interface{}]interface{} to a map[string]interface{} value
// required to marshall to JSON.
func convertToStringKeys(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convertToStringKeys(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convertToStringKeys(v)
		}
	}
	return i
}

// RootCmd represents the the command to run when kubeval is run
var RootCmd = &cobra.Command{
	Use:   "kubeval <file> [file...]",
	Short: "Validate a Kubernetes YAML file against the relevant schema",
	Long:  `Validate a Kubernetes YAML file against the relevant schema`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			error("You must pass at least one file as an argument")
			os.Exit(1)
		}
		success := true
		for _, file := range args {
			valid := validate(file)
			if success {
				success = valid
			}
		}
		if !success {
			os.Exit(1)
		}
	},
}

// Validate a Kubernetes YAML file according to a relevant schema
// TODO This function requires a judicious amount of refactoring.
func validate(element string) bool {
	// Open the YAML file, convert to a Go interface and then
	// load that as a document for gojsonschema to process
	filename, _ := filepath.Abs(element)
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		error("Could not open file", filename)
		return false
	}

	var spec interface{}
	err = yaml.Unmarshal(yamlFile, &spec)
	if err != nil {
		error("Failed to decode YAML from", filename)
		return false
	}

	body := convertToStringKeys(spec)

	documentLoader := gojsonschema.NewGoLoader(body)

	cast, _ := body.(map[string]interface{})
	if _, ok := cast["kind"]; !ok {
		error("Missing a kind key in", filename)
		return false
	}

	kind := cast["kind"].(string)

	// We have both the upstream Kubernetes schemas and the OpenShift schemas available
	// the tool can toggle between then using the --openshift boolean flag and here we
	// use that to select which repository to get the schema from
	var schemaType string
	if OpenShift {
		schemaType = "openshift"
	} else {
		schemaType = "kubernetes"
	}

	// Most of the directories which store the schemas are prefixed with a v so as to
	// match the tagging in the Kubernetes repository, apart from master.
	normalisedVersion := Version
	if Version != "master" {
		normalisedVersion = "v" + normalisedVersion
	}

	schema := fmt.Sprintf("https://raw.githubusercontent.com/garethr/%s-json-schema/master/%s-standalone/%s.json", schemaType, normalisedVersion, strings.ToLower(kind))

	schemaLoader := gojsonschema.NewReferenceLoader(schema)

	// Without forcing these types the schema fails to load
	// Need to Work out proper handling for these types
	gojsonschema.FormatCheckers.Add("int64", ValidFormat{})
	gojsonschema.FormatCheckers.Add("byte", ValidFormat{})
	gojsonschema.FormatCheckers.Add("int32", ValidFormat{})
	gojsonschema.FormatCheckers.Add("int-or-string", ValidFormat{})

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		error("Problem loading schema from the network")
		info(err)
		return false
	}

	if result.Valid() {
		success("The document", element, "is a valid", kind)
		return true
	}

	// The Kubernetes JSON schemas usage isn't valid in my view
	// Integers are passed to string fields but with a format hint saying
	// "don't worry, it's not really a string". This goes by the name of
	// int-or-string. I don't have a complete solution to resolving this
	// here, and will be reporting upstream when I get a moment to write up
	// properly. But rather than mark valid files as invalid here's a work around

	// Grab all the formats used in the current schema and see if
	// int-or-string is present. If it's not we know for definitely a
	// number passed to a string field is invalid
	schemaJson, _ := schemaLoader.LoadJSON()
	formats, _ := jsonpath.Read(schemaJson, "$..format")
	schemaContainsIntOrString := stringInSlice("int-or-string", formats.([]interface{}))

	// Some errors might be caused by int-or-string
	// so we need to track whether we have one or more non int-or-string candidates
	noRealErrors := true
	for _, desc := range result.Errors() {
		isNumber := isJsonNumber(desc.Value())
		// Here is the heuristic we're using to say something is likely a valid
		// Kubernetes config even though it's not valid according to the schema
		// is it a number?
		// does the schema contain any int-or-string formated fields?
		// is the validation error invalid type
		// This will raise some false negatives unfortunately so we
		// print a warning rather than fail completely
		if isNumber && schemaContainsIntOrString && desc.Type() == "invalid_type" {
			warn("--->", desc.Field(), "is an integer, but might be an int_or_string property")
		} else {
			info("-->", desc)
			if noRealErrors {
				noRealErrors = false
			}
		}
	}
	if noRealErrors {
		success("The document", element, "is a valid", kind)
		return true
	} else {
		warn("The document", element, "is not a valid", kind)
		return false
	}
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		error(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.Flags().StringVarP(&Version, "kubernetes-version", "v", "master", "Version of Kubernetes to validate against")
	RootCmd.Flags().BoolVarP(&OpenShift, "openshift", "", false, "Use OpenShift schemas instead of upstream Kubernetes")
}
