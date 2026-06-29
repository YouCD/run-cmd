package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

func physicalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		name := iface.Name
		if strings.HasPrefix(name, "docker") || strings.HasPrefix(name, "veth") ||
			strings.HasPrefix(name, "br-") || strings.HasPrefix(name, "virbr") ||
			strings.HasPrefix(name, "lo") {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

var debugMode bool

func debugf(format string, v ...interface{}) {
	if debugMode {
		log.Printf("[DEBUG] "+format, v...)
	}
}

func main() {
	configPath := flag.String("c", "", "配置文件路径 (默认 ~/.run-cmd.yaml)")
	initCfg := flag.Bool("init", false, "生成默认配置文件")
	flag.BoolVar(&debugMode, "v", false, "开启 debug 日志")
	flag.Parse()

	if *initCfg {
		if err := InitConfig(*configPath); err != nil {
			log.Fatalf("初始化配置失败: %v", err)
		}
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("用法: run-cmd [选项] <命令>")
		fmt.Println()
		fmt.Println("选项:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("示例:")
		fmt.Println("  run-cmd \"sleep 10 && echo done\"")
		fmt.Println("  run-cmd -v \"make build\"")
		fmt.Println("  run-cmd -c /path/to/config.yaml \"long task\"")
		fmt.Println("  run-cmd --init")
		fmt.Println()
		fmt.Println("配置文件 ~/.run-cmd.yaml: MQTT 和钉钉机器人通知")
		os.Exit(1)
	}

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	cmdStr := strings.Join(args, " ")
	if strings.HasPrefix(cmdStr, "ls") {
		cmdStr = "ls --color=auto" + cmdStr[2:]
	}
	notifier := NewNotifier(cfg)

	debugf("开始执行: %s", cmdStr)
	start := time.Now()

	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	exitCode := 0
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
		debugf("命令执行失败，退出码: %d", exitCode)
	} else {
		debugf("命令执行成功")
	}

	end := time.Now()

	if cfg.MQTT.Enabled || cfg.DingTalk.Enabled {
		notifier.Send(cmdStr, start, end, exitCode, physicalIP())
	}
}
