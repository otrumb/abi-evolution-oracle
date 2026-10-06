package experiment

import "local/abi-evolution-oracle-validation/internal/corpus"

type Assessment struct {
	Runtime    string `json:"runtime_behavior"`
	Storage    string `json:"storage_compatibility"`
	Security   string `json:"security"`
	Deployment string `json:"deployment_compatibility"`
}
type Observation struct {
	FixtureID       string     `json:"fixture_id"`
	Class           string     `json:"class"`
	Probe           string     `json:"probe"`
	Status          string     `json:"status"`
	Detail          string     `json:"detail"`
	Assessment      Assessment `json:"assessment"`
	Actionable      bool       `json:"actionable"`
	CandidateCount  int        `json:"candidate_count,omitempty"`
	ArbitraryWinner bool       `json:"arbitrary_winner,omitempty"`
}
type Baseline struct {
	FixtureID    string               `json:"fixture_id"`
	ExitCode     int                  `json:"exit_code"`
	Bump         string               `json:"bump"`
	Breaking     int                  `json:"breaking"`
	Additions    int                  `json:"additions"`
	OutputParsed bool                 `json:"output_parsed"`
	Sources      []BaselineSource     `json:"sources"`
	Capabilities BaselineCapabilities `json:"capabilities"`
}
type BaselineCapabilities struct {
	Facts []BaselineFact `json:"facts"`
}

type BaselineSource struct {
	Section   string `json:"section"`
	Index     int    `json:"index"`
	Kind      string `json:"kind"`
	Signature string `json:"signature"`
	Message   string `json:"message"`
}

type BaselineFact struct {
	Name   string `json:"name"`
	Value  bool   `json:"value"`
	Source int    `json:"source"`
}

func (value BaselineCapabilities) DirectionalEquivalent() bool {
	return value.value("call_identity_unchanged") && value.value("old_consumer_new_producer_decode") && value.value("new_consumer_old_producer_decode")
}
func (value BaselineCapabilities) NamesEquivalent() bool {
	return value.value("wire_identity_unchanged") && value.value("generated_api_impact") && value.value("source_compile_impact")
}
func (value BaselineCapabilities) EventsEquivalent() bool {
	return value.value("topic_identity") && value.value("topic_layout_impact") && value.value("data_layout_impact") && value.value("filter_impact") && value.value("cross_decode_impact")
}
func (value BaselineCapabilities) value(name string) bool {
	for _, fact := range value.Facts {
		if fact.Name == name && fact.Value {
			return true
		}
	}
	return false
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
	Score                  int           `json:"score"`
	TechnicalVerdict       string        `json:"technical_verdict"`
	ClassesBeatingBaseline []string      `json:"classes_beating_baseline"`
	HardGateFailures       []string      `json:"hard_gate_failures"`
	Sections               ScoreSections `json:"sections"`
}
type SectionScore struct {
	Earned    int `json:"earned"`
	Available int `json:"available"`
	Minimum   int `json:"minimum"`
}
type ScoreSections struct {
	Consumer        SectionScore `json:"consumer_evidence"`
	Corpus          SectionScore `json:"corpus_integrity"`
	Differentiation SectionScore `json:"baseline_differentiation"`
	Reproducibility SectionScore `json:"reproducibility"`
	Scope           SectionScore `json:"scope_and_publication"`
}
type GateEvidence struct {
	CorpusVerified       bool `json:"corpus_verified"`
	HashesVerified       bool `json:"hashes_verified"`
	ProvenanceVerified   bool `json:"provenance_verified"`
	ExpectationsVerified bool `json:"expectations_verified"`
	PinsVerified         bool `json:"pins_verified"`
	TestsVerified        bool `json:"tests_verified"`
	Deterministic        bool `json:"deterministic"`
	EvidencePolicy       bool `json:"evidence_policy"`
	PublicationLocked    bool `json:"publication_locked"`
}
type ScoringEvidence struct {
	Observations []Observation
	Baselines    []Baseline
	Corpus       corpus.Summary
	Gates        GateEvidence
}

func scope() Assessment { return Assessment{"not_assessed", "not_assessed", "not_assessed", "unknown"} }
