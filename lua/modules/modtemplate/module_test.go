package modtemplate

import (
	"path/filepath"
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/internal/testutil/testruntime"
	"github.com/LucasAVasco/falcula/lua/luaerror"

	lua "github.com/yuin/gopher-lua"
)

func newTemplateState(t *testing.T) *lua.LState {
	t.Helper()
	luaerror.ConfigureAbortOnError(false)
	t.Cleanup(func() { luaerror.ConfigureAbortOnError(true) })

	r := testruntime.New(t, nil, nil)
	return testutil.NewLuaState(t, map[string]testutil.ModuleLoader{"template": testruntime.Configure(r, New())})
}

func TestParseString_DataTypes(t *testing.T) {
	L := newTemplateState(t)

	err := L.DoString(`
		local template = require("template")
		local function check(tpl, data, want)
			local got, err = template.parse_string(tpl, data)
			assert(err == nil, tostring(err))
			assert(got == want, string.format("%q: got %q, want %q", tpl, got, want))
		end

		check("{{ .name }}={{ .port }}", { name = "web", port = 8080 }, "web=8080")
		check("{{ if .enabled }}on{{ else }}off{{ end }}", { enabled = false }, "off")
		check("{{ index . 1 }}", { "a", "b" }, "b")
		check("{{ range . }}[{{ . }}]{{ end }}", { "x", "y" }, "[x][y]")
		check("{{ .list | len }}", { list = { 1, 2, 3 } }, "3")
		check("{{ .nested.key }}", { nested = { key = "deep" } }, "deep")
		check("{{ .missing }}", { other = 1 }, "<no value>") -- NOTE: an empty table is converted to a list
		check("{{ upper . }}", "sprig", "SPRIG") -- Sprig functions are available
		check("# !{{ .value }}", { value = "sanitized" }, "   sanitized")
		check("{{ .value }}! comment", { value = "x" }, "x         ")
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseString_InvalidTemplate_ReturnsError(t *testing.T) {
	L := newTemplateState(t)

	err := L.DoString(`
		local template = require("template")
		local result, err = template.parse_string("{{ .a ", {})
		assert(result == nil and err:find("error parsing template"), tostring(err))

		result, err = template.parse_string("{{ index . 5 }}", { "a" })
		assert(result == nil and err:find("error executing template"), tostring(err))
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseFile_AndParseAndSaveFile(t *testing.T) {
	L := newTemplateState(t)
	dir := t.TempDir()
	src := testutil.WriteFile(t, filepath.Join(dir, "src.tpl"), "Hello {{ .name }}.")
	dest := filepath.Join(dir, "dest.txt")

	L.SetGlobal("src", lua.LString(src))
	L.SetGlobal("dest", lua.LString(dest))
	L.SetGlobal("missing", lua.LString(filepath.Join(dir, "missing.tpl")))

	err := L.DoString(`
		local template = require("template")
		local out, perr = template.parse_file(src, { name = "file" })
		assert(out == "Hello file.", string.format("%q %s", tostring(out), tostring(perr)))
		assert(template.parse_and_save_file(src, dest, { name = "saved" }) == "Hello saved.")

		local result, err = template.parse_file(missing, {})
		assert(result == nil and err:find("error reading file"), tostring(err))
	`)
	if err != nil {
		t.Fatal(err)
	}

	if got := testutil.ReadFile(t, filepath.Join(dir, "dest.txt")); got != "Hello saved." {
		t.Fatalf("saved file = %q", got)
	}
}

func TestTemplateClass_CustomFunctions(t *testing.T) {
	L := newTemplateState(t)

	err := L.DoString(`
		local template = require("template")
		local tpl = template.Template:new()
		tpl:set_func("greet", function(name, punctuation) return "Hi " .. name .. punctuation end)

		assert(tpl:parse_string('{{ greet .name "!" }}', { name = "Ana" }) == "Hi Ana!")

		-- Functions are per object
		local result, err = template.Template:new():parse_string('{{ greet "x" "y" }}', {})
		assert(result == nil and err ~= nil)
	`)
	if err != nil {
		t.Fatal(err)
	}
}
