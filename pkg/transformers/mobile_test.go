package transformers

import (
	"testing"
)

func TestNewMobileTransformer(t *testing.T) {
	fields := []string{"id", "name", "avatar"}
	tr := NewMobileTransformer(fields)
	input := map[string]interface{}{
		"id":     1,
		"name":   "Alice",
		"avatar": "img.png",
		"extra":  42,
		"_meta":  map[string]interface{}{"source": "test"},
	}
	out, err := tr.Transform(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := out["_meta"]; !ok {
		t.Error("expected _meta field")
	}
	if _, ok := out["extra"]; ok {
		t.Error("unexpected field 'extra'")
	}
}

func TestNewMobileFlattenTransformer(t *testing.T) {
	tr := NewMobileFlattenTransformer("pre")
	input := map[string]interface{}{"foo": map[string]interface{}{"bar": 1}}
	out, err := tr.Transform(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for k := range out {
		if k == "pre_foo_bar" {
			found = true
		}
	}
	if !found {
		t.Error("flattened key not found")
	}
}

func TestNewMobileResponseTransformer(t *testing.T) {
	fields := []string{"id", "name"}
	tr := NewMobileResponseTransformer(fields)
	input := map[string]interface{}{"id": 1, "name": "Bob", "_debug": 123, "_internal": 456, "_temp": 789}
	out, err := tr.Transform(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := out["_debug"]; ok {
		t.Error("_debug should be excluded")
	}
	meta, ok := out["_meta"].(map[string]interface{})
	if !ok || !meta["mobile_optimized"].(bool) {
		t.Error("mobile_optimized meta missing or not true")
	}
}

func TestNewMobileListTransformer(t *testing.T) {
	tr := NewMobileListTransformer([]string{"id", "name"})
	input := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"id": 1, "name": "A", "extra": 9},
			map[string]interface{}{"id": 2, "name": "B"},
		},
		"page": 1,
	}
	out, err := tr.Transform(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items, ok := out["items"].([]interface{})
	if !ok || len(items) != 2 {
		t.Error("items not transformed correctly")
	}
	meta, ok := out["_meta"].(map[string]interface{})
	if !ok || !meta["mobile_list"].(bool) {
		t.Error("mobile_list meta missing or not true")
	}
	if meta["item_count"] != 2 {
		t.Error("item_count meta incorrect")
	}
}

func TestNewMobileErrorTransformer(t *testing.T) {
	tr := NewMobileErrorTransformer()
	input := map[string]interface{}{"error": "fail", "code": 404, "details": "not found"}
	out, err := tr.Transform(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["success"] != false {
		t.Error("success should be false")
	}
	if out["error"] != "fail" {
		t.Error("error field incorrect")
	}
	if out["code"] != 404 {
		t.Error("code field incorrect")
	}
	if out["details"] != "not found" {
		t.Error("details field incorrect")
	}
	if _, ok := out["timestamp"]; !ok {
		t.Error("timestamp missing")
	}
}

func TestNewMobilePaginationTransformer(t *testing.T) {
	tr := NewMobilePaginationTransformer()
	input := map[string]interface{}{"page": 2, "total_pages": 5, "total": 100, "limit": 20}
	out, err := tr.Transform(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pagination, ok := out["pagination"].(map[string]interface{})
	if !ok {
		t.Fatal("pagination missing")
	}
	if pagination["current_page"] != 2 || pagination["total_pages"] != 5 {
		t.Error("pagination fields incorrect")
	}
	if pagination["has_next"] != true || pagination["has_prev"] != true {
		t.Error("has_next/has_prev incorrect")
	}
}

func TestNewMobileImageTransformer(t *testing.T) {
	tr := NewMobileImageTransformer()
	input := map[string]interface{}{"image": "img.jpg", "nested": map[string]interface{}{"avatar_url": "a.png"}}
	out, err := tr.Transform(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["image"] != "img.jpg" {
		t.Error("image not optimized (should be unchanged in placeholder)")
	}
	nested := out["nested"].(map[string]interface{})
	if nested["avatar_url"] != "a.png" {
		t.Error("nested image not optimized (should be unchanged in placeholder)")
	}
}

func TestMobileUserProfileTransformer(t *testing.T) {
	tr := MobileUserProfileTransformer()
	input := map[string]interface{}{"id": 1, "username": "u", "avatar": "a.png", "extra": 99}
	out, err := tr.Transform(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := out["extra"]; ok {
		t.Error("unexpected field 'extra'")
	}
}

func TestMobileProductTransformer(t *testing.T) {
	tr := MobileProductTransformer()
	input := map[string]interface{}{"id": 1, "name": "p", "image": "img.jpg", "extra": 99}
	out, err := tr.Transform(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := out["extra"]; ok {
		t.Error("unexpected field 'extra'")
	}
}

func TestMobileOrderTransformer(t *testing.T) {
	tr := MobileOrderTransformer()
	input := map[string]interface{}{"id": 1, "status": "ok", "items": []interface{}{1, 2, 3}, "extra": 99}
	out, err := tr.Transform(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := out["extra"]; ok {
		t.Error("unexpected field 'extra'")
	}
}

func TestMobileTransformers_EdgeCases(t *testing.T) {
	tr := NewMobileTransformer([]string{"id"})
	_, err := tr.Transform(nil)
	if err == nil {
		t.Error("expected error for nil input")
	}
	tr2 := NewMobileListTransformer([]string{"id"})
	_, err = tr2.Transform(nil)
	if err == nil {
		t.Error("expected error for nil input")
	}
}
