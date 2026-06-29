# run-cmd

执行长时间命令，完成后通过 MQTT / 钉钉 / ntfy 发送通知。自动记录命令、起止时间、退出码和机器 IP。

## 安装

```bash
git clone <repo-url> && cd run-cmd && make install
# 或
cd run-cmd && go build -o run-cmd . && sudo make install
```

## 用法

```bash
run-cmd "sleep 60 && make build"
run-cmd -v "long-running-task"
run-cmd -c /path/to/config.yaml "command"
run-cmd --init
```

### 选项

| 选项 | 说明 |
|------|------|
| `-c string` | 配置文件路径 (默认 ~/.config/run_cmd/config.yaml) |
| `-v` | 开启 debug 日志 |
| `--init` | 生成默认配置文件 |

## 配置

`run-cmd --init` 生成配置文件于 `~/.config/run_cmd/config.yaml`：

```yaml
mqtt:
    enabled: false
    broker: wss://mq-client.youcd.online/mqtt
    topic: sms
    username: ""
    password: ""

dingtalk:
    enabled: false
    webhook: ""

ntfy:
    enabled: false
    server: "https://ntfy.sh"
    topic: ""
    priority: 3
```

环境变量 `RUN_CMD_CONFIG` 可指定自定义路径。

## 通知通道

| 通道 | 说明 |
|------|------|
| **MQTT** | 支持 wss/tcp 连接，可配合手机 App 订阅 |
| **钉钉** | 通过 Webhook 发送群机器人消息 |
| **ntfy** | 通过 [ntfy.sh](https://ntfy.sh) 发送推送通知 |

多个通道可同时启用，通知发送**并发执行**。

## 通知格式

```
CMD:   ls
Start: 2026-06-29 15:00:00
End:   2026-06-29 15:00:05
Code:  0
IP:    192.168.1.100
```

## 功能

- 命令输出实时显示到终端
- 自动对 `ls` 添加 `--color=auto` 保持颜色输出
- 多个通知通道并发推送
- 完成后发通知，不阻塞终端
- 日志分级（默认 info，-v 开启 debug）
