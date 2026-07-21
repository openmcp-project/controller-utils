# FIPS

The `pkg/fips` package contains the `Verify` utility functions that checks if FIPS support is intended by this binary.
If yes, it checks if FIPS is enabled.
If it is not enabled, the process exits immediately.
