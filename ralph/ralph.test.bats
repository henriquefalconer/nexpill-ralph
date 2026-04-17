#!/usr/bin/env bats
#
# Integration tests for ralph/ralph.sh.
#
# Run with:   bats ralph/ralph.test.bats
# Install:    sudo apt install bats   # or: brew install bats-core
#
# Strategy:
#   Each test spins up a fresh temp dir containing a git repo, a copy of
#   ralph.sh, empty prompt files, and a fake `claude` binary on PATH. The
#   fake `claude` is the only knob we twist per test: by controlling what
#   it prints and how long it runs, we can drive ralph.sh through every
#   exit path (timeout, no-signal, max-iterations, clean completion).

setup() {
    TEST_DIR="$(mktemp -d)"
    export PATH="$TEST_DIR/bin:$PATH"
    mkdir -p "$TEST_DIR/bin"

    cd "$TEST_DIR"
    git init -q .
    git -c user.email=t@t -c user.name=t commit --allow-empty -q -m init

    mkdir ralph
    cp "$BATS_TEST_DIRNAME/ralph.sh" ralph/ralph.sh
    chmod +x ralph/ralph.sh
    : > ralph/prompt-build.md
    : > ralph/prompt-plan.md
    : > ralph/prompt-security.md
}

teardown() {
    rm -rf "$TEST_DIR"
}

# Write a fake `claude` whose body is the given shell snippet.
stub_claude() {
    cat > "$TEST_DIR/bin/claude" <<EOF
#!/usr/bin/env bash
$1
EOF
    chmod +x "$TEST_DIR/bin/claude"
}

@test "reports timeout (not generic failure) when ralph/progress.txt stalls past ITERATION_TIMEOUT" {
    stub_claude 'sleep 30'
    WATCHDOG_POLL_INTERVAL=1 ITERATION_TIMEOUT=1 run ralph/ralph.sh 1
    [ "$status" -ne 0 ]
    [[ "$output" == *"timed out"* ]]
    [[ "$output" == *"idle for 1s"* ]]
    [[ "$output" != *"finished without completing"* ]]
}

@test "does not time out while claude keeps appending to ralph/progress.txt" {
    stub_claude 'for i in 1 2 3 4 5; do
    echo "tick $i" >> "'"$TEST_DIR"'/ralph/progress.txt"
    sleep 1
done
echo "<promise>COMPLETE</promise>"'
    WATCHDOG_POLL_INTERVAL=1 ITERATION_TIMEOUT=3 run ralph/ralph.sh 1
    [ "$status" -eq 0 ]
    [[ "$output" == *"signaled completion"* ]]
    [[ "$output" != *"timed out"* ]]
}

@test "reports generic failure when claude exits without completion signal" {
    stub_claude 'echo "no signal emitted"; exit 0'
    run ralph/ralph.sh 1
    [ "$status" -ne 0 ]
    [[ "$output" == *"finished without completing"* ]]
    [[ "$output" != *"timed out"* ]]
}

@test "exits 0 when claude emits <promise>COMPLETE</promise>" {
    stub_claude 'echo "<promise>COMPLETE</promise>"'
    run ralph/ralph.sh 1
    [ "$status" -eq 0 ]
    [[ "$output" == *"signaled completion"* ]]
}

@test "reports max iterations when only <promise>NEXT</promise> is emitted" {
    stub_claude 'echo "<promise>NEXT</promise>"'
    run ralph/ralph.sh 1
    [ "$status" -ne 0 ]
    [[ "$output" == *"reached max iterations"* ]]
}

@test "security mode requires --ref-branch" {
    stub_claude 'echo "<promise>COMPLETE</promise>"'
    run ralph/ralph.sh security 1
    [ "$status" -ne 0 ]
    [[ "$output" == *"Usage: ./ralph.sh security"* ]]
    [[ "$output" == *"--ref-branch"* ]]
}

# ─── Unit tests for run_with_stall_watchdog ───
# Sources ralph.sh in guarded mode (RALPH_SOURCE_ONLY=1) so the function is
# callable, then relaxes the strict-mode options ralph.sh enables.
load_wrapper() {
    export RALPH_SOURCE_ONLY=1
    # shellcheck disable=SC1091
    source "$TEST_DIR/ralph/ralph.sh"
    set +e
    set +u
    set +o pipefail
    : > "$PROGRESS_FILE"
}

@test "run_with_stall_watchdog: prints captured output on stdout for \$(...) callers" {
    load_wrapper
    captured=$(run_with_stall_watchdog bash -c 'echo STDOUT_CAPTURE_PROOF' 2>/dev/null)
    [[ "$captured" == *"STDOUT_CAPTURE_PROOF"* ]]
}

@test "run_with_stall_watchdog: tees live output to stderr for user visibility" {
    load_wrapper
    run_with_stall_watchdog bash -c 'echo STDERR_TEE_PROOF' > "$TEST_DIR/out" 2> "$TEST_DIR/err"
    grep -q STDERR_TEE_PROOF "$TEST_DIR/err"
    grep -q STDERR_TEE_PROOF "$TEST_DIR/out"
}

@test "run_with_stall_watchdog: returns the command's own exit code on non-stall completion" {
    load_wrapper
    rc=0
    run_with_stall_watchdog bash -c 'exit 7' >/dev/null 2>&1 || rc=$?
    [ "$rc" -eq 7 ]
}
