// Package npm provides functions for running npm commands in a generated
// Astro Starlight documentation site. All npm/node interactions are
// concentrated in this file so they are easy to find, review, and modify.
//
// Each function runs a single npm command via os/exec, streaming stdout
// and stderr to the caller-supplied writers (typically os.Stdout/os.Stderr).
package npm

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// RunOptions controls how npm commands are executed.
type RunOptions struct {
	// Dir is the working directory (the generated site root).
	Dir string
	// Stdout and Stderr receive the command's output.
	// If nil, os.Stdout / os.Stderr are used.
	Stdout io.Writer
	Stderr io.Writer
}

func (o *RunOptions) stdout() io.Writer {
	if o.Stdout != nil {
		return o.Stdout
	}
	return os.Stdout
}

func (o *RunOptions) stderr() io.Writer {
	if o.Stderr != nil {
		return o.Stderr
	}
	return os.Stderr
}

// Install runs "npm install" in the site directory.
// This installs the dependencies declared in package.json (Astro, Starlight, etc.).
func Install(opts RunOptions) error {
	return run(opts, "install")
}

// Build runs "npm run build" in the site directory.
// This invokes the Astro SSG build, producing static HTML in dist/.
func Build(opts RunOptions) error {
	return run(opts, "run", "build")
}

// Dev runs "npm run dev" in the site directory.
// This starts the Astro dev server for local preview with hot-reload.
func Dev(opts RunOptions) error {
	return run(opts, "run", "dev")
}

// Preview runs "npm run preview" in the site directory.
// This serves the built static site locally for final review.
func Preview(opts RunOptions) error {
	return run(opts, "run", "preview")
}

// run executes an npm command with the given arguments.
func run(opts RunOptions, args ...string) error {
	npmPath, err := exec.LookPath("npm")
	if err != nil {
		return fmt.Errorf("npm not found in PATH: %w (install Node.js from https://nodejs.org)", err)
	}

	cmd := exec.Command(npmPath, args...)
	cmd.Dir = opts.Dir
	cmd.Stdout = opts.stdout()
	cmd.Stderr = opts.stderr()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm %s failed: %w", args[0], err)
	}
	return nil
}
