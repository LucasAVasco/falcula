// Package sanitizer contains functions to sanitize strings (include Go templates)
package sanitizer

import "strings"

// getIndent returns indentation of a line. Does not compute the indentation if override is not empty, just returns it
func getIndent(str string, override string) string {
	if override != "" {
		return override
	}

	if str[0] == '\t' {
		return "\t"
	}

	return " "
}

// Removes all text around the templates if it starts with '!{{'
func SanitizeTemplate(str string, indent string) (string, error) {
	output := strings.Builder{}

	for i, line := range strings.Split(str, "\n") {
		// Replaces everything before '!{{' by indentation
		index := strings.Index(line, "{{")
		if index > 0 {
			if line[index-1] == '!' {
				line = strings.Repeat(getIndent(line, indent), index) + line[index:]
			}
		}

		// Removes everything after '}}!' by spaces
		index = strings.LastIndex(line, "}}")
		if index > 0 && len(line) > index+2 {
			index += 2
			if line[index] == '!' {
				line = line[:index] + strings.Repeat(" ", len(line)-index)
			}
		}

		// Writes the line
		if i > 0 {
			output.WriteString("\n")
		}
		output.WriteString(line)
	}

	return output.String(), nil
}
