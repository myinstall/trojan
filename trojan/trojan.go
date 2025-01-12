package trojan

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"trojan/core"
	"trojan/util"
)

// ControllMenu Trojan控制菜单
func ControllMenu() {
	fmt.Println()
	tType := Type()
	if tType == "trojan" {
		tType = "trojan-go"
	} else {
		tType = "trojan"
	}
	menu := []string{"启动程序", "停止程序", "重启程序", "查看状态", "查看log"}
	menu = append(menu, "切换为"+tType)
	switch util.LoopInput("请选择: ", menu, true) {
	case 1:
		Start()
	case 2:
		Stop()
	case 3:
		Restart()
	case 4:
		Status(true)
	case 5:
		go Log(300)
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		//阻塞
		<-c
	case 6:
		if err := SwitchType(tType); err != nil {
			fmt.Println(err)
		}
	}
}

// Restart 重启trojan
func Restart() {
	util.OpenPort(core.GetConfig().LocalPort)
	if err := util.ExecCommand("systemctl restart trojan"); err != nil {
		fmt.Println(util.Red("重启t失败!"))
	} else {
		fmt.Println(util.Green("重启t成功!"))
	}
}

// Start 启动trojan
func Start() {
	util.OpenPort(core.GetConfig().LocalPort)
	if err := util.ExecCommand("systemctl start trojan"); err != nil {
		fmt.Println(util.Red("启动t失败!"))
	} else {
		fmt.Println(util.Green("启动t成功!"))
	}
}

// Stop 停止trojan
func Stop() {
	if err := util.ExecCommand("systemctl stop trojan"); err != nil {
		fmt.Println(util.Red("停止t失败!"))
	} else {
		fmt.Println(util.Green("停止t成功!"))
	}
}

// Status 获取trojan状态
func Status(isPrint bool) string {
	result := util.ExecCommandWithResult("systemctl status trojan")
	if isPrint {
		fmt.Println(result)
	}
	return result
}

// RunTime Trojan运行时间
func RunTime() string {
	result := strings.TrimSpace(util.ExecCommandWithResult("ps -Ao etime,args|grep -v grep|grep /usr/local/etc/trojan/config.json"))
	resultSlice := strings.Split(result, " ")
	if len(resultSlice) > 0 {
		return resultSlice[0]
	}
	return ""
}

// Version Trojan版本
func Version() string {
	flag := "-v"
	if Type() == "trojan-go" {
		flag = "-version"
	}
	result := strings.TrimSpace(util.ExecCommandWithResult("/usr/bin/trojan/trojan " + flag))
	if len(result) == 0 {
		return ""
	}
	firstLine := strings.Split(result, "\n")[0]
	tempSlice := strings.Split(firstLine, " ")
	return tempSlice[len(tempSlice)-1]
}

// SwitchType 切换Trojan类型
func SwitchType(tType string) error {
	ARCH := runtime.GOARCH
	if ARCH != "amd64" && ARCH != "arm64" {
		return errors.New("not support " + ARCH + " machine")
	}
	if tType == "trojan" && ARCH != "amd64" {
		return errors.New("trojan not support " + ARCH + " machine")
	}
	if err := core.SetValue("trojanType", tType); err != nil {
		return err
	}
	InstallTrojan("")
	return nil
}

// Type Trojan类型
func Type() string {
	tType, _ := core.GetValue("trojanType")
	if tType == "" {
		if strings.Contains(Status(false), "trojan-go") {
			tType = "trojan-go"
		} else {
			tType = "trojan"
		}
		_ = core.SetValue("trojanType", tType)
	}
	return tType
}

// Log 实时打印trojan日志
func Log(line int) {
	result, _ := LogChan("-n "+strconv.Itoa(line), make(chan byte))
	for line := range result {
		fmt.Println(line)
	}
}

// LogChan trojan实时日志, 返回chan
func LogChan(param string, closeChan chan byte) (chan string, error) {
	cmd := exec.Command("bash", "-c", "journalctl -f -u trojan -o cat "+param)

	stdout, _ := cmd.StdoutPipe()

	if err := cmd.Start(); err != nil {
		fmt.Println("Error:The command is err: ", err.Error())
		return nil, err
	}
	ch := make(chan string, 100)
	stdoutScan := bufio.NewScanner(stdout)
	go func() {
		for stdoutScan.Scan() {
			select {
			case <-closeChan:
				stdout.Close()
				return
			default:
				ch <- stdoutScan.Text()
			}
		}
	}()
	return ch, nil
}
