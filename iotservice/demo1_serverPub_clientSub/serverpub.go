// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 mochi-mqtt, mochi-co
// SPDX-FileContributor: mochi-co

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mochi-mqtt/server/v2/hooks/auth"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
)

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

	// you may also find this useful...
	// d, _ := authRules.ToYAML()
	// d, _ := authRules.ToJSON()
	// fmt.Println(string(d))

	server := mqtt.New(&mqtt.Options{
		InlineClient: true, // you must enable inline client to use direct publishing and subscribing.
	})
	err := server.AddHook(new(auth.Hook), &auth.Options{
		Ledger: authRules,
	})
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

	// Demonstration of using an inline client to directly subscribe to a topic and receive a message when
	// that subscription is activated. The inline subscription method uses the same internal subscription logic
	// as used for external (normal) clients.
	go func() {
		// Inline subscriptions can also receive retained messages on subscription.
		_ = server.Publish("melon/retained1", []byte("Hi Andy! this is a retained message"), true, 0)
		_ = server.Publish("melon/retained2", []byte("Hello World! this is a retained message"), true, 0)
	}()

	// There is a shorthand convenience function, Publish, for easily sending  publish packets if you are not
	// concerned with creating your own packets.  If you want to have more control over your packets, you can
	//directly inject a packet of any kind into the broker. See examples/hooks/main.go for usage.
	go func() {
		for range time.Tick(time.Second * 10) {
			err := server.Publish("melon/retained3", []byte("Hi Zeng, this is not retained msg"), false, 0)
			if err != nil {
				server.Log.Error("server.Publish", "error", err)
			}
			server.Log.Info("main.go issued direct message to direct/publish")
		}
	}()

	<-done
	server.Log.Warn("caught signal, stopping...")
	_ = server.Close()
	server.Log.Info("main.go finished")
}
