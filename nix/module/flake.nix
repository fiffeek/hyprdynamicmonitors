{
  description = "Ephemeral NixOS VM with Hyprland Wayland session";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    hyprdynamicmonitors.url = "path:../..";
  };

  outputs = { self, nixpkgs, hyprdynamicmonitors, ... }:
  let
    system = "x86_64-linux";
    pkgs = import nixpkgs { inherit system; };
  in {
    nixosConfigurations.hypr-vm = nixpkgs.lib.nixosSystem {
      inherit system;
      modules = [
        hyprdynamicmonitors.nixosModules.default
        {
          services.dbus.enable = true;
          services.pipewire.enable = true;
          hardware.opengl.enable = true;

          users.users.demo = {
            isNormalUser = true;
            extraGroups = [ "wheel" "video" "audio" "input" ];
            password = "demo";
          };

          programs.hyprland.enable = true;

          services.hyprdynamicmonitors = {
            enable = true;
            mode = "user";
            configFile = "${hyprdynamicmonitors}/examples/full/config.toml";
            extraFiles = {
              "xdg/hyprdynamicmonitors/hyprconfigs" = "${hyprdynamicmonitors}/examples/full/hyprconfigs";
            };
            extraFlags = [ "--debug" ];
            systemdTarget = "graphical-session.target";
          };

          services.xserver.enable = true;
          services.xserver.displayManager.startx.enable = true;
          services.xserver.displayManager.autoLogin = {
            enable = true;
            user = "demo";
          };

          environment.systemPackages = with pkgs; [ git vim waybar foot wofi hyprpaper ];

          networking.hostName = "hypr-vm";
        }
      ];
    };
  };
}
