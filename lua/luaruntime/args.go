package luaruntime

// GetScriptCurrentArgs gets the current arguments for the current running script
func (r *Runtime) GetScriptCurrentArgs() []string {
	return r.scriptCurrentArgs
}

// SetScriptAvailableArgs sets the available arguments for the current running script
func (r *Runtime) SetScriptAvailableArgs(args [][]string) {
	r.scriptAvailableArgs = args
	r.onSetScriptAvailableArgs(args)
}

// GetScriptAvailableArgs gets the available arguments for the current running script
func (r *Runtime) GetScriptAvailableArgs() [][]string {
	return r.scriptAvailableArgs
}

// SetOnCurrentScriptArgsChange sets the function to be called when the current script arguments change. Overrides the last one if called
// multiple times
func (r *Runtime) SetOnCurrentScriptArgsChange(f func(args []string)) {
	r.onSetScriptCurrentArgs = f
}

// SetOnScriptAvailableArgsChange sets the function to be called when the available script arguments change. Overrides the last one if
// called multiple times
func (r *Runtime) SetOnScriptAvailableArgsChange(f func(args [][]string)) {
	r.onSetScriptAvailableArgs = f
}
