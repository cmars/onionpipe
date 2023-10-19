{ pkgs ? import <nixpkgs> {} }:
  pkgs.mkShell {
    nativeBuildInputs = with pkgs.buildPackages; [
      go tor openssl libevent zlib goreleaser podman podman-compose
    ];
    shellHook = ''
      export GOPATH="$HOME/.cache/gopaths/$(sha256sum <<<$(pwd) | awk '{print $1}')"
      export GOBIN="$HOME/.cache/gopaths/$(sha256sum <<<$(pwd) | awk '{print $1}')/bin"
    '';
}

