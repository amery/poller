package poller

// epoll(2) poller

import (
	"os"
	"syscall"
	"unsafe"
)

// New creates a new Poller.
func New() (*Poller, error) {
	p, err := newEpoll()
	return &Poller{poller: p}, err
}

// newEpoll returns an epoll(2) poller implementation.
func newEpoll() (poller, error) {
	fd, err := epollCreate1()
	if err != nil {
		return nil, err
	}
	e := epoll{
		pollfd:  fd,
		running: true,
		abort:   make(chan error),
		done:    make(chan error),
	}
	go e.loop()
	return &e, nil
}

// epoll implements an epoll(2) based poller.
type epoll struct {
	pollfd  uintptr
	running bool
	abort   chan error
	done    chan error
}

func (e *epoll) register(fd uintptr, h EventHandler, data interface{}) (*Pollable, error) {
	p := Pollable{
		fd:      fd,
		data:    data,
		handler: h,
		poller:  e,
	}
	ev := event{
		events: EPOLLERR | EPOLLHUP,
	}
	ev.setdata(&p)
	if err := epollctl(e.pollfd, syscall.EPOLL_CTL_ADD, fd, &ev); err != nil {
		return nil, err
	}
	debug("epoll: register: fd: %v, %p", p.fd, &p)
	return &p, nil
}

func (e *epoll) deregister(p *Pollable) error {
	// TODO(dfc) // wakeup all other waiters ?
	return epollctl(e.pollfd, syscall.EPOLL_CTL_DEL, p.fd, nil)
}

func (e *epoll) loop() {
	ch := make(chan event)
	go e.waiter(ch)

	defer syscall.Close(int(e.pollfd))

	for e.running {
		select {
		case ev, ok := <-ch:
			if ok {
				p := ev.getdata()
				p.handler(p.fd, ev.events, p.data)
			}
		case err, ok := <-e.abort:
			if ok {
				e.running = false
				e.done <- err
			}
		}
	}
}

func (e *epoll) waiter(out chan event) {
	var events []event
	var eventbuf [0x10]event

	const msec = -1

	for e.running {
		if len(events) > 0 {
			ev := events[0]
			events = events[1:]

			debug("epoll: wait: %0x, %p, fd: %v", ev.events, ev.getdata(), ev.getdata().fd)
			out <- ev
		} else if n, err := epollwait(e.pollfd, eventbuf[0:], msec); err != nil {
			debug("epoll: epollwait: %s", err)
			if err != syscall.EAGAIN && err != syscall.EINTR {
				err = os.NewSyscallError("epoll_wait", err)
				e.running = false
				e.done <- err
			}
		} else if n > 0 {
			debug("epoll: epollwait: %v", n)
			events = eventbuf[0:n]
		}
	}
}

func (e *epoll) WantEvents(p *Pollable, events uint32, oneshot bool) error {
	if events == 0 {
		events = EPOLLERR | EPOLLHUP
	} else if oneshot {
		events |= EPOLLONESHOT
	} else {
		events &= ^EPOLLONESHOT
	}

	if p.events != events {
		ev := event{
			events: events,
		}
		ev.setdata(p)
		debug("epoll: WantEvents: %d,  %0x, %p", p.fd, ev.events, ev.getdata())

		if err := epollctl(e.pollfd, syscall.EPOLL_CTL_MOD, p.fd, &ev); err != nil {
			return err
		}

		if oneshot {
			p.events = 0
		} else {
			p.events = events
		}
	}

	return nil
}

func (e *epoll) Abort(err error) error {
	e.abort <- err
	return nil
}

func (e *epoll) Done() error {
	return <-e.done
}

func epollCreate1() (uintptr, error) {
	fd, _, e := syscall.Syscall(syscall.SYS_EPOLL_CREATE1, syscall.EPOLL_CLOEXEC, 0, 0)
	if e != 0 {
		return 0, e
	}
	return fd, nil
}

// Single-word zero for use when we need a valid pointer to 0 bytes.
var _zero uintptr

func epollwait(epfd uintptr, events []event, msec int) (n int, err error) {
	var _p0 unsafe.Pointer
	if len(events) > 0 {
		_p0 = unsafe.Pointer(&events[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	r0, _, e1 := syscall.Syscall6(syscall.SYS_EPOLL_WAIT, epfd, uintptr(_p0), uintptr(len(events)), uintptr(msec), 0, 0)
	n = int(r0)
	if e1 != 0 {
		err = e1
	}
	return
}

func epollctl(epfd uintptr, op int, fd uintptr, event *event) (err error) {
	_, _, e1 := syscall.RawSyscall6(syscall.SYS_EPOLL_CTL, uintptr(epfd), uintptr(op), fd, uintptr(unsafe.Pointer(event)), 0, 0)
	if e1 != 0 {
		err = e1
	}
	return
}
