package modules

import (
	"github.com/LucasAVasco/falcula/lua/luaruntime"
)

// LoadAllModules loads all available modules and returns the loader. Closing the loader also closes all loaded modules
func LoadAllModules(runtime *luaruntime.Runtime, config *AllModulesLoaderOptions) *Loader {
	loader := NewLoader(runtime, &ModulesConfig{
		Runtime: runtime,
	})
	loader.LoadAllModules(config)

	return loader
}
