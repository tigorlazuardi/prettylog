{
  description = "Description for the project";

  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
      ];
      perSystem =
        {
          pkgs,
          ...
        }:
        {
          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              go
              gotools
              (writeShellScriptBin "doc-server" ''
                xdg-open http://localhost:6060/pkg/github.com/tigorlazuardi/prettylog/
                godoc -http=:6060
              '')
            ];

            shellHook = # sh
              ''
                export GOROOT=${pkgs.go}/share/go 
              '';
          };
        };
    };
}
