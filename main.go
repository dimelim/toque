package main

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

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

func instanceID() string {
	exe, _ := os.Executable()
	sum := sha256.Sum256([]byte(strings.ToLower(filepath.Clean(exe))))
	return "toque-" + hex.EncodeToString(sum[:8])
}

func main() {
	if runCLI() {
		return
	}
	var started atomic.Value
	withContext := func(fn func(context.Context)) {
		if ctx, ok := started.Load().(context.Context); ok {
			fn(ctx)
		}
	}
	svc := app.New(presentMonPath(), func(name string, data any) {
		withContext(func(ctx context.Context) { runtime.EventsEmit(ctx, name, data) })
	})
	err := wails.Run(&options.App{
		Title:            "Toque",
		Width:            1200,
		Height:           780,
		MinWidth:         1040,
		MinHeight:        680,
		Frameless:        true,
		BackgroundColour: &options.RGBA{R: 15, G: 18, B: 13, A: 255},
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup: func(ctx context.Context) {
			started.Store(ctx)
			svc.Recover()
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: instanceID(),
			OnSecondInstanceLaunch: func(options.SecondInstanceData) {
				withContext(func(ctx context.Context) {
					runtime.WindowUnminimise(ctx)
					runtime.WindowShow(ctx)
				})
			},
		},
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
