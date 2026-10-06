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
	FixtureID          string               `json:"fixture_id"`
	ExitCode           int                  `json:"exit_code"`
	Bump               string               `json:"bump"`
	Breaking           int                  `json:"breaking"`
	Additions          int                  `json:"additions"`
	Capabilities       BaselineCapabilities `json:"capabilities"`
	CapabilitiesParsed bool                 `json:"capabilities_parsed"`
}
type BaselineCapabilities struct {
	CallIdentityUnchanged        bool `json:"call_identity_unchanged"`
	OldConsumerNewProducerDecode bool `json:"old_consumer_new_producer_decode"`
	NewConsumerOldProducerDecode bool `json:"new_consumer_old_producer_decode"`
	WireIdentityUnchanged        bool `json:"wire_identity_unchanged"`
	GeneratedAPIImpact           bool `json:"generated_api_impact"`
	SourceCompileImpact          bool `json:"source_compile_impact"`
	TopicIdentity                bool `json:"topic_identity"`
	TopicLayoutImpact            bool `json:"topic_layout_impact"`
	DataLayoutImpact             bool `json:"data_layout_impact"`
	FilterImpact                 bool `json:"filter_impact"`
	CrossDecodeImpact            bool `json:"cross_decode_impact"`
}

func (value BaselineCapabilities) DirectionalEquivalent() bool {
	return value.CallIdentityUnchanged && value.OldConsumerNewProducerDecode && value.NewConsumerOldProducerDecode
}
func (value BaselineCapabilities) NamesEquivalent() bool {
	return value.WireIdentityUnchanged && value.GeneratedAPIImpact && value.SourceCompileImpact
}
func (value BaselineCapabilities) EventsEquivalent() bool {
	return value.TopicIdentity && value.TopicLayoutImpact && value.DataLayoutImpact && value.FilterImpact && value.CrossDecodeImpact
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
