package bubl

import (
	"github.com/charmbracelet/bubbles/list"
)

type myListItem interface {
	FilterValue() string
	Title() string
	Description() string
	RenderLength() int
}

func listMaxItemLength(items []list.Item) (max int) {
	if len(items) == 0 {
		return 25
	}
	for _, item := range items {
		length := item.(myListItem).RenderLength()
		if length > max {
			max = length
		}
	}
	return
}
