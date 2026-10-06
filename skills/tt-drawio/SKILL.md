---
name: tt-drawio
description: 用 T100 组件库画 drawio 原型图（界面草图）。当用户要求「排一个 XX 表单」「画个 XX 界面」「用 T100 控件布局」，或提到 tt drawio / 原型图 / 低保真界面时使用。
---

# 用 T100 组件库画原型图

需求调研完、要跟客户确认方案时用。控件尺寸与配色都是 **T100 设计器的惯例** ——
目的是让这些原型图长得一样，而不是各画各的。

## 组件列表

**表单控件**（库 `controls`，缺省）：

| id | 名称 | 尺寸 | 文字槽位 |
| --- | --- | --- | --- |
| `Label` | 标签 | 100×22 | `value` |
| `ButtonQuery` | 串查字段 | 100×22 | `value` |
| `ReferenceLabel` | 参考字段 | 100×22 | `value` |
| `Button` | 按钮 | 105×22 | `value` |
| `Edit` | 单行输入 | 209×22 | `label` `field` |
| `ButtonEdit` | 编辑开窗 | 209×22 | `label` `field` |
| `MultiLangButtonEdit` | 多语言数据字段 | 209×22 | `label` `field` `letter` |
| `SpinEdit` | 数值微调 | 209×22 | `label` `field` |
| `ComboBox` | 下拉选单 | 209×22 | `label` `field` |
| `DateEdit` | 日期 | 209×22 | `label` `field` |
| `CheckBox` | 多选 | 209×22 | `label` |
| `RadioGroup` | 单选组 | 209×22 | `label` |
| `TextEdit` | 多行文字 | 314×110 | `label` `field` |
| `Tree` | 树形 | 152×120 | `n1`…`n5` |
| `Table` | 明细表 | 由列行算 | `h0`… / `r0c0`… |
| `Grid` | 分区框 | 400×220 | `value` |
| `Folder` | 页签容器 | 400×244 | `label` |
| `Group` | 分区框（带标题） | 400×228 | `label` |

**业务组件**（库 `business`，画模块图/接口图用）：
`module-erp` `module-sub` `doc-purchase` `doc-sales` `doc-inventory` `database` `external-system` `interface` `gateway-composite`（都是 `value` 槽位，`gateway-composite` 不能改字）

上面是速查；**尺寸与槽位的权威以命令现打的为准**：

```
tt drawio lib --catalog
```

**只用清单里有的 id**，不要发明控件。

## 怎么用

### 1. 写排版规格（spec）

只写「用哪个控件、放第几列第几行、写什么字」—— **坐标别自己算**：

```json
{
  "title": "采购订单",
  "grid": { "originX": 60, "originY": 60, "colGap": 24, "rowGap": 12 },
  "items": [
    { "shape": "ButtonEdit", "col": 0, "row": 0, "text": { "label": "供应商" } },
    { "shape": "DateEdit",   "col": 1, "row": 0, "text": { "label": "下单日期" } },
    { "shape": "SpinEdit",   "col": 0, "row": 1, "text": { "label": "项次" } },
    { "shape": "Button",     "col": 1, "row": 2, "text": { "value": "查询" } }
  ]
}
```

明细表这类**宽控件**别参与栅格 —— 在 `items` 里写这样一条、给绝对坐标（下面这段是**一条 item**，不是完整 spec）：

```json
{ "shape": "Table", "x": 60, "y": 200,
  "table": { "columns": [ { "title": "品号", "w": 120 }, { "title": "数量", "w": 80 } ], "rows": 3 } }
```

| 字段 | 说明 |
| --- | --- |
| `library` | `controls`（缺省）或 `business` |
| `title` | 图纸名 |
| `grid` | `originX`/`originY`/`colGap`/`rowGap`（缺省 40/40/16/10）；`colWidths` 钉死列宽 |
| `page` | 图纸尺寸，缺省 850×1100 |
| `items[].col` / `row` | **零基**的行列号，脚本自动算像素。**默认用这个** |
| `items[].x` / `y` / `w` / `h` | 给了就绝对定位 / 覆盖尺寸 —— 整屏分区时才用 |
| `items[].text` | 覆盖文字，键是槽位名（见清单），如 `{"label": "供应商"}` |
| `items[].table` | 仅 `Table`：直接给列与行，列数行数随便（`rows` 也可以只给个数字） |
| `items[].pages` / `active` | 仅 `Folder`：多页签。`pages` 给全所有页名，`active`（缺省 0）是当前页 |

### 2. 出图

```bash
tt drawio compose spec.json -o 采购订单.drawio
```

## 三条容易踩的

1. **不要手写 XML。** 库里的控件是「外层分组 + 多个部件」（`DateEdit` 有 10 个单元格），
   手写必然出错。你只输出 spec，脚本负责展开。

2. **列宽是隐式的。** 不写 `colWidths` 时，列宽 = **该列最宽控件的宽度**。`TextEdit` 宽 314，
   放进第一列会把整列撑开、右列被整体推远几十像素 —— 图看着被拉稀，但**不会报任何错**。
   要么用 `colWidths` 钉死，要么让它不参与栅格、单独一行。

3. **`compose` 成功 ≠ 图对。** 它只验结构（id 不重复、引用不悬空、XML 良构），**不验语义** ——
   文字写错控件、分区摆错位置、行列数不对，它都不会报。交付前抽查几处文字，
   把分区对照需求核一遍。

## 一些 T100 惯例

- **字段行**：标签在左、控件在右 —— 字段类控件已内建，摆栅格就行，不用另外放 `Label`
- **一列一个字段**；两列并排时左列 `col:0`、右列 `col:1`
- **代码 + 名称成对**：开窗字段（`ButtonEdit`）旁边常跟一个只读描述，用 `ReferenceLabel`
- **按钮**通常单独放最后一行
- **明细区**在 T100 里一律叫 `s_detail1`、`s_detail2`…，表头外的标题文字照此
- **标准工具栏四件**（`datainfo` / `insert` / `output` / `query`）查不到官方中文，
  按 id 原样画，**不要编中文**
