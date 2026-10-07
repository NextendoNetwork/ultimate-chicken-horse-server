"""Compare native UNET PE files; mask only parsed timestamps and RSDS GUIDs.

Reads local inputs without executing them. No binary or PDB path is emitted.
Usage: python compare-unet-pe.py TESTED_DLL OFFICIAL_JUNE_ZIP
"""
import hashlib
import json
import struct
import sys
import zipfile
from pathlib import Path


def digest(data):
    return hashlib.sha256(data).hexdigest()


def inspect(data):
    def u16(offset):
        return struct.unpack_from("<H", data, offset)[0]

    def u32(offset):
        return struct.unpack_from("<I", data, offset)[0]

    if data[:2] != b"MZ":
        raise ValueError("Not a PE file")
    pe = u32(0x3C)
    if data[pe:pe + 4] != b"PE\0\0":
        raise ValueError("Missing PE signature")
    optional = pe + 24
    magic = u16(optional)
    if magic not in (0x10B, 0x20B):
        raise ValueError("Unsupported PE optional header")
    table = optional + u16(pe + 20)
    sections = []
    for i in range(u16(pe + 6)):
        base = table + i * 40
        sections.append((data[base:base + 8].rstrip(b"\0"),
                         u32(base + 12), u32(base + 16), u32(base + 20)))

    def rva_offset(rva):
        for _, address, size, start in sections:
            if address <= rva < address + size:
                return start + rva - address
        raise ValueError("RVA outside file-backed sections")

    fields = [(pe + 8, 4)]
    directories = optional + (112 if magic == 0x20B else 96)
    debug_rva, debug_size = struct.unpack_from("<II", data, directories + 6 * 8)
    if debug_size % 28:
        raise ValueError("Invalid debug directory size")
    if debug_size:
        debug = rva_offset(debug_rva)
        for i in range(debug_size // 28):
            entry = debug + 28 * i
            fields.append((entry + 4, 4))
            kind, size, _, pointer = struct.unpack_from("<IIII", data, entry + 12)
            if kind == 2 and size >= 24 and data[pointer:pointer + 4] == b"RSDS":
                fields.append((pointer + 4, 16))
    normalized = bytearray(data)
    for start, size in fields:
        if start < 0 or start + size > len(data):
            raise ValueError("Metadata field outside file")
        normalized[start:start + size] = bytes(size)
    text_sections = [digest(data[start:start + size])
                     for name, _, size, start in sections if name == b".text"]
    return bytes(normalized), fields, text_sections


def main():
    if len(sys.argv) != 3:
        raise SystemExit(__doc__)
    tested = Path(sys.argv[1]).read_bytes()
    with zipfile.ZipFile(sys.argv[2]) as archive:
        names = [n for n in archive.namelist()
                 if n == "lib/UNETServerDLL.dll" or n.endswith("/lib/UNETServerDLL.dll")]
        if len(names) != 1:
            raise ValueError("Expected one official lib/UNETServerDLL.dll")
        official = archive.read(names[0])
    left, fields_left, text_left = inspect(tested)
    right, fields_right, text_right = inspect(official)
    print(json.dumps({
        "tested_sha256": digest(tested), "official_sha256": digest(official),
        "same_size": len(tested) == len(official),
        "differing_byte_positions": (sum(a != b for a, b in zip(tested, official))
                                     if len(tested) == len(official) else None),
        "same_metadata_layout": fields_left == fields_right,
        "masked_fields": [{"offset": hex(start), "size": size}
                          for start, size in fields_left],
        "metadata_normalized_equal": left == right,
        "tested_normalized_sha256": digest(left),
        "official_normalized_sha256": digest(right),
        "tested_text_section_sha256": text_left,
        "official_text_section_sha256": text_right,
        "licensing_permission": "not established by this comparison",
    }, indent=2))


if __name__ == "__main__":
    main()
