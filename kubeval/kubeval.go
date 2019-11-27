package kubeval

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"
)

// ValidFormat is a type for quickly forcing
// new formats on the gojsonschema loader
type ValidFormat struct{}

var cacheMu sync.Mutex

// IsFormat always returns true and meets the
// gojsonschema.FormatChecker interface
func (f ValidFormat) IsFormat(input interface{}) bool {
	return true
}

// ValidationResult contains the details from
// validating a given Kubernetes resource
type ValidationResult struct {
	FileName               string
	Kind                   string
	APIVersion             string
	ValidatedAgainstSchema bool
	Errors                 []gojsonschema.ResultError
}

// VersionKind returns a string representation of this result's apiVersion and kind
func (v *ValidationResult) VersionKind() string {
	return v.APIVersion + "/" + v.Kind
}

func SetupFormatCheckers() {
	// Without forcing these types the schema fails to load
	// Need to Work out proper handling for these types
	gojsonschema.FormatCheckers.Add("int64", ValidFormat{})
	gojsonschema.FormatCheckers.Add("byte", ValidFormat{})
	gojsonschema.FormatCheckers.Add("int32", ValidFormat{})
	gojsonschema.FormatCheckers.Add("int-or-string", ValidFormat{})
}

func determineSchemaURL(baseURL, kind, apiVersion string, config *Config) string {
	// We have both the upstream Kubernetes schemas and the OpenShift schemas available
	// the tool can toggle between then using the config.OpenShift boolean flag and here we
	// use that to format the URL to match the required specification.

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

	if config.OpenShift {
		// If we're using the openshift schemas, there's no further processing required
		return fmt.Sprintf("%s/%s-standalone%s/%s.json", baseURL, normalisedVersion, strictSuffix, strings.ToLower(kind))
	}

	groupParts := strings.Split(apiVersion, "/")
	versionParts := strings.Split(groupParts[0], ".")

	kindSuffix := "-" + strings.ToLower(versionParts[0])
	if len(groupParts) > 1 {
		kindSuffix += "-" + strings.ToLower(groupParts[1])
	}

	return fmt.Sprintf("%s/%s-standalone%s/%s%s.json", baseURL, normalisedVersion, strictSuffix, strings.ToLower(kind), kindSuffix)
}

func determineSchemaBaseURL(config *Config) string {
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
	var body map[string]interface{}
	err := yaml.Unmarshal(data, &body)
	if err != nil {
		return result, fmt.Errorf("Failed to decode YAML from %s: %s", result.FileName, err.Error())
	} else if body == nil {
		return result, nil
	}

	kind, err := getString(body, "kind")
	if err != nil {
		return result, fmt.Errorf("%s: %s", result.FileName, err.Error())
	}
	result.Kind = kind

	apiVersion, err := getString(body, "apiVersion")
	if err != nil {
		return result, fmt.Errorf("%s: %s", result.FileName, err.Error())
	}
	result.APIVersion = apiVersion

	if in(config.KindsToSkip, kind) {
		return result, nil
	}

	schemaErrors, err := validateAgainstSchema(body, &result, schemaCache, config)
	if err != nil {
		return result, fmt.Errorf("%s: %s", result.FileName, err.Error())
	}
	result.Errors = schemaErrors
	return result, nil
}

func validateAgainstSchema(body interface{}, resource *ValidationResult, schemaCache map[string]*gojsonschema.Schema, config *Config) ([]gojsonschema.ResultError, error) {
	var schema *gojsonschema.Schema
	var err error
	var ok bool

	cacheMu.Lock()
	schema, ok = schemaCache[resource.VersionKind()]
	cacheMu.Unlock()
	if !ok {
		schema, err = downloadSchema(resource, config)
		// We cache schemas that are not found in the main registry
		if err == nil || strings.Contains(err.Error(), "404") {
			cacheMu.Lock()
			schemaCache[resource.VersionKind()] = schema
			cacheMu.Unlock()
		}
	}

	if schema == nil {
		return handleMissingSchema(err, config)
	}

	// Without forcing these types the schema fails to load
	// Need to Work out proper handling for these types

	documentLoader := gojsonschema.NewGoLoader(body)
	results, err := schema.Validate(documentLoader)
	if err != nil {
		// This error can only happen if the Object to validate is poorly formed. There's no hope of saving this one
		wrappedErr := fmt.Errorf("Problem validating schema. Check JSON formatting: %s", err)
		return []gojsonschema.ResultError{}, wrappedErr
	}
	resource.ValidatedAgainstSchema = true
	if !results.Valid() {
		return results.Errors(), nil
	}

	return []gojsonschema.ResultError{}, nil
}

func downloadSchema(resource *ValidationResult, config *Config) (*gojsonschema.Schema, error) {
	primarySchemaBaseURL := determineSchemaBaseURL(config)
	primarySchemaRef := determineSchemaURL(primarySchemaBaseURL, resource.Kind, resource.APIVersion, config)
	schemaRefs := []string{primarySchemaRef}

	for _, additionalSchemaURLs := range config.AdditionalSchemaLocations {
		additionalSchemaRef := determineSchemaURL(additionalSchemaURLs, resource.Kind, resource.APIVersion, config)
		schemaRefs = append(schemaRefs, additionalSchemaRef)
	}

	var errors *multierror.Error

	for _, schemaRef := range schemaRefs {
		schemaLoader := gojsonschema.NewReferenceLoader(schemaRef)
		schema, err := gojsonschema.NewSchema(schemaLoader)
		if err == nil {
			// success!
			return schema, nil
		}
		// We couldn't find a schema for this URL, so take a note, then try the next URL
		wrappedErr := fmt.Errorf("Failed initalizing schema %s: %s", schemaRef, err)
		errors = multierror.Append(errors, wrappedErr)
	}

	if errors != nil {
		errors.ErrorFormat = singleLineErrorFormat
	}

	// TODO: this currently triggers a segfault in offline cases
	// We couldn't find a schema for this resource. Cache it's lack of existence, then stop
	//schemaCache[resource.VersionKind()] = nil
	return nil, errors.ErrorOrNil()
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
		results = append(results, result)
		return results, nil
	}

	bits := bytes.Split(input, []byte(detectLineBreak(input)+"---"+detectLineBreak(input)))

	var errors *multierror.Error

	// special case regexp for helm
	helmSourcePattern := regexp.MustCompile(`^(?:---` + detectLineBreak(input) + `)?# Source: (.*)`)

	// Save the fileName we were provided; if we detect a new fileName
	// we'll use that, but we'll need to revert to the default afterward

	filename := ""
	for _, element := range bits {
		if len(element) > 0 {
			result, err := validateResource(element, schemaCache, config)
			if err != nil {
				errors = multierror.Append(errors, err)
				if config.ExitOnError {
					return results, errors
				}
			}
			if found := helmSourcePattern.FindStringSubmatch(string(element)); found != nil {
				filename = found[1]
			}
			result.FileName = filename
			results = append(results, result)
		} else {
			results = append(results, ValidationResult{FileName: filename})
		}
	}

	if errors != nil {
		errors.ErrorFormat = singleLineErrorFormat
	}
	return results, errors.ErrorOrNil()
}

func singleLineErrorFormat(es []error) string {
	messages := make([]string, len(es))
	for i, e := range es {
		messages[i] = e.Error()
	}
	return strings.Join(messages, "\n")
}
