// Package luapath contains functions for working with paths in Lua
package luapath

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/yuin/gopher-lua"
)

func getStackSource(L *lua.LState, index int) (string, error) {
	// get stack level 1 (Lua caller of this function)
	stack, ok := L.GetStack(index)
	if !ok {
		return "", fmt.Errorf("Failed to get stack level 1")
	}

	// Fill debug info with source
	_, err := L.GetInfo("S", stack, lua.LNil)
	if err != nil {
		return "", fmt.Errorf("Failed to get debug info: %w", err)
	}

	return stack.Source, nil
}

// GetCurrentLuaFile returns the path of the current Lua file being executed
func GetCurrentLuaFile(L *lua.LState) (string, error) {
	// Find a stack source that matches a Lua file
	var stack_source string
	for i := 0; i <= 100; i++ {
		var err error
		stack_source, err = getStackSource(L, i)
		if err != nil {
			return "", fmt.Errorf("Failed to get stack level %d: %w", i, err)
		}

		// Ends if file is a Lua file
		if filepath.Ext(stack_source) == ".lua" {
			break
		}
	}

	// Source file
	file := strings.TrimPrefix(stack_source, "@")
	file, err := filepath.Abs(file)
	if err != nil {
		return "", fmt.Errorf("Failed to get absolute path of '%s': %w", file, err)
	}
	file = filepath.Clean(file)

	return file, nil
}

// GetCurrentLuaDirectory returns the directory of the current Lua file being executed
func GetCurrentLuaDirectory(L *lua.LState) (string, error) {
	// Get file
	file, err := GetCurrentLuaFile(L)
	if err != nil {
		return "", fmt.Errorf("Failed to get current Lua file: %w", err)
	}

	// Get directory
	dir := filepath.Dir(file)

	return dir, nil
}

// GetAbs converts a path to an absolute path. If the path is relative, uses the directory of the current Lua script as base
func GetAbs(L *lua.LState, path string) (string, error) {
	if !filepath.IsAbs(path) {
		currentDir, err := GetCurrentLuaDirectory(L)
		if err != nil {
			return "", fmt.Errorf("Failed to get current Lua directory: %w", err)
		}
		path = filepath.Join(currentDir, path)
	}

	// Makes it absolute
	path, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("Failed to get absolute path: %w", err)
	}

	// Cleans (normalizes)
	return filepath.Clean(path), nil
}

// GetRel converts a path to a relative path. If the path is relative, uses the directory of the current Lua script as base. If the base is
// empty, uses the directory of the current Lua script
func GetRel(L *lua.LState, path string, base string) (string, error) {
	path, err := GetAbs(L, path)
	if err != nil {
		return "", fmt.Errorf("Failed to get absolute path of 'path': %w", err)
	}

	if base == "" {
		base, err = GetCurrentLuaDirectory(L)
		if err != nil {
			return "", fmt.Errorf("Failed to get current Lua directory: %w", err)
		}
	} else {
		base, err = GetAbs(L, base)
		if err != nil {
			return "", fmt.Errorf("Failed to get absolute path of 'base': %w", err)
		}
	}

	return filepath.Rel(base, path)
}
