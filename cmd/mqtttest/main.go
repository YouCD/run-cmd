// go run mqtt_test.go
package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	broker := "wss://mq-client.youcd.online/mqtt"
	username := "DJrVfQNN1yIpsjDo"
	password := "xI6I0Mrjdfma9B4r9Br8wV2wYhI9ALP2IioBuX8vFO89BZM"
	topic := "sms"

	// 尝试直接 TCP 连接（如果 MQTT 同时支持 TCP）
	tcpBroker := "tcp://mq-client.youcd.online:1883"

	fmt.Println("=== MQTT 连接测试 ===")
	fmt.Printf("Broker: %s\n", broker)
	fmt.Printf("TCP: %s\n", tcpBroker)
	fmt.Printf("User: %s\n", username)
	fmt.Printf("Topic: %s\n", topic)
	fmt.Println()

	for _, addr := range []string{broker, tcpBroker} {
		fmt.Printf("--- 尝试连接: %s ---\n", addr)
		testConnect(addr, username, password, topic)
		fmt.Println()
	}
}

func testConnect(broker, username, password, topic string) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(fmt.Sprintf("mqtt-test-%d", time.Now().UnixNano()))
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetConnectTimeout(10 * time.Second)
	opts.SetWriteTimeout(10 * time.Second)
	opts.SetCleanSession(true)
	opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		fmt.Println("  [OK] 连接成功 (OnConnect 回调触发)")
	})
	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		fmt.Printf("  [ERR] 连接断开: %v\n", err)
	})

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(15 * time.Second) {
		fmt.Println("  [ERR] 连接超时")
		return
	}
	if token.Error() != nil {
		fmt.Printf("  [ERR] 连接失败: %v\n", token.Error())
		return
	}

	fmt.Println("  [OK] Token 返回成功")

	msg := fmt.Sprintf("MQTT 测试消息 from run-cmd at %s", time.Now().Format("2006-01-02 15:04:05"))
	token = client.Publish(topic, 1, false, msg)
	if !token.WaitTimeout(10 * time.Second) {
		fmt.Println("  [ERR] 发布超时")
		client.Disconnect(500)
		return
	}
	if token.Error() != nil {
		fmt.Printf("  [ERR] 发布失败: %v\n", token.Error())
		client.Disconnect(500)
		return
	}

	fmt.Printf("  [OK] 发布成功: %s\n", msg)

	// 自己也订阅一下看是否能收到
	received := make(chan bool)
	subToken := client.Subscribe(topic, 1, func(c mqtt.Client, m mqtt.Message) {
		fmt.Printf("  [OK] 收到订阅消息: %s\n", string(m.Payload()))
		received <- true
	})
	if subToken.WaitTimeout(5*time.Second) && subToken.Error() == nil {
		fmt.Println("  [OK] 订阅成功")
	}

	time.Sleep(2 * time.Second)
	client.Disconnect(500)
	fmt.Println("  [OK] 已断开连接")
}

func init() {
	if v := os.Getenv("MQTT_BROKER"); v != "" {
		_ = v
	}
}
