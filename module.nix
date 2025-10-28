{
  config,
  lib,
  pkgs,
  ...
}: let
  inherit (lib) getExe mkEnableOption mkIf mkOption mkOrder types;
  pkg = pkgs.callPackage ./. {};
  cfg = config.programs.chromashift;
in {
  options.programs.chromashift = {
    enable = mkEnableOption "Enable chromashift";

    package = mkOption {
      type = types.package;
      default = pkg;
      description = "ChromaShift package to use";
    };

    enableBashIntegration = lib.hm.shell.mkBashIntegrationOption {inherit config;};
    enableFishIntegration = lib.hm.shell.mkFishIntegrationOption {inherit config;};
    enableZshIntegration = lib.hm.shell.mkZshIntegrationOption {inherit config;};
  };

  config = mkIf cfg.enable {
    home.packages = mkIf (cfg.package != null) [cfg.package];
    programs.bash.initExtra = mkIf cfg.enableBashIntegration (
      mkOrder 2000 ''
        eval "$(${getExe cfg.package} alias bash)"
      ''
    );

    programs.zsh.initContent = mkIf cfg.enableZshIntegration (
      mkOrder 2000 ''
        eval "$(${getExe cfg.package} alias zsh)"
      ''
    );

    programs.fish.interactiveShellInit = mkIf cfg.enableFishIntegration (
      mkOrder 2000 ''
        ${getExe cfg.package} alias fish | source
      ''
    );
  };
}
