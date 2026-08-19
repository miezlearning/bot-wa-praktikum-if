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
