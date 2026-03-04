# 🌍 Watchtower

Temiz, minimal, terminal tabanlı küresel istihbarat gösterge paneli.

> **📢 Bu Fork'ta Yeni:** OpenRouter ve Ollama desteği eklendi! API key gerektirmeyen yerel LLM modelleri artık kullanabilirsiniz.

![wt](https://i.imgur.com/p4BiORi.gif)

## Neden Watchtower?

İnternet bilgiyi erişilebilir kıldı—ancak gürültüde yolunu bulmak bunaltıcı hale geldi. WorldMonitor gibi OSINT araçları güçlüdür, ancak sadece veri noktasına ihtiyaç duyan istihbarat profesyonelleri için tasarlanmıştır. Veri selinde boğulmadan bilgi sahibi olmak isteyen ortalama kullanıcı için bir boşluk var.

**Watchtower bu boşluğu dolduruyor.** Tamamen terminalinizde yaşar—tarayıcı sekmesi yok, ağır web uygulamaları yok. Hafif, hızlıdır ve yalnızca tek bir API anahtarı gerektirir (ve bu da sadece AI özet özelliği için isteğe bağlıdır). Terminalinizi açın ve dünyada neler olduğunu görün.

## ✨ Bu Fork'taki Yeni Özellikler

### 🆕 Orijinal Repoda OLMAYAN Özellikler:

| Özellik | Açıklama |
|---------|----------|
| **🔌 OpenRouter** | Ücretsiz LLM modelleri (meta-llama/llama-3-8b-instruct:free) |
| **🏠 Ollama** | Yerel, offline LLM çalıştırma (tamamen ücretsiz, API key gerekmez) |
| **🔓 API Key Opsiyonel** | Ollama ve Local sağlayıcılar için API key **gerekmez** |
| **⏱️ Timeout Artırıldı** | 30s → 120s (uzun AI yanıtları için daha güvenilir) |
| **🇹🇷 Türkçe Arayüz** | Tam Türkçe TUI, menüler ve AI prompt'ları |

### Desteklenen LLM Sağlayıcıları:

| Sağlayıcı | API Key | Ücretsiz Katman | Notlar |
|-----------|---------|-----------------|--------|
| **Groq** | ✅ Gerekli | ✅ Var | Hızlı, ücretsiz katman cömert |
| **OpenRouter** | ✅ Gerekli | ✅ Var | Ücretsiz modeller mevcut |
| **Ollama** | ❌ Gerekmez | ✅ Tamamen ücretsiz | Yerel model, offline çalışır |
| **Local** | ❌ Gerekmez | ✅ Tamamen ücretsiz | Yerel model |
| OpenAI | ✅ Gerekli | ❌ Yok | - |
| Deepseek | ✅ Gerekli | ✅ Var | - |
| Gemini | ✅ Gerekli | ✅ Var | - |
| Anthropic | ✅ Gerekli | ❌ Yok | - |

## Özellikler

| Sekme | İçerik |
|-----|----------|
| **Küresel Haberler** | 100+ RSS beslemesi, anahtar kelime tehdit sınıflandırması (KRİTİK/YÜKSEK/ORTA/DÜŞÜK/BİLGİ) |
| **Piyasalar** | Canlı kripto (CoinGecko) + Polymarket tahmin piyasaları + hisseler + emtialar |
| **Yerel** | Open-Meteo hava durumu (ücretsiz, anahtar gerekmez) + coğrafi hedefli yerel haberler |
| **İstihbarat Özeti** | En önemli manşetlerin AI sentezi |

Tüm API'ler ücretsiz — yalnızca LLM anahtarı gerektirir (Groq ücretsiz katmanı cömerttir).

## Kurulum

Platformunuza en uygun seçeneği seçin.

### Evrensel kurulum betiği
```bash
curl -fsSL https://raw.githubusercontent.com/lajosdeme/watchtower/main/install.sh
```

### Homebrew
```bash
brew tap lajosdeme/watchtower
brew install watchtower
```

### AUR
```bash
yay -S watchtower-bin
```

### .deb (Ubuntu/Debian)
```bash
# sürüm sayfasından indirin, sonra:
sudo dpkg -i watchtower_1.0.0_linux_amd64.deb
watchtower --version
```

### .rpm (Fedora)
```bash
# sürüm sayfasından indirin, sonra:
sudo rpm -i watchtower_1.0.0_linux_amd64.rpm
watchtower --version
```

### Scoop (Windows)
```bash
scoop bucket add watchtower https://github.com/lajosdeme/scoop-watchtower
scoop install watchtower
```

### Kaynak koddan
```bash
git clone https://github.com/lajosdeme/watchtower
cd watchtower
go mod tidy
make run
# veya docker kullanarak:
make docker-run
```

### Go PATH'te ise

```bash
go install github.com/lajosdeme/watchtower@latest
```

## Kurulum

İlk çalıştırmada, Watchtower birkaç şeyi yapılandırmanızı isteyecek:

1. **LLM sağlayıcı seçin** — Groq, OpenRouter, Ollama, OpenAI, Deepseek, Gemini, Anthropic veya yerel model
   - 💡 **Öneri:** Ücretsiz kullanım için **Groq** veya **OpenRouter**
   - 🏠 **Tamamen ücretsiz:** **Ollama** veya **Local** (API key gerekmez!)
2. **API anahtarınızı yapıştırın** — Sadece cloud sağlayıcılar için gerekli
   - Ollama/Local seçtiyseniz bu adım **atlanır**
   - API anahtarınız `~/.config/watchtower/config.yaml` dosyasında saklanır, cihazınızdan asla çıkmaz
3. **Konumunuzu belirtin** — Yerel hava durumu ve haberler için şehrinizi ve koordinatlarınızı girin

![setup](https://i.imgur.com/7L4soxv.gif)

Bu kadar! Uygulama ayarlarınızı kaydeder ve kullanmaya hazırsınız.

### 🆓 Ücretsiz AI Kullanımı

**Tamamen ücretsiz kullanmak için:**

1. **Ollama** (Önerilen):
   ```bash
   # Ollama'yı kurun
   curl -fsSL https://ollama.com/install.sh | sh
   
   # Model indirin
   ollama pull llama3
   
   # Watchtower'da "ollama" seçin - API key gerekmez!
   ```

2. **Groq** (Ücretsiz katman):
   - console.groq.com adresinden ücretsiz API key alın
   - Hızlı ve cömert ücretsiz limitler

3. **OpenRouter** (Ücretsiz modeller):
   - openrouter.ai adresinden API key alın
   - Meta-Llama-3-8b-Instruct:free gibi ücretsiz modeller

## Tuş Atamaları

| Tuş | Aksiyon |
|-----|--------|
| `1` `2` `3` `4` | Sekmeye atla |
| `Tab` / `Shift+Tab` | Sonraki / önceki sekme |
| `← →` / `h l` | Sekme değiştir |
| `↑ ↓` / `j k` | İçeriği kaydır |
| `d` / `u` | Yarım sayfa aşağı/yukarı |
| `g` / `G` | En üst / en alt |
| `r` | Tüm verileri zorla yenile |
| `b` | AI özeti oluştur (Özet sekmesinde) |
| `q` / `Ctrl+C` | Çık |

## Veri Kaynakları

| Kaynak | Ne | Anahtar? |
|--------|------|------|
| Reuters, BBC, AP, Al Jazeera, vb. | Küresel haberler | Yok (RSS) |
| Google News | Yerel haberler | Yok (RSS) |
| CoinGecko | Kripto fiyatları | Yok (public API) |
| Polymarket | Tahmin piyasaları | Yok (public API) |
| Yahoo Finance | Hisseler ve emtialar | Yok |
| Open-Meteo | Hava durumu | Yok |
| **Groq / OpenRouter / Ollama** | AI özet | **Yok** (ücretsiz) |
| OpenAI / Anthropic / Deepseek / Gemini | AI özet | Gerekli |

## Teknoloji Yığını

- **Dil:** Go 1.22
- **TUI:** [bubbletea](https://github.com/charmbracelet/bubbletea) + [lipgloss](https://github.com/charmbracelet/lipgloss) + [bubbles](https://github.com/charmbracelet/bubbles)
- **RSS:** [gofeed](https://github.com/mmcdole/gofeed)
- **Yapılandırma:** [viper](https://github.com/spf13/viper)

## Katkıda Bulunma

Katkılar memnuniyetle karşılanır! Yeni özellikler eklerken, hataları düzeltirken veya belgeleri iyileştirirken:

1. Repoyu fork edin
2. Bir özellik dalı oluşturun (`git checkout -b feature/amazing-feature`)
3. Değişikliklerinizi commit edin (`git commit -m 'Add amazing feature'`)
4. Dala push edin (`git push origin feature/amazing-feature`)
5. Pull Request açın

Kodun formatlandığından (`go fmt ./`) ve testleri geçtiğinden (`go test ./...`) emin olun.

## Watchtower'ı Destekleme

Watchtower'ı faydalı buluyorsanız, projeyi desteklemeyi düşünün:

- **Repo'ya yıldız verin** — görünürlüğe yardımcı olur
- **Paylaşın** — arkadaşlarınıza ve meslektaşlarınıza anlatın
- **Katkıda bulunun** — kod, belgeler, geri bildirim
- **Hata bildirin** — daha iyi hale gelmesine yardımcı olun

## Lisans

MIT Lisansı — detaylar için [LICENSE](LICENSE) dosyasına bakın.

## Yazar

- Orijinal: [Lajos Deme](https://github.com/lajosdeme)
- Bu fork: [cenktekin](https://github.com/cenktekin)

## 🔗 Bağlantılar

- **Orijinal Repo:** https://github.com/lajosdeme/watchtower
- **Bu Fork:** https://github.com/cenktekin/watchtower
- **Değişiklikler:** [CHANGES.md](CHANGES.md) dosyasında detaylı döküm
