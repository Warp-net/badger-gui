package main

import (
	"encoding/json"
	"testing"

	"github.com/filinvadim/badger-gui/domain"
)

// TestMarshalString verifies that marshalString properly encodes strings as valid JSON
func TestMarshalString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "ok",
			expected: `"ok"`,
		},
		{
			name:     "error message",
			input:    "db isn't running",
			expected: `"db isn't running"`,
		},
		{
			name:     "message with special characters",
			input:    "invalid character 'o' looking for beginning of value",
			expected: `"invalid character 'o' looking for beginning of value"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := marshalString(tt.input)
			if string(result) != tt.expected {
				t.Errorf("marshalString(%q) = %q, want %q", tt.input, string(result), tt.expected)
			}

			// Verify that the result is valid JSON
			var decoded string
			if err := json.Unmarshal(result, &decoded); err != nil {
				t.Errorf("marshalString(%q) produced invalid JSON: %v", tt.input, err)
			}
			if decoded != tt.input {
				t.Errorf("After unmarshaling, got %q, want %q", decoded, tt.input)
			}
		})
	}
}

// MockStorer is a mock implementation of the Storer interface for testing
type MockStorer struct {
	running bool
	openErr error
	data    map[string][]byte
}

func (m *MockStorer) Open(dbPath, decryptKey, compression string) error {
	if m.openErr != nil {
		return m.openErr
	}
	m.running = true
	m.data = make(map[string][]byte)
	return nil
}

func (m *MockStorer) Set(key string, value []byte) error {
	m.data[key] = value
	return nil
}

func (m *MockStorer) Get(key string) ([]byte, error) {
	return m.data[key], nil
}

func (m *MockStorer) Delete(key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockStorer) List(prefix string, limit *int, cursor *string) (domain.Items, string, error) {
	return domain.Items{}, "", nil
}

func (m *MockStorer) IsRunning() bool {
	return m.running
}

func (m *MockStorer) Close() {
	m.running = false
}

// TestCallOpenSuccess tests that the Call method returns valid JSON for successful open operation
func TestCallOpenSuccess(t *testing.T) {
	mock := &MockStorer{}
	app := NewApp(mock)

	msgBody, _ := json.Marshal(MessageOpen{
		Path:          "/tmp/test",
		DecryptionKey: "",
		Compression:   "none",
		Delimiter:     "/",
	})

	msg := AppMessage{
		Type: TypeOpen,
		Body: msgBody,
	}

	response := app.Call(msg)

	if response.Type != TypeOpen {
		t.Errorf("Expected response type %q, got %q", TypeOpen, response.Type)
	}

	// Verify that the response body is valid JSON
	var responseText string
	if err := json.Unmarshal(response.Body, &responseText); err != nil {
		t.Fatalf("Response body is not valid JSON: %v, body: %s", err, string(response.Body))
	}

	if responseText != "ok" {
		t.Errorf("Expected response text 'ok', got %q", responseText)
	}
}

// TestCallOpenAlreadyRunning tests that the Call method returns valid JSON for already running error
func TestCallOpenAlreadyRunning(t *testing.T) {
	mock := &MockStorer{running: true}
	app := NewApp(mock)

	msgBody, _ := json.Marshal(MessageOpen{
		Path:          "/tmp/test",
		DecryptionKey: "",
		Compression:   "none",
		Delimiter:     "/",
	})

	msg := AppMessage{
		Type: TypeOpen,
		Body: msgBody,
	}

	response := app.Call(msg)

	// Verify that the response body is valid JSON
	var responseText string
	if err := json.Unmarshal(response.Body, &responseText); err != nil {
		t.Fatalf("Response body is not valid JSON: %v, body: %s", err, string(response.Body))
	}

	if responseText != "already running" {
		t.Errorf("Expected response text 'already running', got %q", responseText)
	}
}

// TestCallSetSuccess tests that the Call method returns valid JSON for successful set operation
func TestCallSetSuccess(t *testing.T) {
	mock := &MockStorer{running: true, data: make(map[string][]byte)}
	app := NewApp(mock)

	msgBody, _ := json.Marshal(MessageSet{
		Key:   "testkey",
		Value: json.RawMessage(`"testvalue"`),
	})

	msg := AppMessage{
		Type: TypeSet,
		Body: msgBody,
	}

	response := app.Call(msg)

	// Verify that the response body is valid JSON
	var responseText string
	if err := json.Unmarshal(response.Body, &responseText); err != nil {
		t.Fatalf("Response body is not valid JSON: %v, body: %s", err, string(response.Body))
	}

	if responseText != "ok" {
		t.Errorf("Expected response text 'ok', got %q", responseText)
	}
}

// TestAppMessageMarshaling tests that AppMessage can be properly marshaled and unmarshaled
func TestAppMessageMarshaling(t *testing.T) {
	msg := AppMessage{
		Type: TypeOpen,
		Body: marshalString("ok"),
	}

	// Marshal the AppMessage
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal AppMessage: %v", err)
	}

	// Unmarshal it back
	var decoded AppMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal AppMessage: %v, data: %s", err, string(data))
	}

	if decoded.Type != msg.Type {
		t.Errorf("Expected type %q, got %q", msg.Type, decoded.Type)
	}

	var bodyText string
	if err := json.Unmarshal(decoded.Body, &bodyText); err != nil {
		t.Fatalf("Failed to unmarshal body: %v, body: %s", err, string(decoded.Body))
	}

	if bodyText != "ok" {
		t.Errorf("Expected body text 'ok', got %q", bodyText)
	}
}
