#NAME
sudoku - a golang sudoku solver

#SYNOPSIS
./sudoku [--verbose]|[--version]|[--format=9.|90|1.|10] < puzzle > solution

#DESCRIPTION
'sudoku' implements a heuristic-based command-line Sudoku solver.

The solver attempts to make progress using heuristics and only falls back
to a brute force (or backtracking) approach when available heuristics are
exhausted.

##INPUT
One or more puzzles formatted according to PUZZLE FORMAT are read from stdin.

##PUZZLE FORMAT
Puzzles are formatted as 9 lines of 9 characters each or as a single line of 81 characters (lengths exclude the line ending in both cases). Positive numbers are used for clues; a period (.) or zero (0) is used to indicate a missing clue.

Blank lines or lines beginning with a leading comment indicator ('#') are ignored.

For example, examples/puzzle.txt:

```
....4.7..
.....1.5.
84...2..3
1..5...3.
6.9..7...
5..1...2.
78...5..1
.....6.4.
....3.2..
```

Other examples may be found in the examples/ subdirectory.

##OUTPUT
If the solver can solve the puzzle, it outputs the solution on stdout. Otherwise, it outputs a partial solution on stdout.

##OUTPUT FORMAT
The default output format is a 9x9 grid with (.) to indicate a missing clue.
The --format option can be used to select an alternative output format. The
following options are available: 9., 90, 1., 10 where the first character indicates
the number of lines per solution and the second character indicates which
character is used to indicate a missing clue.

##DIAGNOSTICS
To view the reasoning of the solver, invoke sudoku with ```--verbose```.

##EXIT CODE
If the solver solves the last input puzzle read from stdin, it exits with a status of 0. Otherwise, it exits with a status of 1. If the last puzzle read from stdin is formatted incorrectly, sudoku exits with a non-zero exit code > 1.

##HEURISTICS
The solver currently implements the following heuristics.

##Direct Exclusion
When a value is asserted in a cell, the value is rejected in all other cells of the cell's 3 intersecting groups.

##Naked Single
When it known that the only possible value in a cell is a particular value, then that value is asserted for the cell.

##Hidden Single
When there is only one possible remaining cell for a given value in a given group, then that value is asserted in that cell.

##Naked Pair
When a group contains two cells whose values are known to be restricted to a pair of values, then any other cell in the same group cannot hold either of the two values, so we reject such values in those cells.

##Naked Triple
When a group contains three cells whose values are known to be restricted to a triple of values, then any other cell in the same group cannot hold any of the three values, so we reject such values in those cells.

##Exclude Complement
If a block contains 2 or 3 unsolved cells in a single row (or column) and those
cells are the only cells in the block that can contain a particular value, then that value can be rejected from the same row (or column) in other blocks.

This heuristic is known in other places as Pointed Pairs/Triples or Block/Line Reduction, depending on the context.

##Speculative Assertion
When we run out of other heuristics to try, we clone the current state of the
solution and speculatively assert one of the possible cell/value pairs and see
what happens.

##Contradiction Found
If a speculative assertion produced a contradiction, then we reject the cell/value
pair that was the subject of the speculative assertion in the original solution.

##Verify Uniqueness
If a speculative assertion finds a solution, we need to verify that the solution
is the only solution. We do this by rejecting the speculatively asserted cell/value
pair in a new clone of the puzzle taken prior to the speculative fork. If this
does not produce a contradiction, then the speculatively determined solution
is not unique and therefore the original puzzle does not have a unique completion.

##Coloring Conflict
When a value is found that is restricted to a pair of cells within a group then if one cell contains the value, the other one must not and vice versa. We note this by coloring one cell with an 'on' color and the other cell with an 'off' color. Furthermore, if cells 'a' and 'b' are a pair in the same coloring network and 'b' and 'c' are a pair in the same coloring network, then all of 'a', 'b' and 'c' must be in the same coloring network and cells 'a' and 'c' must be of the same color.

Each coloring network has two sets - an 'on' and and 'off' set. The 'on' set consists of those cells that are colored 'on' by the network, the 'off' set consists of the other cells which, by definition, are colored 'off'.

Each coloring network also has a two sets of neighbours - one for each color. These neighbours are cells that may contain the same value as the coloring network and intersect with groups that intersect with cells that are in the coloring network.

As each group restricted pair of cells is discovered we create, extend or merge the coloring networks associated with each member of the pair and also extend the neighbourhood of each color (for that coloring network). If we ever discover a non-empty intersection between the neighbourhoods of each color, then we can exclude the value from the cells in the intersection since a given cell can't simultaneously be both on and off.

#BUILDING
Install the golang tool chain for your host, then run:

```go install```

$GOPATH/bin/sudoku will contain the compiled binary.

#TERMINOLOGY
##Cell
A single cell
##Row
A horizontal group of 9 cells. Rows are numbered 1-9 from top to bottom.
##Column
A vertical group of 3 cells. Columns are numbered 1-9 from left to right.
##Block
A group of 9 cells arranged in a 3x3 grid, aligned on boundaries that are multiples of 3 + 1. Blocks are numbered 1-9 from top-left to bottom-right.
##Group
A group is a collection of 9 cells organized as a either a row, column or a block.
##Intersecting Group
Each cell intersects with 3 groups - the so-called 'intersecting groups' of the cell. Each cell has one intersecting group of each type: row, column and block.

#REVISION HISTORY
##1.2.1 - 11th October, 2015
* extended with coloring support

##1.2 - 10th October, 2015
* allow parser to accept puzzles using Royle's format.
* fixed an error in backtracker which caused some contradictions to be missed
* updated to use the same terminology as sudokuwiki and other places
* try to find the most constrained guess during backtracking
* refine documentation of PUZZLE FORMAT and INPUT
* factored out stream io into its own package
* add support for a --format option to select the output puzzle format

##1.1 - 9th October, 2015
* reorganized source code of model package
* add support for clue counting
* improved handling of bad inputs
* added support for queuing heuristics at different priorities
* added additional examples
* added support for backtracking and ambiguous puzzle detection

##1.0 - 5th October, 2015
* initial release

#REFERENCES

* [1] "Minimal Sudoku" - http://staffhome.ecm.uwa.edu.au/~00013890/sudokumin.php
* [2] "Sudoku" - http://mathworld.wolfram.com/Sudoku.html
* [3] "Sudoku" - https://en.wikipedia.org/wiki/Sudoku
* [4] www.sudoku.com - http://www.sudoku.com/
* [5] "Sudoku Wiki" - http://www.sudokuwiki.org/

#COPYRIGHT
(c) Jon Seymour 2015

