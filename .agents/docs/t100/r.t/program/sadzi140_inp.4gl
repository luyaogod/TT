{
  異動型態    異動單號       日 期       異動者      異動內容
  ========= ============= ========== ========== ===============================================
  Modify                  20191209   02160      1.修改支援 MSSQLServer(1.1.071)   
  Modify                  20200330   02160      1.修改支援 PosgreSQL(1.1.072)   
  Modify                  20231017   02160      1.修改支援 DM
}  

&include "../4gl/sadzi140_mcr.inc"
IMPORT os

SCHEMA ds

&include "../4gl/sadzp000_cnst.inc"
&include "../4gl/sadzi140_cnst.inc"

CONSTANT cs_left  STRING = "sr_left"
CONSTANT cs_right STRING = "sr_right"

GLOBALS "../../cfg/top_global.inc"

&include "../4gl/sadzi140_type.inc"

DEFINE ms_lang STRING

FUNCTION sadzi140_inp_run(p_t_dzeb_t)
DEFINE 
  p_t_dzeb_t T_DZEB_T
DEFINE
  lb_return BOOLEAN  
  
  CALL sadzi140_inp_initialize()
  CALL sadzi140_inp_initial_form()
  CALL sadzi140_inp_start(p_t_dzeb_t.*) RETURNING lb_return 
  CALL sadzi140_inp_finalize()

  RETURN lb_return
  
END FUNCTION 
  
FUNCTION sadzi140_inp_initialize()

  &ifndef DEBUG
    LET ms_lang = g_lang
  &else
    LET ms_lang = cs_default_lang
  &endif
  
END FUNCTION

FUNCTION sadzi140_inp_initial_form()
DEFINE 
  lw_window   ui.Window,
  lf_form     ui.Form,
  ls_cfg_path STRING,
  ls_4st_path STRING,
  ls_img_path STRING
  
  OPTIONS INPUT WRAP

  &ifndef DEBUG
    OPEN WINDOW w_sadzi140_inp WITH FORM cl_ap_formpath("ADZ","sadzi140_inp")
  &else
    OPEN WINDOW w_sadzi140_inp WITH FORM sadzi140_util_get_form_path("sadzi140_inp")
  &endif
  
  CURRENT WINDOW IS w_sadzi140_inp
  
  &ifndef DEBUG
  CALL cl_ui_wintitle(1) #工具抬頭名稱
  CALL cl_load_4ad_interface(NULL)
  &endif

  CALL sadzi140_util_set_form_title('sadzi140_inp',ms_lang)
  CALL sadzi140_util_set_form_style(ms_lang)
  
END FUNCTION 


FUNCTION sadzi140_inp_start(p_t_dzeb_t)
DEFINE 
  p_t_dzeb_t T_DZEB_T
DEFINE 
  lo_t_dzeb_t     T_DZEB_T,
  lb_result       BOOLEAN,
  ls_error_code   STRING,
  ls_assign_value STRING,
  li_index        INTEGER,
  ls_dzeb012      STRING,
  lo_t_dzig_t     T_DZIG_T
DEFINE 
  lo_assign_value RECORD
                    column_name  VARCHAR(100),
                    assign_value VARCHAR(100)
                  END RECORD
DEFINE
  lb_return BOOLEAN   

  LET lo_t_dzeb_t.* = p_t_dzeb_t.*

  DIALOG ATTRIBUTE(UNBUFFERED)
    
    INPUT lo_assign_value.* FROM sr_assign_value.* ATTRIBUTE(WITHOUT DEFAULTS)
    END INPUT

    BEFORE DIALOG
      #判斷型態給預設賦予值
      CASE UPSHIFT(lo_t_dzeb_t.dzeb007)
        WHEN "BLOB"
          LET ls_assign_value = "' '"
        WHEN "CLOB"
          LET ls_assign_value = "' '"
        WHEN "DATE"
          LET ls_assign_value = cs_get_date --"sysdate"
        WHEN "NUMBER"
          LET ls_assign_value = "0"
        WHEN "NUMERIC" #20240529
          LET ls_assign_value = "0"
        WHEN "TIMESTAMP"
          LET ls_assign_value = cs_get_time_stamp
        WHEN "VARCHAR2"
          LET ls_assign_value = "' '"
        #20191209 begin  
        WHEN "DATETIME2"
          LET ls_assign_value = cs_get_time_stamp
        WHEN "DECIMAL"
          LET ls_assign_value = "0"
        WHEN "NVARCHAR"
          LET ls_assign_value = "' '"
        WHEN "VARBINARY"            
          LET ls_assign_value = "' '"
        #20191209 end  
        #20200330 begin
        --PostgreSQL
        WHEN "VARCHAR"            
          LET ls_assign_value = "' '"
        WHEN "TEXT"            
          LET ls_assign_value = "' '"
        WHEN "INT2"            
          LET ls_assign_value = "0"
        WHEN "INT4"            
          LET ls_assign_value = "0"
        WHEN "INT8"            
          LET ls_assign_value = "0"
        WHEN "BYTEA"            
          LET ls_assign_value = "' '"
        #20200330 end
        #20231017 begin
        WHEN "SMALLINT"            
          LET ls_assign_value = "0"
        WHEN "INTEGER"            
          LET ls_assign_value = "0"
        WHEN "BIGINT"            
          LET ls_assign_value = "0"
        #20231017 end  
      END CASE
      LET ls_dzeb012 = lo_t_dzeb_t.dzeb012
      LET ls_dzeb012 = ls_dzeb012.trim()
      LET lo_assign_value.column_name  = lo_t_dzeb_t.dzeb002
      LET lo_assign_value.assign_value = ls_assign_value
    
    ON ACTION btn_ok
      LET lb_result = TRUE 
      LET ls_assign_value = lo_assign_value.assign_value
      LET ls_assign_value = ls_assign_value.trim() 
      IF (ls_assign_value IS NULL) OR (ls_assign_value = "''") THEN
        LET ls_error_code = "adz-00653"
        CALL sadzp000_msg_show_error(NULL, ls_error_code, NULL, 0)
        NEXT FIELD ed_assign_value
      END IF
      #進行一般資料檢核
      CALL sadzi140_util_check_default_value(lo_t_dzeb_t.dzeb007,lo_t_dzeb_t.dzeb008,lo_assign_value.assign_value) RETURNING lb_result,ls_error_code
      IF NOT lb_result THEN
        CALL sadzp000_msg_show_error(NULL, ls_error_code, NULL, 0)
        NEXT FIELD ed_assign_value
      ELSE #20200330
        CALL sadzi140_util_refresh_default_value(lo_assign_value.assign_value) RETURNING lo_assign_value.assign_value #20200330      
      END IF
      LET lo_t_dzig_t.dzig001 = lo_t_dzeb_t.dzeb001
      LET lo_t_dzig_t.dzig002 = lo_t_dzeb_t.dzeb002
      LET lo_t_dzig_t.dzig003 = lo_assign_value.assign_value
      CALL sadzi140_inp_insert_update_dzig_t(lo_t_dzig_t.*)
      EXIT DIALOG
      
    ON ACTION btn_cancel 
      LET lb_result = FALSE 
      EXIT DIALOG
      
    ON ACTION CLOSE
      EXIT DIALOG
  
  END DIALOG

  LET lb_return = lb_result

  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_inp_finalize()
  CLOSE WINDOW w_sadzi140_inp
END FUNCTION

FUNCTION sadzi140_inp_insert_update_dzig_t(p_t_dzig_t)
DEFINE
  p_t_dzig_t   T_DZIG_T
DEFINE 
  lo_t_dzig_t  T_DZIG_T

  LET lo_t_dzig_t.* = p_t_dzig_t.*

  BEGIN WORK 
  
  TRY 
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    insert into dzig_t(
                        dzig001,dzig002,dzig003
                      )
               values (
                        lo_t_dzig_t.dzig001,lo_t_dzig_t.dzig002,lo_t_dzig_t.dzig003
                      )
                       
    DISPLAY cs_message_tag,"Insert DZIG_T [",lo_t_dzig_t.dzig001,",",lo_t_dzig_t.dzig002,",",lo_t_dzig_t.dzig003,"] success."
  CATCH
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif
    TRY
      &ifdef is_pgs
      SAVEPOINT sp_tran
      &endif
      UPDATE dzig_t                                 
         SET dzig003 = lo_t_dzig_t.dzig003
       WHERE dzig001 = lo_t_dzig_t.dzig001
         AND dzig002 = lo_t_dzig_t.dzig002
      DISPLAY cs_message_tag,"Update DZIG_T [",lo_t_dzig_t.dzig001,",",lo_t_dzig_t.dzig002,",",lo_t_dzig_t.dzig003,"] success."
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      DISPLAY cs_warning_tag,"Insert or Update DZIG_T unsuccess : ",SQLCA.sqlcode
    END TRY
  END TRY

  COMMIT WORK

END FUNCTION
