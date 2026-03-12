// Package base implements the base Lua module
package base

// BaseModule is the base module that all other modules should inherit
type BaseModule struct {
	Config *Config // It is set by the module loader
}

// SetConfig sets the configurations common for all modules. Is called by the module loader
func (b *BaseModule) SetBaseModuleConfig(opts *Config) {
	b.Config = opts
}

func (b *BaseModule) Close() error {
	return nil
}
