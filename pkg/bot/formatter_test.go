package bot

import (
	"strings"
	"testing"
	"time"

	"botwa_go_ascii/pkg/asciiapi"
)

func TestFormatHelp(t *testing.T) {
	help := FormatHelp("!", "https://ascii.web.id")
	if !strings.Contains(help, "!jadwal") {
		t.Errorf("expected !jadwal in help output, got: %s", help)
	}
	if !strings.Contains(help, "https://ascii.web.id") {
		t.Errorf("expected web url in help output, got: %s", help)
	}
}

func TestFormatSchedules(t *testing.T) {
	items := []asciiapi.ScheduleItem{
		{
			Course:   "Pemrograman Web",
			Class:    "IF-A",
			Day:      "Senin",
			TimeSlot: 2,
			Location: "Lab 1",
		},
	}
	formatted := FormatSchedules(items, "https://ascii.web.id")
	if !strings.Contains(formatted, "Pemrograman Web") || !strings.Contains(formatted, "IF-A") {
		t.Errorf("unexpected schedule format output: %s", formatted)
	}
}

func TestFormatClasses(t *testing.T) {
	items := []asciiapi.ClassItem{
		{
			KodeKelas:  "IF-B",
			MataKuliah: "Struktur Data",
			NamaKelas:  "Kelas B",
		},
	}
	formatted := FormatClasses(items, "struktur", "https://ascii.web.id")
	if !strings.Contains(formatted, "Struktur Data") {
		t.Errorf("expected Struktur Data in filtered class output, got: %s", formatted)
	}
}

func TestFormatModul(t *testing.T) {
	items := []asciiapi.ModulItem{
		{
			Title:        "Modul 1: Pengenalan HTML",
			MataKuliah:   "Pemrograman Web",
			DriveFileURL: "https://drive.google.com/file/d/xyz/view",
			Tahun:        2026,
		},
	}
	formatted := FormatModul(items, "HTML", "https://ascii.web.id")
	if !strings.Contains(formatted, "Modul 1: Pengenalan HTML") {
		t.Errorf("expected module title in output, got: %s", formatted)
	}
}

func TestFormatPing(t *testing.T) {
	formatted := FormatPing(10*time.Millisecond, 25*time.Millisecond, nil)
	if !strings.Contains(formatted, "PONG!") || !strings.Contains(formatted, "OK") {
		t.Errorf("unexpected ping output: %s", formatted)
	}
}

func TestFormatBimbinganSummary(t *testing.T) {
	data := &asciiapi.BimbinganSummaryResult{
		UserFound: true,
		UserName:  "Kak Budi",
		UserRole:  "aslab",
		Mode:      "mentees",
		Groups: []asciiapi.BimbinganGroupItem{
			{
				ID:             "6fae7264-b049-43c2-bf89-58b292e0717c",
				ShortID:        "6fae72",
				NamaKelompok:   "Kelompok Web 1",
				MataKuliahName: "Pemrograman Web",
				Kelas:          "B1 2024",
				Judul:          "Sistem Portal Lab",
				StatusKonsul0:  "acc",
				StatusKonsul1:  "revisi",
				CatatanKonsul1: "Perbaiki flow diagram",
				StatusKonsul2:  "pending",
				IsAccFinal:     false,
				Members: []asciiapi.BimbinganMemberItem{
					{NIM: "2101552001", Nama: "Andi", Role: "Ketua"},
					{NIM: "2101552002", Nama: "Siti", Role: "Anggota"},
				},
			},
		},
	}

	formatted := FormatBimbinganSummary(data, "!", "https://ascii.web.id")
	if !strings.Contains(formatted, "Kelompok Web 1") || !strings.Contains(formatted, "6fae72") {
		t.Errorf("expected group name and short id, got: %s", formatted)
	}
	if !strings.Contains(formatted, "!revisi") || !strings.Contains(formatted, "!acc") {
		t.Errorf("expected aslab command guide in output, got: %s", formatted)
	}
}

func TestFormatReviewSuccess(t *testing.T) {
	res := &asciiapi.ReviewKonsulResult{
		Success:     true,
		StageName:   "Konsul 1 (Flow Program & Desain)",
		StatusLabel: "⚠️ TELAH DI-ACC",
		AslabName:   "Kak Alex",
		Group: &asciiapi.BimbinganGroupItem{
			NamaKelompok:   "Kelompok Web 1",
			MataKuliahName: "Pemrograman Web",
			Kelas:          "B1 2024",
			Judul:          "Sistem Portal Lab",
			CatatanKonsul1: "Flowchart sudah baik.",
		},
		Recipients: []asciiapi.BimbinganMemberItem{
			{Nama: "Andi", PhoneNumber: "628123456789"},
		},
	}

	formatted := FormatReviewSuccess(res, "https://ascii.web.id")
	if !strings.Contains(formatted, "Kelompok Web 1") || !strings.Contains(formatted, "Flowchart sudah baik.") {
		t.Errorf("expected review success details, got: %s", formatted)
	}
}

func TestFormatStudentReviewNotification(t *testing.T) {
	group := &asciiapi.BimbinganGroupItem{
		NamaKelompok:   "Kelompok Web 1",
		MataKuliahName: "Pemrograman Web",
	}
	msg := FormatStudentReviewNotification(group, "Kak Alex", "Konsul 1", "revisi", "Perbaiki modul auth.", "https://ascii.web.id")
	if !strings.Contains(msg, "Kelompok Web 1") || !strings.Contains(msg, "Perbaiki modul auth.") {
		t.Errorf("expected student review notif content, got: %s", msg)
	}
}

func TestFormatLoginAndProfile(t *testing.T) {
	user := &asciiapi.UserInfo{
		ID:          "user-1",
		Name:        "Kak Alex",
		Username:    "2101552001",
		Role:        "aslab",
		PhoneNumber: "6281234567890",
		Email:       "alex@student.ascii.local",
	}

	loginMsg := FormatLoginSuccess(user, "!", "https://ascii.web.id")
	if !strings.Contains(loginMsg, "Kak Alex") || !strings.Contains(loginMsg, "2101552001") || !strings.Contains(loginMsg, "Asisten Laboratorium") {
		t.Errorf("unexpected login msg: %s", loginMsg)
	}

	profMsg := FormatProfile(user, "!", "https://ascii.web.id")
	if !strings.Contains(profMsg, "Kak Alex") || !strings.Contains(profMsg, "6281234567890") {
		t.Errorf("unexpected profile msg: %s", profMsg)
	}

	logoutMsg := FormatLogoutSuccess(user)
	if !strings.Contains(logoutMsg, "Kak Alex") || !strings.Contains(logoutMsg, "LOGOUT BERHASIL") {
		t.Errorf("unexpected logout msg: %s", logoutMsg)
	}
}
