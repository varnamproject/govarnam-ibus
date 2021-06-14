package ibus

const (
	IBUS_CONTROL_MASK = 1 << 2
	IBUS_MOD1_MASK    = 1 << 3
	IBUS_RELEASE_MASK = 1 << 30

	IBUS_space     = 0x020
	IBUS_Return    = 0xff0d
	IBUS_Escape    = 0xff1b
	IBUS_Left      = 0xff51
	IBUS_Right     = 0xff53
	IBUS_Up        = 0xff52
	IBUS_Down      = 0xff54
	IBUS_BackSpace = 0xff08
	IBUS_Delete    = 0xffff

	IBUS_0 = 0x030
	IBUS_1 = 0x031
	IBUS_2 = 0x032
	IBUS_3 = 0x033
	IBUS_4 = 0x034
	IBUS_5 = 0x035
	IBUS_6 = 0x036
	IBUS_7 = 0x037
	IBUS_8 = 0x038
	IBUS_9 = 0x039

	IBUS_KP_0 = 0xffb0
	IBUS_KP_1 = 0xffb1
	IBUS_KP_2 = 0xffb2
	IBUS_KP_3 = 0xffb3
	IBUS_KP_4 = 0xffb4
	IBUS_KP_5 = 0xffb5
	IBUS_KP_6 = 0xffb6
	IBUS_KP_7 = 0xffb7
	IBUS_KP_8 = 0xffb8
	IBUS_KP_9 = 0xffb9
)
