// Package luainspect is a package that inspects a Lua value and pretty prints it
package luainspect

import (
	"fmt"
	"io"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// enhancedWriter is writer with improvements to write inspected Lua values
type enhancedWriter struct {
	writer io.Writer
}

// Write writes a message. It returns only an error instead of returning the number of bytes written and an error
func (w *enhancedWriter) Write(message []byte) error {
	n, err := w.writer.Write(message)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	if n != len(message) {
		return fmt.Errorf("error writing message: wrote %d bytes instead of %d", n, len(message))
	}

	return nil
}

// WriteString is a convenience wrapper for Write that writes a string
func (w *enhancedWriter) WriteString(message string) error {
	return w.Write([]byte(message))
}

// WriteIndent writes a number of indent bytes. indentLevel starts in 0
func (w *enhancedWriter) WriteIndent(indentLevel uint) error {
	identBytes := []byte("    ")

	for range indentLevel {
		err := w.Write(identBytes)
		if err != nil {
			return fmt.Errorf("error writing indent: %w", err)
		}
	}

	return nil
}

// inspect inspects a Lua value and writes it to the writer.
//
// firstIndentLevel is the indent level of the first line. indentLevel is the indent level of the rest of the lines. Both start in 0. Inner
// values are indented by indentLevel + 1.
func inspect(writer *enhancedWriter, firstIndentLevel uint, indentLevel uint, value lua.LValue) error {
	err := writer.WriteIndent(firstIndentLevel)
	if err != nil {
		return fmt.Errorf("error writing first indent: %w", err)
	}

	switch value.Type() {
	case lua.LTNil:
		err := writer.WriteString("<nil>")
		if err != nil {
			return fmt.Errorf("error writing nil: %w", err)
		}

	case lua.LTFunction:
		err := writer.WriteString("function() end")
		if err != nil {
			return fmt.Errorf("error writing function: %w", err)
		}

	case lua.LTUserData:
		value := value.(*lua.LUserData)

		err := inspectUserData(writer, indentLevel, value)
		if err != nil {
			return fmt.Errorf("error inspecting user data: %w", err)
		}

	case lua.LTTable:
		value := value.(*lua.LTable)
		err := inspectLuaTable(writer, indentLevel, value)
		if err != nil {
			return fmt.Errorf("error inspecting table: %w", err)
		}

	case lua.LTBool, lua.LTNumber:
		err := writer.WriteString(value.String())
		if err != nil {
			return fmt.Errorf("error writing value: %w", err)
		}

	case lua.LTString, lua.LTThread, lua.LTChannel:
		err := writer.WriteString("\"" + value.String() + "\"")
		if err != nil {
			return fmt.Errorf("error writing value: %w", err)
		}

	default:
		return fmt.Errorf("unknown type: %s", value.Type())
	}

	return nil
}

// Inspect inspects a Lua value and writes it to the writer
func InspectWriter(value lua.LValue, writer io.Writer) error {
	err := inspect(&enhancedWriter{writer: writer}, 0, 0, value)
	if err != nil {
		return fmt.Errorf("error inspecting value: %w", err)
	}
	return nil
}

// Inspect inspects a Lua value and returns it as a string
func Inspect(value lua.LValue) (string, error) {
	buffer := strings.Builder{}
	err := inspect(&enhancedWriter{writer: &buffer}, 0, 0, value)
	if err != nil {
		return "", fmt.Errorf("error inspecting value: %w", err)
	}
	return buffer.String(), nil
}
