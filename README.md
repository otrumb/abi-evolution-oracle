# ABI Evolution Oracle

`abi-evolution-oracle` is a reproducible experiment and CLI for measuring
Ethereum ABI consumer compatibility across frozen old/new ABI pairs. It executes
go-ethereum decoding, `abigen` source generation and compilation, event layout
probes, collision handling, structural controls, and a pinned `abidiff` baseline.

## Scope

The checked-in corpus contains exactly 50 original synthetic cases: 12
directional, 12 name-only, 12 event, 8 collision, and 6 structural cases. The
corpus is frozen and dedicated under CC0-1.0; no public ABI was copied.

The recorded experiment uses:

- Go 1.24.0 with `GOTOOLCHAIN=local` and `CGO_ENABLED=0`
- go-ethereum and `abigen` v1.15.11
- `jay-tank/abidiff` commit `c9488370c95cd0f9406d3fbec6aa26cc95c1f3b6`

The checked-in evidence reports a technical score of 100/100 for this experiment.
That score is experiment evidence, not market validation.

**Disclaimer:** Reports consumer compatibility evidence only for tested corpus,
consumers, versions, and configurations. Runtime behavior, storage compatibility,
and security are not assessed. Deployment compatibility is unknown. Demand and
adoption are unvalidated. Results must not be generalized into universal
safe/breaking claims.

## Install and build

```bash
go install github.com/otrumb/abi-evolution-oracle/cmd/abi-evolution-oracle@latest
```

For a pinned source checkout:

```bash
export GOTOOLCHAIN=local
export CGO_ENABLED=0
go build -trimpath -o abi-evolution-oracle ./cmd/abi-evolution-oracle
```

PowerShell uses `$env:GOTOOLCHAIN='local'` and `$env:CGO_ENABLED='0'`.

## Commands

Run commands from repository root:

```bash
go run ./cmd/abi-evolution-oracle verify-corpus
go run ./cmd/abi-evolution-oracle score evidence evidence/gates.json
go run ./cmd/abi-evolution-oracle probe-one D06 /path/to/abigen
go run ./cmd/abi-evolution-oracle run /tmp/evidence /path/to/abidiff /path/to/abigen
```

- `verify-corpus` validates corpus shape, hashes, provenance, and expectations.
- `score` recomputes `score.json` from recorded observations, baseline records,
  corpus verification, and gate input.
- `probe-one` executes one fixture with a supplied `abigen` v1.15.11 binary.
- `run` executes all probes and the supplied pinned baseline into a chosen
  evidence directory. It performs no live RPC.

Exit code `0` means command success, `1` means execution or validation failure,
and `2` means invalid command usage.

## Reproduce checked-in evidence

Build `abigen` from go-ethereum v1.15.11 and check out `abidiff` at the exact
commit above. Its CLI entry point is `src/cli.js` and needs Node 18 or newer.

```bash
GOTOOLCHAIN=local CGO_ENABLED=0 go build -trimpath -o .tools/abigen github.com/ethereum/go-ethereum/cmd/abigen
git clone --no-checkout https://github.com/jay-tank/abidiff.git .tools/abidiff
git -C .tools/abidiff checkout c9488370c95cd0f9406d3fbec6aa26cc95c1f3b6
node .tools/abidiff/src/cli.js --help
go run ./cmd/abi-evolution-oracle run evidence-replay .tools/abidiff/src/cli.js .tools/abigen
go run ./cmd/abi-evolution-oracle score evidence evidence/gates.json
```

On Windows, use `.exe` for the `abigen` output. CI performs bounded full
baseline replay on Ubuntu only. No live RPC, credentials, provider allowlist,
or chain allowlist is used.

See [`evidence/final-verdict.md`](evidence/final-verdict.md),
[`docs/tdd-evidence.md`](docs/tdd-evidence.md), and
[`docs/correction-log.md`](docs/correction-log.md) for evidence and correction
history. Technical closure does not erase earlier corrections or historical
publication-lock semantics.

## License

Project source is MIT licensed; see [`LICENSE`](LICENSE). Corpus files under
`corpus/` are separately dedicated under CC0-1.0; see
[`corpus/LICENSE-CC0-1.0.txt`](corpus/LICENSE-CC0-1.0.txt). Third-party tools and
dependencies retain their own licenses and ownership.
