# bin/bash

curl -X POST http://localhost:5000/alerts \
     -H 'Content-Type: application/json' \
     -H 'Authorization: Bearer YOUR_SECRET_TOKEN' \
     -d '{"alerts": [{"status": "firing", "labels": {"alertname": "TestAlert"}}]}'
