package ai

import (
	"bytes"
	"context"
	"equinox/internal/security"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type WhisperStatus string

const (
	StatusPending    WhisperStatus = "pending"
	StatusProcessing WhisperStatus = "processing"
	StatusCompleted  WhisperStatus = "completed"
	StatusFailed     WhisperStatus = "failed"
)

type WhisperTask struct {
	ID             string
	UserID         string
	EncryptedAudio []byte // Encrypted via AES-256-GCM
	UserKey        []byte // Master Key for decryption
	Status         WhisperStatus
	Result         string
	Error          string
	CreatedAt      time.Time
}

type WhisperWorker struct {
	taskQueue   chan *WhisperTask
	results     map[string]*WhisperTask
	resultsLock sync.RWMutex
	concurrency int
	engineURL   string
}

func NewWhisperWorker(queueSize int, concurrency int, engineURL string) *WhisperWorker {
	return &WhisperWorker{
		taskQueue:   make(chan *WhisperTask, queueSize),
		results:     make(map[string]*WhisperTask),
		concurrency: concurrency,
		engineURL:   engineURL,
	}
}

func (w *WhisperWorker) Start(ctx context.Context) {
	for i := 0; i < w.concurrency; i++ {
		go w.workerLoop(ctx)
	}
	log.Printf("🎙️ WhisperWorker started with %d workers", w.concurrency)
}

func (w *WhisperWorker) Submit(task *WhisperTask) error {
	w.resultsLock.Lock()
	w.results[task.ID] = task
	w.resultsLock.Unlock()

	select {
	case w.taskQueue <- task:
		return nil
	default:
		return fmt.Errorf("whisper queue is full")
	}
}

func (w *WhisperWorker) GetStatus(id string) (*WhisperTask, bool) {
	w.resultsLock.RLock()
	defer w.resultsLock.RUnlock()
	task, ok := w.results[id]
	return task, ok
}

func (w *WhisperWorker) workerLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-w.taskQueue:
			w.processTask(task)
		}
	}
}

func (w *WhisperWorker) processTask(task *WhisperTask) {
	task.Status = StatusProcessing
	log.Printf("🎙️ Processing Whisper task %s for user %s", task.ID, task.UserID)

	// 1. Create Decrypt Reader
	audioReader := bytes.NewReader(task.EncryptedAudio)
	decryptReader := security.NewStreamDecryptReader(task.UserKey, audioReader)
	defer decryptReader.Close()

	log.Printf("🎙️ Decrypting %d bytes for whisper task %s", len(task.EncryptedAudio), task.ID)

	// 2. Pipe to Whisper Container
	req, err := http.NewRequest("POST", w.engineURL, decryptReader)
	if err != nil {
		task.Error = "failed to create whisper request: " + err.Error()
		task.Status = StatusFailed
		return
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		task.Error = "whisper engine unreachable: " + err.Error()
		task.Status = StatusFailed
		log.Printf("❌ Whisper task %s failed: %v", task.ID, err)
		return
	}
	defer resp.Body.Close()

	log.Printf("🎙️ Whisper engine response: %d (Task %s)", resp.StatusCode, task.ID)

	if resp.StatusCode != http.StatusOK {
		task.Error = fmt.Sprintf("whisper engine returned error %d", resp.StatusCode)
		task.Status = StatusFailed
		return
	}

	// 3. Read result
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		task.Error = "failed to read transcription: " + err.Error()
		task.Status = StatusFailed
		return
	}

	task.Result = string(resBody)
	task.Status = StatusCompleted
	
	// 4. Secure Purge
	task.EncryptedAudio = nil 
	security.ZeroMemory(task.UserKey)
	task.UserKey = nil

	log.Printf("✅ Completed Whisper task %s: %d chars transcribed", task.ID, len(task.Result))
}

// ClearResults periodically removes old results to prevent RAM leak
func (w *WhisperWorker) ClearResults(olderThan time.Duration) {
	w.resultsLock.Lock()
	defer w.resultsLock.Unlock()
	now := time.Now()
	for id, task := range w.results {
		if now.Sub(task.CreatedAt) > olderThan {
			delete(w.results, id)
		}
	}
}
