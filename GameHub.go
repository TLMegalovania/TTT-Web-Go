package main

import (
	"errors"
	"log"
	"ttt/piece"
	"ttt/player"
	"ttt/win"

	"github.com/google/uuid"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/philippseith/signalr"
)

type GameHub struct {
	players, rooms, boards *cmap.ConcurrentMap

	signalr.Hub
}

func (hub *GameHub) OnConnected(connectionID string) {
	log.Printf("%s connected\n", connectionID)
}

func leave(hub *GameHub, connId string, toHall bool) error {
	pi, ok := hub.players.Get(connId)
	if !ok {
		return errors.New("unauthorized")
	}
	pin := pi.(PlayerInfo)

	roomId, tp := pin.roomId, pin.tp
	if toHall {
		if roomId == "" {
			return errors.New("not in a room")
		}

		pin.roomId = ""
		pin.tp = player.None
		hub.players.Set(connId, pin)
		hub.Groups().RemoveFromGroup(roomId, connId)
		hub.Groups().AddToGroup("hall", connId)
	} else {
		hub.players.Remove(connId)
	}
	if roomId == "" || tp == player.Observer {
		return nil
	}

	rm, _ := hub.rooms.Get(roomId)
	rmd := rm.(RoomDetail)
	if pin.tp == player.Host {
		rmd.Player1 = ""
		rmd.P1Ready = false
	} else {
		rmd.Player2 = ""
		rmd.P2Ready = false
	}

	bd, ok := hub.boards.Get(roomId)
	if ok {
		bdi := bd.(BoardInfo)
		bdi.Result = win.Flee
		rmd.P1Ready = false
		rmd.P2Ready = false
		hub.rooms.Set(roomId, rmd)
		hub.Clients().Group(roomId).Send("gotRoom", rmd)

		hub.Clients().Group(roomId).Send("gotBoard", bdi)
		hub.boards.Remove(roomId)
	} else if rmd.Player1 == "" && rmd.Player2 == "" {
		hub.rooms.Remove(roomId)
		hub.Clients().Group(roomId).Send("deletedRoom")
	} else {
		hub.rooms.Set(roomId, rmd)
		hub.Clients().Group(roomId).Send("gotRoom", rmd)
	}

	hub.Clients().Group("hall").Send("gotRooms", roomDetailsToInfo(hub.rooms))

	return nil
}

func (hub *GameHub) OnDisconnected(connectionID string) {
	log.Printf("%s disconnected\n", connectionID)
	leave(hub, connectionID, false)
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

func (hub *GameHub) CreateRoom(name string) error {
	connId := hub.ConnectionID()
	pi, ok := hub.players.Get(connId)
	if !ok {
		return errors.New("unauthorized")
	}
	pin := pi.(PlayerInfo)
	roomId := uuid.NewString()
	pin.roomId = roomId
	pin.tp = player.Host
	hub.players.Set(connId, pin)
	hub.rooms.Set(roomId, RoomDetail{Name: name, Player1: pin.name})
	hub.Groups().RemoveFromGroup("hall", connId)
	hub.Groups().AddToGroup(roomId, connId)
	hub.Clients().Caller().Send("createdRoom", hub.ConnectionID())
	hub.Clients().Group("hall").Send("gotRooms", roomDetailsToInfo(hub.rooms))

	return nil
}

func (hub *GameHub) JoinRoom(id string) error {
	connId := hub.ConnectionID()
	pi, ok := hub.players.Get(connId)
	if !ok {
		return errors.New("unauthorized")
	}
	rm, ok := hub.rooms.Get(id)
	if !ok {
		hub.Clients().Caller().Send("joinedRoom", player.None)
		return nil
	}
	pin := pi.(PlayerInfo)
	rmd := rm.(RoomDetail)
	if pin.roomId == id && pin.tp != player.Observer {
		hub.Clients().Caller().Send("gotRoom", rmd)
		return nil
	}
	hub.LeaveRoom()

	pin.roomId = id
	roomMod := true
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
	hub.Clients().Caller().Send("joinedRoom", pin.tp)
	hub.players.Set(connId, pin)
	hub.Groups().RemoveFromGroup("hall", hub.ConnectionID())
	hub.Groups().AddToGroup(id, hub.ConnectionID())

	if roomMod {
		hub.rooms.Set(id, rmd)
		hub.Clients().Group("hall").Send("gotRooms", roomDetailsToInfo(hub.rooms))
		hub.Clients().Group(id).Send("gotRoom", rmd)
	} else if b, ok := hub.boards.Get(id); ok {
		hub.Clients().Caller().Send("gotBoard", b)
	}
	return nil
}

func (hub *GameHub) LeaveRoom() error {
	return leave(hub, hub.ConnectionID(), true)
}

func (hub *GameHub) StartGame() error {
	connId := hub.ConnectionID()
	pi, ok := hub.players.Get(connId)
	if !ok {
		return errors.New("unauthorized")
	}
	pin := pi.(PlayerInfo)
	roomId, pt := pin.roomId, pin.tp
	if roomId == "" {
		return errors.New("not in a room")
	}
	if pt == player.Observer {
		return errors.New("not a player")
	}

	rm, _ := hub.rooms.Get(roomId)
	rmd := rm.(RoomDetail)
	if rmd.P1Ready && rmd.P2Ready {
		return errors.New("game has started")
	}
	if pt == player.Host {
		rmd.P1Ready = !rmd.P1Ready
	} else {
		rmd.P2Ready = !rmd.P2Ready
	}
	hub.Clients().Group(roomId).Send("gotRoom", rmd)
	if rmd.P1Ready && rmd.P2Ready {
		bd := BoardInfo{Board: make([]int, Row*Col), Turn: piece.Black}
		hub.boards.Set(roomId, bd)
		hub.Clients().Group(roomId).Send("gotBoard", bd)
	}
	return nil
}

func (hub *GameHub) Go(index int) error {
	connId := hub.ConnectionID()
	pi, ok := hub.players.Get(connId)
	if !ok {
		return errors.New("unauthorized")
	}
	pin := pi.(PlayerInfo)
	roomId, pt := pin.roomId, pin.tp
	if roomId == "" {
		return errors.New("not in a room")
	}
	if pt == player.Observer {
		return errors.New("not a player")
	}

	b, ok := hub.boards.Get(roomId)
	if !ok {
		return errors.New("game not started")
	}
	bd := b.(BoardInfo)
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
	logic(&bd, index)
	if bd.Result != win.Null {
		hub.boards.Remove(roomId)
		rm, _ := hub.rooms.Get(roomId)
		rmd := rm.(RoomDetail)
		rmd.P1Ready = false
		rmd.P2Ready = false
		hub.boards.Set(roomId, rmd)
		hub.Clients().Group(roomId).Send("gotRoom", rmd)
	}
	hub.Clients().Group(roomId).Send("gotBoard", bd)
	return nil
}
