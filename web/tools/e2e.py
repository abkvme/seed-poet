"""End-to-end browser tests for the Seed Poet website (dev tool).

    uv run --with playwright --with pypdf --with mnemonic python web/tools/e2e.py [base-url]

Default base-url: http://127.0.0.1:8080 (`go run ./cmd/seedpoet serve`). Chromium + Firefox;
E2E_BROWSERS=webkit (or a comma list) picks engines. WebKit (Safari's engine) needs Ubuntu: run
it in the official Playwright container or in CI.
Generated poems are checked with the reference BIP39 library (python-mnemonic). Downloaded
PDFs and offline files go to a temp dir that is deleted at the end: they hold real seeds.
"""
import io
import os
import re
import shutil
import sys
import tempfile
from urllib.parse import urlparse

from mnemonic import Mnemonic
from playwright.sync_api import sync_playwright
from pypdf import PdfReader

BASE = (sys.argv[1] if len(sys.argv) > 1 else "http://127.0.0.1:8080").rstrip("/")
ORIGIN = urlparse(BASE).netloc
M = Mnemonic("english")
CLOCK = "(function(){var n=Date.now.bind(Date);window.__off=0;Date.now=function(){return n()+window.__off;};})();"
failures = []


def check(cond, msg):
    print(("PASS " if cond else "FAIL ") + msg)
    if not cond:
        failures.append(msg)


def poem_words(page):
    return page.eval_on_selector_all("#numbered-list li", "els => els.map(e => e.textContent)")


def wait_poem(page, n):
    for _ in range(60):
        if len(poem_words(page)) == n:
            return poem_words(page)
        page.wait_for_timeout(100)
    return poem_words(page)


def test_ua(b):
    """The browser's own user agent plus a marker, so seedpoet.com's visit counter skips test runs."""
    pg = b.new_page()
    ua = pg.evaluate("navigator.userAgent")
    pg.close()
    return ua + " SeedPoetE2E"


def browser_suite(bt, tmp):
    name = bt.name
    b = bt.launch()
    ctx = b.new_context(accept_downloads=True, user_agent=test_ua(b))
    ctx.add_init_script(CLOCK)
    page = ctx.new_page()
    errors, foreign = [], []
    page.on("console", lambda m: m.type == "error" and errors.append(m.text))
    page.on("pageerror", lambda e: errors.append(str(e)))
    # WebKit reports the service worker script download (GET /sw.js, no body) as "fetch".
    page.on("request", lambda r: (urlparse(r.url).netloc not in (ORIGIN, "")
                                  or (r.resource_type in ("fetch", "xhr") and r.url != BASE + "/sw.js"))
            and foreign.append(r.method + " " + r.url))

    page.goto(BASE + "/")
    check(poem_words(page) == [] and page.is_visible("#placeholder"), f"{name}: / starts with no poem")
    for n in (12, 24):
        page.click(f"[data-generate='{n}']")
        words = wait_poem(page, n)
        check(len(words) == n and M.check(" ".join(words)), f"{name}: {n} words button writes a valid BIP39 poem")
    check(page.url.rstrip("/") == BASE, f"{name}: stays on one page ({page.url})")
    check(re.match(r"1[45]:\d\d$", page.inner_text("#timer-value")) is not None and page.is_visible("#timer"),
          f"{name}: timer counts from 15:00 ({page.inner_text('#timer-value')})")
    check("for your security" in page.inner_text("#timer").lower(), f"{name}: security note next to timer")
    check(page.is_visible(".chains") and "Bitcoin" in page.inner_text(".chains"), f"{name}: chain tags shown")

    # Pressing again makes a new poem; switching length works in place.
    first = poem_words(page)
    page.click("[data-generate='24']")
    page.wait_for_timeout(200)
    second = poem_words(page)
    check(second != first and M.check(" ".join(second)), f"{name}: pressing again makes a different valid poem")
    page.click("[data-generate='12']")
    w12 = wait_poem(page, 12)
    check(len(w12) == 12 and M.check(" ".join(w12)), f"{name}: switch to 12 works in place")

    # Old links and app shortcuts: /24 -> /?words=24 starts a poem, URL cleaned.
    page.goto(f"{BASE}/24")
    w = wait_poem(page, 24)
    check(len(w) == 24 and page.url.rstrip("/") == BASE, f"{name}: /24 redirect starts a 24-word poem ({page.url})")
    page.click("[data-generate='12']")
    w12 = wait_poem(page, 12)

    # PDF download with clickable links.
    with page.expect_download() as d:
        page.click("#pdf-btn")
    pdf_path = tmp + f"/{name}.pdf"
    d.value.save_as(pdf_path)
    reader = PdfReader(pdf_path)
    text = reader.pages[0].extract_text()
    uris = [a.get_object()["/A"]["/URI"] for a in reader.pages[0].get("/Annots", [])]
    check(all(w in text for w in w12) and f"12. {w12[-1]}" in text, f"{name}: PDF holds the poem and numbered words")
    check("https://oxisoft.io" in uris and "https://seedpoet.com" in uris, f"{name}: PDF links {uris}")
    check("hide the poem now" in page.text_content("#hide-btn").lower(), f"{name}: hide button nudges after PDF")

    page.click("#hide-btn")
    check(poem_words(page) == [] and not page.is_visible("#timer"), f"{name}: hide now wipes the poem")

    # Drill: an offline quiz on the poem, in memory only, wiped with the poem.
    page.click("[data-generate='12']")
    w12 = wait_poem(page, 12)
    page.click("#drill-btn")
    page.wait_for_timeout(150)

    def drill_fill(wrong=0):
        labels = page.eval_on_selector_all("#poem-lines input.drill-word", "els => els.map(e => e.getAttribute('aria-label'))")
        for n, lab in enumerate(labels):
            i, k = map(int, re.findall(r"\d+", lab))
            page.fill(f"#poem-lines input[aria-label='{lab}']", "nope" if n < wrong else w12[(i - 1) * 3 + k - 1].upper())
        return len(labels)

    attrs = page.eval_on_selector_all("#poem-lines input.drill-word",
                                      "els => els.map(e => e.autocomplete + e.spellcheck + e.getAttribute('autocapitalize'))")
    check(len(attrs) == 4 and set(attrs) == {"offfalsenone"}, f"{name}: drill round 1 hides 1 word per line, no autocomplete/spellcheck {attrs}")
    check(not page.is_visible("#copy-btn") and not page.is_visible("#pdf-btn") and not page.is_visible("#numbered"),
          f"{name}: drill hides copy, PDF and the numbered list")
    drill_fill()
    page.click("#check-btn")
    check("4 of 4 right" in page.inner_text("#poem-status"), f"{name}: drill grades a perfect round ({page.inner_text('#poem-status')})")
    page.click("#check-btn")
    page.wait_for_timeout(100)
    check(drill_fill(wrong=1) == 8, f"{name}: drill round 2 hides 2 words per line")
    page.click("#check-btn")
    bad = page.eval_on_selector_all("#poem-lines input.bad", "els => els.map(e => e.value)")
    check("7 of 8 right" in page.inner_text("#poem-status") and len(bad) == 1 and bad[0] in w12,
          f"{name}: drill marks the wrong word and shows the right one")
    page.click("#drill-btn")
    check(page.eval_on_selector_all("#poem-lines input", "e => e.length") == 0 and page.is_visible("#copy-btn")
          and all(wd in page.inner_text("#poem-lines") for wd in w12), f"{name}: Show poem leaves the drill")
    page.click("#drill-btn")
    page.fill("#poem-lines input.drill-word >> nth=0", "typed")
    page.click("#hide-btn")
    check(page.eval_on_selector("#poem-lines", "e => e.textContent + e.querySelectorAll('input').length") == "0"
          and not page.is_visible("#drill-btn") and not page.is_visible("#check-btn"), f"{name}: hide now wipes the drill")
    page.click("[data-generate='12']")
    wait_poem(page, 12)
    page.click("#drill-btn")
    page.evaluate("window.__off = 15*60*1000 + 1000")
    page.wait_for_timeout(1300)
    page.evaluate("window.__off = 0")
    check(page.eval_on_selector("#poem-lines", "e => e.textContent + e.querySelectorAll('input').length") == "0",
          f"{name}: drill wiped after 15 min")
    check(page.evaluate("localStorage.length + sessionStorage.length") == 0 and page.url.rstrip("/") == BASE,
          f"{name}: drill keeps nothing in storage or the URL")

    # Back never shows an old poem.
    page.goto(f"{BASE}/")
    page.click("[data-generate='24']")
    old = wait_poem(page, 24)
    page.click(".site-nav a[href$='/about']")
    page.wait_for_load_state()
    page.go_back()
    page.wait_for_timeout(800)
    check(poem_words(page) != old, f"{name}: Back does not bring back the old poem")

    # Expiry after 15 minutes.
    page.goto(f"{BASE}/")
    page.click("[data-generate='12']")
    wait_poem(page, 12)
    page.evaluate("window.__off = 14*60*1000")
    page.wait_for_timeout(1300)
    check(len(poem_words(page)) == 12, f"{name}: still shown at 14 min")
    page.evaluate("window.__off = 15*60*1000 + 1000")
    page.wait_for_timeout(1300)
    check(poem_words(page) == [] and page.eval_on_selector("#poem-lines", "e => e.textContent") == "",
          f"{name}: wiped after 15 min")

    # Strength popover.
    page.goto(BASE + "/")
    page.click(".strength-btn")
    page.wait_for_timeout(200)
    vis = page.is_visible("#strength-info")
    page.click("#strength-info .btn")
    page.wait_for_timeout(200)
    check(vis and not page.is_visible("#strength-info"), f"{name}: strength popup opens and closes")

    check(not errors, f"{name}: no console/CSP errors {errors}")
    check(not foreign, f"{name}: no fetch/xhr or third-party requests {foreign}")

    # Desktop: the whole generator with a 24-word poem fits one screen, no scrolling.
    for vw, vh in ((1366, 660), (1280, 720), (1920, 950)):
        pg = ctx.new_page()
        pg.set_viewport_size({"width": vw, "height": vh})
        pg.goto(BASE + "/")
        pg.click("[data-generate='24']")
        wait_poem(pg, 24)
        over = pg.evaluate("document.documentElement.scrollHeight - window.innerHeight")
        check(over <= 0, f"{name}: 24-word poem page fits {vw}x{vh} without scrolling (over {over})")
        pg.click("#drill-btn")
        ws = poem_words(pg)
        for _ in range(2):  # two perfect rounds: round 3 makes every word an input
            pg.eval_on_selector_all("#poem-lines input.drill-word",
                                    "(els, ws) => els.forEach(e => { var n = e.getAttribute('aria-label').match(/\\d+/g); e.value = ws[(n[0] - 1) * 3 + (n[1] - 1)]; })", ws)
            pg.click("#check-btn")
            pg.click("#check-btn")
        check(pg.eval_on_selector_all("#poem-lines input.drill-word", "e => e.length") == 24, f"{name}: drill round 3 hides every word")
        over = pg.evaluate("document.documentElement.scrollHeight - window.innerHeight")
        check(over <= 0, f"{name}: drill on a 24-word poem fits {vw}x{vh} (over {over})")
        pg.close()

    # Small screens: no horizontal scrolling.
    for w in (320, 360, 412):
        pg = ctx.new_page()
        pg.set_viewport_size({"width": w, "height": 740})
        for path in ("/", "/about", "/wordlist"):
            pg.goto(BASE + path)
            pg.wait_for_timeout(300)
            over = pg.evaluate("document.documentElement.scrollWidth - window.innerWidth")
            check(over <= 0, f"{name}: {path} fits {w}px (overflow {over})")
        pg.close()

    # /wordlist: the whole BIP39 list, every word once, no script needed.
    pg = ctx.new_page()
    pg.goto(BASE + "/wordlist")
    listed = pg.eval_on_selector_all("ul.words li", "els => els.map(e => e.textContent.trim())")
    check(len(listed) == 2048 and len(set(listed)) == 2048 and set(listed) == set(M.wordlist),
          f"{name}: /wordlist lists 2048 unique BIP39 words ({len(listed)}, {len(set(listed))} unique)")
    check(pg.get_attribute("link[rel=canonical]", "href") == "https://seedpoet.com/wordlist", f"{name}: /wordlist canonical")
    pg.close()
    ctx.close()

    # Without JavaScript: server fallback.
    ctx = b.new_context(java_script_enabled=False, user_agent=test_ua(b))
    page = ctx.new_page()
    page.goto(f"{BASE}/")
    check(poem_words(page) == [], f"{name} noJS: GET page has no poem")
    page.click("button[formaction='/nojs/24']")
    page.wait_for_load_state()
    words = poem_words(page)
    check(len(words) == 24 and M.check(" ".join(words)), f"{name} noJS: server poem is valid")
    check(page.get_attribute("meta[http-equiv=refresh]", "content") == "900;url=/expired", f"{name} noJS: 15-min refresh")
    check(page.get_attribute("#hide-btn", "href") == "/expired", f"{name} noJS: hide link")
    ctx.close()
    b.close()


def chromium_extras(p, tmp):
    b = p.chromium.launch()
    ctx = b.new_context(accept_downloads=True, user_agent=test_ua(b))
    page = ctx.new_page()

    # Offline single file, opened from disk.
    page.goto(BASE + "/")
    with page.expect_download() as d:
        page.click("a[href='/seedpoet-offline.html'][download]:visible")
    off = tmp + "/seedpoet-offline.html"
    d.value.save_as(off)
    errs = []
    pg = ctx.new_page()
    pg.on("pageerror", lambda e: errs.append(str(e)))
    pg.goto("file://" + off)
    pg.click("[data-generate='24']")
    words = wait_poem(pg, 24)
    check(len(words) == 24 and M.check(" ".join(words)) and not errs, f"offline file: valid poem from disk {errs}")
    pg.close()

    # PWA: installable, service worker, works offline.
    if BASE.startswith("https://") or "localhost" in BASE:
        page.goto(BASE + "/")
        page.wait_for_timeout(1500)
        cdp = ctx.new_cdp_session(page)
        inst = cdp.send("Page.getInstallabilityErrors")["installabilityErrors"]
        check(inst == [], f"PWA installable {inst}")
        page.reload()
        page.wait_for_timeout(1000)
        check(page.evaluate("!!navigator.serviceWorker.controller"), "service worker controls the page")
        ctx.set_offline(True)
        page.goto(BASE + "/")
        page.click("[data-generate='24']")
        words = wait_poem(page, 24)
        check(len(words) == 24 and M.check(" ".join(words)), "works offline after first visit")
        ctx.set_offline(False)
    else:
        print("SKIP PWA checks (needs https or localhost)")
    b.close()


def main():
    tmp = tempfile.mkdtemp(prefix="seedpoet-e2e-")
    try:
        with sync_playwright() as p:
            names = [n.strip() for n in os.environ.get("E2E_BROWSERS", "chromium,firefox").split(",") if n.strip()]
            for name in names:
                browser_suite(getattr(p, name), tmp)
            if "chromium" in names:
                chromium_extras(p, tmp)
    finally:
        shutil.rmtree(tmp, ignore_errors=True)
    print("ALL OK" if not failures else f"{len(failures)} FAILED")
    sys.exit(1 if failures else 0)


if __name__ == "__main__":
    main()
