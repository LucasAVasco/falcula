package modules

import (
	"fmt"

	"github.com/LucasAVasco/falcula/lua/modules/modcmd"
	"github.com/LucasAVasco/falcula/lua/modules/moddockercompose"
	"github.com/LucasAVasco/falcula/lua/modules/modfalcula"
	"github.com/LucasAVasco/falcula/lua/modules/modinspect"
	"github.com/LucasAVasco/falcula/lua/modules/modjson"
	"github.com/LucasAVasco/falcula/lua/modules/modmanager"
	"github.com/LucasAVasco/falcula/lua/modules/modpath"
	"github.com/LucasAVasco/falcula/lua/modules/modprocess"
	"github.com/LucasAVasco/falcula/lua/modules/modtemplate"
	"github.com/LucasAVasco/falcula/lua/modules/modtui"
	"github.com/LucasAVasco/falcula/lua/modules/modyaml"
	"github.com/LucasAVasco/falcula/service/enhanced"
	"github.com/LucasAVasco/falcula/service/manager"
)

// AllModulesLoaderOptions is the configuration required by the module loader in order to load all modules
type AllModulesLoaderOptions struct {
	RawMode      bool                // Disables the TUI. Runs non-interactive
	OnSelectArgs func(args []string) // Called when the user selects the arguments to re-run the script
}

// LoadAllModules loads all available modules
func (l *Loader) LoadAllModules(config *AllModulesLoaderOptions) error {
	// Tui module
	tuiModuleConfig := modtui.Config{
		RawMode:      config.RawMode,
		OnSelectArgs: config.OnSelectArgs,
	}
	tuiModule := modtui.New(&tuiModuleConfig)

	// Manager module
	managerMod := modmanager.New(l.getManagerCallbacks(tuiModule))
	modCmd := modcmd.New()

	// Runtime callbacks
	l.runtime.SetOnCurrentScriptArgsChange(func(args []string) {
		modCmd.SetCurrentScriptArgs(args)
		if tui := tuiModule.GetTui(); tui != nil {
			tui.SetCurrentScriptArgs(args)
		}
	})

	l.runtime.SetOnScriptAvailableArgsChange(func(args [][]string) {
		if tui := tuiModule.GetTui(); tui != nil {
			tui.SetAvailableScriptArgs(args)
		}
	})

	// Other modules
	l.LoadModule("falcula", modfalcula.New())
	l.LoadModuleFromFunction("falcula.inspect", modinspect.LoadFunction)
	l.LoadModule("falcula.json", modjson.New())
	l.LoadModule("falcula.yaml", modyaml.New())
	l.LoadModule("falcula.manager", managerMod)
	l.LoadModule("falcula.cmd", modCmd)
	l.LoadModule("falcula.template", modtemplate.New())
	l.LoadModule("falcula.path", modpath.New())
	l.LoadModule("falcula.tui", tuiModule)

	// Service providers modules
	composeModule := moddockercompose.New()
	l.LoadModule("falcula.compose", composeModule)
	l.LoadModule("falcula.docker.compose", composeModule)
	l.LoadModule("falcula.process", modprocess.New())

	return nil
}

// getManagerCallbacks gets the callbacks for the manager module
func (l *Loader) getManagerCallbacks(tuiModule *modtui.Module) *modmanager.Callbacks {
	managerCallbacks := modmanager.Callbacks{}

	logError := func(err error) {
		l.runtime.Logger.LogError(err)
	}

	managerCallbacks.OnNewManager = func(man *manager.Manager) {
		l.runtime.AddManager(man)

		if tui := tuiModule.GetTui(); tui != nil {
			err := tui.AddManagerToSidebar(man)
			if err != nil {
				logError(fmt.Errorf("error adding manager: %w", err))
			}
		}
	}

	managerCallbacks.OnDeleteManager = func(man *manager.Manager) {
		l.runtime.RemoveManager(man)

		if tui := tuiModule.GetTui(); tui != nil {
			err := tui.RemoveManagerFromSidebar(man)
			if err != nil {
				logError(fmt.Errorf("error removing manager: %w", err))
			}
		}
	}

	managerCallbacks.OnAddService = func(man *manager.Manager, svc *enhanced.EnhancedService) {
		if tui := tuiModule.GetTui(); tui != nil {
			err := tui.AddServiceToSidebar(man, svc)
			if err != nil {
				logError(fmt.Errorf("error adding service: %w", err))
			}
		}
	}

	managerCallbacks.OnServiceStatusChanged = func(man *manager.Manager, svc *enhanced.EnhancedService) {
		if tui := tuiModule.GetTui(); tui != nil {
			err := tui.UpdateServiceStatusInSidebar(man, svc)
			if err != nil {
				logError(fmt.Errorf("error updating service status: %w", err))
			}
		}
	}

	return &managerCallbacks
}
