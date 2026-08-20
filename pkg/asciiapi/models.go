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

// BimbinganMemberItem represents a member of a bimbingan group
type BimbinganMemberItem struct {
	ID          string `json:"id"`
	GroupID     string `json:"groupId"`
	UserID      string `json:"userId,omitempty"`
	NIM         string `json:"nim"`
	Nama        string `json:"nama"`
	Role        string `json:"role"` // "Ketua" or "Anggota"
	PhoneNumber string `json:"phoneNumber,omitempty"`
}

// BimbinganGroupItem represents a practical bimbingan group
type BimbinganGroupItem struct {
	ID                  string                `json:"id"`
	ShortID             string                `json:"shortId,omitempty"`
	MataKuliahID        string                `json:"mataKuliahId"`
	MataKuliahName      string                `json:"mataKuliahName"`
	MataKuliahSingkatan string                `json:"mataKuliahSingkatan,omitempty"`
	Kelas               string                `json:"kelas"`
	NamaKelompok        string                `json:"namaKelompok"`
	Judul               string                `json:"judul"`
	Deskripsi           string                `json:"deskripsi"`
	SlotID              string                `json:"slotId,omitempty"`
	FlowURL             string                `json:"flowUrl,omitempty"`
	RepoURL             string                `json:"repoUrl,omitempty"`
	StatusKonsul0       string                `json:"statusKonsul0"` // "pending", "revisi", "acc"
	CatatanKonsul0      string                `json:"catatanKonsul0,omitempty"`
	StatusKonsul1       string                `json:"statusKonsul1"`
	CatatanKonsul1      string                `json:"catatanKonsul1,omitempty"`
	StatusKonsul2       string                `json:"statusKonsul2"`
	CatatanKonsul2      string                `json:"catatanKonsul2,omitempty"`
	IsAccFinal          bool                  `json:"isAccFinal"`
	LastNotifiedAt      *time.Time            `json:"lastNotifiedAt,omitempty"`
	CreatedAt           time.Time             `json:"createdAt,omitempty"`
	UpdatedAt           time.Time             `json:"updatedAt,omitempty"`
	AslabID             string                `json:"aslabId,omitempty"`
	AslabName           string                `json:"aslabName,omitempty"`
	AslabPhoneNumber    string                `json:"aslabPhoneNumber,omitempty"`
	Members             []BimbinganMemberItem `json:"members,omitempty"`
}

// BimbinganSummaryResult contains resolved user info and associated bimbingan groups
type BimbinganSummaryResult struct {
	UserFound   bool                 `json:"userFound"`
	UserID      string               `json:"userId,omitempty"`
	UserName    string               `json:"userName,omitempty"`
	UserRole    string               `json:"userRole,omitempty"` // "aslab", "pengurus", "koordinator", "praktikan", "guest"
	UserPhone   string               `json:"userPhone,omitempty"`
	Mode        string               `json:"mode"` // "mentees", "my_group", "search", "not_registered"
	SearchQuery string               `json:"searchQuery,omitempty"`
	Groups      []BimbinganGroupItem `json:"groups"`
}

// ReviewKonsulParams contains arguments for reviewing a bimbingan stage
type ReviewKonsulParams struct {
	SenderPhone string `json:"senderPhone"`
	GroupID     string `json:"groupId"`
	Stage       int    `json:"stage"` // 0, 1, 2
	Status      string `json:"status"` // "acc", "revisi", "pending"
	Catatan     string `json:"catatan"`
}

// ReviewKonsulResult contains the result of a review operation
type ReviewKonsulResult struct {
	Success     bool                  `json:"success"`
	Message     string                `json:"message"`
	StageName   string                `json:"stageName"`
	StatusLabel string                `json:"statusLabel"`
	Group       *BimbinganGroupItem   `json:"group,omitempty"`
	AslabName   string                `json:"aslabName,omitempty"`
	Recipients  []BimbinganMemberItem `json:"recipients,omitempty"`
}

// AccFinalParams contains arguments for approving the final project
type AccFinalParams struct {
	SenderPhone string `json:"senderPhone"`
	GroupID     string `json:"groupId"`
	IsAccFinal  bool   `json:"isAccFinal"`
	Catatan     string `json:"catatan,omitempty"`
}
