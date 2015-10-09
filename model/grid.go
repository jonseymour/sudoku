package model

import (
	"fmt"
)

type Priority int

const (
	IMMEDIATE Priority = iota
	DEFERRED
)

// A grid consists of cells and groups and state that tracks the number
// of asserted clues and a queue of pending heuristics.
type Grid struct {
	Groups [NUM_GROUPS]*Group
	Cells  [NUM_CELLS]*Cell

	clues int
	queue [NUM_PRIORITIES][]func()
}

// Initialize a new grid. No clues are asserted.
// All cells are initialized with references to their
// intersecting groups and vice versa.
//
// Counters are initialized to maximum values for the
//  - counts by value by group (Group.Counts)
//  - maybes by cell (Cell.Maybes)
func NewGrid() *Grid {
	grid := &Grid{
		queue: [NUM_PRIORITIES][]func(){
			[]func(){},
			[]func(){},
		},
	}
	for x, _ := range grid.Groups {
		g := &Group{}
		grid.Groups[x] = g
		for i, _ := range g.Counts {
			g.Counts[i] = GROUP_SIZE
			if x < GROUP_SIZE {
				g.Name = fmt.Sprintf("Row:%d", x+1)
			} else if x < 2*GROUP_SIZE {
				g.Name = fmt.Sprintf("Column:%d", x-GROUP_SIZE+1)
			} else {
				g.Name = fmt.Sprintf("Block:%d", x-(2*GROUP_SIZE)+1)
			}
		}
	}
	for r := 0; r < GROUP_SIZE; r++ {
		for c := 0; c < GROUP_SIZE; c++ {
			i := &CellIndex{Row: r, Column: c}

			cell := &Cell{}
			grid.Cells[i.GridIndex()] = cell
			cell.Bits = 1<<GROUP_SIZE - 1
			cell.GridIndex = i.GridIndex()
			cell.Maybes = GROUP_SIZE

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

// Creates a clone of the receiving grid which duplicates all the state
// except for the work queue.
func (gd *Grid) Clone() *Grid {
	grid := NewGrid()
	grid.clues = gd.clues

	for i, g := range gd.Groups {
		for j, c := range g.Counts {
			grid.Groups[i].Counts[j] = c
		}
	}

	for i, c := range gd.Cells {
		cell := grid.Cells[i]
		cell.Maybes = c.Maybes
		cell.Value = c.Value
		cell.Bits = c.Bits
		for v, s := range c.ValueStates {
			cell.ValueStates[v] = s
		}
	}
	return grid
}

// Enqueue a heuristic to try with the specified priority
func (gd *Grid) Enqueue(p Priority, f func()) {
	gd.queue[p] = append(gd.queue[p], f)
}

// Decrement the specified value counts for all intersecting groups of the specified
// cell by 1. If a count drops to 1, apply the unique in group heuristic.
func (gd *Grid) adjustValueCounts(cell *Cell, value int) {
	for _, g := range cell.Groups {
		g.Counts[value] -= 1
		if g.Counts[value] == 1 {
			gd.Enqueue(DEFERRED, gd.heuristicUniqueInGroup(g, value))
		}
	}

	f := func(g1 *Group, g2 *Group) {
		c := g1.Counts[value]
		switch c {
		case 2, 3:
			if g2.Counts[value] > c {
				gd.Enqueue(DEFERRED, gd.heuristicExcludeComplement(g1, g2, value, c))
			}
		}
	}

	f(cell.Groups[BLOCK], cell.Groups[ROW])
	f(cell.Groups[BLOCK], cell.Groups[COLUMN])
	f(cell.Groups[ROW], cell.Groups[BLOCK])
	f(cell.Groups[COLUMN], cell.Groups[BLOCK])
}

// Assert that the grid contains the specified value at the specified index.
// 'reason' contains an English language justification for the belief.
func (gd *Grid) Assert(i CellIndex, value int, reason string) {
	fmt.Fprintf(LogFile, "assert: value=%d, cell=%s, reason=%s\n", value+1, i, reason)
	cell := gd.Cells[i.GridIndex()]
	if cell.Value != nil && *cell.Value != value {
		panic(fmt.Errorf("contradictory assertion: already asserted %d @ %s, now trying to assert %d", *cell.Value+1, i, value+1))
	}

	switch cell.ValueStates[value] {
	case MAYBE:
		cell.ValueStates[value] = YES
		cell.Value = &value
		cell.Bits = 1 << uint(value)
		cell.Maybes = 1

		gd.clues++

		// available slots for all maybes in newly asserted cell are
		// reduced by one in each intersecting group update those totals
		// now and schedule work if a unique value is found in one group

		for v, _ := range cell.ValueStates {
			if cell.ValueStates[v] == MAYBE {
				cell.ValueStates[v] = NO
				gd.adjustValueCounts(cell, v)
			}
		}

		for _, g := range cell.Groups {
			for _, c := range g.Cells {
				if c.ValueStates[value] == MAYBE {
					gd.Enqueue(IMMEDIATE, gd.heuristicExcludeNeighbours(i, c.Index(), value))
				}
			}
		}

	case NO:
		panic(fmt.Errorf("contradiction: attempted to assert %d @ %v, but this value was previously rejected", value+1, i))
	case YES:
		// do nothing
	}
}

// Assert that the grid does not contain the specified value at the specified
// cell index. 'reason' contains am English language justification for the belief.
func (gd *Grid) Reject(i CellIndex, value int, reason string) {
	fmt.Fprintf(LogFile, "reject: value=%d, cell=%s, reason=%s\n", value+1, i, reason)
	cell := gd.Cells[i.GridIndex()]
	switch cell.ValueStates[value] {
	case MAYBE:
		bit := 1 << uint(value)
		cell.ValueStates[value] = NO
		cell.Bits &^= bit
		cell.Maybes -= 1

		gd.adjustValueCounts(cell, value)

		if cell.Maybes == 1 {
			gd.Enqueue(IMMEDIATE, gd.heuristicExcludeSingleton(cell))
		} else if cell.Maybes == 2 {
			gd.Enqueue(DEFERRED, gd.heuristicExcludePairs(cell))
		} else if cell.Maybes == 3 {
			gd.Enqueue(DEFERRED, gd.heuristicExcludeTriples(cell))
		}
	case YES:
		if value == *cell.Value {
			panic(fmt.Errorf("contradiction: attempt to reject value=%d @ %s, but this value was previously asserted", value+1, i))
		}
	case NO:
		// do nothing
	}
}

// Solve the grid by iterating over the work queue until MIN_CLUES
// are obtained or until there is nothing else to try.
// Work is prioritsed so that cheaper heuristics are tried first and
// more expensive heuristics are only tried if there are no more
// cheap heuristics to try. Returns true if the solution is obtained, false
// otherwise.
func (gd *Grid) Solve() (bool, error) {
	if gd.clues < MIN_CLUES {
		return false, fmt.Errorf("too few clues (%d) to solve", gd.clues)
	}

	result := make(chan error)

	go func() {

		defer func() {
			if r := recover(); r != nil {
				result <- fmt.Errorf("%v", r)
			}
		}()

	mainloop:
		for gd.clues < NUM_CELLS {
			for len(gd.queue[0]) > 0 && gd.clues < NUM_CELLS {
				next := gd.queue[0][0]
				gd.queue[0] = gd.queue[0][1:]
				next()
			}

			if gd.clues < NUM_CELLS {

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
		result <- nil

	}()

	return gd.clues == NUM_CELLS, <-result
}
