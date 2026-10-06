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
