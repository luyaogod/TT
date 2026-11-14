{
  異動型態    異動單號       日 期       異動者      異動內容
  ========= ============= ========== ========== ===============================================
  Modify                  20200330   02160      1.修改支援 PosgreSQL(1.1.072)   
}

&include "../4gl/sadzi140_mcr.inc" 

IMPORT os
IMPORT util

SCHEMA ds  

&include "../4gl/sadzp000_cnst.inc"
&include "../4gl/sadzi140_cnst.inc"

CONSTANT cs_reserve_days STRING = "120"  
CONSTANT cs_db_log_name  STRING = "sadzi140_logs"

&include "../4gl/sadzi140_type.inc"

PUBLIC TYPE T_DZHS_T RECORD
              DZHS001  LIKE DZHS_T.DZHS001,	
              DZHS002  LIKE DZHS_T.DZHS002,	
              DZHS003  LIKE DZHS_T.DZHS003,	
              DZHS004  LIKE DZHS_T.DZHS004,	
              DZHS005  LIKE DZHS_T.DZHS005,	
              DZHS006  LIKE DZHS_T.DZHS006	
            END RECORD 
             
&ifndef DEBUG
GLOBALS "../../cfg/top_global.inc"
&endif
             
DEFINE ms_owner_name  STRING
DEFINE ms_object_name STRING
DEFINE ms_lang        STRING 
DEFINE mb_result      BOOLEAN
             
FUNCTION sadzi140_logs_write_log(p_level,p_type,p_message)
DEFINE
  p_level    STRING,
  p_type     STRING,
  p_message  STRING
DEFINE
  ls_level    STRING,
  ls_type     STRING,
  ls_message  STRING,
  lb_result   BOOLEAN,
  lb_in_trans BOOLEAN,
  lo_dzhl_t   T_DZHL_T
DEFINE
  lb_return   BOOLEAN
  
  LET ls_level   = p_level
  LET ls_type    = p_type
  LET ls_message = p_message

  LET lb_in_trans = FALSE

  LET lo_dzhl_t.DZHL002 = NVL(ls_level,cs_logs_level_information)
  LET lo_dzhl_t.DZHL003 = NVL(ls_type,cs_logs_type_column)
  LET lo_dzhl_t.DZHL005 = ls_message

  TRY 
    BEGIN WORK
  CATCH
    LET lb_in_trans = TRUE
  END TRY  
  
  CALL sadzi140_logs_set_log_data(lo_dzhl_t.*) RETURNING lb_result     

  IF NOT lb_in_trans THEN 
    IF NOT lb_result THEN 
      ROLLBACK WORK
    ELSE
      COMMIT WORK
    END IF
  END IF  
  
  LET lb_return = lb_result
  
  RETURN lb_return  
  
END FUNCTION

FUNCTION sadzi140_logs_set_log_data(p_T_DZHL_T)
DEFINE
  p_T_DZHL_T  T_DZHL_T
DEFINE
  lo_T_DZHL_T    T_DZHL_T,
  li_seqence_val INTEGER,  
  lb_result      BOOLEAN
DEFINE
  lb_return  BOOLEAN

  LET lo_T_DZHL_T.* = p_T_DZHL_T.*

  LET lb_return = TRUE

  CALL sadzi140_logs_sequence_get_next_value(cs_db_log_name) RETURNING li_seqence_val
  LET lo_T_DZHL_T.DZHL001 = li_seqence_val
  
  CALL sadzi140_logs_delete_expired_log_data(cs_reserve_days) RETURNING lb_result
  IF NOT lb_result THEN LET lb_return = FALSE END IF
  
  CALL sadzi140_logs_insert_dzhl_data(lo_T_DZHL_T.*) RETURNING lb_result
  IF NOT lb_result THEN LET lb_return = FALSE END IF

  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_logs_delete_expired_log_data(p_reserve_days)
DEFINE
  p_reserve_days STRING
DEFINE
  ls_reserve_days STRING,
  ls_exec_sql     STRING
DEFINE
  lb_return  BOOLEAN  

  LET ls_reserve_days = p_reserve_days
  
  LET lb_return = TRUE
  
  #刪除過期的 Log 資料
  LET ls_exec_sql = "delete from dzhl_t                                             ",
                    " where dzhlcrtdt <= ",cs_get_date," - ",ls_reserve_days," "
                    
  TRY
    PREPARE lpre_exec_sql FROM ls_exec_sql
    EXECUTE lpre_exec_sql
  CATCH
    LET lb_return = FALSE
    DISPLAY cs_error_tag,"Execute SQL ",ls_exec_sql," not success." 
  END TRY

  RETURN lb_return  

END FUNCTION

FUNCTION sadzi140_logs_insert_dzhl_data(p_dzhl_t)
DEFINE
  p_dzhl_t  T_DZHL_T
DEFINE
  lo_dzhl_t    T_DZHL_T,
  ls_user      VARCHAR(100),
  ldt_datetime DATETIME YEAR TO SECOND
DEFINE
  lb_return  BOOLEAN  

  LET lo_dzhl_t.* = p_dzhl_t.*

  LET lb_return = TRUE

  &ifndef DEBUG
  LET ls_user = g_user
  LET ldt_datetime = cl_get_current()
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  LET ldt_datetime = CURRENT YEAR TO SECOND
  &endif

  LET lo_dzhl_t.DZHL004   = ldt_datetime
  LET lo_dzhl_t.DZHLCRTID = ls_user
  LET lo_dzhl_t.DZHLCRTDT = ldt_datetime
  LET lo_dzhl_t.DZHLMODID = ls_user
  LET lo_dzhl_t.DZHLMODDT = ldt_datetime

  TRY 
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    insert into dzhl_t (
                         dzhl001,  dzhl002,  dzhl003,  dzhl004,  dzhl005,	  
                         dzhlcrtid,dzhlcrtdt,dzhlmodid,dzhlmoddt   
                       ) 
                values (
                         lo_dzhl_t.DZHL001,  lo_dzhl_t.DZHL002,  lo_dzhl_t.DZHL003,  lo_dzhl_t.DZHL004,  lo_dzhl_t.DZHL005,	  
                         lo_dzhl_t.DZHLCRTID,lo_dzhl_t.DZHLCRTDT,lo_dzhl_t.DZHLMODID,lo_dzhl_t.DZHLMODDT                          
                       )
  CATCH 
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif
    LET lb_return = FALSE
  END TRY  
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_logs_sequence_get_next_value(p_seq_name)
DEFINE
  p_seq_name  STRING 
DEFINE
  ls_seq_name STRING,
  ls_sql      STRING,
  li_sequence INTEGER
DEFINE
  li_return  INTEGER  

  LET ls_seq_name = p_seq_name

  LET ls_sql = "select coalesce(max(hl.dzhl001),0) seq_value ",
               "  from dzhl_t hl                             "
  
  PREPARE lpre_sequence_get_next_value FROM ls_sql
  DECLARE lcur_sequence_get_next_value CURSOR FOR lpre_sequence_get_next_value

  OPEN lcur_sequence_get_next_value
  FETCH lcur_sequence_get_next_value INTO li_sequence
  CLOSE lcur_sequence_get_next_value
  
  FREE lpre_sequence_get_next_value
  FREE lcur_sequence_get_next_value  
  
  LET li_return = li_sequence + 1 

  RETURN li_return
  
END FUNCTION


FUNCTION sadzi140_logs_insert_dzhs_data(p_dzhs_t)
DEFINE
  p_dzhs_t  T_DZHS_T
DEFINE
  lo_dzhs_t  T_DZHS_T
DEFINE
  lb_return  BOOLEAN  

  LET lo_dzhs_t.* = p_dzhs_t.*

  LET lb_return = TRUE

  TRY 
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    insert into dzhs_t (
                         dzhs001,dzhs002,dzhs003,dzhs004,dzhs005,	  
                         dzhs006
                       ) 
                values (
                         lo_dzhs_t.DZHS001,lo_dzhs_t.DZHS002,lo_dzhs_t.DZHS003,lo_dzhs_t.DZHS004,lo_dzhs_t.DZHS005,	  
                         lo_dzhs_t.DZHS006
                       )
  CATCH 
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif
    LET lb_return = FALSE
  END TRY  
  
  RETURN lb_return
  
END FUNCTION
