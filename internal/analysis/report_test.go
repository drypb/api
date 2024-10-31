package analysis

import (
	"os"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"gitlab.c3sl.ufpr.br/saci/api/internal/data"
)

func TestReportFileIO(t *testing.T) {
	if err := os.Mkdir(data.DefaultReportPath, os.ModePerm); err != nil {
		t.Fatalf("Failed to create reports directory")
	}
	defer os.RemoveAll(data.DefaultReportPath)

	r := &Report{
		Request: RequestMetadata{
			ID:            uuid.New().String(),
			DriverVersion: "1.0.0",
			TemplateID:    9011,
			StartTime:     "20-08-2024 14:18:40 -03",
			EndTime:       "20-08-2024 14:32:07 -03",
			Log:           []string{"inf0", "inf1"},
			Error:         "error!",
			File: FileMetadata{
				Filename:     "malware.exe",
				Extension:    ".exe",
				MimeType:     "application/octet-stream",
				Size:         1234567,
				LastModified: "12-05-2017 01:47:32 -03",
				MD5Sum:       "8e7ac89b4b050ec9e9f8e19cb54d3ede",
				SHA1Sum:      "589a39a1fecd04ff549cb6944625ffd3137328ef",
				SHA256Sum:    "157eb7e0e4b861b9b107fe43219d39b8d1f629e6fb3d089bfedb933de11ea190",
			},
		},
	}

	if err := r.SaveAll(); err != nil {
		t.Fatalf("Failed to generate example status file: %v", err)
	}

	tmp := &Report{}
	if err := tmp.Load(r.Request.ID); err != nil {
		t.Fatalf("Failed to load status file: %v", err)
	}

	if !reflect.DeepEqual(r, tmp) {
		t.Error("Repor struct is not the same after saving and loading.")
	}
}
