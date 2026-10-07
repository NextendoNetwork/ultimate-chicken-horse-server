"""Reject residential LAN address examples in indexed files without printing IPs."""
import re
import subprocess
import sys


def main():
    names = subprocess.check_output(["git", "ls-files", "-z"]).split(b"\0")
    failures = []
    pattern = re.compile(rb"\b192\.168\.\d{1,3}\.\d{1,3}\b")
    for raw in names:
        if not raw:
            continue
        name = raw.decode("utf-8")
        data = subprocess.check_output(["git", "show", ":" + name])
        if pattern.search(data):
            failures.append(name)
    for name in failures:
        print("FAIL: personal LAN address literal in " + name, file=sys.stderr)
    if failures:
        return 1
    print("Indexed address privacy check passed. Review real public IPs and publication metadata separately.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
