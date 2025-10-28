package image

import (
	"fmt"
	"strings"
)

// ParseImage parses a container image string and returns the image name, tag, and digest.
// If no tag is specified, it defaults to "latest". If a digest is present, it is returned as well.
// Examples of valid image strings:
// - "nginx:1.19.0" -> imageName: "nginx", tag: "1.19.0", digest: ""
// - "nginx" -> imageName: "nginx", tag: "latest", digest: ""
// - "nginx@sha256:abcdef..." -> imageName: "nginx", tag: "", digest: "sha256:abcdef..."
// - "nginx:1.19.0@sha256:abcdef..." -> imageName: "nginx", tag: "1.19.0", digest: "sha256:abcdef..."
func ParseImage(image string) (imageName string, tag string, digest string, err error) {
	if image == "" {
		return "", "", "", fmt.Errorf("image string cannot be empty")
	}

	// Check if the image contains a digest (indicated by @)
	digestIndex := strings.LastIndex(image, "@")

	var tagPart string
	if digestIndex != -1 {
		// Extract digest
		digest = image[digestIndex+1:]
		tagPart = image[:digestIndex]
	} else {
		tagPart = image
	}

	// Find the last colon to separate the tag from the image name
	// We use LastIndex to handle registry URLs with ports (e.g., registry.io:5000/image:tag)
	colonIndex := strings.LastIndex(tagPart, ":")

	// If there's a digest but no colon in the tag part, it's a digest-only image
	if digestIndex != -1 && colonIndex == -1 {
		imageName = tagPart
		return imageName, "", digest, nil
	}

	// If there's no colon, it means no explicit tag was provided
	// In this case, default to "latest" tag
	if colonIndex == -1 {
		imageName = tagPart
		return imageName, "latest", digest, nil
	}

	// Extract image name (everything before the last colon)
	imageName = tagPart[:colonIndex]
	// Extract tag (everything after the last colon)
	tag = tagPart[colonIndex+1:]

	return imageName, tag, digest, nil
}
