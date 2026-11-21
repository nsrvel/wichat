# search domain (ms-chat)

Kafka-driven message indexing into Elasticsearch. Public search/query RPCs are **deferred** (no `SearchService` proto yet).

## Layout

```text
internal/search/
  repository/   # EsRepository → wi-shared/infra/elasticsearch
  service/      # IndexKafkaValue, EnsureIndex
  delivery/
    kafka/      # SearchIndexer consumer (commit on success; retry + skip on failure)
```

## Events indexed

- `message.created` → index document
- `message.updated` → re-index body/metadata
- `message.deleted` → delete document by public message id

Other chat event types are ignored.