package analysis

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/drypb/api/internal/config"
)

// Report represents the final artifact of the analysis process.
type Report struct {
	Request RequestMetadata `json:"request_metadata"`
	Process ProcessMetadata `json:"process_metadata"`
}

type RequestMetadata struct {
	Status        string       `json:"status"`
	ID            string       `json:"id"`
	DriverVersion string       `json:"driver_version"`
	TemplateID    int          `json:"template_id"`
	StartTime     string       `json:"start_time"`
	EndTime       string       `json:"end_time"`
	Log           []string     `json:"log"`
	Error         string       `json:"error"`
	File          FileMetadata `json:"file_metadata"`
}

// Malware sample file information.
type FileMetadata struct {
	Filename     string `json:"filename"`
	Extension    string `json:"extension"`
	MimeType     string `json:"mimetype"`
	Size         int64  `json:"size"`
	LastModified string `json:"last_modified"`
	MD5Sum       string `json:"md5sum"`
	SHA1Sum      string `json:"sha1sum"`
	SHA256Sum    string `json:"sha256sum"`
}

// Malware process information
type ProcessMetadata struct {
	WindowsRegisters      []WindowsRegisters      `json:"windows_registers"`
	WindowsFS             []WindowsFileSystem     `json:"windows_fs"`
	WindowsBinariesLoaded []WindowsBinariesLoaded `json:"windows_binaries_loaded"`
	WindowsProcess        []WindowsProcess        `json:"windows_process"`
}

type WindowsRegisters struct {
	Date              string `json:"date"`
	Time              string `json:"time"`
	InfoType          string `json:"info_type"`
	RegistryOperation string `json:"registry_operation"`
	Name              string `json:"name"`
	DataType          string `json:"data_type"`
	Data              string `json:"data"`
}

type Privileges struct {
	SeIncreaseQuotaPrivilege                  string `json:"SeIncreaseQuotaPrivilege"`
	SeSecurityPrivilege                       string `json:"SeSecurityPrivilege"`
	SeTakeOwnershipPrivilege                  string `json:"SeTakeOwnershipPrivilege"`
	SeLoadDriverPrivilege                     string `json:"SeLoadDriverPrivilege"`
	SeSystemProfilePrivilege                  string `json:"SeSystemProfilePrivilege"`
	SeSystemtimePrivilege                     string `json:"SeSystemtimePrivilege"`
	SeProfileSingleProcessPrivilege           string `json:"SeProfileSingleProcessPrivilege"`
	SeIncreaseBasePriorityPrivilege           string `json:"SeIncreaseBasePriorityPrivilege"`
	SeCreatePagefilePrivilege                 string `json:"SeCreatePagefilePrivilege"`
	SeBackupPrivilege                         string `json:"SeBackupPrivilege"`
	SeRestorePrivilege                        string `json:"SeRestorePrivilege"`
	SeShutdownPrivilege                       string `json:"SeShutdownPrivilege"`
	SeDebugPrivilege                          string `json:"SeDebugPrivilege"`
	SeSystemEnvironmentPrivilege              string `json:"SeSystemEnvironmentPrivilege"`
	SeChangeNotifyPrivilege                   string `json:"SeChangeNotifyPrivilege"`
	SeRemoteShutdownPrivilege                 string `json:"SeRemoteShutdownPrivilege"`
	SeUndockPrivilege                         string `json:"SeUndockPrivilege"`
	SeManageVolumePrivilege                   string `json:"SeManageVolumePrivilege"`
	SeImpersonatePrivilege                    string `json:"SeImpersonatePrivilege"`
	SeCreateGlobalPrivilege                   string `json:"SeCreateGlobalPrivilege"`
	SeIncreaseWorkingSetPrivilege             string `json:"SeIncreaseWorkingSetPrivilege"`
	SeTimeZonePrivilege                       string `json:"SeTimeZonePrivilege"`
	SeCreateSymbolicLinkPrivilege             string `json:"SeCreateSymbolicLinkPrivilege"`
	SeDelegateSessionUserImpersonatePrivilege string `json:"SeDelegateSessionUserImpersonatePrivilege"`
}

type WindowsFileSystem struct {
	Date            string     `json:"date"`
	Time            string     `json:"time"`
	InfoType        string     `json:"info_type"`
	MJFunc          string     `json:"mj_func"`
	PID             string     `json:"pid"`
	TID             string     `json:"tid"`
	SID             string     `json:"sid"`
	TokenType       string     `json:"token_type"`
	Privileges      Privileges `json:"privileges"`
	ElevationStatus string     `json:"elevation_status"`
	ImageName       string     `json:"image_name"`
	Path            string     `json:"path"`
	FileName        string     `json:"file_name"`
}

type WindowsBinariesLoaded struct {
	Date          string `json:"date"`
	Time          string `json:"time"`
	InfoType      string `json:"info_type"`
	PID           string `json:"pid"`
	FullImageName string `json:"full_image_name"`
	FileName      string `json:"file_name"`
}

type WindowsProcess struct {
	Date            string     `json:"date"`
	Time            string     `json:"time"`
	InfoType        string     `json:"info_type"`
	PPID            string     `json:"ppid"`
	PID             string     `json:"pid"`
	Operation       string     `json:"operation"`
	TokenType       string     `json:"token_type"`
	Privileges      Privileges `json:"privileges"`
	ElevationStatus string     `json:"elevation_status"`
	ParentName      string     `json:"parent_name"`
	ChildName       string     `json:"child_name"`
}

// Load loads the report from the disk to the memory.
func (r *Report) Load(id string) error {
	path := filepath.Join(config.ReportPath, id+".json")

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
	statusPath := filepath.Join(config.StatusPath, r.Request.ID+".json")

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
	reportPath := filepath.Join(config.ReportPath, r.Request.ID+".json")

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
