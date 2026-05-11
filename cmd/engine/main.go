package main

import (
	"log"
	"os"
	"os/exec"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	services := []string{
		"./cmd/storage",
		"./cmd/gateway",
		"./cmd/web",
	}

	for _, svc := range services {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			cmd := exec.Command("go", "run", path)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Printf("Ошибка в %s: %v", path, err)
			}
		}(svc)
	}

	wg.Wait()
}
