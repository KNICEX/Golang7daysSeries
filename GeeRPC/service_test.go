package geerpc

import (
	"fmt"
	"reflect"
	"testing"
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

func _assert(condition bool, msg string, v ...any) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}

func TestNewService(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	_assert(len(s.method) == 1, "wrong service Method, expect 1, but got %d", len(s.method))
	mType := s.method["Sum"]
	_assert(mType != nil, "wrong Method, Sum shouldn't nil")
}

func TestMethodType_Call(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	mType := s.method["Sum"]

	argv := mType.newArgv()
	replyv := mType.newReplyv()
	argv.Set(reflect.ValueOf(Args{Num1: 1, Num2: 3}))
	err := s.call(mType, argv, replyv)
	_assert(err == nil && *replyv.Interface().(*int) == 4 && mType.NumCalls() == 1,
		"failed to call Foo.Sum ")
}

func TestChan(t *testing.T) {
	c1 := make(chan struct{})
	c2 := make(chan struct{})
	go func() {
		fmt.Println("exec goroutine")
		time.Sleep(time.Second * 2)
		close(c1)
		fmt.Println("send message")
		close(c2)
		fmt.Println("end goroutine")
	}()

	select {
	case <-time.After(time.Second):
		fmt.Println("timeout")
	case <-c1:
		<-c2
	}
	time.Sleep(time.Second * 3)
	fmt.Println("end main")
}
