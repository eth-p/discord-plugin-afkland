{ pkgs, ... }:

{
  packages = [
    pkgs.git
  ];

  languages.go = {
    enable = true;
  };

  languages.javascript = {
    enable = true;
    package = pkgs.nodejs;

    npm.enable = true;
    npm.install.enable = true;
  };
}
