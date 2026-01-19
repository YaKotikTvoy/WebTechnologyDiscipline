#!/bin/bash

sudo -u postgres pg_ctl start -D /var/lib/pgsql/data
sleep 3

psql -U postgres -f /opt/catpc/init-db.sql

PGPASSWORD=barsik_password psql -U barsikuser -d barsikdb -f /opt/catpc/schema.sql

cd /opt/catpc/backend
./catpc-backend &

cd /opt/catpc/frontend
npx serve -s dist -l 5173 &

echo "Frontend: http://localhost:5173"
echo "Backend: http://localhost:1323"
echo "PostgreSQL: localhost:5432"

wait