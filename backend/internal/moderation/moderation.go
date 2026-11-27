// Package moderation is Maplas' simple, rule-based "AI" moderator.
//
// Every new place gets a trust score that starts at 100. Each suspicious
// signal lowers the score; a trusted user raises it. Then:
//
//	score >= 70  -> approved automatically (shown on the map)
//	score >= 40  -> pending (an admin takes a look)
//	score <  40  -> rejected (not saved; the user is told why)
//
// Some problems (swear words, a place that already exists, an invalid
// category or location) reject immediately, whatever the score.
// Comments and usernames are checked with the same rules: they are either
// accepted or rejected.
package moderation

import (
	"math"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"backend/internal/db"
)

type Decision string

const (
	Approve Decision = "approved"
	Review  Decision = "pending"
	Reject  Decision = "rejected"
)

type Result struct {
	Decision Decision `json:"decision"`
	Score    int      `json:"score"`
	Reasons  []string `json:"reasons"`
}

const (
	approveScore = 70
	reviewScore  = 40
)

var AllowedCategories = map[string]bool{
	"Tarihi": true, "Doğa": true, "Plaj": true, "Müze": true, "Antik Kent": true,
	"Alışveriş": true, "Manzara": true, "Eğlence": true, "Diğer": true,
}

// Whole words that are never allowed.
var bannedWords = map[string]bool{
	"amk": true, "aq": true, "sik": true, "piç": true, "göt": true, "ibne": true,
	"bok": true, "fuck": true, "shit": true, "bitch": true,
}

// Word beginnings that are never allowed (also checked without Turkish letters).
var bannedPrefixes = []string{
	"orospu", "siktir", "sikerim", "sikeyim", "sikik", "yarrak", "pezevenk",
	"amına", "amina", "gavat", "kahpe", "fucking", "motherfuck",
}

var reservedUsernames = map[string]bool{
	"admin": true, "administrator": true, "root": true, "maplas": true,
	"moderator": true, "support": true, "destek": true, "yonetici": true, "yönetici": true,
}

var (
	linkRe     = regexp.MustCompile(`(?i)(https?://|www\.|\b[a-z0-9-]+\.(com|net|org|xyz|info|ru|tk|io|co|biz|top|site|online)\b)`)
	phoneRe    = regexp.MustCompile(`\+?\d[\d\s().-]{8,}\d`) // candidate; see hasPhone
	usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_çğıöşüÇĞİÖŞÜ]{3,20}$`)
)

// Database lookups; tests replace these.
var (
	lookupDuplicate = isDuplicate
	lookupHistory   = userHistory
)

// PlaceInput is what a user submits when adding a place.
type PlaceInput struct {
	Name        string
	Description string
	Category    string
	City        string
	Lat, Lng    float64
	ImageURL    string
	UserID      int // 0 = not logged in
}

// CheckPlace decides whether a new place is approved, sent to an admin, or rejected.
func CheckPlace(in PlaceInput) Result {
	name := strings.TrimSpace(in.Name)
	desc := strings.TrimSpace(in.Description)

	// --- Hard rules: reject immediately ---
	if n := utf8.RuneCountInString(name); n < 3 || n > 100 {
		return reject("Yer adı 3 ile 100 karakter arasında olmalı.")
	}
	if n := utf8.RuneCountInString(desc); n < 10 || n > 2000 {
		return reject("Açıklama 10 ile 2000 karakter arasında olmalı.")
	}
	if !AllowedCategories[in.Category] {
		return reject("Geçersiz kategori.")
	}
	if in.Lat < -90 || in.Lat > 90 || in.Lng < -180 || in.Lng > 180 || (in.Lat == 0 && in.Lng == 0) {
		return reject("Haritadan geçerli bir konum seçin.")
	}
	if containsBanned(name) || containsBanned(desc) || containsBanned(in.City) {
		return reject("Uygunsuz ifade içeriyor.")
	}
	if lookupDuplicate(name, in.Lat, in.Lng) {
		return reject("Bu yer zaten haritada var.")
	}

	// --- Soft rules: lower the score ---
	r := Result{Score: 100}
	text := name + " " + desc

	if linkRe.MatchString(text) {
		r.penalize(50, "Bağlantı (link) içeriyor.")
	}
	if hasPhone(text) {
		r.penalize(40, "Telefon numarası içeriyor.")
	}
	if hasRepeatedRun(text, 4) {
		r.penalize(25, "Anlamsız tekrar eden harfler var (ör. 'aaaa').")
	}
	if isShouting(name) || isShouting(desc) {
		r.penalize(15, "Tamamı büyük harfle yazılmış.")
	}
	if symbolRatio(text) > 0.3 {
		r.penalize(20, "Çok fazla sembol veya emoji var.")
	}
	if len(strings.Fields(desc)) < 4 {
		r.penalize(15, "Açıklama çok kısa.")
	}
	if strings.EqualFold(name, desc) {
		r.penalize(20, "Açıklama, yer adıyla aynı.")
	}
	if strings.TrimSpace(in.ImageURL) == "" {
		r.penalize(5, "Fotoğraf eklenmemiş.")
	}

	// --- Who is adding it? ---
	forceReview := false
	if in.UserID == 0 {
		r.penalize(25, "Giriş yapılmadan eklendi.")
	} else {
		approved, recent := lookupHistory(in.UserID)
		if approved >= 3 {
			r.Score += 15 // trusted contributor
		}
		if recent >= 5 {
			forceReview = true
			r.Reasons = append(r.Reasons, "Son bir saatte çok fazla yer eklendi.")
		}
	}

	switch {
	case r.Score < reviewScore:
		r.Decision = Reject
	case r.Score < approveScore || forceReview:
		r.Decision = Review
	default:
		r.Decision = Approve
	}
	return r
}

// CheckComment accepts or rejects a comment.
func CheckComment(content string, rating int) Result {
	content = strings.TrimSpace(content)
	if rating < 1 || rating > 5 {
		return reject("Puan 1 ile 5 arasında olmalı.")
	}
	if n := utf8.RuneCountInString(content); n < 2 || n > 1000 {
		return reject("Yorum 2 ile 1000 karakter arasında olmalı.")
	}
	if containsBanned(content) {
		return reject("Uygunsuz ifade içeriyor.")
	}
	if linkRe.MatchString(content) {
		return reject("Yorumlarda bağlantı (link) paylaşılamaz.")
	}
	if hasPhone(content) {
		return reject("Yorumlarda telefon numarası paylaşılamaz.")
	}
	if hasRepeatedRun(content, 8) {
		return reject("Yorum anlamsız tekrarlar içeriyor.")
	}
	return Result{Decision: Approve, Score: 100}
}

// CheckUsername accepts or rejects a new username.
// Reserved names (like "admin") are only allowed when allowReserved is true.
func CheckUsername(username string, allowReserved bool) Result {
	if !usernameRe.MatchString(username) {
		return reject("Kullanıcı adı 3-20 karakter olmalı; sadece harf, rakam ve _ kullanılabilir.")
	}
	if containsBanned(username) {
		return reject("Kullanıcı adı uygunsuz ifade içeriyor.")
	}
	if !allowReserved && reservedUsernames[lowerTR(username)] {
		return reject("Bu kullanıcı adı kullanılamaz.")
	}
	return Result{Decision: Approve, Score: 100}
}

// ---------- helpers ----------

func reject(reason string) Result {
	return Result{Decision: Reject, Score: 0, Reasons: []string{reason}}
}

func (r *Result) penalize(points int, reason string) {
	r.Score -= points
	r.Reasons = append(r.Reasons, reason)
}

func lowerTR(s string) string {
	return strings.ToLowerSpecial(unicode.TurkishCase, s)
}

// foldTR turns Turkish letters into plain Latin ones (ş->s, ı->i, ...).
var trFolder = strings.NewReplacer("ç", "c", "ğ", "g", "ı", "i", "ö", "o", "ş", "s", "ü", "u")

func words(s string) []string {
	return strings.FieldsFunc(lowerTR(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func containsBanned(s string) bool {
	for _, w := range words(s) {
		if bannedWords[w] {
			return true
		}
		folded := trFolder.Replace(w)
		for _, p := range bannedPrefixes {
			if strings.HasPrefix(w, p) || strings.HasPrefix(folded, p) {
				return true
			}
		}
	}
	return false
}

// hasPhone looks for 10-13 digits written together (spaces, dashes, dots and
// brackets allowed), so date ranges like "1453 - 1923" don't count.
func hasPhone(s string) bool {
	for _, m := range phoneRe.FindAllString(s, -1) {
		digits := 0
		for _, r := range m {
			if unicode.IsDigit(r) {
				digits++
			}
		}
		if digits >= 10 && digits <= 13 {
			return true
		}
	}
	return false
}

// hasRepeatedRun reports whether any letter repeats n or more times in a row.
func hasRepeatedRun(s string, n int) bool {
	var prev rune
	count := 0
	for _, r := range lowerTR(s) {
		if r == prev && unicode.IsLetter(r) {
			count++
			if count >= n {
				return true
			}
		} else {
			prev, count = r, 1
		}
	}
	return false
}

func isShouting(s string) bool {
	letters, upper := 0, 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			letters++
			if unicode.IsUpper(r) {
				upper++
			}
		}
	}
	return letters >= 6 && float64(upper)/float64(letters) > 0.8
}

func symbolRatio(s string) float64 {
	total, symbols := 0, 0
	for _, r := range s {
		if unicode.IsSpace(r) {
			continue
		}
		total++
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !strings.ContainsRune(".,!?;:'\"()-–", r) {
			symbols++
		}
	}
	if total == 0 {
		return 0
	}
	return float64(symbols) / float64(total)
}

// isDuplicate looks for an approved or pending place with the same name within ~300 m.
func isDuplicate(name string, lat, lng float64) bool {
	const deg = 0.003 // about 330 m of latitude
	lngDeg := deg / math.Max(math.Cos(lat*math.Pi/180), 0.1)
	rows, err := db.DB.Query(`
		SELECT COALESCE(name->>'tr', '') FROM places
		WHERE status IN ('approved', 'pending')
		  AND lat BETWEEN $1 AND $2 AND lng BETWEEN $3 AND $4`,
		lat-deg, lat+deg, lng-lngDeg, lng+lngDeg)
	if err != nil {
		return false
	}
	defer rows.Close()
	target := normalizeName(name)
	for rows.Next() {
		var existing string
		if rows.Scan(&existing) == nil && normalizeName(existing) == target {
			return true
		}
	}
	return false
}

func normalizeName(s string) string {
	return trFolder.Replace(strings.Join(words(s), " "))
}

// userHistory returns how many of the user's places are approved, and how many
// places they added in the last hour.
func userHistory(userID int) (approved, recent int) {
	db.DB.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE status = 'approved'),
			COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '1 hour')
		FROM places WHERE creator_id = $1`, userID).Scan(&approved, &recent)
	return
}
