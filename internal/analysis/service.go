package analysis

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// AnalysisStatus represents the status of an analysis task.
type AnalysisStatus string

const (
	// StatusPending means the analysis is waiting to be processed.
	StatusPending AnalysisStatus = "pending"
	// StatusProcessing means the analysis is currently being processed.
	StatusProcessing AnalysisStatus = "processing"
	// StatusCompleted means the analysis has been completed.
	StatusCompleted AnalysisStatus = "completed"
	// StatusFailed means the analysis failed.
	StatusFailed AnalysisStatus = "failed"
)

// Analysis represents a single analysis task.
type Analysis struct {
	ID        string         `json:"id"`
	URL       string         `json:"url"`
	Status    AnalysisStatus `json:"status"`
	Result    interface{}    `json:"result,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// Service provides operations for managing analysis tasks.
type Service struct {
	mu       sync.RWMutex
	analyses map[string]*Analysis
	queue    chan string
}

// NewService creates a new analysis service.
func NewService(queueSize int) *Service {
	return &Service{
		analyses: make(map[string]*Analysis),
		queue:    make(chan string, queueSize),
	}
}

// Create new analysis task and add it to the queue.
func (s *Service) Create(url string) *Analysis {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.New().String()
	analysis := &Analysis{
		ID:        id,
		URL:       url,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.analyses[analysis.ID] = analysis
	s.queue <- analysis.ID
	return analysis
}

// GetAll returns all analysis tasks.
func (s *Service) GetAll() []*Analysis {
	s.mu.RLock()
	defer s.mu.RUnlock()

	analyses := make([]*Analysis, 0, len(s.analyses))
	for _, analysis := range s.analyses {
		analyses = append(analyses, analysis)
	}
	return analyses
}

// GetCompleted returns all completed analysis tasks.
func (s *Service) GetCompleted() []*Analysis {
	s.mu.RLock()
	defer s.mu.RUnlock()

	analyses := make([]*Analysis, 0, len(s.analyses))
	for _, analysis := range s.analyses {
		if analysis.Status == StatusCompleted {
			analyses = append(analyses, analysis)
		}
	}
	return analyses
}

// GetByID returns an analysis task by its ID.
func (s *Service) GetByID(id string) (*Analysis, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	analysis, ok := s.analyses[id]
	return analysis, ok
}

// GetNextFromQueue gets the next analysis ID from the queue. This will block if the queue is empty.
func (s *Service) GetNextFromQueue() string {
	return <-s.queue
}

// UpdateStatus updates the status of an analysis task.
func (s *Service) UpdateStatus(id string, status AnalysisStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if analysis, ok := s.analyses[id]; ok {
		analysis.Status = status
		analysis.UpdatedAt = time.Now()
	}
}

// UpdateResult updates the result of an analysis task.
func (s *Service) UpdateResult(id string, status AnalysisStatus, result interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if analysis, ok := s.analyses[id]; ok {
		analysis.Status = status
		analysis.Result = result
		analysis.UpdatedAt = time.Now()
	}
}
