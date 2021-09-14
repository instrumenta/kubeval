package kubeval

// References:
// https://github.com/windyroad/JUnit-Schema/blob/master/JUnit.xsd
// https://llg.cubic.org/docs/junit/
// https://github.com/jstemmer/go-junit-report/blob/master/formatter/formatter.go
// https://github.com/junit-team/junit5/blob/main/platform-tests/src/test/resources/jenkins-junit.xsd

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"time"
)

type TestSuiteCollection struct {
	XMLName  xml.Name    `xml:"testsuites"`
	Name     string      `xml:"name,attr"`
	Time     float64     `xml:"time,attr"`
	Tests    int         `xml:"tests,attr"`
	Failures int         `xml:"failures,attr"`
	Disabled int         `xml:"disabled,attr"`
	Errors   int         `xml:"errors,attr"`
	Suites   []TestSuite `xml:"testsuite"`
}

type Property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type TestSuite struct {
	XMLName    xml.Name    `xml:"testsuite"`
	Properties []*Property `xml:"properties>property,omitempty"`
	Cases      []TestCase  `xml:"testcase"`
	Name       string      `xml:"name,attr"`
	Id         int         `xml:"id,attr"`
	Tests      int         `xml:"tests,attr"`
	Failures   int         `xml:"failures,attr"`
	Errors     int         `xml:"errors,attr"`
	Disabled   int         `xml:"disabled,attr"`
	Skipped    int         `xml:"skipped,attr"`
}

type TestCase struct {
	XMLName   xml.Name         `xml:"testcase"`
	Name      string           `xml:"name,attr"`
	ClassName string           `xml:"classname,attr"`
	Skipped   *TestCaseSkipped `xml:"skipped,omitempty"`
	Error     *TestCaseError   `xml:"error,omitempty"`
	Failure   []TestCaseError  `xml:"failure,omitempty"`
}

type TestCaseSkipped struct {
	Message string `xml:"message,attr"`
}

type TestCaseError struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

type junito struct {
	id                                  int
	w                                   io.Writer
	withSummary                         bool
	verbose                             bool
	suites                              map[string]*TestSuite // map filename to corresponding suite
	nValid, nInvalid, nErrors, nSkipped int
	startTime                           time.Time
}

func junitOutput(w io.Writer, withSummary bool, isStdin, verbose bool) outputManager {
	return &junito{
		id:          0,
		w:           w,
		withSummary: withSummary,
		verbose:     verbose,
		suites:      make(map[string]*TestSuite),
		nValid:      0,
		nInvalid:    0,
		nErrors:     0,
		nSkipped:    0,
		startTime:   time.Now(),
	}
}

// Put adds a result to the report.
func (o *junito) Put(result ValidationResult) error {
	var suite *TestSuite
	suite, found := o.suites[result.FileName]

	if !found {
		o.id++
		suite = &TestSuite{
			Name:  result.FileName,
			Id:    o.id,
			Tests: 0, Failures: 0, Errors: 0, Disabled: 0, Skipped: 0,
			Cases:      make([]TestCase, 0),
			Properties: make([]*Property, 0),
		}
		o.suites[result.FileName] = suite
	}

	var objectName string
	if len(result.ResourceNamespace) > 0 {
		objectName = fmt.Sprintf("%s/%s", result.ResourceNamespace, result.ResourceName)
	} else {
		objectName = result.ResourceName
	}
	typeName := fmt.Sprintf("%s@%s", result.Kind, result.APIVersion)
	testCase := TestCase{ClassName: typeName, Name: objectName}

	switch getStatus(result) {
	case statusValid:
		o.nValid++
	case statusInvalid:
		for _, err := range result.Errors {
			o.nInvalid++
			failure := TestCaseError{Message: err.Description()}
			testCase.Failure = append(testCase.Failure, failure)
		}

	//case validator.Error:
	//	o.nErrors++
	//	testCase.Error = &TestCaseError{Message: result.Err.Error()}
	case statusSkipped:
		testCase.Skipped = &TestCaseSkipped{Message: fmt.Sprintf( "%s containing a %s (%s) was not validated against a schema", result.FileName, result.Kind, result.QualifiedName())}
		o.nSkipped++
	//case validator.Empty:
	//	return nil
	}

	suite.Tests++
	suite.Cases = append(suite.Cases, testCase)

	return nil
}

// Flush outputs the results as XML
func (o *junito) Flush() error {
	runtime := time.Now().Sub(o.startTime)

	var suites = make([]TestSuite, 0)
	for _, suite := range o.suites {
		suites = append(suites, *suite)
	}

	root := TestSuiteCollection{
		Name:     "kubeval",
		Time:     runtime.Seconds(),
		Tests:    o.nValid + o.nInvalid + o.nErrors + o.nSkipped,
		Failures: o.nInvalid,
		Errors:   o.nErrors,
		Disabled: o.nSkipped,
		Suites:   suites,
	}

	// 2-space indentation
	content, err := xml.MarshalIndent(root, "", "  ")

	if err != nil {
		return err
	}

	writer := bufio.NewWriter(o.w)
	writer.Write(content)
	writer.WriteByte('\n')
	writer.Flush()

	return nil
}