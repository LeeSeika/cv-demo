package jsonx

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestNewBool(t *testing.T) {
	b := NewBool(true)
	if b == nil {
		t.Fatal("expected non-nil value")
	}
	if b.Result().Type != gjson.True {
		t.Fatalf("expected true, got %v", b.Result().Type)
	}

	b = NewBool(false)
	if b == nil {
		t.Fatal("expected non-nil value")
	}
	if b.Result().Type != gjson.False {
		t.Fatalf("expected false, got %v", b.Result().Type)
	}
}

func TestNewString(t *testing.T) {
	s, err := NewString("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil value")
	}
	if !s.IsString() {
		t.Fatalf("expected string, got %v", s.Result().Type)
	}
	if s.String() != "hello" {
		t.Fatalf("expected 'hello', got '%s'", s.Result().Str)
	}

	s, err = NewString("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil value")
	}
	if !s.IsString() {
		t.Fatalf("expected string, got %v", s.Result().Type)
	}
	if s.String() != "" {
		t.Fatalf("expected empty string, got '%s'", s.Result().Str)
	}

	s, err = NewString("key:\"value\"")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil value")
	}
	if !s.IsString() {
		t.Fatalf("expected string, got %v", s.Result().Type)
	}
	if s.String() != "key:\"value\"" {
		t.Fatalf("expected 'key:\"value\"', got '%s'", s.Result().Str)
	}
}

func TestNewNumber(t *testing.T) {
	n := NewNumber(int64(42))
	if n == nil {
		t.Fatal("expected non-nil value")
	}
	if n.Result().Type != gjson.Number {
		t.Fatalf("expected number, got %v", n.Result().Type)
	}
	if n.Result().Int() != 42 {
		t.Fatalf("expected 42, got %d", n.Result().Int())
	}

	n = NewNumber(int64(-1))
	if n == nil {
		t.Fatal("expected non-nil value")
	}
	if n.Result().Type != gjson.Number {
		t.Fatalf("expected number, got %v", n.Result().Type)
	}
	if n.Result().Int() != -1 {
		t.Fatalf("expected -1, got %d", n.Result().Int())
	}

	n = NewNumber(float64(3.14))
	if n == nil {
		t.Fatal("expected non-nil value")
	}
	if n.Result().Type != gjson.Number {
		t.Fatalf("expected number, got %v", n.Result().Type)
	}
	if n.Result().Float() != 3.14 {
		t.Fatalf("expected 3.14, got %f", n.Result().Float())
	}

	n = NewNumber(float64(-3.14151719))
	if n == nil {
		t.Fatal("expected non-nil value")
	}
	if n.Result().Type != gjson.Number {
		t.Fatalf("expected number, got %v", n.Result().Type)
	}
	if n.Result().Float() != -3.14151719 {
		t.Fatalf("expected -3.14151719, got %f", n.Result().Float())
	}
}
