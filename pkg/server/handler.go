package server

import (
	"embed"
	"net/http"
	"text/template"

	"github.com/Valentin-Kaiser/go-dbase-export/pkg/config"
	"github.com/Valentin-Kaiser/go-dbase-export/pkg/job"
)

//go:embed template/index.html
var templates embed.FS

// status is the data structure for the index template
type status struct {
	Filename string
	Jobs     []job.Job
}

// IndexHandler handles the index page
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Get the index template
	index, err := templates.ReadFile("template/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("index").Parse(string(index))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the template
	err = tmpl.Execute(w, status{
		Filename: config.GetConfig().DBPath,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {

	// Get the format from form data

	// Get the files from form data

	// Read table names from the files

	// Present the user with a form to select the tables to export

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ExportHandler(w http.ResponseWriter, r *http.Request) {
	// Get the tables to export from the form data

	// Start a new job

	// Add the job to the job list
}

func JobHandler(w http.ResponseWriter, r *http.Request) {
	// Get the job id from the url

	// Get the job from the job list

	// Show all job information and the log
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	// Get the job id from the url

	// Get the job from the job list

	// Check if the job is finished

	// Check if the job has an error

	// Get the export file path

	// Open the file

	// Set the content type

	// Set the content disposition

	// Copy the file to the response writer
}
