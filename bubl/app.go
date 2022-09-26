package bubl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zmnpl/ding/core"
)

const (
	useHighPerformanceRenderer = false

	FOCUS_INBOUND     = 0
	FOCUS_DIRECTORIES = 1
	FOCUS_NEWNAME     = 2

	STATUS_MOVE_OK     = "Ok"
	STATUS_MOVE_FAILED = "Failed"
	STATUS_ERR         = "Err"
)

func Run() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m = m.reactToWindowSize(msg)

	case tea.KeyMsg:
		if m.inboundList.FilterState() == list.Filtering || m.directoryList.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			switch m.focus {
			case FOCUS_INBOUND:
				m = m.focusDirectories()
			case FOCUS_DIRECTORIES:
				m = m.focusNewName()
			case FOCUS_NEWNAME:
				if m.selectedInbound != nil && m.selectedInbound.(inboundItem).file != nil && m.selectedDirectory != nil && m.selectedDirectory.(directory).dir != nil {
					cmds = append(cmds, makeMoveCommand(m.selectedInbound.(inboundItem).file.Name(), m.timeStamp+m.newNameInput.Value(), m.selectedDirectory.(directory).dir.Name()))
				}
				m = m.focusInbound()
				return m, tea.Batch(cmds...)
			}

		case "f1":
			filename := m.selectedInbound.(inboundItem).file.Name()
			err := core.OpenDocExternal(filepath.Join(core.Inbound, filename))
			if err != nil {
				m.statusMessage = fmt.Sprintf("could not open file in default application: %s", STATUS_ERR)

			}

		case "f2":
			if !m.ocrRunning && len(m.inboundList.Items()) > m.inboundList.Index() {
				m.ocrRunning = true
				itm := m.inboundList.SelectedItem().(inboundItem)
				m.statusMessage = "Running ocf for " + itm.name
				return m, itm.makeOcrCommand(true)
			}

		case "f3":
			// start ocr
			if !m.ocrRunning && len(m.inboundList.Items()) > 0 {
				m.ocrRunning = true
				m.ocrIndex = 0
				itm := m.inboundList.Items()[m.ocrIndex].(inboundItem)
				m.statusMessage = "Running ocf for " + itm.name
				return m, itm.makeOcrCommand(false)
			}
		}

	case ocrMessageMulti:
		m.statusMessage = msg.message
		m.ocrIndex++
		if len(m.inboundList.Items()) > m.ocrIndex {
			itm := m.inboundList.Items()[m.ocrIndex].(inboundItem)
			m.statusMessage = "Running ocf for " + itm.name
			return m, itm.makeOcrCommand(false)
		}
		m.ocrRunning = false

	case ocrMessageSingle:
		m.statusMessage = msg.message
		m.ocrRunning = false

	case moveMsg:
		m.statusMessage = msg.messageText
		m.inboundList.RemoveItem(m.inboundList.Index())
		m.newNameInput.SetValue("")
		return m, nil

	default:
		if m.inboundList.FilterState() == list.Filtering || m.directoryList.FilterState() == list.Filtering {
			break
		}

		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// only update (pass message to) the ui element that has focus...
	var cmd tea.Cmd
	switch m.focus {
	case FOCUS_INBOUND:
		m.inboundList, cmd = m.inboundList.Update(msg)
	case FOCUS_DIRECTORIES:
		m.directoryList, cmd = m.directoryList.Update(msg)
	case FOCUS_NEWNAME:
		m.newNameInput, cmd = m.newNameInput.Update(msg)
	}

	// function which upates the previews
	m = m.updatePreviewViews()

	return m, cmd
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	var foo strings.Builder

	m.help.ShowAll = false
	foo.WriteString(myStyle.docStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			m.mainSection(),
			m.selectedFile(),
			m.newNameSection(),
			m.statusBar(),
			m.helpView(),
		)))

	return foo.String()
}
