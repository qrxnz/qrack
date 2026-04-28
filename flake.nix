{
  inputs.utils.url = "github:numtide/flake-utils";

  outputs = {
    self,
    nixpkgs,
    utils,
  }:
    utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {inherit system;};
      in rec {
        packages.default = pkgs.buildGoModule rec {
          pname = "qrack";
          version = "2.0.6";
          src = self;

          vendorHash = "sha256-GqAk9SdbBMGGo6IQp7CMi5LjWf/IFB897vcd4XC867k=";
        };

        defaultPackage = packages.default;

        devShells.default = pkgs.mkShell rec {
          buildInputs = with pkgs; [
            # Go
            go
            gopls
            delve

            # Formatters
            treefmt
            mdformat
            alejandra
            prettier

            # Others
            go-task
          ];
        };
      }
    );
}
