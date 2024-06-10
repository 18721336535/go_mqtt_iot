package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// 创建全局mqtt publish消息处理 handler
var messagePubHandler3 mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Println("pubclient 发布了消息：")
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	opts := mqtt.NewClientOptions().AddBroker("tcp://127.0.0.1:1883").SetClientID("mqttx_b6989f91xx")
	opts.SetUsername("melon")
	opts.SetPassword("password2")
	opts.SetKeepAlive(120 * time.Second)
	opts.SetDefaultPublishHandler(messagePubHandler3)
	opts.SetPingTimeout(10 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	i := 0
	for range time.Tick(time.Second * 3) {
		i++
		text := fmt.Sprintf("Hi Zeng #%d!", i)
		fmt.Println("\n pubclient 发布了消息：" + text)
		if token := c.Publish("melon/Wendu", 2, false, text); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			panic(token.Error())
		}
	}
	<-done
	os.Exit(1)
}