# Final Verdict

**TECHNICAL GO, PUBLICATION LOCKED**

Score: 100/100. Threshold: 85/100. Hard-gate failures: none.

## Executed results

- Directional: 12/12 go-ethereum v1.15.11 old/new consumer directions executed.
- Names: 12/12 `abigen v1.15.11` old/new bindings generated and compiled.
- Events: 12/12 synthetic topic/data layouts, filters, and cross-version decodes executed.
- Collisions: 8/8 candidate sets preserved; outcomes remain ambiguous, never silently resolved.
- Structural controls: 6/6 baseline comparisons executed; S06 reports `none`.
- Pinned `abidiff`: upstream tests 9/9 and corpus comparisons 50/50.
- Full determinism: two runs byte-identical.
- Fresh-process stability: D06, N03, E07, C01, S06 each 30/30 identical.

Directional, names, and events beat admitted baseline by adding direction-specific decode,
generated-source compilation, and filter/topic/data consequences respectively.

## Assessed scope

- runtime_behavior: not_assessed
- storage_compatibility: not_assessed
- security: not_assessed
- deployment_compatibility: unknown

No publication, remote, outreach, RPC, deployment, or public mutation occurred.
