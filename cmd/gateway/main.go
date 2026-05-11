package main

import (
	"RTL-SDR/engine/internal/handlers"
	"RTL-SDR/engine/internal/storage"
	"RTL-SDR/engine/internal/ws/server"
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	if err := loadEnvFile(".env"); err != nil {
		log.Fatalf("Не удалось загрузить .env: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Клиент для storage API
	storageClient := storage.NewClient(getEnv("STORAGE_API_URL", "http://localhost:8081"))

	// WebSocket-сервер
	wsAddr := getEnv("WS_ADDR", "0.0.0.0:8080")
	srv := server.NewServer(
		wsAddr,
		getEnv("WS_PATH", "/ws"),
	)

	// Регистрируем обработчики с передачей storageClient
	srv.Handle("detection_data", handlers.HandleDetection(storageClient, srv))
	srv.Handle("command", handlers.HandleCommand(storageClient))
	srv.Handle("get_history", handlers.HandleGetHistory(storageClient))

	log.Printf("Gateway запущен на %s", wsAddr)
	if err := srv.ListenAndServe(ctx); err != nil {
		log.Printf("Сервер остановлен: %v", err)
	}
}

func loadEnvFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		if key == "" {
			continue
		}

		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
