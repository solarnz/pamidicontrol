# pamidicontrol
A utility to control the volume of PulseAudio streams / sinks / sources with a midi device

This has been tested with a [KORG nanoKontrol2](https://www.korg.com/au/products/computergear/nanokontrol2/) on ArchLinux. The
nanoKontrol2 is an in-expensive USB midi device with 8 sliders, 8 knobs and 24 buttons, which is great for having more
fine-grained control over your audio setup!

# Installation

You will need the `portmidi` library installed. In Arch, install this with `pacman -S portmidi` under Debian Based systems, you will need to install `libportmidi-dev`

```
go get github.com/solarnz/pamidicontrol
```

# Configuration

pamidicontrol requires the use of a configuration file. Place the config file under `$HOME/.config/pamidicontrol/config.yaml`.
You can checkout the [example configuration file](https://github.com/solarnz/pamidicontrol/blob/master/config.yaml) to see how to configure pamidicontrol.
You must set a bare-minimum the Input and Output midi device names.

pamidicontrol will print to stderr all of the midi control messages it gets, so you can easily build up your configuration file iteratively.
