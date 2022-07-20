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

func logic(bin *BoardInfo, index int) {
	bin.Board[index] = bin.Turn
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
			bin.Result = win.Black
		} else {
			bin.Result = win.White
		}
	} else if lines == 1 {
		if bin.Turn == piece.
			Black {
			bin.Result = win.White
		} else {
			bin.Result = win.Black
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
			bin.Result = win.Tie
		}
	}
}
