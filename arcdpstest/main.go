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

const (
	CBTS_NONE uint8 = iota // not used - not this kind of event
	// not used - not this kind of event

	CBTS_ENTERCOMBAT // agent entered combat
	// src_agent: relates to agent
	// dst_agent: subgroup
	// value: prof id
	// buff_dmg: elite spec id
	// evtc: limited to squad outside instances
	// realtime: limited to squad

	CBTS_EXITCOMBAT // agent left combat
	// src_agent: relates to agent
	// evtc: limited to squad outside instances
	// realtime: limited to squad

	CBTS_CHANGEUP // agent is alive at time of event
	// src_agent: relates to agent
	// evtc: limited to agent table outside instances
	// realtime: limited to squad

	CBTS_CHANGEDEAD // agent is dead at time of event
	// src_agent: relates to agent
	// evtc: limited to agent table outside instances
	// realtime: limited to squad

	CBTS_CHANGEDOWN // agent is down at time of event
	// src_agent: relates to agent
	// evtc: limited to agent table outside instances
	// realtime: limited to squad

	CBTS_SPAWN // agent entered tracking
	// src_agent: relates to agent
	// evtc: limited to agent table outside instances
	// realtime: no

	CBTS_DESPAWN // agent left tracking
	// src_agent: relates to agent
	// evtc: limited to agent table outside instances
	// realtime: no

	CBTS_HEALTHPCTUPDATE // agent health percentage changed
	// src_agent: relates to agent
	// dst_agent: percent * 10000 eg. 99.5% will be 9950
	// evtc: limited to agent table outside instances
	// realtime: no

	CBTS_SQCOMBATSTART // squad combat start, first player enter combat. previously named log start
	// value: as uint32_t, server unix timestamp
	// buff_dmg: local unix timestamp
	// evtc: yes
	// realtime: yes

	CBTS_LOGEND // squad combat stop, last player left combat. previously named log end
	// value: as uint32_t, server unix timestamp
	// buff_dmg: local unix timestamp
	// evtc: yes
	// realtime: yes

	CBTS_WEAPSWAP // agent weapon set changed
	// src_agent: relates to agent
	// dst_agent: new weapon set id
	// value: old weapon seet id
	// evtc: yes
	// realtime: yes

	CBTS_MAXHEALTHUPDATE // agent maximum health changed
	// src_agent: relates to agent
	// dst_agent: new max health
	// evtc: limited to non-players
	// realtime: no

	CBTS_POINTOFVIEW // "recording" player
	// src_agent: relates to agent
	// evtc: yes
	// realtime: no

	CBTS_LANGUAGE // text language id
	// src_agent: text language id
	// evtc: yes
	// realtime: no

	CBTS_GWBUILD // game build
	// src_agent: game build number
	// evtc: yes
	// realtime: no

	CBTS_SHARDID // server shard id
	// src_agent: shard id
	// evtc: yes
	// realtime: no

	CBTS_REWARD // wiggly box reward
	// dst_agent: reward id
	// value: reward type
	// evtc: yes
	// realtime: yes

	CBTS_BUFFINITIAL // buff application for buffs already existing at time of event
	// refer to cbtevent struct, identical to buff application. statechange is set to this
	// evtc: limited to squad outside instances
	// realtime: limited to squad

	CBTS_POSITION // agent position changed
	// src_agent: relates to agent
	// dst_agent: (float*)&dst_agent is float[3], x/y/z
	// evtc: limited to agent table outside instances
	// realtime: no

	CBTS_VELOCITY // agent velocity changed
	// src_agent: relates to agent
	// dst_agent: (float*)&dst_agent is float[3], x/y/z
	// evtc: limited to agent table outside instances
	// realtime: no

	CBTS_FACING // agent facing direction changed
	// src_agent: relates to agent
	// dst_agent: (float*)&dst_agent is float[2], x/y
	// evtc: limited to agent table outside instances
	// realtime: no

	CBTS_TEAMCHANGE // agent team id changed
	// src_agent: relates to agent
	// dst_agent: new team id
	// value: old team id
	// evtc: limited to agent table outside instances
	// realtime: limited to squad

	CBTS_ATTACKTARGET // attacktarget to gadget association
	// src_agent: relates to agent, the attacktarget
	// dst_agent: the gadget
	// evtc: limited to agent table outside instances
	// realtime: no

	CBTS_TARGETABLE // agent targetable state
	// src_agent: relates to agent
	// dst_agent: new targetable state
	// evtc: limited to agent table outside instances
	// realtime: no

	CBTS_MAPID // map id
	// src_agent: map id
	// evtc: yes
	// realtime: no

	CBTS_REPLINFO // internal use
	// internal use

	CBTS_STACKACTIVE // buff instance is now active
	// src_agent: relates to agent
	// dst_agent: buff instance id
	// value: current buff duration
	// evtc: limited to squad outside instances
	// realtime: limited to squad

	CBTS_STACKRESET // buff instance duration changed, value is the duration to reset to (also marks inactive), pad61-pad64 buff instance id
	// src_agent: relates to agent
	// value: new duration
	// evtc: limited to squad outside instances
	// realtime: limited to squad

	CBTS_GUILD // agent is a member of guild
	// src_agent: relates to agent
	// dst_agent: (uint8_t*)&dst_agent is uint8_t[16], guid of guild
	// value: new duration
	// evtc: limited to squad outside instances
	// realtime: no

	CBTS_BUFFINFO // buff information
	// skillid: skilldef id of buff
	// overstack_value: max combined duration
	// src_master_instid:
	// is_src_flanking: likely an invuln
	// is_shields: likely an invert
	// is_offcycle: category
	// pad61: buff stacking type
	// pad62: likely a resistance
	// evtc: yes
	// realtime: no

	CBTS_BUFFFORMULA // buff formula, one per event of this type
	// skillid: skilldef id of buff
	// time: (float*)&time is float[9], type attribute1 attribute2 parameter1 parameter2 parameter3 trait_condition_source trait_condition_self content_reference
	// src_instid: (float*)&src_instid is float[2], buff_condition_source buff_condition_self
	// evtc: yes
	// realtime: no

	CBTS_SKILLINFO // skill information
	// skillid: skilldef id of skill
	// time: (float*)&time is float[4], cost range0 range1 tooltiptime
	// evtc: yes
	// realtime: no

	CBTS_SKILLTIMING // skill timing, one per event of this type
	// skillid: skilldef id of skill
	// src_agent: timing type
	// dst_agent: at time since activation in milliseconds
	// evtc: yes
	// realtime: no

	CBTS_BREAKBARSTATE // agent breakbar state changed
	// src_agent: relates to agent
	// dst_agent: new breakbar state
	// evtc: limited to agent table outside instances
	// realtime: no

	CBTS_BREAKBARPERCENT // agent breakbar percentage changed
	// src_agent: relates to agent
	// value: (float*)&value is float[1], new percentage
	// evtc: limited to agent table outside instances
	// realtime: no

	CBTS_INTEGRITY // one event per message. previously named error
	// time: (char*)&time is char[32], a short null-terminated message with reason
	// evtc: yes
	// realtime: no

	CBTS_MARKER // one event per marker on an agent
	// src_agent: relates to agent
	// value: markerdef id. if value is 0, remove all markers presently on agent
	// buff: marker is a commander tag
	// evtc: limited to agent table outside instances
	// realtime: no

	CBTS_BARRIERPCTUPDATE // agent barrier percentage changed
	// src_agent: relates to agent
	// dst_agent: percent * 10000 eg. 99.5% will be 9950
	// evtc: limited to agent table outside instances
	// realtime: no

	CBTS_STATRESET // arcdps stats reset
	// src_agent: species id of agent that triggered the reset, eg boss species id
	// evtc: yes
	// realtime: yes

	CBTS_EXTENSION // for extension use. not managed by arcdps
	// evtc: yes
	// realtime: yes

	CBTS_APIDELAYED // one per cbtevent that got deemed unsafe for realtime but safe for posting after squadcombat
	// evtc: no
	// realtime: yes

	CBTS_INSTANCESTART // map instance start
	// src_agent: milliseconds ago instance was started
	// evtc: yes
	// realtime: yes

	CBTS_RATEHEALTH // tick health. previously named tickrate
	// src_agent: 25 - tickrate, when tickrate <= 20
	// evtc: yes
	// realtime: no

	CBTS_LAST90BEFOREDOWN // retired, not used since 240529+
	// retired

	CBTS_EFFECT // retired, not used since 230716+
	// retired

	CBTS_IDTOGUID // content id to guid association for volatile types
	// src_agent: (uint8_t*)&src_agent is uint8_t[16] guid of content
	// overstack_value: is of enum contentlocal
	// evtc: yes
	// realtime: no

	CBTS_LOGNPCUPDATE // log boss agent changed
	// src_agent: species id of agent
	// dst_agent: related to agent
	// value: as uint32_t, server unix timestamp
	// evtc: yes
	// realtime: yes

	CBTS_IDLEEVENT // internal use
	// internal use

	CBTS_EXTENSIONCOMBAT // for extension use. not managed by arcdps
	// assumed to be cbtevent struct, skillid will be processed as such for purpose of buffinfo/skillinfo
	// evtc: yes
	// realtime: yes

	CBTS_FRACTALSCALE // fractal scale for fractals
	// src_agent: scale
	// evtc: yes
	// realtime: no

	CBTS_EFFECT2 // play graphical effect
	// src_agent: related to agent
	// dst_agent: effect at location of agent (if applicable)
	// value: (float*)&value is float[3], location x/y/z (if not at agent location)
	// iff: (uint32_t*)&iff is uint32_t[1], effect duration
	// buffremove: (uint32_t*)&buffremove is uint32_t[1], trackable id of effect. id dst_agent and location is 0/0/0, effect was stopped
	// is_shields: (int16_t*)&is_shields is int16_t[3], orientation x/y/z, values are original*1000
	// is_flanking: effect is on a non-static platform
	// evtc: limited to agent table outside instances
	// realtime: no

	CBTS_RULESET // ruleset for self
	// src_agent: bit0: pve, bit1: wvw, bit2: pvp
	// evtc: yes
	// realtime: no

	CBTS_SQUADMARKER // squad ground markers
	// src_agent: (float*)&src_agent is float[3], x/y/z of marker location. if values are all zero or infinity, this marker is removed
	// skillid: index of marker eg. 0 is arrow
	// evtc: yes
	// realtime: no

	CBTS_ARCBUILD // arc build info
	// src_agent: (char*)&src_agent is a null-terminated string matching the full build string in arcdps.log
	// evtc: yes
	// realtime: no

	CBTS_GLIDER // glider status change
	// src_agent: related to agent
	// value: 1 deployed, 0 stowed
	// evtc: limited to agent table outside instances
	// realtime: no

	CBTS_STUNBREAK // disable stopped early
	// src_agent: related to agent
	// value: duration remaining
	// evtc: limited to agent table outside instances
	// realtime: no

	CBTS_UNKNOWN // unknown/unsupported type newer than this list maybe
)
