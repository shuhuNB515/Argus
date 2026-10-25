#!/usr/bin/env python3
"""静态导出网页 UI(`web/out`)的 HTML 中文(汉字)检查是否有泄漏。.

该分叉的主要成就和差异化是 "向用户显示的屏幕上没有中文。"是
(积压 B4 建立者 `out` HTML 汉字 0 属性). 如果您只将该属性留给人工验证，, 贡献者
即使存在诸如硬编码中文字符串或缺少翻译之类的回归，下一个人也可以看到它。
在被发现之前一直被悄悄破坏. 该脚本可防止合并门中的回归.

检查方法
---------
`web/out` 以下全部 `.html` 文件的 **原始内容**(内联标记和渲染标记
脚本·React Server Components 飞行数据)如果有一个汉字，
退出代码 1 它结束于. 查看原始内容的原因是, 用户可见的文本是
飞行数据(JSON 字体字符串)精选于 HTML 因为它与. 仅标签
如果你把它剥掉，你就会错过该路径中的泄漏.

超出检验范围 (诚实的限制)
-----------------------
该门只能看到静态导出结果. 所以 **MOCK 模式(`NEXT_PUBLIC_MOCK=1`,
公开demo使用的模式)浏览器运行时呈现的内容**直到.
`web/src/lib/mock/data.ts`·`handler.ts` 每页上的硬编码样本都是水合作用
回来 `useEffect` 去 `mockHandle()` 把它带到 React 因为它是在状态中渲染的, `out` HTML 这个
不是运行时 JS 仅包含在捆绑包中. 也就是说，汉语通过这条路进来，是经过这个门的。
(过去的公开示威活动。 `/system/logs` 中文在样本日志中就这样静静地存在着,
积压 G114 改为中文).

表面受到来源审查的保护. 渲染目标文件(`web/src` 的 `.tsx`·客户数据)在
其余的中国人必须是以下之一：: 代码注释, `t.has(...)` 实际上被阻止
未使用 i18n 后备, 后端有线格式令牌(聊天提及分类标签·`[模型]` 决策源
哨兵), locale 去 `zh` 仅在以下情况下可见的分支文字, 或者不进行翻译以保持性能。
不是特工大脑(段 [A]) 按原样反映提示的镜子. 不属于其中任何一个
没有 "屏幕上只有中文"将拉面翻译成中文(积压 G115 这个分类
全部确认).

允许·检测范围
-------------
- **中文字符(`去-呵呵` 等等)被允许**做. 因为汉字块和中文字符块在 Unicode 中不重叠，,
  如果只检测到汉字块，中文字符会自动通过。.
- **汉字(CJK 汉字)仅检测**做: 整合汉字及扩展 A(`㐀`–`鿿`), 兼容汉字
  (`豈`–`﫿`), 以及补充平面的扩展。 B 超过(`𠀀`–). 加纳·全角标点符号是
  "汉字"既然不是，那就不是这扇门的目标。(在原文中保存代码块或证据引用
  将范围缩小到汉字，防止误报混入。).

需要上游保存的有意汉字
-----------------------------
现在 `out` HTML 中没有任何汉字，因此允许列表为空。. 保持未来的表现
原文(中国人)当您需要故意将其留在屏幕上时(是的: 语言切换器 `中文` 标签)去
如果发生的话, 下面的字母 `ALLOWED_HAN` 添加它并写一条评论解释为什么留下它。. 这样
大门被严格维护，只允许特定的例外情况通过。.

仅使用标准库，不连接网络.
跑: `python3 -I scripts/check-web-cjk.py [out 目录]`
(默认是存储库根目录 `web/out`. CI 是 `web` 在你的工作目录中
`python3 -I ../scripts/check-web-cjk.py` 它被称为.)
"""
import os
import re
import subprocess
import sys

# CJK 仅汉字块. 中文字符(AC00–D7A3)·加纳·不包括全角标点符号。.
# 范围: 扩张 A + 综合汉字(U+3400–U+9FFF = 㐀–鿿), 兼容汉字(U+F900–U+FAFF = 豈–﫿),
#       补充平面延伸 B 超过(U+20000–U+2FFFF).
HAN = re.compile(r"[㐀-鿿豈-﫿\U00020000-\U0002ffff]")

# 故意在屏幕上留下汉字(是的: 语言切换器 `中文`). 现在是空的 —
# 如果是空的，门上就会挂任何汉字。. 添加例外时，请将文本和原因写在一起。.
# 是的) ALLOWED_HAN = {"中", "文"}  # 语言切换器 '中文' 标签(用于上游控制)
ALLOWED_HAN: set = set()


def find_out_dir() -> str:
    """检查 `out` 决定一个目录.

    如果你给出一个路径作为参数，它将是, 如果没有 git 存储库根目录的 `web/out` 写.
    """
    if len(sys.argv) > 1:
        return os.path.abspath(sys.argv[1])
    root = subprocess.check_output(
        ["git", "rev-parse", "--show-toplevel"], text=True
    ).strip()
    return os.path.join(root, "web", "out")


def main() -> int:
    out_dir = find_out_dir()
    if not os.path.isdir(out_dir):
        print(
            f"错误: 没有要扫描的目录 — {out_dir}\n"
            "第一 `npm run build:static` 使用创建静态导出并运行它."
        )
        return 2

    checked = 0
    offenders = []  # (相对路径, 汉字数, 预览)
    for dirpath, _dirnames, filenames in os.walk(out_dir):
        for name in filenames:
            if not name.endswith(".html"):
                continue
            checked += 1
            path = os.path.join(dirpath, name)
            with open(path, encoding="utf-8", errors="ignore") as fh:
                text = fh.read()
            found = [ch for ch in HAN.findall(text) if ch not in ALLOWED_HAN]
            if found:
                rel = os.path.relpath(path, out_dir)
                kinds = "".join(sorted(set(found)))
                offenders.append((rel, len(found), kinds))

    print(f"检查过的 HTML {checked}狗 (目标: {out_dir})")
    if offenders:
        print(f"中国人(汉字) 泄漏 {len(offenders)}狗档案 — 用户曝光 HTML 里面有汉字:")
        for rel, count, kinds in offenders:
            preview = kinds if len(kinds) <= 30 else kinds[:30] + "…"
            print(f"  out/{rel}: 汉字 {count}狗 · 类型 {preview}")
        print(
            "\n翻译成中文或, 如果是上游需要保留的有意的汉字， "
            "scripts/check-web-cjk.py 的 ALLOWED_HAN 请添加理由."
        )
        return 1
    if checked == 0:
        print("警告: 检查过的 HTML 没有 — 检查是否构建成功.")
        return 2
    print("中国人(汉字) 泄漏 0 — 用户曝光 HTML 里面没有中文字符.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
