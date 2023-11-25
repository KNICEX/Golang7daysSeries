package main

import (
	"context"
	"geerpc"
	"geerpc/registry"
	"geerpc/xclient"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type Foo int
type Args struct {
	Num1, Num2 int
}

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

func startServer(registryAddr string, wg *sync.WaitGroup) {
	// :0系统自动分配
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error: ", err)
	}
	server := geerpc.NewServer()
	log.Println("start rpc server on", lis.Addr())
	var foo Foo
	_ = server.Register(foo)
	registry.Heartbeat(registryAddr, "tcp@"+lis.Addr().String(), 0)
	wg.Done()
	server.Accept(lis)
}

func startRegistry(wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", ":9999")
	registry.HandleHTTP()
	wg.Done()
	_ = http.Serve(l, nil)
}

func main() {
	log.SetFlags(0)
	registryAddr := "http://localhost:9999/_geerpc_/registry"
	var wg sync.WaitGroup
	wg.Add(1)
	go startRegistry(&wg)
	wg.Wait()
	time.Sleep(time.Second)
	wg.Add(2)
	go startServer(registryAddr, &wg)
	go startServer(registryAddr, &wg)

	wg.Wait()
	call(registryAddr)
	broadcast(registryAddr)
}

func foo(xc *xclient.XClient, ctx context.Context, typ, serviceMethod string, args *Args) {
	var reply int
	var err error
	switch typ {
	case "call":
		err = xc.CallWithCtx(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.BroadCast(ctx, serviceMethod, args, &reply)
	}

	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
	}
}

func call(registryAddr string) {
	d := xclient.NewGeeRegistryDiscovery(registryAddr, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() {
		_ = xc.Close()
	}()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := Args{
				Num1: i,
				Num2: i * 3,
			}
			time.Sleep(time.Second * time.Duration(i))
			foo(xc, context.Background(), "call", "Foo.Sum", &args)
		}(i)
	}
	wg.Wait()
}

func broadcast(registryAddr string) {
	d := xclient.NewGeeRegistryDiscovery(registryAddr, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() {
		_ = xc.Close()
	}()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := Args{
				Num1: i,
				Num2: i * 3,
			}

			foo(xc, context.Background(), "broadcast", "Foo.Sum", &args)
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			foo(xc, ctx, "broadcast", "Foo.Sleep", &args)
		}(i)
	}
	wg.Wait()
}
