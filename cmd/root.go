package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/garethr/kubeval/kubeval"
	"github.com/garethr/kubeval/log"
)

// RootCmd represents the the command to run when kubeval is run
var RootCmd = &cobra.Command{
	Use:   "kubeval <file> [file...]",
	Short: "Validate a Kubernetes YAML file against the relevant schema",
	Long:  `Validate a Kubernetes YAML file against the relevant schema`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Error("You must pass at least one file as an argument")
			os.Exit(1)
		}
		success := true
		for _, fileName := range args {
			filePath, _ := filepath.Abs(fileName)
			yamlFile, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Error("Could not open file", fileName)
			}
			valid := kubeval.Validate(yamlFile, fileName)
			if success {
				success = valid
			}
		}
		if !success {
			os.Exit(1)
		}
	},
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
	RootCmd.Flags().StringVarP(&kubeval.Version, "kubernetes-version", "v", "master", "Version of Kubernetes to validate against")
	RootCmd.Flags().BoolVarP(&kubeval.OpenShift, "openshift", "", false, "Use OpenShift schemas instead of upstream Kubernetes")
}
