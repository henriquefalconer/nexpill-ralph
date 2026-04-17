# Packages Plan

Implementation plan for `specs/nexai-architecture.md`. One bullet tree per package, organized into Claude Code sessions. Each session is self-contained and can be run independently once its prerequisites are merged. Citations are `specs/nexai-architecture.md:<line>` for the spec and `src/...:<line>` for current code.

> Note: TypeScript 6.0 requires Node 22+ due to use of `using` declarations. Downgraded to ^5.8.0 for Node 20 compatibility.

## Completed

### Phase 0: Workspace foundation (Session 00)

- Created `pnpm-workspace.yaml` with `packages: ["packages/*"]`.
- Rewrote root `package.json` as a virtual manifest (private, no runtime deps, migration-window devDeps included).
- Added `.npmrc` with strict peer deps config.
- Created `tsconfig.base.json` with shared compiler options.
- Created `tsconfig.references.json` listing `nexai-common-protocol`.
- Kept original `tsconfig.json` for legacy `src/` compilation during the migration window.
- Added root scripts (`build`, `test`, `test:dispatcher`, `test:provider-basaglia`).
- Added `scripts/require-local-env.mjs`.

### Phase 1, Session 01: `nexai-common-protocol`

- `packages/nexai-common-protocol/package.json`
- `packages/nexai-common-protocol/tsconfig.json`
- `packages/nexai-common-protocol/src/routes.ts` (ROUTES constants)
- `packages/nexai-common-protocol/src/anthropic.ts` (AnthropicMessagesRequest/Response zod schemas)
- `packages/nexai-common-protocol/src/voyage.ts` (VoyageEmbedRequest/Response)
- `packages/nexai-common-protocol/src/slack.ts` (SlackWebRequest/Response)
- `packages/nexai-common-protocol/src/db.ts` (DbQueryRequest/DbExecRequest with scope)
- `packages/nexai-common-protocol/src/mixpanel.ts` (MixpanelTrackRequest with dispatcher-stamped field enforcement)
- `packages/nexai-common-protocol/src/auth.ts` (SvidClaims, BootstrapRequest/Response, RefreshRequest/Response)
- `packages/nexai-common-protocol/src/heartbeat.ts` (HeartbeatRequest/Response)
- `packages/nexai-common-protocol/src/manifest.ts` (NexaiCapabilities, ProviderManifest)
- `packages/nexai-common-protocol/src/index.ts` (re-exports all)
- Builds successfully with `tsc --build`.

### Session 02: `nexai-common-http` - DONE

### Session 03: `nexai-common-logging` - DONE

### Session 04: `nexai-common-errors` - DONE

### Session 05: nexai-dispatcher-db - DONE

- `packages/nexai-dispatcher-db/src/connection.ts`: openDb(path), openAllDbs(dataDir, providerNames) → DbRegistry with shared + lazy per-provider handles
- `packages/nexai-dispatcher-db/src/pragmas.ts`: WAL, NORMAL sync, busy_timeout=5000
- `packages/nexai-dispatcher-db/src/schema.ts`: createTables lifted verbatim
- `packages/nexai-dispatcher-db/src/vector.ts`: searchByVector takes explicit db handle
- `packages/nexai-dispatcher-db/src/scoping.ts`: assertScopeAllowed/resolveDb enforce shared_read/shared_write from manifest with table-level SQL extraction
- `packages/nexai-dispatcher-db/src/seeds/*`: jobs.ts + knowledge.ts lifted; knowledge.ts accepts injectable EmbedFn for Voyage (wired in Session 14)

### Session 06: nexai-dispatcher-auth - DONE

- `packages/nexai-dispatcher-auth/src/backend.ts`: SvidBackend interface (sign/verify) for test injection
- `packages/nexai-dispatcher-auth/src/issuer.ts`: issueBootstrap (validates secret + issues 15min HS256 JWT), issueRefresh
- `packages/nexai-dispatcher-auth/src/verifier.ts`: verify(bearer, opts) → SvidClaims; extractBearer helper

### Session 07: nexai-dispatcher-railway - DONE

- `packages/nexai-dispatcher-railway/src/client.ts`: RailwayClient interface + createRailwayClient() using fetchJson; all Railway GraphQL mutations
- `packages/nexai-dispatcher-railway/src/dry-run.ts`: DryRunOptions + isDryRun() gates destructive mutations
- `packages/nexai-dispatcher-railway/src/diff.ts`: MANAGED_PREFIX="nexai-provider-", diffServices() classifies missing/drifted/orphan/kept
- `packages/nexai-dispatcher-railway/src/ops.ts`: serviceCreate, serviceReconverge, serviceRollForward, serviceDelete, deploymentRestart (composed higher-level ops)
- `packages/nexai-dispatcher-railway/src/reconciler.ts`: reconcile() reads package.json via fs.readFileSync, diffs, applies changes
- 14 tests passing (diff + reconciler suites)

### Session 08: `nexai-dispatcher-proxy-anthropic` - DONE

- `packages/nexai-dispatcher-proxy-anthropic/src/handler.ts`: validates body, verifies SVID, gates model + budget, forwards to Anthropic, streams back or returns JSON.
- `packages/nexai-dispatcher-proxy-anthropic/src/stream.ts`: `passthroughStream` wrapper for SSE.
- `packages/nexai-dispatcher-proxy-anthropic/src/budget.ts`: per-provider monthly usage table in shared SQLite, `checkBudget`/`recordUsage`.
- `packages/nexai-dispatcher-proxy-anthropic/src/context.ts`: structural `AnthropicProxyContext` interface (subset of DispatcherContext).
- 13 tests passing.

### Session 09: `nexai-dispatcher-proxy-voyage` - DONE

- `packages/nexai-dispatcher-proxy-voyage/src/handler.ts`: validates with `VoyageEmbedRequest` from `nexai-common-protocol`, gates on `voyage.models` + `monthly_budget_usd`, forwards, logs.

### Session 10: `nexai-dispatcher-proxy-slack` - DONE

- `packages/nexai-dispatcher-proxy-slack/src/handler.ts`: `/v1/slack/web/*`; looks up caller's `xoxb-` keyed by SVID claim `provider_name`, forwards.
- `packages/nexai-dispatcher-proxy-slack/src/token-store.ts`: reads per-provider `xoxb-` from Railway variables set by the reconciler.
- `packages/nexai-dispatcher-proxy-slack/src/scopes.ts`: gates method against `slack.scopes` in manifest.

### Session 11: `nexai-dispatcher-proxy-mixpanel` - DONE

- `packages/nexai-dispatcher-proxy-mixpanel/src/handler.ts`: validates with `MixpanelTrackRequest`, strips client-supplied enrichment fields, stamps server-side values; reads core provider `package.json:version` for `app_version`.
- `packages/nexai-dispatcher-proxy-mixpanel/src/client.ts`: thin wrapper around the `mixpanel` SDK holding the single project token.

### Session 12: `nexai-dispatcher-health` - DONE

- `packages/nexai-dispatcher-health/src/aggregator.ts`: parallel checks (SQLite `SELECT 1` 500ms timeout, Railway `viewer { id }`, Mixpanel token echo, Anthropic/Voyage `/models` ping). Returns 200/503/207 per `specs/nexai-architecture.md:314`.
- `packages/nexai-dispatcher-health/src/heartbeat.ts`: `/v1/heartbeat` handler; in-memory `last_seen_at` map keyed by `provider_name` (spec 316). Staleness detector flags `now - last_seen_at > 2min` for reconciler to restart.

### Session 13: `nexai-dispatcher-analytics` - DONE

- `packages/nexai-dispatcher-analytics/src/registry.ts`: reads every `packages/nexai-provider-<name>/package.json`, exposes `getManifest(provider_name)` returning `NexaiCapabilities | null`
- `packages/nexai-dispatcher-analytics/src/events.ts`: emits 9 event types (proxy latency, reconciler outcomes, Railway API calls, capability denials, budget blocks, SVID issuance, DB proxy latency, provider boot-to-ready, heartbeat staleness); reads dispatcher version from `packages/nexai-dispatcher/package.json` at module load
- `packages/nexai-dispatcher-analytics/src/index.ts`: re-exports all public API
- 18 tests passing (registry + events suites)
- Added `packages/nexai-dispatcher-analytics` to `tsconfig.references.json`

### Session 14: `nexai-dispatcher` - DONE

- `packages/nexai-dispatcher/src/env.ts`: zod-validated env snapshot (RAILWAY_PROJECT_TOKEN, ANTHROPIC_API_KEY, VOYAGE_API_KEY, MIXPANEL_TOKEN, SVID_SIGNING_SECRET, PORT, DATA_DIR, PACKAGES_DIR, etc.)
- `packages/nexai-dispatcher/src/context.ts`: DispatcherContext with db, railway, svid, mixpanel, registry, heartbeat, tokenStore, logger, env, getProviderVersion
- `packages/nexai-dispatcher/src/server.ts`: Hono app with all routes (auth bootstrap/refresh, anthropic/voyage/slack/mixpanel proxies, db query/exec, heartbeat, health). deriveBootstrapSecret() for deterministic provider secrets.
- `packages/nexai-dispatcher/src/index.ts`: boot - opens DBs, seeds data, builds context, reconciles Railway, starts stale heartbeat loop, listens on PORT
- Added `@hono/node-server` and `hono` as dependencies
- All 260 tests still pass

### Session 15: `nexai-provider-sdk` - DONE

- `packages/nexai-provider-sdk/src/bootstrap.ts`: `createSvidClient()` exchanges PROVIDER_BOOTSTRAP_SECRET at ROUTES.AUTH_BOOTSTRAP, schedules JWT refresh 2min before expiry
- `packages/nexai-provider-sdk/src/anthropic.ts`: `createAnthropicClient()` with overloaded `messages()` - returns AnthropicMessagesResponse or ReadableStream for streaming
- `packages/nexai-provider-sdk/src/voyage.ts`: `createVoyageClient()` with `embed()` against ROUTES.VOYAGE_EMBED
- `packages/nexai-provider-sdk/src/slack-web.ts`: `createSlackWebClient()` with `call(method, args)` against ROUTES.SLACK_WEB/{method}
- `packages/nexai-provider-sdk/src/db.ts`: `createDbClient()` with `query(scope, sql, params)` and `exec(scope, sql, params)`
- `packages/nexai-provider-sdk/src/mixpanel.ts`: `createMixpanelClient()` with `track(event, properties)` against ROUTES.MIXPANEL_TRACK
- `packages/nexai-provider-sdk/src/heartbeat.ts`: `createHeartbeatClient()` with 30s interval loop
- `packages/nexai-provider-sdk/src/request-tracker.ts`: `createRequestTracker(mixpanel)` factory - lifted sanitizeHeaders, extractBodySnippet, trackRequestDetails; calls sdk.mixpanel.track
- `packages/nexai-provider-sdk/src/process-scope.ts`: `createProcessScope(mixpanel)` factory - lifted registerProcess, unregisterProcess, isProcessActive; calls sdk.mixpanel.track
- `packages/nexai-provider-sdk/src/sdk.ts`: `createSdk(opts)` bootstraps SVID and wires all proxy clients into NexaiSdk
- `packages/nexai-provider-sdk/src/index.ts`: re-exports all
- Added to tsconfig.references.json
- 260 tests still passing

### Session 16: `nexai-provider-messaging` - DONE

- `packages/nexai-provider-messaging/src/types.ts`: lifted from `src/lib/messaging/types.ts` unchanged
- `packages/nexai-provider-messaging/src/mrkdwn.ts`: lifted `toSlackMrkdwn` from `src/lib/messaging/mrkdwn.ts`
- `packages/nexai-provider-messaging/src/slack.ts`: lifted `initSlackApp`/`getSlackApp` from `src/lib/messaging/slack.ts`
- `packages/nexai-provider-messaging/src/handlers.ts`: lifted `registerMessageHandler` (now takes `slackWeb: SlackWebClient` as 2nd arg) and `enqueueForKey` from `src/lib/messaging/handlers.ts`
- `packages/nexai-provider-messaging/src/lifecycle.ts`: lifted `respondWithPlaceholder`, `runResponder`, `deliverFinal`; outbound Slack calls (`chat.postMessage`, `chat.update`, `chat.delete`) route through `slackWeb.call(...)` per spec 290
- Added to `tsconfig.references.json`
- Tests migrated to `packages/nexai-provider-messaging/src/__tests__/`, updated to use new 4-arg `registerMessageHandler` signature with `slackWeb` mock
- Deleted `src/lib/messaging/`
- 260 tests still passing

### Session 17: `nexai-provider-sessions` - DONE

- `packages/nexai-provider-sessions/src/types.ts`: lifted from `src/lib/sessions/types.ts` unchanged
- `packages/nexai-provider-sessions/src/loader.ts`: lifted from `src/lib/sessions/loader.ts`, rewritten to use async `DbClient` (from `@nexai/provider-sdk`) with `db.query("private", ...)` / `db.exec("private", ...)` instead of `getDb()`. All functions are now async.
- `packages/nexai-provider-sessions/src/index.ts`: re-exports all
- Added to `tsconfig.references.json`
- Note: `src/lib/sessions/` NOT deleted yet - Basaglia agent (`src/@agents/basaglia/`) still imports from it. Will be deleted in Session 19+ when Basaglia is migrated.

### Session 18: `nexai-provider-orchestration` - DONE

- `packages/nexai-provider-orchestration/src/types.ts`: `OrchestratorInput` shape (`text`, `threadTs`, `userId`, `traceId`)
- `packages/nexai-provider-orchestration/src/state-machine.ts`: `abstract class DeterministicStateMachine<TState, TContext>` with `transition(ctx, event)`, `currentState`, `canTransition`, `reset`, `onEnter/onExit` hooks, optional logger
- `packages/nexai-provider-orchestration/src/index.ts`: re-exports all
- Added to `tsconfig.references.json`
- 12 tests passing

## Phase 4: Basaglia provider - DONE

### Session 19: `nexai-provider-basaglia` (core + orchestrator + prompts) - DONE

- **Spec**: `specs/nexai-architecture.md:39, 97-101, 175-180, 184-203`.
- **Purpose**: worker entry + orchestrator + shared prompts + session types.
- **Create**
  - `packages/nexai-provider-basaglia/package.json`: declares the `nexai.capabilities` block per `specs/nexai-architecture.md:188-203` (anthropic models, voyage models, slack app_id/scopes, storage shared_read on `jobs`+`knowledge`, mixpanel super_properties).
  - `packages/nexai-provider-basaglia/src/context.ts`: `ProviderContext` as `specs/nexai-architecture.md:98-101` (`sdk`, `slackApp`, `sessions`, `logger`, `env`).
  - `packages/nexai-provider-basaglia/src/index.ts`: boot. Reads `DISPATCHER_URL`, `PROVIDER_NAME`, `PROVIDER_BOOTSTRAP_SECRET`, `SLACK_BOT_TOKEN`, `SLACK_APP_TOKEN` per `specs/nexai-architecture.md:177-178`. Calls `sdk.bootstrap()`, builds `ProviderContext`, registers the Bolt handler via `@nexai/provider-messaging`, starts heartbeat.
  - `packages/nexai-provider-basaglia/src/agent.ts`: lift `AGENT_NAME`, `initBasagliaAgent`, `buildTraceId`, `readSessionContext`, `handleMessage` from `src/@agents/basaglia/index.ts:11-73`. The `initBasagliaAgent(anthropicClient)` signature changes to `initBasagliaAgent(ctx: ProviderContext)`; it uses `ctx.sdk.anthropic` instead of a direct `Anthropic` client (spec 248).
  - `packages/nexai-provider-basaglia/src/orchestrator.ts`: lift `src/@agents/basaglia/orchestrator.ts:1-2791`. Every `anthropicClient.messages.create(...)` becomes `ctx.sdk.anthropic.messages(...)`. Analytics calls route through `ctx.sdk.mixpanel.track` (the current `src/@agents/basaglia/orchestrator.ts` import of `../../lib/analytics/track.js` becomes an import from `@nexai/provider-sdk`). The capability branches stay: `handlePostEvaluation` (`orchestrator.ts:988`), `handleIdle` (`:1395`), `handleCandidateEvaluation` (`:1712`), `handleInterviewQuestions` (`:2044`), `handleGenericTranscriptAnalysis` (`:2189`), `handleScreeningEvaluation` (`:2319`), `handleHrSupport` (`:2552`). They now import from sub-capability packages (sessions 20-22).
  - `packages/nexai-provider-basaglia/src/session-state.ts`, `session-context.ts`: lift `src/@agents/basaglia/session-state.ts` and `src/@agents/basaglia/session-context.ts` unchanged.
  - `packages/nexai-provider-basaglia/src/prompts/*.ts`: lift `src/@agents/basaglia/prompts/{system.ts,candidate-evaluation.ts,screening-evaluation.ts,hr-support.ts}`. Sub-capability packages re-import these.
- **Dependencies**: `@nexai/provider-sdk`, `@nexai/provider-messaging`, `@nexai/provider-sessions`, `@nexai/provider-orchestration`, `@nexai/common-*`, `@nexai/provider-basaglia-candidate-evaluation`, `@nexai/provider-basaglia-screening-evaluation`, `@nexai/provider-basaglia-hr-support`. Never `@nexai/dispatcher-*` per `specs/nexai-architecture.md:52`.

### Session 20: `nexai-provider-basaglia-candidate-evaluation` - DONE

- **Spec**: `specs/nexai-architecture.md:40`.
- **Purpose**: candidate-evaluation capability, one-package-per-capability per loom spirit (spec 38).
- **Create**
  - `packages/nexai-provider-basaglia-candidate-evaluation/src/fetch-transcription.ts`: lift `src/@agents/basaglia/candidate-evaluation/fetch-transcription.ts:54-177`. Analytics import switches from `@lib/analytics/*` to `@nexai/provider-sdk`.
  - `packages/nexai-provider-basaglia-candidate-evaluation/src/interview-evaluation.ts`: lift `src/@agents/basaglia/candidate-evaluation/interview-evaluation.ts:27-132`. Rewrite `new Anthropic({...}).messages.create(...)` to `ctx.sdk.anthropic.messages(...)`. Model `claude-sonnet-4-20250514` stays (must be present in the manifest's `anthropic.models` block).
  - `packages/nexai-provider-basaglia-candidate-evaluation/src/get-job-description.ts`: lift `src/@agents/basaglia/candidate-evaluation/get-job-description.ts:10-38`. Rewrite `getDb()` usage to `ctx.sdk.db.query("shared", ...)` per `specs/nexai-architecture.md:198, 265` (basaglia's manifest grants `shared_read: ["jobs", "knowledge"]`).
  - `packages/nexai-provider-basaglia-candidate-evaluation/src/list-job-titles.ts`: same treatment as `get-job-description.ts`, lifting `src/@agents/basaglia/candidate-evaluation/list-job-titles.ts:15-51`.
  - Tests lifted as-is, but their fixtures stop constructing a real SQLite; they stub `sdk.db` instead.
- **Dependencies**: `@nexai/provider-sdk`, `@nexai/common-*`. Never sibling provider packages (`@nexai/provider-basaglia-screening-evaluation` etc.) per `specs/nexai-architecture.md:52`.

### Session 21: `nexai-provider-basaglia-screening-evaluation` - DONE

- **Spec**: `specs/nexai-architecture.md:41`.
- **Create**
  - `packages/nexai-provider-basaglia-screening-evaluation/src/screening-evaluation.ts`: lift `src/@agents/basaglia/screening-evaluation/screening-evaluation.ts:22-118`. Anthropic + analytics rewrites as in Session 20.
  - `packages/nexai-provider-basaglia-screening-evaluation/src/fetch-screening-transcription.ts`: lift `src/@agents/basaglia/screening-evaluation/fetch-screening-transcription.ts:3-9`. Today it delegates to the candidate-evaluation package's `fetchTranscription`. Dependency rule at `specs/nexai-architecture.md:52` forbids sibling dependency; extract the shared fetcher into **`nexai-provider-basaglia`** core (session 19) and both sub-packages import from there. Alternatively move the transcription fetcher to `@nexai/provider-sdk` if `MCP_PLAN.md`'s hoisting plan lands first (agent 5 flagged the MCP spec also hoists this code).
  - Prompts already live in `@nexai/provider-basaglia/prompts`; import from core.
- **Dependencies**: `@nexai/provider-sdk`, `@nexai/provider-basaglia` (for the transcription fetcher + prompts), `@nexai/common-*`.

### Session 22: `nexai-provider-basaglia-hr-support` - DONE

- **Spec**: `specs/nexai-architecture.md:42`.
- **Create**
  - `packages/nexai-provider-basaglia-hr-support/src/search-knowledge-base.ts`: lift `src/@agents/basaglia/hr-support/search-knowledge-base.ts:79-120`. The current direct Voyage fetch (and `process.env.VOYAGE_API_KEY` at the top of the file) disappears; replace with `ctx.sdk.voyage.embed({...})`. The vector search hits `ctx.sdk.db.query("shared", ...)` against the `knowledge` table (shared read granted by the basaglia manifest per `specs/nexai-architecture.md:198`).
- **Dependencies**: `@nexai/provider-sdk`, `@nexai/provider-basaglia` (prompts), `@nexai/common-*`.
- **End-of-session cleanup**: delete `src/@agents/` and `src/index.ts`. The monolith is gone.

## Phase 5: Runtime isolation

### Session 23: Dockerfiles + reconciler wiring - DONE

- Created `docker/dispatcher.Dockerfile` verbatim from spec lines 210-222.
- Created `docker/provider.Dockerfile` verbatim from spec lines 226-241.
- Created `.nexai-never-match/.gitkeep` sentinel path for watchPatterns.
- Reconciler wiring (dockerfilePath, buildArgs, watchPatterns) was already implemented in packages/nexai-dispatcher-railway/src/ops.ts from Session 07.

## Phase 6: Test harness

### Session 24: Vitest workspace + system suite - DONE

- Created `vitest.workspace.ts` at root, discovers `packages/*/vitest.config.ts` and `tests/system/vitest.config.ts`.
- Created `vitest.config.ts` in all 21 packages that have test scripts.
- Created `tests/system/` as a pnpm workspace package (`@nexai/system-tests`) with package.json declaring all `@nexai/*` workspace dependencies.
- Created `tests/system/dispatcher.test.ts`: 11 system integration tests covering auth bootstrap, DB scoping/enforcement, Mixpanel enrichment, heartbeat recording, and SVID refresh. Uses Hono's in-memory `app.request()` transport.
- Added `tests/system` to `pnpm-workspace.yaml`.
- All package tsconfigs already had proper `exclude` patterns for test files.
- 312 tests pass total (up from 260).

## Execution order

- Sessions run in the numbered order. Each session is independently reviewable and (modulo the monorepo retrofit in Session 00) revertible.
- A session that lifts code deletes the original path **only at the end** of that session. Until then both the old and new paths build, because Session 00 kept the legacy `src/` compiling.
- Every cross-package import is `workspace:*` per `specs/nexai-architecture.md:72`. Semver bumps follow the existing repo rule (`CLAUDE.md` "App version (semver)"): a change to any lifted package bumps that package plus every dependent package in the same commit.

Note: All sessions 00-24 are now complete. The full monorepo migration is done.

## IMPORTANT

- For each package, author property based tests or unit tests (which ever is best)

## Completed post-migration work

### Iteration 2 (2026-04-17)
- Added `nexai-dispatcher-db/src/__tests__/connection.test.ts`: 21 tests covering openDb, applyPragmas, createTables, openAllDbs, and searchByVector per `specs/nexai-test-database-connection.md`
- Added `nexai-provider-sessions/src/__tests__/loader.test.ts`: 16 unit tests for all 5 loader functions per `specs/nexai-test-sessions-loader.md`
- Added `nexai-provider-sessions/src/__tests__/lifecycle.integration.test.ts`: 10 integration tests covering the full 4-phase lifecycle and 5 invariants per `specs/nexai-test-sessions-lifecycle.md`
- Total tests: 359 (up from 312)

### Iteration 3 (2026-04-17)
- Added `nexai-provider-sdk/src/__tests__/request-tracker.test.ts`: 9 tests covering sanitizeHeaders (redaction, key preservation, no mutation), extractBodySnippet (default window, JSON serialize, custom window, MAX_BODY_LENGTH cap, undefined input), and createRequestTracker (event naming, header sanitization, omit when absent, body_snippet present/absent) per `specs/nexai-test-analytics-request-tracker.md`
- Added `nexai-provider-sdk/src/__tests__/process-scope.test.ts`: 7 tests covering createProcessScope factory (register/active, process.start event, unregister/inactive, process.end event, unknown ID, concurrent processes, instance isolation) per `specs/nexai-test-analytics-process-scope.md`
- Added `nexai-provider-sdk/src/__tests__/analytics.test.ts`: 8 tests covering trackToolExecution (tool.name event, extras merge, no-op with default extras, namespace) and trackError (error.source event, namespace, extras merge, non-Error values) per `specs/nexai-test-analytics-track.md`
- Total tests: 388 (up from 359)

### Iteration 4 (2026-04-17)
- Added `nexai-provider-sdk/src/__tests__/analytics-integration.test.ts`: 14 integration tests covering the full analytics flow per `specs/nexai-test-analytics-integration.md`
  - Full sequential flow (identify, process.start, tool, HTTP, process.end)
  - Event namespacing invariants (tool.*, http.*, process.start/end, error.*)
  - `app_version` stamping via `createMixpanelClient` with `appVersion` option
  - Header sanitization in HTTP events
  - Body snippet truncation invariants
  - Concurrent process tracking
  - Identify writes to people profile (not `track`)
- Extended `MixpanelClient` interface with optional `identify?` method
- Extended `createMixpanelClient` with `appVersion?` option that stamps on every `track`
- Extended `SdkOptions` with `appVersion?` option
- Total tests: 407 (up from 388)

## Remaining work
- `specs/nexai-test-basaglia-candidate-evaluation-lifecycle.md`: complex integration test for the full candidate-evaluation flow against the orchestrator in `packages/nexai-provider-basaglia/`
- `specs/nexai-test-basaglia-hr-support-lifecycle.md`: complex integration test for the hr-support RAG flow
- `specs/nexai-test-basaglia-screening-evaluation-lifecycle.md`: integration test for the screening-evaluation flow
