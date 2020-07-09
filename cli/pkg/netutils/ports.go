package netutils

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

func NextFreePort(port int) (int, error) {
	for p := port; p < 65535; p++ {
		if !PortIsOpen(p) {
			return p, nil
		}
	}
	return 0, fmt.Errorf("No free ports available")
}

func WaitForPort(port int, timeout time.Duration) error {
	start := time.Now()
	for {
		if PortIsOpen(port) {
			return nil
		}

		now := time.Now()
		if now.Sub(start) > timeout {
			return fmt.Errorf("Timed out")
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func PortIsOpen(port int) bool {
	return HostPortIsOpen("", port)
}

func HostPortIsOpen(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), 100*time.Millisecond)
	if conn != nil {
		conn.Close()
	}
	return err == nil
}
