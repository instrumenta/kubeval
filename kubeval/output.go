package kubeval

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	kLog "github.com/instrumenta/kubeval/log"
)

// TODO (brendanryan) move these structs to `/log` once we have removed the potential
// circular dependancy between this package and `/log`

// outputManager controls how results of the `kubeval` evaluation will be recorded
// and reported to the end user.
// This interface is kept private to ensure all implementations are closed within
// this package.
type outputManager interface {
	Put(r ValidationResult) error
	Flush() error
}

const (
	outputSTD  = "stdout"
	outputJSON = "json"
	outputTAP  = "tap"
)

func validOutputs() []string {
	return []string{
		outputSTD,
		outputJSON,
		outputTAP,
	}
}

func GetOutputManager(outFmt string) outputManager {
	switch outFmt {
	case outputSTD:
		return newSTDOutputManager()
	case outputJSON:
		return newDefaultJSONOutputManager()
	case outputTAP:
		return newDefaultTAPOutputManager()
	default:
		return newSTDOutputManager()
	}
}

// STDOutputManager reports `kubeval` results to stdout.
type STDOutputManager struct {
}

// newSTDOutputManager instantiates a new instance of STDOutputManager.
func newSTDOutputManager() *STDOutputManager {
	return &STDOutputManager{}
}

func (s *STDOutputManager) Put(result ValidationResult) error {
	if len(result.Errors) > 0 {
		for _, desc := range result.Errors {
			kLog.Warn(result.FileName, "contains an invalid", result.Kind, fmt.Sprintf("(%s)", result.QualifiedName()), "-", desc.String())
		}
	} else if result.Kind == "" {
		kLog.Success(result.FileName, "contains an empty YAML document")
	} else if !result.ValidatedAgainstSchema {
		kLog.Warn(result.FileName, "containing a", result.Kind, fmt.Sprintf("(%s)", result.QualifiedName()), "was not validated against a schema")
	} else {
		kLog.Success(result.FileName, "contains a valid", result.Kind, fmt.Sprintf("(%s)", result.QualifiedName()))
	}

	return nil
}

func (s *STDOutputManager) Flush() error {
	// no op
	return nil
}

type status string

const (
	statusInvalid = "invalid"
	statusValid   = "valid"
	statusSkipped = "skipped"
)

type dataEvalResult struct {
	Filename string   `json:"filename"`
	Kind     string   `json:"kind"`
	Status   status   `json:"status"`
	Errors   []string `json:"errors"`
}

// jsonOutputManager reports `ccheck` results to `stdout` as a json array..
type jsonOutputManager struct {
	logger *log.Logger

	data []dataEvalResult
}

func newDefaultJSONOutputManager() *jsonOutputManager {
	return newJSONOutputManager(log.New(os.Stdout, "", 0))
}

func newJSONOutputManager(l *log.Logger) *jsonOutputManager {
	return &jsonOutputManager{
		logger: l,
	}
}

func getStatus(r ValidationResult) status {
	if r.Kind == "" {
		return statusSkipped
	}

	if !r.ValidatedAgainstSchema {
		return statusSkipped
	}

	if len(r.Errors) > 0 {
		return statusInvalid
	}

	return statusValid
}

func (j *jsonOutputManager) Put(r ValidationResult) error {
	// stringify gojsonschema errors
	// use a pre-allocated slice to ensure the json will have an
	// empty array in the "zero" case
	errs := make([]string, 0, len(r.Errors))
	for _, e := range r.Errors {
		errs = append(errs, e.String())
	}

	j.data = append(j.data, dataEvalResult{
		Filename: r.FileName,
		Kind:     r.Kind,
		Status:   getStatus(r),
		Errors:   errs,
	})

	return nil
}

func (j *jsonOutputManager) Flush() error {
	b, err := json.Marshal(j.data)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		return err
	}

	j.logger.Print(out.String())
	return nil
}

// tapOutputManager reports `conftest` results to stdout.
type tapOutputManager struct {
	logger *log.Logger

	data []dataEvalResult
}

// newDefaultTapOutManager instantiates a new instance of tapOutputManager
// using the default logger.
func newDefaultTAPOutputManager() *tapOutputManager {
	return newTAPOutputManager(log.New(os.Stdout, "", 0))
}

// newTapOutputManager constructs an instance of tapOutputManager given a
// logger instance.
func newTAPOutputManager(l *log.Logger) *tapOutputManager {
	return &tapOutputManager{
		logger: l,
	}
}

func (j *tapOutputManager) Put(r ValidationResult) error {
	errs := make([]string, 0, len(r.Errors))
	for _, e := range r.Errors {
		errs = append(errs, e.String())
	}

	j.data = append(j.data, dataEvalResult{
		Filename: r.FileName,
		Kind:     r.Kind,
		Status:   getStatus(r),
		Errors:   errs,
	})

	return nil
}

func (j *tapOutputManager) Flush() error {
	issues := len(j.data)
	if issues > 0 {
		total := 0
		for _, r := range j.data {
			if len(r.Errors) > 0 {
				total = total + len(r.Errors)
			} else {
				total = total + 1
			}
		}
		j.logger.Print(fmt.Sprintf("1..%d", total))
		count := 0
		for _, r := range j.data {
			count = count + 1
			var kindMarker string
			if r.Kind == "" {
				kindMarker = ""
			} else {
				kindMarker = fmt.Sprintf(" (%s)", r.Kind)
			}
			if r.Status == "valid" {
				j.logger.Print("ok ", count, " - ", r.Filename, kindMarker)
			} else if r.Status == "skipped" {
				j.logger.Print("ok ", count, " - ", r.Filename, kindMarker, " # SKIP")
			} else if r.Status == "invalid" {
				for i, e := range r.Errors {
					j.logger.Print("not ok ", count, " - ", r.Filename, kindMarker, " - ", e)

					// We have to skip adding 1 if it's the last error
					if len(r.Errors) != i+1 {
						count = count + 1
					}
				}
			}
		}
	}
	return nil
}
