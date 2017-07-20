package kubeval

import "testing"

func TestValidateBlankInput(t *testing.T) {
	blank := []byte("")
	_, err := Validate(blank, "sample")
	if err == nil {
		t.Errorf("Validate should fail when passed a blank string")
	}
}

func TestDetermineSchema(t *testing.T) {
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

func TestDetermineKind(t *testing.T) {
	_, err := determineKind("sample")
	if err == nil {
		t.Errorf("Shouldn't be able to find a kind  when passed a blank string")
	}
}
