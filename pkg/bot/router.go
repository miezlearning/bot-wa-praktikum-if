package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
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

	var cmd string
	var args []string

	matchedPrefix := ""
	if strings.HasPrefix(text, r.cfg.BotPrefix) {
		matchedPrefix = r.cfg.BotPrefix
	} else if strings.HasPrefix(text, "!") {
		matchedPrefix = "!"
	} else if strings.HasPrefix(text, ".") {
		matchedPrefix = "."
	} else if strings.HasPrefix(text, "/") {
		matchedPrefix = "/"
	} else if strings.HasPrefix(text, "#") {
		matchedPrefix = "#"
	}

	if matchedPrefix != "" {
		rawCommand := strings.TrimPrefix(text, matchedPrefix)
		parts := strings.Fields(rawCommand)
		if len(parts) == 0 {
			return
		}
		cmd = strings.ToLower(parts[0])
		args = parts[1:]
	} else {
		// Support direct numeric shortcuts and plain text keywords without prefix
		parts := strings.Fields(text)
		if len(parts) == 0 {
			return
		}
		first := strings.ToLower(parts[0])

		switch first {
		case "1", "bimbingan", "mentee", "progres":
			cmd = "bimbingan"
			args = parts[1:]
		case "2", "jadwal":
			cmd = "jadwal"
			args = parts[1:]
		case "3", "kelas":
			cmd = "kelas"
			args = parts[1:]
		case "4", "modul":
			cmd = "modul"
			args = parts[1:]
		case "5", "pengumuman":
			cmd = "pengumuman"
			args = parts[1:]
		case "6", "berita":
			cmd = "berita"
			args = parts[1:]
		case "7", "kontak", "aslab":
			cmd = "kontak"
			args = parts[1:]
		case "8", "profil", "whoami", "akun":
			cmd = "profil"
			args = parts[1:]
		case "foto", "photo":
			cmd = "foto"
			args = parts[1:]
		case "9", "aturan", "rules":
			cmd = "aturan"
			args = parts[1:]
		case "0", "menu", "help", "halo", "hi", "p", "start":
			cmd = "help"
			args = parts[1:]
		case "login", "masuk":
			cmd = "login"
			args = parts[1:]
		case "logout", "keluar":
			cmd = "logout"
			args = parts[1:]
		case "revisi", "rev":
			cmd = "revisi"
			args = parts[1:]
		case "acc":
			cmd = "acc"
			args = parts[1:]
		case "accfinal":
			cmd = "accfinal"
			args = parts[1:]
		default:
			return
		}
	}

	// Map numeric command with prefix e.g. "!1" -> "bimbingan", "!2" -> "jadwal", etc.
	switch cmd {
	case "1":
		cmd = "bimbingan"
	case "2":
		cmd = "jadwal"
	case "3":
		cmd = "kelas"
	case "4":
		cmd = "modul"
	case "5":
		cmd = "pengumuman"
	case "6":
		cmd = "berita"
	case "7":
		cmd = "kontak"
	case "8":
		cmd = "profil"
	case "9":
		cmd = "aturan"
	case "0":
		cmd = "help"
	}

	argStr := strings.Join(args, " ")
	senderPhone := evt.Info.Sender.User
	if evt.Info.Sender.Server == "lid" && !evt.Info.IsGroup && evt.Info.Chat.Server != "lid" && evt.Info.Chat.User != "" {
		senderPhone = evt.Info.Chat.User
	}
	sender := senderPhone
	if evt.Info.PushName != "" {
		sender = fmt.Sprintf("%s (%s)", evt.Info.PushName, sender)
	}
	log.Printf("[Bot] Received command '%s' with args '%s' from %s", cmd, argStr, sender)

	r.dispatch(evt.Info.Chat, senderPhone, cmd, argStr)
}

func (r *Router) dispatch(target types.JID, senderPhone, cmd, args string) {
	start := time.Now()

	switch cmd {
	case "help", "menu", "start":
		currentUser, _ := r.apiClient.FindUserByPhone(senderPhone)
		reply := FormatHelp(r.cfg.BotPrefix, r.apiClient.WebURL(), currentUser)
		r.SendMessage(target, reply)

	case "login", "masuk", "auth":
		existing, _ := r.apiClient.FindUserByPhone(senderPhone)
		parts := strings.Fields(args)
		if len(parts) < 2 {
			if existing != nil {
				r.SendMessage(target, fmt.Sprintf("ℹ️ *Anda Sudah Login!*\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n• *Nama:* %s\n• *NIM / Username:* %s\n• *Peran (Role):* %s\n• *Status:* 🟢 Terhubung & Aktif\n\n_Ketik `%slogout` terlebih dahulu jika ingin mengganti akun lain._", existing.Name, existing.Username, roleLabel(existing.Role), r.cfg.BotPrefix))
				return
			}
			guide := fmt.Sprintf("⚠️ *Format Perintah Login:*\n`%slogin <NIM> <Password>`\n\n*Contoh:*\n`%slogin 2101552001 Password123`\n\n_Akun portal praktikum ASCII Anda akan otomatis ditautkan ke nomor WhatsApp ini secara permanen._", r.cfg.BotPrefix, r.cfg.BotPrefix)
			r.SendMessage(target, guide)
			return
		}

		nim := parts[0]
		password := parts[1]
		u, err := r.apiClient.LoginUser(senderPhone, nim, password)
		if err != nil {
			log.Printf("[Bot] Login error for %s: %v", nim, err)
			r.SendMessage(target, "❌ Gagal login: "+err.Error())
			return
		}

		reply := FormatLoginSuccess(u, r.cfg.BotPrefix, r.apiClient.WebURL())
		r.SendMessage(target, reply)

	case "logout", "keluar":
		existing, _ := r.apiClient.FindUserByPhone(senderPhone)
		if existing == nil {
			r.SendMessage(target, fmt.Sprintf("ℹ️ *Anda Belum Login*\nNomor WhatsApp ini belum tertaut dengan akun portal manapun.\n\nKetik `%slogin <NIM> <password>` untuk masuk.", r.cfg.BotPrefix))
			return
		}
		u, err := r.apiClient.LogoutUser(senderPhone)
		if err != nil {
			r.SendMessage(target, "❌ "+err.Error())
			return
		}
		reply := FormatLogoutSuccess(u)
		r.SendMessage(target, reply)

	case "profil", "profile", "whoami", "akun", "me", "status", "foto", "photo":
		if strings.TrimSpace(args) != "" {
			targetNIM := strings.TrimSpace(args)
			photo, err := FetchStudentPhoto(targetNIM)
			if err != nil || len(photo) == 0 {
				r.SendMessage(target, fmt.Sprintf("ℹ️ Foto untuk NIM `%s` tidak ditemukan di AIS Unmul.", targetNIM))
				return
			}
			caption := fmt.Sprintf("👤 *Foto Mahasiswa Unmul*\n🆔 *NIM:* %s\n🌐 *Sumber:* AIS Universitas Mulawarman", targetNIM)
			_ = r.SendImageMessage(target, photo, caption)
			return
		}

		u, err := r.apiClient.GetProfile(senderPhone)
		if err != nil {
			log.Printf("[Bot] Profile fetch error: %v", err)
			r.SendMessage(target, "❌ Gagal memuat profil: "+err.Error())
			return
		}
		reply := FormatProfile(u, r.cfg.BotPrefix, r.apiClient.WebURL())
		if u != nil && u.Username != "" {
			photo, photoErr := FetchStudentPhoto(u.Username)
			if photoErr == nil && len(photo) > 0 {
				_ = r.SendImageMessage(target, photo, reply)
				return
			}
		}
		r.SendMessage(target, reply)

	case "bimbingan", "bimbing", "mentee", "mentees", "progres", "tubes", "tugasbesar":
		summary, err := r.apiClient.GetBimbinganSummary(senderPhone, args)
		if err != nil {
			log.Printf("[Bot] Error fetching bimbingan summary: %v", err)
			r.SendMessage(target, "❌ Gagal memuat data bimbingan: "+err.Error())
			return
		}
		reply := FormatBimbinganSummary(summary, r.cfg.BotPrefix, r.apiClient.WebURL())
		r.SendMessage(target, reply)

	case "revisi", "rev":
		caller, _ := r.apiClient.FindUserByPhone(senderPhone)
		if caller == nil {
			r.SendMessage(target, fmt.Sprintf("🔒 *Akses Ditolak (Belum Login)*\nAnda harus login sebagai Asisten Laboratorium terlebih dahulu untuk dapat memberikan revisi.\n\n👉 Ketik: `%slogin <NIM> <password>`", r.cfg.BotPrefix))
			return
		}
		if caller.Role != "aslab" && caller.Role != "pengurus" && caller.Role != "koordinator" && caller.Role != "dosen" && caller.Role != "admin" {
			r.SendMessage(target, fmt.Sprintf("🔒 *Akses Ditolak*\nAkun Anda terdaftar sebagai *%s*. Perintah ini khusus untuk Asisten Laboratorium.", roleLabel(caller.Role)))
			return
		}

		parts := strings.Fields(args)
		if len(parts) < 3 {
			guide := fmt.Sprintf("⚠️ *Format Perintah Revisi:*\n`%srevisi <No/ID_Kelompok> <Tahap: 0/1/2> <Catatan revisi>`\n\n*Contoh:*\n• `%srevisi 1 0 Judul terlalu luas, tolong batasi ruang lingkupnya.`\n• `%srevisi 1 1 Flowchart perbaiki modul auth.`\n• `%srevisi 1 2 Fitur upload berkas belum selesai.`\n\n*Keterangan Tahap:*\n• `0` = Konsul 0 (Konsep & Judul)\n• `1` = Konsul 1 (Flow Program & Desain)\n• `2` = Konsul 2 (70%% Koding & Implementasi)\n\n_Ketik `%sbimbingan` untuk melihat nomor urut & ID kelompok bimbingan Anda._", r.cfg.BotPrefix, r.cfg.BotPrefix, r.cfg.BotPrefix, r.cfg.BotPrefix, r.cfg.BotPrefix)
			r.SendMessage(target, guide)
			return
		}

		groupID := parts[0]
		stage, err := strconv.Atoi(parts[1])
		if err != nil || stage < 0 || stage > 2 {
			r.SendMessage(target, "⚠️ Tahap konsul tidak valid. Gunakan `0` (Konsep), `1` (Flow), atau `2` (70% Koding).")
			return
		}
		catatan := strings.TrimSpace(strings.Join(parts[2:], " "))
		if catatan == "" {
			r.SendMessage(target, "⚠️ Harap cantumkan catatan revisi untuk mahasiswa.")
			return
		}

		res, err := r.apiClient.ReviewKonsul(asciiapi.ReviewKonsulParams{
			SenderPhone: senderPhone,
			GroupID:     groupID,
			Stage:       stage,
			Status:      "revisi",
			Catatan:     catatan,
		})
		if err != nil {
			log.Printf("[Bot] Error giving revisi: %v", err)
			r.SendMessage(target, "❌ Gagal mengirim catatan revisi: "+err.Error())
			return
		}

		// Send success confirmation to Aslab
		reply := FormatReviewSuccess(res, r.apiClient.WebURL())
		r.SendMessage(target, reply)

		// Dispatch notifications to students asynchronously
		if len(res.Recipients) > 0 {
			go func(result *asciiapi.ReviewKonsulResult, note string) {
				for _, m := range result.Recipients {
					if m.PhoneNumber == "" {
						continue
					}
					studentJID := types.NewJID(m.PhoneNumber, types.DefaultUserServer)
					msg := FormatStudentReviewNotification(result.Group, result.AslabName, result.StageName, "revisi", note, r.apiClient.WebURL())
					_ = r.SendMessage(studentJID, msg)
					time.Sleep(300 * time.Millisecond)
				}
			}(res, catatan)
		}

	case "acc":
		caller, _ := r.apiClient.FindUserByPhone(senderPhone)
		if caller == nil {
			r.SendMessage(target, fmt.Sprintf("🔒 *Akses Ditolak (Belum Login)*\nAnda harus login sebagai Asisten Laboratorium terlebih dahulu untuk dapat menyetujui (ACC) asistensi.\n\n👉 Ketik: `%slogin <NIM> <password>`", r.cfg.BotPrefix))
			return
		}
		if caller.Role != "aslab" && caller.Role != "pengurus" && caller.Role != "koordinator" && caller.Role != "dosen" && caller.Role != "admin" {
			r.SendMessage(target, fmt.Sprintf("🔒 *Akses Ditolak*\nAkun Anda terdaftar sebagai *%s*. Perintah ini khusus untuk Asisten Laboratorium.", roleLabel(caller.Role)))
			return
		}

		parts := strings.Fields(args)
		if len(parts) < 2 {
			guide := fmt.Sprintf("⚠️ *Format Perintah ACC:*\n`%sacc <No/ID_Kelompok> <Tahap: 0/1/2> [Catatan opsional]`\n\n*Contoh:*\n• `%sacc 1 0 Konsep & judul disetujui, lanjut flowchart.`\n• `%sacc 1 1 Flowchart sudah baik, lanjut koding 70%%.`\n• `%sacc 1 2 Progres koding sesuai, siap demo.`\n\n*Keterangan Tahap:*\n• `0` = Konsul 0 (Konsep & Judul)\n• `1` = Konsul 1 (Flow Program & Desain)\n• `2` = Konsul 2 (70%% Koding & Implementasi)\n\n_Ketik `%sbimbingan` untuk melihat nomor urut & ID kelompok bimbingan Anda._", r.cfg.BotPrefix, r.cfg.BotPrefix, r.cfg.BotPrefix, r.cfg.BotPrefix, r.cfg.BotPrefix)
			r.SendMessage(target, guide)
			return
		}

		groupID := parts[0]
		stage, err := strconv.Atoi(parts[1])
		if err != nil || stage < 0 || stage > 2 {
			r.SendMessage(target, "⚠️ Tahap konsul tidak valid. Gunakan `0` (Konsep), `1` (Flow), atau `2` (70% Koding).")
			return
		}

		catatan := ""
		if len(parts) > 2 {
			catatan = strings.TrimSpace(strings.Join(parts[2:], " "))
		}

		res, err := r.apiClient.ReviewKonsul(asciiapi.ReviewKonsulParams{
			SenderPhone: senderPhone,
			GroupID:     groupID,
			Stage:       stage,
			Status:      "acc",
			Catatan:     catatan,
		})
		if err != nil {
			log.Printf("[Bot] Error giving ACC: %v", err)
			r.SendMessage(target, "❌ Gagal memproses ACC: "+err.Error())
			return
		}

		reply := FormatReviewSuccess(res, r.apiClient.WebURL())
		r.SendMessage(target, reply)

		// Dispatch notifications to students asynchronously
		if len(res.Recipients) > 0 {
			go func(result *asciiapi.ReviewKonsulResult, note string) {
				for _, m := range result.Recipients {
					if m.PhoneNumber == "" {
						continue
					}
					studentJID := types.NewJID(m.PhoneNumber, types.DefaultUserServer)
					msg := FormatStudentReviewNotification(result.Group, result.AslabName, result.StageName, "acc", note, r.apiClient.WebURL())
					_ = r.SendMessage(studentJID, msg)
					time.Sleep(300 * time.Millisecond)
				}
			}(res, catatan)
		}

	case "accfinal", "acctubes", "accproyek":
		caller, _ := r.apiClient.FindUserByPhone(senderPhone)
		if caller == nil {
			r.SendMessage(target, fmt.Sprintf("🔒 *Akses Ditolak (Belum Login)*\nAnda harus login sebagai Asisten Laboratorium terlebih dahulu untuk dapat memberikan ACC Final.\n\n👉 Ketik: `%slogin <NIM> <password>`", r.cfg.BotPrefix))
			return
		}
		if caller.Role != "aslab" && caller.Role != "pengurus" && caller.Role != "koordinator" && caller.Role != "dosen" && caller.Role != "admin" {
			r.SendMessage(target, fmt.Sprintf("🔒 *Akses Ditolak*\nAkun Anda terdaftar sebagai *%s*. Perintah ini khusus untuk Asisten Laboratorium.", roleLabel(caller.Role)))
			return
		}

		parts := strings.Fields(args)
		if len(parts) < 1 {
			guide := fmt.Sprintf("⚠️ *Format Perintah ACC Final:*\n`%saccfinal <No/ID_Kelompok> [Catatan opsional]`\n\n*Contoh:*\n• `%saccfinal 1 Proyek siap untuk demo praktikum.`\n• `%saccfinal KEL-01`\n\n_Ketik `%sbimbingan` untuk melihat nomor urut & ID kelompok bimbingan Anda._", r.cfg.BotPrefix, r.cfg.BotPrefix, r.cfg.BotPrefix, r.cfg.BotPrefix)
			r.SendMessage(target, guide)
			return
		}

		groupID := parts[0]
		catatan := ""
		if len(parts) > 1 {
			catatan = strings.TrimSpace(strings.Join(parts[1:], " "))
		}

		res, err := r.apiClient.ToggleAccFinal(asciiapi.AccFinalParams{
			SenderPhone: senderPhone,
			GroupID:     groupID,
			IsAccFinal:  true,
			Catatan:     catatan,
		})
		if err != nil {
			log.Printf("[Bot] Error giving ACC Final: %v", err)
			r.SendMessage(target, "❌ Gagal memproses ACC Final: "+err.Error())
			return
		}

		reply := FormatReviewSuccess(res, r.apiClient.WebURL())
		r.SendMessage(target, reply)

		// Dispatch notifications to students asynchronously
		if len(res.Recipients) > 0 {
			go func(result *asciiapi.ReviewKonsulResult, note string) {
				for _, m := range result.Recipients {
					if m.PhoneNumber == "" {
						continue
					}
					studentJID := types.NewJID(m.PhoneNumber, types.DefaultUserServer)
					msg := FormatStudentReviewNotification(result.Group, result.AslabName, result.StageName, "accfinal", note, r.apiClient.WebURL())
					_ = r.SendMessage(studentJID, msg)
					time.Sleep(300 * time.Millisecond)
				}
			}(res, catatan)
		}

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

// SendImageMessage sends an image with caption to WhatsApp recipient
func (r *Router) SendImageMessage(target types.JID, imageBytes []byte, caption string) error {
	if r.waClient == nil || !r.waClient.IsConnected() {
		return fmt.Errorf("whatsapp client is not connected")
	}

	uploaded, err := r.waClient.Upload(context.Background(), imageBytes, whatsmeow.MediaImage)
	if err != nil {
		log.Printf("[Bot] Failed to upload image media (%v), falling back to text", err)
		return r.SendMessage(target, caption)
	}

	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			Caption:       proto.String(caption),
			Mimetype:      proto.String("image/jpeg"),
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
		},
	}

	_, err = r.waClient.SendMessage(context.Background(), target, msg)
	if err != nil {
		log.Printf("[Bot] Failed to send image message to %s (%v), falling back to text", target.String(), err)
		return r.SendMessage(target, caption)
	}
	return nil
}

// FetchStudentPhoto fetches student photo from AIS Unmul
func FetchStudentPhoto(nim string) ([]byte, error) {
	nim = strings.TrimSpace(nim)
	if nim == "" {
		return nil, fmt.Errorf("nim is empty")
	}

	url := fmt.Sprintf("https://ais.unmul.ac.id/file/foto/%s", nim)
	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil || len(data) < 200 {
		return nil, fmt.Errorf("invalid image data")
	}

	// Reject HTML error pages
	if bytes.HasPrefix(data, []byte("<")) {
		return nil, fmt.Errorf("received HTML instead of image")
	}

	return data, nil
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
	if msg.ButtonsResponseMessage != nil {
		if msg.ButtonsResponseMessage.SelectedButtonID != nil && *msg.ButtonsResponseMessage.SelectedButtonID != "" {
			return *msg.ButtonsResponseMessage.SelectedButtonID
		}
	}
	if msg.ListResponseMessage != nil {
		if msg.ListResponseMessage.SingleSelectReply != nil && msg.ListResponseMessage.SingleSelectReply.SelectedRowID != nil {
			return *msg.ListResponseMessage.SingleSelectReply.SelectedRowID
		}
		if msg.ListResponseMessage.Title != nil && *msg.ListResponseMessage.Title != "" {
			return *msg.ListResponseMessage.Title
		}
	}
	if msg.TemplateButtonReplyMessage != nil {
		if msg.TemplateButtonReplyMessage.SelectedID != nil && *msg.TemplateButtonReplyMessage.SelectedID != "" {
			return *msg.TemplateButtonReplyMessage.SelectedID
		}
		if msg.TemplateButtonReplyMessage.SelectedDisplayText != nil && *msg.TemplateButtonReplyMessage.SelectedDisplayText != "" {
			return *msg.TemplateButtonReplyMessage.SelectedDisplayText
		}
	}
	if msg.InteractiveResponseMessage != nil {
		if nfm := msg.InteractiveResponseMessage.GetNativeFlowResponseMessage(); nfm != nil {
			if nfm.ParamsJSON != nil && *nfm.ParamsJSON != "" {
				var p struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal([]byte(*nfm.ParamsJSON), &p); err == nil && p.ID != "" {
					return p.ID
				}
				return *nfm.ParamsJSON
			}
		}
		if msg.InteractiveResponseMessage.Body != nil && msg.InteractiveResponseMessage.Body.Text != nil {
			return *msg.InteractiveResponseMessage.Body.Text
		}
	}
	if msg.ViewOnceMessage != nil && msg.ViewOnceMessage.Message != nil {
		return extractMessageText(msg.ViewOnceMessage.Message)
	}
	if msg.EphemeralMessage != nil && msg.EphemeralMessage.Message != nil {
		return extractMessageText(msg.EphemeralMessage.Message)
	}
	if msg.DocumentWithCaptionMessage != nil && msg.DocumentWithCaptionMessage.Message != nil {
		return extractMessageText(msg.DocumentWithCaptionMessage.Message)
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
