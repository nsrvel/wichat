# WiChat

Self-hosted, plugin-based team communication — structured enough for work, flexible enough for communities.

WiChat sits between casual chat (Discord, WhatsApp) and rigid enterprise suites (Microsoft Teams, Mattermost). Core chat stays lightweight; task boards, whiteboards, and optional spatial “virtual office” features ship as plugins, not baked-in bloat.

**Status:** Pre-development (architecture finalized). See [System Design](docs/architecture/wichat-system-design.md).

## Why WiChat

| | WiChat | Typical OSS chat | Enterprise SaaS |
|---|---|---|---|
| **Hosting** | Self-hosted, your data | Self-hosted | Vendor cloud |
| **Shape** | Flat channels + pro UX | Often corporate or casual-only | Heavy hierarchy |
| **Extensibility** | First-class plugins (gRPC + UI bundles) | Varies | App store / integrations |

Target scale: **2–500 people per workspace** on a single instance (`docker compose`), without sharding at launch.

## Repository layout

```
backend/     # Go microservices + wi-shared (see backend/README.md)
frontend/    # apps/web + @wichat packages (see frontend/README.md)
deploy/      # Docker Compose & ops
docs/        # Architecture & product spec
```

Monorepo ownership (env, gitignore, modules): [docs/architecture/monorepo-layout.md](docs/architecture/monorepo-layout.md).

## Architecture (short)

- **Core:** `backend/` — Go monorepo microservices — see system design §5.1
- **Realtime media:** [LiveKit](https://livekit.io/) (self-hosted SFU)
- **Data:** PostgreSQL, Redis, Elasticsearch, MinIO, Kafka
- **Clients:** Next.js (web + iOS PWA), Electron (desktop), React Native (Android APK)

Diagrams: [C4 Context](docs/architecture/architecture-context.mermaid) · [C4 Containers](docs/architecture/architecture-container.mermaid)

## License

WiChat is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**. See [LICENSE](LICENSE).

If you modify WiChat and offer it as a network service, AGPL requires making corresponding source available to users. Commercial support, hosting, and proprietary plugins on top of the AGPL core are compatible with this model (similar to Mattermost, Rocket.Chat, n8n).

## Self-hosting

Install documentation will land closer to the first release. Deployment target: one `docker-compose.yml` stack (Postgres, Redis, MinIO, Kafka, Elasticsearch, LiveKit, WiChat core).

Deploying organizations act as **data controllers** — customize [PRIVACY_TEMPLATE.md](PRIVACY_TEMPLATE.md) for your instance.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Security issues: [SECURITY.md](SECURITY.md).

## Roadmap (phases)

| Phase | Focus |
|------:|--------|
| 0 | Repo, compose, Go skeleton, CI |
| 1 | Auth & workspaces |
| 2 | Core chat |
| 3 | Voice/video (LiveKit) |
| 4 | Plugin manager |
| 5 | Guest access |
| 6 | Search, notifications, i18n |
| 7 | Real plugins |
| 8 | Hardening (Kafka, audit, observability) |

Details: [System Design §15](docs/architecture/wichat-system-design.md#15-build-roadmap-phase-summary).
