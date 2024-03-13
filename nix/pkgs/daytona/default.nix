args@{ lib
, buildGo122Module
, src
, version
}:

buildGo122Module rec {
  pname = "daytona";
  version = "dev";

  inherit src;

  vendorHash = "sha256-6wjooiaYKkyaEttcYEtDIYV4m7mA35v87Q9DBdKq8n8=";

  ldflags = [
    "-s"
    "-w"
    "-X github.com/daytonaio/daytona/internal.Version=${version}-${args.version}"
  ];

  meta = {
    changelog = "https://github.com/daytonaio/daytona/releases/tag/v${version}";
    description = "The Open Source Dev Environment Manager";
    homepage = "https://github.com/daytonaio/daytona";
    license = lib.licenses.asl20;
    mainProgram = "daytona";
    maintainers = with lib.maintainers; [ ];
  };
}
