package crisp

import (
	"os"
	"path"
)

func writeToFile(filePath string, data []byte) error {
	dirPath := path.Dir(filePath)
	if _, err := os.Stat(dirPath); err != nil {
		if os.IsNotExist(err) {
			// Create file
			err := os.MkdirAll(dirPath, 0755)
			if err != nil {
				return err
			}
		}
	}
	// Write data to file
	return os.WriteFile(filePath, data, 0644)
}
