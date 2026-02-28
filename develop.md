运行主程序需要先启动依赖服务：

1. 启动 PostgreSQL
# 使用 Docker
docker run -d --name postgres \
  -e POSTGRES_USER=user \
  -e POSTGRES_PASSWORD=pass \
  -e POSTGRES_DB=app \
  -p 5432:5432 \
  postgres:16

2. 启动 Redis
# 使用 Docker
docker run -d --name redis \
  -p 6379:6379 \
  redis:7

3. 运行迁移
# 安装 migrate
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
# 确保路径在 PATH 中
export PATH=$PATH:$(go env GOPATH)/bin

make migrate-up

4. 启动应用
go run cmd/api/main.go

或编译后运行：
go build -o bin/api cmd/api/main.go
./bin/api

服务默认监听 :8080，可测试：
curl http://localhost:8080/users
