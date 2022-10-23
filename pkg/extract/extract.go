package extract

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Valentin-Kaiser/go-dbase-export/pkg/model"
	"github.com/Valentin-Kaiser/go-dbase/dbase"
	"github.com/schollz/progressbar/v3"
)

func Extract(path string, export string, format string) (*model.DatabaseSchema, error) {
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
		progresses[tablename].RenderBlank()
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
			progresses[tablename].Finish()
			continue
		}

		for !tables[tablename].EOF() {
			progresses[tablename].Add(1)
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

// ToByteString returns the number of bytes as a string with a unit
func toByteString(b int) string {
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
