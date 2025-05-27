package validators

import (
	"testing"
)

func TestRequiredFieldsValidator(t *testing.T) {
	v := NewRequiredFieldsValidator("foo", "bar")
	data := map[string]interface{}{"foo": 1, "bar": "x"}
	if err := v.Validate(data); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	missing := map[string]interface{}{"foo": 1}
	if err := v.Validate(missing); err == nil {
		t.Error("expected error for missing bar")
	}
}

func TestRequiredFieldsValidator_AllowEmpty(t *testing.T) {
	v := NewRequiredFieldsValidator("foo").WithAllowEmpty(true)
	data := map[string]interface{}{"foo": ""}
	if err := v.Validate(data); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRequiredStringFieldsValidator(t *testing.T) {
	v := NewRequiredStringFieldsValidator("foo").WithMinLength(2)
	ok := map[string]interface{}{"foo": "ab"}
	if err := v.Validate(ok); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	short := map[string]interface{}{"foo": "a"}
	if err := v.Validate(short); err == nil {
		t.Error("expected error for short string")
	}
}

func TestRequiredNestedFieldsValidator(t *testing.T) {
	v := NewRequiredNestedFieldsValidator("user.profile.name")
	ok := map[string]interface{}{"user": map[string]interface{}{"profile": map[string]interface{}{"name": "x"}}}
	if err := v.Validate(ok); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	missing := map[string]interface{}{"user": map[string]interface{}{"profile": map[string]interface{}{}}}
	if err := v.Validate(missing); err == nil {
		t.Error("expected error for missing nested field")
	}
}

func TestConditionalRequiredValidator(t *testing.T) {
	v := NewConditionalRequiredValidator().When(
		func(data map[string]interface{}) bool { return data["type"] == "A" },
		[]string{"foo"},
		"foo required if type is A",
	)
	ok := map[string]interface{}{"type": "A", "foo": 1}
	if err := v.Validate(ok); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	missing := map[string]interface{}{"type": "A"}
	if err := v.Validate(missing); err == nil {
		t.Error("expected error for missing foo")
	}
}

func TestEmailRequiredValidator(t *testing.T) {
	v := EmailRequiredValidator("email")
	ok := map[string]interface{}{"email": "a@b.com"}
	if err := v.Validate(ok); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bad := map[string]interface{}{"email": "abc"}
	if err := v.Validate(bad); err == nil {
		t.Error("expected error for invalid email")
	}
}

func TestIDRequiredValidator(t *testing.T) {
	v := IDRequiredValidator("id")
	ok := map[string]interface{}{"id": 123}
	if err := v.Validate(ok); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	bad := map[string]interface{}{"id": ""}
	if err := v.Validate(bad); err == nil {
		t.Error("expected error for empty id")
	}
}