#!/usr/bin/env ruby
# frozen_string_literal: true

require 'daytona'

def main
  daytona = Daytona::Daytona.new

  owner = daytona.create
  puts "Owner sandbox ready: id=#{owner.id}"

  # Linked sandboxes must be ephemeral — `ephemeral: true` sets
  # `auto_delete_interval=0` automatically.
  follower = daytona.create(
    Daytona::CreateSandboxFromSnapshotParams.new(
      linked_sandbox: owner.id,
      ephemeral: true
    )
  )
  puts "Follower sandbox ready: id=#{follower.id}"

  begin
    # Background the http server with nohup, then poll locally until it
    # binds — so the follower's curl below doesn't race startup.
    puts "\nStarting `python3 -m http.server 3000` in owner #{owner.id}"
    start_script = <<~SH
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
    SH
    start_res = owner.process.exec(command: start_script, timeout: 30)
    raise "Failed to start server in owner: #{start_res.result}" if start_res.exit_code != 0

    puts start_res.result.strip

    # The link network registers each container's name as a DNS hostname.
    # The runner sets the container name to the sandbox id, so the follower
    # can reach the owner by id even when no explicit name is set on create.
    puts "\nReaching #{owner.id} from the follower over the link network"
    curl_res = follower.process.exec(
      command: "curl -sS --max-time 5 http://#{owner.id}:3000/",
      timeout: 10
    )
    if curl_res.exit_code != 0
      raise "Follower could not reach owner: exit=#{curl_res.exit_code} output=#{curl_res.result}"
    end

    puts "Response from owner: #{curl_res.result.strip}"
  ensure
    puts "\nDeleting follower #{follower.id}"
    daytona.delete(follower)
    puts "Deleting owner #{owner.id}"
    daytona.delete(owner)
  end
end

main if __FILE__ == $PROGRAM_NAME
