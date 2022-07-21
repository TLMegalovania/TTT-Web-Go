package main

import (
	"errors"
	"log"
	"strconv"
	"ttt/carr"
	"ttt/piece"
	"ttt/player"
	"ttt/win"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/philippseith/signalr"
)

type GameHub struct {
	players *cmap.ConcurrentMap[PlayerInfo]
	rooms   *carr.CArray[RoomDetail]
	boards  *carr.CArray[BoardInfo]

	signalr.Hub
}

func (hub *GameHub) OnConnected(connectionID string) {
	log.Printf("%s connected\n", connectionID)
}

func roomToHall(hub *GameHub, pin PlayerInfo) {
	connId := hub.ConnectionID()
	roomId, tp := pin.roomId, pin.tp
	srid := strconv.Itoa(roomId)
	pin.tp = player.None
	hub.players.Set(hub.ConnectionID(), pin)
	hub.Groups().RemoveFromGroup(srid, connId)
	hub.Groups().AddToGroup("hall", connId)
	if tp != player.Observer {
		setter := func(rmd *RoomDetail) {
			if tp == player.Host {
				rmd.Player1 = ""
				rmd.P1Ready = false
			} else if tp == player.Guest {
				rmd.Player2 = ""
				rmd.P2Ready = false
			}
		}

		hub.rooms.Set(roomId, setter)
		hub.Clients().Group(srid).Send("gotRoom", hub.rooms.Get(roomId))
		hub.Clients().Group("hall").Send("gotRooms", roomDetailsToInfo(hub.rooms))
	} else {
		hub.Clients().Caller().Send("gotRooms", roomDetailsToInfo(hub.rooms))
	}
}

func hallToOff(hub *GameHub, connId string) {
	hub.Groups().RemoveFromGroup("hall", connId)
}

func roomToOff(hub *GameHub, connId string, pin PlayerInfo) {
	roomId, tp := pin.roomId, pin.tp
	srid := strconv.Itoa(roomId)
	hub.Groups().RemoveFromGroup(srid, connId)
	if tp != player.Observer {
		setter := func(rmd *RoomDetail) {
			if tp == player.Host {
				rmd.Player1 = ""
				rmd.P1Ready = false
			} else if tp == player.Guest {
				rmd.Player2 = ""
				rmd.P2Ready = false
			}
		}

		hub.rooms.Set(roomId, setter)
		hub.Clients().Group(srid).Send("gotRoom", hub.rooms.Get(roomId))
		hub.Clients().Group("hall").Send("gotRooms", roomDetailsToInfo(hub.rooms))
	}
}

func gameToHall(hub *GameHub, pin PlayerInfo) {
	connId := hub.ConnectionID()
	roomId, tp := pin.roomId, pin.tp
	pin.tp = player.None
	hub.players.Set(connId, pin)
	srid := strconv.Itoa(roomId)
	hub.Groups().RemoveFromGroup(srid, connId)
	hub.Groups().AddToGroup("hall", connId)
	setter := func(rmd *RoomDetail) {
		if tp == player.Host {
			rmd.Player1 = ""
		} else if tp == player.Guest {
			rmd.Player2 = ""
		}
		rmd.P1Ready = false
		rmd.P2Ready = false
	}
	hub.rooms.Set(roomId, setter)
	hub.Clients().Group(srid).Send("gotRoom", hub.rooms.Get(roomId))
	hub.Clients().Group("hall").Send("gotRooms", roomDetailsToInfo(hub.rooms))

	bds := func(bd *BoardInfo) {
		bd.Result = win.Flee
	}
	hub.boards.Set(roomId, bds)
	hub.Clients().Group(srid).Send("gotBoard", hub.boards.Get(roomId))
}

func gameToOff(hub *GameHub, connId string, pin PlayerInfo) {
	roomId, tp := pin.roomId, pin.tp
	srid := strconv.Itoa(roomId)
	hub.Groups().RemoveFromGroup(srid, connId)
	setter := func(rmd *RoomDetail) {
		if tp == player.Host {
			rmd.Player1 = ""
		} else if tp == player.Guest {
			rmd.Player2 = ""
		}
		rmd.P1Ready = false
		rmd.P2Ready = false
	}
	hub.rooms.Set(roomId, setter)
	hub.Clients().Group(srid).Send("gotRoom", hub.rooms.Get(roomId))
	hub.Clients().Group("hall").Send("gotRooms", roomDetailsToInfo(hub.rooms))

	bds := func(bd *BoardInfo) {
		bd.Result = win.Flee
	}
	hub.boards.Set(roomId, bds)
	hub.Clients().Group(srid).Send("gotBoard", hub.boards.Get(roomId))

}

func (hub *GameHub) OnDisconnected(connectionID string) {
	log.Printf("%s disconnected\n", connectionID)
	pin, _ := hub.players.Pop(connectionID)
	tp := pin.tp
	if tp != player.None {
		if tp == player.Observer {
			roomToOff(hub, connectionID, pin)
		} else {
			rmd := hub.rooms.Get(pin.roomId)
			if rmd.P1Ready && rmd.P2Ready {
				gameToOff(hub, connectionID, pin)
			} else {
				roomToOff(hub, connectionID, pin)
			}
		}
	} else {
		hallToOff(hub, connectionID)
	}
}

func (hub *GameHub) Login(nickname string) error {
	if nickname == "" {
		return errors.New("empty name not allowed")
	}
	hub.players.Set(hub.ConnectionID(), PlayerInfo{name: nickname})

	hub.Clients().Caller().Send("gotRooms", roomDetailsToInfo(hub.rooms))
	hub.Groups().AddToGroup("hall", hub.ConnectionID())

	return nil
}

func (hub *GameHub) JoinRoom(roomId int) error {
	connId := hub.ConnectionID()
	pin, ok := hub.players.Get(connId)
	if !ok {
		return errors.New("unauthorized")
	}
	if roomId < 0 || roomId >= roomCount {
		return errors.New("no such room")
	}

	//leave from previous room if there is
	hub.LeaveRoom()

	srid := strconv.Itoa(roomId)
	hub.Groups().RemoveFromGroup("hall", connId)
	hub.Groups().AddToGroup(srid, connId)

	pin.roomId = roomId
	roomMod := true
	setter := func(rmd *RoomDetail) {
		if rmd.Player1 == "" {
			rmd.Player1 = pin.name
			pin.tp = player.Host
		} else if rmd.Player2 == "" {
			rmd.Player2 = pin.name
			pin.tp = player.Guest
		} else {
			pin.tp = player.Observer
			roomMod = false
		}
	}
	hub.rooms.Set(roomId, setter)
	hub.players.Set(connId, pin)
	hub.Clients().Caller().Send("joinedRoom", pin.tp)
	if roomMod {
		hub.Clients().Group("hall").Send("gotRooms", roomDetailsToInfo(hub.rooms))
		hub.Clients().Group(srid).Send("gotRoom", hub.rooms.Get(roomId))
	}
	hub.Clients().Caller().Send("gotBoard", hub.boards.Get(roomId))

	return nil
}

func (hub *GameHub) LeaveRoom() error {
	connId := hub.ConnectionID()
	pin, ok := hub.players.Get(connId)
	if !ok {
		return errors.New("unauthorized")
	}
	pt := pin.tp
	if pt == player.None {
		return errors.New("not in a room")
	}
	if pt == player.Observer {
		roomToHall(hub, pin)
	} else {
		roomId := pin.roomId
		rmd := hub.rooms.Get(roomId)
		if rmd.P1Ready && rmd.P2Ready {
			gameToHall(hub, pin)
		} else {
			roomToHall(hub, pin)
		}
	}
	return nil
}

func (hub *GameHub) StartGame() error {
	connId := hub.ConnectionID()
	pin, ok := hub.players.Get(connId)
	if !ok {
		return errors.New("unauthorized")
	}
	roomId, pt := pin.roomId, pin.tp
	if pt == player.None {
		return errors.New("not in a room")
	}
	if pt == player.Observer {
		return errors.New("not a player")
	}

	rmd := hub.rooms.Get(roomId)
	if rmd.P1Ready && rmd.P2Ready {
		return errors.New("game has started")
	}
	setter := func(rmd *RoomDetail) {
		if pt == player.Host {
			rmd.P1Ready = !rmd.P1Ready
		} else {
			rmd.P2Ready = !rmd.P2Ready
		}
	}
	hub.rooms.Set(roomId, setter)
	srid := strconv.Itoa(roomId)
	rmd = hub.rooms.Get(roomId)
	hub.Clients().Group(srid).Send("gotRoom", rmd)
	if rmd.P1Ready && rmd.P2Ready {
		setter := func(bd *BoardInfo) {
			bd.Board = make([]int, Row*Col)
			bd.Result = win.Null
			bd.Turn = piece.Black
		}
		hub.boards.Set(roomId, setter)
		hub.Clients().Group(srid).Send("gotBoard", hub.boards.Get(roomId))
	}
	return nil
}

func (hub *GameHub) Go(index int) error {
	connId := hub.ConnectionID()
	pin, ok := hub.players.Get(connId)
	if !ok {
		return errors.New("unauthorized")
	}
	pt := pin.tp
	if pt == player.None {
		return errors.New("not in a room")
	}
	if pt == player.Observer {
		return errors.New("not a player")
	}
	roomId := pin.roomId
	rmd := hub.rooms.Get(roomId)
	if !(rmd.P1Ready && rmd.P2Ready) {
		return errors.New("game not started")
	}
	bd := hub.boards.Get(roomId)
	bt := bd.Turn
	if bt == piece.Black && pt != player.Host || bt == piece.White && pt != player.Guest {
		return errors.New("not your turn")
	}
	if index < 0 || index >= Col*Row {
		return errors.New("position out of bounds")
	}
	if bd.Board[index] != piece.Null {
		return errors.New("there's already a piece")
	}
	bd.Board[index] = bd.Turn
	result := logic(bd, index)
	srid := strconv.Itoa(roomId)
	if result != win.Null {
		bd.Result = result
		setter := func(rmd *RoomDetail) {
			rmd.P1Ready = false
			rmd.P2Ready = false
		}
		hub.rooms.Set(roomId, setter)
		hub.Clients().Group(srid).Send("gotRoom", hub.rooms.Get(roomId))
	} else if bt == piece.Black {
		bd.Turn = piece.White
	} else {
		bd.Turn = piece.Black
	}
	hub.boards.Set(roomId, func(bin *BoardInfo) {
		bin.Board = bd.Board //maybe useless
		bin.Result = bd.Result
		bin.Turn = bd.Turn
	})
	hub.Clients().Group(srid).Send("gotBoard", bd)
	return nil
}
