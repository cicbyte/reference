package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewReferenceApp()

	err := wails.Run(&options.App{
		Title:     "Reference",
		Width:     1200,
		Height:    800,
		MinWidth:  800,
		MinHeight: 600,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
			// Inject a Content-Security-Policy on every response.
			// - 'unsafe-eval': vue-i18n compiles message formats ({name} interpolation)
			//   at runtime via new Function; this is safe (no external input reaches it,
			//   DOMPurify handles untrusted HTML separately).
			// - 'unsafe-inline' style-src: Ant Design Vue injects inline styles.
			// - ws:/wails.localhost: dev-mode HMR websocket (harmless in production builds).
			// - img data:/blob:: mermaid PNG export.
			Middleware: func(next http.Handler) http.Handler {
				const csp = "default-src 'self'; script-src 'self' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self' data:; connect-src 'self' ws://wails.localhost:* ws://localhost:*; object-src 'none'; base-uri 'self'; frame-ancestors 'none'"
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Security-Policy", csp)
					next.ServeHTTP(w, r)
				})
			},
		},
		BackgroundColour: &options.RGBA{R: 15, G: 23, B: 42, A: 255},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: false,
			Theme:                             windows.SystemDefault,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
