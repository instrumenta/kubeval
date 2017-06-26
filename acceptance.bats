#!/usr/bin/env bats

@test "Pass when parsing a valid Kubernetes config file" {
  run ./kubeval fixtures/valid.yaml
	[ "$status" -eq 0 ]
  [ "$output" = "The document fixtures/valid.yaml is a valid ReplicationController" ]
}

@test "Fail when parsing an invalid Kubernetes config file" {
  run ./kubeval fixtures/invalid.yaml
	[ "$status" -eq 1 ]
}

@test "Return relevant error for non-existent file" {
  run ./kubeval fixtures/not-here
	[ "$status" -eq 1 ]
  [ $(expr "$output" : "^Could not open file") -ne 0 ]
}

@test "Return relevant error for blank file" {
  skip
  run ./kubeval fixtures/blank.yaml
	[ "$status" -eq 1 ]
  [ $(expr "$output" : "^Missing a kind key") -ne 0 ]
}

@test "Return relevant error for YAML missing kind key" {
  run ./kubeval fixtures/missing-kind.yaml
	[ "$status" -eq 1 ]
  [ $(expr "$output" : "^Missing a kind key") -ne 0 ]
}
