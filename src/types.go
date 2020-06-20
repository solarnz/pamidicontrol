package pamidicontrol

type MidiActionType string

const (
	ControlChange MidiActionType = "ControlChange"
)

type PulseAudioActionType string

const (
	VolumeChange PulseAudioActionType = "VolumeChange"
	Mute                              = "Mute"
)

type PulseAudioTargetType string

const (
	PlaybackStream PulseAudioTargetType = "PlaybackStream"
	RecordStream                        = "RecordStream"
	Sink                                = "Sink"
	Source                              = "Source"
)

type PulseAudioAction struct {
	TargetType PulseAudioTargetType
	TargetName string

	ActionType PulseAudioActionType
}

type MidiAction struct {
	ActionType MidiActionType

	Channel    uint8
	Controller uint8

	MaxInputValue uint

	Action PulseAudioAction
}

type Config struct {
	MidiActions    []MidiAction
	InputMidiName  string
	OutputMidiName string
}
