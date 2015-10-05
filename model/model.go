package model

import (
	"io"
	"io/ioutil"
)

type ValueState int
type GroupType int

var LogFile io.Writer = ioutil.Discard

const (
	MAYBE ValueState = 0 + iota
	NO
	YES
)
