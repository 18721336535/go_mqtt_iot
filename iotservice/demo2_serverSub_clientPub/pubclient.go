package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// 创建全局mqtt publish消息处理 handler
var messagePubHandler3 mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Println("发布消息：")
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

	mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stdout, "", 0)
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
	for range time.Tick(time.Second * 10) {
		i++
		text := fmt.Sprintf("Hi Zeng #%d!", i)
		if token := c.Publish("melon/Wendu", 0, false, text); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
	}
	<-done
}
