# WiChat data model (ERD)

## Apa yang ada sebelumnya?

Repo ini **belum punya ERD** terpisah. Yang ada:

- **C4** (bukan ERD): [`architecture-context.mermaid`](../architecture-context.mermaid), [`architecture-container.mermaid`](../architecture-container.mermaid)
- **Skema yang sudah jalan**: SQL migrations di [`backend/ms-user/internal/migrations/`](../../../backend/ms-user/internal/migrations/)
- **Produk / domain**: [`wichat-system-design.md`](../wichat-system-design.md) §6–§7 (belum diturunkan ke kolom per tabel)

File di folder ini adalah **sumber gambaran ERD** yang bisa di-preview dan di-review bareng tim.

## File

| File | Isi | Status |
|------|-----|--------|
| [`implemented-auth.dbml`](./implemented-auth.dbml) | Tabel `public` dari migrations ms-user | **Implemented** — harus match migrations |
| [`planned-overview.dbml`](./planned-overview.dbml) | Channel, message, RBAC lanjutan, audit (per §6) | **Planned** — belum ada migrations |
| [`implemented-auth.mermaid`](./implemented-auth.mermaid) | Subset yang sama untuk preview cepat | **Implemented** |

**Istilah:** di UI/produk **Group** = workspace di spec; tabel `groups` = workspace.

## Cara preview

### 0. Web viewer (lokal)

Jalankan web client, buka **`/dev/data-model`** (mis. `http://localhost:8080/dev/data-model`).

Tab **Implemented** / **Planned** — render file `.mermaid` dari folder ini (ikut dark mode).

### 1. Mermaid (paling gampang di GitHub / Cursor)

- Buka [`implemented-auth.mermaid`](./implemented-auth.mermaid)
- **GitHub**: preview diagram otomatis di file `.mermaid` (atau paste ke Markdown dengan fence ` ```mermaid `)
- **Cursor / VS Code**: extension “Mermaid” → preview diagram

### 2. DBML (paling matang untuk kolom + relasi)

- Buka [`implemented-auth.dbml`](./implemented-auth.dbml) atau [`planned-overview.dbml`](./planned-overview.dbml)
- **[dbdiagram.io](https://dbdiagram.io)** → *Import* → paste isi `.dbml`
- **VS Code / Cursor**: extension “DBML Entity Relationship Diagrams” → preview dari file `.dbml`

### 3. Dari database lokal

Setelah `docker compose` + migrations:

```bash
# contoh (butuh pg_dump + schemaonly)
pg_dump --schema-only --schema=public --table='users' --table='groups' "$DATABASE_URL"
```

Bandingkan output dengan `implemented-auth.dbml`; kalau beda, **migrations menang**, lalu update DBML.

## Aturan maintain

1. **Implemented:** ubah schema hanya lewat goose migration → update `implemented-auth.dbml` + `implemented-auth.mermaid` di PR yang sama.
2. **Planned:** ubah `planned-overview.dbml` saat spec §6 disepakati; baru jadi *implemented* setelah migration land.
3. **Ownership service** (satu cluster Postgres, satu `public` schema atau prefix tabel per service): tidak ada cross-service SQL join di runtime.

## Layanan vs data (target)

| Area | Service | Keterangan |
|------|---------|------------|
| `users`, `groups`, … | ms-user | Auth, instance, groups, membership (implemented, `public`) |
| channels, profiles, … | ms-user | Planned — tabel baru di `public` atau prefix jelas |
| messages, … | ms-chat | Planned |
| audit events | ms-audit | Planned |

Redis, Elasticsearch, MinIO, Kafka **tidak** masuk ERD relasional ini.
