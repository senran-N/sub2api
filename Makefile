.PHONY: build build-backend build-frontend build-datamanagementd
.PHONY: test test-backend test-frontend test-tools test-datamanagementd
.PHONY: lint lint-backend lint-frontend coverage coverage-backend
.PHONY: fmt fmt-backend clean clean-backend migrate-validate release-smoke-test secret-scan

# 一键编译前后端
build: build-backend build-frontend

# 编译后端（复用 backend/Makefile）
build-backend:
	@$(MAKE) -C backend build

# 编译前端（需要已安装依赖）
build-frontend:
	@pnpm --dir frontend run build

# 编译 datamanagementd（宿主机数据管理进程）
build-datamanagementd:
	@test -d datamanagement || { echo "datamanagement/ not found in this checkout"; exit 1; }
	@cd datamanagement && go build -o datamanagementd ./cmd/datamanagementd

# 运行测试（后端 + 前端 + 工具脚本）
test: test-backend test-frontend test-tools

test-backend:
	@$(MAKE) -C backend test

test-frontend:
	@pnpm --dir frontend run lint:check
	@pnpm --dir frontend run typecheck

test-tools:
	@PYTHONDONTWRITEBYTECODE=1 python3 tools/test_http_extreme_probe.py

test-datamanagementd:
	@test -d datamanagement || { echo "datamanagement/ not found in this checkout"; exit 1; }
	@cd datamanagement && go test ./...

lint: lint-backend lint-frontend

lint-backend:
	@$(MAKE) -C backend lint

lint-frontend:
	@pnpm --dir frontend run lint:check
	@pnpm --dir frontend run typecheck

coverage: coverage-backend

coverage-backend:
	@$(MAKE) -C backend coverage

fmt: fmt-backend

fmt-backend:
	@$(MAKE) -C backend fmt

clean: clean-backend

clean-backend:
	@$(MAKE) -C backend clean

migrate-validate:
	@$(MAKE) -C backend migrate-validate

release-smoke-test:
	@test -n "$(IMAGE)" || { echo "usage: make release-smoke-test IMAGE=ghcr.io/<owner>/sub2api:<tag>"; exit 2; }
	@chmod +x deploy/release-smoke-test.sh
	@./deploy/release-smoke-test.sh "$(IMAGE)"

secret-scan:
	@python3 tools/secret_scan.py
