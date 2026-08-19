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
			MataKuliah: "Pemrograman Web",
			KodeKelas:  "IF-A",
			Hari:       "Senin",
			Jam:        "08:00 - 10:00",
			Ruang:      "Lab 1",
			Aslab:      "Miez",
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
			Hari:       "Selasa",
			Jam:        "10:00 - 12:00",
			Aslab:      "Alex",
		},
	}
	formatted := FormatClasses(items, "struktur", "https://ascii.web.id")
	if !strings.Contains(formatted, "Struktur Data") {
		t.Errorf("expected Struktur Data in filtered class output, got: %s", formatted)
	}
}

func TestFormatPing(t *testing.T) {
	formatted := FormatPing(10*time.Millisecond, 25*time.Millisecond, nil)
	if !strings.Contains(formatted, "PONG!") || !strings.Contains(formatted, "OK") {
		t.Errorf("unexpected ping output: %s", formatted)
	}
}
