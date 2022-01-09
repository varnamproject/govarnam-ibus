# IBus Engine For GoVarnam

An easy way to type Indian languages on GNU/Linux systems.

goibus - golang implementation of libibus

Thanks to [sarim](https://github.com/sarim/goibus) and [haunt98](https://github.com/haunt98/goibus) for developing goibus from which this IBus engine is developed.

## Installation

See instructions in website: https://varnamproject.github.io/download/linux

## Development

### Building

* Install dependencies:

```bash
sudo apt install libgtk-3-dev libcairo2-dev libglib2.0-dev
```

* Build

```
make ubuntu-14
```

The preferences dialog depends on GTK. The above command will build for **GTK 3.10** which means it'll work on Ubuntu 14.04 and later versions.

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
