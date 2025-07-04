# JSON Patch

The `api/jsonpatch` package contains a `JSONPatches` type that represents a [JSON Patch](https://datatracker.ietf.org/doc/html/rfc6902).
The type is ready to be used in a kubernetes resource type.

The corresponding `pkg/jsonpatch` package contains helper functions to apply JSON patches specified via the aforementioned API type to a given JSON document or arbitrary go type.

## Embedding the API Type

```golang
import jpapi "github.com/openmcp-project/controller-utils/api/jsonpatch"

type MyTypeSpec struct {
  Patches jpapi.JSONPatches `json:"patches"`
}
```

## Patch Syntax

The `pkg/jsonpatch` package handles JSON patches in form of the `JSONPatches` type from the `api/jsonpatch` package. The type can be safely embedded in k8s resources.
```yaml
patches:
- op: add
  path: /foo/bar
  value: foobar
- op: copy
  path: /foo/baz
  from: /foo/bar
```

`op` and `path` are required for each patch, `value` and `from` depend on the chosen operation.
Valid operations are `add`, `remove`, `replace`, `move`, `copy`, and `test`.

### Path Notation

The are two options for the notation of the `path` attribute:

#### JSON Pointer Notation

The first option is the JSON Pointer notation as described in [RFC 6901](https://datatracker.ietf.org/doc/html/rfc6901). Basically, each path segment is prefixed with `/`, with no differentiation between object fields and array indices. 

There are two special characters which need to be substituted:
- `~` has to be written as `~0`
- `/` has to be written as `~1`

Examples:
- `/foo/bar`
- `/mylist/0/asdf`
- `/metadata/annotations/foo.bar.baz~1foobar`

#### JSON Path Notation

The second option is a simplified variant of the JSON Path notation as described in [RFC 9535](https://datatracker.ietf.org/doc/html/rfc9535). While the RFC specifies a full query language with function evaluation, the implementation here just allows referencing a single path.

In short:
- path segments are separated by `.` or by using `[...]`
  - the leading `.` is optional
- backslashes `\` are used to escape the special characters `\`, `.`, `[`, `]`, `'`, and `"`
- if the bracket notation is used to separate a path segment, single `'` or double `"` quotes may be used within the brackets
  - the quote character has to immediately follow the opening bracket and immediately precede the closing bracket
  - no escaping is required within brackets with quotes

The table below shows a few examples of paths in the JSON Path notation and the corresponding JSON Pointer notation they are converted into.
| JSON Path Notation | JSON Pointer Notation |
| --- | --- |
| `.metadata.annotations.foo\.bar\.baz/foobar` | `/metadata/annotations/foo.bar.baz~1foobar` |
| `metadata.annotations.foo\.bar\.baz/foobar` | `/metadata/annotations/foo.bar.baz~1foobar` |
| `.metadata.annotations[foo\.bar\.baz/foobar]` | `/metadata/annotations/foo.bar.baz~1foobar` |
| `metadata.annotations["foo.bar.baz/foobar"]` | `/metadata/annotations/foo.bar.baz~1foobar` |
| `.metadata[annotations]['foo.bar.baz/foobar']` | `/metadata/annotations/foo.bar.baz~1foobar` |
| `.mylist[0].asdf` | `/mylist/0/asdf` |
| `mylist.0.asdf` | `/mylist/0/asdf` |

## Applying the Patches

### To a JSON Document

```golang
import "github.com/openmcp-project/controller-utils/pkg/jsonpatch"

// mytype.Spec is of type MyTypeSpec as defined in the above example
patch := jsonpatch.New(mytype.Spec.Patches)
// doc and modified are of type []byte
modified, err := patch.Apply(doc)
```

### To an Arbitrary Type

The library supports applying JSON patches to arbitrary types. Internally, the object is marshalled to JSON, then the patch is applied, and then the object is unmarshalled into its original type again. The usual limitations of JSON (un)marshalling (no cyclic structures, etc.) apply.

```golang
import "github.com/openmcp-project/controller-utils/pkg/jsonpatch"

// mytype.Spec is of type MyTypeSpec as defined in the above example
patch := jsonpatch.NewTyped[MyPatchedType](mytype.Spec.Patches)
// obj and modified are of type MyPatchedType
modified, err := patch.Apply(doc)
```

### Options

The `Apply` method optionally takes some options which can be constructed from functions contained in the package:
```golang
modified, err := patch.Apply(doc, jsonpatch.Indent("  "))
```

The available options are:
- `SupportNegativeIndices`
- `AccumulatedCopySizeLimit`
- `AllowMissingPathOnRemove`
- `EnsurePathExistsOnAdd`
- `EscapeHTML`
- `Indent`

The options are simply passed into the [library which is used internally](https://github.com/evanphx/json-patch).
