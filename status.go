package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runStatus(stateDir string) {
	url := cachedURL(stateDir)
	pid := servicePID()

	if pid > 0 {
		fmt.Printf("running  (PID %d)\n", pid)
	} else {
		fmt.Println("stopped")
	}

	if url != "" {
		fmt.Printf("https    https://%s\n", url)
		fmt.Printf("http     http://%s\n", url)
	} else {
		fmt.Println("url      unknown (not yet connected)")
	}

	home, _ := os.UserHomeDir()
	fmt.Printf("logs     %s\n", filepath.Join(home, "Library", "Logs", "tsopencode.log"))
	fmt.Printf("state    %s\n", stateDir)
}

func cachedURL(stateDir string) string {
	b, err := os.ReadFile(filepath.Join(stateDir, "url"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func writeURL(stateDir, dnsName string) {
	os.MkdirAll(stateDir, 0755)
	os.WriteFile(filepath.Join(stateDir, "url"), []byte(dnsName+"\n"), 0644)
}

// servicePID returns the PID of the running launchd service, or 0.
func servicePID() int {
	out, err := exec.Command("launchctl", "list", "homebrew.mxcl.tsopencode").Output()
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, `"PID"`) {
			var pid int
			fmt.Sscanf(line, `"PID" = %d;`, &pid)
			return pid
		}
	}
	return 0
}
