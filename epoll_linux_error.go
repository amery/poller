package poller

import (
	"fmt"
	"strings"
	"syscall"
)

const (
	EPOLLERR     = uint32(syscall.EPOLLERR)
	EPOLLHUP     = uint32(syscall.EPOLLHUP)
	EPOLLIN      = uint32(syscall.EPOLLIN)
	EPOLLOUT     = uint32(syscall.EPOLLOUT)
	EPOLLONESHOT = uint32(syscall.EPOLLONESHOT)
)

func EventNames(revents uint32) string {
	var s []string

	if revents == 0 {
		return ""
	}

	if revents&EPOLLERR != 0 {
		revents &= ^EPOLLERR
		s = append(s, "EPOLLERR")
	}

	if revents&EPOLLHUP != 0 {
		revents &= ^EPOLLHUP
		s = append(s, "EPOLLHUP")
	}

	if revents&EPOLLIN != 0 {
		revents &= ^EPOLLIN
		s = append(s, "EPOLLIN")
	}

	if revents&EPOLLOUT != 0 {
		revents &= ^EPOLLOUT
		s = append(s, "EPOLLOUT")
	}

	if revents&EPOLLONESHOT != 0 {
		revents &= ^EPOLLONESHOT
		s = append(s, "EPOLLONESHOT")
	}

	if revents != 0 {
		s = append(s, fmt.Sprintf("0x%x", revents))
	}

	return strings.Join(s, "|")
}

type EventError struct {
	fd      uintptr
	revents uint32
}

func NewEventError(fd uintptr, revents uint32) EventError {
	e := EventError{
		fd:      fd,
		revents: revents,
	}
	return e
}

func (e EventError) Error() string {
	return fmt.Sprintf("epoll: Unexpected event %s (0x%x) on fd:%d",
		EventNames(e.revents),
		e.revents,
		e.fd)
}

func (e EventError) Fd() uintptr {
	return e.fd
}

func (e EventError) Events() uint32 {
	return e.revents
}
