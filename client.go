package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

func (c *Client) writeAction() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		action, ok := <-c.Actions
		if !ok {
			return
		}
		c.Conn.WriteJSON(action)
	}
}

func (c *Client) readAction(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, a, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		var action Actions
		if err := json.Unmarshal(a, &action); err != nil {
			log.Printf("error: %v", err)
		}
		act := &Actions{
			UserID:           action.UserID,
			MoveX:            action.MoveX,
			MoveY:            action.MoveY,
			IsAttack:         action.IsAttack,
			IsAttackInactive: action.IsAttackInactive,
			IsBlock:          action.IsBlock,
		}
		hub.Broadcast <- act
	}
}
