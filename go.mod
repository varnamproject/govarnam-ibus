module github.com/varnamproject/govarnam-ibus

go 1.15

require (
	github.com/godbus/dbus/v5 v5.0.3
	github.com/gotk3/gotk3 v0.6.2-0.20211107090813-1d544513fb74
	github.com/varnamproject/govarnam v0.0.0-00010101000000-000000000000
)

replace github.com/varnamproject/govarnam-ibus/ibus => ./ibus

// Use this only for development
replace github.com/varnamproject/govarnam => ../govarnam
