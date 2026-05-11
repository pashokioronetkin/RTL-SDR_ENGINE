package handlers

import (
	"RTL-SDR/engine/internal/storage"
	"RTL-SDR/engine/internal/ws/protocol"
	"context"
	"encoding/json"
)

// HandleGetHistory обрабатывает команду get_history – получает историю через storage API
func HandleGetHistory(storageClient *storage.Client) protocol.HandlerFunc {
	return func(ctx context.Context, payload json.RawMessage) (interface{}, error) {
		var args struct {
			Limit int `json:"limit"`
		}
		if err := json.Unmarshal(payload, &args); err != nil {
			args.Limit = 20
		}
		if args.Limit <= 0 || args.Limit > 100 {
			args.Limit = 20
		}
		sessions, err := storageClient.GetHistory(args.Limit)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"history": sessions,
			"count":   len(sessions),
		}, nil
	}
}
