package asciiapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type UserInfo struct {
	ID          string
	Name        string
	Email       string
	Username    string
	Role        string
	PhoneNumber string
}

// NormalizePhoneNumber normalizes phone number to standard international format (e.g. 628123456789)
func NormalizePhoneNumber(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// Remove JID server suffix
	if idx := strings.Index(raw, "@"); idx != -1 {
		raw = raw[:idx]
	}

	// Remove non-digit characters
	reg := regexp.MustCompile(`\D`)
	cleaned := reg.ReplaceAllString(raw, "")
	if cleaned == "" {
		return ""
	}

	if strings.HasPrefix(cleaned, "0") {
		cleaned = "62" + cleaned[1:]
	} else if strings.HasPrefix(cleaned, "8") {
		cleaned = "62" + cleaned
	}

	return cleaned
}

func (c *Client) getDB() (*sql.DB, error) {
	dbPath := c.dbPath
	if dbPath == "" {
		dbPath = "../ascii-if/local.db"
	}

	// Check if file exists or resolve from candidate paths
	candidates := []string{
		dbPath,
		"../ascii-if/local.db",
		"ascii-if/local.db",
		"./local.db",
		"./data/local.db",
		"/data/local.db",
		"/app/local.db",
		filepath.Join(os.Getenv("USERPROFILE"), "Documents", "Programming", "Web Ascii", "ascii-if", "local.db"),
	}

	foundPath := ""
	for _, cand := range candidates {
		if cand == "" {
			continue
		}
		if _, err := os.Stat(cand); err == nil {
			foundPath = cand
			break
		}
	}

	// If neither exists, fallback to ./local.db or dbPath and create directory if needed
	if foundPath == "" {
		foundPath = dbPath
		if foundPath == "../ascii-if/local.db" {
			foundPath = "./local.db"
		}
		if dir := filepath.Dir(foundPath); dir != "." && dir != "" {
			_ = os.MkdirAll(dir, 0755)
		}
	}

	db, err := sql.Open("sqlite", foundPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db at %s: %w", foundPath, err)
	}

	// Ensure required tables exist for standalone VPS deployment and session pairing
	_, _ = db.Exec(`
		CREATE TABLE IF NOT EXISTS user (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT,
			username TEXT,
			role TEXT DEFAULT 'praktikan',
			phone_number TEXT,
			password_hash TEXT,
			created_at INTEGER,
			updated_at INTEGER
		);

		CREATE TABLE IF NOT EXISTS bot_wa_sessions (
			sender_id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		ALTER TABLE bot_wa_sessions ADD COLUMN token TEXT;

		CREATE TABLE IF NOT EXISTS mata_kuliah (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			singkatan TEXT,
			is_legend INTEGER DEFAULT 0,
			is_pilihan INTEGER DEFAULT 0,
			created_at INTEGER,
			updated_at INTEGER
		);

		CREATE TABLE IF NOT EXISTS bimbingan_slot (
			id TEXT PRIMARY KEY,
			aslab_id TEXT,
			mata_kuliah_id TEXT,
			kelas TEXT,
			kuota INTEGER DEFAULT 3,
			catatan_keahlian TEXT,
			created_at INTEGER,
			updated_at INTEGER
		);

		CREATE TABLE IF NOT EXISTS bimbingan_group (
			id TEXT PRIMARY KEY,
			mata_kuliah_id TEXT,
			kelas TEXT,
			nama_kelompok TEXT,
			judul TEXT,
			deskripsi TEXT,
			slot_id TEXT,
			flow_url TEXT,
			repo_url TEXT,
			status_konsul_0 TEXT DEFAULT 'pending',
			catatan_konsul_0 TEXT,
			status_konsul_1 TEXT DEFAULT 'pending',
			catatan_konsul_1 TEXT,
			status_konsul_2 TEXT DEFAULT 'pending',
			catatan_konsul_2 TEXT,
			is_acc_final INTEGER DEFAULT 0,
			last_notified_at INTEGER,
			created_at INTEGER,
			updated_at INTEGER
		);

		CREATE TABLE IF NOT EXISTS bimbingan_member (
			id TEXT PRIMARY KEY,
			group_id TEXT NOT NULL,
			user_id TEXT,
			nim TEXT,
			nama TEXT,
			role TEXT DEFAULT 'Anggota',
			created_at INTEGER,
			updated_at INTEGER
		);

		UPDATE user SET phone_number = NULL WHERE phone_number LIKE '15%' AND length(phone_number) >= 14;
	`)

	// Safe column migrations in case tables were created with legacy columns
	_, _ = db.Exec("ALTER TABLE mata_kuliah ADD COLUMN name TEXT;")
	_, _ = db.Exec("UPDATE mata_kuliah SET name = nama WHERE (name IS NULL OR name = '') AND nama IS NOT NULL;")

	return db, nil
}

// FindUserByPhone finds a user by normalized phone number or WhatsApp LID session
func (c *Client) FindUserByPhone(phone string) (*UserInfo, error) {
	norm := NormalizePhoneNumber(phone)
	raw := strings.TrimSpace(phone)
	if raw == "" && norm == "" {
		return nil, fmt.Errorf("nomor telepon kosong atau tidak valid")
	}

	db, err := c.getDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var u UserInfo

	// 1. Check in bot_wa_sessions first (supports WhatsApp LID and linked phones)
	sessionQuery := `
		SELECT u.id, u.name, u.email, COALESCE(u.username, ''), u.role, COALESCE(u.phone_number, '')
		FROM bot_wa_sessions s
		JOIN user u ON s.user_id = u.id
		WHERE s.sender_id = ? OR s.sender_id = ?
		LIMIT 1
	`
	err = db.QueryRow(sessionQuery, raw, norm).Scan(
		&u.ID, &u.Name, &u.Email, &u.Username, &u.Role, &u.PhoneNumber,
	)
	if err == nil {
		return &u, nil
	}

	// 2. Check candidate phone formats in DB: "628...", "08...", "+628..."
	var localFormat string
	if strings.HasPrefix(norm, "62") {
		localFormat = "0" + norm[2:]
	}
	plusFormat := "+" + norm

	query := `
		SELECT id, name, email, COALESCE(username, ''), role, COALESCE(phone_number, '')
		FROM user
		WHERE phone_number = ? OR phone_number = ? OR phone_number = ? OR phone_number = ?
		LIMIT 1
	`

	err = db.QueryRow(query, norm, localFormat, plusFormat, raw).Scan(
		&u.ID, &u.Name, &u.Email, &u.Username, &u.Role, &u.PhoneNumber,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &u, nil
}

func (c *Client) getUserSession(phone string) (string, *UserInfo) {
	caller, err := c.FindUserByPhone(phone)
	if err != nil || caller == nil {
		return "", caller
	}

	db, err := c.getDB()
	if err != nil {
		return "", caller
	}
	defer db.Close()

	var token sql.NullString
	raw := strings.TrimSpace(phone)
	norm := NormalizePhoneNumber(phone)
	_ = db.QueryRow("SELECT token FROM bot_wa_sessions WHERE sender_id = ? OR sender_id = ? LIMIT 1", raw, norm).Scan(&token)

	if token.Valid && token.String != "" {
		return token.String, caller
	}
	return "", caller
}

// GetBimbinganSummary retrieves bimbingan summary for the sender or search query
func (c *Client) GetBimbinganSummary(phone, query string) (*BimbinganSummaryResult, error) {
	query = strings.TrimSpace(query)
	normPhone := NormalizePhoneNumber(phone)
	token, caller := c.getUserSession(phone)

	result := &BimbinganSummaryResult{
		Groups: make([]BimbinganGroupItem, 0),
	}

	if caller != nil {
		result.UserFound = true
		result.UserID = caller.ID
		result.UserName = caller.Name
		result.UserRole = caller.Role
		result.UserPhone = caller.PhoneNumber
	}

	// 1. Try Live Web API first
	if (token != "" || c.apiKey != "") && query == "" && caller != nil {
		endpoint := "/bimbingan/mentees"
		if caller.Role == "praktikan" {
			endpoint = "/bimbingan/my"
		}
		respBytes, err := c.doRequestWithAuth(http.MethodGet, endpoint, nil, token)
		if err == nil {
			var groups []BimbinganGroupItem
			if err := json.Unmarshal(respBytes, &groups); err == nil && len(groups) > 0 {
				for i := range groups {
					if len(groups[i].ID) >= 6 {
						groups[i].ShortID = groups[i].ID[:6]
					} else {
						groups[i].ShortID = groups[i].ID
					}
				}
				result.Groups = groups
				if caller.Role == "aslab" || caller.Role == "pengurus" || caller.Role == "koordinator" {
					result.Mode = "mentees"
				} else {
					result.Mode = "my_group"
				}
				return result, nil
			}
		}
	}

	// 2. Fallback to Local SQLite DB
	db, err := c.getDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Case 1: Search query provided
	if query != "" {
		result.Mode = "search"
		result.SearchQuery = query

		groups, err := c.searchGroupsDB(db, query)
		if err != nil {
			return nil, err
		}
		result.Groups = groups
		return result, nil
	}

	// Case 2: No query, look up by caller phone
	if caller == nil {
		result.Mode = "not_registered"
		result.UserPhone = normPhone
		return result, nil
	}

	// Case 2A: Aslab / Pengurus / Koordinator -> Fetch mentees
	if caller.Role == "aslab" || caller.Role == "pengurus" || caller.Role == "koordinator" {
		result.Mode = "mentees"
		groups, err := c.getAslabMenteesDB(db, caller.ID)
		if err != nil {
			return nil, err
		}

		// If pengurus/koordinator and no directly mentored groups, fetch recent active groups
		if len(groups) == 0 && (caller.Role == "pengurus" || caller.Role == "koordinator") {
			groups, err = c.getAllActiveGroupsDB(db)
			if err != nil {
				return nil, err
			}
		}

		result.Groups = groups
		return result, nil
	}

	// Case 2B: Praktikan -> Fetch their own groups
	result.Mode = "my_group"
	groups, err := c.getMyGroupsDB(db, caller.ID, caller.Username, normPhone)
	if err != nil {
		return nil, err
	}
	result.Groups = groups
	return result, nil
}

func (c *Client) getAslabMenteesDB(db *sql.DB, aslabID string) ([]BimbinganGroupItem, error) {
	query := `
		SELECT 
			bg.id, bg.mata_kuliah_id, COALESCE(mk.name, ''), COALESCE(mk.singkatan, ''),
			bg.kelas, bg.nama_kelompok, bg.judul, bg.deskripsi,
			COALESCE(bg.slot_id, ''), COALESCE(bg.flow_url, ''), COALESCE(bg.repo_url, ''),
			bg.status_konsul_0, COALESCE(bg.catatan_konsul_0, ''),
			bg.status_konsul_1, COALESCE(bg.catatan_konsul_1, ''),
			bg.status_konsul_2, COALESCE(bg.catatan_konsul_2, ''),
			bg.is_acc_final, bg.created_at, bg.updated_at,
			COALESCE(u.id, ''), COALESCE(u.name, ''), COALESCE(u.phone_number, '')
		FROM bimbingan_group bg
		INNER JOIN bimbingan_slot bs ON bg.slot_id = bs.id
		INNER JOIN mata_kuliah mk ON bg.mata_kuliah_id = mk.id
		LEFT JOIN user u ON bs.aslab_id = u.id
		WHERE bs.aslab_id = ?
		ORDER BY bg.updated_at DESC
	`

	rows, err := db.Query(query, aslabID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return c.scanGroupsWithMembers(db, rows)
}

func (c *Client) getAllActiveGroupsDB(db *sql.DB) ([]BimbinganGroupItem, error) {
	query := `
		SELECT 
			bg.id, bg.mata_kuliah_id, COALESCE(mk.name, ''), COALESCE(mk.singkatan, ''),
			bg.kelas, bg.nama_kelompok, bg.judul, bg.deskripsi,
			COALESCE(bg.slot_id, ''), COALESCE(bg.flow_url, ''), COALESCE(bg.repo_url, ''),
			bg.status_konsul_0, COALESCE(bg.catatan_konsul_0, ''),
			bg.status_konsul_1, COALESCE(bg.catatan_konsul_1, ''),
			bg.status_konsul_2, COALESCE(bg.catatan_konsul_2, ''),
			bg.is_acc_final, bg.created_at, bg.updated_at,
			COALESCE(u.id, ''), COALESCE(u.name, ''), COALESCE(u.phone_number, '')
		FROM bimbingan_group bg
		INNER JOIN mata_kuliah mk ON bg.mata_kuliah_id = mk.id
		LEFT JOIN bimbingan_slot bs ON bg.slot_id = bs.id
		LEFT JOIN user u ON bs.aslab_id = u.id
		ORDER BY bg.updated_at DESC
		LIMIT 20
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return c.scanGroupsWithMembers(db, rows)
}

func (c *Client) getMyGroupsDB(db *sql.DB, userID, username, normPhone string) ([]BimbinganGroupItem, error) {
	query := `
		SELECT DISTINCT
			bg.id, bg.mata_kuliah_id, COALESCE(mk.name, ''), COALESCE(mk.singkatan, ''),
			bg.kelas, bg.nama_kelompok, bg.judul, bg.deskripsi,
			COALESCE(bg.slot_id, ''), COALESCE(bg.flow_url, ''), COALESCE(bg.repo_url, ''),
			bg.status_konsul_0, COALESCE(bg.catatan_konsul_0, ''),
			bg.status_konsul_1, COALESCE(bg.catatan_konsul_1, ''),
			bg.status_konsul_2, COALESCE(bg.catatan_konsul_2, ''),
			bg.is_acc_final, bg.created_at, bg.updated_at,
			COALESCE(u.id, ''), COALESCE(u.name, ''), COALESCE(u.phone_number, '')
		FROM bimbingan_group bg
		INNER JOIN mata_kuliah mk ON bg.mata_kuliah_id = mk.id
		INNER JOIN bimbingan_member bm ON bg.id = bm.group_id
		LEFT JOIN bimbingan_slot bs ON bg.slot_id = bs.id
		LEFT JOIN user u ON bs.aslab_id = u.id
		LEFT JOIN user mu ON bm.user_id = mu.id OR bm.nim = mu.username
		WHERE bm.user_id = ? OR bm.nim = ? OR mu.phone_number = ?
		ORDER BY bg.created_at DESC
	`

	rows, err := db.Query(query, userID, username, normPhone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return c.scanGroupsWithMembers(db, rows)
}

func (c *Client) searchGroupsDB(db *sql.DB, search string) ([]BimbinganGroupItem, error) {
	pattern := "%" + strings.ToLower(search) + "%"
	query := `
		SELECT DISTINCT
			bg.id, bg.mata_kuliah_id, COALESCE(mk.name, ''), COALESCE(mk.singkatan, ''),
			bg.kelas, bg.nama_kelompok, bg.judul, bg.deskripsi,
			COALESCE(bg.slot_id, ''), COALESCE(bg.flow_url, ''), COALESCE(bg.repo_url, ''),
			bg.status_konsul_0, COALESCE(bg.catatan_konsul_0, ''),
			bg.status_konsul_1, COALESCE(bg.catatan_konsul_1, ''),
			bg.status_konsul_2, COALESCE(bg.catatan_konsul_2, ''),
			bg.is_acc_final, bg.created_at, bg.updated_at,
			COALESCE(u.id, ''), COALESCE(u.name, ''), COALESCE(u.phone_number, '')
		FROM bimbingan_group bg
		INNER JOIN mata_kuliah mk ON bg.mata_kuliah_id = mk.id
		LEFT JOIN bimbingan_member bm ON bg.id = bm.group_id
		LEFT JOIN bimbingan_slot bs ON bg.slot_id = bs.id
		LEFT JOIN user u ON bs.aslab_id = u.id
		WHERE LOWER(bg.nama_kelompok) LIKE ?
		   OR LOWER(bg.judul) LIKE ?
		   OR LOWER(mk.name) LIKE ?
		   OR LOWER(bm.nim) LIKE ?
		   OR LOWER(bm.nama) LIKE ?
		ORDER BY bg.updated_at DESC
		LIMIT 15
	`

	rows, err := db.Query(query, pattern, pattern, pattern, pattern, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return c.scanGroupsWithMembers(db, rows)
}

func (c *Client) scanGroupsWithMembers(db *sql.DB, rows *sql.Rows) ([]BimbinganGroupItem, error) {
	var groups []BimbinganGroupItem
	var groupIDs []string

	for rows.Next() {
		var g BimbinganGroupItem
		var createdAtRaw, updatedAtRaw int64
		var isAccFinalInt int

		err := rows.Scan(
			&g.ID, &g.MataKuliahID, &g.MataKuliahName, &g.MataKuliahSingkatan,
			&g.Kelas, &g.NamaKelompok, &g.Judul, &g.Deskripsi,
			&g.SlotID, &g.FlowURL, &g.RepoURL,
			&g.StatusKonsul0, &g.CatatanKonsul0,
			&g.StatusKonsul1, &g.CatatanKonsul1,
			&g.StatusKonsul2, &g.CatatanKonsul2,
			&isAccFinalInt, &createdAtRaw, &updatedAtRaw,
			&g.AslabID, &g.AslabName, &g.AslabPhoneNumber,
		)
		if err != nil {
			return nil, err
		}

		g.IsAccFinal = isAccFinalInt == 1
		if createdAtRaw > 0 {
			g.CreatedAt = time.UnixMilli(createdAtRaw)
		}
		if updatedAtRaw > 0 {
			g.UpdatedAt = time.UnixMilli(updatedAtRaw)
		}

		if len(g.ID) >= 6 {
			g.ShortID = g.ID[:6]
		} else {
			g.ShortID = g.ID
		}

		if strings.HasPrefix(g.AslabPhoneNumber, "15") && len(g.AslabPhoneNumber) >= 14 {
			g.AslabPhoneNumber = ""
		}

		groups = append(groups, g)
		groupIDs = append(groupIDs, g.ID)
	}

	if len(groups) == 0 {
		return groups, nil
	}

	// Fetch members for all found groups
	membersMap, err := c.fetchMembersForGroups(db, groupIDs)
	if err == nil {
		for i := range groups {
			if members, ok := membersMap[groups[i].ID]; ok {
				groups[i].Members = members
			}
		}
	}

	return groups, nil
}

func (c *Client) fetchMembersForGroups(db *sql.DB, groupIDs []string) (map[string][]BimbinganMemberItem, error) {
	if len(groupIDs) == 0 {
		return map[string][]BimbinganMemberItem{}, nil
	}

	placeholders := make([]string, len(groupIDs))
	args := make([]interface{}, len(groupIDs))
	for i, id := range groupIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT 
			bm.id, bm.group_id, COALESCE(bm.user_id, ''), bm.nim, bm.nama, bm.role,
			COALESCE(u.phone_number, '')
		FROM bimbingan_member bm
		LEFT JOIN user u ON bm.user_id = u.id OR bm.nim = u.username
		WHERE bm.group_id IN (%s)
		ORDER BY bm.role DESC, bm.nim ASC
	`, strings.Join(placeholders, ","))

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string][]BimbinganMemberItem)
	for rows.Next() {
		var m BimbinganMemberItem
		if err := rows.Scan(&m.ID, &m.GroupID, &m.UserID, &m.NIM, &m.Nama, &m.Role, &m.PhoneNumber); err != nil {
			continue
		}
		if strings.HasPrefix(m.PhoneNumber, "15") && len(m.PhoneNumber) >= 14 {
			m.PhoneNumber = ""
		}
		res[m.GroupID] = append(res[m.GroupID], m)
	}

	return res, nil
}

// ResolveGroup resolves group by numeric index (relative to aslab mentees), short ID prefix, exact UUID, or group name
func (c *Client) ResolveGroup(identifier string, senderPhone string) (*BimbinganGroupItem, *UserInfo, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return nil, nil, fmt.Errorf("ID atau nomor kelompok wajib diisi")
	}

	caller, err := c.FindUserByPhone(senderPhone)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal memverifikasi nomor asisten: %w", err)
	}
	if caller == nil {
		return nil, nil, fmt.Errorf("nomor WhatsApp Anda belum terdaftar sebagai Asisten Lab / Pengurus di portal ASCII")
	}

	if caller.Role != "aslab" && caller.Role != "pengurus" && caller.Role != "koordinator" {
		return nil, nil, fmt.Errorf("hanya Asisten Lab, Koordinator, atau Pengurus yang dapat memberikan revisi/ACC bimbingan")
	}

	db, err := c.getDB()
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	// Check if identifier is an index number (e.g. "1", "2", "3")
	if idx, err := strconv.Atoi(identifier); err == nil && idx > 0 {
		mentees, err := c.getAslabMenteesDB(db, caller.ID)
		if err != nil {
			return nil, nil, err
		}
		if len(mentees) == 0 && (caller.Role == "pengurus" || caller.Role == "koordinator") {
			mentees, err = c.getAllActiveGroupsDB(db)
		}
		if idx <= len(mentees) {
			return &mentees[idx-1], caller, nil
		}
		return nil, nil, fmt.Errorf("nomor urut kelompok %d tidak ditemukan di daftar bimbingan Anda (total: %d)", idx, len(mentees))
	}

	// Try finding by exact ID or UUID prefix
	query := `
		SELECT 
			bg.id, bg.mata_kuliah_id, COALESCE(mk.name, ''), COALESCE(mk.singkatan, ''),
			bg.kelas, bg.nama_kelompok, bg.judul, bg.deskripsi,
			COALESCE(bg.slot_id, ''), COALESCE(bg.flow_url, ''), COALESCE(bg.repo_url, ''),
			bg.status_konsul_0, COALESCE(bg.catatan_konsul_0, ''),
			bg.status_konsul_1, COALESCE(bg.catatan_konsul_1, ''),
			bg.status_konsul_2, COALESCE(bg.catatan_konsul_2, ''),
			bg.is_acc_final, bg.created_at, bg.updated_at,
			COALESCE(u.id, ''), COALESCE(u.name, ''), COALESCE(u.phone_number, '')
		FROM bimbingan_group bg
		INNER JOIN mata_kuliah mk ON bg.mata_kuliah_id = mk.id
		LEFT JOIN bimbingan_slot bs ON bg.slot_id = bs.id
		LEFT JOIN user u ON bs.aslab_id = u.id
		WHERE bg.id = ? 
		   OR bg.id LIKE ? 
		   OR LOWER(bg.nama_kelompok) = LOWER(?)
		   OR LOWER(bg.nama_kelompok) LIKE ?
		LIMIT 1
	`
	prefixPattern := identifier + "%"
	likeNamePattern := "%" + strings.ToLower(identifier) + "%"

	rows, err := db.Query(query, identifier, prefixPattern, identifier, likeNamePattern)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	groups, err := c.scanGroupsWithMembers(db, rows)
	if err != nil {
		return nil, nil, err
	}
	if len(groups) == 0 {
		return nil, nil, fmt.Errorf("kelompok dengan ID/nama '%s' tidak ditemukan", identifier)
	}

	group := &groups[0]

	// Verify authorization: caller must be group's mentor or pengurus/koordinator
	if caller.Role != "pengurus" && caller.Role != "koordinator" && group.AslabID != caller.ID {
		return nil, nil, fmt.Errorf("Anda bukan asisten pembimbing untuk kelompok '%s' (Aslab pembimbing: Kak %s)", group.NamaKelompok, group.AslabName)
	}

	return group, caller, nil
}

// ReviewKonsul executes a consultation review (revisi / acc)
func (c *Client) ReviewKonsul(params ReviewKonsulParams) (*ReviewKonsulResult, error) {
	group, caller, err := c.ResolveGroup(params.GroupID, params.SenderPhone)
	if err != nil {
		return nil, err
	}

	if params.Stage < 0 || params.Stage > 2 {
		return nil, fmt.Errorf("tahap konsul harus 0 (Konsep), 1 (Flow), atau 2 (70%% Koding)")
	}

	status := strings.ToLower(params.Status)
	if status != "acc" && status != "revisi" && status != "pending" {
		return nil, fmt.Errorf("status harus 'acc', 'revisi', atau 'pending'")
	}

	// 1. Sync to Live Web API if session token / apiKey is present
	token, _ := c.getUserSession(params.SenderPhone)
	if token != "" || c.apiKey != "" {
		payload, _ := json.Marshal(map[string]interface{}{
			"groupId": group.ID,
			"stage":   params.Stage,
			"status":  status,
			"catatan": params.Catatan,
		})
		_, _ = c.doRequestWithAuth(http.MethodPost, "/bimbingan/konsul-review", bytes.NewReader(payload), token)
	}

	// 2. Update local database
	db, err := c.getDB()
	if err == nil {
		defer db.Close()
		now := time.Now().UnixMilli()

		var updateColStatus, updateColCatatan string
		switch params.Stage {
		case 0:
			updateColStatus = "status_konsul_0"
			updateColCatatan = "catatan_konsul_0"
		case 1:
			updateColStatus = "status_konsul_1"
			updateColCatatan = "catatan_konsul_1"
		case 2:
			updateColStatus = "status_konsul_2"
			updateColCatatan = "catatan_konsul_2"
		}

		updateQuery := fmt.Sprintf(`
			UPDATE bimbingan_group
			SET %s = ?, %s = ?, updated_at = ?
			WHERE id = ?
		`, updateColStatus, updateColCatatan)

		_, _ = db.Exec(updateQuery, status, params.Catatan, now, group.ID)
	}

	var stageName string
	switch params.Stage {
	case 0:
		stageName = "Konsul 0 (Konsep & Judul)"
		group.StatusKonsul0 = status
		group.CatatanKonsul0 = params.Catatan
	case 1:
		stageName = "Konsul 1 (Flow Program & Desain)"
		group.StatusKonsul1 = status
		group.CatatanKonsul1 = params.Catatan
	case 2:
		stageName = "Konsul 2 (70% Koding & Implementasi)"
		group.StatusKonsul2 = status
		group.CatatanKonsul2 = params.Catatan
	}

	statusLabel := "⚠️ PERLU REVISI"
	if status == "acc" {
		statusLabel = "✅ TELAH DI-ACC"
	}

	// Collect member recipients
	recipients := make([]BimbinganMemberItem, 0)
	for _, m := range group.Members {
		if norm := NormalizePhoneNumber(m.PhoneNumber); norm != "" {
			m.PhoneNumber = norm
			recipients = append(recipients, m)
		}
	}

	return &ReviewKonsulResult{
		Success:     true,
		Message:     fmt.Sprintf("Status %s berhasil diubah menjadi %s", stageName, statusLabel),
		StageName:   stageName,
		StatusLabel: statusLabel,
		Group:       group,
		AslabName:   caller.Name,
		Recipients:  recipients,
	}, nil
}

// ToggleAccFinal sets the final ACC status for a bimbingan group
func (c *Client) ToggleAccFinal(params AccFinalParams) (*ReviewKonsulResult, error) {
	group, caller, err := c.ResolveGroup(params.GroupID, params.SenderPhone)
	if err != nil {
		return nil, err
	}

	// 1. Sync to Live Web API
	token, _ := c.getUserSession(params.SenderPhone)
	if token != "" || c.apiKey != "" {
		payload, _ := json.Marshal(map[string]interface{}{
			"groupId":    group.ID,
			"isAccFinal": params.IsAccFinal,
		})
		_, _ = c.doRequestWithAuth(http.MethodPost, "/bimbingan/acc-final", bytes.NewReader(payload), token)
	}

	// 2. Update local DB
	db, err := c.getDB()
	if err == nil {
		defer db.Close()
		now := time.Now().UnixMilli()

		isAccInt := 0
		if params.IsAccFinal {
			isAccInt = 1
		}

		updateQuery := `
			UPDATE bimbingan_group
			SET is_acc_final = ?, status_konsul_0 = 'acc', status_konsul_1 = 'acc', status_konsul_2 = 'acc', updated_at = ?
			WHERE id = ?
		`
		_, _ = db.Exec(updateQuery, isAccInt, now, group.ID)
	}

	group.IsAccFinal = params.IsAccFinal
	group.StatusKonsul0 = "acc"
	group.StatusKonsul1 = "acc"
	group.StatusKonsul2 = "acc"

	recipients := make([]BimbinganMemberItem, 0)
	for _, m := range group.Members {
		if norm := NormalizePhoneNumber(m.PhoneNumber); norm != "" {
			m.PhoneNumber = norm
			recipients = append(recipients, m)
		}
	}

	return &ReviewKonsulResult{
		Success:     true,
		Message:     "Kelompok bimbingan berhasil dinyatakan ACC Final (Siap Demo)!",
		StageName:   "ACC Final (Siap Demo)",
		StatusLabel: "🎯 ACC FINAL",
		Group:       group,
		AslabName:   caller.Name,
		Recipients:  recipients,
	}, nil
}

// LoginUser authenticates via NIM and Password and permanently links the sender's WhatsApp phone number
func (c *Client) LoginUser(senderPhone, nim, password string) (*UserInfo, error) {
	norm := NormalizePhoneNumber(senderPhone)
	if norm == "" {
		return nil, fmt.Errorf("nomor WhatsApp pengirim tidak valid")
	}

	nim = strings.TrimSpace(nim)
	password = strings.TrimSpace(password)
	if nim == "" || password == "" {
		return nil, fmt.Errorf("NIM dan password wajib diisi")
	}

	// 1. Attempt login via Web API
	loginPayload, _ := json.Marshal(map[string]interface{}{
		"nim":        nim,
		"password":   password,
		"rememberMe": true,
	})

	respBytes, err := c.doRequest(http.MethodPost, "/login", bytes.NewReader(loginPayload))
	var authUser struct {
		Session struct {
			Token     string `json:"token"`
			ExpiresAt int64  `json:"expiresAt"`
			ID        string `json:"id"`
		} `json:"session"`
		Token string `json:"token"`
		User  struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Email       string `json:"email"`
			Username    string `json:"username"`
			Role        string `json:"role"`
			PhoneNumber string `json:"phoneNumber,omitempty"`
		} `json:"user"`
		Error   string `json:"error,omitempty"`
		Message string `json:"message,omitempty"`
	}

	if err == nil {
		_ = json.Unmarshal(respBytes, &authUser)
	}

	// 2. If API request succeeded, save phone number and session to database
	var matchedUser UserInfo
	if authUser.User.ID != "" {
		realPhone := authUser.User.PhoneNumber
		if realPhone == "" && (strings.HasPrefix(norm, "628") || strings.HasPrefix(norm, "08")) {
			realPhone = norm
		}
		matchedUser = UserInfo{
			ID:          authUser.User.ID,
			Name:        authUser.User.Name,
			Email:       authUser.User.Email,
			Username:    authUser.User.Username,
			Role:        authUser.User.Role,
			PhoneNumber: realPhone,
		}
	} else {
		// Fallback: If web server is unreachable or returned error, check error message
		if err != nil && (strings.Contains(err.Error(), "400") || strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "invalid_credentials")) {
			return nil, fmt.Errorf("NIM atau password salah. Pastikan kredensial Anda sesuai dengan akun portal ASCII")
		}
		if err != nil {
			return nil, fmt.Errorf("gagal menghubungi server autentikasi: %w", err)
		}
		if authUser.Error != "" || authUser.Message != "" {
			msg := authUser.Message
			if msg == "" {
				msg = authUser.Error
			}
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("autentikasi gagal. Periksa kembali NIM dan password Anda")
	}

	// 3. Link session in local database with token
	sessionToken := authUser.Session.Token
	if sessionToken == "" {
		sessionToken = authUser.Token
	}
	if sessionToken == "" {
		sessionToken = authUser.Session.ID
	}

	db, dbErr := c.getDB()
	if dbErr == nil {
		defer db.Close()
		now := time.Now().UnixMilli()
		raw := strings.TrimSpace(senderPhone)

		// 3. Upsert user into user table so it's always accessible
		_, _ = db.Exec(`
			INSERT INTO user (id, name, email, username, role, phone_number, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET 
				name = excluded.name,
				email = excluded.email,
				username = excluded.username,
				role = excluded.role,
				phone_number = CASE 
					WHEN excluded.phone_number != '' AND excluded.phone_number NOT LIKE '15%' THEN excluded.phone_number 
					ELSE user.phone_number 
				END,
				updated_at = excluded.updated_at
		`, matchedUser.ID, matchedUser.Name, matchedUser.Email, matchedUser.Username, matchedUser.Role, matchedUser.PhoneNumber, now, now)

		_, _ = db.Exec(`
			INSERT INTO bot_wa_sessions (sender_id, user_id, token, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(sender_id) DO UPDATE SET user_id = excluded.user_id, token = excluded.token, updated_at = excluded.updated_at
		`, raw, matchedUser.ID, sessionToken, now, now)

		if norm != "" && norm != raw {
			_, _ = db.Exec(`
				INSERT INTO bot_wa_sessions (sender_id, user_id, token, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?)
				ON CONFLICT(sender_id) DO UPDATE SET user_id = excluded.user_id, token = excluded.token, updated_at = excluded.updated_at
			`, norm, matchedUser.ID, sessionToken, now, now)
		}
	}

	return &matchedUser, nil
}

// LogoutUser removes the WhatsApp link for the sender's phone number
func (c *Client) LogoutUser(senderPhone string) (*UserInfo, error) {
	norm := NormalizePhoneNumber(senderPhone)
	raw := strings.TrimSpace(senderPhone)

	caller, err := c.FindUserByPhone(senderPhone)
	if err != nil {
		return nil, err
	}
	if caller == nil {
		return nil, fmt.Errorf("nomor WhatsApp Anda belum terhubung ke akun manapun")
	}

	db, err := c.getDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	_, _ = db.Exec("DELETE FROM bot_wa_sessions WHERE sender_id = ? OR sender_id = ?", raw, norm)
	return caller, nil
}

// GetProfile retrieves the profile linked to the sender's WhatsApp phone number
func (c *Client) GetProfile(senderPhone string) (*UserInfo, error) {
	return c.FindUserByPhone(senderPhone)
}
