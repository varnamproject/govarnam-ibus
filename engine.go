package main

import (
	"fmt"
	"log"

	"gitlab.com/subins2000/govarnam-ibus/ibus"

	"github.com/godbus/dbus/v5"
	"gitlab.com/subins2000/govarnam/govarnam"
)

var handle *govarnam.Varnam

type VarnamEngine struct {
	ibus.Engine
	propList  *ibus.PropList
	preedit   []rune
	cursorPos uint32
	table     *ibus.LookupTable
}

func (e *VarnamEngine) VarnamUpdatePreedit() {
	e.UpdatePreeditText(ibus.NewText(string(e.preedit)), e.cursorPos, true)
}

func (e *VarnamEngine) VarnamClearState() {
	e.preedit = []rune{}
	e.cursorPos = 0
	e.VarnamUpdatePreedit()

	// TODO Is this the correct way to clear ?
	e.table = ibus.NewLookupTable()
	e.HideLookupTable()
}

func (e *VarnamEngine) VarnamCommitText(text *ibus.Text, shouldLearn bool) bool {
	if shouldLearn {
		go handle.Learn(text.Text)
		// TODO error handle
	}
	e.CommitText(text)
	e.VarnamClearState()
	return true
}

func (e *VarnamEngine) VarnamUpdateLookupTable() {
	txt := string(e.preedit)
	if len(e.preedit) == 0 {
		e.HideLookupTable()
		return
	}

	fmt.Println(string(e.preedit))

	// TODO clear lookup table using emitSignal maybe ?
	// Is this the correct way ?
	table := ibus.NewLookupTable()

	result := handle.Transliterate(string(e.preedit))

	// Don't update lookup table if the result is late and next suggestion lookup has begun
	if txt != string(e.preedit) {
		return
	}

	label := 1

	for _, sug := range result.ExactMatch {
		table.AppendCandidate(sug.Word)
		table.AppendLabel(fmt.Sprint(label) + ":")
		label++
	}

	// POINTER1: If no exact matches show greedy first
	if len(result.ExactMatch) == 0 && len(result.GreedyTokenized) > 0 {
		table.AppendCandidate(result.GreedyTokenized[0].Word)
		table.AppendLabel(fmt.Sprint(label) + ":")
		label++
	}

	for _, sug := range result.Suggestions {
		table.AppendCandidate(sug.Word)
		table.AppendLabel(fmt.Sprint(label) + ":")
		label++
	}

	if len(result.Suggestions) == 0 {
		for _, sug := range result.GreedyTokenized {
			table.AppendCandidate(sug.Word)
			table.AppendLabel(fmt.Sprint(label) + ":")
			label++
		}
	}

	// POINTER1: If exact match exist, show greedy last
	if len(result.ExactMatch) > 0 && len(result.GreedyTokenized) > 0 {
		table.AppendCandidate(result.GreedyTokenized[0].Word)
		table.AppendLabel(fmt.Sprint(label) + ":")
		label++
	}

	// Append original string at end
	table.AppendCandidate(string(e.preedit))
	table.AppendLabel(fmt.Sprint(label) + ":")

	// Don't update lookup table if the result is late and next suggestion lookup has begun
	if txt != string(e.preedit) {
		return
	}

	e.table = table
	e.UpdateLookupTable(e.table, true)
}

func (e *VarnamEngine) GetCandidateAt(index uint32) *ibus.Text {
	if len(e.table.Candidates) == 0 {
		return nil
	}
	text := e.table.Candidates[index].Value().(ibus.Text)
	return &text
}

func (e *VarnamEngine) GetCandidate() *ibus.Text {
	return e.GetCandidateAt(e.table.CursorPos)
}

func isWordBreak(ukeyval uint32) bool {
	keyval := int(ukeyval)
	if keyval == 46 || keyval == 44 || keyval == 63 || keyval == 33 || keyval == 40 || keyval == 41 || keyval == 34 || keyval == 59 || keyval == 39 {
		return true
	}
	return false
}

func (e *VarnamEngine) ProcessKeyEvent(keyval uint32, keycode uint32, modifiers uint32) (bool, *dbus.Error) {
	fmt.Println("Process Key Event > ", keyval, keycode, modifiers)

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
	case ibus.IBUS_space:
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
				/* Current backspace has cleared the preedit. Need to reset the engine state */
				e.VarnamClearState()
			}
		}
		return true, nil
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
	fmt.Println("FocusIn")
	e.RegisterProperties(e.propList)
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
		ibus.NewLookupTable()}

	// TODO add SetOrientation method
	// engine.table.emitSignal("SetOrientation", ibus.IBUS_ORIENTATION_VERTICAL)

	var err error
	handle, err = govarnam.InitFromLang("ml")
	if err != nil {
		log.Fatal(err)
	}
	handle.Debug(true)

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
