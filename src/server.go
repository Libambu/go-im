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
	mapLock   sync.RWMutex
	//消息广播的channel
	MessageChan chan string
}

//创建一个server结构体
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:          ip,
		Port:        port,
		OnlineMap:   make(map[string]*User),
		MessageChan: make(chan string),
	}
	return server
}

//hander协程
func (s *Server) hander(conn net.Conn) {
	//	//用户上线，将用户加入Map表中
	user := NewUser(conn)
	fmt.Println("user:" + user.Name + " has login")
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()
	//用户上线信息广播
	s.broadCast("has online", user)
	//如果直接退出那么堆上的user和user的监听携程将无人管理
	select {}
}

//为Server创建一个监听函数
func (s *Server) start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	//关闭套接字
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("Listener close err:", err)
		}
	}(listener)

	//启动监听
	go s.listenMessageChan()

	//循环处理
	for {
		connect, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept() err:  ", err)
			continue
		}
		//创建处理协程
		go s.hander(connect)
	}
}

//Server的消息发送到MessageChan中
func (s *Server) broadCast(msg string, user *User) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.MessageChan <- sendMsg
}

//监听MessageChan广播消息，一旦有消息就发送给全部的在线User
func (s *Server) listenMessageChan() {
	for {
		msg := <-s.MessageChan
		s.mapLock.RLock()
		for _, value := range s.OnlineMap {
			value.C <- msg
		}
		s.mapLock.RUnlock()
	}
}
