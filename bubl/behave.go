package bubl

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
	"github.com/zmnpl/ding/core"
)

type model struct {
	appHeightPercent          float64
	width                     int
	spinner                   spinner.Model
	preview                   viewport.Model
	inboundList               list.Model
	inboundColumnWidth        int
	directoryList             list.Model
	directoryColumnWidth      int
	directoryFileList         list.Model
	directoryFilesColumnWidth int

	help help.Model

	selectedInbound   list.Item
	selectedDirectory list.Item

	newNameInput       textinput.Model
	newNameHeaderStyle lipgloss.Style
	timeStamp          string

	ocrIndex   int
	ocrRunning bool

	statusMessage string

	ready bool
	focus int

	previewWidth int
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = myStyle.highlightStyle

	inbounds := InboundItemsAsBubblesList()
	inboundList := list.New(inbounds, itemDelegate{}, 0, 0)
	inboundList.Title = "Files"
	inboundList.SetShowHelp(false)
	inboundList.SetShowStatusBar(false)
	inboundList.SetFilteringEnabled(true)
	inboundList.Styles.Title = myStyle.titleStyleSelected
	inboundList.Styles.PaginationStyle = myStyle.paginationStyle
	//inboundList.Styles.HelpStyle = helpStyle

	selectedInbound := inboundList.SelectedItem()

	directories := DirectoriesAsBubblesList()
	directoryList := list.New(directories, itemDelegate{}, 0, 0)
	directoryList.Title = "Directories"
	directoryList.SetShowHelp(false)
	directoryList.SetShowStatusBar(false)
	directoryList.SetFilteringEnabled(true)
	directoryList.Styles.Title = myStyle.titleStyle
	directoryList.Styles.PaginationStyle = myStyle.paginationStyle
	//directoryList.Styles.HelpStyle = helpStyle

	selectedDirectory := directoryList.SelectedItem()

	// TODO - Bug, crashes if no directory files available

	directoryFileList := list.New(nil, dimmedDelegate{}, 0, 0)
	directoryFileList.SetShowTitle(false)
	directoryFileList.SetShowHelp(false)
	directoryFileList.SetShowStatusBar(false)
	directoryFileList.SetShowPagination(false)

	newNameInput := textinput.New()
	newNameInput.PlaceholderStyle = myStyle.styleInactiveText
	newNameInput.TextStyle = myStyle.styleActiveText
	newNameInput.Placeholder = "type ..."
	newNameInput.Prompt = ""
	newNameInput.Focus()
	newNameInput.CharLimit = 128
	newNameInput.Width = 32

	m := model{
		appHeightPercent:     0.4,
		spinner:              s,
		inboundList:          inboundList,
		inboundColumnWidth:   listMaxItemLength(inbounds),
		selectedInbound:      selectedInbound,
		directoryList:        directoryList,
		directoryColumnWidth: listMaxItemLength(directories),
		selectedDirectory:    selectedDirectory,
		directoryFileList:    directoryFileList,

		newNameInput:       newNameInput,
		newNameHeaderStyle: myStyle.titleStyleSelected,
		help:               help.New(),
		previewWidth:       35,

		timeStamp: core.GetTimestampFilePrefix(),
	}

	m = m.updateDirectoryFiles()
	m.directoryFilesColumnWidth = m.previewWidth

	return m.focusInbound()
}

func (m model) reactToWindowSize(msg tea.WindowSizeMsg) model {
	top, right, bottom, left := myStyle.docStyle.GetMargin()
	width := msg.Width - left - right
	height := msg.Height - top - bottom // - 5

	height = int(float64(height) * m.appHeightPercent)

	if height < 19 {
		height = 19
	}

	m.width = width

	// dummy to keep width
	if width < 10 {
		m.inboundColumnWidth = 2
	}

	// help
	helpHeight := 3
	m.help.Width = m.width

	m.inboundList.SetSize(m.inboundColumnWidth, height-3-2-helpHeight)
	m.directoryList.SetSize(m.directoryColumnWidth, height-3-2-helpHeight)
	m.directoryFileList.SetSize(m.directoryFilesColumnWidth, height-3-4-helpHeight)
	m.newNameInput.Width = m.width - lipgloss.Width(core.GetTimestampFilePrefix())

	// preview
	headerHeight := 3
	footerHeight := 10
	previewHeight := height - headerHeight - footerHeight - helpHeight

	if !m.ready {
		m.preview = viewport.New(m.previewWidth, previewHeight)
		m.preview.HighPerformanceRendering = useHighPerformanceRenderer

		m.preview.SetContent(myStyle.textDimmedStyle.Render(wrap.String(wordwrap.String(m.docPreview(), m.previewWidth), m.previewWidth)))
		m.ready = true
	} else {
		m.preview.Width = m.previewWidth
		m.preview.Height = previewHeight
	}

	return m
}

func (m model) updatePreviewViews() model {
	// update previews if items have changed
	if m.inboundList.SelectedItem() != m.selectedInbound {
		m.selectedInbound = m.inboundList.SelectedItem()
		if m.selectedInbound != nil {
			m.preview.SetContent(myStyle.textDimmedStyle.Render(wrap.String(wordwrap.String(m.docPreview(), m.previewWidth), m.previewWidth)))
		}
	}
	if m.directoryFileList.SelectedItem() != m.selectedDirectory {
		m.selectedDirectory = m.directoryList.SelectedItem()
		if m.selectedDirectory != nil {
			m.directoryFileList.SetItems(m.selectedDirectory.(directory).GetDirectoryFiles())
		}
	}

	return m
}

func (m model) focusDirectories() model {
	m.focus = FOCUS_DIRECTORIES
	m.statusMessage = "Select a directory..."

	m.inboundList.Styles.Title = myStyle.titleStyle
	m.directoryList.Styles.Title = myStyle.titleStyleSelected

	return m
}

func (m model) focusNewName() model {
	m.focus = FOCUS_NEWNAME
	m.statusMessage = "Enter a file name..."

	m.directoryList.Styles.Title = myStyle.titleStyle
	m.timeStamp = core.GetTimestampFilePrefix()
	m.newNameHeaderStyle = myStyle.titleStyleSelected

	return m
}

func (m model) focusInbound() model {
	m.focus = FOCUS_INBOUND
	m.statusMessage = "Select a file..."

	//m.inboundList.SetDelegate(listItemActiveDelegate)
	m.inboundList.Styles.Title = myStyle.titleStyleSelected
	m.newNameHeaderStyle = myStyle.titleStyle

	return m
}
