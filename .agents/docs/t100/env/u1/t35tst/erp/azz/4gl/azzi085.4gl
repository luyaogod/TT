#該程式已解開Section, 不再透過樣板產出!
#該程式非freestyle程式!
{<section id="azzi085.description" type="s" >}
#應用 a00 樣板自動產生(Version:3)
#+ Standard Version.....: SD版次:0028(2016-11-30 16:09:32), PR版次:0047(2025-10-17 15:16:53)
#+ Customerized Version.: SD版次:0000(1900-01-01 00:00:00), PR版次:0000(1900-01-01 00:00:00)
#+ Build......: 000285
#+ Filename...: azzi085
#+ Description: 資料庫連線設定作業
#+ Creator....: 00845(2013-06-01 00:00:00)
#+ Modifier...: 01856 -SD/PR- 06978
 
{</section>}
 
{<section id="azzi085.global" type="s" >}
#應用 i01 樣板自動產生(Version:81)
#add-point:填寫註解說明 name="global.memo"
#160922-00004 #1 2016/09/22 by jrg542 azzi085 管控 ds開頭不可以刪除
#161124-00001 #1 2016/11/30 by jrg542 azzi085新增一個中間庫的對應欄位
#170216-00032 #1 2016/02/16 by jrg542 在標準環境，型管資料庫編號(scm)，不允許刪除
#170413-00060#2 2017/04/21 by Cynthia 1.dsrept密碼變更時，自動重新產生poijdbc.ini
#                                     2.[整單操作]新增[重產poijdbc.ini]，使用者可手動重產poijdbc
#170628-00013#1 2017/06/28 by jrg542  azzi085  配合 r.t  功能調整 呼叫  crt_new_db_adz_init.sql
#170731-00041#1 2017/08/01 by frank0521 增加cl_log_modified_important()  在失效/有效
#171127-00025#1 2017/11/27 by 00566 dsaps的引擎路徑有改
#180131-00085#1 2018/01/31 by 01856 azzi085使用，呼叫 createdbuser  建立 db user ，不用在分兩次，先建立底層db user 再到 azzi085 建立 資料庫編號
#180312-00040#1 2018/03/12 by 08146 azzi085 修改後，有測試連線，需紀錄log
#180724-00062#1 2018/07/24 by 01856 azzi085 不能刪除只能設定失效、寫入重要事件紀錄訊息優化
#180723-00031#3 2018/07/26 by 01856 調整更新重大事件函式位置
#190423-00046#1 2019/05/02 by 04013 修改密碼時提示操作步驟
#190502-00006#1 2019/05/02 by 04013 create db user 傳入參數修正
#210114-00032#4 2021/01/14 by 12731 開啟 大數據 匯出 excel
#210120-00014#2 2021/01/25 by 12731 (簽出有異常 210712-00018#1 幫追單) 參考 170413-00060#2 製作 multipoijdbc.ini 選項
#210712-00018#1 2021/07/13 by 12731 1. 產生 大數據匯出 excel 相關 ini 檔 不判斷 資料庫類型 ( gzda005 )
#                                   2. 產生 ini 檔原本機制為先刪除在撰寫 ， 調整為直接覆蓋原本檔案內容
#210823-00002#1 2021/08/23 by 12731 產生 ini 方式使用環境 port
#211104-00032#1 2021/11/05 by 12731 產生 multipoijdbc.ini 需要先檢查設定檔
#                                   1. 檔案是否存在 $ORACLE_HOME/jdbc/lib/ojdbc6.jar
#                                   2. 檔案是否存在 $ADZ/42m/adz_adzp523.42m
#                                   ip 抓取 環境設定資訊
#211220-00001#1 2021/12/22 by 12731 1. 畫面欄位資料庫 ip 改由環境來讀取
#                                   2. 產生 multipoijdbc.ini 前先測試連線 ， 沒法測通不產生檔案
#                                   3. 檢查環境變數 CLASSPATH
#220503-00021#1 2022/07/03 by 12731 DB host ip 由 tnsnames.ora 讀取
#221028-00002#1 2022/10/31 by 00845 新增或調整後，只有ORACLE需要問sys/system，其他db種類不需要問(也沒得問)
#221206-00021#1 2022/12/07 by 12731 大數據匯出 excel 新增判斷是否有 adzp523.42r , 點選按鈕才重產 ini
#                                   DB ip 判斷調整如 tnsnames.ora 有找到就先使用
#221206-00021#1 2022/12/07 by 12731 大數據匯出 excel 新增判斷是否有 adzp523.42r , 點選按鈕才重產 ini
#231003-00004#1 2023/10/13 by 02160 1.Modify for MOG/DM commnd tool.
#240319-00048#1 2024/03/20 by 21035 修正jdbc可连接MOG/达梦环境
#240618-00027#1 2024/06/18 by 21035 修正金仓环境取用不到端口号问题
#240709-00026#1 2024/07/09 by 02160 支援Kingbase呼叫command產生設定資料
#241008-00026#1 2024/10/11 by 21035 PGS语法改写
#250110-00014#1 2025/01/10 by 21035 g_db_branch重复定义的修正
#250415-00028#1 2025/04/15 by 18060 支援MogDB模式A呼叫command產生設定資料
#250701-00007#1 2025/07/09 by 06978 修正變更DB密碼，更新底層表格時不需要判斷新舊值是否相同，避免後續實際並未更新為新密碼
#250923-00007#1 2025/09/25 by 21035 取port的sql加上DISTINCT
#end add-point
#add-point:填寫註解說明(客製用) name="global.memo_customerization"

#end add-point
#(ver:55) ---start---
#add-point:填寫註解說明(行業用) name="global.memo_industry"

#end add-point
#(ver:55) --- end --- 
 
IMPORT os
IMPORT util
IMPORT FGL lib_cl_dlg
IMPORT JAVA com.fourjs.fgl.lang.FglRecord  #(ver:56) add
#add-point:增加匯入項目 name="global.import"
IMPORT security   #210114-00032#4 add
IMPORT JAVA java.sql.Connection  #211220-00001#1 add
IMPORT JAVA java.sql.Statement  #211220-00001#1 add
IMPORT JAVA com.digiwin.t100.JdbcConnUtil2  #211220-00001#1 add
#CONSTANT  cs_crt_db_adz_init = "crt_new_db_adz_init.sql"
#231003-00004#1 begin
CONSTANT cs_crt_db_adz_init_pgs = "crt_new_db_adz_init_pgs.sql" #240709-00026#1
CONSTANT cs_crt_db_adz_init_ora = "crt_new_db_adz_init_ora.sql" #240709-00026#1
CONSTANT cs_crt_db_adz_init_kdb = "crt_new_db_adz_init_kdb.sql" #240709-00026#1
#231003-00004#1 end
#end add-point
 
SCHEMA ds
 
GLOBALS "../../cfg/top_global.inc"  
#(ver:56) ---add start---
GLOBALS
#單身匯excel POI變數
TYPE rec_arr_type           ARRAY [] OF com.fourjs.fgl.lang.FglRecord
DEFINE g_export_tag         DYNAMIC ARRAY OF rec_arr_type                 #畫面欄位數
DEFINE g_export_tabname     DYNAMIC ARRAY OF STRING                       #畫面名稱
END GLOBALS
#(ver:56) --- add end ---
 
#add-point:增加匯入變數檔 name="global.inc"
#231003-00004#1 for MogDB begin
GLOBALS
  DEFINE g_default_db_pass STRING
  DEFINE g_db_branch       STRING #POLARDB/MOGDB
END GLOBALS
#231003-00004#1 for MogDB end

DEFINE g_ip                 LIKE type_t.chr50                             #讀取資料庫資訊  #211104-00032#1  add
DEFINE g_port               LIKE type_t.num10                             #讀取資料庫變數  #210823-00002#1  add
#DEFINE g_db_branch          STRING #240618-00027#1 -add #250110-00014 -mark
DEFINE ms_crt_db_adz_init   STRING #240709-00026#1
#end add-point
 
#單頭 type 宣告
PRIVATE TYPE type_g_gzda_m RECORD
       gzda009 LIKE gzda_t.gzda009, 
   gzda002 LIKE gzda_t.gzda002, 
   gzda008 LIKE gzda_t.gzda008, 
   gzda010 LIKE gzda_t.gzda010, 
   gzda005 LIKE gzda_t.gzda005, 
   gzda006 LIKE gzda_t.gzda006, 
   gzda007 LIKE gzda_t.gzda007, 
   gzda001 LIKE gzda_t.gzda001, 
   gzda003 LIKE gzda_t.gzda003, 
   gzda004 LIKE gzda_t.gzda004, 
   gzdastus LIKE gzda_t.gzdastus, 
   gzdaownid LIKE gzda_t.gzdaownid, 
   gzdaownid_desc LIKE ooag_t.ooag011, 
   gzdaowndp LIKE gzda_t.gzdaowndp, 
   gzdaowndp_desc LIKE ooefl_t.ooefl003, 
   gzdacrtid LIKE gzda_t.gzdacrtid, 
   gzdacrtid_desc LIKE ooag_t.ooag011, 
   gzdacrtdp LIKE gzda_t.gzdacrtdp, 
   gzdacrtdp_desc LIKE ooefl_t.ooefl003, 
   gzdacrtdt LIKE gzda_t.gzdacrtdt, 
   gzdamodid LIKE gzda_t.gzdamodid, 
   gzdamodid_desc LIKE ooag_t.ooag011, 
   gzdamoddt LIKE gzda_t.gzdamoddt
       END RECORD
 
DEFINE g_browser    DYNAMIC ARRAY OF RECORD  #查詢方案用陣列 
         b_statepic     LIKE type_t.chr50,
            b_gzda001 LIKE gzda_t.gzda001,
      b_gzda008 LIKE gzda_t.gzda008
      END RECORD 
 
#add-point:自定義模組變數(Module Variable) (請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="global.variable"
DEFINE gs_id        STRING    #180131-00085#1
DEFINE g_ins_logd   BOOLEAN   #180312-00040#1 add 
#end add-point
 
#模組變數(Module Variables)
DEFINE g_gzda_m        type_g_gzda_m  #單頭變數宣告
DEFINE g_gzda_m_t      type_g_gzda_m  #單頭舊值宣告(系統還原用)
DEFINE g_gzda_m_o      type_g_gzda_m  #單頭舊值宣告(其他用途)
DEFINE g_gzda_m_mask_o type_g_gzda_m  #轉換遮罩前資料
DEFINE g_gzda_m_mask_n type_g_gzda_m  #轉換遮罩後資料
 
   DEFINE g_gzda001_t LIKE gzda_t.gzda001
 
   
 
   
 
DEFINE g_wc                  STRING                        #儲存查詢條件
DEFINE g_wc_t                STRING                        #備份查詢條件
DEFINE g_wc_filter           STRING                        #儲存過濾查詢條件
DEFINE g_wc_filter_t         STRING                        #備份過濾查詢條件
DEFINE g_sql                 STRING                        #資料撈取用SQL(含reference)
DEFINE g_forupd_sql          STRING                        #資料鎖定用SQL
DEFINE g_cnt                 LIKE type_t.num10             #指標/統計用變數
DEFINE g_jump                LIKE type_t.num10             #查詢指定的筆數 
DEFINE g_no_ask              LIKE type_t.num5              #是否開啟指定筆視窗 
DEFINE g_rec_b               LIKE type_t.num10             #單身筆數                         
DEFINE l_ac                  LIKE type_t.num10             #目前處理的ARRAY CNT 
DEFINE g_curr_diag           ui.Dialog                     #Current Dialog     
DEFINE gwin_curr             ui.Window                     #Current Window
DEFINE gfrm_curr             ui.Form                       #Current Form
DEFINE g_pagestart           LIKE type_t.num10             #page起始筆數
DEFINE g_page_action         STRING                        #page action
DEFINE g_header_hidden       LIKE type_t.num5              #隱藏單頭
DEFINE g_worksheet_hidden    LIKE type_t.num5              #隱藏工作Panel
DEFINE g_page                STRING                        #第幾頁
DEFINE g_current_sw          BOOLEAN                       #Browser所在筆數用開關
DEFINE g_ch                  base.Channel                  #外串程式用
DEFINE g_state               STRING                        #確認前一個動作是否為新增/複製
DEFINE g_ref_fields          DYNAMIC ARRAY OF VARCHAR(500) #reference用陣列
DEFINE g_ref_vars            DYNAMIC ARRAY OF VARCHAR(500) #reference用陣列
DEFINE g_rtn_fields          DYNAMIC ARRAY OF VARCHAR(500) #reference用陣列
DEFINE g_error_show          LIKE type_t.num5              #是否顯示資料過多的錯誤訊息
DEFINE g_aw                  STRING                        #確定當下點擊的單身(modify_detail用)
DEFINE g_chk                 BOOLEAN                       #助記碼判斷用
DEFINE g_default             BOOLEAN                       #是否有外部參數查詢
DEFINE g_log1                STRING                        #cl_log_modified_record用(舊值)
DEFINE g_log2                STRING                        #cl_log_modified_record用(新值)
 
#快速搜尋用
DEFINE g_searchcol           STRING                        #查詢欄位代碼
DEFINE g_searchstr           STRING                        #查詢欄位字串
DEFINE g_order               STRING                        #查詢排序模式
 
#Browser用
DEFINE g_current_idx         LIKE type_t.num10             #Browser 所在筆數(當下page)
DEFINE g_current_row         LIKE type_t.num10             #Browser 所在筆數(暫存用)
DEFINE g_current_cnt         LIKE type_t.num10             #Browser 總筆數(當下page)
DEFINE g_browser_idx         LIKE type_t.num10             #Browser 所在筆數(所有資料)
DEFINE g_browser_cnt         LIKE type_t.num10             #Browser 總筆數(所有資料)
DEFINE g_row_index           LIKE type_t.num10             #階層樹狀用指標
DEFINE g_add_browse          STRING                        #新增填充用WC
DEFINE g_err_rec             LIKE type_t.chr1              #紀錄是否有錯誤   #161130-00056
 
#190108-00038 ---start---
DEFINE g_simple_mode         LIKE type_t.num5              #是否切換為簡化模式(TRUE:簡化  FALSE:標準)
DEFINE g_simple_name         LIKE dzzc_t.dzzc002           #簡易畫面的名稱
DEFINE g_current_mode        LIKE type_t.chr10             #紀錄目前的畫面是什麼模式(simple:簡化  standard:標準)
DEFINE g_simple_mode_exists  LIKE type_t.chr1              #紀錄是否有簡化模式可供切換(Y/N)
DEFINE gs_first_time         LIKE type_t.chr1              ##紀錄是否為第一次進入init()段
#190108-00038 --- end ---
DEFINE g_autocomplete_param  DYNAMIC ARRAY OF LIKE type_t.chr1000   #autocomplete功能的傳入參數   #190121-00003
DEFINE g_other_wc            LIKE type_t.chr2000           #其他條件   #190121-00003
 
#add-point:自定義客戶專用模組變數(Module Variable) name="global.variable_customerization" 

#end add-point
 
#add-point:傳入參數說明(global.argv) name="global.argv"

#end add-point
 
{</section>}
 
{<section id="azzi085.main" type="s" >}
#應用 a26 樣板自動產生(Version:9)
#+ 作業開始(主程式類型)
MAIN
   #add-point:main段define(客製用) name="main.define_customerization"
   
   #end add-point   
   #add-point:main段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="main.define"
   
   #end add-point   
   
   OPTIONS
   INPUT NO WRAP
   DEFER INTERRUPT
   
   #設定SQL錯誤記錄方式 (模組內定義有效)
   WHENEVER ERROR CALL cl_err_msg_log
       
   #依模組進行系統初始化設定(系統設定)
   CALL cl_ap_init("azz","")
 
   #add-point:作業初始化 name="main.init"
   
   #end add-point
   
   
 
   #LOCK CURSOR (identifier)
   #add-point:SQL_define name="main.define_sql"
   
   #end add-point
   LET g_forupd_sql = " SELECT gzda009,gzda002,gzda008,gzda010,gzda005,gzda006,gzda007,gzda001,gzda003,gzda004,gzdastus,gzdaownid,'',gzdaowndp,'',gzdacrtid,'',gzdacrtdp,'',gzdacrtdt,gzdamodid,'',gzdamoddt", 
                      " FROM gzda_t",
                      " WHERE gzda001=? FOR UPDATE"
   #add-point:SQL_define name="main.after_define_sql"
   
   #end add-point
   LET g_forupd_sql = cl_sql_forupd(g_forupd_sql)                #轉換不同資料庫語法
   LET g_forupd_sql = cl_sql_add_mask(g_forupd_sql)              #遮蔽特定資料
   DECLARE azzi085_cl CURSOR FROM g_forupd_sql                 # LOCK CURSOR
   CALL azzi085_chk_data_exist(g_forupd_sql)   #190806-00019
 
   LET g_sql = " SELECT DISTINCT t0.gzda009,t0.gzda002,t0.gzda008,t0.gzda010,t0.gzda005,t0.gzda006,t0.gzda007,t0.gzda001,t0.gzda003,t0.gzda004,t0.gzdastus,t0.gzdaownid,t0.gzdaowndp,t0.gzdacrtid,t0.gzdacrtdp,t0.gzdacrtdt,t0.gzdamodid,t0.gzdamoddt,t1.ooag011 ,t2.ooefl003 ,t3.ooag011 ,t4.ooefl003 ,t5.ooag011", 
 
               " FROM gzda_t t0",
                              " LEFT JOIN ooag_t t1 ON t1.ooagent="||g_enterprise||" AND t1.ooag001=t0.gzdaownid  ",
               " LEFT JOIN ooefl_t t2 ON t2.ooeflent="||g_enterprise||" AND t2.ooefl001=t0.gzdaowndp AND t2.ooefl002='"||g_dlang||"' ",
               " LEFT JOIN ooag_t t3 ON t3.ooagent="||g_enterprise||" AND t3.ooag001=t0.gzdacrtid  ",
               " LEFT JOIN ooefl_t t4 ON t4.ooeflent="||g_enterprise||" AND t4.ooefl001=t0.gzdacrtdp AND t4.ooefl002='"||g_dlang||"' ",
               " LEFT JOIN ooag_t t5 ON t5.ooagent="||g_enterprise||" AND t5.ooag001=t0.gzdamodid  ",
 
               " WHERE  t0.gzda001 = ?"
   LET g_sql = cl_sql_add_mask(g_sql)              #遮蔽特定資料
   #add-point:SQL_define name="main.after_refresh_sql"
   
   #end add-point
   PREPARE azzi085_master_referesh FROM g_sql
 
    
 
 
   #190108-00038 ---start---
   
   #190108-00038 --- end ---
  
   IF g_bgjob = "Y" THEN
      #add-point:Service Call name="main.servicecall"
      
      #end add-point
   ELSE
      #畫面開啟 (identifier)
      #190108-00038 ---modify start---
      OPEN WINDOW w_azzi085 WITH FORM cl_ap_formpath('azz',g_code)
      #190108-00038 --- modify end ---   
 
      #瀏覽頁簽資料初始化
      CALL cl_ui_init()
   
      #程式初始化
      CALL azzi085_init()   
 
      #進入選單 Menu (="N")
      CALL azzi085_ui_dialog() 
      
      #add-point:畫面關閉前 name="main.before_close"
      
      #end add-point
 
      #畫面關閉
      CLOSE WINDOW w_azzi085
      
   END IF 
   
   CLOSE azzi085_cl
   
   
 
   #add-point:作業離開前 name="main.exit"
   
   #end add-point
 
   #離開作業
   CALL cl_ap_exitprogram("0")
END MAIN
 
 
 
 
{</section>}
 
{<section id="azzi085.init" type="s" >}
#+ 瀏覽頁簽資料初始化
PRIVATE FUNCTION azzi085_init() RETURNS ()   #200213-00032
   #add-point:init段define(客製用) name="init.define_customerization"
   
   #end add-point
   #190108-00038 ---start---
   DEFINE tpl_msg            STRING
   DEFINE tpl_node           om.DomNode
   DEFINE tpl_nodelist       om.NodeList
   DEFINE tpl_node_name      STRING
   DEFINE tpl_node_tagname   STRING
   DEFINE tpl_i              LIKE type_t.num5
   #190108-00038 --- end ---
   #add-point:init段define (請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="init.define"
   
   #end add-point
   
   #add-point:Function前置處理  name="init.pre_function"
   #240709-00026#1 begin
   #240709-00026#1 end
   #end add-point
   
   #定義combobox狀態
      CALL cl_set_combo_scc_part('gzdastus','17','N,Y')
 
      CALL cl_set_combo_scc('gzda005','193') 
 
   LET g_error_show = 1
   LET g_err_rec = "N"   #161130-00056
   LET gwin_curr = ui.Window.getCurrent()
   LET gfrm_curr = gwin_curr.getForm()   
   
   #190108-00038 ---start---
   CALL gfrm_curr.setElementHidden('ui_change',1)   #隱藏畫面切換action
   #190108-00038 --- end ---
  
   #add-point:畫面資料初始化 name="init.init"
   CALL azzi085_set_combo_dbtype() 
   LET g_ins_logd = FALSE            #180312-00040#1 add
   #end add-point
 
   #第一次進入init段時才跑default_search()，以避免後續作畫面轉換時因為重複執行init()段而導致g_wc條件值被異動
   IF cl_null(gs_first_time) OR gs_first_time <> "N" THEN   #190108-00038
      #根據外部參數進行搜尋
      CALL azzi085_default_search()
   END IF   #190108-00038
 
END FUNCTION
 
{</section>}
 
{<section id="azzi085.chk_data_exist" type="s" >}
#+ 異動前先確認該筆資料是否存在
PRIVATE FUNCTION azzi085_chk_data_exist(ps_sql) RETURNS ()   #200213-00032
   #add-point:chk_data_exist段define(客製用) name="chk_data_exist.define_customerization"
   
   #end add-point
   DEFINE ps_sql         STRING
   DEFINE ls_sql         STRING
   DEFINE ls_b           LIKE type_t.num5
   DEFINE ls_e           LIKE type_t.num5
   DEFINE ls_table       LIKE type_t.chr20
   DEFINE ls_wc          STRING
   #add-point:chk_data_exist段define (請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="chk_
   
   #end add-point
 
   #add-point:chk_data_exist段action_chk name="chk_data_exist.action_chk"
   
   #end add-point
 
   #200220-00002 ---modify start---
   #因應多資料庫，需依照各資料庫的語法做截取
   #ORACLE
      LET ls_table = ""
      LET ls_wc = ""
      LET ls_b = ps_sql.getIndexOf("FROM",1)+5
      LET ls_e = ps_sql.getIndexOf("WHERE",ls_b)-1
      LET ls_table = ps_sql.subString(ls_b,ls_e)
 
      LET ls_b = ps_sql.getIndexOf("WHERE",1)
      LET ls_e = ps_sql.getIndexOf(" FOR ",ls_b)
      LET ls_wc = ps_sql.subString(ls_b,ls_e)
 
   #MSV
   #200220-00002 --- modify end ---
 
   LET ls_sql = "SELECT COUNT(1) FROM ",ls_table," ",ls_wc
   PREPARE azzi085_chk_data_exist_pre FROM ls_sql
   IF SQLCA.SQLCODE THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.extend = 'PREPARE azzi085_chk_data_exist_pre ERROR'
      LET g_errparam.code = SQLCA.SQLCODE
      LET g_errparam.popup = TRUE
      CALL cl_err()
   END IF
 
END FUNCTION
 
{</section>}
 
{<section id="azzi085.ui_dialog" type="s" >}
#+ 選單功能實際執行處
PRIVATE FUNCTION azzi085_ui_dialog() RETURNS ()   #200213-00032
   #add-point:ui_dialog段define(客製用) name="ui_dialog.define_customerization"
   
   #end add-point
   DEFINE li_exit   LIKE type_t.num10       #判別是否為離開作業
   DEFINE li_idx    LIKE type_t.num10       #指標變數
   DEFINE ls_wc     STRING                  #wc用
   DEFINE la_param  RECORD                  #程式串查用變數
          prog       STRING,
          actionid   STRING,
          background LIKE type_t.chr1,
          param      DYNAMIC ARRAY OF STRING
                    END RECORD
   DEFINE ls_js     STRING                  #轉換後的json字串
   DEFINE l_cmd_token           base.StringTokenizer   #報表作業cmdrun使用 
   DEFINE l_cmd_next            STRING                 #報表作業cmdrun使用
   DEFINE l_cmd_cnt             LIKE type_t.num5       #報表作業cmdrun使用
   DEFINE l_cmd_prog_arg        STRING                 #報表作業cmdrun使用
   DEFINE l_cmd_arg             STRING                 #報表作業cmdrun使用
   #(ver:53) ---add start---
   DEFINE la_wc      DYNAMIC ARRAY OF RECORD
          tableid    STRING,
          wc         STRING
          END RECORD
   #(ver:53) --- add end ---
   #(ver:56) ---add start---
   DEFINE l_arr_len       INTEGER
   DEFINE l_arr_cnt       INTEGER
   DEFINE lb_get_type     LIKE type_t.chr20
   #(ver:56) --- add end ---
   #add-point:ui_dialog段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="ui_dialog.define"
   
   #end add-point
   
   #add-point:Function前置處理  name="ui_dialog.pre_function"
   
   #end add-point
   
   LET li_exit = FALSE
   LET g_current_row = 0
   LET g_current_idx = 0
 
   #若有外部參數查詢, 則直接顯示資料(隱藏查詢方案)
   IF g_default THEN
      CALL gfrm_curr.setElementHidden("mainlayout",0)
      CALL gfrm_curr.setElementHidden("worksheet",1)
      LET g_main_hidden = 0
   ELSE
      CALL gfrm_curr.setElementHidden("mainlayout",1)
      CALL gfrm_curr.setElementHidden("worksheet",0)
      LET g_main_hidden = 1
   END IF
   
   #action default動作
   #應用 a42 樣板自動產生(Version:3)
   #進入程式時預設執行的動作
   CASE g_actdefault
      WHEN "insert"
         LET g_action_choice="insert"
         LET g_actdefault = ""
         IF cl_auth_chk_act("insert") THEN
            CALL azzi085_insert()
            #add-point:ON ACTION insert name="menu.default.insert"
            
            #END add-point
         END IF
 
      #add-point:action default自訂 name="ui_dialog.action_default"
      
      #end add-point
      OTHERWISE
   END CASE
 
 
 
   
   #add-point:ui_dialog段before dialog  name="ui_dialog.before_dialog"
   
   #end add-point
 
   WHILE li_exit = FALSE
   
      IF g_action_choice = "logistics" THEN
         #清除畫面及相關資料
         CLEAR FORM
         CALL g_browser.clear()       
         INITIALIZE g_gzda_m.* TO NULL
         LET g_wc  = ' 1=2'
         LET g_action_choice = ""
         CALL azzi085_init()
      END IF
 
      #190108-00038 ---start---
      
      #190108-00038 --- end ---
      
      CALL lib_cl_dlg.cl_dlg_before_display()
      CALL cl_notice()
    
      #確保g_current_idx位於正常區間內
      #小於,等於0則指到第1筆
      IF g_current_idx <= 0 THEN
         LET g_current_idx = 1
      END IF
               
      IF g_main_hidden = 0 THEN
         MENU
            BEFORE MENU 
               #先填充browser資料
               #190108-00038 ---modify start---
            #  CALL azzi085_browser_fill(g_wc,"")   #190114-00003 mark
            #  IF g_action_choice <> "ui_change" THEN   #190114-00003
               IF g_action_choice = "ui_change" THEN
                  DISPLAY g_browser_idx TO FORMONLY.b_index   #當下筆數
                  DISPLAY g_browser_cnt TO FORMONLY.b_count   #總筆數
                  DISPLAY g_browser_idx TO FORMONLY.h_index   #當下筆數
                  DISPLAY g_browser_cnt TO FORMONLY.h_count   #總筆數
               ELSE
                  CALL azzi085_browser_fill(g_wc,"")
               END IF
               #190108-00038 --- modify end ---
               CALL cl_navigator_setting(g_current_idx, g_current_cnt)
               
               #還原為原本指定筆數
               IF g_current_row > 0 THEN
                  LET g_current_idx = g_current_row
               END IF
 
               #當每次點任一筆資料都會需要用到  
               IF g_browser_cnt > 0 THEN
                  CALL azzi085_fetch("")   
               END IF               
               #add-point:ui_dialog段 before menu name="ui_dialog.before_menu"
               #170413-00060#2 add(s)
               #只有資料庫編號為dsrept時需執行[重產poijdbc.ini]
               IF g_gzda_m.gzda003 = 'dsrept' THEN
                  CALL cl_set_act_visible("create_ini",TRUE)
               ELSE
                  CALL cl_set_act_visible("create_ini",FALSE)
               END IF
               #170413-00060#2 add(e)
               #210120-00014#2 add(s)
               #只有資料庫編號為 dsdemo 時需執行[重產 multipoijdbc.ini]
               IF g_gzda_m.gzda003 = 'dsdemo' THEN
                  CALL cl_set_act_visible("create_ini_mult",TRUE)
               ELSE
                  CALL cl_set_act_visible("create_ini_mult",FALSE)
               END IF
               #210120-00014#2 add(e)   
               
               
            #210120-00014#2 add(s)
            ON ACTION create_ini_mult  #產生 multipoijdbc.ini
               IF g_gzda_m.gzda003 = 'dsdemo' THEN
                  IF azzi085_create_multipoijdbc() THEN
                     INITIALIZE g_errparam TO NULL
                     LET g_errparam.code = 'azz-01179'
                     LET g_errparam.extend = ' multipoijdbc.ini '
                     LET g_errparam.popup = TRUE
                     CALL cl_err()         
                  END IF
               END IF
            ON ACTION explain_create_ini   #說明如何產生 ini 檔
               INITIALIZE g_errparam TO NULL
               LET g_errparam.code = 'lib-00654'
               LET g_errparam.extend = cl_spreadsheet_ini_path()
               LET g_errparam.popup = TRUE
               CALL cl_err()                    
            #210120-00014#2 add(e)               
               #end add-point
            
            #狀態碼切換
            ON ACTION statechange
               CALL azzi085_statechange()
               LET g_action_choice="statechange"
               #根據資料狀態切換action狀態
               CALL cl_set_act_visible("statechange,modify,delete,reproduce", TRUE)
               CALL azzi085_set_act_visible()
               CALL azzi085_set_act_no_visible()
               IF NOT (g_gzda_m.gzda001 IS NULL
 
                 ) THEN
                  #組合條件
                  LET g_add_browse = " ",
                                     " gzda001 = '", g_gzda_m.gzda001, "' "
 
                  #填到對應位置
                  CALL azzi085_browser_fill(g_wc,"")
               END IF
               
            #第一筆資料
            ON ACTION first
               CALL azzi085_fetch("F") 
               LET g_current_row = g_current_idx
            
            #下一筆資料
            ON ACTION next
               CALL azzi085_fetch("N")
               LET g_current_row = g_current_idx
            
            #指定筆資料
            ON ACTION jump
               CALL azzi085_fetch("/")
               LET g_current_row = g_current_idx
            
            #上一筆資料
            ON ACTION previous
               CALL azzi085_fetch("P")
               LET g_current_row = g_current_idx
            
            #最後筆資料
            ON ACTION last 
               CALL azzi085_fetch("L")  
               LET g_current_row = g_current_idx
            
            #離開程式
            ON ACTION exit
               LET g_action_choice="exit"
               LET INT_FLAG = FALSE
               LET li_exit = TRUE
               EXIT MENU 
            
            #離開程式
            ON ACTION close
               LET g_action_choice="exit"
               LET INT_FLAG = FALSE
               LET li_exit = TRUE
               EXIT MENU
 
            #180226-00040 ---modify start---
            #主頁摺疊
            ON ACTION mainhidden   
               LET g_action_choice = "mainhidden"            
               #(ver:54) ---start---
               #add-point:ui_dialog段 menu段 mainhidden name="ui_dialog.menu_mainhidden"
               
               #end add-point
               #(ver:54) --- end ---
            #180226-00040 --- modify end ---
               #180226-00040 ---modify start---
               IF g_main_hidden THEN
                  CALL gfrm_curr.setElementHidden("mainlayout",0)
                  CALL gfrm_curr.setElementHidden("worksheet",1)
                  LET g_main_hidden = 0
               ELSE
                  CALL gfrm_curr.setElementHidden("mainlayout",1)
                  CALL gfrm_curr.setElementHidden("worksheet",0)
                  LET g_main_hidden = 1
                  CALL cl_notice()
               END IF
               EXIT MENU
               #180226-00040 --- modify end ---
           
            #180226-00040 ---modify start---    
            ON ACTION worksheethidden   #瀏覽頁折疊
               #(ver:54) ---start---
               #add-point:ui_dialog段 menu段 worksheethidden name="ui_dialog.menu_worksheethidden"
               
               #end add-point
               #(ver:54) --- end ---
            #180226-00040 --- modify end ---
               IF g_main_hidden THEN
                  CALL gfrm_curr.setElementHidden("mainlayout",0)
                  CALL gfrm_curr.setElementHidden("worksheet",1)
                  LET g_main_hidden = 0
               ELSE
                  CALL gfrm_curr.setElementHidden("mainlayout",1)
                  CALL gfrm_curr.setElementHidden("worksheet",0)
                  LET g_main_hidden = 1
               END IF
               EXIT MENU
            
            #單頭摺疊，可利用hot key "Alt-s"開啟/關閉單頭
            ON ACTION controls   
               IF g_header_hidden THEN
                  CALL gfrm_curr.setElementHidden("vb_master",0)
                  CALL gfrm_curr.setElementImage("controls","small/arr-u.png")
                  LET g_header_hidden = 0     #visible
               ELSE
                  CALL gfrm_curr.setElementHidden("vb_master",1)
                  CALL gfrm_curr.setElementImage("controls","small/arr-d.png")
                  LET g_header_hidden = 1     #hidden     
               END IF
          
            #查詢方案用
            ON ACTION queryplansel
               CALL cl_dlg_qryplan_select() RETURNING ls_wc
               #不是空條件才寫入g_wc跟重新找資料
               IF NOT cl_null(ls_wc) THEN
                  #(ver:53) ---modify start---
                  CALL util.JSON.parse(ls_wc, la_wc)
                  INITIALIZE g_wc TO NULL
                  FOR li_idx = 1 TO la_wc.getLength()
                     CASE
                        WHEN la_wc[li_idx].tableid = "gzda_t"
                           LET g_wc = la_wc[li_idx].wc
                        #190730-00003#4 -S-
                        WHEN la_wc[li_idx].tableid = "EXTENDWC"
                           LET g_wc = la_wc[li_idx].wc
                        #190730-00003#4 -E-                         
                     END CASE
                  END FOR
                  IF NOT cl_null(g_wc) THEN
                     #取得條件後需要重查、跳到結果第一筆資料的功能程式段
                     CALL azzi085_browser_fill(g_wc,"F")
                     IF g_browser_cnt = 0 THEN
                        INITIALIZE g_errparam TO NULL
                        LET g_errparam.extend = ""
                        LET g_errparam.code   = "-100"
                        LET g_errparam.popup  = TRUE
                        CALL cl_err()
                        CLEAR FORM
                     ELSE
                        CALL azzi085_fetch("F")
                     END IF
                  END IF
                  #(ver:53) --- modify end ---
               END IF
               #重新搜尋會將notice區塊清空,此處不可用EXIT DIALOG, SUBDIALOG重讀會導致部分變數消失
               CALL cl_notice()   #(ver:53) modify
            
            #查詢方案用
            ON ACTION qbe_select
               CALL cl_qbe_list("m") RETURNING ls_wc
               IF NOT cl_null(ls_wc) THEN
                  #(ver:53) ---modify start---
                  CALL util.JSON.parse(ls_wc, la_wc)
                  INITIALIZE g_wc TO NULL
                  FOR li_idx = 1 TO la_wc.getLength()
                     CASE
                        WHEN la_wc[li_idx].tableid = "gzda_t"
                           LET g_wc = la_wc[li_idx].wc
                        #190730-00003#4 -S-
                        WHEN la_wc[li_idx].tableid = "EXTENDWC"
                           LET g_wc = la_wc[li_idx].wc
                        #190730-00003#4 -E-  
                     END CASE
                  END FOR
                  IF NOT cl_null(g_wc) THEN
                     #取得條件後需要重查、跳到結果第一筆資料的功能程式段
                     CALL azzi085_browser_fill(g_wc,"F")
                     IF g_browser_cnt = 0 THEN
                        INITIALIZE g_errparam TO NULL
                        LET g_errparam.extend = ""
                        LET g_errparam.code   = "-100"
                        LET g_errparam.popup  = TRUE
                        CALL cl_err()
                        CLEAR FORM
                     ELSE
                        CALL azzi085_fetch("F")
                     END IF
                  END IF
                  #(ver:53) --- modify end ---
               END IF
               #重新搜尋會將notice區塊清空,此處不可用EXIT DIALOG, SUBDIALOG重讀會導致部分變數消失
               CALL cl_notice()   #(ver:53) modify
            
            
         #應用 a43 樣板自動產生(Version:4)
         ON ACTION modify
            LET g_action_choice="modify"
            IF cl_auth_chk_act("modify") THEN
               LET g_aw = ''
               CALL azzi085_modify()
               #add-point:ON ACTION modify name="menu2.modify"
               #180312-00040#1 --START--
               IF g_ins_logd THEN
                  CALL cl_log_db_start(g_gzda_m.gzda001, "", "")
                  LET g_ins_logd = FALSE
               END IF 
               #180312-00040#1 --END--
               #END add-point
               
            END IF
 
 
 
 
         #應用 a43 樣板自動產生(Version:4)
         ON ACTION output
            LET g_action_choice="output"
            IF cl_auth_chk_act("output") THEN
               
               #add-point:ON ACTION output name="menu2.output"
               
               #END add-point
               
            END IF
 
 
 
 
         #應用 a43 樣板自動產生(Version:4)
         ON ACTION quickprint
            LET g_action_choice="quickprint"
            IF cl_auth_chk_act("quickprint") THEN
               
               #add-point:ON ACTION quickprint name="menu2.quickprint"
               
               #END add-point
               
            END IF
 
 
 
 
         #應用 a43 樣板自動產生(Version:4)
         ON ACTION reproduce
            LET g_action_choice="reproduce"
            IF cl_auth_chk_act("reproduce") THEN
               CALL azzi085_reproduce()
               #add-point:ON ACTION reproduce name="menu2.reproduce"
               
               #END add-point
               
            END IF
 
 
 
 
         #應用 a43 樣板自動產生(Version:4)
         ON ACTION delete
            LET g_action_choice="delete"
            IF cl_auth_chk_act("delete") THEN
               CALL azzi085_delete()
               #add-point:ON ACTION delete name="menu2.delete"
               
               #END add-point
               
            END IF
 
 
 
 
         #應用 a43 樣板自動產生(Version:4)
         ON ACTION insert
            LET g_action_choice="insert"
            IF cl_auth_chk_act("insert") THEN
               CALL azzi085_insert()
               #add-point:ON ACTION insert name="menu2.insert"
               
               #END add-point
               
            END IF
 
 
 
 
         #應用 a43 樣板自動產生(Version:4)
         ON ACTION query
            LET g_action_choice="query"
            IF cl_auth_chk_act("query") THEN
               CALL azzi085_query()
               #add-point:ON ACTION query name="menu2.query"
               
               #END add-point
               
            END IF
 
 
 
 
            
            
 
               #180611-00048
            
            #應用 a46 樣板自動產生(Version:3)
         #新增相關文件
         ON ACTION related_document
            CALL azzi085_set_pk_array()
            IF cl_auth_chk_act("related_document") THEN
               #add-point:ON ACTION related_document name="ui_dialog.menu.related_document"
               
               #END add-point
               CALL cl_doc()
            END IF
            
         ON ACTION agendum
            CALL azzi085_set_pk_array()
            #add-point:ON ACTION agendum name="ui_dialog.menu.agendum"
            
            #END add-point
            CALL cl_user_overview()
            CALL cl_user_overview_set_follow_pic()
         
         ON ACTION followup
            CALL azzi085_set_pk_array()
            #add-point:ON ACTION followup name="ui_dialog.menu.followup"

        #170413-00060#2 add(s)
        ON ACTION create_ini
           IF g_gzda_m.gzda003 = 'dsrept' THEN
              IF cl_spreadsheet_create_ini() THEN
                 INITIALIZE g_errparam TO NULL
                 LET g_errparam.code = 'azz-01179'
                 LET g_errparam.extend = cl_spreadsheet_ini_path()
                 LET g_errparam.popup = TRUE
                 CALL cl_err()         
              END IF
           END IF           
        #170413-00060#2 add(e)
        
            #END add-point
            CALL cl_user_overview_follow('')
 
 
 
 
            #190108-00038 ---start---
            
            #190108-00038 --- end ---
            
            #主選單用ACTION
            &include "main_menu_exit_menu.4gl"
            &include "relating_action.4gl"
            #交談指令共用ACTION
            &include "common_action.4gl"
            
         END MENU
      
      ELSE
      
         DIALOG ATTRIBUTES(UNBUFFERED,FIELD ORDER FORM)
           
      
            #左側瀏覽頁簽
            DISPLAY ARRAY g_browser TO s_browse.* ATTRIBUTE(COUNT=g_rec_b)
            
               BEFORE ROW
                  #回歸舊筆數位置 (回到當時異動的筆數)
                  LET g_current_idx = DIALOG.getCurrentRow("s_browse")
                  IF g_current_idx = 0 THEN
                     LET g_current_idx = 1
                  END IF
                  LET g_current_row = g_current_idx  #目前指標
                  LET g_current_sw = TRUE
                  CALL cl_show_fld_cont()     
                  
                  #當每次點任一筆資料都會需要用到               
                  CALL azzi085_fetch("")
 
               ON ACTION qbefield_user   #欄位隱藏設定 
                  LET g_action_choice="qbefield_user"
                  CALL cl_ui_qbefield_user()
    
               
 
               #180725-00030#1 - S -
               #add-point:ui_dialog段自定義browser_action name="ui_dialog.browser_action"
               
               #end add-point
               #180725-00030#1 - E -
 
            END DISPLAY
 
            #add-point:ui_dialog段自定義display array name="ui_dialog.more_displayarray"
            
            #end add-point
 
            #查詢方案用
            SUBDIALOG lib_cl_dlg.cl_dlg_qryplan
            SUBDIALOG lib_cl_dlg.cl_dlg_relateapps
         
            BEFORE DIALOG
               CALL cl_ap_code_set_action() #171110-00013#1
               #先填充browser資料
               IF g_action_choice <> "mainhidden" THEN
                  #190108-00038 ---modify start---
               #  CALL azzi085_browser_fill(g_wc,"")
                  IF g_action_choice = "ui_change" THEN
                     DISPLAY g_browser_idx TO FORMONLY.b_index #當下筆數
                     DISPLAY g_browser_cnt TO FORMONLY.b_count #總筆數
                     DISPLAY g_browser_idx TO FORMONLY.h_index #當下筆數
                     DISPLAY g_browser_cnt TO FORMONLY.h_count #總筆數
                  ELSE
                     CALL azzi085_browser_fill(g_wc,"")
                  END IF
                  #190108-00038 --- modify end ---
               END IF
               CALL cl_navigator_setting(g_current_idx, g_current_cnt)
               LET g_curr_diag = ui.DIALOG.getCurrent()
               #還原為原本指定筆數
               IF g_current_row > 1 THEN
                  #當刪除最後一筆資料時可能產生錯誤, 進行額外判斷
                  IF g_current_row > g_browser.getLength() THEN
                     LET g_current_row = g_browser.getLength()
                  END IF 
                  LET g_current_idx = g_current_row
                  CALL DIALOG.setCurrentRow("s_browse",g_current_idx)
               END IF
 
               #當每次點任一筆資料都會需要用到  
               IF g_browser_cnt > 0 THEN
                  CALL azzi085_fetch("")   
               END IF          
               CALL cl_notice()
               
               #add-point:ui_dialog段before name="ui_dialog.b_dialog"
 
               #end add-point  
            
            AFTER DIALOG
               #add-point:ui_dialog段 after dialog name="ui_dialog.after_dialog"
               
               #end add-point
            
            #狀態碼切換
            ON ACTION statechange
               CALL azzi085_statechange()
               LET g_action_choice="statechange"
               #根據資料狀態切換action狀態
               CALL cl_set_act_visible("statechange,modify,delete,reproduce", TRUE)
               CALL azzi085_set_act_visible()
               CALL azzi085_set_act_no_visible()
               IF NOT (g_gzda_m.gzda001 IS NULL
 
                 ) THEN
                  #組合條件
                  LET g_add_browse = " ",
                                     " gzda001 = '", g_gzda_m.gzda001, "' "
 
                  #填到對應位置
                  CALL azzi085_browser_fill(g_wc,"")
               END IF
         
            #應用 a49 樣板自動產生(Version:4)
            #過濾瀏覽頁資料
            ON ACTION filter
               LET g_action_choice = "fetch"
               #add-point:filter action name="ui_dialog.action.filter"
               
               #end add-point
               CALL azzi085_filter()
               EXIT DIALOG
 
 
 
            
            #第一筆資料
            ON ACTION first
               CALL azzi085_fetch("F") 
               LET g_current_row = g_current_idx
            
            #下一筆資料
            ON ACTION next
               CALL azzi085_fetch("N")
               LET g_current_row = g_current_idx
         
            #指定筆資料
            ON ACTION jump
               CALL azzi085_fetch("/")
               LET g_current_row = g_current_idx
         
            #上一筆資料
            ON ACTION previous
               CALL azzi085_fetch("P")
               LET g_current_row = g_current_idx
          
            #最後筆資料
            ON ACTION last 
               CALL azzi085_fetch("L")  
               LET g_current_row = g_current_idx
         
            #離開程式
            ON ACTION exit
               LET g_action_choice="exit"
               LET INT_FLAG = FALSE
               LET li_exit = TRUE
               EXIT DIALOG 
         
            #離開程式
            ON ACTION close
               LET g_action_choice="exit"
               LET INT_FLAG = FALSE
               LET li_exit = TRUE
               EXIT DIALOG 
    
            #180226-00040 ---modify start---
            #主頁摺疊
            ON ACTION mainhidden 
               LET g_action_choice = "mainhidden"                
               #(ver:54) ---start---
               #add-point:ui_dialog段 mainhidden name="ui_dialog.mainhidden"
               
               #end add-point
               #(ver:54) --- end ---
            #180226-00040 --- modify end ---
               #180226-00040 ---modify start---
               IF g_main_hidden THEN
                  CALL gfrm_curr.setElementHidden("mainlayout",0)
                  CALL gfrm_curr.setElementHidden("worksheet",1)
                  LET g_main_hidden = 0
               ELSE
                  CALL gfrm_curr.setElementHidden("mainlayout",1)
                  CALL gfrm_curr.setElementHidden("worksheet",0)
                  LET g_main_hidden = 1
                  CALL cl_notice()
               END IF
               #EXIT DIALOG
               #180226-00040 --- modify end ---
               
            ON ACTION worksheethidden   #瀏覽頁折疊
               #(ver:54) ---start---
               #add-point:ui_dialog段 worksheethidden name="ui_dialog.worksheethidden"
               
               #end add-point
               #(ver:54) --- end ---
               IF g_main_hidden THEN
                  CALL gfrm_curr.setElementHidden("mainlayout",0)
                  CALL gfrm_curr.setElementHidden("worksheet",1)
                  LET g_main_hidden = 0
               ELSE
                  CALL gfrm_curr.setElementHidden("mainlayout",1)
                  CALL gfrm_curr.setElementHidden("worksheet",0)
                  LET g_main_hidden = 1
               END IF
               #EXIT DIALOG
         
            ON ACTION exporttoexcel
               LET g_action_choice="exporttoexcel"
               IF cl_auth_chk_act("exporttoexcel") THEN
                  #(ver:56) ---start---
                  #因應匯出excel規則調整，需同時符合新舊做法
                  LET lb_get_type = cl_get_exporttoexcel_type()
                  #add-point:ON ACTION exporttoexcel name="ui_dialog.exporttoexcel_type_change"
                  
                  #END add-point
                  DISPLAY "lb_get_type:",lb_get_type
 
                  CASE
                     WHEN lb_get_type = "poi"   #使用poi樣板
                        #browser
                        CALL g_export_node.clear()
                        LET l_arr_len = g_browser.getLength()
                        LET g_export_tabname[1] = "s_browse"
                        LET g_export_tag[1] = rec_arr_type.create(l_arr_len)
                        FOR l_arr_cnt = 1 TO l_arr_len
                           LET g_export_tag[1][l_arr_cnt] = g_browser[l_arr_cnt]
                        END FOR
 
                        #add-point:ON ACTION exporttoexcel name="ui_dialog.exporttoexcel_browser_poi"
                        
                        #END add-point
                        CALL cl_export_poi()
                        #add-point:ON ACTION exporttoexcel name="ui_dialog.exporttoexcel_browser_after_poi"
                        
                        #END add-point
 
                     OTHERWISE   #使用原始樣板
                        #browser
                        CALL g_export_node.clear()
                        LET g_export_node[1] = base.typeInfo.create(g_browser)
                        LET g_export_id[1]   = "s_browse"
                        
                        #add-point:ON ACTION exporttoexcel name="ui_dialog.exporttoexcel_browser"   #(ver:52)
                        
                        #END add-point
                        CALL cl_export_to_excel()
                        #add-point:ON ACTION exporttoexcel name="ui_dialog.exporttoexcel_browser_after"   #(ver:52)
                        
                        #END add-point
                  END CASE
                  #(ver:56) --- end ---
               END IF
         
            #單頭摺疊，可利用hot key "Alt-s"開啟/關閉單頭
            ON ACTION controls   
               IF g_header_hidden THEN
                  CALL gfrm_curr.setElementHidden("vb_master",0)
                  CALL gfrm_curr.setElementImage("controls","small/arr-u.png")
                  LET g_header_hidden = 0     #visible
               ELSE
                  CALL gfrm_curr.setElementHidden("vb_master",1)
                  CALL gfrm_curr.setElementImage("controls","small/arr-d.png")
                  LET g_header_hidden = 1     #hidden     
               END IF
 
            
            #查詢方案用
            ON ACTION queryplansel
               CALL cl_dlg_qryplan_select() RETURNING ls_wc
               #不是空條件才寫入g_wc跟重新找資料
               IF NOT cl_null(ls_wc) THEN
                  #(ver:53) ---modify start---
                  CALL util.JSON.parse(ls_wc, la_wc)
                  INITIALIZE g_wc TO NULL
                  FOR li_idx = 1 TO la_wc.getLength()
                     CASE
                        WHEN la_wc[li_idx].tableid = "gzda_t"
                           LET g_wc = la_wc[li_idx].wc
                        #190730-00003#4 -S-
                        WHEN la_wc[li_idx].tableid = "EXTENDWC"
                           LET g_wc = la_wc[li_idx].wc
                        #190730-00003#4 -E-  
                     END CASE
                  END FOR
                  IF NOT cl_null(g_wc) THEN
                     #取得條件後需要重查、跳到結果第一筆資料的功能程式段
                     CALL azzi085_browser_fill(g_wc,"F")
                     IF g_browser_cnt = 0 THEN
                        INITIALIZE g_errparam TO NULL
                        LET g_errparam.extend = ""
                        LET g_errparam.code   = "-100"
                        LET g_errparam.popup  = TRUE
                        CALL cl_err()
                        CLEAR FORM
                     ELSE
                        CALL azzi085_fetch("F")
                     END IF
                  END IF
                  #(ver:53) --- modify end ---
               END IF
               #重新搜尋會將notice區塊清空,此處不可用EXIT DIALOG, SUBDIALOG重讀會導致部分變數消失
               CALL cl_notice()   #(ver:53) modify
            
            #查詢方案用
            ON ACTION qbe_select
               CALL cl_qbe_list("m") RETURNING ls_wc
               IF NOT cl_null(ls_wc) THEN
                  #(ver:53) ---modify start---
                  CALL util.JSON.parse(ls_wc, la_wc)
                  INITIALIZE g_wc TO NULL
                  FOR li_idx = 1 TO la_wc.getLength()
                     CASE
                        WHEN la_wc[li_idx].tableid = "gzda_t"
                           LET g_wc = la_wc[li_idx].wc
                        #190730-00003#4 -S-
                        WHEN la_wc[li_idx].tableid = "EXTENDWC"
                           LET g_wc = la_wc[li_idx].wc
                        #190730-00003#4 -E-  
                     END CASE
                  END FOR
                  IF NOT cl_null(g_wc) THEN
                     #取得條件後需要重查、跳到結果第一筆資料的功能程式段
                     CALL azzi085_browser_fill(g_wc,"F")
                     IF g_browser_cnt = 0 THEN
                        INITIALIZE g_errparam TO NULL
                        LET g_errparam.extend = ""
                        LET g_errparam.code   = "-100"
                        LET g_errparam.popup  = TRUE
                        CALL cl_err()
                        CLEAR FORM
                     ELSE
                        CALL azzi085_fetch("F")
                     END IF
                  END IF
                  #(ver:53) --- modify end ---
               END IF
               #重新搜尋會將notice區塊清空,此處不可用EXIT DIALOG, SUBDIALOG重讀會導致部分變數消失
               CALL cl_notice()   #(ver:53) modify
               
            
         #應用 a43 樣板自動產生(Version:4)
         ON ACTION modify
            LET g_action_choice="modify"
            IF cl_auth_chk_act("modify") THEN
               LET g_aw = ''
               CALL azzi085_modify()
               #add-point:ON ACTION modify name="menu.modify"
               #180312-00040#1 --START--
               IF g_ins_logd THEN
                  CALL cl_log_db_start(g_gzda_m.gzda001, "", "")
                  LET g_ins_logd = FALSE
               END IF 
               #180312-00040#1 --END--
               #END add-point
               
            END IF
 
 
 
 
         #應用 a43 樣板自動產生(Version:4)
         ON ACTION output
            LET g_action_choice="output"
            IF cl_auth_chk_act("output") THEN
               
               #add-point:ON ACTION output name="menu.output"
               
               #END add-point
               
            END IF
 
 
 
 
         #應用 a43 樣板自動產生(Version:4)
         ON ACTION quickprint
            LET g_action_choice="quickprint"
            IF cl_auth_chk_act("quickprint") THEN
               
               #add-point:ON ACTION quickprint name="menu.quickprint"
               
               #END add-point
               
            END IF
 
 
 
 
         #應用 a43 樣板自動產生(Version:4)
         ON ACTION reproduce
            LET g_action_choice="reproduce"
            IF cl_auth_chk_act("reproduce") THEN
               CALL azzi085_reproduce()
               #add-point:ON ACTION reproduce name="menu.reproduce"
               
               #END add-point
               
            END IF
 
 
 
 
         #應用 a43 樣板自動產生(Version:4)
         ON ACTION delete
            LET g_action_choice="delete"
            IF cl_auth_chk_act("delete") THEN
               CALL azzi085_delete()
               #add-point:ON ACTION delete name="menu.delete"
               
               #END add-point
               
            END IF
 
 
 
 
         #應用 a43 樣板自動產生(Version:4)
         ON ACTION insert
            LET g_action_choice="insert"
            IF cl_auth_chk_act("insert") THEN
               CALL azzi085_insert()
               #add-point:ON ACTION insert name="menu.insert"
               
               #END add-point
               
            END IF
 
 
 
 
         #應用 a43 樣板自動產生(Version:4)
         ON ACTION query
            LET g_action_choice="query"
            IF cl_auth_chk_act("query") THEN
               CALL azzi085_query()
               #add-point:ON ACTION query name="menu.query"
               
               #END add-point
               
            END IF
 
 
 
 
            
            
 
               #180611-00048
 
            #應用 a46 樣板自動產生(Version:3)
         #新增相關文件
         ON ACTION related_document
            CALL azzi085_set_pk_array()
            IF cl_auth_chk_act("related_document") THEN
               #add-point:ON ACTION related_document name="ui_dialog.dialog.related_document"
               
               #END add-point
               CALL cl_doc()
            END IF
            
         ON ACTION agendum
            CALL azzi085_set_pk_array()
            #add-point:ON ACTION agendum name="ui_dialog.dialog.agendum"
            
            #END add-point
            CALL cl_user_overview()
            CALL cl_user_overview_set_follow_pic()
         
         ON ACTION followup
            CALL azzi085_set_pk_array()
            #add-point:ON ACTION followup name="ui_dialog.dialog.followup"
        
            #END add-point
            CALL cl_user_overview_follow('')
 
 
 
            
            #190108-00038 ---start---
            
            #190108-00038 --- end ---
 
            #主選單用ACTION
            &include "main_menu_exit_dialog.4gl"
            &include "relating_action.4gl"
            #交談指令共用ACTION
            &include "common_action.4gl"
            
         END DIALOG 
      
      END IF
      
      #add-point:ui_dialog段 after dialog name="ui_dialog.exit_dialog"
      
      #end add-point
      
      #(ver:50) ---start---
      IF li_exit THEN
         #add-point:ui_dialog段離開dialog前 name="ui_dialog.b_exit"
         
         #end add-point
         EXIT WHILE
      END IF
      #(ver:50) --- end ---
 
   END WHILE
 
END FUNCTION
 
{</section>}
 
{<section id="azzi085.browser_fill" type="s" >}
#應用 a29 樣板自動產生(Version:18)
#+ 瀏覽頁簽資料填充(一般單檔)
PRIVATE FUNCTION azzi085_browser_fill(p_wc,ps_page_action) RETURNS ()   #200213-00032
   #add-point:browser_fill段define name="browser_fill.define_customerization"
   
   #end add-point
   DEFINE p_wc              STRING
   DEFINE ls_wc             STRING
   DEFINE ps_page_action    STRING
   DEFINE l_searchcol       STRING
   DEFINE l_sql             STRING
   DEFINE l_sql_rank        STRING
   #add-point:browser_fill段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="browser_fill.define"
   
   #end add-point
   
   #add-point:Function前置處理  name="browser_fill.pre_function"
   
   #end add-point
   
   LET l_searchcol = "gzda001"
 
   LET p_wc = p_wc.trim() #當查詢按下Q時 按下放棄 g_wc = "  " 所以要清掉空白
   IF cl_null(p_wc) THEN  #p_wc 查詢條件
      LET p_wc = " 1=1 " 
   END IF
   #add-point:browser_fill段wc控制 name="browser_fill.wc"
   
   #end add-point
 
   LET g_sql = " SELECT COUNT(1) FROM gzda_t ",
               "  ",
               "  ",
               " WHERE  ", 
               p_wc CLIPPED, cl_sql_add_filter("gzda_t")
                
   #add-point:browser_fill段cnt_sql name="browser_fill.cnt_sql"
   
   #end add-point
                
   IF g_sql.getIndexOf(" 1=2",1) THEN
      DISPLAY "INFO: 1=2 jumped!"
   ELSE
      PREPARE header_cnt_pre FROM g_sql
      EXECUTE header_cnt_pre INTO g_browser_cnt
      FREE header_cnt_pre 
   END IF
   
   #若超過最大顯示筆數
   IF g_browser_cnt > g_max_browse THEN
      #若是因為切換畫面的話，不用再次顯示超過最大顯示筆數的訊息
      IF cl_null(g_action_choice) OR g_action_choice <> "ui_change" THEN   #190108-00038
         IF g_error_show = 1 THEN
            INITIALIZE g_errparam TO NULL 
            LET g_errparam.extend = g_browser_cnt 
            LET g_errparam.code   = 9035
            LET g_errparam.popup  = TRUE 
            CALL cl_err()
         END IF
      END IF   #190108-00038
   END IF
   
   LET g_error_show = 0
   
   IF ps_page_action = "F" OR
      ps_page_action = "P"  OR
      ps_page_action = "N"  OR
      ps_page_action = "L"  THEN
      LET g_page_action = ps_page_action
   END IF
   
   IF cl_null(g_add_browse) THEN
      #清除畫面
      CLEAR FORM
      INITIALIZE g_gzda_m.* TO NULL
      CALL g_browser.clear()
      LET g_cnt = 1
      LET ls_wc = p_wc
   ELSE
      LET ls_wc = g_add_browse
      LET g_cnt = g_current_idx
   END IF
   
   LET g_sql = " SELECT t0.gzdastus,t0.gzda001,t0.gzda008",
               " FROM gzda_t t0 ",
               "  ",
               
               " WHERE  ", ls_wc, cl_sql_add_filter("gzda_t")
   #add-point:browser_fill段fill_wc name="browser_fill.fill_wc"
   
   #end add-point 
   LET g_sql = g_sql, " ORDER BY ",l_searchcol," ",g_order
   #add-point:browser_fill段before_pre name="browser_fill.before_pre"
   
   #end add-point                    
 
   #LET g_sql = cl_sql_add_tabid(g_sql,"gzda_t")             #WC重組
   LET g_sql = cl_sql_add_mask(g_sql)              #遮蔽特定資料
   
   IF g_sql.getIndexOf(" 1=2",1) THEN
      DISPLAY "INFO: 1=2 jumped!"
   ELSE
      PREPARE browse_pre FROM g_sql
      DECLARE browse_cur CURSOR FOR browse_pre
      
      FOREACH browse_cur INTO g_browser[g_cnt].b_statepic,g_browser[g_cnt].b_gzda001,g_browser[g_cnt].b_gzda008
         IF SQLCA.sqlcode THEN
            INITIALIZE g_errparam TO NULL 
            LET g_errparam.extend = "foreach:" 
            LET g_errparam.code   = SQLCA.sqlcode 
            LET g_errparam.popup  = TRUE 
            CALL cl_err()
            EXIT FOREACH
         END IF
         
         
         
         #add-point:browser_fill段reference name="browser_fill.reference"
         
         #end add-point
         
         #遮罩相關處理
         CALL azzi085_browser_mask()
         
               #應用 a24 樣板自動產生(Version:3)
      #browser顯示圖片
      CASE g_browser[g_cnt].b_statepic
         WHEN "N"
            LET g_browser[g_cnt].b_statepic = "stus/16/inactive.png"
         WHEN "Y"
            LET g_browser[g_cnt].b_statepic = "stus/16/active.png"
         
      END CASE
 
 
 
         LET g_cnt = g_cnt + 1
         #IF g_cnt > g_max_rec THEN #180330-00002#1
         IF g_cnt > g_max_browse THEN  #180330-00002#1
            EXIT FOREACH
         END IF
      END FOREACH
 
      FREE browse_pre
 
   END IF
 
   #清空g_add_browse, 並指定指標位置
   IF NOT cl_null(g_add_browse) THEN
      LET g_add_browse = ""
      CALL g_curr_diag.setCurrentRow("s_browse",g_current_idx)
   END IF
   
   IF cl_null(g_browser[g_cnt].b_gzda001) THEN
      CALL g_browser.deleteElement(g_cnt)
   END IF
   
   LET g_header_cnt = g_browser.getLength()
   LET g_current_cnt = g_browser.getLength()
   LET g_browser_cnt = g_browser.getLength()
   LET g_rec_b = g_browser.getLength()
   LET g_cnt = 0
   DISPLAY g_browser_cnt TO FORMONLY.b_count
   DISPLAY g_browser_cnt TO FORMONLY.h_count
   
   #若無資料則關閉相關功能
   IF g_browser_cnt = 0 THEN
      CALL cl_set_act_visible("statechange,modify,delete,reproduce,mainhidden", FALSE)
      CALL cl_navigator_setting(0,0)
   ELSE
      CALL cl_set_act_visible("mainhidden", TRUE)
   END IF
   
   #add-point:browser_fill段結束前 name="browser_fill.after"
   
   #end add-point   
   
END FUNCTION
 
 
 
 
{</section>}
 
{<section id="azzi085.construct" type="s" >}
#+ QBE資料查詢
PRIVATE FUNCTION azzi085_construct() RETURNS ()   #200213-00032
   #add-point:cs段define(客製用) name="cs.define_customerization"
   
   #end add-point
   DEFINE ls_return      STRING
   DEFINE ls_result      STRING 
   DEFINE ls_wc          STRING 
   #(ver:53) ---add start---
   DEFINE la_wc       DYNAMIC ARRAY OF RECORD
          tableid     STRING,
          wc          STRING
          END RECORD
   DEFINE li_idx      LIKE type_t.num10
   DEFINE lp_str      STRING   #201109-00009
   #(ver:53) --- add end ---
   #add-point:cs段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="cs.define"
   
   #end add-point
   
   #add-point:Function前置處理  name="cs.pre_function"
   
   #end add-point
   
   #清空畫面&資料初始化
   CLEAR FORM
   INITIALIZE g_gzda_m.* TO NULL
   INITIALIZE g_wc TO NULL
   LET g_current_row = 1
 
   LET g_qryparam.state = "c"
 
   DIALOG ATTRIBUTES(UNBUFFERED,FIELD ORDER FORM)
   
      #螢幕上取條件
      CONSTRUCT BY NAME g_wc ON gzda009,gzda002,gzda008,gzda010,gzda005,gzda006,gzda007,gzda001,gzda003,gzda004,gzdastus,gzdaownid,gzdaowndp,gzdacrtid,gzdacrtdp,gzdacrtdt,gzdamodid,gzdamoddt
      
         BEFORE CONSTRUCT                                    
            #add-point:cs段more_construct name="cs.before_construct"
            
            #end add-point             
      
         #公用欄位開窗相關處理
         #應用 a11 樣板自動產生(Version:3)
         #共用欄位查詢處理  
         ##----<<gzdacrtdt>>----
         AFTER FIELD gzdacrtdt
            CALL FGL_DIALOG_GETBUFFER() RETURNING ls_result
            IF NOT cl_null(ls_result) THEN
               IF NOT cl_chk_date_symbol(ls_result) THEN
                  LET ls_result = cl_add_date_extra_cond(ls_result)
               END IF
            END IF
            CALL FGL_DIALOG_SETBUFFER(ls_result)
 
         #----<<gzdamoddt>>----
         AFTER FIELD gzdamoddt
            CALL FGL_DIALOG_GETBUFFER() RETURNING ls_result
            IF NOT cl_null(ls_result) THEN
               IF NOT cl_chk_date_symbol(ls_result) THEN
                  LET ls_result = cl_add_date_extra_cond(ls_result)
               END IF
            END IF
            CALL FGL_DIALOG_SETBUFFER(ls_result)
         
         #----<<gzdacnfdt>>----
         
         #----<<gzdapstdt>>----
 
 
 
      
         #一般欄位
                  #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda009
            #add-point:BEFORE FIELD gzda009 name="construct.b.gzda009"
            
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda009
            
            #add-point:AFTER FIELD gzda009 name="construct.a.gzda009"
            
            #END add-point
                        #應用 a77 樣板自動產生(Version:1)
            CALL FGL_DIALOG_GETBUFFER() RETURNING lp_str
            IF NOT cl_null(lp_str) THEN
               LET lp_str = cl_ap_code_fuzzyquery(lp_str)   #模糊查詢功能
               DISPLAY lp_str TO gzda009
            END IF
 
 
 
 
         #Ctrlp:construct.c.gzda009
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda009
            #add-point:ON ACTION controlp INFIELD gzda009 name="construct.c.gzda009"
            
            #END add-point
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda002
            #add-point:BEFORE FIELD gzda002 name="construct.b.gzda002"
            
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda002
            
            #add-point:AFTER FIELD gzda002 name="construct.a.gzda002"
            
            #END add-point
                        #應用 a77 樣板自動產生(Version:1)
            CALL FGL_DIALOG_GETBUFFER() RETURNING lp_str
            IF NOT cl_null(lp_str) THEN
               LET lp_str = cl_ap_code_fuzzyquery(lp_str)   #模糊查詢功能
               DISPLAY lp_str TO gzda002
            END IF
 
 
 
 
         #Ctrlp:construct.c.gzda002
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda002
            #add-point:ON ACTION controlp INFIELD gzda002 name="construct.c.gzda002"
            
            #END add-point
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda008
            #add-point:BEFORE FIELD gzda008 name="construct.b.gzda008"
            
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda008
            
            #add-point:AFTER FIELD gzda008 name="construct.a.gzda008"
            
            #END add-point
                        #應用 a77 樣板自動產生(Version:1)
            CALL FGL_DIALOG_GETBUFFER() RETURNING lp_str
            IF NOT cl_null(lp_str) THEN
               LET lp_str = cl_ap_code_fuzzyquery(lp_str)   #模糊查詢功能
               DISPLAY lp_str TO gzda008
            END IF
 
 
 
 
         #Ctrlp:construct.c.gzda008
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda008
            #add-point:ON ACTION controlp INFIELD gzda008 name="construct.c.gzda008"
            
            #END add-point
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda010
            #add-point:BEFORE FIELD gzda010 name="construct.b.gzda010"
            
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda010
            
            #add-point:AFTER FIELD gzda010 name="construct.a.gzda010"
            
            #END add-point
            
 
 
         #Ctrlp:construct.c.gzda010
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda010
            #add-point:ON ACTION controlp INFIELD gzda010 name="construct.c.gzda010"
            
            #END add-point
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda005
            #add-point:BEFORE FIELD gzda005 name="construct.b.gzda005"
            
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda005
            
            #add-point:AFTER FIELD gzda005 name="construct.a.gzda005"
            
            #END add-point
            
 
 
         #Ctrlp:construct.c.gzda005
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda005
            #add-point:ON ACTION controlp INFIELD gzda005 name="construct.c.gzda005"
            
            #END add-point
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda006
            #add-point:BEFORE FIELD gzda006 name="construct.b.gzda006"
            
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda006
            
            #add-point:AFTER FIELD gzda006 name="construct.a.gzda006"
            
            #END add-point
            
 
 
         #Ctrlp:construct.c.gzda006
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda006
            #add-point:ON ACTION controlp INFIELD gzda006 name="construct.c.gzda006"
            
            #END add-point
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda007
            #add-point:BEFORE FIELD gzda007 name="construct.b.gzda007"
            
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda007
            
            #add-point:AFTER FIELD gzda007 name="construct.a.gzda007"
            
            #END add-point
            
 
 
         #Ctrlp:construct.c.gzda007
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda007
            #add-point:ON ACTION controlp INFIELD gzda007 name="construct.c.gzda007"
            
            #END add-point
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda001
            #add-point:BEFORE FIELD gzda001 name="construct.b.gzda001"
            
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda001
            
            #add-point:AFTER FIELD gzda001 name="construct.a.gzda001"
            
            #END add-point
                        #應用 a77 樣板自動產生(Version:1)
            CALL FGL_DIALOG_GETBUFFER() RETURNING lp_str
            IF NOT cl_null(lp_str) THEN
               LET lp_str = cl_ap_code_fuzzyquery(lp_str)   #模糊查詢功能
               DISPLAY lp_str TO gzda001
            END IF
 
 
 
 
         #Ctrlp:construct.c.gzda001
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda001
            #add-point:ON ACTION controlp INFIELD gzda001 name="construct.c.gzda001"
            
            #END add-point
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda003
            #add-point:BEFORE FIELD gzda003 name="construct.b.gzda003"
            
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda003
            
            #add-point:AFTER FIELD gzda003 name="construct.a.gzda003"
            
            #END add-point
                        #應用 a77 樣板自動產生(Version:1)
            CALL FGL_DIALOG_GETBUFFER() RETURNING lp_str
            IF NOT cl_null(lp_str) THEN
               LET lp_str = cl_ap_code_fuzzyquery(lp_str)   #模糊查詢功能
               DISPLAY lp_str TO gzda003
            END IF
 
 
 
 
         #Ctrlp:construct.c.gzda003
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda003
            #add-point:ON ACTION controlp INFIELD gzda003 name="construct.c.gzda003"
            
            #END add-point
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda004
            #add-point:BEFORE FIELD gzda004 name="construct.b.gzda004"
            
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda004
            
            #add-point:AFTER FIELD gzda004 name="construct.a.gzda004"
            
            #END add-point
                        #應用 a77 樣板自動產生(Version:1)
            CALL FGL_DIALOG_GETBUFFER() RETURNING lp_str
            IF NOT cl_null(lp_str) THEN
               LET lp_str = cl_ap_code_fuzzyquery(lp_str)   #模糊查詢功能
               DISPLAY lp_str TO gzda004
            END IF
 
 
 
 
         #Ctrlp:construct.c.gzda004
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda004
            #add-point:ON ACTION controlp INFIELD gzda004 name="construct.c.gzda004"
            
            #END add-point
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzdastus
            #add-point:BEFORE FIELD gzdastus name="construct.b.gzdastus"
            
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzdastus
            
            #add-point:AFTER FIELD gzdastus name="construct.a.gzdastus"
            
            #END add-point
            
 
 
         #Ctrlp:construct.c.gzdastus
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzdastus
            #add-point:ON ACTION controlp INFIELD gzdastus name="construct.c.gzdastus"
            
            #END add-point
 
 
         #Ctrlp:construct.c.gzdaownid
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzdaownid
            #add-point:ON ACTION controlp INFIELD gzdaownid name="construct.c.gzdaownid"
            #此段落由子樣板a08產生
            #開窗c段
            LET g_qryparam.reqry = FALSE
            CALL q_ooag001()                           #呼叫開窗
            DISPLAY g_qryparam.return1 TO gzdaownid  #顯示到畫面上

            NEXT FIELD gzdaownid                     #返回原欄位


            #END add-point
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzdaownid
            #add-point:BEFORE FIELD gzdaownid name="construct.b.gzdaownid"
            
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzdaownid
            
            #add-point:AFTER FIELD gzdaownid name="construct.a.gzdaownid"
            
            #END add-point
                        #應用 a77 樣板自動產生(Version:1)
            CALL FGL_DIALOG_GETBUFFER() RETURNING lp_str
            IF NOT cl_null(lp_str) THEN
               LET lp_str = cl_ap_code_fuzzyquery(lp_str)   #模糊查詢功能
               DISPLAY lp_str TO gzdaownid
            END IF
 
 
 
 
         #Ctrlp:construct.c.gzdaowndp
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzdaowndp
            #add-point:ON ACTION controlp INFIELD gzdaowndp name="construct.c.gzdaowndp"
            #此段落由子樣板a08產生
            #開窗c段
            LET g_qryparam.reqry = FALSE
            CALL q_ooeg001_9()                           #呼叫開窗
            DISPLAY g_qryparam.return1 TO gzdaowndp  #顯示到畫面上

            NEXT FIELD gzdaowndp                     #返回原欄位


            #END add-point
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzdaowndp
            #add-point:BEFORE FIELD gzdaowndp name="construct.b.gzdaowndp"
            
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzdaowndp
            
            #add-point:AFTER FIELD gzdaowndp name="construct.a.gzdaowndp"
            
            #END add-point
                        #應用 a77 樣板自動產生(Version:1)
            CALL FGL_DIALOG_GETBUFFER() RETURNING lp_str
            IF NOT cl_null(lp_str) THEN
               LET lp_str = cl_ap_code_fuzzyquery(lp_str)   #模糊查詢功能
               DISPLAY lp_str TO gzdaowndp
            END IF
 
 
 
 
         #Ctrlp:construct.c.gzdacrtid
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzdacrtid
            #add-point:ON ACTION controlp INFIELD gzdacrtid name="construct.c.gzdacrtid"
            #此段落由子樣板a08產生
            #開窗c段
            LET g_qryparam.reqry = FALSE
            CALL q_ooag001()                           #呼叫開窗
            DISPLAY g_qryparam.return1 TO gzdacrtid  #顯示到畫面上

            NEXT FIELD gzdacrtid                     #返回原欄位


            #END add-point
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzdacrtid
            #add-point:BEFORE FIELD gzdacrtid name="construct.b.gzdacrtid"
            
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzdacrtid
            
            #add-point:AFTER FIELD gzdacrtid name="construct.a.gzdacrtid"
            
            #END add-point
                        #應用 a77 樣板自動產生(Version:1)
            CALL FGL_DIALOG_GETBUFFER() RETURNING lp_str
            IF NOT cl_null(lp_str) THEN
               LET lp_str = cl_ap_code_fuzzyquery(lp_str)   #模糊查詢功能
               DISPLAY lp_str TO gzdacrtid
            END IF
 
 
 
 
         #Ctrlp:construct.c.gzdacrtdp
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzdacrtdp
            #add-point:ON ACTION controlp INFIELD gzdacrtdp name="construct.c.gzdacrtdp"
            #此段落由子樣板a08產生
            #開窗c段
            LET g_qryparam.reqry = FALSE
            CALL q_ooeg001_9()                           #呼叫開窗
            DISPLAY g_qryparam.return1 TO gzdacrtdp  #顯示到畫面上

            NEXT FIELD gzdacrtdp                     #返回原欄位


            #END add-point
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzdacrtdp
            #add-point:BEFORE FIELD gzdacrtdp name="construct.b.gzdacrtdp"
            
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzdacrtdp
            
            #add-point:AFTER FIELD gzdacrtdp name="construct.a.gzdacrtdp"
            
            #END add-point
                        #應用 a77 樣板自動產生(Version:1)
            CALL FGL_DIALOG_GETBUFFER() RETURNING lp_str
            IF NOT cl_null(lp_str) THEN
               LET lp_str = cl_ap_code_fuzzyquery(lp_str)   #模糊查詢功能
               DISPLAY lp_str TO gzdacrtdp
            END IF
 
 
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzdacrtdt
            #add-point:BEFORE FIELD gzdacrtdt name="construct.b.gzdacrtdt"
            
            #END add-point
 
 
         #Ctrlp:construct.c.gzdamodid
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzdamodid
            #add-point:ON ACTION controlp INFIELD gzdamodid name="construct.c.gzdamodid"
            #此段落由子樣板a08產生
            #開窗c段
            LET g_qryparam.reqry = FALSE
            CALL q_ooag001()                           #呼叫開窗
            DISPLAY g_qryparam.return1 TO gzdamodid  #顯示到畫面上

            NEXT FIELD gzdamodid                     #返回原欄位


            #END add-point
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzdamodid
            #add-point:BEFORE FIELD gzdamodid name="construct.b.gzdamodid"
            
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzdamodid
            
            #add-point:AFTER FIELD gzdamodid name="construct.a.gzdamodid"
            
            #END add-point
                        #應用 a77 樣板自動產生(Version:1)
            CALL FGL_DIALOG_GETBUFFER() RETURNING lp_str
            IF NOT cl_null(lp_str) THEN
               LET lp_str = cl_ap_code_fuzzyquery(lp_str)   #模糊查詢功能
               DISPLAY lp_str TO gzdamodid
            END IF
 
 
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzdamoddt
            #add-point:BEFORE FIELD gzdamoddt name="construct.b.gzdamoddt"
            
            #END add-point
 
 
 
           
      END CONSTRUCT
      
      #add-point:cs段more_construct name="cs.more_construct"
      
      #end add-point   
      
      BEFORE DIALOG
         CALL cl_qbe_init()
         #add-point:cs段b_dialog name="cs.b_dialog"
         
         #end add-point  
      
      ON ACTION accept
         CALL cl_qbe_auto_save() #180530-00049#1
         ACCEPT DIALOG
 
      ON ACTION cancel
         LET INT_FLAG = 1
         EXIT DIALOG
 
      #查詢方案列表
      ON ACTION qbe_select
         LET ls_wc = ""
         CLEAR FORM #200326-00032#1
         CALL cl_qbe_list("c") RETURNING ls_wc
         #(ver:53) ---add start---
         IF NOT cl_null(ls_wc) THEN
            CALL util.JSON.parse(ls_wc, la_wc)
            INITIALIZE g_wc TO NULL
            FOR li_idx = 1 TO la_wc.getLength()
               CASE
                  WHEN la_wc[li_idx].tableid = "gzda_t"
                     LET g_wc = la_wc[li_idx].wc
               END CASE
            END FOR
         END IF
         #(ver:53) --- add end ---
    
      #條件儲存為方案
      ON ACTION qbe_save
         CALL cl_qbe_save()
 
      #交談指令共用ACTION
      &include "common_action.4gl"
         CONTINUE DIALOG
   END DIALOG
  
   #add-point:cs段after_construct name="cs.after_construct"
   
   #end add-point
  
END FUNCTION
 
{</section>}
 
{<section id="azzi085.filter" type="s" >}
#應用 a50 樣板自動產生(Version:11)
#+ filter過濾功能
PRIVATE FUNCTION azzi085_filter() RETURNS ()   #200213-00032
   #add-point:filter段define name="filter.define_customerization"
   
   #end add-point   
   #add-point:filter段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="filter.define"
   
   #end add-point   
   
   #add-point:Function前置處理  name="filter.pre_function"
   
   #end add-point
   
   #切換畫面
   IF NOT g_main_hidden THEN
      CALL gfrm_curr.setElementHidden("mainlayout",1)
      CALL gfrm_curr.setElementHidden("worksheet",0)
      LET g_main_hidden = 1
   END IF   
 
   LET INT_FLAG = 0
 
   LET g_qryparam.state = 'c'
 
   LET g_wc_filter_t = g_wc_filter.trim()
   LET g_wc_t = g_wc
 
   LET g_wc = cl_replace_str(g_wc, g_wc_filter_t, '')
 
   #使用DIALOG包住 單頭CONSTRUCT及單身CONSTRUCT
   DIALOG ATTRIBUTES(UNBUFFERED,FIELD ORDER FORM)
 
      #單頭
      CONSTRUCT g_wc_filter ON gzda001,gzda008
                          FROM s_browse[1].b_gzda001,s_browse[1].b_gzda008
 
         BEFORE CONSTRUCT
               DISPLAY azzi085_filter_parser('gzda001') TO s_browse[1].b_gzda001
            DISPLAY azzi085_filter_parser('gzda008') TO s_browse[1].b_gzda008
      
         #add-point:filter段cs_ctrl name="filter.cs_ctrl"
         
         #end add-point
      
      END CONSTRUCT
 
      #add-point:filter段add_cs name="filter.add_cs"
      
      #end add-point
 
      BEFORE DIALOG
         #add-point:filter段b_dialog name="filter.b_dialog"
         
         #end add-point  
      
      ON ACTION accept
         ACCEPT DIALOG
 
      ON ACTION cancel
         LET INT_FLAG = 1
         EXIT DIALOG 
 
      #交談指令共用ACTION
      &include "common_action.4gl" 
         CONTINUE DIALOG
   
   END DIALOG
 
   IF NOT INT_FLAG THEN
      IF NOT g_wc.getIndexOf(g_wc_filter,1) THEN   #200828-00027檢查後進入,避免條件重複 (代追單)
         LET g_wc_filter = "   AND   ", g_wc_filter, "   "
         LET g_wc = g_wc , g_wc_filter   #200917-00047
      END IF   #200828-00027
   ELSE
      LET g_wc_filter = g_wc_filter_t
      LET g_wc = g_wc_t
   END IF
 
      CALL azzi085_filter_show('gzda001')
   CALL azzi085_filter_show('gzda008')
 
END FUNCTION
 
{</section>}
 
{<section id="azzi085.filter_parser" type="s" >}
#+ filter過濾功能
PRIVATE FUNCTION azzi085_filter_parser(ps_field) RETURNS (STRING)   #200213-00032
   #add-point:filter段define name="filter_parser.define_customerization"
   
   #end add-point    
   DEFINE ps_field   STRING
   DEFINE ls_tmp     STRING
   DEFINE li_tmp     LIKE type_t.num10
   DEFINE li_tmp2    LIKE type_t.num10
   DEFINE ls_var     STRING
   #add-point:filter段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="filter_parser.define"
   
   #end add-point    
   
   #一般條件解析
   LET ls_tmp = ps_field, "='"
   LET li_tmp = g_wc_filter.getIndexOf(ls_tmp,1)
   IF li_tmp > 0 THEN
      LET li_tmp = ls_tmp.getLength() + li_tmp
      LET li_tmp2 = g_wc_filter.getIndexOf("'",li_tmp + 1) - 1
      LET ls_var = g_wc_filter.subString(li_tmp,li_tmp2)
   END IF
 
   #模糊條件解析
   LET ls_tmp = ps_field, " like '"
   LET li_tmp = g_wc_filter.getIndexOf(ls_tmp,1)
   IF li_tmp > 0 THEN
      LET li_tmp = ls_tmp.getLength() + li_tmp
      LET li_tmp2 = g_wc_filter.getIndexOf("'",li_tmp + 1) - 1
      LET ls_var = g_wc_filter.subString(li_tmp,li_tmp2)
      LET ls_var = cl_replace_str(ls_var,'%','*')
   END IF
 
   RETURN ls_var
 
END FUNCTION
 
{</section>}
 
{<section id="azzi085.filter_show" type="s" >}
#+ 顯示過濾條件
PRIVATE FUNCTION azzi085_filter_show(ps_field) RETURNS ()   #200213-00032
   DEFINE ps_field         STRING
   DEFINE lnode_item       om.DomNode
   DEFINE ls_title         STRING
   DEFINE ls_name          STRING
   DEFINE ls_condition     STRING
 
   LET ls_name = "formonly.b_", ps_field
   LET lnode_item = gfrm_curr.findNode("TableColumn", ls_name)
   LET ls_title = lnode_item.getAttribute("text")
   IF ls_title.getIndexOf('※',1) > 0 THEN
      LET ls_title = ls_title.subString(1,ls_title.getIndexOf('※',1)-1)
   END IF
 
   #顯示資料組合
   LET ls_condition = azzi085_filter_parser(ps_field)
   IF NOT cl_null(ls_condition) THEN
      LET ls_title = ls_title, '※', ls_condition, '※'
   END IF
 
   #將資料顯示回去
   CALL lnode_item.setAttribute("text",ls_title)
 
END FUNCTION
 
{</section>}
 
{<section id="azzi085.query" type="s" >}
#+ 資料查詢QBE功能準備
PRIVATE FUNCTION azzi085_query() RETURNS ()   #200213-00032
   #add-point:query段define(客製用) name="query.define_customerization"
   
   #end add-point
   DEFINE ls_wc STRING
   #add-point:query段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="query.define"
   
   #end add-point
   
   #add-point:Function前置處理  name="query.pre_function"
   
   #end add-point
   
   LET INT_FLAG = 0
   LET ls_wc = g_wc
   
   #切換畫面
   IF g_main_hidden THEN
      CALL gfrm_curr.setElementHidden("mainlayout",0)
      CALL gfrm_curr.setElementHidden("worksheet",1)
      LET g_main_hidden = 0
   END IF
 
   CALL g_browser.clear() 
 
   #browser panel折疊
   IF g_worksheet_hidden THEN
      CALL gfrm_curr.setElementHidden("worksheet_vbox",0)
      CALL gfrm_curr.setElementImage("worksheethidden","worksheethidden-24.png")
      LET g_worksheet_hidden = 0
   END IF
   
   #單頭折疊
   IF g_header_hidden THEN
      CALL gfrm_curr.setElementHidden("vb_master",0)
      CALL gfrm_curr.setElementImage("controls","headerhidden-24")
      LET g_header_hidden = 0
   END IF
  
   INITIALIZE g_gzda_m.* TO NULL
   ERROR ""
 
   DISPLAY " " TO FORMONLY.b_count
   DISPLAY " " TO FORMONLY.h_count
   CALL azzi085_construct()
 
   IF INT_FLAG THEN
      #取消查詢
      LET INT_FLAG = 0
      #LET g_wc = ls_wc
      LET g_wc = " 1=2"
      CALL azzi085_browser_fill(g_wc,"F")
      CALL azzi085_fetch("")
      RETURN
   ELSE
      LET g_current_row = 1
      LET g_current_cnt = 0
   END IF
   
   #根據條件從新抓取資料
   LET g_error_show = 1
   CALL azzi085_browser_fill(g_wc,"F")   # 移到第一頁
   
   #儲存WC資訊
   CALL cl_dlg_save_user_latestqry("("||g_wc||")")
   
   #備份搜尋條件
   LET ls_wc = g_wc
   
   IF g_browser.getLength() = 0 THEN
      INITIALIZE g_errparam TO NULL 
      LET g_errparam.extend = "" 
      LET g_errparam.code = "-100" 
      LET g_errparam.popup = TRUE 
      CALL cl_err()
   ELSE
      CALL azzi085_fetch("F") 
   END IF
   
   LET g_wc_filter = ""
   
END FUNCTION
 
{</section>}
 
{<section id="azzi085.fetch" type="s" >}
#+ 指定PK後抓取單頭其他資料
PRIVATE FUNCTION azzi085_fetch(p_fl) RETURNS ()   #200213-00032
   #add-point:fetch段define(客製用) name="fetch.define_customerization"
   
   #end add-point
   DEFINE p_fl       LIKE type_t.chr1
   DEFINE ls_msg     STRING
   #add-point:fetch段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="fetch.define"
   
   #end add-point  
   
   #add-point:Function前置處理  name="fetch.pre_function"
   
   #end add-point
   
   #根據傳入的條件決定抓取的資料
   CASE p_fl
      WHEN "F" 
         LET g_current_idx = 1
      WHEN "P"
         IF g_current_idx > 1 THEN               
            LET g_current_idx = g_current_idx - 1
         END IF 
      WHEN "N"
         IF g_current_idx < g_header_cnt THEN
            LET g_current_idx =  g_current_idx + 1
         END IF        
      WHEN "L" 
         #LET g_current_idx = g_header_cnt        
         LET g_current_idx = g_browser.getLength()    
      WHEN "/"
         #詢問要指定的筆數
         IF (NOT g_no_ask) THEN      
            CALL cl_getmsg("fetch", g_lang) RETURNING ls_msg
            LET INT_FLAG = 0
 
            PROMPT ls_msg CLIPPED,": " FOR g_jump
               #交談指令共用ACTION
               &include "common_action.4gl"
            END PROMPT
            
            IF INT_FLAG THEN
               LET INT_FLAG = 0
               EXIT CASE  
            END IF           
         END IF
         IF g_jump > 0 THEN
            LET g_current_idx = g_jump
         END IF
         LET g_no_ask = FALSE     
   END CASE
 
   #筆數顯示
   LET g_browser_idx = g_current_idx 
   IF g_browser_cnt > 0 THEN
      DISPLAY g_browser_idx TO FORMONLY.b_index #當下筆數
      DISPLAY g_browser_cnt TO FORMONLY.b_count #總筆數
      DISPLAY g_browser_idx TO FORMONLY.h_index #當下筆數
      DISPLAY g_browser_cnt TO FORMONLY.h_count #總筆數
   ELSE
      DISPLAY '' TO FORMONLY.b_index #當下筆數
      DISPLAY '' TO FORMONLY.b_count #總筆數
      DISPLAY '' TO FORMONLY.h_index #當下筆數
      DISPLAY '' TO FORMONLY.h_count #總筆數
   END IF
   
   
   
   #避免超出browser資料筆數上限
   IF g_current_idx > g_browser.getLength() THEN
      LET g_browser_idx = g_browser.getLength()
      LET g_current_idx = g_browser.getLength() 
   END IF
   
   # 設定browse索引
   CALL g_curr_diag.setCurrentRow("s_browse", g_current_idx)
   CALL cl_navigator_setting(g_browser_idx, g_browser_cnt) 
 
   #代表沒有資料, 無需做後續資料撈取之動作
   IF g_current_idx = 0 THEN
      RETURN
   END IF
 
   #根據選定的筆數給予key欄位值
   LET g_gzda_m.gzda001 = g_browser[g_current_idx].b_gzda001
 
                       
   #讀取單頭所有欄位資料
   EXECUTE azzi085_master_referesh USING g_gzda_m.gzda001 INTO g_gzda_m.gzda009,g_gzda_m.gzda002,g_gzda_m.gzda008,g_gzda_m.gzda010,g_gzda_m.gzda005,g_gzda_m.gzda006,g_gzda_m.gzda007,g_gzda_m.gzda001,g_gzda_m.gzda003,g_gzda_m.gzda004,g_gzda_m.gzdastus,g_gzda_m.gzdaownid,g_gzda_m.gzdaowndp,g_gzda_m.gzdacrtid, 
       g_gzda_m.gzdacrtdp,g_gzda_m.gzdacrtdt,g_gzda_m.gzdamodid,g_gzda_m.gzdamoddt,g_gzda_m.gzdaownid_desc,g_gzda_m.gzdaowndp_desc,g_gzda_m.gzdacrtid_desc,g_gzda_m.gzdacrtdp_desc,g_gzda_m.gzdamodid_desc
   
   #遮罩相關處理
   LET g_gzda_m_mask_o.* =  g_gzda_m.*
   CALL azzi085_gzda_t_mask()
   LET g_gzda_m_mask_n.* =  g_gzda_m.*
   
   #根據資料狀態切換action狀態
   CALL cl_set_act_visible("statechange,modify,delete,reproduce", TRUE)
   CALL azzi085_set_act_visible()
   CALL azzi085_set_act_no_visible()
 
   #add-point:fetch段action控制 name="fetch.action_control"
   #170413-00060#2 add(s)
   #只有資料庫編號為dsrept時需執行[重產poijdbc.ini]
   IF g_gzda_m.gzda003 = 'dsrept' THEN
      CALL cl_set_act_visible("create_ini",TRUE)
   ELSE
      CALL cl_set_act_visible("create_ini",FALSE)
   END IF
   #170413-00060#2 add(e)
   #210120-00014#2 add(s)
   #只有資料庫編號為 dsdemo 時需執行[重產 multipoijdbc.ini]
   IF g_gzda_m.gzda003 = 'dsdemo' THEN
      CALL cl_set_act_visible("create_ini_mult",TRUE)
   ELSE
     CALL cl_set_act_visible("create_ini_mult",FALSE)
   END IF
   #210120-00014#2 add(e)   
   #end add-point  
   
   
   
   #保存單頭舊值
   LET g_gzda_m_t.* = g_gzda_m.*
   LET g_gzda_m_o.* = g_gzda_m.*
   
   LET g_data_owner = g_gzda_m.gzdaownid      
   LET g_data_dept  = g_gzda_m.gzdaowndp
   LET g_data_crtid = g_gzda_m.gzdacrtid   #180201-00054
   LET g_data_crtdp = g_gzda_m.gzdacrtdp   #180201-00054
   
   #重新顯示
   CALL azzi085_show()
   
 
END FUNCTION
 
{</section>}
 
{<section id="azzi085.insert" type="s" >}
#+ 資料新增
PRIVATE FUNCTION azzi085_insert() RETURNS ()   #200213-00032
   #add-point:insert段define(客製用) name="insert.define_customerization"
   
   #end add-point
   #add-point:insert段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="insert.define"
   DEFINE ls_tmp STRING   #221101-00001#1
   #end add-point    
   
   #add-point:Function前置處理  name="insert.pre_function"
   
   #end add-point
   
   CLEAR FORM #清畫面欄位內容
   INITIALIZE g_gzda_m.* TO NULL             #DEFAULT 設定
   LET g_gzda001_t = NULL
 
   
   #add-point:insert段before name="insert.before"
   
   #end add-point    
   
   CALL s_transaction_begin()
   
   WHILE TRUE
      
      #公用欄位給值
      #應用 a14 樣板自動產生(Version:5)    
      #公用欄位新增給值  
      LET g_gzda_m.gzdaownid = g_user
      LET g_gzda_m.gzdaowndp = g_dept
      LET g_gzda_m.gzdacrtid = g_user
      LET g_gzda_m.gzdacrtdp = g_dept 
      LET g_gzda_m.gzdacrtdt = cl_get_current()
      LET g_gzda_m.gzdamodid = g_user
      LET g_gzda_m.gzdamoddt = cl_get_current()
      LET g_gzda_m.gzdastus = 'Y'
 
 
 
 
      #append欄位給值
      
     
      #一般欄位給值
            LET g_gzda_m.gzda009 = "(default)"
      LET g_gzda_m.gzdastus = "Y"
 
 
      #add-point:單頭預設值 name="insert.default"
      LET ls_tmp = FGL_GETRESOURCE("dbi.default.driver")   #221101-00001#1
      IF ls_tmp.getLength() > 0 THEN
         LET g_gzda_m.gzda006 = ls_tmp.subString(4,6)
         LET g_gzda_m.gzda007 = ls_tmp.subString(7,ls_tmp.getLength())
      ELSE
         #新增 預設給oracle 
         LET g_gzda_m.gzda006 = "ora" #oracle
         LET g_gzda_m.gzda007 = "B1x" #11g R1
      END IF      
      CALL azzi085_set_combo_db_ver()        #180312-00040#1 add  重新刷新 combo box的值
      #CALL azzi085_set_gzda002_visible() #15/06/24 註解開放
      LET g_gzda_m.gzdastus = "N"
      #end add-point   
     
      #顯示狀態(stus)圖片
            #應用 a21 樣板自動產生(Version:3)
	  #根據當下狀態碼顯示圖片
      CASE g_gzda_m.gzdastus 
         WHEN "N"
            CALL gfrm_curr.setElementImage("statechange", "stus/32/inactive.png")
         WHEN "Y"
            CALL gfrm_curr.setElementImage("statechange", "stus/32/active.png")
         
      END CASE
 
 
 
     
      #資料輸入
      CALL azzi085_input("a")
      
      #add-point:單頭輸入後 name="insert.after_insert"
      
      #end add-point
      
      IF INT_FLAG THEN
         #取消
         LET INT_FLAG = 0
         DISPLAY g_current_cnt TO FORMONLY.h_count     #總筆數
         DISPLAY g_current_idx TO FORMONLY.h_index     #當下筆數
         INITIALIZE g_gzda_m.* TO NULL
         CALL azzi085_show()
      #  CALL s_transaction_end('N','0')   #(ver:51)
         INITIALIZE g_errparam TO NULL 
         LET g_errparam.extend = "" 
         LET g_errparam.code = 9001 
         LET g_errparam.popup = FALSE 
         CALL s_transaction_end('N','0')   #(ver:51)
         CALL cl_err()
         RETURN
      END IF
 
      LET g_rec_b = 0
      EXIT WHILE
   END WHILE
   
   #根據資料狀態切換action狀態
   CALL cl_set_act_visible("statechange,modify,delete,reproduce", TRUE)
   CALL azzi085_set_act_visible()
   CALL azzi085_set_act_no_visible()
 
   #將新增的資料併入搜尋條件中
   LET g_state = "insert"
   
   LET g_gzda001_t = g_gzda_m.gzda001
 
   
   #組合新增資料的條件
   LET g_add_browse = " ",
                      " gzda001 = '", g_gzda_m.gzda001, "' "
 
   #填到最後面
   LET g_current_idx = g_browser.getLength() + 1
   CALL azzi085_browser_fill("","")
   
   DISPLAY g_browser_cnt TO FORMONLY.h_count    #總筆數
   DISPLAY g_current_idx TO FORMONLY.h_index    #當下筆數
   CALL cl_navigator_setting(g_current_idx, g_browser_cnt)
 
   #撈取異動後的資料(主要是帶出reference)
   EXECUTE azzi085_master_referesh USING g_gzda_m.gzda001 INTO g_gzda_m.gzda009,g_gzda_m.gzda002,g_gzda_m.gzda008,g_gzda_m.gzda010,g_gzda_m.gzda005,g_gzda_m.gzda006,g_gzda_m.gzda007,g_gzda_m.gzda001,g_gzda_m.gzda003,g_gzda_m.gzda004,g_gzda_m.gzdastus,g_gzda_m.gzdaownid,g_gzda_m.gzdaowndp,g_gzda_m.gzdacrtid, 
       g_gzda_m.gzdacrtdp,g_gzda_m.gzdacrtdt,g_gzda_m.gzdamodid,g_gzda_m.gzdamoddt,g_gzda_m.gzdaownid_desc,g_gzda_m.gzdaowndp_desc,g_gzda_m.gzdacrtid_desc,g_gzda_m.gzdacrtdp_desc,g_gzda_m.gzdamodid_desc
   
   
   #遮罩相關處理
   LET g_gzda_m_mask_o.* =  g_gzda_m.*
   CALL azzi085_gzda_t_mask()
   LET g_gzda_m_mask_n.* =  g_gzda_m.*
   
   #將資料顯示到畫面上
   DISPLAY BY NAME g_gzda_m.gzda009,g_gzda_m.gzda002,g_gzda_m.gzda008,g_gzda_m.gzda010,g_gzda_m.gzda005,g_gzda_m.gzda006,g_gzda_m.gzda007,g_gzda_m.gzda001,g_gzda_m.gzda003,g_gzda_m.gzda004,g_gzda_m.gzdastus,g_gzda_m.gzdaownid,g_gzda_m.gzdaownid_desc,g_gzda_m.gzdaowndp,g_gzda_m.gzdaowndp_desc,g_gzda_m.gzdacrtid, 
       g_gzda_m.gzdacrtid_desc,g_gzda_m.gzdacrtdp,g_gzda_m.gzdacrtdp_desc,g_gzda_m.gzdacrtdt,g_gzda_m.gzdamodid,g_gzda_m.gzdamodid_desc,g_gzda_m.gzdamoddt
 
   #add-point:新增結束後 name="insert.after"
   
   #end add-point 
 
   LET g_data_owner = g_gzda_m.gzdaownid      
   LET g_data_dept  = g_gzda_m.gzdaowndp
   LET g_data_crtid = g_gzda_m.gzdacrtid   #180201-00054
   LET g_data_crtdp = g_gzda_m.gzdacrtdp   #180201-00054
 
   #功能已完成,通報訊息中心
   CALL azzi085_msgcentre_notify('insert')
   
END FUNCTION
 
{</section>}
 
{<section id="azzi085.modify" type="s" >}
#+ 資料修改
PRIVATE FUNCTION azzi085_modify() RETURNS ()   #200213-00032
   #add-point:modify段define(客製用) name="modify.define_customerization"
   
   #end add-point
   DEFINE lp_cnt       LIKE type_t.num5   #190806-00019
   #add-point:modify段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="modify.define"
   
   #end add-point
   
   #add-point:Function前置處理 name="modify.pre_function"
   
   #end add-point
   
   #先確定key值無遺漏
   IF g_gzda_m.gzda001 IS NULL
 
   THEN
      INITIALIZE g_errparam TO NULL 
      LET g_errparam.extend = "" 
      LET g_errparam.code = "std-00003" 
      LET g_errparam.popup = FALSE 
      CALL cl_err()
      RETURN
   END IF 
 
   ERROR ""
  
   #備份key值
   LET g_gzda001_t = g_gzda_m.gzda001
 
   
   CALL s_transaction_begin()
   
   #191101-00032 ---modify start---
   #190806-00019 ---start---
   #修改前須先確認資料是否存在
   IF g_action_choice = "modify" OR g_action_choice = "modify_detail" THEN   #191114-00036
      LET lp_cnt = 0
      EXECUTE azzi085_chk_data_exist_pre INTO lp_cnt USING g_gzda_m.gzda001
      IF cl_null(lp_cnt) OR lp_cnt = 0 THEN
         INITIALIZE g_errparam TO NULL
         LET g_errparam.extend = ""
         LET g_errparam.code = "-100"
         LET g_errparam.popup = TRUE
         CLOSE azzi085_cl
         CALL s_transaction_end('N','0')
         CALL cl_err()
         RETURN
      END IF
   END IF
   #190806-00019 --- end ---
   #191101-00032 --- modify end ---
 
   #先lock資料
   OPEN azzi085_cl USING g_gzda_m.gzda001
   IF SQLCA.SQLCODE THEN    #(ver:49)
      INITIALIZE g_errparam TO NULL 
      LET g_errparam.extend = "OPEN azzi085_cl:",SQLERRMESSAGE 
      LET g_errparam.code = SQLCA.SQLCODE
      LET g_errparam.popup = TRUE 
      CLOSE azzi085_cl
      CALL s_transaction_end('N','0')
      CALL cl_err()
      RETURN
   END IF
 
   #200804-00006 ---start---
   FETCH azzi085_cl
   IF SQLCA.SQLCODE THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.extend = "FETCH azzi085_cl:",SQLERRMESSAGE
      LET g_errparam.code = SQLCA.SQLCODE
      LET g_errparam.popup = TRUE
      CLOSE azzi085_cl
      #PostgreSQL需額外重新開啟, 無論主程式子程式都須關閉
      CALL s_transaction_end('N','0')
      CALL s_transaction_begin()
      CALL cl_err()
      RETURN
   END IF
   #200804-00006 --- end ---
 
   #顯示最新的資料
   EXECUTE azzi085_master_referesh USING g_gzda_m.gzda001 INTO g_gzda_m.gzda009,g_gzda_m.gzda002,g_gzda_m.gzda008,g_gzda_m.gzda010,g_gzda_m.gzda005,g_gzda_m.gzda006,g_gzda_m.gzda007,g_gzda_m.gzda001,g_gzda_m.gzda003,g_gzda_m.gzda004,g_gzda_m.gzdastus,g_gzda_m.gzdaownid,g_gzda_m.gzdaowndp,g_gzda_m.gzdacrtid, 
       g_gzda_m.gzdacrtdp,g_gzda_m.gzdacrtdt,g_gzda_m.gzdamodid,g_gzda_m.gzdamoddt,g_gzda_m.gzdaownid_desc,g_gzda_m.gzdaowndp_desc,g_gzda_m.gzdacrtid_desc,g_gzda_m.gzdacrtdp_desc,g_gzda_m.gzdamodid_desc
 
   #檢查是否允許此動作
   IF NOT azzi085_action_chk() THEN
      CALL s_transaction_end('N','0')
      RETURN
   END IF
 
   #遮罩相關處理
   LET g_gzda_m_mask_o.* =  g_gzda_m.*
   CALL azzi085_gzda_t_mask()
   LET g_gzda_m_mask_n.* =  g_gzda_m.*
   
   
 
   #顯示資料
   CALL azzi085_show()
   
   WHILE TRUE
      LET g_gzda_m.gzda001 = g_gzda001_t
 
      
      #寫入修改者/修改日期資訊
      LET g_gzda_m.gzdamodid = g_user 
LET g_gzda_m.gzdamoddt = cl_get_current()
LET g_gzda_m.gzdamodid_desc = cl_get_username(g_gzda_m.gzdamodid)
      
      #add-point:modify段修改前 name="modify.before_input"
      LET g_gzda_m_t.* = g_gzda_m.*
      #end add-point
 
      #資料輸入
      CALL azzi085_input("u")     
 
      #add-point:modify段修改後 name="modify.after_input"
      
      #end add-point
      
      IF INT_FLAG THEN
      #  CALL s_transaction_end('N','0')   #(ver:51)
         LET INT_FLAG = 0
         LET g_gzda_m.* = g_gzda_m_t.*
         CALL azzi085_show()
         INITIALIZE g_errparam TO NULL 
         LET g_errparam.extend = "" 
         LET g_errparam.code = 9001 
         LET g_errparam.popup = FALSE 
         CALL s_transaction_end('N','0')   #(ver:51)
         CALL cl_err()
         EXIT WHILE
      END IF
 
      #若有modid跟moddt則進行update
      UPDATE gzda_t SET (gzdastus,gzdamodid,gzdamoddt) = (g_gzda_m.gzdastus,g_gzda_m.gzdamodid,g_gzda_m.gzdamoddt)
       WHERE  gzda001 = g_gzda001_t
 
 
      EXIT WHILE
      
   END WHILE
 
   #根據資料狀態切換action狀態
   CALL cl_set_act_visible("statechange,modify,delete,reproduce", TRUE)
   CALL azzi085_set_act_visible()
   CALL azzi085_set_act_no_visible()
 
   #組合新增資料的條件
   LET g_add_browse = " ",
                      " gzda001 = '", g_gzda_m.gzda001, "' "
 
   #填到對應位置
   CALL azzi085_browser_fill(g_wc,"")
 
   CLOSE azzi085_cl
   CALL s_transaction_end('Y','0')
 
   #功能已完成,通報訊息中心
   CALL azzi085_msgcentre_notify('modify')
   
   LET g_worksheet_hidden = 0
   
END FUNCTION   
 
{</section>}
 
{<section id="azzi085.input" type="m" >}
#+ 資料輸入
PRIVATE FUNCTION azzi085_input(p_cmd) RETURNS ()   #200213-00032
   #add-point:input段define(客製用) name="input.define_customerization"

   #end add-point
   DEFINE p_cmd           LIKE type_t.chr1
   DEFINE l_ac_t          LIKE type_t.num10       #未取消的ARRAY CNT 
   DEFINE l_n             LIKE type_t.num10       #檢查重複用  
   DEFINE l_cnt           LIKE type_t.num10       #檢查重複用  
   DEFINE l_lock_sw       LIKE type_t.chr1        #單身鎖住否  
   DEFINE l_allow_insert  LIKE type_t.num5        #可新增否 
   DEFINE l_allow_delete  LIKE type_t.num5        #可刪除否  
   DEFINE l_count         LIKE type_t.num10
   DEFINE l_i             LIKE type_t.num10
   DEFINE l_insert        LIKE type_t.num10
   DEFINE ls_return       STRING
   DEFINE l_var_keys      DYNAMIC ARRAY OF STRING
   DEFINE l_var_keys_bak  DYNAMIC ARRAY OF STRING
   DEFINE l_field_keys    DYNAMIC ARRAY OF STRING
   DEFINE l_vars          DYNAMIC ARRAY OF STRING
   DEFINE l_fields        DYNAMIC ARRAY OF STRING
   DEFINE ls_mdl_str      STRING   #180411-00062
   DEFINE lp_js           STRING   #190412-00022
   #add-point:input段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="input.define"
   DEFINE ls_js_o         STRING   #舊值
   DEFINE ls_js_n         STRING   #新值
   DEFINE la_param        RECORD
                          gzda002  LIKE gzda_t.gzda002,
                          gzda003  LIKE gzda_t.gzda003,
                          gzda005  LIKE gzda_t.gzda005,
                          gzda006  LIKE gzda_t.gzda006,
                          gzda007  LIKE gzda_t.gzda007
                          END RECORD
   DEFINE ls_apsid        STRING
   DEFINE ls_apsport      STRING
   DEFINE ls_gzda004_orig LIKE gzda_t.gzda004
   DEFINE ls_cmd          STRING
   #180131-00085#--start--
   DEFINE ls_tmp          STRING 
   DEFINE ls_system_pw    STRING 
   DEFINE ls_sys_pw       STRING 
   DEFINE lch_pipe        base.Channel  
   #180131-00085#--end--   
   DEFINE l_accept_flag   LIKE type_t.chr1   #210114-00032#4 add
   DEFINE l_result        BOOLEAN            #210114-00032#4 add   
   #end add-point
   
   #add-point:Function前置處理  name="input.pre_function"

   #end add-point
   
   #切換至輸入畫面
   IF g_main_hidden THEN
      CALL gfrm_curr.setElementHidden("mainlayout",0)
      CALL gfrm_curr.setElementHidden("worksheet",1)
      LET g_main_hidden = 0
   END IF
   
   #將資料輸出到畫面上
   DISPLAY BY NAME g_gzda_m.gzda009,g_gzda_m.gzda002,g_gzda_m.gzda008,g_gzda_m.gzda010,g_gzda_m.gzda005,g_gzda_m.gzda006,g_gzda_m.gzda007,g_gzda_m.gzda001,g_gzda_m.gzda003,g_gzda_m.gzda004,g_gzda_m.gzdastus,g_gzda_m.gzdaownid,g_gzda_m.gzdaownid_desc,g_gzda_m.gzdaowndp,g_gzda_m.gzdaowndp_desc,g_gzda_m.gzdacrtid, 
       g_gzda_m.gzdacrtid_desc,g_gzda_m.gzdacrtdp,g_gzda_m.gzdacrtdp_desc,g_gzda_m.gzdacrtdt,g_gzda_m.gzdamodid,g_gzda_m.gzdamodid_desc,g_gzda_m.gzdamoddt
   
   CALL cl_set_head_visible("","YES")  
   
   #a-新增,r-複製,u-修改
   IF p_cmd = 'r' THEN
      #此段落的r動作等同於a
      LET p_cmd = 'a'
   END IF
 
   LET l_insert = FALSE
   LET g_action_choice = ""
 
   LET g_qryparam.state = "i"
   
   #控制key欄位可否輸入
   CALL azzi085_set_entry(p_cmd)
   #add-point:set_entry後 name="input.after_set_entry"

   #end add-point
   CALL azzi085_set_no_entry(p_cmd)
   
   #關閉被遮罩相關欄位輸入, 無法確定USER是否會需要輸入此欄位
   #因此先行關閉, 若有需要可於下方add-point中自行開啟
   CALL cl_mask_set_no_entry()
   #add-point:資料輸入前 name="input.before_input"

   #end add-point
   
   DIALOG ATTRIBUTES(UNBUFFERED,FIELD ORDER FORM)
   
      #單頭段
      INPUT BY NAME g_gzda_m.gzda009,g_gzda_m.gzda002,g_gzda_m.gzda008,g_gzda_m.gzda010,g_gzda_m.gzda005,g_gzda_m.gzda006,g_gzda_m.gzda007,g_gzda_m.gzda001,g_gzda_m.gzda003,g_gzda_m.gzda004,g_gzda_m.gzdastus 
         ATTRIBUTE(WITHOUT DEFAULTS)
         
         #自訂ACTION(master_input)
         
         
         BEFORE INPUT
            IF s_transaction_chk("N",0) THEN
               CALL s_transaction_begin()
            END IF
            #其他table資料備份(確定是否更改用)
            
            #add-point:input開始前 name="input.before.input"

            #end add-point
   
                  #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda009
            #add-point:BEFORE FIELD gzda009 name="input.b.gzda009"

            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda009
            
            #add-point:AFTER FIELD gzda009 name="input.a.gzda009"

            #END add-point
            
 
 
         #應用 a04 樣板自動產生(Version:3)
         ON CHANGE gzda009
            #add-point:ON CHANGE gzda009 name="input.g.gzda009"

            #END add-point 
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda002
            #add-point:BEFORE FIELD gzda002 name="input.b.gzda002"

            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda002
            
            #add-point:AFTER FIELD gzda002 name="input.a.gzda002"

            #END add-point
            
 
 
         #應用 a04 樣板自動產生(Version:3)
         ON CHANGE gzda002
            #add-point:ON CHANGE gzda002 name="input.g.gzda002"

            #END add-point 
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda008
            #add-point:BEFORE FIELD gzda008 name="input.b.gzda008"

            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda008
            
            #add-point:AFTER FIELD gzda008 name="input.a.gzda008"

            #END add-point
            
 
 
         #應用 a04 樣板自動產生(Version:3)
         ON CHANGE gzda008
            #add-point:ON CHANGE gzda008 name="input.g.gzda008"

            #END add-point 
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda010
            #add-point:BEFORE FIELD gzda010 name="input.b.gzda010"

            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda010
            
            #add-point:AFTER FIELD gzda010 name="input.a.gzda010"
            #161124-00001 start
            IF NOT cl_null(g_gzda_m.gzda010) THEN
               IF p_cmd = "a" OR (p_cmd = "u" AND g_gzda_m.gzda010 <> g_gzda_m_t.gzda010) THEN              

                  SELECT COUNT(1) INTO l_cnt FROM gzda_t  
                   WHERE gzda001 = g_gzda_m.gzda010 
                  IF l_cnt > 0 THEN 
                     INITIALIZE g_errparam TO NULL 
                     LET g_errparam.extend = g_gzda_m.gzda010
                     LET g_errparam.code   = "azz-00421" 
                     LET g_errparam.popup  = TRUE 
                     CALL cl_err()

                     NEXT FIELD CURRENT
                  END IF 
               END IF  
            END IF
            #161124-00001 end            
            #END add-point
            
 
 
         #應用 a04 樣板自動產生(Version:3)
         ON CHANGE gzda010
            #add-point:ON CHANGE gzda010 name="input.g.gzda010"

            #END add-point 
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda005
            #add-point:BEFORE FIELD gzda005 name="input.b.gzda005"

            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda005
            
            #add-point:AFTER FIELD gzda005 name="input.a.gzda005"
            IF NOT cl_null(g_gzda_m.gzda005) THEN
                IF p_cmd = "a" OR (p_cmd = "u" AND g_gzda_m.gzda005 <> g_gzda_m_t.gzda005) THEN
                  IF p_cmd = "u" AND g_gzda_m.gzda005 <> g_gzda_m_t.gzda005 THEN 
                     IF NOT azzi085_chk_db_setup("4") THEN 
                        NEXT FIELD gzda005
                     END IF
                  END IF
                  #161124-00001 #先註解 等整合組實作完成 
                  #IF g_gzda_m.gzda005 = "Y" THEN 
                  #   CALL cl_set_comp_entry("gzda010",TRUE) 
                  #END IF
                  #161124-00001                  
                END IF
            END IF
            #END add-point
            
 
 
         #應用 a04 樣板自動產生(Version:3)
         ON CHANGE gzda005
            #add-point:ON CHANGE gzda005 name="input.g.gzda005"

            #END add-point 
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda006
            #add-point:BEFORE FIELD gzda006 name="input.b.gzda006"

            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda006
            
            #add-point:AFTER FIELD gzda006 name="input.a.gzda006"
            CALL azzi085_set_combo_db_ver()
            #CALL azzi085_set_gzda002_visible() #15/06/24 註解開放
            #END add-point
            
 
 
         #應用 a04 樣板自動產生(Version:3)
         ON CHANGE gzda006
            #add-point:ON CHANGE gzda006 name="input.g.gzda006"

            #END add-point 
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda007
            #add-point:BEFORE FIELD gzda007 name="input.b.gzda007"

            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda007
            
            #add-point:AFTER FIELD gzda007 name="input.a.gzda007"
            
            #END add-point
            
 
 
         #應用 a04 樣板自動產生(Version:3)
         ON CHANGE gzda007
            #add-point:ON CHANGE gzda007 name="input.g.gzda007"

            #END add-point 
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda001
            #add-point:BEFORE FIELD gzda001 name="input.b.gzda001"

            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda001
            
            #add-point:AFTER FIELD gzda001 name="input.a.gzda001"
            #此段落由子樣板a05產生
            IF  NOT cl_null(g_gzda_m.gzda001) THEN 
               IF p_cmd = 'a' OR ( p_cmd = 'u' AND ( p_cmd = 'u' AND (g_gzda_m.gzda001 != g_gzda001_t ))) THEN 
#                  IF NOT ap_chk_notDup("","SELECT COUNT(*) FROM gzda_t WHERE "||"gzda001 = '"||g_gzda_m.gzda001 ||"'",'std-00004',0) THEN   #170413-00060#2 mark
                  IF NOT ap_chk_notDup("","SELECT COUNT(1) FROM gzda_t WHERE "||"gzda001 = '"||g_gzda_m.gzda001 ||"'",'std-00004',0) THEN    #170413-00060#2 add
                     NEXT FIELD CURRENT
                  END IF
                  IF NOT azzi085_chk_db_setup("2") THEN
                     NEXT FIELD CURRENT
                  END IF 
               END IF
            END IF
            #END add-point
            
 
 
         #應用 a04 樣板自動產生(Version:3)
         ON CHANGE gzda001
            #add-point:ON CHANGE gzda001 name="input.g.gzda001"

            #END add-point 
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda003
            #add-point:BEFORE FIELD gzda003 name="input.b.gzda003"

            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda003
            
            #add-point:AFTER FIELD gzda003 name="input.a.gzda003"

            #END add-point
            
 
 
         #應用 a04 樣板自動產生(Version:3)
         ON CHANGE gzda003
            #add-point:ON CHANGE gzda003 name="input.g.gzda003"

            #END add-point 
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzda004
            #add-point:BEFORE FIELD gzda004 name="input.b.gzda004"
            #190423-00046#1--start--
            INITIALIZE g_errparam TO NULL
            LET g_errparam.code = "azz-01040"
            LET g_errparam.extend = ""
            LET g_errparam.popup = TRUE
            CALL cl_err()
            #190423-00046#1--end--
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzda004
            
            #add-point:AFTER FIELD gzda004 name="input.a.gzda004"
            #密碼不可空白
            IF cl_null(g_gzda_m.gzda004) THEN
               NEXT FIELD gzda004
            END IF

            #密碼加密
            IF p_cmd = "a" OR (p_cmd = "u" AND g_gzda_m.gzda004 <> g_gzda_m_t.gzda004) THEN
               LET ls_gzda004_orig = g_gzda_m.gzda004   #保留未加密前的密碼
               LET g_gzda_m.gzda004 = cl_hashkey_base64_encode(g_gzda_m.gzda004)
            END IF
            #END add-point
            
 
 
         #應用 a04 樣板自動產生(Version:3)
         ON CHANGE gzda004
            #add-point:ON CHANGE gzda004 name="input.g.gzda004"
          
            #END add-point 
 
 
         #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD gzdastus
            #add-point:BEFORE FIELD gzdastus name="input.b.gzdastus"

            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:3)
         AFTER FIELD gzdastus
            
            #add-point:AFTER FIELD gzdastus name="input.a.gzdastus"
            IF NOT cl_null(g_gzda_m.gzdastus) THEN
               IF p_cmd = "u" AND g_gzda_m.gzdastus <> g_gzda_m_t.gzdastus THEN 
                  CASE g_gzda_m.gzdastus 
                     WHEN "N"
                     
                        IF NOT azzi085_chk_db_setup("1") THEN 
                           NEXT FIELD gzdastus 
                        END IF 
                     
                        CALL gfrm_curr.setElementImage("statechange", "stus/32/inactive.png")
                     WHEN "Y"
                        CALL gfrm_curr.setElementImage("statechange", "stus/32/active.png")
                  END CASE
               END IF 
            END IF 
            #END add-point
            
 
 
         #應用 a04 樣板自動產生(Version:3)
         ON CHANGE gzdastus
            #add-point:ON CHANGE gzdastus name="input.g.gzdastus"

            #END add-point 
 
 
 #欄位檢查
                  #Ctrlp:input.c.gzda009
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda009
            #add-point:ON ACTION controlp INFIELD gzda009 name="input.c.gzda009"

            #END add-point
 
 
         #Ctrlp:input.c.gzda002
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda002
            #add-point:ON ACTION controlp INFIELD gzda002 name="input.c.gzda002"

            #END add-point
 
 
         #Ctrlp:input.c.gzda008
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda008
            #add-point:ON ACTION controlp INFIELD gzda008 name="input.c.gzda008"

            #END add-point
 
 
         #Ctrlp:input.c.gzda010
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda010
            #add-point:ON ACTION controlp INFIELD gzda010 name="input.c.gzda010"

            #END add-point
 
 
         #Ctrlp:input.c.gzda005
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda005
            #add-point:ON ACTION controlp INFIELD gzda005 name="input.c.gzda005"

            #END add-point
 
 
         #Ctrlp:input.c.gzda006
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda006
            #add-point:ON ACTION controlp INFIELD gzda006 name="input.c.gzda006"

            #END add-point
 
 
         #Ctrlp:input.c.gzda007
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda007
            #add-point:ON ACTION controlp INFIELD gzda007 name="input.c.gzda007"

            #END add-point
 
 
         #Ctrlp:input.c.gzda001
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda001
            #add-point:ON ACTION controlp INFIELD gzda001 name="input.c.gzda001"

            #END add-point
 
 
         #Ctrlp:input.c.gzda003
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda003
            #add-point:ON ACTION controlp INFIELD gzda003 name="input.c.gzda003"

            #END add-point
 
 
         #Ctrlp:input.c.gzda004
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzda004
            #add-point:ON ACTION controlp INFIELD gzda004 name="input.c.gzda004"

            #END add-point
 
 
         #Ctrlp:input.c.gzdastus
         #應用 a03 樣板自動產生(Version:3)
         ON ACTION controlp INFIELD gzdastus
            #add-point:ON ACTION controlp INFIELD gzdastus name="input.c.gzdastus"

            #END add-point
 
 
 #欄位開窗
 
         AFTER INPUT
            #若點選cancel則離開dialog
            IF INT_FLAG THEN
               EXIT DIALOG
            END IF
            
            #錯誤訊息統整顯示
            #CALL cl_err_collect_show()
            #CALL cl_showmsg()
  
            IF p_cmd <> "u" THEN
               #當p_cmd不為u代表為新增/複製
               LET l_count = 1  
 
               #確定新增的資料不存在(不重複)
               SELECT COUNT(1) INTO l_count FROM gzda_t
                WHERE  gzda001 = g_gzda_m.gzda001
 
               IF l_count = 0 THEN
               
                  #add-point:單頭新增前 name="input.head.b_insert"
                  LET g_gzda_m.gzdastus = "Y"
                  CALL gfrm_curr.setElementImage("statechange", "stus/32/active.png")
                  #180131-00085#1--start--
                  #資料庫類型:N 建立 
                  #1. 新建一個 "完全獨立無法跟其他 db 溝通的 dbuser" 給抓取資料人使用
                  #2. 把要開放給人家看的資料，使用 book_export 工具匯出，並匯入該新的 dbuser 下
                  #   用這樣的做法期望能滿足管控。那麼缺口在 1. 處，creadbuser 建出來的都是能 grant rights 的
                  #   增加一個 creatdbuser 的功能，建置相關 dbuser 給 azzi085 的 N其他用途 使用
                  #參數位置 createdbuser dbusername     dbtype     dbpw        systempw       syspw
                  #                     登入連線user  資料庫類型   資料庫密碼    system密碼     sys密碼

                  #221028-00002#1 只有ORACLE有 sys/system其他沒有
                  IF g_gzda_m.gzda006 IS NOT NULL AND g_gzda_m.gzda006 = "ora" THEN
                    #LET gs_id = "createdbuser",FGL_GETPID()
                     LET gs_id = FGL_GETPID() USING "<<<<<<<","_createdbuser"

                     IF cl_null(g_gzda_m.gzda005) THEN 
                        INITIALIZE g_errparam TO NULL 
                        LET g_errparam.extend = g_gzda_m.gzda001
                        LET g_errparam.code   = "azz-00370" #資料庫類型不符合
                        LET g_errparam.popup  = TRUE 
                        CALL cl_err()
                        NEXT FIELD CURRENT 
                     END IF 

                     #輸入(ORACLE) db需要的sys/system密碼 (連動cratedbuser用)
                     CALL azzi085_input_dba_password() RETURNING ls_system_pw ,ls_sys_pw
                     IF INT_FLAG THEN
                        LET INT_FLAG = FALSE
                        NEXT FIELD CURRENT 
                     END IF

                    #LET ls_cmd = "createdbuser ", g_gzda_m.gzda001 ," ",g_gzda_m.gzda005  #190502-00006 #1 mark
                     LET ls_cmd = "createdbuser ", g_gzda_m.gzda003 ," ",g_gzda_m.gzda005  #190502-00006 #1 add
                     LET ls_cmd = ls_cmd ," ",ls_gzda004_orig," ",ls_system_pw," ",ls_sys_pw  ," ",gs_id
                     LET lch_pipe = base.Channel.create()
                     CALL lch_pipe.openPipe(ls_cmd.trim(), "r")
                     LET l_cnt = TRUE
                     WHILE lch_pipe.read(ls_return)
                        IF ls_return.trim() IS NULL THEN CONTINUE WHILE END IF
                        LET ls_return = DOWNSHIFT(ls_return)
                        IF ls_return.getIndexOf("error",1) THEN
                           IF ls_return.getIndexOf("system",ls_return.getIndexOf("error",1)+6) THEN
                              LET ls_cmd = "system"
                              LET l_cnt = FALSE
                              EXIT WHILE
                           END IF
                           IF ls_return.getIndexOf("sys",ls_return.getIndexOf("error",1)+6) THEN
                              LET ls_cmd = "sys"
                              LET l_cnt = FALSE
                              EXIT WHILE 
                           END IF
                        END IF
                     END WHILE
                     CALL lch_pipe.close()
                     IF NOT l_cnt THEN
                        INITIALIZE g_errparam TO NULL
                        LET g_errparam.extend = ""
                        LET g_errparam.code   = "azz-00469"
                        LET g_errparam.popup  = TRUE
                        LET g_errparam.replace[1] = ls_cmd
                        CALL cl_err()
                        NEXT FIELD CURRENT
                     END IF

                     #確認底層db user 建立成功且正常結束
                     CALL azzi085_monitor_create_dbuser()
                  END IF

                  #180131-00085#1--end--
                  IF NOT cl_db_test_connect(g_gzda_m.gzda001,g_gzda_m.gzda003,g_gzda_m.gzda004,g_gzda_m.gzda006,g_gzda_m.gzda007) THEN
                     INITIALIZE g_errparam TO NULL 
                     LET g_errparam.extend = g_gzda_m.gzda001
                     LET g_errparam.code   = "azz-00370" 
                     LET g_errparam.popup  = TRUE 
                     CALL cl_err()
                     NEXT FIELD CURRENT
                  END IF
                  #180131-00085#--start--
                  #Y:屬於ERP 類型 才需呼叫 ms_crt_db_adz_init 建立synonym 到 ds #240709-00026#1
                  IF g_gzda_m.gzda005 = "Y" THEN 
                     #170628-00013 start
                     IF azzi085_create_new_db_adz_init(g_gzda_m.gzda003,ls_gzda004_orig) THEN 
                        ELSE 
                          INITIALIZE g_errparam TO NULL 
                          LET g_errparam.extend = ""
                          LET g_errparam.code   = "azz-00446" 
                          LET g_errparam.popup  = TRUE 
                          LET g_errparam.replace[1] = ms_crt_db_adz_init  #240709-00026#1
                          LET g_errparam.replace[2] = g_gzda_m.gzda001
                          CALL cl_err()
                          NEXT FIELD CURRENT
                     END IF
                  END IF   
                  #170628-00013 end
                  #180131-00085#--end--
                  #end add-point
               
                  #190412-00022 ---start---
                  #資料所有者與資料所有部門改為呼叫FUNCTION取得
                  LET lp_js = util.JSON.stringify(g_gzda_m)
                  CALL cl_auth_belong_data(lp_js,"gzda_t") RETURNING g_gzda_m.gzdaownid,g_gzda_m.gzdaowndp
                  #190412-00022 --- end ---
 
                  #將新增的單頭資料寫入資料庫
                  INSERT INTO gzda_t (gzda009,gzda002,gzda008,gzda010,gzda005,gzda006,gzda007,gzda001,gzda003,gzda004,gzdastus,gzdaownid,gzdaowndp,gzdacrtid,gzdacrtdp,gzdacrtdt,gzdamodid,gzdamoddt)
                  VALUES (g_gzda_m.gzda009,g_gzda_m.gzda002,g_gzda_m.gzda008,g_gzda_m.gzda010,g_gzda_m.gzda005,g_gzda_m.gzda006,g_gzda_m.gzda007,g_gzda_m.gzda001,g_gzda_m.gzda003,g_gzda_m.gzda004,g_gzda_m.gzdastus,g_gzda_m.gzdaownid,g_gzda_m.gzdaowndp,g_gzda_m.gzdacrtid,g_gzda_m.gzdacrtdp,g_gzda_m.gzdacrtdt, 
                      g_gzda_m.gzdamodid,g_gzda_m.gzdamoddt) 
                  
                  #add-point:單頭新增中 name="input.head.m_insert"

                  #end add-point
                  
                  #若寫入錯誤則提示錯誤訊息並返回輸入頁面
                  IF SQLCA.SQLCODE THEN
                     INITIALIZE g_errparam TO NULL 
                     LET g_errparam.extend = "gzda_t:",SQLERRMESSAGE 
                     LET g_errparam.code = SQLCA.SQLCODE
                     LET g_errparam.popup = TRUE 
                     CALL s_transaction_end('N','0')   #161130-00056
                     CALL cl_err()
                     NEXT FIELD CURRENT
                  END IF
                  
                  
                  
                  #資料多語言用-增/改
                  
                     #161130-00056
                  
                  #add-point:單頭新增後 name="input.head.a_insert"
                  #將密碼寫到相關檔案
                  #IF g_gzda_m.gzda004 <> g_gzda_m_t.gzda004 THEN    #250701-00007#1 mark
                     #g_gzda_m.gzda001 = ds 有修改密碼要把修改的密碼寫到$FBIN/4gl/gendbs.inc
                     CASE
                        WHEN g_gzda_m.gzda001 = "ds"
                           IF p_cmd = 'u' AND (g_gzda_m.gzda004 != g_gzda_m_t.gzda004 ) THEN
                              CALL azzi085_change_password()
                           END IF

                        WHEN g_gzda_m.gzda001 = "dsaps"
                           IF os.Path.chdir(os.Path.join(FGL_GETENV("APSE"),"std/engine")) THEN ##171127-00025#1
                              #"java -jar console.jar  ",FGL_GETENV("ORACLE_SID")," ",[DB User]," ", [User Pwd]," ORACLE set_pw"
                              LET ls_cmd = "java -jar console.jar ",1 SPACES,FGL_GETENV("ORACLE_SID"),1 SPACES,g_gzda_m.gzda003 CLIPPED,1 SPACES, ls_gzda004_orig,1 SPACES ," ORACLE set_pw "
#                              LET ls_apsid = g_gzda_m.gzda002 CLIPPED
#                              LET ls_apsport = "1521"
#                              LET ls_cmd = "java -Xms1024M -Xmx3096M -jar console.jar",
#                                           1 SPACES,ls_apsid.trim(),1 SPACES,ls_apsport.trim(),
#                                           1 SPACES,FGL_GETENV("ZONE"),1 SPACES,g_gzda_m.gzda003 CLIPPED,
#                                           1 SPACES,ls_gzda004_orig,1 SPACES," ORACLE set_pw"
                              RUN ls_cmd
                           END IF
                     END CASE
                  #END IF                 #250701-00007#1 mark
                  #180131-00085#1--start--
                  IF NOT SQLCA.SQLCODE THEN
                     INITIALIZE g_errparam TO NULL 
                     LET g_errparam.code = "apm-00538"
                     LET g_errparam.popup = TRUE 
                     CALL cl_err()
                  END IF
                  #180131-00085#1--end--     
                  CALL azzi085_noticed_event("insert",g_gzda_m.gzda001)     #180312-00040#1 add  azzi085_noticed_event                             
                  #end add-point
                  
                  CALL s_transaction_end('Y','0')
               ELSE
               #  CALL s_transaction_end('N','0')   #(ver:51)
                  INITIALIZE g_errparam TO NULL 
                  LET g_errparam.extend = g_gzda_m.gzda001
                  LET g_errparam.code = "std-00006" 
                  LET g_errparam.popup = TRUE 
                  CALL s_transaction_end('N','0')   #(ver:51)
                  CALL cl_err()
               END IF 
            ELSE
               #add-point:單頭修改前 name="input.head.b_update"
               IF NOT cl_db_test_connect(g_gzda_m.gzda001,g_gzda_m.gzda003,g_gzda_m.gzda004,g_gzda_m.gzda006,g_gzda_m.gzda007) THEN
                  INITIALIZE g_errparam TO NULL 
                  LET g_errparam.extend = g_gzda_m.gzda001
                  LET g_errparam.code   = "azz-00370" 
                  LET g_errparam.popup  = TRUE 
                  CALL cl_err()
                  LET g_ins_logd = TRUE    #180312-00040#1 add
                  NEXT FIELD CURRENT
               END IF 
               #end add-point
               
               #將遮罩欄位還原
               CALL azzi085_gzda_t_mask_restore('restore_mask_o')
               
               UPDATE gzda_t SET (gzda009,gzda002,gzda008,gzda010,gzda005,gzda006,gzda007,gzda001,gzda003,gzda004,gzdastus,gzdaownid,gzdaowndp,gzdacrtid,gzdacrtdp,gzdacrtdt,gzdamodid,gzdamoddt) = (g_gzda_m.gzda009,g_gzda_m.gzda002,g_gzda_m.gzda008,g_gzda_m.gzda010,g_gzda_m.gzda005,g_gzda_m.gzda006,g_gzda_m.gzda007, 
                   g_gzda_m.gzda001,g_gzda_m.gzda003,g_gzda_m.gzda004,g_gzda_m.gzdastus,g_gzda_m.gzdaownid,g_gzda_m.gzdaowndp,g_gzda_m.gzdacrtid,g_gzda_m.gzdacrtdp,g_gzda_m.gzdacrtdt,g_gzda_m.gzdamodid,g_gzda_m.gzdamoddt)
                WHERE  gzda001 = g_gzda001_t #
 
               #add-point:單頭修改中 name="input.head.m_update"

               #end add-point
               CASE
                  WHEN SQLCA.sqlerrd[3] = 0  #更新不到的處理
                  #  CALL s_transaction_end('N','0')   #(ver:51)
                     INITIALIZE g_errparam TO NULL 
                     LET g_errparam.extend = "gzda_t" 
                     LET g_errparam.code = "std-00009" 
                     LET g_errparam.popup = TRUE 
                     CALL s_transaction_end('N','0')   #(ver:51)
                     CALL cl_err()
                     NEXT FIELD CURRENT
 
                  WHEN SQLCA.SQLCODE #其他錯誤
                  #  CALL s_transaction_end('N','0')   #(ver:51)
                     INITIALIZE g_errparam TO NULL 
                     LET g_errparam.extend = "gzda_t:",SQLERRMESSAGE 
                     LET g_errparam.code = SQLCA.SQLCODE
                     LET g_errparam.popup = TRUE 
                     CALL s_transaction_end('N','0')   #(ver:51)
                     CALL cl_err()
                     NEXT FIELD CURRENT
 
                  OTHERWISE
                     
                     #資料多語言用-增/改
                     
                        #161130-00056
                     
                     #將遮罩欄位進行遮蔽
                     CALL azzi085_gzda_t_mask_restore('restore_mask_n')
                     
                     #add-point:單頭修改後 name="input.head.a_update"
                     # 紀錄重要資料修改歷程
                     FOR l_i = 1 TO 2
                        INITIALIZE la_param.* TO NULL
                        # 利用JSON的方式組合字串
                        IF l_i = 1 THEN
                           LET la_param.gzda002 = g_gzda_m_t.gzda002
                           LET la_param.gzda003 = g_gzda_m_t.gzda003
                           LET la_param.gzda005 = g_gzda_m_t.gzda005
                           LET la_param.gzda006 = g_gzda_m_t.gzda006
                           LET la_param.gzda007 = g_gzda_m_t.gzda007
                           LET ls_js_o = util.JSON.stringify(la_param)   # 舊值
                        ELSE
                           LET la_param.gzda002 = g_gzda_m.gzda002
                           LET la_param.gzda003 = g_gzda_m.gzda003
                           LET la_param.gzda005 = g_gzda_m.gzda005
                           LET la_param.gzda006 = g_gzda_m.gzda006
                           LET la_param.gzda007 = g_gzda_m.gzda007
                           LET ls_js_n = util.JSON.stringify(la_param)   # 新值
                        END IF
                     END FOR
                     CALL cl_log_modified_important(ls_js_o, ls_js_n)
                     
                     #將密碼寫到相關檔案
                     IF g_gzda_m.gzda004 <> g_gzda_m_t.gzda004 THEN
                        #g_gzda_m.gzda001 = ds 有修改密碼要把修改的密碼寫到$FBIN/4gl/gendbs.inc
                        CASE
                           WHEN g_gzda_m.gzda001 = "ds"
                              IF p_cmd = 'u' AND (g_gzda_m.gzda004 != g_gzda_m_t.gzda004 ) THEN
                                 CALL azzi085_change_password()
                              END IF

                           WHEN g_gzda_m.gzda001 = "dsaps"
                              IF os.Path.chdir(os.Path.join(FGL_GETENV("APSE"),"std/engine")) THEN ##171127-00025#1
                              #"java -jar console.jar  ",FGL_GETENV("ORACLE_SID")," ",[DB User]," ", [User Pwd]," ORACLE set_pw"
                                 LET ls_cmd = "java -jar console.jar ",1 SPACES,FGL_GETENV("ORACLE_SID"),1 SPACES,g_gzda_m.gzda003 CLIPPED,1 SPACES, ls_gzda004_orig,1 SPACES ," ORACLE set_pw "
#                                LET ls_apsid = g_gzda_m.gzda002 CLIPPED
#                                LET ls_apsport = "1521"
#                                LET ls_cmd = "java -Xms1024M -Xmx3096M -jar console.jar",
#                                           1 SPACES,ls_apsid.trim(),1 SPACES,ls_apsport.trim(),
#                                           1 SPACES,FGL_GETENV("ZONE"),1 SPACES,g_gzda_m.gzda003 CLIPPED,
#                                           1 SPACES,ls_gzda004_orig,1 SPACES," ORACLE set_pw"
                              RUN ls_cmd
                           END IF
                        END CASE
                     END IF
                     CALL azzi085_noticed_event("modify",g_gzda001_t)   #180312-00040#1 add azzi085_noticed_event
                     #end add-point
                     #修改歷程記錄(單頭修改)
                     LET g_log1 = util.JSON.stringify(g_gzda_m_t)
                     LET g_log2 = util.JSON.stringify(g_gzda_m)
                     IF NOT cl_log_modified_record(g_log1,g_log2) THEN 
                        CALL s_transaction_end('N','0')
                     ELSE
                        CALL s_transaction_end('Y','0')
                     END IF
               END CASE
               
            END IF
           #controlp
      END INPUT
      
      #add-point:input段more input  name="input.more_input"

      #end add-point
    
      BEFORE DIALOG
         #CALL cl_err_collect_init()
         #add-point:input段before_dialog  name="input.before_dialog"

         #end add-point
          
      ON ACTION controlf
         CALL cl_set_focus_form(ui.Interface.getRootNode()) RETURNING g_fld_name,g_frm_name
         CALL cl_fldhelp(g_frm_name, g_fld_name, g_lang)
 
      ON ACTION controlr
         CALL cl_show_req_fields()
 
      ON ACTION controls
         IF g_header_hidden THEN
            CALL gfrm_curr.setElementHidden("vb_master",0)
            CALL gfrm_curr.setElementImage("controls","small/arr-u.png")
            LET g_header_hidden = 0     #visible
         ELSE
            CALL gfrm_curr.setElementHidden("vb_master",1)
            CALL gfrm_curr.setElementImage("controls","small/arr-d.png")
            LET g_header_hidden = 1     #hidden     
         END IF
 
      ON ACTION accept
       #210120-00014#2 mark (s)
         ##210114-00032#4 add(s)
         #IF g_gzda_m.gzda003 != "ds" AND g_gzda_m.gzda005 = "Y" AND g_gzda_m.gzdastus = "Y" THEN
         #   LET l_accept_flag = "Y"
         #END IF
         ##210114-00032#4 add(e)   
       #210120-00014#2 mark (e)         
         ACCEPT DIALOG
         
      #放棄輸入
      ON ACTION cancel
         LET g_action_choice=""
         LET INT_FLAG = TRUE 
         EXIT DIALOG
 
      #在dialog 右上角 (X)
      ON ACTION close 
         LET INT_FLAG = TRUE 
         EXIT DIALOG
    
      #toolbar 離開
      ON ACTION exit
         LET INT_FLAG = TRUE 
         EXIT DIALOG
   
      #交談指令共用ACTION
      &include "common_action.4gl" 
         CONTINUE DIALOG 
   END DIALOG
    
   #add-point:input段after input  name="input.after_input"
   #170413-00060#2 add(s)
   #dsrept密碼變更時，自動重新產生poijdbc.ini
   IF g_gzda_m.gzda003 = 'dsrept' AND g_gzda_m.gzda004 <> g_gzda_m_t.gzda004 THEN
      IF cl_spreadsheet_create_ini() THEN
         INITIALIZE g_errparam TO NULL
         LET g_errparam.code = 'azz-01179'
         LET g_errparam.extend = cl_spreadsheet_ini_path()
         LET g_errparam.popup = TRUE
         CALL cl_err()         
      END IF
   END IF   
   #170413-00060#2 add(e)  

   ##221206-00021#1 mark (s)
   #210114-00032#1 add(s)
   #dsdemo 密碼變更時，自動重新產生 multipoijdbc.ini
   #IF g_gzda_m.gzda003 != "ds" AND g_gzda_m.gzda005 = "Y" AND g_gzda_m.gzdastus = "Y" THEN  
   #   LET l_accept_flag = "Y"
   #END IF 
   #IF l_accept_flag = "Y" THEN
   #   DISPLAY "CREATE multipoijdbc.ini"
   #   LET l_result = azzi085_create_multipoijdbc()
   #END IF
   #210114-00032#1 add(e)
   ##221206-00021#1 mark (e)

   #end add-point    
 
END FUNCTION
 
{</section>}
 
{<section id="azzi085.reproduce" type="s" >}
#+ 資料複製
PRIVATE FUNCTION azzi085_reproduce() RETURNS ()   #200213-00032
   #add-point:reproduce段define(客製用) name="reproduce.define_customerization"
   
   #end add-point
   DEFINE l_newno     LIKE gzda_t.gzda001 
   DEFINE l_oldno     LIKE gzda_t.gzda001 
 
   DEFINE l_master    RECORD LIKE gzda_t.* #此變數樣板目前無使用
   DEFINE l_cnt       LIKE type_t.num10
   #add-point:reproduce段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="reproduce.define"
   
   #end add-point   
   
   #add-point:Function前置處理  name="reproduce.pre_function"
   
   #end add-point
   
   #切換畫面
   IF g_main_hidden THEN
      CALL gfrm_curr.setElementHidden("mainlayout",0)
      CALL gfrm_curr.setElementHidden("worksheet",1)
      LET g_main_hidden = 0
   END IF
   
   #先確定key值無遺漏
   IF g_gzda_m.gzda001 IS NULL
 
   THEN
      INITIALIZE g_errparam TO NULL 
      LET g_errparam.extend = "" 
      LET g_errparam.code = "std-00003" 
      LET g_errparam.popup = FALSE 
      CALL cl_err()
      RETURN
   END IF
   
   #備份key值
   LET g_gzda001_t = g_gzda_m.gzda001
 
   
   #清空key值
   LET g_gzda_m.gzda001 = ""
 
    
   CALL azzi085_set_entry("a")
   CALL azzi085_set_no_entry("a")
   
   #公用欄位給予預設值
   #應用 a14 樣板自動產生(Version:5)    
      #公用欄位新增給值  
      LET g_gzda_m.gzdaownid = g_user
      LET g_gzda_m.gzdaowndp = g_dept
      LET g_gzda_m.gzdacrtid = g_user
      LET g_gzda_m.gzdacrtdp = g_dept 
      LET g_gzda_m.gzdacrtdt = cl_get_current()
      LET g_gzda_m.gzdamodid = g_user
      LET g_gzda_m.gzdamoddt = cl_get_current()
      LET g_gzda_m.gzdastus = 'Y'
 
 
 
   
   CALL s_transaction_begin()
   
   #add-point:複製輸入前 name="reproduce.head.b_input"
   
   #end add-point
   
   #顯示狀態(stus)圖片
         #應用 a21 樣板自動產生(Version:3)
	  #根據當下狀態碼顯示圖片
      CASE g_gzda_m.gzdastus 
         WHEN "N"
            CALL gfrm_curr.setElementImage("statechange", "stus/32/inactive.png")
         WHEN "Y"
            CALL gfrm_curr.setElementImage("statechange", "stus/32/active.png")
         
      END CASE
 
 
 
   
   #清空key欄位的desc
   
   
   #資料輸入
   CALL azzi085_input("r")
   
   IF INT_FLAG THEN
      #取消
      INITIALIZE g_gzda_m.* TO NULL
      CALL azzi085_show()
   #  CALL s_transaction_end('N','0')   #(ver:51)
      LET INT_FLAG = 0
      INITIALIZE g_errparam TO NULL 
      LET g_errparam.extend = "" 
      LET g_errparam.code = 9001 
      LET g_errparam.popup = FALSE 
      CALL s_transaction_end('N','0')   #(ver:51)
      CALL cl_err()
      RETURN
   END IF
   
   CALL s_transaction_begin()
   
   #add-point:單頭複製前 name="reproduce.head.b_insert"
   
   #end add-point
   
   #add-point:單頭複製中 name="reproduce.head.m_insert"
   
   #end add-point
   
   IF SQLCA.SQLCODE THEN
      INITIALIZE g_errparam TO NULL 
      LET g_errparam.extend = "gzda_t:",SQLERRMESSAGE 
      LET g_errparam.code = SQLCA.SQLCODE
      LET g_errparam.popup = TRUE 
      CALL s_transaction_end('N','0')
      CALL cl_err()
      RETURN
   END IF
   
   #add-point:單頭複製後 name="reproduce.head.a_insert"
   
   #end add-point
   
   CALL s_transaction_end('Y','0')
   
   #根據資料狀態切換action狀態
   CALL cl_set_act_visible("statechange,modify,delete,reproduce", TRUE)
   CALL azzi085_set_act_visible()
   CALL azzi085_set_act_no_visible()
 
   #將新增的資料併入搜尋條件中
   LET g_state = "insert"
   
   LET g_gzda001_t = g_gzda_m.gzda001
 
   
   #組合新增資料的條件
   LET g_add_browse = " ",
                      " gzda001 = '", g_gzda_m.gzda001, "' "
 
   #填到最後面
   LET g_current_idx = g_browser.getLength() + 1
   CALL azzi085_browser_fill("","")
   
   DISPLAY g_browser_cnt TO FORMONLY.h_count    #總筆數
   DISPLAY g_current_idx TO FORMONLY.h_index    #當下筆數
   CALL cl_navigator_setting(g_current_idx, g_browser_cnt)
   
   #add-point:完成複製段落後 name="reproduce.after_reproduce"
   
   #end add-point
              
   LET g_data_owner = g_gzda_m.gzdaownid      
   LET g_data_dept  = g_gzda_m.gzdaowndp
   LET g_data_crtid = g_gzda_m.gzdacrtid   #180201-00054
   LET g_data_crtdp = g_gzda_m.gzdacrtdp   #180201-00054
              
   #功能已完成,通報訊息中心
   CALL azzi085_msgcentre_notify('reproduce')
                 
END FUNCTION
 
{</section>}
 
{<section id="azzi085.show" type="s" >}
#+ 資料顯示 
PRIVATE FUNCTION azzi085_show() RETURNS ()   #200213-00032
   #add-point:show段define(客製用) name="show.define_customerization"
   
   #end add-point
   #add-point:show段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="show.define"
   
   #end add-point  
   
   #add-point:show段Function前置處理  name="show.before"
   
   #end add-point
   
   
   
   #帶出公用欄位reference值
   #應用 a12 樣板自動產生(Version:4)
 
 
 
    
   #顯示followup圖示
   #應用 a48 樣板自動產生(Version:3)
   CALL azzi085_set_pk_array()
   #add-point:ON ACTION agendum name="show.follow_pic"
   
   #END add-point
   CALL cl_user_overview_set_follow_pic()
  
 
 
 
   
   #讀入ref值(單頭)
   #add-point:show段reference name="show.head.reference"
   CALL  azzi085_set_combo_db_ver()
   #CALL azzi085_set_gzda002_visible() #15/06/24 註解開放        
   #end add-point
 
   #將資料輸出到畫面上
   DISPLAY BY NAME g_gzda_m.gzda009,g_gzda_m.gzda002,g_gzda_m.gzda008,g_gzda_m.gzda010,g_gzda_m.gzda005,g_gzda_m.gzda006,g_gzda_m.gzda007,g_gzda_m.gzda001,g_gzda_m.gzda003,g_gzda_m.gzda004,g_gzda_m.gzdastus,g_gzda_m.gzdaownid,g_gzda_m.gzdaownid_desc,g_gzda_m.gzdaowndp,g_gzda_m.gzdaowndp_desc,g_gzda_m.gzdacrtid, 
       g_gzda_m.gzdacrtid_desc,g_gzda_m.gzdacrtdp,g_gzda_m.gzdacrtdp_desc,g_gzda_m.gzdacrtdt,g_gzda_m.gzdamodid,g_gzda_m.gzdamodid_desc,g_gzda_m.gzdamoddt
   
   #儲存PK
   LET l_ac = g_current_idx
   CALL azzi085_set_pk_array()
   
   #顯示狀態(stus)圖片
         #應用 a21 樣板自動產生(Version:3)
	  #根據當下狀態碼顯示圖片
      CASE g_gzda_m.gzdastus 
         WHEN "N"
            CALL gfrm_curr.setElementImage("statechange", "stus/32/inactive.png")
         WHEN "Y"
            CALL gfrm_curr.setElementImage("statechange", "stus/32/active.png")
         
      END CASE
 
 
 
 
   #顯示有特殊格式設定的欄位或說明
   CALL cl_show_fld_cont()
 
   #add-point:show段之後 name="show.after"
   
   #end add-point
 
END FUNCTION
 
{</section>}
 
{<section id="azzi085.delete" type="s" >}
#+ 資料刪除 
PRIVATE FUNCTION azzi085_delete() RETURNS ()   #200213-00032
   #add-point:delete段define(客製用) name="delete.define_customerization"
   
   #end add-point
   DEFINE  l_var_keys      DYNAMIC ARRAY OF STRING
   DEFINE  l_field_keys    DYNAMIC ARRAY OF STRING
   DEFINE  l_vars          DYNAMIC ARRAY OF STRING
   DEFINE  l_fields        DYNAMIC ARRAY OF STRING
   DEFINE  l_var_keys_bak  DYNAMIC ARRAY OF STRING
   #add-point:delete段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="delete.define"
   
   #end add-point  
   
   #add-point:Function前置處理  name="delete.pre_function"
   
   #end add-point
   
   #先確定key值無遺漏
   IF g_gzda_m.gzda001 IS NULL
 
   THEN
      INITIALIZE g_errparam TO NULL 
      LET g_errparam.extend = "" 
      LET g_errparam.code = "std-00003" 
      LET g_errparam.popup = FALSE 
      CALL cl_err()
      RETURN
   END IF
 
   CALL s_transaction_begin()
    
   LET g_gzda001_t = g_gzda_m.gzda001
 
   
   
 
   OPEN azzi085_cl USING g_gzda_m.gzda001
   IF SQLCA.SQLCODE THEN    #(ver:49)
      INITIALIZE g_errparam TO NULL 
      LET g_errparam.extend = "OPEN azzi085_cl:",SQLERRMESSAGE 
      LET g_errparam.code = SQLCA.SQLCODE
      LET g_errparam.popup = TRUE 
      CLOSE azzi085_cl
      CALL s_transaction_end('N','0')
      CALL cl_err()
      RETURN
   END IF
 
   #200804-00006 ---start---
   FETCH azzi085_cl
   IF SQLCA.SQLCODE THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.extend = "FETCH azzi085_cl:",SQLERRMESSAGE
      LET g_errparam.code = SQLCA.SQLCODE
      LET g_errparam.popup = TRUE
      CLOSE azzi085_cl
      #PostgreSQL需額外重新開啟, 無論主程式子程式都須關閉
      CALL s_transaction_end('N','0')
      CALL s_transaction_begin()
      CALL cl_err()
      RETURN
   END IF
   #200804-00006 --- end ---
 
   #顯示最新的資料
   EXECUTE azzi085_master_referesh USING g_gzda_m.gzda001 INTO g_gzda_m.gzda009,g_gzda_m.gzda002,g_gzda_m.gzda008,g_gzda_m.gzda010,g_gzda_m.gzda005,g_gzda_m.gzda006,g_gzda_m.gzda007,g_gzda_m.gzda001,g_gzda_m.gzda003,g_gzda_m.gzda004,g_gzda_m.gzdastus,g_gzda_m.gzdaownid,g_gzda_m.gzdaowndp,g_gzda_m.gzdacrtid, 
       g_gzda_m.gzdacrtdp,g_gzda_m.gzdacrtdt,g_gzda_m.gzdamodid,g_gzda_m.gzdamoddt,g_gzda_m.gzdaownid_desc,g_gzda_m.gzdaowndp_desc,g_gzda_m.gzdacrtid_desc,g_gzda_m.gzdacrtdp_desc,g_gzda_m.gzdamodid_desc
   
   
   #檢查是否允許此動作
   IF NOT azzi085_action_chk() THEN
      CALL s_transaction_end('N','0')
      RETURN
   END IF
   
   #遮罩相關處理
   LET g_gzda_m_mask_o.* =  g_gzda_m.*
   CALL azzi085_gzda_t_mask()
   LET g_gzda_m_mask_n.* =  g_gzda_m.*
   
   #將最新資料顯示到畫面上
   CALL azzi085_show()
   
   IF cl_ask_delete() THEN
 
      #add-point:單頭刪除前 name="delete.head.b_delete"
      IF NOT azzi085_chk_db_setup("3") THEN 
         RETURN 
      END IF  
      #end add-point
 
      #應用 a47 樣板自動產生(Version:4)
      #刪除相關文件
      CALL azzi085_set_pk_array()
      #add-point:相關文件刪除前 name="delete.befroe.related_document_remove"
      
      #end add-point   
      CALL cl_doc_remove()  
 
 
 
 
 
      DELETE FROM gzda_t 
       WHERE  gzda001 = g_gzda_m.gzda001 
 
 
      #add-point:單頭刪除中 name="delete.head.m_delete"
      
      #end add-point
         
      IF SQLCA.SQLCODE THEN
         INITIALIZE g_errparam TO NULL 
         LET g_errparam.extend = "gzda_t:",SQLERRMESSAGE 
         LET g_errparam.code = SQLCA.SQLCODE
         LET g_errparam.popup = FALSE 
         CALL s_transaction_end('N','0')
         CALL cl_err()
         RETURN   #161130-00056
      END IF
  
      
         #161130-00056
      
      #add-point:單頭刪除後 name="delete.head.a_delete"
      CALL azzi085_noticed_event("delete",g_gzda_m.gzda001)   #180312-00040#1 add azzi085_noticed_event
      #end add-point
      
       
 
      #修改歷程記錄(刪除)
      LET g_log1 = util.JSON.stringify(g_gzda_m)   #(ver:49)
      IF NOT cl_log_modified_record(g_log1,'') THEN    #(ver:49)
         CLOSE azzi085_cl
         CALL s_transaction_end('N','0')
         RETURN
      END IF
      
      CLEAR FORM
      CALL azzi085_ui_browser_refresh()
      
      #清空單頭資料
      INITIALIZE g_gzda_m.* TO NULL   #180420-00043
 
      #確保畫面上保有資料
      IF g_browser_cnt > 0 THEN
         #CALL azzi085_browser_fill(g_wc,"")
         CALL azzi085_fetch("P")
      ELSE
         CLEAR FORM
      END IF
      CALL s_transaction_end('Y','0')
   ELSE    
      CALL s_transaction_end('N','0')
   END IF
 
   CLOSE azzi085_cl
 
   #功能已完成,通報訊息中心
   CALL azzi085_msgcentre_notify('delete')
 
   #add-point:單頭刪除完成後 name="delete.a_delete"
   
   #end add-point
 
END FUNCTION
 
{</section>}
 
{<section id="azzi085.ui_browser_refresh" type="s" >}
#+ 瀏覽頁簽資料重新顯示
PRIVATE FUNCTION azzi085_ui_browser_refresh() RETURNS ()   #200213-00032
   #add-point:ui_browser_refresh段define(客製用) name="ui_browser_refresh.define_customerization"
   
   #end add-point
   DEFINE l_i  LIKE type_t.num10
   #add-point:ui_browser_refresh段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="ui_browser_refresh.define"
   
   #end add-point    
   
   #add-point:Function前置處理  name="ui_browser_refresh.pre_function"
   
   #end add-point
   
   LET g_browser_cnt = g_browser.getLength()
   LET g_header_cnt  = g_browser.getLength()
   FOR l_i =1 TO g_browser.getLength()
      IF g_browser[l_i].b_gzda001 = g_gzda_m.gzda001
 
         THEN
         CALL g_browser.deleteElement(l_i)
       END IF
   END FOR
   LET g_browser_cnt = g_browser_cnt - 1
   LET g_header_cnt = g_header_cnt - 1
   
   DISPLAY g_browser_cnt TO FORMONLY.b_count     #page count
   DISPLAY g_header_cnt  TO FORMONLY.h_count     #page count
  
   #若無資料則關閉相關功能
   IF g_browser_cnt = 0 THEN
      CALL cl_set_act_visible("statechange,modify,modify_detail,delete,reproduce,mainhidden", FALSE)
      CALL cl_navigator_setting(0,0)
      CLEAR FORM
   ELSE
      CALL cl_set_act_visible("mainhidden", TRUE)
   END IF
  
END FUNCTION
 
{</section>}
 
{<section id="azzi085.set_entry" type="s" >}
#+ 單頭欄位開啟設定
PRIVATE FUNCTION azzi085_set_entry(p_cmd) RETURNS ()   #200213-00032
   #add-point:set_entry段define(客製用) name="set_entry.define_customerization" 
   
   #end add-point
   DEFINE p_cmd LIKE type_t.chr1
   #add-point:set_entry段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="set_entry.define"
   
   #end add-point     
    
   #add-point:Function前置處理 name="set_entry.pre_function"
   
   #end add-point
   
   IF p_cmd = "a" THEN
      CALL cl_set_comp_entry("gzda001",TRUE)
      #根據azzi850使用者身分開關特定欄位
      IF NOT cl_null(g_no_entry) THEN
         CALL cl_set_comp_entry(g_no_entry,TRUE)
      END IF
      #add-point:set_entry段欄位控制 name="set_entry.field_control"
      
      #end add-point 
   END IF
   
   #add-point:set_entry段欄位控制後 name="set_entry.after_control"
   
   #end add-point 
 
END FUNCTION
 
{</section>}
 
{<section id="azzi085.set_no_entry" type="s" >}
#+ 單頭欄位關閉設定
PRIVATE FUNCTION azzi085_set_no_entry(p_cmd) RETURNS ()   #200213-00032
   #add-point:set_no_entry段define(客製用) name="set_no_entry.define_customerization"
   
   #end add-point
   DEFINE p_cmd LIKE type_t.chr1
   #add-point:set_no_entry段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="set_no_entry.define"
   
   #end add-point     
   
   #add-point:Function前置處理  name="set_no_entry.pre_function"
   
   #end add-point
   
   IF p_cmd = 'u' AND g_chkey = 'N' THEN
      CALL cl_set_comp_entry("gzda001",FALSE)
      #根據azzi850使用者身分開關特定欄位
      IF NOT cl_null(g_no_entry) THEN
         CALL cl_set_comp_entry(g_no_entry,FALSE)
      END IF
      #add-point:set_no_entry段欄位控制 name="set_no_entry.field_control"
      
      #end add-point 
   END IF
   
   #add-point:set_no_entry段欄位控制後 name="set_no_entry.after_control"
   
   #end add-point 
 
END FUNCTION
 
{</section>}
 
{<section id="azzi085.set_act_visible" type="s" >}
#+ 單頭權限開啟
PRIVATE FUNCTION azzi085_set_act_visible() RETURNS ()   #200213-00032
   #add-point:set_act_visible段define(客製用) name="set_act_visible.define_customerization" 
   
   #end add-point  
   #add-point:set_act_visible段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="set_act_visible.define"
   
   #end add-point
   #add-point:set_act_visible段 name="set_act_visible.set_act_visible"
   
   #end add-point
END FUNCTION
 
{</section>}
 
{<section id="azzi085.set_act_no_visible" type="s" >}
#+ 單頭權限關閉
PRIVATE FUNCTION azzi085_set_act_no_visible() RETURNS ()   #200213-00032
   #add-point:set_act_no_visible段define(客製用) name="set_act_no_visible.define_customerization"
   
   #end add-point
   #add-point:set_act_no_visible段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="set_act_no_visible.define"
   
   #end add-point
   #add-point:set_act_no_visible段 name="set_act_no_visible.set_act_no_visible"
   
   #end add-point
END FUNCTION
 
{</section>}
 
{<section id="azzi085.default_search" type="s" >}
#+ 外部參數搜尋
PRIVATE FUNCTION azzi085_default_search() RETURNS ()   #200213-00032
   #add-point:default_search段define(客製用) name="default_search.define_customerization" 
   
   #end add-point
   DEFINE li_idx  LIKE type_t.num10
   DEFINE li_cnt  LIKE type_t.num10
   DEFINE ls_wc   STRING
   #add-point:default_search段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="default_search.define"
   
   #end add-point  
   
   #add-point:Function前置處理  name="default_search.pre_function"
   
   #end add-point
   
   IF cl_null(g_order) THEN
      LET g_order = "ASC"
   END IF
   
   #add-point:default_search段開始前 name="default_search.before"
   
   #end add-point  
   
   #根據外部參數(g_argv)組合wc
   IF NOT cl_null(g_argv[01]) THEN
      LET ls_wc = ls_wc, " gzda001 = '", g_argv[01], "' AND "
   END IF
   
 
   
   #add-point:default_search段after sql name="default_search.after_sql"
   
   #end add-point  
   
   #200330-00044 ---start---
   IF NOT cl_null(g_default_wc) THEN
      LET g_wc = g_default_wc
      LET g_default = TRUE
   ELSE
   #200330-00044 --- end ---
      IF NOT cl_null(ls_wc) THEN
         #若有外部參數則根據該參數組合
         LET g_wc = ls_wc.subString(1,ls_wc.getLength()-5)
         LET g_default = TRUE
      ELSE
         #若無外部參數則預設為1=2
         LET g_default = FALSE
         #預設查詢條件
         LET g_wc = cl_qbe_get_default_qryplan()
         IF cl_null(g_wc) THEN
            LET g_wc = " 1=2"
         END IF
      END IF
   END IF   #200330-00044
   
   #add-point:default_search段結束前 name="default_search.after"
   
   #end add-point  
   
   #181128-00048 ---start---
   #若有外部串查條件,則以該條件為主
   IF NOT cl_null(g_default_wc) THEN
      LET g_wc = g_default_wc
   END IF
   #181128-00048 --- end ---
 
   IF g_wc.getIndexOf(" 1=2", 1) THEN
      LET g_default = TRUE
   END IF
 
 
END FUNCTION
 
{</section>}
 
{<section id="azzi085.mask_functions" type="s" >}
&include "erp/azz/azzi085_mask.4gl"
 
{</section>}
 
{<section id="azzi085.state_change" type="m" >}
   #應用 a09 樣板自動產生(Version:25)
#+ 確認碼變更 
PRIVATE FUNCTION azzi085_statechange() RETURNS ()   #200213-00032
   #add-point:statechange段define(客製用) name="statechange.define_customerization"

   #end add-point  
   DEFINE lc_state LIKE type_t.chr5
   #add-point:statechange段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="statechange.define"
   ###170731-00041 #1--add--start--
   DEFINE la_param        RECORD
                          gzda002  LIKE gzda_t.gzda002,
                          gzda003  LIKE gzda_t.gzda003,
                          gzda005  LIKE gzda_t.gzda005,
                          gzda006  LIKE gzda_t.gzda006,
                          gzda007  LIKE gzda_t.gzda007
                          END RECORD
   DEFINE l_i             LIKE type_t.num10
   DEFINE ls_js_n         STRING   #新值
   DEFINE ls_js_o         STRING   #舊值
   ###170731-00041 #1--add--end--
   #end add-point  
   
   #add-point:Function前置處理 name="statechange.before"

   #end add-point  
   
   ERROR ""     #清空畫面右下側ERROR區塊
 
   #180130-00006 ---start---
   #保存單頭舊值
   LET g_gzda_m_t.* = g_gzda_m.*
   LET g_gzda_m_o.* = g_gzda_m.*
   #180130-00006 --- end ---
 
   IF g_gzda_m.gzda001 IS NULL
 
   THEN
      INITIALIZE g_errparam TO NULL 
      LET g_errparam.extend = "" 
      LET g_errparam.code   = "std-00003" 
      LET g_errparam.popup = FALSE 
      CALL cl_err()
      RETURN
   END IF
 
   CALL s_transaction_begin()
   
   OPEN azzi085_cl USING g_gzda_m.gzda001
   IF SQLCA.SQLCODE THEN
   #  CLOSE azzi085_cl
   #  CALL s_transaction_end('N','0')   #(ver:19)
      INITIALIZE g_errparam TO NULL 
      LET g_errparam.extend = "OPEN azzi085_cl:",SQLERRMESSAGE
      LET g_errparam.code   = SQLCA.SQLCODE
      LET g_errparam.popup = TRUE
      CLOSE azzi085_cl
      CALL s_transaction_end('N','0')   #(ver:19)
      CALL cl_err()
      RETURN
   END IF
   
   #顯示最新的資料
   EXECUTE azzi085_master_referesh USING g_gzda_m.gzda001 INTO g_gzda_m.gzda009,g_gzda_m.gzda002,g_gzda_m.gzda008,g_gzda_m.gzda010,g_gzda_m.gzda005,g_gzda_m.gzda006,g_gzda_m.gzda007,g_gzda_m.gzda001,g_gzda_m.gzda003,g_gzda_m.gzda004,g_gzda_m.gzdastus,g_gzda_m.gzdaownid,g_gzda_m.gzdaowndp,g_gzda_m.gzdacrtid, 
       g_gzda_m.gzdacrtdp,g_gzda_m.gzdacrtdt,g_gzda_m.gzdamodid,g_gzda_m.gzdamoddt,g_gzda_m.gzdaownid_desc,g_gzda_m.gzdaowndp_desc,g_gzda_m.gzdacrtid_desc,g_gzda_m.gzdacrtdp_desc,g_gzda_m.gzdamodid_desc
   
 
   #檢查是否允許此動作
   IF NOT azzi085_action_chk() THEN
      CLOSE azzi085_cl
      CALL s_transaction_end('N','0')
      RETURN
   END IF
 
   #將資料顯示到畫面上
   DISPLAY BY NAME g_gzda_m.gzda009,g_gzda_m.gzda002,g_gzda_m.gzda008,g_gzda_m.gzda010,g_gzda_m.gzda005,g_gzda_m.gzda006,g_gzda_m.gzda007,g_gzda_m.gzda001,g_gzda_m.gzda003,g_gzda_m.gzda004,g_gzda_m.gzdastus,g_gzda_m.gzdaownid,g_gzda_m.gzdaownid_desc,g_gzda_m.gzdaowndp,g_gzda_m.gzdaowndp_desc,g_gzda_m.gzdacrtid, 
       g_gzda_m.gzdacrtid_desc,g_gzda_m.gzdacrtdp,g_gzda_m.gzdacrtdp_desc,g_gzda_m.gzdacrtdt,g_gzda_m.gzdamodid,g_gzda_m.gzdamodid_desc,g_gzda_m.gzdamoddt
 
   CASE g_gzda_m.gzdastus
      WHEN "N"
         CALL gfrm_curr.setElementImage("statechange", "stus/32/inactive.png")
      WHEN "Y"
         CALL gfrm_curr.setElementImage("statechange", "stus/32/active.png")
      
   END CASE
 
   #add-point:資料刷新後 name="statechange.after_refresh"

   #end add-point
 
   MENU "" ATTRIBUTES (STYLE="popup")
      BEFORE MENU
         HIDE OPTION "approved"
         HIDE OPTION "rejection"
         CASE g_gzda_m.gzdastus
            
            WHEN "N"
               HIDE OPTION "inactive"
            WHEN "Y"
               HIDE OPTION "active"
         END CASE
     
         #181128-00048 ---start---
         #181211 SA說暫時先mark此段功能
      #  #BPM功能
      #  CALL cl_set_act_visible("reunconfirmed",FALSE)
      #  CASE g_gzda_m.gzdastus
      #      WHEN "A"
      #         CALL cl_set_act_visible("reunconfirmed",TRUE)
      #      WHEN "D"
      #         CALL cl_set_act_visible("reunconfirmed",TRUE)
      #      WHEN "R"
      #         CALL cl_set_act_visible("reunconfirmed",TRUE)
      #  END CASE
         #181128-00048 --- end ---
 
      #add-point:menu前 name="statechange.before_menu"

      #end add-point
      
      
	  
      ON ACTION inactive
         IF cl_auth_chk_act("inactive") THEN
            LET lc_state = "N"
            #add-point:action控制 name="statechange.inactive"
            IF NOT azzi085_chk_db_setup("1") THEN 
               RETURN 
            END IF 
            #end add-point
         END IF
         EXIT MENU
      ON ACTION active
         IF cl_auth_chk_act("active") THEN
            LET lc_state = "Y"
            #add-point:action控制 name="statechange.active"

            #end add-point
         END IF
         EXIT MENU
 
      #add-point:stus控制 name="statechange.more_control"

      #end add-point
      
   END MENU
   LET INT_FLAG=FALSE  #(ver.18)
   
   #確認被選取的狀態碼在清單中
   IF (lc_state <> "N" 
      AND lc_state <> "Y"
      ) OR 
      g_gzda_m.gzdastus = lc_state OR cl_null(lc_state) THEN
      CLOSE azzi085_cl
      CALL s_transaction_end('N','0')
      RETURN
   END IF
   
   #add-point:stus修改前 name="statechange.b_update"

   #end add-point
   
   LET g_gzda_m.gzdamodid = g_user
   LET g_gzda_m.gzdamoddt = cl_get_current()
   LET g_gzda_m.gzdastus = lc_state
   
   #異動狀態碼欄位/修改人/修改日期
   UPDATE gzda_t 
      SET (gzdastus,gzdamodid,gzdamoddt) 
        = (g_gzda_m.gzdastus,g_gzda_m.gzdamodid,g_gzda_m.gzdamoddt)     
    WHERE  gzda001 = g_gzda_m.gzda001
 
    
   IF SQLCA.SQLCODE THEN
      #171011-00030 ---start---
      CLOSE azzi085_cl
      CALL s_transaction_end('N','0')
      #171011-00030 --- end ---
      INITIALIZE g_errparam TO NULL 
      LET g_errparam.extend = "" 
      LET g_errparam.code = SQLCA.SQLCODE 
      LET g_errparam.popup = FALSE 
      CALL cl_err()
      #170801-00033 ---start---
   #  CLOSE azzi085_cl   #171011-00030 mark
   #  CALL s_transaction_end('N','0')   #171011-00030 mark
      RETURN
      #170801-00033 --- end ---
   ELSE
      CASE lc_state
         WHEN "N"
            CALL gfrm_curr.setElementImage("statechange", "stus/32/inactive.png")
         WHEN "Y"
            CALL gfrm_curr.setElementImage("statechange", "stus/32/active.png")
         
      END CASE
 
      #180130-00006 ---start---
      #修改歷程記錄(狀態碼修改)
      LET g_log1 = util.JSON.stringify(g_gzda_m_t)
      LET g_log2 = util.JSON.stringify(g_gzda_m)
      IF NOT cl_log_modified_record(g_log1,g_log2) THEN
         CALL s_transaction_end('N','0')
         RETURN
      END IF
      #180130-00006 --- end ---
    
      #撈取異動後的資料
      EXECUTE azzi085_master_referesh USING g_gzda_m.gzda001 INTO g_gzda_m.gzda009,g_gzda_m.gzda002,g_gzda_m.gzda008,g_gzda_m.gzda010,g_gzda_m.gzda005,g_gzda_m.gzda006,g_gzda_m.gzda007,g_gzda_m.gzda001,g_gzda_m.gzda003,g_gzda_m.gzda004,g_gzda_m.gzdastus,g_gzda_m.gzdaownid,g_gzda_m.gzdaowndp,g_gzda_m.gzdacrtid, 
          g_gzda_m.gzdacrtdp,g_gzda_m.gzdacrtdt,g_gzda_m.gzdamodid,g_gzda_m.gzdamoddt,g_gzda_m.gzdaownid_desc,g_gzda_m.gzdaowndp_desc,g_gzda_m.gzdacrtid_desc,g_gzda_m.gzdacrtdp_desc,g_gzda_m.gzdamodid_desc
      
      #將資料顯示到畫面上
      DISPLAY BY NAME g_gzda_m.gzda009,g_gzda_m.gzda002,g_gzda_m.gzda008,g_gzda_m.gzda010,g_gzda_m.gzda005,g_gzda_m.gzda006,g_gzda_m.gzda007,g_gzda_m.gzda001,g_gzda_m.gzda003,g_gzda_m.gzda004,g_gzda_m.gzdastus,g_gzda_m.gzdaownid,g_gzda_m.gzdaownid_desc,g_gzda_m.gzdaowndp,g_gzda_m.gzdaowndp_desc,g_gzda_m.gzdacrtid, 
          g_gzda_m.gzdacrtid_desc,g_gzda_m.gzdacrtdp,g_gzda_m.gzdacrtdp_desc,g_gzda_m.gzdacrtdt,g_gzda_m.gzdamodid,g_gzda_m.gzdamodid_desc,g_gzda_m.gzdamoddt
   END IF
 
   #add-point:stus修改後 name="statechange.a_update"
   ###170731-00041 #1--add--start--
   #狀態碼異動後 紀錄重要資料修改歷程
   FOR l_i = 1 TO 2
      INITIALIZE la_param.* TO NULL
      # 利用JSON的方式組合字串
      IF l_i = 1 THEN
         LET la_param.gzda002 = g_gzda_m_t.gzda002
         LET la_param.gzda003 = g_gzda_m_t.gzda003
         LET la_param.gzda005 = g_gzda_m_t.gzda005
         LET la_param.gzda006 = g_gzda_m_t.gzda006
         LET la_param.gzda007 = g_gzda_m_t.gzda007
         LET ls_js_o = util.JSON.stringify(la_param)   # 舊值
      ELSE
         LET la_param.gzda002 = g_gzda_m.gzda002
         LET la_param.gzda003 = g_gzda_m.gzda003
         LET la_param.gzda005 = g_gzda_m.gzda005
         LET la_param.gzda006 = g_gzda_m.gzda006
         LET la_param.gzda007 = g_gzda_m.gzda007
         LET ls_js_n = util.JSON.stringify(la_param)   # 新值
      END IF
   END FOR
   CALL cl_log_modified_important(ls_js_o, ls_js_n)
   ###170731-00041 #1--add--end--
   #end add-point
 
   #add-point:statechange段結束前 name="statechange.after"

   #end add-point  
 
   CLOSE azzi085_cl
   CALL s_transaction_end('Y','0')
   
   #(ver:20) ---start---
   #add-point:transaction結束後 name="statechange.transaction_after"

   #end add-point  
   #(ver:20) --- end --- 
 
   #功能已完成,通報訊息中心
   CALL azzi085_msgcentre_notify('statechange:'||lc_state)
   
END FUNCTION
 
 
 
 
{</section>}
 
{<section id="azzi085.signature" type="s" >}
   
 
{</section>}
 
{<section id="azzi085.set_pk_array" type="s" >}
   #應用 a51 樣板自動產生(Version:10)
#+ 給予pk_array內容
PRIVATE FUNCTION azzi085_set_pk_array() RETURNS ()   #200213-00032
   #add-point:set_pk_array段define name="set_pk_array.define_customerization"
   
   #end add-point
   #add-point:set_pk_array段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="set_pk_array.define"
   
   #end add-point
   
   #add-point:Function前置處理 name="set_pk_array.before"
   
   #end add-point  
 
   #181105-00036 ---modify start---  
   #應檢查單頭是否有資料，而非檢查單身
#  #若l_ac<=0代表沒有資料
#  IF l_ac <= 0 THEN
#     RETURN
#  END IF
 
   IF cl_null(g_gzda_m.gzda001) THEN
      RETURN
   END IF
   #181105-00036 --- modify end ---
   
   CALL g_pk_array.clear()
   LET g_pk_array[1].values = g_gzda_m.gzda001
   LET g_pk_array[1].column = 'gzda001'
 
   
   #add-point:set_pk_array段之後 name="set_pk_array.after"
   
   #end add-point  
   
END FUNCTION
 
 
 
 
{</section>}
 
{<section id="azzi085.other_dialog" readonly="Y" type="s" >}
 
 
{</section>}
 
{<section id="azzi085.msgcentre_notify" type="s" >}
#應用 a66 樣板自動產生(Version:7)
PRIVATE FUNCTION azzi085_msgcentre_notify(lc_state) RETURNS ()   #200213-00032
   #add-point:msgcentre_notify段define name="msgcentre_notify.define_customerization"
   
   #end add-point   
   DEFINE lc_state LIKE type_t.chr80
   #add-point:msgcentre_notify段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="msgcentre_notify.define"
   
   #end add-point
   
   #add-point:Function前置處理  name="msgcentre_notify.pre_function"
   
   #end add-point
   
   INITIALIZE g_msgparam TO NULL
 
   #action-id與狀態填寫
   LET g_msgparam.state = lc_state
 
   #PK資料填寫
   CALL azzi085_set_pk_array()
   #單頭資料填寫
   LET g_msgparam.data[1] = util.JSON.stringify(g_gzda_m)
 
   #add-point:msgcentre其他通知 name="msgcentre_notify.process"
   
   #end add-point
 
   #呼叫訊息中心傳遞本關完成訊息
   CALL cl_msgcentre_notify()
 
END FUNCTION
 
 
 
 
{</section>}
 
{<section id="azzi085.action_chk" type="s" >}
#+ 修改/刪除前行為檢查(是否可允許此動作), 若有其他行為須管控也可透過此段落
PRIVATE FUNCTION azzi085_action_chk() RETURNS (INT)   #200213-00032
   #add-point:action_chk段define(客製用) name="action_chk.define_customerization" 
   
   #end add-point
   #add-point:action_chk段define(請盡量不要在客製環境修改此段落內容, 否則將後續patch的調整需人工處理) name="action_chk.define"
   
   #end add-point
   
   #add-point:action_chk段action_chk name="action_chk.action_chk"
   
   #end add-point
   
   RETURN TRUE
   
END FUNCTION
 
{</section>}
 
{<section id="azzi085.other_function" readonly="Y" type="s" >}
#DB 廠牌
PRIVATE FUNCTION azzi085_set_combo_dbtype()
   DEFINE ls_values  STRING 
   DEFINE ls_items   STRING 
   DEFINE lc_gzdb001 LIKE gzdb_t.gzdb001
   DEFINE lc_gzdb002 LIKE gzdb_t.gzdb002
   DEFINE ls_gzdb002 STRING  
  
   DECLARE azzi085_set_combo_db_cs CURSOR FOR
       SELECT  DISTINCT (gzdb001),gzdb002 FROM  gzdb_t 
         WHERE gzdbstus='Y' 
         ORDER BY gzdb001   
   IF SQLCA.SQLCODE THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code =  "lib-045"
      LET g_errparam.extend = "gzdb_t"
      LET g_errparam.popup = TRUE
      CALL cl_err()

      RETURN
   END IF
 
   LET ls_values = "" 
   LET ls_items = ""

   #在gzdb001和gzdb002原本空白處加入冒號(:)
   FOREACH azzi085_set_combo_db_cs INTO lc_gzdb001, lc_gzdb002
      LET ls_values = ls_values, lc_gzdb001 , ',' #值
      LET ls_items = ls_items, lc_gzdb002 CLIPPED, ',' #items
   END FOREACH
   LET ls_values = ls_values.subString(1, ls_values.getLength() - 1)
   LET ls_items = ls_items.subString(1, ls_items.getLength() - 1)

   CALL cl_set_combo_items("gzda006", ls_values, ls_items)
END FUNCTION
#DB版本
PRIVATE FUNCTION azzi085_set_combo_db_ver()
   DEFINE ls_values  STRING 
   DEFINE ls_items   STRING 
   DEFINE lc_gzdb003 LIKE gzdb_t.gzdb003
   DEFINE lc_gzdb004 LIKE gzdb_t.gzdb004
   DEFINE ls_gzdb003 STRING  
   DEFINE li         LIKE type_t.num5

   DECLARE azzi085_set_combo_db_ver_cs CURSOR FOR
    SELECT gzdb003,gzdb004 FROM gzdb_t 
     WHERE gzdbstus='Y' AND gzdb001 = g_gzda_m.gzda006  
     ORDER BY gzdb001  
   IF SQLCA.SQLCODE THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code =  "lib-045"
      LET g_errparam.extend = "gzdb_t"
      LET g_errparam.popup = TRUE
      CALL cl_err()
      RETURN
   END IF

   LET ls_values = "" 
   LET ls_items = ""
   LET li = 1 
   FOREACH azzi085_set_combo_db_ver_cs INTO lc_gzdb003, lc_gzdb004
      LET ls_values = ls_values, lc_gzdb003 , ',' #值
      LET ls_items = ls_items, lc_gzdb004 CLIPPED, ',' #items
   END FOREACH
   #DISPLAY "g_gzda_m.gzda007:",g_gzda_m.gzda007
   LET ls_values = ls_values.subString(1, ls_values.getLength() - 1)
   LET ls_items = ls_items.subString(1, ls_items.getLength() - 1)
   CALL cl_set_combo_items("gzda007", ls_values, ls_items) 
  #LET g_gzda_m.gzda007 = lc_gzdb003  #221101-00001#1
END FUNCTION

################################################################################
# Descriptions...: 變更密碼時要寫到gendbs.inc
# Memo...........:
# Usage..........: CALL azzi085_change_password()
#                  RETURNING 
# Input parameter: 
#                : 
# Return code....: 
#                : 
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi085_change_password()
   DEFINE lc_channel  base.Channel
   DEFINE ls_file     STRING 
   DEFINE ls_text     STRING 
   
   LET ls_file = os.Path.join(os.Path.join(FGL_GETENV("FBIN"),"4gl"),"gendbs.inc") 
   IF os.Path.delete(ls_file) THEN END IF
   #檔案
   LET lc_channel = base.Channel.create()
   CALL lc_channel.openFile(ls_file CLIPPED, "w" )
   CALL lc_channel.setDelimiter("")
   CALL lc_channel.write(g_gzda_m.gzda004)
   CALL lc_channel.close()
END FUNCTION

################################################################################
# Descriptions...: 設定隱藏
# Memo...........:
# Usage..........: CALL azzi085_set_gzda002_visible()
# Input parameter: 
# Return code....: 
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi085_set_gzda002_visible()
   CASE              
     WHEN g_gzda_m.gzda006 = "SQL Server" OR g_gzda_m.gzda006 = "FreeTDS"
        #FreeTDS is a set of libraries for Unix and Linux that allows your programs 
        #to natively talk to Microsoft SQL Server and Sybase databases.
        CALL cl_set_head_visible("gzda002","YES")  
     OTHERWISE 
        CALL cl_set_head_visible("gzda002","NO")
  END CASE 
END FUNCTION

################################################################################
# Descriptions...: 資料庫編號:ds開頭不可以更改失效及資料庫類型
#                  如果資料庫類型 = Y 當變成失效或delete 時，要檢查azzi500是否存在
#                  如果客製區DGENV = c 資料庫編號:ds開頭不可以更改失效及資料庫類型
#                  當資料庫類型變更其他類型時，要檢查azzi500是否存在
# Memo...........:
# Usage..........: CALL azzi085_chk_db_setup(pc_type)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi085_chk_db_setup(pc_type)
   DEFINE li_cnt   LIKE type_t.num5
   DEFINE pc_type  LIKE type_t.chr1

   CASE
      #此為系統用的連線資料庫，不可以變更失效或資料庫類型
      WHEN pc_type = "1" OR pc_type = "4" #1:失效  4:AFTER FIELD gzda005
         IF g_gzda_m.gzda001[1,2] = "ds" AND NOT g_gzda_m.gzda001 = "dscom" THEN 
            INITIALIZE g_errparam TO NULL 
            LET g_errparam.extend = "" 
            LET g_errparam.code   = "azz-00328"    #此為系統用的連線資料庫，不可以變更失效或資料庫類別
            LET g_errparam.popup  = TRUE 
            CALL cl_err()
            RETURN FALSE
         END IF


         IF NOT azzi085_chk_db_setup2(pc_type) THEN 
            RETURN FALSE
         END IF  

      WHEN pc_type = "2" OR pc_type = "3"  #2:AFTER FIELD gzda001  3:單頭刪除前
         #如果客製區DGENV = c 資料庫編號:ds開頭不可以更改失效及資料庫類型
         IF FGL_GETENV("DGENV") = "c" THEN 
            IF g_gzda_m.gzda001[1,2] = "ds" THEN 
               INITIALIZE g_errparam TO NULL 
               LET g_errparam.extend = "" 
               LET g_errparam.code   = "azz-00329" 
               LET g_errparam.popup  = TRUE 
               CALL cl_err()
               RETURN FALSE 
            END IF 
         END IF 
         IF pc_type = "3" THEN 
            IF NOT azzi085_chk_db_setup2(pc_type) THEN 
               RETURN FALSE
            END IF
         END IF 
   END CASE 
 
   RETURN TRUE  
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzi085_chk_db_setup2(pc_type)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi085_chk_db_setup2(pc_type)
   DEFINE li_cnt           LIKE type_t.num5
   DEFINE pc_type          LIKE type_t.chr1
   DEFINE ls_code          STRING
   DEFINE ls_gzda004_orig  STRING    
   
   #如果資料庫類型 = Y 當變成失效或delete 時，要檢查azzi500是否存在
   #當資料庫類型 切換 其他類型時，要檢查azzi500是否存在
   IF g_gzda_m.gzda005 = "Y" OR (pc_type = "4" AND g_gzda_m.gzda005 <> "Y") THEN  
#      SELECT COUNT(*) INTO li_cnt FROM gzou_t   #170413-00060#2 mark
      SELECT COUNT(1) INTO li_cnt FROM gzou_t    #170413-00060#2 add
       WHERE gzou001 = g_enterprise 
         AND gzou003 = g_gzda_m.gzda001
      IF li_cnt > 0 THEN 
         INITIALIZE g_errparam TO NULL 
         LET g_errparam.extend = "" 

         IF pc_type = "4" THEN 
            LET ls_code = "azz-00330"  #在azzi500有此筆資料，不可變更資料庫類型  
         ELSE 
            LET ls_code = "azz-00324"  #此為系統用的連線資料庫，不可以切換為失效或刪除
         END IF 
        
         LET g_errparam.code   = ls_code  
         LET g_errparam.popup  = TRUE 
         CALL cl_err()
         RETURN FALSE 
     END IF
     
     #170628-00013 start
     LET ls_gzda004_orig = cl_hashkey_base65_anti(g_gzda_m.gzda004)    
     IF azzi085_create_new_db_adz_init(g_gzda_m.gzda003,ls_gzda004_orig) THEN 
     ELSE
         INITIALIZE g_errparam TO NULL 
         LET g_errparam.extend = ""
         LET g_errparam.code   = "azz-00446"             #無法建立初始化資料:%1，資料庫編號:%2
         LET g_errparam.popup  = TRUE 
         LET g_errparam.replace[1] = ms_crt_db_adz_init  #240709-00026#1
         LET g_errparam.replace[2] = g_gzda_m.gzda001
         CALL cl_err()
         RETURN FALSE
     END IF
     #170628-00013 end     
   END IF 

   IF pc_type = "3" THEN 
      #160922-00004 start
      #ds資料庫 不可以刪除
      IF g_gzda_m.gzda001 = "ds" OR g_gzda_m.gzda001 = "dsaps" OR g_gzda_m.gzda001 = "dsaws" OR 
         g_gzda_m.gzda001 = "dscom" OR g_gzda_m.gzda001 = "dsdata"  OR g_gzda_m.gzda001 = "dsdemo" OR 
         g_gzda_m.gzda001 = "dspatch" OR g_gzda_m.gzda001 = "dsrept" OR g_gzda_m.gzda001 = "scm"   THEN   #170216-00032  
          
         #170216-00032 start
         #在客戶家環境是可以刪除scm 因為scm 是在產中使用
         IF FGL_GETENV("DGENV") = "c" AND g_gzda_m.gzda001 = "scm" THEN
            RETURN TRUE
         END IF  
         #170216-00032 end
         INITIALIZE g_errparam TO NULL 
         LET g_errparam.extend = "" 
         LET g_errparam.code   = "azz-00324"  #此為系統用的連線資料庫，不可以切換為失效或刪除
         LET g_errparam.popup  = TRUE 
         CALL cl_err()
         RETURN FALSE 
       #180724-00062#-start--       
       ELSE          
         IF g_t100debug = 9 THEN
            RETURN TRUE  
         END IF 
         INITIALIZE g_errparam TO NULL 
         LET g_errparam.extend = "" 
         LET g_errparam.code   = "azz-00485"  #不能刪除，只能設為失效
         LET g_errparam.popup  = TRUE 
         CALL cl_err()
         RETURN FALSE 
       #180724-00062#-end--         
       END IF 
       #160922-00004 end  
   END IF 
   RETURN TRUE 
END FUNCTION

################################################################################
# Descriptions...: 設定對應中間庫combobox
# Memo...........:
# Usage..........: CALL azzi085_set_combo_middle_db()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 16/11/30 By jrg542
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi085_set_combo_middle_db()
   #161124-00001 start
   DEFINE ls_values  STRING 
   DEFINE ls_items   STRING 
   DEFINE lc_gzdb001 LIKE gzdb_t.gzdb001
   DEFINE lc_gzdb002 LIKE gzdb_t.gzdb002
   DEFINE ls_gzdb002 STRING  
  
   DECLARE azzi085_set_combo_middle_db_cs CURSOR FOR
    SELECT DISTINCT gzda001 FROM gzda_t 
     WHERE gzdastus ='Y'
       AND gzda005 = 'M'
   IF SQLCA.SQLCODE THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code =  "lib-045"
      LET g_errparam.extend = "gzda_t"
      LET g_errparam.popup = TRUE
      CALL cl_err()
      RETURN
   END IF
 
   FOREACH azzi085_set_combo_middle_db_cs INTO lc_gzdb001
      LET ls_values = ls_values, lc_gzdb001 , ',' #值
      LET ls_items = ls_items, lc_gzdb001 CLIPPED, ',' #items
   END FOREACH
   LET ls_values = ls_values.subString(1, ls_values.getLength() - 1)
   LET ls_items = ls_items.subString(1, ls_items.getLength() - 1)

   CALL cl_set_combo_items("gzda010", ls_values, ls_items)
   #161124-00001 end
END FUNCTION

################################################################################
# Descriptions...: 建立db synonym
# Memo...........:
# Usage..........: CALL azzi085_create_new_db_adz_init()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2017/06/28 By jrg542
# Modify.........: #170628-00013 #1
################################################################################
PRIVATE FUNCTION azzi085_create_new_db_adz_init(p_db_user,p_db_passwd)
   DEFINE p_db_user      STRING,    
          p_db_passwd    STRING
   DEFINE ls_db_user     STRING,
          ls_db_passwd   STRING,
          ls_exec_sql    STRING,
          ls_command     STRING,
          ls_root_path   STRING,
          ls_disql_bef   STRING, #231003-00004#1
          ls_disql_cmd   STRING, #231003-00004#1
          ls_disql_aft   STRING, #231003-00004#1
          ls_db_type     STRING, #231003-00004#1
          ls_db_source   STRING  #231003-00004#1
   DEFINE li_return      LIKE type_t.num5  
  
   LET ls_db_user   = p_db_user   
   LET ls_db_passwd = p_db_passwd 

   LET li_return = TRUE

   #LET ls_root_path = os.Path.join(os.Path.join(FGL_GETENV("UTL"),"bin"),ms_crt_db_adz_init)  #240709-00026#1
   LET ls_exec_sql = os.Path.join(os.Path.join(FGL_GETENV("UTL"),"bin"),ms_crt_db_adz_init)  #240709-00026#1

   IF os.Path.exists(ls_exec_sql) THEN 
     CALL FGL_DB_DRIVER_TYPE() RETURNING ls_db_type #231003-00004#1
     
     #231003-00004#1 begin
     IF NVL(ls_db_type,"N/A") = "dmg" THEN 
       LET ls_db_source = FGL_GETENV("DM_DATABASE")
       LET ls_disql_bef = "export DM_SVC_PATH_ORG=$DM_SVC_PATH;mkdir -p ~/.dm_svc;cp -f $DM_SVC_PATH/dm_svc.conf ~/.dm_svc;sed -i 's/LANGUAGE=.*/LANGUAGE='\\(en\\)'/' ~/.dm_svc/dm_svc.conf;export DM_SVC_PATH=~/.dm_svc;"
       LET ls_disql_cmd = "disql ",ls_db_user,"/",ls_db_passwd,"@",ls_db_source," \\`",ls_exec_sql," -c 'SET ECHO OFF';"
       LET ls_disql_aft = "export DM_SVC_PATH=$DM_SVC_PATH_ORG;unset DM_SVC_PATH_ORG;" -- "rm -f ~/.dm_svc/dm_svc.conf;"
       LET ls_command = ls_disql_bef,ls_disql_cmd,ls_disql_aft
     ELSE
     #231003-00004#1 end 
       #240709-00026#1 begin
       #250415-00028#1 modify(s)
       CASE
         WHEN g_db_branch = 'KINGBASE'
           LET ls_command = "ksql 'user=",ls_db_user," password=",ls_db_passwd,"' -f '",ls_exec_sql,"' -t"
         WHEN g_db_branch = 'MOGDBA'
           LET ls_command =  "export MOGDB_GSQL_SUPPORT_EXIT=ON;",
                          "gsql -U ",ls_db_user," -W ",ls_db_passwd," -f ",ls_exec_sql," -t"
         OTHERWISE
           LET ls_command = "sqlplus ",ls_db_user,"/",ls_db_passwd," @",ls_exec_sql
       END CASE
       #IF g_db_branch = 'KINGBASE' THEN
       #  LET ls_command = "ksql 'user=",ls_db_user," password=",ls_db_passwd,"' -f '",ls_exec_sql,"' -t"
       #ELSE
       #  LET ls_command = "sqlplus ",ls_db_user,"/",ls_db_passwd," @",ls_exec_sql
       #END IF
       #250415-00028#1 modify(e)
       #240709-00026#1 end      
     END IF #231003-00004#1
     
     DISPLAY "[COMMAND]",ls_command
     RUN ls_command
   ELSE
     LET li_return = FALSE
     DISPLAY "[ERROR]Could not find SQL file : ",ms_crt_db_adz_init  #240709-00026#1
   END IF  

   RETURN li_return 
END FUNCTION

################################################################################
# Descriptions...: 輸入DBA 密碼
# Memo...........:
# Usage..........: CALL azzi085_input_dba_password()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2018/01/31 By 01856
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi085_input_dba_password()
   DEFINE ls_system_pw  STRING 
   DEFINE ls_sys_pw     STRING 
   
   #Fix for GBC add --- start ---
   #OPEN WINDOW azzi085_input_dba_w WITH FORM cl_ap_formpath("azz","azzi085_s01") #Fix for GBC mark
   IF FGL_GETENV('T100NANAWEB') = 'Y' THEN
      OPEN WINDOW azzi085_input_dba_w WITH FORM cl_ap_formpath("azz","azzi085_s01") ATTRIBUTES(STYLE='layoutwin')
   ELSE
      OPEN WINDOW azzi085_input_dba_w WITH FORM cl_ap_formpath("azz","azzi085_s01")
   END IF
   #Fix for GBC add --- end ---
   CALL cl_ui_init()

   DIALOG ATTRIBUTES(UNBUFFERED)

      INPUT ls_system_pw, ls_sys_pw 
       FROM formonly.system_pw, formonly.sys_pw
         ATTRIBUTES(WITHOUT DEFAULTS)
         BEFORE INPUT

      END INPUT
   BEFORE DIALOG

   ON ACTION accept
      ACCEPT DIALOG
 
      ON ACTION cancel
         LET INT_FLAG = TRUE 
         EXIT DIALOG
      #在dialog 右上角 (X)
      ON ACTION CLOSE  
         LET INT_FLAG = TRUE 
         EXIT DIALOG
   
   END DIALOG 

   CLOSE WINDOW azzi085_input_dba_w
   RETURN ls_system_pw,ls_sys_pw
END FUNCTION

################################################################################
# Descriptions...: 查看shell db user 建立完成
# Memo...........:
# Usage..........: CALL azzi085_monitor_create_dbuser()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2018/01/31 By 01856
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi085_monitor_create_dbuser()
   DEFINE lch_pipe         base.Channel
   DEFINE ls_tmp           STRING
   DEFINE ls_cmd           STRING
   DEFINE ls_temp          STRING
   DEFINE li_chk           LIKE type_t.num5
   DEFINE li_cnt           LIKE type_t.num5

   LET ls_cmd = "ps -ef | grep createdbuser | grep $ZONE  | grep ",gs_id ," | grep -v 'sh -c'| wc -l "

   DISPLAY "ls_cmd:",ls_cmd

   WHILE TRUE
      LET li_cnt = 0
      LET ls_tmp = ""
      LET lch_pipe = base.Channel.create()
      CALL lch_pipe.openPipe(ls_cmd, "r")

      WHILE lch_pipe.read(ls_tmp)
           DISPLAY "ls_tmp:",ls_tmp
           IF ls_tmp.trim() IS NULL THEN CONTINUE WHILE END IF
           LET li_cnt = ls_tmp
      END WHILE 
      CALL lch_pipe.close()
      IF li_cnt > 0 THEN
         SLEEP 1
      ELSE 
         EXIT WHILE    
      END IF    
   END WHILE   
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzi085_noticed_event(pc_action_id,pc_gzda001)
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: #180312-00040#1 by 08146
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi085_noticed_event(pc_action_id,pc_gzda001)
   DEFINE pc_gzda001       LIKE gzda_t.gzda001
   #DEFINE ps_action_id     STRING                        #180724-00062#1   add
   DEFINE pc_action_id     LIKE gzzq_t.gzzq002            #180724-00062#1   add
   DEFINE ls_reponse       STRING 
#180723-00031#3--start--    
#   DEFINE lc_event         RECORD
#      lc_user                LIKE gzxa_t.gzxa003,         #user:操作者(g_user)
#      ls_action              STRING ,                     #action:行為(insert,modify,delete,etc...) 包含 生失效
#      ld_time                DATETIME YEAR TO SECOND ,    #time:時間 
#      ls_prog                STRING ,                     #prog:作業代碼(azzi500,azzi085)
#      lc_cust                STRING ,                     #cust:客戶代碼(客服編號) 客服編號
#      lc_key                 STRING                       #key:鍵值欄位(ent代碼,DB代碼) 
#
#                END RECORD
   DEFINE lc_event         RECORD
      user                LIKE gzxa_t.gzxa003,         #user:操作者(g_user)
      action              STRING ,                     #action:行為(insert,modify,delete,etc...) 包含 生失效
      time                DATETIME YEAR TO SECOND ,    #time:時間 
      prog                STRING ,                     #prog:作業代碼(azzi500,azzi085)
      cust                STRING ,                     #cust:客戶代碼(客服編號) 客服編號
      key                 STRING                       #key:鍵值欄位(ent代碼,DB代碼) 
                END RECORD  
#180723-00031#3--end--                 
   DEFINE  ls_js                  STRING 
   DEFINE lc_gzzq004       LIKE gzzq_t.gzzq004            #180724-00062#1    add
--ls_event_desc請傳入以下資訊, 使用json字串:
 --
--1. user:操作者(g_user)
--2. action:行為(insert,modify,delete,etc...) 包含 生失效
--3. time:時間
--4. prog:作業代碼(azzi500,azzi085)
--5. cust:客戶代碼(客服編號) 客服編號
--6. key:鍵值欄位(ent代碼,DB代碼)  

   #180724-00062#1--start--
   SELECT gzzq004 INTO lc_gzzq004 FROM gzzq_t
    WHERE gzzq001 = 'standard'
     AND  gzzq002 = pc_action_id
     AND  gzzq003 = g_dlang
     AND  gzzq006 = 's'
 
#180723-00031#3--start-- 
   #LET lc_event.lc_user = g_user
   #LET lc_event.ls_action = ps_action_id
#   LET lc_event.lc_user = g_user,"(",cl_get_username(g_user),")"
#   LET lc_event.ls_action = lc_gzzq004 
#   #180724-00062#1--end--   
#   LET lc_event.ld_time = cl_get_current()
#   #180724-00062#1--start--
#   #LET lc_event.ls_prog = g_prog
#   LET lc_event.ls_prog = cl_get_progname(g_prog,g_dlang,"2")
#   #180724-00062#1--end--   
#   LET lc_event.lc_cust = FGL_GETENV("CUST")
#   LET lc_event.lc_key = pc_gzda001
   
   LET lc_event.user = g_user
   LET lc_event.action = pc_action_id,"(",lc_gzzq004,")"   
   LET lc_event.time = cl_get_current()
   LET lc_event.prog = cl_get_progname(g_prog,g_dlang,"2")  
   LET lc_event.cust = FGL_GETENV("CUST")
   LET lc_event.key = pc_gzda001
#180723-00031#3--end--   
   
   LET ls_js = util.JSON.stringify(lc_event)
   CALL cl_set_noticed_event(ls_js) RETURNING ls_reponse 
   
   #IF ls_reponse.getindexOf("ERROR",1) THEN  #180723-00031#3  mark        
   IF ls_reponse.getindexOf("Error",1) THEN   #180723-00031#3  add  
      DISPLAY " azzi085 寫入重要事件紀錄 有錯誤:",ls_reponse
   END IF 

END FUNCTION

################################################################################
# Descriptions...: 大數據讀取 資料庫認證資訊
#
#                  重產multipoijdbc.ini
#                  範例如下
#                  topprd.dsdemo.host=10.40.41.146
#                  topprd.dsdemo.port=1521
#                  topprd.dsdemo.user=dsdemo
#                  topprd.dsdemo.pass=<REDACTED>
#
# Memo...........: 210114-00032#4
# Usage..........: CALL azzi085_create_multipoijdbc() 
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2021/01/14 By 12731
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi085_create_multipoijdbc()

   DEFINE lr_gzda       RECORD
          gzda001       LIKE gzda_t.gzda001, #資料庫名稱
          gzda002       LIKE gzda_t.gzda002, #資料庫主機IP(host)
          gzda003       LIKE gzda_t.gzda003, #登入連線user(user)
          gzda004       LIKE gzda_t.gzda004  #資料庫密碼(pass)
          END RECORD
   DEFINE l_channel     base.Channel
   DEFINE l_sql         STRING
   DEFINE l_file        STRING
   DEFINE l_path        STRING
   DEFINE l_text        STRING
   DEFINE l_num         INTEGER  #211104-00032#1 add
   
   #211220-00001#1 add 連線檢查測試 (s)
   DEFINE conn          Connection
   DEFINE stat          Statement
   #211220-00001#1 add 連線檢查測試 (e)

   #取得ds除外的所有的資料庫IP與密碼
   LET l_sql = "SELECT gzda001,gzda002,gzda003,gzda004 FROM gzda_t ",
   #            " WHERE gzda003 <> 'ds' AND gzda005 = 'Y' AND gzdastus = 'Y' ORDER BY gzda001 "  #210712-00018#1 mark
   #210712-00018#1 add (s) 調整不管理設定 資料庫類型 是否為 ERP 專用資料庫
                " WHERE gzda003 <> 'ds' ",
                "   AND gzdastus = 'Y' ",
                " ORDER BY gzda001 "
   #210712-00018#1 add (e)            
   DECLARE azzi085_create_ini_cs CURSOR FROM l_sql

   LET l_channel = base.Channel.create()

   #240319-00048#1 mark-s
   #211104-00032#1 add (s)
   #LET l_text = "ls -l $ORACLE_HOME/jdbc/lib/ojdbc6.jar | wc -l"
   #CALL    l_channel.openPipe(l_text, "r")
   #LET     l_num = l_channel.readLine()
   #IF l_num > 0 THEN
       
      #LET l_text = "ls -l $ADZ/42m/adz_adzp523.42m | wc -l"
      #CALL    l_channel.openPipe(l_text, "r")
      #LET     l_num = l_channel.readLine()
      #IF l_num < 1 THEN
         #LET l_text = "請檢查 adz 模組 42m 目錄有 adz_adzp523.42m \n",
                      #"沒有請嘗試到 $ADZ/4gl 目錄 r.c adzp523 "
         #MENU "錯誤" ATTRIBUTE(STYLE="dialog",COMMENT=l_text,IMAGE="stop")
             #ON ACTION ok
                #EXIT MENU
         #END MENU
         #RETURN FALSE
      #221206-00021#1 add (s)
      #ELSE
         #LET l_text = "ls -l $ADZ/42r/adzp523.42r | wc -l"
         #CALL    l_channel.openPipe(l_text, "r")
         #LET     l_num = l_channel.readLine()
         #IF l_num < 1 THEN
            #LET l_text = "請檢查 adz 模組 42r 目錄有 adzp523.42r \n",
                         #"沒有請嘗試到 $ADZ/4gl 目錄 r.l adzp523 "
            #MENU "錯誤" ATTRIBUTE(STYLE="dialog",COMMENT=l_text,IMAGE="stop")
                #ON ACTION ok
                   #EXIT MENU
            #END MENU
            #RETURN FALSE
         #END IF
      #221206-00021#1 add (e)
      #END IF
      #240319-00048#1 mark-e

   #ELSE
      #LET l_text = "請檢查 $ORACLE_HOME/jdbc/lib/ojdbc6.jar 是否存在 \n",
                   #"不存在請到 /u1/src/oracle 找到檔案放到 \n",
                   #"$ORACLE_HOME/jdbc/lib \n",
                   #"額外檢查下 \n",
                   #" adz 模組目錄內 42m 有 adzp523.42m "
      #MENU "錯誤" ATTRIBUTE(STYLE="dialog",COMMENT=l_text,IMAGE="stop")
          #ON ACTION ok
             #EXIT MENU
      #END MENU
      #RETURN FALSE
   #END IF
   #211104-00032#1 add (e)

   #211220-00001#1 add (s)  檢查環境變數設定
   IF NOT cl_null(FGL_GETENV("CLASSPATH")) THEN
      LET l_text = "echo $CLASSPATH | grep utilitytools | wc -l "
      CALL    l_channel.openPipe(l_text, "r")
      LET     l_num = l_channel.readLine()
      IF l_num > 0 THEN
         LET l_text = "echo $CLASSPATH | grep ojdbc6.jar | wc -l "
         CALL    l_channel.openPipe(l_text, "r")
         LET     l_num = l_channel.readLine()
         IF l_num < 0 THEN
            DISPLAY "echo $CLASSPATH | grep ojdbc6.jar | wc -l"
            LET l_file = "N"
         END IF
      ELSE
         DISPLAY "echo $CLASSPATH | grep utilitytools | wc -l"
         LET l_file = "N"
      END IF

   ELSE 
      DISPLAY "cl_null(FGL_GETENV('CLASSPATH'))"
      LET l_file = "N"
   END IF

   IF l_file = "N" THEN

      LET l_text = "請檢查環境變數 CLASSPATH 是不是缺少 \n ",
                   "utilitytools.jar 和 ojdbc6.jar \n",
                   "調整好請重開 azzi085 程序 "
      MENU "錯誤" ATTRIBUTE(STYLE="dialog",COMMENT=l_text,IMAGE="stop")
       ON ACTION ok
          EXIT MENU
      END MENU
      RETURN FALSE
   END IF
   #211220-00001#1 add (e)

   LET l_file = os.Path.join(os.Path.join(os.Path.join(os.Path.join(FGL_GETENV("TOP"),"utl"),"java"),"jar"),"multipoijdbc.ini")
   
   #210712-00018#1 mark (s) 刪除檔案需要判斷目錄權限 ， 調整為確認檔案存在 ， 重新撰寫到檔案內 ( 覆蓋原本內容 )
   #刪除舊檔案
#   IF os.Path.delete(l_file) THEN END IF
#
#   IF NOT os.Path.exists(l_file) THEN
#      #舊檔案刪除成功
#      DISPLAY cl_getmsg("adz-01165",g_lang) CLIPPED,":",l_file
#   ELSE
#      #舊檔案刪除失敗
#      INITIALIZE g_errparam TO NULL
#
#      LET g_errparam.code = 'adz-01144'
#      LET g_errparam.extend = l_file
#      LET g_errparam.popup = TRUE
#      CALL cl_err()
#      RETURN FALSE
#   END IF
   #210712-00018#1 mark (e) 刪除檔案需要判斷目錄權限 ， 調整為確認檔案存在 ， 重新撰寫到檔案內 ( 覆蓋原本內容 )

   #210823-00002#1 add (s)
   #211104-00032#1 add (s)
   LET l_sql = "select substr(a,1,instr(a,')')-1) as ip ",
               "  from ( select substr(v2.value,instr(v2.value,'HOST')+5,instr(v2.value,'(PORT')-(instr(v2.value,'HOST')+5)) a ",
               #透過 LOCAL LISTENER 與 SERVICE NAME 來篩選
               "           from v$listener_network v1 left join v$listener_network v2 ",
               #LOCAL LISTENER 為 PORT 格式能用欄位顯示
               "             on v2.con_id = v1.con_id and v2.type = 'LOCAL LISTENER' ",
               #抓取 SERVICE NAME 確認返回值為 1 個
               "          where v1.type = 'SERVICE NAME' and lower(v1.value) = '",FGL_GETENV('ORACLE_SID'),"' ) "
   PREPARE azzi085_getip FROM l_sql #241008-00026 -add
   EXECUTE azzi085_getip INTO g_ip  #241008-00026 -add
   #PREPARE azzi085_getip FROM l_sql #241008-00026 -mark    
   #EXECUTE azzi085_getip INTO g_ip  #241008-00026 -mark   
   #211104-00032#1 add (e)

   #220503-00021#1 add (s)
   LET l_text = "echo '",FGL_GETENV('ORACLE_SID'),"' | awk '{print toupper($0)}'"
   CALL    l_channel.openPipe(l_text, "r")
   LET     l_text = l_channel.readLine()   

   LET l_text = "grep -A 10 ",'"',l_text,'" $ORACLE_HOME/network/admin/tnsnames.ora ',
                " | sed -n '/ADDRESS/, /ADDRESS/p' | grep HOST ",
                " | awk -F 'HOST =' '{print $2}' | awk -F ')' '{print $1}' | sed -n '1,1p' "
   DISPLAY "get ip : ",l_text
   CALL    l_channel.openPipe(l_text, "r")
   LET     l_text = l_channel.readLine()
   
   
   
   DISPLAY "config file to get ip : ",l_text
   
   LET l_sql = "SELECT DISTINCT gzda002 FROM gzda_t ",
                " WHERE gzda003 = 'dsdemo' ",
                "   AND gzdastus = 'Y' "
   
   PREPARE jdbc_getip FROM l_sql
   EXECUTE jdbc_getip INTO lr_gzda.gzda002

   #1. tnsnames.ora 不為 null
   #讓 ip = tnsnames.ora
   #2. v$listener_network , tnsnames.ora 為 null
   #讓 ip = azzi085 dsdemo gzda002
   # -----------------------------
   #判斷 azzi085 dsdemo gzda002 不可有 : 分號
   
   LET l_num = 0
   LET l_sql = lr_gzda.gzda002 CLIPPED
   IF l_sql.getIndexOf(":",1) > 0 THEN
      LET l_sql = ""
      LET l_num = 1
   END IF
   
   IF cl_null(g_ip) AND cl_null(l_text) AND cl_null(l_sql) THEN
      IF l_num = 0 THEN
         LET l_text = "azzi085 not find ip , Please enter db ip on user dsdemo "
      ELSE
         LET l_text = "Please check /etc/hosts has DB ip and hostname"
      END IF
      MENU "錯誤" ATTRIBUTE(STYLE="dialog",COMMENT=l_text,IMAGE="")
       ON ACTION ok
          EXIT MENU
      END MENU
      RETURN FALSE
   ELSE
      IF cl_null(g_ip) AND cl_null(l_text) THEN
         LET g_ip = l_sql CLIPPED
      ELSE
         #IF cl_null(g_ip) THEN  #221206-00021#1 mark
         IF NOT cl_null(l_text) THEN  #221206-00021#1 add
            LET g_ip = l_text
         END IF
      END IF
   END IF            
   #220503-00021#1 add (e)

   #ORALCE listener 格式 ( ... (PORT=1521)...)
   #240618-00027 -add -s
   IF g_db_branch = 'KINGBASE' THEN
      #LET l_sql = "SELECT setting "," FROM pg_settings "," WHERE lower(name) = 'port' "         #250923-00007 mark
      LET l_sql = "SELECT DISTINCT setting "," FROM pg_settings "," WHERE lower(name) = 'port' " #250923-00007 add
   ELSE
      #240618-00027 -add -e
      #LET l_sql = "select substr(a,1,instr(a,')')-1) as port ",  #找到左括號 ) 位置後扣除括號 ( 1 個位數 )  #250923-00007 mark
      LET l_sql = "select DISTINCT substr(a,1,instr(a,')')-1) as port ",  #找到左括號 ) 位置後扣除括號 ( 1 個位數 )  #250923-00007 add
                  #找到 PORT 取 5 個值之後 15 個位數 ( 不確定 PORT 長度直接找 15 個位數 )
                  "  from ( select substrb(v2.value,instr(v2.value,'PORT')+5,15) a ",
                  #透過 LOCAL LISTENER 與 SERVICE NAME 來篩選
                  "           from v$listener_network v1 left join v$listener_network v2 ",
                  #LOCAL LISTENER 為 PORT 格式能用欄位顯示
                  "             on v2.con_id = v1.con_id and v2.type = 'LOCAL LISTENER' ",
                  #抓取 SERVICE NAME 確認返回值為 1 個
                  "          where v1.type = 'SERVICE NAME' and lower(v1.value) = '",FGL_GETENV('ORACLE_SID'),"' ) "
   END IF #240618-00027 -add -s
   PREPARE azzi085_getport FROM l_sql
   EXECUTE azzi085_getport INTO g_port
   #210823-00002#1 add (e)

   CALL l_channel.openFile(l_file, "w" )

   IF os.Path.exists(l_file) THEN
      FOREACH azzi085_create_ini_cs INTO lr_gzda.*
         #IF NOT (cl_null(lr_gzda.gzda002) OR cl_null(lr_gzda.gzda003) OR cl_null(lr_gzda.gzda004)) THEN  #211220-00001#1 mark
         IF NOT (cl_null(lr_gzda.gzda003) OR cl_null(lr_gzda.gzda004)) THEN  #211220-00001#1 add 保留帳密需輸入
            #topprd.dsdemo.host=10.40.41.146
            LET l_text = FGL_GETENV("ZONE") CLIPPED,".",lr_gzda.gzda001 CLIPPED,
                         #".host=",lr_gzda.gzda002 CLIPPED  #211104-00032#1 mark
                         ".host=",g_ip CLIPPED  #211104-00032#1 add                         
                         
            CALL l_channel.writeLine(l_text)

            #topprd.dsdemo.port=1521
            #LET l_text = FGL_GETENV("ZONE") CLIPPED,".",lr_gzda.gzda001 CLIPPED,".port=1521"  #210823-00002#1 mark
            LET l_text = FGL_GETENV("ZONE") CLIPPED,".",lr_gzda.gzda001 CLIPPED,".port=",g_port USING "<<<<<"  #210823-00002#1 add
            CALL l_channel.writeLine(l_text)

            #topprd.dsdemo.user=dsdemo
            LET l_text = FGL_GETENV("ZONE") CLIPPED,".",lr_gzda.gzda001 CLIPPED,
                         ".user=",lr_gzda.gzda003 CLIPPED
            CALL l_channel.writeLine(l_text)

            #topprd.dsdemo.pass=<REDACTED>
            LET l_text = cl_hashkey_base65_anti(lr_gzda.gzda004) CLIPPED
            LET l_text = security.Base64.FromString(l_text)
            LET l_text = FGL_GETENV("ZONE") CLIPPED,".",lr_gzda.gzda001 CLIPPED,
                         ".pass=",l_text
            CALL l_channel.writeLine(l_text)

            #增加空白行分隔
            CALL l_channel.writeLine("")
         END IF
      END FOREACH                         

   ELSE
      #檔案不存在
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = 'adz-00339'
      LET g_errparam.extend = ""
      LET g_errparam.popup = TRUE
      LET g_errparam.replace[1] = l_file
      CALL cl_err()
       RETURN FALSE
   END IF

   #211220-00001#1 add (s)
   LET l_text = FGL_GETENV("ZONE")
   TRY 
     LET conn = JdbcConnUtil2.getConnection(l_text,g_dbs)
     LET stat = conn.createStatement()
     LET l_text = "Java 測試 \n",
                  "區域 : ",l_text,"\n",
                  "DB user : ",g_dbs,"\n",
                  "連線 ok \n",
                  "P.S. 如需測試其他 ENT 請在切換 ENT 後重產"
     MENU "提醒" ATTRIBUTE(STYLE="dialog",COMMENT=l_text,IMAGE="")
       ON ACTION ok
          EXIT MENU
      END MENU
   CATCH
       INITIALIZE g_errparam TO NULL
       #LET g_errparam.code = 'lib-00431'  #210120-00014#1 mark
       LET g_errparam.code = 'adz-00594'   #210120-00014#1 add
       LET g_errparam.extend = 'JdbcConn : ',sqlca.sqlerrm
       LET g_errparam.replace[1]='Java use ',l_file,"\n","請確認檔案內資料庫連線資訊"
       LET g_errparam.popup = TRUE
       DISPLAY "JdbcConn : ",sqlca.sqlerrm  #210715-00035#1 add
       CALL cl_err()

       RETURN FALSE
   END TRY
   #211220-00001#1 add (e)

   #211104-00032#1 add (s)
   LET l_text = "cat ",l_file," | wc -l "
   CALL l_channel.openPipe(l_text,"r")
   LET  l_num = l_channel.readLine()
   IF l_num < 2 THEN 
      LET l_text = "請檢查 azzi085 設定是否 正確 \n",
                   l_file,"\n ",
                   "檔案內容是異常"
      MENU "錯誤" ATTRIBUTE(STYLE="dialog",COMMENT=l_text,IMAGE="stop")
          ON ACTION ok
             EXIT MENU
      END MENU
      RETURN FALSE

   END IF
   #211104-00032#1 add (e)

   CALL l_channel.close()

#   #檔案權限775(7*64+7*8+5)
#   IF os.Path.chrwx(l_file,509) THEN END IF   #170525 #170503-00066#1   mark
   #檔案權限777(7*64+7*8+7)
   IF os.Path.chrwx(l_file,511) THEN END IF   #170525 #170503-00066#1   add
   DISPLAY "ini path:",l_file

   RETURN TRUE

END FUNCTION

 
{</section>}
 
