package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

var serialMutex sync.Mutex

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

// ExecATCommand sends a raw AT command to a physical serial port with global mutex lock protection.
func ExecATCommand(port string, cmd string, waitDuration time.Duration) (string, error) {
	serialMutex.Lock()
	defer serialMutex.Unlock()

	fd, err := syscall.Open(port, syscall.O_RDWR|syscall.O_NONBLOCK, 0666)
	if err != nil {
		return "", err
	}
	defer syscall.Close(fd)

	if !strings.HasSuffix(cmd, "\r\n") {
		cmd += "\r\n"
	}

	_, err = syscall.Write(fd, []byte(cmd))
	if err != nil {
		return "", err
	}

	if waitDuration <= 0 {
		waitDuration = 300 * time.Millisecond
	}

	deadline := time.Now().Add(waitDuration)
	var output []byte
	buf := make([]byte, 4096)

	for time.Now().Before(deadline) {
		time.Sleep(30 * time.Millisecond)
		n, _ := syscall.Read(fd, buf)
		if n > 0 {
			output = append(output, buf[:n]...)
			if strings.Contains(string(output), "OK") || strings.Contains(string(output), "ERROR") {
				break
			}
		}
	}

	return string(output), nil
}
