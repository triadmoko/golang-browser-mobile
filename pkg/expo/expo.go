package expo

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/codemi-be/golang-browser-mobile/pkg/utils"
)

// Expo represents the Expo app builder and runner
type Expo struct {
	RootDir     string
	FrontendDir string
}

// NewExpo creates a new Expo instance
func NewExpo(rootDir string) *Expo {
	return &Expo{
		RootDir:     rootDir,
		FrontendDir: filepath.Join(rootDir, "frontend"),
	}
}

// Setup initializes the Expo environment
func (e *Expo) Setup() error {
	fmt.Println("Setting up Expo development environment...")

	// First, ensure dependencies are installed in the frontend project
	if err := e.InstallDependencies(); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// Check if npx is available
	if err := exec.Command("which", "npx").Run(); err != nil {
		fmt.Println("Installing NPX (part of Node.js)...")
		if err := utils.RunCmd("apt-get", "update"); err != nil {
			// If system packages can't be updated (permissions), inform the user
			fmt.Println("Warning: Could not update package lists. You may need sudo privileges.")
			fmt.Println("Please ensure Node.js and NPM are installed on your system.")
		} else if err := utils.RunCmd("apt-get", "install", "-y", "nodejs", "npm"); err != nil {
			fmt.Println("Warning: Could not install Node.js and NPM. You may need sudo privileges.")
			fmt.Println("Please ensure Node.js and NPM are installed on your system.")
		}
	}

	// Skip symlink creation that requires root - just note the location
	expoPath := filepath.Join(e.FrontendDir, "node_modules", ".bin", "expo")
	if _, err := os.Stat(expoPath); err == nil {
		fmt.Println("Local expo found at " + expoPath)
		fmt.Println("You can add this to your PATH temporarily with:")
		fmt.Println("export PATH=\"" + expoPath + ":$PATH\"")
	}

	// Check if npx can run expo - use this approach rather than global installation
	if err := exec.Command("npx", "--no-install", "expo", "--version").Run(); err != nil {
		fmt.Println("Installing Expo CLI in project...")
		if err := utils.RunCmdWithDir(e.FrontendDir, "npm", "install", "--save-dev", "expo-cli"); err != nil {
			return fmt.Errorf("failed to install expo-cli in project: %w", err)
		}
	}

	// Install EAS CLI locally in the project
	if err := exec.Command("npx", "--no-install", "eas", "--version").Run(); err != nil {
		fmt.Println("Installing EAS CLI in project...")
		if err := utils.RunCmdWithDir(e.FrontendDir, "npm", "install", "--save-dev", "eas-cli"); err != nil {
			return fmt.Errorf("failed to install eas-cli in project: %w", err)
		}
	}

	return e.CreateEasConfig()
}

// CreateEasConfig creates or updates the EAS configuration file
func (e *Expo) CreateEasConfig() error {
	fmt.Println("Creating EAS configuration...")

	config := map[string]interface{}{
		"build": map[string]interface{}{
			"development": map[string]interface{}{
				"developmentClient": true,
				"distribution":      "internal",
			},
			"preview": map[string]interface{}{
				"distribution": "internal",
			},
			"production": map[string]interface{}{},
		},
	}

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	configPath := filepath.Join(e.FrontendDir, "eas.json")
	return os.WriteFile(configPath, configJSON, 0644)
}

// StartDevServer starts the Expo development server
func (e *Expo) StartDevServer(platform string) error {
	fmt.Println("Starting Expo development server for", platform, "...")

	// First, try using npx which is the most reliable
	var npxArgs []string

	switch platform {
	case "web":
		npxArgs = []string{"expo", "start", "--web"}
	case "android":
		npxArgs = []string{"expo", "start", "--android"}
	case "ios":
		npxArgs = []string{"expo", "start", "--ios"}
	default:
		npxArgs = []string{"expo", "start"}
	}

	err := utils.RunCmdWithDir(e.FrontendDir, "npx", npxArgs...)
	if err != nil {
		fmt.Println("Direct execution with npx failed, trying npm scripts as fallback...")

		// If npx approach fails, try with npm scripts as fallback
		var npmArgs []string
		switch platform {
		case "web":
			npmArgs = []string{"run", "expo-web"}
		case "android":
			npmArgs = []string{"run", "expo-android"}
		case "ios":
			npmArgs = []string{"run", "expo-ios"}
		default:
			npmArgs = []string{"run", "expo-start"}
		}

		return utils.RunCmdWithDir(e.FrontendDir, "npm", npmArgs...)
	}

	return nil
}

// BuildApp builds the app using EAS for the specified platform
func (e *Expo) BuildApp(platform string) error {
	fmt.Println("Building Expo app for", platform, "...")

	var args []string

	switch platform {
	case "android":
		args = []string{"eas", "build", "--platform", "android", "--profile", "production"}
	case "ios":
		args = []string{"eas", "build", "--platform", "ios", "--profile", "production"}
	default:
		return fmt.Errorf("unsupported platform for Expo build: %s", platform)
	}

	return utils.RunCmdWithDir(e.FrontendDir, "npx", args...)
}

// InstallDependencies installs all Expo dependencies
func (e *Expo) InstallDependencies() error {
	fmt.Println("Installing Expo dependencies...")
	return utils.RunCmdWithDir(e.FrontendDir, "npm", "install", "--legacy-peer-deps")
}
