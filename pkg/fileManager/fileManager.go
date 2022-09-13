package fileManager

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
)

func WriteToFile(filepath string, content []byte) error {
	filedir := path.Dir(filepath)
	os.MkdirAll(filedir, 0755)
	err := os.WriteFile(filepath, content, 0644)
	return err
}

func ReadFile(filepath string) ([]byte, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func GetAlleroHomedir() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		homedir = "."
	}

	return fmt.Sprintf("%s/.allero", homedir)
}

func ReadFolder(folderPath string) []fs.FileInfo {
	files, _ := ioutil.ReadDir(folderPath)
	return files
}

func IsExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
