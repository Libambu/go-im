package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

//创建一个server结构体
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
	}
	return server
}

//hander协程
func (s *Server) hander(conn net.Conn) {
	fmt.Println("accept success")
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
