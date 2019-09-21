package poller

type event struct {
	events uint32
	_      uint32
	data   *pollable
}

func (e *event) setdata(p *pollable) {
	e.data = p
}

func (e *event) getdata() *pollable {
	return e.data
}
