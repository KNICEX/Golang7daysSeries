package singleflight

import (
	"sync"
)

type Loader interface {
	Do(key string, fn func() (any, error)) (any, error)
}

type blocker interface {
	block()
	through()
	set()
}

type call struct {
	blocker blocker
	val     any
	err     error
}

type syncBlocker struct {
	wg sync.WaitGroup
}

func (s *syncBlocker) block() {
	s.wg.Wait()
}

func (s *syncBlocker) through() {
	s.wg.Done()
}

func (s *syncBlocker) set() {
	s.wg.Add(1)
}

type chanBlocker struct {
	c chan struct{}
}

func (c *chanBlocker) block() {
	<-c.c
}

func (c *chanBlocker) through() {
	close(c.c)
}

func (c *chanBlocker) set() {
	c.c = make(chan struct{})
}

type group struct {
	mu          sync.Mutex
	m           map[string]*call
	callFactory func() *call
}

func NewWGLoader() Loader {
	return &group{
		callFactory: func() *call {
			return &call{blocker: &syncBlocker{}}
		},
	}
}
func NewChanLoader() Loader {
	return &group{
		callFactory: func() *call {
			return &call{blocker: &chanBlocker{}}
		},
	}
}

func (g *group) Do(key string, fn func() (any, error)) (any, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.blocker.block()
		return c.val, c.err
	}

	c := g.callFactory()
	c.blocker.set()
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.blocker.through()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
