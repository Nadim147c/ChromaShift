{
  lib,
  buildGoModule,
  fetchFromGitHub,
}:
buildGoModule rec {
  pname = "chromashift";
  version = "2.0.1";

  src = fetchFromGitHub {
    owner = "Nadim147c";
    repo = "ChromaShift";
    rev = "v${version}";
    hash = "sha256-5Wk8hhDrqYLTHOIQWqjS6z1DUflI1jZ4YHkYDr/t1t4=";
  };

  vendorHash = "sha256-OjW2NMFk6EmS/iL2kZwSl+AYZ1bg1ylFNWYrnuHU6tY=";

  ldflags = ["-s" "-w" "-X" "main.Version=v${version}"];

  meta = {
    description = "A output colorizer for your favorite commands";
    homepage = "https://github.com/Nadim147c/ChromaShift";
    license = lib.licenses.gpl3Only;
    mainProgram = "cshift";
  };
}
