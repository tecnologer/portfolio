**REY DAVID DOMÍNGUEZ SOTO**
Senior Go Engineer  •  Distributed Systems  •  Backend, Infrastructure & Developer Tooling

[rdominguez@tecnologer.net](mailto:rdominguez@tecnologer.net)  •  [tecnologer.net](http://tecnologer.net)  •  Remote-first  •  San José del Cabo, México (CST/MST)

### SUMMARY

Senior Go Engineer with 8+ years in Go and 13+ years delivering production systems — building high-throughput distributed systems (gRPC, Protobuf, GraphQL), backend infrastructure, developer tooling (CLI/TUI), and MCP/agentic systems. I own architecture end-to-end, from the database layer through API and network design to CI/CD, and have delivered 10x–40x performance improvements by rebuilding systems from first principles. Comfortable close to the machine (concurrency, memory management, OS-level I/O, raw protocol handling) and equally at home operating my own bare-metal and self-hosted infrastructure. Daily Linux user (hand-configured distro, custom boot chain) and open source contributor to widely-used Go projects. I've led people directly (up to 8 reports), mentored engineers into new roles, and spoken publicly on topics from introductory engineering to advanced multithreaded performance. Proven record architecting microservices at scale, integrating 15+ third-party platforms, keeping migrations safe with parallel-run and fallback strategies, and leading fully remote delivery across distributed teams.

### EXPERIENCE

**DefectDojo Inc – Sr Software Engineer – Remote**
*Apr 2024 – Jul 2026  •  Widely adopted open-source DevSecOps / vulnerability-management platform*

* Designed the Translator Pattern — a standardized conversion architecture (Builder + Fluent Interface + Strategy) in which each vendor model owns its parsing logic, unifying 12+ connectors, eliminating duplicated code, and giving the team a single, predictable place to debug and extend.
* Expanded the platform's vulnerability ingestion pipeline by architecting and shipping 7+ vendor connectors from scratch — Checkmarx One, Wiz, Snyk, SonarCloud/SonarQube, Dependency-Track, IriusRisk, and Akamai — integrating each platform via REST and GraphQL APIs.
* Extended the connector interface hierarchy to support new grouping semantics, with changes isolated to a single connector while maintaining full backward compatibility across the platform.
* Implemented security workflow integrations with engineering and ITSM platforms (GitHub Boards, GitLab Boards, Azure DevOps, ServiceNow) as a dedicated microservice, syncing findings to external project-management tools.
* Diagnosed a silent-failure bug in which a connector imported nothing even though the vendor API returned data: traced a third-party library's swallowed panic ("scheduler recovered from fatal error"), patched a local fork to surface the stack trace, root-caused a field assumed to be always-present that arrived nil, and flagged the silent recovery as a reliability risk — restoring imports and preventing silent data loss for affected customers.
* Led development of the **Deployment Manager**, a Go CLI (urfave/cli) simplifying installation, upgrades, and lifecycle management for enterprise on-premise environments; re-architected it from config-threading-through-arguments to an idiomatic struct-based design with a runtime orchestrator, making it maintainable and extensible (Ubuntu and RHEL support).
* Built its end-to-end Linux provisioning command: preparing the filesystem (directories via `install`, correct `user:group` ownership), resolving configuration through a file → env → interactive fallback chain with license auto-detection, then pulling images, launching the multi-service platform via Docker Compose, and returning admin credentials once setup completed; contributed CI/linting improvements to internal tooling.
* Led development of the **Findings CLI**, a cross-platform Go tool (Go + TOML) enabling automated import, reimport, and export of vulnerability findings — designed to run unattended in CI/CD pipelines for scans without a native connector, automatically pushing discovered findings; adopted by customers and used in demos and QA workflows.
* Added an interactive TUI (Bubble Tea, my design) with basic/advanced modes that guides users through generating their TOML config for the first time — showing only required fields for quick imports or revealing optional configuration when needed; on completion users can save the TOML, emit a runnable shell script (`.sh`, plus `.ps1` for Windows), or execute the import/reimport immediately. User feedback: "the most beautiful tool we've ever been delivered"; used in customer demos and onboarding.
* Designed a multi-channel security notification system integrating Slack, Telegram, and PagerDuty to support real-time vulnerability alerting and incident workflows (pending release).

**Puller Tech – Freelance Software Consultant – Remote**
*Aug 2025 (1-week engagement)*

* Delivered a 40x throughput improvement (1K → 40K payrolls/min) by replacing a sequential TypeScript system with a distributed Go microservice using gRPC and Protobuf — evaluating client-specific formulas in parallel across leveled dependency stages.
* Kept the service stateless and decoupled from the client DB, passing computed values via the gRPC interface — delivered a working integration within a one-week constraint without introducing architectural dependencies.

**Pentalog – Sr Software Engineer & People Manager – Guadalajara, México**
*Dec 2019 – Feb 2024*

* Held a dual role from 2022 to 2024: as People Manager, owned the growth and unblocking of a team of engineers (avg 4, up to 8 reports), running biweekly 1:1s, removing technical and organizational blockers, and guiding course/skill plans when engineers sought a raise or a profile change.
* Served on the technical evaluation panel for Go and C# candidates, conducting conversational 1-hour interviews (background discussion plus a pair-programming exercise) and assessing not only working code but also the candidate's reasoning, testing habits, and ability to articulate time and space complexity.
* Identified a critical performance bottleneck (1,200 reports in 15–20s via Entity Framework) — built a Go + GORM proof-of-concept on personal time that returned 3,000+ reports in 2–3s, driving the business case for a full platform rewrite.
* Led architecture and delivery of Version 2.0 as a distributed Go microservices system on AWS Lambda + AppSync (GraphQL): an identity/access service (SAML-based SSO), a reporting service, and a generic CRUD forms engine — all backed by PostgreSQL.
* Ran the migration as a parallel run with fallback: V1 remained actively maintained (I shipped fixes to it during V2 development) and ran well past V2 launch, providing a robust rollback path.
* Designed the database transaction layer and the majority of the forms microservice; onboarded a second Go developer who implemented the reporting interface against Data Warehouse views.
* Integrated LLM agents (OpenAI API + LangChain) — a dual-mode conversational agent where prospects queried domain-specific information and admins issued natural-language commands, resolved into structured database operations.

**Ubilogix Inc – Sr Software Engineer – Ensenada, México**
*Mar 2017 – Dec 2019*

* Built a Go microservice (gRPC + Protobuf) to manage serial USB sniffers across multiple hardware vendors (Texas Instruments JN5169, Ubisys, Sewio), handling raw chunk assembly into complete packets and standardizing output via a local HTTP server — my first C#→Go migration, started as a downtime side project that became a product.
* On the C#/WPF processing side, achieved ~5x packet throughput (800 → 4,000+ packets/sec) by restructuring the class hierarchy and introducing multithreaded processing.
* Designed a custom SQLite-backed virtualized collection — keeping ~2K packets in memory with dynamic load/eviction on scroll — raising stable capture from ~370K packets (where the unvirtualized UI saturated RAM and the OS killed the process) to 19M packets while staying under 1.2GB RAM.
* Built a real-time Network Explorer in WPF: raw sniffer chunks → packet assembly → decryption → protocol decoding (Zigbee, Thread, and others) → live UI with user-configurable grouping, filtering, and display.
* Resolved a critical license-corruption bug caused by interrupted disk writes — implemented an idempotent write pattern guaranteeing state consistency before invalidating refresh tokens.

**Alesayi Development Co. (ADCO) – Software Architect / Full Stack Developer – Jeddah, Saudi Arabia**
*Aug 2015 – Jan 2017*

* Joined a 3-month-old ERP project with no shared architecture — established coding standards, defined the architectural pattern for the team, and reduced copy-paste code by ~99% across the codebase.
* Built a dynamic permission engine that scraped views to detect inputs and actions, then enforced visibility, edit, and execution rights per user — eliminating hardcoded permission logic.
* Owned a document and contract version-management system end-to-end for the CEO — from requirements gathering with the CEO and users to delivery — building the C# backend (consistent with the ERP stack) and an AngularJS 1.x front-end chosen for fast delivery; files stored with database-backed version tracking over an IT-provided disk cluster.
* Built inventory management and payroll (multiple contract types) modules; optimized critical stored procedures, improving overall system performance by ~20%.

**RedRabbit MX – Full Stack Developer – Culiacán, México**
*Jun 2012 – Jun 2015*

* Delivered full-stack web projects across multiple teams using MS SQL Server, ColdFusion, JavaScript, and HTML/CSS.

### SPEAKING & TEACHING

* **Advanced multithreaded performance:** delivered talks on concurrent programming in Go, using CPU and memory profilers to demonstrate real improvements and to show how to manage resource limits — making the case that effective concurrency is about bounded, measured parallelism, not just spawning goroutines.
* **Pentalog (internal & external):** presented technical talks internally and at external events, ranging from introductory to advanced topics, to audiences of roughly 15–30.
* **Instituto Tecnológico de Culiacán:** spoke at the Informatics Symposium, with audiences of roughly 40–60.
* **High-school outreach:** talks introducing software engineering and operating systems, plus a live demonstration of AI's power and guidance on preparing to integrate it into a professional career; audiences of roughly 25–35.

### OPEN SOURCE

* Merged contributions to widely-used Go projects: PagerDuty integration in `notify`, pgvector schema support in `langchaingo`.
* Built 2 MCP (Model Context Protocol) servers in Go from scratch: geospatial/environmental tooling, and a local-inference server with two-tier model delegation via Ollama.
* Contributed code to Spoolman (self-hosted, Docker), adding bulk filament loading, consumption logging, and rollback of recorded usage.
* Built [**jobtracker**](https://github.com/tecnologer/jobtracker), an open-source job-application tracker — Go REST API + Vue 3 SPA.
* Built **Cotiza3D**, a quoting app for 3D-printing services (Vue).

### HOMELAB / SELF-HOSTED INFRASTRUCTURE

* Linux daily driver since ~2010 (Ubuntu → Manjaro → Fedora → EndeavourOS with hand-configured systemd-boot / Secure Boot / dracut / ukify boot chain); Pop!_OS on laptop, macOS alongside — self-managed environments end to end.
* Dual-ISP setup for resilient remote work (TP-Link ER605 failover + AXE5400 Wi-Fi 6E); diagnosed and fixed a failover-without-failback bug to restore automatic recovery after an ISP drop.
* Self-hosted services on Raspberry Pi: a Telegram notification bot (Twitch alerts) and a dedicated print server; personal fleet of self-built services including battery-monitoring daemons with Telegram/desktop notification backends.
* 3D printer on Klipper/Moonraker (Creality K1C) with rooted firmware for full API access.
* Built [**Tempura**](https://github.com/tecnologer/tempura), an end-to-end IoT monitor for a vermicompost system: ESP8266 + SHT30 / liquid-level sensors → Go REST API (PostgreSQL, Docker, ARM build for the Pi). Designed the circuit and solar-battery power stage, 3D-printed the enclosures, and wrote the server — a hands-on atoms-to-bits build.

### SKILLS

* **Go:** gRPC · Protobuf · GraphQL · GORM · Cobra · urfave/cli · Bubble Tea (TUI) · golangci-lint · goroutines/channels · concurrency & parallelism patterns · context management · graceful shutdown · CLI/TUI development · standard networking & crypto libraries.
* **Systems & Architecture:** Microservices · distributed systems · event-driven design · REST/GraphQL/gRPC API design · concurrent/parallel processing · performance & latency optimization · observability · OS-level I/O and memory management · raw protocol handling (packet assembly, decode) · network protocol decoding (Zigbee, Thread) · serial/USB device management · memory-bounded data structures · Linux daemons on constrained hardware.
* **Migration & Performance:** Legacy-to-Go migration (Python, C#, TypeScript monoliths → Go microservices) · parallel-run with fallback/rollback · performance profiling & optimization (10x–40x gains) · production debugging & root-cause analysis.
* **Infrastructure:** Docker · Docker Compose · containerized service packaging & deployment · Kubernetes (hands-on: containerized full-stack Go app with PVC-backed persistence — Dockerfiles, Deployments, Services, PersistentVolumeClaims via kubectl/minikube) · Linux (self-managed daily driver + homelab) · AWS (Lambda, AppSync, DynamoDB, RDS) · GitHub Actions · CI/CD pipelines · release automation, build & packaging · immutable/reproducible deploys.
* **Data:** PostgreSQL · MSSQL · SQLite (memory-bounded virtualized collections, custom load/eviction) · Redis · database schema design · transaction layer design.
* **Integrations:** Third-party REST/GraphQL API integrations at scale (15+ external vendors and platforms) · security platforms (Checkmarx One, Wiz, Snyk, SonarCloud/SonarQube, Dependency-Track, IriusRisk, Akamai) · issue-tracker/ITSM integrations (GitHub, GitLab, Azure DevOps, ServiceNow) · notification/alerting integrations (Slack, Telegram, PagerDuty) · OAuth2 client flows (Slack) · SAML SSO.
* **Security:** DevSecOps tooling (vulnerability management) · secure protocols · TLS · idempotent/consistent write patterns · refresh-token invalidation · per-user permission engines.
* **Leadership:** Direct people management (up to 8 reports) · biweekly 1:1s · mentoring & career/skill development · technical interviewing · public speaking & technical teaching · architecture standards & design-pattern advocacy · cross-team technical communication.
* **AI/LLM & Tooling:** Model Context Protocol (MCP) — built 2 MCP servers in Go from scratch · agentic workflows & LLM tool integration · OpenAI API · LangChain · Ollama (local inference).
* **Other Languages & Frontend:** C# · Python · JavaScript · Vue 3 · Java · C++.

### EDUCATION

**B.S. Software Engineering — Instituto Tecnológico de Culiacán**

### CERTIFICATIONS

* **Foundations of Cybersecurity** — Google / Coursera — *Mar 2025*
* **Introduction to Artificial Intelligence** — LinkedIn Learning — *Mar 2025*
* **Generative AI for Software Developers Specialization** — IBM / Coursera — *Apr 2024*
* **Generative AI: Prompt Engineering Basics** — IBM / Coursera — *Apr 2024*
* **Generative AI: Introduction and Applications** — IBM / Coursera — *Mar 2024*

### LANGUAGES

Spanish — Native  •  English — Proficient
