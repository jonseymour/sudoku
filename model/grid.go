package model

import (
	"fmt"
)

type Priority int

const (
	IMMEDIATE Priority = iota
	DEFERRED
)

var gridCount int

// A grid consists of cells and groups and state that tracks the number
// of asserted clues and a queue of pending heuristics.
type Grid struct {
	Groups [NUM_GROUPS]*Group
	Cells  [NUM_CELLS]*Cell

	id        int
	clues     int
	queue     [NUM_PRIORITIES][]func()
	ambiguity error
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
	gridCount++
	grid.id = gridCount
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

func (gd *Grid) copyState(dest *Grid) {
	dest.clues = gd.clues

	for i, g := range gd.Groups {
		for j, c := range g.Counts {
			dest.Groups[i].Counts[j] = c
		}
		dest.Groups[i].clues = g.clues
	}

	for i, c := range gd.Cells {
		cell := dest.Cells[i]
		cell.Maybes = c.Maybes
		cell.Value = c.Value
		cell.Bits = c.Bits
		for v, s := range c.ValueStates {
			cell.ValueStates[v] = s
		}
	}
}

// Creates a clone of the receiving dest which duplicates all the state
// except for the work queue.
func (gd *Grid) Clone() *Grid {
	grid := NewGrid()
	gd.copyState(grid)
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
	fmt.Fprintf(LogFile, "assert: grid=%d, value=%d, cell=%s, reason=%s\n", gd.id, value+1, i, reason)
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
			g.clues++
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
	fmt.Fprintf(LogFile, "reject: grid=%d, value=%d, cell=%s, reason=%s\n", gd.id, value+1, i, reason)
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

// pick one of the undecided cells and one of the
// available values for that cell.
func (gd *Grid) guess() (CellIndex, int) {

	// find the most constrained cell

	var constrained *Cell = nil
	constraints := 0

	for _, c := range gd.Cells {
		if c.Value != nil {
			continue
		}
		actual := c.NumConstraints()
		if actual > constraints {
			constraints = actual
			constrained = c
		}
	}

	if constrained == nil {
		panic("could not find an unresolved cell")
	}

	// find an unresolved value for that cell

	for v, s := range constrained.ValueStates {
		if s == MAYBE {
			return constrained.Index(), v
		}
	}

	panic("could not find an unresolved value")
}

// Try to find the solution for the receiving grid by
// cloning it and speculatively asserting the specified
// value at the specified index.
//
// If we find a solution, then we check that it is unique
// by trying to find an alternative value for the same cell
// that produces a different solution.
//
// If the solution is unique, there should be no such solution.
func (gd *Grid) speculate(index CellIndex, value int) (bool, error) {
	copy := gd.Clone()

	copy.Assert(index, value, fmt.Sprintf("guessing %d @ %s", value+1, index))
	solved, err := copy.Solve()

	if solved {

		// value @ index - produces a valid solution

		// we verify that there is no other solution at the same index
		// by rejecting the value at the cell and trying to solve
		// the resulting grid.

		alt := gd.Clone()
		alt.Reject(index, value, fmt.Sprintf("testing that alternative is not valid"))
		r2, _ := alt.Solve()
		if r2 {
			fmt.Fprintf(LogFile, "ambiguous (%d): %s\n", gd.id, gd)
			fmt.Fprintf(LogFile, "solution 1 (%d): %s\n", copy.id, copy)
			fmt.Fprintf(LogFile, "solution 2 (%d): %s\n", alt.id, alt)
			gd.ambiguity = fmt.Errorf("ambiguity @ %s - both values yield valid solutions - %d (grid=%d), %d (grid=%d)", index, *copy.Cells[index.GridIndex()].Value+1, copy.id, *alt.Cells[index.GridIndex()].Value+1, alt.id)
			return true, gd.ambiguity
		} else if alt.ambiguity != nil {
			// if the modified grid failed because of an ambiguity, then
			// there are multiple solutions in addition to the one we found
			// which implies that the solution is not unique
			gd.ambiguity = alt.ambiguity
			return true, gd.ambiguity
		}

		// the solution in copy is unique, so update gd with copy and indicate
		// we are done.

		copy.copyState(gd)

		return true, nil
	} else {

		if copy.ambiguity != nil {

			// the trial failed because of an ambiguity
			// propagate this ambiguity to the receiver

			gd.ambiguity = copy.ambiguity
			return true, gd.ambiguity
		}

		// since asserting the value @ index produced a contradiction
		// we can now reject the value

		gd.Reject(index, value, fmt.Sprintf("rejecting guess of %d @ %s after contradiction: %s", value+1, index, err))
		return false, nil
	}
}

// Solve the grid by iterating over the work queue until NUM_CELLS
// are obtained or until there is nothing else to try.
//
// Work is prioritsed so that cheaper heuristics are tried first and
// more expensive heuristics are only tried if there are no more
// cheap heuristics to try. Returns true if the solution is obtained, false
// otherwise.
func (gd *Grid) Solve() (bool, error) {

	result := make(chan error)

	go func() {

		defer func() {
			if r := recover(); r != nil {
				result <- fmt.Errorf("%v", r)
			}
		}()

	mainloop:
		for {
			for len(gd.queue[0]) > 0 {
				next := gd.queue[0][0]
				gd.queue[0] = gd.queue[0][1:]
				next()
			}

			if gd.clues < NUM_CELLS {

				// unsolved, but there are some lower priority
				// heuristics to try. move one of these
				// into the priority 0 queue, and continue

				for i, q := range gd.queue {
					if len(q) > 0 {
						gd.queue[0] = append(gd.queue[0], q[0])
						gd.queue[i] = q[1:]
						continue mainloop
					}
				}

				// there are no more heuristics available.
				// guess a cell value, and test whether this
				// produces a unique solution.
				done, err := gd.speculate(gd.guess())
				if done {
					result <- err
					return
				}
			} else {
				result <- nil
				return
			}
		}
	}()

	err := <-result
	return gd.clues == NUM_CELLS && err == nil, err
}

func (gd *Grid) String() string {
	result := [81]byte{}
	for i, c := range gd.Cells {
		if c.Value == nil {
			result[i] = byte('.')
		} else {
			result[i] = byte(int32(*c.Value) + int32('1'))
		}
	}
	return string(result[0:])
}
