package model

import (
	"fmt"
	"io"
	"io/ioutil"
)

type ValueState int
type GroupType int

var LogFile io.Writer = ioutil.Discard

const (
	MAYBE ValueState = 0 + iota
	NO
	YES
)

const (
	ROW GroupType = 0 + iota
	COLUMN
	BLOCK
)

type Group struct {
	Name   string
	Counts [9]int
	Cells  [9]*Cell
}

func (g *Group) String() string {
	return g.Name
}

type CellIndex struct {
	Row    int
	Column int
}

func (i CellIndex) BoardIndex() int {
	return i.Row*9 + i.Column
}

func (i CellIndex) RowGroup() int {
	return i.Row
}

func (i CellIndex) ColumnGroup() int {
	return 9 + i.Column
}

func (i CellIndex) BlockGroup() int {
	return 18 + (i.Row/3)*3 + (i.Column / 3)
}

func (i CellIndex) RowIndex() int {
	return i.Column
}

func (i CellIndex) ColumnIndex() int {
	return i.Row
}

func (i CellIndex) BlockIndex() int {
	return (i.Row%3)*3 + i.Column%3
}

func (i CellIndex) String() string {
	return fmt.Sprintf("(Row:%d, Column:%d, Block:%d)", i.RowGroup()+1, i.ColumnGroup()-8, i.BlockGroup()-17)
}

type Cell struct {
	BoardIndex  int
	Maybes      int
	Bits        int
	Value       *int
	ValueStates [9]ValueState
	Groups      [3]*Group
}

func (c *Cell) Index() CellIndex {
	return CellIndex{Row: c.BoardIndex / 9, Column: c.BoardIndex % 9}
}

type Board struct {
	queue  []func()
	Groups [27]*Group
	Cells  [81]*Cell
}

func NewBoard() *Board {
	board := &Board{}
	for x, _ := range board.Groups {
		g := &Group{}
		board.Groups[x] = g
		for i, _ := range g.Counts {
			g.Counts[i] = 9
			if x < 9 {
				g.Name = fmt.Sprintf("Row:%d", x+1)
			} else if x < 18 {
				g.Name = fmt.Sprintf("Column:%d", x-9+1)
			} else {
				g.Name = fmt.Sprintf("Block:%d", x-18+1)
			}
		}
	}
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			i := &CellIndex{Row: r, Column: c}

			cell := &Cell{}
			board.Cells[i.BoardIndex()] = cell
			cell.Bits = 1<<9 - 1
			cell.BoardIndex = i.BoardIndex()
			cell.Maybes = 9

			cell.Groups[ROW] = board.Groups[i.RowGroup()]
			cell.Groups[COLUMN] = board.Groups[i.ColumnGroup()]
			cell.Groups[BLOCK] = board.Groups[i.BlockGroup()]

			cell.Groups[ROW].Cells[i.RowIndex()] = cell
			cell.Groups[COLUMN].Cells[i.ColumnIndex()] = cell
			cell.Groups[BLOCK].Cells[i.BlockIndex()] = cell
		}
	}
	return board
}

func (b *Board) Assert(i CellIndex, value int, reason string) {
	fmt.Fprintf(LogFile, "assert: value=%d, cell=%s, reason=%s\n", value+1, i, reason)
	cell := b.Cells[i.BoardIndex()]
	cell.Bits = 1 << uint(value)
	if cell.Value != nil && *cell.Value != value {
		panic(fmt.Errorf("contradictory assertion: already asserted %d @ %s, now trying to assert %d", *cell.Value+1, i, value+1))
	}
	cell.Value = &value
	switch cell.ValueStates[value] {
	case MAYBE:
		cell.Maybes = 1

		// reduce the counts associated with the other values in
		// the intersecting groups

		for v, _ := range cell.ValueStates {
			if cell.ValueStates[v] == MAYBE && v != value {
				for _, g := range cell.Groups {
					g.Counts[v] -= 1
					if g.Counts[v] == 1 {
						b.processUnique(g, v)
					}
				}
			}
			cell.ValueStates[v] = NO
		}
		cell.ValueStates[value] = YES

		for _, g := range cell.Groups {
			for _, c := range g.Cells {
				if c.BoardIndex != i.BoardIndex() && c.ValueStates[value] == MAYBE {
					x := c.Index()
					b.queue = append(b.queue, func() {
						b.Reject(x, value, fmt.Sprintf("excluded by assertion of %d @ %s", value+1, i))
					})
				}
			}
		}

	case YES:
	case NO:
		panic(fmt.Errorf("contradiction: attempted to assert %d @ %v, but this value was previously rejected", value+1, i))
	}
}

func (b *Board) processUnique(g *Group, value int) {
	for _, c := range g.Cells {
		x := c.Index()
		cell := c
		if c.ValueStates[value] == MAYBE {
			b.queue = append(b.queue, func() {
				if cell.ValueStates[value] == MAYBE {
					b.Assert(x, value, fmt.Sprintf("unique value found in group %s", g))
				}
			})
		}
	}
}

func (b *Board) Reject(i CellIndex, value int, reason string) {
	fmt.Fprintf(LogFile, "reject: value=%d, cell=%s, reason=%s\n", value+1, i, reason)
	cell := b.Cells[i.BoardIndex()]
	switch cell.ValueStates[value] {
	case MAYBE:
		bit := 1 << uint(value)
		cell.ValueStates[value] = NO
		cell.Bits &^= bit
		cell.Maybes -= 1

		for _, g := range cell.Groups {
			g.Counts[value] -= 1
			if g.Counts[value] == 1 {
				b.processUnique(g, value)
			}
		}

		if cell.Maybes == 1 {
			// if a cell has only one maybe, then we assert that value
			// as the cell's value

			b.queue = append(b.queue, func() {
				for v, s := range cell.ValueStates {
					if s == MAYBE {
						x := cell.Index()
						b.queue = append(b.queue, func() {
							b.Assert(x, v, "only possible value in cell")
						})
						return
					}
				}

			})
		} else if cell.Maybes == 2 {

			// if a cell can only contain one value in a pair,
			// then check if there is another cell restricted
			// to the same pair in the same group
			//
			// if so, reject the pair values from every other cell
			// in the same group
			//

			b.queue = append(b.queue, func() {
				if cell.Maybes == 2 {
					pair := []int{}
					for v, s := range cell.ValueStates {
						if s == MAYBE {
							pair = append(pair, v)
						}
					}
					if len(pair) != 2 {
						panic("expected: len(pair) == 2")
					}
					for _, g := range cell.Groups {
						for _, c := range g.Cells {
							if c.BoardIndex != cell.BoardIndex && c.Bits == cell.Bits {
								// found a matching pair in the same group
								for _, c1 := range g.Cells {
									x := c1.Index()
									if c1.BoardIndex != c.BoardIndex && c1.BoardIndex != cell.BoardIndex {
										b.queue = append(b.queue, func() {
											if c1.ValueStates[pair[0]] == MAYBE {
												b.Reject(x, pair[0], fmt.Sprintf("excluded by pair (%v,%v) @ %s, %s", pair[0], pair[1], cell.Index(), c.Index()))
											}
											if c1.ValueStates[pair[1]] == MAYBE {
												b.Reject(x, pair[1], fmt.Sprintf("excluded by pair (%v,%v) @ %s, %s", pair[0], pair[1], cell.Index(), c.Index()))
											}
										})
									}
								}
							}
						}
					}
				}
			})
		} else if cell.Maybes == 3 {
			b.queue = append(b.queue, func() {
				fmt.Fprintf(LogFile, "info: triple found at %s - %03x\n", cell.Index(), cell.Bits)
			})
		}
	case YES:
		if value == *cell.Value {
			panic(fmt.Errorf("contradiction: attempt to reject value=%d @ %s, but this value was previously asserted", value+1, i))
		}
	case NO:
	}
}

func (b *Board) Solve() bool {
	for len(b.queue) > 0 {
		next := b.queue[0]
		b.queue = b.queue[1:]
		next()
	}
	for _, c := range b.Cells {
		if c.Value == nil {
			return false
		}
	}
	return true
}
