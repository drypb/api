package analysis

import (
	"os"
	"reflect"
	"testing"

	"github.com/drypb/api/internal/assert"
	"github.com/drypb/api/internal/data"
	"github.com/google/uuid"
)

var mockedReport = &Report{
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

func TestSaveReport(t *testing.T) {
	err := os.Mkdir(data.DefaultReportPath, os.ModePerm)
	assert.NilError(t, err)
	defer os.RemoveAll(data.DefaultReportPath)
	err = mockedReport.saveReport()
	assert.NilError(t, err)
}

func TestSaveStatus(t *testing.T) {
	err := os.Mkdir(data.DefaultStatusPath, os.ModePerm)
	assert.NilError(t, err)
	defer os.RemoveAll(data.DefaultStatusPath)
	err = mockedReport.saveStatus()
	assert.NilError(t, err)
}

func TestSave(t *testing.T) {
	err := mockedReport.Save("status")
	assert.NilError(t, err)
	err = mockedReport.Save("status")
	assert.NilError(t, err)
	err = mockedReport.Save("asdf")
	assert.Error(t, err)
}

func TestLoad(t *testing.T) {
	err := os.Mkdir(data.DefaultReportPath, os.ModePerm)
	assert.NilError(t, err)
	defer os.RemoveAll(data.DefaultReportPath)

	err = mockedReport.Save("report")
	assert.NilError(t, err)

	tmp := &Report{}
	err = tmp.Load(mockedReport.Request.ID)
	assert.NilError(t, err)

	if !reflect.DeepEqual(mockedReport, tmp) {
		t.Error("Repor struct is not the same after saving and loading.")
	}
}
