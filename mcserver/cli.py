from __future__ import annotations

import argparse
import getpass
import sys
import tempfile
from pathlib import Path
from typing import Optional, Tuple

from .curseforge import CurseForgeClient
from .config import AppConfig, config_path
from .errors import MissingApiKeyError, UserFacingError
from .download import download_to
from .fs_ops import (
    detect_pack_root,
    extract_zip,
    copy_tree_contents,
    update_from_pack_root,
)
from .state import ServerState, utc_now_iso


def _looks_like_url(value: str) -> bool:
    return value.startswith("http://") or value.startswith("https://")


def _looks_like_pack_id(value: str) -> bool:
    return value.isdigit()


def _is_server_dir(server_dir: Path) -> bool:
    return (server_dir / "server.properties").exists()


def _resolve_pack_id(
    cf: CurseForgeClient,
    *,
    source: Optional[str],
    server_dir: Path,
    use_saved: bool,
    use_arg: bool,
    no_prompt: bool,
) -> Tuple[int, Optional[ServerState]]:
    saved = ServerState.load(server_dir)
    saved_pack_id = saved.pack_id if saved else None

    arg_pack_id: Optional[int] = None
    if source:
        if _looks_like_url(source):
            arg_pack_id = cf.resolve_pack_id_from_url(source)
        elif _looks_like_pack_id(source):
            arg_pack_id = int(source)
        else:
            raise UserFacingError(
                "SOURCE must be a digits-only modpack id or a CurseForge modpack URL."
            )

    if saved_pack_id and arg_pack_id and saved_pack_id != arg_pack_id:
        if use_saved and use_arg:
            raise UserFacingError("Choose only one of --use-saved or --use-arg.")
        if use_saved:
            return int(saved_pack_id), saved
        if use_arg:
            return int(arg_pack_id), saved
        if no_prompt or not sys.stdin.isatty():
            raise UserFacingError(
                f"This folder is configured for packId={saved_pack_id} but you provided {arg_pack_id}. "
                "Re-run with --use-saved or --use-arg (or drop --no-prompt)."
            )
        print(
            f"This folder is configured for packId={saved_pack_id} but you provided {arg_pack_id}."
        )
        choice = input("Use [s]aved or [a]rg? (s/a): ").strip().lower()
        if choice.startswith("s"):
            return int(saved_pack_id), saved
        if choice.startswith("a"):
            return int(arg_pack_id), saved
        raise UserFacingError("Invalid choice.")

    if arg_pack_id is not None:
        return int(arg_pack_id), saved

    if saved_pack_id is not None:
        return int(saved_pack_id), saved

    raise UserFacingError(
        "No SOURCE provided and no saved packId found in .mcserver/state.json."
    )


def cmd_cf_resolve(args: argparse.Namespace) -> int:
    cf = CurseForgeClient()
    pack_id = cf.resolve_pack_id_from_url(args.url)
    print(pack_id)
    return 0


def cmd_cf_search(args: argparse.Namespace) -> int:
    cf = CurseForgeClient()
    results = cf.search_modpacks(
        query=args.query, game_version=args.game_version, page_size=args.limit
    )
    for item in results:
        print(f"{item.get('id')}\t{item.get('name')}")
    return 0


def cmd_cf_files(args: argparse.Namespace) -> int:
    cf = CurseForgeClient()
    files = cf.list_files(int(args.pack_id))
    for f in files[: args.limit]:
        if args.server_only and not (f.is_server_pack or f.server_pack_file_id):
            continue
        print(
            f"{f.id}\t{f.file_date}\t{f.display_name}\tserverPack={f.is_server_pack}\tserverPackFileId={f.server_pack_file_id}"
        )
    return 0


def cmd_cf_download_url(args: argparse.Namespace) -> int:
    cf = CurseForgeClient()
    url, server_file_id, display_name = cf.resolve_server_pack_download(
        int(args.pack_id), file_id=args.file_id
    )
    if args.verbose:
        print(f"serverPackFileId={server_file_id}\tdisplayName={display_name}")
    print(url)
    return 0


def _install_or_update(
    *,
    server_dir: Path,
    source: Optional[str],
    file_id: Optional[int],
    accept_eula: bool,
    use_saved: bool,
    use_arg: bool,
    no_prompt: bool,
    check_only: bool,
) -> int:
    cf = CurseForgeClient()
    pack_id, saved_state = _resolve_pack_id(
        cf,
        source=source,
        server_dir=server_dir,
        use_saved=use_saved,
        use_arg=use_arg,
        no_prompt=no_prompt,
    )

    mode_update = _is_server_dir(server_dir)

    url, server_file_id, display_name = cf.resolve_server_pack_download(
        pack_id, file_id=file_id
    )

    if check_only and mode_update:
        installed = saved_state.installed_file_id if saved_state else None
        if installed == server_file_id:
            print("Up to date.")
            return 0
        print(f"Update available: installed={installed} latest={server_file_id}")
        return 0

    with tempfile.TemporaryDirectory(prefix="mcserver_") as tmp:
        tmp_path = Path(tmp)
        zip_path = tmp_path / "serverpack.zip"
        extracted = tmp_path / "extracted"
        extracted.mkdir(parents=True, exist_ok=True)

        print(f"Downloading server pack: {display_name}")
        download_to(url, zip_path)
        extract_zip(zip_path, extracted)
        pack_root = detect_pack_root(extracted)

        if mode_update:
            # Safety check already implied by mode_update
            update_from_pack_root(pack_root, server_dir)
            print("Update complete.")
        else:
            copy_tree_contents(pack_root, server_dir)
            print("Install complete.")

    if accept_eula:
        (server_dir / "eula.txt").write_text("eula=true\n", encoding="utf-8")

    new_state = saved_state or ServerState()
    new_state.pack_id = pack_id
    new_state.installed_file_id = server_file_id
    new_state.installed_display_name = display_name
    new_state.last_updated_at = utc_now_iso()
    new_state.save(server_dir)

    return 0


def cmd_install(args: argparse.Namespace) -> int:
    server_dir = Path(args.dir).resolve()
    return _install_or_update(
        server_dir=server_dir,
        source=args.source,
        file_id=args.file_id,
        accept_eula=args.accept_eula,
        use_saved=args.use_saved,
        use_arg=args.use_arg,
        no_prompt=args.no_prompt,
        check_only=False,
    )


def cmd_update(args: argparse.Namespace) -> int:
    # Interchangeable: update is just an alias to install with the same options.
    server_dir = Path(args.dir).resolve()
    return _install_or_update(
        server_dir=server_dir,
        source=args.source,
        file_id=args.file_id,
        accept_eula=args.accept_eula,
        use_saved=args.use_saved,
        use_arg=args.use_arg,
        no_prompt=args.no_prompt,
        check_only=args.check_only,
    )


def cmd_status(args: argparse.Namespace) -> int:
    server_dir = Path(args.dir).resolve()
    state = ServerState.load(server_dir)
    if not state or not state.pack_id:
        raise UserFacingError("No .mcserver/state.json found in this folder.")
    print(f"packId={state.pack_id}")
    print(f"installedFileId={state.installed_file_id}")
    print(f"installedDisplayName={state.installed_display_name}")
    print(f"lastUpdatedAt={state.last_updated_at}")
    return 0


def cmd_config_set_api_key(args: argparse.Namespace) -> int:
    api_key = args.api_key
    if not api_key:
        if not sys.stdin.isatty():
            raise UserFacingError(
                "No API key provided. Pass it as an argument or run interactively."
            )
        api_key = getpass.getpass("CurseForge API key: ").strip()
    if not api_key:
        raise UserFacingError("API key cannot be empty.")

    cfg = AppConfig.load()
    cfg.curseforge_api_key = api_key
    cfg.save()
    print(f"Saved CurseForge API key to {config_path()}")
    return 0


def cmd_config_path(args: argparse.Namespace) -> int:
    print(str(config_path()))
    return 0


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="mcserver")
    sub = parser.add_subparsers(dest="cmd", required=True)

    p_install = sub.add_parser("install", help="Smart install/update in a directory")
    p_install.add_argument(
        "source", nargs="?", help="Modpack ID (digits) or CurseForge modpack URL"
    )
    p_install.add_argument("--dir", default=".")
    p_install.add_argument("--file-id", type=int, default=None)
    p_install.add_argument("--accept-eula", action="store_true")
    p_install.add_argument("--use-saved", action="store_true")
    p_install.add_argument("--use-arg", action="store_true")
    p_install.add_argument("--no-prompt", action="store_true")
    p_install.set_defaults(func=cmd_install)

    p_update = sub.add_parser("update", help="Alias of install (interchangeable)")
    p_update.add_argument(
        "source", nargs="?", help="Modpack ID (digits) or CurseForge modpack URL"
    )
    p_update.add_argument("--dir", default=".")
    p_update.add_argument("--file-id", type=int, default=None)
    p_update.add_argument("--accept-eula", action="store_true")
    p_update.add_argument("--check-only", action="store_true")
    p_update.add_argument("--use-saved", action="store_true")
    p_update.add_argument("--use-arg", action="store_true")
    p_update.add_argument("--no-prompt", action="store_true")
    p_update.set_defaults(func=cmd_update)

    p_status = sub.add_parser("status", help="Show saved pack/version for a directory")
    p_status.add_argument("--dir", default=".")
    p_status.set_defaults(func=cmd_status)

    p_cf = sub.add_parser("cf", help="CurseForge helper commands")
    cf_sub = p_cf.add_subparsers(dest="cf_cmd", required=True)

    p_cf_resolve = cf_sub.add_parser("resolve", help="Resolve modpack URL to pack ID")
    p_cf_resolve.add_argument("url")
    p_cf_resolve.set_defaults(func=cmd_cf_resolve)

    p_cf_search = cf_sub.add_parser("search", help="Search modpacks")
    p_cf_search.add_argument("query")
    p_cf_search.add_argument("--game-version", default=None)
    p_cf_search.add_argument("--limit", type=int, default=10)
    p_cf_search.set_defaults(func=cmd_cf_search)

    p_cf_files = cf_sub.add_parser("files", help="List modpack files")
    p_cf_files.add_argument("pack_id")
    p_cf_files.add_argument("--server-only", action="store_true")
    p_cf_files.add_argument("--limit", type=int, default=20)
    p_cf_files.set_defaults(func=cmd_cf_files)

    p_cf_dl = cf_sub.add_parser(
        "download-url", help="Resolve direct download URL for server pack"
    )
    p_cf_dl.add_argument("pack_id")
    p_cf_dl.add_argument("--file-id", type=int, default=None)
    p_cf_dl.add_argument("--verbose", action="store_true")
    p_cf_dl.set_defaults(func=cmd_cf_download_url)

    p_config = sub.add_parser(
        "config", help="Persist and inspect local mcserver config"
    )
    cfg_sub = p_config.add_subparsers(dest="config_cmd", required=True)

    p_cfg_set = cfg_sub.add_parser(
        "set-api-key", help="Save CurseForge API key to config file"
    )
    p_cfg_set.add_argument(
        "api_key", nargs="?", help="If omitted, you will be prompted"
    )
    p_cfg_set.set_defaults(func=cmd_config_set_api_key)

    p_cfg_path = cfg_sub.add_parser("path", help="Print the config file path")
    p_cfg_path.set_defaults(func=cmd_config_path)

    return parser


def main(argv: Optional[list[str]] = None) -> int:
    argv = argv if argv is not None else sys.argv[1:]

    parser = build_parser()
    args = parser.parse_args(argv)

    try:
        return int(args.func(args))
    except MissingApiKeyError as e:
        print(str(e), file=sys.stderr)
        return 2
    except UserFacingError as e:
        print(str(e), file=sys.stderr)
        return 2
    except KeyboardInterrupt:
        print("Interrupted", file=sys.stderr)
        return 130


if __name__ == "__main__":
    raise SystemExit(main())
