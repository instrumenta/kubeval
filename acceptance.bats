#!/usr/bin/env bats

@test "Pass when parsing a valid Kubernetes config YAML file" {
  run kubeval fixtures/valid.yaml
	[ "$status" -eq 0 ]
  [ "$output" = "The document fixtures/valid.yaml is a valid ReplicationController" ]
}

@test "Pass when parsing a valid Kubernetes config JSON file" {
  run kubeval fixtures/valid.json
	[ "$status" -eq 0 ]
  [ "$output" = "The document fixtures/valid.json is a valid Deployment" ]
}

@test "Pass when parsing a valid Kubernetes config file with int_to_string vars" {
  run kubeval fixtures/int_or_string.yaml
	[ "$status" -eq 0 ]
  [ "${lines[0]}" = "---> spec.ports.0.targetPort is an integer, but might be an int_or_string property" ]
  [ "${lines[1]}" = "The document fixtures/int_or_string.yaml is a valid Service" ]
}

@test "Fail when parsing an invalid Kubernetes config file" {
  run kubeval fixtures/invalid.yaml
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
  [ $(expr "$output" : "^Missing a kind key") -ne 0 ]
}

@test "Return relevant error for YAML missing kind key" {
  run kubeval fixtures/missing-kind.yaml
	[ "$status" -eq 1 ]
  [ $(expr "$output" : "^Missing a kind key") -ne 0 ]
}
