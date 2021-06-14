# IBus Engine For govarnam

goibus - golang implementation of libibus

Thanks to [sarim](https://github.com/sarim/goibus) and [haunt98](https://github.com/haunt98/goibus) for developing goibus from which `govarnam-ibus` is developed.

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