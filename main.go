package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
)

func main(){
	http.HandleFunc("/wol", handleHttp)
	er := http.ListenAndServe(":2333", nil)
	if er != nil{
		log.Fatal(er)
	}
}

func makeOnLan(mac, nic string) error{
	hw := strings.Replace(strings.Replace(mac, ":", "", -1), "-", "", -1)
	if len(hw) != 12 {
		return errors.New("hw != 12")
	}

	macHex, er := hex.DecodeString(hw)
	if er != nil {
		return er
	}

	var bcast = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	var buf bytes.Buffer
	buf.Write(bcast)
	for i := 0; i < 16; i++ {
		buf.Write(macHex)
	}

	mp := buf.Bytes()
	if len(mp) != 102 {
		return errors.New("mp != 102")
	}

	send := net.UDPAddr{}
	if len(nic) != 0 {
		inter, err := net.InterfaceByName(nic)
		if err != nil {
			return err
		}
		if (inter.Flags & net.FlagUp) == 0 {
			return errors.New("网卡未工作")
		}
		addrs, err := inter.Addrs()
		if err != nil {
			return err
		}
		isOK := false
		for _, addr := range addrs {
			if ip, ok := addr.(*net.IPNet); ok {
				if ipv4 := ip.IP.To4(); ipv4 != nil {
					isOK = true
					send.IP = ipv4
					break
				}
			}
		}
		if !isOK {
			return errors.New("未找到网卡所绑定的ip")
		}

		target := net.UDPAddr{IP: net.IPv4bcast}
		conn, er := net.DialUDP("udp", &send, &target)
		if er != nil {
			return er
		}
		defer conn.Close()

		_, err = conn.Write(mp)
		if err != nil {
			return errors.New("发送失败:" + err.Error())
		} else {
			return nil
		}
	}

	return errors.New("nic == 0")
}

func handleHttp(w http.ResponseWriter, re *http.Request){
	er := re.ParseForm()
	if er != nil{
		_, wer := w.Write([]byte("参数处理失败"))
		if wer != nil{
			log.Println("消息发送失败")
		}
		return
	}

	mac := re.Form.Get("mac")
	if mac == ""{
		_, wer := w.Write([]byte("无法寻找到mac参数"))
		if wer != nil{
			log.Println("发送失败")
		}
		return
	}
	nic := re.Form.Get("nic")
	var mer error
	if nic == ""{
		mer = makeOnLan(mac, "eth0")
	}else{
		mer = makeOnLan(mac, nic)
	}

	if mer != nil{
		_, wer := w.Write([]byte(mer.Error()))
		if wer != nil{
			log.Println("发送失败")
		}
	}else{
		_, wer := w.Write([]byte("已发送唤醒指令"))
		if wer != nil{
			log.Println("发送失败")
		}
	}
}
