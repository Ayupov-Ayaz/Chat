package configs

import (
	"github.com/gorilla/websocket"
	"time"
)

const (
	// Время для написания сообщения коллеге
	WriteWait = 10 * time.Second
	// Время на чтение сообщения от коллеги
	PongWait = 60 * time.Second
	// Отправка ping запроса для проверки доступности коллеги, это значение должно быть меньше, чем PongWait
	PingWait = (PongWait * 9 ) / 10
	// Максимальный размер сообщения доступный для получения от коллеги
	MaxMessageSize = 512

	ReadBufferSize = 1024

	WriteBufferSize = 1024
)


var WsUpgrader = websocket.Upgrader{
	ReadBufferSize: ReadBufferSize,
	WriteBufferSize: WriteBufferSize,
}