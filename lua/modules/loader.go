// Package modules is a package to load Lua modules
package modules

import (
	"fmt"

	"github.com/LucasAVasco/falcula/lua/luaruntime"
	"github.com/LucasAVasco/falcula/lua/modules/base"

	lua "github.com/yuin/gopher-lua"
)

// ModulesConfig is the base configuration that a module requires to work
type ModulesConfig = base.Config

// Loader is a loader for Lua modules
type Loader struct {
	runtime       *luaruntime.Runtime
	modulesConfig *ModulesConfig
	modules       map[string]Module
}

func NewLoader(runtime *luaruntime.Runtime, config *ModulesConfig) (*Loader, error) {
	return &Loader{
		runtime:       runtime,
		modulesConfig: config,
		modules:       make(map[string]Module),
	}, nil
}

func (l *Loader) LoadModuleFromFunction(moduleName string, function func(name string, L *lua.LState) int) {
	l.runtime.GetLuaState().PreloadModule(moduleName, func(l *lua.LState) int {
		return function(moduleName, l)
	})
}

func (l *Loader) LoadModule(moduleName string, module Module) {
	l.modules[moduleName] = module

	module.SetBaseModuleConfig(l.modulesConfig)
	l.runtime.GetLuaState().PreloadModule(moduleName, func(L *lua.LState) int {
		moduleTable := L.NewTable()

		err := module.Loader(L, moduleName, moduleTable)
		if err != nil {
			l.modulesConfig.Runtime.Logger.LogError(fmt.Errorf("error loading module '%s': %w", moduleName, err))
			return 0
		}

		L.Push(moduleTable)
		return 1
	})
}

// Close closes the loader and all loaded modules. Can be called multiple times.
func (l *Loader) Close() error {
	for name, module := range l.modules {
		err := module.Close()
		if err != nil {
			return fmt.Errorf("error closing module '%s': %w", name, err)
		}

		// Removes the module from the map (required to be able to call `Close` multiple times)
		delete(l.modules, name)
	}

	return nil
}

// GetModule returns the module with the given name
func (l *Loader) GetModule(name string) Module {
	return l.modules[name]
}
