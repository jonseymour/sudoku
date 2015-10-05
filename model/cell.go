package model

import (
	"fmt"
)

type CellIndex struct {
	Row    int
	Column int
}

func (i CellIndex) GridIndex() int {
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
	GridIndex   int
	Maybes      int
	Bits        int
	Value       *int
	ValueStates [9]ValueState
	Groups      [3]*Group
}

func (c *Cell) Index() CellIndex {
	return CellIndex{Row: c.GridIndex / 9, Column: c.GridIndex % 9}
}
