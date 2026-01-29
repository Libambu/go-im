package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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
	user := NewUser(conn, s)
	fmt.Println("user:" + user.Name + " has login")
	//将用户上线功能给user
	user.login()
	//检测用户是否活跃的库
	isLive := make(chan bool)
	//接受用户发的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			len, err := conn.Read(buf)
			if len == 0 {
				//s.broadCast("has logout", user)
				user.logout()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("messge get err:", err)
				return
			}
			msg := string(buf[:len])
			user.doMessage(msg)
			//用户任意消息代表当前用户是否活跃
			isLive <- true
		}
	}()
	//如果直接退出那么堆上的user和user的监听携程将无人管理
	for {
		select {
		case <-isLive:
			//用户存活，应该重置定时器
		//time.After(duration) 的本质：它会创建一个新的定时器，并返回一个通道（channel）。
		//如果你不读它，在这个时间段后，系统会往这个通道里发一个当前时间。
		case <-time.After(time.Second * 10):
			//用户超时踢出
			user.sendMsg("you has been logout")
			close(user.C)
			conn.Close()
			delete(s.OnlineMap, user.Name)
			/*
				一旦 socket 被主协程 Close() 了，conn.Read 会立刻收到一个错误（通常是 use of closed network connection）。
				子协程收到 error，进入 if err != nil 分支。
				子协程执行 return。
			*/
			return
		}
	}
}

//为Server创建一个监听函数
func (s *Server) start() {
	//
	fmt.Println("listen begin")
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
