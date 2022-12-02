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

	action := func() (string, error) {
		err := core.OcrPdf(i.name)
		message := "success"
		if err != nil {
			message = err.Error()
		}
		return message, err
	}

	if single {
		return func() tea.Msg {
			message, err := action()
			return ocrMessageSingle{message: message, err: err}
		}
	}

	return func() tea.Msg {
		message, err := action()
		return ocrMessageSingle{message: message, err: err}
	}
}
