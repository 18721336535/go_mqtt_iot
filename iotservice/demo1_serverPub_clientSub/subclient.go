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

// 创建全局mqtt sub消息处理 handler
var messageSubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Println("收到订阅消息：")
	fmt.Printf("Sub Client Topic : %s \n", msg.Topic())
	fmt.Printf("Sub Client msg : %s \n", msg.Payload())
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
	//opts := mqtt.NewClientOptions().AddBroker("tcp://broker.emqx.io:1883").SetClientID("emqx_test_client")
	// 创建消息订阅go程
	// go func() {
	// 	opts := mqtt.NewClientOptions().AddBroker("tcp://192.168.1.2:1883")
	// 	opts.SetClientID("mqttx_ec55ddbc")
	// 	opts.SetUsername("MQTT2")
	// 	opts.SetPassword("mqtt2123")
	// 	opts.SetKeepAlive(60 * time.Second)

	// 	client := mqtt.NewClient(opts)
	// 	if token := client.Connect(); token.Wait() && token.Error() != nil {
	// 		panic(token.Error())
	// 	}

	// 	if token := client.Subscribe("MQTT1/WenDu", 1, messageSubHandler); token.Wait() && token.Error() != nil {
	// 		fmt.Println(token.Error())
	// 		os.Exit(1)
	// 	}
	// 	// }
	// 	fmt.Println("订阅客户端断开与broker的连接")
	// 	// client.Disconnect(250)
	// }()

	go func() {
		opts := mqtt.NewClientOptions().AddBroker("tcp://127.0.0.1:1883")
		opts.SetClientID("mqttx_ec55ddbc")
		opts.SetUsername("melon")
		opts.SetPassword("password2")
		opts.SetKeepAlive(60 * time.Second)

		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}

		if token := client.Subscribe("melon/#", 1, messageSubHandler); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
		<-done
		// client.Disconnect(250)
	}()

	<-done
}
