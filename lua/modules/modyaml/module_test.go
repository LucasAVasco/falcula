package modyaml

import (
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/lua/luaerror"
)

func TestEncodeDecode(t *testing.T) {
	luaerror.ConfigureAbortOnError(false)
	t.Cleanup(func() { luaerror.ConfigureAbortOnError(true) })

	L := testutil.NewLuaState(t, map[string]testutil.ModuleLoader{"yaml": New()})
	err := L.DoString(`
		local yaml = require("yaml")

		local value = yaml.decode("name: svc\nports:\n  - 80\n  - 443\nempty: null\n")
		assert(value.name == "svc" and value.ports[2] == 443 and value.empty == nil)

		-- Round trip
		local decoded = yaml.decode(yaml.encode({ key = "value", list = { "a", "b" } }))
		assert(decoded.key == "value" and decoded.list[2] == "b")

		local result, err = yaml.decode("key: [unclosed")
		assert(result == nil and err:find("error decoding YAML"), tostring(err))
	`)
	if err != nil {
		t.Fatal(err)
	}
}
