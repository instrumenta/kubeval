package kubeval

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/viper"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"

	"github.com/instrumenta/kubeval/log"
)

// Version represents the version of Kubernetes
// for which we should load the schema
var Version string

// SchemaLocation represents what is the schema location,
/// where default value is maintener github project, but can be overriden
/// to either different github repo, or a local file
var SchemaLocation string

// DefaultSchemaLocation is the default value for
var DefaultSchemaLocation = "https://kubernetesjsonschema.dev"

// OpenShiftSchemaLocation is the alternative location for OpenShift specific schemas
var OpenShiftSchemaLocation = "https://raw.githubusercontent.com/garethr/openshift-json-schema/master"

// OpenShift represents whether to test against
// upstream Kubernetes of the OpenShift schemas
var OpenShift bool

// Strict tells kubeval whether to prohibit properties not in
// the schema. The API allows them, but kubectl does not
var Strict bool

// IgnoreMissingSchemas tells kubeval whether to skip validation
// for resource definitions without an available schema
var IgnoreMissingSchemas bool

// ValidFormat is a type for quickly forcing
// new formats on the gojsonschema loader
type ValidFormat struct{}

// IsFormat always returns true and meets the
// gojsonschema.FormatChecker interface
func (f ValidFormat) IsFormat(input interface{}) bool {
	return true
}

// ValidationResult contains the details from
// validating a given Kubernetes resource
type ValidationResult struct {
	FileName   string
	Kind       string
	APIVersion string
	Errors     []gojsonschema.ResultError
}

// detectLineBreak returns the relevant platform specific line ending
func detectLineBreak(haystack []byte) string {
	windowsLineEnding := bytes.Contains(haystack, []byte("\r\n"))
	if windowsLineEnding && runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

func determineSchema(kind string, apiVersion string) string {
	// We have both the upstream Kubernetes schemas and the OpenShift schemas available
	// the tool can toggle between then using the --openshift boolean flag and here we
	// use that to select which repository to get the schema from

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

	// Check Viper for environment variable support first.
	// Then check for an override in SchemaLocation
	// Finally settle on the default value
	baseURLFromEnv := viper.GetString("schema_location")
	var baseURL string
	if baseURLFromEnv != "" {
		baseURL = baseURLFromEnv
	} else if SchemaLocation == "" {
		if OpenShift {
			baseURL = OpenShiftSchemaLocation
		} else {
			baseURL = DefaultSchemaLocation
		}
	} else {
		baseURL = SchemaLocation
	}

	var strictSuffix string
	if Strict {
		strictSuffix = "-strict"
	} else {
		strictSuffix = ""
	}

	var kindSuffix string

	groupParts := strings.Split(apiVersion, "/")
	versionParts := strings.Split(groupParts[0], ".")

	if OpenShift {
		kindSuffix = ""
	} else {
		if len(groupParts) == 1 {
			kindSuffix = "-" + strings.ToLower(versionParts[0])
		} else {
			kindSuffix = fmt.Sprintf("-%s-%s", strings.ToLower(versionParts[0]), strings.ToLower(groupParts[1]))
		}
	}

	return fmt.Sprintf("%s/%s-standalone%s/%s%s.json", baseURL, normalisedVersion, strictSuffix, strings.ToLower(kind), kindSuffix)
}

func determineKind(body interface{}) (string, error) {
	cast, _ := body.(map[string]interface{})
	if _, ok := cast["kind"]; !ok {
		return "", errors.New("Missing a kind key")
	}
	if cast["kind"] == nil {
		return "", errors.New("Missing a kind value")
	}
	return cast["kind"].(string), nil
}

func determineAPIVersion(body interface{}) (string, error) {
	cast, _ := body.(map[string]interface{})
	if _, ok := cast["apiVersion"]; !ok {
		return "", errors.New("Missing a apiVersion key")
	}
	if cast["apiVersion"] == nil {
		return "", errors.New("Missing a apiVersion value")
	}
	return cast["apiVersion"].(string), nil
}

// validateResource validates a single Kubernetes resource against
// the relevant schema, detecting the type of resource automatically
func validateResource(data []byte, fileName string, schemaCache map[string]*gojsonschema.Schema) (ValidationResult, error) {
	var spec interface{}
	result := ValidationResult{}
	if IgnoreMissingSchemas {
		log.Warn("Warning: Set to ignore missing schemas")
	}
	result.FileName = fileName
	err := yaml.Unmarshal(data, &spec)
	if err != nil {
		return result, errors.New("Failed to decode YAML from " + fileName)
	}

	body := convertToStringKeys(spec)

	if body == nil {
		return result, nil
	}

	cast, _ := body.(map[string]interface{})
	if len(cast) == 0 {
		return result, nil
	}

	documentLoader := gojsonschema.NewGoLoader(body)

	kind, err := determineKind(body)
	if err != nil {
		return result, err
	}
	result.Kind = kind

	apiVersion, err := determineAPIVersion(body)
	if err != nil {
		return result, err
	}
	result.APIVersion = apiVersion

	schemaRef := determineSchema(kind, apiVersion)

	schema, ok := schemaCache[schemaRef]
	if !ok {
		if IgnoreMissingSchemas {
			return result, nil
		}
		schemaLoader := gojsonschema.NewReferenceLoader(schemaRef)
		schema, err = gojsonschema.NewSchema(schemaLoader)
		if err != nil {
			return result, fmt.Errorf("Failed initalizing schema %s: %s", schemaRef, err)
		}
		schemaCache[schemaRef] = schema
	}

	// Without forcing these types the schema fails to load
	// Need to Work out proper handling for these types
	gojsonschema.FormatCheckers.Add("int64", ValidFormat{})
	gojsonschema.FormatCheckers.Add("byte", ValidFormat{})
	gojsonschema.FormatCheckers.Add("int32", ValidFormat{})
	gojsonschema.FormatCheckers.Add("int-or-string", ValidFormat{})

	results, err := schema.Validate(documentLoader)
	if err != nil {
		return result, fmt.Errorf("Problem loading schema from the network at %s: %s", schemaRef, err)
	}

	if results.Valid() {
		return result, nil
	}

	result.Errors = results.Errors()
	return result, nil
}

// NewSchemaCache returns a new schema cache to be used with
// ValidateWithCache
func NewSchemaCache() map[string]*gojsonschema.Schema {
	return make(map[string]*gojsonschema.Schema, 0)
}

// Validate a Kubernetes YAML file, parsing out individual resources
// and validating them all according to the  relevant schemas
// TODO This function requires a judicious amount of refactoring.
func Validate(config []byte, fileName string) ([]ValidationResult, error) {
	schemaCache := NewSchemaCache()
	return ValidateWithCache(config, fileName, schemaCache)
}

// ValidateWithCache validates a Kubernetes YAML file, parsing out individual resources
// and validating them all according to the relevant schemas
// Allows passing a kubeval.NewSchemaCache() to cache schemas in-memory
// between validations
func ValidateWithCache(config []byte, fileName string, schemaCache map[string]*gojsonschema.Schema) ([]ValidationResult, error) {
	results := make([]ValidationResult, 0)

	if len(config) == 0 {
		result := ValidationResult{}
		result.FileName = fileName
		results = append(results, result)
		return results, nil
	}

	bits := bytes.Split(config, []byte(detectLineBreak(config)+"---"+detectLineBreak(config)))

	// special case regexp for helm
	helmSourcePattern := regexp.MustCompile(`^(?:---` + detectLineBreak(config) + `)?# Source: (.*)`)

	var errors *multierror.Error

	// Start with the filename we were provided; if we detect a new filename
	// we'll use that until we find a new one.
	detectedFileName := fileName

	for _, element := range bits {
		if len(element) > 0 {
			if found := helmSourcePattern.FindStringSubmatch(string(element)); found != nil {
				detectedFileName = found[1]
			}

			result, err := validateResource(element, detectedFileName, schemaCache)
			results = append(results, result)
			if err != nil {
				errors = multierror.Append(errors, err)
			}
		} else {
			result := ValidationResult{}
			result.FileName = detectedFileName
			results = append(results, result)
		}
	}
	return results, errors.ErrorOrNil()
}
