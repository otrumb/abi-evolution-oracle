# TDD Evidence

## RED

No RED output is claimed. Initial corpus test could not run until fixtures existed;
that state was not captured as command output and is therefore not represented as an
executed RED result.

## GREEN

Executed corpus verification returned:

```text
corpus verified: total=50 directional=12 names=12 events=12 collisions=8 structural=6
```

Executed Go suite returned green for `internal/corpus`. Pinned `abidiff` upstream suite
passed 9/9. Semantic review then failed frozen fixture expectations, forcing NO-GO.
