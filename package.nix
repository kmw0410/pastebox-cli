{
  lib,
  buildGo127Module,
  cacert,
  makeWrapper,
  version ? "dev",
  revision ? "unknown",
}:

let
  # Match Go's Linux CA bundle search; leave its certificate directories intact.
  caSetup = ''
    if [ -z "''${SSL_CERT_FILE:-}" ] && [ -z "''${SSL_CERT_DIR:-}" ]; then
      pastebox_system_ca=
      for pastebox_ca in /etc/ssl/certs/ca-certificates.crt /etc/pki/tls/certs/ca-bundle.crt /etc/ssl/ca-bundle.pem /etc/pki/tls/cacert.pem /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem /etc/ssl/cert.pem; do
        if [ -r "$pastebox_ca" ] && [ -s "$pastebox_ca" ]; then
          pastebox_system_ca=1
          break
        fi
      done
      if [ -z "$pastebox_system_ca" ]; then
        export SSL_CERT_FILE=${cacert}/etc/ssl/certs/ca-bundle.crt
      fi
      unset pastebox_system_ca pastebox_ca
    fi
  '';
in
buildGo127Module {
  pname = "pastebox-cli";
  inherit version;

  src = lib.cleanSourceWith {
    src = ./.;
    filter =
      path: type:
      lib.cleanSourceFilter path type
      && !(builtins.elem (baseNameOf path) [
        "WORK.md"
        "result"
        "dist"
        "vendor"
      ])
      && !(type == "regular" && baseNameOf path == "pastebox-cli");
  };

  vendorHash = "sha256-fxxp7ECuMMmLw7L5/lPi7lfmJM2p+Te9AqXm5Xed3G0=";
  nativeBuildInputs = [ makeWrapper ];
  subPackages = [ "." ];
  env.CGO_ENABLED = "0";
  buildFlags = [ "-buildvcs=false" ];

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
    "-X main.commit=${revision}"
  ];

  postInstall = ''
    mv "$out/bin/pastebox-cli" "$out/bin/pb"
    install -Dm644 LICENSE "$out/share/doc/pastebox-cli/LICENSE"
    install -m644 package.md package_ko.md "$out/share/doc/pastebox-cli/"
    install -Dm644 packaging/completions/pb.bash "$out/share/bash-completion/completions/pb"
    install -Dm644 packaging/completions/_pb "$out/share/zsh/site-functions/_pb"
    install -Dm644 packaging/completions/pb.fish "$out/share/fish/vendor_completions.d/pb.fish"
  '';

  postFixup = ''
    wrapProgram "$out/bin/pb" --run ${lib.escapeShellArg caSetup}
  '';

  doInstallCheck = true;
  installCheckPhase = ''
    runHook preInstallCheck
    "$out/bin/pb" version > version-output.txt
    grep -Fx ${lib.escapeShellArg "pb ${version} (commit ${revision})"} version-output.txt
    "$out/bin/pb" completion bash > bash-completion.txt
    cmp bash-completion.txt "$out/share/bash-completion/completions/pb"
    "$out/bin/pb" completion zsh > zsh-completion.txt
    cmp zsh-completion.txt "$out/share/zsh/site-functions/_pb"
    "$out/bin/pb" completion fish > fish-completion.txt
    cmp fish-completion.txt "$out/share/fish/vendor_completions.d/pb.fish"
    runHook postInstallCheck
  '';

  meta = {
    description = "Terminal client for the Pastebox text sharing service";
    homepage = "https://github.com/kmw0410/pastebox-cli";
    license = lib.licenses.mit;
    mainProgram = "pb";
    platforms = [
      "x86_64-linux"
      "aarch64-linux"
    ];
  };
}
