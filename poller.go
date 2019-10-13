// Package poller provides level triggered readiness notification and
// reliable closing of file descriptors.
package poller

import (
	"syscall"
)

const (
	ClosedFd = ^uintptr(0)
)

type EventHandler func(fd uintptr, revents uint32, data interface{})

type Pollable struct {
	fd      uintptr
	data    interface{}
	events  uint32
	handler EventHandler
	poller  poller
}

// A Poller provides readiness notification and reliable closing of
// registered file descriptors.
type Poller struct {
	poller
}

// RegisterHandler registers a file descriptor with the Poller and returns a
// Pollable which can be used for reading/writing as well as readiness
// notification.
//
// File descriptors registered with the poller will be placed into
// non-blocking mode.
func (p *Poller) RegisterHandler(fd uintptr, h EventHandler, data interface{}) (*Pollable, error) {
	if err := syscall.SetNonblock(int(fd), true); err != nil {
		return nil, err
	}
	return p.register(fd, h, data)
}

func (p *Pollable) Fd() uintptr {
	return p.fd
}

func (p *Pollable) Close() error {
	if fd := p.fd; fd != ClosedFd {
		p.poller.deregister(p)
		p.fd = ClosedFd
		return syscall.Close(int(fd))
	}
	return nil
}

type poller interface {
	register(fd uintptr, h EventHandler, data interface{}) (*Pollable, error)
	WantEvents(*Pollable, uint32, bool) error
	deregister(*Pollable) error
}
