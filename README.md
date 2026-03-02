# Go + Next.js Scaffold

全栈开发模板，基于 Go/Gin 后端 + Next.js 前端。

## 系统功能概览

### 后端 (Go/Gin)

- **用户管理**: 完整的 CRUD 接口
- **数据库**: PostgreSQL + sqlc 类型安全查询
- **缓存**: Redis Cache-Aside 模式
- **迁移**: golang-migrate

### 前端 (Next.js)

- **用户管理**: 列表、创建、编辑、删除
- **UI**: shadcn/ui + Tailwind CSS
- **状态**: TanStack Query (服务端) + Zustand (全局)
- **表单**: React Hook Form + Zod 验证

## 快速开始

### 后端

```bash
# 1. 启动 PostgreSQL
docker run -d --name postgres \
  -e POSTGRES_USER=xmap \
  -e POSTGRES_PASSWORD=xmap \
  -e POSTGRES_DB=app \
  -p 5432:5432 \
  postgres:16

# 2. 启动 Redis
docker run -d --name redis \
  -p 6379:6379 \
  redis:7

# 3. 运行迁移
export DB_URL="postgres://xmap:xmap@localhost:5432/app?sslmode=disable"
make migrate-up

# 4. 启动后端
go run cmd/api/main.go
```

### 前端

```bash
cd web
npm install
npm run dev
```

访问 http://localhost:3000/users

## 后端开发指南

详细架构说明见 [backend_architect.md](backend_architect.md)。

### 目录结构

```
/cmd/api/main.go           # 入口
/internal
  /api/http               # Gin handlers
  /domain                 # 错误码
  /repo                   # 数据访问层
  /service                # 业务逻辑
  /db                     # PostgreSQL
  /cache                  # Redis
/internal/gen/sqlc       # sqlc 生成代码
/sql                     # SQL 查询
/migrations              # 数据库迁移
```

### 添加新实体

1. **创建迁移**: `make migrate-new name=create_xxx`
2. **编写 SQL**: `sql/xxx.sql`
3. **生成代码**: `make sqlc`
4. **实现 Repo**: `internal/repo/xxx_repo.go`
5. **实现 Service**: `internal/service/xxx_service.go`
6. **实现 Handler**: `internal/api/http/handler_xxx.go`
7. **注册路由**: `cmd/api/main.go`

### API 命令

```bash
make sqlc           # 生成 sqlc 代码
make migrate-up     # 执行迁移
make migrate-down   # 回滚迁移
make migrate-new name=xxx  # 创建新迁移
```

## 前端开发指南

### 目录结构

```
web/src/
├── app/                    # Next.js App Router
│   ├── (auth)/           # 认证页面
│   ├── (dashboard)/      # 管理后台
│   └── page.tsx          # 首页
├── components/
│   ├── ui/              # shadcn/ui 组件
│   ├── forms/           # 表单组件
│   └── users/           # 业务组件
├── lib/
│   ├── api.ts           # API 客户端
│   └── utils.ts         # 工具函数
├── stores/              # Zustand stores
├── types/               # TypeScript 类型
└── schemas/             # Zod 验证 schemas
```

### 添加新页面

1. **创建页面**: `src/app/(dashboard)/xxx/page.tsx`
2. **添加导航**: `src/components/layout/nav-header.tsx`
3. **定义类型**: `src/types/index.ts`
4. **定义 Schema**: `src/schemas/xxx-schema.ts`

### 添加 shadcn/ui 组件

```bash
npx shadcn-ui@latest add button
```

## 环境变量

### 后端

无配置文件，环境变量在 `cmd/api/main.go` 中硬编码。如需配置化，可使用 `.env` 或环境变量。

### 前端

在 `web/.env.local` 中配置：

```
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## 技术栈

### 后端

- Go 1.25+
- Gin
- PostgreSQL + pgx/v5
- sqlc
- Redis
- golang-migrate

### 前端

- Next.js 14 (App Router)
- Tailwind CSS
- shadcn/ui
- TanStack Query
- Zustand
- React Hook Form + Zod

## License

MIT
