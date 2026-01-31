// Package status implements the list of possible statuses for a service
package status

type Status uint

const (
	None            Status = 0
	Preparing       Status = 1
	Ready           Status = 2
	AbortingPrepare Status = 3
	PrepareAborted  Status = 4
	Running         Status = 5
	Ended           Status = 6 // Service completed without error
	Stopping        Status = 7
	Stopped         Status = 8 // Service manually exited without error
	Error           Status = 9
)

// IsDoingNothing returns true if the service is doing nothing. Returns false if the status is 'Error'
func (s Status) IsDoingNothing() bool {
	return s == None || s == Ready || s == PrepareAborted || s == Ended || s == Stopped
}

func (s Status) ToString() string {
	switch s {
	case None:
		return "None"
	case Preparing:
		return "Preparing"
	case Ready:
		return "Ready"
	case AbortingPrepare:
		return "AbortingPrepare"
	case PrepareAborted:
		return "PrepareAborted"
	case Running:
		return "Running"
	case Ended:
		return "Ended"
	case Stopping:
		return "Stopping"
	case Stopped:
		return "Stopped"
	case Error:
		return "Error"
	default:
		return "Unknown"
	}
}
