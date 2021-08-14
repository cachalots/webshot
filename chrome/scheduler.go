package chrome

import (
	"errors"
	"fmt"
	"github.com/4everland/screenshot/lib"
	"net/url"
	"sync"
	"time"
)

type Scheduler struct {
	Chrome  *Chrome
	Threads chan bool
}

type Task struct {
	ImageCh chan []byte
	Url     *url.URL
	EndTime time.Time
}

var (
	scheduler *Scheduler
	once      sync.Once
)

func NewScheduler(maxThread int, chrome *Chrome) *Scheduler {
	once.Do(func() {
		scheduler = &Scheduler{
			Chrome:  chrome,
			Threads: make(chan bool, maxThread),
		}
	})

	return scheduler
}

func (s *Scheduler) Exec(ch chan<- []byte, o ScreenshotOptions) {
	s.Threads <- true
	if o.EndTime.Before(time.Now()) {
		lib.Logger().Info(fmt.Sprintf("%s wait thread timeout, now: %d", o.URL.String(), time.Now().Unix()))
		<-s.Threads
		return
	}

	b := s.Chrome.Screenshot(o)
	if o.EndTime.After(time.Now()) {
		lib.Logger().Info(fmt.Sprintf("%s screenshot success, now: %d", o.URL.String(), time.Now().Unix()))
		ch <- b
	}

	close(ch)

	<-s.Threads
}

func Screenshot(o ScreenshotOptions) (b []byte, err error) {
	ch := make(chan []byte)
	lib.Logger().Info(fmt.Sprintf("%s request %d end %d", o.URL.String(), o.ReqTime.Unix(), o.EndTime.Unix()))
	go scheduler.Exec(ch, o)

	select {
	case b := <-ch:
		return b, nil
	case <-time.After(o.EndTime.Sub(time.Now())):
		lib.Logger().Info(fmt.Sprintf("%s channel select timeout endtime:%d now:%d",
			o.URL.String(), o.EndTime.Unix(), time.Now().Unix()))

		return b, errors.New("time out")
	}

	return
}
