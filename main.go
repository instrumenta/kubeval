package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/fatih/color"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/instrumenta/kubeval/kubeval"
	"github.com/instrumenta/kubeval/log"
)

var (
	version     = "dev"
	commit      = "none"
	date        = "unknown"
	directories = []string{}

	// forceColor tells kubeval to use colored output even if
	// stdout is not a TTY
	forceColor bool

	config = kubeval.NewDefaultConfig()
)


func validateFiles(filePaths []string, schemaCache map[string]*gojsonschema.Schema, config *kubeval.Config) ([]kubeval.ValidationResult, error){
	aggResults := []kubeval.ValidationResult{}

	for _, filePath := range filePaths {
		absFilePath, _ := filepath.Abs(filePath)
		fileContents, err := ioutil.ReadFile(absFilePath)
		if err != nil {
			log.Error(fmt.Errorf("Could not open file %v", filePath))
			earlyExit()
			continue
		}

		config.FileName = filePath
		results, err := kubeval.ValidateWithCache(fileContents, schemaCache, config)
		if err != nil {
			log.Error(err)
			earlyExit()
			continue
		}

		aggResults = append(aggResults, results...)
	}

	return aggResults, nil
}

// RootCmd represents the the command to run when kubeval is run
var RootCmd = &cobra.Command{
	Use:     "kubeval <file> [file...]",
	Short:   "Validate a Kubernetes YAML file against the relevant schema",
	Long:    `Validate a Kubernetes YAML file against the relevant schema`,
	Version: fmt.Sprintf("Version: %s\nCommit: %s\nDate: %s\n", version, commit, date),
	Run: func(cmd *cobra.Command, args []string) {
		if config.IgnoreMissingSchemas && !config.Quiet {
			log.Warn("Set to ignore missing schemas")
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
		if (len(args) < 1 || args[0] == "-") && !windowsStdinIssue && ((stat.Mode() & os.ModeCharDevice) == 0) {
			var buffer bytes.Buffer
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				buffer.WriteString(scanner.Text() + "\n")
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
			nWorkers := 4

			filesQueue := make(chan[]string)
			results := make(chan[]kubeval.ValidationResult)
			var wg sync.WaitGroup
			for i:=0; i<=nWorkers; i++ {
				wg.Add(1)
				go func() {
					for pathsBatch := range filesQueue {
						res, _ := validateFiles(pathsBatch, schemaCache, config)
						results <- res
					}
					wg.Done()
				}()
			}

			go func() {
				for resultBatch := range results {
					for _, result := range resultBatch {
						err := outputManager.Put(result)
						if err != nil {
							log.Error(err)
							os.Exit(1)
						}
					}

					// only use result of hasErrors check if `success` is currently truthy
					success = !hasErrors(resultBatch)
				}
			}()

			err := aggregateFiles(args, filesQueue, 100)
			if err != nil {
				log.Error(err)
				success = false
			}

			close(filesQueue)
			wg.Wait()
			close(results)
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

func aggregateFiles(args []string, fileBatches chan<-[]string, batchSize int) error {
	files := []string{}
	copy(files, args)

	var allErrors *multierror.Error
	for _, directory := range directories {
		err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
				files = append(files, path)
				fmt.Printf("%d\n", len(files))
				if len(files) > batchSize {
					fileBatches <-files
					files = nil
				}
			}
			return nil
		})
		if err != nil {
			allErrors = multierror.Append(allErrors, err)
		}
	}

	return allErrors.ErrorOrNil()
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
	kubeval.AddKubevalFlags(RootCmd, config)
	RootCmd.Flags().BoolVarP(&forceColor, "force-color", "", false, "Force colored output even if stdout is not a TTY")
	RootCmd.SetVersionTemplate(`{{.Version}}`)
	RootCmd.Flags().StringSliceVarP(&directories, "directories", "d", []string{}, "A comma-separated list of directories to recursively search for YAML documents")

	viper.SetEnvPrefix("KUBEVAL")
	viper.AutomaticEnv()
	viper.BindPFlag("schema_location", RootCmd.Flags().Lookup("schema-location"))
	viper.BindPFlag("filename", RootCmd.Flags().Lookup("filename"))
}

func main() {
	Execute()
}
