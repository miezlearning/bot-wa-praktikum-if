package bot

import (
	"fmt"
	"strings"
	"time"

	"botwa_go_ascii/pkg/asciiapi"
)

// FormatHelp builds the help/menu message
func FormatHelp(prefix, webURL string) string {
	var sb strings.Builder
	sb.WriteString("🤖 *ASCII INFORMATIKA BOT (WHATSAPP)*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("Halo! Saya adalah bot asisten resmi Laboratorium Informatika ASCII.\n\n")
	sb.WriteString("📋 *Daftar Perintah:*\n")
	sb.WriteString(fmt.Sprintf("• `%sjadwal` - Cek jadwal praktikum aktif\n", prefix))
	sb.WriteString(fmt.Sprintf("• `%skelas` [kode/matkul] - Cek daftar kelas\n", prefix))
	sb.WriteString(fmt.Sprintf("• `%smodul` [kata kunci] - Cari & unduh modul praktikum\n", prefix))
	sb.WriteString(fmt.Sprintf("• `%spengumuman` - Pengumuman lab terkini\n", prefix))
	sb.WriteString(fmt.Sprintf("• `%sberita` - Berita & update praktikum\n", prefix))
	sb.WriteString(fmt.Sprintf("• `%skontak` - Kontak & tautan resmi lab\n", prefix))
	sb.WriteString(fmt.Sprintf("• `%saturan` - Tata tertib & aturan lab\n", prefix))
	sb.WriteString(fmt.Sprintf("• `%sfaq` - Pertanyaan yang sering diajukan\n", prefix))
	sb.WriteString(fmt.Sprintf("• `%sping` - Cek konektivitas bot & backend\n\n", prefix))
	sb.WriteString(fmt.Sprintf("🌐 *Portal Web:* %s\n", webURL))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("_Laboratorium Informatika Universitas Mulawarman_")
	return sb.String()
}

// FormatSchedules formats schedule items
func FormatSchedules(items []asciiapi.ScheduleItem, webURL string) string {
	if len(items) == 0 {
		return "ℹ️ Belum ada jadwal praktikum yang aktif saat ini.\nKunjungi web: " + webURL + "/jadwal-praktikum"
	}

	var sb strings.Builder
	sb.WriteString("📅 *JADWAL PRAKTIKUM LABORATORIUM ASCII*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	timeSlotNames := map[int]string{
		1: "07:30 - 09:00",
		2: "09:10 - 10:40",
		3: "10:50 - 12:20",
		4: "13:00 - 14:30",
		5: "14:40 - 16:10",
		6: "16:20 - 17:50",
	}

	for i, item := range items {
		course := item.Course
		if course == "" {
			course = item.MataKuliah
		}
		classCode := item.Class
		if classCode == "" {
			classCode = item.KodeKelas
		}
		day := item.Day
		if day == "" {
			day = item.Hari
		}
		loc := item.Location
		if loc == "" {
			loc = item.Tempat
		}

		timeStr := item.Jam
		if timeStr == "" && item.TimeSlot > 0 {
			if slotStr, ok := timeSlotNames[item.TimeSlot]; ok {
				timeStr = slotStr
			} else {
				timeStr = fmt.Sprintf("Sesi %d", item.TimeSlot)
			}
		}

		sb.WriteString(fmt.Sprintf("*%d. %s* (%s)\n", i+1, course, classCode))
		if day != "" {
			sb.WriteString(fmt.Sprintf("   📆 Hari: %s\n", day))
		}
		if timeStr != "" {
			sb.WriteString(fmt.Sprintf("   ⏰ Waktu: %s\n", timeStr))
		}
		if loc != "" {
			sb.WriteString(fmt.Sprintf("   📍 Lokasi: %s\n", loc))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("🔗 Selengkapnya: %s/jadwal-praktikum", webURL))
	return sb.String()
}

// FormatClasses formats class list or filtered classes
func FormatClasses(items []asciiapi.ClassItem, filter string, webURL string) string {
	if len(items) == 0 {
		return "ℹ️ Tidak ditemukan data kelas praktikum."
	}

	var sb strings.Builder
	sb.WriteString("🏫 *PEMBAGIAN KELAS PRAKTIKUM*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	count := 0
	for _, item := range items {
		if filter != "" {
			f := strings.ToLower(filter)
			matchKode := strings.Contains(strings.ToLower(item.KodeKelas), f)
			matchMatkul := strings.Contains(strings.ToLower(item.MataKuliah), f)
			matchNama := strings.Contains(strings.ToLower(item.NamaKelas), f)
			if !matchKode && !matchMatkul && !matchNama {
				continue
			}
		}

		count++
		title := item.MataKuliah
		if item.NamaKelas != "" && item.NamaKelas != item.KodeKelas {
			title = fmt.Sprintf("%s (%s)", item.MataKuliah, item.NamaKelas)
		}
		sb.WriteString(fmt.Sprintf("*[%s] %s*\n", item.KodeKelas, title))
		if item.IsPilihan {
			sb.WriteString("   🏷️ Mata Kuliah Pilihan\n")
		}
		sb.WriteString("\n")

		if count >= 20 {
			sb.WriteString(fmt.Sprintf("_...dan %d kelas lainnya._\n\n", len(items)-count))
			break
		}
	}

	if count == 0 {
		return fmt.Sprintf("ℹ️ Tidak ditemukan kelas dengan kata kunci *\"%s\"*.\nKetik `!kelas` untuk melihat semua.", filter)
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("🔗 Detail pembagian kelompok: %s/pembagian-kelas", webURL))
	return sb.String()
}

// FormatModul formats practical modules
func FormatModul(items []asciiapi.ModulItem, filter string, webURL string) string {
	if len(items) == 0 {
		return "ℹ️ Belum ada modul praktikum yang diunggah."
	}

	var sb strings.Builder
	sb.WriteString("📚 *MODUL PRAKTIKUM LABORATORIUM ASCII*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	count := 0
	for _, m := range items {
		title := m.Title
		if title == "" {
			title = "Modul Praktikum"
		}

		if filter != "" {
			f := strings.ToLower(filter)
			matchTitle := strings.Contains(strings.ToLower(title), f)
			matchMatkul := strings.Contains(strings.ToLower(m.MataKuliah), f)
			matchDesc := strings.Contains(strings.ToLower(m.Description), f)
			if !matchTitle && !matchMatkul && !matchDesc {
				continue
			}
		}

		count++
		sb.WriteString(fmt.Sprintf("📌 *%s*\n", title))
		if m.MataKuliah != "" {
			sb.WriteString(fmt.Sprintf("   📖 Mata Kuliah: %s\n", m.MataKuliah))
		}
		if m.Tahun > 0 {
			sb.WriteString(fmt.Sprintf("   📅 Tahun: %d\n", m.Tahun))
		}
		if m.Description != "" {
			sb.WriteString(fmt.Sprintf("   📝 %s\n", m.Description))
		}
		if m.DriveFileURL != "" {
			sb.WriteString(fmt.Sprintf("   📥 Unduh: %s\n", m.DriveFileURL))
		}
		sb.WriteString("\n")

		if count >= 10 {
			sb.WriteString(fmt.Sprintf("_...dan %d modul lainnya._\n\n", len(items)-count))
			break
		}
	}

	if count == 0 {
		return fmt.Sprintf("ℹ️ Tidak ditemukan modul dengan kata kunci *\"%s\"*.\nKetik `!modul` untuk melihat daftar terbaru.", filter)
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("🔗 Akses modul lengkap: %s/modul", webURL))
	return sb.String()
}

// FormatAnnouncements formats announcements
func FormatAnnouncements(items []asciiapi.AnnouncementItem, webURL string) string {
	if len(items) == 0 {
		return "ℹ️ Belum ada pengumuman lab saat ini."
	}

	var sb strings.Builder
	sb.WriteString("📢 *PENGUMUMAN LABORATORIUM ASCII*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for i, item := range items {
		if i >= 5 {
			break
		}
		title := item.Title
		if title == "" {
			title = "Pengumuman"
		}
		sb.WriteString(fmt.Sprintf("*%d. %s*\n", i+1, title))
		if item.Description != "" {
			sb.WriteString(fmt.Sprintf("%s\n", item.Description))
		}
		if item.Link != "" {
			sb.WriteString(fmt.Sprintf("🔗 %s\n", item.Link))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("🔗 Kunjungi: %s", webURL))
	return sb.String()
}

// FormatBerita formats news articles
func FormatBerita(items []asciiapi.BeritaItem, webURL string) string {
	if len(items) == 0 {
		return "ℹ️ Belum ada berita praktikum yang dipublikasikan."
	}

	var sb strings.Builder
	sb.WriteString("📰 *BERITA & ARTIKEL PRAKTIKUM ASCII*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for i, item := range items {
		if i >= 4 {
			break
		}
		title := item.Title
		if title == "" {
			title = "Berita Lab"
		}
		sb.WriteString(fmt.Sprintf("*%d. %s*\n", i+1, title))
		if item.Description != "" {
			sb.WriteString(fmt.Sprintf("   _%s_\n", item.Description))
		}
		if item.Slug != "" {
			sb.WriteString(fmt.Sprintf("   🔗 %s/berita-praktikum/%s\n\n", webURL, item.Slug))
		} else {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("🔗 Baca selengkapnya: %s/berita-praktikum", webURL))
	return sb.String()
}

// FormatContacts formats lab contacts
func FormatContacts(items []asciiapi.ContactItem, webURL string) string {
	if len(items) == 0 {
		return "ℹ️ Belum ada daftar kontak praktikum."
	}

	var sb strings.Builder
	sb.WriteString("📞 *KONTAK & TAUTAN RESMI LAB ASCII*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for i, item := range items {
		title := item.Title
		if title == "" {
			title = item.Platform
		}
		sb.WriteString(fmt.Sprintf("*%d. %s*", i+1, title))
		if item.Platform != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", item.Platform))
		}
		sb.WriteString("\n")
		if item.Description != "" {
			sb.WriteString(fmt.Sprintf("   📝 %s\n", item.Description))
		}
		if item.URL != "" {
			sb.WriteString(fmt.Sprintf("   🔗 %s\n", item.URL))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("🔗 Kunjungi: %s/kontak", webURL))
	return sb.String()
}

// FormatAturan formats lab rules
func FormatAturan(items []asciiapi.AturanItem, webURL string) string {
	if len(items) == 0 {
		return "ℹ️ Belum ada data tata tertib laboratorium."
	}

	var sb strings.Builder
	sb.WriteString("📜 *TATA TERTIB & ATURAN PRAKTIKUM*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for i, item := range items {
		title := item.Title
		if title == "" {
			title = "Tata Tertib"
		}
		sb.WriteString(fmt.Sprintf("*%d. %s*\n", i+1, title))
		if item.Description != "" {
			sb.WriteString(fmt.Sprintf("   _%s_\n", item.Description))
		}
		if item.Content != "" {
			sb.WriteString(fmt.Sprintf("%s\n", item.Content))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("🔗 Selengkapnya: %s/aturan-praktikum", webURL))
	return sb.String()
}

// FormatFAQ formats FAQ items
func FormatFAQ(items []asciiapi.FAQItem, webURL string) string {
	if len(items) == 0 {
		return "ℹ️ Belum ada daftar FAQ laboratorium saat ini.\nKunjungi web: " + webURL
	}

	var sb strings.Builder
	sb.WriteString("❓ *FAQ (FREQUENTLY ASKED QUESTIONS)*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for i, item := range items {
		if i >= 8 {
			break
		}
		question := item.Title
		if question == "" {
			question = "Pertanyaan"
		}
		sb.WriteString(fmt.Sprintf("*Q: %s*\n", question))
		if item.Description != "" {
			sb.WriteString(fmt.Sprintf("A: %s\n\n", item.Description))
		}
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("🔗 Kunjungi portal: %s", webURL))
	return sb.String()
}

// FormatPing formats ping response
func FormatPing(botLatency time.Duration, apiLatency time.Duration, apiErr error) string {
	var sb strings.Builder
	sb.WriteString("🏓 *PONG!*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("⚡ *Bot Engine:* Aktif (Process: %v)\n", botLatency.Round(time.Millisecond)))
	if apiErr != nil {
		sb.WriteString(fmt.Sprintf("🌐 *Web API ASCII:* ❌ Error (%v)\n", apiErr))
	} else {
		sb.WriteString(fmt.Sprintf("🌐 *Web API ASCII:* ✅ OK (Latency: %v)\n", apiLatency.Round(time.Millisecond)))
	}
	sb.WriteString(fmt.Sprintf("🕒 *Waktu Server:* %s\n", time.Now().Format("02 Jan 2006 15:04:05 MST")))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return sb.String()
}
