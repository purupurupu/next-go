package testutil

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// JSONResponse parses the response body as JSON and returns it as a map
func JSONResponse(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	var response map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	return response
}

// ExtractStatusCode extracts the status code from a standard API response
func ExtractStatusCode(response map[string]any) int {
	status, ok := response["status"].(map[string]any)
	if !ok {
		return 0
	}
	code, ok := status["code"].(float64)
	if !ok {
		return 0
	}
	return int(code)
}

// ExtractMessage extracts the message from a standard API response
func ExtractMessage(response map[string]any) string {
	status, ok := response["status"].(map[string]any)
	if !ok {
		return ""
	}
	message, ok := status["message"].(string)
	if !ok {
		return ""
	}
	return message
}

// ExtractData extracts the data object from a standard API response
func ExtractData(response map[string]any) map[string]any {
	data, ok := response["data"].(map[string]any)
	if !ok {
		return nil
	}
	return data
}

// ExtractTodo extracts a todo from the response (direct "todo" key)
func ExtractTodo(response map[string]any) map[string]any {
	todo, ok := response["todo"].(map[string]any)
	if !ok {
		return nil
	}
	return todo
}

// ExtractTodoFromData extracts a todo from the data object (nested "data.todo")
func ExtractTodoFromData(response map[string]any) map[string]any {
	data := ExtractData(response)
	if data == nil {
		return nil
	}
	todo, ok := data["todo"].(map[string]any)
	if !ok {
		return nil
	}
	return todo
}

// ExtractTodos extracts the todos array from the response
func ExtractTodos(response map[string]any) []any {
	todos, ok := response["todos"].([]any)
	if !ok {
		return nil
	}
	return todos
}

// TodoAt returns the todo at the given index as a map
func TodoAt(todos []any, index int) map[string]any {
	if index >= len(todos) {
		return nil
	}
	todo, ok := todos[index].(map[string]any)
	if !ok {
		return nil
	}
	return todo
}

// ExtractCategory extracts a category from the response (direct "category" key)
func ExtractCategory(response map[string]any) map[string]any {
	category, ok := response["category"].(map[string]any)
	if !ok {
		return nil
	}
	return category
}

// ExtractCategoryFromData extracts a category from the data object
func ExtractCategoryFromData(response map[string]any) map[string]any {
	data := ExtractData(response)
	if data == nil {
		return nil
	}
	category, ok := data["category"].(map[string]any)
	if !ok {
		return nil
	}
	return category
}

// ExtractCategories extracts the categories array from the response
func ExtractCategories(response map[string]any) []any {
	categories, ok := response["categories"].([]any)
	if !ok {
		return nil
	}
	return categories
}

// CategoryAt returns the category at the given index as a map
func CategoryAt(categories []any, index int) map[string]any {
	if index >= len(categories) {
		return nil
	}
	category, ok := categories[index].(map[string]any)
	if !ok {
		return nil
	}
	return category
}

// ExtractTag extracts a tag from the response (direct "tag" key)
func ExtractTag(response map[string]any) map[string]any {
	tag, ok := response["tag"].(map[string]any)
	if !ok {
		return nil
	}
	return tag
}

// ExtractTagFromData extracts a tag from the data object
func ExtractTagFromData(response map[string]any) map[string]any {
	data := ExtractData(response)
	if data == nil {
		return nil
	}
	tag, ok := data["tag"].(map[string]any)
	if !ok {
		return nil
	}
	return tag
}

// ExtractTags extracts the tags array from the response
func ExtractTags(response map[string]any) []any {
	tags, ok := response["tags"].([]any)
	if !ok {
		return nil
	}
	return tags
}

// TagAt returns the tag at the given index as a map
func TagAt(tags []any, index int) map[string]any {
	if index >= len(tags) {
		return nil
	}
	tag, ok := tags[index].(map[string]any)
	if !ok {
		return nil
	}
	return tag
}
