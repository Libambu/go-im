package main

import (
	"net"
	"strings"
)

//用户连接对象
type User struct {
	Name   string
	Addr   string
	C      chan string
	coon   net.Conn
	Server *Server
}

//用户创建api
func NewUser(conn net.Conn, server *Server) *User {
	user := &User{
		Name:   conn.RemoteAddr().String(),
		Addr:   conn.RemoteAddr().String(),
		C:      make(chan string),
		coon:   conn,
		Server: server,
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

//用户的上线业务
func (u *User) login() {
	u.Server.mapLock.Lock()
	u.Server.OnlineMap[u.Name] = u
	u.Server.mapLock.Unlock()
	//用户上线信息广播
	u.Server.broadCast("has login", u)
}

//用户的下线业务
func (u *User) logout() {
	//用户下线用Map中删除
	u.Server.mapLock.Lock()
	delete(u.Server.OnlineMap, u.Name)
	u.Server.mapLock.Unlock()
	//用户上线信息广播
	u.Server.broadCast("has logout", u)
}

//用户处理消息业务
func (u *User) doMessage(msg string) {
	if msg == "who" {
		//用户查询有多少人在线
		for _, value := range u.Server.OnlineMap {
			msg := "user:" + value.Name + " is onlne"
			u.sendMsg(msg)
		}
	} else if len(msg) > 7 && msg[:7] == "rename" {
		//用户进行昵称的修改操作
		//消息格式 rename|newName
		newName := strings.Split(msg, "|")[1]
		if _, ok := u.Server.OnlineMap[newName]; ok {
			//用户昵称已经存在
			u.sendMsg("this name has exist")
			return
		}
		u.Server.mapLock.Lock()
		delete(u.Server.OnlineMap, u.Name)
		u.Server.OnlineMap[newName] = u
		u.Server.mapLock.Unlock()
		u.Name = newName
		u.sendMsg("change name:" + newName + " successful")
	} else if len(msg) > 2 && msg[:1] == "@" {
		//私聊 消息格式 @张三|你好啊我是……
		// 2. 使用 SplitN 切割，限制只切成 2 份
		// 这样如果消息内容里也有 "|" (比如 "@张三|你好|再见")，内容部分不会被切断
		parts := strings.SplitN(msg, "|", 2)

		// 3. 【关键】安全检查：防止数组越界
		// 如果没有 "|"，parts 的长度就是 1，直接访问 parts[1] 会崩溃
		if len(parts) < 2 {
			u.sendMsg("fmt err，use: @targetName|targetMsg")
			return
		}
		targetName := parts[0][1:]
		targetMsg := parts[1]
		// 5. 简单的非空检查
		if targetName == "" {
			u.sendMsg("targetName can not been null")
			return
		}
		if targetMsg == "" {
			u.sendMsg("targetMsg can not been null")
			return
		}
		//根据name获取User对象
		targetUser, ok := u.Server.OnlineMap[targetName]
		if !ok {
			u.sendMsg("this user is not exist")
			return
		}
		targetUser.sendMsg(u.Name + " send msg:" + targetMsg)
	} else {
		u.Server.broadCast(msg, u)
	}
}

func (u *User) sendMsg(msg string) {
	u.C <- msg
}
