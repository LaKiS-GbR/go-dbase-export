package serialize

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/Valentin-Kaiser/go-dbase-export/pkg/model"
	"github.com/pelletier/go-toml/v2"
	"github.com/schollz/progressbar/v3"
	"gopkg.in/yaml.v3"
)

func SerializeSchema(dbSchema *model.DatabaseSchema, export string, format string) {
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
	err := serializationBar.RenderBlank()
	if err != nil {
		log.Fatalf("Rendering progress bar failed with error: %v", err)
	}
	for i, table := range dbSchema.TableReferences {
		serializationBar.Describe(fmt.Sprintf("%-20.20s %-10.10s %10.10s", "Serialising table...", table.Name, fmt.Sprintf("(%d/%d)", i+1, len(dbSchema.TableReferences)+1)))
		data, err := serialize(table, export, format)
		if err != nil {
			log.Fatalf("Table serialization failed with error: %v", err)
		}

		serializationBar.Describe(fmt.Sprintf("%-20.20s %-10.10s %10.10s", "Saving table to file", table.Name, fmt.Sprintf("(%d/%d)", i+1, len(dbSchema.TableReferences)+1)))
		path, err := getPath(table, export, format)
		if err != nil {
			log.Fatalf("Getting path failed with error: %v", err)
		}

		err = saveFile(path, data)
		if err != nil {
			log.Fatalf("Saving file failed with error: %v", err)
		}

		err = serializationBar.Add(1)
		if err != nil {
			log.Fatalf("Incrementing progress bar failed with error: %v", err)
		}
	}

	serializationBar.Describe(fmt.Sprintf("%-20.20s %-10.10s %10.10s", "Serializing database", dbSchema.Name, fmt.Sprintf("(%d/%d)", len(dbSchema.TableReferences)+1, len(dbSchema.TableReferences)+1)))
	data, err := serialize(dbSchema, export, format)
	if err != nil {
		log.Fatalf("Database serialization failed with error: %v", err)
	}

	serializationBar.Describe(fmt.Sprintf("%-20.20s %-10.10s %10.10s", "Saving database", dbSchema.Name, fmt.Sprintf("(%d/%d)", len(dbSchema.TableReferences)+1, len(dbSchema.TableReferences)+1)))
	dbExportPath, err := getPath(dbSchema, export, format)
	if err != nil {
		log.Fatalf("Getting path failed with error: %v", err)
	}

	err = saveFile(dbExportPath, data)
	if err != nil {
		log.Fatalf("Saving file failed with error: %v", err)
	}
	err = serializationBar.Add(1)
	if err != nil {
		log.Fatalf("Incrementing progress bar failed with error: %v", err)
	}

	fmt.Println("\nDone!")
}

func saveFile(path string, data []byte) error {
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func getPath(v interface{}, path, format string) (string, error) {
	filename := ""
	switch v := v.(type) {
	case *model.DatabaseSchema:
		filename = v.Name + "." + format
	case *model.Table:
		filename = v.Name + "." + format
	default:
		return "", fmt.Errorf("type %T is not supported", v)
	}

	return filepath.Join(path, filename), nil
}

func serialize(v interface{}, exportPath string, format string) ([]byte, error) {
	data, err := serializeFormat(v, exportPath, format)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func serializeFormat(v interface{}, exportPath string, format string) ([]byte, error) {
	switch format {
	case "json":
		return json.MarshalIndent(v, "", "  ")
	case "toml":
		return toml.Marshal(v)
	case "yaml", "yml":
		return yaml.Marshal(v)
	case "csv":
		return serializeCSV(v)
	case "xlsx":
		return serializeXLSX(v)
	default:
		return nil, fmt.Errorf("format %s is not supported", format)
	}
}

func serializeCSV(v interface{}) ([]byte, error) {
	data := make([][]string, 0)

	switch v := v.(type) {
	case *model.DatabaseSchema:
		data = append(data, []string{"Name", "Columns", "Records", "FirstRow", "RowLength", "FileSize", "Modified"})

		for _, table := range v.TableReferences {
			data = append(data, []string{
				table.Name,
				fmt.Sprintf("%d", table.Columns),
				fmt.Sprintf("%d", table.Records),
				fmt.Sprintf("%d", table.FirstRow),
				fmt.Sprintf("%d", table.RowLength),
				fmt.Sprintf("%d", table.FileSize),
				table.Modified.Format("2006-01-02 15:04:05"),
			})
		}

	case *model.Table:
		columns := make([]string, 0)
		for _, column := range v.Fields {
			columns = append(columns, column.Name)
		}
		data = append(data, columns)

		for _, row := range v.Data {
			values := make([]string, 0)
			for _, value := range row {
				values = append(values, fmt.Sprintf("%v", value))
			}
			data = append(data, values)
		}
	default:
		return nil, fmt.Errorf("type %T is not supported", v)
	}

	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	err := w.WriteAll(data)
	if err != nil {
		return nil, err
	}
	w.Flush()

	return buf.Bytes(), nil
}

func serializeXLSX(v interface{}) ([]byte, error) {
	f := excelize.NewFile()

	switch v := v.(type) {
	case *model.DatabaseSchema:
		f.SetSheetName("Sheet1", "Tables")
		f.SetCellValue("Tables", "A1", "Name")
		f.SetCellValue("Tables", "B1", "Fields per record")
		f.SetCellValue("Tables", "C1", "Total Records")
		f.SetCellValue("Tables", "D1", "Header length (bytes)")
		f.SetCellValue("Tables", "E1", "RowLength (bytes)")
		f.SetCellValue("Tables", "F1", "FileSize (Bytes)")
		f.SetCellValue("Tables", "G1", "Modified")
		for i, table := range v.TableReferences {
			f.SetCellValue("Tables", fmt.Sprintf("A%d", i+2), table.Name)
			f.SetCellValue("Tables", fmt.Sprintf("B%d", i+2), table.Columns)
			f.SetCellValue("Tables", fmt.Sprintf("C%d", i+2), table.Records)
			f.SetCellValue("Tables", fmt.Sprintf("D%d", i+2), table.FirstRow)
			f.SetCellValue("Tables", fmt.Sprintf("E%d", i+2), table.RowLength)
			f.SetCellValue("Tables", fmt.Sprintf("F%d", i+2), table.FileSize)
			f.SetCellValue("Tables", fmt.Sprintf("G%d", i+2), table.Modified.Format("2006-01-02 15:04:05"))
		}
	case *model.Table:
		f.SetSheetName("Sheet1", v.Name)
		columns := make([]string, 0)
		for _, column := range v.Fields {
			columns = append(columns, column.Name)
		}
		f.SetSheetRow(v.Name, "A1", &columns)

		for i, row := range v.Data {
			values := make([]string, 0)
			for _, value := range columns {
				values = append(values, fmt.Sprintf("%v", row[value]))
			}
			f.SetSheetRow(v.Name, fmt.Sprintf("A%d", i+2), &values)
		}
	default:
		return nil, fmt.Errorf("type %T is not supported", v)
	}

	// Get excel file as a slice of bytes
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func IsFormatSupported(format string) bool {
	switch format {
	case "json", "toml", "yaml", "yml", "csv", "xlsx":
		return true
	default:
		return false
	}
}
