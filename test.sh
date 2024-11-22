# bin/bash

curl -X POST http://localhost:5000/alerts \
     -H 'Content-Type: application/json' \
     -H 'Authorization: Bearer YOUR_SECRET_TOKEN' \
     -d '{"alerts": [{"status": "firing", "labels": {"alertname": "TestAlert"}}]}'

# curl -X POST http://localhost:5000/alerts \
#      -H 'Content-Type: application/json' \
#      -H 'Authorization: Bearer YOUR_SECRET_TOKEN' \
#      -d '{
#        "alerts": [{
#          "status": "firing",
#          "labels": {"alertname": "TestAlert222222222"},
#          "annotations": {"summary": "This is a test alert"},
#          "startsAt": "2024-11-23T00:00:00Z",
#          "endsAt": "2024-11-23T01:00:00Z",
#          "generatorURL": "http://example.com/alert",
#          "fingerprint": "abc123"
#        }]
#      }'
