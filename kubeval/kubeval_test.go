package kubeval

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestValidateBlankInput(t *testing.T) {
	blank := []byte("")
	_, err := Validate(blank, "sample")
	if err == nil {
		t.Errorf("Validate should fail when passed a blank string")
	}
}

func TestValidateValidInputs(t *testing.T) {
	var tests = []string{
		"valid.yaml",
		"valid.json",
		"multi_valid.yaml",
		"int_or_string.yaml",
		"null_array.yaml",
		"quantity.yaml",
		"extra_property.yaml",
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

func TestValidateInvalidInputs(t *testing.T) {
	var tests = []string{
		"blank.yaml",
		"missing-kind.json",
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
	schema := determineSchema("sample")
	if schema != "https://raw.githubusercontent.com/garethr/kubernetes-json-schema/master/master-standalone/sample.json" {
		t.Errorf("Schema should default to master")
	}
}

func TestDetermineSchemaForOpenShift(t *testing.T) {
	OpenShift = true
	schema := determineSchema("sample")
	if schema != "https://raw.githubusercontent.com/garethr/openshift-json-schema/master/master-standalone/sample.json" {
		t.Errorf("Should be able to toggle to OpenShift schemas")
	}
}

func TestDetermineSchemaForVersions(t *testing.T) {
	Version = "1.0"
	schema := determineSchema("sample")
	if schema != "https://raw.githubusercontent.com/garethr/openshift-json-schema/master/v1.0-standalone/sample.json" {
		t.Errorf("Should be able to specify a version")
	}
}

func TestDetermineSchemaForSchemaLocation(t *testing.T) {
	SchemaLocation = "file:///home/me"
	schema := determineSchema("sample")
	expectedSchema := "file:///home/me/openshift-json-schema/master/v1.0-standalone/sample.json"
	if schema != expectedSchema {
		t.Errorf("Should be able to specify a schema location, expected %s, got %s instead ", expectedSchema, schema)
	}
}

func TestDetermineKind(t *testing.T) {
	_, err := determineKind("sample")
	if err == nil {
		t.Errorf("Shouldn't be able to find a kind  when passed a blank string")
	}
}
