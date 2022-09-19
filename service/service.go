package service

import (
	"context"
	"distributed/registry"
	"fmt"
	"log"
	"net/http"
)

func Start(ctx context.Context, reg registry.Registration, host, port string,
	regsiterHandlersFunc func()) (context.Context, error) {
	regsiterHandlersFunc()
	//webService启动
	ctx = startService(ctx, reg.ServiceName, host, port)
	//注册服务
	err := registry.RegisterService(reg)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func startService(ctx context.Context, serviceName registry.ServiceName, host,
	port string) context.Context {

	ctx, cancel := context.WithCancel(ctx)

	var srv http.Server
	srv.Addr = ":" + port
	//发生错误时打印错误并取消
	go func() {
		log.Println(srv.ListenAndServe())
		err := registry.ShutdownService(fmt.Sprintf("http://%s:%s", host, port))
		if err != nil {
			log.Println(err)
		}
		cancel()
	}()
	//用户可以手动停止service
	go func() {
		fmt.Printf("%v start. Press any key to stop. \n", serviceName)
		var s string
		fmt.Scanln(&s)
		err := registry.ShutdownService(fmt.Sprintf("http://%s:%s", host, port))
		if err != nil {
			log.Println(err)
		}
		srv.Shutdown(ctx)
		cancel()
	}()
	return ctx
}
