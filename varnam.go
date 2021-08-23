package main

/**
 * gittu-engine - An IBus Engine in Go
 * goibus - golang implementation of libibus
 * Copyright Sarim Khan, 2016
 * Copyright Nguyen Tran Hau, 2021
 * https://github.com/sarim/goibus
 * Licensed under Mozilla Public License 1.1 ("MPL")
 *
 * Derivative Changes: Modified names, added preferences
 * Copyright Subin Siby, 2021
 */

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/varnamproject/govarnam-ibus/ibus"

	"github.com/godbus/dbus/v5"
)

var installPrefix = "/usr/local"

var engineName = "Varnam"
var engineCode = "varnam"

// Bus name related to the engine used which is govarnam
var busName = "org.freedesktop.IBus.GoVarnam"

var schemeIDFlag = flag.String("s", "", "Scheme ID")
var engineLang = flag.String("lang", "", "Language")
var schemeID = ""

var inscriptMode = false

var debug = flag.Bool("debug", false, "Enable debugging")
var embeded = flag.Bool("ibus", false, "Run the embeded ibus component")
var standalone = flag.Bool("standalone", false, "Run standalone by creating new component")
var generatexml = flag.String("xml", "", "Write xml representation of component to file or stdout if file == \"-\"")
var prefix = flag.String("prefix", "", "Prefix location")
var prefs = flag.Bool("prefs", false, "Show preferences window")

func makeComponent() *ibus.Component {
	component := ibus.NewComponent(
		busName,
		engineName+" Input Engine",
		"1.0.0",
		"AGPL-3.0",
		"Subin Siby",
		"https://subinsb.com/varnam",
		installPrefix+"/bin/varnam-ibus-engine -ibus -s "+schemeID+" -lang "+*engineLang,
		"ibus-varnam-"+schemeID)

	avroenginedesc := ibus.SmallEngineDesc(
		engineCode,
		engineName,
		engineName+" Input Method",
		*engineLang,
		"AGPL-3.0",
		"Subin Siby",
		installPrefix+"/share/varnam/ibus/icons/"+engineCode+".png",
		"en",
		installPrefix+"/bin/varnam-ibus-engine -prefs -s "+schemeID+" -lang "+*engineLang,
		"1.0.0")

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

	if *schemeIDFlag == "" {
		log.Fatal("Need a scheme ID. Pass it to -s option")
	}
	if *engineLang == "" {
		log.Fatal("Need a language identifier. Pass it to -lang option")
	}

	schemeID = *schemeIDFlag
	engineName += "-" + schemeID
	engineCode += "-" + schemeID
	busName += "." + schemeID

	if strings.Contains(schemeID, "inscript") {
		inscriptMode = true
	}

	if *generatexml != "" {
		if *prefix == "" {
			log.Fatal("Install prefix location needed")
		}
		installPrefix = *prefix

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
		bus.RequestName(busName, 0)
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

	} else if *prefs {
		showPrefs()
	} else {
		Usage()
		os.Exit(1)
	}
}
