package core

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

const ()

var (
	DEPENDENCIES = map[string]string{
		"pdftotext": "display textual preview of pdf",
		"xdg-open":  "open pdf in your default viewer",
		"ocrmypdf":  "run ocr on pdf",
		"img2pdf":   "convert image to pdf",
		"ag":        "list your documents very fast",
		"fzf":       "fuzzy search through your documents",
		"rga":       "ripgrep-all - use in combination with fzf to fuzzy search your documents",
	}

	redPrinter   = color.New(color.FgRed).SprintFunc()
	greenPrinter = color.New(color.FgGreen).SprintFunc()
)

func checkDep(dep string) bool {
	if _, err := exec.LookPath(dep); err == nil {
		return true
	}
	return false
}

// functions which run external commands

// PrintCheckDeps runs a check for all external dependecies and prints out the result
func PrintCheckDeps() {
	fmt.Printf("Running depency check for 'ding'...\n")
	for dep, depFunction := range DEPENDENCIES {
		if checkDep(dep) {
			fmt.Printf("%s: %-10s - You can %s.\n", greenPrinter("FOUND"), dep, depFunction)
			continue
		}
		fmt.Printf("%s: %-10s - You won't be able to %s.\n", redPrinter("MISSING"), dep, depFunction)
	}
}

// MissingExternDependencies returns a slice with the names of executables which are missing on the
// system to use all the functionality
func CheckDependencies() (available, missing []string) {
	available = make([]string, 0)
	missing = make([]string, 0)
	for _, dep := range DEPENDENCIES {
		if checkDep(dep) {
			available = append(available, dep)
			continue
		}
		missing = append(missing, dep)
	}
	return available, missing
}

// OpenDocExternal tries to open the document in the systems default application for given type
func OpenDocExternal(path string) error {
	cmd := exec.Command("xdg-open", path)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("could not open document in external preferred application: %v", err)
	}
	return nil
}

// GetDocPreview returns the text layer of a pdf as simple string by running the external commant
// pdftotext on it
func GetDocPreview(name string) string {
	//cmd := exec.Command("pdftotext", "-layout", "-f", "1", "-l", "1", filepath.Join(Inbound, name), "-")
	cmd := exec.Command("pdftotext", "-f", "1", "-l", "1", filepath.Join(Inbound, name), "-")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("could not get preview:\n\n%s\n\n%v", string(output), err)
	}
	out := strings.TrimSpace(string(output))
	if out == "" {
		out = "- no OCR content -"
	}
	return string(out)
}

// GetOcrInboundFunc returns a function that can be run async to iterate all inbound files
// and add a text layer to scans
// progress can be watched by the channels that are passed here
func GetOcrInboundFunc(progress chan float32, currentFile chan string) (runner func(), err error) {
	files, err := GetInboundFiles()
	fileCnt := len(files)

	if err == nil {
		return func() {
			defer close(progress)
			defer close(currentFile)

			for i, f := range files {
				progress <- float32(i) / float32(fileCnt)
				currentFile <- f.Name()
				OcrPdf(f.Name())
			}
			progress <- 1.0
			currentFile <- ""
		}, nil
	}

	return nil, err
}

// OcrPdf runs the external command ocrmypdf and so tries to add a text layer to scans
func OcrPdf(name string) error {
	f := filepath.Join(Inbound, name)
	cmd := exec.Command("ocrmypdf", "-q", "-l", "deu", f, f)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not ocr your pdf:\n\n%v", err)
	}

	go UpdateFilePreviewCache(name)

	return nil
}
