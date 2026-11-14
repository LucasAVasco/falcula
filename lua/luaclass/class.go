// Package luaclass is a class generator for Lua
package luaclass

import (
	"errors"
	"fmt"

	"github.com/LucasAVasco/falcula/lua/luadata"
	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/lua/luatable"

	lua "github.com/yuin/gopher-lua"
)

var ErrClassNameEmpty = errors.New("class name cannot be empty")

type Method = lua.LGFunction

type Info struct {
	Name string // Must be provided

	// Class method that generates a new object. `newObj` is the new object returned by the constructor. It is automatically created and
	// configured to use the class table as meta-table. You do not need to do it manually
	//
	// The constructor must be provided
	Constructor func(L *lua.LState, newObj *lua.LTable) error

	// Class methods. Automatically added to the class, you do not need to do it manually. Optional value.
	Methods map[string]Method
}

// Register creates a new class table (see `New`) and sets it in the `info.Name` field of the module table. Returns the class table
func Register(L *lua.LState, mod *lua.LTable, info *Info) (*lua.LTable, error) {
	class, err := New(L, info)
	if err != nil {
		return nil, fmt.Errorf("error creating class '%s': %w", info.Name, err)
	}

	L.SetField(mod, info.Name, class)
	return class, nil
}

// New creates a new class table
func New(L *lua.LState, info *Info) (*lua.LTable, error) {
	if info.Name == "" {
		return nil, ErrClassNameEmpty
	}

	// Meta-class
	class := L.NewTable()
	L.SetField(class, "__index", class)

	// Methods
	if info.Methods != nil {
		L.SetFuncs(class, info.Methods)
	}

	// Constructor
	if info.Constructor == nil {
		return class, nil
	}

	L.SetField(class, "new", L.NewFunction(func(L *lua.LState) int {
		class := L.ToTable(1)
		newObj := CreateInstance(L, class)

		err := info.Constructor(L, newObj)
		if err != nil {
			return luaerror.Push(L, 1, fmt.Errorf("error creating object of class '%s': %w", info.Name, err))
		}
		L.Push(newObj)

		return 1
	}))

	// List constructor
	L.SetField(class, "new_list", L.NewFunction(func(L *lua.LState) int {
		argsList := L.CheckTable(2)

		newObjectList := L.NewTable()
		err := luatable.ForEach(argsList, func(_, value lua.LValue) error {
			// Arguments (the 'value' is a table with all the arguments of a 'new' command)
			argsTable, ok := value.(*lua.LTable)
			if !ok {
				return fmt.Errorf("the arguments of each object must be a table, but got %s", value.Type())
			}
			args := luatable.GetValuesFromLuaTable(argsTable)

			// Calls the 'new' method
			ret, err := CallMethod(L, class, "new", 1, args...)
			if err != nil {
				return err
			}

			// Adds the new object to the list
			newObjectList.Append(ret[0])
			return nil
		})
		if err != nil {
			return luaerror.Push(L, 1, fmt.Errorf("error creating new object for class '%s': %w", info.Name, err))
		}

		// Returns the list
		L.Push(newObjectList)

		return 1
	}))

	return class, nil
}

// CreateInstance creates a new object of the class
func CreateInstance(L *lua.LState, class *lua.LTable) *lua.LTable {
	newObj := L.NewTable()
	L.SetMetatable(newObj, class)
	return newObj
}

// SetMethods adds methods to the class. Overrides old methods if they have the same name.
func SetMethods(L *lua.LState, class *lua.LTable, methods map[string]Method) {
	L.SetFuncs(class, methods)
}

// GetAttribute gets a object attribute. Must not be used outside a class method. Raises a Lua error if the method was not called on an
// object with this attribute (e.g. `obj.method()` instead of `obj:method()`)
func GetAttribute(L *lua.LState, name string) any {
	table, ok := L.Get(1).(*lua.LTable)
	if !ok {
		L.ArgError(1, "object expected (call methods with ':')")
	}

	userData, ok := table.RawGetString(name).(*lua.LUserData)
	if !ok {
		L.ArgError(1, fmt.Sprintf("object without the '%s' attribute (call methods with ':')", name))
	}

	return userData.Value
}

// GetAttributeOf gets a object attribute of the type T. Must not be used outside a class method. Raises a Lua error if the attribute does
// not exist or has a different type (the method was called on an object of another class)
func GetAttributeOf[T any](L *lua.LState, name string) T {
	value, ok := GetAttribute(L, name).(T)
	if !ok {
		L.ArgError(1, fmt.Sprintf("the '%s' attribute of the object has an unexpected type", name))
	}

	return value
}

// SetAttribute sets a object attribute. Must not be used outside a class method
func SetAttribute(L *lua.LState, object *lua.LTable, field string, value any) {
	L.SetField(object, field, luadata.NewUserData(L, value))
}

// CallMethod calls a object method
func CallMethod(L *lua.LState, object *lua.LTable, method string, numReturn int, args ...lua.LValue) ([]lua.LValue, error) {
	// Method function
	function := L.GetField(object, method)
	if function.Type() == lua.LTNil {
		return nil, fmt.Errorf("method '%s' not found", method)
	}
	L.Push(function)

	// Pushes the 'self' argument
	L.Push(object)

	// Makes all arguments as the arguments to the function
	for _, arg := range args {
		L.Push(arg)
	}

	// Calling the function
	err := L.PCall(len(args)+1, numReturn, nil) // 'numArgs' has a '+1' because of the 'self' argument
	if err != nil {
		return nil, fmt.Errorf("error calling method '%s': %w", method, err)
	}

	// List of return values (in the same order that the method returns them)
	retList := make([]lua.LValue, numReturn)
	for i := range numReturn {
		retList[i] = L.Get(i - numReturn)
	}
	L.Pop(numReturn)

	return retList, nil
}

// SetInstance sets the instance of the class when called inside a method. Must not be used outside a method
func SetInstance(L *lua.LState, luaInstance *lua.LTable, goInstance any) {
	SetAttribute(L, luaInstance, "_class_instance", goInstance)
}

// GetInstance gets the instance of the class when called inside a method. Must not be used outside a method.
func GetInstance[T any](L *lua.LState) T {
	return GetAttributeOf[T](L, "_class_instance")
}
