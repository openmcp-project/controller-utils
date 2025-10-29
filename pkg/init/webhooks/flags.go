package webhooks

import (
	"flag"
	"strconv"
	"strings"
)

const (
	defaultPort = 9443
)

type Flags struct {
	Install        bool
	InstallOptions []InstallOption
	CertOptions    []CertOption
	BindHost       string
	BindPort       int
}

//nolint:lll
func BindFlags(fs *flag.FlagSet) *Flags {
	result := &Flags{
		BindPort: defaultPort,
	}
	fs.BoolVar(&result.Install, "install-webhooks", false, "Install webhooks")

	fs.BoolFunc("webhooks-without-ca", "Do not include CA data in the webhooks, e.g. when using a custom URL that has a valid certificate from a well-known CA.", func(s string) error {
		val, err := strconv.ParseBool(s)
		if val {
			result.InstallOptions = append(result.InstallOptions, WithoutCA)
		}
		return err
	})

	fs.Func("webhooks-base-url", "Base URL to the webhooks service, e.g. when calling the webhook from another cluster", func(s string) error {
		result.InstallOptions = append(result.InstallOptions, WithCustomBaseURL(s))
		return nil
	})

	fs.Func("webhooks-service-port", "Port of the webhooks service (not the webhook server itself)", func(s string) error {
		port, err := strconv.Atoi(s)
		result.InstallOptions = append(result.InstallOptions, WithWebhookServicePort(port))
		return err
	})

	fs.Func("webhooks-additional-sans", "Additional Subject Alternative Names (SANs) that should be added to the self-signed webhook certificate. Multiple domains can be specified as comma-separated values.", func(s string) error {
		result.CertOptions = append(result.CertOptions, WithAdditionalDNSNames(strings.Split(s, ",")))
		return nil
	})

	fs.Func("webhooks-bind-address", "The address the webhook endpoint binds to.", func(s string) error {
		host, portStr, found := strings.Cut(s, ":")
		result.BindHost = host
		if found {
			port, err := strconv.Atoi(portStr)
			if err != nil {
				return err
			}
			result.BindPort = port
		}
		return nil
	})

	return result
}
