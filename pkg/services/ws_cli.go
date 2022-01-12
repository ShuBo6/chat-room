package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"sync"
)

type WebsocketMessage struct {
	ClientId string `json:"clientId"`
	UserName string `json:"userName"`
	Ip       string `json:"ip"`
	Data     string `json:"data"`
}

type WebsocketClient struct {
	ID         string
	Socket     *websocket.Conn
	In         chan *WebsocketMessage
	Out        chan []byte
	ctx        context.Context
	CancelFunc context.CancelFunc
	closed     bool
	mutex      *sync.Mutex
	closeChan  []chan bool
}

func Dial(ip string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: ip, Path: "/websocket"}
	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logrus.Warningf("websocket: dial:%s failed, error:%s", u.String(), err.Error())
		return nil, err
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		msg := fmt.Sprintf("websocket: dial:%s failed, status code:%d", u.String(), resp.StatusCode)
		logrus.Warningf(msg)
		return nil, errors.New(msg)
	}
	return conn, nil
}
func NewWebsocketClient(conn *websocket.Conn, readChan chan *WebsocketMessage) *WebsocketClient {
	//conn, _ := dial(ip)
	client := &WebsocketClient{Socket: conn, In: readChan, mutex: &sync.Mutex{}, closeChan: []chan bool{}}
	uuid := uuid.NewV4()
	client.ID = uuid.String()
	client.Out = make(chan []byte, 1000)
	client.ctx, client.CancelFunc = context.WithCancel(context.Background())
	return client
}

// return a chan which will be push 'true' when client closed
func (c *WebsocketClient) ClosedChan() chan bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	ch := make(chan bool, 1)
	c.closeChan = append(c.closeChan, ch)
	return ch
}

func (c *WebsocketClient) Write(message WebsocketMessage) {
	//c.mutex.Lock()
	//defer c.mutex.Unlock()

	if !c.Closed() {
		data, _ := json.Marshal(message)
		logrus.Debugf("write data %s", string(data))
		full := false
		if len(c.Out) == cap(c.Out) {
			logrus.Warningf("web socket client:%s writer data chan is full", c.ID)
			full = true
		}
		c.Out <- data
		if full {
			logrus.Warningf("write success after web socket client:%s writer data chan is full", c.ID)
		}
	}
}

func (c *WebsocketClient) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.closed {
		return
	}
	c.closed = true
	c.CancelFunc()
	//maybe close repeated, so ignore error
	c.Socket.Close()
	//close(c.Out)
	for _, ch := range c.closeChan {
		ch <- true
	}
	//TODO read all message from chan avoid write hang
	//ticker := time.NewTicker(time.Second)
	//for {
	//	select {
	//	case _, ok := <-c.Out:
	//		if !ok {
	//			return
	//		}
	//	case <-ticker.C:
	//		logrus.Debugf("close websocket client:%s success", c.ID)
	//		return
	//	}
	//}
}

func (c *WebsocketClient) Closed() bool {
	return c.closed
}

func (c *WebsocketClient) WriteRoutine() {
	logrus.Debugf("start client:%s write routine", c.ID)
	//ctx, _ := context.WithCancel(c.ctx)
	for {
		if c.Closed() {
			break
		}
		select {
		case message, ok := <-c.Out:
			//socket out chan closed only occur when client closed
			if !ok {
				err := c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					logrus.Warningf("close socket failed")
				}
				return
			}
			if message != nil {
				logrus.Debug(string(message))
				err := c.Socket.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					c.Close()
					logrus.Warningf("write data failed, error:%s, closed", err.Error())
					return
				}
			}
		case <-c.ctx.Done():
			logrus.Debugf("client:%s closed, exit write routine", c.ID)
			return
		}
	}
	logrus.Debugf("stop client:%s write routine", c.ID)
}

type PipelineRequest struct {
	Ids []string `json:"ids"`
}

func (c *WebsocketClient) ReadRoutine() {
	logrus.Debugf("start client:%s read routine", c.ID)
	for {
		if c.Closed() {
			break
		}
		messageType, data, err := c.Socket.ReadMessage()
		if err != nil {
			c.Close()
			logrus.Warningf("read message failed, error:%s, closed", err.Error())
			break
		}
		switch messageType {
		//upstream client close, closed server client
		case websocket.CloseMessage:
			logrus.Debugf("receive closed message, close client:%s", c.ID)
			c.Close()
			return
		case websocket.TextMessage:

			websocketMessage := WebsocketMessage{}
			if err = json.Unmarshal(data, &websocketMessage); err != nil {
				logrus.Warningf("read message failed, error:%s, closed", err.Error())
				continue
			}
			logrus.Debugf("receive TextMessage:%s", string(data))
			//fmt.Printf("\nreceive TextMessage:%s\n->", string(data))
			fmt.Printf("\n[%s-%s]:%s\n->", websocketMessage.UserName, websocketMessage.Ip, websocketMessage.Data)
			c.In <- &websocketMessage

			//fmt.Printf("\nreceive TextMessage:%+v\n->", websocketMessage)
		}
	}
	logrus.Debugf("stop client:%s read routine， +%v", c.ID, c.closed)
}
func (c *WebsocketClient) ProxyRoutine() {
	for {
		select {
		//case <-closeCh:
		//	return
		case <-c.ctx.Done():
			return
		case taskQ, ok := <-c.In:
			//fmt.Printf("\nWebsocketClient.ProxyRoutine taskQ:%+v\n", taskQ)
			if !ok {
				logrus.Warning("client task queue warning.")
			}
			if taskQ != nil {
				GetServer().MessageQueueChan <- taskQ
			}
		case message, ok := <-GetServer().MessageQueueChan:
			if !ok {
				logrus.Warning("client task queue warning.")
			}
			if message != nil {
				data, _ := json.Marshal(message)
				c.Out <- data
			}

		}
	}

}
