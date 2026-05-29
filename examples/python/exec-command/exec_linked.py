from daytona import CreateSandboxFromSnapshotParams, Daytona


def main():
    daytona = Daytona()

    owner = None
    follower = None
    try:
        owner = daytona.create()
        print(f"Owner sandbox ready: id={owner.id} name={owner.name}")

        # Linked sandboxes must be ephemeral — `ephemeral=True` sets
        # `auto_delete_interval=0` automatically.
        follower = daytona.create(
            CreateSandboxFromSnapshotParams(
                linked_sandbox=owner.id,
                ephemeral=True,
            ),
        )
        print(f"Follower sandbox ready: id={follower.id} name={follower.name}")
        print(f"  follower.linked_sandbox_id = {follower.linked_sandbox_id}")

        # Background the http server with nohup, then poll locally until it
        # binds — so the follower's curl below doesn't race startup.
        print(f"\nStarting `python3 -m http.server 3000` in owner {owner.name!r}")
        start_script = """
set -e
mkdir -p /tmp/lnk
echo 'hello from owner' > /tmp/lnk/index.html
cd /tmp/lnk
nohup python3 -m http.server 3000 > /tmp/lnk/srv.log 2>&1 &
for _ in $(seq 1 20); do
  if curl -sS --max-time 1 http://127.0.0.1:3000/ >/dev/null 2>&1; then
    echo READY
    exit 0
  fi
  sleep 0.5
done
echo "server failed to start"
cat /tmp/lnk/srv.log
exit 1
"""
        start_res = owner.process.exec(start_script, timeout=30)
        if start_res.exit_code != 0:
            raise RuntimeError(f"Failed to start server in owner: {start_res.result}")
        print(start_res.result.strip())

        # The link network registers the owner under its sandbox name as a
        # DNS alias, so the follower can reach it by name.
        print(f"\nReaching {owner.name!r} from the follower over the link network")
        curl_res = follower.process.exec(
            f"curl -sS --max-time 5 http://{owner.name}:3000/",
            timeout=10,
        )
        if curl_res.exit_code != 0:
            raise RuntimeError(f"Follower could not reach owner: exit={curl_res.exit_code} output={curl_res.result}")
        print(f"Response from owner: {curl_res.result.strip()}")
    finally:
        if follower is not None:
            print(f"\nDeleting follower {follower.id}")
            daytona.delete(follower)
        if owner is not None:
            print(f"Deleting owner {owner.id}")
            daytona.delete(owner)


if __name__ == "__main__":
    main()
