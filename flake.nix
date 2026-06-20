{
  description = "Oktopus - a governed, agent-agnostic operations control plane";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };

        # Go runtime + dev tools (deps pinned in go.mod).
        goTools = [ pkgs.go pkgs.gopls pkgs.gotools pkgs.git ];

        # Python runtime only (mkdocs-material pinned in pyproject.toml).
        docsTools = [ pkgs.python3 ];
      in
      {
        # Full shell for local development (Go + docs).
        devShells.default = pkgs.mkShell {
          packages = goTools ++ docsTools;
        };

        # Targeted shells so CI jobs pull only the closure they need.
        devShells.go = pkgs.mkShell { packages = goTools; };
        devShells.docs = pkgs.mkShell { packages = docsTools; };
      });
}
