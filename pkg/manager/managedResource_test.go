package manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeWriter is a test implementation of ResourceStatusWriter that records
// what ProjectResources sets on it.
type fakeWriter struct {
	ref     ResourceRef
	phase   string
	message string
}

func (w *fakeWriter) SetReference(ref ResourceRef) { w.ref = ref }
func (w *fakeWriter) SetPhase(phase, message string) {
	w.phase = phase
	w.message = message
}

func TestProjectResources(t *testing.T) {
	t.Run("returns one writer per result with reference and phase/message applied", func(t *testing.T) {
		results := []ManagedResource{
			{
				APIGroup:  "source.toolkit.fluxcd.io",
				Kind:      "OCIRepository",
				Name:      "eso",
				Namespace: "tenant-ns",
				Location:  PlatformCluster,
				Status:    ManagedResourceStatus{Phase: StatusPhaseReady, Message: "Resource exists."},
			},
			{
				Kind:     "Secret",
				Name:     "pull-secret",
				Location: ManagedControlPlane,
				Status:   ManagedResourceStatus{Phase: StatusPhaseProgressing, Message: "creating"},
			},
		}

		out := ProjectResources(results, func() *fakeWriter { return &fakeWriter{} })

		require.Len(t, out, 2)

		assert.Equal(t, ResourceRef{
			APIGroup:  "source.toolkit.fluxcd.io",
			Kind:      "OCIRepository",
			Name:      "eso",
			Namespace: "tenant-ns",
			Location:  string(PlatformCluster),
		}, out[0].ref)
		assert.Equal(t, StatusPhaseReady, out[0].phase)
		assert.Equal(t, "Resource exists.", out[0].message)

		assert.Equal(t, ResourceRef{
			Kind:     "Secret",
			Name:     "pull-secret",
			Location: string(ManagedControlPlane),
		}, out[1].ref)
		assert.Equal(t, StatusPhaseProgressing, out[1].phase)
		assert.Equal(t, "creating", out[1].message)
	})

	t.Run("returns empty non-nil slice for no results", func(t *testing.T) {
		out := ProjectResources(nil, func() *fakeWriter { return &fakeWriter{} })
		assert.NotNil(t, out)
		assert.Empty(t, out)
	})
}
