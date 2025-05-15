package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/codemi-be/golang-browser-mobile/pkg/utils"
)

// Android represents the Android app builder
type Android struct {
	RootDir     string
	ShellDir    string
	GradlewPath string
}

// NewAndroid creates a new Android builder
func NewAndroid(rootDir string) *Android {
	shellDir := filepath.Join(rootDir, "mobile-shell", "android")
	var gradlewPath string

	if runtime.GOOS == "windows" {
		gradlewPath = filepath.Join(shellDir, "gradlew.bat")
	} else {
		gradlewPath = filepath.Join(shellDir, "gradlew")
	}

	return &Android{
		RootDir:     rootDir,
		ShellDir:    shellDir,
		GradlewPath: gradlewPath,
	}
}

// Build builds the Android app
func (a *Android) Build() error {
	fmt.Println("Building Android app...")

	if _, err := os.Stat(a.GradlewPath); err != nil {
		return fmt.Errorf("Android build tools not found. Make sure the Android project is set up correctly: %w", err)
	}

	return utils.RunCmdWithDir(a.ShellDir, a.GradlewPath, "assembleDebug")
}

// InstallApp installs the app on the device
func (a *Android) InstallApp(deviceID string) error {
	fmt.Println("Installing Android app on device/emulator...")

	apkPath := filepath.Join(a.ShellDir, "app", "build", "outputs", "apk", "debug", "app-debug.apk")

	args := []string{"install", "-r"}
	if deviceID != "" {
		args = append(args, "-s", deviceID)
	}
	args = append(args, apkPath)

	return utils.RunCmd("adb", args...)
}

// LaunchApp launches the app on the device
func (a *Android) LaunchApp(deviceID string) error {
	fmt.Println("Launching Android app...")

	launchArgs := []string{"shell", "am", "start", "-n", "com.example.golangmobile/.MainActivity"}
	if deviceID != "" {
		launchArgs = []string{"-s", deviceID, "shell", "am", "start", "-n", "com.example.golangmobile/.MainActivity"}
	}

	return utils.RunCmd("adb", launchArgs...)
}

// SetupPortForwarding sets up port forwarding for development
func (a *Android) SetupPortForwarding(deviceID, port string) error {
	fmt.Println("Setting up port forwarding to device/emulator...")

	args := []string{"reverse", fmt.Sprintf("tcp:%s", port), fmt.Sprintf("tcp:%s", port)}
	if deviceID != "" {
		args = []string{"-s", deviceID, "reverse", fmt.Sprintf("tcp:%s", port), fmt.Sprintf("tcp:%s", port)}
	}

	return utils.RunCmd("adb", args...)
}
