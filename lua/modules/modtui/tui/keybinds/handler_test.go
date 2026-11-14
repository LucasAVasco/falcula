package keybinds

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestInputCaptureFunction_DispatchesKeyBinds(t *testing.T) {
	var calls []string
	asyncDone := make(chan struct{})

	h := NewHandler("test")
	h.AddKeyBinds([]*KeyBind{
		{Rune: 'r', Desc: "restart", Bind: func() { calls = append(calls, "r") }},
		{Key: tcell.KeyEnter, Desc: "open", Bind: func() { calls = append(calls, "enter") }},
		{Rune: 'h', Desc: "description only"},
		{Rune: 'a', Desc: "async", Async: true, Bind: func() { close(asyncDone) }},
	})
	capture := h.GetInputCaptureFunction()

	runeEvent := tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone)
	if capture(runeEvent) != nil {
		t.Error("a handled key must be consumed")
	}
	if capture(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)) != nil {
		t.Error("a handled special key must be consumed")
	}

	unknown := tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone)
	if capture(unknown) != unknown {
		t.Error("an unknown key must be passed through")
	}
	descOnly := tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone)
	if capture(descOnly) != descOnly {
		t.Error("a key bind without function must pass the event through")
	}

	if capture(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone)) != nil {
		t.Error("an async key bind must be consumed")
	}
	<-asyncDone

	if len(calls) != 2 || calls[0] != "r" || calls[1] != "enter" {
		t.Fatalf("calls = %v", calls)
	}
}

func TestAddKeyBinds_WithoutKey_Panics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic")
		}
	}()
	NewHandler("test").AddKeyBinds([]*KeyBind{{Desc: "no key"}})
}

func TestKeyBindNamesAndChildren(t *testing.T) {
	if name := (&KeyBind{Rune: 'x'}).GetName(); name != "x" {
		t.Errorf("rune name = %q", name)
	}
	if name := (&KeyBind{Key: tcell.KeyEnter}).GetName(); name != "<Enter>" {
		t.Errorf("key name = %q", name)
	}

	parent, child := NewHandler("parent"), NewHandler("child")
	parent.AddChildHandler(child)
	if parent.GetName() != "parent" || len(parent.GetChildHandlers()) != 1 || parent.GetChildHandlers()[0] != child {
		t.Error("child handlers were not stored")
	}
}
