package main

type RoomInfo struct {
	Id               int
	Player1, Player2 string
	State            int
}

type RoomDetail struct {
	Player1, Player2 string
	P1Ready, P2Ready bool
}

type BoardInfo struct {
	Board  []int
	Turn   int
	Result int
}

type PlayerInfo struct {
	name   string
	roomId int
	tp     int
}
