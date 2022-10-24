{ buildGoModule, fetchFromGitHub }:

buildGoModule rec {
  pname   = "popular-pixiv";
  version = "2967360";
  meta    = {
    description = "A website which sorts pixiv search results by popularity";
    homepage    = "https://github.com/Xpl0itR/popular-pixiv";
  };

  src = fetchFromGitHub {
    owner  = "Xpl0itR";
    repo   = "popular-pixiv";
    rev    = "2967360fcfa433b5346f4a5bf31b187805b79326";
    sha256 = "sha256-f2Uw6tSRyzyob+/QhjepeHVHTHlHb3vMRS4ej+OG/To=";
  };

  vendorSha256 = null;
  postInstall  = "cp -r html $out";
}