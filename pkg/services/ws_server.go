package services

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var (
	gWebSocketServer *WebSocketServer
)

type WebSocketServer struct {
	//ClientMap     *sync.Map
	MessageQueueChan chan *WebsocketMessage
	clientChan       chan *WebsocketClient
	con              websocket.Conn
	context          context.Context
	cancelFunc       context.CancelFunc
}

func NewWebSocketServer() *WebSocketServer {
	ch := make(chan *WebsocketMessage, 1000)
	ctx, cancelFunc := context.WithCancel(context.Background())

	return &WebSocketServer{
		MessageQueueChan: ch,
		context:          ctx,
		cancelFunc:       cancelFunc,
		clientChan:       make(chan *WebsocketClient,1000),
	}
}
func (w *WebSocketServer) AddClient(conn *websocket.Conn) {
	wsCli := NewWebsocketClient(conn, make(chan *WebsocketMessage, 1000))
	//w.ClientMap.Store(wsCli.ID, wsCli)
	logrus.Debugf("wsCli:%s connected\n", wsCli.ID)
	//w.clientChan <- wsCli
	go wsCli.ProxyRoutine()
	go wsCli.ReadRoutine()
	go wsCli.WriteRoutine()
}
func (w *WebSocketServer) GetClientChan() chan *WebsocketClient {
	return w.clientChan
}
func GetServer() *WebSocketServer {
	if gWebSocketServer == nil {
		gWebSocketServer = NewWebSocketServer()
	}
	return gWebSocketServer
}
