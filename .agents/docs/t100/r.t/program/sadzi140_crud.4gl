{
  異動型態    異動單號       日 期       異動者      異動內容
  ========= ============= ========== ========== ===============================================
  Modify                  20170214   Ernestliou 1.產出 tbl 檔改為判斷 DGENV來做決定是否產出客戶家的語言資料(除了zh_TW,zh_CN以外)
  Modify                  20170425   Ernestliou 1.Mark insert dzebl_t warning message
  Modify                  20170706   Ernestliou 1.修正產生多語言資料時除了zh_TW,zh_CN以外, 其他語系均不呼叫轉換函式
  Modify                  20170912   Ernestliou 1.索引匯入時判斷有無已存在 index 欄位, 有則匯入方設失效
  Modify                  20171120   02160      1.dzep005 若非為 "Y", 則將 dzeb005 的值設進來
  Modify                  20180302   02160      1. 刪除標準欄位時給予警告, 並將刪除的欄位紀錄在表格(dzhj_t)中
  Modify                  20190424   02160      1.支援在建表時同步建立實體紀錄 Trigger 及 adzi190 上的 Trigger Scripts
  Modify                  20200330   02160      1.修改支援 PosgreSQL(1.1.072)   
  Modify                  20210607   02160      1.修正T100表格複製問題(簡繁以外其他語系不會複製)
  Modify                  20221017   02160      1.當自定義欄位型態是 cdfUserDefineVarchar 時, 則該欄位大小寫清空
}  

&include "../4gl/sadzi140_mcr.inc"

IMPORT os
IMPORT util

SCHEMA ds

&include "../4gl/sadzp000_cnst.inc"
&include "../4gl/sadzi140_cnst.inc"
&include "../4gl/sadzp200_cnst.inc"

&include "../4gl/sadzp000_type.inc"
&include "../4gl/sadzi140_type.inc"

&ifndef DEBUG
GLOBALS "../../cfg/top_global.inc"
&endif

FUNCTION sadzi140_crud_insert_dzea_t(p_create_table)
DEFINE
  p_create_table T_DZEA_CREATE_TABLE
DEFINE  
  lo_create_table T_DZEA_CREATE_TABLE
DEFINE
  ls_replace_arg STRING,
  lb_result      BOOLEAN,
  ls_user        VARCHAR(100),
  ldt_datetime    DATETIME YEAR TO SECOND

  LET lo_create_table.* = p_create_table.*
  LET lb_result = TRUE

  &ifndef DEBUG
  LET ls_user = g_user
  LET ldt_datetime = cl_get_current()
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  LET ldt_datetime = CURRENT YEAR TO SECOND
  &endif
  
  TRY
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    insert into dzea_t(
                       dzea001,dzea002,dzea003,dzea004,dzea005,
                       dzea006,dzea007,dzea008,dzea009,dzea010,
                       dzea011,dzea012,dzea013,dzea014,dzea015,
                       dzea016,dzea017,dzea018,dzea019,dzea020,
                       dzea021,
                       dzeacrtid,dzeacrtdt,dzeamodid,dzeamoddt
                      ) 
               values (
                       lo_create_table.dct_table_name,  lo_create_table.dct_table_description, lo_create_table.dct_module_name, lo_create_table.dct_table_type,    lo_create_table.dct_is_multi_lang_table,
                       lo_create_table.dct_is_system_table,lo_create_table.dct_create_notice_trigger, lo_create_table.dct_define_group,    lo_create_table.dct_up_level_table, lo_create_table.dct_table_space, #20190424
                       "Y",lo_create_table.dct_alm_construct_version,lo_create_table.dct_alm_sd_version,lo_create_table.dct_industry_type,lo_create_table.dct_dgenv,
                       lo_create_table.dct_orig_module_name,lo_create_table.dct_shipping_notice,lo_create_table.dct_orig_dgenv,"","",
                       "",
                       ls_user,ldt_datetime,ls_user,ldt_datetime
                      )
    
  CATCH 
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif
    LET lb_result = FALSE
  END TRY

  RETURN lb_result
  
END FUNCTION

FUNCTION sadzi140_crud_insert_update_dzeb_t(p_t_dzeb_t)
DEFINE
  p_t_dzeb_t   T_DZEB_T
DEFINE 
  lo_t_dzeb_t  T_DZEB_T,
  ls_user      VARCHAR(100),
  ldt_datetime  DATETIME YEAR TO SECOND

  LET lo_t_dzeb_t.* = p_t_dzeb_t.*

  &ifndef DEBUG
  LET ls_user = g_user
  LET ldt_datetime = cl_get_current()
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  LET ldt_datetime = CURRENT YEAR TO SECOND
  &endif

  TRY 
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
  
    insert into dzeb_t(
                        dzeb001,dzeb002,dzeb003,dzeb004,dzeb005,
                        dzeb006,dzeb007,dzeb008,dzeb012,dzeb021,
                        dzeb022,dzeb023,dzeb024,dzeb027,dzeb028,
                        dzeb029,dzeb030,dzeb031,
                        dzebcrtid,dzebcrtdt,dzebmodid,dzebmoddt
                      )
               values (
                        lo_t_dzeb_t.dzeb001,lo_t_dzeb_t.dzeb002,lo_t_dzeb_t.dzeb003,lo_t_dzeb_t.dzeb004,lo_t_dzeb_t.dzeb005,
                        lo_t_dzeb_t.dzeb006,lo_t_dzeb_t.dzeb007,lo_t_dzeb_t.dzeb008,lo_t_dzeb_t.dzeb012,lo_t_dzeb_t.dzeb021,
                        lo_t_dzeb_t.dzeb022,lo_t_dzeb_t.dzeb023,lo_t_dzeb_t.dzeb024,lo_t_dzeb_t.dzeb027,lo_t_dzeb_t.dzeb028,
                        lo_t_dzeb_t.dzeb029,lo_t_dzeb_t.dzeb030,lo_t_dzeb_t.dzeb031,
                        ls_user,ldt_datetime,ls_user,ldt_datetime
                      )
                       
  CATCH
    TRY
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif

      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
    
      #DISPLAY "INSERT NOT SUCCESS TRY Update"
      update dzeb_t                                 
         set dzeb003   = lo_t_dzeb_t.dzeb003,
             dzeb004   = lo_t_dzeb_t.dzeb004,
             dzeb005   = lo_t_dzeb_t.dzeb005,
             dzeb006   = lo_t_dzeb_t.dzeb006,
             dzeb007   = lo_t_dzeb_t.dzeb007,
             dzeb008   = lo_t_dzeb_t.dzeb008,
             dzeb012   = lo_t_dzeb_t.dzeb012,
             dzeb021   = lo_t_dzeb_t.dzeb021,
             dzeb022   = lo_t_dzeb_t.dzeb022,
             dzeb023   = lo_t_dzeb_t.dzeb023,
             dzeb024   = lo_t_dzeb_t.dzeb024,
             dzeb027   = lo_t_dzeb_t.dzeb027,
             dzeb028   = lo_t_dzeb_t.dzeb028,
             dzeb029   = lo_t_dzeb_t.dzeb029,
             dzeb030   = lo_t_dzeb_t.dzeb030,
             dzeb031   = lo_t_dzeb_t.dzeb031,
             dzebmodid = ls_user,
             dzebmoddt = ldt_datetime
       where dzeb001 = lo_t_dzeb_t.dzeb001
         and dzeb002 = lo_t_dzeb_t.dzeb002
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      DISPLAY cs_warning_tag,"Insert or Update DZEB_T with column ",lo_t_dzeb_t.DZEB002," unsuccess : ",SQLCA.sqlcode
    END TRY
  END TRY

END FUNCTION

FUNCTION sadzi140_crud_insert_update_dzec_t(p_t_dzec_t)
DEFINE
  p_t_dzec_t  T_DZEC_T
DEFINE 
  lo_t_dzec_t  T_DZEC_T,
  lo_str_buf   base.StringBuffer,
  ls_user      VARCHAR(100),
  ldt_datetime  DATETIME YEAR TO SECOND

  LET lo_t_dzec_t.* = p_t_dzec_t.*

  &ifndef DEBUG
  LET ls_user = g_user
  LET ldt_datetime = cl_get_current()
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  LET ldt_datetime = CURRENT YEAR TO SECOND
  &endif

  #移除空格 begin
  LET lo_str_buf = base.StringBuffer.create()
  CALL lo_str_buf.clear()
  
  CALL lo_str_buf.append(lo_t_dzec_t.dzec004)
  CALL lo_str_buf.replace(" ","",0)
  LET lo_t_dzec_t.dzec004 = lo_str_buf.toString()
  #移除空格 end
  
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
    TRY
      #DISPLAY "INSERT NOT SUCCESS TRY Update"
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      update dzec_t                                 
         set dzec003   = lo_t_dzec_t.dzec003,
             dzec004   = lo_t_dzec_t.dzec004,
             dzec005   = lo_t_dzec_t.dzec005,
             dzec006   = lo_t_dzec_t.dzec006,
             dzec007   = lo_t_dzec_t.dzec007,
             dzec008   = lo_t_dzec_t.dzec008,
             dzecmodid = ls_user,
             dzecmoddt = ldt_datetime
       where dzec001   = lo_t_dzec_t.dzec001
         and dzec002   = lo_t_dzec_t.dzec002
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      DISPLAY cs_warning_tag,"Insert or Update DZEC_T with index ",lo_t_dzec_t.dzec002," unsuccess : ",SQLCA.sqlcode
    END TRY
  END TRY

END FUNCTION

FUNCTION sadzi140_crud_insert_update_dzed_t(p_t_dzed_t)
DEFINE
  p_t_dzed_t  T_DZED_T
DEFINE 
  lo_t_dzed_t  T_DZED_T,
  lo_str_buf   base.StringBuffer,
  ls_user      VARCHAR(100),
  ldt_datetime DATETIME YEAR TO SECOND

  LET lo_t_dzed_t.* = p_t_dzed_t.*

  &ifndef DEBUG
  LET ls_user = g_user
  LET ldt_datetime = cl_get_current()
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  LET ldt_datetime = CURRENT YEAR TO SECOND
  &endif

  #移除空格 begin
  LET lo_str_buf = base.StringBuffer.create()

  CALL lo_str_buf.clear()
  CALL lo_str_buf.append(lo_t_dzed_t.dzed004)
  CALL lo_str_buf.replace(" ","",0)
  LET lo_t_dzed_t.dzed004 = lo_str_buf.toString()

  CALL lo_str_buf.clear()
  CALL lo_str_buf.append(lo_t_dzed_t.dzed006)
  CALL lo_str_buf.replace(" ","",0)
  LET lo_t_dzed_t.dzed006 = lo_str_buf.toString()
  #移除空格 end
  
  TRY 
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    insert into dzed_t(
                        dzed001,dzed002,dzed003,dzed004,dzed005,
                        dzed006,dzed007,dzed008,dzed009,dzed010,
                        dzedcrtid,dzedcrtdt,dzedmodid,dzedmoddt
                      )
               values (
                        lo_t_dzed_t.dzed001,lo_t_dzed_t.dzed002,lo_t_dzed_t.dzed003,lo_t_dzed_t.dzed004,lo_t_dzed_t.dzed005,
                        lo_t_dzed_t.dzed006,lo_t_dzed_t.dzed007,lo_t_dzed_t.dzed008,lo_t_dzed_t.dzed009,lo_t_dzed_t.dzed010,
                        ls_user,ldt_datetime,ls_user,ldt_datetime
                      )
                       
  CATCH
    TRY
      #DISPLAY "INSERT NOT SUCCESS TRY Update"
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      update dzed_t                                 
         set dzed003   = lo_t_dzed_t.dzed003,
             dzed004   = lo_t_dzed_t.dzed004,
             dzed005   = lo_t_dzed_t.dzed005,
             dzed006   = lo_t_dzed_t.dzed006,
             dzed007   = lo_t_dzed_t.dzed007,
             dzed008   = lo_t_dzed_t.dzed008,
             dzed009   = lo_t_dzed_t.dzed009,
             dzed010   = lo_t_dzed_t.dzed010,
             dzedmodid = ls_user,
             dzedmoddt = ldt_datetime
       where dzed001 = lo_t_dzed_t.dzed001
         and dzed002 = lo_t_dzed_t.dzed002
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      DISPLAY cs_warning_tag,"Insert or Update DZED_T with key ",lo_t_dzed_t.dzed002," unsuccess : ",SQLCA.sqlcode
    END TRY
  END TRY

END FUNCTION

FUNCTION sadzi140_crud_insert_update_dzen_t(p_t_dzen_t)
DEFINE
  p_t_dzen_t  T_DZEN_T
DEFINE 
  lo_t_dzen_t   T_DZEN_T,
  ls_user       VARCHAR(100),
  ldt_datetime  DATETIME YEAR TO SECOND

  LET lo_t_dzen_t.* = p_t_dzen_t.*

  &ifndef DEBUG
  LET ls_user = g_user
  LET ldt_datetime = cl_get_current()
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  LET ldt_datetime = CURRENT YEAR TO SECOND
  &endif

  TRY 
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    insert into dzen_t(
                        dzen001,dzen002,dzen003,dzen004,dzen005,
                        dzen006,dzen007,dzen008,dzen009,dzen010,
                        dzen011,dzen012,dzen013,
                        dzencrtid,dzencrtdt,dzenmodid,dzenmoddt
                      )
               values (
                        lo_t_dzen_t.dzen001,lo_t_dzen_t.dzen002,lo_t_dzen_t.dzen003,lo_t_dzen_t.dzen004,lo_t_dzen_t.dzen005,
                        lo_t_dzen_t.dzen006,lo_t_dzen_t.dzen007,lo_t_dzen_t.dzen008,lo_t_dzen_t.dzen009,lo_t_dzen_t.dzen010,
                        lo_t_dzen_t.dzen011,lo_t_dzen_t.dzen012,lo_t_dzen_t.dzen013,
                        ls_user,ldt_datetime,ls_user,ldt_datetime
                      )
    
  CATCH
    TRY
      #DISPLAY "INSERT NOT SUCCESS TRY Update"
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      update dzen_t
         set dzen004  = lo_t_dzen_t.dzen004,
             dzen005  = lo_t_dzen_t.dzen005,
             dzen006  = lo_t_dzen_t.dzen006,
             dzen007  = lo_t_dzen_t.dzen007,
             dzen008  = lo_t_dzen_t.dzen008,
             dzen009  = lo_t_dzen_t.dzen009,
             dzen010  = lo_t_dzen_t.dzen010,
             dzen011  = lo_t_dzen_t.dzen011,
             dzen012  = lo_t_dzen_t.dzen012,
             dzen013  = lo_t_dzen_t.dzen013,
             dzenmodid = ls_user,
             dzenmoddt = ldt_datetime
       where dzen001  = lo_t_dzen_t.dzen001
         and dzen002  = lo_t_dzen_t.dzen002
         and dzen003  = lo_t_dzen_t.dzen003
         
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      DISPLAY cs_warning_tag,"Insert or Update DZEN_T with schema ",lo_t_dzen_t.dzen002," unsuccess : ",SQLCA.sqlcode
    END TRY
  END TRY

END FUNCTION

#從dzeb_t新增dzep_t
FUNCTION sadzi140_crud_insert_update_dzep_t(p_t_dzeb_t,p_common_fields,p_lang)
DEFINE
  p_t_dzeb_t      T_DZEB_T,
  p_common_fields DYNAMIC ARRAY WITH DIMENSION 2 OF VARCHAR(50),
  p_lang          STRING 
DEFINE 
  lo_t_dzeb_t        T_DZEB_T,
  lo_common_fields   DYNAMIC ARRAY WITH DIMENSION 2 OF VARCHAR(50),
  ls_lang            STRING, 
  lo_t_dzep_t        T_DZEP_T,
  li_index           INTEGER,
  ls_insert_sql      STRING,
  ls_upd_dzep005_sql STRING, -- 20171120 
  ls_upd_dzep012_sql STRING,
  ls_upd_dzep021_sql STRING,
  ls_upd_dzep023_sql STRING,
  li_loop            INTEGER,
  lb_fault           BOOLEAN,
  ls_dgenv           STRING,
  lo_str_buf         base.StringBuffer,
  lo_property        T_PROPERTY,
  lo_mapping_data    DYNAMIC ARRAY OF T_MAPPING_DATA

  LET lo_t_dzeb_t.* = p_t_dzeb_t.*
  LET lo_common_fields.* = p_common_fields.*
  LET ls_lang = p_lang
  LET lb_fault = FALSE

  LET ls_dgenv = FGL_GETENV("DGENV")

  CALL sadzi140_util_get_property(lo_t_dzeb_t.dzeb006,ls_lang) RETURNING lo_property.*
  #將共用欄位的查詢時開窗設定值塞入dzeb018, 再塞入 dzep018
  FOR li_index = 1 TO lo_common_fields.getLength()
    IF lo_common_fields[li_index,1] = lo_t_dzeb_t.dzeb022 THEN
      LET lo_t_dzep_t.dzep018 = lo_common_fields[li_index,3]
      EXIT FOR
    ELSE  
      LET lo_t_dzep_t.dzep018 = ""
    END IF 
  END FOR

  {
  #2015.04.15 begin
  #字串型態的資料, 存進dzep_t前先將單引號拿掉
  IF lo_t_dzeb_t.dzeb006 LIKE "C%" THEN 
    LET lo_str_buf = base.StringBuffer.create()
    CALL lo_str_buf.clear()
    CALL lo_str_buf.append(lo_t_dzeb_t.dzeb012)
    CALL lo_str_buf.replace("'","",0)
    LET lo_t_dzeb_t.dzeb012 = lo_str_buf.toString()
  END IF 
  #2015.04.15 end
  }
  
  CALL sadzi140_util_get_global_var_to_db_func_mapping_data() RETURNING lo_mapping_data
  #取得對應的全域變數名稱
  FOR li_loop = 1 TO lo_mapping_data.getLength()
    IF UPSHIFT(lo_t_dzeb_t.dzeb012) = lo_mapping_data[li_loop].md_DB_FUNCTION THEN
      LET lo_t_dzeb_t.dzeb012 = lo_mapping_data[li_loop].md_GLOBAL_VARIABLE 
      EXIT FOR 
    END IF 
  END FOR  
  
  LET lo_t_dzep_t.dzep009 = lo_property.D_WIDGET_WIDTH
  LET lo_t_dzep_t.dzep010 = lo_property.D_WIDGETS
  --LET lo_t_dzep_t.dzep015 = lo_property.D_PERCENT
  LET lo_t_dzep_t.dzep021 = lo_property.D_FORMAT
  LET lo_t_dzep_t.dzep023 = IIF(lo_t_dzeb_t.dzeb022==cs_cdf_user_define_varchar,NULL,lo_property.D_WORD_CASE) #20221017
  LET lo_t_dzep_t.dzep027 = lo_property.D_REPORT_WIDTH
  LET lo_t_dzep_t.dzep028 = lo_property.D_REPORT_DIGIT
  
  LET ls_insert_sql = "insert into dzep_t (                                                            ",
                      "                     dzep001,dzep002,dzep005,dzep009,dzep010,                   ",
                      "                     dzep012,dzep015,dzep016,dzep018,dzep021,                   ",
                      "                     dzep022,dzep023,dzep027,dzep028,                           ",
                      "                     dzepstus                                                   ",  
                      "                    )                                                           ",  
                      "            values (                                                            ",
                      "                     '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb001),"',         ", 
                      "                     '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb002),"',         ", 
                      "                     '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb005),"',         ", 
                      "                     '",sadzi140_util_trim_str(lo_t_dzep_t.dzep009),"',         ", 
                      "                     '",sadzi140_util_trim_str(lo_t_dzep_t.dzep010),"',         ", 
                      "                     ?,                                                         ", 
                      "                     '",sadzi140_util_trim_str(NVL(lo_t_dzep_t.dzep015,"Y")),"',", 
                      "                     '",sadzi140_util_trim_str(NVL(lo_t_dzep_t.dzep016,"Y")),"',", 
                      "                     '",sadzi140_util_trim_str(lo_t_dzep_t.dzep018),"',         ",
                      "                     '",sadzi140_util_trim_str(lo_t_dzep_t.dzep021),"',         ",
                      "                     '",sadzi140_util_trim_str(lo_t_dzep_t.dzep022),"',         ",
                      "                     '",sadzi140_util_trim_str(lo_t_dzep_t.dzep023),"',         ",
                      "                     '",sadzi140_util_trim_str(lo_t_dzep_t.dzep027),"',         ",
                      "                     '",sadzi140_util_trim_str(lo_t_dzep_t.dzep028),"',         ",
                      "                     '",ls_dgenv.trim(),"'                                      ",
                      "                   )                                                            "



  -- 20171120 begin                    
  #更新必要欄位
  LET ls_upd_dzep005_sql = " update dzep_t                                                      ",
                           "    set dzep005 = (                                                 ",
                           "                    select case dzep005                             ",
                           "                             when 'Y' then dzep005                  ",
                           "                           else                                     ",
                           "                             dzeb005                                ",  
                           "                           end as dzep005                           ", --decode(ep.dzep005,'Y',ep.dzep005,eb.dzeb005) dzep005 ",
                           "                      from dzeb_t                                   ",
                           "                     where dzeb001 = dzep001                        ",
                           "                       and dzeb002 = dzep002                        ",
                           "                  )                                                 ",
                           "  where dzep001 = '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb001),"' ",
                           "    and dzep002 = '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb002),"' "
  -- 20171120 end                    
                           
  #更新預設值    
  -- 20160506
  LET ls_upd_dzep012_sql = "update dzep_t                                                                                                   ",
                           "   set dzep012 = ?                                                                                              ",
                           " where dzep001 = '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb001),"'                                              ", 
                           "   and dzep002 = '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb002),"'                                              "
                           --"   AND NVL(TRIM(REPLACE(REPLACE(DZEP012,CHR(13),''),CHR(10),'')),'",cs_null_default,"') = '",cs_null_default,"' " -- 20160506 marked
                      
  #更新顯示格式        
  -- 20160506
  LET ls_upd_dzep021_sql = "update dzep_t                                                                                                   ",
                           "   set dzep021 = ?                                                                                              ",
                           " where dzep001 = '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb001),"'                                              ", 
                           "   and dzep002 = '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb002),"'                                              "
                           --"   AND NVL(TRIM(REPLACE(REPLACE(DZEP021,CHR(13),''),CHR(10),'')),'",cs_null_default,"') = '",cs_null_default,"' " -- 20160506 marked
                           
  #更新欄位大小寫        
  &ifdef is_msv
  LET ls_upd_dzep023_sql = "update dzep_t                                                                                                          ",
                           "   set dzep023 = ?                                                                                                     ",
                           " where dzep001 = '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb001),"'                                                     ", 
                           "   and dzep002 = '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb002),"'                                                     ",
                           "   and coalesce(trim(replace(replace(dzep023,char(13),''),char(10),'')),'",cs_null_default,"') = '",cs_null_default,"' "
  &else
  LET ls_upd_dzep023_sql = "update dzep_t                                                                                                        ",
                           "   set dzep023 = ?                                                                                                   ",
                           " where dzep001 = '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb001),"'                                                   ", 
                           "   and dzep002 = '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb002),"'                                                   ",
                           "   and coalesce(trim(replace(replace(dzep023,chr(13),''),chr(10),'')),'",cs_null_default,"') = '",cs_null_default,"' "
  &endif
  {                    
  LET ls_update_sql = "update dzep_t                                                                    ",
                      "   set dzep009 = '",sadzi140_util_trim_str(lo_t_dzep_t.dzep009),"',              ",
                      "       dzep010 = '",sadzi140_util_trim_str(lo_t_dzep_t.dzep010),"',              ",
                      "       dzep015 = '",sadzi140_util_trim_str(coalesce(lo_t_dzep_t.dzep015,"Y")),"',",
                      "       dzep016 = '",sadzi140_util_trim_str(coalesce(lo_t_dzep_t.dzep016,"Y")),"',",
                      "       dzep018 = '",sadzi140_util_trim_str(lo_t_dzep_t.dzep018),"'               ",
                      " where dzep001 = '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb001),"'               ", 
                      "   and dzep002 = '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb002),"'               "
  }
  
  TRY
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    PREPARE lpre_insert_dzep_t FROM ls_insert_sql
    EXECUTE lpre_insert_dzep_t USING lo_t_dzeb_t.dzeb012
  CATCH
    #DISPLAY cs_warning_tag,"Insert DZEP_T with column ",lo_t_dzeb_t.dzeb002," unsuccess." #20170425
    #更新必要欄位        
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif
    TRY    
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      PREPARE lpre_update_dzep005 FROM ls_upd_dzep005_sql
      EXECUTE lpre_update_dzep005
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      LET lb_fault = TRUE
      DISPLAY cs_warning_tag,"Update DZEP005 unsuccess."
    END TRY
    #更新預設值        
    TRY    
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      PREPARE lpre_update_dzep012 FROM ls_upd_dzep012_sql
      EXECUTE lpre_update_dzep012 USING lo_t_dzeb_t.dzeb012
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      LET lb_fault = TRUE
      DISPLAY cs_warning_tag,"Update DZEP012 unsuccess."
    END TRY
    #更新顯示格式        
    TRY    
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      PREPARE lpre_update_dzep021 FROM ls_upd_dzep021_sql
      EXECUTE lpre_update_dzep021 USING lo_t_dzep_t.dzep021
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      LET lb_fault = TRUE
      DISPLAY cs_warning_tag,"Update DZEP021 unsuccess."
    END TRY
    #更新欄位大小寫        
    TRY    
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      PREPARE lpre_update_dzep023 FROM ls_upd_dzep023_sql
      EXECUTE lpre_update_dzep023 USING lo_t_dzep_t.dzep023
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      LET lb_fault = TRUE
      DISPLAY cs_warning_tag,"Update DZEP023 unsuccess."
    END TRY
    #更新有錯誤則顯示SQL
    IF lb_fault THEN 
      DISPLAY "[Insert SQL] ",ls_insert_sql
      DISPLAY "[Update SQL] ",ls_upd_dzep012_sql
    END IF 
  END TRY  

END FUNCTION

FUNCTION sadzi140_crud_clone_dzep_t(p_t_dzep_t)
DEFINE
  p_t_dzep_t T_DZEP_T
DEFINE 
  lo_t_dzep_t T_DZEP_T,
  ls_dgenv    VARCHAR(2)

  LET ls_dgenv = FGL_GETENV("DGENV")
  
  LET lo_t_dzep_t.* = p_t_dzep_t.*

  TRY
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    insert into dzep_t(
                        dzep001,dzep002,dzep003,dzep004,dzep005,
                        dzep006,dzep007,dzep008,dzep009,dzep010,
                        dzep011,dzep012,dzep013,dzep014,dzep015,
                        dzep016,dzep017,dzep018,dzep019,dzep020,
                        dzep021,dzep022,dzep023,dzep024,dzep025,
                        dzep026,dzep027,dzep028,
                        dzepstus  
                      )
               values (
                        lo_t_dzep_t.DZEP001,lo_t_dzep_t.DZEP002,lo_t_dzep_t.DZEP003,lo_t_dzep_t.DZEP004,lo_t_dzep_t.DZEP005,
                        lo_t_dzep_t.DZEP006,lo_t_dzep_t.DZEP007,lo_t_dzep_t.DZEP008,lo_t_dzep_t.DZEP009,lo_t_dzep_t.DZEP010,
                        lo_t_dzep_t.DZEP011,lo_t_dzep_t.DZEP012,lo_t_dzep_t.DZEP013,lo_t_dzep_t.DZEP014,lo_t_dzep_t.DZEP015,
                        lo_t_dzep_t.DZEP016,lo_t_dzep_t.DZEP017,lo_t_dzep_t.DZEP018,lo_t_dzep_t.DZEP019,lo_t_dzep_t.DZEP020,
                        lo_t_dzep_t.DZEP021,lo_t_dzep_t.DZEP022,lo_t_dzep_t.DZEP023,lo_t_dzep_t.DZEP024,lo_t_dzep_t.DZEP025,
                        lo_t_dzep_t.DZEP026,lo_t_dzep_t.DZEP027,lo_t_dzep_t.DZEP028,
                        ls_dgenv
                      )
  CATCH
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif
    TRY
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      update dzep_t
         set dzep003 = lo_t_dzep_t.DZEP003,
             dzep004 = lo_t_dzep_t.DZEP004,
             dzep005 = lo_t_dzep_t.DZEP005,
             dzep006 = lo_t_dzep_t.DZEP006,
             dzep007 = lo_t_dzep_t.DZEP007,
             dzep008 = lo_t_dzep_t.DZEP008,
             dzep009 = lo_t_dzep_t.DZEP009,
             dzep010 = lo_t_dzep_t.DZEP010,
             dzep011 = lo_t_dzep_t.DZEP011,
             dzep012 = lo_t_dzep_t.DZEP012,
             dzep013 = lo_t_dzep_t.DZEP013,
             dzep014 = lo_t_dzep_t.DZEP014,
             dzep015 = lo_t_dzep_t.DZEP015,
             dzep016 = lo_t_dzep_t.DZEP016,
             dzep017 = lo_t_dzep_t.DZEP017,
             dzep018 = lo_t_dzep_t.DZEP018,
             dzep019 = lo_t_dzep_t.DZEP019,
             dzep020 = lo_t_dzep_t.DZEP020,
             dzep021 = lo_t_dzep_t.DZEP021,
             dzep022 = lo_t_dzep_t.DZEP022,
             dzep023 = lo_t_dzep_t.DZEP023,
             dzep024 = lo_t_dzep_t.DZEP024,
             dzep025 = lo_t_dzep_t.DZEP025,
             dzep026 = lo_t_dzep_t.DZEP026, 
             dzep027 = lo_t_dzep_t.DZEP027,
             dzep028 = lo_t_dzep_t.DZEP028 
       where dzep001 = lo_t_dzep_t.DZEP001
         and dzep002 = lo_t_dzep_t.DZEP002
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      DISPLAY "INSERT or Update dzep error"
    END TRY     
  END TRY   

END FUNCTION

FUNCTION sadzi140_crud_insert_update_dzef_t(p_t_dzeb_t,p_common_fields)
DEFINE
  p_t_dzeb_t T_DZEB_T,
  p_common_fields DYNAMIC ARRAY WITH DIMENSION 2 OF VARCHAR(50)
DEFINE 
  lo_t_dzeb_t   T_DZEB_T,
  lo_common_fields DYNAMIC ARRAY WITH DIMENSION 2 OF VARCHAR(50),
  ls_insert_sql STRING,
  ls_update_sql STRING,
  ls_pre_name   STRING,
  ls_dgenv      STRING,
  li_index      INTEGER

  LET lo_t_dzeb_t.* = p_t_dzeb_t.*
  LET lo_common_fields.* = p_common_fields.*

  LET ls_dgenv = FGL_GETENV("DGENV")
  
  #取得表格的前置名稱
  CALL sadzi140_util_get_table_pre_name(lo_t_dzeb_t.dzeb001) RETURNING ls_pre_name

  #將共用欄位的查設定值塞入
  FOR li_index = 1 TO lo_common_fields.getLength()
    IF lo_common_fields[li_index,1] = lo_t_dzeb_t.dzeb022 THEN
    
      LET ls_insert_sql = "insert into dzef_t (                                                             ",
                          "                     dzef001,dzef002,dzef003,dzef006,dzef007,                    ",
                          "                     dzef008,dzef009,                                            ",
                          "                     dzefstus                                                    ",
                          "                    )                                                            ",  
                          "            values (                                                             ",
                          "                     '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb001),"',          ", 
                          "                     '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb002),"',          ", 
                          "                     replace('",sadzi140_util_trim_str(lo_common_fields[li_index,4]),"','*','",ls_pre_name,"'),", 
                          "                     '",sadzi140_util_trim_str(lo_common_fields[li_index,5]),"', ", 
                          "                     '",sadzi140_util_trim_str(lo_common_fields[li_index,6]),"', ", 
                          "                     '",sadzi140_util_trim_str(lo_common_fields[li_index,8]),"', ", 
                          "                     '",sadzi140_util_trim_str(lo_common_fields[li_index,7]),"', ",
                          "                     '",ls_dgenv.trim(),"'                                       ",
                          "                   )                                                             "

      {                    
      LET ls_update_sql = "update dzef_t                                                              ",
                          "   set dzef003 = replace('",sadzi140_util_trim_str(lo_common_fields[li_index,4]),"','*','",ls_pre_name,"'),",
                          "       dzef006 = '",sadzi140_util_trim_str(lo_common_fields[li_index,5]),"',",
                          "       dzef007 = '",sadzi140_util_trim_str(lo_common_fields[li_index,6]),"',",
                          "       dzef008 = '",sadzi140_util_trim_str(lo_common_fields[li_index,8]),"',",
                          "       dzef009 = '",sadzi140_util_trim_str(lo_common_fields[li_index,7]),"' ",
                          " where dzef001 = '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb001),"'         ", 
                          "   and dzef002 = '",sadzi140_util_trim_str(lo_t_dzeb_t.dzeb002),"'         "
      }                   
      TRY
        &ifdef is_pgs
        SAVEPOINT sp_tran
        &endif
        PREPARE lpre_insert_dzef_t FROM ls_insert_sql
        EXECUTE lpre_insert_dzef_t
      CATCH
        &ifdef is_pgs
        ROLLBACK WORK TO SAVEPOINT sp_tran
        &endif
        #DISPLAY cs_warning_tag,"Insert DZEF_T unsuccess."
        {
        TRY    
          &ifdef is_pgs
          SAVEPOINT sp_tran
          &endif
          PREPARE lpre_update_dzef_t FROM ls_update_sql
          EXECUTE lpre_update_dzef_t
        CATCH
          &ifdef is_pgs
          ROLLBACK WORK TO SAVEPOINT sp_tran
          &endif
          DISPLAY cs_warning_tag,"Insert or Update DZEF_T unsuccess."
          DISPLAY "[Insert SQL] ",ls_insert_sql
          DISPLAY "[Update SQL] ",ls_update_sql
        END TRY
        }
      END TRY  
      EXIT FOR
    END IF 
  END FOR

END FUNCTION

FUNCTION sadzi140_crud_clone_dzef_t(p_t_DZEF_t)
DEFINE
  p_t_dzef_t T_DZEF_T
DEFINE 
  lo_t_dzef_t T_DZEF_T,
  ls_dgenv    VARCHAR(2)

  LET lo_t_dzef_t.* = p_t_dzef_t.*

  LET ls_dgenv = FGL_GETENV("DGENV")
  
  TRY
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    insert into dzef_t(
                        dzef001,dzef002,dzef003,dzef004,dzef005,
                        dzef006,dzef007,dzef008,dzef009,dzef010,
                        dzefstus
                      )
               values (
                        lo_t_DZEF_t.DZEF001,lo_t_DZEF_t.DZEF002,lo_t_DZEF_t.DZEF003,lo_t_DZEF_t.DZEF004,lo_t_DZEF_t.DZEF005,
                        lo_t_DZEF_t.DZEF006,lo_t_DZEF_t.DZEF007,lo_t_DZEF_t.DZEF008,lo_t_DZEF_t.DZEF009,lo_t_DZEF_t.DZEF010,
                        ls_dgenv
                      )
  CATCH
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif

    TRY
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      update dzef_t
         set dzef003 = lo_t_dzef_t.DZEF003,
             dzef004 = lo_t_dzef_t.DZEF004,
             dzef005 = lo_t_dzef_t.DZEF005,
             dzef006 = lo_t_dzef_t.DZEF006,
             dzef007 = lo_t_dzef_t.DZEF007,
             dzef008 = lo_t_dzef_t.DZEF008,
             dzef009 = lo_t_dzef_t.DZEF009,
             dzef010 = lo_t_dzef_t.DZEF010
       where dzef001 = lo_t_dzef_t.DZEF001
         and dzef002 = lo_t_dzef_t.DZEF002
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      DISPLAY "INSERT or Update DZEF error"
    END TRY     
  END TRY   

END FUNCTION

FUNCTION sadzi140_crud_insert_update_gzcc_t(p_dzeb_t)
DEFINE
  p_dzeb_t  T_DZEB_T
DEFINE 
  lo_dzeb_t      T_DZEB_T,
  lo_gzcc_t      DYNAMIC ARRAY OF T_GZCC_T,
  li_loop        INTEGER,
  ls_column_name STRING

  LET lo_dzeb_t.* = p_dzeb_t.* 

  LET ls_column_name = lo_dzeb_t.dzeb002
  LET ls_column_name = ls_column_name.toLowerCase()

  IF (ls_column_name.getIndexOf('stus',1) > 0) THEN
    IF NOT (sadzi140_crud_get_status_code_data_exists(lo_dzeb_t.dzeb001)) THEN

      #若不存在, 則由 gzcb_t 抓取對應的代碼清單新增
      CALL sadzi140_crud_get_status_list(lo_dzeb_t.dzeb001) RETURNING lo_gzcc_t

      FOR li_loop = 1 TO lo_gzcc_t.getLength() 
        TRY 
          &ifdef is_pgs
          SAVEPOINT sp_tran
          &endif
          #只新增不更新
          insert into gzcc_t(
                              gzcc001, gzcc002, gzcc003, gzcc004, gzcc005,
                              gzcc006, gzccstus
                            )
                     VALUES (
                              lo_gzcc_t[li_loop].gzcc001, lo_gzcc_t[li_loop].gzcc002, lo_gzcc_t[li_loop].gzcc003, lo_gzcc_t[li_loop].gzcc004, lo_gzcc_t[li_loop].gzcc005,
                              lo_gzcc_t[li_loop].gzcc006, lo_gzcc_t[li_loop].gzccstus                                
                            )
                             
        CATCH
          &ifdef is_pgs
          ROLLBACK WORK TO SAVEPOINT sp_tran
          &endif
          #DISPLAY cs_warning_tag,"Insert GZCC_T unsuccess."
        END TRY
      END FOR
    END IF  
  END IF  

END FUNCTION

#檢核狀態碼表格資料是否存在
FUNCTION sadzi140_crud_get_status_code_data_exists(p_table)
DEFINE
  p_table  STRING
DEFINE
  lb_return     BOOLEAN,
  ls_sql        STRING,
  li_rec_count  INTEGER,
  ls_table_name STRING

  LET ls_table_name = p_table

  #取得資料筆數
  LET ls_sql = "select count(1) rec_cnt                 ", 
               "  from gzcc_t cc                        ", 
               " where cc.gzcc001 = '",ls_table_name,"' " 

  PREPARE lpre_get_status_code_count FROM ls_sql
  DECLARE lcur_get_status_code_count CURSOR FOR lpre_get_status_code_count
  OPEN lcur_get_status_code_count
  FETCH lcur_get_status_code_count INTO li_rec_count
  CLOSE lcur_get_status_code_count
  FREE lcur_get_status_code_count
  FREE lpre_get_status_code_count

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_crud_clone_gzcc_t(p_t_gzcc_t)
DEFINE
  p_t_gzcc_t T_GZCC_T
DEFINE 
  lo_t_gzcc_t T_GZCC_T

  LET lo_t_gzcc_t.* = p_t_gzcc_t.*

  TRY
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    insert into gzcc_t(
                        gzcc001,gzcc002,gzcc003,gzcc004,gzcc005,
                        gzcc006,gzccstus
                      )
               values (
                        lo_t_gzcc_t.GZCC001,lo_t_gzcc_t.GZCC002,lo_t_gzcc_t.GZCC003,lo_t_gzcc_t.GZCC004,lo_t_gzcc_t.GZCC005,
                        lo_t_gzcc_t.GZCC006,lo_t_gzcc_t.GZCCSTUS
                      )
  CATCH
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif
    TRY 
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      update gzcc_t
         set gzcc002  = lo_t_gzcc_t.gzcc002, 
             gzcc003  = lo_t_gzcc_t.gzcc003, 
             gzcc005  = lo_t_gzcc_t.gzcc005, 
             gzcc006  = lo_t_gzcc_t.gzcc006, 
             gzccstus = lo_t_gzcc_t.gzccstus
       where gzcc001  = lo_t_gzcc_t.gzcc001 
         and gzcc004  = lo_t_gzcc_t.gzcc004
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      DISPLAY "INSERT or Update GZCC_T error"
    END TRY    
  END TRY   

END FUNCTION

FUNCTION sadzi140_crud_insert_update_dzer_t(p_t_dzer_t)
DEFINE
  p_t_dzer_t T_DZER_T
DEFINE 
  lo_t_dzer_t T_DZER_T,
  ls_dgenv    VARCHAR(2)

  LET lo_t_dzer_t.* = p_t_dzer_t.*

  LET ls_dgenv = FGL_GETENV("DGENV")
  
  TRY 
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    insert into dzer_t(
                        dzer001,dzer002,dzer003,dzer004,dzer005,
                        dzer006,dzer007,dzer008,
                        dzerstus
                      )
               values (
                        lo_t_dzer_t.DZER001,lo_t_dzer_t.DZER002,lo_t_dzer_t.DZER003,lo_t_dzer_t.DZER004,lo_t_dzer_t.DZER005,
                        lo_t_dzer_t.DZER006,lo_t_dzer_t.DZER007,lo_t_dzer_t.DZER008,
                        ls_dgenv
                      )
  CATCH 
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif
    DISPLAY "INSERT or Update DZER_T error"
  END TRY   

END FUNCTION

FUNCTION sadzi140_crud_insert_update_dzeq_t(p_t_dzeq_t)
DEFINE
  p_t_dzeq_t T_DZEQ_T
DEFINE 
  lo_t_dzeq_t T_DZEQ_T,
  ls_dgenv    VARCHAR(2)

  LET lo_t_dzeq_t.* = p_t_dzeq_t.*

  LET ls_dgenv = FGL_GETENV("DGENV")

  TRY
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    insert into dzeq_t(
                        dzeq001,dzeq002,dzeq003,dzeq004,dzeq005,
                        dzeq006,dzeq007,dzeq008,dzeq009,
                        dzeqstus
                      )
               values (
                        lo_t_dzeq_t.dzeq001,lo_t_dzeq_t.dzeq002,lo_t_dzeq_t.dzeq003,lo_t_dzeq_t.dzeq004,lo_t_dzeq_t.dzeq005,
                        lo_t_dzeq_t.dzeq006,lo_t_dzeq_t.dzeq007,lo_t_dzeq_t.dzeq008,lo_t_dzeq_t.dzeq009,
                        ls_dgenv
                      )
  CATCH
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif
    DISPLAY "INSERT or Update DZEQ_T error"
  END TRY   

END FUNCTION

FUNCTION sadzi140_crud_insert_update_dzhh_t(p_t_dzhh_t)
DEFINE
  p_t_dzhh_t  T_DZHH_T
DEFINE 
  lo_t_dzhh_t   T_DZHH_T,
  ls_user       VARCHAR(100),
  ldt_datetime  DATETIME YEAR TO SECOND

  LET lo_t_dzhh_t.* = p_t_dzhh_t.*

  &ifndef DEBUG
  LET ls_user = g_user
  LET ldt_datetime = cl_get_current()
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  LET ldt_datetime = CURRENT YEAR TO SECOND
  &endif

  TRY 
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    insert into dzhh_t(
                        dzhh001,dzhh002,dzhh003,dzhh004,dzhh005,
                        dzhhcrtid,dzhhcrtdt,dzhhmodid,dzhhmoddt
                      )
               values (
                        lo_t_dzhh_t.dzhh001,lo_t_dzhh_t.dzhh002,lo_t_dzhh_t.dzhh003,lo_t_dzhh_t.dzhh004,lo_t_dzhh_t.dzhh005,
                        ls_user,ldt_datetime,ls_user,ldt_datetime
                      )
                       
  CATCH
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif
    TRY
      #DISPLAY "INSERT NOT SUCCESS TRY Update"
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      update dzhh_t                                 
         set dzhh003   = lo_t_dzhh_t.dzhh003,
             dzhh004   = lo_t_dzhh_t.dzhh004,
             dzhh005   = lo_t_dzhh_t.dzhh005,
             dzhhmodid = ls_user,
             dzhhmoddt = ldt_datetime
       where dzhh001   = lo_t_dzhh_t.dzhh001
         and dzhh002   = lo_t_dzhh_t.dzhh002
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      DISPLAY cs_warning_tag,"Insert or Update DZHH_T unsuccess : ",SQLCA.sqlcode
    END TRY
  END TRY

END FUNCTION

FUNCTION sadzi140_crud_get_status_list(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name STRING,
  lo_gzcc_t     DYNAMIC ARRAY OF T_GZCC_T
DEFINE  
  ls_sql      STRING,
  li_rec_cnt  INTEGER,
  lo_return   DYNAMIC ARRAY OF T_GZCC_T

  LET ls_table_name = p_table_name

  {
  LET ls_sql = "SELECT EB.DZEB001                     GZCC001,                                 ",
               "       EB.DZEB002                     GZCC002,                                 ",
               "       DECODE(EA.DZEA004,'B',17,'P',17,'M',50,'H',13,'T',13,NULL)   GZCC003,   ",
               "       CB.GZCB002                     GZCC004,                                 ",
               "       's'                            GZCC005,                                 ",
               "       ROWNUM                         GZCC006,                                 ",
               "       DECODE(EA.DZEA004,'B','Y','P','Y','M','Y','H','Y','T','Y','N') GZCCSTUS ",
               "  FROM DZEB_T EB,DZEA_T EA,                                                    ",
               "       GZCB_T CB                                                               ",
               " WHERE 1=1                                                                     ",
               "   AND EA.DZEA001 = '",ls_table_name.toLowercase(),"'                          ",
               "   AND EA.DZEA001 = EB.DZEB001                                                 ",
               "   AND EB.DZEB022 = 'cdfStatus'                                                ",
               "   AND CB.GZCB001 = DECODE(EA.DZEA004,'B',17,'P',17,'M',50,'H',13,'T',13,NULL) ",
               "   AND CB.GZCB003 = 'Y'                                                        ", 
               " ORDER BY EB.DZEB001,CB.GZCB002                                                "
  }
  LET ls_sql = "select eb.dzeb001                             gzcc001,  ",                                
               "       eb.dzeb002                             gzcc002,  ",                                
               "       case ea.dzea004                                  ", 
               "         when 'B' then 17                               ", 
               "         when 'P' then 17                               ", 
               "         when 'M' then 50                               ", 
               "         when 'H' then 13                               ", 
               "         when 'T' then 13                               ", 
               "       else                                             ", 
               "         null                                           ", 
               "       end                                    gzcc003,  ",  
               "       cb.gzcb002                             gzcc004,  ",                                
               "       's'                                    gzcc005,  ",                                
               "       ROW_NUMBER()                                     ", 
               "         OVER(ORDER BY eb.dzeb001,cb.gzcb002) gzcc006,  ",                                
               "       case ea.dzea004                                  ", 
               "         when 'B' then 'Y'                              ", 
               "         when 'P' then 'Y'                              ", 
               "         when 'M' then 'Y'                              ", 
               "         when 'H' then 'Y'                              ", 
               "         when 'T' then 'Y'                              ", 
               "       else                                             ", 
               "         'N'                                            ", 
               "       end                                    gzccstus  ", 
               "  from dzeb_t eb,dzea_t ea,                             ",                        
               "       gzcb_t cb                                        ",                        
               " where 1=1                                              ",                        
               "   and ea.dzea001 = '",ls_table_name.toLowercase(),"'   ", 
               "   and ea.dzea001 = eb.dzeb001                          ",                        
               "   and eb.dzeb022 = 'cdfStatus'                         ",                        
               "   and cb.gzcb001 = case ea.dzea004                     ", 
               "                      when 'B' then 17                  ", 
               "                      when 'P' then 17                  ", 
               "                      when 'M' then 50                  ", 
               "                      when 'H' then 13                  ", 
               "                      when 'T' then 13                  ", 
               "                    else                                ", 
               "                      null                              ", 
               "                    end                                 ", 
               "   and cb.gzcb003 = 'Y'                                 ",                        
               " order by eb.dzeb001,cb.gzcb002                         "                  
  
  PREPARE lpre_get_status_list FROM ls_sql
  DECLARE lcur_get_status_list CURSOR FOR lpre_get_status_list

  LET li_rec_cnt = 1 
  
  OPEN lcur_get_status_list
  FOREACH lcur_get_status_list INTO lo_gzcc_t[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF
    
    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_status_list
  IF li_rec_cnt > 1 THEN
    CALL lo_gzcc_t.deleteElement(li_rec_cnt)
  END IF  
    
  FREE lpre_get_status_list
  FREE lcur_get_status_list  

  LET lo_return.* = lo_gzcc_t.*
  
  RETURN lo_return
  
END FUNCTION

FUNCTION sadzi140_crud_insert_update_dzeal_t(p_create_table,p_lang)
DEFINE
  p_create_table T_DZEA_CREATE_TABLE,
  p_lang         STRING 
DEFINE
  lo_create_table T_DZEA_CREATE_TABLE,
  ls_lang         VARCHAR(10), 
  ls_orig_lang    STRING, 
  ls_insert_sql   STRING,
  ls_update_sql   STRING,
  ls_lang_content VARCHAR(1024),
  li_loop         INTEGER,
  lo_lang_arr     DYNAMIC ARRAY OF T_LANGUAGE_TYPE

  LET lo_create_table.* = p_create_table.*
  LET ls_lang      = p_lang
  LET ls_orig_lang = p_lang
    
  LET ls_lang_content = lo_create_table.dct_table_description

  &ifndef DEBUG
  CALL sadzi140_crud_get_static_lang_list() RETURNING lo_lang_arr

  FOR li_loop = 1 TO lo_lang_arr.getLength()
    LET ls_lang = lo_lang_arr[li_loop]
    IF ls_lang <> ls_orig_lang THEN 
      TRY 
        IF (ls_lang = cs_lang_zh_tw) OR (ls_lang = cs_lang_zh_cn) THEN #20170706 add
          CALL cl_trans_code_tw_cn(ls_lang,lo_create_table.dct_table_description) RETURNING ls_lang_content
        ELSE #20170706 add begin
          LET ls_lang_content = ""
        END IF #20170706 add end 
      CATCH
        DISPLAY cs_error_tag,"Translate language data '",ls_lang,"' error."
        LET ls_lang_content = lo_create_table.dct_table_description
      END TRY
    ELSE   
      LET ls_lang_content = lo_create_table.dct_table_description
    END IF     
  &endif
    LET ls_insert_sql = "insert into dzeal_t(dzeal001,dzeal002,dzeal003) ",  
                        " values('",lo_create_table.dct_table_name,"','",ls_lang,"','",ls_lang_content,"')"
  
    LET ls_update_sql = "update dzeal_t                                        ",
                        "   set dzeal003 = '",ls_lang_content,"'               ",
                        " where dzeal001 = '",lo_create_table.dct_table_name,"'", 
                        "   and dzeal002 = '",ls_lang,"'                       "
                       
    TRY
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      PREPARE lpre_insert_dzeal_t FROM ls_insert_sql
      EXECUTE lpre_insert_dzeal_t
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      DISPLAY cs_warning_tag,"Insert DZEAL_T with table ",lo_create_table.dct_table_name," unsuccess."
      #20160705 如果語系相同才更新多語言的內容
      IF ls_lang = ls_orig_lang THEN
        TRY    
          &ifdef is_pgs
          SAVEPOINT sp_tran
          &endif
          PREPARE lpre_update_dzeal_t FROM ls_update_sql
          EXECUTE lpre_update_dzeal_t
        CATCH
          &ifdef is_pgs
          ROLLBACK WORK TO SAVEPOINT sp_tran
          &endif
          DISPLAY cs_warning_tag,"Insert or Update DZEAL_T with table ",lo_create_table.dct_table_name," unsuccess."
          DISPLAY "[Insert SQL] ",ls_insert_sql
          DISPLAY "[Update SQL] ",ls_update_sql
        END TRY
      END IF  
    END TRY
    
  &ifndef DEBUG
  END FOR    
  &endif

END FUNCTION

FUNCTION sadzi140_crud_insert_update_dzebl_t(p_t_dzeb_t,p_lang)
DEFINE
  p_t_dzeb_t T_DZEB_T,
  p_lang     STRING  
DEFINE 
  lo_t_dzeb_t     T_DZEB_T,
  ls_lang         STRING,  
  ls_orig_lang    STRING, 
  ls_insert_sql   STRING,
  ls_update_sql   STRING,
  ls_lang_content STRING,
  ls_lang_memo    STRING,
  li_loop         INTEGER,
  lo_lang_arr     DYNAMIC ARRAY OF T_LANGUAGE_TYPE

  LET lo_t_dzeb_t.* = p_t_dzeb_t.*
  LET ls_lang = p_lang
  LET ls_orig_lang = p_lang

  LET ls_lang_content = lo_t_dzeb_t.DZEB003
  LET ls_lang_memo    = lo_t_dzeb_t.DZEB024

  #如果多語系沒有修改就不用更新
  IF lo_t_dzeb_t.lang_modified <> cs_lang_modified THEN RETURN END IF
  
  &ifndef DEBUG
  CALL sadzi140_crud_get_static_lang_list() RETURNING lo_lang_arr

  FOR li_loop = 1 TO lo_lang_arr.getLength()
    LET ls_lang = lo_lang_arr[li_loop]
    IF (ls_lang <> ls_orig_lang) THEN
      TRY 
        IF (ls_lang = cs_lang_zh_tw) OR (ls_lang = cs_lang_zh_cn) THEN #20170706 add
          CALL cl_trans_code_tw_cn(ls_lang,lo_t_dzeb_t.DZEB003) RETURNING ls_lang_content
          CALL cl_trans_code_tw_cn(ls_lang,lo_t_dzeb_t.DZEB024) RETURNING ls_lang_memo
        ELSE   #20170706 begin 
          LET ls_lang_content = ""
          LET ls_lang_memo    = ""
        END IF #20170706 end 
      CATCH
        DISPLAY cs_error_tag,"Translate language data '",ls_lang,"' error."
        LET ls_lang_content = lo_t_dzeb_t.DZEB003
        LET ls_lang_memo    = lo_t_dzeb_t.DZEB024
      END TRY
    ELSE  
      LET ls_lang_content = lo_t_dzeb_t.DZEB003
      LET ls_lang_memo    = lo_t_dzeb_t.DZEB024
    END IF     
  &endif
    
    LET ls_insert_sql = "insert into dzebl_t (dzebl001,dzebl002,dzebl003,dzebl004,dzebl000)                                                                ",  
                        "            values ('",lo_t_dzeb_t.DZEB002 ,"','",ls_lang,"','",ls_lang_content,"','",ls_lang_memo,"','",lo_t_dzeb_t.DZEB001 ,"') "
                        
    LET ls_update_sql = "update dzebl_t                              ",
                        "   set dzebl003 = '",ls_lang_content,"',    ",
                        "       dzebl004 = '",ls_lang_memo,"',       ",
                        "       dzebl000 = '",lo_t_dzeb_t.DZEB001,"' ", 
                        " where dzebl001 = '",lo_t_dzeb_t.DZEB002,"' ", 
                        "   and dzebl002 = '",ls_lang,"'             "
                       
    TRY
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      PREPARE lpre_insert_dzebl_t FROM ls_insert_sql
      EXECUTE lpre_insert_dzebl_t
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      #DISPLAY cs_warning_tag,"Insert DZEBL_T with column ",lo_t_dzeb_t.DZEB002," unsuccess." #20170425

      #20160705 如果語系相同才更新多語言的內容
      IF ls_lang = ls_orig_lang THEN
        TRY    
          &ifdef is_pgs
          SAVEPOINT sp_tran
          &endif
          PREPARE lpre_update_dzebl_t FROM ls_update_sql
          EXECUTE lpre_update_dzebl_t
        CATCH
          &ifdef is_pgs
          ROLLBACK WORK TO SAVEPOINT sp_tran
          &endif
          DISPLAY cs_warning_tag,"Insert or Update DZEBL_T with column ",lo_t_dzeb_t.DZEB002," unsuccess."
          DISPLAY "[Insert SQL] ",ls_insert_sql
          DISPLAY "[Update SQL] ",ls_update_sql
        END TRY
      END IF  
      
    END TRY  

  &ifndef DEBUG
  END FOR  
  &endif

END FUNCTION

FUNCTION sadzi140_crud_insert_update_dzeu_t(p_table_type)
DEFINE
  p_table_type  T_TABLE_TYPE 
DEFINE
  lo_table_type  T_TABLE_TYPE, 
  ls_insert_sql  STRING,
  ls_update_sql  STRING

  LET lo_table_type.* = p_table_type.*
    
  LET ls_insert_sql = "insert into dzeu_t (dzeu001,dzeu002,dzeu003,dzeu004,dzeu005) ",  
                      "            values (                                         ",
                      "                    '",lo_table_type.DZEU001,"',             ",
                      "                    '",lo_table_type.DZEU002,"',             ",
                      "                    '",lo_table_type.DZEU003,"',             ",
                      "                    '",lo_table_type.DZEU004,"',             ",
                      "                     ",lo_table_type.DZEU005,"               ",
                      "                    )                                        "

  LET ls_update_sql = "update dzeu_t                                ",
                      "   set dzeu003 = '",lo_table_type.DZEU003,"',",
                      "       dzeu004 = '",lo_table_type.DZEU004,"',",
                      "       dzeu005 =  ",lo_table_type.DZEU005,"  ",
                      " where dzeu001 = '",lo_table_type.DZEU001,"' ", 
                      "   and dzeu002 = '",lo_table_type.DZEU002,"' "
                     
  TRY
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    PREPARE lpre_insert_dzeu_t FROM ls_insert_sql
    EXECUTE lpre_insert_dzeu_t
  CATCH
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif
    TRY    
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      PREPARE lpre_update_dzeu_t FROM ls_update_sql
      EXECUTE lpre_update_dzeu_t
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      DISPLAY cs_warning_tag,"Insert or Update dzeu_T with table ",lo_table_type.DZEU001," unsuccess."
      DISPLAY "[Insert SQL] ",ls_insert_sql
      DISPLAY "[Update SQL] ",ls_update_sql
    END TRY
    
  END TRY  

END FUNCTION

FUNCTION sadzi140_crud_insert_update_dzey_t(p_dzey_t)
DEFINE
  p_dzey_t  T_DZEY_T 
DEFINE
  lo_dzey_t      T_DZEY_T, 
  ldt_today      DATETIME YEAR TO DAY,
  ls_insert_sql  STRING,
  ls_update_sql  STRING

  LET lo_dzey_t.* = p_dzey_t.*

  LET ldt_today = CURRENT YEAR TO DAY
  LET lo_dzey_t.DZEY001 = ldt_today
  
  LET ls_insert_sql = "insert into dzey_t (dzey001,dzey002)          ",  
                      "            values (                          ",
                      "                     ?,                       ",
                      "                     '",lo_dzey_t.DZEY002,"'  ",
                      "                   )                          "

  {                       
  LET ls_update_sql = "update dzey_t                            ",
                      "   set dzey001 = '",lo_dzey_t.DZEY001,"',",
                      "       dzey002 = '",lo_dzey_t.DZEY002,"' ",
                      " where dzey001 = '",lo_dzey_t.DZEY001,"' ", 
                      "   and dzey002 = '",lo_dzey_t.DZEY002,"' "
  }
  
  TRY
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    PREPARE lpre_insert_dzey_t FROM ls_insert_sql
    EXECUTE lpre_insert_dzey_t USING lo_dzey_t.DZEY001
  CATCH
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif
    DISPLAY cs_warning_tag,"Insert DZEY_T with table ",lo_dzey_t.DZEY002," unsuccess."
    {
    TRY    
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      PREPARE lpre_update_dzey_t FROM ls_update_sql
      EXECUTE lpre_update_dzey_t
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      DISPLAY cs_warning_tag,"Insert or Update DZEY_T with table ",lo_dzey_t.DZEY002," unsuccess."
      DISPLAY "[Insert SQL] ",ls_insert_sql
      DISPLAY "[Update SQL] ",ls_update_sql
    END TRY
    }
  END TRY  

END FUNCTION

FUNCTION sadzi140_crud_get_static_lang_list()
DEFINE
  lo_lang_arr DYNAMIC ARRAY OF T_LANGUAGE_TYPE
DEFINE
  lo_return      DYNAMIC ARRAY OF T_LANGUAGE_TYPE,
  ls_sql         STRING, #20170214
  li_count       INTEGER, #20170214
  ls_DGENV       STRING, #20170214
  ls_full_locale STRING #20230821

  #設定靜態區域別
  LET lo_lang_arr[1] = cs_lang_zh_cn
  LET lo_lang_arr[2] = cs_lang_zh_tw

  #20170214 begin
  LET ls_DGENV = FGL_GETENV("DGENV")
  LET ls_full_locale = NVL(FGL_GETENV(cs_env_full_locale),"N/A")  #20230821 
  
  #如果是在客制環境就抓除了簡繁中文外的其他語言資料
  IF (ls_DGENV = cs_dgenv_customize) OR (ls_full_locale = "Y") THEN  #20230821
  
    LET ls_sql = "select gzzy001                                                  ", 
                 "  from gzzy_t                                                   ", 
                 " where gzzy001 not in ('",cs_lang_zh_cn,"','",cs_lang_zh_tw,"') ",
                 "   and gzzystus = 'Y'                                           ", --20230807
                 " order by gzzy001                                               "
                 
    PREPARE lpre_get_static_lang_list FROM ls_sql
    DECLARE lcur_get_static_lang_list CURSOR FOR lpre_get_static_lang_list

    LET li_count = lo_lang_arr.getLength() + 1
    
    OPEN lcur_get_static_lang_list
    FOREACH lcur_get_static_lang_list INTO lo_lang_arr[li_count]  
      IF SQLCA.sqlcode THEN
        EXIT FOREACH
      END IF

      LET li_count = li_count + 1

    END FOREACH
    CLOSE lcur_get_static_lang_list

    FREE lcur_get_static_lang_list
    FREE lpre_get_static_lang_list
    
    CALL lo_lang_arr.deleteElement(li_count)
    
  END IF
  #20170214 end
  
  LET lo_return = lo_lang_arr

  RETURN lo_return

END FUNCTION 

#20180302 begin
FUNCTION sadzi140_crud_insert_update_dzhj_t(p_t_dzhj_t)
DEFINE
  p_t_dzhj_t  T_DZHJ_T
DEFINE 
  lo_t_dzhj_t   T_DZHJ_T,
  ls_user       VARCHAR(100),
  ldt_datetime  DATETIME YEAR TO SECOND,
  lb_success    BOOLEAN  
DEFINE
  lb_return BOOLEAN  

  LET lo_t_dzhj_t.* = p_t_dzhj_t.*

  LET lb_success = TRUE
  LET lb_return  = TRUE
  
  &ifndef DEBUG
  LET ls_user = g_user
  LET ldt_datetime = cl_get_current()
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  LET ldt_datetime = CURRENT YEAR TO SECOND
  &endif

  TRY 
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    insert into dzhj_t(
                        dzhj001,dzhj002,dzhj003,dzhj004,dzhj005,
                        dzhjcrtid,dzhjcrtdt,dzhjmodid,dzhjmoddt
                      )
               values (
                        lo_t_dzhj_t.dzhj001,lo_t_dzhj_t.dzhj002,lo_t_dzhj_t.dzhj003,lo_t_dzhj_t.dzhj004,lo_t_dzhj_t.dzhj005,
                        ls_user,ldt_datetime,ls_user,ldt_datetime
                      )
                       
  CATCH
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif
    TRY
      #DISPLAY "INSERT NOT SUCCESS TRY Update"
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      update dzhj_t                                 
         set dzhj003   = lo_t_dzhj_t.dzhj003,
             dzhj004   = lo_t_dzhj_t.dzhj004,
             dzhj005   = lo_t_dzhj_t.dzhj005,
             dzhjmodid = ls_user,
             dzhjmoddt = ldt_datetime
       where dzhj001 = lo_t_dzhj_t.dzhj001
         and dzhj002 = lo_t_dzhj_t.dzhj002
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      LET lb_success = FALSE
      DISPLAY cs_warning_tag,"Insert or Update DZHJ_T for ",lo_t_dzhj_t.dzhj001,".",lo_t_dzhj_t.dzhj002," unsuccess : ",SQLCA.sqlcode
    END TRY
  END TRY

  LET lb_return = lb_success
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_crud_delete_dzhj_t(p_t_dzhj_t)
DEFINE
  p_t_dzhj_t  T_DZHJ_T
DEFINE 
  lo_t_dzhj_t   T_DZHJ_T,
  ls_user       VARCHAR(100),
  lb_success    BOOLEAN,  
  ldt_datetime  DATETIME YEAR TO SECOND
DEFINE
  lb_return BOOLEAN  

  LET lo_t_dzhj_t.* = p_t_dzhj_t.*

  LET lb_success = TRUE
  LET lb_return  = TRUE

  &ifndef DEBUG
  LET ls_user = g_user
  LET ldt_datetime = cl_get_current()
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  LET ldt_datetime = CURRENT YEAR TO SECOND
  &endif

  TRY
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    
    delete from dzhj_t 
     where dzhj001 = lo_t_dzhj_t.dzhj001
       and dzhj002 = lo_t_dzhj_t.dzhj002
       
  CATCH
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif
    LET lb_success = FALSE
    DISPLAY cs_warning_tag,"DELETE DZHJ_T for ",lo_t_dzhj_t.dzhj001,".",lo_t_dzhj_t.dzhj002," unsuccess : ",SQLCA.sqlcode
  END TRY

  LET lb_return = lb_success
  
  RETURN lb_return
  
END FUNCTION
#20180302 end

#20210607 begin
FUNCTION sadzi140_crud_clone_dzebl_t(p_multi_lang_data)
DEFINE
  p_multi_lang_data  DYNAMIC ARRAY OF T_DZEBL_ORIG_DATA
DEFINE 
  lo_multi_lang_data  DYNAMIC ARRAY OF T_DZEBL_ORIG_DATA,
  ls_insert_sql       STRING,
  ls_update_sql       STRING,
  li_loop             INTEGER

  LET lo_multi_lang_data = p_multi_lang_data

  FOR li_loop = 1 TO lo_multi_lang_data.getLength()
    
    LET ls_insert_sql = "insert into dzebl_t (dzebl001,dzebl002,dzebl003,dzebl004,dzebl000) values (?,?,?,?,?) "
                        
    LET ls_update_sql = "update dzebl_t       ",
                        "   set dzebl003 = ?, ",
                        "       dzebl004 = ?, ",
                        "       dzebl000 = ?  ", 
                        " where dzebl001 = ?  ", 
                        "   and dzebl002 = ?  "
                       
    TRY
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      PREPARE lpre_clone_ins_dzebl_t FROM ls_insert_sql
      EXECUTE lpre_clone_ins_dzebl_t USING lo_multi_lang_data[li_loop].dzebl001,
                                           lo_multi_lang_data[li_loop].dzebl002,  
                                           lo_multi_lang_data[li_loop].dzebl003,  
                                           lo_multi_lang_data[li_loop].dzebl004,  
                                           lo_multi_lang_data[li_loop].dzebl000  
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      #DISPLAY cs_warning_tag,"Insert DZEBL_T with column ",lo_multi_lang_data[li_loop].dzebl001," unsuccess."

      TRY    
        &ifdef is_pgs
        SAVEPOINT sp_tran
        &endif
        PREPARE lpre_clone_upd_dzebl_t FROM ls_update_sql
        EXECUTE lpre_clone_upd_dzebl_t USING lo_multi_lang_data[li_loop].dzebl003,  
                                             lo_multi_lang_data[li_loop].dzebl004,  
                                             lo_multi_lang_data[li_loop].dzebl000,
                                             lo_multi_lang_data[li_loop].dzebl001,
                                             lo_multi_lang_data[li_loop].dzebl002  
                                          
      CATCH
        &ifdef is_pgs
        ROLLBACK WORK TO SAVEPOINT sp_tran
        &endif
        DISPLAY cs_warning_tag,"Insert or Update DZEBL_T with column ",lo_multi_lang_data[li_loop].dzebl001," unsuccess."
        DISPLAY "[Insert SQL] ",ls_insert_sql
        DISPLAY "[Update SQL] ",ls_update_sql
      END TRY
      
    END TRY  

  END FOR  

END FUNCTION
#20210607 end