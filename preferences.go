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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gotk3/gotk3/gtk"
	"github.com/varnamproject/govarnam/govarnamgo"
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

func makeNewHorizontalGrid() *gtk.Grid {
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

func loadConfig() {
	config = getVarnamDefaultConfig()

	path := getConfPath()
	if fileExists(path) {
		confFile, err := ioutil.ReadFile(path)

		if err != nil {
			log.Print("Reading config JSON failed, err: %s", err.Error())
		}

		confJSON := string(confFile)

		err = json.NewDecoder(strings.NewReader(confJSON)).Decode(&config)
		if err != nil {
			log.Print("Parsing config JSON failed, err: %s", err.Error())
		}
	}
}

func getVarnamDefaultConfig() govarnamgo.Config {
	config := govarnamgo.Config{
		IndicDigits:                       false,
		DictionarySuggestionsLimit:        4,
		PatternDictionarySuggestionsLimit: 3,
		TokenizerSuggestionsLimit:         10,
		TokenizerSuggestionsAlways:        true,
		DictionaryMatchExact:              false,
	}

	if inscriptMode {
		config.IndicDigits = true
	}
	return config
}

var config govarnamgo.Config

func makeSettingsPage() *gtk.Box {
	/* Page 1 - Settings */

	settingsPage, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	checkError(err)

	// gtk_widget_set_margin_left & right is deprecated since 3.12
	// It got removed in 3.98: https://github.com/GNOME/gtk/blob/main/NEWS.pre-4.0#L939
	// Our minimum GTK support needed is 3.10 (Ubuntu 14.04)
	settingsPage.SetMarginLeft(12)
	settingsPage.SetMarginRight(12)
	settingsPage.SetMarginBottom(12)

	/* Dictionary Match Exact Preference */
	dictMatchExactGrid := makeNewHorizontalGrid()

	dictMatchExactLabel, err := gtk.LabelNew("Strictly Follow Scheme For Dictionary Results")
	checkError(err)

	dictMatchExactCheck, err := gtk.CheckButtonNew()
	checkError(err)

	dictMatchExactCheck.SetActive(config.DictionaryMatchExact)

	dictMatchExactGrid.Add(dictMatchExactLabel)
	dictMatchExactGrid.Add(dictMatchExactCheck)

	/* Dictionary Match Exact Information */
	dictMatchExactInfoGrid := makeNewHorizontalGrid()

	dictMatchExactInfoLabel, err := gtk.LabelNew("")
	checkError(err)

	dictMatchExactInfoLabel.SetMarkup("If this ^ is turned on then suggestions will be more accurate according to <a href='" + upstreamURL + "/editor/#/scheme'>scheme</a>. But you will need to learn the <a href='" + upstreamURL + "/editor/#/scheme'>language scheme</a> thoroughly for the best experience.")
	dictMatchExactInfoLabel.SetLineWrap(true)

	dictMatchExactInfoGrid.Add(dictMatchExactInfoLabel)

	/* Dictionary Suggestion Preference */
	dictSugsSizeGrid := makeNewHorizontalGrid()

	dictSugsSizeLabel, err := gtk.LabelNew("Dictionary Suggestions Limit")
	checkError(err)

	dictSugsSizeInput, err := gtk.EntryNew()
	checkError(err)

	dictSugsSizeInput.SetInputPurpose(gtk.INPUT_PURPOSE_DIGITS)
	dictSugsSizeInput.Connect("changed", stripNonNumericChars)
	dictSugsSizeInput.SetText(fmt.Sprint(config.DictionarySuggestionsLimit))

	dictSugsSizeGrid.Add(dictSugsSizeLabel)
	dictSugsSizeGrid.Add(dictSugsSizeInput)

	/* Pattern Dictionary Suggestion Preference */
	patternDictSugsSizeGrid := makeNewHorizontalGrid()

	patternDictSugsSizeLabel, err := gtk.LabelNew("Pattern Dictionary Suggestions Limit")
	checkError(err)

	patternDictSugsSizeInput, err := gtk.EntryNew()
	checkError(err)

	patternDictSugsSizeInput.SetInputPurpose(gtk.INPUT_PURPOSE_DIGITS)
	patternDictSugsSizeInput.Connect("changed", stripNonNumericChars)
	patternDictSugsSizeInput.SetText(fmt.Sprint(config.PatternDictionarySuggestionsLimit))

	patternDictSugsSizeGrid.Add(patternDictSugsSizeLabel)
	patternDictSugsSizeGrid.Add(patternDictSugsSizeInput)

	/* Tokenizer Suggestion Preference */
	tokenizerSugsSizeGrid := makeNewHorizontalGrid()

	tokenizerSugsSizeLabel, err := gtk.LabelNew("Tokenizer Suggestions Limit")
	checkError(err)

	tokenizerSugsSizeInput, err := gtk.EntryNew()
	checkError(err)

	tokenizerSugsSizeInput.SetInputPurpose(gtk.INPUT_PURPOSE_DIGITS)
	tokenizerSugsSizeInput.Connect("changed", stripNonNumericChars)
	tokenizerSugsSizeInput.SetText(fmt.Sprint(config.TokenizerSuggestionsLimit))

	tokenizerSugsSizeGrid.Add(tokenizerSugsSizeLabel)
	tokenizerSugsSizeGrid.Add(tokenizerSugsSizeInput)

	actionButtons, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	checkError(err)

	saveButton, err := gtk.ButtonNewWithLabel("Save")
	checkError(err)

	var savedMessage *gtk.Label
	saveButton.Connect("clicked", func(btn *gtk.Button) {
		text, err := dictSugsSizeInput.GetText()
		checkError(err)

		i, _ := strconv.Atoi(text)
		config.DictionarySuggestionsLimit = i

		config.DictionaryMatchExact = dictMatchExactCheck.GetActive()

		text, err = tokenizerSugsSizeInput.GetText()
		checkError(err)

		i, _ = strconv.Atoi(text)
		config.TokenizerSuggestionsLimit = i

		saveConf(config)

		if savedMessage == nil {
			savedMessage, err = gtk.LabelNew("Saved. Restart ibus for changes to take effect")
			checkError(err)

			settingsPage.Add(savedMessage)
			savedMessage.Show()
		}
	})

	actionButtons.PackEnd(saveButton, true, true, 10)

	settingsPage.Add(dictMatchExactGrid)
	settingsPage.Add(dictMatchExactInfoGrid)

	settingsPage.Add(dictSugsSizeGrid)
	settingsPage.Add(patternDictSugsSizeGrid)
	settingsPage.Add(tokenizerSugsSizeGrid)
	settingsPage.Add(actionButtons)

	return settingsPage
}

var (
	rlwOffset int = 0
	rlwLimit  int = 30
)

// Shorthand function for getting recently learned words
func getRLW() []govarnamgo.Suggestion {
	words, err := varnam.GetRecentlyLearntWords(context.Background(), rlwOffset, rlwLimit)
	if err != nil {
		return []govarnamgo.Suggestion{}
	}
	return words
}

func refreshRLWList(list *gtk.ListBox, words []govarnamgo.Suggestion) {
	// Clear rows
	for {
		row := list.GetRowAtIndex(0)
		if row == nil {
			break
		}
		row.Destroy()
	}

	if *debug {
		log.Println(words)
	}

	for _, wordInfo := range words {
		word := wordInfo.Word

		row, err := gtk.ListBoxRowNew()
		checkError(err)

		box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
		checkError(err)

		timeLabel, err := gtk.LabelNew(
			time.Unix(int64(wordInfo.LearnedOn), 0).String(),
		)
		checkError(err)
		timeLabel.SetSelectable(true)

		wordLabel, err := gtk.LabelNew(word)
		checkError(err)
		wordLabel.SetSelectable(true)

		box.PackStart(timeLabel, false, false, 0)
		box.PackStart(wordLabel, true, true, 0)

		unlearnButton, err := gtk.ButtonNewWithLabel("Unlearn")
		unlearnButton.Connect("clicked", func() {
			err := varnam.Unlearn(word)
			log.Println(err)

			refreshRLWList(list, getRLW())
		})

		box.PackEnd(unlearnButton, false, true, 0)

		row.Add(box)
		list.Add(row)
	}

	list.ShowAll()
}

func makeRLWPage() *gtk.Box {
	/* Page 2 - Recently Learned Words */

	rlwPage, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	checkError(err)

	rlwPage.SetMarginLeft(12)
	rlwPage.SetMarginRight(12)
	rlwPage.SetMarginTop(12)
	rlwPage.SetMarginBottom(12)

	pageBtnBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	checkError(err)

	prevBtn, err := gtk.ButtonNewWithLabel("Previous")
	checkError(err)
	prevBtn.SetSensitive(false)

	nextBtn, err := gtk.ButtonNewWithLabel("Next")
	checkError(err)

	pageBtnBox.PackStart(prevBtn, false, false, 0)
	pageBtnBox.PackEnd(nextBtn, false, true, 0)

	list, err := gtk.ListBoxNew()
	checkError(err)
	list.SetSelectionMode(gtk.SELECTION_NONE)

	refreshRLWList(list, getRLW())

	prevBtn.Connect("clicked", func(btn *gtk.Button) {
		nextBtn.SetSensitive(true)

		rlwOffset -= rlwLimit
		if rlwOffset <= 0 {
			rlwOffset = 0
			prevBtn.SetSensitive(false)
		}

		refreshRLWList(list, getRLW())
	})

	nextBtn.Connect("clicked", func(btn *gtk.Button) {
		prevBtn.SetSensitive(true)

		rlwOffset += rlwLimit
		words := getRLW()

		if len(words) == 0 {
			rlwOffset -= rlwLimit
			nextBtn.SetSensitive(false)
		} else {
			refreshRLWList(list, words)
		}
	})

	rlwPage.Add(pageBtnBox)
	rlwPage.Add(list)

	return rlwPage
}

func showPrefs() {
	gtk.Init(nil)

	loadConfig()

	// Create a new toplevel window, set its title, and connect it to the
	// "destroy" signal to exit the GTK main loop when it is destroyed.
	mainWin, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}

	win, err := gtk.ScrolledWindowNew(nil, nil)
	checkError(err)

	mainWin.Add(win)

	varnam, err = govarnamgo.InitFromID(schemeID)
	if err != nil {
		dialog := gtk.MessageDialogNew(mainWin, 0, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, "Varnam Error: "+err.Error())
		dialog.Run()
		return
	}

	mainWin.SetDefaultSize(640, 480)
	mainWin.SetResizable(false)
	mainWin.SetPosition(gtk.WIN_POS_CENTER)
	mainWin.SetTitle("Varnam " + varnam.GetSchemeDetails().DisplayName + " Preferences (" + engineName + ")")
	mainWin.Connect("destroy", func() {
		gtk.MainQuit()
	})

	notebook, err := gtk.NotebookNew()
	checkError(err)

	notebook.SetScrollable(true)

	notebook.SetMarginLeft(12)
	notebook.SetMarginRight(12)
	notebook.SetMarginBottom(12)

	settingsLabel, err := gtk.LabelNew("Settings")
	checkError(err)
	notebook.AppendPage(makeSettingsPage(), settingsLabel)

	rlwLabel, err := gtk.LabelNew("Recently Learnt Words")
	checkError(err)
	notebook.AppendPage(makeRLWPage(), rlwLabel)

	win.Add(notebook)

	// Recursively show all widgets contained in this window.
	mainWin.ShowAll()

	// Begin executing the GTK main loop.  This blocks until
	// gtk.MainQuit() is run.
	gtk.Main()
}
