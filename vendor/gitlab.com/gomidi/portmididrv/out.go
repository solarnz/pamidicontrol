package portmididrv

import (
	"fmt"
	"sync"

	"github.com/rakyll/portmidi"
	"gitlab.com/gomidi/midi"
)

func newOut(driver *driver, id portmidi.DeviceID, name string) midi.Out {
	return &out{driver: driver, id: id, name: name}
}

type out struct {
	id     portmidi.DeviceID
	stream *portmidi.Stream
	name   string
	driver *driver
	sync.RWMutex
}

// IsOpen returns, wether the port is open
func (o *out) IsOpen() bool {
	o.RLock()
	defer o.RUnlock()
	return o.stream != nil
}

// Write writes a MIDI message to the outut port
// If the output port is closed, it returns midi.ErrPortClosed
func (o *out) Write(b []byte) (int, error) {
	o.Lock()
	defer o.Unlock()
	if o.stream == nil {
		return 0, midi.ErrPortClosed
	}

	if len(b) < 2 {
		return 0, fmt.Errorf("cannot send less than two message bytes")
	}

	var last int64
	// ProgramChange messages only have 2 bytes
	if len(b) > 2 {
		last = int64(b[2])
	}

	err := o.stream.WriteShort(int64(b[0]), int64(b[1]), last)
	if err != nil {
		return 0, fmt.Errorf("could not send message to MIDI out %v (%s): %v", o.Number(), o, err)
	}
	return len(b), nil
}

// Underlying returns the underlying *portmidi.Stream. It will be nil, if the port is closed.
// Use it with type casting:
//   portOut := o.Underlying().(*portmidi.Stream)
func (o *out) Underlying() interface{} {
	return o.stream
}

// Number returns the number of the MIDI out port.
// Note that with portmidi, out and in ports are counted together.
// That means there should not be an out port with the same number as an in port.
func (o *out) Number() int {
	return int(o.id)
}

// String returns the name of the MIDI out port.
func (o *out) String() string {
	return o.name
}

// Close closes the MIDI out port
func (o *out) Close() error {
	o.Lock()
	defer o.Unlock()
	if o.stream == nil {
		return nil
	}

	err := o.stream.Close()
	if err != nil {
		return fmt.Errorf("can't close MIDI out %v (%s): %v", o.Number(), o, err)
	}
	o.stream = nil
	return nil
}

// Open opens the MIDI output port
func (o *out) Open() (err error) {
	o.Lock()
	defer o.Unlock()
	if o.stream != nil {
		return nil
	}

	o.stream, err = portmidi.NewOutputStream(o.id, o.driver.buffersizeOut, 0)
	if err != nil {
		o.stream = nil
		return fmt.Errorf("can't open MIDI out port %v (%s): %v", o.Number(), o, err)
	}
	o.driver.Lock()
	o.driver.opened = append(o.driver.opened, o)
	o.driver.Unlock()
	return nil
}
