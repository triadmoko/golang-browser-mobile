package app

import (
	"fmt"
	"log"
	"sync"

	"github.com/codemi-be/golang-browser-mobile/pkg/builder"
	"github.com/codemi-be/golang-browser-mobile/pkg/config"
	"github.com/codemi-be/golang-browser-mobile/pkg/server"
)

// App represents the main application
type App struct {
	Config   *config.Config
	Frontend *builder.Frontend
	Android  *builder.Android
	IOS      *builder.IOS
	Server   *server.PreviewServer
}

// New creates a new application instance
func New(cfg *config.Config) *App {
	return &App{
		Config:   cfg,
		Frontend: builder.NewFrontend(cfg.RootDir),
		Android:  builder.NewAndroid(cfg.RootDir),
		IOS:      builder.NewIOS(cfg.RootDir),
		Server:   server.NewPreviewServer(cfg.RootDir, cfg.PreviewPort),
	}
}

// Run executes the application based on the configuration
func (a *App) Run() error {

	// Run in development mode
	if a.Config.DevMode {
		return a.runDevMode()
	}

	// Run in production/build mode
	return a.runBuildMode()
}

// runDevMode runs the application in development mode
func (a *App) runDevMode() error {
	fmt.Println("Running in development mode...")

	// Development with WebView
	var wg sync.WaitGroup
	wg.Add(1)

	// Start frontend dev server
	if err := a.Frontend.StartDevServer(); err != nil {
		return fmt.Errorf("failed to start dev server: %w", err)
	}

	// If previewing on device/emulator
	if a.Config.Preview {
		if err := a.setupDevicePreview(); err != nil {
			log.Printf("Warning: Preview setup issue: %v", err)
		}
	} else {
		// Just serve the assets locally for quick preview
		a.Server.StartBackground()
	}

	// Keep the program running
	wg.Wait()
	return nil
}

// setupDevicePreview sets up preview on a device
func (a *App) setupDevicePreview() error {
	// Set up port forwarding for Android
	if a.Config.DeviceID != "" && a.Config.BuildAndroid {
		if err := a.Android.SetupPortForwarding(a.Config.DeviceID, a.Config.DevPort); err != nil {
			return fmt.Errorf("port forwarding failed: %w", err)
		}
	}

	// Launch on Android or iOS
	if a.Config.BuildAndroid {
		if err := a.launchAndroidPreview(); err != nil {
			return err
		}
	} else if a.Config.BuildIOS {
		if err := a.launchIOSPreview(); err != nil {
			return err
		}
	}

	return nil
}

// launchAndroidPreview installs and launches the app on an Android device
func (a *App) launchAndroidPreview() error {
	if err := a.Android.InstallApp(a.Config.DeviceID); err != nil {
		return fmt.Errorf("android app installation failed: %w", err)
	}

	if err := a.Android.LaunchApp(a.Config.DeviceID); err != nil {
		return fmt.Errorf("android app launch failed: %w", err)
	}

	return nil
}

// launchIOSPreview installs and launches the app on an iOS device
func (a *App) launchIOSPreview() error {
	if err := a.IOS.InstallApp(a.Config.DeviceID); err != nil {
		return fmt.Errorf("iOS app installation failed: %w", err)
	}

	if err := a.IOS.LaunchApp(a.Config.DeviceID); err != nil {
		return fmt.Errorf("iOS app launch failed: %w", err)
	}

	return nil
}

// runBuildMode builds the application for production
func (a *App) runBuildMode() error {
	return a.buildWithWebView()
}

// buildWithWebView builds the app using WebView approach
func (a *App) buildWithWebView() error {
	// 1. Build frontend
	if err := a.Frontend.Build(); err != nil {
		return fmt.Errorf("frontend build failed: %w", err)
	}

	// 2. Copy build output to mobile shell
	if err := a.Frontend.CopyBuildToMobile(); err != nil {
		return fmt.Errorf("copying build output failed: %w", err)
	}

	// 3. Build mobile apps if requested
	if a.Config.BuildAndroid {
		if err := a.Android.Build(); err != nil {
			return fmt.Errorf("android build failed: %w", err)
		}
	}

	if a.Config.BuildIOS {
		if err := a.IOS.Build(); err != nil {
			return fmt.Errorf("iOS build failed: %w", err)
		}
	}

	// 4. Preview on device/simulator if requested
	if a.Config.Preview {
		if a.Config.BuildAndroid {
			if err := a.launchAndroidPreview(); err != nil {
				return fmt.Errorf("android preview failed: %w", err)
			}
		} else if a.Config.BuildIOS {
			if err := a.launchIOSPreview(); err != nil {
				return fmt.Errorf("iOS preview failed: %w", err)
			}
		}
	}

	fmt.Println("Build process completed successfully!")
	return nil
}
