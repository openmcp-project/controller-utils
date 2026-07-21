package fips

import (
	"context"
	"crypto/fips140"
	"fmt"
	"os"

	"github.com/openmcp-project/controller-utils/pkg/logging"
)

// fipsEnforced is set to "true" at link time when built with FIPS support.
// Standard builds leave it empty — enforcement never triggers.
var fipsEnforced string

// Verify checks if the binary is built with FIPS compliance.
// If yes, it checks if FIPS mode is enabled.
// If it is not enabled but is expected to be enabled it exits the process.
func Verify(ctx context.Context) {
	log, ctx := logging.FromContextOrNew(ctx, nil)

	if fipsEnforced != "true" {
		return
	}

	if !fips140.Enabled() {
		err := fmt.Errorf("FIPS 140 is disabled but enforcement is enabled")
		log.Error(err, "FIPS 140 is disabled but enforcement is active")
		os.Exit(1)
	}
}
