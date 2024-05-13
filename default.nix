{ buildGoModule, fetchFromGitHub }:

buildGoModule rec {
  pname   = "popular-pixiv";
  version = "08d9467";
  meta    = {
    description = "A website which sorts pixiv search results by popularity";
    homepage    = "https://github.com/Xpl0itR/popular-pixiv";
  };

  src = fetchFromGitHub {
    owner  = "Xpl0itR";
    repo   = "popular-pixiv";
    rev    = "08d9467ade39b62f8475a9c08ea1bcc637f3b076";
    sha256 = "1a8mgxji00dk47avmvkx3qa2gq9fpj82816ff8r0z614p36k2pip";
  };

  vendorHash = null;
  postInstall = "cp -r html $out";
}
