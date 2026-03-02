下面给你一套「用 Google AI Studio 把一个典型 Web 应用从 0 做到可验收」的工作流：需求整理 → 原型&视觉 → 原型导出 HTML → 前后端脚手架 → 用例实现 → 测试验收，并穿插一套可直接照抄的 AI 工具集与提示词模板。

---

## 0) 总体架构先定死（避免后面返工）

**关键原则：Gemini API Key 不要直接放在浏览器里。** Google 官方明确提示：Web 客户端直连 Gemini API 只适合原型，生产环境要走更安全的方式（例如服务端调用，或迁移到 Firebase AI Logic）。 ([Google for Developers][1])

推荐两种常见落地方式：

1. **前端（Next.js/React）+ 服务端（Node/Go）代理调用 Gemini**

* 浏览器 → 你的后端（鉴权/限流/审计）→ Gemini API
* 适合你已有 Go（gin + Postgres）的体系：把“调用 LLM”做成一个内部服务或模块。

2. **前端直连（仅原型）**

* 用 Google Gen AI JS/TS SDK 直连，速度快，但要把 Key 暴露风险控制好（只在 demo/内测）。 ([Google for Developers][1])

---

## 1) 需求整理（PRD / 用户故事 / 验收标准）

**AI 工具建议**

* **Google AI Studio（Chat Prompt）**：把“需求对齐、边界条件、验收口径”做扎实。AI Studio 快速做对话式 prompt 迭代是它的强项。([Google AI for Developers][2])
* （可选）NotebookLM/文档 AI：把已有材料喂进去做摘要与冲突检查（如果你有一堆散乱需求很有用）。

**产出物**

* 一页 PRD：目标、非目标、用户画像、核心流程、异常流、数据字典（最小化）。
* 用户故事（User Story）+ 验收标准（AC）
* 风险清单：安全/隐私/成本/延迟

**可复制提示词（直接丢 AI Studio）**

* “把下面的想法整理成：目标/非目标/用户故事（含AC）/核心流程（含异常流）/数据模型草案/接口草案（REST）/风险与假设。想法如下：…”
* “针对这些用户故事，补齐可测试的验收标准（Given-When-Then），并列出边界条件与反例。”

---

## 2) 原型与视觉（低保真 → 高保真 → 设计规范）

**AI 工具建议（设计侧）**

* Figma + AI 插件（用于信息架构、文案、组件变体生成等）
* 你如果要“从设计到前端代码/HTML”，业界常见路线是 **Figma →（Anima 等）→ HTML/React**：Anima 明确提供从 Figma 导出 HTML 的路径。([Anima][3])

  > 注意：这类导出代码往往“可跑但不一定生产级”，适合当作起点（样式与结构参考），后续仍建议工程化重构。

**产出物**

* 关键页面低保真（流程优先）
* 设计系统：字体/间距/颜色/组件状态（hover/disabled/error）
* 高保真页面 + 交互说明（表单校验、空状态、加载态、错误态）

---

## 3) 原型导出 HTML / 前端落地（从“能看”到“能用”）

这里建议把“视觉稿 → 组件化”拆成两步：

1. **从 Figma 导出（或半自动生成）基础结构**

* 用 Anima 之类导出 HTML/React 作为骨架起点。([Anima][3])

2. **用 AI Studio 帮你“工程化重构”**

* 目标：语义化 HTML、可复用组件、Tailwind/shadcn 统一样式、可测试性（data-testid）、可访问性（aria）。

你可以把导出的页面代码片段贴进 AI Studio，让它按你的栈重构：

* “把这段导出的 HTML 重构成 React 组件（TypeScript），拆分为 Header/Form/Table/Modal 子组件；使用 Tailwind；加入 a11y；保留像素级布局；输出文件结构建议。”

---

## 4) 前后端代码脚手架（AI Studio Build mode + 本地 IDE）

### 4.1 用 Google AI Studio 的 Build mode 快速起一个可运行全栈原型

Google AI Studio 已经支持 **Build mode**，主打用自然语言快速“vibe code”并支持**全栈运行时**（服务端逻辑、secrets 管理、npm 包）。([Google AI for Developers][4])
适合你在 1~2 小时内把“主流程跑通”。

建议你在 Build mode 里要求它生成：

* 前端路由结构
* 服务端接口（最少：auth / CRUD / LLM 调用）
* secrets 不写死（走环境变量/secret 管理）
* 最小可用数据存储（先 sqlite/mock，后续替换 Postgres）

### 4.2 再把原型“迁回你的工程栈”

你如果生产要用 gin + Postgres：

* Build mode 产出当“交互与流程参考实现”
* 然后把：

  * API 设计（路由、请求/响应 schema）
  * 前端页面结构与状态机
  * LLM 调用策略（prompt、工具调用结构、错误重试）
    迁移到你的正式 repo（Go + Next.js 等）。

---

## 5) 用例实现（把“AI 生成代码”变成“可维护功能”）

**建议用“用例驱动”来喂 AI：**
每个用例固定四段输入：

1. User Story + AC
2. API contract（OpenAPI/JSON Schema）
3. 数据模型（SQL migration 草案）
4. UI 状态机（loading/empty/error/success）

然后让 AI 输出：

* 后端 handler/service/repo（含错误码映射）
* 前端页面 + hooks
* 最少一组集成测试样例

**强烈建议：把 LLM 调用封装成一个“LLM Adapter”**

* 统一：超时、重试、降级、成本统计、日志脱敏
* 统一：prompt 模板版本化（例如 `prompts/v1/*.md`）

---

## 6) 测试与验收（把 AI 变成“测试生成器 + 审查员”）

**测试层级建议**

* 单测：核心业务函数、数据校验、权限判断
* 接口测试：Postman/Newman 或超轻量自写
* E2E：Playwright（推荐）跑关键路径（登录→下单/创建→列表→详情→删除）
* 回归清单：由 AC 自动生成 checklist

**AI Studio 里非常实用的两类提示词**

* 生成测试：

  * “根据这些用户故事与 AC，生成 Playwright E2E 用例（含稳定选择器策略），并给出测试数据准备与清理步骤。”
* 做审查（非常重要）：

  * “下面是我实现的接口与前端逻辑，逐条对照 AC 检查是否覆盖；列出未覆盖场景、潜在 bug、以及必须补的监控/日志点。”

---

## 7) 推荐的一套“AI 工具集”清单（按阶段配齐）

**需求/文档**

* Google AI Studio（Chat Prompt）：需求拆解、AC、边界条件、接口草案 ([Google AI for Developers][2])
* （可选）NotebookLM：资料多时做归纳与冲突检测 ([blog.google][5])（说明它覆盖在 Google 生态里，但你仍需要自己验证输出）

**原型/视觉**

* Figma +（Anima 等）导出 HTML/React 骨架 ([Anima][3])

**代码原型**

* Google AI Studio Build mode：快速起全栈原型、支持 server-side 逻辑与 secrets/npm ([Google AI for Developers][4])

**Web 端集成**

* Google Gen AI SDK（JS/TS）：原型期直连；生产迁移到更安全方案（Firebase AI Logic 或服务端） ([Google for Developers][1])

**测试/验收**

* Playwright + AI 生成 E2E 用例
* 契约（OpenAPI）+ AI 生成接口测试与 mock

---

## 8) 一页“落地操作顺序”（你照这个跑就行）

1. 在 AI Studio（Chat Prompt）把 PRD + 用户故事 + AC 定稿（导出到你的文档）。([Google AI for Developers][2])
2. Figma 画低保真 → 高保真，补齐空/错/加载态。
3. 用 Anima 导出 HTML/React 当骨架。([Anima][3])
4. 用 AI Studio（Build mode）起一个“端到端可跑的全栈原型”，把主流程跑通。([Google AI for Developers][4])
5. 回到你的正式栈：后端（gin+Postgres）实现接口与权限；前端（Next/React）重构组件与状态。
6. 用 AI 生成并补齐：OpenAPI、Playwright E2E、回归 checklist。
7. 最后按 AC 做验收，并把日志/监控/成本统计补上。

---

如果你告诉我：你更偏向 **Next.js 全栈** 还是 **前端 Next.js + 后端 gin**，以及你应用类型（管理后台 / 内容站 / SaaS 表单流 / Chat 类工具），我可以直接给你一份“可复制到 AI Studio 的项目总提示词（Project Spec Prompt）”，让 Build mode 一次性产出更贴合你栈的目录结构、接口、页面与测试骨架。

[1]: https://developers.google.com/learn/pathways/solution-ai-gemini-getting-started-web?utm_source=chatgpt.com "Getting started with the Gemini API and Web apps"
[2]: https://ai.google.dev/gemini-api/docs/ai-studio-quickstart?utm_source=chatgpt.com "Google AI Studio quickstart | Gemini API"
[3]: https://www.animaapp.com/blog/design-to-code/how-to-export-figma-to-html/?utm_source=chatgpt.com "Figma to HTML: How to export a Figma design into HTML"
[4]: https://ai.google.dev/gemini-api/docs/aistudio-build-mode?utm_source=chatgpt.com "Build apps in Google AI Studio | Gemini API"
[5]: https://blog.google/innovation-and-ai/models-and-research/gemini-models/gemini-3-1-pro/?utm_source=chatgpt.com "Gemini 3.1 Pro: A smarter model for your most complex tasks"
