package modtui

import (
	"testing"

	"github.com/LucasAVasco/falcula/internal/testutil"
	"github.com/LucasAVasco/falcula/internal/testutil/testruntime"
)

func TestShowArgs_RawMode_DoesNothing(t *testing.T) {
	r := testruntime.New(t, nil, nil)
	m := testruntime.Configure(r, New(&Config{RawMode: true}))
	L := testutil.NewLuaState(t, map[string]testutil.ModuleLoader{"falcula.tui": m})

	if err := L.DoString(`local tui = require("falcula.tui"); tui.show_args(); tui.show()`); err != nil {
		t.Fatal(err)
	}
	if getTui() != nil {
		t.Error("the text user interface must not be created in raw mode")
	}
}
