package crds

import (
	"flag"
	"strconv"
)

type Flags struct {
	Install        bool
	InstallOptions []installOption
}

//nolint:lll
func BindFlags(fs *flag.FlagSet) *Flags {
	result := &Flags{}
	fs.BoolVar(&result.Install, "install-crds", false, "Install CRDs")

	fs.BoolFunc("crd-conversion-without-ca", "Do not include CA data in the CRD conversion webhook, e.g. when using a custom URL that has a valid certificate from a well-known CA.", func(s string) error {
		val, err := strconv.ParseBool(s)
		if val {
			result.InstallOptions = append(result.InstallOptions, WithoutCA)
		}
		return err
	})

	fs.Func("crd-conversion-base-url", "Base URL to the CRD conversion webhook service, e.g. when calling the webhook from another cluster", func(s string) error {
		result.InstallOptions = append(result.InstallOptions, WithCustomBaseURL(s))
		return nil
	})

	fs.Func("crd-conversion-service-port", "Port of the CRD conversion webhook service (not the webhook server itself)", func(s string) error {
		port, err := strconv.Atoi(s)
		result.InstallOptions = append(result.InstallOptions, WithWebhookServicePort(port))
		return err
	})

	return result
}
