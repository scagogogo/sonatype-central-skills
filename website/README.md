# Sonatype Central SDK 官网

基于 [VitePress](https://vitepress.dev) 构建的官方文档站点，部署在 GitHub Pages。

## 本地开发

```bash
# 安装依赖
npm install

# 启动开发服务器（热更新）
npm run dev

# 构建生产版本
npm run build

# 本地预览构建产物
npm run preview
```

构建产物输出到 `.vitepress/dist`。

## 目录结构

```
website/
├── .vitepress/
│   └── config.ts          # VitePress 配置（导航、侧边栏、搜索）
├── public/
│   └── logo.svg           # 站点 logo
├── index.md               # 首页（含 AI Agent 一键接入提示词）
├── guide/                 # 使用指南
│   ├── introduction.md
│   ├── quick-start.md
│   ├── what-it-solves.md
│   ├── how-it-works.md
│   ├── search.md
│   ├── download.md
│   ├── publish.md
│   ├── advanced-search.md
│   ├── batch-iterator.md
│   ├── cache-retry.md
│   ├── deprecated.md
│   └── faq.md
├── ai-agent/              # AI Agent 接入指南
│   ├── index.md
│   ├── claude-code.md
│   └── codex.md
├── api/                   # API 参考
│   ├── search.md
│   ├── download.md
│   ├── publisher.md
│   └── response-types.md
├── package.json
└── package-lock.json
```

## 部署

通过 GitHub Actions 自动部署（见 `.github/workflows/website.yml`）。

### 一次性配置（仓库管理员）

部署前需在 GitHub 仓库设置：

> **Settings → Pages → Build and deployment → Source** 选择 **"GitHub Actions"**

配置后，每次推送到 `main` 分支且改动 `website/` 目录时，工作流会自动构建并部署。

部署 URL 通常为：`https://<owner>.github.io/<repo>/`（本项目即 `https://scagogogo.github.io/sonatype-central-skills/`）。

### 手动触发

也可在 GitHub 仓库的 **Actions → Deploy Website → Run workflow** 手动触发部署。

## 写作约定

- 所有面向用户的文档使用**简体中文**
- 代码块标注语言（`go`、`bash`、`markdown`）以获得语法高亮
- VitePress 默认为每个代码块提供复制按钮，无需额外标记
- 侧边栏分组在 `.vitepress/config.ts` 中维护，新增页面需同步更新导航配置
