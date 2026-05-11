package main

import (
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	if err := loadEnvFile(".env"); err != nil {
		log.Fatalf("Не удалось загрузить .env: %v", err)
	}

	webPort := getEnv("WEB_PORT", "8082")
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)
	log.Printf("Web сервер запущен на :%s", webPort)
	if err := http.ListenAndServe(":"+webPort, nil); err != nil {
		log.Fatal(err)
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
