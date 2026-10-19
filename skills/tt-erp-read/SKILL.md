---
name: tt-erp-read
description: 正在阅读或需要阅读T100代码时自动触发本技能
---

# ERP 代码阅读

## 获取源码：在线现读 / 离线镜像

1. 区分环境。tt的用户主要是ERP顾问或者ERP开发，他们的电脑上可能配置了多个不同的客户环境。用户提问时可能带了环境信息，比如："帮我看一下XX公司的XX作业的代码"。如果足够明确，可以判断出是哪个客户环境的作业，那么就可以直接去对应的环境中获取源码；如果不够明确，则需要先询问用户，确认客户环境后再去获取源码。

查看环境列表的命令：
```
tt env list
```

2. 找到程序编号。作业编号是用户在客户端看到的功能入口，而程序编号是作业的实现代码。一个作业可能调用多个程序，而一个程序也可能被多个作业复用。作业不等于程序，一个程序也可能被多个作业复用。在处理问题时，我们要优先判断用户说的是作业还是程序。用户大概率讲的是作业，我们需要借助 `tt dict prog <程序编号>` 来找到对应的程序，才能确定去看哪一份源代码。

```
tt dict prog <程序编号> --at <环境>@<企业>
tt dict prog --kw <业务词> --at <环境>@<企业>
```

3. 优先从本地环境读取代码，节省时间。如果tt没有配置代码，则再去尝试从远程环境读取代码。不要帮用户通过 `mirror` 命令去下载代码，时间会很长。

获取本地源码镜像目录的命令：
```
tt file mirror path --at <环境>
```

从远程SFTP读取源码的命令：
```
tt file ls --at <环境> erp/<模块>/4gl
tt file find --at <环境> erp -name '*<程序>*'
tt file grep -rn --at <环境> --include '*.4gl' '<关键字>' erp/<模块>/4gl
tt file head -n 80 --at <环境> erp/<模块>/4gl/<程序>.4gl
```

## 通过数据字典理解字段含义

在T100的数据库中，所有的表和字段都是流水号，无法直接理解其含义。我们需要借助数据字典来获取其真实的中文含义。T100中跟数据库相关的字段有很显而易见的特征，具体的编码规则请参考 [tt-erp-architecture/SKILL.md](../tt-erp-architecture/SKILL.md)。

例如下面一段代码中，`pmds007`、`pmdt036`、`apba010`、`apba103`、`apba104`、`apba105` 都是数据库中的字段编号，我们需要通过数据字典来查询其对应的中文含义，才能理解这段代码的逻辑。
```
INSERT INTO capp115_tmp (sel,docno,seq,pmds007,pmdt036,apba010,apba103,apba104,apba105)
VALUES (gr_qry[li_i].check,gr_qry[li_i].pmdsdocno,gr_qry[li_i].pmdtseq,gr_qry[li_i].pmds007,
        l_price,gr_qry[li_i].apba010,l_amt_ex,l_tax,l_amt_in)
```

从字典中查询表和字段的中文含义的命令：
```
tt dict r.t --brief "capp115_tmp,pmds_t,pmdt_t,apba_t" --at <环境>@<企业>
tt dict r.t apba_t --at <环境>@<企业>
tt dict r.t capp115_tmp --at <环境>@<企业>
```

## 从系统分类码解读业务逻辑

系统分类码SCC是T100中所有下拉框的数据来源。示例 `CALL cl_set_combo_scc("gzpa003","75")`、`CALL cl_set_combo_scc('b_pmds048','2087')`，这是一个常见的绑定下拉框的函数调用，`gzpa003` 和 `b_pmds048` 就是前端控件的编号，而 `75` 和 `2087` 就是系统分类码SCC的值。我们可以通过查询系统分类码SCC来获取其对应的中文含义。

下面的逻辑判断是一个很典型的例子，如果不借助SCC来查询其中文含义，我们是无法理解其逻辑的，因为都是编号：
```
CASE 
  WHEN g_schedule.gzpa003 = "0"
        CALL capp115_process(ls_js)

  WHEN g_schedule.gzpa003 = "1"
        LET ls_js = capp115_transfer_argv(ls_js)
        CALL cl_cmdrun(ls_js)

  WHEN g_schedule.gzpa003 = "2"
        CALL cl_schedule_update_data(g_jobid,ls_js)

  WHEN g_schedule.gzpa003 = "3"
        CALL cl_schedule_update_data(g_jobid,ls_js)
END CASE  
```

从系统分类码SCC中查询中文含义的命令：
```
tt dict scc 75 --at <环境>@<企业>
tt dict scc 2087 --at <环境>@<企业>
```

## 系统参数和单据参数

除了系统分类码，系统参数和单据参数也是理解程序中业务逻辑的重要手段。系统参数主要分为企业级参数(示例`E-CIR-0759`)和据点级参数(示例`S-BWF-0001`)，单据参数也称单据别参数，则是用于设置，单别参数的示例`D-BWF-0004`，大概率是类似的前缀，区分不出来的话，你可以两个查询工具都试一下

查询系统参数示例：

```
tt dict sysp E-CIR-0759 --at <环境>@<企业>
tt dict sysp S-BWF-0001 --at <环境>@<企业>
```

查询单据参数示例：

```
tt dict docp D-BWF-0004 --at <环境>@<企业>
```

## 从系统消息解读报错和提示

T100中很多弹窗、消息的内容是可以复用的，特征也比较明显，比如模块+流水的 `aap-00228`、`aap-00245`，也有很多纯数字的 `-3101`、`100`。在它们附近往往伴随着一些弹窗、提示相关的组件，不借助工具查询也是看不出来含义的。

```
INITIALIZE g_errparam TO NULL
LET g_errparam.code = 'asf-00150'
LET g_errparam.popup = TRUE
CALL cl_err()
```

查询系统消息、错误消息的命令：

```
tt dict msg asf-00150 --at <环境>@<企业>
tt dict msg --at <环境>@<企业> -- -3101
```

## 开窗和校验带值

T100中的开窗查询组件是可以复用到不同的字段的，特征为 `q_xxx()`，它是由设计器生成的。

```
ON ACTION controlp INFIELD sfajwfdocno
  ...
  CALL q_sfajwfdocno_01_wf()
  ...
```

> 开窗是由实体代码的，可以在 `com/qry` 看到生成的代码，但基本上不需要，因为通过工具就可以查到SQL，也要结合查询组件附近的参数看，因为会注入参数。


查询开窗对应的SQL等信息的命令：
```
tt dict r.q q_sfajwfdocno_01_wf --at <环境>@<企业>
tt dict r.q --kw sfajwfdocno --at <环境>@<企业>
```

过栏位的判断也是可以复用的，特征是 `v_xx`，借助工具我们可以方便地查询到其封装的SQL逻辑。

```
AFTER FIELD wfapwf023

#add-point:AFTER FIELD wfapwf023 name="input.a.page3.wfapwf023"
IF NOT cl_null(g_sfakwf3_d[l_ac].wfapwf023) THEN 
    ...
    IF cl_chk_exist("v_wfalwf001_wf") THEN
      ...
    ELSE
    ...
END IF   
```

## 查询数据库

查询数据库也是学习和验证代码逻辑的重要手段，为了安全的只读查询，我们也封装了工具

选择环境，执行一句查询的示例
```
tt db sql --at <环境>@<企业> "select count(*) from capp115_tmp"
tt db sql --at <环境>@<企业> "select bmaa001,bmaastus from bmaa_t where bmaa001='FCPU010100003'"
```
