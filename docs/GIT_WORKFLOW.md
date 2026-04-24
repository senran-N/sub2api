# Git Workflow

本文档定义当前 fork 的分支模型、上游同步方式和选择性移植流程。目标是在持续开发自有特性的同时，能够安全、可追踪地吸收 `upstream` 中真正需要的更新。

## 1. 远端角色

当前仓库默认使用两个远端：

| 远端 | 作用 | 写入规则 |
| --- | --- | --- |
| `origin` | 自己的 fork，承载当前项目主线和自有特性 | 可以 fetch / push |
| `upstream` | 原项目，只作为更新来源 | 只允许 fetch，不向上游 push |

检查远端配置：

```bash
git remote -v
```

`upstream` 的 push 地址应保持禁用或不可用，避免误推：

```bash
git remote set-url --push upstream DISABLED
```

## 2. 分支模型

| 分支模式 | 用途 | 生命周期 |
| --- | --- | --- |
| `main` | 当前 fork 的稳定主线，只放验证过的改动 | 长期保留 |
| `feature/<name>` | 自有特性开发 | 合入 `main` 后删除 |
| `sync/upstream-YYYYMMDD` | 整体同步上游的临时集成分支 | 验证通过并合入后删除 |
| `port/upstream-<name>` | 选择性移植上游某个特性、修复或提交组 | 验证通过并合入后删除 |
| `backup/pre-<action>-YYYYMMDD-HHMMSS` | 高风险操作前的本地保护点 | 确认稳定后按需清理 |

`main` 必须保持可构建、可测试、可回滚。不要在 `main` 上长期堆积未完成代码。

## 3. 开发自有特性

从最新 `main` 拉出特性分支：

```bash
git checkout main
git pull origin main
git checkout -b feature/my-feature
```

开发过程中保持提交聚焦：

- 一个提交只表达一个明确意图。
- 不把重构、格式化、业务行为变更混在同一个提交里。
- 涉及接口、配置、脚本、架构、部署或开发流程时，同步更新 `docs/`。

合回主线前先更新本地状态并验证：

```bash
git fetch origin upstream
git checkout feature/my-feature
git status --short
```

合入 `main`：

```bash
git checkout main
git pull origin main
git merge --no-ff feature/my-feature
git push origin main
```

## 4. 整体同步上游

当需要吸收上游近期整体更新时，不直接在 `main` 上合并，先创建同步分支：

```bash
git fetch upstream origin
git checkout main
git pull origin main
git branch backup/pre-upstream-merge-$(date +%Y%m%d-%H%M%S)
sync_branch=sync/upstream-$(date +%Y%m%d)
git checkout -b "$sync_branch"
git merge upstream/main
```

合并后必须检查差异：

```bash
git diff --stat main...HEAD
git diff main...HEAD
```

验证通过后再合入 `main`：

```bash
git checkout main
git merge --no-ff "$sync_branch"
git push origin main
```

整体同步适合下面场景：

- 上游改动规模可控，且与当前 fork 的差异不冲突。
- 需要跟随上游结构、依赖、接口或安全修复。
- 当前 fork 没有大量长期偏离上游的核心逻辑。

如果上游改动很大，优先选择第 5 节的选择性移植。

## 5. 选择性移植上游更新

当只需要上游某个特性或修复时，优先使用 `cherry-pick -x`，不要手工复制代码。

先查看上游比当前 `main` 多出的提交：

```bash
git fetch upstream
git log --oneline --decorate main..upstream/main
```

查看目标提交内容：

```bash
git show --stat <commit-sha>
git show <commit-sha>
```

创建移植分支并挑选提交：

```bash
git checkout main
git pull origin main
git checkout -b port/upstream-some-feature
git cherry-pick -x <commit-sha>
```

连续提交使用范围挑选：

```bash
git cherry-pick -x <oldest-sha>^..<newest-sha>
```

`-x` 会在提交信息中记录来源提交，后续排查问题时可以快速追溯到上游。

移植完成后验证并合入：

```bash
git checkout main
git merge --no-ff port/upstream-some-feature
git push origin main
```

只有在上游提交无法直接适配当前 fork 时，才允许手工移植。手工移植必须满足：

- 提交信息写明上游来源提交、PR 或文件路径。
- 只移植当前真正需要的行为，不顺手带入无关重构。
- 不复制重复逻辑，必要时先抽取已有公共能力。
- 同步检查 `docs/` 中对应契约是否需要更新。

## 6. 冲突处理规则

遇到冲突时，先理解双方意图，再编辑文件。不要机械选择 `ours` 或 `theirs`。

推荐流程：

```bash
git status --short
git diff --cc
```

逐个解决冲突后：

```bash
git add <resolved-file>
git status --short
```

继续当前操作：

```bash
# merge 场景
git commit

# cherry-pick 场景
git cherry-pick --continue
```

如果发现方向错误，及时中止：

```bash
git merge --abort
git cherry-pick --abort
```

冲突解决必须遵守以下原则：

- 保留当前 fork 已验证的自有特性，不被上游同名逻辑覆盖。
- 吸收上游修复时要消除根因，不为了通过编译加入无意义兜底。
- 对公共模块的修改优先保持低耦合、高内聚，不引入平行实现。
- 高频路径中的移植代码要检查内存分配、缓存和重复查询问题。

## 7. 同步前后检查清单

同步或移植前：

- `git status --short` 必须干净，或确认未提交改动与本次操作无关。
- `git fetch origin upstream` 已执行。
- 高风险操作前创建 `backup/pre-*` 分支。
- 已确认要整体同步，还是只移植目标提交。

同步或移植后：

- 用 `git diff` 检查真实改动，而不是只看冲突文件。
- 按 [docs/DEVELOPMENT_GUIDELINES.md](/home/senran/Desktop/sub2api/docs/DEVELOPMENT_GUIDELINES.md) 运行与改动范围匹配的验证。
- 涉及接口、配置、脚本、架构、部署或开发流程时，更新对应 `docs/`。
- 确认没有误改生成文件，除非同时更新了生成源。

常用对比命令：

```bash
git log --oneline upstream/main..main
git log --oneline main..upstream/main
git diff --stat main..upstream/main
git diff main..upstream/main
```

## 8. 推荐策略

默认策略是“小步同步、明确来源、先验后合”：

1. 自有功能用 `feature/*` 独立开发。
2. 上游大范围更新用 `sync/*` 先集成验证。
3. 上游单点能力用 `port/*` 和 `cherry-pick -x` 移植。
4. 每次高风险操作前保留 `backup/pre-*`。
5. 合入 `main` 前完成最小但足够的验证和文档同步。
