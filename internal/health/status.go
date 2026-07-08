package health

type NodeStatus struct {
	State                HealthState
	ConsecutiveFailures  int
	ConsecutiveSuccesses int
	LastError            error
}
