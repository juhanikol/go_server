# GoServer: Modular & Hardened Server Framework

GoServer is a production-ready, "library" style framework designed to be imported into other projects. It follows a "Plug-in" architecture, allowing you to add routes, logic, and templates from external modules without modifying the core server files.

## 1. General Features

* Zero-Touch Modularity: Add functionality via the GoServerRouteRegisterer interface.
* Hardened Security: Built-in protection against crashes (Panic Recovery) and HTTP verb enforcement.
* Tiered Rendering: Automatic fallback system that searches for Project templates, then Server templates, and finally a hardcoded safety UI.
* Structured Diagnostics: Integrated JSON/Text logging that captures status codes and response sizes via a custom ResponseWriter wrapper.
* Clean Shutdown: Supports OS signal interception for graceful termination without data loss.

## 2. Technical Architecture & Reasoning

##### The "Inwards-to-Out" Middleware Chain

The server uses a functional decorator pattern. Every request is wrapped in layers that provide features without the developer needing to write them manually:

* Recovery Layer: Traps panics to prevent server crashes.
* Logging Layer: Records request metadata using the GoServerResponseWriter.
* Method Layer: Ensures only allowed HTTP verbs (GET, POST, etc.) reach your logic.
* Logic Layer: Your specific business code executes last.

##### The Response Wrapper (GoServerResponseWriter)

Standard Go http.ResponseWriter is "blind" to what was sent. Our wrapper intercepts WriteHeader and Write calls to capture the Status Code and Bytes Written, allowing the logger to report exactly what the user received.

## 3. Core Methods & Functions

##### Server Management (server.go)

* NewGoServer(addr string, logger *logging.Logger): The constructor that initializes the router and hardened HTTP settings.
* Start(): Begins the blocking listen-and-serve process with middleware injection.
* Stop(ctx context.Context): Gracefully closes all active connections.

##### Rendering & Errors (routes.go & errors.go)

* RenderGoServerTemplate(w, name, data, status): The tiered renderer. It looks for name.html (Project), then servername.html (Internal), then falls back to renderFallbackHTML.
* RenderGoServerError(w, GoServerError): The unified error handler. It logs technical details for you while showing a friendly message to the user.

## 4. Usage & "How To" Guide

**Importing as a Module**
To use this in a new project, initialize your project and point to the go_server location:

Bash

```
go get "gitea.com/this address will be updated"
```

**Implementing a Plugin**
Create a struct that satisfies the GoServerRouteRegisterer interface to inject your own routes.

Go

```
type MyModule struct {}

func (m *MyModule) RegisterGoServerRoutes(gs *httpserver.GoServer) {
   // Add a protected GET route
   gs.GoServerHandler("/my-page", httpserver.MethodMiddleware("GET", m.HandlePage()))
}

func (m *MyModule) HandlePage() http.HandlerFunc {
   return func(w http.ResponseWriter, r *http.Request) {
   fmt.Fprint(w, "Hello from the plugin!")
}
}
```

**Starting the Server**
Use the serverapp convenience package or manual initialization in your main.go:

Go

```
func main() {
    srv := httpserver.NewGoServer(":8080", myLogger)
  
    // Plug in your modules
    myPlugin := &MyModule{}
    myPlugin.RegisterGoServerRoutes(srv)
  
    // Run with graceful shutdown support
    srv.Start()
}
```

## 5. Practical Structs & Instructions

To ensure no errors, always populate these structs when interacting with the server:

* GenericTemplateData: Use for successful page renders. Ensure IsSuccess is true to trigger correct UI states in templates.
* GoServerError: Use for failures.
  * TechnicalErr: Fill this with err.Error() for your logs.
  * Message: Fill this with a user-friendly explanation (e.g., "Item not found").

#### GenericTemplateData

This struct is the standard bridge between your Go logic and the HTML templates. It is used for all "Happy Path" scenarios where a page is rendered successfully.

**Struct Definition and Property Meanings:**

* **`PageTitle`** : The string that appears in the browser tab (usually injected into `<title>`).
* **`MainHeading`** : The primary `<h1>` or title displayed on the page content.
* **`DisplayValue`** : A generic field for additional instructional text or primary content strings.
* **`ErrorMessage`** : Should be left empty for successful renders; if populated, it can trigger alert banners in the UI.
* **`IsSuccess`** : A boolean flag. Set to `true` for normal pages to ensure the UI doesn't trigger error-state styling.

Example Usage in `routes.go`:

```
func (Module *MyModule) HandleDashboard(GoServer *httpserver.GoServer) http.HandlerFunc {
    return func(ResponseWriter http.ResponseWriter, Request *http.Request) {
        // Populating the struct for a successful page
        Data := httpserver.GenericTemplateData{
            PageTitle:    "Admin Dashboard",
            MainHeading:  "System Overview",
            DisplayValue: "Welcome to the project-specific dashboard.",
            IsSuccess:    true, // Correctly signals a successful state
        }

        // Affects: RenderGoServerTemplate checks if "dashboard.html" exists in web/templates/
        GoServer.RenderGoServerTemplate(ResponseWriter, "dashboard.html", Data, http.StatusOK)
    }
}
```

## 6. Useful Knowledge

* Template Priority: If you want to override the default error page, simply create web/templates/index.html in your project. The server will detect it and ignore its internal serverindex.html.
* JSON Logging: Logs are stored in myserver.log by default. Use a JSON viewer or grep to filter by "component" (e.g., grep "GoServer-Recovery" myserver.log).
* Panic Safety: If your code triggers a "nil pointer" error, the server will not crash. It will log the stack trace and show the servererror.html page automatically.
