package luaerror

import (
	"errors"
	"strings"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newStateWithFailingFunction(t *testing.T) *lua.LState {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(L.Close)

	L.SetGlobal("fail", L.NewFunction(func(L *lua.LState) int {
		return Push(L, 2, errors.New("boom"))
	}))
	return L
}

func TestPush_AbortDisabled_ReturnsNilsAndTheMessage(t *testing.T) {
	ConfigureAbortOnError(false)
	t.Cleanup(func() { ConfigureAbortOnError(true) })

	L := newStateWithFailingFunction(t)
	err := L.DoString(`
		local a, b, err = fail()
		assert(select("#", fail()) == 3)
		assert(a == nil and b == nil and err == "boom", tostring(err))
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPush_AbortEnabled_RaisesLuaError(t *testing.T) {
	ConfigureAbortOnError(true)

	L := newStateWithFailingFunction(t)
	err := L.DoString(`fail()`)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected a Lua error with the message, got %v", err)
	}
}
