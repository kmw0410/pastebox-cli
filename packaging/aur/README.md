# AUR Packaging

This directory prepares the source-based `pastebox-cli` AUR package. It is
not an AUR repository and does not publish or push anything to the AUR.

## Update for a release

1. Set `_tag` in `PKGBUILD` to the exact Git release tag.
2. Set `pkgver` to the tag without `v`, replacing each `-` with `.`.
3. Set `_commit` to the short commit ID referenced by the tag.
4. Reset `pkgrel` to `1` for a new upstream version.
5. Update the source checksum and regenerate `.SRCINFO`:

   ```bash
   updpkgsums
   makepkg --printsrcinfo > .SRCINFO
   ```

Keep the original tag in `_tag` so `pb version` matches GitHub Releases while
the normalized `pkgver` remains valid for an Arch package version.

## Validate locally

Run these commands from this directory on an Arch Linux system:

```bash
makepkg --verifysource
makepkg --cleanbuild
makepkg --printsrcinfo > .SRCINFO
namcap PKGBUILD
namcap pastebox-cli-*.pkg.tar.zst
./pkg/pastebox-cli/usr/bin/pb version
```

`namcap` is an optional validation dependency. Before a future AUR submission,
copy only `PKGBUILD` and `.SRCINFO` into the separate AUR Git repository and
add the maintainer comment expected for that repository.
