package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

var serverIp string
var serverPort int

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	coon       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}
	userConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net err:", err)
		return nil
	}
	client.coon = userConn
	return client
}

func (c *Client) nemu() bool {
	var flag int
	fmt.Println("input 1 all chat")
	fmt.Println("input 2 private chat")
	fmt.Println("input 3 rename")
	fmt.Println("input 0 exit")
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		//用户行为模式
		c.flag = flag
		return true
	} else {
		fmt.Println("input correct num")
		return false
	}
}

func (c *Client) publicChat() {
	var msg string
	fmt.Println(">>>input chat msg and input exit exit")
	fmt.Scanln(&msg)
	for msg != "exit" {
		if len(msg) != 0 {
			sendMsg := msg
			_, err := c.coon.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("send err:", err)
				break
			}
		}
		msg = ""
		fmt.Println(">>>input chat msg and input exit exit")
		fmt.Scanln(&msg)
	}

}

func (c *Client) selectUsers() {
	//查询哪些用户在线
	msg := "who"
	_, err := c.coon.Write([]byte(msg))
	if err != nil {
		fmt.Println("send err:", err)
		return
	}
}

func (c *Client) PrivateChat() {
	//查询哪些用户在线
	c.selectUsers()
	//请输入聊天对象的用户名
	fmt.Println(">>>请输入聊天对象的用户名,exit退出")
	var targetName string
	var targetMsg string
	fmt.Scanln(&targetName)
	for targetName != "exit" {
		fmt.Println(">>>请输入聊天消息")
		fmt.Scanln(&targetMsg)
		sendMsg := "@" + targetName + "|" + targetMsg
		_, err := c.coon.Write([]byte(sendMsg))
		if err != nil {
			fmt.Println("send err:", err)
			return
		}
		targetName = ""
		targetMsg = ""
		fmt.Println(">>>请输入聊天对象的用户名,exit退出")
		fmt.Scanln(&targetName)
	}
}

func (c *Client) updateName() bool {
	fmt.Println(">>>input you newName")
	fmt.Scanln(&c.Name)
	sendMsg := "rename" + "|" + c.Name
	_, err := c.coon.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("client write err:", err)
		return false
	}
	return true
}

//客户端主业务
func (c *Client) Run() {
	for c.flag != 0 {
		for c.nemu() != true {
		}
		switch c.flag {
		case 1:
			//公聊
			c.publicChat()
			break
		case 2:
			c.PrivateChat()
			//私聊
			break
		case 3:
			c.updateName()
			break
		}
		time.Sleep(time.Millisecond * 500)
	}
}

//读携程
func (c *Client) DealResponse() {
	buf := make([]byte, 1024)
	for {
		len, err := c.coon.Read(buf)
		// 【修复点 1】只要读到的长度为 0，就说明连接断了
		// 无论是正常断开还是被踢，n 都会是 0
		if len == 0 {
			fmt.Println("检测到服务器已断开连接，客户端退出...")

			// 【修复点 2】直接强行终止整个程序
			// 因为主协程可能还卡在 fmt.Scanln 等待输入，
			// 如果这里只 return，主协程不会停，程序就变成了“僵尸”
			os.Exit(0)
		}
		if err != nil && err != io.EOF {
			fmt.Println("messge get err:", err)
			return
		}
		msg := string(buf[:len])
		fmt.Println(msg)
	}
}

// .\Client -ip 127.0.01 -port 8888
//flag 的绑定代码只要在 flag.Parse() 执行之前 运行都可以
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "set server ip (default localhost)")
	flag.IntVar(&serverPort, "port", 8888, "set server port (default 8888)")
}

func main() {
	//命令行解析
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("client connect fail")
		return
	}
	//处理Server返回的消息
	go client.DealResponse()
	fmt.Println("client connect successful")
	//启动客户端
	client.Run()
}
