package crds

import (
	"embed"
	"os"
	"path"

	"k8s.io/apimachinery/pkg/types"
)

func readAllFiles(fs embed.FS, dir string) ([][]byte, error) {
	fileContents := [][]byte{}

	entries, err := fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		fullpath := path.Join(dir, entry.Name())
		if entry.IsDir() {
			subdirContents, err := readAllFiles(fs, fullpath)
			if err != nil {
				return nil, err
			}
			fileContents = append(fileContents, subdirContents...)
			continue
		}

		content, err := fs.ReadFile(fullpath)
		if err != nil {
			return nil, err
		}
		fileContents = append(fileContents, content)
	}

	return fileContents, nil
}

func getWebhookServiceFromEnv() types.NamespacedName {
	return types.NamespacedName{
		Name:      os.Getenv("WEBHOOK_SERVICE_NAME"),
		Namespace: os.Getenv("WEBHOOK_SERVICE_NAMESPACE"),
	}
}

func getWebhookSecretFromEnv() types.NamespacedName {
	return types.NamespacedName{
		Name:      os.Getenv("WEBHOOK_SECRET_NAME"),
		Namespace: os.Getenv("WEBHOOK_SECRET_NAMESPACE"),
	}
}
