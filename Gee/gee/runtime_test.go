package gee

import (
	"fmt"
	"runtime"
	"testing"
)

func TestDebug(t *testing.T) {
	fmt.Println("GOROOT --> ", runtime.GOROOT())
	fmt.Println("os/platform --> ", runtime.GOOS)
	fmt.Println("CPU numbers --> ", runtime.NumCPU())
	fmt.Println("mem state --> ", runtime.MemProfileRate)
}
