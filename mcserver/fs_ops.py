from __future__ import annotations

import os
import shutil
import zipfile
from pathlib import Path


REPLACE_DIRS = ("mods", "config", "scripts", "kubejs", "libraries", "defaultconfigs")


def detect_pack_root(extracted_dir: Path) -> Path:
    # Look for a directory containing a 'mods' folder within depth 2.
    for root, dirs, _files in os.walk(extracted_dir):
        rel = Path(root).relative_to(extracted_dir)
        if len(rel.parts) > 2:
            continue
        if "mods" in dirs:
            return Path(root)
    return extracted_dir


def extract_zip(zip_path: Path, dest_dir: Path) -> None:
    with zipfile.ZipFile(zip_path, "r") as zf:
        zf.extractall(dest_dir)


def copy_tree_contents(src_dir: Path, dest_dir: Path) -> None:
    dest_dir.mkdir(parents=True, exist_ok=True)
    for child in src_dir.iterdir():
        target = dest_dir / child.name
        if child.is_dir():
            if target.exists():
                shutil.rmtree(target)
            shutil.copytree(child, target)
        else:
            shutil.copy2(child, target)


def update_from_pack_root(pack_root: Path, server_dir: Path) -> None:
    # Replace modpack-managed directories
    for d in REPLACE_DIRS:
        src = pack_root / d
        if not src.exists() or not src.is_dir():
            continue
        dest = server_dir / d
        if dest.exists():
            shutil.rmtree(dest)
        shutil.copytree(src, dest)

    # Copy top-level executables (*.jar, *.sh, *.bat) except user_jvm_args.txt
    for child in pack_root.iterdir():
        if not child.is_file():
            continue
        name = child.name
        if name == "user_jvm_args.txt":
            continue
        if name.endswith(".jar") or name.endswith(".sh") or name.endswith(".bat"):
            shutil.copy2(child, server_dir / name)
