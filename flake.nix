{
  description = "Noodle: architecture maps as code";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";

  outputs = { self, nixpkgs, ... }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAll = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in {
      packages = forAll (pkgs: rec {
        noodle = pkgs.buildGoModule {
          pname = "noodle";
          version = self.shortRev or self.dirtyShortRev or "dev";
          src = self;
          subPackages = [ "cmd/noodle" ];
          # Dependencies are vendored so the plugin wrapper can build offline too.
          vendorHash = null;
          env.CGO_ENABLED = 0;
          meta.mainProgram = "noodle";
        };
        default = noodle;
      });

      devShells = forAll (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [ go gopls golangci-lint drawio-headless jetbrains-mono ];
        };
      });
    };
}
