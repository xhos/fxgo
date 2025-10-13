{
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.git-hooks.url = "github:cachix/git-hooks.nix";
  inputs.git-hooks.inputs.nixpkgs.follows = "nixpkgs";

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    git-hooks,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      formatter = pkgs.alejandra;

      checks.pre-commit = git-hooks.lib.${system}.run {
        src = ./.;
        hooks = {
          gotest.enable = true;
          govet.enable = true;
          alejandra.enable = true;
          golangci-lint = {
            enable = true;
            name = "golangci-lint";
            entry = "${pkgs.golangci-lint}/bin/golangci-lint fmt";
            types = ["go"];
          };
        };
      };

      devShells.default = pkgs.mkShell {
        shellHook = self.checks.${system}.pre-commit.shellHook;
        packages = with pkgs; [
          go
          air
          golangci-lint

          (writeShellScriptBin "fmt" ''
            ${golangci-lint}/bin/golangci-lint fmt
          '')

          (writeShellScriptBin "cover" ''
            go test -coverprofile=coverage.out ./... && \
            go tool cover -html=coverage.out -o coverage.html
          '')
        ];
      };
    });
}
