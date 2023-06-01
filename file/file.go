package file

import (
	"bufio"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Чтение файла полностью (в строку)
func Read(filename string) (string, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// Чтение файла построчно (в массив)
func ReadLines(filename string) ([]string, error) {
	result := make([]string, 0)

	file, err := os.Open(filename)
	if err != nil {
		return result, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		result = append(result, scanner.Text())
	}

	return result, nil
}

// Запись в файл
func Write(filename, filetext string) (int, error) {
	if filename == "" {
		return 0, errors.New("имя файла не задано")
	}
	file, err := os.Create(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return file.WriteString(filetext)
}

// Смена рабочего каталога
func ChangeDir(dir string) {
	path, _ := os.Getwd()
	dir = filepath.Join(path, dir)

	rel, err := filepath.Rel(path, dir)
	if err == nil {
		os.Chdir(rel)
	}
}
