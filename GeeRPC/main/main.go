package main

import (
	"context"
	"geerpc"
	"log"
	"net"
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

func startServer(addr chan string) {
	// :0系统自动分配
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error: ", err)
	}
	log.Println("start rpc server on", lis.Addr())
	var foo Foo
	_ = geerpc.Register(foo)
	addr <- lis.Addr().String()
	geerpc.Accept(lis)
}

func main() {
	addr := make(chan string)
	go startServer(addr)

	client, _ := geerpc.Dial("tcp", <-addr)
	defer func() { _ = client.Close() }()
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := Args{
				Num1: i,
				Num2: i * 3,
			}
			var reply int
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			if err := client.CallWithCxt(ctx, "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Alice error:", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}
