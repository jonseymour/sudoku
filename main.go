package main

import (
	"flag"
	"fmt"
	gridio "github.com/jonseymour/sudoku/io"
	"github.com/jonseymour/sudoku/model"
	"io"
	"os"
	"runtime/pprof"
)

const (
	VERSION = "1.2.1-pre"
)

func main() {
	var verbose = flag.Bool("verbose", false, "Set the verbosity of the logging")
	var version = flag.Bool("version", false, "Report the version number")
	var format = flag.String("format", "9.", "Output format. One of: 9., 90, 1., 10")
	var cpuprofile = flag.Bool("cpuprofile", false, "Enable CPU profiling")
	var noverify = flag.Bool("no-verify-uniqueness", false, "Disable uniqueness check")
	var nosolve = flag.Bool("no-solve", false, "Disable solver - reformatting only.")

	flag.BoolVar(&model.ColoringDisabled, "no-coloring", false, "Disable coloring.")
	flag.BoolVar(&model.NoBacktracking, "no-backtracking", false, "Disable backtracking.")

	flag.Parse()

	if *verbose {
		model.LogFile = os.Stderr
		model.Verbose = *verbose
	}

	if *version {
		fmt.Fprintf(os.Stdout, "%s\n", VERSION)
		os.Exit(0)
	}

	model.VerifyUniqueness = !*noverify

	var err error

	var f io.WriteCloser
	if *cpuprofile {
		if f, err = os.Create("sudoku.pprof"); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		} else {
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
			defer f.Close()
		}
	}

	var solved = false
	var w gridio.GridWriter
	rdr := gridio.NewGridReader(os.Stdin)
	w, err = gridio.NewGridWriter(os.Stdout, *format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}

	var grid *model.Grid

	for grid, err = rdr.Read(); grid != nil && err == nil; grid, err = rdr.Read() {

		orig := grid.Clone()

		if !*nosolve {
			if solved, err = grid.Solve(); err != nil {
				fmt.Fprintf(os.Stderr, "invalid puzzle: %d: %v\n", rdr.ReadCount(), err)
			} else if !solved {
				fmt.Fprintf(os.Stderr, "failed to solve: %s\n", orig)
			}
		}

		w.Write(grid)
		w.Flush()
	}

	pprof.StopCPUProfile()
	if f != nil {
		f.Close()
	}

	if err == nil || err == io.EOF {
		if solved {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
