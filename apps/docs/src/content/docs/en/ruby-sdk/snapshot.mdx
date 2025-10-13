---
title: "SnapshotService"
hideTitleOnPage: true
---

## SnapshotService

SnapshotService class for Daytona SDK.

### Constructors

#### new SnapshotService()

```ruby
def initialize(snapshots_api:, object_storage_api:, default_region_id:, otel_state:)

```

**Parameters**:

- `snapshots_api` _DaytonaApiClient:SnapshotsApi_ - The snapshots API client
- `object_storage_api` _DaytonaApiClient:ObjectStorageApi_ - The object storage API client
- `default_region_id` _String, nil_ - Default region ID for snapshot creation
- `otel_state` _Daytona:OtelState, nil_ -

**Returns**:

- `SnapshotService` - a new instance of SnapshotService

### Methods

#### list()

```ruby
def list(page:, limit:)

```

List all Snapshots.

**Parameters**:

- `page` _Integer, Nil_ -
- `limit` _Integer, Nil_ -

**Returns**:

- `Daytona:PaginatedResource` - Paginated list of all Snapshots

**Raises**:

- `Daytona:Sdk:Error` -

**Examples:**

```ruby
daytona = Daytona::Daytona.new
response = daytona.snapshot.list(page: 1, limit: 10)
snapshots.items.each { |snapshot| puts "#{snapshot.name} (#{snapshot.image_name})" }

```

#### delete()

```ruby
def delete(snapshot)

```

Delete a Snapshot.

**Parameters**:

- `snapshot` _Daytona:Snapshot_ - Snapshot to delete

**Returns**:

- `void`

**Examples:**

```ruby
daytona = Daytona::Daytona.new
snapshot = daytona.snapshot.get("demo")
daytona.snapshot.delete(snapshot)
puts "Snapshot deleted"

```

#### get()

```ruby
def get(name)

```

Get a Snapshot by name.

**Parameters**:

- `name` _String_ - Name of the Snapshot to get

**Returns**:

- `Daytona:Snapshot` - The Snapshot object

**Examples:**

```ruby
daytona = Daytona::Daytona.new
snapshot = daytona.snapshot.get("demo")
puts "#{snapshot.name} (#{snapshot.image_name})"

```

#### create()

```ruby
def create(params, on_logs:)

```

Creates and registers a new snapshot from the given Image definition.

**Parameters**:

- `params` _Daytona:CreateSnapshotParams_ - Parameters for snapshot creation
- `on_logs` _Proc, Nil_ - Callback proc handling snapshot creation logs

**Returns**:

- `Daytona:Snapshot` - The created snapshot

**Examples:**

```ruby
image = Image.debianSlim('3.12').pipInstall('numpy')
params = CreateSnapshotParams.new(name: 'my-snapshot', image: image)
snapshot = daytona.snapshot.create(params) do |chunk|
  print chunk
end

```

#### activate()

```ruby
def activate(snapshot)

```

Activate a snapshot

**Parameters**:

- `snapshot` _Daytona:Snapshot_ - The snapshot instance

**Returns**:

- `Daytona:Snapshot`
