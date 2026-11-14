package luatable

import (
	"errors"
	"slices"
	"strings"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// newTables evaluates a Lua expression that returns a list of tables
func newTables(t *testing.T, expr string) []*lua.LTable {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(L.Close)

	if err := L.DoString("result = " + expr); err != nil {
		t.Fatal(err)
	}

	tables := []*lua.LTable{}
	L.GetGlobal("result").(*lua.LTable).ForEach(func(_, v lua.LValue) {
		tables = append(tables, v.(*lua.LTable))
	})
	return tables
}

func TestExtend_Behaviors(t *testing.T) {
	tests := []struct {
		behavior ExtendBehavior
		wantA    lua.LValue
		wantErr  bool
	}{
		{ExtendBehaviorKeep, lua.LNumber(1), false},
		{ExtendBehaviorForce, lua.LNumber(2), false},
		{ExtendBehaviorError, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.behavior.String(), func(t *testing.T) {
			tables := newTables(t, `{ { a = 1, 10 }, { a = 2, b = 3, 20 } }`)
			dest := tables[0]

			err := Extend(tt.behavior, dest, tables[1])
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), "duplicated key 'a'") {
					t.Fatalf("expected a duplicated key error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if got := dest.RawGetString("a"); got != tt.wantA {
				t.Errorf("a = %v, want %v", got, tt.wantA)
			}
			if got := dest.RawGetString("b"); got != lua.LNumber(3) {
				t.Errorf("new keys must be added, b = %v", got)
			}
			if dest.Len() != 2 || dest.RawGetInt(2) != lua.LNumber(20) {
				t.Errorf("sequence items must be appended, got len %d", dest.Len())
			}
		})
	}
}

func TestExtend_IsShallow(t *testing.T) {
	tables := newTables(t, `{ { n = { x = 1 } }, { n = { y = 2 } } }`)
	if err := Extend(ExtendBehaviorForce, tables[0], tables[1]); err != nil {
		t.Fatal(err)
	}

	n := tables[0].RawGetString("n").(*lua.LTable)
	if n.RawGetString("x") != lua.LNil || n.RawGetString("y") != lua.LNumber(2) {
		t.Fatalf("Extend must replace nested tables, got %v", n)
	}
}

func TestDeepExtend_MergesNestedTables(t *testing.T) {
	tables := newTables(t, `{ { n = { x = 1, deep = { a = 1 } } }, { n = { y = 2, deep = { b = 2 } } }, { n = { z = 3 } } }`)
	if err := DeepExtend(ExtendBehaviorKeep, tables[0], tables[1], tables[2]); err != nil {
		t.Fatal(err)
	}

	n := tables[0].RawGetString("n").(*lua.LTable)
	deep := n.RawGetString("deep").(*lua.LTable)
	for key, want := range map[string]lua.LValue{"x": lua.LNumber(1), "y": lua.LNumber(2), "z": lua.LNumber(3)} {
		if got := n.RawGetString(key); got != want {
			t.Errorf("n.%s = %v, want %v", key, got, want)
		}
	}
	if deep.RawGetString("a") != lua.LNumber(1) || deep.RawGetString("b") != lua.LNumber(2) {
		t.Errorf("deep = %v", deep)
	}
}

func TestDeepExtend_NestedConflict_ReturnsErrorWithTableIndex(t *testing.T) {
	tables := newTables(t, `{ { n = { x = 1 } }, {}, { n = { x = 2 } } }`)

	err := DeepExtend(ExtendBehaviorError, tables[0], tables[1], tables[2])
	if err == nil || !strings.Contains(err.Error(), "index 1") || !strings.Contains(err.Error(), "duplicated key 'x'") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtendBehavior_StringConversion(t *testing.T) {
	for _, b := range []ExtendBehavior{ExtendBehaviorKeep, ExtendBehaviorForce, ExtendBehaviorError} {
		if got := ExtendBehaviorFromString(b.String()); got != b {
			t.Errorf("ExtendBehaviorFromString(%q) = %v", b.String(), got)
		}
	}
	if ExtendBehaviorFromString("other") != ExtendBehaviorInvalid || ExtendBehaviorInvalid.String() != "invalid" {
		t.Error("unknown behaviors must be invalid")
	}
}

func TestStringConversions(t *testing.T) {
	tables := newTables(t, `{ { "a", 2, name = "x" } }`)

	if got := GetStringsFromLuaTableThatKeyIsNumber(tables[0]); !slices.Equal(got, []string{"a", "2"}) {
		t.Errorf("GetStringsFromLuaTableThatKeyIsNumber = %v", got)
	}

	got := GetStringsFromLuaTable(tables[0])
	slices.Sort(got)
	if !slices.Equal(got, []string{"2", "a", "x"}) {
		t.Errorf("GetStringsFromLuaTable = %v", got)
	}

	L := lua.NewState()
	defer L.Close()
	table := GetLuaTableFromStrings(L, []string{"x", "y"})
	if table.Len() != 2 || table.RawGetInt(1) != lua.LString("x") {
		t.Errorf("GetLuaTableFromStrings = %v", table)
	}
	if values := GetValuesFromLuaTable(table); len(values) != 2 {
		t.Errorf("GetValuesFromLuaTable = %v", values)
	}
}

func TestForEach_StopsAtTheFirstError(t *testing.T) {
	tables := newTables(t, `{ { 1, 2, 3 } }`)
	stop := errors.New("stop")

	calls := 0
	err := ForEach(tables[0], func(_, value lua.LValue) error {
		calls++
		if value == lua.LNumber(2) {
			return stop
		}
		return nil
	})

	if !errors.Is(err, stop) || calls != 2 {
		t.Fatalf("ForEach = %v after %d calls, want the wrapped error after 2 calls", err, calls)
	}
}
