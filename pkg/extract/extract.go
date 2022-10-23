package extract

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Valentin-Kaiser/go-dbase-export/pkg/model"
	"github.com/Valentin-Kaiser/go-dbase/dbase"
	"github.com/schollz/progressbar/v3"
)

func Extract(path string) (*model.DatabaseSchema, error) {
	defer func() {
		// Add spacing to the end of the progress bar to make it look nicer
		fmt.Printf("\n")
	}()

	start := time.Now()
	db, err := dbase.OpenDatabase(&dbase.Config{
		Filename:   path,
		TrimSpaces: true,
	})
	if err != nil {
		panic(dbase.GetErrorTrace(err))
	}
	defer db.Close()

	schema := db.Schema()
	tables := db.Tables()

	// length := len(tables)
	databaseSchema := &model.DatabaseSchema{
		Name:   strings.Trim(filepath.Base(path), filepath.Ext(path)),
		Tables: make(map[string]string),
	}

	keys := make([]string, 0)
	for table := range schema {
		keys = append(keys, table)
	}
	sort.Strings(keys)

	progresses := make(map[string]*progressbar.ProgressBar)
	for i, name := range keys {
		progresses[name] = progressbar.NewOptions(
			int(tables[name].Header().RecordsCount()),
			progressbar.OptionShowCount(),
			// progressbar.OptionShowIts(),
			progressbar.OptionSetPredictTime(false),
			progressbar.OptionSetWidth(30),
			progressbar.OptionSetDescription(fmt.Sprintf("Extracting data from %-10.10s %10.10s", strings.ToUpper(name), fmt.Sprintf("(%v/%v)", i+1, len(tables)))),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionSetItsString("records"),
		)
	}

	for _, tablename := range keys {
		fmt.Println() // new line for progress bar to display correctly one per line
		err := progresses[tablename].RenderBlank()
		if err != nil {
			log.Printf("Error rendering blank progress bar: %v", err)
		}
		table := model.Table{
			Name:      strings.ToUpper(tablename),
			Columns:   tables[tablename].Header().ColumnsCount(),
			Records:   tables[tablename].Header().RecordsCount(),
			FileSize:  tables[tablename].Header().FileSize(),
			Modified:  tables[tablename].Header().Modified(),
			RowLength: tables[tablename].Header().RowLength,
			FirstRow:  int(tables[tablename].Header().FirstRow),
			Fields:    make(map[string]*model.Field),
			Data:      make([]map[string]interface{}, 0),
		}

		for _, field := range schema[tablename] {
			table.Fields[field.Name()] = &model.Field{
				Name:   field.Name(),
				Type:   field.Type(),
				GoType: field.Reflect().String(),
				Length: int(field.Length),
			}
		}

		if table.Records == 0 {
			err = progresses[tablename].Finish()
			if err != nil {
				log.Printf("Error finishing progress bar: %v", err)
			}
			continue
		}

		for !tables[tablename].EOF() {
			err = progresses[tablename].Add(1)
			if err != nil {
				log.Printf("Error incrementing progress bar: %v", err)
			}
			row, err := tables[tablename].Next()
			if err != nil {
				// Skip faulty data
				continue
			}

			m, err := row.ToMap()
			if err != nil {
				// Skip conversion error
				continue
			}

			table.Data = append(table.Data, m)
		}

		databaseSchema.Tables[strings.ToUpper(tablename)] = table.Name
		databaseSchema.TableReferences = append(databaseSchema.TableReferences, &table)
	}
	duration := time.Since(start)
	databaseSchema.Generated = duration
	return databaseSchema, nil
}
