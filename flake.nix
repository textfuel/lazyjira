{
  description = "Terminal UI for Jira";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };
  outputs =
    {
      nixpkgs,
      flake-utils,
      gomod2nix,
      self,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system}.extend gomod2nix.overlays.default;
        inherit (pkgs) callPackage;
      in
      {
        packages = rec {
          lazyjira = callPackage ./nix/package.nix {
            version = self.shortRev or self.dirtyShortRev or "dev";
          };
          default = lazyjira;
        };

        devShells.default = pkgs.callPackage ./nix/shell.nix { };
      }
    );
}
