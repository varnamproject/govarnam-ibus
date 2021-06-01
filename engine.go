package main

import (
	"fmt"
	"log"

	"./ibus"
	"github.com/godbus/dbus"
	"gitlab.com/subins2000/govarnam/govarnam"
)

const IBUS_CONTROL_MASK = 1 << 2
const IBUS_MOD1_MASK = 1 << 3
const IBUS_ORIENTATION_VERTICAL = 1
const IBUS_RELEASE_MASK = 1 << 30

const IBUS_space = 0x020
const IBUS_Return = 0xff0d
const IBUS_Escape = 0xff1b
const IBUS_Left = 0xff51
const IBUS_Right = 0xff53
const IBUS_Up = 0xff52
const IBUS_Down = 0xff54
const IBUS_BackSpace = 0xff08
const IBUS_Delete = 0xffff

const IBUS_1 = 0xff0d
const IBUS_2 = 0xff0d
const IBUS_3 = 0xff0d
const IBUS_4 = 0xff0d
const IBUS_5 = 0xff0d
const IBUS_6 = 0xff0d
const IBUS_7 = 0xff0d
const IBUS_8 = 0xff0d
const IBUS_9 = 0xff0d

const IBUS_KP_1 = 0xff0d
const IBUS_KP_2 = 0xff0d
const IBUS_KP_3 = 0xff0d
const IBUS_KP_4 = 0xff0d
const IBUS_KP_5 = 0xff0d
const IBUS_KP_6 = 0xff0d
const IBUS_KP_7 = 0xff0d
const IBUS_KP_8 = 0xff0d
const IBUS_KP_9 = 0xff0d

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
		handle.Learn(text.Text)
		// TODO error handle
	}
	e.CommitText(text)
	e.VarnamClearState()
	return true
}

func (e *VarnamEngine) VarnamUpdateLookupTable() {
	if len(e.preedit) == 0 {
		e.HideLookupTable()
		return
	}

	fmt.Println(string(e.preedit))

	// TODO clear lookup table using emitSignal maybe ?
	// Is this the correct way ?
	e.table = ibus.NewLookupTable()

	result := handle.Transliterate(string(e.preedit), 2)

	e.table.AppendCandidate(result.GreedyTokenized[0].Word)
	e.table.AppendLabel("1")

	label := 2
	for _, sug := range result.Suggestions {
		e.table.AppendCandidate(sug.Word)
		e.table.AppendLabel(fmt.Sprint(label) + ":")
		label++
	}

	if len(result.Suggestions) == 0 {
		for _, sug := range result.GreedyTokenized {
			e.table.AppendCandidate(sug.Word)
			e.table.AppendLabel(fmt.Sprint(label) + ":")
			label++
		}
	}

	// Append original string at end
	e.table.AppendCandidate(string(e.preedit))

	e.UpdateLookupTable(e.table, true)
	fmt.Println(string(e.preedit))
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

func (e *VarnamEngine) ProcessKeyEvent(keyval uint32, keycode uint32, modifiers uint32) (bool, *dbus.Error) {
	fmt.Println("Process Key Event > ", keyval, keycode, modifiers)

	// Ignore key release events
	is_press := modifiers&IBUS_RELEASE_MASK == 0
	if !is_press {
		return false, nil
	}

	modifiers = modifiers & (IBUS_CONTROL_MASK | IBUS_MOD1_MASK)

	if modifiers != 0 {
		if len(e.preedit) == 0 {
			return false, nil
		} else {
			return true, nil
		}
	}

	switch keyval {
	case IBUS_space:
		text := e.GetCandidate()
		if text == nil {
			e.VarnamCommitText(ibus.NewText(string(e.preedit)+" "), false)
			return false, nil
		} else {
			e.VarnamCommitText(ibus.NewText(text.Text+" "), true)
		}
		return true, nil

	case IBUS_Return:
		text := e.GetCandidate()
		if text == nil {
			e.VarnamCommitText(ibus.NewText(string(e.preedit)), false)
			return false, nil
		} else {
			e.VarnamCommitText(text, true)
		}
		return true, nil

	case IBUS_Left:
		if len(e.preedit) == 0 {
			return false, nil
		}
		if e.cursorPos > 0 {
			e.cursorPos--
			e.VarnamUpdatePreedit()
		}
		return true, nil

	case IBUS_Right:
		if len(e.preedit) == 0 {
			return false, nil
		}
		if int(e.cursorPos) < len(e.preedit) {
			e.cursorPos++
			e.VarnamUpdatePreedit()
		}
		return true, nil

	case IBUS_Up:
		if len(e.preedit) == 0 {
			return false, nil
		}
		e.table.CursorPos--
		e.UpdateLookupTable(e.table, true)
		return true, nil

	case IBUS_Down:
		if len(e.preedit) == 0 {
			return false, nil
		}
		e.table.CursorPos++
		e.UpdateLookupTable(e.table, true)
		return true, nil

	case IBUS_BackSpace:
		if len(e.preedit) == 0 {
			return false, nil
		}
		if e.cursorPos > 0 {
			e.cursorPos--
			e.preedit = removeAtIndex(e.preedit, e.cursorPos)
			e.VarnamUpdatePreedit()
			e.VarnamUpdateLookupTable()
			if len(e.preedit) == 0 {
				e.VarnamClearState()
			}
		}
		return true, nil
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
	// engine.table.emitSignal("SetOrientation", IBUS_ORIENTATION_VERTICAL)

	var err error
	handle, err = govarnam.InitFromLang("ml")
	if err != nil {
		log.Fatal(err)
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
