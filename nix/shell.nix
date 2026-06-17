{
  mkShell,
  go,
  gomod2nix,
}:
mkShell {
  name = "lazyjira-dev";
  packages = [
    go
    gomod2nix
  ];
}
