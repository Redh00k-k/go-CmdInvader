package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

type bullet struct {
	x, y int
}

type model struct {
	text       [][]rune
	playerX    int
	bullets    []bullet
	width      int
	height     int
	pierceMode bool
	score      int64
}

func initialModel() model {
	var input []byte
	var err error

	if term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Println("Usage: <command> | go-CmdInvader")
		os.Exit(0)
	}

	// Read pipe
	input, err = io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(1)
	}

	if len(strings.TrimSpace(string(input))) == 0 {
		fmt.Fprintln(os.Stderr, "[Error] 'input' is empty.")
		os.Exit(1)
	}

	lines := strings.Split(string(input), "\n")
	var text [][]rune
	for _, line := range lines {
		text = append(text, []rune(line))
	}

	width := 0
	for _, line := range text {
		if len(line) > width {
			width = len(line)
		}
	}

	return model{
		text:    text,
		playerX: width / 2,
		width:   width,
		height:  len(text) + 5,
		score:   0,
	}
}

func (m model) Init() tea.Cmd {
	return tick()
}

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

type tickMsg struct{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h", "a":
			if m.playerX > 0 {
				m.playerX--
			}
		case "right", "l", "d":
			if m.playerX < m.width-1 {
				m.playerX++
			}
		case " ":
			// append bullet
			m.bullets = append(m.bullets, bullet{x: m.playerX, y: m.height - 2})
		case "p":
			// enable penetration mode
			if !m.pierceMode {
				m.pierceMode = true
			} else {
				m.pierceMode = false
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tickMsg:
		var newBullets []bullet
		for _, b := range m.bullets {
			if b.y > 0 {
				b.y--
				// Hit a block
				if b.y < len(m.text) && b.x < len(m.text[b.y]) {
					if m.text[b.y][b.x] != ' ' {
						// Delete character
						m.text[b.y][b.x] = ' '
						m.score++

						if !m.pierceMode {
							// Delete the bullet by continue to next b loop
							continue
						}
					}
				}
				newBullets = append(newBullets, b)
			}
		}
		m.bullets = newBullets
		return m, tick()
	}
	return m, nil
}

func (m model) View() string {
	// Initialize display
	display := make([][]rune, m.height)
	for y := 0; y < m.height; y++ {
		display[y] = make([]rune, m.width)
		for x := 0; x < m.width; x++ {
			display[y][x] = ' '
		}
	}

	// Text
	for y := 0; y < len(m.text); y++ {
		for x := 0; x < len(m.text[y]); x++ {
			display[y][x] = m.text[y][x]
		}
	}

	// Bullets
	for _, b := range m.bullets {
		if b.y >= 0 && b.y < m.height && b.x >= 0 && b.x < m.width {
			display[b.y][b.x] = '|'
		}
	}

	// Player
	if m.playerX >= 0 && m.playerX < m.width {
		display[m.height-1][m.playerX] = 'A'
	}

	// Show text, bullets, player
	var b strings.Builder
	for y := 0; y < m.height; y++ {
		b.WriteString(string(display[y]))
		b.WriteRune('\n')
	}

	// Show bottom
	b.WriteString(fmt.Sprintf("\nScore : %d\n", m.score))
	if m.pierceMode {
		b.WriteString("Pierce Mode\n")
	}
	b.WriteString("\n←/→ or A/D or H/L: move player | SPACE: shot | P: Pierce mode | Q or CTRL+C: quit\n")

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
