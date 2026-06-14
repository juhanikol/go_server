package httpserver

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// GoServerRouteRegisterer is the modular "Plug-in" interface.
// If you import this package elsewhere, you can create a struct that
// implements this to inject routes without modifying this file.
type GoServerRouteRegisterer interface {
	RegisterGoServerRoutes(GoServer *GoServer)
}

// GenericTemplateData demonstrates how to pass data to HTML files.
// Use this as a reference for your own project-specific data structs.
type GenericTemplateData struct {
	PageTitle    string
	MainHeading  string
	DisplayValue string
	ErrorMessage string
	IsSuccess    bool
}

// apiEndpointInfo contains the data for a single endpoint displayed on the server index page.
type apiEndpointInfo struct {
	Path        string
	Description string
}

// serverIndexData is the data structure for the internal serverindex.html template.
// It matches all fields used by that template.
type serverIndexData struct {
	PageTitle    string
	MainHeading  string
	DisplayValue string
	ErrorMessage string
	IsSuccess    bool
	APIEndpoints []apiEndpointInfo
}

// ScanProjectResources performs automated discovery of templates and sets up auto-routing.
// It fulfills the "Zero-Boilerplate" goal by making files immediately available.
// ScanProjectResources performs Phase 1: Discovery.
func (GoServer *GoServer) ScanProjectResources() {
	if GoServer.Manifest.TemplateDir == "" {
		return
	}

	if _, err := os.Stat(GoServer.Manifest.TemplateDir); os.IsNotExist(err) {
		GoServer.GoServerLogger.Error("CRITICAL: TemplateDir missing", "path", GoServer.Manifest.TemplateDir)
		return
	}

	filepath.Walk(GoServer.Manifest.TemplateDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".html") {
			Tmpl, _ := template.New(info.Name()).ParseFiles(path)
			GoServer.TemplateCache[info.Name()] = Tmpl
			GoServer.autoRegisterTemplateRoute(info.Name())
		}
		return nil
	})
}

// autoRegisterTemplateRoute creates an implicit GET route for a template file.
func (GoServer *GoServer) autoRegisterTemplateRoute(FileName string) {
	if FileName == "index.html" {
		return
	}

	// Calculate route: "contact.html" -> "/contact"
	RoutePath := "/" + strings.TrimSuffix(FileName, ".html")

	// Logic: If the developer already defined this path in the Manifest, do nothing.
	if _, exists := GoServer.Manifest.RouteMap[RoutePath]; exists {
		return
	}

	// 2. Check if the path was already registered via RegisterRoute() or SetHomeRoute().
	if GoServer.RegisteredPaths[RoutePath] {
		GoServer.GoServerLogger.Info("Scanner: Skipping auto-route; path already claimed", "path", RoutePath)
		return
	}

	// Register a simple "Just Render" handler.
	GoServer.RegisterRoute(RoutePath, http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		GoServer.RenderGoServerTemplate(w, FileName, nil, http.StatusOK)
	})
}

// AddDefaultGoServerRoutes registers the base routes for the server.
// Project-specific API routes should be registered by the developer here BUT preferably in another module.
func (GoServer *GoServer) AddDefaultGoServerRoutes() {

	GoServer.ensureHomeRoute()
	GoServer.GoServerRouter.Handle("/{$}", GoServer.GoServerHomeHandler)

	// Static files: Use the manifest path if provided, otherwise default to web/static.
	StaticPath := GoServer.Manifest.StaticDir
	if StaticPath == "" {
		StaticPath = "web/static"
	}

	GoServer.GoServerRouter.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(StaticPath))))

	// GoServer internal embedded assets such as fallback CSS.
	GoServer.registerInternalAssetRoutes()
	// SILENT HEALTH CHECK
	// use only GoServerRouter.HandleFunc instead of GoServer.GoServerHandler.
	// This avoids the LoggingMiddleware entirely for this path.
	GoServer.GoServerRouter.HandleFunc("/health", GoServer.HandleHealthCheck())
}

// registerInternalAssetRoutes serves embedded GoServer-owned assets such as
// the internal fallback stylesheet. These are separate from the importing
// project's own static files.
func (GoServer *GoServer) registerInternalAssetRoutes() {
	EmbeddedFS, err := getEmbeddedAssetsFS()
	if err != nil {
		GoServer.GoServerLogger.Error("Failed to initialize embedded assets filesystem", "error", err)
		return
	}

	EmbeddedFileServer := http.FileServer(http.FS(EmbeddedFS))
	GoServer.GoServerRouter.Handle(
		InternalAssetRoutePrefix,
		http.StripPrefix(InternalAssetRoutePrefix, LoggingMiddleware(EmbeddedFileServer)),
	)
}

// ensureHomeRoute guarantees that "/" always has a valid handler.
func (GoServer *GoServer) ensureHomeRoute() {
	if GoServer.GoServerHomeHandler != nil {
		return
	}

	// Fallback 1: If index.html was scanned, use it.
	if _, exists := GoServer.TemplateCache["index.html"]; exists {
		GoServer.SetHomeRoute(func(w http.ResponseWriter, r *http.Request) {
			GoServer.RenderGoServerTemplate(w, "index.html", nil, http.StatusOK)
		})
		return
	}

	// Fallback 2: Use internal GoServer landing page.
	GoServer.GoServerHomeHandler = GoServer.BuildRouteHandler(http.MethodGet, GoServer.HandleDefaultLandingPage())
}

// HandleDefaultLandingPage is the built-in fallback homepage for GoServer.
// Importing projects may replace this by calling SetHomeRoute or SetHomeTemplate.
// This is displayed if developers have not defined home route
func (GoServer *GoServer) HandleDefaultLandingPage() http.HandlerFunc {
	return func(ResponseWriter http.ResponseWriter, Request *http.Request) {
		if Request.URL.Path != "/" {
			// Instead of white 404, show the nice GoServer error page
			GoServer.RenderGoServerError(ResponseWriter, GoServerError{
				StatusCode: http.StatusNotFound,
				Title:      "Page Not Found",
				Message:    "The requested resource does not exist on this server.",
			})
			return
		}

		LandingData := serverIndexData{
			PageTitle:    "GoServer Home 404",
			MainHeading:  "(404) Welcome to your Modular Server",
			DisplayValue: "No custom homepage was registered, so GoServer is displaying its own landing page.",
			IsSuccess:    true,
			APIEndpoints: nil,
		}

		GoServer.renderEmbeddedTemplateOrFallback(ResponseWriter, "serverindex.html", LandingData, http.StatusOK)
	}
}

// RenderGoServerHome should be used by the project's homepage handler.
// It tries the project's own homepage template first.
// If that fails or is missing, GoServer shows serverindex.html.
// If even that fails, GoServer shows hardcoded fallback HTML.
func (GoServer *GoServer) RenderGoServerHome(ResponseWriter http.ResponseWriter, TemplateName string, TemplateData any, Status int) {
	if GoServer.renderProjectTemplate(ResponseWriter, TemplateName, TemplateData, Status) {
		return
	}

	GoServer.GoServerLogger.Warn("Homepage template failed, switching to GoServer internal index page",
		"requested_template", TemplateName,
	)

	FallbackData := serverIndexData{
		PageTitle:    "GoServer Internal Homepage",
		MainHeading:  "GoServer Internal Landing Page",
		DisplayValue: fmt.Sprintf("The requested homepage template %q could not be loaded. GoServer is showing its internal landing page instead.", TemplateName),
		ErrorMessage: "The project homepage could not be rendered.",
		IsSuccess:    false,
		APIEndpoints: nil,
	}

	GoServer.renderEmbeddedTemplateOrFallback(ResponseWriter, "serverindex.html", FallbackData, Status)
}

// RenderGoServerTemplate is the normal renderer for project pages other than homepage.
// If the requested page fails, GoServer switches to the internal error page.
func (GoServer *GoServer) RenderGoServerTemplate(ResponseWriter http.ResponseWriter, TemplateName string, TemplateData any, Status int) {

	// 1. Check the Cache first
	Tmpl, exists := GoServer.TemplateCache[TemplateName]

	if exists {
		var RenderBuffer bytes.Buffer
		if err := Tmpl.Execute(&RenderBuffer, TemplateData); err == nil {
			ResponseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
			ResponseWriter.WriteHeader(Status)
			RenderBuffer.WriteTo(ResponseWriter)
			return // SUCCESS: Exit here
		}
	}

	// 2. Fallback to Disk (Legacy/intended)
	if GoServer.renderProjectTemplate(ResponseWriter, TemplateName, TemplateData, Status) {
		return // SUCCESS: Exit here
	}

	if !exists {
		// If missing, trigger the standard error handler
		GoServer.RenderGoServerError(ResponseWriter, GoServerError{
			StatusCode:   http.StatusNotFound,
			Title:        "Template Missing",
			Message:      fmt.Sprintf("The template '%s' was not found in the cache OR \n The requested page could not be rendered, so GoServer switched to its internal error page.", TemplateName),
			TechnicalErr: "ensure file exists in Manifest.TemplateDir",
		})
		return
	}

	// 2. Execute from memory
	var RenderBuffer bytes.Buffer
	if err := Tmpl.Execute(&RenderBuffer, TemplateData); err != nil {
		GoServer.GoServerLogger.Error("Template execution failed", "template", TemplateName, "error", err)
		GoServer.RenderGoServerError(ResponseWriter, GoServerError{
			StatusCode:   http.StatusInternalServerError,
			Title:        "Rendering Error",
			Message:      "The requested page could not be displayed.",
			TechnicalErr: err.Error(),
		})
		return
	}

	// 3. Ultimate Failure
	GoServer.RenderGoServerError(ResponseWriter, GoServerError{
		StatusCode: http.StatusNotFound,
		Title:      "Template Missing",
		Message:    fmt.Sprintf("The template '%s' was not found in cache or on disk.", TemplateName),
	})
}

// renderProjectTemplate tries only the project-requested template.
// It does not try internal templates and does not recurse into error rendering.
func (GoServer *GoServer) renderProjectTemplate(ResponseWriter http.ResponseWriter, TemplateName string, TemplateData any, Status int) bool {
	TemplatePath := filepath.Join("web", "templates", TemplateName)

	if _, err := os.Stat(TemplatePath); err != nil {
		GoServer.GoServerLogger.Warn("Project template missing",
			"template", TemplateName,
			"path", TemplatePath,
			"error", err,
		)
		return false
	}

	return GoServer.renderTemplateFromDisk(ResponseWriter, TemplateName, TemplatePath, TemplateData, Status)
}

// renderTemplateFromDisk parses and executes a project-owned template file
// from the filesystem using buffered rendering.
func (GoServer *GoServer) renderTemplateFromDisk(ResponseWriter http.ResponseWriter, TemplateName string, TemplatePath string, TemplateData any, Status int) bool {
	Tmpl, err := template.New(TemplateName).ParseFiles(TemplatePath)
	if err != nil {
		GoServer.GoServerLogger.Error("Template parse failed",
			"template", TemplateName,
			"path", TemplatePath,
			"error", err,
		)
		return false
	}

	var RenderBuffer bytes.Buffer

	if err := Tmpl.Execute(&RenderBuffer, TemplateData); err != nil {
		GoServer.GoServerLogger.Error("Template execute failed",
			"template", TemplateName,
			"path", TemplatePath,
			"error", err,
		)
		return false
	}

	ResponseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
	ResponseWriter.WriteHeader(Status)

	if _, err := RenderBuffer.WriteTo(ResponseWriter); err != nil {
		GoServer.GoServerLogger.Error("Buffered response write failed",
			"template", TemplateName,
			"path", TemplatePath,
			"error", err,
		)
		return false
	}

	return true
}

// renderEmbeddedTemplate renders one GoServer-owned embedded template file
// using buffered rendering.
func (GoServer *GoServer) renderEmbeddedTemplate(ResponseWriter http.ResponseWriter, TemplateName string, TemplateData any, Status int) bool {
	TemplateBytes, err := readEmbeddedAsset(TemplateName)
	if err != nil {
		GoServer.GoServerLogger.Error("Embedded template read failed",
			"template", TemplateName,
			"error", err,
		)
		return false
	}

	Tmpl, err := template.New(TemplateName).Parse(string(TemplateBytes))
	if err != nil {
		GoServer.GoServerLogger.Error("Embedded template parse failed",
			"template", TemplateName,
			"error", err,
		)
		return false
	}

	var RenderBuffer bytes.Buffer

	if err := Tmpl.Execute(&RenderBuffer, TemplateData); err != nil {
		GoServer.GoServerLogger.Error("Embedded template execute failed",
			"template", TemplateName,
			"error", err,
		)
		return false
	}

	ResponseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
	ResponseWriter.WriteHeader(Status)

	if _, err := RenderBuffer.WriteTo(ResponseWriter); err != nil {
		GoServer.GoServerLogger.Error("Buffered embedded response write failed",
			"template", TemplateName,
			"error", err,
		)
		return false
	}

	return true
}

// renderEmbeddedTemplateOrFallback tries one embedded template first.
// If that fails, it falls back to hardcoded HTML.
func (GoServer *GoServer) renderEmbeddedTemplateOrFallback(ResponseWriter http.ResponseWriter, TemplateName string, TemplateData any, Status int) {
	if GoServer.renderEmbeddedTemplate(ResponseWriter, TemplateName, TemplateData, Status) {
		return
	}

	GoServer.GoServerLogger.Warn("Embedded internal template failed, using hardcoded fallback",
		"template", TemplateName,
	)

	GoServer.renderFallbackHTML(ResponseWriter, TemplateData, Status)
}

// renderDirectTemplate tries to parse and execute one exact template file. (legacy but not removed yet)
// It returns true on success, false on any failure.
// IMPORTANT:
// The template is first executed into memory, and only after successful
// execution are headers and body written to the real ResponseWriter.
// This prevents partial or broken output from being sent before GoServer
// has a chance to switch to a fallback page.
func (GoServer *GoServer) renderDirectTemplate(ResponseWriter http.ResponseWriter, TemplateName string, TemplateData any, Status int) bool {
	TemplatePath := filepath.Join("web", "templates", TemplateName)

	Tmpl, err := template.New(TemplateName).ParseFiles(TemplatePath)
	if err != nil {
		GoServer.GoServerLogger.Error("Template parse failed",
			"template", TemplateName,
			"path", TemplatePath,
			"error", err,
		)
		return false
	}

	var RenderBuffer bytes.Buffer

	if err := Tmpl.Execute(&RenderBuffer, TemplateData); err != nil {
		GoServer.GoServerLogger.Error("Template execute failed",
			"template", TemplateName,
			"path", TemplatePath,
			"error", err,
		)
		return false
	}

	ResponseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
	ResponseWriter.WriteHeader(Status)

	if _, err := RenderBuffer.WriteTo(ResponseWriter); err != nil {
		GoServer.GoServerLogger.Error("Buffered response write failed",
			"template", TemplateName,
			"path", TemplatePath,
			"error", err,
		)
		return false
	}

	return true
}

// renderDirectTemplateOrServerIndex is a safe helper for homepage fallback (legacy but not removed yet).
func (GoServer *GoServer) renderDirectTemplateOrServerIndex(ResponseWriter http.ResponseWriter, TemplateName string, TemplateData any, Status int) {
	if GoServer.renderDirectTemplate(ResponseWriter, TemplateName, TemplateData, Status) {
		return
	}

	GoServer.GoServerLogger.Warn("Internal homepage template failed, using hardcoded fallback",
		"template", TemplateName,
	)

	GoServer.renderFallbackHTML(ResponseWriter, TemplateData, Status)
}

// HandleHealthCheck is a simple 'ping' response for monitoring tools.
func (GoServer *GoServer) HandleHealthCheck() http.HandlerFunc {
	return func(ResponseWriter http.ResponseWriter, Request *http.Request) {
		ResponseWriter.WriteHeader(http.StatusOK)
		_, _ = ResponseWriter.Write([]byte("OK"))
	}
}

func (GoServer *GoServer) renderFallbackHTML(ResponseWriter http.ResponseWriter, Data any, Status int) {
	ResponseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
	ResponseWriter.WriteHeader(Status)

	fmt.Fprintf(ResponseWriter, `
		<html>
			<head><title>GoServer Fallback</title></head>
			<body style="font-family:sans-serif; background:#1a1a1a; color:#f0f0f0; padding:20px;">
				<h1>GoServer System Message</h1>
				<p>Status Code: %d</p>
				<hr/>
				<p><strong>Note to Developer:</strong> The requested template was not found or failed to parse.</p>
				<p>Please check your <code>web/templates/</code> directory.</p>
				<pre style="background:#333; padding:10px; border-radius:4px;">%v</pre>
			</body>
		</html>`, Status, Data)
}

func (GoServer *GoServer) renderHardcodedErrorHTML(ResponseWriter http.ResponseWriter, Data GenericTemplateData, Status int) {
	ResponseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
	ResponseWriter.WriteHeader(Status)

	fmt.Fprintf(ResponseWriter, `
		<html>
			<head><title>%s</title></head>
			<body style="font-family:sans-serif; background:#1a1a1a; color:#f0f0f0; padding:20px;">
				<h1>%s</h1>
				<p>%s</p>
				<hr/>
				<p><strong>Developer note:</strong> The internal error template could not be rendered.</p>
				<p>Check <code>web/templates/servererror.html</code>.</p>
			</body>
		</html>`,
		Data.PageTitle,
		Data.MainHeading,
		Data.ErrorMessage,
	)
}
