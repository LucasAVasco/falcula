package manager

import (
	"github.com/LucasAVasco/falcula/process"
	"github.com/LucasAVasco/falcula/service/enhanced"
)

func applyDefaultExitProcessCallback(callback enhanced.ExitStepCallback) enhanced.ExitStepCallback {
	if callback == nil {
		return func(svc *enhanced.EnhancedService, exitInfo *process.ExitInfo, err error) {}
	}

	return callback
}
