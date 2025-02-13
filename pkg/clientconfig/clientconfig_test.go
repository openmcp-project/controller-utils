package clientconfig

import (
	"os"
	"testing"

	"github.com/openmcp-project/controller-utils/api"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type test_input struct {
	config         api.Target
	kubeconfigFile string
}

type test_want struct {
	err           error
	host          string
	impersonation rest.ImpersonationConfig
	token         string
	tokenFile     string
	caFile        string
	caData        string
}

var (
	noerror = test_want{
		err:   nil,
		host:  "https://api.example.com",
		token: "G1FUzrd3FCgLVhIy6kj7",
		caData: `-----BEGIN CERTIFICATE-----
MIID5jCCAk6gAwIBAgIQc3k9EZlBPzaqkHrAFNDOyjANBgkqhkiG9w0BAQsFADAN
MQswCQYDVQQDEwJjYTAeFw0yMzExMTMxMjE0MjBaFw0zMzExMTMxMjE0MjBaMA0x
CzAJBgNVBAMTAmNhMIIBojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKCAYEAv/FZ
jnFA9LDXPUcFQJgeIHEDJ5pfK6hw1Q133AfIyRFYpbTiKbh10z4g7JK2oLzL5CKg
W8yj8d909Ysbf40N4B+yl5nn1M143Tg6Z3QMmaGg7CLbXHSgbIXwD9veRvW9AN81
g444UVSijLkp7SSZWuaMdn6tP9DPpFHOqE6SpnGfHd/iQ3XkFuLyPdrCRlcUuRno
6rf+ANli5deamqjXS1KEsilmhSCYRPQ61nOsbTteWjivPaBxlxYMhNk1l017kV/C
z8/1cpauLJCmgPuNNqMI8CvduWpFtgW7420DPC2vFH4JCqqHIhH+hF7B+jzMEvcz
14oOXMOq9+AsVlK1Epqg0yvGb7wscbrJL5IJ6vaUNB3v43sZAHxGzsoCCf4wI6dB
l1xmwm+kctE6bxEA9ynAJLeog2bYZuKwZm0QBqSUXumPb6ctMRvX2JA4jRrRqTnV
nj8u0cChLe1Ij0IFMkJDGf+VHQB+9Q52BvKVg1gqWjlJ3o+n17fhknDhHnz7AgMB
AAGjQjBAMA4GA1UdDwEB/wQEAwIBpjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQW
BBR0hDfRF7z2nr8nIUx+y8UkuKJcYzANBgkqhkiG9w0BAQsFAAOCAYEAuNH9T+Aq
0mzNrwFhwc/nUqC+0F921VVryIbR6I2amt56GFXO0QPy837WISFqkKKC7bM02uRN
4ORNHYhwedSrR6NkQihYHpq52CruKjKn296lyCxlyEWzH8poYW+kjfuzugwJ+Ih9
RIgGnKZiNWwzc3PLOW4zUzfyWVQVUkGZuN4qTqwoBn2dJwnIBqep3gkdPZbZZpGI
UOpVlZu0zDtJH+F1QzUftJdWeqbMl/YTbOfBKasDepqUbrZioDWnuXHzhF7iqMnN
6k/jHbJ3kTRgH1d262iGgbGOjO3ZRLt1sijxucKfIMjM2H4yW9zmUWuYGdZsJTu2
oQYRIgCpagRDwiQI7gBPLwdIgWbiFMUUbaaNzeQSBlxGzwpwKTB/kGSjCOJN/p8c
Jj6XYmmcvVurcBUlce+YThzpBND4YrYCbfyjH+WAZkDP458JONbLojjjLNRDtN4f
l0nUqSl1FdVvsCo8hUS8XZuciDluhp4Lq6fXx5002SWC1jP4rOvZvs7F
-----END CERTIFICATE-----
`,
	}
)

func Test_GetRESTConfig(t *testing.T) {
	testCases := []struct {
		desc  string
		input test_input
		want  test_want
	}{
		{
			desc: "should read inline kubeconfig",
			input: test_input{
				config: api.Target{
					Kubeconfig: &v1.JSON{
						Raw: []byte(`apiVersion: v1
kind: Config
current-context: shoot--unit-test
contexts:
  - name: shoot--unit-test
    context:
      cluster: shoot--unit-test
      user: shoot--unit-test
clusters:
  - name: shoot--unit-test
    cluster:
      server: https://api.example.com
      certificate-authority-data: >-
        LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUQ1akNDQWs2Z0F3SUJBZ0lRYzNrOUVabEJQemFxa0hyQUZORE95akFOQmdrcWhraUc5dzBCQVFzRkFEQU4KTVFzd0NRWURWUVFERXdKallUQWVGdzB5TXpFeE1UTXhNakUwTWpCYUZ3MHpNekV4TVRNeE1qRTBNakJhTUEweApDekFKQmdOVkJBTVRBbU5oTUlJQm9qQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FZOEFNSUlCaWdLQ0FZRUF2L0ZaCmpuRkE5TERYUFVjRlFKZ2VJSEVESjVwZks2aHcxUTEzM0FmSXlSRllwYlRpS2JoMTB6NGc3Sksyb0x6TDVDS2cKVzh5ajhkOTA5WXNiZjQwTjRCK3lsNW5uMU0xNDNUZzZaM1FNbWFHZzdDTGJYSFNnYklYd0Q5dmVSdlc5QU44MQpnNDQ0VVZTaWpMa3A3U1NaV3VhTWRuNnRQOURQcEZIT3FFNlNwbkdmSGQvaVEzWGtGdUx5UGRyQ1JsY1V1Um5vCjZyZitBTmxpNWRlYW1xalhTMUtFc2lsbWhTQ1lSUFE2MW5Pc2JUdGVXaml2UGFCeGx4WU1oTmsxbDAxN2tWL0MKejgvMWNwYXVMSkNtZ1B1Tk5xTUk4Q3ZkdVdwRnRnVzc0MjBEUEMydkZINEpDcXFISWhIK2hGN0IranpNRXZjegoxNG9PWE1PcTkrQXNWbEsxRXBxZzB5dkdiN3dzY2JySkw1SUo2dmFVTkIzdjQzc1pBSHhHenNvQ0NmNHdJNmRCCmwxeG13bStrY3RFNmJ4RUE5eW5BSkxlb2cyYlladUt3Wm0wUUJxU1VYdW1QYjZjdE1SdlgySkE0alJyUnFUblYKbmo4dTBjQ2hMZTFJajBJRk1rSkRHZitWSFFCKzlRNTJCdktWZzFncVdqbEozbytuMTdmaGtuRGhIbno3QWdNQgpBQUdqUWpCQU1BNEdBMVVkRHdFQi93UUVBd0lCcGpBUEJnTlZIUk1CQWY4RUJUQURBUUgvTUIwR0ExVWREZ1FXCkJCUjBoRGZSRjd6Mm5yOG5JVXgreThVa3VLSmNZekFOQmdrcWhraUc5dzBCQVFzRkFBT0NBWUVBdU5IOVQrQXEKMG16TnJ3Rmh3Yy9uVXFDKzBGOTIxVlZyeUliUjZJMmFtdDU2R0ZYTzBRUHk4MzdXSVNGcWtLS0M3Yk0wMnVSTgo0T1JOSFlod2VkU3JSNk5rUWloWUhwcTUyQ3J1S2pLbjI5Nmx5Q3hseUVXekg4cG9ZVytramZ1enVnd0orSWg5ClJJZ0duS1ppTld3emMzUExPVzR6VXpmeVdWUVZVa0dadU40cVRxd29CbjJkSnduSUJxZXAzZ2tkUFpiWlpwR0kKVU9wVmxadTB6RHRKSCtGMVF6VWZ0SmRXZXFiTWwvWVRiT2ZCS2FzRGVwcVViclppb0RXbnVYSHpoRjdpcU1uTgo2ay9qSGJKM2tUUmdIMWQyNjJpR2diR09qTzNaUkx0MXNpanh1Y0tmSU1qTTJINHlXOXptVVd1WUdkWnNKVHUyCm9RWVJJZ0NwYWdSRHdpUUk3Z0JQTHdkSWdXYmlGTVVVYmFhTnplUVNCbHhHendwd0tUQi9rR1NqQ09KTi9wOGMKSmo2WFltbWN2VnVyY0JVbGNlK1lUaHpwQk5ENFlyWUNiZnlqSCtXQVprRFA0NThKT05iTG9qampMTlJEdE40ZgpsMG5VcVNsMUZkVnZzQ284aFVTOFhadWNpRGx1aHA0THE2Zlh4NTAwMlNXQzFqUDRyT3ZadnM3RgotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
users:
  - name: shoot--unit-test
    user:
      token: >-
        G1FUzrd3FCgLVhIy6kj7
`),
					},
				},
			},
			want: noerror,
		},
		{
			desc: "should fail because multiple methods are configured",
			input: test_input{
				config: api.Target{
					Kubeconfig:     &v1.JSON{Raw: []byte("hello world")},
					KubeconfigRef:  &api.KubeconfigReference{},
					ServiceAccount: &api.ServiceAccountConfig{},
				},
			},
			want: test_want{
				err: ErrInvalidConnectionMethod,
			},
		},
		{
			desc: "should fail because no methods are configured",
			input: test_input{
				config: api.Target{},
			},
			want: test_want{
				err: ErrInvalidConnectionMethod,
			},
		},
		{
			desc: "should read kubeconfig from file",
			input: test_input{
				kubeconfigFile: "testdata/valid.yaml",
				config: api.Target{
					ServiceAccount: &api.ServiceAccountConfig{},
				},
			},
			want: noerror,
		},
		{
			desc: "should read kubeconfig from file (explicitly)",
			input: test_input{
				config: api.Target{
					KubeconfigFile: ptr.To("testdata/valid.yaml"),
				},
			},
			want: noerror,
		},
		{
			desc: "should read kubeconfig from file and set custom values",
			input: test_input{
				kubeconfigFile: "testdata/valid.yaml",
				config: api.Target{
					ServiceAccount: &api.ServiceAccountConfig{
						Name:      "myuser",
						Namespace: "mynamespace",
						Host:      "https://custom-api.example.com",
						CAFile:    ptr.To("/etc/custom/ca.crt"),
						CAData:    ptr.To("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUQ1akNDQWs2Z0F3SUJBZ0lRYzNrOUVabEJQemFxa0hyQUZORE95akFOQmdrcWhraUc5dzBCQVFzRkFEQU4KTVFzd0NRWURWUVFERXdKallUQWVGdzB5TXpFeE1UTXhNakUwTWpCYUZ3MHpNekV4TVRNeE1qRTBNakJhTUEweApDekFKQmdOVkJBTVRBbU5oTUlJQm9qQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FZOEFNSUlCaWdLQ0FZRUF2L0ZaCmpuRkE5TERYUFVjRlFKZ2VJSEVESjVwZks2aHcxUTEzM0FmSXlSRllwYlRpS2JoMTB6NGc3Sksyb0x6TDVDS2cKVzh5ajhkOTA5WXNiZjQwTjRCK3lsNW5uMU0xNDNUZzZaM1FNbWFHZzdDTGJYSFNnYklYd0Q5dmVSdlc5QU44MQpnNDQ0VVZTaWpMa3A3U1NaV3VhTWRuNnRQOURQcEZIT3FFNlNwbkdmSGQvaVEzWGtGdUx5UGRyQ1JsY1V1Um5vCjZyZitBTmxpNWRlYW1xalhTMUtFc2lsbWhTQ1lSUFE2MW5Pc2JUdGVXaml2UGFCeGx4WU1oTmsxbDAxN2tWL0MKejgvMWNwYXVMSkNtZ1B1Tk5xTUk4Q3ZkdVdwRnRnVzc0MjBEUEMydkZINEpDcXFISWhIK2hGN0IranpNRXZjegoxNG9PWE1PcTkrQXNWbEsxRXBxZzB5dkdiN3dzY2JySkw1SUo2dmFVTkIzdjQzc1pBSHhHenNvQ0NmNHdJNmRCCmwxeG13bStrY3RFNmJ4RUE5eW5BSkxlb2cyYlladUt3Wm0wUUJxU1VYdW1QYjZjdE1SdlgySkE0alJyUnFUblYKbmo4dTBjQ2hMZTFJajBJRk1rSkRHZitWSFFCKzlRNTJCdktWZzFncVdqbEozbytuMTdmaGtuRGhIbno3QWdNQgpBQUdqUWpCQU1BNEdBMVVkRHdFQi93UUVBd0lCcGpBUEJnTlZIUk1CQWY4RUJUQURBUUgvTUIwR0ExVWREZ1FXCkJCUjBoRGZSRjd6Mm5yOG5JVXgreThVa3VLSmNZekFOQmdrcWhraUc5dzBCQVFzRkFBT0NBWUVBdU5IOVQrQXEKMG16TnJ3Rmh3Yy9uVXFDKzBGOTIxVlZyeUliUjZJMmFtdDU2R0ZYTzBRUHk4MzdXSVNGcWtLS0M3Yk0wMnVSTgo0T1JOSFlod2VkU3JSNk5rUWloWUhwcTUyQ3J1S2pLbjI5Nmx5Q3hseUVXekg4cG9ZVytramZ1enVnd0orSWg5ClJJZ0duS1ppTld3emMzUExPVzR6VXpmeVdWUVZVa0dadU40cVRxd29CbjJkSnduSUJxZXAzZ2tkUFpiWlpwR0kKVU9wVmxadTB6RHRKSCtGMVF6VWZ0SmRXZXFiTWwvWVRiT2ZCS2FzRGVwcVViclppb0RXbnVYSHpoRjdpcU1uTgo2ay9qSGJKM2tUUmdIMWQyNjJpR2diR09qTzNaUkx0MXNpanh1Y0tmSU1qTTJINHlXOXptVVd1WUdkWnNKVHUyCm9RWVJJZ0NwYWdSRHdpUUk3Z0JQTHdkSWdXYmlGTVVVYmFhTnplUVNCbHhHendwd0tUQi9rR1NqQ09KTi9wOGMKSmo2WFltbWN2VnVyY0JVbGNlK1lUaHpwQk5ENFlyWUNiZnlqSCtXQVprRFA0NThKT05iTG9qampMTlJEdE40ZgpsMG5VcVNsMUZkVnZzQ284aFVTOFhadWNpRGx1aHA0THE2Zlh4NTAwMlNXQzFqUDRyT3ZadnM3RgotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="),
						TokenFile: "testdata/token",
					},
				},
			},
			want: test_want{
				err:  nil,
				host: "https://custom-api.example.com",
				impersonation: rest.ImpersonationConfig{
					UserName: "system:serviceaccount:myuser:mynamespace",
				},
				token:     "",
				tokenFile: "testdata/token",
				caFile:    "/etc/custom/ca.crt",
				caData:    noerror.caData,
			},
		},
		{
			desc: "should read kubeconfig from file and set custom pem-encoded CA data",
			input: test_input{
				kubeconfigFile: "testdata/valid.yaml",
				config: api.Target{
					ServiceAccount: &api.ServiceAccountConfig{
						CAData: ptr.To(noerror.caData),
					},
				},
			},
			want: noerror,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			if tC.input.kubeconfigFile != "" {
				os.Setenv("KUBECONFIG", tC.input.kubeconfigFile)
			} else {
				os.Unsetenv("KUBECONFIG")
			}

			wrapped := New(tC.input.config)
			conf, reloadFunc, err := wrapped.GetRESTConfig()
			client, reloadFunc2, clienterr := wrapped.GetClient(client.Options{})

			if err != nil {
				assert.ErrorIs(t, err, tC.want.err)
				// GetClient should return the same error
				assert.ErrorIs(t, clienterr, tC.want.err)
				assert.Nil(t, client)
				return
			}

			assert.NoError(t, err)
			// GetClient should not return any error
			assert.NoError(t, clienterr)
			assert.NotNil(t, client)
			assert.Equal(t, tC.want.host, conf.Host)
			assert.Equal(t, tC.want.impersonation, conf.Impersonate)
			assert.Equal(t, tC.want.token, conf.BearerToken)
			assert.Equal(t, tC.want.tokenFile, conf.BearerTokenFile)
			assert.Equal(t, tC.want.caFile, conf.CAFile)
			assert.Equal(t, tC.want.caData, string(conf.CAData))

			if assert.NotNil(t, reloadFunc) {
				assert.NoError(t, reloadFunc())
			}
			if assert.NotNil(t, reloadFunc2) {
				assert.NoError(t, reloadFunc2())
			}
		})
	}
}

func Test_KubeconfigFile_Reload(t *testing.T) {
	wrapped := New(api.Target{
		KubeconfigFile: ptr.To("testdata/valid.yaml"),
	})

	conf, reloadFunc, err := wrapped.GetRESTConfig()
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.NotNil(t, reloadFunc)
	assert.Equal(t, "https://api.example.com", conf.Host)
	assert.Equal(t, "G1FUzrd3FCgLVhIy6kj7", conf.BearerToken)

	wrapped.KubeconfigFile = ptr.To("testdata/valid2.yaml")
	assert.NoError(t, reloadFunc())
	assert.Equal(t, "https://api.example.org", conf.Host)
	assert.Equal(t, "vp98rIsJJZ3qcoHAsUhg", conf.BearerToken)
}
