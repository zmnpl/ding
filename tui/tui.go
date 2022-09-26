package tui

import (
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/zmnpl/ding/core"
)

// TODO
// finish layout
// add status line
// - showing summary of current action
// - when started show progress or status message
// add minimal keymap (somewhere)

var (
	titleColorString       = "[blue]"
	titleColor             = tcell.ColorBlue
	subtileColorString     = "[purple]"
	subtileColor           = tcell.ColorPurple
	deactivatedColorString = "[grey]"
	deactivatedColor       = tcell.ColorGrey

	borderColor      = tcell.ColorGrey
	borderTitleColor = tcell.ColorGrey

	app *tview.Application

	contextKeyMap             *tview.TextView
	mainfunctionKeyMap        *tview.TextView
	keymapTemplate            = titleColorString + "%s " + subtileColorString + "%s"
	deactivatedKeymapTemplate = deactivatedColorString + "%s %s"
	keymapSep                 = "[white] • "

	documentView   *tview.TextView
	fileListHeader *tview.TextView
	fileList       *tview.List

	directoryList       *tview.List
	directoryListHeader *tview.TextView
	directoryFileList   *tview.List

	newNameFlex                   *tview.Flex
	newNamePrefixInput            *tview.InputField
	newNameInput                  *tview.InputField
	autocompleteSelectedDirectory func(pathText string) (entries []string)

	statusLine *tview.TextView
)

func reset() {
	app.SetFocus(fileList)

	newNameInput.SetText("")
	newNamePrefixInput.SetText("")

	newNameFlex.Clear()
	newNameFlex.AddItem(newNamePrefixInput, 0, 0, false).AddItem(newNameInput, 0, 1, false)
}

var autocompleteDirectoryMaker = func(path *tview.InputField, extensionFilter map[string]bool) func(pathText string) (entries []string) {
	var mutex sync.Mutex
	prefixMap := make(map[string][]string)
	firstStart := true

	return func(pathText string) (entries []string) {
		prefix := strings.TrimSpace(strings.ToLower(pathText))
		if prefix == "" {
			return nil
		}

		mutex.Lock()
		defer mutex.Unlock()
		// prevent autocomplete to be shown when the options panel is drawn initially
		if firstStart {
			firstStart = false
			return nil
		}
		entries, ok := prefixMap[prefix]
		if ok {
			return entries
		}

		go func() {
			selectedDirectory, _ := directoryList.GetItemText(directoryList.GetCurrentItem())
			files, err := os.ReadDir(filepath.Join(core.Dest, selectedDirectory))

			if err != nil {
				return
			}

			entries := make([]string, 0, len(files))
			dupesBlock := make(map[string]bool)
			for _, file := range files {
				// don't show hidden folders
				if strings.HasPrefix(file.Name(), ".") {
					continue
				}
				// or directories
				if file.IsDir() {
					continue
				}
				// filter specific extensions
				if len(extensionFilter) > 0 {
					_, extensionOk := extensionFilter[strings.ToLower(filepath.Ext(file.Name()))]
					if !extensionOk {
						continue
					}
				}

				fileNameWithoutTimestamp := core.RemoveTimeStampFilePrefix(file.Name())
				if strings.HasPrefix(strings.TrimSpace(strings.ToLower(fileNameWithoutTimestamp)), prefix) {
					if _, in := dupesBlock[fileNameWithoutTimestamp]; !in {
						entries = append(entries, fileNameWithoutTimestamp)
						dupesBlock[fileNameWithoutTimestamp] = true
					}
				}
			}

			mutex.Lock()
			prefixMap[prefix] = entries
			mutex.Unlock()

			path.Autocomplete()

			app.Draw()

		}()
		return nil
	}
}

func Start() {
	contextKeyMap = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).SetWrap(true).SetTextAlign(tview.AlignCenter)

	mainfunctionKeyMap = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).SetWrap(true).SetTextAlign(tview.AlignCenter)

	mft := fmt.Sprintf(keymapTemplate, "ctrl+c", "exit") +
		keymapSep +
		fmt.Sprintf(keymapTemplate, "f5", "ocr inbound")
	mainfunctionKeyMap.SetText(mft)

	documentView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)
	documentView.SetBorderPadding(0, 1, 1, 1)

	fileListHeader = tview.NewTextView().SetDynamicColors(true).SetText("inbound")
	fileListHeader.SetBackgroundColor(subtileColor)
	fileList = tview.NewList()
	directoryListHeader = tview.NewTextView().SetDynamicColors(true).SetText("directories")
	directoryListHeader.SetBackgroundColor(subtileColor)
	directoryList = tview.NewList()

	newNamePrefixInput = tview.NewInputField()
	newNameInput = tview.NewInputField().SetPlaceholder("type new name")
	directoryFileList = tview.NewList().SetSelectedFocusOnly(true)

	statusLine = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)

	setupInboundFileList()
	setupDirectoryList()
	setupNewName()

	// set up main grid layout
	layout := tview.NewGrid()
	layout.SetRows(2, 2, -1, 1, 1, 1)
	layout.SetColumns(45, -1, -1, 45)
	layout.SetBorderPadding(1, 1, 1, 1)

	// main functions
	layout.AddItem(mainfunctionKeyMap, 0, 0, 1, 4, 0, 0, false)

	// left side; inbound
	infoPanelThreshold := 160
	fh := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(fileListHeader, 7, 0, false).
			AddItem(tview.NewTextView(), 0, 1, false), 1, 0, false).
		AddItem(tview.NewTextView(), 0, 1, false)

	layout.AddItem(fh, 1, 0, 1, 2, 0, 0, false)
	layout.AddItem(fileList, 2, 0, 1, 2, 0, 0, true)
	layout.AddItem(documentView, 2, 0, 1, 1, 0, infoPanelThreshold, false)
	layout.AddItem(fh, 1, 1, 1, 1, 0, infoPanelThreshold, false)
	layout.AddItem(fileList, 2, 1, 1, 1, 0, infoPanelThreshold, true)
	// fill empty cells with dummy to draw background
	layout.AddItem(tview.NewTextView(), 1, 0, 1, 1, 0, infoPanelThreshold, false)

	// reight side; destination
	bh := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(directoryListHeader, 7, 0, false).
			AddItem(tview.NewTextView(), 0, 1, false), 1, 0, false).
		AddItem(tview.NewTextView(), 0, 1, false)

	layout.AddItem(bh, 1, 2, 1, 2, 0, 0, false)
	layout.AddItem(directoryList, 2, 2, 1, 2, 0, 0, false)
	layout.AddItem(directoryFileList, 2, 3, 1, 1, 0, infoPanelThreshold, false)
	layout.AddItem(bh, 1, 2, 1, 2, 0, infoPanelThreshold, false)
	layout.AddItem(directoryList, 2, 2, 1, 1, 0, infoPanelThreshold, false)

	newNameFlex = tview.NewFlex().AddItem(newNamePrefixInput, 0, 0, false).AddItem(newNameInput, 0, 1, false)
	layout.AddItem(newNameFlex, 3, 0, 1, 4, 0, 0, false)
	layout.AddItem(newNameFlex, 3, 1, 1, 2, 0, infoPanelThreshold, false)
	// fill empty cells with dummy to draw background
	layout.AddItem(tview.NewTextView(), 3, 0, 1, 1, 0, infoPanelThreshold, false)
	layout.AddItem(tview.NewTextView(), 3, 3, 1, 1, 0, infoPanelThreshold, false)
	populateNewName()

	layout.AddItem(contextKeyMap, 4, 0, 1, 4, 0, 0, false)

	// bottom row
	layout.AddItem(statusLine, 5, 0, 1, 4, 0, 0, false)

	//layout.AddItem(dummy, 1, 0, 1, 1, 0, 0, false)
	//layout.AddItem(dummy, 1, 3, 1, 1, 0, 0, false)

	// general app setup
	app = tview.NewApplication()
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// if event.Key() == tcell.KeyF8 {
		// 	reset()
		// 	return nil
		// }
		if event.Key() == tcell.KeyF5 {
			runOcr()
			return nil
		}

		return event
	})
	app.SetRoot(layout, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func runOcr() {
	ocrStatusTemplate := titleColorString + "OCR: " +
		subtileColorString + "%v%%" +
		"[white] | " +
		titleColorString + "File: " +
		subtileColorString + "%s"

	progress := make(chan float32)
	currentfile := make(chan string)
	o, _ := core.GetOcrInboundFunc(progress, currentfile)
	go o()

	go func() {
		for p := range progress {
			app.QueueUpdateDraw(func() {
				statusLine.SetText(fmt.Sprintf(ocrStatusTemplate, int(p*100), <-currentfile))
			})
		}
	}()
}

func setupInboundFileList() {

	fileList.SetFocusFunc(func() {
		text := fmt.Sprintf(keymapTemplate, "🠕🠗", "navigate") +
			keymapSep +
			fmt.Sprintf(keymapTemplate, "enter", "select") +
			keymapSep +
			fmt.Sprintf(keymapTemplate, "f1", "open in external viewer")

		contextKeyMap.SetText(text)
	})

	fileList.Clear()

	inboundFiles, err := core.GetInboundFiles()
	if err != nil {
		fileList.AddItem("error", "see status line", 0, nil)
		statusLine.SetText(fmt.Sprintf("[red]could not read inbound file list: %s", err))
		return
	}

	for i, f := range inboundFiles {
		if i == 0 {
			populateDocPreview(core.GetCachedDocPreview(f.Name()))
		}
		inf, _ := f.Info()
		sizeMiBs := math.Round(float64(inf.Size())*100/1048576) / 100
		fileList.AddItem(f.Name(), fmt.Sprintf("%v MiB", sizeMiBs), 0, func() {
			app.SetFocus(directoryList)
		})
	}

	fileList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		populateDocPreview(core.GetCachedDocPreview(mainText))
	})

	fileList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		k := event.Key()
		if k == tcell.KeyF1 {
			filename, _ := fileList.GetItemText(fileList.GetCurrentItem())
			err := core.OpenDocExternal(filepath.Join(core.Inbound, filename))
			if err != nil {
				statusLine.SetText(fmt.Sprintf("[red]could not open file in default application: %s", err))
			}
		}

		return event
	})
}

func setupDirectoryList() {
	directoryList.SetFocusFunc(func() {
		text := fmt.Sprintf(keymapTemplate, "🠕🠗", "navigate") +
			keymapSep +
			fmt.Sprintf(keymapTemplate, "enter", "select directory") +
			keymapSep +
			fmt.Sprintf(keymapTemplate, "shift+tab", "back")
		contextKeyMap.SetText(text)
	})

	directories, err := core.GetDirectories()
	if err != nil {
		directoryList.AddItem("Could not get directories", fmt.Sprintf("%s", err), 0, nil)
		return
	}

	for _, directory := range directories {
		directoryName := directory.Name()

		directoryList.AddItem(directoryName, fmt.Sprintf("Files: %v", core.CountDirectory(directoryName)), 0, func() {
			populateNewName()
			app.SetFocus(newNameInput)
		})

		directoryList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
			directoryFiles, err := core.GetCachedDirectoryFiles(mainText)
			if err != nil {
				// TODO
			}
			populateDirectoryFileList(directoryFiles)
		})
	}

	directoryList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		k := event.Key()
		if k == tcell.KeyBacktab {
			app.SetFocus(fileList)
			return nil
		}

		return event
	})

	// init directory files
	directoryName, _ := directoryList.GetItemText(directoryList.GetCurrentItem())
	directoryFiles, _ := core.GetCachedDirectoryFiles(directoryName)
	populateDirectoryFileList(directoryFiles)
}

func updateSelectedDirectory() {
	index := directoryList.GetCurrentItem()
	directoryName, _ := directoryList.GetItemText(index)
	directoryList.SetItemText(index, directoryName, fmt.Sprintf("Files: %v", core.CountDirectory(directoryName)))
}

func setupNewName() {
	newNameInput.SetFocusFunc(func() {
		keymap := fmt.Sprintf(keymapTemplate, "enter", "ingest file") +
			keymapSep +
			fmt.Sprintf(keymapTemplate, "f1", "original name") +
			keymapSep +
			fmt.Sprintf(keymapTemplate, "shift+tab", "back")

		contextKeyMap.SetText(keymap)
	})

	newNameInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			fileName, _ := fileList.GetItemText(fileList.GetCurrentItem())
			directoryName, _ := directoryList.GetItemText(directoryList.GetCurrentItem())
			newFileName := newNamePrefixInput.GetText() + newNameInput.GetText()

			statusLine.SetText(fmt.Sprintf("Wait a second, moving %s to %s", fileName, filepath.Join(core.Dest, directoryName, newFileName)))

			// actual move, blocking
			core.MoveFileToDirectory(fileName, newFileName, directoryName)
			// newly setup ui to reflect changes
			setupInboundFileList()
			//setupDirectoryList()
			updateSelectedDirectory()

			reset()

			statusLine.SetText(fmt.Sprintf("Moved "+titleColorString+"%s[white] to "+subtileColorString+"%s[white]", fileName, filepath.Join(core.Dest, directoryName, newFileName)))
		}
	})

	newNameInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		k := event.Key()
		if k == tcell.KeyF1 {
			fileName, _ := fileList.GetItemText(fileList.GetCurrentItem())
			newNameInput.SetText(fileName)
			return nil
		}
		if k == tcell.KeyBacktab {
			newNameInput.SetText("")
			//newNamePrefixInput.SetText("")
			app.SetFocus(directoryList)
		}

		return event
	})
}

func populateNewName() {
	prefix := core.GetTimestampFilePrefix()
	newNamePrefixInput.SetText(prefix)

	newNameFlex.Clear()
	newNameFlex.AddItem(newNamePrefixInput, len(prefix), 0, false).AddItem(newNameInput, 0, 1, false)

	autocompleteSelectedDirectory = autocompleteDirectoryMaker(newNameInput, map[string]bool{".pdf": true})
	newNameInput.SetAutocompleteFunc(autocompleteSelectedDirectory)
}

// display only controls
func populateDirectoryFileList(directoryFiles []fs.DirEntry) {
	directoryFileList.Clear()
	for _, file := range directoryFiles {
		inf, _ := file.Info()
		sizeMiBs := math.Round(float64(inf.Size())*100/1048576) / 100
		directoryFileList.AddItem(file.Name(), fmt.Sprintf("%v Mib", sizeMiBs), 0, nil)
	}
}

func populateDocPreview(text string) {
	documentView.Clear()
	fmt.Fprintf(documentView, "%s", text)
}

func hexStringFromColor(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("[#%02x%02x%02x]", r, g, b)
}
