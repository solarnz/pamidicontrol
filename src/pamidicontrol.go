package pamidicontrol

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/sqp/pulseaudio"
)

func Run() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/pamidicontrol")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	var c Config
	err = viper.Unmarshal(&c)
	if err != nil {
		panic(err)
	}

	pulse, err := pulseaudio.New()
	if err != nil {
		panic(err)
	}

	paclient := NewPAClient(pulse)
	pulse.Register(paclient)
	paclient.RefreshStreams()

	midiClient := &MidiClient{
		PAClient:       paclient,
		MidiActions:    c.MidiActions,
		InputMidiName:  c.InputMidiName,
		OutputMidiName: c.OutputMidiName,
	}

	if c.InputMidiName == "" || c.OutputMidiName == "" {
		ins, outs, err := midiClient.ListDevices()
		if err != nil {
			panic(err)
		}

		log.Error().Msgf(
			"Input and Output Midi devices must be set.\nPossible input values are: \n\n%s\n\nPossible output values are\n\n%s",
			strings.Join(ins, "\n"),
			strings.Join(outs, "\n"),
		)
		os.Exit(1)
	}

	go midiClient.Run()

	pulse.Listen()
}
