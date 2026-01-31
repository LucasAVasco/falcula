// Package luaclass is a class generator for Lua
package luaclass

import (
	"errors"
	"fmt"

	"github.com/LucasAVasco/falcula/lua/luatable"

	lua "github.com/yuin/gopher-lua"
)

var ErrClassNameEmpty = errors.New("class name cannot be empty")
var ErrClassConstructorEmpty = errors.New("constructor cannot be empty")

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

// New creates a new class table
func New(L *lua.LState, info *Info, onError func(err error)) (*lua.LTable, error) {
	if onError == nil {
		onError = func(err error) { fmt.Printf("error in class context: %v\n", err) }
	}

	if info.Name == "" {
		return nil, ErrClassNameEmpty
	}

	// Meta-class
	class := L.NewTable()
	L.SetField(class, "__index", class)

	// Constructor
	if info.Constructor == nil {
		return nil, ErrClassConstructorEmpty
	}

	L.SetField(class, "new", L.NewFunction(func(L *lua.LState) int {
		class := L.ToTable(1)
		newObj := L.NewTable()
		L.SetMetatable(newObj, class)

		err := info.Constructor(L, newObj)
		if err != nil {
			onError(fmt.Errorf("error creating object: %w", err))
			L.Panic(L)
			return 0
		}
		L.Push(newObj)

		return 1
	}))

	// List constructor
	L.SetField(class, "new_list", L.NewFunction(func(L *lua.LState) int {
		argsList := L.ToTable(2)

		newObjectList := L.NewTable()
		argsList.ForEach(func(_, value lua.LValue) {
			// Arguments
			argsTable := value.(*lua.LTable) // The 'value' is a table with all the arguments of a 'new' command
			args := luatable.GetValuesFromLuaTable(argsTable)

			// Calls the 'new' method
			ret, err := CallMethod(L, class, "new", 1, args...)
			if err != nil {
				onError(fmt.Errorf("error creating new object for class '%s': %w", info.Name, err))
				L.Panic(L)
			}

			// Adds the new object to the list
			newObjectList.Append(ret[0])
		})

		// Returns the list
		L.Push(newObjectList)

		return 1
	}))

	// Methods
	if info.Methods != nil {
		L.SetFuncs(class, info.Methods)
	}

	return class, nil
}

// GetAttribute gets a object attribute. Must not be used outside a class method
func GetAttribute(L *lua.LState, name string) any {
	table := L.ToTable(1)
	userData := table.RawGetString(name).(*lua.LUserData)
	return userData.Value
}

// SetAttribute sets a object attribute. Must not be used outside a class method
func SetAttribute(L *lua.LState, object *lua.LTable, field string, value any) {
	userData := L.NewUserData()
	userData.Value = value
	L.SetField(object, field, userData)
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

	// List of return values
	retList := make([]lua.LValue, numReturn)
	for i := range numReturn {
		retList[i] = L.Get(-1)
		L.Pop(1)
	}

	return retList, nil
}
