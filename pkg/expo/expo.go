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
			return fmt.Errorf("failed to update package lists: %w", err)
		}
		if err := utils.RunCmd("apt-get", "install", "-y", "nodejs", "npm"); err != nil {
			return fmt.Errorf("failed to install Node.js and NPM: %w", err)
		}
	}

	// Create a symlink for expo in node_modules/.bin if available
	expoPath := filepath.Join(e.FrontendDir, "node_modules", ".bin", "expo")
	if _, err := os.Stat(expoPath); err == nil {
		// Create symlink to make expo available in PATH
		binDir := "/usr/local/bin"
		if err := os.MkdirAll(binDir, 0755); err != nil {
			fmt.Printf("Warning: Failed to create bin directory: %v\n", err)
		} else {
			symlinkPath := filepath.Join(binDir, "expo")
			// Remove existing symlink if it exists
			_ = os.Remove(symlinkPath)
			if err := os.Symlink(expoPath, symlinkPath); err != nil {
				fmt.Printf("Warning: Failed to create symlink to expo: %v\n", err)
			} else {
				fmt.Println("Created symlink to expo in " + binDir)
			}
		}
	}

	// Check if global expo-cli is installed or use local one
	if err := exec.Command("npx", "expo", "--version").Run(); err != nil {
		fmt.Println("Installing Expo CLI in project...")
		if err := utils.RunCmdWithDir(e.FrontendDir, "npm", "install", "--save-dev", "expo-cli"); err != nil {
			return fmt.Errorf("failed to install expo-cli in project: %w", err)
		}
	}

	// Install EAS CLI for building
	if err := exec.Command("npx", "eas", "--version").Run(); err != nil {
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

	// First check if expo is installed or use npx
	var cmd string = "expo"
	expoPath := filepath.Join(e.FrontendDir, "node_modules", ".bin", "expo")
	if _, err := os.Stat(expoPath); err != nil {
		// If we don't have direct access to expo in node_modules, use npx
		cmd = "npx"
	}

	var args []string
	switch platform {
	case "web":
		if cmd == "npx" {
			args = []string{"expo", "start", "--web"}
		} else {
			args = []string{"start", "--web"}
		}
	case "android":
		if cmd == "npx" {
			args = []string{"expo", "start", "--android"}
		} else {
			args = []string{"start", "--android"}
		}
	case "ios":
		if cmd == "npx" {
			args = []string{"expo", "start", "--ios"}
		} else {
			args = []string{"start", "--ios"}
		}
	default:
		if cmd == "npx" {
			args = []string{"expo", "start"}
		} else {
			args = []string{"start"}
		}
	}

	// Try running with the selected command
	err := utils.RunCmdWithDir(e.FrontendDir, cmd, args...)
	if err != nil {
		// If direct command failed, try with npm scripts as fallback
		fmt.Println("Falling back to npm scripts...")

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
