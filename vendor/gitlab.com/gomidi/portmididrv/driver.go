package portmididrv

import (
	"fmt"
	"sync"
	"time"

	"github.com/rakyll/portmidi"
	"gitlab.com/gomidi/midi"
)

type driver struct {
	buffersizeRead int
	buffersizeIn   int64
	buffersizeOut  int64
	sleepingTime   time.Duration
	opened         []midi.Port
	sync.RWMutex
}

func (d *driver) String() string {
	return "portmididrv"
}

// Close closes all open ports. It must be called at the end of a session.
func (d *driver) Close() (err error) {
	d.Lock()
	defer d.Unlock()
	for _, p := range d.opened {
		err = p.Close()
	}
	d.opened = nil
	// return just the last error to allow closing the other ports.
	// to ensure that all ports have been closed, this function must
	// return nil anyways
	return
}

// New returns a new driver
func New(options ...Option) (midi.Driver, error) {
	err := portmidi.Initialize()
	if err != nil {
		return nil, fmt.Errorf("can't initialize portmidi: %v", err)
	}
	dr := &driver{}

	dr.buffersizeRead = 1024
	dr.buffersizeIn = 1024
	dr.buffersizeOut = 1024

	// sleepingTime of 0.1ms should be fine to prevent busy waiting
	// and still fast enough for performances
	//dr.sleepingTime = time.Nanosecond * 1000 * 100
	//dr.sleepingTime = time.Nanosecond * 1000 * 500
	dr.sleepingTime = time.Microsecond * 500

	for _, opt := range options {
		opt(dr)
	}

	return dr, nil
}

// Ins returns the available MIDI in ports
func (d *driver) Ins() (ins []midi.In, err error) {
	d.Lock()
	defer d.Unlock()
	for i := 0; i < portmidi.CountDevices(); i++ {
		info := portmidi.Info(portmidi.DeviceID(i))
		if info != nil && info.IsInputAvailable {
			ins = append(ins, newIn(d, portmidi.DeviceID(i), info.Name))
		}
	}
	return
}

// Outs returns the available MIDI out ports
func (d *driver) Outs() (outs []midi.Out, err error) {
	d.Lock()
	defer d.Unlock()
	for i := 0; i < portmidi.CountDevices(); i++ {
		info := portmidi.Info(portmidi.DeviceID(i))
		if info != nil && info.IsOutputAvailable {
			outs = append(outs, newOut(d, portmidi.DeviceID(i), info.Name))
		}
	}
	return
}
