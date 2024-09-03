using Blish_HUD;
using Blish_HUD.Content;
using Blish_HUD.Controls;
using Blish_HUD.Graphics.UI;
using Microsoft.Xna.Framework;
using Microsoft.Xna.Framework.Graphics;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using static System.Net.Mime.MediaTypeNames;

namespace EnemyCount.CountWindow
{
    public class CountContainer : IDisposable
    {
        AsyncTexture2D simpleTexture;
        StandardWindow sw;
        public FlowPanel fp;

        public CountContainer()
        {
            using (var ctx = GameService.Graphics.LendGraphicsDeviceContext())
            {
                int size = 500;
                var tex = new Texture2D(ctx.GraphicsDevice, size, size);
                var dataColors = new Color[size * size];
                var solid = new Color(0, 0, 0, 200);
                for (var i = 0; i < dataColors.Count(); i++)
                {
                    dataColors[i] = solid;
                }
                tex.SetData(0, new Rectangle(0, 0, size, size), dataColors, 0, size * size);

                simpleTexture = new AsyncTexture2D(tex);
            }

            sw = new StandardWindow(simpleTexture, new Rectangle(0, 0, 500, 500), new Rectangle(10, 10, 480, 480))
            {
                Title = "Teams",
                SavesPosition = true,
                Id = "io.arcsin.EnemyCount: count window",
                Parent = GameService.Graphics.SpriteScreen,
            };

            fp = new FlowPanel
            {
                WidthSizingMode = SizingMode.Fill,
                HeightSizingMode = SizingMode.Fill,
                FlowDirection = ControlFlowDirection.SingleTopToBottom,
                CanScroll = true,
                Parent = sw,
            };
        }

        public void Show()
        {
            sw.Show();
        }

        public void Dispose()
        {
            sw?.Dispose();
            fp?.Dispose();
        }
    }
}
