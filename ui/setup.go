package ui

import (
	"context"
	"fmt"
	"watchtower/config"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	stepSelectProvider = iota
	stepAPIKey
	stepLocation
	stepTempUnit
	stepSaving
	stepDone
)

var providers = []string{"groq", "openai", "deepseek", "gemini", "claude", "ollama", "openrouter", "local"}

var tempUnits = []string{"celsius", "fahrenheit"}

var asciiTitle = ` 888     888          888            888      888                                           
 888   o  888          888            888      888                                           
 888  d8b  888          888            888      888                                           
 888 d888b 888  8888b.  888888 .d8888b 88888b.  888888 .d88b.  888  888  888  .d88b.  888d888 
 888d88888b888     "88b 888   d88P"    888 "88b 888   d88""88b 888  888  888 d8P  Y8b 888P"   
 88888P Y88888 .d888888 888   888      888  888 888   888  888 888  888  888 88888888 888     
 8888P   Y8888 888  888 Y88b. Y88b.    888  888 Y88b. Y88..88P Y88b 888 d88P Y8b.     888     
 888P     Y888 "Y888888  "Y888 "Y8888P 888  888  "Y888 "Y88P"   "Y8888888P"   "Y8888  888`

type SetupModel struct {
	step                int
	selectedIdx         int
	tempUnitSelectedIdx int

	apiKeyInput  textinput.Model
	cityInput    textinput.Model
	countryInput textinput.Model

	spinner   spinner.Model
	geocoding bool
	saving    bool
	err       string

	width  int
	height int
}

func NewSetupModel() SetupModel {
	apiKeyInput := textinput.New()
	apiKeyInput.Placeholder = "API anahtarınızı buraya yapıştırın"
	apiKeyInput.EchoMode = textinput.EchoPassword
	apiKeyInput.EchoCharacter = '*'
	apiKeyInput.Focus()

	cityInput := textinput.New()
	cityInput.Placeholder = "örn., İstanbul"
	cityInput.Focus()

	countryInput := textinput.New()
	countryInput.Placeholder = "örn., TR"
	countryInput.CharLimit = 2

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = StyleSpinner

	return SetupModel{
		step:         stepSelectProvider,
		selectedIdx:  0,
		apiKeyInput:  apiKeyInput,
		cityInput:    cityInput,
		countryInput: countryInput,
		spinner:      sp,
	}
}

func (m SetupModel) Init() tea.Cmd {
	return func() tea.Msg {
		return spinner.TickMsg{}
	}
}

func (m SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.apiKeyInput.Width = minInt(50, msg.Width-20)
		m.cityInput.Width = minInt(30, msg.Width-20)
		m.countryInput.Width = 4

	case tea.KeyMsg:
		switch m.step {
		case stepSelectProvider:
			switch msg.Type {
			case tea.KeyUp, tea.KeyShiftTab:
				m.selectedIdx = (m.selectedIdx - 1 + len(providers)) % len(providers)
			case tea.KeyDown, tea.KeyTab:
				m.selectedIdx = (m.selectedIdx + 1) % len(providers)
			case tea.KeyEnter:
				// Skip API key step for providers that don't require it
				selectedProvider := providers[m.selectedIdx]
				if selectedProvider == "ollama" || selectedProvider == "local" {
					m.step = stepLocation
				} else {
					m.step = stepAPIKey
				}
				cmds = append(cmds, func() tea.Msg {
					return tea.WindowSizeMsg{
						Width:  m.width,
						Height: m.height,
					}
				})
			}

		case stepAPIKey:
			switch msg.Type {
			case tea.KeyEnter:
				if m.apiKeyInput.Value() != "" {
					m.step = stepLocation
				}
				cmds = append(cmds, func() tea.Msg {
					return tea.WindowSizeMsg{
						Width:  m.width,
						Height: m.height,
					}
				})
			default:
				var cmd tea.Cmd
				m.apiKeyInput, cmd = m.apiKeyInput.Update(msg)
				cmds = append(cmds, cmd)
			}

		case stepLocation:
			switch msg.Type {
			case tea.KeyEnter:
				if m.cityInput.Value() != "" && m.countryInput.Value() != "" {
					m.step = stepTempUnit
					cmds = append(cmds, func() tea.Msg {
						return tea.WindowSizeMsg{
							Width:  m.width,
							Height: m.height,
						}
					})

				}
			case tea.KeyTab:
				m.cityInput.Blur()
				m.countryInput.Focus()
			default:
				var cmd1, cmd2 tea.Cmd
				m.cityInput, cmd1 = m.cityInput.Update(msg)
				m.countryInput, cmd2 = m.countryInput.Update(msg)
				cmds = append(cmds, cmd1, cmd2)
			}

		case stepTempUnit:
			switch msg.Type {
			case tea.KeyUp, tea.KeyShiftTab:
				m.tempUnitSelectedIdx = (m.tempUnitSelectedIdx - 1 + len(tempUnits)) % len(tempUnits)
			case tea.KeyDown, tea.KeyTab:
				m.tempUnitSelectedIdx = (m.tempUnitSelectedIdx + 1) % len(tempUnits)
			case tea.KeyEnter:
				m.step = stepSaving
				m.geocoding = true
				cmds = append(cmds, m.doGeocode())
				cmds = append(cmds, func() tea.Msg {
					return tea.WindowSizeMsg{
						Width:  m.width,
						Height: m.height,
					}
				})
			}

		case stepSaving:
			if msg.Type == tea.KeyEnter && m.err != "" {
				m.step = stepLocation
				m.err = ""
			}

		case stepDone:
			return m, tea.Quit
		}

		switch msg.Type {
		case tea.KeyEsc:
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case geocodeResultMsg:
		m.geocoding = false
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.saving = true
			cmds = append(cmds, m.doSave(msg.lat, msg.lon))
		}

	case saveResultMsg:
		m.saving = false
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.step = stepDone
		}
	}

	return m, tea.Batch(cmds...)
}

func (m SetupModel) View() string {
	if m.width == 0 {
		return "Kurulum başlatılıyor..."
	}

	stepIndicator := StyleStepIndicator.Render(fmt.Sprintf("[%d/5]", m.step+1))
	header := lipgloss.JoinVertical(
		lipgloss.Center,
		stepIndicator,
		StyleSetupTitle.Render("WATCHTOWER"),
	)

	var content string
	switch m.step {
	case stepSelectProvider:
		content = m.renderProviderStep()
	case stepAPIKey:
		content = m.renderAPIKeyStep()
	case stepLocation:
		content = m.renderLocationStep()
	case stepTempUnit:
		content = m.renderTempUnitStep()
	case stepSaving:
		content = m.renderSavingStep()
	case stepDone:
		content = m.renderDoneStep()
	}

	footer := StyleMuted.Render("↑↓ seçim  tab/enter onay  esc çık")

	centeredContent := lipgloss.Place(
		m.width-4, m.height-6,
		lipgloss.Center, lipgloss.Center,
		content,
	)

	container := lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		"",
		centeredContent,
		"",
		footer,
	)

	return StyleSetupPane.Width(m.width).Render(container)
}

func (m SetupModel) renderProviderStep() string {
	selected := providers[m.selectedIdx]

	var items []string
	for i, p := range providers {
		label := p
		// Add indicator for free/local providers
		if p == "ollama" || p == "local" {
			label = p + " (anahtar gerekmez)"
		} else if p == "groq" || p == "openrouter" {
			label = p + " (ücretsiz katman mevcut)"
		}

		if i == m.selectedIdx {
			items = append(items, StyleSelectedItem.Render("> "+label))
		} else {
			items = append(items, StyleMuted.Render("  "+label))
		}
	}

	content := StyleAccent.Render(asciiTitle) + "\n\n"
	content += StylePrompt.Render("Tercih ettiğiniz LLM'yi seçin:") + "\n\n"
	content += lipgloss.JoinVertical(lipgloss.Left, items...)
	content += "\n\n" + StyleMuted.Render("Seçilen: "+StyleAccent.Render(selected))

	return content
}

func (m SetupModel) renderAPIKeyStep() string {
	selectedProvider := providers[m.selectedIdx]

	content := StyleAccent.Render(asciiTitle) + "\n\n"
	content += StylePrompt.Render("Seçilen: "+selectedProvider) + "\n\n"
	content += selectedProvider + " API anahtarınızı girin:\n\n"
	content += m.apiKeyInput.View() + "\n\n"

	// Provider-specific hints
	hint := "API anahtarınız yerel olarak saklanır ve cihazınızdan asla çıkmaz."
	if selectedProvider == "openrouter" {
		hint = "OpenRouter ücretsiz modeller sağlar. Anahtarınızı openrouter.ai adresinden alın"
	} else if selectedProvider == "groq" {
		hint = "Groq ücretsiz katman sağlar. Anahtarınızı console.groq.com adresinden alın"
	}

	content += StyleHint.Render(hint)

	return content
}

func (m SetupModel) renderLocationStep() string {
	content := StyleAccent.Render(asciiTitle) + "\n\n"
	content += StylePrompt.Render("Hava durumu ve yerel haberler için konumunuzu girin:") + "\n\n"
	content += "  Şehir:          " + m.cityInput.View() + "\n"
	content += "  Ülke kodu: " + m.countryInput.View() + "\n\n"

	if m.err != "" {
		content += StyleError.Render("Hata: "+m.err) + "\n"
		content += StyleHint.Render("Geri dönüp tekrar denemek için Enter'a basın.")
	} else {
		content += StyleHint.Render("Örnek: Lisbon / PT, New York / US, London / GB")
	}

	return content
}

func (m SetupModel) renderTempUnitStep() string {
	content := StyleAccent.Render(asciiTitle) + "\n\n"
	content += StylePrompt.Render("Sıcaklık birimini seçin:") + "\n\n"

	for i, unit := range tempUnits {
		if i == m.tempUnitSelectedIdx {
			content += StyleSelectedItem.Render("> "+unit) + "\n"
		} else {
			content += StyleMuted.Render("  "+unit) + "\n"
		}
	}

	content += "\n" + StyleHint.Render("Seçmek için ↑↓ veya tab kullanın, devam etmek için Enter")

	return content
}

func (m SetupModel) renderSavingStep() string {
	var lines []string

	if m.geocoding {
		lines = append(lines, m.spinner.View()+" Koordinatlar aranıyor...")
	}
	if m.saving {
		lines = append(lines, m.spinner.View()+" Yapılandırma kaydediliyor...")
	}
	if m.err != "" {
		lines = append(lines, StyleError.Render("Hata: "+m.err))
		lines = append(lines, StyleHint.Render("Geri dönüp tekrar denemek için Enter'a basın."))
	}

	return lipgloss.JoinVertical(lipgloss.Center, lines...)
}

func (m SetupModel) renderDoneStep() string {
	provider := providers[m.selectedIdx]
	location := m.cityInput.Value() + ", " + m.countryInput.Value()

	msg := StyleSuccess.Render("Kurulum tamamlandı!") + "\n\n"
	msg += "  Sağlayıcı: " + StyleAccent.Render(provider) + "\n"
	msg += "  Konum: " + StyleAccent.Render(location) + "\n\n"
	msg += StyleHint.Render("Watchtower'ı başlatmak için herhangi bir tuşa basın...")

	return msg
}

func (m SetupModel) doGeocode() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		city := m.cityInput.Value()
		country := m.countryInput.Value()
		lat, lon, err := config.Geocode(ctx, city, country)
		return geocodeResultMsg{lat: lat, lon: lon, err: err}
	}
}

func (m SetupModel) doSave(lat, lon float64) tea.Cmd {
	return func() tea.Msg {
		// No API key needed for ollama/local
		apiKey := m.apiKeyInput.Value()
		selectedProvider := providers[m.selectedIdx]
		if selectedProvider == "ollama" || selectedProvider == "local" {
			apiKey = ""
		}

		cfg := &config.Config{
			LLMProvider: selectedProvider,
			LLMAPIKey:   apiKey,
			Location: config.Location{
				City:      m.cityInput.Value(),
				Country:   m.countryInput.Value(),
				Latitude:  lat,
				Longitude: lon,
			},
			TempUnit:       tempUnits[m.tempUnitSelectedIdx],
			RefreshSec:     120,
			BriefCacheMins: 60,
			CryptoPairs:    []string{"bitcoin", "ethereum", "dogecoin", "usd-coin"},
		}
		err := config.Save(cfg)
		return saveResultMsg{err: err}
	}
}

type geocodeResultMsg struct {
	lat float64
	lon float64
	err error
}

type saveResultMsg struct {
	err error
}
