# IBus Engine For GoVarnam

An easy way to type Indian languages on GNU/Linux systems.

goibus - golang implementation of libibus

Thanks to [sarim](https://github.com/sarim/goibus) and [haunt98](https://github.com/haunt98/goibus) for developing goibus from which this IBus engine is developed.

## Installation

* Install and setup [IBus](https://wiki.archlinux.org/title/IBus)
* Install [GoVarnam](https://github.com/varnamproject/govarnam)
* Download the [latest release](https://github.com/varnamproject/govarnam-ibus/releases)
* Extract the zip file
* Open the extracted folder
* Now we need to run the install script
  * Double click `install.sh`, and choose "Run In Terminal"
  * Or for KDE, Right click, choose "Run In Konsole"
  * Or right click on folder, choose "Open In Terminal", type this and press enter:
  ```bash
  ./install.sh
  ```
* Type your password and press enter
* Restart ibus (no sudo)
```bash
ibus restart
```
* Go to IBus settings, add Varnam input method.
* Maybe set an easy to use switch key to switch between languages (Default is `Meta Key (Windows Key) + Space`)

To uninstall:
```
sudo ./install.sh uninstall
```

## Development

### Building

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
