package luatable

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// ExtendBehavior is the behavior of the Extend function
type ExtendBehavior uint

const (
	// ExtendBehaviorKeep keeps the value from the left most table
	ExtendBehaviorKeep ExtendBehavior = iota

	// ExtendBehaviorForce keeps the value from the right most table
	ExtendBehaviorForce

	// ExtendBehaviorError returns an error if there are duplicated keys
	ExtendBehaviorError

	// ExtendBehaviorInvalid is an invalid behavior
	ExtendBehaviorInvalid
)

// String returns the string representation of the behavior
func (behavior ExtendBehavior) String() string {
	switch behavior {
	case ExtendBehaviorKeep:
		return "keep"
	case ExtendBehaviorForce:
		return "force"
	case ExtendBehaviorError:
		return "error"
	default:
		return "invalid"
	}
}

// ExtendBehaviorFromString converts a string (ExtendBehaviorKeep representation) to an ExtendBehavior
func ExtendBehaviorFromString(s string) ExtendBehavior {
	switch s {
	case "keep":
		return ExtendBehaviorKeep
	case "force":
		return ExtendBehaviorForce
	case "error":
		return ExtendBehaviorError
	default:
		return ExtendBehaviorInvalid
	}
}

// Extend merge several tables into the first one. Only the first table is modified. The behavior defines how to handle duplicated keys
func Extend(behavior ExtendBehavior, destTable *lua.LTable, tables ...*lua.LTable) error {
	for i, table := range tables {
		err := ForEach(table, func(key, value lua.LValue) error {
			// Sequences are appended
			if key.Type() == lua.LTNumber {
				destTable.Append(value)
				return nil
			}

			// Key-value pairs are merged
			if behavior == ExtendBehaviorForce {
				destTable.RawSet(key, value)
				return nil
			}

			if destTable.RawGet(key) == lua.LNil {
				destTable.RawSet(key, value)
				return nil
			}

			if behavior == ExtendBehaviorError {
				return fmt.Errorf("duplicated key '%s'", key)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("error extending table at index %d: %w", i, err)
		}
	}

	return nil
}

// DeepExtend merge several tables into the first one recursively. Only the first table is modified. The behavior defines how to handle
// duplicated keys
func DeepExtend(behavior ExtendBehavior, destTable *lua.LTable, tables ...*lua.LTable) error {
	for i, table := range tables {
		err := ForEach(table, func(key, value lua.LValue) error {
			// Sequences are appended
			if key.Type() == lua.LTNumber {
				destTable.Append(value)
				return nil
			}

			// If both values are tables, then deep extend
			tValue := destTable.RawGet(key)
			if tValue.Type() == lua.LTTable && value.Type() == lua.LTTable {
				err := DeepExtend(behavior, tValue.(*lua.LTable), value.(*lua.LTable))
				if err != nil {
					return fmt.Errorf("error deep extending table: %w", err)
				}
				return nil
			}

			// Both values are not tables, need to override
			if behavior == ExtendBehaviorForce {
				destTable.RawSet(key, value)
				return nil
			}

			if destTable.RawGet(key) == lua.LNil {
				destTable.RawSet(key, value)
				return nil
			}

			if behavior == ExtendBehaviorError {
				return fmt.Errorf("duplicated key '%s'", key)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("error extending table at index %d: %w", i, err)
		}
	}

	return nil
}
