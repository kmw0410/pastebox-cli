{
  description = "Pastebox CLI";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          package = pkgs.callPackage ./package.nix {
            version = "0-unstable-${builtins.substring 0 8 (self.lastModifiedDate or "19700101")}";
            revision = self.rev or self.dirtyRev or "unknown";
          };
        in
        {
          pastebox-cli = package;
          default = package;
        }
      );

      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.pastebox-cli}/bin/pb";
          meta.description = "Pastebox terminal client";
        };
      });

      checks = forAllSystems (system: {
        package = self.packages.${system}.pastebox-cli;
      });
    };
}
