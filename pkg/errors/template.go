package errors

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const (
	// DefaultLinesBefore the number of lines before the error line that are printed.
	DefaultLinesBefore = 5
	// DefaultLinesAfter the number of lines after the error line that are printed.
	DefaultLinesAfter = 5
	// DefaultLineNumberPrefixLength is the fixed width of the code line prefix containing the line number.
	DefaultLineNumberPrefixLength = 6
)

var (
	errorLineColumnRegexp = regexp.MustCompile("(?m):([0-9]+)(:([0-9]+))?:")
	// yamlLineRegexp matches the "line N" location marker used by
	// sigs.k8s.io/yaml / go-yaml when the generated YAML fails to parse
	// (e.g. "error converting YAML to JSON: yaml: line 19: could not find expected ':'").
	yamlLineRegexp = regexp.MustCompile(`line ([0-9]+)`)
)

// TemplateError wraps a go templating error and adds more human-readable information.
type TemplateError struct {
	err                              error
	source                           *string
	output                           *string
	input                            map[string]any
	message                          string
	inputFormatter                   *TemplateInputFormatter
	sourceCodePrepend                int
	sourceCodeAppend                 int
	sourceCodeLineNumberPrefixLength int
}

// TemplateErrorBuilder creates a new TemplateError.
func TemplateErrorBuilder(err error) *TemplateError {
	return &TemplateError{
		err:                              err,
		message:                          err.Error(),
		sourceCodePrepend:                DefaultLinesBefore,
		sourceCodeAppend:                 DefaultLinesAfter,
		sourceCodeLineNumberPrefixLength: DefaultLineNumberPrefixLength,
	}
}

// WithSource adds the template source code to the error.
func (e *TemplateError) WithSource(source *string) *TemplateError {
	e.source = source
	return e
}

// WithSourceSnippetFormat adds the number of lines before and after the error line that are printed in the source code snippet.
// The lineNumberPrefixLength specifies the total length of the line number prefix, consisting of the line number, a colon, and then whitespace to align the code.
func (e *TemplateError) WithSourceSnippetFormat(linesBefore, linesAfter, lineNumberPrefixLength int) *TemplateError {
	e.sourceCodePrepend = linesBefore
	e.sourceCodeAppend = linesAfter
	e.sourceCodeLineNumberPrefixLength = lineNumberPrefixLength
	return e
}

// WithFormattedOutput adds the rendered template output to the error. Used
// when the failure occurs after successful Go-template execution (e.g. the
// generated YAML is syntactically invalid or does not match the expected
// schema) - the user cannot otherwise see the generated document, so surface
// a snippet around the failing line.
func (e *TemplateError) WithFormattedOutput(output *string) *TemplateError {
	e.output = output
	return e
}

// WithInput adds the template input with a formatter to the error.
func (e *TemplateError) WithInput(input map[string]any, inputFormatter *TemplateInputFormatter) *TemplateError {
	e.input = input
	e.inputFormatter = inputFormatter
	return e
}

// Build builds the error message.
func (e *TemplateError) Build() *TemplateError {
	builder := strings.Builder{}
	builder.WriteString(e.err.Error())

	if e.source != nil {
		builder.WriteString("\ntemplate source:\n")
		builder.WriteString(e.formatSource())
	}

	if e.output != nil {
		builder.WriteString("\ntemplated output:\n")
		builder.WriteString(e.formatOutput())
	}

	if e.input != nil && e.inputFormatter != nil {
		builder.WriteString("\ntemplate input:\n")
		builder.WriteString(e.inputFormatter.Format(e.input, "\t"))
	}

	e.message = builder.String()
	return e
}

// Error returns the error message.
func (e *TemplateError) Error() string {
	return e.message
}

// formatSource extracts the significant template source code that was the reason of the template error.
func (e *TemplateError) formatSource() string {
	line, column := extractLineColumn(e.err.Error())
	if line == 0 {
		return ""
	}
	return CreateSourceSnippet(line, column, strings.Split(*e.source, "\n"), e.sourceCodePrepend, e.sourceCodeAppend, e.sourceCodeLineNumberPrefixLength)
}

// formatOutput extracts the significant rendered output lines around the
// location reported by a downstream YAML parser.
func (e *TemplateError) formatOutput() string {
	line, column := extractLineColumn(e.err.Error())
	if line == 0 {
		return ""
	}
	return CreateSourceSnippet(line, column, strings.Split(*e.output, "\n"), e.sourceCodePrepend, e.sourceCodeAppend, e.sourceCodeLineNumberPrefixLength)
}

// extractLineColumn tries the Go-template ":N:M:" location marker first,
// then falls back to the YAML "line N" marker. Returns (0, 0) if neither
// matches.
func extractLineColumn(errStr string) (int, int) {
	if m := errorLineColumnRegexp.FindStringSubmatch(errStr); m != nil {
		line, err := strconv.Atoi(m[1])
		if err != nil {
			return 0, 0
		}
		column := 0
		if len(m) >= 4 && m[3] != "" {
			if c, err := strconv.Atoi(m[3]); err == nil {
				column = c
			}
		}
		return line, column
	}
	if m := yamlLineRegexp.FindStringSubmatch(errStr); m != nil {
		line, err := strconv.Atoi(m[1])
		if err != nil {
			return 0, 0
		}
		return line, 0
	}
	return 0, 0
}

// CreateSourceSnippet creates an excerpt of lines of source code, containing some lines before
// and after the error line.
// The error line and column will be highlighted and looks like this:
// 14:     updateStrategy: patch
// 15:
// 16:     name: test
// 17:     namespace: {{ .imports.namespace }}
//
//	ˆ≈≈≈≈≈≈≈
//
// 18:
// 19:     exportsFromManifests:
// 20:     - key: ingressClass
func CreateSourceSnippet(errorLine, errorColumn int, source []string, linesBefore, linesAfter, lineNumberPrefixLength int) string {
	var (
		sourceStartLine, sourceEndLine int
		formatted                      = strings.Builder{}
	)

	// convert to zero base index
	errorLine -= 1

	// calculate the starting line of the source code
	sourceStartLine = max(errorLine-linesBefore, 0)

	errorLine -= sourceStartLine
	source = source[sourceStartLine:]

	// calculate the ending line of the source code
	sourceEndLine = min(errorLine+linesAfter+1, len(source))

	source = source[:sourceEndLine]

	for i, line := range source {
		// for printing, the line has to be converted back to one based index
		realLine := sourceStartLine + i + 1
		realLineWidth := int(math.Log10(float64(realLine)) + 1)
		// account for the colon after the line number
		repeat := max(lineNumberPrefixLength-realLineWidth-1, 0)
		// the prefix contains the line number and some amount of whitespaces to keep the correct indentation
		prefix := fmt.Sprintf("%d:%s", realLine, strings.Repeat(" ", repeat))
		fmt.Fprintf(&formatted, "%s%s\n", prefix, line)

		if i == errorLine {
			fmt.Fprintf(&formatted, "%s\u02c6≈≈≈≈≈≈≈\n", strings.Repeat(" ", errorColumn+len(prefix)))
		}
	}

	return formatted.String()
}
