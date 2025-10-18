{
  description = "Dynamic monitor configuration for Hyprland";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let
      # NixOS module (supports mode = "user"|"system"|"none")
      nixosModule = { config, pkgs, lib, ... }:
        let
          cfg = config.services.hyprdynamicmonitors;
          defaultPkg = self.packages.${pkgs.system}.default;
          etcKey = lib.replaceStrings [ "/etc/" ] [ "" ] cfg.configPath;
        in
        with lib; {
          options.services.hyprdynamicmonitors = {
            enable = mkEnableOption "HyprDynamicMonitors service";

            package = mkOption {
              type = types.package;
              default = defaultPkg;
              defaultText = literalExpression "inputs.hyprdynamicmonitors.packages.${pkgs.system}.default";
              description = "Package providing the hyprdynamicmonitors binary.";
            };

            mode = mkOption {
              type = types.enum [ "user" "system" "none" ];
              default = "user";
              description = "Whether to install a user unit, system unit, or no unit at all.";
            };

            configPath = mkOption {
              type = types.str;
              default = "/etc/xdg/hyprdynamicmonitors/config.toml";
              description = "Absolute path the binary will read the TOML config from.";
            };

            config = mkOption {
              type = types.nullOr types.str;
              default = null;
              description = "Inline TOML config to be written to configPath. Takes precedence over configFile.";
            };

            configFile = mkOption {
              type = types.nullOr types.path;
              default = null;
              description = "Path to a TOML file whose contents will be copied to configPath.";
            };

            extraFiles = mkOption {
              type = types.nullOr (types.attrsOf types.path);
              default = null;
              description = "Attrset mapping relative /etc paths (e.g. \"hyprdynamicmonitors/asset\") to source paths.";
            };

            installExamples = mkOption {
              type = types.bool;
              default = true;
              description = "If true and no config/configFile is provided, install a minimal example config.";
            };

            serviceOptions = mkOption {
              type = types.attrs;
              default = {};
              description = "Extra fields merged into serviceConfig (passed through to systemd).";
            };

            systemdTarget = mkOption {
              type = types.str;
              default = "graphical-session.target";
              description = "Systemd target to bind to.";
            };

            extraFlags = mkOption {
              type = types.listOf types.str;
              default = [];
              example = [ "--debug" ];
              description = "Extra command-line flags to pass to hyprdynamicmonitors run.";
            };
          };

          config = mkIf cfg.enable (let
            tomlAttr = if cfg.config != null then
              { "${etcKey}" = { text = cfg.config; }; }
            else if cfg.configFile != null then
              { "${etcKey}" = { source = cfg.configFile; }; }
            else if cfg.installExamples then
              { "${etcKey}" = {
                  text = ''
                    # Example HyprDynamicMonitors config
                  '';
                }; }
            else
              { };

            extraEtc = if cfg.extraFiles != null then
              builtins.listToAttrs (map (k: { name = k; value = { source = cfg.extraFiles.${k}; }; }) (builtins.attrNames cfg.extraFiles))
            else {};

            etcAttrs = tomlAttr // extraEtc;

            unitSpec = {
              description = "HyprDynamicMonitors - Dynamic monitor configuration for Hyprland";
              documentation = [ "https://fiffeek.github.io/hyprdynamicmonitors/" ];
              partOf = [ cfg.systemdTarget ];
              requires = [ cfg.systemdTarget ];
              after = [ cfg.systemdTarget ];

              serviceConfig = {
                Type = "simple";
                ExecStart = "${cfg.package}/bin/hyprdynamicmonitors run --config ${cfg.configPath} ${lib.escapeShellArgs cfg.extraFlags}";
                Slice = "session.slice";
                Restart = "on-failure";
                RestartSec = 5;
              } // cfg.serviceOptions;

              wantedBy = [ cfg.systemdTarget ];
            };
          in mkMerge [
            { environment.etc = etcAttrs; }

            (mkIf (cfg.mode == "system") {
              systemd.services.hyprdynamicmonitors = unitSpec;
            })
            (mkIf (cfg.mode == "user") {
              systemd.user.services.hyprdynamicmonitors = unitSpec;
            })
          ]);
        };

      # Home Manager module (installs under ~/.config and registers user systemd unit)
      homeModule = { config, pkgs, lib, ... }:
        let
          hmCfg = config.home.hyprdynamicmonitors;
          defaultPkg = self.packages.${pkgs.system}.default;
          defaultConfigPath = "${config.home.homeDirectory}/.config/hyprdynamicmonitors/config.toml";
        in
        with lib; {
          options.home.hyprdynamicmonitors = {
            enable = mkEnableOption "HyprDynamicMonitors home config";

            package = mkOption {
              type = types.package;
              default = defaultPkg;
              defaultText = literalExpression "inputs.hyprdynamicmonitors.packages.${pkgs.system}.default";
            };

            configPath = mkOption {
              type = types.str;
              default = defaultConfigPath;
              description = "Path under the user's home where the TOML config will be written.";
            };

            config = mkOption {
              type = types.nullOr types.str;
              default = null;
            };

            configFile = mkOption {
              type = types.nullOr types.path;
              default = null;
            };

            extraFiles = mkOption {
              type = types.nullOr (types.attrsOf types.path);
              default = null;
              description = "Attrset mapping paths under ~/.config (relative) to source paths.";
            };

            installExamples = mkOption {
              type = types.bool;
              default = true;
            };

            serviceOptions = mkOption {
              type = types.attrs;
              default = {};
            };

            systemdTarget = mkOption {
              type = types.str;
              default = if (config.wayland.systemd.target or null) != null then config.wayland.systemd.target else "graphical-session.target";
              defaultText = literalExpression "config.wayland.systemd.target";
              description = "Systemd target to bind to.";
            };

            extraFlags = mkOption {
              type = types.listOf types.str;
              default = [];
              example = [ "--debug" ];
              description = "Extra command-line flags to pass to hyprdynamicmonitors run.";
            };
          };

          config = mkIf hmCfg.enable (let
            tomlEntry = if hmCfg.config != null then
              { "${hmCfg.configPath}" = { text = hmCfg.config; }; }
            else if hmCfg.configFile != null then
              { "${hmCfg.configPath}" = { source = hmCfg.configFile; }; }
            else if hmCfg.installExamples then
              { "${hmCfg.configPath}" = { text = ''
                  # Example HyprDynamicMonitors config
                ''; }; }
            else {};

            extraFilesHM = if hmCfg.extraFiles != null then
              builtins.listToAttrs (map (k: { name = "${config.home.homeDirectory}/.config/${k}"; value = { source = hmCfg.extraFiles.${k}; }; }) (builtins.attrNames hmCfg.extraFiles))
            else {};
          in {
            home.file = tomlEntry // extraFilesHM;

            systemd.user.services.hyprdynamicmonitors = {
              Unit = {
                Description = "HyprDynamicMonitors - Dynamic monitor configuration for Hyprland";
                After = hmCfg.systemdTarget;
                PartOf = hmCfg.systemdTarget;
                Requires = hmCfg.systemdTarget;
              };
              Service = {
                ExecStart = "${hmCfg.package}/bin/hyprdynamicmonitors run --config ${hmCfg.configPath} ${lib.escapeShellArgs hmCfg.extraFlags}";
                Restart = "on-failure";
                RestartSec = 5;
              } // hmCfg.serviceOptions;
              Install = {
                WantedBy = [ hmCfg.systemdTarget ];
              };
            };
          });
        };

    in
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        version =
          if self ? rev
          then "git-${builtins.substring 0 7 (toString self.rev)}"
          else "dev";

        hyprdynamicmonitors = pkgs.buildGoModule {
          pname = "hyprdynamicmonitors";
          inherit version;

          src = ./.;
          vendorHash = "sha256-xM65zdq9Rk0ZggJBeu5tguYhYiDPMCpce3ra6tzOlwA=";

          subPackages = [ "." ];

          ldflags = [
            "-s" "-w"
            "-X github.com/fiffeek/hyprdynamicmonitors/cmd.Version=${version}"
            "-X github.com/fiffeek/hyprdynamicmonitors/cmd.Commit=${self.rev or "unknown"}"
            "-X github.com/fiffeek/hyprdynamicmonitors/cmd.BuildDate=${self.lastModifiedDate or "1970-01-01T00:00:00Z"}"
          ];

          nativeBuildInputs = [ pkgs.installShellFiles ];

          postInstall = ''
            installShellCompletion --cmd hyprdynamicmonitors \
              --bash <($out/bin/hyprdynamicmonitors completion bash) \
              --fish <($out/bin/hyprdynamicmonitors completion fish) \
              --zsh <($out/bin/hyprdynamicmonitors completion zsh)
          '';

          meta = with pkgs.lib; {
            description = "Dynamic monitor configuration for Hyprland";
            homepage = "https://github.com/fiffeek/hyprdynamicmonitors";
            license = licenses.mit;
            platforms = platforms.linux;
            mainProgram = "hyprdynamicmonitors";
          };
        };
      in {
        packages.default = hyprdynamicmonitors;

        apps.default = flake-utils.lib.mkApp {
          drv = hyprdynamicmonitors;
        };
      })
      // {
        nixosModules = {
          hyprdynamicmonitors = nixosModule;
          default = nixosModule;
        };

        homeManagerModules = {
          hyprdynamicmonitors = homeModule;
          default = homeModule;
        };
      };
}
