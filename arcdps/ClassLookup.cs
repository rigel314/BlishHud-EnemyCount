using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace EnemyCount.arcdps
{
    internal class ClassLookup
    {
        public static string Table(uint profession, uint elite)
        {
            return profession switch
            {
                1 => GuardianLookup(elite),
                2 => WarriorLookup(elite),
                3 => EngineerLookup(elite),
                4 => RangerLookup(elite),
                5 => ThiefLookup(elite),
                6 => ElementalistLookup(elite),
                7 => MesmerLookup(elite),
                8 => NecromancerLookup(elite),
                9 => RevenantLookup(elite),
                _ => "unknown-" + profession + "-" + elite,
            };
        }

        // to find the magic numbers for elite mapping:
        // for i in {1..72}; do curl -fsSL "https://api.guildwars2.com/v2/specializations/$i" > specilization_$i.json; done
        // cat *.json | jq '.|select(.elite == true)|"\(.profession): \(.id): \(.name)"'|sort

        // "Guardian"
        private static string GuardianLookup(uint elite)
        {
            return elite switch
            {
                0 => "Guardian",
                27 => "Dragonhunter",
                62 => "Firebrand",
                65 => "Willbender",
                _ => "unknown-Guardian-" + elite,
            };
        }
        // "Warrior"
        private static string WarriorLookup(uint elite)
        {
            return elite switch
            {
                0 => "Warrior",
                18 => "Berserker",
                61 => "Spellbreaker",
                68 => "Bladesworn",
                _ => "unknown-Warrior-" + elite,
            };
        }
        // "Engineer"
        private static string EngineerLookup(uint elite)
        {
            return elite switch
            {
                0 => "Engineer",
                43 => "Scrapper",
                57 => "Holosmith",
                70 => "Mechanist",
                _ => "unknown-Engineer-" + elite,
            };
        }
        // "Ranger"
        private static string RangerLookup(uint elite)
        {
            return elite switch
            {
                0 => "Ranger",
                5 => "Druid",
                55 => "Soulbeast",
                72 => "Untamed",
                _ => "unknown-Ranger-" + elite,
            };
        }
        // "Thief"
        private static string ThiefLookup(uint elite)
        {
            return elite switch
            {
                0 => "Thief",
                7 => "Daredevil",
                58 => "Deadeye",
                71 => "Specter",
                _ => "unknown-Thief-" + elite,
            };
        }
        // "Elementalist"
        private static string ElementalistLookup(uint elite)
        {
            return elite switch
            {
                0 => "Elementalist",
                48 => "Tempest",
                56 => "Weaver",
                67 => "Catalyst",
                _ => "unknown-Elementalist-" + elite,
            };
        }
        // "Mesmer"
        private static string MesmerLookup(uint elite)
        {
            return elite switch
            {
                0 => "Mesmer",
                40 => "Chronomancer",
                59 => "Mirage",
                66 => "Virtuoso",
                _ => "unknown-Mesmer-" + elite,
            };
        }
        // "Necromancer"
        private static string NecromancerLookup(uint elite)
        {
            return elite switch
            {
                0 => "Necromancer",
                34 => "Reaper",
                60 => "Scourge",
                64 => "Harbinger",
                _ => "unknown-Necromancer-" + elite,
            };
        }
        // "Revenant"
        private static string RevenantLookup(uint elite)
        {
            return elite switch
            {
                0 => "Revenant",
                52 => "Herald",
                63 => "Renegade",
                69 => "Vindicator",
                _ => "unknown-Revenant-" + elite,
            };
        }
    }
}
