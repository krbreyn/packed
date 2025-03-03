package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// open zip file
// read dir.packed for structure
// read all the data in advance (could be exploited with really large files)
// display menu and choices
// use bubble tea viewport bubble for reading

func main() {
	m := initialModel()

	if len(os.Args) != 2 {
		fmt.Println("usage: packer [fiile]")
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

	p := tea.NewProgram(m)

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

func unpackFile(file *os.File) (map[string]string, error) {
	tr := tar.NewReader(file)

	entries := make(map[string]string)

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
			entries[hdr.Name] = buf.String()
		}
	}

	return entries, nil
}

// bubble tea stuff
func initialModel() model {
	m := model{exitMsg: "goodbye, world!", displayMsg: "hello, world!"}
	return m
}

type model struct {
	exitMsg    string
	displayMsg string
	entries    map[string]string
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) View() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("===\n\t%s\n\n", m.displayMsg))

	sb.WriteString("Files:\n")
	for name, data := range m.entries {
		sb.WriteString(fmt.Sprintf("%s: %s", name, data))
	}

	return sb.String()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case tea.KeyMsg:
		key := msg.String()

		switch key {

		case "ctrl+c", "ctrl+d", "q":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
	}

	return m, tea.Batch(cmds...)
}
