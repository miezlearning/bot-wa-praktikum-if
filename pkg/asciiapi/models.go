package asciiapi

import "time"

// ScheduleItem represents an entry in the lab schedule table
type ScheduleItem struct {
	Class     string `json:"class"`
	Course    string `json:"course"`
	Day       string `json:"day"`
	TimeSlot  int    `json:"time"`
	Location  string `json:"location"`

	// Fallback keys if raw table format is returned
	Hari       string `json:"Hari,omitempty"`
	Jam        string `json:"Jam,omitempty"`
	KodeKelas  string `json:"Kode Kelas,omitempty"`
	MataKuliah string `json:"Mata Kuliah,omitempty"`
	Tempat     string `json:"Tempat,omitempty"`
}

// ClassItem represents class information
type ClassItem struct {
	KodeKelas  string `json:"Kode Kelas"`
	NamaKelas  string `json:"Nama Kelas,omitempty"`
	MataKuliah string `json:"Mata Kuliah"`
	IsPilihan  bool   `json:"isPilihan,omitempty"`
}

// ModulItem represents a practical module
type ModulItem struct {
	ID           string    `json:"id"`
	MataKuliahID string    `json:"mataKuliahId"`
	MataKuliah   string    `json:"mataKuliah,omitempty"` // populated via lookup if available
	Tahun        int       `json:"tahun"`
	Title        string    `json:"title"`
	Description  string    `json:"description,omitempty"`
	DriveFileURL string    `json:"driveFileUrl"`
	Type         string    `json:"type"` // "pdf" or "markdown"
	Urutan       int       `json:"urutan"`
	CreatedAt    time.Time `json:"createdAt,omitempty"`
}

// MataKuliahItem represents a course
type MataKuliahItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Singkatan string `json:"singkatan,omitempty"`
	IsLegend  bool   `json:"isLegend,omitempty"`
	IsPilihan bool   `json:"isPilihan,omitempty"`
}

// AnnouncementItem represents a lab announcement
type AnnouncementItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link,omitempty"`
	SortOrder   int       `json:"sortOrder,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
}

// BeritaItem represents news articles on the portal
type BeritaItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	Content     string    `json:"content,omitempty"`
	ImageURL    string    `json:"imageUrl,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
}

// ContactItem represents laboratory contacts / links
type ContactItem struct {
	ID          string `json:"id"`
	Platform    string `json:"platform"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ButtonText  string `json:"buttonText,omitempty"`
	URL         string `json:"url"`
	SortOrder   int    `json:"sortOrder,omitempty"`
}

// AturanItem represents lab guidelines and rules
type AturanItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Content     string `json:"content"`
}

// FAQItem represents frequently asked questions
type FAQItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder,omitempty"`
}
