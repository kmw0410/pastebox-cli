pkgname=pastebox-cli
pkgver=26.10.20
pkgrel=2
pkgdesc="Lightweight command-line client for self-hosted Pastebox servers"
arch=('x86_64' 'aarch64')
url="https://github.com/kmw0410/pastebox-cli"
license=('MIT')
makedepends=('go>=1.26.10')
_tag=v26.10.20
_commit=56a60a3
source=("${pkgname}-${pkgver}.tar.gz::${url}/archive/refs/tags/${_tag}.tar.gz")
sha256sums=('3413f448dde5a06aa964cecf40b55346a3dbbb0b8e2a2e945d318f9f089e7ab9')

build() {
  cd "${pkgname}-${_tag#v}"

  CGO_ENABLED=0 go build \
    -buildvcs=false \
    -trimpath \
    -ldflags "-s -w -X main.version=${_tag} -X main.commit=${_commit}" \
    -o pb .

  # Generate from the archived release binary so older tags remain rebuildable.
  ./pb completion bash > pb.bash
  ./pb completion zsh > _pb
  ./pb completion fish > pb.fish
}

check() {
  cd "${pkgname}-${_tag#v}"

  CGO_ENABLED=0 go test -buildvcs=false ./...
}

package() {
  cd "${pkgname}-${_tag#v}"

  install -Dm755 pb "${pkgdir}/usr/bin/pb"
  install -Dm644 pb.bash "${pkgdir}/usr/share/bash-completion/completions/pb"
  install -Dm644 _pb "${pkgdir}/usr/share/zsh/site-functions/_pb"
  install -Dm644 pb.fish "${pkgdir}/usr/share/fish/vendor_completions.d/pb.fish"
  install -Dm644 LICENSE "${pkgdir}/usr/share/licenses/${pkgname}/LICENSE"
  install -Dm644 package.md "${pkgdir}/usr/share/doc/${pkgname}/package.md"
  install -Dm644 package_ko.md "${pkgdir}/usr/share/doc/${pkgname}/package_ko.md"
}
