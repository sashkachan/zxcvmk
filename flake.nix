{
  description = "";

  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default =
          with pkgs;
          mkShell {
            buildInputs = [
              gopls
              delve
              go
              gotests
              gomodifytags
              gore
              go-tools
              golangci-lint
            ];
            hardeningDisable = [ "all" ];
            shellHook = ''
              echo Welcome to zxcvmk devshell!
              echo To build and run the project:
              echo "go run cmd/main.go"
            '';
          };
      }
    );
}