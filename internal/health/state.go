package health

type HealthState int

const (
	StateUnknown HealthState = iota
	StateHealthy
	StateUnhealthy
)

func (hs HealthState) String() string {
	switch hs {
	case StateUnknown:
		return "Unknown"
	case StateHealthy:
		return "Healthy"
	case StateUnhealthy:
		return "Unhealthy"
	default:
		return "Unknown"

	}
}
