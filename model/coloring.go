package model

import (
	"fmt"
)

// This structure is used to support coloring.
//
// Two cells are linked by a coloring if they are the only two possible
// cells that can contain a particular value in a particular unit (group).
//
// A coloring relationship is transitive. If a and b are linked in a coloring
// and b and c are linked in a coloring then a and c are linked by the same coloring
//
// A cell can be a member of at most one coloring for a given value. When two
// cells that are currently a member of different colorings are linked, one
// of the colorings is merged into the other.
//
type Coloring struct {
	id               int
	on               *BitSet
	off              *BitSet
	onNeighbourhood  *BitSet
	offNeighbourhood *BitSet
}

func (c *Coloring) IsOn(cell *Cell) bool {
	return c.on.Test(cell.GridIndex)
}

func (c *Coloring) Set(cell *Cell, on bool, value int) {
	if on {
		c.on.Set(cell.GridIndex)
		c.onNeighbourhood = c.onNeighbourhood.Or(cell.Neighbourhood(value))
		c.offNeighbourhood.Clear(cell.GridIndex)
	} else {
		c.off.Set(cell.GridIndex)
		c.offNeighbourhood = c.offNeighbourhood.Or(cell.Neighbourhood(value))
		c.onNeighbourhood.Clear(cell.GridIndex)
	}
	cell.Coloring[value] = c
}

func (gd *Grid) ResetColoring(c *Coloring, value int) {
	delete(gd.colorings, c.id)
	for _, i := range c.on.AsInts() {
		gd.Cells[i].Coloring[value] = nil
	}
	for _, i := range c.off.AsInts() {
		gd.Cells[i].Coloring[value] = nil
	}
}

func (gd *Grid) RemoveColoring(c *Coloring, cell *Cell, value int) {
	c.on.Clear(cell.GridIndex)
	c.off.Clear(cell.GridIndex)
	c.onNeighbourhood.Clear(cell.GridIndex)
	c.offNeighbourhood.Clear(cell.GridIndex)
	cell.Coloring[value] = nil
	gd.Enqueue(IMMEDIATE, func() {
		gd.Reject(cell.Index(), value, fmt.Sprintf("Coloring conflict: coloring=%d", c.id))
	})
}

//
// Pre-condition: both cells are a member of the same unit (group) and the number
// of possible cells for that value in that group is exactly 2.
//
func (gd *Grid) Color(cell1 *Cell, cell2 *Cell, value int) {
	var coloring *Coloring
	if (cell1.Coloring[value] == nil) && (cell2.Coloring[value] == nil) {
		on := (&BitSet{}).Set(cell1.GridIndex)
		off := (&BitSet{}).Set(cell2.GridIndex)
		coloring = &Coloring{
			id:               gd.numColorings,
			on:               on,
			onNeighbourhood:  cell1.Neighbourhood(value).AndNot(off).AndNot(on),
			off:              off,
			offNeighbourhood: cell2.Neighbourhood(value).AndNot(on).AndNot(off),
		}
		gd.numColorings++
		gd.colorings[gd.id] = coloring
		cell1.Coloring[value] = coloring
		cell2.Coloring[value] = coloring
		if Verbose {
			fmt.Fprintf(
				LogFile,
				"coloring: new coloring grid=%d, coloring=%d, cell1=%s, cell2=%s, value=%d\n",
				gd.id,
				coloring.id,
				cell1.Index(),
				cell2.Index(),
				value+1)
		}
	} else if cell1.Coloring[value] == cell2.Coloring[value] {
		coloring = cell1.Coloring[value]
		if coloring.IsOn(cell1) == coloring.IsOn(cell2) {

			if Verbose {
				fmt.Fprintf(
					LogFile,
					"coloring: contradiction grid=%d, coloring=%d, cell1=%s, cell2=%s, value=%d\n",
					gd.id,
					coloring.id,
					cell1.Index(),
					cell2.Index(),
					value+1)
			}
			panic("contradiction: Coloring inconsistency")
		}

	} else if cell1.Coloring[value] == nil {
		coloring = cell2.Coloring[value]
		color := !coloring.IsOn(cell2)
		coloring.Set(cell1, color, value)
		if Verbose {
			fmt.Fprintf(
				LogFile,
				"coloring: extension grid=%d, coloring=%d, cell1=%s, cell2=%s, value=%d\n",
				gd.id,
				coloring.id,
				cell2.Index(),
				cell1.Index(),
				value+1)
		}
	} else if cell2.Coloring[value] == nil {
		coloring = cell1.Coloring[value]
		color := !coloring.IsOn(cell1)
		coloring.Set(cell2, color, value)
		if Verbose {
			fmt.Fprintf(
				LogFile,
				"coloring: extension grid=%d, coloring=%d, cell1=%s, cell2=%s, value=%d\n",
				gd.id,
				coloring.id,
				cell1.Index(),
				cell2.Index(),
				value+1)
		}
	} else {
		// merge the two colorings into a single coloring and ensure
		// a consistent coloring is used for each
		kept := cell1.Coloring[value]
		discarded := cell2.Coloring[value]
		coloring = kept
		if kept.IsOn(cell1) == discarded.IsOn(cell2) {
			kept.on = kept.on.Or(discarded.off)
			kept.onNeighbourhood = kept.onNeighbourhood.Or(discarded.offNeighbourhood).AndNot(discarded.on)
			kept.off = kept.off.Or(discarded.on)
			kept.offNeighbourhood = kept.offNeighbourhood.Or(discarded.onNeighbourhood).AndNot(discarded.off)
		} else {
			kept.on = kept.on.Or(discarded.on)
			kept.onNeighbourhood = kept.onNeighbourhood.Or(discarded.onNeighbourhood).AndNot(discarded.off)
			kept.off = kept.off.Or(discarded.off)
			kept.offNeighbourhood = kept.offNeighbourhood.Or(discarded.offNeighbourhood).AndNot(discarded.on)
		}
		for _, i := range discarded.on.AsInts() {
			gd.Cells[i].Coloring[value] = kept
		}
		for _, i := range discarded.off.AsInts() {
			gd.Cells[i].Coloring[value] = kept
		}
		delete(gd.colorings, discarded.id)
		if Verbose {
			fmt.Fprintf(
				LogFile,
				"coloring: merge grid=%d, kept=%d, discarded=%d, cell1=%s, cell2=%s, value=%d\n",
				gd.id,
				kept.id,
				discarded.id,
				cell1.Index(),
				cell2.Index(),
				value+1)
		}
	}
	intersection := coloring.offNeighbourhood.And(coloring.onNeighbourhood)
	if intersection.Size() > 0 {
		for _, i := range intersection.AsInts() {
			gd.RemoveColoring(coloring, gd.Cells[i], value)
		}
	}
}
