package maplua

import (
	"fmt"
	"reflect"

	"github.com/LucasAVasco/falcula/lua/luatable"
	lua "github.com/yuin/gopher-lua"
)

// Unmarshal decodes a Lua object to a Golang object. The destination must be a pointer to a value
func Unmarshal(luaValue lua.LValue, dest any) error {
	// Validation
	if dest == nil {
		return fmt.Errorf("destination is nil")
	}

	if luaValue.Type() == lua.LTNil {
		return fmt.Errorf("Lua value is nil")
	}

	destType := reflect.TypeOf(dest)
	destValue := reflect.ValueOf(dest)

	// Destination must be a pointer, but we use its dereferenced value
	if destType.Kind() != reflect.Pointer {
		return fmt.Errorf("destination must be a pointer")
	}
	destType = destType.Elem()
	destValue = destValue.Elem()

	switch destType.Kind() {

	case reflect.Interface:
		err := unmarshalInterface(luaValue, destValue)
		if err != nil {
			return fmt.Errorf("error unmarshaling interface: %w", err)
		}

	case reflect.String:
		if luaValue.Type() != lua.LTString {
			return fmt.Errorf("Lua value is not a string")
		}
		destValue.SetString(luaValue.String())

	case reflect.Bool:
		if luaValue.Type() != lua.LTBool {
			return fmt.Errorf("Lua value is not a boolean")
		}
		destValue.SetBool(bool(luaValue.(lua.LBool)))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if luaValue.Type() != lua.LTNumber {
			return fmt.Errorf("Lua value is not an number")
		}
		destValue.SetInt(int64(luaValue.(lua.LNumber)))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if luaValue.Type() != lua.LTNumber {
			return fmt.Errorf("Lua value is not an number")
		}
		destValue.SetUint(uint64(luaValue.(lua.LNumber)))

	case reflect.Float32, reflect.Float64:
		if luaValue.Type() != lua.LTNumber {
			return fmt.Errorf("Lua value is not a number")
		}
		destValue.SetFloat(float64(luaValue.(lua.LNumber)))

	case reflect.Slice, reflect.Array:
		if luaValue.Type() != lua.LTTable {
			return fmt.Errorf("Lua value is not a table")
		}

		err := unmarshalSlice(luaValue.(*lua.LTable), destType, destValue)
		if err != nil {
			return fmt.Errorf("error unmarshaling slice: %w", err)
		}

	case reflect.Map:
		if luaValue.Type() != lua.LTTable {
			return fmt.Errorf("Lua value is not a table")
		}

		err := unmarshalMap(luaValue.(*lua.LTable), destType, destValue)
		if err != nil {
			return fmt.Errorf("error unmarshaling map: %w", err)
		}

	case reflect.Struct:
		if luaValue.Type() != lua.LTTable {
			return fmt.Errorf("Lua value is not a table")
		}

		err := unmarshalStruct(luaValue.(*lua.LTable), destType, destValue)
		if err != nil {
			return fmt.Errorf("error unmarshaling struct: %w", err)
		}

	default:
		return fmt.Errorf("destination is not a supported type: %v", destType)
	}

	return nil
}

// unmarshalInterface unmarshals a Lua value to an interface (any)
func unmarshalInterface(luaValue lua.LValue, destValue reflect.Value) error {
	switch luaValue.Type() {
	case lua.LTNil:
		destValue.Set(reflect.Zero(reflect.TypeFor[any]()))
	case lua.LTBool:
		destValue.Set(reflect.ValueOf(bool(luaValue.(lua.LBool))))
	case lua.LTNumber:
		destValue.Set(reflect.ValueOf(float64(luaValue.(lua.LNumber))))
	case lua.LTString:
		destValue.Set(reflect.ValueOf(luaValue.String()))
	case lua.LTFunction:
		destValue.Set(reflect.ValueOf(luaValue.(*lua.LFunction)))
	case lua.LTUserData:
		destValue.Set(reflect.ValueOf(luaValue.(*lua.LUserData)))
	case lua.LTThread:
		destValue.Set(reflect.ValueOf(luaValue))
	case lua.LTTable:
		luaValue := luaValue.(*lua.LTable)
		if luaValue.Len() == 0 {
			mapValue := reflect.New(reflect.MapOf(reflect.TypeFor[string](), reflect.TypeFor[any]()))
			destValue.Set(mapValue)
			err := Unmarshal(luaValue, mapValue.Interface())
			if err != nil {
				return fmt.Errorf("error unmarshaling table in map of type []any: %w", err)
			}
		} else {
			sliceValue := reflect.New(reflect.SliceOf(reflect.TypeFor[any]()))
			destValue.Set(sliceValue)
			err := Unmarshal(luaValue, sliceValue.Interface())
			if err != nil {
				return fmt.Errorf("error unmarshaling table in slice of type []any: %w", err)
			}
		}
	case lua.LTChannel:
		destValue.Set(reflect.ValueOf(luaValue.(*lua.LChannel)))
	}

	return nil
}

// unmarshalSlice unmarshals a Lua table to a slice
func unmarshalSlice(luaValue *lua.LTable, destType reflect.Type, destValue reflect.Value) error {
	isSliceOfPointers := destType.Elem().Kind() == reflect.Pointer
	if luaValue.Len() == 0 {
		reflectValue := reflect.MakeSlice(destType, 0, luaValue.Len())
		destValue.Set(reflectValue)
		return nil
	}
	sliceValue := destValue
	for i := range luaValue.Len() {
		// Create a element. Its is a pointer to the literal, container or structure
		var element reflect.Value
		if isSliceOfPointers {
			element = reflect.New(destType.Elem().Elem())
		} else {
			element = reflect.New(destType.Elem())
		}

		// Unmarshal the Lua value in the element
		err := Unmarshal(luaValue.RawGetInt(i+1), element.Interface())
		if err != nil {
			return fmt.Errorf("error unmarshaling slice element '%#v' in index '%v': %w", luaValue.RawGetInt(i+1), i, err)
		}

		// We need to dereference the element in order to save in a slice that elements are not pointers
		if !isSliceOfPointers {
			element = element.Elem()
		}

		sliceValue = reflect.Append(sliceValue, element)
	}
	destValue.Set(sliceValue)

	return nil
}

// unmarshalMap unmarshals a Lua table to a map
func unmarshalMap(luaValue *lua.LTable, destType reflect.Type, destValue reflect.Value) error {
	keyType := destType.Key()
	isMapOfPointers := destType.Elem().Kind() == reflect.Pointer
	if destValue.IsNil() {
		destValue.Set(reflect.MakeMap(reflect.MapOf(keyType, destType.Elem())))
	}
	err := luatable.ForEach(luaValue, func(key, value lua.LValue) error {
		// Create a element. Its is a pointer to the literal, container or structure
		var element reflect.Value
		if isMapOfPointers {
			element = reflect.New(destType.Elem().Elem())
		} else {
			element = reflect.New(destType.Elem())
		}

		// Unmarshal
		err := Unmarshal(value, element.Interface())
		if err != nil {
			return fmt.Errorf("error unmarshaling map value '%#v' in key '%v': %w", value, key, err)
		}

		// We need to dereference the element in order to save in a map that elements are not pointers
		if !isMapOfPointers {
			element = element.Elem()
		}

		// Converts the key to the map key type
		keyValue := reflect.ValueOf(key)
		if !keyValue.CanConvert(keyType) {
			return fmt.Errorf("cannot convert key '%v' to map key type '%v'", key, keyType)
		}
		keyValue = keyValue.Convert(keyType)

		// Adds the element to the map
		destValue.SetMapIndex(keyValue, element)
		return nil
	})
	if err != nil {
		return fmt.Errorf("error iterating over map values: %w", err)
	}

	return nil
}

// unmarshalStruct unmarshals a Lua table to a structure
func unmarshalStruct(luaValue *lua.LTable, destType reflect.Type, destValue reflect.Value) error {
	luaTable := luaValue
	for i := 0; i < destType.NumField(); i++ {
		if !destValue.Field(i).CanInterface() {
			continue
		}

		// Gets the tag info
		tagInfo, err := getTagInfo(destType.Field(i))
		if err != nil {
			return fmt.Errorf("error getting tag information for field %v: %w", destType.Field(i), err)
		}
		if tagInfo.Ignore {
			continue
		}

		// Lua key and value
		luaKey := tagInfo.Name

		var luaValue lua.LValue = luaTable
		if !tagInfo.Inline {
			luaValue = luaTable.RawGetString(luaKey)
		}

		// Ignores if the value is nil
		if luaValue == nil {
			continue
		}
		if luaValue.Type() == lua.LTNil {
			continue
		}

		// Field content (dereferenced if it is a pointer)
		field := destValue.Field(i)
		if field.Kind() == reflect.Pointer {
			// Creates a new pointer if it is nil
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}

			// Filed must be the element instead of a pointer to it
			field = field.Elem()
		}

		// Unmarshal the Lua value in the field
		err = Unmarshal(luaValue, field.Addr().Interface())
		if err != nil {
			return fmt.Errorf("error unmarshaling struct value '%v' in key '%v': %w", luaValue, luaKey, err)
		}
	}

	return nil
}
