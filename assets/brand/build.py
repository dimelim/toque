import math
import os
import sys

OUT = os.path.dirname(os.path.abspath(__file__))
R, r, S = 76, 40, 36
TOP, CY, BASE, ASC, DESC = 44, 120, 196, 0, 256


def f(v):
    return f"{v:.2f}".rstrip("0").rstrip(".")


def circle(cx, cy, rad, cw=True):
    sweep = 1 if cw else 0
    return (f"M{f(cx - rad)} {f(cy)}A{f(rad)} {f(rad)} 0 1 {sweep} {f(cx + rad)} {f(cy)}"
            f"A{f(rad)} {f(rad)} 0 1 {sweep} {f(cx - rad)} {f(cy)}Z")


def ring(cx, cy, ocx=None, ocy=None, ir=r):
    ocx = cx if ocx is None else ocx
    ocy = cy if ocy is None else ocy
    return circle(cx, cy, R, True) + circle(ocx, ocy, ir, False)


def rect(x0, y0, x1, y1):
    return f"M{f(x0)} {f(y0)}H{f(x1)}V{f(y1)}H{f(x0)}Z"


def letter_t(x):
    stem = rect(x + 24, ASC, x + 24 + S, BASE)
    bar = rect(x, TOP, x + 96, TOP + 32)
    return [stem, bar]


def letter_o(cx):
    return [ring(cx, CY)]


def letter_q(cx, offset=0, cut=True):
    bowl = ring(cx, CY, cx - offset, CY - offset, r - offset * 0.2)
    x0, x1 = cx + r, cx + R
    if cut:
        stem = f"M{f(x0)} {f(CY)}H{f(x1)}V{f(DESC)}L{f(x0)} {f(DESC - S * 0.8)}Z"
    else:
        stem = rect(x0, CY, x1, DESC)
    return [bowl, stem]


def letter_u(cx):
    p = (f"M{f(cx - R)} {f(TOP)}V{f(CY)}"
         f"A{f(R)} {f(R)} 0 0 0 {f(cx)} {f(BASE)}"
         f"H{f(cx + R)}V{f(TOP)}H{f(cx + r)}V{f(CY)}"
         f"A{f(r)} {f(r)} 0 0 1 {f(cx - r)} {f(CY)}V{f(TOP)}Z")
    return [p]


def letter_e(cx):
    h = S * 0.44
    pa = math.asin(h / R)
    ax, ay = cx + R * math.cos(pa), CY + h
    pb = math.radians(40)
    bx, by = cx + R * math.cos(pb), CY + R * math.sin(pb)
    bix, biy = cx + r * math.cos(pb), CY + r * math.sin(pb)
    cix = cx - math.sqrt(r * r - h * h)
    outer = (f"M{f(ax)} {f(ay)}A{f(R)} {f(R)} 0 1 0 {f(bx)} {f(by)}"
             f"L{f(bix)} {f(biy)}A{f(r)} {f(r)} 0 0 1 {f(cix)} {f(CY + h)}H{f(ax)}Z")
    dx = math.sqrt(r * r - h * h)
    eye = f"M{f(cx - dx)} {f(CY - h)}A{f(r)} {f(r)} 0 0 1 {f(cx + dx)} {f(CY - h)}Z"
    return [outer + eye]


def svg(paths, w, h, fill="#0B0D0A", vb=None, bg=None, rule="evenodd"):
    vb = vb or f"0 0 {f(w)} {f(h)}"
    body = ""
    if bg:
        body += f'<rect width="100%" height="100%" fill="{bg}"/>'
    body += "".join(f'<path fill="{fill}" fill-rule="{rule}" d="{d}"/>' for d in paths)
    return f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="{vb}" width="{f(w)}" height="{f(h)}">{body}</svg>'


def wordmark(offset=0, cut=True):
    paths = []
    paths += letter_t(0)
    paths += letter_o(188)
    paths += letter_q(360, offset, cut)
    paths += letter_u(540)
    paths += letter_e(714)
    return paths


def symbol_paths(offset=0, cut=True):
    cx = 112
    paths = letter_q(cx, offset, cut)
    return paths


def write(name, text):
    with open(os.path.join(OUT, name), "w", encoding="utf-8") as fh:
        fh.write(text)


def icon(offset, fg, bg, name, radius=56):
    q = symbol_paths(offset)
    inner = "".join(f'<path fill="{fg}" fill-rule="evenodd" d="{d}"/>' for d in q)
    scale = 0.80
    tx = 128 - (112 + 0) * scale
    ty = 128 - 150 * scale
    text = (f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" width="256" height="256">'
            f'<rect width="256" height="256" rx="{radius}" fill="{bg}"/>'
            f'<g transform="translate({f(tx)} {f(ty)}) scale({scale})">{inner}</g></svg>')
    write(name, text)


INK, PAPER, VOLT = "#0B0D0A", "#F3F5EE", "#D4FF3A"

if __name__ == "__main__":
    for old in os.listdir(OUT):
        if old.endswith(".svg"):
            os.remove(os.path.join(OUT, old))
    vb = "-8 -8 808 272"
    write("wordmark.svg", svg(wordmark(), 800, 256, fill=INK, vb=vb))
    write("wordmark-light.svg", svg(wordmark(), 800, 256, fill=PAPER, vb=vb))
    write("wordmark-volt.svg", svg(wordmark(), 800, 256, fill=VOLT, vb=vb))
    write("symbol.svg", svg(symbol_paths(), 256, 256, fill=INK, vb="28 36 168 228"))
    write("symbol-volt.svg", svg(symbol_paths(), 256, 256, fill=VOLT, vb="28 36 168 228"))
    icon(0, INK, VOLT, "icon.svg")
    icon(0, VOLT, INK, "icon-dark.svg")
    print("ok")
