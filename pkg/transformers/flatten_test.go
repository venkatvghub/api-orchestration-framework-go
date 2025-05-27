package transformers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlattenTransformer(t *testing.T) {
	data := map[string]interface{}{
		"id": 1,
		"info": map[string]interface{}{
			"name": "foo",
			"meta": map[string]interface{}{"v": 2},
		},
		"arr": []interface{}{
			map[string]interface{}{"x": 10},
			20,
		},
		"_meta": map[string]interface{}{"source": "test"},
	}

	t.Run("Basic flatten", func(t *testing.T) {
		ft := NewFlattenTransformer("basic", "")
		result, err := ft.Transform(data)
		assert.NoError(t, err)
		assert.Equal(t, 1, result["id"])
		assert.Equal(t, "foo", result["info.name"])
		assert.Equal(t, 2, result["info.meta.v"])
		assert.Equal(t, 10, result["arr.0.x"])
		assert.Equal(t, 20, result["arr.1"])
		assert.Equal(t, "test", result["_meta.source"])
	})

	t.Run("With prefix and separator", func(t *testing.T) {
		ft := NewFlattenTransformer("prefix", "pfx").WithSeparator(":")
		result, err := ft.Transform(data)
		assert.NoError(t, err)
		assert.Contains(t, result, "pfx:id")
		assert.Contains(t, result, "pfx:info:name")
		assert.Contains(t, result, "pfx:arr:0:x")
	})

	t.Run("MaxDepth", func(t *testing.T) {
		ft := NewFlattenTransformer("maxdepth", "").WithMaxDepth(1)
		result, err := ft.Transform(data)
		assert.NoError(t, err)
		assert.Contains(t, result, "info")
		assert.Contains(t, result, "arr.0")
		assert.Contains(t, result, "arr.1")
	})

	t.Run("Exclude meta fields", func(t *testing.T) {
		ft := NewFlattenTransformer("excludeMeta", "").WithMeta(false)
		result, err := ft.Transform(data)
		assert.NoError(t, err)
		for key := range result {
			assert.False(t, strings.HasPrefix(key, "_meta"), "Should not contain _meta fields")
		}
	})

	t.Run("Nil input", func(t *testing.T) {
		ft := NewFlattenTransformer("nil", "")
		_, err := ft.Transform(nil)
		assert.Error(t, err)
	})
}

func TestUnflattenTransformer(t *testing.T) {
	flat := map[string]interface{}{
		"a.b": 1,
		"a.c": 2,
		"x":   3,
	}
	ut := NewUnflattenTransformer("unflat")
	result, err := ut.Transform(flat)
	assert.NoError(t, err)
	assert.Contains(t, result, "a")
	assert.Contains(t, result, "x")
	if nested, ok := result["a"].(map[string]interface{}); ok {
		assert.Equal(t, 1, nested["b"])
		assert.Equal(t, 2, nested["c"])
	}
}

func TestPrefixTransformer(t *testing.T) {
	data := map[string]interface{}{"foo": 1, "bar": 2}
	pt := NewPrefixTransformer("pt", "pre").WithSeparator("-")
	result, err := pt.Transform(data)
	assert.NoError(t, err)
	assert.Equal(t, 1, result["pre-foo"])
	assert.Equal(t, 2, result["pre-bar"])
}

func TestRemovePrefixTransformer(t *testing.T) {
	data := map[string]interface{}{"pre_foo": 1, "pre_bar": 2, "baz": 3}
	rpt := NewRemovePrefixTransformer("rpt", "pre")
	result, err := rpt.Transform(data)
	assert.NoError(t, err)
	assert.Equal(t, 1, result["foo"])
	assert.Equal(t, 2, result["bar"])
	assert.Equal(t, 3, result["baz"])
}
