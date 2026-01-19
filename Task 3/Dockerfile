FROM registry.altlinux.org/alt/alt:p11

RUN apt-get update && apt-get install -y \
    golang \
    postgresql17 \
    postgresql17-server \
    nodejs \
    npm \
    sudo

RUN mkdir -p /var/lib/pgsql/data
RUN chown -R postgres:postgres /var/lib/pgsql

USER postgres
RUN initdb -D /var/lib/pgsql/data

USER root
RUN echo "root ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

RUN mkdir -p /opt/catpc
WORKDIR /opt/catpc

COPY backend/ backend/
COPY frontend/ frontend/
COPY init-db.sql .
COPY schema.sql .
COPY start.sh .

WORKDIR /opt/catpc/backend
RUN go build -mod=vendor -o catpc-backend

WORKDIR /opt/catpc/frontend
RUN npm ci && npm run build

RUN chmod +x /opt/catpc/start.sh

EXPOSE 1323 5173 5432

WORKDIR /opt/catpc
CMD ["/bin/bash", "/opt/catpc/start.sh"]