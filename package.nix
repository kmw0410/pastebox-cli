{
  lib,
  buildGo127Module,
  version ? "dev",
  revision ? "unknown",
}:

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
