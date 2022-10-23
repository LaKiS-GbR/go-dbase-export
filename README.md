# FoxProd/Base exporter

This is a simple tool to export data from dBase/FoxPro databases. 
Based on [go-dbase](github.com/Valentin-Kaiser/go-dbase).

## Build

```bash
cd cmd/
go build -o dbase-exporter
```

## Usage

```bash
dbase-exporter -h

Usage of dbase-exporter:
-debug-file string
        Path to the debug file
  -debug-screen
        Log debug information to the screen
  -export string
        Path to the export folder (default "./export/")
  -format string
        Format type of the export (json, yaml/yml, toml, csv, xlsx) (default "json")
  -path string
        Path to the database file
```
