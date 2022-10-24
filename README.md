# 🦊 FoxPro/dBase exporter

[![License](https://img.shields.io/badge/License-BSD_3--Clause-blue.svg)](https://github.com/Valentin-Kaiser/go-dbase-export/blob/main/LICENSE)
[![Linters](https://github.com/Valentin-Kaiser/go-dbase-export/workflows/Linters/badge.svg)](https://github.com/Valentin-Kaiser/go-dbase-export)
[![CodeQL](https://github.com/Valentin-Kaiser/go-dbase-export/workflows/CodeQL/badge.svg)](https://github.com/Valentin-Kaiser/go-dbase-export)
[![Go Report](https://goreportcard.com/badge/github.com/Valentin-Kaiser/go-dbase-export)](https://goreportcard.com/report/github.com/Valentin-Kaiser/go-dbase-export)

**This is a simple tool to export data from dBase/FoxPro databases.**
Based on [go-dbase](github.com/Valentin-Kaiser/go-dbase).

## Build

```bash
cd cmd/
go build -o dbase-exporter
```

## Usage

The application can be used as a CLI tool or via the web browser.

> If executed without any parameters, the application starts with web server and the export can only be executed via it. For configuration, a file is created in which the port, database path and export path can be entered.

```txt
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
        Path to the FoxPro/dBase database  file (DATABASE.DBC)
  -repository string
        Path to the repository folder (Used to store the uploaded files) (default "./repository")
  -run
        Run the export in cli
```
