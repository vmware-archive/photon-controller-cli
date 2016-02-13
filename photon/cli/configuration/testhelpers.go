package configuration

import (
	"os"
)

func RemoveConfigFile() error {
	filepath, err := getConfigurationFilePath()
	if err != nil {
		return err
	}

	if isFileExist(filepath) {
		err = os.Remove(filepath)
		if err != nil {
			return err
		}
	}

	return nil
}

func ChangeConfigFileContents(content string) error {
	filepath, err := getConfigurationFilePath()
	if err != nil {
		return err
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	err = file.Sync()
	if err != nil {
		return err
	}

	return nil
}
