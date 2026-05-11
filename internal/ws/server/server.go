package server

import (
	"RTL-SDR/engine/internal/ws/protocol"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Server struct {
	Addr     string
	WsPath   string
	Handlers map[string]protocol.HandlerFunc

	upgrader  websocket.Upgrader
	clients   map[*Client]bool
	clientsMu sync.RWMutex
	broadcast chan []byte
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

func NewServer(addr, wsPath string) *Server {
	return &Server{
		Addr:     addr,
		WsPath:   wsPath,
		Handlers: make(map[string]protocol.HandlerFunc),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients:   make(map[*Client]bool),
		broadcast: make(chan []byte, 256),
	}
}

func (s *Server) Handle(typeName string, handler protocol.HandlerFunc) {
	s.Handlers[typeName] = handler
}

func (s *Server) Broadcast(message []byte) {
	s.broadcast <- message
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc(s.WsPath, s.handleWS)

	srv := &http.Server{
		Addr:    s.Addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
		close(s.broadcast)
	}()

	go s.broadcastLoop()

	log.Printf("Gateway WebSocket слушает: ws://%s%s", s.Addr, s.WsPath)
	return srv.ListenAndServe()
}

func (s *Server) broadcastLoop() {
	for msg := range s.broadcast {
		s.clientsMu.RLock()
		clients := make([]*Client, 0, len(s.clients))
		for client := range s.clients {
			clients = append(clients, client)
		}
		s.clientsMu.RUnlock()

		for _, client := range clients {
			select {
			case client.send <- msg:
			default:
				s.clientsMu.Lock()
				delete(s.clients, client)
				s.clientsMu.Unlock()
				close(client.send)
			}
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}

	s.clientsMu.Lock()
	s.clients[client] = true
	s.clientsMu.Unlock()

	go client.writePump()

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, client)
		s.clientsMu.Unlock()
		close(client.send)
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var env protocol.Envelope
		if err := json.Unmarshal(msg, &env); err != nil {
			s.sendError(client, "Неверный JSON")
			continue
		}

		if err := env.Validate(); err != nil {
			s.sendError(client, "Ошибка валидации: "+err.Error())
			continue
		}

		handler, ok := s.Handlers[env.Type]
		if !ok {
			s.sendError(client, "Неизвестный тип: "+env.Type)
			continue
		}

		result, err := handler(context.Background(), env.Payload)
		if err != nil {
			s.sendError(client, err.Error())
			continue
		}

		if result != nil {
			resp := map[string]interface{}{
				"type":    env.Type + "_response",
				"payload": result,
			}
			if data, err := json.Marshal(resp); err == nil {
				client.send <- data
			}
		}
	}
}

func (s *Server) sendError(client *Client, msg string) {
	errMsg := map[string]interface{}{
		"type":    "error",
		"payload": map[string]string{"message": msg},
	}
	data, _ := json.Marshal(errMsg)
	client.send <- data
}
