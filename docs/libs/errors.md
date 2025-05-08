# Error Handling

The `pkg/errors` package contains the `ReasonableError` type, which combines a normal `error` with a reason string. This is useful for errors that happen during reconciliation for updating the resource's status with a reason for the error later on.

### Noteworthy Functions:

- `WithReason(...)` can be used to wrap a standard error together with a reason into a `ReasonableError`.
- `Errorf(...)` can be used to wrap an existing `ReasonableError` together with a new error, similarly to how `fmt.Errorf(...)` does it for standard errors.
- `NewReasonableErrorList(...)` or `Join(...)` can be used to work with lists of errors. `Aggregate()` turns them into a single error again.

