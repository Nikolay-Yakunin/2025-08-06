package archiver

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// Нужно нормальное название.
type Archiver interface {
	CreateZip(files []string, dest string) error
}

type ZipArchiver struct{} // any не подходит.

// Конструктор архиватора.
func NewZipArchiver() *ZipArchiver {
	return &ZipArchiver{}
}

// Создает и заполняет zip архив.
func (a *ZipArchiver) CreateZip(files []string, dest string) error {
	// Проверка существования директории.
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	zipFile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer func() {
			if err := zipFile.Close(); err != nil {
				return
			}
		}()

	zipWriter := zip.NewWriter(zipFile)
	defer func() {
			if err := zipWriter.Close(); err != nil {
				return
			}
		}()

	// Добавление файлов в архив.
	for _, file := range files {
		if err := addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

// Добавляет файл в архив.
func addFileToZip(zipWriter *zip.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
			if err := file.Close(); err != nil {
				return
			}
		}()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}
