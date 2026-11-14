package maplua

import (
	"reflect"
	"strings"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

type inner struct {
	Value int `lua:"value"`
}

type sample struct {
	Name     string            `lua:"name"`
	Count    int               `lua:"count"`
	Ratio    float64           `lua:"ratio"`
	Enabled  bool              `lua:"enabled"`
	Tags     []string          `lua:"tags"`
	Labels   map[string]string `lua:"labels"`
	Inner    inner             `lua:"inner"`
	Optional *inner            `lua:"optional"`
	List     []*inner          `lua:"list"`
	Flat     inner             `lua:",inline"`
	Skipped  string            `lua:"skipped,ignore"`
	Default  string            // Uses the field name as key
	private  string
}

func TestUnmarshal_StructWithAllFieldKinds(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	err := L.DoString(`value = {
		name = "svc", count = 3, ratio = 0.5, enabled = true,
		tags = { "a", "b" }, labels = { env = "dev" },
		inner = { value = 1 }, optional = { value = 2 }, list = { { value = 3 }, { value = 4 } },
		value = 5, skipped = "x", Default = "d",
	}`)
	if err != nil {
		t.Fatal(err)
	}

	var got sample
	if err := Unmarshal(L.GetGlobal("value"), &got); err != nil {
		t.Fatal(err)
	}

	want := sample{
		Name: "svc", Count: 3, Ratio: 0.5, Enabled: true,
		Tags: []string{"a", "b"}, Labels: map[string]string{"env": "dev"},
		Inner: inner{1}, Optional: &inner{2}, List: []*inner{{3}, {4}},
		Flat: inner{5}, Default: "d",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Unmarshal =\n%+v\nwant\n%+v", got, want)
	}
}

func TestUnmarshal_MissingFields_KeepTheirValues(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	got := sample{Name: "kept", Count: 7}
	if err := Unmarshal(L.NewTable(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != "kept" || got.Count != 7 || got.Optional != nil {
		t.Fatalf("unexpected changes: %+v", got)
	}
}

func TestUnmarshal_Interface_TableBecomesMapOrSlice(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	if err := L.DoString(`list = { 1, "a" }; dict = { k = true }`); err != nil {
		t.Fatal(err)
	}

	var list, dict any
	if err := Unmarshal(L.GetGlobal("list"), &list); err != nil {
		t.Fatal(err)
	}
	if err := Unmarshal(L.GetGlobal("dict"), &dict); err != nil {
		t.Fatal(err)
	}

	if got := *(list.(*[]any)); !reflect.DeepEqual(got, []any{1.0, "a"}) {
		t.Errorf("list = %#v", got)
	}
	if got := *(dict.(*map[string]any)); !reflect.DeepEqual(got, map[string]any{"k": true}) {
		t.Errorf("dict = %#v", got)
	}
}

func TestUnmarshal_Errors(t *testing.T) {
	type badTag struct {
		A string `lua:"a,unknown"`
	}

	tests := []struct {
		name    string
		value   lua.LValue
		dest    any
		wantErr string
	}{
		{"nil value", lua.LNil, new(string), "Lua value is nil"},
		{"nil destination", lua.LString("x"), nil, "destination is nil"},
		{"non pointer destination", lua.LString("x"), "", "must be a pointer"},
		{"string into int", lua.LString("x"), new(int), "not an number"},
		{"number into string", lua.LNumber(1), new(string), "not a string"},
		{"number into bool", lua.LNumber(1), new(bool), "not a boolean"},
		{"string into slice", lua.LString("x"), new([]string), "not a table"},
		{"unknown tag option", &lua.LTable{Metatable: lua.LNil}, new(badTag), "unknown option"},
		{"unsupported destination", lua.LString("x"), new(chan int), "not a supported type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal(tt.value, tt.dest)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected an error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestMarshalUnmarshal_RoundTrip(t *testing.T) {
	original := sample{
		Name: "svc", Count: 3, Ratio: 0.5, Enabled: true,
		Tags: []string{"a"}, Labels: map[string]string{"env": "dev"},
		Inner: inner{1}, Optional: &inner{2}, List: []*inner{{3}},
		Flat: inner{5}, Default: "d",
	}

	luaValue, err := Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	var decoded sample
	if err := Unmarshal(luaValue, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, original) {
		t.Fatalf("round trip =\n%+v\nwant\n%+v", decoded, original)
	}
}

func TestMarshal_Values(t *testing.T) {
	str := "pointer"

	tests := []struct {
		name  string
		value any
		check func(t *testing.T, v lua.LValue)
	}{
		{"string", "x", expectValue(lua.LString("x"))},
		{"pointer", &str, expectValue(lua.LString("pointer"))},
		{"bool", true, expectValue(lua.LTrue)},
		{"int", -2, expectValue(lua.LNumber(-2))},
		{"uint", uint8(2), expectValue(lua.LNumber(2))},
		{"float", 1.5, expectValue(lua.LNumber(1.5))},
		{"nil", nil, expectValue(lua.LNil)},
		{"slice", []int{1, 2}, func(t *testing.T, v lua.LValue) {
			table := v.(*lua.LTable)
			if table.Len() != 2 || table.RawGetInt(2) != lua.LNumber(2) {
				t.Errorf("slice = %v", table)
			}
		}},
		{"map with int keys", map[int]string{3: "c"}, func(t *testing.T, v lua.LValue) {
			if got := v.(*lua.LTable).RawGetInt(3); got != lua.LString("c") {
				t.Errorf("map[3] = %v", got)
			}
		}},
		{"ignored struct fields are not marshaled", sample{Skipped: "x"}, func(t *testing.T, v lua.LValue) {
			table := v.(*lua.LTable)
			if table.RawGetString("skipped") != lua.LNil || table.RawGetString("private") != lua.LNil {
				t.Errorf("ignored fields were marshaled: %v", table)
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := Marshal(tt.value)
			if err != nil {
				t.Fatal(err)
			}
			tt.check(t, v)
		})
	}

	if _, err := Marshal(make(chan int)); err == nil {
		t.Error("marshaling a channel must fail")
	}
}

func TestMarshal_Tables_WorkWithTheStandardLibrary(t *testing.T) {
	value, err := Marshal(map[string]any{"list": []any{1.0}})
	if err != nil {
		t.Fatal(err)
	}

	L := lua.NewState()
	defer L.Close()
	L.SetGlobal("value", value)

	// Tables created without a Lua state used to have an invalid meta-table, which crashed these functions
	if err := L.DoString(`assert(getmetatable(value) == nil); assert(getmetatable(value.list) == nil); assert(#value.list == 1)`); err != nil {
		t.Fatal(err)
	}
}

func expectValue(want lua.LValue) func(t *testing.T, v lua.LValue) {
	return func(t *testing.T, v lua.LValue) {
		t.Helper()
		if v != want {
			t.Errorf("got %#v, want %#v", v, want)
		}
	}
}
