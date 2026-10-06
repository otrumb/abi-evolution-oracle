# ABI Evolution Oracle Validation

Local-only pre-code experiment measuring consumer compatibility with
go-ethereum v1.15.11. Corpus is original synthetic work dedicated under CC0-1.0.

No runtime behavior, storage compatibility, security, deployment compatibility,
generic decoder behavior, or universal indexer behavior is assessed.

Publication remains locked regardless of technical verdict.

```powershell
$env:GOTOOLCHAIN='local'
$env:CGO_ENABLED='0'
go run ./cmd/abi-oracle verify-corpus
go run ./cmd/abi-oracle run -evidence evidence
```
