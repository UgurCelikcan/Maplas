# Developer Roadmap - Maplas Project

Bu belge, projenin mevcut durumuna ve standart yazılım geliştirme pratiklerine dayanarak oluşturulmuş bir yol haritasıdır.

## 🚀 Faz 1: Mimari ve Kod Kalitesi (Refactoring)

Mevcut prototipin daha ölçeklenebilir ve yönetilebilir hale getirilmesi.

- [x] **Backend Refactoring**
  - [x] `backend/main.go` dosyasını modüler hale getir:
    - `internal/handlers`: HTTP isteklerini karşılayan fonksiyonlar.
    - `internal/models`: Veritabanı yapıları (structs).
    - `internal/db`: Veritabanı bağlantısı ve sorgu fonksiyonları.
    - `internal/middleware`: Auth (JWT), CORS, vb. ara katmanlar.
  - [x] Hardcoded secret'ları (JWT Key, Admin Code) `.env` dosyasına taşı.

- [ ] **Frontend State Management**
  - [ ] **Pinia** kütüphanesini projeye dahil et.
  - [ ] `App.vue` içindeki karmaşık state mantığını (User auth, Places list) Pinia store'larına taşı.
  - [ ] API çağrılarını store action'ları üzerinden yönet.

- [ ] **Tip Güvenliği (Type Safety)**
  - [ ] Backend ve Frontend arasında paylaşılan tipleri (Place, User) senkronize tutacak bir yapı kur (veya manuel olarak sıkı takip et).

## ✨ Faz 2: Kullanıcı Deneyimi ve Yeni Özellikler

Uygulamanın yeteneklerini artırma.

- [ ] **UI/UX İyileştirmeleri**
  - [ ] `alert()` kullanımı yerine modern "Toast" bildirimleri (örn: `vue-toastification`).
  - [ ] Yükleniyor (Loading) durumları için Skeleton loading ekranları.
  - [ ] Form validasyonlarını (VeeValidate veya manuel) görsel olarak iyileştir.

- [ ] **Harita Özellikleri**
  - [ ] Kategoriye özel harita pinleri (Restoran için çatal-bıçak ikonu, Park için ağaç vb.).
  - [ ] "Beni Bul" butonu ile kullanıcı konumuna hızlı odaklanma ve canlı takip.
  - [ ] Harita üzerinde "Cluster" (kümeleme) ayarlarını optimize et.

- [ ] **Kullanıcı Etkileşimi**
  - [ ] Favori Yerler / Kaydedilenler listesi.
  - [ ] Profil sayfası düzenleme (Avatar yükleme, Biyo güncelleme - Backend hazır, Frontend entegrasyonu kontrol edilmeli).
  - [ ] Yorumlara yanıt verme (Admin veya Mekan sahibi için).

## 🛡️ Faz 3: Test ve Güvenlik

Uygulamanın kararlılığını sağlama.

- [ ] **Backend Testleri**
  - [ ] Kritik API endpointleri için Unit Testler (`go test`).
  - [ ] Veritabanı işlemleri için Mock testler.

- [ ] **Frontend Testleri**
  - [ ] **Vitest** kurulumu.
  - [ ] Kritik bileşenlerin (MapDisplay, AuthModal) unit testleri.

- [ ] **Güvenlik**
  - [x] Rate Limiting ekle (API spam koruması).
  - [x] Resim yüklemeleri için boyut ve tür kontrolünü sıkılaştır.

## 🚢 Faz 4: DevOps ve Dağıtım (Deployment)

Yayına alma süreçleri.

- [ ] **Dockerizasyon**
  - [ ] Backend için multi-stage `Dockerfile`.
  - [ ] Frontend için Nginx tabanlı production `Dockerfile`.
  - [ ] Tüm yapıyı tek komutla kaldırmak için güncel `docker-compose.yml`.

- [ ] **CI/CD**
  - [ ] GitHub Actions veya GitLab CI ile otomatik build ve lint kontrolü.

---
*Son Güncelleme: 12 Ocak 2026*
