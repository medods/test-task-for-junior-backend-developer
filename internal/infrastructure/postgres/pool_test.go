package postgres

import (
	"context"
	"testing"
)

func TestOpen_EmptyDSN(t *testing.T) {
	_, err := Open(context.Background(), "")
	if err == nil {
		t.Fatalf("expected error for empty dsn")
	}
}
