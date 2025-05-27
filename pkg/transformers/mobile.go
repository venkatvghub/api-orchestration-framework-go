package transformers

import (
	"time"
)

// NewMobileTransformer creates a transformer optimized for mobile responses
// This is now a configured instance of the generic FieldTransformer
func NewMobileTransformer(fields []string) Transformer {
	return NewFieldTransformer("mobile", fields).
		WithMeta(true).
		WithFlatten(false).
		WithSeparator("_")
}

// NewMobileFlattenTransformer creates a mobile transformer that flattens nested structures
func NewMobileFlattenTransformer(prefix string) Transformer {
	return NewFlattenTransformer("mobile_flatten", prefix).
		WithSeparator("_").
		WithMeta(true).
		WithMaxDepth(3) // Limit depth for mobile optimization
}

// NewMobileResponseTransformer creates a comprehensive mobile response transformer
func NewMobileResponseTransformer(fields []string) Transformer {
	return NewTransformerChain("mobile_response").
		Add(NewFuncTransformer("add_mobile_metadata", func(data map[string]interface{}) (map[string]interface{}, error) {
			// Add mobile-specific metadata
			if data["_meta"] == nil {
				data["_meta"] = make(map[string]interface{})
			}

			if meta, ok := data["_meta"].(map[string]interface{}); ok {
				meta["timestamp"] = time.Now().Unix()
				meta["version"] = "1.0"
				meta["mobile_optimized"] = true
			}

			return data, nil
		})).
		Add(NewMobileTransformer(fields)).
		Add(ExcludeFieldsTransformer("_debug", "_internal", "_temp"))
}

// NewMobileListTransformer creates a transformer for mobile list responses
func NewMobileListTransformer(itemFields []string) Transformer {
	return NewFuncTransformer("mobile_list", func(data map[string]interface{}) (map[string]interface{}, error) {
		result := make(map[string]interface{})

		// Transform list items if present
		if items, ok := data["items"].([]interface{}); ok {
			transformedItems := make([]interface{}, len(items))
			itemTransformer := NewMobileTransformer(itemFields)

			for i, item := range items {
				if itemMap, ok := item.(map[string]interface{}); ok {
					transformed, err := itemTransformer.Transform(itemMap)
					if err != nil {
						return nil, err
					}
					transformedItems[i] = transformed
				} else {
					transformedItems[i] = item
				}
			}
			result["items"] = transformedItems
		}

		// Copy pagination and metadata
		for key, value := range data {
			if key != "items" {
				result[key] = value
			}
		}

		// Add mobile list metadata
		if result["_meta"] == nil {
			result["_meta"] = make(map[string]interface{})
		}

		if meta, ok := result["_meta"].(map[string]interface{}); ok {
			meta["mobile_list"] = true
			if items, ok := result["items"].([]interface{}); ok {
				meta["item_count"] = len(items)
			}
		}

		return result, nil
	})
}

// NewMobileErrorTransformer creates a transformer for mobile error responses
func NewMobileErrorTransformer() Transformer {
	return NewFuncTransformer("mobile_error", func(data map[string]interface{}) (map[string]interface{}, error) {
		result := map[string]interface{}{
			"success":   false,
			"timestamp": time.Now().Unix(),
			"mobile":    true,
		}

		// Extract error information
		if errorMsg, ok := data["error"]; ok {
			result["error"] = errorMsg
		} else if message, ok := data["message"]; ok {
			result["error"] = message
		} else {
			result["error"] = "An error occurred"
		}

		// Extract error code if present
		if code, ok := data["code"]; ok {
			result["code"] = code
		}

		// Extract details if present
		if details, ok := data["details"]; ok {
			result["details"] = details
		}

		return result, nil
	})
}

// NewMobilePaginationTransformer creates a transformer for mobile pagination
func NewMobilePaginationTransformer() Transformer {
	return NewFuncTransformer("mobile_pagination", func(data map[string]interface{}) (map[string]interface{}, error) {
		result := make(map[string]interface{})

		// Copy all data
		for key, value := range data {
			result[key] = value
		}

		// Enhance pagination for mobile
		pagination := make(map[string]interface{})

		if page, ok := data["page"]; ok {
			pagination["current_page"] = page
		}
		if totalPages, ok := data["total_pages"]; ok {
			pagination["total_pages"] = totalPages
		}
		if total, ok := data["total"]; ok {
			pagination["total_items"] = total
		}
		if limit, ok := data["limit"]; ok {
			pagination["page_size"] = limit
		}

		// Add mobile-specific pagination info
		if currentPage, ok := pagination["current_page"].(int); ok {
			if totalPages, ok := pagination["total_pages"].(int); ok {
				pagination["has_next"] = currentPage < totalPages
				pagination["has_prev"] = currentPage > 1
			}
		}

		if len(pagination) > 0 {
			result["pagination"] = pagination
		}

		return result, nil
	})
}

// NewMobileImageTransformer creates a transformer for mobile image optimization
func NewMobileImageTransformer() Transformer {
	return NewFuncTransformer("mobile_images", func(data map[string]interface{}) (map[string]interface{}, error) {
		result := deepCopyMap(data)

		// Transform image URLs for mobile optimization
		transformImages(result)

		return result, nil
	})
}

// transformImages recursively transforms image URLs in the data
func transformImages(data map[string]interface{}) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			// Check if this looks like an image URL
			if isImageField(key) {
				data[key] = optimizeImageURL(v)
			}
		case map[string]interface{}:
			transformImages(v)
		case []interface{}:
			for _, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					transformImages(itemMap)
				}
			}
		}
	}
}

// isImageField checks if a field name suggests it contains an image URL
func isImageField(fieldName string) bool {
	imageFields := []string{"image", "avatar", "photo", "picture", "thumbnail", "icon"}
	for _, field := range imageFields {
		if fieldName == field ||
			fieldName == field+"_url" ||
			fieldName == field+"Url" {
			return true
		}
	}
	return false
}

// optimizeImageURL adds mobile optimization parameters to image URLs
func optimizeImageURL(url string) string {
	// This is a placeholder - in a real implementation, you would add
	// mobile-specific parameters like width, quality, format, etc.
	// For example: url + "?mobile=true&width=400&quality=80"
	return url
}

// Common mobile transformer configurations

// MobileUserProfileTransformer creates a transformer for user profile data
func MobileUserProfileTransformer() Transformer {
	return NewMobileTransformer([]string{
		"id", "username", "email", "name", "avatar",
		"created_at", "last_login", "verified",
	})
}

// MobileProductTransformer creates a transformer for product data
func MobileProductTransformer() Transformer {
	return NewTransformerChain("mobile_product").
		Add(NewMobileTransformer([]string{
			"id", "name", "price", "currency", "image",
			"description", "category", "in_stock", "rating",
		})).
		Add(NewMobileImageTransformer())
}

// MobileOrderTransformer creates a transformer for order data
func MobileOrderTransformer() Transformer {
	return NewMobileTransformer([]string{
		"id", "status", "total", "currency", "created_at",
		"items", "shipping_address", "payment_method",
	})
}
