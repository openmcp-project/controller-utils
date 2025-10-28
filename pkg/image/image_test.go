package image

import (
	"testing"
)

func TestParseImage(t *testing.T) {
	tests := []struct {
		name              string
		image             string
		expectedImageName string
		expectedTag       string
		expectedDigest    string
		expectError       bool
	}{
		{
			name:              "image with tag only",
			image:             "nginx:1.21.0",
			expectedImageName: "nginx",
			expectedTag:       "1.21.0",
			expectedDigest:    "",
			expectError:       false,
		},
		{
			name:              "image with tag and digest",
			image:             "nginx:1.21.0@sha256:abc123def456",
			expectedImageName: "nginx",
			expectedTag:       "1.21.0",
			expectedDigest:    "sha256:abc123def456",
			expectError:       false,
		},
		{
			name:              "image with latest tag",
			image:             "ubuntu:latest",
			expectedImageName: "ubuntu",
			expectedTag:       "latest",
			expectedDigest:    "",
			expectError:       false,
		},
		{
			name:              "image with version tag and digest",
			image:             "registry.io/myapp:v2.1.3@sha256:fedcba987654",
			expectedImageName: "registry.io/myapp",
			expectedTag:       "v2.1.3",
			expectedDigest:    "sha256:fedcba987654",
			expectError:       false,
		},
		{
			name:              "image without explicit tag (defaults to latest)",
			image:             "nginx",
			expectedImageName: "nginx",
			expectedTag:       "latest",
			expectedDigest:    "",
			expectError:       false,
		},
		{
			name:              "empty image string",
			image:             "",
			expectedImageName: "",
			expectedTag:       "",
			expectedDigest:    "",
			expectError:       true,
		},
		{
			name:              "image with multiple colons in name",
			image:             "registry.io:5000/namespace/image:v1.0.0",
			expectedImageName: "registry.io:5000/namespace/image",
			expectedTag:       "v1.0.0",
			expectedDigest:    "",
			expectError:       false,
		},
		{
			name:              "image with multiple colons and digest",
			image:             "registry.io:5000/namespace/image:v1.0.0@sha256:123456789abc",
			expectedImageName: "registry.io:5000/namespace/image",
			expectedTag:       "v1.0.0",
			expectedDigest:    "sha256:123456789abc",
			expectError:       false,
		},
		{
			name:              "image with digest only (no tag)",
			image:             "nginx@sha256:abc123def456",
			expectedImageName: "nginx",
			expectedTag:       "",
			expectedDigest:    "sha256:abc123def456",
			expectError:       false,
		},
		{
			name:              "complex registry with namespace and tag",
			image:             "ghcr.io/openmcp-project/components/github.com/openmcp-project/openmcp:v0.0.11",
			expectedImageName: "ghcr.io/openmcp-project/components/github.com/openmcp-project/openmcp",
			expectedTag:       "v0.0.11",
			expectedDigest:    "",
			expectError:       false,
		},
		{
			name:              "image with port and path",
			image:             "localhost:5000/my-namespace/my-image:1.2.3",
			expectedImageName: "localhost:5000/my-namespace/my-image",
			expectedTag:       "1.2.3",
			expectedDigest:    "",
			expectError:       false,
		},
		{
			name:              "image with port, path and digest",
			image:             "localhost:5000/my-namespace/my-image:1.2.3@sha256:abcdef123456",
			expectedImageName: "localhost:5000/my-namespace/my-image",
			expectedTag:       "1.2.3",
			expectedDigest:    "sha256:abcdef123456",
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imageName, tag, digest, err := ParseImage(tt.image)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if imageName != tt.expectedImageName {
				t.Errorf("expected image name %q, got %q", tt.expectedImageName, imageName)
			}

			if tag != tt.expectedTag {
				t.Errorf("expected tag %q, got %q", tt.expectedTag, tag)
			}

			if digest != tt.expectedDigest {
				t.Errorf("expected digest %q, got %q", tt.expectedDigest, digest)
			}
		})
	}
}
