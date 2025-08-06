package task

import (
	"testing"
)

func TestNewTask(t *testing.T) {
	taskID := "test-task-id"
	urls := []string{"http://example.com/file1.txt", "http://example.com/file2.txt"}
	maxFiles := 5

	task := NewTask(taskID, urls, maxFiles)

	if task.TaskID != taskID {
		t.Errorf("Expected TaskID %s, got %s", taskID, task.TaskID)
	}
	if len(task.URLs) != len(urls) {
		t.Errorf("Expected %d URLs, got %d", len(urls), len(task.URLs))
	}
	if task.MaxFiles != maxFiles {
		t.Errorf("Expected MaxFiles %d, got %d", maxFiles, task.MaxFiles)
	}
	if task.Status != StatusPending {
		t.Errorf("Expected Status %s, got %s", StatusPending, task.Status)
	}
	if len(task.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(task.Errors))
	}
}

func TestAddURL_Success(t *testing.T) {
	task := NewTask("test-task", []string{}, 3)
	url := "http://example.com/file.txt"

	err := task.AddURL(url)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	urls := task.GetURLs()
	if len(urls) != 1 {
		t.Errorf("Expected 1 URL, got %d", len(urls))
	}
	if urls[0] != url {
		t.Errorf("Expected URL %s, got %s", url, urls[0])
	}
}

func TestAddURL_Error_MaxFilesReached(t *testing.T) {
	task := NewTask("test-task", []string{"url1", "url2"}, 2)
	url := "http://example.com/file.txt"

	err := task.AddURL(url)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	expectedErrMsg := "too many urls, max count is 2"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
	urls := task.GetURLs()
	if len(urls) != 2 {
		t.Errorf("Expected 2 URLs, got %d", len(urls))
	}
}

func TestSetStatus(t *testing.T) {
	task := NewTask("test-task", []string{}, 5)
	newStatus := StatusProcessing

	task.SetStatus(newStatus)

	if task.GetStatus() != newStatus {
		t.Errorf("Expected status %s, got %s", newStatus, task.GetStatus())
	}
}

func TestAddError(t *testing.T) {
	task := NewTask("test-task", []string{}, 5)
	url := "http://example.com/badfile.txt"
	errMsg := "connection timeout"

	task.AddError(url, errMsg)

	errors := task.GetErrors()
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}
	if errors[0].URL != url {
		t.Errorf("Expected URL %s, got %s", url, errors[0].URL)
	}
	if errors[0].Error != errMsg {
		t.Errorf("Expected error message %s, got %s", errMsg, errors[0].Error)
	}
}

func TestAddMultipleErrors(t *testing.T) {
	task := NewTask("test-task", []string{}, 5)
	errorsToAdd := []struct {
		url string
		err string
	}{
		{"http://example.com/file1.txt", "not found"},
		{"http://example.com/file2.txt", "timeout"},
		{"http://example.com/file3.txt", "forbidden"},
	}

	for _, e := range errorsToAdd {
		task.AddError(e.url, e.err)
	}

	errors := task.GetErrors()
	if len(errors) != len(errorsToAdd) {
		t.Errorf("Expected %d errors, got %d", len(errorsToAdd), len(errors))
	}

	for i, expected := range errorsToAdd {
		if errors[i].URL != expected.url {
			t.Errorf("Expected URL %s at index %d, got %s", expected.url, i, errors[i].URL)
		}
		if errors[i].Error != expected.err {
			t.Errorf("Expected error %s at index %d, got %s", expected.err, i, errors[i].Error)
		}
	}
}

func TestGetStatus(t *testing.T) {
	task := NewTask("test-task", []string{}, 5)
	task.SetStatus(StatusCompleted)

	status := task.GetStatus()

	if status != StatusCompleted {
		t.Errorf("Expected status %s, got %s", StatusCompleted, status)
	}
}

func TestGetURLs(t *testing.T) {
	originalURLs := []string{"url1", "url2", "url3"}
	task := NewTask("test-task", originalURLs, 5)

	urls := task.GetURLs()

	if len(urls) != len(originalURLs) {
		t.Errorf("Expected %d URLs, got %d", len(originalURLs), len(urls))
	}
	for i, url := range originalURLs {
		if urls[i] != url {
			t.Errorf("Expected URL %s at index %d, got %s", url, i, urls[i])
		}
	}

	// Копия
	urls[0] = "modified"
	originalURLsAfter := task.GetURLs()
	if originalURLsAfter[0] == "modified" {
		t.Error("GetURLs should return a copy, not the original slice")
	}
}

func TestGetErrors(t *testing.T) {
	task := NewTask("test-task", []string{}, 5)
	task.AddError("url1", "error1")
	task.AddError("url2", "error2")

	errors := task.GetErrors()

	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}

	// Должно вернуть копию
	errors[0] = FileError{URL: "modified", Error: "modified"}
	errorsAfter := task.GetErrors()
	if errorsAfter[0].URL == "modified" {
		t.Error("GetErrors should return a copy, not the original slice")
	}
}

func TestConcurrentAccess(t *testing.T) {
	task := NewTask("test-task", []string{}, 10)
	done := make(chan bool)

	// Concurrent writers
	for i := 0; i < 5; i++ {
		go func(i int) {
			_ = task.AddURL("http://example.com/file" + string(rune(i)) + ".txt")
			task.SetStatus(TaskStatus("status" + string(rune(i))))
			task.AddError("url"+string(rune(i)), "error"+string(rune(i)))
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	urls := task.GetURLs()
	errors := task.GetErrors()
	status := task.GetStatus()

	// Не должно быть паники
	_ = urls
	_ = errors
	_ = status
}
