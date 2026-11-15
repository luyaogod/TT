#該程式已解開Section, 不再透過樣板產出!
#該程式非freestyle程式!
{<section id="azzq080.description" type="s" >}
#+ Version..: T100-ERP-1.00.00(SD版次:1,PR版次:1) Build-000025
#+ 
#+ Filename...: azzq080
#+ Description: 授權人數清單查詢
#+ Creator....: 00413(2014-08-07 17:05:53)
#+ Modifier...: 00413(2014-09-19 17:41:51) -SD/PR- 20992
#+ Buildtype..: 應用 q02 樣板自動產生
#+ 以上段落由子樣板a00產生
 
{</section>}
 
{<section id="azzq080.global" type="m" >}
#+ Modifier...: No.170814-00017#1  2017/08/16  frank0521  增加轉出excel及搜尋功能
#+ Modifier...: No.170821-00043#1  2017/08/23  frank0521  增加程式註解及log紀錄
#+ Modifier...: No.180822-00047#1  2018/08/22  01856      調整顯示資訊當無法顯示使用者名稱，顯示此使用者由其他區域登入,無法取得使用者名稱
#+ Modifier...: No.181108-00014#1  2018/11/08  04013      調整grep PID指令
#+ Modifier...: No.181117-00002#1  2018/11/07  01856      調整授權清單人員執行作業清單 顯示順序 ，把企業編號 移到 作業編號前面
#+ Modifier...: No.181117-00002#1  2018/12/07  01856      調整azzq080 顯示資訊 此使用者由其主機登入,無法取得使用者名稱 
                                                          #調整顯示畫面 把執行區域 移到 企業編號前面
#+ Modifier...: No.181207-00062#2  2018/12/10  01856      azzq080 優化 檢查PID 是否存在主機上和修正授權人員目前執行作業清單，只顯示以這台主機運行清單
#+ Modifier...: No.190503-00015#1  2019/05/03  01856      解析報表主機ip 錯誤原因是報表主機的經測試，在報表主機使用xshell執行echo $FGLSERVER會顯示為t100rep而非IP
#+ Modifier...: No.190925-00017#2  2019/09/26  13070      报表WHERE条件中漏加：cl_sql_add_filter导致权限过滤段缺失
#+ Modifier...: No.191111-00008#1  2019/05/03  01856      修正上一版次單號#190926-00017 對應應該190925-00017#2 報表WHERE條件中漏加cl_sql_add_filter的SQL 錯誤
#+ Modifier...: No.191120-00006#1  2019/11/20  01856      建立版本資訊 r.r azzq080 -v
#+ Modifier...: No.200827-00023#1  2020/08/27  05598      修正顯示清單與數量不一致問題，並於清單增加序號欄位
#+ Modifier...: No.220801-00038#1  2022/08/01  05598      授權機制優化，1.登入畫面使用CPU授權，登入後改用Runtime授權 2.防連點機制改判斷weblogin log   
#                                                                      3.重複登入控卡改為checkbox 4.閒置控卡機優化，拿掉一次性登入設定，拿掉提示閒置到期警告，改為閒置時間到強制關閉。
#+ Modifier...: No.230104-00013#1  2023/01/04 By 05598    用戶數已滿以CPU license彈窗顯示訊息
#+ Modifier...: No.230106-00023#1  2023/01/06 By 05598    使用AD/LDAP帳號登入時，在login畫面不佔用授權導致登入失敗
#+ Modifier...: No.230116-00015#2  2023/01/16 By 05598    閒置到期顯示訊息不佔用授權
#+ Modifier...: No.231019-00014#1  2023/10/20 By 20992    調整畫面增加備註，並增加版次訊息給自助更新程式判定版次用
#+ Modifier...: No.240111-00016#1  2024/01/11 By 20992    CPU授權型式資訊呈現處理
#+ Modifier...: No.240320-00003#1  2024/03/20 By 20992    CPU授權卡控呈現調整
#+ Modifier...: No.240628-00031#1  2024/06/28 By 20992    畫面呈現、原始授權數據呈現調整
#+ Modifier...: No.250328-00007#1  2025/03/28 By 20992    統計功能修正及增加在線人數功能
#+ Modifier...: No.250709-00026#1  2025/08/13 By 20992    查詢功能問題修正
#+ Modifier...: No.250901-00010#1  2025/09/01 By 20992    查詢功能問題修正
#+ Modifier...: No.251028-00009#1  2025/10/29 By 20992    修正畫面欄位對不上導致查詢功能問題部分
IMPORT os
IMPORT util
#add-point:增加匯入項目
#220801-00038#1 add s
#cl_poi_fillExcel
IMPORT JAVA java.sql.Connection
IMPORT JAVA java.sql.DriverManager
IMPORT JAVA java.sql.ResultSet
IMPORT JAVA java.sql.SQLException
IMPORT JAVA java.sql.Statement
IMPORT JAVA java.sql.ResultSetMetaData
IMPORT JAVA java.sql.Timestamp
IMPORT JAVA com.digiwin.t100.JdbcUtil
IMPORT JAVA java.lang.Double
#cl_poi_fillExcel
IMPORT JAVA org.apache.poi.xssf.streaming.SXSSFWorkbook
IMPORT JAVA org.apache.poi.ss.usermodel.Cell
IMPORT JAVA org.apache.poi.ss.usermodel.Row
IMPORT JAVA org.apache.poi.ss.usermodel.Sheet
IMPORT JAVA java.io.FileOutputStream
#220801-00038#1 add e
#end add-point
 
SCHEMA ds
 
GLOBALS "../../cfg/top_global.inc"
 
#add-point:增加匯入變數檔
#CONSTANT ls_version2 = "1.0.01 build-20231020"    #231019-00014#1 調整版次判定方式，給自助更新程式判定版次用
#CONSTANT ls_version2 = "1.0.02 build-20240111"    #240111-00016#1 CPU授權型式資訊呈現處理
#CONSTANT ls_version2 = "1.1.01 build-20240628"    #240628-00031#1 畫面呈現、原始授權數據呈現調整
#CONSTANT ls_version2 = "1.1.02 build-20250328"    #250328-00007#1 統計功能修正及增加在線人數功能
#CONSTANT ls_version2 = "1.1.03 build-20250813"    #250709-00026#1 查詢功能問題修正
CONSTANT ls_version2 = "1.1.04 build-20250901"    #250901-00010#1 查詢功能問題修正
#end add-point
 
#單身 type 宣告
PRIVATE TYPE type_g_gzwl_d RECORD
       #statepic       LIKE type_t.chr1,
       sel LIKE type_t.chr1, 
   gzwl001 LIKE gzwl_t.gzwl001,     #使用者編號
   gzwl002 LIKE type_t.chr20,       #使用者帳號          #240628-00031#1 add
   gzwl009 LIKE gzwl_t.gzwl009,      #員工編號             #240628-00031#1 add
   gzwl002_desc LIKE type_t.chr500, #使用者全名          #240628-00031#1 add
   gzxa007 LIKE gzxa_t.gzxa007,     #使用者預設營運據點    #240628-00031#1 add
   gzwl006 LIKE gzwl_t.gzwl006,      #使用者IP          #181207-00062#2
   gzwl010 LIKE gzwl_t.gzwl010,     #執行區域          #181207-00005#1 
   gzwl012 LIKE gzwl_t.gzwl012,     #企業編號          #181117-00002#1 
   gzwl003 LIKE gzwl_t.gzwl003,     #作業編號
   gzwl003_desc LIKE type_t.chr500, 
   gzwl004 LIKE gzwl_t.gzwl004,     #執行開始時間
   gzwl015 LIKE gzwl_t.gzwl015,     #AP主機
   gzwl008 LIKE gzwl_t.gzwl008,     #PID
   #gzwl010 LIKE gzwl_t.gzwl010,     #執行區域         #181207-00005#1 
   gzwl007 LIKE gzwl_t.gzwl007,     #執行模式
   gzwl014 LIKE gzwl_t.gzwl014,     #背景作業
   gzwl013 LIKE gzwl_t.gzwl013     #WS產品來源
       END RECORD
PRIVATE TYPE type_g_gzwl2_d RECORD
   num LIKE type_t.num5,         #序號                  #200827-00023#1 add
   gzwl005 LIKE type_t.chr100,  #授權編號
   gzwl005_desc LIKE type_t.chr500,  #FrontEndId 
   gzwl006 LIKE gzwl_t.gzwl006,     #使用者IP
   gzwl002 LIKE type_t.chr20,       #使用者編號
   gzwl002_desc LIKE type_t.chr500, #使用者全名
   port LIKE type_t.chr20,         #連接port                  #220801-00038#1 add
   gzxa007 LIKE gzxa_t.gzxa007,     #使用者預設營運據點
   gzwl015 LIKE gzwl_t.gzwl015,     #AP主機 
   pid LIKE type_t.chr500           #pid               #240628-00031#1 add
       END RECORD
 
 
 
#模組變數(Module Variables)
DEFINE g_master                     type_g_gzwl_d
DEFINE g_master_t                   type_g_gzwl_d
DEFINE g_gzwl_d          DYNAMIC ARRAY OF type_g_gzwl_d
DEFINE g_gzwl_d_t        type_g_gzwl_d
DEFINE g_gzwl2_d   DYNAMIC ARRAY OF type_g_gzwl2_d
DEFINE g_gzwl2_d_t type_g_gzwl2_d
 
 
      
DEFINE g_wc                 STRING
DEFINE g_wc_t               STRING                        #儲存 user 的查詢條件
DEFINE g_wc2                STRING
DEFINE g_wc_filter          STRING
DEFINE g_wc_filter_t        STRING
DEFINE g_sql                STRING
DEFINE g_forupd_sql         STRING                        #SELECT ... FOR UPDATE SQL
DEFINE g_before_input_done  LIKE type_t.num5
DEFINE g_cnt                LIKE type_t.num10    
DEFINE l_ac                 LIKE type_t.num5              
DEFINE l_ac_d               LIKE type_t.num5              #單身idx 
DEFINE g_curr_diag          ui.Dialog                     #Current Dialog
DEFINE gwin_curr            ui.Window                     #Current Window
DEFINE gfrm_curr            ui.Form                       #Current Form
DEFINE g_current_page       LIKE type_t.num5              #目前所在頁數
DEFINE g_detail_cnt         LIKE type_t.num5              #單身 總筆數(所有資料)
DEFINE g_detail_cnt2        LIKE type_t.num5              #單身 總筆數(所有資料)
DEFINE g_ref_fields         DYNAMIC ARRAY OF VARCHAR(500) #ap_ref用陣列
DEFINE g_rtn_fields         DYNAMIC ARRAY OF VARCHAR(500) #ap_ref用陣列
DEFINE g_ref_vars           DYNAMIC ARRAY OF VARCHAR(500) #ap_ref用陣列
DEFINE gs_keys              DYNAMIC ARRAY OF VARCHAR(500) #同步資料用陣列
DEFINE gs_keys_bak          DYNAMIC ARRAY OF VARCHAR(500) #同步資料用陣列
DEFINE g_insert             LIKE type_t.chr5              #是否導到其他page
DEFINE g_error_show         LIKE type_t.num5
DEFINE g_master_idx         LIKE type_t.num5
DEFINE g_detail_idx         LIKE type_t.num5
DEFINE g_detail_idx2        LIKE type_t.num5
DEFINE g_hyper_url          STRING                        #hyperlink的主要網址
 
 
#多table用wc
DEFINE g_wc_table           STRING
 
 
 
DEFINE g_wc_filter_table           STRING
 
 
 
 
#add-point:自定義模組變數(Module Variable)
DEFINE g_auth_cnt     LIKE type_t.num5   #授權人數
DEFINE g_limit_cnt    LIKE type_t.num5   #執行上限
DEFINE g_auth_notover BOOLEAN            #授權是否已滿
DEFINE g_exec_notover BOOLEAN            #執行上限是否已滿
DEFINE g_keyword_h    DYNAMIC ARRAY OF RECORD
          keyname     STRING,
          wordnum     SMALLINT,
          field       STRING
                      END RECORD
DEFINE g_keyword_d    DYNAMIC ARRAY OF RECORD
          keyname     STRING,
          wordnum     SMALLINT,
          field       STRING
                      END RECORD
DEFINE g_fglwrt       STRING
DEFINE g_dbs_old      STRING
#160726-00036#1
DEFINE g_aid          DYNAMIC ARRAY OF RECORD
          gzwl009     LIKE gzwl_t.gzwl009,     #員工編號
          gzwl012     LIKE gzwl_t.gzwl012      #企業編號
                      END RECORD
DEFINE g_wc_table2    STRING
#181207-00062#2--start--
PRIVATE TYPE type_license_d RECORD

      fdid        LIKE type_t.chr100,  #VARCHAR(100),   #gzwx001
      pid         LIKE type_t.chr20,   #VARCHAR(20),    #gzwx002
      name        LIKE type_t.chr50,   #VARCHAR(50),    #gzwx003
      ip          LIKE type_t.chr20,   #VARCHAR(20),    #gzwx004 
      port        LIKE type_t.chr20,   #VARCHAR(20),    #port        #220801-00038#1 add
      exemode     LIKE type_t.chr20,   #VARCHAR(20),    #gzwx005
      agent       LIKE type_t.chr20,   #VARCHAR(20),    #gzwx006
      appname     LIKE type_t.chr50,   #VARCHAR(50),    #gzwx007
      host        LIKE type_t.chr50,   #VARCHAR(50)     #gzwx008
      sessionkey  LIKE type_t.chr50           
      END RECORD 
DEFINE g_license_d   DYNAMIC ARRAY OF type_license_d
DEFINE g_gzwl_d_bck      DYNAMIC ARRAY OF type_g_gzwl_d
DEFINE g_log_order       LIKE type_t.num5
DEFINE g_log             STRING  

#181207-00062#2--end--
#CONSTANT ls_version = "191111-00008"                     #191120-00006#1   add
DEFINE ls_version    STRING                              #191120-00006#1   add
DEFINE g_start_time  DATETIME YEAR TO FRACTION(5)        #191120-00006#1   add
DEFINE g_end_time    DATETIME YEAR TO FRACTION(5)        #191120-00006#1   add
#end add-point
 
#add-point:傳入參數說明

#end add-point
 
{</section>}
 
{<section id="azzq080.main" type="m" >}
#+ 此段落由子樣板a26產生
#OPTIONS SHORT CIRCUIT
#+ 作業開始 
MAIN
   #add-point:main段define
 
   #end add-point   
 
   OPTIONS
   INPUT NO WRAP
   DEFER INTERRUPT
   
   #設定SQL錯誤記錄方式 (模組內定義有效)
   WHENEVER ERROR CALL cl_err_msg_log
   LET g_start_time = cl_get_timestamp()        #191120-00006#1   add   
   #依模組進行系統初始化設定(系統設定)
   CALL cl_ap_init("azz","")
 
   #add-point:作業初始化
   #啟動dscom連線, 但先指回原DB, dscom不能跟LIB一樣用完即斷, 因為裡面有temptable相串, 開關會丟
   #191120-00006#1--start--
   CASE
      #查版本
      WHEN g_argv[01] = "-v"  
         #CALL azzq080_version()RETURNING ls_version     #231019-00014#1 mark
         #DISPLAY "INFO:azzq080 版本最新單號:",ls_version #231019-00014#1 mark
         DISPLAY "INFO:azzq080 版本最新版號:",ls_version2 #231019-00014#1 add
         CALL cl_ap_exitprogram("0")
   END CASE
   #191120-00006#1--end--

   LET g_t100debug = FGL_GETENV("T100DEBUG") 
   LET g_dbs_old = g_dbs
   IF NOT cl_db_setconnect("dscom") THEN
      CALL cl_db_connect("dscom", FALSE)       #啟動dscom連線
   END IF
   IF cl_db_setconnect(g_dbs_old) THEN END IF  #先指回原DB
   #end add-point
   
   
 
   #LOCK CURSOR (identifier)
   #add-point:SQL_define

   #end add-point
   LET g_forupd_sql = " ", 
                      " FROM ",
                      " "
   #add-point:SQL_define

   #end add-point
   LET g_forupd_sql = cl_sql_forupd(g_forupd_sql)                #轉換不同資料庫語法
   LET g_forupd_sql = cl_sql_add_mask(g_forupd_sql)              #遮蔽特定資料
   DECLARE azzq080_cl CURSOR FROM g_forupd_sql                 # LOCK CURSOR
 
   LET g_sql = " SELECT  ",
               " FROM  t0",
               
               " WHERE  "
   LET g_sql = cl_sql_add_mask(g_sql)              #遮蔽特定資料
   #add-point:SQL_define

   #end add-point
   PREPARE azzq080_master_referesh FROM g_sql
 
   #add-point:main段define_sql

   #end add-point 
   LET g_forupd_sql = ""
   #add-point:main段define_sql

   #end add-point 
   LET g_forupd_sql = cl_sql_forupd(g_forupd_sql)
   LET g_forupd_sql = cl_sql_add_mask(g_forupd_sql)              #遮蔽特定資料
   DECLARE azzq080_bcl CURSOR FROM g_forupd_sql
 
 
   
   IF g_bgjob = "Y" THEN
      #add-point:Service Call

      #end add-point
   ELSE
      #畫面開啟 (identifier)
      OPEN WINDOW w_azzq080 WITH FORM cl_ap_formpath("azz",g_code)
   
      #瀏覽頁簽資料初始化
      CALL cl_ui_init()
   
      #程式初始化
      CALL azzq080_init()   
 
      #進入選單 Menu (="N")
      CALL azzq080_ui_dialog() 
      
      #add-point:畫面關閉前
      #dscom斷線, 連回原DB, 讓exitprogram可以正常
      IF cl_db_setconnect("dscom") THEN
         DISCONNECT CURRENT
         IF NOT cl_db_setconnect(g_dbs_old) THEN END IF
      END IF
      #end add-point
 
      #畫面關閉
      CLOSE WINDOW w_azzq080
      
   END IF 
   
   CLOSE azzq080_cl
   
   
 
   #add-point:作業離開前

   #end add-point
 
   #離開作業
   CALL cl_ap_exitprogram("0")
   
END MAIN
 
 
 
{</section>}
 
{<section id="azzq080.init" type="s" >}
#+ 畫面資料初始化
PRIVATE FUNCTION azzq080_init()
   #add-point:init段define
   DEFINE ls_title   STRING
   #end add-point   
   
   LET g_error_show  = 1
   LET g_wc_filter   = " 1=1"
   LET g_wc_filter_t = " 1=1"
   
   
   
   #add-point:畫面資料初始化
   LET gwin_curr = ui.Window.getCurrent()
   LET gfrm_curr = gwin_curr.getForm()

   INITIALIZE g_ref_fields TO NULL
   LET g_ref_fields[1] = g_prog
   LET g_ref_fields[2] = "lbl_auth_title"
   CALL ap_ref_array2(g_ref_fields," SELECT gzzd005 FROM gzzd_t WHERE gzzd001 = ? AND gzzd002 = '"||g_lang CLIPPED||"' AND gzzd003 = ? AND gzzd004 = 's'","") RETURNING g_rtn_fields
   LET ls_title = "<span style=\"color: #0054A3; font-size: 14pt; font-weight: bold;\">", g_rtn_fields[1], "</span>"
   CALL gfrm_curr.setElementText("lbl_auth_title", ls_title)
   INITIALIZE g_ref_fields TO NULL
   LET g_ref_fields[1] = g_prog
   LET g_ref_fields[2] = "lbl_process_title"
   CALL ap_ref_array2(g_ref_fields," SELECT gzzd005 FROM gzzd_t WHERE gzzd001 = ? AND gzzd002 = '"||g_lang CLIPPED||"' AND gzzd003 = ? AND gzzd004 = 's'","") RETURNING g_rtn_fields
   LET ls_title = "<span style=\"color: #0054A3; font-size: 14pt; font-weight: bold;\">", g_rtn_fields[1], "</span>"
   CALL gfrm_curr.setElementText("lbl_process_title", ls_title)
   #240628-00031#1 mark --- start ---
#   #231019-00014#1 --start--
#   #設定備註字樣顏色
#   INITIALIZE g_ref_fields TO NULL
#   LET g_ref_fields[1] = g_prog
#   LET g_ref_fields[2] = "lbl_note"
#   CALL ap_ref_array2(g_ref_fields," SELECT gzzd005 FROM gzzd_t WHERE gzzd001 = ? AND gzzd002 = '"||g_lang CLIPPED||"' AND gzzd003 = ? AND gzzd004 = 's'","") RETURNING g_rtn_fields
#   LET ls_title = "<span style=\"color: #D10606; font-size: 10pt; font-weight: bold;\">", g_rtn_fields[1], "</span>"
#   CALL gfrm_curr.setElementText("lbl_note", ls_title)
#   #231019-00014#1 -- end --
   #240628-00031#1 mark --- end ---
   CALL gfrm_curr.setFieldHidden("lbl_note", TRUE) #240628-00031#1 add
   CALL gfrm_curr.setElementText("lbl_note", "")   #240628-00031#1 add
   #240320-00003#1 add --- start ---
   #240320-00003#1 add --- start ---
   INITIALIZE g_ref_fields TO NULL
   LET g_ref_fields[1] = g_prog
   LET g_ref_fields[2] = "lbl_auth_count"
   CALL ap_ref_array2(g_ref_fields," SELECT gzzd005 FROM gzzd_t WHERE gzzd001 = ? AND gzzd002 = '"||g_lang CLIPPED||"' AND gzzd003 = ? AND gzzd004 = 's'","") RETURNING g_rtn_fields
   IF cl_tlm_gl_get_fglwrtinfo("Users_CPU") = "CPU" AND cl_tlm_gl_get_licensenum("G03") = 0 THEN
      LET ls_title = "*", g_rtn_fields[1]
   ELSE
      LET ls_title = g_rtn_fields[1]
   END IF
   CALL gfrm_curr.setElementText("lbl_auth_count", ls_title)
   #240320-00003#1 add --- end ---
   CALL azzq080_add_style()

   #CALL azzq080_license_keyword_define() #240628-00031#1 mark
   #end add-point
 
   CALL azzq080_default_search()  
END FUNCTION
 
{</section>}
 
{<section id="azzq080.default_search" type="s" >}
PRIVATE FUNCTION azzq080_default_search()
   #add-point:default_search段define
   
   #end add-point
 
   #add-point:default_search段開始前
   
   #end add-point
 
   #+ 此段落由子樣板qs27產生
   #+ 組承接外部參數時資料庫欄位對應條件(單身)
   IF NOT cl_null(g_argv[01]) THEN
      LET g_wc = g_wc, " gzwl001 = '", g_argv[01], "' AND "
   END IF
 
 
 
 
 
   IF NOT cl_null(g_wc) THEN
      LET g_wc = g_wc.subString(1,g_wc.getLength()-5)
   ELSE
      #預設查詢條件
      LET g_wc = cl_qbe_get_default_qryplan()
      IF cl_null(g_wc) THEN
         LET g_wc = " 1=1"
      END IF
   END IF
 
   #add-point:default_search段開始後
   #一開始就需要有資料
   #LET g_wc = "gzwl001 IS NOT NULL" #240628-00031#1 mark
   #end add-point
 
END FUNCTION
 
{</section>}
 
{<section id="azzq080.ui_dialog" type="m" >}
#+ 功能選單 
PRIVATE FUNCTION azzq080_ui_dialog()
   DEFINE ls_wc    STRING
   DEFINE li_idx   LIKE type_t.num5
   #add-point:ui_dialog段define
   DEFINE ls_sql   STRING
   DEFINE li_cnt   LIKE type_t.num5
   DEFINE ls_str   STRING  #240111-00016#1 add
   #end add-point 
 
   LET gwin_curr = ui.Window.getCurrent()
   LET gfrm_curr = gwin_curr.getForm()   
   
   LET g_action_choice = " "  
   CALL cl_set_act_visible("accept,cancel", FALSE)
         
   #add-point:ui_dialog段before dialog 
   #240628-00031#1 mark --- start ---
#   #240111-00016#1 add --- start ---
#   #先判斷授權類型再處理欄位隱藏無資料欄位 
#   LET ls_str = cl_tlm_gl_get_fglwrtinfo("Users_CPU")
#   IF ls_str = "CPU" THEN
#      CALL gfrm_curr.setFieldHidden("formonly.lbl_port", TRUE)
#      CALL gfrm_curr.setFieldHidden("formonly.gzwl005_desc", TRUE)
#   END IF
#   #240111-00016#1 add --- end ---
   #240628-00031#1 mark --- end ---
   #240628-00031#1 add --- start ---
   CALL gfrm_curr.setFieldHidden("formonly.gzwl002", TRUE)
   CALL gfrm_curr.setFieldHidden("formonly.gzwl009_desc", TRUE)
   CALL gfrm_curr.setFieldHidden("formonly.lbl_port", TRUE)
   CALL gfrm_curr.setFieldHidden("formonly.gzwl002", TRUE)
   CALL gfrm_curr.setFieldHidden("formonly.gzxa007", TRUE)
   CALL gfrm_curr.setFieldHidden("formonly.gzwl015", TRUE)
   CALL gfrm_curr.setFieldHidden("formonly.pid", TRUE)
   CALL gfrm_curr.setFieldHidden("lbl_process_count", TRUE)
   CALL gfrm_curr.setFieldHidden("lbl_process_limit", TRUE)
   CALL gfrm_curr.setElementText("lbl_process_count", "")
   CALL gfrm_curr.setElementText("lbl_process_limit", "")
   #240628-00031#1 add --- end ---
   #end add-point
 
   LET g_detail_idx = 1  #240628-00031#1 add
   CALL azzq080_b_fill() #240628-00031#1 add
   #240628-00031#1 mark --- start ---
#   IF NOT cl_null(g_wc) AND g_wc != " 1=1" THEN
#      LET g_detail_idx = 1
#      CALL azzq080_b_fill()
#   ELSE
#      CALL azzq080_query()
#   END IF
   #240628-00031#1 mark --- end ---
   
   WHILE TRUE
   
      DIALOG ATTRIBUTES(UNBUFFERED,FIELD ORDER FORM)
         DISPLAY ARRAY g_gzwl_d TO s_detail1.* ATTRIBUTE(COUNT=g_detail_cnt) 
      
            BEFORE DISPLAY 
               LET g_current_page = 1
 
            BEFORE ROW
               LET g_detail_idx = DIALOG.getCurrentRow("s_detail1")
               LET l_ac = g_detail_idx
               DISPLAY g_detail_idx TO FORMONLY.h_index
               DISPLAY g_gzwl_d.getLength() TO FORMONLY.h_count
               CALL azzq080_fetch()
               LET g_master_idx = l_ac
               #add-point:input段before row
 
               #end add-point  
            
            #自訂ACTION(detail_show,page_1)
            
               
         END DISPLAY
      
         DISPLAY ARRAY g_gzwl2_d TO s_detail2.*
            ATTRIBUTES(COUNT=g_detail_cnt2)  
         
            BEFORE DISPLAY 
#               LET g_current_page = 2
            
            BEFORE ROW
               LET g_detail_idx2 = DIALOG.getCurrentRow("s_detail2")
               LET l_ac_d = g_detail_idx2
#               DISPLAY g_detail_idx TO FORMONLY.h_index
#               DISPLAY g_gzwl2_d.getLength() TO FORMONLY.h_count
#               CALL azzq080_fetch()
#               LET g_master_idx = l_ac
               #add-point:input段before row
               #根據授權編號, 重整右側執行清單
               #181207-00062#2--start--
               #CALL azzq080_process_b_fill(g_gzwl2_d[g_detail_idx2].gzwl005)
               #240628-00031#1 mark --- start ----
#               IF g_gzwl2_d[g_detail_idx2].gzwl015 = cl_get_ap_hostname() THEN 
#                  CALL azzq080_process_b_fill(g_gzwl2_d[g_detail_idx2].gzwl005)
#               ELSE 
#                  CALL g_gzwl_d.clear()
#                  CALL azzq080_show_proc_cnt_and_lim(g_gzwl2_d[g_detail_idx2].gzwl005)
#               END IF
               #240628-00031#1 mark --- end ----
               CALL azzq080_process_b_fill_licinfo(g_gzwl2_d[g_detail_idx2].gzwl005_desc) #240628-00031#1 add
               #181207-00062#2--end--
               #end add-point  
 
            #自訂ACTION(detail_show,page_2)
            
               
         END DISPLAY
 
 
         
 
      
         #add-point:ui_dialog段自定義display array

         #end add-point
         
         BEFORE DIALOG      
            CALL DIALOG.setSelectionMode("s_detail1", 1)
 
            #add-point:ui_dialog段before dialog
            LET g_curr_diag = ui.Dialog.getCurrent()
            #不需要標準功能
            CALL DIALOG.setActionActive("insert", FALSE)
            CALL DIALOG.setActionActive("output", FALSE)
            CALL DIALOG.setActionActive("filter", FALSE)
            #end add-point
 
         
         #+ 此段落由子樣板a43產生
         ON ACTION insert
            LET g_action_choice="insert"
            IF cl_auth_chk_act("insert") THEN
               CALL azzq080_insert()
               #add-point:ON ACTION insert

               #END add-point
               EXIT DIALOG
            END IF
 
 
         #+ 此段落由子樣板a43產生
         ON ACTION output
            LET g_action_choice="output"
            IF cl_auth_chk_act("output") THEN
               
               #add-point:ON ACTION output

               #END add-point
               EXIT DIALOG
            END IF
 
 
         #+ 此段落由子樣板a43產生
         ON ACTION query
            LET g_action_choice="query"
            IF cl_auth_chk_act("query") THEN
               CALL azzq080_query()
               #add-point:ON ACTION query

               #END add-point
               
            END IF
 
 
         #+ 此段落由子樣板a43產生
         ON ACTION datainfo
            LET g_action_choice="datainfo"
            IF cl_auth_chk_act("datainfo") THEN
               
               #add-point:ON ACTION datainfo
               #160726-00036#1
               LET ls_sql = "SELECT COUNT(*) FROM ooag_t WHERE ooagent = ? AND ooag001 = ?"
               PREPARE azzq080_ui_dialog_getooag_pre FROM ls_sql
               EXECUTE azzq080_ui_dialog_getooag_pre USING g_enterprise, g_aid[DIALOG.getCurrentRow("s_detail2")].gzwl009 INTO li_cnt
               IF li_cnt > 0 THEN
                  CALL cl_user_contact("aooi130", "ooag_t", "ooag002", "ooag001",g_aid[DIALOG.getCurrentRow("s_detail2")].gzwl009)
               ELSE
                  INITIALIZE g_errparam TO NULL 
                  LET g_errparam.extend = g_gzwl2_d[DIALOG.getCurrentRow("s_detail2")].gzwl002," [",g_aid[DIALOG.getCurrentRow("s_detail2")].gzwl009,"]"
                  LET g_errparam.code   = "azz-00395"  
                  CALL cl_err()
               END IF
               CONTINUE DIALOG   #這樣就不會點完info後，左側的CurrentRow一直跳回最開頭
               #END add-point
               EXIT DIALOG
            END IF
 
 
      
         ON ACTION filter
            LET g_action_choice="filter"
            CALL azzq080_filter()
            #add-point:ON ACTION filter

            #END add-point
            EXIT DIALOG
 
         ON ACTION close
            LET INT_FLAG=FALSE         
            LET g_action_choice = "exit"
            EXIT DIALOG
 
         ON ACTION exit
            LET g_action_choice="exit"
            EXIT DIALOG
 
         ON ACTION datarefresh   # 重新整理
            LET g_error_show = 1
            CALL azzq080_b_fill()
 
         ON ACTION agendum   # 待辦事項
            #add-point:ON ACTION agendum

            #END add-point
            CALL cl_user_overview()
 
         #+ 此段落由子樣板qs19產生
         #有關於sel欄位選取的action段落
         #選擇全部
         ON ACTION selall
            CALL DIALOG.setSelectionRange("s_detail1", 1, -1, 1)
            FOR li_idx = 1 TO g_gzwl_d.getLength()
               LET g_gzwl_d[li_idx].sel = "Y"
            END FOR
 
            #add-point:ui_dialog段on action selall

            #end add-point
 
         #取消全部
         ON ACTION selnone
            CALL DIALOG.setSelectionRange("s_detail1", 1, -1, 0)
            FOR li_idx = 1 TO g_gzwl_d.getLength()
               LET g_gzwl_d[li_idx].sel = "N"
            END FOR
 
            #add-point:ui_dialog段on action selnone

            #end add-point
 
         #勾選所選資料
         ON ACTION sel
            FOR li_idx = 1 TO g_gzwl_d.getLength()
               IF DIALOG.isRowSelected("s_detail1", li_idx) THEN
                  LET g_gzwl_d[li_idx].sel = "Y"
               END IF
            END FOR
 
            #add-point:ui_dialog段on action sel

            #end add-point
 
         #取消所選資料
         ON ACTION unsel
            FOR li_idx = 1 TO g_gzwl_d.getLength()
               IF DIALOG.isRowSelected("s_detail1", li_idx) THEN
                  LET g_gzwl_d[li_idx].sel = "N"
               END IF
            END FOR
 
            #add-point:ui_dialog段on action unsel

            #end add-point
 
 
 
         
 
         #add-point:ui_dialog段自定義action
         ON ACTION killprocess
            CALL azzq080_kill_process()

         ON ACTION licenseinfo_view
            CALL azzq080_fglwrt_display()
         #220801-00038#1 add s
         ON ACTION licenseusage
            CALL azzq080_license_usage_dialog()
            #CALL azzq080_export_usage("2022-08-11", "0") 
         #220801-00038#1 add e
         
         #170814-00017#1--add--start--
         #excel匯出功能          
         ON ACTION exporttoexcel
            LET g_action_choice="exporttoexcel"
            IF cl_auth_chk_act("exporttoexcel") THEN
               #browser
               CALL g_export_node.clear()

               LET g_export_node[1] = base.typeInfo.create(g_gzwl_d)
               LET g_export_id[1]   = "s_detail1"
               LET g_export_node[2] = base.typeInfo.create(g_gzwl2_d)
               LET g_export_id[2]   = "s_detail2"
                
               #add-point:ON ACTION exporttoexcel name="menu.exporttoexcel"
                
               #END add-point
               CALL cl_export_to_excel_getpage()
               CALL cl_export_to_excel()
            END IF
         #170814-00017#1--add--end--   
         #end add-point
      
         #主選單用ACTION
         &include "main_menu.4gl"
         &include "relating_action.4gl"
         #交談指令共用ACTION
         &include "common_action.4gl"
            CONTINUE DIALOG
 
         #add-point:查詢方案相關ACTION設定前
 
         #end add-point
 
         ON ACTION qbe_select
            CALL cl_qbe_list("m") RETURNING ls_wc
            IF NOT cl_null(ls_wc) THEN
               LET g_wc = ls_wc
               CALL azzq080_b_fill()
            END IF
      
      #  ON ACTION qbeclear   # 條件清除
      #     CLEAR FORM
 
         #add-point:查詢方案相關ACTION設定後

         #end add-point
      END DIALOG
      
      IF g_action_choice = "exit" AND NOT cl_null(g_action_choice) THEN
         EXIT WHILE
      END IF
      
   END WHILE
 
   CALL cl_set_act_visible("accept,cancel", TRUE)
 
END FUNCTION
 
{</section>}
 
{<section id="azzq080.query" type="m" >}
#+ QBE資料查詢
PRIVATE FUNCTION azzq080_query()
   DEFINE ls_wc      LIKE type_t.chr500
   DEFINE ls_return  STRING
   DEFINE ls_result  STRING 
   #add-point:query段define

   #end add-point 
   
   LET INT_FLAG = 0
   CLEAR FORM
   CALL g_gzwl_d.clear()
   LET g_wc_filter = " 1=1"
   
   CALL gfrm_curr.setFieldHidden("formonly.sel", TRUE)
   CALL gfrm_curr.setFieldHidden("formonly.statepic", TRUE)
   
   LET g_qryparam.state = "c"
   LET g_detail_idx  = 1
   LET g_detail_idx2 = 1
   
   #wc備份
   LET ls_wc = g_wc
   LET g_master_idx = l_ac
 
   
 
   DIALOG ATTRIBUTES(UNBUFFERED,FIELD ORDER FORM)
 
      #單身根據table分拆construct
      #240628-00031#1 mark --- start ---
#      CONSTRUCT g_wc_table ON gzwl001,gzwl003,gzwl004,gzwl007,gzwl014,gzwl008
#           FROM s_detail1[1].b_gzwl001,s_detail1[1].b_gzwl003,s_detail1[1].b_gzwl004,s_detail1[1].b_gzwl007, 
#               s_detail1[1].b_gzwl014,s_detail1[1].b_gzwl008
      #240628-00031#1 mark --- end ---  
         #240628-00031#1 add --- start ---
         CONSTRUCT g_wc_table ON gzwl002,gzwl009,gzwl002_desc,gzwl006,gzwl010,gzwl003,gzwl003_desc,gzwl004
              FROM s_detail1[1].b_gzwl002,s_detail1[1].b_gzwl009,s_detail1[1].b_gzwl009_desc,s_detail1[1].b_gzwl006, 
                  s_detail1[1].b_gzwl010,s_detail1[1].b_gzwl003,s_detail1[1].b_gzwl003_desc,s_detail1[1].b_gzwl004
         #240628-00031#1 add --- end ---
                      
         BEFORE CONSTRUCT
            #add-point:cs段more_construct

            #end add-point 
            
       #單身公用欄位開窗相關處理
       
         
       #單身一般欄位開窗相關處理
                #----<<sel>>----
         #----<<b_gzwl001>>----
         #此段落由子樣板a01產生
         #BEFORE FIELD b_gzwl001 #251028-00009#1 mark
            #add-point:BEFORE FIELD b_gzwl001

            #END add-point
 
         #此段落由子樣板a02產生
         #AFTER FIELD b_gzwl001 #251028-00009#1 mark
            
            #add-point:AFTER FIELD b_gzwl001

            #END add-point
            
 
         #Ctrlp:construct.c.page1.b_gzwl001
         #ON ACTION controlp INFIELD b_gzwl001 #251028-00009#1 mark
            #add-point:ON ACTION controlp INFIELD b_gzwl001

            #END add-point
 
         #----<<b_gzwl003>>----
         #此段落由子樣板a01產生
         BEFORE FIELD b_gzwl003
            #add-point:BEFORE FIELD b_gzwl003

            #END add-point
 
         #此段落由子樣板a02產生
         AFTER FIELD b_gzwl003
            
            #add-point:AFTER FIELD b_gzwl003

            #END add-point
            
 
         #Ctrlp:construct.c.page1.b_gzwl003
         ON ACTION controlp INFIELD b_gzwl003 
            #add-point:ON ACTION controlp INFIELD b_gzwl003
            #170814-00017#1--add--start--
            INITIALIZE g_qryparam.* TO NULL
            LET g_qryparam.state = 'c'
            LET g_qryparam.reqry = FALSE
            CALL q_gzwl003()                              #呼叫開窗
            DISPLAY g_qryparam.return1 TO b_gzwl003       #顯示到畫面上
            DISPLAY g_qryparam.return2 TO b_gzwl003_desc  #顯示到畫面上

            NEXT FIELD b_gzwl003                          #返回原欄位
            #170814-00017#1--add--end--
            #END add-point
 
         #----<<b_gzwl003_desc>>----
         #----<<b_gzwl004>>----
         #此段落由子樣板a01產生
         BEFORE FIELD b_gzwl004
            #add-point:BEFORE FIELD b_gzwl004

            #END add-point
 
         #此段落由子樣板a02產生
         AFTER FIELD b_gzwl004
            
            #add-point:AFTER FIELD b_gzwl004

            #END add-point
            
 
         #Ctrlp:construct.c.page1.b_gzwl004
         ON ACTION controlp INFIELD b_gzwl004
            #add-point:ON ACTION controlp INFIELD b_gzwl004

            #END add-point
 
         #----<<b_gzwl007>>----
         #此段落由子樣板a01產生
         #BEFORE FIELD b_gzwl007 #251028-00009#1 mark
            #add-point:BEFORE FIELD b_gzwl007

            #END add-point
 
         #此段落由子樣板a02產生
         #AFTER FIELD b_gzwl007 #251028-00009#1 mark
            
            #add-point:AFTER FIELD b_gzwl007

            #END add-point
            
 
         #Ctrlp:construct.c.page1.b_gzwl007
         #ON ACTION controlp INFIELD b_gzwl007 #251028-00009#1 mark
            #add-point:ON ACTION controlp INFIELD b_gzwl007

            #END add-point
 
         #----<<b_gzwl008>>----
         #此段落由子樣板a01產生
         #BEFORE FIELD b_gzwl008 #251028-00009#1 mark
            #add-point:BEFORE FIELD b_gzwl008

            #END add-point
 
         #此段落由子樣板a02產生
         #AFTER FIELD b_gzwl008 #251028-00009#1 mark
            
            #add-point:AFTER FIELD b_gzwl008

            #END add-point
            
 
         #Ctrlp:construct.c.page1.b_gzwl008
         #ON ACTION controlp INFIELD b_gzwl008 #251028-00009#1 mark
            #add-point:ON ACTION controlp INFIELD b_gzwl008

            #END add-point
 
         #----<<gzwl005>>----
         #----<<gzwl002>>----
         #----<<gzwl002_desc>>----
         #----<<gzwl006>>----
   
       
      END CONSTRUCT
      
 
      
 
  
      #add-point:query段more_construct
      #160726-00036#1
#      CONSTRUCT g_wc_table2 ON gzwl002,gzwl006            #240628-00031#1 mark
#           FROM s_detail2[1].gzwl002,s_detail2[1].gzwl006 #240628-00031#1 mark
          CONSTRUCT g_wc_table2 ON gzwl006 #240628-00031#1 add 
             FROM s_detail2[1].gzwl006     #240628-00031#1 add 
          
         #170814-00017#1--add--start--     
         #251028-00009#1 mark --- start ---         
#         ON ACTION controlp INFIELD gzwl002
#            INITIALIZE g_qryparam.* TO NULL
#            LET g_qryparam.state = 'c'
#            LET g_qryparam.reqry = FALSE
#            CALL q_gzwl002()                            #呼叫開窗
#            DISPLAY g_qryparam.return1 TO gzwl002       #顯示到畫面上
#            DISPLAY g_qryparam.return2 TO gzwl009_desc  #顯示到畫面上
#            DISPLAY g_qryparam.return3 TO gzwl006       #顯示到畫面上
#
#            NEXT FIELD gzwl002                          #返回原欄位
         #251028-00009#1 mark --- end ---

         ON ACTION controlp INFIELD gzwl006
            INITIALIZE g_qryparam.* TO NULL
            LET g_qryparam.state = 'c'
            LET g_qryparam.reqry = FALSE
            CALL q_gzwl002()                            #呼叫開窗
            DISPLAY g_qryparam.return1 TO gzwl002       #顯示到畫面上
            DISPLAY g_qryparam.return2 TO gzwl009_desc  #顯示到畫面上
            DISPLAY g_qryparam.return3 TO gzwl006       #顯示到畫面上

            NEXT FIELD gzwl006                          #返回原欄位
         #170814-00017#1--add--end--
      END CONSTRUCT
      #end add-point 
 
      ON ACTION accept
         ACCEPT DIALOG
         
      ON ACTION cancel
         LET INT_FLAG = 1
         EXIT DIALOG
      
      #交談指令共用ACTION
      &include "common_action.4gl"
         CONTINUE DIALOG 
 
      #add-point:query段查詢方案相關ACTION設定前

      #end add-point 
 
      #查詢方案列表
      ON ACTION qbe_select
         LET ls_wc = ""
         CALL cl_qbe_list("c") RETURNING ls_wc
 
      #條件儲存為方案
      ON ACTION qbe_save
         CALL cl_qbe_save()
 
      ON ACTION qbeclear   # 條件清除
         CLEAR FORM
 
      #add-point:query段查詢方案相關ACTION設定後

      #end add-point 
 
   END DIALOG
 
   
 
   IF INT_FLAG THEN
      LET INT_FLAG = 0
      #還原
      LET g_wc = ls_wc
   ELSE
      LET g_master_idx = 1
   END IF
   
   LET g_wc = g_wc_table 
 
 
        
   LET g_wc2 = " 1=1"
 
        
   #add-point:cs段after_construct
   #LET g_wc = g_wc, " AND ", g_wc_table2    #170814-00017#1 add #250901-00010#1 mark
   #250901-00010#1 add --- start --- 
   IF cl_null(g_wc_table) THEN
      LET g_wc = " 1=1 " 
   END IF
   IF NOT cl_null(g_wc_table2) THEN
      LET g_wc = g_wc, " AND ", g_wc_table2 
   END IF
   #250901-00010#1 --- end ---    
   #160726-00036#1
   IF g_wc_table = " 1=1" AND g_wc_table2 = " 1=1" THEN
      #LET g_wc = "gzwl001 IS NOT NULL"     #250709-00026#1 mark
   END IF
   #end add-point
        
   LET g_error_show = 1
   CALL azzq080_b_fill()
   LET l_ac = g_master_idx
   IF g_detail_cnt = 0 AND NOT INT_FLAG THEN
      INITIALIZE g_errparam TO NULL 
      LET g_errparam.extend = "" 
      LET g_errparam.code   = -100 
      LET g_errparam.popup  = TRUE 
      CALL cl_err()
 
   END IF
   
   CALL gfrm_curr.setFieldHidden("formonly.sel", FALSE)
   CALL gfrm_curr.setFieldHidden("formonly.statepic", FALSE)
   
END FUNCTION
 
{</section>}
 
{<section id="azzq080.b_fill" type="m" >}
#+ 單身陣列填充
PRIVATE FUNCTION azzq080_b_fill()
   DEFINE ls_wc           STRING
   #add-point:b_fill段define
   DEFINE ls_sql          STRING
   DEFINE ls_dbs          STRING
   DEFINE lc_gzwl009      LIKE gzwl_t.gzwl009
   DEFINE lc_gzwl012      LIKE gzwl_t.gzwl012
   DEFINE lc_gzou003      LIKE gzou_t.gzou003
   DEFINE ls_str          STRING  #240111-00016#1 add
   #end add-point
 
   #add-point:b_fill段sql_before
   #抓取清單前, 先整理PID使用情況
   #預計是透過森雄的工具重整

   #240111-00016#1 mark --- start ---
   #必須先處理左側授權列表, 
   #CALL azzq080_license_collect()
   #顯示上方人數資訊
   #IF g_auth_cnt = 32767 THEN
   #   DISPLAY g_detail_cnt2, "∞" TO FORMONLY.auth_count, FORMONLY.auth_total
   #ELSE
   #   DISPLAY g_detail_cnt2, g_auth_cnt TO FORMONLY.auth_count, FORMONLY.auth_total
   #END IF
   #240111-00016#1 mark --- end ---
   #240628-00031#1 mark --- start ---
#   #240111-00016#1 add --- start ---
#   #先判斷授權類型再處理顯示
#   LET ls_str = cl_tlm_gl_get_fglwrtinfo("Users_CPU")
#   IF ls_str = "CPU" THEN
#      #必須先處理左側授權列表 
#      CALL azzq080_license_collect_TUE()
#      #顯示上方人數資訊
#      #整理資訊要用到全體gzwl_t, 需先連結dscom, 處理完後才連回原DB
#      IF cl_db_setconnect("dscom") THEN END IF
#      LET g_auth_cnt = cl_tlm_gl_get_licensenum("TUE")
#      SELECT COUNT(gzwl006) INTO g_detail_cnt2
#         FROM ( SELECT DISTINCT gzwl006 FROM gzwl_t WHERE TRIM(gzwl006) IS NOT NULL ) A
#      DISPLAY g_detail_cnt2, g_auth_cnt TO FORMONLY.auth_count, FORMONLY.auth_total
#      #最後連回原來DB
#      IF cl_db_setconnect(g_dbs_old) THEN END IF 
#   ELSE
#      #必須先處理左側授權列表, 
#      CALL azzq080_license_collect()
#      #顯示上方人數資訊
#      IF g_auth_cnt = 32767 THEN
#         DISPLAY g_detail_cnt2, "∞" TO FORMONLY.auth_count, FORMONLY.auth_total
#      ELSE
#         DISPLAY g_detail_cnt2, g_auth_cnt TO FORMONLY.auth_count, FORMONLY.auth_total
#      END IF
#   END IF
#   #240111-00016#1 add --- end ---
   #240628-00031#1 mark --- end ---
   #240628-00031#1 add --- start ---
   #先判斷授權類型再處理顯示
   #必須先處理左側授權列表, 
   CALL azzq080_license_collect()
   #顯示上方人數資訊
   IF g_auth_cnt = 32767 THEN
      DISPLAY g_detail_cnt2, "∞" TO FORMONLY.auth_count, FORMONLY.auth_total
   ELSE
      DISPLAY g_detail_cnt2, g_auth_cnt TO FORMONLY.auth_count, FORMONLY.auth_total
   END IF
   #240628-00031#1 add --- end ---
   
   #在檢核一次 PID 存在否                #181207-00062#2  add
   #CALL azzq080_chk_g_gzwl2_d_data()    #181207-00062#2  add   #191111-00008#1  mark 不檢查了避免開啟畫面時間過久

   IF g_gzwl2_d.getLength() > 0 THEN
      CALL g_curr_diag.setCurrentRow("s_detail2", 1)
      LET g_detail_idx2 = g_curr_diag.getCurrentRow("s_detail2")
      IF g_detail_idx2 = 0 OR cl_null(g_detail_idx2) THEN
         LET g_detail_idx2 = 1
      END IF
      #240628-00031#1 mark --- start ----
#      #181207-00062#2--start--
#      #CALL azzq080_process_b_fill(g_gzwl2_d[g_detail_idx2].gzwl005
#      IF g_gzwl2_d[g_detail_idx2].gzwl015 = cl_get_ap_hostname() THEN 
#         CALL azzq080_process_b_fill(g_gzwl2_d[g_detail_idx2].gzwl005)
#      ELSE 
#         CALL g_gzwl_d.clear()
#         CALL azzq080_show_proc_cnt_and_lim(g_gzwl2_d[g_detail_idx2].gzwl005)
#      END IF
#      #181207-00062#2--end--
      #240628-00031#1 mark --- end ----
      CALL azzq080_process_b_fill_licinfo(g_gzwl2_d[g_detail_idx2].gzwl005_desc) #240628-00031#1 add
   END IF
   #不繼續往下執行
   RETURN

   IF cl_null(g_wc_filter) THEN
      LET g_wc_filter = " 1=1"
   END IF
   IF cl_null(g_wc) THEN
      LET g_wc = " 1=1"
   END IF

   LET ls_sql = "SELECT gzou003 FROM gzou_t WHERE gzou001 = ?"
   PREPARE b_fill_getdbs_pre FROM ls_sql
#   LET g_sql = "SELECT DISTINCT gzwl005, gzwl002, ooag011, gzwl006 FROM gzwl_t",
#               " LEFT OUTER JOIN ooag_t ON ooagent = gzwl012 AND ooag001 = gzwl009",
#               " WHERE ", g_wc, " AND ", g_wc_filter,cl_sql_add_filter("gzwl_t"),
#               " ORDER BY gzwl_t.gzwl002" #ODI-ORA#
   LET ls_sql = "SELECT DISTINCT gzwl005, gzwl002, '', gzwl006, gzwl009, gzwl012 FROM gzwl_t",
                " WHERE ", g_wc, " AND ", g_wc_filter,cl_sql_add_filter("gzwl_t"),
                " ORDER BY gzwl_t.gzwl002" #ODI-ORA#
   LET ls_sql = cl_sql_add_mask(ls_sql)              #遮蔽特定資料
   #LET ls_sql = ls_sql,cl_sql_add_filter("gzwl_t")  #190926-00017#2 add   #191111-00008#1   mark 上面已經有 cl_sql_add_filter 、cl_sql_add_mask 所以不用呼叫 
   #LET ls_sql = cl_sql_add_mask(ls_sql)             #190926-00017#2 add   #191111-00008#1   mark 上面已經有 cl_sql_add_filter 、cl_sql_add_mask 所以不用呼叫 
   PREPARE azzq080_pb2 FROM ls_sql
   DECLARE b2_fill_curs CURSOR FOR azzq080_pb2
   CALL g_gzwl2_d.clear()
   LET l_ac_d = 1
   FOREACH b2_fill_curs INTO g_gzwl2_d[l_ac_d].gzwl005, g_gzwl2_d[l_ac_d].gzwl002,
                             g_gzwl2_d[l_ac_d].gzwl002_desc, g_gzwl2_d[l_ac_d].gzwl006,
                             lc_gzwl009, lc_gzwl012
      #授權先全抓, 應該尚未大量到開不出來
      IF cl_null(lc_gzwl012) THEN
         LET lc_gzwl012 = g_enterprise
         LET lc_gzou003 = g_dbs
      ELSE
         INITIALIZE lc_gzou003 TO NULL
         EXECUTE b_fill_getdbs_pre INTO lc_gzou003 USING lc_gzwl012
         IF cl_null(lc_gzou003) THEN
            LET lc_gzou003 = g_dbs
         END IF
      END IF
      LET ls_sql = "SELECT ooag011 FROM ", lc_gzou003 CLIPPED,".ooag_t WHERE ooagent = ? AND ooag001 = ?"
      PREPARE b_fill_getname_pre FROM ls_sql
      EXECUTE b_fill_getname_pre INTO g_gzwl2_d[l_ac_d].gzwl002_desc USING lc_gzwl012, lc_gzwl009
      LET g_gzwl2_d[l_ac_d].num = l_ac_d                 #200827-00023#1 add
      LET l_ac_d = l_ac_d + 1
   END FOREACH
   #end add-point
 
#   IF cl_null(g_wc_filter) THEN
#      LET g_wc_filter = " 1=1"
#   END IF
#   IF cl_null(g_wc) THEN
#      LET g_wc = " 1=1"
#   END IF
#   IF cl_null(g_wc2) THEN
#      LET g_wc2 = " 1=1"
#   END IF
#   
#   LET ls_wc = g_wc, " AND ", g_wc2, " AND ", g_wc_filter
#   
#   LET g_sql = "SELECT  UNIQUE '',gzwl001,gzwl003,'',gzwl004,gzwl007,gzwl014,gzwl008,'','','','' FROM gzwl_t", 
#
# 
# 
#               "",
#               " WHERE 1=1 AND ", ls_wc,cl_sql_add_filter("gzwl_t")
#    
#   LET g_sql = g_sql, cl_sql_add_filter("gzwl_t"),
#                      " ORDER BY gzwl_t.gzwl001" #ODI-ORA#
#  
   #add-point:b_fill段sql_after

   #end add-point
 
#   LET g_sql = cl_sql_add_mask(g_sql)              #遮蔽特定資料
#   PREPARE azzq080_pb FROM g_sql
#   DECLARE b_fill_curs CURSOR FOR azzq080_pb
#   
#   OPEN b_fill_curs
# 
#   CALL g_gzwl_d.clear()
#   CALL g_gzwl2_d.clear()   
 
 
   #add-point:陣列清空
 
   #end add-point
 
#   LET g_cnt = l_ac
#   LET l_ac = 1   
#   ERROR "Searching!" 
# 
#   FOREACH b_fill_curs INTO g_gzwl_d[l_ac].sel,g_gzwl_d[l_ac].gzwl001,g_gzwl_d[l_ac].gzwl003,g_gzwl_d[l_ac].gzwl003_desc, 
#       g_gzwl_d[l_ac].gzwl004,g_gzwl_d[l_ac].gzwl007,g_gzwl_d[l_ac].gzwl014,g_gzwl_d[l_ac].gzwl008,g_gzwl2_d[l_ac].gzwl005, 
#       g_gzwl2_d[l_ac].gzwl002,g_gzwl2_d[l_ac].gzwl002_desc,g_gzwl2_d[l_ac].gzwl006
#      IF SQLCA.sqlcode THEN
#         INITIALIZE g_errparam TO NULL 
#         LET g_errparam.extend = "FOREACH:" 
#         LET g_errparam.code   = SQLCA.sqlcode 
#         LET g_errparam.popup  = TRUE 
#         CALL cl_err()
# 
#         EXIT FOREACH
#      END IF
      
      #LET g_gzwl_d[l_ac].statepic = cl_get_actipic(g_gzwl_d[l_ac].statepic)
 
      
 
      #add-point:b_fill段資料填充

      #end add-point
      
#      CALL azzq080_detail_show("'1'")      
 
      #同一授權process列表全部顯示
#      IF l_ac > g_max_rec THEN
#         IF g_error_show = 1 THEN
#            INITIALIZE g_errparam TO NULL 
#            LET g_errparam.extend =  "" 
#            LET g_errparam.code   =  9035 
#            LET g_errparam.popup  = TRUE 
#            CALL cl_err()
# 
#         END IF
#         EXIT FOREACH
#      END IF
#      LET l_ac = l_ac + 1
#      
#   END FOREACH
#   LET g_error_show = 0
#   
# 
#   
#   CALL g_gzwl_d.deleteElement(g_gzwl_d.getLength())   
#   CALL g_gzwl2_d.deleteElement(g_gzwl2_d.getLength())
 
 
   #add-point:陣列長度調整

   #end add-point
   
   #add-point:b_fill段資料填充(其他單身)
 
   #end add-point
#   LET g_detail_cnt = g_gzwl_d.getLength()
#   DISPLAY g_detail_cnt TO FORMONLY.h_count
#   LET l_ac = g_cnt
#   LET g_cnt = 0
#   
#   CLOSE b_fill_curs
#   FREE azzq080_pb
#   
#   LET l_ac = 1
#   IF g_gzwl_d.getLength() > 0 THEN
#      CALL azzq080_fetch()
#   END IF
   
      CALL azzq080_filter_show('gzwl001','b_gzwl001')
   CALL azzq080_filter_show('gzwl003','b_gzwl003')
   CALL azzq080_filter_show('gzwl004','b_gzwl004')
   CALL azzq080_filter_show('gzwl007','b_gzwl007')
   CALL azzq080_filter_show('gzwl014','b_gzwl014')
   CALL azzq080_filter_show('gzwl008','b_gzwl008')
 
   
END FUNCTION
 
{</section>}
 
{<section id="azzq080.fetch" type="s" >}
#+ 單身陣列填充2
PRIVATE FUNCTION azzq080_fetch()
   DEFINE li_ac           LIKE type_t.num5
   #add-point:fetch段define
   
   #end add-point
   
 
   #add-point:陣列清空
   
   #end add-point
   
   LET li_ac = l_ac 
   
 
   
   #add-point:單身填充後
   
   #end add-point 
   
 
   #add-point:陣列筆數調整
   
   #end add-point
 
   LET l_ac = li_ac
   
END FUNCTION
 
{</section>}
 
{<section id="azzq080.detail_show" type="s" >}
#+ 顯示相關資料
PRIVATE FUNCTION azzq080_detail_show(ps_page)
   DEFINE ps_page    STRING
   DEFINE ls_sql     STRING
   #add-point:show段define
   
   #end add-point
 
   #add-point:detail_show段之前
   
   #end add-point
   
   
 
   #讀入ref值
   IF ps_page.getIndexOf("'1'",1) > 0 THEN
      #帶出公用欄位reference值page1
      
 
      #add-point:show段單身reference
      
      #end add-point
   END IF
   
   IF ps_page.getIndexOf("'2'",1) > 0 THEN
      #帶出公用欄位reference值page2
      
 
      #add-point:show段單身reference
      
      #end add-point
   END IF
 
 
 
   #add-point:detail_show段之後
   
   #end add-point
 
END FUNCTION
 
{</section>}
 
{<section id="azzq080.filter" type="m" >}
#+ filter過濾功能
PRIVATE FUNCTION azzq080_filter()
   #add-point:filter段define

   #end add-point    
   
   LET l_ac = 1
   LET g_detail_idx = 1
   LET g_detail_idx2 = 1
 
   LET INT_FLAG = 0
 
   LET g_qryparam.state = 'c'
 
   LET g_wc_filter_t = g_wc_filter
   LET g_wc_t = g_wc
   
   CALL gfrm_curr.setFieldHidden("formonly.sel", TRUE)
   CALL gfrm_curr.setFieldHidden("formonly.statepic", TRUE)
 
   LET g_wc = cl_replace_str(g_wc, g_wc_filter, '')
 
   #使用DIALOG包住 單頭CONSTRUCT及單身CONSTRUCT
   DIALOG ATTRIBUTES(UNBUFFERED,FIELD ORDER FORM)
 
      #單頭
      CONSTRUCT g_wc_filter ON gzwl001,gzwl003,gzwl004,gzwl007,gzwl014,gzwl008
                          FROM s_detail1[1].b_gzwl001,s_detail1[1].b_gzwl003,s_detail1[1].b_gzwl004, 
                              s_detail1[1].b_gzwl007,s_detail1[1].b_gzwl014,s_detail1[1].b_gzwl008
 
         BEFORE CONSTRUCT
#saki       CALL cl_qbe_init()
                     DISPLAY azzq080_filter_parser('gzwl001') TO s_detail1[1].b_gzwl001
            DISPLAY azzq080_filter_parser('gzwl003') TO s_detail1[1].b_gzwl003
            DISPLAY azzq080_filter_parser('gzwl004') TO s_detail1[1].b_gzwl004
            DISPLAY azzq080_filter_parser('gzwl007') TO s_detail1[1].b_gzwl007
            DISPLAY azzq080_filter_parser('gzwl014') TO s_detail1[1].b_gzwl014
            DISPLAY azzq080_filter_parser('gzwl008') TO s_detail1[1].b_gzwl008
 
 
      END CONSTRUCT
 
      #add-point:filter段add_cs

      #end add-point
 
      BEFORE DIALOG
         #add-point:filter段b_dialog

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
      LET g_wc_filter = g_wc_filter, " "
   ELSE
      LET g_wc_filter = g_wc_filter_t
   END IF
   
      CALL azzq080_filter_show('gzwl001','b_gzwl001')
   CALL azzq080_filter_show('gzwl003','b_gzwl003')
   CALL azzq080_filter_show('gzwl004','b_gzwl004')
   CALL azzq080_filter_show('gzwl007','b_gzwl007')
   CALL azzq080_filter_show('gzwl014','b_gzwl014')
   CALL azzq080_filter_show('gzwl008','b_gzwl008')
 
    
   CALL azzq080_b_fill()
   
   CALL gfrm_curr.setFieldHidden("formonly.sel", FALSE)
   CALL gfrm_curr.setFieldHidden("formonly.statepic", FALSE)
 
END FUNCTION
 
{</section>}
 
{<section id="azzq080.filter_parser" type="s" >}
#+ filter欄位解析
PRIVATE FUNCTION azzq080_filter_parser(ps_field)
   DEFINE ps_field   STRING
   DEFINE ls_tmp     STRING
   DEFINE li_tmp     LIKE type_t.num5
   DEFINE li_tmp2    LIKE type_t.num5
   DEFINE ls_var     STRING
   #add-point:filter段define
   
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
 
{<section id="azzq080.filter_show" type="s" >}
#+ Browser標題欄位顯示搜尋條件
PRIVATE FUNCTION azzq080_filter_show(ps_field,ps_object)
   DEFINE ps_field         STRING
   DEFINE ps_object        STRING
   DEFINE lnode_item       om.DomNode
   DEFINE ls_title         STRING
   DEFINE ls_name          STRING
   DEFINE ls_condition     STRING
 
   LET ls_name = "formonly.", ps_object
 
   LET lnode_item = gfrm_curr.findNode("TableColumn", ls_name)
   LET ls_title = lnode_item.getAttribute("text")
   IF ls_title.getIndexOf('※',1) > 0 THEN
      LEt ls_title = ls_title.subString(1,ls_title.getIndexOf('※',1)-1)
   END IF
 
   #顯示資料組合
   LET ls_condition = azzq080_filter_parser(ps_field)
   IF NOT cl_null(ls_condition) THEN
      LET ls_title = ls_title, '※', ls_condition, '※'
   END IF
 
   #將資料顯示回去
   CALL lnode_item.setAttribute("text",ls_title)
 
END FUNCTION
 
{</section>}
 
{<section id="azzq080.insert" type="s" >}
#+ insert
PRIVATE FUNCTION azzq080_insert()
   #add-point:insert段define
   
   #end add-point
 
   #add-point:insert段control
   
   #end add-point    
   
END FUNCTION
 
{</section>}
 
{<section id="azzq080.modify" type="s" >}
#+ modify
PRIVATE FUNCTION azzq080_modify()
   #add-point:modify段define
   
   #end add-point
 
   #add-point:modify段control
   
   #end add-point 
END FUNCTION
 
{</section>}
 
{<section id="azzq080.reproduce" type="s" >}
#+ reproduce
PRIVATE FUNCTION azzq080_reproduce()
   #add-point:reproduce段define
   
   #end add-point
 
   #add-point:reproduce段control
   
   #end add-point 
END FUNCTION
 
{</section>}
 
{<section id="azzq080.delete" type="s" >}
#+ delete
PRIVATE FUNCTION azzq080_delete()
   #add-point:delete段define
   
   #end add-point
 
   #add-point:delete段control
   
   #end add-point 
END FUNCTION
 
{</section>}
 
{<section id="azzq080.other_function" type="s" >}
################################################################################
# Descriptions...: 抓取指定授權下的目前執行清單
# Memo...........: 
# Usage..........: CALL azzq080_process_b_fill(pc_gzwl005)
# Input parameter: pc_gzwl005   授權編號
# Return code....: None
# Date & Author..: 2014/09/22 By Saki
# Modify.........:
################################################################################
PUBLIC FUNCTION azzq080_process_b_fill(pc_gzwl005)
   DEFINE pc_gzwl005   LIKE gzwl_t.gzwl005
#181207-00062#2--strt--   
#   DEFINE lr_others    RECORD
#             pid       LIKE type_t.chr20,
#             exemode   LIKE type_t.chr20,
#             agent     LIKE type_t.chr20,
#             appname   LIKE type_t.chr50,
#             host      LIKE type_t.chr50
#                       END RECORD

   DEFINE lr_others    type_license_d
   DEFINE lc_gzwl002   LIKE gzwl_t.gzwl002
   DEFINE ls_ent       STRING        
#181207-00062#2--end--   
   DEFINE li_i         LIKE type_t.num5
   DEFINE li_repeat    BOOLEAN
   DEFINE lc_channel   base.Channel
   DEFINE ls_cmd       STRING
   DEFINE ls_str       STRING

   #組出右側第一筆資料的where條件
   IF cl_db_setconnect("dscom") THEN END IF
   #181117-00002#1--start--
                                #sel、Session Key、企業編號、作業編號
   #LET g_sql = "SELECT  UNIQUE '',gzwl001,gzwl003,'',gzwl004,gzwl015,gzwl008,gzwl010,gzwl007,gzwl012,gzwl014,gzwl013 FROM gzwl_t",
   #LET g_sql = "SELECT  UNIQUE '',gzwl001,gzwl012,gzwl003,'',gzwl004,gzwl015,gzwl008,gzwl010,gzwl007,gzwl014,gzwl013 FROM gzwl_t",     #181207-00062#2 mark
   #181117-00002#1--end--
   #LET g_sql = "SELECT  UNIQUE '',gzwl001,gzwl010,gzwl012,gzwl003,'',gzwl004,gzwl015,gzwl008,gzwl007,gzwl014,gzwl013,gzwl006 FROM gzwl_t",      #181207-00062#2 add #240628-00031#1 mark 
   LET g_sql = "SELECT  UNIQUE '',gzwl001,gzwl010,gzwl012,gzwl003,'',gzwl004,gzwl015,gzwl008,gzwl007,gzwl014,gzwl013,gzwl006,gzwl002 FROM gzwl_t",      #240628-00031#1 add  
                " WHERE ", g_wc, " AND ", g_wc_filter,cl_sql_add_filter("gzwl_t"),
               "   AND gzwl005 = '", pc_gzwl005, "'",
               " ORDER BY gzwl_t.gzwl004" #ODI-ORA#
   #LET g_sql = g_sql,cl_sql_add_filter("gzwl_t")  #190926-00017#2 add   #191111-00008#1   mark 上面已經有 cl_sql_add_filter 、cl_sql_add_mask 所以不用呼叫
   #LET g_sql = cl_sql_add_mask(g_sql)             #190926-00017#2 add   #191111-00008#1   mark 上面已經有 cl_sql_add_filter 、cl_sql_add_mask 所以不用呼叫   
   LET g_sql = cl_sql_add_mask(g_sql)              #遮蔽特定資料
   PREPARE azzq080_process_b_fill_pre FROM g_sql
   DECLARE azzq080_process_b_fill_curs CURSOR FOR azzq080_process_b_fill_pre
   OPEN azzq080_process_b_fill_curs
   CALL g_gzwl_d.clear()
   LET l_ac = 1
   #181117-00002#1--start--   
   #FOREACH azzq080_process_b_fill_curs INTO g_gzwl_d[l_ac].sel,g_gzwl_d[l_ac].gzwl001,g_gzwl_d[l_ac].gzwl003,g_gzwl_d[l_ac].gzwl003_desc,
   #g_gzwl_d[l_ac].gzwl004,g_gzwl_d[l_ac].gzwl015,g_gzwl_d[l_ac].gzwl008,g_gzwl_d[l_ac].gzwl010,g_gzwl_d[l_ac].gzwl007,g_gzwl_d[l_ac].gzwl012,g_gzwl_d[l_ac].gzwl014,g_gzwl_d[l_ac].gzwl013
   #FOREACH azzq080_process_b_fill_curs INTO g_gzwl_d[l_ac].sel,g_gzwl_d[l_ac].gzwl001,g_gzwl_d[l_ac].gzwl012,g_gzwl_d[l_ac].gzwl003,g_gzwl_d[l_ac].gzwl003_desc,                                                #181207-00062#2 mark            
   #                                        g_gzwl_d[l_ac].gzwl004,g_gzwl_d[l_ac].gzwl015,g_gzwl_d[l_ac].gzwl008,g_gzwl_d[l_ac].gzwl010,g_gzwl_d[l_ac].gzwl007,g_gzwl_d[l_ac].gzwl014,g_gzwl_d[l_ac].gzwl013     #181207-00062#2 mark
   #181117-00002#1--end--
   #240628-00031#1 mark --- start ---
#   FOREACH azzq080_process_b_fill_curs INTO g_gzwl_d[l_ac].sel,g_gzwl_d[l_ac].gzwl001,g_gzwl_d[l_ac].gzwl010,g_gzwl_d[l_ac].gzwl012,g_gzwl_d[l_ac].gzwl003,g_gzwl_d[l_ac].gzwl003_desc,                          #181207-00062#2 add 
#                                            g_gzwl_d[l_ac].gzwl004,g_gzwl_d[l_ac].gzwl015,g_gzwl_d[l_ac].gzwl008,g_gzwl_d[l_ac].gzwl007,g_gzwl_d[l_ac].gzwl014,g_gzwl_d[l_ac].gzwl013,                            #181207-00062#2 add
#                                            g_gzwl_d[l_ac].gzwl006        
   #240628-00031#1 mark --- end ---
   #240628-00031#1 add --- start ---
   FOREACH azzq080_process_b_fill_curs INTO g_gzwl_d[l_ac].sel,g_gzwl_d[l_ac].gzwl001,g_gzwl_d[l_ac].gzwl010,g_gzwl_d[l_ac].gzwl012,g_gzwl_d[l_ac].gzwl003,g_gzwl_d[l_ac].gzwl003_desc,                          #181207-00062#2 add 
                                            g_gzwl_d[l_ac].gzwl004,g_gzwl_d[l_ac].gzwl015,g_gzwl_d[l_ac].gzwl008,g_gzwl_d[l_ac].gzwl007,g_gzwl_d[l_ac].gzwl014,g_gzwl_d[l_ac].gzwl013,                            #181207-00062#2 add
                                            g_gzwl_d[l_ac].gzwl006,g_gzwl_d[l_ac].gzwl002
   #240628-00031#1 add --- end ---
   
      
      IF SQLCA.sqlcode THEN
         INITIALIZE g_errparam TO NULL 
         LET g_errparam.extend = "FOREACH:" 
         LET g_errparam.code   = SQLCA.sqlcode 
         LET g_errparam.popup  = TRUE 
         CALL cl_err()
         EXIT FOREACH
      END IF
      #181207-00062#2--start--
      #判斷同一台主機，PID 存在否
      IF g_gzwl_d[l_ac].gzwl015 = cl_get_ap_hostname() AND NOT azzq080_chk_prosess(g_gzwl_d[l_ac].gzwl008) THEN  
         LET g_gzwl_d_bck[g_gzwl_d_bck.getLength()+1].* = g_gzwl_d[l_ac].*

         
         CALL g_gzwl_d.deleteElement(l_ac)
         LET l_ac = g_gzwl_d.getLength() + 1
         CONTINUE FOREACH 
      END IF  
     #181207-00062#2--end--  
      LET g_gzwl_d[l_ac].sel = "N"
      #同一授權process列表全部顯示
      LET l_ac = l_ac + 1
   END FOREACH
   CALL g_gzwl_d.deleteElement(l_ac)

   #160726-00036#1 如果有下條件, 就不要額外列出非gzwl資料
   IF g_wc.trim() = "1=1" OR g_wc.trim() = "gzwl001 IS NOT NULL" THEN
   #再去找TERMINAL沒有落在gzwl資料的process, 沒有gzwl005對應, 先拿g_detail_idx2的fdid來尋找
   #LET g_sql = "SELECT pid, exemode, agent, appname, host FROM licinfo",               #181207-00062#2 mark
   LET g_sql = "SELECT pid, exemode, agent, appname, host ,fdid,ip FROM licinfo",       #181207-00062#2 add
               " WHERE fdid = '", g_gzwl2_d[g_detail_idx2].gzwl005_desc, "'" ,
               "   AND ip = '", g_gzwl2_d[g_detail_idx2].gzwl006,"'"
   LET g_sql = g_sql,cl_sql_add_filter("licinfo")  #190926-00017#2 add
   LET g_sql = cl_sql_add_mask(g_sql)             #190926-00017#2 add 
   PREPARE azzq080_process_b_fill_pre2 FROM g_sql
   DECLARE azzq080_process_b_fill_curs2 CURSOR FOR azzq080_process_b_fill_pre2

 IF g_t100debug >=6 THEN
    DISPLAY " 左側執行清單"
    DISPLAY " gzwl005_desc:",g_gzwl2_d[g_detail_idx2].gzwl005_desc
    DISPLAY " gzwl006:",g_gzwl2_d[g_detail_idx2].gzwl006
    DISPLAY " gzwl002:",g_gzwl2_d[g_detail_idx2].gzwl002 
    FOR li_i = 1 TO g_gzwl_d.getLength()
    DISPLAY " 右側執行清單筆數 g_gzwl_d.getLength():",g_gzwl_d.getLength() 
    DISPLAY " gzwl001:",g_gzwl_d[li_i].gzwl001
    DISPLAY " gzwl010:",g_gzwl_d[li_i].gzwl010
    DISPLAY " gzwl012:",g_gzwl_d[li_i].gzwl012
    DISPLAY " gzwl003:",g_gzwl_d[li_i].gzwl003
    DISPLAY " gzwl003_desc:",g_gzwl_d[li_i].gzwl003_desc
    DISPLAY " gzwl004:",g_gzwl_d[li_i].gzwl004
    DISPLAY " gzwl015:",g_gzwl_d[li_i].gzwl015
    DISPLAY " gzwl008:",g_gzwl_d[li_i].gzwl008
    DISPLAY " gzwl007:",g_gzwl_d[li_i].gzwl007
    DISPLAY " gzwl014:",g_gzwl_d[li_i].gzwl014
    DISPLAY " gzwl013:",g_gzwl_d[li_i].gzwl013
    DISPLAY " gzwl006:",g_gzwl_d[li_i].gzwl006
    END FOR
 END IF
   #181207-00062#2--start--
   #LET l_ac = g_gzwl_d.getLength() + 1
   LET l_ac = 1
   #FOREACH azzq080_process_b_fill_curs2 INTO lr_others.pid, lr_others.exemode, lr_others.agent, lr_others.appname, lr_others.host
   FOREACH azzq080_process_b_fill_curs2 INTO lr_others.pid, lr_others.exemode, lr_others.agent, lr_others.appname, lr_others.host,lr_others.fdid,lr_others.ip
      IF g_t100debug >=6 THEN 
        DISPLAY "lr_others.fdid:",lr_others.fdid," lr_others.pid:",lr_others.pid , " lr_others.exemode:",lr_others.exemode ," lr_others.agent:",lr_others.agent ," lr_others.appname:",lr_others.appname ," lr_others.host:",lr_others.host ," lr_others.ip:",lr_others.ip 
      END IF      
      #主機不同而且PID 不存在，也不顯示出來 
      IF lr_others.host <> cl_get_ap_hostname() AND NOT azzq080_chk_prosess(lr_others.pid) THEN  
         CONTINUE FOREACH 
      END IF
      #主機相同而且PID 不存在，也不顯示出來 
      IF lr_others.host = cl_get_ap_hostname() AND NOT azzq080_chk_prosess(lr_others.pid) THEN
         IF g_t100debug >=6 THEN
            DISPLAY "=======" 
            DISPLAY "pid 不存在"
            DISPLAY "lr_others.pid:",lr_others.pid , " lr_others.host:",lr_others.host
            DISPLAY "======="  
            CONTINUE FOREACH  
         END IF
      END IF  

      #判斷左測清單這筆使帳號是否跟右邊同一筆資料，同一筆就放進右測清單單身  
      SELECT gzwl002 INTO lc_gzwl002 FROM gzwl_t
       WHERE gzwl008 = lr_others.pid 
        AND  gzwl006 = lr_others.ip
        AND  gzwl015 = lr_others.host

      #IF lc_gzwl002 <> g_gzwl2_d[g_detail_idx2].gzwl002 THEN 
      #   IF g_t100debug >=6 THEN
      #      DISPLAY "======= " 
      #      DISPLAY "右測清單使用者編號跟左測清單使使用者帳號不同"
      #      DISPLAY "lc_gzwl002:",lc_gzwl002, " g_gzwl2_d[g_detail_idx2].gzwl002:",g_gzwl2_d[g_detail_idx2].gzwl002
      #      DISPLAY "======= "  
      #      CONTINUE FOREACH  
      #   END IF
      #   CONTINUE FOREACH 
      #END IF   
     #191111-00008#1--start--   #不檢查了避免開啟畫面時間過久
      #FOR li_i = 1 TO g_gzwl_d.getLength()
      #    IF g_gzwl_d[li_i].gzwl008 = lr_others.pid THEN
      #       IF g_gzwl_d[li_i].gzwl015 = lr_others.host THEN 
      #          IF NOT azzq080_chk_prosess(lr_others.pid) THEN
      #             IF g_t100debug >=6 THEN
      #             DISPLAY "======= " 
      #             DISPLAY "pid 不存在"
      #             DISPLAY "lr_others.pid:",lr_others.pid ," lr_others.host:",lr_others.host
      #             DISPLAY "======= "  
      #             CALL g_gzwl_d.deleteElement(li_i)
      #             LET li_i = li_i - 1
      #             CONTINUE FOR  
      #             END IF 
      #         END IF 
      #         #IF g_gzwl_d[li_i].gzwl006             
      #         DISPLAY " ======= 使用者IP gzwl006:",g_gzwl_d[li_i].gzwl006
      #         DISPLAY " ======= PID gzwl008:",g_gzwl_d[li_i].gzwl008 
      #       END IF
      #    END IF
      #END FOR 
     #191111-00008#1--end--      
      #181207-00062#2--end--     
      LET li_repeat = FALSE
      FOR li_i = 1 TO g_gzwl_d.getLength()
          IF g_gzwl_d[li_i].gzwl008 = lr_others.pid THEN
             LET li_repeat = TRUE
             EXIT FOR
          END IF
      END FOR
      IF NOT li_repeat THEN
         LET l_ac = g_gzwl_d.getLength() + 1        #181207-00062#2  add
         LET g_gzwl_d[l_ac].sel = "N" 
         LET g_gzwl_d[l_ac].gzwl008 = lr_others.pid
         LET g_gzwl_d[l_ac].gzwl007 = lr_others.exemode, "-", lr_others.agent
         LET g_gzwl_d[l_ac].gzwl015 = lr_others.host
         #181207-00062#2--start--
         IF cl_null(g_gzwl_d[l_ac].gzwl014) THEN 
            LET g_gzwl_d[l_ac].gzwl014 = "N"
         END IF
         #181207-00062#2--end--          
         IF lr_others.appname IS NOT NULL THEN
            LET g_gzwl_d[l_ac].gzwl003_desc = lr_others.appname
         ELSE
            #181207-00062#2--start--
            #LET lc_channel = base.Channel.create()
           #LET ls_cmd = "ps -ef | grep -v grep | grep '", lr_others.pid, "' | awk '{print $9}'"                   #181108-00014#1   mark
            #LET ls_cmd = "ps -ef | grep -v grep | grep ' ", lr_others.pid, " ' | grep fglrun | awk '{print $9}'"   #181108-00014#1   add
            #CALL lc_channel.openPipe(ls_cmd, "r")
            #CALL lc_channel.setDelimiter("")
            #WHILE lc_channel.read(ls_str)
            #   EXIT WHILE
            #END WHILE        
            #LET g_gzwl_d[l_ac].gzwl003_desc = ls_str
            #不同區域主機忽略，針對同一台主機搜尋程式資訊
            IF g_gzwl_d[l_ac].gzwl015 <> cl_get_ap_hostname() THEN
               DISPLAY " 不同區域主機忽略 gzwl015:",g_gzwl_d[l_ac].gzwl015
               DISPLAY " 不同區域主機忽略 gzwl008:",g_gzwl_d[l_ac].gzwl008            
            ELSE 
                #用 grep 指令搜尋程式資訊
                #CALL azzq080_get_grep_prog(lr_others.*) RETURNING ls_str,ls_time,ls_ent
                CALL azzq080_get_grep_prog(lr_others.*) RETURNING ls_str,ls_ent
                IF cl_null(ls_str) THEN 
                   CALL g_gzwl_d.deleteElement(l_ac)
                   CONTINUE FOREACH 
                END IF 
                #判斷有沒有 fgldeb 表示r.d 
                IF ls_str.getIndexOf("fgldeb",1) THEN 
                   LET g_gzwl_d[l_ac].gzwl003_desc = ls_str  
                ELSE
                   LET g_gzwl_d[l_ac].gzwl003 = os.Path.basename(ls_str)           #作業編號
                   LET ls_str = os.Path.rootname(ls_str)
                   IF ls_str.getIndexOf("erp",1) THEN 
                      LET ls_str = ls_str.subString(1,ls_str.getIndexOf("erp",1)-2)
                      LET ls_str = os.Path.basename(ls_str)
                      LET g_gzwl_d[l_ac].gzwl010 = ls_str                          #執行區域
                   END IF  
                   LET g_gzwl_d[l_ac].gzwl012 = ls_ent                   
                END IF 
                
            END IF
            #181207-00062#2--end--
         END IF
         LET l_ac = l_ac + 1
      END IF
   END FOREACH
   END IF

   #關聯原DB資料
   IF cl_db_setconnect(g_dbs_old) THEN END IF
   LET g_sql = "SELECT gzzal003 FROM gzzal_t WHERE gzzal001 = ? AND gzzal002 = ?"
   PREPARE azzq080_process_b_fill_getgzzal_pre FROM g_sql
   FOR l_ac = 1 TO g_gzwl_d.getLength()
       EXECUTE azzq080_process_b_fill_getgzzal_pre USING g_gzwl_d[l_ac].gzwl003, g_lang INTO g_gzwl_d[l_ac].gzwl003_desc
   END FOR

   #181207-00062#2--start--
   #CALL cl_process_user_limit(pc_gzwl005) RETURNING g_limit_cnt, g_detail_cnt, g_exec_notover
   #160726-00036#1 執行作業數已經不只gzwl_t內的資料
   #LET g_detail_cnt = g_gzwl_d.getLength()
   #IF g_limit_cnt = 32767 THEN
   #   DISPLAY g_detail_cnt, "∞" TO FORMONLY.process_count, FORMONLY.process_limit
   #ELSE
   #   DISPLAY g_detail_cnt, g_limit_cnt TO FORMONLY.process_count, FORMONLY.process_limit
   #END IF

   CALL azzq080_show_proc_cnt_and_lim(pc_gzwl005) 

   #表示右邊清單 已經沒有資料，把左邊清單自動刪除不存在資料 
   IF g_gzwl_d.getLength() = 0 THEN 
      CALL g_gzwl2_d.deleteElement(g_detail_idx2)
      LET g_detail_idx2 = g_detail_idx2 - 1
   END IF    
   CALL azzq080_delete_gzwl()   
   #181207-00062#2--end--   
END FUNCTION
################################################################################
# Descriptions...: Genero License清單整理
# Memo...........: 初測, 不同主機同FGLSERVER, 會是一樣的FrontEndId2, License看來是同一個, 主機資訊應該在Process列表呈現
# Usage..........: CALL azzq080_license_collect()
# Input parameter: None
# Return code....: None
# Date & Author..: 2015/06/30 By Saki
# Modify.........:
################################################################################
PRIVATE FUNCTION azzq080_license_collect()
   DEFINE lc_channel    base.Channel
   DEFINE ls_str        STRING
   DEFINE ls_info       STRING
   DEFINE ls_host       STRING
   DEFINE ls_fdid       STRING
   DEFINE ls_tmp        STRING
   DEFINE li_i          SMALLINT
   DEFINE ls_jsonstr_d  STRING
   DEFINE lr_listchk RECORD
             fdid       LIKE type_t.chr100,
             pid        LIKE type_t.chr20,
             name       LIKE type_t.chr50,
             ip         LIKE type_t.chr20,
             port       LIKE type_t.chr20,      #220801-00038#1
             exemode    LIKE type_t.chr20,
             agent      LIKE type_t.chr20,
             appname    LIKE type_t.chr50,
             host       LIKE type_t.chr50
   END RECORD
   
   DEFINE ls_sql        STRING
   DEFINE li_idx        LIKE type_t.num5
   DEFINE ls_cmd        STRING
   DEFINE li_cnt        LIKE type_t.num5
   DEFINE lc_write      base.Channel             #170821-00043#1
   DEFINE lc_write2     base.Channel             #200827-00023#1
   DEFINE ls_path       STRING                   #170821-00043#1
   DEFINE ld_date       LIKE type_t.dat          #170821-00043#1
   DEFINE ls_log_file   STRING                   #170821-00043#1
   DEFINE ls_log_file2  STRING                   #200827-00023#1
   DEFINE li_count      LIKE type_t.num5         #170821-00043#1
   DEFINE li_count_fdid LIKE type_t.num5         #170821-00043#1
   DEFINE ls_msg_desc   STRING                   #180822-00047#1   
   DEFINE la_gzou       DYNAMIC ARRAY OF RECORD  #181207-00005#1
          gzou003    LIKE gzou_t.gzou003         
                     END RECORD         
   DEFINE li_ex_license LIKE type_t.num5         #200827-00023#1
   DEFINE ls_ex_ip		STRING                   #200827-00023#1
   DEFINE ls_port		   STRING                   #220801-00038#1
   #240628-00031#1 add --- start ---
   DEFINE li_count_keyword        LIKE type_t.num5
   DEFINE la_infoip       DYNAMIC ARRAY OF RECORD  
          ip         LIKE type_t.chr20, 
          pid        LIKE type_t.chr50             
   END RECORD
   DEFINE ls_info_host    LIKE type_t.chr50
   DEFINE la_infodata    DYNAMIC ARRAY OF RECORD  
             fdid       LIKE type_t.chr100,
             pid        LIKE type_t.chr20,
             name       LIKE type_t.chr50,
             ip         LIKE type_t.chr20,
             port       LIKE type_t.chr20,     
             exemode    LIKE type_t.chr20,
             agent      LIKE type_t.chr20,
             appname    LIKE type_t.chr50,
             host       LIKE type_t.chr50,
             gzwl001    LIKE gzwl_t.gzwl001,
             gzwl002    LIKE gzwl_t.gzwl002,
             gzwl003    LIKE gzwl_t.gzwl003,
             gzwl004    LIKE gzwl_t.gzwl004,
             gzwl005    LIKE gzwl_t.gzwl005,
             gzwl006    LIKE gzwl_t.gzwl006,
             gzwl007    LIKE gzwl_t.gzwl007,
             gzwl008    LIKE gzwl_t.gzwl008,
             gzwl009    LIKE gzwl_t.gzwl009,
             gzwl010    LIKE gzwl_t.gzwl010,
             gzwl011    LIKE gzwl_t.gzwl011,
             gzwl012    LIKE gzwl_t.gzwl012,
             gzwl013    LIKE gzwl_t.gzwl013,
             gzwl014    LIKE gzwl_t.gzwl014,
             gzwl015    LIKE gzwl_t.gzwl015,
             gzwl002_desc LIKE type_t.chr500, 
             gzxa007     LIKE gzxa_t.gzxa007,
             gzwl003_desc LIKE type_t.chr500, 
             originfo   STRING,
             ls_detail_cnt2   LIKE type_t.num5,
             ls_auth_cnt      LIKE type_t.num5
   END RECORD
   
   
   #先執行一下process紀錄檢測並清除過期的
   CALL cl_process_auth_refresh()          #r.r adzp444
   CALL azzq080_counting_number()          
   
   LET ls_str = cl_tlm_gl_license_collect(g_dbs_old)
   CALL util.JSON.parse(ls_str, la_infodata)
   
   #取出其他資訊
   INITIALIZE g_fglwrt TO NULL
   LET li_idx = la_infodata.getLength()
   LET g_fglwrt = la_infodata[li_idx].originfo            #原始授權字串
   LET g_detail_cnt2 = la_infodata[li_idx].ls_detail_cnt2 #目前使用數
   LET g_auth_cnt = la_infodata[li_idx].ls_auth_cnt       #總授權數
   CALL la_infodata.deleteElement(li_idx)    
   
   #建立暫存表解析呈現用
   CREATE TEMP TABLE license_info(
      fdid        VARCHAR(100),
      pid         VARCHAR(20),
      name        VARCHAR(50),
      ip          VARCHAR(20),
      port        VARCHAR(20),              
      exemode     VARCHAR(20),
      agent       VARCHAR(20),
      appname     VARCHAR(50),
      host        VARCHAR(50),
      gzwl001     VARCHAR(40),
      gzwl002     VARCHAR(20),
      gzwl003     VARCHAR(40),
      gzwl004     DATE,
      gzwl005     VARCHAR(100),
      gzwl006     VARCHAR(40),
      gzwl007     VARCHAR(10),
      gzwl008     VARCHAR(40),
      gzwl009     VARCHAR(20),
      gzwl010     VARCHAR(40),
      gzwl011     VARCHAR(10),
      gzwl012     VARCHAR(5),
      gzwl013     VARCHAR(40),
	   gzwl014     VARCHAR(1),
	   gzwl015     VARCHAR(40),
	   gzwl002_desc VARCHAR(500),
      gzxa007     VARCHAR(100),
      gzwl003_desc VARCHAR(500)
   );
   #資料轉入暫存表使用
   DELETE FROM license_info
   LET ls_sql = "INSERT INTO license_info(fdid,pid,name,ip,port,exemode,agent,appname,host, ",
                " gzwl001,gzwl002,gzwl003,gzwl004,gzwl005,gzwl006,gzwl007,gzwl008,gzwl009,gzwl010, ",
                " gzwl011,gzwl012,gzwl013,gzwl014,gzwl015,gzwl002_desc,gzxa007,gzwl003_desc )",
                " VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
   PREPARE azzq080_license_collect_license_info FROM ls_sql
   
   #開始寫入暫存檔
   FOR li_i = 1 TO la_infodata.getLength()
      #250709-00026#1 add --- start ---
      IF cl_null(la_infodata[li_i].gzwl006) THEN
         LET ls_ex_ip = la_infodata[li_i].ip
         LET la_infodata[li_i].gzwl006 = ls_ex_ip.trim()
      END IF
      #250709-00026#1 add --- end ---
      EXECUTE azzq080_license_collect_license_info
        USING la_infodata[li_i].fdid,la_infodata[li_i].pid,la_infodata[li_i].name,
              la_infodata[li_i].ip,la_infodata[li_i].port,la_infodata[li_i].exemode,
              la_infodata[li_i].agent,la_infodata[li_i].appname,la_infodata[li_i].host,
              la_infodata[li_i].gzwl001,la_infodata[li_i].gzwl002,la_infodata[li_i].gzwl003,
              la_infodata[li_i].gzwl004,la_infodata[li_i].gzwl005,la_infodata[li_i].gzwl006,
              la_infodata[li_i].gzwl007,la_infodata[li_i].gzwl008,la_infodata[li_i].gzwl009,
              la_infodata[li_i].gzwl010,la_infodata[li_i].gzwl011,la_infodata[li_i].gzwl012,
              la_infodata[li_i].gzwl013,la_infodata[li_i].gzwl014,la_infodata[li_i].gzwl015,
              la_infodata[li_i].gzwl002_desc,la_infodata[li_i].gzxa007,la_infodata[li_i].gzwl003_desc
              
      IF g_t100debug >=6 THEN 
         DISPLAY " ------------------------- "
         DISPLAY " la_infodata[li_i].fdid:",la_infodata[li_i].fdid
         DISPLAY " la_infodata[li_i].pid:",la_infodata[li_i].pid
         DISPLAY " la_infodata[li_i].name:",la_infodata[li_i].name
         DISPLAY " la_infodata[li_i].ip:",la_infodata[li_i].ip
         DISPLAY " la_infodata[li_i].exemode:",la_infodata[li_i].exemode
         DISPLAY " la_infodata[li_i].agent:",la_infodata[li_i].agent
         DISPLAY " la_infodata[li_i].appname:",la_infodata[li_i].appname
         DISPLAY " la_infodata[li_i].host:",la_infodata[li_i].host  
         DISPLAY " la_infodata[li_i].gzwl001:",la_infodata[li_i].gzwl001
         DISPLAY " la_infodata[li_i].gzwl002:",la_infodata[li_i].gzwl002
         DISPLAY " la_infodata[li_i].gzwl003:",la_infodata[li_i].gzwl003
         DISPLAY " la_infodata[li_i].gzwl004:",la_infodata[li_i].gzwl004
         DISPLAY " la_infodata[li_i].gzwl005:",la_infodata[li_i].gzwl005
         DISPLAY " la_infodata[li_i].gzwl006:",la_infodata[li_i].gzwl006
         DISPLAY " la_infodata[li_i].gzwl007:",la_infodata[li_i].gzwl007
         DISPLAY " la_infodata[li_i].gzwl008:",la_infodata[li_i].gzwl008
         DISPLAY " la_infodata[li_i].gzwl009:",la_infodata[li_i].gzwl009
         DISPLAY " la_infodata[li_i].gzwl010:",la_infodata[li_i].gzwl010
         DISPLAY " la_infodata[li_i].gzwl011:",la_infodata[li_i].gzwl011
         DISPLAY " la_infodata[li_i].gzwl012:",la_infodata[li_i].gzwl012
         DISPLAY " la_infodata[li_i].gzwl013:",la_infodata[li_i].gzwl013
         DISPLAY " la_infodata[li_i].gzwl014:",la_infodata[li_i].gzwl014
         DISPLAY " la_infodata[li_i].gzwl015:",la_infodata[li_i].gzwl015
         DISPLAY " la_infodata[li_i].gzwl002_desc:",la_infodata[li_i].gzwl002_desc
         DISPLAY " la_infodata[li_i].gzxa007:",la_infodata[li_i].gzxa007
         DISPLAY " la_infodata[li_i].gzwl003_desc:",la_infodata[li_i].gzwl003_desc
         DISPLAY " ------------------------- "
         DISPLAY " "      
      END IF 
   END FOR

   #寫入左側資料顯示
   LET ls_sql = " SELECT DISTINCT fdid FROM license_info WHERE fdid IS NOT NULL "
               , " AND ", g_wc, " AND ", g_wc_filter, cl_sql_add_filter("gzwl_t")
   LET ls_sql = cl_sql_add_mask(ls_sql)
   
   PREPARE azzq080_license_collect_pre FROM ls_sql
   DECLARE azzq080_license_collect_curs CURSOR FOR azzq080_license_collect_pre

   #處理 IP顯示
   LET ls_sql = " SELECT DISTINCT ip FROM license_info WHERE fdid = ? "
   PREPARE azzq080_license_collect_ip_pre FROM ls_sql
   DECLARE azzq080_license_collect_ip_curs CURSOR FOR azzq080_license_collect_ip_pre

   CALL g_gzwl2_d.clear()
   LET li_idx = 1
   FOREACH azzq080_license_collect_curs INTO g_gzwl2_d[li_idx].gzwl005_desc
      LET g_gzwl2_d[li_idx].num = li_idx                          
      #ip 可能有重複，另外處理加入
      CALL la_infoip.clear()  
      LET li_i = 1 
      FOREACH azzq080_license_collect_ip_curs USING g_gzwl2_d[li_idx].gzwl005_desc INTO la_infoip[li_i].ip
         IF NOT cl_null(la_infoip[li_i].ip) THEN
            LET ls_tmp = la_infoip[li_i].ip
            IF li_i = 1 THEN
               LET g_gzwl2_d[li_idx].gzwl006 = ls_tmp.trim()
            ELSE
               LET g_gzwl2_d[li_idx].gzwl006 = g_gzwl2_d[li_idx].gzwl006, ",", ls_tmp.trim()
            END IF
         END IF 
         LET li_i = li_i + 1
      END FOREACH
      LET li_idx = li_idx + 1
   END FOREACH
   CALL g_gzwl2_d.deleteElement(li_idx)   
   
   LET g_detail_cnt = g_gzwl2_d.getlength()

   #240628-00031#1 add --- end ---
   #240628-00031#1 mark --- start ---
#
#   CALL g_gzwl2_d.clear()
#   CALL g_aid.clear()                     
#   CALL la_gzou.clear()                     #181207-00005#1
#   CALL g_license_d.clear()                 #181207-00062#2
#   INITIALIZE g_fglwrt TO NULL
#
#   #先執行一下process紀錄檢測並清除過期的
#   CALL cl_process_auth_refresh()          #r.r adzp444
#   CALL azzq080_counting_number()          #181207-00062#2  add
#
#   #整理資訊要用到全體gzwl_t, 需先連結dscom, 處理完後才連回原DB
#   IF cl_db_setconnect("dscom") THEN END IF
#   CREATE TEMP TABLE licinfo(
#      fdid        VARCHAR(100),
#      pid         VARCHAR(20),
#      name        VARCHAR(50),
#      ip          VARCHAR(20),
#      port        VARCHAR(20),              #220801-00038#1
#      exemode     VARCHAR(20),
#      agent       VARCHAR(20),
#      appname     VARCHAR(50),
#      host        VARCHAR(50)
#   );
#   DELETE FROM licinfo
#   LET ls_sql = "INSERT INTO licinfo(fdid,pid,name,ip,port,exemode,agent,appname,host)",
#                "             VALUES(?,?,?,?,?,?,?,?,?)"
#   PREPARE azzq080_license_collect_tmp_pre FROM ls_sql
#
#   LET lc_channel = base.Channel.create()
#   CALL lc_channel.openPipe("fglWrt -a info users", "r")
#   CALL lc_channel.setDelimiter("")
#   #170821-00043#1--寫log--start--
#   LET li_count = 1
#   LET li_count_fdid = 1
#   LET lc_write = base.Channel.create()
#   LET lc_write2 = base.Channel.create()                        #200827-00023#1  add
#   LET ld_date = cl_get_today()
#   LET ls_tmp = ld_date USING "yyyymmdd"
#   #LET ls_path = "azzq080_", ls_tmp CLIPPED, ".log"            #181207-00062#2  mark
#   LET ls_path = "azzq080_", g_log CLIPPED, ".log"              #181207-00062#2  add
#   LET ls_log_file = os.Path.join(FGL_GETENV("LOGDIR"), ls_path)
#   IF NOT os.Path.exists(ls_log_file) THEN
#      CALL lc_write.openFile(ls_log_file,"w")
#   ELSE
#      CALL lc_write.openFile(ls_log_file,"a")
#   END IF
#   LET ls_log_file2 = os.Path.join(FGL_GETENV("LOGDIR"), 'azzq080_list.log')     #200827-00023#1
#   CALL lc_write2.openFile(ls_log_file2,"w")                      #200827-00023#1   add
#   CALL lc_write.setDelimiter("")
#   CALL lc_write2.setDelimiter("")                                #200827-00023#1   add
#   LET ls_tmp = "#--------------------------- (", cl_get_current(), ") ------------------------#"
#   CALL lc_write.write(ls_tmp)
#   LET ls_tmp = "[info] 開始解析 'fglWrt -a info users'..... "
#   CALL lc_write.write(ls_tmp)
#   #170821-00043#1--寫log--end--
#   WHILE lc_channel.read(ls_str)
#      CALL lc_write2.write(ls_str)                                #200827-00023#1   add
#      IF ls_str IS NULL THEN
#         CONTINUE WHILE
#      END IF
#      LET g_fglwrt = g_fglwrt, ls_str, "\n"
#
#      #處理單頭
#      CASE
#         WHEN ls_str.getIndexOf("CPU(s)", 1)  
#            LET g_auth_cnt = 32767
#         WHEN ls_str.getIndexOf("Users", 1)    #目前使用數/總授權數 eg. 20/250
#            LET ls_tmp = "[info] 讀取授權使用人數 ",ls_str       #170821-00043#1  寫log
#            CALL lc_write.write(ls_tmp)                         #170821-00043#1  寫log
#            
#            LET ls_info = ls_str.subString(ls_str.getIndexOf(":",1)+1, ls_str.getLength())
#            IF ls_info IS NULL THEN
#               LET ls_info = ls_str.subString(ls_str.getIndexOf("Users",1)+5, ls_str.getLength())
#            END IF
#            LET ls_info = ls_info.trim()
#            LET ls_tmp = ls_info.subString(1, ls_info.getIndexOf("/",1)-1)
#            LET g_detail_cnt2 = ls_tmp       #目前使用數
#            LET ls_tmp = ls_info.subString(ls_info.getIndexOf("/",1)+1, ls_info.getLength())
#            LET g_auth_cnt = ls_tmp          #總授權數
#      END CASE
#
#      #Host是取到一次以後給單身輪著用的
#      IF ls_str.getIndexOf("Host",1) THEN
#         LET ls_info = ls_str.subString(ls_str.getIndexOf(":",1)+1, ls_str.getLength())
#         IF ls_info IS NULL THEN
#            LET ls_info = ls_str.subString(ls_str.getIndexOf("Host",1)+4, ls_str.getLength())
#         END IF
#         LET ls_host = ls_info.trim()
#         INITIALIZE ls_fdid TO NULL                               #200827-00023#1   add
#         CONTINUE WHILE
#      END IF
# 
#      #授權編號
#      IF ls_str.getIndexOf("FrontEndId2",1) THEN   
#         LET ls_tmp = ls_str.trim()        #170821-00043#1  寫log
#         CALL lc_write.write(ls_tmp)       #170821-00043#1  寫log
#         LET ls_info = ls_str.subString(ls_str.getIndexOf(":",1)+1, ls_str.getLength())
#         IF ls_info IS NULL THEN
#            LET ls_info = ls_str.subString(ls_str.getIndexOf("FrontEndId2",1)+11, ls_str.getLength())
#         END IF
#         LET ls_fdid = ls_info.trim()
#         CONTINUE WHILE
#      END IF
#      #200827-00023#1   add start
#      IF ls_str.getIndexOf("GUI Server",1) THEN
#         LET li_ex_license = 1
#         INITIALIZE ls_ex_ip TO NULL
#         IF NOT ls_str.getIndexOf("Process Id",1) THEN
#           LET ls_tmp = "FrontEndId2: {Unable to analyze}"
#           CALL lc_write.write(ls_tmp)       #170821-00043#1  寫log 
#           LET ls_fdid = "Unable to analyze"
#           LET li_ex_license = 0
#           LET ls_ex_ip = ls_str.subString(ls_str.getIndexOf("GUI Server",1)+10, ls_str.getLength())
#           #220801-00038#1 add s
#           LET ls_ex_ip = ls_str.subString(ls_str.getIndexOf("GUI Server",1)+10, ls_str.getLength())
#           INITIALIZE ls_port TO NULL
#           LET ls_port = ls_str.subString(ls_str.getIndexOf(":",1)+1, ls_str.getLength())
#         ELSE
#           INITIALIZE ls_port TO NULL
#           LET ls_port = ls_str.subString(ls_str.getIndexOf(":",1)+1, ls_str.getIndexOf(" - ",1)-1)
#           #220801-00038#1 add e
#         END IF
#      END IF
#      #200827-00023#1   add end
#      #找單身
#      FOR li_i = 1 TO g_keyword_d.getLength()
#          IF ls_str.getIndexOf(g_keyword_d[li_i].keyname,1) THEN
#             INITIALIZE ls_info TO NULL
#             IF g_keyword_d[li_i].keyname = "Process Id" THEN        #讀取到 "Process Id" 開始做 insert table的動作 
#                #IF ls_jsonstr_d IS NOT NULL THEN                    #200827-00023#1   mark
#                #200827-00023#1   add start
#                IF ls_jsonstr_d IS NULL THEN
#                  #第一次讀到"Process Id"，ls_jsonstr_d是空的，給予初值
#                   LET ls_jsonstr_d = ls_jsonstr_d, "\"host\":\"", ls_host, "\",\"fdid\":\"", ls_fdid, "\","
#                   LET ls_info = ls_str.subString(ls_str.getIndexOf(g_keyword_d[li_i].keyname,1)+g_keyword_d[li_i].wordnum, ls_str.getLength())
#                   LET ls_jsonstr_d = ls_jsonstr_d, "\"port\":\"", ls_port, "\","               #220801-00038#1 add
#                END IF
#                IF NOT ls_str.getIndexOf("GUI Server",1) THEN
#                   LET li_ex_license = 0
#                   LET ls_info = ls_info.trim()
#                   LET ls_jsonstr_d = ls_jsonstr_d, "\"", g_keyword_d[li_i].field, "\":\"", ls_info, "\","
#                   LET ls_jsonstr_d = ls_jsonstr_d, "\"exemode\":\"UTA\",\"name\":\"GUI Server\"," 
#                   LET ls_jsonstr_d = ls_jsonstr_d, "\"ip\":\"", ls_ex_ip, "\","
#                   LET ls_jsonstr_d = ls_jsonstr_d, "\"port\":\"", ls_port, "\","               #220801-00038#1 add
#                END IF
#             END IF
#             IF li_i = g_keyword_d.getLength() OR li_ex_license = 0 THEN        #讀取到 "Process Id" 開始做 insert table的動作 
#             #200827-00023#1   add end
#                   #要先做一個很白癡的比對方式
#                   CALL util.JSON.parse("{"||ls_jsonstr_d.subString(1, ls_jsonstr_d.getLength()-1)||"}", lr_listchk)
#                   #推測是GWC進入
#                   IF lr_listchk.agent IS NOT NULL THEN
#                      INITIALIZE lr_listchk.fdid TO NULL
#                      LET lr_listchk.fdid = "Web Process - ", lr_listchk.pid
#                      LET ls_jsonstr_d = util.JSON.stringify(lr_listchk)
#                      LET ls_jsonstr_d = ls_jsonstr_d.subString(2, ls_jsonstr_d.getLength()-1),","
#                   END IF
#                   #推測是WS進入
#                   IF lr_listchk.pid   IS NOT NULL AND lr_listchk.name    IS NULL AND
#                      lr_listchk.ip    IS NULL     AND lr_listchk.exemode IS NULL AND
#                      lr_listchk.agent IS NULL     AND lr_listchk.appname IS NULL THEN
#                      INITIALIZE lr_listchk.fdid TO NULL
#                      LET lr_listchk.name = "Web Service"
#                      LET lr_listchk.fdid = "Process - ", lr_listchk.pid
#                      LET ls_jsonstr_d = util.JSON.stringify(lr_listchk)
#                      LET ls_jsonstr_d = ls_jsonstr_d.subString(2, ls_jsonstr_d.getLength()-1),","
#                   END IF
#                   
#                   #170821-00043#1--寫log--start--
#                   LET ls_tmp = "[INSERT TEMP TABLE 第", li_count, "筆 {fdid :", lr_listchk.fdid, 
#                                                                     " ;  ip :", lr_listchk.ip, 
#                                                                     " ;  port :", lr_listchk.port,           #220801-00038#1 add
#                                                                     " ; pid :",lr_listchk.pid,"}]"  
#                   LET li_count = li_count + 1                                 
#                   CALL lc_write.write(ls_tmp)
#                   #170821-00043#1--寫log--end--     
#                   
#                   #INSERT TEMP TABLE
#                   #181207-00062#2--start--
#                   #EXECUTE azzq080_license_collect_tmp_pre                   
#                   #  USING lr_listchk.fdid,lr_listchk.pid,lr_listchk.name,lr_listchk.ip,
#                   #        lr_listchk.exemode,lr_listchk.agent,lr_listchk.appname,
#                   #        lr_listchk.host
#                   LET li_cnt = g_license_d.getLength()+1
#                   LET g_license_d[li_cnt].fdid = lr_listchk.fdid  
#                   LET g_license_d[li_cnt].pid = lr_listchk.pid
#                   LET g_license_d[li_cnt].name = lr_listchk.name 
#                   LET g_license_d[li_cnt].ip = lr_listchk.ip 
#                   LET g_license_d[li_cnt].port = lr_listchk.port                           #220801-00038#1 add
#                   LET g_license_d[li_cnt].exemode = lr_listchk.exemode 
#                   LET g_license_d[li_cnt].agent = lr_listchk.agent 
#                   LET g_license_d[li_cnt].appname = lr_listchk.appname 
#                   LET g_license_d[li_cnt].host = lr_listchk.host 
#                   LET g_license_d[li_cnt].sessionkey = g_sessionkey
#                   #181207-00062#2--end-- 
#                   LET li_ex_license = 1           #200827-00023#1   add
#                   #insert完，清空ls_jsonstr_d，給予初值以便做下次insert
#                   INITIALIZE ls_jsonstr_d TO NULL
#                   EXIT FOR                        #200827-00023#1   add
##                   LET ls_jsonstr_d = ls_jsonstr_d, "\"host\":\"", ls_host, "\",\"fdid\":\"", ls_fdid, "\","           #200827-00023#1 mark
##                ELSE                                                                                                   #200827-00023#1 mark
##                   #第一次讀到"Process Id"，ls_jsonstr_d是空的，給予初值                                                  #200827-00023#1 mark
##                   LET ls_jsonstr_d = ls_jsonstr_d, "\"host\":\"", ls_host, "\",\"fdid\":\"", ls_fdid, "\","           #200827-00023#1 mark
##                END IF                                                                                                 #200827-00023#1 mark
##                LET ls_info = ls_str.subString(ls_str.getIndexOf(g_keyword_d[li_i].keyname,1)+g_keyword_d[li_i].wordnum, ls_str.getLength())   #200827-00023#1 mark
#             END IF
#             IF ls_info IS NULL THEN
#                LET ls_info = ls_str.subString(ls_str.getIndexOf(":",1)+1, ls_str.getLength())
#             END IF
#             IF ls_info IS NULL THEN
#                LET ls_info = ls_str.subString(ls_str.getIndexOf(g_keyword_d[li_i].keyname,1)+g_keyword_d[li_i].wordnum, ls_str.getLength())
#             END IF
#             LET ls_info = ls_info.trim()
#             LET ls_jsonstr_d = ls_jsonstr_d, "\"", g_keyword_d[li_i].field, "\":\"", ls_info, "\","
#             EXIT FOR
#          END IF
#      END FOR
#   END WHILE
#   #最後一筆資料要加進去呀!   讀取到最後才做insert，所以最後一筆資料還需另外insert
#   #200827-00023#1 mark start
##   IF ls_jsonstr_d IS NOT NULL THEN
##      CALL util.JSON.parse("{"||ls_jsonstr_d.subString(1, ls_jsonstr_d.getLength()-1)||"}", lr_listchk)
##      IF lr_listchk.agent IS NOT NULL THEN
##         INITIALIZE lr_listchk.fdid TO NULL
##         LET lr_listchk.fdid = "Web Process - ", lr_listchk.pid
##         LET ls_jsonstr_d = util.JSON.stringify(lr_listchk)
##         LET ls_jsonstr_d = ls_jsonstr_d.subString(2, ls_jsonstr_d.getLength()-1),","
##      END IF
##      #推測是WS進入
##      IF lr_listchk.pid IS NOT NULL AND lr_listchk.name IS NULL AND
##         lr_listchk.ip IS NULL AND lr_listchk.exemode IS NULL AND
##         lr_listchk.agent IS NULL AND lr_listchk.appname IS NULL THEN
##         INITIALIZE lr_listchk.fdid TO NULL
##         LET lr_listchk.name = "Web Service"
##         LET lr_listchk.fdid = "Process - ", lr_listchk.pid
##         LET ls_jsonstr_d = util.JSON.stringify(lr_listchk)
##         LET ls_jsonstr_d = ls_jsonstr_d.subString(2, ls_jsonstr_d.getLength()-1),","
##      END IF
##      
##      #170821-00043#1--寫log--start--
##      LET ls_tmp = "[INSERT TEMP TABLE 第", li_count, "筆 {fdid :", lr_listchk.fdid, 
##                                                        " ;  ip :", lr_listchk.ip, 
##                                                        " ; pid :", lr_listchk.pid,"}]"  
##      LET li_count = li_count + 1                               
##      CALL lc_write.write(ls_tmp)     
##      #170821-00043#1--寫log--end--
##      #181207-00062#2--start--
##      #EXECUTE azzq080_license_collect_tmp_pre
##      #  USING lr_listchk.fdid,lr_listchk.pid,lr_listchk.name,lr_listchk.ip,
##      #        lr_listchk.exemode,lr_listchk.agent,lr_listchk.appname,
##      #        lr_listchk.host
##     LET li_cnt = g_license_d.getLength()+1        
##      LET g_license_d[li_cnt].fdid = lr_listchk.fdid  
##      LET g_license_d[li_cnt].pid = lr_listchk.pid
##      LET g_license_d[li_cnt].name = lr_listchk.name 
##      LET g_license_d[li_cnt].ip = lr_listchk.ip 
##      LET g_license_d[li_cnt].exemode = lr_listchk.exemode 
##      LET g_license_d[li_cnt].agent = lr_listchk.agent 
##      LET g_license_d[li_cnt].appname = lr_listchk.appname 
##      LET g_license_d[li_cnt].host = lr_listchk.host
##      LET g_license_d[li_cnt].sessionkey = g_sessionkey
##      #181207-00062#2--end--
##      
##      INITIALIZE ls_jsonstr_d TO NULL
##   END IF
#   #200827-00023#1 mark end
#   IF g_t100debug >=6 THEN 
#      DISPLAY "解析授權資訊後長度 g_license_d.getLength:",g_license_d.getLength()   
#   END IF
#   #191111-00008#1--start--   #不檢查了避免開啟畫面時間過久
#   #FOR li_i = 1 TO g_license_d.getLength()
#   #     #登入主機相同也確認主機上 process 是否存在
#   #    IF g_license_d[li_i].host = cl_get_ap_hostname() THEN 
#   #       IF NOT azzq080_chk_prosess(g_license_d[li_i].pid) THEN
#   #          IF g_t100debug >=6 THEN
#   #             DISPLAY "檢查授權資訊 chk_prosess pid 不存在:",g_license_d[li_i].pid ," li_i:",li_i
#   #          END IF          
#   #          CALL g_license_d.deleteElement(li_i)
#   #          LET li_i = li_i - 1
#   #          CONTINUE FOR 
#   #       #ELSE 
#   #          #CALL g_license_d.deleteElement(li_i)
#   #          #LET li_i = 1 
#   #       END IF 
#   #    END IF 
#   #END FOR
#   #191111-00008#1--end--   
#   #181207-00062#2--start--
#   #開始寫入暫存檔
#   FOR li_i = 1 TO g_license_d.getLength()
#      EXECUTE azzq080_license_collect_tmp_pre
#        USING g_license_d[li_i].fdid,g_license_d[li_i].pid,g_license_d[li_i].name,
#              g_license_d[li_i].ip,g_license_d[li_i].port,g_license_d[li_i].exemode,g_license_d[li_i].agent,
#              g_license_d[li_i].appname,g_license_d[li_i].host
#      IF g_t100debug >=6 THEN 
#         DISPLAY " ------------------------- "
#         DISPLAY " g_license_d[li_i].fdid:",g_license_d[li_i].fdid
#         DISPLAY " g_license_d[li_i].pid:",g_license_d[li_i].pid
#         DISPLAY " g_license_d[li_i].name:",g_license_d[li_i].name
#         DISPLAY " g_license_d[li_i].ip:",g_license_d[li_i].ip
#         DISPLAY " g_license_d[li_i].ip:",g_license_d[li_i].port                  #220801-00038#1 add
#         DISPLAY " g_license_d[li_i].exemode:",g_license_d[li_i].exemode
#         DISPLAY " g_license_d[li_i].agent:",g_license_d[li_i].agent
#         DISPLAY " g_license_d[li_i].appname:",g_license_d[li_i].appname
#         DISPLAY " g_license_d[li_i].host:",g_license_d[li_i].host  
#         DISPLAY " ------------------------- "
#         DISPLAY " "      
#      END IF 
#   END FOR
#   #181207-00062#2--end--
#
#   #170821-00043#1--寫log--start--
#   SELECT COUNT (DISTINCT fdid) INTO li_count_fdid FROM licinfo   #UNIQUE FrontEndId2 個數
#   
#   LET ls_tmp = "***************************************************************************"
#   CALL lc_write.write(ls_tmp) 
#   LET ls_tmp = "[info] temp_table資訊： 共",li_count-1,"筆"
#   CALL lc_write.write(ls_tmp)   
#   LET ls_tmp = "[info] temp_table資訊： UNIQUE FrontEndId2 共",li_count_fdid,"筆"
#   CALL lc_write.write(ls_tmp)   
#   LET ls_tmp = "[info] 開始寫入授權人數統計清單....."
#   CALL lc_write.write(ls_tmp) 
#   #170821-00043#1--寫log--end--
#   
#   
#   #整理, 並結合process資訊 where 條件是 FrontEndId 跟 ip
#   LET ls_sql = "SELECT t1.gzwl005, t1.gzwl002, t1.gzwl012, t1.gzwl006, t0.port, t1.gzwl015, t1.gzwl009, t0.fdid, t0.name, t0.host",
#                "  FROM licinfo t0",
#                "  LEFT JOIN gzwl_t t1 ON t1.gzwl008 = t0.pid AND t1.gzwl015 = t0.host",
#                " WHERE t0.fdid = ? AND t0.ip = ?"
#   PREPARE azzq080_license_collect_gzwl_pre FROM ls_sql
#   DECLARE azzq080_license_collect_gzwl_curs CURSOR FOR azzq080_license_collect_gzwl_pre
#   LET ls_sql = "SELECT t1.gzwl005, t1.gzwl002, t1.gzwl012, t1.gzwl006, t0.port, t1.gzwl015, t1.gzwl009, t0.fdid, t0.name, t0.host",
#                "  FROM licinfo t0",
#                "  LEFT JOIN gzwl_t t1 ON t1.gzwl008 = t0.pid AND t1.gzwl015 = t0.host",
#                " WHERE t0.pid = ?"
#   PREPARE azzq080_license_collect_gzwl2_pre FROM ls_sql
#   DECLARE azzq080_license_collect_gzwl2_curs CURSOR FOR azzq080_license_collect_gzwl2_pre
#
#   #160726-00036#1 有下查詢條件時, 左側選單要經過篩選的SQL處理
#   LET ls_sql = "SELECT COUNT(*) FROM gzwl_t WHERE ", g_wc," AND gzwl005 = ?"
#   PREPARE azzq080_license_collect_qbe_pre FROM ls_sql
#   DECLARE azzq080_license_collect_qbe_curs CURSOR FOR azzq080_license_collect_qbe_pre
#
#   LET ls_sql = "SELECT fdid, ip, port, MAX(pid) FROM licinfo GROUP BY fdid, ip, port"
#   PREPARE azzq080_license_collect_analyze_pre FROM ls_sql
#   DECLARE azzq080_license_collect_analyze_curs CURSOR FOR azzq080_license_collect_analyze_pre
#   LET li_idx = 1
#   #從temp table中抓取每個授權使用者
#   FOREACH azzq080_license_collect_analyze_curs INTO lr_listchk.fdid, lr_listchk.ip, lr_listchk.port, lr_listchk.pid
#      INITIALIZE lr_listchk.name, lr_listchk.host TO NULL
#      IF lr_listchk.fdid IS NOT NULL AND lr_listchk.ip IS NOT NULL THEN
#         #授權編號不是NULL AND　IP 不是NULL 
#         #同一個FGLSERVER會混雜不紀錄gzwl的process進來, 最好找到其中有gzwl的來呈現
#         FOREACH azzq080_license_collect_gzwl_curs USING lr_listchk.fdid, lr_listchk.ip
#                 #授權編號、使用者編號、使用者編號、企業編號、使用者ip、AP 主機、員工編號、FrontEndId、name、host         
#            INTO g_gzwl2_d[li_idx].gzwl005, g_gzwl2_d[li_idx].gzwl002, g_aid[li_idx].gzwl012, g_gzwl2_d[li_idx].gzwl006, g_gzwl2_d[li_idx].port, g_gzwl2_d[li_idx].gzwl015, g_aid[li_idx].gzwl009, g_gzwl2_d[li_idx].gzwl005_desc, lr_listchk.name, lr_listchk.host
#            IF g_gzwl2_d[li_idx].gzwl005 IS NOT NULL THEN
#               EXIT FOREACH
#            END IF
#         END FOREACH
#         #gzwl005:授權編號
#         IF g_gzwl2_d[li_idx].gzwl005 IS NULL AND lr_listchk.fdid[1] = "{" THEN
#            LET g_gzwl2_d[li_idx].gzwl002_desc = "Developer"
#         END IF
#      ELSE
#         #GWC進入沒有fdid, 而且一個process就佔一個user license, 用pid連結gzwl忠實呈現
#         FOREACH azzq080_license_collect_gzwl2_curs USING lr_listchk.pid
#            INTO g_gzwl2_d[li_idx].gzwl005, g_gzwl2_d[li_idx].gzwl002, g_aid[li_idx].gzwl012, g_gzwl2_d[li_idx].gzwl006, g_gzwl2_d[li_idx].port, g_gzwl2_d[li_idx].gzwl015, g_aid[li_idx].gzwl009, g_gzwl2_d[li_idx].gzwl005_desc, lr_listchk.name, lr_listchk.host
#            IF g_gzwl2_d[li_idx].gzwl005 IS NOT NULL THEN
#               EXIT FOREACH
#            END IF
#         END FOREACH
#      END IF
#      #160726-00036#1 有下查詢條件時, gzwl資料有存在才顯示在左方列表
#      IF g_wc.trim() != "1=1" AND g_wc.trim() != "gzwl001 IS NOT NULL" THEN
#         IF g_gzwl2_d[li_idx].gzwl005 IS NULL THEN
#            CALL g_gzwl2_d.deleteElement(li_idx)
#            CONTINUE FOREACH
#         ELSE
#            EXECUTE azzq080_license_collect_qbe_curs USING g_gzwl2_d[li_idx].gzwl005 INTO li_cnt
#            IF li_cnt <= 0 THEN
#               CALL g_gzwl2_d.deleteElement(li_idx)
#               CONTINUE FOREACH
#            END IF
#         END IF
#      END IF
#      #gzwl005:授權編號
#      IF g_gzwl2_d[li_idx].gzwl005 IS NULL THEN
#         LET g_gzwl2_d[li_idx].gzwl002 = lr_listchk.name         #使用者編號
#         LET g_gzwl2_d[li_idx].gzwl006 = lr_listchk.ip           #使用者ip
#         LET g_gzwl2_d[li_idx].port = lr_listchk.port            #使用者port          #220801-00038#1 add
#         LET g_gzwl2_d[li_idx].gzwl015 = lr_listchk.host         #AP 主機  
#         LET g_gzwl2_d[li_idx].gzwl005_desc = lr_listchk.fdid    #FrontEndId
#      END IF
#      
#      #170821-00043#1--寫log--start--
#      LET ls_tmp = "[INSERT g_gzwl2_d(左側清單) 第", li_idx,"筆 {fdid :", g_gzwl2_d[li_idx].gzwl005_desc, 
#                                                              " ; ip :", g_gzwl2_d[li_idx].gzwl006, 
#                                                              " ; port :", g_gzwl2_d[li_idx].port,       #220801-00038#1 add
#                                                              " ; pid :", lr_listchk.pid,"}]"
#      CALL lc_write.write(ls_tmp)
#      LET g_gzwl2_d[li_idx].num = li_idx                          #200827-00023#1 add
#      LET li_idx = li_idx + 1
#      #170821-00043#1--寫log--end--
#   END FOREACH
#   
#   #170821-00043#1--寫log--start--
#   LET ls_tmp = "***************************************************************************"
#   CALL lc_write.write(ls_tmp) 
#   LET ls_tmp = "[info] g_gzwl2_d(左側清單)資訊： 共",g_gzwl2_d.getLength(),"筆"
#   CALL lc_write.write(ls_tmp)
#   #170821-00043#1--寫log--end--
#
#   #再結合原有資料庫資訊
#   IF cl_db_setconnect(g_dbs_old) THEN END IF
#   LET ls_sql = "SELECT ooag011 FROM ooag_t WHERE ooagent = ? AND ooag001 = ?"
#   PREPARE azzq080_license_collect_ooag_pre FROM ls_sql
#   LET ls_sql = "SELECT gzxa007 FROM gzxa_t WHERE gzxaent = ? AND gzxa001 = ?"
#   PREPARE azzq080_license_collect_gzxa_pre FROM ls_sql
#   FOR li_idx = 1 TO g_gzwl2_d.getLength()
#       EXECUTE azzq080_license_collect_ooag_pre USING g_aid[li_idx].gzwl012, g_aid[li_idx].gzwl009 INTO g_gzwl2_d[li_idx].gzwl002_desc
#       #左側參雜fglWrt取出且gzwl並不存在的資訊, fglWrt的電腦id可能被誤解讀為使用者帳號
#       EXECUTE azzq080_license_collect_gzxa_pre USING g_aid[li_idx].gzwl012, g_gzwl2_d[li_idx].gzwl002 INTO g_gzwl2_d[li_idx].gzxa007
#   END FOR
#   #181207-00005#1--start--
#   LET ls_sql = "SELECT DISTINCT gzou003 FROM gzou_t ",
#                " WHERE  gzoustus = 'Y' " 
#
#   PREPARE azzq080_license_collect_gzou_pre FROM ls_sql
#   DECLARE azzq080_license_collect_gzou_cs CURSOR FOR azzq080_license_collect_gzou_pre
#   LET li_i = 1
#   FOREACH azzq080_license_collect_gzou_cs INTO la_gzou[li_i].gzou003
#
#      LET li_i = li_i + 1
#   END FOREACH
#   CALL la_gzou.deleteElement(li_i)
#
#   FOR li_i =  1 TO la_gzou.getLength()
#
#       IF la_gzou[li_i].gzou003 != g_dbs_old THEN
#          IF NOT cl_db_setconnect(la_gzou[li_i].gzou003) THEN        
#             CALL cl_db_connect(la_gzou[li_i].gzou003, FALSE)
#          END IF 
#
#          LET ls_sql = "SELECT ooag011 FROM ooag_t WHERE ooagent = ? AND ooag001 = ?"
#          PREPARE azzq080_license_collect_ooag_pre2 FROM ls_sql
#          LET ls_sql = "SELECT gzxa007 FROM gzxa_t WHERE gzxaent = ? AND gzxa001 = ?"
#          PREPARE azzq080_license_collect_gzxa_pre2 FROM ls_sql
#          FOR li_idx = 1 TO g_gzwl2_d.getLength()
#              IF cl_null(g_gzwl2_d[li_idx].gzwl002_desc) THEN 
#                 EXECUTE azzq080_license_collect_ooag_pre2 USING g_aid[li_idx].gzwl012, g_aid[li_idx].gzwl009 INTO g_gzwl2_d[li_idx].gzwl002_desc
#              END IF 
#
#              IF cl_null(g_gzwl2_d[li_idx].gzxa007) THEN 
#                 #左側參雜fglWrt取出且gzwl並不存在的資訊, fglWrt的電腦id可能被誤解讀為使用者帳號
#                 EXECUTE azzq080_license_collect_gzxa_pre2 USING g_aid[li_idx].gzwl012, g_gzwl2_d[li_idx].gzwl002 INTO g_gzwl2_d[li_idx].gzxa007
#              END IF 
#          END FOR  
#       END IF 
#
#   END FOR 
# 
#
#   #最後在連回原來DB
#   IF cl_db_setconnect(g_dbs_old) THEN 
#
#   END IF 
#   #181207-00005#1--end--   
#
##180822-00047#1--start-- 
#LET ls_msg_desc = cl_getmsg("azz-00488",g_lang)         #"此使用者由其他區域登入,無法取得使用者名稱"
#FOR li_idx = 1 TO g_gzwl2_d.getLength()
#   #181207-00005#1--start--
#   #IF g_gzwl2_d[li_idx].gzwl002_desc = "Developer" THEN 
#   IF g_gzwl2_d[li_idx].gzwl002_desc = "Developer" AND g_gzwl2_d[li_idx].gzwl015 != cl_get_ap_hostname() THEN 
#       LET g_gzwl2_d[li_idx].gzwl002_desc = ls_msg_desc
#   END IF 
#   #181207-00005#1--end--
#END FOR  
##180822-00047#1--end--
#
#   #181117-00002#1--start--
#   FOR li_idx = 1 TO g_gzwl2_d.getLength()
#       IF g_gzwl2_d[li_idx].gzwl002_desc = "Developer" THEN 
#          LET g_gzwl2_d[li_idx].gzwl002_desc = ""
#       END IF 
#   END FOR
#   #181117-00002#1--end--
#   CALL lc_write.close()
#   CALL lc_write2.close()                                #200827-00023#1 add
   #240628-00031#1 mark --- end ---
   
FOR li_idx = 1 TO g_gzwl2_d.getLength()
    IF g_t100debug >=6 THEN   
      DISPLAY"***************************************************************************" 
      DISPLAY " g_gzwl2_d gzwl002:", g_gzwl2_d[li_idx].gzwl002
      DISPLAY " g_gzwl2_d gzwl002_desc:", g_gzwl2_d[li_idx].gzwl002_desc
      DISPLAY " g_gzwl2_d gzwl005:", g_gzwl2_d[li_idx].gzwl005
      DISPLAY " g_gzwl2_d gzwl005_desc:", g_gzwl2_d[li_idx].gzwl005_desc
      DISPLAY " g_gzwl2_d gzwl006:", g_gzwl2_d[li_idx].gzwl006
      DISPLAY " g_gzwl2_d gzwl015:", g_gzwl2_d[li_idx].gzwl015
      DISPLAY " g_gzwl2_d gzxa007:", g_gzwl2_d[li_idx].gzxa007
      DISPLAY"***************************************************************************"
      DISPLAY " "      
   END IF
END FOR  

END FUNCTION
################################################################################
# Descriptions...: 定義fglWrt單身分析的keyword
# Memo...........: 
# Usage..........: CALL azzq080_license_keyword_define()
# Input parameter: None
# Return code....: None
# Date & Author..: 2015/06/30 By Saki
# Modify.........:
################################################################################
PRIVATE FUNCTION azzq080_license_keyword_define()
   DEFINE ls_str STRING
   DEFINE li_i SMALLINT

   #建立關鍵字集
   LET ls_str = "[",
                "{\"keyname\":\"License\",\"field\":\"license\"},",
                "{\"keyname\":\"Product\",\"field\":\"product\"},",
                "{\"keyname\":\"Type\",\"field\":\"type\"},",
                "{\"keyname\":\"Users\",\"field\":\"status\"},",
                "{\"keyname\":\"CPU(s)\",\"field\":\"status\"}",
                "]"
   CALL util.JSON.parse(ls_str, g_keyword_h)
   FOR li_i = 1 TO g_keyword_h.getLength()
       LET g_keyword_h[li_i].wordnum = g_keyword_h[li_i].keyname.getLength()
   END FOR
   LET ls_str = "[",
                "{\"keyname\":\"Process Id\",\"field\":\"pid\"},",
                "{\"keyname\":\"user-name\",\"field\":\"name\"},",
                "{\"keyname\":\"host-addr\",\"field\":\"ip\"},",
                "{\"keyname\":\"cx-mode\",\"field\":\"exemode\"},",
                "{\"keyname\":\"user-agent\",\"field\":\"agent\"},",
#                "{\"keyname\":\"app-name\",\"field\":\"appname\"}",          #200827-00023#1 mark
                "{\"keyname\":\"host-name\",\"field\":\"appname\"}",          #200827-00023#1 add
                "]"
   CALL util.JSON.parse(ls_str, g_keyword_d)
   FOR li_i = 1 TO g_keyword_d.getLength()
       LET g_keyword_d[li_i].wordnum = g_keyword_d[li_i].keyname.getLength()
   END FOR
END FUNCTION

################################################################################
# Descriptions...: 定義fglWrt單身分析的keyword
# Memo...........: 
# Usage..........: CALL azzq080_license_keyword_define()
# Input parameter: None
# Return code....: None
# Date & Author..: 2015/06/30 By Saki
# Modify.........:
################################################################################
PRIVATE FUNCTION azzq080_fglwrt_display()
   DEFINE lwin_curr ui.Window
   DEFINE lfrm_curr ui.Form

   OPEN WINDOW azzq080_fglwrt_display_w WITH FORM cl_ap_formpath("azz","azzq080_s02")
      ATTRIBUTES(STYLE="layoutwin")
   CALL cl_ui_init()
   LET lwin_curr = ui.Window.getCurrent()
   LET lfrm_curr = lwin_curr.getForm()

   DISPLAY g_fglwrt TO formonly.fglwrt
   MENU ""
      ON ACTION close
         EXIT MENU
      ON ACTION exit
         EXIT MENU
   END MENU

   CLOSE WINDOW azzq080_fglwrt_display_w
END FUNCTION

################################################################################
# Descriptions...: 動態新增Style, 以便fglWrt視窗有find功能
# Memo...........: 
# Usage..........: CALL azzq080_add_style()
# Input parameter: None
# Return code....: None
# Date & Author..: 2015/07/16 By Saki
# Modify.........:
################################################################################
PRIVATE FUNCTION azzq080_add_style()
   DEFINE lnode_root,lnode_item,lnode_child,lnode_att om.DomNode
   DEFINE llst_styles om.NodeList
   DEFINE li_i SMALLINT

   LET lnode_root = ui.Interface.getRootNode()
   LET llst_styles = lnode_root.selectByTagName("StyleList")
   FOR li_i = 1 TO llst_styles.getLength()
       LET lnode_item = llst_styles.item(li_i)
       LET lnode_child = lnode_item.createChild("Style")
       CALL lnode_child.setAttribute("name", "TextEdit.integratedSearch")
       LET lnode_att = lnode_child.createChild("StyleAttribute")
       CALL lnode_att.setAttribute("name", "integratedSearch")
       CALL lnode_att.setAttribute("value", "yes")
       EXIT FOR
   END FOR
END FUNCTION
################################################################################
# Descriptions...: 刪除線上process, 並記錄log
# Memo...........: 如果提供不同主機hostname, 是否能刪掉其他主機上的process
# Usage..........: CALL azzq080_kill_process()
# Input parameter: None
# Return code....: None
# Date & Author..: 2015/07/31 By Saki
# Modify.........:
################################################################################
PRIVATE FUNCTION azzq080_kill_process()
   DEFINE ls_ttpswd     STRING
   DEFINE li_idx        LIKE type_t.num5
   DEFINE li_sel_flag   BOOLEAN
   DEFINE ls_cmd        STRING
   DEFINE li_success    BOOLEAN
   DEFINE ls_result     STRING
   DEFINE li_allow_del  BOOLEAN
   DEFINE lr_gzwl       RECORD LIKE gzwl_t.*
   DEFINE lr_gzwn       DYNAMIC ARRAY OF RECORD LIKE gzwn_t.*
   DEFINE ls_sql        STRING
   DEFINE ls_sql_sel    STRING
   DEFINE ls_sql_ins    STRING
   DEFINE ls_sql_del    STRING
   DEFINE ld_date       LIKE type_t.dat
   DEFINE ls_date       STRING
   DEFINE ls_slip       STRING

   LET li_success = TRUE
   LET ls_sql_sel = "SELECT gzwl001, gzwl002, gzwl003, gzwl004, gzwl005, gzwl006, gzwl007,",
                    "       gzwl008, gzwl009, gzwl010, gzwl011, gzwl012, gzwl013, gzwl014,",
                    "       gzwl015 FROM gzwl_t",
                    " WHERE gzwl001 = ?"
   LET ls_sql_ins = "INSERT INTO gzwn_t(gzwndocno, gzwncrtid, gzwncrtdt, gzwn001,",
                    "                   gzwn002, gzwn003, gzwn004, gzwn005, gzwn006,",
                    "                   gzwn007, gzwn008, gzwn009, gzwn010, gzwn011,",
                    "                   gzwn012, gzwn013, gzwn014, gzwn015, gzwn099)",
                    " VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
   LET ls_sql_del = "DELETE FROM gzwl_t WHERE gzwl001 = ?"

   #有選擇的話才開始進行步驟
   LET li_sel_flag = FALSE
   FOR li_idx = 1 TO g_gzwl_d.getLength()
       IF g_gzwl_d[li_idx].sel = "Y" THEN
          LET li_sel_flag = TRUE
          EXIT FOR
       END IF
   END FOR

   IF li_sel_flag THEN
      INITIALIZE ls_ttpswd TO NULL
      IF cl_ask_confirm("lib-00208") THEN
         CALL azzq080_01() RETURNING li_allow_del, ls_ttpswd
         IF li_allow_del AND NOT cl_null(ls_ttpswd) THEN
            IF NOT cl_db_setconnect("dscom") THEN END IF
            PREPARE azzq080_kill_process_getinfo1_pre FROM ls_sql_sel
            PREPARE azzq080_kill_process_log1_pre FROM ls_sql_ins
            PREPARE azzq080_kill_process1_pre FROM ls_sql_del
            FOR li_idx = 1 TO g_gzwl_d.getLength()
                IF g_gzwl_d[li_idx].sel = "Y" THEN
                   LET ls_cmd = "echo \"",ls_ttpswd.trim(),"\"|sudo -S -k -u root kill -15 ", g_gzwl_d[li_idx].gzwl008," 2>&1 >&-"
                   RUN ls_cmd RETURNING ls_result
                   IF ls_result != "0" THEN
                      LET li_success = FALSE
                   ELSE
                      #寫入Log
                      INITIALIZE lr_gzwl.* TO NULL
                      EXECUTE azzq080_kill_process_getinfo1_pre USING g_gzwl_d[li_idx].gzwl001
                         INTO lr_gzwl.gzwl001, lr_gzwl.gzwl002, lr_gzwl.gzwl003, lr_gzwl.gzwl004,
                              lr_gzwl.gzwl005, lr_gzwl.gzwl006, lr_gzwl.gzwl007, lr_gzwl.gzwl008,
                              lr_gzwl.gzwl009, lr_gzwl.gzwl010, lr_gzwl.gzwl011, lr_gzwl.gzwl012,
                              lr_gzwl.gzwl013, lr_gzwl.gzwl014, lr_gzwl.gzwl015
                      IF cl_null(lr_gzwl.gzwl001) THEN
                         #這時候應該是沒有進入gzwl的process
                         LET lr_gzwl.gzwl003 = g_gzwl_d[l_ac].gzwl003_desc
                         LET lr_gzwl.gzwl007 = g_gzwl_d[l_ac].gzwl007
                         LET lr_gzwl.gzwl008 = g_gzwl_d[l_ac].gzwl008
                      END IF
                      EXECUTE azzq080_kill_process1_pre USING g_gzwl_d[li_idx].gzwl001
                      #取單號, 可能會搶號, 又不能放棄紀錄, 因為已經被砍掉了, 迴圈很恐怖, 鎖Table也很恐怖, 先不管搶號的insert不成功
                      LET ld_date = cl_get_current()
                      LET ls_date = ld_date USING 'yyyymmdd'
                      LET ls_sql = "SELECT MAX(gzwndocno) FROM gzwn_t WHERE gzwndocno LIKE '", ls_date,"-%'"
                      PREPARE azzq080_kill_process_docno_pre FROM ls_sql
                      EXECUTE azzq080_kill_process_docno_pre INTO lr_gzwn[li_idx].gzwndocno
                      IF cl_null(lr_gzwn[li_idx].gzwndocno) THEN
                         LET ls_slip = "1"
                      ELSE
                         LET ls_slip = lr_gzwn[li_idx].gzwndocno
                         LET ls_slip = ls_slip.subString(10, ls_slip.getLength())
                         LET ls_slip = ls_slip + 1
                      END IF
                      LET lr_gzwn[li_idx].gzwndocno = ls_date, "-", ls_slip USING "&&&&&&"
                      LET lr_gzwn[li_idx].gzwncrtid = g_user
                      LET lr_gzwn[li_idx].gzwncrtdt = cl_get_current()
                      LET lr_gzwn[li_idx].gzwn099 = "1"
                      LET lr_gzwn[li_idx].gzwn001 = lr_gzwl.gzwl001
                      LET lr_gzwn[li_idx].gzwn002 = lr_gzwl.gzwl002
                      LET lr_gzwn[li_idx].gzwn003 = lr_gzwl.gzwl003
                      LET lr_gzwn[li_idx].gzwn004 = lr_gzwl.gzwl004
                      LET lr_gzwn[li_idx].gzwn005 = lr_gzwl.gzwl005
                      LET lr_gzwn[li_idx].gzwn006 = lr_gzwl.gzwl006
                      LET lr_gzwn[li_idx].gzwn007 = lr_gzwl.gzwl007
                      LET lr_gzwn[li_idx].gzwn008 = lr_gzwl.gzwl008
                      LET lr_gzwn[li_idx].gzwn009 = lr_gzwl.gzwl009
                      LET lr_gzwn[li_idx].gzwn010 = lr_gzwl.gzwl010
                      LET lr_gzwn[li_idx].gzwn011 = lr_gzwl.gzwl011
                      LET lr_gzwn[li_idx].gzwn012 = lr_gzwl.gzwl012
                      LET lr_gzwn[li_idx].gzwn013 = lr_gzwl.gzwl013
                      LET lr_gzwn[li_idx].gzwn014 = lr_gzwl.gzwl014
                      LET lr_gzwn[li_idx].gzwn015 = lr_gzwl.gzwl015
                      EXECUTE azzq080_kill_process_log1_pre
                        USING lr_gzwn[li_idx].gzwndocno, lr_gzwn[li_idx].gzwncrtid, lr_gzwn[li_idx].gzwncrtdt,
                              lr_gzwn[li_idx].gzwn001, lr_gzwn[li_idx].gzwn002, lr_gzwn[li_idx].gzwn003, lr_gzwn[li_idx].gzwn004,
                              lr_gzwn[li_idx].gzwn005, lr_gzwn[li_idx].gzwn006, lr_gzwn[li_idx].gzwn007, lr_gzwn[li_idx].gzwn008,
                              lr_gzwn[li_idx].gzwn009, lr_gzwn[li_idx].gzwn010, lr_gzwn[li_idx].gzwn011, lr_gzwn[li_idx].gzwn012,
                              lr_gzwn[li_idx].gzwn013, lr_gzwn[li_idx].gzwn014, lr_gzwn[li_idx].gzwn015, lr_gzwn[li_idx].gzwn099
                   END IF
                END IF
            END FOR
         END IF
      END IF
   END IF

   IF NOT cl_db_setconnect(g_dbs_old) THEN END IF
   IF li_success THEN
      #依照dscom裡面所砍掉的、紀錄的log一樣在原DB內執行, 砍不掉的可能是不同的DB, 也不管他, 讓排程工具處理
      PREPARE azzq080_kill_process_log2_pre FROM ls_sql_ins
      PREPARE azzq080_kill_process2_pre FROM ls_sql_del
      FOR li_idx = 1 TO lr_gzwn.getLength()
          IF lr_gzwn[li_idx].gzwndocno IS NULL THEN
             CONTINUE FOR
          END IF
          EXECUTE azzq080_kill_process2_pre USING lr_gzwn[li_idx].gzwn001
          IF SQLCA.sqlcode OR SQLCA.sqlerrd[3] = 0 THEN
          ELSE
             EXECUTE azzq080_kill_process_log2_pre
               USING lr_gzwn[li_idx].gzwndocno, lr_gzwn[li_idx].gzwncrtid, lr_gzwn[li_idx].gzwncrtdt,
                     lr_gzwn[li_idx].gzwn001, lr_gzwn[li_idx].gzwn002, lr_gzwn[li_idx].gzwn003, lr_gzwn[li_idx].gzwn004,
                     lr_gzwn[li_idx].gzwn005, lr_gzwn[li_idx].gzwn006, lr_gzwn[li_idx].gzwn007, lr_gzwn[li_idx].gzwn008,
                     lr_gzwn[li_idx].gzwn009, lr_gzwn[li_idx].gzwn010, lr_gzwn[li_idx].gzwn011, lr_gzwn[li_idx].gzwn012,
                     lr_gzwn[li_idx].gzwn013, lr_gzwn[li_idx].gzwn014, lr_gzwn[li_idx].gzwn015, lr_gzwn[li_idx].gzwn099
          END IF
      END FOR
   ELSE
      INITIALIZE g_errparam TO NULL 
      LET g_errparam.code   = "lib-00209"
      LET g_errparam.popup  = TRUE 
      CALL cl_err()
   END IF
   CALL azzq080_b_fill()
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzq080_chk_prosess(ps_pid)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2018/12/10 By 01856
# Modify.........:
################################################################################
PRIVATE FUNCTION azzq080_chk_prosess(ps_pid)
   DEFINE ps_pid        STRING 
   DEFINE ls_cmd       STRING 
   DEFINE lc_channel   base.Channel
   DEFINE ls_str       STRING  
   DEFINE li_cnt       LIKE type_t.num5  

   LET li_cnt = 0
   LET ls_cmd = "ps -ef | grep -v grep | grep ' ", ps_pid, " ' | grep fglrun | awk '{print $2}' | wc -l "   
   LET lc_channel = base.Channel.create()   
   CALL lc_channel.openPipe(ls_cmd, "r")
   CALL lc_channel.setDelimiter("")
   WHILE lc_channel.read(ls_str)
      LET li_cnt = ls_str
      EXIT WHILE
   END WHILE   
   #DISPLAY "chk_prosess pid:",ps_pid ," 筆數:",li_cnt
   IF li_cnt > 0 THEN 
      RETURN TRUE 
   ELSE
      RETURN FALSE  
   END IF 
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzq080_get_grep_prog(pr_others)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2018/12/10 By 01856
# Modify.........:
################################################################################
PRIVATE FUNCTION azzq080_get_grep_prog(pr_others)
   DEFINE pr_others    type_license_d
   DEFINE lc_channel   base.Channel
   DEFINE ls_cmd        STRING 
   DEFINE ls_str        STRING 
   DEFINE li_idx        LIKE type_t.num5
   DEFINE l_42r_dir     STRING 
   DEFINE ls_tmp        STRING 
   DEFINE li_ok         LIKE type_t.num5
   DEFINE li_chk        LIKE type_t.num5
   DEFINE ls_time       STRING 
   DEFINE ls_ent        STRING 
   DEFINE ld_hour_to_min      DATETIME YEAR TO SECOND

   DEFINE l_interval    INTERVAL HOUR TO SECOND   
                        
   LET lc_channel = base.Channel.create()
   IF cl_null(FGL_GETENV("DEV_VER")) THEN 
      LET l_42r_dir = "42r"
   ELSE 
      LET l_42r_dir = "42r",FGL_GETENV("DEV_VER")
   END IF  

   LET ls_cmd = "ps -ef | grep -v grep | grep -v ' sh -c ' | grep ' ", pr_others.pid, " ' | grep fglrun "  
   LET li_idx = 7
   CALL lc_channel.openPipe(ls_cmd, "r")
   CALL lc_channel.setDelimiter("")
   #DISPLAY " ls_cmd:",ls_cmd
   WHILE lc_channel.read(ls_str)
      #搜尋fglrun
      IF NOT ls_str.getIndexOf("fglrun",1) THEN 
         EXIT WHILE
      ELSE  
         DISPLAY "get_grep_prog ls_str:",ls_str 
         IF ls_str.getIndexOf("fglrun-bin -d",1) THEN
            LET ls_str = ""
            EXIT WHILE
         END IF 
         #DISPLAY " ls_str:",ls_str
         LET li_ok = FALSE 
         WHILE TRUE 
            LET ls_cmd = "ps -ef | grep -v grep | grep -v ' sh -c ' | grep ' ", pr_others.pid, " ' | grep fglrun | awk '{print $",li_idx USING "<<<" ,"}'" 
            CALL lc_channel.openPipe(ls_cmd, "r")
            CALL lc_channel.setDelimiter("")
            WHILE lc_channel.read(ls_str)

            
               #搜尋 fgldeb
               IF ls_str.getIndexOf("fgldeb",1) THEN 
                  LET ls_tmp = ls_str
                  LET li_idx = li_idx + 1 
                  EXIT WHILE
               END IF

               #搜尋 fgldeb
               IF ls_str.getIndexOf("fglrun-bin ",1) THEN 
                  LET ls_tmp = ls_str.subString(ls_str.getIndexOf("fglrun-bin",1),ls_str.getLength())
                  #LET li_idx = li_idx + 1 
                  LET ls_str = ls_tmp 
                  LET li_ok =  TRUE 
                  EXIT WHILE
               END IF
               
               #搜尋42r
               IF ls_str.getIndexOf(l_42r_dir,1) THEN 
                   LET li_ok = TRUE 
               ELSE 
                  #搜尋不到 42r 往下一個
                  LET li_idx = li_idx + 1   
               END IF 
               EXIT WHILE 
            END WHILE  #end while TRUE 內層
            IF li_ok THEN 
               EXIT WHILE 
            END IF 
         END WHILE  #end while TRUE 中層
      END IF 
      EXIT WHILE
   END WHILE  #end while TRUE 外層
   
   IF li_ok THEN 
      #假如有fgldeb(r.d) 就多一個fgldeb ，沒有就是 r.r 
      LET ls_str = ls_tmp ,1 SPACE , ls_str
      LET li_ok = FALSE
      IF cl_null(ls_tmp) THEN 

         #取出企業編號
         LET ls_cmd = "ps -ef | grep -v grep | grep -v ' sh -c ' | grep ' ", pr_others.pid, " ' | grep fglrun | awk '{print $10}' " 
         CALL lc_channel.openPipe(ls_cmd, "r")
         CALL lc_channel.setDelimiter("")
         LET ls_tmp = "" 
         WHILE lc_channel.read(ls_tmp)
            LET ls_ent = ls_tmp
         END WHILE
         IF ls_ent.getIndexOf("static_ent",1) THEN 
            LET ls_ent = ls_ent.subString(ls_ent.getIndexOf("static_ent",1)+12,ls_ent.getLength()-1)
            #DISPLAY "static_ent ls_ent:",ls_ent
         END IF 
      END IF 

      
   END IF 

   IF cl_null(ls_str) THEN 
      CALL azzq080_chk_prosess(pr_others.pid) RETURNING li_chk 
      IF NOT li_chk THEN 
         DISPLAY " get_grep_prog and chk_prosess li_chk:",li_chk ," ls_str:",ls_str
         
         LET ls_str = ""
         LET ls_ent = ""
         RETURN ls_str,ls_ent
      END IF 
   END IF 
   RETURN ls_str,ls_ent
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzq080_show_proc_cnt_and_lim(pc_gzwl005)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2019/01/04 By 01856
# Modify.........:
################################################################################
PRIVATE FUNCTION azzq080_show_proc_cnt_and_lim(pc_gzwl005)
   DEFINE pc_gzwl005   LIKE gzwl_t.gzwl005 #授權資訊

   CALL cl_process_user_limit(pc_gzwl005) RETURNING g_limit_cnt, g_detail_cnt, g_exec_notover

   #160726-00036#1 執行作業數已經不只gzwl_t內的資料
   LET g_detail_cnt = g_gzwl_d.getLength()
   
   IF g_limit_cnt = 32767 THEN
      DISPLAY g_detail_cnt, "∞" TO FORMONLY.process_count, FORMONLY.process_limit
   ELSE
      DISPLAY g_detail_cnt, g_limit_cnt TO FORMONLY.process_count, FORMONLY.process_limit
   END IF
END FUNCTION

################################################################################
# Descriptions...: 再次檢核左側清單授權
# Memo...........:
# Usage..........: CALL azzq080_chk_g_gzwl2_d_data()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2019/01/04 By 01856
# Modify.........:
################################################################################
PRIVATE FUNCTION azzq080_chk_g_gzwl2_d_data()
   DEFINE li_i       LIKE type_t.num5
   DEFINE li_cnt     LIKE type_t.num5 
   DEFINE li_count   LIKE type_t.num5 
   DEFINE lr_license    type_license_d     
   DEFINE li_chk     LIKE type_t.num5    #190503-00015#1   add
   
   #先切到 dscom
   IF cl_db_setconnect("dscom") THEN END IF

   LET g_sql = "SELECT COUNT(1) FROM licinfo ",
               " WHERE fdid = ?  AND ip = ? AND host = ? "
               
   PREPARE azzq080_chk_g_gzwl2_d_data_pre FROM g_sql
   DECLARE azzq080_chk_g_gzwl2_d_data_curs CURSOR FOR azzq080_chk_g_gzwl2_d_data_pre

   LET g_sql = "SELECT pid, host ,fdid,ip FROM licinfo ",
               " WHERE fdid = ?  AND ip = ? AND host = ? "
   PREPARE azzq080_chk_g_gzwl2_d_data_pre2 FROM g_sql
   DECLARE azzq080_chk_g_gzwl2_d_data_curs2 CURSOR FOR azzq080_chk_g_gzwl2_d_data_pre2               
   
   #190503-00015#1#1--start--
   LET g_sql = "SELECT COUNT(1) FROM licinfo ",
               " WHERE fdid = ?  AND host = ? "
               
   PREPARE azzq080_chk_g_gzwl2_d_data_pre3 FROM g_sql
   DECLARE azzq080_chk_g_gzwl2_d_data_curs3 CURSOR FOR azzq080_chk_g_gzwl2_d_data_pre3   

   LET g_sql = "SELECT pid, host ,fdid,ip FROM licinfo ",
               " WHERE fdid = ? AND host = ? "
   PREPARE azzq080_chk_g_gzwl2_d_data_pre4 FROM g_sql
   DECLARE azzq080_chk_g_gzwl2_d_data_curs4 CURSOR FOR azzq080_chk_g_gzwl2_d_data_pre4   
   #190503-00015#1#1--start--   

   FOR li_i = 1 TO g_gzwl2_d.getLength()
       #190503-00015#1#1--start--   
       DISPLAY " ip    gzwl006:",g_gzwl2_d[li_i].gzwl006  
       DISPLAY " AP主機 gzwl015:",g_gzwl2_d[li_i].gzwl015
       DISPLAY " hostname:",cl_get_ap_hostname()
       DISPLAY " FrontEndId gzwl005_desc:",g_gzwl2_d[li_i].gzwl005_desc
       #190503-00015#1#1--end--
       
       IF g_gzwl2_d[li_i].gzwl015 = cl_get_ap_hostname() THEN 
          #190503-00015#1#1--start-- 
          IF NOT cl_chk_ipv4(g_gzwl2_d[li_i].gzwl006) THEN
             LET li_chk = FALSE  
             LET li_count = 0
             LET li_cnt = 0
                                                            #FrontEndId ,hostname
             FOREACH azzq080_chk_g_gzwl2_d_data_curs3 USING g_gzwl2_d[li_i].gzwl005_desc,g_gzwl2_d[li_i].gzwl015 INTO li_cnt
                #fglWrt 有資料 ，判斷有沒有存在主機PID
                IF li_cnt > 0 THEN
                   LET li_chk = TRUE  
                   FOREACH azzq080_chk_g_gzwl2_d_data_curs4 USING g_gzwl2_d[li_i].gzwl005_desc,g_gzwl2_d[li_i].gzwl015 
                      INTO lr_license.pid,lr_license.host,lr_license.ip 
                      DISPLAY " 判斷有沒有存在主機PID  :" 
                      DISPLAY " pid  :",lr_license.pid 
                      DISPLAY " host :",lr_license.host        
                      DISPLAY " ip :",lr_license.ip  
                      DISPLAY " 判斷有沒有存在主機PID  :"    
                      IF azzq080_chk_prosess(lr_license.pid) THEN
                         LET li_count = li_count + 1 
                      ELSE 
                         
                      END IF 
                  END FOREACH
                END IF 
             END FOREACH 
             #沒有資料就刪除就刪除左側清單單身資料
             IF NOT li_chk OR li_count = 0 THEN 
                DISPLAY " 準備刪除 delete FrontEndId:",g_gzwl2_d[li_i].gzwl005_desc ," ip:",g_gzwl2_d[li_i].gzwl006
                CALL g_gzwl2_d.deleteElement(li_i)
                LET li_i = li_i - 1                        
             END IF 
             CONTINUE FOR 
          END IF
          #190503-00015#1#1--end-- 
          
          FOREACH azzq080_chk_g_gzwl2_d_data_curs USING g_gzwl2_d[li_i].gzwl005_desc,g_gzwl2_d[li_i].gzwl006,g_gzwl2_d[li_i].gzwl015 INTO li_cnt

             IF li_cnt = 0 THEN 
                
                CALL g_gzwl2_d.deleteElement(li_i)
                LET li_i = li_i - 1
                CONTINUE FOR  
             ELSE 
                LET li_count = 0
                FOREACH azzq080_chk_g_gzwl2_d_data_curs2 USING g_gzwl2_d[li_i].gzwl005_desc,g_gzwl2_d[li_i].gzwl006,g_gzwl2_d[li_i].gzwl015 
                   INTO lr_license.pid,lr_license.host,lr_license.ip             
                    
                    IF NOT azzq080_chk_prosess(lr_license.pid) THEN
                       LET li_count = li_count + 1 
                    END IF 
                END FOREACH 
                #假如PID 都不存在 就刪除左側清單單身資料

                IF li_count = li_cnt THEN 
                   CALL g_gzwl2_d.deleteElement(li_i) 
                   LET li_i = li_i - 1
                   CONTINUE FOR 
                END IF    
             END IF
           #END IF 
          END FOREACH      
       END IF 
   END FOR 
   #再切回原來 DB
   IF cl_db_setconnect(g_dbs_old) THEN END IF
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzq080_counting_number()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2019/01/04 By 01856
# Modify.........:
################################################################################
PRIVATE FUNCTION azzq080_counting_number()
   DEFINE ls_log_order STRING

   LET g_log_order = g_log_order + 1             #每次按下重新整理,就自動將順序加上1.
   LET ls_log_order = g_log_order USING "&&&"
   LET g_log = FGL_GETPID() || ls_log_order      #改變g_log

END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzq080_delete_gzwl()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 2019/01/04 By 01856
# Modify.........:
################################################################################
PRIVATE FUNCTION azzq080_delete_gzwl()
   DEFINE li_cnt LIKE type_t.num5

   #先切到 dscom
   IF cl_db_setconnect("dscom") THEN END IF
   FOR li_cnt = 1 TO g_gzwl_d_bck.getLength()
       DISPLAY " dscom g_gzwl_d_bck PID 不存在:",g_gzwl_d_bck[li_cnt].gzwl008 
       #DELETE FROM gzwl_t WHERE gzwl008 = g_gzwl_d_bck[li_cnt].gzwl008  
   END FOR 

   IF cl_db_setconnect(g_dbs_old) THEN END IF
   FOR li_cnt = 1 TO g_gzwl_d_bck.getLength()
       DISPLAY " g_gzwl_d_bck PID 不存在:",g_gzwl_d_bck[li_cnt].gzwl008  
       #DELETE FROM gzwl_t WHERE gzwl008 = g_gzwl_d_bck[li_cnt].gzwl008      
   END FOR 
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzq080_version()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PRIVATE FUNCTION azzq080_version()
   DEFINE lchannel_read  base.Channel
   DEFINE ls_tmp         STRING
   DEFINE ls_tmp2        STRING
   DEFINE ls_backup      STRING 
   DEFINE ls_source      STRING
   DEFINE li_idx         LIKE type_t.num5  
   DEFINE li_idx2        LIKE type_t.num5 
   DEFINE ls_file_path   STRING   

   LET ls_file_path = os.Path.join(os.Path.join(os.Path.join(FGL_GETENV("ERP"), "azz"),"4gl"),g_code)
   LET ls_file_path = ls_file_path,".4gl"

   #檢查檔案是否存在
   IF NOT os.Path.exists(ls_file_path) THEN
      LET ls_tmp2 = "4gl 檔案不存在:",ls_file_path
      RETURN ls_tmp2
   END IF
 
   LET lchannel_read = base.Channel.create()
   CALL lchannel_read.setDelimiter("")
   CALL lchannel_read.openFile(ls_file_path CLIPPED, "r" )

   WHILE TRUE
      LET ls_source = lchannel_read.readLine()
      IF lchannel_read.isEof() THEN
         EXIT WHILE
      END IF
      LET ls_tmp = ls_source.trim()
      IF ls_tmp.getIndexOf("#+ Modifier...:",1) THEN 
      ELSE
         IF ls_tmp.getIndexOf("IMPORT os",1) OR ls_backup.getIndexOf("IMPORT os",1) THEN 
            EXIT WHILE  
         END IF
         CONTINUE WHILE                        #220801-00038#1 add         
      END IF 

      LET ls_source = ls_source.trim()
      IF ls_source.getLength() = 0 THEN 

      ELSE 
         LET ls_backup = ls_source  
      END IF 
         
   END WHILE  

   IF ls_backup.getIndexOf("Modifier...:",1) THEN 

      LET li_idx = ls_backup.getIndexOf("Modifier...:",1)
      LET li_idx2 = ls_backup.getIndexOf("No.",li_idx+1)
      LET li_idx = ls_backup.getIndexOf(" ",li_idx2+1)
      LET ls_tmp2 = ls_backup.subString(li_idx2,li_idx)
   END IF  

   RETURN ls_tmp2
END FUNCTION
#220801-00038#1 add
################################################################################
# Descriptions...: 描述说明
# Memo...........: 匯出流量統計圖表
# Usage..........: CALL azzq080_export_usage(p_date,p_mode)
# Input parameter: p_date   統計流量日期
#                : mode     統計流量模式: 0:by day, 1:by month
# Date & Author..: 2022-08-11 By 05598
# Modify.........:
################################################################################
PRIVATE FUNCTION azzq080_export_usage(p_date,p_mode)
   DEFINE p_date           DATETIME YEAR TO SECOND
   DEFINE p_mode           STRING
   DEFINE l_date           STRING
   DEFINE l_msg            STRING
   DEFINE l_str            STRING
   DEFINE l_cmd            STRING
   DEFINE l_usagelog       STRING
   DEFINE l_current        STRING
   DEFINE l_filename       STRING
   DEFINE l_filepath       STRING
   DEFINE l_url_path       STRING
   DEFINE l_date_a         DYNAMIC ARRAY OF STRING
   DEFINE l_value          DYNAMIC ARRAY OF STRING
   DEFINE l_hour           DYNAMIC ARRAY OF STRING
   DEFINE l_max            INTEGER
   DEFINE l_int            INTEGER
   DEFINE l_strindex       INTEGER 
   DEFINE l_circle_row     INTEGER 
   DEFINE l_circle_max     INTEGER 
   DEFINE lb_return        BOOLEAN
   DEFINE l_result         LIKE type_t.num5
   DEFINE l_channel        base.Channel
   DEFINE sxWorkBook       SXSSFWorkbook
   DEFINE sheet            Sheet
   DEFINE row              Row
   DEFINE cell             Cell
   DEFINE outFile          FileOutputStream
   #250328-00007#1 add --- start ---
   DEFINE l_ctype          INTEGER
   DEFINE l_ac             INTEGER 
   DEFINE l_i              INTEGER 
   DEFINE l_j              INTEGER 
   DEFINE l_d              INTEGER 
   DEFINE l_logfile        STRING
   DEFINE l_logfile2       STRING
   DEFINE l_tmp            STRING
   DEFINE l_tmpi           INTEGER
   DEFINE l_rdata         DYNAMIC ARRAY OF RECORD
         lc_date           STRING,
         lc_time           STRING,
         lc_used           INTEGER,
         lc_unlogin        INTEGER,
         lc_ck             STRING
      END RECORD
   DEFINE l_nums           DYNAMIC ARRAY OF RECORD
         lc_ck             STRING,
         lc_used           INTEGER
      END RECORD
   
   LET l_channel = base.Channel.create()
   LET lb_return = TRUE
   #抓取 excel 樣板
   CASE p_mode
      WHEN "1"
         LET l_filepath = FGL_GETENV("FBIN"),"/42m/azzq080_usage_day.xlsx" 
      WHEN "2"
         LET l_filepath = FGL_GETENV("FBIN"),"/42m/azzq080_usage_week.xlsx" 
      WHEN "3"
         LET l_filepath = FGL_GETENV("FBIN"),"/42m/azzq080_usage_month.xlsx" 
   END CASE
   LET l_filename = "license_usage_", l_current,".xlsx"
   LET l_cmd = "cp -rfp ", l_filepath, " ", FGL_GETENV("TEMPDIR"), "/", l_filename
   RUN l_cmd
   LET l_filepath = FGL_GETENV("TEMPDIR"),"/", l_filename
   
   #處理 excel 檔案初始化
   LET sxWorkBook = cl_poi_open_sxssf(l_filepath)
   IF sxWorkBook IS NULL THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = '-461'
      LET g_errparam.extend = ''
      LET g_errparam.popup = TRUE
      CALL cl_err()    
      LET lb_return = FALSE
   END IF
   
   LET sheet = sxWorkBook.getSheetAt(0)
   #找不到Sheet彈窗警告並離開
   IF sheet IS NULL THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = 'azz-01169'
      LET g_errparam.extend = ''
      LET g_errparam.popup = TRUE
      CALL cl_err()        
      LET lb_return = FALSE
   END IF
   
   #移除工作表非空行
   CALL cl_poi_removeAllRows(0)
   
   #建立樣式與資料格式
   CALL cl_poi_crateCellStyle()
   
   #處理記錄資料
   LET l_ctype = 0
   IF cl_tlm_gl_get_licensenum("TUE") > 0 AND cl_tlm_gl_get_licensenum("G03") = 0 THEN
      LET l_ctype = 1
   END IF
   IF l_ctype = 0 THEN
      LET l_logfile = "/u3/log/lic_c1.log"
   ELSE
      LET l_logfile = "/u3/log/lic_c2.log"
   END IF
   LET l_logfile2 = "/u3/log/lic_c3.log"

   #記錄檔不存在就不處理
   IF NOT os.Path.exists(l_logfile) OR NOT os.Path.exists(l_logfile2) THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = 'azz-13614'
      LET g_errparam.extend = ''
      LET g_errparam.popup = TRUE
      CALL cl_err()        
      LET lb_return = FALSE
   END IF
   #整理匯出
   IF lb_return THEN
      #依據匯出類型整理
      CALL l_rdata.clear()
      CASE p_mode
         WHEN "1"
            #以小時為單位抓取峰值呈現，日期限定前1天
            LET l_date = TODAY -1 USING "yyyy-mm-dd" 
            #處理小時代碼
            FOR l_i = 1 TO 24 
               IF l_i < 10 THEN
                  LET l_hour[l_i] = "0",  l_i  USING "<"
               ELSE
                  LET l_hour[l_i] = l_i  USING "<<"
               END IF
            END FOR
            #以小時為單位搜尋並抓出峰值
            FOR l_i = 1 TO 24 
               LET l_str = l_date, "-", l_hour[l_i]
               #組合搜尋字串
               LET l_cmd = "grep '", l_str, "' ", l_logfile
               CALL l_channel.openPipe(l_cmd, "r")   #執行指令
               CALL l_nums.clear()
               WHILE TRUE
                  LET l_str = l_channel.readLine()
                  IF l_channel.isEof() THEN 
                     EXIT WHILE 
                  END IF
                  LET l_ac = l_nums.getLength() + 1 
                  LET l_nums[l_ac].lc_ck = l_str.substring(1, l_str.getIndexOf('Users :',1) - 1)
                  LET l_tmp = l_str.substring(l_str.getIndexOf('Users :',1) + 8, l_str.getLength())
                  IF l_tmp.getIndexOf('/',1) > 0 THEN
                     LET l_nums[l_ac].lc_used = l_tmp.substring(1, l_tmp.getIndexOf('/',1) - 1)
                  ELSE
                     LET l_nums[l_ac].lc_used = l_tmp
                  END IF
               END WHILE
               #抓出峰值
               LET l_max = 0
               FOR l_j = 1 TO l_nums.getLength()
                  IF l_nums[l_j].lc_used > l_max THEN
                     LET l_max = l_nums[l_j].lc_used
                     LET l_tmp = l_nums[l_j].lc_ck
                  ENd IF
               END FOR
               #存入資料
               IF l_nums.getLength() > 0 THEN
                  LET l_ac = l_rdata.getLength() + 1 
                  LET l_rdata[l_ac].lc_date = l_date
                  LET l_rdata[l_ac].lc_time = l_hour[l_i]
                  LET l_rdata[l_ac].lc_used = l_max
                  LET l_rdata[l_ac].lc_ck = l_tmp
               END IF            
            END FOR
            #抓取在線資料
            FOR l_i = 1 TO l_rdata.getLength()
               LET l_str = l_rdata[l_i].lc_ck
               #組合搜尋字串
               LET l_cmd = "grep '", l_str, "' ", l_logfile2," | awk '{print $4}' | awk -F'/' '{print $1}' "            
               CALL l_channel.openPipe(l_cmd, "r")   #執行指令
               LET l_str = l_channel.readLine()
               IF cl_null(l_str) THEN
                  LET l_rdata[l_i].lc_unlogin = 0
               ELSE
                  LET l_rdata[l_i].lc_unlogin = l_str
               END IF 
            END FOR
         WHEN "2"
            #以天為單位抓取峰值呈現，日期限定前7天
            FOR l_d = 7 TO 1 STEP -1
               LET l_date = TODAY - l_d USING "yyyy-mm-dd" 
               #以天為單位搜尋並抓出峰值
               LET l_str = l_date
               #組合搜尋字串
               LET l_cmd = "grep '", l_str, "' ", l_logfile
               CALL l_channel.openPipe(l_cmd, "r")   #執行指令
               CALL l_nums.clear()
               WHILE TRUE
                  LET l_str = l_channel.readLine()
                  IF l_channel.isEof() THEN 
                     EXIT WHILE 
                  END IF
                  LET l_ac = l_nums.getLength() + 1 
                  LET l_nums[l_ac].lc_ck = l_str.substring(1, l_str.getIndexOf('Users :',1) - 1)
                  LET l_tmp = l_str.substring(l_str.getIndexOf('Users :',1) + 8, l_str.getLength())
                  IF l_tmp.getIndexOf('/',1) > 0 THEN
                     LET l_nums[l_ac].lc_used = l_tmp.substring(1, l_tmp.getIndexOf('/',1) - 1)
                  ELSE
                     LET l_nums[l_ac].lc_used = l_tmp
                  END IF
               END WHILE
               #抓出峰值
               LET l_max = 0
               FOR l_j = 1 TO l_nums.getLength()
                  IF l_nums[l_j].lc_used > l_max THEN
                     LET l_max = l_nums[l_j].lc_used
                     LET l_tmp = l_nums[l_j].lc_ck
                  ENd IF
               END FOR
               #存入資料
               IF l_nums.getLength() > 0 THEN
                  LET l_ac = l_rdata.getLength() + 1 
                  LET l_rdata[l_ac].lc_date = l_date
                  #LET l_rdata[l_ac].lc_time = l_hour[l_i]
                  LET l_rdata[l_ac].lc_used = l_max
                  LET l_rdata[l_ac].lc_ck = l_tmp
               END IF  
               #抓取在線資料
               FOR l_i = 1 TO l_rdata.getLength()
                  LET l_str = l_rdata[l_i].lc_ck
                  #組合搜尋字串
                  LET l_cmd = "grep '", l_str, "' ", l_logfile2," | awk '{print $4}' | awk -F'/' '{print $1}' "            
                  CALL l_channel.openPipe(l_cmd, "r")   #執行指令
                  LET l_str = l_channel.readLine()
                  IF cl_null(l_str) THEN
                     LET l_rdata[l_i].lc_unlogin = 0
                  ELSE
                     LET l_rdata[l_i].lc_unlogin = l_str
                  END IF 
               END FOR
            END FOR
         WHEN "3"
            #以天為單位抓取峰值呈現，日期限定前30天
            FOR l_d = 30 TO 1 STEP -1
               LET l_date = TODAY - l_d USING "yyyy-mm-dd" 
               #以天為單位搜尋並抓出峰值
               LET l_str = l_date
               #組合搜尋字串
               LET l_cmd = "grep '", l_str, "' ", l_logfile
               CALL l_channel.openPipe(l_cmd, "r")   #執行指令
               CALL l_nums.clear()
               WHILE TRUE
                  LET l_str = l_channel.readLine()
                  IF l_channel.isEof() THEN 
                     EXIT WHILE 
                  END IF
                  LET l_ac = l_nums.getLength() + 1 
                  LET l_nums[l_ac].lc_ck = l_str.substring(1, l_str.getIndexOf('Users :',1) - 1)
                  LET l_tmp = l_str.substring(l_str.getIndexOf('Users :',1) + 8, l_str.getLength())
                  IF l_tmp.getIndexOf('/',1) > 0 THEN
                     LET l_nums[l_ac].lc_used = l_tmp.substring(1, l_tmp.getIndexOf('/',1) - 1)
                  ELSE
                     LET l_nums[l_ac].lc_used = l_tmp
                  END IF
               END WHILE
               #抓出峰值
               LET l_max = 0
               FOR l_j = 1 TO l_nums.getLength()
                  IF l_nums[l_j].lc_used > l_max THEN
                     LET l_max = l_nums[l_j].lc_used
                     LET l_tmp = l_nums[l_j].lc_ck
                  ENd IF
               END FOR
               #存入資料
               IF l_nums.getLength() > 0 THEN
                  LET l_ac = l_rdata.getLength() + 1 
                  LET l_rdata[l_ac].lc_date = l_date
                  #LET l_rdata[l_ac].lc_time = l_hour[l_i]
                  LET l_rdata[l_ac].lc_used = l_max
                  LET l_rdata[l_ac].lc_ck = l_tmp
               END IF  
               #抓取在線資料
               FOR l_i = 1 TO l_rdata.getLength()
                  LET l_str = l_rdata[l_i].lc_ck
                  #組合搜尋字串
                  LET l_cmd = "grep '", l_str, "' ", l_logfile2," | awk '{print $4}' | awk -F'/' '{print $1}' "            
                  CALL l_channel.openPipe(l_cmd, "r")   #執行指令
                  LET l_str = l_channel.readLine()
                  IF cl_null(l_str) THEN
                     LET l_rdata[l_i].lc_unlogin = 0
                  ELSE
                     LET l_rdata[l_i].lc_unlogin = l_str
                  END IF 
               END FOR
            END FOR
      END CASE
      
      #匯入 EXCLE 中
      LET l_ac = 1 #excel 第一列為名稱
      FOR l_i = 1 TO l_rdata.getLength()
         LET l_int = 0
         LET row = sheet.createRow(l_ac)
         
         #日期
         LET cell = row.createCell(l_int)
         LET l_str = l_rdata[l_i].lc_date
         CALL cell.setCellValue(l_str)
         LET l_int = l_int + 1
         
         #時間
         LET cell = row.createCell(l_int)
         LET l_str = l_rdata[l_i].lc_time
         CALL cell.setCellValue(l_str)
         LET l_int = l_int + 1
         
         #使用數
         LET cell = row.createCell(l_int)
         LET l_tmpi = l_rdata[l_i].lc_used
         CALL cell.setCellValue(l_tmpi)
         LET l_int = l_int + 1
         
         #含在線數
         LET cell = row.createCell(l_int)
         LET l_tmpi = l_rdata[l_i].lc_used + l_rdata[l_i].lc_unlogin
         CALL cell.setCellValue(l_tmpi)
         LET l_int = l_int + 1
         
         #註冊數量
         LET cell = row.createCell(l_int)
         IF g_auth_cnt = 32767 THEN
            LET l_tmpi = 0
         ELSE
           LET l_tmpi = g_auth_cnt
         END IF
         CALL cell.setCellValue(l_tmpi)
         LET l_int = l_int + 1      
         
         LET l_ac = l_ac + 1      
      END FOR
      
      
      #250328-00007#1 add --- end ---
      #250328-00007#1 mark --- start ---
   #   LET l_hour[1] = "00" 
   #   LET l_hour[2] = "01" 
   #   LET l_hour[3] = "02" 
   #   LET l_hour[4] = "03" 
   #   LET l_hour[5] = "04" 
   #   LET l_hour[6] = "05" 
   #   LET l_hour[7] = "06" 
   #   LET l_hour[8] = "07" 
   #   LET l_hour[9] = "08" 
   #   LET l_hour[10] = "09" 
   #   LET l_hour[11] = "10" 
   #   LET l_hour[12] = "11" 
   #   LET l_hour[13] = "12" 
   #   LET l_hour[14] = "13" 
   #   LET l_hour[15] = "14" 
   #   LET l_hour[16] = "15" 
   #   LET l_hour[17] = "16" 
   #   LET l_hour[18] = "17" 
   #   LET l_hour[19] = "18" 
   #   LET l_hour[20] = "19" 
   #   LET l_hour[21] = "20" 
   #   LET l_hour[22] = "21" 
   #   LET l_hour[23] = "22" 
   #   LET l_hour[24] = "23" 
   #   LET l_channel = base.Channel.create()
   #   LET l_usagelog = "/u3/log/lic.log"
   #   DISPLAY l_usagelog
   #   LET l_current = CURRENT SECOND TO FRACTION(3)
   #   CASE p_mode
   #      WHEN "1"
   #         LET l_filepath = FGL_GETENV("FBIN"),"/42m/azzq080_usage_day.xlsx" 
   #      WHEN "2"
   #         LET l_filepath = FGL_GETENV("FBIN"),"/42m/azzq080_usage_week.xlsx" 
   #      WHEN "3"
   #         LET l_filepath = FGL_GETENV("FBIN"),"/42m/azzq080_usage_month.xlsx" 
   #   END CASE
   #   LET l_filename = "license_usage_", l_current,".xlsx"
   #   LET l_cmd = "cp -rfp ", l_filepath, " ", FGL_GETENV("TEMPDIR"), "/", l_filename
   #   RUN l_cmd
   #   LET l_filepath = FGL_GETENV("TEMPDIR"),"/", l_filename
   #   
   #   DISPLAY "[WithStream]"
   #   LET sxWorkBook = cl_poi_open_sxssf(l_filepath)
   #   
   #   IF sxWorkBook IS NULL THEN
   #      INITIALIZE g_errparam TO NULL
   #      LET g_errparam.code = '-461'
   #      LET g_errparam.extend = ''
   #      LET g_errparam.popup = TRUE
   #      CALL cl_err()    
   #      LET lb_return = FALSE
   #   END IF
   #   
   #   LET sheet = sxWorkBook.getSheetAt(0)
   #   
   #   #找不到Sheet彈窗警告並離開
   #   IF sheet IS NULL THEN
   #      INITIALIZE g_errparam TO NULL
   #      LET g_errparam.code = 'azz-01169'
   #      LET g_errparam.extend = ''
   #      LET g_errparam.popup = TRUE
   #      CALL cl_err()        
   #      LET lb_return = FALSE
   #   END IF
   #   
   #   #移除工作表非空行
   #   CALL cl_poi_removeAllRows(0)
   #   
   #   #建立樣式與資料格式
   #   CALL cl_poi_crateCellStyle()
   #   
   #   CASE p_mode
   #      WHEN "1"
   #         #計算用量鋒值(FOR day)
   #         LET l_circle_max = 24
   #         LET l_date = p_date
   #         LET l_date = l_date.subString(12,13)
   #         LET l_int = l_date
   #         LET l_circle_max = l_circle_max + l_int + 1
   #         LET l_date = p_date - (1 UNITS DAY)
   #         LET l_date = l_date.subString(1,10)
   #      WHEN "2"
   #         #計算用量鋒值(FOR week)
   #         LET l_circle_max = 7
   #      WHEN "3"
   #         #計算用量鋒值(FOR month)
   #         LET l_circle_max = 30
   #   END CASE
   # 
   #   LET l_max = 0
   #   FOR l_circle_row = 1 to l_circle_max 
   #      CASE p_mode
   #         WHEN "1"
   #            #計算用量鋒值(FOR day)
   #            IF l_circle_row <= 24 THEN
   #               LET l_cmd = "cat ", l_usagelog, " | grep '", l_date, " ", l_hour[l_circle_row], "' | awk '{print$6}' | awk -F '/' '{print$1}'| awk 'BEGIN {max = 0} {if ($1+0 > max+0) max=$1} END {print max}'"
   #               LET l_date_a[l_circle_row] = l_date, " ",l_hour[l_circle_row], ":00:"
   #            ELSE
   #               LET l_date = p_date
   #               LET l_date = l_date.subString(1,10)
   #               LET l_cmd = "cat ", l_usagelog, " | grep '", l_date, " ", l_hour[l_circle_row-24], "' | awk '{print$6}' | awk -F '/' '{print$1}'| awk 'BEGIN {max = 0} {if ($1+0 > max+0) max=$1} END {print max}'"
   #               LET l_date_a[l_circle_row] = l_date, " ",l_hour[l_circle_row-24], ":00:"
   #            END IF
   #         WHEN "2"
   #            LET l_date = p_date - ((8 - l_circle_row) UNITS DAY)
   #            LET l_date = l_date.subString(1,10)
   #            LET l_date_a[l_circle_row] = l_date, " 00:00:"
   #            #計算用量鋒值(FOR week)
   #            LET l_cmd = "cat ", l_usagelog, " | grep '", l_date, "' | awk '{print$6}' | awk -F '/' '{print$1}'| awk 'BEGIN {max = 0} {if ($1+0 > max+0) max=$1} END {print max}'"
   #         WHEN "3"
   #            LET l_date = p_date - ((31 - l_circle_row) UNITS DAY)
   #            LET l_date = l_date.subString(1,10)
   #            LET l_date_a[l_circle_row] = l_date, " 00:00:"
   #            #計算用量鋒值(FOR month)
   #            LET l_cmd = "cat ", l_usagelog, " | grep '", l_date, "' | awk '{print$6}' | awk -F '/' '{print$1}'| awk 'BEGIN {max = 0} {if ($1+0 > max+0) max=$1} END {print max}'"
   #      END CASE
   #      CALL l_channel.openPipe(l_cmd,"r")
   #      WHILE TRUE
   #         LET l_msg = l_channel.readLine()
   #         IF l_channel.isEof() THEN
   #            EXIT WHILE
   #         END IF
   #         LET l_value[l_circle_row] = l_msg
   #         IF l_value[l_circle_row] > l_max THEN
   #            LET l_max = l_value[l_circle_row]
   #         END IF
   #      END WHILE
   #   END FOR
   ##   CASE p_mode
   ##      WHEN "1"
   ##         #將資料填入試算表樣板 (for day)
   ##         LET l_cmd = "cat ",l_usagelog, " | grep '", l_date,"' | grep ':00:'"
   ##      WHEN "2"
   ##         #將資料填入試算表樣板 (for week)
   ##         LET l_cmd = "cat ",l_usagelog, " | grep '", l_date,"' | grep '00:00:'"
   ##      WHEN "3"
   ##         #將資料填入試算表樣板 (for month)
   ##         LET l_cmd = "cat ",l_usagelog, " | grep '", l_date,"' | grep '00:00:'"
   ##   END CASE
   #   
   #   LET l_circle_row = 1
   #   WHILE TRUE
   #      LET l_cmd = "cat ",l_usagelog, " | grep '", l_date_a[l_circle_row],"'"
   #      CALL l_channel.openPipe(l_cmd,"r")
   #      LET l_msg = l_channel.readLine()
   #      IF l_channel.isEof() THEN
   #         CONTINUE WHILE
   #      END IF
   #
   #      LET row = sheet.createRow(l_circle_row)
   # 
   #      # 日期
   #      LET l_str = l_msg.subString(1,10)
   #      LET cell = row.createCell(0)
   #      CALL cell.setCellValue(l_str)
   #      # 時間
   #      IF p_mode = "1" THEN
   #         LET l_str = l_msg.subString(6,14)
   #         LET l_str = l_str,"00"
   #      END IF
   #      
   #      IF p_mode = "2" THEN
   #         LET l_str = l_msg.substring(6,10)
   #      END IF
   #      
   #      IF p_mode = "3" THEN
   #         LET l_str = l_msg.substring(6,10)
   #      END IF
   #      LET cell = row.createCell(1)
   #      CALL cell.setCellValue(l_str)
   #      # 授權用量
   #      CALL l_msg.getIndexOf("/", 33) RETURNING l_strindex
   #      LET l_str = l_msg.subString(33,l_strindex-1)
   #      LET cell = row.createCell(2)
   #      LET l_int = l_value[l_circle_row]
   #      CALL cell.setCellValue(l_int)
   #      #授權用量最大值
   #      LET cell = row.createCell(26)
   #      IF l_int = l_max THEN
   #         CALL cell.setCellValue(l_int)
   #         LET l_max = 32767
   #      END IF 
   #         
   #      # 註冊授權數量
   #      LET l_str = l_msg.subString(l_strindex+1, l_msg.getLength())
   #      LET cell = row.createCell(3)
   #      LET l_int = l_str
   #      CALL cell.setCellValue(l_int)
   #      
   #      LET l_circle_row = l_circle_row +1
   #      IF l_circle_row > l_circle_max THEN
   #         EXIT WHILE
   #      END IF
   #   END WHILE
      #250328-00007#1 mark --- end ---

   #寫入檔案
   LET outFile = FileOutputStream.create(l_filepath)
   CALL sxWorkBook.write(outFile)
   CALL outFile.close()
   CALL sxWorkBook.dispose()

   CALL l_channel.close()
   
   LET l_url_path = os.Path.join(os.Path.join(FGL_GETENV("FGLASIP"),"out"),l_filename CLIPPED)
   CALL cl_client_open_url(l_url_path) RETURNING l_result
   END IF #250328-00007#1 add
END FUNCTION
#220801-00038#1 add
################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzq080_license_usage_dialog
# Input parameter: None
# Return code....: None
# Date & Author..: 2022/08/16 By 05598
# Modify.........:
################################################################################
PRIVATE FUNCTION azzq080_license_usage_dialog()
   DEFINE lwin_curr ui.Window
   DEFINE lfrm_curr ui.Form
   DEFINE l_date DATETIME YEAR TO SECOND
   DEFINE l_mode STRING
   DEFINE exportmode STRING

   OPEN WINDOW azzq080_license_usage_w WITH FORM cl_ap_formpath("azz","azzq080_s03")
      ATTRIBUTES(STYLE="layoutwin")
   CALL cl_ui_init()
   LET lwin_curr = ui.Window.getCurrent()
   LET lfrm_curr = lwin_curr.getForm()
   INITIALIZE l_mode TO NULL
    
    INPUT BY NAME exportmode ATTRIBUTE(WITHOUT DEFAULTS)
      BEFORE INPUT
         LET exportmode = '3'
         LET l_mode = '3'
      ON CHANGE exportmode
         LET l_mode = exportmode
      ON ACTION usage_export
         LET l_date = cl_get_current()
         CALL azzq080_export_usage(l_date, l_mode)
      ON ACTION close
         EXIT INPUT
      ON ACTION exit
         EXIT INPUT
   END INPUT
      #交談指令共用ACTION
#      &include "common_action.4gl"

   CLOSE WINDOW azzq080_license_usage_w
END FUNCTION

################################################################################
# Descriptions...: 描述说明 CPU 授權 License清單整理
# Memo...........:
# Usage..........: CALL azzq080_license_collect_TUE()
# Input parameter: None
# Return code....: None
# Date & Author..: 2024/01/05 By 20992
# Modify.........:
################################################################################
PRIVATE FUNCTION azzq080_license_collect_TUE()
   DEFINE lc_channel    base.Channel
   DEFINE ls_str        STRING
   DEFINE ls_sql        STRING
   DEFINE li_idx        LIKE type_t.num5
   DEFINE li_i          INTEGER
   DEFINE la_gzou       DYNAMIC ARRAY OF RECORD 
          gzou003       LIKE gzou_t.gzou003         
     END RECORD
   DEFINE la_gzwl       DYNAMIC ARRAY OF RECORD 
          num           LIKE type_t.num5, #項次
          gzwl002       LIKE gzwl_t.gzwl002, #使用者編號
          gzwl005       LIKE gzwl_t.gzwl005, #授權編號
          gzwl006       LIKE gzwl_t.gzwl006, #使用者IP
          gzwl009       LIKE gzwl_t.gzwl009, #員工編號
          gzwl012       LIKE gzwl_t.gzwl012, #企業編號
          gzwl015       LIKE gzwl_t.gzwl015, #AP主機
          gzwl002_desc  LIKE type_t.chr500, #使用者全名 
          gzxa007       LIKE gzxa_t.gzxa007     #使用者預設營運據點         
     END RECORD											
   
   CALL g_gzwl2_d.clear()
   CALL la_gzwl.clear()                     
   CALL la_gzou.clear()
   INITIALIZE g_fglwrt TO NULL

   #先執行一下process紀錄檢測並清除過期的
   CALL cl_process_auth_refresh()          #r.r adzp444
   CALL azzq080_counting_number()          #181207-00062#2  add

   #整理資訊要用到全體gzwl_t, 需先連結dscom, 處理完後才連回原DB
   IF cl_db_setconnect("dscom") THEN END IF
   LET ls_sql = " SELECT ROW_NUMBER() OVER (ORDER BY gzwl002) RN, gzwl002, gzwl005, gzwl006, gzwl009, gzwl012, gzwl015 FROM ",
                " ( SELECT DISTINCT  gzwl002, gzwl005, gzwl006, gzwl009, gzwl012, gzwl015 FROM gzwl_t WHERE TRIM(gzwl006) IS NOT NULL ) A ",
                " ORDER BY RN, gzwl006  "
   PREPARE azzq080_license_collect_TUE_pre FROM ls_sql
   DECLARE azzq080_license_collect_TUE_cs CURSOR FOR azzq080_license_collect_TUE_pre
   LET li_idx = 1
   FOREACH azzq080_license_collect_TUE_cs INTO la_gzwl[li_idx].num,la_gzwl[li_idx].gzwl002,la_gzwl[li_idx].gzwl005,la_gzwl[li_idx].gzwl006,la_gzwl[li_idx].gzwl009,la_gzwl[li_idx].gzwl012,la_gzwl[li_idx].gzwl015
      IF SQLCA.SQLCODE THEN
         INITIALIZE g_errparam TO NULL 
         LET g_errparam.extend = "FOREACH:",SQLERRMESSAGE 
         LET g_errparam.code = SQLCA.SQLCODE 
         LET g_errparam.popup = TRUE 
         CALL cl_err()
         EXIT FOREACH
      END IF

      LET li_idx = li_idx + 1
   END FOREACH
   CALL la_gzwl.deleteElement(la_gzwl.getLength())

   #再結合原有資料庫資訊
   IF cl_db_setconnect(g_dbs_old) THEN END IF
   LET ls_sql = "SELECT ooag011 FROM ooag_t WHERE ooagent = ? AND ooag001 = ?"
   PREPARE azzq080_license_collect_TUE_ooag_pre FROM ls_sql
   LET ls_sql = "SELECT gzxa007 FROM gzxa_t WHERE gzxaent = ? AND gzxa001 = ?"
   PREPARE azzq080_license_collect_TUE_gzxa_pre FROM ls_sql
   FOR li_idx = 1 TO la_gzwl.getLength()
       EXECUTE azzq080_license_collect_TUE_ooag_pre USING la_gzwl[li_idx].gzwl012, la_gzwl[li_idx].gzwl009 INTO la_gzwl[li_idx].gzwl002_desc
       EXECUTE azzq080_license_collect_TUE_gzxa_pre USING la_gzwl[li_idx].gzwl012, la_gzwl[li_idx].gzwl002 INTO la_gzwl[li_idx].gzxa007
   END FOR
   
   #撈取所有DB庫名稱，處理使用者全名空白部分
   LET ls_sql = "SELECT DISTINCT gzou003 FROM gzou_t ",
                " WHERE  gzoustus = 'Y' " 
   PREPARE azzq080_license_collect_TUE_gzou_pre FROM ls_sql
   DECLARE azzq080_license_collect_TUE_gzou_cs CURSOR FOR azzq080_license_collect_TUE_gzou_pre
   LET li_i = 1
   FOREACH azzq080_license_collect_TUE_gzou_cs INTO la_gzou[li_i].gzou003
      LET li_i = li_i + 1
   END FOREACH
   CALL la_gzou.deleteElement(li_i)

   FOR li_i =  1 TO la_gzou.getLength()
       IF la_gzou[li_i].gzou003 != g_dbs_old THEN
          IF NOT cl_db_setconnect(la_gzou[li_i].gzou003) THEN        
             CALL cl_db_connect(la_gzou[li_i].gzou003, FALSE)
          END IF 
          LET ls_sql = "SELECT ooag011 FROM ooag_t WHERE ooagent = ? AND ooag001 = ?"
          PREPARE azzq080_license_collect_TUE_ooag_pre2 FROM ls_sql
          LET ls_sql = "SELECT gzxa007 FROM gzxa_t WHERE gzxaent = ? AND gzxa001 = ?"
            PREPARE azzq080_license_collect_TUE_gzxa_pre2 FROM ls_sql
          FOR li_idx = 1 TO la_gzwl.getLength()
              IF cl_null(la_gzwl[li_idx].gzwl002_desc) THEN 
                 EXECUTE azzq080_license_collect_TUE_ooag_pre2 USING la_gzwl[li_idx].gzwl012, la_gzwl[li_idx].gzwl009 INTO la_gzwl[li_idx].gzwl002_desc
                 EXECUTE azzq080_license_collect_TUE_gzxa_pre2 USING la_gzwl[li_idx].gzwl012, la_gzwl[li_idx].gzwl002 INTO la_gzwl[li_idx].gzxa007
              END IF 
          END FOR  
       END IF 
   END FOR 

   #最後連回原來DB
   IF cl_db_setconnect(g_dbs_old) THEN END IF 
   
   #呈現資訊
   FOR li_idx = 1 TO la_gzwl.getLength()
       LET g_gzwl2_d[li_idx].num = la_gzwl[li_idx].num
       LET g_gzwl2_d[li_idx].gzwl005 = la_gzwl[li_idx].gzwl005
       LET g_gzwl2_d[li_idx].gzwl002 = la_gzwl[li_idx].gzwl002
       LET g_gzwl2_d[li_idx].gzwl002_desc = la_gzwl[li_idx].gzwl002_desc
       LET g_gzwl2_d[li_idx].gzwl006 = la_gzwl[li_idx].gzwl006
       LET g_gzwl2_d[li_idx].gzwl015 = la_gzwl[li_idx].gzwl015
       LET g_gzwl2_d[li_idx].gzxa007 = la_gzwl[li_idx].gzxa007
   END FOR
END FUNCTION

################################################################################
# Descriptions...: 描述说明：抓取指定授權項目下的目前執行清單
# Memo...........:
# Usage..........: CALL azzq080_process_b_fill_licinfo(pc_fdid)
#                  RETURNING 回传参数
# Input parameter: pc_fdid   fdid
# Return code....: None
# Date & Author..: 2024/07/01 By 20992
# Modify.........: 240628-00031#1 add
################################################################################
PRIVATE FUNCTION azzq080_process_b_fill_licinfo(pc_fdid)
  DEFINE pc_fdid      LIKE type_t.chr500
   DEFINE pc_gzwl015   LIKE gzwl_t.gzwl015
   
   DEFINE lr_others    type_license_d
   DEFINE lc_gzwl002   LIKE gzwl_t.gzwl002
   DEFINE ls_ent       STRING        
   DEFINE li_i         LIKE type_t.num5
   DEFINE li_repeat    BOOLEAN
   DEFINE lc_channel   base.Channel
   DEFINE ls_cmd       STRING
   DEFINE ls_str       STRING
   DEFINE ls_sql       STRING
   DEFINE la_gzou      DYNAMIC ARRAY OF RECORD  
          gzou003    LIKE gzou_t.gzou003         
   END RECORD
   DEFINE la_licinfo   DYNAMIC ARRAY OF RECORD  
          fdid         LIKE type_t.chr100,      
          pid          LIKE type_t.chr20,
          host         LIKE type_t.chr50          
   END RECORD
   DEFINE lc_gzwl005    LIKE gzwl_t.gzwl005
   DEFINE ls_info_host  LIKE gzwl_t.gzwl015
   DEFINE ls_gzwl008  LIKE gzwl_t.gzwl008
   DEFINE li_idx        LIKE type_t.num5
   DEFINE la_pids       DYNAMIC ARRAY OF STRING
   DEFINE ls_tok        base.StringTokenizer
        
   LET ls_sql = " SELECT DISTINCT '', gzwl001,gzwl002,gzwl009,gzwl002_desc,gzxa007,gzwl006, ",
                " gzwl010,gzwl012,gzwl003,gzwl003_desc,gzwl004,gzwl015,gzwl008,gzwl007, ",
                " gzwl014,gzwl013 FROM license_info WHERE fdid = '", pc_fdid, "'",
                " AND gzwl008 Is NOT NULL"
                , " AND ", g_wc, " AND ", g_wc_filter, cl_sql_add_filter("gzwl_t")
   LET ls_sql = cl_sql_add_mask(ls_sql)
   
   PREPARE azzq080_process_b_fill_licinfo_pre FROM ls_sql
   DECLARE azzq080_process_b_fill_licinfo_cur CURSOR FOR azzq080_process_b_fill_licinfo_pre

   CALL g_gzwl_d.clear()
   
   LET l_ac = 1
   FOREACH azzq080_process_b_fill_licinfo_cur INTO g_gzwl_d[l_ac].sel,g_gzwl_d[l_ac].gzwl001,g_gzwl_d[l_ac].gzwl002,
               g_gzwl_d[l_ac].gzwl009,g_gzwl_d[l_ac].gzwl002_desc,g_gzwl_d[l_ac].gzxa007,g_gzwl_d[l_ac].gzwl006,
               g_gzwl_d[l_ac].gzwl010,g_gzwl_d[l_ac].gzwl012,g_gzwl_d[l_ac].gzwl003,g_gzwl_d[l_ac].gzwl003_desc, 
               g_gzwl_d[l_ac].gzwl004,g_gzwl_d[l_ac].gzwl015,g_gzwl_d[l_ac].gzwl008,g_gzwl_d[l_ac].gzwl007,
               g_gzwl_d[l_ac].gzwl014,g_gzwl_d[l_ac].gzwl013
      #同一授權process列表全部顯示         
      LET g_gzwl_d[l_ac].sel = "N"
      LET l_ac = l_ac + 1
   END FOREACH
   CALL g_gzwl_d.deleteElement(l_ac)
   
   IF g_gzwl_d.getLength() <= 0 THEN
      #此使用者於其他 主機/區域 操作, 無法取得資料
      LET ls_str = cl_getmsg("azz-13598",g_lang)         #此使用者於其他 主機/區域 操作, 無法取得資料
      LET g_gzwl_d[1].gzwl002_desc = ls_str
      LET g_gzwl_d[1].sel = "N"
   ELSE
      FOR l_ac = 1 TO g_gzwl_d.getLength()
         IF cl_null(g_gzwl_d[l_ac].gzwl002_desc) THEN
            LET ls_str = cl_getmsg("azz-13598",g_lang)         #此使用者於其他 主機/區域 操作, 無法取得資料
            LET g_gzwl_d[l_ac].gzwl002_desc = ls_str
         END IF
      END FOR
   END IF
     
   IF g_t100debug >=6 THEN
      DISPLAY " 左側執行清單"
      DISPLAY "    pc_fdid:",pc_fdid
      DISPLAY " 右側執行清單筆數 g_gzwl_d.getLength():",g_gzwl_d.getLength() 
      FOR li_i = 1 TO g_gzwl_d.getLength()
         DISPLAY " ======================================= "
         DISPLAY "    gzwl001:",g_gzwl_d[li_i].gzwl001
         DISPLAY "    gzwl010:",g_gzwl_d[li_i].gzwl010
         DISPLAY "    gzwl012:",g_gzwl_d[li_i].gzwl012
         DISPLAY "    gzwl003:",g_gzwl_d[li_i].gzwl003
         DISPLAY "    gzwl003_desc:",g_gzwl_d[li_i].gzwl003_desc
         DISPLAY "    gzwl004:",g_gzwl_d[li_i].gzwl004
         DISPLAY "    gzwl015:",g_gzwl_d[li_i].gzwl015
         DISPLAY "    gzwl008:",g_gzwl_d[li_i].gzwl008
         DISPLAY "    gzwl007:",g_gzwl_d[li_i].gzwl007
         DISPLAY "    gzwl014:",g_gzwl_d[li_i].gzwl014
         DISPLAY "    gzwl013:",g_gzwl_d[li_i].gzwl013
         DISPLAY "    gzwl006:",g_gzwl_d[li_i].gzwl006
         DISPLAY "    gzwl002:",g_gzwl_d[li_i].gzwl002
         DISPLAY "    gzwl002_desc:",g_gzwl_d[li_i].gzwl002_desc
         DISPLAY "    gzxa007:",g_gzwl_d[li_i].gzxa007
         DISPLAY "    gzwl009:",g_gzwl_d[li_i].gzwl009
      END FOR
   END IF
   
END FUNCTION

 
{</section>}
 
