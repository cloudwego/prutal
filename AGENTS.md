# AGENTS.md

## Compatibility

- Keep the root module compatible with Go 1.20.
  `tests` is a separate module that requires Go 1.25.
- Public compatibility and unsupported features are tracked in `README.md`; update that table when behavior changes.

## Architecture constraints

- Preserve the descriptor design: parse tags once, cache `StructDesc` by type, and resolve coder functions during finalization rather than on every encode or decode.
- For every new unsafe layout assumption, document it, extend `internal/hack`, and consider both 64-bit and 32-bit behavior.
- Validate the parsed model before generator emission and keep generated Go compact.
- `internal/protowire`, `prutalgen/internal/antlr`, `prutalgen/internal/parser`, and `prutalgen/internal/protobuf` contain generated, ported, or vendored code.
  Preserve their existing license headers and avoid unrelated reformatting or refactoring.
- New first-party Go and assembly files need the Apache-2.0 header enforced by `.licenserc.yaml`.

## Generated files

- Regenerate tracked `internal/prutal/testdata.pb.go` and `internal/prutal/testdata_edition2023.pb.go` with `internal/prutal/testdata.sh`.
  The script restores their required Apache-2.0 headers after generation.
- Regenerate the ANTLR parser only for an intentional parser upgrade, using `prutalgen/internal/update_parser.sh` with a local `antlr` executable.

## Verification

Run the subset appropriate to the change:

```sh
go test ./...
go test -race ./...
GOOS=linux GOARCH=386 go build ./...
cd tests && make test
go test -bench=. -benchmem -run=none ./...
```

The integration suite requires network access, downloads the pinned `protoc`, and installs the pinned `protoc-gen-go` tools.
Use the 386 build for runtime, descriptor, or unsafe-layout changes; use the integration suite for generator and protobuf-compatibility changes; use benchmarks for performance-sensitive runtime changes.

## Git

Follow the Angular-style commit subject convention required by `CONTRIBUTING.md`.

## Pinned reference implementation

- Repository: https://github.com/protocolbuffers/protobuf-go
- Commit: `cdd4c5f7406e82462949c7a65defa9f3029c162d`

Use this revision to resolve protobuf semantics, wire-format edge cases, error behavior, and generated-code expectations.
Change the pin only for a compatibility reason.
