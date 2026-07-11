package health

import "time"

type NodeStatus struct {
	State                State
	ConsecutiveFailures  int
	ConsecutiveSuccesses int
	LastCheck            time.Time
	LastError            error
}
