# Fix: VitePress base 路径导致 Pages 布局错乱

> **For agentic workers:** REQUIRED SUB-SKILL: `superpowers:subagent-driven-development`
> Steps use checkbox (`- [ ]`) syntax.

**Goal:** 修复线上 GitHub Pages 布局错乱——根因是 VitePress `base` 配置缺失，资源 URL 锚定在域名根路径而非 `/sonatype-central-skills/` 子路径，导致 CSS/JS 全部 404。

**Architecture:** VitePress 构建时把所有资源引用写成绝对路径。`base` 决定这个前缀：`base:'/'` 产出 `/assets/xxx.css`，`base:'/sonatype-central-skills/'` 产出 `/sonatype-central-skills/assets/xxx.css`。当前 config.ts 无 base → 默认 `/` → 资源 404 → 无样式。数据流：config.ts 加 base → `npm run build` → dist 资源路径带子前缀 → push → website.yml CI 重新部署 → Pages 资源 200 → 布局恢复。复用现有 CI（website.yml + deploy-pages），不改 CI，只改一个配置字段 + push 已有提交。

**Tech Stack:** VitePress 1.6.4, vitepress-plugin-mermaid 2.0.17, GitHub Actions (website.yml, deploy-pages@v4), GitHub Pages (sonatype-central-skills 子路径)

**Root Cause:**
- 线上 HTML 第7行 `<link rel="preload stylesheet" href="/assets/style.BOFw7W2O.css">` —— 资源路径锚定域名根
- `curl https://scagogogo.github.io/assets/style.BOFw7W2O.css` → **HTTP 404**
- `curl https://scagogogo.github.io/sonatype-central-skills/assets/style.BOFw7W2O.css` → **HTTP 200**（资源实际部署在子路径）
- `window.__VP_SITE_DATA__` 里 `"base":"/"`，应为 `"/sonatype-central-skills/"`
- `website/.vitepress/config.ts` 无 `base` 字段（VitePress 默认 `/`）

**Risks:**
- base 改动会影响所有内部链接与资源路径 → 缓解：VitePress 对 base 一致处理，构建后用 curl 抽查 2-3 个资源验证
- 本地 dev server 不受 base 影响（dev 用根路径），无法本地复现 404 → 缓解：靠 `npm run build` + 检查 dist HTML 的资源路径是否带子前缀验证
- 线上当前是旧提交（aef267e），push 后 CI 部署新版（be4f9a9 + 本次 fix）→ 缓解：push 后用 curl 复验线上资源 200

---

### Task 1: 给 config.ts 加 base 配置

**Depends on:** None
**Files:**
- Modify: `website/.vitepress/config.ts:6-7`

- [ ] **Step 1: 修改 config.ts — 在 withMermaid 配置对象顶部加 base 字段**

文件: `website/.vitepress/config.ts:6-7`（`withMermaid({` 之后、`lang:` 之前）

```typescript
export default withMermaid({
  // 站点部署在 GitHub Pages 子路径 /sonatype-central-skills/ 下，
  // 必须显式声明 base，否则所有 CSS/JS/图片资源会锚定域名根路径而 404，
  // 导致布局错乱（无样式裸 HTML）。VitePress 默认 base 为 '/'，仅适用于根域名部署。
  base: '/sonatype-central-skills/',
  lang: 'zh-CN',
```

说明：只新增 4 行（1 行配置 + 3 行注释），不改其他。`withMermaid` 透传 base 给 VitePress。

- [ ] **Step 2: 构建并验证 dist 资源路径带子前缀 — 确认 base 生效**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills/website && rm -rf .vitepress/dist && npm run build 2>&1 | tail -3 && echo "--- 检查首页HTML资源路径 ---" && grep -oE 'href="/[^"]*\.css|src="/[^"]*\.js' .vitepress/dist/index.html | head -5`

Expected:
  - Exit code: 0
  - build 输出含 "build complete"
  - 资源路径全部以 `/sonatype-central-skills/assets/` 开头，**不含** 裸 `/assets/`（根路径）

- [ ] **Step 3: 验证 __VP_SITE_DATA__ 的 base 字段已更新**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills/website && grep -oE '"base":"[^"]*"' .vitepress/dist/index.html`

Expected:
  - Exit code: 0
  - Output: `"base":"/sonatype-central-skills/"`（不是 `/`）

---

### Task 2: 本地 preview 复验渲染正常

**Depends on:** Task 1
**Files:**
- 无修改（仅本地起服务验证）

- [ ] **Step 1: 本地 preview 构建产物并抓取首页 — 确认资源加载无 404**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills/website && (nohup npm run preview -- --port 5189 > /tmp/vp-preview.log 2>&1 &) && sleep 4 && echo "--- 首页 HTTP 状态 ---" && curl -s -o /dev/null -w "%{http_code}\n" http://localhost:5189/sonatype-central-skills/ && echo "--- CSS 资源状态 ---" && CSS=$(grep -oE '/sonatype-central-skills/assets/[^"]*\.css' .vitepress/dist/index.html | head -1) && curl -s -o /dev/null -w "%{http_code}\n" "http://localhost:5189${CSS}"`

Expected:
  - Exit code: 0
  - 首页返回 200
  - CSS 资源返回 200（无 404）

- [ ] **Step 2: 关闭 preview server**

Run: `pkill -f "vitepress preview" 2>/dev/null; pkill -f "5189" 2>/dev/null; echo "preview 已关闭"`

Expected:
  - Exit code: 0

---

### Task 3: push 并复验线上 Pages

**Depends on:** Task 1, Task 2
**Files:**
- 无修改（git push + 等 CI + curl 复验）

- [ ] **Step 1: 提交 base 修复**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && git add website/.vitepress/config.ts && git commit -m "$(cat <<'EOF'
fix(website): set VitePress base to /sonatype-central-skills/ for GitHub Pages

线上 Pages 布局错乱的根因：VitePress 默认 base:'/'，但站点部署在
GitHub Pages 子路径 /sonatype-central-skills/ 下。所有 CSS/JS 资源
锚定域名根路径（/assets/xxx）→ 浏览器从 scagogogo.github.io 根域名取资源
→ 404 → 无样式裸 HTML。

curl 实测：
- https://scagogogo.github.io/assets/style.BOFw7W2O.css → 404
- https://scagogogo.github.io/sonatype-central-skills/assets/style.BOFw7W2O.css → 200

修复：显式设置 base: '/sonatype-central-skills/'。

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>
EOF
)"`

Expected:
  - Exit code: 0
  - 输出含新 commit hash

- [ ] **Step 2: push 到 main 触发 website.yml CI**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && git push origin main`

Expected:
  - Exit code: 0
  - 输出含 "main -> main"

- [ ] **Step 3: 等待 CI 构建部署完成 — 轮询 workflow 状态**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && for i in $(seq 1 30); do status=$(gh run list --workflow=website.yml --limit 1 --json status,conclusion --jq '.[0]'); echo "[$i] $status"; concl=$(echo "$status" | jq -r '.conclusion'); [ "$conc" != "null" ] && break; sleep 10; done; echo "--- 最终结论 ---"; gh run list --workflow=website.yml --limit 1 --json status,conclusion,databaseId --jq '.[0]'`

Expected:
  - Exit code: 0
  - conclusion 为 "success"

- [ ] **Step 4: 复验线上首页资源 200 — 确认布局修复**

Run: `echo "--- 线上首页 HTTP ---" && curl -s -o /dev/null -w "%{http_code}\n" https://scagogogo.github.io/sonatype-central-skills/ && echo "--- 线上 CSS(base 修复后应为子路径) ---" && CSS=$(curl -sL https://scagogogo.github.io/sonatype-central-skills/ | grep -oE 'href="[^"]*\.css' | head -1 | sed 's/href="//') && echo "CSS路径: $CSS" && curl -s -o /dev/null -w "%{http_code}\n" "https://scagogogo.github.io${CSS}" && echo "--- 线上 base 字段 ---" && curl -sL https://scagogogo.github.io/sonatype-central-skills/ | grep -oE '"base":"[^"]*"'`

Expected:
  - 首页 200
  - CSS 路径以 `/sonatype-central-skills/assets/` 开头，且该资源返回 200
  - `"base":"/sonatype-central-skills/"`

---

## Self-Review Results

| # | Check | Result | Action Taken |
|---|-------|--------|-------------|
| 1 | Header 含 Goal+Architecture+Tech Stack? | PASS | — |
| 2 | 每个 Task 标注 Depends on? | PASS | — |
| 3 | 每个 Task 列出精确文件路径? | PASS | config.ts:6-7 |
| 4 | 每个 Task 有 3-8 Steps? | PASS | T1=3,T2=2,T3=4 |
| 5 | 创建文件步骤含完整代码? | N/A | 仅 Modify |
| 6 | 修改步骤含替换后完整内容? | PASS | 含上下文 |
| 7 | 代码块大小合理? | PASS | 均 <10 行 |
| 8 | 无悬空引用? | PASS | — |
| 9 | 每个 Task 有验证命令? | PASS | grep/curl/gh |
| 10 | 需求全覆盖? | PASS | 含根因验证+线上复验 |
| 11 | 可独立验证? | PASS | — |
| 12 | 无 TBD/TODO? | PASS | — |
| 13 | 无抽象指令? | PASS | — |
| 14 | 跨 Task 一致性? | PASS | base 值统一 |
| 15 | 保存位置? | PASS | docs/superpowers/plans/ |

**Status:** ✅ ALL PASS

---

## Execution Selection

**Tasks:** 3
**Dependencies:** yes（T2 依赖 T1，T3 依赖 T1/T2）
**User Preference:** inline（单文件配置改动 + push + 在线验证，需根据 CI/curl 结果即时决策，不适合并行 subagent）
**Decision:** Inline
**Reasoning:** 这是一个根因明确、改动极小（4 行配置）但需要端到端在线验证的 bug fix。串行执行能根据每步 curl 结果判断是否继续，subagent 并行无意义且会干扰 push 顺序。采用 inline。

⏹️ Phase 4 Complete: Execution selected — Inline

---

## 根因总结（写给用户）

**线上确实基于 VitePress 部署**（HTML meta `<generator>VitePress v1.6.4</generator>` 可证）。布局错乱的原因不是"没部署 VitePress"，而是 **VitePress 的 `base` 路径没配置对**：

- 站点部署在 GitHub Pages 子路径 `https://scagogogo.github.io/sonatype-central-skills/`
- 但 config.ts 没设 `base`，VitePress 默认 `base: '/'`
- 于是构建产出的 HTML 引用的 CSS/JS 全是根路径：`/assets/style.xxx.css`
- 浏览器从 `scagogogo.github.io` 根域名取这些资源 → **404**（资源实际在子路径下）
- 没有 CSS → 页面变成无样式的裸 HTML，看起来"布局错乱"

实测：`curl https://scagogogo.github.io/assets/style.BOFw7W2O.css` 返回 404，而 `curl .../sonatype-central-skills/assets/style.BOFw7W2O.css` 返回 200。

修复只需在 config.ts 加一行 `base: '/sonatype-central-skills/'`，重新构建 push 即可。同时会把上次未 push 的 Mermaid 图、URL 笔误修正一并部署上线。
