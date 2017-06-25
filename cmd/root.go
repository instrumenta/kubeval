package cmd

import (
	"fmt"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var Version string
var OpenShift bool

type ValidFormat struct{}

// Ensure it meets the gojsonschema.FormatChecker interface
func (f ValidFormat) IsFormat(input string) bool {
	return true
}

// Based on https://stackoverflow.com/questions/40737122/convert-yaml-to-json-without-struct-golang
// We unmarshal yaml into a value of type interface{},
// go through the result recursively, and convert each encountered
// map[interface{}]interface{} to a map[string]interface{} value
// required to marshall to JSON.
func convert_to_string_keys(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convert_to_string_keys(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert_to_string_keys(v)
		}
	}
	return i
}

var cfgFile string

var RootCmd = &cobra.Command{
	Use:   "kubeval",
	Short: "Validate a Kubernetes YAML file against the relevant schema",
	Long:  `Validate a Kubernetes YAML file against the relevant schema`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("You must pass at least one file as an argument")
			os.Exit(1)
		}
		for _, file := range args {
			validate(file)
		}
	},
}

// Validate a Kubernetes YAML file accoring to a relevant schema
func validate(element string) {
	// Open the YAML file, convert to a Go interface and then
	// load that as a document for gojsonschema to process
	filename, _ := filepath.Abs(element)
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var spec interface{}
	err = yaml.Unmarshal(yamlFile, &spec)
	if err != nil {
		panic(err)
	}

	body := convert_to_string_keys(spec)
	documentLoader := gojsonschema.NewGoLoader(body)

	cast := body.(map[string]interface{})
	kind := strings.ToLower((cast["kind"].(string)))

	// We have both the upstream Kubernetes schemas and the OpenShift schemas available
	// the tool can toggle between then using the --openshift boolean flag and here we
	// use that to select which repository to get the schema from
	var schema_type string
	if OpenShift {
		schema_type = "openshift"
	} else {
		schema_type = "kubernetes"
	}

	// Most of the directories which store the schemas are prefixed with a v so as to
	// match the tagging in the Kubernetes repository, apart from master.
	normalised_version := Version
	if Version != "master" {
		normalised_version = "v" + normalised_version
	}

	schema := fmt.Sprintf("https://raw.githubusercontent.com/garethr/%s-json-schema/master/%s/%s.json", schema_type, normalised_version, kind)

	schemaLoader := gojsonschema.NewReferenceLoader(schema)

	// Without forcing these types the schema fails to load
	// Need to Work out proper handling for these types
	gojsonschema.FormatCheckers.Add("int64", ValidFormat{})
	gojsonschema.FormatCheckers.Add("byte", ValidFormat{})
	gojsonschema.FormatCheckers.Add("int32", ValidFormat{})
	gojsonschema.FormatCheckers.Add("int-or-string", ValidFormat{})

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		panic(err)
	}

	if result.Valid() {
		fmt.Println("The document is valid")
	} else {
		fmt.Println("The document is not valid. see errors :")
		for _, desc := range result.Errors() {
			fmt.Println(desc)
			fmt.Println(desc.Type())
			fmt.Println(desc.Field())
			fmt.Println(desc.Description())
			fmt.Println(desc.Details())
		}
		os.Exit(1)
	}
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.Flags().StringVarP(&Version, "kubernetes-version", "v", "master", "Version of Kubernetes to validate against")
	RootCmd.Flags().BoolVarP(&OpenShift, "openshift", "", false, "Use OpenShift schemas instead of upstream Kubernetes")
}
