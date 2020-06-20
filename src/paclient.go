package pamidicontrol

import (
	"github.com/godbus/dbus"
	"github.com/rs/zerolog/log"
	"github.com/sqp/pulseaudio"
)

type PAClient struct {
	*pulseaudio.Client

	playbackStreamsByName map[string]dbus.ObjectPath
	recordStreamsByName   map[string]dbus.ObjectPath
	sourcesByName         map[string]dbus.ObjectPath
	sinksByName           map[string]dbus.ObjectPath
}

func NewPAClient(c *pulseaudio.Client) *PAClient {
	client := &PAClient{
		Client:                c,
		playbackStreamsByName: make(map[string]dbus.ObjectPath, 0),
		recordStreamsByName:   make(map[string]dbus.ObjectPath, 0),
		sourcesByName:         make(map[string]dbus.ObjectPath, 0),
		sinksByName:           make(map[string]dbus.ObjectPath, 0),
	}
	return client
}

func (c *PAClient) NewPlaybackStream(path dbus.ObjectPath) {
	c.RefreshStreams()
}

func (c *PAClient) PlaybackStreamRemoved(path dbus.ObjectPath) {
	c.RefreshStreams()
}

func (c *PAClient) RefreshStreams() error {
	playbackStreamsByName := make(map[string]dbus.ObjectPath, 0)
	recordStreamsByName := make(map[string]dbus.ObjectPath, 0)
	sinksByName := make(map[string]dbus.ObjectPath, 0)
	sourcesByName := make(map[string]dbus.ObjectPath, 0)

	streams, err := c.Core().ListPath("PlaybackStreams")
	if err != nil {
		return err
	}

	for _, streamPath := range streams {
		stream := c.Stream(streamPath)
		props, err := stream.MapString("PropertyList")
		if err != nil {
			return err
		}

		if applicationName, ok := props["application.name"]; ok {
			playbackStreamsByName[applicationName] = streamPath
		}
	}

	streams, err = c.Core().ListPath("RecordStreams")
	if err != nil {
		return err
	}

	for _, streamPath := range streams {
		stream := c.Stream(streamPath)
		props, err := stream.MapString("PropertyList")
		if err != nil {
			return err
		}

		if applicationName, ok := props["application.name"]; ok {
			recordStreamsByName[applicationName] = streamPath
		}
	}

	sinks, err := c.Core().ListPath("Sinks")
	if err != nil {
		return err
	}
	for _, sinkPath := range sinks {
		device := c.Device(sinkPath)
		props, err := device.MapString("PropertyList")
		if err != nil {
			panic(err)
		}

		if deviceDescription, ok := props["device.description"]; ok {
			sinksByName[deviceDescription] = sinkPath
		}
	}

	sources, err := c.Core().ListPath("Sources")
	if err != nil {
		return err
	}
	for _, sourcePath := range sources {
		device := c.Device(sourcePath)
		props, err := device.MapString("PropertyList")
		if err != nil {
			panic(err)
		}

		if deviceDescription, ok := props["device.description"]; ok {
			sourcesByName[deviceDescription] = sourcePath
		}
	}

	c.playbackStreamsByName = playbackStreamsByName
	c.recordStreamsByName = recordStreamsByName
	c.sinksByName = sinksByName
	c.sourcesByName = sourcesByName

	return nil
}

func (c *PAClient) ProcessVolumeAction(action PulseAudioAction, volume float32) error {
	pa100perc := 65535
	newVol := uint32(volume * float32(pa100perc))

	var obj *pulseaudio.Object

	if action.TargetType == Sink {
		if sinkPath, ok := c.sinksByName[action.TargetName]; ok {
			obj = c.Device(sinkPath)
		}
	}

	if action.TargetType == Source {
		if sourcePath, ok := c.sourcesByName[action.TargetName]; ok {
			obj = c.Device(sourcePath)
		}
	}

	if action.TargetType == PlaybackStream {
		if streamPath, ok := c.playbackStreamsByName[action.TargetName]; ok {
			obj = c.Stream(streamPath)
		}
	}

	if action.TargetType == RecordStream {
		if streamPath, ok := c.recordStreamsByName[action.TargetName]; ok {
			obj = c.Stream(streamPath)
		}
	}

	if obj != nil {
		err := obj.Set("Volume", []uint32{newVol, newVol})
		if err != nil {
			return err
		}
	} else {
		var paType string
		switch action.TargetType {
		case Sink:
			paType = "sink"
		case Source:
			paType = "source"
		case PlaybackStream:
			paType = "playback stream"
		case RecordStream:
			paType = "record stream"
		}

		log.Warn().Msgf("Could not find %s by name [%s] to set its volume", paType, action.TargetName)
	}
	return nil
}
