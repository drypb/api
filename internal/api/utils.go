package api

import (
	"os"

	"github.com/drypb/api/internal/config"
)

func createEssentialDirs() error {
	dirs := []string{
		config.SamplePath,
		config.ReportPath,
		config.LogPath,
		config.StatusPath,
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
