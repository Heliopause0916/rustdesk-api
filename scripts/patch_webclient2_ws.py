#!/usr/bin/env python3
# 用途：
#   给 vendored rustdesk web client 1.5.3 构建产物 resources/web2/js/dist/index.js 打补丁，
#   修正函数 Ye（构造 WebSocket 地址）中 l6（域名:端口）与 B6（纯域名）两个分支丢弃端口、
#   隐含走 443 的问题。
#
# 同源判定语义：
#   - l6 分支（"域名:端口"）：若冒号前主机名 t 等于 window.location.hostname，
#     返回 `${a}://${window.location.host}${n}`（页面源，含端口，如 8043）；
#     否则保持上游原行为 `${a}://${t}${n}`。
#   - B6 分支（纯域名）：若 u === window.location.hostname，
#     返回 `${a}://${window.location.host}${n}`；否则保持 `${a}://${u}${n}`。
#   其余分支（ws:// 或 wss:// 前缀原样返回、IPv4、IPv6、含冒号但不匹配 l6）一律不动。
#
# 适用部署模型：
#   单 vhost 反向代理在同一端口按路径转发 /ws/id（→hbbs 21118）与 /ws/relay（→hbbr 21119），
#   页面形如 https://rustdesk.ustc-chorus.art:8043/webclient2/。同域地址收敛到页面源
#   （含端口），由代理按路径转发到 hbbs/hbbr。
#
# 已知取舍：
#   罕见的分体部署——页面与非 443 的同域地址并存、但 WebSocket 故意走 443——会受影响：
#   该场景下同域地址将改走页面源（含端口）而非 443。如需恢复上游行为，撤销本补丁。
#
# 幂等：文件已包含两处 window.location.hostname 同源判定标记时，输出 "already patched"
#   并以 0 退出；只含一处标记（部分打补丁）视为异常，非 0 退出。
# fast-fail：未打补丁时，两处旧片段必须各自恰好出现 1 次，否则非 0 退出并打印诊断，
#   不写任何内容；替换仅使用精确字符串，不做模糊匹配。

from typing import NoReturn

import sys

DEFAULT_TARGET = "resources/web2/js/dist/index.js"

# 旧片段：从 minified 文件按真实字节提取（含反引号与模板字面量），各自恰好出现 1 次
OLD_L6 = 'if(l6(u))return`${a}://${t}${n}`}'
NEW_L6 = ('if(l6(u))return t===window.location.hostname'
          '?`${a}://${window.location.host}${n}`'
          ':`${a}://${t}${n}`}')

OLD_B6 = 'else if(B6(u))return`${a}://${u}${n}`;return u}'
NEW_B6 = ('else if(B6(u))return u===window.location.hostname'
          '?`${a}://${window.location.host}${n}`'
          ':`${a}://${u}${n}`;return u}')

# 幂等标记：两处同源判定
MARK_L6 = 't===window.location.hostname'
MARK_B6 = 'u===window.location.hostname'


def fail(msg) -> NoReturn:
    print(f"ERROR: {msg}", file=sys.stderr)
    sys.exit(1)


def main():
    path = sys.argv[1] if len(sys.argv) > 1 else DEFAULT_TARGET
    try:
        with open(path, "r", encoding="utf-8", newline="") as f:
            content = f.read()
    except OSError as exc:
        fail(f"cannot read {path}: {exc}")

    has_l6 = MARK_L6 in content
    has_b6 = MARK_B6 in content
    if has_l6 and has_b6:
        print("already patched")
        return
    if has_l6 or has_b6:
        present = MARK_L6 if has_l6 else MARK_B6
        missing = MARK_B6 if has_l6 else MARK_L6
        fail(f"partially patched: marker {present!r} present but {missing!r} absent in {path}")

    counts = []
    for name, old in (("l6 branch", OLD_L6), ("B6 branch", OLD_B6)):
        count = content.count(old)
        counts.append((name, count))
        if count != 1:
            for n, c in counts:
                print(f"DIAGNOSTIC: {n}: expected exactly 1 occurrence, found {c}", file=sys.stderr)
            fail(f"old fragment count mismatch in {path}; aborting without writing")

    new_content = content.replace(OLD_L6, NEW_L6).replace(OLD_B6, NEW_B6)
    with open(path, "w", encoding="utf-8", newline="") as f:
        f.write(new_content)
    print(f"patched {path}: l6/B6 branches now return page origin (window.location.host) for same-host addresses")


if __name__ == "__main__":
    main()
