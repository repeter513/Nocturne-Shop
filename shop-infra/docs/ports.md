# Порты локального стека

| Сервис | Порт (host) | URL |
|--------|-------------|-----|
| web | 3000 | http://localhost:3000 |
| bff | 8090 | http://localhost:8090 |
| envoy (gRPC gateway) | 443 | grpc://localhost:443 |
| envoy admin | 9901 | http://localhost:9901 |
| auth | 8081 | grpc://localhost:8081 |
| catalog | 8082 | grpc://localhost:8082 |
| cart | 8083 | grpc://localhost:8083 |
| order | 8084 | grpc://localhost:8084 |
| notification | 8085 | grpc://localhost:8085 |
| payment | 8086 | grpc://localhost:8086 |
| PostgreSQL | 5432 | postgres://shop:shop@localhost:5432 |
| Adminer | 8089 | http://localhost:8089 |

## Базы данных

| БД | Сервис |
|----|--------|
| auth_db | auth |
| catalog_db | catalog |
| cart_db | cart |
| order_db | order |
| payment_db | payment |
| notification_db | notification |
