package health

type NodeStatus struct {
	State                State
	ConsecutiveFailures  int
	ConsecutiveSuccesses int
	LastError            error
}
