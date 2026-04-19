// Package maplua is a package that contains functions to map a Lua object to a Golang object and vice versa
package maplua

import (
	"fmt"
	"reflect"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// tagInfo is a structure that contains information about a Tag of a structure field
type tagInfo struct {
	Name   string // Name of the field
	Inline bool   // If the field is inline
	Ignore bool   // If the field should be ignored
}

// getTagInfo gets the TagInfo of a structure field
func getTagInfo(t reflect.StructField) (*tagInfo, error) {
	info := tagInfo{}
	tag := strings.Split(t.Tag.Get("lua"), ",")

	// Name
	info.Name = tag[0]
	if info.Name == "" {
		info.Name = t.Name
	}

	if info.Name == "-" {
		return &info, nil
	}

	// Other options
	tag = tag[1:] // Removes the name

	for _, option := range tag {
		switch option {

		case "inline":
			info.Inline = true

		case "ignore":
			info.Ignore = true

		default:
			return nil, fmt.Errorf("unknown option '%s'", option)
		}

	}

	return &info, nil
}

// Marshal encodes a Golang object to a Lua object
func Marshal(value any) (lua.LValue, error) {
	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)

	// Dereference pointers
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
		v = v.Elem()
	}

	// Handles `nil` (there is not a reflection `Kind` for nil)
	if v == reflect.ValueOf(nil) {
		return lua.LNil, nil
	}

	switch t.Kind() {
	case reflect.String:
		return lua.LString(v.String()), nil

	case reflect.Bool:
		return lua.LBool(v.Bool()), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return lua.LNumber(v.Int()), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return lua.LNumber(v.Uint()), nil

	case reflect.Float32, reflect.Float64:
		return lua.LNumber(v.Float()), nil

	case reflect.Slice, reflect.Array:
		slice, err := marshalSlice(v)
		if err != nil {
			return nil, fmt.Errorf("error marshaling slice/array: %w", err)
		}
		return slice, nil

	case reflect.Map:
		table, err := marshalMap(v)
		if err != nil {
			return nil, fmt.Errorf("error marshaling map: %w", err)
		}
		return table, nil

	case reflect.Struct:
		table, err := marshalStruct(t, v)
		if err != nil {
			return nil, fmt.Errorf("error marshaling struct: %w", err)
		}
		return table, nil

	default:
		return nil, fmt.Errorf("not implemented for type %v", t)
	}
}

// marshalSlice marshals a slice or an array to a Lua table
func marshalSlice(v reflect.Value) (*lua.LTable, error) {
	slice := lua.LTable{}
	for i := 0; i < v.Len(); i++ {
		element, err := Marshal(v.Index(i).Interface())
		if err != nil {
			return nil, fmt.Errorf("error marshaling slice/array value: %w", err)
		}
		slice.RawSetInt(i+1, element)
	}

	return &slice, nil
}

// marshalMap marshals a map to a Lua table
func marshalMap(v reflect.Value) (*lua.LTable, error) {
	table := lua.LTable{}
	for _, key := range v.MapKeys() {
		element, err := Marshal(v.MapIndex(key).Interface())
		if err != nil {
			return nil, fmt.Errorf("error marshaling map value: %w", err)
		}

		// Sets the value in the table
		switch key.Kind() {
		case reflect.String:
			table.RawSetString(key.String(), element)
		case reflect.Int:
			table.RawSetInt(int(key.Int()), element)
		case reflect.Uint:
			table.RawSetInt(int(key.Uint()), element)
		}
	}

	return &table, nil
}

// marshalStruct marshals a struct to a Lua table
func marshalStruct(t reflect.Type, v reflect.Value) (*lua.LTable, error) {
	table := lua.LTable{}
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanInterface() {
			continue
		}

		// Gets the tag info
		tagInfo, err := getTagInfo(t.Field(i))
		if err != nil {
			return nil, fmt.Errorf("error getting tag information for field %v: %w", t.Field(i), err)
		}
		if tagInfo.Ignore {
			continue
		}

		// Lua key and value
		luaKey := tagInfo.Name

		luaValue, err := Marshal(v.Field(i).Interface())
		if err != nil {
			return nil, fmt.Errorf("error marshaling structure value: %w", err)
		}

		// Should not set a nil value
		if luaValue.Type() == lua.LTNil {
			continue
		}

		// Sets the value in the table
		if tagInfo.Inline && luaValue.Type() == lua.LTTable {
			// Merge the tables
			luaValue.(*lua.LTable).ForEach(func(l1, l2 lua.LValue) {
				// If the key already exists, do not override
				if table.RawGet(l1) != lua.LNil {
					return
				}

				table.RawSet(l1, l2)
			})
		} else {
			table.RawSetString(luaKey, luaValue)
		}
	}

	return &table, nil
}
