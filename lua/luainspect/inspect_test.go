package luainspect

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestInspect(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	eval := func(expr string) lua.LValue {
		t.Helper()
		if err := L.DoString("value = " + expr); err != nil {
			t.Fatal(err)
		}
		return L.GetGlobal("value")
	}

	userData := L.NewUserData()
	userData.Value = lua.LString("inner")

	tests := []struct {
		name  string
		value lua.LValue
		want  string
	}{
		{"nil", lua.LNil, "<nil>"},
		{"number", lua.LNumber(1.5), "1.5"},
		{"bool", lua.LTrue, "true"},
		{"string", lua.LString("a"), `"a"`},
		{"function", eval("function() end"), "function() end"},
		{"empty table", eval("{}"), "{}"},
		{"sequence", eval(`{ 1, "a" }`), "{\n    1, \"a\"\n}"},
		{"key value", eval(`{ k = 1 }`), "{\n    k = 1\n}"},
		{"sequence and key value", eval(`{ 1, k = 2 }`), "{\n    1,\n    k = 2\n}"},
		{"nested table", eval(`{ { 1 } }`), "{\n    {\n        1\n    }\n}"},
		{"nested key value", eval(`{ k = { x = 1 } }`), "{\n    k = {\n        x = 1\n    }\n}"},
		{"user data", userData, "UserData: {\n    Value: \"inner\"\n}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Inspect(tt.value)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("Inspect =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}
