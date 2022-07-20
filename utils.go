package main

import (
	"ttt/room"

	cmap "github.com/orcaman/concurrent-map"
)

func roomDetailsToInfo(rd *cmap.ConcurrentMap) (values []RoomInfo) {
	values = make([]RoomInfo, 0, rd.Count())
	for tp := range rd.IterBuffered() {
		info := tp.Val.(RoomDetail)
		var state int
		if info.P1Ready && info.P2Ready {
			state = room.InGame
		} else if info.Player1 == "" || info.Player2 == "" {
			state = room.Available
		} else {
			state = room.Full
		}
		values = append(values, RoomInfo{Id: tp.Key, Name: info.Name, Player1: info.Player1, Player2: info.Player2, State: state})
	}
	return
}
