package model

import (
	"fmt"
)

// Reject the asserted value in a neighbour of the asserted cell.
func (gd *Grid) heuristicExcludeNeighbours(asserted CellIndex, neighbour CellIndex, assertedValue int) func() {
	return func() {
		gd.Reject(neighbour, assertedValue, fmt.Sprintf("Direct Exclusion by %s", asserted))
	}
}

// The cell is a candidate for being the only remaining member of the group
// who can hold the value. Note: at the time it is registered, pending rejects
// may still be in effect, so we need to check again when it executes to be sure.
func (gd *Grid) heuristicUniqueInGroup(g *Group, value int) func() {
	return func() {
		for _, c := range g.Cells {
			if c.ValueStates[value] == MAYBE {
				gd.Assert(c.Index(), value, fmt.Sprintf("Hidden Single in %s", g))
			}
		}
	}
}

// When the cell has only one possible value, find and assert that cell.
func (gd *Grid) heuristicExcludeSingleton(cell *Cell) func() {
	return func() {
		for v, s := range cell.ValueStates {
			if s == MAYBE {
				gd.Assert(cell.Index(), v, "Naked Single")
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
			gd.Reject(exclude.Index(), pair[0], fmt.Sprintf("Naked Pair (%v,%v) @ %s, %s", pair[0]+1, pair[1]+1, p1.Index(), p2.Index()))
		}
		if exclude.ValueStates[pair[1]] == MAYBE {
			gd.Reject(exclude.Index(), pair[1], fmt.Sprintf("Naked Pair (%v,%v) @ %s, %s", pair[0]+1, pair[1]+1, p1.Index(), p2.Index()))
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

func (gd *Grid) heuristicExcludeTriple(t1 *Cell, t2 *Cell, t3 *Cell, exclude *Cell, triple []int) func() {
	return func() {
		if exclude.ValueStates[triple[0]] == MAYBE {
			gd.Reject(exclude.Index(), triple[0], fmt.Sprintf("Naked Triple (%v,%v,%v) @ (%s, %s, %s)", triple[0]+1, triple[1]+1, triple[2]+1, t1.Index(), t2.Index(), t3.Index()))
		}
		if exclude.ValueStates[triple[1]] == MAYBE {
			gd.Reject(exclude.Index(), triple[1], fmt.Sprintf("Naked Triple (%v,%v,%v) @ (%s, %s, %s)", triple[0]+1, triple[1]+1, triple[2]+1, t1.Index(), t2.Index(), t3.Index()))
		}
		if exclude.ValueStates[triple[2]] == MAYBE {
			gd.Reject(exclude.Index(), triple[2], fmt.Sprintf("Naked Triple (%v,%v,%v) @ (%s, %s, %s)", triple[0]+1, triple[1]+1, triple[2]+1, t1.Index(), t2.Index(), t3.Index()))
		}
	}
}

func (gd *Grid) heuristicExcludeTriples(cell *Cell) func() {
	return func() {
		if cell.Maybes == 3 {
			triple := []int{}
			for v, s := range cell.ValueStates {
				if s == MAYBE {
					triple = append(triple, v)
				}
			}
			if len(triple) != 3 {
				panic("expected: len(triple) == 3")
			}
			for _, g := range cell.Groups {
				tripleIndex := []int{cell.GridIndex}
				for _, c := range g.Cells {
					if c.GridIndex != cell.GridIndex &&
						c.Value == nil &&
						(c.Bits&^cell.Bits) == 0 {
						tripleIndex = append(tripleIndex, c.GridIndex)
						// found a matching pair in the same group
					}
				}
				if len(tripleIndex) == 3 {
					a := tripleIndex[0]
					b := tripleIndex[1]
					c := tripleIndex[2]
					for _, t1 := range g.Cells {
						if (t1.GridIndex != a) && (t1.GridIndex != b) && (t1.GridIndex != c) {
							gd.Enqueue(IMMEDIATE, gd.heuristicExcludeTriple(gd.Cells[a], gd.Cells[b], gd.Cells[c], t1, triple))
						}
					}
				}
			}
		}
	}
}

//
func (gd *Grid) heuristicExcludeComplement(g1 *Group, g2 *Group, value int, count int) func() {
	return func() {
		tmp := []*Cell{}
		i := 0
		j := 0
		k := 0
		for i < GROUP_SIZE && j < GROUP_SIZE {
			diff := g1.Cells[i].GridIndex - g2.Cells[j].GridIndex
			if diff > 0 {
				tmp = append(tmp, g2.Cells[j])
				j++
			} else if diff < 0 {
				i++
			} else {
				if g2.Cells[j].ValueStates[value] == MAYBE {
					k++
				}
				j++
				i++
			}
		}
		for j < GROUP_SIZE {
			tmp = append(tmp, g2.Cells[j])
			j++
		}
		if k == count {
			for _, c := range tmp {
				if c.ValueStates[value] == MAYBE {
					gd.Reject(c.Index(), value, fmt.Sprintf("Exclude Complement of %s's intersection with %s", g2, g1))
				}
			}
		}
	}
}
