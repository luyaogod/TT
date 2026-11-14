{
  異動型態    異動單號       日 期       異動者      異動內容
  ========= ============= ========== ========== ===============================================
  Modify                  20160506   Ernestliou 1.修改欄位型態時同步更新adzi150預設值及顯示格式
  Modify                  20160510   Ernestliou 1.客制環境才顯示實際表格上的欄位將會根據 r.t 上的資料進行異動
  Modify                  20160513   Ernestliou 1.當產出多語言表時呼叫 r.s 用openpipe, 否則用 RUN
  Add                     20160518   Ernestliou 1.新增 Release Button 
  Modify                  20160519   Ernestliou 1.AZZ模組在客戶端不能簽出
  Add                     20160530   Ernestliou 1.新增匯入比較功能
  Modify                  20160614   Ernestliou 1.FK亂敲防呆措施
  Add                     20160615   Ernestliou 1.新增外鍵（FK）的時候新增提醒訊息
  Add                     20160616   Ernestliou 1.計算 Key 總和長度超過DB定義最大值要警告
  Modify                  20160617   Ernestliou 1.表格主名稱長度不可以少於 n 碼
  Modify                  20160705   Ernestliou 1.如果語系相同才更新多語言的內容
  Modify                  20160706   Ernestliou 1.避免到客戶端還要再執行tbl重產才能有 ud 的欄位
  Modify                  20160803   Ernestliou 1.修正客戶端dzlu008為null的問題
  Modify                  20160812   Ernestliou 1.查詢系統權限清單, 過濾只取 gzda_t 中有定義的
  Modify                  20160824   Ernestliou 1.更改取得當地多語言來源, 以作為匯入多語言的參考
  Modify                  20160825   Ernestliou 1.權限賦予在客戶端可以隨時執行修改, 但是在產中依然要簽出
  Modify                  20160826   Ernestliou 1.控管已出貨的Index或者為PK的Index不能刪除及異動
  Fix                     20160901   Ernestliou 1.修正表格複製時多語言轉換失準的問題
  Fix                     20160914   Ernestliou 1.當為客戶環境時, 不顯示"已出貨"欄位
  Modify                  20161024   Ernestliou 1.提供類似CONSTRUCT查詢的功能
  Modify                  20161102   Ernestliou 1.執行異動或是匯入完成時要進行 Grant 動作
  Modify                  20161209   Ernestliou 1.修正寫死語系
                                                2.修正抓取表格型態來源
  Modify                  20161220   Ernestliou 1.針對客戶自行新增的欄位, 一律不處理 drop 程序
  Modify                  20161229   Ernestliou 1.Command 模式顯示 version
  Modify                  20170208   Ernestliou 1.單表匯入功能加入產生 tbl 檔
  Modify                  20170214   Ernestliou 1.輸入欄位的驗證（輸入ent會過）
                                                2.勾選 PK 時屬性必須先輸入
                                                3.表格重建時對應 Trigger 的異動
                                                4.產出 tbl 檔改為判斷 DGENV來做決定是否產出客戶家的語言資料(除了zh_TW,zh_CN以外)
  Fix                     20170310   Ernestliou 1.修正匯入表格取消的時候會有錯誤訊息
  Modify                  20170315   Ernestliou 1.修正 Trigger 重建時to_char失敗的問題
  Modify                  20170317   Ernestliou 1.Trigger Scripts 改為利用 SPOOL 匯出, 在 Table Rebuild完成後呼叫 sqlplus 執行
  Modify                  20170405   Ernestliou 1.產生 dsgrant.sql 到 /u1/t10tst/apse/std/script/ OR /u1/t10tst/apse/script/
                                                2.取消行業別匯入比對
  Modify                  20170413   Ernestliou 1.ADZ模組只能選擇Z開頭的欄位屬性; 非ADZ模組不能選擇Z開頭的欄位屬性
  Modify                  20170414   Ernestliou 1.異動欄位時先設定 default 為 null
  Modify                  20170417   Ernestliou 1.topstd不可以匯出/入table
  modify                  20170419   Ernestliou 1.匯出 gztd_t 時, 只匯出該表格有使用的部份
                                                2.匯入時, 先匯入 gztd_t 再匯入其他表格
  Modify                  20170425   Ernestliou 1.Mark insert dzebl_t warning message
                                                2.Mark grant to APS warning message
                                                3.僅能讓使用者挑選azzi070內[功能類別]屬於"ERP"的模組
                                                4.啟動時,先抓有簽出的清單,若沒有簽出則以第一個模組顯示
                                                5.拖拉共用欄位時, ADZ 模組的屬性也要跟著變換為Z開頭
                                                6.[執行異動]時先彈窗 : 是否要進行表格的異動
  Modify                  20170428   Ernestliou 1.Fix for APS , modify version only
  Modify                  20170523   Ernestliou 1.新增彙整訊息功能及視窗
                                                2.將儲存時的通知訊息統一彙整到彙整訊息窗中
                                                3.提昇儲存時的效能, 不做資料庫即時比對
                                                4.當 DGENV 為 c 時, 不刪除任何客制欄位
  Modify                  20170706   Ernestliou 1.[表格重建]移到 "DBA專用"中, 但 DGENV="s" 時, toolbar 上面仍然顯示出來
                                                2.[執行異動]最終結果為'失敗'的話, 若透過[表格重建]就可以解決的話, 則提醒使用者自行處理  
  Modify                  20170714   Ernestliou 1.表格批次異動改呼叫 adzp090
  Modify                  20170801   Ernestliou 1.修正權限設定的授予權限角色沒有隨搜尋更新顯示的問題
  Modify                  20170907   Ernestliou 1.版次比對時Debug模式則可以看所有的歷程
  Modify                  20170908   Ernestliou 1.加強產出 tbl/tables.xml 檔失敗的訊息提示
  Modify                  20170911   Ernestliou 1.建表時清空暫存內容
  Modify                  20170912   02160      1. r.t 匯入前判斷是否有比匯入檔較多的欄位, 有則警告
                                                2. 索引匯入時判斷有無已存在 index 欄位, 有則匯入方設失效
                                                3. 產生 schema 因為權限失敗時將訊息改為背景顯示, 但與設計流程有關的部份要顯示出來, 如 tbl 檔
                                                4. 表格異動歷程比對 debug 模式
                                                5. 非設計資料執行刪除訊息的觸發點修正 
                                                6. About 新增顯示 OS Process ID
  Modify                  20171120   02160      1. 新增欄位挑選式匯入功能
                                                2. Analyze table 之前, 先解統計 Lock
                                                3. tbl 的 req 欄位改抓 dzep005  
                                                4. dzep005若非為 "Y", 則將 dzeb005 的值設進來
                                                5. r.t 在客戶端要可以不受 alm 控制簽入表格
  Modify                  20180129   02160      1. 匯入時判斷, 若匯出檔的 ERPVER 和目標區的不一樣則報錯
  Modify                  20180213   02160      1. 針對 GST 3.0 dynamic array Compile 失敗的問題進行修正
  Modify                  20180302   02160      1. 刪除標準欄位時給予警告, 並將刪除的欄位紀錄在表格(dzhj_t)中
                                                2. 從 r.t 單表匯入時, 提示 user 有某些標準欄位已經被刪除, 不會進行匯入,若要匯入標準欄位, 請用挑選式匯入功能
                                                3. 挑選式匯入功能可以匯入已經刪除的標準欄位, 並在匯入時給予提示將匯入已經刪除的欄位, 並將已經刪除的紀錄移除
                                                4. 提供欄位序號修正的功能以修正設計資料的欄位序號和實體表格不一致的情況
  Modify                  20180307   02160      1. 增加Patch Mode, 以讓 adzi140_patch 順利執行
  Modify                  20180313   02160      1. 修正自建索引，但改了名稱為xxxx_ pk*然後執行異動失敗，想要再刪除r.t這筆資料也因為pk規則卡控被擋下的問題
  Modify                  20180314   02160      1. 新增批次簽出
  Modify                  20180410   02160      1. 鼎新內部開發環境支持全新程式支持跨版本匯入, 包含表格
  Modify                  20180412   02160      1. 表格欄位修改說明(r.t, adzi150)後要提醒 : 請問是否要檢視 xxxx_t 多語言的影響範圍?
  Modify                  20180420   02160      1. 讓'全新開發程式'可以在v1.1與v3.0雙向匯出入
                                                2. 可以啟動 ctrl+g 以跑其他程式
  Modify                  20180514   02160      1. 刪除表格時不要關注已經不存在於 azzi085 上的資料庫
  Modify                  20180522   02160      1. 匯入時詢問 : 匯入表格XXX會直接影響該表格的設計資料及實體資料表結構確定要執行嗎
  Modify                  20180903   02160      1. r.t在存檔時要檢查並提醒, 但不阻擋存檔, 這樣才比較有彈性 : ARP,AWS,AZZ,ADZ模組之外,其他的表都需要ent.
                                                2. 分鏡匯入功能也要改
                                                3. 某些alter table的情境是不合法的, 但需要提供人工排除問題的做法 : 只需要提供給客製環境
                                                4. 修正授予權限角色ComboBox無即時更新的問題
                                                5. 修正比對Constratint時惕除系統Constraint的問題
  Modify                  20181218   02160      1.修正抓取比對客制索引的來源為dzec007
  Modify                  20181224   02160      1.修正比對欄位型態或長度不同時所給予的背景訊息
  Modify                  20190108   02160      1.修正ERPVER 1.0 及 3.0 的卡控規則
  Modify                  20190219   02160      1.存檔時判斷是否完成相關欄位跳脫時的檢測
  Modify                  20190304   02160      1.只可以增加(或修改,刪除)"流水號"類的行業專屬欄位,並依據行業欄位命名規則.
                                                2.但不可以將這類專屬欄位設定為PK,以免影響使用同一張表的標準程式的正常運行.
                                                3.不可以對標準欄位做任何異動.
  Modify                  20190415   02160      1.r.t 加行業搜尋條件 adz-01582
  Modify                  20190424   02160      1.支援在建表時同步建立實體紀錄 Trigger 及 adzi190 上的 Trigger Scripts
  Modify                  20190513   02160      1.r.t 在簽出表格前先判斷該表是否存在於 gzdc_t 中, 若有則開啟詢問視窗詢問 :
                                                  "adz-xxxxx 請注意 : 此張表格為系統重要表格, 執行異動及過單後需要做全系統 Rebuild, 請問是否繼續簽出程序 ?"
  Modify                  20190521   02160      1.特定模式才顯示 dzlm_t 的資訊
  Modify                  20190619   02160      1.移除 Notice Trigger 的處理段落
  Modify                  20190819   02160      1.當PK設定超過 10 個欄位的時候就加以控卡
  Modify                  20190912   02160      1.修正索引比對使用 ds 做比較
  Modify                  20191004   02160      1.修正預設值存檔時若為Varchar2型態,必需有前後單引號
  Modify                  20191205   02160      1.在客戶端, 無論是r.t或是後台新增的標準欄位, 一律不去刪除, 只做新增與修改
                                                2.取消參數加 debug 模式
  ------------------------------------------------------------------------------
  Modify                  20191209   02160      1.修改支援 MSSQLServer(1.1.071)   
  Modify                  20200330   02160      1.修改支援 PosgreSQL(1.1.072)   
  Modify                  20210218   02160      1.修正關於勾選欄位為 pk 時出現預設欄位更新值若按下取消會造成後續更新為 null 的問題
  Modify                  20210222   02160      1.修正當下欄位按下Enter(ACCEPT)時需要檢核 :
                                                  a.判斷當下欄位輸入欄位名稱是否符合命名原則
                                                  b.判斷當下欄位輸入欄位屬性是否存在  
  Modify                  20210607   02160      1.修正T100表格複製問題(簡繁以外其他語系不會複製)
  Modify                  20211112   02160      1.當執行在客戶環境時進行權限控管
                                                2.修正輸入預設值存檔時預設值會消失的問題
  Modify                  20211122   02160      1.修改欄位轉為PK時內容更新預設值
  Modify                  20220617   02160      1.修正 Footing data 無法在新增客製 pk 後調整的問題
                                                2.修正刪除單一欄位後,若承接欄位為ud欄位,則會無法儲存的問題
  Modify                  20221017   02160      1.當自定義欄位型態是 cdfUserDefineVarchar 時, 則該欄位大小寫清空
  Modify                  20220222   02160      1.xml 匯出加入 XR_TYPE    
  Modify                  20230606   02160      1.修改支援 MogDB (1.1.082)
  Modify                  20230808   02160      1.修改支援 DM (1.1.083)
  Modify                  20240320   02160      1.呼叫 adzp156 之前先 r.s ds inaj_t
  Modify                  20240416   02160      1.若超過10個PK後就不控管了,只控管與資料庫相關的最大值
                                                2.新增控卡行業環境下不可以對標準表格新增 PK 欄位(adz-01818)
  Modify                  20240521   02160      1.修改支援 KINGBASE (1.1.088)
  Modify                  20240719   02160      1.客戶端純客制表不管制 PK 數量
  Modify                  20250113   02160      1.修正在 PostgreSQL 環境下, 當型態從 int4 轉成 decimal (int4 -> decimal(20,6))時會失敗的問題
}

&include "../4gl/sadzi140_mcr.inc"

IMPORT os
IMPORT util
IMPORT security

SCHEMA ds

&include "../4gl/sadzp000_type.inc"
&include "../4gl/sadzp000_cnst.inc"

&include "../4gl/sadzp200_cnst.inc"
&include "../4gl/sadzp200_type.inc"

&include "../4gl/sadzi140_cnst.inc"

CONSTANT cs_erp_path      STRING = "ERP"
CONSTANT cs_program_title STRING = "adzi140-Table Designer"
CONSTANT cs_version       STRING = "1.1.093"
CONSTANT cs_modify_date   STRING = "2025.01.13"

CONSTANT cs_data_modified STRING = "*"
CONSTANT cs_data_added    STRING = "+"
CONSTANT cs_data_delete   STRING = "-"
CONSTANT cs_data_drag     STRING = ">"
CONSTANT cs_table_tag     STRING = "{TABLE}"

GLOBALS "../../cfg/top_global.inc"

#20230606 for MogDB begin
GLOBALS
  DEFINE g_default_db_pass STRING
  DEFINE g_db_branch       STRING #POLARDB/MOGDB
END GLOBALS
#20230606 for MogDB end

################################################################################
#資料表 TYPE 宣告

&include "../4gl/sadzi140_type.inc"

PRIVATE TYPE T_DZEL_T RECORD
               recordtype LIKE type_t.chr1,
               dzel001    LIKE dzel_t.dzel001,
               dzel002    LIKE dzel_t.dzel002
             END RECORD

PRIVATE TYPE T_DZEK_T_R RECORD
               dzek001 LIKE dzek_t.dzek001,
               dzek002 LIKE dzek_t.dzek002,
               dzek003 LIKE dzek_t.dzek003,
               dzek004 LIKE dzek_t.dzek004,
               gztd002 LIKE gztd_t.gztd002,
               gztd003 LIKE gztd_t.gztd003,
               gztd008 LIKE gztd_t.gztd008,
               dzek005 LIKE dzek_t.dzek005,
               dzek006 LIKE dzek_t.dzek006,
               dzek007 LIKE dzek_t.dzek007,
               dzek008 LIKE dzek_t.dzek008,
               dzek009 LIKE dzek_t.dzek009,
               dzek010 LIKE dzek_t.dzek010,
               DEFAULT_VALUE VARCHAR(100),
               dzej004 LIKE dzej_t.dzej004 #20170425
             END RECORD

PRIVATE TYPE T_COL_PROPERTY RECORD
               c_property      VARCHAR(50),
               c_desc          VARCHAR(100),
               c_widget_width  VARCHAR(30)
             END RECORD

PRIVATE TYPE T_STEP RECORD
               NEXT    BOOLEAN,
               PREV    BOOLEAN,
               CONFIRM BOOLEAN,
               CANCEL  BOOLEAN
             END RECORD

PUBLIC TYPE T_DBOBJ_DEFINE RECORD
               p_version             STRING,
               p_exec_type           STRING,
               p_obj_type            STRING,
               p_obj_name            STRING,
               p_subexec_type        STRING,
               p_subobj_type         STRING,
               p_subobj_name         STRING,
               p_data_type           STRING,
               p_data_length_or_list STRING,
               p_notnull             STRING,
               p_comment             STRING,
               p_req_no              STRING
             END RECORD

PUBLIC TYPE T_FUNCTION_TABLE_TYPE_LIST RECORD
              p_module_name      VARCHAR(15),
              p_front_word       VARCHAR(15),
              p_table_name       VARCHAR(15),
              p_pre_table_name   VARCHAR(15),
              p_aux_type         VARCHAR(15),
              p_table_type       VARCHAR(50),
              p_type             VARCHAR(2),
              p_checkbox         VARCHAR(1),
              p_orig_table_type  VARCHAR(50),
              p_table_type_desc  VARCHAR(512)
            END RECORD

PUBLIC TYPE T_COMMON_COLUMN_LIST RECORD
              p_module_name      VARCHAR(15),
              p_front_word       VARCHAR(15),
              p_table_name       VARCHAR(15),
              p_pre_table_name   VARCHAR(15),
              p_aux_type         VARCHAR(15),
              p_table_type       VARCHAR(50),
              p_type             VARCHAR(2),
              p_checkbox         VARCHAR(1),
              p_orig_table_type  VARCHAR(50),
              p_table_type_desc  VARCHAR(512)
            END RECORD

PUBLIC TYPE T_GRANT_ACCEPT_LIST RECORD
              p_grant_schema  VARCHAR(30),
              p_accept_schema VARCHAR(30)
            END RECORD

#20170523 begin
PUBLIC TYPE T_COMMON_COLUMN_DATAS RECORD
              p_column_type      VARCHAR(30),
              p_short_word       VARCHAR(20),
              p_short_desc       VARCHAR(500)
            END RECORD
#20170523 end
            
################################################################################
## Define Combobox related SQLs
DEFINE ms_sql_modules         STRING
DEFINE ms_sql_cust_modules    STRING
DEFINE ms_sql_data_type       STRING
DEFINE ms_sql_table_type      STRING
DEFINE ms_sql_key_type        STRING
DEFINE ms_sql_index_type      STRING
DEFINE ms_sql_mod_names       STRING
DEFINE ms_sql_genero_widgets  STRING
DEFINE ms_sql_industry        STRING
DEFINE ms_sql_main_form_modules STRING
DEFINE ms_sql_erp_schema_list STRING
DEFINE ms_sql_all_schema_list STRING
DEFINE ms_table_name          STRING
DEFINE ms_concat_sign         STRING #20191209
################################################################################

#定義模組變數
DEFINE m_dzea_t DYNAMIC ARRAY OF T_DZEA_T
DEFINE m_dzeb_t DYNAMIC ARRAY OF T_DZEB_T
DEFINE m_dzec_t DYNAMIC ARRAY OF T_DZEC_T
DEFINE m_dzed_t DYNAMIC ARRAY OF T_DZED_T
DEFINE m_dzel_t DYNAMIC ARRAY OF T_DZEL_T
DEFINE m_dzen_t DYNAMIC ARRAY OF T_DZEN_T

DEFINE m_dzea_t_color DYNAMIC ARRAY OF T_DZEA_T_COLOR
DEFINE m_dzeb_t_color DYNAMIC ARRAY OF T_DZEB_T_COLOR
DEFINE m_dzec_t_color DYNAMIC ARRAY OF T_DZEC_T_COLOR

DEFINE m_old_dzeb_t DYNAMIC ARRAY OF T_DZEB_T

DEFINE m_table_synonym_list DYNAMIC ARRAY OF T_TABLE_SYNONYM

DEFINE m_function_table_type_list DYNAMIC ARRAY OF T_FUNCTION_TABLE_TYPE_LIST

DEFINE m_common_fields DYNAMIC ARRAY WITH DIMENSION 2 OF VARCHAR(50)

DEFINE m_multi_lang_fields  DYNAMIC ARRAY OF VARCHAR(500)
DEFINE m_multi_lang_returns DYNAMIC ARRAY OF VARCHAR(500) #ap_ref用陣列

DEFINE mi_dzea_index  INTEGER
DEFINE mi_dzeb_index  INTEGER
DEFINE mi_dzec_index  INTEGER
DEFINE mi_dzed_index  INTEGER
DEFINE mi_dzen_index  INTEGER

DEFINE ms_module                  STRING
DEFINE ms_search                  STRING
DEFINE ms_version                 STRING
DEFINE ms_alm_version             STRING
DEFINE mb_enable_alm              BOOLEAN
DEFINE ms_show_detail_color       STRING
DEFINE ms_table_user_define_code  STRING
DEFINE ms_table_user_custom_code  STRING
DEFINE ms_column_user_define_code STRING
DEFINE ms_column_user_custom_code STRING
DEFINE ms_column_user_alter_code  STRING
DEFINE ms_column_industry_code    STRING --20190304
DEFINE ms_enable_customize        STRING
DEFINE ms_enable_industry         STRING
DEFINE ms_check_table_tail_code   STRING
DEFINE ms_table_tail_code         STRING
DEFINE ms_table_tail_combine_code STRING
DEFINE ms_key_user_alter_leading_code   STRING
DEFINE ms_index_user_alter_leading_code STRING
DEFINE ms_max_table_name_length   STRING

DEFINE ms_synonym_define         STRING
DEFINE ms_drag_and_drop_internal STRING
DEFINE ms_delete_table_control   STRING

#Master DB 及 Master User
DEFINE ms_master_db    STRING
DEFINE ms_master_user  STRING
DEFINE ms_enable_user_define STRING

DEFINE ms_lang   STRING
DEFINE ms_user   STRING
DEFINE ms_dept   STRING
DEFINE ms_DGENV  STRING
DEFINE ms_enterprise STRING
DEFINE ms_message STRING
DEFINE ms_topind  STRING
DEFINE ms_ind_tail_code STRING

DEFINE mb_drop_confirm  BOOLEAN
DEFINE ms_edit_mode     STRING
DEFINE mb_clone_table   BOOLEAN
DEFINE mb_tool_table    BOOLEAN

DEFINE mb_enable_checkout      BOOLEAN
DEFINE mb_enable_checkout_lib  BOOLEAN
DEFINE mb_enable_softscore     BOOLEAN
DEFINE mb_drop_abnormal_column BOOLEAN

DEFINE mb_debug BOOLEAN
DEFINE mb_patch BOOLEAN
DEFINE ms_GUID  STRING
DEFINE mdt_login_time DATETIME YEAR TO SECOND

DEFINE mb_get_issue_table BOOLEAN
DEFINE mb_modifiable BOOLEAN #20160825
DEFINE mb_first_run BOOLEAN #20170425
DEFINE mb_get_checkout_tables BOOLEAN #20170425
DEFINE mb_get_industry_table BOOLEAN #20190415
DEFINE ms_erpver  STRING #20190424
DEFINE ms_cust  STRING #20190424

MAIN
  CALL adzi140_initialize()
  CALL adzi140_initial_form()
  CALL adzi140_start()
  CALL adzi140_finalize(TRUE)
END MAIN

FUNCTION adzi140_initialize()
DEFINE
  ls_param_enable_alm     STRING,
  ls_replace_arg          STRING,
  ls_erpalm               STRING,
  lb_exists               BOOLEAN,
  lb_error                BOOLEAN,
  lb_result               BOOLEAN,
  ls_result               STRING,
  ls_enable_checkout      STRING,
  ls_enable_checkout_lib  STRING,
  ls_enable_softscore     STRING,
  ls_user                 STRING,
  ls_GUID                 STRING,
  ls_arg_val              STRING,
  ls_drop_abnormal_column STRING,
  ls_version              STRING #20161229

  LET mb_drop_confirm = FALSE

  LET mb_debug = FALSE
  LET mb_patch = FALSE
  LET ms_GUID = ""
  LET mb_modifiable = TRUE #20160825
  LET mb_first_run = TRUE #20170425
  LET ms_erpver = NVL(FGL_GETENV("ERPVER_ALM"),FGL_GETENV("ERPVER")) #20190424 #20200518
  #20211112 begin
  &ifndef DEBUG
    LET ms_DGENV = FGL_GETENV("DGENV")
  &else
    LET ms_DGENV = "s"
    CALL FGL_SETENV("DGENV",ms_DGENV)
  &endif
  #20211112 end
  
  &ifndef DEBUG
    #Debug mode 及 Patch Mode
    LET ls_arg_val = UPSHIFT(ARG_VAL(2))
    LET ls_GUID = ARG_VAL(3)
    CASE
      --WHEN ls_arg_val = "DEBUG" #20191205
      --  CALL FGL_SETENV(cs_env_RT_DEBUG_MODE,"Y") #20191205
      WHEN ls_arg_val = "PATCH"
        LET mb_patch = TRUE
        LET ms_GUID = ls_GUID.Trim()
        IF NVL(ms_GUID,cs_null_value) = cs_null_value THEN
          CALL FGL_SETENV(cs_env_RT_DEBUG_MODE,"N")
        ELSE
          CALL FGL_SETENV(cs_env_RT_DEBUG_MODE,"Y")
        END IF
      #20161229 begin  
      WHEN ls_arg_val = "-V"
        CALL adzi140_get_version()
        CALL adzi140_finalize(FALSE)         
      #20161229 end  
    END CASE
    LET mb_debug = IIF(FGL_GETENV(cs_env_RT_DEBUG_MODE)=="Y",TRUE,FALSE)
    
    #依模組進行系統初始化設定(系統設定)
    CALL cl_tool_init()
    #20211112 add begin
    IF (ms_DGENV = cs_dgenv_customize) AND (NOT cl_auth_user_init()) THEN              
      CALL adzi140_finalize(TRUE)                              
    END IF
    #20211112 add end
    CALL cl_db_connect("ds", TRUE)
    LET ms_lang = g_lang
    LET ms_user = g_account --g_user --
    LET ms_dept = g_dept
    LET ms_enterprise = g_enterprise
    #CALL sadzi140_db_connect_db(cs_master_db) RETURNING lb_result
    #IF NOT lb_result THEN CALL adzi140_finalize(FALSE) END IF
    LET g_bgjob = "N"
    LET mdt_login_time = cl_get_current()
  &else
    CALL FGL_GETENV("USER") RETURNING ls_user
    LET ms_lang = cs_default_lang
    LET ms_user = ls_user
    LET ms_dept = cs_dept
    LET ms_enterprise = cs_enterprise
    &ifdef is_pgs
    CALL sadzi140_db_connect_db(cs_test_pgs) RETURNING lb_result
    &else
    &ifdef is_msv
    CALL sadzi140_db_connect_db(cs_test_msv) RETURNING lb_result
    &else
    CALL sadzi140_db_connect_db(cs_test_kdb) RETURNING lb_result
    #CALL sadzi140_db_connect_db(cs_t10dit_db) RETURNING lb_result
    &endif
    &endif
    IF NOT lb_result THEN CALL adzi140_finalize(FALSE) END IF
    CALL FGL_SETENV("RT_DEBUG_MODE","Y")
    LET mb_debug = IIF(FGL_GETENV(cs_env_RT_DEBUG_MODE)=="Y",TRUE,FALSE)
    LET mb_patch = FALSE
    LET ms_GUID = "FEF0742D-F5C0-4A5E-819E-69FF67F7B02E"
    LET mdt_login_time = CURRENT YEAR TO SECOND
    LET ms_erpver = '3.0'
    CALL FGL_SETENV("ERPVER",ms_erpver)
    LET ms_cust   = "local"
    CALL FGL_SETENV("CUST",ms_cust)
    LET g_db_branch = cs_pgs_sub_kdb
  &endif

  CALL sadzi140_util_get_concat_sign() RETURNING ms_concat_sign #20191209

  #檢核資料庫設定的完整性
  DISPLAY cs_information_tag,"Verify the integrity of DB initial data ..."
  CALL sadzi140_util_identify_db_data(FALSE,ms_lang) RETURNING lb_result
  IF NOT lb_result THEN
    CALL adzi140_finalize(FALSE)
  END IF

  CALL sadzi140_util_get_common_fields(m_common_fields) #取得DZEF要用到的共用欄位設定
  CALL sadzi140_util_set_execute_path(os.path.pwd()) #設定執行路徑

  #取得Master DB 及 User 及 使用者定義欄位
  CALL sadzi140_db_get_parameter(cs_param_level_system,cs_param_master_db) RETURNING ms_master_db
  CALL sadzi140_db_get_parameter(cs_param_level_system,cs_param_master_user) RETURNING ms_master_user
  CALL sadzi140_db_get_parameter(cs_param_level_system,cs_param_enable_customize) RETURNING ms_enable_customize
  CALL sadzi140_db_get_parameter(cs_param_level_system,cs_param_enable_industry) RETURNING ms_enable_industry

  CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_enable_user_define) RETURNING ms_enable_user_define
  CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_version) RETURNING ms_version
  CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_show_detail_color) RETURNING ms_show_detail_color
  CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_drag_and_drop_internal) RETURNING ms_drag_and_drop_internal
  CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_delete_table_control) RETURNING ms_delete_table_control
  CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_synonym_define) RETURNING ms_synonym_define
  CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_check_table_tail_code) RETURNING ms_check_table_tail_code
  CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_table_tail_code) RETURNING ms_table_tail_code

  #判斷是否要進行 Columns, Key, Index 等的 Drop 程序
  CALL sadzi140_db_get_parameter(cs_param_level_alter_table,cs_param_drop_abnormal_column) RETURNING ls_drop_abnormal_column
  LET mb_drop_abnormal_column = IIF(NVL(ls_drop_abnormal_column,"N")=="Y",TRUE,FALSE)

  &ifndef DEBUG
    #行業別
    LET ms_topind = NVL(NVL(FGL_GETENV("INDSD"),FGL_GETENV("TOPIND")),cs_default_topind) #20201007
  &else
    #行業別
    LET ms_topind = "sd"
    CALL FGL_SETENV("TOPIND",ms_topind)
  &endif
  LET ms_column_industry_code = IIF(ms_topind==cs_default_topind,NULL,ms_topind) -- 20190304

  IF (ms_enable_customize = "Y") OR (ms_DGENV = cs_dgenv_customize) THEN
    CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_table_user_define_code) RETURNING ms_table_user_define_code
    CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_table_user_custom_code) RETURNING ms_table_user_custom_code
    CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_column_user_define_code) RETURNING ms_column_user_define_code
    CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_column_user_custom_code) RETURNING ms_column_user_custom_code
    CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_column_user_alter_code) RETURNING ms_column_user_alter_code

    CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_key_user_alter_leading_code) RETURNING ms_key_user_alter_leading_code
    CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_index_user_alter_leading_code) RETURNING ms_index_user_alter_leading_code
    #客戶端一律不冠上行業別碼
    LET ms_ind_tail_code = "" 
  ELSE
    #標準區則視TOPIND, 若為sd一律不冠上行業別碼,否則冠上行業別碼
    LET ms_ind_tail_code = IIF(ms_topind==cs_default_topind,"",ms_topind) 
  END IF
  
  #組合表格尾碼
  LET ms_table_tail_combine_code = ms_table_user_custom_code,ms_ind_tail_code,ms_table_tail_code
  
  CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_column_user_define_code) RETURNING ms_column_user_define_code
  CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_max_table_name_length) RETURNING ms_max_table_name_length

  #是否啟動ALM
  LET ms_alm_version = "0"
  CALL sadzi140_db_get_parameter(cs_param_level_alm,cs_param_enable_alm) RETURNING ls_param_enable_alm

  &ifndef DEBUG
    #系統是否允許LIB相關資源被簽出
    CALL cl_get_para("","","A-SYS-0040") RETURNING ls_enable_checkout_lib
    #設計器是否准許分鏡系統串接
    CALL cl_get_para("","","A-SYS-0041") RETURNING ls_enable_softscore
    #是否可簽出
    CALL FGL_GETENV("TOPCHKOUT") RETURNING ls_enable_checkout
    #是否啟動ALM
    CALL FGL_GETENV("TOPALM") RETURNING ls_erpalm
  &else
    LET ls_enable_checkout = "Y"
    LET ls_erpalm = "N"
  &endif

  LET mb_enable_checkout_lib = IIF(NVL(ls_enable_checkout_lib,"N")=="Y",TRUE,FALSE)
  LET mb_enable_softscore    = IIF(NVL(ls_enable_softscore,"N")=="Y",TRUE,FALSE)
  LET mb_enable_checkout     = IIF(NVL(ls_enable_checkout,"N")=="Y",TRUE,FALSE)

  IF (NVL(ls_param_enable_alm,"N") = "Y") OR (NVL(ls_erpalm,"N") = "Y") THEN
    LET mb_enable_alm = TRUE
    #檢查DZLM_T是否存在
    CALL sadzp200_alm_check_alm_table_exist(ms_master_user) RETURNING lb_exists
    IF NOT lb_exists THEN
      LET ls_replace_arg = "|"
      CALL sadzp000_msg_show_error('adz-00191', 'adz-00191', ls_replace_arg, 1)
      CALL adzi140_finalize(FALSE)
    END IF
  ELSE
    LET mb_enable_alm  = FALSE
  END IF

  &ifdef is_msv
  &else
  &ifdef is_pgs
  &else
    IF sadzi140_util_get_db_type() <> cs_db_driver_type_kdb THEN #20240521
      #建立 Stored procedure
      DISPLAY cs_information_tag,"Check and create stored procedure which needed by table designer ..."
      CALL sadzi140_util_create_procedure(ms_master_db,ms_master_user,cs_prc_get_col_default,FALSE) RETURNING lb_error
      IF lb_error THEN
        #失敗則離開
        CALL sadzp000_msg_show_error('adz-00469', 'adz-00469', NULL, 1)
        CALL adzi140_finalize(FALSE)
      END IF
      CALL sadzi140_util_create_procedure(ms_master_db,ms_master_user,cs_prc_get_col_default_by_owner,FALSE) RETURNING lb_error
      IF lb_error THEN
        #失敗則離開
        CALL sadzp000_msg_show_error('adz-00469', 'adz-00469', NULL, 1)
        CALL adzi140_finalize(FALSE)
      END IF
    END IF  
  &endif   
  &endif

END FUNCTION

FUNCTION adzi140_initial_form()
DEFINE
  ls_designer_4ad STRING,
  ls_top_menu    STRING,
  ls_toolbar     STRING,
  ls_erp_path    STRING,
  lo_combobox    ui.ComboBox,
  lo_window      ui.Window,
  lo_form        ui.Form,
  ls_separator   STRING,
  ls_logo_path   STRING,
  ls_sys_info    STRING,
  ls_db_type     STRING #20240521
#Being:20150417 by Hiko
DEFINE 
  ls_cfg_path STRING,
  ls_4st_path STRING,
  ls_img_path STRING
#End:20150417 by Hiko

  LET ls_separator = os.Path.separator()

  CALL FGL_GETENV(cs_erp_path) RETURNING ls_erp_path
  CALL adzi140_initial_combobox_sql()

  #按 Enter 自動跳欄位
  OPTIONS INPUT WRAP
  CLOSE WINDOW SCREEN

  &ifndef DEBUG
    OPEN WINDOW w_adzi140 WITH FORM cl_ap_formpath("ADZ","adzi140")
    CURRENT WINDOW IS w_adzi140

    LET ls_top_menu = ls_erp_path,ls_separator,"adz",ls_separator,"4tm",ls_separator,ms_lang,ls_separator,"adzi140.4tm"
    LET ls_toolbar  = ls_erp_path,ls_separator,"cfg",ls_separator,"4tb",ls_separator,"toolbar_adzi140_",ms_lang,".4tb"
    LET ls_img_path = os.Path.join(os.Path.join(FGL_GETENV("RES"), "img"), "ui")
    LET ls_cfg_path = os.Path.join(FGL_GETENV("ERP"), "cfg")

    LET lo_window = ui.Window.getCurrent()
    LET lo_form = lo_window.getForm()
    DISPLAY cs_information_tag,"Loading top menu : ",ls_top_menu
    CALL lo_form.loadTopMenu(ls_top_menu)
    DISPLAY cs_information_tag,"Loading toolbar : ",ls_toolbar
    CALL lo_form.loadToolBar(ls_toolbar)

    #Being:20150417 by Hiko
    #CALL adzi140_get_system_information(TRUE) RETURNING ls_sys_info #20150417 by Hiko
    CALL cl_ui_wintitle(1) #工具抬頭名稱
    #CALL cl_load_4ad_interface(NULL)
    LET ls_logo_path = os.Path.join(os.Path.join(ls_img_path, "logo"), "dsc_logo.ico")
    DISPLAY cs_information_tag,"Set logo : ",ls_logo_path
    CALL lo_window.setImage(ls_logo_path)
    CALL sadzi140_util_set_form_style(ms_lang)
    #End:20150417 by Hiko
    
    #CALL ui.Interface.loadStyles(ls_erp_path||ls_separator||"cfg"||ls_separator||"4st"||ls_separator||"adzi140") #20150417 by Hiko
    LET ls_designer_4ad = os.Path.join(os.Path.join(os.Path.join(ls_erp_path,"cfg"),"4ad"),"designer.4ad")
    --CALL lo_form.loadActionDefaults(ls_designer_4ad)
    DISPLAY cs_information_tag,"Loading 4ad : ",ls_designer_4ad
    CALL ui.Interface.loadActionDefaults(ls_designer_4ad)
  &else
    OPEN WINDOW w_adzi140 WITH FORM sadzi140_util_get_form_path("adzi140")
    CURRENT WINDOW IS w_adzi140
    LET lo_window = ui.Window.getCurrent()
    LET lo_form   = lo_window.getForm()

    CALL adzi140_get_system_information(TRUE) RETURNING ls_sys_info
    LET ls_top_menu = ls_erp_path,ls_separator,"adz",ls_separator,"4tm",ls_separator,ms_lang,ls_separator,"adzi140.4tm"
    LET ls_toolbar  = ls_erp_path,ls_separator,"cfg",ls_separator,"4tb",ls_separator,"toolbar_adzi140_",ms_lang,".4tb"

    CALL lo_form.loadTopMenu(ls_top_menu)
    CALL lo_form.loadToolBar(ls_toolbar)
    CALL sadzi140_util_set_form_style(ms_lang)
    #LET ls_designer_4ad = os.Path.join(os.Path.join(os.Path.join(ls_erp_path,"cfg"),"4ad"),"designer.4ad")
    #CALL ui.Interface.loadActionDefaults(ls_designer_4ad)
  &endif

  CALL ui.Dialog.setDefaultUnbuffered(TRUE)

  #20160914 begin
  IF (ms_DGENV = cs_dgenv_customize) THEN
    CALL lo_form.setElementHidden("formonly.lbl_dzeb031",TRUE)
    CALL lo_form.setElementHidden("tbi_table_rebuild",TRUE) #20170706
    CALL lo_form.setElementHidden("tm_table_rebuild",TRUE) #20240521
  ELSE
    CALL lo_form.setElementHidden("formonly.lbl_dzeb031",FALSE)
    CALL lo_form.setElementHidden("tbi_table_rebuild",FALSE) #20170706
    CALL lo_form.setElementHidden("tm_table_rebuild",FALSE) #20240521
  END IF
  #20160914 end

  &ifdef is_msv
  CALL lo_form.setElementHidden("page_pageprivilege",TRUE)
  CALL lo_form.setElementHidden("tbi_table_rebuild",TRUE)
  CALL lo_form.setElementHidden("tm_table_rebuild",TRUE)
  &else
  &ifdef is_pgs
  CALL lo_form.setElementHidden("tbi_table_rebuild",TRUE)
  CALL lo_form.setElementHidden("tm_table_rebuild",TRUE)
  &else
  #20240521 begin
  LET ls_db_type = sadzi140_util_get_db_type()
  IF ls_db_type <> cs_db_driver_type_ora THEN 
    CALL lo_form.setElementHidden("tbi_table_rebuild",TRUE)
    CALL lo_form.setElementHidden("tm_table_rebuild",TRUE)
  END IF
  #20240521 end
  &endif
  &endif
  
  DISPLAY cs_information_tag,"Loading combobox data ..."
  LET lo_combobox = ui.ComboBox.forName("formonly.cb_moduleselect")
  CALL sadzi140_util_fill_combobox_by_sql(lo_combobox,ms_sql_main_form_modules)

  #LET lo_combobox = ui.ComboBox.forName("formonly.dzeb006")
  #CALL sadzi140_util_fill_combobox_by_sql(lo_combobox,ms_sql_data_type)

  LET lo_combobox = ui.ComboBox.forName("formonly.lbl_dzea004")
  CALL sadzi140_util_fill_combobox_by_sql(lo_combobox,ms_sql_table_type)

  LET lo_combobox = ui.ComboBox.forName("formonly.lbl_dzed003")
  CALL sadzi140_util_fill_combobox_by_sql(lo_combobox,ms_sql_key_type)

  LET lo_combobox = ui.ComboBox.forName("formonly.lbl_dzec003")
  CALL sadzi140_util_fill_combobox_by_sql(lo_combobox,ms_sql_index_type)

  LET lo_combobox = ui.ComboBox.forName("formonly.lbl_dzea003")
  CALL sadzi140_util_fill_combobox_by_sql(lo_combobox,ms_sql_mod_names)

  #LET lo_combobox = ui.ComboBox.forName("formonly.lbl_dzeb010")
  #CALL sadzi140_util_fill_combobox_by_sql(lo_combobox,ms_sql_genero_widgets)

  #LET lo_combobox = ui.ComboBox.forName("formonly.cbox_tabletype")
  #CALL sadzi140_util_fill_combobox_by_sql(lo_combobox,ms_sql_table_type)

  LET lo_combobox = ui.ComboBox.forName("formonly.lbl_dzen002")
  CALL sadzi140_util_fill_combobox_by_sql(lo_combobox,ms_sql_erp_schema_list)

  LET lo_combobox = ui.ComboBox.forName("formonly.lbl_dzen003")
  CALL sadzi140_util_fill_combobox_by_sql(lo_combobox,ms_sql_all_schema_list)

END FUNCTION

FUNCTION adzi140_start()
DEFINE
  lo_window               ui.Window,
  lo_form                 ui.Form, #20190415
  ls_module               STRING,
  ls_schema_sql           STRING,
  ls_table_name           STRING,
  ls_module_name          STRING,
  lb_indsd                BOOLEAN, #20201007 
  ls_top_module_name      STRING,
  ls_top_env              STRING,
  ls_module_env           STRING,
  ls_search               STRING,
  lb_drop_status          BOOLEAN,
  ls_question             STRING,
  ls_tbl_Name             STRING,
  lo_dnd                  ui.DragDrop,
  drag_source             STRING,
  drag_index              INTEGER,
  drop_index              INTEGER,
  drag_value              STRING,
  ls_group_id             STRING,
  li_index                INTEGER,
  ls_parameter            STRING,
  ls_key_list             STRING,
  ls_sql_script           STRING,
  lo_create_table         T_DZEA_CREATE_TABLE,
  lo_schema_type          DYNAMIC ARRAY OF T_TABLE_SYNONYM,
  lo_function_table_list  DYNAMIC ARRAY OF T_FUNCTION_TABLE_TYPE_LIST,
  ls_version              STRING,
  #li_step                 INTEGER,
  lo_db_connstr           T_DB_CONNSTR,
  li_loop                 INTEGER,
  ls_all_message          STRING,
  ls_message              STRING,
  ls_table_sql_filename   STRING,
  ls_synonym_sql_filename STRING,
  ls_master_user          STRING,
  ls_db_username          STRING,
  ls_curr_version         STRING,
  ls_new_version          STRING,
  ls_table_grant          STRING,
  lo_table_in_db_type     DYNAMIC ARRAY OF T_TABLE_IN_DB_TYPE,
  ls_working              STRING,
  ls_working_path         STRING,
  ls_alm_version          STRING,
  ls_user                 STRING,
  ls_replace_arg          STRING,
  lb_error                BOOLEAN,
  ls_module_path          STRING,
  lb_kill_tbl_files       BOOLEAN,
  ls_current_item         STRING,
  lb_diff                 BOOLEAN,
  lb_table_lock           BOOLEAN,
  lb_table_rebuild        BOOLEAN,
  ls_rtn_code             STRING,
  ls_table_type           STRING,
  ls_get_issue_table      STRING,
  ls_get_industry_table   STRING, #20190415
  ls_os_separator         STRING,
  ls_4gl_path             STRING,
  ls_command              STRING,
  ls_file_name            STRING,
  ls_program_name         STRING,
  ls_pwd_path             STRING,
  ls_source_Table         STRING,
  ls_dest_Table           STRING,
  lb_success              BOOLEAN,
  li_return               INTEGER,
  ls_tdi_name             STRING,
  ls_export_path          STRING,
  ls_export_desc          STRING,
  ls_export_title         STRING,
  ls_import_path          STRING,
  ls_import_desc          STRING,
  ls_import_title         STRING,
  ls_clinet_path_name     STRING,
  ls_separator            STRING,
  lb_exist_foreign_key    BOOLEAN,
  ls_alm_request_no       STRING,
  ls_dgenv                STRING,
  ls_GUID                 STRING,
  ls_update_type          STRING,
  ls_table_desc           STRING,
  ls_alm_sd_version       STRING,
  li_dzlu                 INTEGER,
  ls_request_no           STRING,
  ls_sequence_no          STRING,
  ls_check_out_full_name  STRING,
  lb_commit               BOOLEAN,
  lb_active               BOOLEAN,
  ls_error                STRING,
  ls_table_list           STRING,
  lb_enable_next          BOOLEAN,
  lb_enable_confirm       BOOLEAN,
  li_prev_step            INTEGER,
  li_next_step            INTEGER,
  li_step                 INTEGER,
  li_arr_curr             INTEGER,
  ls_about                STRING,
  ls_module_code          STRING,
  ls_lang                 STRING,
  ls_table_xml            STRING,
  ls_industry_type        STRING,
  ls_dzeb002              STRING,
  ls_orig_path            STRING,
  lb_data_altered         BOOLEAN,
  lb_dzeb006_changed      BOOLEAN,
  lb_dzeb003_changed      BOOLEAN, #20180412
  lb_insert               BOOLEAN,
  lb_clone_import         BOOLEAN,
  ls_new_work_dir         STRING,
  ls_tar_name             STRING,
  ls_dzea003              STRING,
  ls_err_msg              STRING,
  ls_gen_tbl_list         STRING, #20170908
  ls_drop_backdoor_column STRING, #20161220
  ls_backdoor_col_list    STRING, #20161220
  li_backdoor_col_counts  INTEGER,#20161220
  lb_alter_backdoor       BOOLEAN,#20161220
  ls_result               STRING, #160504-00002
  ls_res_code             STRING, #160504-00002
  ls_get_checkout_tables  STRING, #20170425
  li_columnid_rec         INTEGER,#20170706
  ls_column_queue         STRING, #20180302
  ls_columns              STRING, #20180302
  ls_del_list             STRING, #20180302 
  lb_filter_hidden        BOOLEAN, #20190415
  ls_adv_filter_text      STRING,  #20190415
  lo_pick_col_arr         DYNAMIC ARRAY OF T_COLUMN_LIST, #20180302
  lo_left_array           DYNAMIC ARRAY OF T_COLUMN_LIST, #20180302
  lo_right_array          DYNAMIC ARRAY OF T_COLUMN_LIST, #20180302
  lo_check_out_owner      T_CHECK_OUT_OWNER_LIST, #160504-00002
  lo_mark_info            T_MARK_INFORMATION,  #160504-00002 
  lo_new_dir_list         DYNAMIC ARRAY OF T_DIRECTORY_LIST, 
  lo_DZLU_T               DYNAMIC ARRAY OF T_DZLU_T,
  lo_combobox             ui.ComboBox,
  lo_step                 T_STEP,
  lo_REVISION             T_REVISION,
  lo_user_info            T_USER_INFO, -- 使用者資訊物件
  lo_export_parameter     T_DZLM_T_SCR,
  lo_import_parameter     T_DZLM_T_SCR,
  lo_file_dialog          T_FILE_DIALOG,
  lo_dzlm_t               T_DZLM_T,
  lo_dzlm_info            T_DZLM_INFO,
  lo_export_info          T_EXPORT_INFO,
  lo_import_info          T_EXPORT_INFO,
  lo_putfile_para         T_PUT_GET_FILE_PARA,
  lo_getfile_para         T_PUT_GET_FILE_PARA,
  lo_tar_file_info        T_TAR_FILE_INFO,
  lo_table_info           T_PROGRAM_INFO,
  lo_DZAF_T               T_DZAF_T,
  lo_base_table_info      T_BASE_TABLE_INFO,
  ls_msg_no               STRING, #20170523
  lo_collect_messages     DYNAMIC ARRAY OF T_COLLECT_MESSAGE, #20170523
  lo_lang_arr             DYNAMIC ARRAY OF T_LANGUAGE_TYPE,
  lo_directory_list       DYNAMIC ARRAY OF T_DIRECTORY_LIST
DEFINE
  lb_return      BOOLEAN,
  ls_retmsg      STRING,
  lb_result      BOOLEAN,
  lb_gen_tbl     BOOLEAN, #20170908
  ls_gen_list    STRING, #20170908
  lb_isFault     BOOLEAN,
  lb_DropSuccess BOOLEAN

  LET ls_top_module_name = cs_top_module_name
  LET ls_top_env         = FGL_GETENV(ls_top_module_name)
  LET ls_user            = ms_user
  LET ls_master_user     = ms_master_user

  LET ls_separator = os.Path.separator()

  LET mb_clone_table = FALSE
  LET lb_clone_import = FALSE
  LET lb_alter_backdoor = FALSE #20161220

  #取得作業系統的分隔符號
  CALL os.Path.separator() RETURNING ls_os_separator

  DIALOG ATTRIBUTES(UNBUFFERED)

    #Table Lists
    DISPLAY ARRAY m_dzea_t TO sr_TableList.*
      BEFORE ROW
        LET mi_dzea_index = ARR_CURR()
        LET ms_table_name = m_dzea_t[mi_dzea_index].dzea001
        LET ls_table_name = ms_table_name
        LET ls_dzea003    = m_dzea_t[mi_dzea_index].dzea003
        DISPLAY m_dzea_t[mi_dzea_index].dzea001 TO formonly.lbl_table_name
        CALL adzi140_refresh_detail(DIALOG,ls_table_name)
        #設定 Import 功能及建立function table 的功能是否 Enable
        CALL adzi140_set_import_sco_enable(DIALOG)
        CALL adzi140_set_editor_mode(DIALOG,cs_editor_mode_browse)
        CALL adzi140_set_editable(DIALOG,mi_dzea_index)
        CALL adzi140_set_check_out_mode(DIALOG,m_dzea_t[mi_dzea_index].*)
        #填充欄位定義清單
        CALL adzi140_fill_dzel_t(ms_dgenv,m_dzea_t[mi_dzea_index].dzea014) #20190304
    END DISPLAY

    #Column Define( Drag and Drop )
    DISPLAY ARRAY m_dzel_t TO sr_cdfList.*
      ON DRAG_START(lo_dnd)
        LET drag_source = "sr_cdflist"
        LET drag_index = arr_curr()
        LET drag_value = m_dzel_t[drag_index].DZEL001

      ON DRAG_FINISHED(lo_dnd)
        INITIALIZE drag_source TO NULL

      ON DRAG_ENTER(lo_dnd)
        IF drag_source IS NULL THEN
          CALL lo_dnd.setOperation(NULL)
        END IF

      ON DROP(lo_dnd)
        {
        IF drag_source == "sr_cdflist" THEN
          IF ms_drag_and_drop_internal = "Y" THEN
            CALL lo_dnd.dropInternal()
            IF adzi140_get_editor_mode() <> cs_editor_mode_design THEN
              CALL adzi140_set_editor_mode(DIALOG,cs_editor_mode_design)
              NEXT FIELD lbl_dzel001
            END IF
          END IF
        ELSE
        }
          IF adzi140_check_if_check_out_active(mi_dzea_index) THEN
            IF adzi140_get_editor_mode() <> cs_editor_mode_design THEN
              CALL adzi140_set_editor_mode(DIALOG,cs_editor_mode_design)
              NEXT FIELD lbl_dzel001
            END IF
            LET drop_index = lo_dnd.getLocationRow()
            FOR li_index = m_dzeb_t.getLength() TO 1 STEP -1
              IF m_dzeb_t[li_index].RECORDTYPE = cs_data_drag THEN
                IF m_dzeb_t[li_index].GROUPID = ls_group_id THEN
                  CALL DIALOG.deleteRow("sr_tablecolumns", li_index)
                END IF
              END IF
            END FOR
          END IF
        {
        END IF
        }

    END DISPLAY

    #Table Columns( Drag and Drop )
    DISPLAY ARRAY m_dzeb_t TO sr_TableColumns.*
      BEFORE ROW
        LET mi_dzeb_index = ARR_CURR()

      ON DRAG_START(lo_dnd)
        LET drag_source = "sr_tablecolumns"
        LET drag_index  = arr_curr()
        LET drag_value  = m_dzeb_t[drag_index].DZEB002
        LET ls_group_id = m_dzeb_t[drag_index].GROUPID
        FOR li_index = 1 TO m_dzeb_t.getLength()
          IF m_dzeb_t[li_index].RECORDTYPE = cs_data_drag THEN
            IF m_dzeb_t[li_index].GROUPID = ls_group_id THEN
              CALL DIALOG.setSelectionRange("sr_tablecolumns",li_index,li_index,TRUE)
            END IF
          END IF
        END FOR

      ON DRAG_FINISHED(lo_dnd)
        INITIALIZE drag_source TO NULL

      ON DRAG_ENTER(lo_dnd)
        IF drag_source IS NULL THEN
          CALL lo_dnd.setOperation(NULL)
        END IF

      ON DROP(lo_dnd)
        IF drag_source == "sr_tablecolumns" THEN
          IF (ms_drag_and_drop_internal = "Y") AND (adzi140_check_if_check_out_active(mi_dzea_index)) AND
             (ms_dgenv = NVL(m_dzea_t[mi_dzea_index].dzea018,ms_dgenv)) AND (m_dzea_t[mi_dzea_index].dzea017 = "N") THEN
            CALL lo_dnd.dropInternal()
            #有拖拉順序, 則將所有的欄位都標示為已修改
            FOR li_loop = 1 TO m_dzeb_t.getLength()
              IF m_dzeb_t[li_loop].dzeb002 = drag_value THEN
                LET m_dzeb_t[li_loop].recordtype = cs_data_drag
              END IF
              IF (NVL(m_dzeb_t[li_loop].recordtype,cs_null_default) <> cs_data_drag) AND
                 (NVL(m_dzeb_t[li_loop].recordtype,cs_null_default) <> cs_data_added) THEN
                LET m_dzeb_t[li_loop].recordtype = NVL(m_dzeb_t[li_loop].recordtype,cs_data_modified)
              END IF
              LET m_dzeb_t[li_loop].dzebseq = li_loop
            END FOR
            IF adzi140_get_editor_mode() <> cs_editor_mode_design THEN
              CALL adzi140_set_editor_mode(DIALOG,cs_editor_mode_design)
              NEXT FIELD lbl_dzel001
            END IF
          ELSE
            -- 已經出貨表格無法進行欄位順序拖曳
            IF m_dzea_t[mi_dzea_index].dzea017 = "Y" THEN
              LET ls_replace_arg = m_dzea_t[mi_dzea_index].dzea001,"|"
              IF ms_dgenv = cs_dgenv_standard THEN
                CALL sadzp000_msg_show_error('adz-00461', 'adz-00461', ls_replace_arg, 1)
              END IF
              IF ms_dgenv = cs_dgenv_customize THEN
                CALL sadzp000_msg_show_error('adz-00462', 'adz-00462', ls_replace_arg, 1)
              END IF
            END IF
          END IF
        ELSE
          IF adzi140_check_if_check_out_active(mi_dzea_index) THEN
            LET drop_index = lo_dnd.getLocationRow()
            LET ls_table_name = m_dzea_t[mi_dzea_index].dzea001
            LET ls_module_name = m_dzea_t[mi_dzea_index].dzea003 #20170425
            CALL sadzi140_db_get_GUID() RETURNING ls_group_id
            LET m_old_dzeb_t.* = m_dzeb_t.*

            IF drag_value LIKE "cdf%" THEN
              #已出貨則只能 append 在最後
              LET lb_insert = IIF(m_dzea_t[mi_dzea_index].dzea017 = "Y",FALSE,TRUE)
              CALL adzi140_drag_column_define_to_table(DIALOG,ls_module_name,ls_table_name,drag_value,drop_index,ls_group_id,lb_insert) #20170425
            ELSE
              CALL adzi140_drag_column_define_from_group(DIALOG,ls_module_name,ls_table_name,drag_value,drop_index,ls_group_id) #20170425
            END IF

            #有拖拉新欄位, 則將所有的欄位都標示為已修改
            FOR li_loop = 1 TO m_dzeb_t.getLength()
              IF (NVL(m_dzeb_t[li_loop].recordtype,cs_null_default) <> cs_data_drag) AND
                 (NVL(m_dzeb_t[li_loop].recordtype,cs_null_default) <> cs_data_added) THEN
                LET m_dzeb_t[li_loop].recordtype = NVL(m_dzeb_t[li_loop].recordtype,cs_data_modified)
              END IF
            END FOR

            IF NOT mb_drop_confirm THEN
              LET mb_drop_confirm = TRUE
              #CALL sadzp000_msg_show_info('adz-00100', 'adz-00100', NULL, 0)
            END IF
            IF adzi140_get_editor_mode() <> cs_editor_mode_design THEN
              CALL adzi140_set_editor_mode(DIALOG,cs_editor_mode_design)
              NEXT FIELD lbl_dzel001
            END IF
          END IF
        END IF

    END DISPLAY

    #Table Index
    DISPLAY ARRAY m_dzec_t TO sr_TableIndex.*
      BEFORE ROW
        LET mi_dzec_index = ARR_CURR()
    END DISPLAY

    #Table Keys
    DISPLAY ARRAY m_dzed_t TO sr_TableKey.*
      BEFORE ROW
        LET mi_dzed_index = ARR_CURR()
    END DISPLAY

    #Table Privileges
    DISPLAY ARRAY m_dzen_t TO sr_TablePrivileges.*
      BEFORE ROW
        LET mi_dzen_index = ARR_CURR()
    END DISPLAY

    #輸入模組搜尋條件
    INPUT ls_module FROM cb_ModuleSelect ATTRIBUTE(WITHOUT DEFAULTS)
      ON CHANGE cb_ModuleSelect
        LET ms_module = ls_module
        CALL DIALOG.setCurrentRow("sr_tablelist",1)
        LET mi_dzea_index = 1
        #更新模組清單, 所以只傳入ms_module
        --LET ls_search = "" LET ms_search = ""
        CALL adzi140_refresh_master(DIALOG,NVL(ms_search,ms_module))
        CALL adzi140_refresh_detail(DIALOG,m_dzea_t[mi_dzea_index].dzea001)
    END INPUT

    #輸入搜尋條件
    INPUT ls_search FROM edt_Search ATTRIBUTE(WITHOUT DEFAULTS)
      AFTER INPUT
        LET ms_search = sadzi140_util_replace_string(ls_search,"'","")
      ON CHANGE edt_Search
        LET ms_search = sadzi140_util_replace_string(ls_search,"'","")
      ON ACTION ACCEPT
        #20170425 begin
        IF ls_search IS NULL THEN 
          LET ls_module = "*" 
          LET ms_module = ls_module
        ELSE
          LET ls_module = "" 
          LET ms_module = ls_module
        END IF  
        #20170425 end

        LET ls_get_issue_table = "N" #20170425
        LET mb_get_issue_table = IIF(ls_get_issue_table == "Y",TRUE,FALSE) #20170425
        LET ls_get_checkout_tables = "N" #20170425
        LET mb_get_checkout_tables = IIF(ls_get_checkout_tables == "Y",TRUE,FALSE) #20170425

        LET ms_search = sadzi140_util_replace_string(ls_search,"'","")
        CALL DIALOG.setCurrentRow("sr_tablelist",1)
        CLEAR sr_tablelist.*
        #搜尋, 所以只傳入ms_search
        CALL adzi140_refresh_master(DIALOG,NVL(ms_search,ms_module))
        LET mi_dzea_index = 1
        LET ms_table_name = m_dzea_t[mi_dzea_index].dzea001 #20170801
        LET ls_table_name = ms_table_name #20170801
        CALL adzi140_refresh_detail(DIALOG,m_dzea_t[mi_dzea_index].dzea001)
        #填充欄位定義清單
        CALL adzi140_fill_dzel_t(ms_dgenv,m_dzea_t[mi_dzea_index].dzea014) #20190304
    END INPUT

    #輸入搜尋條件
    INPUT ls_get_issue_table,
          ls_get_industry_table #20190415
     FROM cb_get_issue_tables,
          cb_get_industry_table #20190415
     ATTRIBUTE(WITHOUT DEFAULTS) 
      ON CHANGE cb_get_issue_tables
        LET mb_get_issue_table = IIF(ls_get_issue_table == "Y",TRUE,FALSE)
        CALL adzi140_refresh_master(DIALOG,NVL(ms_search,ms_module))
        LET mi_dzea_index = 1
        CALL adzi140_refresh_detail(DIALOG,m_dzea_t[mi_dzea_index].dzea001)
      #20190415 begin  
      ON CHANGE cb_get_industry_table
        LET mb_get_industry_table = IIF(ls_get_industry_table == "Y",TRUE,FALSE)
        CALL adzi140_refresh_master(DIALOG,NVL(ms_search,ms_module))
        LET mi_dzea_index = 1
        CALL adzi140_refresh_detail(DIALOG,m_dzea_t[mi_dzea_index].dzea001)
      #20190415 end  
    END INPUT

    #20170425 begin
    INPUT ls_get_checkout_tables FROM cb_get_checkout_tables ATTRIBUTE(WITHOUT DEFAULTS)
      ON CHANGE cb_get_checkout_tables
        --LET ms_search = "" LET ls_search = ""
        --LET ms_module = "" LET ls_module = ""
        --LET lo_combobox = ui.ComboBox.forName("formonly.cb_moduleselect")
        --LET ls_module = lo_combobox.getItemName(lo_combobox.getItemCount()+1)
        LET mb_get_checkout_tables = IIF(ls_get_checkout_tables == "Y",TRUE,FALSE)
        CALL adzi140_refresh_master(DIALOG,NVL(ms_search,ms_module))
        LET mi_dzea_index = 1
        CALL adzi140_refresh_detail(DIALOG,m_dzea_t[mi_dzea_index].dzea001)
    END INPUT
    #20170425 end

    BEFORE DIALOG
      LET ls_get_issue_table = "N"
      LET ls_get_industry_table = "N" #20170425
      LET ls_get_checkout_tables = "N" #20170425
      LET lo_window = ui.Window.getCurrent()
      LET lo_form   = lo_window.getForm() #20190415
      LET lb_filter_hidden = TRUE #20190415
      CALL sadzp000_msg_get_message('adz-01582',ms_lang) RETURNING ls_adv_filter_text #20190415
      CALL lo_form.setElementText("btn_adv_filter","︾  "||ls_adv_filter_text) #20190415 
      #CALL adzi140_set_editor_mode(DIALOG,cs_editor_mode_browse) #20170706 mark

      {
      #20170425 mark begin
      CALL adzi140_get_check_out_table_list(ms_user) RETURNING ls_table_list
      IF ls_table_list.trim() IS NOT NULL THEN
        LET ls_get_checkout_tables = "Y"
      ELSE  
        LET ls_get_checkout_tables = "N"
        #找不到簽出資訊則秀第一個模組
        LET lo_combobox = ui.ComboBox.forName("formonly.cb_moduleselect")
        LET ls_module = lo_combobox.getItemName(1)
      END IF
      LET mb_get_checkout_tables = IIF(ls_get_checkout_tables == "Y",TRUE,FALSE)
      #20170425 mark end
      }

      {
      #20170425 mark begin
      #秀第一個模組
      LET lo_combobox = ui.ComboBox.forName("formonly.cb_moduleselect")
      LET ls_module = lo_combobox.getItemName(2)
      LET ms_module = ls_module
      CALL adzi140_fill_dzea_t(ms_module) #20170425 
      #20170425 mark end
      }

      #填充欄位定義清單
      CALL adzi140_fill_dzel_t(ms_dgenv,ms_topind) #20190304
      #填充表格清單 
      CALL adzi140_fill_dzea_t(null) #20170425 
      LET mi_dzea_index = ARR_CURR() #20170706
      LET ms_table_name = m_dzea_t[mi_dzea_index].dzea001 #20170706
      LET ls_table_name = ms_table_name #20170706
      CALL DIALOG.setArrayAttributes("sr_tablelist", m_dzea_t_color)
      CALL DIALOG.setSelectionMode("sr_tablecolumns",TRUE)
      #20170706 begin
      #設定 Import 功能及建立function table 的功能是否 Enable
      CALL adzi140_set_import_sco_enable(DIALOG)
      CALL adzi140_set_editor_mode(DIALOG,cs_editor_mode_browse)
      CALL adzi140_set_editable(DIALOG,mi_dzea_index)
      CALL adzi140_set_check_out_mode(DIALOG,m_dzea_t[mi_dzea_index].*)
      #20170706 end
      
    ON ACTION btn_table_detail
      {
      MENU "table_detail" ATTRIBUTES (STYLE="popup")
        ON ACTION refresh
        ON ACTION cancel
      END MENU
      }

    #20190415 begin
    ON ACTION btn_adv_filter
      LET lb_filter_hidden = NOT lb_filter_hidden
      CALL lo_form.setElementHidden("grid_adv_filter", lb_filter_hidden) 
      CALL lo_form.setElementText("btn_adv_filter",IIF(lb_filter_hidden,"︾  "||ls_adv_filter_text,"︽  "||ls_adv_filter_text))  
    #20190415 end
    
    ON ACTION find_data
      #20170425 begin
      IF ls_search IS NULL THEN 
        LET ls_module = "*" 
        LET ms_module = ls_module
      ELSE
        LET ls_module = "" 
        LET ms_module = ls_module
      END IF  
      #20170425 end
      
      LET ls_get_issue_table = "N" #20170425
      LET mb_get_issue_table = IIF(ls_get_issue_table == "Y",TRUE,FALSE) #20170425
      LET ls_get_checkout_tables = "N" #20170425
      LET mb_get_checkout_tables = IIF(ls_get_checkout_tables == "Y",TRUE,FALSE) #20170425
      
      CALL DIALOG.setCurrentRow("sr_tablelist",1)
      CLEAR sr_tablelist.*
      LET ms_search = sadzi140_util_replace_string(ls_search,"'","")
      #搜尋, 所以只傳入ms_search
      CALL adzi140_refresh_master(DIALOG,NVL(ms_search,ms_module))
      LET mi_dzea_index = 1
      LET ms_table_name = m_dzea_t[mi_dzea_index].dzea001 #20170801
      LET ls_table_name = ms_table_name #20170801
      CALL adzi140_refresh_detail(DIALOG,m_dzea_t[mi_dzea_index].dzea001)
      NEXT FIELD lbl_dzea001

    ############################################################################
    #複製表格
    ON ACTION tbi_clone_table
      IF m_dzea_t[IIF(mi_dzea_index==0,1,mi_dzea_index)].dzea001 IS NULL THEN 
        CALL sadzp000_msg_show_error("adz-00824", "adz-00824", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF
      LET mb_clone_table = TRUE
      LET ls_source_Table = m_dzea_t[mi_dzea_index].dzea001
      CALL sadzi140_db_get_module_code_by_dgenv(m_dzea_t[mi_dzea_index].dzea003,ms_dgenv) RETURNING ls_module_code
      INITIALIZE lo_create_table TO NULL #20170911
      LET lo_create_table.dct_module_name = ls_module_code
      LET lo_create_table.dct_table_type = m_dzea_t[mi_dzea_index].dzea004
      LET lo_create_table.dct_orig_module_name = m_dzea_t[mi_dzea_index].dzea016
      CALL adzi140_create_new_table_wizard(DIALOG,lo_create_table.*,ls_source_Table) RETURNING lb_result,lo_create_table.*
      LET ls_search = lo_create_table.dct_table_name
      LET ms_search = ls_search
      CALL adzi140_refresh_and_locate_master(DIALOG,lo_create_table.dct_table_name)
      INITIALIZE lo_create_table TO NULL

    #建立新表格
    ON ACTION tbi_create_new_table
      LET mb_clone_table = FALSE
      LET ls_source_Table = NULL #20170911
      LET ls_table_name = m_dzea_t[mi_dzea_index].dzea001
      INITIALIZE lo_create_table TO NULL #20170911
      CALL adzi140_create_new_table_wizard(DIALOG,lo_create_table.*,ls_source_Table) RETURNING lb_result,lo_create_table.*
      LET ls_search = IIF(ls_search.trim() IS NULL,lo_create_table.dct_table_name,ls_search||','||lo_create_table.dct_table_name) -- NVL(ls_search,lo_create_table.dct_table_name)
      LET ms_search = ls_search
      CALL adzi140_refresh_and_locate_master(DIALOG,NVL(lo_create_table.dct_table_name,ls_table_name))
      #填充欄位定義清單
      CALL adzi140_fill_dzel_t(ms_dgenv,m_dzea_t[mi_dzea_index].dzea014) #20190304
      INITIALIZE lo_create_table TO NULL
      
    ############################################################################
    #建立附屬表格
    ON ACTION tbi_create_function_table
      IF m_dzea_t[IIF(mi_dzea_index==0,1,mi_dzea_index)].dzea001 IS NULL THEN 
        CALL sadzp000_msg_show_error("adz-00824", "adz-00824", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF
      LET lb_result = TRUE
      INITIALIZE lo_step TO NULL
      CALL adzi140_set_step_data(FALSE,TRUE,FALSE,FALSE) RETURNING lo_step.*
      IF (m_dzea_t[mi_dzea_index].DZEA004 = 'L') OR (m_dzea_t[mi_dzea_index].DZEA004 = 'V') THEN
        LET ls_replace_arg = m_dzea_t[mi_dzea_index].dzea001,"|"
        CALL sadzp000_msg_show_error('adz-00093', 'adz-00093', ls_replace_arg, 1)
      ELSE
        #呼叫ALM
        CALL adzi140_step_call_alm_check_out(mb_enable_alm,ms_user,ms_lang,cs_user_role_sd,lo_step.*,TRUE) RETURNING lb_result,lo_step.*,lo_USER_INFO.*,lo_DZLU_T

        {
        #20190304 mark begin
        #若在標準環境, 且簽出行業別和程式行業別不同的, 不能簽出
        #IF (lb_result) AND (mb_enable_alm) AND (ms_DGENV = cs_dgenv_standard) AND (NVL(lo_DZLU_T[1].DZLU008,cs_null_default) <> NVL(m_dzea_t[mi_dzea_index].dzea014,cs_null_default)) THEN #20160803 marked
        IF (lb_result) AND (mb_enable_alm) AND (ms_DGENV = cs_dgenv_standard) AND (NVL(lo_DZLU_T[1].DZLU008,NVL(ms_topind,cs_default_topind)) <> NVL(m_dzea_t[mi_dzea_index].dzea014,NVL(ms_topind,cs_default_topind))) THEN #20160803 Modified
          LET ls_replace_arg = ""
          CALL sadzp000_msg_show_error("adz-00516", "adz-00516", ls_replace_arg, 1)
          CONTINUE DIALOG
        END IF
        #20190304 mark end
        }

        IF (lb_result) AND (lo_step.NEXT) THEN
          CALL adzi140_step_create_table_type_list(DIALOG,m_dzea_t[mi_dzea_index].*,lo_step.*) RETURNING lb_result, lo_step.*, lo_function_table_list
          IF lb_result THEN
            FOR li_index = 1 TO lo_function_table_list.getLength()
              IF NVL(lo_function_table_list[li_index].p_checkbox,cs_null_value) = "Y" THEN

                #ALM 給定編號
                CALL adzi140_step_get_alm_info(lo_function_table_list[li_index].p_table_name,lo_function_table_list[li_index].p_module_name,lo_function_table_list[li_index].p_table_type_desc,cs_spec_type_table,lo_USER_INFO.ui_ROLE,lo_DZLU_T) RETURNING lb_result, lo_dzlm_t.*
                IF NOT lb_result THEN EXIT FOR END IF

                ------------------------------- end ALM 給定編號 -----------------------------
                #給定ALM資訊
                LET ls_alm_version    = lo_dzlm_t.DZLM005
                LET ls_alm_sd_version = lo_dzlm_t.DZLM006
                LET ls_alm_request_no = lo_dzlm_t.DZLM012
                LET ls_dgenv          = ms_dgenv

                #建立表格及回傳 Key 清單
                CALL adzi140_create_auxilary_table(lo_function_table_list[li_index].*,ls_alm_version,ls_alm_sd_version,ls_alm_request_no,ls_dgenv) RETURNING ls_table_name

              END IF
            END FOR
            #CALL adzi140_refresh_master(DIALOG,ms_search)
            IF NOT lb_result THEN
              LET ls_replace_arg = lo_function_table_list[mi_dzea_index].p_table_name,"|"
              CALL sadzp000_msg_show_error('adz-00364', 'adz-00364', ls_replace_arg, 1)
            END IF
            LET ls_search = NVL(ls_search,ls_table_name)
            LET ms_search = ls_search
            LET ls_replace_arg = ls_table_name,"|" #20170523
            CALL sadzp000_msg_show_message_bar("adz-01201",ls_replace_arg,ms_lang) #20170523
            CALL adzi140_refresh_and_locate_master(DIALOG,ls_table_name)
          END IF
        END IF
      END IF

    ############################################################################

    ON ACTION tbi_delete_table
      IF m_dzea_t[IIF(mi_dzea_index==0,1,mi_dzea_index)].dzea001 IS NULL THEN 
        CALL sadzp000_msg_show_error("adz-00824", "adz-00824", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF
      #20190304 begin
      IF (ms_topind <> cs_default_topind) AND (m_dzea_t[mi_dzea_index].dzea014 <> ms_topind) THEN 
        LET ls_msg_no = "adz-01577"
        CALL sadzp000_msg_show_error(ls_msg_no, ls_msg_no, NULL, 1)
        CONTINUE DIALOG
      END IF
      #20190304 end
      LET li_arr_curr    = m_dzea_t[mi_dzea_index].dzeaseq
      LET ls_table_name  = m_dzea_t[mi_dzea_index].dzea001
      LET ls_table_type  = m_dzea_t[mi_dzea_index].dzea004
      LET ls_master_user = ms_master_user
      LET ls_replace_arg = ls_table_name,"|"
      CALL sadzp000_msg_alert_box("adz-00114", "adz-00114", ls_replace_arg, 0) RETURNING ls_question
      IF ls_question = cs_question_yes THEN
        LET lb_DropSuccess = FALSE
        CALL sadzi140_util_delete_master_table(ms_master_user,"D",ls_table_name,FALSE) RETURNING lb_DropSuccess
        IF lb_DropSuccess THEN
          #20160617 begin
          IF NVL(m_dzea_t[mi_dzea_index].dzea017,"N") = "Y" THEN 
            #已出貨的表格要刪除則會紀錄
            CALL sadzi140_util_set_log_for_shipped_table(ls_table_name,ms_user)
          END IF  
          #20160617 end
          --CALL adzi140_refresh_master(DIALOG,ms_search)
          CALL sadzi140_db_gen_db_schema(ls_master_user.toLowerCase(),ls_table_name,ls_table_type) RETURNING lb_result
          IF NOT lb_result THEN
            CALL sadzp000_msg_show_error('adz-00381', 'adz-00381', '', 1)
          END IF
        END IF
      END IF
      --CALL DIALOG.setCurrentRow("sr_tablelist",li_arr_curr)
      --CALL adzi140_refresh_master(DIALOG,ms_search)
      CALL sadzp000_msg_show_message_bar(NULL,NULL,ms_lang) #20170523
      CALL adzi140_refresh_and_locate_master(DIALOG,IIF(lb_DropSuccess,ms_search,ls_table_name))

    #begin Column define work
    ON ACTION btn_AddCDFConfirm
      LET ls_table_name = m_dzea_t[mi_dzea_index].dzea001
      FOR li_index = 1 TO m_dzeb_t.getLength()
        IF m_dzeb_t[li_index].RECORDTYPE = cs_data_drag THEN
          LET m_dzeb_t[li_index].RECORDTYPE = cs_data_added
        END IF
      END FOR
      CALL adzi140_alter_column_data(DIALOG,FALSE) RETURNING lb_data_altered,lb_dzeb006_changed,lb_dzeb003_changed #20180412
      CALL adzi140_refresh_and_locate_master(DIALOG,ls_table_name)
      CALL DIALOG.setCurrentRow("sr_tablecolumns",1)
      CALL adzi140_set_editable(DIALOG,mi_dzea_index)

    ON ACTION btn_CancelCDF
      LET ls_table_name = m_dzea_t[mi_dzea_index].dzea001
      CALL adzi140_refresh_and_locate_master(DIALOG,ls_table_name)
      #CALL adzi140_refresh_detail(DIALOG,ls_table_name)
      #設定 Import 功能及建立function table 的功能是否 Enable
      CALL adzi140_set_import_sco_enable(DIALOG)
      CALL adzi140_set_editor_mode(DIALOG,cs_editor_mode_browse)
      CALL adzi140_set_editable(DIALOG,mi_dzea_index)
    #end Column define work

    #20160825 begin
    #權限賦予在客戶端可以隨時執行修改, 但是在產中則要簽出
    ON ACTION modify_grant
      IF m_dzea_t[IIF(mi_dzea_index==0,1,mi_dzea_index)].dzea001 IS NULL THEN 
        CALL sadzp000_msg_show_error("adz-00824", "adz-00824", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF
      CALL adzi140_set_editor_mode(DIALOG,cs_editor_mode_browse)
      CALL adzi140_set_editable(DIALOG,mi_dzea_index)
      IF mb_modifiable THEN 
        LET ls_current_item = DIALOG.getCurrentItem()
        GOTO _ACTION_MODIFY
      END IF
    #20160825 end
    
    #20180302 begin
    ON ACTION tm_fix_column_sequence
      IF m_dzea_t[IIF(mi_dzea_index==0,1,mi_dzea_index)].dzea001 IS NULL THEN 
        CALL sadzp000_msg_show_error("adz-00824", "adz-00824", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF

      #檢核欄位和索引與實際表格有沒有差異
      CALL sadzi140_shot_check_diff(ls_master_user,ls_table_name,mb_drop_abnormal_column) RETURNING lb_diff
      IF lb_diff THEN
        CALL sadzp000_msg_show_error("adz-00135", "adz-00135", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF
      
      LET ls_replace_arg = m_dzea_t[mi_dzea_index].dzea001,"|"
      CALL sadzp000_msg_alert_box("adz-01337", "adz-01337", ls_replace_arg, 0) RETURNING ls_question
      IF ls_question = cs_question_no THEN CONTINUE DIALOG END IF

      LET ls_table_name  = m_dzea_t[mi_dzea_index].dzea001
      CALL sadzi140_util_fix_design_data_column_sequence(ls_table_name) RETURNING lb_result      
      -- 檢核r.t資料是否和資料庫中有差異
      CALL adzi140_check_db_data_diff(ls_master_user,ls_table_name) RETURNING lb_result
      CALL adzi140_refresh_and_locate_master(DIALOG,ls_table_name)
      
    ON ACTION tm_del_std_column
      IF m_dzea_t[IIF(mi_dzea_index==0,1,mi_dzea_index)].dzea001 IS NULL THEN 
        CALL sadzp000_msg_show_error("adz-00824", "adz-00824", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF
      
      LET ls_replace_arg = m_dzea_t[mi_dzea_index].dzea001,"|"
      CALL sadzp000_msg_alert_box("adz-01336", "adz-01336", ls_replace_arg, 0) RETURNING ls_question
      IF ls_question = cs_question_no THEN CONTINUE DIALOG END IF
      
      LET ls_current_item = "sr_tablecolumns"
      LET ls_table_name  = m_dzea_t[mi_dzea_index].dzea001
      INITIALIZE lo_left_array TO NULL
      INITIALIZE lo_right_array TO NULL
      CALL adzi140_get_standard_column_list(ls_table_name) RETURNING lo_left_array
      CALL sadzi140_cols_run(lo_left_array,lo_right_array) RETURNING ls_columns,lo_pick_col_arr
      IF lo_pick_col_arr.getLength() > 0 THEN 
        LET ls_del_list = ""
        FOR li_loop = 1 TO lo_pick_col_arr.getLength()
          LET ls_del_list = ls_del_list,"'",lo_pick_col_arr[li_loop].COLUMN_NAME,"'",","
        END FOR
        LET ls_del_list = ls_del_list.subString(1,ls_del_list.getLength()-1)
        CALL sadzi140_util_set_delete_standard_column_info(ls_table_name,lo_pick_col_arr) RETURNING lb_result
        IF lb_result THEN 
          LET lb_data_altered = TRUE
          CALL adzi140_delete_column_design_datas(ls_table_name,ls_del_list) RETURNING lb_result
          IF NOT lb_result THEN 
            CALL sadzi140_util_recover_standard_column_info(ls_table_name,lo_pick_col_arr) RETURNING lb_result
          END IF
        END IF  
        CALL adzi140_refresh_and_locate_master(DIALOG,ls_table_name)
        GOTO _ACTION_MODIFY
      END IF
      
    #20180302 end
    
    #begin Alter data process
    ON ACTION MODIFY
      LET lb_data_altered = FALSE #20180302
      LET ls_current_item = NULL

      LABEL _ACTION_MODIFY: #20160825

      IF m_dzea_t[IIF(mi_dzea_index==0,1,mi_dzea_index)].dzea001 IS NULL THEN 
        CALL sadzp000_msg_show_error("adz-00824", "adz-00824", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF
      LET ls_master_user = ms_master_user
      LET ls_table_name  = m_dzea_t[mi_dzea_index].dzea001
      LET ls_module_name = m_dzea_t[mi_dzea_index].dzea003
      LET ls_table_desc  = m_dzea_t[mi_dzea_index].dzea002
      LET ls_current_item = IIF(ls_current_item IS NULL,DIALOG.getCurrentItem(),ls_current_item) #20180302
      CASE
        WHEN ls_current_item = "sr_tablelist"
          #20190304 begin
          IF (ms_topind <> cs_default_topind) AND (m_dzea_t[mi_dzea_index].dzea014 <> ms_topind) THEN 
            LET ls_msg_no = "adz-01575"
            CALL sadzp000_msg_show_error(ls_msg_no, ls_msg_no, NULL, 1)
            CONTINUE DIALOG
          END IF
          #20190304 end
          LET mi_dzea_index = ARR_CURR()
          CALL adzi140_modify_master_table(DIALOG) RETURNING lb_data_altered
          LET lb_data_altered = FALSE -- 目前只針對Column PK, 及 FK 異動呼叫 adzp144
        WHEN ls_current_item = "sr_tablecolumns"
          CALL adzi140_alter_column_data(DIALOG,lb_data_altered) RETURNING lb_data_altered,lb_dzeb006_changed,lb_dzeb003_changed  #20180302 #20180412
          #如果型態有異動, 就做備份
          IF lb_dzeb006_changed THEN
            CALL sadzi140_util_create_curr_snapshot(ls_master_user,ls_table_name,ls_module_name,ls_table_desc,ms_dgenv)
          END IF
          { 
          #20180412 begin
          #如果有修改多語言, 就提醒詢問是否整批產生相關畫面多語言
          IF lb_dzeb003_changed THEN
            LET ls_replace_arg = ls_table_name,"|"
            CALL sadzp000_msg_alert_box("adz-01355", "adz-01355", ls_replace_arg, 0) RETURNING ls_question
            IF ls_question = cs_question_yes THEN
              LET ls_program_name = "adzp130"
              LET ls_parameter = "r.r ",ls_program_name," -T ",ls_table_name
              &ifndef DEBUG
              IF sadzi140_util_check_program_exists(ls_program_name) THEN
                CALL cl_cmdrun_openpipe("r.r",ls_parameter,FALSE) RETURNING lb_return,ls_retmsg
              END IF
              &endif
            END IF            
          END IF
          #20180412 end
          }
        WHEN ls_current_item = "sr_tablekey"
          #20190304 begin
          IF (ms_topind <> cs_default_topind) AND (m_dzea_t[mi_dzea_index].dzea014 <> ms_topind) THEN 
            LET ls_msg_no = "adz-01572"
            CALL sadzp000_msg_show_error(ls_msg_no, ls_msg_no, NULL, 1)
            CONTINUE DIALOG
          END IF
          #20190304 end
          CALL adzi140_alter_key_data(DIALOG) RETURNING lb_data_altered
        WHEN ls_current_item = "sr_tableindex"
          #20190304 begin
          IF (ms_topind <> cs_default_topind) AND (m_dzea_t[mi_dzea_index].dzea014 <> ms_topind) THEN 
            LET ls_msg_no = "adz-01574"
            CALL sadzp000_msg_show_error(ls_msg_no, ls_msg_no, NULL, 1)
            CONTINUE DIALOG
          END IF
          #20190304 end
          CALL adzi140_alter_index_data(DIALOG) RETURNING lb_data_altered
          LET lb_data_altered = FALSE -- 目前只針對Column PK, 及 FK 異動呼叫 adzp144
        WHEN ls_current_item = "sr_tableprivileges"
          #20190304 begin
          IF (ms_topind <> cs_default_topind) AND (m_dzea_t[mi_dzea_index].dzea014 <> ms_topind) THEN 
            LET ls_msg_no = "adz-01573"
            CALL sadzp000_msg_show_error(ls_msg_no, ls_msg_no, NULL, 1)
            CONTINUE DIALOG
          END IF
          #20190304 end
          CALL adzi140_alter_privilege_data(DIALOG) RETURNING lb_data_altered
          IF lb_data_altered THEN
            IF sadzi140_shot_check_privileges_if_diff(ls_table_name) THEN
              CALL sadzi140_util_grant_revoke_privileges(ls_table_name) RETURNING ls_message
              CALL sadzp000_msg_show_info('adz-00880', 'adz-00880', NULL, 0)
              #有 APS 的 grant 要重產該檔案及通知 user 過單
              IF adzi140_check_if_exist_aps_grant(m_dzen_t) THEN 
                CALL sadzi140_util_regen_grant_aps_privilege_file() RETURNING ls_file_name
                #標準環境才顯示訊息
                IF (ms_DGENV = cs_dgenv_standard) THEN  
                  LET ls_replace_arg = ls_file_name,"|"
                  CALL sadzp000_msg_show_info('adz-00846', 'adz-00846', ls_replace_arg, 0)
                END IF  
              END IF  
            END IF
          END IF
          NEXT FIELD lbl_dzea001 #20160825
      END CASE

      -- 針對Column PK, 及 FK 異動呼叫 adzp144
      IF (lb_data_altered) AND (ls_current_item = "sr_tablecolumns") THEN
        &ifndef DEBUG
          #產出tables.xml檔
          #CALL sadzi140_xml_get_lang_type_list() RETURNING lo_lang_arr
          CALL sadzi140_crud_get_static_lang_list() RETURNING lo_lang_arr
          FOR li_loop = 1 TO lo_lang_arr.getLength()
            LET ls_lang = lo_lang_arr[li_loop]
            LET ls_table_xml = ls_top_env,ls_separator,"com",ls_separator,"mta",ls_separator,ls_lang,ls_separator,"tables.xml"
            CALL sadzi140_xml_regen_base_tables_data(ls_table_xml,ls_lang,ls_table_name) RETURNING lb_result #20170908
          END FOR
        &else
          LET ls_table_xml = "c:\\temp\\tbl\\tables.xml"
          CALL sadzi140_xml_regen_base_tables_data(ls_table_xml,ls_lang,ls_table_name) RETURNING lb_result #20170908
        &endif
        CALL sadzi140_util_collect_affect_prog_list(ls_table_name)
      END IF

      CALL adzi140_refresh_and_locate_master(DIALOG,ls_table_name)
      CALL adzi140_set_editable(DIALOG,mi_dzea_index)
    #end Alter data process

    #begin Top Menu Actions
    ON ACTION tm_gen_sql
      LET ls_table_name  = m_dzea_t[mi_dzea_index].dzea001
      LET ls_module_name = m_dzea_t[mi_dzea_index].DZEA003
      #產出 table.sch 檔
      #CALL sadzi140_xml_get_lang_type_list() RETURNING lo_lang_arr
      CALL sadzi140_crud_get_static_lang_list() RETURNING lo_lang_arr
      FOR li_loop = 1 TO lo_lang_arr.getLength()
        LET ls_lang = lo_lang_arr[li_loop]
        CALL sadzi140_db_gen_table_schema(ls_table_name,ls_module_name,FALSE,ls_lang) RETURNING ls_schema_sql
      END FOR
      LET ls_replace_arg = ls_table_name,"|"
      CALL sadzp000_msg_show_info('adz-00101', 'adz-00101', ls_replace_arg, 0)

    ON ACTION tm_batch_gen_sql
      #批次產出 {table}.sch 檔
      #CALL sadzi140_xml_get_lang_type_list() RETURNING lo_lang_arr
      CALL sadzi140_crud_get_static_lang_list() RETURNING lo_lang_arr
      &ifndef DEBUG
      CALL cl_progress_bar(m_dzea_t.getLength()*lo_lang_arr.getLength())
      &endif
      FOR li_index = 1 TO m_dzea_t.getLength()
        LET ls_table_name  = m_dzea_t[li_index].dzea001
        LET ls_module_name = m_dzea_t[li_index].DZEA003
        FOR li_loop = 1 TO lo_lang_arr.getLength()
          LET ls_lang = lo_lang_arr[li_loop]
          &ifndef DEBUG
          CALL cl_progress_ing(ls_lang||ls_os_separator||ls_table_name||'.sch')
          &endif
          CALL sadzi140_db_gen_table_schema(ls_table_name,ls_module_name,TRUE,ls_lang) RETURNING ls_schema_sql
        END FOR
      END FOR
      CALL sadzp000_msg_show_info('adz-00102', 'adz-00102', NULL, 0)

    ON ACTION tm_gen_42s
      LET ls_module_name  = "adz"
      LET ls_program_name = "adzi140"
      CALL FGL_GETENV(ls_module_name.toUpperCase()) RETURNING ls_module_path
      LET ls_4gl_path = ls_module_path,ls_os_separator,"4gl"
      CALL os.Path.pwd() RETURNING ls_pwd_path
      CALL sadzi140_lbl_run(ls_module_name,ls_program_name,ms_lang)
      &ifndef DEBUG
      CALL os.Path.chdir(ls_4gl_path) RETURNING lb_success
      LET ls_command = "r.c ",ls_program_name
      RUN ls_command RETURNING li_return
      LET lb_success = IIF(li_return==0,TRUE,FALSE)
      LET ls_command = "r.l ",ls_program_name
      RUN ls_command RETURNING li_return
      LET lb_success = IIF(li_return==0,TRUE,FALSE)
      &endif
      CALL os.Path.chdir(ls_pwd_path) RETURNING lb_success
      CALL sadzp000_msg_show_info('adz-00217', 'adz-00217', NULL, 0)

    ON ACTION tm_gen_table_list
      LET ls_table_name = m_dzea_t[mi_dzea_index].dzea001
      LET ls_gen_tbl_list = "" #20170908
      LET lb_gen_tbl = TRUE #20170908
      &ifndef DEBUG
        #產出tables.xml檔
        #CALL sadzi140_xml_get_lang_type_list() RETURNING lo_lang_arr
        CALL sadzi140_crud_get_static_lang_list() RETURNING lo_lang_arr
        CALL cl_progress_bar(lo_lang_arr.getLength())
        FOR li_loop = 1 TO lo_lang_arr.getLength()
          LET ls_lang = lo_lang_arr[li_loop]
          LET ls_table_xml = ls_top_env,ls_separator,"com",ls_separator,"mta",ls_separator,ls_lang,ls_separator,"tables.xml"
          CALL cl_progress_ing(ls_table_xml)
          CALL sadzi140_xml_gen_table_catalog(ls_table_xml,ls_lang) RETURNING lb_result
          #20170908 begin
          IF NOT lb_result THEN
            LET ls_gen_tbl_list = ls_gen_tbl_list,ls_table_xml,",","\n"
          END IF
          LET lb_gen_tbl = IIF(lb_gen_tbl,lb_result,lb_gen_tbl)
          #20170908 end
        END FOR
      &else
        LET ls_table_xml = "c:\\temp\\tbl\\tables.xml"
        CALL sadzi140_xml_gen_table_catalog(ls_table_xml,ms_lang) RETURNING lb_result
        #20170908 begin
        IF NOT lb_result THEN
          LET ls_gen_tbl_list = ls_gen_tbl_list,ls_table_xml,",","\n"
        END IF
        LET lb_gen_tbl = IIF(lb_gen_tbl,lb_result,lb_gen_tbl)
        #20170908 end
      &endif
      #20170908 begin
      IF NOT lb_gen_tbl THEN 
        LET ls_msg_no = "adz-01240"
        LET ls_replace_arg = "\n",ls_gen_tbl_list,"\n","|"
        CALL sadzp000_msg_show_error(ls_msg_no, ls_msg_no, ls_replace_arg, 1)
      END IF
      #20170908 end
      #資料表清單 "tables.xml" 產出程序結束 !!
      CALL sadzp000_msg_show_info('adz-00105', 'adz-00105', NULL, 0)

    ON ACTION tm_gen_table_structure
      IF m_dzea_t[IIF(mi_dzea_index==0,1,mi_dzea_index)].dzea001 IS NULL THEN 
        CALL sadzp000_msg_show_error("adz-00824", "adz-00824", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF
      LET ls_table_name  = m_dzea_t[mi_dzea_index].dzea001
      LET ls_module_name = m_dzea_t[mi_dzea_index].DZEA003
      LET ls_module_env = FGL_GETENV(ls_module_name)
      LET ls_gen_tbl_list = "" #20170908
      LET lb_gen_tbl = TRUE #20170908
      CALL sadzi140_xml_update_widget_width()
      -- 依照多語言產出對應的 tbl 檔
      #CALL sadzi140_xml_get_lang_type_list() RETURNING lo_lang_arr
      CALL sadzi140_crud_get_static_lang_list() RETURNING lo_lang_arr
      FOR li_loop = 1 TO lo_lang_arr.getLength()
        LET ls_lang = lo_lang_arr[li_loop]
        &ifndef DEBUG
        LET ls_tbl_Name = ls_module_env,ls_separator,"dzx",ls_separator,"tbl",ls_separator,ls_lang,ls_separator,ls_table_name,".tbl"
        &else
        LET ls_tbl_Name = "c:\\temp\\tbl\\",ls_lang,ls_separator,ls_table_name,".tbl"
        &endif
        CALL sadzi140_xml_gen_table_describe(ls_table_name,ls_tbl_Name,ls_lang) RETURNING lb_result
        #20170908 begin
        IF NOT lb_result THEN
          LET ls_gen_tbl_list = ls_gen_tbl_list,ls_tbl_Name,",","\n"
        END IF
        LET lb_gen_tbl = IIF(lb_gen_tbl,lb_result,lb_gen_tbl)
        #20170908 end
      END FOR
      #20170908 begin
      IF NOT lb_gen_tbl THEN 
        LET ls_msg_no = "adz-01240"
        LET ls_replace_arg = "\n",ls_gen_tbl_list,"\n","|"
        CALL sadzp000_msg_show_error(ls_msg_no, ls_msg_no, ls_replace_arg, 1)
      END IF
      #20170908 end
      #資料表結構 "ls_table_name".tbl 產出程序結束 !!
      LET ls_replace_arg = ls_table_name,"|"
      CALL sadzp000_msg_show_info('adz-00106', 'adz-00106', ls_replace_arg, 0)

    ON ACTION tm_gen_aps_grant_file
      CALL sadzi140_util_regen_grant_aps_privilege_file() RETURNING ls_file_name
      LET ls_replace_arg = "<br>",ls_file_name,"|"
      #CALL sadzp000_msg_show_info('adz-00847', 'adz-00847', ls_replace_arg, 0)
      CALL sadzp000_msg_show_info('adz-00991', 'adz-00991', ls_replace_arg, 0)
    
    ON ACTION tm_batch_gen_table_structure
      IF m_dzea_t[IIF(mi_dzea_index==0,1,mi_dzea_index)].dzea001 IS NULL THEN 
        CALL sadzp000_msg_show_error("adz-00824", "adz-00824", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF
      {
      #先刪除各模組的 tbl 檔案
      FOR li_index = 1 TO m_dzea_t.getLength()
        LET ls_module_name = m_dzea_t[li_index].DZEA003
        &ifndef DEBUG
        LET ls_module_env = FGL_GETENV(ls_module_name)
        LET ls_tbl_Name   = ls_module_env,ls_separator,"dzx",ls_separator,"tbl",ls_separator,ms_lang,ls_separator,"*.tbl"
        &else
        LET ls_module_env = os.Path.pwd()
        LET ls_tbl_Name   = ls_module_env,ls_separator,"tbl",ls_separator,"*.tbl"
        &endif
        CALL os.Path.delete(ls_tbl_Name) RETURNING lb_kill_tbl_files
        DISPLAY "Batch kill tbl files :",ls_tbl_Name," --> ",lb_kill_tbl_files
      END FOR
      }
      #建立各模組的 tbl 檔案
      CALL sadzi140_xml_update_widget_width()
      -- 依照多語言產出對應的 tbl 檔
      #CALL sadzi140_xml_get_lang_type_list() RETURNING lo_lang_arr
      CALL sadzi140_crud_get_static_lang_list() RETURNING lo_lang_arr
      &ifndef DEBUG
      CALL cl_progress_bar(m_dzea_t.getLength()*lo_lang_arr.getLength())
      &endif
      LET lb_gen_tbl = TRUE #20170908
      LET ls_gen_tbl_list = "" #20170908
      FOR li_index = 1 TO m_dzea_t.getLength()
        LET ls_table_name  = m_dzea_t[li_index].dzea001
        LET ls_module_name = m_dzea_t[li_index].DZEA003
        LET ls_module_env  = FGL_GETENV(ls_module_name)
        FOR li_loop = 1 TO lo_lang_arr.getLength()
          LET ls_lang = lo_lang_arr[li_loop]
          &ifndef DEBUG
          LET ls_tbl_Name = ls_module_env,ls_separator,"dzx",ls_separator,"tbl",ls_separator,ls_lang,ls_separator,ls_table_name,".tbl"
          CALL cl_progress_ing(ls_tbl_Name)
          &else
          LET ls_tbl_Name = "c:\\temp\\tbl\\",ls_lang,ls_separator,ls_table_name,".tbl"
          &endif
          DISPLAY "Generate tbl File :",ls_tbl_Name
          CALL sadzi140_xml_gen_table_describe(ls_table_name,ls_tbl_Name,ls_lang) RETURNING lb_result
          #20170908 begin
          IF NOT lb_result THEN
            LET ls_gen_tbl_list = ls_gen_tbl_list,ls_tbl_Name,",","\n"
          END IF
          LET lb_gen_tbl = IIF(lb_gen_tbl,lb_result,lb_gen_tbl)
          #20170908 end
        END FOR
      END FOR
      #20170908 begin
      IF NOT lb_gen_tbl THEN 
        LET ls_msg_no = "adz-01240"
        LET ls_replace_arg = "\n",ls_gen_tbl_list,"\n","|"
        CALL sadzp000_msg_show_error(ls_msg_no, ls_msg_no, ls_replace_arg, 1)
      END IF
      #20170908 end
      CALL sadzp000_msg_show_info('adz-00107', 'adz-00107', NULL, 0)

    ON ACTION tm_batch_exec_alter
      #20170714 begin
      LET ls_program_name = 'adzp090'
      IF sadzi140_util_check_program_exists(ls_program_name) THEN
        LET ls_command = "r.r ",ls_program_name
        RUN ls_command WITHOUT WAITING
      END IF
      #20170714 end
      CALL adzi140_refresh_and_locate_master(DIALOG,ls_table_name)

    ON ACTION tm_batch_table_rebuild
      IF m_dzea_t[IIF(mi_dzea_index==0,1,mi_dzea_index)].dzea001 IS NULL THEN 
        CALL sadzp000_msg_show_error("adz-01225", "adz-01225", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF
      &ifndef DEBUG
      CALL cl_progress_bar(m_dzea_t.getLength())
      &endif
      LET ls_table_list = ""
      LET ms_message = ""
      FOR li_index = 1 TO m_dzea_t.getLength()
        LET ls_master_user = ms_master_user
        LET ls_table_name  = m_dzea_t[li_index].dzea001
        LET ls_module_name = m_dzea_t[li_index].dzea003
        LET ls_table_desc  = m_dzea_t[li_index].dzea002
        LET lb_table_rebuild = FALSE

        LET ls_table_list = ls_table_list,",",ls_table_name

        &ifndef DEBUG
        CALL cl_progress_ing(ls_table_name)
        &endif

        #呼叫表格重建函式
        CALL sadzi140_util_table_rebuild(ls_master_user,ls_table_name,TRUE,ms_alm_version) RETURNING lb_table_rebuild,ls_rtn_code,ls_message
        LET ms_message = ms_message,"\n",ls_message
        CALL sadzi140_db_update_alter_code(ls_master_user,ls_table_name)
        IF m_dzea_t[li_index].dzea004 = "L" THEN
          #產生多語言 4gl 檔
          CALL sadzi140_util_gen_multi_lang(m_dzea_t[li_index].dzea001)
        END IF
        #產生延伸檔
        CALL sadzi140_util_gen_extend_inc(ls_table_name) RETURNING lb_result
        #重新編譯失效物件
        CALL sadzi140_util_recompile_invalid_db_objects(ls_table_name) RETURNING lb_result
        #Grant APS 權限 #20161102
        CALL sadzi140_util_grant_revoke_privileges(ls_table_name) RETURNING ls_message
      END FOR

      IF ls_table_list.trim() <> "" THEN
        LET ls_table_list = ls_table_list.subString(2,ls_table_list.getLength())
        CALL sadzi140_db_gen_db_schema(ls_master_user.toLowerCase(),ls_table_list,NULL) RETURNING lb_result
      END IF
      LET ls_error = ms_message.toUpperCase()
      IF (ls_error.getIndexOf("ERROR",1) > 0) THEN
        CALL sadzp000_msg_show_error('adz-00167', 'adz-00167', NULL, 1)
        CALL sadzi140_rev_view_logresult(ms_message)
      END IF
      #Grant APS 權限, 未來要移除
      CALL adzi140_grant_all_db_APS_privilege()
      #CALL adzi140_refresh_master(DIALOG,ms_search)
      CALL adzi140_refresh_and_locate_master(DIALOG,ls_table_name)

    {
    #移除此功能回歸到 azzi090
    ON ACTION tm_gen_datatypes
      CALL sadzi140_xml_gen_table_data_types_xml(m_dzea_t[mi_dzea_index].dzea001,ls_top_env||ls_separator||"com"||ls_separator||"mta"||ls_separator||ms_lang||ls_separator||"datatypes.xml") RETURNING lb_result
      #資料表欄位屬性資料 "datatypes.xml" 產出程序結束 !!
      CALL sadzp000_msg_show_info('adz-00108', 'adz-00108', NULL, 0)
    }

    ON ACTION tm_sql_querizer
      LET ls_program_name = "adzi170"
      LET ls_parameter = "r.r ",ls_program_name
      &ifndef DEBUG
      IF sadzi140_util_check_program_exists(ls_program_name) THEN
        CALL cl_cmdrun_openpipe("r.r",ls_parameter,FALSE) RETURNING lb_return,ls_retmsg
      END IF
      &endif

    ON ACTION tm_affect_prog_list
      LET ls_table_name  = m_dzea_t[mi_dzea_index].dzea001
      CALL sadzi140_util_collect_affect_prog_list(ls_table_name)

    ON ACTION tm_export
      LET ls_table_name  = m_dzea_t[mi_dzea_index].dzea001
      CALL sadzi140_util_export_table_schema(ls_table_name,ms_lang,cs_rt_design_data_code,cs_rt_design_sd_version,cs_rt_design_request_no,cs_rt_design_sequence_no)

    ON ACTION tm_import
      #LET ls_replace_arg = m_dzea_t[mi_dzea_index].dzea001,"|"
      --20180522 mark begin
      --LET ls_replace_arg = " ","|"
      --CALL sadzp000_msg_alert_box("adz-00751", "adz-00751", ls_replace_arg, 0) RETURNING ls_question
      --IF ls_question = cs_question_yes THEN
      --20180522 mark end
        LET lb_error = FALSE
        #LET ls_table_name  = m_dzea_t[mi_dzea_index].dzea001
        --CALL sadzi140_util_import_table_schema(ms_master_user,ms_lang,cs_rt_design_request_no,cs_rt_design_sequence_no) RETURNING lb_result,ls_table_name
        CALL sadzi140_util_import_table_schema(ms_master_user,ms_lang) RETURNING lb_error,ls_table_name
        IF NOT lb_error THEN #20170310 沒有錯誤才執行
          #Grant APS 權限 #20161102
          CALL sadzi140_util_grant_revoke_privileges(ls_table_name) RETURNING ls_message
          CALL adzi140_refresh_and_locate_master(DIALOG,ls_table_name)
          CALL adzi140_set_editable(DIALOG,mi_dzea_index)
          #異動成功要新增Snapshot
          LET ls_user = ms_user
          LET ls_master_user = ms_master_user
          --LET ls_table_name  = m_dzea_t[mi_dzea_index].dzea001
          LET ls_module_name = sadzi140_exim_get_table_module(ls_table_name) --m_dzea_t[mi_dzea_index].dzea003
          LET ls_table_desc  = sadzi140_exim_get_table_desc(ls_table_name) --m_dzea_t[mi_dzea_index].dzea002
          IF NOT lb_error THEN 
            #設定表格資訊
            CALL sadzi140_util_create_curr_snapshot(ls_master_user,ls_table_name,ls_module_name,ls_table_desc,ms_dgenv)

            #20170208 begin
            #產出 table.sch .tbl 檔
            #CALL sadzi140_xml_get_lang_type_list() RETURNING lo_lang_arr
            CALL sadzi140_crud_get_static_lang_list() RETURNING lo_lang_arr

            &ifndef DEBUG
            CALL cl_progress_bar(lo_lang_arr.getLength())
            &endif

            LET lb_gen_tbl = TRUE #20170908
            LET ls_gen_list = "" #20170908
            FOR li_loop = 1 TO lo_lang_arr.getLength()
              LET ls_lang = lo_lang_arr[li_loop]
              &ifndef DEBUG
              CALL cl_progress_ing("Generate table : '"||ls_table_name||"' data (.tbl and .sch). Language : '"||ls_lang||"'")
              &endif
              #產出tables.xml及tbl檔
              CALL sadzi140_xml_gen_table_datas(ls_table_name,ls_module_name,ls_lang,ls_gen_list) RETURNING lb_gen_tbl,ls_gen_list #20170908
              CALL sadzi140_db_gen_table_schema(ls_table_name,ls_module_name,FALSE,ls_lang) RETURNING ls_schema_sql

              #將有異動的表格資訊存到dzey_t中
              CALL adzi140_set_altered_table_info(ls_table_name)

            END FOR
            #產出AlterTableList.xml供Designer讀取
            CALL sadzi140_xml_gen_web_table_datas(ls_lang)
            IF m_dzea_t[mi_dzea_index].dzea004 = "L" THEN
              #產生多語言 4gl 檔
              CALL sadzi140_util_gen_multi_lang(m_dzea_t[mi_dzea_index].dzea001)
            END IF
            #產生延伸檔
            CALL sadzi140_util_gen_extend_inc(ls_table_name) RETURNING lb_result
            #20170208 end
            #20170908 begin
            IF NOT lb_gen_tbl THEN 
              LET ls_msg_no = "adz-01240"
              LET ls_replace_arg = ls_gen_list,"|"
              CALL sadzp000_msg_show_error(ls_msg_no, ls_msg_no, ls_replace_arg, 1)
            END IF
            #20170908 end
          END IF
        END IF #20170310 
        CALL adzi140_refresh_and_locate_master(DIALOG,NVL(ls_table_name,m_dzea_t[mi_dzea_index].dzea001)) --20180522 modify
      --END IF --20180522 mark

      #DISPLAY "Clinet path : ",lo_export_info.CLIENT_PATH
      #DISPLAY "Server Path :",lo_export_info.WORKING_PATH
      #DISPLAY "Tar Name : ",lo_export_info.TAR_NAME

    ON ACTION tm_clone_import
      LET mb_clone_table = FALSE
      LET ls_orig_path = os.Path.pwd()
      
      CALL sadzi140_exim_reset_parameters() 

      #上傳原始壓縮檔
      CALL sadzi140_util_catch_table_pkg_to_server(ls_lang) RETURNING lb_result,lo_getfile_para.*,lo_tar_file_info.*
      
      IF lb_result THEN 
        #解縮原始包 
        CALL sadzi140_util_untar_file(lo_getfile_para.SERVER_FILE_PATH,lo_getfile_para.SERVER_FILE_NAME,TRUE) RETURNING lb_result
        #取得檔案清單
        CALL sadzi140_util_get_file_list(ls_orig_path,lo_getfile_para.SERVER_FILE_PATH) RETURNING lo_directory_list
        CALL os.Path.chdir(ls_orig_path) RETURNING lb_result

        #建立新表格
        INITIALIZE lo_create_table TO NULL #20170911
        CALL adzi140_create_new_table_wizard(DIALOG,lo_create_table.*,ls_source_Table) RETURNING lb_result,lo_create_table.*

        IF lb_result THEN 
        
          #取得新表的基本資訊
          CALL adzi140_get_base_table_info(lo_create_table.*) RETURNING lo_base_table_info.*
          #把壓縮檔內的檔案複製到另外一個工作目錄, 並且更名為新表名稱
          CALL sadzi140_util_rename_clone(lo_tar_file_info.TABLE_NAME,lo_getfile_para.SERVER_FILE_PATH,lo_directory_list,lo_base_table_info.bti_table_name) RETURNING lb_result,ls_new_work_dir,lo_new_dir_list
          #把檔案內容也變更為新表名稱
          CALL sadzi140_util_replace_clone_contents(lo_tar_file_info.TABLE_NAME,ls_new_work_dir,lo_new_dir_list,lo_base_table_info.bti_table_name,ms_table_tail_code) RETURNING lb_result
          #重新壓縮
          CALL sadzi140_exim_get_parameter(cs_exim_import,lo_base_table_info.bti_table_name,lo_tar_file_info.CONSTRUCT_VERSION,lo_tar_file_info.SD_VERSION,lo_tar_file_info.REQUEST_NO,lo_tar_file_info.SEQUENCE_NO) RETURNING lo_import_parameter.*
          CALL sadzi140_exim_tar_file(lo_import_parameter.*) RETURNING lb_result,ls_tar_name
          CALL os.Path.chdir(ls_orig_path) RETURNING lb_result  
          
          IF lb_result THEN
            DISPLAY cs_information_tag,"Start to import and alter table." 

            LET lo_getfile_para.SERVER_FILE_PATH = ls_new_work_dir -- 重新指定 Server 端的工作路徑
            LET lo_getfile_para.SERVER_FILE_NAME = ls_tar_name     -- 重新指定解包檔案
            LET lo_tar_file_info.TABLE_NAME      = lo_base_table_info.bti_table_name -- 重新指定要更新的表格
            
            #將複製後的表的資料匯入及異動
            CALL sadzi140_util_verify_import_and_alter_table(ls_master_user,ms_lang,lo_getfile_para.*,lo_tar_file_info.*) RETURNING lb_result
            CALL os.Path.chdir(ls_orig_path) RETURNING lb_result  
            IF NOT lb_result THEN
              DISPLAY cs_error_tag,"Import and alter table failed !!"
            END IF 
          ELSE   
            DISPLAY cs_warning_tag,"Catch table package from client to server failed or no select any file." 
          END IF
          
        END IF    
      END IF   
      
      CALL sadzi140_exim_reset_parameters()

      #返回原路徑  
      CALL os.Path.chdir(ls_orig_path) RETURNING lb_result  
      LET ls_search = lo_base_table_info.bti_table_name
      LET ms_search = ls_search
      CALL adzi140_refresh_and_locate_master(DIALOG,lo_base_table_info.bti_table_name)

      INITIALIZE lo_create_table TO NULL

    #20171120 begin  
    ON ACTION tm_selection_import
      LET ls_replace_arg = " ","|"
      CALL sadzp000_msg_alert_box("adz-01290", "adz-01290", ls_replace_arg, 0) RETURNING ls_question
      IF ls_question = cs_question_yes THEN
        LET lb_error = FALSE
        CALL sadzi140_util_selection_import_table_schema(ms_master_user,ms_lang) RETURNING lb_error,ls_table_name
        IF NOT lb_error THEN 
          #Grant APS 權限
          CALL sadzi140_util_grant_revoke_privileges(ls_table_name) RETURNING ls_message
          CALL adzi140_refresh_and_locate_master(DIALOG,ls_table_name)
          CALL adzi140_set_editable(DIALOG,mi_dzea_index)
          #異動成功要新增Snapshot
          LET ls_user = ms_user
          LET ls_master_user = ms_master_user
          LET ls_module_name = sadzi140_exim_get_table_module(ls_table_name)
          LET ls_table_desc  = sadzi140_exim_get_table_desc(ls_table_name)
          IF NOT lb_error THEN 
            #設定表格資訊
            CALL sadzi140_util_create_curr_snapshot(ls_master_user,ls_table_name,ls_module_name,ls_table_desc,ms_dgenv)

            #產出 table.sch .tbl 檔
            #CALL sadzi140_xml_get_lang_type_list() RETURNING lo_lang_arr
            CALL sadzi140_crud_get_static_lang_list() RETURNING lo_lang_arr

            &ifndef DEBUG
            CALL cl_progress_bar(lo_lang_arr.getLength())
            &endif

            LET lb_gen_tbl = TRUE
            LET ls_gen_list = ""
            FOR li_loop = 1 TO lo_lang_arr.getLength()
              LET ls_lang = lo_lang_arr[li_loop]
              &ifndef DEBUG
              CALL cl_progress_ing("Generate table : '"||ls_table_name||"' data (.tbl and .sch). Language : '"||ls_lang||"'")
              &endif
              #產出tables.xml及tbl檔
              CALL sadzi140_xml_gen_table_datas(ls_table_name,ls_module_name,ls_lang,ls_gen_list) RETURNING lb_gen_tbl,ls_gen_list
              CALL sadzi140_db_gen_table_schema(ls_table_name,ls_module_name,FALSE,ls_lang) RETURNING ls_schema_sql

              #將有異動的表格資訊存到dzey_t中
              CALL adzi140_set_altered_table_info(ls_table_name)

            END FOR
            #產出AlterTableList.xml供Designer讀取
            CALL sadzi140_xml_gen_web_table_datas(ls_lang)
            IF m_dzea_t[mi_dzea_index].dzea004 = "L" THEN
              #產生多語言 4gl 檔
              CALL sadzi140_util_gen_multi_lang(m_dzea_t[mi_dzea_index].dzea001)
            END IF
            #產生延伸檔
            CALL sadzi140_util_gen_extend_inc(ls_table_name) RETURNING lb_result
            IF NOT lb_gen_tbl THEN 
              LET ls_msg_no = "adz-01240"
              LET ls_replace_arg = ls_gen_list,"|"
              CALL sadzp000_msg_show_error(ls_msg_no, ls_msg_no, ls_replace_arg, 1)
            END IF
          END IF
        END IF
        CALL adzi140_refresh_and_locate_master(DIALOG,ls_table_name)
      END IF
      #20171120 end

    #20180314 begin  
    ON ACTION tm_batch_checkout
      LET ls_program_name = "adzp120"
      LET ls_parameter = "r.r ",ls_program_name
      &ifndef DEBUG
      IF sadzi140_util_check_program_exists(ls_program_name) THEN
        CALL cl_cmdrun_openpipe("r.r",ls_parameter,FALSE) RETURNING lb_return,ls_retmsg
      END IF
      &endif
    #20180314 end  
      
    ON ACTION tm_about
      LET ls_about = adzi140_get_system_information(FALSE)
      #About
      MENU 'About' ATTRIBUTE (STYLE="dialog", COMMENT=ls_about.trim() CLIPPED, IMAGE="information")
        ON ACTION ok
          EXIT MENU
      END MENU

    ON ACTION tm_quit
      EXIT DIALOG
    #end Top Menu Actions

    #begin Toolbar Actions
    ON ACTION tbi_adzi150
      LET ls_program_name = "adzi150"
      LET ls_parameter = "r.r ",ls_program_name," ",m_dzea_t[mi_dzea_index].dzea001," ",ls_module
      &ifndef DEBUG
      IF sadzi140_util_check_program_exists(ls_program_name) THEN
        CALL cl_cmdrun_openpipe("r.r",ls_parameter,FALSE) RETURNING lb_return,ls_retmsg
      END IF
      &endif

    #20161220 begin
    ON ACTION tm_drop_backdoor_column
      IF mi_dzea_index = 0 THEN CONTINUE DIALOG END IF
      LET ls_table_name = m_dzea_t[mi_dzea_index].dzea001
      CALL sadzi140_util_toggle_columns_comments(ls_table_name) RETURNING lb_result
      CALL sadzi140_db_get_backdoor_columns(ls_table_name) RETURNING li_backdoor_col_counts,ls_backdoor_col_list
      IF li_backdoor_col_counts > 0 THEN 
        LET ls_replace_arg = "\n",ls_backdoor_col_list,"|"
        CALL sadzp000_msg_alert_box("adz-00948", "adz-00948", ls_replace_arg, 0) RETURNING ls_question
        IF (ls_question = cs_question_yes) THEN
          LET lb_alter_backdoor = TRUE
          LET ls_drop_backdoor_column = "Y"
          CALL FGL_SETENV(cs_env_RT_DROP_BACKDOOR_COLUMN,ls_drop_backdoor_column)
          LET ls_message = cs_logs_type_set_drop_backdoor_column," : ",cs_env_RT_DROP_BACKDOOR_COLUMN,"=",ls_drop_backdoor_column
          CALL sadzi140_logs_write_log(cs_logs_level_warning,cs_logs_type_set_drop_backdoor_column,ls_message) RETURNING lb_result
          GOTO _tbi_exec_alter
        END IF
      ELSE
        LET ls_replace_arg = ls_table_name,"|"
        CALL sadzp000_msg_show_info('adz-00949', 'adz-00949', ls_replace_arg, 0)
      END IF  
    #20161220 end
    
    ON ACTION tbi_exec_alter
      LABEL _tbi_exec_alter:
      IF m_dzea_t[IIF(mi_dzea_index==0,1,mi_dzea_index)].dzea001 IS NULL THEN 
        CALL sadzp000_msg_show_error("adz-00824", "adz-00824", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF

      #20170425 begin      
      CALL sadzp000_msg_question_box("adz-00115", "adz-00115", ls_replace_arg, 0) RETURNING ls_question
      IF ls_question = cs_question_no THEN
        CONTINUE DIALOG
      END IF 
      #20170425 end      
      
      LET ls_user = ms_user
      LET ls_master_user = ms_master_user
      LET ls_table_name  = m_dzea_t[mi_dzea_index].dzea001
      LET ls_module_name = m_dzea_t[mi_dzea_index].dzea003
      LET ls_table_desc  = m_dzea_t[mi_dzea_index].dzea002
      LET ls_table_type  = m_dzea_t[mi_dzea_index].dzea004

      -- 檢核欄位資料是否存在
      CALL adzi140_check_column_data_exists() RETURNING lb_return
      IF NOT lb_return THEN CONTINUE DIALOG END IF
      #20161220 begin
      IF lb_alter_backdoor THEN
        LET lb_diff = lb_alter_backdoor 
      ELSE
      #20161220 end
        -- 檢核r.t資料是否和資料庫中有差異
        CALL adzi140_check_db_data_diff(ls_master_user,ls_table_name) RETURNING lb_return
        LET lb_diff = lb_return
      END IF #20161220
      IF lb_diff THEN

        {
        #20170523 mark begin
        -- 檢查是否需含共用欄位
        CALL adzi140_check_if_common_column_data_exists(ls_table_name,ls_table_type,ms_lang,lo_t_dzeb_t) RETURNING lb_return
        IF NOT lb_return THEN CONTINUE DIALOG END IF
        -- 檢查是否需含自定義欄位(標準環境及非 topstd 才檢核)
        IF (ms_DGENV = cs_dgenv_standard) AND (ms_user <> cs_topstd_user_name) THEN
          CALL adzi140_check_if_user_define_column_data_exists(ls_table_name,ls_table_type) RETURNING lb_return
          IF NOT lb_return THEN CONTINUE DIALOG END IF
        END IF
        #20170523 mark end
        }
        
        #實際異動
        CALL sadzi140_util_make_alter_table(ls_master_user,ls_table_name,FALSE) RETURNING lb_error,ls_message
        CALL sadzi140_db_update_alter_code(ls_master_user,ls_table_name)
        #異動成功要新增Snapshot
        IF NOT lb_error THEN
          #設定表格資訊
          CALL sadzi140_util_create_curr_snapshot(ls_master_user,ls_table_name,ls_module_name,ls_table_desc,ms_dgenv)

          #產出 table.sch .tbl 檔
          #CALL sadzi140_xml_get_lang_type_list() RETURNING lo_lang_arr
          CALL sadzi140_crud_get_static_lang_list() RETURNING lo_lang_arr

          {
          #20170523 mark begin
          &ifndef DEBUG
          CALL cl_progress_bar(lo_lang_arr.getLength())
          &endif
          #20170523 mark end
          }

          LET lb_gen_tbl = TRUE #20170908
          LET ls_gen_list = "" #20170908
          FOR li_loop = 1 TO lo_lang_arr.getLength()
            LET ls_lang = lo_lang_arr[li_loop]
            {
            #20170523 mark begin
            &ifndef DEBUG
            CALL cl_progress_ing("Generate table : '"||ls_table_name||"' data (.tbl and .sch). Language : '"||ls_lang||"'")
            &endif
            #20170523 mark end
            }
            CALL sadzp000_msg_show_message_bar("Generate table : '"||ls_table_name||"' data (.tbl and .sch). Language : '"||ls_lang||"'",NULL,ls_lang)
            #產出tables.xml及tbl檔
            CALL sadzi140_xml_gen_table_datas(ls_table_name,ls_module_name,ls_lang,ls_gen_list) RETURNING lb_gen_tbl,ls_gen_list #20170908
            CALL sadzi140_db_gen_table_schema(ls_table_name,ls_module_name,FALSE,ls_lang) RETURNING ls_schema_sql

            #將有異動的表格資訊存到dzey_t中
            CALL adzi140_set_altered_table_info(ls_table_name)

          END FOR
          #20170908 begin
          IF NOT lb_gen_tbl THEN 
            LET ls_msg_no = "adz-01240"
            LET ls_replace_arg = ls_gen_list,"|"
            CALL sadzp000_msg_show_error(ls_msg_no, ls_msg_no, ls_replace_arg, 1)
          END IF
          #20170908 end
          #產出AlterTableList.xml供Designer讀取
          CALL sadzi140_xml_gen_web_table_datas(ls_lang)
          #有差異要重產ds.sch
          IF lb_diff THEN
            CALL sadzi140_db_gen_db_schema(ls_master_user.toLowerCase(),ls_table_name,ls_table_type) RETURNING lb_result
            IF NOT lb_result THEN
              CALL sadzp000_msg_show_error('adz-00381', 'adz-00381', '', 1)
            END IF
          END IF
          IF m_dzea_t[mi_dzea_index].dzea004 = "L" THEN
            #產生多語言 4gl 檔
            CALL sadzi140_util_gen_multi_lang(m_dzea_t[mi_dzea_index].dzea001)
          END IF
          #產生延伸檔
          CALL sadzi140_util_gen_extend_inc(ls_table_name) RETURNING lb_result
          #重新編譯失效物件
          CALL sadzi140_util_recompile_invalid_db_objects(ls_table_name) RETURNING lb_result
        END IF

        #20161220 begin
        IF lb_alter_backdoor THEN 
          LET ls_drop_backdoor_column = "N"
          CALL FGL_SETENV(cs_env_RT_DROP_BACKDOOR_COLUMN,ls_drop_backdoor_column)
          LET ls_message = cs_logs_type_set_drop_backdoor_column," : ",cs_env_RT_DROP_BACKDOOR_COLUMN,"=",ls_drop_backdoor_column
          CALL sadzi140_logs_write_log(cs_logs_level_warning,cs_logs_type_set_drop_backdoor_column,ls_message) RETURNING lb_result
          LET lb_alter_backdoor = FALSE
        END IF          
        #20161220 end
      END IF
      #20170706 begin
      #檢核欄位ID有沒有差異
      CALL sadzi140_db_get_column_id_status(ls_master_user,ls_table_name) RETURNING li_columnid_rec 
      LET lb_diff = IIF(li_columnid_rec >= 1,TRUE,FALSE) 
      IF lb_diff THEN
        LET ls_replace_arg = ""
        IF (ms_dgenv = cs_dgenv_customize) THEN 
          CALL sadzp000_msg_show_info("adz-01221", "adz-01221", ls_replace_arg, 1)
        ELSE 
          CALL sadzp000_msg_show_info("adz-01220", "adz-01220", ls_replace_arg, 1)
        END IF
      END IF
      #20170706 end
      LET ls_replace_arg = ls_table_name,"|" #20170523
      CALL sadzp000_msg_show_message_bar("adz-01201",ls_replace_arg,ms_lang) #20170523
      #CALL adzi140_refresh_master(DIALOG,ms_search)
      CALL adzi140_refresh_and_locate_master(DIALOG,ls_table_name)
      
    #20170706 begin  
    ON ACTION tm_table_rebuild
      GOTO _tbi_table_rebuild
    #20170706 end  
    ON ACTION tbi_table_rebuild 
      LABEL _tbi_table_rebuild: #20170706
      IF m_dzea_t[IIF(mi_dzea_index==0,1,mi_dzea_index)].dzea001 IS NULL THEN 
        CALL sadzp000_msg_show_error("adz-00824", "adz-00824", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF
      LET ls_master_user = ms_master_user
      LET ls_table_name  = m_dzea_t[mi_dzea_index].dzea001
      LET ls_module_name = m_dzea_t[mi_dzea_index].dzea003
      LET ls_table_desc  = m_dzea_t[mi_dzea_index].dzea002
      LET ls_table_type  = m_dzea_t[mi_dzea_index].dzea004

      LET lb_table_rebuild = FALSE
      LET ls_replace_arg = "|"
      CALL sadzp000_msg_alert_box("adz-00170", "adz-00170", ls_replace_arg, 0) RETURNING ls_question
      IF ls_question = cs_question_yes THEN
        {
        #異動前檢核欄位和索引與實際表格有沒有差異
        CALL sadzi140_shot_check_diff(ls_master_user,ls_table_name,mb_drop_abnormal_column) RETURNING lb_diff
        }
        #呼叫表格重建函式
        CALL sadzi140_util_table_rebuild(ls_master_user,ls_table_name,FALSE,ms_alm_version) RETURNING lb_table_rebuild,ls_rtn_code,ls_message
        IF (lb_table_rebuild AND ls_rtn_code = "1") THEN
          #begin 異動成功要新增Snapshot
          #設定表格資訊
          CALL sadzi140_util_create_curr_snapshot(ls_master_user,ls_table_name,ls_module_name,ls_table_desc,ms_dgenv)
          #end 異動成功要新增Snapshot
          CALL sadzi140_db_update_alter_code(ls_master_user,ls_table_name)
          #IF lb_diff THEN
            CALL sadzi140_db_gen_db_schema(ls_master_user.toLowerCase(),ls_table_name,ls_table_type) RETURNING lb_result
            IF NOT lb_result THEN
              CALL sadzp000_msg_show_error('adz-00381', 'adz-00381', '', 1)
            END IF
          #END IF
          IF m_dzea_t[mi_dzea_index].dzea004 = "L" THEN
            #產生多語言 4gl 檔
            CALL sadzi140_util_gen_multi_lang(m_dzea_t[mi_dzea_index].dzea001)
          END IF
          #產生延伸檔
          CALL sadzi140_util_gen_extend_inc(ls_table_name) RETURNING lb_result
          #重新編譯失效物件
          CALL sadzi140_util_recompile_invalid_db_objects(ls_table_name) RETURNING lb_result
          #CALL adzi140_refresh_master(DIALOG,ms_search)
          CALL adzi140_refresh_and_locate_master(DIALOG,ls_table_name)
        END IF
        #Grant APS 權限 #20161102
        CALL sadzi140_util_grant_revoke_privileges(ls_table_name) RETURNING ls_message
        #Grant APS 權限, 未來要移除
        CALL adzi140_grant_all_db_APS_privilege()
      END IF

    ON ACTION tm_import_sco
      #20170706 begin
      IF m_dzea_t[IIF(mi_dzea_index==0,1,mi_dzea_index)].dzea001 IS NULL THEN 
        CALL sadzp000_msg_show_error("adz-00824", "adz-00824", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF
      #20170706 end
      LET ls_master_user  = ms_master_user
      LET ls_table_name   = m_dzea_t[mi_dzea_index].dzea001
      LET ls_module_name  = m_dzea_t[mi_dzea_index].dzea003
      LET ls_table_type   = m_dzea_t[mi_dzea_index].dzea004 #20180903
      LET ls_working_path = os.Path.join(os.Path.join(os.Path.join(FGL_GETENV("ERP"),ls_module_name.toLowerCase()),"sa"),"tsc")
      CALL os.Path.chdir(ls_working_path) RETURNING lb_return
      #CALL sadzi140_util_sco_delete_data(ls_master_user,ls_table_name)
      LET ls_program_name = 'adzp160'
      LET ls_parameter = "r.r ",ls_program_name," ",ls_table_name
      &ifndef DEBUG
      IF sadzi140_util_check_program_exists(ls_program_name) THEN
        CALL cl_cmdrun_openpipe("r.r",ls_parameter,FALSE) RETURNING lb_return,ls_retmsg
      END IF
      &endif
      CALL adzi140_refresh_and_locate_master(DIALOG,ls_table_name)
      
      #20180903 begin
      CALL lo_collect_messages.clear()
      CALL adzi140_check_common_fields_if_exists(ls_table_name,ls_table_type,ls_module_name,m_dzeb_t,lo_collect_messages) RETURNING lo_collect_messages
      IF lo_collect_messages.getLength() <> 0 THEN 
        CALL sadzp000_msg_show_collect_message(lo_collect_messages,ms_lang) 
        CALL lo_collect_messages.clear()
      END IF
      #20180903 end
      
      #CALL adzi140_refresh_detail(DIALOG,ls_table_name)
      #設定 Import 功能及建立function table 的功能是否 Enable
      CALL adzi140_set_import_sco_enable(DIALOG)
      CALL adzi140_set_editable(DIALOG,mi_dzea_index)
      LET ls_replace_arg = ls_table_name,"|"
      CALL sadzp000_msg_show_info('adz-00109', 'adz-00109', ls_replace_arg, 0)

    ON ACTION tbi_export_excel
      IF m_dzea_t[IIF(mi_dzea_index==0,1,mi_dzea_index)].dzea001 IS NULL THEN 
        CALL sadzp000_msg_show_error("adz-00824", "adz-00824", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF
      &ifndef DEBUG
        #Begin:2014/08/27 by Hiko
        #建立匯出excel所需要的陣列內容
        CALL g_export_node.clear()
        LET g_export_node[1] = base.typeInfo.create(m_dzea_t)
        LET g_export_node[2] = base.typeInfo.create(m_dzel_t)
        LET g_export_node[3] = base.typeInfo.create(m_dzeb_t)
        LET g_export_node[4] = base.typeInfo.create(m_dzed_t)
        LET g_export_node[5] = base.typeInfo.create(m_dzec_t)
        LET g_export_node[6] = base.typeInfo.create(m_dzen_t) #2016.02.17 by ernest
        TRY
          CALL cl_export_to_excel_getpage()
          CALL cl_export_to_excel()
        CATCH
        END TRY
        #End:2014/08/27 by Hiko
      &endif

    ON ACTION tbi_table_diff
      IF m_dzea_t[IIF(mi_dzea_index==0,1,mi_dzea_index)].dzea001 IS NULL THEN 
        CALL sadzp000_msg_show_error("adz-00824", "adz-00824", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF
      LET ls_master_user = ms_master_user
      LET ls_table_name  = m_dzea_t[mi_dzea_index].dzea001
      CALL sadzi140_diff_run(ls_master_user,ls_table_name)

    ON ACTION tm_rev_list
      GOTO _tbi_rev_list

    ON ACTION tbi_rev_list
      LABEL _tbi_rev_list:
      IF mi_dzea_index = 0 THEN CONTINUE DIALOG END IF
      LET ls_master_user = ms_master_user
      LET ls_table_name  = m_dzea_t[mi_dzea_index].dzea001
      LET lb_active = adzi140_check_if_check_out_active(mi_dzea_index) OR mb_debug
      CALL sadzi140_rev_run(ls_table_name,ls_master_user,ms_lang,lb_active,mb_debug) RETURNING lb_return #20170907
      IF lb_return THEN
        CALL sadzi140_db_update_alter_code(ls_master_user,ls_table_name)
        #CALL adzi140_refresh_master(DIALOG,ms_search)
        CALL adzi140_refresh_and_locate_master(DIALOG,ls_table_name)
      END IF

    {  
    ON ACTION tm_optimizer
      #設定最佳化架構
      CALL sadzi140_db_alter_optimizer_feature()
    }
    
    #ALM Control Begin
    ON ACTION btn_check_out
      #20170706 begin
      IF m_dzea_t[IIF(mi_dzea_index==0,1,mi_dzea_index)].dzea001 IS NULL THEN 
        CALL sadzp000_msg_show_error("adz-00824", "adz-00824", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF
      #20170706 end
      LET ls_replace_arg = m_dzea_t[mi_dzea_index].dzea001,"|"
      CALL sadzp000_msg_question_box("adz-00369", "adz-00369", ls_replace_arg, 0) RETURNING ls_question
      IF NOT IIF(ls_question == cs_question_yes,TRUE,FALSE) THEN
        CONTINUE DIALOG
      END IF
      #20190513 begin
      IF sadzi140_util_check_if_critical_table(m_dzea_t[mi_dzea_index].dzea001) THEN 
        LET ls_replace_arg = m_dzea_t[mi_dzea_index].dzea001,"|"
        CALL sadzp000_msg_question_box("adz-01590", "adz-01590", ls_replace_arg, 0) RETURNING ls_question
        IF NOT IIF(ls_question == cs_question_yes,TRUE,FALSE) THEN
          CONTINUE DIALOG
        END IF
      END IF  
      #20190513 end
      #20201007 begin
      IF (ms_DGENV = cs_dgenv_standard) and (ms_user <> cs_topstd_user_name) AND (ms_topind = cs_indsd_ux) THEN
        LET ls_module_name = m_dzea_t[mi_dzea_index].dzea003
        CALL sadzi140_util_get_if_indsd_module_exists(ls_module_name) RETURNING lb_indsd
        IF NOT lb_indsd THEN 
          LET ls_replace_arg  = ls_module_name,"|"
          CALL sadzp000_msg_show_error("adz-01634", "adz-01634", ls_replace_arg, 1)
          CONTINUE DIALOG
        END IF              
      END IF            
      #20201007 end

      LET lo_create_table.dct_table_name            = m_dzea_t[mi_dzea_index].dzea001
      LET lo_create_table.dct_table_description     = m_dzea_t[mi_dzea_index].dzea002
      LET lo_create_table.dct_module_name           = m_dzea_t[mi_dzea_index].dzea003
      LET lo_create_table.dct_table_type            = m_dzea_t[mi_dzea_index].dzea004
      LET lo_create_table.dct_is_multi_lang_table   = m_dzea_t[mi_dzea_index].dzea005
      LET lo_create_table.dct_is_system_table       = m_dzea_t[mi_dzea_index].dzea006
      LET lo_create_table.dct_create_notice_trigger = m_dzea_t[mi_dzea_index].dzea007 #20190424

      CALL adzi140_set_step_data(FALSE,TRUE,FALSE,FALSE) RETURNING lo_step.*
      #呼叫 ALM 取得 DZLU 資訊
      CALL adzi140_step_call_alm_check_out(mb_enable_alm,ms_user,ms_lang,cs_user_role_sd,lo_step.*,FALSE) RETURNING lb_result,lo_step.*,lo_USER_INFO.*,lo_DZLU_T

      {
      #20190304 mark begin
      #若在標準環境, 且簽出行業別和程式行業別不同的, 不能簽出
      #IF (lb_result) AND (mb_enable_alm) AND (ms_DGENV = cs_dgenv_standard) AND (NVL(lo_DZLU_T[1].DZLU008,cs_null_default) <> NVL(m_dzea_t[mi_dzea_index].dzea014,cs_null_default)) THEN #20160803 marked
      IF (lb_result) AND (mb_enable_alm) AND (ms_DGENV = cs_dgenv_standard) AND (NVL(lo_DZLU_T[1].DZLU008,NVL(ms_topind,cs_default_topind)) <> NVL(m_dzea_t[mi_dzea_index].dzea014,NVL(ms_topind,cs_default_topind))) THEN #20160803 modified
        LET ls_replace_arg = ""
        CALL sadzp000_msg_show_error("adz-00516", "adz-00516", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF
      #20190304 mark end
      }

      #ALM 給定編號
      CALL adzi140_step_get_alm_info(lo_create_table.dct_table_name,lo_create_table.dct_module_name,lo_create_table.dct_table_description,cs_spec_type_table,lo_USER_INFO.ui_ROLE,lo_DZLU_T) RETURNING lb_result, lo_dzlm_t.*
      #重新更新清單及按鈕狀態
      #CALL adzi140_refresh_master(DIALOG,ms_search)
      #CALL DIALOG.setCurrentRow("sr_tablelist",mi_dzea_index)
      CALL adzi140_refresh_and_locate_master(DIALOG,lo_create_table.dct_table_name)
      CALL adzi140_set_check_out_mode(DIALOG,m_dzea_t[mi_dzea_index].*)
      CALL adzi140_set_editable(DIALOG,mi_dzea_index)

    #20160518 begin  
    ON ACTION btn_release
      #取得現今DZLM的資料
      LET lo_table_info.pi_NAME   = m_dzea_t[mi_dzea_index].dzea001
      LET lo_table_info.pi_MODULE = m_dzea_t[mi_dzea_index].dzea003
      LET lo_table_info.pi_DESC   = m_dzea_t[mi_dzea_index].dzea002
      LET lo_table_info.pi_TYPE   = cs_spec_type_table

      CALL sadzp200_alm_get_dzlm(lo_table_info.*,cs_user_role_sd) RETURNING lo_dzlm_t.*
      CALL sadzp200_alm_check_if_could_release(lo_dzlm_t.*,cs_user_role_sd) RETURNING ls_result,ls_res_code,lo_mark_info.*,lo_check_out_owner.*
      CASE
        #曾過正&階段<=QC,不能釋出
        WHEN ls_result = cs_release_no
          LET ls_replace_arg  = "|"
          CALL sadzp000_msg_show_error(ls_res_code, ls_res_code, ls_replace_arg, 1)
          CONTINUE DIALOG 
        #曾過正(dzlm022='P')&階段>QC,或者已經 remark掉修改段就可以釋出   
        WHEN ls_result = cs_release_yes
          LET ls_err_msg = NULL
          IF lo_check_out_owner.cool_ID IS NOT NULL THEN 
            LET ls_err_msg  = lo_check_out_owner.cool_ROLE,"|",lo_check_out_owner.cool_ID,"|"
          END IF 
         CALL sadzp000_msg_question_box(ls_res_code,ls_res_code,ls_err_msg,0) RETURNING ls_result
         IF ls_result <> cs_response_yes THEN
           CONTINUE DIALOG 
         ELSE   
           GOTO _CHECK_IN
         END IF
      END CASE 
      
    #20160518 end 
      
    ON ACTION btn_check_in
      LET ls_replace_arg = m_dzea_t[mi_dzea_index].dzea001,"|"
      CALL sadzp000_msg_question_box("adz-01442", "adz-01442", ls_replace_arg, 0) RETURNING ls_question
      IF NOT IIF(ls_question == cs_question_yes,TRUE,FALSE) THEN
        CONTINUE DIALOG
      END IF

      #20160518
      LABEL _CHECK_IN: 
      
      #檢核欄位和索引與實際表格有沒有差異
      CALL sadzi140_shot_check_diff(ls_master_user,ls_table_name,mb_drop_abnormal_column) RETURNING lb_diff
      IF lb_diff THEN
        CALL sadzp000_msg_show_error("adz-00135", "adz-00135", ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF
      #20170706 begin
      #檢核欄位ID有沒有差異
      CALL sadzi140_db_get_column_id_status(ls_master_user,ls_table_name) RETURNING li_columnid_rec 
      LET lb_diff = IIF(li_columnid_rec >= 1,TRUE,lb_diff) 
      IF lb_diff THEN
        LET ls_replace_arg = ""
        IF (ms_dgenv = cs_dgenv_customize) THEN 
          CALL sadzp000_msg_show_info("adz-01221", "adz-01221", ls_replace_arg, 1)
        ELSE 
          CALL sadzp000_msg_show_info("adz-01220", "adz-01220", ls_replace_arg, 1)
        END IF
        CONTINUE DIALOG
      END IF
      #20170706 end

      #取得現今DZLM的資料
      LET lo_table_info.pi_NAME   = m_dzea_t[mi_dzea_index].dzea001
      LET lo_table_info.pi_MODULE = m_dzea_t[mi_dzea_index].dzea003
      LET lo_table_info.pi_DESC   = m_dzea_t[mi_dzea_index].dzea002
      LET lo_table_info.pi_TYPE   = cs_spec_type_table

      -------------------------------------------------------------------
      -------------------------- ALM 相關資訊 ----------------------------

      CALL sadzp200_alm_get_dzlm(lo_table_info.*,cs_user_role_sd) RETURNING lo_dzlm_t.*

      BEGIN WORK

      CALL sadzp200_alm_update_check_in_code(lo_DZLM_T.*,cs_user_role_sd) RETURNING lb_result
      IF NOT lb_result THEN GOTO _CheckWork END IF

      #檢核DZLM資料是否還有簽出的
      CALL sadzp200_alm_check_data_valid(lo_DZLM_T.*,cs_check_out) RETURNING lb_result
      IF NOT lb_result THEN
        #當DZLM中的SD或PR項均無簽出者,則 DZLM 資訊更新遷入日期結案
        CALL sadzp200_alm_update_check_in_date(lo_DZLM_T.*) RETURNING lb_result
        IF NOT lb_result THEN GOTO _CheckWork END IF
      END IF

      LABEL _CheckWork:

      IF lb_result THEN
        COMMIT WORK
      ELSE
        ROLLBACK WORK
      END IF

      IF NOT lb_result THEN
        LET ls_replace_arg = m_dzea_t[mi_dzea_index].dzea001,"|"
        CALL sadzp000_msg_show_error('adz-00326', 'adz-00326', ls_replace_arg, 1)
      END IF

      -------------------------------------------------------------------
      -------------------------------------------------------------------

      #重新更新清單及按鈕狀態
      #CALL adzi140_refresh_master(DIALOG,ms_search)
      #CALL DIALOG.setCurrentRow("sr_tablelist",mi_dzea_index)
      CALL adzi140_refresh_and_locate_master(DIALOG,m_dzea_t[mi_dzea_index].dzea001)
      CALL adzi140_set_check_out_mode(DIALOG,m_dzea_t[mi_dzea_index].*)
      CALL adzi140_set_editable(DIALOG,mi_dzea_index)

    ON ACTION btn_recall
      LET ls_replace_arg = m_dzea_t[mi_dzea_index].dzea001,"|"
      CALL sadzp000_msg_question_box("adz-00370", "adz-00370", ls_replace_arg, 0) RETURNING ls_question
      IF NOT IIF(ls_question == cs_question_yes,TRUE,FALSE) THEN
        CONTINUE DIALOG
      END IF

      LET lo_table_info.pi_NAME   = m_dzea_t[mi_dzea_index].dzea001
      LET lo_table_info.pi_MODULE = m_dzea_t[mi_dzea_index].dzea003
      LET lo_table_info.pi_DESC   = m_dzea_t[mi_dzea_index].dzea002
      LET lo_table_info.pi_TYPE   = cs_spec_type_table

      #取得該表格最大的異動版次
      LET lb_success = TRUE
      CALL sadzi140_rev_get_revision_data(lo_table_info.pi_NAME) RETURNING lo_REVISION.*
      IF lo_REVISION.rv_ALM_VERSION IS NOT NULL THEN
        LET lb_success = FALSE
        CALL sadzi140_rev_make_revision(lo_table_info.pi_NAME,lo_REVISION.*) RETURNING lb_success
      END IF

      IF lb_success THEN

        -------------------------------------------------------------------------
        --------------------------- begin ALM 相關資訊 ----------------------------

        CALL sadzp200_alm_get_dzlm(lo_table_info.*,cs_user_role_sd) RETURNING lo_dzlm_t.*

        BEGIN WORK

        LET lb_commit = TRUE
        #依據要取消的ROLE將相關資料清空並重新回傳
        CALL sadzp200_alm_preset_dzlm(lo_DZLM_T.*,cs_user_role_sd) RETURNING lo_DZLM_T.*
        #依據角色清空對應的版次欄位
        CALL sadzp200_alm_set_dzlm(lo_DZLM_T.*,cs_user_role_sd) RETURNING lb_result,ls_update_type
        IF NOT lb_result THEN LET lb_commit = FALSE GOTO _CheckRecall END IF
        #檢核DZLM資料是否還有效(還存在SD或PR的資訊)
        IF NOT sadzp200_alm_check_data_valid(lo_DZLM_T.*,cs_check_out) AND NOT sadzp200_alm_check_data_valid(lo_DZLM_T.*,cs_check_in) THEN
          #當DZLM中的SD或PR項的板次都不存在時, 表示該簽出均已取消, DZLM及DZAF的資訊均可刪除
          CALL sadzp200_alm_delete_dzlm_data(lo_DZLM_T.*) RETURNING lb_result
          IF NOT lb_result THEN LET lb_commit = FALSE GOTO _CheckRecall END IF
          CALL sadzp200_ver_delete_data(lo_DZLM_T.*,ms_DGENV) RETURNING lb_result
          IF NOT lb_result THEN LET lb_commit = FALSE GOTO _CheckRecall END IF
        ELSE
          #更新版次資訊(DZAF_T)
          LET lo_DZAF_T.DZAF001 = lo_table_info.pi_NAME
          LET lo_DZAF_T.DZAF002 = lo_DZLM_T.DZLM005
          LET lo_DZAF_T.DZAF005 = lo_table_info.pi_TYPE
          LET lo_DZAF_T.DZAF006 = lo_table_info.pi_MODULE

          #先取得現行的DZAF版次相關資料
          CALL sadzp200_ver_get_curr_ver_info(lo_DZAF_T.*) RETURNING lo_DZAF_T.*
          #遞減對應的版次號碼
          CALL sadzp200_ver_decreas_dzaf_ver(lo_DZAF_T.*,cs_user_role_sd) RETURNING lo_DZAF_T.*
          #更新資料庫
          CALL sadzp200_ver_set_ver(lo_DZAF_T.*,cs_user_role_sd) RETURNING lb_result
          IF NOT lb_result THEN LET lb_commit = FALSE GOTO _CheckRecall END IF
          #檢核DZLM資料是否還有簽出的
          CALL sadzp200_alm_check_data_valid(lo_DZLM_T.*,cs_check_out) RETURNING lb_result
          IF NOT lb_result THEN
            #當DZLM中的SD或PR項均無簽出者,則 DZLM 資訊更新遷入日期結案
            CALL sadzp200_alm_update_check_in_date(lo_DZLM_T.*) RETURNING lb_result
            IF NOT lb_result THEN LET lb_commit = FALSE GOTO _CheckRecall END IF
          END IF
        END IF

        LABEL _CheckRecall:

        IF lb_commit THEN
          COMMIT WORK
        ELSE
          ROLLBACK WORK
        END IF

        IF NOT lb_commit THEN
          LET ls_replace_arg = m_dzea_t[mi_dzea_index].dzea001,"|"
          CALL sadzp000_msg_show_error('adz-00297', 'adz-00297', ls_replace_arg, 1)
        END IF

        ---------------------------- end ALM 相關資訊 ----------------------------
        -------------------------------------------------------------------------
        LET ls_replace_arg = "|"
        CALL sadzp000_msg_show_info('adz-00189', 'adz-00189', ls_replace_arg, 0)
      ELSE
        LET ls_replace_arg = "|"
        CALL sadzp000_msg_show_error('adz-00190', 'adz-00190', ls_replace_arg, 0)
      END IF

      #重新更新清單及按鈕狀態
      CALL sadzi140_db_update_alter_code(ms_master_user,m_dzea_t[mi_dzea_index].dzea001)
      #CALL adzi140_refresh_master(DIALOG,ms_search)
      #CALL DIALOG.setCurrentRow("sr_tablelist",mi_dzea_index)
      CALL adzi140_refresh_and_locate_master(DIALOG,m_dzea_t[mi_dzea_index].dzea001)
      CALL adzi140_set_check_out_mode(DIALOG,m_dzea_t[mi_dzea_index].*)
      CALL adzi140_set_editable(DIALOG,mi_dzea_index)

    #ALM Control End

    #20170912 begin
    ON ACTION controlg
      LET ls_program_name = 'adzp009'
      IF sadzi140_util_check_program_exists(ls_program_name) THEN
        LET ls_command = "r.r ",ls_program_name
        RUN ls_command WITHOUT WAITING
      END IF
    #20170912 end  

    ON ACTION EXIT
      LET INT_FLAG=FALSE
      EXIT DIALOG

    ON ACTION CLOSE
      LET INT_FLAG=FALSE
      EXIT DIALOG

    CONTINUE DIALOG

  END DIALOG

END FUNCTION

FUNCTION adzi140_finalize(p_value)
DEFINE
  p_value BOOLEAN

  IF p_value THEN
    EXIT PROGRAM
  ELSE
    EXIT PROGRAM -1
  END IF

END FUNCTION

FUNCTION adzi140_create_new_table_wizard(p_dialog,p_create_table,p_source_Table)
DEFINE 
  p_dialog        ui.Dialog,
  p_create_table  T_DZEA_CREATE_TABLE,
  p_source_Table  STRING
DEFINE
  lo_create_table      T_DZEA_CREATE_TABLE,
  ls_source_Table      STRING,
  lb_result            BOOLEAN,
  li_step              INTEGER,
  ls_all_message       STRING,
  ls_master_user       STRING,
  ls_industry_type     STRING,
  ls_table_name        STRING,
  ls_alm_version       STRING,
  ls_alm_request_no    STRING,
  ls_dgenv             STRING,
  ls_dest_Table        STRING,
  ls_version           STRING,
  ls_replace_arg       STRING,
  lb_enable_next       BOOLEAN,
  lb_enable_confirm    BOOLEAN,
  lb_success           BOOLEAN,
  lb_exist_foreign_key BOOLEAN,
  lb_finished          BOOLEAN,
  lo_dzlm_t            T_DZLM_T,
  lo_dzlm_info         T_DZLM_INFO,
  lo_step              T_STEP,
  lo_user_info         T_USER_INFO, -- 使用者資訊物件
  lo_DZLU_T            DYNAMIC ARRAY OF T_DZLU_T,
  lo_schema_type       DYNAMIC ARRAY OF T_TABLE_SYNONYM,
  li_loop              INTEGER
DEFINE
  lb_return  BOOLEAN,
  lo_return  T_DZEA_CREATE_TABLE

  LET lo_create_table.* = p_create_table.*
  LET ls_source_Table = p_source_Table
  
  LET lb_result = TRUE
  LET lb_finished = FALSE
  LET li_step = 0
  LET ls_all_message = ""
  LET ls_master_user = ms_master_user
  LET ls_industry_type = cs_industry_type_standard
  --LET ls_table_name = m_dzea_t[mi_dzea_index].dzea001
  INITIALIZE lo_dzlm_info TO NULL
  
  #begin 建立新表格步驟精靈
  WHILE TRUE

    #呼叫 ALM 取得 DZLU 資訊
    IF (lb_result) AND (li_step = 0) THEN
      CALL adzi140_set_step_data(FALSE,TRUE,FALSE,FALSE) RETURNING lo_step.*
      CALL adzi140_step_call_alm_check_out(mb_enable_alm,ms_user,ms_lang,cs_user_role_sd,lo_step.*,TRUE) RETURNING lb_result,lo_step.*,lo_USER_INFO.*,lo_DZLU_T
      LET ls_industry_type = ms_topind --lo_DZLU_T[1].DZLU008 -- 設定行業別,因為只有SD所以只要抓第一筆就可以
      LET li_step = li_step + 1
    END IF

    #呼叫產生新表格的畫面
    IF (lb_result) AND (li_step = 1) THEN

      LET lb_enable_next = IIF((ms_synonym_define = "Y") OR (ms_enable_industry = "Y"),TRUE,FALSE)
      LET lb_enable_confirm = TRUE --IIF((NVL(ls_industry_type,cs_industry_type_standard)==cs_industry_type_standard) ,TRUE,FALSE)
      
      CALL adzi140_step_create_new_table(lo_create_table.*,ls_source_Table,lb_enable_next,lb_enable_confirm) RETURNING lb_result,lo_step.*,lo_create_table.*

      IF ((lo_step.NEXT) OR (lo_step.CONFIRM)) AND NOT (lo_step.PREV) THEN
        LET li_step = li_step + 1
      ELSE
        LET li_step = li_step - 1
      END IF

    END IF

    #設定行業別
    IF (lb_result) AND (li_step = 2) THEN
      IF (lo_step.NEXT) THEN
        IF (ms_enable_industry = "Y") THEN
          LET lb_enable_next = IIF((ms_synonym_define = "Y"),TRUE,FALSE)
          CALL adzi140_step_set_industry_type(lo_create_table.*,lb_enable_next) RETURNING lb_result, lo_step.*, lo_create_table.*
        END IF
      END IF

      IF ((lo_step.NEXT) OR (lo_step.CONFIRM)) AND NOT (lo_step.PREV) THEN
        LET li_step = li_step + 1
      ELSE
        LET li_step = li_step - 1
      END IF

    END IF

    #呼叫設定 Schema 為 Table 或是 Synonym
    IF (lb_result) AND (li_step = 3) THEN

      IF (ms_synonym_define = "Y") AND (lo_step.NEXT) THEN
        LET lb_enable_next = FALSE
        CALL adzi140_step_set_table_synonym(p_dialog,lo_create_table.*,lb_enable_next) RETURNING lb_result, lo_step.*, lo_schema_type
      ELSE
        CALL adzi140_step_set_table_synonym_no_window(p_dialog,lo_create_table.*) RETURNING lb_result, lo_step.*, lo_schema_type
      END IF

      IF ((lo_step.NEXT) OR (lo_step.CONFIRM)) AND NOT (lo_step.PREV) THEN
        LET li_step = li_step + 1
      ELSE
        LET li_step = li_step - 1
      END IF

    END IF

    #開始執行實際建同意字及表格作業
    IF (lb_result) AND (li_step = 4) THEN

      IF ((lo_step.NEXT) OR (lo_step.CONFIRM)) AND NOT (lo_step.PREV) THEN
        #合併表格名稱及表尾
        LET lo_create_table.dct_table_name = lo_create_table.dct_table_name,lo_create_table.dct_tail_code
        LET ls_table_name = lo_create_table.dct_table_name

        #ALM 給定編號
        CALL adzi140_step_get_alm_info(lo_create_table.dct_table_name,lo_create_table.dct_module_name,lo_create_table.dct_table_description,cs_spec_type_table,lo_USER_INFO.ui_ROLE,lo_DZLU_T) RETURNING lb_result, lo_dzlm_t.*
        IF NOT lb_result THEN GOTO _RESULT END IF

        #給定ALM資訊
        LET ls_all_message    = ""
        LET ls_alm_version    = lo_dzlm_t.DZLM005
        LET ls_alm_request_no = lo_dzlm_t.DZLM012
        LET ls_dgenv          = ms_dgenv

        LET lo_create_table.dct_alm_construct_version = lo_dzlm_t.DZLM005
        LET lo_create_table.dct_alm_sd_version        = lo_dzlm_t.DZLM006
        LET lo_create_table.dct_alm_request_no        = lo_dzlm_t.DZLM012
        LET lo_create_table.dct_dgenv                 = ms_dgenv
        LET lo_create_table.dct_shipping_notice       = "N"
        LET lo_create_table.dct_orig_dgenv            = ms_dgenv

        CALL sadzp000_msg_show_message_bar(cs_message_tag||"Create design data ...",NULL,ms_lang) #20170523
        CALL adzi140_create_dzea_data(p_dialog,lo_create_table.*) RETURNING lb_result
        IF lb_result THEN
          CALL sadzp000_msg_show_message_bar(cs_message_tag||"Create table or synonym data ...",NULL,ms_lang) #20170523
          CALL adzi140_create_table_synonym_data(lo_schema_type) RETURNING lb_result

          CALL sadzp000_msg_show_message_bar(cs_message_tag||"Create real table, please wait ...",NULL,ms_lang) #20170523
          CALL sadzi140_util_create_real_table(ls_master_user,lo_create_table.*,ls_version,ls_alm_version,ls_alm_request_no,ls_dgenv,TRUE) RETURNING ls_all_message
          #CALL sadzi140_db_update_alter_code(ls_master_user,lo_create_table.dct_table_name)#20170523 mark
          CALL sadzi140_db_set_alter_code(ls_master_user,lo_create_table.dct_table_name,"N") #20170523 add

          #執行表格複製
          IF mb_clone_table THEN
            LET ls_dest_Table = lo_create_table.dct_table_name
            CALL sadzi140_clone_start(ls_source_Table,ls_dest_Table) RETURNING lb_success,lb_exist_foreign_key
            #判斷複製的表是否有Foreign Key, 有則提醒
            IF lb_exist_foreign_key THEN
              LET ls_replace_arg = ls_dest_Table,"|"
              CALL sadzp000_msg_show_info('adz-00237', 'adz-00237', ls_replace_arg, 0)
            END IF
          END IF
          
          #Grant APS 權限, 未來要移除
          CALL adzi140_grant_all_db_APS_privilege()
          CALL adzi140_refresh_master(p_dialog,ms_search)
          CALL adzi140_set_import_sco_enable(p_dialog)
          CALL sadzp000_msg_show_message_bar(NULL,NULL,ms_lang) #20170523
        END IF
        
        LET lb_finished = TRUE

        LET li_step = 0

      ELSE
        LET li_step = li_step - 1
      END IF
    END IF

    LABEL _RESULT:

    IF lb_finished OR NOT lb_result THEN
      --LET lo_create_table.dct_module_name = NULL
      --LET lo_create_table.dct_table_name = NULL
      --LET lo_create_table.dct_table_type = NULL
      --LET lo_create_table.dct_table_description = NULL
      EXIT WHILE
    END IF

  END WHILE

  LET lb_return = lb_result
  LET lo_return.* = lo_create_table.*

  RETURN lb_return,lo_return.*

END FUNCTION 

#開窗新增Table
FUNCTION adzi140_step_create_new_table(p_create_table,p_clone_table,p_enable_next,p_enable_confirm)
DEFINE
  p_create_table   T_DZEA_CREATE_TABLE,
  p_clone_table    STRING,
  p_enable_next    BOOLEAN,
  p_enable_confirm BOOLEAN
DEFINE
  lo_create_table          T_DZEA_CREATE_TABLE,
  lo_dzhh_t                T_DZHH_T, #20160617
  ls_clone_table           STRING,
  lb_enable_next           BOOLEAN,
  lb_enable_confirm        BOOLEAN,
  lo_module_combobox       ui.ComboBox,
  lo_step                  T_STEP,
  lo_table_type_combobox   ui.ComboBox,
  lo_window                ui.window,
  lo_form                  ui.form,
  lb_result                BOOLEAN,
  ls_module_combobox       VARCHAR(50),
  ls_table_type_combobox   VARCHAR(50),
  ls_define_group          VARCHAR(100),
  ls_master_db             STRING,
  ls_master_user           STRING,
  ls_table_name            STRING,
  ls_table_leading_code    STRING,
  ls_module_name           STRING,
  ls_module_leading_code   STRING,
  ls_leading_code          STRING,
  ls_table_char            STRING,
  li_loop                  INTEGER,
  lb_table_exist           BOOLEAN,
  ls_window_desc           STRING,
  ls_table_type            STRING,
  ls_replace_arg           STRING,
  ls_tail_code             STRING,
  ls_separator             STRING,
  ls_industry_msg          STRING,
  ls_table_tail            STRING,
  ls_orig_table_name       STRING,
  ls_message               STRING,
  li_max_table_name_length INTEGER,
  li_table_length          INTEGER

  LET lo_create_table.* = p_create_table.*
  LET ls_clone_table    = p_clone_table
  LET lb_enable_next    = p_enable_next
  LET lb_enable_confirm = p_enable_confirm

  LET ls_separator = os.Path.separator()

  LET lb_result = TRUE

  &ifndef DEBUG
  OPEN WINDOW w_sadzi140_wcrt WITH FORM cl_ap_formpath("ADZ","sadzi140_wcrt")
  &else
  OPEN WINDOW w_sadzi140_wcrt WITH FORM sadzi140_util_get_form_path("sadzi140_wcrt")
  &endif

  CURRENT WINDOW IS w_sadzi140_wcrt

  CALL sadzi140_util_set_form_title("sadzi140_wcrt",ms_lang) #20150417 by Hiko

  #begin of 設定 Wizard Title 及圖示
  LET lo_window = ui.Window.getCurrent()
  LET lo_form = lo_window.getForm()

  IF mb_clone_table THEN
    CALL sadzp000_msg_get_message('adz-00220',ms_lang) RETURNING ls_window_desc
    CALL FGL_SETTITLE(ls_window_desc)
    CALL lo_form.setElementImage('img_wcrt','tabledesigner'||ls_separator||'tbd_clone_table.png')
    CALL lo_form.setElementHidden('lbl_clone_table',FALSE)
    CALL lo_form.setFieldHidden('formonly.edt_clone_table',FALSE)
    DISPLAY ls_clone_table TO formonly.edt_clone_table
  ELSE
    CALL sadzp000_msg_get_message('adz-00219',ms_lang) RETURNING ls_window_desc
    CALL FGL_SETTITLE(ls_window_desc)
    CALL lo_form.setElementImage('img_wcrt','tabledesigner'||ls_separator||'tbd_new_schema.png')
    CALL lo_form.setElementHidden('lbl_clone_table',TRUE)
    CALL lo_form.setFieldHidden('formonly.edt_clone_table',TRUE)
  END IF

  #end of 設定 Wizard Title 及圖示

  #填入Table型態
  LET lo_table_type_combobox = ui.ComboBox.forName("formonly.cb_tabletype")
  CALL sadzi140_util_fill_combobox_by_sql(lo_table_type_combobox,ms_sql_table_type)

  #填入模組別資料
  LET mb_tool_table = FALSE
  CALL adzi140_initial_combobox_sql()
  CALL adzi140_fill_module_combobox(lo_module_combobox) RETURNING lo_module_combobox
  LET ls_module_combobox = lo_module_combobox.getItemName(1)
  LET lo_create_table.dct_module_name = NVL(NVL(lo_create_table.dct_module_name,ms_module),ls_module_combobox)

  #預設給標準行業別
  LET lo_create_table.dct_industry_type = NVL(ms_topind,cs_default_topind) --cs_industry_type_standard

  #設定表格結尾字
  LET ls_table_tail = ms_table_tail_combine_code
  LET lo_create_table.dct_tail_code = ms_table_tail_combine_code

  #設定表格型態
  LET ls_table_type_combobox = "M" --lo_table_type_combobox.getItemName(1)
  LET lo_create_table.dct_table_type = NVL(lo_create_table.dct_table_type,ls_table_type_combobox)

  #設定 Checkbox
  LET lo_create_table.dct_is_multi_lang_table = "N"
  LET lo_create_table.dct_is_system_table = "N"
  LET lo_create_table.dct_is_altered = "Y"

  IF NVL(p_create_table.dct_table_name,cs_null_value) <> cs_null_value  THEN
    LET lo_create_table.* = p_create_table.*
    DISPLAY lo_create_table.* TO sr_CreateTable.*
  END IF

  #取得Master DB 及 User
  LET ls_master_db   = ms_master_db
  LET ls_master_user = ms_master_user

  DIALOG
    INPUT lo_create_table.* FROM sr_CreateTable.* ATTRIBUTE(WITHOUT DEFAULTS)
      BEFORE INPUT
        CALL DIALOG.setFieldActive("cb_modulename",TRUE)
        CALL DIALOG.setFieldActive("ed_tablename",TRUE)
        CALL DIALOG.setFieldActive("ed_tabledescription",TRUE)

      ON CHANGE cb_TableType
        #判斷是否為多語言表格
        LET ls_table_type = lo_create_table.dct_table_type
        CALL adzi140_check_if_multi_lang_table(ls_table_type) RETURNING lo_create_table.dct_is_multi_lang_table
        CALL adzi140_alarm_if_multi_lang_table(lo_create_table.dct_is_multi_lang_table)
        CALL adzi140_check_if_tool_table(ls_table_type) RETURNING mb_tool_table
        #重填Module別
        CALL adzi140_initial_combobox_sql()
        CALL adzi140_fill_module_combobox(lo_module_combobox) RETURNING lo_module_combobox
        #觸發Module名稱更換
        GOTO _cb_change_module_name

      ON CHANGE cb_ModuleName
        LABEL _cb_change_module_name:
        CALL sadzi140_db_get_module_leading_code(lo_create_table.dct_module_name,lo_create_table.dct_table_name) RETURNING ls_leading_code
        LET lo_create_table.dct_table_name = ls_leading_code
        LET ls_module_name = lo_create_table.dct_module_name
        CALL adzi140_check_if_system_table(ls_module_name) RETURNING lo_create_table.dct_is_system_table
        CALL adzi140_get_fitting_table_type(ls_module_name,lo_create_table.dct_table_type) RETURNING lo_create_table.dct_table_type

      --BEFORE FIELD ed_TableName
      --  LET lo_create_table.dct_table_name = ls_orig_table_name

      ON CHANGE ed_TableName
        --LET ls_orig_table_name = lo_create_table.dct_table_name
        IF NVL(lo_create_table.dct_table_name,cs_null_value) <> cs_null_value THEN
          CALL sadzi140_db_get_module_leading_code(lo_create_table.dct_module_name,lo_create_table.dct_table_name) RETURNING ls_leading_code

          LET ls_table_name = lo_create_table.dct_table_name
          LET ls_table_leading_code = ls_table_name.subString(1,ls_leading_code.getLength())
          LET ls_module_name = lo_create_table.dct_module_name
          LET ls_module_leading_code = ls_module_name.subString(1,1)

          #判斷模組別的第一碼是不是客制模組碼"E", 若不是則要作一般表格檢核
          IF ls_module_leading_code <> cs_custom_module_leading_code THEN
            #檢核系統定義的前綴字和輸入的前綴字是否相符
            IF NVL(ls_table_leading_code,cs_null_value) <> ls_leading_code THEN
              CALL sadzp000_msg_show_error("adz-00117", "adz-00117", NULL, 1)
              LET lo_create_table.dct_table_name = ls_leading_code
              NEXT FIELD ed_TableName
              CALL DIALOG.setActionActive("btn_next", FALSE)
              CALL DIALOG.setActionActive("btn_ok", FALSE)
            ELSE
              CALL DIALOG.setActionActive("btn_next", lb_enable_next)
              CALL DIALOG.setActionActive("btn_ok", lb_enable_confirm)
            END IF

            #判斷表格主名稱長度
            LET li_max_table_name_length = NVL(ms_max_table_name_length,cs_max_table_name_length)
            IF ls_table_name.getLength() > li_max_table_name_length THEN
              LET ls_message = ms_max_table_name_length,"|"
              CALL sadzp000_msg_show_error("adz-00562", "adz-00562", ls_message, 1)
              LET lo_create_table.dct_table_name = ls_leading_code
              NEXT FIELD ed_TableName
              CALL DIALOG.setActionActive("btn_next", FALSE)
              CALL DIALOG.setActionActive("btn_ok", FALSE)
            ELSE
              CALL DIALOG.setActionActive("btn_next", lb_enable_next)
              CALL DIALOG.setActionActive("btn_ok", lb_enable_confirm)
            END IF

            --20160617 begin
            IF ls_table_name.getLength() < li_max_table_name_length THEN
              LET ls_message = ms_max_table_name_length,"|"
              CALL sadzp000_msg_show_error("adz-00878", "adz-00878", ls_message, 1)
              LET lo_create_table.dct_table_name = ls_leading_code
              NEXT FIELD ed_TableName
              CALL DIALOG.setActionActive("btn_next", FALSE)
              CALL DIALOG.setActionActive("btn_ok", FALSE)
            ELSE
              CALL DIALOG.setActionActive("btn_next", lb_enable_next)
              CALL DIALOG.setActionActive("btn_ok", lb_enable_confirm)
            END IF
            --20160617 end
            
          END IF

          #判斷輸入表格名稱是否含有其他 Unicode
          FOR li_loop = 1 TO ls_table_name.getLength()
            LET ls_table_char = NVL(ls_table_name.subString(li_loop,li_loop),"@")
            --IF ls_table_char NOT MATCHES "[a-z]" AND ls_table_char NOT MATCHES "[_]" THEN
            IF ls_table_char NOT MATCHES "[a-z]" THEN
              CALL sadzp000_msg_show_error("adz-00118", "adz-00118", NULL, 1)
              LET lo_create_table.dct_table_name = ls_leading_code
              NEXT FIELD ed_TableName
              CALL DIALOG.setActionActive("btn_next", FALSE)
              CALL DIALOG.setActionActive("btn_ok", FALSE)
              EXIT FOR
            ELSE
              CALL DIALOG.setActionActive("btn_next", lb_enable_next)
              CALL DIALOG.setActionActive("btn_ok", lb_enable_confirm)
            END IF
          END FOR

          #判斷表格尾字串是否為Combine tail code
          IF ms_check_table_tail_code = "Y" THEN
            LET li_table_length = ls_table_name.getLength()
            LET ls_tail_code = ls_table_name.subString(li_table_length-1,li_table_length)
            IF ls_tail_code <> ms_table_tail_code THEN
              CALL sadzp000_msg_show_error("adz-00250", "adz-00250", NULL, 1)
              NEXT FIELD ed_TableName
              CALL DIALOG.setActionActive("btn_next", FALSE)
              CALL DIALOG.setActionActive("btn_ok", FALSE)
            ELSE
              CALL DIALOG.setActionActive("btn_next", lb_enable_next)
              CALL DIALOG.setActionActive("btn_ok", lb_enable_confirm)
            END IF
          END IF

          {
          #顯示行業別下一步的訊息
          IF (ms_enable_industry = "Y") THEN
            CALL sadzp000_msg_get_message('adz-00274',ms_lang) RETURNING ls_industry_msg
            DISPLAY ls_industry_msg TO formonly.lbl_memo
          END IF
          }
         
        ELSE
          DISPLAY "" TO formonly.lbl_memo
        END IF
    END INPUT

    BEFORE DIALOG
      #啟動 ALM 時設定上一步
      IF mb_enable_alm THEN
        CALL DIALOG.setActionActive("btn_back", TRUE)
      ELSE
        CALL DIALOG.setActionActive("btn_back", FALSE)
      END IF
      #啟動表格複製時
      IF mb_clone_table THEN
        CALL DIALOG.setFieldActive("cb_tabletype",FALSE)
      ELSE
        CALL DIALOG.setFieldActive("cb_tabletype",TRUE)
      END IF
      IF NVL(lo_create_table.dct_table_name,cs_null_value) = cs_null_value THEN
        CALL DIALOG.setActionActive("btn_ok", FALSE)
        CALL DIALOG.setActionActive("btn_next", FALSE)
      ELSE
        CALL DIALOG.setActionActive("btn_ok", lb_enable_confirm)
        CALL DIALOG.setActionActive("btn_next", lb_enable_next)
      END IF

    ON ACTION btn_CANCEL
      INITIALIZE lo_create_table TO NULL
      LET lb_result = FALSE
      CALL adzi140_set_step_data(FALSE,FALSE,FALSE,TRUE) RETURNING lo_step.*
      EXIT DIALOG

    ON ACTION btn_back
      CALL adzi140_set_step_data(TRUE,TRUE,FALSE,FALSE) RETURNING lo_step.*
      EXIT DIALOG

    ON ACTION btn_next
      CALL adzi140_set_step_data(FALSE,TRUE,FALSE,FALSE) RETURNING lo_step.*

      GOTO _btn_ok

    ON ACTION btn_ok

      CALL adzi140_set_step_data(FALSE,FALSE,TRUE,FALSE) RETURNING lo_step.*

      #檢核表格是否存在
      LET ls_table_name = lo_create_table.dct_table_name,ms_table_tail_combine_code
      #CALL sadzi140_db_get_module_leading_code(lo_create_table.dct_module_name,ls_table_name) RETURNING ls_leading_code
      CALL adzi140_check_table_data(ls_table_name) RETURNING lb_table_exist
      IF lb_table_exist THEN
        LET ls_replace_arg = ls_table_name,"|"
        CALL sadzp000_msg_show_error('adz-00095', 'adz-00095', ls_replace_arg, 1)
        #LET lo_create_table.dct_table_name = ls_leading_code
        NEXT FIELD ed_TableName
        CALL DIALOG.setActionActive("btn_next", FALSE)
        CALL DIALOG.setActionActive("btn_ok", FALSE)
      ELSE
        CALL DIALOG.setActionActive("btn_next", lb_enable_next)
        CALL DIALOG.setActionActive("btn_ok", lb_enable_confirm)
      END IF

      #20160617 begin
      #檢查表格是否已經出貨過且刪除
      CALL adzi140_check_delete_table_if_exists(ls_table_name) RETURNING lo_dzhh_t.*
      IF lo_dzhh_t.DZHH001 IS NOT NULL THEN
        LET ls_replace_arg = lo_dzhh_t.DZHH001,"|",lo_dzhh_t.DZHH003,"|"
        CALL sadzp000_msg_show_error('adz-00879', 'adz-00879', ls_replace_arg, 1)
        NEXT FIELD ed_TableName
        CALL DIALOG.setActionActive("btn_next", FALSE)
        CALL DIALOG.setActionActive("btn_ok", FALSE)
      ELSE  
        CALL DIALOG.setActionActive("btn_next", lb_enable_next)
        CALL DIALOG.setActionActive("btn_ok", lb_enable_confirm)
      END IF
      #20160617 end

      LABEL _btn_ok:

      #模組、表格型態及表格均需選擇才能繼續建表動作!
      IF (NVL(lo_create_table.dct_module_name,cs_null_value) = cs_null_value) OR (NVL(lo_create_table.dct_table_type,cs_null_value) = cs_null_value) OR
         (NVL(lo_create_table.dct_table_name,cs_null_value) = cs_null_value) THEN
        LET ls_replace_arg = "|"
        CALL sadzp000_msg_show_error('adz-00177', 'adz-00177', ls_replace_arg, 1)
        CONTINUE DIALOG
      END IF

      LET ls_define_group = cs_group_leading_code||lo_create_table.dct_table_type
      LET lo_create_table.dct_define_group          = ls_define_group
      LET lo_create_table.dct_is_multi_lang_table   = NVL(lo_create_table.dct_is_multi_lang_table,"N")
      LET lo_create_table.dct_is_system_table       = NVL(lo_create_table.dct_is_system_table,"N")
      LET lo_create_table.dct_create_notice_trigger = NVL(lo_create_table.dct_create_notice_trigger,"N") #20190424
      LET lo_create_table.dct_table_space           = NVL(lo_create_table.dct_table_space,"")
      LET lo_create_table.dct_master_db             = NVL(ls_master_db,cs_master_db)
      LET lo_create_table.dct_master_user           = NVL(ls_master_user,cs_master_user)
      LET lo_create_table.dct_is_altered            = NVL(lo_create_table.dct_is_altered,"Y")
      LET lo_create_table.dct_shipping_notice       = "N"
      LET lo_create_table.dct_orig_module_name      = lo_create_table.dct_module_name
      EXIT DIALOG

    ON ACTION CLOSE
      EXIT DIALOG

  END DIALOG

  CLOSE WINDOW w_sadzi140_wcrt

  RETURN lb_result, lo_step.*, lo_create_table.*

END FUNCTION

#設定行業別
FUNCTION adzi140_step_set_industry_type(p_create_table,p_enable_next)
DEFINE
  p_create_table  T_DZEA_CREATE_TABLE,
  p_enable_next   BOOLEAN
DEFINE
  lo_create_table         T_DZEA_CREATE_TABLE,
  lo_step                 T_STEP,
  lb_enable_next          BOOLEAN,
  lo_industry_type        T_INDUSTRY_TYPE,
  ls_orig_table_name      STRING,
  lo_cb_industry_type     ui.ComboBox,
  lo_table_type_combobox  ui.ComboBox,
  lo_window               ui.window,
  lo_form                 ui.form,
  lb_result               BOOLEAN,
  ls_industry_type        STRING,
  ls_table_type_combobox  VARCHAR(50),
  ls_define_group         VARCHAR(100),
  ls_master_db            STRING,
  ls_master_user          STRING,
  ls_table_name           STRING,
  ls_table_leading_code   STRING,
  ls_leading_code         STRING,
  ls_table_char           STRING,
  li_loop                 INTEGER,
  lb_table_exist          BOOLEAN,
  ls_window_desc          STRING,
  ls_module_name          STRING,
  ls_table_type           STRING,
  ls_replace_arg          STRING,
  ls_tail_code            STRING,
  ls_separator            STRING,
  ls_dct_table_name       STRING,
  li_table_length         INTEGER

  LET lo_create_table.* = p_create_table.*
  LET lb_enable_next    = p_enable_next

  LET ls_separator = os.Path.separator()

  LET lb_result = TRUE

  &ifndef DEBUG
  OPEN WINDOW w_sadzi140_ind WITH FORM cl_ap_formpath("ADZ","sadzi140_ind")
  &else
  OPEN WINDOW w_sadzi140_ind WITH FORM sadzi140_util_get_form_path("sadzi140_ind")
  &endif

  CURRENT WINDOW IS w_sadzi140_ind

  CALL sadzi140_util_set_form_title("sadzi140_ind",ms_lang) #20150417 by Hiko

  #填入行業別資料
  LET lo_cb_industry_type = ui.ComboBox.forName("formonly.cb_industry_type")
  CALL sadzi140_util_fill_combobox_by_sql(lo_cb_industry_type,ms_sql_industry)

  #先預設Industry Type 及 Table Name
  LET lo_industry_type.INDUSTRY_TYPE = lo_create_table.dct_industry_type
  LET lo_industry_type.TABLE_NAME    = lo_create_table.dct_table_name
  LET lo_industry_type.TAIL_CODE     = lo_create_table.dct_tail_code

  #紀錄表格原始名稱
  LET ls_orig_table_name = lo_create_table.dct_table_name
  IF lo_industry_type.INDUSTRY_TYPE IS NOT NULL THEN
    DISPLAY lo_industry_type.INDUSTRY_TYPE TO formonly.cb_industry_type
  END IF
  {
  IF lo_industry_type.INDUSTRY_TYPE IS NOT NULL THEN
    LET ls_dct_table_name = lo_create_table.dct_table_name
    LET ls_dct_table_name = ls_dct_table_name.subString(1,ls_dct_table_name.getIndexOf(lo_industry_type.INDUSTRY_TYPE,1)-1)
    LET ls_orig_table_name = ls_dct_table_name
  ELSE
    LET ls_orig_table_name = lo_create_table.dct_table_name
  END IF
  }
  DIALOG
    INPUT lo_industry_type.* FROM sr_industry_type.* ATTRIBUTE(WITHOUT DEFAULTS)
      ON CHANGE cb_industry_type
        LET ls_table_name = ls_orig_table_name

        {
        #判斷表格尾字串是否為"_t"
        LET li_table_length = ls_table_name.getLength()
        LET ls_tail_code = ls_table_name.subString(li_table_length-1,li_table_length)
        }

        LET ls_industry_type = lo_industry_type.INDUSTRY_TYPE
        LET ls_industry_type = ls_industry_type.trim()
        LET lo_industry_type.TABLE_NAME = ls_orig_table_name,ls_industry_type
        {
        IF ls_tail_code <> cs_table_tail_code THEN
          LET lo_industry_type.TABLE_NAME = ls_orig_table_name,ls_industry_type
        ELSE
          LET lo_industry_type.TABLE_NAME = ls_orig_table_name.subString(1,li_table_length-cs_table_tail_code.getLength()),ls_industry_type,cs_table_tail_code
        END IF
        }
        LET lo_create_table.dct_table_name    = lo_industry_type.TABLE_NAME
        LET lo_create_table.dct_industry_type = lo_industry_type.INDUSTRY_TYPE

    END INPUT

    BEFORE DIALOG
      CALL DIALOG.setActionActive("btn_next", lb_enable_next)

    ON ACTION btn_CANCEL
      INITIALIZE lo_create_table TO NULL
      LET lb_result = FALSE
      CALL adzi140_set_step_data(FALSE,FALSE,FALSE,TRUE) RETURNING lo_step.*

      EXIT DIALOG

    ON ACTION btn_back
      CALL adzi140_set_step_data(TRUE,TRUE,FALSE,FALSE) RETURNING lo_step.*
      LET lo_create_table.dct_table_name    = ls_orig_table_name
      LET lo_create_table.dct_industry_type = NVL(ms_topind,cs_default_topind) --cs_industry_type_standard
      EXIT DIALOG

    ON ACTION btn_next
      CALL adzi140_set_step_data(FALSE,TRUE,FALSE,FALSE) RETURNING lo_step.*

      EXIT DIALOG

    ON ACTION btn_ok
      CALL adzi140_set_step_data(FALSE,FALSE,TRUE,FALSE) RETURNING lo_step.*
      #檢核表格是否存在
      LET ls_table_name = lo_create_table.dct_table_name,lo_create_table.dct_tail_code
      CALL adzi140_check_table_data(ls_table_name) RETURNING lb_table_exist
      IF lb_table_exist THEN
        LET ls_replace_arg = ls_table_name,"|"
        CALL sadzp000_msg_show_error('adz-00095', 'adz-00095', ls_replace_arg, 1)
        #LET lo_industry_type.INDUSTRY_TYPE = ""
        #LET lo_industry_type.TABLE_NAME    = ls_orig_table_name
        #LET lo_create_table.dct_industry_type = lo_industry_type.INDUSTRY_TYPE
        #LET lo_create_table.dct_table_name    = lo_industry_type.TABLE_NAME
        NEXT FIELD cb_industry_type
      END IF

      EXIT DIALOG

    ON ACTION CLOSE
      EXIT DIALOG

  END DIALOG

  CLOSE WINDOW w_sadzi140_ind

  RETURN lb_result, lo_step.*, lo_create_table.*

END FUNCTION

#開窗新增Table
FUNCTION adzi140_step_set_table_synonym(p_dialog,p_create_table,p_enable_next)
DEFINE
  p_dialog        ui.Dialog,
  p_create_table  T_DZEA_CREATE_TABLE,
  p_enable_next   BOOLEAN
DEFINE
  lo_create_table T_DZEA_CREATE_TABLE,
  lo_step         T_STEP,
  lb_enable_next  BOOLEAN,
  lo_schema_type  DYNAMIC ARRAY OF T_TABLE_SYNONYM,
  lb_result       BOOLEAN,
  ls_table_name   STRING,
  ed_TableName    STRING,
  ls_master_db    STRING
  LET lb_result = TRUE

  LET lo_create_table.* = p_create_table.*
  LET lb_enable_next    = p_enable_next

  &ifndef DEBUG
  OPEN WINDOW w_sadzi140_syn WITH FORM cl_ap_formpath("ADZ","sadzi140_syn")
  &else
  OPEN WINDOW w_sadzi140_syn WITH FORM sadzi140_util_get_form_path("sadzi140_syn")
  &endif

  CURRENT WINDOW IS w_sadzi140_syn

  CALL sadzi140_util_set_form_title("sadzi140_syn",ms_lang) #20150417 by Hiko

  LET ed_TableName  = lo_create_table.dct_table_name,lo_create_table.dct_tail_code
  LET ls_table_name = lo_create_table.dct_table_name,lo_create_table.dct_tail_code

  #取得Master DB
  LET ls_master_db = ms_master_db

  CALL adzi140_fill_table_synonym_list(ls_master_db,ls_table_name,lo_create_table.dct_module_name)

  DIALOG
    INPUT ARRAY lo_schema_type FROM sr_schema_type.* ATTRIBUTE(WITHOUT DEFAULTS)
      BEFORE INPUT
        CALL lo_schema_type.clear()
        LET lo_schema_type.* = m_table_synonym_list.*
        DISPLAY BY NAME ed_TableName
        CALL DIALOG.setActionHidden("insert",TRUE)
        CALL DIALOG.setActionHidden("append",TRUE)
        CALL DIALOG.setActionHidden("delete",TRUE)
    END INPUT

    BEFORE DIALOG
      CALL DIALOG.setActionActive("btn_next", lb_enable_next)
      CALL DIALOG.setActionActive("btn_ok", TRUE)

    ON ACTION btn_CANCEL
      LET lb_result = FALSE
      CALL adzi140_set_step_data(FALSE,FALSE,FALSE,TRUE) RETURNING lo_step.*
      EXIT DIALOG

    ON ACTION btn_back
      CALL adzi140_set_step_data(TRUE,TRUE,FALSE,FALSE) RETURNING lo_step.*
      EXIT DIALOG

    ON ACTION btn_next
      CALL adzi140_set_step_data(FALSE,TRUE,FALSE,FALSE) RETURNING lo_step.*
      EXIT DIALOG

    ON ACTION btn_OK
      CALL adzi140_set_step_data(FALSE,FALSE,TRUE,FALSE) RETURNING lo_step.*
      EXIT DIALOG

    ON ACTION CLOSE
      EXIT DIALOG

  END DIALOG

  CLOSE WINDOW w_sadzi140_syn

  RETURN lb_result, lo_step.*, lo_schema_type

END FUNCTION

#開窗新增Table
FUNCTION adzi140_step_set_table_synonym_no_window(p_dialog,p_create_table)
DEFINE
  p_dialog        ui.Dialog,
  p_create_table  T_DZEA_CREATE_TABLE
DEFINE
  lo_create_table T_DZEA_CREATE_TABLE,
  lo_step         T_STEP,
  lo_schema_type  DYNAMIC ARRAY OF T_TABLE_SYNONYM,
  lb_result       BOOLEAN,
  ls_table_name   STRING,
  ed_TableName    STRING,
  ls_master_db    STRING

  LET lb_result = TRUE

  LET lo_create_table.* = p_create_table.*

  LET ed_TableName  = lo_create_table.dct_table_name,lo_create_table.dct_tail_code
  LET ls_table_name = lo_create_table.dct_table_name,lo_create_table.dct_tail_code

  #取得Master DB
  LET ls_master_db = ms_master_db

  CALL adzi140_fill_table_synonym_list(ls_master_db,ls_table_name,lo_create_table.dct_module_name)

  LET lo_schema_type.* = m_table_synonym_list.*

  CALL adzi140_set_step_data(FALSE,TRUE,FALSE,FALSE) RETURNING lo_step.*

  RETURN lb_result, lo_step.*, lo_schema_type

END FUNCTION

#開窗新增附屬表格
FUNCTION adzi140_step_create_table_type_list(p_dialog,p_dzea_t,p_step)
DEFINE
  p_dialog      ui.Dialog,
  p_dzea_t      T_DZEA_T,
  p_step        T_STEP
DEFINE
  lo_function_table_list  DYNAMIC ARRAY OF T_FUNCTION_TABLE_TYPE_LIST,
  lo_table_info   T_DZEA_T,
  lb_result       BOOLEAN,
  lo_step         T_STEP,
  ed_TableName    STRING,
  li_aux_count    INTEGER,
  ls_replace_arg  STRING

  LET lb_result = TRUE
  LET lo_step.* = p_step.*

  LET lo_table_info.* = p_dzea_t.*

  &ifndef DEBUG
  OPEN WINDOW w_sadzi140_fnct WITH FORM cl_ap_formpath("ADZ","sadzi140_fnct")
  &else
  OPEN WINDOW w_sadzi140_fnct WITH FORM sadzi140_util_get_form_path("sadzi140_fnct")
  &endif

  CURRENT WINDOW IS w_sadzi140_fnct

  CALL sadzi140_util_set_form_title("sadzi140_fnct",ms_lang) #20150417 by Hiko

  LET ed_TableName = lo_table_info.DZEA001
  CALL adzi140_fill_function_table_type_list(lo_table_info.*) RETURNING li_aux_count

  IF li_aux_count = 0 THEN
    LET ls_replace_arg = lo_table_info.DZEA001,"|"
    CALL sadzp000_msg_show_error('adz-00094', 'adz-00094', ls_replace_arg, 1)
  ELSE

    DIALOG
      INPUT ARRAY lo_function_table_list FROM sr_function_table_list.* ATTRIBUTE(WITHOUT DEFAULTS)
        BEFORE INPUT
          CALL lo_function_table_list.clear()
          LET lo_function_table_list.* = m_function_table_type_list.*
          DISPLAY BY NAME ed_TableName
          CALL DIALOG.setActionHidden("insert",TRUE)
          CALL DIALOG.setActionHidden("append",TRUE)
          CALL DIALOG.setActionHidden("delete",TRUE)
      END INPUT

      BEFORE DIALOG
        CALL DIALOG.setActionActive("btn_back", FALSE)
        CALL DIALOG.setActionActive("btn_next", FALSE)

      ON ACTION btn_CANCEL
        LET lb_result = FALSE
        CALL adzi140_set_step_data(FALSE,FALSE,FALSE,TRUE) RETURNING lo_step.*
        EXIT DIALOG

      ON ACTION btn_back
        CALL adzi140_set_step_data(TRUE,FALSE,FALSE,FALSE) RETURNING lo_step.*
        EXIT DIALOG

      ON ACTION btn_next
        CALL adzi140_set_step_data(FALSE,TRUE,FALSE,FALSE) RETURNING lo_step.*
        EXIT DIALOG

      ON ACTION btn_OK
        CALL adzi140_set_step_data(FALSE,FALSE,TRUE,FALSE) RETURNING lo_step.*
        EXIT DIALOG

      ON ACTION CLOSE
        EXIT DIALOG

    END DIALOG
  END IF

  CLOSE WINDOW w_sadzi140_fnct

  RETURN lb_result, lo_step.*, lo_function_table_list

END FUNCTION

#實際將設定的資料 Insert 到Table中
FUNCTION adzi140_create_dzea_data(p_dialog, p_create_table)
DEFINE
  p_dialog       ui.Dialog,
  p_create_table T_DZEA_CREATE_TABLE
DEFINE
  li_dzea_t_len   INTEGER,
  li_index        INTEGER,
  ls_group_id     STRING,
  lb_result       BOOLEAN,
  ls_replace_arg  STRING,
  lo_create_table T_DZEA_CREATE_TABLE,
  lb_return       BOOLEAN

  LET lo_create_table.* = p_create_table.*

  CALL m_dzea_t.appendElement()
  LET li_dzea_t_len = m_dzea_t.getLength()
  LET m_dzea_t[li_dzea_t_len].DZEA001 = lo_create_table.dct_table_name
  LET m_dzea_t[li_dzea_t_len].DZEA002 = lo_create_table.dct_table_description

  #將資料新增進DZEA_T
  BEGIN WORK
  CALL sadzi140_crud_insert_dzea_t(lo_create_table.*) RETURNING lb_result
  IF NOT lb_result THEN
    ROLLBACK WORK
    LET ls_replace_arg = lo_create_table.dct_table_name,"|"
    CALL sadzp000_msg_show_error('adz-00095', 'adz-00095', ls_replace_arg, 1)
  ELSE
    CALL sadzi140_crud_insert_update_dzeal_t(lo_create_table.*,ms_lang)
    COMMIT WORK
  END IF

  CALL adzi140_refresh_master(p_dialog,ms_search)

  IF lb_result THEN
    FOR li_index = 1 TO m_dzea_t.getLength()
      IF m_dzea_t[li_index].DZEA001 = lo_create_table.dct_table_name THEN
        LET mi_dzea_index = li_index
        EXIT FOR
      END IF
    END FOR

    CALL p_dialog.setCurrentRow("sr_tablelist",mi_dzea_index)
    CALL adzi140_refresh_master(p_dialog,ms_search)
    CALL p_dialog.setCurrentRow("sr_tablecolumns",1)

  END IF

  LET lb_return = lb_result

  RETURN lb_return

END FUNCTION

#修改 Master Table 的資料
FUNCTION adzi140_modify_master_table(p_dialog)
DEFINE
  p_dialog ui.dialog
DEFINE
  lo_t_dzea_t     DYNAMIC ARRAY OF T_DZEA_T,
  li_index        INTEGER,
  ls_replace_arg  STRING,
  ls_table_name   STRING,
  ls_owner        STRING,
  li_rec_index    INTEGER,
  ls_current_item STRING,
  ls_module_code  STRING,
  li_arr_curr     INTEGER,
  ls_group_type   VARCHAR(10),
  lb_active       BOOLEAN,
  lb_data_altered BOOLEAN,
  lo_create_table T_DZEA_CREATE_TABLE
DEFINE
  lb_return  BOOLEAN

  LET lb_data_altered = FALSE

  CALL adzi140_set_editor_mode(p_dialog,cs_editor_mode_edit)

  DIALOG

    INPUT ARRAY lo_t_dzea_t FROM sr_TableList.* ATTRIBUTE(WITHOUT DEFAULTS)
      BEFORE INPUT
        LET lo_t_dzea_t.* = m_dzea_t.*
        CALL DIALOG.setActionHidden("insert",TRUE)
        CALL DIALOG.setActionHidden("append",TRUE)
        CALL DIALOG.setActionHidden("delete",TRUE)
        CALL DIALOG.setActionActive("delete",FALSE)
        CALL DIALOG.setCurrentRow("sr_tablelist",mi_dzea_index)
        
      ON CHANGE lbl_dzea002
        LET li_rec_index = ARR_CURR()
        LET ls_table_name = m_dzea_t[li_rec_index].dzea001
        LET m_dzea_t[li_rec_index].recordtype = NVL(m_dzea_t[li_rec_index].recordtype,cs_data_modified)
        LET lb_data_altered = TRUE

      ON ACTION act_lang_modify INFIELD lbl_dzea002
        LET li_rec_index = ARR_CURR()
        LET ls_table_name = m_dzea_t[li_rec_index].dzea001
        LET m_dzea_t[li_rec_index].recordtype = NVL(m_dzea_t[li_rec_index].recordtype,cs_data_modified)
        &ifndef DEBUG
        BEGIN WORK
          CALL n_dzeal(ls_table_name)
          INITIALIZE m_multi_lang_fields TO NULL
          LET m_multi_lang_fields[1] = ls_table_name
          CALL ap_ref_array2(m_multi_lang_fields," select dzeal003 from dzeal_t where dzeal001 = ? and dzeal002 = '"||ms_lang||"'","") RETURNING m_multi_lang_returns
          LET m_dzea_t[li_rec_index].dzea002 = m_multi_lang_returns[1]
        COMMIT WORK
        &else
        DISPLAY cs_message_tag,"act_lang_modify"
        &endif
        LET lb_data_altered = TRUE

      ON CHANGE lbl_dzea004
        LET li_rec_index = ARR_CURR()
        LET m_dzea_t[li_rec_index].dzea015 = ms_dgenv
        #判斷當為客制環境時,若更改表格資料,則需要異動模組別為 Cxx 模組.
        IF ms_dgenv = cs_dgenv_customize THEN
          LET ls_module_code = m_dzea_t[li_rec_index].dzea003
          CALL sadzi140_db_get_module_code_by_dgenv(ls_module_code,ms_dgenv) RETURNING ls_module_code
          LET m_dzea_t[li_rec_index].dzea003 = ls_module_code
        END IF
        LET lb_data_altered = TRUE

    END INPUT

    ON ACTION CANCEL
      LET lb_data_altered = FALSE
      EXIT DIALOG

    ON ACTION ACCEPT
      BEGIN WORK
        LET mi_dzea_index = ARR_CURR()
        TRY
          FOR li_index = 1 TO m_dzea_t.getLength()
            IF (m_dzea_t[li_index].recordtype = cs_data_modified) OR (li_index = mi_dzea_index) THEN
              LET ls_group_type = cs_group_leading_code,m_dzea_t[li_index].dzea004

              UPDATE dzea_t
                 SET dzea002 = m_dzea_t[li_index].dzea002,
                     dzea003 = m_dzea_t[li_index].dzea003,
                     dzea004 = m_dzea_t[li_index].dzea004,
                     dzea005 = m_dzea_t[li_index].dzea005,
                     dzea006 = m_dzea_t[li_index].dzea006,
                     dzea007 = m_dzea_t[li_index].dzea007,
                     dzea008 = ls_group_type,
                     dzea012 = m_dzea_t[li_index].dzea012
               WHERE dzea001 = m_dzea_t[li_index].dzea001

              #更新多語言檔
              LET lo_create_table.dct_table_name            = m_dzea_t[li_index].dzea001
              LET lo_create_table.dct_table_description     = m_dzea_t[li_index].dzea002
              LET lo_create_table.dct_module_name           = m_dzea_t[li_index].dzea003
              LET lo_create_table.dct_table_type            = m_dzea_t[li_index].dzea004
              LET lo_create_table.dct_is_multi_lang_table   = m_dzea_t[li_index].dzea005
              LET lo_create_table.dct_is_system_table       = m_dzea_t[li_index].dzea006
              LET lo_create_table.dct_create_notice_trigger = m_dzea_t[li_index].dzea007 #20190424
              CALL sadzi140_crud_insert_update_dzeal_t(lo_create_table.*,ms_lang)
            END IF
          END FOR
        CATCH
          LET ls_replace_arg = "DZEA_T","|"
          CALL sadzp000_msg_show_error('adz-00089', 'adz-00089', ls_replace_arg, 1)
        END TRY
      COMMIT WORK

      CALL DIALOG.setCurrentRow("sr_tablelist",mi_dzea_index)
      LET ls_table_name = m_dzea_t[mi_dzea_index].dzea001
      LET ls_owner      = ms_master_user

      #CALL sadzi140_db_update_alter_code(ls_owner,ls_table_name)#20170523 mark
      CALL sadzi140_db_set_alter_code(ls_owner,ls_table_name,"N") #20170523 add

      EXIT DIALOG

    ON ACTION CLOSE
      EXIT DIALOG

  END DIALOG

  CALL adzi140_set_editor_mode(p_dialog,cs_editor_mode_browse)

  LET lb_return = lb_data_altered

  RETURN lb_return

END FUNCTION

FUNCTION adzi140_check_db_data_diff(p_master_user,p_table_name)
DEFINE
  p_master_user STRING,
  p_table_name  STRING
DEFINE
  ls_master_user    STRING,
  ls_table_name     STRING,
  ls_new_version    STRING,
  lb_diff           BOOLEAN, #20170706
  ls_question       STRING,
  lb_alter          BOOLEAN,
  ls_replace_arg    STRING,
  li_columnid_rec   INTEGER, #20170706
  lb_result         BOOLEAN  #20170523
DEFINE
  lb_return        BOOLEAN

  LET ls_master_user = p_master_user
  LET ls_table_name  = p_table_name

  LET lb_alter = FALSE

  #檢核欄位和索引有沒有差異
  CALL sadzi140_shot_check_diff(ls_master_user,ls_table_name,mb_drop_abnormal_column) RETURNING lb_diff
  #檢核欄位ID有沒有差異
  CALL sadzi140_db_get_column_id_status(ls_master_user,ls_table_name) RETURNING li_columnid_rec #20170706
  LET lb_diff = IIF(li_columnid_rec >= 1,TRUE,IIF(mb_debug,TRUE,lb_diff)) #20170706

  #20170425 mark begin
  {
  IF lb_alter THEN
    LET ls_replace_arg = ls_table_name,"|"
    CALL sadzp000_msg_question_box("adz-00115", "adz-00115", ls_replace_arg, 0) RETURNING ls_question
    IF ls_question = cs_question_yes THEN
      LET lb_alter = TRUE
    ELSE
      LET lb_alter = FALSE
    END IF
  ELSE
  }
  #20170425 mark end
  
  IF NOT lb_diff THEN #20170706
    LET lb_alter = FALSE
    CALL sadzi140_db_check_table_if_altered(ls_table_name) RETURNING lb_result #20170523 add
    IF NOT lb_result THEN 
      CALL sadzi140_db_set_alter_code(ls_master_user,ls_table_name,"Y") #20170523 add
      #LET ls_replace_arg = ls_table_name,"|"
      #CALL sadzp000_msg_show_info('adz-01201', 'adz-01201', ls_replace_arg, 0) 
    ELSE   
      LET ls_replace_arg = ls_table_name,"|"
      CALL sadzp000_msg_show_info('adz-00111', 'adz-00111', ls_replace_arg, 0) 
      #CALL sadzp000_msg_show_message_bar('adz-00111',ls_replace_arg,ms_lang) #20170523 add
    END IF  
  ELSE #20170706
   LET lb_alter = lb_diff #20170706
  END IF

  LET lb_return = lb_alter

  RETURN lb_return

END FUNCTION

#檢核表格資料是否存在
FUNCTION adzi140_check_table_data(p_table)
DEFINE
  p_table  STRING
DEFINE
  lb_return     BOOLEAN,
  ls_sql        STRING,
  li_rec_count  INTEGER,
  ls_table_name STRING

  LET ls_table_name = p_table

  #取得資料筆數
  LET ls_sql = "select count(1)                        ",
               "  from dzea_t ea                       ",
               " where 1=1                             ",
               "   and ea.dzea001 = '",ls_table_name,"'"

  PREPARE lpre_check_table_data FROM ls_sql
  DECLARE lcur_check_table_data CURSOR FOR lpre_check_table_data
  OPEN lcur_check_table_data
  FETCH lcur_check_table_data INTO li_rec_count
  CLOSE lcur_check_table_data
  FREE lcur_check_table_data
  FREE lpre_check_table_data

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF

  RETURN lb_return

END FUNCTION

#檢核表格Column值是否存在
FUNCTION adzi140_check_column_data(p_table)
DEFINE
  p_table  STRING
DEFINE
  lb_return     BOOLEAN,
  ls_sql        STRING,
  li_rec_count  INTEGER,
  ls_table_name STRING

  LET ls_table_name = p_table

  #取得資料筆數
  LET ls_sql = "select count(1)                        ",
               "  from dzeb_t eb                       ",
               " where 1=1                             ",
               "   and eb.dzeb001 = '",ls_table_name,"'"

  PREPARE lpre_check_column_data FROM ls_sql
  DECLARE lcur_check_column_data CURSOR FOR lpre_check_column_data
  OPEN lcur_check_column_data
  FETCH lcur_check_column_data INTO li_rec_count
  CLOSE lcur_check_column_data
  FREE lcur_check_column_data
  FREE lpre_check_column_data

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF

  RETURN lb_return

END FUNCTION

#異動(Create,Modify,Delete)Column資料
FUNCTION adzi140_alter_column_data(p_dialog,p_schema_altered) #20180302
DEFINE
  p_dialog        ui.dialog,
  p_schema_altered  BOOLEAN #20180302
DEFINE
  lo_dialog       ui.dialog,
  lb_schema_altered BOOLEAN, #20180302
  lo_t_dzeb_t     DYNAMIC ARRAY OF T_DZEB_T,
  lo_del_list     DYNAMIC ARRAY OF T_DZEB_T, #20160615 modify
  lo_null_list    DYNAMIC ARRAY OF STRING,
  li_rec_cnt      INTEGER,
  li_counts       INTEGER,
  li_null_cnts    INTEGER,
  ls_key_list     STRING,
  ls_null_list    STRING,
  li_arr_curr     INTEGER,
  ls_dzeb002      STRING,
  ls_dzeb011      STRING,
  ls_dzeb017      STRING,
  ls_dzeb018      STRING,
  ls_dzeb022      STRING,
  lo_column_define T_COLUMN_DEFINE,
  ls_dzeb023      STRING,
  lb_error        BOOLEAN,
  lb_success      BOOLEAN,
  ls_schema_sql   STRING,
  ls_err_message  STRING,
  ls_replace_arg  STRING,
  ls_table_name   STRING,
  ls_column_name  STRING,
  ls_column_name_orig STRING,
  ls_column_char  STRING,
  li_loop         INTEGER,
  ls_del_list     STRING,
  ls_main_table   STRING,
  ls_main_module  STRING,
  ls_owner        STRING,
  ls_sql          STRING,
  ls_leading_code STRING,
  ls_custom_code  STRING, #20190304
  ls_dzeb027      STRING, #20190304
  ls_topind       STRING, #20190304
  ls_prog_list    STRING,
  ls_old_dzeb003  STRING,
  ls_new_dzeb003  STRING,
  ls_old_dzeb006  STRING,
  ls_new_dzeb006  STRING,
  ls_old_dzeb012  STRING,
  ls_new_dzeb012  STRING,
  lb_active       BOOLEAN,
  lb_eb004_active BOOLEAN, #20190304
  li_rec_index    INTEGER,
  lb_key_checked  BOOLEAN,
  lb_result       BOOLEAN,
  ls_result       STRING, #20190819
  ls_module_code  STRING,
  ls_ship_notice  STRING,
  ls_lang         STRING,
  ls_master_user  STRING,
  ls_error_code   STRING,
  lb_column_exist BOOLEAN,
  lb_duplicate    BOOLEAN,
  ls_db_function  STRING,
  ls_dzeb012      STRING,
  ls_message      STRING,
  ls_module_name  STRING, #20180903
  li_num_begin    INTEGER, #20190304
  li_num_end      INTEGER, #20190304
  li_pk_cols_cnts INTEGER, #20190819
  ls_effect_table_list STRING,
  ls_no_mapping_data   STRING,
  ls_table_type        STRING, #20170523
  lb_check_value       BOOLEAN,
  lb_data_altered      BOOLEAN,
  lb_dzeb006_changed   BOOLEAN,
  ls_current_item      STRING, #20201006
  lb_dzeb003_changed   BOOLEAN, #20180412
  lo_dzeb027_json      T_DZEB027_JSON, #20190304
  lo_column_info       DYNAMIC ARRAY OF T_COLUMN_INFO, #20171120
  lo_collect_messages  DYNAMIC ARRAY OF T_COLLECT_MESSAGE, #20170523
  lo_mapping_data      DYNAMIC ARRAY OF T_MAPPING_DATA,
  lo_lang_arr          DYNAMIC ARRAY OF T_LANGUAGE_TYPE,
  lo_col_property      T_COL_PROPERTY,
  lo_KEY_IF_EXCEED     T_KEY_IF_EXCEED, #20160616
  lo_dzhl_t            T_DZHL_T
DEFINE
  lb_return BOOLEAN

  LET lo_dialog         = p_dialog
  LET lb_schema_altered = p_schema_altered #20180302
  
  LET m_old_dzeb_t.*     = m_dzeb_t.*
  
  LET ls_del_list        = ""
  LET lb_data_altered    = FALSE
  LET lb_dzeb006_changed = FALSE
  LET lb_error           = FALSE
  LET lb_success         = TRUE
  LET lb_key_checked     = FALSE
  LET ls_master_user     = ms_master_user
  LET lb_dzeb003_changed = FALSE #20180412

  CALL adzi140_set_editor_mode(p_dialog,cs_editor_mode_edit)
  LET li_pk_cols_cnts = adzi140_get_table_pk_cols_cnts(m_dzea_t[mi_dzea_index].dzea001) -- 20190819

  DIALOG

    INPUT ARRAY lo_t_dzeb_t FROM sr_TableColumns.* ATTRIBUTES(WITHOUT DEFAULTS)

      BEFORE INPUT
        LET lo_t_dzeb_t.* = m_dzeb_t.*
        CALL DIALOG.setActionHidden("insert",TRUE)
        CALL DIALOG.setActionHidden("append",TRUE)
        CALL DIALOG.setActionHidden("delete",TRUE)
        CALL DIALOG.setCurrentRow("sr_tablecolumns",mi_dzeb_index)

      BEFORE ROW
        LET li_arr_curr = ARR_CURR()
        #紀錄舊值設定 begin
        LET ls_old_dzeb003 = lo_t_dzeb_t[li_arr_curr].dzeb003
        LET ls_old_dzeb006 = lo_t_dzeb_t[li_arr_curr].dzeb006
        LET ls_old_dzeb012 = lo_t_dzeb_t[li_arr_curr].dzeb012
        #紀錄舊值設定 end
        LET ls_dzeb022 = lo_t_dzeb_t[li_arr_curr].dzeb022
        CALL adzi140_parse_dzeb027_json_data(lo_t_dzeb_t[li_arr_curr].*) RETURNING lo_dzeb027_json.* #20190304

        LET lb_active = IIF(
                            (
                             (
                               (NVL(lo_t_dzeb_t[li_arr_curr].dzeb030,ms_dgenv) = ms_dgenv)
                               OR
                               (ls_dzeb022.subString(1,cs_cdf_user_define.getLength()) = cs_cdf_user_define)
                               OR
                               mb_debug
                             ) AND NVL(lo_t_dzeb_t[li_arr_curr].dzeb031,cs_null_value) <> cs_column_shipped
                            ),
                            TRUE,FALSE
                           )

        #20190304 begin                   
        IF (ms_DGENV = cs_dgenv_standard) THEN                            
          LET lb_active = IIF(NVL(lo_dzeb027_json.TOPIND,ms_topind) <> ms_topind,FALSE,TRUE)
        END IF
        #20190304 end  
        
        CALL DIALOG.setFieldActive("sr_tablecolumns.lbl_dzeb002", lb_active)
        #CALL DIALOG.setFieldActive("sr_tablecolumns.lbl_dzeb003", lb_active)
        #20190304 begin                   
        IF (ms_DGENV = cs_dgenv_standard) THEN   
          --IF ((NVL(lo_dzeb027_json.TOPIND,ms_topind) <> cs_default_topind) OR (lo_t_dzeb_t[li_arr_curr].dzeb030 <> ms_dgenv)) AND -- 20210826 mark 
          --   (m_dzea_t[mi_dzea_index].dzea014 <> ms_topind) THEN -- 20210826 mark
          IF ((NVL(lo_dzeb027_json.TOPIND,ms_topind) <> ms_topind) OR (lo_t_dzeb_t[li_arr_curr].dzeb030 <> ms_dgenv)) THEN --20210826 modify cs_default_topind to ms_topind
            LET lb_eb004_active = FALSE
          ELSE
            LET lb_eb004_active = lb_active 
          END IF         
          CALL DIALOG.setFieldActive("sr_tablecolumns.lbl_dzeb004",lb_eb004_active)
        ELSE
          CALL DIALOG.setFieldActive("sr_tablecolumns.lbl_dzeb004",lb_active)
        END IF 
        #20190304 end  
        CALL DIALOG.setFieldActive("sr_tablecolumns.lbl_dzeb005", lb_active)
        #2190304 begin
        IF ms_topind <> cs_default_topind THEN  
          CALL DIALOG.setFieldActive("sr_tablecolumns.lbl_dzeb006", lb_active)
        ELSE
          CALL DIALOG.setFieldActive("sr_tablecolumns.lbl_dzeb006", TRUE)  
        END IF
        #2190304 end
        #CALL DIALOG.setFieldActive("sr_tablecolumns.lbl_dzeb012", lb_active)
        CALL DIALOG.setFieldActive("sr_tablecolumns.lbl_dzeb024", lb_active)

        #只要欄位已經出貨或者位於客制環境但該欄位為標準欄位, 則該欄位不可以刪除
        LET lb_active = IIF(
                            (NVL(lo_t_dzeb_t[li_arr_curr].dzeb031,cs_null_value) = "Y")
                            OR 
                            ((ms_DGENV = cs_dgenv_customize) AND (lo_t_dzeb_t[li_arr_curr].dzeb030 = cs_dgenv_standard))
                            OR
                            (NVL(lo_dzeb027_json.TOPIND,ms_topind) <> ms_topind) #20190304
                            OR
                            lb_schema_altered, #20180302
                            FALSE,TRUE
                           )
        CALL DIALOG.setActionActive("delete", lb_active)

      ON ACTION DELETE
        LET li_arr_curr = ARR_CURR()
        LET ls_dzeb002 = lo_t_dzeb_t[li_arr_curr].dzeb002
        IF lo_t_dzeb_t[li_arr_curr].dzeb002 IS NOT NULL THEN 
          LET ls_del_list = ls_del_list,",'",lo_t_dzeb_t[li_arr_curr].dzeb002,"'"
          LET lo_del_list[lo_del_list.getLength()+1].* = lo_t_dzeb_t[li_arr_curr].*
          #刪除的欄位是Key才作用
          IF lo_t_dzeb_t[li_arr_curr].dzeb004 = "Y" THEN
            LET lb_data_altered = TRUE
          END IF
          CALL DIALOG.deleteRow("sr_tablecolumns",li_arr_curr)
          #有刪除欄位, 則將所有非新增或是拖動欄位都標示為已修改以更動順序
          FOR li_loop = 1 TO lo_t_dzeb_t.getLength()
            IF (NVL(lo_t_dzeb_t[li_loop].recordtype,cs_null_default) <> cs_data_drag) AND
               (NVL(lo_t_dzeb_t[li_loop].recordtype,cs_null_default) <> cs_data_added) THEN
              LET lo_t_dzeb_t[li_loop].recordtype = NVL(lo_t_dzeb_t[li_loop].recordtype,cs_data_modified)
            END IF
            LET lo_t_dzeb_t[li_loop].dzebseq = li_loop
          END FOR
          #20160615 begin
          #新增 Log
          CALL adzi140_set_column_alter_log(lo_del_list[lo_del_list.getLength()].*,cs_data_delete)
          #20160615 end
          DISPLAY cs_warning_tag,"Column ",ls_dzeb002," deleted !!"  
        ELSE 
          DISPLAY cs_warning_tag,"Column is null !!"  
        END IF   

      ON ROW CHANGE
        LET lo_t_dzeb_t[ARR_CURR()].RECORDTYPE = NVL(lo_t_dzeb_t[ARR_CURR()].RECORDTYPE,cs_data_modified)
        DISPLAY lo_t_dzeb_t[ARR_CURR()].* TO sr_TableColumns[ARR_CURR()].*

      ON CHANGE lbl_dzeb002
        LABEL _lbl_dzeb002:
        #判斷輸入欄位名稱是否重複
        LET lb_duplicate = FALSE
        LET li_arr_curr = ARR_CURR()
        LET ls_column_name = lo_t_dzeb_t[li_arr_curr].dzeb002
        FOR li_rec_cnt = 1 TO lo_t_dzeb_t.getLength()
          IF li_rec_cnt <> li_arr_curr THEN
            IF lo_t_dzeb_t[li_rec_cnt].dzeb002 = ls_column_name THEN
              LET lb_duplicate = TRUE
              EXIT FOR
            END IF
          END IF
        END FOR
        IF lb_duplicate THEN
          LET ls_replace_arg = ls_column_name,"|"
          CALL sadzp000_msg_show_error("adz-00116", "adz-00116", ls_replace_arg, 1)
          LET lo_t_dzeb_t[li_arr_curr].dzeb002 = ""
          NEXT FIELD lbl_dzeb002
          CONTINUE DIALOG
        END IF

        #判斷輸入欄位名稱是否已經存在資料庫
        LET ls_column_name = lo_t_dzeb_t[ARR_CURR()].dzeb002
        LET ls_column_name_orig = lo_t_dzeb_t[ARR_CURR()].dzeb002_orig

        IF NVL(ls_column_name_orig,cs_null_value) <> cs_null_value THEN
          IF ls_column_name <> ls_column_name_orig THEN
            CALL sadzp000_msg_show_error("adz-00211", "adz-00211", NULL, 1)
            LET lo_t_dzeb_t[ARR_CURR()].dzeb002 = ls_column_name_orig
            NEXT FIELD lbl_dzeb002
            CONTINUE DIALOG
          END IF
          {
          CALL sadzi140_db_check_db_column_exist_by_single_column(ls_master_user,ls_column_name_orig,ls_column_name) RETURNING lb_column_exist
          IF lb_column_exist THEN
            CALL sadzp000_msg_show_error("adz-00211", "adz-00211", NULL, 1)
            LET lo_t_dzeb_t[ARR_CURR()].dzeb002 = ls_column_name_orig
            NEXT FIELD lbl_dzeb002
            CONTINUE DIALOG
          ELSE
            LET ls_del_list = ls_del_list,",'",ls_column_name_orig,"'"
          END IF
          }
        END IF

        #判斷輸入欄位名稱是否含有其他 Unicode
        LET ls_column_name = lo_t_dzeb_t[ARR_CURR()].dzeb002
        FOR li_loop = 1 TO ls_column_name.getLength()
          LET ls_column_char = NVL(ls_column_name.subString(li_loop,li_loop),"@")
          IF ls_column_char NOT MATCHES "[a-z]" AND ls_column_char NOT MATCHES "[0-9]" THEN
            CALL sadzp000_msg_show_error("adz-00127", "adz-00127", NULL, 1)
            LET lo_t_dzeb_t[ARR_CURR()].dzeb002 = ""
            NEXT FIELD lbl_dzeb002
            EXIT FOR
          END IF
        END FOR

        #判斷輸入欄位名稱是否符合命名原則
        CALL adzi140_get_column_define_type(m_dzea_t[mi_dzea_index].dzea001,lo_t_dzeb_t[ARR_CURR()].dzeb002,m_dzea_t[mi_dzea_index].dzea014) RETURNING lo_column_define.*,ls_leading_code,ls_custom_code #20190304
        IF lo_column_define.cd_DEFINE_TYPE = cs_no_define THEN
          CALL sadzp000_msg_show_error('adz-00126', 'adz-00126', ls_replace_arg, 1)
          LET lo_t_dzeb_t[ARR_CURR()].dzeb002 = ls_leading_code
          NEXT FIELD lbl_dzeb002
          CONTINUE DIALOG
        ELSE
          #20170214 begin
          LET ls_main_table = m_dzea_t[mi_dzea_index].dzea001
          CALL sadzi140_util_get_table_leading_code(ls_main_table) RETURNING ls_leading_code
          #20170214 end
          #若為序號欄位, 則要將序號值填入dzeb023
          IF lo_column_define.cd_DEFINE_TYPE = cs_cdf_serial_number THEN
            LET ls_dzeb002 = lo_t_dzeb_t[ARR_CURR()].dzeb002
            #20170214 begin
            IF ls_dzeb002.getIndexOf(ls_leading_code,1) = 0 THEN
              LET ls_dzeb002 = ls_leading_code,ls_dzeb002
              LET lo_t_dzeb_t[ARR_CURR()].dzeb002 = ls_dzeb002 
              GOTO _lbl_dzeb002 #重新做一次檢查
            END IF 
            #20170214 end
            #20190304 begin
            CALL adzi140_get_column_numeric_digit(ls_dzeb002) RETURNING li_num_begin,li_num_end 
            LET ls_dzeb023 = ls_dzeb002.subString(li_num_begin,li_num_end)
            LET lo_t_dzeb_t[ARR_CURR()].dzeb023 = ls_dzeb023
            LET ls_custom_code = adzi140_get_custom_code(ls_main_table,ls_dzeb002,ms_column_user_define_code)
            CALL adzi140_get_dzeb027_json_data(ms_topind,ls_dzeb023,ls_custom_code) RETURNING ls_dzeb027
            LET lo_t_dzeb_t[ARR_CURR()].DZEB027 = ls_dzeb027
            #20190304 end
          ELSE
            #20170214 begin
            LET ls_dzeb002 = lo_t_dzeb_t[ARR_CURR()].dzeb002
            IF ls_dzeb002.getIndexOf(ls_leading_code,1) = 0 THEN
              LET ls_dzeb002 = ls_leading_code,ls_dzeb002
              LET lo_t_dzeb_t[ARR_CURR()].dzeb002 = ls_dzeb002 
              GOTO _lbl_dzeb002 #重新做一次檢查
            END IF 
            #20170214 end
            #否則取得對應的預設值
            LET lo_t_dzeb_t[ARR_CURR()].dzeb003 = lo_column_define.cd_DESC
            LET lo_t_dzeb_t[ARR_CURR()].dzeb006 = lo_column_define.cd_COLUMN_TYPE
            LET lo_t_dzeb_t[ARR_CURR()].dzeb022 = lo_column_define.cd_DEFINE_TYPE
            CALL adzi140_on_change_dzeb006(lo_t_dzeb_t[ARR_CURR()].*) RETURNING lo_t_dzeb_t[ARR_CURR()].*
          END IF
          LET lo_t_dzeb_t[ARR_CURR()].dzeb022 = lo_column_define.cd_DEFINE_TYPE
        END IF
        #LET lb_data_altered = TRUE

      AFTER FIELD lbl_dzeb003
        LET ls_new_dzeb003 = GET_FLDBUF(lbl_dzeb003)
        IF ls_new_dzeb003 <> ls_old_dzeb003 THEN
          GOTO _lbl_dzeb003
        END IF

      ON CHANGE lbl_dzeb003
        LABEL _lbl_dzeb003:
        LET li_arr_curr = ARR_CURR()
        LET lo_t_dzeb_t[li_arr_curr].RECORDTYPE = NVL(lo_t_dzeb_t[li_arr_curr].RECORDTYPE,cs_data_modified)
        LET lo_t_dzeb_t[li_arr_curr].lang_modified = cs_lang_modified
        #設定客制標示, 已經設為客制的則不再重設
        IF lo_t_dzeb_t[li_arr_curr].dzeb029 <> cs_dgenv_customize THEN
          LET lo_t_dzeb_t[li_arr_curr].dzeb029 = ms_dgenv
        END IF
        &ifndef DEBUG
        #檢查輸入的資料在unicode下是否屬於繁體或簡體字的範圍
        IF NOT cl_chk_tworcn(ms_lang,lo_t_dzeb_t[li_arr_curr].dzeb003,5) THEN
          #LET lo_t_dzeb_t[li_arr_curr].dzeb003 = ""
          NEXT FIELD lbl_dzeb003
        END IF
        &endif
        LET lb_dzeb003_changed = TRUE #20180412
        #LET lb_data_altered = TRUE

      ON CHANGE lbl_dzeb004
        LET lb_key_checked = TRUE
        LET li_arr_curr = ARR_CURR()

        #20170214 begin
        IF lo_t_dzeb_t[li_arr_curr].dzeb004 = "Y" AND lo_t_dzeb_t[li_arr_curr].dzeb006 IS NULL THEN
          CALL sadzp000_msg_show_error("adz-00957", "adz-00957", NULL, 1) #勾選欄位為 Key 之前, 請先設定其欄位屬性
          LET lo_t_dzeb_t[li_arr_curr].dzeb004 = "N"
          NEXT FIELD lbl_dzeb006
        END IF  
        #20170214 end

        #20240416 begin
        IF (m_dzea_t[mi_dzea_index].dzea014 = cs_default_topind) AND 
           (m_dzea_t[mi_dzea_index].dzea018 = cs_dgenv_standard) AND 
           (ms_topind <> cs_default_topind) THEN
          CALL sadzp000_msg_show_error("adz-01818", "adz-01818", NULL, 1) #行業環境下不可以對標準表格新增 PK 欄位
          LET lo_t_dzeb_t[li_arr_curr].dzeb004 = "N"
          NEXT FIELD lbl_dzeb006
        END IF
        #20240416 end

        #20160615 begin
        IF lo_t_dzeb_t[li_arr_curr].dzeb006 LIKE "B%" THEN
          LET lo_t_dzeb_t[li_arr_curr].dzeb004 = "N"
          CALL sadzp000_msg_show_error("adz-00871", "adz-00871", NULL, 1) #文件或二進位型態的欄位無法設為主鍵
          CONTINUE DIALOG 
        END IF 
        #20160615 end

        #20190819 begin
        CALL sadzi140_db_check_key_cols_if_exceed(m_dzea_t[mi_dzea_index].dzea018,lo_t_dzeb_t,li_pk_cols_cnts) RETURNING lb_result,ls_result
        IF lb_result THEN 
          LET lo_t_dzeb_t[li_arr_curr].dzeb004 = "N"
          LET ls_replace_arg = ls_result,"|"
          CALL sadzp000_msg_show_error("adz-01598", "adz-01598", ls_replace_arg, 1) #主鍵欄位個數已經超過上限 %1, 將自動取消勾選.
          CONTINUE DIALOG 
        END IF
        #20190819 end
        
        #20160616 begin
        CALL sadzi140_db_check_key_length_if_exceed(lo_t_dzeb_t) RETURNING lo_KEY_IF_EXCEED.*
        IF lo_KEY_IF_EXCEED.kie_result THEN
          LET lo_t_dzeb_t[li_arr_curr].dzeb004 = "N"
          LET ls_replace_arg = lo_KEY_IF_EXCEED.kie_user_def,"|",lo_KEY_IF_EXCEED.kie_sys_def,"|"
          CALL sadzp000_msg_show_error("adz-00877", "adz-00877", ls_replace_arg, 1) #您所設定的主鍵或索引總長度 (%1) 超出系統設定值 (%2)
          CONTINUE DIALOG 
        END IF 
        #20160616 end
        
        LET lo_t_dzeb_t[li_arr_curr].RECORDTYPE = NVL(lo_t_dzeb_t[li_arr_curr].RECORDTYPE,cs_data_modified)
        IF lo_t_dzeb_t[li_arr_curr].dzeb004 = "Y" THEN
          LET ls_table_name = m_dzea_t[mi_dzea_index].dzea001
          #查看看DUMMY欄位是否存在, 若不存在, 則需輸入Patch賦予值
          CALL sadzi140_db_check_db_column_exist_by_single_column(ms_master_user,ls_table_name,cs_dummy_column_name) RETURNING lb_result
          IF NOT lb_result THEN
            -- 如果表格已存在實體資料庫, 欲設定新Key欄位, 要先設定欄位屬性
            IF lo_t_dzeb_t[li_arr_curr].dzeb006 IS NULL THEN
              CALL sadzp000_msg_show_error("adz-00689", "adz-00689", NULL, 1)
              LET lo_t_dzeb_t[li_arr_curr].dzeb004 = "N"
              NEXT FIELD lbl_dzeb006
            ELSE
              CALL sadzi140_inp_run(lo_t_dzeb_t[li_arr_curr].*) RETURNING lb_result
              LET lo_t_dzeb_t[li_arr_curr].dzeb004 = IIF(lb_result,"Y","N")
            END IF
          END IF
        END IF
        LET lo_t_dzeb_t[li_arr_curr].dzeb005 = lo_t_dzeb_t[li_arr_curr].dzeb004
        LET lb_data_altered = TRUE

        #Log begin
        LET ls_message = sadzp000_msg_get_message("adz-00783",ms_lang)
        LET ls_replace_arg = lo_t_dzeb_t[li_arr_curr].dzeb002,"|",lo_t_dzeb_t[li_arr_curr].dzeb004,"|"
        LET ls_message = sadzp000_msg_replace_message(ls_message,ls_replace_arg)

        CALL sadzi140_logs_write_log(cs_logs_level_information,cs_logs_type_column,ls_message) RETURNING lb_result
        #Log end

      ON CHANGE lbl_dzeb005
        LET lo_t_dzeb_t[ARR_CURR()].RECORDTYPE = NVL(lo_t_dzeb_t[ARR_CURR()].RECORDTYPE,cs_data_modified)

      AFTER FIELD lbl_dzeb006
        LET ls_new_dzeb006 = GET_FLDBUF(lbl_dzeb006)
        IF ls_new_dzeb006 <> ls_old_dzeb006 THEN
          #20190219 begin 
          LET ls_main_module = m_dzea_t[mi_dzea_index].dzea003
          CALL adzi140_check_column_attribute_valid(lo_t_dzeb_t[ARR_CURR()].*,ls_main_module,ms_dgenv,ls_old_dzeb006) RETURNING lb_result,ls_old_dzeb006,lo_t_dzeb_t[ARR_CURR()].*
          IF NOT lb_result THEN 
            NEXT FIELD lbl_dzeb006
            CONTINUE DIALOG
          END IF
          #20190219 end 
        ELSE  
          LET lb_dzeb006_changed = TRUE
        END IF
        
      #欄位屬性
      ON CHANGE lbl_dzeb006
        #20190219 begin 
        LET ls_main_module = m_dzea_t[mi_dzea_index].dzea003
        CALL adzi140_check_column_attribute_valid(lo_t_dzeb_t[ARR_CURR()].*,ls_main_module,ms_dgenv,ls_old_dzeb006) RETURNING lb_result,ls_old_dzeb006,lo_t_dzeb_t[ARR_CURR()].*
        IF NOT lb_result THEN 
          NEXT FIELD lbl_dzeb006
          CONTINUE DIALOG
        ELSE  
          LET lb_dzeb006_changed = TRUE
        END IF
        #20190219 end 

      #欄位屬性
      #ON ACTION controlp INFIELD lbl_dzeb006
      ON ACTION lbl_dzeb006 INFIELD lbl_dzeb006
        LABEL _type_modify_lbl_dzeb006: 
        LET li_arr_curr = ARR_CURR()
        #Control-p begin
        &ifndef DEBUG
          #CALL q_dzeb006(FALSE,TRUE,lo_t_dzeb_t[li_arr_curr].dzeb006,lo_t_dzeb_t[li_arr_curr].dzeb006_DESC) RETURNING lo_col_property.* #2012.11.23
          #henry:20130304->修改開窗的參數設定
          LET g_qryparam.state = "i"
          LET g_qryparam.reqry = TRUE
          LET g_qryparam.default1 = lo_t_dzeb_t[li_arr_curr].dzeb006
          LET g_qryparam.default2 = lo_t_dzeb_t[li_arr_curr].dzeb006_DESC
          LET g_qryparam.default3 = ''
          CALL q_dzeb006()
          LET lo_col_property.C_PROPERTY     = g_qryparam.return1
          LET lo_col_property.C_DESC         = g_qryparam.return2
          LET lo_col_property.C_WIDGET_WIDTH = g_qryparam.return3
        &endif
        LET lo_t_dzeb_t[li_arr_curr].RECORDTYPE   = cs_data_modified
        LET lo_t_dzeb_t[li_arr_curr].dzeb006      = lo_col_property.C_PROPERTY
        LET lo_t_dzeb_t[li_arr_curr].dzeb006_DESC = lo_col_property.C_DESC
        #Control-p end

        #20190219 begin 
        LET ls_main_module = m_dzea_t[mi_dzea_index].dzea003
        CALL adzi140_check_column_attribute_valid(lo_t_dzeb_t[li_arr_curr].*,ls_main_module,ms_dgenv,ls_old_dzeb006) RETURNING lb_result,ls_old_dzeb006,lo_t_dzeb_t[li_arr_curr].*
        IF NOT lb_result THEN 
          NEXT FIELD lbl_dzeb006
          CONTINUE DIALOG
        ELSE  
          LET lb_dzeb006_changed = TRUE
        END IF
        #20190219 end 

        NEXT FIELD lbl_dzeb006

      AFTER FIELD lbl_dzeb012
        LET ls_new_dzeb012 = GET_FLDBUF(lbl_dzeb012)
        IF ls_new_dzeb012 <> ls_old_dzeb012 THEN
          LET li_arr_curr = ARR_CURR()
          #離開欄位時也做 on change 的判定
          IF (lo_t_dzeb_t[li_arr_curr].dzeb012 IS NOT NULL) AND (lo_t_dzeb_t[li_arr_curr].dzeb012 <> ASCII(32)) THEN
            GOTO _lbl_dzeb012
          END IF
        END IF
       
      ON CHANGE lbl_dzeb012
        LABEL _lbl_dzeb012:
        LET li_arr_curr = ARR_CURR()
        LET lb_check_value = TRUE
        CALL sadzi140_util_check_if_enable_set_default_value(lo_t_dzeb_t[li_arr_curr].dzeb007) RETURNING lb_result
        IF NOT lb_result THEN
          LET ls_replace_arg = UPSHIFT(lo_t_dzeb_t[li_arr_curr].dzeb007),"|"
          CALL sadzp000_msg_show_error('adz-00577', 'adz-00577', ls_replace_arg, 0)
          LET lo_t_dzeb_t[li_arr_curr].dzeb012 = NULL
        ELSE
          #先取得Mapping資料
          CALL sadzi140_util_get_global_var_to_db_func_mapping_data() RETURNING lo_mapping_data
          #再從Mapping Data中取得 DB Function
          CALL sadzi140_util_get_db_func_by_mapping_data(lo_t_dzeb_t[li_arr_curr].dzeb012,lo_t_dzeb_t[li_arr_curr].dzeb006,lo_mapping_data) RETURNING ls_db_function
          IF ls_db_function IS NOT NULL THEN
            #取回的 DB function 不為空, 則直接設回預設值
            LET lo_t_dzeb_t[li_arr_curr].dzeb012 = ls_db_function
          ELSE
            #":" 開頭的請 User 到 adzi150 設定
            LET ls_dzeb012 = lo_t_dzeb_t[li_arr_curr].dzeb012
            IF ls_dzeb012.subString(1,1) = ":" THEN
              LET ls_no_mapping_data = ""
              FOR li_loop = 1 TO lo_mapping_data.getLength()
                IF lo_mapping_data[li_loop].md_DB_FUNCTION IS NULL THEN
                  LET ls_no_mapping_data = ls_no_mapping_data," '",lo_mapping_data[li_loop].md_GLOBAL_VARIABLE,"' ,"
                END IF
              END FOR
              LET ls_replace_arg = "\n",ls_no_mapping_data,"|"
              CALL sadzp000_msg_show_error('adz-00581', 'adz-00581', ls_replace_arg, 0)
              LET lo_t_dzeb_t[li_arr_curr].dzeb012 = ""
              LET lb_check_value = FALSE
              NEXT FIELD lbl_dzeb012
            END IF
          END IF
          IF lb_check_value THEN
            #進行一般資料檢核
            CALL sadzi140_util_check_default_value(lo_t_dzeb_t[li_arr_curr].dzeb007,lo_t_dzeb_t[li_arr_curr].dzeb008,lo_t_dzeb_t[li_arr_curr].dzeb012) RETURNING lb_result,ls_error_code
            IF NOT lb_result THEN
              CALL sadzp000_msg_show_error(ls_error_code, ls_error_code, NULL, 0)
              LET lo_t_dzeb_t[li_arr_curr].dzeb012 = "" #20191004
              NEXT FIELD lbl_dzeb012
            ELSE #20200330
              CALL sadzi140_util_refresh_default_value(lo_t_dzeb_t[li_arr_curr].dzeb012) RETURNING lo_t_dzeb_t[li_arr_curr].dzeb012 #20200330      
            END IF
          END IF
          #設定客制標示, 已經設為客制的則不再重設
          IF lo_t_dzeb_t[li_arr_curr].dzeb029 <> cs_dgenv_customize THEN
            LET lo_t_dzeb_t[li_arr_curr].dzeb029 = ms_dgenv
          END IF
          LET lo_t_dzeb_t[li_arr_curr].RECORDTYPE = NVL(lo_t_dzeb_t[li_arr_curr].RECORDTYPE,cs_data_modified)

        END IF
        #LET lb_data_altered = TRUE

      #ON ACTION act_lang_modify INFIELD lbl_dzeb003 
      ON ACTION lbl_dzeb003 INFIELD lbl_dzeb003 
        LABEL _act_lang_modify_lbl_dzeb003:
        LET li_rec_index = ARR_CURR()
        LET ls_column_name = lo_t_dzeb_t[li_rec_index].dzeb002
        &ifndef DEBUG
        BEGIN WORK
        CALL n_dzebl(ls_column_name)
        INITIALIZE m_multi_lang_fields TO NULL
        LET m_multi_lang_fields[1] = ls_column_name
        CALL ap_ref_array2(m_multi_lang_fields," select dzebl003 from dzebl_t where dzebl001 = ? and dzebl002 = '"||ms_lang||"'","") RETURNING m_multi_lang_returns
        LET lo_t_dzeb_t[li_rec_index].dzeb003 = m_multi_lang_returns[1]
        COMMIT WORK
        &else
        DISPLAY cs_message_tag,"act_lang_modify"
        &endif
        LET lo_t_dzeb_t[li_rec_index].RECORDTYPE = NVL(lo_t_dzeb_t[li_rec_index].RECORDTYPE,cs_data_modified)
        LET lb_dzeb003_changed = TRUE #20180412
        #LET lb_data_altered = TRUE

      ON ACTION act_lang_modify INFIELD lbl_dzeb024
        LET li_rec_index = ARR_CURR()
        LET ls_column_name = lo_t_dzeb_t[li_rec_index].dzeb002
        &ifndef DEBUG
        BEGIN WORK
        CALL n_dzebl(ls_column_name)
        INITIALIZE m_multi_lang_fields TO NULL
        LET m_multi_lang_fields[1] = ls_column_name
        CALL ap_ref_array2(m_multi_lang_fields," select dzebl004 from dzebl_t where dzebl001 = ? and dzebl002 = '"||ms_lang||"'","") RETURNING m_multi_lang_returns
        LET lo_t_dzeb_t[li_rec_index].dzeb024 = m_multi_lang_returns[1]
        COMMIT WORK
        &else
        DISPLAY cs_message_tag,"act_lang_modify"
        &endif
        LET lo_t_dzeb_t[li_rec_index].RECORDTYPE = NVL(lo_t_dzeb_t[li_rec_index].RECORDTYPE,cs_data_modified)
        #LET lb_data_altered = TRUE

      #20201006 begin  
      ON ACTION controlp -- ATTRIBUTES(ACCELERATOR="control-p")
        LET ls_current_item = DIALOG.getCurrentItem()
        CASE
          WHEN ls_current_item = "sr_tablecolumns.lbl_dzeb003"
            GOTO _act_lang_modify_lbl_dzeb003
          WHEN ls_current_item = "sr_tablecolumns.lbl_dzeb006"
            GOTO _type_modify_lbl_dzeb006
        END CASE
      #20201006 end
      
    END INPUT

    {
    ON ACTION btn_ColumnCreate
      #CALL DIALOG.insertRow("sr_tablecolumns",ARR_CURR())
      CALL DIALOG.appendRow("sr_tablecolumns")
    }

    {
    ON ACTION btn_ColumnDelete
      CALL DIALOG.deleteRow("sr_tablecolumns",ARR_CURR())
    }

    #20180302 begin
    BEFORE DIALOG
      IF lb_schema_altered THEN 
        CALL DIALOG.setActionActive("cancel", FALSE)
        CALL DIALOG.setActionActive("close", FALSE)
      ELSE
        CALL DIALOG.setActionActive("cancel", TRUE)
        CALL DIALOG.setActionActive("close", TRUE)
      END IF 
    #20180302 end
    
    ON ACTION CANCEL
      LET lb_data_altered = FALSE
      LET lb_dzeb006_changed = FALSE
      LET ls_del_list = ""
      CALL adzi140_fill_dzeb_t(m_dzea_t[mi_dzea_index].dzea001)
      EXIT DIALOG

    ON ACTION ACCEPT
      LET li_arr_curr = ARR_CURR()

      #20210222 begin
      LET ls_current_item = DIALOG.getCurrentItem()
      IF ls_current_item = "sr_tablecolumns.lbl_dzeb002" AND lo_t_dzeb_t[li_arr_curr].dzeb022 IS NULL THEN
        #判斷當下欄位輸入欄位名稱是否符合命名原則
        CALL adzi140_get_column_define_type(m_dzea_t[mi_dzea_index].dzea001,lo_t_dzeb_t[li_arr_curr].dzeb002,m_dzea_t[mi_dzea_index].dzea014) RETURNING lo_column_define.*,ls_leading_code,ls_custom_code #20190304
        IF lo_column_define.cd_DEFINE_TYPE = cs_no_define THEN
          CALL sadzp000_msg_show_error('adz-00126', 'adz-00126', ls_replace_arg, 1)
          LET lo_t_dzeb_t[li_arr_curr].dzeb002 = ls_leading_code
          NEXT FIELD lbl_dzeb002
          CONTINUE DIALOG
        END IF  
      END IF
      
      #判斷當下欄位輸入欄位屬性是否存在
      LET ls_main_module = m_dzea_t[mi_dzea_index].dzea003
      CALL adzi140_check_column_attribute_valid(lo_t_dzeb_t[ARR_CURR()].*,ls_main_module,ms_dgenv,ls_old_dzeb006) RETURNING lb_result,ls_old_dzeb006,lo_t_dzeb_t[ARR_CURR()].*
      IF NOT lb_result THEN 
        NEXT FIELD lbl_dzeb006
        CONTINUE DIALOG
      END IF
      #20210222 end
      
      #20190219 begin 
      LET ls_new_dzeb006 = GET_FLDBUF(lbl_dzeb006)
      IF ls_new_dzeb006 <> ls_old_dzeb006 THEN
        LET ls_main_module = m_dzea_t[mi_dzea_index].dzea003
        CALL adzi140_check_column_attribute_valid(lo_t_dzeb_t[li_arr_curr].*,ls_main_module,ms_dgenv,ls_old_dzeb006) RETURNING lb_result,ls_old_dzeb006,lo_t_dzeb_t[li_arr_curr].*
        IF NOT lb_result THEN 
          NEXT FIELD lbl_dzeb006
          CONTINUE DIALOG
        END IF
      END IF  
      #20190219 end

      #20191004 begin
      LET ls_new_dzeb012 = GET_FLDBUF(lbl_dzeb012)
      IF NVL(ls_new_dzeb012,cs_null_value) <> NVL(ls_old_dzeb012,cs_null_value) THEN
        #進行一般資料檢核
        CALL sadzi140_util_check_default_value(lo_t_dzeb_t[li_arr_curr].dzeb007,lo_t_dzeb_t[li_arr_curr].dzeb008,lo_t_dzeb_t[li_arr_curr].dzeb012) RETURNING lb_result,ls_error_code
        IF NOT lb_result THEN
          CALL sadzp000_msg_show_error(ls_error_code, ls_error_code, NULL, 0)
          LET lo_t_dzeb_t[li_arr_curr].dzeb012 = ""
          NEXT FIELD lbl_dzeb012
          CONTINUE DIALOG
        ELSE #20200330
          CALL sadzi140_util_refresh_default_value(lo_t_dzeb_t[li_arr_curr].dzeb012) RETURNING lo_t_dzeb_t[li_arr_curr].dzeb012 #20200330      
        END IF
      END IF  
      #20191004 end
      
      CALL lo_collect_messages.clear() #20170523
      #讓當下所在Cursor的資料進入已異動
      LET lo_t_dzeb_t[li_arr_curr].RECORDTYPE = NVL(lo_t_dzeb_t[li_arr_curr].RECORDTYPE,cs_data_modified)
      LET lo_t_dzeb_t[li_arr_curr].lang_modified = cs_lang_modified

      {
      #避免資料沒有更新到當下欄位, 呼叫一次adzi140_on_change_dzeb006
      LET ls_new_dzeb006 = lo_t_dzeb_t[li_arr_curr].dzeb006
      IF NVL(ls_new_dzeb006,cs_null_value) <> NVL(ls_old_dzeb006,cs_null_value) THEN
        CALL adzi140_on_change_dzeb006(lo_t_dzeb_t[li_arr_curr].*) RETURNING lo_t_dzeb_t[li_arr_curr].*
        DISPLAY lo_t_dzeb_t[li_arr_curr].* TO sr_TableColumns[li_arr_curr].*
      END IF
      }

      FOR li_rec_cnt = 1 TO lo_t_dzeb_t.getLength()
        #判斷輸入欄位名稱是否已經存在資料庫
        LET ls_column_name = lo_t_dzeb_t[li_rec_cnt].dzeb002
        LET ls_column_name_orig = lo_t_dzeb_t[li_rec_cnt].dzeb002_orig
        IF NVL(ls_column_name_orig,cs_null_value) <> cs_null_value THEN
          IF ls_column_name <> ls_column_name_orig THEN
            CALL sadzp000_msg_show_error("adz-00211", "adz-00211", NULL, 1)
            LET lo_t_dzeb_t[ARR_CURR()].dzeb002 = ls_column_name_orig
            NEXT FIELD lbl_dzeb002
            CONTINUE DIALOG
          END IF
          {
          CALL sadzi140_db_check_db_column_exist_by_single_column(ls_master_user,ls_column_name_orig,ls_column_name) RETURNING lb_column_exist
          IF lb_column_exist THEN
            CALL sadzp000_msg_show_error("adz-00211", "adz-00211", NULL, 1)
            LET lo_t_dzeb_t[ARR_CURR()].dzeb002 = ls_column_name_orig
            NEXT FIELD lbl_dzeb002
            CONTINUE DIALOG
          END IF
          }
        END IF
        
      END FOR

      #數字類型欄位預設值預設為0, 沒有輸入要提示
      LET li_null_cnts = 1
      CALL lo_null_list.clear()
      FOR li_rec_cnt = 1 TO lo_t_dzeb_t.getLength()
        LET ls_dzeb012 = sadzi140_util_trim_str(lo_t_dzeb_t[li_rec_cnt].dzeb012)
        IF NVL(lo_t_dzeb_t[li_rec_cnt].dzeb022,cs_null_value) = cs_cdf_serial_number AND
           NVL(lo_t_dzeb_t[li_rec_cnt].dzeb006,cs_null_value) LIKE "N%" AND
           ((ls_dzeb012.getLength() = 0) OR ls_dzeb012 IS NULL) THEN
          LET lo_null_list[li_null_cnts] = lo_t_dzeb_t[li_rec_cnt].dzeb002
          LET li_null_cnts = li_null_cnts + 1  
        END IF   
      END FOR    
      
      LET li_rec_cnt = 1
      IF lo_null_list.getLength() > 0 THEN
        FOR li_null_cnts = 1 TO lo_null_list.getLength()
          LET ls_null_list = ls_null_list,lo_null_list[li_null_cnts],", "
          IF li_rec_cnt >= 5 THEN
            LET li_rec_cnt = 0
            LET ls_null_list = ls_null_list,"\n"   
          END IF
          LET li_rec_cnt = li_rec_cnt + 1
        END FOR 
        LET ls_replace_arg = "\n",ls_null_list,"|"
        #CALL sadzp000_msg_show_info('adz-00822', 'adz-00822', ls_replace_arg, 0) -- 20170523 mark
        CALL sadzp000_msg_set_collect_message('adz-00822',ls_replace_arg,cs_msg_work_type_add,ms_lang,lo_collect_messages) RETURNING lo_collect_messages -- 20170523
      END IF

      LET lb_error = FALSE
      LET ls_main_table  = m_dzea_t[mi_dzea_index].dzea001
      LET ls_main_module = m_dzea_t[mi_dzea_index].dzea003
      LET ls_table_type  = m_dzea_t[mi_dzea_index].dzea004 #20170523 
      LET ls_owner       = ms_master_user

      #若有做主鍵的勾選或取消, 則需要顯示提示訊息
      IF lb_key_checked THEN
        #CALL sadzp000_msg_show_info("adz-00215", "adz-00215", NULL, 1) -- 20170523 mark
        CALL sadzp000_msg_set_collect_message('adz-00215',NULL,cs_msg_work_type_add,ms_lang,lo_collect_messages) RETURNING lo_collect_messages -- 20170523
        CALL sadzi140_db_get_effect_table_list(ls_main_table,ms_lang) RETURNING ls_effect_table_list #20191209
        IF NVL(ls_effect_table_list.trim(),cs_null_value) <> cs_null_value THEN
          LET ls_replace_arg = ls_effect_table_list,"|"
          #CALL sadzp000_msg_show_info('adz-00749', 'adz-00749', ls_replace_arg, 0) -- 20170523 mark
          CALL sadzp000_msg_set_collect_message('adz-00749',ls_replace_arg,cs_msg_work_type_add,ms_lang,lo_collect_messages) RETURNING lo_collect_messages -- 20170523
        END IF
      END IF

      #若有刪除欄位, 則需要顯示提示訊息
      IF ls_del_list.getLength() > 1 THEN
        CALL sadzi140_util_split_string_to_group(ls_del_list,10,",") RETURNING ls_replace_arg
        LET ls_replace_arg = "\n",ls_replace_arg,"|"
        --CALL sadzp000_msg_show_info('adz-00781', 'adz-00781', ls_replace_arg, 0) -- 20170523 mark
        CALL sadzp000_msg_set_collect_message('adz-00781',ls_replace_arg,cs_msg_work_type_add,ms_lang,lo_collect_messages) RETURNING lo_collect_messages -- 20170523
        #客制環境才顯示實際表格上的欄位將會根據 r.t 上的資料進行異動
        IF (ms_DGENV = cs_dgenv_customize) THEN
          --CALL sadzp000_msg_show_info('adz-00848', 'adz-00848', NULL, 0) -- 20160510 --20170523 mark
          CALL sadzp000_msg_set_collect_message('adz-00848',NULL,cs_msg_work_type_add,ms_lang,lo_collect_messages) RETURNING lo_collect_messages -- 20170523
        END IF
      END IF

      #20180903 begin
      LET ls_module_name = m_dzea_t[mi_dzea_index].dzea003
      CALL adzi140_check_common_fields_if_exists(ls_main_table,ls_table_type,ls_module_name,lo_t_dzeb_t,lo_collect_messages) RETURNING lo_collect_messages
      #20180903 end

      {
      #20170523 begin
      IF lo_collect_messages.getLength() <> 0 THEN 
        CALL sadzp000_msg_show_collect_message(lo_collect_messages,ms_lang) 
      END IF
      #20170523 end
      }
      
      {
      #20170523 mark
      &ifndef DEBUG
      CALL cl_progress_bar(4) #20170214
      &endif
      }

      #20170523 begin
      FOR li_rec_cnt = 1 TO lo_t_dzeb_t.getLength()

        LET lo_t_dzeb_t[li_rec_cnt].dzeb001 = ls_main_table
        LET lo_t_dzeb_t[li_rec_cnt].dzeb021 = li_rec_cnt

        IF (NVL(lo_t_dzeb_t[li_rec_cnt].dzeb006,cs_null_value) = cs_null_value) AND (NVL(lo_t_dzeb_t[li_rec_cnt].dzeb002,cs_null_value) <> cs_null_value) THEN
          CALL DIALOG.setCurrentRow("sr_tablecolumns",li_rec_cnt)
          LET ls_replace_arg = lo_t_dzeb_t[li_rec_cnt].dzeb002,"|",lo_t_dzeb_t[li_rec_cnt].dzeb003,"|"
          CALL sadzp000_msg_show_error('adz-00096', 'adz-00096', ls_replace_arg, 1)
          NEXT FIELD lbl_dzeb006
          CONTINUE DIALOG
        END IF
          
      END FOR      
      #20170523 end
      
      #begin 處理新增及異動
      BEGIN WORK
      LET li_counts = 0
      LET ls_key_list = ""

      {
      #20170523 mark
      &ifndef DEBUG
      CALL cl_progress_ing('Preparing design data ...') #20170214
      &endif
      }
      
      CALL sadzp000_msg_show_message_bar('Preparing design data ...',NULL,NULL) -- 20170523
      
      FOR li_rec_cnt = 1 TO lo_t_dzeb_t.getLength()

        {
        #彙整 Key 清單
        IF NVL(lo_t_dzeb_t[li_rec_cnt].dzeb004,"N") = "Y" THEN
          LET ls_key_list = ls_key_list,",",lo_t_dzeb_t[li_rec_cnt].dzeb002
        END IF
        }

        #有標示異動/新增碼的資料列才更新
        IF (lo_t_dzeb_t[li_rec_cnt].recordtype = cs_data_modified) OR 
           (lo_t_dzeb_t[li_rec_cnt].recordtype = cs_data_added) OR 
           (sadzi140_util_trim_str(lo_t_dzeb_t[li_rec_cnt].dzeb022)) IS NULL THEN #20190304

          LET lo_t_dzeb_t[li_rec_cnt].dzeb001 = ls_main_table
          LET lo_t_dzeb_t[li_rec_cnt].dzeb021 = li_rec_cnt

          #20190304 begin
          LET ls_dzeb023 = NULL
          IF lo_t_dzeb_t[li_rec_cnt].dzeb022 = cs_cdf_serial_number OR sadzi140_util_trim_str(lo_t_dzeb_t[li_rec_cnt].dzeb022) IS NULL THEN
            LET ls_dzeb002 = lo_t_dzeb_t[li_rec_cnt].dzeb002
            CALL adzi140_get_column_numeric_digit(ls_dzeb002) RETURNING li_num_begin,li_num_end
            IF li_num_begin > 0 THEN 
              LET ls_dzeb023 = ls_dzeb002.subString(li_num_begin,li_num_end)
              LET lo_t_dzeb_t[li_rec_cnt].dzeb023 = ls_dzeb023
              IF lo_t_dzeb_t[li_rec_cnt].dzeb022 IS NULL THEN
                LET lo_t_dzeb_t[li_rec_cnt].dzeb022 = cs_cdf_serial_number 
              END IF
            END IF  
          END IF  
          LET ls_dzeb027 = lo_t_dzeb_t[li_rec_cnt].DZEB027 
          IF ls_dzeb027.getIndexOf("TOPIND",1) <= 0 THEN
            LET ls_custom_code = adzi140_get_custom_code(lo_t_dzeb_t[li_rec_cnt].dzeb001,lo_t_dzeb_t[li_rec_cnt].DZEB002,ms_column_user_define_code)
            IF ls_custom_code IS NULL OR ((ls_custom_code <> m_dzea_t[mi_dzea_index].dzea014) AND (ls_custom_code <> ms_topind)) THEN   
              LET ls_topind = m_dzea_t[mi_dzea_index].dzea014 
            ELSE
              LET ls_topind = ms_topind
            END IF
            CALL adzi140_get_dzeb027_json_data(ls_topind,NVL(ls_dzeb023,lo_t_dzeb_t[li_rec_cnt].DZEB023),ls_custom_code) RETURNING ls_dzeb027
            LET lo_t_dzeb_t[li_rec_cnt].DZEB027 = ls_dzeb027
          END IF
          #20190304 end

          #20170523 mark begin
          { 
          IF (NVL(lo_t_dzeb_t[li_rec_cnt].dzeb006,cs_null_value) = cs_null_value) AND (NVL(lo_t_dzeb_t[li_rec_cnt].dzeb002,cs_null_value) <> cs_null_value) THEN
            LET lb_error = TRUE
            CALL DIALOG.setCurrentRow("sr_tablecolumns",li_rec_cnt)
            #Error : 欄位沒有設定屬性值
            LET ls_replace_arg = lo_t_dzeb_t[li_rec_cnt].dzeb002,"|",lo_t_dzeb_t[li_rec_cnt].dzeb003,"|"
            CALL sadzp000_msg_show_error('adz-00096', 'adz-00096', ls_replace_arg, 1)

            ROLLBACK WORK
            BEGIN WORK
            EXIT FOR
          ELSE
          }
          #20170523 mark end
            #新增到 DZEB_T 中
            CALL sadzi140_crud_insert_update_dzeb_t(lo_t_dzeb_t[li_rec_cnt].*)
            #新增到欄位規格設計資料中
            CALL sadzi140_crud_insert_update_dzep_t(lo_t_dzeb_t[li_rec_cnt].*,m_common_fields,ms_lang)
            #新增到欄位參考設定中
            CALL sadzi140_crud_insert_update_dzef_t(lo_t_dzeb_t[li_rec_cnt].*,m_common_fields)
            #修改多語言資料
            CALL sadzi140_crud_insert_update_dzebl_t(lo_t_dzeb_t[li_rec_cnt].*,ms_lang)
            #新增狀態相關資料
            CALL sadzi140_crud_insert_update_gzcc_t(lo_t_dzeb_t[li_rec_cnt].*)
            #20160615 begin
            #新增 Log
            CALL adzi140_set_column_alter_log(lo_t_dzeb_t[li_rec_cnt].*,lo_t_dzeb_t[li_rec_cnt].recordtype)
            #20160615 end

            LET li_counts = li_counts + 1
            IF li_counts >= 100 THEN
              COMMIT WORK
              BEGIN WORK
              LET li_counts = 0
            END IF

          --END IF #20170523 mark
        END IF
      END FOR

      COMMIT WORK
      #end 處理新增及異動

      {
      #20170523 mark
      &ifndef DEBUG
      CALL cl_progress_ing('Processing deleted data ...') #20170214
      &endif
      }

      CALL sadzp000_msg_show_message_bar('Processing deleted data ...',NULL,NULL) -- 20170523
      
      #begin 處理刪除
      IF ls_del_list.getLength() > 1 THEN
        LET ls_del_list = ls_del_list.subString(2,ls_del_list.getLength())
        CALL adzi140_delete_column_design_datas(ls_main_table,ls_del_list) RETURNING lb_result #20180302

        {
        BEGIN WORK

        #刪除DZEB
        CALL adzi140_delete_dzeb_data(ls_main_table,ls_del_list) RETURNING lb_success
        IF NOT lb_success THEN
          ROLLBACK WORK
          BEGIN WORK
        END IF

        IF lb_success THEN
          #刪除DZEBL
          CALL adzi140_delete_dzebl_data(ls_main_table,ls_del_list) RETURNING lb_success
          IF NOT lb_success THEN
            ROLLBACK WORK
            BEGIN WORK
          END IF

          #刪除DZEP
          CALL adzi140_delete_dzep_data(ls_main_table,ls_del_list) RETURNING lb_success
          IF NOT lb_success THEN
            ROLLBACK WORK
            BEGIN WORK
          END IF
        END IF

        COMMIT WORK
        }
        
      END IF
      #end 處理刪除

      #判斷當為客制環境時,則需要異動模組別為 Cxx 模組.
      {
      #2014.11.18 為不異動模組別
      IF ms_dgenv = cs_dgenv_customize THEN
        LET ls_module_code = m_dzea_t[mi_dzea_index].dzea003
        #Check Column Data
        CALL sadzi140_db_check_column_if_customize(ls_main_table) RETURNING lb_result
        IF lb_result THEN
          CALL sadzi140_db_get_module_code_by_dgenv(ls_module_code,ms_dgenv) RETURNING ls_module_code
          CALL sadzi140_db_update_module_code(ls_main_table,ls_module_code,ms_dgenv)
        ELSE
          CALL sadzi140_db_get_module_code_by_dgenv(ls_module_code,cs_dgenv_standard) RETURNING ls_module_code
          CALL sadzi140_db_update_module_code(ls_main_table,ls_module_code,cs_dgenv_standard)
        END IF
      END IF
      }

      {
      #20170523 mark
      &ifndef DEBUG
      CALL cl_progress_ing('Verifying db columns with design data, please wait ...') #20170214
      &endif
      }
      
      CALL sadzp000_msg_show_message_bar('Verifying db columns with design data, please wait ...',NULL,NULL) -- 20170523
      
      #更新異動碼
      #CALL sadzi140_db_update_alter_code(ls_owner,ls_main_table) #20170523 mark
      CALL sadzi140_db_set_alter_code(ls_owner,ls_main_table,"N") #20170523 add

      {
      #20170523
      &ifndef DEBUG
      CALL cl_progress_ing('Finalizing ...') #20170214
      &endif
      }

      CALL sadzp000_msg_show_message_bar('Finalizing ...',NULL,NULL) -- 20170523
      
      IF NOT lb_error THEN
        #重Load一次表格欄位
        CALL sadzi140_db_get_column_with_sequence(ls_main_table) RETURNING lo_column_info
        #檢核並刪除無效的Key或索引資料
        CALL sadzi140_db_check_and_remove_invalid_key_index(lo_column_info,ls_main_table,ms_lang) RETURNING lb_result
        #建立表格的 Key 資料
        CALL sadzi140_db_set_constraint_data(ls_main_table) RETURNING lb_result
        #CALL sadzi140_db_set_key_data(ls_main_table,ls_key_list) RETURNING lb_result
        #CALL sadzi140_db_set_index_data(ls_main_table,ls_key_list) RETURNING lb_result
        CALL adzi140_fill_dzeb_t(ls_main_table)
        CALL adzi140_fill_dzed_t(ls_main_table)
        #20190619 begin
        #20190619 end
        CALL sadzp000_msg_show_message_bar(NULL,NULL,NULL) -- 20170523
        #EXIT DIALOG
      ELSE
        LET lb_error = FALSE
        #CONTINUE DIALOG
      END IF

      #20170523 begin
      IF lo_collect_messages.getLength() <> 0 THEN 
        CALL sadzp000_msg_show_collect_message(lo_collect_messages,ms_lang) 
      END IF
      #20170523 end

      IF NOT lb_error THEN 
        EXIT DIALOG
      ELSE
        CONTINUE DIALOG
      END IF    
      
    ON ACTION CLOSE
      EXIT DIALOG

  END DIALOG

  LET p_dialog = lo_dialog #20170214
  
  CALL adzi140_set_editor_mode(p_dialog,cs_editor_mode_browse)

  LET lb_return = IIF(lb_schema_altered==TRUE,lb_schema_altered,lb_data_altered) #20180302

  RETURN lb_return,lb_dzeb006_changed,lb_dzeb003_changed #20180412

END FUNCTION

FUNCTION adzi140_on_change_dzeb006(p_t_dzeb_t)
DEFINE
  p_t_dzeb_t  T_DZEB_T
DEFINE
  ls_replace_arg STRING,
  ls_type        STRING,
  ls_default     STRING,
  lo_t_dzeb_t    T_DZEB_T,
  lo_property    T_PROPERTY

  LET lo_t_dzeb_t.* = p_t_dzeb_t.*

  CALL sadzi140_util_get_property(lo_t_dzeb_t.dzeb006,ms_lang) RETURNING lo_property.*

  IF NVL(lo_property.D_TYPE,cs_null_value) = cs_null_value THEN
    INITIALIZE lo_property TO NULL
    IF lo_t_dzeb_t.dzeb006 IS NOT NULL THEN 
      LET lo_t_dzeb_t.dzeb006 = ""
      LET ls_replace_arg = ""
      CALL sadzp000_msg_show_error('adz-00231', 'adz-00231', ls_replace_arg, 1)
    END IF  
  END IF

  LET lo_t_dzeb_t.dzeb006_DESC = lo_property.D_DESC
  LET lo_t_dzeb_t.dzeb007      = lo_property.D_TYPE
  LET lo_t_dzeb_t.dzeb008      = lo_property.D_LENGTH
  #LET lo_t_dzeb_t.dzeb009      = lo_property.D_WIDGET_WIDTH
  #LET lo_t_dzeb_t.dzeb010      = lo_property.D_WIDGETS
  LET lo_t_dzeb_t.dzeb012      = NVL(lo_t_dzeb_t.dzeb012,lo_property.D_DEFAULT) --20211112 fix
  #LET lo_t_dzeb_t.dzeb015      = lo_property.D_PERCENT
  #LET lo_t_dzeb_t.dzeb016      = lo_property.D_FORMAT
  LET lo_t_dzeb_t.dzeb029      = ms_dgenv
  IF (lo_t_dzeb_t.dzeb030) IS NULL THEN
    LET lo_t_dzeb_t.dzeb030 = ms_dgenv
  END IF
  IF (lo_t_dzeb_t.dzeb031) IS NULL THEN
    LET lo_t_dzeb_t.dzeb031 = "N"
  END IF

  RETURN lo_t_dzeb_t.*

END FUNCTION

#建立 Index 的資料
FUNCTION adzi140_alter_index_data(p_dialog)
DEFINE
  p_dialog ui.dialog
DEFINE
  lo_t_dzec_t    DYNAMIC ARRAY OF T_DZEC_T,
  lo_del_list    DYNAMIC ARRAY OF STRING,
  lo_left_array   DYNAMIC ARRAY OF T_COLUMN_LIST,
  lo_right_array  DYNAMIC ARRAY OF T_COLUMN_LIST,
  lo_column_array DYNAMIC ARRAY OF T_COLUMN_LIST, #20160614
  lo_INDEX_IF_EXCEED  T_KEY_IF_EXCEED, #20160616
  lo_pick_col_arr DYNAMIC ARRAY OF T_COLUMN_LIST, #20171120
  li_rec_cnt     INTEGER,
  li_counts      INTEGER,
  ls_table_name  STRING,
  ls_index_list  STRING,
  ls_index_type  STRING,
  ls_old_index_type  STRING,
  lb_error       BOOLEAN,
  ls_schema_sql  STRING,
  ls_main_table  STRING,
  ls_main_module STRING,
  ls_owner       STRING,
  ls_del_list    STRING,
  li_arr_curr    INTEGER,
  ls_columns     STRING,
  ls_column_queue STRING,
  ls_identify_name STRING,
  ls_identify_col_name STRING,
  ls_replace_arg STRING,
  lb_success     BOOLEAN,
  lb_active      BOOLEAN,
  ls_leading_code STRING,
  ls_index_name_keyword STRING,
  ls_index_name_orig STRING,
  ls_index_name  STRING,
  ls_module_code STRING,
  ls_ship_notice STRING,
  lb_result      BOOLEAN,
  li_loop        INTEGER,
  ls_lang        STRING,
  ls_index_leading STRING,
  lb_data_altered BOOLEAN,
  lo_lang_arr     DYNAMIC ARRAY OF T_LANGUAGE_TYPE
DEFINE
  lb_return  BOOLEAN

  LET lb_error = FALSE
  LET lb_data_altered = FALSE 

  CALL adzi140_set_editor_mode(p_dialog,cs_editor_mode_edit)

  DIALOG

    INPUT ARRAY lo_t_dzec_t FROM sr_TableIndex.* ATTRIBUTE(WITHOUT DEFAULTS)

      BEFORE INPUT
        LET lo_t_dzec_t.* = m_dzec_t.*
        CALL DIALOG.setActionHidden("insert",TRUE)
        CALL DIALOG.setActionHidden("append",TRUE)
        CALL DIALOG.setActionHidden("delete",TRUE)
        CALL DIALOG.setCurrentRow("sr_tableindex",mi_dzec_index)

      BEFORE ROW
        LET li_arr_curr = ARR_CURR()
        LET ls_index_type = lo_t_dzec_t[li_arr_curr].dzec003
        LET ls_index_name = lo_t_dzec_t[li_arr_curr].dzec002
        LET ls_old_index_type = ls_index_type

        #20160826 Add begin
        #控管已出貨的Index或者為PK的Index不能刪除及異動
        IF (NVL(lo_t_dzec_t[li_arr_curr].dzec008,cs_null_value) = "Y") OR
           (NVL(lo_t_dzec_t[li_arr_curr].dzec007,ms_dgenv) <> ms_dgenv) OR
           ((ls_index_type == "U") AND (ls_index_name.subString(ls_index_name.getLength()-cs_pk_keyword.getLength()+1,ls_index_name.getLength()) = cs_pk_keyword)) THEN --20180313 --(ls_index_name.getIndexOf(cs_pk_keyword,1) > 0)) THEN
          CALL DIALOG.setActionActive("delete", FALSE)
          LET lb_active = FALSE
        ELSE
          CALL DIALOG.setActionActive("delete", TRUE)
          LET lb_active = TRUE
        END IF
        CALL DIALOG.setFieldActive("sr_tableindex.lbl_dzec002", lb_active)
        CALL DIALOG.setFieldActive("sr_tableindex.lbl_dzec004", lb_active)
        #20160826 Add end

        #20160826 Mark begin
        {
        #控管PK的Index不能異動
        LET lb_active = IIF(((ls_index_type == "U") AND ((ls_index_name.subString(ls_index_name.getLength()-cs_pk_keyword.getLength()+1,ls_index_name.getLength()) = cs_pk_keyword))),FALSE,TRUE) --20180313
        CALL DIALOG.setFieldActive("sr_tableindex.lbl_dzec002", lb_active)
        CALL DIALOG.setFieldActive("sr_tableindex.lbl_dzec004", lb_active)
        }
        #20160826 Mark end
        
      AFTER ROW
        LET li_arr_curr = ARR_CURR()
        LET ls_index_name = lo_t_dzec_t[li_arr_curr].dzec002
        FOR li_rec_cnt = 1 TO lo_t_dzec_t.getLength()
          IF li_arr_curr <> li_rec_cnt THEN
            IF lo_t_dzec_t[li_rec_cnt].dzec002 = ls_index_name THEN
              CALL sadzp000_msg_show_error("adz-00198", "adz-00198", NULL, 1)
              NEXT FIELD lbl_dzec002
              CONTINUE DIALOG
            END IF
          END IF
        END FOR

      ON CHANGE lbl_dzec002
        LET ls_index_name = lo_t_dzec_t[ARR_CURR()].dzec002
        LET ls_index_name_orig = lo_t_dzec_t[ARR_CURR()].dzec002_orig

        #判斷輸入名稱是否合乎命名原則
        LET ls_table_name = m_dzea_t[mi_dzea_index].dzea001
        CALL sadzi140_util_get_table_leading_code(ls_table_name) RETURNING ls_leading_code
        IF m_dzea_t[mi_dzea_index].dzea018 = cs_dgenv_standard THEN
          LET ls_index_leading = ms_index_user_alter_leading_code,ls_leading_code
        ELSE
          LET ls_index_leading = ls_leading_code
        END IF
        IF (ls_index_name.getIndexOf(ls_index_leading||cs_idx_normal_keyword,1) = 0) AND
           (ls_index_name.getIndexOf(ls_index_leading||cs_idx_unique_keyword,1) = 0) AND
           (ls_index_name IS NOT NULL) THEN
          CALL sadzp000_msg_show_error('adz-00126', 'adz-00126', ls_replace_arg, 1)
          #預設給 Footing Data
          LET lo_t_dzec_t[ARR_CURR()].dzec002 = ls_index_leading,cs_idx_normal_keyword
          LET lo_t_dzec_t[ARR_CURR()].dzec003 = "N"
          NEXT FIELD lbl_dzec002
          CONTINUE DIALOG
        END IF

        IF NVL(ls_index_name_orig,cs_null_value) <> cs_null_value THEN
          IF ls_index_name <> ls_index_name_orig THEN
            CALL sadzp000_msg_show_error("adz-00213", "adz-00213", NULL, 1)
            LET lo_t_dzec_t[ARR_CURR()].dzec002 = ls_index_name_orig
            NEXT FIELD lbl_dzec002
            CONTINUE DIALOG
          END IF
        END IF

        LET li_arr_curr = ARR_CURR()
        LET ls_identify_name = lo_t_dzec_t[li_arr_curr].dzec002
        IF adzi140_check_key_index_name_exist(ls_identify_name) THEN
          #LET lo_t_dzec_t[li_arr_curr].dzec002 = ""
          LET ls_replace_arg = "|"
          CALL sadzp000_msg_show_error('adz-00198', 'adz-00198', ls_replace_arg, 1)
          NEXT FIELD lbl_dzec002
        END IF

        LET lb_data_altered = TRUE

      ON CHANGE lbl_dzec003
        LET li_arr_curr = ARR_CURR()

        #檢核是否可異動索引屬性
        LET ls_ship_notice = lo_t_dzec_t[li_arr_curr].dzec008
        CALL adzi140_check_attribute_could_modify(ls_ship_notice,ms_dgenv) RETURNING lb_result
        IF NOT lb_result THEN
          LET lo_t_dzec_t[li_arr_curr].dzec003 = ls_old_index_type
          CONTINUE DIALOG
        END IF

        LET ls_index_name = lo_t_dzec_t[li_arr_curr].dzec002
        IF (ls_old_index_type = "U") AND (ls_index_name.subString(ls_index_name.getLength()-cs_pk_keyword.getLength()+1,ls_index_name.getLength()) = cs_pk_keyword) THEN --20180313 -- (ls_index_name.getIndexOf(cs_pk_keyword,1) > 0) THEN
          LET lo_t_dzec_t[li_arr_curr].dzec003 = ls_old_index_type
          LET ls_replace_arg = "|"
          CALL sadzp000_msg_show_error('adz-00380', 'adz-00380', ls_replace_arg, 1)
        ELSE
          #Index 給預設名稱
          IF NVL(lo_t_dzec_t[li_arr_curr].dzec002,cs_null_value) = cs_null_value THEN
            LET ls_table_name = m_dzea_t[mi_dzea_index].dzea001
            CALL sadzi140_util_get_table_leading_code(ls_table_name) RETURNING ls_leading_code
            CASE lo_t_dzec_t[li_arr_curr].dzec003
              WHEN "U"
                LET ls_index_name_keyword = cs_idx_unique_keyword
              WHEN "N"
                LET ls_index_name_keyword = cs_idx_normal_keyword
            END CASE
            IF m_dzea_t[mi_dzea_index].dzea018 = cs_dgenv_standard THEN
              #LET lo_t_dzec_t[li_arr_curr].dzec002 = ls_leading_code,ms_table_user_custom_code,ls_index_name_keyword
              #LET lo_t_dzec_t[li_arr_curr].dzec002 = ms_index_user_alter_leading_code,ls_leading_code,ms_column_user_alter_code,ls_index_name_keyword
              LET lo_t_dzec_t[li_arr_curr].dzec002 = ms_index_user_alter_leading_code,ls_leading_code,ls_index_name_keyword
            ELSE
              #LET lo_t_dzec_t[li_arr_curr].dzec002 = ls_leading_code,ls_index_name_keyword
              LET lo_t_dzec_t[li_arr_curr].dzec002 = ls_leading_code,ls_index_name_keyword
            END IF
          END IF
        END IF
        LET lb_data_altered = TRUE

      ON ACTION controlp INFIELD lbl_dzec004
        LET li_arr_curr   = ARR_CURR()
        LET ls_table_name = m_dzea_t[mi_dzea_index].dzea001
        LET ls_columns    = lo_t_dzec_t[li_arr_curr].dzec004
        CALL adzi140_decode_columns(ls_columns) RETURNING ls_column_queue
        CALL adzi140_get_table_column_list(ls_table_name,ls_column_queue,cs_side_left) RETURNING lo_left_array
        CALL adzi140_get_table_column_list(ls_table_name,ls_column_queue,cs_side_right) RETURNING lo_right_array
        CALL sadzi140_cols_run(lo_left_array,lo_right_array) RETURNING ls_columns,lo_pick_col_arr #20171120
        IF NVL(ls_columns,cs_null_value) <> cs_result_cancel THEN
          LET lo_t_dzec_t[li_arr_curr].dzec004 = ls_columns
          LET lb_data_altered = TRUE
          GOTO _on_ahange_lbl_dzec004_INFIELD #20160614 
        END IF

      #20160614 begin  
      AFTER FIELD lbl_dzec004
        --GOTO _on_ahange_lbl_dzec004  #20170912 marked
      ON CHANGE lbl_dzec004
        LABEL _on_ahange_lbl_dzec004:
        LET li_arr_curr = ARR_CURR()
        LABEL _on_ahange_lbl_dzec004_INFIELD:
        
        LET ls_table_name = m_dzea_t[mi_dzea_index].dzea001
        LET ls_columns    = lo_t_dzec_t[li_arr_curr].dzec004

        IF ls_columns.trim() IS NOT NULL THEN 
          INITIALIZE lo_column_array TO NULL
          CALL adzi140_get_input_column_list(ls_columns) RETURNING lo_column_array
          FOR li_rec_cnt = 1 TO lo_column_array.getLength()
            CALL sadzi140_util_check_table_column_if_exists(ls_table_name,lo_column_array[li_rec_cnt].COLUMN_NAME) RETURNING lb_result
            IF NOT lb_result THEN
              LET ls_replace_arg = lo_column_array[li_rec_cnt].COLUMN_NAME,"|",ls_table_name,"|"
              CALL sadzp000_msg_show_error("adz-00869", "adz-00869", ls_replace_arg, 1) #欄位 xxxx 不存在r.t資料表設計器中
              LET lo_t_dzec_t[li_arr_curr].dzec004 = ""
              NEXT FIELD lbl_dzec004
              CONTINUE DIALOG
            END IF
          END FOR

          #20160616 begin
          CALL sadzi140_db_check_index_length_if_exceed(m_dzeb_t,lo_column_array) RETURNING lo_INDEX_IF_EXCEED.*
          IF lo_INDEX_IF_EXCEED.kie_result THEN
            LET ls_replace_arg = lo_INDEX_IF_EXCEED.kie_user_def,"|",lo_INDEX_IF_EXCEED.kie_sys_def,"|"
            CALL sadzp000_msg_show_error("adz-00877", "adz-00877", ls_replace_arg, 1) #您所設定的主鍵或索引總長度 (%1) 超出系統設定值 (%2)
            LET lo_t_dzec_t[li_arr_curr].dzec004 = ""
            NEXT FIELD lbl_dzec004
            CONTINUE DIALOG 
          END IF 
          #20160616 end

          #20170912 begin
          LET lo_t_dzec_t[li_arr_curr].dzec001 = m_dzea_t[mi_dzea_index].dzea001
          IF adzi140_check_if_duplicate_index_columns(lo_t_dzec_t[li_arr_curr].*) THEN 
            LET ls_replace_arg = ""
            CALL sadzp000_msg_show_error("adz-01242", "adz-01242", ls_replace_arg, 1) #您所挑選或輸入的索引欄位在另外一個索引已經存在, 請確認是否正確.
            LET lo_t_dzec_t[li_arr_curr].dzec004 = ""
            NEXT FIELD lbl_dzec004
            CONTINUE DIALOG 
          END IF
          #20170912 end
          
          LET lb_data_altered = TRUE
        END IF  
      #20160614 end
      
      ON ACTION DELETE
        LET li_arr_curr = ARR_CURR()
        LET ls_index_type = lo_t_dzec_t[li_arr_curr].dzec003
        LET ls_index_name = lo_t_dzec_t[li_arr_curr].dzec002
        IF ls_index_type = "U" AND (ls_index_name.subString(ls_index_name.getLength()-cs_pk_keyword.getLength()+1,ls_index_name.getLength()) = cs_pk_keyword) THEN --20180313 --ls_index_name.getIndexOf(cs_pk_keyword,1) > 0 THEN
          CALL sadzp000_msg_show_error("adz-00379", "adz-00379", NULL, 1)
          CONTINUE DIALOG
        END IF
        LET ls_del_list = ls_del_list,",'",lo_t_dzec_t[li_arr_curr].dzec002,"'"
        CALL lo_del_list.appendElement()
        LET lo_del_list[lo_del_list.getLength()] = lo_t_dzec_t[li_arr_curr].dzec002
        CALL DIALOG.deleteRow("sr_tableindex",li_arr_curr)
        LET lb_data_altered = TRUE

    END INPUT

    ON ACTION CANCEL
      LET lb_data_altered = FALSE
      CALL adzi140_fill_dzec_t(m_dzea_t[mi_dzea_index].dzea001)
      EXIT DIALOG

    ON ACTION ACCEPT
      #做存檔前的驗證
      FOR li_rec_cnt = 1 TO lo_t_dzec_t.getLength()
        LET ls_index_name = lo_t_dzec_t[li_rec_cnt].dzec002
        LET ls_index_name_orig = lo_t_dzec_t[li_rec_cnt].dzec002_orig
        IF NVL(ls_index_name_orig,cs_null_value) <> cs_null_value THEN
          IF ls_index_name <> ls_index_name_orig THEN
            CALL sadzp000_msg_show_error("adz-00213", "adz-00213", NULL, 1)
            LET lo_t_dzec_t[li_rec_cnt].dzec002 = ls_index_name_orig
            NEXT FIELD lbl_dzec002
            CONTINUE DIALOG
          END IF
        END IF
        LET ls_identify_name = lo_t_dzec_t[li_rec_cnt].dzec002
        LET ls_index_type = lo_t_dzec_t[li_rec_cnt].dzec003
        IF (NVL(ls_identify_name,cs_null_value) = cs_null_value) AND (NVL(ls_index_type,cs_null_value) <> cs_null_value) THEN
          LET ls_replace_arg = "|"
          CALL sadzp000_msg_show_error('adz-00201', 'adz-00201', ls_replace_arg, 1)
          CALL DIALOG.setCurrentRow("sr_tableindex", li_rec_cnt)
          NEXT FIELD lbl_dzec002
          CONTINUE DIALOG
        END IF
        LET ls_identify_col_name = lo_t_dzec_t[li_rec_cnt].dzec004
        IF NVL(ls_identify_name,cs_null_value) <> cs_null_value AND NVL(ls_identify_col_name,cs_null_value) = cs_null_value THEN
          LET ls_replace_arg = "|"
          CALL sadzp000_msg_show_error('adz-00202', 'adz-00202', ls_replace_arg, 1)
          CALL DIALOG.setCurrentRow("sr_tableindex", li_rec_cnt)
          NEXT FIELD lbl_dzec004
          CONTINUE DIALOG
        END IF
      END FOR

      LET lb_error = FALSE
      LET ls_main_table  = m_dzea_t[mi_dzea_index].dzea001
      LET ls_main_module = m_dzea_t[mi_dzea_index].dzea003
      LET ls_owner       = ms_master_user

      #begin 處理新增及異動
      BEGIN WORK
      LET li_counts = 0
      FOR li_rec_cnt = 1 TO lo_t_dzec_t.getLength()

        LET lo_t_dzec_t[li_rec_cnt].dzec001 = ls_main_table
        LET lo_t_dzec_t[li_rec_cnt].dzec006 = ms_dgenv
        IF lo_t_dzec_t[li_rec_cnt].dzec007 IS NULL THEN
          LET lo_t_dzec_t[li_rec_cnt].dzec007 = ms_dgenv
        END IF
        IF lo_t_dzec_t[li_rec_cnt].dzec008 IS NULL THEN
          LET lo_t_dzec_t[li_rec_cnt].dzec008 = "N"
        END IF
        LET lo_t_dzec_t[li_rec_cnt].dzecstus = "Y"  #20170912
        CALL sadzi140_crud_insert_update_dzec_t(lo_t_dzec_t[li_rec_cnt].*)

        LET li_counts = li_counts + 1
        IF li_counts >= 100 THEN
          COMMIT WORK
          BEGIN WORK
          LET li_counts = 0
        END IF

        IF NOT lb_error THEN
        END IF

      END FOR
      COMMIT WORK
      #end 處理新增及異動

      #begin 處理刪除
      IF ls_del_list.getLength() > 1 THEN
        BEGIN WORK

        LET ls_del_list = ls_del_list.subString(2,ls_del_list.getLength())

        #刪除DZEC
        CALL adzi140_delete_dzec_data(ls_main_table,ls_del_list) RETURNING lb_success
        IF NOT lb_success THEN
          ROLLBACK WORK
          BEGIN WORK
        END IF

        COMMIT WORK
      END IF
      #end 處理刪除

      #判斷當為客制環境時,則需要異動模組別為 Cxx 模組.
      {
      #2014.11.18 為不異動模組別
      IF ms_dgenv = cs_dgenv_customize THEN
        LET ls_module_code = m_dzea_t[mi_dzea_index].dzea003
        #Check Key Data
        CALL sadzi140_db_check_index_if_customize(ls_main_table) RETURNING lb_result
        IF lb_result THEN
          CALL sadzi140_db_get_module_code_by_dgenv(ls_module_code,ms_dgenv) RETURNING ls_module_code
          CALL sadzi140_db_update_module_code(ls_main_table,ls_module_code,ms_dgenv)
        ELSE
          CALL sadzi140_db_get_module_code_by_dgenv(ls_module_code,cs_dgenv_standard) RETURNING ls_module_code
          CALL sadzi140_db_update_module_code(ls_main_table,ls_module_code,cs_dgenv_standard)
        END IF
      END IF
      }

      #CALL sadzi140_db_update_alter_code(ls_owner,ls_main_table) #20170523 mark
      CALL sadzi140_db_set_alter_code(ls_owner,ls_main_table,"N") #20170523 add
      {
      #產出 table.sch 檔
      #CALL sadzi140_xml_get_lang_type_list() RETURNING lo_lang_arr
      CALL sadzi140_crud_get_static_lang_list() RETURNING lo_lang_arr
      LET lb_gen_tbl = TRUE #20170908
      LET ls_gen_list = "" #20170908
      FOR li_loop = 1 TO lo_lang_arr.getLength()
        LET ls_lang = lo_lang_arr[li_loop]
        CALL sadzi140_xml_gen_table_datas(ls_main_table,NVL(ls_module_code,ls_main_module),ls_lang,ls_gen_list) RETURNING lb_gen_tbl,ls_gen_list #20170908
        CALL sadzi140_db_gen_table_schema(ls_main_table,NVL(ls_module_code,ls_main_module),FALSE,ls_lang) RETURNING ls_schema_sql
      END FOR
      }
      CALL adzi140_fill_dzec_t(ls_main_table)
      EXIT DIALOG

    ON ACTION CLOSE
      EXIT DIALOG

  END DIALOG

  CALL adzi140_set_editor_mode(p_dialog,cs_editor_mode_browse)

  LET lb_return = lb_data_altered

  RETURN lb_return

END FUNCTION

#異動 Key 的資料
FUNCTION adzi140_alter_key_data(p_dialog)
DEFINE
  p_dialog ui.dialog
DEFINE
  lo_t_dzed_t     DYNAMIC ARRAY OF T_DZED_T,
  lo_del_list     DYNAMIC ARRAY OF STRING,
  lo_left_array   DYNAMIC ARRAY OF T_COLUMN_LIST,
  lo_right_array  DYNAMIC ARRAY OF T_COLUMN_LIST,
  lo_column_array DYNAMIC ARRAY OF T_COLUMN_LIST, #20160614
  lo_pick_col_arr DYNAMIC ARRAY OF T_COLUMN_LIST, #20171120
  li_rec_cnt      INTEGER,
  li_counts       INTEGER,
  li_arr_curr     INTEGER,
  ls_key_list     VARCHAR(1024),
  ls_schema_sql   STRING,
  ls_replace_arg  STRING,
  ls_main_table   STRING,
  ls_main_module  STRING,
  ls_owner        STRING,
  ls_del_list     STRING,
  ls_current_item STRING,
  ls_columns      STRING,
  ls_column_queue STRING,
  ls_table_name   STRING,
  ls_dzed005      STRING,
  ls_if_foreign   STRING,
  ls_identify_name STRING,
  ls_key_type     STRING,
  ls_old_key_type STRING,
  ls_identify_col_name STRING,
  lb_active       BOOLEAN,
  lb_success      BOOLEAN,
  lb_error        BOOLEAN,
  ls_key_name_orig STRING,
  ls_key_name     STRING,
  ls_leading_code STRING,
  ls_module_code  STRING,
  ls_ship_notice  STRING,
  lb_result       BOOLEAN,
  li_foreign_cnts INTEGER,
  li_loop         INTEGER,
  ls_lang         STRING,
  ls_key_leading  STRING,
  ls_fd_table_name STRING,
  lb_data_altered BOOLEAN,
  lb_add_new_footing_data BOOLEAN,
  lo_lang_arr     DYNAMIC ARRAY OF T_LANGUAGE_TYPE
DEFINE
  lb_return BOOLEAN

  LET lb_success = TRUE
  LET lb_data_altered = FALSE
  LET lb_add_new_footing_data = FALSE

  CALL adzi140_set_editor_mode(p_dialog,cs_editor_mode_edit)

  DIALOG

    INPUT ARRAY lo_t_dzed_t FROM sr_TableKey.* ATTRIBUTE(WITHOUT DEFAULTS)

      BEFORE INPUT
        LET lo_t_dzed_t.* = m_dzed_t.*
        CALL DIALOG.setActionHidden("insert",TRUE)
        CALL DIALOG.setActionHidden("append",TRUE)
        CALL DIALOG.setActionHidden("delete",TRUE)
        CALL DIALOG.setCurrentRow("sr_tablekey",mi_dzed_index)

      BEFORE ROW
        LET li_arr_curr = ARR_CURR()
        LET ls_key_type = lo_t_dzed_t[li_arr_curr].dzed003
        LET ls_old_key_type = ls_key_type

        IF (NVL(lo_t_dzed_t[li_arr_curr].dzed010,cs_null_value) = "Y") OR
           NVL(lo_t_dzed_t[li_arr_curr].dzed009,ms_dgenv) <> ms_dgenv THEN
          CALL DIALOG.setActionActive("delete", FALSE)
          LET lb_active = FALSE
          CALL DIALOG.setFieldActive("sr_tablekey.lbl_dzed002", lb_active)
          CALL DIALOG.setFieldActive("sr_tablekey.lbl_dzed005", lb_active)
          LET lb_active = IIF(ls_key_type == "P",FALSE,TRUE) -- 2022xxxx
          CALL DIALOG.setFieldActive("sr_tablekey.lbl_dzed004", lb_active) 
          CALL DIALOG.setFieldActive("sr_tablekey.lbl_dzed006", lb_active) 
        ELSE
          CALL DIALOG.setActionActive("delete", TRUE)
          LET lb_active = IIF(((ls_key_type == "F") OR (ls_key_type == "R")),TRUE,FALSE)
          CALL DIALOG.setFieldActive("sr_tablekey.lbl_dzed005", lb_active)
          CALL DIALOG.setFieldActive("sr_tablekey.lbl_dzed006", lb_active) 
          LET lb_active = IIF(ls_key_type == "P",FALSE,TRUE)
          CALL DIALOG.setFieldActive("sr_tablekey.lbl_dzed002", lb_active)
          CALL DIALOG.setFieldActive("sr_tablekey.lbl_dzed004", lb_active) 
        END IF

      AFTER ROW
        LET li_arr_curr = ARR_CURR()
        LET ls_key_name = lo_t_dzed_t[li_arr_curr].dzed002
        FOR li_rec_cnt = 1 TO lo_t_dzed_t.getLength()
          IF li_arr_curr <> li_rec_cnt THEN
            IF lo_t_dzed_t[li_rec_cnt].dzed002 = ls_key_name THEN
              CALL sadzp000_msg_show_error("adz-00198", "adz-00198", NULL, 1)
              NEXT FIELD lbl_dzed002
              CONTINUE DIALOG
            END IF
          END IF
        END FOR

      ON ACTION controlp INFIELD lbl_dzed004
        LET li_arr_curr   = ARR_CURR()
        LET ls_table_name = m_dzea_t[mi_dzea_index].dzea001
        LET ls_columns    = lo_t_dzed_t[li_arr_curr].dzed004
        CALL adzi140_decode_columns(ls_columns) RETURNING ls_column_queue
        CALL adzi140_get_table_column_list(ls_table_name,ls_column_queue,cs_side_left) RETURNING lo_left_array
        CALL adzi140_get_table_column_list(ls_table_name,ls_column_queue,cs_side_right) RETURNING lo_right_array
        CALL sadzi140_cols_run(lo_left_array,lo_right_array) RETURNING ls_columns,lo_pick_col_arr #20171120
        IF NVL(ls_columns,cs_null_value) <> cs_result_cancel THEN
          LET lo_t_dzed_t[li_arr_curr].dzed004 = ls_columns
          LET lo_t_dzed_t[li_arr_curr].recordtype = cs_data_modified  #2022xxxx
          LET lb_data_altered = TRUE
        END IF

      ON ACTION controlp INFIELD lbl_dzed005
        LET li_arr_curr = ARR_CURR()
        LET ls_dzed005 = lo_t_dzed_t[li_arr_curr].dzed005
        &ifndef DEBUG
        INITIALIZE g_qryparam.* TO NULL
        LET g_qryparam.state = 'i'
        LET g_qryparam.reqry = FALSE
        CALL q_dzea002()
        LET ls_dzed005 = NVL(g_qryparam.return1,ls_dzed005)
        &endif
        LET lo_t_dzed_t[li_arr_curr].dzed005 = ls_dzed005
        LET lo_t_dzed_t[li_arr_curr].recordtype = cs_data_modified  #2022xxxx
        LET lb_data_altered = TRUE

      ON ACTION controlp INFIELD lbl_dzed006
        LET li_arr_curr   = ARR_CURR()
        LET ls_table_name = lo_t_dzed_t[li_arr_curr].dzed005
        LET ls_columns    = lo_t_dzed_t[li_arr_curr].dzed006
        CALL adzi140_decode_columns(ls_columns) RETURNING ls_column_queue
        CALL adzi140_get_table_column_list(ls_table_name,ls_column_queue,cs_side_left) RETURNING lo_left_array
        CALL adzi140_get_table_column_list(ls_table_name,ls_column_queue,cs_side_right) RETURNING lo_right_array
        CALL sadzi140_cols_run(lo_left_array,lo_right_array) RETURNING ls_columns,lo_pick_col_arr #20171120
        IF NVL(ls_columns,cs_null_value) <> cs_result_cancel THEN
          LET lo_t_dzed_t[li_arr_curr].dzed006 = ls_columns
          LET lo_t_dzed_t[li_arr_curr].recordtype = cs_data_modified  #2022xxxx
          LET lb_data_altered = TRUE
        END IF

      ON ACTION DELETE
        LET ls_key_type = lo_t_dzed_t[li_arr_curr].dzed003
        IF ls_key_type = "P" THEN
          CALL sadzp000_msg_show_error("adz-00214", "adz-00214", NULL, 1)
          CONTINUE DIALOG
        END IF
        LET ls_del_list = ls_del_list,",'",lo_t_dzed_t[ARR_CURR()].dzed002,"'"
        CALL lo_del_list.appendElement()
        LET lo_del_list[lo_del_list.getLength()] = lo_t_dzed_t[ARR_CURR()].dzed002
        CALL DIALOG.deleteRow("sr_tablekey",ARR_CURR())
        LET lb_data_altered = TRUE

      ON CHANGE lbl_dzed002
        LET ls_key_name = lo_t_dzed_t[ARR_CURR()].dzed002
        LET ls_key_name_orig = lo_t_dzed_t[ARR_CURR()].dzed002_orig

        #判斷輸入名稱是否合乎命名原則
        LET ls_table_name = m_dzea_t[mi_dzea_index].dzea001
        CALL sadzi140_util_get_table_leading_code(ls_table_name) RETURNING ls_leading_code
        IF m_dzea_t[mi_dzea_index].dzea018 = cs_dgenv_standard THEN
          LET ls_key_leading = ms_key_user_alter_leading_code,ls_leading_code
        ELSE
          LET ls_key_leading = ls_leading_code
        END IF
        IF (ls_key_name.getIndexOf(ls_key_leading||cs_fk_keyword,1) = 0) AND
           (ls_key_name.getIndexOf(ls_key_leading||cs_fd_keyword,1) = 0) AND
           (ls_key_name.getIndexOf(ls_key_leading||cs_rk_keyword,1) = 0) AND
           (ls_key_name.getIndexOf(ls_key_leading||cs_pk_keyword,1) = 0) AND
           (ls_key_name IS NOT NULL) THEN
          CALL sadzp000_msg_show_error('adz-00126', 'adz-00126', ls_replace_arg, 1)
          #預設給 Footing Data
          LET lo_t_dzed_t[ARR_CURR()].dzed002 = ls_key_leading,cs_fd_keyword
          LET lo_t_dzed_t[ARR_CURR()].dzed003 = "F"
          NEXT FIELD lbl_dzed002
          CONTINUE DIALOG
        END IF

        #判斷輸入Key名稱是否已經存在資料庫
        IF NVL(ls_key_name_orig,cs_null_value) <> cs_null_value THEN
          IF ls_key_name <> ls_key_name_orig THEN
            CALL sadzp000_msg_show_error("adz-00212", "adz-00212", NULL, 1) #Key名稱已經存在資料庫,無法修改.
            LET lo_t_dzed_t[ARR_CURR()].dzed002 = ls_key_name_orig
            NEXT FIELD lbl_dzed002
            CONTINUE DIALOG
          END IF
        END IF

        LET li_arr_curr = ARR_CURR()
        LET ls_identify_name = lo_t_dzed_t[li_arr_curr].dzed002
        IF adzi140_check_key_index_name_exist(ls_identify_name) THEN
          #LET lo_t_dzed_t[li_arr_curr].dzed002 = ""
          LET ls_replace_arg = "|"
          CALL sadzp000_msg_show_error('adz-00198', 'adz-00198', ls_replace_arg, 1) #Key或Index名稱重複
          NEXT FIELD lbl_dzed002
        END IF

        LET lo_t_dzed_t[li_arr_curr].recordtype = cs_data_modified #2022xxxx
        
        LET lb_data_altered = TRUE

      ON CHANGE lbl_dzed003
        LET li_arr_curr = ARR_CURR()

        #檢核是否可異動欄位屬性
        LET ls_ship_notice = lo_t_dzed_t[li_arr_curr].dzed010
        CALL adzi140_check_attribute_could_modify(ls_ship_notice,ms_dgenv) RETURNING lb_result
        IF NOT lb_result THEN
          LET lo_t_dzed_t[li_arr_curr].dzed003 = ls_old_key_type
          CONTINUE DIALOG
        END IF

        IF ls_old_key_type = "P" THEN
          LET lo_t_dzed_t[li_arr_curr].dzed003 = ls_old_key_type
          LET ls_replace_arg = "|"
          CALL sadzp000_msg_show_error('adz-00210', 'adz-00210', ls_replace_arg, 1)
        ELSE
          #Primary key 需由 r.t 自己產生
          IF lo_t_dzed_t[li_arr_curr].dzed003 = "P" THEN
            LET lo_t_dzed_t[li_arr_curr].dzed003 = ls_old_key_type #2022xxxx
            LET ls_replace_arg = "|"
            CALL sadzp000_msg_show_error('adz-00209', 'adz-00209', ls_replace_arg, 1)
          END IF

          #Foreign Key 給預設名稱
          IF ((lo_t_dzed_t[li_arr_curr].dzed003 = "F") OR (lo_t_dzed_t[li_arr_curr].dzed003 = "R")) THEN --AND NVL(lo_t_dzed_t[li_arr_curr].dzed002,cs_null_value) = cs_null_value THEN
            LET ls_table_name = m_dzea_t[mi_dzea_index].dzea001
            CALL sadzi140_util_get_table_leading_code(ls_table_name) RETURNING ls_leading_code

            IF ((lo_t_dzed_t[li_arr_curr].dzed003 = "F") AND NVL(lo_t_dzed_t[li_arr_curr].dzed002,cs_null_value) = cs_null_value) THEN
              IF m_dzea_t[mi_dzea_index].dzea018 = cs_dgenv_standard THEN
                #LET lo_t_dzed_t[li_arr_curr].dzed002 = ls_leading_code,ms_table_user_custom_code,cs_fk_keyword
                #LET lo_t_dzed_t[li_arr_curr].dzed002 = ms_key_user_alter_leading_code,ls_leading_code,ms_column_user_alter_code,cs_fk_keyword
                LET lo_t_dzed_t[li_arr_curr].dzed002 = ms_key_user_alter_leading_code,ls_leading_code,cs_fd_keyword
              ELSE
                #LET lo_t_dzed_t[li_arr_curr].dzed002 = ls_leading_code,cs_fk_keyword
                LET lo_t_dzed_t[li_arr_curr].dzed002 = ls_leading_code,cs_fd_keyword
              END IF
            END IF

            IF ((lo_t_dzed_t[li_arr_curr].dzed003 = "R") AND NVL(lo_t_dzed_t[li_arr_curr].dzed002,cs_null_value) = cs_null_value) THEN
              IF m_dzea_t[mi_dzea_index].dzea018 = cs_dgenv_standard THEN
                #LET lo_t_dzed_t[li_arr_curr].dzed002 = ls_leading_code,ms_table_user_custom_code,cs_rk_keyword
                #LET lo_t_dzed_t[li_arr_curr].dzed002 = ms_key_user_alter_leading_code,ls_leading_code,ms_column_user_alter_code,cs_rk_keyword
                LET lo_t_dzed_t[li_arr_curr].dzed002 = ms_key_user_alter_leading_code,ls_leading_code,cs_rk_keyword
              ELSE
                #LET lo_t_dzed_t[li_arr_curr].dzed002 = ls_leading_code,cs_rk_keyword
                LET lo_t_dzed_t[li_arr_curr].dzed002 = ls_leading_code,cs_rk_keyword
              END IF
            END IF
          END IF

          {
          #只能有一個 Foreign Key
          IF lo_t_dzed_t[li_arr_curr].dzed003 = "R" THEN
            LET li_foreign_cnts = 0
            FOR li_rec_cnt = 1 TO lo_t_dzed_t.getLength()
              IF lo_t_dzed_t[li_rec_cnt].dzed003 = "R" THEN
                LET li_foreign_cnts = li_foreign_cnts + 1
              END IF
            END FOR
            IF li_foreign_cnts > 1 THEN
              LET lo_t_dzed_t[li_arr_curr].dzed003 = ""
              LET ls_replace_arg = "|"
              CALL sadzp000_msg_show_error('adz-00235', 'adz-00235', ls_replace_arg, 1)
            END IF
          END IF
          }
        END IF
        LET lb_active = TRUE
        IF ((lo_t_dzed_t[li_arr_curr].dzed003 <> "F") AND (lo_t_dzed_t[li_arr_curr].dzed003 <> "R")) THEN
          LET lo_t_dzed_t[li_arr_curr].dzed005 = ""
          LET lo_t_dzed_t[li_arr_curr].dzed006 = ""
          LET lb_active = FALSE
        END IF
        CALL DIALOG.setFieldActive("sr_tablekey.lbl_dzed005", lb_active)
        CALL DIALOG.setFieldActive("sr_tablekey.lbl_dzed006", lb_active)
        DISPLAY lo_t_dzed_t[li_arr_curr].* TO sr_TableKey[li_arr_curr].*

        LET lo_t_dzed_t[li_arr_curr].recordtype = cs_data_modified #2022xxxx
        
        LET lb_data_altered = TRUE

      #20160614 begin  
      AFTER FIELD lbl_dzed004
        GOTO _on_ahange_lbl_dzed004  
      ON CHANGE lbl_dzed004
        LABEL _on_ahange_lbl_dzed004:
        
        LET li_arr_curr = ARR_CURR()

        LET ls_table_name = m_dzea_t[mi_dzea_index].dzea001
        LET ls_columns    = lo_t_dzed_t[li_arr_curr].dzed004

        IF ls_columns.trim() IS NOT NULL THEN 
          INITIALIZE lo_column_array TO NULL
          CALL adzi140_get_input_column_list(ls_columns) RETURNING lo_column_array
          FOR li_rec_cnt = 1 TO lo_column_array.getLength()
            CALL sadzi140_util_check_table_column_if_exists(ls_table_name,lo_column_array[li_rec_cnt].COLUMN_NAME) RETURNING lb_result
            IF NOT lb_result THEN
              LET ls_replace_arg = lo_column_array[li_rec_cnt].COLUMN_NAME,"|",ls_table_name,"|"
              CALL sadzp000_msg_show_error("adz-00869", "adz-00869", ls_replace_arg, 1) #欄位 xxxx 不存在r.t資料表設計器中
              NEXT FIELD lbl_dzed004
              CONTINUE DIALOG
            END IF
          END FOR
        END IF
      #20160614 end
      
      ON CHANGE lbl_dzed005
        LET li_arr_curr = ARR_CURR()
        LET ls_fd_table_name = lo_t_dzed_t[li_arr_curr].dzed005

        #20160614 begin  
        IF ls_fd_table_name.trim() IS NOT NULL THEN 
          CALL sadzi140_util_check_table_if_exists(ls_fd_table_name) RETURNING lb_result
          IF NOT lb_result THEN
            LET ls_replace_arg = ls_fd_table_name,"|"
            CALL sadzp000_msg_show_error("adz-00493", "adz-00493", ls_replace_arg, 1) #xxxx_t表格不存在r.t資料表設計器中
            LET lo_t_dzed_t[li_arr_curr].dzed005 = ""
            NEXT FIELD lbl_dzed005
            CONTINUE DIALOG
          END IF
        END IF  
        #20160614 end

        #外來鍵表格不可重複
        FOR li_rec_cnt = 1 TO lo_t_dzed_t.getLength()
          IF li_arr_curr <> li_rec_cnt THEN
            IF lo_t_dzed_t[li_rec_cnt].dzed005 = ls_fd_table_name THEN
              CALL sadzp000_msg_show_error("adz-00534", "adz-00534", NULL, 1)
              LET lo_t_dzed_t[li_arr_curr].dzed005 = ""
              LET lo_t_dzed_t[li_arr_curr].dzed006 = ""
              NEXT FIELD lbl_dzed005
              CONTINUE DIALOG
            END IF
          END IF
        END FOR

        IF (((lo_t_dzed_t[li_arr_curr].dzed003 = "F") OR (lo_t_dzed_t[li_arr_curr].dzed003 = "R")) AND (lo_t_dzed_t[li_arr_curr].dzed005 IS NOT NULL)) THEN
          CALL adzi140_get_foreign_key_data(lo_t_dzed_t[li_arr_curr].dzed005) RETURNING ls_key_list
          LET lo_t_dzed_t[li_arr_curr].dzed006 = ls_key_list
        ELSE
          LET lo_t_dzed_t[li_arr_curr].dzed006 = ""
        END IF
        DISPLAY lo_t_dzed_t[li_arr_curr].* TO sr_TableKey[li_arr_curr].*

        LET lo_t_dzed_t[li_arr_curr].recordtype = cs_data_modified #2022xxxx

        LET lb_data_altered = TRUE

      #20160614 begin  
      AFTER FIELD lbl_dzed006
        GOTO _on_ahange_lbl_dzed006  
      ON CHANGE lbl_dzed006
        LABEL _on_ahange_lbl_dzed006:
        LET li_arr_curr = ARR_CURR()
        
        LET ls_table_name = lo_t_dzed_t[li_arr_curr].dzed005
        LET ls_columns    = lo_t_dzed_t[li_arr_curr].dzed006

        IF ls_columns.trim() IS NOT NULL THEN 
          INITIALIZE lo_column_array TO NULL
          CALL adzi140_get_input_column_list(ls_columns) RETURNING lo_column_array
          FOR li_rec_cnt = 1 TO lo_column_array.getLength()
            CALL sadzi140_util_check_table_column_if_exists(ls_table_name,lo_column_array[li_rec_cnt].COLUMN_NAME) RETURNING lb_result
            IF NOT lb_result THEN
              LET ls_replace_arg = lo_column_array[li_rec_cnt].COLUMN_NAME,"|",ls_table_name,"|"
              CALL sadzp000_msg_show_error("adz-00869", "adz-00869", ls_replace_arg, 1) #欄位 xxxx 不存在r.t資料表設計器中
              NEXT FIELD lbl_dzed006
              CONTINUE DIALOG
            END IF
          END FOR
        END IF  
        #20160614 end
        
    END INPUT

    ON ACTION CANCEL
      LET lb_data_altered = FALSE
      CALL adzi140_fill_dzed_t(m_dzea_t[mi_dzea_index].dzea001)
      EXIT DIALOG

    ON ACTION ACCEPT
      #做存檔前的驗證
      FOR li_rec_cnt = 1 TO lo_t_dzed_t.getLength()
        LET ls_key_name = lo_t_dzed_t[li_rec_cnt].dzed002
        LET ls_key_name_orig = lo_t_dzed_t[li_rec_cnt].dzed002_orig
        IF NVL(ls_key_name_orig,cs_null_value) <> cs_null_value THEN
          IF ls_key_name <> ls_key_name_orig THEN
            CALL sadzp000_msg_show_error("adz-00212", "adz-00212", NULL, 1) #Key名稱已經存在資料庫,無法修改.
            LET lo_t_dzed_t[li_rec_cnt].dzed002 = ls_key_name_orig
            NEXT FIELD lbl_dzed002
            CONTINUE DIALOG
          END IF
        END IF
        LET ls_identify_name = lo_t_dzed_t[li_rec_cnt].dzed002
        LET ls_key_type =  lo_t_dzed_t[li_rec_cnt].dzed003
        IF (NVL(ls_identify_name,cs_null_value) = cs_null_value) AND (NVL(ls_key_type,cs_null_value) <> "") THEN
          LET ls_replace_arg = "|"
          CALL sadzp000_msg_show_error('adz-00200', 'adz-00200', ls_replace_arg, 1) #請輸入Key名稱
          CALL DIALOG.setCurrentRow("sr_tablekey", li_rec_cnt)
          NEXT FIELD lbl_dzed002
          CONTINUE DIALOG
        END IF
        LET ls_identify_col_name = lo_t_dzed_t[li_rec_cnt].dzed004
        IF (NVL(ls_identify_name,cs_null_value) <> cs_null_value) AND (NVL(ls_identify_col_name,cs_null_value) = cs_null_value) THEN
          LET ls_replace_arg = "|"
          CALL sadzp000_msg_show_error('adz-00203', 'adz-00203', ls_replace_arg, 1) #請輸入Key欄位
          CALL DIALOG.setCurrentRow("sr_tablekey", li_rec_cnt)
          NEXT FIELD lbl_dzed004
          CONTINUE DIALOG
        END IF
        LET ls_if_foreign = lo_t_dzed_t[li_rec_cnt].dzed003
        #Foreign Key 時需判段表格及欄位是否輸入
        IF ((ls_if_foreign = "F") OR (ls_if_foreign = "R")) THEN
          LET ls_identify_name = lo_t_dzed_t[li_rec_cnt].dzed005
          IF NVL(ls_identify_name,cs_null_value) = cs_null_value THEN
            LET ls_replace_arg = "|"
            CALL sadzp000_msg_show_error('adz-00204', 'adz-00204', ls_replace_arg, 1) #請輸入ForeignKey名稱
            CALL DIALOG.setCurrentRow("sr_tablekey", li_rec_cnt)
            NEXT FIELD lbl_dzed005
            CONTINUE DIALOG
          END IF
          LET ls_identify_col_name = lo_t_dzed_t[li_rec_cnt].dzed006
          IF (NVL(ls_identify_name,cs_null_value) <> cs_null_value) AND (NVL(ls_identify_col_name,cs_null_value) = cs_null_value) THEN
            LET ls_replace_arg = "|"
            CALL sadzp000_msg_show_error('adz-00205', 'adz-00205', ls_replace_arg, 1) #請輸入ForeignKey欄位
            CALL DIALOG.setCurrentRow("sr_tablekey", li_rec_cnt)
            NEXT FIELD lbl_dzed006
            CONTINUE DIALOG
          END IF
          #20160615 begin
          IF lo_t_dzed_t[li_rec_cnt].recordtype = cs_data_added THEN
            LET lb_add_new_footing_data = TRUE
          END IF
          #20160615 end
        END IF
      END FOR

      LET lb_error = FALSE
      LET ls_main_table  = m_dzea_t[mi_dzea_index].dzea001
      LET ls_main_module = m_dzea_t[mi_dzea_index].dzea003
      LET ls_owner       = ms_master_user

      #bein 處理新增及異動
      BEGIN WORK
      LET li_counts = 0

      FOR li_rec_cnt = 1 TO lo_t_dzed_t.getLength()

        LET lo_t_dzed_t[li_rec_cnt].dzed001 = ls_main_table
        IF (lo_t_dzed_t[li_rec_cnt].recordtype = cs_data_modified) OR (lo_t_dzed_t[li_rec_cnt].dzed008 IS NULL) THEN #2022xxxx 
          LET lo_t_dzed_t[li_rec_cnt].dzed008 = ms_dgenv
        END IF #2022xxxx  
        IF lo_t_dzed_t[li_rec_cnt].dzed009 IS NULL THEN
          LET lo_t_dzed_t[li_rec_cnt].dzed009 = ms_dgenv
        END IF
        IF lo_t_dzed_t[li_rec_cnt].dzed010 IS NULL THEN
          LET lo_t_dzed_t[li_rec_cnt].dzed010 = "N"
        END IF
        CALL sadzi140_crud_insert_update_dzed_t(lo_t_dzed_t[li_rec_cnt].*)

        LET li_counts = li_counts + 1
        IF li_counts >= 100 THEN
          COMMIT WORK
          BEGIN WORK
          LET li_counts = 0
        END IF

      END FOR
      COMMIT WORK
      #end 處理新增及異動

      #begin 處理刪除
      IF ls_del_list.getLength() > 1 THEN
        BEGIN WORK

        LET ls_del_list = ls_del_list.subString(2,ls_del_list.getLength())

        #刪除DZED
        CALL adzi140_delete_dzed_data(ls_main_table,ls_del_list) RETURNING lb_success
        IF NOT lb_success THEN
          ROLLBACK WORK
          BEGIN WORK
        END IF

        COMMIT WORK
      END IF
      #end 處理刪除

      #判斷當為客制環境時,則需要異動模組別為 Cxx 模組.
      {
      #2014.11.18 為不異動模組別
      IF ms_dgenv = cs_dgenv_customize THEN
        LET ls_module_code = m_dzea_t[mi_dzea_index].dzea003
        #Check Key Data
        CALL sadzi140_db_check_key_if_customize(ls_main_table) RETURNING lb_result
        IF lb_result THEN
          CALL sadzi140_db_get_module_code_by_dgenv(ls_module_code,ms_dgenv) RETURNING ls_module_code
          CALL sadzi140_db_update_module_code(ls_main_table,ls_module_code,ms_dgenv)
        ELSE
          CALL sadzi140_db_get_module_code_by_dgenv(ls_module_code,cs_dgenv_standard) RETURNING ls_module_code
          CALL sadzi140_db_update_module_code(ls_main_table,ls_module_code,cs_dgenv_standard)
        END IF
      END IF
      }

      #CALL sadzi140_db_update_alter_code(ls_owner,ls_main_table) #20170523 mark
      CALL sadzi140_db_set_alter_code(ls_owner,ls_main_table,"N") #20170523 add
      {
      #產出 table.sch 檔
      #CALL sadzi140_xml_get_lang_type_list() RETURNING lo_lang_arr
      CALL sadzi140_crud_get_static_lang_list() RETURNING lo_lang_arr
      LET lb_gen_tbl = TRUE #20170908
      LET ls_gen_list = "" #20170908
      FOR li_loop = 1 TO lo_lang_arr.getLength()
        LET ls_lang = lo_lang_arr[li_loop]
        CALL sadzi140_xml_gen_table_datas(ls_main_table,NVL(ls_module_code,ls_main_module),ls_lang,ls_gen_list) RETURNING lb_gen_tbl,ls_gen_list #20170908
        CALL sadzi140_db_gen_table_schema(ls_main_table,NVL(ls_module_code,ls_main_module),FALSE,ls_lang) RETURNING ls_schema_sql
      END FOR
      }
      
      #20160615 begin
      IF lb_add_new_footing_data THEN
        CALL sadzp000_msg_show_info('adz-00870', 'adz-00870', ls_replace_arg, 1) #您有新增Footing Data, 請確認對應的欄位是否正確.
      END IF  
      #20160615 end

      CALL adzi140_fill_dzed_t(ls_main_table)
      EXIT DIALOG

    ON ACTION CLOSE
      EXIT DIALOG

  END DIALOG

  CALL adzi140_set_editor_mode(p_dialog,cs_editor_mode_browse)

  LET lb_return = lb_data_altered

  RETURN lb_return

END FUNCTION

#建立 Grant Privileges 的資料
FUNCTION adzi140_alter_privilege_data(p_dialog)
DEFINE
  p_dialog ui.dialog
DEFINE
  lo_t_dzen_t     DYNAMIC ARRAY OF T_DZEN_T,
  lo_del_list     DYNAMIC ARRAY OF T_GRANT_ACCEPT_LIST,
  li_rec_cnt      INTEGER,
  li_counts       INTEGER,
  ls_table_name   STRING,
  lb_error        BOOLEAN,
  ls_owner        STRING,
  li_arr_curr     INTEGER,
  ls_replace_arg  STRING,
  lb_success      BOOLEAN,
  lb_active       BOOLEAN,
  lb_data_altered BOOLEAN,
  ls_grant_schema  STRING,
  ls_accept_schema STRING
DEFINE
  lb_return  BOOLEAN

  LET lb_error = FALSE
  LET lb_data_altered = FALSE

  CALL adzi140_set_editor_mode(p_dialog,cs_editor_mode_edit)

  DIALOG

    INPUT ARRAY lo_t_dzen_t FROM sr_TablePrivileges.* ATTRIBUTE(WITHOUT DEFAULTS)

      BEFORE INPUT
        LET lo_t_dzen_t.* = m_dzen_t.*
        CALL DIALOG.setActionHidden("insert",TRUE)
        CALL DIALOG.setActionHidden("append",TRUE)
        CALL DIALOG.setActionHidden("delete",TRUE)
        CALL DIALOG.setCurrentRow("sr_tableprivileges",mi_dzen_index)

      BEFORE ROW
        LET li_arr_curr = ARR_CURR()
        LET ls_grant_schema  = lo_t_dzen_t[li_arr_curr].dzen002
        LET ls_accept_schema = lo_t_dzen_t[li_arr_curr].dzen003

      {
      AFTER ROW
        LET li_arr_curr = ARR_CURR()
        LET ls_schema_name   = lo_t_dzen_t[li_arr_curr].dzen002
        LET ls_accept_schema = lo_t_dzen_t[li_arr_curr].dzen003
        FOR li_rec_cnt = 1 TO lo_t_dzen_t.getLength()
          IF li_arr_curr <> li_rec_cnt THEN
            IF lo_t_dzen_t[li_rec_cnt].dzen002 = ls_schema_name THEN
              CALL sadzp000_msg_show_error("adz-00742", "adz-00742", NULL, 1)
              NEXT FIELD lbl_dzen002
              CONTINUE DIALOG
            END IF
          END IF
        END FOR
      }

      ON CHANGE lbl_dzen002
        GOTO _check_grant

      ON CHANGE lbl_dzen003
        LABEL _check_grant:
        LET li_arr_curr = ARR_CURR()
        LET ls_table_name    = m_dzea_t[mi_dzea_index].dzea001
        LET ls_grant_schema  = lo_t_dzen_t[li_arr_curr].dzen002
        LET ls_accept_schema = lo_t_dzen_t[li_arr_curr].dzen003
        IF (ls_grant_schema = ls_accept_schema) THEN
          #授與權限與取得權限的Schema名稱不可相同
          LET lo_t_dzen_t[li_arr_curr].dzen003 = ""
          LET ls_replace_arg = "|"
          CALL sadzp000_msg_show_error('adz-00746', 'adz-00746', ls_replace_arg, 1)
          NEXT FIELD lbl_dzen003
        END IF
        IF adzi140_check_privilege_name_exist(li_arr_curr,lo_t_dzen_t) THEN
          #欲賦予權限的Schema名稱重複
          LET lo_t_dzen_t[li_arr_curr].dzen003 = ""
          LET ls_replace_arg = "|"
          CALL sadzp000_msg_show_error('adz-00742', 'adz-00742', ls_replace_arg, 1)
          NEXT FIELD lbl_dzen003
        END IF
        #設定客制標示, 已經設為客制的則不再重設
        IF lo_t_dzen_t[li_arr_curr].dzen011 <> cs_dgenv_customize THEN
          LET lo_t_dzen_t[li_arr_curr].dzen011 = ms_dgenv
        END IF

        #LET lb_data_altered = TRUE

      ON CHANGE lbl_dzen004
        GOTO _CUSTOMIZE_CODE
      ON CHANGE lbl_dzen005
        GOTO _CUSTOMIZE_CODE
      ON CHANGE lbl_dzen006
        GOTO _CUSTOMIZE_CODE
      ON CHANGE lbl_dzen007
        GOTO _CUSTOMIZE_CODE
      ON CHANGE lbl_dzen008
        GOTO _CUSTOMIZE_CODE
      ON CHANGE lbl_dzen009
        GOTO _CUSTOMIZE_CODE
      ON CHANGE lbl_dzen010
        LABEL _CUSTOMIZE_CODE:
        IF lo_t_dzen_t[li_arr_curr].dzen011 <> cs_dgenv_customize THEN
          LET lo_t_dzen_t[li_arr_curr].dzen011 = ms_dgenv
        END IF

      ON ACTION DELETE
        LET li_arr_curr = ARR_CURR()
        LET ls_grant_schema  = lo_t_dzen_t[li_arr_curr].dzen002
        LET ls_accept_schema = lo_t_dzen_t[li_arr_curr].dzen003
        CALL lo_del_list.appendElement()
        LET lo_del_list[lo_del_list.getLength()].p_grant_schema  = lo_t_dzen_t[li_arr_curr].dzen002
        LET lo_del_list[lo_del_list.getLength()].p_accept_schema = lo_t_dzen_t[li_arr_curr].dzen003
        CALL DIALOG.deleteRow("sr_tableprivileges",li_arr_curr)
        #LET lb_data_altered = TRUE

    END INPUT

    ON ACTION CANCEL
      LET lb_data_altered = FALSE
      CALL adzi140_fill_dzen_t(m_dzea_t[mi_dzea_index].dzea001)
      EXIT DIALOG

    ON ACTION ACCEPT
      LET lb_data_altered = TRUE
      LET lb_error = FALSE
      LET ls_table_name  = m_dzea_t[mi_dzea_index].dzea001
      LET ls_owner       = ms_master_user

      #做存檔前的驗證
      FOR li_rec_cnt = 1 TO lo_t_dzen_t.getLength()
        LET ls_grant_schema  = lo_t_dzen_t[li_rec_cnt].dzen002
        LET ls_accept_schema = lo_t_dzen_t[li_rec_cnt].dzen003
        IF ls_grant_schema IS NOT NULL OR ls_accept_schema IS NOT NULL THEN
          #請輸入欲賦予權限的Schema名稱
          IF (NVL(ls_grant_schema,cs_null_value) = cs_null_value) THEN
            LET ls_replace_arg = "|"
            CALL sadzp000_msg_show_error('adz-00743', 'adz-00743', ls_replace_arg, 1)
            CALL DIALOG.setCurrentRow("sr_tableprivileges", li_rec_cnt)
            NEXT FIELD lbl_dzen002
            CONTINUE DIALOG
          END IF
          #請輸入欲取得權限的Schema名稱
          IF (NVL(ls_accept_schema,cs_null_value) = cs_null_value) THEN
            LET ls_replace_arg = "|"
            CALL sadzp000_msg_show_error('adz-00744', 'adz-00744', ls_replace_arg, 1)
            CALL DIALOG.setCurrentRow("sr_tableprivileges", li_rec_cnt)
            NEXT FIELD lbl_dzen003
            CONTINUE DIALOG
          END IF
          #授與權限項目至少需勾選一項
          IF lo_t_dzen_t[li_rec_cnt].dzen004 = "N" AND
             lo_t_dzen_t[li_rec_cnt].dzen005 = "N" AND
             lo_t_dzen_t[li_rec_cnt].dzen006 = "N" AND
             lo_t_dzen_t[li_rec_cnt].dzen007 = "N" AND
             lo_t_dzen_t[li_rec_cnt].dzen008 = "N" AND
             lo_t_dzen_t[li_rec_cnt].dzen009 = "N" AND
             lo_t_dzen_t[li_rec_cnt].dzen010 = "N" THEN
            LET ls_replace_arg = "|"
            CALL sadzp000_msg_show_error('adz-00747', 'adz-00747', ls_replace_arg, 1)
            CALL DIALOG.setCurrentRow("sr_tableprivileges", li_rec_cnt)
            NEXT FIELD lbl_dzen002
            CONTINUE DIALOG
          END IF
        END IF
      END FOR

      #begin 處理新增及異動
      BEGIN WORK
      LET li_counts = 0
      FOR li_rec_cnt = 1 TO lo_t_dzen_t.getLength()

        LET lo_t_dzen_t[li_rec_cnt].dzen001 = ls_table_name

        IF lo_t_dzen_t[li_rec_cnt].dzen011 IS NULL THEN
          LET lo_t_dzen_t[li_rec_cnt].dzen011 = ms_dgenv
        END IF
        IF lo_t_dzen_t[li_rec_cnt].dzen012 IS NULL THEN
          LET lo_t_dzen_t[li_rec_cnt].dzen012 = ms_dgenv
        END IF
        IF lo_t_dzen_t[li_rec_cnt].dzen013 IS NULL THEN
          LET lo_t_dzen_t[li_rec_cnt].dzen013 = "N"
        END IF
        CALL sadzi140_crud_insert_update_dzen_t(lo_t_dzen_t[li_rec_cnt].*)

        LET li_counts = li_counts + 1
        IF li_counts >= 100 THEN
          COMMIT WORK
          BEGIN WORK
          LET li_counts = 0
        END IF

        IF NOT lb_error THEN
        END IF

      END FOR
      COMMIT WORK
      #end 處理新增及異動

      #begin 處理刪除
      IF lo_del_list.getLength() >= 1 THEN
        BEGIN WORK

        #刪除DZEN
        CALL adzi140_delete_dzen_data(ls_table_name,lo_del_list) RETURNING lb_success
        IF NOT lb_success THEN
          ROLLBACK WORK
          BEGIN WORK
        END IF

        COMMIT WORK
      END IF
      #end 處理刪除

      #CALL sadzi140_db_update_alter_code(ls_owner,ls_table_name) #20170523 mark
      CALL adzi140_fill_dzen_t(ls_table_name)
      EXIT DIALOG

    ON ACTION CLOSE
      EXIT DIALOG

  END DIALOG

  CALL adzi140_set_editor_mode(p_dialog,cs_editor_mode_browse)

  LET lb_return = lb_data_altered

  RETURN lb_return

END FUNCTION

#20180302 begin
FUNCTION adzi140_delete_column_design_datas(p_table_name,p_del_list)
DEFINE
  p_table_name STRING,
  p_del_list   STRING
DEFINE
  ls_table_name  STRING,
  ls_del_list    STRING,
  lb_result      BOOLEAN,
  lb_success     BOOLEAN
DEFINE
  lb_return  BOOLEAN  

  LET ls_table_name = p_table_name.trim()
  LET ls_del_list   = p_del_list.trim()

  LET lb_result = TRUE
  LET lb_success = TRUE

  BEGIN WORK

  #刪除DZEB
  CALL adzi140_delete_dzeb_data(ls_table_name,ls_del_list) RETURNING lb_result
  IF NOT lb_result THEN
    LET lb_success = FALSE
    ROLLBACK WORK
    BEGIN WORK
  END IF

  IF lb_success THEN
    #刪除DZEBL
    CALL adzi140_delete_dzebl_data(ls_table_name,ls_del_list) RETURNING lb_result
    IF NOT lb_result THEN
      LET lb_success = FALSE
      ROLLBACK WORK
      BEGIN WORK
    END IF
  END IF

  IF lb_success THEN
    #刪除DZEP
    CALL adzi140_delete_dzep_data(ls_table_name,ls_del_list) RETURNING lb_result
    IF NOT lb_result THEN
      LET lb_success = FALSE
      ROLLBACK WORK
      BEGIN WORK
    END IF
  END IF
  
  COMMIT WORK  

  LET lb_return = lb_success
  
  RETURN lb_return
  
END FUNCTION

#刪除 DZHJ 的資料
FUNCTION adzi140_delete_dzhj_data(p_table_name,p_del_list)
DEFINE
  p_table_name STRING,
  p_del_list   STRING
DEFINE
  ls_table_name  STRING,
  ls_del_list    STRING,
  ls_sql         STRING,
  ls_replace_arg STRING
DEFINE
  lb_success BOOLEAN,
  lb_return  BOOLEAN

  LET ls_table_name = p_table_name.trim()
  LET ls_del_list   = p_del_list.trim()

  LET lb_success = TRUE
  LET lb_return  = TRUE

  IF (ls_table_name IS NOT NULL) AND (ls_del_list IS NOT NULL) THEN
    LET ls_sql = "delete from dzhj_t                   ",
                 " where dzhj001 = '",ls_table_name,"' ",
                 "   and dzhj002 in (",ls_del_list,")  "
    TRY
      PREPARE lpre_delete_dzhj_data FROM ls_sql
      EXECUTE lpre_delete_dzhj_data
    CATCH
      LET lb_success = FALSE
      LET ls_replace_arg = "DZHJ_T","|"
      CALL sadzp000_msg_show_error('adz-00091', 'adz-00091', ls_replace_arg, 1)
    END TRY
  END IF

  LET lb_return = lb_success

  RETURN lb_return

END FUNCTION
#20180302 end

#刪除 DZEB 的資料
FUNCTION adzi140_delete_dzeb_data(p_table_name,p_del_list)
DEFINE
  p_table_name STRING,
  p_del_list   STRING
DEFINE
  ls_table_name  STRING,
  ls_del_list    STRING,
  ls_sql         STRING,
  ls_replace_arg STRING
DEFINE
  lb_success BOOLEAN,
  lb_return  BOOLEAN

  LET ls_table_name = p_table_name.trim()
  LET ls_del_list   = p_del_list.trim()

  LET lb_success = TRUE
  LET lb_return  = TRUE

  IF (ls_table_name IS NOT NULL) AND (ls_del_list IS NOT NULL) THEN
    LET ls_sql = "delete from dzeb_t                   ",
                 " where dzeb001 = '",ls_table_name,"' ",
                 "   and dzeb002 in (",ls_del_list,")  "
    TRY
      PREPARE lpre_delete_dzeb_data FROM ls_sql
      EXECUTE lpre_delete_dzeb_data
    CATCH
      LET lb_success = FALSE
      LET ls_replace_arg = "DZEB_T","|"
      CALL sadzp000_msg_show_error('adz-00091', 'adz-00091', ls_replace_arg, 1)
    END TRY
  END IF

  LET lb_return = lb_success

  RETURN lb_return

END FUNCTION

#刪除 DZEBL 的資料
FUNCTION adzi140_delete_dzebl_data(p_table_name,p_del_list)
DEFINE
  p_table_name STRING,
  p_del_list   STRING
DEFINE
  ls_table_name  STRING,
  ls_del_list    STRING,
  ls_sql         STRING,
  ls_replace_arg STRING
DEFINE
  lb_success BOOLEAN,
  lb_return  BOOLEAN

  LET ls_table_name = p_table_name.trim()
  LET ls_del_list   = p_del_list.trim()

  LET lb_success = TRUE
  LET lb_return  = TRUE

  IF (ls_del_list IS NOT NULL) THEN
    LET ls_sql = "delete from dzebl_t                   ",
                 " where dzebl001 in (",ls_del_list,")  "
    TRY
      PREPARE lpre_delete_dzebl_data FROM ls_sql
      EXECUTE lpre_delete_dzebl_data
    CATCH
      LET lb_success = FALSE
      LET ls_replace_arg = "DZEBL_T","|"
      CALL sadzp000_msg_show_error('adz-00091', 'adz-00091', ls_replace_arg, 1)
    END TRY
  END IF

  LET lb_return = lb_success

  RETURN lb_return

END FUNCTION

#刪除 DZEC 的資料
FUNCTION adzi140_delete_dzec_data(p_table_name,p_del_list)
DEFINE
  p_table_name STRING,
  p_del_list   STRING
DEFINE
  ls_table_name  STRING,
  ls_del_list    STRING,
  ls_sql         STRING,
  ls_replace_arg STRING
DEFINE
  lb_success BOOLEAN,
  lb_return  BOOLEAN

  LET ls_table_name = p_table_name.trim()
  LET ls_del_list   = p_del_list.trim()

  LET lb_success = TRUE
  LET lb_return  = TRUE

  IF (ls_table_name IS NOT NULL) AND (ls_del_list IS NOT NULL) THEN
    LET ls_sql = "delete from dzec_t                   ",
                 " where dzec001 = '",ls_table_name,"' ",
                 "   and dzec002 in (",ls_del_list,")  "
    TRY
      PREPARE lpre_delete_dzec_data FROM ls_sql
      EXECUTE lpre_delete_dzec_data
    CATCH
      LET lb_success = FALSE
      LET ls_replace_arg = "DZEC_T","|"
      CALL sadzp000_msg_show_error('adz-00091', 'adz-00091', ls_replace_arg, 1)
    END TRY
  END IF

  LET lb_return = lb_success

  RETURN lb_return

END FUNCTION

#刪除 DZED 的資料
FUNCTION adzi140_delete_dzed_data(p_table_name,p_del_list)
DEFINE
  p_table_name STRING,
  p_del_list   STRING
DEFINE
  ls_table_name  STRING,
  ls_del_list    STRING,
  ls_sql         STRING,
  ls_replace_arg STRING
DEFINE
  lb_success BOOLEAN,
  lb_return  BOOLEAN

  LET ls_table_name = p_table_name.trim()
  LET ls_del_list   = p_del_list.trim()

  LET lb_success = TRUE
  LET lb_return  = TRUE

  IF (ls_table_name IS NOT NULL) AND (ls_del_list IS NOT NULL) THEN
    LET ls_sql = "delete from dzed_t                   ",
                 " where dzed001 = '",ls_table_name,"' ",
                 "   and dzed002 in (",ls_del_list,")  "
    TRY
      PREPARE lpre_delete_dzed_data FROM ls_sql
      EXECUTE lpre_delete_dzed_data
    CATCH
      LET lb_success = FALSE
      LET ls_replace_arg = "DZED_T","|"
      CALL sadzp000_msg_show_error('adz-00091', 'adz-00091', ls_replace_arg, 1)
    END TRY
  END IF

  LET lb_return = lb_success

  RETURN lb_return

END FUNCTION

#刪除 DZEN 的資料
FUNCTION adzi140_delete_dzen_data(p_table_name,p_del_list)
DEFINE
  p_table_name STRING,
  p_del_list   DYNAMIC ARRAY OF T_GRANT_ACCEPT_LIST
DEFINE
  ls_table_name  STRING,
  lo_del_list    DYNAMIC ARRAY OF T_GRANT_ACCEPT_LIST,
  ls_sql         STRING,
  ls_replace_arg STRING,
  li_loop        INTEGER
DEFINE
  lb_success BOOLEAN,
  lb_return  BOOLEAN

  LET ls_table_name = p_table_name.trim()
  LET lo_del_list   = p_del_list

  LET lb_success = TRUE
  LET lb_return  = TRUE

  FOR li_loop = 1 TO lo_del_list.getLength()
    LET ls_sql = "delete from dzen_t                                          ",
                 " where dzen001 = '",ls_table_name,"'                        ",
                 "   and dzen002 = '",lo_del_list[li_loop].p_grant_schema,"'  ",
                 "   and dzen003 = '",lo_del_list[li_loop].p_accept_schema,"' "
    TRY
      PREPARE lpre_delete_dzen_data FROM ls_sql
      EXECUTE lpre_delete_dzen_data
    CATCH
      LET lb_success = FALSE
    END TRY
  END FOR

  IF NOT lb_success THEN
    LET ls_replace_arg = "DZEN_T","|"
    CALL sadzp000_msg_show_error('adz-00091', 'adz-00091', ls_replace_arg, 1)
  END IF

  LET lb_return = lb_success

  RETURN lb_return

END FUNCTION

#刪除 DZEP 的資料
FUNCTION adzi140_delete_dzep_data(p_table_name,p_del_list)
DEFINE
  p_table_name STRING,
  p_del_list   STRING
DEFINE
  ls_table_name  STRING,
  ls_del_list    STRING,
  ls_sql         STRING,
  ls_replace_arg STRING
DEFINE
  lb_success BOOLEAN,
  lb_return  BOOLEAN

  LET ls_table_name = p_table_name.trim()
  LET ls_del_list   = p_del_list.trim()

  LET lb_success = TRUE
  LET lb_return  = TRUE

  IF (ls_table_name IS NOT NULL) AND (ls_del_list IS NOT NULL) THEN
    #刪除欄位規格設計資料
    LET ls_sql = "delete from dzep_t                   ",
                 " where dzep001 = '",ls_table_name,"' ",
                 "   and dzep002 in (",ls_del_list,")  "
    TRY
      PREPARE lpre_delete_dzep_data FROM ls_sql
      EXECUTE lpre_delete_dzep_data
    CATCH
      LET lb_success = FALSE
      LET ls_replace_arg = "DZEP_T","|"
      CALL sadzp000_msg_show_error('adz-00091', 'adz-00091', ls_replace_arg, 1)
    END TRY
  END IF

  LET lb_return = lb_success

  RETURN lb_return

END FUNCTION

FUNCTION adzi140_refresh_master(p_dialog,p_search_string)
DEFINE
  p_dialog        ui.Dialog,
  p_search_string STRING
DEFINE
  lo_combobox      ui.ComboBox,
  ls_search_string STRING

  INITIALIZE m_dzea_t TO NULL
  INITIALIZE m_dzeb_t TO NULL
  INITIALIZE m_dzec_t TO NULL
  INITIALIZE m_dzed_t TO NULL

  LET ls_search_string = p_search_string

  LET lo_combobox = ui.ComboBox.forName("formonly.cb_moduleselect")
  CALL sadzi140_util_fill_combobox_by_sql(lo_combobox,ms_sql_main_form_modules)

  CALL adzi140_fill_dzea_t(ls_search_string)
  CALL p_dialog.setArrayAttributes("sr_tablelist", m_dzea_t_color)

  LET mi_dzea_index = IIF(mi_dzea_index==0,1,mi_dzea_index)
  #CALL adzi140_refresh_detail(p_dialog,m_dzea_t[mi_dzea_index].DZEA001)
  #設定 Import 功能及建立function table 的功能是否 Enable
  CALL adzi140_set_import_sco_enable(p_dialog)
  CALL adzi140_set_editor_mode(p_dialog,cs_editor_mode_browse)
  CALL adzi140_set_editable(p_dialog,mi_dzea_index)
  CALL adzi140_set_check_out_mode(p_dialog,m_dzea_t[mi_dzea_index].*)

  DISPLAY m_dzea_t[mi_dzea_index].dzea001 TO formonly.lbl_table_name

END FUNCTION

FUNCTION adzi140_refresh_and_locate_master(p_dialog,p_table_name)
DEFINE
  p_dialog     ui.Dialog,
  p_table_name STRING
DEFINE
  ls_table_name STRING,
  li_index INTEGER

  LET ls_table_name = p_table_name
  LET mi_dzea_index = 1

  CALL adzi140_refresh_master(p_dialog,NVL(ms_search,ls_table_name))
  
  FOR li_index = 1 TO m_dzea_t.getLength()
    IF m_dzea_t[li_index].dzea001 = ls_table_name THEN
      LET mi_dzea_index = li_index
      EXIT FOR
    END IF
  END FOR

  CALL adzi140_refresh_master(p_dialog,NVL(ms_search,ls_table_name))

  CALL p_dialog.setCurrentRow("sr_tablelist",mi_dzea_index)
  CALL adzi140_refresh_detail(p_dialog,ls_table_name)

END FUNCTION

FUNCTION adzi140_refresh_detail(p_dialog,p_master_table)
DEFINE
  p_dialog       ui.Dialog,
  p_master_table STRING
DEFINE
  ls_master_table STRING,
  ls_current_item STRING

  LET ls_master_table = p_master_table
  LET ls_current_item = p_dialog.getCurrentItem()

  CALL adzi140_fill_dzeb_t(ls_master_table)
  CALL adzi140_fill_dzec_t(ls_master_table)
  CALL adzi140_fill_dzed_t(ls_master_table)
  CALL adzi140_fill_dzen_t(ls_master_table)
  CALL adzi140_refresh_combobox(ls_master_table) #20170801 #20180903
  
  IF ms_show_detail_color = "Y" THEN
    #非編輯模式才能更新Detail顏色
    IF adzi140_get_editor_mode() = cs_editor_mode_browse THEN
      CALL p_dialog.setArrayAttributes("sr_tablecolumns", m_dzeb_t_color)
      CALL p_dialog.setArrayAttributes("sr_tableindex", m_dzec_t_color)
    END IF
  END IF

END FUNCTION

#20180903 begin
FUNCTION adzi140_refresh_combobox(p_master_table)
DEFINE
  p_master_table STRING
DEFINE
  ls_master_table STRING,
  lo_combobox  ui.ComboBox,
  ls_sql_erp_schema_list STRING,
  lo_str_buf  base.StringBuffer

  LET ls_master_table = p_master_table

  #重新定義撈取授與權限者的資料, 因每一個表格不一定建在所有schema中
  LET lo_str_buf = base.StringBuffer.create()
  CALL lo_str_buf.append(ms_sql_erp_schema_list)
  CALL lo_str_buf.replace(cs_table_tag,ls_master_table,0)
  LET ls_sql_erp_schema_list = lo_str_buf.toString()

  LET lo_combobox = ui.ComboBox.forName("formonly.lbl_dzen002")
  CALL sadzi140_util_fill_combobox_by_sql(lo_combobox,ls_sql_erp_schema_list)

END FUNCTION
#20180903 end

FUNCTION adzi140_fill_module_combobox(p_module_combobox)
DEFINE
  p_module_combobox ui.combobox

  #填入模組別資料
  LET p_module_combobox = ui.ComboBox.forName("formonly.cb_modulename")
  IF (ms_enable_customize = "Y") OR (ms_DGENV = cs_dgenv_customize) THEN
    CALL sadzi140_util_fill_combobox_by_sql(p_module_combobox,ms_sql_cust_modules)
  ELSE
    CALL sadzi140_util_fill_combobox_by_sql(p_module_combobox,ms_sql_modules)
  END IF

  RETURN p_module_combobox

END FUNCTION

FUNCTION adzi140_fill_dzea_t(p_search_cond)
DEFINE
  p_search_cond STRING
DEFINE
  ls_search_cond STRING,
  ls_sql         STRING,
  ls_sql_cond    STRING,
  ls_sql_alm     STRING,
  ls_sql_issue   STRING,
  ls_sql_ooag    STRING,
  ls_sql_module  STRING,#20170425
  ls_column_alm  STRING,
  ls_sql_patch   STRING,
  ls_ent_db      STRING, #20191209
  ls_sql_dzlm    STRING, #20190521
  ls_sql_check_out STRING, #20170425
  ls_sql_industry_table STRING, #20190415
  li_count       INTEGER

  LET ls_search_cond = p_search_cond

  IF (ls_search_cond IS NULL) OR (ls_search_cond = "*") THEN
  
    IF (ls_search_cond IS NULL) THEN 
      LET ls_sql_cond = " AND 1=2"
    END IF  
    
    IF (ls_search_cond = "*") THEN 
      LET ls_sql_cond = ""
    END IF
    
  ELSE
    CALL adzi140_parse_search_contents(ls_search_cond) RETURNING ls_sql_cond
  END IF

  #是否取得未異動的表格
  IF mb_get_issue_table THEN
    LET ls_sql_issue = " and ea.dzea011 = 'N' "
  ELSE
    LET ls_sql_issue = ""
  END IF

  #20170425 begin
  #是否取得有簽出的表格
  IF mb_get_checkout_tables THEN
    LET ls_sql_check_out = " and exists (                                    ", 
                           "               select 1                          ", 
                           "                 from dzlm_t lm                  ",
                           "                where 1=1                        ",
                           "                  and lm.dzlm002 = ea.dzea001    ",
                           "                  and lm.dzlm001 = 'T'           ",
                           "                  and lm.dzlm008 = 'O'           ",
                           "                  and lm.dzlm007 = '",ms_user,"' ",
                           "             )                                   "
  ELSE
    LET ls_sql_check_out = ""
  END IF

  IF ms_module IS NOT NULL AND ms_module <> "*" THEN
    LET ls_sql_module = " and ea.dzea003 = '",ms_module,"' "
  ELSE
    LET ls_sql_module = ""
  END IF
  #20170425 end

  #20190415 begin
  #是否取得行業別表格
  IF mb_get_industry_table THEN
    &ifdef is_msv
    LET ls_sql_industry_table = " and (  ",
                                "       ea.dzea014 <> '",cs_industry_type_standard,"' ",
                                "       or exists (select 1 from dzeb_t eb            ",   
                                "                   where eb.dzeb001 = ea.dzea001     ",
                                "                     and replace(eb.dzeb027,'\"','') not like '%TOPIND:",cs_standard_topind,"%' ",
                                "                     and charindex('TOPIND',eb.dzeb027) > 0 ",
                                "                     and trim(eb.dzeb027) is not null)  ", 
                                "      ) "
    &else
    LET ls_sql_industry_table = " and (  ",
                                "       ea.dzea014 <> '",cs_industry_type_standard,"' ",
                                "       or exists (select 1 from dzeb_t eb            ",   
                                "                   where eb.dzeb001 = ea.dzea001     ",
                                "                     and replace(eb.dzeb027,'\"','') not like '%TOPIND:",cs_standard_topind,"%' ",
                                "                     and instr(eb.dzeb027,'TOPIND') > 0 ",
                                "                     and trim(eb.dzeb027) is not null)  ", 
                                "      ) "
    &endif                                
  ELSE
    LET ls_sql_industry_table = ""
  END IF
  #20190415 end
  
  #Patch Mode
  IF mb_patch THEN
    LET ls_sql_patch = " and exists (                                   ",
                       "              select 1                          ",
                       "                from dzez_t ez                  ",
                       "               where ez.dzez001 = '",ms_GUID,"' ",
                       "                 and ez.dzez002 = ea.dzea001    ",
                       "             )                                  "
  ELSE
    LET ls_sql_patch = ""
  END IF

  IF mb_enable_alm THEN
    LET ls_sql_alm = " left outer join (                                                 ",
                     "                   select *                                        ",
                     "                     from dzlm_t lmo                               ",
                     "                    where 1=1                                      ",
                     "                      and exists (                                 ",
                     "                                   select 1                        ",
                     "                                     from dzlu_t lu                ",
                     "                                    where lu.dzlu003 = lmo.dzlm012 ",
                     "                                 )                                 ",
                     "                 ) lm on lm.dzlm001 = '",cs_spec_type_table,"'     ",
                     "                     and lm.dzlm002 = ea.dzea001                   ",
                     "                     and lm.dzlm008 = '",cs_check_out,"'           "
    
  ELSE
    LET ls_sql_alm = " left outer join dzlm_t lm on lm.dzlm001 = '",cs_spec_type_table,"' ",
                     "                          and lm.dzlm002 = ea.dzea001               ",
                     "                          and lm.dzlm008 = '",cs_check_out,"'       "
  END IF

  #20190521 begin
  LET ls_sql_dzlm = ""
  IF NOT mb_debug THEN 
    LET ls_sql_dzlm = " and ea.dzea001 not in ('dzlm_t') " 
  END IF
  #20190521 end

  CALL sadzi140_db_get_ent_mapping_db() RETURNING ls_ent_db #20191209

  
  LET ls_sql = --"SELECT DISTINCT DECODE(EA.DZEA011,'N','!','I','#','') RECORD_TYPE, ", #20191209 mark #20170706 
               "select distinct case ea.dzea011                                    ", #20191209 begin
               "                  when 'N' then '!'                                ",
               "                  when 'I' then '#'                                ",
               "                  else ''                                          ",
               "                end record_type,                                   ", #20191209 end  
               "       '' seq_no,                                                  ", 
               "       ea.dzea003,ea.dzea001,eal.dzeal003,ea.dzea004,ea.dzea005,   ",
               "       ea.dzea006,ea.dzea007,ea.dzea011,                           ",
               "       lm.dzlm005,lm.dzlm006,lm.dzlm007,lm.dzlm008,lm.dzlm012,     ",
               "       lm.dzlm015,ag.ooag011,ea.dzea012,ea.dzea013,ea.dzea014,     ",
               "       ea.dzea015,ea.dzea016,ea.dzea017,ea.dzea018,ea.dzea019,     ",
               "       ea.dzea020,ea.dzea021,ea.dzeastus                           ",
               "  from dzea_t ea                                                   ",
               "       left outer join dzeal_t eal on eal.dzeal001 = ea.dzea001    ",
               "                                  and eal.dzeal002 = '",ms_lang,"' ",
               ls_sql_alm,
               "       left outer join (                                           ",
               "                         select xa.gzxa001,ag.ooag011              ",
               --"                           from dsdemo.ooag_t ag,                  ", #20191209marked
               --"                                gzxa_t xa                          ", #20191209marked
               "                           from ",ls_ent_db,".ooag_t ag,           ", #20191209 
               "                                gzxa_t xa                          ", #20191209
               "                          where xa.gzxaent = ag.ooagent            ",
               "                            and ag.ooag001 = xa.gzxa003            ",
               "                            and xa.gzxaent = '",ms_enterprise,"'   ",
               "                          group by xa.gzxa001,ag.ooag011           ",
               --"                          order by xa.gzxa001                      ", #20191209marked
               "                       ) ag on ag.gzxa001 = lm.dzlm007             ",
               " where 1=1                                                         ",
               "   and ea.dzea003 not in ('",cs_mdm_module_name,"')                ",
               ls_sql_cond,
               ls_sql_issue,
               ls_sql_patch,
               ls_sql_module, #20170425
               ls_sql_check_out, #20170425
               ls_sql_industry_table, #20190415
               ls_sql_dzlm, #20190521
               " order by ea.dzea003,ea.dzea001                                    "

  PREPARE lpre_fill_dzea_t FROM ls_sql
  DECLARE lcur_fill_dzea_t CURSOR FOR lpre_fill_dzea_t

  CALL m_dzea_t.clear()
  CALL m_dzea_t_color.clear()

  LET li_count = 1

  FOREACH lcur_fill_dzea_t INTO m_dzea_t[li_count].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF
    LET m_dzea_t[li_count].DZEASEQ = li_count

    CASE 
      WHEN m_dzea_t[li_count].dzea011 = "N"
        LET m_dzea_t_color[li_count].recordtype = "red"
        LET m_dzea_t_color[li_count].dzeaseq    = "red reverse"
        LET m_dzea_t_color[li_count].dzea003    = "red reverse"
        LET m_dzea_t_color[li_count].dzea001    = "red reverse"
        LET m_dzea_t_color[li_count].dzea002    = "red reverse"
        LET m_dzea_t_color[li_count].dzea004    = "red reverse"
        LET m_dzea_t_color[li_count].dzea005    = "red reverse"
        LET m_dzea_t_color[li_count].dzea006    = "red reverse"
        LET m_dzea_t_color[li_count].dzea007    = "red reverse"
        LET m_dzea_t_color[li_count].dzea011    = "red reverse"
      WHEN m_dzea_t[li_count].dzea011 = "I"
        LET m_dzea_t_color[li_count].recordtype = "yellow"
        LET m_dzea_t_color[li_count].dzeaseq    = "yellow reverse"
        LET m_dzea_t_color[li_count].dzea003    = "yellow reverse"
        LET m_dzea_t_color[li_count].dzea001    = "yellow reverse"
        LET m_dzea_t_color[li_count].dzea002    = "yellow reverse"
        LET m_dzea_t_color[li_count].dzea004    = "yellow reverse"
        LET m_dzea_t_color[li_count].dzea005    = "yellow reverse"
        LET m_dzea_t_color[li_count].dzea006    = "yellow reverse"
        LET m_dzea_t_color[li_count].dzea007    = "yellow reverse"
        LET m_dzea_t_color[li_count].dzea011    = "yellow reverse"
    OTHERWISE  
      LET m_dzea_t_color[li_count].recordtype = NULL
      LET m_dzea_t_color[li_count].dzeaseq    = NULL
      LET m_dzea_t_color[li_count].dzea003    = NULL
      LET m_dzea_t_color[li_count].dzea001    = NULL
      LET m_dzea_t_color[li_count].dzea002    = NULL
      LET m_dzea_t_color[li_count].dzea004    = NULL
      LET m_dzea_t_color[li_count].dzea005    = NULL
      LET m_dzea_t_color[li_count].dzea006    = NULL
      LET m_dzea_t_color[li_count].dzea007    = NULL
      LET m_dzea_t_color[li_count].dzea011    = NULL
    END CASE

    LET li_count = li_count + 1
  END FOREACH
  CALL m_dzea_t.deleteElement(li_count)

END FUNCTION

FUNCTION adzi140_fill_dzeb_t(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_table_name  STRING,
  ls_sql         STRING,
  li_count       INTEGER

  LET ls_table_name = p_table_name

  &ifdef is_msv
    #20191209 begin
    LET ls_sql = "select case eb.dzeb028 when 'N' then '' else '' end record_type,                  ",
                 "       '', '', eb.dzeb001, eb.dzeb002, ebl.dzebl003,                              ",
                 "       eb.dzeb004, eb.dzeb005, eb.dzeb006, tdl.gztdl003, td.gztd003,              ",
                 "       td.gztd008, eb.dzeb012, eb.dzeb021, eb.dzeb022, eb.dzeb023,                ",
                 "       ebl.dzebl004, eb.dzeb027, eb.dzeb028, eb.dzeb029, eb.dzeb030,              ",
                 "       eb.dzeb031, eb.dzeb002, cat.dzep012, cat.cate, 'N' lang_modified           ",
                 "  from dzeb_t eb                                                                  ",
                 "  left outer join gztd_t td                                                       ",
                 "    on td.gztd001 = eb.dzeb006                                                    ",
                 "  left outer join gztdl_t tdl                                                     ",
                 "    on tdl.gztdl001 = eb.dzeb006                                                  ",
                 "   and tdl.gztdl002 = '",ms_lang,"'                                               ",
                 "  left outer join dzebl_t ebl                                                     ",
                 "    on ebl.dzebl001 = eb.dzeb002                                                  ",
                 "   and ebl.dzebl002 = '",ms_lang,"'                                               ",
                 "  left outer join (select ep.dzep001,                                             ",
                 "                          ep.dzep002,                                             ",
                 "                          ep.dzep011,                                             ",
                 "                          ep.dzep012,                                             ",
                 "                          (select top 10                                          ",
                 "                                  case                                            ",
                 "                                    string_agg(gzcbl002 + '.' + gzcbl004,',')     ",
                 "                                    within group(order by gzcbl002)               ",
                 "                                  when '.' then null else                         ",
                 "                                    string_agg(gzcbl002 + '.' + gzcbl004,',')     ",
                 "                                    within group(order by gzcbl002)               ",
                 "                                  end cate                                        ",
                 "                             from gzcbl_t cbl                                     ",
                 "                            where cbl.gzcbl003 = '",ms_lang,"'                    ",
                 "                              and cbl.gzcbl001 = coalesce(trim(ep.dzep011), 0)    ",
                 "                          ) cate                                                  ",
                 "                     from dzep_t ep                                               ",
                 "                    group by ep.dzep001, ep.dzep002, ep.dzep011, ep.dzep012) cat  ",
                 "    on cat.dzep001 = eb.dzeb001                                                   ",
                 "   and cat.dzep002 = eb.dzeb002                                                   ",
                 " where eb.dzeb001 = '",ls_table_name,"'                                           ",
                 " order by eb.dzeb021, eb.dzeb002                                                  "
    #20191209 end
  &else
  &ifdef is_pgs
    #20200330 begin
    LET ls_sql = "select case eb.dzeb028 when 'N' then '' else '' end record_type,                  ",
                 "       '', '', eb.dzeb001, eb.dzeb002, ebl.dzebl003,                              ",
                 "       eb.dzeb004, eb.dzeb005, eb.dzeb006, tdl.gztdl003, td.gztd003,              ",
                 "       td.gztd008, eb.dzeb012, eb.dzeb021, eb.dzeb022, eb.dzeb023,                ",
                 "       ebl.dzebl004, eb.dzeb027, eb.dzeb028, eb.dzeb029, eb.dzeb030,              ",
                 "       eb.dzeb031, eb.dzeb002, cat.dzep012, cat.cate, 'N' lang_modified           ",
                 "  from dzeb_t eb                                                                  ",
                 "  left outer join gztd_t td                                                       ",
                 "    on td.gztd001 = eb.dzeb006                                                    ",
                 "  left outer join gztdl_t tdl                                                     ",
                 "    on tdl.gztdl001 = eb.dzeb006                                                  ",
                 "   and tdl.gztdl002 = '",ms_lang,"'                                               ",
                 "  left outer join dzebl_t ebl                                                     ",
                 "    on ebl.dzebl001 = eb.dzeb002                                                  ",
                 "   and ebl.dzebl002 = '",ms_lang,"'                                               ",
                 "  left outer join (select ep.dzep001,                                             ",
                 "                          ep.dzep002,                                             ",
                 "                          ep.dzep011,                                             ",
                 "                          ep.dzep012,                                             ",
                 "                          (select case                                            ",
                 "                                    string_agg(gzcbl002||'.'||gzcbl004,','        ",
                 "                                               order by gzcbl002)                 ",
                 "                                  when '.' then null else                         ",
                 "                                    string_agg(gzcbl002||'.'||gzcbl004,','        ",
                 "                                               order by gzcbl002)                 ",
                 "                                  end cate                                        ",
                 "                             from gzcbl_t cbl                                     ",
                 "                            where cbl.gzcbl003 = '",ms_lang,"'                    ",
                 "                              and cbl.gzcbl001 = case                             ", 
                 "                                                   when length(trim(ep.dzep011)) = 0 ",
                 "                                                     then 0                       ",
                 "                                                   else                           ", 
                 "                                                     cast(coalesce(trim(ep.dzep011),'0') as integer) ", 
                 "                                                 end                              ",
                 "                          ) cate                                                  ",
                 "                     from dzep_t ep                                               ",
                 "                    group by ep.dzep001, ep.dzep002, ep.dzep011, ep.dzep012) cat  ",
                 "    on cat.dzep001 = eb.dzeb001                                                   ",
                 "   and cat.dzep002 = eb.dzeb002                                                   ",
                 " where eb.dzeb001 = '",ls_table_name,"'                                           ",
                 " order by eb.dzeb021, eb.dzeb002                                                  "
    #20200330 end
  &else
    LET ls_sql = "SELECT DECODE(EB.DZEB028,'N','','') RECORD_TYPE,'','', EB.DZEB001,                                                             ",
                 "       EB.DZEB002,  EBL.DZEBL003,EB.DZEB004,EB.DZEB005,EB.DZEB006,                                                             ",
                 "       TDL.GZTDL003,TD.GZTD003,  TD.GZTD008,EB.DZEB012,EB.DZEB021,                                                             ",
                 "       EB.DZEB022,EB.DZEB023,EBL.DZEBL004,EB.DZEB027,EB.DZEB028,                                                               ",
                 "       EB.DZEB029,EB.DZEB030,EB.DZEB031,EB.DZEB002,CAT.DZEP012,                                                                ",#最後的 dzeb002 作為判斷欄位是否被修改的依據
                 "       CAT.CATE,'N' LANG_MODIFIED                                                                                              ",
                 "  FROM DZEB_T EB                                                                                                               ",
                 "       LEFT OUTER JOIN GZTD_T TD ON TD.GZTD001 = EB.DZEB006                                                                    ",
                 "       LEFT OUTER JOIN GZTDL_T TDL ON TDL.GZTDL001 = EB.DZEB006                                                                ",
                 "                                  AND TDL.GZTDL002 =  '",ms_lang,"'                                                            ",
                 "       LEFT OUTER JOIN DZEBL_T EBL ON EBL.DZEBL001 = EB.DZEB002                                                                ",
                 "                                  AND EBL.DZEBL002 =  '",ms_lang,"'                                                            ",
                 "       LEFT OUTER JOIN (                                                                                                       ",
                 "                         SELECT EP.DZEP001,EP.DZEP002,EP.DZEP011,EP.DZEP012,                                                   ",
                 "                                (                                                                                              ",
                 "                                  SELECT DECODE(                                                                               ",
                 "                                                LISTAGG(GZCBL002||'.'||GZCBL004,',') WITHIN GROUP (ORDER BY GZCBL002),'.',NULL,",
                 "                                                LISTAGG(GZCBL002||'.'||GZCBL004,',') WITHIN GROUP (ORDER BY GZCBL002)          ",
                 "                                               ) CATE                                                                          ",
                 "                                    FROM GZCBL_T CBL                                                                           ",
                 "                                   WHERE CBL.GZCBL001 = NVL(TRIM(EP.DZEP011),0)                                                ",
                 "                                     AND CBL.GZCBL003 = '",ms_lang,"'                                                          ",
                 "                                     AND ROWNUM <= 10                                                                          ",
                 "                                ) CATE                                                                                         ",
                 "                           FROM DZEP_T EP                                                                                      ",
                 "                          GROUP BY EP.DZEP001,EP.DZEP002,EP.DZEP011,EP.DZEP012                                                 ",
                 "                        ) CAT ON CAT.DZEP001 = EB.DZEB001                                                                      ",
                 "                             AND CAT.DZEP002 = EB.DZEB002                                                                      ",
                 " WHERE EB.DZEB001 = '",ls_table_name,"'                                                                                        ",
                 " ORDER BY EB.DZEB021,EB.DZEB002                                                                                                "
  &endif
  &endif
  
  PREPARE lpre_fill_dzeb_t FROM ls_sql
  DECLARE lcur_fill_dzeb_t CURSOR FOR lpre_fill_dzeb_t

  CALL m_dzeb_t.clear()
  CALL m_dzeb_t_color.clear()

  LET li_count = 1

  #DISPLAY ls_table_name
  FOREACH lcur_fill_dzeb_t INTO m_dzeb_t[li_count].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF
    LET m_dzeb_t[li_count].DZEBSEQ = li_count

    IF m_dzeb_t[li_count].dzeb029 = cs_dgenv_customize THEN --m_dzeb_t[li_count].dzeb030 THEN --20190304
      LET m_dzeb_t_color[li_count].recordtype     = "blue"
      {
      LET m_dzeb_t_color[li_count].groupid        = "red"
      LET m_dzeb_t_color[li_count].dzebseq        = "red"
      LET m_dzeb_t_color[li_count].dzeb001        = "red"
      }
      LET m_dzeb_t_color[li_count].dzeb002        = "blue"
      LET m_dzeb_t_color[li_count].dzeb003        = "blue"
      {
      LET m_dzeb_t_color[li_count].dzeb004        = "red"
      LET m_dzeb_t_color[li_count].dzeb005        = "red"
      LET m_dzeb_t_color[li_count].dzeb006        = "red"
      LET m_dzeb_t_color[li_count].dzeb006_desc   = "red"
      LET m_dzeb_t_color[li_count].dzeb007        = "red"
      LET m_dzeb_t_color[li_count].dzeb008        = "red"
      LET m_dzeb_t_color[li_count].dzeb012        = "red"
      LET m_dzeb_t_color[li_count].dzeb021        = "red"
      LET m_dzeb_t_color[li_count].dzeb022        = "red"
      LET m_dzeb_t_color[li_count].dzeb023        = "red"
      LET m_dzeb_t_color[li_count].dzeb024        = "red"
      LET m_dzeb_t_color[li_count].dzeb027        = "red"
      LET m_dzeb_t_color[li_count].dzeb028        = "red"
      }
    ELSE
      LET m_dzeb_t_color[li_count].recordtype     = NULL
      {
      LET m_dzeb_t_color[li_count].groupid        = NULL
      LET m_dzeb_t_color[li_count].dzebseq        = NULL
      LET m_dzeb_t_color[li_count].dzeb001        = NULL
      }
      LET m_dzeb_t_color[li_count].dzeb002        = NULL
      LET m_dzeb_t_color[li_count].dzeb003        = NULL
      {
      LET m_dzeb_t_color[li_count].dzeb004        = NULL
      LET m_dzeb_t_color[li_count].dzeb005        = NULL
      LET m_dzeb_t_color[li_count].dzeb006        = NULL
      LET m_dzeb_t_color[li_count].dzeb006_desc   = NULL
      LET m_dzeb_t_color[li_count].dzeb007        = NULL
      LET m_dzeb_t_color[li_count].dzeb008        = NULL
      LET m_dzeb_t_color[li_count].dzeb012        = NULL
      LET m_dzeb_t_color[li_count].dzeb021        = NULL
      LET m_dzeb_t_color[li_count].dzeb022        = NULL
      LET m_dzeb_t_color[li_count].dzeb023        = NULL
      LET m_dzeb_t_color[li_count].dzeb024        = NULL
      LET m_dzeb_t_color[li_count].dzeb027        = NULL
      LET m_dzeb_t_color[li_count].dzeb028        = NULL
      }
    END IF

    LET li_count = li_count + 1
  END FOREACH
  CALL m_dzeb_t.deleteElement(li_count)

END FUNCTION

FUNCTION adzi140_fill_dzec_t(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_table_name   STRING,
  ls_sql         STRING,
  li_count       INTEGER

  LET ls_table_name = p_table_name

  LET ls_sql = "select case dzec005 when 'N' then '' else '' end  record_type,   ",
               "       dzec001,dzec002,dzec003,dzec004,dzec005,                  ",
               "       dzec006,dzec007,dzec008,dzecstus,                         ", #20170912 add dzecstus
               "       dzec002 dzec002_orig                                      ", #最後的DZEC002是為了比對用
               "  from dzec_t                                                    ",
               " where dzec001 = '",ls_table_name, "'                            ",
               " order by dzecstus desc, dzec002                                 "  #20170912 add dzecstus

  PREPARE lpre_fill_dzec_t FROM ls_sql
  DECLARE lcur_fill_dzec_t CURSOR FOR lpre_fill_dzec_t

  CALL m_dzec_t.clear()

  LET li_count = 1

  #DISPLAY ls_table_name
  FOREACH lcur_fill_dzec_t INTO m_dzec_t[li_count].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    #20170912 begin
    IF m_dzec_t[li_count].dzecstus = "N" THEN
      LET m_dzec_t_color[li_count].recordtype = "lightgray reverse"
      LET m_dzec_t_color[li_count].dzec001    = "lightgray reverse"
      LET m_dzec_t_color[li_count].dzec002    = "lightgray reverse"
      LET m_dzec_t_color[li_count].dzec003    = "lightgray reverse"
      LET m_dzec_t_color[li_count].dzec004    = "lightgray reverse"
      LET m_dzec_t_color[li_count].dzec005    = "lightgray reverse"
      LET m_dzec_t_color[li_count].dzecstus   = "lightgray reverse"
    ELSE
      LET m_dzec_t_color[li_count].recordtype = NULL
      LET m_dzec_t_color[li_count].dzec001    = NULL
      LET m_dzec_t_color[li_count].dzec002    = NULL
      LET m_dzec_t_color[li_count].dzec003    = NULL
      LET m_dzec_t_color[li_count].dzec004    = NULL
      LET m_dzec_t_color[li_count].dzec005    = NULL
      LET m_dzec_t_color[li_count].dzecstus   = NULL
    END IF
    #20170912 end

    LET li_count = li_count + 1
  END FOREACH
  CALL m_dzec_t.deleteElement(li_count)

END FUNCTION

FUNCTION adzi140_fill_dzed_t(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_table_name  STRING,
  ls_sql         STRING,
  li_count       INTEGER

  LET ls_table_name = p_table_name

  LET ls_sql = "select '',                                      ",
               "       dzed001,dzed002,dzed003,dzed004,dzed005, ",
               "       dzed006,dzed007,dzed008,dzed009,dzed010, ",
               "       dzed002 dzed002_orig                     ", #最後面的dzed002是為了比對用
               "  from dzed_t                                   ",
               " where dzed001 = '",ls_table_name, "'           ",
               " order by dzed002                               "

  PREPARE lpre_fill_dzed_t FROM ls_sql
  DECLARE lcur_fill_dzed_t CURSOR FOR lpre_fill_dzed_t

  CALL m_dzed_t.clear()

  LET li_count = 1

  #DISPLAY ls_table_name
  FOREACH lcur_fill_dzed_t INTO m_dzed_t[li_count].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF
    LET li_count = li_count + 1
  END FOREACH
  CALL m_dzed_t.deleteElement(li_count)

END FUNCTION

FUNCTION adzi140_fill_dzen_t(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_table_name  STRING,
  ls_sql         STRING,
  li_count       INTEGER

  LET ls_table_name = p_table_name

  LET ls_sql = "select '',                                      ",
               "       dzen001,dzen002,dzen003,dzen004,dzen005, ",
               "       dzen006,dzen007,dzen008,dzen009,dzen010, ",
               "       dzen011,dzen012,dzen013,                 ",
               "       dzen002 dzen002_orig                     ", #最後面的dzen002是為了比對用
               "  from dzen_t                                   ",
               " where dzen001 = '",ls_table_name, "'           ",
               " order by dzen002                               "

  PREPARE lpre_fill_dzen_t FROM ls_sql
  DECLARE lcur_fill_dzen_t CURSOR FOR lpre_fill_dzen_t

  CALL m_dzen_t.clear()

  LET li_count = 1

  #DISPLAY ls_table_name
  FOREACH lcur_fill_dzen_t INTO m_dzen_t[li_count].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF
    LET li_count = li_count + 1
  END FOREACH
  CALL m_dzen_t.deleteElement(li_count)

END FUNCTION

FUNCTION adzi140_fill_dzel_t(p_dgenv,p_topind) #20190304
DEFINE
  p_dgenv   STRING,
  p_topind  STRING #20190304
DEFINE
  ls_dgenv      STRING,
  ls_topind     STRING,
  ls_sql        STRING,
  ls_sql_topind STRING, #20190304
  li_count      INTEGER

  LET ls_dgenv  = p_dgenv.toLowerCase()
  LET ls_topind = p_topind #20190304

  #20190304 begin
  LET ls_sql_topind = ""
  IF ms_dgenv = cs_dgenv_standard THEN 
    IF ls_topind <> ms_topind THEN
      LET ls_sql_topind = " and el.dzel001 = '",cs_cdf_serial_number,"' " 
    END IF
  END IF  
  #20190304 end

  LET ls_sql = "select '',                                     ",
               "       el.dzel001   cdf,                       ",
               "       ell.dzell003 cdf_desc                   ",
               "  from dzel_t el                               ",
               "  left outer join dzell_t ell                  ",
               "               on ell.dzell001 = el.dzel001    ",
               "              and ell.dzell002 = '",ms_lang,"' ",
               " where el.dzel006 = 'Y'                        ",
               "   and el.dzel003 like '%",ls_dgenv.trim(),"%' ",
               ls_sql_topind, #20190304
               " order by el.dzel009                           "

  PREPARE lpre_fill_dzel_t FROM ls_sql
  DECLARE lcur_fill_dzel_t CURSOR FOR lpre_fill_dzel_t

  CALL m_dzel_t.clear()

  LET li_count = 1

  FOREACH lcur_fill_dzel_t INTO m_dzel_t[li_count].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF
    LET li_count = li_count + 1
  END FOREACH
  CALL m_dzel_t.deleteElement(li_count)

END FUNCTION

FUNCTION adzi140_fill_table_synonym_list(p_master_db,p_table_name,p_module_name)
DEFINE
  p_master_db   STRING,
  p_table_name  STRING,
  p_module_name STRING
DEFINE
  ls_master_db   STRING,
  ls_table_name  STRING,
  ls_module_name STRING,
  ls_sql         STRING,
  ls_db_type     STRING, #20200330
  li_count       INTEGER

  LET ls_master_db   = p_master_db
  LET ls_table_name  = p_table_name
  LET ls_module_name = p_module_name.toUpperCase()

  LET ls_sql = ""

  LET ls_db_type = sadzi140_util_get_db_type() #20200330

  #20200330 begin
  CASE
    WHEN ls_db_type = cs_db_driver_type_pgs OR
         ls_db_type = cs_db_driver_type_msv
      LET ls_sql = "select case da.gzda001                                            ",
                   "         when '",ls_master_db,"' then 0                           ",
                   "       else                                                       ",
                   "         row_number() over(order by da.gzda001 asc)               ",
                   "       end                                            row_number, ",
                   "       '",ls_master_db,"'                             master_db,  ",
                   "       ''                                             table_name, ",
                   "       ''                                             table_type, ",
                   "       'Y'                                            checkboxs,  ",
                   "       da.gzda001                                     ent_db,     ",
                   "       ''                                             sch_type    ",
                   "  from gzda_t da                                                  ", 
                   " where 1=1                                                        ", 
                   "   and da.gzda005  = 'Y'                                          ", 
                   "   and da.gzdastus = 'Y'                                          ", 
                   " order by 1                                                       " 
  OTHERWISE
    LET ls_sql = "select decode(da.gzda001,'",ls_master_db,"',0,rownum) row_number,",
                 "       '",ls_master_db,"'                             master_db, ",
                 "       ''                                             table_name,",
                 "       ''                                             table_type,",
                 "       'Y'                                            checkboxs, ",
                 "       da.gzda001                                     ent_db,    ",
                 "       ''                                             sch_type   ",
                 "  from gzda_t da                                                 ",
                 " where 1=1                                                       ",
                 "   and da.gzda005  = 'Y'                                         ",
                 "   and da.gzdastus = 'Y'                                         ",
                 " order by 1                                                      "
  END CASE
  #20200330 end
  
  #DISPLAY ls_sql

  PREPARE lpre_fill_table_synonym_list FROM ls_sql
  DECLARE lcur_fill_table_synonym_list CURSOR FOR lpre_fill_table_synonym_list
  OPEN lcur_fill_table_synonym_list

  CALL m_table_synonym_list.clear()

  LET li_count = 1

  #DISPLAY "ls_sql : ",ls_sql
  #DISPLAY ls_table_name

  FOREACH lcur_fill_table_synonym_list INTO m_table_synonym_list[li_count].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF
    LET m_table_synonym_list[li_count].p_table_name = ls_table_name
    IF (m_table_synonym_list[li_count].p_ent_db = m_table_synonym_list[li_count].p_master_db) OR NOT
       (ls_module_name = "ADZ" OR ls_module_name = "AZZ" OR ls_module_name = "AWS" OR ls_module_name = "ARP") THEN
      LET m_table_synonym_list[li_count].p_sch_type = "T" #Table
    ELSE
      LET m_table_synonym_list[li_count].p_sch_type = "S" #Synonym
    END IF
    LET li_count = li_count + 1
  END FOREACH
  CALL m_table_synonym_list.deleteElement(li_count)

END FUNCTION

FUNCTION adzi140_fill_function_table_type_list(p_table_info)
DEFINE
  p_table_info   T_DZEA_T
DEFINE
  lo_table_info       T_DZEA_T,
  ls_sql              STRING,
  li_count            INTEGER,
  ls_pre_table_name   STRING,
  ls_front_word       STRING,
  ls_table_desc       STRING,
  ls_module_name      STRING,
  ls_table_type       STRING,
  li_under_line_index INTEGER,
  ls_rear_word        STRING,
  ls_dzlm_sql         STRING
DEFINE
  li_return  INTEGER

  LET lo_table_info.* = p_table_info.*

  LET ls_pre_table_name = lo_table_info.DZEA001
  LET ls_table_desc     = lo_table_info.DZEA002
  LET ls_module_name    = lo_table_info.DZEA003

  IF (m_dzea_t[mi_dzea_index].dzea018 = cs_dgenv_standard) AND (ms_topind = cs_standard_topind) THEN
    LET li_under_line_index = ls_pre_table_name.getIndexOf(ms_table_tail_code,1)

    IF (li_under_line_index = 0) THEN
      LET ls_front_word = ls_pre_table_name
      LET ls_rear_word  = ""
    ELSE
      LET ls_front_word = ls_pre_table_name.subString(1,li_under_line_index-1)
      LET ls_rear_word  = ms_table_tail_combine_code
    END IF
  ELSE
    #行業別
    IF ms_topind <> cs_standard_topind THEN 
      LET li_under_line_index = ls_pre_table_name.getIndexOf(ms_table_tail_code,1)
    ELSE
      LET li_under_line_index = ls_pre_table_name.getIndexOf(ms_table_tail_combine_code,1)
    END IF

    IF (li_under_line_index = 0) THEN
      LET ls_front_word = ls_pre_table_name
      LET ls_rear_word  = ""
    ELSE
      LET ls_front_word = ls_pre_table_name.subString(1,li_under_line_index-1)
      #行業別
      IF ms_topind <> cs_standard_topind THEN 
        LET ls_rear_word  = ms_table_tail_code
      ELSE
        LET ls_rear_word  = ms_table_tail_combine_code
      END IF  
    END IF
  END IF

  #20200330 begin
  LET ls_sql = "select '",ls_module_name,"',                                                               ",
               "       '",ls_front_word,"',                                                                ",
               "       '",ls_front_word,"'",ms_concat_sign,"el.dzel008",ms_concat_sign,"'",ls_rear_word,"',",
               "       '",ls_pre_table_name,"',                                                            ",
               "       el.dzel008,                                                                         ",
               "       '',                                                                                 ",
               "       '',                                                                                 ",
               "       'N',                                                                                ",
               "       el.dzel001,                                                                         ",
               "       '",ls_table_desc,"'",ms_concat_sign,"el.dzel002                                     ",
               "  from dzel_t el                                                                           ",
               " where el.dzel007 = 'Y'                                                                    ",
               "   and not exists (                                                                        ",
               "                    select 1                                                               ",
               "                      from dzea_t ea                                                       ",
               "                     where ea.dzea001 = '",ls_front_word,"'",ms_concat_sign,"el.dzel008",ms_concat_sign,"'",ls_rear_word,"' ",
               #"                     WHERE EA.DZEA008 = EL.DZEL001                                          ",
               #"                       AND EA.DZEA009 = '",ls_pre_table_name,"'                             ",
               "                  )                                                                        ",
               ls_dzlm_sql,
               " order by el.dzel009                                                                       "
  #20200330 end

  PREPARE lpre_fill_function_table_type_list FROM ls_sql
  DECLARE lcur_fill_function_table_type_list CURSOR FOR lpre_fill_function_table_type_list

  CALL m_function_table_type_list.clear()

  LET li_count = 1

  FOREACH lcur_fill_function_table_type_list INTO m_function_table_type_list[li_count].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET ls_table_type = m_function_table_type_list[li_count].p_orig_table_type
    LET ls_table_type = ls_table_type.subString(4,ls_table_type.getLength())
    LET m_function_table_type_list[li_count].p_table_type = ls_table_type

    LET li_count = li_count + 1

  END FOREACH
  CALL m_function_table_type_list.deleteElement(li_count)

  LET li_return = li_count - 1

  RETURN li_return

END FUNCTION

#抓取欄位定義群組(遞迴)
FUNCTION adzi140_drag_column_define_from_group(p_dialog,p_module_name,p_table_name,p_value,p_index,p_group_id) #20170425
DEFINE
  p_dialog      ui.Dialog,
  p_module_name STRING, #20170425
  p_table_name  STRING,
  p_value       STRING,
  p_index       INTEGER,
  p_group_id    STRING
TYPE
  T_DZEK_RECV RECORD
                dzek001 LIKE dzek_t.dzek001,
                dzek002 LIKE dzek_t.dzek002,
                dzek003 LIKE dzek_t.dzek003
              END RECORD
DEFINE
  ls_module_name   STRING, #20170425
  ls_table_name    STRING,
  ls_value         STRING,
  li_index         INTEGER,
  ls_sql           STRING,
  li_counts        INTEGER,
  li_for           INTEGER,
  ls_old_dzek001   STRING,
  ls_group_id      STRING,
  lo_dzek_recv     T_DZEK_RECV,
  lr_dzek_recv     DYNAMIC ARRAY OF T_DZEK_RECV

  LET ls_module_name = p_module_name #20170425
  LET ls_table_name  = p_table_name
  LET ls_value       = p_value
  LET li_index       = p_index
  LET ls_group_id    = p_group_id

  LET ls_sql = "select ek.dzek001,ek.dzek002,ek.dzek003 ",
               "  from dzek_t ek                        ",
               " where ek.dzek001 = '",ls_value,"'      ",
               "   and ek.dzek008 = 'Y'                 ",
               " order by ek.dzek006                    "

  PREPARE lpre_get_column_define_from_group FROM ls_sql
  DECLARE lcur_get_column_define_from_group CURSOR FOR lpre_get_column_define_from_group

  LET li_counts = 0;
  CALL lr_dzek_recv.clear()

  OPEN lcur_get_column_define_from_group
  FOREACH lcur_get_column_define_from_group INTO lo_dzek_recv.*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_counts = li_counts + 1;

    LET lr_dzek_recv[li_counts].dzek001 = lo_dzek_recv.dzek001
    LET lr_dzek_recv[li_counts].dzek002 = lo_dzek_recv.dzek002
    LET lr_dzek_recv[li_counts].dzek003 = lo_dzek_recv.dzek003

  END FOREACH
  CLOSE lcur_get_column_define_from_group

  FREE lpre_get_column_define_from_group
  FREE lcur_get_column_define_from_group

  LET ls_old_dzek001 = "NULL"
  IF ls_group_id IS NULL THEN
    CALL sadzi140_db_get_GUID() RETURNING ls_group_id
  END IF

  FOR li_for = 1 TO li_counts
    IF lo_dzek_recv.dzek003 IS NOT NULL THEN
      IF ls_old_dzek001 <> lo_dzek_recv.dzek001 THEN
        LET ls_old_dzek001 = lo_dzek_recv.dzek001
        CALL adzi140_drag_column_define_to_table(p_dialog,p_module_name,p_table_name,lr_dzek_recv[li_for].dzek001,p_index,ls_group_id,FALSE) #20170425
      END IF
    ELSE
      CALL adzi140_drag_column_define_from_group(p_dialog,p_module_name,p_table_name,lr_dzek_recv[li_for].dzek002,p_index,ls_group_id) #20170425
    END IF;
  END FOR

END FUNCTION

#將欄位定義的名稱抓到 Table 表格中
FUNCTION adzi140_drag_column_define_to_table(p_dialog,p_module_name,p_table_name,p_value,p_index,p_group_id,p_insert) #20170425
DEFINE
  p_dialog      ui.Dialog,
  p_module_name STRING, #20170425
  p_table_name  STRING,
  p_value       STRING,
  p_index       INTEGER,
  p_group_id    STRING,
  p_insert      BOOLEAN
DEFINE
  ls_module_name     STRING, #20170425
  ls_table_name      STRING,
  ls_value           STRING,
  li_index           INTEGER,
  ls_group_id        STRING,
  lb_insert          BOOLEAN,
  ls_sql             STRING,
  li_counts          INTEGER,
  ls_pre_column_name STRING,
  lo_dzek_t          T_DZEK_T_R,
  ls_field_name      STRING,
  ls_custom_code     STRING,
  li_max_srn         INTEGER, #20190304
  ls_dzeb027         STRING, #20190304
  lb_duplicate       BOOLEAN

  LET ls_module_name = p_module_name #20170425
  LET ls_table_name  = p_table_name
  LET ls_value       = p_value
  LET li_index       = p_index
  LET ls_group_id    = p_group_id
  LET lb_insert      = p_insert

  LET ls_pre_column_name = NVL(ls_table_name.subString(1,ls_table_name.getIndexOf(ms_table_tail_code,1)-1),ls_table_name)

  LET ls_sql = "select ek.dzek001,ek.dzek002,ek.dzek003,ek.dzek004,td.gztd002,           ",           
               "       td.gztd003,replace(replace(td.gztd008,' ',''),'.',',') gztd008,   ",           
               "       ekl.dzekl004 dzek005,ek.dzek006,ek.dzek007,ek.dzek008,ek.dzek009, ",           
               "       ek.dzek010,                                                       ",           
               "       case upper(trim(td.gztd003))                                      ",           
               "         when 'NUMBER' then '0'                                          ",           
               "         when 'NUMERIC' then '0'                                         ",           
               "         when 'DECIMAL' then '0'                                         ",           
               "       else                                                              ",           
               "         ''                                                              ",           
               "       end default_value,                                                ",  
               "       coalesce(ej.dzej004,ek.dzek004) dzej004                           ", #20170425               
               "  from dzek_t ek                                                         ",           
               "  left outer join gztd_t td on td.gztd001 = ek.dzek004                   ",           
               "  left outer join dzekl_t ekl on ekl.dzekl001 = ek.dzek001               ",           
               "                             and ekl.dzekl002 = ek.dzek002               ",           
               "                             and ekl.dzekl003 = '",ms_lang,"'            ", 
               "  left outer join dzej_t ej on ej.dzej001 = 'adzi140_data_type_mapping'  ", #20170425 
               "                           and ej.dzej003 = ek.dzek004                   ", #20170425           
               " where ek.dzek001 = '",ls_value,"'                                       ",           
               "   and ek.dzek008 = 'Y'                                                  ",            
               " order by ek.dzek006                                                     "                                                     

  PREPARE lpre_get_column_define_to_table FROM ls_sql
  DECLARE lcur_get_column_define_to_table CURSOR FOR lpre_get_column_define_to_table

  LET li_counts = 0;

  IF ls_group_id IS NULL THEN
    CALL sadzi140_db_get_GUID() RETURNING ls_group_id
  END IF

  #appendRow
  IF NOT lb_insert THEN
    LET li_index = m_dzeb_t.getLength() + 1
  END IF

  OPEN lcur_get_column_define_to_table
  FOREACH lcur_get_column_define_to_table INTO lo_dzek_t.*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_counts = li_counts + 1;

    CASE
      WHEN (lo_dzek_t.DZEK002 = "{SerialNumber}")
        #呼叫產出序列碼欄位
        CALL adzi140_gen_serial_number_column(p_dialog,ls_module_name,ls_pre_column_name,ls_value,li_index,lo_dzek_t.*,ls_group_id) #20170425
      WHEN (lo_dzek_t.DZEK002 = "{UserDefineVarchar}") OR
           (lo_dzek_t.DZEK002 = "{UserDefineNumber}") OR
           (lo_dzek_t.DZEK002 = "{UserDefineDateTime}")
        #呼叫產出自定義欄位
        CALL adzi140_gen_user_defined_column(p_dialog,ls_module_name,ls_pre_column_name,ls_value,li_index,lo_dzek_t.*,ls_group_id) #20170425
    OTHERWISE
      IF lb_insert THEN
        CALL p_dialog.insertRow("sr_tablecolumns", li_index)
        CALL p_dialog.setCurrentRow("sr_tablecolumns", li_index)
      ELSE
        CALL p_dialog.appendRow("sr_tablecolumns") #appendRow
      END IF

      #判斷表格是否為客制的,若是,則給Custom Code(uc),否則給Alter Code(ua)
      IF m_dzea_t[mi_dzea_index].dzea018 = cs_dgenv_standard THEN
        #20190304 begin
        IF m_dzea_t[mi_dzea_index].dzea014 = ms_topind THEN 
          LET ls_custom_code = ms_column_user_alter_code 
        ELSE
          LET ls_custom_code = IIF(ms_dgenv==cs_dgenv_customize,ms_column_user_alter_code,ms_column_industry_code) -- 20190304
        END IF
        #20190304 end  
      END IF

      LET ls_field_name = ls_pre_column_name,ls_custom_code,lo_dzek_t.DZEK002
      CALL adzi140_check_record_duplicate(p_dialog,ls_field_name,li_index) RETURNING lb_duplicate

      IF NOT lb_duplicate THEN
        CALL adzi140_get_max_serial_number(m_dzeb_t) RETURNING li_max_srn #20190304
        
        LET m_dzeb_t[li_index].RECORDTYPE   = cs_data_drag
        LET m_dzeb_t[li_index].GROUPID      = ls_group_id
        LET m_dzeb_t[li_index].DZEB002      = ls_field_name
        LET m_dzeb_t[li_index].DZEB003      = lo_dzek_t.DZEK005
        LET m_dzeb_t[li_index].DZEB004      = lo_dzek_t.DZEK003
        LET m_dzeb_t[li_index].DZEB005      = lo_dzek_t.DZEK003 --NVL(lo_dzek_t.DZEK009,"N") -- 20180903
        LET m_dzeb_t[li_index].DZEB006      = IIF(ls_module_name="ADZ",lo_dzek_t.DZEJ004,lo_dzek_t.DZEK004) #20170425
        LET m_dzeb_t[li_index].DZEB006_DESC = lo_dzek_t.GZTD002
        LET m_dzeb_t[li_index].DZEB012      = lo_dzek_t.DEFAULT_VALUE
        LET m_dzeb_t[li_index].DZEB022      = ls_value
        --LET m_dzeb_t[li_index].DZEB023      = li_max_srn + 1 #20190304
        CALL adzi140_get_dzeb027_json_data(ms_topind,m_dzeb_t[li_index].DZEB023,ls_custom_code) RETURNING ls_dzeb027 #20190304
        LET m_dzeb_t[li_index].DZEB027      = ls_dzeb027 #20190304

        CALL adzi140_on_change_dzeb006(m_dzeb_t[li_index].*) RETURNING m_dzeb_t[li_index].*

        LET li_index = li_index + 1
      END IF

    END CASE

    --IF lo_dzek_t.DZEK002 = "{SerialNumber}" THEN
      #呼叫產出序列碼欄位
      --CALL adzi140_gen_serial_number_column(p_dialog,ls_module_name,ls_pre_column_name,ls_value,li_index,lo_dzek_t.*,ls_group_id) #20170425
    --ELSE
      --IF lb_insert THEN
        --CALL p_dialog.insertRow("sr_tablecolumns", li_index)
        --CALL p_dialog.setCurrentRow("sr_tablecolumns", li_index)
      --ELSE
        --CALL p_dialog.appendRow("sr_tablecolumns") #appendRow
      --END IF
--
      --LET ls_field_name = ls_pre_column_name,lo_dzek_t.DZEK002
      --CALL adzi140_check_record_duplicate(p_dialog,ls_field_name,li_index) RETURNING lb_duplicate
--
      --IF NOT lb_duplicate THEN
        --LET m_dzeb_t[li_index].RECORDTYPE   = cs_data_drag
        --LET m_dzeb_t[li_index].GROUPID      = ls_group_id
        --LET m_dzeb_t[li_index].DZEB002      = ls_field_name
        --LET m_dzeb_t[li_index].DZEB003      = lo_dzek_t.DZEK005
        --LET m_dzeb_t[li_index].DZEB004      = lo_dzek_t.DZEK003
        --LET m_dzeb_t[li_index].DZEB005      = NVL(lo_dzek_t.DZEK009,"N")
        --LET m_dzeb_t[li_index].DZEB006      = IIF(ls_module_name="ADZ",lo_dzek_t.DZEJ004,lo_dzek_t.DZEK004) #20170425
        --LET m_dzeb_t[li_index].DZEB006_DESC = lo_dzek_t.GZTD002
        --LET m_dzeb_t[li_index].DZEB022      = ls_value
--
        --CALL adzi140_on_change_dzeb006(m_dzeb_t[li_index].*) RETURNING m_dzeb_t[li_index].*
--
        --LET li_index = li_index + 1
      --END IF
      --
    --END IF

  END FOREACH
  CLOSE lcur_get_column_define_to_table

  FREE lpre_get_column_define_to_table
  FREE lcur_get_column_define_to_table

END FUNCTION

FUNCTION adzi140_gen_serial_number_column(p_dialog,p_module_name,p_pre_column_name,p_value,p_index,p_dzek_t,p_group_id)
DEFINE
  p_dialog          ui.Dialog,
  p_module_name     STRING, #20170425
  p_pre_column_name STRING,
  p_value           STRING,
  p_index           INTEGER,
  p_dzek_t          T_DZEK_T_R,
  p_group_id        STRING
DEFINE
  lo_serial_number RECORD
                     serial_begin INTEGER,
                     serial_end   INTEGER
                   END RECORD
DEFINE
  ls_module_name     STRING, #20170425
  ls_pre_column_name STRING,
  ls_value           STRING,
  li_index           INTEGER,
  lo_dzek_t          T_DZEK_T_R,
  li_count           INTEGER,
  ls_group_id        STRING,
  ls_field_name      STRING,
  li_max_srn         INTEGER,
  li_curr_srn        INTEGER,
  li_serial_begin    INTEGER,
  li_serial_end      INTEGER,
  ls_custom_code     STRING,
  ls_dzeb027         STRING, #20190304
  lb_duplicate       BOOLEAN

  &ifndef DEBUG
  OPEN WINDOW w_sadzi140_srn WITH FORM cl_ap_formpath("ADZ","sadzi140_srn")
  &else
  OPEN WINDOW w_sadzi140_srn WITH FORM sadzi140_util_get_form_path("sadzi140_srn")
  &endif

  CALL sadzi140_util_set_form_title("sadzi140_srn",ms_lang) #20150417 by Hiko

  LET ls_module_name     = p_module_name #20170425
  LET ls_pre_column_name = p_pre_column_name
  LET ls_value           = p_value
  LET li_index           = p_index
  LET lo_dzek_t.*        = p_dzek_t.*
  LET ls_group_id        = p_group_id

  LET li_max_srn = 0
  LET li_curr_srn = 0

  DIALOG
    INPUT lo_serial_number.* FROM sr_SerialNumber.* ATTRIBUTE(WITHOUT DEFAULTS)
    END INPUT

    BEFORE DIALOG
      #抓取流水號最大值
      CALL adzi140_get_max_serial_number(m_dzeb_t) RETURNING li_max_srn #20190304
      
      LET li_serial_begin = li_max_srn + 1
      LET li_serial_end   = li_serial_begin + 1
      LET lo_serial_number.serial_begin = li_serial_begin
      LET lo_serial_number.serial_end   = li_serial_end

    ON ACTION btn_OK
      IF lo_serial_number.serial_end <= lo_serial_number.serial_begin THEN
        #序號迄值必須大於序號起值
        CALL sadzp000_msg_show_error('adz-00099', 'adz-00099', NULL, 1)
        CONTINUE DIALOG
      END IF
      
      IF ls_group_id IS NULL THEN
        CALL sadzi140_db_get_GUID() RETURNING ls_group_id
      END IF
      
      LET li_index = m_dzeb_t.getLength() + 1 #appendRow
      IF lo_serial_number.serial_begin IS NOT NULL AND lo_serial_number.serial_end IS NOT NULL THEN

        FOR li_count = lo_serial_number.serial_begin TO lo_serial_number.serial_end
          #CALL p_dialog.insertRow("sr_tablecolumns", li_index)
          #CALL p_dialog.setCurrentRow("sr_tablecolumns", li_index)
          CALL p_dialog.appendRow("sr_tablecolumns") #appendRow

          #判斷表格是否為客制的,若是,則給Custom Code(uc),否則給Alter Code(ua)
          IF m_dzea_t[mi_dzea_index].dzea018 = cs_dgenv_standard THEN
            #20190304 begin
            IF m_dzea_t[mi_dzea_index].dzea014 = ms_topind THEN 
              LET ls_custom_code = ms_column_user_alter_code 
            ELSE
              LET ls_custom_code = IIF(ms_dgenv==cs_dgenv_customize,ms_column_user_alter_code,ms_column_industry_code) -- 20190304
            END IF
            #20190304 end  
          END IF

          LET ls_field_name = ls_pre_column_name,ls_custom_code,(li_count USING lo_dzek_t.DZEK007)
          CALL adzi140_check_record_duplicate(p_dialog,ls_field_name,li_index) RETURNING lb_duplicate

          IF NOT lb_duplicate THEN
            LET m_dzeb_t[li_index].RECORDTYPE   = cs_data_drag
            LET m_dzeb_t[li_index].GROUPID      = ls_group_id
            LET m_dzeb_t[li_index].DZEB002      = ls_field_name
            LET m_dzeb_t[li_index].DZEB003      = lo_dzek_t.DZEK005
            LET m_dzeb_t[li_index].DZEB004      = lo_dzek_t.DZEK003
            LET m_dzeb_t[li_index].DZEB006      = IIF(ls_module_name="ADZ",lo_dzek_t.DZEJ004,lo_dzek_t.DZEK004) #20170425
            LET m_dzeb_t[li_index].DZEB006_DESC = lo_dzek_t.GZTD002
            --LET m_dzeb_t[li_index].DZEB007      = lo_dzek_t.GZTD003
            --LET m_dzeb_t[li_index].DZEB008      = lo_dzek_t.GZTD008
            LET m_dzeb_t[li_index].DZEB012      = lo_dzek_t.DEFAULT_VALUE
            LET m_dzeb_t[li_index].DZEB022      = ls_value
            LET m_dzeb_t[li_index].DZEB023      = li_count
            CALL adzi140_get_dzeb027_json_data(ms_topind,m_dzeb_t[li_index].DZEB023,ls_custom_code) RETURNING ls_dzeb027 #20190304
            LET m_dzeb_t[li_index].DZEB027      = ls_dzeb027 #20190304
            LET m_dzeb_t[li_index].DZEB029      = ms_dgenv
            LET m_dzeb_t[li_index].DZEB030      = ms_dgenv
            LET m_dzeb_t[li_index].DZEB031      = "N"
            LET li_index = li_index + 1 #appendRow
          END IF
        END FOR

      END IF
      
      CALL m_dzeb_t.deleteElement(li_index) #appendRow
      EXIT DIALOG

    ON ACTION btn_CANCEL
      EXIT DIALOG

    ON ACTION CLOSE
      EXIT DIALOG

  END DIALOG

  CLOSE WINDOW w_sadzi140_srn

END FUNCTION

FUNCTION adzi140_gen_user_defined_column(p_dialog,p_module_name,p_pre_column_name,p_value,p_index,p_dzek_t,p_group_id)
DEFINE
  p_dialog          ui.Dialog,
  p_module_name     STRING, #20170425
  p_pre_column_name STRING,
  p_value           STRING,
  p_index           INTEGER,
  p_dzek_t          T_DZEK_T_R,
  p_group_id        STRING
DEFINE
  lo_serial_number RECORD
                     serial_begin INTEGER,
                     serial_end   INTEGER
                   END RECORD
DEFINE
  ls_module_name     STRING, #20170425
  ls_pre_column_name STRING,
  ls_value           STRING,
  li_index           INTEGER,
  lo_dzek_t          T_DZEK_T_R,
  li_count           INTEGER,
  ls_group_id        STRING,
  ls_field_name      STRING,
  li_max_srn         INTEGER,
  li_curr_srn        INTEGER,
  li_serial_begin    INTEGER,
  li_serial_end      INTEGER,
  ls_number_fields   STRING,
  li_number_fields   INTEGER,
  ls_dzeb027         STRING, #20190304
  lb_duplicate       BOOLEAN,
  ls_column_user_define_code STRING

  LET ls_module_name     = p_module_name #20170425
  LET ls_pre_column_name = p_pre_column_name
  LET ls_value           = p_value
  LET li_index           = p_index
  LET lo_dzek_t.*        = p_dzek_t.*
  LET ls_group_id        = p_group_id

  LET li_max_srn = 0
  LET li_curr_srn = 0

  #抓取流水號最大值
  FOR li_count = 1 TO m_dzeb_t.getLength()
    IF (m_dzeb_t[li_count].dzeb023 IS NOT NULL) AND (
                                                     (m_dzeb_t[li_count].dzeb022 = cs_cdf_user_define_varchar) OR
                                                     (m_dzeb_t[li_count].dzeb022 = cs_cdf_user_define_number) OR
                                                     (m_dzeb_t[li_count].dzeb022 = cs_cdf_user_define_datetime)
                                                    ) THEN
      LET li_curr_srn = m_dzeb_t[li_count].dzeb023
      IF li_curr_srn > li_max_srn THEN
        LET li_max_srn = li_curr_srn
      END IF
    END IF
  END FOR

  #取得自定義欄位的中間設定碼
  CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_column_user_define_code) RETURNING ls_column_user_define_code
  #取得自定義欄位的應有欄位個數
  CASE
    WHEN (ls_value = cs_cdf_user_define_varchar)
      CALL sadzi140_db_get_parameter(cs_param_level_column_define,cs_param_number_of_varchar_fields) RETURNING ls_number_fields
    WHEN (ls_value = cs_cdf_user_define_number)
      CALL sadzi140_db_get_parameter(cs_param_level_column_define,cs_param_number_of_number_fields) RETURNING ls_number_fields
    WHEN (ls_value = cs_cdf_user_define_datetime)
      CALL sadzi140_db_get_parameter(cs_param_level_column_define,cs_param_number_of_datetime_fields) RETURNING ls_number_fields
  END CASE
  LET li_number_fields = NVL(ls_number_fields.trim(),"0")

  LET li_serial_begin = li_max_srn + 1
  LET li_serial_end   = li_serial_begin + li_number_fields - 1

  LET lo_serial_number.serial_begin = li_serial_begin
  LET lo_serial_number.serial_end   = li_serial_end

  IF ls_group_id IS NULL THEN
    CALL sadzi140_db_get_GUID() RETURNING ls_group_id
  END IF

  LET li_index = m_dzeb_t.getLength() + 1 #appendRow
  IF lo_serial_number.serial_begin IS NOT NULL AND lo_serial_number.serial_end IS NOT NULL THEN

    FOR li_count = lo_serial_number.serial_begin TO lo_serial_number.serial_end
      #CALL p_dialog.insertRow("sr_tablecolumns", li_index)
      #CALL p_dialog.setCurrentRow("sr_tablecolumns", li_index)
      CALL p_dialog.appendRow("sr_tablecolumns") #appendRow

      LET ls_field_name = ls_pre_column_name,ls_column_user_define_code,(li_count USING lo_dzek_t.DZEK007)
      CALL adzi140_check_record_duplicate(p_dialog,ls_field_name,li_index) RETURNING lb_duplicate

      IF NOT lb_duplicate THEN
        LET m_dzeb_t[li_index].RECORDTYPE = cs_data_drag
        LET m_dzeb_t[li_index].GROUPID      = ls_group_id
        LET m_dzeb_t[li_index].DZEB002      = ls_field_name
        LET m_dzeb_t[li_index].DZEB003      = lo_dzek_t.DZEK005,(li_count USING lo_dzek_t.DZEK007)
        LET m_dzeb_t[li_index].DZEB004      = lo_dzek_t.DZEK003
        LET m_dzeb_t[li_index].DZEB006      = IIF(ls_module_name="ADZ",lo_dzek_t.DZEJ004,lo_dzek_t.DZEK004) #20170425
        LET m_dzeb_t[li_index].DZEB006_DESC = lo_dzek_t.GZTD002
        LET m_dzeb_t[li_index].DZEB007      = lo_dzek_t.GZTD003
        LET m_dzeb_t[li_index].DZEB008      = lo_dzek_t.GZTD008
        LET m_dzeb_t[li_index].DZEB012      = lo_dzek_t.DEFAULT_VALUE
        LET m_dzeb_t[li_index].DZEB022      = ls_value
        LET m_dzeb_t[li_index].DZEB023      = li_count
        CALL adzi140_get_dzeb027_json_data(ms_topind,m_dzeb_t[li_index].DZEB023,NULL) RETURNING ls_dzeb027 #20190304
        LET m_dzeb_t[li_index].DZEB027      = ls_dzeb027 #20190304
        LET m_dzeb_t[li_index].DZEB029      = ms_dgenv
        LET m_dzeb_t[li_index].DZEB030      = ms_dgenv
        LET m_dzeb_t[li_index].DZEB031      = "N"
        LET li_index = li_index + 1 #appendRow
      END IF
    END FOR

  END IF
  CALL m_dzeb_t.deleteElement(li_index) #appendRow

END FUNCTION

FUNCTION adzi140_initial_combobox_sql()
DEFINE
  ls_customize_sql               STRING,
  ls_modules_tool_table_sql      STRING,
  ls_cust_modules_tool_table_sql STRING

  #Display Module name in main window left hand table
  LET ms_sql_mod_names = "select zo.gzzo001  module_name,                    ",
                         "       zo.gzzo001  module_desc                     ",
                         "  from gzzo_t zo                                   ",
                         "union all                                          ",
                         "select zj.gzzj001 module_name,                     ",
                         "       zj.gzzj001 module_desc                      ",
                         "  from gzzj_t zj                                   ",
                         " where not exists (                                ",
                         "                    select 1                       ",
                         "                      from gzzo_t zo               ",
                         "                     where zo.gzzo001 = zj.gzzj001 ",
                         "                  )                                ",
                         " order by 1                                        "

  #begin of module select on main window
  IF (ms_enable_customize = "Y") OR (ms_DGENV = cs_dgenv_customize) THEN
    LET ls_customize_sql = "union all                                          ",
                           "select zj.gzzj001                     module_name, ",
                           "       zj.gzzj001",ms_concat_sign,"'. '",
                                               ms_concat_sign,"zol.gzzol003 module_desc  ", #20191209
                           "  from gzzj_t zj                                   ",
                           "  left outer join gzzol_t zol                      ",
                           "               on zol.gzzol001 = zj.gzzj003        ",
                           "              and zol.gzzol002 = '",ms_lang,"'     ",
                           " where 1=1                                         ",
                           "   and zj.gzzj002 = 'ERP'                          ", #20170425
                           "   and not exists (                                ",
                           "                    select 1                       ",
                           "                      from gzzo_t zo               ",
                           "                     where zo.gzzo001 = zj.gzzj001 ",
                           "                  )                                "
  ELSE
    LET ls_customize_sql = ""
  END IF

  LET ms_sql_main_form_modules = "select zo.gzzo001                     module_name,           ",
                                 "       zo.gzzo001",ms_concat_sign,"'. '",
                                                     ms_concat_sign,"zol.gzzol003 module_desc  ", #20191209
                                 "  from (                                                     ", #20170425 begin
                                 "        select zo.*                                          ",
                                 "          from gzzo_t zo                                     ",       
                                 "         where 1=1                                           ",       
                                 "           and zo.gzzo002 = 'ERP'                            ",       
                                 "           and zo.gzzo007 = 'sd'                             ",
                                 "        union                                                ",
                                 "        select zo.*                                          ",
                                 "          from gzzo_t zo                                     ",       
                                 "         where 1=1                                           ",       
                                 "           and zo.gzzo002 = 'ERP'                            ",       
                                 "           and zo.gzzo007 <> 'sd'                            ",
                                 "           and (                                             ",
                                 "                 (zo.gzzo007 = '",ms_topind,"')              ",
                                 "                 or                                          ",
                                 "                 ('",ms_dgenv,"' = '",cs_dgenv_customize,"') ", 
                                 "               )                                             ", 
                                 "       ) zo                                                  ", #20170425 end
                                 "  left outer join gzzol_t zol                                ",
                                 "               on zol.gzzol001 = zo.gzzo001                  ",
                                 "              and zol.gzzol002 = '",ms_lang,"'               ",
                                 " where 1=1                                                   ",
                                 ls_customize_sql,                                           
                                 "union all                                                    ",
                                 "select '*'  module_name,                                     ",
                                 "       '*'  module_desc                                      ",
                                 "  from dual                                                  ",
                                 " order by 1                                                  "  
                                 
  #end of module select on main window

  #begin of module select on create table wizard window
  IF mb_tool_table THEN
    LET ls_modules_tool_table_sql = " and zo.gzzo001 in ('AZZ','ADZ') "
    LET ls_cust_modules_tool_table_sql = " and zj.gzzj003 in ('AZZ','ADZ') "
  ELSE
    LET ls_modules_tool_table_sql = ""
    LET ls_cust_modules_tool_table_sql = ""
  END IF

  LET ms_sql_modules = "select zo.gzzo001                     module_name,            ",
                       "       zo.gzzo001",ms_concat_sign,"'. '",
                                           ms_concat_sign,"zol.gzzol003 module_desc   ", #20191209
                       "  from (                                                      ", #20170425 begin
                       "        select zo.*                                           ",
                       "          from gzzo_t zo                                      ",       
                       "         where 1=1                                            ",       
                       "           and zo.gzzo002 = 'ERP'                             ",       
                       "           and zo.gzzo007 = 'sd'                              ",
                       "        union                                                 ",
                       "        select zo.*                                           ",
                       "          from gzzo_t zo                                      ",       
                       "         where 1=1                                            ",       
                       "           and zo.gzzo002 = 'ERP'                             ",       
                       "           and zo.gzzo007 <> 'sd'                             ",
                       "           and zo.gzzo007 = '",ms_topind,"'                   ",
                       "       ) zo                                                   ", #20170425 end
                       "  left outer join gzzol_t zol on zol.gzzol001 = zo.gzzo001    ",
                       "                             and zol.gzzol002 = '",ms_lang,"' ",
                       " where 1=1                                                    ",
                       "   and exists (                                               ",
                       "                select 1                                      ",
                       "                  from gzzt_t zt                              ",
                       "                 where zt.gzzt001 = zo.gzzo001                ",
                       "              )                                               ",
                       ls_modules_tool_table_sql,
                       " order by zo.gzzo001                                          "

  LET ms_sql_cust_modules = "select zj.gzzj001                     module_name,            ",
                            "       zj.gzzj001",ms_concat_sign,"'. '",
                                                ms_concat_sign,"zol.gzzol003 module_desc   ", #20191209
                            "  from gzzj_t zj                                              ",
                            "  left outer join gzzol_t zol                                 ",
                            "               on zol.gzzol001 = zj.gzzj003                   ",
                            "              and zol.gzzol002 = '",ms_lang,"'                ",
                            " where 1=1                                                    ",
                            "   and zj.gzzj002 = 'ERP'                                     ", #20170425
                            "   and not exists (                                           ",
                            "                    select 1                                  ",
                            "                      from gzzo_t zo                          ",
                            "                     where zo.gzzo001 = zj.gzzj001            ",
                            "                  )                                           ",
                            "   and zj.gzzj003 not in ('ADZ')                              ",
                            ls_cust_modules_tool_table_sql,
                            "union all                                                     ",
                            "select zo.gzzo001                     module_name,            ",
                            "       zo.gzzo001",ms_concat_sign,"'. '",
                                                ms_concat_sign,"zol.gzzol003 module_desc   ", #20191209
                            "  from gzzo_t zo                                              ",
                            "  left outer join gzzol_t zol on zol.gzzol001 = zo.gzzo001    ",
                            "                             and zol.gzzol002 = '",ms_lang,"' ",
                            " where zo.gzzo006 = 'E'                                       ",
                            " order by 1                                                   "
  #end of module select on create table wizard window

  LET ms_sql_industry = "select gzoi001                industry_type, ",
                        "       gzoi001",ms_concat_sign,"'. '",
                                         ms_concat_sign,"gzoi002 industry_desc  ",#20191209
                        "  from gzoi_t                                ",
                        " where gzoi001 <> 'sd'                       ",
                        " order by gzoi001                            "

  LET ms_sql_data_type = "select td.gztd001                   type_name,",
                         "       td.gztd001",ms_concat_sign,"'. '",
                                             ms_concat_sign,"td.gztd002 type_desc ", #20191209
                         "  from gztd_t td                              ",
                         " order by td.gztd001                          "

{
  LET ms_sql_table_type = "SELECT EJ.DZEJ003                   TABLE_TYPES,     ",
                          "       EJ.DZEJ003",ms_concat_sign,"'. '",ms_concat_sign,"EJ.DZEJ004 TABLE_TYPES_DESC ",
                          "  FROM DZEJ_T EJ                                     ",
                          " WHERE EJ.DZEJ001 = 'adzi140_TableTypes'             ",
                          " ORDER BY EJ.DZEJ001,EJ.DZEJ002                      "
}

  LET ms_sql_table_type = "select cb.gzcb002                     table_types,            ",
                          "       cb.gzcb002",ms_concat_sign,"'. '",
                                              ms_concat_sign,"cbl.gzcbl004 table_types_desc ", #20191209
                          "  from gzcb_t cb                                              ",
                          "  left outer join gzcbl_t cbl on cbl.gzcbl001 = cb.gzcb001    ",
                          "                             and cbl.gzcbl002 = cb.gzcb002    ",
                          "                             and cbl.gzcbl003 = '",ms_lang,"' ",
                          " where cb.gzcb001 = '256'                                     ",
                          "   and cb.gzcb002 not in ('V')                                ", -- 排除提速檔
                          " order by cb.gzcb002                                          "

  LET ms_sql_key_type = "select ej.dzej002 key_types,           ",
                        "       ej.dzej003 key_types_desc       ",
                        "  from dzej_t ej                       ",
                        " where ej.dzej001 = 'adzi140_Key_types'",
                        "   and ej.dzej005 = 'Y'                ",
                        " order by ej.dzej001                   "

  LET ms_sql_index_type = "select ej.dzej002 index_types,           ",
                          "       ej.dzej003 index_types_desc       ",
                          "  from dzej_t ej                         ",
                          " where ej.dzej001 = 'adzi140_Index_types'",
                          " order by ej.dzej001                     "

  LET ms_sql_genero_widgets =  "select ej.dzej002 widgets_types,                        ",
                               "       ej.dzej002",ms_concat_sign,"'. '",
                                                   ms_concat_sign,"ej.dzej003 widgets_types_desc  ", #20191209
                               "  from dzej_t ej                                        ",
                               " where ej.dzej001 = 'GENERO_WIDGETS'                    ",
                               " order by ej.dzej002                                    "

  LET ms_sql_erp_schema_list = "select da.gzda001 db_name,                               ",
                               "       da.gzda001 db_desc                                ",
                               "  from gzda_t da                                         ",
                               " where da.gzdastus = 'Y'                                 ",
                               "   and da.gzda005  = 'Y'                                 ",
                               "   and (                                                 ",
                               "         exists (                                        ",
                               "                 select 1                                ",
                               "                   from dzeu_t eu                        ",
                               "                  where eu.dzeu002 = da.gzda001          ",
                               "                    and eu.dzeu001 = '",cs_table_tag,"'  ",
                               "                    and eu.dzeu003 = 'T'                 ",
                               "                )                                        ",
                               "         or                                              ", #20180903 begin
                               "         exists (                                        ",
                               "                  select 1                               ",
                               "                    from dzen_t en                       ",
                               "                   where en.dzen001 = '",cs_table_tag,"' ",
                               "                     and en.dzen002 = da.gzda001         ",
                               "                )                                        ",
                               "       )                                                 ",        
                               " order by 1                                              "  #20180903 end

  LET ms_sql_all_schema_list = "select da.gzda001 db_name,             ",
                               "       da.gzda001 db_desc              ",
                               "  from gzda_t da                       ",
                               " where da.gzdastus = 'Y'               ",
                               " order by 1                            "

END FUNCTION

FUNCTION adzi140_get_system_information(p_set_title)
DEFINE
  p_set_title BOOLEAN
DEFINE
  lb_set_title     BOOLEAN,
  ls_zone          STRING,
  ls_cust          STRING,
  ls_user          STRING,
  ls_window_title  STRING,
  lo_db_info       T_DB_INFO,
  ls_program_title STRING,
  ls_about         STRING,
  ls_mode          STRING,
  ls_erpid         STRING,
  ls_topind        STRING,
  ls_DB_TYPE       STRING,
  ls_branch        STRING, #20230606 
  ls_drop_backdoor_col STRING,
  lo_session_data  T_SESSION_DATA, #20170912 
  ls_os_process_id STRING, #20170912
  ldt_curr_time    DATETIME YEAR TO SECOND
DEFINE
  ls_return STRING

  LET lb_set_title = p_set_title

  CALL FGL_GETENV("ZONE") RETURNING ls_zone
  CALL FGL_GETENV("CUST") RETURNING ls_cust
  CALL FGL_GETENV("ERPID") RETURNING ls_erpid
  CALL sadzi140_util_get_db_type() RETURNING ls_DB_TYPE #20200330 
  CALL sadzi140_util_get_pgs_branch_type() RETURNING ls_branch #20230606
  LET ls_DB_TYPE = IIF(ls_DB_TYPE = cs_db_driver_type_pgs,ls_DB_TYPE||'('||ls_branch||')',ls_DB_TYPE) #20230606
  
  LET ls_topind = NVL(ms_topind,FGL_GETENV("TOPIND"))
  CALL FGL_GETENV(cs_env_RT_DROP_BACKDOOR_COLUMN) RETURNING ls_drop_backdoor_col

  &ifndef DEBUG
  LET ldt_curr_time = cl_get_current()
  &else
  LET ldt_curr_time = CURRENT YEAR TO SECOND
  &endif

  #For get Windows information
  LET ls_user = ms_user

  CALL sadzi140_db_get_db_info() RETURNING lo_db_info.*
  CALL sadzi140_util_get_program_title('adzi140',ms_lang) RETURNING ls_program_title
  CALL sadzi140_db_get_connect_session_data() RETURNING lo_session_data.* #20170912
  LET ls_os_process_id = lo_session_data.sd_PROCESS #20170912

  LET ls_zone = ls_zone
  LET ls_user = ms_user

  IF lb_set_title THEN
    LET ls_window_title = NVL(ls_program_title,cs_program_title)," [ Cust：",ls_cust," ] [ Zone：",ls_zone," ] [ User：",ls_user," ] [DB Name：",lo_db_info.DB_NAME,"] [DB User：",lo_db_info.USER_NAME,"] [ Login Time：",CURRENT YEAR TO SECOND," ]"
    CALL FGL_SETTITLE(ls_window_title)
  END IF

  #Execution Mode
  CASE
    WHEN mb_patch
      LET ls_mode = "Patch"
    WHEN mb_debug
      LET ls_mode = "Debug"
    OTHERWISE
      LET ls_mode = "Normal"
  END CASE

  LET ls_about = "Table Designer：",cs_version,"\n",
                 "Build date：",cs_modify_date,"\n\n",
                 "Mode：",ls_mode,"\n",
                 "Environment：",IIF(ms_dgenv=cs_dgenv_standard,"Standard","Customize"),"\n",
                 "Cust：",ls_cust,"\n",
                 "Zone：",ls_zone,"\n",
                 "ERPID：",ls_erpid,"\n",
                 "ERPVER：",ms_erpver,"\n", #20180420
                 "TOPIND：",NVL(ls_topind,cs_default_topind),"\n",
                 "DB type：",ls_DB_TYPE,"\n",
                 "DB name：",lo_db_info.DB_NAME,"\n",
                 "DB user：",lo_db_info.USER_NAME,"\n",
                 "OS process ID：",ls_os_process_id,"\n", #20170912
                 "Drop backdoor columns：",NVL(ls_drop_backdoor_col,"N"),"\n",
                 "Login user：",ls_user,"\n",
                 "Login time：",mdt_login_time,"\n",
                 "Current time：",ldt_curr_time

  LET ls_return = ls_about

  RETURN ls_return

END FUNCTION

#判斷欄位是否重複
FUNCTION adzi140_check_record_duplicate(p_dialog,p_field_name,p_index)
DEFINE
  p_dialog     ui.Dialog,
  p_field_name STRING,
  p_index      INTEGER
DEFINE
  ls_field_name  STRING,
  li_index       INTEGER,
  li_rec_cnt     INTEGER,
  ls_question    STRING,
  lb_delete      BOOLEAN,
  ls_replace_arg STRING

  LET ls_field_name  = p_field_name
  LET li_index       = p_index

  LET lb_delete = FALSE

  FOR li_rec_cnt = 1 TO m_old_dzeb_t.getLength()
    IF m_old_dzeb_t[li_rec_cnt].DZEB002 = ls_field_name THEN
      LET ls_replace_arg = ls_field_name,"|"
      CALL sadzp000_msg_show_error("adz-00116", "adz-00116", ls_replace_arg, 1)
      LET lb_delete = TRUE
      CALL p_dialog.deleteRow("sr_tablecolumns", li_index)
    END IF
  END FOR

  RETURN lb_delete

END FUNCTION

FUNCTION adzi140_get_foreign_key_data(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name STRING,
  ls_sql        STRING,
  ls_key_list    VARCHAR(1024)

  LET ls_table_name  = p_table_name

  LET ls_sql = "select ed.dzed004 keylist              ",
               "  from dzed_t ed                       ",
               " where 1=1                             ",
               "   and ed.dzed003 = 'P'                ",
               "   and ed.dzed001 = '",ls_table_name,"'"

  PREPARE lpre_get_foreign_key_data FROM ls_sql
  DECLARE lcur_get_foreign_key_data CURSOR FOR lpre_get_foreign_key_data

  LET ls_key_list = ""

  OPEN lcur_get_foreign_key_data
  FETCH lcur_get_foreign_key_data INTO ls_key_list
  CLOSE lcur_get_foreign_key_data

  FREE lpre_get_foreign_key_data
  FREE lcur_get_foreign_key_data

  RETURN ls_key_list

END FUNCTION

FUNCTION adzi140_get_key_type_desc(p_key_type)
CONSTANT lcs_keyword STRING = "KEY"
DEFINE
  p_key_type STRING
DEFINE
  ls_key_type      STRING,
  ls_key_type_desc STRING,
  ls_return        STRING

  LET ls_key_type = p_key_type

  CASE ls_key_type.toUpperCase()
    WHEN "P"
      LET ls_key_type_desc = "PRIMARY"
    WHEN "F"
      LET ls_key_type_desc = "FOOTING"
    WHEN "R"
      LET ls_key_type_desc = "REFERENCED"
    WHEN "U"
      LET ls_key_type_desc = "UNIQUE"
  OTHERWISE
    LET ls_key_type_desc = ""
  END CASE

  LET ls_return = ls_key_type_desc," ",lcs_keyword

  RETURN ls_return

END FUNCTION

FUNCTION adzi140_get_index_type_desc(p_index_type)
DEFINE
  p_index_type STRING
DEFINE
  ls_index_type      STRING,
  ls_index_type_desc STRING,
  ls_return          STRING

  LET ls_index_type = p_index_type

  CASE ls_index_type.toUpperCase()
    WHEN "N"
      LET ls_index_type_desc = "NORMAL"
    WHEN "U"
      LET ls_index_type_desc = "UNIQUE"
  OTHERWISE
    LET ls_index_type_desc = ""
  END CASE

  LET ls_return = ls_index_type_desc

  RETURN ls_return

END FUNCTION

FUNCTION adzi140_create_table_synonym_data(p_schema_type)
DEFINE
  p_schema_type  DYNAMIC ARRAY OF T_TABLE_SYNONYM
DEFINE
  lo_schema_type DYNAMIC ARRAY OF T_TABLE_SYNONYM,
  li_counts      INTEGER,
  li_rec_cnt     INTEGER,
  lb_result      BOOLEAN

  LET lo_schema_type.* = p_schema_type.*

  BEGIN WORK

  LET li_counts = 0
  LET lb_result = TRUE

  FOR li_rec_cnt = 1 TO lo_schema_type.getLength()

    TRY
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      INSERT INTO dzeu_t(
                          dzeu001,dzeu002,dzeu003,dzeu004,dzeu005
                        )
                 VALUES (
                          lo_schema_type[li_rec_cnt].p_table_name,
                          lo_schema_type[li_rec_cnt].p_ent_db,
                          lo_schema_type[li_rec_cnt].p_sch_type,
                          lo_schema_type[li_rec_cnt].p_checkbox,
                          lo_schema_type[li_rec_cnt].p_row_number
                        )
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      TRY
        &ifdef is_pgs
        SAVEPOINT sp_tran
        &endif
        UPDATE dzeu_t
           SET dzeu003 = lo_schema_type[li_rec_cnt].p_sch_type,
               dzeu004 = lo_schema_type[li_rec_cnt].p_checkbox,
               dzeu005 = lo_schema_type[li_rec_cnt].p_row_number
         WHERE dzeu001 = lo_schema_type[li_rec_cnt].p_table_name
           AND dzeu002 = lo_schema_type[li_rec_cnt].p_ent_db
      CATCH
        &ifdef is_pgs
        ROLLBACK WORK TO SAVEPOINT sp_tran
        &endif
        LET lb_result = FALSE
      END TRY
    END TRY

    LET li_counts = li_counts + 1
    IF li_counts >= 100 THEN
      COMMIT WORK
      BEGIN WORK
      LET li_counts = 0
    END IF

  END FOR
  COMMIT WORK

  RETURN lb_result

END FUNCTION

FUNCTION adzi140_create_auxilary_table(p_function_table_type_list,p_alm_version,p_alm_sd_version,p_alm_request_no,p_dgenv)
DEFINE
  p_function_table_type_list T_FUNCTION_TABLE_TYPE_LIST,
  p_alm_version              STRING,
  p_alm_sd_version           STRING,
  p_alm_request_no           STRING,
  p_dgenv                    STRING
DEFINE
  lo_function_table_type_list T_FUNCTION_TABLE_TYPE_LIST,
  ls_alm_version              STRING,
  ls_alm_sd_version           STRING,
  ls_alm_request_no           STRING,
  ls_dgenv                    STRING
DEFINE
  lo_aux_table_base RECORD
                      dzeb001 LIKE dzeb_t.dzeb001,
                      dzeb002 LIKE dzeb_t.dzeb002,
                      dzeb003 LIKE dzeb_t.dzeb003,
                      dzeb004 LIKE dzeb_t.dzeb004,
                      dzeb005 LIKE dzeb_t.dzeb005,
                      dzeb006 LIKE dzeb_t.dzeb006,
                      dzeb007 LIKE dzeb_t.dzeb007,
                      dzeb008 LIKE dzeb_t.dzeb008,
                      dzeb012 LIKE dzeb_t.dzeb012,
                      dzeb021 LIKE dzeb_t.dzeb021,
                      dzeb022 LIKE dzeb_t.dzeb022,
                      dzeb023 LIKE dzeb_t.dzeb023,
                      dzeb024 LIKE dzeb_t.dzeb024,
                      dzeb028 LIKE dzeb_t.dzeb028,
                      dzeb029 LIKE dzeb_t.dzeb029,
                      dzeb030 LIKE dzeb_t.dzeb030,
                      dzeb031 LIKE dzeb_t.dzeb031
                    END RECORD
DEFINE
  lo_t_dzeb_t                 T_DZEB_T,
  lo_dzlm_t                   T_DZLM_T,
  lo_create_table             T_DZEA_CREATE_TABLE,
  lo_DZEA_info                T_DZEA_INFO
DEFINE
  ls_pre_table_name      STRING,
  ls_table_name          STRING,
  ls_sql                 STRING,
  ls_column_name         STRING,
  ls_front_word          STRING,
  ls_table_type          STRING,
  li_max_column_value    INTEGER,
  li_max_value           INTEGER,
  li_max_sequence        INTEGER,
  li_max_column_sequence INTEGER,
  li_column_seq          INTEGER,
  li_column_name_seq     INTEGER,
  ls_key_list            STRING,
  ls_digit_numbers       STRING,
  li_count               INTEGER,
  lo_schema_type         DYNAMIC ARRAY OF T_TABLE_SYNONYM,
  ls_all_message         STRING,
  ls_version             STRING,
  li_loop                INTEGER,
  lb_result              BOOLEAN,
  lb_gen_tbl             BOOLEAN, #20170908
  ls_gen_list            STRING, #20170908
  lb_return_val          BOOLEAN,
  lb_success             BOOLEAN,
  ls_master_db           STRING,
  ls_master_user         STRING,
  lb_error               BOOLEAN,
  ls_user                STRING,
  ls_replace_arg         STRING,
  ls_message             STRING,
  ls_schema_sql          STRING,
  ls_lang                STRING,
  ls_pre_table_lead_code STRING,
  lo_column_info         DYNAMIC ARRAY OF T_COLUMN_INFO, #20171120
  lo_lang_arr            DYNAMIC ARRAY OF T_LANGUAGE_TYPE
DEFINE
  ls_return  STRING #20170425  

  LET lo_function_table_type_list.* = p_function_table_type_list.*
  LET ls_alm_version                = p_alm_version
  LET ls_alm_sd_version             = p_alm_sd_version
  LET ls_alm_request_no             = p_alm_request_no
  LET ls_dgenv                      = p_dgenv

  LET ls_key_list        = ""
  LET li_column_seq      = 1
  LET li_column_name_seq = 0
  LET ls_digit_numbers   = ""
  LET ls_user            = ms_user

  #取得數字位數
  CALL adzi140_get_digit_number() RETURNING ls_digit_numbers

  ##############################################################################

  BEGIN WORK

  #抓取上層表格
  LET ls_pre_table_name = lo_function_table_type_list.p_pre_table_name
  LET ls_table_name     = lo_function_table_type_list.p_table_name
  CALL sadzi140_db_get_DZEA_info(ls_pre_table_name) RETURNING lo_DZEA_info.*

  LET lo_create_table.dct_table_name            = lo_function_table_type_list.p_table_name
  LET lo_create_table.dct_table_description     = lo_function_table_type_list.p_table_type_desc
  LET lo_create_table.dct_module_name           = lo_function_table_type_list.p_module_name
  LET lo_create_table.dct_table_type            = lo_function_table_type_list.p_table_type
  LET lo_create_table.dct_is_multi_lang_table   = "N"
  LET lo_create_table.dct_is_system_table       = "N"
  LET lo_create_table.dct_create_notice_trigger = "N" #20190424
  LET lo_create_table.dct_define_group          = lo_function_table_type_list.p_orig_table_type
  LET lo_create_table.dct_up_level_table        = lo_function_table_type_list.p_pre_table_name
  LET lo_create_table.dct_table_space           = ""
  LET lo_create_table.dct_is_altered            = "Y"
  LET lo_create_table.dct_industry_type         = lo_DZEA_info.DZEA014
  LET lo_create_table.dct_alm_construct_version = ls_alm_version
  LET lo_create_table.dct_alm_sd_version        = ls_alm_sd_version
  LET lo_create_table.dct_alm_request_no        = ls_alm_request_no
  LET lo_create_table.dct_dgenv                 = ls_dgenv
  LET lo_create_table.dct_orig_module_name      = lo_function_table_type_list.p_module_name
  LET lo_create_table.dct_shipping_notice       = "N"
  LET lo_create_table.dct_orig_dgenv            = ls_dgenv

  #將資料新增到主檔 DZEA_T 中
  CALL sadzi140_crud_insert_dzea_t(lo_create_table.*) RETURNING lb_success
  IF NOT lb_success THEN
    ROLLBACK WORK
    LET ls_replace_arg = lo_create_table.dct_table_name,"|"
    CALL sadzp000_msg_show_error('adz-00095', 'adz-00095', ls_replace_arg, 1)
  ELSE
    CALL sadzi140_crud_insert_update_dzeal_t(lo_create_table.*,ms_lang)
    COMMIT WORK
  END IF
  ##############################################################################

  IF lb_success THEN
    ## 建立表格 ####################################################################
    #以上層表格為條件抓取其為Table或是Synonym
    CALL sadzi140_db_get_main_table_synonym_data(ls_pre_table_name) RETURNING lo_schema_type
    FOR li_loop = 1 TO lo_schema_type.getLength()
      LET lo_schema_type[li_loop].p_table_name = ls_table_name
    END FOR
    CALL adzi140_create_table_synonym_data(lo_schema_type) RETURNING lb_result
    LET ls_all_message = ""
    LET ls_master_db   = ms_master_db
    LET ls_master_user = ms_master_user
    LET lo_create_table.dct_master_db   = NVL(ls_master_db,cs_master_db)
    LET lo_create_table.dct_master_user = NVL(ls_master_user,cs_master_user)
    CALL sadzi140_util_create_real_table(ms_master_user,lo_create_table.*,ls_version,ls_alm_version,ls_alm_request_no,ls_dgenv,TRUE) RETURNING ls_all_message
    ##############################################################################

    ################### 抓取原表格鍵值欄位並新增到多語言或是提速檔資料中 ##################
    BEGIN WORK

    LET li_count = 1
    LET ls_key_list = ""

    LET ls_sql = "select eb.dzeb001,eb.dzeb002,eb.dzeb003,eb.dzeb004,eb.dzeb005,",
                 "       eb.dzeb006,eb.dzeb007,eb.dzeb008,eb.dzeb012,eb.dzeb021,",
                 "       eb.dzeb022,eb.dzeb023,eb.dzeb024,'N','",ls_dgenv,"'    ",
                 "  from dzeb_t eb                                              ",
                 "  left outer join gztd_t td on td.gztd001 = eb.dzeb006        ",
                 " where eb.dzeb001 = '",ls_pre_table_name,"'                   ",
                 "   and eb.dzeb004 = 'Y'                                       ",
                 " order by eb.dzeb021                                          "

    PREPARE lpre_CreateAuxTable FROM ls_sql
    DECLARE lcur_CreateAuxTable CURSOR FOR lpre_CreateAuxTable

    FOREACH lcur_CreateAuxTable INTO lo_aux_table_base.*
      IF SQLCA.sqlcode THEN
        EXIT FOREACH
      END IF

      LET ls_front_word = lo_function_table_type_list.p_front_word
      LET ls_pre_table_lead_code = ls_pre_table_name.subString(1,ls_pre_table_name.getIndexOf(ms_table_tail_code,1)-1)

      LET ls_column_name = lo_aux_table_base.DZEB002

      LET ls_column_name = ls_column_name.subString(ls_pre_table_lead_code.getLength()+1,ls_column_name.getLength())
      IF ls_column_name MATCHES "*[0-9]" THEN
        LET li_column_name_seq = ls_column_name
      END IF
      #LET ls_column_name = (li_column_seq USING ls_digit_numbers)

      LET lo_aux_table_base.DZEB001 = lo_function_table_type_list.p_table_name
      LET lo_aux_table_base.DZEB002 = ls_front_word,lo_function_table_type_list.p_aux_type,
                                      ms_column_user_custom_code,
                                      ls_column_name
      LET lo_aux_table_base.DZEB021 = li_column_seq
      LET lo_aux_table_base.DZEB030 = ls_dgenv
      LET lo_aux_table_base.DZEB031 = "N"

      --LET ls_key_list = ls_key_list,",",lo_aux_table_base.DZEB002

      LET lo_t_dzeb_t.DZEB001 = lo_aux_table_base.DZEB001
      LET lo_t_dzeb_t.DZEB002 = lo_aux_table_base.DZEB002
      LET lo_t_dzeb_t.DZEB003 = lo_aux_table_base.DZEB003
      LET lo_t_dzeb_t.DZEB004 = lo_aux_table_base.DZEB004
      LET lo_t_dzeb_t.DZEB005 = lo_aux_table_base.DZEB005
      LET lo_t_dzeb_t.DZEB006 = lo_aux_table_base.DZEB006
      LET lo_t_dzeb_t.DZEB007 = lo_aux_table_base.DZEB007
      LET lo_t_dzeb_t.DZEB008 = lo_aux_table_base.DZEB008
      LET lo_t_dzeb_t.DZEB012 = lo_aux_table_base.DZEB012
      LET lo_t_dzeb_t.DZEB021 = lo_aux_table_base.DZEB021
      LET lo_t_dzeb_t.DZEB022 = lo_aux_table_base.DZEB022
      LET lo_t_dzeb_t.DZEB023 = lo_aux_table_base.DZEB023
      LET lo_t_dzeb_t.DZEB024 = lo_aux_table_base.DZEB024
      LET lo_t_dzeb_t.DZEB028 = lo_aux_table_base.DZEB028
      LET lo_t_dzeb_t.DZEB029 = lo_aux_table_base.DZEB029
      LET lo_t_dzeb_t.DZEB030 = lo_aux_table_base.DZEB030
      LET lo_t_dzeb_t.DZEB031 = lo_aux_table_base.DZEB031
      LET lo_t_dzeb_t.lang_modified = cs_lang_modified

      #將資料新增到明細檔 DZEB_T, DZEP_T 中
      CALL sadzi140_crud_insert_update_dzeb_t(lo_t_dzeb_t.*)
      #新增到欄位規格設計資料中
      CALL sadzi140_crud_insert_update_dzep_t(lo_t_dzeb_t.*,m_common_fields,ms_lang)
      #新增到欄位參考設定中
      CALL sadzi140_crud_insert_update_dzef_t(lo_t_dzeb_t.*,m_common_fields)
      #修改多語言資料
      CALL sadzi140_crud_insert_update_dzebl_t(lo_t_dzeb_t.*,ms_lang)
      #新增狀態相關資料
      CALL sadzi140_crud_insert_update_gzcc_t(lo_t_dzeb_t.*)

      LET li_count = li_count + 1
      LET li_column_seq = li_column_seq + 1
    END FOREACH

    COMMIT WORK
    ##############################################################################

    ############## 抓取多語言或是提速檔設定並新增到多語言或是提速檔明細資料中 ###############
    BEGIN WORK

    LET ls_table_type = ""
    LET li_count = 1

    IF li_column_name_seq = 0 OR li_column_name_seq IS NULL THEN
      LET li_column_name_seq = 1
    ELSE
      LET li_column_name_seq = li_column_name_seq + 1
    END IF

    #CALL adzi140_get_max_column_number(lo_function_table_type_list.*) RETURNING li_max_value
    #CALL adzi140_get_max_column_sequence(lo_function_table_type_list.*) RETURNING li_max_sequence

    IF lo_function_table_type_list.p_orig_table_type = cs_group_leading_code||"L" THEN
      LET ls_table_type = "lng"
    END IF

    IF lo_function_table_type_list.p_orig_table_type = cs_group_leading_code||"V" THEN
      LET ls_table_type = "spd"
    END IF

    LET ls_sql = "select '','',ekl.dzekl004 dzek005,ek.dzek003,ek.dzek003,      ",
                 "       ek.dzek004,td.gztd003,td.gztd008,'','',                ",
                 "       '",cs_cdf_serial_number,"','','','N','",ls_dgenv,"'    ",
                 "  from dzek_t ek                                              ",
                 "  left outer join dzekl_t ekl on ekl.dzekl001 = ek.dzek001    ",
                 "                             and ekl.dzekl002 = ek.dzek002    ",
                 "                             and ekl.dzekl003 = '",ms_lang,"' ",
                 "  left outer join gztd_t td on td.gztd001 = ek.dzek004        ",
                 " where ek.dzek002 = '{MaxNumber}'                             ",
                 "   and ek.dzek001 like '",ls_table_type,"%'                   ",
                 " order by ek.dzek006                                          "

    PREPARE lpre_CreateLangSpdColumn FROM ls_sql
    DECLARE lcur_CreateLangSpdColumn CURSOR FOR lpre_CreateLangSpdColumn

    FOREACH lcur_CreateLangSpdColumn INTO lo_aux_table_base.*
      IF SQLCA.sqlcode THEN
        EXIT FOREACH
      END IF

      #LET li_max_column_value    = li_max_value + li_count
      #LET li_max_column_sequence = li_max_sequence + li_count

      LET ls_front_word = lo_function_table_type_list.p_front_word

      LET lo_aux_table_base.DZEB001 = lo_function_table_type_list.p_table_name
      #LET lo_aux_table_base.DZEB002 = ls_front_word,lo_function_table_type_list.p_aux_type,
      #                                (li_max_column_value USING lo_aux_table_base.DZEB024)
      LET lo_aux_table_base.DZEB002 = ls_front_word,lo_function_table_type_list.p_aux_type,
                                      ms_column_user_custom_code,
                                      (li_column_name_seq USING ls_digit_numbers)
      #LET lo_aux_table_base.DZEB021 = li_max_column_sequence
      LET lo_aux_table_base.DZEB021 = li_column_seq
      LET lo_aux_table_base.DZEB030 = ls_dgenv
      LET lo_aux_table_base.DZEB031 = "N"

      {
      IF NVL(lo_aux_table_base.DZEB004,cs_null_value) = "Y" THEN
        LET ls_key_list = ls_key_list,",",lo_aux_table_base.DZEB002
      END IF
      }

      LET lo_t_dzeb_t.DZEB001 = lo_aux_table_base.DZEB001
      LET lo_t_dzeb_t.DZEB002 = lo_aux_table_base.DZEB002
      LET lo_t_dzeb_t.DZEB003 = lo_aux_table_base.DZEB003
      LET lo_t_dzeb_t.DZEB004 = lo_aux_table_base.DZEB004
      LET lo_t_dzeb_t.DZEB005 = lo_aux_table_base.DZEB005
      LET lo_t_dzeb_t.DZEB006 = lo_aux_table_base.DZEB006
      LET lo_t_dzeb_t.DZEB007 = lo_aux_table_base.DZEB007
      LET lo_t_dzeb_t.DZEB008 = lo_aux_table_base.DZEB008
      LET lo_t_dzeb_t.DZEB012 = lo_aux_table_base.DZEB012
      LET lo_t_dzeb_t.DZEB021 = lo_aux_table_base.DZEB021
      LET lo_t_dzeb_t.DZEB022 = lo_aux_table_base.DZEB022
      LET lo_t_dzeb_t.DZEB023 = lo_aux_table_base.DZEB023
      LET lo_t_dzeb_t.DZEB024 = lo_aux_table_base.DZEB024
      LET lo_t_dzeb_t.DZEB028 = lo_aux_table_base.DZEB028
      LET lo_t_dzeb_t.DZEB029 = lo_aux_table_base.DZEB029
      LET lo_t_dzeb_t.DZEB030 = lo_aux_table_base.DZEB030
      LET lo_t_dzeb_t.DZEB031 = lo_aux_table_base.DZEB031
      LET lo_t_dzeb_t.lang_modified = cs_lang_modified

      #將資料新增到明細檔 DZEB_T, DZEP_T 中
      CALL sadzi140_crud_insert_update_dzeb_t(lo_t_dzeb_t.*)
      #新增到欄位規格設計資料中
      CALL sadzi140_crud_insert_update_dzep_t(lo_t_dzeb_t.*,m_common_fields,ms_lang)
      #新增到欄位參考設定中
      CALL sadzi140_crud_insert_update_dzef_t(lo_t_dzeb_t.*,m_common_fields)
      #修改多語言資料
      CALL sadzi140_crud_insert_update_dzebl_t(lo_t_dzeb_t.*,ms_lang)
      #新增狀態相關資料
      CALL sadzi140_crud_insert_update_gzcc_t(lo_t_dzeb_t.*)

      LET li_count = li_count + 1
      LET li_column_seq = li_column_seq + 1
      LET li_column_name_seq = li_column_name_seq + 1
    END FOREACH

    COMMIT WORK

    #重Load一次表格欄位
    CALL sadzi140_db_get_column_with_sequence(ls_table_name) RETURNING lo_column_info
    #檢核並刪除無效的Key或索引資料
    CALL sadzi140_db_check_and_remove_invalid_key_index(lo_column_info,ls_table_name,ms_lang) RETURNING lb_result
    #建立表格的 Key 資料
    CALL sadzi140_db_set_constraint_data(ls_table_name) RETURNING lb_result
    #CALL sadzi140_db_set_key_data(ls_table_name,ls_key_list) RETURNING lb_result
    #CALL sadzi140_db_set_index_data(ls_table_name,ls_key_list) RETURNING lb_result

    #產出 table.sch 檔
    #CALL sadzi140_xml_get_lang_type_list() RETURNING lo_lang_arr
    CALL sadzi140_crud_get_static_lang_list() RETURNING lo_lang_arr
    LET lb_gen_tbl = TRUE #20170908
    LET ls_gen_list = "" #20170908
    FOR li_loop = 1 TO lo_lang_arr.getLength()
      LET ls_lang = lo_lang_arr[li_loop]
      CALL sadzi140_xml_gen_table_datas(lo_function_table_type_list.p_table_name,lo_function_table_type_list.p_module_name,ls_lang,ls_gen_list) RETURNING lb_gen_tbl,ls_gen_list #20170908
      CALL sadzi140_db_gen_table_schema(lo_function_table_type_list.p_table_name,lo_function_table_type_list.p_module_name,FALSE,ls_lang) RETURNING ls_schema_sql
    END FOR

    ##############################################################################
    CALL adzi140_check_db_data_diff(ls_master_user,ls_table_name) RETURNING lb_return_val
    IF lb_return_val THEN
      CALL sadzi140_util_make_alter_table(ls_master_user,ls_table_name,FALSE) RETURNING lb_error,ls_message
      IF NOT lb_error THEN
        CALL sadzi140_util_make_snapshot(ls_master_user,ls_table_name,ls_alm_version,ls_alm_request_no,ls_dgenv)
      END IF
      CALL sadzi140_db_update_alter_code(ls_master_user,ls_table_name)
      CALL sadzi140_db_gen_db_schema(ls_master_user.toLowerCase(),ls_table_name,lo_function_table_type_list.p_table_type) RETURNING lb_result
      IF NOT lb_result THEN
        CALL sadzp000_msg_show_error('adz-00381', 'adz-00381', '', 1)
      END IF
      IF lo_create_table.dct_table_type = "L" THEN
        #產生多語言 4gl 檔
        CALL sadzi140_util_gen_multi_lang(lo_create_table.dct_table_name) #[mi_dzea_index].dzea001)
      END IF
      #產生延伸檔
      CALL sadzi140_util_gen_extend_inc(ls_table_name) RETURNING lb_result
      #重新編譯失效物件
      CALL sadzi140_util_recompile_invalid_db_objects(ls_table_name) RETURNING lb_result
    END IF
    ##############################################################################
  END IF

  LET ls_return = ls_table_name

  RETURN ls_return

END FUNCTION

FUNCTION adzi140_get_digit_number()
DEFINE
  ls_sql      STRING,
  ls_digits   VARCHAR(10),
  ls_return   STRING

  LET ls_sql = "select ek.dzek007                             ",
               "  from dzek_t ek                              ",
               " where ek.dzek001 = '",cs_cdf_serial_number,"'"

  PREPARE lpre_get_digit_number FROM ls_sql
  DECLARE lcur_get_digit_number CURSOR FOR lpre_get_digit_number

  LET ls_digits = cs_digits

  OPEN lcur_get_digit_number
  FETCH lcur_get_digit_number INTO ls_digits
  CLOSE lcur_get_digit_number

  FREE lpre_get_digit_number
  FREE lcur_get_digit_number

  LET ls_return = NVL(ls_digits,cs_digits)

  RETURN ls_return

END FUNCTION

FUNCTION adzi140_get_column_define_type(p_table_name,p_column_name,p_topind) #20190304
DEFINE
  p_table_name  STRING,
  p_column_name STRING,
  p_topind      STRING  #20190304
DEFINE
  ls_table_name      STRING,
  ls_column_name     STRING,
  ls_topind          STRING,  #20190304
  ls_column_mean     STRING,
  ls_column_char     STRING,
  ls_table_char      STRING,
  ls_table_pre_name  STRING,
  lo_column_define   T_COLUMN_DEFINE,
  li_loop            INTEGER
DEFINE
  lo_return        T_COLUMN_DEFINE,
  ls_rtn_prename   STRING,
  ls_rtn_cust_code STRING #20190304

  LET ls_table_name  = p_table_name
  LET ls_column_name = p_column_name
  LET ls_topind      = p_topind #20190304

  LET ls_rtn_cust_code = ""  -- 20190304
  
  #先抓取表格Leading Code
  CALL sadzi140_util_get_table_leading_code(ls_table_name) RETURNING ls_table_pre_name

  #當表頭為標準時,則欄位名稱Leading Code要加上Alter Code
  IF m_dzea_t[mi_dzea_index].dzea018 = cs_dgenv_standard THEN
    #20190304 begin
    IF ls_topind = ms_topind THEN 
      LET ls_rtn_cust_code = ms_column_user_alter_code 
    ELSE
      LET ls_rtn_cust_code = IIF(ms_dgenv==cs_dgenv_customize,ms_column_user_alter_code,ms_column_industry_code) -- 20190304
    END IF  
    #20190304 end
    LET ls_table_pre_name = ls_table_pre_name,ls_rtn_cust_code -- 20190304
  ELSE
    LET ls_rtn_cust_code = ""  -- 20190304
    LET ls_table_pre_name = ls_table_pre_name
  END IF

  IF ls_column_name.getIndexOf(ls_table_pre_name,1) > 0 THEN
    LET ls_column_mean = ls_column_name.subString(ls_table_pre_name.getLength() + 1,ls_column_name.getLength())
  ELSE
    LET ls_column_mean = ls_column_name
  END IF

  #判斷輸入的欄位為字母或是數字
  LET lo_column_define.cd_DEFINE_TYPE = ""
  FOR li_loop = 1 TO ls_column_mean.getLength()
    LET ls_column_char = NVL(ls_column_mean.subString(li_loop,li_loop),"@")
    IF ls_column_char MATCHES "[a-z]" OR ls_column_char MATCHES "[A-Z]" THEN
      EXIT FOR
    ELSE
      LET lo_column_define.cd_DEFINE_TYPE = cs_cdf_serial_number
    END IF
  END FOR

  #若為非數字的欄位, 則抓定義
  IF lo_column_define.cd_DEFINE_TYPE IS NULL THEN
    CALL adzi140_get_column_define_data(ls_column_mean,ms_enable_user_define,m_dzea_t[mi_dzea_index].dzea014) RETURNING lo_column_define.* #20190304
  END IF

  LET lo_return.* = lo_column_define.*
  LET ls_rtn_prename = ls_table_pre_name

  RETURN lo_return.*,ls_rtn_prename,ls_rtn_cust_code #20190304

END FUNCTION

FUNCTION adzi140_get_column_define_data(p_column_mean,p_enable_user_define,p_topind) #20190304
DEFINE
  p_column_mean        STRING,
  p_enable_user_define STRING,
  p_topind             STRING #20190304
DEFINE
  ls_column_mean        STRING,
  ls_topind             STRING, #20190304
  ls_enable_user_define STRING,
  ls_sql                STRING,
  ls_sql_topind         STRING, #20190304
  lo_column_define      T_COLUMN_DEFINE
DEFINE
  lo_return  T_COLUMN_DEFINE

  LET ls_column_mean = p_column_mean
  LET ls_enable_user_define = p_enable_user_define
  LET ls_topind = p_topind #20190304

  #20190304 begin
  LET ls_sql_topind = ""
  IF ms_dgenv = cs_dgenv_standard THEN 
    IF ls_topind <> ms_topind THEN
      LET ls_sql_topind = " and ek.dzek001 = '",cs_cdf_serial_number,"' " 
    END IF
  END IF  
  #20190304 end
  
  LET ls_sql = "select ek.dzek001,ek.dzek002,ek.dzek004,       ",
               "       ekl.dzekl004 dzek005                    ",
               "  from dzek_t ek                               ",
               "  left outer join dzekl_t ekl                  ",
               "               on ekl.dzekl001 = ek.dzek001    ",
               "              and ekl.dzekl002 = ek.dzek002    ",
               "              and ekl.dzekl003 = '",ms_lang,"' ",
               " where ek.dzek002 = '",ls_column_mean,"'       ",
               ls_sql_topind

  PREPARE lpre_get_column_define FROM ls_sql
  DECLARE lcur_get_column_define CURSOR FOR lpre_get_column_define

  OPEN lcur_get_column_define
  FETCH lcur_get_column_define INTO lo_column_define.*
  CLOSE lcur_get_column_define

  FREE lpre_get_column_define
  FREE lcur_get_column_define

  LET lo_return.* = lo_column_define.*

  #若不能讓使用者自行定義, 則預設回傳 cs_no_define
  IF ls_enable_user_define = 'N' THEN
    LET lo_return.cd_DEFINE_TYPE = NVL(lo_column_define.cd_DEFINE_TYPE,cs_no_define)
  ELSE
    LET lo_return.cd_DEFINE_TYPE = NVL(lo_column_define.cd_DEFINE_TYPE,cs_cdf_user_define)
  END IF

  RETURN lo_return.*

END FUNCTION

FUNCTION adzi140_set_editor_mode(p_dialog,p_mode)
DEFINE
  p_dialog ui.dialog,
  p_mode   STRING
DEFINE
  lb_active   BOOLEAN,
  ls_mode     STRING,
  ls_mode_msg STRING,
  ls_db_type  STRING

  LET ms_edit_mode = p_mode
  LET ls_mode = p_mode

  CASE
    WHEN ls_mode = cs_editor_mode_browse
      CALL sadzp000_msg_get_message('adz-00206',ms_lang) RETURNING ls_mode_msg
    WHEN ls_mode = cs_editor_mode_edit
      CALL sadzp000_msg_get_message('adz-00207',ms_lang) RETURNING ls_mode_msg
    WHEN ls_mode = cs_editor_mode_design
      CALL sadzp000_msg_get_message('adz-00208',ms_lang) RETURNING ls_mode_msg
  END CASE

  DISPLAY "[ "||ls_mode_msg||" ]" TO formonly.lbl_editor_mode

  LET lb_active = IIF(ls_mode == cs_editor_mode_design,TRUE,FALSE)

  #因應增刪修模式切換功能按鈕
  CALL p_dialog.setActionActive("btn_addcdfconfirm", lb_active)
  CALL p_dialog.setActionActive("btn_cancelcdf", lb_active)
  CALL p_dialog.setActionActive("tbi_create_new_table", NOT lb_active)
  CALL p_dialog.setActionActive("tbi_create_function_table", NOT lb_active)
  CALL p_dialog.setActionActive("tbi_clone_table", NOT lb_active)
  CALL p_dialog.setActionActive("tbi_delete_table", NOT lb_active)
  CALL p_dialog.setActionActive("tbi_adzi150", NOT lb_active)
  CALL p_dialog.setActionActive("tbi_exec_alter", NOT lb_active)
  
  &ifdef is_msv
  &else
  &ifdef is_pgs
  &else
  LET ls_db_type = sadzi140_util_get_db_type()
  IF ls_db_type <> cs_db_driver_type_ora THEN
    CALL p_dialog.setActionActive("tbi_table_rebuild", FALSE)
    CALL p_dialog.setActionActive("tm_table_rebuild", FALSE) 
  ELSE 
    CALL p_dialog.setActionActive("tbi_table_rebuild", NOT lb_active)
    CALL p_dialog.setActionActive("tm_table_rebuild", NOT lb_active) #20170706
  END IF
  &endif
  &endif
  
  CALL p_dialog.setActionActive("tbi_rev_list", NOT lb_active)
  #CALL p_dialog.setActionActive("tm_rev_list", NOT lb_active)
  CALL p_dialog.setActionActive("modify", NOT lb_active)
  #CALL p_dialog.setActionActive("accept", NOT lb_active)
  #CALL p_dialog.setActionActive("cancel", NOT lb_active)
  #CALL p_dialog.setActionActive("delete", NOT lb_active)
  CALL p_dialog.setActionActive("exit", NOT lb_active)

  CALL p_dialog.setActionActive("btn_check_out", NOT lb_active)
  --CALL p_dialog.setActionActive("btn_recall", NOT lb_active)
  CALL p_dialog.setActionActive("btn_recall", FALSE)
  CALL p_dialog.setActionActive("btn_check_in", NOT lb_active)
  CALL p_dialog.setActionActive("btn_release", NOT lb_active)

END FUNCTION

FUNCTION adzi140_set_check_out_mode(p_dialog,p_dzea_t)
DEFINE
  p_dialog ui.dialog,
  p_dzea_t T_DZEA_T
DEFINE
  lo_dzea_t   T_DZEA_T,
  ls_check_out_full_name STRING,
  ls_mode_msg STRING

  LET lo_dzea_t.* = p_dzea_t.*

  CASE
    WHEN NVL(lo_dzea_t.dzlm008,cs_null_value) <> cs_check_out
      CALL sadzp000_msg_get_message('adz-00365',ms_lang) RETURNING ls_mode_msg
      LET ls_check_out_full_name = "[ ",ls_mode_msg," ]"
    WHEN NVL(lo_dzea_t.dzlm008,cs_null_value) = cs_check_out
      CALL sadzp000_msg_get_message('adz-00367',ms_lang) RETURNING ls_mode_msg
      LET ls_check_out_full_name = "[ ",ls_mode_msg,"：",lo_dzea_t.dzlm007,"-",lo_dzea_t.ooag011," ] [ ",lo_dzea_t.dzlm012,"#",lo_dzea_t.dzlm015," ]"
  END CASE

  DISPLAY ls_check_out_full_name TO formonly.lbl_check_out_full_name

END FUNCTION

FUNCTION adzi140_get_editor_mode()
DEFINE
  ls_return STRING

  LET ls_return = ms_edit_mode

  RETURN ls_return

END FUNCTION

FUNCTION adzi140_set_import_sco_enable(p_dialog)
DEFINE
  p_dialog ui.Dialog
DEFINE
  lo_window         ui.Window,
  lo_form           ui.Form

  LET lo_window = ui.Window.getCurrent()
  LET lo_form = lo_window.getForm()

  #若DZEB_T中有資料, 則不可以啟動匯入 SCO 功能
  #若DZEB_T中有資料, 才可以啟動建立附屬表格功能
  IF m_dzeb_t.getLength() >= 1 THEN
    CALL p_dialog.setActionActive("tm_import_sco", FALSE)
    CALL p_dialog.setActionActive("tbi_create_function_table", TRUE)
  ELSE
    CALL p_dialog.setActionActive("tm_import_sco", TRUE)
    CALL p_dialog.setActionActive("tbi_create_function_table", FALSE)
  END IF

  #是否啟動匯入 SCO 檔
  IF mb_enable_softscore OR mb_debug THEN
    CALL lo_form.setElementHidden("tm_import_sco",FALSE)
  ELSE
    CALL lo_form.setElementHidden("tm_import_sco",TRUE)
  END IF

END FUNCTION

FUNCTION adzi140_check_if_check_out_active(p_index)
DEFINE
  p_index INTEGER
DEFINE
  lb_check_out      BOOLEAN,
  lb_check_out_user BOOLEAN  
DEFINE
  lb_return BOOLEAN

  LET lb_check_out      = IIF(m_dzea_t[p_index].dzlm008 == cs_check_out,TRUE,FALSE)
  LET lb_check_out_user = IIF(m_dzea_t[p_index].dzlm007 == ms_user,TRUE,FALSE)
  
  LET lb_return =  lb_check_out AND lb_check_out_user
                  
  RETURN lb_return

END FUNCTION

FUNCTION adzi140_set_editable(p_dialog,p_index)
DEFINE
  p_dialog          ui.Dialog,
  p_index           INTEGER
DEFINE
  ls_table_pre_name STRING,
  lo_window         ui.Window,
  lo_form           ui.Form,
  ls_current_item   STRING #20160825
DEFINE
  lb_active BOOLEAN

  LET lo_window = ui.Window.getCurrent()
  LET lo_form = lo_window.getForm()
  LET ls_current_item = p_dialog.getCurrentItem() #20160825

  IF mb_debug THEN
    #CALL p_dialog.setActionActive("tm_batch_exec_alter", TRUE) #20170714 mark
    CALL p_dialog.setActionActive("tm_batch_table_rebuild", TRUE)
    #CALL p_dialog.setActionActive("tm_optimizer", TRUE)
    GOTO _RETURN
  ELSE
    #CALL p_dialog.setActionActive("tm_batch_exec_alter", FALSE) #20170714 mark
    CALL p_dialog.setActionActive("tm_batch_table_rebuild", FALSE)
    #CALL p_dialog.setActionActive("tm_optimizer", FALSE)
  END IF

  IF mb_enable_alm THEN
    #20171120 begin    
    IF (ms_DGENV <> cs_dgenv_standard) THEN  
      CALL lo_form.setElementHidden("btn_check_in",FALSE)
    ELSE
      CALL lo_form.setElementHidden("btn_check_in",TRUE)
    END IF  
    #20171120 end    
    IF (ms_user <> cs_topstd_user_name) AND (ms_DGENV = cs_dgenv_standard) THEN
      CALL lo_form.setElementHidden("btn_release",FALSE)
    ELSE   
      CALL lo_form.setElementHidden("btn_release",TRUE)
    END IF  
  ELSE
    CALL lo_form.setElementHidden("btn_check_in",FALSE)
    CALL lo_form.setElementHidden("btn_release",TRUE)
  END IF

  LET lb_active = adzi140_check_if_check_out_active(p_index)

  #從是否可讀寫狀態碼管制功能按鈕
  #CALL p_dialog.setActionActive("modify", lb_active)  #20160825 mark
  #20160825 begin
  IF ls_current_item = "sr_tableprivileges" AND ms_dgenv <> cs_dgenv_standard THEN 
    LET mb_modifiable = TRUE 
  ELSE
    LET mb_modifiable = lb_active
  END IF  
  CALL p_dialog.setActionActive("modify", mb_modifiable)
  #20160825 end
  CALL p_dialog.setActionActive("tbi_rev_list", lb_active)
  #CALL p_dialog.setActionActive("tm_rev_list", lb_active)
  &ifdef is_msv
  &else
  &ifdef is_pgs
  &else
  CALL p_dialog.setActionActive("tbi_table_rebuild", lb_active)
  CALL p_dialog.setActionActive("tm_table_rebuild", lb_active) #20170706
  &endif
  &endif
  CALL p_dialog.setActionActive("tbi_exec_alter", lb_active)
  CALL p_dialog.setActionActive("tbi_delete_table", lb_active)
  #CALL p_dialog.setActionActive("tbi_clone_table", lb_active)
  #CALL p_dialog.setActionActive("tbi_create_function_table", lb_active)
  CALL p_dialog.setActionActive("tm_import_sco", IIF(m_dzeb_t.getLength() >= 1,FALSE,lb_active))
  CALL p_dialog.setActionActive("tm_import", IIF(ms_user == cs_topstd_user_name,FALSE,TRUE))
  CALL p_dialog.setActionActive("tm_export", IIF(ms_user == cs_topstd_user_name,FALSE,TRUE)) --20170417
  CALL p_dialog.setActionActive("tm_clone_import", IIF(ms_user == cs_topstd_user_name,FALSE,TRUE)) --20170417
  CALL p_dialog.setActionActive("tm_selection_import", IIF(ms_user == cs_topstd_user_name,FALSE,TRUE)) --20171120
  CALL p_dialog.setActionActive("tm_del_std_column", IIF(ms_dgenv <> cs_dgenv_standard,TRUE,FALSE)) --20180302

  CALL p_dialog.setActionActive("btn_check_out", NOT lb_active)
  #CALL p_dialog.setActionActive("btn_recall", lb_active)
  CALL p_dialog.setActionActive("btn_recall", FALSE)
  CALL p_dialog.setActionActive("btn_check_in", lb_active)
  CALL p_dialog.setActionActive("btn_release", lb_active)

  #已經簽出,但簽出者不是自己,則所有簽核機能按鈕都Disable
  #或者在標準環境下,當表格所屬的行業別和環境行業別不同時,也不能簽出
  IF ((m_dzea_t[p_index].dzlm007 IS NOT NULL) AND (NVL(m_dzea_t[p_index].dzlm007,cs_null_value) <> ms_user)) THEN --OR #20190304 modify
     --((ms_dgenv = cs_dgenv_standard) AND (m_dzea_t[p_index].dzea014 <> ms_topind)) THEN #20190304 mark
    CALL p_dialog.setActionActive("btn_check_out", FALSE)
  END IF

  #ADZ,AZZ,LIB模組和安全機制的表格(gzy開頭)在客戶家不能簽出 -- 20160519
  LET ls_table_pre_name = m_dzea_t[p_index].dzea001
  LET ls_table_pre_name = ls_table_pre_name.subString(1,3)
  IF ((m_dzea_t[p_index].dzea003 = "ADZ") OR (m_dzea_t[p_index].dzea003 = "AZZ") OR (m_dzea_t[p_index].dzea003 = "LIB" AND mb_enable_checkout_lib) OR (ls_table_pre_name = "gzy")) AND
     (ms_dgenv = cs_dgenv_customize) THEN
    CALL p_dialog.setActionActive("btn_check_out", FALSE)
  END IF

  #只要SYS-0002設定為不能簽出,或user是topstd則一律管制簽出
  IF (NOT mb_enable_checkout) OR (ms_user = cs_topstd_user_name) THEN
    CALL p_dialog.setActionActive("btn_check_out", FALSE)
    CALL p_dialog.setActionActive("tbi_create_new_table", FALSE)
    CALL p_dialog.setActionActive("tbi_create_function_table", FALSE)
    CALL p_dialog.setActionActive("tbi_clone_table", FALSE)
    CALL p_dialog.setActionActive("tm_batch_checkout", FALSE) #20180314
  END IF

  #系統表格管制刪除
  IF ms_delete_table_control = "Y" THEN
    LET lb_active = IIF(m_dzea_t[p_index].dzea006 == "Y",FALSE,lb_active)
    CALL p_dialog.setActionActive("tbi_delete_table", lb_active)
  END IF

  #提速及多語言檔不開放表格複製
  {
  LET lb_active = IIF(m_dzea_t[p_index].dzea004 == "L" OR m_dzea_t[p_index].dzea004 == "V",FALSE,lb_active)
  CALL p_dialog.setActionActive("tbi_clone_table", lb_active)
  }

  #已出貨表格不可刪除
  LET lb_active = IIF((NVL(m_dzea_t[p_index].dzea017,cs_null_value) == "Y") OR (NVL(m_dzea_t[p_index].dzeastus,cs_null_value) == cs_pass_through_status_code),FALSE,TRUE)
  IF NOT lb_active THEN
    CALL p_dialog.setActionActive("tbi_delete_table", lb_active)
  END IF

  #客戶端不可以刪除標準表格
  IF ms_dgenv = cs_dgenv_customize THEN
    LET lb_active = IIF((NVL(m_dzea_t[p_index].dzea018,ms_dgenv) = ms_dgenv),TRUE,FALSE)
    CALL p_dialog.setActionActive("tbi_delete_table", lb_active)
    CALL p_dialog.setFieldActive("sr_tablelist.lbl_dzea004", lb_active)
  END IF

  LABEL _RETURN:

END FUNCTION

FUNCTION adzi140_check_column_data_exists()
DEFINE
  ls_replace_arg STRING,
  lb_return      BOOLEAN

  LET lb_return = TRUE

  IF m_dzeb_t.getLength() = 0 THEN
    LET ls_replace_arg = "|"
    CALL sadzp000_msg_show_error('adz-00150', 'adz-00150', ls_replace_arg, 1)
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF

  RETURN lb_return

END FUNCTION

FUNCTION adzi140_check_if_common_column_data_exists(p_table_name,p_table_type,p_lang,p_dzeb_t) #20170523
DEFINE
  p_table_name STRING,
  p_table_type STRING,
  p_lang       STRING,
  p_dzeb_t     DYNAMIC ARRAY OF T_DZEB_T #20170523
DEFINE
  ls_table_name  STRING,
  ls_table_type  STRING,
  ls_lang        STRING,
  lo_dzeb_t      DYNAMIC ARRAY OF T_DZEB_T, #20170523
  ls_question    STRING,
  ls_replace_arg STRING,
  li_loop        INTEGER, #20170523
  lj_loop        INTEGER, #20170523
  lb_data_exists BOOLEAN, #20170523
  lb_return      BOOLEAN,
  lo_SRC_COMMON_COLUMN_DATAS  DYNAMIC ARRAY OF T_COMMON_COLUMN_DATAS,
  lo_DST_COMMON_COLUMN_DATAS  DYNAMIC ARRAY OF T_COMMON_COLUMN_DATAS

  LET ls_table_name = p_table_name
  LET ls_table_type = p_table_type
  LET ls_lang       = p_lang
  LET lo_dzeb_t     = p_dzeb_t #20170523

  LET lb_return = TRUE

  CALL adzi140_get_common_column_list(ls_table_type,ls_lang) RETURNING lo_SRC_COMMON_COLUMN_DATAS
  CALL adzi140_get_table_common_column_list(ls_table_name,ls_table_type,ls_lang) RETURNING lo_DST_COMMON_COLUMN_DATAS

  {
  #20170523 mark begin
  #檢核共用欄位是否已經被設定
  IF lo_SRC_COMMON_COLUMN_DATAS.getLength() > 0 AND lo_DST_COMMON_COLUMN_DATAS.getLength() = 0 THEN
    LET ls_replace_arg = "|"
    CALL sadzp000_msg_question_box("adz-00590", "adz-00590", ls_replace_arg, 0) RETURNING ls_question
    LET lb_return = IIF(ls_question == cs_question_yes,TRUE,FALSE)
  END IF
  #20170523 mark end
  }

  #20170523 begin
  IF lo_DST_COMMON_COLUMN_DATAS.getLength() = 0 THEN
    LET lb_data_exists = FALSE
    FOR li_loop = 1 TO lo_SRC_COMMON_COLUMN_DATAS.getLength()
      FOR lj_loop = 1 TO lo_dzeb_t.getLength()
        IF lo_dzeb_t[lj_loop].dzeb022 = lo_SRC_COMMON_COLUMN_DATAS[li_loop].p_column_type THEN
          LET lb_data_exists = TRUE
          EXIT FOR
        END IF
      END FOR 
      IF lb_data_exists THEN EXIT FOR END IF
    END FOR 
  END IF
  
  IF lo_SRC_COMMON_COLUMN_DATAS.getLength() > 0 AND lo_DST_COMMON_COLUMN_DATAS.getLength() = 0 AND (NOT lb_data_exists) THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  #20170523 end

  RETURN lb_return

END FUNCTION

FUNCTION adzi140_get_common_column_list(p_table_type,p_lang)
DEFINE
  p_table_type STRING,
  p_lang       STRING
DEFINE
  ls_table_type  STRING,
  ls_lang        STRING,
  ls_sql         STRING,
  ls_recursive   STRING, #20200330
  li_cnt         INTEGER,
  lo_common_column_datas  DYNAMIC ARRAY OF T_COMMON_COLUMN_DATAS
DEFINE
  lo_Return DYNAMIC ARRAY OF T_COMMON_COLUMN_DATAS

  LET ls_table_type = p_table_type
  LET ls_lang = p_lang

  {
  LET ls_sql = "select ekx.dzek001,ekx.dzek002,ekl.dzekl004                   ",
               "  from (                                                      ",
               "          select dzek001,dzek002                              ",
               "            from dzek_t                                       ",
               "           start with dzek001 = 'grp",ls_table_type,"'        ",
               "         connect by prior dzek002 = dzek001                   ",
               "           order by dzek001                                   ",
               "       ) ekx                                                  ",
               "  left outer join dzekl_t ekl on ekl.dzekl001 = ekx.dzek001   ",
               "                             and ekl.dzekl002 = ekx.dzek002   ",
               "                             and ekl.dzekl003 = '",ls_lang,"' ",
               " where ekx.dzek001 not like 'grp%'                            ",
               " order by ekx.dzek001                                         "
  }

  LET ls_recursive = sadzi140_util_get_cte_recursive_sign() #20200330
  
  LET ls_sql = "with ",ls_recursive," ekr(dzek001,dzek002) as        ", #20200330
               "        (                                            ",
               "          select ek.dzek001,ek.dzek002               ",
               "            from dzek_t ek                           ",
               "           where ek.dzek001 = 'grp",ls_table_type,"' ",
               "           union all                                 ",
               "          select ekx.dzek001,ekx.dzek002             ",                        
               "            from dzek_t ekx, ekr                     ",
               "           where ekx.dzek001 = ekr.dzek002           ",
               "        )                                            ",
               "select ekr.dzek001,ekr.dzek002,ekl.dzekl004          ",
               "  from ekr                                           ",
               "  left outer join dzekl_t ekl                        ", 
               "               on ekl.dzekl001 = ekr.dzek001         ", 
               "              and ekl.dzekl002 = ekr.dzek002         ", 
               "              and ekl.dzekl003 = '",ls_lang,"'       ",
               " where ekr.dzek001 not like 'grp%'                   ",                
               " order by ekr.dzek001                                "

  PREPARE lpre_get_common_column_list FROM ls_sql
  DECLARE lcur_get_common_column_list CURSOR FOR lpre_get_common_column_list

  LET li_cnt = 1
  OPEN lcur_get_common_column_list
  FOREACH lcur_get_common_column_list INTO lo_common_column_datas[li_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_cnt = li_cnt + 1

  END FOREACH
  CLOSE lcur_get_common_column_list

  FREE lpre_get_common_column_list
  FREE lcur_get_common_column_list

  CALL lo_common_column_datas.deleteElement(li_cnt)

  LET lo_return = lo_common_column_datas

  RETURN lo_return

END FUNCTION

FUNCTION adzi140_get_table_common_column_list(p_table_name,p_table_type,p_lang)
DEFINE
  p_table_name STRING,
  p_table_type STRING,
  p_lang       STRING
DEFINE
  ls_table_name  STRING,
  ls_table_type  STRING,
  ls_lang        STRING,
  ls_sql         STRING,
  ls_recursive   STRING, #20200330
  li_cnt         INTEGER,
  lo_common_column_datas  DYNAMIC ARRAY OF T_COMMON_COLUMN_DATAS
DEFINE
  lo_Return DYNAMIC ARRAY OF T_COMMON_COLUMN_DATAS

  LET ls_table_name = p_table_name
  LET ls_table_type = p_table_type
  LET ls_lang = p_lang

  {
  LET ls_sql = "select eb.dzeb022,ek.dzek002,ekl.dzekl004                               ",
               "  from dzeb_t eb                                                        ",
               "  left outer join dzek_t ek on ek.dzek001 = eb.dzeb022                  ",
               "  left outer join dzekl_t ekl on ekl.dzekl001 = eb.dzeb022              ",
               "                             and ekl.dzekl002 = ek.dzek002              ",
               "                             and ekl.dzekl003 = '",ls_lang,"'           ",
               " where eb.dzeb001 = '",ls_table_name,"'                                 ",
               "   and exists (                                                         ",
               "                select 1                                                ",
               "                  from (                                                ",
               "                          select dzek001,dzek002                        ",
               "                            from dzek_t                                 ",
               "                           start with dzek001 = 'grp",ls_table_type,"'  ",
               "                         connect by prior dzek002 = dzek001             ",
               "                           order by dzek001                             ",
               "                       ) ekx                                            ",
               "                 where ekx.dzek001 not like 'grp%'                      ",
               "                   and ekx.dzek001 = eb.dzeb022                         ",
               "              )                                                         ",
               " order by eb.dzeb022                                                    "
  }

  LET ls_recursive = sadzi140_util_get_cte_recursive_sign() #20200330
  
  LET ls_sql = "with ",ls_recursive," ekr(dzek001,dzek002) as                 ", #20200330
               "        (                                                     ",
               "          select ek.dzek001,ek.dzek002                        ",
               "            from dzek_t ek                                    ",
               "           where ek.dzek001 = 'grp",ls_table_type,"'          ",
               "           union all                                          ",
               "          select ekx.dzek001,ekx.dzek002                      ",                        
               "            from dzek_t ekx, ekr                              ",
               "           where ekx.dzek001 = ekr.dzek002                    ",
               "        )                                                     ",
               "select eb.dzeb022,ek.dzek002,ekl.dzekl004                     ",
               "  from dzeb_t eb                                              ",
               "  left outer join dzek_t ek on ek.dzek001 = eb.dzeb022        ",
               "  left outer join dzekl_t ekl on ekl.dzekl001 = eb.dzeb022    ",
               "                             and ekl.dzekl002 = ek.dzek002    ",
               "                             and ekl.dzekl003 = '",ls_lang,"' ",
               " where eb.dzeb001 = '",ls_table_name,"'                       ",
               "   and exists (                                               ",
               "                select 1                                      ",
               "                  from ekr                                    ",
               "                 where ekr.dzek001 not like 'grp%'            ",
               "                   and ekr.dzek001 = eb.dzeb022               ",
               "              )                                               ",
               " order by eb.dzeb022                                          "
  
  PREPARE lpre_get_table_common_column_list FROM ls_sql
  DECLARE lcur_get_table_common_column_list CURSOR FOR lpre_get_table_common_column_list

  LET li_cnt = 1
  OPEN lcur_get_table_common_column_list
  FOREACH lcur_get_table_common_column_list INTO lo_common_column_datas[li_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_cnt = li_cnt + 1

  END FOREACH
  CLOSE lcur_get_table_common_column_list

  FREE lpre_get_table_common_column_list
  FREE lcur_get_table_common_column_list

  CALL lo_common_column_datas.deleteElement(li_cnt)

  LET lo_return = lo_common_column_datas

  RETURN lo_return

END FUNCTION

FUNCTION adzi140_check_if_user_define_column_data_exists(p_table_name,p_table_type,p_dzeb_t) #20170523
DEFINE
  p_table_name STRING,
  p_table_type STRING,
  p_dzeb_t     DYNAMIC ARRAY OF T_DZEB_T #20170523
DEFINE
  ls_table_name  STRING,
  ls_table_type  STRING,
  lo_dzeb_t      DYNAMIC ARRAY OF T_DZEB_T, #20170523
  ls_question    STRING,
  ls_replace_arg STRING,
  li_loop        INTEGER, #20170523
  lb_data_exists BOOLEAN, #20170523
  ls_dzeb022     STRING, #20170523
  lb_result      BOOLEAN,
  lb_udc_exist   BOOLEAN,
  lb_return      BOOLEAN

  LET ls_table_name = p_table_name
  LET ls_table_type = p_table_type
  LET lo_dzeb_t     = p_dzeb_t #20170523

  LET lb_return = TRUE
  LET lb_result = FALSE
  LET lb_udc_exist = FALSE
  LET lb_data_exists = FALSE #20170523

  CALL adzi140_check_if_need_user_define_column(ls_table_type) RETURNING lb_result

  #檢核共用欄位是否已經被設定
  IF lb_result THEN
    CALL adzi140_check_user_define_column_if_exists(ls_table_name) RETURNING lb_udc_exist
    #20170523 begin
    IF NOT lb_udc_exist THEN
      FOR li_loop = 1 TO lo_dzeb_t.getLength()
        LET ls_dzeb022 = lo_dzeb_t[li_loop].dzeb022 
        IF ls_dzeb022.subString(1,cs_cdf_user_define.getLength()) = cs_cdf_user_define THEN
          LET lb_data_exists = TRUE
          EXIT FOR
        END IF
      END FOR
    END IF
    IF NOT lb_udc_exist AND NOT lb_data_exists THEN
      LET lb_return = FALSE 
    ELSE
      LET lb_return = TRUE 
    END IF
    #20170523 end
    
    {
    #20170523 mark begin
    IF NOT lb_udc_exist THEN
      LET ls_replace_arg = "|"
      CALL sadzp000_msg_question_box("adz-00591", "adz-00591", ls_replace_arg, 0) RETURNING ls_question
      LET lb_return = IIF(ls_question == cs_question_yes,TRUE,FALSE)
      LET lb_return = TRUE #20170523
    ELSE #20170523   
      LET lb_return = FALSE #20170523
    END IF #20170523
    #20170523 
    }
    
  END IF

  RETURN lb_return

END FUNCTION

FUNCTION adzi140_get_table_column_list(p_table_name,p_columns,p_side)
DEFINE
  p_table_name STRING,
  p_columns    STRING,
  p_side       STRING
DEFINE
  ls_table_name    STRING,
  ls_columns       STRING,
  ls_side          STRING,
  lo_column_array  DYNAMIC ARRAY OF T_COLUMN_LIST,
  li_rec_cnt       INTEGER,
  ls_sql           STRING,
  ls_sql_where     STRING
DEFINE
  lo_return  DYNAMIC ARRAY OF T_COLUMN_LIST

  LET ls_table_name = p_table_name
  LET ls_columns    = p_columns
  LET ls_side       = p_side

  IF ls_side = cs_side_left THEN
    IF ls_columns <> "''" THEN
      LET ls_sql_where = " and eb.dzeb002 not in (",ls_columns,")"
    ELSE
      LET ls_sql_where = " and 1=1 "
    END IF
  ELSE
    IF ls_side = cs_side_right THEN
      IF ls_columns <> "''" THEN
        LET ls_sql_where = " and eb.dzeb002 in (",ls_columns,")"
      ELSE
        LET ls_sql_where = " and 1=2 "
      END IF
    END IF
  END IF

  LET ls_sql = "select eb.dzeb002,  ebl.dzebl003               ",
               "  from dzeb_t eb                               ",
               "  left outer join dzebl_t ebl                  ",
               "               on ebl.dzebl001 = eb.dzeb002    ",
               "              and ebl.dzebl002 = '",ms_lang,"' ",
               " where eb.dzeb001 = '",ls_table_name,"'        ",
               ls_sql_where,
               " order by eb.dzeb021,eb.dzeb002                "

  PREPARE lpre_get_table_column_list FROM ls_sql
  DECLARE lcur_get_table_column_list CURSOR FOR lpre_get_table_column_list

  LET li_rec_cnt = 1

  CALL lo_column_array.Clear()

  OPEN lcur_get_table_column_list
  FOREACH lcur_get_table_column_list INTO lo_column_array[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CALL lo_column_array.deleteElement(li_rec_cnt)

  CLOSE lcur_get_table_column_list

  FREE lpre_get_table_column_list
  FREE lcur_get_table_column_list

  LET lo_return = lo_column_array

  RETURN lo_return

END FUNCTION

#20180302 begin
FUNCTION adzi140_get_standard_column_list(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_table_name    STRING,
  lo_column_array  DYNAMIC ARRAY OF T_COLUMN_LIST,
  li_rec_cnt       INTEGER,
  ls_sql           STRING,
  ls_sql_where     STRING
DEFINE
  lo_return  DYNAMIC ARRAY OF T_COLUMN_LIST

  LET ls_table_name = p_table_name

  LET ls_sql = "select eb.dzeb002,  ebl.dzebl003               ",
               "  from dzeb_t eb                               ",
               "  left outer join dzebl_t ebl                  ",
               "               on ebl.dzebl001 = eb.dzeb002    ",
               "              and ebl.dzebl002 = '",ms_lang,"' ",
               " where eb.dzeb001 =  '",ls_table_name,"'       ",
               "   and eb.dzeb030 =  '",cs_dgenv_standard,"'   ", #標準欄位
               "   and eb.dzeb029 <> '",cs_dgenv_customize,"'  ", #沒有客制過
               "   and eb.dzeb004 <> 'Y'                       ", #非 key 欄位 
               " order by eb.dzeb021,eb.dzeb002                "

  PREPARE lpre_get_standard_column_list FROM ls_sql
  DECLARE lcur_get_standard_column_list CURSOR FOR lpre_get_standard_column_list

  LET li_rec_cnt = 1

  CALL lo_column_array.Clear()

  OPEN lcur_get_standard_column_list
  FOREACH lcur_get_standard_column_list INTO lo_column_array[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CALL lo_column_array.deleteElement(li_rec_cnt)

  CLOSE lcur_get_standard_column_list

  FREE lpre_get_standard_column_list
  FREE lcur_get_standard_column_list

  LET lo_return = lo_column_array

  RETURN lo_return

END FUNCTION
#20180302 end

#20160614 begin  
FUNCTION adzi140_get_input_column_list(p_columns)
DEFINE
  p_columns STRING
DEFINE
  li_index          INTEGER,
  ls_column_char    STRING,
  ls_columns        STRING,
  ls_columns_string STRING,
  lo_column_array   DYNAMIC ARRAY OF T_COLUMN_LIST
DEFINE
  lo_return DYNAMIC ARRAY OF T_COLUMN_LIST

  LET ls_columns = p_columns

  LET ls_columns_string = ""

  FOR li_index = 1 TO ls_columns.getLength()
    LET ls_column_char = NVL(ls_columns.subString(li_index,li_index),cs_null_value)
    IF ls_column_char MATCHES "," THEN
      IF ls_columns_string.trim() IS NOT NULL THEN 
        LET lo_column_array[lo_column_array.getLength() + 1].COLUMN_NAME = ls_columns_string
        LET ls_columns_string = ""
      END IF  
    ELSE
      LET ls_columns_string = ls_columns_string,ls_column_char
    END IF
  END FOR

  #最後一組或僅有一組要加入Queue
  IF ls_columns_string.trim() IS NOT NULL THEN 
    LET lo_column_array[lo_column_array.getLength() + 1].COLUMN_NAME = ls_columns_string
  END IF
  
  LET lo_return = lo_column_array
  
  RETURN lo_return

END FUNCTION
#20160614 end

FUNCTION adzi140_decode_columns(p_columns)
DEFINE
  p_columns STRING
DEFINE
  li_index          INTEGER,
  ls_column_char    STRING,
  ls_columns        STRING,
  ls_columns_string STRING,
  ls_columns_queue  STRING
DEFINE
  ls_return STRING

  LET ls_columns = p_columns

  LET ls_columns_queue  = ""
  LET ls_columns_string = "'"

  FOR li_index = 1 TO ls_columns.getLength()
    LET ls_column_char = NVL(ls_columns.subString(li_index,li_index),cs_null_value)
    IF ls_column_char MATCHES "," THEN
      LET ls_columns_string = ls_columns_string,"'"
      LET ls_columns_queue  = ls_columns_queue,ls_columns_string,","
      LET ls_columns_string = "'"
    ELSE
      LET ls_columns_string = ls_columns_string,ls_column_char
    END IF
  END FOR

  #最後一組或僅有一組要加入Queue
  LET ls_columns_string = ls_columns_string,"'"
  LET ls_columns_queue  = ls_columns_queue,ls_columns_string,","
  LET ls_columns_string = "'"

  LET ls_return = ls_columns_queue.subString(1,ls_columns_queue.getLength()-1)

  RETURN ls_return

END FUNCTION

#檢核Key及Index名稱是否存在
FUNCTION adzi140_check_key_index_name_exist(p_name)
DEFINE
  p_name  STRING
DEFINE
  ls_name   STRING,
  lb_result BOOLEAN
DEFINE
  lb_return BOOLEAN

  LET ls_name = p_name

  CALL adzi140_check_key_name_exist(ls_name) RETURNING lb_result
  IF lb_result THEN GOTO _return END IF
  CALL adzi140_check_index_name_exist(ls_name) RETURNING lb_result
  IF lb_result THEN GOTO _return END IF

  LABEL _return:

  LET lb_return = lb_result

  RETURN lb_return

END FUNCTION

#檢核Key名稱是否存在
FUNCTION adzi140_check_key_name_exist(p_name)
DEFINE
  p_name  STRING
DEFINE
  ls_name       STRING,
  ls_sql        STRING,
  li_rec_count  INTEGER,
  lb_return     BOOLEAN

  LET ls_name = p_name

  #取得資料筆數
  LET ls_sql = "select count(1)                   ",
               "  from dzed_t ed                  ",
               " where ed.dzed002 = '",ls_name,"' "

  PREPARE lpre_check_key_name_exist FROM ls_sql
  DECLARE lcur_check_key_name_exist CURSOR FOR lpre_check_key_name_exist
  OPEN lcur_check_key_name_exist
  FETCH lcur_check_key_name_exist INTO li_rec_count
  CLOSE lcur_check_key_name_exist
  FREE lcur_check_key_name_exist
  FREE lpre_check_key_name_exist

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF

  RETURN lb_return

END FUNCTION

#檢核Index名稱是否存在
FUNCTION adzi140_check_index_name_exist(p_name)
DEFINE
  p_name  STRING
DEFINE
  ls_name       STRING,
  ls_sql        STRING,
  li_rec_count  INTEGER,
  lb_return     BOOLEAN

  LET ls_name = p_name

  #取得資料筆數
  LET ls_sql = "select count(1)                   ",
               "  from dzec_t ec                  ",
               " where ec.dzec002 = '",ls_name,"' "

  PREPARE lpre_check_index_name_exist FROM ls_sql
  DECLARE lcur_check_index_name_exist CURSOR FOR lpre_check_index_name_exist
  OPEN lcur_check_index_name_exist
  FETCH lcur_check_index_name_exist INTO li_rec_count
  CLOSE lcur_check_index_name_exist
  FREE lcur_check_index_name_exist
  FREE lpre_check_index_name_exist

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF

  RETURN lb_return

END FUNCTION

FUNCTION adzi140_check_privilege_name_exist(p_rec_pos,p_t_dzen_t)
DEFINE
  p_rec_pos   INTEGER,
  p_t_dzen_t  DYNAMIC ARRAY OF T_DZEN_T
DEFINE
  li_rec_pos     INTEGER,
  lo_t_dzen_t    DYNAMIC ARRAY OF T_DZEN_T,
  li_loop        INTEGER,
  ls_key_data    STRING,
  ls_loop_data   STRING,
  lb_exists      BOOLEAN
DEFINE
  lb_return  BOOLEAN

  LET li_rec_pos  = p_rec_pos
  LET lo_t_dzen_t = p_t_dzen_t

  LET lb_exists = FALSE

  #組合Key資料
  LET ls_key_data = lo_t_dzen_t[li_rec_pos].dzen002,lo_t_dzen_t[li_rec_pos].dzen003

  FOR li_loop = 1 TO lo_t_dzen_t.getLength()
    IF li_loop <> li_rec_pos THEN
      LET ls_loop_data = lo_t_dzen_t[li_loop].dzen002,lo_t_dzen_t[li_loop].dzen003
      IF ls_loop_data.trim() = ls_key_data.trim() THEN
        LET lb_exists = TRUE
        EXIT FOR
      END IF
    END IF
  END FOR

  LET lb_return = lb_exists

  RETURN lb_return

END FUNCTION

FUNCTION adzi140_check_if_system_table(p_module_name)
DEFINE
  p_module_name STRING
DEFINE
  ls_module_name STRING,
  ls_is_system_table STRING
DEFINE
  ls_return STRING

  LET ls_module_name = p_module_name

  IF (ls_module_name = "ADZ") OR (ls_module_name = "AZZ") THEN
    LET ls_is_system_table = "Y"
  ELSE
    LET ls_is_system_table = "N"
  END IF

  LET ls_return = ls_is_system_table

  RETURN ls_return

END FUNCTION

FUNCTION adzi140_get_fitting_table_type(p_module_name,p_orig_table_type)
DEFINE
  p_module_name STRING,
  p_orig_table_type  STRING
DEFINE
  ls_module_name STRING,
  ls_orig_table_type  STRING,
  ls_table_type  STRING
DEFINE
  ls_return STRING

  LET ls_module_name     = p_module_name
  LET ls_orig_table_type = p_orig_table_type

  IF (ls_module_name = "ADZ") THEN
    LET ls_table_type = "X"
  ELSE
    LET ls_table_type = ls_orig_table_type
  END IF

  LET ls_return = ls_table_type

  RETURN ls_return

END FUNCTION

FUNCTION adzi140_check_if_multi_lang_table(p_table_type)
DEFINE
  p_table_type STRING
DEFINE
  ls_table_type    STRING,
  ls_is_multi_lang STRING
DEFINE
  ls_return STRING

  LET ls_table_type = p_table_type

  IF ls_table_type = "L" THEN
    LET ls_is_multi_lang = 'Y'
  ELSE
    LET ls_is_multi_lang = 'N'
  END IF

  LET ls_return = ls_is_multi_lang

  RETURN ls_return

END FUNCTION

FUNCTION adzi140_alarm_if_multi_lang_table(p_if_multi_lang)
DEFINE
  p_if_multi_lang STRING
DEFINE
  lb_if_multi_lang BOOLEAN

  LET lb_if_multi_lang = IIF(p_if_multi_lang=="Y",TRUE,FALSE)

  IF lb_if_multi_lang THEN 
    CALL sadzp000_msg_show_info('adz-00823', 'adz-00823', "|", 1)
  END IF

END FUNCTION

FUNCTION adzi140_check_if_tool_table(p_table_type)
DEFINE
  p_table_type STRING
DEFINE
  ls_table_type STRING,
  lb_is_tool    BOOLEAN
DEFINE
  lb_return BOOLEAN

  LET ls_table_type = p_table_type

  IF ls_table_type = "X" THEN
    LET lb_is_tool = TRUE
  ELSE
    LET lb_is_tool = FALSE
  END IF

  LET lb_return = lb_is_tool

  RETURN lb_return

END FUNCTION

FUNCTION adzi140_get_DZAC_prog_list(p_column_name)
DEFINE
  p_column_name STRING
DEFINE
  ls_line          STRING,
  ls_sql           STRING,
  ls_column_name   STRING,
  ls_prog_list     STRING,
  ls_temp          STRING,
  li_count         INTEGER,
  ls_prog_name     VARCHAR(250)
DEFINE
  ls_Return STRING

  LET ls_column_name  = p_column_name
  LET ls_prog_list = NULL

  LET ls_prog_name = ""
  LET li_count = 0

  LET ls_sql = "select distinct ac.dzac001               ",
               "  from dzac_t ac                         ",
               " where ac.dzac002 = '",ls_column_name,"' ",
               " order by ac.dzac001                     "

  PREPARE lpre_get_DZAC_prog_list FROM ls_sql
  DECLARE lcur_get_DZAC_prog_list CURSOR FOR lpre_get_DZAC_prog_list

  OPEN lcur_get_DZAC_prog_list
  FOREACH lcur_get_DZAC_prog_list INTO ls_prog_name
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET ls_line = ls_prog_name
    LET ls_prog_list = ls_prog_list,ls_line,","

    LET li_count = li_count + 1
    IF li_count >= 10 THEN
      LET li_count = 0
      LET ls_prog_list = ls_prog_list,"\n"
    END IF

  END FOREACH
  CLOSE lcur_get_DZAC_prog_list

  FREE lpre_get_DZAC_prog_list
  FREE lcur_get_DZAC_prog_list

  LET ls_temp = "\n",ls_prog_list,"\n"

  IF ls_prog_list IS NOT NULL THEN
    LET ls_Return = ls_temp
  ELSE
    LET ls_Return = ""
  END IF

  RETURN ls_Return

END FUNCTION

#呼叫 ALM 取得 DZLU 資訊
FUNCTION adzi140_step_call_alm_check_out(p_enable_alm,p_user,p_lang,p_role,p_step,p_wizard)
DEFINE
  p_enable_alm BOOLEAN,
  p_user       STRING,
  p_lang       STRING,
  p_role       STRING,
  p_step       T_STEP,
  p_wizard     BOOLEAN
DEFINE
  lb_enable_alm BOOLEAN,
  ls_user       STRING,
  ls_lang       STRING,
  ls_role       STRING,
  lo_step       T_STEP,
  li_step       INTEGER,
  lb_wizard     BOOLEAN,
  lo_DZLU_T     DYNAMIC ARRAY OF T_DZLU_T,
  lo_user_info  T_USER_INFO,
  lb_result     BOOLEAN

  LET lb_enable_alm = p_enable_alm
  LET ls_user       = p_user
  LET ls_lang       = p_lang
  LET ls_role       = p_role
  LET lo_step.*     = p_step.*
  LET lb_wizard     = p_wizard

  #設定使用者資訊
  LET lo_user_info.ui_NUMBER = ls_user
  &ifndef DEBUG
    LET lo_user_info.ui_NAME = cl_get_username(ls_user)
  &else
    LET lo_user_info.ui_NAME = cs_null_username_default
  &endif
  LET lo_user_info.ui_LANG = ls_lang
  LET lo_user_info.ui_ROLE = ls_role

  IF lo_step.NEXT THEN
    IF (lb_enable_alm) THEN
      #取得DZLU資料
      CALL sadzp200_ckout_run(lo_USER_INFO.ui_ROLE,lo_USER_INFO.ui_NUMBER,lo_USER_INFO.ui_LANG,0,lb_wizard) RETURNING lb_result,li_step,lo_DZLU_T
    ELSE
      CALL sadzp200_ckout_get_dzlu_without_alm(lo_USER_INFO.ui_ROLE,lo_USER_INFO.ui_NUMBER) RETURNING lb_result,lo_DZLU_T
    END IF
  END IF

  RETURN lb_result,lo_step.*,lo_user_info.*,lo_DZLU_T

END FUNCTION

FUNCTION adzi140_step_get_alm_info(p_table_name,p_module_name,p_table_desc,p_spec_type,p_role,p_DZLU_T)
DEFINE
  p_table_name  STRING,
  p_module_name STRING,
  p_table_desc  STRING,
  p_spec_type   STRING,
  p_role        STRING,
  p_DZLU_T      DYNAMIC ARRAY OF T_DZLU_T
DEFINE
  ls_table_name  STRING,
  ls_module_name STRING,
  ls_table_desc  STRING,
  ls_spec_type   STRING,
  ls_role        STRING,
  lo_DZLU_T      DYNAMIC ARRAY OF T_DZLU_T,
  lb_result      BOOLEAN,
  ls_update_type STRING,
  li_dzlu        INTEGER,
  ls_GUID        STRING,
  lo_dzlm_t      T_DZLM_T,
  lo_DZAF_T      T_DZAF_T,
  lo_table_info  T_PROGRAM_INFO
DEFINE
  lb_return  BOOLEAN,
  lo_return  T_DZLM_T


  LET ls_table_name  = p_table_name
  LET ls_module_name = p_module_name
  LET ls_table_desc  = p_table_desc
  LET ls_spec_type   = p_spec_type
  LET ls_role        = p_role
  LET lo_DZLU_T      = p_DZLU_T

  #設定表格資訊
  LET lo_table_info.pi_NAME   = ls_table_name
  LET lo_table_info.pi_MODULE = ls_module_name
  LET lo_table_info.pi_DESC   = ls_table_desc
  LET lo_table_info.pi_TYPE   = ls_spec_type

  #設定版號輸入資料
  LET lo_DZAF_T.DZAF001 = lo_table_info.pi_NAME
  LET lo_DZAF_T.DZAF005 = lo_table_info.pi_TYPE
  LET lo_DZAF_T.DZAF006 = lo_table_info.pi_MODULE

  #取得新版號
  BEGIN WORK -- 20200330 
  
  CALL sadzp200_ver_get_new_ver_info(lo_DZAF_T.*,ls_role,FALSE) RETURNING lb_result,lo_DZAF_T.*
  IF NOT lb_result THEN DISPLAY cs_error_tag,"Get New version information fault." GOTO _RETURN END IF
  
  #彙整入DZLM_T
  CALL security.RandomGenerator.CreateUUIDString() RETURNING ls_GUID
  FOR li_dzlu = 1 TO lo_DZLU_T.getLength()
    IF lo_DZLU_T[li_dzlu].DZLU001 IS NOT NULL THEN
      CALL sadzp200_alm_set_dzlm_mix_info(lo_DZLU_T[li_dzlu].*,lo_DZAF_T.*,lo_table_info.*,ls_GUID) RETURNING lb_result,ls_update_type
      EXIT FOR
    END IF
  END FOR
  IF NOT lb_result THEN DISPLAY cs_error_tag,"Set DZLM data fault." GOTO _RETURN END IF
  CALL sadzp200_alm_get_dzlm(lo_table_info.*,cs_user_role_sd) RETURNING lo_dzlm_t.*
  IF NOT lb_result THEN DISPLAY cs_error_tag,"Get DZLM information fault." GOTO _RETURN END IF

  LABEL _RETURN:

  -- 20200330 begin
  IF NOT lb_result THEN 
    ROLLBACK WORK 
  ELSE 
    COMMIT WORK
  END IF
  -- 20200330 end

  LET lb_return = lb_result
  LET lo_return.* = lo_dzlm_t.*

  RETURN lb_return,lo_return.*

END FUNCTION

FUNCTION adzi140_grant_all_db_APS_privilege()
DEFINE
  lo_db_connstr DYNAMIC ARRAY OF T_DB_CONNSTR,
  li_loop       INTEGER,
  li_db_counts  INTEGER

  CALL sadzi140_db_get_all_db_connect_string() RETURNING lo_db_connstr

  FOR li_loop = 1 TO lo_db_connstr.getLength()
    CALL sadzi140_db_grant_APS_privilege(lo_db_connstr[li_loop].*,ms_master_db)
  END FOR

END FUNCTION

FUNCTION adzi140_check_if_exist_aps_grant(p_dzen_t)
DEFINE 
  p_dzen_t DYNAMIC ARRAY OF T_DZEN_T
DEFINE
  lo_dzen_t DYNAMIC ARRAY OF T_DZEN_T,
  li_loop   INTEGER,
  lb_result BOOLEAN
DEFINE
  lb_return BOOLEAN
  
  LET lo_dzen_t = p_dzen_t

  LET lb_result = FALSE

  FOR li_loop = 1 TO lo_dzen_t.getLength()
    IF lo_dzen_t[li_loop].dzen003 = cs_privilege_receiver_dsaps THEN
      LET lb_result = TRUE
    END IF
  END FOR  

  LET lb_return = lb_result

  RETURN lb_return

END FUNCTION

FUNCTION adzi140_set_step_data(p_prev,p_next,p_ok,p_cancel)
DEFINE
  p_prev,p_next,p_ok,p_cancel BOOLEAN
DEFINE
  lo_step T_STEP
DEFINE
  lo_return T_STEP

  LET lo_step.PREV    = NVL(p_prev,FALSE)
  LET lo_step.NEXT    = NVL(p_next,FALSE)
  LET lo_step.CONFIRM = NVL(p_ok,FALSE)
  LET lo_step.CANCEL  = NVL(p_cancel,FALSE)

  LET lo_return.* = lo_step.*

  RETURN lo_return.*

END FUNCTION

FUNCTION adzi140_check_data_type_could_modify(p_ship_notice,p_dgenv,p_new_dzeb006,p_old_dzeb006)
DEFINE
  p_ship_notice  STRING,
  p_dgenv        STRING,
  p_new_dzeb006  STRING,
  p_old_dzeb006  STRING
DEFINE
  ls_ship_notice STRING,
  ls_dgenv       STRING,
  ls_new_dzeb006 STRING,
  ls_old_dzeb006 STRING,
  ls_replace_arg STRING,
  ls_err_type    STRING, #20181224
  lb_result      BOOLEAN
DEFINE
  lb_return BOOLEAN

  LET ls_ship_notice = p_ship_notice
  LET ls_dgenv       = p_dgenv
  LET ls_new_dzeb006 = p_new_dzeb006
  LET ls_old_dzeb006 = p_old_dzeb006

  LET ls_replace_arg = "|"

  LET lb_result = TRUE

  #已出貨欄位才做判斷
  IF ls_ship_notice = cs_column_shipped THEN
    #標準環境
    IF ls_dgenv = cs_dgenv_standard THEN
      CALL sadzi140_db_column_type_validation(ls_new_dzeb006,ls_old_dzeb006) RETURNING lb_result,ls_err_type #20181224
      IF NOT lb_result THEN
        CALL sadzp000_msg_show_error('adz-00555', 'adz-00555', ls_replace_arg, 1)
      END IF
    END IF
    #客制環境
    IF ls_dgenv = cs_dgenv_customize THEN
      CALL sadzi140_db_column_type_validation(ls_new_dzeb006,ls_old_dzeb006) RETURNING lb_result,ls_err_type #20181224
      IF NOT lb_result THEN
        CALL sadzp000_msg_show_error('adz-00556', 'adz-00556', ls_replace_arg, 1)
      END IF
    END IF
  END IF

  LET lb_return = lb_result

  RETURN lb_return

END FUNCTION

FUNCTION adzi140_check_attribute_could_modify(p_ship_notice,p_dgenv)
DEFINE
  p_ship_notice STRING,
  p_dgenv       STRING
DEFINE
  ls_ship_notice STRING,
  ls_dgenv       STRING,
  ls_replace_arg STRING
DEFINE
  lb_return BOOLEAN

  LET ls_ship_notice = p_ship_notice
  LET ls_dgenv       = p_dgenv

  LET ls_replace_arg = "|"

  IF ls_ship_notice = "Y" THEN
    IF ls_dgenv = cs_dgenv_standard THEN
      CALL sadzp000_msg_show_error('adz-00398', 'adz-00398', ls_replace_arg, 1)
    END IF
    IF ls_dgenv = cs_dgenv_customize THEN
      CALL sadzp000_msg_show_error('adz-00399', 'adz-00399', ls_replace_arg, 1)
    END IF
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF

  RETURN lb_return

END FUNCTION

FUNCTION adzi140_set_altered_table_info(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_table_name STRING,
  lo_dzey_t     T_DZEY_T

  LET ls_table_name = p_table_name

  LET lo_dzey_t.DZEY002 = ls_table_name
  
  BEGIN WORK 
  CALL sadzi140_crud_insert_update_dzey_t(lo_dzey_t.*)
  COMMIT WORK

END FUNCTION

FUNCTION adzi140_check_if_need_user_define_column(p_table_type)
DEFINE
  p_table_type  STRING
DEFINE
  ls_table_type STRING,
  ls_sql        STRING,
  li_rec_count  INTEGER
DEFINE
  lb_return  BOOLEAN

  LET ls_table_type = p_table_type.trim()

  IF ls_table_type MATCHES '[MTDB]' THEN 
    LET lb_return = TRUE
  ELSE
    LET lb_return = FALSE
  END IF

  RETURN lb_return

END FUNCTION

FUNCTION adzi140_check_user_define_column_if_exists(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_count  INTEGER
DEFINE
  lb_return  BOOLEAN

  LET ls_table_name = p_table_name

  #取得資料筆數
  LET ls_sql = "select count(*)                                  ",
               "  from dzeb_t eb                                 ",
               " where eb.dzeb022 like '",cs_cdf_user_define,"%' ",
               "   and eb.dzeb001 = '",ls_table_name,"'          "

  PREPARE lpre_check_user_define_column_if_exists FROM ls_sql
  DECLARE lcur_check_user_define_column_if_exists CURSOR FOR lpre_check_user_define_column_if_exists
  OPEN lcur_check_user_define_column_if_exists
  FETCH lcur_check_user_define_column_if_exists INTO li_rec_count
  CLOSE lcur_check_user_define_column_if_exists
  FREE lcur_check_user_define_column_if_exists
  FREE lpre_check_user_define_column_if_exists

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF

  RETURN lb_return

END FUNCTION

FUNCTION adzi140_get_base_table_info(p_create_table)
DEFINE 
  p_create_table   T_DZEA_CREATE_TABLE
DEFINE 
  lo_create_table     T_DZEA_CREATE_TABLE,
  lo_base_table_info  T_BASE_TABLE_INFO
DEFINE
  lo_return   T_BASE_TABLE_INFO

  LET lo_create_table.* = p_create_table.*

  LET lo_base_table_info.bti_if_rename         = TRUE 
  LET lo_base_table_info.bti_table_name        = lo_create_table.dct_table_name
  LET lo_base_table_info.bti_table_type        = lo_create_table.dct_table_type
  LET lo_base_table_info.bti_table_description = lo_create_table.dct_table_description
  LET lo_base_table_info.bti_module_name       = lo_create_table.dct_module_name

  LET lo_return.* = lo_base_table_info.*
  
  RETURN lo_return.*
  
END FUNCTION 

FUNCTION adzi140_set_column_alter_log(p_t_dzeb_t,p_alter_type)
DEFINE
  p_t_dzeb_t   T_DZEB_T,
  p_alter_type STRING 
DEFINE
  lo_t_dzeb_t    T_DZEB_T,
  ls_alter_type  STRING,
  ls_msg_no      STRING, 
  ls_message     STRING,
  ls_replace_arg STRING,
  lb_result      BOOLEAN

  LET lo_t_dzeb_t.* = p_t_dzeb_t.*
  LET ls_alter_type = p_alter_type

  IF ls_alter_type IS NULL THEN RETURN END IF
  
  CASE
    WHEN ls_alter_type = cs_data_added
      LET ls_msg_no = "adz-00873"
    WHEN ls_alter_type = cs_data_modified
      LET ls_msg_no = "adz-00874"
    WHEN ls_alter_type = cs_data_delete
      LET ls_msg_no = "adz-00875"
  END CASE 

  LET ls_message = sadzp000_msg_get_message(ls_msg_no,ms_lang)
  LET ls_replace_arg = lo_t_dzeb_t.dzeb001,",",lo_t_dzeb_t.dzeb002,",",lo_t_dzeb_t.dzeb003,",",lo_t_dzeb_t.dzeb006,",",lo_t_dzeb_t.dzeb007,",",lo_t_dzeb_t.dzeb008,"|"
  LET ls_message = sadzp000_msg_replace_message(ls_message,ls_replace_arg)

  CALL sadzi140_logs_write_log(cs_logs_level_information,cs_logs_type_column,ls_message) RETURNING lb_result

END FUNCTION

#20160617 begin
FUNCTION adzi140_check_delete_table_if_exists(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name  STRING,
  ls_sql         STRING,
  li_rec_count   INTEGER,
  lo_DZHH_T      T_DZHH_T
DEFINE
  lo_return  T_DZHH_T

  LET ls_table_name  = p_table_name

  LET ls_sql = "select hh.dzhh001,hh.dzhh002,hh.dzhh003, ",
               "       hh.dzhh004,hh.dzhh005             ",
               "  from dzhh_t hh                         ",
               " where hh.dzhh001 = '",ls_table_name,"'  "

  PREPARE lpre_check_delete_table_if_exists FROM ls_sql
  DECLARE lcur_check_delete_table_if_exists CURSOR FOR lpre_check_delete_table_if_exists
  OPEN lcur_check_delete_table_if_exists
  FETCH lcur_check_delete_table_if_exists INTO lo_DZHH_T.*
  CLOSE lcur_check_delete_table_if_exists
  FREE lcur_check_delete_table_if_exists
  FREE lpre_check_delete_table_if_exists

  LET lo_return.* = lo_DZHH_T.*

  RETURN lo_return.*

END FUNCTION
#20160617 end

#20161024 begin
FUNCTION adzi140_parse_search_contents(p_search_string)
DEFINE
  p_search_string STRING
DEFINE
  ls_search_string STRING,
  li_loop     INTEGER,
  ls_char     STRING,
  ls_contant  VARCHAR(1024),
  li_curr     INTEGER,
  ls_full_sql STRING,
  lo_search_contents DYNAMIC ARRAY OF T_SEARCH_CONTENTS
DEFINE
  ls_return STRING  

  LET ls_search_string = sadzi140_util_replace_string(p_search_string,"'","")

  LET ls_contant = ""
  FOR li_loop = 1 TO ls_search_string.getLength()
    LET ls_char = ls_search_string.getCharAt(li_loop) --.subString(li_loop,li_loop)
    IF ls_char = "," THEN
      CALL sadzi140_util_trim_str(ls_contant) RETURNING ls_contant
      LET li_curr = lo_search_contents.getLength()+1
      LET lo_search_contents[li_curr].sc_contents = ls_contant
      LET lo_search_contents[li_curr].sc_sql      = adzi140_generate_contant_sql(ls_contant) 
      LET ls_contant = ""
    ELSE 
      LET ls_contant = ls_contant,ls_char  
    END IF
  END FOR

  CALL sadzi140_util_trim_str(ls_contant) RETURNING ls_contant
  IF ls_contant IS NOT NULL THEN 
    LET li_curr = lo_search_contents.getLength()+1
    LET lo_search_contents[li_curr].sc_contents = ls_contant
    LET lo_search_contents[li_curr].sc_sql      = adzi140_generate_contant_sql(ls_contant) 
    LET ls_contant = ""
  END IF  

  FOR li_loop = 1 TO lo_search_contents.getLength()
    LET ls_full_sql = ls_full_sql,lo_search_contents[li_loop].sc_sql,"\n"
  END FOR 

  LET ls_full_sql = ls_full_sql.subString(ls_full_sql.getIndexOf("or",1)+2,ls_full_sql.getLength())
  
  LET ls_return = " and ( ",NVL(ls_full_sql,"1=1")," )"

  RETURN ls_return
  
END FUNCTION

FUNCTION adzi140_generate_contant_sql(p_contant)
DEFINE
  p_contant STRING
DEFINE
  ls_content     STRING,
  li_loop        INTEGER,
  ls_char        STRING,
  ls_obj_name    STRING,
  ls_sql         STRING,
  ls_all_sql     STRING,
  ls_contant_sql STRING
DEFINE 
  ls_return  STRING  

  LET ls_content = p_contant

  FOR li_loop = 1 TO ls_content.getLength()
    LET ls_char = ls_content.getCharAt(li_loop) --subString(li_loop,li_loop)
    IF ls_char = "|" THEN
      CALL adzi140_parse_table_condition(ls_obj_name) RETURNING ls_sql
      LET ls_all_sql = ls_all_sql,ls_sql,"\n"
      LET ls_sql = ""
      LET ls_obj_name = ""
    ELSE 
      LET ls_obj_name = ls_obj_name,ls_char.trim()
    END IF
  END FOR

  IF ls_obj_name IS NOT NULL THEN 
    CALL adzi140_parse_table_condition(ls_obj_name) RETURNING ls_sql
    LET ls_all_sql = ls_all_sql,ls_sql,"\n"
    LET ls_sql = ""
    LET ls_obj_name = ""
  END IF    

  LET ls_return = ls_all_sql

  RETURN ls_return  
  
END FUNCTION

FUNCTION adzi140_parse_table_condition(p_obj_name)
DEFINE 
  p_obj_name STRING 
DEFINE 
  ls_obj_name STRING,
  ls_sql      STRING
DEFINE
  ls_return STRING  

  LET ls_obj_name = p_obj_name.trim()

  IF ls_obj_name.getIndexOf("*",1) > 0 THEN
    CALL sadzi140_util_replace_string(ls_obj_name,"*","%") RETURNING ls_obj_name
    LET ls_sql = " or ea.dzea001 like '",ls_obj_name,"' "
  ELSE
    LET ls_sql = " or ea.dzea003 = '",ls_obj_name.toUpperCase(),"'      ",
                 " or ea.dzea003 like '%",ls_obj_name.toUpperCase(),"%' ",
                 " or ea.dzea001 like '%",ls_obj_name,"%'               ",
                 " or eal.dzeal003 like '%",ls_obj_name,"%'             "

  END IF

  LET ls_return = ls_sql

  RETURN ls_return
  
END FUNCTION  
#20161024 end

#20161229 begin
FUNCTION adzi140_get_version()
  DISPLAY "\n",cs_program_title," version : ",cs_version,"\n"
END FUNCTION
#20161229 end

#20170425 begin
FUNCTION adzi140_get_check_out_table_list(p_user_id)
DEFINE
  p_user_id STRING
DEFINE
  ls_user_id     STRING,
  ls_sql         STRING,
  ls_table_name  VARCHAR(50),
  ls_table_list  STRING,
  li_count       INTEGER
DEFINE
  ls_return STRING  

  LET ls_user_id = p_user_id

  LET li_count = 1

  LET ls_sql = "select lm.dzlm002                    ",
               "  from dzlm_t lm                     ",
               " where 1=1                           ",
               "   and lm.dzlm001 = 'T'              ",
               "   and lm.dzlm008 = 'O'              ",
               "   and lm.dzlm007 = '",ls_user_id,"' "  

  PREPARE lpre_get_check_out_table_list FROM ls_sql
  DECLARE lcur_get_check_out_table_list CURSOR FOR lpre_get_check_out_table_list

  OPEN lcur_get_check_out_table_list
  FOREACH lcur_get_check_out_table_list INTO ls_table_name
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET ls_table_list = ls_table_list,ls_table_name,","
    
    LET li_count = li_count + 1

  END FOREACH
  CLOSE lcur_get_check_out_table_list

  FREE lpre_get_check_out_table_list
  FREE lcur_get_check_out_table_list  

  LET ls_return = ls_table_list

  RETURN ls_return
  
END FUNCTION
#20170425 end

#20170912 begin
FUNCTION adzi140_check_if_duplicate_index_columns(p_dzec_t)
DEFINE
  p_dzec_t  T_DZEC_T
DEFINE
  lo_dzec_t        T_DZEC_T,
  li_idx_cnt       INTEGER, 
  ls_txt_idx_list  STRING, 
  ls_arr_idx_list  STRING, 
  lb_duplicate_idx BOOLEAN,
  ls_replace_arg   STRING,
  lo_resort_cols   DYNAMIC ARRAY OF T_RESORT_COLUMN, 
  lo_arr_cols      T_RESORT_COLUMN, 
  lo_txt_cols      T_RESORT_COLUMN
DEFINE
  lb_return  BOOLEAN  

  LET lo_dzec_t.* = p_dzec_t.*

  LET lb_duplicate_idx = FALSE
  
  CALL sadzi140_db_get_table_index_list(lo_dzec_t.dzec001,FALSE) RETURNING lo_resort_cols
  #比對匯入的索引的欄位清單是否有重複, 若有, 則匯入的資料設為失效
  IF lo_resort_cols.getLength() > 0 AND NVL(lo_dzec_t.dzecstus,"Y") <> "N" THEN 
    LET lo_txt_cols.rc_TABLE_NAME  = lo_dzec_t.dzec001
    LET lo_txt_cols.rc_INDEX_NAME  = lo_dzec_t.dzec002
    LET lo_txt_cols.rc_COLUMN_LIST = lo_dzec_t.dzec004
    CALL sadzi140_db_get_resort_index_columns_list(lo_txt_cols.*,cs_resort_index_type_text) RETURNING ls_txt_idx_list
    FOR li_idx_cnt = 1 TO lo_resort_cols.getLength()
      CALL sadzi140_db_get_resort_index_columns_list(lo_resort_cols[li_idx_cnt].*,cs_resort_index_type_text) RETURNING ls_arr_idx_list
      IF ls_arr_idx_list = ls_txt_idx_list THEN
        LET lb_duplicate_idx = TRUE
        EXIT FOR 
      END IF
    END FOR
  END IF

  LET lb_return = lb_duplicate_idx

  RETURN lb_return
  
END FUNCTION
#20170912 end

#20180903 begin
FUNCTION adzi140_check_if_enterprise_field_exists(p_module_name,p_dzeb_t)
DEFINE
  p_module_name STRING,
  p_dzeb_t      DYNAMIC ARRAY OF T_DZEB_T
DEFINE
  ls_module_name STRING,
  lo_dzeb_t      DYNAMIC ARRAY OF T_DZEB_T,
  li_loop        INTEGER
DEFINE
  lb_return BOOLEAN  

  LET ls_module_name = p_module_name.toUpperCase()
  LET lo_dzeb_t      = p_dzeb_t

  LET lb_return = FALSE

  IF (ls_module_name = "ADZ" OR ls_module_name = "AZZ" OR ls_module_name = "AWS" OR ls_module_name = "ARP") THEN
    LET lb_return = TRUE
  ELSE
    FOR li_loop = 1 TO lo_dzeb_t.getLength()
      IF lo_dzeb_t[li_loop].dzeb022 = cs_cdf_enterprise THEN
        LET lb_return = TRUE
        EXIT FOR 
      END IF
    END FOR
  END IF  

  RETURN lb_return
  
END FUNCTION

FUNCTION adzi140_check_common_fields_if_exists(p_table_name,p_table_type,p_module_name,p_dzeb_t,p_collect_messages)
DEFINE
  p_table_name  STRING,
  p_table_type  STRING,
  p_module_name STRING,
  p_dzeb_t      DYNAMIC ARRAY OF T_DZEB_T,
  p_collect_messages DYNAMIC ARRAY OF T_COLLECT_MESSAGE
DEFINE
  ls_table_name  STRING,
  ls_table_type  STRING,
  ls_module_name STRING,
  lo_dzeb_t      DYNAMIC ARRAY OF T_DZEB_T,
  lo_collect_messages DYNAMIC ARRAY OF T_COLLECT_MESSAGE,
  lb_result      BOOLEAN
DEFINE
  lo_return  DYNAMIC ARRAY OF T_COLLECT_MESSAGE    

  LET ls_table_name  = p_table_name
  LET ls_table_type  = p_table_type
  LET ls_module_name = p_module_name
  LET lo_dzeb_t      = p_dzeb_t
  LET lo_collect_messages = p_collect_messages
  
  -- 檢查是否需含共用欄位
  CALL adzi140_check_if_common_column_data_exists(ls_table_name,ls_table_type,ms_lang,lo_dzeb_t) RETURNING lb_result 
  IF NOT lb_result THEN
    CALL sadzp000_msg_set_collect_message('adz-00590',NULL,cs_msg_work_type_add,ms_lang,lo_collect_messages) RETURNING lo_collect_messages
  END IF      
  
  -- 檢查是否需含自定義欄位(標準環境及非 topstd 才檢核)
  IF (ms_DGENV = cs_dgenv_standard) AND (ms_user <> cs_topstd_user_name) THEN
    CALL adzi140_check_if_user_define_column_data_exists(ls_table_name,ls_table_type,lo_dzeb_t) RETURNING lb_result
    IF NOT lb_result THEN
      CALL sadzp000_msg_set_collect_message('adz-00591',NULL,cs_msg_work_type_add,ms_lang,lo_collect_messages) RETURNING lo_collect_messages
    END IF      
  END IF

  #檢查Enterprise欄位是否有輸入
  CALL adzi140_check_if_enterprise_field_exists(ls_module_name,lo_dzeb_t) RETURNING lb_result
  IF NOT lb_result THEN
    CALL sadzp000_msg_set_collect_message('adz-01432',NULL,cs_msg_work_type_add,ms_lang,lo_collect_messages) RETURNING lo_collect_messages
  END IF      

  LET lo_return = lo_collect_messages

  RETURN lo_return
  
END FUNCTION 

#20180903 end

#20190219 begin
FUNCTION adzi140_check_column_attribute_valid(p_t_dzeb_t,p_main_module,p_dgenv,p_old_dzeb006)
DEFINE
  p_t_dzeb_t     T_DZEB_T,
  p_main_module  STRING,
  p_dgenv        STRING,
  p_old_dzeb006  STRING
DEFINE
  lo_t_dzeb_t     T_DZEB_T,
  ls_main_module  STRING,
  ls_dgenv        STRING,
  ls_old_dzeb006  STRING,
  ls_new_dzeb006  STRING,
  ls_ship_notice  STRING,
  ls_replace_arg  STRING,
  ls_column_name  STRING,
  ls_prog_list    STRING,
  lb_result       BOOLEAN
DEFINE
  lb_return      BOOLEAN,
  ls_rtn_old_006 STRING,
  lo_return      T_DZEB_T
  
  LET lo_t_dzeb_t.*  = p_t_dzeb_t.*
  LET ls_main_module = p_main_module
  LET ls_dgenv       = p_dgenv
  LET ls_old_dzeb006 = p_old_dzeb006

  LET lb_return = TRUE

  LET lo_t_dzeb_t.RECORDTYPE = NVL(lo_t_dzeb_t.RECORDTYPE,cs_data_modified)
  #檢核是否可異動欄位屬性(不為自定義欄位才檢核)
  LET ls_new_dzeb006 = lo_t_dzeb_t.dzeb006
  LET ls_ship_notice = lo_t_dzeb_t.dzeb031
  CALL adzi140_check_data_type_could_modify(ls_ship_notice,ls_dgenv,ls_new_dzeb006,ls_old_dzeb006) RETURNING lb_result
  IF NOT lb_result THEN
    LET lo_t_dzeb_t.dzeb006 = ls_old_dzeb006
    LET lb_return = FALSE
  ELSE
    #設定客制標示, 已經設為客制的則不再重設
    IF lo_t_dzeb_t.dzeb029 <> cs_dgenv_customize THEN
      LET lo_t_dzeb_t.dzeb029 = ls_dgenv
    END IF
  END IF

  #20170413 begin
  #判斷資料屬性名稱是否符合模組, ADZ只能用 "Z" 開頭的
  IF ls_main_module.toUpperCase() = "ADZ" THEN
    IF ls_new_dzeb006.subString(1,1) <> "Z" THEN
      LET ls_replace_arg = "|"
      CALL sadzp000_msg_show_error('adz-00992', 'adz-00992', ls_replace_arg, 0) #ADZ模組只能使用 Z 開頭的資料屬性
      LET lo_t_dzeb_t.dzeb006 = ""
      CALL adzi140_on_change_dzeb006(lo_t_dzeb_t.*) RETURNING lo_t_dzeb_t.*
      LET lb_return = FALSE
    END IF
  ELSE
    IF ls_new_dzeb006.subString(1,1) = "Z" THEN
      LET ls_replace_arg = "|"
      CALL sadzp000_msg_show_error('adz-00993', 'adz-00993', ls_replace_arg, 0) #非ADZ模組不能使用 Z 開頭的資料屬性
      LET lo_t_dzeb_t.dzeb006 = ""
      CALL adzi140_on_change_dzeb006(lo_t_dzeb_t.*) RETURNING lo_t_dzeb_t.*
      LET lb_return = FALSE
    END IF
  END IF
  #20170413 end
          
  LET ls_column_name = lo_t_dzeb_t.dzeb002
  LET lo_t_dzeb_t.RECORDTYPE = NVL(lo_t_dzeb_t.RECORDTYPE,cs_data_modified)
  CALL adzi140_on_change_dzeb006(lo_t_dzeb_t.*) RETURNING lo_t_dzeb_t.*
  IF lo_t_dzeb_t.dzeb006 IS NULL THEN 
    LET lb_return = FALSE
  ELSE 
    CALL adzi140_get_DZAC_prog_list(ls_column_name) RETURNING ls_prog_list
    --DISPLAY lo_t_dzeb_t.* TO sr_TableColumns[li_arr_curr].*
    LET ls_new_dzeb006 = lo_t_dzeb_t.dzeb006
    IF (ls_new_dzeb006 <> ls_old_dzeb006) AND (ls_prog_list.trim() IS NOT NULL) THEN
      LET ls_old_dzeb006 = lo_t_dzeb_t.dzeb006
      LET ls_replace_arg = ls_column_name,"|",ls_prog_list,"|"
      CALL sadzp000_msg_show_info('adz-00244', 'adz-00244', ls_replace_arg, 0) #您修改了欄位%1的欄位屬性，將會影響底下所列程式的規格設計資料：%2
    END IF
  END IF  

  LET lo_return.* = lo_t_dzeb_t.*
  LET ls_rtn_old_006 = ls_old_dzeb006
  
  RETURN lb_return,ls_rtn_old_006,lo_return.*

END FUNCTION
#20190219 end

#20190304 begin
FUNCTION adzi140_get_column_numeric_digit(p_column_name)
DEFINE
  p_column_name  STRING
DEFINE
  ls_column_name  STRING,
  li_count        INTEGER,
  ls_digit        STRING
DEFINE
  li_rtn_begin  INTEGER,
  li_rtn_end    INTEGER

  LET ls_column_name = p_column_name

  LET li_rtn_begin = 0
  LET li_rtn_end   = 0
  
  FOR li_count = 1 TO ls_column_name.getLength()
    LET ls_digit = ls_column_name.getCharAt(li_count)

    IF ls_digit.trim() MATCHES "[0-9]" AND li_rtn_begin = 0 THEN
      LET li_rtn_begin = li_count
    END IF

    IF ls_digit.trim() MATCHES "[a-z]" AND li_rtn_begin > 0 AND li_rtn_end = 0 THEN
      LET li_rtn_end = li_count
      IF ls_digit.trim() MATCHES "[a-z]" THEN
        LET li_rtn_end = li_rtn_end - 1
        EXIT FOR
      END IF
    END IF
    
  END FOR

  LET li_rtn_end = IIF(li_rtn_end==0,ls_column_name.getLength(),li_rtn_end)

  RETURN li_rtn_begin,li_rtn_end
  
END FUNCTION

FUNCTION adzi140_get_max_serial_number(p_dzeb_t)
DEFINE 
  p_dzeb_t DYNAMIC ARRAY OF T_DZEB_T
DEFINE
  li_count         INTEGER,
  li_curr_srn      INTEGER,
  li_max_srn       INTEGER,
  lo_dzeb027_json  T_DZEB027_JSON
DEFINE
  li_return  INTEGER  

  LET li_max_srn = 0
  
  #抓取流水號最大值
  FOR li_count = 1 TO p_dzeb_t.getLength()
    IF p_dzeb_t[li_count].dzeb022 = cs_cdf_serial_number THEN 
      CALL adzi140_parse_dzeb027_json_data(p_dzeb_t[li_count].*) RETURNING lo_dzeb027_json.*

      CASE    
        WHEN ms_DGENV = cs_dgenv_standard
          CASE 
            WHEN lo_dzeb027_json.COL_TYPE = ms_topind AND ms_topind <> cs_default_topind
              LET li_curr_srn = lo_dzeb027_json.SER_NO --p_dzeb_t[li_count].dzeb023
              IF li_curr_srn > li_max_srn THEN
                LET li_max_srn = li_curr_srn
              END IF
            WHEN ms_topind = cs_default_topind    
              LET li_curr_srn = p_dzeb_t[li_count].dzeb023
              IF li_curr_srn > li_max_srn THEN
                LET li_max_srn = li_curr_srn
              END IF
          END CASE
        WHEN ms_DGENV = cs_dgenv_customize
          IF lo_dzeb027_json.COL_TYPE = ms_column_user_alter_code THEN
            LET li_curr_srn = lo_dzeb027_json.SER_NO --p_dzeb_t[li_count].dzeb023
            IF li_curr_srn > li_max_srn THEN
              LET li_max_srn = li_curr_srn
            END IF
          END IF
      END CASE

      {
      IF (p_dzeb_t[li_count].dzeb023 IS NOT NULL) AND (p_dzeb_t[li_count].dzeb022 = cs_cdf_serial_number) THEN
        LET li_curr_srn = p_dzeb_t[li_count].dzeb023
        IF li_curr_srn > li_max_srn THEN
          LET li_max_srn = li_curr_srn
        END IF
      END IF
      }
    END IF
  END FOR

  LET li_return = li_max_srn --IIF(li_max_srn==0,2,li_max_srn)

  RETURN li_return
  
END FUNCTION

FUNCTION adzi140_get_dzeb027_json_data(p_top_ind,p_ser_no,p_col_type)
DEFINE 
  p_top_ind   STRING,
  p_ser_no    INTEGER,
  p_col_type  STRING
DEFINE 
  ls_top_ind       STRING,
  ls_ser_no        STRING,
  ls_col_type      STRING,
  lo_dzeb027_json  T_DZEB027_JSON,
  lo_json_obj      util.JSONObject,
  ls_json_str      STRING
DEFINE
  ls_return  STRING  

  LET ls_top_ind  = p_top_ind
  LET ls_ser_no   = p_ser_no
  LET ls_col_type = p_col_type

  LET lo_dzeb027_json.TOPIND   = ls_top_ind
  LET lo_dzeb027_json.SER_NO   = ls_ser_no
  LET lo_dzeb027_json.COL_TYPE = ls_col_type

  LET lo_json_obj = util.JSONObject.fromFGL(lo_dzeb027_json)
  LET ls_json_str = lo_json_obj.toString() 
  
  LET ls_return = ls_json_str

  RETURN ls_return

END FUNCTION

FUNCTION adzi140_parse_dzeb027_json_data(p_dzeb_t)
DEFINE 
  p_dzeb_t T_DZEB_T
DEFINE 
  lo_dzeb_t        T_DZEB_T,
  ls_dzeb027       STRING,
  lo_dzeb027_json  T_DZEB027_JSON,
  lo_json_obj      util.JSONObject
DEFINE
  lo_return  T_DZEB027_JSON  

  LET lo_dzeb_t.* = p_dzeb_t.*
  LET ls_dzeb027 = lo_dzeb_t.dzeb027
  LET ls_dzeb027 = ls_dzeb027.trim()

  IF ls_dzeb027.trim() IS NOT NULL THEN
    TRY  
      LET lo_json_obj = util.JSONObject.parse(ls_dzeb027)
      LET lo_dzeb027_json.TOPIND   = lo_json_obj.get("TOPIND")
      LET lo_dzeb027_json.SER_NO   = lo_json_obj.get("SER_NO")
      LET lo_dzeb027_json.COL_TYPE = lo_json_obj.get("COL_TYPE")
      IF lo_dzeb027_json.TOPIND IS NULL OR lo_dzeb027_json.SER_NO IS NULL THEN
        LET lo_dzeb027_json.TOPIND   = NULL
        LET lo_dzeb027_json.SER_NO   = 0
        LET lo_dzeb027_json.COL_TYPE = NULL
      END IF 
    CATCH
      DISPLAY cs_warning_tag,"Parse JSON for column ",lo_dzeb_t.dzeb002," failed."
      LET lo_dzeb027_json.TOPIND   = NULL
      LET lo_dzeb027_json.SER_NO   = 0
      LET lo_dzeb027_json.COL_TYPE = NULL
    END TRY
  ELSE  
    DISPLAY cs_warning_tag,"Column ",lo_dzeb_t.dzeb002," JSON data is empty."
    LET lo_dzeb027_json.TOPIND   = NULL
    LET lo_dzeb027_json.SER_NO   = 0
    LET lo_dzeb027_json.COL_TYPE = NULL
  END IF  
  
  LET lo_return.*  = lo_dzeb027_json.*

  RETURN lo_return.*

END FUNCTION

FUNCTION adzi140_get_custom_code(p_table_name,p_column_name,p_column_user_define_code)
DEFINE
  p_table_name  STRING,
  p_column_name STRING,
  p_column_user_define_code STRING
DEFINE
  ls_table_name   STRING,
  ls_column_name  STRING,
  ls_column_user_define_code STRING,
  ls_leading_code STRING,
  ls_remain_code  STRING,
  li_loop         INTEGER,
  lb_num_col      BOOLEAN,
  ls_char         STRING,
  ls_pad_char     STRING
DEFINE
  ls_return  STRING  

  LET ls_table_name  = p_table_name
  LET ls_column_name = p_column_name
  LET ls_column_user_define_code = p_column_user_define_code

  LET ls_pad_char = ""
  LET lb_num_col  = FALSE
  
  CALL sadzi140_util_get_table_leading_code(ls_table_name) RETURNING ls_leading_code
  LET ls_remain_code = ls_column_name.subString(ls_leading_code.getLength()+1,ls_column_name.getLength())

  FOR li_loop = 1 TO ls_remain_code.getLength()
    LET ls_char = ls_remain_code.getCharAt(li_loop)
    IF ls_char MATCHES "[0-9]" THEN
      LET lb_num_col = TRUE
      EXIT FOR 
    ELSE
      LET ls_pad_char = ls_pad_char,ls_char
    END IF
  END FOR 

  IF (NOT lb_num_col) OR (ls_pad_char = ls_column_user_define_code) THEN
    LET ls_pad_char = ""
  END IF

  LET ls_return = ls_pad_char
  
  RETURN ls_return
  
END FUNCTION
#20190304 end

#20190819 begin
FUNCTION adzi140_get_table_pk_cols_cnts(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_count  INTEGER
DEFINE
  li_return  INTEGER

  LET ls_table_name = p_table_name

  #取得資料筆數
  LET ls_sql = "select count(1)                         ", 
               "  from dzeb_t eb                        ", 
               " where eb.dzeb001 = '",ls_table_name,"' ",  
               "   and eb.dzeb004 = 'Y'                 "  
   
  PREPARE lpre_get_table_pk_cols_cnts FROM ls_sql
  DECLARE lcur_get_table_pk_cols_cnts CURSOR FOR lpre_get_table_pk_cols_cnts
  OPEN lcur_get_table_pk_cols_cnts
  FETCH lcur_get_table_pk_cols_cnts INTO li_rec_count
  CLOSE lcur_get_table_pk_cols_cnts
  FREE lcur_get_table_pk_cols_cnts
  FREE lpre_get_table_pk_cols_cnts

  LET li_return = li_rec_count

  RETURN li_return

END FUNCTION
#20190819 end
