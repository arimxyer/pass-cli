class PassCli < Formula
  desc "Secure CLI password manager with AES-256-GCM encryption"
  homepage "https://github.com/ari1110/pass-cli"
  version "0.8.51"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/ari1110/pass-cli/releases/download/v0.0.1/pass-cli_0.0.1_darwin_amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_DARWIN_AMD64"
    end
    on_arm do
      url "https://github.com/ari1110/pass-cli/releases/download/v0.0.1/pass-cli_0.0.1_darwin_arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_DARWIN_ARM64"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/ari1110/pass-cli/releases/download/v0.0.1/pass-cli_0.0.1_linux_amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_LINUX_AMD64"
    end
    on_arm do
      url "https://github.com/ari1110/pass-cli/releases/download/v0.0.1/pass-cli_0.0.1_linux_arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_LINUX_ARM64"
    end
  end

  def install
    bin.install "pass-cli"

    # Generate shell completions
    generate_completions_from_executable(bin/"pass-cli", "completion")

    # Install documentation
    doc.install "README.md" if File.exist?("README.md")
    doc.install "LICENSE" if File.exist?("LICENSE")
  end

  def caveats
    <<~EOS
      Pass-CLI is a secure password manager that stores credentials locally.

      To get started:
        1. Initialize your vault: pass-cli init
        2. Add a credential: pass-cli add myservice
        3. Retrieve it: pass-cli get myservice

      Your vault is stored at: ~/.pass-cli/

      For more information, run: pass-cli --help
    EOS
  end

  test do
    # Test that the binary exists and is executable
    assert_match version.to_s, shell_output("#{bin}/pass-cli version")

    # Test help command
    assert_match "A secure CLI password manager", shell_output("#{bin}/pass-cli --help")

    # Test init command in a temporary directory
    testdir = testpath/"test-vault"
    mkdir_p testdir
    ENV["HOME"] = testdir
    system bin/"pass-cli", "init"
    assert_predicate testdir/".pass-cli", :exist?
  end
end
