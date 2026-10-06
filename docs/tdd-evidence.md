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
