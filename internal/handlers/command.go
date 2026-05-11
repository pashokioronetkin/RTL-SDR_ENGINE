package handlers

import (
	"RTL-SDR/engine/internal/storage"
	"RTL-SDR/engine/internal/ws/protocol"
	"context"
	"encoding/json"
	"log"
)

type CommandPayload struct {
	Command string          `json:"command"`
	Args    json.RawMessage `json:"args"`
}

type CommandResponse struct {
	Status  string      `json:"status"`
	Result  interface{} `json:"result,omitempty"`
	Message string      `json:"message,omitempty"`
}

func HandleCommand(storageClient *storage.Client) protocol.HandlerFunc {
	return func(ctx context.Context, payload json.RawMessage) (interface{}, error) {
		var cmd CommandPayload
		if err := json.Unmarshal(payload, &cmd); err != nil {
			return nil, err
		}
		log.Printf("Получена команда: %s", cmd.Command)
		switch cmd.Command {
		case "ping":
			return CommandResponse{Status: "ok", Result: "pong"}, nil
		default:
			return CommandResponse{Status: "error", Message: "unknown command"}, nil
		}
	}
}
