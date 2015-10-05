package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/jonseymour/sudoku/model"
	"os"
)

const (
	VERSION = "1.1-pre"
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
	br := bufio.NewReader(os.Stdin)
	bw := bufio.NewWriter(os.Stdout)
	for {
		grid := model.NewGrid()
		for r := 0; r < 9; r++ {
			if row, err := br.ReadString('\n'); err != nil {
				if r != 0 {
					fmt.Fprintf(os.Stderr, "read error: %v", err)
					os.Exit(2)
				} else {
					if solved {
						os.Exit(0)
					} else {
						os.Exit(1)
					}
				}
			} else {
				if len(row) < 9 {
					fmt.Fprintf(os.Stderr, "invalid row\n")
				} else {
					row = row[0:9]
					for c, v := range row {
						if v == '.' {
							continue
						}
						if v >= '1' && v <= '9' {
							value := int(v - int32('1'))
							grid.Assert(model.CellIndex{Row: r, Column: c}, value, "problem initialization")
						}
					}
				}
			}
		}
		solved = grid.Solve()
		for r := 0; r < 9; r++ {
			for c := 0; c < 9; c++ {
				x := r*9 + c
				v := grid.Cells[x].Value
				if v == nil {
					bw.WriteString(".")
				} else {
					bw.WriteRune(rune(int32(*v) + int32('1')))
				}
			}
			bw.WriteString("\n")
			bw.Flush()
		}
	}
}
