//go:build linux

package cmd

import (
	"fmt"
	"os/exec"
)

// /curl
func CurlCommand(target string) string {
	cmd := exec.Command("curl", "-v", "-IL", "--max-time", "10", target)
	out, err := cmd.CombinedOutput()
	raw := string(out)
	if err != nil {
		raw = fmt.Sprintf("Curl error: %v\nOutpout:\n%s", err, raw)
	}
	return raw
}

// /ping
func PingCommand(target string) string {
	cmd := exec.Command("ping", "-c", "5", target)
	out, err := cmd.CombinedOutput()
	raw := string(out)
	if err != nil {
		raw = fmt.Sprintf("Ping error: %v\nOutpout:\n%s", err, raw)
	}
	return raw
}

// /nslookup
func NslookupCommand(target string) string {
	cmd := exec.Command("nslookup", target)
	out, err := cmd.CombinedOutput()
	raw := string(out)
	if err != nil {
		raw = fmt.Sprintf("Nslookup error: %v\nOutpout:\n%s", err, raw)
	}
	return raw
}
