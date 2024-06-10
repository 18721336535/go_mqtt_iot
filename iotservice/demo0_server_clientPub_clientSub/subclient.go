package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// 创建全局mqtt sub消息处理 handler
var messageSubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("\n subclient 收到订阅消息： : %s \n", msg.Payload())
	//to do persist to db (redis)
}

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()
	opts := mqtt.NewClientOptions().AddBroker("tcp://127.0.0.1:1883").SetClientID("mqttx_b6989f91x2")
	opts.SetUsername("melon")
	opts.SetPassword("password2")
	opts.SetKeepAlive(120 * time.Second)
	opts.SetPingTimeout(10 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if token := c.Subscribe("melon/Wendu", 2, messageSubHandler); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		panic(token.Error())
	}
	<-done
	os.Exit(1)
}
