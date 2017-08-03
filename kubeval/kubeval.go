package kubeval

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"runtime"

	"github.com/hashicorp/go-multierror"
	"github.com/xeipuuv/gojsonschema"
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

// IsFormat always returns true and meets the
// gojsonschema.FormatChecker interface
func (f ValidFormat) IsFormat(input string) bool {
	return true
}

// ValidationResult contains the details from
// validating a given Kubernetes resource
type ValidationResult struct {
	FileName string
	Kind     string
	Errors   []gojsonschema.ResultError
}

// lineBreak returns the relevant platform specific line ending
func lineBreak() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

func determineSchema(kind string) string {
	// We have both the upstream Kubernetes schemas and the OpenShift schemas available
	// the tool can toggle between then using the --openshift boolean flag and here we
	// use that to select which repository to get the schema from
	var schemaType string
	if OpenShift {
		schemaType = "openshift"
	} else {
		schemaType = "kubernetes"
	}

	// Set a default Version to make usage as a library easier
	if Version == "" {
		Version = "master"
	}
	// Most of the directories which store the schemas are prefixed with a v so as to
	// match the tagging in the Kubernetes repository, apart from master.
	normalisedVersion := Version
	if Version != "master" {
		normalisedVersion = "v" + normalisedVersion
	}

	return fmt.Sprintf("https://raw.githubusercontent.com/garethr/%s-json-schema/master/%s-standalone/%s.json", schemaType, normalisedVersion, strings.ToLower(kind))
}

func determineKind(body interface{}) (string, error) {
	cast, _ := body.(map[string]interface{})
	if _, ok := cast["kind"]; !ok {
		return "", errors.New("Missing a kind key")
	}

	return cast["kind"].(string), nil
}

// validateResource validates a single Kubernetes resource against
// the relevant schema, detecting the type of resource automatically
func validateResource(data []byte, fileName string) (ValidationResult, error) {
	var spec interface{}
	result := ValidationResult{}
	result.FileName = fileName
	err := yaml.Unmarshal(data, &spec)
	if err != nil {
		return result, errors.New("Failed to decode YAML from " + fileName)
	}

	body := convertToStringKeys(spec)
	documentLoader := gojsonschema.NewGoLoader(body)

	kind, err := determineKind(body)
	if err != nil {
		return result, err
	}
	result.Kind = kind
	schema := determineSchema(kind)

	schemaLoader := gojsonschema.NewReferenceLoader(schema)

	// Without forcing these types the schema fails to load
	// Need to Work out proper handling for these types
	gojsonschema.FormatCheckers.Add("int64", ValidFormat{})
	gojsonschema.FormatCheckers.Add("byte", ValidFormat{})
	gojsonschema.FormatCheckers.Add("int32", ValidFormat{})
	gojsonschema.FormatCheckers.Add("int-or-string", ValidFormat{})

	results, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return result, errors.New("Problem loading schema from the network")
	}

	if results.Valid() {
		return result, nil
	}

	result.Errors = results.Errors()
	return result, nil
}

// Validate a Kubernetes YAML file, parsing out individual resources
// and validating them all according to the  relevant schemas
// TODO This function requires a judicious amount of refactoring.
func Validate(config []byte, fileName string) ([]ValidationResult, error) {
	if len(config) == 0 {
		return nil, errors.New("The document " + fileName + " appears to be empty")
	}

	bits := bytes.Split(config, []byte("---" + lineBreak()))

	results := make([]ValidationResult, 0)
	var errors *multierror.Error
	for _, element := range bits {
		if len(element) > 0 {
			result, err := validateResource(element, fileName)
			results = append(results, result)
			if err != nil {
				errors = multierror.Append(errors, err)
			}
		}
	}
	return results, errors.ErrorOrNil()
}
