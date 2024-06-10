package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mochi-mqtt/server/v2/hooks/auth"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
)

var TMP_HSET_KEY string = "/melon/temp"

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	authRules := &auth.Ledger{
		Auth: auth.AuthRules{ // Auth disallows all by default
			{Username: "peach", Password: "password1", Allow: true},
			{Username: "melon", Password: "password2", Allow: true},
			{Remote: "127.0.0.1:*", Allow: true},
			{Remote: "localhost:*", Allow: true},
			{Remote: "192.168.1.6:*", Allow: true},
			{Remote: "192.168.1.7:*", Allow: true},
		},
		ACL: auth.ACLRules{ // ACL allows all by default
			{Remote: "127.0.0.1:*"}, // local superuser allow all
			{
				// user melon can read and write to their own topic
				Username: "melon", Filters: auth.Filters{
					"melon/#":   auth.ReadWrite,
					"updates/#": auth.WriteOnly, // can write to updates, but can't read updates from others
				},
			},
			{
				// Otherwise, no clients have publishing permissions
				Filters: auth.Filters{
					"#":         auth.ReadOnly,
					"updates/#": auth.Deny,
				},
			},
		},
	}

	server := mqtt.New(&mqtt.Options{
		//must enable inline client to use direct publishing and subscribing.
		InlineClient: true,
	})
	err := server.AddHook(new(auth.Hook), &auth.Options{
		Ledger: authRules,
	})
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:      "t1",
		Address: ":1883",
	})
	err = server.AddListener(tcp)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {

		callbackFn := func(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
			server.Log.Info("服务器收到 订阅的消息：",
				"client", cl.ID,
				"subscriptionId", sub.Identifier,
				"topic", pk.TopicName,
				"订阅得到的新温度:", string(pk.Payload)+"°C")
		}

		server.Log.Info("订阅 1 : inline client subscribing")
		_ = server.Subscribe("melon/#", 2, callbackFn)

	}()

	<-done
	server.Log.Warn("caught signal, stopping...")
	_ = server.Close()
	server.Log.Info("main.go finished")
}
