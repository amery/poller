// Package poller provides level triggered readiness notification and
// reliable closing of file descriptors.
package poller

import (
	"io"
	"syscall"
)

// A Poller provides readiness notification and reliable closing of
// registered file descriptors.
type Poller struct {
	poller
}

// Register registers a file descriptor with the Poller and returns a
// pollable which can be used for reading/writing as well as readiness
// notification.
//
// File descriptors registered with the poller will be placed into
// non-blocking mode.
func (p *Poller) Register(fd uintptr) (*WaitPollable, error) {
	if err := syscall.SetNonblock(int(fd), true); err != nil {
		return nil, err
	}
	return p.register(fd)
}

// WaitPollable represents a file descriptor that can be read/written
// and polled/waited for readiness notification.
type WaitPollable struct {
	fd     uintptr
	cr, cw chan error
	poller
}

// Read reads up to len(b) bytes from the underlying fd. It returns the number of
// bytes read and an error, if any. EOF is signaled by a zero count with
// err set to io.EOF.
//
// Callers to Read will block if there is no data available to read.
func (p *WaitPollable) Read(b []byte) (int, error) {
	n, e := p.read(b)
	if n < 0 {
		n = 0
	}
	if n == 0 && len(b) > 0 && e == nil {
		return 0, io.EOF
	}
	if e != nil {
		return n, e
	}
	return n, nil
}

func (p *WaitPollable) read(b []byte) (int, error) {
	for {
		n, e := syscall.Read(int(p.fd), b)
		if e != syscall.EAGAIN {
			return n, e
		}
		if err := p.WaitRead(); err != nil {
			return 0, err
		}
	}
}

// Write writes len(b) bytes to the fd. It returns the number of bytes
// written and an error, if any. Write returns a non-nil error when n !=
// len(b).
//
// Callers to Write will block if there is no buffer capacity available.
func (p *WaitPollable) Write(b []byte) (int, error) {
	n, e := p.write(b)
	if n < 0 {
		n = 0
	}
	if n != len(b) {
		return n, io.ErrShortWrite
	}
	if e != nil {
		return n, e
	}
	return n, nil
}

func (p *WaitPollable) write(b []byte) (int, error) {
	for {
		// TODO(dfc) this is wrong
		n, e := syscall.Write(int(p.fd), b)
		if e != syscall.EAGAIN {
			return n, e
		}
		if err := p.WaitWrite(); err != nil {
			return 0, err
		}
	}
}

// Close deregisters the WaitPollable and closes the underlying file descriptor.
func (p *WaitPollable) Close() error {
	err := p.deregister(p)
	// p.fd = uintptr(-1) // TODO(dfc) fix
	return err
}

// WaitRead waits for the WaitPollable to become ready for
// reading.
func (p *WaitPollable) WaitRead() error {
	debug("pollable: %p, fd: %v  waitread", p, p.fd)
	if err := p.poller.waitRead(p); err != nil {
		return err
	}
	return <-p.cr
}

// WaitWrite waits for the WaitPollable to become ready for
// writing.
func (p *WaitPollable) WaitWrite() error {
	if err := p.poller.waitWrite(p); err != nil {
		return err
	}
	return <-p.cw
}

func (p *WaitPollable) wake(mode int, err error) {
	debug("pollable: %p, fd: %v wake: %c, %v", p, p.fd, mode, err)
	if mode == 'r' {
		p.cr <- err
	} else {
		p.cw <- err
	}
}

type poller interface {
	register(fd uintptr) (*WaitPollable, error)
	waitRead(*WaitPollable) error
	waitWrite(*WaitPollable) error
	deregister(*WaitPollable) error
}
