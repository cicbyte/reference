#!/usr/bin/env python3
"""Generate logo assets from the HTML design source.

Workflow
--------
``logo.html`` (under ``gui/frontend/src/assets/``) is the single design source:
edit the ``<svg class="design-svg">`` inside it to change the logo. This script
then:

1. Extracts that SVG markup → writes ``assets/logo.svg``.
2. Renders it at several sizes on a **transparent** background and screenshots
   each into PNG files:
     - ``assets/logo-<size>.png``       (in-app usage, e.g. sidebar)
     - ``assets/logo.png``              (= the 256 variant, convenience copy)
     - ``public/icon.png``              (browser favicon, 256)
     - ``build/appicon.png``            (Wails desktop app icon, 256)
3. Builds ``build/windows/icon.ico`` (multi-size: 16→256) — embedded into the
   Windows .exe by Wails, shown in the taskbar / window title / Alt-Tab.

Requires: ``pip install playwright`` and ``playwright install chromium``.
Optional but default-on: ``pip install pillow`` (for the .ico).
"""

from pathlib import Path

try:
    from playwright.sync_api import sync_playwright
except ImportError:  # pragma: no cover
    sync_playwright = None

try:
    from PIL import Image
except ImportError:  # pragma: no cover
    Image = None

ROOT = Path(__file__).resolve().parent.parent
ASSETS_DIR = ROOT / "gui" / "frontend" / "src" / "assets"
PUBLIC_DIR = ROOT / "gui" / "frontend" / "public"
BUILD_DIR = ROOT / "gui" / "build"
WINDOWS_DIR = BUILD_DIR / "windows"
HTML_FILE = ASSETS_DIR / "logo.html"

# Sizes generated as individual PNGs.
SIZES = [256, 128, 64, 32]

# Sizes embedded in the Windows .ico (smaller ones shown in the taskbar /
# window title at different DPIs; 256 for Alt-Tab / explorer large icons).
ICO_SIZES = [16, 24, 32, 48, 64, 128, 256]

# Re-export target files. All point to the 256 render which is the canonical
# high-resolution source.
LOGO_SVG = ASSETS_DIR / "logo.svg"
LOGO_PNG_256 = ASSETS_DIR / "logo-256.png"
ICON_PNG = PUBLIC_DIR / "icon.png"
APPICON_PNG = BUILD_DIR / "appicon.png"
LOGO_PNG_GENERIC = ASSETS_DIR / "logo.png"
ICO_FILE = WINDOWS_DIR / "icon.ico"


def extract_svg() -> str:
    """Pull the design SVG markup out of logo.html via Playwright.

    We load the page and read ``outerHTML`` of the first
    ``<svg class="design-svg">`` element. This is robust against the same
    string appearing in comments or CSS (which broke the earlier text-scan
    approach) because the browser only returns real DOM nodes.
    """
    if sync_playwright is None:
        raise SystemExit("playwright is required (pip install playwright)")
    url = HTML_FILE.resolve().as_uri()
    with sync_playwright() as p:
        browser = p.chromium.launch()
        page = browser.new_page()
        page.goto(url)
        page.wait_for_load_state("networkidle")
        svg = page.eval_on_selector(
            'svg.design-svg',
            'el => el.outerHTML',
        )
        browser.close()
    # Normalise the root tag: drop the design-only class so logo.svg is clean.
    svg = svg.replace(' class="design-svg"', "", 1).strip()
    return svg


def write_svg(svg: str) -> None:
    LOGO_SVG.write_text(svg + "\n", encoding="utf-8")
    print(f"  wrote {LOGO_SVG.relative_to(ROOT)}")


def _bold_svg() -> str:
    """A chunky, high-fill variant for tiny taskbar sizes (16/24/32/48).

    The regular logo is line-based (open chevrons + hollow diamond) and at
    16×16 the strokes collapse to sub-pixel slivers, so the icon looks tiny and
    faint. This variant replaces the strokes with **solid** thick chevron bars
    and a large filled diamond, roughly doubling the pixel coverage so the mark
    reads clearly at a glance in the taskbar.
    """
    # viewBox cropped tight to the geometry so it fills the square frame.
    return (
        '<svg viewBox="28 44 200 168" xmlns="http://www.w3.org/2000/svg">'
        # left solid chevron "<"  (thick bar, ~30 wide)
        '<path d="M104 72 L60 128 L104 184 L74 184 L30 128 L74 72 Z" fill="#0d9488"/>'
        # right solid chevron ">"
        '<path d="M152 72 L196 128 L152 184 L182 184 L226 128 L182 72 Z" fill="#0d9488"/>'
        # large filled centre diamond (focal node)
        '<path d="M128 88 L168 128 L128 168 L88 128 Z" fill="#2dd4bf"/>'
        '</svg>'
    )


def _render_one(browser, svg: str, size: int, out: Path) -> None:
    """Render ``svg`` at exactly ``size``×``size`` (transparent) → ``out``."""
    page_html = (
        "<!DOCTYPE html><html><head><meta charset='utf-8'>"
        "<style>html,body{margin:0;padding:0;background:transparent;"
        "overflow:hidden;}svg{display:block;}</style></head>"
        f"<body>{svg}</body></html>"
    )
    ctx = browser.new_context(
        viewport={"width": size, "height": size},
        device_scale_factor=2,  # crisp on hi-dpi
    )
    page = ctx.new_page()
    page.set_content(page_html)
    page.wait_for_load_state("networkidle")
    page.eval_on_selector(
        "svg",
        "(el, s) => { el.setAttribute('width', s); el.setAttribute('height', s);"
        " el.style.width = s+'px'; el.style.height = s+'px'; }",
        size,
    )
    page.wait_for_timeout(60)
    page.locator("svg").screenshot(path=str(out), omit_background=True)
    ctx.close()


def render_pngs(svg: str) -> None:
    """Screenshot the SVG at each size with a transparent background.

    We render one standalone page per size where the SVG is the *only* element
    sized exactly to the viewport, then screenshot the element directly.
    Element screenshots are robust against any coordinate offset and ignore
    surrounding page chrome.
    """
    if sync_playwright is None:
        print("  [skip] playwright not installed (pip install playwright)")
        return

    PUBLIC_DIR.mkdir(parents=True, exist_ok=True)
    BUILD_DIR.mkdir(parents=True, exist_ok=True)

    with sync_playwright() as p:
        browser = p.chromium.launch()
        for size in SIZES:
            out = ASSETS_DIR / f"logo-{size}.png"
            _render_one(browser, svg, size, out)
            print(f"  wrote {out.relative_to(ROOT)} ({size}x{size} @2x)")
        # Taskbar-sized PNGs for the .ico: a chunky variant that stays legible
        # at 16/24/32. Stored in build/ (not bundled in the app).
        for size in ICO_SIZES:
            out = WINDOWS_DIR / f"_ico-{size}.png"
            _render_one(browser, _bold_svg(), size, out)
        browser.close()

    # Re-export the 256 render to the three canonical locations.
    data = LOGO_PNG_256.read_bytes()
    for target in (LOGO_PNG_GENERIC, ICON_PNG, APPICON_PNG):
        target.write_bytes(data)
        print(f"  wrote {target.relative_to(ROOT)}")


def write_ico() -> None:
    """Build a multi-size Windows .ico from the 256 master PNG.

    The .ico is what Wails embeds into the compiled .exe via the Windows
    resource file in ``build/windows/`` — it drives the taskbar, window title,
    Alt-Tab thumbnail and Explorer icon.

    Small sizes (≤48, what the taskbar actually shows) are rendered from a
    dedicated chunky variant (``_bold_svg``) so the mark stays legible at
    16×16; larger sizes use the regular logo downscaled from the 256 master.
    """
    if Image is None:
        print("  [skip] Pillow not installed (pip install pillow) — .ico skipped")
        return
    if not LOGO_PNG_256.exists():
        print(f"  [skip] {LOGO_PNG_256.relative_to(ROOT)} missing — run PNG render first")
        return

    WINDOWS_DIR.mkdir(parents=True, exist_ok=True)
    master = Image.open(LOGO_PNG_256).convert("RGBA")

    frames = []
    for s in ICO_SIZES:
        bold_png = WINDOWS_DIR / f"_ico-{s}.png"
        if bold_png.exists():
            frames.append(Image.open(bold_png).convert("RGBA"))
        else:
            frames.append(master.resize((s, s), Image.LANCZOS))
        # tidy up the temp render
        if bold_png.exists():
            bold_png.unlink()

    frames[0].save(
        str(ICO_FILE),
        format="ICO",
        sizes=[(s, s) for s in ICO_SIZES],
        append_images=frames[1:],
    )
    print(f"  wrote {ICO_FILE.relative_to(ROOT)} (sizes {ICO_SIZES})")


def main() -> None:
    print(f"Reading design source: {HTML_FILE.relative_to(ROOT)}")
    svg = extract_svg()
    print("Generating assets:")
    write_svg(svg)
    render_pngs(svg)
    write_ico()
    print("\nDone.")


if __name__ == "__main__":
    main()
