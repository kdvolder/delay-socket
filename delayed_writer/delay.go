package delayed_writer

import (
	"fmt"
	"io"
	"time"
)

type delayedWriter struct {
	count    int
	delegate io.Writer
	delay    time.Duration
}

func (d *delayedWriter) Write(p []byte) (n int, err error) {
	d.count = d.count + 1
	fmt.Printf("writes = %d\n", d.count)
	time.Sleep(d.delay)
	return d.delegate.Write(p)
}

func New(w io.Writer, delayBy time.Duration) io.Writer {
	return &delayedWriter{delegate: w, delay: delayBy}
}

var _ io.Writer = &delayedWriter{}
