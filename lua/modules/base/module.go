// Package base implements the base Lua module
package base

// BaseModule is the base module that all other modules should inherit
type BaseModule struct {
	Opts *Options // Is set by the module loader
}

// SetOpts sets the options for the module. Is called by the module loader
func (b *BaseModule) SetOpts(opts *Options) {
	b.Opts = opts
}

func (b *BaseModule) Close() error {
	return nil
}
