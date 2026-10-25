# Packaging

## Nix

The root `package.nix` uses `buildGo127Module`; `flake.nix` exposes it as
`packages.<system>.pastebox-cli` and the default package/application.
`flake.lock` pins nixpkgs. The Nix Build workflow builds and tests both supported
Linux architectures on native runners, including `pb version` after installation.

The flake uses a source-date development version and the Git revision when
available. Release builders can override `version` with the original release tag
and `revision` with the release commit when calling `package.nix`; both values
are injected into `pb version`. A plain path source has no Git revision and
reports `unknown`.

After changing Go dependencies, recalculate `vendorHash` in `package.nix`:
temporarily set it to `lib.fakeHash`, run `nix build .#pastebox-cli`, and replace
it with the actual hash from Nix's mismatch error. Never commit the placeholder.
Update nixpkgs deliberately with `nix flake update nixpkgs`, then run:

```bash
nix flake check --no-update-lock-file --print-build-logs
nix build .#pastebox-cli --no-update-lock-file
nix run . --no-update-lock-file -- version
```

Nix uses its own package definition and store paths; nFPM continues to define
the distribution archive packages. No nixpkgs registration or binary cache
publication is performed by this workflow.

## Void Linux

`void-package-build.yml` builds native `x86_64` and `aarch64` XBPS packages
for glibc and musl. It uses `packaging/nfpm.yaml` to produce a temporary DEB
payload, then creates and installs the XBPS packages in Void containers.
The combined release workflow includes all four XBPS assets in its checksums.

## AUR

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

`aur-publish.yml` publishes after a successful Release workflow on `main`, or
by manual dispatch for an existing release tag. Automated runs honor `[skip ci]`
and skip Dependabot commits. Configure `AUR_USERNAME`, `AUR_EMAIL`, and
`AUR_SSH_PRIVATE_KEY` as repository secrets before running publication.

## Workflow validation

`quality-checks.yml` validates GitHub configuration and test-only changes on
pushes and pull requests, including actionlint and the Go workflow assertions.
It has read-only permissions and does not create releases. The existing release
workflow keeps its metadata exclusions and quality gate before tagging.
