package kubeval

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
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

func TestValidateSourceExtraction(t *testing.T) {
	expectedFileNames := []string{
		"chart/templates/primary.yaml",   // first from primary template
		"chart/templates/primary.yaml",   // second resource from primary template
		"chart/templates/secondary.yaml", // first resource from secondary template
		"chart/templates/secondary.yaml", // second resource from secondary template
		"chart/templates/frontend.yaml",  // first resource from frontend template
		"chart/templates/frontend.yaml",  // second resource from frontend template
		"chart/templates/frontend.yaml",  // empty resource no comment
		"chart/templates/frontend.yaml",  // empty resource with comment
	}
	filePath, _ := filepath.Abs("../fixtures/multi_valid_source.yaml")
	fileContents, _ := ioutil.ReadFile(filePath)
	results, err := Validate(fileContents, "multi_valid_source.yaml")
	if err != nil {
		t.Fatalf("Unexpected error while validating source: %v", err)
	}
	for i, r := range results {
		if r.FileName != expectedFileNames[i] {
			t.Errorf("%v: expected filename [%v], got [%v]", i, expectedFileNames[i], r.FileName)
		}
	}
}

func TestStrictCatchesAdditionalErrors(t *testing.T) {
	config := NewDefaultConfig()
	config.Strict = true
	filePath, _ := filepath.Abs("../fixtures/extra_property.yaml")
	fileContents, _ := ioutil.ReadFile(filePath)
	results, _ := Validate(fileContents, "extra_property.yaml", config)
	if len(results[0].Errors) == 0 {
		t.Errorf("Validate should not pass when testing for additional properties not in schema")
	}
}

func TestValidateMultipleVersions(t *testing.T) {
	config := NewDefaultConfig()
	config.Strict = true
	config.KubernetesVersion = "1.14.0"
	filePath, _ := filepath.Abs("../fixtures/valid_version.yaml")
	fileContents, _ := ioutil.ReadFile(filePath)
	results, err := Validate(fileContents, "valid_version.yaml", config)
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

func TestValidateMultipleResourcesWithErrors(t *testing.T) {
	var tests = []string{
		"multi_invalid_resources.yaml",
	}
	for _, test := range tests {
		config := NewDefaultConfig()
		filePath, _ := filepath.Abs("../fixtures/" + test)
		fileContents, _ := ioutil.ReadFile(filePath)
		config.ExitOnError = true
		_, err := Validate(fileContents, test, config)
		if err == nil {
			t.Errorf("Validate should not pass when testing invalid configuration in " + test)
		} else if merr, ok := err.(*multierror.Error); ok {
			if len(merr.Errors) != 1 {
				t.Errorf("Validate should encounter exactly 1 error when testing invalid configuration in " + test + " with ExitOnError=true")
			}
		}
		config.ExitOnError = false
		_, err = Validate(fileContents, test, config)
		if err == nil {
			t.Errorf("Validate should not pass when testing invalid configuration in " + test)
		} else if merr, ok := err.(*multierror.Error); ok {
			if len(merr.Errors) != 5 {
				t.Errorf("Validate should encounter exactly 5 errors when testing invalid configuration in " + test)
			}
		} else if !ok {
			t.Errorf("Validate should encounter exactly 5 errors when testing invalid configuration in " + test)
		}
	}
}

func TestDetermineSchema(t *testing.T) {
	config := NewDefaultConfig()
	schema := determineSchema("sample", "v1", config)
	if schema != "https://kubernetesjsonschema.dev/master-standalone/sample-v1.json" {
		t.Errorf("Schema should default to master, instead %s", schema)
	}
}

func TestDetermineSchemaForVersions(t *testing.T) {
	config := NewDefaultConfig()
	config.KubernetesVersion = "1.0"
	schema := determineSchema("sample", "v1", config)
	if schema != "https://kubernetesjsonschema.dev/v1.0-standalone/sample-v1.json" {
		t.Errorf("Should be able to specify a version, instead %s", schema)
	}
}

func TestDetermineSchemaForOpenShift(t *testing.T) {
	config := NewDefaultConfig()
	config.OpenShift = true
	schema := determineSchema("sample", "v1", config)
	if schema != "https://raw.githubusercontent.com/garethr/openshift-json-schema/master/master-standalone/sample.json" {
		t.Errorf("Should be able to toggle to OpenShift schemas, instead %s", schema)
	}
}

func TestDetermineSchemaForSchemaLocation(t *testing.T) {
	config := NewDefaultConfig()
	config.SchemaLocation = "file:///home/me"
	schema := determineSchema("sample", "v1", config)
	expectedSchema := "file:///home/me/master-standalone/sample-v1.json"
	if schema != expectedSchema {
		t.Errorf("Should be able to specify a schema location, expected %s, got %s instead ", expectedSchema, schema)
	}
}

func TestDetermineSchemaForEnvVariable(t *testing.T) {
	oldVal, found := os.LookupEnv("KUBEVAL_SCHEMA_LOCATION")
	defer func() {
		if found {
			os.Setenv("KUBEVAL_SCHEMA_LOCATION", oldVal)
		} else {
			os.Unsetenv("KUBEVAL_SCHEMA_LOCATION")
		}
	}()
	config := NewDefaultConfig()
	os.Setenv("KUBEVAL_SCHEMA_LOCATION", "file:///home/me")
	schema := determineSchema("sample", "v1", config)
	expectedSchema := "file:///home/me/master-standalone/sample-v1.json"
	if schema != expectedSchema {
		t.Errorf("Should be able to specify a schema location, expected %s, got %s instead ", expectedSchema, schema)
	}
}

func TestGetString(t *testing.T) {
	var tests = []struct{
		body map[string]interface{}
		key string
		expectedVal string
		expectError bool
	}{
		{
			body: map[string]interface{}{"goodKey": "goodVal"},
			key: "goodKey",
			expectedVal: "goodVal",
			expectError: false,
		},
		{
			body: map[string]interface{}{},
			key: "missingKey",
			expectedVal: "",
			expectError: true,
		},
		{
			body: map[string]interface{}{"nilKey": nil},
			key: "nilKey",
			expectedVal: "",
			expectError: true,
		},
		{
			body: map[string]interface{}{"badKey": 5},
			key: "badKey",
			expectedVal: "",
			expectError: true,
		},
	}

	for _, test := range tests {
		actualVal, err := getString(test.body, test.key)
		if err != nil {
			if !test.expectError {
				t.Errorf("Unexpected error: %s", err.Error())
			}
			// We expected this error, so move to the next test
			continue
		}
		if test.expectError {
			t.Errorf("Expected an error, but didn't receive one")
			continue
		}
		if actualVal != test.expectedVal {
			t.Errorf("Expected %s, got %s", test.expectedVal, actualVal)
		}
	}
}

func TestSkipCrdSchemaMiss(t *testing.T) {
	filePath, _ := filepath.Abs("../fixtures/test_crd.yaml")
	fileContents, _ := ioutil.ReadFile(filePath)
	_, err := Validate(fileContents, "test_crd.yaml")
	if err == nil {
		t.Errorf("For custom CRD's with schema missing we should error without IgnoreMissingSchemas flag")
	}

	config := NewDefaultConfig()
	config.IgnoreMissingSchemas = true
	results, _ := Validate(fileContents, "test_crd.yaml", config)
	if len(results[0].Errors) != 0 {
		t.Errorf("For custom CRD's with schema missing we should skip with IgnoreMissingSchemas flag")
	}

	config.IgnoreMissingSchemas = false
	config.KindsToSkip = []string{"SealedSecret"}
	results, _ = Validate(fileContents, "test_crd.yaml", config)
	if len(results[0].Errors) != 0 {
		t.Errorf("We should skip resources listed in KindsToSkip")
	}
}

func TestFlagAdding(t *testing.T) {
	cmd := &cobra.Command{}
	config := &Config{}

	AddKubevalFlags(cmd, config)

	expectedFlags := []string{
		"exit-on-error",
		"ignore-missing-schemas",
		"openshift",
		"strict",
		"filename",
		"skip-kinds",
		"schema-location",
		"kubernetes-version",
	}

	for _, expected := range expectedFlags {
		flag := cmd.Flags().Lookup(expected)
		if flag == nil {
			t.Errorf("Could not find flag '%s'", expected)
		}
	}
}
