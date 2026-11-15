#該程式未解開Section, 採用最新樣板產出!
#該程式為freestyle程式!
{<section id="cl_get.description" type="s" >}
#應用 a00 樣板自動產生(Version:3)
#+ Standard Version.....: SD版次:0047(1900-01-01 00:00:00), PR版次:0047(2025-06-19 15:29:25)
#+ Customerized Version.: SD版次:0000(1900-01-01 00:00:00), PR版次:0000(1900-01-01 00:00:00)
#+ Build......: 000739
#+ Filename...: cl_get
#+ Description: 抓取指定參考資料功能
#+ Creator....: 01258(2013-08-16 10:00:04)
#+ Modifier...: 00000 -SD/PR- 00824
 
{</section>}
 
{<section id="cl_get.global" type="s" >}
#應用 p00 樣板自動產生(Version:6)
#add-point:填寫註解說明 name="main.memo"
#               2015/12/29 by madey   增加函式cl_get_date_format()
#                                     增加函式cl_get_today_with_format()
#160118-00011#2 新增陣列存放取過的系統變數
#160426-00026#1 2016/04/26 by jrg542  增加取得fglrun -V 版本 lib
#161018-00027#1 2016/10/18 by jrg542  修正 抓取cl_get_para 參數資料 當是1:改抓設定時的預設值
#161109-00070#1 2016/11/10 by yscheng 新增cl_get_owndept_list()及cl_get_ownuser_list()
#161225-00003#1 2016/12/25 by jrg542  新增 cl_get_prog_code() 傳入作業編號 回傳程式編號
#170502-00044#1 2017/05/18 by jrg542  在參數配置表中新增一個欄位，標示這參數是否會時時變更，若為Y的話就不能放入緩存
#170525-00010#2 2017/05/25 by frank0521 修改COUNT(*) 改成 COUNT(1)
#170718-00057#1 2017/07/19 by yscheng 新增cl_get_exporttoexcel_type()，判斷匯出excel時要使用何種類型
#180117-00034#1 2018/01/19 by yscheng 調整cl_get_exporttoexcel_type()中客製否的判斷
#180417-00044#1 2018/04/17 by 01856   取得superuser role
#190314-00035#2 2019/03/15 by 04013   fglrun -V 判斷方式調整
#190902-00024#1 2019/09/04 by 01856   取參數cl_get_para()日期格式處理
#190926-00044#1 2019/09/30 by 01856   取參數cl_get_para()日期格式處理 yyyy/mm/dd dd/mm/yyyy
#191008-00018#1 2019/10/09 by 04013   調整cl_get_ap_hostname()取hostname判斷方式
#190918-00025#1 2019/10/15 by 01856   調整cl_get_today()日期格式處理
#191017-00034#2 2019/10/18 by 01856   新增cl_get_data_trans_yymmdd() 時間資料型態是char把日期類別轉成年月日和cl_get_data_trans_dbdate() 時間類別是char要轉成dbdate 日期格式
#191025-00018#1 2019/10/25 by 01856   修正譯錯誤訊息 incompatible types, found: 'base.StringTokenizer'
#201026-00009#2 2020/10/26 by 10167   組SQL前再次檢查，取不到值的時候也要讓指令對，不要變成 201 而留在 6372內
#201221-00029#1 2020/12/21 By 00845   調整語法
#201228-00059#1 2020/12/29 By 00845   取消判斷後生成PREPARE，PREPARE無法保留到下次調用function，必須每次重設
#200818-00050#1 2020/11/20 By 11234   效能优化，在進入cl_get_tableid()后初步判斷明顯不符合實体表字段的情況直接不計算實体表名
#210319-00004#1 2021/03/22 by 00845   效能優化，處理gzsz的暫存功能，以空間換取時間
#210603-00002#1 2021/06/03 by 00845   210319-00004變量使用太小,調整放大並設定暫存上限,避免超量使用時系統損毀
#211105-00021#1 2021/11/05 by 12731   XG 執行排程時會有紀錄到 cl_get 異常 ， 主要是 channel 開啟位置異常
#                                     補上 1 , 31 , 31p 程式同步部分 
#211125-00020#1 2021/12/15 by 02647   卡控次數限制 並設置debug log
#220125-00020#2 2022/02/17 by 00824   新增取得首頁配合報表相關資訊的FUNCTION (FOR 3.0首頁搭配億信ABI調整)
#231031-00024#1 2023/11/09 by 07556   偶發異常，調整cl_get_ap_hostname寫法。
#240216-00011#1 2024/02/16 by 00824   修正跨ENT處理時，重複透過cl_get_para()取同一個系統參數值時，會取到舊值的問題
#240312-00011#1 2024/03/12 by 20992   修正取企業級參數部分不應用ds 身份執行
#250606-00022#1 2025/06/19 By 00824   修正分管人權限問題
#end add-point
#add-point:填寫註解說明(客製用) name="main.memo_customerization"

#end add-point
#(ver:6) ---start---
#add-point:填寫註解說明(行業用) name="global.memo_industry"

#end add-point
#(ver:6) --- end ---
 
IMPORT os
#add-point:增加匯入項目 name="main.import"

#end add-point
 
SCHEMA ds
 
GLOBALS "../../cfg/top_global.inc"
#add-point:增加匯入變數檔 name="global.inc"

#end add-point
 
{</section>}
 
{<section id="cl_get.free_style_variable" type="s" >}
#add-point:free_style模組變數(Module Variable) name="free_style.variable"

#end add-point
 
{</section>}
 
{<section id="cl_get.global_variable" type="s" >}
#add-point:自定義模組變數(Module Variable) name="global.variable"
GLOBALS
   DEFINE gs_hostname    STRING
   #161109-00070 #1 ---start---
   DEFINE mc_gzyb003     LIKE gzyb_t.gzyb003   #部門資料授權 gzyb003,gzxb006
   DEFINE mc_gzyb004     LIKE gzyb_t.gzyb004   #個人資料授權 gzyb004,gzxb007
   DEFINE mc_gzyb005     LIKE gzyb_t.gzyb005   #單身處理方式 gzyb005,gzxb008
   DEFINE ma_site        DYNAMIC ARRAY OF LIKE gzxc_t.gzxc004  #可用據點陣列
   DEFINE ma_action      DYNAMIC ARRAY OF RECORD               #授權功能陣列
             gzyc003     LIKE gzyc_t.gzyc003,  #功能編號
             gzyc004     LIKE gzyc_t.gzyc004,  #功能授權
             gzyc005     LIKE gzyc_t.gzyc005   #資料授權類別
                         END RECORD
   DEFINE ma_deptgroup   DYNAMIC ARRAY OF RECORD
             gzxe002     LIKE gzxe_t.gzxe002,  #關聯部門編號
             gzxe005     LIKE gzxe_t.gzxe005   #是否允許下展
                         END RECORD
   DEFINE ma_dept_list   STRING                #相關部門暫存字串
   DEFINE mc_gzyb003_own_q      LIKE type_t.chr1   #查詢功能所屬部門資料授權
   DEFINE mc_gzyb003_own_m      LIKE type_t.chr1   #修改功能所屬部門資料授權
   DEFINE mc_gzyb003_own_d      LIKE type_t.chr1   #刪除功能所屬部門資料授權
   DEFINE mc_gzyb003_rela_q     LIKE type_t.chr1   #查詢功能相關部門資料授權
   DEFINE mc_gzyb003_rela_m     LIKE type_t.chr1   #修改功能相關部門資料授權
   DEFINE mc_gzyb003_rela_d     LIKE type_t.chr1   #刪除功能相關部門資料授權
   DEFINE mc_gzyb003_diff_m     LIKE type_t.chr1   #修改功能不同部門資料授權
   DEFINE mc_gzyb003_diff_q     LIKE type_t.chr1   #查詢功能不同部門資料授權
   DEFINE mc_gzyb003_diff_d     LIKE type_t.chr1   #刪除功能不同部門資料授權
   DEFINE mc_gzya003            LIKE gzya_t.gzya003 #是否為superuser
   #161109-00070 #1 --- end ---
   DEFINE mc_revalue  STRING   #是否重新取系統參數值   #240216-00011
   DEFINE g_responsibility STRING   #210223-00024#1 By 10167   #250606-00022
END GLOBALS
#210319-00004#1
#抓設定資料用的, 標示有沒有處理過這筆參數
DEFINE ma_para_def DYNAMIC ARRAY OF RECORD
          paracode STRING,
          gzsz001  LIKE gzsz_t.gzsz001,
          gzsz003  LIKE gzsz_t.gzsz003
               END RECORD
#緩存資料用的, 參數在 A(系統級),E(企業級)不受site影響,只有一筆,其他多筆               
DEFINE ma_para_set DYNAMIC ARRAY OF RECORD
          sitecode STRING,
          paracode STRING,
          value    STRING
               END RECORD
DEFINE gs_fglrun_ver STRING
DEFINE maw_get_tabname_open BOOLEAN  #201221-00029#1
DEFINE mas_sql_for_a  STRING  #201221-00029#1 系統級
DEFINE mas_sql_for_e  STRING  #201221-00029#1 企業級
DEFINE mas_sql_for_s  STRING  #201221-00029#1 據點級
#end add-point
 
{</section>}
 
{<section id="cl_get.other_dialog" type="s" >}

 
{</section>}
 
{<section id="cl_get.other_function" readonly="Y" type="s" >}
############################################################
#+ @code
#+ 函式目的 取出系統時間
#+
#+ @return  DATETIME YEAR  系統時間
############################################################
PUBLIC FUNCTION cl_get_syscurrent()
   DEFINE ld_local_time     DATETIME YEAR TO SECOND

   WHENEVER ERROR CALL cl_err_msg_log

   LET ld_local_time = cl_time_trans_by_tz("system_time")
   RETURN ld_local_time #系統時間 ex2012-09-10 15:13:53
END FUNCTION
############################################################
#+ @code
#+ 函式目的  取得table name
#+ @param    ps_tableid  CHAR(15)   欄位代碼
#+ @param    p_lang      NUMBER(5)  語言別
#+
#+ @return   CHAR(80)    表格名稱
############################################################
PUBLIC FUNCTION cl_get_table_name(ps_tableid,p_lang)
   DEFINE ps_tableid        LIKE dzeal_t.dzeal001
   DEFINE p_lang            LIKE dzeal_t.dzeal002
   DEFINE lc_tabname        LIKE type_t.chr80

   WHENEVER ERROR CALL cl_err_msg_log

   SELECT dzeal003 INTO lc_tabname FROM dzeal_t
    WHERE dzeal001 = ps_tableid AND dzeal002 = p_lang

   RETURN lc_tabname
END FUNCTION
############################################################
#+ @code
#+ 函式目的 擷取有設定關聯的reference節點
#+ @param   pnode_target  om.DomNode  欄位節點-FormField
#+ @param   ls_reference  STRING      reference欄位名稱
#+
#+ @return  om.DomNode    關聯欄位節點
############################################################
PUBLIC FUNCTION cl_get_reference(pnode_target,ls_reference)
   DEFINE   pnode_target om.DomNode
   DEFINE   ls_reference STRING
   DEFINE   lnode_next   om.DomNode
   DEFINE   ls_item_name STRING

   WHENEVER ERROR CALL cl_err_msg_log

   LET lnode_next = pnode_target.getNext()
   IF lnode_next IS NOT NULL THEN
      LET ls_item_name = lnode_next.getAttribute("colname")
      IF ls_item_name IS NULL THEN
         LET ls_item_name = lnode_next.getAttribute("name")
      END IF
   END IF
   IF (ls_item_name = ls_reference) OR
      (ls_item_name.subString(10,ls_item_name.getLength()) = ls_reference) OR
      (ls_item_name = ls_reference.subString(10,ls_reference.getLength())) THEN
      RETURN lnode_next
   ELSE
      RETURN NULL
   END IF
END FUNCTION
############################################################
#+ 查詢營運據點名稱
############################################################
PUBLIC FUNCTION cl_get_sitename(lc_ooag001)
   DEFINE lc_ooag001        LIKE ooag_t.ooag001
   DEFINE lc_ooag003        LIKE ooag_t.ooag003

   WHENEVER ERROR CALL cl_err_msg_log

   SELECT ooag003 INTO lc_ooag003 FROM ooag_t
    WHERE ooag001 = lc_ooag001 AND ooag002 = g_dlang AND ooagent = g_enterprise
   IF SQLCA.SQLCODE THEN
      LET lc_ooag003 = "User Not Exist"
   END IF

   RETURN lc_ooag003
END FUNCTION
############################################################
#+ @code
#+ 函式目的  尋找欄位在資料庫對應的table name
#+ @param    ps_fldname  STRING  欄位代碼
#+
#+ @return   STRING      表格代碼
############################################################
PUBLIC FUNCTION cl_get_tableid(ps_fldname)
   DEFINE ps_fldname        STRING
   DEFINE lc_tabname        LIKE type_t.chr30
   DEFINE ls_tabname        STRING
   DEFINE ls_sql            STRING
   DEFINE ps_fldname_char   LIKE gztz_t.gztz002
   DEFINE g_db_type         LIKE type_t.chr3    

   WHENEVER ERROR CALL cl_err_msg_log

   #200818-00050#1 add(s)
   IF ps_fldname MATCHES '*l_*' OR ps_fldname MATCHES 'formonly.*' THEN
      LET ls_tabname = NULL
      RETURN ls_tabname
   END IF
   #200818-00050#1 add(e)

   LET ps_fldname_char = ps_fldname

   #201221-00029#1 已經設定過的話可以複用
   IF maw_get_tabname_open IS NULL OR maw_get_tabname_open = FALSE THEN
      LET ls_sql = " SELECT gztz001 FROM gztz_t WHERE gztz002 = ? ",
                   "     AND gztz001 LIKE '%_t' "
      PREPARE get_tabname FROM ls_sql
      DECLARE get_tabname_curs CURSOR FOR get_tabname
      LET maw_get_tabname_open = TRUE
   END IF   
   FOREACH get_tabname_curs USING ps_fldname_char INTO lc_tabname
      IF NOT cl_null(lc_tabname) THEN
         EXIT FOREACH
      END IF
   END FOREACH

   IF g_db_type = "ORA" THEN
      LET ls_tabname = lc_tabname CLIPPED
      LET ls_tabname = ls_tabname.toLowerCase()
   ELSE
      LET ls_tabname = lc_tabname CLIPPED
   END IF

#   CLOSE get_tabname_curs
#201221-00029#1 有可能重複使用, 此處不移除
#   FREE get_tabname_curs
#   FREE get_tabname
   RETURN ls_tabname
END FUNCTION
############################################################
#+ @code
#+ 函式目的  尋找可登入系統的員工姓名
#+ @param    lc_gzxa003  CHAR(10)   使用者工號
#+
#+ @return   CHAR(80)    員工姓名
############################################################
PUBLIC FUNCTION cl_get_username(lc_gzxa003)
   DEFINE lc_gzxa003    LIKE gzxa_t.gzxa003
   DEFINE lc_ooag011    LIKE ooag_t.ooag011

   WHENEVER ERROR CALL cl_err_msg_log

   SELECT ooag011 INTO lc_ooag011 FROM ooag_t
    WHERE ooagent = g_enterprise
      AND ooag001 = lc_gzxa003
   IF SQLCA.SQLCODE THEN
      LET lc_ooag011 = cl_getmsg("sui-00005",g_lang)
   END IF  

   RETURN lc_ooag011
END FUNCTION
############################################################
#+ @code
#+ 函式目的 取出指定時區當下日期
#+
#+ @return  DATETIME YEAR  對方當地時間
############################################################
PUBLIC FUNCTION cl_get_today()
   #DEFINE ld_local_time     DATETIME YEAR TO DAY    #190918-00025#1  mark	
   DEFINE ld_local_time     DATE                     #190918-00025#1  add

   WHENEVER ERROR CALL cl_err_msg_log

   LET ld_local_time = cl_time_trans_by_tz("today_date")
   LET g_today = ld_local_time

   RETURN ld_local_time   #對方當地日期 #ex: 2012-09-10
END FUNCTION
############################################################
#+ @code
#+ 函式目的 取出指定時區當下時間
#+
#+ @return  DATETIME HOUR  對方當地時間
############################################################
PUBLIC FUNCTION cl_get_time()
   DEFINE ld_local_time     DATETIME HOUR TO SECOND

   WHENEVER ERROR CALL cl_err_msg_log

   LET ld_local_time = cl_time_trans_by_tz("today_time")
   RETURN ld_local_time   #對方當地時間 #ex: 15:13:53
END FUNCTION

################################################################################
# Descriptions...: 回傳毫秒
# Usage..........: cl_get_timestamp()
# Return code....: DATETIME YEAR TO FRACTION(5)
################################################################################
PUBLIC FUNCTION cl_get_timestamp()
   DEFINE ld_local_time     DATETIME YEAR TO FRACTION(5)

   WHENEVER ERROR CALL cl_err_msg_log

   LET ld_local_time = cl_time_trans_by_tz("current_time")
   RETURN ld_local_time   #對方當地時間 #ex: 2012-09-10 15:13:53
END FUNCTION
############################################################
#+ @code
#+ 函式目的  查詢指定語言別下的作業程式名稱
#+ @param    pc_gzzal001  CHAR(20)   程式代碼
#+ @param    pc_gzzal002  NUMBER(5)  語言別 (NULL時以g_lang取代)
#+ @param    pc_type     NUMBER(5)  回傳樣式 "1"_default單獨回傳  "2" 程式名稱 (id)
#+
#+ @return   CHAR(80)    指定語言別下的作業程式名稱
############################################################
PUBLIC FUNCTION cl_get_progname(pc_gzzal001,pc_gzzal002,pc_type)
   DEFINE pc_gzzal001  LIKE gzzal_t.gzzal001
   DEFINE pc_gzzal002  LIKE gzzal_t.gzzal002
   DEFINE rc_gzzal003  LIKE gzzal_t.gzzal003
   DEFINE rs_text      STRING
   DEFINE pc_type      LIKE type_t.num5
   DEFINE ls_sql       STRING

   WHENEVER ERROR CALL cl_err_msg_log

   LET pc_gzzal001 = pc_gzzal001 CLIPPED
   LET pc_gzzal002 = pc_gzzal002 CLIPPED

   #若傳入的語言別是空的 (未傳入),取用g_lang
   IF pc_gzzal002 IS NULL THEN
      LET pc_gzzal002 = g_lang
   END IF

   IF pc_type <> 2 THEN
      LET pc_type = 1
   END IF

   #取用名稱
   LET ls_sql = "SELECT gzzal003 FROM gzzal_t WHERE gzzal001 = ? AND gzzal002 = ? "
   PREPARE lcurs_gzza_pre FROM ls_sql
   EXECUTE lcurs_gzza_pre INTO rc_gzzal003 USING pc_gzzal001,pc_gzzal002

   IF pc_type = "1" THEN
      RETURN rc_gzzal003
   ELSE
      LET rs_text = rc_gzzal003 CLIPPED, " (" , pc_gzzal001 CLIPPED, ")"
      RETURN rs_text.trim()
   END IF
END FUNCTION
############################################################
#+ @code
#+ 函式目的  抓取AP Server的hostname(若不同ap server未依慣例而設定相同hostname時,將會造成系統誤判誤砍)
#+
#+ @return   STRING  AP_Server_Hostname
############################################################
PUBLIC FUNCTION cl_get_ap_hostname()
   DEFINE ls_host   STRING
   DEFINE ch        base.Channel
   DEFINE li_error  LIKE type_t.num5 #231031-00024#1 add

   WHENEVER ERROR CALL cl_err_msg_log


   #231031-00024#1 mark-s
   #LET ch = base.Channel.create()  #211105-00021#1 add
   #
   #IF cl_null(gs_hostname) THEN
   #   #LET ch = base.Channel.create()  #211105-00021#1 mark
   #   #CALL ch.setDelimiter(".")  #191008-00018 mark    原以"."做切分，調整為get full hostname ex:ecs-4a31.novalocal
   #   CALL ch.openPipe("hostname","r")
   #
   #   LET ls_host = ""
   #   WHILE ch.read(ls_host)
   #   END WHILE
   #ELSE
   #   LET ls_host = gs_hostname
   #END IF
   #
   #CALL ch.close()
   #RETURN ls_host.trim()
   #231031-00024#1 mark-e
   #231031-00024#1 add-s
   IF NOT cl_null(gs_hostname) THEN
      RETURN gs_hostname
   END IF

   TRY
      LET li_error = 0
      LET ch = base.Channel.create()
      LET li_error = 1
      LET ls_host = ""
      CALL ch.setDelimiter("")
      LET li_error = 2
      CALL ch.openPipe("hostname","r")
      LET li_error = 3
      WHILE ch.read(ls_host)
      END WHILE
      LET li_error = 4
      LET gs_hostname = ls_host.trim()
      CALL ch.close()
      LET li_error = 5
   CATCH
      DISPLAY "ERROR: cl_get_hostname error status:",status
      DISPLAY "ERROR: cl_get_hostname error code:",li_error
      RETURN "HOSTNAME_GET_FAIL"
   END TRY

   RETURN gs_hostname
   #231031-00024#1 add-e
END FUNCTION
############################################################
#+ @code
#+ 函式目的 取得要改變文字的節點
#+ @param   pnode_target  om.DomNode  Parent Node
#+
#+ @return  om.DomNode    lnode_pre
#+ @return  NUMBER(5)     li_bind_lable
############################################################
PUBLIC FUNCTION cl_get_change_text_node(pnode_target)
   DEFINE pnode_target    om.DomNode
   DEFINE lnode_pre       om.DomNode
   DEFINE lnode_child     om.DomNode
   DEFINE li_bind_label   LIKE type_t.num5
   DEFINE ls_tmp          STRING

   WHENEVER ERROR CALL cl_err_msg_log
   
   LET li_bind_label = FALSE
   IF pnode_target.getTagName() = "Button" THEN
      RETURN pnode_target, li_bind_label
   END IF

   LET lnode_pre = pnode_target.getPrevious()
   LET lnode_child = pnode_target.getFirstChild()

   IF lnode_child.getTagName() = "CheckBox" THEN
      RETURN lnode_pre, li_bind_label
   END IF

 --Getting related "Label" node for this element
   IF lnode_pre IS NOT NULL THEN
      IF ( lnode_pre.getTagName() = "Label" ) THEN
         LET ls_tmp = lnode_pre.getAttribute("name")
         LET ls_tmp = ls_tmp.toUpperCase()
         IF ls_tmp.getLength() <= 5 OR NOT ls_tmp.subString(1,5) = "DUMMY" THEN
            LET li_bind_label = TRUE
         END IF
      #for欄位前會有串聯程式功能的button處理
      ELSE
         IF lnode_pre.getTagName() = "Button" AND
            lnode_pre.getAttribute("tag") = "+" THEN
            LET lnode_pre = lnode_pre.getPrevious()
            IF lnode_pre IS NOT NULL THEN
               IF ( lnode_pre.getTagName() = "Label" ) AND
                  ( cl_null(lnode_pre.getAttribute("name")) ) THEN
                  LET li_bind_label = TRUE
               END IF
            END IF
         END IF
      END IF
   END IF

   RETURN lnode_pre, li_bind_label
END FUNCTION
############################################################
#+ @code
#+ 函式目的  尋找可登入系統的員工姓名
#+ @param    lc_gzxa001  CHAR(10)   登入帳號代碼
#+
#+ @return   CHAR(80)    員工姓名
############################################################
PUBLIC FUNCTION cl_get_accountname(lc_gzxa001)
   DEFINE lc_gzxa001        LIKE gzxa_t.gzxa001
   DEFINE lc_gzxa002        LIKE gzxa_t.gzxa002
   DEFINE lc_gzxa003        LIKE gzxa_t.gzxa003
   DEFINE lc_ooag011        LIKE ooag_t.ooag011

   WHENEVER ERROR CALL cl_err_msg_log

   SELECT gzxa002,gzxa003 INTO lc_gzxa002,lc_gzxa003 FROM gzxa_t
    WHERE gzxaent = g_enterprise
      AND gzxa001 = lc_gzxa001
   IF SQLCA.SQLCODE THEN
      LET lc_ooag011 = cl_getmsg("sui-00005",g_lang)
   ELSE
      #當gzxa002不為空時，則取作SCC，讀取聯絡對象檔ooag_t
      IF NOT cl_null(lc_gzxa003) THEN
         SELECT ooag011 INTO lc_ooag011 FROM ooag_t
          WHERE ooagent = g_enterprise
            AND ooag001 = lc_gzxa003
      END IF
   END IF

   RETURN lc_ooag011
END FUNCTION
############################################################
#+取得有效無效的圖片
############################################################
PUBLIC FUNCTION cl_get_actipic(lc_acticode)
   DEFINE lc_acticode   LIKE type_t.chr1
   DEFINE lc_picpath    LIKE type_t.chr50

   WHENEVER ERROR CALL cl_err_msg_log

   CASE lc_acticode
      WHEN "Y"
         LET lc_picpath = "stus/16/active.png"
      WHEN "N"
         LET lc_picpath = "stus/16/inactive.png"
   END CASE

   RETURN lc_picpath
END FUNCTION
############################################################
#+ @code
#+ 函式目的 將系統時間(CURRENT)轉換為指定時區時間
#+
#+ @return  DATETIME YEAR  對方當地時間
############################################################
PUBLIC FUNCTION cl_get_current()
   DEFINE ld_local_time     DATETIME YEAR TO SECOND

   WHENEVER ERROR CALL cl_err_msg_log

   LET ld_local_time = cl_time_trans_by_tz("current_time")
   LET g_today = ld_local_time

   RETURN ld_local_time   #對方當地時間 #ex: 2012-09-10 15:13:53
END FUNCTION
############################################################
#+ @code
#+ 函式目的  尋找部門名稱
#+ @param   lc_ooefl001  CHAR(10)  部門代碼
#+
#+ @return   CHAR(80)    部門名稱
############################################################
PUBLIC FUNCTION cl_get_deptname(lc_ooefl001)
   DEFINE lc_ooefl001    LIKE ooefl_t.ooefl001 
   DEFINE lc_ooefl002    LIKE ooefl_t.ooefl002 
   DEFINE lc_ooefl003    LIKE ooefl_t.ooefl003 

   WHENEVER ERROR CALL cl_err_msg_log

   SELECT ooefl003 INTO lc_ooefl003 FROM ooefl_t
    WHERE ooeflent = g_enterprise
      AND ooefl001= lc_ooefl001 AND ooefl002 = g_lang
   IF SQLCA.SQLCODE THEN
      LET lc_ooefl003 = "Dept is Not Exist"
   END IF
   RETURN lc_ooefl003
END FUNCTION
############################################################
#+ @code
#+ 函式目的  找出該 4gl 所屬的 Module 名稱
#+ @param    lc_gzza001  CHAR(20)  4gl 檔名
#+ @param    lc_type     CHAR(10)  回傳形態 S:大寫英文 D:小寫英文 0,1,2:語言別
#+
#+ @return   STRING      所需的模組名稱
############################################################
PUBLIC FUNCTION cl_get_module(lc_gzza001,lc_type)
   DEFINE lc_gzza001   LIKE gzza_t.gzza001
   DEFINE lc_gzza011   LIKE gzza_t.gzza011   #客製
   DEFINE lc_gzza003   LIKE gzza_t.gzza003
   DEFINE lc_return    LIKE gzza_t.gzza003   #回傳用值
   DEFINE lc_type      LIKE gzzol_t.gzzol002
   DEFINE lc_gzzol003  LIKE gzzol_t.gzzol003

   WHENEVER ERROR CALL cl_err_msg_log

   # gzzol_t 有錯或查不到 gzzol003 資料則回傳空值
   SELECT gzza003,gzza011 INTO lc_gzza003,lc_gzza011 FROM gzza_t 
    WHERE gzza001 = lc_gzza001
   IF STATUS THEN
      RETURN " "
   ELSE
      LET lc_return = lc_gzza003 CLIPPED
   END IF

   #A/B 要因為勾選了客製, 更換編號為 C/D
   IF lc_gzza011 = "c" THEN
      IF lc_return[1,1] = "A" THEN LET lc_return = "C",lc_return[2,3] END IF
      IF lc_return[1,1] = "B" THEN LET lc_return = "D",lc_return[2,3] END IF
   END IF      
 
   CASE
     WHEN lc_type = "S" OR lc_type = "s"
        LET lc_gzza003 = UPSHIFT(lc_return) CLIPPED
        RETURN lc_gzza003
     WHEN lc_type = "D" OR lc_type = "d"
        LET lc_gzza003 = DOWNSHIFT(lc_return) CLIPPED
        RETURN lc_gzza003
     OTHERWISE
        #因為C/D開頭沒有名字, 所以要轉回A/B
        IF lc_gzza011 = "c" THEN
           IF lc_return[1,1] = "C" THEN LET lc_return = "A",lc_return[2,3] END IF
           IF lc_return[1,1] = "D" THEN LET lc_return = "B",lc_return[2,3] END IF
        END IF 
        SELECT gzzol003 INTO lc_gzzol003 FROM gzzol_t
         WHERE gzzol001 = lc_gzza003 AND gzzol002 = lc_type
        RETURN lc_gzzol003
   END CASE
END FUNCTION
############################################################
#+ @code
#+ 函式目的  依欄位代碼取出系統欄位名稱設定維護作業(p_feldname)內的欄位名稱
#+ @param    p_gztu001  CHAR(20)   欄位代碼
#+ @param    p_lang     NUMBER(5)  語言別
#+
#+ @return   CHAR(80)   欄位說明
############################################################
PUBLIC FUNCTION cl_get_feldname(p_dzebl001,p_lang)
   DEFINE p_dzebl001     LIKE dzebl_t.dzebl001
   DEFINE p_lang         LIKE dzebl_t.dzebl002
   DEFINE l_dzebl003     LIKE dzebl_t.dzebl003
 
   WHENEVER ERROR CALL cl_err_msg_log

   LET l_dzebl003 = ''
   SELECT dzebl003 INTO l_dzebl003 FROM dzebl_t
    WHERE dzebl001 = p_dzebl001 AND dzebl002 = p_lang

   RETURN l_dzebl003
END FUNCTION
############################################################
#+ 抓取參照表參數
############################################################
PUBLIC FUNCTION cl_get_ref_para(lc_enterprise,lc_referencetable,lc_para_id)
   DEFINE lc_enterprise   LIKE type_t.num5
   DEFINE lc_para_data    LIKE type_t.chr80
   DEFINE lc_para_id      LIKE gzsz_t.gzsz002
   DEFINE lc_referencetable LIKE type_t.chr10   
   
   RETURN lc_para_data
END FUNCTION
############################################################
#+ 查詢設定資料
############################################################
PUBLIC FUNCTION cl_get_para(lc_enterprise,lc_site,lc_para_id)
   DEFINE lc_enterprise   LIKE type_t.num5
   DEFINE lc_site         LIKE type_t.chr10
   DEFINE lc_para_id      LIKE gzsz_t.gzsz002 #參數編號
   DEFINE lc_gzsz001      LIKE gzsz_t.gzsz001 #參數所屬table
   DEFINE lc_gzsz003      LIKE gzsz_t.gzsz003 #輸入限制資料型態   #190902-00024#1    
   DEFINE lc_para_data    LIKE type_t.chr80
   DEFINE lc_type         LIKE type_t.chr1    #參數型態,取參數編號第1碼
   DEFINE li_cnt          LIKE type_t.num5
   DEFINE ls_sql          STRING
   DEFINE ls_tableid      STRING
   DEFINE lc_gzsz008      LIKE gzsz_t.gzsz008 #參數預設值
   DEFINE lc_gzsz017      LIKE gzsz_t.gzsz017 #抓取異常值處理方法 1:參數未設置時取系統預設值 2:顯示Error 3:回傳Null
   DEFINE lc_gzsz019      LIKE gzsz_t.gzsz019 #允許參數即時抓取   #170502-00044
   DEFINE li_row          LIKE type_t.num10   #需用到 NUM10
   DEFINE ls_sqlcode      STRING
   DEFINE ls_gzou003      LIKE gzou_t.gzou003 #240312-00011#1 add
   
   WHENEVER ERROR CALL cl_err_msg_log

   #空參數處理
   IF lc_para_id IS NULL THEN
      RETURN lc_para_data
   END IF

   IF cl_null(mc_revalue) OR mc_revalue = 'N' THEN   #240216-00011
      #160118-00011 - start -
      IF ma_para_set.getLength() > 0 THEN
         CASE
            WHEN lc_para_id[1,1] = "E" OR lc_para_id[1,1] = "A"
               FOR li_row = 1 TO ma_para_set.getLength()
                  IF lc_para_id = ma_para_set[li_row].paracode THEN
                     &ifdef v1
                     CALL cl_get_para_date(lc_enterprise,lc_site,lc_para_id,
                                           ma_para_set[li_row].value) 
                     &endif
                     &ifdef v3
                     CALL cl_get_para_date(lc_enterprise,lc_site,
                                           lc_para_id,ma_para_set[li_row].value)
                     &endif                                        
                        RETURNING ma_para_set[li_row].value     #190902-00024#1
                     RETURN ma_para_set[li_row].value               
                  END IF
               END FOR
            OTHERWISE
               FOR li_row = 1 TO ma_para_set.getLength()
                  IF lc_site = ma_para_set[li_row].sitecode AND lc_para_id = ma_para_set[li_row].paracode THEN
                     CALL cl_get_para_date(lc_enterprise,lc_site,lc_para_id,ma_para_set[li_row].value) 
                        RETURNING ma_para_set[li_row].value     #190902-00024#1  
                     RETURN ma_para_set[li_row].value                                    
                  END IF
               END FOR
         END CASE
      END IF
      #160118-00011 -  end  -
   END IF   #240216-00011
   
   #取出參數型態
   LET lc_type = lc_para_id[1,1]

   #--frank
   #找出該參數的所屬table及預設值

   SELECT gzsz001,gzsz003,gzsz008,gzsz017,gzsz019 
     INTO lc_gzsz001,lc_gzsz003,lc_gzsz008,lc_gzsz017,lc_gzsz019     #170502-00044    #190902-00024#1   add gzsz003
     FROM gzsz_t                                                                               #170502-00044
    WHERE gzsz002 = lc_para_id
   IF STATUS = 100 THEN
     RETURN lc_para_data
   END IF

   #210319-00004#1
   #寫入ma_para_def #前面不檢的原因在於一定會走過這一段,後面也不檢的原因在於這一段只會走一次(靠ma_para_set)
   LET li_row = ma_para_def.getLength()+1
   IF li_row < 50000 THEN           #210603-00002 為避免內存爆量, 超過50000筆後，停止暫存    #211125-00020#1修正大於不寫入
     IF g_t100debug >= 6 THEN display 'AW-DEF:pa-code:',lc_para_id,' 001:',lc_gzsz001,' 003:',lc_gzsz003 END IF #211125-00020#1 t100debug log紀錄
      LET ma_para_def[li_row].paracode = lc_para_id
      LET ma_para_def[li_row].gzsz001 = lc_gzsz001
      LET ma_para_def[li_row].gzsz003 = lc_gzsz003
   END IF

   #組合出需要的SQL敘述
   LET ls_tableid = lc_gzsz001 CLIPPED
   LET ls_tableid = ls_tableid.subString(1,ls_tableid.getIndexOf("_t",1)-1)
   
   #201026-00009#2 by 10167 -s-
   #組SQL前再次檢查，取不到值的時候也要讓指令對，不要變成 201 而留在 6372內
   &ifdef v1
   CASE lc_para_id[1,1]
      WHEN "A" #假定為系統級的
         IF ls_tableid.getLength() <= 3 THEN
            LET ls_tableid = "gzsa"
            LET lc_gzsz001 = "gzsa_t"
         END IF
         
      WHEN "E" #假定為企業級的
         IF ls_tableid.getLength() <= 3 THEN
            LET ls_tableid = "ooaa"
            LET lc_gzsz001 = "ooaa_t"
         END IF
         
      WHEN "S" #假定為據點級的
         IF ls_tableid.getLength() <= 3 THEN
            LET ls_tableid = "ooab"
            LET lc_gzsz001 = "ooab_t"
         END IF
         
   END CASE
   &endif
   &ifdef v3
   CASE
      WHEN lc_type = "A" #假定為系統級的
         IF ls_tableid.getLength() <= 3 THEN
            LET ls_tableid = "gzsa"
            LET lc_gzsz001 = "gzsa_t"
         END IF
         LET ls_sql = "SELECT ",ls_tableid,"002 FROM ",lc_gzsz001 CLIPPED,
                      " WHERE ",ls_tableid,"001 = ? "

         PREPARE cl_get_para_cs_a FROM ls_sql   #201228-00059#1
         EXECUTE cl_get_para_cs_a USING lc_para_id INTO lc_para_data
         LET ls_sqlcode = SQLCA.SQLCODE
         FREE cl_get_para_cs_a
         
      WHEN lc_type = "E" #假定為企業級的
         IF ls_tableid.getLength() <= 3 THEN
            LET ls_tableid = "ooaa"
            LET lc_gzsz001 = "ooaa_t"
         END IF
         #240312-00011#1 add --- start ---
         LET ls_gzou003 = g_dbs
         IF g_dbs = "ds" THEN
            SELECT gzou003 INTO ls_gzou003 FROM gzou_t WHERE gzou001 = lc_enterprise
            IF cl_null(ls_gzou003) THEN 
               LET ls_gzou003 = g_dbs
            END IF
         END IF
         #240312-00011#1 add --- end ---
         #LET ls_sql = "SELECT ",ls_tableid,"002 FROM ",g_dbs CLIPPED,".",lc_gzsz001 CLIPPED, #240312-00011#1 mark
         LET ls_sql = "SELECT ",ls_tableid,"002 FROM ",ls_gzou003 CLIPPED,".",lc_gzsz001 CLIPPED, #240312-00011#1 add
                      " WHERE ",ls_tableid,"001 = ? ",
                        " AND ",ls_tableid,"ent = ? "

         PREPARE cl_get_para_cs_e FROM ls_sql   #201228-00059#1
         EXECUTE cl_get_para_cs_e USING lc_para_id,lc_enterprise INTO lc_para_data
         FREE cl_get_para_cs_e
         LET ls_sqlcode = SQLCA.SQLCODE

      WHEN lc_type = "S" #假定為據點級的   
         IF ls_tableid.getLength() <= 3 THEN
            LET ls_tableid = "ooab"
            LET lc_gzsz001 = "ooab_t"
         END IF
         LET ls_sql = "SELECT ",ls_tableid,"002 FROM ",g_dbs CLIPPED,".",lc_gzsz001 CLIPPED,
                      " WHERE ",ls_tableid,"001 = ? ",
                        " AND ",ls_tableid,"ent = ? ",
                        " AND ",ls_tableid,"site = ? "

         PREPARE cl_get_para_cs_s FROM ls_sql   #201228-00059#1
         EXECUTE cl_get_para_cs_s USING lc_para_id,lc_enterprise,lc_site INTO lc_para_data
         LET ls_sqlcode = SQLCA.SQLCODE
         FREE cl_get_para_cs_s

      OTHERWISE
          DISPLAY "ERROR : Setting Error, check azzi990 for ",lc_para_id
          RETURN ""
   END CASE
   &endif

   &ifdef v1
   #調整寫法，table前加上DB name
   CASE
      WHEN lc_type = "A" #A:系統級
         LET ls_sql = "SELECT ",ls_tableid,"002 FROM ",lc_gzsz001 CLIPPED,
                      " WHERE ",ls_tableid,"001 = '",lc_para_id CLIPPED,"' "
      WHEN lc_type = "E" #E:企業級
         #240312-00011#1 add --- start ---
         LET ls_gzou003 = g_dbs
         IF g_dbs = "ds" THEN
            SELECT gzou003 INTO ls_gzou003 FROM gzou_t WHERE gzou001 = lc_enterprise
            IF cl_null(ls_gzou003) THEN 
               LET ls_gzou003 = g_dbs
            END IF
         END IF
         #240312-00011#1 add --- end ---
         #LET ls_sql = "SELECT ",ls_tableid,"002 FROM ",g_dbs CLIPPED,".",lc_gzsz001 CLIPPED, #240312-00011#1 mark
         LET ls_sql = "SELECT ",ls_tableid,"002 FROM ",ls_gzou003 CLIPPED,".",lc_gzsz001 CLIPPED, #240312-00011#1 add
                      " WHERE ",ls_tableid,"001 = '",lc_para_id CLIPPED,"' ",
                        " AND ",ls_tableid,"ent = ",lc_enterprise USING "<<<<<"
      WHEN lc_type = "S" #S:據點級
         LET ls_sql = "SELECT ",ls_tableid,"002 FROM ",g_dbs CLIPPED,".",lc_gzsz001 CLIPPED,
                      " WHERE ",ls_tableid,"001 = '",lc_para_id CLIPPED,"' ",
                        " AND ",ls_tableid,"ent = ",lc_enterprise USING "<<<<<",
                        " AND ",ls_tableid,"site = '",lc_site CLIPPED,"' "
      OTHERWISE
   END CASE
   

   #準備,執行
   PREPARE cl_get_para_cs FROM ls_sql
   EXECUTE cl_get_para_cs INTO lc_para_data
   
   IF SQLCA.SQLCODE THEN
   &endif
   &ifdef v3
   #準備,執行
   IF ls_sqlcode THEN
   &endif
      CASE
         WHEN lc_gzsz017 = "1"    #1:參數未設置時取系統預設值
         #  IF SQLCA.SQLCODE = 100 THEN
               LET lc_para_data = lc_gzsz008 CLIPPED
         #  END IF

         WHEN lc_gzsz017 = "2"    #2:顯示Error
            IF g_bgjob = "N" AND g_dbs IS NOT NULL THEN
               LET lc_para_data = NULL

               INITIALIZE g_errparam TO NULL
               LET g_errparam.code = SQLCA.SQLCODE
               LET g_errparam.extend = lc_para_id
               LET g_errparam.popup = TRUE
               CALL cl_err()
            END IF

         WHEN lc_gzsz017 = "3"    #3:回傳Null
            LET lc_para_data = NULL
      END CASE
   ELSE
   #參數值檢查 (如果設定成NOT_DEFINED，則顯示錯誤提示訊息)
      IF lc_para_data = "NOT_DEFINED" OR lc_para_data IS NULL THEN
         #有資料 可能是NOT_DEFINED 或是 NULL
         IF g_bgjob = "N" AND g_dbs IS NOT NULL THEN
            CASE

               WHEN lc_gzsz017 = "1" #1:改抓設定時的預設值  #161018-00027 #1
                  LET lc_para_data = lc_gzsz008 #給預設值  #161018-00027 #1  

               WHEN lc_gzsz017 = "2"   #2:顯示Error
                  LET lc_para_data = NULL

                  INITIALIZE g_errparam TO NULL
                  LET g_errparam.code = "lib-00086"
                  LET g_errparam.extend = NULL
                  LET g_errparam.popup = TRUE
                  LET g_errparam.replace[1] = lc_para_id CLIPPED
                  CALL cl_err()

               WHEN lc_gzsz017 = "3"   #3:回傳Null
                  LET lc_para_data = NULL
            END CASE
         END IF
      END IF
   END IF
   &ifdef v1
   CLOSE cl_get_para_cs
   FREE cl_get_para_cs
   &endif

   #160118-00011 - start -
   
   #170502-00044 start
   #回寫陣列--設定gzsz019=Y 要即時搜尋的,不寫入陣列
   IF lc_gzsz019 IS NOT NULL THEN
      IF lc_gzsz019 = "Y" THEN
         
         CALL cl_get_para_date(lc_enterprise,lc_site,lc_para_id,lc_para_data) RETURNING lc_para_data   #190902-00024#1
         RETURN lc_para_data

      END IF
   END IF
   #170502-00044 end
   #191017-00034#2--start--
   IF lc_gzsz019 = "N" OR cl_null(lc_gzsz019) THEN
      CALL cl_get_para_date(lc_enterprise,lc_site,lc_para_id,lc_para_data) RETURNING lc_para_data   #191017-00034 #2
   END IF 
   #191017-00034#2--end--
   
   #回寫陣列
   LET li_row = ma_para_set.getLength() + 1
   IF li_row < 50000 THEN           #210603-00002 為避免內存爆量, 超過50000筆後，停止暫存  #211125-00020#1修正大於不寫入
     IF g_t100debug >= 6 THEN display 'AW-SET:pa-code:',lc_para_id,' site:',lc_site,' value:',lc_para_data END IF   #211125-00020#1 t100 debug log紀錄
      LET ma_para_set[li_row].sitecode = lc_site
      LET ma_para_set[li_row].paracode = lc_para_id
      LET ma_para_set[li_row].value    = lc_para_data
   END IF
   #160118-00011 -  end  -

   RETURN lc_para_data
END FUNCTION
############################################################
#+ 抓取單據別參數
############################################################
PUBLIC FUNCTION cl_get_doc_para(lc_enterprise,lc_site,lc_documenttype,lc_para_id)
   DEFINE lc_enterprise     LIKE type_t.num5
   DEFINE lc_site           LIKE type_t.chr10   
   DEFINE lc_para_data      LIKE type_t.chr80
   DEFINE lc_para_id        LIKE gzsz_t.gzsz002
   DEFINE lc_gzsz001        LIKE gzsz_t.gzsz001
   DEFINE lc_gzsz003        LIKE gzsz_t.gzsz003   #210319-00004#1
   DEFINE lc_referencetable LIKE type_t.chr10
   DEFINE lc_documenttype   LIKE type_t.chr10
   DEFINE li_cnt            LIKE type_t.num5
   DEFINE ls_sql            STRING
   DEFINE ls_tableid        STRING   
   DEFINE ls_chk_ent_sql    STRING
   DEFINE li_ma_para_def    BOOLEAN

   WHENEVER ERROR CALL cl_err_msg_log

   #找出參照表號
   SELECT ooef004 INTO lc_referencetable FROM ooef_t 
    WHERE ooefent = lc_enterprise 
      AND ooef001 = lc_site

   #找出該參數的所屬table  ==>doc_只做定義緩存,沒有做資料緩存
   #210319-00004#1
   LET li_ma_para_def = FALSE
   #檢測暫存陣列
   FOR li_cnt = 1 TO ma_para_def.getLength()     #211125-00020  _set -> _def
      IF lc_para_id = ma_para_def[li_cnt].paracode THEN
         LET lc_gzsz001 = ma_para_def[li_cnt].gzsz001
         LET lc_gzsz003 = ma_para_def[li_cnt].gzsz003
         LET li_ma_para_def = TRUE
         EXIT FOR
      END IF
   END FOR
   #不存在只好查db後放入
   IF NOT li_ma_para_def THEN
      SELECT gzsz001,gzsz003 INTO lc_gzsz001,lc_gzsz003 FROM gzsz_t
       WHERE gzsz002 = lc_para_id
      IF NOT SQLCA.SQLCODE THEN
         LET li_cnt = ma_para_def.getLength()+1
         IF li_cnt < 50000 THEN  #211125-00020
           IF g_t100debug >= 6 THEN display 'AW-DEF:pa-code:',lc_para_id,' 001:',lc_gzsz001,' 003:',lc_gzsz003 END IF #211125-00020
           LET ma_para_def[li_cnt].paracode = lc_para_id
           LET ma_para_def[li_cnt].gzsz001 = lc_gzsz001
           LET ma_para_def[li_cnt].gzsz003 = lc_gzsz003
         END IF
      ELSE
         IF SQLCA.SQLCODE = 100 THEN
            RETURN lc_para_data
         END IF
      END IF
   END IF
   
   #201026-00009#2 by 10167 -s-
   #增加一組錯誤判斷
   IF SQLCA.SQLCODE THEN
      DISPLAY "ERROR:參數無法被取出 SQLCODE:",SQLCA.SQLCODE," MESSAGE:",SQLERRMESSAGE
      RETURN lc_para_data
   END IF
   #201026-00009#2 by 10167 -e-

   #組合出需要的SQL敘述
   LET ls_tableid = lc_gzsz001 CLIPPED
   LET ls_tableid = ls_tableid.subString(1,ls_tableid.getIndexOf("_t",1)-1)
   LET ls_sql = "SELECT ",ls_tableid,"004 FROM ",lc_gzsz001 CLIPPED,
                " WHERE ",ls_tableid,"001 = '",lc_referencetable CLIPPED,"' ",
                  " AND ",ls_tableid,"002 = '",lc_documenttype CLIPPED,"' ",
                  " AND ",ls_tableid,"003 = '",lc_para_id CLIPPED,"' "

   #確認該單據別table是否有ent欄位
   LET li_cnt = 0
   LET ls_chk_ent_sql = "SELECT COUNT(1) FROM gztz_t ",   #170525-00010#2
                        " WHERE gztz001 = '",ls_tableid,"_t'",
                          " AND gztz002 = '",ls_tableid,"ent'"

   PREPARE cl_get_doc_para_chk_ent_cs FROM ls_chk_ent_sql
   EXECUTE cl_get_doc_para_chk_ent_cs INTO li_cnt
   IF li_cnt > 0 THEN
      LET ls_sql = ls_sql, " AND ",ls_tableid,"ent = '",g_enterprise CLIPPED,"'"
   END IF

   FREE cl_get_doc_para_chk_ent_cs

   #準備,執行
   PREPARE cl_get_doc_para_cs FROM ls_sql
   EXECUTE cl_get_doc_para_cs INTO lc_para_data

   FREE cl_get_doc_para_cs

   RETURN lc_para_data
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
PUBLIC FUNCTION cl_get_module_by_4gl(ls_checking,lc_type)
   DEFINE ls_checking  STRING
   DEFINE lc_gzzo001   LIKE gzzo_t.gzzo001
   DEFINE lc_type      LIKE gzzol_t.gzzol002
   DEFINE lc_gzzol003  LIKE gzzol_t.gzzol003
   DEFINE li_n,li_error LIKE type_t.num5
   DEFINE ls_return   STRING

   WHENEVER ERROR CALL cl_err_msg_log

   CASE
      # 1.當 frm_name 以 a 開頭時, 必為標準 package
      WHEN ls_checking.subString(1,1)="a"
         FOR li_n=3 TO ls_checking.getLength()
            LET lc_gzzo001 = ls_checking.subString(1,li_n)
            LET lc_gzzo001 = UPSHIFT(lc_gzzo001 CLIPPED)
            SELECT gzzo001 FROM gzzo_t WHERE gzzo001 = lc_gzzo001
            IF NOT SQLCA.SQLCODE THEN
               LET li_error=FALSE
               EXIT FOR
            ELSE
               LET li_error=TRUE
            END IF
         END FOR

      # 5.當 frm_name 以 s_ 開頭時, 必為 sub
      WHEN ls_checking.subString(1,2)="s_"
         LET lc_gzzo001 = "sub"

      # 5.當 frm_name 以 cl_ 開頭時, 必為 lib
      WHEN ls_checking.subString(1,3)="cl_"
         LET lc_gzzo001 = "lib"

      # 6.當 frm_name 以 q_ 開頭時, 必為 qry
      WHEN ls_checking.subString(1,2)="q_"
         LET lc_gzzo001 = "qry"

      # 6.當 frm_name 以 n_ 開頭時, 必為 lng
      WHEN ls_checking.subString(1,2)="n_"
         LET lc_gzzo001 = "lng"

      OTHERWISE
         LET li_error=TRUE
   END CASE
   CASE
     WHEN lc_type = "S" OR lc_type = "s"
        LET ls_return = UPSHIFT(lc_gzzo001) CLIPPED
     WHEN lc_type = "D" OR lc_type = "d"
        LET ls_return = DOWNSHIFT(lc_gzzo001) CLIPPED
     OTHERWISE
        SELECT gzzol003 INTO lc_gzzol003 FROM gzzol_t
         WHERE gzzol003 = lc_gzzo001 AND gzzol002 = lc_type
        LET ls_return = lc_gzzol003
   END CASE
   RETURN ls_return
END FUNCTION

################################################################################
# Descriptions...: 取得程式的行業別代碼
# Memo...........:
# Usage..........: cl_get_industry_code()
#                  RETURNING p_industry
# Input parameter: 無
# Return code....: p_industry     行業別代碼
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_get_industry_code()
   DEFINE p_industry   STRING
   DEFINE ls_str       STRING


   LET ls_str = g_code
   IF ls_str.getIndexOf("_",1) > 0 THEN
      LET p_industry = ls_str.subString(ls_str.getIndexOf("_",1)+1,ls_str.getLength())
   ELSE
      LET p_industry = "sd"   #一般行業
   END IF

   RETURN p_industry
END FUNCTION

################################################################################
# Descriptions...: 取得單身的每頁顯示筆數
# Memo...........:
# Usage..........: CALL cl_get_num_in_page()
#                  RETURNING ls_num_in_page
# Return code....: ls_num_in_page 單身每頁顯示筆數
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_get_num_in_page()
   DEFINE ls_gzza005         LIKE gzza_t.gzza005    #單身顯示筆數處理方法
   DEFINE ls_gzza007         LIKE gzza_t.gzza007    #最大單身顯示筆數
   DEFINE ls_num_in_page     LIKE type_t.num10      #單身每頁顯示筆數


   LET ls_gzza005 = NULL
   LET ls_gzza007 = 0
   LET ls_num_in_page = 0

   SELECT gzza005,gzza007 INTO ls_gzza005,ls_gzza007 FROM gzza_t
    WHERE gzza001 = g_prog

   IF NOT cl_null(ls_gzza005) THEN
      # 若單身顯示筆數處理方式設定為 "3 依azzi900設定管制"，但卻又未設定顯示筆數，
      # 則改為依系統參數設定的顯示筆數
      IF ls_gzza005 = "3" AND (cl_null(ls_gzza007) OR ls_gzza007 < 1) THEN
      #   INITIALIZE g_errparam TO NULL
      #   LET g_errparam.extend = ""
      #   LET g_errparam.code   = "lib-00249"
      #   LET g_errparam.popup  = TRUE
      #   CALL cl_err()

         LET ls_gzza005 = "2"
      END IF

      CASE
         WHEN ls_gzza005 = "1"   #不管制
            LET ls_num_in_page = 2000000000   #20億筆

         WHEN ls_gzza005 = "2"   #依系統參數設定
            #取aoos010中設定的 "查詢作業單身每頁資料顯示筆數"
            CALL cl_get_para(g_enterprise,"","E-SYS-0710") RETURNING ls_num_in_page

         WHEN ls_gzza005 = "3"   #依azzi900設定管制
            LET ls_num_in_page = ls_gzza007

      END CASE
     
   END IF

   IF cl_null(ls_num_in_page) OR ls_num_in_page < 1 THEN
   #   INITIALIZE g_errparam TO NULL
   #   LET g_errparam.extend = ""
   #   LET g_errparam.code   = "lib-00250"
   #   LET g_errparam.popup  = TRUE
   #   CALL cl_err()

      LET ls_num_in_page = g_max_rec
   END IF

   RETURN ls_num_in_page

END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_get_date_format()
#                  RETURNING 回传参数
# Input parameter: 
#                : 
# Return code....: 回传参数变量1   YYYY/MM/DD
#                : 
# Date & Author..: 2015/12/29 By madey
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_get_date_format()
   DEFINE ls_dbdate    STRING
   DEFINE lc_time_unit LIKE type_t.chr1
   DEFINE lc_date      LIKE type_t.chr4
   DEFINE ls_date      STRING
   DEFINE li_i         LIKE type_t.num5

   #ex DBDATE=Y4MD/時，回傳YYYY/MM/DD
   CALL FGL_GETENV("DBDATE") RETURNING ls_dbdate

   #get time_unit
   LET lc_time_unit = "/" #default
   IF ls_dbdate MATCHES "*-*" THEN 
      LET lc_time_unit="-"
   END IF
   IF ls_dbdate MATCHES "*.*" THEN 
      LET lc_time_unit="."
   END IF
   IF ls_dbdate MATCHES "*0*" THEN  #不帶
      LET lc_time_unit="0"
   END IF

   LET lc_date = ls_dbdate.subString(1,4)
   LET ls_date=""
   FOR li_i = 1 TO 3 STEP 2
      CASE lc_date[li_i,li_i+1]
         WHEN "Y2" 
             IF lc_time_unit="0" THEN
                LET ls_date=ls_date.trim(),"YY"
             ELSE
                LET ls_date=ls_date.trim(),"YY",lc_time_unit
             END IF

         WHEN "Y4" 
             IF lc_time_unit="0" THEN
               LET ls_date=ls_date.trim(),"YYYY"
             ELSE
               LET ls_date=ls_date.trim(),"YYYY",lc_time_unit
             END IF

         WHEN "MD"
             IF lc_time_unit="0" THEN
                LET ls_date=ls_date.trim(),"MMDD"
             ELSE
                LET ls_date=ls_date.trim(),"MM",lc_time_unit,"DD",lc_time_unit
             END IF

         WHEN "DM" 
             IF lc_time_unit="0" THEN
               LET ls_date=ls_date.trim(),"DDMM"
             ELSE
               LET ls_date=ls_date.trim(),"DD",lc_time_unit,"MM",lc_time_unit
             END IF

      END CASE
   END FOR

   IF lc_time_unit="0" THEN
      RETURN ls_date
   ELSE
      RETURN ls_date.subString(1,ls_date.getLength()-1)
   END IF
   
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_get_today_with_format()
#                  RETURNING 回传参数
# Input parameter: 
#                : 
# Return code....: 回传参数变量1   TO_DATE('2015/12/29','YYYY/MM/DD')
#                : 
# Date & Author..: 2015/12/29 By madey
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_get_today_with_format()
   DEFINE l_today        STRING
   DEFINE r_date         STRING  
   DEFINE ls_date_format STRING
  #DEFINE la_ref_fields  DYNAMIC ARRAY OF VARCHAR(500) #ap_ref用陣列
  #DEFINE la_rtn_fields  DYNAMIC ARRAY OF VARCHAR(500) #ap_ref用陣列
  #DEFINE l_gzou005      LIKE gzou_t.gzou005           #日期年月格式


    #init
    LET l_today = g_today
   #IF cl_null(g_lang) THEN
   #   LET g_lang = "zh_TW"
   #END IF
    
   ##取當下enterprice的系統日期格式
   ##抓azzi500 資料
   #INITIALIZE la_ref_fields TO NULL
   #INITIALIZE la_rtn_fields TO NULL
   #LET la_ref_fields[1] = g_enterprise
   #CALL ap_ref_array2(la_ref_fields,"SELECT gzou005 FROM gzou_t WHERE gzou001=?","") RETURNING la_rtn_fields
   #IF NOT cl_null(la_rtn_fields[1]) THEN
   #   LET l_gzou005 = la_rtn_fields[1]
   #ELSE
   #   DISPLAY "Warning: gzou005 is null, set default value=1"
   #   LET l_gzou005 = "1" #default 
   #END IF
   #
   ##抓azzi600 (scc) 說明
   #INITIALIZE la_ref_fields TO NULL
   #INITIALIZE la_rtn_fields TO NULL
   #LET la_ref_fields[1] = "10" #scc                                                                   
   #LET la_ref_fields[2] = l_gzou005
   #CALL ap_ref_array2(la_ref_fields,"SELECT gzcbl004 FROM gzcbl_t WHERE gzcbl001=? AND gzcbl002=? AND gzcbl003='"||g_lang||"'","") RETURNING la_rtn_fields
   #IF NOT cl_null(la_rtn_fields[1]) THEN
   #   LET ls_date_format = la_rtn_fields[1]
   #ELSE
   #   DISPLAY "Warning: get scc (10)  fail! set default format=YYYY/MM/DD"
   #   LET ls_date_format = "YYYY/MM/DD"
   #END IF

    
    #轉成TO_DATE語法 ex. TO_DATE('2015/12/31','YYYY/MM/DD')
    CALL cl_get_date_format() RETURNING ls_date_format
    LET r_date= "TO_DATE('",l_today,"','",ls_date_format,"')"
    RETURN r_date
    
END FUNCTION

################################################################################
# Descriptions...: 取得系统分类值的说明
# Memo...........:
# Usage..........: CALL cl_get_scc_desc(p_scc,p_scc_value)
#                  RETURNING l_gzcbl004
# Input parameter: p_scc         系统分类码
#                : p_scc_value   系统分类值
# Return code....:  l_gzcbl004    系统分类值的说明
#                : 
# Date & Author..: 2016/03/21 By June
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_get_scc_desc(p_scc,p_scc_value)
   DEFINE p_scc STRING,
          p_scc_value STRING
   DEFINE ls_trigger STRING,
          ls_sql STRING,
          l_gzcbl004 LIKE gzcbl_t.gzcbl004

   TRY
      LET p_scc = p_scc.trim()
      LET p_scc_value = p_scc_value.trim()

      LET ls_trigger = "cl_get_scc_desc"
      LET ls_sql = "SELECT gzcbl004",
                    " FROM gzcbl_t",
                   " WHERE gzcbl001='",p_scc,"' AND gzcbl002='",p_scc_value,"' AND gzcbl003='",g_dlang,"'"
      PREPARE gzcbl_prep FROM ls_sql
      EXECUTE gzcbl_prep INTO l_gzcbl004
      FREE gzcbl_prep

      RETURN l_gzcbl004 CLIPPED
   CATCH
      RETURN NULL
   END TRY

END FUNCTION

################################################################################
# Descriptions...: 取得应用分类值的说明
# Memo...........:
# Usage..........: CALL cl_get_acc_desc(p_acc,p_acc_value)
#                  RETURNING l_oocql004
# Input parameter: p_acc         应用分类码
#                : p_acc_value   应用分类值
# Return code....: l_oocql004    应用分类值的说明
#                :
# Date & Author..: 2016/03/21 By June
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_get_acc_desc(p_acc,p_acc_value)
   DEFINE p_acc STRING,
          p_acc_value STRING
   DEFINE ls_trigger STRING,
          ls_sql STRING,
          l_oocql004 LIKE oocql_t.oocql004

   TRY
      LET p_acc = p_acc.trim()
      LET p_acc_value = p_acc_value.trim()

      LET ls_trigger = "cl_get_acc_desc"
      LET ls_sql = "SELECT oocql004",
                    " FROM oocql_t",
                   " WHERE oocqlent = ",g_enterprise," AND oocql001='",p_acc,"' ",
                   " AND oocql002='",p_acc_value,"' AND oocql003='",g_dlang,"'"
      PREPARE oocql_prep FROM ls_sql
      EXECUTE oocql_prep INTO l_oocql004
      FREE oocql_prep

      RETURN l_oocql004 CLIPPED
   CATCH
      RETURN NULL
   END TRY
END FUNCTION

################################################################################
# Descriptions...: 取得fglrun -V 版本
# Memo...........:
# Usage..........: CALL cl_get_fglrun_ver()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2016/04/26 By jrg542
# Modify.........: 160426-00026 #1
################################################################################
PUBLIC FUNCTION cl_get_fglrun_ver()
   DEFINE ls_host   STRING
   DEFINE ch        base.Channel
   DEFINE l_token     base.StringTokenizer    #190314-00035	add
   DEFINE ls_tmp      STRING                  #190314-00035	add
   
   WHENEVER ERROR CALL cl_err_msg_log

   LET ch = base.Channel.create()  #211105-00021#1 add

   IF cl_null(gs_fglrun_ver) THEN
      #LET ch = base.Channel.create()  #211105-00021#1 mark
      CALL ch.setDelimiter("")
      #CALL ch.openPipe("fglrun -V|grep build","r")     #190314-00035	mark
      CALL ch.openPipe("fglrun -V | grep fglrun ","r")  #190314-00035	add
      LET ls_host = ""
      WHILE ch.read(ls_host)
      END WHILE
      #LET ls_host = ls_host.subString(ls_host.getIndexOf("fglrun",1)+6,ls_host.getIndexOf("build",1)-1)  #190314-00035#2	mark
      #190314-00035--start--
      LET ls_host = ls_host.subString(ls_host.getIndexOf("fglrun",1)+6,ls_host.getLength()-1)

      LET l_token = base.StringTokenizer.create(ls_host," ")
      WHILE l_token.hasMoreTokens()
        LET ls_tmp = l_token.nextToken() 
        EXIT WHILE  
      END WHILE 
      LET ls_host = ls_tmp 
      #190314-00035--end--
      
      LET gs_fglrun_ver = ls_host
   ELSE
      LET ls_host = gs_fglrun_ver
   END IF

   CALL ch.close()
   RETURN ls_host.trim()

END FUNCTION

################################################################################
# Descriptions...: 依據要檢核的表格及執行動作，取得有權限執行此動作的部門清單
# Memo...........:
# Usage..........: CALL cl_get_owndept_list(ps_table,ps_action)
#                  RETURNING ps_dept_list
# Input parameter: ps_table       要檢核的表格
#                : ps_action      執行動作
# Return code....: ps_dept_list   有權限執行的部門清單
#                : 
# Date & Author..: 日期 By 作者
# Modify.........: add by 161109-00070 #1
################################################################################
PUBLIC FUNCTION cl_get_owndept_list(ps_table,ps_action)
   DEFINE ps_table      STRING      #撈取的table名稱
   DEFINE ps_action     STRING      #進行的動作(query / modify / delete)
   DEFINE ps_owndp      STRING      #該table的owndp欄位名稱
   DEFINE ps_dept_list  STRING      #同部門群組清單
   DEFINE ps_diff_dept  STRING      #不同部門間的權限
   DEFINE ps_rela_dept  STRING      #所屬部門間的權限
   DEFINE ls_cnt        LIKE type_t.num5
   
   
   LET ps_dept_list = NULL
   LET ps_diff_dept = NULL
   LET ps_rela_dept = NULL

   #若是super user，則不控卡
   IF mc_gzya003 = "Y" THEN
      DISPLAY "擁有super user權限，所以不控卡"
      RETURN ps_dept_list
   END IF
   
   #250606-00022 ---start---
   #分管人進入可操作的分管作業時才做判斷，當更精準的判斷時，不會在其他作業浪費時間
   IF g_code="azzi850" OR g_code="azzi860" OR g_code="azzi861" OR g_code="azzi802" OR #權限(gzyp003)
      g_code="aooi130" OR g_code="azzi800" OR g_code="azzi801"                        #人員(002)
   THEN
      IF g_responsibility = g_account THEN
         DISPLAY "擁有分管人權限，所以不控卡"
         RETURN ps_dept_list
      END IF
   END IF
   #250606-00022 --- end ---

   #先確定該table有owndp
   LET ps_owndp = ps_table.subString(1,ps_table.getIndexOf('_',1)-1), 'owndp'

   IF cl_getField(ps_table,ps_owndp) THEN
      #依據目前要做的動作，檢核相關動作的權限
      CASE
         WHEN ps_action = "query"
            LET ps_diff_dept = mc_gzyb003_diff_q
            LET ps_rela_dept = mc_gzyb003_rela_q
         
         WHEN ps_action = "modify"
            LET ps_diff_dept = mc_gzyb003_diff_m
            LET ps_rela_dept = mc_gzyb003_rela_m

         WHEN ps_action = "delete"
            LET ps_diff_dept = mc_gzyb003_diff_d
            LET ps_rela_dept = mc_gzyb003_rela_d
      END CASE
      
      #檢核是否有權限可查看資料
      #不同部門間允許查詢/修改/刪除，表示所有資料都可查詢/修改/刪除，就不需增加過濾條件
      IF ps_diff_dept = "N" THEN
         #若有設定相關部門
         IF ma_deptgroup.getLength() > 0 THEN
            IF ps_rela_dept = "Y" THEN   #僅相關部門有權限
               FOR ls_cnt = 1 TO ma_deptgroup.getLength()
                  IF cl_null(ps_dept_list) THEN
                     LET ps_dept_list = "'",ma_deptgroup[ls_cnt].gzxe002,"'"
                  ELSE
                     LET ps_dept_list = ps_dept_list,", '",ma_deptgroup[ls_cnt].gzxe002,"'"
                  END IF
               END FOR
            ELSE                  #僅所屬部門有權限
               LET ps_dept_list = "'",g_dept,"'"
            END IF
         ELSE
            LET ps_dept_list = "'",g_dept,"'"
         END IF
      END IF
   END IF
   
#   DISPLAY "ps_dept_list:",ps_dept_list

   RETURN ps_dept_list
END FUNCTION

################################################################################
# Descriptions...: 依據要檢核的表格及執行動作，取得有權限執行此動作的使用者清單
# Memo...........:
# Usage..........: CALL cl_get_ownuser_list(ps_table,ps_action)
#                  RETURNING ps_ownid_list
# Input parameter: ps_table       要檢核的表格
#                : ps_action      執行動作
# Return code....: ps_ownid_list  有權限執行此動作的使用者清單
#                : 
# Date & Author..: 日期 By 作者
# Modify.........: add by 161109-00070 #1
################################################################################
PUBLIC FUNCTION cl_get_ownuser_list(ps_table,ps_action)
   DEFINE ps_table      STRING      #撈取的table名稱
   DEFINE ps_action     STRING      #進行的動作(query / modify / delete)
   DEFINE ps_ownid      STRING      #該table的ownid欄位名稱
   DEFINE ps_ownid_list STRING      #資料擁有者清單
   DEFINE ls_cnt        LIKE type_t.num5


   LET ps_ownid_list = NULL

   #若是super user，則不控卡
   IF mc_gzya003 = "Y" THEN
      DISPLAY "擁有super user權限，所以不控卡"
      RETURN ps_ownid_list
   END IF
   
   #250606-00022 ---start---
   #分管人進入可操作的分管作業時才做判斷，當更精準的判斷時，不會在其他作業浪費時間
   IF g_code="azzi850" OR g_code="azzi860" OR g_code="azzi861" OR g_code="azzi802" OR #權限(gzyp003)
      g_code="aooi130" OR g_code="azzi800" OR g_code="azzi801"                        #人員(002)
   THEN
      IF g_responsibility = g_account THEN
         DISPLAY "擁有分管人權限，所以不控卡"
         RETURN ps_ownid_list
      END IF
   END IF
   #250606-00022 --- end ---

   #先確定該table有ownid
   LET ps_ownid = ps_table.subString(1,ps_table.getIndexOf('_',1)-1), 'ownid'
   
   IF cl_getField(ps_table,ps_ownid) THEN
      #依據目前要做的動作，檢核相關動作的權限
      CASE
         WHEN ps_action = "query"
            IF mc_gzyb004 = "4" THEN    #非擁有者不可查詢
               #增加過濾條件
               IF cl_null(ps_ownid_list) THEN
                  LET ps_ownid_list = "'",g_user CLIPPED,"'"
               ELSE
                  LET ps_ownid_list = ps_ownid_list,", '",g_user CLIPPED,"'"
               END IF
            END IF

         WHEN ps_action = "modify"
            IF mc_gzyb004 = "2" OR mc_gzyb004 = "3" OR mc_gzyb004 = "4" THEN    #非擁有者不可修改
               #增加過濾條件
               IF cl_null(ps_ownid_list) THEN
                  LET ps_ownid_list = "'",g_user CLIPPED,"'"
               ELSE
                  LET ps_ownid_list = ps_ownid_list,", '",g_user CLIPPED,"'"
               END IF
            END IF
         
         WHEN ps_action = "delete"
            IF mc_gzyb004 = "1" OR mc_gzyb004 = "3" OR mc_gzyb004 = "4" THEN    #非擁有者不可刪除
               #增加過濾條件
               IF cl_null(ps_ownid_list) THEN
                  LET ps_ownid_list = "'",g_user CLIPPED,"'"
               ELSE
                  LET ps_ownid_list = ps_ownid_list,", '",g_user CLIPPED,"'"
               END IF
            END IF

      END CASE
   END IF

#   DISPLAY "ps_ownid_list:",ps_ownid_list

   RETURN ps_ownid_list
END FUNCTION

################################################################################
# Descriptions...: 傳入作業編號 回傳程式編號
# Memo...........:
# Usage..........: CALL cl_get_prog_code(pc_gzzz001)
#                  RETURNING 回传参数
# Input parameter: pc_gzzz001 作業編號
# Return code....: lc_gzzz002
# Date & Author..: 日期 By jrg542
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_get_prog_code(pc_gzzz001)
   DEFINE lc_gzzz002     LIKE gzzz_t.gzzz002
   DEFINE pc_gzzz001     LIKE gzzz_t.gzzz001

   #由作業名稱轉換為程式名稱
   SELECT gzzz002 INTO lc_gzzz002 FROM gzzz_t
    WHERE gzzz001 = pc_gzzz001

   RETURN lc_gzzz002
END FUNCTION

################################################################################
# Descriptions...: 取得要使用何種類型匯出excel
# Memo...........:
# Usage..........: CALL cl_get_exporttoexcel_type()
#                  RETURNING ps_type
# Input parameter: 
# Return code....: ps_type (org:使用原始寫法  poi:使用poi寫法)
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_get_exporttoexcel_type()
   DEFINE ps_sql             STRING
   DEFINE ps_type            LIKE type_t.chr20      #回傳值 (org:使用原始寫法  poi:使用poi寫法)
   DEFINE ps_use_org         LIKE type_t.chr1       #是否使用原始寫法
   DEFINE ps_use_poi         LIKE type_t.chr1       #是否使用poi寫法
   DEFINE ps_add_point       DYNAMIC ARRAY OF RECORD
          dzba003            LIKE dzba_t.dzba003,   #程式設計點(add-point)
          max_dzba004        LIKE dzba_t.dzba004,   #設計點版次
      #   dzba010            LIKE dzba_t.dzba010,   #客製碼(add-point單頭)   #180117-00034 mark
          dzba005            LIKE dzba_t.dzba005,   #客製碼(add-point單身)   #180117-00034 add
          dzbb098            LIKE dzbb_t.dzbb098,   #設計點內容
          rep_type           LIKE type_t.chr20      #區分是哪一種寫法 (org:原始寫法  poi:poi寫法)
                             END RECORD
   DEFINE ps_cnt             LIKE type_t.num5
   DEFINE li_count           LIKE type_t.num5
   DEFINE lc_dzbb098         LIKE dzbb_t.dzbb098
   DEFINE ls_tmp             STRING
   
   
   #exporttoexcel段，目前有8個add-point可能會有原始樣板與poi樣版的差異，
   #分別為 原始樣板用的add-point：menu.exporttoexcel_browser，menu.exporttoexcel_browser_after，
   #                              menu.exporttoexcel，menu.exporttoexcel_after，
   #                              ui_dialog.exporttoexcel_browser，ui_dialog.exporttoexcel_browser_after
   #        poi樣板用的add-point：menu.exporttoexcel_browser_poi，menu.exporttoexcel_browser_poi_after，
   #                              menu.exporttoexcel_poi，menu.exporttoexcel_after_poi，
   #                              ui_dialog.exporttoexcel_browser_poi，ui_dialog.exporttoexcel_browser_after_poi
   
   LET ps_type = NULL
   LET ps_use_org = "N"
   LET ps_use_poi = "N"
   CALL ps_add_point.clear()

   #先將要查找的add-point放入陣列中，並做好分類，以供後續取用
   LET ps_cnt = 1
   #原始樣板用的add-point
   LET ps_add_point[ps_cnt].dzba003 = "menu.exporttoexcel_browser"            LET ps_add_point[ps_cnt].rep_type = "org"  LET ps_cnt = ps_cnt + 1
   LET ps_add_point[ps_cnt].dzba003 = "menu.exporttoexcel_browser_after"      LET ps_add_point[ps_cnt].rep_type = "org"  LET ps_cnt = ps_cnt + 1
   LET ps_add_point[ps_cnt].dzba003 = "menu.exporttoexcel"                    LET ps_add_point[ps_cnt].rep_type = "org"  LET ps_cnt = ps_cnt + 1
   LET ps_add_point[ps_cnt].dzba003 = "menu.exporttoexcel_after"              LET ps_add_point[ps_cnt].rep_type = "org"  LET ps_cnt = ps_cnt + 1
   #poi樣板用的add-point
   LET ps_add_point[ps_cnt].dzba003 = "menu.exporttoexcel_browser_poi"        LET ps_add_point[ps_cnt].rep_type = "poi"  LET ps_cnt = ps_cnt + 1
   LET ps_add_point[ps_cnt].dzba003 = "menu.exporttoexcel_browser_poi_after"  LET ps_add_point[ps_cnt].rep_type = "poi"  LET ps_cnt = ps_cnt + 1
   LET ps_add_point[ps_cnt].dzba003 = "menu.exporttoexcel_poi"                LET ps_add_point[ps_cnt].rep_type = "poi"  LET ps_cnt = ps_cnt + 1
   LET ps_add_point[ps_cnt].dzba003 = "menu.exporttoexcel_poi_after"          LET ps_add_point[ps_cnt].rep_type = "poi"  LET ps_cnt = ps_cnt + 1
   
   #查找規則，先個別確定每個add-point是標準還是客製，及其最大的設計版次
   #並確認該add-point是否已有資料
   FOR ps_cnt = 1 TO ps_add_point.getLength()
      #分別判斷各add-point是否已被客製
      LET li_count = 0
      SELECT count(1) INTO li_count FROM dzba_t
       WHERE dzba001 = g_code   #因為有可能有程式共用的狀況，因此dzba_t中紀錄的是程式代碼
         AND dzba003 = ps_add_point[ps_cnt].dzba003
         AND dzbastus = "Y"
         AND dzba005 = "c"   #AND dzba010 = "c"   #180117-00034 modify (應檢查單身add-point，而非單頭add-point)
      IF li_count > 0 THEN
      #  LET ps_add_point[ps_cnt].dzba010 = "c"   #180117-00034 mark
         LET ps_add_point[ps_cnt].dzba005 = "c"   #180117-00034 add (應檢查單身add-point，而非單頭add-point)
      ELSE
      #  LET ps_add_point[ps_cnt].dzba010 = "s"   #180117-00034 mark
         LET ps_add_point[ps_cnt].dzba005 = "s"   #180117-00034 add (應檢查單身add-point，而非單頭add-point)
      END IF
      
      #取出此add-point標準或客製的最大程式版次及最大設計點版次
      LET li_count = 0
      SELECT MAX(dzba004) INTO li_count
        FROM dzba_t
       WHERE dzba001 = g_code   #程式編號  因為有可能有程式共用的狀況，因此dzba_t中紀錄的是程式代碼
         AND dzba003 = ps_add_point[ps_cnt].dzba003   #程式設計點
      #  AND dzba010 = ps_add_point[ps_cnt].dzba010   #標準或客製(單頭)   #180117-00034 mark
         AND dzba010 = ps_add_point[ps_cnt].dzba005   #標準或客製(單身)   #180117-00034 add
      IF li_count IS NULL OR li_count = 0 THEN
         LET ps_add_point[ps_cnt].max_dzba004 = 0
         
         #若為0或null，表示這個add-point沒有被使用過，因此不須再往下進行
         CONTINUE FOR
      ELSE
         LET ps_add_point[ps_cnt].max_dzba004 = li_count
      END IF

      #取出該add-point的設計點內容
      LOCATE lc_dzbb098 IN FILE
      
      SELECT dzbb098 INTO lc_dzbb098
        FROM dzbb_t
       WHERE dzbb001 = g_code   #程式編號
         AND dzbb002 = ps_add_point[ps_cnt].dzba003       #程式設計點
         AND dzbb003 = ps_add_point[ps_cnt].max_dzba004   #最大設計點版次
      #  AND dzbb004 = ps_add_point[ps_cnt].dzba010       #標準或客製(單頭)   #180117-00034 mark
         AND dzbb004 = ps_add_point[ps_cnt].dzba005       #標準或客製(單身)   #180117-00034 add

      LET ls_tmp = lc_dzbb098
      FREE lc_dzbb098
      
      #判斷設計點內容是否有資料
      IF ps_add_point[ps_cnt].rep_type = "poi" THEN
         IF NOT cl_null(ls_tmp) THEN
            LET ps_use_poi = "Y"
         END IF
      ELSE
         IF NOT cl_null(ls_tmp) THEN
            LET ps_use_org = "Y"
         END IF
      END IF
   END FOR
   
   #依據所得資訊，判斷應該使用何種報表樣板
   CASE
      #poi類的add-point有資料，原始的也有資料，表示已調整過，用poi的樣板
      WHEN ps_use_poi = "Y" AND ps_use_org = "Y"
         LET ps_type = "poi"

      #poi類的add-point有資料，原始的沒有資料，直接可使用poi的樣板
      WHEN ps_use_poi = "Y" AND ps_use_org = "N"
         LET ps_type = "poi"

      #poi類的add-point沒有資料，但原始的有資料，表示還沒調整成poi樣板的寫法，因此要用原始的樣板
      WHEN ps_use_poi = "N" AND ps_use_org = "Y"
         LET ps_type = "org"

      #若poi類的與原始的add-point都沒有資料，表示不用調整，因此可直接用poi的樣板
      WHEN ps_use_poi = "N" AND ps_use_org = "N"
         LET ps_type = "poi"
      
   END CASE
   
   RETURN ps_type
   
END FUNCTION

################################################################################
# Descriptions...: 取得suoerrole
# Memo...........:
# Usage..........: CALL cl_get_role_superuser()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: Y/N 
# Date & Author..: 2018/04/17 By 01856
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_get_role_superuser()
   DEFINE lst_token     base.StringTokenizer
   DEFINE ls_token      STRING
   DEFINE lc_gzya003    LIKE gzya_t.gzya003
   DEFINE li_cnt        LIKE type_t.num5

   #取superuser 
   SELECT COUNT(gzya003) INTO li_cnt FROM gzxb_t 
    INNER JOIN gzya_t ON gzxbent = gzyaent AND gzxb003 = gzya001 AND gzyastus = 'Y' 
    WHERE gzxbent = g_enterprise
      AND gzxb001 = g_user
      AND gzxbstus = 'Y'
   IF li_cnt > 0 THEN 
      LET mc_gzya003 = "Y"
   ELSE 
      LET mc_gzya003 = "N"
   END IF 
   
   RETURN mc_gzya003
END FUNCTION

################################################################################
# Descriptions...: 取主程式模組編號
# Memo...........:
# Usage..........: CALL cl_get_main_prog_module(pc_prog)
#                  RETURNING module
# Input parameter: pc_prog 程式編號
# Return code....: ls_str  模組編號
# Date & Author..: 2018/12/03 By 01856
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_get_main_prog_module(pc_prog)
   DEFINE pc_prog     LIKE gzza_t.gzza001
   DEFINE lc_gzza004  LIKE gzza_t.gzza004
   DEFINE ls_str      STRING    

   SELECT gzza004 INTO lc_gzza004 FROM gzza_t
    WHERE gzza001 = pc_prog

   LET ls_str = lc_gzza004
   LET ls_str =  ls_str.subString(ls_str.getIndexOf("$",2)+1,ls_str.getIndexOf("$",2)+3)
   LET ls_str =  DOWNSHIFT(ls_str)    

   RETURN ls_str
END FUNCTION

################################################################################
# Descriptions...: 日期格式特別處理
# Memo...........:
# Usage..........: CALL cl_get_para_date(lc_enterprise,lc_site,lc_para_id,ls_value)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2019/09/04 By 01856
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_get_para_date(lc_enterprise,lc_site,lc_para_id,ls_value)
   DEFINE lc_enterprise  LIKE type_t.num5
   DEFINE lc_site        LIKE type_t.chr10
   DEFINE lc_para_id     LIKE gzsz_t.gzsz002 #參數編號
   DEFINE lc_gzsz001     LIKE gzsz_t.gzsz001 #參數所屬table  #210319-00004#1
   DEFINE lc_gzsz003     LIKE gzsz_t.gzsz003 #輸入限制資料型態
   DEFINE ls_format      STRING
   DEFINE ls_value       STRING
   DEFINE pi_type        LIKE type_t.num5
   DEFINE ls_token       STRING
   DEFINE l_token        base.StringTokenizer
   DEFINE ls_year        STRING
   DEFINE ls_mm          STRING
   DEFINE ls_dd          STRING
   DEFINE li_num         LIKE type_t.num5    #190926-00044#1
   DEFINE li_format      LIKE type_t.num5    #190926-00044#1
   DEFINE li_cnt         LIKE type_t.num10   #210603-00002#1 放大為 NUM10
   DEFINE li_ma_para_def BOOLEAN

   #210319-00004#1
   LET li_ma_para_def = FALSE
   #檢測暫存陣列
   FOR li_cnt = 1 TO ma_para_def.getLength()
      IF lc_para_id = ma_para_def[li_cnt].paracode THEN  #211125-00020 _set -> _def
         LET lc_gzsz001 = ma_para_def[li_cnt].gzsz001
         LET lc_gzsz003 = ma_para_def[li_cnt].gzsz003
         LET li_ma_para_def = TRUE
         EXIT FOR
      END IF
   END FOR
   #不存在只好查db後放入
   IF NOT li_ma_para_def THEN
      SELECT gzsz001,gzsz003 INTO lc_gzsz001,lc_gzsz003 FROM gzsz_t
       WHERE gzsz002 = lc_para_id
      IF NOT SQLCA.SQLCODE THEN
         LET li_cnt = ma_para_def.getLength()+1
         IF li_cnt < 50000 THEN                     #210603-00002#1 上限50000筆超過不存   #211125-00020
            IF g_t100debug >= 6 THEN display 'AW-DEF:pa-code:',lc_para_id,' 001:',lc_gzsz001,' 003:',lc_gzsz003 END IF   #211125-00020
            LET ma_para_def[li_cnt].paracode = lc_para_id
            LET ma_para_def[li_cnt].gzsz001 = lc_gzsz001
            LET ma_para_def[li_cnt].gzsz003 = lc_gzsz003
         END IF
      END IF
   END IF

   #日期格式特別處理
   IF NOT lc_gzsz003 = "5" THEN
      RETURN ls_value
   END IF

   LET ls_format = FGL_GETENV("DBDATE")
   #190926-00044#1--start--
   #LET l_token = base.StringTokenizer.create(ls_value,"/")
   CASE 
      WHEN ls_value.getIndexOf("/",1) 
         LET l_token = base.StringTokenizer.create(ls_value,"/")
      WHEN ls_value.getIndexOf("-",1) 
         LET l_token = base.StringTokenizer.create(ls_value,"-")
      OTHERWISE 
         LET l_token = base.StringTokenizer.create(ls_value,"/")
   END CASE   
   #190926-00044#1--end--
        
   WHILE l_token.hasMoreTokens()
      LET ls_year = l_token.nextToken()
      LET ls_mm = l_token.nextToken() 
      LET ls_dd = l_token.nextToken()  
      EXIT WHILE
   END WHILE 
   #190926-00044#1--start--
   #ex:30-01-2014 #2014-30-01 #01-30-2014
   IF ls_year.getLength() = 4 THEN 
      LET li_format = 0
   ELSE 
      #表示 日期或月份
      LET li_num = ls_year
      IF li_num > 12 THEN 
         LET li_format = 1
         #表示月份
      ELSE 
         #表示日期
         LET li_format = 2
      END IF 
   END IF 
   #190926-00044#1--end--
       
   CASE 
      WHEN ls_format = "Y4MD/" #2014/01/01
         RETURN ls_value          
      WHEN ls_format = "Y4MD-" #2014-01-01
         LET ls_value = ls_year,"-",ls_mm,"-",ls_dd
      WHEN ls_format = "DMY4-" #30-01-2014
         #190926-00044#1--start-- 
         IF li_format = 0 THEN 
            LET ls_value = ls_dd,"-",ls_mm,"-",ls_year
         END IF 
         IF li_format = 1 OR li_format = 2 THEN 
            #表示月份
            LET ls_value = ls_year ,"-",ls_mm,"-",ls_dd
         END IF 
         #LET ls_value = ls_dd,"-",ls_mm,"-",ls_year
         #190926-00044#1--end--              
      WHEN ls_format = "DMY4/" #30/01/2014
         #190926-00044#1--start--            
         IF li_format = 0 THEN 
            LET ls_value = ls_dd,"/",ls_mm,"/",ls_year
         END IF 

         IF li_format = 1 OR li_format = 2 THEN 
            #表示月份 日期
            LET ls_value = ls_year ,"/",ls_mm,"/",ls_dd
         END IF           
         #LET ls_value = ls_dd,"/",ls_mm,"/",ls_year
         #190926-00044#1--end--               
             
      WHEN ls_format = "MDY4/" #01/30/2014
         #190926-00044#1--start--               
         IF li_format = 0 THEN 
            LET ls_value = ls_mm,"/",ls_dd,"/",ls_year
         END IF 

         IF li_format = 1 OR li_format = 2 THEN 
            #表示月份 日期
            LET ls_value = ls_year ,"/",ls_mm,"/",ls_dd
         END IF               
         #LET ls_value = ls_mm,"/",ls_dd,"/",ls_year
         #190926-00044#1--end--               
   END CASE
   RETURN ls_value 
END FUNCTION

################################################################################
# Descriptions...: 時間資料型態是varchar 把日期資料轉成年月日
# Memo...........:
# Usage..........: CALL cl_get_data_trans_yymmdd(ps_value)
#                  RETURNING 回传参数
# Input parameter: ps_value STRING 日期
# Return code....: ls_value STRING 日期 yyyy/mm/dd
# Date & Author..: 2019/10/18 By 01856
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_get_data_trans_yymmdd(ps_value)
   DEFINE ps_value    STRING
   DEFINE ls_value    STRING   
   DEFINE l_token     base.StringTokenizer
   DEFINE ls_str      STRING
   DEFINE ls_str2     STRING  
   DEFINE ls_str3     STRING    
   DEFINE ls_format   STRING    

   LET ls_format = FGL_GETENV("DBDATE")
   
   CASE 
      WHEN ps_value.getIndexOf("/",1) 
         LET l_token = base.StringTokenizer.create(ps_value,"/")
      WHEN ps_value.getIndexOf("-",1) 
         LET l_token = base.StringTokenizer.create(ps_value,"-")
      OTHERWISE 
         #LET ps_value = base.StringTokenizer.create(ps_value,"/")   #191025-00018#1   mark
         LET l_token = base.StringTokenizer.create(ps_value,"/")     #191025-00018#1   add
   END CASE 

   WHILE l_token.hasMoreTokens()
      LET ls_str = l_token.nextToken()
      LET ls_str2 = l_token.nextToken() 
      LET ls_str3 = l_token.nextToken()  
      EXIT WHILE
   END WHILE 

   CASE 
      WHEN ls_format = "Y4MD/" #2014/01/01
         #轉成年月日
         LET ls_value = ls_str,"/",ls_str2,"/",ls_str3 #2014/01/01
   
      WHEN ls_format = "Y4MD-" #2014-01-01
         #轉成年月日
         LET ls_value = ls_str,"/",ls_str2,"/",ls_str3 #2014/01/01
     
      WHEN ls_format = "DMY4-" #30-01-2014
         #轉成年月日
         LET ls_value = ls_str3,"/",ls_str2,"/",ls_str #2014/01/01
         
      WHEN ls_format = "DMY4/" #30/01/2014
         #轉成年月日
         LET ls_value = ls_str3,"/",ls_str2,"/",ls_str #2014/01/01           
   
      WHEN ls_format = "MDY4/" #01/30/2014
         #轉成年月日
         LET ls_value = ls_str3,"/",ls_str,"/",ls_str2 #2014/01/01          
   
      WHEN ls_format = "Y2MD/"
         #轉成年月日
         LET ls_value = ls_str,"/",ls_str2,"/",ls_str3 #2014-01-01
   
      OTHERWISE   
       #2014/01/01 預設
         LET ls_value = ls_str,"/",ls_str2,"/",ls_str3 #2014-01-01
   END CASE

    RETURN ls_value
END FUNCTION

################################################################################
# Descriptions...: 時間類別是char 要轉成對應dbdate 
# Memo...........:
# Usage..........: CALL cl_get_data_trans_dbdate(ps_value)
#                  RETURNING 回传参数
# Input parameter: ps_value STRING 日期
# Return code....: ls_value STRING dbdate 日期格式
# Date & Author..: 2019/10/18 By 10856
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_get_data_trans_dbdate(ps_value)
   DEFINE ps_value    STRING
   DEFINE ls_value    STRING 
   DEFINE ls_format   STRING 
   DEFINE ls_token    STRING 
   DEFINE l_token     base.StringTokenizer
   DEFINE ls_year     STRING
   DEFINE ls_mm       STRING  
   DEFINE ls_dd       STRING   

   LET ls_format = FGL_GETENV("DBDATE")

   CASE 
      WHEN ps_value.getIndexOf("/",1) 
         LET l_token = base.StringTokenizer.create(ps_value,"/")
      WHEN ps_value.getIndexOf("-",1) 
         LET l_token = base.StringTokenizer.create(ps_value,"-")
      OTHERWISE 
         LET l_token = base.StringTokenizer.create(ps_value,"/")
   END CASE   
   
   WHILE l_token.hasMoreTokens()
      LET ls_year = l_token.nextToken()
      LET ls_mm = l_token.nextToken() 
      LET ls_dd = l_token.nextToken()  
      EXIT WHILE
   END WHILE 
   
   CASE 
      WHEN ls_format = "Y4MD/" #2014/01/01
         LET ls_value = ps_value
         RETURN ls_value          
   
      WHEN ls_format = "Y4MD-" #2014-01-01
     
         LET ls_value = ls_year,"-",ls_mm,"-",ls_dd
         
      WHEN ls_format = "DMY4-" #30-01-2014
   
         LET ls_value = ls_dd,"-",ls_mm,"-",ls_year
       
      WHEN ls_format = "DMY4/" #30/01/2014
      
         LET ls_value = ls_dd,"/",ls_mm,"/",ls_year
                    
      WHEN ls_format = "MDY4/" #01/30/2014
         LET ls_value = ls_mm,"/",ls_dd,"/",ls_year
       
      WHEN ls_format = "Y2MD/" 
         LET ls_value = ps_value
         RETURN ls_value
   
      OTHERWISE    
         LET ls_value = ps_value      
         RETURN ls_value   
   END CASE
   
   RETURN ls_value    
END FUNCTION

################################################################################
# Descriptions...: 關閉CURSOR
# Memo...........:
# Usage..........: cl_get_para_close_res(lc_type)
#                  RETURNING 回传参数
# Input parameter: 類型 A/E/S
################################################################################
PUBLIC FUNCTION cl_get_para_close_res(lc_type)

   DEFINE lc_type  LIKE type_t.chr1

&ifdef v3
   CASE
      WHEN lc_type = "A"   #系統級
         CLOSE cl_get_para_cs_a
         FREE cl_get_para_cs_a
      WHEN lc_type = "E"   #企業級
         CLOSE cl_get_para_cs_e
         FREE cl_get_para_cs_e
      WHEN lc_type = "S"   #據點級
         CLOSE cl_get_para_cs_s
         FREE cl_get_para_cs_s
   END CASE
&endif

END FUNCTION

################################################################################
# Descriptions...: 取得首頁報表位址及認證碼
# Memo...........:
# Usage..........: CALL cl_get_rep_choose()
#                  RETURNING lc_link,lc_gzgh001
# Input parameter: 
# Return code....: lc_link        報表位址
#                : lc_gzgh001     驗證碼
# Date & Author..: No:220125-00020#2 by 00824
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_get_rep_choose()
   DEFINE lc_link    STRING
   DEFINE lc_gzgh001 LIKE gzgh_t.gzgh001

   #先由環境變數(DIGIWINPROD)判斷是否為產中環境
   &ifdef env_is_dgw
      #若是產中環境，為便於開發及維護，FR與ABI的程式段落需同時存在，只是在執行的時候才由環境變數(TOPVRPT)判斷要走FR或是ABI
      IF FGL_GETENV("TOPVRPT") = "ABI" THEN
         CALL cl_abi_view() RETURNING lc_link,lc_gzgh001
         DISPLAY "dgw abi lc_link:",lc_link
      ELSE
         #預設是FR，因此若為NULL也是走FR段
         CALL cl_fr_view() RETURNING lc_link,lc_gzgh001
         DISPLAY "dgw FR lc_link:",lc_link
      END IF
   &else
      #若是客戶環境，FR與ABI的段落只會有一種，不會同時存在
      #編譯時需依據環境變數(TOPVRPT)決定要編FR或是ABI的段落
      &ifdef rpt_is_abi
         CALL cl_abi_view() RETURNING lc_link,lc_gzgh001
         DISPLAY "cust abi lc_link:",lc_link
      &else
         #預設是FR，因此若為NULL也是走FR段
         CALL cl_fr_view() RETURNING lc_link,lc_gzgh001
         DISPLAY "cust FR lc_link:",lc_link
      &endif
   &endif

   RETURN lc_link,lc_gzgh001
   
END FUNCTION

################################################################################
# Descriptions...: 組合報表欄位資料的SQL
# Memo...........:
# Usage..........: CALL cl_get_report_sql()
#                  RETURNING ls_sql
# Input parameter: 
# Return code....: ls_sql         SQL
# Date & Author..: No:220125-00020#2 By 00824
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_get_report_sql()
   DEFINE ls_sql     STRING


   LET ls_sql = NULL

   #先由環境變數(DIGIWINPROD)判斷是否為產中環境
   &ifdef env_is_dgw
      #若是產中環境，為便於開發及維護，FR與ABI的程式段落需同時存在，只是在執行的時候才由環境變數(TOPVRPT)判斷要走FR或是ABI
      IF FGL_GETENV("TOPVRPT") = "ABI" THEN
         LET ls_sql = "SELECT gzxi002,gzxi005,gzgu021,gzgu020,gzgul002 FROM gzxi_t",
                       " INNER JOIN gzgu_t ON gzgu021 = gzxi002 AND gzgu007 = '1' ",
                       " INNER JOIN gzgul_t ON gzgul000 = gzgu000 AND gzgul001 = '",g_lang,"'",
                       " WHERE gzxi001 = '",g_user,"'",
                       "   AND gzxi003 = 'ABI'",
                       "   AND gzxient = ",g_enterprise,
                       " ORDER BY gzxi007"
      ELSE
         #預設是FR，因此若為NULL也是走FR段
         LET ls_sql = "SELECT gzxi002,gzxi005,gzgq021,gzgq020,gzgql002 FROM gzxi_t",
                       " INNER JOIN gzgq_t ON gzgq021 = gzxi002 AND gzgq007 = '1' ",
                       " INNER JOIN gzgql_t ON gzgql000 = gzgq000 AND gzgql001 = '",g_lang,"'",
                       " WHERE gzxi001 = '",g_user,"'",
                       "   AND gzxi003 = 'FR'",
                       "   AND gzxient = ",g_enterprise,
                       " ORDER BY gzxi007"
      END IF
   &else
      #若是客戶環境，FR與ABI的段落只會有一種，不會同時存在
      #編譯時需依據環境變數(TOPVRPT)決定要編FR或是ABI的段落
      &ifdef rpt_is_abi
         LET ls_sql = "SELECT gzxi002,gzxi005,gzgu021,gzgu020,gzgul002 FROM gzxi_t",
                       " INNER JOIN gzgu_t ON gzgu021 = gzxi002 AND gzgu007 = '1' ",
                       " INNER JOIN gzgul_t ON gzgul000 = gzgu000 AND gzgul001 = '",g_lang,"'",
                       " WHERE gzxi001 = '",g_user,"'",
                       "   AND gzxi003 = 'ABI'",
                       "   AND gzxient = ",g_enterprise,
                       " ORDER BY gzxi007"
      &else
         #預設是FR，因此若為NULL也是走FR段
         LET ls_sql = "SELECT gzxi002,gzxi005,gzgq021,gzgq020,gzgql002 FROM gzxi_t",
                       " INNER JOIN gzgq_t ON gzgq021 = gzxi002 AND gzgq007 = '1' ",
                       " INNER JOIN gzgql_t ON gzgql000 = gzgq000 AND gzgql001 = '",g_lang,"'",
                       " WHERE gzxi001 = '",g_user,"'",
                       "   AND gzxi003 = 'FR'",
                       "   AND gzxient = ",g_enterprise,
                       " ORDER BY gzxi007"
      &endif
   &endif

   RETURN ls_sql
   
END FUNCTION

################################################################################
# Descriptions...: 組合取得報表ID的SQL
# Memo...........:
# Usage..........: CALL cl_get_report_id_sql()
#                  RETURNING ls_sql
# Input parameter: 
# Return code....: ls_sql         SQL
# Date & Author..: No:220125-00020#2 By 00824
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_get_report_id_sql()
   DEFINE ls_sql     STRING


   LET ls_sql = NULL

   #先由環境變數(DIGIWINPROD)判斷是否為產中環境
   &ifdef env_is_dgw
      #若是產中環境，為便於開發及維護，FR與ABI的程式段落需同時存在，只是在執行的時候才由環境變數(TOPVRPT)判斷要走FR或是ABI
      IF FGL_GETENV("TOPVRPT") = "ABI" THEN
         LET ls_sql = "SELECT gzgu021,gzgu020",
                      "  FROM gzgu_t ",
                      " WHERE gzgu021 = ? AND gzgu007 = '1' "
      ELSE
         #預設是FR，因此若為NULL也是走FR段
         LET ls_sql = "SELECT gzgq021,gzgq020",
                      "  FROM gzgq_t ",
                      " WHERE gzgq021 = ? AND gzgq007 = '1' "
      END IF
   &else
      #若是客戶環境，FR與ABI的段落只會有一種，不會同時存在
      #編譯時需依據環境變數(TOPVRPT)決定要編FR或是ABI的段落
      &ifdef rpt_is_abi
         LET ls_sql = "SELECT gzgu021,gzgu020",
                      "  FROM gzgu_t ",
                      " WHERE gzgu021 = ? AND gzgu007 = '1' "
      &else
         #預設是FR，因此若為NULL也是走FR段
         LET ls_sql = "SELECT gzgq021,gzgq020",
                      "  FROM gzgq_t ",
                      " WHERE gzgq021 = ? AND gzgq007 = '1' "
      &endif
   &endif

   RETURN ls_sql
   
END FUNCTION

 
{</section>}
 
