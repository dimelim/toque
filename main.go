package main

import (
	"context"
	"embed"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"toque/internal/app"
	"toque/internal/winx"
)

//go:embed all:frontend/dist
var assets embed.FS

func presentMonPath() string {
	exe, _ := os.Executable()
	for _, p := range []string{
		filepath.Join(filepath.Dir(exe), "PresentMon.exe"),
		filepath.Join(filepath.Dir(exe), "..", "..", ".dev", "bin", "PresentMon.exe"),
	} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "PresentMon.exe"
}

func main() {
	if runCLI() {
		return
	}
	var wctx context.Context
	svc := app.New(presentMonPath(), func(name string, data any) {
		if wctx != nil {
			runtime.EventsEmit(wctx, name, data)
		}
	})
	svc.StartSession()
	defer svc.StopSession()
	err := wails.Run(&options.App{
		Title:            "Toque",
		Width:            1200,
		Height:           780,
		MinWidth:         1040,
		MinHeight:        680,
		Frameless:        true,
		BackgroundColour: &options.RGBA{R: 15, G: 18, B: 13, A: 255},
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        func(ctx context.Context) { wctx = ctx },
		OnShutdown: func(ctx context.Context) { svc.StopSession() },
		Bind: []interface{}{svc},
		Windows: &windows.Options{
			Theme:                windows.Dark,
			WebviewUserDataPath:  filepath.Join(winx.LocalAppData(), "Toque", "WebView2"),
			DisableWindowIcon:    false,
			WebviewGpuIsDisabled: false,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
