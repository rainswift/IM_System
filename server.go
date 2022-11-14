package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	//在线用户列表
	OnlineMap map[string]*User
	//同步锁
	mapLock sync.RWMutex
	//消息广播的channer
	Message chan string
}

//创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

//监听message广播消息channel的goroutine，一旦有消息就发送给全部的在线user
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

//广播消息的方法
func (this *Server) BoradCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	user := NewUser(conn)
	//用户上线
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()
	//广播当前用户上线消息
	this.BoradCast(user, "以上线")

	//阻塞当前handler
	select {}
}

//启动服务器的接口
func (this *Server) Start() {
	Listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.listen err:", err)
		return
	}
	defer Listener.Close()
	//启动监听
	go this.ListenMessager()

	for {
		conn, err := Listener.Accept()
		if err != nil {
			fmt.Println("listenr accept err", err)
			continue
		}

		go this.Handler(conn)
	}
}
