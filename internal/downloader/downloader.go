package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Downloader interface {
	Download(ctx context.Context, url, dest string) error
}

type HTTPDownloader struct {
	Timeout     time.Duration
	MaxSize     int64
	AllowedExts []string
}

// Конструктор загрузчика
// timeout - время на запрос,
// maxSize - максимальный размер файла MB,
// allowedExts - массив расширений.
func NewHTTPDownloader(timeout time.Duration, maxSize int64, allowedExts []string) *HTTPDownloader {
	return &HTTPDownloader{
		Timeout:     timeout,
		MaxSize:     maxSize,
		AllowedExts: allowedExts,
	}
}

func (d *HTTPDownloader) Download(ctx context.Context, url, dest string) error {
	client := &http.Client{Timeout: d.Timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	extCaugh := 0
	for _, ext := range d.AllowedExts {
		if ext == filepath.Ext(dest) {
			extCaugh++
		}
	} // Тупая проверка, но тут нет arr.include(), поэтому так.
	if extCaugh == 0 { // Нужно было запихать сюда логер.
		return errors.New("extention is not allowed: " + filepath.Ext(dest))
	}

	if resp.ContentLength > 0 && resp.ContentLength > d.MaxSize {
		return errors.New("file too large: " + fmt.Sprint(resp.ContentLength))
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to download: " + resp.Status)
	}

	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0755); err != nil { // rwxr-xr-x виндой игнорится.
		return err
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
