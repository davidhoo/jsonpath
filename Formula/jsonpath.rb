class Jsonpath < Formula
  desc "RFC 9535 compliant JSONPath processor with beautiful colored output"
  homepage "https://github.com/davidhoo/jsonpath"
  version "1.0.2"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/davidhoo/jsonpath/releases/download/v1.0.2/jp-darwin-arm64.tar.gz"
      sha256 "53bae8aeb8a7c32f983ba9ddb794eb4ad57d3c35d980e617c36f31fb3eff9e6b"
    else
      url "https://github.com/davidhoo/jsonpath/releases/download/v1.0.2/jp-darwin-amd64.tar.gz"
      sha256 "6f2830389413ab1497d6bd74df37577527854adca5e23ec0ad22f1da64ac9377"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/davidhoo/jsonpath/releases/download/v1.0.2/jp-linux-arm64.tar.gz"
      sha256 "f69ac99d468737a5bd09d067e1a1c30ffe0c60adf706dbf75d9c3262a2833b50"
    else
      url "https://github.com/davidhoo/jsonpath/releases/download/v1.0.2/jp-linux-amd64.tar.gz"
      sha256 "baddd0334c7a8589a4f5997d31aa71b5943f2e0eb71e80f2aa05d1c512bf060e"
    end
  end

  def install
    bin.install "jp"
  end

  test do
    assert_equal "\"jp\"", shell_output("#{bin}/jp -p '$.name' <<< '{\"name\":\"jp\"}'").strip
  end
end 