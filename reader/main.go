package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := initialModel()

	if len(os.Args) != 2 {
		fmt.Println("usage: packer [file]")
		os.Exit(0)
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	entries, err := unpackFile(file)
	if err != nil {
		panic(err)
	}
	m.entries = entries

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if m, err := p.Run(); err != nil {
		fmt.Println("err:", err)
		os.Exit(1)
	} else {
		s, ok := m.(model)
		if ok && s.exitMsg != "" {
			fmt.Println(s.exitMsg)
			fmt.Println(s.entries)
		}
		os.Exit(0)
	}
}

func unpackFile(file *os.File) ([]entry, error) {
	tr := tar.NewReader(file)

	var entries []entry

	for {
		hdr, err := tr.Next()

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if hdr.Typeflag == tar.TypeReg {
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, tr); err != nil {
				return nil, err
			}
			entries = append(entries, entry{name: hdr.Name, data: buf.String()})
		}
	}

	return entries, nil
}

/*
	bubble tea stuff
*/

func initialModel() model {
	m := model{}
	return m
}

type model struct {
	exitMsg  string
	loc      int
	picked   entry
	entries  []entry
	viewport viewport.Model
	ready    bool
}

type entry struct {
	name, data string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) View() string {
	var sb strings.Builder

	switch m.picked {

	case entry{}:
		sb.WriteString(fmt.Sprintf("===\n\t%s %s\n\n", "hello, world!", m.picked.name))

		var i int
		sb.WriteString("Files:\n")
		for _, entry := range m.entries {
			if i == m.loc {
				sb.WriteString(">")
			}
			sb.WriteString(fmt.Sprintf(" %s\n", entry.name))
			i++
		}
		return sb.String()

	default:
		return m.viewport.View()

	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case tea.KeyMsg:
		key := msg.String()

		switch key {

		case "ctrl+c", "ctrl+d", "q":
			return m, tea.Quit

		case "esc":
			m.picked = entry{}

		case "enter":
			m.picked = m.entries[m.loc]
			m.viewport.SetContent(m.picked.data)

		case "down":
			if m.loc < len(m.entries)-1 {
				m.loc++
			}

		case "up":
			if m.loc > 0 {
				m.loc--
			}
		}

	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height)
			m.viewport.SetContent(m.picked.data)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
