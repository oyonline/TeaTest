# TeaTest 阶段 2 Claude 修订稿复审回执

2026-11-09，用户明确要求再复审一次修订后的阶段卡。本轮为方案复审，不构成施工或部署授权。

## 当前结论

**VERDICT=APPROVE，P0=0、P1=0、P2=0。** 首审的 1 个 P1、2 个 P2 均被复审判断为已消除，无新成立问题，未发现卡外改造。

- 被审卡：`/Users/linshen/Cursor/TeaTest/docs/stages/20261109-stage-2-text-practice-and-bank-access.md`。
- 被审 SHA256：`e2a78825d392f3f1f6f5e7a0cda5d3dfd8dcf5c284cf5150492070131a160b79`。
- 当前文件 SHA256：`16dcd05bfc5a44bca10996660ce03721981ac015d090b073a7f096893e090cdb`。复审后只更新状态行；已机械校验，其他卡文与被审稿一致。
- 状态：方案待用户批准施工；业务代码未改，数据库未迁移，未提交推送或部署。
- 首审历史：同目录 `20261109-stage-2-text-practice-and-bank-access-review.md`，原始证据保留。

## 模型与执行回执

- requested model `opus`，期望标签 `claude-opus-5-5`；CLI reported model `claude-opus-5-5`，status `reported_match`。标签回读不等于独立证明底层身份。
- requested effort high，通过 `CLAUDE_CODE_EFFORT_LEVEL=high` 设置；actual effort 未报告，记 unknown。
- 工具全部禁用、空 MCP、TOOL_BUDGET=0、MULTI_AGENT=false；transcript tool_use=0、tool_result=0。
- session `e8ec7003-999a-42ee-81d1-a05e814e71f4`；exit_code=0，terminal.ok=true，cleanup_ok=true，num_turns=1。
- watchdog：declared_timeout_s=600、stall_threshold_s=300、poll_interval_s=30。
- CLI：2.1.319 (Claude Code)；模型返回标价成本 $0.2989964（不是另行核对的实付金额）。
- 原始回执：`/Users/linshen/.dsh/artifacts/cross-model/2026-11-09/L1/184214-9247-teatest-stage2-rereview-EWxWCINZ/receipt.json`；原始输出：`/Users/linshen/.dsh/artifacts/cross-model/2026-11-09/L1/184214-9247-teatest-stage2-rereview-EWxWCINZ/out.txt`。

## Claude 复审原文

VERDICT=APPROVE

P0=0　P1=0　P2=0

## 首审问题消除情况

| 首审问题 | 修订稿中的对应内容 | 结论 |
|---|---|---|
| [P1] 两个入口支持的题型不明确，混合题型的考试无法完成 | §2「两个入口共用同一答题与结果组件，均支持单选、多选、判断、填空、简答，题型不受题库分类限制」；§3「两个入口各完成一份五题型混合卷（覆盖原客观题）」 | 已消除 [原文·文档] |
| [P2] 统计和进行中提示可能串入口 | §2「列表、统计及进行中提示按"授权∩入口分类"过滤；入口参数只能缩小范围」；§3「统计/进行中提示仅当前入口」 | 已消除 [原文·文档] |
| [P2] utf8mb4 和其他写入路径未覆盖 | §3「连接/答案列为 utf8mb4；含 emoji 的 4000 字混合文本…」；§2「新题沿用 CSV 导入，不新增单题编辑器」 | 已消除 [原文·文档] |

## 本轮复核的检查项

以下检查项均未发现会导致错误交付或验收歧义的问题：

- **授权隔离**：覆盖题库列表、题目、续答、结果、统计和作答，按服务端身份校验。验收中的「拒绝请求写入 0 行」能覆盖越权开新考试和越权提交。
- **存量授权**：有持久完成标记，与初始化数据同事务。验收已写明重启不恢复已撤销的权限，也不放行后来新增的题库。
- **撤销与历史未分类考试**：未分类考试不再走全库读取和提交，也不计入当前统计，没有留下旁路。
- **文字答案**：
    - 字段从 varchar(10) 扩容。
    - 文字题型不经过 `normalizeAnswer`，也不做 uppercase。
    - 只去首尾空白、统一换行。
    - 4001 字和纯空白一律拒绝，不截断。
- **未评分**：用可空的 is_correct 或明确状态表达，不计入错题数和正确率分母。2 填空＋2 简答的验收数字前后自洽。
- **新入口回流**：登录失败、JWT 过期、刷新、从结果页返回都保留入口；`request.js` 的硬跳 `/` 已在补充说明中纳入修复。

## 卡外改造

未发现卡外改造，§4 的改动范围与「不做」清单一致。

