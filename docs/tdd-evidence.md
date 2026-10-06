# TDD Evidence

## RED

`Test_Verify_rejects_placeholder_fixture_semantics` failed because `Verify` returned
nil on placeholder corpus. Captured output: `An error is expected but got nil.`

## GREEN

Executed corpus verification returned:

```text
corpus verified: total=50 directional=12 names=12 events=12 collisions=8 structural=6
```

After repair, corpus verifier rejects missing explicit expectations and duplicate ABI
pairs. Full Go suite passes. Pinned `abidiff` upstream suite passes 9/9 and all 50
fixtures execute. All consumer probes and determinism gates pass.

## Closure RED

- Unchanged E01 still returned `observed`.
- Invalid C03 still emitted `candidate_count=2`.
- Score mutation tests failed to compile because no evidence scoring API existed.

## Closure GREEN

- Unchanged E01 and semantic-addition S06 return `rejected`.
- Collision candidate counts and winner state derive from parsed ABI candidates.
- Mutated status/count/baseline/actionability/determinism evidence prevents 100 or GO.
- Machine scorer returns 100 only after recorded gates pass.

## Baseline Capability RED/GREEN

RED: baseline records had no capability parser and `ConsumerDetail` was assigned false
unconditionally. Equivalent synthetic baseline outputs could not remove class wins.

GREEN: strict parsing rejects malformed/missing output shape, persists capability facts,
and removes only classes whose complete frozen equivalence fact set is present. Setting
all fact sets true produces zero wins and NO-GO. Actual pinned output yields zero
equivalent records for directional, names, and events.

## FR-002 Repair RED/GREEN

RED: actual upstream output without synthetic `consumer_facts` became eleven false
values while `capabilities_parsed=true`. Parser accepted unknown and trailing JSON and
could not preserve source provenance or distinguish unsupported facts from false facts.

GREEN: strict typed decoding requires `breaking`, `additions`, `notes`, `bump`, and
non-empty `kind`, `signature`, `message` on every entry; unknown fields and trailing JSON
reject. Ordered source records persist separately from exact supported fact records.
Scoring requires parsed output and complete class fact support, so missing evidence earns
no baseline win. Pinned output parses 50/50 and now yields zero wins and NO-GO.

## FR-002 Equivalence Direction RED/GREEN

RED: a parsed exhaustive baseline with an empty or incomplete equivalent fact set was
rejected as unsupported, so all three otherwise qualifying classes lost their wins.

GREEN: each observation now requires exactly one parsed baseline record. A complete true
equivalent fact set removes only its matching class win; incomplete or empty sets permit
the win. Missing, duplicate, or `OutputParsed=false` records remain hard NO-GO. Strict
raw parsing remains unchanged. Fresh pinned execution parses 50/50; machine scoring
derives directional, events, and names wins, score 100, and technical GO.
