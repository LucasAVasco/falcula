package modjson

import (
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/lua/luaerror"
)

func TestEncodeDecode(t *testing.T) {
	luaerror.ConfigureAbortOnError(false)
	t.Cleanup(func() { luaerror.ConfigureAbortOnError(true) })

	L := testutil.NewLuaState(t, map[string]testutil.ModuleLoader{"json": New()})
	err := L.DoString(`
		local json = require("json")

		-- Encoding (map keys are sorted by encoding/json)
		assert(json.encode({ b = 1, a = { 1, "x", true } }) == '{"a":[1,"x",true],"b":1}', json.encode({ b = 1, a = { 1, "x", true } }))
		assert(json.encode("text") == '"text"')

		-- Decoding
		local value = json.decode('{"list":[1,2],"nested":{"ok":true},"missing":null}')
		assert(value.list[2] == 2 and value.nested.ok == true and value.missing == nil)

		-- Errors
		local result, err = json.decode("{invalid")
		assert(result == nil and err:find("error decoding JSON"), tostring(err))
		result, err = json.encode(function() end)
		assert(result == nil and err ~= nil)
	`)
	if err != nil {
		t.Fatal(err)
	}
}
