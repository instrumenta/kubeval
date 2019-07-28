package kubeval

import (
	"github.com/instrumenta/kubeval/log"
)

// TODO (brendanryan) move these structs to `/log` once we have removed the potential
// circular dependancy between this package and `/log`

// OutputManager controls how results of the `ccheck` evaluation will be recorded
// and reported to the end user.
type OutputManager interface {
	Put(r ValidationResult) error
	Flush() error
}

// STDOutputManager reports `ccheck` results to stdout.
type STDOutputManager struct {
}

// NewDefaultStdOutputManager instantiates a new instance of STDOutputManager using
// the default logger.
func NewSTDOutputManager() *STDOutputManager {
	return &STDOutputManager{}
}

func (s *STDOutputManager) Put(result ValidationResult) error {

	if len(result.Errors) > 0 {
		log.Warn("The file", result.FileName, "contains an invalid", result.Kind)
		for _, desc := range result.Errors {
			log.Info("--->", desc)
		}
	} else if result.Kind == "" {
		log.Success("The file", result.FileName, "contains an empty YAML document")
	} else if !result.ValidatedAgainstSchema {
		log.Warn("The file", result.FileName, "containing a", result.Kind, "was not validated against a schema")
	} else {
		log.Success("The file", result.FileName, "contains a valid", result.Kind)
	}

	return nil
}

func (s *STDOutputManager) Flush() error {
	// no op
	return nil
}
