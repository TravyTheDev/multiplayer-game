package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type WSHandler struct {
	hub *Hub
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWSHandler(hub *Hub) *WSHandler {
	return &WSHandler{
		hub: hub,
	}
}

func (h *WSHandler) registerRoutes(router *mux.Router) {
	router.HandleFunc("/ws/game/{user_id}", h.joinGame)
	router.HandleFunc("/game/isPlayer1", h.isPlayer1).Methods("GET")
}

func (h *WSHandler) joinGame(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error: %v", err)
	}
	defer conn.Close()
	userID := mux.Vars(r)["user_id"]

	var x = 0.0

	if len(h.hub.Clients) == 0 {
		x = 50
	} else {
		x = 150
	}
	intUserID, _ := strconv.Atoi(userID)
	cl := &Client{
		Conn:    conn,
		Actions: make(chan *Actions, 10),
		ID:      intUserID,
		X:       x,
		Y:       346,
	}
	act := &Actions{
		UserID:           intUserID,
		MoveX:            cl.X,
		MoveY:            cl.Y,
		IsAttack:         false,
		IsAttackInactive: false,
		IsBlock:          false,
	}
	h.hub.Register <- cl
	h.hub.Broadcast <- act

	go cl.writeAction()
	cl.readAction(h.hub)
}

func (h *WSHandler) isPlayer1(w http.ResponseWriter, r *http.Request) {
	clients := len(h.hub.Clients)

	var isPlayer1 = clients == 1

	if err := json.NewEncoder(w).Encode(isPlayer1); err != nil {
		fmt.Println(err)
		return
	}
}
