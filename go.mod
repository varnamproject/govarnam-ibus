module gitlab.com/subins2000/govarnam-ibus

go 1.15

require (
	github.com/godbus/dbus/v5 v5.0.3
	gitlab.com/subins2000/govarnam v0.0.0-20210601183813-de1d79352eb0
)

replace (
    gitlab.com/subins2000/govarnam-ibus/ibus => "./ibus"
)
