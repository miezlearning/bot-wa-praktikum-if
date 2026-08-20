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
	sb.WriteString("📋 *Menu & Navigasi Cepat:*\n")
	sb.WriteString(fmt.Sprintf("*[1]* 👥 *Bimbingan Praktikum* (`%sbimbingan`)\n", prefix))
	sb.WriteString("     _Cek kelompok bimbingan & status asistensi_\n")
	sb.WriteString(fmt.Sprintf("*[2]* 📅 *Jadwal Praktikum* (`%sjadwal`)\n", prefix))
	sb.WriteString("     _Cek jadwal aktif praktikum lab_\n")
	sb.WriteString(fmt.Sprintf("*[3]* 🏫 *Pembagian Kelas* (`%skelas`)\n", prefix))
	sb.WriteString("     _Cek kelas, ruangan & aslab pendamping_\n")
	sb.WriteString(fmt.Sprintf("*[4]* 📚 *Modul Praktikum* (`%smodul`)\n", prefix))
	sb.WriteString("     _Cari & unduh berkas modul_\n")
	sb.WriteString(fmt.Sprintf("*[5]* 📢 *Pengumuman Lab* (`%spengumuman`)\n", prefix))
	sb.WriteString("     _Pengumuman lab terkini_\n")
	sb.WriteString(fmt.Sprintf("*[6]* 📰 *Berita Praktikum* (`%sberita`)\n", prefix))
	sb.WriteString("     _Berita & update laboratorium_\n")
	sb.WriteString(fmt.Sprintf("*[7]* 📞 *Kontak Aslab* (`%skontak`)\n", prefix))
	sb.WriteString("     _Daftar kontak asisten lab_\n")
	sb.WriteString(fmt.Sprintf("*[8]* 🔐 *Profil Akun* (`%sprofil`)\n", prefix))
	sb.WriteString("     _Cek akun yang tertaut ke nomor ini_\n\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("✍️ *Perintah Asistensi (Khusus Aslab):*\n")
	sb.WriteString(fmt.Sprintf("• `%srevisi <No> <0/1/2> <catatan>` - Beri revisi\n", prefix))
	sb.WriteString(fmt.Sprintf("• `%sacc <No> <0/1/2> [catatan]` - Beri ACC\n", prefix))
	sb.WriteString(fmt.Sprintf("• `%saccfinal <No> [catatan]` - ACC Final siap demo\n\n", prefix))
	sb.WriteString("🔐 *Autentikasi Akun:*\n")
	sb.WriteString(fmt.Sprintf("• `%slogin <NIM> <password>` - Tautkan akun portal\n", prefix))
	sb.WriteString(fmt.Sprintf("• `%slogout` - Putuskan tautan akun\n\n", prefix))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("👉 _Ketik angka *1* s/d *8* untuk akses langsung!_\n")
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

func stageStatusBadge(status string) string {
	switch strings.ToLower(status) {
	case "acc":
		return "✅ ACC"
	case "revisi":
		return "⚠️ REVISI"
	case "pending":
		return "⏳ Pending"
	default:
		return "⏳ Belum"
	}
}

// FormatBimbinganSummary formats bimbingan list / status
func FormatBimbinganSummary(data *asciiapi.BimbinganSummaryResult, prefix, webURL string) string {
	var sb strings.Builder

	if data.Mode == "not_registered" {
		sb.WriteString("ℹ️ *Nomor WhatsApp Belum Terhubung*\n")
		sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		sb.WriteString("Nomor WhatsApp Anda belum tertaut dengan akun di portal praktikum ASCII.\n\n")
		sb.WriteString("💡 *Cara Menautkan Nomor:*\n")
		sb.WriteString(fmt.Sprintf("1. Login ke portal web: %s\n", webURL))
		sb.WriteString("2. Buka menu *Pengaturan Profil* dan masukkan nomor WhatsApp Anda.\n\n")
		sb.WriteString("🔍 *Atau cari bimbingan dengan NIM/Nama Kelompok:*\n")
		sb.WriteString(fmt.Sprintf("Ketik: `%sbimbingan <NIM / Nama Kelompok>`\n", prefix))
		sb.WriteString(fmt.Sprintf("Contoh: `%sbimbingan 2101552001`\n", prefix))
		sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		return sb.String()
	}

	if data.Mode == "search" {
		sb.WriteString(fmt.Sprintf("🔍 *HASIL PENCARIAN BIMBINGAN: \"%s\"*\n", data.SearchQuery))
		sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	} else if data.Mode == "mentees" {
		sb.WriteString("📋 *DAFTAR KELOMPOK BIMBINGAN*\n")
		if data.UserName != "" {
			sb.WriteString(fmt.Sprintf("👤 *Asisten:* Kak %s\n", data.UserName))
		}
		sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	} else {
		sb.WriteString("📋 *STATUS BIMBINGAN PRAKTIKUM*\n")
		if data.UserName != "" {
			sb.WriteString(fmt.Sprintf("👤 *Praktikan:* %s\n", data.UserName))
		}
		sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	}

	if len(data.Groups) == 0 {
		if data.Mode == "search" {
			sb.WriteString(fmt.Sprintf("ℹ️ Tidak ditemukan kelompok bimbingan dengan kata kunci *\"%s\"*.\n", data.SearchQuery))
			sb.WriteString(fmt.Sprintf("Coba gunakan NIM atau kata kunci lain: `%sbimbingan <kata kunci>`", prefix))
		} else if data.Mode == "mentees" {
			sb.WriteString("ℹ️ Belum ada kelompok bimbingan yang terdaftar di bawah bimbingan Anda saat ini.\n")
			sb.WriteString(fmt.Sprintf("🔗 Buka portal untuk cek slot: %s/bimbingan", webURL))
		} else {
			sb.WriteString("ℹ️ Anda belum terdaftar dalam kelompok bimbingan manapun.\n")
			sb.WriteString(fmt.Sprintf("🔗 Daftarkan kelompok bimbingan di portal: %s/bimbingan", webURL))
		}
		return sb.String()
	}

	for i, g := range data.Groups {
		sb.WriteString(fmt.Sprintf("*%d. %s* (ID: `%s`)\n", i+1, g.NamaKelompok, g.ShortID))
		sb.WriteString(fmt.Sprintf("   📖 *Mata Kuliah:* %s (%s)\n", g.MataKuliahName, g.Kelas))
		sb.WriteString(fmt.Sprintf("   📌 *Judul:* %s\n", g.Judul))
		if g.AslabName != "" {
			aslabInfo := "Kak " + g.AslabName
			cleanPhone := asciiapi.NormalizePhoneNumber(g.AslabPhoneNumber)
			if cleanPhone != "" && !strings.HasPrefix(cleanPhone, "15") && len(cleanPhone) <= 14 {
				aslabInfo += fmt.Sprintf(" (wa.me/%s)", cleanPhone)
			}
			sb.WriteString(fmt.Sprintf("   👨‍🏫 *Aslab:* %s\n", aslabInfo))
		}

		if len(g.Members) > 0 {
			sb.WriteString("   👥 *Anggota:*\n")
			for _, m := range g.Members {
				roleLabel := ""
				if m.Role == "Ketua" {
					roleLabel = " *(Ketua)*"
				}
				sb.WriteString(fmt.Sprintf("      • %s (%s)%s\n", m.Nama, m.NIM, roleLabel))
			}
		}

		accFinalBadge := "⏳ Belum"
		if g.IsAccFinal {
			accFinalBadge = "🎯 *ACC FINAL (SIAP DEMO)*"
		}

		sb.WriteString("   📊 *Tahap Asistensi:*\n")
		sb.WriteString(fmt.Sprintf("      • Konsul 0 (Konsep): %s\n", stageStatusBadge(g.StatusKonsul0)))
		if g.CatatanKonsul0 != "" {
			sb.WriteString(fmt.Sprintf("        📝 Catatan: _\"%s\"_\n", g.CatatanKonsul0))
		}
		sb.WriteString(fmt.Sprintf("      • Konsul 1 (Flowchart): %s\n", stageStatusBadge(g.StatusKonsul1)))
		if g.CatatanKonsul1 != "" {
			sb.WriteString(fmt.Sprintf("        📝 Catatan: _\"%s\"_\n", g.CatatanKonsul1))
		}
		sb.WriteString(fmt.Sprintf("      • Konsul 2 (70%% Koding): %s\n", stageStatusBadge(g.StatusKonsul2)))
		if g.CatatanKonsul2 != "" {
			sb.WriteString(fmt.Sprintf("        📝 Catatan: _\"%s\"_\n", g.CatatanKonsul2))
		}
		sb.WriteString(fmt.Sprintf("      • ACC Final: %s\n", accFinalBadge))

		if g.FlowURL != "" {
			sb.WriteString(fmt.Sprintf("   🔗 *Flow:* %s\n", g.FlowURL))
		}
		if g.RepoURL != "" {
			sb.WriteString(fmt.Sprintf("   🔗 *Repo:* %s\n", g.RepoURL))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	if data.Mode == "mentees" || data.UserRole == "aslab" || data.UserRole == "pengurus" || data.UserRole == "koordinator" {
		sb.WriteString("💡 *Aksi Aslab via Chat Bot:*\n")
		sb.WriteString(fmt.Sprintf("• `%srevisi <No/ID> <Tahap: 0/1/2> <Catatan>`\n", prefix))
		sb.WriteString(fmt.Sprintf("  _Contoh:_ `%srevisi 1 1 Flowchart perbaiki alur autentikasi`\n", prefix))
		sb.WriteString(fmt.Sprintf("• `%sacc <No/ID> <Tahap: 0/1/2> [Catatan]`\n", prefix))
		sb.WriteString(fmt.Sprintf("  _Contoh:_ `%sacc 1 1 Flowchart sudah baik, lanjut koding 70%%`\n", prefix))
		sb.WriteString(fmt.Sprintf("• `%saccfinal <No/ID> [Catatan]`\n", prefix))
		sb.WriteString(fmt.Sprintf("  _Contoh:_ `%saccfinal 1 Selamat siap untuk demo!`\n\n", prefix))
	}

	sb.WriteString(fmt.Sprintf("🔗 Portal Bimbingan: %s/bimbingan", webURL))
	return sb.String()
}

// FormatReviewSuccess formats the response when a review (revisi or acc) is successfully recorded
func FormatReviewSuccess(res *asciiapi.ReviewKonsulResult, webURL string) string {
	var sb strings.Builder
	if res.StatusLabel == "✅ TELAH DI-ACC" {
		sb.WriteString("🎉 *ASISTENSI BERHASIL DI-ACC!*\n")
	} else {
		sb.WriteString("📝 *CATATAN REVISI BERHASIL DIKIRIM!*\n")
	}
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	if res.Group != nil {
		sb.WriteString(fmt.Sprintf("📋 *Kelompok:* %s (%s)\n", res.Group.NamaKelompok, res.Group.Kelas))
		sb.WriteString(fmt.Sprintf("📖 *Mata Kuliah:* %s\n", res.Group.MataKuliahName))
		sb.WriteString(fmt.Sprintf("📌 *Judul:* %s\n", res.Group.Judul))
	}

	sb.WriteString(fmt.Sprintf("📌 *Tahapan:* %s\n", res.StageName))
	sb.WriteString(fmt.Sprintf("📊 *Status Baru:* %s\n", res.StatusLabel))

	var catatan string
	if res.Group != nil {
		if strings.Contains(res.StageName, "0") {
			catatan = res.Group.CatatanKonsul0
		} else if strings.Contains(res.StageName, "1") {
			catatan = res.Group.CatatanKonsul1
		} else if strings.Contains(res.StageName, "2") {
			catatan = res.Group.CatatanKonsul2
		}
	}

	if catatan != "" {
		sb.WriteString(fmt.Sprintf("📝 *Catatan:* \"%s\"\n\n", catatan))
	} else {
		sb.WriteString("\n")
	}

	if len(res.Recipients) > 0 {
		sb.WriteString(fmt.Sprintf("📢 *Notifikasi WhatsApp:* Berhasil dikirimkan otomatis ke %d anggota kelompok.\n\n", len(res.Recipients)))
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("🔗 Dashboard: %s/bimbingan", webURL))
	return sb.String()
}

// FormatAccFinalSuccess formats the response when ACC final is given
func FormatAccFinalSuccess(res *asciiapi.ReviewKonsulResult, webURL string) string {
	var sb strings.Builder
	sb.WriteString("🎯 *PROYEK DINYATAKAN ACC FINAL (SIAP DEMO)!*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	if res.Group != nil {
		sb.WriteString(fmt.Sprintf("📋 *Kelompok:* %s (%s)\n", res.Group.NamaKelompok, res.Group.Kelas))
		sb.WriteString(fmt.Sprintf("📖 *Mata Kuliah:* %s\n", res.Group.MataKuliahName))
		sb.WriteString(fmt.Sprintf("📌 *Judul:* %s\n\n", res.Group.Judul))
	}

	sb.WriteString("✅ Seluruh tahapan asistensi (Konsul 0, 1, dan 2) otomatis dinyatakan *ACC*.\n")
	if len(res.Recipients) > 0 {
		sb.WriteString(fmt.Sprintf("📢 *Notifikasi WhatsApp:* Telah dikirimkan otomatis ke %d anggota kelompok.\n\n", len(res.Recipients)))
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("🔗 Dashboard: %s/bimbingan", webURL))
	return sb.String()
}

// FormatStudentReviewNotification formats official notification sent to students
func FormatStudentReviewNotification(group *asciiapi.BimbinganGroupItem, aslabName, stageName, status, catatan, webURL string) string {
	statusLabel := "⚠️ *PERLU REVISI*"
	if status == "acc" {
		statusLabel = "✅ *TELAH DI-ACC*"
	}

	timeWIB := time.Now().Format("15:04")
	ref := strings.ToUpper(fmt.Sprintf("%x", time.Now().UnixNano()%0xFFFF))

	var sb strings.Builder
	sb.WriteString("📢 *Update Hasil Bimbingan Praktikum ASCII*\n\n")
	sb.WriteString(fmt.Sprintf("Halo Rekan Mahasiswa (*%s*),\n", group.NamaKelompok))
	sb.WriteString(fmt.Sprintf("Asisten pembimbing Kak *%s* telah memberikan catatan untuk tahapan *%s* (%s):\n\n", aslabName, stageName, group.MataKuliahName))
	sb.WriteString(fmt.Sprintf("Status: %s\n", statusLabel))
	if catatan != "" {
		sb.WriteString(fmt.Sprintf("📝 *Catatan Aslab:*\n\"%s\"\n\n", catatan))
	} else {
		sb.WriteString("\n")
	}
	sb.WriteString("Cek detail dan tindak lanjut pada portal:\n")
	sb.WriteString(fmt.Sprintf("🔗 %s/bimbingan\n\n", webURL))
	sb.WriteString(fmt.Sprintf("_Ref: #%s • %s WIB_\n", ref, timeWIB))
	sb.WriteString("_Pesan otomatis dari Sistem Portal Lab ASCII Informatika_")
	return sb.String()
}

// FormatStudentAccFinalNotification formats official notification sent to students on ACC final
func FormatStudentAccFinalNotification(group *asciiapi.BimbinganGroupItem, aslabName, catatan, webURL string) string {
	timeWIB := time.Now().Format("15:04")
	ref := strings.ToUpper(fmt.Sprintf("%x", time.Now().UnixNano()%0xFFFF))

	var sb strings.Builder
	sb.WriteString("📢 *Update Hasil Bimbingan Praktikum ASCII*\n\n")
	sb.WriteString(fmt.Sprintf("Halo Rekan Mahasiswa (*%s*),\n", group.NamaKelompok))
	sb.WriteString(fmt.Sprintf("Asisten pembimbing Kak *%s* telah menyatakan bimbingan proyek kelompok Anda:\n\n", aslabName))
	sb.WriteString("Status: 🎯 *ACC FINAL (SIAP DEMO)*\n")
	if catatan != "" {
		sb.WriteString(fmt.Sprintf("📝 *Catatan Aslab:*\n\"%s\"\n\n", catatan))
	} else {
		sb.WriteString("📝 *Catatan Aslab:*\n\"Selamat! Proyek bimbingan kelompok Anda telah disetujui ACC Final dan siap untuk jadwal demo praktikum.\"\n\n")
	}
	sb.WriteString("Cek detail dan tindak lanjut pada portal:\n")
	sb.WriteString(fmt.Sprintf("🔗 %s/bimbingan\n\n", webURL))
	sb.WriteString(fmt.Sprintf("_Ref: #%s • %s WIB_\n", ref, timeWIB))
	sb.WriteString("_Pesan otomatis dari Sistem Portal Lab ASCII Informatika_")
	return sb.String()
}

func roleLabel(role string) string {
	switch role {
	case "aslab":
		return "👨‍🏫 Asisten Laboratorium"
	case "pengurus":
		return "👑 Pengurus Lab"
	case "koordinator":
		return "🎖️ Koordinator Praktikum"
	case "praktikan":
		return "🎓 Praktikan / Mahasiswa"
	default:
		return "👤 Pengguna"
	}
}

// FormatLoginSuccess formats successful login response
func FormatLoginSuccess(u *asciiapi.UserInfo, prefix, webURL string) string {
	var sb strings.Builder
	sb.WriteString("✅ *LOGIN BERHASIL!*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("Halo, *%s*! Akun Anda telah berhasil ditautkan ke WhatsApp ini.\n\n", u.Name))
	sb.WriteString(fmt.Sprintf("📋 *Detail Akun:*\n"))
	sb.WriteString(fmt.Sprintf("• *Nama Lengkap:* %s\n", u.Name))
	sb.WriteString(fmt.Sprintf("• *NIM / Username:* %s\n", u.Username))
	sb.WriteString(fmt.Sprintf("• *Peran (Role):* %s\n", roleLabel(u.Role)))
	if u.Email != "" {
		sb.WriteString(fmt.Sprintf("• *Email:* %s\n", u.Email))
	}
	if u.PhoneNumber != "" && !strings.HasPrefix(u.PhoneNumber, "15") && len(u.PhoneNumber) <= 14 {
		sb.WriteString(fmt.Sprintf("• *Nomor HP Terdaftar:* +%s\n", u.PhoneNumber))
	}
	sb.WriteString("• *Status WhatsApp:* 🟢 Terhubung & Aktif\n\n")

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("💡 *Perintah yang Dapat Digunakan:*\n")
	if u.Role == "aslab" || u.Role == "pengurus" || u.Role == "koordinator" {
		sb.WriteString(fmt.Sprintf("• `%s1` atau `%sbimbingan` - Lihat seluruh kelompok yang Anda bimbing\n", prefix, prefix))
		sb.WriteString(fmt.Sprintf("• `%srevisi <No> <Tahap> <Catatan>` - Beri revisi bimbingan\n", prefix))
		sb.WriteString(fmt.Sprintf("• `%sacc <No> <Tahap>` - Beri ACC asistensi\n", prefix))
		sb.WriteString(fmt.Sprintf("• `%saccfinal <No>` - ACC Final proyek praktikum\n", prefix))
	} else {
		sb.WriteString(fmt.Sprintf("• `%s1` atau `%sbimbingan` - Cek status kelompok & progres asistensi Anda\n", prefix, prefix))
	}
	sb.WriteString(fmt.Sprintf("• `%s8` atau `%sprofil` - Cek status profil akun WhatsApp ini\n", prefix, prefix))
	sb.WriteString(fmt.Sprintf("• `%slogout` - Putuskan tautan akun dari nomor ini\n\n", prefix))
	sb.WriteString(fmt.Sprintf("🔗 Portal Lab: %s", webURL))
	return sb.String()
}

// FormatLogoutSuccess formats logout message
func FormatLogoutSuccess(u *asciiapi.UserInfo) string {
	var sb strings.Builder
	sb.WriteString("🚪 *LOGOUT BERHASIL*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("Tautan akun *%s* (%s) dengan WhatsApp ini telah berhasil dilepas.\n\n", u.Name, u.Username))
	sb.WriteString("Ketik `!login <NIM> <password>` kapan saja jika ingin menghubungkan kembali akun Anda.\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return sb.String()
}

// FormatProfile formats profile status message
func FormatProfile(u *asciiapi.UserInfo, prefix, webURL string) string {
	var sb strings.Builder
	if u == nil {
		sb.WriteString("ℹ️ *BELUM ADA AKUN TERHUBUNG*\n")
		sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		sb.WriteString("WhatsApp ini belum tertaut dengan akun di portal praktikum ASCII.\n\n")
		sb.WriteString("🔐 *Cara Menghubungkan Akun:*\n")
		sb.WriteString(fmt.Sprintf("Ketik: `%slogin <NIM> <password>`\n", prefix))
		sb.WriteString(fmt.Sprintf("Contoh: `%slogin 2101552001 Password123`\n\n", prefix))
		sb.WriteString(fmt.Sprintf("🔗 Portal Web: %s\n", webURL))
		sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		return sb.String()
	}

	sb.WriteString("👤 *PROFIL AKUN TERTAUT*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("• *Nama:* %s\n", u.Name))
	sb.WriteString(fmt.Sprintf("• *NIM / Username:* %s\n", u.Username))
	sb.WriteString(fmt.Sprintf("• *Peran (Role):* %s\n", roleLabel(u.Role)))
	if u.Email != "" {
		sb.WriteString(fmt.Sprintf("• *Email:* %s\n", u.Email))
	}
	if u.PhoneNumber != "" && !strings.HasPrefix(u.PhoneNumber, "15") && len(u.PhoneNumber) <= 14 {
		sb.WriteString(fmt.Sprintf("• *Nomor HP Terdaftar:* +%s\n", u.PhoneNumber))
	}
	sb.WriteString("• *Status WhatsApp:* 🟢 Terhubung & Aktif\n\n")

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("• Ketik `1` atau `%sbimbingan` untuk melihat data bimbingan.\n", prefix))
	sb.WriteString(fmt.Sprintf("• Ketik `%slogout` untuk memutuskan tautan akun dari nomor ini.\n", prefix))
	return sb.String()
}
