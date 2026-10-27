package health

type Status string

const (
	StatusOK       Status = "ok"
	StatusDegraded Status = "degraded"
	StatusError    Status = "error"
)

type CheckResult struct {
	Status Status `json:"status"`
}

type Report struct {
	Status Status                 `json:"status"`
	Checks map[string]CheckResult `json:"checks"`
}
