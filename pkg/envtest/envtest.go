package envtest

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	errFailedToGetWD           = errors.New("failed to get working directory")
	errMakefileIsDir           = errors.New("expected Makefile to be a file but it is a directory")
	errMakefileNotFound        = errors.New("reached fs root and did not find Makefile")
	errFailedToRunMake         = errors.New("failed to run make")
	errFailedToRunSetupEnvtest = errors.New("failed to run setup-envtest")
)

// Install uses make to install the envtest dependencies and sets the
// KUBEBUILDER_ASSETS environment variable.
// k8sVersion is the version of Kubernetes to install, e.g. "1.30.0" or "latest".
func Install(k8sVersion string) error {
	wd, err := os.Getwd()
	if err != nil {
		return errors.Join(errFailedToGetWD, err)
	}

	makefilePath, err := findMakefile(wd)
	if err != nil {
		return err
	}
	repoDir := filepath.Dir(makefilePath)

	if err := runMakeEnvtest(repoDir); err != nil {
		return err
	}

	assetsDir, err := runSetupEnvtest(repoDir, k8sVersion)
	if err != nil {
		return err
	}

	return os.Setenv("KUBEBUILDER_ASSETS", assetsDir)
}

func runMakeEnvtest(repoDir string) error {
	cmd := exec.Command("make", "envtest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		return errors.Join(errFailedToRunMake, err)
	}
	return nil
}

func runSetupEnvtest(workingDir, k8sVersion string) (string, error) {
	binDir := filepath.Join(workingDir, "bin")
	binary := filepath.Join(binDir, "setup-envtest")
	cmd := exec.Command(binary, "use", k8sVersion, "--bin-dir", binDir, "-p", "path")
	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = workingDir
	if err := cmd.Run(); err != nil {
		return "", errors.Join(errFailedToRunSetupEnvtest, err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func findMakefile(root string) (string, error) {
	if !filepath.IsAbs(root) {
		var err error
		if root, err = filepath.Abs(root); err != nil {
			return "", err
		}
	}

	if root == "/" {
		return "", errMakefileNotFound
	}

	makefilePath := filepath.Join(root, "Makefile")
	finfo, err := os.Stat(makefilePath)
	if errors.Is(err, fs.ErrNotExist) {
		parent := filepath.Dir(root)
		return findMakefile(parent)
	}
	if err != nil {
		return "", err
	}
	if finfo.IsDir() {
		return "", errMakefileIsDir
	}

	return makefilePath, nil
}
