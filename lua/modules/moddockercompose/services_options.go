package moddockercompose

import (
	"fmt"

	"github.com/LucasAVasco/falcula/lua/luatable"
	"github.com/LucasAVasco/falcula/provider/dockercompose"
	lua "github.com/yuin/gopher-lua"
)

// parseBuildServiceOpts parses the build service options
func parseBuildServiceOpts(_ *lua.LState, argument lua.LValue) (*dockercompose.BuildServiceOpts, error) {
	if argument == lua.LNil {
		return nil, nil
	}

	// Result
	opts := dockercompose.BuildServiceOpts{}

	// Checks if the argument is a table
	table, ok := argument.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("the options list must be a table, got %T", argument)
	}

	// 'no_pull'
	if noPull := table.RawGetString("no_pull"); noPull.Type() == lua.LTBool {
		opts.NoPull = bool(noPull.(lua.LBool))
	}

	// Parses the 'builds' field
	if builds := table.RawGetString("builds"); builds.Type() == lua.LTTable {
		builds := builds.(*lua.LTable)

		opts.Builds = make([]*dockercompose.BuildInfo, 0)

		err := luatable.ForEach(builds, func(key, value lua.LValue) error {
			options, ok := value.(*lua.LTable)
			if !ok {
				return fmt.Errorf("the build information must be a table, got %T", value)
			}

			// 'builds[n]'
			buildInfo := &dockercompose.BuildInfo{}

			// 'builds[n].services'
			services := options.RawGetString("services")
			if services.Type() == lua.LTTable {
				buildInfo.ServicesNames = luatable.GetStringsFromLuaTable(services.(*lua.LTable))
			} else if services == lua.LNil {
			} else {
				return fmt.Errorf("services list must be a table of strings, got %T", services)
			}

			// 'builds[n].platforms'
			platforms := options.RawGetString("platforms")
			if platforms.Type() == lua.LTTable {
				buildInfo.Platforms = luatable.GetStringsFromLuaTable(platforms.(*lua.LTable))
			} else if services == lua.LNil {
			} else {
				return fmt.Errorf("platforms list must be a table of strings, got %T", platforms)
			}

			opts.Builds = append(opts.Builds, buildInfo)

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error iterating through the list of build information: %w", err)
		}
	}

	return &opts, nil
}
