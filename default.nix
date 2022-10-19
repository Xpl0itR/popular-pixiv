{ buildGoModule, fetchFromGitHub }:

buildGoModule rec {
  pname   = "popular-pixiv";
  version = "fbe50e3";
  meta    = {
    description = "A website which sorts pixiv search results by popularity";
    homepage    = "https://github.com/Xpl0itR/popular-pixiv";
  };

  src = fetchFromGitHub {
    owner  = "Xpl0itR";
    repo   = "popular-pixiv";
    rev    = "fbe50e3524a1383319bc89b3ddc3248b837fc1d3";
    sha256 = "sha256-GIIoXBnnoUL25zoTEpewPiWrczHb44Ttu7+DiLYOOaM=";
  };

  vendorSha256 = null;
  postInstall  = "cp -r html $out";
}