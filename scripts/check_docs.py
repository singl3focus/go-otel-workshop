"""Check local Markdown links and the two-language documentation contract.

Проверяет локальные Markdown-ссылки и правила двуязычной документации.
"""
from pathlib import Path
from urllib.parse import unquote
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]


def tracked_markdown():
    # EN: Include new unignored files so the check also works before a commit.
    # RU: Учитываем новые неигнорируемые файлы и до создания коммита.
    names = subprocess.check_output(
        ["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"],
        cwd=ROOT,
    ).decode("utf-8").split("\0")
    return sorted({ROOT / name for name in names if name.endswith(".md")})


def anchors(text):
    counts = {}
    result = set()
    for line in text.splitlines():
        heading = re.match(r"^#{1,6}\s+(.+?)\s*#*\s*$", line)
        if not heading:
            continue
        slug = heading.group(1).strip().lower()
        slug = re.sub(r"[^\w\- ]", "", slug).replace(" ", "-")
        count = counts.get(slug, 0)
        counts[slug] = count + 1
        result.add(f"{slug}-{count}" if count else slug)
    return result


def code_blocks(text):
    blocks = re.findall(r"```([^\n]*)\n(.*?)```", text, re.S)
    return [
        (lang.strip(), "\n".join(line for line in code.splitlines() if not line.lstrip().startswith("#")))
        for lang, code in blocks if lang.strip() in {"sh", "bash", "powershell", "go"}
    ]


def check():
    errors = []
    files = tracked_markdown()
    for path in files:
        text = path.read_text(encoding="utf-8")
        # EN: Ignore fenced examples when extracting prose links.
        # RU: При проверке ссылок пропускаем блоки примеров кода.
        prose = re.sub(r"```.*?```", "", text, flags=re.S)
        for target in re.findall(r"\]\(([^)]+)\)", prose):
            target = unquote(target.split(' "', 1)[0].strip("<>"))
            if re.match(r"^[a-zA-Z][\w+.-]*:", target) or target.startswith("//"):
                continue
            name, _, fragment = target.partition("#")
            dest = (path.parent / name).resolve() if name else path
            if not dest.exists():
                errors.append(f"{path.relative_to(ROOT)}: missing target {target}")
            elif fragment and dest.suffix == ".md" and fragment not in anchors(dest.read_text(encoding="utf-8")):
                errors.append(f"{path.relative_to(ROOT)}: missing anchor {target}")
        if ".github" not in path.relative_to(ROOT).parts and not path.name.endswith(".ru.md"):
            ru = path.with_name(path.stem + ".ru.md")
            if not ru.exists():
                errors.append(f"{path.relative_to(ROOT)}: missing Russian counterpart")
                continue
            other = ru.read_text(encoding="utf-8")
            for document, body in [(path, text), (ru, other)]:
                if f"]({path.name})" not in body or f"]({ru.name})" not in body:
                    errors.append(f"{document.relative_to(ROOT)}: missing language switch")
            if code_blocks(text) != code_blocks(other):
                errors.append(f"{path.relative_to(ROOT)}: English/Russian runnable blocks differ")
    for error in errors:
        print(error, file=sys.stderr)
    if errors:
        return 1
    print(f"Documentation OK: {len(files)} Markdown files, local links and paired commands checked.")
    return 0


if __name__ == "__main__":
    raise SystemExit(check())
