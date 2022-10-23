package server

import (
	"embed"
	"fmt"
	"net/http"
	"text/template"
)

//go:embed template/index.html
var templates embed.FS

// IndexHandler handles the index page
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Get the index template
	index, err := templates.ReadFile("template/index.html")
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("index").Parse(string(index))
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the template
	tmpl.Execute(w, tmpl)
}

func ExportHandler(w http.ResponseWriter, r *http.Request) {
	// Not implemented yet
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}
