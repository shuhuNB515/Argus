#!/usr/bin/env python3
"""存储库中 Markdown 文档的内部链接·图像·检查锚点引用完整性.

一切都被追踪 `.md` 扫描文档并检查两件事.

1. 指向存储库中另一个文件的相对路径链接(`[text](path)`)和图像
   另请参阅(`![alt](path)`·`<img src="path">`)它是否指向一个实际存在的文件？.
2. 文档锚链接(`[text](#heading)`·`[text](other.md#heading)`)目标 `.md`
   它是否指的是文档中实际存在的标题？. 锚 slug 是 GitHub 的规则
   (小写 → `\\p{Word}`·连字符·删除只留下空格 → 用连字符替换空格,
   一样 slug 如果它们重叠，则按出现顺序 `-1`·`-2` 后缀)到标准库
   复制和创造.

如果有一个损坏的引用，则退出代码 1 因为它结束于, CI 合并门
(`.github/workflows/docs.yml`)它可以按原样用于贡献者的本地验证。.

哪些情况不属于检查范围:
- 外面 URL(`http://`·`https://`·`mailto:`·`tel:`·`data:`): 取决于网络
  flaky 因此，它不属于这个确定性门的范围。. 外部链接状态是独立的
  检查.
- `.md` 指向除以下文件之外的文件的链接 `#` 后片(是的: 源代码行号): 标题
  由于它不是锚点，因此它只检查文件是否存在，不检查锚点。.
- 图片参考 `#` 一块: 标题锚在图像中没有意义.
- fenced code block(``` 或 ~~~ 块包裹着) 内部链接示例·标题样式注释:
  实际参考·这是代码示例，而不是标题，因此请跳过它。.

仅使用标准库，不连接网络. 跑: `python3 -I scripts/check-doc-links.py`
"""
import os
import re
import subprocess
import sys
import unicodedata
from collections import defaultdict

# [text](path) 中间图像(`![...]`)非链接. 路径以空格开头(标题 "..." 排除).
MD_LINK = re.compile(r"(?<!\!)\[[^\]]*\]\(([^)\s]+)")
# ![alt](path) 图片链接.
MD_IMAGE = re.compile(r"!\[[^\]]*\]\(([^)\s]+)")
# <img ... src="path" ...> HTML 图像.
HTML_IMAGE = re.compile(r"<img[^>]*\bsrc=[\"']([^\"']+)[\"']", re.IGNORECASE)

# 网络·非文件方案. 锚(`#`)是对同一文档的引用，因此此处不包含它。.
EXTERNAL_PREFIXES = ("http://", "https://", "mailto:", "tel:", "data:")

ATX_HEADING = re.compile(r"^(#{1,6})\s+(.*?)\s*#*\s*$")
# 从标题文本中删除内联降价，只留下纯文本.
MD_INLINE_LINK = re.compile(r"\[([^\]]*)\]\([^)]*\)")


def is_external(target: str) -> bool:
    return target.startswith(EXTERNAL_PREFIXES)


def _is_word_char(ch: str) -> bool:
    """GitHub 的锚 slug 规则留下 `\\p{Word}` 判断是否是文本.

    Ruby 正则表达式 `\\p{Word}` = Letter(L*) + Mark(M*) + Decimal_Number(Nd)
    + Connector_Punctuation(Pc) + Join_Control(U+200C·U+200D). 中文字符·汉字·
    下划线·表情符号的 variation selector(Mn) 背面保持不变。.
    """
    if ch == "_":
        return True
    if ch in ("‌", "‍"):
        return True
    category = unicodedata.category(ch)
    if category[0] in ("L", "M"):
        return True
    return category in ("Nd", "Pc")


def slugify(text: str) -> str:
    """标题文字 GitHub 的锚 slug 转换为(排除重复后缀的处理)."""
    text = text.strip().lower()
    out = []
    for ch in text:
        if ch == " ":
            out.append("-")
        elif ch == "-" or _is_word_char(ch):
            out.append(ch)
    return "".join(out)


def heading_slugs(root: str, md: str) -> set:
    """全部在文件中 ATX 从标题 GitHub 锚 slug 创建一个集合.

    一样 slug 如果它们重叠 GitHub 按照出现顺序，比如 `-1`·`-2` 添加后缀.
    """
    slugs = set()
    counts = defaultdict(int)
    in_fence = False
    with open(os.path.join(root, md), encoding="utf-8", errors="ignore") as fh:
        for line in fh:
            stripped = line.lstrip()
            if stripped.startswith("```") or stripped.startswith("~~~"):
                in_fence = not in_fence
                continue
            if in_fence:
                continue
            match = ATX_HEADING.match(line.rstrip("\n"))
            if not match:
                continue
            raw = match.group(2)
            raw = MD_INLINE_LINK.sub(r"\1", raw)
            raw = raw.replace("`", "").replace("*", "").replace("_", "")
            base = slugify(raw)
            seen = counts[base]
            counts[base] += 1
            slugs.add(base if seen == 0 else f"{base}-{seen}")
    return slugs


def main() -> int:
    root = subprocess.check_output(
        ["git", "rev-parse", "--show-toplevel"], text=True
    ).strip()
    tracked = subprocess.check_output(
        ["git", "ls-files"], cwd=root, text=True
    ).splitlines()
    md_files = [f for f in tracked if f.endswith(".md")]

    slug_cache: dict = {}

    def slugs_of(md_path: str) -> set:
        if md_path not in slug_cache:
            slug_cache[md_path] = heading_slugs(root, md_path)
        return slug_cache[md_path]

    checked_files = 0
    checked_anchors = 0
    broken_files = []
    broken_anchors = []
    for md in md_files:
        with open(os.path.join(root, md), encoding="utf-8", errors="ignore") as fh:
            lines = fh.readlines()
        base_dir = os.path.dirname(md)
        in_fence = False
        for lineno, line in enumerate(lines, 1):
            stripped = line.lstrip()
            if stripped.startswith("```") or stripped.startswith("~~~"):
                in_fence = not in_fence
                continue
            if in_fence:
                continue
            for pattern in (MD_LINK, MD_IMAGE, HTML_IMAGE):
                for match in pattern.finditer(line):
                    target = match.group(1).strip()
                    if is_external(target):
                        continue

                    if target.startswith("#"):
                        file_part, anchor = "", target[1:]
                    elif "#" in target:
                        file_part, anchor = target.split("#", 1)
                    else:
                        file_part, anchor = target, ""
                    file_part = file_part.split("?", 1)[0]
                    anchor = anchor.strip()

                    # 指点 .md 文件(锚检查目标). 如果它是另一个文件的链接，则该文件,
                    # 无文件碎片 `#锚`棉质磁性文件.
                    target_md = None
                    if file_part:
                        checked_files += 1
                        resolved = os.path.normpath(
                            os.path.join(root, base_dir, file_part)
                        )
                        if not os.path.exists(resolved):
                            broken_files.append((md, lineno, target))
                            continue
                        if file_part.endswith(".md"):
                            target_md = os.path.normpath(
                                os.path.join(base_dir, file_part)
                            )
                    else:
                        target_md = md

                    # 锚点是文本的链接(MD_LINK)仅有意义. 图片参考
                    # `#` 跳过碎片。.
                    if anchor and target_md is not None and pattern is MD_LINK:
                        checked_anchors += 1
                        if anchor not in slugs_of(target_md):
                            broken_anchors.append((md, lineno, target, target_md))

    print(
        f"追踪降价 {len(md_files)}狗 · 文件参考 {checked_files}狗 · "
        f"锚参考 {checked_anchors}狗检查"
    )
    if broken_files:
        print(f"损坏的文件参考 {len(broken_files)}狗 — 存储库中不存在指向的文件:")
        for md, lineno, target in broken_files:
            print(f"  {md}:{lineno} -> {target}")
    if broken_anchors:
        print(f"锚点损坏参考 {len(broken_anchors)}狗 — 该标题在目标文档中不存在:")
        for md, lineno, target, target_md in broken_anchors:
            print(f"  {md}:{lineno} -> {target}  (目标: {target_md})")
    if broken_files or broken_anchors:
        return 1
    print("损坏的参考 0 — 所有内部链接·图像·锚点指向真实对象.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
