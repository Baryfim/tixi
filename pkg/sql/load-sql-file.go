package sql

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

func LoadSQLFile(filePath string) (string, error) {
	path := filepath.Join("sql", filePath)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("не удалось прочитать файл: %w", err)
	}
	return string(content), nil
}
