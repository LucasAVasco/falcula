package modtemplate

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/LucasAVasco/falcula/sanitizer"

	"github.com/Masterminds/sprig/v3"
)

// parseString parses a string as a Go template and executes it with the given data. Returns the result
func parseString(str string, data any, funcMap template.FuncMap) (string, error) {
	// Sanitizes
	str, err := sanitizer.SanitizeTemplate(str, "")
	if err != nil {
		return "", fmt.Errorf("error sanitizing string: %w", err)
	}

	// New template
	tpl := template.New("main").Funcs(sprig.FuncMap())
	if funcMap != nil {
		tpl = tpl.Funcs(funcMap)
	}

	// Parses
	tpl, err = tpl.Parse(str)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}

	// Executes
	writer := &strings.Builder{}
	err = tpl.Execute(writer, data)
	if err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	return writer.String(), nil
}

// parseFile parses a file as a Go template and executes it with the given data. Returns the result
func parseFile(srcFile string, data any, funcMap template.FuncMap) (string, error) {
	// Reads file
	source, err := os.ReadFile(srcFile)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	// Parses
	output, err := parseString(string(source), data, funcMap)
	if err != nil {
		return "", fmt.Errorf("error parsing string: %w", err)
	}

	return string(output), nil
}

func parseAndSaveFile(srcFile string, destFile string, data any, funcMap template.FuncMap) (string, error) {
	// Parses
	output, err := parseFile(srcFile, data, funcMap)
	if err != nil {
		return "", fmt.Errorf("error parsing file: %w", err)
	}

	// Writes
	err = os.WriteFile(destFile, []byte(output), 0644)
	if err != nil {
		return "", fmt.Errorf("error writing file: %w", err)
	}

	return output, nil
}
