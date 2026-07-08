package health

type NodeStatus struct {
	Healthy              bool
	ConsecutiveFailures  int
	ConsecutiveSuccesses int
}
