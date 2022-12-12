package bazel

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.skia.org/infra/go/skerr"

	"go.skia.org/infra/go/exec"
	"go.skia.org/infra/task_driver/go/lib/os_steps"
	"go.skia.org/infra/task_driver/go/td"
)

// Bazel provides a Task Driver API for working with Bazel (via the Bazelisk launcher, see
// https://github.com/bazelbuild/bazelisk).
type Bazel struct {
	rbeCredentialFile string
	workspace         string
}

// NewWithRamdisk returns a new Bazel instance which uses a ramdisk as the Bazel cache.
//
// Using a ramdisk as the Bazel cache prevents CockroachDB "disk stall detected" errors on GCE VMs
// due to slow I/O.
func NewWithRamdisk(ctx context.Context, workspace string, rbeCredentialFile string, sizeGb int) (*Bazel, func(), error) {
	// Create and mount ramdisk.
	//
	// At the time of writing, a full build of the Buildbot repository on an empty Bazel cache takes
	// ~20GB on cache space. Infra-PerCommit-Test-Bazel-Local runs on GCE VMs with 64GB of RAM.
	ramdiskDir, err := os_steps.TempDir(ctx, "", "ramdisk_*")
	if err != nil {
		return nil, nil, skerr.Wrap(err)
	}
	if _, err := exec.RunCwd(ctx, workspace, "sudo", "mount", "-t", "tmpfs", "-o", fmt.Sprintf("size=%dg", sizeGb), "tmpfs", ramdiskDir); err != nil {
		return nil, nil, skerr.Wrap(err)
	}

	// Create Bazel cache directory inside the ramdisk.
	//
	// Using the ramdisk's mount point directly as the Bazel cache causes Bazel to fail with a file
	// permission error. Using a directory within the ramdisk as the Bazel cache prevents this error.
	cachePath := filepath.Join(ramdiskDir, "bazel_cache")
	if err := os_steps.MkdirAll(ctx, cachePath); err != nil {
		return nil, nil, skerr.Wrap(err)
	}
	repositoryCachePath := filepath.Join(ramdiskDir, "bazel_repo_cache")
	if err := os_steps.MkdirAll(ctx, repositoryCachePath); err != nil {
		return nil, nil, skerr.Wrap(err)
	}
	opts := BazelOptions{
		CachePath:           cachePath,
		RepositoryCachePath: repositoryCachePath,
	}
	if err := EnsureBazelRCFile(ctx, opts); err != nil {
		return nil, nil, skerr.Wrap(err)
	}

	absCredentialFile, err := os_steps.Abs(ctx, rbeCredentialFile)
	if err != nil {
		return nil, nil, skerr.Wrap(err)
	}
	bzl := &Bazel{
		rbeCredentialFile: absCredentialFile,
		workspace:         workspace,
	}

	cleanup := func() {
		// Shut down the Bazel server. This ensures that there are no processes with open files under
		// the ramdisk, which would otherwise cause a "target is busy" when we unmount the ramdisk.
		if _, err := bzl.Do(ctx, "shutdown"); err != nil {
			td.Fatal(ctx, err)
		}

		if _, err := exec.RunCwd(ctx, workspace, "sudo", "umount", ramdiskDir); err != nil {
			td.Fatal(ctx, err)
		}
		if err := os_steps.RemoveAll(ctx, ramdiskDir); err != nil {
			td.Fatal(ctx, err)
		}
	}
	return bzl, cleanup, nil
}

type BazelOptions struct {
	CachePath           string
	RepositoryCachePath string
}

// EnsureBazelRCFile makes sure the user .bazelrc file exists and matches the provided
// configuration. This makes it easy for all subsequent calls to Bazel use the right command
// line args, even if Bazel is not invoked directly from task_driver (e.g. from a Makefile).
func EnsureBazelRCFile(ctx context.Context, bazelOpts BazelOptions) error {
	var bazelRCLines []string
	if bazelOpts.CachePath != "" {
		// https://docs.bazel.build/versions/main/output_directories.html#current-layout
		bazelRCLines = append(bazelRCLines, fmt.Sprintf("startup --output_user_root=%s", bazelOpts.CachePath))
	}
	if bazelOpts.RepositoryCachePath != "" {
		// Also keep the repository cache on disk so that it survives reboots.
		// https://bazel.build/docs/build#repository-cache
		bazelRCLines = append(bazelRCLines, fmt.Sprintf("build --repository_cache=%s", bazelOpts.RepositoryCachePath))
	}

	// https://docs.bazel.build/versions/main/guide.html#where-are-the-bazelrc-files
	// We go for the user's .bazelrc file instead of the system one because the
	// swarming user does not have access to write to /etc/bazel.bazelrc
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return skerr.Wrap(err)
	}
	userBazelRCLocation := filepath.Join(homeDir, ".bazelrc")
	bazelRCContent := strings.Join(bazelRCLines, "\n")
	return os_steps.WriteFile(ctx, userBazelRCLocation, []byte(bazelRCContent), 0666)
}

// New returns a new Bazel instance.
func New(ctx context.Context, workspace, rbeCredentialFile string, opts BazelOptions) (*Bazel, error) {
	if err := EnsureBazelRCFile(ctx, opts); err != nil {
		return nil, skerr.Wrap(err)
	}

	absCredentialFile := ""
	var err error
	if rbeCredentialFile != "" {
		absCredentialFile, err = os_steps.Abs(ctx, rbeCredentialFile)
		if err != nil {
			return nil, skerr.Wrap(err)
		}
	}
	return &Bazel{
		rbeCredentialFile: absCredentialFile,
		workspace:         workspace,
	}, nil
}

// Do executes a Bazel subcommand.
func (b *Bazel) Do(ctx context.Context, subCmd string, args ...string) (string, error) {
	cmd := []string{"bazelisk", subCmd}
	cmd = append(cmd, args...)
	return exec.RunCwd(ctx, b.workspace, cmd...)
}

// DoOnRBE executes a Bazel subcommand on RBE.
func (b *Bazel) DoOnRBE(ctx context.Context, subCmd string, args ...string) (string, error) {
	// See https://bazel.build/reference/command-line-reference
	cmd := []string{
		"--config=remote",
		"--sandbox_base=/dev/shm", // Make builds faster by using a RAM disk for the sandbox.
	}
	if b.rbeCredentialFile != "" {
		cmd = append(cmd, "--google_credentials="+b.rbeCredentialFile)
	} else {
		cmd = append(cmd, "--google_default_credentials")
	}
	cmd = append(cmd, args...)
	return b.Do(ctx, subCmd, cmd...)

}
