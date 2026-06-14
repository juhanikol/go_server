# GO_SERVER

Modular go server to serve the needs for creating server once in a while. Use as private library or as template to your project.

## Features

* Interface to add routes from other projects without sweat
* Custom logger to save all logs to file and also stdout
* save errors and warnings to the logger
* graceful termination (when Ctrl+C pressed)
* Decoupling for the api/other endpoints

## Design and technical features

* **Usage as** : "library"
* **Usage as template** : having structure to conform go-idioms/standards
* **One clear constructor** : Developer should only write business logic and let the server compose the outer concerns.
* **Decoupling** : The server doesn't care *where* data comes from. You could swap the API for a Database without touching `server.go
* **Safety** : Using `RegisterRoute` in your routes prevents the most beginner mistakes (allow a `DELETE` action on a `GET` request`) and header checks
* **Logging** : method, path, remote, duration, final status code, bytes written
* **Fallback/error behaviour** : using premade errorpages for clear error messaging
* **Global recovery** : panic recovery wrapped around the whole route

## Practical functions

Use these as your working rule set:

* `SetHomeRoute(...)` : for custom homepage handler
* `SetHomeTemplate(...)` : for simple homepage without custom logic
* `RegisterRoute(...)` : for normal routes and API endpoints
* `RenderGoServerHome(...)` : inside homepage handler
* `RenderGoServerTemplate(...)` : inside other HTML pages
* `RenderGoServerError(...)` : for explicit 4xx / 5xx pages


## Usage

### Private library

```
go get "gitea.com/this address will be updated"
```

create a struct in your new project that implements the GoServerRouteRegisterer interface in routes.go


### The "Go Workspace" Approach (go.work)

If you have multiple projects locally (e.g., Project_A and Go_Modular_Server) and you don't want to push to GitHub yet, use a Go Workspace.

How: Create a go.work file in your root folder:

```

Go
go 1.25.3
use (
./Go_Modular_Server
./Project_A
)
```

## Examples

#### A normal project route

Developer writes only the logic:

```
func HandleDashboard(ResponseWriter http.ResponseWriter, Request *http.Request) {
	ResponseWriter.Write([]byte("Dashboard page"))
}
```

Then register it

```
GoServerInstance.RegisterRoute("/dashboard", http.MethodGet, HandleDashboard)
```

#### A homepage handler

Homepage can still be set separately, because homepage has special fallback behavior:

```
GoServerInstance.SetHomeRoute(func(ResponseWriter http.ResponseWriter, Request *http.Request) {
	Data := httpserver.GenericTemplateData{
		PageTitle:    "My App Home",
		MainHeading:  "Welcome",
		DisplayValue: "This homepage belongs to the project.",
		IsSuccess:    true,
	}

	GoServerInstance.RenderGoServerHome(ResponseWriter, "index.html", Data, http.StatusOK)
})
```

#### A simple homepage without custom logic

```
GoServerInstance.SetHomeTemplate(
	"index.html",
	httpserver.GenericTemplateData{
		PageTitle:    "My App Home",
		MainHeading:  "Welcome",
		DisplayValue: "Configured with SetHomeTemplate.",
		IsSuccess:    true,
	},
	http.StatusOK,
)
```

That is the shortest way to define a homepage.

#### An API route

```
func HandlePing(ResponseWriter http.ResponseWriter, Request *http.Request) {
	ResponseWriter.Header().Set("Content-Type", "application/json")
	ResponseWriter.WriteHeader(http.StatusOK)
	ResponseWriter.Write([]byte(`{"status":"ok"}`))
}

GoServerInstance.RegisterRoute("/api/ping", http.MethodGet, HandlePing)
```

This keeps API routes project-owned

#### Example of full usage in a module

A simple external module could look like this:

```
package mymodule

import (
	"net/http"

	"go_server/internal/httpserver"
)

type MyModule struct{}

func (Module *MyModule) RegisterGoServerRoutes(GoServer *httpserver.GoServer) {
	GoServer.SetHomeRoute(Module.HandleHome(GoServer))
	GoServer.RegisterRoute("/dashboard", http.MethodGet, Module.HandleDashboard(GoServer))
	GoServer.RegisterRoute("/api/ping", http.MethodGet, Module.HandlePing())
}

func (Module *MyModule) HandleHome(GoServer *httpserver.GoServer) http.HandlerFunc {
	return func(ResponseWriter http.ResponseWriter, Request *http.Request) {
		Data := httpserver.GenericTemplateData{
			PageTitle:    "My Module Home",
			MainHeading:  "Welcome to My Module",
			DisplayValue: "This is the custom homepage from an external module.",
			IsSuccess:    true,
		}

		GoServer.RenderGoServerHome(ResponseWriter, "index.html", Data, http.StatusOK)
	}
}

func (Module *MyModule) HandleDashboard(GoServer *httpserver.GoServer) http.HandlerFunc {
	return func(ResponseWriter http.ResponseWriter, Request *http.Request) {
		Data := httpserver.GenericTemplateData{
			PageTitle:    "Dashboard",
			MainHeading:  "Dashboard",
			DisplayValue: "Project dashboard content.",
			IsSuccess:    true,
		}

		GoServer.RenderGoServerTemplate(ResponseWriter, "dashboard.html", Data, http.StatusOK)
	}
}

func (Module *MyModule) HandlePing() http.HandlerFunc {
	return func(ResponseWriter http.ResponseWriter, Request *http.Request) {
		ResponseWriter.Header().Set("Content-Type", "application/json")
		ResponseWriter.WriteHeader(http.StatusOK)
		ResponseWriter.Write([]byte(`{"status":"ok"}`))
	}
}
```


`GoServerRouteRegisterer` interface. This allows an entirely different project to "plug in" to this modular server.

Imagine you have a new project called **"InventoryModule"** . You don't want to change the `go_server` code; you just want to add inventory routes to it.

File: `external_project/inventory.go`

**Bash**

```
package inventory

import (
	"fmt"
	"net/http"

	// Import your modular server
	"go_server/internal/httpserver"
)

// InventoryPlugin implementing the GoServerRouteRegisterer interface
type InventoryPlugin struct {
	Name string
}

// RegisterGoServerRoutes is the required method for the interface.
// It allows this external module to inject its own logic into GoServer.
func (IP *InventoryPlugin) RegisterGoServerRoutes(GS *httpserver.GoServer) {
	GS.GoServerLogger.Info("Registering external Inventory Plugin", "module", IP.Name)

	// Add a new route using the modular server's handler
	GS.GoServerHandler("/inventory", httpserver.MethodMiddleware(http.MethodGet, IP.HandleInventoryList()))
}

func (IP *InventoryPlugin) HandleInventoryList() http.HandlerFunc {
	return func(RW http.ResponseWriter, R *http.Request) {
		fmt.Fprintf(RW, "Welcome to the %s inventory list!", IP.Name)
	}
}
```

**How to trigger this in `runserver.go`:**
In `runserver.go`--> run(), you would simply add one line after initializing the server:

```
InventoryModule := &inventory.InventoryPlugin{Name: "Warehouse-Alpha"}
InventoryModule.RegisterGoServerRoutes(GoServerInstance)
```

## Project Structure (WIP)

* **main.go** : The entry point for the Go-based Web Server.
* **api/** : Shared API models for entire server.
* **internal/httpserver/** : Implementation of the HTTP server, including routes, handlers, and the `renderTemplate` engine.
* **internal/services/api1/** : The `apicaller.go` service responsible for external API communication.
* **static/** : Contains HTML files (`index.html`, `details.html`, `compare.html`) and CSS styles.

## USAGE

This is the important usage rule:

* for homepage route /: use RenderGoServerHome(...)
* for all other HTML pages: use RenderGoServerTemplate(...)
* for explicit 4xx / 5xx: use RenderGoServerError(...)

That lets the project define pages without changing the server code

## USAGE EXAMPLES

Example homepage in external module

```
func (App *MyModule) HandleHome(GS *httpserver.GoServer) http.HandlerFunc {
	return func(ResponseWriter http.ResponseWriter, Request *http.Request) {
		Data := httpserver.GenericTemplateData{
			PageTitle:    "My App Home",
			MainHeading:  "Welcome",
			DisplayValue: "This page belongs to the project, not the GoServer core.",
			IsSuccess:    true,
		}

		GS.RenderGoServerHome(ResponseWriter, "index.html", Data, http.StatusOK)
	}
}
```

Example other page:

```
func (App *MyModule) HandleDashboard(GS *httpserver.GoServer) http.HandlerFunc {
	return func(ResponseWriter http.ResponseWriter, Request *http.Request) {
		Data := DashboardData{
			Title: "Dashboard",
		}

		GS.RenderGoServerTemplate(ResponseWriter, "dashboard.html", Data, http.StatusOK)
	}
}
```

Example explicit 404:

```
GS.RenderGoServerError(ResponseWriter, httpserver.GoServerError{
	StatusCode:   http.StatusNotFound,
	Title:        "Page Not Found",
	Message:      "The requested resource does not exist.",
	TechnicalErr: "No matching route or record",
})
```

external project module should register something like this:

```
func (App *MyModule) RegisterGoServerRoutes(GS *httpserver.GoServer) {
	GS.GoServerHandler("/", httpserver.MethodMiddleware(http.MethodGet, App.HandleHome(GS)))
	GS.GoServerHandler("/dashboard", httpserver.MethodMiddleware(http.MethodGet, App.HandleDashboard(GS)))
}
```

And the homepage handler should use `RenderGoServerHome(...)`, because homepage failure should fall back to `serverindex.html`, not directly to the generic error page.

## Authors

* **Juhani Kolehmainen** (Juhani.kolehmainen@gmail.com
