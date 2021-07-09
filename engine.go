package main

import (
	"context"
	"fmt"
	"log"

	"gitlab.com/subins2000/govarnam-ibus/ibus"

	"github.com/godbus/dbus/v5"
	"gitlab.com/subins2000/govarnam/govarnamgo"
)

var varnam *govarnamgo.VarnamHandle

type VarnamEngine struct {
	ibus.Engine
	propList          *ibus.PropList
	preedit           []rune
	cursorPos         uint32
	table             *ibus.LookupTable
	transliterateCTX  context.Context
	updateTableCancel context.CancelFunc
}

func (e *VarnamEngine) VarnamUpdatePreedit() {
	e.UpdatePreeditText(ibus.NewText(string(e.preedit)), e.cursorPos, true)
}

func (e *VarnamEngine) VarnamClearState() {
	e.preedit = []rune{}
	e.cursorPos = 0
	e.VarnamUpdatePreedit()

	e.table.Clear()
	e.HideLookupTable()
}

func (e *VarnamEngine) VarnamCommitText(text *ibus.Text, shouldLearn bool) bool {
	if shouldLearn {
		go varnam.Learn(text.Text, 0)
		// TODO error handle
	}
	e.CommitText(text)
	e.VarnamClearState()
	return true
}

func getVarnamResult(ctx context.Context, channel chan<- govarnamgo.TransliterationResult, word string) {
	channel <- varnam.Transliterate(ctx, word)
	close(channel)
}

func (e *VarnamEngine) InternalUpdateTable(ctx context.Context) {
	resultChannel := make(chan govarnamgo.TransliterationResult)

	go getVarnamResult(ctx, resultChannel, string(e.preedit))

	select {
	case <-ctx.Done():
		return
	case result := <-resultChannel:
		e.table.Clear()

		for _, sug := range result.ExactMatches {
			e.table.AppendCandidate(sug.Word)
		}

		for _, sug := range result.PatternDictionarySuggestions {
			e.table.AppendCandidate(sug.Word)
		}

		for _, sug := range result.DictionarySuggestions {
			e.table.AppendCandidate(sug.Word)
		}

		for _, sug := range result.GreedyTokenized {
			e.table.AppendCandidate(sug.Word)
		}

		for _, sug := range result.TokenizerSuggestions {
			e.table.AppendCandidate(sug.Word)
		}

		label := uint32(1)
		for label <= e.table.PageSize {
			e.table.AppendLabel(fmt.Sprint(label) + ":")
			label++
		}

		// Append original string at end
		e.table.AppendCandidate(string(e.preedit))
		e.table.AppendLabel("0:")

		e.UpdateLookupTable(e.table, true)
	}
}

func (e *VarnamEngine) VarnamUpdateLookupTable() {
	if e.updateTableCancel != nil {
		e.updateTableCancel()
		e.updateTableCancel = nil
	}

	if len(e.preedit) == 0 {
		e.HideLookupTable()
		return
	}

	ctx, cancel := context.WithCancel(e.transliterateCTX)
	e.updateTableCancel = cancel

	e.InternalUpdateTable(ctx)
}

func (e *VarnamEngine) GetCandidateAt(index uint32) *ibus.Text {
	if int(index) > len(e.table.Candidates)-1 {
		return nil
	}
	text := e.table.Candidates[index].Value().(ibus.Text)
	return &text
}

func (e *VarnamEngine) GetCandidate() *ibus.Text {
	return e.GetCandidateAt(e.table.CursorPos)
}

func (e *VarnamEngine) VarnamCommitCandidateAt(index uint32) (bool, *dbus.Error) {
	page := uint32(e.table.CursorPos / e.table.PageSize)

	index = page*e.table.PageSize + index

	if *debug {
		fmt.Println("Pagination picker:", len(e.table.Candidates), e.table.CursorPos, page, index)
	}

	text := e.GetCandidateAt(uint32(index))
	if text != nil {
		return e.VarnamCommitText(text, true), nil
	}
	return false, nil
}

func isWordBreak(ukeyval uint32) bool {
	keyval := int(ukeyval)
	if keyval == 46 || keyval == 44 || keyval == 63 || keyval == 33 || keyval == 40 || keyval == 41 || keyval == 34 || keyval == 59 || keyval == 39 {
		return true
	}
	return false
}

func (e *VarnamEngine) ProcessKeyEvent(keyval uint32, keycode uint32, modifiers uint32) (bool, *dbus.Error) {
	if *debug {
		fmt.Println("Process Key Event > ", keyval, keycode, modifiers)
	}

	// Ignore key release events
	is_press := modifiers&ibus.IBUS_RELEASE_MASK == 0
	if !is_press {
		return false, nil
	}

	modifiers = modifiers & (ibus.IBUS_CONTROL_MASK | ibus.IBUS_MOD1_MASK)

	if modifiers != 0 {
		if len(e.preedit) == 0 {
			return false, nil
		} else {
			return true, nil
		}
	}

	switch keyval {
	case ibus.IBUS_Space:
		text := e.GetCandidate()
		if text == nil {
			e.VarnamCommitText(ibus.NewText(string(e.preedit)+" "), false)
		} else {
			e.VarnamCommitText(ibus.NewText(text.Text+" "), true)
		}
		return true, nil

	case ibus.IBUS_Return:
		text := e.GetCandidate()
		if text == nil {
			e.VarnamCommitText(ibus.NewText(string(e.preedit)), false)
			return false, nil
		} else {
			e.VarnamCommitText(text, true)
		}
		return true, nil

	case ibus.IBUS_Escape:
		if len(e.preedit) == 0 {
			return false, nil
		}
		e.VarnamCommitText(ibus.NewText(string(e.preedit)), false)
		return false, nil

	case ibus.IBUS_Left:
		if len(e.preedit) == 0 {
			return false, nil
		}
		if e.cursorPos > 0 {
			e.cursorPos--
			e.VarnamUpdatePreedit()
		}
		return true, nil

	case ibus.IBUS_Right:
		if len(e.preedit) == 0 {
			return false, nil
		}
		if int(e.cursorPos) < len(e.preedit) {
			e.cursorPos++
			e.VarnamUpdatePreedit()
		}
		return true, nil

	case ibus.IBUS_Up:
		if len(e.preedit) == 0 {
			return false, nil
		}
		e.table.CursorUp()
		e.UpdateLookupTable(e.table, true)
		return true, nil

	case ibus.IBUS_Down:
		if len(e.preedit) == 0 {
			return false, nil
		}
		e.table.CursorDown()
		e.UpdateLookupTable(e.table, true)
		return true, nil

	case ibus.IBUS_BackSpace:
		if len(e.preedit) == 0 {
			return false, nil
		}
		if e.cursorPos > 0 {
			e.cursorPos--
			e.preedit = removeAtIndex(e.preedit, e.cursorPos)
			e.VarnamUpdatePreedit()
			e.VarnamUpdateLookupTable()
			if len(e.preedit) == 0 {
				/* Current backspace has cleared the preedit. Need to reset the engine state */
				e.VarnamClearState()
			}
		}
		return true, nil

	case ibus.IBUS_Delete:
		if len(e.preedit) == 0 {
			return false, nil
		}
		if int(e.cursorPos) < len(e.preedit) {
			e.preedit = removeAtIndex(e.preedit, e.cursorPos)
			e.VarnamUpdatePreedit()
			e.VarnamUpdateLookupTable()
			if len(e.preedit) == 0 {
				/* Current delete has cleared the preedit. Need to reset the engine state */
				e.VarnamClearState()
			}
		}
		return true, nil

	case ibus.IBUS_Home, ibus.IBUS_KP_Home:
		if len(e.preedit) == 0 {
			return false, nil
		}
		e.cursorPos = 0
		e.VarnamUpdatePreedit()
		return true, nil

	case ibus.IBUS_End, ibus.IBUS_KP_End:
		if len(e.preedit) == 0 {
			return false, nil
		}
		e.cursorPos = uint32(len(e.preedit))
		e.VarnamUpdatePreedit()
		return true, nil

	case ibus.IBUS_0, ibus.IBUS_KP_0:
		// Commit the text itself
		e.VarnamCommitText(ibus.NewText(string(e.preedit)), false)
		return true, nil
	case ibus.IBUS_1, ibus.IBUS_KP_1:
		return e.VarnamCommitCandidateAt(0)
	case ibus.IBUS_2, ibus.IBUS_KP_2:
		return e.VarnamCommitCandidateAt(1)
	case ibus.IBUS_3, ibus.IBUS_KP_3:
		return e.VarnamCommitCandidateAt(2)
	case ibus.IBUS_4, ibus.IBUS_KP_4:
		return e.VarnamCommitCandidateAt(3)
	case ibus.IBUS_5, ibus.IBUS_KP_5:
		return e.VarnamCommitCandidateAt(4)
	case ibus.IBUS_6, ibus.IBUS_KP_6:
		return e.VarnamCommitCandidateAt(5)
	case ibus.IBUS_7, ibus.IBUS_KP_7:
		return e.VarnamCommitCandidateAt(6)
	case ibus.IBUS_8, ibus.IBUS_KP_8:
		return e.VarnamCommitCandidateAt(7)
	case ibus.IBUS_9, ibus.IBUS_KP_9:
		return e.VarnamCommitCandidateAt(8)
	}

	if isWordBreak(keyval) {
		text := e.GetCandidate()
		if text != nil {
			e.VarnamCommitText(ibus.NewText(text.Text+string(keyval)), true)
			return true, nil
		}
		return false, nil
	}

	if keyval <= 128 {
		if len(e.preedit) == 0 {
			/* We are starting a new word. Now there could be a word selected in the text field
			 * and we may be typing over the selection. In this case to clear the selection
			 * we commit a empty text which will trigger the textfield to clear the selection.
			 * If there is no selection, this won't affect anything */
			e.CommitText(ibus.NewText(""))
		}

		// Appending at cursor position
		e.preedit = insertAtIndex(e.preedit, e.cursorPos, rune(keyval))
		e.cursorPos++

		e.VarnamUpdatePreedit()

		e.VarnamUpdateLookupTable()

		return true, nil
	}
	return false, nil
}

func (e *VarnamEngine) FocusIn() *dbus.Error {
	e.RegisterProperties(e.propList)
	return nil
}

func (e *VarnamEngine) FocusOut() *dbus.Error {
	e.VarnamClearState()
	return nil
}

func (e *VarnamEngine) PropertyActivate(prop_name string, prop_state uint32) *dbus.Error {
	fmt.Println("PropertyActivate", prop_name)
	return nil
}

var eid = 0

func VarnamEngineCreator(conn *dbus.Conn, engineName string) dbus.ObjectPath {
	eid++
	fmt.Println("Creating Varnam Engine #", eid)
	objectPath := dbus.ObjectPath(fmt.Sprintf("/org/freedesktop/IBus/Engine/VarnamGo/%d", eid))

	propp := ibus.NewProperty(
		"setup",
		ibus.PROP_TYPE_NORMAL,
		"Preferences - Varnam",
		"gtk-preferences",
		"Configure Varnam Engine",
		true,
		true,
		ibus.PROP_STATE_UNCHECKED)

	engine := &VarnamEngine{
		ibus.BaseEngine(conn, objectPath),
		ibus.NewPropList(propp),
		[]rune{},
		0,
		ibus.NewLookupTable(),
		context.Background(),
		nil}

	// TODO add SetOrientation method
	// engine.table.emitSignal("SetOrientation", ibus.IBUS_ORIENTATION_VERTICAL)

	var err error
	varnam, err = govarnamgo.InitFromID("ml")
	if err != nil {
		log.Fatal(err)
	}

	varnam.Debug(*debug)

	configLocal := retrieveSavedConf()
	if configLocal != nil {
		varnam.SetConfig(*configLocal)
	}

	ibus.PublishEngine(conn, objectPath, engine)
	return objectPath
}

func removeAtIndex(s []rune, index uint32) []rune {
	return append(s[0:index], s[index+1:]...)
}

// Thanks wasmup https://stackoverflow.com/a/61822301/1372424
// 0 <= index <= len(a)
func insertAtIndex(a []rune, index uint32, value rune) []rune {
	if uint32(len(a)) == index { // nil or empty slice or after last element
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...) // index < len(a)
	a[index] = value
	return a
}
