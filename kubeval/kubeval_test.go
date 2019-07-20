package kubeval

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

func TestValidateBlankInput(t *testing.T) {
	blank := []byte("")
	_, err := Validate(blank, "sample")
	if err != nil {
		t.Errorf("Validate should pass when passed a blank string")
	}
}

func TestValidateValidInputs(t *testing.T) {
	var tests = []string{
		"blank.yaml",
		"comment.yaml",
		"valid.yaml",
		"valid.json",
		"multi_valid.yaml",
		"int_or_string.yaml",
		"null_array.yaml",
		"quantity.yaml",
		"extra_property.yaml",
		"full_domain_group.yaml",
		"unconventional_keys.yaml",
	}
	for _, test := range tests {
		filePath, _ := filepath.Abs("../fixtures/" + test)
		fileContents, _ := ioutil.ReadFile(filePath)
		_, err := Validate(fileContents, test)
		if err != nil {
			t.Errorf("Validate should pass when testing valid configuration in " + test)
		}
	}
}

func TestValidateValidInputsWithCache(t *testing.T) {
	var tests = []string{
		"blank.yaml",
		"comment.yaml",
		"valid.yaml",
		"valid.json",
		"multi_valid.yaml",
		"int_or_string.yaml",
		"null_array.yaml",
		"quantity.yaml",
		"extra_property.yaml",
		"full_domain_group.yaml",
		"unconventional_keys.yaml",
	}
	schemaCache := make(map[string]*gojsonschema.Schema, 0)

	for _, test := range tests {
		filePath, _ := filepath.Abs("../fixtures/" + test)
		fileContents, _ := ioutil.ReadFile(filePath)
		_, err := ValidateWithCache(fileContents, test, schemaCache)
		if err != nil {
			t.Errorf("Validate should pass when testing valid configuration in " + test)
		}
	}
}

func TestValidateInvalidInputs(t *testing.T) {
	var tests = []string{
		"missing_kind.yaml",
		"missing_kind_value.yaml",
	}
	for _, test := range tests {
		filePath, _ := filepath.Abs("../fixtures/" + test)
		fileContents, _ := ioutil.ReadFile(filePath)
		_, err := Validate(fileContents, test)
		if err == nil {
			t.Errorf("Validate should not pass when testing invalid configuration in " + test)
		}
	}
}

func TestStrictCatchesAdditionalErrors(t *testing.T) {
	Strict = true
	filePath, _ := filepath.Abs("../fixtures/extra_property.yaml")
	fileContents, _ := ioutil.ReadFile(filePath)
	results, _ := Validate(fileContents, "extra_property.yaml")
	if len(results[0].Errors) == 0 {
		t.Errorf("Validate should not pass when testing for additional properties not in schema")
	}
}

func TestValidateMultipleVersions(t *testing.T) {
	Strict = true
	Version = "1.14.0"
	OpenShift = false
	filePath, _ := filepath.Abs("../fixtures/valid_version.yaml")
	fileContents, _ := ioutil.ReadFile(filePath)
	results, err := Validate(fileContents, "valid_version.yaml")
	Version = ""
	if err != nil || len(results[0].Errors) > 0 {
		t.Errorf("Validate should pass when testing valid configuration with multiple versions: %v", err)
	}
}

func TestValidateInputsWithErrors(t *testing.T) {
	var tests = []string{
		"invalid.yaml",
		"multi_invalid.yaml",
	}
	for _, test := range tests {
		filePath, _ := filepath.Abs("../fixtures/" + test)
		fileContents, _ := ioutil.ReadFile(filePath)
		results, _ := Validate(fileContents, test)
		if len(results[0].Errors) == 0 {
			t.Errorf("Validate should not pass when testing invalid configuration in " + test)
		}
	}
}

func TestDetermineSchema(t *testing.T) {
	Strict = false
	schema := determineSchema("sample", "v1")
	if schema != "https://kubernetesjsonschema.dev/master-standalone/sample-v1.json" {
		t.Errorf("Schema should default to master, instead %s", schema)
	}
}

func TestDetermineSchemaForVersions(t *testing.T) {
	Version = "1.0"
	OpenShift = false
	schema := determineSchema("sample", "v1")
	if schema != "https://kubernetesjsonschema.dev/v1.0-standalone/sample-v1.json" {
		t.Errorf("Should be able to specify a version, instead %s", schema)
	}
}

func TestDetermineSchemaForOpenShift(t *testing.T) {
	OpenShift = true
	Version = "master"
	schema := determineSchema("sample", "v1")
	if schema != "https://raw.githubusercontent.com/garethr/openshift-json-schema/master/master-standalone/sample.json" {
		t.Errorf("Should be able to toggle to OpenShift schemas, instead %s", schema)
	}
}

func TestDetermineSchemaForSchemaLocation(t *testing.T) {
	OpenShift = false
	Version = "master"
	SchemaLocation = "file:///home/me"
	schema := determineSchema("sample", "v1")
	expectedSchema := "file:///home/me/master-standalone/sample-v1.json"
	if schema != expectedSchema {
		t.Errorf("Should be able to specify a schema location, expected %s, got %s instead ", expectedSchema, schema)
	}
}

func TestDetermineKind(t *testing.T) {
	_, err := determineKind("sample")
	if err == nil {
		t.Errorf("Shouldn't be able to find a kind when passed a blank string")
	}
}

func TestSkipCrdSchemaMiss(t *testing.T) {
	filePath, _ := filepath.Abs("../fixtures/test_crd.yaml")
	fileContents, _ := ioutil.ReadFile(filePath)
	_, err := Validate(fileContents, "test_crd.yaml")
	if err == nil {
		t.Errorf("For custom CRD's with schema missing we should error without IgnoreMissingSchemas flag")
	}

	IgnoreMissingSchemas = true
	results, _ := Validate(fileContents, "test_crd.yaml")
	if len(results[0].Errors) != 0 {
		t.Errorf("For custom CRD's with schema missing we should skip with IgnoreMissingSchemas flag")
	}
}
