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
	config := NewDefaultConfig()
	config.FileName = "blank"
	_, err := Validate(blank, config)
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
		config := NewDefaultConfig()
		config.FileName = test
		_, err := Validate(fileContents, config)
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
		config := NewDefaultConfig()
		config.FileName = test
		_, err := ValidateWithCache(fileContents, schemaCache, config)
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
		config := NewDefaultConfig()
		config.FileName = test
		_, err := Validate(fileContents, config)
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
	config := NewDefaultConfig()
	config.FileName = "multi_valid_source.yaml"
	results, err := Validate(fileContents, config)
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
	config.FileName = "extra_property.yaml"
	filePath, _ := filepath.Abs("../fixtures/extra_property.yaml")
	fileContents, _ := ioutil.ReadFile(filePath)
	results, _ := Validate(fileContents, config)
	if len(results[0].Errors) == 0 {
		t.Errorf("Validate should not pass when testing for additional properties not in schema")
	}
}

func TestValidateMultipleVersions(t *testing.T) {
	config := NewDefaultConfig()
	config.Strict = true
	config.FileName = "valid_version.yaml"
	config.KubernetesVersion = "1.14.0"
	filePath, _ := filepath.Abs("../fixtures/valid_version.yaml")
	fileContents, _ := ioutil.ReadFile(filePath)
	results, err := Validate(fileContents, config)
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
		config := NewDefaultConfig()
		config.FileName = test
		results, _ := Validate(fileContents, config)
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
		config.FileName = test
		_, err := Validate(fileContents, config)
		if err == nil {
			t.Errorf("Validate should not pass when testing invalid configuration in " + test)
		} else if merr, ok := err.(*multierror.Error); ok {
			if len(merr.Errors) != 1 {
				t.Errorf("Validate should encounter exactly 1 error when testing invalid configuration in " + test + " with ExitOnError=true")
			}
		}
		config.ExitOnError = false
		_, err = Validate(fileContents, config)
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

func TestDetermineSchemaURL(t *testing.T) {
	var tests = []struct {
		config   *Config
		baseURL  string
		kind     string
		version  string
		expected string
	}{
		{
			config:   NewDefaultConfig(),
			baseURL:  "https://base",
			kind:     "sample",
			version:  "v1",
			expected: "https://base/master-standalone/sample-v1.json",
		},
		{
			config:   &Config{KubernetesVersion: "2"},
			baseURL:  "https://base",
			kind:     "sample",
			version:  "v1",
			expected: "https://base/v2-standalone/sample-v1.json",
		},
		{
			config:   &Config{KubernetesVersion: "master", Strict: true},
			baseURL:  "https://base",
			kind:     "sample",
			version:  "v1",
			expected: "https://base/master-standalone-strict/sample-v1.json",
		},
		{
			config:   NewDefaultConfig(),
			baseURL:  "https://base",
			kind:     "sample",
			version:  "extensions/v1beta1",
			expected: "https://base/master-standalone/sample-extensions-v1beta1.json",
		},
		{
			config:   &Config{KubernetesVersion: "master", OpenShift: true},
			baseURL:  "https://base",
			kind:     "sample",
			version:  "v1",
			expected: "https://base/master-standalone/sample.json",
		},
	}
	for _, test := range tests {
		schemaURL := determineSchemaURL(test.baseURL, test.kind, test.version, test.config)
		if schemaURL != test.expected {
			t.Errorf("Schema URL should be %s, got %s", test.expected, schemaURL)
		}
	}
}

func TestDetermineSchemaForSchemaLocation(t *testing.T) {
	oldVal, found := os.LookupEnv("KUBEVAL_SCHEMA_LOCATION")
	defer func() {
		if found {
			os.Setenv("KUBEVAL_SCHEMA_LOCATION", oldVal)
		} else {
			os.Unsetenv("KUBEVAL_SCHEMA_LOCATION")
		}
	}()

	var tests = []struct {
		config   *Config
		envVar   string
		expected string
	}{
		{
			config:   &Config{OpenShift: true},
			envVar:   "",
			expected: OpenShiftSchemaLocation,
		},
		{
			config:   &Config{SchemaLocation: "https://base"},
			envVar:   "",
			expected: "https://base",
		},
		{
			config:   &Config{},
			envVar:   "https://base",
			expected: "https://base",
		},
		{
			config:   &Config{},
			envVar:   "",
			expected: DefaultSchemaLocation,
		},
	}
	for i, test := range tests {
		os.Setenv("KUBEVAL_SCHEMA_LOCATION", test.envVar)
		schemaBaseURL := determineSchemaBaseURL(test.config)
		if schemaBaseURL != test.expected {
			t.Errorf("test #%d: Schema Base URL should be %s, got %s", i, test.expected, schemaBaseURL)
		}
	}
}

func TestGetString(t *testing.T) {
	var tests = []struct {
		body        map[string]interface{}
		key         string
		expectedVal string
		expectError bool
	}{
		{
			body:        map[string]interface{}{"goodKey": "goodVal"},
			key:         "goodKey",
			expectedVal: "goodVal",
			expectError: false,
		},
		{
			body:        map[string]interface{}{},
			key:         "missingKey",
			expectedVal: "",
			expectError: true,
		},
		{
			body:        map[string]interface{}{"nilKey": nil},
			key:         "nilKey",
			expectedVal: "",
			expectError: true,
		},
		{
			body:        map[string]interface{}{"badKey": 5},
			key:         "badKey",
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
	config := NewDefaultConfig()
	config.FileName = "test_crd.yaml"
	filePath, _ := filepath.Abs("../fixtures/test_crd.yaml")
	fileContents, _ := ioutil.ReadFile(filePath)
	_, err := Validate(fileContents)
	if err == nil {
		t.Errorf("For custom CRD's with schema missing we should error without IgnoreMissingSchemas flag")
	}

	config.IgnoreMissingSchemas = true
	results, _ := Validate(fileContents, config)
	if len(results[0].Errors) != 0 {
		t.Errorf("For custom CRD's with schema missing we should skip with IgnoreMissingSchemas flag")
	}

	config.IgnoreMissingSchemas = false
	config.KindsToSkip = []string{"SealedSecret"}
	results, _ = Validate(fileContents, config)
	if len(results[0].Errors) != 0 {
		t.Errorf("We should skip resources listed in KindsToSkip")
	}
}

func TestAdditionalSchemas(t *testing.T) {
	// This test uses a hack - first tell kubeval to use a bogus URL as its
	// primary search location, then give the DefaultSchemaLocation as an
	// additional schema.
	// This should cause kubeval to fail when looking for the schema in the
	// primary location, then succeed when it finds the schema at the
	// "additional location"
	config := NewDefaultConfig()
	config.SchemaLocation = "testLocation"
	config.AdditionalSchemaLocations = []string{DefaultSchemaLocation}

	config.FileName = "valid.yaml"
	filePath, _ := filepath.Abs("../fixtures/valid.yaml")
	fileContents, _ := ioutil.ReadFile(filePath)
	results, err := Validate(fileContents, config)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	} else if len(results[0].Errors) != 0 {
		t.Errorf("Validate should pass when testing a valid configuration using additional schema")
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
		"additional-schema-locations",
		"kubernetes-version",
	}

	for _, expected := range expectedFlags {
		flag := cmd.Flags().Lookup(expected)
		if flag == nil {
			t.Errorf("Could not find flag '%s'", expected)
		}
	}
}
