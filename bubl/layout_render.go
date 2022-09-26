package bubl

import (
	"github.com/charmbracelet/lipgloss"
)

// part renders

func (m model) mainSection() string {
	ms := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Margin(0, 1, 0, 0).Width(m.inboundColumnWidth).Render(m.inboundList.View()),
		lipgloss.NewStyle().Margin(2, 1, 0, 0).Width(m.previewWidth).Render(myStyle.previewInactiveColorStyle.Render(m.preview.View())),
		lipgloss.NewStyle().Margin(0, 1, 0, 0).Width(m.directoryColumnWidth).Render(m.directoryList.View()),
		lipgloss.NewStyle().Margin(1, 0, 0, 0).Width(m.directoryFilesColumnWidth).Render(m.directoryFileList.View()),
	)

	if m.focus == FOCUS_INBOUND {
		ms = lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Margin(0, 1, 0, 0).Width(m.inboundColumnWidth).Render(m.inboundList.View()),
			lipgloss.NewStyle().Margin(2, 1, 0, 0).Width(m.previewWidth).Render(myStyle.previewInactiveColorStyle.Render(m.preview.View())),
		)

	} else {
		ms = lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Margin(0, 1, 0, 0).Width(m.directoryColumnWidth).Render(m.directoryList.View()),
			lipgloss.NewStyle().Margin(1, 0, 0, 0).Width(m.directoryFilesColumnWidth).Render(m.directoryFileList.View()),
		)
	}

	return ms
}

func (m model) selectedFile() string {
	return lipgloss.NewStyle().Margin(1, 0, 0, 0).Padding(0, 0).Render("  " +
		myStyle.titleStyle.Render("Selected File") + "  " +
		m.inboundList.SelectedItem().(inboundItem).Title())
}

func (m model) newNameSection() string {
	return lipgloss.NewStyle().Margin(0, 0, 0, 0).Padding(0, 0).Render("  " +
		m.newNameHeaderStyle.Render("New Name") + "       " +
		myStyle.textDimmedStyle.Render(m.timeStamp) +
		m.newNameInput.View())
}

func (m model) statusBar() string {
	statusWidth := m.width - 2

	statusText := m.statusMessage
	if len(statusText) > statusWidth {
		statusText = statusText[:statusWidth-3] + "..."
	}

	return lipgloss.NewStyle().Margin(1, 1, 0, 2).Render(
		myStyle.titleStyle.Render("$ ") +
			myStyle.statusBarStyle.Copy().Width(m.width-6).Render(statusText))
}

func (m model) helpView() string {
	return lipgloss.NewStyle().Margin(1, 0, 0, 2).Render(m.help.FullHelpView(keys.FullHelp()))
}
