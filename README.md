#NAME
sudoku - a golang sudoku solver

#SYNOPSIS
./sudoku [--verbose]|[--version] < puzzle > solution

#DESCRIPTION
'sudoku' implements a heuristic-based command-line Sudoku solver.

The solver does not currently implement a backtracking algorithm. Instead, in order
to make progress, it applies heuristics in the same way that a human player would. As a result, there are some puzzles that the solver cannot currently solve.

##INPUT
Puzzles formatted according to PUZZLE FORMAT are read from stdin.

##PUZZLE FORMAT
Puzzles are formatted as 9 lines of 9 characters each (excluding the line ending).
Numbers are used for clues. Period is used to indicate an absent clue.

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

##DIAGNOSTICS
To view the reasoning of the solver, invoke sudoku with ```--verbose```.

##EXIT CODE
If the solver solves the last input puzzle read from stdin, it exits with a status of 0. Otherwise, it exits with a status of 1. If the last puzzle read from stdin is formatted incorrectly, sudoku exits with a non-zero exit code > 1.

##HEURISTICS
The solver currently implements the following heuristics.

##exclude asserted value from other cells same group
When a value is asserted in a cell, the value is rejected in all other cells of the cell's 3 intersecting groups.

##only value in cell
When it known that the only possible value in a cell is a particular value, then that value is asserted for the cell.

##single remaining position in group for value
When there is only one possible remaining cell for a given value in a given group, then that value is asserted in that cell.

##pair exclusion
When a group contains two cells whose values are known to be restricted to a pair of values, then any other cell in the same group cannot hold either of the two values, so we reject such values in those cells.

##triple exclusion
When a group contains three cells whose values are known to be restricted to a triple of values, then any other cell in the same group cannot hold any of the three values, so we reject such values in those cells.

##MISSING HEURISTICS

The following heuristics are not currently implemented.

##Single Dimensional Block Constraints
If a block group contains 2 or 3 unsolved cells in the same row (column) and there are candidate values are constrained to be in that same row (column) group, then those values can be eliminated from the adjacent rows (columns) in the adjacent row (column) blocks.

#BUILDING
Install the golang tool chain for your host, then run:

```go install```

$GOPATH/bin/sudoku will contain the compiled binary.

#LIMITATIONS
The solver is currently incomplete; there are some puzzles - such as examples/toohard.txt - that cannot be solved using the currently implemented heuristics.

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
##1.1 - 5th October, 2015
* reorganized source code of model package
* add support for clue counting
* improved handling of bad inputs
* added support for queuing heuristics at different priorities
* added additional examples

##1.0 - 5th October, 2015
* initial release

#COPYRIGHT
(c) Jon Seymour 2015

