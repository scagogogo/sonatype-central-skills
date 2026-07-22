import { withMermaid } from 'vitepress-plugin-mermaid'

// https://vitepress.dev/reference/site-config
// withMermaid 直接接收 VitePress 配置对象（含 mermaid 字段），
// 无需 defineConfig 包装 —— 包装会干扰插件对 markdown.config 的注入。
export default withMermaid({
  // 站点部署在 GitHub Pages 子路径 /sonatype-central-skills/ 下，
  // 必须显式声明 base，否则所有 CSS/JS/图片资源会锚定域名根路径而 404，
  // 导致布局错乱（无样式裸 HTML）。VitePress 默认 base 为 '/'，仅适用于根域名部署。
  base: '/sonatype-central-skills/',
  lang: 'zh-CN',
  title: 'Sonatype Central SDK',
  description: '一个全面、类型安全的 Go SDK，用于 Sonatype Central Repository API — 搜索、下载、发布 Maven 制品',
  lastUpdated: true,
  cleanUrls: true,

  head: [
    ['meta', { name: 'theme-color', content: '#3c8cff' }],
    ['meta', { name: 'author', content: 'scagogogo' }],
    ['meta', { property: 'og:title', content: 'Sonatype Central SDK' }],
    ['meta', { property: 'og:description', content: 'Go SDK for the Sonatype Central Repository API' }],
    ['meta', { property: 'og:type', content: 'website' }],
  ],

  themeConfig: {
    // 站点级别的社交链接
    socialLinks: [
      { icon: 'github', link: 'https://github.com/scagogogo/sonatype-central-skills' },
    ],

    search: {
      provider: 'local',
      options: {
        translations: {
          button: {
            buttonText: '搜索文档',
            buttonAriaLabel: '搜索文档',
          },
          modal: {
            noResultsText: '无法找到相关结果',
            resetButtonTitle: '清除查询条件',
            footer: {
              selectText: '选择',
              navigateText: '切换',
              closeText: '关闭',
            },
          },
        },
      },
    },

    // 顶部导航
    nav: [
      { text: '指南', link: '/guide/introduction', activeMatch: '/guide/' },
      { text: 'AI Agent 接入', link: '/ai-agent/', activeMatch: '/ai-agent/' },
      { text: 'API', link: '/api/search', activeMatch: '/api/' },
      {
        text: 'GitHub',
        link: 'https://github.com/scagogogo/sonatype-central-skills',
      },
    ],

    // 侧边栏配置：根据路径分组
    sidebar: {
      '/guide/': [
        {
          text: '开始',
          collapsed: false,
          items: [
            { text: '介绍', link: '/guide/introduction' },
            { text: '快速开始', link: '/guide/quick-start' },
            { text: '它解决了什么问题', link: '/guide/what-it-solves' },
            { text: '工作原理', link: '/guide/how-it-works' },
          ],
        },
        {
          text: '核心能力',
          collapsed: false,
          items: [
            { text: '搜索', link: '/guide/search' },
            { text: '下载', link: '/guide/download' },
            { text: '发布到 Maven Central', link: '/guide/publish' },
            { text: '高级搜索', link: '/guide/advanced-search' },
            { text: '批量操作与迭代器', link: '/guide/batch-iterator' },
            { text: '命令行工具（CLI）', link: '/guide/cli' },
          ],
        },
        {
          text: '更多',
          collapsed: true,
          items: [
            { text: '缓存与重试', link: '/guide/cache-retry' },
            { text: '已弃用的 API', link: '/guide/deprecated' },
            { text: 'FAQ', link: '/guide/faq' },
          ],
        },
      ],
      '/ai-agent/': [
        {
          text: 'AI Agent 接入',
          collapsed: false,
          items: [
            { text: '总览', link: '/ai-agent/' },
            { text: 'Claude Code 接入', link: '/ai-agent/claude-code' },
            { text: 'Codex 接入', link: '/ai-agent/codex' },
          ],
        },
      ],
      '/api/': [
        {
          text: 'API 参考',
          collapsed: false,
          items: [
            { text: '搜索 API', link: '/api/search' },
            { text: '下载 API', link: '/api/download' },
            { text: '发布 API', link: '/api/publisher' },
            { text: '响应类型', link: '/api/response-types' },
          ],
        },
      ],
    },

    // 页脚
    footer: {
      message: '基于 MIT 许可证发布',
      copyright: 'Copyright © 2024-present scagogogo',
    },

    // 上一页/下一页导航文本
    docFooter: {
      prev: '上一页',
      next: '下一页',
    },

    // 大纲配置
    outline: {
      level: [2, 3],
      label: '本页导航',
    },

    // 编辑此页链接
    editLink: {
      pattern: 'https://github.com/scagogogo/sonatype-central-skills/edit/main/website/:path',
      text: '在 GitHub 上编辑此页',
    },

    // 最后更新时间文本
    lastUpdated: {
      text: '最后更新于',
    },

    // 侧边栏组标签
    sidebarMenuLabel: '菜单',
    returnToTopLabel: '回到顶部',
  },
})
