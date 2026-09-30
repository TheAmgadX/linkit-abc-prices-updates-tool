package gui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails-bound presentation layer.
// It is the only struct exposed to the JavaScript frontend.
// It delegates all processing to the application service.
type App struct {
	ctx context.Context
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{}
}

// Startup is called by Wails when the application starts.
// The context is saved so that Wails runtime functions (dialogs etc.) can be called.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// SelectExcelFile opens a native OS file picker filtered to Excel files.
// Returns the selected path, or an empty string if cancelled.
func (a *App) SelectExcelFile() string {
	selection, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Select Excel File",
		Filters: []wailsruntime.FileFilter{
			{
				DisplayName: "Excel Files (*.xlsx, *.xls)",
				Pattern:     "*.xlsx;*.xls",
			},
		},
	})
	if err != nil {
		return ""
	}
	return selection
}

// SelectOutputDirectory opens a native OS directory picker.
// Returns the selected path, or an empty string if cancelled.
func (a *App) SelectOutputDirectory() string {
	selection, err := wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Select Output Directory",
	})
	if err != nil {
		return ""
	}
	return selection
}

// OpenOutputDirectory opens the given directory in the system file manager.
func (a *App) OpenOutputDirectory(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		// Linux — try common file managers
		for _, fm := range []string{"xdg-open", "nautilus", "dolphin", "thunar", "nemo"} {
			if _, err := exec.LookPath(fm); err == nil {
				cmd = exec.Command(fm, path)
				break
			}
		}
	}

	if cmd == nil {
		return fmt.Errorf("could not find a file manager to open the directory")
	}

	return cmd.Start()
}

// ProcessFiles validates inputs, runs the processing pipeline, and returns the result.
// This method runs synchronously in Go — Wails ensures it does not block the JS event loop.
func (a *App) ProcessFiles(input ProcessInput) ProcessingResult {
	return appService.Process(input)
}
