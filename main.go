package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/zmnpl/ding/bubl"
	"github.com/zmnpl/ding/core"
	"github.com/zmnpl/ding/tui"
)

func flagUsage() {
	fmt.Printf(`Usage: fast-p [OPTIONS]
Reads a list of PDF filenames from STDIN and returns a list of null-byte
separated items of the form filename[TAB]text
where "text" is the text extracted from the first two pages of the PDF
by pdftotext and [TAB] denotes a tab character "\t".

Common usage of this tool is to pipe the result to FZF with a command in
your .bashrc as explained in https://github.com/bellecp/fast-p.`)

	flag.PrintDefaults()
}

func main() {
	checkDeps := flag.Bool("checkDependencies", false, "Check external depencies for different programm functions")
	tview := flag.Bool("ui", false, "Run with tview UI")
	out := flag.String("out", core.Dest, "Root path of your documents directory; where the documents should go")
	in := flag.String("in", core.Inbound, "Path where your scans / inbound documents land")

	//flag.Usage = flagUsage
	flag.Parse()

	if _, err := os.Stat(*out); !os.IsNotExist(err) {
		if *out != core.Dest {
			core.Dest = *out
		}
	} else {
		log.Fatal("The given documentPath does not exist")
	}

	if _, err := os.Stat(*in); !os.IsNotExist(err) {
		if *in != core.Inbound {
			core.Inbound = *in
		}
	} else {
		log.Fatal("The given inboundPath does not exist")
	}

	if *checkDeps {
		core.PrintCheckDeps()
		os.Exit(0)
	}

	if *tview {
		tui.Start()
		os.Exit(0)
	}

	bubl.Run()
}
