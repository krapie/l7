#!/bin/bash
for i in {1..100}; do curl -s http://localhost | grep Hostname; done | sort | uniq -c | sort -nr
