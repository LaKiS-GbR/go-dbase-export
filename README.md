# dBase/FoxPro exporter

This is a simple tool to export data from dBase/FoxPro databases. Based on [go-dbase](github.com/Valentin-Kaiser/go-dbase) and [progressbar](github.com/schollz/progressbar/v3).


## Build

```bash
go build -o dbase-exporter
```


## Usage

```bash
dbase-exporter -h

Usage of ./main.go:
  -debug
        Log debug information
  -debug-file string
        Path to the debug file (default "debug.log")
  -debug-screen
        Log debug information to the screen
  -export string
        Path to the export folder (default "./export/")
  -path string
        Path to the database file
```
