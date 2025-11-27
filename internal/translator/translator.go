// Package translator needs to translate monitoring points (aka 'tochka monitoringa' from RUS to ENG).
// It uses the predefined 'locations.json' file.
package translator

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	locationRusToEng map[string]string
	once             sync.Once
	initErr          error
)

// Init loads the locations file once.
func Init(filePath string) error {
	once.Do(func() {
		log := logrus.WithFields(logrus.Fields{
			"component": "translator",
			"path":      filePath,
		})
		log.Info("Loading translations...")

		file, err := os.ReadFile(filePath)
		if err != nil {
			initErr = fmt.Errorf("failed to read translations file: %w", err)
			log.Error(initErr)
			return
		}

		if err = json.Unmarshal(file, &locationRusToEng); err != nil {
			initErr = fmt.Errorf("failed to parse translations file: %w", err)
			log.Error(initErr)
			return
		}
		log.Info("Translations loaded successfully.")
	})

	return initErr
}

// GetEngLocation returns Monitoring Point name in English
func GetEngLocation(rus string) string {
	if initErr != nil {
		return rus
	}

	if val, ok := locationRusToEng[rus]; ok {
		return val
	}

	logrus.WithField("location", rus).Warn("Translation not found for location")
	return rus
}
