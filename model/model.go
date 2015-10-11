package model

import (
	"io"
	"io/ioutil"
)

type GroupType int

var LogFile io.Writer = ioutil.Discard
var Verbose bool = false
var VerifyUniqueness bool = true

const (
	BLOCK_SIZE      = 3
	GROUP_SIZE      = 9
	NUM_CELLS       = 81
	NUM_GROUP_TYPES = 3
	NUM_GROUPS      = 27
	NUM_PRIORITIES  = 2
	MIN_CLUES       = 17
)
