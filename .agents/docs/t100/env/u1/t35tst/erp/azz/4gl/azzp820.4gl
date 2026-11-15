#該程式未解開Section, 採用最新樣板產出!
#該程式非freestyle程式!
{<section id="azzp820.description" type="s" >}
#應用 a00 樣板自動產生(Version:3)
#+ Standard Version.....: SD版次:0015(2018-03-14 14:03:42), PR版次:0015(2024-03-18 15:23:20)
#+ Customerized Version.: SD版次:0000(1900-01-01 00:00:00), PR版次:0000(1900-01-01 00:00:00)
#+ Build......: 000643
#+ Filename...: azzp820
#+ Description: 逾期LOG資料批量移除作業
#+ Creator....: 01856(2016-03-29 10:22:59)
#+ Modifier...: 08146 -SD/PR- 01856
 
{</section>}
 
{<section id="azzp820.global" type="s" >}
#應用 p01 樣板自動產生(Version:25)
#add-point:填寫註解說明 name="global.memo" name="global.memo"
#171003-00012#1  By 08146 增加刪除gzpc_t 、gzpd_t 兩個表格
#180313-00029#1  By 08146 azzp820 優化 增加選擇刪除日期天數
#180312-00040#1  By 08146 刪除時，增加刪除 mongodb 資料
#180710-00028#1  bY 08146 刪除時，需檢查 mongodb 是否為目前的 db
#180807-00046#10 BY 00824 調整日期格式及TO_DATE用法
#200904-00093#1  By 01856 優化刪除log，依azzi500 ENT 有效且屬於ERP資料庫 進行刪除
#201209-00048#1  By 01856 優化程式當大量資料刪除時，避免第一次刪除造成系統卡住
#211104-00021#1  By 16615 新增通知中心表格 lokg_t 適用於T100 v1，T100 v2(含)以上不支援 
#221206-00022#1  By 01856 修正刪除wsfa_t
#221230-00039#1  By 01856 優化azzp820 刪除log(xx)_t日誌表格，並增加wsfa_t wsfb_t lsxm_t，利用SCC:385設定表格編號刪除日誌選項
#231221-00028#1  By 01856 增加wsgj_t，利用SCC:385設定表格編號刪除日誌選項，並修正當選擇 1=1，無添加任何查詢條件
#231228-00038#1  By 01856 客戶家的DSBLOB會爆掉，azzp820是用delete語法，故是無法釋放tablespace的空間  
#240213-00001#1  By 01856 無法有效釋放空間release DSBLOB 出現 SQLCA.SQLCODE: -6372 ORA-01031: insufficient privileges
#240318-00005#1  By 01856 修正取不到E-SYS-0735參數的問題
#end add-point
#add-point:填寫註解說明(客製用) name="global.memo_customerization"

#end add-point
#(ver:20) ---start---
#add-point:填寫註解說明(行業用) name="global.memo_industry"

#end add-point
#(ver:20) --- end ---
 
IMPORT os
IMPORT util
IMPORT FGL lib_cl_schedule
#add-point:增加匯入項目 name="global.import"

#end add-point
 
SCHEMA ds
 
GLOBALS "../../cfg/top_global.inc"
GLOBALS "../../cfg/top_schedule.inc"
GLOBALS
   DEFINE gwin_curr2  ui.Window
   DEFINE gfrm_curr2  ui.Form
   DEFINE gi_hiden_asign       LIKE type_t.num5
   DEFINE gi_hiden_exec        LIKE type_t.num5
   DEFINE gi_hiden_spec        LIKE type_t.num5
   DEFINE gi_hiden_exec_end    LIKE type_t.num5
   DEFINE g_chk_jobid          LIKE type_t.num5
END GLOBALS
 
PRIVATE TYPE type_parameter RECORD
   #add-point:自定背景執行須傳遞的參數(Module Variable) name="global.parameter"
   
   #end add-point
        wc               STRING
                     END RECORD
 
DEFINE g_sql             STRING        #組 sql 用
DEFINE g_forupd_sql      STRING        #SELECT ... FOR UPDATE  SQL
DEFINE g_error_show      LIKE type_t.num5
DEFINE g_jobid           STRING
DEFINE g_wc              STRING
 
PRIVATE TYPE type_master RECORD
       ld_date LIKE type_t.dat, 
   stagenow LIKE type_t.chr80,
       wc               STRING
       END RECORD
 
#模組變數(Module Variables)
DEFINE g_master type_master
DEFINE g_condition_display LIKE type_t.chr1   #180813-00044
 
 
#add-point:自定義模組變數(Module Variable) name="global.variable"
#180313-00029#1--START--
DEFINE g_master_1 DYNAMIC ARRAY OF RECORD
    ls_table      LIKE dzea_t.dzea001,
    li_day_cnt    LIKE type_t.num10
    END RECORD
DEFINE l_chk_delete BOOLEAN             
#180313-00029#1--END--
#200904-00093#1--start--
DEFINE ga_gzou       DYNAMIC ARRAY OF RECORD
          gzou003    LIKE gzou_t.gzou003
                     END RECORD
DEFINE g_dbs_old      STRING 
#200904-00093#1--end--
#201209-00048#1--start--
DEFINE g_table_list DYNAMIC ARRAY OF RECORD
              gzou003    LIKE gzou_t.gzou003,
              dzea001    LIKE dzea_t.dzea001,
              chk_flag   LIKE type_t.num5
                     END RECORD
#201209-00048#1--end--                     
#221230-00039#1 add(s)
DEFINE g_new_ver LIKE type_t.num5  #啟用新版刪除log機制  
#221230-00039#1 add(e)
#231228-00038#1 add(s)
DEFINE g_table_list2 DYNAMIC ARRAY OF RECORD
              dzea001    LIKE dzea_t.dzea001,
              chk_cnt    LIKE type_t.num5
                     END RECORD
DEFINE g_obj         util.JSONObject
#231228-00038#1 add(e)
#240318-00005#1 add(s)
DEFINE g_pram DYNAMIC ARRAY OF RECORD
              lc_value    LIKE ooaa_t.ooaa002
                     END RECORD
#240318-00005#1 add(e)                     
#end add-point
 
#add-point:自定義客戶專用模組變數(Module Variable) name="global.variable_customerization"

#end add-point
 
#add-point:傳入參數說明 name="global.argv"

#end add-point
 
{</section>}
 
{<section id="azzp820.main" type="s" >}
MAIN
   #add-point:main段define (客製用) name="main.define_customerization"
   
   #end add-point 
   DEFINE ls_js    STRING
   DEFINE lc_param type_parameter  
   #add-point:main段define name="main.define"
   
   #end add-point 
  
   #設定SQL錯誤記錄方式 (模組內定義有效)
   WHENEVER ERROR CALL cl_err_msg_log
 
   #add-point:初始化前定義 name="main.before_ap_init"
   
   #end add-point
   #依模組進行系統初始化設定(系統設定)
   CALL cl_ap_init("azz","")
 
   #add-point:定義背景狀態與整理進入需用參數ls_js name="main.background"
   DISPLAY "g_bgjob:",g_bgjob," g_argv[01] ",g_argv[01] ," g_argv[02]:",g_argv[02] ," g_argv[03]:",g_argv[03]
   #201209-00048#1--start--
   IF g_argv[01] = "init" THEN
      IF azzp820_chk_prog() THEN 
         LET g_bgjob = "Y"
      END IF        
   END IF 
   #201209-00048#1--end--
   
   #end add-point
 
   #背景(Y) 或半背景(T) 時不做主畫面開窗
   IF g_bgjob = "Y" OR g_bgjob = "T" THEN
      #排程參數由01開始，若不是1開始，表示有保留參數
      LET ls_js = g_argv[01]
     #CALL util.JSON.parse(ls_js,g_master)   #p類主要使用l_param,此處不解析
      #add-point:Service Call name="main.servicecall"
      #201209-00048#1#1--start--
      #判斷高頻設定是否存在 
      IF azzp820_chk_prog() THEN 
         #判斷ps -ef | grep azzp820 設定是否存在          
         IF azzp820_chk_exist() THEN 
            DISPLAY "azzp820目前正在執行 參數:" ,g_argv[01]
            CALL cl_ap_exitprogram("0")     
         ELSE 
           IF g_argv[01] = "init" THEN #表示高頻排程
           ELSE 
              DISPLAY "azzp820有高頻排程設定 :參數:" ,g_argv[01],"暫時不執行"
              CALL cl_ap_exitprogram("0")  
           END IF 
         END IF  
      ELSE  
      END IF
      #201209-00048#1#1--end--      
      #end add-point
      CALL azzp820_process(ls_js)
   ELSE
      #畫面開啟 (identifier)
      OPEN WINDOW w_azzp820 WITH FORM cl_ap_formpath("azz",g_code)
 
      #瀏覽頁簽資料初始化
      CALL cl_ui_init()
 
      #程式初始化
      CALL azzp820_init()
 
      #進入選單 Menu (="N")
      CALL azzp820_ui_dialog()
 
      #add-point:畫面關閉前 name="main.before_close"
      
      #end add-point
      #畫面關閉
      CLOSE WINDOW w_azzp820
   END IF
 
   #add-point:作業離開前 name="main.exit"
   
   #end add-point
 
   #離開作業
   CALL cl_ap_exitprogram("0")
END MAIN
 
{</section>}
 
{<section id="azzp820.init" type="s" >}
#+ 初始化作業
PRIVATE FUNCTION azzp820_init() RETURNS ()   #200213-00032
 
   #add-point:init段define (客製用) name="init.define_customerization"
   
   #end add-point
   #add-point:ui_dialog段define name="init.define"
   
   #end add-point
 
   LET g_error_show = 1
   LET gwin_curr2 = ui.Window.getCurrent()
   LET gfrm_curr2 = gwin_curr2.getForm()
   CALL cl_schedule_import_4fd()
   CALL cl_set_combo_scc("gzpa003","75")
   IF cl_get_para(g_enterprise,"","E-SYS-0005") = "N" THEN
       CALL cl_set_comp_visible("scheduling_page,history_page",FALSE)
   END IF
   LET g_condition_display = "N"   #180813-00044
 
   #190121-00002#1 - S -
   #帶起程式來源為azzi950時僅呈現條件查詢頁
   IF NOT cl_null(g_parentprog) AND g_parentprog = "azzi950"  THEN
       CALL cl_set_comp_visible("scheduling_page,history_page",FALSE)
   END IF      
   #190121-00002#1 - E -
 
   #add-point:畫面資料初始化 name="init.init"
   CALL cl_set_combo_scc('formonly.li_day_cnt','306')
   #end add-point
   
END FUNCTION
 
{</section>}
 
{<section id="azzp820.ui_dialog" type="s" >}
#+ 選單功能實際執行處
PRIVATE FUNCTION azzp820_ui_dialog() RETURNS ()   #200213-00032
 
   #add-point:ui_dialog段define (客製用) name="ui_dialog.define_customerization"
   
   #end add-point
   DEFINE li_exit  LIKE type_t.num5    #判別是否為離開作業
   DEFINE li_idx   LIKE type_t.num10
   DEFINE ls_js    STRING
   DEFINE ls_wc    STRING
   DEFINE l_dialog ui.DIALOG
   DEFINE lc_param type_parameter
   DEFINE lp_str   STRING   #201109-00009
   #add-point:ui_dialog段define name="ui_dialog.define"
   DEFINE ls_sql          STRING              #180313-00029#1 add
   DEFINE li_cnt          LIKE type_t.num10   #180313-00029#1 add
   DEFINE li_cnt2         LIKE type_t.num10   #221230-00039#1 add   
   #end add-point
   
   #add-point:ui_dialog段before dialog name="ui_dialog.before_dialog"
   
   #end add-point
 
   WHILE TRUE
      #add-point:ui_dialog段before dialog2 name="ui_dialog.before_dialog2"
      
      #end add-point
 
      DIALOG ATTRIBUTES(UNBUFFERED,FIELD ORDER FORM)
         
         
         #應用 a58 樣板自動產生(Version:3)
         CONSTRUCT BY NAME g_master.wc ON ld_date
            BEFORE CONSTRUCT
               #add-point:cs段before_construct name="cs.head.before_construct"
               
               #end add-point 
         
            #公用欄位開窗相關處理
            
               
            #一般欄位開窗相關處理    
                     #應用 a01 樣板自動產生(Version:2)
         BEFORE FIELD ld_date
            #add-point:BEFORE FIELD ld_date name="construct.b.ld_date"
            
            #END add-point
 
 
         #應用 a02 樣板自動產生(Version:4)
         AFTER FIELD ld_date
            
            #add-point:AFTER FIELD ld_date name="construct.a.ld_date"
            
            #END add-point
            
 
            
 
         #Ctrlp:construct.c.ld_date
         #應用 a03 樣板自動產生(Version:3)
#此ctrlp無內容         ON ACTION controlp INFIELD ld_date
            #add-point:ON ACTION controlp INFIELD ld_date name="construct.c.ld_date"
            
            #END add-point
 
 
 
            
            #add-point:其他管控 name="cs.other"
            
            #end add-point
            
         END CONSTRUCT
 
 
 
      
         #add-point:ui_dialog段construct name="ui_dialog.more_construct"
         
         #end add-point
         #add-point:ui_dialog段input name="ui_dialog.more_input"
         #180313-00029#1--start--
         INPUT ARRAY g_master_1 FROM s_detail1.*
            AFTER INPUT 
               FOR li_cnt = 1 TO g_master_1.getLength()
                  IF g_master_1[li_cnt].li_day_cnt <> 0 THEN
                     LET l_chk_delete =  TRUE
                     EXIT FOR
                  END IF 
               END FOR
         END INPUT
         #180313-00029#1--end--
         #end add-point
         #add-point:ui_dialog段自定義display array name="ui_dialog.more_displayarray"
         
         #end add-point
 
         SUBDIALOG lib_cl_schedule.cl_schedule_setting
         SUBDIALOG lib_cl_schedule.cl_schedule_setting_exec_call
         SUBDIALOG lib_cl_schedule.cl_schedule_select_show_history
         SUBDIALOG lib_cl_schedule.cl_schedule_show_history
 
         BEFORE DIALOG
            LET l_dialog = ui.DIALOG.getCurrent()
            CALL azzp820_get_buffer(l_dialog)
            #180813-00044 ---start---
            #顯示上次使用者下的條件
            IF g_condition_display = "Y" THEN
               CALL cl_qbe_display_condition('100001', g_user)
            END IF
            #180813-00044 --- end ---
            #add-point:ui_dialog段before dialog name="ui_dialog.before_dialog3"
            #180313-00029#1--start--
            LET li_cnt = 1
            #211104-00021#1 --- start --- 在顯示欄位限定 T100 v1.0 才會有lokg_t的表格
            IF FGL_GETENV("ERPVER") = '1.0' THEN
                LET ls_sql = "SELECT dzea001 FROM dzea_t",
                         " WHERE dzea001 LIKE 'log%'",
                         "   AND dzea001 <> 'logr_t'",
                         "    OR dzea001 = 'wsfa_t'",
                         "    OR dzea001 = 'lokg_t'"

            ELSE
               LET ls_sql = "SELECT dzea001 FROM dzea_t",
                         " WHERE dzea001 LIKE 'log%'",
                         "   AND dzea001 <> 'logr_t'",
                         "    OR dzea001 = 'wsfa_t'"
            END IF

            # LET ls_sql = "SELECT dzea001 FROM dzea_t",
            #             " WHERE dzea001 LIKE 'log%'",
            #             "   AND dzea001 <> 'logr_t'",
            #             "    OR dzea001 = 'wsfa_t'"
            #211104-00021#1 --- end ---
            #221230-00039#1 add(s)                 
                SELECT COUNT(1) INTO li_cnt2 FROM gzca_t
                  WHERE gzca001 = 385  
                IF li_cnt2 > 0 THEN 
                   LET ls_sql = "SELECT gzcb002 FROM  gzca_t 
                                  LEFT JOIN gzcb_t ON gzca001 = gzcb001
                                   WHERE  gzca001 = 385
                                    ORDER  BY  gzcb012 "
                END IF        
            #221230-00039#1 add(e)                            
            DECLARE azzp820_table_init_cs CURSOR FROM ls_sql
            FOREACH azzp820_table_init_cs INTO g_master_1[li_cnt].ls_table
               LET g_master_1[li_cnt].li_day_cnt = "0"
               LET li_cnt = li_cnt + 1
            END FOREACH
            CALL g_master_1.deleteElement(g_master_1.getLength())
            LET l_chk_delete = FALSE    #預設為FALSE，代表走刪除日期不走下方table
            #180313-00029#1--end--
            #end add-point
 
         ON ACTION batch_execute
            LET g_action_choice = "batch_execute"
            #180813-00044 ---start---
            #將此次查詢的條件儲存起來
            LET g_condition_display = "Y"
            CALL cl_qbe_auto_save()
            #180813-00044 --- end ---
            ACCEPT DIALOG
 
         #add-point:ui_dialog段before_qbeclear name="ui_dialog.before_qbeclear"
         
         #end add-point
 
         ON ACTION qbeclear         
            CLEAR FORM
            INITIALIZE g_master.* TO NULL   #畫面變數清空
            INITIALIZE lc_param.* TO NULL   #傳遞參數變數清空
            #add-point:ui_dialog段qbeclear name="ui_dialog.qbeclear"
            
            #end add-point
 
         ON ACTION history_fill
            CALL cl_schedule_history_fill()
 
         ON ACTION close
            LET INT_FLAG = TRUE
            EXIT DIALOG
         
         ON ACTION exit
            LET INT_FLAG = TRUE
            EXIT DIALOG
 
         #add-point:ui_dialog段action name="ui_dialog.more_action"
         
         #end add-point
 
         #主選單用ACTION
         &include "main_menu_exit_dialog.4gl"
         &include "relating_action.4gl"
         #交談指令共用ACTION
         &include "common_action.4gl"
            CONTINUE DIALOG
      END DIALOG
      
      IF g_action_choice = "logistics" THEN
         #清除畫面及相關資料
         CLEAR FORM   
         INITIALIZE g_master.* TO NULL
      #   LET g_wc  = ' 1=2' #250121-00005 mark
         LET g_wc  = ' 1=1'  #250121-00005 modify
         LET g_action_choice = ""
         CALL azzp820_init()
         CONTINUE WHILE
      END IF
 
      #檢查批次設定是否有錯(或未設定完成)
      IF cl_null(g_parentprog) THEN #190121-00002#1
         IF NOT cl_schedule_exec_check() THEN
            CONTINUE WHILE
         END IF
      END IF #190121-00002#1
      
      LET lc_param.wc = g_master.wc    #把畫面上的wc傳遞到參數變數
      #請在下方的add-point內進行把畫面的輸入資料(g_master)轉換到傳遞參數變數(lc_param)的動作
      #add-point:ui_dialog段exit dialog name="process.exit_dialog"
      #180313-00029#1--start--
      IF l_chk_delete THEN
         LET ls_js = util.JSON.stringify(g_master_1)
         LET lc_param.wc = ls_js
      END IF
      #180313-00029#1#1--end--      
      #end add-point
 
      LET ls_js = util.JSON.stringify(lc_param)  #r類使用g_master/p類使用lc_param
 
      #190121-00002#1 - S -
      #帶起程式來源為azzi950時僅回傳條件不做執行
      IF NOT cl_null(g_parentprog) AND g_parentprog = "azzi950"  THEN
         IF g_action_choice = "batch_execute" THEN
            #將條件寫入暫存檔並結束
            CALL cl_schedule_write_json_data(ls_js)
            EXIT WHILE
         END IF   
      END IF      
      #190121-00002#1 - E -
      
      IF INT_FLAG THEN
         LET INT_FLAG = FALSE
         EXIT WHILE
      ELSE
         IF g_chk_jobid THEN 
            LET g_jobid = g_schedule.gzpa001
         ELSE 
            LET g_jobid = cl_schedule_get_jobid(g_prog)
         END IF 
 
         #依照指定模式執行報表列印
         CASE 
            WHEN g_schedule.gzpa003 = "0"
                 CALL azzp820_process(ls_js)
 
            WHEN g_schedule.gzpa003 = "1"
                 LET ls_js = azzp820_transfer_argv(ls_js)
                 CALL cl_cmdrun(ls_js)
 
            WHEN g_schedule.gzpa003 = "2"
                 CALL cl_schedule_update_data(g_jobid,ls_js)
 
            WHEN g_schedule.gzpa003 = "3"
                 CALL cl_schedule_update_data(g_jobid,ls_js)
         END CASE  
 
         IF g_schedule.gzpa003 = "2" OR g_schedule.gzpa003 = "3" THEN 
            CALL cl_ask_confirm3("std-00014","") #設定完成
         END IF    
         LET g_schedule.gzpa003 = "0" #預設一開始 立即於前景執行
 
         #add-point:ui_dialog段after schedule name="process.after_schedule"
         IF g_bgjob = "N" THEN
            #當刪除完成 畫面清空
            CALL cl_progress_bar_no_window(0)
            CALL cl_progress_no_window_ing("")
         END IF 
         #end add-point
 
         #欄位初始資訊 
         CALL cl_schedule_init_info("all",g_schedule.gzpa003) 
         LET gi_hiden_asign = FALSE 
         LET gi_hiden_exec = FALSE 
         LET gi_hiden_spec = FALSE 
         LET gi_hiden_exec_end = FALSE 
         CALL cl_schedule_hidden()
      END IF
   END WHILE
 
END FUNCTION
 
{</section>}
 
{<section id="azzp820.transfer_argv" type="s" >}
#+ 轉換本地參數至cmdrun參數內,準備進入背景執行
PRIVATE FUNCTION azzp820_transfer_argv(ls_js) RETURNS (STRING)   #200213-00032
 
   #add-point:transfer_agrv段define (客製用) name="transfer_agrv.define_customerization"
   
   #end add-point
   DEFINE ls_js       STRING
   DEFINE la_cmdrun   RECORD
             prog       STRING,
             actionid   STRING,
             background LIKE type_t.chr1,
             param      DYNAMIC ARRAY OF STRING
                  END RECORD
   DEFINE la_param    type_parameter
   #add-point:transfer_agrv段define name="transfer_agrv.define"
   
   #end add-point
 
   LET la_cmdrun.prog = g_prog
   LET la_cmdrun.background = "Y"
   LET la_cmdrun.param[1] = ls_js
 
   #add-point:transfer.argv段程式內容 name="transfer.argv.define"
   
   #end add-point
 
   RETURN util.JSON.stringify( la_cmdrun )
END FUNCTION
 
{</section>}
 
{<section id="azzp820.process" type="s" >}
#+ 資料處理   (r類使用g_master為主處理/p類使用l_param為主)
PRIVATE FUNCTION azzp820_process(ls_js) RETURNS ()   #200213-00032
 
   #add-point:process段define (客製用) name="process.define_customerization"
   
   #end add-point
   DEFINE ls_js         STRING
   DEFINE lc_param      type_parameter
   DEFINE li_stus       LIKE type_t.num5
   DEFINE li_count      LIKE type_t.num10  #progressbar計量
   DEFINE ls_sql        STRING             #主SQL
   DEFINE li_p01_status LIKE type_t.num5
   #add-point:process段define name="process.define"
   
   #end add-point
 
  #INITIALIZE lc_param TO NULL           #p類不可以清空
   CALL util.JSON.parse(ls_js,lc_param)  #r類作業被t類呼叫時使用, p類主要解開參數處
   LET li_p01_status = 1
 
  #IF lc_param.wc IS NOT NULL THEN
  #   LET g_bgjob = "T"       #特殊情況,此為t類作業鬆耦合串入報表主程式使用
  #END IF
 
   #add-point:process段前處理 name="process.pre_process"
   CALL azzp820_get_gzou003()   #200904-00093#1
   #221230-00039#1 add(s)
   SELECT COUNT(1) INTO li_count FROM gzca_t
       WHERE  gzca001 = 385  
   IF li_count > 0 THEN 
       LET g_new_ver = TRUE 
   END IF 
   #221230-00039#1 add(e)
   #231228-00038#1 add(s)   
   LET g_obj = util.JSONObject.create()  
   CALL g_table_list2.clear()
   #231228-00038#1 add(e)   
   #end add-point
 
   #預先計算progressbar迴圈次數
   IF g_bgjob <> "Y" THEN
      #add-point:process段count_progress name="process.count_progress"
      
      #end add-point
   END IF
 
   #主SQL及相關FOREACH前置處理
#  DECLARE azzp820_process_cs CURSOR FROM ls_sql
#  FOREACH azzp820_process_cs INTO
   #add-point:process段process name="process.process"
   
   #end add-point
#  END FOREACH
 
   IF g_bgjob = "N" THEN
      #前景作業完成處理
      #add-point:process段foreground完成處理 name="process.foreground_finish"
      #CALL cl_progress_bar_no_window(12)  #221230-00039#1 mark
      #221230-00039#1 add(s)
      IF g_new_ver THEN 
         CALL cl_progress_bar_no_window(g_master_1.getLength())
      ELSE 
         CALL cl_progress_bar_no_window(12)
      END IF 
      #221230-00039#1 add(e)
      #開始刪除 
      CALL azzp820_delete_log(ls_js)
      #end add-point
      CALL cl_ask_confirm3("std-00012","")
   ELSE
      #背景作業完成處理
      #add-point:process段background完成處理 name="process.background_finish"
      CALL azzp820_delete_log(ls_js)
      #end add-point
      CALL cl_schedule_exec_call(li_p01_status)
   END IF
 
   #呼叫訊息中心傳遞本關完成訊息
   CALL azzp820_msgcentre_notify()
 
END FUNCTION
 
{</section>}
 
{<section id="azzp820.get_buffer" type="s" >}
PRIVATE FUNCTION azzp820_get_buffer(p_dialog) RETURNS ()   #200213-00032
 
   #add-point:process段define (客製用) name="get_buffer.define_customerization"
   
   #end add-point
   DEFINE p_dialog   ui.DIALOG
   #add-point:process段define name="get_buffer.define"
   
   #end add-point
 
   
 
   CALL cl_schedule_get_buffer(p_dialog)
 
   #add-point:get_buffer段其他欄位處理 name="get_buffer.others"
   
   #end add-point
END FUNCTION
 
{</section>}
 
{<section id="azzp820.msgcentre_notify" type="s" >}
PRIVATE FUNCTION azzp820_msgcentre_notify() RETURNS ()   #200213-00032
 
   #add-point:process段define (客製用) name="msgcentre_notify.define_customerization"
   
   #end add-point
   DEFINE lc_state LIKE type_t.chr5
   #add-point:process段define name="msgcentre_notify.define"
   
   #end add-point
 
   INITIALIZE g_msgparam TO NULL
 
   #action-id與狀態填寫
   LET g_msgparam.state = "process"
 
   #add-point:msgcentre其他通知 name="msg_centre.process"
   
   #end add-point
 
   #呼叫訊息中心傳遞本關完成訊息
   CALL cl_msgcentre_notify()
 
END FUNCTION
 
{</section>}
 
{<section id="azzp820.other_function" readonly="Y" type="s" >}
#add-point:自定義元件(Function) name="other.function"

################################################################################
# Descriptions...: 刪除log
# Memo...........:
# Usage..........: CALL azzp820_delete_log(ls_js)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2016/03/29 By jrg542
# Modify.........:
################################################################################
PRIVATE FUNCTION azzp820_delete_log(ls_js)
   DEFINE ls_js         STRING
   DEFINE pc_date       LIKE type_t.chr80 
   DEFINE lc_param      type_parameter
   DEFINE ld_dist_date  DATETIME YEAR TO SECOND   
   DEFINE ls_date       STRING 
   DEFINE li_idx        LIKE type_t.num5
   DEFINE li_ok         LIKE type_t.num5
   DEFINE lt_start      DATETIME YEAR TO FRACTION(3)
   DEFINE lt_end        DATETIME YEAR TO FRACTION(3)
   DEFINE ls_js_array   util.JSONArray
   DEFINE li_cnt        LIKE type_t.num5    #200904-00093#1   
   DEFINE li_chk        LIKE type_t.num5
   DEFINE li_cnt2       LIKE type_t.num5
   DEFINE days          LIKE type_t.num5
   #221206-00022#1 add(s) 2022/12/07   
   DEFINE ls_begin       STRING             
   DEFINE lc_time        LIKE type_t.chr50  
   DEFINE ld_day         DATETIME YEAR TO SECOND    
   #221206-00022#1 add(e) 2022/12/07   
   #221230-00039#1 add(s)
   DEFINE lc_table_rec  RECORD 
       gzcb002 LIKE gzcb_t.gzcb002,
       gzcb003 LIKE gzcb_t.gzcb003,
       gzcb004 LIKE gzcb_t.gzcb004,
       gzcb005 LIKE gzcb_t.gzcb005
       END RECORD  
   DEFINE lc_dzeb007 LIKE  dzeb_t.dzeb007 
   DEFINE ld_dist_date2  DATETIME YEAR TO SECOND
   DEFINE li_cnt3       LIKE type_t.num5
   #221230-00039#1 add(e)


   #201209-00048#1--start--
   #確認是否為JSON方式傳入
   IF ls_js.subString(1,1) = "{" THEN
      CALL util.JSON.parse(ls_js,lc_param)   
      LET ls_date = lc_param.wc
   ELSE    
   END IF 
   #CALL util.JSON.parse(ls_js,lc_param)   
   #LET ls_date = lc_param.wc
   #201209-00048#1--end--   
   
   #180313-00029#1--START--
   #檢查傳入的參數中是否有 ls_table，有的話代表是傳入 s_detail1 的資料。就跳至 azzp820_delete_s_setail1
   IF ls_date.getIndexOf("ls_table",1) > 0 THEN
      LET ls_js_array = util.JSONArray.parse(ls_date)    #傳入的 json格式的string parse為 json array
      CALL ls_js_array.toFGL(g_master_1)                 #將 json array 轉換成 record
      #CALL azzp820_delete_s_detail1()   #221230-00039#1 mark
      #221230-00039#1 add(s)    
      #CALL azzp820_delete_s_detail1()
      DISPLAY " g_new_ver ",g_new_ver
      IF g_new_ver THEN 
         #取當天日期當基礎點
         CALL azzp820_delete_s_detail1_V2(cl_get_today())  
      ELSE
         CALL azzp820_delete_s_detail1()
      END IF   
      #221230-00039#1 add(e)          
      RETURN 
   END IF 
   #180313-00029#1--END--
   
   #231221-00028#1 add(s)
   LET ls_date = ls_date.trim()
   IF ls_date = "1=1" THEN 
       LET ls_date = "'",cl_get_current(),"'" 
   END IF
   #231221-00028#1 add(e)     
   
   IF g_bgjob = "Y" THEN
      DISPLAY " ls_js ",ls_js
      DISPLAY " ls_date ",ls_date
#221230-00039#1 add(s)  
      DISPLAY " g_new_ver ",g_new_ver       
      IF g_new_ver THEN      
         IF ls_js = "init" THEN          
            IF cl_null(g_argv[03]) THEN 
               LET g_argv[03] = 100000 
            END IF 
         
            IF cl_null(g_argv[04])  THEN 
               LET days = 30
            ELSE
               LET days = g_argv[04]    
            END IF 
      
            CALL azzp820_get_table_list(days)
            LET g_dbs_old = g_dbs
            FOR li_cnt = 1 TO ga_gzou.getLength()
                DISPLAY " 資料庫編號:",ga_gzou[li_cnt].gzou003 
                IF g_dbs_old = ga_gzou[li_cnt].gzou003 THEN 
                ELSE 
                   CALL cl_db_connect(ga_gzou[li_cnt].gzou003,FALSE)  
                END IF
                CALL azzp820_get_para(g_enterprise,"","E-SYS-0735",ga_gzou[li_cnt].gzou003) #240318-00005#1 add
                #確認每個表格資料筆數，小於設定的參數上限就不刪除
                FOR li_cnt2 = 1 TO g_master_1.getLength()
                    CALL azzp820_get_table_info(g_master_1[li_cnt2].ls_table)  RETURNING lc_table_rec.*
                    CALL azzp820_chk_table_amount(lc_table_rec.gzcb002,ga_gzou[li_cnt].gzou003) RETURNING li_chk   
                    LET g_table_list[g_table_list.getLength()+1].gzou003 = ga_gzou[li_cnt].gzou003
                    LET g_table_list[g_table_list.getLength()].dzea001 = lc_table_rec.gzcb002
                    LET g_table_list[g_table_list.getLength()].chk_flag = li_chk 
                    DISPLAY "資料是否大於 ",g_argv[03]," (符合 1:不符合 0):",li_chk," table id:",lc_table_rec.gzcb002,  " DB:",ga_gzou[li_cnt].gzou003                    
                    IF li_chk THEN 
                       LET ld_dist_date = cl_get_current()
                       LET ld_dist_date = ld_dist_date - (g_master_1[li_cnt2].li_day_cnt UNITS DAY) #保留天數
                       CALL azzp820_get_field_type(lc_table_rec.gzcb002,lc_table_rec.gzcb003) RETURNING lc_dzeb007  
                       CALL azzp820_delete_log_use_count_V2(lc_table_rec.gzcb002,lc_table_rec.gzcb003,ld_dist_date,g_argv[02],lc_dzeb007)        
                       IF NOT cl_null(lc_table_rec.gzcb004) THEN 
                          CALL azzp820_get_field_type(lc_table_rec.gzcb002,lc_table_rec.gzcb004) RETURNING lc_dzeb007  
                          CALL azzp820_delete_log_use_count_V2(lc_table_rec.gzcb002,lc_table_rec.gzcb004,ld_dist_date,g_argv[02],lc_dzeb007)
                       END IF                    

                    END IF   
                END FOR 
                #231228-00038#1 add(s)
                #IF ga_gzou[li_cnt].gzou003 = "ds" THEN  #240213-00001#1 mark
                   CALL azzp820_release_blob_tablespcace(ga_gzou[li_cnt].gzou003)  #240213-00001#1 add a_gzou[li_cnt].gzou003
                #END IF                                  #240213-00001#1 mark
                #231228-00038#1 add(e)                 
                  IF g_dbs_old = ga_gzou[li_cnt].gzou003 THEN 
                  ELSE
                    #斷開目前連線   
                    CALL cl_db_disconnect_current()
                    #連到一開始程式執行時的對應連線 
                    IF cl_db_setconnect(g_dbs_old) THEN END IF
                 END IF
	        
            END FOR # ga_gzou    
             LET li_cnt2 = 0
             FOR li_cnt = 1 TO g_table_list.getLength()                 
                 IF NOT g_table_list[li_cnt].chk_flag THEN 
                    LET li_cnt2 = li_cnt2 + 1             
                 END IF 
             END FOR
             DISPLAY " 確認是否符合小於:",g_argv[03] ," li_cnt2:",li_cnt2 ," g_table_list.getLength():",g_table_list.getLength()
             #判斷所有資料是否小於 10000，就不可以不啟用高頻排程，採用一般排程刪除
             IF li_cnt2 = g_table_list.getLength() THEN             
                UPDATE gzpg_t SET gzpgstus = 'N'
                 WHERE gzpg002 = 'azzp820'
             END IF        
          
            RETURN
          END IF   
      END IF      
#221230-00039#1 add(e)      
      
     #201209-00048#1--start--
	  IF ls_js = "init" THEN 
         IF cl_null(g_argv[03]) THEN 
            LET g_argv[03] = 100000 
         END IF 
         
         IF cl_null(g_argv[04])  THEN 
            LET days = 30
         ELSE
            LET days = g_argv[04]    
         END IF 
         LET ld_dist_date = cl_get_current()
         LET ld_dist_date = ld_dist_date - (days UNITS DAY) #保留天數
DISPLAY "保留天數 days: ",days                
         LET g_dbs_old = g_dbs
         FOR li_cnt = 1 TO ga_gzou.getLength()
         DISPLAY " 資料庫編號:",ga_gzou[li_cnt].gzou003 
         IF g_dbs_old = ga_gzou[li_cnt].gzou003 THEN 
         ELSE 
            CALL cl_db_connect(ga_gzou[li_cnt].gzou003,FALSE)  
         END IF
	      CALL azzp820_get_para(g_enterprise,"","E-SYS-0735",ga_gzou[li_cnt].gzou003) #240318-00005#1 add  
           CALL azzp820_delete_log_use_count("loga_t","loga006",ld_dist_date,g_argv[02])
           CALL azzp820_delete_log_use_count("loga_t","loga007",ld_dist_date,g_argv[02])
           CALL azzp820_delete_log_use_count("logb_t","logb007",ld_dist_date,g_argv[02])
           CALL azzp820_delete_log_use_count("logc_t","logc003",ld_dist_date,g_argv[02])
           CALL azzp820_delete_log_use_count("logd_t","logd006",ld_dist_date,g_argv[02])
           CALL azzp820_chk_table_amount("logd_t",ga_gzou[li_cnt].gzou003) RETURNING li_chk
           DISPLAY "資料是否大於 ",g_argv[03]," 符合:",li_chk, " table id: logd_t",  " DB:",ga_gzou[li_cnt].gzou003
           LET g_table_list[g_table_list.getLength()+1].gzou003 = ga_gzou[li_cnt].gzou003
           LET g_table_list[g_table_list.getLength()].dzea001 = "logd_t"
           LET g_table_list[g_table_list.getLength()].chk_flag = li_chk

           CALL azzp820_delete_log_use_count("logd_t","logd007",ld_dist_date,g_argv[02])
           CALL azzp820_delete_log_use_count("loge_t","loge006",ld_dist_date,g_argv[02])
           CALL azzp820_delete_log_use_count("logi_t","logi002",ld_dist_date,g_argv[02])
           CALL azzp820_delete_log_use_count("logs_t","logs006",ld_dist_date,g_argv[02])
           CALL azzp820_delete_log_use_count("logt_t","logt002",ld_dist_date,g_argv[02])
           CALL azzp820_delete_log_use_count("logv_t","logv002",ld_dist_date,g_argv[02])
           CALL azzp820_delete_log_use_count("logw_t","logw003",ld_dist_date,g_argv[02])
           CALL azzp820_delete_log_use_count("logx_t","logx001",ld_dist_date,g_argv[02])
           CALL azzp820_delete_log_use_count("logz_t","logz001",ld_dist_date,g_argv[02])
           CALL azzp820_delete_log_use_count("gzpc_t","gzpc002",ld_dist_date,g_argv[02])
           CALL azzp820_delete_log_use_count("gzpd_t","gzpd008",ld_dist_date,g_argv[02])
           #211104-00021 ---start----
           IF FGL_GETENV("ERPVER") = '1.0' THEN
              CALL azzp820_delete_log_use_count("lokg_t","lokg015",ld_dist_date,g_argv[02])    #211104-00021  
           END IF
           #211104-00021 ---END----
           CALL azzp820_delete_log_use_count("wsfa_t","wsfa003",ld_dist_date,g_argv[02])    
           #231228-00038#1 add(s)
           CALL azzp820_delete_log_use_count("wsfb_t","wsfb003",ld_dist_date,g_argv[02]) 
           #IF ga_gzou[li_cnt].gzou003 = "ds" THEN  #240213-00001#1 mark
              CALL azzp820_release_blob_tablespcace(ga_gzou[li_cnt].gzou003) #240213-00001#1 add a_gzou[li_cnt].gzou003
           #END IF                                  #240213-00001#1 mark 
           #231228-00038#1 add(e)
           IF g_dbs_old = ga_gzou[li_cnt].gzou003 THEN 
           ELSE
             #斷開目前連線   
             CALL cl_db_disconnect_current()
             #連到一開始程式執行時的對應連線 
             IF cl_db_setconnect(g_dbs_old) THEN END IF
          END IF
	     
          END FOR     
          LET li_cnt2 = 0
          FOR li_cnt = 1 TO g_table_list.getLength()
              
              IF NOT g_table_list[li_cnt].chk_flag THEN 
                 LET li_cnt2 = li_cnt2 + 1 

              END IF 
          END FOR
          DISPLAY " 確認是否符合小於",g_argv[03] ,"li_cnt2:",li_cnt2 ," g_table_list.getLength():",g_table_list.getLength()
          #判斷所有資料是否小於 10000，就不可以不啟用高頻排程，採用一般排程刪除
          IF li_cnt2 = g_table_list.getLength() THEN 

             UPDATE gzpg_t SET gzpgstus = 'N'
              WHERE gzpg002 = 'azzp820'
          END IF        
          RETURN 
      END IF 
      #201209-00048#1--end--      

   END IF 
   LET li_idx = ls_date.getIndexOf("'",1)
   LET ls_date = ls_date.subString(li_idx+1,ls_date.getIndexOf("'",li_idx+1)-1)
   LET ld_dist_date = ls_date
#   LET ld_dist_date = ld_dist_date - (90 UNITS DAY)     #180313-00029#1 mark  
   LET ld_dist_date = ld_dist_date - (30 UNITS DAY)      #180313-00029#1 add  改為30天

#221230-00039#1 add(s)       
     IF g_new_ver THEN 
        #取開窗日期當基礎點或啟動排程當天當基礎點減30天 
        LET ld_dist_date2 = ls_date
        CALL azzp820_get_table_list(30)
        CALL azzp820_delete_s_detail1_V2(ld_dist_date2)                 
        RETURN  
     END IF  
#221230-00039#1 add(e) 

   #200904-00093#1--start--
   LET g_dbs_old = g_dbs
   FOR li_cnt = 1 TO ga_gzou.getLength()
      DISPLAY " 資料庫編號:",ga_gzou[li_cnt].gzou003 
      IF g_dbs_old = ga_gzou[li_cnt].gzou003 THEN 
      ELSE 
         CALL cl_db_connect(ga_gzou[li_cnt].gzou003,FALSE)  
      END IF
   #200904-00093#1--end--   
  CALL azzp820_get_para(g_enterprise,"","E-SYS-0735",ga_gzou[li_cnt].gzou003) #240318-00005#1 add

  DISPLAY "azzp820_delete_log 目標日期:",ls_date," 刪一個月前的資料:", ld_dist_date 
   LET lt_start =  cl_get_current()
   IF g_bgjob = "N" THEN 
      CALL cl_progress_no_window_ing("DELETE loga_t")
   END IF
   DELETE FROM loga_t WHERE loga006 < ld_dist_date
   LET lt_end =  cl_get_current()
   DISPLAY "delete loga_t WHERE loga006 count: ",sqlca.SQLERRD[3] ," spent time:",lt_end - lt_start 
   
   #200904-00093#1--start--
   DELETE FROM loga_t WHERE loga007 < ld_dist_date 
   LET lt_end =  cl_get_current() 
   DISPLAY "delete loga_t WHERE loga007 count: ",sqlca.SQLERRD[3] ," spent time:",lt_end - lt_start 
   #200904-00093#1--end--
   
   CALL azzp820_delete_data_mongo("loga_t", ld_dist_date)   #180312-00040#1 add
   #LET lt_end =  cl_get_current()                                                        #200904-00093#1 #2020/08/30
   #DISPLAY "delete loga_t count: ",sqlca.SQLERRD[3] ," spent time:",lt_end - lt_start    #200904-00093#1 #2020/08/30
   
   LET lt_start =  cl_get_current()
   IF g_bgjob = "N" THEN 
      CALL cl_progress_no_window_ing("DELETE logb_t")
   END IF 
   DELETE  FROM logb_t where logb007 < ld_dist_date
   LET lt_end =  cl_get_current()
   DISPLAY "delete logb_t where logb007 count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start

   
   CALL azzp820_delete_data_mongo("logb_t", ld_dist_date)   #180312-00040#1 add
   #LET lt_end =  cl_get_current()                                                     #200904-00093#1 #2020/08/30
   #DISPLAY "delete logb_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start  #200904-00093#1 #2020/08/30

   LET lt_start =  cl_get_current()
   IF g_bgjob = "N" THEN 
      CALL cl_progress_no_window_ing("DELETE logc_t")
   END IF    
   DELETE FROM logc_t WHERE logc003 < ld_dist_date
   LET lt_end =  cl_get_current()
   DISPLAY "delete logc_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start   
   CALL azzp820_delete_data_mongo("logc_t", ld_dist_date)   #180312-00040#1 add
   #LET lt_end =  cl_get_current()                                                     #200904-00093#1 #2020/08/30 
   #DISPLAY "delete logc_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start  #200904-00093#1 #2020/08/30

   LET lt_start =  cl_get_current()
   IF g_bgjob = "N" THEN 
      CALL cl_progress_no_window_ing("DELETE logd_t")
   END IF 
   DELETE FROM logd_t WHERE logd006 < ld_dist_date
   LET lt_end =  cl_get_current()
   DISPLAY "delete logd_t where logd006 count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start

   DELETE FROM logd_t WHERE logd007 < ld_dist_date
   LET lt_end =  cl_get_current()
   DISPLAY "delete logd_t where logd007 count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start
   CALL azzp820_delete_data_mongo("logd_t", ld_dist_date)   #180312-00040#1 add
   #LET lt_end =  cl_get_current()                                                     #200904-00093#1 #2020/08/30
   #DISPLAY "delete logd_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start  #200904-00093#1 #2020/08/30
 
   LET lt_start =  cl_get_current()
   IF g_bgjob = "N" THEN 
      CALL cl_progress_no_window_ing("DELETE loge_t")
   END IF 
   DELETE FROM loge_t WHERE loge006 < ld_dist_date
   
   LET lt_end =  cl_get_current()
   DISPLAY "delete loge_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start   
   CALL azzp820_delete_data_mongo("loge_t", ld_dist_date)   #180312-00040#1 add
   #LET lt_end =  cl_get_current()                                                    #200904-00093#1 #2020/08/30
   #DISPLAY "delete loge_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start #200904-00093#1 #2020/08/30
    #231228-00038#1 add(s)
    IF sqlca.SQLERRD[3] > 0 THEN 
       IF NOT g_obj.has("loge_t") THEN 
          CALL g_obj.put("loge_t" , "loge_t" )
          LET g_table_list2[g_table_list2.getLength()+1].chk_cnt = TRUE
          LET g_table_list2[g_table_list2.getLength()].dzea001 = "loge_t"   
       ELSE
       END IF
    END IF
    #231228-00038#1 add(e)  
   LET lt_start =  cl_get_current()
   IF g_bgjob = "N" THEN 
      CALL cl_progress_no_window_ing("DELETE logi_t")
   END IF    
   DELETE FROM logi_t WHERE logi002 < ld_dist_date
   LET lt_end =  cl_get_current()
   DISPLAY "delete logi_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start
   CALL azzp820_delete_data_mongo("logi_t", ld_dist_date)   #180312-00040#1 add
   #LET lt_end =  cl_get_current()                                                     #200904-00093#1 #2020/08/30
   #DISPLAY "delete logi_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start  #200904-00093#1 #2020/08/30

   LET lt_start =  cl_get_current()
   IF g_bgjob = "N" THEN 
      CALL cl_progress_no_window_ing("DELETE logs_t")
   END IF
   DELETE FROM logs_t WHERE logs006 < ld_dist_date
   LET lt_end =  cl_get_current()
   DISPLAY "delete logs_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start   
   CALL azzp820_delete_data_mongo("logs_t", ld_dist_date)   #180312-00040#1 add
   #LET lt_end =  cl_get_current()                                                      #200904-00093#1 #2020/08/30
   #DISPLAY "delete logs_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start   #200904-00093#1 #2020/08/30

   LET lt_start =  cl_get_current()
   IF g_bgjob = "N" THEN 
      CALL cl_progress_no_window_ing("DELETE logt_t")
   END IF
   DELETE FROM logt_t WHERE logt002 < ld_dist_date
   LET lt_end =  cl_get_current()   
   DISPLAY "delete logt_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start 
   CALL azzp820_delete_data_mongo("logt_t", ld_dist_date)   #180312-00040#1 add
   #LET lt_end =  cl_get_current()                                                      #200904-00093#1 #2020/08/30
   #DISPLAY "delete logt_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_star t  #200904-00093#1 #2020/08/30

   LET lt_start =  cl_get_current()
   IF g_bgjob = "N" THEN 
      CALL cl_progress_no_window_ing("DELETE logv_t")
   END IF
   DELETE FROM logv_t WHERE logv002 < ld_dist_date
   LET lt_end =  cl_get_current()      
   DISPLAY "delete logv_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start
   CALL azzp820_delete_data_mongo("logv_t", ld_dist_date)   #180312-00040#1 add
   #LET lt_end =  cl_get_current()                                                     #200904-00093#1 #2020/08/30
   #DISPLAY "delete logv_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start  #200904-00093#1 #2020/08/30

   LET lt_start =  cl_get_current()
   IF g_bgjob = "N" THEN 
      CALL cl_progress_no_window_ing("DELETE logw_t")
   END IF    
   DELETE FROM logw_t WHERE logw003 < ld_dist_date
   LET lt_end =  cl_get_current() 
   DISPLAY "delete logw_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start
   CALL azzp820_delete_data_mongo("logw_t", ld_dist_date)   #180312-00040#1 add
   #LET lt_end =  cl_get_current()                                                     #200904-00093#1 #2020/08/30
   #DISPLAY "delete logw_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start  #200904-00093#1 #2020/08/30

   LET lt_start =  cl_get_current()
   IF g_bgjob = "N" THEN 
      CALL cl_progress_no_window_ing("DELETE logx_t")
   END IF 
   DELETE FROM logx_t WHERE logx001 < ld_dist_date
   LET lt_end =  cl_get_current()  
   DISPLAY "delete logx_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start
   CALL azzp820_delete_data_mongo("logx_t", ld_dist_date)   #180312-00040#1 add
   #LET lt_end =  cl_get_current()                                                      #200904-00093#1 #2020/08/30
   #DISPLAY "delete logx_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start   #200904-00093#1 #2020/08/30

   LET lt_start =  cl_get_current()
   IF g_bgjob = "N" THEN 
      CALL cl_progress_no_window_ing("DELETE logz_t")
   END IF 
   DELETE FROM logz_t WHERE logz001 < ld_dist_date
   LET lt_end =  cl_get_current() 
   DISPLAY "delete logz_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start
   CALL azzp820_delete_data_mongo("logz_t", ld_dist_date)   #180312-00040#1 add
   #LET lt_end =  cl_get_current()                                                     #200904-00093#1 #2020/08/30
   #DISPLAY "delete logz_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start  #200904-00093#1 #2020/08/30
   
   #171003-00012#1--start--
   LET lt_start =  cl_get_current()
   IF g_bgjob = "N" THEN
      CALL cl_progress_no_window_ing("DELETE gzpc_t")
   END IF
   DELETE FROM gzpc_t WHERE gzpc002 < ld_dist_date
   #DISPLAY "delete gzpc_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start  #200904-00093#1 #2020/08/30
   LET lt_end =  cl_get_current()
   DISPLAY "delete gzpc_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start

   LET lt_start =  cl_get_current()
   IF g_bgjob = "N" THEN
      CALL cl_progress_no_window_ing("DELETE gzpd_t")
   END IF
   DELETE FROM gzpd_t WHERE gzpd008 < ld_dist_date
   LET lt_end =  cl_get_current()   
   DISPLAY "delete gzpd_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start
   #LET lt_end =  cl_get_current()  #200904-00093#1                                                    #2020/08/30
   #DISPLAY "delete gzpd_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start  #200904-00093#1 #2020/08/30
   #171003-00012#1--end--
   
   
   #211104-00021#1--start--
   IF FGL_GETENV("ERPVER") = '1.0' THEN
      LET lt_start = cl_get_current()
      IF g_bgjob = "N" THEN
         CALL cl_progress_no_window_ing("DELETE lokg_t")
      END IF
      DELETE FROM lokg_t WHERE　lokg015 < ld_dist_date
      LET lt_end = cl_get_current()
      DISPLAY "delete lokg_t coun: ",sqlca.SQLERRD[3]," spent time",lt_end = lt_start
      IF sqlca.SQLERRD[3] > 0 THEN 
         IF NOT g_obj.has("lokg_t") THEN 
            CALL g_obj.put("lokg_t" , "lokg_t")
            LET g_table_list2[g_table_list2.getLength()+1].chk_cnt = TRUE
            LET g_table_list2[g_table_list2.getLength()].dzea001 = "lokg_t"  
         ELSE
         END IF
      END IF             
   END IF
   #211104-00021#1--end--
   
   
   #180313-00029#1 --start--
   LET lt_start =  cl_get_current()
   IF g_bgjob = "N" THEN
      CALL cl_progress_no_window_ing("DELETE wsfa_t")
   END IF
   #221206-00022#1 add(s)
   LET ls_begin = ld_dist_date 
   LET lc_time = ls_begin
   DISPLAY "wsfa_t ls_begin:",ls_begin 
   DISPLAY "wsfa_t lc_time:",lc_time
   #DELETE FROM wsfa_t WHERE wsfa003 < ld_dist_date
   DELETE FROM wsfa_t WHERE wsfa003 < lc_time
   #221206-00022#1 add(e)
   LET lt_end =  cl_get_current()
   DISPLAY "delete wsfa_t count: ",sqlca.SQLERRD[3]," spent time:",lt_end - lt_start
   #180313-00029#1 --end--
    IF sqlca.SQLERRD[3] > 0 THEN 
       IF NOT g_obj.has("wsfa_t") THEN 
          CALL g_obj.put("wsfa_t"   , "wsfa_t"   )
          LET g_table_list2[g_table_list2.getLength()+1].chk_cnt = TRUE
          LET g_table_list2[g_table_list2.getLength()].dzea001 = "wsfa_t"  
       ELSE
       END IF
    END IF       
   IF g_bgjob = "N" THEN 
      CALL cl_progress_no_window_ing("DELETE Finish")
   END IF
   #221206-00022#1 add(s)
   #IF ga_gzou[li_cnt].gzou003 = "ds" THEN #240213-00001#1 mark
      CALL  azzp820_release_blob_tablespcace(ga_gzou[li_cnt].gzou003) #240213-00001#1 add a_gzou[li_cnt].gzou003
   #END IF                                 #240213-00001#1 mark    
   #221206-00022#1 add(e)   
   #200904-00093#1--start-- 2020/08/30--start--
   IF g_dbs_old = ga_gzou[li_cnt].gzou003 THEN 
   ELSE
      #斷開目前連線   
      CALL cl_db_disconnect_current()
      #連到一開始程式執行時的對應連線 
      IF cl_db_setconnect(g_dbs_old) THEN END IF
   END IF 
   #200904-00093#1--end--
   
   END FOR  #200904-00093#1 #2020/08/30--end--
END FUNCTION

################################################################################
# Descriptions...: 針對各個table做刪除的動作
# Memo...........:
# Usage..........: CALL azzp820_delete_s_detail1()
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: #180313-00029#1 by 08146
# Modify.........:
################################################################################
PRIVATE FUNCTION azzp820_delete_s_detail1()
   DEFINE ld_day         DATETIME YEAR TO SECOND 
   DEFINE li_cnt         LIKE type_t.num10
   DEFINE li_cnt2        LIKE type_t.num5   #200904-00093#1
   DEFINE ls_begin       STRING             #200904-00093#1
   DEFINE lc_time        LIKE type_t.chr50  #200904-00093#1

   #200904-00093#1--start--
   LET g_dbs_old = g_dbs   
   FOR li_cnt2 = 1 TO ga_gzou.getLength()
      IF g_dbs_old = ga_gzou[li_cnt2].gzou003 THEN 
      ELSE 
         CALL cl_db_connect(ga_gzou[li_cnt2].gzou003,FALSE)  
      END IF   
   #200904-00093#1--end--    
   CALL azzp820_get_para(g_enterprise,"","E-SYS-0735",ga_gzou[li_cnt2].gzou003)  #240318-00005#1 add
   FOR li_cnt = 1 TO g_master_1.getLength()
      IF g_master_1[li_cnt].li_day_cnt <> 0 THEN
#         LET ld_day = cl_get_today() - (g_master_1[li_cnt].li_day_cnt UNITS DAY) + (90 UNITS DAY)   #180312-00040#1 mark
         LET ld_day = cl_get_today() - (g_master_1[li_cnt].li_day_cnt UNITS DAY)                     #180312-00040#1 add  
         CASE g_master_1[li_cnt].ls_table 
            WHEN "loga_t"
               DISPLAY "loga_t ld_day:",ld_day
            #   DELETE FROM loga_t WHERE loga006 < TO_DATE(ld_day, 'YY/MM/DD')   #180807-00046 mark
               DELETE FROM loga_t WHERE loga006 < ld_day   #180807-00046
               DISPLAY "delete loga_t WHERE loga006 count: ",sqlca.SQLERRD[3] , " SQLCA.SQLCODE: ",SQLCA.SQLCODE

               DELETE FROM loga_t WHERE loga007 < ld_day   #180807-00046
               DISPLAY "delete loga_t WHERE loga007 count: ",sqlca.SQLERRD[3] , " SQLCA.SQLCODE: ",SQLCA.SQLCODE #200904-00093#1
               
               CALL azzp820_delete_data_mongo("loga_t", ld_day)   #180312-00040#1 add
               #DISPLAY "delete loga_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE  #200904-00093#1

            WHEN "logb_t"
               DISPLAY "logb_t ld_day:",ld_day
            #   DELETE  FROM logb_t where logb007 < TO_DATE(ld_day, 'YY/MM/DD')   #180807-00046 mark                            
               DELETE  FROM logb_t where logb007 < ld_day   #180807-00046
               DISPLAY "delete logb_t where logb007 count: ",sqlca.SQLERRD[3] , " SQLCA.SQLCODE: ",SQLCA.SQLCODE  #200904-00093#1
                    
               
               CALL azzp820_delete_data_mongo("logb_t", ld_day)   #180312-00040#1 add
               #DISPLAY "delete logb_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE  #200904-00093#1

            WHEN "logc_t"
               DISPLAY "logc_t ld_day:",ld_day
            #   DELETE FROM logc_t WHERE logc003 < TO_DATE(ld_day, 'YY/MM/DD')   #180807-00046 mark
               DELETE FROM logc_t WHERE logc003 < ld_day   #180807-00046
               DISPLAY "delete logc_t count: ",sqlca.SQLERRD[3] , " SQLCA.SQLCODE: ",SQLCA.SQLCODE
               CALL azzp820_delete_data_mongo("logc_t", ld_day)   #180312-00040#1 add
               #DISPLAY "delete logc_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE

            WHEN "logd_t"
               DISPLAY "logd_t ld_day:",ld_day
            #   DELETE FROM logd_t WHERE logd006 < TO_DATE(ld_day, 'YY/MM/DD')   #180807-00046 mark
               DELETE FROM logd_t WHERE logd006 < ld_day   #180807-00046
               DISPLAY "delete logd_t WHERE logd006 count: ",sqlca.SQLERRD[3] , " SQLCA.SQLCODE: ",SQLCA.SQLCODE

               DELETE FROM logd_t WHERE logd007 < ld_day   #180807-00046
               DISPLAY "delete logd_t WHERE logd007 count: ",sqlca.SQLERRD[3] , " SQLCA.SQLCODE: ",SQLCA.SQLCODE
               
               CALL azzp820_delete_data_mongo("logd_t", ld_day)   #180312-00040#1 add
               #DISPLAY "delete logd_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE #200904-00093#1

            WHEN "loge_t"
               DISPLAY "logt_t ld_day:",ld_day
            #   DELETE FROM loge_t WHERE loge006 < TO_DATE(ld_day, 'YY/MM/DD')   #180807-00046 mark
               DELETE FROM loge_t WHERE loge006 < ld_day   #180807-00046
               DISPLAY "delete loge_t count: ",sqlca.SQLERRD[3] , " SQLCA.SQLCODE: ",SQLCA.SQLCODE
               CALL azzp820_delete_data_mongo("loge_t", ld_day)   #180312-00040#1 add
               #DISPLAY "delete loge_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE  #200904-00093#1
               #231228-00038#1 add(s)
               IF sqlca.SQLERRD[3] > 0 THEN 
                 IF NOT g_obj.has(g_master_1[li_cnt].ls_table ) THEN 
                    CALL g_obj.put(g_master_1[li_cnt].ls_table , g_master_1[li_cnt].ls_table )
                    LET g_table_list2[g_table_list2.getLength()+1].chk_cnt = TRUE
                    LET g_table_list2[g_table_list2.getLength()].dzea001 = g_master_1[li_cnt].ls_table   
                 ELSE
                 END IF
               END IF
               #231228-00038#1 add(e)               
            WHEN "logi_t"
               DISPLAY "logi_t ld_day:",ld_day
            #   DELETE FROM logi_t WHERE logi002 < TO_DATE(ld_day, 'YY/MM/DD')   #180807-00046 mark
               DELETE FROM logi_t WHERE logi002 < ld_day   #180807-00046
               DISPLAY "delete logi_t count: ",sqlca.SQLERRD[3] , " SQLCA.SQLCODE: ",SQLCA.SQLCODE
               CALL azzp820_delete_data_mongo("logi_t", ld_day)   #180312-00040#1 add
               #DISPLAY "delete logi_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE #200904-00093#1

            WHEN "logs_t"
               DISPLAY "logs_t ld_day:",ld_day
            #   DELETE FROM logs_t WHERE logs006 < TO_DATE(ld_day, 'YY/MM/DD')   #180807-00046 mark
               DELETE FROM logs_t WHERE logs006 < ld_day   #180807-00046
               DISPLAY "delete logs_t count: ",sqlca.SQLERRD[3] , " SQLCA.SQLCODE: ",SQLCA.SQLCODE
               CALL azzp820_delete_data_mongo("logs_t", ld_day)   #180312-00040#1 add
               #DISPLAY "delete logs_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE #200904-00093#1

            WHEN "logt_t"
               DISPLAY "logt_t ld_day:",ld_day
            #   DELETE FROM logt_t WHERE logt002 < TO_DATE(ld_day, 'YY/MM/DD')   #180807-00046 mark
               DELETE FROM logt_t WHERE logt002 < ld_day   #180807-00046
               DISPLAY "delete logt_t count: ",sqlca.SQLERRD[3] , " SQLCA.SQLCODE: ",SQLCA.SQLCODE
               CALL azzp820_delete_data_mongo("logt_t", ld_day)   #180312-00040#1 add
               #DISPLAY "delete logt_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE #200904-00093#1

            WHEN "logv_t"
               DISPLAY "logv_t ld_day:",ld_day
            #   DELETE FROM logv_t WHERE logv002 < TO_DATE(ld_day, 'YY/MM/DD')   #180807-00046 mark
               DELETE FROM logv_t WHERE logv002 < ld_day   #180807-00046
               DISPLAY "delete logv_t count: ",sqlca.SQLERRD[3] , " SQLCA.SQLCODE: ",SQLCA.SQLCODE
               CALL azzp820_delete_data_mongo("logv_t", ld_day)   #180312-00040#1 add
               #DISPLAY "delete logv_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE #200904-00093#1
            WHEN "logw_t"
               DISPLAY "logw_t ld_day:",ld_day
            #   DELETE FROM logw_t WHERE logw003 < TO_DATE(ld_day, 'YY/MM/DD')   #180807-00046 mark
               DELETE FROM logw_t WHERE logw003 < ld_day   #180807-00046
               DISPLAY "delete logw_t count: ",sqlca.SQLERRD[3] , " SQLCA.SQLCODE: ",SQLCA.SQLCODE
               CALL azzp820_delete_data_mongo("logw_t", ld_day)   #180312-00040#1 add
               #DISPLAY "delete logw_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE #200904-00093#1

            WHEN "logx_t"
               DISPLAY "logx_t ld_day:",ld_day
            #   DELETE FROM logx_t WHERE logx001 < TO_DATE(ld_day, 'YY/MM/DD')   #180807-00046 mark
               DELETE FROM logx_t WHERE logx001 < ld_day   #180807-00046
               DISPLAY "delete logx_t count: ",sqlca.SQLERRD[3] , " SQLCA.SQLCODE: ",SQLCA.SQLCODE
               CALL azzp820_delete_data_mongo("logx_t", ld_day)   #180312-00040#1 add
               #DISPLAY "delete logx_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE #200904-00093#1

            WHEN "logz_t"
               DISPLAY "logz_t ld_day:",ld_day
            #   DELETE FROM logz_t WHERE logz001 < TO_DATE(ld_day, 'YY/MM/DD')   #180807-00046 mark
               DELETE FROM logz_t WHERE logz001 < ld_day   #180807-00046
               DISPLAY "delete logz_t count: ",sqlca.SQLERRD[3] , " SQLCA.SQLCODE: ",SQLCA.SQLCODE
               CALL azzp820_delete_data_mongo("logz_t", ld_day)   #180312-00040#1 add
               #DISPLAY "delete logz_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE #200904-00093#1

            #211104-00021#1--start--
            WHEN "lokg_t"
              IF FGL_GETENV("ERPVER") = '1.0' THEN
                 DISPLAY "lokg_t ld_day",ld_day
                 DELETE FROM lokg_t WHERE lokg015 < ld_day
                 DISPLAY "delete lokg_t count: ",sqlca.SQLERRD[3] , " SQLCA.SQLCODE: ",SQLCA.SQLCODE
                 CALL azzp820_delete_data_mongo("lokg_t", ld_day)
              END IF
            #211104-00021#1--end--
            #231228-00038#1 add(s) 
                 IF sqlca.SQLERRD[3] > 0 THEN 
                    IF NOT g_obj.has(g_master_1[li_cnt].ls_table ) THEN 
                       CALL g_obj.put(g_master_1[li_cnt].ls_table , g_master_1[li_cnt].ls_table )
                       LET g_table_list2[g_table_list2.getLength()+1].chk_cnt = TRUE
                       LET g_table_list2[g_table_list2.getLength()].dzea001 = g_master_1[li_cnt].ls_table   
                    ELSE
                    END IF
                 END IF
           #231228-00038#1 add(e)      
            WHEN "wsfa_t" #因wsfa003 是vharcher型態
               LET ls_begin = ld_day 
               LET lc_time = ls_begin
               DISPLAY "wsfa_t ls_begin:",ls_begin 
               DISPLAY "wsfa_t ld_day:",ld_day
            #   DELETE FROM wsfa_t WHERE wsfa003 < TO_DATE(ld_day, 'YY/MM/DD')   #180807-00046 mark
               #DELETE FROM wsfa_t WHERE wsfa003 < ld_day   #180807-00046  #200904-00093#1 #2020/08/30
               DELETE FROM wsfa_t WHERE wsfa003 < lc_time   #180807-00046  #200904-00093#1 #2020/08/30
               DISPLAY "delete wsfa_t count: ",sqlca.SQLERRD[3] , " SQLCA.SQLCODE: ",SQLCA.SQLCODE ," SQLERRMESSAGE:",SQLERRMESSAGE 
              #231228-00038#1 add(s) 
              IF sqlca.SQLERRD[3] > 0 THEN 
                 IF NOT g_obj.has(g_master_1[li_cnt].ls_table ) THEN 
                    CALL g_obj.put(g_master_1[li_cnt].ls_table , g_master_1[li_cnt].ls_table )
                    LET g_table_list2[g_table_list2.getLength()+1].chk_cnt = TRUE
                    LET g_table_list2[g_table_list2.getLength()].dzea001 = g_master_1[li_cnt].ls_table   
                 ELSE
                 END IF
              END IF
              #231228-00038#1 add(e)               
         END CASE
      END IF 
   END FOR
   #231228-00038#1 add(s) 
   #IF ga_gzou[li_cnt2].gzou003 = "ds" THEN  #240213-00001#1 mark             
      CALL  azzp820_release_blob_tablespcace(ga_gzou[li_cnt2].gzou003)  #240213-00001#1 add a_gzou[li_cnt2].gzou003
   #END IF                                   #240213-00001#1 mark 
   #231228-00038#1 add(e) 
   #200904-00093#1--start--
   IF g_dbs_old = ga_gzou[li_cnt2].gzou003 THEN 
   ELSE 
      #斷開目前連線   
      CALL cl_db_disconnect_current()
      #連到一開始程式執行時的對應連線 
      IF cl_db_setconnect(g_dbs_old) THEN END IF
   END IF
   #200904-00093#1--end--   

   END FOR   #200904-00093#1

   #LET l_chk_delete = FALSE   #200904-00093#1
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
PRIVATE FUNCTION azzp820_delete_data_mongo(ps_table,ps_date)
   DEFINE ps_table      STRING 
   DEFINE ps_date       DATETIME YEAR TO SECOND
   DEFINE ls_query      STRING
   DEFINE ls_err         util.JSONObject
   DEFINE ls_result      STRING
   DEFINE ls_res_etag    STRING
   DEFINE li_cnt         LIKE type_t.num5   #240318-00005#1 add


   FOR li_cnt = 1 TO g_pram.getLength()     #240318-00005#1 add
     IF g_pram[li_cnt].lc_value = "2" THEN  #240318-00005#1 add
        CONTINUE FOR                        #240318-00005#1 add
     END IF                                 #240318-00005#1 add

   #180710-00028#1 --START--
   #判斷目前是否為 mongoDB，如果不是則直接 RETURN                #240318-00005#1 mark
   #IF cl_get_para(g_enterprise,"","E-SYS-0735 ") = "2" THEN  #240318-00005#1 mark
   #   RETURN                                                 #240318-00005#1 mark
   #END IF                                                    #240318-00005#1 mark
   #180710-00028#1 -- END --
   
   LET ls_query = "/*?filter={'$or':["
   
   CASE 
      WHEN "loga_t"
         LET ls_query = ls_query ,'{"loga006" : {"$lt":"', ps_date, '"}}]}'
         DISPLAY "delete loga_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE
      WHEN "logb_t"
         LET ls_query = ls_query ,'{"logb007" : {"$lt":"', ps_date, '"}}]}'
         DISPLAY "delete logb_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE
      WHEN "logc_t"
         LET ls_query = ls_query ,'{"logc003" : {"$lt":"', ps_date, '"}}]}'
         DISPLAY "delete logc_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE
      WHEN "logd_t"
         LET ls_query = ls_query ,'{"logd006" : {"$lt":"', ps_date, '"}}]}'
         DISPLAY "delete logd_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE
      WHEN "loge_t"
         LET ls_query = ls_query ,'{"loge006" : {"$lt":"', ps_date, '"}}]}'
         DISPLAY "delete loge_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE
      WHEN "logi_t"
         LET ls_query = ls_query ,'{"logi002" : {"$lt":"', ps_date, '"}}]}'
         DISPLAY "delete logi_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE
      WHEN "logs_t"
         LET ls_query = ls_query ,'{"logs006" : {"$lt":"', ps_date, '"}}]}'
         DISPLAY "delete logs_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE
      WHEN "logt_t"
         LET ls_query = ls_query ,'{"logt002" : {"$lt":"', ps_date, '"}}]}'
         DISPLAY "delete logt_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE
      WHEN "logv_t"
         LET ls_query = ls_query ,'{"logv002" : {"$lt":"', ps_date, '"}}]}'
         DISPLAY "delete logv_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE
      WHEN "logw_t"
         LET ls_query = ls_query ,'{"logw003" : {"$lt":"', ps_date, '"}}]}'
         DISPLAY "delete logw_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE
      WHEN "logx_t"
         LET ls_query = ls_query ,'{"logx001" : {"$lt":"', ps_date, '"}}]}'
         DISPLAY "delete logx_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE
      WHEN "logz_t"
         LET ls_query = ls_query ,'{"logz001" : {"$lt":"', ps_date, '"}}]}'
         DISPLAY "delete logz_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE
      #211104-00021 start
      WHEN 'lokg_t'
         IF FGL_GETENV("ERPVER") = '1.0' THEN
            LET ls_query = ls_query ,'{"lokg015" : {"$lt":"', ps_date, '"}}]}'
            DISPLAY "delete lokg_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE
         END IF
      #211104-00021 end
      WHEN "wsfa_t"
         LET ls_query = ls_query ,'{"wsfa003" : {"$lt":"', ps_date, '"}}]}'
         DISPLAY "delete wsfa_t count: ",sqlca.SQLERRD[3] , "SQLCA.SQLCODE: ",SQLCA.SQLCODE
   END CASE
   
   CALL cl_log_restheart ("DELETE", ps_table, NULL, ls_query, NULL) 
      RETURNING ls_err, ls_res_etag, ls_result
   END FOR #240318-00005#1 add
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzp820_get_gzou003()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2020/09/04 By 01856
# Modify.........:
################################################################################
PRIVATE FUNCTION azzp820_get_gzou003()
   DEFINE ls_sql   STRING 
   DEFINE li_cnt   LIKE type_t.num5
   CALL ga_gzou.clear()
   LET ls_sql = "SELECT DISTINCT gzou003 FROM gzou_t,gzda_t ",
                " WHERE gzoustus = 'Y' ",
                " AND gzda001=gzou003 AND gzdastus = 'Y' AND gzda005 = 'Y' ",
                " ORDER BY gzou003 "

   LET li_cnt = 1         
   DECLARE azzp820_get_gzou003_cs CURSOR FROM ls_sql
      FOREACH azzp820_get_gzou003_cs INTO ga_gzou[li_cnt].gzou003

         LET li_cnt = li_cnt + 1
   END FOREACH
   #221230-00039#1 add(s)
   LET  ga_gzou[li_cnt].gzou003 = "ds"
   FOR li_cnt = 1 TO ga_gzou.getLength()

       DISPLAY " li_cnt :",li_cnt ," gzou003 :", ga_gzou[li_cnt].gzou003
   END FOR
   #221230-00039#1 add(e)    
   
   #CALL ga_gzou.deleteElement(li_cnt)  #221230-00039#1 mark 
   CLOSE azzp820_get_gzou003_cs
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL app820_get_fglrun_ver()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2020/12/13 By 01856
# Modify.........:
################################################################################
PRIVATE FUNCTION app820_get_fglrun_ver()
   DEFINE ls_host     STRING
   DEFINE ch          base.Channel
   DEFINE l_token     base.StringTokenizer    #190314-00035	add
   DEFINE ls_tmp      STRING                  #190314-00035	add
   
   WHENEVER ERROR CALL cl_err_msg_log

   #IF cl_null(gs_fglrun_ver) THEN
      LET ch = base.Channel.create()
      CALL ch.setDelimiter("")
      CALL ch.openPipe("fglrun -V | grep fglrun ","r")  
      LET ls_host = ""
      WHILE ch.read(ls_host)
      END WHILE
      LET ls_host = ls_host.subString(ls_host.getIndexOf("fglrun",1)+6,ls_host.getLength()-1)

      LET l_token = base.StringTokenizer.create(ls_host," ")
      WHILE l_token.hasMoreTokens()
        LET ls_tmp = l_token.nextToken() 
        EXIT WHILE  
      END WHILE 
      LET ls_host = ls_tmp       
      #LET gs_fglrun_ver = ls_host
   #ELSE
   #   LET ls_host = gs_fglrun_ver
   #END IF

   CALL ch.close()
   RETURN ls_host.trim()
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzp820_chk_exist()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2020/12/13 By 01856
# Modify.........:
################################################################################
PRIVATE FUNCTION azzp820_chk_exist()
   DEFINE ls_cmd STRING 
   DEFINE li_exis820    LIKE type_t.num5
   DEFINE l_ch          base.Channel
   DEFINE ls_buf        STRING   
   DEFINE ls_ver        STRING  #fglrun -V 的版本  
   DEFINE ls_str        STRING 
   DEFINE ls_str2       STRING  
   DEFINE li_idx        LIKE type_t.num5 
   DEFINE li_idx2       LIKE type_t.num5    
   DEFINE li_cnt        LIKE type_t.num5
   DEFINE li_sessionkey LIKE type_t.num5    

   LET ls_cmd = "ps -ef | grep azzp820 | grep $ZONE  | grep -v 'sh -c'  "
   LET l_ch = base.Channel.create()    
   CALL l_ch.openPipe(ls_cmd,"r")   #執行指令

   LET li_exis820 = FALSE
   LET ls_ver = app820_get_fglrun_ver()
   IF ls_ver.subString(1,1)>2 THEN 
      LET ls_cmd = "fglrun" 
   ELSE 
      LET ls_cmd = "fglrun-bin"
   END IF 

  WHILE TRUE
      CALL l_ch.readLine() RETURNING ls_buf
      IF ls_buf IS NULL THEN EXIT WHILE END IF

      DISPLAY "ls_buf:",ls_buf
    
      LET li_idx =  ls_buf.getIndexOf(ls_cmd,1) 
      IF(ls_buf.getIndexOf("\"sessionkey\"",li_idx)>0) THEN 

         LET li_idx2 = ls_buf.getIndexOf("\"sessionkey\"",li_idx) 
      END IF 
      IF li_idx2 > 0 THEN 
         LET ls_str2 = ls_buf.subString(li_idx2,ls_buf.getIndexOf("\"parent\"",li_idx2+1)-2)
         LET ls_str = ls_buf.subString(li_idx,li_idx2-2)
      END IF 
      
      DISPLAY "ls_str:",ls_str
      IF ls_buf.getIndexOf(ls_cmd,1) AND ls_buf.getIndexOf("azzp820",1) AND ls_buf.getIndexOf(ls_str,1) THEN #160426-00026 #1 add
         LET li_exis820 = TRUE
         LET li_cnt = li_cnt + 1         
      END IF

   END WHILE   
   LET li_sessionkey = FALSE 
   IF ls_str2.getIndexOf(g_sessionkey,1) THEN 
       LET li_sessionkey = TRUE #表示目前是自己正在執行要排除
   END IF    
   DISPLAY "ls_str2:",ls_str2 ," g_sessionkey:",g_sessionkey 
   #當下時間確保只有一個執行 
   IF li_sessionkey AND li_cnt < 2 THEN 
      LET li_exis820 = FALSE 
   END IF  
  
   RETURN li_exis820
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzp820_chk_prog()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2020/12/13 By 01856
# Modify.........:
################################################################################
PRIVATE FUNCTION azzp820_chk_prog()
  DEFINE li_cnt LIKE type_t.num5
  
  SELECT COUNT(1) INTO li_cnt FROM gzpg_t
   WHERE gzpg002 = 'azzp820'
    AND gzpgstus = 'Y'

  IF li_cnt > 0 THEN 
     RETURN TRUE
  ELSE 
     RETURN FALSE  
  END IF  
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2020/12/13 By 01856
# Modify.........:
################################################################################
PRIVATE FUNCTION azzp820_delete_log_use_count(ls_table,ls_field,ld_dist_date,amount)
   DEFINE ls_sql    STRING 
   DEFINE ls_table  STRING 
   DEFINE ls_field  STRING 
   DEFINE amount    LIKE type_t.num5
   DEFINE ld_dist_date  DATETIME YEAR TO SECOND
   DEFINE lt_start      DATETIME YEAR TO FRACTION(3)
   DEFINE lt_end        DATETIME YEAR TO FRACTION(3)   
   DEFINE ld_day         DATETIME YEAR TO SECOND 
   DEFINE li_cnt         LIKE type_t.num10
   DEFINE li_cnt2        LIKE type_t.num5   
   DEFINE ls_begin       STRING             
   DEFINE lc_time        LIKE type_t.chr50  
   
                
   LET lt_start =  cl_get_current()   
   IF ls_table = "wsfa_t"  THEN 
      LET ls_begin = ld_dist_date 
      LET lc_time = ls_begin
   
      LET ls_sql = "DELETE FROM ",ls_table,  #"loga_t",
                " WHERE ",ls_field, " IN ", # loga006
                " (SELECT ",ls_field,       #loga006 
                "   FROM (SELECT * ",
                "    FROM ",ls_table ,      #loga_t ",
                "    WHERE ",ls_field ," < ? ",
                "     ORDER BY ",ls_field," ASC) lo ",
                " WHERE ROWNUM  < ",amount ," ) "
     PREPARE azzp820_delete_log_count_pre FROM ls_sql

     EXECUTE azzp820_delete_log_count_pre USING lc_time       

   ELSE 
      LET ls_sql = "DELETE FROM ",ls_table,  #"loga_t",
                " WHERE ",ls_field, " IN ", # loga006
                " (SELECT ",ls_field,       #loga006 
                "   FROM (SELECT * ",
                "    FROM ",ls_table ,      #loga_t ",
                "    WHERE ",ls_field ," < ? ",
                "     ORDER BY ",ls_field," ASC) lo ", #loga006
                " WHERE ROWNUM < ",amount ," ) "  #300
      PREPARE azzp820_delete_log_count_pre2 FROM ls_sql                

      EXECUTE azzp820_delete_log_count_pre2 USING ld_dist_date                 
   END  IF 
   
   LET lt_end =  cl_get_current()
   IF ls_table = "wsfa_t" THEN 
      DISPLAY "delete FROM ",ls_table, "  WHERE ",ls_field  , " < ",lc_time ," count: ",sqlca.SQLERRD[3] ," spent time:",lt_end - lt_start   
   ELSE 
      DISPLAY "delete FROM ",ls_table, "  WHERE ",ls_field  ," < " ,ld_dist_date ," count: ",sqlca.SQLERRD[3] ," spent time:",lt_end - lt_start   
   END IF
   #231228-00038#1 add(s)
   IF sqlca.SQLERRD[3] > 0 AND azzp820_get_clob_type(ls_table,ls_field) THEN 
      IF NOT g_obj.has(ls_table) THEN 
         CALL g_obj.put(ls_table, ls_table)
         LET g_table_list2[g_table_list2.getLength()+1].chk_cnt = TRUE
         LET g_table_list2[g_table_list2.getLength()].dzea001 = ls_table  
      ELSE
      END IF 
   END IF
   #231228-00038#1 add(e)   
   FREE azzp820_delete_log_count_pre
   FREE azzp820_delete_log_count_pre2   
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzp820_chk_table_amount(table_id,db_id)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2020/12/13 By 01856
# Modify.........:
################################################################################
PRIVATE FUNCTION azzp820_chk_table_amount(table_id,db_id)
   DEFINE table_id LIKE dzea_t.dzea001
   DEFINE db_id LIKE gzou_t.gzou003
   DEFINE ls_sql   STRING 
   DEFINE li_cnt   LIKE type_t.num10

   LET ls_sql = "SELECT COUNT(1) FROM  ",db_id,".",table_id
   PREPARE azzp820_chk_table_pre FROM ls_sql 
   EXECUTE azzp820_chk_table_pre INTO li_cnt 
   
   FREE azzp820_chk_table_pre 
   DISPLAY "azzp820_chk_table_amount :",table_id ," cnt:",li_cnt
   IF li_cnt > g_argv[03] THEN  #上限

      RETURN TRUE 
   END IF  

   RETURN FALSE  
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzp820_delete_s_detail1_V2(p_d_day)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 日期 By 作者
# Modify.........: #221230-00039#1 add
################################################################################
PRIVATE FUNCTION azzp820_delete_s_detail1_V2(p_d_day)
   DEFINE li_cnt2  LIKE type_t.num5 
   DEFINE li_cnt   LIKE type_t.num5
   DEFINE lc_table_rec  RECORD 
     gzcb002 LIKE gzcb_t.gzcb002,
     gzcb003 LIKE gzcb_t.gzcb003,
     gzcb004 LIKE gzcb_t.gzcb004,
     gzcb005 LIKE gzcb_t.gzcb005
     END RECORD 
   DEFINE lc_dzeb007 LIKE dzeb_t.dzeb007 
   DEFINE ld_day         DATETIME YEAR TO SECOND 
   DEFINE p_d_day        DATE
   DEFINE ls_msg    STRING 
   DEFINE li_cnt3  LIKE type_t.num5  #240318-00005#1 add

   
   LET g_dbs_old = g_dbs   
   FOR li_cnt2 = 1 TO ga_gzou.getLength()

      DISPLAY " 資料庫編號:",ga_gzou[li_cnt2].gzou003    
      IF g_dbs_old = ga_gzou[li_cnt2].gzou003 THEN 
      ELSE 
         CALL cl_db_connect(ga_gzou[li_cnt2].gzou003,FALSE)  
      END IF   
      CALL azzp820_get_para(g_enterprise,"","E-SYS-0735",ga_gzou[li_cnt2].gzou003) #240318-00005#1 add
   
      FOR li_cnt = 1 TO g_master_1.getLength()
         IF g_master_1[li_cnt].li_day_cnt <> 0 THEN
           INITIALIZE lc_table_rec.* TO NULL
           LET ld_day = p_d_day - (g_master_1[li_cnt].li_day_cnt UNITS DAY) 
           DISPLAY "azzp820_delete_s_detail1_V2 目標日期:",p_d_day," 刪一個月前的資料:", p_d_day 
           DISPLAY "保留天數 days: ",g_master_1[li_cnt].li_day_cnt    
           #SELECT gzcb002,gzcb003,gzcb004,gzcb005 INTO lc_table_rec.gzcb002,lc_table_rec.gzcb003,
           #                                            lc_table_rec.gzcb004,lc_table_rec.gzcb005
           #  FROM gzca_t
           #  LEFT JOIN gzcb_t
           #    ON gzca001 = gzcb001        
           # WHERE gzca001 = 385
           # AND  gzcb002 = g_master_1[li_cnt].ls_table  
            CALL azzp820_get_table_info(g_master_1[li_cnt].ls_table)  RETURNING lc_table_rec.*            

            CALL azzp820_get_field_type(lc_table_rec.gzcb002,lc_table_rec.gzcb003) RETURNING lc_dzeb007  
            FOR li_cnt3 = 1 TO g_pram.getLength()               #240318-00005#1 add
            IF g_pram[li_cnt3].lc_value = "2" THEN              #240318-00005#1 add
            #IF cl_get_para(g_enterprise,"","E-SYS-0735") = "2" THEN #240318-00005#1 mark
               #2:oracle db
                 #適用於T100 v2(含)以上不支援，刪除 lokg_t
                 IF FGL_GETENV("ERPVER") = '3.0' AND lc_table_rec.gzcb002 = "lokg_t" THEN
                    DISPLAY "V3.0: 不支援刪除 ",lc_table_rec.gzcb002
                    CONTINUE FOR 
                 END IF 
                 LET ls_msg = "DELETE ",lc_table_rec.gzcb002 
                 CALL azzp820_show_progress(ls_msg)                 

                 CALL azzp820_delete_log_v2(lc_table_rec.gzcb002,lc_table_rec.gzcb003,lc_dzeb007,ld_day)

                 IF NOT cl_null(lc_table_rec.gzcb004) THEN 
                    CALL azzp820_get_field_type(lc_table_rec.gzcb002,lc_table_rec.gzcb004) RETURNING lc_dzeb007  
                    CALL azzp820_delete_log_v2(lc_table_rec.gzcb002,lc_table_rec.gzcb004,lc_dzeb007,ld_day)
                 END IF                                  
            ELSE 
                #1:mongo db
                LET ls_msg = "DELETE ",lc_table_rec.gzcb002 
                CALL azzp820_show_progress(ls_msg) 
                
                CALL azzp820_delete_data_mongo_V2(lc_table_rec.gzcb002,lc_table_rec.gzcb003,ld_day)
                 
                 IF NOT cl_null(lc_table_rec.gzcb004) THEN 
                    CALL azzp820_get_field_type(lc_table_rec.gzcb002,lc_table_rec.gzcb004) RETURNING lc_dzeb007  
                    CALL azzp820_delete_data_mongo_V2(lc_table_rec.gzcb002,lc_table_rec.gzcb004,ld_day)
                 END IF                 
            END IF
            END FOR  #240318-00005#1 add          
         END IF 
      END FOR 
      #231228-00038#1 add(s)
      #IF ga_gzou[li_cnt2].gzou003 = "ds" THEN  #240213-00001#1 mark
         CALL azzp820_release_blob_tablespcace(ga_gzou[li_cnt2].gzou003) #240213-00001#1 add a_gzou[li_cnt2].gzou003
      #END IF                                   #240213-00001#1 mark
      #231228-00038#1 add(e)
      IF g_dbs_old = ga_gzou[li_cnt2].gzou003 THEN 
      ELSE 
         #斷開目前連線   
         CALL cl_db_disconnect_current()
         #連到一開始程式執行時的對應連線 
         IF cl_db_setconnect(g_dbs_old) THEN END IF
      END IF      
   END FOR
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzp820_get_field_type(p_table,p_colname)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 日期 By 作者
# Modify.........: #221230-00039#1 add
################################################################################
PRIVATE FUNCTION azzp820_get_field_type(p_table,p_colname)
   DEFINE p_table     LIKE dzeb_t.dzeb001
   DEFINE p_colname   LIKE dzeb_t.dzeb002
   DEFINE r_dzeb007  LIKE dzeb_t.dzeb007

  
   LET r_dzeb007 = ""
   SELECT dzeb007 INTO r_dzeb007  FROM  dzeb_t
    WHERE  dzeb001 = p_table #'wsfa_t'
     AND dzeb002 = p_colname #'wsfa003' 
   DISPLAY " azzp820_get_field_type: p_table:",p_table ," p_colname:",p_colname, " r_dzeb007:",r_dzeb007
   RETURN r_dzeb007
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzp820_delete_log_v2(pc_gzcb002,pc_gzcb003,pc_dzeb007,ld_dist_date)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 日期 By 作者
# Modify.........: #221230-00039#1 add
################################################################################
PRIVATE FUNCTION azzp820_delete_log_v2(pc_gzcb002,pc_gzcb003,pc_dzeb007,ld_dist_date)
   DEFINE pc_gzcb002      LIKE gzcb_t.gzcb002
   DEFINE pc_gzcb003      LIKE gzcb_t.gzcb002
   DEFINE pc_dzeb007      LIKE dzeb_t.dzeb007  
   DEFINE ld_dist_date    DATETIME YEAR TO SECOND 
   DEFINE ls_sql          STRING 
   DEFINE lt_start      DATETIME YEAR TO FRACTION(3)
   DEFINE lt_end        DATETIME YEAR TO FRACTION(3)
   DEFINE ls_begin       STRING             
   DEFINE lc_time        LIKE type_t.chr50  
   
   LET lt_start =  cl_get_current()
   LET ls_sql = "DELETE FROM ",pc_gzcb002,
                " WHERE ",pc_gzcb003 ," <= ? "


   PREPARE azzp820_delete_log_pre FROM ls_sql
 
   IF DOWNSHIFT(pc_dzeb007) = "varchar2" OR UPSHIFT(pc_dzeb007) = "VARCHAR2" THEN 

      LET ls_begin = ld_dist_date 
      LET lc_time = ls_begin
      
      EXECUTE azzp820_delete_log_pre USING lc_time
   ELSE 
      EXECUTE azzp820_delete_log_pre USING ld_dist_date
   END IF   
   
   LET lt_end =  cl_get_current() 
   DISPLAY "delete ", pc_gzcb002 ," WHERE ", pc_gzcb003 , " < ",ld_dist_date," count: ",sqlca.SQLERRD[3] ,  " SQLCA.SQLCODE: ",SQLCA.SQLCODE ," SQLERRMESSAGE:",SQLERRMESSAGE   #231221-00028#1 add SQLERRMESSAGE
   DISPLAY "spent time:",lt_end - lt_start   
   #231228-00038#1 add(s)
   IF sqlca.SQLERRD[3] > 0 AND azzp820_get_clob_type(pc_gzcb002,pc_gzcb003) THEN 
      IF NOT g_obj.has(pc_gzcb002) THEN 
         CALL g_obj.put(pc_gzcb002, pc_gzcb002)
         LET g_table_list2[g_table_list2.getLength()+1].chk_cnt = TRUE
         LET g_table_list2[g_table_list2.getLength()].dzea001 = pc_gzcb002  
      ELSE
      END IF 

   END IF 
  #231228-00038#1 add(e)
   FREE azzp820_delete_log_pre            
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzp820_delete_data_mongo_V2(ps_table,ps_field,ps_date)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 日期 By 作者
# Modify.........: #221230-00039#1 add
################################################################################
PRIVATE FUNCTION azzp820_delete_data_mongo_V2(ps_table,ps_field,ps_date)
   DEFINE ps_table      STRING 
   DEFINE ps_field      STRING
   DEFINE ps_date       DATETIME YEAR TO SECOND
   DEFINE ls_query      STRING
   DEFINE ls_err         util.JSONObject
   DEFINE ls_result      STRING
   DEFINE ls_res_etag    STRING
   
   LET ls_query = "/*?filter={'$or':["
   #LET ls_query = ls_query ,'{\"',ps_field,'\" : {"$lt":"', ps_date, '"}}]}'
   LET ls_query = ls_query ,'{\"',ps_field,'\" : {"$lt":"', ps_date, '"}}]}'
   CALL cl_log_restheart ("DELETE", ps_table, NULL, ls_query, NULL) 
      RETURNING ls_err, ls_res_etag, ls_result
      DISPLAY "delete azzp820_delete_data_mongo2 ls_err: ",ls_err, " ls_res_etag: ",ls_res_etag," ls_result:",ls_result  

   #ex:{"loga006" : {"$lt":"2022-12-18 00:00:00"}}]}
   #ex:/*?filter={'$or':[{"loga006" : {"$lt":"2022-12-18 00:00:00"}}]}
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzp820_delete_log_use_count_V2(ls_table,ls_field,ld_dist_date,amount,pc_dzeb007)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 日期 By 作者
# Modify.........: #221230-00039#1 add
################################################################################
PRIVATE FUNCTION azzp820_delete_log_use_count_V2(ls_table,ls_field,ld_dist_date,amount,pc_dzeb007)
   DEFINE ls_sql        STRING 
   DEFINE ls_table      STRING 
   DEFINE ls_field      STRING 
   DEFINE amount        LIKE type_t.num5
   DEFINE pc_dzeb007    LIKE dzeb_t.dzeb007
   DEFINE ld_dist_date  DATETIME YEAR TO SECOND
   DEFINE lt_start      DATETIME YEAR TO FRACTION(3)
   DEFINE lt_end        DATETIME YEAR TO FRACTION(3)   
   DEFINE ld_day         DATETIME YEAR TO SECOND 
   DEFINE li_cnt         LIKE type_t.num10
   DEFINE li_cnt2        LIKE type_t.num5   
   DEFINE ls_begin       STRING             
   DEFINE lc_time        LIKE type_t.chr50         
   
   DISPLAY " azzp820_delete_log_use_count_V2 pc_dzeb007:",pc_dzeb007     
   LET lt_start =  cl_get_current()   
   IF DOWNSHIFT(pc_dzeb007) = "varchar2" OR UPSHIFT(pc_dzeb007) = "VARCHAR2"  THEN
      LET ls_begin = ld_dist_date 
      LET lc_time = ls_begin
      LET ls_sql = "DELETE FROM ",ls_table,  #"loga_t",
                " WHERE ",ls_field, " IN ", # loga006
                " (SELECT ",ls_field,       #loga006 
                "   FROM (SELECT * ",
                "    FROM ",ls_table ,      #loga_t ",
                "    WHERE ",ls_field ," < ? ",
                "     ORDER BY ",ls_field," ASC) lo ",
                " WHERE ROWNUM  <= ",amount ," ) "
     PREPARE azzp820_delete_log_count_pre3 FROM ls_sql
     EXECUTE azzp820_delete_log_count_pre3 USING lc_time 
     LET li_cnt = SQLCA.sqlerrd[3]
     FREE azzp820_delete_log_count_pre3     
      DISPLAY " is varchar2 ls_sql:",ls_sql       
   ELSE 
      LET ls_sql = "DELETE FROM ",ls_table,  #"loga_t",
                " WHERE ",ls_field, " IN ", # loga006
                " (SELECT ",ls_field,       #loga006 
                "   FROM (SELECT * ",
                "    FROM ",ls_table ,      #loga_t ",
                "    WHERE ",ls_field ," < ? ",
                "     ORDER BY ",ls_field," ASC) lo ", #loga006
                " WHERE ROWNUM <= ",amount ," ) "  #300
      PREPARE azzp820_delete_log_count_pre4 FROM ls_sql                
      EXECUTE azzp820_delete_log_count_pre4 USING ld_dist_date
      LET li_cnt = SQLCA.sqlerrd[3]
      FREE azzp820_delete_log_count_pre4  
      DISPLAY " is not varchar2 ls_sql:",ls_sql        
   END  IF 
   
   LET lt_end =  cl_get_current()
   DISPLAY "delete FROM ",ls_table, "  WHERE ",ls_field  , " <= ",lc_time ," count: ",li_cnt ," spent time:",lt_end - lt_start   
   #231228-00038#1 add(s)
   IF sqlca.SQLERRD[3] > 0 AND azzp820_get_clob_type(ls_table,ls_field) THEN 
      IF NOT g_obj.has(ls_table) THEN 
         CALL g_obj.put(ls_table, ls_table)
         LET g_table_list2[g_table_list2.getLength()+1].chk_cnt = TRUE
         LET g_table_list2[g_table_list2.getLength()].dzea001 = ls_table  
      ELSE
      END IF 
   END IF    
   #231228-00038#1 add(e)   
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzp820_get_table_list(p_days)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 日期 By 作者
# Modify.........: #221230-00039#1 add
################################################################################
PRIVATE FUNCTION azzp820_get_table_list(p_days)
   DEFINE li_cnt     LIKE type_t.num5
   DEFINE p_days     LIKE type_t.num5
   
    DECLARE azzp820_get_table_list_cs CURSOR FOR 
     SELECT gzcb002 FROM  gzca_t 
      LEFT JOIN gzcb_t ON gzca001 = gzcb001
       WHERE  gzca001 = 385
      ORDER  BY  gzcb012 

    LET li_cnt = 1  
    FOREACH azzp820_get_table_list_cs INTO g_master_1[li_cnt].ls_table
       LET g_master_1[li_cnt].li_day_cnt = p_days
       LET li_cnt = li_cnt + 1
    END FOREACH
    CALL g_master_1.deleteElement(g_master_1.getLength())
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzp820_show_progress(ps_msg)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 日期 By 作者
# Modify.........: #221230-00039#1 add
################################################################################
PRIVATE FUNCTION azzp820_show_progress(ps_msg)
  DEFINE ps_msg STRING 

   IF g_bgjob = "N" THEN 
      CALL cl_progress_no_window_ing(ps_msg)
   END IF
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzp820_get_table_info(p_gzcb002)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 日期 By 作者
# Modify.........: #221230-00039#1 add
################################################################################
PRIVATE FUNCTION azzp820_get_table_info(p_gzcb002)
   DEFINE p_gzcb002 LIKE gzcb_t.gzcb002
   DEFINE lc_table_rec  RECORD 
       gzcb002 LIKE gzcb_t.gzcb002,
       gzcb003 LIKE gzcb_t.gzcb003,
       gzcb004 LIKE gzcb_t.gzcb004,
       gzcb005 LIKE gzcb_t.gzcb005
       END RECORD 
   DEFINE lc_dzeb007 LIKE dzeb_t.dzeb007

    SELECT gzcb002,gzcb003,gzcb004,gzcb005 INTO lc_table_rec.gzcb002,lc_table_rec.gzcb003,
                                                lc_table_rec.gzcb004,lc_table_rec.gzcb005
        FROM gzca_t
        LEFT JOIN gzcb_t ON gzca001 = gzcb001        
    WHERE gzca001 = 385
    AND  gzcb002 = p_gzcb002

    RETURN lc_table_rec.*
END FUNCTION

################################################################################
# Descriptions...: 釋放DSBLOB 空間(clob型態占用的就是DSBLOB的空間，
#                  使用DELETE語法，刪除table資料，無法發釋放空間)
# Memo...........:
# Usage..........: CALL azzp820_release_blob_tablespcace(ps_now_dbs)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2024/01/11 By 作者
# Modify.........:
################################################################################
PRIVATE FUNCTION azzp820_release_blob_tablespcace(ps_now_dbs)
  DEFINE ls_sql       STRING 
  DEFINE ls_sql2       STRING 
  DEFINE ls_sql3       STRING 
  DEFINE ls_sql4       STRING  
  DEFINE li_cnt       LIKE type_t.num5
  DEFINE l_buf        base.StringBuffer
  DEFINE l_buf2        base.StringBuffer  
  DEFINE ls_str       STRING
  DEFINE ls_str2      STRING  
  DEFINE ls_table     LIKE dzea_t.dzea001
  DEFINE l_remove     RECORD 
       lc_table          LIKE dzeb_t.dzeb001,
       lc_column         LIKE dzeb_t.dzeb002,
       lc_sql_text       LIKE type_t.chr500,
       lc_status         LIKE type_t.chr500,
       tablespace_name   LIKE type_t.chr500,
       owner             LIKE type_t.chr100   #240213-00001#1  
       END RECORD
  DEFINE li_chk    LIKE type_t.num5 
  DEFINE ps_now_dbs    LIKE gzou_t.gzou003    #240213-00001#1
     
  LET l_buf = base.StringBuffer.create()
  LET l_buf2 = base.StringBuffer.create()
  CALL l_buf.clear()
  FOR li_cnt = 1 TO g_table_list2.getLength()
      IF g_table_list2[li_cnt].chk_cnt THEN 

         CALL l_buf.append("'")
         CALL l_buf.append(g_table_list2[li_cnt].dzea001)
         CALL l_buf.append("',")
      END IF 

  END FOR 
  DISPLAY "azzp820_release_blob_tablespcace g_master_1.getLength():",g_master_1.getLength()
  IF l_buf.getLength() > 0 THEN 
     CALL l_buf2.clear()
     LET ls_str = l_buf.subString(1,l_buf.getLength()-1)
     LET ls_str = "(",ls_str,")"

     LET ls_sql  = "SELECT DISTINCT dzeb001 FROM  dzeb_t 
                     WHERE  (dzeb006  = 'B001' or dzeb006  = 'Z001') 
                      AND  dzeb001 IN  " ,ls_str
 DISPLAY "azzp820_release_blob_tablespcace ls_sql:",ls_sql                 
      DECLARE azzp820_release_blob_cs CURSOR FROM ls_sql
      FOREACH azzp820_release_blob_cs INTO ls_table
 DISPLAY "azzp820_release_blob_tablespcace ls_table:",ls_table                 
         CALL l_buf2.append("'")
         CALL l_buf2.append(ls_table)
         CALL l_buf2.append("',")
      END FOREACH
 DISPLAY "azzp820_release_blob_tablespcace l_buf.getLength():",l_buf.getLength()           
      IF l_buf2.getLength() > 0 THEN       


         LET ls_str2 = l_buf2.subString(1,l_buf2.getLength()-1)
         LET ls_str2 = "(",ls_str2,")" 
         LET ls_sql2 = "SELECT  table_name, column_name, 'alter table ' || owner || '.' || table_name || ' move lob(' || column_name || ') store as (tablespace dsblob)' sqltext
                         FROM  dba_lobs
                          WHERE  tablespace_name = 'DSBLOB' AND table_name IN ",UPSHIFT(ls_str2) 
                         , " AND owner ='",UPSHIFT(ps_now_dbs),"'"   #240213-00001#1 add " AND owner ='",UPSHIFT(ps_now_dbs),"'

         LET ls_sql3 = "SELECT  tablespace_name ,status , 'alter index '||owner||'.'||index_name||' rebuild' rebuild_txt
                        FROM  dba_indexes 
                        WHERE status<>'VALID' AND table_name IN "                         
                          
         DECLARE azzp820_release_blob_cs2 CURSOR FROM ls_sql2
         FOREACH azzp820_release_blob_cs2 INTO l_remove.lc_table,l_remove.lc_column,l_remove.lc_sql_text
            #LET ls_sql4 = l_remove.lc_sql_text
            PREPARE azzp820_release_blob_pre FROM l_remove.lc_sql_text
            EXECUTE azzp820_release_blob_pre
            DISPLAY "azzp820_release_blob_pre ",ps_now_dbs," SQLCA.SQLCODE: ",SQLCA.SQLCODE , " SQLCA.sqlerrd[3]:",SQLCA.sqlerrd[3]," SQLERRMESSAGE:",SQLERRMESSAGE 
            IF NOT SQLCA.SQLCODE  THEN 

               LET ls_sql4 = ls_sql3 ,UPSHIFT(ls_str) , " AND owner ='",UPSHIFT(ps_now_dbs),"'"    #240213-00001#1 add " AND owner ='",UPSHIFT(ps_now_dbs),"'  
               
               
               DECLARE azzp820_release_blob_cs3 CURSOR FROM ls_sql4
               FOREACH azzp820_release_blob_cs3 INTO l_remove.tablespace_name,l_remove.lc_status,l_remove.lc_sql_text 
                  #LET ls_sql4 = l_remove.lc_sql_text  
               
                  PREPARE azzp820_release_blob_pre2 FROM l_remove.lc_sql_text 
                  EXECUTE azzp820_release_blob_pre2
                  DISPLAY "azzp820_release_blob_pre2 ",ps_now_dbs, " SQLCA.SQLCODE: ",SQLCA.SQLCODE ," SQLCA.sqlerrd[3]:",SQLCA.sqlerrd[3] ," SQLERRMESSAGE:",SQLERRMESSAGE             
                  FREE azzp820_release_blob_pre2          
               END FOREACH 
            ELSE 

            END IF  
            FREE azzp820_release_blob_pre  
            # 
             
         END FOREACH                                       
      END IF                
  END IF 
END FUNCTION

################################################################################
# Descriptions...: 取clob 欄位型態
# Memo...........:
# Usage..........: CALL azzp820_get_clob_type(p_table,p_colname)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2024/01/11 By 作者
# Modify.........:
################################################################################
PRIVATE FUNCTION azzp820_get_clob_type(p_table,p_colname)
   DEFINE p_table     LIKE dzeb_t.dzeb001
   DEFINE p_colname   LIKE dzeb_t.dzeb002
   DEFINE r_dzeb007   LIKE dzeb_t.dzeb007
   DEFINE li_cnt      LIKE type_t.num5

  
   LET r_dzeb007 = ""
   SELECT COUNT(1) INTO li_cnt FROM  dzeb_t
    WHERE  dzeb001 = p_table #'wsfa_t'
    AND (dzeb007 = 'clob' OR dzeb007 = 'blob') 
   IF li_cnt > 0 THEN 

     RETURN TRUE
   ELSE 
     RETURN FALSE   
   END IF 
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzp820_get_para(pc_enterprise,pc_site,pc_para_id,pc_gzou003)
#                  RETURNING 回传参数
# Input parameter: pc_enterprise   企業編號
#                : pc_site         據點編號
#                : pc_para_id      參數編號
#                : pc_gzou003      資料庫編號
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PRIVATE FUNCTION azzp820_get_para(pc_enterprise,pc_site,pc_para_id,pc_gzou003)
   DEFINE ls_sql         STRING 
   DEFINE pc_enterprise   LIKE type_t.num5
   DEFINE pc_site         LIKE type_t.chr20
   DEFINE pc_para_id      LIKE gzsz_t.gzsz002 #參數編號
   DEFINE lc_ooaa002      LIKE ooaa_t.ooaa002
   DEFINE li_cnt2         LIKE type_t.num5
   DEFINE li_cnt          LIKE type_t.num5
   DEFINE pc_gzou003      LIKE gzou_t.gzou003


   SELECT  COUNT(DISTINCT ooaa002) INTO li_cnt2 FROM ooaa_t
     LEFT  JOIN  gzou_t ON  ooaaent = gzou001
     WHERE  ooaa001 = pc_para_id
     AND  gzou003 = pc_gzou003

   CALL g_pram.clear()
   IF li_cnt2 > 0 THEN 
   
      LET li_cnt = 1
      DECLARE azzp820_get_para_cs CURSOR FOR 

       SELECT  DISTINCT ooaa002 FROM ooaa_t
         LEFT  JOIN  gzou_t ON  ooaaent = gzou001
         WHERE  ooaa001 = pc_para_id
         AND  gzou003 = pc_gzou003      
      FOREACH azzp820_get_para_cs INTO lc_ooaa002

         IF lc_ooaa002 = "NOT_DEFINED" THEN 
            CONTINUE FOREACH 
         END IF 
         IF cl_null(lc_ooaa002) THEN             
            SELECT gzsz008 INTO lc_ooaa002    
             FROM gzsz_t                                                                               #170502-00044
              WHERE gzsz002 = pc_para_id
         END IF 
         DISPLAY "0 參數設定lc_ooaa002:",lc_ooaa002  ," 資料庫編號:",pc_gzou003," 參數編號:",pc_para_id
         LET g_pram[li_cnt].lc_value = lc_ooaa002
         LET li_cnt = li_cnt + 1
      END FOREACH

   ELSE
       LET li_cnt = li_cnt + 1
      #取預設值
      SELECT gzsz008 
        INTO lc_ooaa002    
        FROM gzsz_t                                                                               #170502-00044
       WHERE gzsz002 = lc_para_id
       DISPLAY "1 參數設定lc_ooaa002:",lc_ooaa002  ," 資料庫編號:",pc_gzou003," 參數編號:",pc_para_id
     LET g_pram[li_cnt].lc_value = lc_ooaa002       
   END IF 
   
   CALL g_pram.deleteElement(li_cnt)

   CLOSE azzp820_get_para_cs
END FUNCTION

#end add-point
 
{</section>}
 
