package bubl

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zmnpl/ding/core"
)

const (
	// TERMINAL THEME COLORS
	BLACK   = lipgloss.Color("0")
	RED     = lipgloss.Color("1")
	GREEN   = lipgloss.Color("2")
	YELLOW  = lipgloss.Color("3")
	BLUE    = lipgloss.Color("4")
	MAGENTA = lipgloss.Color("5")
	CYAN    = lipgloss.Color("6")
	WHITE   = lipgloss.Color("7")

	LIGHT_BLACK   = lipgloss.Color("8")
	LIGHT_RED     = lipgloss.Color("9")
	LIGTH_GREEN   = lipgloss.Color("10")
	LIGHT_YELLOW  = lipgloss.Color("11")
	LIGHT_BLUE    = lipgloss.Color("12")
	LIGHT_MAGENTA = lipgloss.Color("13")
	LIGHT_CYAN    = lipgloss.Color("14")
	LIGHT_WITE    = lipgloss.Color("15")

	OUTER_MARGIN = 1
)

var (
	myStyle appStyle
)

type appStyle struct {
	COLOR_DIMMED_TEXT        lipgloss.TerminalColor
	COLOR_ACTIVE_ITEM        lipgloss.TerminalColor
	COLOR_ACTIVE_ITEM_BORDER lipgloss.TerminalColor
	COLOR_TITLE              lipgloss.TerminalColor

	COLOR_INACTIVE_ITEM        lipgloss.TerminalColor
	COLOR_STATUSBAR_FOREGROUND lipgloss.TerminalColor

	docStyle       lipgloss.Style
	highlightStyle lipgloss.Style

	previewInactiveColorStyle lipgloss.Style
	timeStampStyle            lipgloss.Style

	styleActiveText   lipgloss.Style
	styleInactiveText lipgloss.Style

	titleStyle         lipgloss.Style
	titleStyleSelected lipgloss.Style
	itemStyle          lipgloss.Style
	itemStyleSelected  lipgloss.Style
	paginationStyle    lipgloss.Style
	statusBarStyle     lipgloss.Style
	textDimmedStyle    lipgloss.Style
}

func init() {
	myStyle = defaultStyle()
}

func defaultStyle() (s appStyle) {
	s.COLOR_DIMMED_TEXT = lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"} // some grey

	s.COLOR_ACTIVE_ITEM = lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}        // pink
	s.COLOR_ACTIVE_ITEM_BORDER = lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"} // dimmed pink

	s.COLOR_TITLE = lipgloss.Color("#3282FF")

	//s.COLOR_ACTIVE_ITEM =
	s.docStyle = lipgloss.NewStyle().Margin(OUTER_MARGIN, 0)

	// staring spinner only ... never used again
	s.highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// preview
	s.previewInactiveColorStyle = lipgloss.NewStyle()
	s.timeStampStyle = lipgloss.NewStyle()

	// text styles
	s.styleActiveText = lipgloss.NewStyle()
	s.styleInactiveText = lipgloss.NewStyle()

	s.titleStyle = lipgloss.NewStyle().Foreground(s.COLOR_TITLE)
	s.titleStyleSelected = s.titleStyle.Copy().Underline(true)

	s.itemStyle = lipgloss.NewStyle().PaddingLeft(2)
	s.itemStyleSelected = lipgloss.NewStyle().Foreground(s.COLOR_ACTIVE_ITEM)

	s.paginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(2).Margin(-1)

	s.statusBarStyle = lipgloss.NewStyle().Foreground(s.COLOR_DIMMED_TEXT)

	s.textDimmedStyle = lipgloss.NewStyle().Foreground(s.COLOR_DIMMED_TEXT)
	return
}

// -----------------------------------------------------------------------------
// list delegates

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(list.DefaultItem)
	if !ok {
		return
	}

	str := i.Title()

	fn := myStyle.itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return myStyle.itemStyleSelected.Render("> " + s)
		}
	}

	fmt.Fprintf(w, fn(str)+" "+myStyle.textDimmedStyle.Render(i.Description()))
}

type dimmedDelegate struct{}

func (d dimmedDelegate) Height() int                               { return 1 }
func (d dimmedDelegate) Spacing() int                              { return 0 }
func (d dimmedDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d dimmedDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(list.DefaultItem)
	if !ok {
		return
	}

	str := i.Title()
	if core.TimestampPrefixMatch.Match([]byte(i.Title())) {
		str = fmt.Sprintf("...%s", core.RemoveTimeStampFilePrefix(i.Title()))
	}

	if len(str) > 33 {
		str = fmt.Sprintf("%s...", str[:30])
	}

	fmt.Fprintf(w, myStyle.textDimmedStyle.Render(myStyle.itemStyle.Render(str)))
}
