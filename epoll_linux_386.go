package poller

type event struct {
	events uint32
	_      uint32
	data   *WaitPollable
}

func (e *event) setdata(p *WaitPollable) {
	e.data = p
}

func (e *event) getdata() *WaitPollable {
	return e.data
}
