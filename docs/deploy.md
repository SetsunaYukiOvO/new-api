# New API Docker 部署指南

## 环境要求

- Linux 服务器（Debian/Ubuntu）
- 最低配置：1核 1G 内存

## 完整部署步骤

### 1. 安装 Docker

```bash
apt-get update
apt-get install -y docker.io docker-compose
systemctl start docker
systemctl enable docker
```

### 2. 创建项目目录

```bash
mkdir -p /www/wwwroot/new-api-main
cd /www/wwwroot/new-api-main
```

### 3. 创建 docker-compose.yml

```bash
cat > docker-compose.yml << 'EOF'
services:
  new-api:
    image: ghcr.io/setsunayukiovo/new-api:latest
    container_name: new-api
    restart: always
    command: --log-dir /app/logs
    ports:
      - "3000:3000"
    volumes:
      - ./data:/data
      - ./logs:/app/logs
    environment:
      - SQL_DSN=postgresql://root:YOUR_PASSWORD@postgres:5432/new-api
      - REDIS_CONN_STRING=redis://redis
      - TZ=Asia/Shanghai
      - ERROR_LOG_ENABLED=true
      - BATCH_UPDATE_ENABLED=true
    depends_on:
      - redis
      - postgres
    networks:
      - new-api-network
    healthcheck:
      test: ["CMD-SHELL", "wget -q -O - http://localhost:3000/api/status | grep -o '\"success\":\\s*true' || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:latest
    container_name: redis
    restart: always
    networks:
      - new-api-network

  postgres:
    image: postgres:15
    container_name: postgres
    restart: always
    environment:
      POSTGRES_USER: root
      POSTGRES_PASSWORD: YOUR_PASSWORD
      POSTGRES_DB: new-api
    volumes:
      - pg_data:/var/lib/postgresql/data
    networks:
      - new-api-network

volumes:
  pg_data:

networks:
  new-api-network:
    driver: bridge
EOF
```

> **重要：** 将 `YOUR_PASSWORD` 替换为你自己的强密码，`SQL_DSN` 和 `POSTGRES_PASSWORD` 中的密码必须一致。

### 4. 拉取镜像并启动

```bash
docker-compose pull
docker-compose up -d
```

### 5. 验证运行状态

```bash
docker-compose ps
```

三个容器（new-api、redis、postgres）都应显示为 `Up` 状态。

### 6. 访问

浏览器打开 `http://你的服务器IP:3000`

## 更新部署

代码推送到 GitHub main 分支后，GitHub Actions 会自动构建新的 Docker 镜像（约 4-5 分钟）。

在服务器上执行：

```bash
cd /www/wwwroot/new-api-main
docker-compose pull
docker-compose down
docker-compose up -d
```

## 查看日志

```bash
# 查看 new-api 容器日志
docker logs -f new-api

# 查看应用日志文件
ls /www/wwwroot/new-api-main/logs/
```

## 数据备份

数据存储位置：

| 数据 | 位置 |
|------|------|
| 应用数据 | `/www/wwwroot/new-api-main/data/` |
| 应用日志 | `/www/wwwroot/new-api-main/logs/` |
| PostgreSQL | Docker volume `new-api-main_pg_data` |

备份 PostgreSQL：

```bash
docker exec postgres pg_dump -U root new-api > backup.sql
```

恢复 PostgreSQL：

```bash
cat backup.sql | docker exec -i postgres psql -U root new-api
```

## 可选：使用 MySQL 替代 PostgreSQL

将 `docker-compose.yml` 中的 `postgres` 服务替换为：

```yaml
  mysql:
    image: mysql:8.2
    container_name: mysql
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: YOUR_PASSWORD
      MYSQL_DATABASE: new-api
    volumes:
      - mysql_data:/var/lib/mysql
    networks:
      - new-api-network
```

并将 `SQL_DSN` 改为：

```
SQL_DSN=root:YOUR_PASSWORD@tcp(mysql:3306)/new-api
```

同时在 `volumes` 中添加 `mysql_data:`，在 `depends_on` 中将 `postgres` 改为 `mysql`。
