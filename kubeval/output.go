package kubeval

import (
	"bytes"
	"encoding/json"
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
)

func validOutputs() []string {
	return []string{
		outputSTD,
		outputJSON,
	}
}

func GetOutputManager(outFmt string) outputManager {
	switch outFmt {
	case outputSTD:
		return newSTDOutputManager()
	case outputJSON:
		return newDefaultJSONOutputManager()
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
		kLog.Warn("The file", result.FileName, "contains an invalid", result.Kind)
		for _, desc := range result.Errors {
			kLog.Info("--->", desc)
		}
	} else if result.Kind == "" {
		kLog.Success("The file", result.FileName, "contains an empty YAML document")
	} else if !result.ValidatedAgainstSchema {
		kLog.Warn("The file", result.FileName, "containing a", result.Kind, "was not validated against a schema")
	} else {
		kLog.Success("The file", result.FileName, "contains a valid", result.Kind)
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

type jsonEvalResult struct {
	Filename string   `json:"filename"`
	Kind     string   `json:"kind"`
	Status   status   `json:"status"`
	Errors   []string `json:"errors"`
}

// jsonOutputManager reports `ccheck` results to `stdout` as a json array..
type jsonOutputManager struct {
	logger *log.Logger

	data []jsonEvalResult
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

	j.data = append(j.data, jsonEvalResult{
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
