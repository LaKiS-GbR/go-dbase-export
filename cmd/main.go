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
	"github.com/Valentin-Kaiser/go-dbase/dbase"
	"github.com/schollz/progressbar/v3"
)

const exportPath = "./export/"

func main() {
	start := time.Now()
	// Flags
	path := flag.String("path", "", "Path to the database file")
	export := flag.String("export", exportPath, "Path to the export folder")
	format := flag.String("format", "json", "Format type of the export (json, yaml/yml, toml, csv, xlsx)")
	debugScreen := flag.Bool("debug-screen", false, "Log debug information to the screen")
	debugFile := flag.String("debug-file", "", "Path to the debug file")

	flag.Parse()

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

	dbSchema, err := extract.Extract(*path, *export, *format)
	if err != nil {
		log.Fatalf("Data extraction failed with error: %v", err)
	}

	serializationBar := progressbar.NewOptions(
		len(dbSchema.TableReferences)+1,
		progressbar.OptionShowCount(),
		// progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetDescription(fmt.Sprintf("%-32.32s %10.10s", "Saving files ", fmt.Sprintf("(%d/%d)", 0, len(dbSchema.TableReferences)+1))),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetItsString("records"),
	)

	fmt.Println()
	err = serializationBar.RenderBlank()
	if err != nil {
		log.Fatalf("Rendering progress bar failed with error: %v", err)
	}
	for i, table := range dbSchema.TableReferences {
		serializationBar.Describe(fmt.Sprintf("%-20.20s %-10.10s %10.10s", "Serialising table...", table.Name, fmt.Sprintf("(%d/%d)", i+1, len(dbSchema.TableReferences)+1)))
		data, err := serialize.Serialize(table, *export, *format)
		if err != nil {
			log.Fatalf("Table serialization failed with error: %v", err)
		}

		serializationBar.Describe(fmt.Sprintf("%-20.20s %-10.10s %10.10s", "Saving table to file", table.Name, fmt.Sprintf("(%d/%d)", i+1, len(dbSchema.TableReferences)+1)))
		path, err := serialize.GetPath(table, *export, *format)
		if err != nil {
			log.Fatalf("Getting path failed with error: %v", err)
		}

		err = serialize.SaveFile(path, data)
		if err != nil {
			log.Fatalf("Saving file failed with error: %v", err)
		}

		err = serializationBar.Add(1)
		if err != nil {
			log.Fatalf("Incrementing progress bar failed with error: %v", err)
		}
	}

	serializationBar.Describe(fmt.Sprintf("%-20.20s %-10.10s %10.10s", "Serializing database", dbSchema.Name, fmt.Sprintf("(%d/%d)", len(dbSchema.TableReferences)+1, len(dbSchema.TableReferences)+1)))
	data, err := serialize.Serialize(dbSchema, *export, *format)
	if err != nil {
		log.Fatalf("Database serialization failed with error: %v", err)
	}

	serializationBar.Describe(fmt.Sprintf("%-20.20s %-10.10s %10.10s", "Saving database", dbSchema.Name, fmt.Sprintf("(%d/%d)", len(dbSchema.TableReferences)+1, len(dbSchema.TableReferences)+1)))
	dbExportPath, err := serialize.GetPath(dbSchema, *export, *format)
	if err != nil {
		log.Fatalf("Getting path failed with error: %v", err)
	}

	err = serialize.SaveFile(dbExportPath, data)
	if err != nil {
		log.Fatalf("Saving file failed with error: %v", err)
	}
	err = serializationBar.Add(1)
	if err != nil {
		log.Fatalf("Incrementing progress bar failed with error: %v", err)
	}

	elapsed := time.Since(start)
	fmt.Println()
	log.Printf("Export finished in %s", elapsed)
}
