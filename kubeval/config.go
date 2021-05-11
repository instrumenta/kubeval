package kubeval

import (
	"fmt"

	"github.com/spf13/cobra"
)

// DefaultSchemaLocation is the default location to search for schemas
const DefaultSchemaLocation = "https://kubernetesjsonschema.dev"

// OpenShiftSchemaLocation is the alternative location for OpenShift specific schemas
const OpenShiftSchemaLocation = "https://raw.githubusercontent.com/garethr/openshift-json-schema/master"

// A Config object contains various configuration data for kubeval
type Config struct {
	// DefaultNamespace is the namespace to assume in resources
	// if no namespace is set in `metadata:namespace` (as used with
	// `kubectl apply --namespace ...` or `helm install --namespace ...`,
	// for example)
	DefaultNamespace string

	// KubernetesVersion represents the version of Kubernetes
	// for which we should load the schema
	KubernetesVersion string

	// SchemaLocation is the base URL from which to search for schemas.
	// It can be either a remote location or a local directory
	SchemaLocation string

	// AdditionalSchemaLocations is a list of alternative base URLs from
	// which to search for schemas, given that the desired schema was not
	// found at SchemaLocation
	AdditionalSchemaLocations []string

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

	// KindsToReject is a list of case-sensitive prohibited kubernetes resources types
	KindsToReject []string

	// FileName is the name to be displayed when testing manifests read from stdin
	FileName string

	// OutputFormat is the name of the output formatter which will be used when
	// reporting results to the user.
	OutputFormat string

	// LoggingLevel configures the logging level for whether we should log output when using the human layout
	LoggingLevel string

	// Quiet indicates whether non-results output should be emitted to the applications
	// log.
	Quiet bool

	// IgnoreHelmSource if enabled we don't replace the file name used in errors reported with
	// the contents of the `# Source` comment file name
	IgnoreHelmSource bool

	// InsecureSkipTLSVerify controls whether to skip TLS certificate validation
	// when retrieving schema content over HTTPS
	InsecureSkipTLSVerify bool
}

// NewDefaultConfig creates a Config with default values
func NewDefaultConfig() *Config {
	return &Config{
		DefaultNamespace:  "default",
		FileName:          "stdin",
		KubernetesVersion: "master",
	}
}

// AddKubevalFlags adds the default flags for kubeval to cmd
func AddKubevalFlags(cmd *cobra.Command, config *Config) *cobra.Command {
	cmd.Flags().StringVarP(&config.DefaultNamespace, "default-namespace", "n", "default", "Namespace to assume in resources if no namespace is set in metadata:namespace")
	cmd.Flags().BoolVar(&config.ExitOnError, "exit-on-error", false, "Immediately stop execution when the first error is encountered")
	cmd.Flags().BoolVar(&config.IgnoreMissingSchemas, "ignore-missing-schemas", false, "Skip validation for resource definitions without a schema")
	cmd.Flags().BoolVar(&config.OpenShift, "openshift", false, "Use OpenShift schemas instead of upstream Kubernetes")
	cmd.Flags().BoolVar(&config.Strict, "strict", false, "Disallow additional properties not in schema")
	cmd.Flags().StringVarP(&config.FileName, "filename", "f", "stdin", "filename to be displayed when testing manifests read from stdin")
	cmd.Flags().StringSliceVar(&config.KindsToSkip, "skip-kinds", []string{}, "Comma-separated list of case-sensitive kinds to skip when validating against schemas")
	cmd.Flags().StringSliceVar(&config.KindsToReject, "reject-kinds", []string{}, "Comma-separated list of case-sensitive kinds to prohibit validating against schemas")
	cmd.Flags().StringVarP(&config.SchemaLocation, "schema-location", "s", "", "Base URL used to download schemas. Can also be specified with the environment variable KUBEVAL_SCHEMA_LOCATION.")
	cmd.Flags().StringSliceVar(&config.AdditionalSchemaLocations, "additional-schema-locations", []string{}, "Comma-seperated list of secondary base URLs used to download schemas")
	cmd.Flags().StringVarP(&config.KubernetesVersion, "kubernetes-version", "v", "master", "Version of Kubernetes to validate against")
	cmd.Flags().StringVarP(&config.OutputFormat, "output", "o", "", fmt.Sprintf("The format of the output of this script. Options are: %v", validOutputs()))
	cmd.Flags().StringVarP(&config.LoggingLevel, "log-level", "", "", fmt.Sprintf("The logging level for whether to log all levels or just warnings. Options are: %v", validLevels()))
	cmd.Flags().BoolVar(&config.Quiet, "quiet", false, "Silences any output aside from the direct results")
	cmd.Flags().BoolVar(&config.IgnoreHelmSource, "ignore-helm-source", false, "If enabled we preserve the file name for errors rather than replace it with the contents of the '# Source' comment in the generated helm resource")
	cmd.Flags().BoolVar(&config.InsecureSkipTLSVerify, "insecure-skip-tls-verify", false, "If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure")

	return cmd
}
