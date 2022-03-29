package main

import (
	"context"
	"fmt"
	"syscall"

	//_ "net/http/pprof"
	"os/exec"
	"time"
)

//当 Command 为 多个语句或者有后台进程时,在进行 kill 时可能会产生孤儿进程
func main() {
	// 使用 pprof 查看堆栈
	//go func() {
	//	err := http.ListenAndServe(":6060", nil)
	//	if err != nil {
	//		fmt.Printf("failed to start pprof monitor:%s", err)
	//	}
	//}()
	x := time.AfterFunc( //规定时间后自动提交
		time.Second*5,
		func() {
			println("cancelFn!!!!!!")
		},
	)
	defer x.Stop()
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelFn()
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", " echo hello;sleep 20; echo world")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	output, err := cmd.CombinedOutput()
	fmt.Printf("output：【%s】err:【%s】", string(output), err)
}
