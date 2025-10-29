package webhooks

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testValidArgs = []string{
		"-install-webhooks=true",
		"--webhooks-without-ca",
		"-webhooks-base-url=https://webhooks.example.com",
		"--webhooks-service-port", "1234",
		"-webhooks-additional-sans", "webhooks.example.com,webhooks.example.org",
		"--webhooks-bind-address=someaddr:4567",
	}
)

func Test_BindFlags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	flags := BindFlags(fs)
	err := fs.Parse(testValidArgs)

	assert.NoError(t, err)
	assert.NotNil(t, flags)
	assert.Equal(t, &Flags{
		Install: true,
		InstallOptions: []InstallOption{
			WithoutCA,
			WithCustomBaseURL("https://webhooks.example.com"),
			WithWebhookServicePort(1234),
		},
		CertOptions: []CertOption{
			WithAdditionalDNSNames{"webhooks.example.com", "webhooks.example.org"},
		},
		BindHost: "someaddr",
		BindPort: 4567,
	}, flags)
}
