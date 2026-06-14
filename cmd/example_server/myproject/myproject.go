// Package myproject presents a demonstration of the possibilites of go_server
package myproject

import (
	"go_server/pkg/httpserver"
	"net/http"
)

// GetManifest provides the configuration for the smoketest.
func GetManifest() httpserver.ProjectManifest {
	return httpserver.ProjectManifest{
		TemplateDir: "web/templates",
		StaticDir:   "web/static",
		RouteMap: map[string]any{
			// 1. Zero-Boilerplate: If about.html exists, it maps to /about
			"/about": "about.html",

			// 2. No-Handler Logic: Pure data, template name inferred as "contact.html"
			"/contact": func() httpserver.GenericTemplateData {
				return httpserver.GenericTemplateData{
					PageTitle:   "Contact Us",
					MainHeading: "Get in Touch",
					IsSuccess:   true,
				}
			},

			// 3. Advanced Logic: Access the request for data, but return a Template Override
			"/user": func(r *http.Request) any {
				user := r.URL.Query().Get("name")
				if user == "" {
					return httpserver.GoServerError{
						StatusCode:   http.StatusBadRequest,
						Title:        "Invalid Request",
						Message:      "You must provide a name parameter (e.g., /user?name=Alpha)",
						TechnicalErr: "missing_query_param",
					}
				}

				return httpserver.TemplateResponse{
					Name: "serverindex.html", // Explicitly use a different template
					Data: httpserver.GenericTemplateData{
						PageTitle:    "Profile",
						MainHeading:  "User Dashboard",
						DisplayValue: "Welcome back, " + user,
						IsSuccess:    true,
					},
				}
			},
		},
	}
}
