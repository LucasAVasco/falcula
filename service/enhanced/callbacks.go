package enhanced

import (
	"github.com/LucasAVasco/falcula/service/iface"
)

// ExitStepCallback is a function that is called when a Step related to the service is done
type ExitStepCallback func(svc *EnhancedService, exitInfo *iface.ExitInfo, err error)

type Callbacks struct {
	OnServiceStatusChanged func(svc *EnhancedService) // Called when the status of the service changes
	OnExitProcess          ExitStepCallback           // Called when a step of the service is done
}

func fillCallbacksWithDefaults(callbacks *Callbacks) *Callbacks {
	if callbacks == nil {
		callbacks = &Callbacks{}
	}

	if callbacks.OnServiceStatusChanged == nil {
		callbacks.OnServiceStatusChanged = func(svc *EnhancedService) {}
	}

	if callbacks.OnExitProcess == nil {
		callbacks.OnExitProcess = func(svc *EnhancedService, exitInfo *iface.ExitInfo, err error) {}
	}

	return callbacks
}
