package main

import (
	"crypto/rand"
	"embed"
	"github.com/filinvadim/badger-gui/database"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"log"
	"net/http"
	"strings"
)

//go:embed frontend/dist
var assets embed.FS

//go:embed icon.png
var icon []byte

// utf8Middleware ensures proper UTF-8 charset headers for HTML, CSS, and JS files
func utf8Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set charset for HTML files
		if strings.HasSuffix(r.URL.Path, ".html") || r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		} else if strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		} else if strings.HasSuffix(r.URL.Path, ".json") {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	db, err := database.New(nil)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	app := NewApp(db)

	setLinuxDesktopIcon(icon)

	err = wails.Run(&options.App{
		Title:            "badger-gui",
		Width:            1024,
		Height:           1024,
		WindowStartState: options.Maximised,
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Middleware: assetserver.ChainMiddleware(utf8Middleware),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.close,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: rand.Text(),
			OnSecondInstanceLaunch: func(_ options.SecondInstanceData) {
				panic("second instance launched")
			},
		},
		Bind: []interface{}{
			app,
		},
		Linux: &linux.Options{
			Icon:                icon,
			WindowIsTranslucent: false,
			WebviewGpuPolicy:    linux.WebviewGpuPolicyNever,
			ProgramName:         "badger-gui",
		},
		Mac: &mac.Options{
			TitleBar:             nil,
			Appearance:           mac.DefaultAppearance,
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			Preferences:          nil,
			DisableZoom:          false,
			About: &mac.AboutInfo{
				Title:   "badger-gui",
				Message: "",
				Icon:    icon,
			},
			OnFileOpen: nil,
			OnUrlOpen:  nil,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:                false,
			WindowIsTranslucent:                 false,
			DisableFramelessWindowDecorations:   false,
			Theme:                               windows.Dark,
			BackdropType:                        windows.Auto,
			Messages:                            nil,
			ResizeDebounceMS:                    0,
			OnSuspend:                           nil,
			OnResume:                            nil,
			WebviewGpuIsDisabled:                true,
			WebviewDisableRendererCodeIntegrity: false,
			EnableSwipeGestures:                 false,
			WindowClassName:                     "badger-gui",
		},
	})
	if err != nil {
		db.Close()
		log.Fatalf("failed to start application: %s", err)
	}
}
