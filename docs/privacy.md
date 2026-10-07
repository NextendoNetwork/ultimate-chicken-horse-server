# Network address privacy

Do not publish the operator's PC, console, home-network or public residential IP addresses. Keep actual endpoints in private configuration, and logs/captures/account data outside Git. Examples use placeholders; protocol tests use documentation address blocks. Loopback addresses describe local interfaces and are not operator-specific addresses.

Run `go run ./cmd/check-address-privacy` before publication. It inspects tracked and nonignored new working-tree files, rejects private network address literals and accidentally selected binaries/private inputs without printing the matched values. Review commit messages, pull request descriptions and new captures separately.

Sanitizing current files does not erase older commits, forks, caches or existing clones. Coordinate any necessary historical removal separately.
