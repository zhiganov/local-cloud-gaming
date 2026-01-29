package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type GPU struct {
	ID        int       `json:"id"`
	Available bool      `json:"available"`
	UserID    string    `json:"user_id,omitempty"`
	SessionStart time.Time `json:"session_start,omitempty"`
	Port      int       `json:"port"`
}

type Session struct {
	UserID    string    `json:"user_id"`
	GPUID     int       `json:"gpu_id"`
	Port      int       `json:"port"`
	PIN       string    `json:"pin"`
	StartedAt time.Time `json:"started_at"`
}

type Orchestrator struct {
	mu       sync.RWMutex
	gpus     []GPU
	sessions map[string]*Session
	basePort int
	timeout  time.Duration
}

func NewOrchestrator(gpuCount, basePort int, timeout time.Duration) *Orchestrator {
	gpus := make([]GPU, gpuCount)
	for i := 0; i < gpuCount; i++ {
		gpus[i] = GPU{
			ID:        i,
			Available: true,
			Port:      basePort + (i * 10), // 10 ports per instance
		}
	}
	return &Orchestrator{
		gpus:     gpus,
		sessions: make(map[string]*Session),
		basePort: basePort,
		timeout:  timeout,
	}
}

func (o *Orchestrator) Status() []GPU {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.gpus
}

func (o *Orchestrator) Claim(userID string) (*Session, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Check if user already has a session
	if session, exists := o.sessions[userID]; exists {
		return session, nil
	}

	// Find available GPU
	for i := range o.gpus {
		if o.gpus[i].Available {
			o.gpus[i].Available = false
			o.gpus[i].UserID = userID
			o.gpus[i].SessionStart = time.Now()

			session := &Session{
				UserID:    userID,
				GPUID:     o.gpus[i].ID,
				Port:      o.gpus[i].Port,
				PIN:       generatePIN(),
				StartedAt: time.Now(),
			}
			o.sessions[userID] = session

			// TODO: Start/configure Sunshine container for this GPU
			log.Printf("Assigned GPU %d to user %s on port %d", i, userID, session.Port)
			return session, nil
		}
	}

	return nil, ErrNoAvailableGPU
}

func (o *Orchestrator) Release(userID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	session, exists := o.sessions[userID]
	if !exists {
		return ErrSessionNotFound
	}

	o.gpus[session.GPUID].Available = true
	o.gpus[session.GPUID].UserID = ""
	o.gpus[session.GPUID].SessionStart = time.Time{}
	delete(o.sessions, userID)

	log.Printf("Released GPU %d from user %s", session.GPUID, userID)
	return nil
}

func (o *Orchestrator) CleanupExpired() {
	o.mu.Lock()
	defer o.mu.Unlock()

	now := time.Now()
	for userID, session := range o.sessions {
		if now.Sub(session.StartedAt) > o.timeout {
			o.gpus[session.GPUID].Available = true
			o.gpus[session.GPUID].UserID = ""
			delete(o.sessions, userID)
			log.Printf("Expired session for user %s on GPU %d", userID, session.GPUID)
		}
	}
}

var (
	ErrNoAvailableGPU  = &HTTPError{Status: http.StatusServiceUnavailable, Message: "no GPU available"}
	ErrSessionNotFound = &HTTPError{Status: http.StatusNotFound, Message: "session not found"}
)

type HTTPError struct {
	Status  int    `json:"-"`
	Message string `json:"error"`
}

func (e *HTTPError) Error() string { return e.Message }

func generatePIN() string {
	// Simple 4-digit PIN for Moonlight pairing
	return strconv.Itoa(1000 + int(time.Now().UnixNano()%9000))
}

func main() {
	gpuCount, _ := strconv.Atoi(getEnv("GPU_COUNT", "1"))
	basePort, _ := strconv.Atoi(getEnv("BASE_PORT", "47984"))
	timeout, _ := time.ParseDuration(getEnv("SESSION_TIMEOUT", "4h"))

	orch := NewOrchestrator(gpuCount, basePort, timeout)

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			orch.CleanupExpired()
		}
	}()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// API routes
	r.Get("/api/status", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(orch.Status())
	})

	r.Post("/api/claim", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "user_id required", http.StatusBadRequest)
			return
		}
		session, err := orch.Claim(userID)
		if err != nil {
			if httpErr, ok := err.(*HTTPError); ok {
				http.Error(w, httpErr.Message, httpErr.Status)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(session)
	})

	r.Post("/api/release", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "user_id required", http.StatusBadRequest)
			return
		}
		if err := orch.Release(userID); err != nil {
			if httpErr, ok := err.(*HTTPError); ok {
				http.Error(w, httpErr.Message, httpErr.Status)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Serve dashboard
	r.Handle("/*", http.FileServer(http.Dir("./dashboard/dist")))

	log.Printf("Starting orchestrator on :8080 with %d GPUs", gpuCount)
	log.Fatal(http.ListenAndServe(":8080", r))
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
