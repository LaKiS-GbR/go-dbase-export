package server

import (
	"bytes"
	"embed"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/LaKiS-GbR/go-dbase-export/pkg/config"
	"github.com/LaKiS-GbR/go-dbase-export/pkg/job"
)

var RepositoryName string

//go:embed template/index.html
var templates embed.FS

// Only one job can run at a time
var runningJob *job.Job

// status is the data structure for the index template
type status struct {
	Filename   string
	Exported   bool
	Running    bool
	Error      error
	Time       time.Time
	Duration   time.Duration
	Repository []string
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

	if runningJob == nil {
		err := tmpl.Execute(w, status{Running: false, Filename: config.GetConfig().DBPath})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Append export repository if an job has run successfully
	var repository []string
	if runningJob != nil && runningJob.IsFinished() && runningJob.GetError() == nil {
		files, err := os.ReadDir(config.GetConfig().ExportPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, file := range files {
			repository = append(repository, file.Name())
		}
	}

	// Render the template
	err = tmpl.Execute(w, status{
		Running:    !runningJob.IsFinished(),
		Error:      runningJob.GetError(),
		Exported:   runningJob != nil && runningJob.IsFinished() && runningJob.GetError() == nil,
		Repository: repository,
		Time:       runningJob.Time,
		Duration:   runningJob.Elapsed,
		Filename:   config.GetConfig().DBPath,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ExportHandler(w http.ResponseWriter, r *http.Request) {
	if runningJob != nil && !runningJob.IsFinished() {
		http.Error(w, "A job is already running", http.StatusInternalServerError)
		return
	}

	// Get the format from the url arg
	format := r.URL.Query().Get("format")
	if len(format) == 0 {
		http.Error(w, "Please provide a format", http.StatusInternalServerError)
		return
	}

	// Clean the repository
	if err := os.RemoveAll(config.GetConfig().ExportPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	runningJob = job.New(bytes.NewBuffer(nil), nil)
	go runningJob.Run(
		config.GetConfig().DBPath,
		config.GetConfig().ExportPath,
		format,
	)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	// Get the file name from the url arg
	fileName := r.URL.Query().Get("file")
	if len(fileName) == 0 {
		http.Error(w, "No file name provided", http.StatusBadRequest)
		return
	}

	path := filepath.Join(config.GetConfig().ExportPath, fileName)
	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.Error(w, "File does not exist", http.StatusNotFound)
		return
	}

	// serve the file
	http.Header.Add(w.Header(), "Content-Disposition", "attachment; filename="+fileName)
	http.ServeFile(w, r, path)
}
