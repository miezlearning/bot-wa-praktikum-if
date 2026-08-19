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
	sb.WriteString(fmt.Sprintf("• `%skelas` [kode] - Cek daftar / detail kelas\n", prefix))
	sb.WriteString(fmt.Sprintf("• `%smodul` [kata kunci] - Cari & download modul\n", prefix))
	sb.WriteString(fmt.Sprintf("• `%spengumuman` - Pengumuman lab terkini\n", prefix))
	sb.WriteString(fmt.Sprintf("• `%sberita` - Berita & update praktikum\n", prefix))
	sb.WriteString(fmt.Sprintf("• `%skontak` - Kontak Aslab & Koordinator\n", prefix))
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

	for i, item := range items {
		sb.WriteString(fmt.Sprintf("*%d. %s* (%s)\n", i+1, item.MataKuliah, item.KodeKelas))
		sb.WriteString(fmt.Sprintf("   📆 Hari: %s\n", item.Hari))
		sb.WriteString(fmt.Sprintf("   ⏰ Jam: %s\n", item.Jam))
		if item.Ruang != "" {
			sb.WriteString(fmt.Sprintf("   📍 Ruang: %s\n", item.Ruang))
		}
		if item.Aslab != "" {
			sb.WriteString(fmt.Sprintf("   👤 Aslab: %s\n", item.Aslab))
		}
		if item.Dosen != "" {
			sb.WriteString(fmt.Sprintf("   🎓 Dosen: %s\n", item.Dosen))
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
			matchAslab := strings.Contains(strings.ToLower(item.Aslab), f)
			if !matchKode && !matchMatkul && !matchAslab {
				continue
			}
		}

		count++
		sb.WriteString(fmt.Sprintf("*[%s] %s*\n", item.KodeKelas, item.MataKuliah))
		if item.Hari != "" && item.Jam != "" {
			sb.WriteString(fmt.Sprintf("   📆 Waktu: %s, %s\n", item.Hari, item.Jam))
		}
		if item.Ruang != "" {
			sb.WriteString(fmt.Sprintf("   📍 Ruang: %s\n", item.Ruang))
		}
		if item.Aslab != "" {
			sb.WriteString(fmt.Sprintf("   👤 Aslab: %s\n", item.Aslab))
		}
		sb.WriteString("\n")

		if count >= 15 {
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
		if filter != "" {
			f := strings.ToLower(filter)
			matchJudul := strings.Contains(strings.ToLower(m.Judul), f)
			matchMatkul := strings.Contains(strings.ToLower(m.MataKuliah), f)
			matchDeskripsi := strings.Contains(strings.ToLower(m.Deskripsi), f)
			if !matchJudul && !matchMatkul && !matchDeskripsi {
				continue
			}
		}

		count++
		sb.WriteString(fmt.Sprintf("📌 *%s*\n", m.Judul))
		if m.MataKuliah != "" {
			sb.WriteString(fmt.Sprintf("   📖 Matkul: %s\n", m.MataKuliah))
		}
		if m.Pertemuan > 0 {
			sb.WriteString(fmt.Sprintf("   🔢 Pertemuan: %d\n", m.Pertemuan))
		}
		if m.FileUrl != "" {
			sb.WriteString(fmt.Sprintf("   📥 Unduh: %s\n", m.FileUrl))
		}
		sb.WriteString("\n")

		if count >= 12 {
			sb.WriteString(fmt.Sprintf("_...dan %d modul lainnya._\n\n", len(items)-count))
			break
		}
	}

	if count == 0 {
		return fmt.Sprintf("ℹ️ Tidak ditemukan modul dengan kata kunci *\"%s\"*.", filter)
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
		sb.WriteString(fmt.Sprintf("*%d. %s*\n", i+1, item.Judul))
		sb.WriteString(fmt.Sprintf("%s\n\n", item.Isi))
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
		sb.WriteString(fmt.Sprintf("*%d. %s*\n", i+1, item.Judul))
		if item.Ringkasan != "" {
			sb.WriteString(fmt.Sprintf("   _%s_\n", item.Ringkasan))
		}
		sb.WriteString(fmt.Sprintf("   🔗 %s/berita-praktikum/%s\n\n", webURL, item.Slug))
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
	sb.WriteString("📞 *KONTAK ASISTEN & KOORDINATOR LAB*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for i, item := range items {
		sb.WriteString(fmt.Sprintf("*%d. %s*\n", i+1, item.Nama))
		sb.WriteString(fmt.Sprintf("   🏷️ Posisi/Peran: %s\n", item.Peran))
		if item.Matkul != "" {
			sb.WriteString(fmt.Sprintf("   📚 Mata Kuliah: %s\n", item.Matkul))
		}
		sb.WriteString(fmt.Sprintf("   📱 Kontak: %s\n\n", item.Kontak))
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("🔗 Hubungi via web: %s/kontak", webURL))
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
		sb.WriteString(fmt.Sprintf("*%d. %s*\n", i+1, item.Judul))
		sb.WriteString(fmt.Sprintf("%s\n\n", item.Deskripsi))
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("🔗 Selengkapnya: %s/aturan-praktikum", webURL))
	return sb.String()
}

// FormatFAQ formats FAQ items
func FormatFAQ(items []asciiapi.FAQItem, webURL string) string {
	if len(items) == 0 {
		return "ℹ️ Belum ada daftar FAQ laboratorium."
	}

	var sb strings.Builder
	sb.WriteString("❓ *FAQ (FREQUENTLY ASKED QUESTIONS)*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for i, item := range items {
		if i >= 6 {
			break
		}
		sb.WriteString(fmt.Sprintf("*Q: %s*\n", item.Pertanyaan))
		sb.WriteString(fmt.Sprintf("A: %s\n\n", item.Jawaban))
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
