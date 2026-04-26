# Git Workflow

本文档定义当前 fork 的 main-only 工作方式、上游审计方式和选择性移植规则。当前目标是让 `main` 成为唯一长期开发主线，减少分支漂移和语义不清。

## 1. 远端角色

当前仓库默认使用两个远端：

| 远端 | 作用 | 写入规则 |
| --- | --- | --- |
| `origin` | 自己的 fork，承载当前项目主线 | 可以 fetch / push |
| `upstream` | 原项目，只作为审计和选择性移植来源 | 只允许 fetch，不向上游 push |

检查远端配置：

```bash
git remote -v
```

`upstream` 的 push 地址应保持禁用或不可用，避免误推：

```bash
git remote set-url --push upstream DISABLED
```

## 2. Main-Only 主线策略

当前 fork 不默认创建 `feature/*`、`sync/*`、`port/*` 或 `backup/*` 分支。日常整理、重构、修复和选择性上游移植都直接在 `main` 上完成。

`main` 必须满足：

- 本地开始修改前先确认 `git status --short --branch`。
- 每组改动保持聚焦，避免把文档、格式化、业务行为和大规模重构混成不可审计的一团。
- 修改接口、配置、脚本、架构、部署或开发流程时，同步更新 `docs/`。
- 合入远端前运行与改动范围匹配的验证，失败必须记录具体阻塞和恢复步骤。

## 3. 在 Main 上开发

开始前：

```bash
git checkout main
git pull origin main
git status --short --branch
```

实施中：

- 小步修改，小步验证。
- 用 `git diff --stat` 和 `git diff` 审计真实改动。
- 不使用 `git reset --hard` 或 `git checkout -- <file>` 覆盖未确认来源的改动。
- 如遇到无关未提交改动，保留并绕开；如影响当前任务，先理解再处理。

完成后：

```bash
git status --short --branch
git diff --check
```

根据改动范围运行 [docs/DEVELOPMENT_GUIDELINES.md](/home/senran/Desktop/sub2api/docs/DEVELOPMENT_GUIDELINES.md) 中的验证命令。

## 4. 上游更新审计

默认不整体 merge `upstream/main`。上游只作为更新来源，先审计，再决定是否移植。

查看上游变化：

```bash
git fetch upstream origin
git log --oneline --decorate main..upstream/main
git diff --stat main..upstream/main
```

查看单个提交：

```bash
git show --stat <commit-sha>
git show <commit-sha>
```

移植前先按主题分类：

- 必须移植：安全修复、兼容性修复、会影响当前 fork 正常使用的上游改进。
- 可评估：和当前 fork 架构方向一致，但需要适配已有重构边界。
- 暂缓：大功能、迁移、UI 大改或与当前 fork 方向冲突的变更。
- 跳过：版本号、上游 CI 噪音、与当前 fork 无关的改动。

## 5. 选择性移植

优先使用 `cherry-pick -x` 保留来源信息，但仍然直接在 `main` 上操作：

```bash
git checkout main
git pull origin main
git fetch upstream
git cherry-pick -x <commit-sha>
```

连续提交使用范围挑选：

```bash
git cherry-pick -x <oldest-sha>^..<newest-sha>
```

如果冲突或方向不对，立即中止当前移植：

```bash
git cherry-pick --abort
```

只有在上游提交无法直接适配当前 fork 时，才允许手工移植。手工移植必须满足：

- 在提交说明或 LongCodex handoff 中写明上游来源提交、PR 或文件路径。
- 只移植当前真正需要的行为，不顺手带入无关重构。
- 不复制重复逻辑，必要时先复用已有公共能力。
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
git cherry-pick --continue
```

冲突解决必须遵守以下原则：

- 保留当前 fork 已验证的自有特性，不被上游同名逻辑覆盖。
- 吸收上游修复时要消除根因，不为了通过编译加入无意义兜底。
- 对公共模块的修改优先保持低耦合、高内聚，不引入平行实现。
- 高频路径中的移植代码要检查资源释放、缓存和重复查询问题。

## 7. 检查清单

修改或移植前：

- `git status --short --branch` 已确认。
- `git fetch origin upstream` 已执行，或明确本次不需要上游状态。
- 已确认本次是本 fork 自有改动、上游选择性移植，还是纯文档/流程整理。
- 高风险改动已有 LongCodex handoff 或证据文件记录恢复点；不通过临时分支表达恢复点。

修改或移植后：

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
