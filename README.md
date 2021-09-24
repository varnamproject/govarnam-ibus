# IBus Engine For GoVarnam

An easy way to type Indian languages on GNU/Linux systems.

goibus - golang implementation of libibus

Thanks to [sarim](https://github.com/sarim/goibus) and [haunt98](https://github.com/haunt98/goibus) for developing goibus from which this IBus engine is developed.

## Installation

See instructions in website: https://varnamproject.github.io/download/linux

## Development

### Building

Install dependencies:

```bash
sudo apt install libgtk-3-dev libcairo2-dev libglib2.0-dev
```

For Ubuntu 18.04 & others with old GTK versions, [special build params](https://github.com/gotk3/gotk3/issues/693) are required:
```
go build -tags pango_1_42,gtk_3_22 .
```
You can achieve the above with `make ubuntu-18`. This build will be **usable in majority of GNU/Linux distributions**.

You can also compile with latest GTK version with :
```bash
go build .
```
or `make ubuntu-20`

### Setup

To just try it out:
```
go run . -standalone
```

To use it system wide:
```
make ubuntu-18
sudo ln -s $(realpath varnam-ibus-engine) /usr/local/bin/varnam-ibus-engine
./varnam-ibus-engine -xml govarnam.xml
sudo ln -s $(realpath govarnam.xml) /usr/share/ibus/component/govarnam.xml

# Copy icon
sudo ln -s $(realpath varnam.png) /usr/local/share/varnam/ibus/icons/varnam.png

# Restart ibus
ibus restart
```

Now, go to ibus settings to add the input method. Currently it's just for Malayalam.

(Note that we use `varnam-ibus-engine` as executable name)
