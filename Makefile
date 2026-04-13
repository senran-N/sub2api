.PHONY: build build-backend build-frontend build-datamanagementd test test-backend test-frontend test-datamanagementd lint lint-backend lint-frontend coverage coverage-backend fmt fmt-backend clean clean-backend migrate-validate release-smoke-test secret-scan

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

# 运行测试（后端 + 前端）
test: test-backend test-frontend

test-backend:
	@$(MAKE) -C backend test

test-frontend:
	@pnpm --dir frontend run lint:check
	@pnpm --dir frontend run typecheck

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
