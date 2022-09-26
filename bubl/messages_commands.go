package bubl

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmnpl/ding/core"
)

// -----------------------------------------------------------------------------
// messages
type moveMsg struct {
	messageText string
	err         error
}

type ocrMessageMulti struct {
	message string
	err     error
}

type ocrMessageSingle struct {
	message string
	err     error
}

// -----------------------------------------------------------------------------
// commands
func makeMoveCommand(fileName, newName, directoryName string) func() tea.Msg {
	return func() tea.Msg {
		neeewName, err := core.MoveFileToDirectory(fileName, newName, directoryName)

		messageWaht := fmt.Sprintf("\"%v\" to \"%v\"", fileName, neeewName)
		message := "Moved " + messageWaht
		if err != nil {
			message = "Failed to move " + messageWaht
		}
		return moveMsg{
			messageText: message,
			err:         err,
		}
	}
}

func (i inboundItem) makeOcrCommand(single bool) func() tea.Msg {
	if single {
		return func() tea.Msg {
			err := core.OcrPdf(i.name)
			return ocrMessageSingle{message: "success", err: err}
		}
	}

	return func() tea.Msg {
		err := core.OcrPdf(i.name)
		return ocrMessageMulti{message: "success", err: err}
	}
}
