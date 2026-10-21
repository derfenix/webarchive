{
  pkgs,
  ...
}:

{

  packages = [
    pkgs.git
    pkgs.go
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
