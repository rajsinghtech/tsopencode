package main

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

const plistLabel = "com.rajsinghtech.tsopencode"

const plistTmpl = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>{{.Label}}</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{.BinPath}}</string>{{range .Args}}
		<string>{{.}}</string>{{end}}
	</array>
	<key>EnvironmentVariables</key>
	<dict>
		<key>PATH</key>
		<string>/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin</string>{{if .AuthKey}}
		<key>TS_AUTHKEY</key>
		<string>{{.AuthKey}}</string>{{end}}
	</dict>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>{{.LogPath}}</string>
	<key>StandardErrorPath</key>
	<string>{{.LogPath}}</string>
</dict>
</plist>
`

type plistData struct {
	Label   string
	BinPath string
	Args    []string
	LogPath string
	AuthKey string
}

func plistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", plistLabel+".plist"), nil
}

func installService(authKey, hostname, stateDir, opencodeBin string) error {
	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find executable: %w", err)
	}

	var args []string
	if hostname != "opencode" {
		args = append(args, "--hostname", hostname)
	}
	if stateDir != defaultStateDir() {
		args = append(args, "--state-dir", stateDir)
	}
	if opencodeBin != "opencode" {
		args = append(args, "--opencode-bin", opencodeBin)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	logPath := filepath.Join(home, "Library", "Logs", "tsopencode.log")

	data := plistData{
		Label:   plistLabel,
		BinPath: binPath,
		Args:    args,
		LogPath: logPath,
		AuthKey: authKey,
	}

	path, err := plistPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	tmpl := template.Must(template.New("plist").Parse(plistTmpl))
	if err := tmpl.Execute(f, data); err != nil {
		f.Close()
		return err
	}
	f.Close()

	uid := strconv.Itoa(os.Getuid())
	// unload first in case already registered
	exec.Command("launchctl", "bootout", "gui/"+uid+"/"+plistLabel).Run()
	if out, err := exec.Command("launchctl", "bootstrap", "gui/"+uid, path).CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl bootstrap: %w\n%s", err, out)
	}

	fmt.Printf("tsopencode installed — runs at login\nlogs: %s\n", logPath)
	return nil
}

func uninstallService() error {
	uid := strconv.Itoa(os.Getuid())
	exec.Command("launchctl", "bootout", "gui/"+uid+"/"+plistLabel).Run()

	path, err := plistPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	fmt.Println("tsopencode uninstalled")
	return nil
}
