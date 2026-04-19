// Package moddockercompose is a module that provides functions and classes for working with Docker Compose
package moddockercompose

import (
	"fmt"

	"github.com/LucasAVasco/falcula/lua/luaclass"
	"github.com/LucasAVasco/falcula/lua/luadata"
	"github.com/LucasAVasco/falcula/lua/luaerror"
	"github.com/LucasAVasco/falcula/lua/luapath"
	"github.com/LucasAVasco/falcula/lua/luatable"
	"github.com/LucasAVasco/falcula/lua/maplua"
	"github.com/LucasAVasco/falcula/lua/modules/base"
	"github.com/LucasAVasco/falcula/provider/dockercompose"

	lua "github.com/yuin/gopher-lua"
)

type Loader struct {
	base.BaseModule
}

func New() *Loader {
	return &Loader{}
}

func (l *Loader) Loader(L *lua.LState, name string, mod *lua.LTable) error {
	info := luaclass.Info{
		Name: "Provider",
		Constructor: func(L *lua.LState, newObj *lua.LTable) error {
			providerName := L.ToString(2)
			composeFile := L.ToString(3)

			// Provider configurations
			config := dockercompose.ProviderConfig{}
			config.Name = providerName
			config.Multiplexer = l.Config.Runtime.Logger.GetServicesMultiplexer()
			err := maplua.Unmarshal(L.OptTable(4, L.NewTable()), &config.Opts)
			if err != nil {
				return fmt.Errorf("error getting provider configuration: %w", err)
			}

			// Creates the provider for the compose file
			composeFile, err = luapath.GetAbs(L, composeFile)
			if err != nil {
				return fmt.Errorf("error getting absolute path: %w\n", err)
			}

			provider := dockercompose.New(&config, composeFile)

			// Sets the provider in the instance
			luaclass.SetAttribute(L, newObj, "_provider", provider)

			return nil
		},
		Methods: methods,
	}
	class, err := luaclass.New(L, &info, l.Config.Runtime.Logger.GetOnErrorWithoutReturn())
	if err != nil {
		return fmt.Errorf("error creating class '%s' of '%s' module: %w", info.Name, name, err)

	}

	L.SetField(mod, info.Name, class)

	return nil
}

// getProvider gets the docker compose provider when called inside a method. Must not be used outside a method
func getProvider(L *lua.LState) *dockercompose.Provider {
	return luaclass.GetAttribute(L, "_provider").(*dockercompose.Provider)
}

var methods = map[string]lua.LGFunction{
	"get_name": func(L *lua.LState) int {
		provider := getProvider(L)
		L.Push(lua.LString(provider.GetName()))
		return 1
	},

	"add_image": func(L *lua.LState) int {
		provider := getProvider(L)
		image := L.ToString(2)
		provider.AddDefaultPushImage(image)
		return 0
	},

	"add_push_images": func(L *lua.LState) int {
		provider := getProvider(L)
		images := luatable.GetStringsFromLuaTable(L.ToTable(2))
		provider.AddDefaultPushImages(images)
		return 0
	},

	"new_build_service": func(L *lua.LState) int {
		provider := getProvider(L)
		opts := dockercompose.BuildServiceOpts{}
		err := maplua.Unmarshal(L.OptTable(2, L.NewTable()), &opts)
		if err != nil {
			return luaerror.Push(L, 1, fmt.Errorf("error parsing build options: %w", err))
		}
		L.Push(luadata.NewUserData(L, provider.NewBuildService(&opts)))
		return 1
	},

	"new_up_service": func(L *lua.LState) int {
		provider := getProvider(L)
		platform := L.OptString(2, "")

		opts := dockercompose.ServiceOpts{}
		err := maplua.Unmarshal(L.OptTable(3, L.NewTable()), &opts)
		if err != nil {
			return luaerror.Push(L, 1, err)
		}
		L.Push(luadata.NewUserData(L, provider.NewUpService(platform, &opts)))
		return 1
	},

	"new_down_service": func(L *lua.LState) int {
		provider := getProvider(L)
		opts := dockercompose.DownServiceOpts{}
		err := maplua.Unmarshal(L.OptTable(2, L.NewTable()), &opts)
		if err != nil {
			return luaerror.Push(L, 1, err)
		}
		L.Push(luadata.NewUserData(L, provider.NewDownService(&opts)))
		return 1
	},

	"new_push_service": func(L *lua.LState) int {
		provider := getProvider(L)
		opts := dockercompose.PushServiceOpts{}
		err := maplua.Unmarshal(L.OptTable(2, L.NewTable()), &opts)
		if err != nil {
			return luaerror.Push(L, 1, fmt.Errorf("error parsing push options: %w", err))
		}
		L.Push(luadata.NewUserData(L, provider.NewPushService(&opts)))
		return 1
	},
}
