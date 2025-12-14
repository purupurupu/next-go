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
	return ExtractObject(response, "data")
}

// --- Generic helper functions ---

// ExtractObject extracts an object by key from the response
func ExtractObject(response map[string]any, key string) map[string]any {
	obj, ok := response[key].(map[string]any)
	if !ok {
		return nil
	}
	return obj
}

// ExtractObjectOrFail extracts an object by key from the response and fails the test if not found
func ExtractObjectOrFail(t *testing.T, response map[string]any, key string) map[string]any {
	t.Helper()
	obj := ExtractObject(response, key)
	require.NotNil(t, obj, "expected key '%s' to be present in response", key)
	return obj
}

// ExtractObjectFromData extracts an object by key from the data object
func ExtractObjectFromData(response map[string]any, key string) map[string]any {
	data := ExtractData(response)
	if data == nil {
		return nil
	}
	return ExtractObject(data, key)
}

// ExtractArray extracts an array by key from the response
func ExtractArray(response map[string]any, key string) []any {
	arr, ok := response[key].([]any)
	if !ok {
		return nil
	}
	return arr
}

// ExtractArrayOrFail extracts an array by key from the response and fails the test if not found
func ExtractArrayOrFail(t *testing.T, response map[string]any, key string) []any {
	t.Helper()
	arr := ExtractArray(response, key)
	require.NotNil(t, arr, "expected key '%s' to be present in response", key)
	return arr
}

// ItemAt returns the item at the given index as a map
func ItemAt(items []any, index int) map[string]any {
	if index >= len(items) {
		return nil
	}
	item, ok := items[index].(map[string]any)
	if !ok {
		return nil
	}
	return item
}

// --- Todo helpers (using generic functions) ---

// ExtractTodo extracts a todo from the response
func ExtractTodo(response map[string]any) map[string]any {
	return ExtractObject(response, "todo")
}

// ExtractTodoFromData extracts a todo from the data object
func ExtractTodoFromData(response map[string]any) map[string]any {
	return ExtractObjectFromData(response, "todo")
}

// ExtractTodos extracts the todos array from the response
func ExtractTodos(response map[string]any) []any {
	return ExtractArray(response, "todos")
}

// TodoAt returns the todo at the given index
func TodoAt(todos []any, index int) map[string]any {
	return ItemAt(todos, index)
}

// --- Category helpers (using generic functions) ---

// ExtractCategory extracts a category from the response
func ExtractCategory(response map[string]any) map[string]any {
	return ExtractObject(response, "category")
}

// ExtractCategoryFromData extracts a category from the data object
func ExtractCategoryFromData(response map[string]any) map[string]any {
	return ExtractObjectFromData(response, "category")
}

// ExtractCategories extracts the categories array from the response
func ExtractCategories(response map[string]any) []any {
	return ExtractArray(response, "categories")
}

// CategoryAt returns the category at the given index
func CategoryAt(categories []any, index int) map[string]any {
	return ItemAt(categories, index)
}

// --- Tag helpers (using generic functions) ---

// ExtractTag extracts a tag from the response
func ExtractTag(response map[string]any) map[string]any {
	return ExtractObject(response, "tag")
}

// ExtractTagFromData extracts a tag from the data object
func ExtractTagFromData(response map[string]any) map[string]any {
	return ExtractObjectFromData(response, "tag")
}

// ExtractTags extracts the tags array from the response
func ExtractTags(response map[string]any) []any {
	return ExtractArray(response, "tags")
}

// TagAt returns the tag at the given index
func TagAt(tags []any, index int) map[string]any {
	return ItemAt(tags, index)
}
