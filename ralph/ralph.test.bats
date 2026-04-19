#!/usr/bin/env bats
#
# Integration tests for ralph/ralph.
#
# Run with:   bats ralph/ralph.test.bats
# Install:    sudo apt install bats   # or: brew install bats-core
#
# Strategy:
#   Each test spins up a fresh temp dir containing a git repo, a copy of
#   ralph, empty prompt files, and a fake `claude` binary on PATH. The
#   fake `claude` is the only knob we twist per test: by controlling what
#   it prints and how long it runs, we can drive ralph through every
#   exit path (timeout, no-signal, max-iterations, clean completion).

setup() {
    TEST_DIR="$(mktemp -d)"
    export PATH="$TEST_DIR/bin:$PATH"
    mkdir -p "$TEST_DIR/bin"

    cd "$TEST_DIR"
    git init -q .
    git -c user.email=t@t -c user.name=t commit --allow-empty -q -m init

    mkdir ralph
    cp "$BATS_TEST_DIRNAME/ralph" ralph/ralph
    chmod +x ralph/ralph
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
    WATCHDOG_POLL_INTERVAL=1 ITERATION_TIMEOUT=1 run ralph/ralph 1
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
    WATCHDOG_POLL_INTERVAL=1 ITERATION_TIMEOUT=3 run ralph/ralph 1
    [ "$status" -eq 0 ]
    [[ "$output" == *"signaled completion"* ]]
    [[ "$output" != *"timed out"* ]]
}

@test "reports generic failure when claude exits without completion signal" {
    stub_claude 'echo "no signal emitted"; exit 0'
    run ralph/ralph 1
    [ "$status" -ne 0 ]
    [[ "$output" == *"finished without completing"* ]]
    [[ "$output" != *"timed out"* ]]
}

@test "exits 0 when claude emits <promise>COMPLETE</promise>" {
    stub_claude 'echo "<promise>COMPLETE</promise>"'
    run ralph/ralph 1
    [ "$status" -eq 0 ]
    [[ "$output" == *"signaled completion"* ]]
}

@test "reports max iterations when only <promise>NEXT</promise> is emitted" {
    stub_claude 'echo "<promise>NEXT</promise>"'
    run ralph/ralph 1
    [ "$status" -ne 0 ]
    [[ "$output" == *"reached max iterations"* ]]
}

@test "security mode requires --ref-branch" {
    stub_claude 'echo "<promise>COMPLETE</promise>"'
    run ralph/ralph security 1
    [ "$status" -ne 0 ]
    [[ "$output" == *"Usage: ./ralph security"* ]]
    [[ "$output" == *"--ref-branch"* ]]
}

# ─── Unit tests for run_with_stall_watchdog ───
# Sources ralph in guarded mode (RALPH_SOURCE_ONLY=1) so the function is
# callable, then relaxes the strict-mode options ralph enables.
load_wrapper() {
    export RALPH_SOURCE_ONLY=1
    # shellcheck disable=SC1091
    source "$TEST_DIR/ralph/ralph"
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

@test "rotate_log: trims oversized file to last max bytes, keeps newest content" {
    load_wrapper
    seq 1 5000 > "$TEST_DIR/big.log"
    initial=$(stat -c %s "$TEST_DIR/big.log")
    rotate_log "$TEST_DIR/big.log" 100
    final=$(stat -c %s "$TEST_DIR/big.log")
    [ "$initial" -gt 100 ]
    [ "$final" -le 100 ]
    tail -1 "$TEST_DIR/big.log" | grep -q "^5000$"
    # Under-cap and missing files are no-ops.
    echo "tiny" > "$TEST_DIR/small.log"
    rotate_log "$TEST_DIR/small.log" 1000000
    [ "$(cat "$TEST_DIR/small.log")" = "tiny" ]
    rotate_log "$TEST_DIR/missing.log" 100
    [ ! -f "$TEST_DIR/missing.log" ]
}

# ─── Per-mode behavior ───

@test "plan mode defaults to 1 iteration when max_iterations not passed" {
    stub_claude 'echo "<promise>NEXT</promise>"'
    run ralph/ralph plan --goal "feature"
    [ "$status" -ne 0 ]
    [[ "$output" == *"reached max iterations (1)"* ]]
}

@test "security mode defaults to 1 iteration when max_iterations not passed" {
    stub_claude 'echo "<promise>NEXT</promise>"'
    run ralph/ralph security --ref-branch main
    [ "$status" -ne 0 ]
    [[ "$output" == *"reached max iterations (1)"* ]]
}

@test "plan mode respects explicit max_iterations" {
    stub_claude 'echo "<promise>NEXT</promise>"'
    run ralph/ralph plan 3 --goal "feature"
    [ "$status" -ne 0 ]
    [[ "$output" == *"reached max iterations (3)"* ]]
}

@test "security mode respects explicit max_iterations" {
    stub_claude 'echo "<promise>NEXT</promise>"'
    run ralph/ralph security 3 --ref-branch main
    [ "$status" -ne 0 ]
    [[ "$output" == *"reached max iterations (3)"* ]]
}

@test "plan mode requires --goal" {
    stub_claude 'echo "<promise>COMPLETE</promise>"'
    run ralph/ralph plan 1
    [ "$status" -ne 0 ]
    [[ "$output" == *"Usage: ./ralph plan"* ]]
    [[ "$output" == *"--goal"* ]]
}

@test "plan mode completes on <promise>COMPLETE</promise>" {
    stub_claude 'echo "<promise>COMPLETE</promise>"'
    run ralph/ralph plan 1 --goal "feature"
    [ "$status" -eq 0 ]
    [[ "$output" == *"signaled completion"* ]]
}

@test "security mode completes on <promise>COMPLETE</promise>" {
    stub_claude 'echo "<promise>COMPLETE</promise>"'
    run ralph/ralph security 1 --ref-branch main
    [ "$status" -eq 0 ]
    [[ "$output" == *"signaled completion"* ]]
}

@test "build mode header advertises unlimited iterations when no count passed" {
    stub_claude 'echo "<promise>COMPLETE</promise>"'
    run ralph/ralph
    [ "$status" -eq 0 ]
    [[ "$output" == *"Max iterations: unlimited"* ]]
}

@test "build mode loops past iteration 1 by default (unlimited)" {
    # Stub emits NEXT for the first two iterations, then COMPLETE.
    # If build defaulted to 1 iter, the run would stop after iter 1 with
    # "reached max iterations" instead of signaling completion at iter 3.
    stub_claude '
counter="'"$TEST_DIR"'/counter"
n=$(cat "$counter" 2>/dev/null || echo 0)
n=$((n+1))
echo "$n" > "$counter"
if [ "$n" -lt 3 ]; then
    echo "<promise>NEXT</promise>"
else
    echo "<promise>COMPLETE</promise>"
fi'
    run ralph/ralph
    [ "$status" -eq 0 ]
    [[ "$output" == *"signaled completion at iteration 3"* ]]
}

@test "plan mode substitutes [project-specific goal] in prompt" {
    # Give the prompt a placeholder; have the stub echo stdin back so we can
    # confirm the substitution reached claude.
    echo "GOAL_WAS=[project-specific goal]" > "$TEST_DIR/ralph/prompt-plan.md"
    stub_claude '
cat > "'"$TEST_DIR"'/claude-stdin"
echo "<promise>COMPLETE</promise>"'
    run ralph/ralph plan 1 --goal "MY_SPECIAL_GOAL"
    [ "$status" -eq 0 ]
    grep -q "GOAL_WAS=MY_SPECIAL_GOAL" "$TEST_DIR/claude-stdin"
}

@test "security mode substitutes [ref-branch] in prompt" {
    echo "REF=[ref-branch]" > "$TEST_DIR/ralph/prompt-security.md"
    stub_claude '
cat > "'"$TEST_DIR"'/claude-stdin"
echo "<promise>COMPLETE</promise>"'
    run ralph/ralph security 1 --ref-branch MY_SPECIAL_BRANCH
    [ "$status" -eq 0 ]
    grep -q "REF=MY_SPECIAL_BRANCH" "$TEST_DIR/claude-stdin"
}
