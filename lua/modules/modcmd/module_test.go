package modcmd

import (
	"slices"
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

func TestAvailableArgs_AreStoredInTheRuntime(t *testing.T) {
	r := testruntime.New(t, nil, nil)
	var notified [][]string
	r.SetOnScriptAvailableArgsChange(func(args [][]string) { notified = args })

	L := testutil.NewLuaState(t, map[string]testutil.ModuleLoader{"cmd": testruntime.Configure(r, New())})
	err := L.DoString(`
		local cmd = require("cmd")
		cmd.set_available_args({ { "dev" }, { "prod", 2 } })
		local args = cmd.get_available_args()
		assert(#args == 2 and args[2][1] == "prod" and args[2][2] == "2")
	`)
	if err != nil {
		t.Fatal(err)
	}

	want := [][]string{{"dev"}, {"prod", "2"}}
	if !slices.EqualFunc(r.GetScriptAvailableArgs(), want, slices.Equal) || !slices.EqualFunc(notified, want, slices.Equal) {
		t.Fatalf("available args = %v, notified = %v, want %v", r.GetScriptAvailableArgs(), notified, want)
	}
}

func TestSetAvailableArgs_InvalidInput_RaisesArgumentError(t *testing.T) {
	r := testruntime.New(t, nil, nil)
	L := testutil.NewLuaState(t, map[string]testutil.ModuleLoader{"cmd": testruntime.Configure(r, New())})

	for _, code := range []string{
		`require("cmd").set_available_args("dev")`,
		`require("cmd").set_available_args({ "dev" })`,
	} {
		if err := L.DoString(code); err == nil {
			t.Errorf("%s: expected an error", code)
		}
	}
}
