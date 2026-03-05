package cli

import "os"

// IsTTY reports whether stdout is a terminal. It is a package-level variable for testability.
var IsTTY = func() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
