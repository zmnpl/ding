package bubl

import (
	"fmt"
	"io/fs"
	"math"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/zmnpl/ding/core"
)

func (m model) docPreview() string {
	if m.selectedInbound != nil && m.selectedInbound.(inboundItem).file != nil {
		return core.GetCachedDocPreview(m.selectedInbound.(inboundItem).file.Name())
	}
	return "-"
}

func (m model) updateDirectoryFiles() model {
	var directoryFiles []list.Item
	if m.selectedDirectory != nil {
		directoryFiles = m.selectedDirectory.(directory).GetDirectoryFiles()
	}
	m.directoryFileList.SetItems(directoryFiles)
	return m
}

// -----------------------------------------------------------------------------
// directory file
type directoryFile struct {
	name string
	size int64
	file fs.DirEntry
}

func NewDirectoryFile(file fs.DirEntry) directoryFile {
	info, _ := file.Info()
	return directoryFile{
		name: info.Name(),
		size: info.Size(),
		file: file,
	}
}

func (bf directoryFile) Title() string {
	return core.RemoveTimeStampFilePrefix(bf.name)
}

func (bf directoryFile) Description() string {
	sizeMiBs := math.Round(float64(bf.size)*100/1048576) / 100
	return fmt.Sprintf("(%v Mib)", sizeMiBs)
}

func (bf directoryFile) FilterValue() string { return bf.name }

func (bf directoryFile) RenderLength() int {
	return len(bf.Title()) + len(bf.Description()) + 3 // "> " + " "
}

func (b directory) GetDirectoryFiles() []list.Item {
	var files []fs.DirEntry
	if b.dir != nil {
		files, _ = core.GetCachedDirectoryFiles(b.dir.Name())
	} else {
		files, _ = core.GetCachedDirectoryFiles("-")
	}

	result := make([]list.Item, len(files))
	for i, file := range files {
		result[i] = NewDirectoryFile(file)
	}

	return result
}

// -----------------------------------------------------------------------------
// inbound item
type inboundItem struct {
	name string
	size int64
	file fs.DirEntry
}

func NewInboundItem(file fs.DirEntry) inboundItem {
	info, _ := file.Info()
	return inboundItem{
		name: file.Name(),
		size: info.Size(),
		file: file,
	}
}

func (i inboundItem) Title() string {
	return i.name
}

func (i inboundItem) Description() string {
	sizeMiBs := math.Round(float64(i.size)*100/1048576) / 100
	return fmt.Sprintf("(%v Mib)", sizeMiBs)
}

func (i inboundItem) FilterValue() string {
	// file name + preview to enable filtering by doc content
	return i.name // + " " +core.GetCachedDocPreview(i.file.Name())
}

// RenderLength gives the length of the rendered item text
// TODO: This is not in any way connected to the delegate render method ... how can this be done?
func (i inboundItem) RenderLength() int {
	return len(i.Title()) + len(i.Description()) + 3
}

func InboundItemsAsBubblesList() []list.Item {
	inboundFiles, _ := core.GetInboundFiles()
	items := make([]list.Item, len(inboundFiles))
	for i, file := range inboundFiles {
		items[i] = NewInboundItem(file)
	}

	return items
}

// -----------------------------------------------------------------------------
// directory
type directory struct {
	name string
	dir  fs.DirEntry
}

func NewDirectory(dir fs.DirEntry) directory {
	return directory{
		name: dir.Name(),
		dir:  dir,
	}
}

func (b directory) Title() string {
	return b.name
}

func (b directory) Description() string {
	return fmt.Sprintf("(%v files)", core.CountDirectory(b.name))
}

func (b directory) FilterValue() string { return b.dir.Name() }

func (b directory) RenderLength() int {
	return len(b.Title()) + len(b.Description()) + 3
}

func DirectoriesAsBubblesList() []list.Item {
	directories, _ := core.GetDirectories()
	items := make([]list.Item, len(directories))
	for i, dir := range directories {
		items[i] = NewDirectory(dir)
	}

	return items
}

// -----------------------------------------------------------------------------
// help

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type keyMap struct {
	Up          key.Binding
	Down        key.Binding
	Filter      key.Binding
	Confirm     key.Binding
	OpenPreview key.Binding
	OcrSingle   key.Binding
	OcrMultiple key.Binding
	Quit        key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Confirm, k.Quit},       // first column
		{k.Up, k.Down},            // second column
		{k.OpenPreview, k.Filter}, //...
		{k.OcrSingle, k.OcrMultiple},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	OpenPreview: key.NewBinding(
		key.WithKeys("f1"),
		key.WithHelp("f1", "open selected doc in pdf viewer"),
	),
	OcrSingle: key.NewBinding(
		key.WithKeys("f2"),
		key.WithHelp("f2", "ocr selected"),
	),
	OcrMultiple: key.NewBinding(
		key.WithKeys("f3"),
		key.WithHelp("f3", "ocr all"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
