//go:build !prod

package resources

import (
	"embed"
	"log/slog"
	"net/http"
	"os"

	"github.com/benbjohnson/hashfs"
)

var (
	//go:embed static
	StaticDirectory embed.FS
	StaticSys       = hashfs.NewFS(StaticDirectory)
)

func Handler() http.Handler {
	slog.Info("Serving static assets from dev server", "path", StaticDirectoryPath)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		http.StripPrefix("/static/", http.FileServerFS(os.DirFS(StaticDirectoryPath))).ServeHTTP(w, r)
	})
}

func StaticPath(path string) string {
	return "/static/" + path
}
