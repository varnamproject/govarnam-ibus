package main

import (
	"flag"
	"fmt"
	"os"

	"./ibus"

	"github.com/godbus/dbus"
)

var embeded = flag.Bool("ibus", false, "Run the embeded ibus component")
var standalone = flag.Bool("standalone", false, "Run standalone by creating new component")
var generatexml = flag.String("xml", "", "Write xml representation of component to file or stdout if file == \"-\"")

func makeComponent() *ibus.Component {

	component := ibus.NewComponent(
		"org.freedesktop.IBus.Varnam",
		"Varnam Input Engine",
		"0.2",
		"AGPL-3.0",
		"Subin Siby",
		"https://subinsb.com/varnam",
		"/usr/local/bin/govarnam-ibus -ibus",
		"ibus-varnam")

	avroenginedesc := ibus.SmallEngineDesc(
		"varnam",
		"Varnam",
		"Varnam Input Method",
		"ml",
		"AGPL-3.0",
		"Subin Siby",
		"/usr/local/share/varnam/ibus/icons/varnam.png",
		"en",
		"/usr/local/bin/govarnam-ibus -pref",
		"0.2")

	component.AddEngine(avroenginedesc)

	return component
}

func main() {

	var Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.CommandLine.VisitAll(func(f *flag.Flag) {
			format := "  -%s: %s\n"
			fmt.Fprintf(os.Stderr, format, f.Name, f.Usage)
		})
	}

	flag.Parse()

	if *generatexml != "" {
		c := makeComponent()

		if *generatexml == "-" {
			c.OutputXML(os.Stdout)
		} else {
			f, err := os.Create(*generatexml)
			if err != nil {
				panic(err)
			}

			c.OutputXML(f)
			f.Close()
		}
	} else if *embeded {
		bus := ibus.NewBus()
		fmt.Println("Got Bus, Running Embeded")

		conn := bus.GetDbusConn()
		ibus.NewFactory(conn, VarnamEngineCreator)
		bus.RequestName("org.freedesktop.IBus.Varnam", 0)
		select {}
	} else if *standalone {
		bus := ibus.NewBus()
		fmt.Println("Got Bus, Running Standalone")

		conn := bus.GetDbusConn()
		ibus.NewFactory(conn, VarnamEngineCreator)
		bus.RegisterComponent(makeComponent())

		fmt.Println("Setting Global Engine to me")
		bus.CallMethod("SetGlobalEngine", 0, "varnam")

		c := make(chan *dbus.Signal, 10)
		conn.Signal(c)

		select {
		case <-c:
		}

	} else {
		Usage()
		os.Exit(1)
	}
}