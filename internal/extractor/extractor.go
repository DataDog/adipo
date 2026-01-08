package extractor

import (
	"fmt"
	"os"
)

// Extractor is the interface for binary extraction strategies
type Extractor interface {
	// Extract extracts binary data and returns the path to execute
	Extract(data []byte, name string) (path string, cleanup func(), err error)

	// Name returns the extractor name
	Name() string
}

// ExecutionOptions contains options for executing the binary
type ExecutionOptions struct {
	Args            []string          // Command-line arguments
	Env             []string          // Environment variables
	PreferMemory    bool              // Prefer memory extraction over disk
	TempDir         string            // Custom temp directory (for disk extraction)
	Verbose         bool              // Verbose output
}

// DefaultExecutionOptions returns default execution options
func DefaultExecutionOptions() *ExecutionOptions {
	return &ExecutionOptions{
		Args:         os.Args[1:],
		Env:          os.Environ(),
		PreferMemory: true,
		TempDir:      "",
		Verbose:      false,
	}
}

// ExtractAndExecute extracts a binary and executes it with the given options
func ExtractAndExecute(data []byte, name string, opts *ExecutionOptions) error {
	if opts == nil {
		opts = DefaultExecutionOptions()
	}

	var extractor Extractor
	var fallback Extractor

	// Choose extraction strategy
	if opts.PreferMemory {
		extractor = &MemoryExtractor{}
		fallback = &DiskExtractor{TempDir: opts.TempDir}
	} else {
		extractor = &DiskExtractor{TempDir: opts.TempDir}
	}

	// Try primary extraction method
	path, cleanup, err := extractor.Extract(data, name)
	if err != nil {
		if fallback == nil {
			return fmt.Errorf("%s extraction failed: %w", extractor.Name(), err)
		}

		// Try fallback
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Warning: %s extraction failed, falling back to %s\n",
				extractor.Name(), fallback.Name())
		}

		path, cleanup, err = fallback.Extract(data, name)
		if err != nil {
			return fmt.Errorf("all extraction methods failed: %w", err)
		}
	}

	// Ensure cleanup happens
	if cleanup != nil {
		defer cleanup()
	}

	// Execute the binary
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Executing: %s\n", path)
	}

	return Execute(path, opts.Args, opts.Env)
}
