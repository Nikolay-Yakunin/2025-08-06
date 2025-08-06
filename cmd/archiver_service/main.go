package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitlab.com/Nikolay-Yakunin/2025-08-06/internal/taskmanager"
	"gitlab.com/Nikolay-Yakunin/2025-08-06/pkg/config"
)

func main() {
	cfg := config.NewConfig()

	// Дебаг мод
	debug := cfg.Mode == "debug"

	taskManager := taskmanager.NewTaskManager(cfg.MaxTasks, log.Default(), debug)

	// GET /task - создать новую таску, вернуть uuid
	http.HandleFunc("/task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		id, err := taskManager.CreateTask([]string{})
		if err != nil {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"error": "server busy"})
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"task_id": id})
	})

	// POST /task/{task_id} и GET /task/{task_id}
	http.HandleFunc("/task/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 2 || parts[0] != "task" {
			log.Printf("Not found: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		taskID := parts[1]

		if r.Method == http.MethodPost {
			var req struct {
				URL string `json:"url"`
			}

			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
				log.Printf("Invalid body for task %s", taskID)
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid body"})
				return
			}

			log.Printf("Add url to task %s: %s", taskID, req.URL)
			err := taskManager.AddURL(taskID, []string{req.URL})
			if err != nil {
				log.Printf("Task not found: %s", taskID)
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "task not found"})
				return
			}

			status, _ := taskManager.GetStatus(taskID)
			log.Printf("Task %s status after add: %s", taskID, status)
			json.NewEncoder(w).Encode(map[string]string{"status": string(status)})
			return
		}

		if r.Method == http.MethodGet {
			status, err := taskManager.GetStatus(taskID)
			if err != nil {
				log.Printf("Task not found: %s", taskID)
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "task not found"})
				return
			}

			resp := map[string]string{"status": string(status)}
			if status == "completed" {
				resp["download_url"] = "/download/" + taskID
				log.Printf("Task %s completed, archive ready", taskID)
			}

			log.Printf("Task %s status: %s", taskID, status)
			json.NewEncoder(w).Encode(resp)
			return
		}

		log.Printf("Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// GET /download/{task_id}
	http.HandleFunc("/download/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 2 || parts[0] != "download" {
			log.Printf("Not found: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Берем с url id
		taskID := parts[1]
		archivePath := filepath.Join("/tmp/archiver", taskID, "archive.zip")
		f, err := os.Open(archivePath)
		if err != nil {
			log.Printf("Archive not found for task %s", taskID)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "archive not found"})
			return
		}
		defer f.Close()

		// Заголовки
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=archive.zip")
		log.Printf("Serving archive for task %s", taskID)
		io.Copy(w, f)
	})

	server := &http.Server{
		Addr: cfg.Port,
	}

	stop := make(chan os.Signal, 1)

	go func() {
		log.Printf("Server starting on %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
