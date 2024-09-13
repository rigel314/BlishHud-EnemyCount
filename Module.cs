using Blish_HUD;
using Blish_HUD.ArcDps;
using Blish_HUD.ArcDps.Models;
using Blish_HUD.Content;
using Blish_HUD.Controls;
using Blish_HUD.Modules;
using Blish_HUD.Modules.Managers;
using Blish_HUD.Overlay.UI.Views;
using Blish_HUD.Settings;
using EnemyCount.arcdps;
using Microsoft.Xna.Framework;
using Microsoft.Xna.Framework.Content;
using Microsoft.Xna.Framework.Graphics;
using SharpDX.MediaFoundation;
using System;
using System.Collections;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.ComponentModel.Composition;
using System.IO;
using System.Linq;
using System.Net.Sockets;
using System.Text.Json;
using System.Threading;
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
        private Dictionary<string, blar> active = new Dictionary<string, blar>();
        private Mutex activeLock = new Mutex();
        private Dictionary<string, blar> display = new Dictionary<string, blar>();
        private ConcurrentDictionary<string, blar> latest = new ConcurrentDictionary<string, blar>();

        public struct blar
        {
            public Ag ag { get; set; }
            public RawCombatEventArgs args { get; set; }
        }

        CountWindow.CountContainer cc;

        volatile bool reset = false;
        volatile bool resetDisplay = false;
        volatile bool logStart = false;
        volatile bool logEnd = false;
        volatile bool logEnded = false;
        int startctr = 0;

        StreamWriter logFileWriter;

        Label statusLabel;

        CornerIcon ci;

        volatile SocketError se = SocketError.Success;

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
            logFileWriter.WriteLine(JsonSerializer.Serialize(args));

            if (args.CombatEvent.Ev == null)
            {
                return;
            }

            if (args.CombatEvent.Ev.IsStateChange == ArcDpsEnums.StateChange.LogStart)
            {
                if (startctr == 0)
                {
                    //Console.WriteLine(active);
                    active = new Dictionary<string, blar>();
                    //reset = true;
                    latest.Clear();
                    logStart = true;
                }
                startctr++;
            }

            if (args.CombatEvent.Ev.IsStateChange == ArcDpsEnums.StateChange.LogEnd)
            {
                if (startctr != 0)
                {
                    startctr--;
                }
                if (startctr == 0)
                {
                    //latest = active;
                    resetDisplay = true;
                    foreach (var x in active)
                    {
                        latest.TryAdd(x.Key, x.Value);
                    }
                    logEnd = true;
                    logFileWriter.Flush();
                }
                // TODO: update the sessions dictionary from the active one
            }

            //activeLock.WaitOne();
            if (args.CombatEvent.Src != null && args.CombatEvent.Src.Elite != 0xffffffff && args.CombatEvent.Src.Profession != 0)
            {
                var key = args.CombatEvent.Src.Team + "." + args.CombatEvent.Ev.SrcInstId;
                if (!active.ContainsKey(key))
                {
                    active.Add(key, new blar() { ag = args.CombatEvent.Src, args = args });
                }
            }

            if (args.CombatEvent.Dst != null && args.CombatEvent.Dst.Elite != 0xffffffff && args.CombatEvent.Dst.Profession != 0)
            {
                var key = args.CombatEvent.Dst.Team + "." + args.CombatEvent.Ev.DstInstId;
                if (!active.ContainsKey(key))
                {
                    active.Add(key, new blar() { ag = args.CombatEvent.Dst, args = args });
                }
            }
            //activeLock.ReleaseMutex();
        }

        private void handleArcDpsErrors(object sender, SocketError args)
        {
            se = args;
        }

        protected override async Task LoadAsync()
        {
            logFileWriter = File.AppendText("C:\\temp\\enemyCount.log");

            ArcDpsService.ArcDps.RawCombatEvent += handleArcDpsEvents;
            ArcDpsService.ArcDps.Error += handleArcDpsErrors;

            cc = new CountWindow.CountContainer();
            new StandardButton()
            {
                Text = "Reset",
                Parent = cc.fp2,
            }.Click += (s, e) =>
            {
                reset = true;
            };

            statusLabel = new Label()
            {
                AutoSizeWidth = true,
                Parent = cc.fp2,
            };

            ci = new CornerIcon()
            {
                Icon = ContentService.Content.GetTexture(@"fallback.png"),
                HoverIcon = ContentService.Content.GetTexture(@"fallback.png"),
                BasicTooltipText = "Toggle Counts",
                Priority = 0x7621e8bc,
                Parent = GameService.Graphics.SpriteScreen,
            };
            ci.Click += (s, e) =>
            {
                cc.Toggle();
            };
        }

        protected override void OnModuleLoaded(EventArgs e)
        {
            // Base handler must be called
            base.OnModuleLoaded(e);
        }

        protected override void Update(GameTime gameTime)
        {
            statusLabel.Text = "adrp:" + ArcDpsService.ArcDps.RenderPresent + ",adr:" + ArcDpsService.ArcDps.Running + ",err:" + se;

            if (reset)
            {
                foreach (var ch in cc.fp.GetChildrenOfType<Label>())
                {
                    ch.Parent = null;
                }
                display.Clear();
                reset = false;
            }

            if (resetDisplay)
            {
                display.Clear();
                resetDisplay = false;
            }

            if (logStart)
            {
                //new Label()
                //{
                //    Text = "logStart",
                //    AutoSizeWidth = true,
                //    Parent = cc.fp,
                //};
                Logger.Debug("logStart");
                logStart = false;
            }
            if (logEnd)
            {
                //new Label()
                //{
                //    Text = "logEnd",
                //    AutoSizeWidth = true,
                //    Parent = cc.fp,
                //};

                Logger.Debug("logEnd");
                logEnd = false;
                logEnded = true;
            }

            //activeLock.WaitOne();
            var tmpLatestDict = latest;
            foreach (var kv in tmpLatestDict)
            {
                var value = kv.Value;
                if (!display.ContainsKey(kv.Key) && value.ag.Team != 2763)
                {
                    if (logEnded)
                    {
                        new Label()
                        {
                            Text = "_______",
                            AutoSizeWidth = true,
                            Parent = cc.fp,
                        };
                        logEnded = false;
                        Logger.Debug(JsonSerializer.Serialize(latest));
                    }
                    var x = value.ag.Team +
                        ":" + ClassLookup.Table(value.ag.Profession, value.ag.Elite) +
                        "-" + +value.ag.Id +
                        //"-" + value.ev.SrcAgent + 
                        "-" + value.args.CombatEvent.Ev.SrcInstId +
                        //"-" + value.ev.DstAgent + 
                        "-" + value.args.CombatEvent.Ev.DstInstId;
                    if (value.ag.Name != "")
                    {
                        x += " (" + value.ag.Name + ")";
                    }
                    new Label()
                    {
                        Text = x,
                        AutoSizeWidth = true,
                        Parent = cc.fp,
                    };
                    Logger.Debug(x);
                    display.Add(kv.Key, value);
                }
            }
            //activeLock.ReleaseMutex();
        }

        /// <inheritdoc />
        protected override void Unload()
        {
            ArcDpsService.ArcDps.RawCombatEvent -= handleArcDpsEvents;
            ArcDpsService.ArcDps.Error -= handleArcDpsErrors;
            ci.Dispose();
            cc.Dispose();
        }

    }

}
