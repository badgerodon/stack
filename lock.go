package main

import (
	"net"
	"strconv"
	"time"
)

type PortLock struct {
	hostport string
	ln       net.Listener
}

func NewPortLock(port int) PortLock {
	return PortLock{hostport: net.JoinHostPort("127.0.0.1", strconv.Itoa(port))}
}

func (p *PortLock) Lock() {
	t := 1 * time.Second
	for {
		if l, err := net.Listen("tcp", p.hostport); err == nil {
			p.ln = l
			return
		}
		time.Sleep(t)
	}
}

// Unlock unlocks the port lock
func (p *PortLock) Unlock() {
	if p.ln != nil {
		p.ln.Close()
	}
}
