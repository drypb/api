package analysis

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"os"

	"github.com/gabriel-vasile/mimetype"
)

const (
	TimeFormat string = "02-01-2006 15:04:05 MST"
)

func getMimeType(path string) (string, error) {
	mime, err := mimetype.DetectFile(path)
	if err != nil {
		return "", err
	}

	return mime.String(), nil
}

func getSize(path string) (int64, error) {
	file, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return file.Size(), nil
}

func getLastModified(path string) (string, error) {
	file, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	modtime := file.ModTime()
	timestamp := modtime.Format(TimeFormat)

	return timestamp, nil
}

func getMD5Sum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	io.Copy(hash, file)

	md5sum := hex.EncodeToString(hash.Sum(nil))

	return md5sum, nil
}

func getSHA1Sum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha1.New()
	io.Copy(hash, file)

	sha1sum := hex.EncodeToString(hash.Sum(nil))

	return sha1sum, nil
}

func getSHA256Sum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	io.Copy(hash, file)

	sha256sum := hex.EncodeToString(hash.Sum(nil))

	return sha256sum, nil
}

func getDriverVersion(template int) string {
	version := map[int]string{
		9011: "1.0.3.2",
	}
	return version[template]
}

func getFatalEnv(env string) string {
	e := os.Getenv(env)
	if e == "" {
		log.Fatalf("%s is empty", e)
	}
	return e
}
