#!/usr/bin/env bats

@test "Pass when parsing a valid Kubernetes config YAML file" {
  run kubeval fixtures/valid.yaml
	[ "$status" -eq 0 ]
  [ "$output" = "The document fixtures/valid.yaml contains a valid ReplicationController" ]
}

@test "Pass when parsing a valid Kubernetes config YAML file on stdin" {
  run bash -c "cat fixtures/valid.yaml | kubeval"
	[ "$status" -eq 0 ]
  [ "$output" = "The document stdin contains a valid ReplicationController" ]
}

@test "Pass when parsing a valid Kubernetes config JSON file" {
  run kubeval fixtures/valid.json
	[ "$status" -eq 0 ]
  [ "$output" = "The document fixtures/valid.json contains a valid Deployment" ]
}

@test "Pass when parsing a Kubernetes file with string and integer quantities" {
  run kubeval fixtures/quantity.yaml
	[ "$status" -eq 0 ]
  [ "$output" = "The document fixtures/quantity.yaml contains a valid LimitRange" ]
}

@test "Pass when parsing a valid Kubernetes config file with int_to_string vars" {
  run kubeval fixtures/int_or_string.yaml
	[ "$status" -eq 0 ]
  [ "$output" = "The document fixtures/int_or_string.yaml contains a valid Service" ]
}

@test "Pass when parsing a valid Kubernetes config file with null arrays" {
  run kubeval fixtures/null_array.yaml
	[ "$status" -eq 0 ]
  [ "$output" = "The document fixtures/null_array.yaml contains a valid Deployment" ]
}

@test "Pass when parsing a valid Kubernetes config file with null strings" {
  run kubeval fixtures/null_string.yaml
	[ "$status" -eq 0 ]
  [ "$output" = "The document fixtures/null_string.yaml contains a valid Service" ]
}

@test "Pass when parsing a multi-document config file" {
  run kubeval fixtures/multi_valid.yaml
	[ "$status" -eq 0 ]
}

@test "Fail when parsing a multi-document config file with one invalid resource" {
  run kubeval fixtures/multi_invalid.yaml
	[ "$status" -eq 1 ]
}

@test "Fail when parsing an invalid Kubernetes config file" {
  run kubeval fixtures/invalid.yaml
	[ "$status" -eq 1 ]
}

@test "Fail when parsing an invalid Kubernetes config file on stdin" {
  run bash -c "cat fixtures/invalid.yaml | kubeval"
	[ "$status" -eq 1 ]
}

@test "Return relevant error for non-existent file" {
  run kubeval fixtures/not-here
	[ "$status" -eq 1 ]
  [ $(expr "$output" : "^Could not open file") -ne 0 ]
}

@test "Return relevant error for blank file" {
  run kubeval fixtures/blank.yaml
	[ "$status" -eq 1 ]
  [ "$output" = "The document fixtures/blank.yaml appears to be empty" ]
}

@test "Return relevant error for YAML missing kind key" {
  run kubeval fixtures/missing-kind.yaml
	[ "$status" -eq 1 ]
}

@test "Fail when parsing a config with additional properties and strict set" {
  run kubeval --strict fixtures/extra_property.yaml
	[ "$status" -eq 1 ]
}

@test "Pass when parsing a config with additional properties" {
  run kubeval fixtures/extra_property.yaml
	[ "$status" -eq 0 ]
}
