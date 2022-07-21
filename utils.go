package main

import (
	"ttt/carr"
	"ttt/room"
)

func roomDetailsToInfo(rd *carr.CArray[RoomDetail]) (values []RoomInfo) {
	values = make([]RoomInfo, 0, rd.Count())
	for tp := range rd.Iter() {
		info := tp.Val
		var state int
		if info.P1Ready && info.P2Ready {
			state = room.InGame
		} else if info.Player1 == "" || info.Player2 == "" {
			state = room.Available
		} else {
			state = room.Full
		}
		values = append(values, RoomInfo{Id: tp.Key, Player1: info.Player1, Player2: info.Player2, State: state})
	}
	return
}
