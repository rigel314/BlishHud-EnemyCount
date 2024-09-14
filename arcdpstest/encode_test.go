package main

import (
	"encoding/binary"
	"encoding/hex"
	"testing"
)

func TestEncode(t *testing.T) {
	args := CombatArgs{
		EventType: wire_AREA,
		Fields:    0x0f,
		CombatEvent: CombatEvent{
			Ev: &Ev{
				Time:            1,
				SrcAgent:        2,
				DstAgent:        3,
				Value:           4,
				BuffDmg:         5,
				OverStackValue:  6,
				SkillId:         7,
				SrcInstId:       8,
				DstInstId:       9,
				SrcMasterInstId: 10,
				DstMasterInstId: 11,
				Iff:             12,
				Buff:            false,
				Result:          14,
				IsActivation:    15,
				IsBuffRemove:    16,
				IsNinety:        false,
				IsFifty:         false,
				IsMoving:        false,
				IsStateChange:   20,
				IsFlanking:      false,
				IsShields:       false,
				IsOffCycle:      false,
				Pad61:           24,
				Pad62:           25,
				Pad63:           26,
				Pad64:           27,
			},
			Src: &Ag{
				Name:       "1",
				Id:         2,
				Profession: 3,
				Elite:      4,
				Self:       5,
				Team:       6,
			},
			Dst: &Ag{
				Name:       "7",
				Id:         8,
				Profession: 9,
				Elite:      10,
				Self:       11,
				Team:       12,
			},
			SkillName: String("blar"),
			Id:        1,
			Revision:  2,
		},
	}

	b := make([]byte, 1024)
	n, err := Encode(b, binary.LittleEndian, args)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if n > 0 {
		t.Log(hex.EncodeToString(b[:n]))
	}
}

func String(x string) *string {
	return &x
}
