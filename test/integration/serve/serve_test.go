package serve_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestServeCommand проверяет базовую функциональность команды serve
func TestServeCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Проверяем наличие OPENAI_API_KEY
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("OPENAI_API_KEY not set, skipping serve integration test")
	}

	// Создаем временную директорию для вывода
	tempDir, err := os.MkdirTemp("", "aiwf-serve-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Путь к тестовой конфигурации
	configPath := filepath.Join("..", "..", "..", "templates", "api-server", "config.yaml")

	// Путь к aiwf бинарнику
	aiwfBinary := filepath.Join("..", "..", "..", "aiwf")

	// Проверяем существование бинарника
	if _, err := os.Stat(aiwfBinary); os.IsNotExist(err) {
		t.Skipf("aiwf binary not found at %s, run 'go build -o aiwf ./cmd/aiwf' first", aiwfBinary)
	}

	// Запускаем сервер
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	port := "18080" // Используем нестандартный порт для тестов
	cmd := exec.CommandContext(ctx,
		aiwfBinary,
		"serve",
		"-f", configPath,
		"--output", tempDir,
		"--port", port,
		"--host", "127.0.0.1",
	)

	// Перенаправляем вывод для отладки
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Запускаем сервер в фоне
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Убиваем процесс при завершении теста
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// Ждем, пока сервер запустится
	baseURL := fmt.Sprintf("http://127.0.0.1:%s", port)
	if !waitForServer(baseURL+"/health", 30*time.Second) {
		t.Fatal("Server failed to start within timeout")
	}

	t.Logf("Server started successfully on %s", baseURL)

	// Тест 1: Health check
	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			t.Fatalf("Health check failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var result map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result["status"] != "ok" {
			t.Errorf("Expected status 'ok', got '%s'", result["status"])
		}
	})

	// Тест 2: Список агентов
	t.Run("ListAgents", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/agents")
		if err != nil {
			t.Fatalf("List agents failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		agents, ok := result["agents"].([]interface{})
		if !ok {
			t.Fatal("Expected 'agents' array in response")
		}

		if len(agents) != 4 {
			t.Errorf("Expected 4 agents, got %d", len(agents))
		}

		t.Logf("Found %d agents", len(agents))
	})

	// Тест 3: Вызов агента (text_analyzer)
	t.Run("CallTextAnalyzer", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"text":          "This is a great product!",
			"analysis_type": "sentiment",
		}

		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		resp, err := http.Post(
			baseURL+"/agent/text_analyzer",
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			t.Fatalf("Agent call failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Проверяем структуру ответа
		data, ok := result["data"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected 'data' object in response")
		}

		trace, ok := result["trace"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected 'trace' object in response")
		}

		// Проверяем поля в data
		if data["analysis_type"] == nil {
			t.Error("Expected 'analysis_type' in data")
		}

		if data["result"] == nil {
			t.Error("Expected 'result' in data")
		}

		// Проверяем trace
		usage, ok := trace["usage"].(map[string]interface{})
		if !ok {
			t.Error("Expected 'usage' in trace")
		} else {
			if usage["total"] == nil {
				t.Error("Expected 'total' tokens in usage")
			}
		}

		t.Logf("Agent response: %+v", data)
	})

	// Тест 4: Invalid request
	t.Run("InvalidRequest", func(t *testing.T) {
		// Отправляем невалидный JSON
		resp, err := http.Post(
			baseURL+"/agent/text_analyzer",
			"application/json",
			bytes.NewBufferString("{invalid json}"),
		)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Error("Expected non-200 status for invalid request")
		}
	})

	// Тест 5: Method not allowed
	t.Run("MethodNotAllowed", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/agent/text_analyzer")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", resp.StatusCode)
		}
	})
}

// TestServeEphemeralMode проверяет ephemeral mode (без --output)
func TestServeEphemeralMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	configPath := filepath.Join("..", "..", "..", "templates", "api-server", "config.yaml")
	aiwfBinary := filepath.Join("..", "..", "..", "aiwf")

	if _, err := os.Stat(aiwfBinary); os.IsNotExist(err) {
		t.Skip("aiwf binary not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	port := "18081"
	cmd := exec.CommandContext(ctx,
		aiwfBinary,
		"serve",
		"-f", configPath,
		"--port", port,
	)

	// Захватываем вывод
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// Ждем запуска
	baseURL := fmt.Sprintf("http://127.0.0.1:%s", port)
	if !waitForServer(baseURL+"/health", 30*time.Second) {
		t.Logf("Stdout: %s", stdout.String())
		t.Logf("Stderr: %s", stderr.String())
		t.Fatal("Server failed to start")
	}

	// Проверяем, что сервер работает
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	t.Log("Ephemeral mode server started successfully")
}

// waitForServer ждет, пока сервер станет доступен
func waitForServer(url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 1 * time.Second}

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	return false
}
