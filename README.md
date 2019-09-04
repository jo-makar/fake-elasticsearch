# fake-elasticsearch

Simulate an Elasticsearch node

Primarily interested in providing behavior that mimics the [Bulk API](https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html) to support ingesting data from Beat agents.

| Beat agent | Elasticsearch endpoint | Notes |
| ---------- | ---------------------- | ----- |
| filebeat 7.3.1 | GET / | TODO Where is this documented? |
| filebeat 7.3.1 | GET /_xpack?human=false | [Info API](https://www.elastic.co/guide/en/elasticsearch/reference/current/info-api.html) |
| filebeat 7.3.1 | GET,PUT /_ingest/pipeline/&lt;id&gt; | [Ingest APIs](https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest-apis.html) |

## TLS support

Generate a private key and self-signed certificate with:

`openssl req -x509 -nodes -newkey rsa:4096 -keyout server.key -out server.crt -subj '/O=Acme'`

Then replace http.ListenAndServe with http.ListenAndServeTLS in main.go
