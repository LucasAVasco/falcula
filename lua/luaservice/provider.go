package luaservice

import (
	"fmt"

	"github.com/LucasAVasco/falcula/multiplexer"
	"github.com/LucasAVasco/falcula/provider/base"
	lua "github.com/yuin/gopher-lua"
)

// ParseProviderConfig create a provider configuration with the given name and multiplexer. The opts is a Lua table with the options for the
// provider that will be parsed and set in the provider configuration. If not provided (nil), returns a default provider configuration
// (without any options)
func ParseProviderConfig(name string, multiplexer *multiplexer.Multiplexer, opts lua.LValue) (*base.ProviderConfig, error) {
	config := base.ProviderConfig{
		Name:        name,
		Multiplexer: multiplexer,
	}

	switch opts.Type() {

	case lua.LTTable:
		table := opts.(*lua.LTable)

		if serviceOpts := table.RawGetString("service_opts"); serviceOpts.Type() == lua.LTTable {
			if start_disabled := serviceOpts.(*lua.LTable).RawGetString("start_disabled"); start_disabled.Type() == lua.LTBool {
				config.DefaultServiceOpts.StartDisabled = bool(start_disabled.(lua.LBool))
			}
		}

	case lua.LTNil:

	default:
		return nil, fmt.Errorf("the provider configuration must be a table, got %T", opts)
	}

	return &config, nil
}
