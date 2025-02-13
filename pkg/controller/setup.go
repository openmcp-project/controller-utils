package controller

import (
	"fmt"
	"os"
	"path"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// LoadKubeconfig loads a cluster configuration from the given path.
// If the path points to a single file, this file is expected to contain a kubeconfig which is then loaded.
// If the path points to a directory which contains a file named "kubeconfig", that file is used.
// If the path points to a directory which does not contain a "kubeconfig" file, there must be "host", "token", and "ca.crt" files present,
// which are used to configure cluster access based on an OIDC trust relationship.
// If the path is empty, the in-cluster config is returned.
func LoadKubeconfig(configPath string) (*rest.Config, error) {
	if configPath == "" {
		return rest.InClusterConfig()
	}
	fi, err := os.Stat(configPath)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		if kfi, err := os.Stat(path.Join(configPath, "kubeconfig")); err == nil && !kfi.IsDir() {
			// there is a kubeconfig file in the specified folder
			// point configPath to the kubeconfig
			configPath = path.Join(configPath, "kubeconfig")
		} else {
			// no kubeconfig file present, load OIDC trust configuration
			host, err := os.ReadFile(path.Join(configPath, "host"))
			if err != nil {
				return nil, fmt.Errorf("error reading host file: %w", err)
			}
			return &rest.Config{
				Host:            string(host),
				BearerTokenFile: path.Join(configPath, "token"),
				TLSClientConfig: rest.TLSClientConfig{
					CAFile: path.Join(configPath, "ca.crt"),
				},
			}, nil
		}
	}
	// at this point, configPath points to a single file which is expected to contain a kubeconfig
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	return clientcmd.RESTConfigFromKubeConfig(data)
}
