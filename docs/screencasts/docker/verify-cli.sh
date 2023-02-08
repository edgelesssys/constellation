#!/usr/bin/expect -f
# Note: Expects to be able to run 'sudo install' without a password

set timeout -1
set send_human {0.05 0 1 0.05 0.3}
set CTRLC \003
set record_name [lindex $argv 0];

proc expect_prompt {} {
    # make sure this matches your prompt
    expect "$ "
}

proc run_command {cmd} {
    send -h "$cmd"
    send "\r"
    expect -timeout 1
}

proc send_keystroke_to_interactive_process {key {addl_sleep 2}} {
    send "$key"
    expect -timeout 1
    sleep $addl_sleep
}

# Start recording
spawn asciinema rec $record_name
send "\r"
expect_prompt

### Step 0: Requirements
run_command "echo Step 0: Installing requirements"
expect_prompt
run_command "go install github.com/sigstore/cosign/cmd/cosign@latest"
expect_prompt
run_command "go install github.com/sigstore/rekor/cmd/rekor-cli@latest"
expect_prompt

### Step 1: Download CLI
run_command "echo Step 1: Download CLI and signature"
expect_prompt
run_command "curl -sLO https://github.com/edgelesssys/constellation/releases/latest/download/constellation-linux-amd64"
expect_prompt
run_command "curl -sLO https://github.com/edgelesssys/constellation/releases/latest/download/constellation-linux-amd64.sig"
expect_prompt

### Step 2: Verify the CLI using cosign
run_command "echo Step 2: Verify the CLI using cosign and the public Rekor transparency log"
expect_prompt
run_command "COSIGN_EXPERIMENTAL=1 cosign verify-blob --key https://edgeless.systems/es.pub --signature constellation-linux-amd64.sig constellation-linux-amd64"
expect_prompt

### Step 2b: Verify the CLI manually
run_command "echo Optional Step 2b: Manually inspect the Rekor transparency log"
expect_prompt
run_command "rekor-cli search --artifact constellation-linux-amd64"
expect -re "\n(\[a-f0-9]+)\r"
set uuid '$expect_out(1,string)'
expect_prompt
run_command "rekor-cli get --uuid=$uuid"
expect_prompt

### Step 3: Install the CLI
run_command "echo Step 4: Install the CLI"
expect_prompt
run_command "sudo install constellation-linux-amd64 /usr/local/bin/constellation"
expect_prompt
run_command "echo Done! You can now use the verified CLI"
expect_prompt
run_command "constellation -h"

# Stop recording
send "exit"
