package model

const (
	ROW GroupType = 0 + iota
	COLUMN
	BLOCK
)

type Group struct {
	Name   string
	Counts [9]int
	Cells  [9]*Cell
}

func (g *Group) String() string {
	return g.Name
}
