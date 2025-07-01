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
