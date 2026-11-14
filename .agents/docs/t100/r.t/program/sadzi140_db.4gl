{
  異動型態    異動單號       日 期       異動者      異動內容
  ========= ============= ========== ========== ===============================================
  Modify                  20161220   Ernestliou 1.針對客戶自行新增的欄位, 一律不處理 drop 程序
  Modify                  20170214   Ernestliou 1.表格重建時對應 Trigger 的異動
  Modify                  20170315   Ernestliou 1.修正 Trigger 重建時to_char失敗的問題
  Modify                  20170317   Ernestliou 1.Trigger Scripts 改為利用 SPOOL 匯出, 在 Table Rebuild完成後呼叫 sqlplus 執行
  Modify                  20170405   Ernestliou 1.產生 dsgrant.sql 到 /u1/t10tst/apse/std/script/ OR /u1/t10tst/apse/script/
  Modify                  20170414   Ernestliou 1.異動欄位時先設定 default 為 null
  Modify                  20170425   Ernestliou 1.Mark grant to APS warning message
  Modify                  20170523   Ernestliou 1.新增彙整訊息功能及視窗
  Modify                  20170706   Ernestliou 1.[執行異動]最終結果為'失敗'的話, 若透過[表格重建]就可以解決的話, 則提醒使用者自行處理  
  Modify                  20170912   Ernestliou 1.索引匯入時判斷有無已存在 index 欄位, 有則匯入方設失效
                                                2. About 新增顯示 OS Process ID
  Modify                  20171120   02160      1. Analyze table 之前, 先解統計 Lock
  Modify                  20180514   02160      1. 刪除表格時不要關注已經不存在於 azzi085 上的資料庫
  Modify                  20180710   02160      1.修正當更改相同型態長度時, 預設值也要重新判定
  Modify                  20181218   02160      1.修正抓取比對客制索引的來源為dzec007
  Modify                  20181224   02160      1.修正比對欄位型態或長度不同時所給予的背景訊息
  ------------------------------------------------------------------------------
  Modify                  20191209   02160      1.修改支援 MSSQLServer(1.1.071)   
  Modify                  20200330   02160      1.修改支援 PosgreSQL(1.1.072)   
  Modify                  20230606   02160      1.修改支援 MogDB 
  Modify                  20230804   02160      1.修改支援 DM
  Modify                  20240416   02160      1.若超過10個PK後就不控管了,只控管與資料庫相關的最大值
  Modify                  20240521   02160      1.修改支援 KINGBASE
  Modify                  20240719   02160      1.客戶端純客制表不管制 PK 數量
  Modify                  20241014   02160      1.針對 MogDB 效能問題修正抓取 backdoor 的 SQL
  Modify                  20241021   02160      1.修改 pk limit 為 11
  Modify                  20250113   02160      1.修正在 PostgreSQL 環境下, 當型態從 int4 轉成 decimal (int4 -> decimal(20,6))時會失敗的問題
}  

&include "../4gl/sadzi140_mcr.inc"

IMPORT util
IMPORT os 
IMPORT security

SCHEMA ds

&include "../4gl/sadzp000_cnst.inc"
&include "../4gl/sadzi140_cnst.inc"
&include "../4gl/sadzp200_cnst.inc"

&include "../4gl/sadzi140_type.inc"

CONSTANT cs_alter_column_section STRING = "--alter column sec "

PRIVATE TYPE T_COLUMN_DIFF_W_COMMENT RECORD
              dzev005  LIKE dzev_t.dzev005,
              dzev006  LIKE dzev_t.dzev006,
              dzev007  LIKE dzev_t.dzev007,
              dzev008  LIKE dzev_t.dzev008,
              dzev009  LIKE dzev_t.dzev009,
              dzev021  LIKE dzev_t.dzev021,
              comments VARCHAR(4000)
            END RECORD

&ifndef DEBUG
GLOBALS "../../cfg/top_global.inc"
&else
GLOBALS "../../cfg/top_global.inc"
&endif

DEFINE
  mo_columns_data_in_mem DYNAMIC ARRAY OF T_COLUMNS_DATA_IN_MEM

FUNCTION sadzi140_db_gen_table_schema(p_table_name,p_module_name,p_batch_mode,p_lang)
DEFINE
  p_table_name  STRING,
  p_module_name STRING,
  p_batch_mode  BOOLEAN,
  p_lang        STRING
DEFINE
  ls_table_name       STRING,
  ls_module_name      STRING,
  lb_batch_mode       BOOLEAN,  
  ls_lang             STRING,
  ##########################
  ls_table_memo       STRING,
  ls_table_contents   STRING,
  ls_table_constraint STRING,
  ls_table_index      STRING,
  ls_table_grant      STRING,
  ls_exit_sign        STRING,
  ls_table_full       STRING,
  ##########################
  ls_erp_env          STRING,
  lchannel_write      base.Channel,
  ls_schema_filename  STRING,
  ls_replace_arg      STRING,
  ls_separator        STRING
  
  LET ls_table_name  = p_table_name
  LET ls_module_name = p_module_name
  LET lb_batch_mode  = p_batch_mode
  LET ls_lang        = p_lang

  LET ls_separator = os.Path.separator()
  
  &ifndef DEBUG
  LET ls_erp_env = FGL_GETENV(ls_module_name)
  LET ls_schema_filename = ls_erp_env,ls_separator,"sch",ls_separator,ls_lang,ls_separator,ls_table_name,".sch"
  &else
  LET ls_erp_env = FGL_GETENV("TEMPDIR")
  LET ls_schema_filename = ls_erp_env,ls_separator,ls_table_name,"_",ls_lang,".sch"
  &endif
  
  LET ls_exit_sign = sadzi140_util_get_exit_sign() --cs_exit_sig --"exit;"
  
  LET lchannel_write = base.Channel.create()
  CALL lchannel_write.setDelimiter("")

  CALL sadzi140_db_gen_table_memo(ls_table_name) RETURNING ls_table_memo
  CALL sadzi140_db_gen_table_contents(ls_table_name) RETURNING ls_table_contents
  CALL sadzi140_db_gen_table_constraint("",ls_table_name,FALSE) RETURNING ls_table_constraint
  CALL sadzi140_db_gen_table_index("",ls_table_name,FALSE) RETURNING ls_table_index
  CALL sadzi140_db_gen_table_grant(ls_table_name) RETURNING ls_table_grant

  LET ls_table_full = ls_table_memo,"\n",
                      ls_table_contents,"\n",
                      ls_table_constraint,"\n",
                      ls_table_index,"\n",
                      ls_table_grant,"\n",
                      ls_exit_sign
  
  #開檔寫入
  TRY
    DISPLAY "Schema File : ",ls_schema_filename
    CALL lchannel_write.openFile(ls_schema_filename, "w" )
    CALL lchannel_write.write(ls_table_full)
    #關檔
    CALL lchannel_write.close()
    IF lb_batch_mode THEN
      DISPLAY cs_success_tag,"Generate Table Schema : "||ls_table_name
    ELSE
      IF lb_batch_mode THEN
        DISPLAY cs_success_tag,"Generate Table Schema : "||ls_table_name
      ELSE
        #CALL sadzi140_rev_view_logresult(ls_table_full)
        DISPLAY cs_success_tag,"Generate Table Schema : "||ls_table_name
      END IF  
    END IF  
  CATCH
    LET ls_table_full = ""
    DISPLAY cs_error_tag,"Generate Table Schema : "||ls_table_name
    IF NOT lb_batch_mode THEN
      #Create Table Schema ERROR
      LET ls_replace_arg = ls_table_name,"|",ls_module_name,"|",SQLCA.SQLCODE,"|"
      CALL sadzp000_msg_show_error(NULL, 'adz-00085', ls_replace_arg, 1)
    END IF  
  END TRY  

  RETURN ls_table_full

END FUNCTION

FUNCTION sadzi140_db_gen_table_contents(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_line           STRING,
  ls_sql            STRING,
  ls_table_name     STRING,
  li_rec_count      INTEGER,
  li_counts         INTEGER,
  ls_comma          STRING, 
  ls_COL_LENGTH     STRING,
  lo_line_str_buf   base.StringBuffer,
  lo_table_contents RECORD
                      col_name    LIKE dzeb_t.dzeb002, 
                      data_type   VARCHAR(50),
                      col_length  VARCHAR(50),
                      default_val VARCHAR(50),
                      NULLABLE    VARCHAR(50),
                      col_desc    VARCHAR(300),
                      eb_col_desc LIKE dzeb_t.dzeb003 
                    END RECORD,
  lo_table_datas RECORD
                   col_name    STRING,
                   data_type   STRING,
                   col_length  STRING,
                   default_val STRING,
                   NULLABLE    STRING,
                   col_desc    STRING
                 END RECORD
DEFINE
  ls_Return   STRING

  LET ls_table_name   = p_table_name
  LET lo_line_str_buf = base.StringBuffer.create()

  #取得資料筆數
  LET ls_sql = "select count(1)                       ",
               "  from dzeb_t eb                      ",
               " where 1=1                            ",
               "   and eb.dzeb001 = '",ls_table_name,"'" 

  PREPARE lpre_get_rec_counts FROM ls_sql
  DECLARE lcur_get_rec_counts CURSOR FOR lpre_get_rec_counts
  OPEN lcur_get_rec_counts
  FETCH lcur_get_rec_counts INTO li_rec_count
  CLOSE lcur_get_rec_counts
  FREE lcur_get_rec_counts
  FREE lpre_get_rec_counts

  #取得欄位資料
  LET ls_sql = "select eb.dzeb002       col_name,             ",      
               "       td.gztd003       data_type,            ",
               "       trim(td.gztd008) col_length,           ",
               "       ''               default_val,          ",
               "       ''               nullable,             ",  
               "       ebl.dzebl003     col_desc,             ",
               "       eb.dzeb003       eb_col_desc           ",
               "  from dzeb_t eb                              ",
               "  left outer join gztd_t td                   ",
               "               on td.gztd001 = eb.dzeb006     ",                                 
               "  left outer join dzebl_t ebl                 ",
               "               on ebl.dzebl001 = eb.dzeb002   ",
               "              and ebl.dzebl002 = '",g_lang,"' ",
               " where 1=1                                    ",
               "   and eb.dzeb001 = '",ls_table_name,"'       ",
               " order by eb.dzeb021,eb.dzeb002               " 

  PREPARE lpre_gen_table_contents FROM ls_sql
  DECLARE lcur_gen_table_contents CURSOR FOR lpre_gen_table_contents

  LET ls_line = "create table ",ls_table_name,"\n"
  CALL lo_line_str_buf.append(ls_line)
  LET ls_line = "(","\n"
  CALL lo_line_str_buf.append(ls_line)

  LET li_counts = 0
  OPEN lcur_gen_table_contents
  FOREACH lcur_gen_table_contents INTO lo_table_contents.*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_counts = li_counts + 1

    LET ls_COL_LENGTH = lo_table_contents.COL_LENGTH
    LET ls_COL_LENGTH = ls_COL_LENGTH.Trim()
    IF ls_COL_LENGTH IS NOT NULL THEN
      LET lo_table_datas.COL_LENGTH = "(",lo_table_contents.COL_LENGTH,")"
    ELSE  
      LET lo_table_datas.COL_LENGTH = lo_table_contents.COL_LENGTH
    END IF

    #2012.12.18 Remove
    {
    IF lo_table_contents.DEFAULT_VAL IS NOT NULL THEN
      LET lo_table_datas.DEFAULT_VAL = "DEFAULT ",lo_table_contents.DEFAULT_VAL
    ELSE
      LET lo_table_datas.DEFAULT_VAL = lo_table_contents.DEFAULT_VAL
    END IF

    IF lo_table_contents.NULLABLE = "N" THEN
      LET lo_table_datas.NULLABLE = "NOT NULL"
    ELSE   
      LET lo_table_datas.NULLABLE = "" 
    END IF
    }
    
    #最後一筆資料不加逗號
    IF li_counts = li_rec_count THEN
      LET ls_comma = ""
    ELSE
      LET ls_comma = ","
    END IF

    LET ls_line = lo_table_contents.col_name,7 SPACE,
                  lo_table_contents.data_type,
                  lo_table_datas.col_length,2 SPACE,
                  lo_table_datas.default_val,2 SPACE,
                  lo_table_datas.nullable,2 SPACE,
                  ls_comma,
                  sadzi140_db_get_comment_sign("B"),nvl(lo_table_contents.col_desc,lo_table_contents.eb_col_desc),sadzi140_db_get_comment_sign("E"),"\n"
    
    CALL lo_line_str_buf.append(ls_line)

  END FOREACH
  CLOSE lcur_gen_table_contents
  
  LET ls_line = ");"
  CALL lo_line_str_buf.append(ls_line)
  
  FREE lpre_gen_table_contents
  FREE lcur_gen_table_contents  

  LET ls_Return = lo_line_str_buf.toString()

  RETURN ls_Return
  
END FUNCTION

FUNCTION sadzi140_db_gen_table_memo(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_line          STRING,
  ls_sql           STRING,
  ls_table_name    STRING,
  lo_line_str_buf  base.StringBuffer,
  ls_comment_sign  STRING,
  lo_table_memo RECORD
                  tm_table_name LIKE dzea_t.dzea001,
                  tm_table_desc VARCHAR(300),
                  tm_table_memo LIKE dzea_t.dzea002,
                  tm_table_type LIKE dzea_t.dzea004,
                  tm_multi_lang LIKE dzea_t.dzea005,
                  tm_sys_table  LIKE dzea_t.dzea006
                END RECORD
DEFINE
  ls_return STRING               

  LET ls_table_name  = p_table_name
  LET lo_line_str_buf = base.StringBuffer.create()

  LET ls_sql = "select ea.dzea001    table_name,              ",
               "       eal.dzeal003  table_desc,              ",
               "       eal.dzeal004  table_memo,              ",
               "       ea.dzea004    table_type,              ",
               "       ea.dzea005    multi_lang,              ",
               "       ea.dzea006    sys_table                ",
               "  from dzea_t ea                              ",
               "  left outer join dzeal_t eal                 ",
               "               on eal.dzeal001 = ea.dzea001   ",                                      
               "              and eal.dzeal002 = '",g_lang,"' ",
               " where 1=1                                    ",
               "   and ea.dzea001 = '",ls_table_name,"'       "

  PREPARE lpre_gen_table_memo FROM ls_sql
  DECLARE lcur_gen_table_memo CURSOR FOR lpre_gen_table_memo

  LET ls_comment_sign = sadzi140_db_get_comment_sign("B")
  LET ls_comment_sign = ls_comment_sign,"\n" 
  CALL lo_line_str_buf.append(ls_comment_sign)
  
  OPEN lcur_gen_table_memo
  FOREACH lcur_gen_table_memo INTO lo_table_memo.*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET ls_line = "================================================================================","\n"
    CALL lo_line_str_buf.append(ls_line)
    LET ls_line = "檔案代號:",lo_table_memo.tm_table_name,"\n" 
    CALL lo_line_str_buf.append(ls_line)
    LET ls_line = "檔案名稱:",lo_table_memo.tm_table_desc,"\n" 
    CALL lo_line_str_buf.append(ls_line)
    LET ls_line = "檔案目的:",lo_table_memo.tm_table_memo,"\n" 
    CALL lo_line_str_buf.append(ls_line)
    LET ls_line = "上游檔案:","\n" 
    CALL lo_line_str_buf.append(ls_line)
    LET ls_line = "下游檔案:","\n" 
    CALL lo_line_str_buf.append(ls_line)
    LET ls_line = "檔案類型:",lo_table_memo.tm_table_type,"\n" 
    CALL lo_line_str_buf.append(ls_line)
    LET ls_line = "多語系檔案:",lo_table_memo.tm_multi_lang,"\n" 
    CALL lo_line_str_buf.append(ls_line)
    LET ls_line = "============.========================.==========================================","\n"  
    CALL lo_line_str_buf.append(ls_line)

  END FOREACH
  CLOSE lcur_gen_table_memo
  
  LET ls_comment_sign = sadzi140_db_get_comment_sign("E")
  CALL lo_line_str_buf.append(ls_comment_sign)

  FREE lpre_gen_table_memo
  FREE lcur_gen_table_memo  

  LET ls_return = lo_line_str_buf.toString()

  RETURN ls_return
                  
END FUNCTION

FUNCTION sadzi140_db_gen_table_constraint(p_owner,p_table_name,p_include_drop)
DEFINE
  p_owner        STRING,
  p_table_name   STRING,
  p_include_drop BOOLEAN
DEFINE
  ls_owner              STRING,
  ls_table_name         STRING,
  lb_include_drop       BOOLEAN,
  ls_line               STRING,
  ls_sql                STRING,
  lo_line_str_buf       base.StringBuffer,
  lb_constraint_exist   BOOLEAN,
  lb_index_exist        BOOLEAN,
  ls_drop_index         STRING,
  ls_enable_foreign_key STRING,
  ls_foreign_key_state  STRING,
  ls_db_type            STRING, #20240521
  ls_foreign_check      STRING, #20191209
  lo_table_constraint RECORD
                        key_table       VARCHAR(30),
                        key_name        VARCHAR(30),
                        key_type        VARCHAR(10),
                        key_columns     VARCHAR(512),
                        foreign_table   VARCHAR(30),
                        foreign_columns VARCHAR(30)
                      END RECORD
DEFINE
  ls_return STRING               

  LET ls_owner        = p_owner
  LET ls_table_name   = p_table_name
  LET lb_include_drop = p_include_drop
  
  LET lo_line_str_buf = base.StringBuffer.create()

  #取得 Foreign 是否啟動
  CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_enable_foreign_key) RETURNING ls_enable_foreign_key
  LET ls_foreign_key_state = IIF(ls_enable_foreign_key=="Y",cs_state_enable,cs_state_disable)

  LET ls_sql = "select bd.dzed001 key_table,            ",  
               "       bd.dzed002 key_name,             ",
               "       bd.dzed003 key_type,             ",
               "       bd.dzed004 key_columns,          ",
               "       bd.dzed005 foreign_table,        ",
               "       bd.dzed006 foreign_columns       ",
               "  from dzed_t bd                        ",
               " where bd.dzed001 = '",ls_table_name,"' ",
               "   and bd.dzed003 <> 'F'                "               

  PREPARE lpre_GenTableConstraint FROM ls_sql
  DECLARE lcur_GenTableConstraint CURSOR FOR lpre_GenTableConstraint

  OPEN lcur_GenTableConstraint
  FOREACH lcur_GenTableConstraint INTO lo_table_constraint.*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET ls_line = ""

    IF lo_table_constraint.KEY_TYPE = "P" THEN
      IF lb_include_drop THEN
        #檢查Constraint是否存在, 若存在則先 Drop 
        CALL sadzi140_db_check_table_constraint_exist(ls_owner,lo_table_constraint.KEY_TABLE,lo_table_constraint.key_name) RETURNING lb_constraint_exist
        IF lb_constraint_exist THEN
          &ifdef is_msv
          LET ls_line = ls_line,
                        "if object_id('dbo.[",lo_table_constraint.key_name,"]') is not null",
                        " alter table ",lo_table_constraint.KEY_TABLE," drop constraint ",lo_table_constraint.key_name,"\n","go","\n" #20191209
          &else              
          &ifdef is_pgs
          LET ls_line = ls_line,
                        "-- PostgreSQL drop primary key.","\n",
                        " alter table ",lo_table_constraint.KEY_TABLE," drop constraint ",lo_table_constraint.key_name,";","\n" #20200330
          &else              
          LET ls_db_type = sadzi140_util_get_db_type() #20240521
          CASE ls_db_type
            WHEN cs_db_driver_type_kdb #20240521
              LET ls_line = ls_line,
                            "-- KDB drop primary key.","\n",
                            " alter table ",lo_table_constraint.KEY_TABLE," drop constraint ",lo_table_constraint.key_name,";","\n"
            WHEN cs_db_driver_type_dmg --20230804
              LET ls_Line = ls_Line,
                            "-- DM drop primary key [SEC1].","\n",
                            "alter table ",lo_table_constraint.KEY_TABLE," drop primary key cascade;","\n"
          OTHERWISE
            #Pure Oracle   
            LET ls_line = ls_line,
                          "-- Oracle drop primary key [SEC1].","\n",
                          "alter table ",lo_table_constraint.KEY_TABLE," drop primary key cascade drop index;","\n"
          END CASE 
          &endif
          &endif
        END IF    
        #檢查Index是否存在, 若存在則Drop 
        CALL sadzi140_db_check_table_index_exist(ls_owner,lo_table_constraint.KEY_NAME) RETURNING lb_index_exist
        IF lb_index_exist THEN
          CALL sadzi140_db_get_drop_index_procedure(lo_table_constraint.key_table,lo_table_constraint.KEY_NAME) RETURNING ls_drop_index #20191209
          LET ls_line = ls_line,
                        "-- sadzi140_db_gen_table_constraint()","\n", 
                        ls_drop_index,"\n"
        END IF                
        
      END IF
      &ifdef is_msv
      LET ls_line = ls_line,
                    "alter table ",lo_table_constraint.KEY_TABLE,
                    " add constraint ",lo_table_constraint.KEY_NAME,
                    " primary key (",lo_table_constraint.KEY_COLUMNS,")","\n","go","\n" #20191209
      &else
      &ifdef is_pgs
      LET ls_line = ls_line,
                    "alter table ",lo_table_constraint.KEY_TABLE,
                    " add constraint ",lo_table_constraint.KEY_NAME,
                    " primary key (",lo_table_constraint.KEY_COLUMNS,");","\n" #20200330
      &else
      LET ls_db_type = sadzi140_util_get_db_type() #20240521
      CASE ls_db_type
        WHEN cs_db_driver_type_kdb #20240521
          LET ls_line = ls_line,
                        "alter table ",lo_table_constraint.KEY_TABLE,
                        " add constraint ",lo_table_constraint.KEY_NAME,
                        " primary key (",lo_table_constraint.KEY_COLUMNS,");","\n" 
        WHEN cs_db_driver_type_dmg --20230804
          LET ls_line = ls_line,
                        "alter table ",lo_table_constraint.KEY_TABLE,
                        " add constraint ",lo_table_constraint.KEY_NAME,
                        " primary key (",lo_table_constraint.KEY_COLUMNS,") enable validate;","\n"
      OTHERWISE  
        LET ls_line = ls_line,
                      "alter table ",lo_table_constraint.KEY_TABLE,
                      " add constraint ",lo_table_constraint.KEY_NAME,
                      " primary key (",lo_table_constraint.KEY_COLUMNS,") enable validate;","\n"
      END CASE  
      &endif              
      &endif              
    END IF

    IF lo_table_constraint.KEY_TYPE = "U" THEN
      &ifdef is_msv
      LET ls_line = "alter table ",lo_table_constraint.KEY_TABLE,
                    " add constraint ",lo_table_constraint.KEY_NAME,
                    " unique (",lo_table_constraint.KEY_COLUMNS,")","\n","go","\n" #20191209
      &else
      &ifdef is_pgs
      LET ls_line = "alter table ",lo_table_constraint.KEY_TABLE,
                    " add constraint ",lo_table_constraint.KEY_NAME,
                    " unique (",lo_table_constraint.KEY_COLUMNS,");","\n" #20200330
      &else
      LET ls_db_type = sadzi140_util_get_db_type() #20240521
      CASE ls_db_type
        WHEN cs_db_driver_type_kdb #20240521
          LET ls_line = "alter table ",lo_table_constraint.KEY_TABLE,
                        " add constraint ",lo_table_constraint.KEY_NAME,
                        " unique (",lo_table_constraint.KEY_COLUMNS,");","\n"
        WHEN cs_db_driver_type_dmg #20230804
          LET ls_line = "alter table ",lo_table_constraint.KEY_TABLE,
                        " add constraint ",lo_table_constraint.KEY_NAME,
                        " unique (",lo_table_constraint.KEY_COLUMNS,") enable validate;","\n"
      OTHERWISE  
        LET ls_line = "alter table ",lo_table_constraint.KEY_TABLE,
                      " add constraint ",lo_table_constraint.KEY_NAME,
                      " unique (",lo_table_constraint.KEY_COLUMNS,") enable validate;","\n"
      END CASE  
      &endif              
      &endif              
    END IF

    IF lo_table_constraint.KEY_TYPE = "R" THEN
      IF (ls_enable_foreign_key = "Y") THEN 
        &ifdef is_msv
        LET ls_line = "alter table ",lo_table_constraint.KEY_TABLE,
                      " add constraint ",lo_table_constraint.KEY_NAME,
                      " foreign key (",lo_table_constraint.KEY_COLUMNS,")",
                      " references ",lo_table_constraint.FOREIGN_TABLE,
                      " (",lo_table_constraint.FOREIGN_COLUMNS,")","\n","go","\n" #20191209
        IF ls_enable_foreign_key <> "Y" THEN
          LET ls_line = ls_line,
                        "alter table ",lo_table_constraint.KEY_TABLE," nocheck constraint ",lo_table_constraint.KEY_NAME,"","\n","go","\n" #20191209
        END IF        
        &else
        &ifdef is_pgs
        LET ls_line = "alter table ",lo_table_constraint.KEY_TABLE,
                      " add constraint ",lo_table_constraint.KEY_NAME,
                      " foreign key (",lo_table_constraint.KEY_COLUMNS,")",
                      " references ",lo_table_constraint.FOREIGN_TABLE,
                      " (",lo_table_constraint.FOREIGN_COLUMNS,") ",";","\n" #20200330
        &else
        LET ls_db_type = sadzi140_util_get_db_type() #20240521
        CASE ls_db_type
          WHEN cs_db_driver_type_kdb #20240521
            LET ls_line = "alter table ",lo_table_constraint.KEY_TABLE,
                          " add constraint ",lo_table_constraint.KEY_NAME,
                          " foreign key (",lo_table_constraint.KEY_COLUMNS,")",
                          " references ",lo_table_constraint.FOREIGN_TABLE,
                          " (",lo_table_constraint.FOREIGN_COLUMNS,") ",";","\n"
          WHEN cs_db_driver_type_dmg #20230804
            LET ls_line = "alter table ",lo_table_constraint.KEY_TABLE,
                          " add constraint ",lo_table_constraint.KEY_NAME,
                          " foreign key (",lo_table_constraint.KEY_COLUMNS,")",
                          " references ",lo_table_constraint.FOREIGN_TABLE,
                          " (",lo_table_constraint.FOREIGN_COLUMNS,") ",ls_foreign_key_state,";","\n"
        OTHERWISE   
          LET ls_line = "alter table ",lo_table_constraint.KEY_TABLE,
                        " add constraint ",lo_table_constraint.KEY_NAME,
                        " foreign key (",lo_table_constraint.KEY_COLUMNS,")",
                        " references ",lo_table_constraint.FOREIGN_TABLE,
                        " (",lo_table_constraint.FOREIGN_COLUMNS,") ",ls_foreign_key_state,";","\n"
        END CASE  
        &endif              
        &endif              
      END IF;      
    END IF
    
    CALL lo_line_str_buf.append(ls_line)

  END FOREACH
  CLOSE lcur_GenTableConstraint
  
  FREE lpre_GenTableConstraint
  FREE lcur_GenTableConstraint  

  LET ls_return = lo_line_str_buf.toString()

  RETURN ls_return
                  
END FUNCTION

FUNCTION sadzi140_db_get_drop_index_procedure(p_table_name,p_index_name) #20191209
DEFINE
  p_table_name STRING, #20191209
  p_index_name STRING
DEFINE
  ls_table_name STRING, #20191209
  ls_index_name STRING,
  ls_procedure  STRING,
  ls_db_type    STRING #20240521
DEFINE
  ls_return STRING  

  LET ls_table_name = p_table_name #20191209
  LET ls_index_name = p_index_name

  &ifdef is_msv
  LET ls_procedure = "drop index if exists ",ls_index_name," on ",ls_table_name,"\n","go","\n" #20191209
  &else
  &ifdef is_pgs
  LET ls_procedure = "drop index if exists ",ls_index_name,";","\n" #20200330
  &else
  LET ls_db_type = sadzi140_util_get_db_type() #20240521
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb #20240521
      LET ls_procedure = "drop index if exists ",ls_index_name,";","\n"
    WHEN cs_db_driver_type_dmg #20230804
      LET ls_procedure = "-- Drop index inline procedure.                                                      ","\n",
                         "SET SERVEROUTPUT ON                                                                  ","\n",
                         "DECLARE                                                                              ","\n",
                         "  INDEX_NOT_EXISTS EXCEPTION;                                                        ","\n",
                         "  PRAGMA EXCEPTION_INIT(INDEX_NOT_EXISTS, -1418);                                    ","\n",
                         "BEGIN                                                                                ","\n",
                         "  DBMS_OUTPUT.ENABLE(100000);                                                        ","\n",
                         "  EXECUTE IMMEDIATE ('DROP INDEX ",ls_index_name.toUpperCase(),"');                  ","\n",
                         "EXCEPTION                                                                            ","\n",
                         "  WHEN INDEX_NOT_EXISTS THEN                                                         ","\n",
                         "    DBMS_OUTPUT.PUT_LINE('",cs_warning_tag," INDEX ",ls_index_name," NOT EXISTS.');  ","\n",
                         "END;                                                                                 ","\n",
                         "/                                                                                    "
  OTHERWISE   
    LET ls_procedure = "-- Drop index inline procedure.                                                      ","\n",
                       "SET SERVEROUTPUT ON                                                                  ","\n",
                       "DECLARE                                                                              ","\n",
                       "  INDEX_NOT_EXISTS EXCEPTION;                                                        ","\n",
                       "  PRAGMA EXCEPTION_INIT(INDEX_NOT_EXISTS, -1418);                                    ","\n",
                       "BEGIN                                                                                ","\n",
                       "  DBMS_OUTPUT.ENABLE(100000);                                                        ","\n",
                       "  EXECUTE IMMEDIATE ('DROP INDEX ",ls_index_name.toUpperCase(),"');                  ","\n",
                       "EXCEPTION                                                                            ","\n",
                       "  WHEN INDEX_NOT_EXISTS THEN                                                         ","\n",
                       "    DBMS_OUTPUT.PUT_LINE('",cs_warning_tag," INDEX ",ls_index_name," NOT EXISTS.');  ","\n",
                       "END;                                                                                 ","\n",
                       "/                                                                                    "
  END CASE  
  &endif                   
  &endif                   

  LET ls_return =  ls_procedure

  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_gen_table_index(p_owner,p_table_name,p_include_drop)
DEFINE
  p_owner        STRING,
  p_table_name   STRING,
  p_include_drop BOOLEAN
DEFINE
  ls_owner        STRING,
  ls_tablename    STRING,
  lb_include_drop BOOLEAN,
  ls_line         STRING,
  ls_sql          STRING,
  ls_unique       STRING,
  ls_drop_index   STRING, #20171120
  #lo_line_str_buf base.StringBuffer,
  lb_index_exist  BOOLEAN,
  lo_table_index RECORD
                   index_table   VARCHAR(50),
                   index_name    VARCHAR(50),
                   index_type    VARCHAR(10),
                   index_columns VARCHAR(150)
                 END RECORD
DEFINE
  ls_return STRING               

  LET ls_owner        = p_owner
  LET ls_tablename    = p_table_name
  LET lb_include_drop = p_include_drop
  
  #LET lo_line_str_buf = base.StringBuffer.create()
  
  LET ls_unique = "";

  LET ls_sql = "select ec.dzec001 index_table,                     ",  
               "       ec.dzec002 index_name,                      ",
               "       ec.dzec003 index_type,                      ",
               "       ec.dzec004 index_columns                    ",
               "  from dzec_t ec                                   ",
               " where ec.dzec001 = '",ls_tablename,"'             ",
               "   and coalesce(ec.dzecstus,'Y') = 'Y'             ", #20191209 #20170912
               "   and not exists (                                ", #20240529 begin
               "                    select 1                       ", 
               "                      from dzed_t ed               ",
               "                     where ed.dzed001 = ec.dzec001 ",
               "                       and ed.dzed002 = ec.dzec002 ",
               "                  )                                "  #20240529 end

  PREPARE lpre_gen_table_index FROM ls_sql
  DECLARE lcur_gen_table_index CURSOR FOR lpre_gen_table_index

  OPEN lcur_gen_table_index
  FOREACH lcur_gen_table_index INTO lo_table_index.*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    IF NVL(lo_table_index.INDEX_TYPE,cs_null_value) = "U" THEN
      LET ls_unique = "unique";
    ELSE
      LET ls_unique = "";
    END IF

    IF lb_include_drop THEN
      CALL sadzi140_db_check_table_index_exist(ls_owner,lo_table_index.INDEX_NAME) RETURNING lb_index_exist
      IF lb_index_exist THEN
        CALL sadzi140_db_get_drop_index_procedure(lo_table_index.index_table,lo_table_index.INDEX_NAME) RETURNING ls_drop_index  #20191209 #20171120
        LET ls_line = ls_line,
                      "-- sadzi140_db_gen_table_index()","\n",
                      ls_drop_index,"\n" #20171120
                      --"drop index ",lo_table_index.INDEX_NAME,";","\n"
      END IF                
    END IF

    &ifdef is_msv
    LET ls_line = ls_line,
                  "create ",ls_unique," index ",lo_table_index.INDEX_NAME,
                  " on ",lo_table_index.INDEX_TABLE,
                  " (",lo_table_index.INDEX_COLUMNS,")","\n","go","\n"
    &else
    --Oracle & PostgreSQL 共用
    LET ls_line = ls_line,
                  "create ",ls_unique," index ",lo_table_index.INDEX_NAME,
                  " on ",lo_table_index.INDEX_TABLE,
                  " (",lo_table_index.INDEX_COLUMNS,");","\n"
    &endif    
                  
    #CALL lo_line_str_buf.append(ls_line)

  END FOREACH
  CLOSE lcur_gen_table_index
  
  FREE lpre_gen_table_index
  FREE lcur_gen_table_index  

  LET ls_return = ls_line #lo_line_str_buf.toString()

  RETURN ls_return
                  
END FUNCTION

FUNCTION sadzi140_db_gen_table_grant(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_line          STRING,
  ls_sql           STRING,
  ls_table_name    STRING,
  ls_union         STRING,
  lo_line_str_buf  base.StringBuffer,
  lo_table_grant RECORD
                   privileges_name   VARCHAR(30),
                   privileges_user   VARCHAR(30)
                 END RECORD
DEFINE
  ls_Return STRING               

  LET ls_table_name   = p_table_name
  LET lo_line_str_buf = base.StringBuffer.create()

  LET ls_union = ""
  IF sadzi140_util_get_pgs_branch_type() = cs_pgs_sub_mogdb THEN --20230804
    LET ls_union = "union                          ",
                   "select 'all'  privileges_name, ",
                   "        user  privileges_user  ",
                   "  from dual                    " 
  END IF
  
  LET ls_sql = "select bj.dzej005 privileges_name, ",
               "       bj.dzej006 privileges_user  ", 
               "  from dzej_t bj                   ", 
               " where bj.dzej001 = 'TABLE_GRANTS' ",
               ls_union, 
               " order by 1,2                      "

  PREPARE lpre_gen_table_grant FROM ls_sql
  DECLARE lcur_gen_table_grant CURSOR FOR lpre_gen_table_grant

  OPEN lcur_gen_table_grant
  FOREACH lcur_gen_table_grant INTO lo_table_grant.*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET ls_line = "grant ",lo_table_grant.privileges_name,
                  " on ",ls_table_name,
                  " to ",lo_table_grant.privileges_user,";","\n"
    CALL lo_line_str_buf.append(ls_line)

  END FOREACH
  CLOSE lcur_gen_table_grant
  
  FREE lpre_gen_table_grant
  FREE lcur_gen_table_grant  

  &ifdef is_msv
  LET ls_Return = ""
  &else
  LET ls_Return = lo_line_str_buf.toString()
  &endif

  RETURN ls_Return
                  
END FUNCTION

FUNCTION sadzi140_db_exec_SQL(p_sql,p_batch_mode)
DEFINE
  p_sql STRING,
  p_batch_mode BOOLEAN
DEFINE
  ls_sql         STRING,
  lb_batch_mode  BOOLEAN,
  ls_replace_arg STRING

  LET ls_sql = p_sql
  LET lb_batch_mode = p_batch_mode

  BEGIN WORK

  TRY
    PREPARE lpre_exec_sql FROM ls_sql
    EXECUTE lpre_exec_sql
    COMMIT WORK
  CATCH
    IF lb_batch_mode THEN
      DISPLAY cs_error_tag,"Execute SQL ",ls_sql," not success." 
    ELSE
      LET ls_replace_arg = ls_sql,"|"
      CALL sadzp000_msg_show_error(NULL, 'adz-00086', ls_replace_arg, 1) 
    END IF   
    ROLLBACK WORK
  END TRY  
  
END FUNCTION

FUNCTION sadzi140_db_exec_SQL_no_commit(p_sql,p_batch_mode)
DEFINE
  p_sql STRING,
  p_batch_mode BOOLEAN
DEFINE
  ls_sql         STRING,
  lb_batch_mode  BOOLEAN,
  ls_replace_arg STRING
DEFINE
  lb_return BOOLEAN 
  
  LET ls_sql = p_sql
  LET lb_batch_mode = p_batch_mode

  LET lb_return = TRUE
  
  TRY
    PREPARE lpre_exec_SQL_no_commit FROM ls_sql
    EXECUTE lpre_exec_SQL_no_commit
  CATCH
    LET lb_return = FALSE
    IF lb_batch_mode THEN
      DISPLAY cs_error_tag,"Execute SQL ",ls_sql," not success." 
    ELSE
      LET ls_replace_arg = ls_sql,"|"
      CALL sadzp000_msg_show_error(NULL, 'adz-00086', ls_replace_arg, 1)
    END IF   
  END TRY  
  
END FUNCTION

FUNCTION sadzi140_db_run_SQL(p_sql)
DEFINE
  p_sql STRING
DEFINE
  ls_sql STRING
DEFINE
  lb_return BOOLEAN 
  
  LET ls_sql = p_sql

  LET lb_return = TRUE
  
  TRY
    PREPARE lpre_run_SQL_no_commit FROM ls_sql
    EXECUTE lpre_run_SQL_no_commit
  CATCH
    LET lb_return = FALSE
  END TRY  

  FREE lpre_run_SQL_no_commit

  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_db_get_fgl_connect_string(p_db_name)
DEFINE
  p_db_name STRING
DEFINE
  lo_db_connstr T_DB_CONNSTR,
  ls_db_spec    STRING,
  ls_db_str     STRING

  LET ls_db_spec = p_db_name
  LET ls_db_str  = "dbi.database.",ls_db_spec,"."
  
  LET lo_db_connstr.db_source   = FGL_GETRESOURCE(ls_db_str||"source")
  LET lo_db_connstr.db_username = FGL_GETRESOURCE(ls_db_str||"username")
  LET lo_db_connstr.db_password = FGL_GETRESOURCE(ls_db_str||"password")
  LET lo_db_connstr.db_schema   = FGL_GETRESOURCE(ls_db_str||"schema")
  LET lo_db_connstr.db_sid      = NVL(FGL_GETENV("ORACLE_SID"),FGL_GETENV("DM_DATABASE")) #20230804
  
  RETURN lo_db_connstr.*
  
END FUNCTION

FUNCTION sadzi140_db_get_db_connect_string(p_db_name)
DEFINE
  p_db_name STRING
DEFINE
  ls_db_name      STRING,
  ls_db_sid       STRING,
  ls_privacy_word STRING,
  lb_success      BOOLEAN,
  lo_db_connstr   T_DB_CONNSTR,
  ls_db_source    STRING, #20191209
  ls_db_ip_port   STRING, #20200330
  ls_db_type      STRING, #20240521
  ls_sql          STRING
DEFINE
  lo_return T_DB_CONNSTR  
  
  LET ls_db_name = p_db_name

  &ifdef is_msv
  LET ls_db_sid = ""
  LET ls_db_source = "'",FGL_GETENV("ZONE"),"_","'+da.gzda001"
  &else
  &ifdef is_pgs
  LET ls_db_sid = FGL_GETENV("PGDATABASE")
  LET ls_db_source = "'",ls_db_sid,"'"
  &else
  LET ls_db_type = sadzi140_util_get_db_type() #20240521
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb #20240521
      LET ls_db_sid = FGL_GETENV("KINGBASE_DATABASE")
    WHEN cs_db_driver_type_dmg --20230804
      LET ls_db_sid = FGL_GETENV("DM_DATABASE")
  OTHERWISE
    LET ls_db_sid = FGL_GETENV("ORACLE_SID")
  END CASE 
  LET ls_db_source = "'",ls_db_sid,"'"
  &endif
  &endif
      
  LET ls_sql = "select ",ls_db_source," db_source,       ", #20191209 --"select '",ls_db_sid,"'  db_source,       ",
               "       da.gzda003       db_username,     ",
               "       da.gzda004       db_password,     ",
               "       da.gzda001       db_schema,       ",
               "       '",ls_db_sid,"'  db_sid,          ",
               "       ''               db_sql_filename, ",
               "       ''               db_version,      ",
               "       da.gzda002       db_host_ip,      ", #20191209 
               "       da.gzda002       db_host_port     ", #20200330
               "  from gzda_t da                         ",
               " where 1=1                               ",
               "   and da.gzda001  = '",ls_db_name,"'    ",
               --"   and da.gzda005  = 'Y'                 ", 
               "   and da.gzdastus = 'Y'                 "

  PREPARE lpre_get_db_connect_string FROM ls_sql
  DECLARE lcur_get_db_connect_string CURSOR FOR lpre_get_db_connect_string

  INITIALIZE lo_db_connstr TO NULL
  
  OPEN lcur_get_db_connect_string
  FETCH lcur_get_db_connect_string INTO lo_db_connstr.*
  CLOSE lcur_get_db_connect_string

  #20200330 begin
  LET ls_db_ip_port = lo_db_connstr.db_host_ip
  IF ls_db_ip_port.getIndexOf(":",1) > 0 THEN
    LET lo_db_connstr.db_host_ip   = ls_db_ip_port.subString(1,ls_db_ip_port.getIndexOf(":",1)-1) 
    LET lo_db_connstr.db_host_port = ls_db_ip_port.subString(ls_db_ip_port.getIndexOf(":",1)+1,ls_db_ip_port.getLength()) 
  END IF  
  #20200330 end

  &ifndef DEBUG
  IF (NVL(lo_db_connstr.db_username,cs_null_value) = cs_master_db) THEN
    CALL sadzi140_db_get_privacy_word() RETURNING lb_success,ls_privacy_word
    IF lb_success THEN 
      LET lo_db_connstr.db_password = cl_hashkey_base65_anti(ls_privacy_word)
    ELSE 
      LET lo_db_connstr.db_password = cl_hashkey_base65_anti(lo_db_connstr.db_password)
    END IF 
  ELSE 
    IF lo_db_connstr.db_username IS NOT NULL THEN 
      LET lo_db_connstr.db_password = cl_hashkey_base65_anti(lo_db_connstr.db_password)
    END IF  
  END IF 
  &endif
  
  FREE lpre_get_db_connect_string
  FREE lcur_get_db_connect_string  

  LET lo_return.* = lo_db_connstr.*
  
  RETURN lo_return.*
  
END FUNCTION

FUNCTION sadzi140_db_get_all_db_connect_string()
DEFINE
  ls_db_sid       STRING,
  li_rec_cnt      INTEGER,
  lo_db_connstr   DYNAMIC ARRAY OF T_DB_CONNSTR,
  ls_privacy_word STRING,
  lb_success      BOOLEAN,
  ls_db_source    STRING, #20191209
  ls_db_ip_port   STRING, #20200330
  ls_db_type      STRING, #20240521
  ls_sql          STRING
DEFINE
  lo_return DYNAMIC ARRAY OF T_DB_CONNSTR  

  &ifdef is_msv
  LET ls_db_sid = ""
  LET ls_db_source = "'",FGL_GETENV("ZONE"),"_","'+da.gzda001"
  &else
  &ifdef is_pgs
  LET ls_db_sid = FGL_GETENV("PGDATABASE")
  LET ls_db_source = "'",ls_db_sid,"'"
  &else
  LET ls_db_type = sadzi140_util_get_db_type() #20240521
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb #20240521
      LET ls_db_sid = FGL_GETENV("KINGBASE_DATABASE")
    WHEN cs_db_driver_type_dmg --20230804
      LET ls_db_sid = FGL_GETENV("DM_DATABASE")
  OTHERWISE
    LET ls_db_sid = FGL_GETENV("ORACLE_SID")
  END CASE 
  LET ls_db_source = "'",ls_db_sid,"'"
  &endif
  &endif
      
  LET ls_sql = "select ",ls_db_source," db_source,       ", #20191209 --"select '",ls_db_sid,"'  db_source,       ",
               "       da.gzda003       db_username,     ",
               "       da.gzda004       db_password,     ",
               "       da.gzda001       db_schema,       ",
               "       '",ls_db_sid,"'  db_sid,          ",
               "       ''               db_sql_filename, ",
               "       ''               db_version,      ",
               "       da.gzda002       db_host_ip,      ", #20191209 
               "       da.gzda002       db_host_port     ", #20200330
               "  from gzda_t da                         ",
               " where 1=1                               ",  
               "   and da.gzda005  = 'Y'                 ", 
               "   and da.gzdastus = 'Y'                 "

  PREPARE lpre_get_all_db_connect_string FROM ls_sql
  DECLARE lcur_get_all_db_connect_string CURSOR FOR lpre_get_all_db_connect_string

  INITIALIZE lo_db_connstr TO NULL

  LET li_rec_cnt = 1
  OPEN lcur_get_all_db_connect_string
  FOREACH lcur_get_all_db_connect_string INTO lo_db_connstr[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    #20200330 begin
    LET ls_db_ip_port = lo_db_connstr[li_rec_cnt].db_host_ip
    IF ls_db_ip_port.getIndexOf(":",1) > 0 THEN
      LET lo_db_connstr[li_rec_cnt].db_host_ip   = ls_db_ip_port.subString(1,ls_db_ip_port.getIndexOf(":",1)-1) 
      LET lo_db_connstr[li_rec_cnt].db_host_port = ls_db_ip_port.subString(ls_db_ip_port.getIndexOf(":",1)+1,ls_db_ip_port.getLength()) 
    END IF  
    #20200330 end
    
    &ifndef DEBUG
    IF (lo_db_connstr[li_rec_cnt].db_username = cs_master_db) THEN
      CALL sadzi140_db_get_privacy_word() RETURNING lb_success,ls_privacy_word
      IF lb_success THEN 
        LET lo_db_connstr[li_rec_cnt].db_password = cl_hashkey_base65_anti(ls_privacy_word)
      ELSE 
        LET lo_db_connstr[li_rec_cnt].db_password = cl_hashkey_base65_anti(lo_db_connstr[li_rec_cnt].db_password)
      END IF 
    ELSE 
      LET lo_db_connstr[li_rec_cnt].db_password = cl_hashkey_base65_anti(lo_db_connstr[li_rec_cnt].db_password)
    END IF 
    &endif
    
    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_all_db_connect_string

  CALL lo_db_connstr.deleteElement(li_rec_cnt)
  
  FREE lpre_get_all_db_connect_string
  FREE lcur_get_all_db_connect_string

  LET lo_return = lo_db_connstr
  
  RETURN lo_return
  
END FUNCTION

FUNCTION sadzi140_db_get_valid_schema_list()
DEFINE
  li_rec_cnt      INTEGER,
  lo_schema_list  DYNAMIC ARRAY OF T_SCHEMA_LIST, 
  ls_sql          STRING
DEFINE
  lo_return DYNAMIC ARRAY OF T_SCHEMA_LIST  
  
  LET ls_sql = "select da.gzda001                        ", #20191209 gzda003 change to gzda001
               "  from gzda_t da                         ", 
               " where 1=1                               ", 
               "   and da.gzda001 not in ('ds','dsdemo') ", #20191209 gzda003 change to gzda001
               "   and da.gzda005  = 'Y'                 ", 
               "   and da.gzdastus = 'Y'                 ",
               " order by da.gzda001                     "  #20191209 gzda003 change to gzda001

  PREPARE lpre_get_valid_schema_list FROM ls_sql
  DECLARE lcur_get_valid_schema_list CURSOR FOR lpre_get_valid_schema_list

  INITIALIZE lo_schema_list TO NULL

  LET li_rec_cnt = 1
  
  OPEN lcur_get_valid_schema_list
  FOREACH lcur_get_valid_schema_list INTO lo_schema_list[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_valid_schema_list
  CALL lo_schema_list.deleteElement(li_rec_cnt)
  
  FREE lpre_get_valid_schema_list
  FREE lcur_get_valid_schema_list

  LET lo_return = lo_schema_list
  
  RETURN lo_return
  
END FUNCTION

FUNCTION sadzi140_db_get_table_type_by_dsdemo(p_table_name,p_schema_name)
DEFINE 
  p_table_name   STRING,
  p_schema_name  STRING
DEFINE
  ls_table_name  STRING,
  ls_schema_name STRING,
  li_rec_cnt     INTEGER,
  lo_table_type  T_TABLE_TYPE, 
  ls_sql         STRING
DEFINE
  lo_return T_TABLE_TYPE  

  LET ls_table_name  = p_table_name.toLowerCase()
  LET ls_schema_name = p_schema_name.toLowerCase()

  #從 DSDEMO 抓參考資料
  &ifdef is_msv
  LET ls_sql = "select eu.dzeu001,da.gzda001 dzeu002,eu.dzeu003,                  ",
               "       eu.dzeu004,cast(eun.dzeu005 as varchar) dzeu005            ",
               "  from dzeu_t eu                                                  ",
               "  left outer join (                                               ",
               "                    select eu3.dzeu001,max(eu3.dzeu005)+1 dzeu005 ",
               "                      from dzeu_t eu3                             ",
               "                     group by eu3.dzeu001                         ",
               "                  ) eun on eun.dzeu001 = eu.dzeu001               ",
               "  left outer join gzda_t da on da.gzda001 = '",ls_schema_name,"'  ", #20191209 gzda003 change to gzda001  
               "                           and da.gzda005 = 'Y'                   ",
               " where eu.dzeu001 = '",ls_table_name,"'                           ",
               "   and eu.dzeu002 = (                                             ",
               "                      select da.gzda001                           ",
               "                        from gzda_t da                            ",
               "                       where da.gzda001 = eu.dzeu002              ",
               "                         and da.gzda001 = '",cs_slave_user,"'     ", #20191209 gzda003 change to gzda001
               "                    )                                             ",
               "   and not exists (                                               ",
               "                    select 1                                      ",
               "                      from dzeu_t eu2                             ",
               "                     where eu2.dzeu001 = eu.dzeu001               ",
               "                       and eu2.dzeu002 = '",ls_schema_name,"'     ",
               "                  )                                               "
  &else
  &ifdef is_pgs
  LET ls_sql = "select eu.dzeu001,da.gzda001 dzeu002,eu.dzeu003,                  ",
               "       eu.dzeu004,eun.dzeu005::varchar as dzeu005                 ",
               "  from dzeu_t eu                                                  ",
               "  left outer join (                                               ",
               "                    select eu3.dzeu001,max(eu3.dzeu005)+1 dzeu005 ",
               "                      from dzeu_t eu3                             ",
               "                     group by eu3.dzeu001                         ",
               "                  ) eun on eun.dzeu001 = eu.dzeu001               ",
               "  left outer join gzda_t da on da.gzda001 = '",ls_schema_name,"'  ", #20191209 gzda003 change to gzda001  
               "                           and da.gzda005 = 'Y'                   ",
               " where eu.dzeu001 = '",ls_table_name,"'                           ",
               "   and eu.dzeu002 = (                                             ",
               "                      select da.gzda001                           ",
               "                        from gzda_t da                            ",
               "                       where da.gzda001 = eu.dzeu002              ",
               "                         and da.gzda001 = '",cs_slave_user,"'     ", #20191209 gzda003 change to gzda001
               "                    )                                             ",
               "   and not exists (                                               ",
               "                    select 1                                      ",
               "                      from dzeu_t eu2                             ",
               "                     where eu2.dzeu001 = eu.dzeu001               ",
               "                       and eu2.dzeu002 = '",ls_schema_name,"'     ",
               "                  )                                               "
  &else
  LET ls_sql = "select eu.dzeu001,da.gzda001 dzeu002,eu.dzeu003,                  ",
               "       eu.dzeu004,to_char(eun.dzeu005) dzeu005                    ",
               "  from dzeu_t eu                                                  ",
               "  left outer join (                                               ",
               "                    select eu3.dzeu001,max(eu3.dzeu005)+1 dzeu005 ",
               "                      from dzeu_t eu3                             ",
               "                     group by eu3.dzeu001                         ",
               "                  ) eun on eun.dzeu001 = eu.dzeu001               ",
               "  left outer join gzda_t da on da.gzda001 = '",ls_schema_name,"'  ", #20191209 gzda003 change to gzda001  
               "                           and da.gzda005 = 'Y'                   ",
               " where eu.dzeu001 = '",ls_table_name,"'                           ",
               "   and eu.dzeu002 = (                                             ",
               "                      select da.gzda001                           ",
               "                        from gzda_t da                            ",
               "                       where da.gzda001 = eu.dzeu002              ",
               "                         and da.gzda001 = '",cs_slave_user,"'     ", #20191209 gzda003 change to gzda001
               "                    )                                             ",
               "   and not exists (                                               ",
               "                    select 1                                      ",
               "                      from dzeu_t eu2                             ",
               "                     where eu2.dzeu001 = eu.dzeu001               ",
               "                       and eu2.dzeu002 = '",ls_schema_name,"'     ",
               "                  )                                               "
  &endif
  &endif
  
  PREPARE lpre_get_table_type_by_dsdemo FROM ls_sql
  DECLARE lcur_get_table_type_by_dsdemo CURSOR FOR lpre_get_table_type_by_dsdemo

  INITIALIZE lo_table_type TO NULL

  LET li_rec_cnt = 1
  
  OPEN lcur_get_table_type_by_dsdemo
  FETCH lcur_get_table_type_by_dsdemo INTO lo_table_type.*
  CLOSE lcur_get_table_type_by_dsdemo
  
  FREE lpre_get_table_type_by_dsdemo
  FREE lcur_get_table_type_by_dsdemo

  LET lo_return.* = lo_table_type.*
  
  RETURN lo_return.*
  
END FUNCTION

FUNCTION sadzi140_db_get_privacy_word()
DEFINE
  ls_file_name    STRING,
  lo_channel      base.Channel,
  ls_privacy_word STRING,
  lb_file_exist   BOOLEAN
DEFINE
  ls_result STRING,
  lb_result BOOLEAN
  
  LET ls_file_name = os.Path.join(os.Path.join(FGL_GETENV("FBIN"),"4gl"),"gendbs.inc")

  IF os.Path.exists(ls_file_name) THEN
    LET lb_file_exist = TRUE
    LET lo_channel = base.Channel.create()
    CALL lo_channel.openFile(ls_file_name.trim(),"r")
    WHILE lo_channel.read(ls_privacy_word)
      EXIT WHILE
    END WHILE
    CALL lo_channel.close()
  ELSE
    LET lb_file_exist = FALSE
    DISPLAY cs_error_tag," Get privacy file error !!"
  END IF

  LET lb_result = lb_file_exist
  LET ls_result = ls_privacy_word

  RETURN lb_result,ls_result
  
END FUNCTION      

FUNCTION sadzi140_db_get_os_type()
DEFINE
  ls_separator STRING
DEFINE
  ls_return STRING

  LET ls_separator = os.Path.separator()

  CASE ls_separator
    WHEN "\\"
      LET ls_return = cs_os_type_win
    WHEN "/"
      LET ls_return = cs_os_type_unix
  OTHERWISE
    LET ls_return = cs_os_type_unix
  END CASE   

  RETURN ls_return

END FUNCTION

FUNCTION sadzi140_db_sqlplus(p_db_connstr)
DEFINE
  p_db_connstr T_DB_CONNSTR
DEFINE
  ls_sqlplus_str STRING,
  ls_message     STRING,
  ls_log_file    STRING,
  ls_file_list   STRING,
  ls_return      STRING,
  ls_disql_bef   STRING,
  ls_disql_cmd   STRING,
  ls_disql_aft   STRING,
  ls_db_type     STRING, #20240521
  lo_db_connstr  T_DB_CONNSTR

  LET lo_db_connstr.* = p_db_connstr.*

  &ifdef is_msv
    &ifdef DEBUG
    LET lo_db_connstr.db_password = "topprd"
    &endif
    LET ls_sqlplus_str = "sqlcmd"," -S ",lo_DB_CONNSTR.db_host_ip," -d ",lo_DB_CONNSTR.db_source," -U ",lo_DB_CONNSTR.db_username," -P ",lo_DB_CONNSTR.db_password," -i ",lo_DB_CONNSTR.db_sql_filename
    IF sadzi140_db_get_os_type() = cs_os_type_win THEN 
      LET ls_sqlplus_str = "chcp 65001&",ls_sqlplus_str
    END IF  
  &else
  &ifdef is_pgs
    &ifdef DEBUG
    LET lo_db_connstr.db_password = lo_DB_CONNSTR.db_username
    &endif
    IF sadzi140_util_get_pgs_branch_type() = cs_pgs_sub_pgs THEN --20230606 
      LET ls_sqlplus_str = "export PGPASSWORD=",lo_DB_CONNSTR.db_password,";psql -U ",lo_DB_CONNSTR.db_username," -f ",lo_DB_CONNSTR.db_sql_filename," -t"
    ELSE
      LET ls_sqlplus_str = "gsql -U ",lo_DB_CONNSTR.db_username," -W ",lo_DB_CONNSTR.db_password," -f ",lo_DB_CONNSTR.db_sql_filename," -t"
    END IF  
  &else
    LET ls_db_type = sadzi140_util_get_db_type()
    CASE ls_db_type
      WHEN cs_db_driver_type_kdb --20240521
        #LET ls_sqlplus_str = "ksql 'user=",lo_DB_CONNSTR.db_username," password=",lo_DB_CONNSTR.db_password,"' -c '\\set SQLTERM /' -f '",lo_DB_CONNSTR.db_sql_filename,"' -t"
        LET ls_sqlplus_str = "ksql 'user=",lo_DB_CONNSTR.db_username," password=",lo_DB_CONNSTR.db_password,"' -f '",lo_DB_CONNSTR.db_sql_filename,"' -t"
      WHEN cs_db_driver_type_dmg --20230804
        LET ls_disql_bef = "export DM_SVC_PATH_ORG=$DM_SVC_PATH;mkdir -p ~/.dm_svc;cp -f $DM_SVC_PATH/dm_svc.conf ~/.dm_svc;sed -i 's/LANGUAGE=.*/LANGUAGE='\\(en\\)'/' ~/.dm_svc/dm_svc.conf;export DM_SVC_PATH=~/.dm_svc;"
        LET ls_disql_cmd = "disql ",lo_DB_CONNSTR.db_username,"/",lo_DB_CONNSTR.db_password,"@",lo_DB_CONNSTR.db_source," \\`",lo_DB_CONNSTR.db_sql_filename," -c 'SET ECHO OFF';"
        LET ls_disql_aft = "export DM_SVC_PATH=$DM_SVC_PATH_ORG;unset DM_SVC_PATH_ORG;" -- "rm -f ~/.dm_svc/dm_svc.conf;"
        LET ls_sqlplus_str = ls_disql_bef,ls_disql_cmd,ls_disql_aft
        #DISPLAY cs_command_tag,ls_sqlplus_str  
    OTHERWISE
      LET ls_sqlplus_str = "sqlplus ",lo_DB_CONNSTR.db_username,"/",lo_DB_CONNSTR.db_password,"@",lo_DB_CONNSTR.db_source," @",lo_DB_CONNSTR.db_sql_filename
    END CASE
  &endif
  &endif
  #DISPLAY cs_command_tag,ls_sqlplus_str
  
  CALL sadzi140_db_run_and_catch_log(ls_sqlplus_str) RETURNING ls_message
  #RUN ls_sqlplus_str
  CALL sadzi140_db_gen_log_file(ls_Message) RETURNING ls_log_file
  #DISPLAY "Log File : ",ls_log_file

  LET ls_file_list = ls_file_list,"SQL File : ",lo_DB_CONNSTR.db_sql_filename,"\n"
  LET ls_file_list = ls_file_list,"Log File : ",ls_log_file,"\n"
  LET ls_file_list = ls_file_list,ls_message
  
  LET ls_return = ls_file_list
  
  RETURN ls_return

END FUNCTION

FUNCTION sadzi140_db_run_and_catch_log(p_command)
DEFINE 
  p_command   STRING  #要執行的指令
DEFINE  
  ls_command    STRING, #要執行的指令
  ls_return     STRING,
  ls_cmd_line   STRING,
  ls_msg_change STRING, #20191209
  ls_message    STRING,
  li_log_line   INTEGER,
  lch_read      base.Channel   

  LET ls_command = p_command   
  
  LET ls_message  = ""
  LET li_log_line = 0

  LET lch_read  = base.Channel.create()
  CALL lch_read.setDelimiter("")

  &ifdef is_pgs
  LET ls_command = ls_command||" 2>&1"
  &endif

  #DISPLAY cs_command_tag,ls_command
  
  CALL lch_read.openPipe(ls_command, "r")
  WHILE TRUE
    LET ls_cmd_line = lch_read.readLine()
    &ifdef is_msv
    LET ls_msg_change = ls_cmd_line.toUpperCase()
    IF ls_msg_change.getIndexOf("MSG",1) > 0 THEN
      LET ls_cmd_line = "Error : ",ls_cmd_line 
    END IF
    &endif
    IF lch_read.isEof() THEN 
      EXIT WHILE 
    END IF

    LET li_log_line = li_log_line + 1 
    LET ls_message = ls_message, 
                     "[",(li_log_line USING "&&&&"),"] ",ls_cmd_line, "\n"
    LET ls_cmd_line = ls_cmd_line.toUpperCase()

  END WHILE
  CALL lch_read.close()

  LET ls_return = ls_message
  
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_gen_log_file(p_log)
DEFINE
  p_log      STRING
DEFINE
  ls_log           STRING, 
  ls_GUID   STRING,
  ls_temp_dir      STRING,
  lchannel_write   base.Channel,
  ls_log_filename  STRING,
  ls_return        STRING,
  ls_separator     STRING
  
  LET ls_log      = p_log

  LET ls_separator = os.Path.separator()
  
  LET ls_temp_dir = FGL_GETENV("TEMPDIR")
  CALL sadzi140_db_get_GUID() RETURNING ls_GUID 
  LET ls_log_filename = ls_temp_dir,ls_separator,"sadzi140_db_",ls_GUID,".log"
  
  LET lchannel_write = base.Channel.create()
  CALL lchannel_write.setDelimiter("")

  LET ls_log = ls_log
  
  #開檔寫入
  TRY
    CALL lchannel_write.openFile(ls_log_filename, "w" )
    CALL lchannel_write.write(ls_log)
    #關檔
    CALL lchannel_write.close()
  CATCH
    DISPLAY "Generate log Error : ",ls_log_filename
  END TRY  

  LET ls_return = ls_log_filename

  RETURN ls_return

END FUNCTION

FUNCTION sadzi140_db_gen_db_schema(p_db_owner,p_table_name,p_table_type)
DEFINE
  p_db_owner   STRING,
  p_table_name STRING,
  p_table_type STRING  
DEFINE
  ls_db_owner   STRING,
  ls_table_name STRING, 
  ls_table_type STRING,  
  ls_command    STRING,
  lb_success    BOOLEAN,
  ls_message    STRING
DEFINE
  lb_return BOOLEAN   

  LET ls_db_owner   = cs_master_db --p_db_owner #20191209
  LET ls_table_name = p_table_name
  LET ls_table_type = NVL(p_table_type,cs_null_value)

  LET lb_success = TRUE
  
  LET ls_command = "r.s ",cs_master_db," ",ls_table_name.trim()
  DISPLAY cs_command_tag,ls_command

  -- 20160513
  IF ls_table_type = "L" THEN  
    &ifndef DEBUG
    IF g_bgjob <> "Y" THEN 
      #CALL cl_progress_bar(2)
      #CALL cl_progress_ing(ls_command)
      CALL sadzp000_msg_show_message_bar("Generate database schema file ... "||ls_command,NULL,g_lang) #20170523
    END IF   
    CALL cl_cmdrun_openpipe("r.s",ls_command,TRUE) RETURNING lb_success,ls_message
    IF g_bgjob <> "Y" THEN
      #CALL cl_progress_ing(NVL(ls_message,cs_status_finalizing))
      CALL sadzp000_msg_show_message_bar(cs_status_finalizing,NULL,g_lang) #20170523
    END IF   
    &endif
  ELSE 
    RUN ls_command WITHOUT WAITING
  END IF;  

  LET lb_return = lb_success

  RETURN lb_return

END FUNCTION

#取得各DB Tool的SQL註解符號
#p_Sign--B:取得開始註解符號; E:取得結束註解符號
FUNCTION sadzi140_db_get_comment_sign(p_sign)
DEFINE 
  p_sign   CHAR(02)
DEFINE 
  ls_sign    CHAR(02),
  ls_comment STRING,
  ls_return  STRING,
  ls_db_type STRING #20200330

  LET ls_sign   = p_sign
  LET ls_return = ""

  LET ls_db_type = sadzi140_util_get_db_type()
   
  IF ls_sign = "B" THEN
    CASE ls_db_type
      WHEN cs_db_driver_type_ifx 
        LET ls_comment= "{ "
      WHEN cs_db_driver_type_kdb #20240521 
        LET ls_comment = "/* "
      WHEN cs_db_driver_type_ora 
        LET ls_comment = "/* "
      WHEN cs_db_driver_type_dmg --20230804 
        LET ls_comment = "/* "
      WHEN cs_db_driver_type_pgs 
        LET ls_comment = "/* "
      WHEN cs_db_driver_type_msv 
        LET ls_comment = "/* "
    OTHERWISE    
      LET ls_comment = "/* " --20230804 
    END CASE
  ELSE
    CASE ls_db_type
      WHEN cs_db_driver_type_ifx 
        LET ls_comment = " }"
      WHEN cs_db_driver_type_kdb #20240521 
        LET ls_comment = " */"
      WHEN cs_db_driver_type_ora 
        LET ls_comment = " */"
      WHEN cs_db_driver_type_dmg --20230804 
        LET ls_comment = " */"
      WHEN cs_db_driver_type_pgs 
        LET ls_comment = " */"
      WHEN cs_db_driver_type_msv 
        LET ls_comment = " */"
    OTHERWISE    
      LET ls_comment = " */"  --20230804 
    END CASE
  END IF

  LET ls_return = ls_comment
 
  RETURN ls_return
   
END FUNCTION

FUNCTION sadzi140_db_gen_drop_script(p_table_name,p_type)
DEFINE
  p_table_name  STRING,
  p_type        STRING
DEFINE
  ##########################
  ls_drop_script         STRING,
  ls_drop_rebuild_script STRING,
  ls_exit_sign           STRING,
  ls_drop_full           STRING,
  ls_object_type         STRING,
  ##########################
  ls_table_name       STRING,
  ls_type             STRING,
  ls_random_name      STRING,
  ls_tempdir          STRING,
  ls_db_type          STRING, #20240521 
  lchannel_write      base.Channel,
  ls_sript_filename   STRING,
  ls_procedure_script STRING,
  ls_separator        STRING 
DEFINE
  ls_return STRING  
  
  LET ls_table_name = p_table_name
  LET ls_type       = p_type

  LET ls_separator = os.Path.separator()
  
  LET ls_tempdir = FGL_GETENV("TEMPDIR")
  CALL sadzi140_db_get_GUID() RETURNING ls_random_name 
  LET ls_sript_filename = ls_tempdir,ls_separator,ls_table_name,ls_random_name,".drp"

  LET ls_exit_sign = sadzi140_util_get_exit_sign() --cs_exit_sig --"exit;"
  
  LET lchannel_write = base.Channel.create()
  CALL lchannel_write.setDelimiter("")

  IF ls_type = 'T' THEN
    LET ls_object_type = "TABLE"
    &ifdef is_msv
    LET ls_drop_rebuild_script = "drop table if exists ",ls_table_name,cs_rebuild_tail,"\n","go","\n"
    LET ls_drop_script = "drop table if exists ",ls_table_name,"\n","go","\n"
    &else
    &ifdef is_pgs
    LET ls_drop_rebuild_script = "drop table if exists ",ls_table_name,cs_rebuild_tail," cascade;","\n"
    LET ls_drop_script = "drop table if exists ",ls_table_name," cascade;","\n"
    &else
    LET ls_db_type = sadzi140_util_get_db_type() #20240521
    CASE ls_db_type
      WHEN cs_db_driver_type_kdb #20240521
        LET ls_drop_rebuild_script = "drop table if exists ",ls_table_name,cs_rebuild_tail,";","\n"
        LET ls_drop_script = "drop table if exists ",ls_table_name,";","\n"
      WHEN cs_db_driver_type_dmg #20230804
        LET ls_drop_rebuild_script = "drop table ",ls_table_name,cs_rebuild_tail," cascade constraints"
        LET ls_drop_script = "drop table ",ls_table_name," cascade constraints"
    OTHERWISE   
      LET ls_drop_rebuild_script = "drop table ",ls_table_name,cs_rebuild_tail," cascade constraints"
      LET ls_drop_script = "drop table ",ls_table_name," cascade constraints"
    END CASE  
    &endif
    &endif
  ELSE
    LET ls_object_type = "SYNONYM"
    &ifdef is_msv
    LET ls_drop_rebuild_script = "","\n","go","\n"
    LET ls_drop_script = "drop synonym if exists ",ls_table_name,"\n","go","\n"
    &else
    &ifdef is_pgs
    IF sadzi140_util_get_pgs_branch_type() = cs_pgs_sub_pgs THEN 
      -- PostgreSQL Using View
      LET ls_drop_rebuild_script = "","\n"
      LET ls_drop_script = "drop view if exists ",ls_table_name,";","\n"
    ELSE
      -- MogDB Using Synonym
      LET ls_drop_rebuild_script = "","\n"
      LET ls_drop_script = "drop synonym if exists ",ls_table_name,";","\n"
    END IF  
    &else
    LET ls_db_type = sadzi140_util_get_db_type() #20240521
    CASE ls_db_type
      WHEN cs_db_driver_type_kdb #20240521
        -- KDB Using Synonym
        LET ls_drop_rebuild_script = "","\n"
        LET ls_drop_script = "drop synonym if exists ",ls_table_name,";","\n"
      WHEN cs_db_driver_type_dmg #20230804
        LET ls_drop_rebuild_script = ""
        LET ls_drop_script = "drop synonym ",ls_table_name
    OTHERWISE   
      LET ls_drop_rebuild_script = ""
      LET ls_drop_script = "drop synonym ",ls_table_name
    END CASE  
    &endif
    &endif
  END IF

  &ifdef is_msv
  LET ls_procedure_script = "","\n",
                            ls_drop_rebuild_script,
                            ls_drop_script 
  &else
  &ifdef is_pgs
  LET ls_procedure_script = "","\n",
                            ls_drop_rebuild_script,
                            ls_drop_script 
  &else
  LET ls_procedure_script = "-- Drop table or synonym procedure.                                        ","\n",
                            "SET SERVEROUTPUT ON                                                        ","\n",
                            "DECLARE                                                                    ","\n",
                            "  ls_SQL        VARCHAR2(1024);                                            ","\n",
                            "  ls_ObjectType VARCHAR2(10);                                              ","\n",
                            "BEGIN                                                                      ","\n",
                            "   DBMS_OUTPUT.ENABLE(100000);                                             ","\n",
                            "                                                                           ","\n",
                            "   ls_ObjectType := '",ls_object_type,"';                                  ","\n",
                            "                                                                           ","\n",
                            "   IF (ls_ObjectType = 'TABLE') THEN                                       ","\n",
                            "     ls_SQL := '",ls_drop_rebuild_script,"';                               ","\n",
                            "     BEGIN                                                                 ","\n",
                            "       EXECUTE IMMEDIATE ls_SQL;                                           ","\n",
                            "       DBMS_OUTPUT.PUT_LINE('[Message] Drop rebuild table success.');      ","\n",                    
                            "     EXCEPTION                                                             ","\n",
                            "       WHEN OTHERS THEN                                                    ","\n",
                            "         DBMS_OUTPUT.PUT_LINE(SQLERRM);                                    ","\n",
                            "     END;                                                                  ","\n",
                            "   END IF;                                                                 ","\n",
                            "                                                                           ","\n",
                            "   ls_SQL := '",ls_drop_script,"';                                         ","\n",
                            "                                                                           ","\n",
                            "   BEGIN                                                                   ","\n",
                            "     EXECUTE IMMEDIATE ls_SQL;                                             ","\n",
                            "     DBMS_OUTPUT.PUT_LINE('[Message] Drop table success.');                ","\n",          
                            "   EXCEPTION                                                               ","\n",
                            "     WHEN OTHERS THEN                                                      ","\n",
                            "         DBMS_OUTPUT.PUT_LINE(SQLERRM);                                    ","\n",
                            "   END;                                                                    ","\n",
                            "                                                                           ","\n",
                            "EXCEPTION                                                                  ","\n",
                            "  WHEN OTHERS THEN                                                         ","\n",
                            "    DBMS_OUTPUT.PUT_LINE(SQLERRM);                                         ","\n",
                            "END;                                                                       ","\n",
                            "/                                                                          ","\n"
                            
  LET ls_db_type = sadzi140_util_get_db_type() #20240521
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb #20240521
      LET ls_procedure_script = "","\n",
                                ls_drop_rebuild_script,
                                ls_drop_script 
    WHEN cs_db_driver_type_dmg #20230804
      --Using Oracle procedure scrupts
  OTHERWISE
    --Using Oracle procedure scrupts  
  END CASE  
  &endif
  &endif
  
  LET ls_drop_full = ls_procedure_script,"\n",
                     ls_exit_sign
  
  #開檔寫入
  TRY
    CALL lchannel_write.openFile(ls_sript_filename, "w" )
    CALL lchannel_write.write(ls_drop_full)
    #關檔
    CALL lchannel_write.close()
  CATCH
    DISPLAY "Generate Script Error !!"
  END TRY  

  LET ls_return = ls_sript_filename
  
  RETURN ls_return

END FUNCTION


FUNCTION sadzi140_db_gen_rebuild_script(p_owner,p_table_name)
DEFINE
  p_owner       STRING, -- 20170317
  p_table_name  STRING,
  p_type        STRING
DEFINE
  ##########################
  ls_drop_script      STRING,
  ls_exit_sign        STRING,
  ls_drop_full        STRING,
  ##########################
  ls_owner            STRING, --20170317
  ls_table_name       STRING,
  ls_type             STRING,
  ls_random_name      STRING,
  ls_tempdir          STRING,
  lchannel_write      base.Channel,
  ls_db_type          STRING, #20240521
  ls_sript_filename   STRING,
  ls_spool_filename   STRING, -- 20170317
  ls_procedure_script STRING,
  ls_separator        STRING 
DEFINE
  ls_rtn_reb STRING,  
  ls_rtn_spl STRING  

  LET ls_owner = p_owner --20170317
  LET ls_table_name = p_table_name

  LET ls_separator = os.Path.separator()
  
  LET ls_tempdir = FGL_GETENV("TEMPDIR")
  CALL sadzi140_db_get_GUID() RETURNING ls_random_name 
  LET ls_sript_filename = ls_tempdir,ls_separator,ls_owner,"_",ls_table_name,"_",ls_random_name,".reb"
  LET ls_spool_filename = ls_tempdir,ls_separator,ls_owner,"_",ls_table_name,"_",ls_random_name,".spl" --20170317
  
  LET ls_exit_sign = sadzi140_util_get_exit_sign() --cs_exit_sig
  
  LET lchannel_write = base.Channel.create()
  CALL lchannel_write.setDelimiter("")

  &ifdef is_msv
  LET ls_procedure_script = "--Rebuild table T-SQL procedure not support ","\n"
  &else
  &ifdef is_pgs
  LET ls_procedure_script = "--Rebuild table PLpgSQL procedure not support ","\n"
  &else
  #Pure Oracle or DM
  #20170214 begin
  #20170317 begin
  LET ls_procedure_script = "-- Rebuild table PL/SQL procedure                                                                            ","\n",
                            "SET SERVEROUTPUT ON                                                                                          ","\n",
                            "SET TRIMOUT ON                                                                                               ","\n",
                            "SET TRIMSPOOL ON                                                                                             ","\n",
                            "SET WRAP OFF                                                                                                 ","\n",
                            "SET TERMOUT OFF                                                                                              ","\n",
                            "SET PAGESIZE 0                                                                                               ","\n",
                            "SET FEEDBACK OFF                                                                                             ","\n",
                            "SET LINESIZE 32000                                                                                           ","\n",
                            "                                                                                                             ","\n",
                            "SPOOL '",ls_spool_filename,"'                                                                                ","\n",
                            "                                                                                                             ","\n",
                            "DECLARE                                                                                                      ","\n",
                            "  type t_trigger_content is record                                                                           ","\n",
                            "  (                                                                                                          ","\n",
                            "    table_name    all_triggers.TABLE_NAME%type,                                                              ","\n",
                            "    trigger_name  all_triggers.TRIGGER_NAME%type,                                                            ","\n",
                            "    trigger_ddl   varchar2(32767),                                                                           ","\n",
                            "    trigger_body  varchar2(32767),                                                                           ","\n",
                            "    trigger_alter varchar2(32767)                                                                            ","\n",
                            "  );                                                                                                         ","\n",
                            "  type t_trigger_list is table of t_trigger_content index by pls_integer;                                    ","\n",
                            "  trigger_list t_trigger_list;                                                                               ","\n",
                            "  -------------------------------------------------------------------------------                            ","\n",
                            "  cursor c_trigger_list(p_owner varchar2,p_table_name varchar2) IS                                           ","\n",
                            "    SELECT TABLE_NAME,TRIGGER_NAME,DBMS_METADATA.GET_DDL('TRIGGER', TRIGGER_NAME) TRIGGER_DDL                ","\n", #20170315
                            "      FROM ALL_TRIGGERS                                                                                      ","\n",
                            "     WHERE OWNER = p_owner                                                                                   ","\n",
                            "       AND TABLE_NAME = p_table_name;                                                                        ","\n",
                            "  r_trigger_list c_trigger_list%ROWTYPE;                                                                     ","\n",
                            "  -------------------------------------------------------------------------------                            ","\n",
                            "  cursor c_index(p_owner varchar2,p_table_name varchar2) IS                                                  ","\n",
                            "    select *                                                                                                 ","\n",
                            "      from all_indexes ais                                                                                   ","\n",
                            "     where ais.OWNER      = p_owner                                                                          ","\n",
                            "       and ais.TABLE_NAME = p_table_name;                                                                    ","\n",
                            "  r_index c_index%ROWTYPE;                                                                                   ","\n",
                            "  -------------------------------------------------------------------------------                            ","\n",
                            "  cursor c_constraints(p_owner varchar2,p_table_name varchar2) IS                                            ","\n",
                            "    select *                                                                                                 ","\n",
                            "      from all_constraints acs                                                                               ","\n",
                            "     where acs.OWNER      = p_owner                                                                          ","\n",
                            "       and acs.TABLE_NAME = p_table_name;                                                                    ","\n",
                            "  r_constraints c_constraints%ROWTYPE;                                                                       ","\n",
                            "  -------------------------------------------------------------------------------                            ","\n",
                            "  cs_ERROR_TAG    CONSTANT VARCHAR2(20) := '",cs_error_tag,"';                                               ","\n",          
                            "  cs_WARNING_TAG  CONSTANT VARCHAR2(20) := '",cs_warning_tag,"';                                             ","\n",            
                            "  cs_SUCCESS_TAG  CONSTANT VARCHAR2(20) := '",cs_success_tag,"';                                             ","\n",            
                            "  ls_SQL          VARCHAR2(1024);                                                                            ","\n",
                            "  ls_TableName    VARCHAR2(512);                                                                             ","\n",
                            "  ls_user         VARCHAR2(512);                                                                             ","\n",
                            "  ls_banner       VARCHAR2(64);                                                                              ","\n", --20230804
                            "  ls_TableSpace   VARCHAR2(512);                                                                             ","\n",
                            "  ls_RebuildTable VARCHAR2(512);                                                                             ","\n",
                            "  li_counts       INTEGER;                                                                                   ","\n",
                            "  ---- For Trigger rebuild ------------------------------------------------------                            ","\n",
                            "  li_index     integer;                                                                                      ","\n",
                            "  ls_exec_sql  varchar2(32767);                                                                              ","\n",
                            "  -------------------------------------------------------------------------------                            ","\n",
                            "BEGIN                                                                                                        ","\n",
                            "  DBMS_OUTPUT.ENABLE(100000);                                                                                ","\n",
                            "                                                                                                             ","\n",
                            "  ls_TableName := '",ls_table_name.toUpperCase(),"';                                                         ","\n",
                            "  ls_RebuildTable := ls_TableName||'",cs_rebuild_tail.toUpperCase(),"';                                      ","\n",                
                            "                                                                                                             ","\n",
                            "  select user                                                                                                ","\n",
                            "    into ls_user                                                                                             ","\n",
                            "    from dual;                                                                                               ","\n",
                            "                                                                                                             ","\n",
                            "  select ats.TABLESPACE_NAME                                                                                 ","\n",
                            "    into ls_TableSpace                                                                                       ","\n",
                            "    from all_tables ats                                                                                      ","\n",
                            "   where ats.OWNER      = ls_user                                                                            ","\n",
                            "     and ats.TABLE_NAME = ls_TableName;                                                                      ","\n",
                            "                                                                                                             ","\n",
                            "  ----- Backup Trigger DDL and drop Trigger begin -----                                                      ","\n",
                            "  li_index := 1;                                                                                             ","\n",
                            "  open c_trigger_list(ls_user,ls_TableName);                                                                 ","\n",
                            "  LOOP                                                                                                       ","\n",
                            "    fetch c_trigger_list into r_trigger_list;                                                                ","\n",
                            "    if c_trigger_list%notfound then                                                                          ","\n",
                            "      exit;                                                                                                  ","\n",
                            "    end if;                                                                                                  ","\n",
                            "                                                                                                             ","\n",
                            "    trigger_list(li_index).table_name   := r_trigger_list.TABLE_NAME;                                        ","\n",
                            "    trigger_list(li_index).trigger_name := r_trigger_list.trigger_name;                                      ","\n",
                            "    trigger_list(li_index).trigger_ddl  := r_trigger_list.trigger_ddl;                                       ","\n",
                            "                                                                                                             ","\n",
                            "    select replace(substr(r_trigger_list.trigger_ddl,1,instr(r_trigger_list.trigger_ddl,'ALTER')-1),'EDITIONABLE','')       ","\n",
                            "      into trigger_list(li_index).trigger_body                                                               ","\n",
                            "      from dual;                                                                                             ","\n",
                            "                                                                                                             ","\n",
                            "    --DBMS_OUTPUT.PUT_LINE(trim(trigger_list(li_index).trigger_body)||CHR(10)||'/'||CHR(10));                ","\n",
                            "    DBMS_OUTPUT.PUT_LINE(trim(trigger_list(li_index).trigger_body));                                         ","\n",
                            "    DBMS_OUTPUT.NEW_LINE;                                                                                    ","\n",
                            "    DBMS_OUTPUT.PUT_LINE('/');                                                                               ","\n",
                            "    DBMS_OUTPUT.NEW_LINE;                                                                                    ","\n",
                            "                                                                                                             ","\n",
                            "    select substr(r_trigger_list.trigger_ddl,instr(r_trigger_list.trigger_ddl,'ALTER'),length(r_trigger_list.trigger_ddl))  ","\n", 
                            "      into trigger_list(li_index).trigger_alter                                                              ","\n",
                            "      from dual;                                                                                             ","\n",
                            "                                                                                                             ","\n",
                            "    --DBMS_OUTPUT.PUT_LINE(trim(trigger_list(li_index).trigger_alter)||';'||CHR(10));                        ","\n",
                            "    DBMS_OUTPUT.PUT_LINE(trim(trigger_list(li_index).trigger_alter)||';');                                   ","\n",
                            "    DBMS_OUTPUT.NEW_LINE;                                                                                    ","\n",
                            "                                                                                                             ","\n",
                            "    BEGIN                                                                                                    ","\n",
                            "      ls_exec_sql := 'drop trigger '||r_trigger_list.trigger_name;                                           ","\n",
                            "      --Dbms_Output.put_line(ls_exec_sql);                                                                   ","\n",
                            "      execute immediate ls_exec_sql;                                                                         ","\n",
                            "    EXCEPTION                                                                                                ","\n",
                            "      WHEN OTHERS THEN                                                                                       ","\n",
                            "        DBMS_OUTPUT.PUT_LINE(cs_WARNING_TAG||'drop trigger '||r_trigger_list.trigger_name);                  ","\n",
                            "    END;                                                                                                     ","\n",
                            "                                                                                                             ","\n",
                            "    li_index := li_index + 1;                                                                                ","\n",
                            "                                                                                                             ","\n",
                            "  END LOOP;                                                                                                  ","\n",
                            "  close c_trigger_list;                                                                                      ","\n",
                            "                                                                                                             ","\n",
                            "  DBMS_OUTPUT.PUT_LINE('exit;');                                                                             ","\n",                                                                                
                            "                                                                                                             ","\n",
                            "  ----- Backup Trigger DDL and drop Trigger end -----                                                        ","\n",
                            "                                                                                                             ","\n",
                            "  -- 先刪除舊的備份表格                                                                                         ","\n",
                            "  BEGIN                                                                                                      ","\n",
                            "    ls_SQL := 'DROP TABLE '||ls_RebuildTable;                                                                ","\n",
                            "    EXECUTE IMMEDIATE ls_SQL;                                                                                ","\n",
                            "  EXCEPTION                                                                                                  ","\n",
                            "    WHEN OTHERS THEN                                                                                         ","\n",
                            "      DBMS_OUTPUT.PUT_LINE(cs_WARNING_TAG||' '||ls_SQL);                                                     ","\n",
                            "  END;                                                                                                       ","\n",
                            "                                                                                                             ","\n",
                            "  -- 將表格更名為備份表格名稱                                                                                    ","\n",
                            "  BEGIN                                                                                                      ","\n",
                            "    ls_SQL := 'ALTER TABLE '||ls_TableName||' RENAME TO '||ls_RebuildTable;                                  ","\n",
                            "    EXECUTE IMMEDIATE ls_SQL;                                                                                ","\n",
                            "  EXCEPTION                                                                                                  ","\n",
                            "    WHEN OTHERS THEN                                                                                         ","\n",
                            "      DBMS_OUTPUT.PUT_LINE(cs_ERROR_TAG||' '||ls_SQL);                                                       ","\n",
                            "      RAISE;                                                                                                 ","\n",
                            "  END;                                                                                                       ","\n",
                            "                                                                                                             ","\n",
                            "  -- 刪除表格的Key及Index                                                                                      ","\n",
                            "  SELECT upper(substr(banner,1,instr(banner,' ')-1)) banner                                                  ","\n",
                            "    INTO ls_banner                                                                                           ","\n",
                            "    FROM v$version                                                                                           ","\n",
                            "   WHERE rownum = 1;                                                                                         ","\n",
                            "                                                                                                             ","\n",
                            "  BEGIN                                                                                                      ","\n",
                            "    IF ls_banner = 'ORACLE' THEN                                                                             ","\n",
                            "      ls_SQL := 'ALTER TABLE '||ls_RebuildTable||' DROP PRIMARY KEY CASCADE DROP INDEX';                     ","\n",
                            "    ELSE                                                                                                     ","\n",
                            "      ls_SQL := 'ALTER TABLE '||ls_RebuildTable||' DROP PRIMARY KEY CASCADE';                                ","\n",
                            "    END IF;                                                                                                  ","\n",
                            "    EXECUTE IMMEDIATE ls_SQL;                                                                                ","\n",
                            "  EXCEPTION                                                                                                  ","\n",
                            "    WHEN OTHERS THEN                                                                                         ","\n",
                            "      DBMS_OUTPUT.PUT_LINE(cs_WARNING_TAG||' '||ls_SQL);                                                     ","\n",
                            "      DBMS_OUTPUT.PUT_LINE(SQLERRM);                                                                         ","\n",
                            "  END;                                                                                                       ","\n",
                            "                                                                                                             ","\n",
                            "  -- 刪除表格其他的Constraint                                                                                ","\n",
                            "  li_counts := 0;                                                                                            ","\n",
                            "  open c_constraints(ls_user,ls_RebuildTable);                                                               ","\n",
                            "  LOOP                                                                                                       ","\n",
                            "    fetch c_constraints into r_constraints;                                                                  ","\n",
                            "    if c_constraints%notfound then                                                                           ","\n",
                            "      exit;                                                                                                  ","\n",
                            "    end if;                                                                                                  ","\n",
                            "                                                                                                             ","\n",
                            "    ls_SQL := 'alter table '||ls_RebuildTable||' drop constraint '||r_constraints.CONSTRAINT_NAME||' cascade'; ","\n",
                            "                                                                                                             ","\n",
                            "    BEGIN                                                                                                    ","\n",
                            "      EXECUTE IMMEDIATE ls_SQL;                                                                              ","\n",
                            "    EXCEPTION                                                                                                ","\n",
                            "      WHEN OTHERS THEN                                                                                       ","\n",
                            "        DBMS_OUTPUT.PUT_LINE(cs_WARNING_TAG||'Drop CONSTRAINT '||r_constraints.CONSTRAINT_NAME||' fault.');  ","\n",
                            "        DBMS_OUTPUT.PUT_LINE(SQLERRM);                                                                       ","\n",
                            "    END;                                                                                                     ","\n",
                            "                                                                                                             ","\n",
                            "    li_counts := li_counts + 1;                                                                              ","\n",
                            "                                                                                                             ","\n",
                            "  END LOOP;                                                                                                  ","\n",
                            "  close c_constraints;                                                                                       ","\n",
                            "                                                                                                             ","\n",
                            "  -- 刪除表格其他的Index                                                                                     ","\n",
                            "  li_counts := 0;                                                                                            ","\n",
                            "  open c_index(ls_user,ls_RebuildTable);                                                                     ","\n",
                            "  LOOP                                                                                                       ","\n",
                            "    fetch c_index into r_index;                                                                              ","\n",
                            "    if c_index%notfound then                                                                                 ","\n",
                            "      exit;                                                                                                  ","\n",
                            "    end if;                                                                                                  ","\n",
                            "                                                                                                             ","\n",
                            "    ls_SQL := 'DROP INDEX '||r_index.INDEX_NAME;                                                             ","\n",
                            "                                                                                                             ","\n",
                            "    BEGIN                                                                                                    ","\n",
                            "      EXECUTE IMMEDIATE ls_SQL;                                                                              ","\n",
                            "    EXCEPTION                                                                                                ","\n",
                            "      WHEN OTHERS THEN                                                                                       ","\n",
                            "        DBMS_OUTPUT.PUT_LINE(cs_WARNING_TAG||'Drop INDEX '||r_index.INDEX_NAME||' fault.');                  ","\n",
                            "        DBMS_OUTPUT.PUT_LINE(SQLERRM);                                                                       ","\n",
                            "    END;                                                                                                     ","\n",
                            "                                                                                                             ","\n",
                            "    li_counts := li_counts + 1;                                                                              ","\n",
                            "                                                                                                             ","\n",
                            "  END LOOP;                                                                                                  ","\n",
                            "  close c_index;                                                                                             ","\n",
                            "                                                                                                             ","\n",
                            "  -- 重新建立表格等待異動                                                                                    ","\n",
                            "  BEGIN                                                                                                      ","\n",
                            "    ls_SQL := 'CREATE TABLE '||ls_TableName||' TABLESPACE '||ls_TableSpace||                                 ","\n",
                            "              '          AS SELECT * FROM DUMMY WHERE 1=2';                                                  ","\n",
                            "    EXECUTE IMMEDIATE ls_SQL;                                                                                ","\n",
                            "  EXCEPTION                                                                                                  ","\n",
                            "    WHEN OTHERS THEN                                                                                         ","\n",
                            "      DBMS_OUTPUT.PUT_LINE(cs_ERROR_TAG||' '||ls_SQL);                                                       ","\n",
                            "      RAISE;                                                                                                 ","\n",
                            "  END;                                                                                                       ","\n",
                            "                                                                                                             ","\n",
                            "  -- 賦予權限(TIPTOP)                                                                                        ","\n",
                            "  BEGIN                                                                                                      ","\n",
                            "    ls_SQL := 'grant select, insert, update, delete on '||ls_TableName||' to TIPTOP';                        ","\n",
                            "    EXECUTE IMMEDIATE ls_SQL;                                                                                ","\n",
                            "  EXCEPTION                                                                                                  ","\n",
                            "    WHEN OTHERS THEN                                                                                         ","\n",
                            "      DBMS_OUTPUT.PUT_LINE(cs_ERROR_TAG||' '||ls_SQL);                                                       ","\n",
                            "      RAISE;                                                                                                 ","\n",
                            "  END;                                                                                                       ","\n",
                            "                                                                                                             ","\n",
                            "  DBMS_OUTPUT.PUT_LINE(cs_SUCCESS_TAG);                                                                      ","\n",
                            "                                                                                                             ","\n",
                            "  /*                                                                                                         ","\n",
                            "  ----- Recreate Trigger begin -----                                                                         ","\n",
                            "  li_index := trigger_list.first;                                                                            ","\n",
                            "  loop                                                                                                       ","\n",
                            "    exit when li_index is null;                                                                              ","\n",
                            "                                                                                                             ","\n",
                            "    --dbms_output.put_line('table_name : '||trigger_list(li_index).table_name||chr(13)||                     ","\n",
                            "    --                     'trigger_name : '||trigger_list(li_index).trigger_name||chr(13)||                 ","\n",
                            "    --                     'trigger_ddl : '||trigger_list(li_index).trigger_ddl||chr(13)||                   ","\n",
                            "    --                     'trigger_body : '||trigger_list(li_index).trigger_body||chr(13)||                 ","\n",
                            "    --                     'trigger_alter : '||trigger_list(li_index).trigger_alter                          ","\n",
                            "    --                    );                                                                                 ","\n",
                            "                                                                                                             ","\n",
                            "    BEGIN                                                                                                    ","\n",
                            "      ls_exec_sql := trigger_list(li_index).trigger_body;                                                    ","\n",
                            "      Dbms_Output.put_line('Recreate trigger : '||trigger_list(li_index).trigger_name);                      ","\n",
                            "      execute immediate ls_exec_sql;                                                                         ","\n",
                            "    EXCEPTION                                                                                                ","\n",
                            "      WHEN OTHERS THEN                                                                                       ","\n",
                            "        DBMS_OUTPUT.PUT_LINE(cs_WARNING_TAG||'Recreate trigger : '||trigger_list(li_index).trigger_name);    ","\n",
                            "    END;                                                                                                     ","\n",
                            "                                                                                                             ","\n",
                            "    BEGIN                                                                                                    ","\n",
                            "      ls_exec_sql := trigger_list(li_index).trigger_alter;                                                   ","\n",
                            "      Dbms_Output.put_line('Alter trigger : '||trigger_list(li_index).trigger_name);                         ","\n",
                            "      execute immediate ls_exec_sql;                                                                         ","\n",
                            "    EXCEPTION                                                                                                ","\n",
                            "      WHEN OTHERS THEN                                                                                       ","\n",
                            "        DBMS_OUTPUT.PUT_LINE(cs_WARNING_TAG||'Alter trigger : '||trigger_list(li_index).trigger_name);       ","\n",
                            "    END;                                                                                                     ","\n",
                            "                                                                                                             ","\n",
                            "    li_index := trigger_list.next(li_index);                                                                 ","\n",
                            "                                                                                                             ","\n",
                            "  end loop;                                                                                                  ","\n",
                            "  ----- Recreate Trigger end -----                                                                           ","\n",
                            "  */                                                                                                         ","\n",
                            "                                                                                                             ","\n",
                            "EXCEPTION                                                                                                    ","\n",
                            "  WHEN OTHERS THEN                                                                                           ","\n",
                            "    DBMS_OUTPUT.PUT_LINE(cs_ERROR_TAG||SQLERRM);                                                             ","\n",
                            "END;                                                                                                         ","\n",
                            "/                                                                                                            ","\n",
                            "                                                                                                             ","\n",
                            "SPOOL OFF                                                                                                    ","\n",
                            "                                                                                                             ","\n"
  #20170317 end
  #20170214 end

  LET ls_db_type = sadzi140_util_get_db_type() #20240521
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb #20240521
      LET ls_procedure_script = "--Rebuild table KDB procedure not support.","\n"
    WHEN cs_db_driver_type_dmg --20230804
      -- Using Oracle Procedure scripts
  OTHERWISE
    -- Using Oracle Procedure scripts
  END CASE
  
  &endif
  &endif
                
  LET ls_drop_full = ls_procedure_script,"\n",
                     ls_exit_sign
  
  #開檔寫入
  TRY
    CALL lchannel_write.openFile(ls_sript_filename, "w" )
    CALL lchannel_write.write(ls_drop_full)
    #關檔
    CALL lchannel_write.close()
  CATCH
    DISPLAY "Generate Script Error !!"
  END TRY  

  LET ls_rtn_reb = ls_sript_filename
  LET ls_rtn_spl = ls_spool_filename #20170317
  
  RETURN ls_rtn_reb,ls_rtn_spl #20170317

END FUNCTION

#異動實際表格型態的 Procedure
#若目標Schema的表格或是同義字不存在,會以ds的內容,並參照dzeu_t的型態建立之
FUNCTION sadzi140_db_get_alter_table_type_procedure(p_table_name,p_schema_name,p_master_schema)
DEFINE
  p_table_name     STRING,
  p_schema_name    STRING,
  p_master_schema  STRING
DEFINE
  ls_table_name       STRING,
  ls_schema_name      STRING,
  ls_master_schema    STRING,
  ls_procedure_script STRING,
  ls_db_type          STRING #20240521
DEFINE
  ls_return STRING  
  
  LET ls_table_name    = p_table_name
  LET ls_schema_name   = p_schema_name
  LET ls_master_schema = p_master_schema

  &ifdef is_msv
  LET ls_procedure_script = "--Create non exist table or synonymous T-SQL procedure. ","\n"
  &else
  &ifdef is_pgs
  LET ls_procedure_script = "--Create non exist table or synonymous PLpgSQL procedure. ","\n"
  &else
  LET ls_procedure_script = "-- Create non exist table or synonymous PL/SQL procedure.                                                                                                ","\n",
                            "DECLARE                                                                                                                                                  ","\n",
                            "  --------------------------------------------------------                                                                                               ","\n",
                            "  cursor c_index(p_owner varchar2,p_table_name varchar2) IS                                                                                              ","\n",
                            "    SELECT *                                                                                                                                             ","\n",
                            "      FROM ALL_INDEXES AIS                                                                                                                               ","\n",
                            "     WHERE AIS.OWNER      = p_owner                                                                                                                      ","\n",
                            "       AND AIS.TABLE_NAME = p_table_name;                                                                                                                ","\n",
                            "  r_index c_index%ROWTYPE;                                                                                                                               ","\n",
                            "  --------------------------------------------------------                                                                                               ","\n",
                            "  li_counts          INTEGER;                                                                                                                            ","\n",
                            "  li_cnts            INTEGER;                                                                                                                            ","\n",
                            "  ls_table_name      VARCHAR2(50);                                                                                                                       ","\n",
                            "  ls_schema_name     VARCHAR2(50);                                                                                                                       ","\n",
                            "  ls_master_schema   VARCHAR2(50);                                                                                                                       ","\n",
                            "  ls_user            VARCHAR2(50);                                                                                                                       ","\n",
                            "  ls_db_name         VARCHAR2(50);                                                                                                                       ","\n",
                            "  ls_table_type      VARCHAR2(50);                                                                                                                       ","\n",
                            "  ls_real_table_type VARCHAR2(50);                                                                                                                       ","\n",
                            "  ls_backup_name     VARCHAR2(50);                                                                                                                       ","\n",
                            "  ls_ExecScript      VARCHAR2(4096);                                                                                                                     ","\n",
                            "  --------------------------------------------------------                                                                                               ","\n",
                            "  cs_ERROR_TAG       CONSTANT VARCHAR2(20) := '",cs_error_tag,"';                                                                                        ","\n",
                            "  cs_WARNING_TAG     CONSTANT VARCHAR2(20) := '",cs_warning_tag,"';                                                                                      ","\n",
                            "  cs_SUCCESS_TAG     CONSTANT VARCHAR2(20) := '",cs_success_tag,"';                                                                                      ","\n",
                            "  --------------------------------------------------------                                                                                               ","\n",
                            "BEGIN                                                                                                                                                    ","\n",
                            "                                                                                                                                                         ","\n",
                            "  ls_table_name    := '",ls_table_name.toUpperCase(),"';                                                                                                 ","\n",
                            "  ls_backup_name   := ls_table_name||'_ALTER';                                                                                                           ","\n",
                            "  ls_schema_name   := '",ls_schema_name.toUpperCase(),"';                                                                                                ","\n",
                            "  ls_master_schema := '",ls_master_schema.toUpperCase(),"';                                                                                              ","\n",
                            "                                                                                                                                                         ","\n",
                            "  SELECT USER                                                                                                                                            ","\n",
                            "    INTO ls_user                                                                                                                                         ","\n",
                            "    FROM DUAL;                                                                                                                                           ","\n",
                            "                                                                                                                                                         ","\n",
                            "  BEGIN                                                                                                                                                  ","\n",
                            "    SELECT DA.GZDA001                                                                                                                                    ","\n",
                            "      INTO ls_db_name                                                                                                                                    ","\n",
                            "      FROM DS.GZDA_T DA                                                                                                                                  ","\n",
                            "     WHERE DA.GZDA001 = LOWER(ls_schema_name);                                                                                                           ","\n", #20191209 gzda003 change to gzda001
                            "  EXCEPTION                                                                                                                                              ","\n",
                            "    WHEN OTHERS THEN                                                                                                                                     ","\n",
                            "      DBMS_OUTPUT.PUT_LINE(cs_ERROR_TAG||'Can not get db name by schema '||ls_schema_name);                                                              ","\n",                           
                            "      ls_db_name := NULL;                                                                                                                                ","\n",      
                            "  END;                                                                                                                                                   ","\n",
                            "                                                                                                                                                         ","\n",
                            "  BEGIN                                                                                                                                                  ","\n",
                            "    SELECT COUNT(*)                                                                                                                                      ","\n",
                            "      INTO li_cnts                                                                                                                                       ","\n",
                            "      FROM ALL_TABLES ATS                                                                                                                                ","\n",
                            "     WHERE ATS.OWNER      = UPPER(ls_schema_name)                                                                                                        ","\n",
                            "       AND ATS.TABLE_NAME = UPPER(ls_table_name);                                                                                                        ","\n",
                            "  EXCEPTION                                                                                                                                              ","\n",
                            "    WHEN OTHERS THEN                                                                                                                                     ","\n",
                            "      li_cnts := 0;                                                                                                                                      ","\n",
                            "  END;                                                                                                                                                   ","\n",
                            "                                                                                                                                                         ","\n",
                            "  BEGIN                                                                                                                                                  ","\n",
                            "    SELECT UPPER(EU.DZEU003) TABLE_TYPE                                                                                                                  ","\n",
                            "      INTO ls_table_type                                                                                                                                 ","\n",
                            "      FROM ",ls_master_schema,".DZEU_T EU                                                                                                                ","\n",
                            "     WHERE EU.DZEU001 = LOWER(ls_table_name)                                                                                                             ","\n",
                            "       AND EU.DZEU002 = LOWER(ls_db_name);                                                                                                               ","\n",
                            "  EXCEPTION                                                                                                                                              ","\n",
                            "    WHEN OTHERS THEN                                                                                                                                     ","\n",
                            "      ls_table_type := 'X';                                                                                                                              ","\n",
                            "  END;                                                                                                                                                   ","\n",
                            "                                                                                                                                                         ","\n",
                            "  BEGIN                                                                                                                                                  ","\n",
                            "    SELECT DECODE(AOS.OBJECT_TYPE,'TABLE','T','SYNONYM','S','INDEX','I','X') TABLE_TYPE                                                                  ","\n",
                            "      INTO ls_real_table_type                                                                                                                            ","\n",
                            "      FROM ALL_OBJECTS AOS                                                                                                                               ","\n",
                            "     WHERE AOS.OWNER       = UPPER(ls_schema_name)                                                                                                       ","\n",
                            "       AND AOS.OBJECT_NAME = UPPER(ls_table_name);                                                                                                       ","\n",
                            "  EXCEPTION                                                                                                                                              ","\n",
                            "    WHEN OTHERS THEN                                                                                                                                     ","\n",
                            "      ls_real_table_type := 'X';                                                                                                                         ","\n",
                            "  END;                                                                                                                                                   ","\n",
                            "  --DBMS_OUTPUT.PUT_LINE(li_cnts);                                                                                                                       ","\n",
                            "                                                                                                                                                         ","\n",
                            "  -- 不為 Master schema 時才做這件事                                                                                                                     ","\n",
                            "  IF (ls_user <> ls_master_schema) AND (ls_db_name IS NOT NULL) THEN                                                                                     ","\n",
                            "    IF (li_cnts = 0) OR (ls_table_type <> ls_real_table_type) THEN                                                                                       ","\n",
                            "                                                                                                                                                         ","\n",
                            "      IF ls_real_table_type = 'T' THEN                                                                                                                   ","\n",
                            "        -- 刪除備份表格名稱                                                                                                                              ","\n",
                            "        BEGIN                                                                                                                                            ","\n",
                            "          ls_ExecScript := 'DROP TABLE '||ls_backup_name;                                                                                                ","\n",
                            "          EXECUTE IMMEDIATE ls_ExecScript;                                                                                                               ","\n",
                            "        EXCEPTION                                                                                                                                        ","\n",
                            "          WHEN OTHERS THEN                                                                                                                               ","\n",
                            "            DBMS_OUTPUT.PUT_LINE(cs_WARNING_TAG||'Drop backup table not success.')  ;                                                                    ","\n",
                            "        END;                                                                                                                                             ","\n",
                            "                                                                                                                                                         ","\n",
                            "        -- 將原表格更名為備份表格名稱                                                                                                                    ","\n",
                            "        BEGIN                                                                                                                                            ","\n",
                            "          ls_ExecScript := 'ALTER TABLE '||ls_table_name||' RENAME TO '||ls_backup_name;                                                                 ","\n",
                            "          EXECUTE IMMEDIATE ls_ExecScript;                                                                                                               ","\n",
                            "        EXCEPTION                                                                                                                                        ","\n",
                            "          WHEN OTHERS THEN                                                                                                                               ","\n",
                            "            DBMS_OUTPUT.PUT_LINE(cs_WARNING_TAG||'Alter original table rename to backup not success.');                                                  ","\n",
                            "            RAISE;                                                                                                                                       ","\n",
                            "        END;                                                                                                                                             ","\n",
                            "                                                                                                                                                         ","\n",
                            "        -- 刪除表格的Key及Index                                                                                                                          ","\n",
                            "        BEGIN                                                                                                                                            ","\n",
                            "          ls_ExecScript := 'ALTER TABLE '||ls_backup_name||' DROP PRIMARY KEY CASCADE DROP INDEX';                                                       ","\n",
                            "          EXECUTE IMMEDIATE ls_ExecScript;                                                                                                               ","\n",
                            "        EXCEPTION                                                                                                                                        ","\n",
                            "          WHEN OTHERS THEN                                                                                                                               ","\n",
                            "            DBMS_OUTPUT.PUT_LINE(cs_WARNING_TAG||'Alter drop primary key and index not success.');                                                       ","\n",
                            "        END;                                                                                                                                             ","\n",
                            "                                                                                                                                                         ","\n",
                            "        -- 刪除表格其他的Index                                                                                                                           ","\n",
                            "        li_counts := 0;                                                                                                                                  ","\n",
                            "        open c_index(ls_user,ls_backup_name);                                                                                                            ","\n",
                            "        LOOP                                                                                                                                             ","\n",
                            "          fetch c_index into r_index;                                                                                                                    ","\n",
                            "          if c_index%notfound then                                                                                                                       ","\n",
                            "            exit;                                                                                                                                        ","\n",
                            "          end if;                                                                                                                                        ","\n",
                            "                                                                                                                                                         ","\n",
                            "          ls_ExecScript := 'DROP INDEX '||r_index.INDEX_NAME;                                                                                            ","\n",
                            "                                                                                                                                                         ","\n",
                            "          BEGIN                                                                                                                                          ","\n",
                            "            EXECUTE IMMEDIATE ls_ExecScript;                                                                                                             ","\n",
                            "          EXCEPTION                                                                                                                                      ","\n",
                            "            WHEN OTHERS THEN                                                                                                                             ","\n",
                            "              DBMS_OUTPUT.PUT_LINE(cs_WARNING_TAG||'Drop INDEX '||r_index.INDEX_NAME||' not success.');                                                  ","\n",
                            "          END;                                                                                                                                           ","\n",
                            "                                                                                                                                                         ","\n",
                            "          li_counts := li_counts + 1;                                                                                                                    ","\n",
                            "                                                                                                                                                         ","\n",
                            "        END LOOP;                                                                                                                                        ","\n",
                            "        close c_index;                                                                                                                                   ","\n",
                            "                                                                                                                                                         ","\n",
                            "      ELSIF ls_real_table_type = 'S' THEN                                                                                                                ","\n",
                            "        BEGIN                                                                                                                                            ","\n",
                            "          ls_ExecScript := 'DROP SYNONYM '||ls_table_name;                                                                                               ","\n",
                            "          EXECUTE IMMEDIATE ls_ExecScript;                                                                                                               ","\n",
                            "        EXCEPTION                                                                                                                                        ","\n",
                            "          WHEN OTHERS THEN                                                                                                                               ","\n",
                            "            DBMS_OUTPUT.PUT_LINE(cs_warning_tag||ls_ExecScript);                                                                                         ","\n",
                            "            DBMS_OUTPUT.PUT_LINE(SQLERRM);                                                                                                               ","\n",
                            "        END;                                                                                                                                             ","\n",
                            "      END IF;                                                                                                                                            ","\n",
                            "                                                                                                                                                         ","\n",
                            "                                                                                                                                                         ","\n",
                            "      IF ls_table_type = 'T' THEN                                                                                                                        ","\n",
                            "        ls_ExecScript := 'CREATE TABLE '||ls_schema_name||'.'||ls_table_name||' AS SELECT * FROM '||ls_master_schema||'.'||ls_table_name||' WHERE 1=2';  ","\n",
                            "        BEGIN                                                                                                                                            ","\n",
                            "          EXECUTE IMMEDIATE ls_ExecScript;                                                                                                               ","\n",
                            "        EXCEPTION                                                                                                                                        ","\n",
                            "          WHEN OTHERS THEN                                                                                                                               ","\n",
                            "            DBMS_OUTPUT.PUT_LINE(cs_ERROR_tag||ls_ExecScript)  ;                                                                                         ","\n",
                            "            DBMS_OUTPUT.PUT_LINE(SQLERRM);                                                                                                               ","\n",
                            "        END;                                                                                                                                             ","\n",
                            "                                                                                                                                                         ","\n",
                            "        ls_ExecScript := 'GRANT SELECT, INSERT, UPDATE, DELETE ON '||ls_table_name||' TO TIPTOP';                                                        ","\n",
                            "        BEGIN                                                                                                                                            ","\n",
                            "          EXECUTE IMMEDIATE ls_ExecScript;                                                                                                               ","\n",
                            "        EXCEPTION                                                                                                                                        ","\n",
                            "          WHEN OTHERS THEN                                                                                                                               ","\n",
                            "            DBMS_OUTPUT.PUT_LINE(cs_ERROR_tag||ls_ExecScript)  ;                                                                                         ","\n",
                            "            DBMS_OUTPUT.PUT_LINE(SQLERRM);                                                                                                               ","\n",
                            "        END;                                                                                                                                             ","\n",
                            "                                                                                                                                                         ","\n",
                            "      ELSIF ls_table_type = 'S' THEN                                                                                                                     ","\n",
                            "        ls_ExecScript := 'CREATE SYNONYM '||ls_table_name||' FOR '||ls_master_schema||'.'||ls_table_name;                                                ","\n",
                            "        BEGIN                                                                                                                                            ","\n",
                            "          EXECUTE IMMEDIATE ls_ExecScript;                                                                                                               ","\n",
                            "        EXCEPTION                                                                                                                                        ","\n",
                            "          WHEN OTHERS THEN                                                                                                                               ","\n",
                            "            DBMS_OUTPUT.PUT_LINE(cs_ERROR_tag||ls_ExecScript)  ;                                                                                         ","\n",
                            "            DBMS_OUTPUT.PUT_LINE(SQLERRM);                                                                                                               ","\n",
                            "        END;                                                                                                                                             ","\n",
                            "      END IF;                                                                                                                                            ","\n",
                            "    END IF;                                                                                                                                              ","\n",
                            "  END IF;                                                                                                                                                ","\n",
                            "                                                                                                                                                         ","\n",
                            "END;                                                                                                                                                     ","\n",
                            "/                                                                                                                                                        ","\n"
  LET ls_db_type = sadzi140_util_get_db_type() #20240521
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb #20240521
      LET ls_procedure_script = "--Create non exist table or synonymous PLpgSQL procedure.","\n"
    WHEN cs_db_driver_type_dmg --20230804
      -- Using Oracle Procedure scripts
  OTHERWISE
    -- Using Oracle Procedure scripts
  END CASE
  &endif 
  &endif 
                            
  LET ls_return = ls_procedure_script,"\n"
  
  RETURN ls_return

END FUNCTION

FUNCTION sadzi140_db_get_update_foreign_key_data(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name       STRING,
  ls_procedure_script STRING,
  ls_separator        STRING,
  ls_db_type          STRING
DEFINE
  ls_return STRING  
  
  LET ls_table_name = p_table_name.toLowerCase()

  LET ls_separator = os.Path.separator()

  &ifdef is_msv
  LET ls_procedure_script = "--Update foreign key data T-SQL procedure. ","\n"
  &else
  &ifdef is_pgs
  LET ls_procedure_script = "--Update foreign key data PLpgSQL procedure. ","\n"
  &else
  LET ls_procedure_script = "-- Update foreign key data P/L SQL procedure.                                                                                                                                                         ","\n",
                            "SET SERVEROUTPUT ON                                                                                                                                                                                   ","\n",
                            "DECLARE                                                                                                                                                                                               ","\n",
                            "  -----------------------------------------------------------------------------                                                                                                                       ","\n",
                            "  cs_ERROR_TAG    CONSTANT VARCHAR2(20) := '",cs_error_tag,"';                                                                                                                                        ","\n",
                            "  cs_WARNING_TAG  CONSTANT VARCHAR2(20) := '",cs_warning_tag,"';                                                                                                                                      ","\n",
                            "  cs_SUCCESS_TAG  CONSTANT VARCHAR2(20) := '",cs_success_tag,"';                                                                                                                                      ","\n",
                            "  -----------------------------------------------------------------------------                                                                                                                       ","\n",
                            "  ls_ExecScript        varchar2(32767);                                                                                                                                                               ","\n",  
                            "  ls_user              varchar2(20);                                                                                                                                                                  ","\n",  
                            "  ls_table_name        varchar2(30);                                                                                                                                                                  ","\n",
                            "  lb_false             boolean;                                                                                                                                                                       ","\n",
                            "  ls_r_constraint_name varchar2(60);                                                                                                                                                                  ","\n",
                            "BEGIN                                                                                                                                                                                                 ","\n",
                            "                                                                                                                                                                                                      ","\n",
                            "  ls_table_name := '",ls_table_name,"';                                                                                                                                                               ","\n",
                            "                                                                                                                                                                                                      ","\n",
                            "  SELECT USER                                                                                                                                                                                         ","\n",
                            "    INTO ls_user                                                                                                                                                                                      ","\n",
                            "    FROM DUAL;                                                                                                                                                                                        ","\n",
                            "                                                                                                                                                                                                      ","\n",
                            "  lb_false := FALSE;                                                                                                                                                                                  ","\n",
                            "  ls_r_constraint_name := NULL;                                                                                                                                                                       ","\n",
                            "                                                                                                                                                                                                      ","\n",
                            "  BEGIN                                                                                                                                                                                               ","\n",
                            "    SELECT ACS.R_CONSTRAINT_NAME                                                                                                                                                                      ","\n",
                            "      INTO ls_r_constraint_name                                                                                                                                                                       ","\n",
                            "      FROM ALL_CONSTRAINTS ACS                                                                                                                                                                        ","\n",
                            "     WHERE EXISTS (                                                                                                                                                                                   ","\n",
                            "                    SELECT 1                                                                                                                                                                          ","\n",
                            "                      FROM DZED_T ED                                                                                                                                                                  ","\n",
                            "                     WHERE ED.DZED001 = LOWER(ACS.TABLE_NAME)                                                                                                                                         ","\n",
                            "                       AND ED.DZED002 = LOWER(ACS.CONSTRAINT_NAME)                                                                                                                                    ","\n",
                            "                       AND ED.DZED003 = DECODE(ACS.CONSTRAINT_TYPE, 'R', 'R', ACS.CONSTRAINT_TYPE)                                                                                                    ","\n",
                            "                       AND ED.DZED003 = 'R'                                                                                                                                                           ","\n",
                            "                       AND ED.DZED001 = lower(ls_table_name)                                                                                                                                          ","\n",
                            "                  )                                                                                                                                                                                   ","\n",
                            "       AND ACS.OWNER = UPPER(ls_user);                                                                                                                                                                ","\n",
                            "  EXCEPTION                                                                                                                                                                                           ","\n",
                            "    WHEN OTHERS THEN                                                                                                                                                                                  ","\n",
                            "      DBMS_OUTPUT.put_line(cs_WARNING_TAG||SQLERRM);                                                                                                                                                  ","\n",
                            "      lb_false := TRUE;                                                                                                                                                                               ","\n",
                            "  END;                                                                                                                                                                                                ","\n",
                            "                                                                                                                                                                                                      ","\n",
                            "  if (not lb_false) and (ls_r_constraint_name is not null) then                                                                                                                                       ","\n",
                            "    BEGIN                                                                                                                                                                                             ","\n",
                            "      ls_ExecScript := 'UPDATE DZED_T ED                                                                                                                                                         '||  ","\n",
                            "                       '   SET (                                                                                                                                                                 '||  ","\n",
                            "                       '         ED.DZED004,                                                                                                                                                     '||  ","\n",
                            "                       '         ED.DZED005,                                                                                                                                                     '||  ","\n",
                            "                       '         ED.DZED006                                                                                                                                                      '||  ","\n",
                            "                       '       ) = (                                                                                                                                                             '||  ","\n",
                            "                       '            SELECT LOWER(ACC.COLUMN_NAMES),                                                                                                                              '||  ","\n",
                            "                       '                   LOWER(ACCR.TABLE_NAME),                                                                                                                               '||  ","\n",
                            "                       '                   LOWER(ACCR.COLUMN_NAMES)                                                                                                                              '||  ","\n",
                            "                       '              FROM ALL_CONSTRAINTS ACS                                                                                                                                   '||  ","\n",
                            "                       '              LEFT OUTER JOIN (                                                                                                                                          '||  ","\n",
                            "                       '                               SELECT ACCS.CONSTRAINT_NAME,ACCS.OWNER,ACCS.TABLE_NAME,LISTAGG(ACCS.COLUMN_NAME, '','') WITHIN GROUP(ORDER BY ACCS.POSITION) COLUMN_NAMES '||  ","\n",
                            "                       '                                 FROM ALL_CONS_COLUMNS ACCS                                                                                                              '||  ","\n",
                            "                       '                                GROUP BY ACCS.CONSTRAINT_NAME,ACCS.OWNER,ACCS.TABLE_NAME                                                                                 '||  ","\n",
                            "                       '                              ) ACC ON ACC.CONSTRAINT_NAME = ACS.CONSTRAINT_NAME                                                                                         '||  ","\n",
                            "                       '                                   AND ACC.OWNER           = ACS.OWNER                                                                                                   '||  ","\n",
                            "                       '              LEFT OUTER JOIN (                                                                                                                                          '||  ","\n",
                            "                       '                               SELECT ACCS.CONSTRAINT_NAME,ACCS.OWNER,ACCS.TABLE_NAME,LISTAGG(ACCS.COLUMN_NAME, '','') WITHIN GROUP(ORDER BY ACCS.POSITION) COLUMN_NAMES '||  ","\n",
                            "                       '                                 FROM ALL_CONS_COLUMNS ACCS                                                                                                              '||  ","\n",
                            "                       '                                GROUP BY ACCS.CONSTRAINT_NAME,ACCS.OWNER,ACCS.TABLE_NAME                                                                                 '||  ","\n",
                            "                       '                              ) ACCR ON ACCR.CONSTRAINT_NAME = ACS.R_CONSTRAINT_NAME                                                                                     '||  ","\n",
                            "                       '                                    AND ACCR.OWNER           = ACS.R_OWNER                                                                                               '||  ","\n",
                            "                       '             WHERE 1 = 1                                                                                                                                                 '||  ","\n",
                            "                       '               AND ACS.TABLE_NAME = UPPER(ED.DZED001)                                                                                                                    '||  ","\n",
                            "                       '               AND ACS.CONSTRAINT_NAME = UPPER(ED.DZED002)                                                                                                               '||  ","\n",
                            "                       '               AND DECODE(ACS.CONSTRAINT_TYPE, ''R'', ''R'', ACS.CONSTRAINT_TYPE) = ED.DZED003                                                                           '||  ","\n",
                            "                       '               AND ACS.OWNER      = UPPER('''||ls_user||''')                                                                                                             '||  ","\n",
                            "                       '           )                                                                                                                                                             '||  ","\n",
                            "                       ' WHERE ED.DZED001 = LOWER('''||ls_table_name||''')                                                                                                                       '||  ","\n",
                            "                       '   AND ED.DZED003 = ''R''                                                                                                                                                ';   ","\n",
                            "                                                                                                                                                                                                      ","\n",
                            "      EXECUTE IMMEDIATE ls_ExecScript;                                                                                                                                                                ","\n",
                            "      COMMIT;                                                                                                                                                                                         ","\n",
                            "    EXCEPTION                                                                                                                                                                                         ","\n",
                            "      WHEN OTHERS THEN ROLLBACK;                                                                                                                                                                      ","\n",
                            "      DBMS_OUTPUT.put_line(cs_ERROR_TAG||SQLERRM);                                                                                                                                                    ","\n",
                            "    END;                                                                                                                                                                                              ","\n",
                            "  end if;                                                                                                                                                                                             ","\n",
                            "                                                                                                                                                                                                      ","\n",
                            "END;                                                                                                                                                                                                  ","\n",
                            "/                                                                                                                                                                                                     ","\n"
  LET ls_db_type = sadzi140_util_get_db_type() #20240521
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb #20240521
      LET ls_procedure_script = "--Update foreign key data PLpgSQL procedure.","\n"
    WHEN cs_db_driver_type_dmg --20230804
      -- Using Oracle Procedure scripts
  OTHERWISE
    -- Using Oracle Procedure scripts
  END CASE
  &endif
  &endif
                            
  LET ls_return = ls_procedure_script
  
  RETURN ls_return

END FUNCTION

FUNCTION sadzi140_db_get_create_dummy_table_sql(p_create_table)
DEFINE 
  p_create_table  T_DZEA_CREATE_TABLE
DEFINE
  lo_create_table T_DZEA_CREATE_TABLE,
  ls_sql          STRING,
  ls_table_space  STRING,
  ls_comment      STRING,
  ls_return       STRING,
  ls_db_type      STRING #20240521

  LET lo_create_table.* = p_create_table.* 

  LET ls_sql = ""
  
  LET ls_table_space = ""
  LET ls_comment     = ""

  &ifdef is_msv
    LET ls_sql = "select * into ",lo_create_table.dct_table_name," from dummy where 1=2","\n",
                 "go","\n"
                 
    LET ls_comment = "declare @user sysname                                           ","\n",
                     "select @user = user_name()                                      ","\n",
                     "begin try                                                       ","\n",
                     "execute sp_addextendedproperty 'comment',                       ","\n",
                     "   '",lo_create_table.dct_table_description,"',                 ","\n",  
                     "   'user', @user, 'table', '",lo_create_table.dct_table_name,"' ","\n",
                     "end try                                                         ","\n",
                     "begin catch                                                     ","\n",
                     "execute sp_updateextendedproperty 'comment',                    ","\n",
                     "   '",lo_create_table.dct_table_description,"',                 ","\n",
                     "   'user', @user, 'table', '",lo_create_table.dct_table_name,"' ","\n", 
                     "end catch                                                       ","\n", 
                     "go                                                              ","\n"
  &else
  &ifdef is_pgs
    LET ls_sql = "create table"," ",
                 lo_create_table.dct_table_name," ",
                 ls_table_space," ", 
                 "as select * from dummy where 1=2",
                 ";"

    #Table的說明             
    IF NOT NVL(lo_create_table.dct_table_description,cs_null_value) = cs_null_value THEN
      LET ls_comment = "comment on table ",lo_create_table.dct_table_name,
                       "  is '",lo_create_table.dct_table_description,"'",
                       ";"
    END IF
  &else

    LET ls_db_type = sadzi140_util_get_db_type() #20240521
    CASE ls_db_type
      WHEN cs_db_driver_type_kdb #20240521
        LET ls_sql = "create table"," ",
                     lo_create_table.dct_table_name," ",
                     ls_table_space," ", 
                     "as select * from dummy where 1=2",
                     ";"

        #Table的說明             
        IF NOT NVL(lo_create_table.dct_table_description,cs_null_value) = cs_null_value THEN
          LET ls_comment = "comment on table ",lo_create_table.dct_table_name,
                           "  is '",lo_create_table.dct_table_description,"'",
                           ";"
        END IF
        
    --WHEN cs_db_driver_type_dmg #20240521
        --與 Oracle 共用
    OTHERWISE   
      #指定 Tablespace 
      IF NVL(lo_create_table.dct_table_space CLIPPED,cs_null_value) <> cs_null_value AND lo_create_table.dct_table_space CLIPPED <> ""
         AND lo_create_table.dct_table_space CLIPPED <> " " THEN
        LET ls_table_space = "tablespace ",
                            lo_create_table.dct_table_space
      END IF

      LET ls_sql = "create table"," ",
                   lo_create_table.dct_table_name," ",
                   ls_table_space," ", 
                   "as select * from dummy where 1=2",
                   ";"

      #Table的說明             
      IF NOT NVL(lo_create_table.dct_table_description,cs_null_value) = cs_null_value THEN
        LET ls_comment = "comment on table ",lo_create_table.dct_table_name,
                         " is '",lo_create_table.dct_table_description,"'",
                         ";"
      END IF
    END CASE  
    
  &endif
  &endif

  LET ls_return = ls_sql,"\n",
                  ls_comment 
 
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_get_create_synonym_view_sql(p_db_connstr,p_create_table)
DEFINE 
  p_db_connstr    T_DB_CONNSTR,
  p_create_table  T_CREATE_SYNONYM_VIEW
DEFINE
  lo_db_connstr    T_DB_CONNSTR,
  lo_create_table  T_CREATE_SYNONYM_VIEW,
  ls_sql           STRING,
  ls_return        STRING

  LET lo_create_table.* = p_create_table.*
  LET lo_db_connstr.* = p_db_connstr.*

  #FOR li_loop = 1 TO lo_SchemaType.getLength()
  #  IF lo_create_table.
  #END FOR
  
  LET ls_sql = ""

  &ifdef is_msv
  LET ls_sql = "create or replace synonym ",
               lo_create_table.csv_table_name,
               " for ",
               lo_create_table.csv_master_user,".dbo.",lo_create_table.csv_table_name,"\n",
               "go"
  &else
  &ifdef is_pgs
  CASE sadzi140_util_get_pgs_branch_type()
    WHEN cs_pgs_sub_mogdb
      LET ls_sql = "create or replace synonym ",
                   lo_db_connstr.db_username,".",lo_create_table.csv_table_name,
                   " for ",
                   lo_create_table.csv_master_user,".",lo_create_table.csv_table_name,
                   ";"
    WHEN cs_pgs_sub_polardb
      {待建構}
  OTHERWISE       
    --PostgreSQL using view
    LET ls_sql = "drop view if exists ",lo_create_table.csv_table_name,";","\n",
                 "create or replace view ",lo_create_table.csv_table_name," as select * from ", lo_create_table.csv_master_user,".",lo_create_table.csv_table_name,";","\n",
                 "grant select,insert,delete,update on  ",lo_create_table.csv_table_name," to tiptop;","\n"
  END CASE 
  &else
  LET ls_sql = "create or replace synonym ",
               lo_db_connstr.db_username,".",lo_create_table.csv_table_name,
               " for ",
               lo_create_table.csv_master_user,".",lo_create_table.csv_table_name,
               ";"
  &endif  
  &endif  

  LET ls_return = ls_sql
 
  RETURN ls_return
  
END FUNCTION

#20200330 begin
FUNCTION sadzi140_db_get_drop_synonym_view_sql(p_db_connstr,p_create_table)
DEFINE 
  p_db_connstr   T_DB_CONNSTR,
  p_create_table T_CREATE_SYNONYM_VIEW
DEFINE
  lo_db_connstr   T_DB_CONNSTR,
  lo_create_table T_CREATE_SYNONYM_VIEW,
  ls_sql          STRING,
  ls_return       STRING

  LET lo_create_table.* = p_create_table.*
  LET lo_db_connstr.* = p_db_connstr.*

  #FOR li_loop = 1 TO lo_SchemaType.getLength()
  #  IF lo_create_table.
  #END FOR
  
  LET ls_sql = ""

  &ifdef is_msv
  LET ls_sql = "drop synonym ",
               lo_create_table.csv_table_name,"\n",
               "go","\n"
  &else
  &ifdef is_pgs
  CASE sadzi140_util_get_pgs_branch_type()
    WHEN cs_pgs_sub_mogdb
  LET ls_sql = "drop synonym ",
               lo_db_connstr.db_username,".",lo_create_table.csv_table_name,
               ";"
    WHEN cs_pgs_sub_polardb
      {待建構}
  OTHERWISE       
    --PostgreSQL using view
    LET ls_sql = "drop view if exists ",lo_create_table.csv_table_name,";","\n"
  END CASE 
  
  &else
  LET ls_sql = "drop synonym ",
               lo_db_connstr.db_username,".",lo_create_table.csv_table_name,
               ";"
  &endif  
  &endif  

  LET ls_return = ls_sql
 
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_get_backup_and_drop_depended_view_sql(p_source_object_info)
DEFINE
  p_source_object_info  T_SOURCE_OBJECT_INFO
DEFINE
  lo_source_object_info  T_SOURCE_OBJECT_INFO,
  ls_sql     STRING,
  ls_return  STRING

  LET lo_source_object_info.* = p_source_object_info.*

  #LET ls_sql = "select t100_deps_save_and_drop_dependencies('",lo_source_object_info.soi_schema,"'::varchar,'",lo_source_object_info.soi_object_name,"'::varchar);","\n"
  LET ls_sql = "do $proc$                                                                                                         ","\n",
               "declare                                                                                                           ","\n",
               "  -------------------------------------------------------------------------------                                 ","\n", 
               "  ls_schema    varchar(128) := '",lo_source_object_info.soi_schema,"';                                            ","\n", 
               "  ls_table     varchar(128) := '",lo_source_object_info.soi_object_name,"';                                       ","\n",  
               "  ls_prc_guid  varchar(128) := '",lo_source_object_info.soi_prc_guid,"';                                          ","\n", 
               "  -------------------------------------------------------------------------------                                 ","\n",
               "  c_deps_list refcursor;                                                                                          ","\n",
               "  ls_deps_list  varchar(4096);                                                                                    ","\n",
               "  cs_deps_list  varchar(4096) :=                                                                                  ","\n",
               "    'select dvd.dependent_schema,dvd.dependent_view,                                                              ","\n",
               "            case                                                                                                  ","\n",
               "              when source_schema = ''",cs_master_db,"'' and dependent_schema <> source_schema then                ","\n",
               "                ''create or replace view ''||dependent_view||                                                     ","\n",
               "                '' as select * from ''||source_schema||''.''||dependent_view||'';''                               ","\n",
               "            else                                                                                                  ","\n",
               "              dvd.scripts                                                                                         ","\n",
               "            end scripts,                                                                                          ","\n",
               "            dvd.source_schema,dvd.source_table                                                                    ","\n",
               "       from dba_view_depend dvd                                                                                   ","\n",
               "      where dvd.dependent_schema = ''{p_schema}''                                                                 ","\n",
               "        and dvd.source_table  = ''{p_table}''';                                                                   ","\n",
               "  r_deps_list RECORD;                                                                                             ","\n",
               "  -------------------------------------------------------------------------------                                 ","\n",
               "  cs_error_tag       varchar(15) := '[ERROR]';                                                                    ","\n",
               "  cs_success_tag     varchar(15) := '[SUCCESS]';                                                                  ","\n",
               "  cs_warning_tag     varchar(15) := '[WARNING]';                                                                  ","\n",
               "  cs_message_tag     varchar(15) := '[MESSAGE]';                                                                  ","\n",
               "  -------------------------------------------------------------------------------                                 ","\n",
               "  lb_error  boolean;                                                                                              ","\n",
               "  lb_deps   boolean;                                                                                              ","\n",
               "  lb_drop   boolean;                                                                                              ","\n",
               "  ls_sql    varchar(65535);                                                                                       ","\n",
               "begin                                                                                                             ","\n",
               "                                                                                                                  ","\n",
               "  lb_error := false;                                                                                              ","\n",
               "                                                                                                                  ","\n",
               "  select replace(replace(cs_deps_list,'{p_schema}',ls_schema),'{p_table}',ls_table) into ls_deps_list;            ","\n",
               "  open c_deps_list for execute ls_deps_list;                                                                      ","\n",
               "  loop                                                                                                            ","\n",
               "    fetch c_deps_list into r_deps_list;                                                                           ","\n",
               "    exit when not found;                                                                                          ","\n",
               "                                                                                                                  ","\n",
               "    lb_deps := true;                                                                                              ","\n",
               "                                                                                                                  ","\n",
               "    begin                                                                                                         ","\n",
               "      ls_sql := 'insert into t100_deps_saved_ddl                                                                  ","\n",
               "                 (deps_view_schema, deps_view_name, deps_ddl_to_run,  sorc_view_schema, sorc_view_name,prc_guid)  ","\n",
               "                 values                                                                                           ","\n",
               "                 ($1,$2,$3,$4,$5,$6)';                                                                            ","\n",
               "      --raise notice '% %',cs_message_tag,'';                                                                     ","\n",
               "      execute ls_sql                                                                                              ","\n",
               "      using r_deps_list.dependent_schema,r_deps_list.dependent_view,r_deps_list.scripts,                          ","\n",
               "            r_deps_list.source_schema,r_deps_list.source_table,ls_prc_guid;                                       ","\n",        
               "    exception                                                                                                     ","\n",
               "      when others then                                                                                            ","\n",
               "        raise notice '% %',cs_error_tag,ls_sql;                                                                   ","\n",
               "        lb_deps := false;                                                                                         ","\n",
               "        raise;                                                                                                    ","\n",
               "    end;                                                                                                          ","\n",
               "                                                                                                                  ","\n",
               "    if lb_deps then                                                                                               ","\n",
               "      lb_drop := true;                                                                                            ","\n",
               "      begin                                                                                                       ","\n",
               "        ls_sql := 'drop view '||r_deps_list.dependent_schema||'.'||r_deps_list.dependent_view;                    ","\n",
               "        execute ls_sql;                                                                                           ","\n",
               "      exception                                                                                                   ","\n",
               "        when others then                                                                                          ","\n",
               "          raise notice '% %',cs_error_tag,ls_sql;                                                                 ","\n",
               "          raise;                                                                                                  ","\n",
               "      end;                                                                                                        ","\n",
               "    end if;                                                                                                       ","\n",
               "                                                                                                                  ","\n",
               "  end loop;                                                                                                       ","\n",
               "  close c_deps_list;                                                                                              ","\n",
               "                                                                                                                  ","\n",
               "exception                                                                                                         ","\n",
               "  when others then                                                                                                ","\n",
               "    raise notice '% %',cs_error_tag,SQLERRM;                                                                      ","\n",
               "end $proc$;                                                                                                       ","\n"

  LET ls_return = ls_sql
 
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_get_restore_depended_view_sql(p_source_object_info)
DEFINE
  p_source_object_info  T_SOURCE_OBJECT_INFO
DEFINE
  lo_source_object_info  T_SOURCE_OBJECT_INFO,
  ls_sql     STRING,
  ls_return  STRING

  LET lo_source_object_info.* = p_source_object_info.*

  --LET ls_sql = "select t100_deps_restore_dependencies('",lo_source_object_info.soi_schema,"'::varchar,'",lo_source_object_info.soi_object_name,"'::varchar);","\n"
  LET ls_sql = "do $proc$                                                                                   ","\n", 
               "declare                                                                                     ","\n", 
               "  -------------------------------------------------------------------------------           ","\n", 
               "  ls_prc_guid  varchar(128) := '",lo_source_object_info.soi_prc_guid,"';                    ","\n", 
               "  ls_schema    varchar(128) := '",lo_source_object_info.soi_schema,"';                      ","\n", 
               "  -------------------------------------------------------------------------------           ","\n", 
               "  c_ddls_list refcursor;                                                                    ","\n", 
               "  ls_ddls_list  varchar(4096);                                                              ","\n", 
               "  cs_ddls_list  varchar(4096) :=                                                            ","\n", 
               "    'select tds.deps_view_schema, tds.deps_view_name, tds.deps_ddl_to_run,                  ","\n", 
               "            tds.sorc_view_schema, tds.sorc_view_name, tds.prc_guid                          ","\n", 
               "       from t100_deps_saved_ddl tds                                                         ","\n", 
               "      where tds.prc_guid = ''{p_prc_guid}''                                                 ","\n", 
               "        and tds.deps_view_schema = ''{p_schema}''                                           ","\n",
               "      order by length(sorc_view_schema)';                                                   ","\n", 
               "  r_ddls_list RECORD;                                                                       ","\n", 
               "  -------------------------------------------------------------------------------           ","\n", 
               "  cs_error_tag       varchar(15) := '[ERROR]';                                              ","\n", 
               "  cs_success_tag     varchar(15) := '[SUCCESS]';                                            ","\n", 
               "  cs_warning_tag     varchar(15) := '[WARNING]';                                            ","\n", 
               "  cs_message_tag     varchar(15) := '[MESSAGE]';                                            ","\n", 
               "  -------------------------------------------------------------------------------           ","\n", 
               "  lb_error  boolean;                                                                        ","\n", 
               "  lb_deps   boolean;                                                                        ","\n", 
               "  lb_drop   boolean;                                                                        ","\n", 
               "  ls_sql    varchar(65535);                                                                 ","\n", 
               "begin                                                                                       ","\n", 
               "                                                                                            ","\n", 
               "  lb_error := false;                                                                        ","\n", 
               "                                                                                            ","\n", 
               "  select replace(replace(cs_ddls_list,'{p_prc_guid}',ls_prc_guid),'{p_schema}',ls_schema)   ","\n", 
               "    into ls_ddls_list;                                                                      ","\n", 
               "                                                                                            ","\n", 
               "  open c_ddls_list for execute ls_ddls_list;                                                ","\n", 
               "  loop                                                                                      ","\n", 
               "    fetch c_ddls_list into r_ddls_list;                                                     ","\n", 
               "    exit when not found;                                                                    ","\n", 
               "                                                                                            ","\n", 
               "    lb_deps := true;                                                                        ","\n", 
               "                                                                                            ","\n", 
               "    begin                                                                                   ","\n", 
               "      ls_sql := r_ddls_list.deps_ddl_to_run;                                                ","\n", 
               "      --raise notice '% %',cs_message_tag,'';                                               ","\n", 
               "      execute ls_sql;                                                                       ","\n", 
               "    exception                                                                               ","\n", 
               "      when others then                                                                      ","\n",
               "        if lower(r_ddls_list.deps_ddl_to_run) like '%dummy%' then                           ","\n",
               "          raise notice '% %',cs_warning_tag,ls_sql;                                         ","\n",
               "        else                                                                                ","\n",
               "          raise notice '% %',cs_error_tag,ls_sql;                                           ","\n",
               "        end if;                                                                             ","\n",
               "        lb_deps := false;                                                                   ","\n", 
               "        raise;                                                                              ","\n", 
               "    end;                                                                                    ","\n", 
               "                                                                                            ","\n", 
               "    if lb_deps then                                                                         ","\n", 
               "      lb_drop := true;                                                                      ","\n", 
               "      begin                                                                                 ","\n", 
               "        ls_sql := 'delete from t100_deps_saved_ddl                                          ","\n", 
               "                    where prc_guid = $1                                                     ","\n", 
               "                      and deps_view_schema = $2                                             ","\n", 
               "                      and deps_view_name = $3 ';                                            ","\n", 
               "        execute ls_sql                                                                      ","\n", 
               "        using r_ddls_list.prc_guid,r_ddls_list.deps_view_schema,r_ddls_list.deps_view_name; ","\n", 
               "      exception                                                                             ","\n", 
               "        when others then                                                                    ","\n", 
               "          raise notice '% %',cs_error_tag,ls_sql;                                           ","\n", 
               "          raise;                                                                            ","\n", 
               "      end;                                                                                  ","\n", 
               "    end if;                                                                                 ","\n", 
               "                                                                                            ","\n", 
               "  end loop;                                                                                 ","\n", 
               "  close c_ddls_list;                                                                        ","\n", 
               "                                                                                            ","\n", 
               "exception                                                                                   ","\n", 
               "  when others then                                                                          ","\n",
               "    if lower(SQLERRM) like '%dummy%' then                                                   ","\n",
               "      raise notice '% %',cs_warning_tag,SQLERRM;                                            ","\n",
               "    else                                                                                    ","\n",
               "      raise notice '% %',cs_error_tag,SQLERRM;                                              ","\n",
               "    end if;                                                                                 ","\n", 
               "end $proc$;                                                                                 ","\n" 

  LET ls_return = ls_sql
 
  RETURN ls_return
  
END FUNCTION
#20200330 end

FUNCTION sadzi140_db_get_alter_constraint_sql(p_owner,p_table_name,p_diff_constraint)
DEFINE 
  p_owner           STRING,
  p_table_name      STRING,
  p_diff_constraint DYNAMIC ARRAY OF T_CONSTRAINT_DIFF 
DEFINE
  ls_owner            STRING,
  ls_table_name       STRING,
  lo_diff_constraint  DYNAMIC ARRAY OF T_CONSTRAINT_DIFF,
  lo_alter_constraint DYNAMIC ARRAY OF T_CONSTRAINT_DIFF,
  li_loop             INTEGER, 
  li_rec_cnt          INTEGER,
  ls_constraints      STRING,
  ls_sql              STRING,
  ls_where            STRING,
  ls_line             STRING,
  ls_db_type          STRING, #20240521
  ls_drop_null_sql    STRING, #20200330
  lb_result           BOOLEAN, #20200330
  lb_constraint_exist BOOLEAN,
  ls_key_type         STRING,
  ls_references       STRING,
  ls_enable_foreign_key STRING,
  ls_foreign_key_state  STRING,
  ls_key_null_default_value STRING,
  lo_string_buffer    base.StringBuffer
DEFINE  
  ls_return STRING

  LET ls_owner           = p_owner
  LET ls_table_name      = p_table_name
  LET lo_diff_constraint = p_diff_constraint

  LET lo_string_buffer = base.StringBuffer.create()

  FOR li_loop = 1 TO lo_diff_constraint.getLength()
    LET ls_constraints = ls_constraints,",'",lo_diff_constraint[li_loop].dzed002,"'"
  END FOR 
  LET ls_constraints = ls_constraints.subString(2,ls_constraints.getLength())

  LET ls_sql = "select ed.dzed001        dzed001,                      ",
               "       ed.dzed002        dzed002,                      ",
               "       upper(ed.dzed003) dzed003,                      ",
               "       ed.dzed004        dzed004,                      ",
               "       ed.dzed005        dzed005,                      ",
               "       ed.dzed006        dzed006                       ",
               "  from dzed_t ed                                       ",
               " where ed.dzed001 = '",ls_table_name.toLowerCase(),"'  ",
               "   and ed.dzed002 in (",ls_constraints.toLowerCase(),")",
               "   and ed.dzed003 <> 'F'                               "

  PREPARE lpre_get_alter_constraint_sql FROM ls_sql
  DECLARE lcur_get_alter_constraint_sql CURSOR FOR lpre_get_alter_constraint_sql

  LET li_rec_cnt = 1
  
  OPEN lcur_get_alter_constraint_sql
  FOREACH lcur_get_alter_constraint_sql INTO lo_alter_constraint[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET ls_Line = ""
    &ifdef is_msv
    -- SQL Server 專用
    LET ls_Line = ls_Line,
                  sadzi140_db_get_drop_index_before_add_pk(lo_alter_constraint[li_rec_cnt].dzed001,lo_alter_constraint[li_rec_cnt].dzed004)
    &endif            

    IF NVL(lo_alter_constraint[li_rec_cnt].DZED003,cs_null_value) = "P" THEN
      LET ls_key_type = "primary key";
      LET ls_references = ""
    END IF
    IF NVL(lo_alter_constraint[li_rec_cnt].DZED003,cs_null_value) = "U" THEN
      LET ls_key_type = "unique";
      LET ls_references = ""
    END IF
    IF NVL(lo_alter_constraint[li_rec_cnt].DZED003,cs_null_value) = "R" THEN
      #取得 Foreign 是否啟動
      CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_enable_foreign_key) RETURNING ls_enable_foreign_key
      LET ls_foreign_key_state = IIF(ls_enable_foreign_key=="Y",cs_state_enable,cs_state_disable)
      LET ls_key_type = "foreign key";
      LET ls_references = "references ",lo_alter_constraint[li_rec_cnt].dzed005," (",lo_alter_constraint[li_rec_cnt].dzed006,") on delete cascade ",ls_foreign_key_state
    END IF

    #Example
    #Add
    --alter table ITMML_T add constraint itmml_pk primary key (ITMMLENT, ITMML001);
    --alter table ITMML_T add constraint itmml_fk foreign key (ITMMLENT) references itmm_t (ITMMENT) ON DELETE CASCADE DISABLE;
    --alter table ITMML_T add constraint itmml_uk unique (ITMML003);    
    #Drop
    --alter table ITMML_T drop constraint ITMML_PK CASCADE;
    --alter table ITMML_T drop constraint ITMML_FK;
  
    CALL sadzi140_db_check_table_constraint_exist(ls_owner,lo_alter_constraint[li_rec_cnt].dzed001,lo_alter_constraint[li_rec_cnt].dzed002) RETURNING lb_constraint_exist
    IF lb_constraint_exist THEN
      IF NVL(lo_alter_constraint[li_rec_cnt].DZED003,cs_null_value) = "P" THEN
        &ifdef is_msv
        LET ls_Line = ls_Line,
                      "-- MS SQL Server Drop primary key.","\n",
                      "if object_id('dbo.[",lo_alter_constraint[li_rec_cnt].DZED002,"]') is not null",
                      " alter table ",lo_alter_constraint[li_rec_cnt].dzed001," drop constraint ",lo_alter_constraint[li_rec_cnt].DZED002,"\n","go","\n"
        &else
        &ifdef is_pgs
        LET ls_drop_null_sql = sadzi140_db_get_drop_not_null_sql(ls_owner,ls_table_name,lo_alter_constraint[li_rec_cnt].DZED002) 
        LET ls_Line = ls_Line,
                      "-- PostgreSQL Drop primary key 1.","\n",
                      "alter table ",lo_alter_constraint[li_rec_cnt].dzed001," drop constraint ",lo_alter_constraint[li_rec_cnt].DZED002,";","\n",
                      ls_drop_null_sql,"\n"
        &else
        LET ls_db_type = sadzi140_util_get_db_type() #20240521
        CASE ls_db_type
          WHEN cs_db_driver_type_kdb --20240521
            LET ls_drop_null_sql = sadzi140_db_get_drop_not_null_sql(ls_owner,ls_table_name,lo_alter_constraint[li_rec_cnt].DZED002) 
            LET ls_Line = ls_Line,
                          "-- KingBASE Drop primary key.","\n",
                          "alter table ",lo_alter_constraint[li_rec_cnt].dzed001," drop constraint ",lo_alter_constraint[li_rec_cnt].DZED002,";","\n",
                          ls_drop_null_sql,"\n"
          WHEN cs_db_driver_type_dmg --20230804
            LET ls_Line = ls_Line,
                          "-- DM drop primary key [SEC2].","\n",
                          "alter table ",lo_alter_constraint[li_rec_cnt].dzed001," drop primary key cascade;","\n"
        OTHERWISE  
          LET ls_Line = ls_Line,
                        "-- Oracle Drop primary key [SEC2].","\n",
                        "alter table ",lo_alter_constraint[li_rec_cnt].dzed001," drop primary key cascade drop index;","\n"
        END CASE
        &endif
        &endif
      ELSE                     
        &ifdef is_msv
        LET ls_Line = ls_Line,
                      "-- MS SQL Server Drop constraint.","\n",
                      "if object_id('dbo.[",lo_alter_constraint[li_rec_cnt].DZED002,"]') is not null",
                      " alter table ",lo_alter_constraint[li_rec_cnt].dzed001," drop constraint ",lo_alter_constraint[li_rec_cnt].DZED002,"\n","go","\n"
        &else
        &ifdef is_pgs
        LET ls_drop_null_sql = sadzi140_db_get_drop_not_null_sql(ls_owner,ls_table_name,lo_alter_constraint[li_rec_cnt].DZED002) 
        LET ls_Line = ls_Line,
                      "-- PostgreSQL Drop constraint.","\n",
                      "alter table ",lo_alter_constraint[li_rec_cnt].dzed001," drop constraint ",lo_alter_constraint[li_rec_cnt].DZED002,";","\n",
                      ls_drop_null_sql,"\n"
        &else
        LET ls_db_type = sadzi140_util_get_db_type() #20240521
        CASE ls_db_type
          WHEN cs_db_driver_type_kdb #20240521
            LET ls_drop_null_sql = sadzi140_db_get_drop_not_null_sql(ls_owner,ls_table_name,lo_alter_constraint[li_rec_cnt].DZED002) 
            LET ls_Line = ls_Line,
                          "-- KDB Drop constraint.","\n",
                          "alter table ",lo_alter_constraint[li_rec_cnt].dzed001," drop constraint ",lo_alter_constraint[li_rec_cnt].DZED002,";","\n",
                          ls_drop_null_sql,"\n"
        --WHEN cs_db_driver_type_dmg #20240521
          --與Oracle共用
        OTHERWISE   
          LET ls_Line = ls_Line,
                        "-- Oracle Drop constraint.","\n",
                        "alter table ",lo_alter_constraint[li_rec_cnt].dzed001," drop constraint ",lo_alter_constraint[li_rec_cnt].DZED002," cascade;","\n"
        END CASE  
        &endif
        &endif
      END IF                     
    END IF                

    CALL sadzi140_db_get_key_null_default_value_list(ls_table_name) RETURNING ls_key_null_default_value 
    &ifdef is_msv
    LET ls_Line = ls_Line,
                  "-- Set Key field default value if key field data is null.","\n",
                  ls_key_null_default_value"\n",
                  "go","\n"
    &else
    &ifdef is_pgs
    LET ls_Line = ls_Line,
                  "-- Set Key field default value if key field data is null.","\n",
                  "alter table ",ls_table_name," disable trigger user",";","\n",  
                  ls_key_null_default_value,"\n",
                  "alter table ",ls_table_name," enable trigger user",";","\n"  
    &else
    LET ls_db_type = sadzi140_util_get_db_type() #20240521
    CASE ls_db_type
      WHEN cs_db_driver_type_kdb #20240521
        LET ls_Line = ls_Line,
                      "-- Set Key field default value if key field data is null.","\n",
                      "alter table ",ls_table_name," disable trigger user",";","\n",  
                      ls_key_null_default_value,"\n",
                      "alter table ",ls_table_name," enable trigger user",";","\n"  
    --WHEN cs_db_driver_type_dmg #20240521
      --與 Oracle 共用
    OTHERWISE   
      LET ls_Line = ls_Line,
                    "-- Set Key field default value if key field data is null.","\n",
                    ls_key_null_default_value,"\n"
    END CASE  
    &endif                       
    &endif                       

                  
    &ifdef is_msv
    &else
    --Oracle & PostgreSQL 共用
    LET ls_Line = ls_Line,"commit;","\n"
    &endif              
    
    &ifdef is_msv
    LET ls_Line = ls_Line,
                  "-- SQL Server alter table add constraint.","\n",
                  "alter table ",lo_alter_constraint[li_rec_cnt].dzed001,
                  " add constraint ",lo_alter_constraint[li_rec_cnt].dzed002," ",ls_key_type,
                  " (",lo_alter_constraint[li_rec_cnt].dzed004,") ",ls_references,"\n","go","\n"
                  
    LET ls_Line = ls_Line,
                  sadzi140_db_get_create_index_after_add_pk(lo_alter_constraint[li_rec_cnt].dzed001,lo_alter_constraint[li_rec_cnt].dzed004)
                  
    &else
    --Oracle & PostgreSQL 共用
    LET ls_Line = ls_Line,
                  "-- Oracle alter table add constraint.","\n",
                  "alter table ",lo_alter_constraint[li_rec_cnt].dzed001,
                  " add constraint ",lo_alter_constraint[li_rec_cnt].dzed002," ",ls_key_type,
                  " (",lo_alter_constraint[li_rec_cnt].dzed004,") ",ls_references,";","\n"
    &endif              
                  
    CALL lo_string_buffer.append(ls_Line)

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_alter_constraint_sql
  CALL lo_alter_constraint.deleteElement(li_rec_cnt)
  
  FREE lpre_get_alter_constraint_sql
  FREE lcur_get_alter_constraint_sql  

  LET ls_return = lo_string_buffer.toString()
 
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_get_drop_constraint_sql(p_owner,p_table_name,p_diff_constraint)
DEFINE 
  p_owner           STRING,
  p_table_name      STRING,
  p_diff_constraint DYNAMIC ARRAY OF T_CONSTRAINT_DIFF 
DEFINE
  ls_owner            STRING,
  ls_table_name       STRING,
  lo_diff_constraint  DYNAMIC ARRAY OF T_CONSTRAINT_DIFF,
  lo_drop_constraint  DYNAMIC ARRAY OF T_CONSTRAINT_DIFF,
  li_loop             INTEGER, 
  li_rec_cnt          INTEGER,
  ls_indexes          STRING,
  ls_sql              STRING,
  ls_where            STRING,
  ls_line             STRING,
  ls_drop_null_sql    STRING,
  ls_db_type          STRING, #20240521  
  lb_table_index_exist    BOOLEAN,
  lb_designer_index_exist BOOLEAN,
  ls_unique               STRING,
  ls_constraints          STRING,
  lo_string_buffer        base.StringBuffer
DEFINE  
  ls_return STRING

  LET ls_owner             = p_owner
  LET ls_table_name        = p_table_name
  LET lo_diff_constraint.* = p_diff_constraint.*

  LET lo_string_buffer = base.StringBuffer.create()

  FOR li_loop = 1 TO lo_diff_constraint.getLength()
    LET ls_constraints = ls_constraints,",'",lo_diff_constraint[li_loop].dzed002,"'"
  END FOR 
  LET ls_constraints = ls_constraints.subString(2,ls_constraints.getLength())

  &ifdef is_msv
  LET ls_sql = "select t.name                          table_name,                                 ", 
               "       i.name                          constraint_name,                            ", 
               "       iif(i.is_primary_key=1,'P','R') constraint_type,                            ", 
               "       string_agg(c.name,',') within group (order by ic.key_ordinal) column_names, ", 
               "       '' r_table_name,                                                            ", 
               "       '' r_column_names                                                           ", 
               "  from sys.indexes i                                                               ", 
               " inner join sys.index_columns ic                                                   ", 
               "         on i.index_id = ic.index_id and i.object_id = ic.object_id                ", 
               " inner join sys.tables as t                                                        ", 
               "         on t.object_id = i.object_id                                              ", 
               " inner join sys.columns c                                                          ", 
               "         on t.object_id = c.object_id and ic.column_id = c.column_id               ", 
               " where 1=1                                                                         ", 
               "   and t.name = '",ls_table_name.toLowerCase(),"'                                  ", 
               " group by t.name,i.name,iif(i.is_primary_key=1,'P','R')                            " 
  &else
  &ifdef is_pgs
  LET ls_sql = "select acs.table_name,                                                     ",                                                                               
               "       acs.constraint_name,                                                ",                                                                               
               "       case acs.constraint_type                                            ",
               "         when 'R' then 'R'                                                 ",
               "       else                                                                ",
               "         acs.constraint_type                                               ",
               "       end as constraint_type,                                             ",                           
               "       acc.column_names,                                                   ",                                                                               
               "       '' r_table_name,                                                    ",                                                                  
               "       '' r_column_names                                                   ",                                                                               
               "  from dba_constraints acs                                                 ",                                                                               
               "  left outer join (                                                        ",                                                                               
               "                   select accs.constraint_name,accs.owner,accs.table_name, ",
               "                          string_agg(accs.column_name, ','                 ",
               "                                     order by accs.position) column_names  ",
               "                     from dba_cons_columns accs                            ",                                                                               
               "                    group by accs.constraint_name,                         ",
               "                             accs.owner,accs.table_name                    ",                                                          
               "                  ) acc                                                    ",
               "               on acc.constraint_name = acs.constraint_name                ",                                                                      
               "              and acc.table_name      = acs.table_name                     ",                                                                             
               "              and acc.owner           = acs.owner                          ",                                                                      
               " where 1 = 1                                                               ",                                                                               
               "   and acs.table_name = '",ls_table_name.toLowerCase(),"'                  ",                                                                               
               "   and acs.owner      = '",ls_owner.toLowerCase(),"'                       ",                                                                               
               "   and acs.constraint_name in (",ls_constraints.toLowerCase(),")           "                                                                               
  &else
  LET ls_sql = "select acs.table_name,                                                             ", 
               "       acs.constraint_name,                                                        ", 
               "       decode(acs.constraint_type, 'R', 'R', acs.constraint_type) constraint_type, ", 
               "       acc.column_names,                                                           ", 
               "       accr.table_name r_table_name,                                               ", 
               "       accr.column_names                                                           ", 
               "  from dba_constraints acs                                                         ", 
               "  left outer join (                                                                ", 
               "                   select accs.constraint_name,accs.owner,accs.table_name,         ",
               "                          listagg(accs.column_name, ',')                           ", 
               "                            within group(order by accs.position) column_names      ", 
               "                     from dba_cons_columns accs                                    ", 
               "                    group by accs.constraint_name,accs.owner,accs.table_name       ", 
               "                  ) acc on acc.constraint_name = acs.constraint_name               ", 
               "                       and acc.table_name      = acs.table_name                    ",                                                                             
               "                       and acc.owner           = acs.owner                         ", 
               "  left outer join (                                                                ", 
               "                   select accs.constraint_name,accs.owner,accs.table_name,         ",
               "                          listagg(accs.column_name, ',')                           ",
               "                            within group(order by accs.position) column_names      ", 
               "                     from dba_cons_columns accs                                    ", 
               "                    group by accs.constraint_name,accs.owner,accs.table_name       ", 
               "                  ) accr on accr.constraint_name = acs.r_constraint_name           ", 
               "                        and accr.owner           = acs.r_owner                     ", 
               " where 1 = 1                                                                       ", 
               "   and acs.table_name = '",ls_table_name.toUpperCase(),"'                          ", 
               "   and acs.owner      = '",ls_owner.toUpperCase(),"'                               ", 
               "   and acs.constraint_name in (",ls_constraints.toUpperCase(),")                   "
  &endif
  &endif
  
  PREPARE lpre_get_drop_constraint_sql FROM ls_sql
  DECLARE lcur_get_drop_constraint_sql CURSOR FOR lpre_get_drop_constraint_sql

  LET li_rec_cnt = 1
  
  OPEN lcur_get_drop_constraint_sql
  FOREACH lcur_get_drop_constraint_sql INTO lo_drop_constraint[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET ls_Line = ""

    IF NVL(lo_drop_constraint[li_rec_cnt].DZED003,cs_null_value) = "P" THEN
      &ifdef is_msv
      LET ls_Line = ls_Line,
                    "-- SQL Server drop primary key.","\n",
                    "if object_id('dbo.[",lo_drop_constraint[li_rec_cnt].DZED002,"]') is not null",
                    " alter table ",lo_drop_constraint[li_rec_cnt].dzed001," drop constraint ",lo_drop_constraint[li_rec_cnt].DZED002,"\n","go","\n"
      &else
      &ifdef is_pgs
      LET ls_drop_null_sql = sadzi140_db_get_drop_not_null_sql(ls_owner,ls_table_name,lo_drop_constraint[li_rec_cnt].DZED002) 
      LET ls_Line = ls_Line,
                    "-- PostgreSQL drop primary key 2.","\n",
                    " alter table ",lo_drop_constraint[li_rec_cnt].dzed001," drop constraint ",lo_drop_constraint[li_rec_cnt].DZED002,";","\n",
                    ls_drop_null_sql,"\n"
      &else
      LET ls_db_type = sadzi140_util_get_db_type() #20240521
      CASE ls_db_type
        WHEN cs_db_driver_type_kdb --20240521
          LET ls_drop_null_sql = sadzi140_db_get_drop_not_null_sql(ls_owner,ls_table_name,lo_drop_constraint[li_rec_cnt].DZED002) 
          LET ls_Line = ls_Line,
                        "-- KingBASE drop primary key 2.","\n",
                        " alter table ",lo_drop_constraint[li_rec_cnt].dzed001," drop constraint ",lo_drop_constraint[li_rec_cnt].DZED002,";","\n",
                        ls_drop_null_sql,"\n"
        WHEN cs_db_driver_type_dmg --20230804
          LET ls_Line = ls_Line,
                        "-- DM drop primary key [SEC3].","\n",
                        "alter table ",lo_drop_constraint[li_rec_cnt].dzed001," drop primary key cascade;","\n"
      OTHERWISE  
        LET ls_Line = ls_Line,
                      "-- Oracle drop primary key [SEC3].","\n",
                      "alter table ",lo_drop_constraint[li_rec_cnt].dzed001," drop primary key cascade drop index;","\n"
      END CASE
      &endif       
      &endif       
    ELSE                     
      &ifdef is_msv
      LET ls_Line = ls_Line,
                    "-- SQL Server drop constraint.","\n",
                    "if object_id('dbo.[",lo_drop_constraint[li_rec_cnt].DZED002,"]') is not null",
                    " alter table ",lo_drop_constraint[li_rec_cnt].dzed001," drop constraint ",lo_drop_constraint[li_rec_cnt].DZED002,"\n","go","\n"
      &else
      &ifdef is_pgs
      LET ls_drop_null_sql = sadzi140_db_get_drop_not_null_sql(ls_owner,ls_table_name,lo_drop_constraint[li_rec_cnt].DZED002) 
      LET ls_Line = ls_Line,
                    "-- PostgreSQL drop constraint.","\n",
                    "alter table ",lo_drop_constraint[li_rec_cnt].dzed001," drop constraint ",lo_drop_constraint[li_rec_cnt].DZED002,";","\n",
                    ls_drop_null_sql,"\n"
      &else
      LET ls_db_type = sadzi140_util_get_db_type() #20240521
      CASE ls_db_type
        WHEN cs_db_driver_type_kdb --20240521
          LET ls_drop_null_sql = sadzi140_db_get_drop_not_null_sql(ls_owner,ls_table_name,lo_drop_constraint[li_rec_cnt].DZED002) 
          LET ls_Line = ls_Line,
                        "-- PostgreSQL drop constraint.","\n",
                        "alter table ",lo_drop_constraint[li_rec_cnt].dzed001," drop constraint ",lo_drop_constraint[li_rec_cnt].DZED002,";","\n",
                        ls_drop_null_sql,"\n"
      --WHEN cs_db_driver_type_dmg --20230804
        --與 Oracle 共用
      OTHERWISE  
        LET ls_Line = ls_Line,
                      "-- Oracle drop constraint.","\n",
                      "alter table ",lo_drop_constraint[li_rec_cnt].dzed001," drop constraint ",lo_drop_constraint[li_rec_cnt].DZED002," cascade;","\n"
      END CASE  
      &endif       
      &endif       
    END IF                     
      
    CALL lo_string_buffer.append(ls_Line)
    
    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_drop_constraint_sql
  CALL lo_drop_constraint.deleteElement(li_rec_cnt)
  
  FREE lpre_get_drop_constraint_sql
  FREE lcur_get_drop_constraint_sql  

  LET ls_return = lo_string_buffer.toString()
 
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_get_drop_not_null_sql(p_owner,p_table_name,p_pk_name)
DEFINE 
  p_owner        STRING,
  p_table_name   STRING,
  p_pk_name      STRING
DEFINE
  ls_owner         STRING,
  ls_table_name    STRING,
  lo_pk_name       STRING,
  ls_db_type       STRING, #20240521
  lv_alter_sql     VARCHAR(1024),
  ls_alter_line    STRING,
  ls_sql           STRING
DEFINE  
  ls_return STRING

  #20240521 begin
  LET ls_db_type = sadzi140_util_get_db_type()
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb
      LET ls_owner      = p_owner.toUpperCase()
      LET ls_table_name = p_table_name.toUpperCase()
      LET lo_pk_name    = p_pk_name.toUpperCase()
  --WHEN cs_db_driver_type_dmg
    --與 Oracle 共用 
  OTHERWISE  
    LET ls_owner      = p_owner.toLowerCase()
    LET ls_table_name = p_table_name.toLowerCase()
    LET lo_pk_name    = p_pk_name.toLowerCase()
  END CASE
  #20240521 end
  
  LET lv_alter_sql = NULL
  LET ls_alter_line = ""

  LET ls_sql = "select 'alter table '||dccs.table_name||' alter column '||dccs.column_name||' drop not null' ",
               "  from dba_cons_columns dccs                                                                 ",
               " where dccs.owner = '",ls_owner,"'                                                           ",
               "   and dccs.table_name = '",ls_table_name,"'                                                 ",
               "   and dccs.constraint_name = '",lo_pk_name,"'                                               "

  PREPARE lpre_get_drop_not_null_sql FROM ls_sql
  DECLARE lcur_get_drop_not_null_sql CURSOR FOR lpre_get_drop_not_null_sql

  OPEN lcur_get_drop_not_null_sql
  FOREACH lcur_get_drop_not_null_sql INTO lv_alter_sql
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET ls_alter_line = ls_alter_line,
                        lv_alter_sql,";","\n"
                        
  END FOREACH
  CLOSE lcur_get_drop_not_null_sql
  
  FREE lpre_get_drop_not_null_sql
  FREE lcur_get_drop_not_null_sql  

  LET ls_return = ls_alter_line
 
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_get_alter_index_sql(p_owner,p_table_name,p_diff_indexes)
DEFINE 
  p_owner        STRING,
  p_table_name   STRING,
  p_diff_indexes DYNAMIC ARRAY OF T_INDEX_DIFF 
DEFINE
  ls_owner         STRING,
  ls_table_name    STRING,
  lo_diff_indexes  DYNAMIC ARRAY OF T_INDEX_DIFF,
  lo_alter_indexes DYNAMIC ARRAY OF T_INDEX_DIFF,
  li_loop          INTEGER, 
  li_rec_cnt       INTEGER,
  ls_indexes       STRING,
  ls_sql           STRING,
  ls_where         STRING,
  ls_line          STRING,
  ls_db_type       STRING, #20240521
  ls_drop_index    STRING, #20171120
  lb_index_exist   BOOLEAN,
  ls_unique        STRING,
  lo_string_buffer base.StringBuffer
DEFINE  
  ls_return STRING

  LET ls_owner          = p_owner
  LET ls_table_name     = p_table_name
  LET lo_diff_indexes   = p_diff_indexes

  LET lo_string_buffer = base.StringBuffer.create()

  FOR li_loop = 1 TO lo_diff_indexes.getLength()
    LET ls_indexes = ls_indexes,",'",lo_diff_indexes[li_loop].dzec002,"'"
  END FOR 
  LET ls_indexes = ls_indexes.subString(2,ls_indexes.getLength())

  LET ls_sql = "select ec.dzec001,                                    ",
               "       ec.dzec002,                                    ",
               "       ec.dzec003,                                    ",
               "       ec.dzec004                                     ",
               "  from dzec_t ec                                      ",
               " where ec.dzec001 = '",ls_table_name.toLowerCase(),"' ",
               "   and ec.dzec002 in (",ls_indexes.toLowerCase(),")   ",
               "   and coalesce(ec.dzecstus,'Y') = 'Y'                "  #20170912

  PREPARE lpre_get_alter_index_sql FROM ls_sql
  DECLARE lcur_get_alter_index_sql CURSOR FOR lpre_get_alter_index_sql

  LET li_rec_cnt = 1
  
  OPEN lcur_get_alter_index_sql
  FOREACH lcur_get_alter_index_sql INTO lo_alter_indexes[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET ls_Line = ""

    IF NVL(lo_alter_indexes[li_rec_cnt].DZEC003,cs_null_value) = "U" THEN
      LET ls_unique = "unique";
    ELSE
      LET ls_unique = "";
    END IF

    CALL sadzi140_db_check_table_index_exist(ls_owner,lo_alter_indexes[li_rec_cnt].DZEC002) RETURNING lb_index_exist
    IF lb_index_exist THEN
      CALL sadzi140_db_get_drop_index_procedure(lo_alter_indexes[li_rec_cnt].dzec001,lo_alter_indexes[li_rec_cnt].DZEC002) RETURNING ls_drop_index  #20191209 #20171120 
      LET ls_Line = ls_Line,
                    "-- sadzi140_db_get_alter_index_sql()","\n",
                    ls_drop_index,"\n" #20171120
                    --"drop index ",lo_alter_indexes[li_rec_cnt].DZEC002,";","\n"
    END IF                
      
    &ifdef is_msv
    LET ls_Line = ls_Line,
                  "create ",ls_unique," index ",lo_alter_indexes[li_rec_cnt].DZEC002,
                  " on ",lo_alter_indexes[li_rec_cnt].DZEC001,
                  " (",lo_alter_indexes[li_rec_cnt].DZEC004,")","\n","go","\n"

    LET ls_Line = ls_Line,
                  "update statistics ",lo_alter_indexes[li_rec_cnt].DZEC001,"\n","go","\n"
    &else
    &ifdef is_pgs
    LET ls_Line = ls_Line,
                  "create ",ls_unique," index ",lo_alter_indexes[li_rec_cnt].DZEC002,
                  " on ",lo_alter_indexes[li_rec_cnt].DZEC001,
                  " (",lo_alter_indexes[li_rec_cnt].DZEC004,");","\n"
                  
    LET ls_Line = ls_Line,
                  "analyze ",lo_alter_indexes[li_rec_cnt].DZEC001,";","\n"
    &else
    LET ls_Line = ls_Line,
                  "create ",ls_unique," index ",lo_alter_indexes[li_rec_cnt].DZEC002,
                  " on ",lo_alter_indexes[li_rec_cnt].DZEC001,
                  " (",lo_alter_indexes[li_rec_cnt].DZEC004,");","\n"

    LET ls_db_type = sadzi140_util_get_db_type() #20240521
    CASE ls_db_type
      WHEN cs_db_driver_type_kdb #20240521
        LET ls_Line = ls_Line,
                      "analyze ",lo_alter_indexes[li_rec_cnt].DZEC001,";","\n"
      WHEN cs_db_driver_type_dmg #20230804
        LET ls_Line = ls_Line,
                      "BEGIN                                                                                           ","\n",
                      "  --CALL SP_TAB_STAT_DEINIT(CURRENT_USER,'",UPSHIFT(lo_alter_indexes[li_rec_cnt].DZEC001),"');  ","\n",
                      "  CALL SP_TAB_STAT_INIT(CURRENT_USER,'",UPSHIFT(lo_alter_indexes[li_rec_cnt].DZEC001),"');      ","\n",
                      "EXCEPTION                                                                                       ","\n",
                      "  WHEN OTHERS THEN                                                                              ","\n",
                      "     dbms_output.put_line('[",cs_warning_tag,"]'||SQLERRM);                                     ","\n",
                      "END;                                                                                            ","\n",
                      "/                                                                                               ","\n"
    OTHERWISE   
      #20171120 begin              
      LET ls_Line = ls_Line,
                    "begin                                                                                      ","\n",
                    "  dbms_stats.unlock_table_stats(user,'",UPSHIFT(lo_alter_indexes[li_rec_cnt].DZEC001),"'); ","\n",
                    "exception                                                                                  ","\n",
                    "  when others then                                                                         ","\n",
                    "    dbms_output.put_line('",cs_warning_tag,"'||SQLERRM);                                   ","\n",
                    "end;                                                                                       ","\n",
                    "/                                                                                          ","\n"
      #20171120 end              
      LET ls_Line = ls_Line,
                    "analyze table ",lo_alter_indexes[li_rec_cnt].DZEC001," estimate statistics;","\n"
    END CASE  
    &endif
    &endif
    
    CALL lo_string_buffer.append(ls_Line)
    
    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_alter_index_sql
  CALL lo_alter_indexes.deleteElement(li_rec_cnt)
  
  FREE lpre_get_alter_index_sql
  FREE lcur_get_alter_index_sql  

  LET ls_return = lo_string_buffer.toString()
 
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_get_drop_index_sql(p_owner,p_table_name,p_prev_version,p_diff_indexes)
DEFINE 
  p_owner        STRING,
  p_table_name   STRING,
  p_prev_version STRING,
  p_diff_indexes DYNAMIC ARRAY OF T_INDEX_DIFF 
DEFINE
  ls_owner         STRING,
  ls_table_name    STRING,
  ls_prev_version  STRING,
  lo_diff_indexes  DYNAMIC ARRAY OF T_INDEX_DIFF,
  lo_drop_indexes  DYNAMIC ARRAY OF T_INDEX_DIFF,
  li_loop          INTEGER, 
  li_rec_cnt       INTEGER,
  ls_indexes       STRING,
  ls_sql           STRING,
  ls_where         STRING,
  ls_line          STRING,
  ls_drop_index    STRING, #20171120
  lb_table_index_exist    BOOLEAN,
  lb_designer_index_exist BOOLEAN,
  ls_unique               STRING,
  lo_string_buffer        base.StringBuffer
DEFINE  
  ls_return STRING

  LET ls_owner          = p_owner
  LET ls_table_name     = p_table_name
  LET ls_prev_version   = p_prev_version
  LET lo_diff_indexes.* = p_diff_indexes.*

  LET lo_string_buffer = base.StringBuffer.create()

  FOR li_loop = 1 TO lo_diff_indexes.getLength()
    LET ls_indexes = ls_indexes,",'",lo_diff_indexes[li_loop].dzec002,"'"
  END FOR 
  LET ls_indexes = ls_indexes.subString(2,ls_indexes.getLength())

  &ifdef is_msv
  LET ls_sql = "select dis.table_name,dis.index_name,substring(dis.uniqueness,1,1) index_type, ",
               "       dis.column_names                                                        ",
               "  from dba_indexes dis                                                         ",
               " where dis.owner = '",ls_owner.toLowerCase(),"'                                ",
               "   and dis.table_name = '",ls_table_name.toLowerCase(),"'                      ",
               "   and dis.index_name in (",ls_indexes.toLowerCase(),")                        "  
  &else
  &ifdef is_pgs
  LET ls_sql = "select dis.table_name,dis.index_name,substring(dis.uniqueness,1,1) index_type,   ",
               "       string_agg(aic.column_name,',' order by aic.column_position) column_names ",
               "  from dba_indexes dis                                                           ",
               " inner join dba_ind_columns aic                                                  ",
               "         on aic.table_owner = dis.table_owner                                    ",
               "        and aic.table_name  = dis.table_name                                     ",
               "        and aic.index_name  = dis.index_name                                     ",
               " where dis.owner = '",ls_owner.toLowerCase(),"'                                  ",
               "   and dis.table_name = '",ls_table_name.toLowerCase(),"'                        ",
               "   and dis.index_name in (",ls_indexes.toLowerCase(),")                          ",  
               " group by dis.table_name,dis.index_name,substring(dis.uniqueness,1,1)            "
  &else
  LET ls_sql = "SELECT AIS.TABLE_NAME,AIS.INDEX_NAME,SUBSTRB(AIS.UNIQUENESS,1,1) INDEX_TYPE,                                 ",
               "       LISTAGG(AICS.COLUMN_NAME,',') WITHIN GROUP (ORDER BY AIS.INDEX_NAME,AICS.COLUMN_POSITION) COLUMN_NAME ",
               "  FROM DBA_INDEXES AIS,DBA_IND_COLUMNS AICS                                                                  ",
               " WHERE AIS.OWNER        = '",ls_owner.toUpperCase(),"'                                                       ",
               "   AND AIS.TABLE_NAME   = '",ls_table_name.toUpperCase(),"'                                                  ",
               "   AND AIS.INDEX_TYPE <> 'CLUSTER'                                                                           ", --20230807    
               "   AND AICS.TABLE_OWNER = AIS.TABLE_OWNER                                                                    ",
               "   AND AICS.TABLE_NAME  = AIS.TABLE_NAME                                                                     ",
               "   AND AICS.INDEX_NAME  = AIS.INDEX_NAME                                                                     ",
               "   AND AIS.INDEX_NAME NOT IN (                                                                               ",
               "                               SELECT ACS.INDEX_NAME                                                         ", --20230807
               "                                 FROM DBA_CONSTRAINTS ACS                                                    ",
               "                                WHERE ACS.TABLE_NAME = AIS.TABLE_NAME                                        ",
               "                                  AND ACS.OWNER      = AIS.OWNER                                             ",
               "                             )                                                                               ",
               "   AND AIS.INDEX_NAME IN (",ls_indexes.toUpperCase(),")                                                      ",
               "  GROUP BY AIS.TABLE_NAME,AIS.INDEX_NAME,SUBSTRB(AIS.UNIQUENESS,1,1)                                         "
  &endif               
  &endif               
               
  PREPARE lpre_get_drop_index_sql FROM ls_sql
  DECLARE lcur_get_drop_index_sql CURSOR FOR lpre_get_drop_index_sql

  LET li_rec_cnt = 1
  
  OPEN lcur_get_drop_index_sql
  FOREACH lcur_get_drop_index_sql INTO lo_drop_indexes[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET ls_Line = ""

    --CALL sadzi140_db_check_table_index_exist(ls_owner,lo_drop_indexes[li_rec_cnt].DZEC002) RETURNING lb_table_index_exist
    --CALL sadzi140_db_check_designer_index_exist(lo_drop_indexes[li_rec_cnt].DZEC002) RETURNING lb_designer_index_exist
    #實際表格 Index 存在及 r.t 的資料不存在時, 才產出 drop index Scripts  
    --IF lb_table_index_exist AND NOT lb_designer_index_exist THEN
      CALL sadzi140_db_get_drop_index_procedure(lo_drop_indexes[li_rec_cnt].dzec001,lo_drop_indexes[li_rec_cnt].DZEC002) RETURNING ls_drop_index  #20191209 #20171120
      LET ls_Line = ls_Line,
                    "-- sadzi140_db_get_drop_index_sql()","\n",
                    ls_drop_index,"\n" #20171120
                    --"drop index ",lo_drop_indexes[li_rec_cnt].DZEC002,";","\n"
    --END IF                
      
    CALL lo_string_buffer.append(ls_Line)
    
    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_drop_index_sql
  CALL lo_drop_indexes.deleteElement(li_rec_cnt)
  
  FREE lpre_get_drop_index_sql
  FREE lcur_get_drop_index_sql  

  LET ls_return = lo_string_buffer.toString()
 
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_get_alter_column_sql(p_master_user,p_table_owner,p_table_name,p_object_type,p_diff_columns)
DEFINE 
  p_master_user  STRING, 
  p_table_owner  STRING,
  p_table_name   STRING,
  p_object_type  STRING,
  p_diff_columns DYNAMIC ARRAY OF T_COLUMN_DIFF 
DEFINE
  ls_master_user    STRING, 
  ls_table_owner    STRING,
  ls_table_name     STRING,
  ls_object_type    STRING,
  lo_diff_columns   DYNAMIC ARRAY OF T_COLUMN_DIFF,
  lo_alter_columns  DYNAMIC ARRAY OF T_COLUMN_ALTER,
  lo_db_column_data     T_COLUMN_DATA,
  lo_design_column_data T_COLUMN_DATA,
  li_loop           INTEGER, 
  li_rec_cnt        INTEGER,
  ls_columns        STRING,
  ls_sql            STRING,
  ls_where          STRING,
  ls_alter_type     STRING,
  ls_alter_script   STRING,
  ls_alter_sql      STRING,
  ls_data_type      STRING,
  ls_type_leading   STRING, #20200330
  ls_type_using     STRING, #20200330
  ls_default        STRING,
  ls_lob_tablespace STRING,
  ls_sql_upd_dzeb_t STRING,
  ls_db_type        STRING, #20230521 
  ls_not_null       STRING, #20191209
  ls_drop_cons      STRING, #20191209
  ls_drop_index     STRING, #20191209
  ls_comment_script STRING, #20161220
  ls_rebuild_index_list STRING,
  lb_column_exist   BOOLEAN
DEFINE  
  ls_return      STRING

  LET ls_master_user    = p_master_user
  LET ls_table_owner    = p_table_owner
  LET ls_table_name     = p_table_name
  LET ls_object_type    = p_object_type
  LET lo_diff_columns.* = p_diff_columns.*

  LET ls_alter_type   = ""
  LET ls_alter_script = ""
  LET ls_alter_sql    = ""

  INITIALIZE mo_columns_data_in_mem TO NULL

  #取得 LOB 型態欄位的Tablespace
  CALL sadzi140_db_get_parameter(cs_param_level_lob,cs_param_tablespace) RETURNING ls_lob_tablespace
  CALL sadzi140_db_get_column_data_to_mem(ls_table_owner,ls_table_name) RETURNING mo_columns_data_in_mem
  
  FOR li_loop = 1 TO lo_diff_columns.getLength()
    LET ls_columns = ls_columns,",'",lo_diff_columns[li_loop].DZEV005,"'"
  END FOR

  LET ls_columns = ls_columns.subString(2,ls_columns.getLength())
  #LET ls_where = " AND EV.DZEV005 IN (",ls_columns,")"
  LET ls_where = " and eb.dzeb002 in (",ls_columns.toLowerCase(),")"

  &ifdef is_msv
  LET ls_sql = "select eb.dzeb002                             dzeb002,      ", 
               "       eb.dzeb004                             dzeb004,      ", 
               "       case eb.dzeb004                                      ", 
               "         when 'Y' then 'N'                                  ", 
               "         when 'N' then 'Y'                                  ", 
               "       else                                                 ", 
               "         'Y'                                                ", 
               "       end                                    dzeb005,      ", 
               "       td.gztd003                             dzeb007,      ", 
               "       case td.gztd003                                      ", 
               "         when 'varbinary'                                   ", 
               "           then 'max'                                       ", 
               "         when 'nvarchar'                                    ", 
               "           then iif(td.gztd009='TEXT','max',td.gztd008)     ", 
               "       else                                                 ", 
               "         td.gztd008                                         ",
               "       end                                    dzeb008,      ",
               "       trim(eb.dzeb012)                       dzeb012       ",
               "  from dzeb_t eb                                            ",
               "       left outer join gztd_t td on td.gztd001 = eb.dzeb006 ", 
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'       ",
               ls_where,
               " order by eb.dzeb021                                        "
  &else
  &ifdef is_pgs
  LET ls_sql = "select eb.dzeb002           dzeb002,                  ", 
               "       eb.dzeb004           dzeb004,                  ", 
               "       case eb.dzeb004                                ", 
               "         when 'Y' then 'N'                            ", 
               "         when 'N' then 'Y'                            ", 
               "       else                                           ", 
               "         'Y'                                          ", 
               "       end                  dzeb005,                  ", 
               "       td.gztd003           dzeb007,                  ", 
               "       td.gztd008           dzeb008,                  ",
               "       trim(eb.dzeb012)     dzeb012                   ",
               "  from dzeb_t eb                                      ",
               "       left outer join gztd_t td                      ",
               "                    on td.gztd001 = eb.dzeb006        ", 
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"' ",
               ls_where,
               " order by eb.dzeb021                                  "
  &else
  LET ls_sql = "select upper(eb.dzeb002)                      dzeb002,       ",
               "       eb.dzeb004                             dzeb004,       ",
               "       decode(eb.dzeb004,'Y','N','N','Y','Y') dzeb005,       ",
               "       upper(td.gztd003)                      dzeb007,       ",
               "       decode(                                               ",
               "              upper(td.gztd003),                             ",
               "              'BLOB','4000',                                 ",
               "              'CLOB','4000',                                 ",
               "              'DATE','7',                                    ",
               "              td.gztd008                                     ",
               "             )                                dzeb008,       ",
               "       trim(decode(eb.dzeb012,'',null,eb.dzeb012))  dzeb012  ",
               "  from dzeb_t eb                                             ",
               "       left outer join gztd_t td on td.gztd001 = eb.dzeb006  ",
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'        ",
               ls_where,
               " order by eb.dzeb021                                         "
  &endif
  &endif
  
  PREPARE lpre_get_alter_column_sql FROM ls_sql
  DECLARE lcur_get_alter_column_sql CURSOR FOR lpre_get_alter_column_sql

  LET li_rec_cnt = 1
  
  OPEN lcur_get_alter_column_sql
  FOREACH lcur_get_alter_column_sql INTO lo_alter_columns[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    #取得欄位資料
    CALL sadzi140_db_get_db_column_data(ls_table_owner,ls_table_name,lo_alter_columns[li_rec_cnt].DZEB002) RETURNING lo_db_column_data.*
    CALL sadzi140_db_get_design_column_data(ls_table_owner,ls_table_name,lo_alter_columns[li_rec_cnt].DZEB002) RETURNING lo_design_column_data.* 

    #查看看欄位是否存在, 若存在, 則做Modify, 否則做 Add
    CALL sadzi140_db_check_db_column_exist(ls_table_owner,ls_table_name,lo_alter_columns[li_rec_cnt].DZEB002) RETURNING lb_column_exist
    IF lb_column_exist THEN
      &ifdef is_msv
      LET ls_alter_type = cs_alter_type_alter_column
      &else
      &ifdef is_pgs
      LET ls_alter_type = cs_alter_type_alter_column
      &else
      LET ls_db_type = sadzi140_util_get_db_type() #20240521
      CASE ls_db_type
        WHEN cs_db_driver_type_kdb #20240521
          LET ls_alter_type = cs_alter_type_alter_column
      --WHEN cs_db_driver_type_dmg 
        --與Oracle共用
      OTHERWISE   
        LET ls_alter_type = cs_alter_type_modify
      END CASE  
      &endif
      &endif
    ELSE
      LET ls_alter_type = cs_alter_type_add
    END IF

    LET ls_default = ""
    &ifdef is_msv
    IF ls_alter_type = cs_alter_type_add THEN 
      LET ls_default = cs_alter_column_section,"1.0","\n",
                       "alter table ",ls_table_name," add constraint df_",ls_table_name,"_",lo_alter_columns[li_rec_cnt].DZEB002," default null for ",lo_alter_columns[li_rec_cnt].DZEB002,""
    END IF  
    &else
    #Oracle & PostgreSQL 共用
    LET ls_default = "default null" -- 20170414
    &endif

    #設定Data Type and clear default
    IF (NVL(lo_db_column_data.DATA_TYPE,cs_null_default) <> NVL(lo_design_column_data.DATA_TYPE,cs_null_default)) OR 
       (NVL(lo_db_column_data.DATA_LENGTH,cs_null_default) <> NVL(lo_design_column_data.DATA_LENGTH,cs_null_default)) THEN
    
      IF (lo_alter_columns[li_rec_cnt].DZEB007 = 'DATE') OR 
         (lo_alter_columns[li_rec_cnt].DZEB007 = 'date') OR --20200210 
         (lo_alter_columns[li_rec_cnt].DZEB007 = 'text') OR --20200330
         (UPSHIFT(lo_alter_columns[li_rec_cnt].DZEB007) = 'INT2') OR --20200330 
         (UPSHIFT(lo_alter_columns[li_rec_cnt].DZEB007) = 'INT4') OR --20200330
         (UPSHIFT(lo_alter_columns[li_rec_cnt].DZEB007) = 'INT8') OR --20200330
         (lo_alter_columns[li_rec_cnt].DZEB007 = 'bytea') OR --20200330
         (lo_alter_columns[li_rec_cnt].DZEB007 = 'SMALLINT') OR --20230804 
         (lo_alter_columns[li_rec_cnt].DZEB007 = 'INTEGER') OR --20230804  
         (lo_alter_columns[li_rec_cnt].DZEB007 = 'BIGINT') OR --20230804 
         (lo_alter_columns[li_rec_cnt].DZEB007 = 'CLOB') OR 
         (lo_alter_columns[li_rec_cnt].DZEB007 = 'BLOB') THEN 
        LET ls_data_type = lo_alter_columns[li_rec_cnt].DZEB007
      ELSE      
        LET ls_data_type = lo_alter_columns[li_rec_cnt].DZEB007,"(",lo_alter_columns[li_rec_cnt].DZEB008,")"
      END IF

      &ifdef is_msv
      LET ls_not_null = ""
      IF (lo_alter_columns[li_rec_cnt].dzeb004 = "Y") THEN --AND (ls_alter_type = cs_alter_type_add) THEN 
        LET ls_not_null = "not null"
      ELSE
        --SQL Server 
        #SQL Server 將 column 設為非 Null 之前要先檢核是否為 Constraint
        LET ls_drop_cons = "declare                                                                                     ","\n",
                           "  @ls_sql         nvarchar(max),                                                            ","\n",
                           "  @ls_table_name  varchar(128) = '",ls_table_name,"',                                       ","\n",
                           "  @ls_column_name varchar(128) = '",lo_alter_columns[li_rec_cnt].DZEB002,"'                 ","\n",
                           "begin                                                                                       ","\n",
                           "  select @ls_sql = 'alter table '+uccs.table_name+ ' drop constraint '+uccs.constraint_name ","\n",
                           "    from user_cons_columns uccs                                                             ","\n",
                           "   where uccs.table_name  = @ls_table_name                                                  ","\n",
                           "     and uccs.column_name = @ls_column_name                                                 ","\n",
                           "  if @@ROWCOUNT <> 0                                                                        ","\n",
                           "    begin try                                                                               ","\n",
                           "      exec(@ls_sql)                                                                         ","\n",
                           "    end try                                                                                 ","\n",
                           "    begin catch                                                                             ","\n",
                           "      print '",cs_message_tag,"'+'Exec '+@ls_sql+' not success.'                            ","\n",
                           "    end catch                                                                               ","\n",
                           "  else                                                                                      ","\n",
                           "    print '",cs_message_tag,"'+'No key field '+@ls_table_name+'.'+@ls_column_name+' found.' ","\n",
                           "end                                                                                         ","\n",
                           "go","\n"
      END IF
      LET ls_drop_index = sadzi140_db_get_drop_index_before_add_pk(ls_table_name,lo_alter_columns[li_rec_cnt].DZEB002)
      LET ls_alter_script = ls_drop_index,"\n",
                            ls_drop_cons,"\n",
                            cs_alter_column_section,"2.0","\n",
                            "alter table",cs_SPACE,
                            ls_table_name,cs_SPACE,
                            ls_alter_type,cs_SPACE,
                            lo_alter_columns[li_rec_cnt].DZEB002,cs_SPACE,
                            ls_data_type,cs_SPACE,
                            ls_not_null,cs_SPACE,
                            "\n","go","\n",
                            ls_default,cs_SPACE,
                            "\n","go","\n"
      &else
      &ifdef is_pgs
      --PostgreSQL
      ----Set datatype----

      LET ls_type_leading = ""
      LET ls_type_using = ""
      CASE
        WHEN ls_alter_type = cs_alter_type_add 
          LET ls_type_leading = "" 
        WHEN ls_alter_type = cs_alter_type_alter_column 
          LET ls_type_leading = "type" 
          CASE ls_data_type.toLowerCase()
            WHEN "int2"
              LET ls_type_using = cs_pgs_alter_using,cs_SPACE,
                                  lo_alter_columns[li_rec_cnt].DZEB002,cs_pgs_type_cast,"smallint"
            WHEN "int4"
              LET ls_type_using = cs_pgs_alter_using,cs_SPACE,
                                  lo_alter_columns[li_rec_cnt].DZEB002,cs_pgs_type_cast,"integer"
            WHEN "int8"
              LET ls_type_using = cs_pgs_alter_using,cs_SPACE,
                                  lo_alter_columns[li_rec_cnt].DZEB002,cs_pgs_type_cast,"bigint"
          OTHERWISE     
            LET ls_type_using = cs_pgs_alter_using,cs_SPACE,
                                lo_alter_columns[li_rec_cnt].DZEB002,cs_pgs_type_cast,ls_data_type
          END CASE
      END CASE 

      -- Drop default
      IF (ls_alter_type = cs_alter_type_alter_column) THEN 
        LET ls_alter_script = ls_alter_script,
                              cs_alter_column_section,"3.0","\n",
                              "alter table",cs_SPACE,
                              ls_table_name,cs_SPACE,
                              ls_alter_type,cs_SPACE,
                              lo_alter_columns[li_rec_cnt].DZEB002,cs_SPACE,
                              "drop default",cs_SPACE,
                              ";","\n"
      END IF                            

      -- Set type
      LET ls_alter_script = ls_alter_script,
                            cs_alter_column_section,"4.0","\n",
                            "alter table",cs_SPACE,
                            ls_table_name,cs_SPACE,
                            ls_alter_type,cs_SPACE,
                            lo_alter_columns[li_rec_cnt].DZEB002,cs_SPACE,
                            ls_type_leading,cs_SPACE,
                            ls_data_type,cs_SPACE,
                            ls_type_using,
                            ";","\n"

      ----Clear default----                             
      LET ls_alter_script = ls_alter_script,
                            cs_alter_column_section,"5.0","\n",
                            "alter table",cs_SPACE,
                            ls_table_name,cs_SPACE,
                            cs_alter_type_alter_column,cs_SPACE,
                            lo_alter_columns[li_rec_cnt].DZEB002,cs_SPACE,
                            "set",cs_SPACE,
                            ls_default,
                            ";","\n"
      &else
      LET ls_db_type = sadzi140_util_get_db_type() #20240521
      CASE ls_db_type
        WHEN cs_db_driver_type_kdb #20240521
          --KDB
          ----Set datatype----
          LET ls_type_leading = ""
          LET ls_type_using = ""
          CASE
            WHEN ls_alter_type = cs_alter_type_add 
              LET ls_type_leading = "" 
            WHEN ls_alter_type = cs_alter_type_alter_column 
              LET ls_type_leading = "type" 
              CASE ls_data_type.toUpperCase()
                WHEN "INT2"
                  LET ls_type_using = cs_pgs_alter_using,cs_SPACE,
                                      lo_alter_columns[li_rec_cnt].DZEB002,cs_pgs_type_cast,"smallint"
                WHEN "INT4"
                  LET ls_type_using = cs_pgs_alter_using,cs_SPACE,
                                      lo_alter_columns[li_rec_cnt].DZEB002,cs_pgs_type_cast,"integer"
                WHEN "INT8"
                  LET ls_type_using = cs_pgs_alter_using,cs_SPACE,
                                      lo_alter_columns[li_rec_cnt].DZEB002,cs_pgs_type_cast,"bigint"
              OTHERWISE     
                LET ls_type_using = cs_pgs_alter_using,cs_SPACE,
                                    lo_alter_columns[li_rec_cnt].DZEB002,cs_pgs_type_cast,ls_data_type
              END CASE
          END CASE 

          -- Drop default
          IF (ls_alter_type = cs_alter_type_alter_column) THEN 
            LET ls_alter_script = ls_alter_script,
                                  cs_alter_column_section,"k3.0","\n",
                                  "alter table",cs_SPACE,
                                  ls_table_name,cs_SPACE,
                                  ls_alter_type,cs_SPACE,
                                  lo_alter_columns[li_rec_cnt].DZEB002,cs_SPACE,
                                  "drop default",cs_SPACE,
                                  ";","\n"
          END IF                            

          -- Set type
          LET ls_alter_script = ls_alter_script,
                                cs_alter_column_section,"k4.0","\n",
                                "alter table",cs_SPACE,
                                ls_table_name,cs_SPACE,
                                ls_alter_type,cs_SPACE,
                                lo_alter_columns[li_rec_cnt].DZEB002,cs_SPACE,
                                ls_type_leading,cs_SPACE,
                                ls_data_type,cs_SPACE,
                                ls_type_using,
                                ";","\n"

          ----Clear default----                             
          LET ls_alter_script = ls_alter_script,
                                cs_alter_column_section,"k5.0","\n",
                                "alter table",cs_SPACE,
                                ls_table_name,cs_SPACE,
                                cs_alter_type_alter_column,cs_SPACE,
                                lo_alter_columns[li_rec_cnt].DZEB002,cs_SPACE,
                                "set",cs_SPACE,
                                ls_default,
                                ";","\n"
                                
      --WHEN cs_db_driver_type_dmg #20240521
        --與 Oracle 共用
      OTHERWISE   
        --Oracle
        LET ls_alter_script = cs_alter_column_section,"6.0","\n",
                              "alter table",cs_SPACE,
                              ls_table_name,cs_SPACE,
                              ls_alter_type,cs_SPACE,
                              lo_alter_columns[li_rec_cnt].DZEB002,cs_SPACE,
                              ls_data_type,cs_SPACE,
                              ls_default, -- 20170414
                              ";","\n"
      END CASE  
      
      &endif
      &endif

      LET ls_alter_sql = ls_alter_sql,ls_alter_script
      LET ls_alter_script = "" #20200330
    END IF 

    #20161220 begin
    #設定 Comment
    IF ls_alter_type = cs_alter_type_add THEN
      &ifdef is_msv
      LET ls_comment_script = "declare                                                                                                      ","\n",
                              "  @user sysname                                                                                              ","\n",
                              "begin                                                                                                        ","\n",
                              "  select @user = user_name()                                                                                 ","\n",
                              "  begin try                                                                                                  ","\n",
                              "     execute sp_addextendedproperty 'comment','",lo_alter_columns[li_rec_cnt].DZEB002,"',                    ","\n",
                              "             'user', @user, 'table', '",ls_table_name,"','column','",lo_alter_columns[li_rec_cnt].DZEB002,"' ","\n",
                              "  end try                                                                                                    ","\n",
                              "  begin catch                                                                                                ","\n",
                              "    execute sp_updateextendedproperty 'comment','",lo_alter_columns[li_rec_cnt].DZEB002,"',                  ","\n",
                              "            'user', @user, 'table', '",ls_table_name,"','column','",lo_alter_columns[li_rec_cnt].DZEB002,"'  ","\n",
                              "  end catch                                                                                                  ","\n",
                              "end                                                                                                          ","\n",
                              "go","\n"
      &else
      --Oracle & PostgreSQL 共用
      LET ls_comment_script = "comment on column ",ls_table_name,".",lo_alter_columns[li_rec_cnt].DZEB002," is '",lo_alter_columns[li_rec_cnt].DZEB002,"'",";","\n"
      &endif
      LET ls_alter_sql = ls_alter_sql,ls_comment_script
      LET ls_comment_script = "" #20200330
    END IF
    #20161220 end
    
    #設定預設值(Set default value)
    IF (NVL(lo_db_column_data.DEFAULT_VALUE,cs_null_default) <> NVL(lo_design_column_data.DEFAULT_VALUE,cs_null_default)) OR
       (NVL(lo_db_column_data.DATA_LENGTH,cs_null_default) <> NVL(lo_design_column_data.DATA_LENGTH,cs_null_default)) THEN -- 20180710 
      
      &ifdef is_msv
      LET ls_alter_type = cs_alter_type_alter_column
      &else
      &ifdef is_pgs
      LET ls_alter_type = cs_alter_type_alter_column
      &else
      LET ls_alter_type = cs_alter_type_modify
      &endif
      &endif
      
      LET ls_sql_upd_dzeb_t = ""
      IF NVL(lo_alter_columns[li_rec_cnt].dzeb012,cs_null_default) <> cs_null_default THEN 
        CASE UPSHIFT(lo_alter_columns[li_rec_cnt].dzeb007)
          WHEN "BLOB"
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          WHEN "CLOB"
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          WHEN "DATE"
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          WHEN "NUMBER"
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          WHEN "NUMERIC" #20240529
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          WHEN "TIMESTAMP"
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          WHEN "VARCHAR2"
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          #20230804 begin  
          WHEN "SMALLINT"
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          WHEN "INTEGER"
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          WHEN "BIGINT"
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          #20230804 end  
          #20191209 begin  
          --SQL Server
          WHEN "DATETIME2"
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          WHEN "DECIMAL"
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          WHEN "NVARCHAR"
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          WHEN "VARBINARY"            
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          #20191209 end
          #20200330 begin
          --PostgreSQL
          WHEN "VARCHAR"
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          WHEN "TEXT"            
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          WHEN "INT2"            
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          WHEN "INT4"            
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          WHEN "INT8"            
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          WHEN "BYTEA"            
            LET ls_default = "default",cs_SPACE,lo_alter_columns[li_rec_cnt].dzeb012
          #20200330 end
        END CASE 
      ELSE
        LET ls_default = "default null"
        #LET ls_sql_upd_dzeb_t = "UPDATE DZEB_T SET DZEB012 = "" WHERE DZEB001 = '",ls_table_name,"' AND DZEB002 = '",lo_alter_columns[li_rec_cnt].DZEB002,"'",";","\n"
        #LET ls_sql_upd_dzeb_t = ls_sql_upd_dzeb_t,
        #                        "commit;",
        #                        "\n"
      END IF 
      
      &ifdef is_msv
      LET ls_alter_script = cs_alter_column_section,"7.0","\n",
                            "alter table ",ls_table_name," drop constraint df_",ls_table_name,"_",lo_alter_columns[li_rec_cnt].DZEB002,
                            "\n","go","\n",
                            cs_alter_column_section,"8.0","\n",
                            "alter table ",ls_table_name," add constraint df_",ls_table_name,"_",lo_alter_columns[li_rec_cnt].DZEB002," ",ls_default," for ",lo_alter_columns[li_rec_cnt].DZEB002,
                            "\n","go","\n"
      &else
      &ifdef is_pgs
      LET ls_alter_script = cs_alter_column_section,"9.0","\n",
                            "alter table",cs_SPACE,
                            ls_table_name,cs_SPACE,
                            ls_alter_type,cs_SPACE,
                            lo_alter_columns[li_rec_cnt].DZEB002,cs_SPACE,
                            "set",cs_SPACE,
                            ls_default,
                            ";","\n"
      &else
      LET ls_db_type = sadzi140_util_get_db_type() #20240521
      CASE ls_db_type
        WHEN cs_db_driver_type_kdb #20240521
          LET ls_alter_script = cs_alter_column_section,"k9.0","\n",
                                "alter table",cs_SPACE,
                                ls_table_name,cs_SPACE,
                                cs_alter_type_alter_column,cs_SPACE,
                                lo_alter_columns[li_rec_cnt].DZEB002,cs_SPACE,
                                "set",cs_SPACE,
                                ls_default,
                                ";","\n"
      --WHEN cs_db_driver_type_dmg #20240521
        --與Oracle共用
      OTHERWISE   
        LET ls_alter_script = cs_alter_column_section,"10.0","\n",
                              "alter table",cs_SPACE,
                              ls_table_name,cs_SPACE,
                              ls_alter_type,cs_SPACE,
                              lo_alter_columns[li_rec_cnt].DZEB002,cs_SPACE,
                              ls_default,
                              ";","\n"
      END CASE  
      &endif
      &endif

      LET ls_alter_sql = ls_alter_sql,ls_alter_script
      LET ls_alter_script = "" #20200330
      
      LET ls_alter_sql = ls_alter_sql,ls_sql_upd_dzeb_t
      LET ls_sql_upd_dzeb_t = "" #20200330
    END IF  

    &ifdef is_msv
    -----------------------------------
    ##SQL Server 無 TableSpace所以不做##
    -----------------------------------
    &else
    &ifdef is_pgs
    -----------------------------------
    ##PostgreSQL 暫不指定 Tablespace
    -----------------------------------
    &else
    LET ls_db_type = sadzi140_util_get_db_type() #20240521
    CASE ls_db_type
      WHEN cs_db_driver_type_kdb #20240521
        -----------------------------------
        ##KDB 暫不指定 Move LOB Tablespace
        -----------------------------------
      WHEN cs_db_driver_type_dmg #20240521
        -----------------------------------
        ##DMG 暫不指定 Move LOB Tablespace
        -----------------------------------
    OTHERWISE   
      #for Oracle 判斷為 LOB 型態的欄位, 將之移到 DSBLOB Tablespace
      IF ((lo_alter_columns[li_rec_cnt].DZEB007 = 'CLOB') OR (lo_alter_columns[li_rec_cnt].DZEB007 = 'BLOB')) THEN
        #Ex. ALTER TABLE DS.DZAB_T MOVE LOB(DZAB099) STORE AS (TABLESPACE BLOBDATA); 
        LET ls_alter_script = cs_alter_column_section,"11.0","\n",
                              "ALTER TABLE",cs_SPACE,
                              ls_table_name,cs_SPACE,
                              "MOVE LOB(",lo_alter_columns[li_rec_cnt].DZEB002,")",cs_SPACE,
                              "STORE AS (TABLESPACE ",ls_lob_tablespace,")",
                              ";","\n"
                             
        LET ls_alter_sql = ls_alter_sql,ls_alter_script
        LET ls_alter_script = "" #20200330

        #新增/異動 BLOB,CLOB 時要重建該表格相關 Index 
        CALL sadzi140_db_get_rebuild_index_list(ls_table_name) RETURNING ls_rebuild_index_list
        LET ls_alter_sql = ls_alter_sql,ls_rebuild_index_list
        LET ls_rebuild_index_list = "" #20200330
        
      END IF
    END CASE  
    &endif    
    &endif    

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_alter_column_sql
  CALL lo_alter_columns.deleteElement(li_rec_cnt)
  
  FREE lpre_get_alter_column_sql
  FREE lcur_get_alter_column_sql  

  LET ls_return = ls_alter_sql
 
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_get_alter_foreign_constraint_sql(p_table_owner,p_table_name)
DEFINE 
  p_table_owner  STRING,
  p_table_name   STRING
DEFINE
  ls_table_owner   STRING,
  ls_table_name    STRING,
  li_loop          INTEGER, 
  li_rec_cnt       INTEGER,
  lo_CONSTRAINT    T_CONSTRAINT_DIFF,
  ls_alter_sql     STRING,
  ls_alter_script  STRING,
  ls_sql           STRING,
  ls_db_type       STRING, --20240521
  ls_foreign_key_state  STRING,
  ls_enable_foreign_key STRING
DEFINE  
  ls_return      STRING

  LET ls_table_owner = p_table_owner.toUpperCase()
  LET ls_table_name  = p_table_name.toLowerCase()

  #取得 Foreign 是否啟動
  CALL sadzi140_db_get_parameter(cs_param_level_editor,cs_param_enable_foreign_key) RETURNING ls_enable_foreign_key
  LET ls_foreign_key_state = IIF(ls_enable_foreign_key=="Y",cs_state_enable,cs_state_disable)

  &ifdef is_msv
  LET ls_sql = "select ed.dzed001,ed.dzed002,ed.dzed003,ed.dzed004,ed.dzed005, ",
               "       ed.dzed006                                              ", 
               "  from dzed_t ed                                               ",
               " where ed.dzed005 = '",ls_table_name,"'                        ",
               "   and not exists (                                            ",
               "                    select 1                                   ",
               "                      from dba_constraints acs                 ",
               "                     where acs.owner = '",ls_table_owner,"'    ",
               "                       and acs.constraint_name = ed.dzed002    ",
               "                       and acs.table_name      = ed.dzed001    ",
               "                  )                                            ",
               "   and ed.dzed003 <> 'F'                                       "
  &else
  &ifdef is_pgs
  LET ls_sql = "select ed.dzed001,ed.dzed002,ed.dzed003,ed.dzed004,ed.dzed005, ",
               "       ed.dzed006                                              ", 
               "  from dzed_t ed                                               ",
               " where ed.dzed005 = '",ls_table_name,"'                        ",
               "   and not exists (                                            ",
               "                    select 1                                   ",
               "                      from dba_constraints acs                 ",
               "                     where acs.owner = '",ls_table_owner,"'    ",
               "                       and acs.constraint_name = ed.dzed002    ",
               "                       and acs.table_name      = ed.dzed001    ",
               "                  )                                            ",
               "   and ed.dzed003 <> 'F'                                       "
  &else
  LET ls_sql = "select ed.dzed001,ed.dzed002,ed.dzed003,ed.dzed004,ed.dzed005,     ",
               "       ed.dzed006                                                  ", 
               "  from dzed_t ed                                                   ",
               " where ed.dzed005 = '",ls_table_name,"'                            ",
               "   and not exists (                                                ",
               "                    select 1                                       ",
               "                      from dba_constraints acs                     ",
               "                     where acs.owner = '",ls_table_owner,"'        ",
               "                       and acs.constraint_name = upper(ed.dzed002) ",
               "                       and acs.table_name      = upper(ed.dzed001) ",
               "                  )                                                ",
               "   and ed.dzed003 <> 'F'                                           "
  &endif               
  &endif               
                              
  PREPARE lpre_get_alter_foreign_constraint_sql FROM ls_sql
  DECLARE lcur_get_alter_foreign_constraint_sql CURSOR FOR lpre_get_alter_foreign_constraint_sql

  LET li_rec_cnt = 1
  
  OPEN lcur_get_alter_foreign_constraint_sql
  FOREACH lcur_get_alter_foreign_constraint_sql INTO lo_CONSTRAINT.*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    &ifdef is_msv
    LET ls_alter_script = "alter table ",lo_CONSTRAINT.dzed001,
                          " add constraint ",lo_CONSTRAINT.dzed002,
                          " foreign key (",lo_CONSTRAINT.dzed004,") references ",lo_CONSTRAINT.dzed005," (",lo_CONSTRAINT.dzed006,")","\n","go","\n"
    &else
    &ifdef is_pgs
    LET ls_alter_script = "alter table ",lo_CONSTRAINT.dzed001,
                          " add constraint ",lo_CONSTRAINT.dzed002,
                          " foreign key (",lo_CONSTRAINT.dzed004,") references ",lo_CONSTRAINT.dzed005," (",lo_CONSTRAINT.dzed006,")",";","\n"
    &else
    LET ls_db_type = sadzi140_util_get_db_type() #20240521
    CASE ls_db_type
      WHEN cs_db_driver_type_kdb #20240521
        LET ls_alter_script = "alter table ",lo_CONSTRAINT.dzed001,
                              " add constraint ",lo_CONSTRAINT.dzed002,
                              " foreign key (",lo_CONSTRAINT.dzed004,") references ",lo_CONSTRAINT.dzed005," (",lo_CONSTRAINT.dzed006,")",";","\n"
    --WHEN cs_db_driver_type_dmg #20240521
      --與Oracle共用
    OTHERWISE   
      LET ls_alter_script = "alter table ",lo_CONSTRAINT.dzed001,
                            " add constraint ",lo_CONSTRAINT.dzed002,
                            " foreign key (",lo_CONSTRAINT.dzed004,") references ",lo_CONSTRAINT.dzed005," (",lo_CONSTRAINT.dzed006,") on delete cascade ",ls_foreign_key_state,";","\n"
    END CASE  
    &endif                          
    &endif                          
       
    LET ls_alter_sql = ls_alter_sql,ls_alter_script
    LET ls_alter_script = ""
    
    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_alter_foreign_constraint_sql
  
  FREE lpre_get_alter_foreign_constraint_sql
  FREE lcur_get_alter_foreign_constraint_sql  

  LET ls_return = ls_alter_sql
 
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_get_drop_column_sql(p_master_user,p_table_owner,p_table_name,p_object_type,p_diff_columns)
DEFINE 
  p_master_user  STRING, 
  p_table_owner  STRING, 
  p_table_name   STRING,
  p_object_type  STRING,
  p_diff_columns DYNAMIC ARRAY OF T_COLUMN_DIFF 
DEFINE
  ls_master_user   STRING, 
  ls_table_owner   STRING, 
  ls_table_name    STRING,
  ls_object_type   STRING,
  lo_diff_columns  DYNAMIC ARRAY OF T_COLUMN_DIFF,
  lo_alter_columns DYNAMIC ARRAY OF T_COLUMN_DIFF_W_COMMENT, #20161220
  lo_undrop_col_list DYNAMIC ARRAY OF T_COLUMN_LIST, #20161220
  li_undrop_length INTEGER,  #20161220
  li_loop          INTEGER, 
  li_rec_cnt       INTEGER,
  ls_columns       STRING,
  ls_sql           STRING,
  ls_where         STRING,
  ls_alter_type    STRING,
  ls_alter_script  STRING,
  ls_alter_sql     STRING,
  ls_data_type     STRING,
  ls_drp_comment   STRING, #20191209
  ls_drop_idxs     STRING, #20191209
  lb_column_exist  BOOLEAN,
  lb_drop_column   BOOLEAN, #20161220
  lb_rt_drop_backdoor_column BOOLEAN #20161220
DEFINE  
  lb_return  BOOLEAN, #20161220
  ls_return  STRING,
  lo_return  DYNAMIC ARRAY OF T_COLUMN_LIST #20161220

  LET ls_master_user    = p_master_user
  LET ls_table_owner    = p_table_owner
  LET ls_table_name     = p_table_name
  LET ls_object_type    = p_object_type
  LET lo_diff_columns.* = p_diff_columns.*

  LET ls_alter_type   = ""
  LET ls_alter_script = ""
  LET ls_alter_sql    = ""
  
  CALL lo_undrop_col_list.clear()

  #20161220 begin
  LET lb_rt_drop_backdoor_column = IIF(NVL(FGL_GETENV(cs_env_RT_DROP_BACKDOOR_COLUMN),"N")=="Y",TRUE,FALSE)
  #LET lb_rt_drop_backdoor_column = TRUE # for test
  #20161220 end
  
  FOR li_loop = 1 TO lo_diff_columns.getLength()
    LET ls_columns = ls_columns,",'",lo_diff_columns[li_loop].DZEV005,"'"
  END FOR

  LET ls_columns = ls_columns.subString(2,ls_columns.getLength())
  #LET ls_where = " AND EV.DZEV005 IN (",ls_columns,")"
  LET ls_where = " and atc.column_name in (",ls_columns.toUpperCase(),")"

  &ifdef is_msv
  LET ls_sql = "select atc.column_name                      dzev005,       ",               
               "       iif(atc.nullable='Y','N','Y')        dzev006,       ",               
               "       atc.nullable                         dzev007,       ",               
               "       atc.data_type                        dzev008,       ",               
               "       atc.data_length                      dzev009,       ",
               "       atc.column_id                        dzev021,       ",               
               "       atc.comment                          comments       ",               
               "  from dba_tab_columns atc                                 ",               
               " where 1=1                                                 ",               
               "   and not exists (                                        ",               
               "                    select 1 from dzeb_t eb                ",               
               "                     where eb.dzeb001 = atc.table_name     ",        
               "                       and eb.dzeb002 = atc.column_name    ",        
               "                  )                                        ",               
               "   and atc.owner      = '",ls_table_owner.toLowerCase(),"' ",
               "   and atc.table_name = '",ls_table_name.toLowerCase(),"'  "  
  &else
  &ifdef is_pgs
  LET ls_sql = "select atc.column_name                      dzev005,                          ",               
               "       case when atc.nullable = 'Y'                                           ", 
               "       then 'N' else 'Y' end                dzev006,                          ",               
               "       atc.nullable                         dzev007,                          ",               
               "       atc.data_type                        dzev008,                          ",               
               "       case                                                                   ",
               "         when atc.data_type = 'text' then null                                ",
               "         when atc.data_type = 'bytea' then null                               ",
               "         when atc.data_type = 'date' then null                                ",
               "         when atc.data_type = 'int2' then null                                ",
               "         when atc.data_type = 'int4' then null                                ",
               "         when atc.data_type = 'int8' then null                                ",
               "         when atc.data_type = 'decimal' then                                  ",
               "           atc.data_precision::varchar||','||                                 ",
               "           atc.data_scale::varchar                                            ",
               "       else                                                                   ",
               "         atc.data_length::varchar                                             ",
               "       end::varchar                         dzev009,                          ",
               "       atc.column_id                        dzev021,                          ",               
               "       dcc.comments                      as comments                          ",               
               "  from dba_tab_columns atc                                                    ",               
               "  left outer join (                                                           ",                            
               "                    select * from dba_col_comments acc                        ",
               "                     where acc.table_name = '",ls_table_name.toLowerCase(),"' ", 
               "                  ) dcc                                                       ",
               "               on dcc.owner       = atc.owner                                 ",        
               "              and dcc.table_name  = atc.table_name                            ",
               "              and dcc.column_name = atc.column_name                           ",
               " where 1=1                                                                    ",               
               "   and not exists (                                                           ",               
               "                    select 1 from dzeb_t eb                                   ",               
               "                     where eb.dzeb001 = atc.table_name                        ",        
               "                       and eb.dzeb002 = atc.column_name                       ",        
               "                  )                                                           ",               
               "   and atc.owner      = '",ls_table_owner.toLowerCase(),"'                    ",
               "   and atc.table_name = '",ls_table_name.toLowerCase(),"'                     "  
  &else
  LET ls_sql = "select atc.column_name                      dzev005,                          ",
               "       decode(acc.column_name,null,'N','Y') dzev006,                          ",
               "       atc.nullable                         dzev007,                          ",
               "       atc.data_type                        dzev008,                          ",
               "       to_char(decode(                                                        ",
               "               atc.data_type,                                                 ",
               "               'BLOB',4000,                                                   ", #20240529
               "               'CLOB',4000,                                                   ", #20240529
               "               'INT2',NULL,                                                   ", #20240529
               "               'INT4',NULL,                                                   ", #20240529
               "               'INT8',NULL,                                                   ", #20240529
               "               'NUMBER',decode(                                               ",
               "                                nvl(atc.data_scale,'0'),                      ",
               "                                '0',atc.data_precision,                       ",
               "                                atc.data_precision||'.'||atc.data_scale       ",
               "                              ),                                              ",
               "               'NUMERIC',DECODE(                                              ", #20240529 begin
               "                                NVL(ATC.DATA_SCALE,'0'),                      ",
               "                                '0',ATC.DATA_PRECISION,                       ",
               "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE       ", 
               "                              ),                                              ", #20240529 end
               "               'DATE',7,                                                      ", #20240529
               "               'TIMESTAMP(0)',7,                                              ", #20240529
               "               'TIMESTAMP(5)',11,                                             ", #20240529
               "               'TIMESTAMP(6)',11,                                             ", #20240529
               "               'TIMESTAMP',decode(                                            ",
               "                                   atc.data_scale,0,7,8,7,11                  ", #20240529
               "                                 ),                                           ",                
               "               'DATE',decode(                                                 ",
               "                              atc.data_length,3,7,atc.data_length             ",
               "                            ),                                                ",                
               "               atc.data_length                                                ",
               "             ))                             dzev009,                          ", #20161220
               "       atc.column_id                        dzev021,                          ", #20161220
               "       dccs.comments                        comments                          ", #20161220
               "  from dba_tab_columns atc                                                    ",
               "  left outer join dba_cons_columns acc on acc.owner       = atc.owner         ",
               "                                      and acc.table_name  = atc.table_name    ",
               "                                      and acc.column_name = atc.column_name   ",
               "                                      and acc.constraint_name like '%PK'      ",
               "  left outer join dba_col_comments dccs on dccs.owner       = atc.owner       ", #20161220
               "                                       and dccs.table_name  = atc.table_name  ", #20161220
               "                                       and dccs.column_name = atc.column_name ", #20161220
               " where 1=1                                                                    ",
               ls_where,                                                                     
               "   and not exists (                                                           ",
               "                    select 1 from dzeb_t eb                                   ",
               "                     where eb.dzeb001 = lower(atc.table_name)                 ",
               "                       and eb.dzeb002 = lower(atc.column_name)                ",
               "                  )                                                           ",  
               "   and atc.owner      = '",ls_table_owner.toUpperCase(),"'                    ",
               "   and atc.table_name = '",ls_table_name.toUpperCase(),"'                     "
  &endif
  &endif
               
  PREPARE lpre_get_drop_column_sql FROM ls_sql
  DECLARE lcur_get_drop_column_sql CURSOR FOR lpre_get_drop_column_sql

  LET li_rec_cnt = 1
  
  OPEN lcur_get_drop_column_sql
  FOREACH lcur_get_drop_column_sql INTO lo_alter_columns[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    #20161220 begin
    LET lb_drop_column = TRUE
    IF UPSHIFT(lo_alter_columns[li_rec_cnt].dzev005) <> cs_dummy_column_name THEN 
      IF NVL(lo_alter_columns[li_rec_cnt].comments,cs_null_value) <> lo_alter_columns[li_rec_cnt].dzev005 THEN
        IF NOT lb_rt_drop_backdoor_column THEN
          LET lb_drop_column = FALSE
        ELSE  
          LET li_undrop_length = lo_undrop_col_list.getLength() + 1
          LET lo_undrop_col_list[li_undrop_length].COLUMN_NAME = lo_alter_columns[li_rec_cnt].DZEV005
          LET lo_undrop_col_list[li_undrop_length].COLUMN_DESC = ls_table_owner,".",lo_alter_columns[li_rec_cnt].DZEV005  
        END IF 
      END IF
    END IF  
    #20161220 end

    IF lb_drop_column THEN #20161220 
      &ifdef is_msv
      #SQL Server Drop column 之前要先 drop comment
      LET ls_drp_comment = "if object_id('df_",ls_table_name,"_",lo_alter_columns[li_rec_cnt].DZEV005,"') is not null ","\n",
                           "begin","\n",
                           "  alter table ",ls_table_name," drop constraint df_",ls_table_name,"_",lo_alter_columns[li_rec_cnt].DZEV005,"\n",
                           "end","\n",
                           "go","\n"
        LET ls_drop_idxs = "declare                                                                     ","\n",                              
                           "  @ls_sql         nvarchar(max),                                            ","\n",                              
                           "  @ls_table_name  varchar(128) = '",ls_table_name,"',                       ","\n",                              
                           "  @ls_column_name varchar(128) = '",lo_alter_columns[li_rec_cnt].DZEV005,"' ","\n",
                           "  ------------------------------------------------------------              ","\n",
                           "  declare c_indexes cursor for                                              ","\n",
                           "    select 'drop index '+uics.table_name+'.'+uics.index_name                ","\n",
                           "    from user_ind_columns uics                                              ","\n",                             
                           "    where uics.table_name  = @ls_table_name                                 ","\n",                               
                           "      and uics.column_name = @ls_column_name                                ","\n",                               
                           "  ------------------------------------------------------------              ","\n",
                           "begin                                                                       ","\n",
                           "                                                                            ","\n",
                           "  open c_indexes                                                            ","\n",
                           "  fetch next from c_indexes into @ls_sql                                    ","\n",
                           "  while @@fetch_status <> -1                                                ","\n",
                           "  begin                                                                     ","\n",
                           "                                                                            ","\n",
                           "    --print 'SQL Command : '+@ls_sql                                        ","\n",
                           "    exec(@ls_sql)                                                           ","\n",
                           "    fetch next from c_indexes into @ls_sql                                  ","\n",
                           "                                                                            ","\n",
                           "  end                                                                       ","\n",
                           "  close c_indexes;                                                          ","\n",
                           "  deallocate c_indexes;                                                     ","\n",
                           "                                                                            ","\n",
                           "end                                                                         ","\n",                              
                           "go","\n"                                                                    
      LET ls_alter_script = "alter table ",
                            ls_table_name,
                            " drop column ",
                            lo_alter_columns[li_rec_cnt].DZEV005,"\n",
                            "go","\n"
      LET ls_alter_script = ls_drp_comment,
                            ls_drop_idxs,
                            ls_alter_script                            
      &else
      --Oracle & PostgreSQL 共用
      LET ls_alter_script = "alter table ",
                           ls_table_name,
                           " drop column ",
                           lo_alter_columns[li_rec_cnt].DZEV005,
                           ";","\n"
      &endif                     

      LET ls_alter_sql = ls_alter_sql,ls_alter_script
      LET ls_alter_script = "" #20200330
    ELSE #20161220 begin
      LET li_undrop_length = lo_undrop_col_list.getLength() + 1
      LET lo_undrop_col_list[li_undrop_length].COLUMN_NAME = lo_alter_columns[li_rec_cnt].DZEV005
      LET lo_undrop_col_list[li_undrop_length].COLUMN_DESC = ls_table_owner,".",lo_alter_columns[li_rec_cnt].DZEV005  
    END IF #20161220 end
      
    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_drop_column_sql
  CALL lo_alter_columns.deleteElement(li_rec_cnt)
  
  FREE lpre_get_drop_column_sql
  FREE lcur_get_drop_column_sql  

  LET lb_return = lb_rt_drop_backdoor_column #20161220
  LET ls_return = ls_alter_sql
  LET lo_return = lo_undrop_col_list #20161220
 
  RETURN lb_return,ls_return,lo_return #20161220
  
END FUNCTION

#檢核實際表格Column值是否存在
FUNCTION sadzi140_db_check_db_column_exist_by_single_column(p_owner,p_table_name,p_column_name)
DEFINE
  p_owner       STRING,
  p_table_name  STRING,
  p_column_name STRING
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_column_name STRING,
  ls_sql         STRING,
  li_rec_count   INTEGER,
  lb_return      BOOLEAN

  &ifdef is_msv
  LET ls_owner       = p_owner.toLowerCase()
  LET ls_table_name  = p_table_name.toLowerCase()
  LET ls_column_name = p_column_name.toLowerCase()
  &else
  &ifdef is_pgs
  LET ls_owner       = p_owner.toLowerCase()
  LET ls_table_name  = p_table_name.toLowerCase()
  LET ls_column_name = p_column_name.toLowerCase()
  &else
  LET ls_owner       = p_owner.toUpperCase()
  LET ls_table_name  = p_table_name.toUpperCase()
  LET ls_column_name = p_column_name.toUpperCase()
  &endif
  &endif

  #取得資料筆數
  LET ls_sql = "select count(1)                               ",
               "  from dba_tab_columns dtc                    ",
               " where dtc.owner       = '",ls_owner,"'       ",
               "   and dtc.table_name  = '",ls_table_name,"'  ",
               "   and dtc.column_name = '",ls_column_name,"' "
 
  PREPARE lpre_check_db_column_exist FROM ls_sql
  DECLARE lcur_check_db_column_exist CURSOR FOR lpre_check_db_column_exist
  OPEN lcur_check_db_column_exist
  FETCH lcur_check_db_column_exist INTO li_rec_count
  CLOSE lcur_check_db_column_exist
  FREE lcur_check_db_column_exist
  FREE lpre_check_db_column_exist

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_db_check_db_column_exist(p_owner,p_table_name,p_column_name)
DEFINE
  p_owner       STRING,
  p_table_name  STRING,
  p_column_name STRING
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_column_name STRING,
  lb_rec_exist   BOOLEAN,
  li_loop        INTEGER,
  lb_return      BOOLEAN

  &ifdef is_msv
  LET ls_owner       = p_owner.toLowerCase()
  LET ls_table_name  = p_table_name.toLowerCase()
  LET ls_column_name = p_column_name.toLowerCase()
  &else
  &ifdef is_pgs
  LET ls_owner       = p_owner.toLowerCase()
  LET ls_table_name  = p_table_name.toLowerCase()
  LET ls_column_name = p_column_name.toLowerCase()
  &else
  LET ls_owner       = p_owner.toUpperCase()
  LET ls_table_name  = p_table_name.toUpperCase()
  LET ls_column_name = p_column_name.toUpperCase()
  &endif
  &endif

  LET lb_rec_exist = FALSE
  
  FOR li_loop = 1 TO mo_columns_data_in_mem.getLength()
    IF mo_columns_data_in_mem[li_loop].OWNER       = ls_owner AND
       mo_columns_data_in_mem[li_loop].TABLE_NAME  = ls_table_name AND
       mo_columns_data_in_mem[li_loop].COLUMN_NAME = ls_column_name THEN

       LET lb_rec_exist = TRUE
       
       EXIT FOR
       
    END IF   
  END FOR  
  
  LET lb_return = lb_rec_exist
  
  RETURN lb_return
  
END FUNCTION

#檢核實際表格 Key 值是否存在
FUNCTION sadzi140_db_check_table_constraint_exist(p_owner,p_table_name,p_constraint_name)
DEFINE
  p_owner           STRING,
  p_table_name      STRING,
  p_constraint_name STRING
DEFINE
  ls_owner           STRING,
  ls_table_name      STRING,
  ls_constraint_name STRING,
  ls_sql             STRING,
  li_rec_count       INTEGER,
  lb_return          BOOLEAN

  &ifdef is_msv
  LET ls_owner           = NVL(p_owner.toLowerCase(),cs_master_user.toLowerCase())
  LET ls_table_name      = p_table_name.toLowerCase()
  LET ls_constraint_name = p_constraint_name.toLowerCase()
  &else
  &ifdef is_pgs
  LET ls_owner           = NVL(p_owner.toLowerCase(),cs_master_user.toLowerCase())
  LET ls_table_name      = p_table_name.toLowerCase()
  LET ls_constraint_name = p_constraint_name.toLowerCase()
  &else
  LET ls_owner           = NVL(p_owner.toUpperCase(),cs_master_user.toUpperCase())
  LET ls_table_name      = p_table_name.toUpperCase()
  LET ls_constraint_name = p_constraint_name.toUpperCase()
  &endif
  &endif

  #取得資料筆數
  LET ls_sql = "select count(1)                                       ",
               "  from all_constraints acs                            ",
               " where acs.owner           = '",ls_owner,"'           ",
               "   and acs.table_name      = '",ls_table_name,"'      ",
               "   and acs.constraint_name = '",ls_constraint_name,"' " 
 
  PREPARE lpre_check_table_constraint_exist FROM ls_sql
  DECLARE lcur_check_table_constraint_exist CURSOR FOR lpre_check_table_constraint_exist
  OPEN lcur_check_table_constraint_exist
  FETCH lcur_check_table_constraint_exist INTO li_rec_count
  CLOSE lcur_check_table_constraint_exist
  FREE lcur_check_table_constraint_exist
  FREE lpre_check_table_constraint_exist

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION

#檢核實際表格Synonym是否存在
FUNCTION sadzi140_db_check_table_synonym_exist(p_owner,p_table_name)
DEFINE
  p_owner       STRING,
  p_table_name  STRING
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_sql         STRING,
  li_rec_count   INTEGER,
  lb_return      BOOLEAN

  &ifdef is_msv
  LET ls_owner      = NVL(p_owner.toLowerCase(),cs_master_user.toLowerCase())
  LET ls_table_name = p_table_name.toLowerCase()
  &else
  &ifdef is_pgs
  LET ls_owner      = NVL(p_owner.toLowerCase(),cs_master_user.toLowerCase())
  LET ls_table_name = p_table_name.toLowerCase()
  &else
  LET ls_owner      = NVL(p_owner.toUpperCase(),cs_master_user.toUpperCase())
  LET ls_table_name = p_table_name.toUpperCase()
  &endif
  &endif

  #取得資料筆數
  LET ls_sql = "select count(1)                             ",
               "  from dba_objects dos                      ",
               " where dos.owner       = '",ls_owner,"'     ",
               "   and dos.object_type = 'SYNONYM'          ",
               "   and dos.object_name = '",ls_table_name,"'"

  PREPARE lpre_check_table_synonym_exist FROM ls_sql
  DECLARE lcur_check_table_synonym_exist CURSOR FOR lpre_check_table_synonym_exist
  OPEN lcur_check_table_synonym_exist
  FETCH lcur_check_table_synonym_exist INTO li_rec_count
  CLOSE lcur_check_table_synonym_exist
  FREE lcur_check_table_synonym_exist
  FREE lpre_check_table_synonym_exist

  DISPLAY li_rec_count
  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION

#檢核實際表格 Index 是否存在
FUNCTION sadzi140_db_check_table_index_exist(p_owner,p_index_name)
DEFINE
  p_owner       STRING,
  p_index_name  STRING
DEFINE
  ls_owner      STRING,
  ls_index_name STRING,
  ls_sql        STRING,
  li_rec_count  INTEGER,
  lb_return     BOOLEAN

  &ifdef is_msv
  LET ls_owner      = NVL(p_owner.toLowerCase(),cs_master_user.toLowerCase())
  LET ls_index_name = p_index_name.toLowerCase()
  &else
  &ifdef is_pgs
  LET ls_owner      = NVL(p_owner.toLowerCase(),cs_master_user.toLowerCase())
  LET ls_index_name = p_index_name.toLowerCase()
  &else
  LET ls_owner      = NVL(p_owner.toUpperCase(),cs_master_user.toUpperCase())
  LET ls_index_name = p_index_name.toUpperCase()
  &endif
  &endif

  #取得資料筆數
  LET ls_sql = "select count(1)                            ",
               "  from dba_indexes ai                      ",
               " where ai.owner = '",ls_owner,"'           ",
               "   and ai.index_name = '",ls_index_name,"' " 
 
  PREPARE lpre_check_table_index_exist FROM ls_sql
  DECLARE lcur_check_table_index_exist CURSOR FOR lpre_check_table_index_exist
  OPEN lcur_check_table_index_exist
  FETCH lcur_check_table_index_exist INTO li_rec_count
  CLOSE lcur_check_table_index_exist
  FREE lcur_check_table_index_exist
  FREE lpre_check_table_index_exist

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION

#檢核 Table Designer Index 資料是否存在
FUNCTION sadzi140_db_check_designer_index_exist(p_index_name)
DEFINE
  p_index_name  STRING
DEFINE
  ls_index_name STRING,
  ls_sql        STRING,
  li_rec_count  INTEGER,
  lb_return     BOOLEAN

  LET ls_index_name = p_index_name

  #取得資料筆數
  LET ls_sql = "select count(1)                         ",
               "  from dzec_t ec                        ",
               " where ec.dzec002 = '",ls_index_name,"' "  
 
  PREPARE lpre_check_designer_index_exist FROM ls_sql
  DECLARE lcur_check_designer_index_exist CURSOR FOR lpre_check_designer_index_exist
  OPEN lcur_check_designer_index_exist
  FETCH lcur_check_designer_index_exist INTO li_rec_count
  CLOSE lcur_check_designer_index_exist
  FREE lcur_check_designer_index_exist
  FREE lpre_check_designer_index_exist

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION

#以亂數產出副檔名
FUNCTION sadzi140_db_get_GUID()
DEFINE
  lr_random_name RECORD
                   segment1 STRING,
                   segment2 STRING,
                   segment3 STRING,
                   segment4 STRING
                 END RECORD
DEFINE 
  li_random_value   INTEGER,
  li_max_random_num INTEGER,
  ls_GUID           STRING,
  ls_using_format   STRING
DEFINE  
  ls_return       STRING

  {
  LET li_max_random_num = 9999
  LET ls_using_format  = "&&&&"
  
  CALL util.math.rand(li_max_random_num) RETURNING li_random_value
  LET lr_random_name.segment1 = li_random_value USING ls_using_format
  CALL util.math.rand(li_max_random_num) RETURNING li_random_value
  LET lr_random_name.segment2 = li_random_value USING ls_using_format
  CALL util.math.rand(li_max_random_num) RETURNING li_random_value
  LET lr_random_name.segment3 = li_random_value USING ls_using_format
  CALL util.math.rand(li_max_random_num) RETURNING li_random_value
  LET lr_random_name.segment4 = li_random_value USING ls_using_format
  
  LET ls_final_name = lr_random_name.segment1,".",
                      lr_random_name.segment2,".",
                      lr_random_name.segment3,".",
                      lr_random_name.segment4
  }

  CALL security.RandomGenerator.CreateUUIDString() RETURNING ls_GUID 

  LET ls_return = ls_GUID
  
  RETURN ls_return                     
  
END FUNCTION

FUNCTION sadzi140_db_get_parameter(p_level,p_param)
DEFINE
  p_level STRING,
  p_param STRING
DEFINE
  ls_level     STRING,
  ls_param     STRING,
  ls_sql       STRING,
  lv_parameter VARCHAR(1024) 
DEFINE
  ls_return    STRING

  LET ls_level = p_level
  LET ls_param = p_param

  LET ls_sql = "select ej.dzej005                        ",
               "  from dzej_t ej                         ",
               " where ej.dzej001 = 'adzi140_parameters' ",
               "   and ej.dzej003 = '",ls_level,"'       ",
               "   and ej.dzej004 = '",ls_param,"'       "
                              
  PREPARE lpre_get_parameter FROM ls_sql
  DECLARE lcur_get_parameter CURSOR FOR lpre_get_parameter

  OPEN lcur_get_parameter
  FETCH lcur_get_parameter INTO lv_parameter
  CLOSE lcur_get_parameter
  
  FREE lpre_get_parameter
  FREE lcur_get_parameter  

  #20191209 begin
  &ifdef is_msv
  CASE 
    WHEN ls_level = cs_param_level_system 
      IF ls_param = cs_param_master_user THEN 
        LET lv_parameter = FGL_GETENV("ZONE"),"_",lv_parameter
      END IF
  END CASE
  &endif
  #20191209 end
  
  LET ls_return = lv_parameter
  
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_get_table_in_db_type(p_table_name,p_order_type)
DEFINE 
  p_table_name STRING,
  p_order_type STRING
DEFINE
  ls_table_name       STRING,
  ls_order_type       STRING,
  ls_sql              STRING,
  li_rec_cnt          INTEGER,         
  lo_table_in_db_type DYNAMIC ARRAY OF T_TABLE_IN_DB_TYPE
DEFINE  
  lo_return DYNAMIC ARRAY OF T_TABLE_IN_DB_TYPE

  LET ls_table_name = p_table_name
  LET ls_order_type = p_order_type
  
  LET li_rec_cnt = 1
  
  LET ls_sql = "select eu.dzeu002, eu.dzeu003                  ",
               "  from dzeu_t eu                               ",
               " where eu.dzeu001 = '",ls_table_name,"'        ",
               "   and eu.dzeu004 = 'Y'                        ",
               "   and exists (                                ", #20200330 begin
               "                select 1 from gzda_t da        ",
               "                 where da.gzda001 = eu.dzeu002 ", 
               "                   and da.gzda005 = 'Y'        ",
               "              )                                ", #20200330 end
               " order by eu.dzeu005 ",ls_order_type,"         "
                              
  PREPARE lpre_get_table_or_synonym FROM ls_sql
  DECLARE lcur_get_table_or_synonym CURSOR FOR lpre_get_table_or_synonym

  OPEN lcur_get_table_or_synonym
  FOREACH lcur_get_table_or_synonym INTO lo_table_in_db_type[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_table_or_synonym
  CALL lo_table_in_db_type.deleteElement(li_rec_cnt)
  
  FREE lpre_get_table_or_synonym
  FREE lcur_get_table_or_synonym  

  LET lo_return.* = lo_table_in_db_type.*
  
  RETURN lo_return
  
END FUNCTION

FUNCTION sadzi140_db_get_main_table_synonym_data(p_main_table)
DEFINE
  p_main_table STRING
DEFINE
  ls_main_table  STRING,
  lo_schema_type DYNAMIC ARRAY OF T_TABLE_SYNONYM,
  li_rec_cnt     INTEGER,
  ls_sql         STRING,
  lo_return      DYNAMIC ARRAY OF T_TABLE_SYNONYM

  LET ls_main_table = p_main_table

  LET li_rec_cnt = 1

  LET ls_sql = "select eu.dzeu005,'',eu.dzeu001,'',eu.dzeu004,eu.dzeu002,eu.dzeu003 ", 
               "  from dzeu_t eu                                                    ", 
               " where eu.dzeu001 = '",ls_main_table,"'                             ", 
               " order by eu.dzeu005                                                "
               
  PREPARE lpre_get_main_table_synonym_data FROM ls_sql
  DECLARE lcur_get_main_table_synonym_data CURSOR FOR lpre_get_main_table_synonym_data

  OPEN lcur_get_main_table_synonym_data
  FOREACH lcur_get_main_table_synonym_data INTO lo_schema_type[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_main_table_synonym_data
  CALL lo_schema_type.deleteElement(li_rec_cnt)
  
  FREE lpre_get_main_table_synonym_data
  FREE lcur_get_main_table_synonym_data  

  LET lo_return.* = lo_schema_type.*

  RETURN lo_return
  
END FUNCTION

FUNCTION sadzi140_db_check_table_exist(p_owner,p_table_name)
DEFINE
  p_owner      STRING,
  p_table_name STRING
DEFINE
  ls_owner      STRING,
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_count  INTEGER,
  lb_return     BOOLEAN

  &ifdef is_msv
  LET ls_owner      = NVL(p_owner.toLowerCase(),cs_master_user.toLowerCase())
  LET ls_table_name = p_table_name.toLowerCase()
  &else
  &ifdef is_pgs
  LET ls_owner      = NVL(p_owner.toLowerCase(),cs_master_user.toLowerCase())
  LET ls_table_name = p_table_name.toLowerCase()
  &else
  LET ls_owner      = NVL(p_owner.toUpperCase(),cs_master_user.toUpperCase())
  LET ls_table_name = p_table_name.toUpperCase()
  &endif
  &endif

  #取得資料筆數
  LET ls_sql = "select count(1)                             ",
               "  from dba_objects dos                      ",
               " where dos.owner       = '",ls_owner,"'     ",
               "   and dos.object_type = 'TABLE'            ",
               "   and dos.object_name = '",ls_table_name,"'"

  PREPARE lpre_check_table_exist FROM ls_sql
  DECLARE lcur_check_table_exist CURSOR FOR lpre_check_table_exist
  OPEN lcur_check_table_exist
  FETCH lcur_check_table_exist INTO li_rec_count
  CLOSE lcur_check_table_exist
  FREE lcur_check_table_exist
  FREE lpre_check_table_exist

  --DISPLAY li_rec_count
  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_db_get_create_table_info(p_table_name)
DEFINE 
  p_table_name STRING
DEFINE
  ls_table_name        STRING,
  ls_sql               STRING,
  li_rec_cnt           INTEGER,         
  lo_table_information T_DZEA_CREATE_TABLE
DEFINE  
  lo_return T_DZEA_CREATE_TABLE

  LET ls_table_name = p_table_name
  LET li_rec_cnt = 1
  
  LET ls_sql = "select ea.dzea003,ea.dzea004,ea.dzea001,ea.dzea002,ea.dzea005, ",
               "       ea.dzea006,ea.dzea007,ea.dzea008,ea.dzea009,ea.dzea010, ",
               "       '',''                                                   ",
               "  from dzea_t ea                                               ",
               " where ea.dzea001 = '",ls_table_name,"'                        " 
                               
  PREPARE lpre_get_table_infomation FROM ls_sql
  DECLARE lcur_get_table_infomation CURSOR FOR lpre_get_table_infomation

  OPEN lcur_get_table_infomation
  FETCH lcur_get_table_infomation INTO lo_table_information.*
  CLOSE lcur_get_table_infomation
  
  FREE lpre_get_table_infomation
  FREE lcur_get_table_infomation  

  LET lo_return.* = lo_table_information.*
  
  RETURN lo_return.*
  
END FUNCTION

FUNCTION sadzi140_db_get_module_leading_code(p_module_name,p_table_name)
DEFINE 
  p_module_name STRING,
  p_table_name  STRING
DEFINE
  ls_module_name  STRING,
  ls_table_name   STRING,
  ls_sql          STRING,
  li_rec_cnt      INTEGER,
  ls_leading_code VARCHAR(10)         
DEFINE  
  ls_return STRING

  LET ls_module_name = p_module_name
  LET ls_table_name  = p_table_name

  LET li_rec_cnt = 1

  &ifdef is_msv
  LET ls_sql = "select zt.gzzt002                                                            ",
               "  from gzzt_t zt                                                             ",
               " where zt.gzzt001 = (                                                        ",
               "                      select zj.gzzj003                                      ",
               "                        from gzzj_t zj                                       ",
               "                       where zj.gzzj001 = '",ls_module_name,"'               ",
               "                    )                                                        ",
               "   and zt.gzzt002 like substring('",ls_table_name,"',1,len(zt.gzzt002))||'%' "
  &else
  &ifdef is_pgs
  LET ls_sql = "select zt.gzzt002                                                               ",
               "  from gzzt_t zt                                                                ",
               " where zt.gzzt001 = (                                                           ",
               "                      select zj.gzzj003                                         ",
               "                        from gzzj_t zj                                          ",
               "                       where zj.gzzj001 = '",ls_module_name,"'                  ",
               "                    )                                                           ",
               "   and zt.gzzt002 like substring('",ls_table_name,"',1,length(zt.gzzt002))||'%' "
  &else
  LET ls_sql = "select zt.gzzt002                                                              ",
               "  from gzzt_t zt                                                               ",
               " where zt.gzzt001 = (                                                          ",
               "                      select zj.gzzj003                                        ",
               "                        from gzzj_t zj                                         ",
               "                       where zj.gzzj001 = '",ls_module_name,"'                 ",
               "                    )                                                          ",
               "   and zt.gzzt002 like substrb('",ls_table_name,"',1,lengthb(zt.gzzt002))||'%' "
  &endif             
  &endif             
                                              
  PREPARE lpre_get_module_leading_code FROM ls_sql
  DECLARE lcur_get_module_leading_code CURSOR FOR lpre_get_module_leading_code

  OPEN lcur_get_module_leading_code
  FETCH lcur_get_module_leading_code INTO ls_leading_code
  CLOSE lcur_get_module_leading_code
  
  FREE lpre_get_module_leading_code
  FREE lcur_get_module_leading_code  

  IF NVL(ls_leading_code,cs_null_value) = cs_null_value THEN

    LET ls_sql = "select zt.gzzt002                                              ",
                 "  from gzzt_t zt                                               ",
                 " where zt.gzzt001 = (                                          ",
                 "                      select zj.gzzj003                        ",
                 "                        from gzzj_t zj                         ",
                 "                       where zj.gzzj001 = '",ls_module_name,"' ",
                 "                    )                                          "
                                                
    PREPARE lpre_get_module_only_leading_code FROM ls_sql
    DECLARE lcur_get_module_only_leading_code CURSOR FOR lpre_get_module_only_leading_code

    OPEN lcur_get_module_only_leading_code
    FETCH lcur_get_module_only_leading_code INTO ls_leading_code
    CLOSE lcur_get_module_only_leading_code
    
    FREE lpre_get_module_only_leading_code
    FREE lcur_get_module_only_leading_code  
    
  END IF

  LET ls_return = ls_leading_code
  
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_get_module_code_by_dgenv(p_module_name,p_dgenv)
DEFINE 
  p_module_name STRING,
  p_dgenv       STRING
DEFINE
  ls_module_name  STRING,
  ls_dgenv        STRING,
  ls_sql          STRING,
  li_rec_cnt      INTEGER,
  ls_module_code  VARCHAR(10)         
DEFINE  
  ls_return STRING

  LET ls_module_name = p_module_name
  LET ls_dgenv       = p_dgenv

  LET li_rec_cnt = 1

  IF ls_dgenv = cs_dgenv_standard THEN
    LET ls_sql = "select distinct zj.gzzj003                ",
                 "  from gzzj_t zj                          ",
                 " where zj.gzzj001 = '",ls_module_name,"'  "
  ELSE 
    LET ls_sql = "select zj.gzzj001                         ",
                 "  from gzzj_t zj                          ",
                 " where zj.gzzj001 = zj.gzzj003            ",
                 "   and zj.gzzj003 = '",ls_module_name,"'  ",
                 "   and 's' = '",ls_dgenv.toLowerCase(),"' ",
                 "union                                     ",
                 "select zj.gzzj001                         ",
                 "  from gzzj_t zj                          ",
                 " where zj.gzzj001 <> zj.gzzj003           ",
                 "   and zj.gzzj003 = '",ls_module_name,"'  ",
                 "   and 'c' = '",ls_dgenv.toLowerCase(),"' "
  END IF               
                                              
  PREPARE lpre_get_module_code_by_dgenv FROM ls_sql
  DECLARE lcur_get_module_code_by_dgenv CURSOR FOR lpre_get_module_code_by_dgenv

  OPEN lcur_get_module_code_by_dgenv
  FETCH lcur_get_module_code_by_dgenv INTO ls_module_code
  CLOSE lcur_get_module_code_by_dgenv
  
  FREE lpre_get_module_code_by_dgenv
  FREE lcur_get_module_code_by_dgenv  

  LET ls_return = NVL(ls_module_code,p_module_name)
  
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_set_key_data(p_table_name,p_key_list)
DEFINE
  p_table_name STRING,
  p_key_list   STRING
DEFINE
  lo_t_dzed_t RECORD
                dzed001 LIKE dzed_t.dzed001,
                dzed002 LIKE dzed_t.dzed002,
                dzed003 LIKE dzed_t.dzed003,
                dzed004 LIKE dzed_t.dzed004,
                dzed005 LIKE dzed_t.dzed005,
                dzed006 LIKE dzed_t.dzed006,
                dzed007 LIKE dzed_t.dzed007,
                dzed008 LIKE dzed_t.dzed008,
                dzed009 LIKE dzed_t.dzed009,
                dzed010 LIKE dzed_t.dzed010
              END RECORD
DEFINE
  ls_table_name      STRING,
  ls_key_list        STRING,
  ls_pre_column_name STRING,
  ls_key_name        STRING,
  lb_chk_key         BOOLEAN,
  ls_dgenv           VARCHAR(100),
  ls_user            VARCHAR(100),
  ldt_datetime       DATETIME YEAR TO SECOND
DEFINE
  lb_return BOOLEAN
  
  LET ls_table_name = p_table_name
  LET ls_key_list   = p_key_list --.subString(1,p_key_list.getLength())

  LET ls_DGENV  = FGL_GETENV("DGENV")
  LET lb_return = TRUE

  &ifndef DEBUG
  LET ls_user = g_user
  LET ldt_datetime = cl_get_current()
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  LET ldt_datetime = CURRENT YEAR TO SECOND
  &endif
  
  LET ls_pre_column_name = NVL(ls_table_name.subString(1,ls_table_name.getIndexOf("_",1)-1),ls_table_name)
  LET ls_key_name = ls_pre_column_name,cs_pk_keyword

  LET lo_t_dzed_t.dzed001 = ls_table_name 
  LET lo_t_dzed_t.dzed002 = ls_key_name 
  LET lo_t_dzed_t.dzed003 = cs_primary_key_type 
  LET lo_t_dzed_t.dzed004 = ls_key_list 
  LET lo_t_dzed_t.dzed005 = ""
  LET lo_t_dzed_t.dzed006 = ""
  LET lo_t_dzed_t.dzed007 = NVL(lo_t_dzed_t.dzed007,"N") #20191209
  LET lo_t_dzed_t.dzed008 = NVL(lo_t_dzed_t.dzed008,ls_dgenv) #20191209 
  LET lo_t_dzed_t.dzed009 = NVL(lo_t_dzed_t.dzed009,ls_dgenv) #20191209
  LET lo_t_dzed_t.dzed010 = NVL(lo_t_dzed_t.dzed010,"N") #20191209

  BEGIN WORK
  IF ls_key_list IS NOT NULL THEN
    TRY
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    insert into dzed_t(
                        dzed001,dzed002,dzed003,dzed004,dzed005,
                        dzed006,dzed007,dzed008,dzed009,dzed010,
                        dzedcrtid,dzedcrtdt,dzedmodid,dzedmoddt
                      )
                 VALUES (
                          lo_t_dzed_t.dzed001,lo_t_dzed_t.dzed002,lo_t_dzed_t.dzed003,lo_t_dzed_t.dzed004,lo_t_dzed_t.dzed005,
                          lo_t_dzed_t.dzed006,lo_t_dzed_t.dzed007,lo_t_dzed_t.dzed008,lo_t_dzed_t.dzed009,lo_t_dzed_t.dzed010,
                          ls_user,ldt_datetime,ls_user,ldt_datetime
                        )
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      ROLLBACK WORK
      BEGIN WORK
      #DISPLAY cs_warning_tag,"Key Duplicate, update it."
      TRY 
        &ifdef is_pgs
        SAVEPOINT sp_tran
        &endif
        UPDATE dzed_t                            
           SET dzed004 = lo_t_dzed_t.dzed004,
               dzed005 = lo_t_dzed_t.dzed005,          
               dzed006 = lo_t_dzed_t.dzed006,
               #dzed007 = lo_t_dzed_t.dzed007,          
               #dzed008 = lo_t_dzed_t.dzed008,
               #dzed009 = lo_t_dzed_t.dzed009,          
               #dzed010 = lo_t_dzed_t.dzed010          
               dzedmodid = ls_user,
               dzedmoddt = ldt_datetime          
         WHERE dzed001 = lo_t_dzed_t.dzed001     
           AND dzed002 = lo_t_dzed_t.dzed002     
      CATCH
        &ifdef is_pgs
        ROLLBACK WORK TO SAVEPOINT sp_tran
        &endif
        ROLLBACK WORK
        BEGIN WORK
        LET lb_return = FALSE
        DISPLAY cs_error_tag,"Key update fail."
      END TRY    
    END TRY
  ELSE
    CALL sadzi140_db_check_key_data(lo_t_dzed_t.dzed001,lo_t_dzed_t.dzed002) RETURNING lb_chk_key
    IF lb_chk_key THEN 
      TRY 
        &ifdef is_pgs
        SAVEPOINT sp_tran
        &endif
        DELETE FROM dzed_t
         WHERE dzed001 = lo_t_dzed_t.dzed001
           AND dzed002 = lo_t_dzed_t.dzed002
      CATCH
        &ifdef is_pgs
        ROLLBACK WORK TO SAVEPOINT sp_tran
        &endif
        ROLLBACK WORK
        BEGIN WORK
        LET lb_return = FALSE
        DISPLAY cs_error_tag,"Key Delete fail."
      END TRY    
    END IF      
  END IF  
  
  COMMIT WORK

  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_db_set_index_data(p_table_name,p_key_list)
DEFINE
  p_table_name STRING,
  p_key_list   STRING
DEFINE
  lo_t_dzec_t RECORD
                dzec001  LIKE dzec_t.dzec001,
                dzec002  LIKE dzec_t.dzec002,
                dzec003  LIKE dzec_t.dzec003,
                dzec004  LIKE dzec_t.dzec004,
                dzec005  LIKE dzec_t.dzec005,
                dzec006  LIKE dzec_t.dzec006,
                dzec007  LIKE dzec_t.dzec007,
                dzec008  LIKE dzec_t.dzec008,
                dzecstus LIKE dzec_t.dzecstus #20170912
              END RECORD
DEFINE
  ls_table_name      STRING,
  ls_key_list        STRING,
  ls_pre_column_name STRING,
  ls_key_name        STRING,
  lb_chk_index       BOOLEAN,
  ls_dgenv           VARCHAR(100),
  ls_user            VARCHAR(100),
  ldt_datetime       DATETIME YEAR TO SECOND
DEFINE
  lb_return  BOOLEAN
  
  LET ls_table_name = p_table_name
  LET ls_key_list   = p_key_list --.subString(1,p_key_list.getLength())

  LET ls_DGENV  = FGL_GETENV("DGENV")
  LET lb_return = TRUE

  &ifndef DEBUG
  LET ls_user = g_user
  LET ldt_datetime = cl_get_current()
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  LET ldt_datetime = CURRENT YEAR TO SECOND
  &endif
  
  LET ls_pre_column_name = NVL(ls_table_name.subString(1,ls_table_name.getIndexOf("_",1)-1),ls_table_name)
  LET ls_key_name = ls_pre_column_name,cs_pk_keyword

  LET lo_t_dzec_t.dzec001  = ls_table_name 
  LET lo_t_dzec_t.dzec002  = ls_key_name 
  LET lo_t_dzec_t.dzec003  = cs_unique_index_type 
  LET lo_t_dzec_t.dzec004  = ls_key_list 
  LET lo_t_dzec_t.dzec005  = NVL(lo_t_dzec_t.dzec005,"N") #20191209
  LET lo_t_dzec_t.dzec006  = NVL(lo_t_dzec_t.dzec006,ls_dgenv) #20191209
  LET lo_t_dzec_t.dzec007  = NVL(lo_t_dzec_t.dzec007,ls_dgenv) #20191209
  LET lo_t_dzec_t.dzec008  = NVL(lo_t_dzec_t.dzec008,"N") #20191209
  LET lo_t_dzec_t.dzecstus = NVL(lo_t_dzec_t.dzecstus,"Y") #20191209
  
  BEGIN WORK
  IF ls_key_list IS NOT NULL THEN
    TRY
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      insert into dzec_t(
                          dzec001,dzec002,dzec003,dzec004,dzec005,
                          dzec006,dzec007,dzec008,dzecstus, #20170912 add dzecstus
                          dzeccrtid,dzeccrtdt,dzecmodid,dzecmoddt
                        )
                 values (
                          lo_t_dzec_t.dzec001,lo_t_dzec_t.dzec002,lo_t_dzec_t.dzec003,lo_t_dzec_t.dzec004,lo_t_dzec_t.dzec005,
                          lo_t_dzec_t.dzec006,lo_t_dzec_t.dzec007,lo_t_dzec_t.dzec008,lo_t_dzec_t.dzecstus, #20170912 add dzecstus
                          ls_user,ldt_datetime,ls_user,ldt_datetime
                        )
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      ROLLBACK WORK
      BEGIN WORK
      #DISPLAY cs_warning_tag,"Key Duplicate, update it."
      TRY 
        &ifdef is_pgs
        SAVEPOINT sp_tran
        &endif
        update dzec_t                            
           set dzec004 = lo_t_dzec_t.dzec004,
               #dzec005 = lo_t_dzec_t.dzec005,
               #dzec006 = lo_t_dzec_t.dzec006,
               #dzec007 = lo_t_dzec_t.dzec007,
               #dzec008 = lo_t_dzec_t.dzec008,
               dzecmodid = ls_user,
               dzecmoddt = ldt_datetime               
         where dzec001 = lo_t_dzec_t.dzec001     
           and dzec002 = lo_t_dzec_t.dzec002     
      CATCH
        &ifdef is_pgs
        ROLLBACK WORK TO SAVEPOINT sp_tran
        &endif
        ROLLBACK WORK
        BEGIN WORK
        LET lb_return = FALSE
        DISPLAY cs_error_tag,"Index update fail."
      END TRY    
    END TRY
  ELSE
    CALL sadzi140_db_check_index_data(lo_t_dzec_t.dzec001,lo_t_dzec_t.dzec002) RETURNING lb_chk_index
    IF lb_chk_index THEN 
      TRY 
        &ifdef is_pgs
        SAVEPOINT sp_tran
        &endif
        DELETE FROM dzec_t
         WHERE dzec001 = lo_t_dzec_t.dzec001
           AND dzec002 = lo_t_dzec_t.dzec002
      CATCH
        &ifdef is_pgs
        ROLLBACK WORK TO SAVEPOINT sp_tran
        &endif
        ROLLBACK WORK
        BEGIN WORK
        LET lb_return = FALSE
        DISPLAY cs_error_tag,"Index delete fail."
      END TRY    
    END IF      
  END IF  
  
  COMMIT WORK

  RETURN lb_return
  
END FUNCTION

#檢核表格是否存在
FUNCTION sadzi140_db_check_table_exists(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  lb_return     BOOLEAN,
  ls_sql        STRING,
  li_rec_count  INTEGER,
  ls_table_name STRING

  &ifdef is_msv
  LET ls_table_name = p_table_name.toLowerCase()
  &else
  &ifdef is_pgs
  LET ls_table_name = p_table_name.toLowerCase()
  &else
  LET ls_table_name = p_table_name.toUpperCase()
  &endif
  &endif

  #取得資料筆數
  LET ls_sql = "select count (*)                            ",
               "  from dba_tables atb                       ",
               " where atb.table_name = '",ls_table_name,"' "
  
  PREPARE lpre_check_table_exists FROM ls_sql
  DECLARE lcur_check_table_exists CURSOR FOR lpre_check_table_exists
  OPEN lcur_check_table_exists
  FETCH lcur_check_table_exists INTO li_rec_count
  CLOSE lcur_check_table_exists
  FREE lcur_check_table_exists
  FREE lpre_check_table_exists

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION

#檢核表格 Key 值是否存在
FUNCTION sadzi140_db_check_key_data(p_table_name,p_key_name)
DEFINE
  p_table_name  STRING,
  p_key_name    STRING
DEFINE
  lb_return     BOOLEAN,
  ls_sql        STRING,
  li_rec_count  INTEGER,
  ls_table_name STRING,
  ls_key_name   STRING

  LET ls_table_name = p_table_name
  LET ls_key_name   = p_key_name

  #取得資料筆數
  LET ls_sql = "select count(1)                        ",
               "  from dzed_t ed                       ",
               " where 1=1                             ",
               "   and ed.dzed001 = '",ls_table_name,"'", 
               "   and ed.dzed002 = '",ls_key_name,"'  " 
 
  PREPARE lpre_check_key_data FROM ls_sql
  DECLARE lcur_check_key_data CURSOR FOR lpre_check_key_data
  OPEN lcur_check_key_data
  FETCH lcur_check_key_data INTO li_rec_count
  CLOSE lcur_check_key_data
  FREE lcur_check_key_data
  FREE lpre_check_key_data

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION

#取得Foreign Key List
FUNCTION sadzi140_db_get_foreign_key_list(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  lb_return      BOOLEAN,
  ls_sql         STRING,
  li_rec_count   INTEGER,
  ls_column_list STRING,
  ls_column_name VARCHAR(50),
  ls_table_name  STRING

  LET ls_table_name = p_table_name
  
  LET li_rec_count = 0
  LET ls_column_list = ""

  #取得 Foreign Key 清單
  LET ls_sql = "select ed.dzed001                      ",
               "  from dzed_t ed                       ",
               " where 1=1                             ",
               "   and ed.dzed003 = 'R'                ",
               "   and ed.dzed005 = '",ls_table_name,"'" 
 
  PREPARE lpre_get_foreign_key_list FROM ls_sql
  DECLARE lcur_get_foreign_key_list CURSOR FOR lpre_get_foreign_key_list
  
  OPEN lcur_get_foreign_key_list
  FOREACH lcur_get_foreign_key_list INTO ls_column_name
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_count = li_rec_count + 1

    LET ls_column_list = ls_column_list,ls_column_name,","
    
  END FOREACH
  CLOSE lcur_get_foreign_key_list
  
  FREE lcur_get_foreign_key_list
  FREE lpre_get_foreign_key_list

  LET ls_column_list = ls_column_list.subString(1,ls_column_list.getLength()-1)
  
  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return,ls_column_list
  
END FUNCTION

#檢核表格 Index 值是否存在
FUNCTION sadzi140_db_check_index_data(p_table_name,p_key_name)
DEFINE
  p_table_name  STRING,
  p_key_name    STRING
DEFINE
  lb_return     BOOLEAN,
  ls_sql        STRING,
  li_rec_count  INTEGER,
  ls_table_name STRING,
  ls_key_name   STRING

  LET ls_table_name = p_table_name
  LET ls_key_name   = p_key_name

  #取得資料筆數
  LET ls_sql = "select count(1)                        ",
               "  from dzec_t ec                       ",
               " where 1=1                             ",
               "   and ec.dzec001 = '",ls_table_name,"'", 
               "   and ec.dzec002 = '",ls_key_name,"'  " 
 
  PREPARE lpre_check_index_data FROM ls_sql
  DECLARE lcur_check_index_data CURSOR FOR lpre_check_index_data
  OPEN lcur_check_index_data
  FETCH lcur_check_index_data INTO li_rec_count
  CLOSE lcur_check_index_data
  FREE lcur_check_index_data
  FREE lpre_check_index_data

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION

#檢核表格是否被鎖定
FUNCTION sadzi140_db_check_table_lock(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE  
  ls_sql        STRING,
  li_rec_count  INTEGER,
  ls_table_name STRING,
  ls_enterprise STRING,
  ls_db_type    STRING, #20240704
  ls_ent_db     STRING, #20191209
  lv_user_name  VARCHAR(500)
DEFINE
  lb_return     BOOLEAN,
  ls_user_name  STRING  

  LET ls_table_name = p_table_name

  &ifndef DEBUG
    LET ls_enterprise = g_enterprise
  &else
    LET ls_enterprise = cs_enterprise
  &endif  

  &ifdef is_msv
  LET ls_sql = "select count(1),dl.login_name+'-'+dl.host_name username   ", 
               "  from dba_locks dl                                       ",                      
               " where dl.object_type = 'TABLE'                           ",
               "   and dl.object_name = '",ls_table_name.toLowerCase(),"' ",
               " group by dl.login_name+'-'+dl.host_name                  "
  &else
  &ifdef is_pgs
  LET ls_sql = "select count(1),coalesce(dl.os_user_name,'N/A')||'-'||dl.machine_name username ", 
               "  from dba_locks dl                                                            ",
               " where dl.object_name = '",ls_table_name.toLowerCase(),"'                      ",
               " group by coalesce(dl.os_user_name,'N/A')||'-'||dl.machine_name                "
               
  LET ls_sql = "select 0 as cnts,'tiptop/tiptop' as username from dual" --For debug
  &else
  CALL sadzi140_db_get_ent_mapping_db() RETURNING ls_ent_db #20191209

  LET ls_db_type = sadzi140_util_get_db_type()
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb
      #取得資料筆數
      LET ls_sql = "SELECT COUNT(1),L.OS_USER_NAME||'-'||AG.OOAG011 USERNAME     ",                  
                   "  FROM GV$LOCKED_OBJECT L                                    ",
                   " INNER JOIN DBA_OBJECTS O ON O.OBJECT_ID = L.OBJECT_ID       ",
                   " INNER JOIN V$SESSION V ON V.SID = L.SESSION_ID              ",
                   " INNER JOIN V$PROCESS P ON P.ADDR= V.PADDR                   ",
                   "  LEFT OUTER JOIN (                                          ",   
                   "                    SELECT XA.GZXA001,AG.OOAG011             ",   
                   "                      FROM ",ls_ent_db,".OOAG_T AG,          ", #20191209   
                   "                           GZXA_T XA                         ",   
                   "                     WHERE XA.GZXAENT = AG.OOAGENT           ",   
                   "                        AND AG.OOAG001 = XA.GZXA003          ",    
                   "                        AND XA.GZXAENT = '",ls_enterprise,"' ",
                   "                      GROUP BY XA.GZXA001,AG.OOAG011         ",    
                   "                      ORDER BY XA.GZXA001                    ",    
                   "                  ) AG                                       ",
                   "                  ON AG.GZXA001 = L.OS_USER_NAME             ",
                   " WHERE 1=1                                                   ",
                   "   AND O.OBJECT_NAME = '",ls_table_name.toUpperCase(),"'     ",
                   "   AND V.LOCKWAIT IS NULL                                    ",
                   " GROUP BY L.OS_USER_NAME||'-'||AG.OOAG011                    ",
                   " ORDER BY 2                                                  "
      LET ls_sql = "select 0 as cnts,'tiptop/tiptop' as username from dual" --For debug
    WHEN cs_db_driver_type_dmg
      #取得資料筆數
      LET ls_sql = "SELECT COUNT(1),S.USER_NAME||'-'||COALESCE(AG.OOAG011,S.CLNT_IP) USERNAME ", 
                   "  FROM V$LOCK v                                                           ", 
                   "  INNER JOIN V$SESSIONS S                                                 ", 
                   "          ON S.TRX_ID = V.TRX_ID                                          ", 
                   "  INNER JOIN dba_objects o                                                ", 
                   "          ON o.object_id = v.table_id                                     ", 
                   "  LEFT OUTER JOIN (                                                       ", 
                   "                    SELECT XA.GZXA001,AG.OOAG011                          ", 
                   "                      FROM ",ls_ent_db,".OOAG_T AG,                       ", 
                   "                           GZXA_T XA                                      ", 
                   "                     WHERE XA.GZXAENT = AG.OOAGENT                        ", 
                   "                        AND AG.OOAG001 = XA.GZXA003                       ", 
                   "                        AND XA.GZXAENT = '",ls_enterprise,"'              ", 
                   "                      GROUP BY XA.GZXA001,AG.OOAG011                      ", 
                   "                      ORDER BY XA.GZXA001                                 ", 
                   "                  ) AG                                                    ", 
                   "                  ON AG.GZXA001 = S.USER_NAME                             ", 
                   " WHERE 1=1                                                                ", 
                   "   AND V.TRX_ID <> 0                                                      ", 
                   "   AND V.IGN_FLAG <> 1                                                    ", 
                   "   AND o.object_name = '",ls_table_name.toUpperCase(),"'                  ",
                   "   AND ROWNUM = 1                                                         ",
                   " GROUP BY S.USER_NAME||'-'||COALESCE(AG.OOAG011,S.CLNT_IP)                "
      LET ls_sql = "select 0 as cnts,'tiptop/tiptop' as username from dual" --For debug
  OTHERWISE  
    #Oracle
    #取得資料筆數
    LET ls_sql = "SELECT COUNT(1),L.OS_USER_NAME||'-'||AG.OOAG011 USERNAME     ",                  
                 "  FROM GV$LOCKED_OBJECT L                                    ",
                 " INNER JOIN DBA_OBJECTS O ON O.OBJECT_ID = L.OBJECT_ID       ",
                 " INNER JOIN V$SESSION V ON V.SID = L.SESSION_ID              ",
                 " INNER JOIN V$PROCESS P ON P.ADDR= V.PADDR                   ",
                 "  LEFT OUTER JOIN (                                          ",   
                 "                    SELECT XA.GZXA001,AG.OOAG011             ",   
                 "                      FROM ",ls_ent_db,".OOAG_T AG,          ", #20191209   
                 "                           GZXA_T XA                         ",   
                 "                     WHERE XA.GZXAENT = AG.OOAGENT           ",   
                 "                        AND AG.OOAG001 = XA.GZXA003          ",    
                 "                        AND XA.GZXAENT = '",ls_enterprise,"' ",
                 "                      GROUP BY XA.GZXA001,AG.OOAG011         ",    
                 "                      ORDER BY XA.GZXA001                    ",    
                 "                  ) AG                                       ",
                 "                  ON AG.GZXA001 = L.OS_USER_NAME             ",
                 " WHERE 1=1                                                   ",
                 "   AND O.OBJECT_NAME = '",ls_table_name.toUpperCase(),"'     ",
                 "   AND V.LOCKWAIT IS NULL                                    ",
                 " GROUP BY L.OS_USER_NAME||'-'||AG.OOAG011                    ",
                 " ORDER BY 2                                                  "
  END CASE 
  &endif
  &endif

  PREPARE lpre_check_table_lock FROM ls_sql
  DECLARE lcur_check_table_lock CURSOR FOR lpre_check_table_lock
  OPEN lcur_check_table_lock
  FETCH lcur_check_table_lock INTO li_rec_count,lv_user_name
  CLOSE lcur_check_table_lock
  FREE lcur_check_table_lock
  FREE lpre_check_table_lock

  LET ls_user_name = lv_user_name
  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return,ls_user_name
  
END FUNCTION

FUNCTION sadzi140_db_get_more_one_record_from_all_tables(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_table_name      STRING,
  li_rec_cnt         INTEGER,
  ls_sql             STRING,
  lv_counts_sql      VARCHAR(1024),
  ls_counts_sql      STRING,
  ls_zone            STRING, #20191209
  ls_db_leading      STRING, #20191209
  lb_more_record     BOOLEAN,
  lb_more_one_record BOOLEAN
DEFINE  
  lb_return BOOLEAN

  LET ls_table_name = p_table_name

  LET li_rec_cnt = 1
  LET lb_more_one_record = FALSE

  LET ls_zone = FGL_GETENV("ZONE") #20191209
  LET ls_db_leading = ls_zone,"_" #20191209

  &ifdef is_msv
  LET ls_sql = "select 'select count(1) from '+atb.owner+'.dbo.'+atb.table_name as sqls ",                   
               "  from dba_tables atb                                                   ",                   
               " where atb.table_name = '",ls_table_name.toLowerCase(),"'               ",                   
               "   and exists (                                                         ",
               "                select 1                                                ",                   
               "                  from gzda_t da                                        ",                   
               "                 where da.gzda005  = 'Y'                                ",                   
               "                   and da.gzdastus = 'Y'                                ",                   
               "                   and '",ls_db_leading,"'+da.gzda001 = atb.owner       ",
               "              )                                                         "
  &else
  &ifdef is_pgs
  LET ls_sql = "select 'select count(1) from '||atb.owner||'.'||atb.table_name as sqls ",                   
               "  from dba_tables atb                                                  ",                   
               " where atb.table_name = '",ls_table_name.toLowerCase(),"'              ",                   
               "   and exists (                                                        ",
               "                select 1                                               ",                   
               "                  from gzda_t da                                       ",                   
               "                 where da.gzda005  = 'Y'                               ",                   
               "                   and da.gzdastus = 'Y'                               ",                   
               "                   and da.gzda001  = atb.owner                         ",
               "              )                                                        "
  &else
  LET ls_sql = "select 'select count(1) from '||atb.owner||'.'||atb.table_name as sqls ",                   
               "  from dba_tables atb                                                  ",                   
               " where atb.table_name = '",ls_table_name.toUpperCase(),"'              ",                   
               "   and exists (                                                        ", -- 20180514 begin 
               "                select 1                                               ",                   
               "                  from gzda_t da                                       ",                   
               "                 where da.gzda005  = 'Y'                               ",                   
               "                   and da.gzdastus = 'Y'                               ",                   
               "                   and da.gzda001  = lower(atb.owner)                  ", -- 20180514 end   
               "              )                                                        "
  &endif               
  &endif               
               
  PREPARE lpre_get_more_one_record_from_all_tables FROM ls_sql
  DECLARE lcur_get_more_one_record_from_all_tables CURSOR FOR lpre_get_more_one_record_from_all_tables

  OPEN lcur_get_more_one_record_from_all_tables
  FOREACH lcur_get_more_one_record_from_all_tables INTO lv_counts_sql
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET ls_counts_sql = lv_counts_sql
    CALL sadzi140_db_get_table_more_one_record(ls_counts_sql) RETURNING lb_more_record
    IF lb_more_record THEN
      LET lb_more_one_record = TRUE
    END IF
    
    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_more_one_record_from_all_tables
  
  FREE lpre_get_more_one_record_from_all_tables
  FREE lcur_get_more_one_record_from_all_tables  

  LET lb_return = lb_more_one_record

  RETURN lb_return
  
END FUNCTION

#抓取表格是否有資料存在
FUNCTION sadzi140_db_get_table_more_one_record(p_sql)
DEFINE
  p_sql    STRING
DEFINE
  ls_sql        STRING,
  li_rec_count  INTEGER
DEFINE
  lb_return     BOOLEAN

  LET ls_sql = p_sql
  
  LET lb_return = FALSE

  PREPARE lpre_get_table_more_one_record FROM ls_sql
  DECLARE lcur_get_table_more_one_record CURSOR FOR lpre_get_table_more_one_record
  TRY 
    OPEN lcur_get_table_more_one_record
    FETCH lcur_get_table_more_one_record INTO li_rec_count
    CLOSE lcur_get_table_more_one_record
  CATCH
    LET li_rec_count = 0
  END TRY   
  FREE lcur_get_table_more_one_record
  FREE lpre_get_table_more_one_record

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_db_update_alter_code(p_owner,p_table_name)
DEFINE
  p_owner      STRING,
  p_table_name STRING
DEFINE
  ls_owner         STRING,
  ls_table_name    STRING,
  li_col_fwd_rec   INTEGER,
  li_col_bwd_rec   INTEGER,
  li_key_fwd_rec   INTEGER,
  li_key_bwd_rec   INTEGER,
  li_idx_fwd_rec   INTEGER,
  li_idx_bwd_rec   INTEGER,
  li_columnid_rec  INTEGER,
  ls_exec_sql      STRING,
  ls_alter_code    STRING

  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name

  #先重置異動碼
  CALL sadzi140_db_update_default_alter_code(ls_table_name,"Y")

  CALL sadzi140_db_get_column_forward_records(ls_owner,ls_table_name) RETURNING li_col_fwd_rec
  CALL sadzi140_db_get_column_backward_records(ls_owner,ls_table_name) RETURNING li_col_bwd_rec
  CALL sadzi140_db_get_key_forward_records(ls_owner,ls_table_name) RETURNING li_key_fwd_rec
  CALL sadzi140_db_get_key_backward_records(ls_owner,ls_table_name) RETURNING li_key_bwd_rec
  CALL sadzi140_db_get_index_forward_records(ls_owner,ls_table_name) RETURNING li_idx_fwd_rec
  CALL sadzi140_db_get_index_backward_records(ls_owner,ls_table_name) RETURNING li_idx_bwd_rec

  CALL sadzi140_db_get_column_id_status(ls_owner,ls_table_name) RETURNING li_columnid_rec
  
  ##############################################################################
  IF (li_col_fwd_rec >= 1) OR (li_col_bwd_rec >= 1) OR
     (li_key_fwd_rec >= 1) OR (li_key_bwd_rec >= 1) OR 
     (li_idx_fwd_rec >= 1) OR (li_idx_bwd_rec >= 1) THEN #20170706
    LET ls_alter_code = "N"
  ELSE
    #LET ls_alter_code = "Y" #20170706 mark 
    #20170706 begin
    IF (li_columnid_rec >= 1) THEN
      LET ls_alter_code = "I"
    ELSE
      LET ls_alter_code = "Y"
    END IF
    #20170706 end
  END IF

  LET ls_exec_sql = " update dzea_t                                      ",
                    "    set dzea011 = '",ls_alter_code,"'               ",
                    "  where dzea001 = '",ls_table_name.toLowerCase(),"' "
                    
  CALL sadzi140_db_exec_SQL(ls_exec_sql,FALSE)                      
  
  ##############################################################################
  
END FUNCTION 

#20170523 begin
FUNCTION sadzi140_db_set_alter_code(p_owner,p_table_name,p_alter_code)
DEFINE
  p_owner      STRING,
  p_table_name STRING,
  p_alter_code STRING 
DEFINE
  ls_owner         STRING,
  ls_table_name    STRING,
  ls_alter_code    STRING,
  ls_exec_sql      STRING

  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name
  LET ls_alter_code = p_alter_code

  #先重置異動碼
  CALL sadzi140_db_update_default_alter_code(ls_table_name,"Y")

  LET ls_exec_sql = " update dzea_t                                      ",
                    "    set dzea011 = '",ls_alter_code,"'               ",
                    "  where dzea001 = '",ls_table_name.toLowerCase(),"' "
                    
  CALL sadzi140_db_exec_SQL(ls_exec_sql,FALSE)                      
  
  ##############################################################################
  
END FUNCTION 

FUNCTION sadzi140_db_check_table_if_altered(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_count  INTEGER,
  lb_return     BOOLEAN

  LET ls_table_name = p_table_name.toLowerCase()

  #取得資料筆數
  LET ls_sql = "select count(1) cnts                    ",
               "  from dzea_t ea                        ",
               " where ea.dzea001 = '",ls_table_name,"' ", 
               "   and ea.dzea011 = 'Y'                 " 
                
  PREPARE lpre_check_table_if_altered FROM ls_sql
  DECLARE lcur_check_table_if_altered CURSOR FOR lpre_check_table_if_altered
  OPEN lcur_check_table_if_altered
  FETCH lcur_check_table_if_altered INTO li_rec_count
  CLOSE lcur_check_table_if_altered
  FREE lcur_check_table_if_altered
  FREE lpre_check_table_if_altered

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION
#20170523 end

FUNCTION sadzi140_db_update_shipping_code(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_table_name    STRING,
  ls_exec_sql      STRING

  LET ls_table_name = p_table_name

  BEGIN WORK 
  
  LET ls_exec_sql = " update dzea_t                                      ",
                    "    set dzea017 = 'Y'                               ",
                    "  where dzea001 = '",ls_table_name.toLowerCase(),"' "
                    
  CALL sadzi140_db_exec_SQL_no_commit(ls_exec_sql,TRUE)

  LET ls_exec_sql = " update dzeb_t                                      ",
                    "    set dzeb031 = 'Y'                               ",
                    "  where dzeb001 = '",ls_table_name.toLowerCase(),"' "
                    
  CALL sadzi140_db_exec_SQL_no_commit(ls_exec_sql,TRUE)
  
  LET ls_exec_sql = " update dzec_t                                      ",
                    "    set dzec008 = 'Y'                               ",
                    "  where dzec001 = '",ls_table_name.toLowerCase(),"' "
                    
  CALL sadzi140_db_exec_SQL_no_commit(ls_exec_sql,TRUE)
  
  LET ls_exec_sql = " update dzed_t                                      ",
                    "    set dzed010 = 'Y'                               ",
                    "  where dzed001 = '",ls_table_name.toLowerCase(),"' "
                    
  CALL sadzi140_db_exec_SQL_no_commit(ls_exec_sql,TRUE)
  
  COMMIT WORK   
  
END FUNCTION 

FUNCTION sadzi140_db_update_default_alter_code(p_table_name,p_alter_code)
DEFINE
  p_table_name STRING,
  p_alter_code STRING
DEFINE
  ls_table_name STRING,
  ls_alter_code STRING,
  ls_exec_sql   STRING

  LET ls_table_name = p_table_name
  LET ls_alter_code = p_alter_code

  BEGIN WORK

  #Column
  LET ls_exec_sql = " update dzeb_t                                      ",
                    "    set dzeb028 = '",ls_alter_code,"'               ",
                    "  where dzeb001 = '",ls_table_name.toLowerCase(),"' "
                    
  CALL sadzi140_db_exec_SQL_no_commit(ls_exec_sql,FALSE)

  #Index
  LET ls_exec_sql = " update dzec_t                                      ",
                    "    set dzec005 = '",ls_alter_code,"'               ",
                    "  where dzec001 = '",ls_table_name.toLowerCase(),"' "
                    
  CALL sadzi140_db_exec_SQL_no_commit(ls_exec_sql,FALSE)

  #Key
  LET ls_exec_sql = " update dzed_t                                      ",
                    "    set dzed007 = '",ls_alter_code,"'               ",
                    "  where dzed001 = '",ls_table_name.toLowerCase(),"' "
                    
  CALL sadzi140_db_exec_SQL_no_commit(ls_exec_sql,FALSE)

  #Main Table
  LET ls_exec_sql = " update dzea_t                                      ",
                    "    set dzea011 = '",ls_alter_code,"'               ",
                    "  where dzea001 = '",ls_table_name.toLowerCase(),"' "
                    
  CALL sadzi140_db_exec_SQL_no_commit(ls_exec_sql,FALSE)                      

  COMMIT WORK
  
END FUNCTION 

FUNCTION sadzi140_db_get_column_forward_records(p_owner,p_table_name)
DEFINE
  p_owner      STRING,
  p_table_name STRING
DEFINE
  lo_column_forward_records T_COLUMN_DATA
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_column_name STRING,
  ls_sql         STRING,
  ls_exec_sql    STRING,
  li_rec_cnt     INTEGER,
  ls_db_type     STRING, #20240606 
  ls_ora_coldef  STRING, #20240529
  ls_alter_table VARCHAR(30)
DEFINE
  li_return INTEGER  

  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name
  
  LET li_rec_cnt = 0

  ------------------------------------------------------------------------------------------------
  -- Column forward
  ------------------------------------------------------------------------------------------------
  &ifdef is_msv
  LET ls_sql = "select eb.dzeb002                       column_name,      ", 
               "       eb.dzeb004                       iskey,            ", 
               "       case eb.dzeb004                                    ",    
               "         when 'Y' then 'N'                                ",    
               "         when 'N' then 'Y'                                ",       
               "       else                                               ", 
               "         'Y'                                              ",    
               "       end                              nullable,         ",    
               "       td.gztd003                       data_type,        ", 
               "      case td.gztd003                                     ", 
               "         when 'varbinary'                                 ",    
               "           then 'max'                                     ",       
               "         when 'nvarchar'                                  ",       
               "           then iif(td.gztd009='TEXT','max',td.gztd008)   ",       
               "       else                                               ", 
               "         td.gztd008                                       ",    
               "       end                              data_length,      ",    
               "       trim(eb.dzeb012)                 data_default      ",
               "  from dzeb_t eb                                          ",
               "  left outer join gztd_t td                               ",
               "               on td.gztd001 = eb.dzeb006                 ",
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'     ",
               "except                                                    ", 
               "select atc.column_name                 column_name,       ",                       
               "       iif(atc.nullable='N','Y','N')   iskey,             ",                       
               "       atc.nullable                    nullable,          ",                       
               "       atc.data_type                   data_type,         ",
               "       atc.data_length                 data_length,       ",                       
               "       atc.data_default                data_default       ",   
               "  from dba_tab_columns atc                                ",                            
               " where 1=1                                                ",                            
               "   and atc.owner      = '",ls_owner.toLowerCase(),"'      ",                            
               "   and atc.table_name = '",ls_table_name.toLowerCase(),"' ",                            
               " order by 1                                               "                            
  &else
  &ifdef is_pgs
  LET ls_sql = "select eb.dzeb002                       column_name,      ",    
               "       eb.dzeb004                       iskey,            ",    
               "       case eb.dzeb004                                    ",    
               "         when 'Y' then 'N'                                ",    
               "         when 'N' then 'Y'                                ",    
               "       else                                               ",    
               "         'Y'                                              ",    
               "       end                              nullable,         ",    
               "       td.gztd003                       data_type,        ",    
               "       td.gztd008                       data_length,      ",    
               "       trim(eb.dzeb012)                 data_default      ",    
               "  from dzeb_t eb                                          ",    
               "  left outer join gztd_t td                                                 ",  
               "               on td.gztd001 = eb.dzeb006                                   ",  
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'                       ",    
               "except                                                                      ",    
               "select atc.column_name                 column_name,                         ",    
               "       case atc.nullable                                                    ",  
               "         when 'N' then 'Y'                                                  ",  
               "       else 'N' end                    iskey,                               ",    
               "       atc.nullable                    nullable,                            ",    
               "       atc.data_type                   data_type,                           ",    
               "       case                                                                 ",
               "         when atc.data_type = 'text' then null                              ",
               "         when atc.data_type = 'bytea' then null                             ",
               "         when atc.data_type = 'date' then null                              ",
               "         when atc.data_type = 'int2' then null                              ",
               "         when atc.data_type = 'int4' then null                              ",
               "         when atc.data_type = 'int8' then null                              ",
               "         when atc.data_type = 'decimal' then                                ",
               "           atc.data_precision::varchar||','||                               ",
               "           atc.data_scale::varchar                                          ",
               "       else                                                                 ",
               "         atc.data_length::varchar                                           ",
               "       end::varchar                    data_length,                         ",
               "       case                                                                 ",
               "         when atc.data_default is null then null                            ",
               "         when atc.data_default = 'NULL' then null                           ",
               "         when atc.data_default = 'null' then null                           ",
               "         when atc.data_default like 'CURRENT%' then lower(atc.data_default) ",
               "         when atc.data_default = '\"sysdate\"()' then 'sysdate'             ",
               "         when atc.data_default = 'SYSDATE' then 'sysdate'                   ",
               "       else                                                                 ",
               "         atc.data_default                                                   ",
               "       end as data_default                                                  ",
               "  from dba_tab_columns atc                                                  ",    
               " where 1=1                                                                  ",    
               "   and atc.owner      = '",ls_owner.toLowerCase(),"'                        ",    
               "   and atc.table_name = '",ls_table_name.toLowerCase(),"'                   ",    
               " order by 1                                                                 "  
  &else
  LET ls_db_type = sadzi140_util_get_db_type()
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb
      LET ls_ora_coldef = " case                                                                 ",
                          "   when atc.data_default is null then null                            ",
                          "   when atc.data_default = 'NULL' then null                           ",
                          "   when atc.data_default = 'null' then null                           ",
                          "   when atc.data_default like 'CURRENT%' then lower(atc.data_default) ",
                          "   when atc.data_default = '\"sysdate\"()' then 'sysdate'             ",
                          "   when atc.data_default = 'SYSDATE' then 'sysdate'                   ",
                          " else                                                                 ",
                          "   atc.data_default                                                   ",
                          " end as data_default                                                  "
    WHEN cs_db_driver_type_dmg 
      LET ls_ora_coldef = " case                                                                 ",
                          "   when atc.data_default is null then null                            ",
                          "   when atc.data_default = 'NULL' then null                           ",
                          "   when atc.data_default = 'null' then null                           ",
                          "   when atc.data_default LIKE 'CURRENT%' then lower(atc.data_default) ",
                          "   when atc.data_default = '\"sysdate\"()' then 'sysdate'             ",
                          "   when atc.data_default = 'SYSDATE' then 'sysdate'                   ",
                          " else                                                                 ",
                          "   atc.data_default                                                   ",
                          " end as data_default                                                  "
  OTHERWISE  
    LET ls_ora_coldef = " decode(trim(ds.get_col_default(atc.table_name,atc.column_name)),'',null,       ",
                        "        trim(ds.get_col_default(atc.table_name,atc.column_name))) data_default  "
  END CASE
  
  LET ls_sql = "select upper(eb.dzeb002)                      column_name,                              ",  
               "       eb.dzeb004                             iskey,                                    ",  
               "       decode(eb.dzeb004,'Y','N','N','Y','Y') nullable,                                 ",  
               "       upper(td.gztd003)                      data_type,                                ",  
               "       replace(to_char(to_number(replace(replace(                                       ",  
               "       decode(                                                                          ",  
               "              upper(td.gztd003),                                                        ",  
               "              'BLOB','4000',                                                            ",  
               "              'CLOB','4000',                                                            ",  
               "              'DATE','7',                                                               ",  
               "              'TIMESTAMP',decode(td.gztd008,'0','7','11'),                              ",  
               "              td.gztd008                                                                ",  
               "             ),',','.'),' ',''))),'.',',')    data_length,                              ",  
               "       trim(decode(eb.dzeb012,'',null,eb.dzeb012))    data_default                      ",  
               "  from dzeb_t eb                                                                        ",  
               "  left outer join gztd_t td                                                             ", 
               "               on td.gztd001 = eb.dzeb006                                               ", 
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'                                   ",  
               "minus                                                                                   ",  
               "select atc.column_name                      column_name,                                ",  
               "       decode(acc.column_name,null,'N','Y') iskey,                                      ",  
               "       atc.nullable                         nullable,                                   ",  
               "       decode(instrb(atc.data_type,'TIMESTAMP'),1,'TIMESTAMP',atc.data_type) data_type, ",  
               "       replace(to_char(decode(                                                          ",  
               "               atc.data_type,                                                           ",  
               "               'BLOB',4000,                                                             ", #20240529
               "               'CLOB',4000,                                                             ", #20240529
               "               'INT2',NULL,                                                             ", #20240529
               "               'INT4',NULL,                                                             ", #20240529
               "               'INT8',NULL,                                                             ", #20240529
               "               'NUMBER',decode(                                                         ",  
               "                                nvl(atc.data_scale,'0'),                                ",  
               "                                '0',atc.data_precision,                                 ",  
               "                                atc.data_precision||'.'||atc.data_scale                 ",  
               "                              ),                                                        ",  
               "               'NUMERIC',DECODE(                                                        ", #20240529 begin
               "                                NVL(ATC.DATA_SCALE,'0'),                                ",
               "                                '0',ATC.DATA_PRECISION,                                 ",
               "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE                 ", 
               "                              ),                                                        ", #20240529 end
               "               'DATE',7,                                                                ", #20240529
               "               'TIMESTAMP(0)',7,                                                        ", #20240529
               "               'TIMESTAMP(5)',11,                                                       ", #20240529
               "               'TIMESTAMP(6)',11,                                                       ", #20240529
               "               'TIMESTAMP',decode(                                                      ",
               "                                   atc.data_scale,0,7,8,7,11                            ", #20240529
               "                                 ),                                                     ",                
               "               'DATE',decode(                                                           ",
               "                              atc.data_length,3,7,atc.data_length                       ",
               "                            ),                                                          ",                
               "               atc.data_length                                                          ",  
               "             )),'.',',')                    data_length,                                ",
               ls_ora_coldef,  
               "  from dba_tab_columns atc                                                              ",  
               "  left outer join dba_cons_columns acc                                                  ", 
               "               on acc.owner       = atc.owner                                           ", 
               "              and acc.table_name  = atc.table_name                                      ", 
               "              and acc.column_name = atc.column_name                                     ", 
               "              and acc.constraint_name like '%PK'                                        ", 
               " where 1=1                                                                              ",  
               "   and atc.owner      = '",ls_owner.toUpperCase(),"'                                    ",  
               "   and atc.table_name = '",ls_table_name.toUpperCase(),"'                               ", 
               "   and atc.column_name not like 'SYS_%'                                                 ", 
               " order by 1                                                                             "
  &endif
  &endif
  
  PREPARE lpre_get_column_forward_records FROM ls_sql
  DECLARE lcur_get_column_forward_records CURSOR FOR lpre_get_column_forward_records

  BEGIN WORK
  
  OPEN lcur_get_column_forward_records
  FOREACH lcur_get_column_forward_records INTO lo_column_forward_records.*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET ls_column_name = lo_column_forward_records.COLUMN_NAME
    
    LET ls_exec_sql = "update dzeb_t                                       ",
                      "   set dzeb028 = 'N'                                ",
                      " where dzeb001 = '",ls_table_name.toLowerCase(),"'  ",
                      "   and dzeb002 = '",ls_column_name.toLowerCase(),"' "
         
    CALL sadzi140_db_exec_SQL_no_commit(ls_exec_sql,FALSE)                      

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_column_forward_records

  COMMIT WORK
  
  FREE lpre_get_column_forward_records
  FREE lcur_get_column_forward_records  
  -------------------------------------------------------------------------------------------------
  
  LET li_return = li_rec_cnt
  
  RETURN li_return
  
END FUNCTION 

FUNCTION sadzi140_db_get_column_backward_records(p_owner,p_table_name)
DEFINE
  p_owner      STRING,
  p_table_name STRING
DEFINE
  lo_column_backward_records T_COLUMN_DATA
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_column_name STRING,
  ls_sql         STRING,
  ls_exec_sql    STRING,
  ls_ora_coldef  STRING, #20240529
  ls_db_type     STRING, #20240606
  li_rec_cnt     INTEGER,
  ls_alter_table VARCHAR(30)
DEFINE
  li_return INTEGER  

  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name
  
  LET li_rec_cnt = 0

  ------------------------------------------------------------------------------------------------
  -- Column backward
  ------------------------------------------------------------------------------------------------
  &ifdef is_msv
  LET ls_sql = "select atc.column_name                 column_name,       ",         
               "       iif(atc.nullable='N','Y','N')   iskey,             ",         
               "       atc.nullable                    nullable,          ",         
               "       atc.data_type                   data_type,         ",
               "       atc.data_length                 data_length,       ",         
               "       atc.data_default                data_default       ",
               "  from dba_tab_columns atc                                ",              
               " where 1=1                                                ",              
               "   and atc.owner      = '",ls_owner.toLowerCase(),"'      ",              
               "   and atc.table_name = '",ls_table_name.toLowerCase(),"' ",              
               "except                                                    ",
               "select eb.dzeb002                       column_name,      ",
               "       eb.dzeb004                       iskey,            ",
               "       case eb.dzeb004                                    ",
               "         when 'Y' then 'N'                                ",
               "         when 'N' then 'Y'                                ",
               "       else                                               ",
               "         'Y'                                              ",
               "       end                              nullable,         ",
               "       td.gztd003                       data_type,        ",
               "      case td.gztd003                                     ",
               "         when 'varbinary'                                 ",
               "           then 'max'                                     ",
               "         when 'nvarchar'                                  ",
               "           then iif(td.gztd009='TEXT','max',td.gztd008)   ",
               "       else                                               ",
               "         td.gztd008                                       ",
               "       end                              data_length,      ",
               "       trim(eb.dzeb012)                 data_default      ",
               "  from dzeb_t eb                                          ",
               "  left outer join gztd_t td                               ",
               "               on td.gztd001 = eb.dzeb006                 ",
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'     ",
               " order by 1                                               "              
  &else
  &ifdef is_pgs
  LET ls_sql = "select atc.column_name                 column_name,                         ",
               "       case atc.nullable                                                    ",
               "         when 'N' then 'Y'                                                  ",
               "       else 'N' end                    iskey,                               ",
               "       atc.nullable                    nullable,                            ",
               "       atc.data_type                   data_type,                           ",
               "       case                                                                 ",
               "         when atc.data_type = 'text' then null                              ",
               "         when atc.data_type = 'bytea' then null                             ",
               "         when atc.data_type = 'date' then null                              ",
               "         when atc.data_type = 'int2' then null                              ",
               "         when atc.data_type = 'int4' then null                              ",
               "         when atc.data_type = 'int8' then null                              ",
               "         when atc.data_type = 'decimal' then                                ",
               "           atc.data_precision::varchar||','||                               ",
               "           atc.data_scale::varchar                                          ",
               "       else                                                                 ",
               "         atc.data_length::varchar                                           ",
               "       end::varchar                    data_length,                         ",
               "       case                                                                 ",
               "         when atc.data_default is null then null                            ",
               "         when atc.data_default = 'NULL' then null                           ",
               "         when atc.data_default = 'null' then null                           ",
               "         when atc.data_default like 'CURRENT%' then lower(atc.data_default) ",
               "         when atc.data_default = '\"sysdate\"()' then 'sysdate'             ",
               "         when atc.data_default = 'SYSDATE' then 'sysdate'                   ",
               "       else                                                                 ",
               "         atc.data_default                                                   ",
               "       end as data_default                                                  ",
               "  from dba_tab_columns atc                                                  ",
               " where 1=1                                                                  ",
               "   and atc.owner      = '",ls_owner.toLowerCase(),"'                        ",
               "   and atc.table_name = '",ls_table_name.toLowerCase(),"'                   ",
               "except                                                                      ",
               "select eb.dzeb002                       column_name,                        ",
               "       eb.dzeb004                       iskey,                              ",
               "       case eb.dzeb004                                                      ",
               "         when 'Y' then 'N'                                                  ",
               "         when 'N' then 'Y'                                                  ",
               "       else 'Y' end                     nullable,                           ",
               "       td.gztd003                       data_type,                          ",
               "       td.gztd008                       data_length,                        ",
               "       trim(eb.dzeb012)                 data_default                        ",
               "  from dzeb_t eb                                                            ",
               "  left outer join gztd_t td                                                 ",
               "               on td.gztd001 = eb.dzeb006                                   ",
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'                       ",
               " order by 1                                                                 "
  &else
  LET ls_db_type = sadzi140_util_get_db_type()
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb
      LET ls_ora_coldef = " case                                                                 ",
                          "   when atc.data_default is null then null                            ",
                          "   when atc.data_default = 'NULL' then null                           ",
                          "   when atc.data_default = 'null' then null                           ",
                          "   when atc.data_default like 'CURRENT%' then lower(atc.data_default) ",
                          "   when atc.data_default = '\"sysdate\"()' then 'sysdate'             ",
                          "   when atc.data_default = 'SYSDATE' then 'sysdate'                   ",
                          " else                                                                 ",
                          "   atc.data_default                                                   ",
                          " end as data_default                                                  "
    WHEN cs_db_driver_type_dmg 
      LET ls_ora_coldef = " case                                                                 ",
                          "   when atc.data_default is null then null                            ",
                          "   when atc.data_default = 'NULL' then null                           ",
                          "   when atc.data_default = 'null' then null                           ",
                          "   when atc.data_default LIKE 'CURRENT%' then lower(atc.data_default) ",
                          "   when atc.data_default = '\"sysdate\"()' then 'sysdate'             ",
                          "   when atc.data_default = 'SYSDATE' then 'sysdate'                   ",
                          " else                                                                 ",
                          "   atc.data_default                                                   ",
                          " end as data_default                                                  "
  OTHERWISE  
    LET ls_ora_coldef = " decode(trim(ds.get_col_default(atc.table_name,atc.column_name)),'',null,       ",
                        "        trim(ds.get_col_default(atc.table_name,atc.column_name))) data_default  "
  END CASE
  
  LET ls_sql = "select atc.column_name                      column_name,                              ",          
               "       decode(acc.column_name,null,'N','Y') iskey,                                    ",          
               "       atc.nullable                         nullable,                                 ",          
               "       decode(instrb(atc.data_type,'TIMESTAMP'),1,'TIMESTAMP',                        ", 
               "              atc.data_type)  data_type,                                              ", 
               "       replace(to_char(decode(                                                        ",          
               "               atc.data_type,                                                         ",          
               "               'BLOB',4000,                                                           ", #20240529
               "               'CLOB',4000,                                                           ", #20240529
               "               'INT2',NULL,                                                           ", #20240529
               "               'INT4',NULL,                                                           ", #20240529
               "               'INT8',NULL,                                                           ", #20240529
               "               'NUMBER',decode(                                                       ",          
               "                                nvl(atc.data_scale,'0'),                              ",          
               "                                '0',atc.data_precision,                               ",          
               "                                atc.data_precision||'.'||atc.data_scale               ",          
               "                              ),                                                      ",          
               "               'NUMERIC',DECODE(                                                      ", #20240529 begin
               "                                NVL(ATC.DATA_SCALE,'0'),                              ",
               "                                '0',ATC.DATA_PRECISION,                               ",
               "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE               ", 
               "                              ),                                                      ", #20240529 end
               "               'DATE',7,                                                              ", #20240529
               "               'TIMESTAMP(0)',7,                                                      ", #20240529
               "               'TIMESTAMP(5)',11,                                                     ", #20240529
               "               'TIMESTAMP(6)',11,                                                     ", #20240529
               "               'TIMESTAMP',decode(                                                    ",
               "                                   atc.data_scale,0,7,8,7,11                          ", #20240529
               "                                 ),                                                   ",                
               "               'DATE',decode(                                                         ",
               "                              atc.data_length,3,7,atc.data_length                     ",
               "                            ),                                                        ",                
               "               atc.data_length                                                        ",          
               "             )),'.',',')  data_length,                                                ",
               ls_ora_coldef, 
               "  from dba_tab_columns atc                                                            ",          
               "  left outer join dba_cons_columns acc                                                ", 
               "               on acc.owner       = atc.owner                                         ", 
               "              and acc.table_name  = atc.table_name                                    ", 
               "              and acc.column_name = atc.column_name                                   ", 
               "              and acc.constraint_name like '%PK'                                      ", 
               " where 1=1                                                                            ",          
               "   and atc.owner      = '",ls_owner.toUpperCase(),"'                                  ",          
               "   and atc.table_name = '",ls_table_name.toUpperCase(),"'                             ",          
               "   and atc.column_name not like 'SYS_%'                                               ",          
               "minus                                                                                 ",          
               "select upper(eb.dzeb002)                      column_name,                            ",          
               "       eb.dzeb004                             iskey,                                  ",          
               "       decode(eb.dzeb004,'Y','N','N','Y','Y') nullable,                               ",          
               "       upper(td.gztd003)                      data_type,                              ",          
               "       replace(to_char(to_number(replace(replace(                                     ",          
               "       decode(                                                                        ",          
               "              upper(td.gztd003),                                                      ",          
               "              'BLOB','4000',                                                          ",          
               "              'CLOB','4000',                                                          ",          
               "              'DATE','7',                                                             ",          
               "              'TIMESTAMP',decode(td.gztd008,'0','7','11'),                            ",          
               "              td.gztd008                                                              ",          
               "             ),',','.'),' ',''))),'.',',')    data_length,                            ",          
               "       trim(decode(eb.dzeb012,'',null,eb.dzeb012))   data_default                     ",          
               "  from dzeb_t eb                                                                      ",          
               "  left outer join gztd_t td                                                           ", 
               "               on td.gztd001 = eb.dzeb006                                             ", 
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'                                 ",          
               " order by 1                                                                           "          
  &endif
  &endif
  
  PREPARE lpre_get_column_backward_records FROM ls_sql
  DECLARE lcur_get_column_backward_records CURSOR FOR lpre_get_column_backward_records

  OPEN lcur_get_column_backward_records
  FOREACH lcur_get_column_backward_records INTO lo_column_backward_records.*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_column_backward_records
  
  FREE lpre_get_column_backward_records
  FREE lcur_get_column_backward_records  
  -------------------------------------------------------------------------------------------------
  
  LET li_return = li_rec_cnt
  
  RETURN li_return
  
END FUNCTION 

FUNCTION sadzi140_db_get_key_forward_records(p_owner,p_table_name)
DEFINE
  p_owner      STRING,
  p_table_name STRING
DEFINE
  lo_key_forward_records T_CONSTRAINT_DIFF
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_key_name    STRING,
  ls_sql         STRING,
  ls_exec_sql    STRING,
  li_rec_cnt     INTEGER,
  ls_alter_table VARCHAR(30)
DEFINE
  li_return INTEGER  

  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name
  
  LET li_rec_cnt = 0

  ------------------------------------------------------------------------------------------------
  -- key forward
  ------------------------------------------------------------------------------------------------
  &ifdef is_msv
  LET ls_sql = "select ed.dzed001       table_name,                                           ",                                                                     
               "       ed.dzed002       constraint_name,                                      ",                                                                      
               "       ed.dzed003       constraint_type,                                      ",                                                                      
               "       ed.dzed004       column_names,                                         ",                                                                      
               "       ed.dzed005       r_table_name,                                         ",                                                                      
               "       ed.dzed006       r_column_names                                        ",                                                                      
               "  from dzed_t ed                                                              ",                                                                             
               " where ed.dzed001 = '",ls_table_name.toLowerCase(),"'                         ",                                                    
               "   and ed.dzed003 <> 'F'                                                      ",                                                                             
               "except                                                                        ",                                                                              
               "select acs.table_name,                                                        ",                                                                             
               "       acs.constraint_name,                                                   ",  
               "       acs.constraint_type,                                                   ",  
               "       acc.column_names,                                                      ",                                                                             
               "       NULL r_table_name,                                                     ",                                                                  
               "       NULL r_column_names                                                    ",                                                                               
               "  from dba_constraints acs                                                    ",                                                                             
               "  left outer join (                                                           ",                                                                             
               "                   select accs.constraint_name,accs.owner,accs.table_name,    ",  
               "                          string_agg(accs.column_name, ',')                   ",  
               "                          within group(order by accs.position) column_names   ",   
               "                     from dba_cons_columns accs                               ",                                                                             
               "                    group by accs.constraint_name,accs.owner,accs.table_name  ",                                                                             
               "                  ) acc on acc.constraint_name = acs.constraint_name          ",                                                                             
               "                       and acc.table_name      = acs.table_name               ",                                                                             
               "                       and acc.owner           = acs.owner                    ",                                                                             
               " where 1 = 1                                                                  ",                                                                             
               "   and acs.table_name = '",ls_table_name.toLowerCase(),"'                     ",                                                    
               "   and acs.owner      = '",ls_owner.toLowerCase(),"'                          "  
  &else
  &ifdef is_pgs
  LET ls_sql = "select ed.dzed001       table_name,                                          ",
               "       ed.dzed002       constraint_name,                                     ",
               "       ed.dzed003       constraint_type,                                     ",
               "       ed.dzed004       column_names,                                        ",
               "       ed.dzed005       r_table_name,                                        ",
               "       ed.dzed006       r_column_names                                       ",
               "  from dzed_t ed                                                             ",
               " where ed.dzed001 = '",ls_table_name.toLowerCase(),"'                        ",
               "   and ed.dzed003 <> 'F'                                                     ",
               "except                                                                       ",
               "select acs.table_name,                                                       ",
               "       acs.constraint_name,                                                  ",
               "       acs.constraint_type,                                                  ",
               "       acc.column_names,                                                     ",
               "       NULL r_table_name,                                                    ",
               "       NULL r_column_names                                                   ",
               "  from dba_constraints acs                                                   ",
               "  left outer join (                                                          ",
               "                   select accs.constraint_name,accs.owner,accs.table_name,   ",
               "                          string_agg(accs.column_name, ','                   ",
               "                            order by accs.position) column_names             ",
               "                     from dba_cons_columns accs                              ",
               "                    group by accs.constraint_name,accs.owner,accs.table_name ",
               "                  ) acc                                                      ",
               "               on acc.constraint_name = acs.constraint_name                  ",
               "              and acc.table_name      = acs.table_name                       ",                                                                             
               "              and acc.owner           = acs.owner                            ",
               " where 1 = 1                                                                 ",
               "   and acs.table_name = '",ls_table_name.toLowerCase(),"'                    ",
               "   and acs.owner      = '",ls_owner.toLowerCase(),"'                         "
  &else
  LET ls_sql = "select upper(ed.dzed001)       table_name,                                         ",                                                                         
               "       upper(ed.dzed002)       constraint_name,                                    ",                                                                         
               "       upper(ed.dzed003)       constraint_type,                                    ",                                                                         
               "       upper(ed.dzed004)       column_names,                                       ",                                                                         
               "       upper(trim(ed.dzed005)) r_table_name,                                       ",                                                                         
               "       upper(trim(ed.dzed006)) r_column_names                                      ",                                                                         
               "  from dzed_t ed                                                                   ",                                                                         
               " where ed.dzed001 = '",ls_table_name.toLowerCase(),"'                              ",                                                                         
               "   and ed.dzed003 <> 'F'                                                           ",                                                                         
               "minus                                                                              ",                                                                         
               "select acs.table_name,                                                             ",                                                                         
               "       acs.constraint_name,                                                        ",                                                                         
               "       decode(acs.constraint_type, 'R', 'R', acs.constraint_type) constraint_type, ",                                                                         
               "       acc.column_names,                                                           ",                                                                         
               "       accr.table_name r_table_name,                                               ",                                                                         
               "       accr.column_names                                                           ",                                                                         
               "  from dba_constraints acs                                                         ",                                                                         
               "  left outer join (                                                                ",                                                                         
               "                   select accs.constraint_name,accs.owner,accs.table_name,         ",
               "                          listagg(accs.column_name, ',')                           ",
               "                            within group(order by accs.position) column_names      ",
               "                     from dba_cons_columns accs                                    ",                                                                         
               "                    group by accs.constraint_name,accs.owner,accs.table_name       ",                                                                         
               "                  ) acc on acc.constraint_name = acs.constraint_name               ",                                                                         
               "                       and acc.table_name      = acs.table_name                    ",                                                                             
               "                       and acc.owner           = acs.owner                         ",                                                                         
               "  left outer join (                                                                ",                                                                         
               "                   select accs.constraint_name,accs.owner,accs.table_name,         ",
               "                          listagg(accs.column_name, ',')                           ",
               "                            within group(order by accs.position) column_names      ",
               "                     from dba_cons_columns accs                                    ",                                                                         
               "                    group by accs.constraint_name,accs.owner,accs.table_name       ",                                                                         
               "                  ) accr on accr.constraint_name = acs.r_constraint_name           ",                                                                         
               "                        and accr.owner           = acs.r_owner                     ",                                                                          
               " where 1 = 1                                                                       ",                                                                         
               "   and acs.table_name = '",ls_table_name.toUpperCase(),"'                          ",                                                                         
               "   and acs.owner      = '",ls_owner.toUpperCase(),"'                               ",                                                                         
               "   and acs.constraint_name not like 'SYS%'                                         "
  &endif 
  &endif 
               
  PREPARE lpre_get_key_forward_records FROM ls_sql
  DECLARE lcur_get_key_forward_records CURSOR FOR lpre_get_key_forward_records

  OPEN lcur_get_key_forward_records
  FOREACH lcur_get_key_forward_records INTO lo_key_forward_records.*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_key_forward_records

  FREE lpre_get_key_forward_records
  FREE lcur_get_key_forward_records  
  -------------------------------------------------------------------------------------------------
  
  LET li_return = li_rec_cnt
  
  RETURN li_return
  
END FUNCTION 

FUNCTION sadzi140_db_get_key_backward_records(p_owner,p_table_name)
DEFINE
  p_owner      STRING,
  p_table_name STRING
DEFINE
  lo_key_backward_records T_CONSTRAINT_DIFF
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_key_name    STRING,
  ls_sql         STRING,
  ls_exec_sql    STRING,
  li_rec_cnt     INTEGER,
  ls_alter_table VARCHAR(30)
DEFINE
  li_return INTEGER  

  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name
  
  LET li_rec_cnt = 0

  ------------------------------------------------------------------------------------------------
  -- key backward
  ------------------------------------------------------------------------------------------------
  &ifdef is_msv
  LET ls_sql = "select acs.table_name,                                                        ",                                                                             
               "       acs.constraint_name,                                                   ",  
               "       acs.constraint_type,                                                   ",  
               "       acc.column_names,                                                      ",                                                                             
               "       NULL r_table_name,                                                     ",                                                                  
               "       NULL r_column_names                                                    ",                                                                               
               "  from dba_constraints acs                                                    ",                                                                             
               "  left outer join (                                                           ",                                                                             
               "                   select accs.constraint_name,accs.owner,accs.table_name,    ",  
               "                          string_agg(accs.column_name, ',')                   ",  
               "                          within group(order by accs.position) column_names   ",   
               "                     from dba_cons_columns accs                               ",                                                                             
               "                    group by accs.constraint_name,accs.owner,accs.table_name  ",                                                                             
               "                  ) acc on acc.constraint_name = acs.constraint_name          ",                                                                             
               "                       and acc.table_name      = acs.table_name               ",                                                                             
               "                       and acc.owner           = acs.owner                    ",                                                                             
               " where 1 = 1                                                                  ",                                                                             
               "   and acs.table_name = '",ls_table_name.toLowerCase(),"'                     ",                                                    
               "   and acs.owner      = '",ls_owner.toLowerCase(),"'                          ",
               "except                                                                        ",                                                                              
               "select ed.dzed001       table_name,                                           ",                                                                     
               "       ed.dzed002       constraint_name,                                      ",                                                                      
               "       ed.dzed003       constraint_type,                                      ",                                                                      
               "       ed.dzed004       column_names,                                         ",                                                                      
               "       ed.dzed005       r_table_name,                                         ",                                                                      
               "       ed.dzed006       r_column_names                                        ",                                                                      
               "  from dzed_t ed                                                              ",                                                                             
               " where ed.dzed001 = '",ls_table_name.toLowerCase(),"'                         ",                                                    
               "   and ed.dzed003 <> 'F'                                                      "                                                                             
  &else               
  &ifdef is_pgs
  LET ls_sql = "select acs.table_name,                                                        ",                                                                             
               "       acs.constraint_name,                                                   ",  
               "       acs.constraint_type,                                                   ",  
               "       acc.column_names,                                                      ",                                                                             
               "       NULL r_table_name,                                                     ",                                                                  
               "       NULL r_column_names                                                    ",                                                                               
               "  from dba_constraints acs                                                    ",                                                                             
               "  left outer join (                                                           ",                                                                             
               "                   select accs.constraint_name,accs.owner,accs.table_name,    ",  
               "                          string_agg(accs.column_name, ','                    ", 
               "                                     order by accs.position) column_names     ",   
               "                     from dba_cons_columns accs                               ",                                                                             
               "                    group by accs.constraint_name,accs.owner,accs.table_name  ",                                                                             
               "                  ) acc on acc.constraint_name = acs.constraint_name          ",                                                                             
               "                       and acc.table_name      = acs.table_name               ",                                                                             
               "                       and acc.owner           = acs.owner                    ",                                                                             
               " where 1 = 1                                                                  ",                                                                             
               "   and acs.table_name = '",ls_table_name.toLowerCase(),"'                     ",                                                    
               "   and acs.owner      = '",ls_owner.toLowerCase(),"'                          ",
               "except                                                                        ",                                                                              
               "select ed.dzed001       table_name,                                           ",                                                                     
               "       ed.dzed002       constraint_name,                                      ",                                                                      
               "       ed.dzed003       constraint_type,                                      ",                                                                      
               "       ed.dzed004       column_names,                                         ",                                                                      
               "       ed.dzed005       r_table_name,                                         ",                                                                      
               "       ed.dzed006       r_column_names                                        ",                                                                      
               "  from dzed_t ed                                                              ",                                                                             
               " where ed.dzed001 = '",ls_table_name.toLowerCase(),"'                         ",                                                    
               "   and ed.dzed003 <> 'F'                                                      "                                                                             
  &else               
  LET ls_sql = "select acs.table_name,                                                             ",                                                                          
               "       acs.constraint_name,                                                        ",                                                                          
               "       decode(acs.constraint_type, 'R', 'R', acs.constraint_type) constraint_type, ",                                                                          
               "       acc.column_names,                                                           ",                                                                          
               "       accr.table_name r_table_name,                                               ",                                                                          
               "       accr.column_names                                                           ",                                                                          
               "  from dba_constraints acs                                                         ",                                                                          
               "  left outer join (                                                                ",                                                                          
               "                   select accs.constraint_name,accs.owner,accs.table_name,         ", 
               "                          listagg(accs.column_name, ',')                           ", 
               "                            within group(order by accs.position) column_names      ", 
               "                     from dba_cons_columns accs                                    ",                                                                          
               "                    group by accs.constraint_name,accs.owner,accs.table_name       ",                                                                          
               "                  ) acc on acc.constraint_name = acs.constraint_name               ",                                                                          
               "                       and acc.table_name      = acs.table_name                    ",                                                                          
               "                       and acc.owner           = acs.owner                         ",                                                                          
               "  left outer join (                                                                ",                                                                          
               "                   select accs.constraint_name,accs.owner,accs.table_name,         ", 
               "                          listagg(accs.column_name, ',')                           ", 
               "                            within group(order by accs.position) column_names      ", 
               "                     from dba_cons_columns accs                                    ",                                                                          
               "                    group by accs.constraint_name,accs.owner,accs.table_name       ",                                                                          
               "                  ) accr on accr.constraint_name = acs.r_constraint_name           ",                                                                          
               "                        and accr.owner           = acs.r_owner                     ",                                                                          
               " where 1 = 1                                                                       ",                                                                          
               "   and acs.table_name = '",ls_table_name.toUpperCase(),"'                          ",                                                                          
               "   and acs.owner      = '",ls_owner.toUpperCase(),"'                               ",                                                                          
               "   and acs.constraint_name not like 'SYS%'                                         ",                                                                          
               "minus                                                                              ",                                                                          
               "select upper(ed.dzed001)       table_name,                                         ",                                                                          
               "       upper(ed.dzed002)       constraint_name,                                    ",                                                                          
               "       upper(ed.dzed003)       constraint_type,                                    ",                                                                          
               "       upper(ed.dzed004)       column_names,                                       ",                                                                          
               "       upper(trim(ed.dzed005)) r_table_name,                                       ",                                                                         
               "       upper(trim(ed.dzed006)) r_column_names                                      ",                                                                         
               "  from dzed_t ed                                                                   ",                                                                         
               " where ed.dzed001 = '",ls_table_name.toLowerCase(),"'                              ",                                                                         
               "   and ed.dzed003 <> 'F'                                                           "                                                                         
  &endif               
  &endif               
                
  PREPARE lpre_get_key_backward_records FROM ls_sql
  DECLARE lcur_get_key_backward_records CURSOR FOR lpre_get_key_backward_records

  OPEN lcur_get_key_backward_records
  FOREACH lcur_get_key_backward_records INTO lo_key_backward_records.*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_key_backward_records
  
  FREE lpre_get_key_backward_records
  FREE lcur_get_key_backward_records  
  -------------------------------------------------------------------------------------------------
  
  LET li_return = li_rec_cnt
  
  RETURN li_return
  
END FUNCTION 

FUNCTION sadzi140_db_get_index_forward_records(p_owner,p_table_name)
DEFINE
  p_owner      STRING,
  p_table_name STRING
DEFINE
  lo_index_forward_records T_INDEX_DATA
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_index_name STRING,
  ls_sql         STRING,
  ls_exec_sql    STRING,
  li_rec_cnt     INTEGER,
  ls_alter_table VARCHAR(30)
DEFINE
  li_return INTEGER  

  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name
  
  LET li_rec_cnt = 0

  ------------------------------------------------------------------------------------------------
  -- Index forward
  ------------------------------------------------------------------------------------------------
  &ifdef is_msv
  LET ls_sql = "select ec.dzec001 table_name,                                         ",                                
               "       ec.dzec002 index_name,                                         ",                                
               "       ec.dzec003 index_type,                                         ",                                
               "       replace(ec.dzec004,' ','') column_name                         ",                                
               "  from dzec_t ec                                                      ",                                       
               " where ec.dzec001 = '",ls_table_name.toLowerCase(),"'                 ",              
               "   and coalesce(ec.dzecstus,'Y') = 'Y'                                ",                                            
               "   and not exists (                                                   ",                                       
               "                    select 1                                          ",                                       
               "                      from dzed_t ed                                  ",                                       
               "                     where ed.dzed001 = ec.dzec001                    ",                                       
               "                       and ed.dzed002 = ec.dzec002                    ",                                       
               "                  )                                                   ",                                       
               "except                                                                ",
               "select ais.table_name table_name,                                     ",                                       
               "       ais.index_name index_name,                                     ",                                       
               "       substring(ais.uniqueness,1,1) index_type,                      ",
               "       ais.column_names column_name                                   ",
               "  from dba_indexes ais                                                ",
               " where ais.owner        = '",ls_owner.toLowerCase(),"'                ",                      
               "   and ais.table_name   = '",ls_table_name.toLowerCase(),"'           ",
               "   and ais.index_name not in (                                        ",                                       
               "                               select acs.constraint_name             ",                                       
               "                                 from dba_constraints acs             ",                                       
               "                                where acs.table_name = ais.table_name ",                                       
               "                                  and acs.owner      = ais.owner      ",                                       
               "                             )                                        "                                       
  &else
  &ifdef is_pgs
  LET ls_sql = "select ec.dzec001 table_name,                                         ",                                
               "       ec.dzec002 index_name,                                         ",                                
               "       ec.dzec003 index_type,                                         ",                                
               "       replace(ec.dzec004,' ','') column_name                         ",                                
               "  from dzec_t ec                                                      ",                                       
               " where ec.dzec001 = '",ls_table_name.toLowerCase(),"'                 ",              
               "   and coalesce(ec.dzecstus,'Y') = 'Y'                                ",                                            
               "   and not exists (                                                   ",                                       
               "                    select 1                                          ",                                       
               "                      from dzed_t ed                                  ",                                       
               "                     where ed.dzed001 = ec.dzec001                    ",                                       
               "                       and ed.dzed002 = ec.dzec002                    ",                                       
               "                  )                                                   ",                                       
               "except                                                                ",
               "select ais.table_name table_name,                                     ",                                       
               "       ais.index_name index_name,                                     ",                                       
               "       substring(ais.uniqueness,1,1) index_type,                      ",
               "       string_agg(aic.column_name,','                                 ",
               "         order by aic.column_position) column_name                    ",
               "  from dba_indexes ais                                                ",
               "  inner join dba_ind_columns aic                                      ",
               "          on aic.table_owner = ais.table_owner                        ",
               "         and aic.table_name  = ais.table_name                         ",
               "         and aic.index_name  = ais.index_name                         ",
               " where ais.owner        = '",ls_owner.toLowerCase(),"'                ",                      
               "   and ais.table_name   = '",ls_table_name.toLowerCase(),"'           ",
               "   and ais.index_name not in (                                        ",                                       
               "                               select acs.constraint_name             ",                                       
               "                                 from dba_constraints acs             ",                                       
               "                                where acs.table_name = ais.table_name ",                                       
               "                                  and acs.owner      = ais.owner      ",                                       
               "                             )                                        ",                                       
               " group by ais.table_name,ais.index_name,substring(ais.uniqueness,1,1) "
  &else
  LET ls_sql = "select upper(ec.dzec001) table_name,                                                    ",                               
               "       upper(ec.dzec002) index_name,                                                    ",                               
               "       upper(ec.dzec003) index_type,                                                    ",                               
               "       upper(replace(ec.dzec004,' ','')) column_name                                    ",                               
               "  from dzec_t ec                                                                        ",                               
               " where ec.dzec001 = '",ls_table_name.toLowerCase(),"'                                   ",                               
               "   and nvl(ec.dzecstus,'Y') = 'Y'                                                       ", #20170912                     
               "   and not exists (                                                                     ",                               
               "                    select 1                                                            ",                               
               "                      from dzed_t ed                                                    ",                               
               "                     where ed.dzed001 = ec.dzec001                                      ",                               
               "                       and ed.dzed002 = ec.dzec002                                      ",                               
               "                  )                                                                     ",                               
               " minus                                                                                  ",                               
               "select ais.table_name table_name,                                                       ",                               
               "       ais.index_name index_name,                                                       ",                               
               "       substrb(decode(ais.uniqueness,'UNIQUE','UNIQUE',ais.index_type),1,1) index_type, ",                               
               "       listagg(aics.column_name,',')                                                    ",           
               "         within group (order by ais.index_name,aics.column_position) column_name        ",           
               "  from dba_indexes ais,dba_ind_columns aics                                             ",                               
               " where ais.owner        = '",ls_owner.toUpperCase(),"'                                  ",                               
               "   and ais.table_name   = '",ls_table_name.toUpperCase(),"'                             ",                               
               "   and ais.index_type <> 'CLUSTER'                                                      ", --20230807    
               "   and aics.table_owner = ais.table_owner                                               ",                               
               "   and aics.table_name  = ais.table_name                                                ",                               
               "   and aics.index_name  = ais.index_name                                                ",                               
               "   and ais.index_name not in (                                                          ",                               
               "                               select acs.index_name                                    ", --20230807                               
               "                                 from dba_constraints acs                               ",                               
               "                                where acs.table_name = ais.table_name                   ",                               
               "                                  and acs.owner      = ais.owner                        ",                               
               "                             )                                                          ",                    
               " group by ais.table_name,ais.index_name,                                                ",
               "          substrb(decode(ais.uniqueness,'UNIQUE','UNIQUE',ais.index_type),1,1)          "
  &endif
  &endif
               
  PREPARE lpre_get_index_forward_records FROM ls_sql
  DECLARE lcur_get_index_forward_records CURSOR FOR lpre_get_index_forward_records

  BEGIN WORK
  
  OPEN lcur_get_index_forward_records
  FOREACH lcur_get_index_forward_records INTO lo_index_forward_records.*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET ls_index_name = lo_index_forward_records.COLUMN_NAME
    
    LET ls_exec_sql = "update dzec_t                                      ",
                      "   set dzec005 = 'N'                               ",
                      " where dzec001 = '",ls_table_name.toLowerCase(),"' ",
                      "   and dzec002 = '",ls_index_name.toLowerCase(),"' "

    CALL sadzi140_db_exec_SQL_no_commit(ls_exec_sql,FALSE)                      

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_index_forward_records
  
  COMMIT WORK
  
  FREE lpre_get_index_forward_records
  FREE lcur_get_index_forward_records  
  -------------------------------------------------------------------------------------------------
  
  LET li_return = li_rec_cnt
  
  RETURN li_return
  
END FUNCTION 

FUNCTION sadzi140_db_get_index_backward_records(p_owner,p_table_name)
DEFINE
  p_owner      STRING,
  p_table_name STRING
DEFINE
  lo_index_backward_records T_INDEX_DATA
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_index_name STRING,
  ls_sql         STRING,
  ls_exec_sql    STRING,
  li_rec_cnt     INTEGER,
  ls_alter_table VARCHAR(30)
DEFINE
  li_return INTEGER  

  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name
  
  LET li_rec_cnt = 0

  ------------------------------------------------------------------------------------------------
  -- Index backward
  ------------------------------------------------------------------------------------------------
  &ifdef is_msv
  LET ls_sql = 
               "select ais.table_name table_name,                                     ",                                       
               "       ais.index_name index_name,                                     ",                                       
               "       substring(ais.uniqueness,1,1) index_type,                      ",
               "       ais.column_names column_name                                   ",
               "  from dba_indexes ais                                                ",
               " where ais.owner        = '",ls_owner.toLowerCase(),"'                ",                      
               "   and ais.table_name   = '",ls_table_name.toLowerCase(),"'           ",
               "   and ais.index_name not in (                                        ",                                       
               "                               select acs.constraint_name             ",                                       
               "                                 from dba_constraints acs             ",                                       
               "                                where acs.table_name = ais.table_name ",                                       
               "                                  and acs.owner      = ais.owner      ",                                       
               "                             )                                        ",
               "except                                                                ",
               "select ec.dzec001 table_name,                                         ",                                
               "       ec.dzec002 index_name,                                         ",                                
               "       ec.dzec003 index_type,                                         ",                                
               "       replace(ec.dzec004,' ','') column_name                         ",                                
               "  from dzec_t ec                                                      ",                                       
               " where ec.dzec001 = '",ls_table_name.toLowerCase(),"'                 ",              
               "   and coalesce(ec.dzecstus,'Y') = 'Y'                                ",                                            
               "   and not exists (                                                   ",                                       
               "                    select 1                                          ",                                       
               "                      from dzed_t ed                                  ",                                       
               "                     where ed.dzed001 = ec.dzec001                    ",                                       
               "                       and ed.dzed002 = ec.dzec002                    ",                                       
               "                  )                                                   "                                       
  &else               
  &ifdef is_pgs
  LET ls_sql = 
               "select ais.table_name table_name,                                     ",                                       
               "       ais.index_name index_name,                                     ",                                       
               "       substring(ais.uniqueness,1,1) index_type,                      ",
               "       string_agg(aic.column_name,','                                 ",
               "         order by aic.column_position) column_name                    ",
               "  from dba_indexes ais                                                ",
               "  inner join dba_ind_columns aic                                      ",
               "          on aic.table_owner = ais.table_owner                        ",
               "         and aic.table_name  = ais.table_name                         ",
               "         and aic.index_name  = ais.index_name                         ",
               " where ais.owner        = '",ls_owner.toLowerCase(),"'                ",                      
               "   and ais.table_name   = '",ls_table_name.toLowerCase(),"'           ",
               "   and ais.index_name not in (                                        ",                                       
               "                               select acs.constraint_name             ",                                       
               "                                 from dba_constraints acs             ",                                       
               "                                where acs.table_name = ais.table_name ",                                       
               "                                  and acs.owner      = ais.owner      ",                                       
               "                             )                                        ",
               " group by ais.table_name,ais.index_name,substring(ais.uniqueness,1,1) ",
               "except                                                                ",
               "select ec.dzec001 table_name,                                         ",                                
               "       ec.dzec002 index_name,                                         ",                                
               "       ec.dzec003 index_type,                                         ",                                
               "       replace(ec.dzec004,' ','') column_name                         ",                                
               "  from dzec_t ec                                                      ",                                       
               " where ec.dzec001 = '",ls_table_name.toLowerCase(),"'                 ",              
               "   and coalesce(ec.dzecstus,'Y') = 'Y'                                ",                                            
               "   and not exists (                                                   ",                                       
               "                    select 1                                          ",                                       
               "                      from dzed_t ed                                  ",                                       
               "                     where ed.dzed001 = ec.dzec001                    ",                                       
               "                       and ed.dzed002 = ec.dzec002                    ",                                       
               "                  )                                                   "                                       
  &else               
  LET ls_sql = "select ais.table_name table_name,                                                       ",                               
               "       ais.index_name index_name,                                                       ",                               
               "       substrb(decode(ais.uniqueness,'UNIQUE','UNIQUE',ais.index_type),1,1) index_type, ",                               
               "       listagg(aics.column_name,',')                                                    ",          
               "         within group (order by ais.index_name,aics.column_position) column_name        ",          
               "  from dba_indexes ais,dba_ind_columns aics                                             ",                               
               " where ais.owner        = '",ls_owner.toUpperCase(),"'                                  ",                               
               "   and ais.table_name   = '",ls_table_name.toUpperCase(),"'                             ",                               
               "   and ais.index_type <> 'CLUSTER'                                                      ", --20230807    
               "   and aics.table_owner = ais.table_owner                                               ",                               
               "   and aics.table_name  = ais.table_name                                                ",                               
               "   and aics.index_name  = ais.index_name                                                ",                               
               "   and ais.index_name not in (                                                          ",                               
               "                               select acs.index_name                                    ", --20230807                               
               "                                 from dba_constraints acs                               ",                               
               "                                where acs.table_name = ais.table_name                   ",                               
               "                                  and acs.owner      = ais.owner                        ",                               
               "                             )                                                          ",                               
               " group by ais.table_name,ais.index_name,                                                ",          
               "          substrb(decode(ais.uniqueness,'UNIQUE','UNIQUE',ais.index_type),1,1)          ",          
               " minus                                                                                  ",                               
               "select upper(ec.dzec001) table_name,                                                    ",                               
               "       upper(ec.dzec002) index_name,                                                    ",                               
               "       upper(ec.dzec003) index_type,                                                    ",                               
               "       upper(replace(ec.dzec004,' ','')) column_name                                    ",                               
               "  from dzec_t ec                                                                        ",                      
               " where ec.dzec001 = '",ls_table_name.toLowerCase(),"'                                   ",                               
               "   and nvl(ec.dzecstus,'Y') = 'Y'                                                       ", #20170912                              
               "   and not exists (                                                                     ",                               
               "                    select 1                                                            ",                               
               "                      from dzed_t ed                                                    ",                               
               "                     where ed.dzed001 = ec.dzec001                                      ",                     
               "                       and ed.dzed002 = ec.dzec002                                      ",                     
               "                  )                                                                     "                     
  &endif 
  &endif
  
  PREPARE lpre_get_index_backward_records FROM ls_sql
  DECLARE lcur_get_index_backward_records CURSOR FOR lpre_get_index_backward_records

  OPEN lcur_get_index_backward_records
  FOREACH lcur_get_index_backward_records INTO lo_index_backward_records.*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_index_backward_records
  
  FREE lpre_get_index_backward_records
  FREE lcur_get_index_backward_records  
  -------------------------------------------------------------------------------------------------
  
  LET li_return = li_rec_cnt
  
  RETURN li_return
  
END FUNCTION 

FUNCTION sadzi140_db_get_column_id_status(p_owner,p_table_name)
DEFINE
  p_owner      STRING,
  p_table_name STRING
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_key_name    STRING,
  ls_sql         STRING,
  ls_exec_sql    STRING,
  li_rec_cnt     INTEGER,
  li_colid_cnt   INTEGER,
  ls_alter_table VARCHAR(30)
DEFINE
  li_return INTEGER  

  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name
  
  LET li_rec_cnt = 0

  ------------------------------------------------------------------------------------------------
  -- key backward
  ------------------------------------------------------------------------------------------------

  &ifdef is_msv
  LET ls_sql = "select 1                                                           ",             
               "  from (                                                           ",             
               "         select atc.column_id   row_num,                           ",            
               "                atc.column_name column_name                        ",            
               "           from dba_tab_columns atc                                ",            
               "          where atc.owner = '",ls_owner.toLowerCase(),"'           ",  
               "            and atc.table_name = '",ls_table_name.toLowerCase(),"' ",
               "            and atc.column_name <> 'DUMMY'                         ",            
               "         except                                                    ",
               "         select eb.dzeb021        row_num,                         ",             
               "                eb.dzeb002 column_name                             ",      
               "           from dzeb_t eb                                          ",             
               "         where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'      ",
               "         union                                                     ",             
               "         select eb.dzeb021        row_num,                         ",             
               "                eb.dzeb002 column_name                             ",      
               "           from dzeb_t eb                                          ",             
               "         where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'      ",
               "         except                                                    ",              
               "         select atc.column_id   row_num,                           ",             
               "                atc.column_name column_name                        ",             
               "           from dba_tab_columns atc                                ",             
               "          where atc.owner = '",ls_owner.toLowerCase(),"'           ",   
               "            and atc.table_name = '",ls_table_name.toLowerCase(),"' ",
               "            and atc.column_name <> 'DUMMY'                         ",             
               "       ) col_status                                                "              
  &else
  &ifdef is_pgs
  LET ls_sql = "select 1                                                           ",             
               "  from (                                                           ",             
               "         select atc.column_id   row_num,                           ",            
               "                atc.column_name column_name                        ",            
               "           from dba_tab_columns atc                                ",            
               "          where atc.owner = '",ls_owner.toLowerCase(),"'           ",  
               "            and atc.table_name = '",ls_table_name.toLowerCase(),"' ",
               "            and atc.column_name <> 'dummy'                         ",            
               "         except                                                    ",
               "         select eb.dzeb021        row_num,                         ",             
               "                eb.dzeb002 column_name                             ",      
               "           from dzeb_t eb                                          ",             
               "         where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'      ",
               "         union                                                     ",             
               "         select eb.dzeb021        row_num,                         ",             
               "                eb.dzeb002 column_name                             ",      
               "           from dzeb_t eb                                          ",             
               "         where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'      ",
               "         except                                                    ",              
               "         select atc.column_id   row_num,                           ",             
               "                atc.column_name column_name                        ",             
               "           from dba_tab_columns atc                                ",             
               "          where atc.owner = '",ls_owner.toLowerCase(),"'           ",   
               "            and atc.table_name = '",ls_table_name.toLowerCase(),"' ",
               "            and atc.column_name <> 'dummy'                         ",             
               "       ) col_status                                                "              
  &else
  LET ls_sql = "SELECT 1                                                    ",
               "  FROM (                                                    ",
               "          SELECT ATC.COLUMN_ID   ROW_NUM,                   ",           
               "                 ATC.COLUMN_NAME COLUMN_NAME                ",           
               "            FROM DBA_TAB_COLUMNS ATC                        ",           
               "           WHERE ATC.OWNER = UPPER('",ls_owner,"')          ",
               "             AND ATC.TABLE_NAME = UPPER('",ls_table_name,"')",
               "             AND ATC.COLUMN_NAME <> 'DUMMY'                 ",           
               "          MINUS                                             ",
               "         SELECT EB.DZEB021        ROW_NUM,                  ",          
               "                UPPER(EB.DZEB002) COLUMN_NAME               ",          
               "           FROM DZEB_T EB                                   ",          
               "         WHERE EB.DZEB001 = LOWER('",ls_table_name,"')      ",
               "         UNION                                              ",
               "         SELECT EB.DZEB021        ROW_NUM,                  ",          
               "                UPPER(EB.DZEB002) COLUMN_NAME               ",          
               "           FROM DZEB_T EB                                   ",          
               "         WHERE EB.DZEB001 = LOWER('",ls_table_name,"')      ",
               "         MINUS                                              ",
               "         SELECT ATC.COLUMN_ID   ROW_NUM,                    ",          
               "                ATC.COLUMN_NAME COLUMN_NAME                 ",          
               "           FROM DBA_TAB_COLUMNS ATC                         ",          
               "          WHERE ATC.OWNER = UPPER('",ls_owner,"')           ",
               "            AND ATC.TABLE_NAME = UPPER('",ls_table_name,"') ",
               "            AND ATC.COLUMN_NAME <> 'DUMMY'                  ",
               "       ) COL_STATUS                                         "
  &endif 
  &endif 
                
  PREPARE lpre_get_column_id_status FROM ls_sql
  DECLARE lcur_get_column_id_status CURSOR FOR lpre_get_column_id_status

  OPEN lcur_get_column_id_status
  FOREACH lcur_get_column_id_status INTO li_colid_cnt
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_column_id_status
  
  FREE lpre_get_column_id_status
  FREE lcur_get_column_id_status  
  -------------------------------------------------------------------------------------------------
  
  LET li_return = li_rec_cnt
  
  RETURN li_return
  
END FUNCTION 

FUNCTION sadzi140_db_get_db_info()
DEFINE
  lo_db_info  T_DB_INFO,
  ls_sql      STRING,
  ls_db_type  STRING
DEFINE
  lo_return T_DB_INFO  

  &ifdef is_msv
  LET ls_sql = "select db_name()   db_name,",
               "       system_user sys_user"
  &else
  &ifdef is_pgs
  LET ls_sql = "select current_database() as db_name,  ",
               "       current_user       as user_name "
  &else
  LET ls_db_type = sadzi140_util_get_db_type()
  CASE
    WHEN ls_db_type = cs_db_driver_type_kdb --20240521
      LET ls_sql = "select upper(current_database) as db_name,  ", 
                   "       upper(current_user)     as user_name ",
                   "  from dual                                 "
    WHEN ls_db_type = cs_db_driver_type_dmg --20230804
      LET ls_sql = "select upper(cur_database) as db_name,  ", 
                   "       current_user        as user_name ",
                   "  from dual                             "
  OTHERWISE  
    LET ls_sql = "select ora_database_name db_name, ",
                 "       user              sys_user ", 
                 "  from dual                       "
  END CASE 
  &endif
  &endif

  PREPARE lpre_get_db_info FROM ls_sql
  DECLARE lcur_get_db_info CURSOR FOR lpre_get_db_info

  INITIALIZE lo_db_info TO NULL
  
  OPEN lcur_get_db_info
  FETCH lcur_get_db_info INTO lo_db_info.*
  CLOSE lcur_get_db_info
  
  FREE lpre_get_db_info
  FREE lcur_get_db_info  

  LET lo_return.* = lo_db_info.*
  
  RETURN lo_return.*
  
END FUNCTION

FUNCTION sadzi140_db_connect_db(p_db)
DEFINE 
  p_db STRING
DEFINE
  lb_return BOOLEAN   

  LET lb_return = TRUE
  
  TRY
    DISCONNECT CURRENT
  CATCH
    DISPLAY cs_message_tag,"No any connection, skip disconnect."
  END TRY  

  TRY 
    CONNECT TO p_db
  CATCH
    LET lb_return = FALSE
    DISPLAY cs_error_tag,"Could not connect to database, please contact DBA !"
  END TRY 

  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_db_get_table_columns(p_owner,p_table_name,p_cursor)
DEFINE
  p_owner      STRING,
  p_table_name STRING,
  p_cursor     STRING 
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_cursor      STRING, 
  li_rec_cnt     INTEGER,
  lv_column      VARCHAR(1024),
  ls_column      STRING,  
  ls_sql         STRING,
  li_column_num  INTEGER,
  ls_CRLF        STRING,
  ls_multi_cols  STRING,
  ls_cursor_cols STRING
DEFINE  
  lb_return BOOLEAN

  &ifdef is_msv
  LET ls_owner      = p_owner.toLowerCase()     
  LET ls_table_name = p_table_name.toLowerCase()
  &else
  &ifdef is_pgs
  LET ls_owner      = p_owner.toLowerCase()     
  LET ls_table_name = p_table_name.toLowerCase()
  &else
  LET ls_owner      = p_owner.toUpperCase()     
  LET ls_table_name = p_table_name.toUpperCase()
  &endif  
  &endif  
  
  LET ls_cursor = p_cursor

  LET ls_multi_cols  = ""
  LET ls_cursor_cols = ""
  
  LET li_rec_cnt = 1
  LET li_column_num = 0

  LET ls_sql = "select atc.column_name                     ",
               "  from dba_tab_columns atc                 ",
               " where atc.owner      = '",ls_owner,"'     ",
               "   and atc.table_name = '",ls_table_name,"'",
               " order by atc.column_id                    "
                                                     
  PREPARE lpre_get_table_columns FROM ls_sql
  DECLARE lcur_get_table_columns CURSOR FOR lpre_get_table_columns

  OPEN lcur_get_table_columns
  FOREACH lcur_get_table_columns INTO lv_column
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_column_num = li_column_num + 1

    IF li_column_num >= 10 THEN 
      LET ls_CRLF = "\n"
    ELSE
      LET ls_CRLF = ""
    END IF 
    
    LET ls_column = lv_column
    LET ls_multi_cols = ls_multi_cols,ls_column,",",ls_CRLF
    LET ls_cursor_cols = ls_cursor_cols,ls_cursor,".",ls_column,",",ls_CRLF

    IF li_column_num >= 10 THEN
      LET li_column_num = 0
    END IF 
    
  END FOREACH
  CLOSE lcur_get_table_columns
  
  LET ls_multi_cols = ls_multi_cols.subString(1,ls_multi_cols.getLength()-1)
  IF ls_multi_cols.subString(ls_multi_cols.getLength(),ls_multi_cols.getLength()) = "," THEN
    LET ls_multi_cols = ls_multi_cols.subString(1,ls_multi_cols.getLength()-1)
  END IF  
  
  LET ls_cursor_cols = ls_cursor_cols.subString(1,ls_cursor_cols.getLength()-1)
  IF ls_cursor_cols.subString(ls_cursor_cols.getLength(),ls_cursor_cols.getLength()) = "," THEN
    LET ls_cursor_cols = ls_cursor_cols.subString(1,ls_cursor_cols.getLength()-1)
  END IF   
  
  FREE lpre_get_table_columns
  FREE lcur_get_table_columns  

  RETURN ls_multi_cols,ls_cursor_cols
  
END FUNCTION

FUNCTION sadzi140_db_get_DZEA_info(p_table_name)
DEFINE 
  p_table_name STRING
DEFINE
  ls_table_name STRING,
  ls_sql        STRING,
  lo_DZEA_info  T_DZEA_INFO
DEFINE  
  lo_return T_DZEA_INFO

  LET ls_table_name = p_table_name
  
  LET ls_sql = "select ea.dzea001,ea.dzea002,ea.dzea003,ea.dzea004,ea.dzea005, ",
               "       ea.dzea006,ea.dzea007,ea.dzea008,ea.dzea009,ea.dzea010, ",
               "       ea.dzea011,ea.dzea012,ea.dzea013,ea.dzea014,ea.dzea015  ",
               "  from dzea_t ea                                               ",
               " where ea.dzea001 = '",ls_table_name,"'                        " 
                               
  PREPARE lpre_get_DZEA_info FROM ls_sql
  DECLARE lcur_get_DZEA_info CURSOR FOR lpre_get_DZEA_info

  OPEN lcur_get_DZEA_info
  FETCH lcur_get_DZEA_info INTO lo_DZEA_info.*
  CLOSE lcur_get_DZEA_info
  
  FREE lpre_get_DZEA_info
  FREE lcur_get_DZEA_info  

  LET lo_return.* = lo_DZEA_info.*
  
  RETURN lo_return.*
  
END FUNCTION

FUNCTION sadzi140_db_gen_move_data_script(p_table_name,p_reb_table,p_multi_cols,p_cursor_cols)
DEFINE
  p_table_name  STRING,
  p_reb_table   STRING,
  p_multi_cols  STRING,
  p_cursor_cols STRING
DEFINE
  ##########################
  ls_drop_script      STRING,
  ls_exit_sign        STRING,
  ls_drop_full        STRING,
  ##########################
  ls_table_name       STRING,
  ls_reb_table        STRING,
  ls_multi_cols       STRING,
  ls_cursor_cols      STRING,
  ls_random_name      STRING,
  ls_tempdir          STRING,
  lchannel_write      base.Channel,
  ls_sript_filename   STRING,
  ls_procedure_script STRING,
  ls_separator        STRING,
  ls_db_type          STRING #20240521
DEFINE
  ls_return STRING  
  
  LET ls_table_name  = p_table_name
  LET ls_reb_table   = p_reb_table
  LET ls_multi_cols  = p_multi_cols
  LET ls_cursor_cols = p_cursor_cols

  LET ls_separator = os.Path.separator()
  
  LET ls_tempdir = FGL_GETENV("TEMPDIR")
  CALL sadzi140_db_get_GUID() RETURNING ls_random_name 
  LET ls_sript_filename = ls_tempdir,ls_separator,ls_table_name,ls_random_name,".mve"
  
  LET lchannel_write = base.Channel.create()
  CALL lchannel_write.setDelimiter("")
  
  LET ls_exit_sign = sadzi140_util_get_exit_sign() --cs_exit_sig

  &ifdef is_msv
  LET ls_procedure_script = "","\n"
  &else
  &ifdef is_pgs
  LET ls_procedure_script = "","\n"
  &else
  LET ls_procedure_script = "-- Fill rebuild data procedure                                       ","\n",
                            "SET SERVEROUTPUT ON                                                  ","\n",
                            "DECLARE                                                              ","\n",
                            "  ------------------------------------------------------------------ ","\n",
                            "  cursor c_datas IS                                                  ","\n",
                            "    SELECT *                                                         ","\n",
                            "      FROM ",ls_reb_table.toUpperCase(),";                           ","\n",
                            "  r_datas c_datas%ROWTYPE;                                           ","\n",
                            "  ------------------------------------------------------------------ ","\n",
                            "  cs_ERROR_TAG        CONSTANT VARCHAR2(20) := '",cs_error_tag,"';   ","\n",                                      
                            "  cs_WARNING_TAG      CONSTANT VARCHAR2(20) := '",cs_warning_tag,"'; ","\n",                                        
                            "  cs_SUCCESS_TAG      CONSTANT VARCHAR2(20) := '",cs_success_tag,"'; ","\n",
                            "  cs_MSG              CONSTANT VARCHAR2(20) := 'Moving Data.';       ","\n",
                            "  ls_SQL              VARCHAR2(1024);                                ","\n",
                            "  ls_columns          VARCHAR2(4000);                                ","\n",
                            "  ls_TableName        VARCHAR2(30);                                  ","\n",
                            "  ls_RebuildTableName VARCHAR2(30);                                  ","\n",
                            "  ls_Owner            VARCHAR2(30);                                  ","\n",
                            "  ln_count            INTEGER;                                       ","\n",
                            "  ln_count_total      INTEGER;                                       ","\n",
                            "  ------------------------------------------------------------------ ","\n",
                            "BEGIN                                                                ","\n",              
                            "  DBMS_OUTPUT.ENABLE(100000);                                        ","\n",
                            "                                                                     ","\n",
                            "  ln_count := 1;                                                     ","\n",
                            "  ln_count_total := 0;                                               ","\n",
                            "                                                                     ","\n",
                            "  open c_datas;                                                      ","\n",
                            "  loop                                                               ","\n",
                            "    fetch c_datas into r_datas;                                      ","\n",
                            "    if c_datas%notfound then                                         ","\n",
                            "       exit;                                                         ","\n",
                            "    end if;                                                          ","\n",
                            "                                                                     ","\n",
                            "    INSERT INTO ",ls_table_name.toUpperCase(),"                      ","\n",
                            "    (                                                                ","\n",
                            ls_multi_cols,                                                        
                            "    )                                                                ","\n",
                            "    VALUES                                                           ","\n",
                            "    (                                                                ","\n",
                            ls_cursor_cols,                                                       
                            "    );                                                               ","\n",
                            "                                                                     ","\n",
                            "    ln_count := ln_count + 1;                                        ","\n",
                            "    if ln_count >= 1000 then                                         ","\n",
                            "      commit;                                                        ","\n",
                            "      ln_count_total := ln_count_total + ln_count;                   ","\n",
                            "      ln_count := 1;                                                 ","\n",
                            "    end if;                                                          ","\n",
                            "                                                                     ","\n",
                            "  end loop;                                                          ","\n",
                            "                                                                     ","\n",
                            "  ln_count_total := ln_count_total + ln_count;                       ","\n",
                            "                                                                     ","\n",
                            "  commit;                                                            ","\n",
                            "                                                                     ","\n",
                            "  DBMS_OUTPUT.PUT_LINE(cs_SUCCESS_TAG||' '||cs_MSG);                 ","\n",                                 
                            "                                                                     ","\n",
                            "EXCEPTION                                                            ","\n",              
                            "  WHEN OTHERS THEN ROLLBACK;                                         ","\n",            
                            "  DBMS_OUTPUT.PUT_LINE(cs_ERROR_TAG||' '||cs_MSG);                   ","\n",
                            "  DBMS_OUTPUT.PUT_LINE(SQLERRM);                                     ","\n",
                            "END;                                                                 ","\n",              
                            "/                                                                    ","\n"                                                
  LET ls_db_type = sadzi140_util_get_db_type() #20240521
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb #20240521
      LET ls_procedure_script = "","\n"
  --WHEN cs_db_driver_type_dmg #20240521
    --與Oracle共用
  OTHERWISE   
    --不修改scripts
  END CASE  
  &endif
  &endif
  
  LET ls_drop_full = ls_procedure_script,"\n",
                     ls_exit_sign
  
  #開檔寫入
  TRY
    CALL lchannel_write.openFile(ls_sript_filename, "w" )
    CALL lchannel_write.write(ls_drop_full)
    #關檔
    CALL lchannel_write.close()
  CATCH
    DISPLAY "Generate Script Error !!"
  END TRY  

  LET ls_return = ls_sript_filename
  
  RETURN ls_return

END FUNCTION

FUNCTION sadzi140_db_get_column_data_to_mem(p_owner,p_table_name)
DEFINE
  p_owner       STRING,
  p_table_name  STRING
DEFINE
  lo_columns_data_in_mem DYNAMIC ARRAY OF T_COLUMNS_DATA_IN_MEM
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_sql         STRING,
  ls_exec_sql    STRING,
  ls_ora_coldef  STRING, #20240529
  li_rec_cnt     INTEGER,
  ls_db_type     STRING, #20240606
  ls_alter_table VARCHAR(30)
DEFINE
  lo_return DYNAMIC ARRAY OF T_COLUMNS_DATA_IN_MEM

  LET ls_owner       = p_owner
  LET ls_table_name  = p_table_name
  
  LET li_rec_cnt = 1

  &ifdef is_msv
  LET ls_sql = "select atc.owner                      owner,               ",       
               "       atc.table_name                 table_name,          ",       
               "       atc.column_name                column_name,         ",
               "       iif(atc.nullable='Y','N','Y')  iskey,               ",
               "       atc.nullable                   nullable,            ",
               "       atc.data_length                data_length,         ",
               "       atc.data_default               data_default         ",
               "  from dba_tab_columns atc                                 ",             
               " where 1=1                                                 ",             
               "   and atc.owner       = '",ls_owner.toLowerCase(),"'      ",
               "   and atc.table_name  = '",ls_table_name.toLowerCase(),"' ",
               " order by 1,2,3                                            "             
  &else
  &ifdef is_pgs
  LET ls_sql = "select atc.owner              as owner,                                      ",
               "       atc.table_name            table_name,                                 ",
               "       atc.column_name           column_name,                                ",
               "       case atc.nullable                                                     ",
               "         when 'Y' then 'N'                                                   ",
               "       else 'Y' end              iskey,                                      ",
               "       atc.nullable              nullable,                                   ",
               "       case                                                                  ",
               "         when atc.data_type = 'text' then null                               ",
               "         when atc.data_type = 'bytea' then null                              ",
               "         when atc.data_type = 'date' then null                               ",
               "         when atc.data_type = 'int2' then null                               ",
               "         when atc.data_type = 'int4' then null                               ",
               "         when atc.data_type = 'int8' then null                               ",
               "         when atc.data_type = 'decimal' then                                 ",
               "           atc.data_precision::varchar||','||                                ",
               "           atc.data_scale::varchar                                           ",
               "       else                                                                  ",
               "         atc.data_length::varchar                                            ",
               "       end::varchar              data_length,                                ",
               "       case                                                                  ",
               "         when atc.data_default is null then null                             ",
               "         when atc.data_default = 'NULL' then null                            ",
               "         when atc.data_default = 'null' then null                            ",
               "         when atc.data_default like 'CURRENT%' then lower(atc.data_default)  ",
               "         when atc.data_default = '\"sysdate\"()' then 'sysdate'              ",
               "         when atc.data_default = 'SYSDATE' then 'sysdate'                    ",
               "       else                                                                  ",
               "         atc.data_default                                                    ",
               "       end as data_default                                                   ",
               "  from dba_tab_columns atc                                                   ",
               " where 1=1                                                                   ",
               "   and atc.owner       = '",ls_owner.toLowerCase(),"'                        ",
               "   and atc.table_name  = '",ls_table_name.toLowerCase(),"'                   ",
               " order by 1,2,3                                                              "
  &else
  LET ls_db_type = sadzi140_util_get_db_type()
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb
      LET ls_ora_coldef = " case                                                                 ",
                          "   when atc.data_default is null then null                            ",
                          "   when atc.data_default = 'NULL' then null                           ",
                          "   when atc.data_default = 'null' then null                           ",
                          "   when atc.data_default like 'CURRENT%' then lower(atc.data_default) ",
                          "   when atc.data_default = '\"sysdate\"()' then 'sysdate'             ",
                          "   when atc.data_default = 'SYSDATE' then 'sysdate'                   ",
                          " else                                                                 ",
                          "   atc.data_default                                                   ",
                          " end as data_default                                                  "
    WHEN cs_db_driver_type_dmg 
      LET ls_ora_coldef = " case                                                                 ",
                          "   when atc.data_default is null then null                            ",
                          "   when atc.data_default = 'NULL' then null                           ",
                          "   when atc.data_default = 'null' then null                           ",
                          "   when atc.data_default LIKE 'CURRENT%' then lower(atc.data_default) ",
                          "   when atc.data_default = '\"sysdate\"()' then 'sysdate'             ",
                          "   when atc.data_default = 'SYSDATE' then 'sysdate'                   ",
                          " else                                                                 ",
                          "   atc.data_default                                                   ",
                          " end as data_default                                                  "
  OTHERWISE  
    LET ls_ora_coldef = " decode(trim(ds.get_col_default(atc.table_name,atc.column_name)),'',null,       ",
                        "        trim(ds.get_col_default(atc.table_name,atc.column_name))) data_default  "
  END CASE
  
  LET ls_sql = "select atc.owner                            as owner,                                      ",
               "       atc.table_name                       as table_name,                                 ",
               "       atc.column_name                      as column_name,                                ",
               "       decode(acc.column_name,null,'N','Y') as iskey,                                      ",
               "       atc.nullable                         as nullable,                                   ",
               "       decode(instrb(atc.data_type,'TIMESTAMP'),1,'TIMESTAMP',atc.data_type) as data_type, ",
               "       replace(to_char(decode(                                                             ",
               "               atc.data_type,                                                              ",
               "               'BLOB',4000,                                                                ", #20240529
               "               'CLOB',4000,                                                                ", #20240529
               "               'INT2',NULL,                                                                ", #20240529
               "               'INT4',NULL,                                                                ", #20240529
               "               'INT8',NULL,                                                                ", #20240529
               "               'NUMBER',decode(                                                            ",
               "                                nvl(atc.data_scale,'0'),                                   ",
               "                                '0',atc.data_precision,                                    ",
               "                                atc.data_precision||'.'||atc.data_scale                    ",
               "                              ),                                                           ",
               "               'NUMERIC',DECODE(                                                           ", #20240529 begin
               "                                NVL(ATC.DATA_SCALE,'0'),                                   ",
               "                                '0',ATC.DATA_PRECISION,                                    ",
               "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE                    ", 
               "                              ),                                                           ", #20240529 end
               "               'DATE',7,                                                                   ", #20240529
               "               'TIMESTAMP(0)',7,                                                           ", #20240529
               "               'TIMESTAMP(5)',11,                                                          ", #20240529
               "               'TIMESTAMP(6)',11,                                                          ", #20240529
               "               'TIMESTAMP',decode(                                                         ",
               "                                   atc.data_scale,0,7,8,7,11                               ", #20240529 
               "                                 ),                                                        ",                
               "               'DATE',decode(                                                              ",
               "                              atc.data_length,3,7,atc.data_length                          ",
               "                            ),                                                             ",                
               "               atc.data_length                                                             ",
               "             )),'.',',')                    as data_length,                                ",
               ls_ora_coldef,
               "  from dba_tab_columns atc                                                                 ",
               "  left outer join dba_cons_columns acc on acc.owner       = atc.owner                      ",
               "                                      and acc.table_name  = atc.table_name                 ",
               "                                      and acc.column_name = atc.column_name                ",
               "                                      and acc.constraint_name like '%PK'                   ",
               " where 1=1                                                                                 ",
               "   and atc.owner       = '",ls_owner.toUpperCase(),"'                                      ",
               "   and atc.table_name  = '",ls_table_name.toUpperCase(),"'                                 ",
               "   and atc.column_name not like 'SYS%'                                                     ",
               " order by 1,2,3                                                                            "
  &endif 
  &endif 
               
  PREPARE lpre_get_columns_data_to_mem FROM ls_sql
  DECLARE lcur_get_columns_data_to_mem CURSOR FOR lpre_get_columns_data_to_mem

  OPEN lcur_get_columns_data_to_mem
  FOREACH lcur_get_columns_data_to_mem INTO lo_columns_data_in_mem[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF
    
    LET li_rec_cnt = li_rec_cnt + 1
    
  END FOREACH
  CLOSE lcur_get_columns_data_to_mem
  CALL lo_columns_data_in_mem.deleteElement(li_rec_cnt)
  
  FREE lpre_get_columns_data_to_mem
  FREE lcur_get_columns_data_to_mem  
  
  LET lo_return = lo_columns_data_in_mem
  
  RETURN lo_return
  
END FUNCTION 

FUNCTION sadzi140_db_get_db_column_data(p_owner,p_table_name,p_column_name)
DEFINE
  p_owner       STRING,
  p_table_name  STRING,
  p_column_name STRING
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_column_name STRING,
  lo_db_column_data  T_COLUMN_DATA,
  li_loop            INTEGER
DEFINE
  lo_return T_COLUMN_DATA

  &ifdef is_msv
  LET ls_owner       = p_owner.toLowerCase()
  LET ls_table_name  = p_table_name.toLowerCase()
  LET ls_column_name = p_column_name.toLowerCase()
  &else
  &ifdef is_pgs
  LET ls_owner       = p_owner.toLowerCase()
  LET ls_table_name  = p_table_name.toLowerCase()
  LET ls_column_name = p_column_name.toLowerCase()
  &else
  LET ls_owner       = p_owner.toUpperCase()
  LET ls_table_name  = p_table_name.toUpperCase()
  LET ls_column_name = p_column_name.toUpperCase()
  &endif
  &endif

  FOR li_loop = 1 TO mo_columns_data_in_mem.getLength()
    IF mo_columns_data_in_mem[li_loop].OWNER       = ls_owner AND
       mo_columns_data_in_mem[li_loop].TABLE_NAME  = ls_table_name AND
       mo_columns_data_in_mem[li_loop].COLUMN_NAME = ls_column_name THEN

       LET lo_db_column_data.COLUMN_NAME   = mo_columns_data_in_mem[li_loop].COLUMN_NAME  
       LET lo_db_column_data.ISKEY         = mo_columns_data_in_mem[li_loop].ISKEY        
       LET lo_db_column_data.NULLABLE      = mo_columns_data_in_mem[li_loop].NULLABLE     
       LET lo_db_column_data.DATA_TYPE     = mo_columns_data_in_mem[li_loop].DATA_TYPE    
       LET lo_db_column_data.DATA_LENGTH   = mo_columns_data_in_mem[li_loop].DATA_LENGTH  
       LET lo_db_column_data.DEFAULT_VALUE = mo_columns_data_in_mem[li_loop].DEFAULT_VALUE
       
       EXIT FOR
       
    END IF   
  END FOR 

  LET lo_return.* = lo_db_column_data.*
  
  RETURN lo_return.*

END FUNCTION 

FUNCTION sadzi140_db_get_design_column_data(p_owner,p_table_name,p_column_name)
DEFINE
  p_owner       STRING,
  p_table_name  STRING,
  p_column_name STRING
DEFINE
  lo_design_column_data T_COLUMN_DATA
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_column_name STRING,
  ls_sql         STRING,
  ls_exec_sql    STRING,
  li_rec_cnt     INTEGER,
  ls_alter_table VARCHAR(30)
DEFINE
  lo_return T_COLUMN_DATA

  LET ls_owner       = p_owner
  LET ls_table_name  = p_table_name
  LET ls_column_name = p_column_name
  
  LET li_rec_cnt = 0

  &ifdef is_msv
  LET ls_sql = "select eb.dzeb002                             column_name,   ", 
               "       eb.dzeb004                             iskey,         ", 
               "       case eb.dzeb004                                       ", 
               "         when 'Y' then 'N'                                   ", 
               "         when 'N' then 'Y'                                   ", 
               "       else                                                  ", 
               "         'Y'                                                 ", 
               "       end                                    nullable,      ", 
               "       td.gztd003                             data_type,     ", 
               "       case td.gztd003                                       ", 
               "         when 'varbinary'                                    ", 
               "           then 'max'                                        ", 
               "         when 'nvarchar'                                     ", 
               "           then iif(td.gztd009='TEXT','max',td.gztd008)      ", 
               "       else                                                  ", 
               "         td.gztd008                                          ", 
               "       end                                    data_length,   ", 
               "       trim(eb.dzeb012)                       data_default   ", 
               "  from dzeb_t eb                                             ", 
               "       left outer join gztd_t td on td.gztd001 = eb.dzeb006  ", 
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'        ", 
               "   and eb.dzeb002 = '",ls_column_name.toLowerCase(),"'       ", 
               " order by 1                                                  "
  &else
  &ifdef is_pgs
  LET ls_sql = "select eb.dzeb002                             column_name,   ",
               "       eb.dzeb004                             iskey,         ",
               "       case eb.dzeb004                                       ",
               "         when 'Y' then 'N'                                   ",
               "         when 'N' then 'Y'                                   ",
               "       else                                                  ",
               "         'Y'                                                 ",
               "       end                                    nullable,      ",
               "       td.gztd003                             data_type,     ",
               "       td.gztd008                             data_length,   ",
               "       trim(eb.dzeb012)                       data_default   ",
               "  from dzeb_t eb                                             ",
               "       left outer join gztd_t td on td.gztd001 = eb.dzeb006  ",
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'        ",
               "   and eb.dzeb002 = '",ls_column_name.toLowerCase(),"'       ",
               " order by 1                                                  "
  &else
  LET ls_sql = "select upper(eb.dzeb002)                      as column_name,       ",
               "       eb.dzeb004                             as iskey,             ",
               "       decode(eb.dzeb004,'Y','N','N','Y','Y') as nullable,          ",
               "       upper(td.gztd003)                      as data_type,         ",
               "       replace(to_char(to_number(replace(replace(                   ",
               "       decode(                                                      ",
               "              upper(td.gztd003),                                    ",
               "              'BLOB','4000',                                        ",
               "              'CLOB','4000',                                        ",
               "              'DATE','7',                                           ",
               "              'TIMESTAMP',decode(td.gztd008,'0','7','11'),          ",
               "              td.gztd008                                            ",
               "             ),',','.'),' ',''))),'.',',')    as data_length,       ",
               "       trim(decode(eb.dzeb012,'',null,eb.dzeb012))  as data_default ",
               "  from dzeb_t eb                                                    ",
               "       left outer join gztd_t td on td.gztd001 = eb.dzeb006         ",
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'               ",
               "   and eb.dzeb002 = '",ls_column_name.toLowerCase(),"'              ",
               " order by 1                                                         "
  &endif             
  &endif             
                
  PREPARE lpre_get_design_column_data FROM ls_sql
  DECLARE lcur_get_design_column_data CURSOR FOR lpre_get_design_column_data

  OPEN lcur_get_design_column_data
  FETCH lcur_get_design_column_data INTO lo_design_column_data.*
  CLOSE lcur_get_design_column_data
  
  FREE lpre_get_design_column_data
  FREE lcur_get_design_column_data  
  
  LET lo_return.* = lo_design_column_data.*
  
  RETURN lo_return.*
  
END FUNCTION 

FUNCTION sadzi140_db_update_module_code(p_table_name,p_module,p_dgenv)
DEFINE
  p_table_name STRING,
  p_module     STRING,
  p_dgenv      STRING
DEFINE
  ls_table_name    STRING,
  ls_module        STRING,
  ls_dgenv         STRING,
  ls_exec_sql      STRING

  LET ls_table_name = p_table_name
  LET ls_module     = p_module
  LET ls_dgenv      = p_dgenv

  LET ls_exec_sql = " update dzea_t                                      ",
                    "    set dzea003 = '",ls_module.toUpperCase(),"',    ",
                    "        dzea015 = '",ls_dgenv.toLowerCase(),"'      ",    
                    "  where dzea001 = '",ls_table_name.toLowerCase(),"' "
                    
  CALL sadzi140_db_exec_SQL(ls_exec_sql,TRUE)                      
  
END FUNCTION 

#檢核欄位是否有客制資料
FUNCTION sadzi140_db_check_column_if_customize(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_count  INTEGER,
  lb_return     BOOLEAN

  LET ls_table_name = p_table_name.toLowerCase()

  #取得資料筆數
  LET ls_sql = "select count(1) cnts                         ",
               "  from dzeb_t eb                             ",
               " where eb.dzeb001 = '",ls_table_name,"'      ", 
               "   and eb.dzeb029 = '",cs_dgenv_customize,"' " 
                
  PREPARE lpre_check_column_if_customize FROM ls_sql
  DECLARE lcur_check_column_if_customize CURSOR FOR lpre_check_column_if_customize
  OPEN lcur_check_column_if_customize
  FETCH lcur_check_column_if_customize INTO li_rec_count
  CLOSE lcur_check_column_if_customize
  FREE lcur_check_column_if_customize
  FREE lpre_check_column_if_customize

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION

#檢核Key是否有客制資料
FUNCTION sadzi140_db_check_key_if_customize(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_count  INTEGER,
  lb_return     BOOLEAN

  LET ls_table_name = p_table_name.toLowerCase()

  #取得資料筆數
  LET ls_sql = "select count(1) cnts                         ",
               "  from dzed_t ed                             ",
               " where ed.dzed001 = '",ls_table_name,"'      ", 
               "   and ed.dzed008 = '",cs_dgenv_customize,"' " 
                
  PREPARE lpre_check_key_if_customize FROM ls_sql
  DECLARE lcur_check_key_if_customize CURSOR FOR lpre_check_key_if_customize
  OPEN lcur_check_key_if_customize
  FETCH lcur_check_key_if_customize INTO li_rec_count
  CLOSE lcur_check_key_if_customize
  FREE lcur_check_key_if_customize
  FREE lpre_check_key_if_customize

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_db_check_index_if_customize(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_count  INTEGER,
  lb_return     BOOLEAN

  LET ls_table_name = p_table_name.toLowerCase()

  #取得資料筆數
  LET ls_sql = "select count(1) cnts                         ",
               "  from dzec_t ec                             ",
               " where ec.dzec001 = '",ls_table_name,"'      ", 
               "   and ec.dzec006 = '",cs_dgenv_customize,"' " 
                
  PREPARE lpre_check_index_if_customize FROM ls_sql
  DECLARE lcur_check_index_if_customize CURSOR FOR lpre_check_index_if_customize
  OPEN lcur_check_index_if_customize
  FETCH lcur_check_index_if_customize INTO li_rec_count
  CLOSE lcur_check_index_if_customize
  FREE lcur_check_index_if_customize
  FREE lpre_check_index_if_customize

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_db_grant_APS_privilege(p_db_connstr,p_master_db)
DEFINE 
  p_db_connstr  T_DB_CONNSTR,
  p_master_db   STRING
DEFINE 
  lo_db_connstr  T_DB_CONNSTR,
  ls_master_db   STRING,
  ls_message     STRING,
  ls_separator   STRING,
  ls_db_type     STRING #20240521

  LET lo_db_connstr.* = p_db_connstr.*
  LET ls_master_db    = p_master_db

  LET ls_separator = os.Path.separator()

  &ifdef is_msv
  DISPLAY cs_message_tag,"SQL server do not grant privileges to APS." 
  &else
  &ifdef is_pgs
  DISPLAY cs_message_tag,"PostgreSQL do not grant privileges to APS." 
  &else
  LET ls_db_type = sadzi140_util_get_db_type() #20240521
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb #20240521
      DISPLAY cs_message_tag,"KDB do not grant privileges to APS." 
  --WHEN cs_db_driver_type_dmg #20240521
    --與Oracle通用
  OTHERWISE   
    #呼叫產生 Grant 的SQL
    LET lo_db_connstr.db_sql_filename = FGL_GETENV("APSE"),ls_separator,"script",ls_separator,"dsgrant.sql"
    IF os.Path.exists(lo_db_connstr.db_sql_filename) THEN 
      #DISPLAY cs_information_tag," Grant ",lo_db_connstr.db_username," privileges for APS system." #20170425
      CALL sadzi140_db_sqlplus(lo_db_connstr.*) RETURNING ls_message
    ELSE
      #20170405 begin
      LET lo_db_connstr.db_sql_filename = FGL_GETENV("APSE"),ls_separator,"std",ls_separator,"script",ls_separator,"dsgrant.sql"
      IF os.Path.exists(lo_db_connstr.db_sql_filename) THEN 
        #DISPLAY cs_information_tag," Grant ",lo_db_connstr.db_username," privileges for APS system." #20170425
        CALL sadzi140_db_sqlplus(lo_db_connstr.*) RETURNING ls_message
      ELSE 
      #20170405 end
        DISPLAY cs_warning_tag,"Can not find the APSE script file for ",lo_db_connstr.db_username," !!"
      END IF #20170405 
    END IF  
  END CASE  
  &endif
  &endif
  
END FUNCTION

FUNCTION sadzi140_db_update_column_name(p_table_name,p_old_col_name,p_new_col_name)
DEFINE
  p_table_name   STRING,
  p_old_col_name STRING,
  p_new_col_name STRING
DEFINE
  ls_table_name   STRING,
  ls_old_col_name STRING,
  ls_new_col_name STRING,
  ls_exec_sql     STRING

  LET ls_table_name   = p_table_name.toLowerCase()
  LET ls_old_col_name = p_old_col_name.toLowerCase()
  LET ls_new_col_name = p_new_col_name.toLowerCase()

  LET ls_exec_sql = " update dzeb_t                         ",
                    "    set dzeb002 = '",ls_new_col_name,"'",
                    "  where dzeb001 = '",ls_table_name,"'  ",
                    "    and dzeb002 = '",ls_old_col_name,"'"
                    
  CALL sadzi140_db_exec_SQL(ls_exec_sql,FALSE)
  
END FUNCTION 

FUNCTION sadzi140_db_get_rebuild_index_list(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name   STRING,
  ls_sql          STRING,
  li_rec_count    INTEGER,
  ls_rebuild_list STRING,
  ls_rebuild_sql  VARCHAR(1024),
  ls_rebuild      STRING
DEFINE
  ls_return STRING
  
  LET ls_table_name = p_table_name
  
  LET li_rec_count = 0
  LET ls_rebuild_list = ""

  #取得 Rebuild Index 清單
  &ifdef is_msv
  LET ls_rebuild = "'alter index '+ec.dzec002+' rebuild'" #20191209
  &else
  --Oracle & PostgreSQL 共用
  LET ls_rebuild = "'alter index '||ec.dzec002||' rebuild'"
  &endif
  
  LET ls_sql = "select ",ls_rebuild," rebuild_sql                                ",
               "  from dzec_t ec                                                 ",
               " where exists (                                                  ",
               "                select 1                                         ",
               "                  from user_indexes uis                          ",
               "                 where uis.index_name = upper(ec.dzec002)        ",
               "                   and uis.table_name = upper(ec.dzec001)        ",
               "                   and uis.table_owner = (select user from dual) ",
               "              )                                                  ",
               "   and ec.dzec001 = '",ls_table_name,"'                          "
                                                                                 
  PREPARE lpre_get_rebuild_index_list FROM ls_sql                                
  DECLARE lcur_get_rebuild_index_list CURSOR FOR lpre_get_rebuild_index_list
  
  OPEN lcur_get_rebuild_index_list
  FOREACH lcur_get_rebuild_index_list INTO ls_rebuild_sql
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_count = li_rec_count + 1

    &ifdef is_msv
    LET ls_rebuild_list = ls_rebuild_list,ls_rebuild_sql,"\n","go","\n"
    &else
    --Oracle & PGS 共用
    LET ls_rebuild_list = ls_rebuild_list,ls_rebuild_sql,";","\n"
    &endif
    
  END FOREACH
  CLOSE lcur_get_rebuild_index_list
  
  FREE lcur_get_rebuild_index_list
  FREE lpre_get_rebuild_index_list

  LET ls_return = ls_rebuild_list
  
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_get_column_type_data_type(p_column_type)
DEFINE
  p_column_type  STRING
DEFINE
  ls_column_type STRING,
  ls_sql         STRING,
  lo_data_type   T_DATA_TYPE
DEFINE  
  lo_return      T_DATA_TYPE

  LET ls_column_type = p_column_type

  #取得Data Type 相關資訊
  &ifdef is_msv
  LET ls_sql = "select tdm.gztd003                                     dt_type,                                         ",       
               "       tdm.gztd008                                     dt_length,                                       ",       
               "       tdm.dt_integer                                  dt_integer,                                      ",       
               "       tdm.dt_decimal                                  dt_decimal,                                      ",       
               "       (coalesce(tdm.dt_integer,0) - coalesce(tdm.dt_decimal,0)) dt_pdecimal                            ",                 
               "  from (                                                                                                ",       
               "        select tdx.gztd003                                                gztd003,                      ", 
               "               tdx.gztd008                                                gztd008,                      ", 
               "               case tdx.comma_pos                                                                       ", 
               "                 when 0 then tdx.gztd008                                                                ", 
               "               else                                                                                     ", 
               "                 substring(tdx.gztd008,1,tdx.comma_pos-1)                                               ", 
               "               end                                                        dt_integer,                   ", 
               "               case tdx.comma_pos                                                                       ", 
               "                 when 0 then 0                                                                          ", 
               "               else                                                                                     ", 
               "                 substring(tdx.gztd008,tdx.comma_pos+1,len(tdx.gztd008))                                ", 
               "               end                                                        dt_decimal                    ", 
               "          from (                                                                                        ",       
               "                select td.gztd003                                    gztd003,                           ", 
               "                       replace(replace(td.gztd008,' ',''),'.',',')   gztd008,                           ", 
               "                       coalesce(charindex(',',replace(replace(td.gztd008,' ',''),'.',',')),0) comma_pos ",               
               "                  from gztd_t td                                                                        ",       
               "                 where td.gztd001 = '",ls_column_type.TRIM(),"'                                         ", 
               "               ) tdx                                                                                    ",       
               "       ) tdm                                                                                            " 
  &else
  &ifdef is_pgs
  LET ls_sql = "select tdm.gztd003                     dt_type,                                                      ",
               "       tdm.gztd008                     dt_length,                                                    ",
               "       tdm.dt_integer                  dt_integer,                                                   ",
               "       tdm.dt_decimal                  dt_decimal,                                                   ",
               "       coalesce(tdm.dt_integer,0)                                                                    ",
               "         - coalesce(tdm.dt_decimal,0)  dt_pdecimal                                                   ",
               "  from (                                                                                             ",
               "        select case tdx.gztd003                                                                      ",
               "                 when 'int2' then 'decimal'                                                          ", #20250113 --當成 decimal(4,0)
               "                 when 'int4' then 'decimal'                                                          ", #20250113 --當成 decimal(9,0)
               "                 when 'int8' then 'decimal'                                                          ", #20250113 --當成 decimal(19,0)
               "               else                                                                                  ",
               "                 tdx.gztd003                                                                         ",
               "               end                                      gztd003,                                     ",
               "               case tdx.gztd003                                                                      ",
               "                 when 'int2' then '4'                                                                ", #20250113 --int2 最大 32,767 未滿 99,999
               "                 when 'int4' then '9'                                                                ", #20250113 --int4 最大 2,147,483,647 未滿 9,999,999,999
               "                 when 'int8' then '19'                                                               ", #20250113 --int8 最大 9,223,372,036,854,775,807 未滿 99,999,999,999,999,999,999
               "               else                                                                                  ",
               "                 coalesce(tdx.gztd008,'0')                                                           ",
               "               end                                       gztd008,                                    ",
               "               case tdx.comma_pos                                                                    ",
               "                 when 0 then tdx.gztd008::integer                                                    ",
               "               else                                                                                  ",
               "                 substring(tdx.gztd008,1,                                                            ",
               "                           tdx.comma_pos-1)::integer                                                 ",
               "               end                                        dt_integer,                                ",
               "               case tdx.comma_pos                                                                    ",
               "                 when 0 then 0                                                                       ",
               "               else                                                                                  ",
               "                 substring(tdx.gztd008,                                                              ",
               "                           tdx.comma_pos+1,                                                          ",
               "                           length(tdx.gztd008))::integer                                             ",
               "               end                                        dt_decimal                                 ",
               "          from (                                                                                     ",
               "                select td.gztd003,                                                                   ",
               "                       coalesce(replace(replace(trim(td.gztd008),' ',''),'.',','),'0')   gztd008,    ",
               "                       coalesce(strpos(replace(replace(td.gztd008,' ',''),'.',','),','),0) comma_pos ", 
               "                  from gztd_t td                                                                     ",
               "                 where td.gztd001 = '",ls_column_type.TRIM(),"'                                      ",
               "               ) tdx                                                                                 ",
               "       ) tdm                                                                                         "
  &else
  LET ls_sql = "SELECT TDM.GZTD003                                     DT_TYPE,                                                 ",
               "       TDM.GZTD008                                     DT_LENGTH,                                               ",
               "       TDM.DT_INTEGER                                  DT_INTEGER,                                              ",
               "       TDM.DT_DECIMAL                                  DT_DECIMAL,                                              ",
               "       (NVL(TDM.DT_INTEGER,0) - NVL(TDM.DT_DECIMAL,0)) DT_PDECIMAL                                              ",
               "  FROM (                                                                                                        ",
               "        SELECT TDX.GZTD003                                                                         GZTD003,     ",
               "               TDX.GZTD008                                                                         GZTD008,     ",
               "               DECODE(TDX.COMMA_POS,0,TDX.GZTD008,SUBSTRB(TDX.GZTD008,1,TDX.COMMA_POS-1))          DT_INTEGER,  ",
               "               DECODE(TDX.COMMA_POS,0,0,SUBSTRB(TDX.GZTD008,TDX.COMMA_POS+1,LENGTHB(TDX.GZTD008))) DT_DECIMAL   ",
               "          FROM (                                                                                                ",
               "                SELECT TD.GZTD003                                                     GZTD003,                  ",
               "                       REPLACE(REPLACE(TD.GZTD008,' ',''),'.',',')                    GZTD008,                  ",
               "                       NVL(INSTRB(REPLACE(REPLACE(TD.GZTD008,' ',''),'.',','),','),0) COMMA_POS                 ",
               "                  FROM GZTD_T TD                                                                                ",
               "                 WHERE TD.GZTD001 = '",ls_column_type.TRIM(),"'                                                 ",
               "               ) TDX                                                                                            ",
               "       ) TDM                                                                                                    " 
  &endif 
  &endif 
  
  PREPARE lpre_get_data_type FROM ls_sql
  DECLARE lcur_get_data_type CURSOR FOR lpre_get_data_type
  OPEN lcur_get_data_type
  FETCH lcur_get_data_type INTO lo_data_type.*
  CLOSE lcur_get_data_type
  FREE lcur_get_data_type
  FREE lpre_get_data_type

  LET lo_return.* = lo_data_type.*
  
  RETURN lo_return.*
  
END FUNCTION

FUNCTION sadzi140_db_column_type_validation(p_new_column_type,p_old_column_type)
DEFINE
  p_new_column_type STRING,
  p_old_column_type STRING
DEFINE
  ls_new_column_type STRING,
  ls_old_column_type STRING,
  lo_new_data_type   T_DATA_TYPE,
  lo_old_data_type   T_DATA_TYPE
DEFINE
  lb_return BOOLEAN,
  ls_return STRING

  LET lb_return = TRUE  

  LET ls_new_column_type = p_new_column_type.trim()
  LET ls_old_column_type = p_old_column_type.trim()

  #先取得兩者的資料型態
  CALL sadzi140_db_get_column_type_data_type(ls_new_column_type) RETURNING lo_new_data_type.*
  CALL sadzi140_db_get_column_type_data_type(ls_old_column_type) RETURNING lo_old_data_type.*

  #新舊資料型態不符時, 則出局
  IF (lo_new_data_type.DT_TYPE <> lo_old_data_type.DT_TYPE) THEN
    LET lb_return = FALSE  
    LET ls_return = cs_valid_err_type_err
    GOTO _OUT
  END IF 

  #減去小數後的兩者整數比對, 新的比舊的小, 則出局
  IF (lo_new_data_type.DT_PINTEGER < lo_old_data_type.DT_PINTEGER) THEN
    LET lb_return = FALSE   
    LET ls_return = cs_valid_err_length_err
    GOTO _OUT
  END IF   

  #兩者小數比對, 新的比舊的小, 則出局
  IF (lo_new_data_type.DT_DECIMAL < lo_old_data_type.DT_DECIMAL) THEN
    LET lb_return = FALSE  
    LET ls_return = cs_valid_err_length_err
    GOTO _OUT
  END IF
  
  LABEL _OUT:

  RETURN lb_return,ls_return
 
END FUNCTION

FUNCTION sadzi140_db_get_key_null_default_value_list(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name   STRING,
  ls_sql          STRING,
  li_rec_count    INTEGER,
  ls_default_list STRING,
  ls_db_type      STRING, #20191209
  ls_default_sql  VARCHAR(1024),
  ls_default      STRING
DEFINE
  ls_return STRING
  
  LET ls_table_name = p_table_name.toLowerCase()
  
  LET li_rec_count = 0
  LET ls_default_list = ""
  LET ls_db_type = sadzi140_util_get_db_type() #20191209

  #抓DZIG_T的賦予值, 沒有的話抓預設賦予值
  &ifdef is_msv
  LET ls_sql = "select 'update '+eb.dzeb001+' set '+eb.dzeb002+' = '+ig.dzig003+' where '+eb.dzeb002+' is null' usql ", 
               "  from dzeb_t eb                                                                                     ",         
               " inner join dzig_t ig                                                                                ", 
               "         on ig.dzig001 = eb.dzeb001                                                                  ", 
               "        and ig.dzig002 = eb.dzeb002                                                                  ", 
               " where 1=1                                                                                           ",         
               "   and eb.dzeb001 = '",ls_table_name,"'                                                              ", 
               "   and eb.dzeb004 = 'Y'                                                                              ",         
               "union all                                                                                            ",         
               "select 'update '+eb.dzeb001+' set '+eb.dzeb002+' = '+ej.dzej005+' where '+eb.dzeb002+' is null' usql ", 
               "  from dzeb_t eb                                                                                     ",         
               "  inner join gztd_t td                                                                               ", 
               "          on td.gztd001 = eb.dzeb006                                                                 ", 
               "  left outer join dzej_t ej                                                                          ", 
               "               on ej.dzej001 = 'adzi140_parameters'                                                  ", 
               "              and ej.dzej003 = '",cs_param_level_patch_Key_value,"'                                  ", 
               "              and td.gztd003 = lower(ej.dzej004)                                                     ", 
               "              and coalesce(ej.dzej006,'",cs_db_driver_type_ora,"') = '",ls_db_type,"'                ", 
               " where 1=1                                                                                           ",         
               "   and eb.dzeb001 = '",ls_table_name,"'                                                              ", 
               "   and eb.dzeb004 = 'Y'                                                                              ",         
               "   and not exists (                                                                                  ",         
               "                    select 1                                                                         ",         
               "                      from dzig_t ig                                                                 ",         
               "                     where ig.dzig001 = eb.dzeb001                                                   ",         
               "                       and ig.dzig002 = eb.dzeb002                                                   ",         
               "                  )                                                                                  " 
  &else
  -- Oracle & PGS 共用
  LET ls_sql = "select 'update '||eb.dzeb001||' set '||eb.dzeb002||' = '||ig.dzig003||' where '||eb.dzeb002||' is null' usql     ",
               "  from dzeb_t eb                                                                                                 ",
               " inner join dzig_t ig on ig.dzig001 = eb.dzeb001                                                                 ",
               "                     and ig.dzig002 = eb.dzeb002                                                                 ",
               " where 1=1                                                                                                       ",
               "   and eb.dzeb001 = '",ls_table_name,"'                                                                          ",
               "   and eb.dzeb004 = 'Y'                                                                                          ",
               "union all                                                                                                        ",
               "select 'update '||eb.dzeb001||' set '||eb.dzeb002||' = '||td.default_val||' where '||eb.dzeb002||' is null' usql ",
               "  from dzeb_t eb                                                                                                 ",
               "  LEFT OUTER JOIN (                                                                                              ", --20230808 begin
               "                    SELECT DISTINCT gztd001,gztd003,                                                             ",
               "                           CASE                                                                                  ",
               "                             WHEN gztd003 LIKE '%char%' THEN ''' '''                                             ",
               "                             WHEN gztd003 = 'smallint' THEN '0'                                                  ",
               "                             WHEN gztd003 = 'integer' THEN '0'                                                   ",
               "                             WHEN gztd003 = 'bigint' THEN '0'                                                    ",
               "                             WHEN upper(gztd003) = 'INT2' THEN '0'                                               ", #20240529
               "                             WHEN upper(gztd003) = 'INT4' THEN '0'                                               ", #20240529
               "                             WHEN upper(gztd003) = 'INT8' THEN '0'                                               ", #20240529
               "                             WHEN gztd003 = 'number' THEN '0'                                                    ",
               "                             WHEN gztd003 = 'numeric' THEN '0'                                                   ", #20240529
               "                             WHEN gztd003 = 'decimal' THEN '0'                                                   ",
               "                             WHEN gztd003 LIKE 'int%' THEN '0'                                                   ",
               "                             WHEN gztd003 = 'date' THEN 'current_date'                                           ",
               "                             WHEN gztd003 = 'timestamp' THEN 'current_timestamp'                                 ",
               "                             WHEN gztd003 = 'bytea' THEN ''' '''                                                 ",
               "                             WHEN gztd003 = 'text' THEN ''' '''                                                  ",
               "                             WHEN gztd003 = 'blob' THEN ''' '''                                                  ",
               "                             WHEN gztd003 = 'clob' THEN ''' '''                                                  ",
               "                           END default_val                                                                       ",
               "                      FROM gztd_t td                                                                             ",
               "                     WHERE td.gztd003 IS NOT NULL                                                                ",
               "                  ) td                                                                                           ", --20230808 end
               "              ON td.gztd001 = eb.dzeb006                                                                         ",
               " where 1=1                                                                                                       ",
               "   and eb.dzeb001 = '",ls_table_name,"'                                                                          ",
               "   and eb.dzeb004 = 'Y'                                                                                          ",
               "   and not exists (                                                                                              ",
               "                    select 1                                                                                     ",
               "                      from dzig_t ig                                                                             ",
               "                     where ig.dzig001 = eb.dzeb001                                                               ",
               "                       and ig.dzig002 = eb.dzeb002                                                               ",
               "                  )                                                                                              "
  &endif                
               
  PREPARE lpre_get_key_null_default_value_list FROM ls_sql                                
  DECLARE lcur_get_key_null_default_value_list CURSOR FOR lpre_get_key_null_default_value_list
  
  OPEN lcur_get_key_null_default_value_list
  FOREACH lcur_get_key_null_default_value_list INTO ls_default_sql
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_count = li_rec_count + 1

    &ifdef is_msv
    LET ls_default_list = ls_default_list,ls_default_sql,"\n","go","\n"
    &else
    LET ls_default_list = ls_default_list,ls_default_sql,";","\n"
    &endif
    
  END FOREACH
  CLOSE lcur_get_key_null_default_value_list
  
  FREE lcur_get_key_null_default_value_list
  FREE lpre_get_key_null_default_value_list

  LET ls_return = ls_default_list
  
  RETURN ls_return
  
END FUNCTION

{
FUNCTION sadzi140_db_alter_optimizer_feature()
DEFINE
  ls_db_version STRING,
  ls_exec_sql   STRING

  CALL sadzi140_db_get_parameter(cs_param_level_session,cs_param_optimizer_features_enable) RETURNING ls_db_version
  LET ls_db_version = NVL(ls_db_version,cs_default_optimizer_features_enable)
  DISPLAY cs_information_tag,"Optimizer feature version set to : ",ls_db_version

  LET ls_exec_sql = "ALTER SESSION SET OPTIMIZER_FEATURES_ENABLE='",ls_db_version,"'"
                   
  CALL sadzi140_db_exec_SQL_no_commit(ls_exec_sql,FALSE)
  
END FUNCTION 
}

FUNCTION sadzi140_db_grant_object_privileges(p_object,p_privilege_info)
DEFINE
  p_object          STRING,
  p_privilege_info  T_PRIVILEGE_INFO
DEFINE
  ls_object         STRING,
  lo_privilege_info T_PRIVILEGE_INFO,
  ls_privileges     STRING,
  ls_sql            STRING,
  ls_db_type        STRING #20240521
DEFINE
  ls_return STRING

  LET ls_object           = p_object
  LET lo_privilege_info.* = p_privilege_info.*

  LET ls_privileges = ""

  #判斷有"Y"的就取用
  IF lo_privilege_info.pi_SELECT = "Y" THEN
    LET ls_privileges = ls_privileges,cs_privilege_select,"," 
  END IF
  IF lo_privilege_info.pi_INSERT = "Y" THEN
    LET ls_privileges = ls_privileges,cs_privilege_insert,"," 
  END IF
  IF lo_privilege_info.pi_UPDATE = "Y" THEN
    LET ls_privileges = ls_privileges,cs_privilege_update,"," 
  END IF
  IF lo_privilege_info.pi_DELETE = "Y" THEN
    LET ls_privileges = ls_privileges,cs_privilege_delete,"," 
  END IF
  IF lo_privilege_info.pi_REFERENCES = "Y" THEN
    LET ls_privileges = ls_privileges,cs_privilege_references,"," 
  END IF
  IF lo_privilege_info.pi_ALTER = "Y" THEN
    LET ls_privileges = ls_privileges,cs_privilege_alter,"," 
  END IF
  IF lo_privilege_info.pi_INDEX = "Y" THEN
    LET ls_privileges = ls_privileges,cs_privilege_index,"," 
  END IF

  #將最後一個逗號捨棄
  LET ls_privileges = ls_privileges.subString(1,ls_privileges.getLength()-1)

  &ifdef is_msv
  LET ls_sql = "grant ",ls_privileges," on ",ls_object," to ",lo_privilege_info.pi_ACCEPTER,"\n","go","\n"
  &else  
  &ifdef is_pgs
  LET ls_sql = "grant usage on schema ",lo_privilege_info.pi_GRANTER," to ",lo_privilege_info.pi_ACCEPTER,";","\n",
               "grant ",ls_privileges," on ",ls_object," to ",lo_privilege_info.pi_ACCEPTER,";","\n"
  &else 
  LET ls_db_type = sadzi140_util_get_db_type() #20240521
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb #20240521
      LET ls_sql = "grant usage on schema ",lo_privilege_info.pi_GRANTER," to ",lo_privilege_info.pi_ACCEPTER,";","\n",
                   "grant ",ls_privileges," on ",ls_object," to ",lo_privilege_info.pi_ACCEPTER,";","\n"
  --WHEN cs_db_driver_type_dmg #20240521
    --與Oracle共用
  OTHERWISE   
    LET ls_sql = "grant ",ls_privileges," on ",ls_object," to ",lo_privilege_info.pi_ACCEPTER,";","\n"
  END CASE  
  &endif
  &endif

  LET ls_return = ls_sql
  
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_revoke_object_privileges(p_object,p_privilege,p_user)
DEFINE
  p_object    STRING,
  p_privilege STRING,
  p_user      STRING
DEFINE
  ls_object    STRING,
  ls_privilege STRING,
  ls_user      STRING,
  ls_sql       STRING
DEFINE
  ls_return STRING
  
  LET ls_object    = p_object
  LET ls_privilege = NVL(p_privilege,cs_privilege_all)
  LET ls_user      = p_user

  &ifdef is_msv
  LET ls_sql = "revoke ",ls_privilege," on ",ls_object," from ",ls_user,"\n","go","\n"
  &else
  &ifdef is_pgs
  LET ls_sql = "revoke ",ls_privilege," on ",ls_object," from ",ls_user,";","\n"
  &else
  #Oracle與DM、KDB共用
  LET ls_sql = "revoke ",ls_privilege," on ",ls_object," from ",ls_user,";","\n"
  &endif
  &endif

  LET ls_return = ls_sql
  
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_get_privilege_granter(p_object)
DEFINE
  p_object STRING
DEFINE
  ls_object  STRING,
  li_rec_cnt INTEGER,
  ls_sql     STRING,
  lo_privilege_granter DYNAMIC ARRAY OF T_PRIVILEGE_GRANTER 
DEFINE
  lo_return DYNAMIC ARRAY OF T_PRIVILEGE_GRANTER 

  LET ls_object = p_object.toLowerCase()

  LET li_rec_cnt = 1

  &ifdef is_msv
  &else
  &ifdef is_pgs
  LET ls_sql = "select lower(dtp.owner) dzen002                                     ",
               "  from dba_tab_privs dtp                                            ",                                   
               " where dtp.table_name = '",ls_object.toLowerCase(),"'               ",          
               "   and dtp.grantee <> '",cs_default_privilege_role.toLowerCase(),"' ",                                   
               " group by dtp.owner,dtp.grantee                                     ",                                   
               "union                                                               ",
               "select en.dzen002                                                   ", 
               "  from dzen_t en                                                    ", 
               " where en.dzen001 = '",ls_object.toLowerCase(),"'                   ",
               " group by en.dzen002                                                ", 
               " order by 1                                                         "
  &else
  LET ls_sql = "select lower(dtp.owner) dzen002                                     ",
               "  from dba_tab_privs dtp                                            ",                                   
               " where dtp.table_name = '",ls_object.toUpperCase(),"'               ",          
               "   and dtp.grantee <> '",cs_default_privilege_role.toUpperCase(),"' ",                                   
               " group by dtp.owner,dtp.grantee                                     ",                                   
               "union                                                               ",
               "select en.dzen002                                                   ", 
               "  from dzen_t en                                                    ", 
               " where en.dzen001 = '",ls_object.toLowerCase(),"'                   ",
               " group by en.dzen002                                                ", 
               " order by 1                                                         "
  &endif                
  &endif                
 
  PREPARE lpre_get_privilege_granter FROM ls_sql
  DECLARE lcur_get_privilege_granter CURSOR FOR lpre_get_privilege_granter

  OPEN lcur_get_privilege_granter
  FOREACH lcur_get_privilege_granter INTO lo_privilege_granter[li_rec_cnt]
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF
    
    LET li_rec_cnt = li_rec_cnt + 1
    
  END FOREACH
  CLOSE lcur_get_privilege_granter
  
  CALL lo_privilege_granter.deleteElement(li_rec_cnt)
  
  FREE lpre_get_privilege_granter
  FREE lcur_get_privilege_granter    

  LET lo_return = lo_privilege_granter

  RETURN lo_return
  
END FUNCTION 

FUNCTION sadzi140_db_get_privilege_granter_from_diff(p_privilege_diff)
DEFINE
  p_privilege_diff  DYNAMIC ARRAY OF T_PRIVILEGE_DIFF
DEFINE
  lo_privilege_diff    DYNAMIC ARRAY OF T_PRIVILEGE_DIFF,
  lo_privilege_granter DYNAMIC ARRAY OF T_PRIVILEGE_GRANTER,
  li_diff              INTEGER,  
  li_loop              INTEGER,
  lb_exist             BOOLEAN, 
  ls_object            STRING,
  li_rec_cnt           INTEGER,
  ls_sql               STRING
DEFINE
  lo_return  DYNAMIC ARRAY OF T_PRIVILEGE_GRANTER 

  LET lo_privilege_diff = p_privilege_diff

  LET li_rec_cnt = 1
  CALL lo_privilege_granter.clear()

  #從差異清單中抓取不重複的賦予權限者
  FOR li_diff = 1 TO lo_privilege_diff.getLength()
    LET lb_exist = FALSE
    #找到有重複則離開
    FOR li_loop = 1 TO lo_privilege_granter.getLength()
      IF lo_privilege_granter[li_loop] = lo_privilege_diff[li_diff].DZEN002 THEN
        LET lb_exist = TRUE
        EXIT FOR
      END IF 
    END FOR
    #不重複則加入清單
    IF NOT lb_exist THEN
      LET lo_privilege_granter[lo_privilege_granter.getLength() + 1] = lo_privilege_diff[li_diff].DZEN002
    END IF
  END FOR 

  LET lo_return = lo_privilege_granter

  RETURN lo_return
  
END FUNCTION 

FUNCTION sadzi140_db_get_privilege_info(p_object,p_granter)
DEFINE
  p_object  STRING,
  p_granter STRING
DEFINE
  ls_object  STRING,
  ls_granter STRING,
  li_rec_cnt INTEGER,
  ls_sql     STRING,
  lo_privilege_info DYNAMIC ARRAY OF T_PRIVILEGE_INFO 
DEFINE
  lo_return DYNAMIC ARRAY OF T_PRIVILEGE_INFO 

  LET ls_object  = p_object.toLowerCase()
  LET ls_granter = p_granter

  LET li_rec_cnt = 1

  LET ls_sql = "select en.dzen002,en.dzen003,en.dzen004,en.dzen005,en.dzen006, ",
               "       en.dzen007,en.dzen008,en.dzen009,en.dzen010             ",
               "  from dzen_t en                                               ",
               " where en.dzen001 = '",ls_object,"'                            ",
               "   and en.dzen002 = '",ls_granter,"'                           ",
               " order by en.dzen003                                           "
 
  PREPARE lpre_get_privilege_info FROM ls_sql
  DECLARE lcur_get_privilege_info CURSOR FOR lpre_get_privilege_info

  OPEN lcur_get_privilege_info
  FOREACH lcur_get_privilege_info INTO lo_privilege_info[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF
    
    LET li_rec_cnt = li_rec_cnt + 1
    
  END FOREACH
  CLOSE lcur_get_privilege_info
  
  CALL lo_privilege_info.deleteElement(li_rec_cnt)
  
  FREE lpre_get_privilege_info
  FREE lcur_get_privilege_info    

  LET lo_return = lo_privilege_info

  RETURN lo_return
  
END FUNCTION 

FUNCTION sadzi140_db_refresh_privilege_info(p_PRIVILEGE_DIFF)
DEFINE
  p_PRIVILEGE_DIFF T_PRIVILEGE_DIFF
DEFINE
  lo_privilege_diff T_PRIVILEGE_DIFF,
  lo_privilege_info T_PRIVILEGE_INFO 
DEFINE
  lo_return T_PRIVILEGE_INFO 
  
  LET lo_PRIVILEGE_DIFF.* = p_PRIVILEGE_DIFF.*

  LET lo_privilege_info.pi_GRANTER    = lo_privilege_diff.DZEN002
  LET lo_privilege_info.pi_ACCEPTER   = lo_privilege_diff.DZEN003
  LET lo_privilege_info.pi_SELECT     = lo_privilege_diff.DZEN004
  LET lo_privilege_info.pi_INSERT     = lo_privilege_diff.DZEN005
  LET lo_privilege_info.pi_UPDATE     = lo_privilege_diff.DZEN006
  LET lo_privilege_info.pi_DELETE     = lo_privilege_diff.DZEN007
  LET lo_privilege_info.pi_REFERENCES = lo_privilege_diff.DZEN008
  LET lo_privilege_info.pi_ALTER      = lo_privilege_diff.DZEN009
  LET lo_privilege_info.pi_INDEX      = lo_privilege_diff.DZEN010
  
  LET lo_return.* = lo_privilege_info.*

  RETURN lo_return.*
  
END FUNCTION 

FUNCTION sadzi140_db_get_effect_table_list(p_table_name,p_lang)
DEFINE
  p_table_name  STRING,
  p_lang        STRING #20191209
DEFINE
  ls_table_name  STRING,
  ls_lang        STRING, #20191209
  li_rec_cnt INTEGER,
  ls_sql     STRING,
  lv_effect_table VARCHAR(1024), 
  ls_effect_table STRING 
DEFINE
  ls_return STRING

  LET ls_table_name = p_table_name.toLowerCase()
  LET ls_lang = p_lang #20191209

  LET li_rec_cnt = 1
  LET ls_effect_table = ""

  &ifdef is_msv
  LET ls_sql = "select distinct ed.dzed001+'-'+eal.dzeal003 effect_table   ",
               "  from dzed_t ed                                           ",
               "  left outer join dzeal_t eal                              ",
               "               on eal.dzeal001 = ed.dzed001                ",
               "              and eal.dzeal002 = '",ls_lang,"'             ",
               " where dzed005 = '",ls_table_name,"'                       ",
               " order by 1                                                "
  &else
  &ifdef is_pgs
  LET ls_sql = "select distinct ed.dzed001||'-'||eal.dzeal003 effect_table ",
               "  from dzed_t ed                                           ",
               "  left outer join dzeal_t eal                              ",
               "               on eal.dzeal001 = ed.dzed001                ",
               "              and eal.dzeal002 = '",ls_lang,"'             ",
               " where dzed005 = '",ls_table_name,"'                       ",
               " order by 1                                                "
  &else
  LET ls_sql = "select distinct ed.dzed001||'-'||eal.dzeal003 effect_table ",
               "  from dzed_t ed                                           ",
               "  left outer join dzeal_t eal                              ",
               "               on eal.dzeal001 = ed.dzed001                ",
               "              and eal.dzeal002 = '",ls_lang,"'             ",
               " where dzed005 = '",ls_table_name,"'                       ",
               " order by 1                                                "
  &endif               
  &endif               
  
  PREPARE lpre_get_effect_table_list FROM ls_sql
  DECLARE lcur_get_effect_table_list CURSOR FOR lpre_get_effect_table_list

  OPEN lcur_get_effect_table_list
  FOREACH lcur_get_effect_table_list INTO lv_effect_table
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF
    
    LET li_rec_cnt = li_rec_cnt + 1
    LET ls_effect_table = ls_effect_table,",",lv_effect_table,"\n"
    
  END FOREACH
  CLOSE lcur_get_effect_table_list
  
  FREE lpre_get_effect_table_list
  FREE lcur_get_effect_table_list    

  IF NVL(ls_effect_table.trim(),cs_null_value) <> cs_null_value THEN 
    LET ls_return = "\n","\n",ls_effect_table.subString(2,ls_effect_table.getLength())
  ELSE
    LET ls_return = ""
  END IF 

  RETURN ls_return
  
END FUNCTION 

FUNCTION sadzi140_db_get_db_schema_list()
DEFINE
  ls_sql         STRING,
  li_counts      INTEGER,
  lo_schema_list DYNAMIC ARRAY OF T_STATIC_DB_SCHEMA
DEFINE
  lo_return  DYNAMIC ARRAY OF T_STATIC_DB_SCHEMA

  LET li_counts = 1
  
  LET ls_sql = "select da.gzda001        ",
               "  from gzda_t da         ",
               " where 1=1               ",
               "   and da.gzda005  = 'Y' ", 
               "   and da.gzdastus = 'Y' ",
               " order by da.gzda001     "
               
  PREPARE lpre_get_db_schema_list FROM ls_sql
  DECLARE lcur_get_db_schema_list CURSOR FOR lpre_get_db_schema_list

  OPEN lcur_get_db_schema_list
  FOREACH lcur_get_db_schema_list INTO lo_schema_list[li_counts].sds_DB_NAME
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_counts = li_counts + 1

  END FOREACH
  CLOSE lcur_get_db_schema_list

  CALL lo_schema_list.deleteElement(li_counts)
  
  FREE lpre_get_db_schema_list
  FREE lcur_get_db_schema_list

  LET lo_return = lo_schema_list
  
  RETURN lo_return

END FUNCTION

FUNCTION sadzi140_db_get_static_db_list()
DEFINE
  li_length           INTEGER,
  ls_static_db        STRING,
  ls_static_db_char   STRING,
  ls_static_db_string STRING,
  lo_static_db_list   DYNAMIC ARRAY OF T_STATIC_DB_SCHEMA
DEFINE
  lo_return DYNAMIC ARRAY OF T_STATIC_DB_SCHEMA

  CALL lo_static_db_list.clear()

  #取得當地語系清單
  CALL sadzi140_db_get_parameter(cs_param_level_system,cs_param_essential_schemas) RETURNING ls_static_db
  
  IF ls_static_db.trim() IS NOT NULL THEN 
    FOR li_length = 1 TO ls_static_db.getLength()
      LET ls_static_db_char = ls_static_db.subString(li_length,li_length)
      IF ls_static_db_char = "," THEN 
        LET lo_static_db_list[lo_static_db_list.getLength() + 1].sds_DB_NAME = ls_static_db_string
        LET ls_static_db_string = ""
      ELSE
        LET ls_static_db_string = ls_static_db_string,ls_static_db_char 
      END IF
    END FOR
  ELSE
    #抓不到給預設值
    LET lo_static_db_list[1].sds_DB_NAME = 'ds'
    LET lo_static_db_list[2].sds_DB_NAME = 'dsdemo'
    LET lo_static_db_list[3].sds_DB_NAME = 'dsdata'
    LET lo_static_db_list[4].sds_DB_NAME = 'dsstd'
  END IF

  LET lo_return = lo_static_db_list

  RETURN lo_return  
    
END FUNCTION 

FUNCTION sadzi140_db_check_complete_of_db_data()
DEFINE
  lo_static_db_list DYNAMIC ARRAY OF T_STATIC_DB_SCHEMA,
  lo_db_data_list   DYNAMIC ARRAY OF T_STATIC_DB_SCHEMA,
  li_static         INTEGER,
  li_db_data        INTEGER,
  lb_total_result   BOOLEAN,
  ls_total_result   STRING,
  lb_result         BOOLEAN
DEFINE
  ls_return  STRING,
  lb_return  BOOLEAN

  CALL sadzi140_db_get_static_db_list() RETURNING lo_static_db_list
  CALL sadzi140_db_get_db_schema_list() RETURNING lo_db_data_list

  LET lb_total_result = TRUE
  LET ls_total_result = ""
  
  #抓出在資料庫設定中是否 ds,dsdemo,dsdata 有設定
  FOR li_static = 1 TO lo_static_db_list.getLength()
    LET lb_result = FALSE
    FOR li_db_data = 1 TO lo_db_data_list.getLength()
      IF lo_static_db_list[li_static].sds_DB_NAME = lo_db_data_list[li_db_data].sds_DB_NAME THEN 
        LET lb_result = TRUE
        EXIT FOR
      END IF
    END FOR
    IF NOT lb_result THEN
      LET lb_total_result = FALSE
      LET ls_total_result = ls_total_result,lo_static_db_list[li_static].sds_DB_NAME,","
    END IF
  END FOR 
  
  LET lb_return = lb_total_result
  LET ls_return = "\n",ls_total_result.subString(1,ls_total_result.getLength()-1)
  
  RETURN lb_return,ls_return 
  
END FUNCTION

#20160616 begin
FUNCTION sadzi140_db_check_index_length_if_exceed(p_DZEB_T,p_column_list)
DEFINE
  p_DZEB_T      DYNAMIC ARRAY OF T_DZEB_T,
  p_column_list DYNAMIC ARRAY OF T_COLUMN_LIST
DEFINE
  lo_DZEB_T           DYNAMIC ARRAY OF T_DZEB_T,
  lo_column_list      DYNAMIC ARRAY OF T_COLUMN_LIST,
  li_index_max_length INTEGER,
  li_loop             INTEGER,
  li_loop2            INTEGER,
  li_type_size        INTEGER,
  li_all_type_size    INTEGER
DEFINE
  lo_return  T_KEY_IF_EXCEED

  LET lo_DZEB_T = p_DZEB_T
  LET lo_column_list = p_column_list

  LET li_type_size = 0
  LET li_all_type_size = 0
  
  CALL sadzi140_db_get_database_key_max_length() RETURNING li_index_max_length

  FOR li_loop = 1 TO lo_column_list.getLength()
    FOR li_loop2 = 1 TO lo_DZEB_T.getLength()
      IF lo_DZEB_T[li_loop2].dzeb002 = lo_column_list[li_loop].COLUMN_NAME THEN 
        CALL sadzi140_db_get_database_column_data_type_size(lo_DZEB_T[li_loop2].*) RETURNING li_type_size
        LET li_all_type_size = li_all_type_size + li_type_size
        EXIT FOR
      END IF 
    END FOR
  END FOR  

  LET lo_return.kie_sys_def = li_index_max_length
  LET lo_return.kie_user_def = li_all_type_size
  
  IF li_all_type_size >= li_index_max_length THEN
    LET lo_return.kie_result = TRUE
  ELSE 
    LET lo_return.kie_result = FALSE
  END IF 
  
  RETURN lo_return.*
  
END FUNCTION 

FUNCTION sadzi140_db_check_key_length_if_exceed(p_DZEB_T)
DEFINE
  p_DZEB_T  DYNAMIC ARRAY OF T_DZEB_T
DEFINE
  lo_DZEB_T         DYNAMIC ARRAY OF T_DZEB_T,
  li_key_max_length INTEGER,
  li_loop           INTEGER,
  li_type_size      INTEGER,
  li_all_type_size  INTEGER
DEFINE
  lo_return  T_KEY_IF_EXCEED

  LET lo_DZEB_T = p_DZEB_T

  LET li_type_size = 0
  LET li_all_type_size = 0
  
  CALL sadzi140_db_get_database_key_max_length() RETURNING li_key_max_length

  FOR li_loop = 1 TO lo_DZEB_T.getLength()
    IF NVL(lo_DZEB_T[li_loop].dzeb004,"N") = "Y" THEN 
      CALL sadzi140_db_get_database_column_data_type_size(lo_DZEB_T[li_loop].*) RETURNING li_type_size
      LET li_all_type_size = li_all_type_size + li_type_size
    END IF 
  END FOR 

  LET lo_return.kie_sys_def = li_key_max_length
  LET lo_return.kie_user_def = li_all_type_size
  
  IF li_all_type_size >= li_key_max_length THEN
    LET lo_return.kie_result = TRUE
  ELSE 
    LET lo_return.kie_result = FALSE
  END IF 
  
  RETURN lo_return.*
  
END FUNCTION 

FUNCTION sadzi140_db_get_database_key_max_length()
DEFINE
  ls_sql            STRING,
  li_block_size     INTEGER,
  ls_db_type        STRING, #20240521
  li_key_max_length INTEGER
DEFINE  
  li_return  INTEGER

  &ifdef is_msv
  LET li_key_max_length = 900
  &else
  &ifdef is_pgs
  #https://wiki.postgresql.org/wiki/FAQ#What_is_the_maximum_size_for_a_row.2C_a_table.2C_and_a_database.3F
  LET li_key_max_length = 8192
  &else
  #取得DB Block Size
  LET ls_db_type = sadzi140_util_get_db_type()
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb --20240521
      LET ls_sql = "select value block_size    ", --8192
                   "  from v$parameter         ",
                   " where name = 'block_size' "
    WHEN cs_db_driver_type_dmg --20230804
      LET ls_sql = "select sf_get_page_size() " --32768
  OTHERWISE  
    LET ls_sql = "select value block_size       ", --8192
                 "  from v$parameter            ",
                 " where name = 'db_block_size' "
  END CASE
                   
  PREPARE lpre_get_oracle_key_max_length FROM ls_sql
  DECLARE lcur_get_oracle_key_max_length CURSOR FOR lpre_get_oracle_key_max_length
  OPEN lcur_get_oracle_key_max_length
  FETCH lcur_get_oracle_key_max_length INTO li_block_size
  CLOSE lcur_get_oracle_key_max_length
  FREE lcur_get_oracle_key_max_length
  FREE lpre_get_oracle_key_max_length
    
  IF sadzi140_util_get_db_type() = cs_db_driver_type_ora THEN --20230804 begin
    #Formula reference to http://blog.csdn.net/yidian815/article/details/16336911
    LET li_key_max_length = (NVL(li_block_size,0) - 192) * 0.8
  ELSE  
    LET li_key_max_length = NVL(li_block_size,0)
  END IF  --20230804 end
  &endif
  &endif
  
  LET li_return = li_key_max_length
  
  RETURN li_return
  
END FUNCTION

FUNCTION sadzi140_db_get_database_column_data_type_size(p_DZEB_T)
DEFINE
  p_DZEB_T  T_DZEB_T
DEFINE
  lo_DZEB_T      T_DZEB_T,
  ls_data_type   STRING,
  ls_temp_size   STRING, #20191209
  li_temp_size   STRING, #20191209
  li_type_size   INTEGER,
  li_define_size INTEGER
DEFINE
  li_return INTEGER
  
  LET lo_DZEB_T.* = p_DZEB_T.*
  LET ls_data_type = UPSHIFT(lo_DZEB_T.dzeb007)

  LET li_type_size = 0
  
  #Size reference to https://docs.oracle.com/cd/B28359_01/appdev.111/b28395/oci03typ.htm
  CASE
    WHEN ls_data_type = "BLOB"
      LET li_type_size = 99999
    WHEN ls_data_type = "CLOB"
      LET li_type_size = 99999
    WHEN ls_data_type = "DATE"
      &ifdef is_msv
      LET li_type_size = 3
      &else
      &ifdef is_pgs
      LET li_type_size = 4
      &else
      LET li_type_size = 7
      &endif
      &endif
    WHEN ls_data_type = "NUMBER"
      LET li_type_size = 21 
    WHEN ls_data_type = "NUMERIC"  #20240529
      LET li_type_size = 38 
    WHEN ls_data_type = "TIMESTAMP"
      &ifdef is_pgs
      LET li_type_size = 8
      &else
      LET li_define_size = NVL(lo_DZEB_T.dzeb008,0)
      LET li_type_size = IIF(li_define_size > 0,13,11)
      &endif
    WHEN ls_data_type = "VARCHAR2"
      LET li_type_size = NVL(lo_DZEB_T.dzeb008,0)
    #20191209 begin
    #SELECT * FROM sys.types;
    WHEN ls_data_type = "VARBINARY"
      LET li_type_size = 8000
    WHEN ls_data_type = "NVARCHAR"
      IF UPSHIFT(lo_DZEB_T.dzeb008 ) = 'MAX' THEN
        LET li_type_size = 8000
      ELSE 
        LET li_type_size = NVL(lo_DZEB_T.dzeb008,0)
      END IF
    WHEN ls_data_type = "DECIMAL"
      LET ls_temp_size = lo_DZEB_T.dzeb008
      IF ls_temp_size.getIndexOf(',',1) > 0 THEN
        LET ls_temp_size = ls_temp_size.subString(1,ls_temp_size.getIndexOf(',',1)-1)
      END IF      
      LET li_temp_size = ls_temp_size
      LET li_type_size = li_temp_size #initial define
      &ifdef is_msv
      #http://www.sqlservertutorial.net/sql-server-basics/sql-server-decimal/
      CASE 
        WHEN li_temp_size <= 9 
          LET li_type_size = 5
        WHEN li_temp_size > 9 AND li_temp_size <= 19 
          LET li_type_size = 9
        WHEN li_temp_size > 19 AND li_temp_size <= 28 
          LET li_type_size = 13
        WHEN li_temp_size > 28 AND li_temp_size <= 38 
          LET li_type_size = 17
      END CASE
      &endif
      &ifdef is_pgs
      #https://www.postgresql.org/docs/8.1/datatype.html
      LET li_type_size = li_temp_size
      &endif
    WHEN ls_data_type = "DATETIME2"
      LET li_type_size = 8
    #20191209 end  
    #20200330 begin
    WHEN ls_data_type = "INT2"
      LET li_type_size = 2
    WHEN ls_data_type = "INT4"
      LET li_type_size = 4
    WHEN ls_data_type = "INT8"
      LET li_type_size = 8
    WHEN ls_data_type = "BYTEA"
      LET li_type_size = 99999
    WHEN ls_data_type = "TEXT"
      LET li_type_size = 99999
    WHEN ls_data_type = "VARCHAR"
      LET li_type_size = NVL(lo_DZEB_T.dzeb008,0)
    #20200330 ebd
    #20230804 begin
    WHEN ls_data_type = "SMALLINT"
      LET li_type_size = 2
    WHEN ls_data_type = "INTEGER"
      LET li_type_size = 4
    WHEN ls_data_type = "BIGINT"
      LET li_type_size = 8
    #20230804 end
  OTHERWISE 
    LET li_type_size = 99999
  END CASE  
  
  LET li_return = li_type_size
  
  RETURN li_return
    
END FUNCTION
#20160616 end

#20161220 begin
FUNCTION sadzi140_db_get_comment_columns_scripts(p_column_comments)
DEFINE
  p_column_comments T_COMMENT_COLUMNS
DEFINE
  lo_column_comments T_COMMENT_COLUMNS,
  ls_scripts  STRING,
  ls_db_type  STRING #20240521
DEFINE
  ls_return STRING

  LET lo_column_comments.* = p_column_comments.*

  &ifdef is_msv
  LET ls_scripts = "declare @user sysname                                                                                                                            ","\n",
                   "select @user = user_name()                                                                                                                       ","\n",
                   "begin try                                                                                                                                        ","\n",
                   "   execute sp_addextendedproperty 'comment','",DOWNSHIFT(lo_column_comments.CC_COLUMN_NAME),"',                                                  ","\n",
                   "           'user', @user, 'table', '",DOWNSHIFT(lo_column_comments.CC_TABLE_NAME),"','column','",DOWNSHIFT(lo_column_comments.CC_COLUMN_NAME),"' ","\n",
                   "end try                                                                                                                                          ","\n",
                   "begin catch                                                                                                                                      ","\n",
                   "  execute sp_updateextendedproperty 'comment','",DOWNSHIFT(lo_column_comments.CC_COLUMN_NAME),"',                                                ","\n",
                   "          'user', @user, 'table', '",DOWNSHIFT(lo_column_comments.CC_TABLE_NAME),"','column','",DOWNSHIFT(lo_column_comments.CC_COLUMN_NAME),"'  ","\n",
                   "end catch                                                                                                                                        ","\n",
                   "go","\n"
  &else
  &ifdef is_pgs
  LET ls_scripts = "do $$                                                    ","\n",
                   "declare                                                  ","\n",        
                   "  ls_exe_sql varchar(1000);                              ","\n",       
                   "begin                                                    ","\n",        
                   "                                                         ","\n",        
                   "  ls_exe_sql := 'comment on column ",
                                    lo_column_comments.CC_OWNER,".",
                                    lo_column_comments.CC_TABLE_NAME,".",
                                    lo_column_comments.CC_COLUMN_NAME,
                                    " is ''",
                                    lo_column_comments.CC_COLUMN_NAME,"''';  ","\n",
                   "                                                         ","\n",        
                   "  execute ls_exe_sql;                                    ","\n",
                   "                                                         ","\n",        
                   "end $$;                                                  ","\n"  
  &else
  LET ls_db_type = sadzi140_util_get_db_type() #20240521
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb #20240521
      LET ls_scripts = "do $$                                                    ","\n",
                       "declare                                                  ","\n",        
                       "  ls_exe_sql varchar(1000);                              ","\n",       
                       "begin                                                    ","\n",        
                       "                                                         ","\n",        
                       "  ls_exe_sql := 'comment on column ",
                                        lo_column_comments.CC_OWNER,".",
                                        lo_column_comments.CC_TABLE_NAME,".",
                                        lo_column_comments.CC_COLUMN_NAME,
                                        " is ''",
                                        lo_column_comments.CC_COLUMN_NAME,"''';  ","\n",
                       "                                                         ","\n",        
                       "  execute ls_exe_sql;                                    ","\n",
                       "                                                         ","\n",        
                       "end $$;                                                  ","\n"  
  --WHEN cs_db_driver_type_dmg #20240521
    --與Oracle共用
  OTHERWISE   
    LET ls_scripts = "DECLARE                                                                      ","\n",
                     "  ls_exe_sql VARCHAR2(1000);                                                 ","\n",
                     "BEGIN                                                                        ","\n",
                     "                                                                             ","\n",
                     "  ls_exe_sql := 'COMMENT ON COLUMN ",
                                       UPSHIFT(lo_column_comments.CC_OWNER),".",
                                       UPSHIFT(lo_column_comments.CC_TABLE_NAME),".",
                                       UPSHIFT(lo_column_comments.CC_COLUMN_NAME),
                                       " IS ''",
                                       UPSHIFT(lo_column_comments.CC_COLUMN_NAME),"'''; ","\n",
                     "                                                                             ","\n",
                     "  BEGIN                                                                      ","\n",
                     "    EXECUTE IMMEDIATE ls_exe_sql;                                            ","\n",
                     "    DBMS_OUTPUT.PUT_LINE('",cs_success_tag,"'||ls_exe_sql||', '||SQLERRM);   ","\n",
                     "  EXCEPTION                                                                  ","\n",
                     "    WHEN OTHERS THEN                                                         ","\n",
                     "      DBMS_OUTPUT.PUT_LINE('",cs_warning_tag,"'||ls_exe_sql||', '||SQLERRM); ","\n",
                     "  END;                                                                       ","\n",
                     "                                                                             ","\n",
                     "END;                                                                         ","\n",
                     "/                                                                            ","\n"
  END CASE  
  &endif                 
  &endif                 

  LET ls_return = ls_scripts

  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_db_get_toggle_comment_columns(p_db_user_name,p_table_name)
DEFINE
  p_db_user_name STRING,
  p_table_name   STRING 
DEFINE
  ls_db_user_name STRING,
  ls_table_name   STRING,
  li_rec_cnt      INTEGER,
  ls_sql          STRING,
  lo_comment_columns  DYNAMIC ARRAY OF T_COMMENT_COLUMNS
DEFINE
  lo_return  DYNAMIC ARRAY OF T_COMMENT_COLUMNS

  LET ls_db_user_name = p_db_user_name
  LET ls_table_name   = p_table_name

  LET li_rec_cnt = 1

  &ifdef is_msv
  LET ls_sql = "select '",ls_db_user_name,"'  as owner,                                               ",
               "       eb.dzeb001             as table_name,                                          ",
               "       eb.dzeb002             as column_name,                                         ",
               "       eb.dzeb002             as comments                                             ",
               "  from dzeb_t eb                                                                      ",      
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'                                 ",
               "   and exists (                                                                       ",      
               "                select 1                                                              ",      
               "                  from dba_tab_columns dtcs                                           ",
               "                 where 1=1                                                            ",      
               "                   and dtcs.owner       = '",ls_db_user_name.toLowerCase(),"'         ",
               "                   and dtcs.table_name  = '",ls_table_name.toLowerCase(),"'           ",
               "                   and dtcs.column_name = eb.dzeb002                                  ",
               "                   and dtcs.column_name <> coalesce(dtcs.comment,'",cs_null_value,"') ",
               "              )                                                                       "  
  &else
  &ifdef is_pgs
  LET ls_sql = "select '",ls_db_user_name.toLowerCase(),"'  as owner,                                  ",  
               "       eb.dzeb001  as table_name,                                                      ",  
               "       eb.dzeb002  as column_name,                                                     ",  
               "       eb.dzeb002  as comments                                                         ",  
               "  from dzeb_t eb                                                                       ",         
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'                                  ", 
               "   and exists (                                                                        ",         
               "                select 1                                                               ",         
               "                  from dba_col_comments dccs                                           ",         
               "                 where 1=1                                                             ",         
               "                   and dccs.owner       = '",ls_db_user_name.toLowerCase(),"'          ", 
               "                   and dccs.table_name  = '",ls_table_name.toLowerCase(),"'            ", 
               "                   and dccs.column_name = eb.dzeb002                                   ",  
               "                   and dccs.column_name <> coalesce(dccs.comments,'",cs_null_value,"') ", 
               "              )                                                                        ",         
               " order by eb.dzeb021                                                                   "         
  &else
  LET ls_sql = "select upper('",ls_db_user_name,"') as owner,                                     ",
               "       upper(eb.dzeb001)            as table_name,                                ",
               "       upper(eb.dzeb002)            as column_name,                               ",
               "       upper(eb.dzeb002)            as comments                                   ",
               "  from dzeb_t eb                                                                  ",
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'                             ",
               "   and exists (                                                                   ", 
               "                select 1                                                          ",
               "                  from dba_col_comments dccs                                      ",
               "                 where 1=1                                                        ",
               "                   and dccs.owner       = upper('",ls_db_user_name,"')            ", 
               "                   and dccs.table_name  = upper('",ls_table_name,"')              ", 
               "                   and dccs.column_name = upper(eb.dzeb002)                       ",
               "                   and dccs.column_name <> nvl(dccs.comments,'",cs_null_value,"') ",
               "              )                                                                   ",
               " order by eb.dzeb021                                                              " 
  &endif 
  &endif 
  
  PREPARE lpre_get_toggle_comment_columns FROM ls_sql
  DECLARE lcur_get_toggle_comment_columns CURSOR FOR lpre_get_toggle_comment_columns

  OPEN lcur_get_toggle_comment_columns
  FOREACH lcur_get_toggle_comment_columns INTO lo_comment_columns[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_toggle_comment_columns

  CALL lo_comment_columns.deleteElement(li_rec_cnt)
  
  FREE lpre_get_toggle_comment_columns
  FREE lcur_get_toggle_comment_columns  
  
  LET lo_return = lo_comment_columns

  RETURN lo_return 
  
END FUNCTION

FUNCTION sadzi140_db_get_backdoor_columns(p_table_name)
DEFINE
  p_table_name  STRING 
DEFINE
  ls_table_name       STRING,
  li_rec_cnt          INTEGER,
  ls_sql              STRING,
  li_col_count        INTEGER,
  ls_zone             STRING, #20191209 
  ls_backdoor_cols    STRING,
  lo_backdoor_columns DYNAMIC ARRAY OF T_COMMENT_COLUMNS
DEFINE
  ls_return  STRING,
  li_return  INTEGER

  LET ls_table_name   = p_table_name

  LET li_rec_cnt = 1
  LET li_col_count = 1
  LET ls_backdoor_cols = ""
  LET ls_zone = DOWNSHIFT(FGL_GETENV("ZONE")),"_" #20191209

  &ifdef is_msv
  LET ls_sql = "select dtcs.owner,dtcs.table_name,dtcs.column_name,dtcs.comment ",
               "  from dba_tab_columns dtcs                                     ",
               " where 1=1                                                      ",
               "   and dtcs.table_name = '",ls_table_name.toLowerCase(),"'      ",
               "   and coalesce(dtcs.comment, 'x') <> dtcs.column_name          ",
               "   and exists (                                                 ",
               "                select *                                        ",
               "                  from gzda_t da                                ",
               "                 where '",ls_zone,"'+da.gzda001 = dtcs.owner    ",
               "                   and da.gzda005 = 'Y'                         ",
               "              )                                                 ",
               " order by dtcs.owner,dtcs.table_name,dtcs.column_name           "
  &else
  &ifdef is_pgs
  CASE sadzi140_util_get_pgs_branch_type()
    WHEN cs_pgs_sub_mogdb
      -- 20241014 begin 
      LET ls_sql = "with                                                                             ", 
                   "  dtcs as materialized (select *                                                 ",
                   "                          from dba_tab_columns                                   ",
                   "                         where table_name = '",ls_table_name.toLowerCase(), "'), ",
                   "  dct as materialized (select *                                                  ",
                   "                         from dba_col_comments                                   ",
                   "                        where table_name = '",ls_table_name.toLowerCase(), "'),  ",
                   "  dos as materialized (select *                                                  ",
                   "                         from dba_objects                                        ",
                   "                        where object_name = '",ls_table_name.toLowerCase(), "'   ",
                   "                          and object_type = 'VIEW')                              ",
                   "select dtcs.owner, dtcs.table_name, dtcs.column_name, dct.comments               ",
                   "  from dtcs                                                                      ",
                   "  left outer join dct                                                            ",
                   "    on dct.owner       = dtcs.owner                                              ",
                   "   and dct.table_name  = dtcs.table_name                                         ",
                   "   and dct.column_name = dtcs.column_name                                        ",
                   " where 1 = 1                                                                     ",
                   "   and coalesce(dct.comments, 'x') <> dtcs.column_name                           ",
                   "   and exists (select 1                                                          ",
                   "                 from gzda_t da                                                  ",
                   "                where da.gzda001 = dtcs.owner                                    ",
                   "                  and da.gzda005 = 'Y')                                          ",
                   "   and not exists (select 1                                                      ",
                   "                     from dos                                                    ",
                   "                    where dos.owner       = dtcs.owner                           ",
                   "                      and dos.object_name = dtcs.table_name                      ",
                   "                      and dos.object_type = 'VIEW')                              ",
                   " order by dtcs.owner, dtcs.table_name, dtcs.column_name                          "
      -- 20241014 end
    WHEN cs_pgs_sub_polardb
      {待建構}
  OTHERWISE       
    --PostgreSQL
    LET ls_sql = "select dtcs.owner,dtcs.table_name,dtcs.column_name,dct.comments                ",
                 "  from dba_tab_columns dtcs                                                    ",
                 "  left outer join (                                                            ",
                 "                    select *                                                   ",
                 "                      from dba_col_comments act                                ",
                 "                     where act.table_name = '",ls_table_name.toLowerCase(), "' ",
                 "                   ) dct                                                       ",
                 "               on dct.owner       = dtcs.owner                                 ",        
                 "              and dct.table_name  = dtcs.table_name                            ",
                 "              and dct.column_name = dtcs.column_name                           ",
                 " where 1=1                                                                     ",
                 "   and dtcs.table_name = '",ls_table_name.toLowerCase(),"'                     ",
                 "   and coalesce(dct.comments, 'x') <> dtcs.column_name                         ",
                 "   and exists (                                                                ",
                 "                select *                                                       ",
                 "                  from gzda_t da                                               ",
                 "                 where da.gzda001 = dtcs.owner                                 ",
                 "                   and da.gzda005 = 'Y'                                        ",
                 "              )                                                                ",
                 "   and not exists (                                                            ", 
                 "                    select *                                                   ", 
                 "                      from dba_objects dos                                     ", 
                 "                     where dos.owner = dtcs.owner                              ", 
                 "                       and dos.object_name = dtcs.table_name                   ", 
                 "                       and dos.object_type = 'VIEW'                            ", 
                 "                  )                                                            ", 
                 " order by dtcs.owner,dtcs.table_name,dtcs.column_name                          " 
  END CASE 
  &else
  LET ls_sql = "select dccs.owner,dccs.table_name,dccs.column_name,dccs.comments ",
               "  from dba_col_comments dccs                                     ",
               " where 1=1                                                       ",
               "   and dccs.table_name = '",ls_table_name.toUpperCase(),"'       ",
               "   and coalesce(trim(dccs.comments), 'x') <> dccs.column_name    ",
               "   and exists (                                                  ",
               "                select 1                                         ",
               "                  from gzda_t da                                 ",
               "                 where da.gzda001 = lower(dccs.owner)            ",
               "                   and da.gzda005 = 'Y'                          ",
               "              )                                                  ",
               " order by dccs.owner,dccs.table_name,dccs.column_name            "
  &endif
  &endif
  
  PREPARE lpre_get_backdoor_columns FROM ls_sql
  DECLARE lcur_get_backdoor_columns CURSOR FOR lpre_get_backdoor_columns

  OPEN lcur_get_backdoor_columns
  FOREACH lcur_get_backdoor_columns INTO lo_backdoor_columns[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET ls_backdoor_cols = ls_backdoor_cols,
                           lo_backdoor_columns[li_rec_cnt].CC_OWNER,".",
                           lo_backdoor_columns[li_rec_cnt].CC_TABLE_NAME,".",
                           lo_backdoor_columns[li_rec_cnt].CC_COLUMN_NAME,","

    LET li_rec_cnt = li_rec_cnt + 1
    LET li_col_count = li_col_count + 1
    IF li_col_count > 5 THEN
      LET li_col_count = 1 
      LET ls_backdoor_cols = ls_backdoor_cols,"\n"
    END IF  

  END FOREACH
  CLOSE lcur_get_backdoor_columns

  CALL lo_backdoor_columns.deleteElement(li_rec_cnt)
  
  FREE lpre_get_backdoor_columns
  FREE lcur_get_backdoor_columns  
  
  LET li_return = lo_backdoor_columns.getLength()
  LET ls_return = ls_backdoor_cols

  RETURN li_return,ls_return 
  
END FUNCTION
#20161220 end

#20170912 begin
FUNCTION sadzi140_db_get_resort_index_columns_list(p_resort_cols,p_type)
DEFINE
  p_resort_cols T_RESORT_COLUMN,
  p_type        STRING
DEFINE
  lo_resort_cols T_RESORT_COLUMN,
  ls_type        STRING,
  ls_sql         STRING,
  ls_sc_sql      STRING,
  ls_db_type     STRING, #20240521
  ls_resort_list VARCHAR(4096)
DEFINE
  ls_return  STRING  

  LET lo_resort_cols.* = p_resort_cols.*
  LET ls_type          = p_type

  CASE 
    WHEN ls_type = cs_resort_index_type_text
      LET ls_sc_sql = "select '",lo_resort_cols.rc_COLUMN_LIST,"' sep_col from dual"
    WHEN ls_type = cs_resort_index_type_table
      LET ls_sc_sql = "select dzec004 sep_col from dzec_t where dzec001 = '",lo_resort_cols.rc_TABLE_NAME,"' and dzec002 = '",lo_resort_cols.rc_INDEX_NAME,"'"
  END CASE

  &ifdef is_msv
  LET ls_sql = "select cast(replace(scl.rearrange_col, ' ', '') as nvarchar) as sc_list ",
               "  from (                                                                ",
               "        select rearrange_col = stuff(fco.rearrange_col, 1, 2, '')       ",
               "          from (                                                        ",
               ls_sc_sql,
               "               ) sc                                                     ",
               "         cross apply (                                                  ",
               "                       select ', ' + ltrim(ss.value)                    ",
               "                         from string_split(sc.sep_col, ',') ss          ",
               "                        order by ltrim(ss.value)                        ",
               "                          for xml path('')                              ",
               "                     ) fco(rearrange_col)                               ",
               "         group by fco.rearrange_col                                     ",
               "       ) as scl                                                         "
  &else
  &ifdef is_pgs
  LET ls_sql = "select string_agg(sc.dzec004_sep,',' order by sc.dzec004_sep) as sc_list       ",              
               "  from (                                                                       ",                          
               "         with temp as                                                          ",                          
               "         (                                                                     ",                          
               ls_sc_sql,
               "         )                                                                     ",                          
               "         select unnest(string_to_array(t.sep_col::varchar,',')) as dzec004_sep ",
               "           from temp t                                                         ",                         
               "       ) sc                                                                    "
  &else
  LET ls_db_type = sadzi140_util_get_db_type()
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb
      LET ls_sql = "select string_agg(sc.dzec004_sep,',' order by sc.dzec004_sep) as sc_list       ",              
                   "  from (                                                                       ",                          
                   "         with temp as                                                          ",                          
                   "         (                                                                     ",                          
                   ls_sc_sql,
                   "         )                                                                     ",                          
                   "         select unnest(string_to_array(t.sep_col::varchar,',')) as dzec004_sep ",
                   "           from temp t                                                         ",                         
                   "       ) sc                                                                    "
    WHEN cs_db_driver_type_dmg
      LET ls_sql = "with temp as                                                                          ",           
                   " (                                                                                    ",           
                   ls_sc_sql,
                   " )                                                                                    ",                   
                   " select listagg(sc.dzec004_sep,',') within group (order by sc.dzec004_sep) as sc_list ",
                   "   from (                                                                             ",
                   "          select regexp_substr(t.sep_col, '[^,]+', 1,level)as dzec004_sep             ",
                   "            from dual,temp t                                                          ",
                   "         connect by level <= length(regexp_replace(t.sep_col,'[^,]+')) + 1            ",  
                   "        ) sc                                                                          "               
  OTHERWISE 
    #Pure Oracle
    LET ls_sql = "select listagg(sc.dzec004_sep,',') within group (order by sc.dzec004_sep) as sc_list                     ",
                 "  from (                                                                                                 ",
                 "         with temp as                                                                                    ",
                 "         (                                                                                               ",
                 ls_sc_sql,
                 "         )                                                                                               ",
                 "         select distinct trim(regexp_substr(t.sep_col, '[^,]+', 1, levels.column_value)) as dzec004_sep  ",
                 "           from temp t,                                                                                  ",
                 "                table(                                                                                   ",
                 "                      cast(                                                                              ",
                 "                            multiset(                                                                    ",
                 "                                      select level                                                       ",
                 "                                        from dual                                                        ",
                 "                                     connect by level <= length(regexp_replace(t.sep_col,'[^,]+')) + 1   ",
                 "                                    ) as sys.odcinumberlist                                              ",
                 "                          )                                                                              ",
                 "                     ) levels                                                                            ",
                 "       ) sc                                                                                              "
  END CASE 
  &endif
  &endif
  
  PREPARE lpre_get_resort_index_columns_list FROM ls_sql
  DECLARE lcur_get_resort_index_columns_list CURSOR FOR lpre_get_resort_index_columns_list
  OPEN lcur_get_resort_index_columns_list
  FETCH lcur_get_resort_index_columns_list INTO ls_resort_list
  CLOSE lcur_get_resort_index_columns_list
  FREE lcur_get_resort_index_columns_list
  FREE lpre_get_resort_index_columns_list

  LET ls_return = ls_resort_list
  LET ls_return = ls_return.trim()

  RETURN ls_return  
  
END FUNCTION

FUNCTION sadzi140_db_get_table_index_list(p_table_name,p_customize)
DEFINE
  p_table_name  STRING,
  p_customize   BOOLEAN
DEFINE
  ls_table_name   STRING,
  lb_customize    BOOLEAN,
  lo_resort_cols  DYNAMIC ARRAY OF T_RESORT_COLUMN,
  ls_sql          STRING,
  ls_sql_dzec007  STRING, #20181218
  li_rec_cnt      INTEGER
DEFINE
  lo_return  DYNAMIC ARRAY OF T_RESORT_COLUMN

  LET ls_table_name = p_table_name.toLowerCase()
  LET lb_customize  = p_customize

  LET li_rec_cnt = 1

  IF lb_customize THEN 
    LET ls_sql_dzec007 = " and ec.dzec007 = '",cs_dgenv_customize,"' " #20181218
  ELSE
    LET ls_sql_dzec007 = "" #20181218
  END IF
  
  LET ls_sql = "select ec.dzec001,ec.dzec002,ec.dzec004,ec.dzecstus ",
               "  from dzec_t ec                                    ",
               " where ec.dzec001 = '",ls_table_name,"'             ",
               ls_sql_dzec007 #20181218
               
  PREPARE lpre_get_customize_index_list FROM ls_sql
  DECLARE lcur_get_customize_index_list CURSOR FOR lpre_get_customize_index_list

  OPEN lcur_get_customize_index_list
  FOREACH lcur_get_customize_index_list INTO lo_resort_cols[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_customize_index_list

  FREE lpre_get_customize_index_list
  FREE lcur_get_customize_index_list    

  CALL lo_resort_cols.deleteElement(li_rec_cnt)

  LET lo_return = lo_resort_cols

  RETURN lo_return
  
END FUNCTION

FUNCTION sadzi140_db_get_connect_session_data()
DEFINE
  ls_sql          STRING,
  ls_db_type      STRING, #20240521
  lo_session_data T_SESSION_DATA
DEFINE  
  lo_return  T_SESSION_DATA

  &ifdef is_msv
  LET ls_sql = "select s.spid,                                              ", 
               "       s.kpid,                                              ", 
               "       rtrim(s.loginame) as username,                       ", 
               "       rtrim(s.loginame) as osuser,                         ", 
               "       decc.client_net_address+':'+                         ",
               "         cast(decc.client_tcp_port as nvarchar) as machine, ",
               "       s.program_name as program,                           ",
               "       s.hostprocess as process,                            ",
               "       db_name(s.dbid) dbname                               ",
               "  from sys.sysprocesses s                                   ",
               " inner join sys.dm_exec_connections as decc                 ",
               "    on s.spid = decc.session_id                             ",
               " where 1 = 1                                                ",
               "   and s.spid = @@spid                                      "
  &else
  &ifdef is_pgs
  LET ls_sql = "select psa.pid              as sid,          ",
               "       psa.usesysid         as serial_no,    ",
               "       psa.usename          as username,     ",
               "       ''                   as osuser,       ",
               "       psa.client_addr      as machine,      ",
               "       psa.application_name as program,      ",
               "       psa.datid            as process,      ",
               "       psa.datname          as service_name  ",
               "  from pg_stat_activity psa                  ",
               " where pid = pg_backend_pid()                "
  &else
  LET ls_db_type = sadzi140_util_get_db_type()
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb #20240521
      LET ls_sql = "select psa.pid              as sid,          ",
                   "       psa.usesysid         as serial_no,    ",
                   "       psa.usename          as username,     ",
                   "       ''                   as osuser,       ",
                   "       psa.client_addr      as machine,      ",
                   "       psa.application_name as program,      ",
                   "       psa.datid            as process,      ",
                   "       psa.datname          as service_name  ",
                   "  from pg_stat_activity psa                  ",
                   " where pid = pg_backend_pid()                "
    WHEN cs_db_driver_type_dmg  #20230804
      LET ls_sql = "select s.sess_id,s.sess_seq,s.user_name,'' osuser,clnt_host,",
                   "       s.appname,s.thrd_id,CUR_DATABASE                     ",
                   "  from v$sessions s                                         ",     
                   " where 1=1                                                  ",                                                         
                   "   and s.sess_id = sys_context('USERENV','SESSIONID')       " 
  OTHERWISE
    --PURE Oracle  
    LET ls_sql = "select s.sid, s.serial# serial_no, s.username, s.osuser, s.machine, ", 
                 "       s.program, s.process, s.service_name                         ",
                 "  from v$session s                                                  ", 
                 " where 1=1                                                          ",
                 "   and s.audsid = userenv('sessionid')                              "
  END CASE
  &endif
  &endif
  
  TRY                
    PREPARE lpre_get_connect_session_data FROM ls_sql
    DECLARE lcur_get_connect_session_data CURSOR FOR lpre_get_connect_session_data
    OPEN lcur_get_connect_session_data
    FETCH lcur_get_connect_session_data INTO lo_session_data.*
    CLOSE lcur_get_connect_session_data
    FREE lcur_get_connect_session_data
    FREE lpre_get_connect_session_data
  CATCH
    INITIALIZE lo_session_data TO NULL
  END TRY  

  LET lo_return.* = lo_session_data.*
  
  RETURN lo_return.*
  
END FUNCTION
#20170912 end

#20171120 begin
FUNCTION sadzi140_db_set_constraint_data(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name  STRING,
  lb_result      BOOLEAN,
  ls_key_field_stream STRING
DEFINE
  lb_return BOOLEAN  

  LET ls_table_name = p_table_name

  #先取得Key Fields
  CALL sadzi140_db_get_key_field_stream(ls_table_name) RETURNING lb_result,ls_key_field_stream
  #設定進去Key及Index中
  CALL sadzi140_db_set_key_data(ls_table_name,ls_key_field_stream) RETURNING lb_result
  CALL sadzi140_db_set_index_data(ls_table_name,ls_key_field_stream) RETURNING lb_result

  LET lb_return = lb_result

  RETURN lb_return
  
END FUNCTION 

#取得Key 欄位清單
FUNCTION sadzi140_db_get_key_field_stream(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name  STRING,
  ls_sql         STRING,
  li_rec_count   INTEGER,
  ls_field_list  STRING,
  ls_field_name  VARCHAR(50)
DEFINE
  lb_return  BOOLEAN,
  ls_return  STRING

  LET ls_table_name = p_table_name
  
  LET li_rec_count = 0
  LET ls_field_list = ""

  #取得 Key field 清單
  LET ls_sql = "select eb.dzeb002                       ",
               "  from dzeb_t eb                        ",          
               " where 1=1                              ",          
               "   and eb.dzeb004 = 'Y'                 ",          
               "   and eb.dzeb001 = '",ls_table_name,"' ", 
               " order by eb.dzeb021                    "
 
  PREPARE lpre_get_key_field_list FROM ls_sql
  DECLARE lcur_get_key_field_list CURSOR FOR lpre_get_key_field_list
  
  OPEN lcur_get_key_field_list
  FOREACH lcur_get_key_field_list INTO ls_field_name
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_count = li_rec_count + 1

    LET ls_field_list = ls_field_list,ls_field_name,","
    
  END FOREACH
  CLOSE lcur_get_key_field_list
  
  FREE lcur_get_key_field_list
  FREE lpre_get_key_field_list

  LET ls_field_list = ls_field_list.subString(1,ls_field_list.getLength()-1)
  
  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF

  LET ls_return = ls_field_list
  
  RETURN lb_return,ls_return
  
END FUNCTION

FUNCTION sadzi140_db_get_key_index_cols_to_array(p_table_name,p_pk_idx_name,p_type,p_lang)
DEFINE
  p_table_name  STRING,
  p_pk_idx_name STRING,
  p_type        STRING,
  p_lang        STRING
DEFINE
  ls_table_name  STRING,
  ls_pk_idx_name STRING,
  ls_type        STRING,
  ls_lang        STRING,
  ls_sql         STRING,
  ls_sc_sql      STRING,
  ls_db_type     STRING, #20240521
  li_count       INTEGER,
  lo_COLUMN_LIST DYNAMIC ARRAY OF T_COLUMN_LIST
DEFINE
  lo_return  DYNAMIC ARRAY OF T_COLUMN_LIST  

  LET ls_table_name  = p_table_name
  LET ls_pk_idx_name = p_pk_idx_name
  LET ls_type        = p_type
  LET ls_lang        = p_lang

  LET li_count = 1

  CASE 
    WHEN ls_type = cs_primary_key_type
      LET ls_sc_sql = "select ed.dzed004 sep_col from dzed_t ed where ed.dzed001 = '",ls_table_name,"' and ed.dzed002 = '",ls_pk_idx_name,"'" -- key
    WHEN ls_type = cs_unique_index_type
      LET ls_sc_sql = "select ec.dzec004 sep_col from dzec_t ec where ec.dzec001 = '",ls_table_name,"' and ec.dzec002 = '",ls_pk_idx_name,"'" -- index
  END CASE

  &ifdef is_msv
  LET ls_sql = "select cols.column_sep,ebl.dzebl003                                            ",
               "  from (                                                                       ",
               "        select ltrim(rtrim(m.n.value('.[1]', 'varchar(8000)'))) as column_sep  ",
               "          from (select cast('<XMLRoot><RowData>' +                             ",
               "                            replace(sc.sep_col, ',', '</RowData><RowData>') +  ",
               "                            '</RowData></XMLRoot>' as xml) as x                ",
               "                  from (                                                       ",
               ls_sc_sql,
               "                        ) as sc) t                                             ",
               "         cross apply x.nodes('/XMLRoot/RowData') m(n)                          ",
               "       ) as cols                                                               ",
               "  left outer join dzebl_t ebl                                                  ",
               "               on ebl.dzebl001 = cols.column_sep                               ",
               "              and ebl.dzebl002 = '",ls_lang,"'                                 "  
  &else
  &ifdef is_pgs
  LET ls_sql = "select cols.column_sep,ebl.dzebl003                                           ", 
               "  from (                                                                      ", 
               "         select unnest(string_to_array(t.sep_col::varchar,',')) as column_sep ", 
               "           from (                                                             ", 
               ls_sc_sql,
               "                ) t                                                           ", 
               "       ) cols                                                                 ", 
               "  left outer join dzebl_t ebl                                                 ", 
               "               on ebl.dzebl001 = cols.column_sep                              ",    
               "              and ebl.dzebl002 = '",ls_lang,"'                                " 
  &else
  LET ls_db_type = sadzi140_util_get_db_type() #20240521
  CASE ls_db_type
    WHEN cs_db_driver_type_kdb --20240521
      LET ls_sql = "select cols.column_sep,ebl.dzebl003                                           ", 
                   "  from (                                                                      ", 
                   "         select unnest(string_to_array(t.sep_col::varchar,',')) as column_sep ", 
                   "           from (                                                             ", 
                   ls_sc_sql,
                   "                ) t                                                           ", 
                   "       ) cols                                                                 ", 
                   "  left outer join dzebl_t ebl                                                 ", 
                   "               on ebl.dzebl001 = cols.column_sep                              ",    
                   "              and ebl.dzebl002 = '",ls_lang,"'                                " 
    WHEN cs_db_driver_type_dmg --20230804
      LET ls_sql = " select cols.column_sep,ebl.dzebl003                                       ",
                   "   from (                                                                  ",           
                   "          select regexp_substr(t.sep_col, '[^,]+', 1,level)as column_sep   ",          
                   "            from dual,(",ls_sc_sql,") t                                    ",                      
                   "         connect by level <= length(regexp_replace(t.sep_col,'[^,]+')) + 1 ",           
                   "        ) cols                                                             ",            
                   "   left outer join dzebl_t ebl                                             ",
                   "                on ebl.dzebl001 = cols.column_sep                          ",       
                   "               and ebl.dzebl002 = '",ls_lang,"'                            "
  OTHERWISE
    --Pure SQL   
    LET ls_sql = "select cols.column_sep,ebl.dzebl003                                                                   ",
                 "  from (                                                                                              ",
                 "        select distinct trim(regexp_substr(t.sep_col, '[^,]+', 1, levels.column_value)) as column_sep ",
                 "          from (",ls_sc_sql,") t,                                                                     ",                                                                      
                 "               table(                                                                                 ",
                 "                     cast(                                                                            ",
                 "                           multiset(                                                                  ",
                 "                                     select level                                                     ",
                 "                                       from dual                                                      ",
                 "                                    connect by level <= length(regexp_replace(t.sep_col,'[^,]+')) + 1 ",
                 "                                   ) as sys.odcinumberlist                                            ",
                 "                         )                                                                            ",
                 "                    ) levels                                                                          ",
                 "        ) cols                                                                                        ",
                 "        left outer join dzebl_t ebl on ebl.dzebl001 = cols.column_sep                                 ",
                 "                                   and ebl.dzebl002 = '",ls_lang,"'                                   "
  END CASE 
  &endif 
  &endif 
  
  PREPARE lpre_get_key_index_cols_to_array FROM ls_sql
  DECLARE lcur_get_key_index_cols_to_array CURSOR FOR lpre_get_key_index_cols_to_array
  OPEN lcur_get_key_index_cols_to_array
  FOREACH lcur_get_key_index_cols_to_array INTO lo_COLUMN_LIST[li_count].*
    IF SQLCA.sqlcode THEN
     EXIT FOREACH
    END IF

    LET li_count = li_count + 1

  END FOREACH
  CLOSE lcur_get_key_index_cols_to_array

  CALL lo_COLUMN_LIST.deleteElement(li_count)
  
  FREE lcur_get_key_index_cols_to_array
  FREE lpre_get_key_index_cols_to_array

  LET lo_return = lo_COLUMN_LIST

  RETURN lo_return
  
END FUNCTION

FUNCTION sadzi140_db_get_key_name_list_without_pk(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name     STRING,
  ls_sql            STRING,
  li_rec_count      INTEGER,
  lo_key_name_list  DYNAMIC ARRAY OF T_KEY_NAME_LIST
DEFINE
  lo_return  DYNAMIC ARRAY OF T_KEY_NAME_LIST

  LET ls_table_name = p_table_name
  
  LET li_rec_count = 1

  #取得 Key name 清單
  LET ls_sql = "select ed.dzed001,ed.dzed002,ed.dzed004        ",
               "  from dzed_t ed                               ",
               " where ed.dzed001  = '",ls_table_name,"'       ", 
               "   and ed.dzed003 <> '",cs_primary_key_type,"' "             
 
  PREPARE lpre_get_key_name_list_without_pk FROM ls_sql
  DECLARE lcur_get_key_name_list_without_pk CURSOR FOR lpre_get_key_name_list_without_pk
  
  OPEN lcur_get_key_name_list_without_pk
  FOREACH lcur_get_key_name_list_without_pk INTO lo_key_name_list[li_rec_count].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_count = li_rec_count + 1

  END FOREACH
  CLOSE lcur_get_key_name_list_without_pk
  
  CALL lo_key_name_list.deleteElement(li_rec_count)
  
  FREE lcur_get_key_name_list_without_pk
  FREE lpre_get_key_name_list_without_pk

  LET lo_return = lo_key_name_list
  
  RETURN lo_return
  
END FUNCTION

FUNCTION sadzi140_db_get_index_name_list_without_unique(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name     STRING,
  ls_sql            STRING,
  li_rec_count      INTEGER,
  lo_index_name_list  DYNAMIC ARRAY OF T_INDEX_NAME_LIST
DEFINE
  lo_return  DYNAMIC ARRAY OF T_INDEX_NAME_LIST

  LET ls_table_name = p_table_name
  
  LET li_rec_count = 1

  #取得 Index name 清單
  LET ls_sql = "select ec.dzec001,ec.dzec002,ec.dzec004         ",
               "  from dzec_t ec                                ",
               " where ec.dzec001  = '",ls_table_name,"'        ", 
               "   and ec.dzec003 <> '",cs_unique_index_type,"' "             
 
  PREPARE lpre_get_index_name_list_without_unique FROM ls_sql
  DECLARE lcur_get_index_name_list_without_unique CURSOR FOR lpre_get_index_name_list_without_unique
  
  OPEN lcur_get_index_name_list_without_unique
  FOREACH lcur_get_index_name_list_without_unique INTO lo_index_name_list[li_rec_count].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_count = li_rec_count + 1

  END FOREACH
  CLOSE lcur_get_index_name_list_without_unique
  
  CALL lo_index_name_list.deleteElement(li_rec_count)
  
  FREE lcur_get_index_name_list_without_unique
  FREE lpre_get_index_name_list_without_unique

  LET lo_return = lo_index_name_list
  
  RETURN lo_return
  
END FUNCTION

FUNCTION sadzi140_db_check_and_remove_invalid_key_index(p_column_info,p_table_name,p_lang)
DEFINE
  p_column_info DYNAMIC ARRAY OF T_COLUMN_INFO,
  p_table_name  STRING,
  p_lang        STRING
DEFINE
  lo_column_info  DYNAMIC ARRAY OF T_COLUMN_INFO,
  ls_table_name   STRING,
  ls_lang         STRING,
  li_loop         INTEGER,
  li_col_info     INTEGER,
  li_pk_idx       INTEGER,
  lb_result       BOOLEAN,
  lb_pk_idx_exist BOOLEAN,
  lo_key_name_list     DYNAMIC ARRAY OF T_KEY_NAME_LIST,
  lo_index_name_list   DYNAMIC ARRAY OF T_INDEX_NAME_LIST,
  lo_key_idx_cols_list DYNAMIC ARRAY OF T_COLUMN_LIST
DEFINE
  lb_return BOOLEAN  

  LET lo_column_info = p_column_info
  LET ls_table_name  = p_table_name
  LET ls_lang        = p_lang

  LET lb_return = TRUE

  #抓出 Key 清單
  CALL sadzi140_db_get_key_name_list_without_pk(ls_table_name) RETURNING lo_key_name_list
  FOR li_loop = 1 TO lo_key_name_list.getLength()
    #抓出每一個Key的欄位
    CALL sadzi140_db_get_key_index_cols_to_array(lo_key_name_list[li_loop].knl_TABLE_NAME,lo_key_name_list[li_loop].knl_KEY_NAME,cs_primary_key_type,ls_lang) RETURNING lo_key_idx_cols_list
    FOR li_pk_idx = 1 TO lo_key_idx_cols_list.getLength()
      LET lb_pk_idx_exist = FALSE
      FOR li_col_info = 1 TO lo_column_info.getLength()
        #如果Key欄位在表格欄位中能找的到, 就接下去下一個搜尋
        IF lo_key_idx_cols_list[li_pk_idx].COLUMN_NAME = lo_column_info[li_col_info].cs_COLUMN_NAME THEN 
          LET lb_pk_idx_exist = TRUE
          EXIT FOR
        END IF 
      END FOR
      #如果這個Key欄位在表格欄位中都抓不到, 就刪掉這個Key
      IF NOT lb_pk_idx_exist THEN
        DISPLAY cs_warning_tag,"Some field can not found in '",lo_key_name_list[li_loop].knl_TABLE_NAME,".",lo_key_name_list[li_loop].knl_KEY_NAME,".",lo_key_name_list[li_loop].knl_KEY_COLS,"', the key data will be killed."
        CALL sadzi140_db_delete_not_exist_key_data(lo_key_name_list[li_loop].knl_TABLE_NAME,lo_key_name_list[li_loop].knl_KEY_NAME) RETURNING lb_result
        EXIT FOR 
      END IF
    END FOR   
  END FOR
  
  #抓出 Index 清單
  CALL sadzi140_db_get_index_name_list_without_unique(ls_table_name) RETURNING lo_index_name_list
  FOR li_loop = 1 TO lo_index_name_list.getLength()
    CALL sadzi140_db_get_key_index_cols_to_array(lo_index_name_list[li_loop].inl_TABLE_NAME,lo_index_name_list[li_loop].inl_INDEX_NAME,cs_unique_index_type,ls_lang) RETURNING lo_key_idx_cols_list 
    FOR li_pk_idx = 1 TO lo_key_idx_cols_list.getLength()
      LET lb_pk_idx_exist = FALSE
      FOR li_col_info = 1 TO lo_column_info.getLength()
        #如果Index欄位在表格欄位中能找的到, 就接下去下一個搜尋
        IF lo_key_idx_cols_list[li_pk_idx].COLUMN_NAME = lo_column_info[li_col_info].cs_COLUMN_NAME THEN 
          LET lb_pk_idx_exist = TRUE
          EXIT FOR
        END IF 
      END FOR
      #如果這個Index欄位在表格欄位中都抓不到, 就刪掉這個Index
      IF NOT lb_pk_idx_exist THEN
        DISPLAY cs_warning_tag,"Some field can not found in '",lo_index_name_list[li_loop].inl_TABLE_NAME,".",lo_index_name_list[li_loop].inl_INDEX_NAME,".",lo_index_name_list[li_loop].inl_INDEX_COLS,"', the index data will be killed."
        CALL sadzi140_db_delete_not_exist_index_data(lo_index_name_list[li_loop].inl_TABLE_NAME,lo_index_name_list[li_loop].inl_INDEX_NAME) RETURNING lb_result
        EXIT FOR 
      END IF
    END FOR   
  END FOR

  RETURN lb_return
  
END FUNCTION


--FUNCTION sadzi140_db_

FUNCTION sadzi140_db_delete_not_exist_key_data(p_table_name,p_key_name)
DEFINE
  p_table_name  STRING,
  p_key_name    STRING
DEFINE
  ls_table_name   STRING,
  ls_key_name     STRING,
  lb_success      BOOLEAN,
  ls_sql          STRING
DEFINE
  lb_return BOOLEAN  
  
  LET ls_table_name = p_table_name.toLowerCase()
  LET ls_key_name   = p_key_name.toLowerCase()

  LET lb_success = TRUE

  LET ls_sql = "delete from dzed_t                   ",
               " where dzed001 = '",ls_table_name,"' ",
               "   and dzed002 = '",ls_key_name,"'   "

  TRY
    PREPARE lpre_delete_not_exist_key_data FROM ls_sql
    EXECUTE lpre_delete_not_exist_key_data
  CATCH
    LET lb_success = FALSE
  END TRY

  LET lb_return = lb_success

  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_db_delete_not_exist_index_data(p_table_name,p_index_name)
DEFINE
  p_table_name  STRING,
  p_index_name  STRING
DEFINE
  ls_table_name   STRING,
  ls_index_name   STRING,
  lb_success      BOOLEAN,
  ls_sql          STRING
DEFINE
  lb_return BOOLEAN  
  
  LET ls_table_name = p_table_name.toLowerCase()
  LET ls_index_name = p_index_name.toLowerCase()

  LET lb_success = TRUE

  LET ls_sql = "delete from dzec_t                   ",
               " where dzec001 = '",ls_table_name,"' ",
               "   and dzec002 = '",ls_index_name,"'  "

  TRY
    PREPARE lpre_delete_not_exist_index_data FROM ls_sql
    EXECUTE lpre_delete_not_exist_index_data
  CATCH
    LET lb_success = FALSE
  END TRY

  LET lb_return = lb_success

  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_db_get_column_with_sequence(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_table_name  STRING,
  ls_sql         STRING,
  li_rec_cnt     INTEGER,
  lo_column_info DYNAMIC ARRAY OF T_COLUMN_INFO
DEFINE
  lo_return DYNAMIC ARRAY OF T_COLUMN_INFO

  LET ls_table_name = p_table_name.toLowerCase()
  
  LET ls_sql = "select eb.dzeb001,eb.dzeb002,eb.dzeb006,eb.dzeb021 ",
               "  from dzeb_t eb                                   ",
               " where eb.dzeb001 = '",ls_table_name,"'            ",
               " order by eb.dzeb021                               "
                
  PREPARE lpre_get_column_with_sequence FROM ls_sql
  DECLARE lcur_get_column_with_sequence CURSOR FOR lpre_get_column_with_sequence

  LET li_rec_cnt = 1
  
  OPEN lcur_get_column_with_sequence
  FOREACH lcur_get_column_with_sequence INTO lo_column_info[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_column_with_sequence
  
  FREE lpre_get_column_with_sequence
  FREE lcur_get_column_with_sequence 

  CALL lo_column_info.deleteElement(li_rec_cnt)
  
  LET lo_return = lo_column_info
  
  RETURN lo_return
  
END FUNCTION 
#20171120 end

#20190819 begin
FUNCTION sadzi140_db_check_key_cols_if_exceed(p_cust_sign,p_DZEB_T,p_key_col_cnts)
DEFINE
  p_cust_sign    STRING, #20240719
  p_DZEB_T       DYNAMIC ARRAY OF T_DZEB_T,
  p_key_col_cnts INTEGER
DEFINE
  ls_cust_sign     STRING, #20240719
  lo_DZEB_T        DYNAMIC ARRAY OF T_DZEB_T,
  li_key_col_cnts  INTEGER,
  li_key_counts    INTEGER,
  lb_disable       BOOLEAN,
  lb_is_cust_env   BOOLEAN, #20240719
  lb_is_cust_table BOOLEAN, #20240719
  li_loop          INTEGER
DEFINE
  lb_return BOOLEAN,
  ls_return STRING

  LET ls_cust_sign    = p_cust_sign
  LET lo_DZEB_T       = p_DZEB_T
  LET li_key_col_cnts = p_key_col_cnts

  LET lb_return  = FALSE
  LET ls_return  = ""
  LET lb_disable = IIF(NVL(FGL_GETENV(cs_disable_pk_cols_limits),"N")=="Y",TRUE,FALSE)
  LET lb_is_cust_env = IIF(NVL(FGL_GETENV(cs_keyword_DGENV),cs_dgenv_standard)==cs_dgenv_customize,TRUE,FALSE) #20240719
  LET lb_is_cust_table  = IIF(NVL(ls_cust_sign,cs_dgenv_standard)==cs_dgenv_customize,TRUE,FALSE) #20240719
  
  FOR li_loop = 1 TO lo_DZEB_T.getLength()
    IF NVL(lo_DZEB_T[li_loop].dzeb004,"N") = "Y" THEN 
      LET li_key_counts = li_key_counts + 1
    END IF 
  END FOR 

  IF NOT lb_disable THEN
    IF li_key_counts = ci_pk_cols_limits THEN #20240416 
      LET lb_return = TRUE
      LET ls_return = ci_pk_cols_limits - 1  #20241021
    END IF
    #20240719 begin
    IF lb_is_cust_env AND lb_is_cust_table THEN 
      LET lb_return = FALSE
      LET ls_return = ""
    END IF
    #20240719 end
  END IF

  #超過資料庫最大 pk 數一律卡
  IF li_key_counts > ci_max_pk_cols_limits THEN 
    LET lb_return = TRUE
    LET ls_return = ci_max_pk_cols_limits
  END IF 

  LET ls_return = ls_return.trim()
  
  RETURN lb_return,ls_return
  
END FUNCTION 
#20190819 end

#20191209 begin
FUNCTION sadzi140_db_get_ent_mapping_db()
DEFINE
  ls_sql      STRING,
  li_topent   INTEGER,
  ls_zone     VARCHAR(100),
  ls_db_name  VARCHAR(100)
DEFINE
  ls_return  STRING
  
  LET li_topent = NVL(FGL_GETENV("TOPENT"),"99")
  LET ls_zone   = FGL_GETENV("ZONE")

  LET ls_sql = "select ou.gzou003     ", 
               "  from gzou_t ou      ", 
               " where ou.gzou001 = ? "

  PREPARE lpre_get_ent_mapping_db FROM ls_sql
  DECLARE lcur_get_ent_mapping_db CURSOR FOR lpre_get_ent_mapping_db
  OPEN lcur_get_ent_mapping_db USING li_topent
  FETCH lcur_get_ent_mapping_db INTO ls_db_name
  CLOSE lcur_get_ent_mapping_db
  FREE lcur_get_ent_mapping_db
  FREE lpre_get_ent_mapping_db

  &ifdef is_msv
  LET ls_db_name = ls_zone,"_",ls_db_name,".dbo"
  &endif
  
  LET ls_return = ls_db_name

  RETURN ls_return  
  
END FUNCTION

-- for SQL Server only begin
FUNCTION sadzi140_db_get_drop_index_before_add_pk(p_table_name,p_column_list) 
DEFINE
  p_table_name  STRING,
  p_column_list STRING
DEFINE
  ls_table_name  STRING,
  ls_column_list STRING,
  lo_column_buf  base.StringBuffer,  
  ls_proc_sql    STRING
DEFINE
  ls_return  STRING  

  LET ls_table_name  = p_table_name
  LET ls_column_list = p_column_list

  LET lo_column_buf = base.StringBuffer.create()
  CALL lo_column_buf.append(ls_column_list)
  CALL lo_column_buf.replace(",","'',''",0)
  LET ls_column_list = "''",lo_column_buf.toString(),"''"
  
  -- for SQL Server only
  LET ls_proc_sql = "declare                                                                                     ","\n",
                    "  ------------------------------------------------------------                              ","\n",
                    "  @cs_error_tag    varchar(200) = '",cs_error_tag,"',                                       ","\n",
                    "  @cs_message_tag  varchar(200) = '",cs_message_tag,"',                                     ","\n",
                    "  @ls_table_name   varchar(200),                                                            ","\n",
                    "  @ls_column_list  varchar(200),                                                            ","\n",
                    "  @ls_index_cursor varchar(4000),                                                           ","\n",
                    "  @ls_index_name   varchar(200),                                                            ","\n",
                    "  @ls_alt_sql      varchar(1000)                                                            ","\n",
                    "  ------------------------------------------------------------                              ","\n",
                    "begin                                                                                       ","\n",
                    "                                                                                            ","\n",
                    "  set @ls_table_name  = '",ls_table_name,"'                                                 ","\n",
                    "  set @ls_column_list = '",ls_column_list,"'                                                ","\n",
                    "  set @ls_index_cursor = 'declare c_indexes cursor for                                      ","\n",
                    "                           select distinct uics.table_name+''.''+uics.index_name            ","\n",
                    "                             from user_ind_columns uics                                     ","\n",
                    "                            where uics.table_name = '''+@ls_table_name+'''                  ","\n",
                    "                              and uics.index_name not like ''%pk''                          ","\n",
                    "                              and uics.column_name in ('+@ls_column_list+')'                ","\n",
                    "                                                                                            ","\n",
                    "  exec(@ls_index_cursor)                                                                    ","\n",
                    "  open c_indexes                                                                            ","\n",
                    "  fetch next from c_indexes into @ls_index_name                                             ","\n",
                    "  while @@fetch_status <> -1                                                                ","\n",
                    "  begin                                                                                     ","\n",
                    "                                                                                            ","\n",
                    "    set @ls_alt_sql = 'drop index if exists ' +@ls_index_name                               ","\n",
                    "    begin try                                                                               ","\n",
                    "      print @cs_message_tag+'Drop index => '+@ls_alt_sql                                    ","\n",
                    "      exec(@ls_alt_sql)                                                                     ","\n",
                    "    end try                                                                                 ","\n",
                    "    begin catch                                                                             ","\n",
                    "      print @cs_error_tag+'Drop index with error => '+ERROR_MESSAGE()+' [SQL]'+@ls_alt_sql  ","\n",
                    "    end catch                                                                               ","\n",
                    "                                                                                            ","\n",
                    "    fetch next from c_indexes into @ls_index_name                                           ","\n",
                    "                                                                                            ","\n",
                    "  end                                                                                       ","\n",
                    "  close c_indexes; deallocate c_indexes;                                                    ","\n",
                    "                                                                                            ","\n",
                    "end                                                                                         ","\n",
                    "go                                                                                          ","\n"

  LET ls_return = ls_proc_sql

  RETURN ls_return

END FUNCTION

FUNCTION sadzi140_db_get_create_index_after_add_pk(p_table_name,p_column_list)
DEFINE
  p_table_name  STRING,
  p_column_list STRING
DEFINE
  ls_table_name  STRING,
  ls_column_list STRING,
  lo_column_buf  base.StringBuffer,  
  ls_proc_sql    STRING
DEFINE
  ls_return  STRING  

  LET ls_table_name  = p_table_name
  LET ls_column_list = p_column_list

  LET lo_column_buf = base.StringBuffer.create()
  CALL lo_column_buf.append(ls_column_list)
  CALL lo_column_buf.replace(",","'',''",0)
  LET ls_column_list = "''",lo_column_buf.toString(),"''"
  
  -- for SQL Server only
  LET ls_proc_sql = "declare                                                                                                                   ","\n",
                    "  ------------------------------------------------------------                                                            ","\n",
                    "  @cs_error_tag    varchar(200) = '[ERROR]',                                                                              ","\n",
                    "  @cs_message_tag  varchar(200) = '[MESSAGE]',                                                                            ","\n",
                    "  @ls_table_name   varchar(200),                                                                                          ","\n",
                    "  @ls_columns      varchar(2048),                                                                                         ","\n",
                    "  @ls_index_cursor varchar(4000),                                                                                         ","\n",
                    "  @ls_index_name   varchar(200),                                                                                          ","\n",
                    "  @ls_alt_sql      varchar(1000)                                                                                          ","\n",
                    "  ------------------------------------------------------------                                                            ","\n",
                    "begin                                                                                                                     ","\n",
                    "                                                                                                                          ","\n",
                    "  set @ls_table_name  = '",ls_table_name,"'                                                                               ","\n",
                    "  set @ls_index_cursor = 'declare c_indexes cursor for                                                                    ","\n",
                    "                            select ec.dzec002,ec.dzec004                                                                  ","\n",
                    "                              from dzec_t ec                                                                              ","\n",
                    "                             where exists (                                                                               ","\n",
                    "                                            select oec.dzec002,cols.value as cols                                         ","\n",
                    "                                              from dzec_t oec                                                             ","\n",
                    "                                             cross apply string_split(oec.dzec004,'','') as cols                          ","\n",
                    "                                             where 1=1                                                                    ","\n",
                    "                                               and oec.dzec002 not like ''%pk''                                           ","\n",
                    "                                               and oec.dzec001 = '''+@ls_table_name+'''                                   ","\n",
                    "                                               and oec.dzec001 = ec.dzec001                                               ","\n",
                    "                                               and oec.dzec002 = ec.dzec002                                               ","\n",
                    "                                               and oec.dzec003 = ec.dzec003                                               ","\n",
                    "                                          )                                                                               ","\n",
                    "                               and not exists (                                                                           ","\n",
                    "                                                select 1                                                                  ","\n",
                    "                                                  from user_indexes uis                                                   ","\n",
                    "                                                 where uis.table_name = ec.dzec001                                        ","\n", 
                    "                                                   and uis.index_name = ec.dzec002                                        ","\n",
                    "                                              )                                                                           ","\n", 
                    "                         '                                                                                                ","\n",
                    "                                                                                                                          ","\n",
                    "  exec(@ls_index_cursor)                                                                                                  ","\n",
                    "  open c_indexes                                                                                                          ","\n",
                    "  fetch next from c_indexes into @ls_index_name,@ls_columns                                                               ","\n",
                    "  while @@fetch_status <> -1                                                                                              ","\n",
                    "  begin                                                                                                                   ","\n",
                    "                                                                                                                          ","\n",
                    "    set @ls_alt_sql = 'create index '+@ls_index_name+' on '+@ls_table_name+' ('+@ls_columns+')'                           ","\n",
                    "    begin try                                                                                                             ","\n",
                    "      print @cs_message_tag+'Create index => '+@ls_alt_sql                                                                ","\n",
                    "      exec(@ls_alt_sql)                                                                                                   ","\n",
                    "    end try                                                                                                               ","\n",
                    "    begin catch                                                                                                           ","\n",
                    "      print @cs_error_tag+'Create index with error => '+ERROR_MESSAGE()+' [SQL]'+@ls_alt_sql                              ","\n",
                    "    end catch                                                                                                             ","\n",
                    "                                                                                                                          ","\n",
                    "    fetch next from c_indexes into @ls_index_name,@ls_columns                                                             ","\n",
                    "                                                                                                                          ","\n",
                    "  end                                                                                                                     ","\n",
                    "  close c_indexes; deallocate c_indexes;                                                                                  ","\n",
                    "                                                                                                                          ","\n",
                    "end                                                                                                                       ","\n",
                    "go                                                                                                                        ","\n"                   

  LET ls_return = ls_proc_sql

  RETURN ls_return

END FUNCTION
-- for SQL Server only end

#20191209 end
