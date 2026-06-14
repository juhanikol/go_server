package httpserver

import (
	"embed"
	"io/fs"
)

// Embedded fallback assets used by GoServer itself.
//
// These files live inside pkg/httpserver/assets and are compiled directly
// into the Go binary. This means GoServer can always access its own
// internal fallback templates and CSS even when another project imports
// the library remotely.
//
//go:embed assets/*
var embeddedAssets embed.FS

// InternalAssetRoutePrefix is the reserved URL prefix used by GoServer
// to serve its own embedded static assets such as fallback CSS.
//
// Keeping this separate avoids collisions with the importing project's
// own /static/... routes.
const InternalAssetRoutePrefix = "/__go_server/static/"

// getEmbeddedAssetsFS returns a filesystem rooted at "assets/" so it can
// be served with http.FileServer when needed.
func getEmbeddedAssetsFS() (fs.FS, error) {
	return fs.Sub(embeddedAssets, "assets")
}

// readEmbeddedAsset reads one embedded asset file by simple name.
//
// Example:
// - "serverindex.html"
// - "servererror.html"
// - "styles.css"
func readEmbeddedAsset(fileName string) ([]byte, error) {
	return embeddedAssets.ReadFile("assets/" + fileName)
}
