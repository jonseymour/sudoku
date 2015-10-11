package model

import (
	"fmt"
)

type ValueState int

type CellIndex struct {
	Row    int
	Column int
}

type Cell struct {
	GridIndex   int
	Maybes      int
	Bits        int
	Value       *int
	ValueStates [GROUP_SIZE]ValueState
	Groups      [NUM_GROUP_TYPES]*Group
	Coloring    [GROUP_SIZE]*Coloring
}

const (
	MAYBE ValueState = 0 + iota
	NO
	YES
)

func (i CellIndex) GridIndex() int {
	return i.Row*GROUP_SIZE + i.Column
}

func (i CellIndex) RowGroup() int {
	return i.Row
}

func (i CellIndex) ColumnGroup() int {
	return GROUP_SIZE + i.Column
}

func (i CellIndex) BlockGroup() int {
	return 2*GROUP_SIZE + (i.Row/BLOCK_SIZE)*BLOCK_SIZE + (i.Column / BLOCK_SIZE)
}

func (i CellIndex) RowIndex() int {
	return i.Column
}

func (i CellIndex) ColumnIndex() int {
	return i.Row
}

func (i CellIndex) BlockIndex() int {
	return (i.Row%BLOCK_SIZE)*BLOCK_SIZE + i.Column%BLOCK_SIZE
}

func (i CellIndex) String() string {
	return fmt.Sprintf("(Row:%d, Column:%d, Block:%d)", i.RowGroup()+1, i.ColumnGroup()-GROUP_SIZE+1, i.BlockGroup()-2*GROUP_SIZE+1)
}

func (c *Cell) Index() CellIndex {
	return CellIndex{Row: c.GridIndex / GROUP_SIZE, Column: c.GridIndex % GROUP_SIZE}
}

func (c *Cell) NumConstraints() int {
	count := (GROUP_SIZE - c.Maybes)
	for _, g := range c.Groups {
		count += g.clues
	}
	return count
}

// Answer the set of all cells in the intersecting units that can contain
// the specified value, not including the cell itself.
func (c *Cell) Neighbourhood(value int) *BitSet {
	r := &BitSet{}
	for _, g := range c.Groups {
		r = r.Or(g.Values[value])
	}
	return r.Clear(c.GridIndex)
}
