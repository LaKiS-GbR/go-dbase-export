package job

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Valentin-Kaiser/go-dbase-export/pkg/extract"
	"github.com/Valentin-Kaiser/go-dbase-export/pkg/serialize"
)

type Job struct {
	ID       string
	Finished bool
	Error    error
	Start    time.Time
	End      time.Time
	Duration time.Duration
	Log      io.Writer
	Debug    io.Writer
}

func New(out, debug io.Writer) *Job {
	return &Job{
		Log:   out,
		Debug: debug,
	}
}

func (j *Job) Run(path, export, format string) *Job {
	j.Start = time.Now()
	if len(strings.TrimSpace(path)) == 0 {
		j.Finish(fmt.Errorf("please provide a path to the database file"))
		return j
	}

	if len(strings.TrimSpace(export)) == 0 {
		j.Finish(fmt.Errorf("please provide a path to the export file"))
		return j
	}

	if len(strings.TrimSpace(format)) == 0 {
		j.Finish(fmt.Errorf("please provide a format type"))
		return j
	}

	// Check if format is supported
	if !serialize.IsFormatSupported(format) {
		j.Finish(fmt.Errorf("format %v is not supported", format))
		return j
	}

	// Create the export folder if it doesn't exist
	if _, err := os.Stat(export); os.IsNotExist(err) {
		err := os.Mkdir(export, 0755)
		if err != nil {
			j.Finish(fmt.Errorf("creating export folder failed with error: %v", err))
			return j
		}
	}

	dbSchema, err := extract.Extract(path)
	if err != nil {
		j.Finish(fmt.Errorf("extracting database failed with error: %v", err))
		return j
	}

	// Serialize the schema
	serialize.SerializeSchema(dbSchema, export, format)

	j.End = time.Now()
	j.Duration = time.Since(j.Start)
	j.Finish(nil)
	return j
}

func (j *Job) IsFinished() bool {
	return j.Finished
}

func (j *Job) GetError() error {
	return j.Error
}

func (j *Job) Finish(err error) {
	j.Error = err
	j.Finished = true
}
