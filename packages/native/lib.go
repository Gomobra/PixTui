package native

import (
	"io"
)

// lib is the entry point module for the native package. It mirrors the Zig
// lib.zig which re-exports the process-wide I/O and standard options, and
// registers ABI export symbols.
//
// In Go this provides package-level access to the standard I/O sink and
// default options used by native resources.

// IO is the process-wide I/O interface used by standalone native resources.
// It mirrors runtime.zig's io() function.
var IO io.Writer = io.Discard

// SetIO sets the process-wide I/O writer.
func SetIO(w io.Writer) {
	IO = w
}

// StdOptions holds standard options for native resources.
// Mirrors runtime-abi.zig's std_options.
type StdOptions struct {
	// LogLevel controls logging verbosity (0=error, 1=warn, 2=info, 3=debug).
	LogLevel int
}

// DefaultStdOptions returns the default standard options.
func DefaultStdOptions() StdOptions {
	return StdOptions{
		LogLevel: 2, // info
	}
}

// DefaultStdOptionsVar is the package-wide default standard options value.
var DefaultStdOptionsVar = DefaultStdOptions()
