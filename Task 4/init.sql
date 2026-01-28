DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'webchat') THEN
        CREATE USER webchat WITH PASSWORD 'web1234';
    END IF;
END
$$;

DROP DATABASE IF EXISTS webchatdb;
CREATE DATABASE webchatdb 
    OWNER webchat 
    ENCODING 'UTF8' 
    LC_COLLATE = 'en_US.UTF-8' 
    LC_CTYPE = 'en_US.UTF-8' 
    TEMPLATE template0;

ALTER USER webchat WITH SUPERUSER;

\encoding UTF8
