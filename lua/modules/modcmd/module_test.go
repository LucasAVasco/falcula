package modcmd

import (
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/internal/testutil/testruntime"
)

func TestArgs_FollowTheCurrentScriptArguments(t *testing.T) {
	r := testruntime.New(t, nil, nil)
	m := testruntime.Configure(r, New())
	L := testutil.NewLuaState(t, map[string]testutil.ModuleLoader{"cmd": m})

	m.SetCurrentScriptArgs([]string{"a", "b", "c"})
	if err := L.DoString(`args = require("cmd").args; assert(#args == 3 and args[3] == "c")`); err != nil {
		t.Fatal(err)
	}

	// The same table is updated in place (the old arguments are removed)
	m.SetCurrentScriptArgs([]string{"x"})
	if err := L.DoString(`assert(#args == 1 and args[1] == "x" and args[2] == nil)`); err != nil {
		t.Fatal(err)
	}
}

func TestGetAvailableArgs_ReturnsTheTaskArgs(t *testing.T) {
	r := testruntime.New(t, nil, nil)
	r.SetScriptAvailableArgs([][]string{{"dev"}, {"prod", "--verbose"}}) // The `args` option of the task

	L := testutil.NewLuaState(t, map[string]testutil.ModuleLoader{"cmd": testruntime.Configure(r, New())})
	err := L.DoString(`
		local cmd = require("cmd")
		local args = cmd.get_available_args()
		assert(#args == 2 and args[1][1] == "dev" and args[2][1] == "prod" and args[2][2] == "--verbose")
		assert(cmd.set_available_args == nil, "set_available_args was replaced by the 'args' option of the task")
	`)
	if err != nil {
		t.Fatal(err)
	}
}
