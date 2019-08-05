package kubeval

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"

	"github.com/instrumenta/kubeval/log"
)

// ValidationResult contains the details from
// validating a given Kubernetes resource
type ValidationResult struct {
       FileName               string
       Kind                   string
       APIVersion             string
       ValidatedAgainstSchema bool
       Errors                 []gojsonschema.ResultError
}

func determineSchema(kind, apiVersion string, config *Config) string {
	// We have both the upstream Kubernetes schemas and the OpenShift schemas available
	// the tool can toggle between then using the config.Openshift boolean flag and here we
	// use that to select which repository to get the schema from

	// Most of the directories which store the schemas are prefixed with a v so as to
	// match the tagging in the Kubernetes repository, apart from master.
	normalisedVersion := config.KubernetesVersion
	if normalisedVersion != "master" {
		normalisedVersion = "v" + normalisedVersion
	}

	strictSuffix := ""
	if config.Strict {
		strictSuffix = "-strict"
	}

	groupParts := strings.Split(apiVersion, "/")
	versionParts := strings.Split(groupParts[0], ".")

	kindSuffix := ""
	if !config.OpenShift {
		if len(groupParts) == 1 {
			kindSuffix = "-" + strings.ToLower(versionParts[0])
		} else {
			kindSuffix = fmt.Sprintf("-%s-%s", strings.ToLower(versionParts[0]), strings.ToLower(groupParts[1]))
		}
	}

	baseURL := determineBaseURL(config)
	return fmt.Sprintf("%s/%s-standalone%s/%s%s.json", baseURL, normalisedVersion, strictSuffix, strings.ToLower(kind), kindSuffix)
}

func determineBaseURL(config *Config) string {
	// Order of precendence:
	// 1. If --openshift is passed, return the openshift schema location
	// 2. If a --schema-location is passed, use it
	// 3. If the KUBEVAL_SCHEMA_LOCATION is set, use it
	// 4. Otherwise, use the DefaultSchemaLocation

	if config.OpenShift {
		return OpenShiftSchemaLocation
	}

	if config.SchemaLocation != "" {
		return config.SchemaLocation
	}

	// We only care that baseURL has a value after this call, so we can
	// ignore LookupEnv's second return value
	baseURL, _ := os.LookupEnv("KUBEVAL_SCHEMA_LOCATION")
	if baseURL != "" {
		return baseURL
	}

	return DefaultSchemaLocation
}

// validateResource validates a single Kubernetes resource against
// the relevant schema, detecting the type of resource automatically
func validateResource(data []byte, schemaCache map[string]*gojsonschema.Schema, config *Config) (ValidationResult, error) {
	result := ValidationResult{}
	result.FileName = config.FileName
	var body map[string]interface{}
	err := yaml.Unmarshal(data, &body)
	if err != nil {
		return result, fmt.Errorf("Failed to decode YAML from %s: %s", result.FileName, err.Error())
	} else if body == nil {
		return result, nil
	}

	kind, err := getString(body, "kind")
	if err != nil {
		return result, err
	}
	result.Kind = kind

	apiVersion, err := getString(body, "apiVersion")
	if err != nil {
		return result, err
	}
	result.APIVersion = apiVersion

	if in(config.KindsToSkip, kind) {
		return result, nil
	}

	schemaErrors, err := validateAgainstSchema(body, &result, schemaCache, config)
	if err != nil {
		return result, err
	}
	result.Errors = schemaErrors
	return result, nil
}

func validateAgainstSchema(body interface{}, resource *ValidationResult, schemaCache map[string]*gojsonschema.Schema, config *Config) ([]gojsonschema.ResultError, error) {
	if config.IgnoreMissingSchemas {
		log.Warn("Warning: Set to ignore missing schemas")
	}
	schemaRef := determineSchema(resource.Kind, resource.APIVersion, config)
	schema, ok := schemaCache[schemaRef]
	if !ok {
		schemaLoader := gojsonschema.NewReferenceLoader(schemaRef)
		var err error
		schema, err = gojsonschema.NewSchema(schemaLoader)
		schemaCache[schemaRef] = schema

		if err != nil {
			return handleMissingSchema(fmt.Errorf("Failed initalizing schema %s: %s", schemaRef, err), config)
		}
	}

	if schema == nil {
		return handleMissingSchema(fmt.Errorf("Failed initalizing schema %s: see first error", schemaRef), config)
	}

	documentLoader := gojsonschema.NewGoLoader(body)
	results, err := schema.Validate(documentLoader)
	if err != nil {
		return []gojsonschema.ResultError{}, fmt.Errorf("Problem loading schema from the network at %s: %s", schemaRef, err)
	}
	resource.ValidatedAgainstSchema = true
	if !results.Valid() {
		return results.Errors(), nil
	}
	return []gojsonschema.ResultError{}, nil
}

func handleMissingSchema(err error, config *Config) ([]gojsonschema.ResultError, error) {
	if config.IgnoreMissingSchemas {
		return []gojsonschema.ResultError{}, nil
	}
	return []gojsonschema.ResultError{}, err
}

// NewSchemaCache returns a new schema cache to be used with
// ValidateWithCache
func NewSchemaCache() map[string]*gojsonschema.Schema {
	return make(map[string]*gojsonschema.Schema, 0)
}

// Validate a Kubernetes YAML file, parsing out individual resources
// and validating them all according to the  relevant schemas
func Validate(input []byte, conf ...*Config) ([]ValidationResult, error) {
	schemaCache := NewSchemaCache()
	return ValidateWithCache(input, schemaCache, conf...)
}

// ValidateWithCache validates a Kubernetes YAML file, parsing out individual resources
// and validating them all according to the relevant schemas
// Allows passing a kubeval.NewSchemaCache() to cache schemas in-memory
// between validations
func ValidateWithCache(input []byte, schemaCache map[string]*gojsonschema.Schema, conf ...*Config) ([]ValidationResult, error) {
	config := NewDefaultConfig()
	if len(conf) == 1 {
		config = conf[0]
	}

	results := make([]ValidationResult, 0)

	if len(input) == 0 {
		result := ValidationResult{}
		result.FileName = config.FileName
		results = append(results, result)
		return results, nil
	}

	bits := bytes.Split(input, []byte(detectLineBreak(input)+"---"+detectLineBreak(input)))

	var errors *multierror.Error

	// special case regexp for helm
	helmSourcePattern := regexp.MustCompile(`^(?:---` + detectLineBreak(input) + `)?# Source: (.*)`)

	// Save the fileName we were provided; if we detect a new fileName
	// we'll use that, but we'll need to revert to the default afterward
	originalFileName := config.FileName
	defer func() {
		// revert the filename back to the original
		config.FileName = originalFileName
	}()

	for _, element := range bits {
		if len(element) > 0 {
			if found := helmSourcePattern.FindStringSubmatch(string(element)); found != nil {
				config.FileName = found[1]
			}

			result, err := validateResource(element, schemaCache, config)
			if err != nil {
				errors = multierror.Append(errors, err)
				if config.ExitOnError {
					return results, errors
				}
			}
			results = append(results, result)
		} else {
			result := ValidationResult{}
			result.FileName = config.FileName
			results = append(results, result)
		}
	}
	return results, errors.ErrorOrNil()
}
