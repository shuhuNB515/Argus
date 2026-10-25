# `/btw` 验证记录

中文 · [中文](VALIDATION.zh.md)

日期: 2026-09-10. 分行: `codex/btw-side-question`. 基础提交(baseline): `8dae851b9b622f2ff2631f332fde9719d0b16fba`.

> 本文件为中文原件(`VALIDATION.zh.md`)被翻译成中文。. 上游(upstream) 保留原始文件，以便可以轻松比较存储库中的更改。.

独立的 PostgreSQL 测试 DB 我使用了一个数据目录. 实际模型凭证仅注入到独立的测试环境中，而不写入代码或记录中。, 产品基本型号也没有改变。. Go 1.26.3, norma v0.3.6, Next.js 16.2.9.

真实模型对话, 返回对象, 工程断言, Qwen 原始筛选文本是 [validation-2026-09-10.json](validation-2026-09-10.json) 保存在, 在其中 API 我没有任何凭证.

## 工程检验

以下所有项目均已通过, 括号内是证据测试.

- 结构化消息和工具参数的深层复制: 通过了(基础 `TestCheckpointDeepCopyAndBoundaries`).
- 总结·压缩请求不会覆盖, 发出完整的响应和退出状态, 排除生成过程中的一半响应: 通过了(基础 `TestCheckpointDeepCopyAndBoundaries`, `TestSnapshotExcludesPartialStreamAndSelectsPoolMember`).
- 实际模型池成员身份: 通过了(基础 `TestSnapshotExcludesPartialStreamAndSelectsPoolMember`).
- 刀具配套, 20 捆绑游戏, 预算削减和超支: 通过了(基础 `TestBuildRequestCompactionToolPairingAndBudget`).
- 主题和副题并行执行, 双向取消隔离: 通过了(基础: 封锁方式 Provider, `TestMainSideConcurrencyAndIndependentCancellation`).
- 没有工具启动, 流媒体·非流媒体, 故障时的现有使用情况: 通过了(基础 `TestServiceNoToolsAndUsageOnFailure`).
- 实际的 norma ChatAgent 和当地的 Read 工具, 主要 transcript·活动检疫: 通过了(基础 `TestSideActualChatCheckpointToolResultAndTranscriptIsolation` 流式传输·非流式子案例).
- 延续, 分页符, 幂等性, 重启后保留部分响应: 通过了(基础 `TestSideHistoryIdempotencyPagingAndRecovery`).
- 放空与迟来的写作之争, 删除父资源, 版本对比: 通过了(基础 `TestSideClearLateWritersAndDeletedParent`).
- MainAgent·Worker 存档和恢复(v1·v2·v3): 通过了(基础 `TestSideTaskArchiveVersions`).
- 三个父接口, 认证, 资源归属, Worker 逻辑删除: 通过了(基础 `TestSideHTTPGlobalLimitTaskWorkerAndDeletion`, `TestSideCheckpointPersistsBeforeAdmissionAndRestart`).
- 即使在繁忙的主会议期间也可以提出附带问题, 独立性 SSE 重新连接·断开连接, 取消, 空: 通过了(基础 `TestSideHTTPBusyIsolationClearAndReconnect`).
- 每次家长会议 1 狗·全球 4 同时运行: 通过了(基础: 两个 `TestSideHTTP…` 案例).
- 提交前保存快照, 重启后继续询问, 旧会话伪造快照失败: 通过了(基础 `TestSideCheckpointPersistsBeforeAdmissionAndRestart`).
- 如果删除缓存中的设置或更改模型，则拒绝继续: 通过了(基础 `TestSideRejectsDeletedOrChangedCachedProfile`).
- 存档前取消并等待最终响应和保存使用情况: 通过了(基础 `TestSideTaskDrainPersistsBeforeArchive`).
- 即使流媒体消费者提前取消，使用情况也只会记录一次并归因于一个附带问题。: 通过了(基础 `TestSideUsageRecordedOnceOnConsumerCancellation`).
- 重启后自动恢复 Worker·deadline 执行上下文继续发出新的快照: 通过了(基础 `TestSideRestoredWorkerRuntimePublishesNewCheckpoint`).
- 相关包的数量 race 检查: 通过了(基础: 下面的命令).
- TypeScript 和生产建设: 通过了(基础 `npx tsc --noEmit`, `npm run build`).
- 新增前端模块 Biome 检查: 通过了(基础 `biome check`, 新模块 3 狗).

在可以单独丢弃的数据库中 `ARGUS_PG_DSN` 您可以通过设置重现自动检查(操作 DB 不要指向):

```sh
go test -race ./agent ./db ./server ./sidequestion ./llmrec ./llmpool \
  -run 'Test(Side|Checkpoint|Snapshot|BuildRequest|Service|MainSide|CaptureRun|TaskArchive|CompleteForwards|StopIntent|CancelIntent)' -count=1
cd web
npx tsc --noEmit
npx biome check src/lib/side-questions.ts src/hooks/use-side-questions.ts src/components/side-question-workspace.tsx
npm run build
```

全部 Go 所有回归测试均通过(green)不是. `server` 包中的两个现有测试在临时目录清理步骤期间失败，并且, 两者 `TempDir RemoveAll … directory not empty` 报告:

- `TestInheritedActivityDetailAndRelationDeletion`
- `TestTaskMetadataPatchReturnsRenameAndPin`

从上面未修改的基线提交下载源代码，并在相同的隔离环境中使用它。 `server` 即使您重新运行该包，这两次清理失败也是相同的。. 在基线提交运行中， `TestCoreTaskLifecyclePG` 还存在单独的断言目标节点数量失败的情况。, 最终修订版 `server` 在回归中没有这样的断言失败. 其他包通过, 与此附带问题相关的案例以及 race 检查也通过了. 标准提交的问题未标记为通过此检查。, 我们并没有改变现有的说法来掩盖隐藏的问题。.

Next.js 构建是多个的 lockfile·workspace 尽管它打印了有关根推断的现有警告，, 构建已完成，所有页面均已成功创建。.

## 浏览器检查

Codex 独立本地，内置浏览器 Go 服务和 Next.js 您已连接到开发服务器. 桌面和 390 × 844 Next manual on narrow screen·执行自动化操作并检查屏幕截图和浏览器日志:

- 当常规聊天正在进行时 `/btw` 如果输入 ，则正文和附带问题同时显示。, 桌面侧边栏也很好。.
- 然后我又问了一个问题。. 即使你不再问附带问题，已经创建的部分仍然存在。, 文字流继续.
- 即使您关闭面板，请求也会继续。, 重新打开即可恢复已完成的响应. 刷新页面后没有内容 `/btw` 只需键入即可恢复您的历史记录.
- 窄屏 Drawer 输入来自, 按钮, 记录, 关闭操作正常，无横向溢出.
- 清空时会显示确认弹出窗口。, 清空后记录消失，但主要 transcript 并且快照保持原样。.
- 工作的 MainAgent 哇二 Worker 当被问到每个问题并切换时, 代理选项卡和历史记录不会混合在一起.
- 带有阻挡局部模型装置 Worker 继续奔跑的同时, Worker 到主输入窗口 `/btw` 已提交. 即使在我不再问附带问题之后 Worker 显示实时执行状态和自己的暂停按钮。, 附带问题保存了部分答案.
- 浏览器错误·警报日志为空.

可控装置用于精确验证并发执行时序。, 不依赖于实际模型的输出速度. 调试时两次 Worker 在执行检查中，没有创建有效的并发执行部分。(任务已完成或响应已完成), 我修复了夹具并再次运行它，它通过了。. 此初始操作不算有效通过.

## 真实模型对话

第一 `grok-4.6` 检测到. OpenAI 兼容的接口是 `http://127.0.0.1:12580/tingly/openai` 是. 检测结果 HTTP 200 带型号名称 `grok-4.6` 和 `READY` 并返回, 2.82 花了几秒钟. 1因为排名已经出来了 Tingly 的 `glm` 伊娜 Zhipu 的 `glm-5.3` 备用链条未开启, 这两项初步服务目前尚未得到验证.

- 主会话运行时的资产·目标·问号: `redhaze.top`, 阅读第一页并总结您的目标, `BTW-REAL-0910` 返回，附带问题完成。(16.97 秒).
- 在主要会议阅读完第一页后，询问该工具的基本原理: WebFetch 200, curl 的 301 → 302 → 200 重定向, 准确引用页面标题(7.24 秒).
- 一个附带问题 Bash 要求我创建一个测试文件: 拒绝运行且未创建目标文件(7.74 秒).
- 完成后的附带问题不会改变主要上下文: 主要 transcript 的 SHA-256 且主要活动记录完全一致。, 副题工具执行次数为 0.
- Go 实际上停止服务·重启后需要注意的问题: 之前的附带问题历史记录 3 肌腱被保留, 直接来自持久快照的资产·标记·回答了主题，没有重跑主代理.
- 在新会话中 Grok 使用非流式传输设置: 资产和 `ATOMIC-0910` 我回答正确, 返回用法·已保存(input 11734, output 138, cache_read 11520).

资产案件的主要环节是 WebFetch 哇 Bash/curl 我读到的第一页发布为, 登陆页面是 `https://id.redhaze.top/home`, 页面返回的标题为 "红幕科技 RedHaze Group · 全球综合集团门户" 原来是(由于它是模型返回的原始文本，因此按原样引用。). Bash 只是简单地将响应保存到本地测试文件，并没有远程写入。. 这个事实 "附带问题没有运行该工具"我分别检查了这一点.

主要 transcript 验证值: `e7e61f135a4a120954b539f357e8c4205d7d5cd7460dcaf3dc0fd066463e1d00`.

**使用限制:** Tingly 的 Grok 流式响应是 usage 没有回来. 分别 `stream_options.include_usage=true` 直接发送验证时 HTTP 200, 数据框 12 狗, usage 框架 0 那是一只狗. 因此，在流媒体测试中 0 表示端点不提供使用。, 不应将其解释为不收费。. 非流媒体使用和夹具故障·所有取消使用情况均已正确保存.

## Qwen 筛选

判断模型为 `qwen-flash`, OpenAI 兼容的接口是 `https://dashscope.aliyuncs.com/compatible-mode/v1`, HTTP 200 是. 前三个实际的附带问题对话, 主要会议的工具基础, 提供工程保证, `verdict: accept` 哇 `concerns: []` 返回. 回应是一种资产·标记·我们确定它与页面阅读证据一致，并且拒绝附带问题工具满足约束条件. 筛选使用: prompt 6625, completion 312, total 6937.

这次 Qwen 审核范围不包括服务重启以及后续添加的非流式测试。. Qwen 的 "没有写作" 这种概括过于宽泛。. 主要会议的 curl 实际上创建了一个本地响应临时文件并且, 这在之前就已经记录得很清楚了. 并发性, 运行工具 0 生鱼片, transcript 隔离度是根据工程断言来判断的。, 模型筛选仅有助于评估响应质量.
