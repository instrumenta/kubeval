package kubeval

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"

	"github.com/garethr/kubeval/log"
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

func validateResource(data []byte, fileName string) bool {
	var spec interface{}
	err := yaml.Unmarshal(data, &spec)
	if err != nil {
		log.Error("Failed to decode YAML from", fileName)
		return false
	}

	body := convertToStringKeys(spec)

	documentLoader := gojsonschema.NewGoLoader(body)

	cast, _ := body.(map[string]interface{})
	if _, ok := cast["kind"]; !ok {
		log.Error("Missing a kind key in", fileName)
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
		log.Error("Problem loading schema from the network")
		log.Info(err)
		return false
	}

	if result.Valid() {
		log.Success("The document", fileName, "is a valid", kind)
		return true
	}

	log.Warn("The document", fileName, "is not a valid", kind)
	for _, desc := range result.Errors() {
		log.Info("-->", desc)
	}
	return false
}

// Validate a Kubernetes YAML file according to a relevant schema
// TODO This function requires a judicious amount of refactoring.
func Validate(config []byte, fileName string) bool {

	bits := bytes.Split(config, []byte("---\n"))

	results := make([]bool, len(bits))
	for i, element := range bits {
		result := validateResource(element, fileName)
		results[i] = result
	}
	for _, a := range results {
		if a == false {
			return false
		}
	}
	return true
}
