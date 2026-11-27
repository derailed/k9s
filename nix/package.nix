{
  lib,
  buildGoModule,
  go,
  installShellFiles,
  stdenv,
  version,
  commit ? "nix",
  date ? "1970-01-01T00:00:00Z",
  src,
}:
(buildGoModule.override { inherit go; }) {
  pname = "k9s";
  inherit version src;

  vendorHash = "sha256-3FdxnRQOcpaz/wwQbX/AXR5h90YoUD0oVBDF5T4CvkM=";

  # Mirrors the `build` target of the Makefile.
  env.CGO_ENABLED = 0;
  tags = [ "netgo" ];
  ldflags = [
    "-s"
    "-w"
    "-X github.com/derailed/k9s/cmd.version=${version}"
    "-X github.com/derailed/k9s/cmd.commit=${commit}"
    "-X github.com/derailed/k9s/cmd.date=${date}"
  ];

  subPackages = [ "." ];

  nativeBuildInputs = [ installShellFiles ];

  # Tests are run from the dev shell (`make test`); several of them expect a
  # writable HOME / kubeconfig, which the sandbox does not provide.
  doCheck = false;

  postInstall = lib.optionalString (stdenv.buildPlatform.canExecute stdenv.hostPlatform) ''
    export HOME=$(mktemp -d)
    installShellCompletion --cmd k9s \
      --bash <($out/bin/k9s completion bash) \
      --fish <($out/bin/k9s completion fish) \
      --zsh <($out/bin/k9s completion zsh)
  '';

  doInstallCheck = true;
  installCheckPhase = ''
    runHook preInstallCheck
    export HOME=$(mktemp -d)
    $out/bin/k9s version --short | grep -F "${version}"
    runHook postInstallCheck
  '';

  meta = {
    description = "Kubernetes CLI To Manage Your Clusters In Style";
    homepage = "https://k9scli.io";
    changelog = "https://github.com/derailed/k9s/releases/tag/${version}";
    license = lib.licenses.asl20;
    mainProgram = "k9s";
    platforms = lib.platforms.unix;
  };
}
