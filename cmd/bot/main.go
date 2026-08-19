package main

import (
	"context"
	"fmt"
	"log"

	"botwa_go_ascii/config"
	"botwa_go_ascii/pkg/asciiapi"
	"botwa_go_ascii/pkg/bot"
	"botwa_go_ascii/pkg/server"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 MEMULAI BOT WHATSAPP ASCII INFORMATIKA")
	fmt.Println("==================================================")

	// 1. Load configuration
	cfg := config.LoadConfig()
	log.Printf("[Main] Config loaded (API: %s, Web: %s, Prefix: '%s')", cfg.AsciiAPIURL, cfg.AsciiWebURL, cfg.BotPrefix)

	// 2. Initialize API Client
	apiClient := asciiapi.NewClient(cfg)

	// Test backend reachability (non-blocking)
	go func() {
		latency, err := apiClient.Ping()
		if err != nil {
			log.Printf("[Main] ⚠️ Peringatan: Tidak dapat menjangkau Web API ASCII (%s): %v", cfg.AsciiAPIURL, err)
			log.Println("[Main] Pastikan server web ascii-if sudah running jika ingin mengetes integrasi API.")
		} else {
			log.Printf("[Main] ✅ Berhasil terhubung ke Web API ASCII (Latency: %v)", latency)
		}
	}()

	// 3. Initialize WhatsApp Bot
	waBot, err := bot.NewBot(cfg, apiClient)
	if err != nil {
		log.Fatalf("[Main] Gagal menginisialisasi WhatsApp Bot: %v", err)
	}

	// 4. Start WhatsApp Bot
	if err := waBot.Start(context.Background()); err != nil {
		log.Fatalf("[Main] Gagal menjalankan koneksi WhatsApp: %v", err)
	}

	// 5. Initialize & Start Webhook Server
	var srv *server.Server
	if cfg.EnableServer {
		srv = server.NewServer(cfg, apiClient, waBot.Router())
		if err := srv.Start(); err != nil {
			log.Printf("[Main] Gagal menjalankan REST server: %v", err)
		}
	}

	fmt.Println("==================================================")
	fmt.Println("🟢 BOT SIAP! Tekan Ctrl+C untuk mematikan bot.")
	fmt.Println("==================================================")

	// 6. Wait for interrupt signal for graceful shutdown
	bot.WaitForInterrupt()

	log.Println("[Main] Mematikan bot dan server...")
	if srv != nil {
		srv.Stop()
	}
	waBot.Stop()
	log.Println("[Main] Selesai. Sampai jumpa!")
}
