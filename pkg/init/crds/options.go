package crds

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//
// CRD Install Options
//

type installOptions struct {
	localClient        client.Client
	remoteClient       client.Client
	caData             []byte
	noResolveCA        bool
	customBaseUrl      *string
	webhookService     types.NamespacedName
	webhookSecret      types.NamespacedName
	webhookServicePort int32
}

type installOption interface {
	ApplyToInstallOptions(o *installOptions)
}

//
// Remote Client
//

type WithRemoteClient struct {
	Client client.Client
}

func (opt WithRemoteClient) ApplyToInstallOptions(o *installOptions) {
	o.remoteClient = opt.Client
}

//
// Custom Base URL
//

type WithCustomBaseURL string

func (opt WithCustomBaseURL) ApplyToInstallOptions(o *installOptions) {
	o.customBaseUrl = ptr.To(string(opt))
}

//
// Custom Certificate Authority
//

type WithCustomCA []byte

func (opt WithCustomCA) ApplyToInstallOptions(o *installOptions) {
	o.caData = []byte(opt)
	o.noResolveCA = true
}

//
// Don't resolve Certificate Authority
//

type withoutCA struct{}

var WithoutCA = withoutCA{}

func (withoutCA) ApplyToInstallOptions(o *installOptions) {
	o.caData = nil
	o.noResolveCA = true
}

//
// Webhook Service Port
//

type WithWebhookServicePort int32

func (opt WithWebhookServicePort) ApplyToInstallOptions(o *installOptions) {
	o.webhookServicePort = int32(opt)
}

//
// Webhook Secret Reference
//

type WithWebhookSecret types.NamespacedName

func (opt WithWebhookSecret) ApplyToInstallOptions(o *installOptions) {
	o.webhookSecret = types.NamespacedName(opt)
}

//
// Webhook Service Reference
//

type WithWebhookService types.NamespacedName

func (opt WithWebhookService) ApplyToInstallOptions(o *installOptions) {
	o.webhookService = types.NamespacedName(opt)
}
