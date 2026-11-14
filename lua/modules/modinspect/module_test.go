package modinspect

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestInspectFunction_ReturnsTheInspectedValue(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	L.PreloadModule("inspect", func(L *lua.LState) int { return LoadFunction("inspect", L) })

	err := L.DoString(`
		local inspect = require("inspect")
		assert(inspect({ k = 1 }) == "{\n    k = 1\n}", inspect({ k = 1 }))
		assert(inspect("x") == '"x"')
	`)
	if err != nil {
		t.Fatal(err)
	}
}
