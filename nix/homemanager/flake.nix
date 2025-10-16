{
  description = "Home Manager configuration example for HyprDynamicMonitors";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    hyprdynamicmonitors.url = "path:../..";
  };

  outputs = { self, nixpkgs, home-manager, hyprdynamicmonitors, ... }:
  let
    system = "x86_64-linux";
    pkgs = import nixpkgs { inherit system; };
  in {
    nixosConfigurations.hypr-vm = nixpkgs.lib.nixosSystem {
      inherit system;
      modules = [
        home-manager.nixosModules.home-manager
        {
          services.dbus.enable = true;
          services.pipewire.enable = true;
          hardware.graphics.enable = true;

          users.users.demo = {
            isNormalUser = true;
            extraGroups = [ "wheel" "video" "audio" "input" ];
            password = "demo";
          };

          programs.hyprland.enable = true;

          services.displayManager.autoLogin = {
            enable = true;
            user = "demo";
          };

          environment.systemPackages = with pkgs; [ git vim waybar foot wofi hyprpaper ];

          networking.hostName = "hypr-vm";

          system.stateVersion = "24.05";

          home-manager.users.demo = {
            imports = [ hyprdynamicmonitors.homeManagerModules.default ];

            home.username = "demo";
            home.homeDirectory = "/home/demo";
            home.stateVersion = "24.05";

            home.hyprdynamicmonitors = {
              enable = true;
              configFile = "${hyprdynamicmonitors}/examples/minimal/config.toml";
              extraFiles = {
                "hyprdynamicmonitors/hyprconfigs" = "${hyprdynamicmonitors}/examples/minimal/hyprconfigs";
              };
              extraFlags = [ "--debug" ];
              systemdTarget = "graphical-session.target";
            };

            wayland.windowManager.hyprland = {
              enable = true;
              systemd.enable = true;
            };

            home.packages = with pkgs; [
              waybar
              foot
              wofi
              hyprpaper
            ];

            programs.home-manager.enable = true;
          };
        }
      ];
    };
  };
}
