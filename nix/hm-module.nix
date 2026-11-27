# Home Manager module for k9s.
#
# Mirrors `programs.k9s` from home-manager
# (https://github.com/nix-community/home-manager/blob/master/modules/programs/k9s.nix)
# and replaces it, so the option set stays identical while `package` defaults
# to the k9s built from this flake.
self:
{
  config,
  lib,
  pkgs,
  ...
}:
let
  inherit (lib)
    literalExpression
    mkIf
    mkOption
    types
    ;

  cfg = config.programs.k9s;
  yamlFormat = pkgs.formats.yaml { };
  inherit (pkgs.stdenv.hostPlatform) isDarwin system;

  enableXdgConfig = !isDarwin || config.xdg.enable;
  configDir = if enableXdgConfig then "k9s" else "Library/Application Support/k9s";

  yamlOption =
    {
      file,
      topic,
      example ? null,
    }:
    mkOption {
      inherit (yamlFormat) type;
      default = { };
      description = ''
        Written to {file}`$XDG_CONFIG_HOME/k9s/${file}` (linux)
        or {file}`Library/Application Support/k9s/${file}` (darwin).
        See <${topic}> for supported values.
      '';
    }
    // lib.optionalAttrs (example != null) { inherit example; };

  skinSetting =
    if !(cfg.settings ? k9s.ui.skin) && cfg.skins != { } then
      { k9s.ui.skin = builtins.head (builtins.attrNames cfg.skins); }
    else
      { };

  # file name -> content written under the k9s config dir
  files =
    lib.filterAttrs (_: v: v != null) {
      "config.yaml" =
        if cfg.settings != { } then
          yamlFormat.generate "k9s-config" (lib.recursiveUpdate skinSetting cfg.settings)
        else
          null;
      "aliases.yaml" =
        if cfg.aliases != { } then yamlFormat.generate "k9s-aliases" { inherit (cfg) aliases; } else null;
      "hotkeys.yaml" =
        if cfg.hotKeys != { } then yamlFormat.generate "k9s-hotkeys" { inherit (cfg) hotKeys; } else null;
      "plugins.yaml" =
        if cfg.plugins != { } then yamlFormat.generate "k9s-plugins" { inherit (cfg) plugins; } else null;
      "jumps.yaml" =
        if cfg.jumps != { } then yamlFormat.generate "k9s-jumps" { inherit (cfg) jumps; } else null;
      "views.yaml" =
        if cfg.views != { } then yamlFormat.generate "k9s-views" { inherit (cfg) views; } else null;
    }
    // lib.mapAttrs' (
      name: value:
      lib.nameValuePair "skins/${name}.yaml" (
        if types.path.check value then value else yamlFormat.generate "k9s-skin-${name}.yaml" value
      )
    ) cfg.skins;

  targetFiles = lib.mapAttrs' (
    name: source: lib.nameValuePair "${configDir}/${name}" { inherit source; }
  ) files;
in
{
  # Replace home-manager's own module to avoid duplicate option declarations.
  disabledModules = [ "programs/k9s.nix" ];

  options.programs.k9s = {
    enable = lib.mkEnableOption "k9s - Kubernetes CLI To Manage Your Clusters In Style";

    package = mkOption {
      type = types.nullOr types.package;
      default = self.packages.${system}.k9s or pkgs.k9s;
      defaultText = literalExpression "k9s.packages.\${system}.k9s";
      description = "The k9s package to install. Set to `null` to only manage configuration.";
    };

    settings = yamlOption {
      file = "config.yaml";
      topic = "https://k9scli.io/topics/config/";
      example = literalExpression ''
        {
          k9s = {
            refreshRate = 2;
          };
        }
      '';
    };

    skins = mkOption {
      type = with types; attrsOf (either yamlFormat.type path);
      default = { };
      description = ''
        Skin files written to {file}`$XDG_CONFIG_HOME/k9s/skins/` (linux)
        or {file}`Library/Application Support/k9s/skins/` (darwin). See
        <https://k9scli.io/topics/skins/> for supported values. When
        `settings.k9s.ui.skin` is unset, the first skin (alphabetically) is
        selected.
      '';
      example = literalExpression ''
        {
          my_blue_skin = {
            k9s = {
              body = {
                fgColor = "dodgerblue";
              };
            };
          };
          my_red_skin = ./red_skin.yaml;
        }
      '';
    };

    aliases = yamlOption {
      file = "aliases.yaml";
      topic = "https://k9scli.io/topics/aliases/";
      example = literalExpression ''
        {
          # Use pp as an alias for Pod
          pp = "v1/pods";
        }
      '';
    };

    hotKeys = yamlOption {
      file = "hotkeys.yaml";
      topic = "https://k9scli.io/topics/hotkeys/";
      example = {
        shift-0 = {
          shortCut = "Shift-0";
          description = "Viewing pods";
          command = "pods";
        };
      };
    };

    plugins = yamlOption {
      file = "plugins.yaml";
      topic = "https://k9scli.io/topics/plugins/";
      example = literalExpression ''
        {
          # Defines a plugin to provide a `ctrl-l` shortcut to
          # tail the logs while in pod view.
          fred = {
            shortCut = "Ctrl-L";
            description = "Pod logs";
            scopes = [ "po" ];
            command = "kubectl";
            background = false;
            args = [ "logs" "-f" "$NAME" "-n" "$NAMESPACE" "--context" "$CLUSTER" ];
          };
        }
      '';
    };

    jumps = yamlOption {
      file = "jumps.yaml";
      topic = "https://github.com/derailed/k9s#resource-jumps";
      example = {
        "argoproj.io/v1alpha1/applications" = {
          targetGVR = "apps/v1/deployments";
          labelSelector = "app.kubernetes.io/instance={{.metadata.name}}";
        };
      };
    };

    views = yamlOption {
      file = "views.yaml";
      topic = "https://k9scli.io/topics/columns/";
      example = {
        "v1/pods" = {
          columns = [
            "AGE"
            "NAMESPACE"
            "NAME"
            "IP"
            "NODE"
            "STATUS"
            "READY"
          ];
        };
      };
    };
  };

  config = mkIf cfg.enable {
    home.packages = mkIf (cfg.package != null) [ cfg.package ];
    xdg.configFile = mkIf enableXdgConfig targetFiles;
    home.file = mkIf (!enableXdgConfig) targetFiles;
  };
}
