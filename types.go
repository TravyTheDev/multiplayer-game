package main

import "github.com/gorilla/websocket"

type Actions struct {
	UserID           int     `json:"id"`
	MoveX            float64 `json:"moveX"`
	MoveY            float64 `json:"moveY"`
	IsAttack         bool    `json:"isAttack"`
	IsAttackInactive bool    `json:"isAttackInactive"`
	IsBlock          bool    `json:"isBlock"`
	IsHit            bool    `json:"isHit"`
}

type Client struct {
	Conn    *websocket.Conn
	Actions chan *Actions
	ID      int `json:"id"`
	X       float64
	Y       float64
}

type Hub struct {
	Clients    map[int]*Client `json:"clients"`
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *Actions
}

type ClientRes struct {
	ID int `json:"id"`
}
