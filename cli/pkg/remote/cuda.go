package remote

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
)

// GetCUDADriverVersion returns the CUDA driver version or empty
// string if the host doesn't have a CUDA driver
func (c *Client) GetCUDADriverVersion() (string, error) {
	versionPath := "/proc/driver/nvidia/version"
	file, err := c.SFTP().Open(versionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("Failed to open %s: %s", versionPath, err)
	}

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("Failed to read %s: %s", versionPath, err)
	}

	re := regexp.MustCompile(`Kernel Module\s+([0-9.]+)\s+`)
	matches := re.FindStringSubmatch(string(contents))
	if len(matches) == 0 {
		return "", fmt.Errorf("Failed to parse %s, no Kernel Module version in file contents:\n%s", versionPath, string(contents))
	}

	return matches[1], nil
}
