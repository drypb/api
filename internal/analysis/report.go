package analysis

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/drypb/api/internal/data"
)

// Report represents the final artifact of the analysis process.
type Report struct {
	Request RequestMetadata `json:"requestMetadata"`
	Process ProcessMetadata `json:"processMetadata"`
}

type RequestMetadata struct {
	Status        string       `json:"status"`
	ID            string       `json:"id"`
	DriverVersion string       `json:"driverVersion"`
	TemplateID    int          `json:"templateID"`
	StartTime     string       `json:"startTime"`
	EndTime       string       `json:"endTime"`
	Log           []string     `json:"log"`
	Error         string       `json:"error"`
	File          FileMetadata `json:"fileMetadata"`
}

// Malware sample file information.
type FileMetadata struct {
	Filename     string `json:"filename"`
	Extension    string `json:"extension"`
	MimeType     string `json:"mimetype"`
	Size         int64  `json:"size"`
	LastModified string `json:"lastModified"`
	MD5Sum       string `json:"md5sum"`
	SHA1Sum      string `json:"sha1sum"`
	SHA256Sum    string `json:"sha256sum"`
}

// Malware process information
type ProcessMetadata struct {
	WindowsRegisters      []WindowsRegisters      `json:"reg"`
	WindowsFS             []WindowsFileSystem     `json:"fs"`
	WindowsBinariesLoaded []WindowsBinariesLoaded `json:"load"`
	WindowsProcess        []WindowsProcess        `json:"proc"`
}

type WindowsRegisters struct {
	Date              string `json:"date"`
	Time              string `json:"time"`
	InfoType          string `json:"info type"`
	RegistryOperation string `json:"registry operation"`
	Name              string `json:"name"`
	DataType          string `json:"data type"`
	Data              string `json:"data"`
}

type WindowsFileSystem struct {
	Date            string              `json:"date"`
	Time            string              `json:"time"`
	InfoType        string              `json:"info type"`
	MJFunc          string              `json:"mjFunc"`
	PID             string              `json:"pid"`
	TID             string              `json:"tid"`
	SID             string              `json:"sid"`
	TokenType       string              `json:"token type"`
	Privileges      []map[string]string `json:"privileges"`
	ElevationStatus string              `json:"elevation status"`
	ImageName       string              `json:"image name"`
	Path            string              `json:"path"`
	FileName        string              `json:"fileName"`
}

type WindowsBinariesLoaded struct {
	Date          string `json:"date"`
	Time          string `json:"time"`
	InfoType      string `json:"info type"`
	PID           string `json:"pid"`
	FullImageName string `json:"full image name"`
	Filename      string `json:"filename"`
}

type WindowsProcess struct {
	Date            string              `json:"date"`
	Time            string              `json:"time"`
	InfoType        string              `json:"info type"`
	PPID            string              `json:"ppid"`
	PID             string              `json:"pid"`
	Operation       string              `json:"operation"`
	TokenType       string              `json:"token type"`
	Privileges      []map[string]string `json:"privileges"`
	ElevationStatus string              `json:"elevation status"`
	ParentName      string              `json:"parent name"`
	ChildName       string              `json:"child name"`
}

// Load loads the report from the disk to the memory.
func (r *Report) Load(id string) error {
	path := filepath.Join(data.DefaultReportPath, id+".json")

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&r)
	if err != nil {
		return err
	}

	return nil
}

// Save saves the report or status from memory to disk.
func (r *Report) Save(what string) error {
	switch what {
	case "status":
		r.saveStatus()
	case "report":
		r.saveReport()
	default:
		return fmt.Errorf("report: invalid option")
	}
	return nil
}

func (r *Report) saveStatus() error {
	if r.Request.ID == "" {
		return errors.New("Analysis ID is not set")
	}
	statusPath := filepath.Join(data.DefaultStatusPath, r.Request.ID+".json")

	file, err := os.Create(statusPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(&r.Request)
	if err != nil {
		return err
	}

	return nil
}

func (r *Report) saveReport() error {
	if r.Request.ID == "" {
		return errors.New("Analysis ID is not set")
	}
	reportPath := filepath.Join(data.DefaultReportPath, r.Request.ID+".json")

	file, err := os.Create(reportPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(&r)
	if err != nil {
		return err
	}
	file.Close()

	return nil
}

// LogThis adds a message to [Report].
func (r *Report) LogThis(message string) {
	r.Request.Log = append(r.Request.Log, formatLog(message))
	r.Save("status")
}

func formatLog(message string) string {
	now := time.Now().Format("15:04:05.00")
	return fmt.Sprintf("[%s] %s", now, message)
}
