package model

import (
	"fmt"
)

type Priority int

const (
	IMMEDIATE Priority = iota
	DEFERRED
)

type Grid struct {
	Groups [27]*Group
	Cells  [81]*Cell

	clues int
	queue [2][]func()
}

func NewGrid() *Grid {
	grid := &Grid{
		queue: [2][]func(){
			[]func(){},
			[]func(){},
		},
	}
	for x, _ := range grid.Groups {
		g := &Group{}
		grid.Groups[x] = g
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
			grid.Cells[i.GridIndex()] = cell
			cell.Bits = 1<<9 - 1
			cell.GridIndex = i.GridIndex()
			cell.Maybes = 9

			cell.Groups[ROW] = grid.Groups[i.RowGroup()]
			cell.Groups[COLUMN] = grid.Groups[i.ColumnGroup()]
			cell.Groups[BLOCK] = grid.Groups[i.BlockGroup()]

			cell.Groups[ROW].Cells[i.RowIndex()] = cell
			cell.Groups[COLUMN].Cells[i.ColumnIndex()] = cell
			cell.Groups[BLOCK].Cells[i.BlockIndex()] = cell
		}
	}
	return grid
}

func (gd *Grid) Enqueue(p Priority, f func()) {
	gd.queue[p] = append(gd.queue[p], f)
}

func (gd *Grid) Assert(i CellIndex, value int, reason string) {
	fmt.Fprintf(LogFile, "assert: value=%d, cell=%s, reason=%s\n", value+1, i, reason)
	cell := gd.Cells[i.GridIndex()]
	cell.Bits = 1 << uint(value)
	if cell.Value != nil && *cell.Value != value {
		panic(fmt.Errorf("contradictory assertion: already asserted %d @ %s, now trying to assert %d", *cell.Value+1, i, value+1))
	}
	cell.Value = &value
	switch cell.ValueStates[value] {
	case MAYBE:
		gd.clues++
		cell.Maybes = 1

		// reduce the counts associated with the other values in
		// the intersecting groups

		for v, _ := range cell.ValueStates {
			if cell.ValueStates[v] == MAYBE && v != value {
				for _, g := range cell.Groups {
					g.Counts[v] -= 1
					if g.Counts[v] == 1 {
						gd.processUnique(g, v)
					}
				}
			}
			cell.ValueStates[v] = NO
		}
		cell.ValueStates[value] = YES

		for _, g := range cell.Groups {
			for _, c := range g.Cells {
				if c.GridIndex != i.GridIndex() && c.ValueStates[value] == MAYBE {
					gd.Enqueue(IMMEDIATE, gd.heuristicExcludeNeighbours(i, c.Index(), value))
				}
			}
		}

	case YES:
	case NO:
		panic(fmt.Errorf("contradiction: attempted to assert %d @ %v, but this value was previously rejected", value+1, i))
	}
}

func (gd *Grid) processUnique(g *Group, value int) {
	for _, c := range g.Cells {
		if c.ValueStates[value] == MAYBE {
			gd.Enqueue(IMMEDIATE, gd.heuristicUniqueInGroup(g, c, value))
		}
	}
}

func (gd *Grid) Reject(i CellIndex, value int, reason string) {
	fmt.Fprintf(LogFile, "reject: value=%d, cell=%s, reason=%s\n", value+1, i, reason)
	cell := gd.Cells[i.GridIndex()]
	switch cell.ValueStates[value] {
	case MAYBE:
		bit := 1 << uint(value)
		cell.ValueStates[value] = NO
		cell.Bits &^= bit
		cell.Maybes -= 1

		for _, g := range cell.Groups {
			g.Counts[value] -= 1
			if g.Counts[value] == 1 {
				gd.processUnique(g, value)
			}
		}

		if cell.Maybes == 1 {
			gd.Enqueue(IMMEDIATE, gd.heuristicOnlyPossibleValue(cell))
		} else if cell.Maybes == 2 {
			gd.Enqueue(DEFERRED, gd.heuristicExcludePairs(cell))
		} else if cell.Maybes == 3 {
			gd.Enqueue(DEFERRED, func() {
				// TODO: add support for processing triples.
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

func (gd *Grid) Solve() bool {
	if gd.clues < 17 {
		panic(fmt.Sprintf("too few clues (%d) to solve", gd.clues))
	}

mainloop:
	for gd.clues < 81 {
		for len(gd.queue[0]) > 0 && gd.clues < 81 {
			next := gd.queue[0][0]
			gd.queue[0] = gd.queue[0][1:]
			next()
		}

		if gd.clues < 81 {

			// still busy - look for some low priority things to do

			for i, q := range gd.queue {
				if len(q) > 0 {
					gd.queue[0] = append(gd.queue[0], q[0])
					gd.queue[i] = q[1:]
					continue mainloop
				}
			}
			break mainloop
		}
	}

	return gd.clues == 81
}
