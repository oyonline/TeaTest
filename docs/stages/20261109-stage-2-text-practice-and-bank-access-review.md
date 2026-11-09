# TeaTest 阶段 2：Claude 跨模型评审回执

> 本件为首审历史记录。用户随后授权的复审已返回 APPROVE；最新结论见同目录 `20261109-stage-2-text-practice-and-bank-access-rereview.md`。下文“未复审”均指首审收口当时。

日期：2026-11-09。评审对象是阶段卡，业务功能尚未施工。

## 结论与版本

- 初审：**REQUEST_CHANGES，P0=0、P1=1、P2=2**；调用成功不等于方案通过。
- 被审原稿：`/Users/linshen/Cursor/TeaTest/docs/stages/20261109-stage-2-text-practice-and-bank-access.md`，SHA256 `04c29fc6cfa37a050c4b05a84242d3b5d36aa2253ede9c75996a116ecf7837b1`；完整原稿随评审输入保存在下方原始制品目录。
- 修订稿：同路径，SHA256 `e2a78825d392f3f1f6f5e7a0cda5d3dfd8dcf5c284cf5150492070131a160b79`。已按下列说明修订，**未追加复审，不能称 Claude 已批准修订稿**。
- 施工状态：待用户批准阶段 2。未改业务代码、未执行数据库迁移、未提交推送或部署。

## 发现处理

1. P1：采纳。明确两入口复用同一答题/结果组件，均支持五种题型，分类不限制题型；验收改为两入口各一份五题型混合卷，包含原客观题回归。
2. P2：采纳入口统计口径说明：列表、统计与进行中提示都按授权与入口分类交集过滤，入口只能缩小范围。原评审“现有首页有通用进行中跳转”的具体触发描述未被当前 Welcome.vue 支持，本记录不将其宣称为已复现故障。
3. P2：补充 utf8mb4 与 emoji 边界验收；确认当前新题仅沿用 CSV 导入，本波不新增单题编辑器。没有为 P2 新开验证或评审轮次。
4. 另补明确既有“提交后对照”要求：最后一题也允许查看参考答案再进入结果，避免复用旧自动跳转时遗漏简答对照。

## 调用凭证

- 入口：`xmodel-call L1`，requested model `opus`，期望标签 `claude-opus-5-5`；回执 reported model `claude-opus-5-5`，status `reported_match`。标签一致不等于独立证明底层供应商身份。
- requested effort `high`，通过单次环境变量 `CLAUDE_CODE_EFFORT_LEVEL=high` 设置；回执未报告 actual effort，记为 unknown，不冒称回读确认。
- effort 设置来源：[Claude 官方模型配置说明](https://code.claude.com/docs/en/model-config#adjust-effort-level)。
- 限制：模型工具全部关闭、空 MCP、MULTI_AGENT=false、TOOL_BUDGET=0；transcript tool_use=0、tool_result=0；单轮评审。
- watchdog：declared_timeout_s=600、stall_threshold_s=300、poll_interval_s=30。
- CLI：2.1.319 (Claude Code)；session `91f9cb66-a940-419d-9fc1-0cf20937d5e8`；exit_code=0，terminal.ok=true，cleanup_ok=true，num_turns=1。
- 模型返回标价成本：$0.3569298；不是另行核验的实际扣款。
- 原始回执：`/Users/linshen/.dsh/artifacts/cross-model/2026-11-09/L1/183112-4955-teatest-stage2-review-N2JXarR7/receipt.json`。
- 原始输出：`/Users/linshen/.dsh/artifacts/cross-model/2026-11-09/L1/183112-4955-teatest-stage2-review-N2JXarR7/out.txt`。原始证据未覆盖。

## Claude 首审原文（发现部分）

VERDICT=REQUEST_CHANGES

P0=0　P1=1　P2=2

### 1．[P1] 两个入口都可能遇到全部五种题型，但卡里没说两个答题页各自支持哪些题型

- **卡内位置**：§2「新增 `fill_blank`、`short_answer`」只说新增两种题型。§2「旧入口展示竞赛/考级，新入口展示通用题库」只按分类分流。§4「前端…文字作答/结果页」没写这是给哪个入口用的。§3 回归只覆盖「旧单选/多选/判断各一条完整答题流程」。
- **真实触发**：导入时由管理员选目标题库，卡里没有限制题型和分类的搭配。管理员可以把填空 CSV 导入竞赛库，也可以往通用库导入单选题。也可以先导入，之后再改题库分类。
- **影响**：
  - 当前 `exam_service.go:490-493` 会把未知题型都当成 `single_choice` 返回。如果只改了新页面，旧入口会拿到一道没有选项的填空题，用户答不了。
  - 所有题都答完才会结算（`finalizeExamIfComplete`），所以这场考试永远无法完成。
  - 反过来，`/practice` 里出现单选题也是同样的问题。
  - 验收表测不到这种情况。
- **最小修正**：在卡里写明「两个入口共用同一套题目渲染，五种题型都支持」。验收表加一行：旧入口完成一份「客观题＋填空＋简答」混合卷，`/practice` 完成一份含单选的卷。这样比按分类限制题型更省事，因为限制题型还要处理导入后再改分类的情况。

### 2．[P2] 统计和「继续答题」提示没有按入口区分，可能串入口

- **卡内位置**：§2「两入口按分类和授权交集展示」只约束了题库列表。现有 `GetExamStats(userID)` 会统计这个用户的所有 `in_progress`、`completed` 记录，并返回 `has_in_progress`（`exam_service.go:411-441`）。卡里只写了「相关统计」按授权过滤，没写按入口过滤。
- **真实触发**：同一个用户同时有竞赛库和通用库的授权，并且在 `/practice` 有一场进行中的练习。他从旧入口登录，欢迎页会显示「有进行中考试」，点进去会被带到练习卷；题目总数里也混入了通用库的题。
- **影响**：违背验收「`/practice` 刷新及返回不串入口」。
- **最小修正**：卡里加一句：「统计和进行中提示先按服务端授权过滤，再按当前入口的分类筛选展示。入口参数只能缩小范围，不能放宽权限。」

### 3．[P2] 「最多 4000 个 Unicode 字符、不截断」只验证了 CSV 导入，没覆盖字符集和其他写入路径

- **卡内位置**：§2「最多 4000 个 Unicode 字符」。§3「4000 字中文/混合文本可导入、作答、刷新」。补充说明里「新类型参考答案不 uppercase」只写在 CSV 那一条下。
- **真实触发**：
  - 如果表或连接用的是 utf8mb3，emoji 这类四字节字符会报错，非严格模式下会被静默截断。纯中文测试发现不了。
  - 如果后台有单题新增或编辑的入口，并且沿用旧的大写化和 A–E 校验，文字答案在编辑后会被改写。
- **最小修正**：验收加两项：
  - 确认受影响的列使用 utf8mb4，测试文本里放一个 emoji。
  - 明确后台单题保存路径（如果存在）对文字题型同样不改写、不截断；如果新题型只能导入，就在卡里写明。

### 卡外改造

未发现卡外改造，方案改动范围与 §4 和「不做」清单一致。

