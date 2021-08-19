# IBus Engine For govarnam

An easy way to type Indian languages on GNU/Linux systems.

goibus - golang implementation of libibus

Thanks to [sarim](https://github.com/sarim/goibus) and [haunt98](https://github.com/haunt98/goibus) for developing goibus from which `govarnam-ibus` is developed.

## Installation

* Install and setup [IBus](https://wiki.archlinux.org/title/IBus)
* Install [GoVarnam](https://github.com/varnamproject/govarnam)
* Download a release from Releases
* Extract the zip file
* Run the install script in the extracted folder (need sudo):
```
sudo ./install.sh install
```
* Restart ibus (no sudo)
```bash
ibus restart
```
* Go to IBus settings, add Varnam input method.
* Maybe set an easy to use switch key to switch between languages

To uninstall:
```
sudo ./install.sh uninstall
```

## Development

## Building

For Ubuntu 18.04 & others with old GTK versions, special build params are required:
```
go build -tags pango_1_42,gtk_3_22 .
```

For others, simply do :

```bash
go build .
```

## Setup

To just try it out:
```
go run . -standalone
```

To use it system wide:
```
go build .
sudo ln -s $(realpath govarnam-ibus) /usr/local/bin/govarnam-ibus
./govarnam-ibus -xml govarnam.xml
sudo ln -s $(realpath govarnam.xml) /usr/share/ibus/component/govarnam.xml

# Copy icon
sudo ln -s $(realpath varnam.png) /usr/local/share/varnam/ibus/icons/varnam.png

# Restart ibus
ibus restart
```

Now, go to ibus settings to add the input method. Currently it's just for Malayalam.
