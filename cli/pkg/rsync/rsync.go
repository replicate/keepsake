package rsync

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adjust/uniuri"

	"path"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/remote"
)

// TODO(andreas): since we're not actually uploading to a consistent
// directory, we're not taking advantage of rsync's ability to upload
// changes. perhaps it'd be better to just use sftp.

const rsyncX86BinaryURL = "https://github.com/JBBgameich/rsync-static/releases/download/continuous/rsync-x86"
const remoteRsyncTempPath = "/opt/replicate/rsync"

func UploadToTempDir(localDir string, remoteOptions *remote.Options) (remoteDir string, err error) {
	remoteDir = path.Join("/tmp/replicate/upload", uniuri.NewLen(20))
	client, err := remote.NewClient(remoteOptions)
	if err != nil {
		return "", err
	}

	if err := client.SFTP().MkdirAll(remoteDir); err != nil {
		return "", fmt.Errorf("Failed to create remote temp directory %s, got error: %s", remoteDir, err)
	}

	if err := Upload(localDir, remoteOptions, remoteDir); err != nil {
		return "", err
	}
	return remoteDir, nil
}

func Upload(localDir string, remoteOptions *remote.Options, remoteDir string) error {
	if err := ensureLocalRsync(); err != nil {
		return err
	}

	remoteRsyncPath, err := ensureRemoteRsync(remoteOptions)
	if err != nil {
		return err
	}

	console.Debug("Uploading %s to %s", localDir, getRemoteSpec(remoteOptions, remoteDir))

	cmd := rsyncCmd(localDir, remoteOptions, remoteDir, remoteRsyncPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Failed to upload %s to %s, got error: %s", localDir, getRemoteSpec(remoteOptions, remoteDir), output)
	}

	return nil
}

func rsyncCmd(localDir string, remoteOptions *remote.Options, remoteDir string, remoteRsyncPath string) *exec.Cmd {
	// rsync wants / suffixes
	if !strings.HasSuffix(localDir, "/") {
		localDir = localDir + "/"
	}
	if !strings.HasSuffix(remoteDir, "/") {
		remoteDir = remoteDir + "/"
	}

	// basic rsync options
	args := []string{"--update", "--archive"}

	// ssh options
	sshCmd := fmt.Sprintf("ssh %s", strings.Join(remoteOptions.SSHArgs(), " "))
	args = append(args, "--rsh", sshCmd)

	// TODO: .replicateignore (just maybe, there might be value in having the remote directory be an exact mirror, with .git and everything?)

	// hack to get rid of core dumps
	args = append(args, "--exclude", "core", "--include", "core/")

	// if we installed rsync
	if remoteRsyncPath != "" {
		args = append(args, "--rsync-path", remoteRsyncPath)
	}

	args = append(args, localDir, getRemoteSpec(remoteOptions, remoteDir))
	return exec.Command("rsync", args...)
}

func getRemoteSpec(remoteOptions *remote.Options, remoteDir string) string {
	remoteSpec := ""
	// FIXME (bfirsh): this could be done with -l in SSHArgs so we don't have to concatenate strings, like how Docker client does it
	if remoteOptions.Username != "" {
		remoteSpec += remoteOptions.Username + "@"
	}
	remoteSpec += remoteOptions.Host + ":" + remoteDir
	return remoteSpec
}

func ensureRemoteRsync(remoteOptions *remote.Options) (remoteRsyncPath string, err error) {
	client, err := remote.NewClient(remoteOptions)
	if err != nil {
		return "", err
	}

	if client.Command("which", "rsync").Run() == nil {
		return "rsync", nil
	}

	if _, err := client.SFTP().Stat(remoteRsyncTempPath); err == nil {
		return remoteRsyncTempPath, nil
	}

	// TODO: check that remote system is linux

	console.Debug("Installing remote rsync binary at %s", remoteRsyncTempPath)

	resp, err := http.Get(rsyncX86BinaryURL)
	if err != nil {
		return "", fmt.Errorf("Failed to download rsync binary")
	}
	defer resp.Body.Close()

	remoteRsyncDir := filepath.Dir(remoteRsyncTempPath)
	if _, err := client.SFTP().Stat(remoteRsyncDir); err != nil {
		if err := client.SFTP().MkdirAll(remoteRsyncDir); err != nil {
			return "", fmt.Errorf("Failed to create remote directory %s, got error: %s", remoteRsyncDir, err)
		}
	}

	remoteFile, err := client.SFTP().Create(remoteRsyncTempPath)
	if err != nil {
		return "", fmt.Errorf("Failed to create remote rsync file at %s, got error: %s", remoteRsyncTempPath, err)
	}

	if _, err := io.Copy(remoteFile, resp.Body); err != nil {
		return "", fmt.Errorf("Failed to write remote rsync file at %s, got error: %s", remoteRsyncTempPath, err)
	}
	if err := remoteFile.Close(); err != nil {
		return "", fmt.Errorf("Failed to close remote rsync file at %s, got error: %s", remoteRsyncTempPath, err)
	}

	if err := client.SFTP().Chmod(remoteRsyncTempPath, 0755); err != nil {
		return "", fmt.Errorf("Failed to make remote rsync file at %s executable, got error: %s", remoteRsyncTempPath, err)
	}

	return remoteRsyncTempPath, nil
}

func ensureLocalRsync() error {
	output, err := exec.Command("which", "rsync").CombinedOutput()
	if err != nil {
		console.Debug("which rsync: %s", output)
		return fmt.Errorf("The rsync command was not found. Please install rsync and try again")
	}
	return nil
}
