package validator

import (
	"strings"
	"testing"
)

type CommonEmbedded struct {
	Value string `validate:"required"`
}

type StructA struct {
	CommonEmbedded
	FieldA string `validate:"required,email"`
}

type StructB struct {
	CommonEmbedded
	FieldB string `validate:"required,numeric"`
}

func containsErrorForField(err error, fieldName string) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), fieldName)
}

func TestStructCacheCollision(t *testing.T) {
	v := New() // Initialize your validator

	// 1. Validate StructA (populates cache)
	a := StructA{
		CommonEmbedded: CommonEmbedded{Value: "test"},
		FieldA:         "not-an-email", // Should fail email validation
	}
	errA := v.Struct(a)
	if errA == nil {
		t.Error("Expected validation error for StructA (invalid email), but got nil")
	}

	// 2. Validate StructB (should NOT use StructA's cached field layout/rules)
	b := StructB{
		CommonEmbedded: CommonEmbedded{Value: "test"},
		FieldB:         "abc", // Should fail numeric validation
	}
	errB := v.Struct(b)
	if errB == nil {
		t.Error("Expected validation error for StructB (invalid numeric), but got nil")
	}

	// Verify that the error messages point to the correct fields (FieldB, not FieldA)
	if errB != nil && !containsErrorForField(errB, "FieldB") {
		t.Errorf("Expected validation error on 'FieldB', but got: %v", errB)
	}
}
