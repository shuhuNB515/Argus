#!/usr/bin/env python3
"""跟踪的 Markdown 文档指向的外部链接(http·https)检查是否还活着.

姐妹脚本 `check-doc-links.py` 是存储库内的内部链接·外部仅通过查看锚点 URL 银
按设计跳过(取决于网络 flaky 因此，它不在合并门中处理。). 这个
脚本是 "单独检查"负责. README 顶部·防御指南·detections README 去
对于访客和防守者 "去这里"并参考外部链接(事故报告窗口 boho.or.kr·
privacy.go.kr·pipc.go.kr·fsec.or.kr, CISA KEV, OWASP·SigmaHQ·Suricata·MITRE·MISP,
原始演示 argus-demo.vercel.app, GitHub 徽章等)恶化·移动·关闭时悄然破裂
它仍然存在, 定期做·手动捕捉.

检查目标 URL 选择规则
- 一切都被追踪 `.md` 看看, fenced code block(``` 或 ~~~)和内联代码跨度
  (`` `...` ``) 里面 URL 被跳过. 在其中 URL 是一个命令示例·设定值·外部引用
  资产(是的: 附带问题验证文件 `id.redhaze.top`)伊拉 "供读者参考的链接"不.
- 降价链接·图像(`[text](url)`·`![alt](url)`), 自动链接(`<url>`), 和其余的
  散文中的人 URL 在 http·https 收集地址.
- 预约·排除占位符主机: localhost·社论/环回 IP(127.·10.·192.168.·
  169.254.·172.16~31.·0.0.0.0·::1), RFC 2606/6761 预约(example.com/org/net/edu·
  `*.example.*`·`.test`·`.invalid`·`.local`·`.tld`), 没有点 FQDN 名字不是(是的: `target`).

如何检查你是否还活着 (MAINTAINING.md 8.2 翻译教训)
国内公众·安全机构网站 HEAD 要求·基本的 User-Agent 拒绝或多次
因为它会重定向, 简单的验证会将健康的链接误认为是损坏的链接。. 所以这个脚本
**浏览器 User-Agent 通过, GET 到, 遵循重定向** 确认. 三种状态
分享.
- OK: 最终状态是 2xx·3xx. 该链接有效.
- RESTRICTED: 401·403·405·429. 主机存活，但验证方式为服务器访问策略。(阻止机器人·
  方法拒绝·限速)这只是一个阻塞，而不是一个断开的链接。. 报告但不视为失败.
- DOWN: 404·410·5xx(即使重试后)·DNS/连接/超时/SSL 错误(即使重试后). 实际上
  很可能坏了.

网络错误·5xx·429 稍有延迟地重试，以区分临时闪烁和真正的故障。.
仅使用标准库. 这个脚本是 **公开文件已经指出** 另请参阅 URL 仅
GET 只需发送它来检查其生存情况即可。, 扫描任何目标·请勿探测.

已知的例外情况(allowlist)
`scripts/external-links-allowlist.txt` 写在 URL 这个 DOWN 当它出来的时候 "ALLOWED" 分别
分类和 strict Don't put it in the exit code. We don't own it so we can't fix it, 上游文本保存
文件(是的: 上游 CHANGELOG.zh.md)透明地记录这些继承的死链接, 循环 strict 检查
由于这一点，它不会永久变红。 **仅当出现新的断开链接时** 让它变红.

跑
- 基本的(报告·退出代码 0): `python3 -I scripts/check-external-links.py`
- 循环而不是合并门/释放检查(如果有新的断开链接，它将呈红色。): `--strict` (allowlist 在
  没有 DOWN 如果存在其中之一，则退出代码为 1). RESTRICTED·ALLOWED 是 strict 这也不被视为失败。.
- 在没有网络的情况下仅预览提取集: `--list` (不来电检查目标 URL 并输出源).
"""
import argparse
import http.client
import os
import re
import socket
import ssl
import subprocess
import sys
import time
import urllib.error
import urllib.request
from collections import defaultdict
from urllib.parse import urlsplit

# [text](url) 文字链接(不是图像)哇 ![alt](url) 图像. 路径为空·右括号之前.
MD_LINK = re.compile(r"(?<!\!)\[[^\]]*\]\(([^)\s]+)")
MD_IMAGE = re.compile(r"!\[[^\]]*\]\(([^)\s]+)")
# <https://...> 自动链接.
AUTOLINK = re.compile(r"<(https?://[^>\s]+)>")
# 散文中的人 URL. 后面的标点符号被去掉。.
BARE_URL = re.compile(r"https?://[^\s)>\]\"'`]+")
# 内联代码跨度 `...` (单反引号). 解压前清理空间.
INLINE_CODE = re.compile(r"`[^`]*`")
# 散文 URL 标点符号通常加在末尾.
TRAILING_PUNCT = ".,;:!?\"'»)]}>"

# IPv4 社论/环回/链接本地频段(172.16~31. 是一个单独的模式).
PRIVATE_IPV4 = re.compile(r"^(127\.|10\.|192\.168\.|169\.254\.|0\.0\.0\.0$)")
PRIVATE_IPV4_172 = re.compile(r"^172\.(1[6-9]|2\d|3[01])\.")

BROWSER_UA = (
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
    "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0 Safari/537.36"
)

# 主机处于活动状态但验证方法被服务器策略阻止的状态代码(没有破损).
RESTRICTED_CODES = {401, 403, 405, 429}

ALLOWLIST_PATH = os.path.join("scripts", "external-links-allowlist.txt")


def load_allowlist(root: str) -> set:
    """我们无法修复的已知问题 DOWN URL 读集(如果不是，则为空集).

    排成一行 URL 一. `#` 后面是评论, 空行·忽略仅注释行.
    """
    path = os.path.join(root, ALLOWLIST_PATH)
    allow = set()
    if not os.path.exists(path):
        return allow
    with open(path, encoding="utf-8") as fh:
        for line in fh:
            line = line.split("#", 1)[0].strip()
            if line:
                allow.add(line)
    return allow


def strip_code(line: str) -> str:
    """fence 从一个外行删除内联代码跨度(在其中 URL 不是参考)."""
    return INLINE_CODE.sub(" ", line)


def is_checkable(url: str) -> bool:
    """预约·过滤掉占位主机, 外观其实还是值得一看的。 URL 只有叶子."""
    parts = urlsplit(url)
    if parts.scheme not in ("http", "https"):
        return False
    host = parts.hostname
    if not host:
        return False
    host = host.lower()
    if host == "localhost" or host.endswith(".localhost"):
        return False
    if host == "::1":
        return False
    if PRIVATE_IPV4.match(host) or PRIVATE_IPV4_172.match(host):
        return False
    if host in ("example.com", "example.org", "example.net", "example.edu"):
        return False
    if host.endswith((".example.com", ".example.org", ".example.net", ".example")):
        return False
    if host.endswith((".test", ".invalid", ".local", ".localdomain", ".tld")):
        return False
    if "." not in host:  # FQDN 不是最上面的名字(是的: target)
        return False
    return True


def extract(root: str, md_files: list) -> dict:
    """追踪 .md 外部检查 URL → [(文件, 行号), ...] 做一本字典."""
    found = defaultdict(list)
    for md in md_files:
        with open(os.path.join(root, md), encoding="utf-8", errors="ignore") as fh:
            lines = fh.readlines()
        in_fence = False
        for lineno, raw in enumerate(lines, 1):
            stripped = raw.lstrip()
            if stripped.startswith("```") or stripped.startswith("~~~"):
                in_fence = not in_fence
                continue
            if in_fence:
                continue
            line = strip_code(raw)
            candidates = []
            for pattern in (MD_LINK, MD_IMAGE, AUTOLINK):
                candidates.extend(m.group(1) for m in pattern.finditer(line))
            for m in BARE_URL.finditer(line):
                candidates.append(m.group(0).rstrip(TRAILING_PUNCT))
            for url in candidates:
                url = url.strip().rstrip(TRAILING_PUNCT)
                if is_checkable(url):
                    where = (md, lineno)
                    if where not in found[url]:
                        found[url].append(where)
    return found


def probe(url: str, timeout: float, retries: int, delay: float):
    """URL 在浏览器中 UA 通过 GET 发送 (分类, 详情) 回报.

    分类 "OK" · "RESTRICTED" · "DOWN". 网络错误·5xx·429 重试.
    """
    req = urllib.request.Request(
        url,
        method="GET",
        headers={
            "User-Agent": BROWSER_UA,
            "Accept": "*/*",
            "Accept-Language": "ko,en;q=0.8",
        },
    )
    last = ""
    for attempt in range(retries + 1):
        try:
            with urllib.request.urlopen(req, timeout=timeout) as resp:
                code = resp.getcode()
                if code in RESTRICTED_CODES:
                    return "RESTRICTED", f"HTTP {code}"
                return "OK", f"HTTP {code}"
        except urllib.error.HTTPError as exc:
            code = exc.code
            if code in RESTRICTED_CODES:
                return "RESTRICTED", f"HTTP {code}"
            if code >= 500 or code == 408:
                last = f"HTTP {code}"  # 这可能是暂时的，所以重试
            else:
                return "DOWN", f"HTTP {code}"  # 404·410 等已确认, 无需重试
        except (
            urllib.error.URLError,
            http.client.HTTPException,
            ssl.SSLError,
            socket.timeout,
            ConnectionError,
            TimeoutError,
            OSError,
        ) as exc:
            reason = getattr(exc, "reason", exc)
            last = f"{type(exc).__name__}: {reason}"
        if attempt < retries:
            time.sleep(delay * (attempt + 1))
    return "DOWN", last


def main() -> int:
    ap = argparse.ArgumentParser(description="检查跟踪 Markdown 中的外部链接生存情况")
    ap.add_argument("--strict", action="store_true",
                    help="DOWN 如果存在其中之一，则退出代码为 1 (循环/用于发布检查)")
    ap.add_argument("--list", action="store_true",
                    help="无需网络调用的检查目标 URL 并仅输出源")
    ap.add_argument("--timeout", type=float, default=15.0, help="请求超时(秒)")
    ap.add_argument("--retries", type=int, default=2, help="网络错误·5xx·429 重试次数")
    ap.add_argument("--delay", type=float, default=0.5, help="通话之间·重试之间的默认延迟(秒)")
    args = ap.parse_args()

    root = subprocess.check_output(
        ["git", "rev-parse", "--show-toplevel"], text=True
    ).strip()
    tracked = subprocess.check_output(["git", "ls-files"], cwd=root, text=True).splitlines()
    md_files = [f for f in tracked if f.endswith(".md")]

    allow = load_allowlist(root)
    found = extract(root, md_files)
    urls = sorted(found)
    print(f"追踪降价 {len(md_files)}狗在外面接受检查 URL {len(urls)}狗收藏")

    if args.list:
        for url in urls:
            where = ", ".join(f"{f}:{ln}" for f, ln in found[url])
            tag = " [allowlist]" if url in allow else ""
            print(f"  {url}{tag}  ({where})")
        return 0

    results = {"OK": [], "RESTRICTED": [], "ALLOWED": [], "DOWN": []}
    for i, url in enumerate(urls):
        if i:
            time.sleep(args.delay)  # 通话之间有轻微延迟(沙箱控制规避)
        verdict, detail = probe(url, args.timeout, args.retries, args.delay)
        if verdict == "DOWN" and url in allow:
            verdict = "ALLOWED"  # 已知无法更正的上游继承。 DOWN
        results[verdict].append((url, detail))
        print(f"  [{verdict:10}] {url} — {detail}")

    print(
        f"\n总结: OK {len(results['OK'])} · RESTRICTED {len(results['RESTRICTED'])} · "
        f"ALLOWED {len(results['ALLOWED'])} · DOWN {len(results['DOWN'])}"
    )
    if results["RESTRICTED"]:
        print("RESTRICTED(宿主生存·验证方法被阻止, 不是损坏的链接):")
        for url, detail in results["RESTRICTED"]:
            print(f"  {url} — {detail}")
    if results["ALLOWED"]:
        print("ALLOWED(allowlist 已知写于 DOWN — 我们无法修复它 strict 排除):")
        for url, detail in results["ALLOWED"]:
            where = ", ".join(f"{f}:{ln}" for f, ln in found[url])
            print(f"  {url} — {detail}  (另请参阅: {where})")
    if results["DOWN"]:
        print("DOWN(很有可能坏了 — 需要确认):")
        for url, detail in results["DOWN"]:
            where = ", ".join(f"{f}:{ln}" for f, ln in found[url])
            print(f"  {url} — {detail}  (另请参阅: {where})")

    if args.strict and results["DOWN"]:
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
