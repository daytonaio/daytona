from datetime import datetime
from daytona_api_client.models.snapshot_dto import SnapshotDto

def test_snapshot_to_json_serializes_datetime():
    snapshot = SnapshotDto(
        id="id",
        general=True,
        name="name",
        state="active",
        size=1,
        entrypoint=[],
        cpu=1,
        gpu=1,
        mem=1,
        disk=1,
        errorReason=None,
        createdAt=datetime.utcnow(),
        updatedAt=datetime.utcnow(),
        lastUsedAt=None,
        buildInfo=None,
    )

    json_str = snapshot.to_json()

    assert '"createdAt"' in json_str
    assert '"updatedAt"' in json_str
