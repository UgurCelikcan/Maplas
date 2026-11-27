package moderation

import "testing"

// fakeDB replaces the database lookups for the duration of a test.
func fakeDB(t *testing.T, duplicate bool, approved, recent int) {
	t.Helper()
	origDup, origHist := lookupDuplicate, lookupHistory
	lookupDuplicate = func(string, float64, float64) bool { return duplicate }
	lookupHistory = func(int) (int, int) { return approved, recent }
	t.Cleanup(func() { lookupDuplicate, lookupHistory = origDup, origHist })
}

func goodPlace() PlaceInput {
	return PlaceInput{
		Name:        "Galata Kulesi",
		Description: "İstanbul'un simgelerinden biri, 1348 yılında Cenevizliler tarafından yapılmış taş kule.",
		Category:    "Tarihi",
		City:        "İstanbul",
		Lat:         41.0256, Lng: 28.9741,
		ImageURL: "/uploads/abc.jpg",
		UserID:   7,
	}
}

func TestCheckPlace(t *testing.T) {
	tests := []struct {
		name     string
		edit     func(p *PlaceInput)
		dup      bool
		approved int
		recent   int
		want     Decision
	}{
		{"normal place from a user", func(p *PlaceInput) {}, false, 0, 0, Approve},
		{"anonymous with photo", func(p *PlaceInput) { p.UserID = 0 }, false, 0, 0, Approve},
		{"anonymous, no photo, short description", func(p *PlaceInput) {
			p.UserID, p.ImageURL, p.Description = 0, "", "Çok güzel yer"
		}, false, 0, 0, Review},
		{"date range is not a phone number", func(p *PlaceInput) {
			p.Description = "Osmanlı döneminde 1453 - 1923 yılları arasında yangın gözetleme kulesi olarak kullanıldı."
		}, false, 0, 0, Approve},
		{"phone number", func(p *PlaceInput) { p.Description += " Rezervasyon: 0532 123 45 67" }, false, 0, 0, Review},
		{"link plus phone is spam", func(p *PlaceInput) {
			p.Description = "Ucuz tur için www.ucuztur.xyz adresine girin veya 0532 123 45 67 arayın hemen"
		}, false, 0, 0, Reject},
		{"swear word", func(p *PlaceInput) { p.Name = "Orospu çocuğu mekanı" }, false, 0, 0, Reject},
		{"swear word without Turkish letters", func(p *PlaceInput) { p.Description += " amina koyim" }, false, 0, 0, Reject},
		{"duplicate", func(p *PlaceInput) {}, true, 0, 0, Reject},
		{"bad category", func(p *PlaceInput) { p.Category = "Kumarhane" }, false, 0, 0, Reject},
		{"no location", func(p *PlaceInput) { p.Lat, p.Lng = 0, 0 }, false, 0, 0, Reject},
		{"name too short", func(p *PlaceInput) { p.Name = "ab" }, false, 0, 0, Reject},
		{"too many posts in an hour", func(p *PlaceInput) {}, false, 0, 6, Review},
		{"trusted user offsets a missing photo and caps", func(p *PlaceInput) {
			p.ImageURL, p.Name = "", "GALATA KULESİ"
		}, false, 5, 0, Approve},
		{"repeated letters and shouting", func(p *PlaceInput) {
			p.UserID, p.Name, p.Description = 0, "SÜPERRRRR YERRRR", "ÇOK GÜZELLLLL GELİN GÖRÜN"
		}, false, 0, 0, Reject},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeDB(t, tt.dup, tt.approved, tt.recent)
			p := goodPlace()
			tt.edit(&p)
			got := CheckPlace(p)
			if got.Decision != tt.want {
				t.Errorf("got %s (score %d, reasons %v), want %s", got.Decision, got.Score, got.Reasons, tt.want)
			}
		})
	}
}

func TestCheckComment(t *testing.T) {
	tests := []struct {
		content string
		rating  int
		want    Decision
	}{
		{"Harika bir yer, mutlaka görün!", 5, Approve},
		{"İyi", 3, Approve},
		{"", 5, Reject},
		{"Güzel", 0, Reject},
		{"Güzel", 6, Reject},
		{"Siktir git", 1, Reject},
		{"Detaylar için https://spam.com", 5, Reject},
		{"Beni arayın 0532 123 45 67", 5, Reject},
		{"Süperrrrrrrrrrrr", 5, Reject},
	}
	for _, tt := range tests {
		if got := CheckComment(tt.content, tt.rating); got.Decision != tt.want {
			t.Errorf("CheckComment(%q, %d) = %s %v, want %s", tt.content, tt.rating, got.Decision, got.Reasons, tt.want)
		}
	}
}

func TestCheckUsername(t *testing.T) {
	tests := []struct {
		username      string
		allowReserved bool
		want          Decision
	}{
		{"ugur_42", false, Approve},
		{"Çağrı", false, Approve},
		{"ab", false, Reject},
		{"bir isim", false, Reject},
		{"cok_uzun_bir_kullanici_adi", false, Reject},
		{"admin", false, Reject},
		{"Admin", false, Reject},
		{"admin", true, Approve},
		{"orospu1", false, Reject},
	}
	for _, tt := range tests {
		if got := CheckUsername(tt.username, tt.allowReserved); got.Decision != tt.want {
			t.Errorf("CheckUsername(%q, %v) = %s %v, want %s", tt.username, tt.allowReserved, got.Decision, got.Reasons, tt.want)
		}
	}
}
