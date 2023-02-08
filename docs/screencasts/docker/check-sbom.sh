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
spawn asciinema rec --overwrite $record_name
send "\r"
expect_prompt

### Step 0: Requirements
run_command "echo Step 0: Installing requirements"
expect_prompt
run_command "curl -sLO https://github.com/anchore/grype/releases/download/v0.56.0/grype_0.56.0_linux_amd64.tar.gz"
expect_prompt
run_command "tar -xvzf grype_0.56.0_linux_amd64.tar.gz"
expect_prompt
run_command "sudo install grype /usr/local/bin/grype"
expect_prompt
run_command "grype --help"
expect_prompt

### Step 1: Download & check SBOM
run_command "echo Step 1: Download Constellation SBOM"
expect_prompt
run_command "curl -sLO https://github.com/edgelesssys/constellation/releases/latest/download/constellation.spdx.sbom"
expect_prompt
run_command "grype constellation.spdx.sbom -o table -q"
expect_prompt
run_command "echo We are safe! :)"

# Stop recording
send "exit"
