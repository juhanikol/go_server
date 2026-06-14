package httpserver

// GoServerError holds the details for displaying a user-friendly error page.
type GoServerError struct {
	StatusCode   int
	Title        string
	Message      string
	TechnicalErr string // Only logged, never shown to user for security
}
