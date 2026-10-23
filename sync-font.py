#!/usr/bin/env python3
"""Import the reviewed, unchanged estate Inter asset and its full legal record."""
import hashlib
import json
from pathlib import Path
import sys

root = Path(__file__).resolve().parent
supplier = Path(sys.argv[1]).resolve()
record = next(x for x in json.loads((supplier / "legal/extras.json").read_text()) if x["name"] == "inter")
asset = (supplier / "frontend/src/fonts/Inter.woff2").read_bytes()
assert hashlib.sha256(asset).hexdigest() == "2dcd3612039a9aaa2660c0c9edd0f4b982e19d317af5ffe4161c37ddc803b5d0"
assert record["license"] == "OFL-1.1"
for doc in record["documents"]:
    assert hashlib.sha256(doc["text"].encode()).hexdigest() == doc["sha256"]
license_text = next(x["text"] for x in record["documents"] if x["path"] == "OFL.txt")
assert "PERMISSION & CONDITIONS" in license_text and "DISCLAIMER" in license_text
for directory in [root / "files", root / "docker/files"]:
    directory.mkdir(parents=True, exist_ok=True)
    (directory / "Inter.woff2").write_bytes(asset)
    (directory / "Inter-OFL.txt").write_text(license_text)
    (directory / "Inter-provenance.json").write_text(json.dumps(record, indent=2) + "\n")
    (directory / "FontsFree-Net-SFProText-Regular.ttf").unlink(missing_ok=True)
print("Inter bytes and full original terms verified; active SF Pro payload removed")
