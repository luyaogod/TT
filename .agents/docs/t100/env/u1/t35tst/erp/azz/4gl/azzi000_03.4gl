#該程式未解開Section, 採用最新樣板產出!
#該程式為freestyle程式!
{<section id="azzi000_03.description" type="s" >}
#應用 a00 樣板自動產生(Version:3)
#+ Standard Version.....: SD版次:0014(2014-09-30 13:56:15), PR版次:0014(2026-04-28 10:51:02)
#+ Customerized Version.: SD版次:0000(1900-01-01 00:00:00), PR版次:0000(1900-01-01 00:00:00)
#+ Build......: 000690
#+ Filename...: azzi000_03
#+ Description: 系統流程
#+ Creator....: 01274(2014-09-30 10:44:49)
#+ Modifier...: 01274 -SD/PR- 20992
 
{</section>}
 
{<section id="azzi000_03.global" type="s" >}
#應用 p00 樣板自動產生(Version:6)
#add-point:填寫註解說明 name="main.memo"
#+ Modifier...: No.161104-00050#1 2016/11/04 By tsai_yen 首頁流程圖目錄多語言要對應企業編號
#+ Modifier...: No.241231-00003#1 2024/12/31 By 20992    gzbc_t 增加 gzbcent 欄位
#+ Modifier...: No.250307-00032#1 2025/03/10 By 20992    修正連接線未正常顯示問題
#+ Modifier...: No.260427-00008#1 2026/04/27 By 20992    修正流程圖 insert 及 update 出現 PK 重複問題
#+ Modifier...: No.260428-00003#1 2026/04/28 By 20992    取消流程圖 update 部分，以 insert 處理即可

#end add-point
#add-point:填寫註解說明(客製用) name="main.memo_customerization"

#end add-point
#(ver:6) ---start---
#add-point:填寫註解說明(行業用) name="global.memo_industry"

#end add-point
#(ver:6) --- end ---
 
IMPORT os
#add-point:增加匯入項目 name="main.import"
IMPORT util
#end add-point
 
SCHEMA ds
 
GLOBALS "../../cfg/top_global.inc"
#add-point:增加匯入變數檔 name="global.inc"

#end add-point
 
{</section>}
 
{<section id="azzi000_03.free_style_variable" type="s" >}
#add-point:free_style模組變數(Module Variable) name="free_style.variable"
TYPE  t_tree  DYNAMIC ARRAY OF RECORD  #樹
         id  STRING,
         pId STRING,
         name STRING,
         open BOOLEAN,
         icon STRING,
         iconOpen STRING,
         iconClose STRING,
         isNode boolean
      END RECORD
TYPE t_position      RECORD
         top            STRING,
         left           STRING
                     END RECORD
TYPE t_size          RECORD
         width          STRING,
         height         STRING
                     END RECORD
TYPE t_containers    DYNAMIC ARRAY OF RECORD
         id             STRING,
         label          STRING,    #gzbbl_t.gzbbl003 簡稱
         desc           STRING,    #gzbbl_t.gzbbl004 全稱
         module         STRING,    #gzbb_t.gzbb003   節點類型
         other          STRING,    #gzbb_t.gzbb005  類型是program用的
         image          STRING,    #gzbb_t.gzbb005  類型非program用的(圖片路徑)
         template       STRING,    #可用節點類型gzbb005
         position       t_position,
         size           t_size,
         css            STRING
                     END RECORD
TYPE t_parameters    RECORD         #gzbc_t.gzbc004   其它資訊
         label          STRING
                     END RECORD
TYPE t_connections   DYNAMIC ARRAY OF RECORD
         parameters  t_parameters,
         source      STRING,  #gzbc_t.gzbc001   來源節點代號
         target      STRING   #gzbc_t.gzbc002   目的節點代號
                     END RECORD
DEFINE g_ref_fields          DYNAMIC ARRAY OF VARCHAR(500) #reference用陣列
DEFINE g_ref_vars            DYNAMIC ARRAY OF VARCHAR(500) #reference用陣列
DEFINE g_rtn_fields          DYNAMIC ARRAY OF VARCHAR(500) #reference用陣列
#end add-point
 
{</section>}
 
{<section id="azzi000_03.global_variable" type="s" >}
#add-point:自定義模組變數(Module Variable) name="global.variable"
DEFINE g_tree           t_tree
DEFINE g_first_gzba001  STRING  #第一個要先展開的樹
#end add-point
 
{</section>}
 
{<section id="azzi000_03.other_dialog" type="s" >}

 
{</section>}
 
{<section id="azzi000_03.other_function" readonly="Y" type="s" >}

################################################################################
# Descriptions...: 系統流程圖
# Memo...........:
# Usage..........: CALL azzi000_03()
#                  RETURNING none
# Input parameter: none
# Return code....: none
# Date & Author..: 劉偉先
# Modify.........:
################################################################################
PUBLIC FUNCTION azzi000_03()
   DEFINE l_flow_wc              STRING
   DEFINE l_flow_cfg             RECORD   #flow的設定
            editable             BOOLEAN, #是否可編輯
            imagepath            STRING,  #節點圖示是相對路徑，必須指定image路徑給WebComponents使用
               lans              RECORD   #要給WebComponent中的多語言文字
                  C_VIEW_FLOW       STRING  #流程檢視的多語言
               END RECORD
          END RECORD
   DEFINE l_container   RECORD
            label          STRING,
            module         STRING
                        END RECORD
   DEFINE la_param      RECORD                  #程式串查用變數
             prog          STRING,                 #串查程式名稱
             param         DYNAMIC ARRAY OF STRING #傳遞變數
                        END RECORD
   DEFINE l_json        STRING
   DEFINE l_sb          base.StringBuffer
   
   INITIALIZE la_param TO null
   LET la_param.prog = "azzi000"
   LET l_sb = base.StringBuffer.create()
   CALL l_sb.append(cl_ap_url("GWC", util.JSON.stringify(la_param)))
   CALL l_sb.append("/DUA_HTML5/Image/")
   CALL l_sb.replace("/wa/r/", "/wa/i/", 1)
   
   #開啟Flow為編輯模式
   LET l_flow_cfg.editable = FALSE
   LET l_flow_cfg.lans.C_VIEW_FLOW = cl_getmsg("azz-00259", g_dlang)
   LET l_flow_cfg.imagepath = l_sb.toString()
   LET l_flow_wc = util.JSON.stringify(l_flow_cfg)
   
   OPEN WINDOW w_azzi000_03 WITH FORM cl_ap_formpath("azz","azzi000_03")
         ATTRIBUTE(STYLE="dialog")
   
   CALL azzi000_03_init()
   
   DIALOG ATTRIBUTES(UNBUFFERED=TRUE)
      INPUT BY NAME l_flow_wc ATTRIBUTES(WITHOUT DEFAULTS=TRUE)
         #WebComponent載入完成後
         ON ACTION wc_init
            LET l_json = util.JSON.stringify(g_tree)
            CALL azzi000_03_submit_wc("l_flow_wc", "tree", l_json)
         
         #點樹節點時
         ON ACTION wc_tree_click
            LET l_json = azzi000_03_get_flow(l_flow_wc)
            CALL azzi000_03_submit_wc("l_flow_wc", "flow", l_json)
            
         #雙擊流程圖節點
         ON ACTION wc_dblclick_node
            CALL util.JSON.parse(l_flow_wc, l_container)

      END INPUT

      ON ACTION close
         EXIT DIALOG
      ON ACTION exit
         EXIT DIALOG
   END DIALOG
   CLOSE WINDOW w_azzi000_03
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PUBLIC FUNCTION azzi000_03_submit_wc(p_wcname,p_type,p_data)
   DEFINE p_wcname      STRING,
       p_type        STRING,
       p_data        STRING
   CALL azzi000_03_prop_comp(p_wcname, "type", p_type)
   CALL azzi000_03_prop_comp(p_wcname, "data", p_data)
END FUNCTION

################################################################################
# Descriptions...: 初始化分類選單及第一個分類流程圖
# Memo...........:
# Usage..........: CALL azzi000_03_init()
#                  RETURNING none
# Input parameter: none
# Return code....: none
# Date & Author..: 劉偉先
# Modify.........:
################################################################################
PUBLIC FUNCTION azzi000_03_init()
   DEFINE l_sql      STRING
   DEFINE l_where_p  STRING
   DEFINE l_gzxa013  LIKE gzxa_t.gzxa013
   DEFINE l_cnt      INTEGER

   CALL azzi000_03_update_gzbc() #241231-00003#1 add
   
   SELECT gzxa013 INTO l_gzxa013 FROM gzxa_t WHERE gzxa001 = g_account AND gzxaent = g_enterprise

   #如果設定的目錄不存在，則將目錄指定為NULL
   IF NOT cl_null(l_gzxa013) THEN
      SELECT COUNT(*) INTO l_cnt FROM gzba_t WHERE gzba001 = l_gzxa013 AND gzbaent = g_enterprise
      IF l_cnt == 0 THEN
         LET l_gzxa013 = NULL
      END IF
   END IF

   IF cl_null(l_gzxa013) THEN
      LET l_where_p = " gzba002 IS NULL "
   ELSE
      LET l_where_p = " gzba002 = '", l_gzxa013, "' "
   END IF

   LET l_sql = "SELECT gzba001, gzba002, gzbal003  FROM gzba_t ",
               " LEFT JOIN gzbal_t ON gzbal001 = gzba001 AND gzbal002 = '", g_lang CLIPPED, "' ",
                               "  AND gzbaent = gzbalent",   #161104-00050#1
               "    WHERE ", l_where_p,
               "  AND gzbaent = ", g_enterprise, " ",
               "  AND gzbastus = 'Y' ",
               "ORDER BY gzba003 ASC"

   DECLARE declare_gzba_root CURSOR FROM l_sql

   LET l_sql = "SELECT COUNT(*) FROM gzba_t WHERE gzba002 = ? "
   PREPARE pre_gzba_child_cnt FROM l_sql

   LET l_sql = "SELECT gzba001, gzba002, gzbal003  FROM gzba_t ",
               "LEFT JOIN gzbal_t ON gzbal001 = gzba001 AND gzbal002 = '", g_lang CLIPPED, "' ",
                               "  AND gzbaent = gzbalent",   #161104-00050#1
               "    WHERE gzba002 = ? ",
               "  AND gzbaent = ", g_enterprise, " ",
               "  AND gzbastus = 'Y' ",
               "ORDER BY gzba003 ASC"
   DECLARE declare_gzba_child CURSOR FROM l_sql

   LET l_sql = "SELECT * FROM gzbb_t WHERE gzbb002 = ? AND gzbbent = '", g_enterprise CLIPPED ,"'"
   PREPARE pre_gzbb_by_gzbb002  FROM l_sql

   #LET l_sql = "SELECT * FROM gzbc_t WHERE gzbc001 = ? AND gzbc003 = ? " #241231-00003#1 mark
   LET l_sql = "SELECT * FROM gzbc_t WHERE gzbc001 = ? AND gzbc003 = ? AND gzbcent = '", g_enterprise CLIPPED ,"'" #241231-00003#1 add
   PREPARE pre_gzbc_by_gzbb001 FROM l_sql

   LET l_sql = "SELECT * FROM gzbbl_t WHERE gzbbl001 = ? AND gzbblent = '", g_enterprise, "' "
   PREPARE pre_gzbbl_by_gzbb001 FROM l_sql

   DECLARE declare_gzbb_by_gzbb002 CURSOR FOR pre_gzbb_by_gzbb002
   DECLARE declare_gzbc_by_gzbb001 CURSOR FOR pre_gzbc_by_gzbb001
   DECLARE declare_gzbbl_by_gzbb001 CURSOR FOR pre_gzbbl_by_gzbb001

END FUNCTION

################################################################################
# Descriptions...: 初始化選單分類樹
# Memo...........:
# Usage..........: CALL azzi000_03_init_tree()
#                  RETURNING none
# Input parameter: none
# Return code....: none
# Date & Author..: 劉偉先
# Modify.........:
################################################################################
PUBLIC FUNCTION azzi000_03_init_tree()
   DEFINE l_gzba_arr    DYNAMIC ARRAY OF RECORD
            gzba001        LIKE gzba_t.gzba001, 
            gzba002        LIKE gzba_t.gzba002,
            gzbal003       LIKE gzbal_t.gzbal003
                        END RECORD
   DEFINE l_gzba_t      RECORD
            gzba001        LIKE gzba_t.gzba001, 
            gzba002        LIKE gzba_t.gzba002,
            gzbal003       LIKE gzbal_t.gzbal003
                        END RECORD
   DEFINE l_len         INTEGER
   DEFINE l_json        RECORD
             expand         STRING,
             nodes          t_tree
                        END RECORD
   DEFINE l_cnt         INTEGER
   DEFINE l_idx         INTEGER
   DEFINE l_gzbb_cnt    INTEGER
   DEFINE l_first_expand   STRING
   
   CALL g_tree.clear()
   OPEN declare_gzba_root
   FOREACH declare_gzba_root INTO l_gzba_t.*
      CALL l_gzba_arr.appendElement()
      LET l_len = l_gzba_arr.getLength()
      LET l_gzba_arr[l_len].*   = l_gzba_t.*
   END FOREACH
   CLOSE declare_gzba_root
   
   FOR l_idx = 1 TO l_gzba_arr.getLength()
      #先把第一筆有資料的記下來，預備載入時先展開
      CALL azzi000_03_first_expand(l_gzba_arr[l_idx].gzba001)
      
      CALL g_tree.appendElement()
      LET l_len = g_tree.getLength()
      LET g_tree[l_len].id   = l_gzba_arr[l_idx].gzba001    #分類代號
      LET g_tree[l_len].pid  = l_gzba_arr[l_idx].gzba002    #上層分類代號
      LET g_tree[l_len].name = l_gzba_arr[l_idx].gzbal003   #分類名稱
      IF cl_null(g_tree[l_len].name) THEN LET g_tree[l_len].name = " " END IF
      #LET g_tree[l_len].open = (l_idx == 1)                         #第一個先幫開啟
      LET g_tree[l_len].open = FALSE                        #展開
      EXECUTE pre_gzba_child_cnt USING l_gzba_arr[l_idx].gzba001 INTO l_cnt
      IF l_cnt > 0 THEN
         LET g_tree[l_len].isnode = TRUE 
         CALL azzi000_03_tree_child(l_gzba_arr[l_idx].gzba001)         
      ELSE
         LET g_tree[l_len].isnode = FALSE
      END IF
   END FOR
   LET l_json.expand = g_first_gzba001
   LET l_json.nodes = g_tree

   RETURN util.JSON.stringify(l_json)
END FUNCTION

################################################################################
# Descriptions...: 填充指定的分類子項
# Memo...........:
# Usage..........: CALL azzi000_03_tree_child(p_gzba002)
#                  RETURNING none
# Input parameter: p_gzba002      上層分類代號
# Return code....: none
# Date & Author..: 劉偉先
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi000_03_tree_child(p_gzba002)
   DEFINE p_gzba002        LIKE gzba_t.gzba002
   DEFINE l_gzba_arr    DYNAMIC ARRAY OF RECORD
            gzba001        LIKE gzba_t.gzba001, 
            gzba002        LIKE gzba_t.gzba002,
            gzbal003       LIKE gzbal_t.gzbal003
                        END RECORD
   DEFINE l_gzba_t      RECORD
            gzba001        LIKE gzba_t.gzba001, 
            gzba002        LIKE gzba_t.gzba002,
            gzbal003       LIKE gzbal_t.gzbal003
                        END RECORD
   DEFINE l_len         INTEGER
   DEFINE l_cnt         INTEGER
   DEFINE l_idx         INTEGER
   
   OPEN declare_gzba_child USING p_gzba002
   FOREACH declare_gzba_child INTO l_gzba_t.*
      CALL l_gzba_arr.appendElement()
      LET l_len = l_gzba_arr.getLength()
      LET l_gzba_arr[l_len].*   = l_gzba_t.*
   END FOREACH
   CLOSE declare_gzba_child
   
   FOR l_idx = 1 TO l_gzba_arr.getLength()
      #先把第一筆有資料的記下來，預備載入時先展開
      CALL azzi000_03_first_expand(l_gzba_arr[l_idx].gzba001)
      
      CALL g_tree.appendElement()
      LET l_len = g_tree.getLength()
      LET g_tree[l_len].id   = l_gzba_arr[l_idx].gzba001    #分類代號
      LET g_tree[l_len].pid  = l_gzba_arr[l_idx].gzba002    #上層分類代號
      LET g_tree[l_len].name = l_gzba_arr[l_idx].gzbal003   #分類名稱
      IF cl_null(g_tree[l_len].name) THEN LET g_tree[l_len].name = " " END IF
      LET g_tree[l_len].open = FALSE                         #開啟分類
      EXECUTE pre_gzba_child_cnt USING l_gzba_arr[l_idx].gzba001 INTO l_cnt
      IF l_cnt > 0 THEN
         LET g_tree[l_len].isnode = TRUE 
         CALL azzi000_03_tree_child(l_gzba_arr[l_idx].gzba001)         
      ELSE
         LET g_tree[l_len].isnode = FALSE
      END IF
   END FOR
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PUBLIC FUNCTION azzi000_03_prop_comp(p_wcname,p_prop_name,p_prop_value)
   DEFINE p_wcname STRING
   DEFINE p_prop_name STRING
   DEFINE p_prop_value STRING
   DEFINE win ui.Window
   DEFINE l_ff    om.DomNode
   DEFINE win_node      om.DomNode
   DEFINE l_nl   om.NodeList
   DEFINE l_wc  om.DomNode
   DEFINE l_dict    om.DomNode
   DEFINE l_prop    om.DomNode

   LET win = ui.Window.getCurrent()
   LET win_node = win.getNode()

   LET l_nl = win_node.selectByPath("//FormField[@name=\"formonly."|| p_wcname ||"\"]")

   IF l_nl.getLength() > 0 THEN
      LET l_ff = l_nl.item(1)
      LET l_wc = l_ff.getFirstChild()
      LET l_dict = l_wc.getFirstChild()
      IF l_dict IS NULL THEN
        LET l_dict = l_wc.createChild("PropertyDict")
        CALL l_dict.setAttribute("name", "properties")
      END IF
      LET l_nl = l_dict.selectByPath("//Property[@name=\"" || p_prop_name || "\"]")
      LET l_prop = l_nl.item(1)
      IF l_prop IS NULL THEN
         LET l_prop = l_dict.createChild("Property")
         CALL l_prop.setAttribute("name", p_prop_name)
      END IF
      CALL l_prop.setAttribute("value", p_prop_value)
   END IF
END FUNCTION

################################################################################
# Descriptions...: 取得指定分類下的流程圖
# Memo...........:
# Usage..........: CALL azzi000_03_get_flow(p_gzba001)
#                  RETURNING l_json
# Input parameter: p_gzba001      分類代碼
# Return code....: l_json         json格式的字串
# Date & Author..: 劉偉先
# Modify.........:
################################################################################
PUBLIC FUNCTION azzi000_03_get_flow(p_gzba001)
   DEFINE p_gzba001     LIKE gzba_t.gzba001
   DEFINE l_gzbb_t      RECORD LIKE gzbb_t.*
   DEFINE l_gzbbl_t     RECORD LIKE gzbbl_t.*
   DEFINE l_gzbc_t      RECORD LIKE gzbc_t.*
   DEFINE l_gzbd003     LIKE gzbd_t.gzbd003 #圖片名稱
   DEFINE l_flow        RECORD          #將轉成JSON格式的model
            mask        BOOLEAN,        #是否可操作
            containers    t_containers, #節點們
            connections   t_connections #連結線們
                        END RECORD
   DEFINE l_len_1       INTEGER
   DEFINE l_len_2       INTEGER
   DEFINE l_len_3       INTEGER
   DEFINE l_len_4       INTEGER
   DEFINE lr_json       STRING
   DEFINE l_opts        RECORD
            size           t_size,
            position       t_position,
            css            STRING
                        END RECORD
   DEFINE l_json        STRING         #l_flow轉換成json後的字串
   
   LET l_flow.mask = FALSE
   IF p_gzba001 == NULL THEN
      RETURN util.JSON.stringify(l_flow)
   END IF
   
      
   #  將分類下的節點及相關資訊取出(連結線、多語言資料)
   #     1.取出節點轉成l_flow.containers
   #     2.取出連結線轉成l_flow.connections
   #     3.
   
   OPEN declare_gzbb_by_gzbb002 USING p_gzba001
   FOREACH declare_gzbb_by_gzbb002 INTO l_gzbb_t.*
      #將節點轉換成預備的JSON model (有機會改用util.JSONObject替代？)
      CALL l_flow.containers.appendElement()
      LET l_len_1 = l_flow.containers.getLength()
      CALL util.JSON.parse(l_gzbb_t.gzbb004, l_opts)
      LET l_flow.containers[l_len_1].id      = l_gzbb_t.gzbb001
      LET l_flow.containers[l_len_1].module  = l_gzbb_t.gzbb003
      LET l_flow.containers[l_len_1].label   = ""
      LET l_flow.containers[l_len_1].desc    = ""
      LET l_flow.containers[l_len_1].template = l_gzbb_t.gzbb005
      
      LET l_gzbd003 = NULL
      SELECT gzbd003 INTO l_gzbd003 FROM gzbd_t 
         WHERE gzbd001 = l_gzbb_t.gzbb005
           AND gzbdent = g_enterprise
      LET l_flow.containers[l_len_1].image = l_gzbd003
      
      IF l_gzbb_t.gzbb003 == "program" THEN
         LET l_flow.containers[l_len_1].other = l_gzbb_t.gzbb006
         LET l_flow.containers[l_len_1].label = l_gzbb_t.gzbb006
         LET l_flow.containers[l_len_1].DESC  = cl_get_progname(l_gzbb_t.gzbb006,g_dlang,1), "(", l_gzbb_t.gzbb006 , ")"
      END IF

      IF l_gzbb_t.gzbb003 == "subflow" THEN
         INITIALIZE g_ref_fields TO NULL 
         LET g_ref_fields[1] = l_gzbb_t.gzbb006
         CALL ap_ref_array2(g_ref_fields,"SELECT gzbal003 FROM gzbal_t WHERE gzbalent='"||g_enterprise||"' AND gzbal001=? AND gzbal002='"||g_dlang||"'","") RETURNING g_rtn_fields
         LET l_flow.containers[l_len_1].other = l_gzbb_t.gzbb006
         LET l_flow.containers[l_len_1].label = g_rtn_fields[1]
      END IF
      
      LET l_flow.containers[l_len_1].size.*     = l_opts.size.*
      LET l_flow.containers[l_len_1].position.* = l_opts.position.*
      LET l_flow.containers[l_len_1].css        = l_opts.css
      
      #將連結線轉換成預備的JSON modle (有機會改用util.JSONObject替代？)     
      OPEN declare_gzbc_by_gzbb001 USING l_gzbb_t.gzbb001, p_gzba001
      FOREACH declare_gzbc_by_gzbb001 INTO l_gzbc_t.*
         CALL l_flow.connections.appendElement()
         LET l_len_4 = l_flow.connections.getLength()
         LET l_flow.connections[l_len_4].source = l_gzbc_t.gzbc001
         LET l_flow.connections[l_len_4].target = l_gzbc_t.gzbc002
         LET l_flow.connections[l_len_4].parameters.label = l_gzbc_t.gzbc004
      END FOREACH
      CLOSE declare_gzbc_by_gzbb001      
      
      OPEN declare_gzbbl_by_gzbb001 USING l_gzbb_t.gzbb001  
      FOREACH declare_gzbbl_by_gzbb001 INTO l_gzbbl_t.*
         #撈到與目前相同的語系時，賦值給節點的多語言label, desc
         IF l_gzbbl_t.gzbbl002 = g_lang THEN
            LET l_flow.containers[l_len_1].label = l_gzbbl_t.gzbbl003
            LET l_flow.containers[l_len_1].desc  = l_gzbbl_t.gzbbl004
            IF l_gzbb_t.gzbb003 = "program" THEN
               LET l_flow.containers[l_len_1].desc  = l_gzbbl_t.gzbbl004, "(", l_gzbbl_t.gzbbl003, ")"
            END IF
         END IF

      END FOREACH
      CLOSE declare_gzbbl_by_gzbb001
      
   END FOREACH 
   
   CLOSE declare_gzbb_by_gzbb002
   
   LET l_json = util.JSON.stringify(l_flow)
   RETURN l_json
END FUNCTION

PRIVATE FUNCTION azzi000_03_first_expand(p_gzba001)
   DEFINE p_gzba001  LIKE gzba_t.gzba001
   DEFINE l_cnt      INTEGER
   IF g_first_gzba001 IS NULL THEN
      SELECT COUNT(*) INTO l_cnt FROM gzbb_t WHERE gzbb002 = p_gzba001
      IF l_cnt > 0 THEN
         LET g_first_gzba001 = p_gzba001
      END IF
   END IF
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzi000_03_update_gzbc()
#                  RETURNING 回传参数
# Input parameter: none
# Return code....: none
# Date & Author..: 2024/12/31 20992
# Modify.........: 241231-00003#1 add
################################################################################
PRIVATE FUNCTION azzi000_03_update_gzbc()
   DEFINE l_gzbc        DYNAMIC ARRAY OF RECORD
         l_gzbc001     LIKE gzbc_t.gzbc001,
         l_gzbc002     LIKE gzbc_t.gzbc002,
         l_gzbc003     LIKE gzbc_t.gzbc003,
         l_gzbcent     LIKE gzbc_t.gzbcent,
         l_gzbaent     LIKE gzba_t.gzbaent
      END RECORD
   DEFINE l_sql         STRING
   DEFINE l_i           INTEGER
   DEFINE l_cn          INTEGER  #260427-00008#1 add
   
   #260428-00003#1 mark --- start ---
#   LET l_sql = " SELECT gzbaent, gzbcent, gzbc001, gzbc002, gzbc003 FROM gzbc_t ", 
#            " LEFT JOIN (SELECT gzbb001, gzbb002, gzbbent, gzba001, gzbaent FROM gzbb_t ", 
#            " LEFT JOIN gzba_t ON gzbb002 = gzba001 AND gzbaent = gzbbent ) tmpf ",
#            " ON gzbc003 = gzba001 AND gzbc001 = gzbb001 ",
#            " WHERE gzbaent IS NOT NULL AND gzbaent <> gzbcent AND gzbcent = '0' "
#   DECLARE azzi000_03_update_gzbc_cur CURSOR FROM l_sql
#   
#   LET l_sql = " UPDATE gzbc_t SET gzbcent=? WHERE gzbc001=? AND gzbc002=? AND gzbc003=?  "
#   PREPARE azzi000_03_update_gzbc_pre FROM l_sql
   #260428-00003#1 mark --- end ---
   
   #260427-00008#1 add --- start ---
   LET l_sql = " SELECT count(*) FROM gzbc_t where gzbcent=? and gzbc001=? AND gzbc002=? AND gzbc003=?  "
   PREPARE azzi000_03_sel_gzbc_pre FROM l_sql
   #260427-00008#1 add --- end ---

   #260428-00003#1 mark --- start ---
#   CALL l_gzbc.clear()
#   LET l_i = 1
#   FOREACH azzi000_03_update_gzbc_cur INTO l_gzbc[l_i].l_gzbaent,l_gzbc[l_i].l_gzbcent,l_gzbc[l_i].l_gzbc001,l_gzbc[l_i].l_gzbc002,l_gzbc[l_i].l_gzbc003
#      #EXECUTE azzi000_03_update_gzbc_pre USING l_gzbc[l_i].l_gzbaent,l_gzbc[l_i].l_gzbc001,l_gzbc[l_i].l_gzbc002,l_gzbc[l_i].l_gzbc003 #260427-00008#1 mark
#      
#      #260427-00008#1 add --- start ---
#      EXECUTE azzi000_03_sel_gzbc_pre INTO l_cn USING l_gzbc[l_i].l_gzbaent,l_gzbc[l_i].l_gzbc001,l_gzbc[l_i].l_gzbc002,l_gzbc[l_i].l_gzbc003
#      IF l_cn = 0 THEN
#         EXECUTE azzi000_03_update_gzbc_pre USING l_gzbc[l_i].l_gzbaent,l_gzbc[l_i].l_gzbc001,l_gzbc[l_i].l_gzbc002,l_gzbc[l_i].l_gzbc003
#      END IF
#      #260427-00008#1 add --- end ---
#      
#      LET l_i = l_i + 1
#   END FOREACH
   #260428-00003#1 mark --- end ---
   
   #250307-00032#1 add --- start ---
   #補上缺漏 ENT 部分，已有的部分要排除不需重複建立
   LET l_sql = " SELECT DISTINCT A.gzbc001, A.gzbc002, A.gzbc003 FROM gzba_t ",
            " LEFT JOIN gzbc_t A ON gzba001 = gzbc003 ",
            " LEFT JOIN gzbc_t B ON gzba001 = B.gzbc003 AND b.gzbcent = gzbaent ",
            " WHERE A.gzbc003 IS NOT NULL AND A.gzbcent <> gzbaent ",
            " AND b.gzbc003 IS NULL AND gzbaent = '", g_enterprise, "' "
   DECLARE azzi000_03_update_gzbc_cur2 CURSOR FROM l_sql
   
   LET l_sql = " INSERT INTO gzbc_t(gzbcent, gzbc001, gzbc002, gzbc003) VALUES (?, ?, ?, ?) "
   PREPARE azzi000_03_insert_gzbc_pre FROM l_sql

   CALL l_gzbc.clear()
   LET l_i = 1
   FOREACH azzi000_03_update_gzbc_cur2 INTO l_gzbc[l_i].l_gzbc001,l_gzbc[l_i].l_gzbc002,l_gzbc[l_i].l_gzbc003
      #EXECUTE azzi000_03_insert_gzbc_pre USING g_enterprise,l_gzbc[l_i].l_gzbc001,l_gzbc[l_i].l_gzbc002,l_gzbc[l_i].l_gzbc003 #260427-00008#1 mark
      
      #260427-00008#1 add --- start ---
      EXECUTE azzi000_03_sel_gzbc_pre INTO l_cn USING g_enterprise,l_gzbc[l_i].l_gzbc001,l_gzbc[l_i].l_gzbc002,l_gzbc[l_i].l_gzbc003
      IF l_cn = 0 THEN
         EXECUTE azzi000_03_insert_gzbc_pre USING g_enterprise,l_gzbc[l_i].l_gzbc001,l_gzbc[l_i].l_gzbc002,l_gzbc[l_i].l_gzbc003
      END IF
      #260427-00008#1 add --- end ---
      
      LET l_i = l_i + 1
   END FOREACH
   #250307-00032#1 add --- end ---

END FUNCTION

 
{</section>}
 
