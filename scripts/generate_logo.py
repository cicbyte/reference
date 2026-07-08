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

Requires: ``pip install playwright`` and ``playwright install chromium``.
"""

from pathlib import Path

try:
    from playwright.sync_api import sync_playwright
except ImportError:  # pragma: no cover
    sync_playwright = None

ROOT = Path(__file__).resolve().parent.parent
ASSETS_DIR = ROOT / "gui" / "frontend" / "src" / "assets"
PUBLIC_DIR = ROOT / "gui" / "frontend" / "public"
BUILD_DIR = ROOT / "gui" / "build"
HTML_FILE = ASSETS_DIR / "logo.html"

# Sizes generated as individual PNGs.
SIZES = [256, 128, 64, 32]

# Re-export target files. All point to the 256 render which is the canonical
# high-resolution source.
LOGO_SVG = ASSETS_DIR / "logo.svg"
LOGO_PNG_256 = ASSETS_DIR / "logo-256.png"
ICON_PNG = PUBLIC_DIR / "icon.png"
APPICON_PNG = BUILD_DIR / "appicon.png"
LOGO_PNG_GENERIC = ASSETS_DIR / "logo.png"


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
            # Force the svg to render at exactly (size × size), pinned to 0,0.
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
            # Resize the SVG to the exact viewport so the element bbox matches.
            page.eval_on_selector(
                "svg",
                "(el, s) => { el.setAttribute('width', s); el.setAttribute('height', s); el.style.width = s+'px'; el.style.height = s+'px'; }",
                size,
            )
            page.wait_for_timeout(60)
            out = ASSETS_DIR / f"logo-{size}.png"
            page.locator("svg").screenshot(
                path=str(out),
                omit_background=True,  # ← transparent
            )
            print(f"  wrote {out.relative_to(ROOT)} ({size}x{size} @2x)")
            ctx.close()
        browser.close()

    # Re-export the 256 render to the three canonical locations.
    data = LOGO_PNG_256.read_bytes()
    for target in (LOGO_PNG_GENERIC, ICON_PNG, APPICON_PNG):
        target.write_bytes(data)
        print(f"  wrote {target.relative_to(ROOT)}")


def main() -> None:
    print(f"Reading design source: {HTML_FILE.relative_to(ROOT)}")
    svg = extract_svg()
    print("Generating assets:")
    write_svg(svg)
    render_pngs(svg)
    print("\nDone.")


if __name__ == "__main__":
    main()
