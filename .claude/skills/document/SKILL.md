---
name: document
description: Write or review documentation (Go doc comments) for code in this repo. Picks the SOP that matches the package being documented. Use when asked to document, add comments to, or review the docs of any package or file, or when adding exported identifiers that need documentation.
---

# Documenting code

This skill is a router. Do not document from memory: first identify the target package, then follow its SOP.

## 1. Identify the target

Resolve the target from, in order: the path(s) the user named, the files being changed in the current task, then the current working directory. If the target spans several packages, handle each with its own SOP.

## 2. Select the SOP

| Target path | SOP to read and follow |
|---|---|
| `usecase/` | [sops/usecase.md](sops/usecase.md) |
| `internal/` (all subpackages) | [sops/internal.md](sops/internal.md) |

If the target has no SOP:
1. Apply only the **Common rules** below.
2. Tell the user no package-specific SOP exists and offer to create one by copying `sops/usecase.md`'s structure (scope, what to document, style, checklist), then add a row to the table above.

## 3. Common rules (apply to every package)

- Go doc convention: comment starts with the identifier name, English, full sentences.
- Default: document intent and behavior, not implementation tools — never name the library, database, broker or service behind an interface. A package SOP may override this (e.g. `internal/` adapters name their technology).
- Comment-only changes: do not alter logic, signatures, log messages, or tests (`*_test.go` are never documented).
- Types follow the project's grouped `type ( ... )` convention; place each comment on its spec inside the block.
- Read the code before describing it; only state facts the code implements.
- Finish with `gofmt -l`, `go vet` and `go test` on the package; the SOP adds package-specific checks.

## Adding a new SOP

Create `sops/<package>.md` with: scope, what to document, vocabulary/style specific to the package, and a verification checklist. Register it in the table in step 2.
