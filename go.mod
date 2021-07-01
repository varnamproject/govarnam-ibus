module gitlab.com/subins2000/govarnam-ibus

go 1.15

require (
	github.com/godbus/dbus/v5 v5.0.3
	github.com/gotk3/gotk3 v0.6.0
	gitlab.com/subins2000/govarnam v0.0.0-00010101000000-000000000000
)

replace gitlab.com/subins2000/govarnam-ibus/ibus => ./ibus

// Use this only for development
replace gitlab.com/subins2000/govarnam => ../govarnam
