package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/fatih/color"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/instrumenta/kubeval/kubeval"
	"github.com/instrumenta/kubeval/log"
)

var (
	version                 = "dev"
	commit                  = "none"
	date                    = "unknown"
	directories             = []string{}
	ignoredPathPatterns = []string{}

	// forceColor tells kubeval to use colored output even if
	// stdout is not a TTY
	forceColor bool

	config = kubeval.NewDefaultConfig()
)

// RootCmd represents the the command to run when kubeval is run
var RootCmd = &cobra.Command{
	Short:   "Validate a Kubernetes YAML file against the relevant schema",
	Long:    `Validate a Kubernetes YAML file against the relevant schema`,
	Version: fmt.Sprintf("Version: %s\nCommit: %s\nDate: %s\n", version, commit, date),
	Run: func(cmd *cobra.Command, args []string) {
		if config.IgnoreMissingSchemas && !config.Quiet {
			log.Warn("Set to ignore missing schemas")
		}

		// This is not particularly secure but we highlight that with the name of
		// the config item. It would be good to also support a configurable set of
		// trusted certificate authorities as in the `--certificate-authority`
		// kubectl option.
		if config.InsecureSkipTLSVerify {
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		success := true
		windowsStdinIssue := false
		outputManager := kubeval.GetOutputManager(config.OutputFormat)

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
		// Assert that colors will definitely be used if requested
		if forceColor {
			color.NoColor = false
		}
		// We detect whether we have anything on stdin to process if we have no arguments
		// or if the argument is a -
		notty := (stat.Mode() & os.ModeCharDevice) == 0
		noFileOrDirArgs := (len(args) < 1 || args[0] == "-") && len(directories) < 1
		if noFileOrDirArgs && !windowsStdinIssue && notty {
			buffer := new(bytes.Buffer)
			_, err := io.Copy(buffer, os.Stdin)
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}
			schemaCache := kubeval.NewSchemaCache()
			config.FileName = viper.GetString("filename")
			results, err := kubeval.ValidateWithCache(buffer.Bytes(), schemaCache, config)
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}
			success = !hasErrors(results)

			for _, r := range results {
				err = outputManager.Put(r)
				if err != nil {
					log.Error(err)
					os.Exit(1)
				}
			}
		} else {
			if len(args) < 1 && len(directories) < 1 {
				log.Error(errors.New("You must pass at least one file as an argument, or at least one directory to the directories flag"))
				os.Exit(1)
			}
			schemaCache := kubeval.NewSchemaCache()
			files, err := aggregateFiles(args)
			if err != nil {
				log.Error(err)
				success = false
			}

			var aggResults []kubeval.ValidationResult
			for _, fileName := range files {
				filePath, _ := filepath.Abs(fileName)
				fileContents, err := ioutil.ReadFile(filePath)
				if err != nil {
					log.Error(fmt.Errorf("Could not open file %v", fileName))
					earlyExit()
					success = false
					continue
				}
				config.FileName = fileName
				results, err := kubeval.ValidateWithCache(fileContents, schemaCache, config)
				if err != nil {
					log.Error(err)
					earlyExit()
					success = false
					continue
				}

				for _, r := range results {
					err := outputManager.Put(r)
					if err != nil {
						log.Error(err)
						os.Exit(1)
					}
				}

				aggResults = append(aggResults, results...)
			}

			// only use result of hasErrors check if `success` is currently truthy
			success = success && !hasErrors(aggResults)
		}

		// flush any final logs which may be sitting in the buffer
		err = outputManager.Flush()
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		if !success {
			os.Exit(1)
		}
	},
}

// hasErrors returns truthy if any of the provided results
// contain errors.
func hasErrors(res []kubeval.ValidationResult) bool {
	for _, r := range res {
		if len(r.Errors) > 0 {
			return true
		}
	}
	return false
}

// isIgnored returns whether the specified filename should be ignored.
func isIgnored(path string) (bool, error) {
	for _, p := range ignoredPathPatterns {
		m, err := regexp.MatchString(p, path)
		if err != nil {
			return false, err
		}
		if m {
			return true, nil
		}
	}
	return false, nil
}

func aggregateFiles(args []string) ([]string, error) {
	files := make([]string, len(args))
	copy(files, args)

	var allErrors *multierror.Error
	for _, directory := range directories {
		err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			ignored, err := isIgnored(path)
			if err != nil {
				return err
			}
			if !info.IsDir() && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) && !ignored {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			allErrors = multierror.Append(allErrors, err)
		}
	}

	return files, allErrors.ErrorOrNil()
}

func earlyExit() {
	if config.ExitOnError {
		os.Exit(1)
	}
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
	rootCmdName := filepath.Base(os.Args[0])
	if strings.HasPrefix(rootCmdName, "kubectl-") {
		rootCmdName = strings.Replace(rootCmdName, "-", " ", 1)
	}
	RootCmd.Use = fmt.Sprintf("%s <file> [file...]", rootCmdName)
	kubeval.AddKubevalFlags(RootCmd, config)
	RootCmd.Flags().BoolVarP(&forceColor, "force-color", "", false, "Force colored output even if stdout is not a TTY")
	RootCmd.SetVersionTemplate(`{{.Version}}`)
	RootCmd.Flags().StringSliceVarP(&directories, "directories", "d", []string{}, "A comma-separated list of directories to recursively search for YAML documents")
	RootCmd.Flags().StringSliceVarP(&ignoredPathPatterns, "ignored-path-patterns", "i", []string{}, "A comma-separated list of regular expressions specifying paths to ignore")
	RootCmd.Flags().StringSliceVarP(&ignoredPathPatterns, "ignored-filename-patterns", "", []string{}, "An alias for ignored-path-patterns")
	
	viper.SetEnvPrefix("KUBEVAL")
	viper.AutomaticEnv()
	viper.BindPFlag("schema_location", RootCmd.Flags().Lookup("schema-location"))
	viper.BindPFlag("filename", RootCmd.Flags().Lookup("filename"))
}

func main() {
	Execute()
}
