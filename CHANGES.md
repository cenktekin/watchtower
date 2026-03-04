# Watchtower Türkçe Çeviri Projesi

Bu fork, Watchtower uygulamasının **tam Türkçe çevirisini** ve **önemli özellik iyileştirmelerini** içerir.

## 🆕 Orijinal Repodan Farklar

### ✨ Yeni Özellikler

#### 1. OpenRouter Desteği
- **Orijinal repo'da YOK**
- OpenRouter API entegrasyonu eklendi
- Ücretsiz LLM modelleri kullanma imkanı
- Endpoint: `https://openrouter.ai/api/v1/chat/completions`
- Varsayılan model: `meta-llama/llama-3-8b-instruct:free`

#### 2. Ollama Yerel LLM Desteği
- **Orijinal repo'da YOK**
- Ollama yerel model desteği eklendi
- Tamamen offline, gizlilik odaklı kullanım
- Endpoint: `http://localhost:11434/v1/chat/completions`
- Varsayılan model: `llama3`

#### 3. API Key Zorunluluğu Kaldırıldı
- **Orijinal:** Tüm sağlayıcılar için API key zorunlu
- **Sizde:** `ollama` ve `local` sağlayıcıları için API key **gerekmez**
- Kurulum sihirbazı bu sağlayıcıları seçtiğinizde API key adımı otomatik atlanır
- Auth header kontrolü eklendi (`if cfg.AuthHeader() != ""`)

#### 4. HTTP Timeout Artırıldı
- **Orijinal:** 30 saniye
- **Sizde:** 120 saniye
- Uzun AI yanıtları için daha güvenilir

### 🔧 İyileştirmeler

#### Kurulum Sihirbazı
- Sağlayıcı tipine göre akıllı akış
- API key gerektirmeyen sağlayıcılar için otomatik atlama
- Kullanıcı dostu etiketler:
  - `(anahtar gerekmez)` - ollama, local
  - `(ücretsiz katman mevcut)` - groq, openrouter

#### AI Prompt'ları
- Global brief prompt'u tamamen yenilendi:
  - Türkçe yanıt zorunluluğu
  - Risk çapaları (100/75/50/25/0)
  - Somut olay odaklı ("bomba, grev, darbe" vb.)
  - Diplomatik dil yasak
  - Aksiyon odaklı ülke riskleri

### 🌍 Türkçe Çeviri

### Eklenen Dosyalar
- `README.tr.md` - Türkçe README dosyası

### Değiştirilen Dosyalar

#### 1. `ui/model.go` (TUI Arayüzü)
- **Sekme isimleri:**
  - "Overview" → "Genel Bakış"
  - "Global News" → "Küresel Haberler"
  - "Local" → "Yerel"

- **Başlıklar ve etiketler:**
  - "real-time intelligence" → "gerçek zamanlı istihbarat"
  - "loading..." → "yükleniyor..."
  - "updated" → "güncellendi"

- **Hava Durumu Paneli:**
  - "Feels like" → "Hissedilen"
  - "Day/Hi/Lo/Rain" → "Gün/Max/Min/Yağış"
  - "Today" → "Bugün"
  - "Humidity/Wind/Visibility" → "Nem/Rüzgar/Görüş"

- **Piyasalar Paneli:**
  - "SYM/NAME/PRICE/24H%" → "SEM/İSİM/FİYAT/24S%"
  - "INDICES" → "ENDEKSLER"
  - "COMMODITIES" → "EMTİALAR"

- **Tahmin Piyasaları:**
  - "QUESTION/YES%/ENDS" → "SORU/EVET%/BİTİŞ"

- **Ülke Risk Paneli:**
  - "COUNTRY RISK INDEX" → "ÜLKE RİSK ENDEKSİ"
  - "Press [b] to generate risk scores" → "Risk puanlarını oluşturmak için [b] tuşuna basın"

- **Yerel Haberler:**
  - "WEATHER" → "HAVA DURUMU"
  - "LOCAL NEWS" → "YEREL HABERLER"
  - "LOCAL BRIEF" → "YEREL ÖZET"
  - "ARTICLES" → "MAKALELER"

- **Durum Mesajları:**
  - "Opening: ..." → "Açılıyor: ..."
  - "No URL available" → "URL mevcut değil"
  - "Brief generated and cached" → "Özet oluşturuldu ve önbelleğe alındı"

- **Tuş İpuçları:**
  - "navigate" → "gezin"
  - "open in browser" → "tarayıcıda aç"
  - "scroll" → "kaydır"
  - "switch" → "değiştir"
  - "refresh" → "yenile"
  - "quit" → "çık"

#### 2. `ui/setup.go` (Kurulum Sihirbazı)
- **Tüm adımlar Türkçe:**
  - "Select your preferred LLM" → "Tercih ettiğiniz LLM'yi seçin"
  - "Enter your API key" → "API anahtarınızı girin"
  - "Enter your location" → "Konumunuzu girin"
  - "Select temperature unit" → "Sıcaklık birimini seçin"
  - "Setup complete" → "Kurulum tamamlandı"

- **Placeholder'lar:**
  - "Paste your API key here" → "API anahtarınızı buraya yapıştırın"
  - "e.g., Lisbon" → "örn., İstanbul"
  - "e.g., PT" → "örn., TR"

- **Yönlendirme metinleri:**
  - "↑↓ select  tab/enter confirm  esc quit" → "↑↓ seçim  tab/enter onay  esc çık"

#### 3. `intel/intel.go` (AI İstihbarat Motoru)
- **AI Prompt'ları:**
  - Global brief prompt'u zaten Türkçe idi
  - Local brief prompt'u Türkçe'ye çevrildi
  - System prompt'ları (Claude için) Türkçe

- **Hata Mesajları:**
  - "No LLM_API_KEY set" → "LLM_API_KEY ayarlanmamış"
  - "No news items available" → "Özetlenecek haber bulunamadı"

- **Veri Bölüm Başlıkları:**
  - "LOCAL NEWS HEADLINES" → "YEREL HABER MANŞETLERİ"
  - "CURRENT WEATHER" → "MEVCUT HAVA DURUMU"
  - "FORECAST" → "HAVA TAHMİNİ"

## 🎯 Kullanım

Uygulama artık tamamen Türkçe arayüz ile çalışır. Tüm menüler, butonlar, durum mesajları ve yardım metinleri Türkçe'dir.

### Kurulum
Orijinal kurulum talimatları geçerlidir:
```bash
git clone https://github.com/cenktekin/watchtower.git
cd watchtower
go mod tidy
make run
```

### İlk Çalıştırma
İlk çalıştırmada Türkçe kurulum sihirbazı sizi karşılayacak:
1. LLM sağlayıcı seçin
2. API anahtarınızı girin (isteğe bağlı, yerel modeller için gerekmez)
3. Konumunuzu girin
4. Sıcaklık birimini seçin

## 📊 İstatistikler

- **Değiştirilen dosya:** 3
- **Yeni dosya:** 1 (README.tr.md)
- **Toplam satır değişikliği:** ~575 satır
  - 399 ekleme
  - 176 silme

## 🔀 Orijinal Repo ile Senkronizasyon

Bu fork, orijinal repo (`lajosdeme/watchtower`) ile senkronize tutulmalıdır:

```bash
git remote add upstream https://github.com/lajosdeme/watchtower.git
git fetch upstream
git merge upstream/main
```

## 📝 Notlar

- Kod yapısı ve işlevselliği değiştirilmemiştir
- Sadece kullanıcıya görünen metinler çevrilmiştir
- API çağrıları ve veri formatları aynı kalmıştır
- Go formatlama araçları (`go fmt`) uygulanmıştır

## 🤝 Katkıda Bulunma

Türkçe çeviriyi iyileştirmek için PR'lar memnuniyetle karşılanır.

## 📄 Lisans

Orijinal proje ile aynı - MIT License
