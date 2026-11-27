{
  description = "K9s - Kubernetes CLI To Manage Your Clusters In Style";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      inherit (nixpkgs) lib;

      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f: lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});

      # Keep the version in sync with the Makefile (VERSION ?= vX.Y.Z).
      version = builtins.head (
        builtins.match ".*VERSION[[:space:]]+\\?=[[:space:]]+(v[0-9.]+).*" (builtins.readFile ./Makefile)
      );

      d = self.lastModifiedDate or "19700101000000";
      date = "${builtins.substring 0 4 d}-${builtins.substring 4 2 d}-${builtins.substring 6 2 d}T${builtins.substring 8 2 d}:${builtins.substring 10 2 d}:${builtins.substring 12 2 d}Z";
      commit = self.shortRev or self.dirtyShortRev or "dirty";

      mkK9s =
        pkgs:
        pkgs.callPackage ./nix/package.nix {
          go = pkgs.go_1_26;
          inherit version date commit;
          src = lib.fileset.toSource {
            root = ./.;
            fileset = lib.fileset.unions [
              ./go.mod
              ./go.sum
              ./main.go
              ./cmd
              ./internal
            ];
          };
        };
    in
    {
      packages = forAllSystems (pkgs: rec {
        k9s = mkK9s pkgs;
        default = k9s;
      });

      apps = forAllSystems (pkgs: {
        default = {
          type = "app";
          program = lib.getExe self.packages.${pkgs.stdenv.hostPlatform.system}.k9s;
          meta.description = "Run k9s";
        };
      });

      overlays.default = final: _prev: { k9s = mkK9s final; };

      homeManagerModules = rec {
        k9s = import ./nix/hm-module.nix self;
        default = k9s;
      };

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          inputsFrom = [ self.packages.${pkgs.stdenv.hostPlatform.system}.k9s ];
          packages = with pkgs; [
            go_1_26
            gopls
            gotools
            delve
            golangci-lint
            gnumake
            kubectl
          ];
          env.CGO_ENABLED = 0;
        };
      });

      formatter = forAllSystems (pkgs: pkgs.nixfmt);
    };
}
