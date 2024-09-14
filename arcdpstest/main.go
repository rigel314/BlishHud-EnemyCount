package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"os"
)

var (
	pid    = flag.Int("pid", 0, "overlay target pid")
	replay = flag.String("replay", "", "path to json file containing arcdps-bhud json combat events")
)

func main() {
	flag.Parse()

	f, err := os.Open(*replay)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	jdec := json.NewDecoder(f)

	port := int(uint16(*pid) | 1<<14 | 1<<15)
	log.Println("listening on port:", port)

	svr, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: port,
	})
	if err != nil {
		panic(err)
	}

	client, err := svr.AcceptTCP()
	if err != nil {
		panic(err)
	}

	log.Println("got client")

	for {
		var args CombatArgs
		err := jdec.Decode(&args)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			panic(err)
		}

		if args.EventType == bhud_AREA {
			args.EventType = wire_AREA
		}
		if args.EventType == bhud_LOCAL {
			args.EventType = wire_LOCAL
		}

		var fields uint8 = 0x0f

		if args.CombatEvent.Ev == nil {
			fields &^= hasEv
		}
		if args.CombatEvent.Src == nil {
			fields &^= hasSrc
		}
		if args.CombatEvent.Dst == nil {
			fields &^= hasDst
		}
		if args.CombatEvent.SkillName == nil {
			fields &^= hasSkillName
		}

		args.Fields = fields

		buf := make([]byte, 1024)
		n, err := Encode(buf, binary.LittleEndian, args)
		if err != nil {
			panic(err)
		}

		lenbuf := make([]byte, 8)
		binary.LittleEndian.PutUint64(lenbuf, uint64(n))

		_, err = client.Write(lenbuf)
		if err != nil {
			panic(err)
		}
		_, err = client.Write(buf[:n])
		if err != nil {
			panic(err)
		}

		// if args.CombatEvent.Ev != nil && args.CombatEvent.Ev.IsStateChange == state_LogEnd {
		// 	time.Sleep(10 * time.Second)
		// }
	}
}

const (
	bhud_AREA  uint8 = 0
	bhud_LOCAL uint8 = 1
)

const (
	wire_AREA  uint8 = 2
	wire_LOCAL uint8 = 3
)

const (
	hasEv        uint8 = 1 << iota
	hasSrc       uint8 = 1 << iota
	hasDst       uint8 = 1 << iota
	hasSkillName uint8 = 1 << iota
)

type CombatArgs struct {
	EventType   uint8
	Fields      uint8
	CombatEvent CombatEvent
}

type CombatEvent struct {
	Ev        *Ev
	Src       *Ag
	Dst       *Ag
	SkillName *string
	Id        uint64
	Revision  uint64
}

type Ev struct {
	Time            uint64
	SrcAgent        uint64
	DstAgent        uint64
	Value           int32
	BuffDmg         int32
	OverStackValue  uint32
	SkillId         uint32
	SrcInstId       uint16
	DstInstId       uint16
	SrcMasterInstId uint16
	DstMasterInstId uint16
	Iff             uint8
	Buff            bool
	Result          uint8
	IsActivation    uint8
	IsBuffRemove    uint8
	IsNinety        bool
	IsFifty         bool
	IsMoving        bool
	IsStateChange   uint8
	IsFlanking      bool
	IsShields       bool
	IsOffCycle      bool
	Pad61           uint8
	Pad62           uint8
	Pad63           uint8
	Pad64           uint8
}

type Ag struct {
	Name       string
	Id         uint64
	Profession uint32
	Elite      uint32
	Self       uint32
	Team       uint16
}

const (
	state_None uint8 = iota
	state_EnterCombat
	state_ExitCombat
	state_ChangeUp
	state_ChangeDead
	state_ChangeDown
	state_Spawn
	state_Despawn
	state_HealthUpdate
	state_LogStart
	state_LogEnd
	state_WeaponSwap
	state_MaxHealthUpdate
	state_PointOfView
	state_Language
	state_GWBuild
	state_ShardId
	state_Reward
	state_BuffInitial
	state_Position
	state_Velocity
	state_Rotation
	state_TeamChange
	state_AttackTarget
	state_Targetable
	state_MapID
	state_ReplInfo
	state_StackActive
	state_StackReset
	state_Guild
	state_BuffInfo
	state_BuffFormula
	state_SkillInfo
	state_SkillTiming
	state_BreakbarState
	state_BreakbarPercent
	state_Error
	state_Tag
	state_Unknown
)
