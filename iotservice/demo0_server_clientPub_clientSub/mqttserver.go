package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	//接收 中断信号
	go func() {
		<-sigs
		done <- true
	}()

	//配置认证规则 配置可以访问server的用户和 ip 白名单
	authRules := &auth.Ledger{
		Auth: auth.AuthRules{ // Auth disallows all by default
			{Username: "melon", Password: "password2", Allow: true},
			{Remote: "127.0.0.1:*", Allow: true},
			{Remote: "localhost:*", Allow: true},
			{Remote: "192.168.1.2:*", Allow: true},
		},
		//配置 用户对主题的访问权限
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

	//启动mqtt server
	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	//收到 中断信号时 server进程 退出执行
	<-done
	server.Log.Warn("caught signal, stopping...")
	_ = server.Close()
	server.Log.Info("main.go finished")
}
