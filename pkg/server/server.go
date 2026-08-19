package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"botwa_go_ascii/config"
	"botwa_go_ascii/pkg/asciiapi"
	"botwa_go_ascii/pkg/bot"

	"go.mau.fi/whatsmeow/types"
)

type Server struct {
	cfg       *config.Config
	apiClient *asciiapi.Client
	botRouter *bot.Router
	httpSrv   *http.Server
}

type SendMessageRequest struct {
	To      string `json:"to"`      // Phone number (e.g. 628123456789 or 628123456789@s.whatsapp.net) or group JID
	Message string `json:"message"` // Message content
}

type BroadcastRequest struct {
	Recipients []string `json:"recipients"` // List of numbers or group JIDs
	Message    string   `json:"message"`
}

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewServer(cfg *config.Config, apiClient *asciiapi.Client, botRouter *bot.Router) *Server {
	return &Server{
		cfg:       cfg,
		apiClient: apiClient,
		botRouter: botRouter,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Public Health Check
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/v1/health", s.handleHealth)

	// Protected Endpoints
	mux.HandleFunc("/api/v1/send-message", s.authMiddleware(s.handleSendMessage))
	mux.HandleFunc("/api/v1/broadcast", s.authMiddleware(s.handleBroadcast))

	addr := fmt.Sprintf(":%d", s.cfg.ServerPort)
	s.httpSrv = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("[Server] REST Webhook Server listening on http://localhost:%d", s.cfg.ServerPort)
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[Server] HTTP server error: %v", err)
		}
	}()

	return nil
}

func (s *Server) Stop() {
	if s.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.httpSrv.Shutdown(ctx)
		log.Println("[Server] REST Webhook Server stopped gracefully.")
	}
}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Verify secret token
		secretHeader := r.Header.Get("X-Bot-Secret")
		authHeader := r.Header.Get("Authorization")

		authorized := false
		if secretHeader != "" && secretHeader == s.cfg.BotSecret {
			authorized = true
		} else if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" && parts[1] == s.cfg.BotSecret {
				authorized = true
			}
		}

		if !authorized {
			sendJSON(w, http.StatusUnauthorized, APIResponse{
				Success: false,
				Error:   "Unauthorized: Invalid X-Bot-Secret or Authorization Bearer token",
			})
			return
		}

		next(w, r)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	apiLatency, apiErr := s.apiClient.Ping()

	status := map[string]interface{}{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
		"ascii_web": map[string]interface{}{
			"url":     s.apiClient.WebURL(),
			"online":  apiErr == nil,
			"latency": apiLatency.String(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "Method not allowed"})
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid JSON body: " + err.Error()})
		return
	}

	if req.To == "" || req.Message == "" {
		sendJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "'to' and 'message' fields are required"})
		return
	}

	targetJID := parseJID(req.To)
	if err := s.botRouter.SendMessage(targetJID, req.Message); err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Error: "Failed to send message: " + err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: "Message sent successfully"})
}

func (s *Server) handleBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "Method not allowed"})
		return
	}

	var req BroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid JSON body: " + err.Error()})
		return
	}

	if len(req.Recipients) == 0 || req.Message == "" {
		sendJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "'recipients' and 'message' fields are required"})
		return
	}

	failed := []string{}
	for _, recipient := range req.Recipients {
		targetJID := parseJID(recipient)
		if err := s.botRouter.SendMessage(targetJID, req.Message); err != nil {
			failed = append(failed, recipient)
		}
		// Slight sleep to avoid WhatsApp spam throttle
		time.Sleep(300 * time.Millisecond)
	}

	if len(failed) > 0 {
		sendJSON(w, http.StatusMultiStatus, APIResponse{
			Success: false,
			Message: fmt.Sprintf("Broadcast completed with %d failures", len(failed)),
			Error:   fmt.Sprintf("Failed for: %s", strings.Join(failed, ", ")),
		})
		return
	}

	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: fmt.Sprintf("Broadcast sent to all %d recipients", len(req.Recipients)),
	})
}

func parseJID(target string) types.JID {
	target = strings.TrimSpace(target)
	if strings.Contains(target, "@") {
		jid, _ := types.ParseJID(target)
		return jid
	}

	// Clean number (e.g. 0812 -> 62812 or +62 -> 62)
	cleaned := strings.ReplaceAll(strings.ReplaceAll(target, "+", ""), "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	if strings.HasPrefix(cleaned, "0") {
		cleaned = "62" + cleaned[1:]
	}

	return types.NewJID(cleaned, types.DefaultUserServer)
}

func sendJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
