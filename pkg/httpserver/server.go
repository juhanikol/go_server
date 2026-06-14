// Package httpserver contains all the necessary methods to launch a Go-server
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

// SimplePageLogic is the ultimate simplification: responsewriter and requests.
// The developer only provides the data to be displayed.
type SimplePageLogic func() GenericTemplateData

// AdvancedPageLogic allows reading the request (e.g., for forms)
// but still hides some of the complexity.
type AdvancedPageLogic func(Request *http.Request) any

// TemplateResponse allows the developer to explicitly name a template
// and provide data, breaking the strict path-to-filename dependency.
type TemplateResponse struct {
	Name string
	Data any
}

// PageDefinition is configuration for a single route.
type PageDefinition struct {
	TemplateName string
	Method       string // e.g., "POST"
	Logic        any    // simplepagelogic, advancedpagelogic, nil...
}

// ProjectManifest defines the automated configuration for a project.
type ProjectManifest struct {
	StaticDir   string
	TemplateDir string
	RouteMap    map[string]any // Mapping of paths to logic, templates, or PageDefinitions
}

// ServerConfig has basic settings for the HTTP server.
type ServerConfig struct {
	ServerAddress      string
	ServerReadTimeout  time.Duration
	ServerWriteTimeout time.Duration
	AllowedHosts       []string
}

// GoServer is the primary orchestrator for the web server.
type GoServer struct {
	GoServerLogger      *slog.Logger
	GoServerRouter      *http.ServeMux
	GoServerServing     *http.Server
	GoServerHomeHandler http.Handler
	DomainMap           map[string]ProjectManifest
	TemplateCache       map[string]*template.Template
	Manifest            ProjectManifest
	// RegisteredPaths tracks all routes to prevent auto-router collisions.
	RegisteredPaths map[string]bool
	AllowedHosts    []string
	Env             string
}

// NewGoServer initializes the GoServer struct.
func NewGoServer(ServerConfig ServerConfig, GoServerLogger *slog.Logger) *GoServer {
	return &GoServer{
		GoServerLogger:  GoServerLogger,
		GoServerRouter:  http.NewServeMux(),
		TemplateCache:   make(map[string]*template.Template),
		RegisteredPaths: make(map[string]bool),
		AllowedHosts:    append([]string(nil), ServerConfig.AllowedHosts...),
		GoServerServing: &http.Server{
			Addr:         ServerConfig.ServerAddress,
			ReadTimeout:  ServerConfig.ServerReadTimeout,
			WriteTimeout: ServerConfig.ServerWriteTimeout,
		},
	}
}

// Launch is the primary entry point. It triggers the scanner and registers all routes.
func (GoServer *GoServer) Launch(Manifest ProjectManifest) error {
	GoServer.Manifest = Manifest

	// 1. Run the Scanner to populate TemplateCache and register implicit routes.
	GoServer.ScanProjectResources()

	// 2. Interpret the RouteMap from the Manifest.
	GoServer.InterpretManifest()

	// 3. Add Infrastructure Routes (Static files, Health check, and Fallbacks).
	GoServer.AddDefaultGoServerRoutes()

	return GoServer.Start()
}

// BuildRouteHandler creates the standard middleware chain for normal routes.
// Order:
// 1. Business logic
// 2. Method check
// 3. Request/response logging
//
// Panic recovery is intentionally applied globally in Start() around the full router.
func (GoServer *GoServer) BuildRouteHandler(AllowedMethod string, LogicHandler http.HandlerFunc) http.Handler {
	var FinalHandler http.Handler = http.HandlerFunc(LogicHandler)

	if AllowedMethod != "" {
		FinalHandler = MethodMiddleware(AllowedMethod, FinalHandler)
	}

	return FinalHandler
}

// RegisterRoute is the main constructor-style entry point for normal routes.
// Developers should prefer this instead of manually composing middleware.
func (GoServer *GoServer) RegisterRoute(Path string, AllowedMethod string, LogicHandler http.HandlerFunc) {
	// Conflict Resolution: Track that this path is manually claimed.
	GoServer.RegisteredPaths[Path] = true

	WrappedHandler := GoServer.BuildRouteHandler(AllowedMethod, LogicHandler)
	GoServer.GoServerRouter.Handle(Path, WrappedHandler)
}

// InterpretManifest processes the RouteMap and converts it into hardened handlers.
// It understands the new SimplePageLogic type
func (GoServer *GoServer) InterpretManifest() {
	for Path, Value := range GoServer.Manifest.RouteMap {
		var Def PageDefinition

		switch v := Value.(type) {
		case string:
			Def = PageDefinition{TemplateName: v, Method: http.MethodGet}
		case func() GenericTemplateData:
			Def = PageDefinition{Logic: v, Method: http.MethodGet, TemplateName: Path[1:] + ".html"}
		case func(*http.Request) any:
			Def = PageDefinition{Logic: v, Method: http.MethodGet, TemplateName: Path[1:] + ".html"}
		case PageDefinition:
			Def = v
		default:
			GoServer.GoServerLogger.Warn("Interpreter: Unknown logic type for route", "path", Path)
			continue
		}
		GoServer.RegisterSimplePage(Path, Def)
	}
}

// RegisterSimplePage wraps a PageDefinition in the standard middleware chain.
func (GoServer *GoServer) RegisterSimplePage(Path string, Def PageDefinition) {
	if Def.Method == "" {
		Def.Method = http.MethodGet
	}

	// This is the core Interpreter Pattern:
	// It converts Simple or Advanced logic into a standard http.HandlerFunc.
	InterpreterBridge := func(w http.ResponseWriter, r *http.Request) {
		var rawData any

		// 1. Execute the logic
		switch logic := Def.Logic.(type) {
		case func() GenericTemplateData:
			rawData = logic()
		case func(*http.Request) any:
			rawData = logic(r)
		default:
			rawData = nil
		}

		// 2. Determine Template and Final Data
		templateToUse := Def.TemplateName // Default from Manifest or Path
		finalData := rawData

		// Check if the developer provided an explicit Template Override
		if override, ok := rawData.(TemplateResponse); ok {
			templateToUse = override.Name
			finalData = override.Data
		}

		// 3. Handle Errors
		if errData, ok := finalData.(GoServerError); ok {
			GoServer.RenderGoServerError(w, errData)
			return
		}

		// 4. Render
		GoServer.RenderGoServerTemplate(w, templateToUse, finalData, http.StatusOK)
	}

	GoServer.RegisterRoute(Path, Def.Method, InterpreterBridge)
}

// GoServerHandler is kept as a compatibility helper.
// It applies the standard logging middleware for prebuilt handlers.
func (GoServer *GoServer) GoServerHandler(Path string, HTTPHandler http.Handler) {
	GoServer.GoServerRouter.Handle(Path, LoggingMiddleware(HTTPHandler))
}

// SetHomeRoute lets the importing project define its own "/" handler.
// This should be called BEFORE AddDefaultGoServerRoutes().
func (GoServer *GoServer) SetHomeRoute(HomeHandler http.HandlerFunc) {
	if HomeHandler == nil {
		return
	}
	// Homepage is registered to the tracker as well.
	GoServer.RegisteredPaths["/"] = true
	GoServer.GoServerHomeHandler = GoServer.BuildRouteHandler(http.MethodGet, HomeHandler)
}

// SetHomeTemplate is a convenience helper for simple homepage cases.
// Use this when the project only wants to supply a template + data,
// without writing a full custom handler.
func (GoServer *GoServer) SetHomeTemplate(TemplateName string, TemplateData any, Status int) {
	GoServer.SetHomeRoute(func(ResponseWriter http.ResponseWriter, Request *http.Request) {
		if Request.URL.Path != "/" {
			http.NotFound(ResponseWriter, Request)
			return
		}

		GoServer.RenderGoServerHome(ResponseWriter, TemplateName, TemplateData, Status)
	})
}

// Start attaches the router and begins listening for requests.
func (GoServer *GoServer) Start() error {
	GoServer.GoServerServing.Handler = GoServer.activeHandler()

	GoServer.GoServerLogger.Info("GoServer is now online", "addr", GoServer.GoServerServing.Addr)

	err := GoServer.GoServerServing.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (GoServer *GoServer) activeHandler() http.Handler {
	// TODO: Restore multi-domain routing here when DomainMap has a complete
	// registration and dispatch model. The current single-project flow must
	// serve the registered router directly.
	handler := GoServer.hostPolicyMiddleware(GoServer.GoServerRouter)
	handler = RecoveryMiddleware(GoServer, handler)
	handler = LoggingMiddleware(handler)
	return handler
}

func (GoServer *GoServer) hostPolicyMiddleware(nextHandler http.Handler) http.Handler {
	allowedHosts := GoServer.normalizedAllowedHosts()
	if len(allowedHosts) == 0 {
		return nextHandler
	}

	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		requestHost := normalizeHostForPolicy(request.Host)
		if allowedHosts[requestHost] {
			nextHandler.ServeHTTP(responseWriter, request)
			return
		}

		GoServer.GoServerLogger.Warn("Rejected request host", "host", request.Host, "normalized_host", requestHost)
		GoServer.RenderGoServerError(responseWriter, GoServerError{
			StatusCode:   http.StatusForbidden,
			Title:        "Host Not Allowed",
			Message:      "The requested host is not allowed by this server.",
			TechnicalErr: fmt.Sprintf("rejected host %q", request.Host),
		})
	})
}

func (GoServer *GoServer) normalizedAllowedHosts() map[string]bool {
	if len(GoServer.AllowedHosts) == 0 {
		return nil
	}

	allowedHosts := make(map[string]bool, len(GoServer.AllowedHosts))
	for _, host := range GoServer.AllowedHosts {
		normalizedHost := normalizeHostForPolicy(host)
		if normalizedHost != "" {
			allowedHosts[normalizedHost] = true
		}
	}

	return allowedHosts
}

func normalizeHostForPolicy(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return ""
	}

	if splitHost, _, err := net.SplitHostPort(host); err == nil {
		return strings.Trim(splitHost, "[]")
	}

	if strings.HasPrefix(host, "[") {
		if end := strings.Index(host, "]"); end > 0 {
			return host[1:end]
		}
	}

	return strings.Trim(host, "[]")
}

// Shutdown provides a clean way to stop the server from main.go or runserver.go.
func (GoServer *GoServer) Shutdown(ctx context.Context) error {
	GoServer.GoServerLogger.Info("Initiating GoServer shutdown...")
	return GoServer.GoServerServing.Shutdown(ctx)
}

// RenderGoServerError provides a unified way to show error pages.
// It logs the technical error and renders the index or a specific error template.
func (GoServer *GoServer) RenderGoServerError(ResponseWriter http.ResponseWriter, Err GoServerError) {
	// 1. LOG THE ERROR
	// use the internal GoServerLogger to record the technical details.
	// Then developer/devops sees the "TechnicalErr" in the logs,
	// while the user only sees the friendly "Message".
	GoServer.GoServerLogger.Error("HTTP Error Occurred",
		"status", Err.StatusCode,
		"title", Err.Title,
		"user_message", Err.Message,
		"technical_details", Err.TechnicalErr,
	)

	// 2. APPLY SAFE DEFAULTS IF SOME FIELDS ARE EMPTY
	if Err.StatusCode == 0 {
		Err.StatusCode = http.StatusInternalServerError
	}
	if Err.Title == "" {
		Err.Title = "System Error. Error title empty or unclear"
	}
	if Err.Message == "" {
		Err.Message = "The request could not be completed. GoServer switched to its internal error page. Error message empty"
	}

	// 3. PREPARE TEMPLATE DATA
	DisplayData := GenericTemplateData{
		PageTitle:    fmt.Sprintf("Error %d", Err.StatusCode),
		MainHeading:  Err.Title,
		ErrorMessage: Err.Message,
		DisplayValue: "GoServer switched to its internal error page. Check your page code to display page content correctly",
		IsSuccess:    false,
	}

	// 4. TRY DEDICATED INTERNAL ERROR TEMPLATE ONLY
	// This helper does not call RenderGoServerError again.
	if GoServer.renderEmbeddedTemplate(ResponseWriter, "servererror.html", DisplayData, Err.StatusCode) {
		return
	}

	// 5. FINAL FALLBACK
	// We use the existing RenderGoServerTemplate method.
	GoServer.renderHardcodedErrorHTML(ResponseWriter, DisplayData, Err.StatusCode)
}
