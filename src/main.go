package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Valentin-Kaiser/go-dbase/dbase"
	"github.com/schollz/progressbar/v3"
)

const exportPath = "./export/"

type DatabaseSchema struct {
	Name      string
	Tables    map[string]string
	Generated time.Duration
}

type Table struct {
	Name      string
	Columns   uint16
	Records   uint32
	FirstRow  int
	RowLength string
	FileSize  int64
	Modified  time.Time
	Fields    map[string]Field
	Data      []map[string]interface{}
}

type Field struct {
	Name   string
	Type   string
	GoType string
	Length int
}

func main() {
	// Flags
	path := flag.String("path", "", "Path to the database file")
	exportPath := flag.String("export", exportPath, "Path to the export folder")
	debug := flag.Bool("debug", false, "Log debug information")
	debugScreen := flag.Bool("debug-screen", false, "Log debug information to the screen")
	debugFile := flag.String("debug-file", "debug.log", "Path to the debug file")

	flag.Parse()

	if len(strings.TrimSpace(*path)) == 0 {
		fmt.Println("Please provide a path to the database file")
		os.Exit(1)
	}

	if len(strings.TrimSpace(*exportPath)) == 0 {
		fmt.Println("Please provide a path to the export folder")
		os.Exit(1)
	}

	// Create the export folder if it doesn't exist
	if _, err := os.Stat(*exportPath); os.IsNotExist(err) {
		err := os.Mkdir(*exportPath, 0755)
		if err != nil {
			fmt.Println("Could not create the export folder")
			os.Exit(1)
		}
	}

	if *debug {
		// Open debug log file so we see what's going on
		f, err := os.OpenFile(filepath.Clean(*debugFile), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}
		if !*debugScreen {
			dbase.Debug(true, f)
		} else {
			dbase.Debug(true, os.Stdout)
		}
	}

	start := time.Now()
	db, err := dbase.OpenDatabase(&dbase.Config{
		Filename:   *path,
		TrimSpaces: true,
	})
	if err != nil {
		panic(dbase.GetErrorTrace(err))
	}
	defer db.Close()

	schema := db.Schema()
	tables := db.Tables()

	// length := len(tables)
	databaseSchema := DatabaseSchema{
		Name:   strings.Trim(filepath.Base(*path), filepath.Ext(*path)),
		Tables: make(map[string]string),
	}

	keys := make([]string, 0)
	for table := range schema {
		keys = append(keys, table)
	}
	sort.Strings(keys)

	total := 0
	progresses := make(map[string]*progressbar.ProgressBar)
	for i, name := range keys {
		total += int(tables[name].Header().RecordsCount())
		progresses[name] = progressbar.NewOptions(
			int(tables[name].Header().RecordsCount()),
			progressbar.OptionShowCount(),
			// progressbar.OptionShowIts(),
			progressbar.OptionSetPredictTime(false),
			progressbar.OptionSetWidth(30),
			progressbar.OptionSetDescription(fmt.Sprintf("Exporting %10.10s  %10.10s", fmt.Sprintf("(%v/%v)", i+1, len(tables)), strings.ToUpper(name))),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionSetItsString("records"),
		)
	}

	for _, tablename := range keys {
		fmt.Print("\n")

		progresses[tablename].RenderBlank()
		t := Table{
			Name:     strings.ToUpper(tablename),
			Columns:  tables[tablename].Header().ColumnsCount(),
			Records:  tables[tablename].Header().RecordsCount(),
			FileSize: tables[tablename].Header().FileSize(),
			Modified: tables[tablename].Header().Modified(),
			Fields:   make(map[string]Field),
			Data:     make([]map[string]interface{}, 0),
		}

		for _, field := range schema[tablename] {
			t.Fields[field.Name()] = Field{
				Name:   field.Name(),
				Type:   field.Type(),
				GoType: field.Reflect().String(),
				Length: int(field.Length),
			}
		}

		if t.Records == 0 {
			progresses[tablename].Finish()
			continue
		}

		// var wg sync.WaitGroup
		for !tables[tablename].EOF() {
			// wg.Add(1)
			// go func(wg *sync.WaitGroup, progress *progressbar.ProgressBar, t *Table) {
			// 	defer wg.Done()
			// This reads the complete row
			progresses[tablename].Add(1)
			row, err := tables[tablename].Row()
			if err != nil {
				// Skip faulty data
				tables[tablename].Skip(1)
				continue
			}

			// Increment the row pointer
			// skip deleted rows
			if row.Deleted {
				tables[tablename].Skip(1)
				continue
			}

			m, err := row.ToMap()
			if err != nil {
				tables[tablename].Skip(1)
				continue
			}

			t.Data = append(t.Data, m)
			tables[tablename].Skip(1)
			// }(&wg, progresses[tablename], &t)
		}
		// wg.Wait()

		// Write table to file
		b, err := json.MarshalIndent(t, "", "  ")
		if err != nil {
			panic(err)
		}
		err = os.WriteFile(filepath.Join(*exportPath, t.Name+".json"), b, 0644)
		if err != nil {
			panic(err)
		}

		databaseSchema.Tables[strings.ToUpper(tablename)] = t.Name
		// fmt.Printf("Export %v/%v table completed \n", it+1, len(tables))
	}
	duration := time.Since(start)
	databaseSchema.Generated = duration

	// JSON encoding
	b, err := json.MarshalIndent(databaseSchema, "", "    ")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Open schema output file
	schemaFile, err := os.OpenFile(filepath.Join(*exportPath, databaseSchema.Name+".json"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}

	_, err = schemaFile.Write(b)
	if err != nil {
		panic(err)
	}
}

// ToByteString returns the number of bytes as a string with a unit
func ToByteString(b int) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
