package base

import (
	"github.com/LucasAVasco/falcula/lua/luaruntime"
)

// Config is the options that a module requires to work
type Config struct {
	Runtime *luaruntime.Runtime
}
