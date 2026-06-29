package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Notifier struct {
	cfg *Config
}

func NewNotifier(cfg *Config) *Notifier {
	return &Notifier{cfg: cfg}
}

func (n *Notifier) Send(cmd string, startTime, endTime time.Time, exitCode int, ip string) {
	msg := fmt.Sprintf("CMD:   %s\nStart: %s\nEnd:   %s\nCode:  %d\nIP:    %s",
		cmd, startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"), exitCode, ip)

	var wg sync.WaitGroup

	if n.cfg.MQTT.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			debugf("正在通过 MQTT 发送通知...")
			if err := n.sendMQTT(msg); err != nil {
				log.Printf("MQTT 通知失败: %v", err)
			} else {
				debugf("MQTT 通知发送成功")
			}
		}()
	}

	if n.cfg.DingTalk.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			debugf("正在通过钉钉发送通知...")
			if err := n.sendDingTalk(msg); err != nil {
				log.Printf("钉钉通知失败: %v", err)
			} else {
				debugf("钉钉通知发送成功")
			}
		}()
	}

	if n.cfg.Ntfy.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			debugf("正在通过 ntfy 发送通知...")
			if err := n.sendNtfy(msg); err != nil {
				log.Printf("ntfy 通知失败: %v", err)
			} else {
				debugf("ntfy 通知发送成功")
			}
		}()
	}

	wg.Wait()
}

func (n *Notifier) sendMQTT(msg string) error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(n.cfg.MQTT.Broker)
	opts.SetClientID(fmt.Sprintf("run-cmd-%d", time.Now().UnixNano()))
	opts.SetUsername(n.cfg.MQTT.Username)
	opts.SetPassword(n.cfg.MQTT.Password)
	opts.SetAutoReconnect(false)
	opts.SetConnectTimeout(15 * time.Second)
	opts.SetWriteTimeout(15 * time.Second)
	opts.SetCleanSession(true)

	// 如果是 wss 协议，跳过证书验证（常用于自签证书场景）
	opts.SetTLSConfig(&tls.Config{
		InsecureSkipVerify: true,
	})

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		debugf("MQTT 已连接到: %s", n.cfg.MQTT.Broker)
	})
	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		log.Printf("MQTT 连接断开: %v", err)
	})

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(20 * time.Second) {
		return fmt.Errorf("MQTT 连接超时 (20s)")
	}
	if token.Error() != nil {
		return fmt.Errorf("MQTT 连接失败: %w", token.Error())
	}

	token = client.Publish(n.cfg.MQTT.Topic, 1, false, msg)
	if !token.WaitTimeout(15 * time.Second) {
		client.Disconnect(1000)
		return fmt.Errorf("MQTT 发布超时 (15s)")
	}
	if token.Error() != nil {
		client.Disconnect(1000)
		return fmt.Errorf("MQTT 发布失败: %w", token.Error())
	}

	client.Disconnect(1000)
	return nil
}

func (n *Notifier) sendDingTalk(msg string) error {
	body := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": msg,
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := http.Post(n.cfg.DingTalk.Webhook, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("钉钉 Webhook 返回非 200: %d", resp.StatusCode)
	}
	return nil
}

func (n *Notifier) sendNtfy(msg string) error {
	url := fmt.Sprintf("%s/%s", n.cfg.Ntfy.Server, n.cfg.Ntfy.Topic)
	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(msg)))
	if err != nil {
		return err
	}
	req.Header.Set("Title", "run-cmd")
	if n.cfg.Ntfy.Priority > 0 {
		req.Header.Set("Priority", fmt.Sprintf("%d", n.cfg.Ntfy.Priority))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ntfy 返回非 200: %d", resp.StatusCode)
	}
	return nil
}
