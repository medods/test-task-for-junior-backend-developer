// Package docs serves the embedded OpenAPI spec and Swagger UI.
package docs

import (
	"embed"
	"net/http"
)

//go:embed openapi.json
var openAPISpec embed.FS

// Handler serves OpenAPI spec and Swagger UI assets.
type Handler struct {
	spec []byte
}

// NewHandler loads the embedded OpenAPI spec.
func NewHandler() *Handler {
	spec, err := openAPISpec.ReadFile("openapi.json")
	if err != nil {
		panic(err)
	}

	return &Handler{spec: spec}
}

// ServeSpec serves the OpenAPI JSON document.
func (h *Handler) ServeSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(h.spec)
}

// ServeUI serves the Swagger UI HTML page.
func (h *Handler) ServeUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(swaggerUIHTML))
}

// RedirectToUI redirects /swagger to /swagger/.
func (h *Handler) RedirectToUI(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
}

var swaggerUIHTML = []byte(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Task Service Swagger</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    body { margin: 0; background: #f6f7fb; }
    #swagger-ui { max-width: 1440px; margin: 0 auto; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: '/swagger/openapi.json',
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis],
        layout: 'BaseLayout'
      });
    };
  </script>
</body>
</html>`)
