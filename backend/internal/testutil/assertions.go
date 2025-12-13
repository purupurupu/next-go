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
