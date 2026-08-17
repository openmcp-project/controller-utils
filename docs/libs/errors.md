# Error Handling

## ReasonableError

The `pkg/errors` package contains the `ReasonableError` type, which combines a normal `error` with a reason string. This is useful for errors that happen during reconciliation for updating the resource's status with a reason for the error later on.

### Noteworthy Functions:

- `WithReason(...)` can be used to wrap a standard error together with a reason into a `ReasonableError`.
- `Errorf(...)` can be used to wrap an existing `ReasonableError` together with a new error, similarly to how `fmt.Errorf(...)` does it for standard errors.
- `NewReasonableErrorList(...)` or `Join(...)` can be used to work with lists of errors. `Aggregate()` turns them into a single error again.

## Ignoring Invalid User Input

`ErrInvalidUserInput` can be used to create errors that can be ignored with `IgnoreInvalidUserInput(...)` at the end of reconciliation. This allows you to handle errors properly and skip unnecessary requeues.

## Templating Errors

The `TemplateErrorBuilder` from `pkg/errors` can be used to construct more useful errors out of the default golang templating errors.

The standard error message for a missing key looks like this:
```
foo: bar
asdf: template: mytemplate:2:16: executing "mytemplate" at <.inputs.asdf>: map has no entry for key "inputs"
```

With the enhanced templating error logic, it can be turned into this:
```
template: mytemplate:2:16: executing "mytemplate" at <.inputs.asdf>: map has no entry for key "inputs"
template source:
1:  foo: bar
2:  asdf: {{ .inputs.asdf }}
                    ˆ≈≈≈≈≈≈≈
3:  test: true

template input:
	foo: "bar"
	asdf: "asdf"
```

Use `TemplateErrorBuilder(<original error>)` to create the builder, then add more context with these builder methods:
  - `WithSource` adds the source template.
  - `WithSourceSnippetFormat` allows to overwrite the default values for
    - lines before the error (default: `5`)
    - lines after the error (default: `5`)
    - length of the prefix (line number, colon, following whitespace) (default: `6`)
  - `WithFormattedOutput` can be used if the templating itself was successful, but the generated output does not match the expected format or is wrong in some other way.
  - `WithInput` sets the template's input values.
Then use the `Build()` method to generate the error message.
