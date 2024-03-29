package main

import (
	"context"
	"fmt"
	"log"

	"github.com/varnamproject/govarnam-ibus/ibus"

	"github.com/godbus/dbus/v5"
	"github.com/varnamproject/govarnam/govarnamgo"
)

var lastTypedCharacter = ""

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

func getVarnamResult(ctx context.Context, channel chan<- []govarnamgo.Suggestion, word string) {
	if inscriptMode {
		result, err := varnam.GetSuggestions(ctx, word)
		if err == nil {
			channel <- result
		} else {
			log.Print(err)
		}
	} else {
		result, err := varnam.Transliterate(ctx, word)
		if err == nil {
			channel <- result
		} else {
			log.Print(err)
		}
	}
	close(channel)
}

func (e *VarnamEngine) InternalUpdateTable(ctx context.Context) {
	resultChannel := make(chan []govarnamgo.Suggestion)

	go getVarnamResult(ctx, resultChannel, string(e.preedit))

	select {
	case <-ctx.Done():
		return
	case result := <-resultChannel:
		e.table.Clear()

		if inscriptMode {
			// Append original string at beginning
			e.table.AppendCandidate(string(e.preedit))
			e.table.AppendLabel("0:")
		}

		for _, sug := range result {
			e.table.AppendCandidate(sug.Word)
		}

		label := uint32(1)
		for label <= e.table.PageSize {
			e.table.AppendLabel(fmt.Sprint(label) + ":")
			label++
		}

		if !inscriptMode {
			// Append original string at end
			e.table.AppendCandidate(string(e.preedit))
			e.table.AppendLabel("0:")
		}

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
	// 46 is .
	// 44 is ,
	// 63 is ?
	// 33 is !
	// 40 is (
	// 41 is )
	if keyval == 46 || keyval == 44 || keyval == 63 || keyval == 33 || keyval == 40 || keyval == 41 {
		return true
	}
	// 59 is ;
	// 39 is '
	// 34 is "
	if !inscriptMode && (keyval == 59 || keyval == 39 || keyval == 34) {
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

	ctrlModifiers := modifiers & ibus.IBUS_CONTROL_MASK
	if ctrlModifiers != 0 {
		if len(e.preedit) == 0 {
			return false, nil
		}
		if keyval == ibus.IBUS_Delete {
			if *debug {
				fmt.Println("CTRL + DEL = Unlearn word")
			}
			text := e.GetCandidate()
			if text != nil {
				varnam.Unlearn(text.Text)
				e.VarnamUpdateLookupTable()
			}
		}
		return true, nil
	}

	altModifiers := modifiers & ibus.IBUS_MOD1_MASK
	if altModifiers != 0 {
		if len(e.preedit) == 0 {
			return false, nil
		}
		if keyval == ibus.IBUS_Down {
			if *debug {
				fmt.Println("ALT + DOWN = Suggestions page down")
			}
			e.table.NextPage()
		} else if keyval == ibus.IBUS_Up {
			if *debug {
				fmt.Println("ALT + UP = Suggestions page up")
			}
			e.table.PreviousPage()
		}
		e.UpdateLookupTable(e.table, true)
		return true, nil
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
		return true, nil

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
		if len(e.preedit) == 0 {
			return false, nil
		}
		// Commit the text itself
		e.VarnamCommitText(ibus.NewText(string(e.preedit)), false)
		return true, nil
	}

	numericKey := uint32(10)

	switch keyval {
	case ibus.IBUS_1, ibus.IBUS_KP_1:
		numericKey = 0
		break
	case ibus.IBUS_2, ibus.IBUS_KP_2:
		numericKey = 1
		break
	case ibus.IBUS_3, ibus.IBUS_KP_3:
		numericKey = 2
		break
	case ibus.IBUS_4, ibus.IBUS_KP_4:
		numericKey = 3
		break
	case ibus.IBUS_5, ibus.IBUS_KP_5:
		numericKey = 4
		break
	case ibus.IBUS_6, ibus.IBUS_KP_6:
		numericKey = 5
		break
	case ibus.IBUS_7, ibus.IBUS_KP_7:
		numericKey = 6
		break
	case ibus.IBUS_8, ibus.IBUS_KP_8:
		numericKey = 7
		break
	case ibus.IBUS_9, ibus.IBUS_KP_9:
		numericKey = 8
	}

	if numericKey != 10 {
		if inscriptMode {
			// Inscript scheme uses ^1 to input ZWJ.
			// In usual enhanced inscript AltGr + 1 is used to achieve the same.
			// ^2 - ZWNJ
			// ^4 - ₹
			// Inscript scheme uses |number to input a native language numeral
			// |1 - Language Numeral 1 - ൧
			// |2 - Language Numeral 2 - ൨

			if lastTypedCharacter != "^" && lastTypedCharacter != "|" {
				return e.VarnamCommitCandidateAt(numericKey)
			}
		} else {
			return e.VarnamCommitCandidateAt(numericKey)
		}
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

		if inscriptMode {
			result := varnam.TransliterateGreedyTokenized(string(rune(keyval)))

			// Appending at cursor position
			if len(result) > 0 {
				e.preedit = insertAtIndex(e.preedit, e.cursorPos, []rune(result[0].Word)[0])
			} else {
				e.preedit = insertAtIndex(e.preedit, e.cursorPos, rune(keyval))
			}
			e.cursorPos++

			lastTypedCharacter = string(keyval)
		} else {
			// Appending at cursor position
			e.preedit = insertAtIndex(e.preedit, e.cursorPos, rune(keyval))
			e.cursorPos++
		}

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

func (c *VarnamEngine) Destroy() *dbus.Error {
	varnam.Close()
	varnam = nil
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

	if varnam == nil {
		var err error
		varnam, err = govarnamgo.InitFromID(schemeID)
		if err != nil {
			log.Fatal(err)
		}

		varnam.Debug(*debug)

		loadConfig()
		varnam.SetConfig(config)
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
