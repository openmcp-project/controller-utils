package controller

import "sigs.k8s.io/controller-runtime/pkg/client"

////////////////////
// STATUS UPDATER //
////////////////////

// NewOpenMCPStatusUpdaterBuilder returns a StatusUpdaterBuilder that expects only ObservedGeneration, Conditions, and Phase as status fields.
// It does not include LastReconcileTime, Reason, or Message.
func NewOpenMCPStatusUpdaterBuilder[Obj client.Object]() *StatusUpdaterBuilder[Obj] {
	return NewStatusUpdaterBuilder[Obj]().WithoutFields(
		STATUS_FIELD_LAST_RECONCILE_TIME,
		STATUS_FIELD_REASON,
		STATUS_FIELD_MESSAGE,
	)
}
