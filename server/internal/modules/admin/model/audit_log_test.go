package model

import "testing"

func TestAuditLogBeforeCreateGeneratesID(t *testing.T) {
	log := &AuditLog{}

	if err := log.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate() error = %v", err)
	}

	if log.ID == "" {
		t.Fatal("expected ID to be generated")
	}
}
