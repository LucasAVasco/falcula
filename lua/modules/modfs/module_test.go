package modfs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func TestWatchParallelRunsCallbacksAndReturnsTheirErrors(t *testing.T) {
	dir := t.TempDir()
	files := []string{filepath.Join(dir, "a"), filepath.Join(dir, "b")}
	for _, f := range files {
		if err := os.WriteFile(f, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	L := lua.NewState()
	defer L.Close()
	L.PreloadModule("fs", func(L *lua.LState) int {
		mod := L.NewTable()
		if err := New().Loader(L, "fs", mod); err != nil {
			t.Fatal(err)
		}
		L.Push(mod)
		return 1
	})
	L.SetGlobal("file_a", lua.LString(files[0]))
	L.SetGlobal("file_b", lua.LString(files[1]))

	// Touches the files while the Lua function is watching them
	go func() {
		for range 20 {
			time.Sleep(50 * time.Millisecond)
			for _, f := range files {
				_ = os.WriteFile(f, []byte("x"), 0o644)
			}
		}
	}()

	err := L.DoString(`
		local fs = require("fs")
		local calls = 0
		local cb = function()
			calls = calls + 1
			if calls >= 4 then error("stop watching") end
		end
		local ok, err = pcall(fs.watch_parallel, {
			{ files = { file_a }, callback = cb },
			{ files = { file_b }, callback = cb },
		})
		assert(not ok and tostring(err):find("stop watching"), tostring(err))
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWatchReturnsWhenTheScriptIsInterrupted(t *testing.T) {
	dir := t.TempDir()

	L := lua.NewState()
	defer L.Close()
	L.PreloadModule("fs", func(L *lua.LState) int {
		mod := L.NewTable()
		if err := New().Loader(L, "fs", mod); err != nil {
			t.Fatal(err)
		}
		L.Push(mod)
		return 1
	})
	L.SetGlobal("dir", lua.LString(dir))

	ctx, cancel := context.WithCancel(context.Background())
	L.SetContext(ctx)
	time.AfterFunc(200*time.Millisecond, cancel)

	done := make(chan error, 1)
	go func() {
		done <- L.DoString(`require("fs").watch({ dir }, function() end)`)
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected an error after the interruption")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("fs.watch did not return after the interruption")
	}
}
