package modprocess

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestStringCommandIsShellCommand(t *testing.T) {
	cmd, err := getProcessCommandFromLuaValue(lua.LString("echo hi"))
	if err != nil {
		t.Fatal(err)
	}
	if !cmd.Shell || len(cmd.Command) != 1 || cmd.Command[0] != "echo hi" {
		t.Fatalf("unexpected command: %+v", cmd)
	}
}
