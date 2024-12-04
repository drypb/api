package analysis

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMimeType(t *testing.T) {
	tmp, err := os.CreateTemp("", "go-test-")
	assert.Nil(t, err)
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	tmp.WriteString("this is a test!")

	got, err := getMimeType(tmp.Name())
	assert.Nil(t, err)
	want := "text/plain; charset=utf-8"
	assert.Equal(t, got, want)
}

func TestGetSize(t *testing.T) {
	tmp, err := os.CreateTemp("", "go-test-")
	assert.Nil(t, err)
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	want, _ := tmp.WriteString("this is a test!")
	got, err := getSize(tmp.Name())
	assert.Nil(t, err)
	assert.Equal(t, got, int64(want))
}

func TestGetMD5Sum(t *testing.T) {
	tmp, err := os.CreateTemp("", "go-test-")
	assert.Nil(t, err)
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	tmp.WriteString("this is a test!")
	got, err := getMD5Sum(tmp.Name())
	assert.Nil(t, err)
	want := "89742a09d9b41329b850b76a76b05e00"
	assert.Equal(t, got, want)
}

func TestGetSHA1Sum(t *testing.T) {
	tmp, err := os.CreateTemp("", "go-test-")
	assert.Nil(t, err)
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	tmp.WriteString("this is a test!")

	got, err := getSHA1Sum(tmp.Name())
	assert.Nil(t, err)
	want := "3aa4cb08d481cfe2b08e4a5e31777f642263d58d"
	assert.Equal(t, got, want)
}

func TestGetSHA256Sum(t *testing.T) {
	tmp, err := os.CreateTemp("", "go-test-")
	assert.Nil(t, err)
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	tmp.WriteString("this is a test!")

	got, err := getSHA256Sum(tmp.Name())
	assert.Nil(t, err)
	want := "ca7f87917e4f5029f81ec74d6711f1c587dca0fe91ec82b87bb77aeb15e6566d"
	assert.Equal(t, got, want)
}

func TestGetFatalEnv(t *testing.T) {
	t.Run("variable exists", func(t *testing.T) {
		envName := "TEST_ENV"
		envValue := "some_value"
		os.Setenv(envName, envValue)
		defer os.Unsetenv(envName)

		got := getFatalEnv(envName)
		want := envValue
		assert.Equal(t, got, want)
	})

	t.Run("variable doesn't exists", func(t *testing.T) {
		envName := "EMPTY_ENV"
		os.Unsetenv(envName)
		assert.Panics(t, func() { getFatalEnv(envName) })
	})
}
