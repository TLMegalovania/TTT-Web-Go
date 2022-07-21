package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"ttt/carr"

	kitlog "github.com/go-kit/log"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/philippseith/signalr"
)

const roomCount = 10

func main() {
	const address = "localhost:19810"
	players, rooms, boards := cmap.New[PlayerInfo](), carr.NewCArray[RoomDetail](roomCount), carr.NewCArray[BoardInfo](roomCount)
	server, err := signalr.NewServer(context.TODO(), signalr.HubFactory(func() signalr.HubInterface {
		return &GameHub{players: &players, rooms: rooms, boards: boards}
	}), signalr.Logger(kitlog.NewLogfmtLogger(os.Stdout), true), signalr.AllowOriginPatterns([]string{"localhost:*", "127.0.0.1:*"}))
	if err != nil {
		log.Fatal(err)
	}
	router := http.NewServeMux()
	server.MapHTTP(signalr.WithHTTPServeMux(router), "/ws")
	log.Printf("Listening for websocket connections on http://%s\n", address)
	if err := http.ListenAndServe(address, router); err != nil {
		log.Fatal(err)
	}
}
