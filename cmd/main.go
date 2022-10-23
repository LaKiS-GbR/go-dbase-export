package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Valentin-Kaiser/go-dbase-export/pkg/extract"
	"github.com/Valentin-Kaiser/go-dbase-export/pkg/serialize"
	"github.com/Valentin-Kaiser/go-dbase-export/pkg/server"
	"github.com/Valentin-Kaiser/go-dbase/dbase"
)

const exportPath = "./export/"

func main() {
	start := time.Now()
	// Flags
	run := flag.Bool("run", false, "Run the export in cli")
	path := flag.String("path", "", "Path to the FoxPro/dBase database  file (DATABASE.DBC)")
	export := flag.String("export", exportPath, "Path to the export folder")
	format := flag.String("format", "json", "Format type of the export (json, yaml/yml, toml, csv, xlsx)")
	debugScreen := flag.Bool("debug-screen", false, "Log debug information to the screen")
	debugFile := flag.String("debug-file", "", "Path to the debug file")
	port := flag.Int("port", 80, "Port to start the server on")

	flag.Parse()

	if !*run {
		// Start the server
		server.Start(*port)
		return
	}

	if len(strings.TrimSpace(*path)) == 0 {
		log.Fatal("Please provide a path to the database file")
	}

	if len(strings.TrimSpace(*export)) == 0 {
		log.Fatal("Please provide a path to the export folder")
	}

	if len(strings.TrimSpace(*format)) == 0 {
		log.Fatal("Please provide a format type")
	}

	// Check if format is supported
	if !serialize.IsFormatSupported(*format) {
		log.Fatalf("Format type %v is not supported", *format)
	}

	if *debugScreen && len(strings.TrimSpace(*debugFile)) > 0 {
		// Open debug log file so we see what's going on in the dbase package
		f, err := os.OpenFile(filepath.Clean(*debugFile), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		dbase.Debug(true, io.MultiWriter(os.Stdout, f))
	} else if *debugScreen && len(strings.TrimSpace(*debugFile)) == 0 {
		dbase.Debug(true, os.Stdout)
	} else if !*debugScreen && len(strings.TrimSpace(*debugFile)) > 0 {
		dbase.Debug(true, os.Stdout)
	}

	// Create the export folder if it doesn't exist
	if _, err := os.Stat(*export); os.IsNotExist(err) {
		err := os.Mkdir(*export, 0755)
		if err != nil {
			log.Fatalf("Creating export folder failed with error: %v", err)
		}
	}

	dbSchema, err := extract.Extract(*path)
	if err != nil {
		log.Fatalf("Data extraction failed with error: %v", err)
	}

	// Serialize the schema
	serialize.SerializeSchema(dbSchema, *export, *format)

	elapsed := time.Since(start)
	fmt.Println()
	log.Printf("Export finished in %s", elapsed)
}
