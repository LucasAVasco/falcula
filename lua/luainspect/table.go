package luainspect

import (
	"fmt"

	"github.com/LucasAVasco/falcula/lua/luatable"
	lua "github.com/yuin/gopher-lua"
)

// tableInspectorState is the state of the table inspector
type tableInspectorState uint

const (
	// tableInspectorEmpty is the state of the table inspector when it is doing nothing (first state)
	tableInspectorEmpty tableInspectorState = iota

	// tableInspectorSequence is the state of the table inspector when it is writing a sequence that each element has only one line.
	// Example: numbers and strings
	tableInspectorSequence

	// tableInspectorMultiLineSequence is the state of the table inspector when it is writing a multi-line sequence (sequence that each
	// element has more than one line). Example: tables, user data
	tableInspectorMultiLineSequence

	// tableInspectorKeyValue is the state of the table inspector when it is writing key-value elements (the non-number keys of a table)
	tableInspectorKeyValue
)

// tableInspector is used to inspect a Lua table
type tableInspector struct {
	table       *lua.LTable
	writer      *enhancedWriter
	indentLevel uint
	state       tableInspectorState
}

// Handle is the main function of the table inspector. It inspects the table and writes it to the writer
func (t *tableInspector) Handle() error {
	// Empty table
	key, value := t.table.Next(lua.LNil)
	if key.Type() == lua.LTNil && value.Type() == lua.LTNil {
		err := t.writer.WriteString("{}")
		if err != nil {
			return fmt.Errorf("error writing empty table: %w", err)
		}
		return nil
	}

	// Non-empty table
	err := t.writer.WriteString("{\n")
	if err != nil {
		return fmt.Errorf("error writing first line of table: %w", err)
	}

	err = luatable.ForEach(t.table, func(key, value lua.LValue) error {
		if key.Type() == lua.LTNumber {
			err = t.handleSequenceElement(key, value)
			if err != nil {
				return fmt.Errorf("error inspecting sequence element [key: %v, value: %v]: %w", key, value, err)
			}
		} else if key.Type() == lua.LTString {
			err = t.handleKeyValueElement(key, value)
			if err != nil {
				return fmt.Errorf("error inspecting key-value element [key: %v, value: %v]: %w", key, value, err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error iterating over table elements: %w", err)
	}

	// Closing table
	err = t.writer.WriteString("\n")
	if err != nil {
		return fmt.Errorf("error writing new line: %w", err)
	}

	err = t.writer.WriteIndent(t.indentLevel)
	if err != nil {
		return fmt.Errorf("error writing indent: %w", err)
	}

	err = t.writer.WriteString("}")
	if err != nil {
		return fmt.Errorf("error writing last line of table (close brace): %w", err)
	}

	return nil
}

// handleSequenceElement is used to inspect a sequence element (key-pair that the key is a number)
func (t *tableInspector) handleSequenceElement(_, value lua.LValue) error {
	multiLineValue := value.Type() == lua.LTTable || value.Type() == lua.LTUserData

	switch t.state {
	case tableInspectorEmpty:
		// Starts a new sequence
		err := t.writer.WriteIndent(t.indentLevel + 1)
		if err != nil {
			return fmt.Errorf("error writing first indent of sequence: %w", err)
		}

	case tableInspectorSequence:
		if multiLineValue {
			// Starts a new line
			err := t.writer.WriteString(",\n")
			if err != nil {
				return fmt.Errorf("error writing element separator of new multi-line sequence: %w", err)
			}

			err = t.writer.WriteIndent(t.indentLevel + 1)
			if err != nil {
				return fmt.Errorf("error writing indent after element separator of new multi-line sequence: %w", err)
			}
		} else {
			// Element separator
			err := t.writer.WriteString(", ")
			if err != nil {
				return fmt.Errorf("error writing element separator of sequence: %w", err)
			}
		}

	case tableInspectorMultiLineSequence:
		// Element separator
		err := t.writer.WriteString(",\n")
		if err != nil {
			return fmt.Errorf("error writing element separator of multi-line sequence: %w", err)
		}

		// Starts a new line
		err = t.writer.WriteIndent(t.indentLevel + 1)
		if err != nil {
			return fmt.Errorf("error writing indent after element separator of multi-line sequence: %w", err)
		}

	case tableInspectorKeyValue:
		// Closes the last key-value pair
		err := t.writer.WriteString(",\n")
		if err != nil {
			return fmt.Errorf("error writing element separator of sequence after key-value pair: %w", err)
		}

		err = t.writer.WriteIndent(t.indentLevel + 1)
		if err != nil {
			return fmt.Errorf("error writing indent after element separator of sequence after key-value pair: %w", err)
		}
	}

	// Next state
	if multiLineValue {
		t.state = tableInspectorMultiLineSequence
	} else {
		t.state = tableInspectorSequence
	}

	// Inspects the value
	err := inspect(t.writer, 0, t.indentLevel+1, value)
	if err != nil {
		return fmt.Errorf("error inspecting value: %w", err)
	}

	return nil
}

// handleKeyValueElement is used to inspect a key-value element (the non-number keys of a table)
func (t *tableInspector) handleKeyValueElement(key, value lua.LValue) error {
	switch t.state {
	case tableInspectorEmpty:
	case tableInspectorSequence, tableInspectorMultiLineSequence:
		// Closes the sequence
		err := t.writer.WriteString(",\n")
		if err != nil {
			return fmt.Errorf("error closing sequence: %w", err)
		}
	case tableInspectorKeyValue:
		// Element separator
		err := t.writer.WriteString(",\n")
		if err != nil {
			return fmt.Errorf("error writing element separator: %w", err)
		}
	}

	t.state = tableInspectorKeyValue

	// Writes the key
	err := t.writer.WriteIndent(t.indentLevel + 1)
	if err != nil {
		return fmt.Errorf("error writing indent: %w", err)
	}

	err = t.writer.WriteString(key.String() + " = ")
	if err != nil {
		return fmt.Errorf("error writing key: %w", err)
	}

	err = inspect(t.writer, 0, t.indentLevel+1, value)
	if err != nil {
		return fmt.Errorf("error inspecting value: %w", err)
	}

	return nil
}

// inspectLuaTable inspects a Lua table and writes it to the writer
func inspectLuaTable(writer *enhancedWriter, indentLevel uint, table *lua.LTable) error {
	tableInspector := &tableInspector{
		table:       table,
		writer:      writer,
		indentLevel: indentLevel,
		state:       tableInspectorEmpty,
	}
	return tableInspector.Handle()
}
