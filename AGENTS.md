# AGENTS.md

Rules for changing this repository. They describe how the code is organised
today, not an aspiration: a change that breaks them needs a reason in the PR.

## What this is

A self-hosted web archive: an HTTP API adds a page by URL, a worker archives it
in the requested formats (headers, PDF, single HTML file) and the results are
stored in badger. Go 1.26, `devenv` provides the toolchain.

## Layers

| Layer | Path | Rule |
|---|---|---|
| domain | `internal/domain` | standard library and `google/uuid` only, no adapters, no config, no logger |
| application | `internal/application` | use cases and the ports they need; imports the domain and `zap`, never an adapter, config or transport |
| infrastructure | `internal/infrastructure` | adapters: badger storage, processors, queue, logging |
| interfaces | `internal/interfaces` | transports: the generated REST handler, the static web UI, the router |
| composition | `cmd/service` | builds every implementation and passes it inward; nothing imports it |

Dependencies point inward: interfaces and infrastructure may import
application and domain, the reverse never happens. `golangci-lint` enforces this
with `depguard` rules in `.golangci.yaml`.

Where things go:

- a new archive format → implement `processor` in
  `internal/infrastructure/processors`, register it in `NewProcessors`, add the
  value to `internal/domain/format.go` and to the `format` enum in
  `internal/api/openapi.yaml`;
- a new storage backend → implement `application.PageRepository` and wire it in
  `cmd/service/main.go`;
- a new endpoint → change `internal/api/openapi.yaml`, run
  `go generate ./internal/api`, implement it in `internal/interfaces/rest`;
  the handler only maps types and errors, the rule belongs in a use case.

## The stored format is a contract

A page is stored as one msgpack record (`internal/infrastructure/storage/badger`).
Renaming or removing a field of `domain.Page`/`domain.PageBase`/`domain.Result`/
`domain.File` silently changes what users already have on disk.

- `internal/infrastructure/storage/badger/testdata/page_v1.msgpack` pins the
  format; `TestMarshalFixture` fails on any change to it.
- For a deliberate change: add the mapping in the storage adapter so old records
  still read, then regenerate the fixture with `UPDATE_FIXTURE=1 go test
  ./internal/infrastructure/storage/badger -run Fixture` and say so in the PR.

## Checks before committing

```shell
go build ./...
go vet ./...
go vet -tags integration ./...
go test -race ./...
golangci-lint run ./...
```

Tests must not reach the network: use `httptest`. A test that needs an external
resource or a binary goes behind the `integration` build tag.

## Code conventions

- Wrap errors with context: `fmt.Errorf("save page: %w", err)`.
- Missing page or file is `domain.ErrNotFound`; the storage adapter maps
  badger's `ErrKeyNotFound` to it and the transport maps it to `404`.
  Invalid input is `application.ValidationError` and becomes `400`.
- The domain has no I/O: fetching, files, clocks and IDs come in from outside.
  New page IDs and timestamps are injected (`Service.newID`, `Service.now`) so
  use cases stay testable.
- Keep the logging library and msgpack out of the domain and the use cases.
- Match the surrounding style; no broad refactors in a feature commit.
