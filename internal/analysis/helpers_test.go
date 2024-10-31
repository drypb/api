package analysis

import (
	"os"
	"testing"
)

func TestGetMimeType(t *testing.T) {
	tmp, err := os.CreateTemp("", "go-test-")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	tmp.WriteString("this is a test!")

	mimetype, err := getMimeType(tmp.Name())
	if err != nil {
		t.Fatalf("GetMimeType returned an error: %v", err)
	}

	expected := "text/plain; charset=utf-8"

	if mimetype != expected {
		t.Error("GetMimeType returned the wrong mimetype")
		t.Logf("Expected: %s\nGot: %s", expected, mimetype)
	}
}

func TestGetSize(t *testing.T) {
	tmp, err := os.CreateTemp("", "go-test-")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	expected, _ := tmp.WriteString("this is a test!")

	size, err := getSize(tmp.Name())
	if err != nil {
		t.Fatalf("GetSize returned an error: %v", err)
	}

	if size != int64(expected) {
		t.Error("GetSize returned the wrong size")
		t.Logf("Expected: %d\nGot: %d", expected, size)
	}
}

func TestGetMD5Sum(t *testing.T) {
	tmp, err := os.CreateTemp("", "go-test-")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	tmp.WriteString("this is a test!")

	sum, err := getMD5Sum(tmp.Name())
	if err != nil {
		t.Fatalf("GetMD5Sum returned an error: %v", err)
	}

	expected := "89742a09d9b41329b850b76a76b05e00"

	if sum != expected {
		t.Error("GetMD5Sum returned the wrong sum")
		t.Logf("Expected: %s\nGot: %s", expected, sum)
	}
}

func TestGetSHA1Sum(t *testing.T) {
	tmp, err := os.CreateTemp("", "go-test-")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	tmp.WriteString("this is a test!")

	sum, err := getSHA1Sum(tmp.Name())
	if err != nil {
		t.Fatalf("GetSHA1Sum returned an error: %v", err)
	}

	expected := "3aa4cb08d481cfe2b08e4a5e31777f642263d58d"

	if sum != expected {
		t.Error("GetSHA1Sum returned the wrong sum")
		t.Logf("Expected: %s\nGot: %s", expected, sum)
	}
}

func TestGetSHA256Sum(t *testing.T) {
	tmp, err := os.CreateTemp("", "go-test-")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	tmp.WriteString("this is a test!")

	sum, err := getSHA256Sum(tmp.Name())
	if err != nil {
		t.Fatalf("GetSHA256Sum returned an error: %v", err)
	}

	expected := "ca7f87917e4f5029f81ec74d6711f1c587dca0fe91ec82b87bb77aeb15e6566d"

	if sum != expected {
		t.Error("GetSHA256Sum returned the wrong sum")
		t.Logf("Expected: %s\nGot: %s", expected, sum)
	}
}
