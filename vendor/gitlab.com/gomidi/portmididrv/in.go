package portmididrv

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/rakyll/portmidi"
	"gitlab.com/gomidi/midi"
)

func newIn(driver *driver, id portmidi.DeviceID, name string) midi.In {
	return &in{driver: driver, id: id, name: name}
}

type in struct {
	id     portmidi.DeviceID
	stream *portmidi.Stream
	name   string

	driver *driver

	lastTimestamp portmidi.Timestamp
	mx            sync.RWMutex
	stopped       bool
}

// IsOpen returns wether the MIDI in port is open.
func (i *in) IsOpen() bool {
	i.mx.RLock()
	defer i.mx.RUnlock()
	return i.stream != nil
}

// Underlying returns the underlying *portmidi.Stream. It will be nil, if the port is closed.
// Use it with type casting:
//   portIn := i.Underlying().(*portmidi.Stream)
func (i *in) Underlying() interface{} {
	return i.stream
}

// Number returns the number of the MIDI in port.
// Note that with portmidi, out and in ports are counted together.
// That means there should not be an out port with the same number as an in port.
func (i *in) Number() int {
	return int(i.id)
}

// String returns the name of the MIDI in port.
func (i *in) String() string {
	return i.name
}

// Close closes the MIDI in port
func (i *in) Close() error {
	i.mx.Lock()
	defer i.mx.Unlock()
	if i.stream == nil {
		return nil
	}

	i.stopped = true

	err := i.stream.Close()
	if err != nil {
		return fmt.Errorf("can't close MIDI in %v (%s): %v", i.Number(), i, err)
	}
	i.stream = nil
	return nil
}

// Open opens the MIDI in port
func (i *in) Open() (err error) {
	i.mx.Lock()
	defer i.mx.Unlock()
	if i.stream != nil {
		return nil
	}
	i.stream, err = portmidi.NewInputStream(i.id, i.driver.buffersizeIn)
	if err != nil {
		i.stream = nil
		return fmt.Errorf("can't open MIDI in port %v (%s): %v", i.Number(), i, err)
	}
	i.driver.Lock()
	i.driver.opened = append(i.driver.opened, i)
	i.driver.Unlock()
	return nil
}

// StopListening cancels the listening
func (i *in) StopListening() error {
	i.mx.Lock()
	i.stopped = true
	i.mx.Unlock()
	return nil
}

// read is an internal helper function
func (i *in) read(cb func([]byte, int64)) error {
	events, err := i.stream.Read(i.driver.buffersizeRead)

	if err != nil {
		bt, err2 := i.stream.ReadSysExBytes(i.driver.buffersizeRead)
		if err2 != nil {
			return err
		}

		cb(bt, int64(i.lastTimestamp)*1000)
		return nil
	}

	for _, ev := range events {

		var b = make([]byte, 3)
		b[0] = byte(ev.Status)
		b[1] = byte(ev.Data1)
		b[2] = byte(ev.Data2)
		// ev.Timestamp is in Milliseconds
		// we want deltaMicroseconds as int64
		cb(b, int64(ev.Timestamp-i.lastTimestamp)*1000)
	}

	return nil
}

// SetListener sets the listener
func (i *in) SetListener(listener func(data []byte, deltaMicroseconds int64)) error {
	go func() {
		i.lastTimestamp = portmidi.Time()
		for {
			i.mx.Lock()
			stopped := i.stopped

			if stopped {
				i.mx.Unlock()
				return
			}
			has, _ := i.stream.Poll()
			if has {
				i.read(listener)
			}
			i.mx.Unlock()
			time.Sleep(i.driver.sleepingTime)
			runtime.Gosched()
		}
	}()
	return nil
}
