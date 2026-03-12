package modules

import (
	"fmt"

	"github.com/LucasAVasco/falcula/lua/luaruntime"
)

// LoadAllModules loads all available modules and returns the loader. Closing the loader also closes all loaded modules
func LoadAllModules(runtime *luaruntime.Runtime, config *AllModulesLoaderOptions) (*Loader, error) {
	loader, err := NewLoader(runtime, &ModulesConfig{
		Runtime: runtime,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating loader: %w", err)
	}

	loader.LoadAllModules(config)

	return loader, nil
}
