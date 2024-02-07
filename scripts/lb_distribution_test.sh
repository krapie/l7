#!/bin/bash
for i in {1..100}; do curl -s -H "x-shard-key: apiKey/document-$i" http://localhost | grep Hostname; done | sort | uniq -c | sort -nr
