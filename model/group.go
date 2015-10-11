package model

const (
	ROW GroupType = 0 + iota
	COLUMN
	BLOCK
)

type Group struct {
	Name   string
	Values [GROUP_SIZE]*BitSet
	Cells  [GROUP_SIZE]*Cell
	Mask   *BitSet

	clues int
}

func (g *Group) String() string {
	return g.Name
}
