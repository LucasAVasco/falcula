package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/LucasAVasco/falcula/lua/modules"
	"github.com/LucasAVasco/falcula/service/enhanced"
	"github.com/LucasAVasco/falcula/service/manager"
	"github.com/LucasAVasco/falcula/tui/ioredirect"

	lua "github.com/yuin/gopher-lua"
)

// fillManagerCallbacks fills the callbacks for the service manager module.
func (t *Tui) fillManagerCallbacks(modulesOpts *modules.AllModulesLoaderOptions) {
	modulesOpts.ManagerCallbacks.OnNewManager = func(man *manager.Manager) {
		err := t.mainPage.SideBar.AddManager(man)
		if err != nil {
			t.mainPage.DebugLogs.Append(fmt.Errorf("error adding manager: %w", err))
		}
	}
	modulesOpts.ManagerCallbacks.OnDeleteManager = func(man *manager.Manager) {
		err := t.mainPage.SideBar.RemoveManager(man)
		if err != nil {
			t.mainPage.DebugLogs.Append(fmt.Errorf("error removing manager: %w", err))
		}
	}
	modulesOpts.ManagerCallbacks.OnAddService = func(man *manager.Manager, svc *enhanced.EnhancedService) {
		err := t.mainPage.SideBar.AddService(man, svc)
		if err != nil {
			t.mainPage.DebugLogs.Append(fmt.Errorf("error adding service: %w", err))
		}
	}
	modulesOpts.ManagerCallbacks.OnServiceStatusChanged = func(man *manager.Manager, svc *enhanced.EnhancedService) {
		err := t.mainPage.SideBar.UpdateServiceStatus(man, svc)
		if err != nil {
			t.mainPage.DebugLogs.Append(fmt.Errorf("error updating service status: %w", err))
		}
	}

	modulesOpts.ManagerCallbacks.OnServiceLog = func(service *enhanced.EnhancedService, message string) {
		t.mainPage.ServiceLogs.Append(message)
	}
	modulesOpts.ManagerCallbacks.OnDebugLog = func(message string) {
		t.mainPage.DebugLogs.Append(message, "\n")
	}
}

// fillCmdCallbacks fills the callbacks for the command module.
func (t *Tui) fillCmdCallbacks(modulesOpts *modules.AllModulesLoaderOptions) {
	modulesOpts.CmdCallbacks.OnSetAvailableCmdArgs = func(args [][]string) {
		t.argsView.SetAvailableArgs(args)
	}
}

// closeLuaState closes the current Lua state
func (t *Tui) closeLuaState() {
	if t.luaState != nil {
		t.luaState.Close()
		t.luaState = nil
	}
}

// Execute the Lua configuration file (Lua script). Blocks until the file execution is finished.
//
// The `args` are the arguments available to the script.
func (t *Tui) executeLuaFile(args []string) error {
	if !t.luaFileMutex.TryLock() {
		return nil
	}
	defer t.luaFileMutex.Unlock()

	// Closes the last loader
	if t.moduleLoader != nil {
		err := t.moduleLoader.Close()
		if err != nil {
			return fmt.Errorf("error closing last loader: %w", err)
		}
	}

	// Closes the last Lua state
	t.closeLuaState()

	// New state
	L := lua.NewState()
	t.luaState = L
	defer t.closeLuaState()

	// Redirects stdout to the debug logs
	ioredirect.Redirect(L, t.mainPage.DebugLogs)

	// Argument view
	t.argsView.SetCurrentArgs(args)

	// Module loader
	loaderOpts := modules.ModulesOptions{
		Multiplexer: t.multi,
		OnDebug: func(msg string) {
			t.mainPage.DebugLogs.Append(msg, "\n")
		},
		OnError: func(err error) {
			t.mainPage.DebugLogs.Append(err)
		},
	}

	loader, err := modules.NewLoader(L, &loaderOpts)
	if err != nil {
		return fmt.Errorf("error creating module loader: %w", err)
	}
	t.moduleLoader = loader

	// Loads all modules
	modulesOpts := modules.AllModulesLoaderOptions{
		CurrentArgs: args,
	}
	t.fillManagerCallbacks(&modulesOpts)
	t.fillCmdCallbacks(&modulesOpts)
	loader.LoadAllModules(&modulesOpts)

	// Sets the command line arguments
	argTable := L.NewTable()
	for i, arg := range args {
		argTable.RawSetInt(i+1, lua.LString(arg))
	}
	L.SetGlobal("arg", argTable)

	// Gets the main file path
	mainFiles := []string{"falcula.lua", "falcfg/init.lua", ".falcula.lua", ".falcfg/init.lua"}
	mainFilePath := ""
	for _, mainFile := range mainFiles {
		if _, err := os.Stat(mainFile); err == nil {
			mainFilePath = mainFile
			break
		}
	}
	if mainFilePath == "" {
		return fmt.Errorf("can not find the main file, possible names: %s", strings.Join(mainFiles, ", "))
	}

	// Executes the main file
	if err := L.DoFile(mainFilePath); err != nil {
		return fmt.Errorf("error running script: %w", err)
	}

	return nil
}

// createLuaRoutine executes the Lua file in a separate goroutine. Does not block.
func (t *Tui) createLuaRoutine(args []string) {
	go func() {
		err := t.executeLuaFile(args)
		if err != nil {
			t.mainPage.DebugLogs.Append(fmt.Errorf("error executing Lua file: %w", err))
		}
	}()
}
