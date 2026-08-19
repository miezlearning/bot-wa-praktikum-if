package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"botwa_go_ascii/config"
	"botwa_go_ascii/pkg/asciiapi"

	_ "modernc.org/sqlite" // Pure Go SQLite driver for whatsmeow sqlstore

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type Bot struct {
	cfg       *config.Config
	apiClient *asciiapi.Client
	waClient  *whatsmeow.Client
	router    *Router
	container *sqlstore.Container
}

func NewBot(cfg *config.Config, apiClient *asciiapi.Client) (*Bot, error) {
	// Custom logger for whatsmeow
	dbLog := waLog.Stdout("Database", "WARN", true)
	waLogger := waLog.Stdout("WhatsApp", "INFO", true)

	// Inisialisasi SQLite database container dengan pure Go driver
	container, err := sqlstore.New(
		context.Background(),
		"sqlite",
		fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)", cfg.DBPath),
		dbLog,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize sqlite store: %w", err)
	}

	// Get or create device store
	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get device store: %w", err)
	}

	if deviceStore == nil {
		deviceStore = container.NewDevice()
		log.Println("[Bot] Created new WhatsApp session device store")
	}

	client := whatsmeow.NewClient(deviceStore, waLogger)

	b := &Bot{
		cfg:       cfg,
		apiClient: apiClient,
		waClient:  client,
		container: container,
	}

	b.router = NewRouter(cfg, apiClient, client)
	client.AddEventHandler(b.eventHandler)

	return b, nil
}

func (b *Bot) Router() *Router {
	return b.router
}

func (b *Bot) Client() *whatsmeow.Client {
	return b.waClient
}

func (b *Bot) Start(ctx context.Context) error {
	// If already logged in, connect directly
	if b.waClient.Store.ID != nil {
		log.Println("[Bot] Existing session found, connecting to WhatsApp...")
		if err := b.waClient.Connect(); err != nil {
			return fmt.Errorf("failed to connect existing session: %w", err)
		}
		log.Println("[Bot] Connected successfully to WhatsApp!")
	} else {
		// New login needed: QR Code or Pairing Code
		log.Println("[Bot] No active session found. Preparing authentication...")
		qrChan, err := b.waClient.GetQRChannel(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get qr channel: %w", err)
		}

		if err := b.waClient.Connect(); err != nil {
			return fmt.Errorf("failed to connect for QR: %w", err)
		}

		// Handle pairing code if configured
		if b.cfg.PairingMode && b.cfg.PhoneNumber != "" {
			code, err := b.waClient.PairPhone(context.Background(), b.cfg.PhoneNumber, true, whatsmeow.PairClientChrome, "Chrome (Windows)")
			if err != nil {
				log.Printf("[Bot] Failed to get pairing code: %v", err)
			} else {
				log.Printf("\n==================================================")
				log.Printf("🔑 KODE PAIRING WHATSAPP: %s", code)
				log.Printf("Masukkan kode ini pada HP Anda di WhatsApp > Perangkat Tertaut")
				log.Printf("==================================================\n")
			}
		}

		// QR code loop in background goroutine
		go func() {
			for evt := range qrChan {
				if evt.Event == "code" {
					DisplayQRCode(evt.Code)
				} else {
					log.Printf("[Bot] QR Event: %s", evt.Event)
				}
			}
		}()
	}

	return nil
}

func (b *Bot) eventHandler(rawEvt interface{}) {
	switch evt := rawEvt.(type) {
	case *events.LoggedOut:
		log.Println("[Bot] ⚠️ Session logged out from phone. Please delete whatsapp.db to rescan.")
	case *events.Connected:
		log.Println("[Bot] ✅ WhatsApp Web socket connected!")
	case *events.Message:
		b.router.HandleMessage(evt)
	}
}

func (b *Bot) Stop() {
	if b.waClient != nil {
		b.waClient.Disconnect()
	}
	if b.container != nil {
		_ = b.container.Close()
	}
	log.Println("[Bot] WhatsApp client disconnected and database closed gracefully.")
}

// WaitForInterrupt blocks until SIGINT/SIGTERM is received
func WaitForInterrupt() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
}
