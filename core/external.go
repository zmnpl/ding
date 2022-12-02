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

var (
	OCRMYPDF_ERRCODES = map[int]string{
		0:   "Everything worked as expected.",
		1:   "Invalid arguments, exited with an error.",
		2:   "The input file does not seem to be a valid PDF.",
		3:   "An external program required by OCRmyPDF is missing.",
		4:   "An output file was created, but it does not seem to be a valid PDF. The file will be available.",
		5:   "The user running OCRmyPDF does not have sufficient permissions to read the input file and write the output file.",
		6:   "The file already appears to contain text so it may not need OCR. See output message.",
		7:   "An error occurred in an external program (child process) and OCRmyPDF cannot continue.",
		8:   "The input PDF is encrypted. OCRmyPDF does not read encrypted PDFs. Use another program such as qpdf to remove encryption.",
		9:   "A custom configuration file was forwarded to Tesseract using --tesseract-config, and Tesseract rejected this file.",
		10:  "A valid PDF was created, PDF/A conversion failed. The file will be available.",
		15:  "Some other error occurred.",
		130: "The program was interrupted by pressing Ctrl+C.",
	}
)

// OcrPdf runs the external command ocrmypdf and so tries to add a text layer to scans
func OcrPdf(name string) error {
	f := filepath.Join(Inbound, name)
	cmd := exec.Command("ocrmypdf", "-q", "-l", "deu", "--redo-ocr", f, f)

	//var outb, errb bytes.Buffer
	//cmd.Stdout = &outb
	//cmd.Stderr = &errb
	//fmt.Println("out:", outb.String(), "err:", errb.String())

	_, err := cmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("ocr error: %v", OCRMYPDF_ERRCODES[exitError.ExitCode()])
		}
		return fmt.Errorf("ocr error: %v", err)
	}

	go UpdateFilePreviewCache(name)

	return nil
}
