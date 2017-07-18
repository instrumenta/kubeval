package cmd

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/garethr/kubeval/version"
)

var Version bool

var versionTemplate = `Version:      {{.BuildVersion}}
Git commit:   {{.BuildSHA}}
Built:        {{.BuildTime}}
Go version:   {{.GoVersion}}
OS/Arch:      {{.Os}}/{{.Arch}}`

func printVersion() {
	renderedVersionOutput, _ := renderVersionTemplate()
	fmt.Println(renderedVersionOutput)
}

func renderVersionTemplate() (string, error) {
	versionTemplate, err := template.New("version").Parse(versionTemplate)
	if err != nil {
		return "", err
	}
	var versionOutputBuffer bytes.Buffer
	err = versionTemplate.Execute(&versionOutputBuffer, version.Version)
	if err != nil {
		return "", err
	}
	return versionOutputBuffer.String(), nil
}
