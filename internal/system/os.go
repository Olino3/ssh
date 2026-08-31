package system

import (
	"fmt"
	"os"
	"strings"
)

type OSInfo struct {
	ID      string
	Version string
	Pretty  string
}

func ParseOSRelease(data []byte) OSInfo {
	var info OSInfo
	for _, line := range strings.Split(string(data), "\n") {
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		v = strings.Trim(v, `"`)
		switch k {
		case "ID":
			info.ID = v
		case "VERSION_ID":
			info.Version = v
		case "PRETTY_NAME":
			info.Pretty = v
		}
	}
	return info
}

func ValidateOS(info OSInfo) error {
	var major int
	switch info.ID {
	case "ubuntu":
		if _, err := fmt.Sscanf(info.Version, "%d", &major); err != nil || major < 20 {
			return fmt.Errorf("ubuntu %q unsupported (need 20.04+)", info.Version)
		}
	case "debian":
		if _, err := fmt.Sscanf(info.Version, "%d", &major); err != nil || major < 11 {
			return fmt.Errorf("debian %q unsupported (need 11+)", info.Version)
		}
	default:
		return fmt.Errorf("%q is not ubuntu or debian", info.ID)
	}
	return nil
}

func DetectOS() (OSInfo, error) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return OSInfo{}, fmt.Errorf("cannot read /etc/os-release: %w", err)
	}
	info := ParseOSRelease(data)
	if err := ValidateOS(info); err != nil {
		return info, err
	}
	return info, nil
}
