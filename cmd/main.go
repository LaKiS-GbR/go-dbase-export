package main

import (
	"flag"
	"io"
	"os"

	"github.com/LaKiS-GbR/go-dbase-export/pkg/job"
	"github.com/LaKiS-GbR/go-dbase-export/pkg/server"
)

const exportPath = "./export/"

func main() {
	// Flags
	run := flag.Bool("run", false, "Run the export in cli")
	path := flag.String("path", "", "Path to the FoxPro/dBase database  file (DATABASE.DBC)")
	export := flag.String("export", exportPath, "Path to the export folder")
	format := flag.String("format", "json", "Format type of the export (json, yaml/yml, toml, csv, xlsx)")
	debugScreen := flag.Bool("debug-screen", false, "Log debug information to the screen")
	debugFile := flag.String("debug-file", "", "Path to the debug file")
	repository := flag.String("repository", "./repository", "Path to the repository folder (Used to store the uploaded files)")

	flag.Parse()

	if !*run {
		server.RepositoryName = *repository
		// Start the server
		server.Start()
		return
	}

	// Debug output
	var debug io.Writer
	if *debugScreen {
		debug = os.Stdout
	} else if len(*debugFile) > 0 {
		f, err := os.OpenFile(*debugFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		debug = f
	}

	job.New(os.Stdout, debug).Run(*path, *export, *format)
}
