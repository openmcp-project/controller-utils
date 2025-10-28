# Image Parsing

The `pkg/image` package provides utilities for parsing container image references into their constituent components.

## ParseImage Function

The `ParseImage` function parses a container image string and extracts the image name, tag, and digest components. This is useful for validating, manipulating, or analyzing container image references in Kubernetes controllers.

### Function Signature

```go
func ParseImage(image string) (imageName string, tag string, digest string, err error)
```

### Behavior

- **Default Tag**: If no tag is specified, it defaults to `"latest"`
- **Digest Support**: Handles images with SHA256 digests (indicated by `@sha256:...`)
- **Registry URLs**: Properly handles registry URLs with ports (e.g., `registry.io:5000/image:tag`)
- **Validation**: Returns an error for empty image strings

### Examples

```go
// Basic image with tag
imageName, tag, digest, err := ParseImage("nginx:1.19.0")
// Returns: "nginx", "1.19.0", "", nil

// Image without tag (defaults to latest)
imageName, tag, digest, err := ParseImage("nginx")
// Returns: "nginx", "latest", "", nil

// Image with digest only
imageName, tag, digest, err := ParseImage("nginx@sha256:abcdef...")
// Returns: "nginx", "", "sha256:abcdef...", nil

// Image with both tag and digest
imageName, tag, digest, err := ParseImage("nginx:1.19.0@sha256:abcdef...")
// Returns: "nginx", "1.19.0", "sha256:abcdef...", nil
```

This function is particularly useful when working with container images in Kubernetes controllers, allowing you to extract and validate image components for further processing or validation.