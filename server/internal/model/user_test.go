package model

import (
	"context"
	"errors"
	"testing"
)

func TestUserModelCreateAndFind(t *testing.T) {
	pool := OpenTestDB(t)
	m := NewUserModel(pool)
	ctx := context.Background()

	id, err := m.Create(ctx, "alice@x.com", "hash-1")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id")
	}

	u, err := m.FindByEmail(ctx, "alice@x.com")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if u.ID != id || u.Email != "alice@x.com" || u.PasswordHash != "hash-1" {
		t.Fatalf("mismatch: %+v", u)
	}

	if _, err := m.FindByEmail(ctx, "nobody@x.com"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
