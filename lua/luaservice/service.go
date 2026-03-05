// Package luaservice is a package that contains functions to work with services and providers in Lua
package luaservice

import (
	"fmt"

	"github.com/LucasAVasco/falcula/provider/base"
	lua "github.com/yuin/gopher-lua"
)

// ParseBaseServiceOpts parses the base service options from a Lua table. If the table does not exist, returns nil
func ParseBaseServiceOpts(_ *lua.LState, table lua.LValue) (*base.ServiceOpts, error) {
	if table == lua.LNil {
		return nil, nil
	}

	// Result
	opts := base.ServiceOpts{}

	if table.Type() != lua.LTTable {
		return nil, fmt.Errorf("the options list must be a table, got %T", table)
	}

	// 'start_disabled'
	if startDisabled := table.(*lua.LTable).RawGetString("start_disabled"); startDisabled.Type() == lua.LTBool {
		startDisabled := bool(startDisabled.(lua.LBool))
		opts.StartDisabled = &startDisabled
	}

	return &opts, nil
}
