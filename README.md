
# poller
    import "github.com/pkg/poller"

Package poller provides level triggered readiness notification and
reliable closing of file descriptors.







## type WaitPollable
``` go
type WaitPollable struct {
    // contains filtered or unexported fields
}
```
WaitPollable represents a file descriptor that can be read/written
and polled/waited for readiness notification.











### func (\*WaitPollable) Close
``` go
func (p *WaitPollable) Close() error
```
Close deregisters the WaitPollable and closes the underlying file descriptor.



### func (\*WaitPollable) Read
``` go
func (p *WaitPollable) Read(b []byte) (int, error)
```
Read reads up to len(b) bytes from the underlying fd. It returns the number of
bytes read and an error, if any. EOF is signaled by a zero count with
err set to io.EOF.

Callers to Read will block if there is no data available to read.



### func (\*WaitPollable) WaitRead
``` go
func (p *WaitPollable) WaitRead() error
```
WaitRead waits for the WaitPollable to become ready for
reading.



### func (\*WaitPollable) WaitWrite
``` go
func (p *WaitPollable) WaitWrite() error
```
WaitWrite waits for the WaitPollable to become ready for
writing.



### func (\*WaitPollable) Write
``` go
func (p *WaitPollable) Write(b []byte) (int, error)
```
Write writes len(b) bytes to the fd. It returns the number of bytes
written and an error, if any. Write returns a non-nil error when n !=
len(b).

Callers to Write will block if there is no buffer capacity available.



## type Poller
``` go
type Poller struct {
    // contains filtered or unexported fields
}
```
A Poller provides readiness notification and reliable closing of
registered file descriptors.









### func New
``` go
func New() (*Poller, error)
```
New creates a new Poller.




### func (\*Poller) Register
``` go
func (p *Poller) Register(fd uintptr) (*WaitPollable, error)
```
Register registers a file descriptor with the Poller and returns a
WaitPollable which can be used for reading/writing as well as readiness
notification.

File descriptors registered with the poller will be placed into
non-blocking mode.









- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
