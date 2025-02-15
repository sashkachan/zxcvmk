{
  description = "";

  inputs.utils.url = "github:numtide/flake-utils";
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  inputs.gomod2nix = {
    url = "github:tweag/gomod2nix";
    inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = { self, nixpkgs, utils, gomod2nix }: 
    utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs {
        inherit system;
        overlays = [ gomod2nix.overlays.default ];
      };

      in {
        devShells.default = with pkgs;
          mkShell {
            buildInputs = [
              gopls 
              delve 
              go 
              gore 
              go-tools 
              golangci-lint 
              gomod2nix.packages.${system}.default
            ];
          };

        packages.default = pkgs.buildGoApplication rec {
          pname = "zxcvmk";
          version = "1.0.0";
          src = ./.;
          modules = ./gomod2nix.toml;
          meta = {
            mainProgram = "cmd/zxcvmk";
          };
        };
      });
}
