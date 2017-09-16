package cmd

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/garethr/kubeval/kubeval"
	"github.com/garethr/kubeval/log"
)

// RootCmd represents the the command to run when kubeval is run
var RootCmd = &cobra.Command{
	Use:   "kubeval <file> [file...]",
	Short: "Validate a Kubernetes YAML file against the relevant schema",
	Long:  `Validate a Kubernetes YAML file against the relevant schema`,
	Run: func(cmd *cobra.Command, args []string) {
		if Version {
			printVersion()
			os.Exit(0)
		}
		success := true
		windowsStdinIssue := false
		stat, err := os.Stdin.Stat()
		if err != nil {
			// Stat() will return an error on Windows in both Powershell and
			// console until go1.9 when nothing is passed on stdin.
			// See https://github.com/golang/go/issues/14853.
			if runtime.GOOS != "windows" {
				log.Error(err)
				os.Exit(1)
			} else {
				windowsStdinIssue = true
			}
		}
		// We detect whether we have anything on stdin to process
		if !windowsStdinIssue && ((stat.Mode() & os.ModeCharDevice) == 0) {
			var buffer bytes.Buffer
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				buffer.WriteString(scanner.Text() + "\n")
			}
			results, err := kubeval.Validate(buffer.Bytes(), "stdin")
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}
			success = logResults(results, success)
		} else {
			if len(args) < 1 {
				log.Error("You must pass at least one file as an argument")
				os.Exit(1)
			}
			for _, fileName := range args {
				filePath, _ := filepath.Abs(fileName)
				fileContents, err := ioutil.ReadFile(filePath)
				if err != nil {
					log.Error("Could not open file", fileName)
					os.Exit(1)
				}
				results, err := kubeval.Validate(fileContents, fileName)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
				success = logResults(results, success)
			}
		}
		if !success {
			os.Exit(1)
		}
	},
}

func logResults(results []kubeval.ValidationResult, success bool) bool {
	for _, result := range results {
		if len(result.Errors) > 0 {
			success = false
			log.Warn("The document", result.FileName, "contains an invalid", result.Kind)
			for _, desc := range result.Errors {
				log.Info("--->", desc)
			}
		} else {
			log.Success("The document", result.FileName, "contains a valid", result.Kind)
		}
	}
	return success
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(-1)
	}
}

func init() {
	viper.SetEnvPrefix("KUBEVAL")
	viper.AutomaticEnv()
	RootCmd.Flags().StringVarP(&kubeval.Version, "kubernetes-version", "v", "master", "Version of Kubernetes to validate against")
	RootCmd.Flags().StringVarP(&kubeval.SchemaLocation, "schema-location", "", kubeval.DefaultSchemaLocation, "Base URL used to download schemas. Can also be specified with the environment variable KUBEVAL_SCHEMA_LOCATION")
	RootCmd.Flags().BoolVarP(&kubeval.OpenShift, "openshift", "", false, "Use OpenShift schemas instead of upstream Kubernetes")
	RootCmd.Flags().BoolVarP(&kubeval.Strict, "strict", "", false, "Disallow additional properties not in schema")
	RootCmd.Flags().BoolVarP(&Version, "version", "", false, "Display the kubeval version information and exit")
	viper.BindPFlag("schema_location", RootCmd.Flags().Lookup("schema-location"))
}
