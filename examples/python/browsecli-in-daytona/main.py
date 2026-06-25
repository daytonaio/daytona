"""Reach any website from a Daytona Sandbox with a Verified Browserbase browser.

Daytona is great at running your **agent loop** in a fast, secure sandbox — but a
sandbox can't browse the real web reliably. Its in-sandbox Chromium (shipped in
Daytona's desktop snapshot) is immature, browser support is on the roadmap, and a
sandbox still has a **datacenter IP** that Cloudflare / Akamai / DataDome block on
sight, with no anti-bot fingerprint hardening and no way to solve a CAPTCHA.

This example keeps the browser **out** of the sandbox. The Daytona Sandbox runs the
`browse` CLI (https://github.com/browserbase/stagehand/tree/main/packages/cli),
which connects out over CDP to a **Verified Browserbase browser** that:

  - uses a **residential / verified IP** — no datacenter-IP blocking
  - runs in **Verified browser mode** — passes bot-detection fingerprinting
  - **auto-solves CAPTCHAs / challenges** server-side

    ┌──────────────────────────┐      CDP over wss       ┌──────────────────────────┐
    │  Daytona Sandbox          │  ───────────────────────▶ │  Browserbase Verified    │
    │  node + `browse` CLI      │                            │  browser (residential IP,│
    │  your agent loop          │ ◀──────────────────────────│  stealth, CAPTCHA solve)  │
    └──────────────────────────┘      page data / refs     └──────────────────────────┘

This mirrors the OpenAI × Daytona computer-use cookbook pattern, but instead of
driving the sandbox's local desktop browser, it uses Browserbase as the hardened
anti-bot / CAPTCHA browser layer Daytona doesn't provide.

Daytona egress prerequisite
---------------------------
Daytona restricts sandbox outbound traffic with TWO layers: an IP firewall
(`networkAllowList` / `networkBlockAll`) AND an Envoy proxy that allowlists by
SNI/domain from a shared list (https://github.com/daytonaio/sandbox-network-whitelist).
By default only "Essential services" domains (github, npm, pypi, major AI
providers, ...) are reachable, so on a restricted tier (Tier 1/2) the TLS
handshake to api.browserbase.com is reset even if you pin its IPs — the IP-level
`networkAllowList` does NOT bypass the SNI filter. To run this template, EITHER:
  - use a Tier 3/4 Daytona org (full internet egress by default), OR
  - have the Browserbase domain-allowlist entry deployed to Daytona's proxies.
    Browserbase was added to the shared list in
    https://github.com/daytonaio/sandbox-network-whitelist/pull/117 (merged);
    once that whitelist is rolled out to the Envoy proxies, lower-tier orgs can
    reach `api.browserbase.com` / `connect.*.browserbase.com` too.

Run it:

    pip install -r requirements.txt
    export DAYTONA_API_KEY=dtn_...
    export BROWSERBASE_API_KEY=bb_live_...
    python main.py
"""

import os
import sys

from daytona import (
    CreateSandboxFromImageParams,
    Daytona,
    Image,
    Resources,
)

TARGET_URL = os.environ.get("TARGET_URL", "https://nowsecure.nl")

# The Browserbase credentials the in-sandbox `browse` CLI needs. We read them from
# the local environment and inject them into the sandbox at exec time so the key
# never has to be baked into the image or snapshot.
BROWSERBASE_ENV = {
    k: os.environ[k]
    for k in ("BROWSERBASE_API_KEY",)
    if os.environ.get(k)
}


def build_image() -> Image:
    """Declaratively build the sandbox image.

    Daytona's `debian_slim` base has no Node, so we add Node 20 via NodeSource,
    install the `browse` CLI globally, and copy in the demo script. **No
    Chrome/Chromium is installed** — the browser lives on Browserbase and is
    reached over CDP at run time.
    """
    return (
        Image.debian_slim("3.12")
        .run_commands(
            # debian_slim has neither curl nor npm; NodeSource needs curl, and its
            # `nodejs` package bundles npm (the distro `nodejs` package does not).
            "apt-get update",
            "apt-get install -y curl ca-certificates gnupg",
            "curl -fsSL https://deb.nodesource.com/setup_20.x | bash -",
            "apt-get install -y nodejs",
            "npm install -g browse@latest",
            "browse --version",
        )
        .add_local_file("browsecli-demo.sh", "/app/browsecli-demo.sh")
        .run_commands("chmod +x /app/browsecli-demo.sh")
        .workdir("/app")
    )


def main() -> int:
    if not BROWSERBASE_ENV.get("BROWSERBASE_API_KEY"):
        print(
            "[browsecli-in-daytona] BROWSERBASE_API_KEY is not set. "
            "Set it before running:\n"
            "  export BROWSERBASE_API_KEY=bb_live_..."
        )
        return 1

    # Reads DAYTONA_API_KEY (and optionally DAYTONA_API_URL / DAYTONA_TARGET) from
    # the environment. Pass DaytonaConfig(api_key=...) for explicit configuration.
    daytona = Daytona()

    image = build_image()

    # Create the sandbox directly from the declarative Image. Daytona builds the
    # image on the fly — no pre-made snapshot is required. (To reuse the image
    # across runs, call daytona.snapshot.create(CreateSnapshotParams(name=...,
    # image=image, resources=...), on_logs=print) once and then create from
    # CreateSandboxFromSnapshotParams(snapshot="...") instead.)
    #
    # NOTE: reaching Browserbase requires Daytona egress to permit
    # api.browserbase.com / connect.*.browserbase.com — see the "Daytona egress
    # prerequisite" in the module docstring (Tier 3/4, or the domain-allowlist PR).
    print("[browsecli-in-daytona] building image + creating sandbox...")
    sandbox = daytona.create(
        CreateSandboxFromImageParams(
            image=image,
            resources=Resources(cpu=1, memory=2, disk=4),
        ),
        on_snapshot_create_logs=lambda chunk: print(chunk, end=""),
    )

    try:
        # Run the same demo every sandbox template runs: create a Verified
        # Browserbase session (--proxies --verified --solve-captchas), open a
        # Cloudflare-protected page over CDP, and assert we reached real content
        # instead of a challenge wall. The Browserbase credentials are injected
        # per-exec so they stay out of the image/snapshot.
        print(f"[browsecli-in-daytona] running browsecli-demo.sh against {TARGET_URL}")
        response = sandbox.process.exec(
            "bash /app/browsecli-demo.sh",
            env={**BROWSERBASE_ENV, "TARGET_URL": TARGET_URL},
            timeout=180,
        )
        print(response.result)
        print(f"[browsecli-in-daytona] exit code: {response.exit_code}")
        return response.exit_code
    finally:
        print("[browsecli-in-daytona] deleting sandbox...")
        daytona.delete(sandbox)


if __name__ == "__main__":
    sys.exit(main())
