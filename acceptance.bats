#!/usr/bin/env bats

@test "Pass when parsing a valid Kubernetes config YAML file" {
  run bin/kubeval fixtures/valid.yaml
  [ "$status" -eq 0 ]
  [ "$output" = "The file fixtures/valid.yaml contains a valid ReplicationController" ]
}

@test "Pass when parsing a valid Kubernetes config YAML file on stdin" {
  run bash -c "cat fixtures/valid.yaml | bin/kubeval"
  [ "$status" -eq 0 ]
  [ "$output" = "The file stdin contains a valid ReplicationController" ]
}

@test "Pass when parsing a valid Kubernetes config YAML file explicitly on stdin" {
  run bash -c "cat fixtures/valid.yaml | bin/kubeval -"
  [ "$status" -eq 0 ]
  [ "$output" = "The file stdin contains a valid ReplicationController" ]
}

@test "Pass when parsing a valid Kubernetes config JSON file" {
  run bin/kubeval fixtures/valid.json
  [ "$status" -eq 0 ]
  [ "$output" = "The file fixtures/valid.json contains a valid Deployment" ]
}

@test "Pass when parsing a Kubernetes file with string and integer quantities" {
  run bin/kubeval fixtures/quantity.yaml
  [ "$status" -eq 0 ]
  [ "$output" = "The file fixtures/quantity.yaml contains a valid LimitRange" ]
}

@test "Pass when parsing a valid Kubernetes config file with int_to_string vars" {
  run bin/kubeval fixtures/int_or_string.yaml
  [ "$status" -eq 0 ]
  [ "$output" = "The file fixtures/int_or_string.yaml contains a valid Service" ]
}

@test "Pass when parsing a valid Kubernetes config file with null arrays" {
  run bin/kubeval fixtures/null_array.yaml
  [ "$status" -eq 0 ]
  [ "$output" = "The file fixtures/null_array.yaml contains a valid Deployment" ]
}

@test "Pass when parsing a valid Kubernetes config file with null strings" {
  run bin/kubeval fixtures/null_string.yaml
  [ "$status" -eq 0 ]
  [ "$output" = "The file fixtures/null_string.yaml contains a valid Service" ]
}

@test "Pass when parsing a multi-document config file" {
  run bin/kubeval fixtures/multi_valid.yaml
  [ "$status" -eq 0 ]
}

@test "Fail when parsing a multi-document config file with one invalid resource" {
  run bin/kubeval fixtures/multi_invalid.yaml
  [ "$status" -eq 1 ]
}

@test "Fail when parsing an invalid Kubernetes config file" {
  run bin/kubeval fixtures/invalid.yaml
  [ "$status" -eq 1 ]
}

@test "Fail when parsing an invalid Kubernetes config file on stdin" {
  run bash -c "cat fixtures/invalid.yaml | bin/kubeval -"
  [ "$status" -eq 1 ]
}

@test "Return relevant error for non-existent file" {
  run bin/kubeval fixtures/not-here
  [ "$status" -eq 1 ]
  [ $(expr "$output" : "^Could not open file") -ne 0 ]
}

@test "Pass when parsing a blank config file" {
   run bin/kubeval fixtures/blank.yaml
   [ "$status" -eq 0 ]
   [ "$output" = "The file fixtures/blank.yaml contains an empty YAML document" ]
 }

 @test "Pass when parsing a blank config file with a comment" {
   run bin/kubeval fixtures/comment.yaml
   [ "$status" -eq 0 ]
   [ "$output" = "The file fixtures/comment.yaml contains an empty YAML document" ]
 }

@test "Return relevant error for YAML missing kind key" {
  run bin/kubeval fixtures/missing_kind.yaml
  [ "$status" -eq 1 ]
}

@test "Fail when parsing a config with additional properties and strict set" {
  run bin/kubeval --strict fixtures/extra_property.yaml
  [ "$status" -eq 1 ]
}

@test "Fail when parsing a config with a kind key but no value" {
  run bin/kubeval fixtures/missing_kind_value.yaml
  [ "$status" -eq 1 ]
}

@test "Pass when parsing a config with additional properties" {
  run bin/kubeval fixtures/extra_property.yaml
  [ "$status" -eq 0 ]
}

@test "Fail when parsing a config with CRD" {
  run bin/kubeval fixtures/test_crd.yaml
  [ "$status" -eq 1 ]
}

@test "Pass when parsing a config with CRD and ignoring missing schemas" {
  run bin/kubeval --ignore-missing-schemas fixtures/test_crd.yaml
  [ "$status" -eq 0 ]
}
