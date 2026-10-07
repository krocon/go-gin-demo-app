package tasks

import (
	"context"
	"errors"
	"testing"
)

func TestIDsAreAssignedSequentially(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()

	first, _ := repo.Create(ctx, "First")
	second, _ := repo.Create(ctx, "Second")

	if first.ID != 1 || second.ID != 2 {
		t.Fatalf("got ids %d and %d, want 1 and 2", first.ID, second.ID)
	}
}

func TestUpdateOnlyChangesGivenFields(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()
	task, _ := repo.Create(ctx, "Original")

	done := true
	updated, err := repo.Update(ctx, task.ID, UpdateTaskRequest{Done: &done})

	if err != nil {
		t.Fatal(err)
	}
	want := Task{ID: task.ID, Title: "Original", Done: true}
	if updated != want {
		t.Fatalf("got %+v, want %+v", updated, want)
	}
}

func TestUpdateOfUnknownIDReturnsErrNotFound(t *testing.T) {
	repo := NewInMemoryRepository()
	title := "Nope"

	_, err := repo.Update(context.Background(), 99, UpdateTaskRequest{Title: &title})

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestDeleteReportsWhetherATaskWasRemoved(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()
	task, _ := repo.Create(ctx, "To delete")

	if err := repo.Delete(ctx, task.ID); err != nil {
		t.Fatalf("first delete: %v", err)
	}
	if err := repo.Delete(ctx, task.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second delete: got %v, want ErrNotFound", err)
	}
}
