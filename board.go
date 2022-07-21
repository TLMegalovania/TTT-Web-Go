package main

import (
	"ttt/piece"
	"ttt/win"
)

const (
	Row = 5
	Col = 5
)

var Dirs = [][][]int{
	{
		{-2, 0},
		{-1, 0},
		{0, 0},
		{1, 0},
		{2, 0},
	},
	{
		{0, -2},
		{0, -1},
		{0, 0},
		{0, 1},
		{0, 2},
	},
	{
		{-2, -2},
		{-1, -1},
		{0, 0},
		{1, 1},
		{2, 2},
	},
	{
		{2, -2},
		{1, -1},
		{0, 0},
		{-1, 1},
		{-2, 2},
	},
}

// pure function
func logic(bin BoardInfo, index int) int {
	lines := 0
	p0, p1 := index/Col, index%Col
	for _, dir := range Dirs {
		line := 0
		for _, dp := range dir {
			dp0, dp1 := p0+dp[0], p1+dp[1]
			if !(dp0 < 0 || dp0 >= Row || dp1 < 0 || dp1 >= Col) {
				if bin.Board[dp0*Col+dp1] == bin.Turn {
					line++
					if line >= 3 {
						break
					}
				} else {
					line = 0
				}
			}
		}
		if line >= 3 {
			lines++
			if lines >= 2 {
				break
			}
		}
	}
	if lines >= 2 {
		if bin.Turn == piece.
			Black {
			return win.Black
		} else {
			return win.White
		}
	} else if lines == 1 {
		if bin.Turn == piece.
			Black {
			return win.White
		} else {
			return win.Black
		}
	} else {
		full := true
		for _, v := range bin.Board {
			if v == piece.Null {
				full = false
				break
			}
		}
		if full {
			return win.Tie
		}
	}
	return win.Null
}
