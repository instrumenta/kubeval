package kubeval

import "github.com/spf13/cobra"

// DefaultSchemaLocation is the default location to search for schemas
const DefaultSchemaLocation = "https://kubernetesjsonschema.dev"

// OpenShiftSchemaLocation is the alternative location for OpenShift specific schemas
const OpenShiftSchemaLocation = "https://raw.githubusercontent.com/garethr/openshift-json-schema/master"

// A Config object contains various configuration data for kubeval
type Config struct {
	// KubernetesVersion represents the version of Kubernetes
	// for which we should load the schema
	KubernetesVersion string

	// SchemaLocation is the base URL from which to search for schemas.
	// It can be either a remote location or a local directory
	SchemaLocation string

	// OpenShift represents whether to test against
	// upstream Kubernetes or the OpenShift schemas
	OpenShift bool

	// Strict tells kubeval whether to prohibit properties not in
	// the schema. The API allows them, but kubectl does not
	Strict bool

	// IgnoreMissingSchemas tells kubeval whether to skip validation
	// for resource definitions without an available schema
	IgnoreMissingSchemas bool

	// ExitOnError tells kubeval whether to halt processing upon the
	// first error encountered or to continue, aggregating all errors
	ExitOnError bool

	// KindsToSkip is a list of kubernetes resources types with which to skip
	// schema validation
	KindsToSkip []string

	// FileName is the name to be displayed when testing manifests read from stdin
	FileName string
}


// NewDefaultConfig creates a Config with default values
func NewDefaultConfig() *Config {
	return &Config{
		FileName: "stdin",
		KubernetesVersion: "master",
	}
}

// AddKubevalFlags adds the default flags for kubeval to cmd
func AddKubevalFlags(cmd *cobra.Command, config *Config) *cobra.Command {
	cmd.Flags().BoolVar(&config.ExitOnError, "exit-on-error", false, "Immediately stop execution when the first error is encountered")
	cmd.Flags().BoolVar(&config.IgnoreMissingSchemas, "ignore-missing-schemas", false, "Skip validation for resource definitions without a schema")
	cmd.Flags().BoolVar(&config.OpenShift, "openshift", false, "Use OpenShift schemas instead of upstream Kubernetes")
	cmd.Flags().BoolVar(&config.Strict, "strict", false, "Disallow additional properties not in schema")
	cmd.Flags().StringP("filename", "f", "stdin", "filename to be displayed when testing manifests read from stdin")
	cmd.Flags().StringSliceVar(&config.KindsToSkip, "skip-kinds", []string{}, "Comma-separated list of case-sensitive kinds to skip when validating against schemas")
	cmd.Flags().StringVar(&config.SchemaLocation, "schema-location", "", "Base URL used to download schemas. Can also be specified with the environment variable KUBEVAL_SCHEMA_LOCATION")
	cmd.Flags().StringVarP(&config.KubernetesVersion, "kubernetes-version", "v", "master", "Version of Kubernetes to validate against")
	return cmd
}
