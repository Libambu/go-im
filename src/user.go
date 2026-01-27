package main

import (
	"net"
)

//用户连接对象
type User struct {
	Name string
	Addr string
	C    chan string
	coon net.Conn
}

//用户创建api
func NewUser(conn net.Conn) *User {
	user := &User{
		Name: conn.RemoteAddr().String(),
		Addr: conn.RemoteAddr().String(),
		C:    make(chan string),
		coon: conn,
	}
	//启动User发送消息go程
	go user.listenChan()
	return user
}

//监听当前User channel 一旦有消息，就直接发给对方客户端
func (u *User) listenChan() {
	for {
		msg := <-u.C
		//将消息写给对应的客户端
		u.coon.Write([]byte(msg + "\r\n"))
	}
}
