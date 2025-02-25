package crds

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testValidArgs = []string{
		"-install-crds=true",
		"--crd-conversion-without-ca",
		"-crd-conversion-base-url=https://webhooks.example.com",
		"--crd-conversion-service-port", "1234",
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
		InstallOptions: []installOption{
			WithoutCA,
			WithCustomBaseURL("https://webhooks.example.com"),
			WithWebhookServicePort(1234),
		},
	}, flags)
}
