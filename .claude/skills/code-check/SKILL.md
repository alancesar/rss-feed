---
name: code-check
description: Check Go code in this repo against the project's style directives (declaration order, grouped type blocks, function and method order, one-line error handling). Use when asked to check, lint or review code style, before committing Go changes, or whenever writing new Go code in this repo.
---

# Code check

Apply every directive below to the target: the path(s) the user named, otherwise the Go files changed in the current task (`git diff --name-only` plus untracked `*.go`). Report each violation as `file:line — directive — fix`, then fix it if the user asked for fixes or the code is being written in this task.

## Directives

### 1. Top-level declaration order

Package-level declarations go at the top of the file, right after the imports, in this order, each as one grouped block:

```go
const (
    ...
)

var (
    // usually sentinel errors
    ErrSomething = errors.New("something")
)

type (
    ...
)

// funcs and methods follow
```

- Omit a block that has nothing in it. Never repeat a block kind (one `const (...)`, one `var (...)`, one `type (...)` per file).
- Use the grouped form even for a single declaration.
- Never declare types "on demand" between functions (`type A struct{...}`, then some funcs, then `type B struct{...}`). Every type in the file belongs in the single `type ( ... )` block at the top.
- This applies to `*_test.go` files too.

### 2. Function and method order

After the declaration blocks, order the file's funcs like this:

1. **Constructors**: `New...` functions that build a type declared in the file, right after the `type ( ... )` block.
2. **Exported methods**: functions with a receiver, `func (s Struct) Method()`.
3. **Unexported methods**: `func (s Struct) method()`.
4. **Exported functions**: no receiver, `func Function()`.
5. **Unexported functions**: `func function()`.

```go
type (
    HandleFeed struct { ... }
)

func NewHandleFeed(...) *HandleFeed { ... }            // 1. constructor

func (uc HandleFeed) Execute(ctx context.Context) error // 2. exported method

func (uc HandleFeed) handleFeed(...) error              // 3. unexported method

func Parse(...)                                         // 4. exported function

func validate(...)                                      // 5. unexported function
```

- When a file has methods on several types, keep each type's methods together, in the order the types are declared, with exported before unexported within each type.
- Go exempts `init()` and `main()` from these rules: put them after the constructors.
- In `*_test.go` files, `Test...` functions count as exported functions. Test-double methods (stubs and fakes) come first, as methods.

### 3. One-line error handling

When a call returns only `error`, handle it in the `if` statement:

```go
// correct
if err := uc.store.Save(ctx, a); err != nil {
    return err
}

// incorrect
err := uc.store.Save(ctx, a)
if err != nil {
    return err
}
```

**Check the scope before applying.** The short form scopes `err` to the `if` block. Any code after the block that reads `err` silently refers to an outer `err` instead, or fails to compile. When the error needs further handling (classifying it, acking or nacking a delivery), move every use of `err` inside the block and leave the block early with `return` or `continue`:

```go
// correct: every use of err is inside the block
if err := uc.handleFeed(ctx, delivery.Payload); err != nil {
    if errors.Is(err, ErrInvalidFeed) {
        delivery.Nack(false)
        continue
    }

    delivery.Nack(true)
    continue
}

delivery.Ack()

// incorrect: the second if reads the outer err (e.g. from an earlier Subscribe), which is nil
if err := uc.handleFeed(ctx, delivery.Payload); err != nil {
    logger.Error().Err(err).Msg("failed to handle feed")
}
if errors.Is(err, ErrInvalidFeed) {
    delivery.Nack(false)
    continue
}
```

If the uses of `err` can't all be moved inside the block, keep the two-line form.

Calls that return a value plus an error (`v, err := f()`) keep the two-line form.

## Verification

Run `gofmt -l`, `go vet` and `go test` on the affected packages. A style fix must not change behaviour, so the tests must stay green.
