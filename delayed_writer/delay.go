package delayed_writer

import (
	"fmt"
	"io"
	"time"
)

type delayedWriter struct {
	requests  int
	processed int
	delegate  io.Writer
	delay     time.Duration
	workItems chan workItem
	errors    chan error
}

var _ io.Writer = &delayedWriter{}

type workItem struct {
	data      []byte
	processAt time.Time
}

func New(w io.Writer, delayBy time.Duration) io.Writer {
	workItems := make(chan workItem, 300)
	self := &delayedWriter{
		delegate:  w,
		delay:     delayBy,
		workItems: workItems,
	}
	go self.processWorkItems()
	return self
}

func (d *delayedWriter) processWorkItems() {
	for item := range d.workItems {
		waitTime := item.processAt.Sub(time.Now())
		if waitTime > 0 {
			time.Sleep(waitTime)
		}
		_, err := d.delegate.Write(item.data)
		d.processed = d.processed + 1
		fmt.Printf("Req: %d Done: %d Pending: %d\n", d.requests, d.processed, d.requests-d.processed)
		if err != nil {
			d.errors <- err
		}
	}
}

func (d *delayedWriter) Write(p []byte) (n int, err error) {
	select {
	case err, ok := <-d.errors:
		if ok {
			//Something bad happened, so stop accepting work items.
			return 0, err
		}
	default:
		// no errors yet... keep accepting work items
	}
	data := make([]byte, len(p))
	copy(data, p)
	d.workItems <- workItem{
		data:      data,
		processAt: time.Now().Add(d.delay),
	}
	d.requests = d.requests + 1
	fmt.Printf("Req: %d Done: %d Pending: %d\n", d.requests, d.processed, d.requests-d.processed)
	return len(p), nil
}
