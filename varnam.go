package main

/**
 * govarnam-ibus - An Indian Language Input Method Engine for GNU/Linux
 * Copyright Subin Siby, 2021
 * Licensed under AGPL-3.0-only
 *
 * gittu-engine - An IBus Engine in Go
 * goibus - golang implementation of libibus
 * Copyright Sarim Khan, 2016
 * Licensed under Mozilla Public License 1.1 ("MPL")
 */

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"gitlab.com/subins2000/govarnam-ibus/ibus"

	"github.com/godbus/dbus/v5"
)

var debug = flag.Bool("debug", false, "Enable debugging")
var embeded = flag.Bool("ibus", false, "Run the embeded ibus component")
var standalone = flag.Bool("standalone", false, "Run standalone by creating new component")
var generatexml = flag.String("xml", "", "Write xml representation of component to file or stdout if file == \"-\"")

func makeComponent() *ibus.Component {

	component := ibus.NewComponent(
		"org.freedesktop.IBus.Varnam",
		"GoVarnam Input Engine", // TODO change to Varnam
		"0.2",
		"AGPL-3.0",
		"Subin Siby",
		"https://subinsb.com/varnam",
		"/usr/local/bin/govarnam-ibus -ibus",
		"ibus-varnam")

	avroenginedesc := ibus.SmallEngineDesc(
		"govarnam",
		"GoVarnam", // TODO change to Varnam
		"GoVarnam Input Method",
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
	if *debug {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

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
