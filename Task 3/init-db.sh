#!/bin/bash

for i in {1..15}; do
    postgres -u psql -c "SELECT 1;" > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        break
    fi
    sleep 2
done

if [ -f "/opt/catpc/barsikdb.sql" ]; then
    postgres -u psql -f /opt/catpc/barsikdb.sql
fi