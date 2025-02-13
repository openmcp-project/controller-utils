package controller

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// HasOwnerReference returns the index of the owner reference if the 'owned' object has a owner reference pointing to the 'owner' object.
// If not, -1 is returned.
// Note that name and uid are only compared if set in the owner object. This means that the function will return a positive index
// for an owner object with empty name and uid if the owned object contains a owner reference which fits just apiversion and kind.
// The scheme argument may be nil if the owners GVK is populated.
func HasOwnerReference(owned, owner client.Object, scheme *runtime.Scheme) (int, error) {
	if owned == nil || owner == nil {
		return -1, fmt.Errorf("neither dependent nor owner may be nil when checking for owner references")
	}
	if owner.GetNamespace() != owned.GetNamespace() && owner.GetNamespace() != "" {
		// cross-namespace owner references are not possible
		return -1, nil
	}
	gvk := owner.GetObjectKind().GroupVersionKind()
	if gvk.Version == "" || gvk.Kind == "" {
		if scheme == nil {
			return -1, fmt.Errorf("scheme must be provided if owner's GVK is not populated")
		}
		var err error
		gvk, err = apiutil.GVKForObject(owner, scheme)
		if err != nil {
			return -1, fmt.Errorf("unable to determine owner's GVK: %w", err)
		}
	}
	gv := gvk.GroupVersion().String()
	for idx, or := range owned.GetOwnerReferences() {
		if (owner.GetName() != "" && or.Name != owner.GetName()) ||
			(owner.GetUID() != "" && or.UID != owner.GetUID()) ||
			(or.APIVersion != gv) ||
			(or.Kind != gvk.Kind) {
			continue
		}
		return idx, nil
	}
	return -1, nil
}
