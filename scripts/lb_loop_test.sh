#!/bin/bash
for i in {1..1000}; do curl -s -H "x-shard-key: apiKey/document-$i" http://localhost | grep Hostname; sleep 0.5; done
