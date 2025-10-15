## Problem

When a Daytona sandbox reaches its storage limit , it becomes unusable with the error:

```
Error response from daemon: mkdir /var/lib/docker/overlay2/...: no space left on device
```

This prevents the container from starting, with no clear recovery path for users to preserve their data and work.

## Expected Behavior

Daytona should:

1. Automatically allocate small increments of additional storage (10kB) when a container hits its limit
2. Allow this automatic expansion up to 10% beyond the original sandbox size
3. Display clear warnings that this extra space is temporary and limited

## Rationale

- Prevents immediate workflow disruption when storage limits are reached
- Gives users time to clean up unnecessary files or properly migrate data
- Reduces support requests for sandbox recovery
- Improves user experience with large development projects

These are the commands i used directly on the runner host to get that desired effect basiclly:

# 1. Get the container's image ID

IMAGE_ID=$(docker inspect dc22c691-a475-43cb-ace7-9a4715411e42 -f '{{.Image}}')

# 2. Rename the original container

docker rename dc22c691-a475-43cb-ace7-9a4715411e42 dc22c691-old

# 3. Create new container with identical settings + 100MB extra storage

docker create \
  --name dc22c691-a475-43cb-ace7-9a4715411e42 \
  --privileged \
  --runtime=sysbox-runc \
  --env "DAYTONA_SANDBOX_ID=dc22c691-a475-43cb-ace7-9a4715411e42" \
  --env "DAYTONA_SANDBOX_SNAPSHOT=cr.app.daytona.io/sbox/4d8eb3d1-2cee-475f-89f7-c430bfd5d209:24.04" \
  --env "DAYTONA_SANDBOX_USER=root" \
  --memory=8589934592 \
  --memory-swap=8589934592 \
  --storage-opt size=10.1G \
  --cpu-period=100000 \
  --cpu-quota=400000 \
  --security-opt label=disable \
  --shm-size=67108864 \
  --cgroupns=private \
  --ipc=private \
  --add-host=host.docker.internal:host-gateway \
  --label "daytona.organization_id=759d646f-5427-4aee-b4e3-3290fa920fce" \
  --label "daytona.organization_name=Magic AI, Inc." \
  --label "org.opencontainers.image.ref.name=ubuntu" \
  --label "org.opencontainers.image.version=24.04" \
  --mount type=bind,source=/opt/daytona-runner/.tmp/binaries/daemon-amd64,destination=/usr/local/bin/daytona,readonly \
  --mount type=bind,source=/opt/daytona-runner/.tmp/binaries/daytona-computer-use,destination=/usr/local/lib/daytona-computer-use,readonly \
  --entrypoint "sleep" \
  $IMAGE_ID infinity

# 4. Start the new container

docker start dc22c691-a475-43cb-ace7-9a4715411e42

# 5. Copy data from old container's upper layer to new container

SOURCE="/var/lib/docker/overlay2/a40d64f4fd706649b959ab440869a5cdec41d85680c69d658305c3459cf0d721/diff"
tar -C $SOURCE -cf - . | docker exec -i dc22c691-a475-43cb-ace7-9a4715411e42 tar -C / -xf -

Besides that recovery on the runner (needs to be implemented agnosticly on the runner as for example microVM runners will implement this differently than this docker runner we have currently).

We have to have a way to notify the user that the sandbox is facing issues (dont implement now but give me suggestions on how we could do this the best, so the user knows this needs to be addressed by them)
