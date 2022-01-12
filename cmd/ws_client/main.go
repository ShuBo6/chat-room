package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	ws_chat "ws-chat/pkg/services"
)

func getClientIp() (string, error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return "", err
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}

		}
	}

	return "", errors.New("can not find the client ip address")

}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("输入用户名")
	fmt.Println("---------------------")
	userName, _ := reader.ReadBytes('\n')
	userName = bytes.TrimRight(userName, "\r\n")
	userName = bytes.TrimRight(userName, "\n")
	fmt.Println("输入服务端地址")
	fmt.Println("---------------------")
	serverIP, _ := reader.ReadBytes('\n')
	serverIP = bytes.TrimRight(serverIP, "\r\n")
	serverIP = bytes.TrimRight(serverIP, "\n")
	conn, _ := ws_chat.Dial(string(serverIP))
	wsCli := ws_chat.NewWebsocketClient(conn, make(chan *ws_chat.WebsocketMessage, 1000))
	go wsCli.ReadRoutine()
	go wsCli.WriteRoutine()
	ip, _ := getClientIp()
	fmt.Println("-----------开始聊天----------")
	for {
		fmt.Print("\n-> ")
		text, _ := reader.ReadBytes('\n')
		text = bytes.TrimRight(text, "\r\n")
		text = bytes.TrimRight(text, "\n")
		fmt.Print("\n")

		wsCli.Write(ws_chat.WebsocketMessage{UserName: string(userName), Ip: ip, ClientId: wsCli.ID, Data: string(text)})
	}

}
