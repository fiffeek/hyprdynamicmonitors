{
  description = "Dynamic monitor configuration for Hyprland";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        version = if self ? rev then "git-${builtins.substring 0 7 (toString self.rev)}" else "dev";

        hyprdynamicmonitors = pkgs.buildGoModule {
          pname = "hyprdynamicmonitors";
          inherit version;

          src = ./.;
          vendorHash = "sha256-irbpDD7hn2SWDQNtRE5TzkqklXyx03ie8dSF+6SR50A=";

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
      in
      {
        packages.default = hyprdynamicmonitors;

        apps.default = flake-utils.lib.mkApp {
          drv = hyprdynamicmonitors;
        };
      });
}
