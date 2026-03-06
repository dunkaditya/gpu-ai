# GPU.ai Phase 1 — Architecture Diagrams

> Complete system workflow covering auth, billing, GPU provisioning, WireGuard networking, health monitoring, and error handling.

---

## 1. System Architecture Overview

```mermaid
flowchart TB
    %% ============================================================
    %% GPU.AI PHASE 1 — HIGH-LEVEL SYSTEM ARCHITECTURE
    %% ============================================================

    subgraph CLIENTS["Clients"]
        direction LR
        browser["Browser<br/>(Dashboard)"]
        cli["CLI / API<br/>Consumer"]
    end

    subgraph FRONTEND["Frontend — Next.js + Clerk"]
        direction TB
        clerkMW["Clerk Middleware<br/>(proxy.ts)"]
        pages["Dashboard Pages<br/>(SWR Data Fetching)"]
        apiClient["Typed API Client<br/>(api.ts)"]
        pages --> apiClient
    end

    subgraph GPUCTL["gpuctl — Go API Server (:9090)"]
        direction TB

        subgraph MIDDLEWARE["Middleware Chain"]
            direction LR
            mwClerk["ClerkAuth<br/>JWT + JWKS"]
            mwOrg["RequireOrg<br/>403 guard"]
            mwRate["OrgRateLimiter<br/>10 req/s burst 20"]
            mwIdemp["Idempotency<br/>POST dedup"]
            mwClerk --> mwOrg --> mwRate --> mwIdemp
        end

        subgraph PUBLIC_API["Public API — /api/v1"]
            direction TB
            instCRUD["Instances<br/>POST GET DELETE"]
            sshCRUD["SSH Keys<br/>POST GET DELETE"]
            billAPI["Billing<br/>Usage + Spending Limits"]
            gpuAvail["GPU Availability<br/>GET /gpu/available"]
            sseStream["SSE Streams<br/>/instances/id/events<br/>/events"]
        end

        subgraph INTERNAL_API["Internal API"]
            direction TB
            readyCB["POST /internal/.../ready<br/>(cloud-init callback)"]
            healthCB["POST /internal/.../health<br/>(health ping)"]
            intAuth["InstanceTokenAuth<br/>(per-instance token)"]
        end

        sseBroker["StatusBroker<br/>+ OrgEventBroker"]
    end

    subgraph CORE["Core Services"]
        direction TB
        engine["Provisioning Engine<br/>(provision.Engine)"]
        billingTicker["Billing Ticker<br/>(60s loop)"]
        availPoller["Availability Poller<br/>(30s loop)"]
        healthMon["Health Monitor<br/>(60s loop)"]
    end

    subgraph DATA["Data Layer"]
        direction LR
        pg[("PostgreSQL<br/>instances, orgs,<br/>billing_sessions,<br/>ssh_keys, events,<br/>spending_limits")]
        redis[("Redis<br/>availability cache,<br/>idempotency keys")]
    end

    subgraph INFRA["WireGuard Proxy Server"]
        direction TB
        wgMgr["WG Manager<br/>(wgctrl-go)"]
        ipam["IPAM<br/>(10.0.0.0/16)"]
        iptables["iptables<br/>DNAT + FORWARD"]
        wgIface["wg0 Interface"]
        wgMgr --> wgIface
        wgMgr --> iptables
    end

    subgraph EXTERNAL["External Services"]
        direction LR
        clerk["Clerk<br/>(Auth + JWKS)"]
        stripe["Stripe<br/>(Billing Meters)"]
        runpod["RunPod<br/>(GraphQL API)"]
    end

    subgraph INSTANCE["GPU Instance (RunPod Pod)"]
        direction TB
        cloudInit["Cloud-Init<br/>Bootstrap Script"]
        wgClient["WireGuard Client<br/>(tunnel to proxy)"]
        sshd["sshd :22"]
        cloudInit --> wgClient
        cloudInit --> sshd
    end

    %% === Client connections ===
    browser --> clerkMW
    cli --> mwClerk
    clerkMW --> pages

    %% === Frontend to API ===
    apiClient -->|"fetch /api/v1/*"| MIDDLEWARE

    %% === Middleware to routes ===
    MIDDLEWARE --> PUBLIC_API
    intAuth --> INTERNAL_API

    %% === Public API to core ===
    instCRUD -->|"Provision / Terminate"| engine
    gpuAvail -->|"Read cache"| redis
    billAPI -->|"Query sessions"| pg
    sshCRUD -->|"CRUD"| pg
    sseStream --> sseBroker

    %% === Core service connections ===
    engine -->|"Create/poll pods"| runpod
    engine -->|"Persist state"| pg
    engine -->|"WG key gen + AddPeer"| wgMgr
    engine -->|"Allocate IP"| ipam
    engine -->|"Status events"| sseBroker

    billingTicker -->|"Read sessions"| pg
    billingTicker -->|"Report GPU-seconds"| stripe
    billingTicker -->|"Stop/terminate"| engine

    availPoller -->|"ListAvailable"| runpod
    availPoller -->|"Cache offerings<br/>(+markup)"| redis

    healthMon -->|"GetStatus"| runpod
    healthMon -->|"Update state"| pg
    healthMon -->|"Publish events"| sseBroker

    %% === Internal callbacks ===
    INSTANCE -->|"POST /ready"| readyCB
    INSTANCE -->|"POST /health"| healthCB

    %% === Instance networking ===
    wgClient -.->|"WireGuard tunnel"| wgIface
    iptables -.->|"DNAT port:N -> 10.x.x.x:22"| wgClient

    %% === Auth ===
    mwClerk -.->|"Verify JWT"| clerk

    %% === Styling ===
    classDef external fill:#e8f4fd,stroke:#1976d2,stroke-width:2px,color:#0d47a1
    classDef data fill:#fce4ec,stroke:#c62828,stroke-width:2px,color:#b71c1c
    classDef core fill:#e8f5e9,stroke:#2e7d32,stroke-width:2px,color:#1b5e20
    classDef infra fill:#fff3e0,stroke:#e65100,stroke-width:2px,color:#bf360c
    classDef instance fill:#f3e5f5,stroke:#6a1b9a,stroke-width:2px,color:#4a148c

    class clerk,stripe,runpod external
    class pg,redis data
    class engine,billingTicker,availPoller,healthMon core
    class wgMgr,ipam,iptables,wgIface infra
    class cloudInit,wgClient,sshd instance
```

---

## 2. GPU Provisioning Lifecycle

```mermaid
sequenceDiagram
    autonumber
    participant Client as Client<br/>(Browser / CLI)
    participant Auth as Clerk Auth<br/>Middleware
    participant API as API Handler<br/>(handleCreateInstance)
    participant DB as PostgreSQL
    participant Engine as Provisioning<br/>Engine
    participant WG as WireGuard<br/>Manager
    participant IPAM as IPAM<br/>(10.0.0.0/16)
    participant Provider as RunPod<br/>(GraphQL)
    participant Pod as GPU Instance<br/>(Pod)
    participant SSE as SSE Broker

    Note over Client,SSE: === PROVISIONING REQUEST ===

    Client->>Auth: POST /api/v1/instances<br/>{gpu_type, gpu_count, tier, ssh_key_ids}
    Auth->>Auth: Verify Clerk JWT + JWKS
    Auth->>API: Inject SessionClaims (org_id, user_id)

    API->>DB: EnsureOrgAndUser(clerk_org_id, clerk_user_id)
    DB-->>API: internal org_id, user_id

    API->>API: Validate request body<br/>(gpu_type, gpu_count 1-8, tier)

    API->>Engine: engine.Provision(req)

    Note over Engine,Provider: === ENGINE ORCHESTRATION ===

    Engine->>Engine: Generate instance ID (gpu-XXXXXXXX)<br/>Generate internal token (32 hex)

    Engine->>DB: GetSSHKeysByIDs() or GetSSHKeysByUserID()
    DB-->>Engine: SSH public keys
    alt No SSH keys found
        Engine-->>API: ErrSSHKeysNotFound
        API-->>Client: 400 ssh-keys-not-found
    end

    Engine->>DB: checkSpendingLimit(org_id)
    alt Spending limit reached
        Engine-->>API: ErrSpendingLimitReached
        API-->>Client: 402 spending_limit_reached
    end

    Engine->>Provider: ListAvailable() (all providers)
    Provider-->>Engine: GPU offerings[]

    Engine->>Engine: Filter by gpu_type, gpu_count, tier, region<br/>Sort candidates by price ascending

    alt No matching providers
        Engine-->>API: ErrNoProvider
        API-->>Client: 409 no-availability
    end

    alt Price cap exceeded
        Engine-->>API: ErrPriceExceeded
        API-->>Client: 409 price-exceeded
    end

    Note over Engine,WG: === WIREGUARD SETUP (once, before retry loop) ===

    Engine->>WG: GenerateKeyPair()
    WG-->>Engine: {publicKey, privateKey}
    Engine->>Engine: EncryptPrivateKey (for DB storage)
    Engine->>IPAM: AllocateAddress() (in transaction)
    IPAM-->>Engine: tunnelIP (e.g. 10.0.0.5/16)
    Engine->>WG: AddPeer(publicKey, tunnelIP, externalPort)
    WG->>WG: wgctrl ConfigureDevice
    WG->>WG: iptables -t nat -A PREROUTING (DNAT)
    WG->>WG: iptables -A FORWARD (ACCEPT)
    Engine->>Engine: RenderBootstrap(cloud-init script)

    Note over Engine,Provider: === PROVIDER RETRY LOOP (max 3 attempts) ===

    loop Attempt 1..3 (sorted by price)
        Engine->>Provider: Provision(instanceID, gpuType, tier, startupScript)
        alt Success
            Provider-->>Engine: {upstreamID, costPerHour}
        else Failure
            Provider-->>Engine: error
            Engine->>Engine: Log warning, try next candidate
        end
    end

    alt All providers failed
        Engine->>WG: RemovePeer (best-effort cleanup)
        Engine-->>API: "all providers failed"
        API-->>Client: 500 provisioning-error
    end

    Note over Engine,DB: === PERSIST & RESPOND ===

    Engine->>DB: CreateInstance(full record)
    Engine-->>API: ProvisionResponse{instanceID, hostname, price, status: creating}

    API->>DB: GetInstance(instanceID)
    DB-->>API: Instance record
    API-->>Client: 201 InstanceResponse<br/>{id, status: starting, connection, price}

    Note over Engine,Pod: === ASYNC STATUS PROGRESSION (goroutine) ===

    Engine->>DB: UpdateStatus(creating -> provisioning)
    Engine->>SSE: Publish "provisioning"

    loop Poll every 5s (10min timeout)
        Engine->>Provider: GetStatus(upstreamID)
        Provider-->>Engine: {status, ip, ports}

        alt Provider says "running" (from provisioning)
            Engine->>DB: UpdateStatus(provisioning -> booting)
            Engine->>DB: CreateBillingSession (billing starts)
            Engine->>SSE: Publish "booting"
        end

        alt Provider says "running" + IP ready (from booting)
            Engine->>DB: SetInstanceRunning(ip)
            Engine->>SSE: Publish "running"
        end

        alt Provider says "terminated" / "error"
            Engine->>DB: SetInstanceError(reason)
            Engine->>DB: CreateZeroBillingSession (audit trail)
        end
    end

    Note over Pod,API: === CLOUD-INIT READY CALLBACK ===

    Pod->>Pod: Bootstrap: configure WireGuard tunnel,<br/>install SSH keys, set hostname
    Pod->>API: POST /internal/instances/{id}/ready<br/>(Bearer: internal_token)
    API->>DB: SetInstanceRunning()
    API->>SSE: Publish StatusEvent{status: running}
    API-->>Pod: 200 OK
```

---

## 3. Instance State Machine

```mermaid
stateDiagram-v2
    [*] --> creating : API creates instance

    creating --> provisioning : progressStatus goroutine<br/>(immediate)
    creating --> error : Engine failure
    creating --> stopping : User DELETE

    provisioning --> booting : Provider reports "running"<br/>(billing session starts)
    provisioning --> error : Provider reports error/terminated<br/>(zero billing session)
    provisioning --> stopping : User DELETE

    booting --> running : Provider "running" + IP ready<br/>OR cloud-init /ready callback
    booting --> error : Provider error<br/>(billing session active)
    booting --> stopping : User DELETE

    running --> stopping : User DELETE
    running --> stopped : Spending limit reached<br/>(billing ticker)
    running --> error : Health monitor detects failure<br/>or spot interruption

    stopped --> running : Spending limit removed
    stopped --> stopping : User DELETE
    stopped --> terminated : 72h auto-terminate<br/>(billing ticker)

    stopping --> terminated : Provider.Terminate() succeeds<br/>(billing closed, WG cleaned)
    stopping --> error : Provider.Terminate() fails

    error --> stopping : User DELETE (cleanup)

    terminated --> [*]

    note right of creating
        External state: "starting"
    end note

    note right of provisioning
        External state: "starting"
    end note

    note right of booting
        External state: "starting"
        Billing begins here
    end note

    note right of running
        External state: "running"
    end note

    note right of stopped
        External state: "stopped"
        Billing paused
    end note

    note left of error
        External state: "error"
        error_reason populated
    end note

    note left of terminated
        External state: "terminated"
        terminated_at set
        Billing closed
        WireGuard peer removed
    end note
```

---

## 4. Authentication & Middleware Pipeline

```mermaid
flowchart LR
    %% ============================================================
    %% AUTH & MIDDLEWARE PIPELINE
    %% ============================================================

    req["Incoming<br/>HTTP Request"] --> split{Route Type?}

    %% === PUBLIC API PATH ===
    split -->|"/api/v1/*"| clerkAuth

    subgraph AUTH_CHAIN["Authenticated Route Chain"]
        direction LR

        clerkAuth{"ClerkAuth<br/>Middleware"} -->|"CLERK_SECRET_KEY set"| jwtVerify["Verify JWT<br/>via Clerk JWKS"]
        clerkAuth -->|"CLERK_SECRET_KEY empty"| devUser["Inject dev-user<br/>dev-org claims"]

        jwtVerify --> jwtCheck{Valid JWT?}
        jwtCheck -->|No| reject401a["401 Unauthenticated"]
        jwtCheck -->|Yes| injectClaims["Inject SessionClaims<br/>into context"]
        devUser --> injectClaims

        injectClaims --> orgCheck{"RequireOrg<br/>Middleware"}
        orgCheck -->|"No active org"| reject403["403 org-required"]
        orgCheck -->|"Org present"| rateLim

        rateLim{"OrgRateLimiter"} -->|"> 10 req/s"| reject429["429 Too Many Requests"]
        rateLim -->|"Under limit"| methodCheck{POST?}

        methodCheck -->|Yes| idemp["Idempotency<br/>Middleware<br/>(check Redis)"]
        methodCheck -->|No| handler["Route Handler"]

        idemp --> idempCheck{Duplicate<br/>request?}
        idempCheck -->|Yes| cached["Return cached<br/>response"]
        idempCheck -->|No| handler
    end

    %% === INTERNAL CALLBACK PATH ===
    split -->|"/internal/*"| tokenAuth

    subgraph INTERNAL_AUTH["Instance Token Auth"]
        direction LR
        tokenAuth["Extract Bearer token<br/>from Authorization header"] --> lookupToken["DB lookup:<br/>match token to instance"]
        lookupToken --> tokenValid{Token valid<br/>for instance ID?}
        tokenValid -->|No| reject401b["401 Unauthorized"]
        tokenValid -->|Yes| intHandler["Internal Handler<br/>(ready / health)"]
    end

    %% === HEALTH ENDPOINT PATH ===
    split -->|"GET /health"| localOnly

    subgraph HEALTH_AUTH["Health Check Auth"]
        direction LR
        localOnly["LocalhostOnly<br/>Middleware"] --> internalAuth["InternalAPI<br/>TokenAuth"]
        internalAuth --> healthHandler["handleHealth<br/>(DB + Redis ping)"]
    end

    %% === CLAIMS EXTRACTION ===
    handler --> claimsExtract["auth.ClaimsFromContext()<br/>Returns {UserID, OrgID}"]
    claimsExtract --> dbLookup["db.EnsureOrgAndUser()<br/>or db.GetOrgIDByClerkOrgID()"]
    dbLookup --> bizLogic["Business Logic<br/>(scoped to org)"]

    %% === Styling ===
    classDef reject fill:#ffcdd2,stroke:#c62828,stroke-width:2px,color:#b71c1c
    classDef success fill:#c8e6c9,stroke:#2e7d32,stroke-width:2px,color:#1b5e20
    classDef check fill:#fff9c4,stroke:#f9a825,stroke-width:2px,color:#f57f17

    class reject401a,reject401b,reject403,reject429 reject
    class handler,intHandler,healthHandler,bizLogic success
    class jwtCheck,orgCheck,tokenValid,idempCheck,methodCheck check
```

---

## 5. Billing & Spending Limit Enforcement

```mermaid
flowchart TD
    %% ============================================================
    %% BILLING TICKER — 60-SECOND LOOP
    %% ============================================================

    start(("Billing Ticker<br/>fires every 60s")) --> getSessions["Get all active<br/>billing_sessions<br/>from PostgreSQL"]

    getSessions --> hasSessions{Any active<br/>sessions?}
    hasSessions -->|No| done(("Sleep 60s"))
    hasSessions -->|Yes| groupByOrg["Group sessions<br/>by org_id"]

    %% === PHASE 1: SPENDING LIMITS (FIRST) ===
    groupByOrg --> limitLoop

    subgraph LIMITS["Phase 1: Enforce Spending Limits (before Stripe)"]
        direction TB
        limitLoop["For each org"] --> getLimit["db.GetSpendingLimit()"]

        getLimit --> hasLimit{Limit<br/>configured?}
        hasLimit -->|No| skipOrg["Skip org"]
        hasLimit -->|Yes| cycleCheck{Billing cycle<br/>rolled over?}

        cycleCheck -->|Yes| resetSpend["db.ResetMonthlySpend()"]
        cycleCheck -->|No| calcSpend

        resetSpend --> calcSpend["db.GetOrgMonthSpendCents()"]
        calcSpend --> calcPct["Calculate<br/>spend percentage"]

        calcPct --> pct72{limit_reached_at<br/>+ 72 hours?}
        pct72 -->|Yes| autoTerminate["engine.TerminateStoppedInstancesForOrg()<br/>Force terminate all stopped"]

        pct72 -->|No| pct100{"spend >= 100%<br/>of limit?"}
        pct100 -->|Yes| stopAll["engine.StopInstancesForOrg()<br/>Stop all running instances<br/>Set limit_reached_at = now"]

        pct100 -->|No| pct95{"spend >= 95%?"}
        pct95 -->|Yes| warn95["Log SPEND_WARNING_95<br/>Set notify_95_sent = true"]

        pct95 -->|No| pct80{"spend >= 80%?"}
        pct80 -->|Yes| warn80["Log SPEND_WARNING_80<br/>Set notify_80_sent = true"]
        pct80 -->|No| skipOrg
    end

    %% === PHASE 2: STRIPE REPORTING (SECOND) ===
    LIMITS --> stripePhase

    subgraph STRIPE["Phase 2: Report to Stripe Billing Meters"]
        direction TB
        stripePhase{Stripe<br/>configured?}
        stripePhase -->|No| done2(("Sleep 60s"))
        stripePhase -->|Yes| calcUnreported["For each session:<br/>unreported = elapsed_seconds - stripe_reported_seconds"]

        calcUnreported --> aggregate["Aggregate unreported<br/>GPU-seconds by org"]
        aggregate --> lookupCust["db.GetOrgStripeCustomerID()<br/>for each org"]
        lookupCust --> buildBatch["Build MeterEventBatch[]<br/>{customer_id, gpu_seconds, timestamp}"]

        buildBatch --> sendStripe["stripe.ReportMeterEvents()<br/>One meter event per org"]
        sendStripe --> updateReported["db.UpdateStripeReportedSeconds()<br/>for each session"]
        updateReported --> done2
    end

    autoTerminate --> STRIPE
    stopAll --> STRIPE
    warn95 --> STRIPE
    warn80 --> STRIPE
    skipOrg --> STRIPE

    %% === BILLING SESSION LIFECYCLE ===
    subgraph SESSION_LIFECYCLE["Billing Session Lifecycle"]
        direction LR
        sessionCreate["CreateBillingSession<br/>(at booting state)"] --> sessionActive["Active Session<br/>(accumulating seconds)"]
        sessionActive --> sessionClose["CloseBillingSession<br/>(at DELETE or error)"]
        sessionClose --> sessionDone["Closed Session<br/>(final cost calculated)"]
    end

    %% === Styling ===
    classDef critical fill:#ffcdd2,stroke:#c62828,stroke-width:2px,color:#b71c1c
    classDef warning fill:#fff9c4,stroke:#f9a825,stroke-width:2px,color:#f57f17
    classDef stripe fill:#e8f4fd,stroke:#1976d2,stroke-width:2px,color:#0d47a1
    classDef ok fill:#c8e6c9,stroke:#2e7d32,stroke-width:2px,color:#1b5e20

    class autoTerminate,stopAll critical
    class warn95,warn80 warning
    class sendStripe,buildBatch,aggregate stripe
    class skipOrg,done,done2 ok
```

---

## 6. Health Monitoring & Error Recovery

```mermaid
flowchart TD
    %% ============================================================
    %% HEALTH MONITOR — 60-SECOND LOOP
    %% ============================================================

    start(("Health Monitor<br/>fires every 60s")) --> listActive["db.ListActiveInstances()<br/>(running + booting)"]

    listActive --> hasInst{Any active<br/>instances?}
    hasInst -->|No| sleep(("Sleep 60s"))
    hasInst -->|Yes| fanOut["Fan out checks<br/>(max 10 concurrent)"]

    fanOut --> checkInst["For each instance"]
    checkInst --> lookupProv["registry.Get(upstream_provider)"]

    lookupProv --> provFound{Provider<br/>found?}
    provFound -->|No| logWarn["Log warning<br/>(orphaned instance)"]
    provFound -->|Yes| pollStatus["provider.GetStatus(upstream_id)"]

    pollStatus --> pollOk{Poll<br/>succeeded?}
    pollOk -->|No| logPollFail["Log warning<br/>(will retry next tick)"]

    pollOk -->|Yes| statusCheck{Upstream<br/>status?}

    %% === HEALTHY ===
    statusCheck -->|"running"| healthy["Healthy<br/>(no action)"]

    %% === SPOT INTERRUPTION ===
    statusCheck -->|"terminated / error"<br/>AND tier = spot| spotPath

    subgraph SPOT["Spot Interruption Handler"]
        direction TB
        spotPath["Detected: spot interruption"] --> spotLock["Optimistic lock:<br/>UpdateInstanceStatus -> error"]
        spotLock --> spotLocked{Lock<br/>acquired?}
        spotLocked -->|No| spotSkip["Status changed concurrently<br/>(skip — already handled)"]
        spotLocked -->|Yes| spotError["SetInstanceError<br/>(spot instance interrupted)"]
        spotError --> spotBilling["CloseBillingSession<br/>(immediate — stop charges)"]
        spotBilling --> spotEvent["CreateInstanceEvent<br/>(type: interrupted)"]
        spotEvent --> spotSSE["Publish to OrgEventBroker<br/>(SSE notification)"]
    end

    %% === NON-SPOT FAILURE ===
    statusCheck -->|"terminated / error"<br/>AND tier = on_demand| nonSpotPath

    subgraph NONSPOT["Non-Spot Failure Handler (with retries)"]
        direction TB
        nonSpotPath["Detected: possible failure"] --> retryLoop

        retryLoop["Retry loop<br/>(3 attempts, 10s apart)"]
        retryLoop --> recheck["Re-read instance from DB"]
        recheck --> statusChanged{Status changed<br/>concurrently?}
        statusChanged -->|Yes| abortRetry["Abort retries<br/>(user terminated, etc.)"]
        statusChanged -->|No| repoll["provider.GetStatus()"]
        repoll --> recovered{Status =<br/>"running"?}
        recovered -->|Yes| falseAlarm["False alarm<br/>(transient blip)"]
        recovered -->|No| nextRetry{More<br/>retries?}
        nextRetry -->|Yes| retryLoop
        nextRetry -->|No| declareFailure["INSTANCE_FAILURE:<br/>all retries exhausted"]

        declareFailure --> failLock["Optimistic lock:<br/>UpdateInstanceStatus -> error"]
        failLock --> failError["SetInstanceError<br/>(provider reports terminated)"]
        failError --> failBilling["CloseBillingSession"]
        failBilling --> failEvent["CreateInstanceEvent<br/>(type: failed)"]
        failEvent --> failSSE["Publish to OrgEventBroker"]
    end

    %% Return paths
    logWarn --> sleep
    logPollFail --> sleep
    healthy --> sleep
    spotSSE --> sleep
    spotSkip --> sleep
    abortRetry --> sleep
    falseAlarm --> sleep
    failSSE --> sleep

    %% === Styling ===
    classDef critical fill:#ffcdd2,stroke:#c62828,stroke-width:2px,color:#b71c1c
    classDef warning fill:#fff9c4,stroke:#f9a825,stroke-width:2px,color:#f57f17
    classDef ok fill:#c8e6c9,stroke:#2e7d32,stroke-width:2px,color:#1b5e20
    classDef info fill:#e8f4fd,stroke:#1976d2,stroke-width:2px,color:#0d47a1

    class spotPath,declareFailure,spotError,failError critical
    class logWarn,logPollFail,nonSpotPath warning
    class healthy,falseAlarm,spotSkip,abortRetry ok
    class spotSSE,failSSE,spotEvent,failEvent info
```

---

## 7. Termination Flow

```mermaid
sequenceDiagram
    autonumber
    participant Client as Client
    participant API as API Handler<br/>(handleDeleteInstance)
    participant DB as PostgreSQL
    participant Engine as Provisioning<br/>Engine
    participant Provider as RunPod
    participant WG as WireGuard<br/>Manager
    participant SSE as SSE Broker

    Client->>API: DELETE /api/v1/instances/{id}
    API->>API: Extract claims (org_id)
    API->>DB: GetOrgIDByClerkOrgID()
    API->>DB: GetInstanceForOrg(id, org_id)

    alt Instance not found or wrong org
        API-->>Client: 404 not-found
    end

    alt Already terminated
        API-->>Client: 200 (idempotent — current state)
    end

    API->>Engine: engine.Terminate(instanceID)

    Note over Engine,WG: === TERMINATION SEQUENCE ===

    Engine->>DB: GetInstance(instanceID)

    Engine->>DB: UpdateInstanceStatus(current -> stopping)<br/>(optimistic lock on current state)

    alt Concurrent state change
        Engine->>DB: Re-read instance
        alt Now terminated
            Engine-->>API: nil (already done)
        end
    end

    Engine->>Provider: provider.Terminate(upstreamID)

    alt Provider terminate fails
        Engine->>DB: SetInstanceError(reason)
        Engine-->>API: error
        API-->>Client: 500 termination-error
    end

    Engine->>DB: TerminateInstance(instanceID)<br/>(set terminated_at, billing_end)

    Engine->>DB: CloseBillingSession(instanceID, now)

    Engine->>DB: CreateInstanceEvent(type: terminated)

    Note over Engine,WG: === WIREGUARD CLEANUP (best-effort) ===

    alt WG configured and instance has tunnel
        Engine->>WG: RemovePeer(publicKey, tunnelIP, port)
        WG->>WG: iptables -t nat -D PREROUTING
        WG->>WG: iptables -D FORWARD
        WG->>WG: wgctrl RemovePeer
    end

    Engine-->>API: nil (success)
    API->>DB: GetInstanceForOrg (re-fetch updated state)
    API-->>Client: 200 InstanceResponse{status: terminated}
```

---

## 8. WireGuard Networking Detail

```mermaid
flowchart TD
    %% ============================================================
    %% WIREGUARD TUNNEL SETUP & SSH ACCESS PATH
    %% ============================================================

    subgraph PROVISION["During Provisioning (Engine)"]
        direction TB
        genKeys["GenerateKeyPair()<br/>{publicKey, privateKey}"] --> encryptKey["EncryptPrivateKey()<br/>(AES for DB storage)"]
        encryptKey --> allocIP["IPAM.AllocateAddress()<br/>(in DB transaction)<br/>Returns: 10.0.x.y/16"]
        allocIP --> calcPort["PortFromTunnelIP()<br/>port = 10000 + ip[2]*256 + ip[3]<br/>e.g. 10.0.0.5 -> port 10005"]
        calcPort --> addPeer["WG Manager: AddPeer()"]
    end

    subgraph WG_SETUP["WireGuard Manager — AddPeer()"]
        direction TB
        addPeer --> configWG["wgctrl.ConfigureDevice(wg0)<br/>Add peer with AllowedIPs=/32<br/>Keepalive=25s"]
        configWG --> dnatRule["iptables -t nat -A PREROUTING<br/>-p tcp --dport PORT<br/>-j DNAT --to TUNNEL_IP:22"]
        dnatRule --> fwdRule["iptables -A FORWARD<br/>-p tcp -d TUNNEL_IP --dport 22<br/>-j ACCEPT"]
    end

    subgraph ROLLBACK["Rollback on Failure"]
        direction TB
        dnatRule -->|Fails| rb1["Rollback: remove WG peer"]
        fwdRule -->|Fails| rb2["Rollback: remove DNAT rule<br/>+ remove WG peer"]
    end

    subgraph CLOUDINIT["Cloud-Init Bootstrap Script"]
        direction TB
        fwdRule -->|Success| renderBS["RenderBootstrap() with:<br/>- InstancePrivateKey<br/>- ProxyEndpoint (203.0.113.1:51820)<br/>- ProxyPublicKey<br/>- InstanceAddress (10.0.x.y/16)<br/>- SSHAuthorizedKeys<br/>- CallbackURL"]
        renderBS --> validate["ValidateBootstrapData()<br/>- Regex check IDs, hostnames<br/>- SSH key format validation<br/>- Shell injection detection"]
    end

    subgraph POD_BOOT["GPU Instance Boot"]
        direction TB
        validate --> podStart["RunPod creates pod<br/>with dockerArgs = bootstrap script"]
        podStart --> wgConf["Configure WireGuard client:<br/>[Interface] PrivateKey, Address<br/>[Peer] PublicKey, Endpoint, AllowedIPs"]
        wgConf --> sshSetup["Write SSH authorized_keys<br/>Set hostname"]
        sshSetup --> tunnelUp["WireGuard tunnel established<br/>10.0.x.y <-> Proxy Server"]
        tunnelUp --> readyCB["POST /internal/instances/{id}/ready<br/>Bearer: internal_token"]
    end

    subgraph SSH_ACCESS["User SSH Access Path"]
        direction TB
        userSSH["ssh -p PORT root@PROXY_HOST"] --> proxyNAT["Proxy: iptables DNAT<br/>PORT -> 10.0.x.y:22"]
        proxyNAT --> wgTunnel["WireGuard tunnel<br/>Proxy wg0 -> Instance wg0"]
        wgTunnel --> instanceSSH["Instance sshd :22<br/>(authorized via SSH keys)"]
    end

    %% === Styling ===
    classDef keygen fill:#e8f4fd,stroke:#1976d2,stroke-width:2px
    classDef iptables fill:#fff3e0,stroke:#e65100,stroke-width:2px
    classDef security fill:#fce4ec,stroke:#c62828,stroke-width:2px
    classDef success fill:#c8e6c9,stroke:#2e7d32,stroke-width:2px

    class genKeys,encryptKey,allocIP keygen
    class dnatRule,fwdRule,proxyNAT iptables
    class validate,rb1,rb2 security
    class tunnelUp,instanceSSH,readyCB success
```

---

## Legend

| Color | Meaning |
|-------|---------|
| Blue (`#e8f4fd`) | External services / info states |
| Red (`#fce4ec` / `#ffcdd2`) | Data stores / critical errors |
| Green (`#c8e6c9` / `#e8f5e9`) | Success / healthy states |
| Orange (`#fff3e0`) | Infrastructure / WireGuard |
| Yellow (`#fff9c4`) | Warnings / decision points |
| Purple (`#f3e5f5`) | GPU instance internals |
