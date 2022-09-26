package main

import (
	"flag"
	"log"
	"os"

	"github.com/zmnpl/ding/bubl"
	"github.com/zmnpl/ding/core"
	"github.com/zmnpl/ding/tui"
)

func main() {
	checkDeps := flag.Bool("checkDependencies", false, "Check external depencies for different programm functions")
	tview := flag.Bool("ui", false, "Run with tview UI")
	out := flag.String("out", core.Dest, "Root path of your documents directory; where the documents should go")
	in := flag.String("in", core.Inbound, "Path where your scans / inbound documents land")

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
