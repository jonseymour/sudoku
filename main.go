package main

import (
	"flag"
	"fmt"
	gridio "github.com/jonseymour/sudoku/io"
	"github.com/jonseymour/sudoku/model"
	"io"
	"os"
)

const (
	VERSION = "1.2-pre"
)

func main() {
	var verbose = flag.Bool("verbose", false, "Set the verbosity of the logging")
	var version = flag.Bool("version", false, "Report the version number")
	flag.Parse()

	if *verbose {
		model.LogFile = os.Stderr
	}

	if *version {
		fmt.Fprintf(os.Stdout, "%s\n", VERSION)
		os.Exit(0)
	}

	var solved = false
	rdr := gridio.NewGridReader(os.Stdin)
	w := gridio.NewGridWriter(os.Stdout)

	var err error
	var grid *model.Grid

	for grid, err = rdr.Read(); grid != nil && err == nil; grid, err = rdr.Read() {

		if solved, err = grid.Solve(); err != nil {
			fmt.Fprintf(os.Stderr, "invalid puzzle: %d: %v\n", rdr.ReadCount(), err)
		}

		w.Write(grid)
		w.Flush()
	}

	if err != nil {
		if solved {
			os.Exit(0)
		} else {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "read error: %s\n", err)
			}
			os.Exit(1)
		}
	}
}
