package kubeval

import (
	"bytes"
	"log"
	"testing"

	"github.com/xeipuuv/gojsonschema"

	"github.com/stretchr/testify/assert"
)

func newResultError(msg string) gojsonschema.ResultError {
	r := &gojsonschema.ResultErrorFields{}

	r.SetContext(gojsonschema.NewJsonContext("error", nil))
	r.SetDescription(msg)

	return r
}

func newResultErrors(msgs []string) []gojsonschema.ResultError {
	var res []gojsonschema.ResultError
	for _, m := range msgs {
		res = append(res, newResultError(m))
	}
	return res
}

func Test_jsonOutputManager_put(t *testing.T) {
	type args struct {
		vr ValidationResult
	}

	tests := []struct {
		msg    string
		args   args
		exp    string
		expErr error
	}{
		{
			msg: "empty input",
			args: args{
				vr: ValidationResult{},
			},
			exp: `[
	{
		"filename": "",
		"kind": "",
		"status": "skipped",
		"errors": []
	}
]
`,
		},
		{
			msg: "file with no errors",
			args: args{
				vr: ValidationResult{
					FileName:               "deployment.yaml",
					Kind:                   "deployment",
					ValidatedAgainstSchema: true,
					Errors:                 nil,
				},
			},
			exp: `[
	{
		"filename": "deployment.yaml",
		"kind": "deployment",
		"status": "valid",
		"errors": []
	}
]
`,
		},
		{
			msg: "file with errors",
			args: args{
				vr: ValidationResult{
					FileName:               "service.yaml",
					Kind:                   "service",
					ValidatedAgainstSchema: true,
					Errors: newResultErrors([]string{
						"i am a error",
						"i am another error",
					}),
				},
			},
			exp: `[
	{
		"filename": "service.yaml",
		"kind": "service",
		"status": "invalid",
		"errors": [
			"error: i am a error",
			"error: i am another error"
		]
	}
]
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := newJSONOutputManager(log.New(buf, "", 0))

			// record results
			err := s.Put(tt.args.vr)
			if err != nil {
				assert.Equal(t, tt.expErr, err)
			}

			// flush final buffer
			err = s.Flush()
			if err != nil {
				assert.Equal(t, tt.expErr, err)
			}

			assert.Equal(t, tt.exp, buf.String())
		})
	}
}

func Test_tapOutputManager_put(t *testing.T) {
	type args struct {
		vr ValidationResult
	}

	tests := []struct {
		msg    string
		args   args
		exp    string
		expErr error
	}{
		{
			msg: "file with no errors",
			args: args{
				vr: ValidationResult{
					FileName:               "deployment.yaml",
					Kind:                   "Deployment",
					ValidatedAgainstSchema: true,
					Errors:                 nil,
				},
			},
			exp: `1..1
ok 1 - deployment.yaml (Deployment)
`,
		},
		{
			msg: "file with errors",
			args: args{
				vr: ValidationResult{
					FileName:               "service.yaml",
					Kind:                   "Service",
					ValidatedAgainstSchema: true,
					Errors: newResultErrors([]string{
						"i am a error",
						"i am another error",
					}),
				},
			},
			exp: `1..2
not ok 1 - service.yaml (Service) - error: i am a error
not ok 2 - service.yaml (Service) - error: i am another error
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := newTAPOutputManager(log.New(buf, "", 0))

			// record results
			err := s.Put(tt.args.vr)
			if err != nil {
				assert.Equal(t, tt.expErr, err)
			}

			// flush final buffer
			err = s.Flush()
			if err != nil {
				assert.Equal(t, tt.expErr, err)
			}

			assert.Equal(t, tt.exp, buf.String())
		})
	}
}
