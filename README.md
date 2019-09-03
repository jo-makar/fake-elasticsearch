# fake-elasticsearch

Simulate an Elasticsearch node

Primarily interested in providing behavior that mimics the [Bulk API](https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html) to support ingesting data from Beat agents.

| Beat agent | Elasticsearch endpoint | Notes |
| ---------- | ---------------------- | ----- |
| filebeat 7.3.1 | GET / | |
