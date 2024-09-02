using Blish_HUD;
using Blish_HUD.ArcDps;
using Blish_HUD.ArcDps.Models;
using Blish_HUD.Modules;
using Blish_HUD.Modules.Managers;
using Blish_HUD.Settings;
using Microsoft.Xna.Framework;
using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.ComponentModel.Composition;
using System.Threading.Tasks;

namespace EnemyCount
{
    [Export(typeof(Blish_HUD.Modules.Module))]
    public class Module : Blish_HUD.Modules.Module
    {
        private static readonly Logger Logger = Logger.GetLogger<Module>();

        private SettingCollection settings_teamID;
        private SettingEntry<string> setting_teamID_red;
        private SettingEntry<string> setting_teamID_green;
        private SettingEntry<string> setting_teamID_blue;

        struct ClassCount
        {
            public int Count;
            public string Class;
        }
        struct Team
        {
            public int Total;
            public ClassCount[] Counts;
        }
        struct Session
        {
            public ConcurrentDictionary<ushort, Team> Teams; // key is teamID
        }
        private ConcurrentDictionary<DateTime, Session> Sessions;
        private ulong activeSessionID;
        private Dictionary<ulong, Ag> active = new Dictionary<ulong, Ag>();

        #region Service Managers
        internal SettingsManager SettingsManager => this.ModuleParameters.SettingsManager;
        internal ContentsManager ContentsManager => this.ModuleParameters.ContentsManager;
        internal DirectoriesManager DirectoriesManager => this.ModuleParameters.DirectoriesManager;
        internal Gw2ApiManager Gw2ApiManager => this.ModuleParameters.Gw2ApiManager;
        #endregion

        [ImportingConstructor]
        public Module([Import("ModuleParameters")] ModuleParameters moduleParameters) : base(moduleParameters) { }

        private SettingValidationResult validateNum(string x)
        {
            if (UInt16.TryParse(x, out ushort y))
            {
                return new SettingValidationResult(true);
            }
            return new SettingValidationResult(false, "could not parse uint16");
        }

        protected override void DefineSettings(SettingCollection settings)
        {
            settings_teamID = settings.AddSubCollection("Team ID Map", false);
            settings_teamID.RenderInUi = true;
            setting_teamID_red = settings_teamID.DefineSetting("teamID_red", "706", () => "Red TeamID", () => "arcdps reported teamID for red team");
            setting_teamID_green = settings_teamID.DefineSetting("teamID_green", "2763", () => "Green TeamID", () => "arcdps reported teamID for green team");
            setting_teamID_blue = settings_teamID.DefineSetting("teamID_blue", "432", () => "Blue TeamID", () => "arcdps reported teamID for blue team");

            setting_teamID_red.SetValidation(validateNum);
        }

        protected override void Initialize()
        {

        }

        private void handleArcDpsEvents(object sender, RawCombatEventArgs args)
        {
            //if (args.CombatEvent.Id != activeSessionID)
            //{
            //    Console.WriteLine(active);
            //    active = new Dictionary<ulong, Ag>();
            //}

            if (args.CombatEvent.Src != null)
            {
                if (!active.ContainsKey(args.CombatEvent.Src.Id))
                {
                    active.Add(args.CombatEvent.Src.Id, args.CombatEvent.Src);
                }
            }

            if (args.CombatEvent.Dst != null)
            {
                if (!active.ContainsKey(args.CombatEvent.Dst.Id))
                {
                    active.Add(args.CombatEvent.Dst.Id, args.CombatEvent.Dst);
                }
            }
        }

        protected override async Task LoadAsync()
        {
            ArcDpsService.ArcDps.RawCombatEvent += handleArcDpsEvents;
        }

        protected override void OnModuleLoaded(EventArgs e)
        {
            // Base handler must be called
            base.OnModuleLoaded(e);
        }

        protected override void Update(GameTime gameTime)
        {

        }

        /// <inheritdoc />
        protected override void Unload()
        {
            // Unload here

            // All static members must be manually unset
        }

    }

}
