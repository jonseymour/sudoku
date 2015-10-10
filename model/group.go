package model

const (
	ROW GroupType = 0 + iota
	COLUMN
	BLOCK
)

type Group struct {
	Name   string
	Counts [GROUP_SIZE]int
	Cells  [GROUP_SIZE]*Cell

	clues int
}

func (g *Group) String() string {
	return g.Name
}
