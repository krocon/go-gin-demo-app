package tasks

import (
	"context"
	"slices"
	"strings"
	"sync"
)

// InMemoryRepository is a thread-safe, in-memory Repository. Data is lost on restart.
// A sync.RWMutex lets many readers in at once, but only one writer.
type InMemoryRepository struct {
	mu     sync.RWMutex
	tasks  map[int64]Task
	nextID int64
}

// Compile-time check that InMemoryRepository implements Repository.
var _ Repository = (*InMemoryRepository)(nil)

// NewInMemoryRepository returns an empty repository.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{tasks: make(map[int64]Task), nextID: 1}
}

// NewInMemoryRepositoryWithSampleData returns a repository pre-filled with a few tasks, handy for demos.
func NewInMemoryRepositoryWithSampleData() *InMemoryRepository {
	r := NewInMemoryRepository()
	for _, title := range []string{"Learn Go", "Build a Gin API", "Record a one-minute video"} {
		_, _ = r.Create(context.Background(), title)
	}
	return r
}

func (r *InMemoryRepository) FindAll(_ context.Context, done *bool) ([]Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		if done == nil || t.Done == *done {
			result = append(result, t)
		}
	}
	slices.SortFunc(result, func(a, b Task) int { return int(a.ID - b.ID) })
	return result, nil
}

func (r *InMemoryRepository) FindByID(_ context.Context, id int64) (Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	return t, nil
}

func (r *InMemoryRepository) Create(_ context.Context, title string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t := Task{ID: r.nextID, Title: strings.TrimSpace(title)}
	r.tasks[t.ID] = t
	r.nextID++
	return t, nil
}

func (r *InMemoryRepository) Update(_ context.Context, id int64, req UpdateTaskRequest) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	if req.Title != nil {
		t.Title = strings.TrimSpace(*req.Title)
	}
	if req.Done != nil {
		t.Done = *req.Done
	}
	r.tasks[id] = t // Task is a value: write the modified copy back
	return t, nil
}

func (r *InMemoryRepository) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(r.tasks, id)
	return nil
}
