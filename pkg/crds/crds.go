package crds

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"

	"github.com/openmcp-project/controller-utils/pkg/clusters"
	"github.com/openmcp-project/controller-utils/pkg/controller"
	"github.com/openmcp-project/controller-utils/pkg/errors"
	"github.com/openmcp-project/controller-utils/pkg/logging"
	"github.com/openmcp-project/controller-utils/pkg/resources"
)

type (
	CRDLabelToClusterMappings map[string]*clusters.Cluster
	CRDList                   func() ([]*apiextv1.CustomResourceDefinition, error)
)

type CRDManager struct {
	mappingLabelName           string
	crdLabelsToClusterMappings CRDLabelToClusterMappings
	crdList                    CRDList
}

func NewCRDManager(mappingLabelName string, crdList CRDList) *CRDManager {
	return &CRDManager{
		mappingLabelName:           mappingLabelName,
		crdLabelsToClusterMappings: make(CRDLabelToClusterMappings),
		crdList:                    crdList,
	}
}

func (m *CRDManager) AddCRDLabelToClusterMapping(labelValue string, cluster *clusters.Cluster) {
	m.crdLabelsToClusterMappings[labelValue] = cluster
}

func (m *CRDManager) CreateOrUpdateCRDs(ctx context.Context, log *logging.Logger) error {
	crds, err := m.crdList()
	if err != nil {
		return fmt.Errorf("error getting CRDs: %w", err)
	}

	var errs error

	for _, crd := range crds {
		c, err := m.getClusterForCRD(crd)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		if log != nil {
			log.Info("creating/updating CRD", "name", crd.Name, "cluster", c.ID())
		}
		err = resources.CreateOrUpdateResource(ctx, c.Client(), resources.NewCRDMutator(crd, crd.Labels, crd.Annotations))
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return fmt.Errorf("error creating/updating CRDs: %w", errs)
	}
	return nil
}

func (m *CRDManager) getClusterForCRD(crd *apiextv1.CustomResourceDefinition) (*clusters.Cluster, error) {
	labelValue, ok := controller.GetLabel(crd, m.mappingLabelName)
	if !ok {
		return nil, fmt.Errorf("missing label '%s' for CRD '%s'", m.mappingLabelName, crd.Name)
	}

	cluster, ok := m.crdLabelsToClusterMappings[labelValue]
	if !ok {
		return nil, fmt.Errorf("no cluster mapping found for label value '%s' in CRD '%s'", labelValue, crd.Name)
	}

	return cluster, nil
}

// CRDsFromFileSystem reads CRDs from the specified filesystem path.
func CRDsFromFileSystem(fsys fs.FS, path string) ([]*apiextv1.CustomResourceDefinition, error) {
	var crds []*apiextv1.CustomResourceDefinition

	entries, err := fs.ReadDir(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(path, entry.Name())
		data, err := fs.ReadFile(fsys, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		var crd apiextv1.CustomResourceDefinition
		if err := yaml.Unmarshal(data, &crd); err != nil {
			return nil, fmt.Errorf("failed to unmarshal CRD from file %s: %w", filePath, err)
		}

		crds = append(crds, &crd)
	}

	return crds, nil
}
