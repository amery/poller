package poller

import (
	"io"
	"syscall"
)

// WaitPollable represents a file descriptor that can be read/written
// and polled/waited for readiness notification.
type WaitPollable struct {
	*pollable
	cr, cw chan error
}

func NewWaitPollable(poller *Poller, fd uintptr) (*WaitPollable, error) {
	p := &WaitPollable{
		cr: make(chan error),
		cw: make(chan error),
	}

	if pp, err := poller.RegisterHandler(fd, waitHandler, p); err != nil {
		return nil, err
	} else {
		p.pollable = pp
		return p, nil
	}
}

func (p *Poller) Register(fd uintptr) (*WaitPollable, error) {
	return NewWaitPollable(p, fd)
}

func waitHandler(fd uintptr, revents uint32, _p interface{}) {
	p := _p.(*WaitPollable)

	if revents&EPOLLOUT != 0 {
		p.wake('w', nil)
		revents &= ^EPOLLOUT
	}

	if revents&EPOLLIN != 0 {
		p.wake('r', nil)
		revents &= ^EPOLLIN
	}

	if revents != 0 {
		p.wake('r', EventError{fd, revents})
	}
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

// Close deregisters the pollable and closes the underlying file descriptor.
func (p *WaitPollable) Close() error {
	err := p.pollable.Close()
	close(p.cr)
	close(p.cw)
	return err
}

// WaitRead waits for the WaitPollable to become ready for
// reading.
func (p *WaitPollable) WaitRead() error {
	debug("pollable: %p, fd: %v  waitread", p, p.fd)
	if err := p.poller.waitRead(p.pollable); err != nil {
		return err
	}
	return <-p.cr
}

// WaitWrite waits for the WaitPollable to become ready for
// writing.
func (p *WaitPollable) WaitWrite() error {
	if err := p.poller.waitWrite(p.pollable); err != nil {
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
