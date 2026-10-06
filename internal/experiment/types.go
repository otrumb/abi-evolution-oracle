package experiment

type Assessment struct {
	Runtime    string `json:"runtime_behavior"`
	Storage    string `json:"storage_compatibility"`
	Security   string `json:"security"`
	Deployment string `json:"deployment_compatibility"`
}
type Observation struct {
	FixtureID  string     `json:"fixture_id"`
	Class      string     `json:"class"`
	Probe      string     `json:"probe"`
	Status     string     `json:"status"`
	Detail     string     `json:"detail"`
	Assessment Assessment `json:"assessment"`
}
type Baseline struct {
	FixtureID string `json:"fixture_id"`
	ExitCode  int    `json:"exit_code"`
	Bump      string `json:"bump"`
	Breaking  int    `json:"breaking"`
	Additions int    `json:"additions"`
}
type Summary struct {
	Directional int `json:"directional"`
	Names       int `json:"names"`
	Events      int `json:"events"`
	Collisions  int `json:"collisions"`
	Structural  int `json:"structural"`
	Baselines   int `json:"baselines"`
}
type Score struct {
	Score                  int      `json:"score"`
	TechnicalVerdict       string   `json:"technical_verdict"`
	ClassesBeatingBaseline []string `json:"classes_beating_baseline"`
	HardGateFailures       []string `json:"hard_gate_failures"`
}

func scope() Assessment { return Assessment{"not_assessed", "not_assessed", "not_assessed", "unknown"} }
