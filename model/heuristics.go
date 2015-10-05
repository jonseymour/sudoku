package model

import (
	"fmt"
)

// Reject the asserted value in a neighbour of the asserted cell.
func (gd *Grid) heuristicExcludeNeighbours(asserted CellIndex, neighbour CellIndex, assertedValue int) func() {
	return func() {
		gd.Reject(neighbour, assertedValue, fmt.Sprintf("excluded by assertion of %d @ %s", assertedValue+1, asserted))
	}
}

// The cell is a candidate for being the only remaining member of the group
// who can hold the value. Note: at the time it is registered, pending rejects
// may still be in effect, so we need to check again when it executes to be sure.
func (gd *Grid) heuristicUniqueInGroup(g *Group, cell *Cell, value int) func() {
	return func() {
		if cell.ValueStates[value] == MAYBE {
			gd.Assert(cell.Index(), value, fmt.Sprintf("unique value found in group %s", g))
		}
	}
}

// When the cell has only one possible value, find and assert that cell.
func (gd *Grid) heuristicOnlyPossibleValue(cell *Cell) func() {
	return func() {
		for v, s := range cell.ValueStates {
			if s == MAYBE {
				gd.Enqueue(IMMEDIATE, func() {
					gd.Assert(cell.Index(), v, "only possible value in cell")
				})
				return
			}
		}
	}
}

// When two cells in g group can only be satisfied by a pair of values, we can
// exclude those values from every other cell in the same group. This heuristic
// does this for one cell (exclude)
func (gd *Grid) heuristicExcludePair(p1 *Cell, p2 *Cell, exclude *Cell, pair []int) func() {
	return func() {
		if exclude.ValueStates[pair[0]] == MAYBE {
			gd.Reject(exclude.Index(), pair[0], fmt.Sprintf("excluded by pair (%v,%v) @ %s, %s", pair[0], pair[1], p1.Index(), p2.Index()))
		}
		if exclude.ValueStates[pair[1]] == MAYBE {
			gd.Reject(exclude.Index(), pair[1], fmt.Sprintf("excluded by pair (%v,%v) @ %s, %s", pair[0], pair[1], p1.Index(), p2.Index()))
		}
	}
}

// When two cells in a group can only be satisfied by a pair of values, we can
// exclude those values from every other cell in the same group. This heuristic
// does this for every intersecting group of the specified cell that has another
// cell containing the matching pair.
func (gd *Grid) heuristicExcludePairs(cell *Cell) func() {
	return func() {
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
							if c1.GridIndex != c.GridIndex && c1.GridIndex != cell.GridIndex {
								gd.Enqueue(IMMEDIATE, gd.heuristicExcludePair(cell, c, c1, pair))
							}
						}
					}
				}
			}
		}
	}
}
