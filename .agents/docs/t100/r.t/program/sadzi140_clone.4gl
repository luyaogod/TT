{
  異動型態    異動單號       日 期       異動者      異動內容
  ========= ============= ========== ========== ===============================================
  Modify                  20170912   Ernestliou 1.索引匯入時判斷有無已存在 index 欄位, 有則匯入方設失效
  Modify                  20200330   02160      1.修改支援 PosgreSQL(1.1.072)   
  Modify                  20210607   02160      1.修正T100表格複製問題(簡繁以外其他語系不會複製)
}  
&include "../4gl/sadzi140_mcr.inc"

IMPORT os
IMPORT util

SCHEMA ds

&include "../4gl/sadzi140_cnst.inc"

&include "../4gl/sadzi140_type.inc"

&ifndef DEBUG
GLOBALS "../../cfg/top_global.inc"
&endif

DEFINE ms_lang STRING 

DEFINE m_common_fields DYNAMIC ARRAY WITH DIMENSION 2 OF VARCHAR(50)

FUNCTION sadzi140_clone_start(p_src_table,p_dest_table)
DEFINE 
  p_src_table, p_dest_table STRING 
DEFINE 
  ls_src_table, ls_dest_table STRING,
  lo_column_src_data     DYNAMIC ARRAY OF T_DZEB_T,
  lo_column_spec_data    DYNAMIC ARRAY OF T_DZEP_T,
  lo_key_src_data        DYNAMIC ARRAY OF T_DZED_T,
  lo_index_src_data      DYNAMIC ARRAY OF T_DZEC_T,
  lo_multi_lang_ref_data DYNAMIC ARRAY OF T_DZER_T, 
  lo_tree_src_data       DYNAMIC ARRAY OF T_DZEQ_T,
  lo_column_ref_data     DYNAMIC ARRAY OF T_DZEF_T,
  lo_state_code_data     DYNAMIC ARRAY OF T_GZCC_T,
  lo_multi_lang_data     DYNAMIC ARRAY OF T_DZEBL_ORIG_DATA, #20210607
  lb_exist_foreign_key   BOOLEAN,
  li_loop    INTEGER,
  lb_success BOOLEAN 
DEFINE
  lb_return BOOLEAN   

  LET ls_src_table  = p_src_table 
  LET ls_dest_table = p_dest_table

  LET lb_success = TRUE
  LET lb_exist_foreign_key = FALSE
  
  &ifndef DEBUG
  LET ms_lang = g_lang
  &else
  LET ms_lang = cs_default_lang
  &endif  

  #取得DZEF要用到的共用欄位設定
  CALL sadzi140_util_get_common_fields(m_common_fields)
  
  #先取得原始表格的相關資料
  CALL sadzi140_clone_get_column_src_data(ls_src_table,ls_dest_table) RETURNING lo_column_src_data
  CALL sadzi140_clone_get_column_spec_data(ls_src_table,ls_dest_table) RETURNING lo_column_spec_data
  CALL sadzi140_clone_get_key_src_data(ls_src_table,ls_dest_table) RETURNING lo_key_src_data
  CALL sadzi140_clone_get_index_src_data(ls_src_table,ls_dest_table) RETURNING lo_index_src_data
  CALL sadzi140_clone_get_column_ref_data(ls_src_table,ls_dest_table) RETURNING lo_column_ref_data
  #CALL sadzi140_clone_get_multi_lang_ref_src_data(ls_src_table,ls_dest_table) RETURNING lo_multi_lang_ref_data #20200330 marked
  CALL sadzi140_clone_get_tree_src_data(ls_src_table,ls_dest_table) RETURNING lo_tree_src_data
  CALL sadzi140_clone_get_state_code_data(ls_src_table,ls_dest_table) RETURNING lo_state_code_data
  CALL sadzi140_clone_get_multi_lang_data(ls_src_table,ls_dest_table) RETURNING lo_multi_lang_data #20210607
  
  #開始將資料匯入
  BEGIN WORK
  
  FOR li_loop = 1 TO lo_column_src_data.getLength()
    #新增到 DZEB_T 中
    CALL sadzi140_crud_insert_update_dzeb_t(lo_column_src_data[li_loop].*)
    #新增欄位多語言資料
    #CALL sadzi140_crud_insert_update_dzebl_t(lo_column_src_data[li_loop].*,ms_lang) #20210607 marked
  END FOR 
  IF NOT lb_success THEN GOTO _clone_error END IF 

  #新增欄位多語言資料 #20210607 add
  CALL sadzi140_crud_clone_dzebl_t(lo_multi_lang_data) #20210607 add
  
  FOR li_loop = 1 TO lo_column_spec_data.getLength()
    #將欄位規格資料匯入
    CALL sadzi140_crud_clone_dzep_t(lo_column_spec_data[li_loop].*)
  END FOR 
  IF NOT lb_success THEN GOTO _clone_error END IF
  
  FOR li_loop = 1 TO lo_column_ref_data.getLength()
    #將欄位參考資料匯入
    CALL sadzi140_crud_clone_dzef_t(lo_column_ref_data[li_loop].*)
  END FOR 
  IF NOT lb_success THEN GOTO _clone_error END IF 

  FOR li_loop = 1 TO lo_key_src_data.getLength()
    #判斷複製的表是否有Foreign Key, 有則提醒
    IF lo_key_src_data[li_loop].dzed003 = "R" THEN
      LET lb_exist_foreign_key = TRUE
    END IF 
    #將鍵值資料匯入
    CALL sadzi140_crud_insert_update_dzed_t(lo_key_src_data[li_loop].*)
  END FOR 
  IF NOT lb_success THEN GOTO _clone_error END IF 

  FOR li_loop = 1 TO lo_index_src_data.getLength()
    #將索引資料匯入
    CALL sadzi140_crud_insert_update_dzec_t(lo_index_src_data[li_loop].*)
  END FOR 
  IF NOT lb_success THEN GOTO _clone_error END IF 

  {
  #20200330 marked
  FOR li_loop = 1 TO lo_multi_lang_ref_data.getLength()
    #將資料多語言設定匯入
    CALL sadzi140_crud_insert_update_dzer_t(lo_multi_lang_ref_data[li_loop].*)
  END FOR   
  IF NOT lb_success THEN GOTO _clone_error END IF
  } 

  FOR li_loop = 1 TO lo_tree_src_data.getLength()
    #將樹狀結構設定匯入
    CALL sadzi140_crud_insert_update_dzeq_t(lo_tree_src_data[li_loop].*)
  END FOR   
  IF NOT lb_success THEN GOTO _clone_error END IF 
  
  FOR li_loop = 1 TO lo_state_code_data.getLength()
    #將狀態碼設定匯入
    CALL sadzi140_crud_clone_gzcc_t(lo_state_code_data[li_loop].*)
  END FOR   
  IF NOT lb_success THEN GOTO _clone_error END IF
  
  LABEL _clone_error:
  
  IF NOT lb_success THEN
    ROLLBACK WORK 
  ELSE 
    COMMIT WORK
  END IF
  
  LET lb_return = lb_success
    
  RETURN lb_return, lb_exist_foreign_key 

END FUNCTION 

FUNCTION sadzi140_clone_get_column_src_data(p_src_table,p_dest_table)
DEFINE
  p_src_table  STRING,
  p_dest_table STRING
DEFINE  
  ls_src_table         STRING,
  ls_dest_table        STRING,
  ls_src_leading_code  STRING,
  ls_dest_leading_code STRING,
  ls_DGENV             STRING, 
  lo_column_src_data   DYNAMIC ARRAY OF T_DZEB_T
DEFINE  
  ls_SQL     STRING,
  li_RecCnt  INTEGER,
  lo_RETURN  DYNAMIC ARRAY OF T_DZEB_T

  LET ls_src_table  = p_src_table
  LET ls_dest_table = p_dest_table

  LET ls_DGENV = FGL_GETENV("DGENV")

  CALL sadzi140_util_get_table_leading_code(ls_src_table) RETURNING ls_src_leading_code
  CALL sadzi140_util_get_table_leading_code(ls_dest_table) RETURNING ls_dest_leading_code
  
  INITIALIZE lo_column_src_data TO NULL
  INITIALIZE lo_RETURN TO NULL
  
  LET ls_SQL = " select '','','',                                                                      ",
               "        replace(dzeb001,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzeb001, ",
               "        replace(dzeb002,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzeb002, ",
               "        dzebl003 dzeb003,                                                              ", #20160901 modify
               "        dzeb004,                                                                       ",
               "        dzeb005,                                                                       ",
               "        dzeb006,                                                                       ",
               "        '' dzeb006_desc,                                                               ",
               "        dzeb007,                                                                       ",
               "        dzeb008,                                                                       ",
               "        dzeb012,                                                                       ",
               "        dzeb021,                                                                       ",
               "        dzeb022,                                                                       ",
               "        dzeb023,                                                                       ",
               "        dzeb024,                                                                       ",
               "        dzeb027,                                                                       ",
               "        'N' dzeb028,                                                                   ",
               "        '",ls_DGENV,"' dzeb029,                                                        ",
               "        '",ls_DGENV,"' dzeb030,                                                        ",
               "        'N'            dzeb031,                                                        ",
               "        ''             dzeb002_orig,                                                   ",
               "        ''             dzep012,                                                        ",
               "        ''             category,                                                       ",
               "        '",cs_lang_modified,"'  LANG_MODIFIED                                          ",
               "   from dzeb_t                                                                         ",
               "   left outer join dzebl_t on dzebl001 = dzeb002                                       ", #20160901 begin
               "                          and dzebl002 = '",ms_lang,"'                                 ", #20160901 end
               "  where dzeb001 = '",ls_src_table,"'                                                   "

  PREPARE lpre_get_column_src_data FROM ls_SQL
  DECLARE lcur_get_column_src_data CURSOR FOR lpre_get_column_src_data

  LET li_RecCnt = 1 
  
  OPEN lcur_get_column_src_data
  FOREACH lcur_get_column_src_data INTO lo_column_src_data[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_get_column_src_data
  CALL lo_column_src_data.deleteElement(li_RecCnt)
  
  FREE lpre_get_column_src_data
  FREE lcur_get_column_src_data  

  LET lo_RETURN.* = lo_column_src_data.* 
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_clone_get_key_src_data(p_src_table,p_dest_table)
DEFINE
  p_src_table  STRING,
  p_dest_table STRING
DEFINE  
  ls_src_table  STRING,
  ls_dest_table STRING,
  ls_src_leading_code  STRING,
  ls_dest_leading_code STRING,
  ls_DGENV        STRING,
  lo_key_src_data DYNAMIC ARRAY OF T_DZED_T
DEFINE  
  ls_SQL     STRING,
  li_RecCnt  INTEGER,
  lo_RETURN  DYNAMIC ARRAY OF T_DZED_T

  LET ls_src_table  = p_src_table
  LET ls_dest_table = p_dest_table

  LET ls_DGENV = FGL_GETENV("DGENV")

  CALL sadzi140_util_get_table_leading_code(ls_src_table) RETURNING ls_src_leading_code
  CALL sadzi140_util_get_table_leading_code(ls_dest_table) RETURNING ls_dest_leading_code
  
  INITIALIZE lo_key_src_data TO NULL
  INITIALIZE lo_RETURN TO NULL
  
  LET ls_SQL = "select '',                                                                            ",
               "       replace(dzed001,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzed001, ",
               "       replace(dzed002,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzed002, ",
               "       dzed003,                                                                       ",
               "       replace(dzed004,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzed004, ",
               "       dzed005,                                                                       ",
               "       dzed006,                                                                       ",
               "       'N' dzed007,                                                                   ",
               "       '",ls_DGENV,"' dzed008,                                                        ",                   
               "       '",ls_DGENV,"' dzed009,                                                        ",                   
               "       'N' dzed010,                                                                   ",
               "       ''                                                                             ",
               "  from dzed_t                                                                         ",
               " where dzed001 = '",ls_src_table,"'                                                   "
   
  PREPARE lpre_get_key_src_data FROM ls_SQL
  DECLARE lcur_get_key_src_data CURSOR FOR lpre_get_key_src_data

  LET li_RecCnt = 1 
  
  OPEN lcur_get_key_src_data
  FOREACH lcur_get_key_src_data INTO lo_key_src_data[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_get_key_src_data
  CALL lo_key_src_data.deleteElement(li_RecCnt)
  
  FREE lpre_get_key_src_data
  FREE lcur_get_key_src_data  

  LET lo_RETURN.* = lo_key_src_data.* 
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_clone_get_index_src_data(p_src_table,p_dest_table)
DEFINE
  p_src_table  STRING,
  p_dest_table STRING
DEFINE  
  ls_src_table  STRING,
  ls_dest_table STRING,
  ls_src_leading_code  STRING,
  ls_dest_leading_code STRING,
  ls_DGENV          STRING,
  lo_index_src_data DYNAMIC ARRAY OF T_DZEC_T
DEFINE  
  ls_SQL     STRING,
  li_RecCnt  INTEGER,
  lo_RETURN  DYNAMIC ARRAY OF T_DZEC_T

  LET ls_src_table  = p_src_table
  LET ls_dest_table = p_dest_table

  LET ls_DGENV = FGL_GETENV("DGENV")
  
  CALL sadzi140_util_get_table_leading_code(ls_src_table) RETURNING ls_src_leading_code
  CALL sadzi140_util_get_table_leading_code(ls_dest_table) RETURNING ls_dest_leading_code
  
  INITIALIZE lo_index_src_data TO NULL
  INITIALIZE lo_RETURN TO NULL
  
  LET ls_SQL = "select '',                                                                            ",
               "       replace(dzec001,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzec001, ",
               "       replace(dzec002,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzec002, ",
               "       dzec003,                                                                       ",
               "       replace(dzec004,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzec004, ",
               "       'N' dzec005,                                                                   ",
               "       '",ls_DGENV,"' dzec006,                                                        ",
               "       '",ls_DGENV,"' dzec007,                                                        ",
               "       'N' dzec008,                                                                   ",                          
               "       dzecstus,                                                                      ", #20170912
               "       ''                                                                             ",
               "  from dzec_t                                                                         ",
               " where dzec001 = '",ls_src_table,"'                                                   "
   
  PREPARE lpre_get_index_src_data FROM ls_SQL
  DECLARE lcur_get_index_src_data CURSOR FOR lpre_get_index_src_data

  LET li_RecCnt = 1 
  
  OPEN lcur_get_index_src_data
  FOREACH lcur_get_index_src_data INTO lo_index_src_data[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_get_index_src_data
  CALL lo_index_src_data.deleteElement(li_RecCnt)
  
  FREE lpre_get_index_src_data
  FREE lcur_get_index_src_data  

  LET lo_RETURN.* = lo_index_src_data.* 
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_clone_get_multi_lang_ref_src_data(p_src_table,p_dest_table)
DEFINE
  p_src_table  STRING,
  p_dest_table STRING
DEFINE  
  ls_src_table  STRING,
  ls_dest_table STRING,
  ls_src_leading_code  STRING,
  ls_dest_leading_code STRING,
  lo_multi_lang_src_data DYNAMIC ARRAY OF T_DZER_T
DEFINE  
  ls_SQL     STRING,
  li_RecCnt  INTEGER,
  lo_RETURN  DYNAMIC ARRAY OF T_DZER_T

  LET ls_src_table  = p_src_table
  LET ls_dest_table = p_dest_table

  CALL sadzi140_util_get_table_leading_code(ls_src_table) RETURNING ls_src_leading_code
  CALL sadzi140_util_get_table_leading_code(ls_dest_table) RETURNING ls_dest_leading_code
  
  INITIALIZE lo_multi_lang_src_data TO NULL
  INITIALIZE lo_RETURN TO NULL
  
  LET ls_SQL = "select '',                                                                            ",
               "       replace(dzer001,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzer001, ",
               "       replace(dzer002,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzer002, ",
               "       dzer003,                                                                       ",
               "       replace(dzer004,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzer004, ",
               "       replace(dzer005,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzer005, ",
               "       replace(dzer006,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzer006, ",
               "       replace(dzer007,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzer007, ",
               "       replace(dzer008,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzer008  ",
               "  from dzer_t                                                                         ",
               " where dzer001 = '",ls_src_table,"'                                                   "
   
  PREPARE lpre_get_multi_lang_src_data FROM ls_SQL
  DECLARE lcur_get_multi_lang_src_data CURSOR FOR lpre_get_multi_lang_src_data

  LET li_RecCnt = 1 
  
  OPEN lcur_get_multi_lang_src_data
  FOREACH lcur_get_multi_lang_src_data INTO lo_multi_lang_src_data[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_get_multi_lang_src_data
  CALL lo_multi_lang_src_data.deleteElement(li_RecCnt)
  
  FREE lpre_get_multi_lang_src_data
  FREE lcur_get_multi_lang_src_data  

  LET lo_RETURN.* = lo_multi_lang_src_data.* 
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_clone_get_tree_src_data(p_src_table,p_dest_table)
DEFINE
  p_src_table  STRING,
  p_dest_table STRING
DEFINE  
  ls_src_table  STRING,
  ls_dest_table STRING,
  ls_src_leading_code  STRING,
  ls_dest_leading_code STRING,
  lo_tree_src_data DYNAMIC ARRAY OF T_DZEQ_T
DEFINE  
  ls_SQL     STRING,
  li_RecCnt  INTEGER,
  lo_RETURN  DYNAMIC ARRAY OF T_DZEQ_T

  LET ls_src_table  = p_src_table
  LET ls_dest_table = p_dest_table

  CALL sadzi140_util_get_table_leading_code(ls_src_table) RETURNING ls_src_leading_code
  CALL sadzi140_util_get_table_leading_code(ls_dest_table) RETURNING ls_dest_leading_code
  
  INITIALIZE lo_tree_src_data TO NULL
  INITIALIZE lo_RETURN TO NULL
  
  LET ls_SQL = "select '',                                                                            ",
               "       replace(dzeq001,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzeq001, ",
               "       dzeq002,                                                                       ",
               "       dzeq003,                                                                       ",
               "       dzeq004,                                                                       ",
               "       dzeq005,                                                                       ",
               "       dzeq006,                                                                       ",
               "       dzeq007,                                                                       ",
               "       replace(dzeq008,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzeq008, ",
               "       replace(dzeq009,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzeq009  ",
               "  from dzeq_t                                                                         ",
               " where dzeq001 = '",ls_src_table,"'                                                   "
   
  PREPARE lpre_get_tree_src_data FROM ls_SQL
  DECLARE lcur_get_tree_src_data CURSOR FOR lpre_get_tree_src_data

  LET li_RecCnt = 1 
  
  OPEN lcur_get_tree_src_data
  FOREACH lcur_get_tree_src_data INTO lo_tree_src_data[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_get_tree_src_data
  CALL lo_tree_src_data.deleteElement(li_RecCnt)
  
  FREE lpre_get_tree_src_data
  FREE lcur_get_tree_src_data  

  LET lo_RETURN.* = lo_tree_src_data.* 
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_clone_get_column_spec_data(p_src_table,p_dest_table)
DEFINE
  p_src_table  STRING,
  p_dest_table STRING
DEFINE  
  ls_src_table         STRING,
  ls_dest_table        STRING,
  ls_src_leading_code  STRING,
  ls_dest_leading_code STRING,
  lo_column_spec_data  DYNAMIC ARRAY OF T_DZEP_T
DEFINE  
  ls_SQL     STRING,
  li_RecCnt  INTEGER,
  lo_RETURN  DYNAMIC ARRAY OF T_DZEP_T

  LET ls_src_table  = p_src_table
  LET ls_dest_table = p_dest_table

  CALL sadzi140_util_get_table_leading_code(ls_src_table) RETURNING ls_src_leading_code
  CALL sadzi140_util_get_table_leading_code(ls_dest_table) RETURNING ls_dest_leading_code
  
  INITIALIZE lo_column_spec_data TO NULL
  INITIALIZE lo_RETURN TO NULL
  
  LET ls_SQL = "select '',                                                                            ",
               "       replace(dzep001,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzep001, ",
               "       replace(dzep002,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzep002, ",
               "       dzep003,                                                                       ",
               "       dzep004,                                                                       ",
               "       dzep005,                                                                       ",
               "       dzep006,                                                                       ",
               "       dzep007,                                                                       ",
               "       dzep008,                                                                       ",
               "       dzep009,                                                                       ",
               "       dzep010,                                                                       ",
               "       dzep011,                                                                       ",
               "       dzep012,                                                                       ",
               "       dzep013,                                                                       ",
               "       dzep014,                                                                       ",
               "       dzep015,                                                                       ",
               "       dzep016,                                                                       ",
               "       dzep017,                                                                       ",
               "       dzep018,                                                                       ",
               "       dzep019,                                                                       ",
               "       dzep020,                                                                       ",
               "       dzep021,                                                                       ",
               "       dzep022,                                                                       ",
               "       dzep023,                                                                       ",
               "       dzep024,                                                                       ",
               "       dzep025,                                                                       ",
               "       dzep026,                                                                       ",
               "       dzep027,                                                                       ",
               "       dzep028                                                                        ",
               "  from dzep_t                                                                         ",
               " where dzep001 = '",ls_src_table,"'                                                   "

  PREPARE lpre_get_column_spec_data FROM ls_SQL
  DECLARE lcur_get_column_spec_data CURSOR FOR lpre_get_column_spec_data

  LET li_RecCnt = 1 
  
  OPEN lcur_get_column_spec_data
  FOREACH lcur_get_column_spec_data INTO lo_column_spec_data[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_get_column_spec_data
  CALL lo_column_spec_data.deleteElement(li_RecCnt)
  
  FREE lpre_get_column_spec_data
  FREE lcur_get_column_spec_data  

  LET lo_RETURN.* = lo_column_spec_data.* 
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_clone_get_column_ref_data(p_src_table,p_dest_table)
DEFINE
  p_src_table  STRING,
  p_dest_table STRING
DEFINE  
  ls_src_table         STRING,
  ls_dest_table        STRING,
  ls_src_leading_code  STRING,
  ls_dest_leading_code STRING,
  lo_column_ref_data   DYNAMIC ARRAY OF T_DZEF_T
DEFINE  
  ls_SQL     STRING,
  li_RecCnt  INTEGER,
  lo_RETURN  DYNAMIC ARRAY OF T_DZEF_T

  LET ls_src_table  = p_src_table
  LET ls_dest_table = p_dest_table

  CALL sadzi140_util_get_table_leading_code(ls_src_table) RETURNING ls_src_leading_code
  CALL sadzi140_util_get_table_leading_code(ls_dest_table) RETURNING ls_dest_leading_code
  
  INITIALIZE lo_column_ref_data TO NULL
  INITIALIZE lo_RETURN TO NULL
  
  LET ls_SQL = "select '',                                                                            ",
               "       replace(dzef001,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzef001, ",
               "       replace(dzef002,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzef002, ",
               "       replace(dzef003,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzef003, ",
               "       dzef004,                                                                       ",
               "       dzef005,                                                                       ",
               "       dzef006,                                                                       ",
               "       dzef007,                                                                       ",
               "       dzef008,                                                                       ",
               "       dzef009,                                                                       ",
               "       dzef010                                                                        ",
               "  from dzef_t                                                                         ",
               " where dzef001 = '",ls_src_table,"'                                                   "

  PREPARE lpre_get_column_ref_data FROM ls_SQL
  DECLARE lcur_get_column_ref_data CURSOR FOR lpre_get_column_ref_data

  LET li_RecCnt = 1 
  
  OPEN lcur_get_column_ref_data
  FOREACH lcur_get_column_ref_data INTO lo_column_ref_data[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_get_column_ref_data
  CALL lo_column_ref_data.deleteElement(li_RecCnt)
  
  FREE lpre_get_column_ref_data
  FREE lcur_get_column_ref_data  

  LET lo_RETURN.* = lo_column_ref_data.* 
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_clone_get_state_code_data(p_src_table,p_dest_table)
DEFINE
  p_src_table  STRING,
  p_dest_table STRING
DEFINE  
  ls_src_table         STRING,
  ls_dest_table        STRING,
  ls_src_leading_code  STRING,
  ls_dest_leading_code STRING,
  lo_state_code_data   DYNAMIC ARRAY OF T_GZCC_T
DEFINE  
  ls_SQL     STRING,
  li_RecCnt  INTEGER,
  lo_RETURN  DYNAMIC ARRAY OF T_GZCC_T

  LET ls_src_table  = p_src_table
  LET ls_dest_table = p_dest_table

  CALL sadzi140_util_get_table_leading_code(ls_src_table) RETURNING ls_src_leading_code
  CALL sadzi140_util_get_table_leading_code(ls_dest_table) RETURNING ls_dest_leading_code
  
  INITIALIZE lo_state_code_data TO NULL
  INITIALIZE lo_RETURN TO NULL
  
  LET ls_SQL = "select replace(gzcc001,'",ls_src_leading_code,"','",ls_dest_leading_code,"') gzcc001, ",
               "       replace(gzcc002,'",ls_src_leading_code,"','",ls_dest_leading_code,"') gzcc002, ",
               "       gzcc003,                                                                       ",
               "       gzcc004,                                                                       ",
               "       gzcc005,                                                                       ",
               "       gzcc006,                                                                       ",
               "       gzccstus                                                                       ",
               "  from gzcc_t                                                                         ",
               " where gzcc001 = '",ls_src_table,"'                                                   "

  PREPARE lpre_get_state_code_data FROM ls_SQL
  DECLARE lcur_get_state_code_data CURSOR FOR lpre_get_state_code_data

  LET li_RecCnt = 1 
  
  OPEN lcur_get_state_code_data
  FOREACH lcur_get_state_code_data INTO lo_state_code_data[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_get_state_code_data
  CALL lo_state_code_data.deleteElement(li_RecCnt)
  
  FREE lpre_get_state_code_data
  FREE lcur_get_state_code_data  

  LET lo_RETURN.* = lo_state_code_data.* 
  
  RETURN lo_RETURN
  
END FUNCTION

#20210607 begin
FUNCTION sadzi140_clone_get_multi_lang_data(p_src_table,p_dest_table)
DEFINE
  p_src_table  STRING,
  p_dest_table STRING
DEFINE  
  ls_src_table         STRING,
  ls_dest_table        STRING,
  ls_src_leading_code  STRING,
  ls_dest_leading_code STRING,
  lo_multi_lang_data   DYNAMIC ARRAY OF T_DZEBL_ORIG_DATA
DEFINE  
  ls_SQL     STRING,
  li_RecCnt  INTEGER,
  lo_RETURN  DYNAMIC ARRAY OF T_DZEBL_ORIG_DATA

  LET ls_src_table  = p_src_table
  LET ls_dest_table = p_dest_table

  CALL sadzi140_util_get_table_leading_code(ls_src_table) RETURNING ls_src_leading_code
  CALL sadzi140_util_get_table_leading_code(ls_dest_table) RETURNING ls_dest_leading_code
  
  INITIALIZE lo_multi_lang_data TO NULL
  INITIALIZE lo_RETURN TO NULL
  
  LET ls_SQL = "select coalesce(replace(dzebl000,'",ls_src_leading_code,"','",ls_dest_leading_code,"'), ",
               "                '",ls_dest_table,"') dzebl000,                                          ",
               "       replace(dzebl001,'",ls_src_leading_code,"','",ls_dest_leading_code,"') dzebl001, ",
               "       dzebl002,                                                                        ",
               "       dzebl003,                                                                        ",
               "       dzebl004                                                                         ",
               "  from dzebl_t                                                                          ",
               " where dzebl001 like '",ls_src_leading_code,"%'                                         ",
               "   and dzebl001 not like '",ls_dest_leading_code,"%'                                    "
   
  PREPARE lpre_get_multi_lang_data FROM ls_SQL
  DECLARE lcur_get_multi_lang_data CURSOR FOR lpre_get_multi_lang_data

  LET li_RecCnt = 1 
  
  OPEN lcur_get_multi_lang_data
  FOREACH lcur_get_multi_lang_data INTO lo_multi_lang_data[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_get_multi_lang_data
  CALL lo_multi_lang_data.deleteElement(li_RecCnt)
  
  FREE lpre_get_multi_lang_data
  FREE lcur_get_multi_lang_data  

  LET lo_RETURN = lo_multi_lang_data 
  
  RETURN lo_RETURN
  
END FUNCTION
#20210607 end