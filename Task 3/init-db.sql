CREATE USER barsikuser WITH PASSWORD 'barsik_password';
CREATE DATABASE barsikdb OWNER barsikuser;
ALTER USER barsikuser WITH SUPERUSER;