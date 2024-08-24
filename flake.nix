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
        packages.default = pkgs.stdenv.mkDerivation {
          pname = "zxcvmk";
          version = "1.0.0";
          src = ./.;

          buildInputs = [ pkgs.go ];

          buildPhase = ''
            mkdir -p $out/bin
            go build -o $out/bin/zxcvmk ./cmd/main.go
          '';

          meta = with pkgs.lib; {
            mainProgram = "zxcvmk";
            description = "zxcvmk";
            license = licenses.mit;
            maintainers = [ maintainers.yourself ];
          };
        };
      }
    );
}
