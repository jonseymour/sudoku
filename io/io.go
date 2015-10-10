package io

import (
	"bufio"
	"fmt"
	"github.com/jonseymour/sudoku/model"
	io "io"
	"strings"
)

type GridReader interface {
	Read() (*model.Grid, error)
	ReadCount() int
}

type gridReader struct {
	buffered *bufio.Reader
	overflow string
	puzzles  int
}

func NewGridReader(r io.Reader) GridReader {
	gr := &gridReader{
		buffered: bufio.NewReader(r),
	}
	return gr
}

func (gr *gridReader) ReadCount() int {
	return gr.puzzles
}

func (gr *gridReader) Read() (*model.Grid, error) {

nextline:
	for {
		var buffer string

		buffer = gr.overflow
		gr.overflow = ""

		for len(buffer) < 81 {
			line, err := gr.buffered.ReadString('\n')
			if err != nil {
				if len(buffer) == 0 {
					return nil, err
				} else {
					return nil, fmt.Errorf("truncated input")
				}
			}
			line = strings.TrimSpace(line)
			if len(line) == 0 || strings.HasPrefix(line, "#") {
				continue nextline
			}
			buffer = buffer + line
		}

		gr.puzzles++

		gr.overflow = buffer[81:]
		buffer = buffer[0:81]
		grid := model.NewGrid()
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
				return nil, fmt.Errorf("invalid cell value: %d: %v\n", gr.puzzles, string(rune(ch)))
			}
		}

		return grid, nil
	}
}

type GridWriter interface {
	Write(g *model.Grid) error
	Flush() error
}

type gridWriter struct {
	buffered *bufio.Writer
	format   string
}

func NewGridWriter(w io.Writer, format string) (GridWriter, error) {
	if format == "" {
		format = "9."
	}
	if len(format) != 2 ||
		(format[0] != '9' && format[0] != '1') ||
		(format[1] != '0' && format[1] != '.') {
		return nil, fmt.Errorf("format must be one of: 9., 90, 1., 10")
	}
	gw := &gridWriter{
		buffered: bufio.NewWriter(w),
		format:   format,
	}
	return gw, nil
}

func (gw *gridWriter) Write(g *model.Grid) error {
	var err error
	wrapModulus := 81 / int(gw.format[0]-'0')
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			x := r*9 + c
			v := g.Cells[x].Value
			if v == nil {
				_, err = gw.buffered.WriteRune(rune(gw.format[1]))
			} else {
				_, err = gw.buffered.WriteRune(rune(int32(*v) + int32('1')))
			}
			if err != nil {
				return err
			}
		}
		if (r+1)*9%wrapModulus == 0 {
			if _, err = gw.buffered.WriteString("\n"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (gw *gridWriter) Flush() error {
	return gw.buffered.Flush()
}
