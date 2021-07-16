package main

/**
 * GoVarnam IBus Engine Preferences
 * Copyright Subin Siby, 2021
 * Licensed under AGPL-3.0
 *
 * For preferences to be applied, ibus
 * need to be restarted: ibus restart
 */

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/gotk3/gotk3/gtk"
	"gitlab.com/subins2000/govarnam/govarnamgo"
)

func getConfPath() string {
	var (
		loc string
		dir string
	)

	home := os.Getenv("XDG_DATA_HOME")
	if home == "" {
		home = os.Getenv("HOME")
		dir = path.Join(home, ".local", "share", "varnam")
	} else {
		dir = path.Join(home, "varnam")
	}
	loc = path.Join(dir, engineCode+"-ibus.conf")

	return loc
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func checkError(err error) {
	if err != nil {
		log.Fatal("Unable to create widget:", err)
	}
}

func makeNewGrid() *gtk.Grid {
	grid, err := gtk.GridNew()
	checkError(err)
	grid.SetMarginTop(12)
	grid.SetRowSpacing(12)
	grid.SetColumnSpacing(12)
	return grid
}

func stripNonNumericChars(input *gtk.Entry) {
	var result strings.Builder
	s, err := input.GetText()
	checkError(err)
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('0' <= b && b <= '9') ||
			b == ' ' {
			result.WriteByte(b)
		}
	}
	input.SetText(result.String())
}

func saveConf(config govarnamgo.Config) {
	jsonBytes, _ := json.Marshal(config)
	err := ioutil.WriteFile(getConfPath(), jsonBytes, 0644)
	checkError(err)
}

func retrieveSavedConf() *govarnamgo.Config {
	path := getConfPath()
	if fileExists(path) {
		var configLocal govarnamgo.Config
		confFile, _ := ioutil.ReadFile(path)

		if err := json.Unmarshal(confFile, &configLocal); err != nil {
			log.Fatal("Parsing conf JSON failed, err: %s", err.Error())
		}
		return &configLocal
	}
	return nil
}

func getVarnamDefaultConfig() govarnamgo.Config {
	config := govarnamgo.Config{IndicDigits: false, DictionarySuggestionsLimit: 5, TokenizerSuggestionsLimit: 10, TokenizerSuggestionsAlways: true}

	if inscriptMode {
		config.IndicDigits = true
	}
	return config
}

func showPrefs() {
	gtk.Init(nil)

	config := getVarnamDefaultConfig()

	configLocal := retrieveSavedConf()
	if configLocal != nil {
		config = *configLocal
	}

	// Create a new toplevel window, set its title, and connect it to the
	// "destroy" signal to exit the GTK main loop when it is destroyed.
	win, err := gtk.WindowNew(gtk.WINDOW_POPUP)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetPosition(gtk.WIN_POS_CENTER)
	win.SetTitle(engineName + " Preferences")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	dialog, err := gtk.DialogNew()
	dialog.SetTitle(engineName + " Preferences")
	dialog.SetTransientFor(win)
	dialog.Connect("destroy", func() {
		gtk.MainQuit()
	})

	gtkBox, err := dialog.GetContentArea()
	checkError(err)

	gtkBox.SetMarginStart(12)
	gtkBox.SetMarginEnd(12)
	gtkBox.SetMarginBottom(12)

	/* Dictionary Suggestion Preference */
	dictSugsSizeGrid := makeNewGrid()

	dictSugsSizeLabel, err := gtk.LabelNew("Dictionary Suggestions Limit")
	checkError(err)

	dictSugsSizeInput, err := gtk.EntryNew()
	checkError(err)

	dictSugsSizeInput.SetInputPurpose(gtk.INPUT_PURPOSE_DIGITS)
	dictSugsSizeInput.Connect("changed", stripNonNumericChars)
	dictSugsSizeInput.SetText(fmt.Sprint(config.DictionarySuggestionsLimit))

	dictSugsSizeGrid.Add(dictSugsSizeLabel)
	dictSugsSizeGrid.Add(dictSugsSizeInput)

	/* Dictionary Suggestion Preference */
	tokenizerSugsSizeGrid := makeNewGrid()

	tokenizerSugsSizeLabel, err := gtk.LabelNew("Tokenizer Suggestions Limit")
	checkError(err)

	tokenizerSugsSizeInput, err := gtk.EntryNew()
	checkError(err)

	tokenizerSugsSizeInput.SetInputPurpose(gtk.INPUT_PURPOSE_DIGITS)
	tokenizerSugsSizeInput.Connect("changed", stripNonNumericChars)
	tokenizerSugsSizeInput.SetText(fmt.Sprint(config.TokenizerSuggestionsLimit))

	tokenizerSugsSizeGrid.Add(tokenizerSugsSizeLabel)
	tokenizerSugsSizeGrid.Add(tokenizerSugsSizeInput)

	gtkBox.Add(dictSugsSizeGrid)
	gtkBox.Add(tokenizerSugsSizeGrid)

	dialog.AddButton("Save", gtk.RESPONSE_APPLY)
	dialog.AddButton("Cancel", gtk.RESPONSE_CANCEL)

	dialog.Connect("response", func(d *gtk.Dialog, response gtk.ResponseType) {
		if response == gtk.RESPONSE_APPLY {
			text, err := dictSugsSizeInput.GetText()
			checkError(err)

			i, _ := strconv.Atoi(text)
			config.DictionarySuggestionsLimit = i

			text, err = tokenizerSugsSizeInput.GetText()
			checkError(err)

			i, _ = strconv.Atoi(text)
			config.TokenizerSuggestionsLimit = i

			saveConf(config)
		}
		gtk.MainQuit()
	})

	dialog.SetDefaultSize(100, 150)
	// Recursively show all widgets contained in this window.
	dialog.ShowAll()

	// Begin executing the GTK main loop.  This blocks until
	// gtk.MainQuit() is run.
	gtk.Main()
}
