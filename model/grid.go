package model

import (
	"fmt"
)

type Grid struct {
	Groups [27]*Group
	Cells  [81]*Cell

	clues int
	queue []func()
}

func NewGrid() *Grid {
	grid := &Grid{}
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
					x := c.Index()
					gd.queue = append(gd.queue, func() {
						gd.Reject(x, value, fmt.Sprintf("excluded by assertion of %d @ %s", value+1, i))
					})
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
		x := c.Index()
		cell := c
		if c.ValueStates[value] == MAYBE {
			gd.queue = append(gd.queue, func() {
				if cell.ValueStates[value] == MAYBE {
					gd.Assert(x, value, fmt.Sprintf("unique value found in group %s", g))
				}
			})
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
			// if a cell has only one maybe, then we assert that value
			// as the cell's value

			gd.queue = append(gd.queue, func() {
				for v, s := range cell.ValueStates {
					if s == MAYBE {
						x := cell.Index()
						gd.queue = append(gd.queue, func() {
							gd.Assert(x, v, "only possible value in cell")
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

			gd.queue = append(gd.queue, func() {
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
							if c.GridIndex != cell.GridIndex && c.Bits == cell.Bits {
								// found a matching pair in the same group
								for _, c1 := range g.Cells {
									x := c1.Index()
									if c1.GridIndex != c.GridIndex && c1.GridIndex != cell.GridIndex {
										gd.queue = append(gd.queue, func() {
											if c1.ValueStates[pair[0]] == MAYBE {
												gd.Reject(x, pair[0], fmt.Sprintf("excluded by pair (%v,%v) @ %s, %s", pair[0], pair[1], cell.Index(), c.Index()))
											}
											if c1.ValueStates[pair[1]] == MAYBE {
												gd.Reject(x, pair[1], fmt.Sprintf("excluded by pair (%v,%v) @ %s, %s", pair[0], pair[1], cell.Index(), c.Index()))
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
			gd.queue = append(gd.queue, func() {
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
	for len(gd.queue) > 0 && gd.clues < 81 {
		next := gd.queue[0]
		gd.queue = gd.queue[1:]
		next()
	}
	return gd.clues == 81
}
