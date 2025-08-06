package task

import (
	"fmt"
	"sync"
)

type TaskStatus string

// Статусы таски.
const (
	StatusPending    TaskStatus = "pending"
	StatusProcessing TaskStatus = "processing"
	StatusCompleted  TaskStatus = "completed"
	StatusFailed     TaskStatus = "failed"
)

// Ошибки
// - URL: Для того чтобы вернуть "имя" файла,
// - Error: Для самой ошибки.
type FileError struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

type Task struct {
	TaskID   string      `json:"task_id"`
	URLs     []string    `json:"urls"`
	MaxFiles int         `json:"-"` // Не должно быть в json-е
	Status   TaskStatus  `json:"status"`
	Errors   []FileError `json:"errors"`
	Mu       sync.RWMutex
	// Должна ли таска знать о пути к архиву? Ну по сути, task_id можно назвать путем.
}

// Конструктор таски
// taskID - uuid,
// urls - массив с url,
// MaxFiles - максимальное количество файлов.
func NewTask(taskID string, urls []string, MaxFiles int) *Task {
	return &Task{
		TaskID:   taskID,
		URLs:     urls,
		MaxFiles: MaxFiles,
		Status:   StatusPending,
		Errors:   make([]FileError, 0),
	}
}

// AddURL добавляет url в таску.
func (t *Task) AddURL(url string) error {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	if len(t.URLs) == t.MaxFiles {
		return fmt.Errorf("too many urls, max count is %d", t.MaxFiles)
	}
	t.URLs = append(t.URLs, url)
	return nil
}

// SetStatus меняет статус таски.
func (t *Task) SetStatus(status TaskStatus) {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	t.Status = status
}

// AddError добавляет ошибки, не ограниченно по размеру,
// так как плохие url, не должны занимать место.
func (t *Task) AddError(url, errMsg string) {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	t.Errors = append(t.Errors, FileError{URL: url, Error: errMsg})
}

// ----- Геттеры -----

// GetStatus возвращает статус таски.
func (t *Task) GetStatus() TaskStatus {
	t.Mu.RLock()
	defer t.Mu.RUnlock()
	return t.Status
}

// GetURLs возвращает копиб URLs таски.
func (t *Task) GetURLs() []string {
	t.Mu.RLock()
	defer t.Mu.RUnlock()
	urls := make([]string, len(t.URLs))
	copy(urls, t.URLs)
	return urls
}

// GetErrors возвращает копию ошибок такси.
func (t *Task) GetErrors() []FileError {
	t.Mu.RLock()
	defer t.Mu.RUnlock()
	errs := make([]FileError, len(t.Errors))
	copy(errs, t.Errors)
	return errs
}
