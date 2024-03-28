{
  description = "Dayton Flake Development Shell";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    systems.url = "github:nix-systems/default";
  };

  outputs = inputs @ { self, flake-parts, systems, ... }: flake-parts.lib.mkFlake { inherit inputs; } {
    systems = import systems;

    imports = [
      inputs.flake-parts.flakeModules.easyOverlay
    ];

    perSystem = { self', inputs', config, pkgs, system, lib, ... }: {
      _module.args.pkgs = import self.inputs.nixpkgs {
        inherit system;
        overlays = [
          self.overlays.default
        ];
      };

      apps = {
        daytona = {
          type = "app";
          program = lib.getExe pkgs.daytona-dev;
        };
      };

      packages = {
        daytona-dev = pkgs.callPackage ./nix/pkgs/daytona {
          src = lib.cleanSource ./.;
          version = self.rev or "dirty";
        };
      };

      devShells = {
        default = pkgs.mkShell {
          name = "development-shell";
          packages = [
            pkgs.go_1_22
            pkgs.nodejs_18
          ];
        };
      };

      overlayAttrs = self'.packages;
    };
  };
}
