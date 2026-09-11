# AUR

The repository-root `PKGBUILD` and `.SRCINFO` define the source-based AUR
package. For a new release:

1. Set `_tag` to the exact Git release tag.
2. Set `pkgver` to the tag without `v`, replacing `-` with `.`.
3. Set `_commit` to the release commit's short ID.
4. Reset `pkgrel` to `1`.
5. Refresh the checksum and `.SRCINFO`.

```bash
updpkgsums
makepkg --printsrcinfo > .SRCINFO
```

Validate the package on an Arch Linux system before copying only `PKGBUILD` and
`.SRCINFO` to the separate AUR repository.

```bash
makepkg --verifysource
makepkg --cleanbuild
namcap PKGBUILD
```
