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

	// Check if file exists or resolve relative to working dir / executable
	if _, err := os.Stat(dbPath); err != nil {
		// Try alternate paths
		candidates := []string{
			"../ascii-if/local.db",
			"ascii-if/local.db",
			"./local.db",
			filepath.Join(os.Getenv("USERPROFILE"), "Documents", "Programming", "Web Ascii", "ascii-if", "local.db"),
		}
		found := false
		for _, cand := range candidates {
			if _, err := os.Stat(cand); err == nil {
				dbPath = cand
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("database file not found at %s: %w", dbPath, err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	return db, nil
}

// FindUserByPhone finds a user by normalized phone number
func (c *Client) FindUserByPhone(phone string) (*UserInfo, error) {
	norm := NormalizePhoneNumber(phone)
	if norm == "" {
		return nil, fmt.Errorf("nomor telepon kosong atau tidak valid")
	}

	db, err := c.getDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Candidate phone formats in DB: "628...", "08...", "+628..."
	var localFormat string
	if strings.HasPrefix(norm, "62") {
		localFormat = "0" + norm[2:]
	}

	query := `
		SELECT id, name, email, COALESCE(username, ''), role, COALESCE(phone_number, '')
		FROM user
		WHERE phone_number = ? OR phone_number = ? OR phone_number = ?
		LIMIT 1
	`
	plusFormat := "+" + norm

	var u UserInfo
	err = db.QueryRow(query, norm, localFormat, plusFormat).Scan(
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

// GetBimbinganSummary retrieves bimbingan summary for the sender or search query
func (c *Client) GetBimbinganSummary(phone, query string) (*BimbinganSummaryResult, error) {
	query = strings.TrimSpace(query)

	db, err := c.getDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	normPhone := NormalizePhoneNumber(phone)
	var caller *UserInfo
	if normPhone != "" {
		caller, _ = c.FindUserByPhone(normPhone)
	}

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

	db, err := c.getDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	now := time.Now().UnixMilli()

	var updateColStatus, updateColCatatan string
	var stageName string

	switch params.Stage {
	case 0:
		updateColStatus = "status_konsul_0"
		updateColCatatan = "catatan_konsul_0"
		stageName = "Konsul 0 (Konsep & Judul)"
	case 1:
		updateColStatus = "status_konsul_1"
		updateColCatatan = "catatan_konsul_1"
		stageName = "Konsul 1 (Flow Program & Desain)"
	case 2:
		updateColStatus = "status_konsul_2"
		updateColCatatan = "catatan_konsul_2"
		stageName = "Konsul 2 (70% Koding & Implementasi)"
	}

	updateQuery := fmt.Sprintf(`
		UPDATE bimbingan_group
		SET %s = ?, %s = ?, updated_at = ?
		WHERE id = ?
	`, updateColStatus, updateColCatatan)

	_, err = db.Exec(updateQuery, status, params.Catatan, now, group.ID)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui status di database: %w", err)
	}

	// Update in-memory group state
	switch params.Stage {
	case 0:
		group.StatusKonsul0 = status
		group.CatatanKonsul0 = params.Catatan
	case 1:
		group.StatusKonsul1 = status
		group.CatatanKonsul1 = params.Catatan
	case 2:
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

	db, err := c.getDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	now := time.Now().UnixMilli()

	isAccInt := 0
	if params.IsAccFinal {
		isAccInt = 1
	}

	// If ACC Final is true, automatically set all stages to 'acc'
	updateQuery := `
		UPDATE bimbingan_group
		SET is_acc_final = ?, status_konsul_0 = 'acc', status_konsul_1 = 'acc', status_konsul_2 = 'acc', updated_at = ?
		WHERE id = ?
	`
	_, err = db.Exec(updateQuery, isAccInt, now, group.ID)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui ACC Final di database: %w", err)
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
		User struct {
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

	// 2. If API request succeeded, save phone number to database
	var matchedUser UserInfo
	if authUser.User.ID != "" {
		matchedUser = UserInfo{
			ID:          authUser.User.ID,
			Name:        authUser.User.Name,
			Email:       authUser.User.Email,
			Username:    authUser.User.Username,
			Role:        authUser.User.Role,
			PhoneNumber: norm,
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

	// 3. Link phone number in local database if available
	db, dbErr := c.getDB()
	if dbErr == nil {
		defer db.Close()
		_, _ = db.Exec("UPDATE user SET phone_number = ?, updated_at = ? WHERE id = ?", norm, time.Now().UnixMilli(), matchedUser.ID)
	}

	return &matchedUser, nil
}

// LogoutUser removes the WhatsApp link for the sender's phone number
func (c *Client) LogoutUser(senderPhone string) (*UserInfo, error) {
	norm := NormalizePhoneNumber(senderPhone)
	if norm == "" {
		return nil, fmt.Errorf("nomor WhatsApp pengirim tidak valid")
	}

	caller, err := c.FindUserByPhone(norm)
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

	_, err = db.Exec("UPDATE user SET phone_number = NULL, updated_at = ? WHERE id = ?", time.Now().UnixMilli(), caller.ID)
	if err != nil {
		return nil, fmt.Errorf("gagal melepas tautan akun di database: %w", err)
	}

	return caller, nil
}

// GetProfile retrieves the profile linked to the sender's WhatsApp phone number
func (c *Client) GetProfile(senderPhone string) (*UserInfo, error) {
	norm := NormalizePhoneNumber(senderPhone)
	if norm == "" {
		return nil, fmt.Errorf("nomor WhatsApp pengirim tidak valid")
	}
	return c.FindUserByPhone(norm)
}
