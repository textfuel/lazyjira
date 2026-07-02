{
  pkgs ? null,
  ...
}@args:
let
  pkgs =
    if builtins.isAttrs pkgs then
      pkgs
    else
      (
        let
          inherit (builtins) fromJSON readFile;
          inherit ((fromJSON (readFile ../flake.lock)).nodes) nixpkgs gomod2nix;
          fetchLocked =
            node:
            let
              inherit (node.locked)
                owner
                repo
                rev
                narHash
                ;
            in

            fetchTarball {
              url = "https://github.com/${owner}/${repo}/archive/${rev}.tar.gz";
              sha256 = narHash;

            };
        in
        import (fetchLocked nixpkgs) {
          inherit (args) system;
          overlays = [
            (import "${fetchLocked gomod2nix}/overlay.nix")
          ];
        }
      );
  buildGoApplication = args.buildGoApplication or pkgs.buildGoApplication;
  version =
    args.version or (
      if builtins.pathExists ../.git then
        builtins.substring 0 7 (pkgs.lib.commitIdFromGitRepo ../.git)
      else
        "dev"
    );
in
buildGoApplication {
  inherit version;

  pname = "lazyjira";
  src = ./..;
  modules = ../gomod2nix.toml;
  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
  ];
  subPackages = [ "cmd/lazyjira" ];
  meta = with pkgs.lib; {
    description = "Terminal UI for Jira";
    homepage = "https://github.com/textfuel/lazyjira";
    license = licenses.mit;
    mainProgram = "lazyjira";
  };
}
