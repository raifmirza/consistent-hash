package health

type State uint8

const (
	StateUnknown State = iota
	StateHealthy
	StateUnhealthy
)

func (s State) String() string {
	switch s {
	case StateHealthy:
		return "Healthy"
	case StateUnhealthy:
		return "Unhealthy"
	default:
		return "Unknown"
	}
}
