package modules

import (
	"github.com/LucasAVasco/falcula/lua/modules/modcmd"
	"github.com/LucasAVasco/falcula/lua/modules/moddockercompose"
	"github.com/LucasAVasco/falcula/lua/modules/modmanager"
	"github.com/LucasAVasco/falcula/lua/modules/modpath"
	"github.com/LucasAVasco/falcula/lua/modules/modprocess"
	"github.com/LucasAVasco/falcula/lua/modules/modtemplate"
)

type AllModulesLoaderOptions struct {
	// Scripts arguments to pass to falcula. If the user runs `falcula run a b c`, then CurrentArgs will be `[]string{"a", "b", "c"}`
	CurrentArgs []string

	CmdCallbacks     modcmd.Callbacks     // Callbacks of the command module
	ManagerCallbacks modmanager.Callbacks // Service manager callbacks
}

// LoadAllModules loads all available modules
func (l *Loader) LoadAllModules(opts *AllModulesLoaderOptions) error {
	l.LoadModule("falcula.manager", modmanager.New(&opts.ManagerCallbacks))
	l.LoadModule("falcula.cmd", modcmd.New(opts.CurrentArgs, &opts.CmdCallbacks))
	l.LoadModule("falcula.template", modtemplate.New())
	l.LoadModule("falcula.path", modpath.New())

	composeModule := moddockercompose.New()
	l.LoadModule("falcula.compose", composeModule)
	l.LoadModule("falcula.docker.compose", composeModule)
	l.LoadModule("falcula.process", modprocess.New())

	return nil
}
