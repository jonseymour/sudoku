package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/jonseymour/sudoku/model"
	"os"
	"strings"
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
	var overflow string
	br := bufio.NewReader(os.Stdin)
	bw := bufio.NewWriter(os.Stdout)
	puzzles := 0

	for {
		grid := model.NewGrid()

		var buffer string

		buffer = overflow
		overflow = ""

		for len(buffer) < 81 {
			line, err := br.ReadString('\n')
			if err != nil {
				if len(buffer) == 0 {
					if solved {
						os.Exit(0)
					} else {
						os.Exit(1)
					}
				} else {
					fmt.Fprintf(os.Stderr, "read error: %v\n", err)
					os.Exit(2)
				}
			}
			line = strings.TrimSpace(line)
			buffer = buffer + line
		}
		puzzles += 1

		overflow = buffer[81:]
		buffer = buffer[0:81]
		for i, ch := range buffer {
			if ch == '.' || ch == '0' {
				continue
			}
			r := i / 9
			c := i % 9
			if ch >= '1' && ch <= '9' {
				value := int(ch - int32('1'))
				grid.Assert(model.CellIndex{Row: r, Column: c}, value, "initial state")
			} else {
				fmt.Fprintf(os.Stderr, "invalid cell value: %d: %v\n", puzzles, string(rune(ch)))
				os.Exit(2)
			}
		}

		var err error
		if solved, err = grid.Solve(); err != nil {
			fmt.Fprintf(os.Stderr, "invalid puzzle: %d: %v\n", puzzles, err)
		}

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
