package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ListSerialPorts discovers connected cellular USB serial ports across Linux / Mac.
func ListSerialPorts() ([]string, error) {
	var ports []string

	if runtime.GOOS == "darwin" {
		matches, err := filepath.Glob("/dev/cu.usbmodem*")
		if err == nil {
			ports = append(ports, matches...)
		}
		matchesTty, err := filepath.Glob("/dev/tty.usbmodem*")
		if err == nil {
			ports = append(ports, matchesTty...)
		}
	} else {
		// Linux (Raspberry Pi, Ubuntu, etc.)
		usbMatches, err := filepath.Glob("/dev/ttyUSB*")
		if err == nil {
			ports = append(ports, usbMatches...)
		}
		acmMatches, err := filepath.Glob("/dev/ttyACM*")
		if err == nil {
			ports = append(ports, acmMatches...)
		}
	}

	// Filter existing device files
	var validPorts []string
	for _, p := range ports {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			validPorts = append(validPorts, p)
		}
	}

	if len(validPorts) == 0 {
		return nil, fmt.Errorf("no cellular modem USB serial ports found")
	}
	return validPorts, nil
}
