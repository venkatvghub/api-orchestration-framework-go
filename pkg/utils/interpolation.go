package utils

import (
	"fmt"
	"strings"

	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
)

// InterpolateString replaces ${variable} patterns in template with values from context
func InterpolateString(template string, ctx interfaces.ExecutionContext) (string, error) {
	result := template
	// Find all ${...} patterns
	for {
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "}")
		if end == -1 {
			return "", fmt.Errorf("unclosed variable reference in template: %s", template)
		}
		end += start

		varName := result[start+2 : end]
		varValue := ""

		if strings.Contains(varName, ".") {
			varValue = GetNestedValueFromContext(varName, ctx)
		} else if val, ok := ctx.Get(varName); ok {
			varValue = fmt.Sprintf("%v", val)
		} else {
			// Keep original placeholder if variable not found
			varValue = "${" + varName + "}"
		}

		result = result[:start] + varValue + result[end+1:]
	}
	return result, nil
}

// InterpolateMap applies string interpolation to all string values in a map
func InterpolateMap(data map[string]string, ctx interfaces.ExecutionContext) (map[string]string, error) {
	result := make(map[string]string)
	for key, value := range data {
		interpolatedValue, err := InterpolateString(value, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to interpolate key %s: %w", key, err)
		}
		result[key] = interpolatedValue
	}
	return result, nil
}

// HasVariables checks if a string contains variable placeholders
func HasVariables(template string) bool {
	return strings.Contains(template, "${") && strings.Contains(template, "}")
}

// ExtractVariables returns all variable names found in a template
func ExtractVariables(template string) []string {
	var variables []string
	text := template

	for {
		start := strings.Index(text, "${")
		if start == -1 {
			break
		}
		end := strings.Index(text[start:], "}")
		if end == -1 {
			break
		}
		end += start

		varName := text[start+2 : end]
		variables = append(variables, varName)
		text = text[end+1:]
	}

	return variables
}
