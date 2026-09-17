#!/usr/bin/env python3
"""给 heat.exe 生成的 .wxs 补上卸载时要删的目录登记。

为什么需要：这个 MSI 是**用户级**安装，目录都落在用户配置文件下（
%LOCALAPPDATA%\\Programs\\TT）。MSI 的 ICE64 校验要求这类目录必须登记进
RemoveFile 表，否则卸载后目录会原样留在磁盘上。heat 只生成组件与文件，
不生成 RemoveFolder，所以在构建流水线里补这一道。

规则（照 MSI 的约束来）：RemoveFolder 必须挂在某个 Component 里，且它指向的
目录要是该组件所在目录或它的上级。于是：
  · 目录自己带组件（skill 目录就是这种）→ 往它的第一个组件里塞一个 RemoveFolder；
  · 目录自己不带组件（skills/ 这种只有子目录的）→ 补一个只挂注册表键的组件，
    专门负责卸载时删掉这个空目录，并把它登记进 ComponentGroup；
  · 只挂注册表键的组件用 Guid="*"：WiX 会按 keypath 推出稳定的 GUID，
    重打包装升级时不会认成另一个组件。

用法：wix_removefolders.py <heat 生成的 .wxs>
就地改写。
"""
import sys
import xml.etree.ElementTree as ET

NS = "http://schemas.microsoft.com/wix/2006/wi"


def q(tag):
    return "{%s}%s" % (NS, tag)


def main(path):
    ET.register_namespace("", NS)  # 输出用默认命名空间,别变成 ns0:
    tree = ET.parse(path)
    root = tree.getroot()

    groups = list(root.iter(q("ComponentGroup")))
    added = 0
    for d in root.iter(q("Directory")):
        did = d.get("Id")
        if not did:
            continue
        comps = [c for c in d if c.tag == q("Component")]
        if comps:
            # 目录里有文件:在它的第一个组件里登记「卸载时删掉本目录」
            rm = ET.SubElement(comps[0], q("RemoveFolder"))
            rm.set("Id", "rm_" + did)
            rm.set("Directory", did)
            rm.set("On", "uninstall")
        else:
            # 只有子目录、自己没文件:补一个「空目录组件」,否则它会留在磁盘上
            c = ET.Element(q("Component"))
            c.set("Id", "rmdir_" + did)
            c.set("Guid", "*")
            # 不写 Directory:组件嵌在 <Directory> 里时该属性必须省略(否则 CNDL0062),
            # 所在目录由嵌套位置决定,也就是它自己。
            rm = ET.SubElement(c, q("RemoveFolder"))
            rm.set("Id", "rm_" + did)
            rm.set("Directory", did)
            rm.set("On", "uninstall")
            rv = ET.SubElement(c, q("RegistryValue"))
            rv.set("Root", "HKCU")
            rv.set("Key", r"Software\TT\dirs")
            rv.set("Name", did)
            rv.set("Type", "integer")
            rv.set("Value", "1")
            rv.set("KeyPath", "yes")
            d.append(c)
            for g in groups:
                ref = ET.SubElement(g, q("ComponentRef"))
                ref.set("Id", "rmdir_" + did)
            added += 1

    # ICE38:用户级安装里,组件的 KeyPath 必须是 HKCU 下的注册表键,不能用文件 ——
    # 用文件做 KeyPath 时,MSI 是按"文件版本"判定该组件装没装,这个判定是全机范围的,
    # 和"按用户安装"的语义对不上。改成注册表键之后,每个用户各自的安装状态是清楚的。
    for c in list(root.iter(q("Component"))):
        f = c.find(q("File"))
        if f is None or f.get("KeyPath") != "yes":
            continue
        cid = c.get("Id")
        del f.attrib["KeyPath"]
        rv = ET.SubElement(c, q("RegistryValue"))
        rv.set("Root", "HKCU")
        rv.set("Key", r"Software\TT\components")
        rv.set("Name", cid)
        rv.set("Type", "integer")
        rv.set("Value", "1")
        rv.set("KeyPath", "yes")

    tree.write(path, encoding="utf-8", xml_declaration=True)
    print("wix_removefolders: %s (补了 %d 个空目录组件)" % (path, added))


if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.exit("用法: wix_removefolders.py <heat 生成的 .wxs>")
    main(sys.argv[1])
