"""生成 TT Desktop 的应用图标（圆角方块 + 黑色 TT）。

用法（需要 Pillow）：
    python desktop/build/make-icon.py

产物写在脚本同目录：
    icon.ico   给 Windows（exe / 快捷方式 / 窗口）—— 16 到 256 多尺寸
    icon-1024.png  设计稿（预览、以后要改尺寸时从它出发）

设计约束（改之前先读）：
  · 字母是**几何笔画**画的，不依赖字体 —— 16px 下靠字体渲染的 T 会糊成一团，
    而这里横竖都是整数矩形，缩到多小都锐利；
  · 方块是浅色、字母是黑色：反过来的话小尺寸下白字会被抗锯齿吃掉；
  · 方块留 3% 外边距：Windows 任务栏/桌面不会给图标留白，贴边的方块看起来比别家的大一圈。
"""

from pathlib import Path

from PIL import Image, ImageDraw

SIZE = 1024
MARGIN = 30          # 3% 外边距
RADIUS = 210         # 圆角半径（22% 左右，接近 Windows 11 的方角观感）
CARD = (247, 247, 247, 255)
EDGE = (214, 214, 214, 255)
INK = (17, 17, 17, 255)

CAP = 430            # 字母高度
STEM = 100           # 竖笔画宽
BAR = 100            # 横笔画高
TW = 292             # 单个 T 的宽度
GAP = 56             # 两个 T 之间的缝（再宽就念成「T T」，不是「TT」）
OPTICAL_DY = 8       # 视觉居中：横笔画在上方，纯几何居中看起来偏高，往下挪一点

OUT = Path(__file__).resolve().parent


def main() -> None:
    img = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
    d = ImageDraw.Draw(img)

    d.rounded_rectangle(
        [MARGIN, MARGIN, SIZE - MARGIN, SIZE - MARGIN],
        radius=RADIUS,
        fill=CARD,
        outline=EDGE,
        width=6,
    )

    total = TW * 2 + GAP
    x0 = (SIZE - total) // 2
    y0 = (SIZE - CAP) // 2 + OPTICAL_DY
    for i in range(2):
        x = x0 + i * (TW + GAP)
        d.rectangle([x, y0, x + TW, y0 + BAR], fill=INK)
        cx = x + TW // 2
        d.rectangle([cx - STEM // 2, y0, cx + STEM // 2, y0 + CAP], fill=INK)

    img.save(OUT / "icon-1024.png")
    img.save(
        OUT / "icon.ico",
        format="ICO",
        sizes=[(16, 16), (24, 24), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)],
    )
    print("wrote", OUT / "icon.ico")


if __name__ == "__main__":
    main()
