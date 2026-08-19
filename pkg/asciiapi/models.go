package asciiapi

import "time"

// ScheduleItem represents an entry in the lab schedule table
type ScheduleItem struct {
	Hari       string `json:"Hari"`
	Jam        string `json:"Jam"`
	KodeKelas  string `json:"Kode Kelas"`
	MataKuliah string `json:"Mata Kuliah"`
	Ruang      string `json:"Ruang"`
	Dosen      string `json:"Dosen"`
	Aslab      string `json:"Aslab"`
}

// ClassItem represents class information
type ClassItem struct {
	ID         string `json:"id"`
	KodeKelas  string `json:"Kode Kelas,omitempty"`
	MataKuliah string `json:"Mata Kuliah,omitempty"`
	Ruang      string `json:"Ruang,omitempty"`
	Hari       string `json:"Hari,omitempty"`
	Jam        string `json:"Jam,omitempty"`
	Dosen      string `json:"Dosen,omitempty"`
	Aslab      string `json:"Aslab,omitempty"`
	Semester   string `json:"Semester,omitempty"`
}

// ModulItem represents a practical module
type ModulItem struct {
	ID           string    `json:"id"`
	MataKuliahID string    `json:"mataKuliahId"`
	MataKuliah   string    `json:"mataKuliah,omitempty"`
	Judul        string    `json:"judul"`
	Deskripsi    string    `json:"deskripsi,omitempty"`
	FileUrl      string    `json:"fileUrl,omitempty"`
	Pertemuan    int       `json:"pertemuan,omitempty"`
	CreatedAt    time.Time `json:"createdAt,omitempty"`
}

// AnnouncementItem represents a lab announcement
type AnnouncementItem struct {
	ID        string    `json:"id"`
	Judul     string    `json:"judul"`
	Isi       string    `json:"isi"`
	Tipe      string    `json:"tipe,omitempty"`
	Urutan    int       `json:"urutan,omitempty"`
	Aktif     bool      `json:"aktif,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

// BeritaItem represents news articles on the portal
type BeritaItem struct {
	ID        string    `json:"id"`
	Judul     string    `json:"judul"`
	Slug      string    `json:"slug"`
	Ringkasan string    `json:"ringkasan,omitempty"`
	Konten    string    `json:"konten,omitempty"`
	Gambar    string    `json:"gambar,omitempty"`
	Author    string    `json:"author,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

// ContactItem represents laboratory contacts / assistants
type ContactItem struct {
	ID      string `json:"id"`
	Nama    string `json:"nama"`
	Peran   string `json:"peran"`
	Kontak  string `json:"kontak"`
	Matkul  string `json:"matkul,omitempty"`
	Urutan  int    `json:"urutan,omitempty"`
}

// AturanItem represents lab guidelines and rules
type AturanItem struct {
	ID        string `json:"id"`
	Judul     string `json:"judul"`
	Deskripsi string `json:"deskripsi"`
	Kategori  string `json:"kategori,omitempty"`
}

// FAQItem represents frequently asked questions
type FAQItem struct {
	ID         string `json:"id"`
	Pertanyaan string `json:"pertanyaan"`
	Jawaban    string `json:"jawaban"`
	Kategori   string `json:"kategori,omitempty"`
}
