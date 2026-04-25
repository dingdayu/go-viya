package viya

import (
	"context"
	"errors"
	"testing"
)

func TestGetDefaultClientReturnsErrorWhenUnset(t *testing.T) {
	SetDefaultClient(nil)

	client, err := GetDefaultClient()
	if !errors.Is(err, ErrDefaultClientNotSet) {
		t.Fatalf("GetDefaultClient() error = %v, want ErrDefaultClientNotSet", err)
	}
	if client != nil {
		t.Fatalf("GetDefaultClient() client = %v, want nil", client)
	}
}

func TestGetDefaultClientReturnsConfiguredClient(t *testing.T) {
	t.Cleanup(func() {
		SetDefaultClient(nil)
	})

	want := NewClient(context.Background(), "https://viya.example.com")
	SetDefaultClient(want)

	got, err := GetDefaultClient()
	if err != nil {
		t.Fatalf("GetDefaultClient() error = %v", err)
	}
	if got != want {
		t.Fatalf("GetDefaultClient() client = %p, want %p", got, want)
	}
}

func TestMustGetDefaultClientPanicsWhenUnset(t *testing.T) {
	SetDefaultClient(nil)

	defer func() {
		recovered := recover()
		if !errors.Is(recoveredAsError(recovered), ErrDefaultClientNotSet) {
			t.Fatalf("panic = %v, want ErrDefaultClientNotSet", recovered)
		}
	}()

	_ = MustGetDefaultClient()
}

func TestMustGetDefaultClientReturnsConfiguredClient(t *testing.T) {
	t.Cleanup(func() {
		SetDefaultClient(nil)
	})

	want := NewClient(context.Background(), "https://viya.example.com")
	SetDefaultClient(want)

	if got := MustGetDefaultClient(); got != want {
		t.Fatalf("MustGetDefaultClient() = %p, want %p", got, want)
	}
}

func recoveredAsError(recovered any) error {
	err, _ := recovered.(error)
	return err
}
