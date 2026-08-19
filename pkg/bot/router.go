package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"botwa_go_ascii/config"
	"botwa_go_ascii/pkg/asciiapi"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

type Router struct {
	cfg       *config.Config
	apiClient *asciiapi.Client
	waClient  *whatsmeow.Client
}

func NewRouter(cfg *config.Config, apiClient *asciiapi.Client, waClient *whatsmeow.Client) *Router {
	return &Router{
		cfg:       cfg,
		apiClient: apiClient,
		waClient:  waClient,
	}
}

// HandleMessage parses and routes incoming WhatsApp messages
func (r *Router) HandleMessage(evt *events.Message) {
	// Ignore messages from self or empty messages
	if evt.Info.IsFromMe {
		return
	}

	text := extractMessageText(evt.Message)
	if text == "" {
		return
	}

	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, r.cfg.BotPrefix) {
		return
	}

	rawCommand := strings.TrimPrefix(text, r.cfg.BotPrefix)
	parts := strings.Fields(rawCommand)
	if len(parts) == 0 {
		return
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]
	argStr := strings.Join(args, " ")

	sender := evt.Info.Sender.User
	if evt.Info.PushName != "" {
		sender = fmt.Sprintf("%s (%s)", evt.Info.PushName, sender)
	}
	log.Printf("[Bot] Received command '%s' with args '%s' from %s", cmd, argStr, sender)

	r.dispatch(evt.Info.Chat, cmd, argStr)
}

func (r *Router) dispatch(target types.JID, cmd, args string) {
	start := time.Now()

	switch cmd {
	case "help", "menu", "start":
		reply := FormatHelp(r.cfg.BotPrefix, r.apiClient.WebURL())
		r.SendMessage(target, reply)

	case "jadwal", "schedule":
		schedules, err := r.apiClient.GetSchedules()
		if err != nil {
			log.Printf("[Bot] Error fetching schedules: %v", err)
			r.SendMessage(target, "❌ Gagal mengambil data jadwal dari web server ASCII: "+err.Error())
			return
		}
		reply := FormatSchedules(schedules, r.apiClient.WebURL())
		r.SendMessage(target, reply)

	case "kelas", "kelompok", "class":
		classes, err := r.apiClient.GetClasses()
		if err != nil {
			log.Printf("[Bot] Error fetching classes: %v", err)
			r.SendMessage(target, "❌ Gagal mengambil data kelas dari web server ASCII: "+err.Error())
			return
		}
		reply := FormatClasses(classes, args, r.apiClient.WebURL())
		r.SendMessage(target, reply)

	case "modul", "module":
		moduls, err := r.apiClient.GetAllModul()
		if err != nil {
			log.Printf("[Bot] Error fetching moduls: %v", err)
			r.SendMessage(target, "❌ Gagal mengambil data modul dari web server ASCII: "+err.Error())
			return
		}
		reply := FormatModul(moduls, args, r.apiClient.WebURL())
		r.SendMessage(target, reply)

	case "pengumuman", "announcement", "info":
		announcements, err := r.apiClient.GetAllAnnouncements()
		if err != nil {
			log.Printf("[Bot] Error fetching announcements: %v", err)
			r.SendMessage(target, "❌ Gagal mengambil pengumuman dari web server ASCII: "+err.Error())
			return
		}
		reply := FormatAnnouncements(announcements, r.apiClient.WebURL())
		r.SendMessage(target, reply)

	case "berita", "news":
		beritaList, err := r.apiClient.GetAllBerita()
		if err != nil {
			log.Printf("[Bot] Error fetching berita: %v", err)
			r.SendMessage(target, "❌ Gagal mengambil berita dari web server ASCII: "+err.Error())
			return
		}
		reply := FormatBerita(beritaList, r.apiClient.WebURL())
		r.SendMessage(target, reply)

	case "kontak", "contact", "aslab":
		contacts, err := r.apiClient.GetAllContacts()
		if err != nil {
			log.Printf("[Bot] Error fetching contacts: %v", err)
			r.SendMessage(target, "❌ Gagal mengambil kontak dari web server ASCII: "+err.Error())
			return
		}
		reply := FormatContacts(contacts, r.apiClient.WebURL())
		r.SendMessage(target, reply)

	case "aturan", "rules", "tatatertib":
		aturan, err := r.apiClient.GetAllAturan()
		if err != nil {
			log.Printf("[Bot] Error fetching aturan: %v", err)
			r.SendMessage(target, "❌ Gagal mengambil aturan dari web server ASCII: "+err.Error())
			return
		}
		reply := FormatAturan(aturan, r.apiClient.WebURL())
		r.SendMessage(target, reply)

	case "faq", "tanya":
		faqs, err := r.apiClient.GetAllFAQ()
		if err != nil {
			log.Printf("[Bot] Error fetching faqs: %v", err)
			r.SendMessage(target, "❌ Gagal mengambil FAQ dari web server ASCII: "+err.Error())
			return
		}
		reply := FormatFAQ(faqs, r.apiClient.WebURL())
		r.SendMessage(target, reply)

	case "ping":
		botLatency := time.Since(start)
		apiLatency, apiErr := r.apiClient.Ping()
		reply := FormatPing(botLatency, apiLatency, apiErr)
		r.SendMessage(target, reply)

	default:
		// Optional: ignore or inform unknown command
		r.SendMessage(target, fmt.Sprintf("❓ Perintah `%s%s` tidak dikenali.\nKetik `%shelp` untuk melihat daftar perintah.", r.cfg.BotPrefix, cmd, r.cfg.BotPrefix))
	}
}

// SendMessage sends a plain text WhatsApp message
func (r *Router) SendMessage(target types.JID, text string) error {
	if r.waClient == nil || !r.waClient.IsConnected() {
		return fmt.Errorf("whatsapp client is not connected")
	}

	msg := &waProto.Message{
		Conversation: proto.String(text),
	}

	_, err := r.waClient.SendMessage(context.Background(), target, msg)
	if err != nil {
		log.Printf("[Bot] Failed to send message to %s: %v", target.String(), err)
		return err
	}
	return nil
}

func extractMessageText(msg *waProto.Message) string {
	if msg == nil {
		return ""
	}
	if msg.Conversation != nil && *msg.Conversation != "" {
		return *msg.Conversation
	}
	if msg.ExtendedTextMessage != nil && msg.ExtendedTextMessage.Text != nil {
		return *msg.ExtendedTextMessage.Text
	}
	if msg.ImageMessage != nil && msg.ImageMessage.Caption != nil {
		return *msg.ImageMessage.Caption
	}
	if msg.VideoMessage != nil && msg.VideoMessage.Caption != nil {
		return *msg.VideoMessage.Caption
	}
	if msg.DocumentMessage != nil && msg.DocumentMessage.Caption != nil {
		return *msg.DocumentMessage.Caption
	}
	return ""
}
