// Package models provides models for servuces
package models

import "time"

type InputMessage struct {
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}

type OutputMessage struct {
	ChatID      string    `json:"chat_id"`
	Sender      ChatUser  `json:"sender"`
	Content     string    `json:"content"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
}

type ChatUser struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type Chat struct {
	ID      string     `json:"id"`
	Name    string     `json:"name"`
	Avatar  string     `json:"avatar"`
	Members []ChatUser `json:"members"`
}
