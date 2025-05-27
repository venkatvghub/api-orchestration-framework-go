package validators

import (
	"testing"
)

func TestNoOpValidator(t *testing.T) {
	v := NewNoOpValidator()
	if err := v.Validate(map[string]interface{}{}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAlwaysFailValidator(t *testing.T) {
	v := NewAlwaysFailValidator("fail")
	if err := v.Validate(map[string]interface{}{}); err == nil {
		t.Error("expected error")
	}
}

func TestIsEmpty(t *testing.T) {
	if !IsEmpty(nil) || !IsEmpty(0) || !IsEmpty(0.0) || !IsEmpty("") {
		t.Error("expected empty values to be empty")
	}
	if IsEmpty("x") || IsEmpty(1) {
		t.Error("expected non-empty values to not be empty")
	}
}

func TestGetFieldValue(t *testing.T) {
	m := map[string]interface{}{"foo": 1}
	v, ok := GetFieldValue(m, "foo")
	if !ok || v != 1 {
		t.Error("expected to get field value")
	}
}

func TestGetStringField(t *testing.T) {
	m := map[string]interface{}{"foo": "bar"}
	v, err := GetStringField(m, "foo")
	if err != nil || v != "bar" {
		t.Error("expected to get string field")
	}
	_, err = GetStringField(m, "baz")
	if err == nil {
		t.Error("expected error for missing field")
	}
}

func TestGetIntField(t *testing.T) {
	m := map[string]interface{}{"foo": 1}
	v, err := GetIntField(m, "foo")
	if err != nil || v != 1 {
		t.Error("expected to get int field")
	}
}

func TestGetBoolField(t *testing.T) {
	m := map[string]interface{}{"foo": true}
	v, err := GetBoolField(m, "foo")
	if err != nil || !v {
		t.Error("expected to get bool field")
	}
}