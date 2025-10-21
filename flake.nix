{
  description = "A output colorizer for your favorite commands";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    systems.url = "github:nix-systems/default-linux";
  };
  outputs = {
    self,
    systems,
    nixpkgs,
  }: let
    forAllSystems = nixpkgs.lib.genAttrs (import systems);
    pkgsFor = nixpkgs.legacyPackages;
  in {
    packages = forAllSystems (system: {
      default = pkgsFor.${system}.callPackage ./. {};
    });
    # TODO: create the nix shell
    # devShells = forAllSystems (system: {
    #   default = pkgsFor.${system}.callPackage ./shell.nix {};
    # });
    # TODO: create the home moduel
    # homeModules = {
    #   rong = import ./module.nix self;
    #   default = self.homeModules.rong;
    # };
  };
}
