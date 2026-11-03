"""Render the site's raster images from their sources (dev tool, not used at runtime).

    uv run --with playwright --with pillow python web/tools/render_images.py [base-url]

Writes into web/static/root/: favicon.ico, icons/*.png (from favicon.svg), og.png
(from web/tools/og.html) and, when base-url is given (a running local site), the
PWA screenshots of the home page. Then rebuild with `go run ./cmd/buildweb`.
"""
import io
import pathlib
import sys

from PIL import Image
from playwright.sync_api import sync_playwright

ROOT = pathlib.Path(__file__).resolve().parents[1]
OUT = ROOT / "static" / "root"
SVG = (OUT / "favicon.svg").read_text()


def svg_page(size, pad=0.0, bg=None):
    inner = size * (1 - 2 * pad)
    off = size * pad
    bgcss = f"background:{bg};" if bg else "background:transparent;"
    return (f"<html><body style='margin:0;{bgcss}width:{size}px;height:{size}px'>"
            f"<div style='position:absolute;left:{off}px;top:{off}px;width:{inner}px;height:{inner}px'>"
            + SVG.replace('width="64" height="64"', 'width="100%" height="100%"') + "</div></body></html>")


def shot(page, html, size, path=None):
    page.set_viewport_size({"width": size, "height": size})
    page.set_content(html)
    png = page.screenshot(omit_background=True)
    if path:
        path.write_bytes(png)
    return png


def main():
    (OUT / "icons").mkdir(parents=True, exist_ok=True)
    (OUT / "screenshots").mkdir(parents=True, exist_ok=True)
    with sync_playwright() as p:
        b = p.chromium.launch()
        page = b.new_page()
        for size, name in ((192, "icon-192.png"), (512, "icon-512.png")):
            shot(page, svg_page(size), size, OUT / "icons" / name)
        # Apple touch icons get no transparency; maskable icons keep the art in the safe zone.
        shot(page, svg_page(180, 0.08, "#020a05"), 180, OUT / "icons" / "apple-touch-icon.png")
        shot(page, svg_page(512, 0.16, "#020a05"), 512, OUT / "icons" / "maskable-512.png")
        frames = [Image.open(io.BytesIO(shot(page, svg_page(s), s))).convert("RGBA") for s in (48, 32, 16)]
        frames[0].save(OUT / "favicon.ico", sizes=[(48, 48), (32, 32), (16, 16)], append_images=frames[1:])

        page.set_viewport_size({"width": 1200, "height": 630})
        page.goto((ROOT / "tools" / "og.html").as_uri())
        page.wait_for_timeout(300)
        page.screenshot(path=str(OUT / "og.png"))

        if len(sys.argv) > 1:
            base = sys.argv[1].rstrip("/")
            for w, h, name in ((1280, 720, "wide.png"), (412, 915, "narrow.png")):
                ctx = b.new_context(viewport={"width": w, "height": h}, reduced_motion="reduce")
                pg = ctx.new_page()
                pg.goto(base + "/")
                pg.wait_for_timeout(500)
                pg.screenshot(path=str(OUT / "screenshots" / name))
                ctx.close()
        b.close()
    print("images written to", OUT)


if __name__ == "__main__":
    main()
