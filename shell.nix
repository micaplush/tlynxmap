{ pkgs ? import <nixpkgs> { }, ... }:

pkgs.mkShellNoCC {
  buildInputs = with pkgs; [
    go
    delve
  ];
}
