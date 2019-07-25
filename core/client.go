package core

import (
	"bytes"
	"chat/configs"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"time"
)

const (
	// Время для написания сообщения коллеге
	writeWait = 10 * time.Second

	// Время на чтение сообщения от коллеги
	pongWait = 60 * time.Second

var (
	newLine = []byte{'\n'}
	space 	= []byte{' '}
)

// Клиент - посредник между websocket подключением и server
type Client struct {
	server *Server

	// the websocket connection
	conn *websocket.Conn

	// Буфферизированный канал для отправки сообщений
	send chan[]byte

	logger zap.Logger
}

// readPump - передает сообщения из websocket подключения серверу
//
// Приложение запускает readPump в отдельной горутине
// По каждому соединение будет реализован не более одного считывателя
// который будет читать из программы
func (c *Client) readPump() {
	l := c.logger.Named("readPump")
	//
	defer func() {
		c.server.unsubscribe <- c
		if err := c.conn.Close(); err != nil {
			l.Warn("Не удалось закрыть websocket подключение, from readPump!", zap.Error(err))
		}
	}()

	c.conn.SetReadLimit(configs.MaxMessageSize)

	if err := c.conn.SetReadDeadline(time.Now().Add(configs.PongWait)); err != nil {
		l.Warn("Не удалось установить крайний срок чтения сообщения", zap.Error(err))
	}

	c.conn.SetPongHandler(func(appData string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(configs.PongWait)); err != nil {
			l.Warn("Не удалось установить крайний срок чтения сообщения в pongHandler", zap.Error(err))
			return err
		}
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				l.Error("Не удалось получить сообщение ", zap.Error(err))
			}
			break
		}
		// убираем лишние пробелы, меняем символ переноса строки на пробел
		message = bytes.TrimSpace(bytes.Replace(message, newLine, space, -1))
		// передает сообщение в server
		c.server.broadcast <- message
	}
}

// writePump - принимает сообщения из сервера и передает их websocket подключению
//
// Приложение запускает writePump в отдельной горутине
// По каждому соединение будет реализован не более одной функции writePump
// который будет передавать все записи из горутины
func (c *Client) writePump() {
	l := c.logger.Named("writePump")
	ticker := time.NewTimer(configs.PingWait)

	defer func() {
		ticker.Stop()
		if err := c.conn.Close(); err != nil {
			l.Error("Не удалось закрыть websocket подключение, from writePump")
		}
	}()

	for {
		select {
		case message, ok := <- c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(configs.WriteWait)); err != nil {
				l.Warn("Не удалось установить крайний срок чтения сообщения в writePong", zap.Error(err))
			}
			if !ok {
				// сервер закрыл подключение
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Добавляем сообщение в очередь
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newLine)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				l.Warn("Не удалось закрыть writer ", zap.Error(err))
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(configs.WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				l.Warn("")
				return
			}
			
		}
	}
}

