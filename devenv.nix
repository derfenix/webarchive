{
  pkgs,
  nixpkgs_stable,
  ...
}:

{

  packages = [
    pkgs.git
    pkgs.go_1_23
  ];

  enterShell = ''
    git --version
    go version
  '';

  enterTest = ''
    echo "Running tests"
    go test -race ./...
  '';
}
