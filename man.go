package main

import (
	crand "crypto/rand"
	"log"
	"math/big"
	mrand "math/rand/v2"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type model struct {
	textarea       textarea.Model
	width          int
	height         int
	overlayVisible bool
	selectedButton int
	idxTextSet     int
	err            error
}

func initialModel() model {
	ta := textarea.New()
	ta.Prompt = "█ "
	ta.ShowLineNumbers = false
	ta.SetVirtualCursor(false)
	ta.Focus()
	placeholder, idxTexts := getRandomIndex()
	ta.Placeholder = placeholder

	return model{
		textarea:       ta,
		err:            nil,
		overlayVisible: false,
		selectedButton: 0,
		idxTextSet:     idxTexts,
	}
}

var confirmBox = lipgloss.NewStyle().
	Width(56).
	Padding(1, 2).
	Background(lipgloss.Alpha(lipgloss.Black, 0.2)).
	Align(lipgloss.Center)

var activeButton = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("0")).
	Background(lipgloss.Color("15")).
	Padding(0, 1)

var inactiveButton = lipgloss.NewStyle().
	Background(lipgloss.Alpha(lipgloss.Black, 0.2)).
	Foreground(lipgloss.Color("252")).
	Padding(0, 1)

var buttonGap = lipgloss.NewStyle().
	Background(lipgloss.Alpha(lipgloss.Black, 0.2))

func (m model) Init() tea.Cmd {
	return textarea.Blink
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.overlayVisible {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.overlayVisible = false
				return m, nil
			case "left", "h", "shift+tab":
				m.selectedButton = 0
				return m, nil
			case "right", "l":
				m.selectedButton = 1
				return m, nil
			case "tab":
				m.selectedButton = 1 - m.selectedButton
				return m, nil
			case "enter":
				if m.selectedButton == 0 {
					placeholder, idxTexts := getRandomIndex()
					m.textarea.Reset()
					m.textarea.Placeholder = placeholder
					m.idxTextSet = idxTexts
				}
				m.overlayVisible = false
				return m, nil
			}

			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, tea.Quit
		case "ctrl+f":
			m.overlayVisible = true
			m.selectedButton = 0
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetHeight(max(1, m.height-2))
		m.textarea.SetWidth(m.width)
		m.textarea.MaxWidth = msg.Width
	case error:
		m.err = msg
		return m, nil
	}

	if m.overlayVisible {
		return m, nil
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	const footer = "\nctrl+c or esc to quit   ctrl+f to feed the void"

	var c *tea.Cursor
	if !m.overlayVisible && !m.textarea.VirtualCursor() {
		c = m.textarea.Cursor()
	}

	f := strings.Join([]string{
		m.textarea.View(),
		footer,
	}, "\n")

	if m.overlayVisible {
		t := textSets[m.idxTextSet]
		body := strings.Join([]string{
			t.confirmation,
			"",
			renderButtons(m.selectedButton, t.buttonConfirm, t.buttonReturn),
		}, "\n")

		box := confirmBox.Render(body)
		x := max(0, (m.width-lipgloss.Width(box))/2)
		y := max(0, (m.height-lipgloss.Height(box))/2)

		baseLayer := lipgloss.NewLayer(f).X(0).Y(0).Z(0)
		boxLayer := lipgloss.NewLayer(box).X(x).Y(y).Z(1)
		f = lipgloss.NewCompositor(baseLayer, boxLayer).Render()
	}

	v := tea.NewView(f)
	v.AltScreen = true
	v.Cursor = c
	return v
}

func renderButtons(selected int, confirmLabel, returnLabel string) string {
	confirm := inactiveButton.Render("[ " + confirmLabel + " ]")
	ret := inactiveButton.Render("[ " + returnLabel + " ]")

	if selected == 0 {
		confirm = activeButton.Render("[ " + confirmLabel + " ]")
	} else {
		ret = activeButton.Render("[ " + returnLabel + " ]")
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, confirm, buttonGap.Render("  "), ret)
}

type textSet struct {
	placeholder   string
	confirmation  string
	buttonConfirm string
	buttonReturn  string
}

var textSets = []textSet{
	{
		placeholder:   "Write it down. Watch it dissolve. It was never yours to keep.",
		confirmation:  "Time to dissolve it.\nIt was never yours to keep anyway.",
		buttonConfirm: "Let it dissolve",
		buttonReturn:  "Not yet",
	},
	{
		placeholder:   "Speak into the dark. The void receives everything and returns nothing.",
		confirmation:  "The void is ready.\nSpeak, release, and return to silence.",
		buttonConfirm: "Send it",
		buttonReturn:  "Keep speaking",
	},
	{
		placeholder:   "Ready to stop feeling that?",
		confirmation:  "Let it stop here.\nThis is where it ends.",
		buttonConfirm: "Yes, done",
		buttonReturn:  "Not quite",
	},
	{
		placeholder:   "This is a safe place to fall apart a little.",
		confirmation:  "You're okay now.\nReady to close this chapter?",
		buttonConfirm: "I'm okay",
		buttonReturn:  "Give me a moment",
	},
	{
		placeholder:   "No one will read this. Say it anyway.",
		confirmation:  "Said. Done. Gone.\nNo one saw it. No one ever will.",
		buttonConfirm: "Clear it",
		buttonReturn:  "Keep going",
	},
	{
		placeholder:   "This input won't remember you. That's the gift.",
		confirmation:  "And now it forgets.\nNo trace. No record. Just you, lighter.",
		buttonConfirm: "Forget it",
		buttonReturn:  "Not yet",
	},
	{
		placeholder:   "Name it here so it loses some of its power.",
		confirmation:  "You named it. It's smaller now.\nReady to leave it behind?",
		buttonConfirm: "Leave it",
		buttonReturn:  "Keep writing",
	},
	{
		placeholder:   "I made this space just for me. Use it.",
		confirmation:  "You used it well.\nThis moment was just for you. Ready to close it?",
		buttonConfirm: "Close it",
		buttonReturn:  "Stay a little longer",
	},
	{
		placeholder:   "Your feelings are real. They deserve somewhere to go.",
		confirmation:  "They went somewhere. It's done.\nYou gave them a place. Now let them rest.",
		buttonConfirm: "Let them rest",
		buttonReturn:  "Not yet",
	},
	{
		placeholder:   "You've been holding this long enough.",
		confirmation:  "You can put it down now.\nFor real this time.",
		buttonConfirm: "Put it down",
		buttonReturn:  "Not yet",
	},
}

func getRandomIndex() (string, int) {
	options := len(textSets)
	max := big.NewInt(int64(options))

	num, err := crand.Int(crand.Reader, max)
	if err != nil {
		num := mrand.IntN(options)
		return textSets[num].placeholder, int(num)
	}
	idx := int(num.Int64())
	return textSets[idx].placeholder, idx
}
