#!/bin/bash

sudo -u postgres pg_ctl start -D /var/lib/pgsql/data
sleep 3

psql -U postgres -f /opt/webchat/init.sql
psql -U webchat -d webchatdb -f /opt/webchat/schema.sql

cd /opt/webchat/frontend
npx serve -s dist -l 3000 &

echo "========================================"
echo "WebChat запущен!"
echo "Frontend: http://localhost:3000"
echo "Backend: http://localhost:8080"
echo "PostgreSQL: localhost:5432"
echo "========================================"

cd /opt/webchat/backend
./webchat

wait
