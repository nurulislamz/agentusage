#!/usr/bin/env python3
"""
Extracts PlantUML diagrams from docs/COMMAND_FLOW_DIAGRAMS.md and renders them to docs/diagrams/*.svg.
"""

import os
import re
import subprocess
import sys
import urllib.request
import zlib

DOC_PATHS = [
    "docs/COMMAND_FLOW_DIAGRAMS.md",
    "docs/user-flow-diagram.md",
]
DIAGRAMS_DIR = "docs/diagrams"
PLANTUML_JAR = os.environ.get("PLANTUML_JAR", "/tmp/plantuml.jar")


def plantuml_encode(text: str) -> str:
    """Encode PlantUML text using deflate and PlantUML custom base64 alphabet."""
    zlibbed = zlib.compress(text.encode("utf-8"))[2:-4]
    alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
    res = []
    i = 0
    while i < len(zlibbed):
        b1 = zlibbed[i]
        b2 = zlibbed[i + 1] if i + 1 < len(zlibbed) else 0
        b3 = zlibbed[i + 2] if i + 2 < len(zlibbed) else 0

        c1 = b1 >> 2
        c2 = ((b1 & 0x3) << 4) | (b2 >> 4)
        c3 = ((b2 & 0xF) << 2) | (b3 >> 6)
        c4 = b3 & 0x3F

        res.append(alphabet[c1 & 0x3F])
        res.append(alphabet[c2 & 0x3F])
        if i + 1 < len(zlibbed):
            res.append(alphabet[c3 & 0x3F])
        if i + 2 < len(zlibbed):
            res.append(alphabet[c4 & 0x3F])
        i += 3
    return "".join(res)


def render_with_jar(jar_path: str, src: str) -> str:
    proc = subprocess.run(
        ["java", "-jar", jar_path, "-tsvg", "-pipe"],
        input=src.encode("utf-8"),
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=True,
    )
    return proc.stdout.decode("utf-8")


def render_with_server(src: str) -> str:
    enc = plantuml_encode(src)
    url = f"https://www.plantuml.com/plantuml/svg/~1{enc}"
    req = urllib.request.Request(url, headers={"User-Agent": "Mozilla/5.0 agentusage-diagram-generator"})
    with urllib.request.urlopen(req, timeout=10) as resp:
        return resp.read().decode("utf-8")


def ensure_plantuml_jar() -> str:
    if os.path.exists(PLANTUML_JAR):
        return PLANTUML_JAR
    try:
        print(f"Downloading plantuml.jar to {PLANTUML_JAR}...")
        urllib.request.urlretrieve(
            "https://github.com/plantuml/plantuml/releases/latest/download/plantuml.jar",
            PLANTUML_JAR,
        )
        return PLANTUML_JAR
    except Exception as e:
        print(f"Warning: could not download plantuml.jar ({e}), falling back to online server.")
        return ""


def extract_diagrams(doc_text: str):
    pattern = re.compile(
        r'!\[(?P<title>[^\]]+)\]\(diagrams/(?P<filename>[^)]+\.svg)\)'
        r'.*?```plantuml\s*\n(?P<puml>@startuml.*?@enduml)\s*\n```',
        re.DOTALL,
    )
    return list(pattern.finditer(doc_text))


def main():
    check_only = "--check" in sys.argv
    cli_files = [arg for arg in sys.argv[1:] if not arg.startswith("--")]
    target_docs = cli_files if cli_files else DOC_PATHS

    matches = []
    for doc in target_docs:
        if not os.path.exists(doc):
            print(f"Warning: {doc} not found, skipping.")
            continue
        with open(doc, "r", encoding="utf-8") as f:
            doc_text = f.read()
        doc_matches = extract_diagrams(doc_text)
        print(f"Found {len(doc_matches)} PlantUML diagrams in {doc}")
        matches.extend(doc_matches)

    if not matches:
        print(f"Error: No diagrams found across target docs: {target_docs}")
        sys.exit(1)

    os.makedirs(DIAGRAMS_DIR, exist_ok=True)

    jar_path = ensure_plantuml_jar() if not check_only else ""

    errors = 0
    for m in matches:
        title = m.group("title")
        filename = m.group("filename")
        puml = m.group("puml")
        out_path = os.path.join(DIAGRAMS_DIR, filename)

        if check_only:
            if not os.path.exists(out_path):
                print(f"FAIL: Missing diagram asset: {out_path}")
                errors += 1
                continue
            with open(out_path, "r", encoding="utf-8") as f:
                content = f.read()
            if "<svg" not in content or "</svg>" not in content or "Syntax Error" in content:
                print(f"FAIL: Corrupt/invalid SVG: {out_path}")
                errors += 1
            else:
                print(f"OK: {filename} exists and is valid SVG ({len(content)} bytes)")
            continue

        print(f"Rendering {filename} ({title})...")
        svg = ""
        try:
            if jar_path and os.path.exists(jar_path):
                svg = render_with_jar(jar_path, puml)
            else:
                svg = render_with_server(puml)
        except Exception as err:
            print(f"ERROR rendering {filename}: {err}")
            errors += 1
            continue

        if "<svg" not in svg or "Syntax Error" in svg:
            print(f"ERROR: rendered output for {filename} is not valid SVG or has syntax errors!")
            errors += 1
            continue

        with open(out_path, "w", encoding="utf-8") as f:
            f.write(svg)
        print(f"Saved {out_path} ({len(svg)} bytes)")

    if errors > 0:
        print(f"\nCompleted with {errors} error(s)")
        sys.exit(1)

    print(f"\nAll {len(matches)} diagrams verified/rendered successfully.")


if __name__ == "__main__":
    main()
