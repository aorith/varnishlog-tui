{
  description = "Varnishlog TUI";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = {nixpkgs, ...}: let
    eachSystem = nixpkgs.lib.genAttrs ["aarch64-darwin" "aarch64-linux" "x86_64-darwin" "x86_64-linux"];
  in {
    packages = eachSystem (system: let
      pkgs = import nixpkgs {inherit system;};
    in {
      default = pkgs.buildGoModule {
        name = "varnishlog-tui";
        src =
          builtins.filterSource
          (path: type: !(baseNameOf path == "go.mod"))
          ./.;
        vendorHash = null;
      };
    });
  };
}
