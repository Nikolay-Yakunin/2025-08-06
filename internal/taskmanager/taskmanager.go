package taskmanager

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"

	"gitlab.com/Nikolay-Yakunin/2025-08-06/internal/actor"
	"gitlab.com/Nikolay-Yakunin/2025-08-06/internal/archiver"
	"gitlab.com/Nikolay-Yakunin/2025-08-06/internal/downloader"
	"gitlab.com/Nikolay-Yakunin/2025-08-06/internal/task"
	"gitlab.com/Nikolay-Yakunin/2025-08-06/pkg/config"
)

// TaskManager: Планировщик задач, использующий паттерн Actor.
type TaskManager struct {
	actor      actor.ActorInterface
	tasks      map[string]*task.Task
	mu         sync.RWMutex // Приватный мьютекс.
	maxTasks   int8         // Примитивная оптимизация, вроде map так улучшили, int на int8 заменили
	logger     *log.Logger
	cfg        *config.Config // bad practic
	downloader downloader.Downloader
	archiver   archiver.Archiver
}

type TaskCommand struct {
	TaskID  string
	URLs    []string
	ReplyCh chan any // Канал для сообщений.
}

// Конструктор TM:
// maxTasks - максимальное количество тасок(задач),
// logger - логгер,
// debug - флаг для деббаг мода, передается в приватное поле actor.
func NewTaskManager(maxTasks int8, logger *log.Logger, debug bool) *TaskManager {
	cfg := config.NewConfig() // По хорошему,
	// это должно быть в main.go, но я плохой :)
	tm := &TaskManager{
		tasks:      make(map[string]*task.Task),
		maxTasks:   maxTasks,
		logger:     logger,
		cfg:        cfg,
		downloader: downloader.NewHTTPDownloader(30*time.Second, cfg.MaxFileSize, cfg.AllowedExtensions),
		archiver:   archiver.NewZipArchiver(),
	}
	// Вообще, нужно давать нормальные имена, типа:
	// get, post, тот же CRUD, но мне было сложно придумать нормальные,
	// универсальные имена.
	actorHandlers := map[string]actor.Handler{
		"create":  tm.handleCreate,
		"add_url": tm.handleAddURL,
		"status":  tm.handleStatus,
	}
	tm.actor = actor.NewActor(10, actorHandlers, logger, debug)
	return tm
}

// handleCreate обработка создания таски.
func (tm *TaskManager) handleCreate(ctx context.Context, payload any) error {
	cmd, ok := payload.(TaskCommand)
	if !ok {
		tm.logger.Printf("Invalid payload type in handleCreate: expected TaskCommand, got %T", payload)
		return nil
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	if int8(len(tm.tasks)) >= tm.maxTasks {
		tm.logger.Printf("Task creation rejected: max tasks limit reached (%d)", tm.maxTasks)
		select {
		case cmd.ReplyCh <- "busy":
		case <-ctx.Done(): // Не самая читаемая запись, но с каналами по другому не получится.
			tm.logger.Printf("Context cancelled while sending 'busy' response")
			return ctx.Err()
		}
		return nil
	}

	id := uuid.New().String() // Просто хотел попробовать uuid.
	t := task.NewTask(id, cmd.URLs, tm.cfg.MaxFiles)
	tm.tasks[id] = t

	select {
	case cmd.ReplyCh <- id:
		tm.logger.Printf("Successfully created task %s with %d initial URLs", id, len(cmd.URLs))
	case <-ctx.Done():
		// Если контекст завершен, удаляем.
		delete(tm.tasks, id)
		tm.logger.Printf("Context cancelled, rolled back task creation for ID: %s", id)
		return ctx.Err()
	}

	return nil
}

// handleAddURL обработка добавления url.
func (tm *TaskManager) handleAddURL(ctx context.Context, payload any) error {
	cmd, ok := payload.(TaskCommand)
	if !ok {
		tm.logger.Printf("Invalid payload type in handleCreate: expected TaskCommand, got %T", payload)
		return nil
	}

	select {
	case <-ctx.Done():
		tm.logger.Printf("Context cancelled before processing add_url for task %s", cmd.TaskID)
		return ctx.Err()
	default:
	}

	tm.mu.RLock()
	t, exists := tm.tasks[cmd.TaskID]
	tm.mu.RUnlock()

	// Проверка существования таски.
	if !exists {
		tm.logger.Printf("Task not found: %s", cmd.TaskID)
		select {
		case cmd.ReplyCh <- "not_found":
		case <-ctx.Done():
			tm.logger.Printf("Context cancelled while sending 'not_found' for task %s", cmd.TaskID)
			return ctx.Err()
		}
		return nil
	}

	// Добавление url, вообще, подразумевается, что их от одного.
	var firstError error
	for _, url := range cmd.URLs {
		err := t.AddURL(url)
		if err != nil {
			tm.logger.Printf("Failed to add URL %s to task %s: %v", url, cmd.TaskID, err)
			if firstError == nil {
				firstError = err
			}
		}
	}

	if firstError != nil {
		select {
		case cmd.ReplyCh <- firstError:
		case <-ctx.Done():
			tm.logger.Printf("Context cancelled while sending error response for task %s", cmd.TaskID)
			return ctx.Err()
		}
		return nil
	}

	// Завершение.
	select {
	case cmd.ReplyCh <- "ok":
	case <-ctx.Done():
		tm.logger.Printf("Context cancelled while sending 'ok' response for task %s", cmd.TaskID)
		return ctx.Err()
	}

	// По тз, если пользователь добавил 3 url, значит нужно начать
	t.Mu.RLock()
	urls := make([]string, len(t.URLs))
	copy(urls, t.URLs)
	currentStatus := t.GetStatus()
	t.Mu.RUnlock()

	urlCount := len(urls)
	if urlCount >= tm.cfg.MaxFiles && currentStatus == task.StatusPending {
		tm.logger.Printf("Auto-starting task %s: %d URLs reached threshold %d",
			cmd.TaskID, urlCount, tm.cfg.MaxFiles)
		go tm.processTask(cmd.TaskID, urls)
	}

	return nil
}

// Планировщик клининга
// Вообще, можно добавить дату:время создания,
// и клинить проверяя в processTask, но я не хотел размывать обязанности,
// так что лучше выделить отдельно горутину для удаления.
func (tm *TaskManager) scheduleCleanup(taskID string) {
	tm.logger.Printf("Scheduled cleanup for task %s in 1 hour", taskID)

	go func() {
		time.Sleep(1 * time.Hour)

		taskDir := filepath.Join(tm.cfg.TmpPath, taskID)
		if _, err := os.Stat(taskDir); os.IsNotExist(err) {
			tm.logger.Printf("Task %s directory already removed", taskID)
			return
		}

		if err := os.RemoveAll(taskDir); err != nil {
			tm.logger.Printf("Failed to clean up task %s: %v", taskID, err)
		} else {
			tm.logger.Printf("Successfully cleaned up task %s directory", taskID)
			// После клининга, удаляем таску.

			// Вот тут, я не совсем понимаю задачу. Типо, я не хочу
			// чтобы таски вообще не удалялись, но в задаче такого нет.
			// И более того, в задаче не уточненно, как должно происходить скачивание.
			// То есть, если пользователь отойдет, процесс завершиться, все файлы скачаются,
			// и статус станет "completed", удалять таску нельзя, так как пользователь,
			// все еще, должен получить url для скачивания. А если он уйдет на 10^999 лет,
			// все это время хранить этот архив?
			tm.mu.Lock()
			delete(tm.tasks, taskID)
			tm.mu.Unlock()
		}
	}()
}

// Главный процесс.
func (tm *TaskManager) processTask(taskID string, urls []string) {
	tm.mu.RLock()
	t, exists := tm.tasks[taskID]
	tm.mu.RUnlock()
	if !exists {
		tm.logger.Printf("Task %s not found for processing", taskID)
		return
	}
	t.SetStatus(task.StatusProcessing)
	tm.logger.Printf("Processing task %s", taskID)

	// Директория для загрузок
	taskDir := filepath.Join(tm.cfg.TmpPath, taskID, "downloads")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.SetStatus(task.StatusFailed)
		tm.logger.Printf("Failed to create dir for task %s: %v", taskID, err)
		tm.scheduleCleanup(taskID)
		return
	}

	var downloadedFiles []string
	successfulDownloads := 0
	failedDownloads := 0

	// Пытаемся скачать urls
	for _, url := range urls {
		fileName := filepath.Base(url)
		destPath := filepath.Join(taskDir, fileName)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		err := tm.downloader.Download(ctx, url, destPath)
		cancel()

		if err != nil {
			failedDownloads++
			t.AddError(url, err.Error())
			tm.logger.Printf("Failed to download %s for task %s: %v", url, taskID, err)
			// Удаление url из urls, 
			// чтобы не забивать "очередь".
			t.Mu.Lock()
			for i, u := range t.URLs {
				if u == url {
					t.URLs = append(t.URLs[:i], t.URLs[i+1:]...)
					break
				}
			}
			t.Mu.Unlock()
			continue
		}

		successfulDownloads++
		downloadedFiles = append(downloadedFiles, destPath)
	}

	if len(downloadedFiles) == 0 {
		t.SetStatus(task.StatusFailed)
		tm.logger.Printf("Task %s failed", taskID)
		tm.scheduleCleanup(taskID)
		return
	}

	// Архивирование.
	archivePath := filepath.Join(tm.cfg.TmpPath, taskID, "archive.zip")
	err := tm.archiver.CreateZip(downloadedFiles, archivePath)
	if err != nil {
		t.SetStatus(task.StatusFailed)
		tm.logger.Printf("Task %s: archiving failed: %v", taskID, err)
		tm.scheduleCleanup(taskID)
		return
	}

	t.SetStatus(task.StatusCompleted)
	tm.logger.Printf("Task %s: completed, archive ready", taskID)
	tm.scheduleCleanup(taskID)
}

func (tm *TaskManager) handleStatus(ctx context.Context, payload any) error {
	cmd, ok := payload.(TaskCommand)
	if !ok {
		return nil
	}
	tm.mu.RLock()
	t, exists := tm.tasks[cmd.TaskID]
	tm.mu.RUnlock()
	if !exists {
		cmd.ReplyCh <- "not_found" // это не нужно логировать здесь
		return nil
	}
	cmd.ReplyCh <- t.GetStatus()
	return nil
}

// TODO: DeleteTask
// TODO: Ref
// ----- API -----
// Я устал писать

func (tm *TaskManager) CreateTask(urls []string) (string, error) {
	reply := make(chan any, 1)
	tm.actor.Send("create", TaskCommand{URLs: urls, ReplyCh: reply})
	res := <-reply
	if id, ok := res.(string); ok && id != "busy" {
		return id, nil
	}

	return "", context.DeadlineExceeded
}

func (tm *TaskManager) AddURL(taskID string, urls []string) error {
	reply := make(chan any, 1)
	tm.actor.Send("add_url", TaskCommand{TaskID: taskID, URLs: urls, ReplyCh: reply})
	res := <-reply
	if res == "ok" {
		return nil
	}

	return context.Canceled
}

func (tm *TaskManager) GetStatus(taskID string) (task.TaskStatus, error) {
	reply := make(chan any, 1)
	tm.actor.Send("status", TaskCommand{TaskID: taskID, ReplyCh: reply})
	res := <-reply
	if status, ok := res.(task.TaskStatus); ok {
		return status, nil
	}

	return "", context.Canceled
}
