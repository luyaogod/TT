{
  異動型態    異動單號       日 期       異動者      異動內容
  ========= ============= ========== ========== ===============================================
  Modify                  20161103   Ernestliou 1.檢視資料的方式更新
  Modify                  20170907   Ernestliou 1.版次比對時Debug模式則可以看所有的歷程
  Modify                  20170912   Ernestliou 1.索引匯入時判斷有無已存在 index 欄位, 有則匯入方設失效
  Modify                  20180903   02160      1. 某些alter table的情境是不合法的, 但需要提供人工排除問題的做法 : 只需要提供給客製環境
  Modify                  20200330   02160      1.修改支援 PosgreSQL(1.1.072)   
}  
  
&include "../4gl/sadzi140_mcr.inc" 

IMPORT os
IMPORT util

SCHEMA ds  

&include "../4gl/sadzp000_cnst.inc"
&include "../4gl/sadzi140_cnst.inc"
&include "../4gl/sadzi140_type.inc"

&include "../4gl/sadzp200_cnst.inc"

PRIVATE TYPE T_COLUMN_REV_DATA RECORD 
               DZEB001 LIKE DZEB_T.DZEB001,
               DZEB002 LIKE DZEB_T.DZEB002,
               DZEB003 LIKE DZEB_T.DZEB003,
               DZEB004 LIKE DZEB_T.DZEB004,
               DZEB005 LIKE DZEB_T.DZEB005,
               DZEB006 LIKE DZEB_T.DZEB006,
               DZEB007 LIKE DZEB_T.DZEB007,
               DZEB008 LIKE DZEB_T.DZEB008,
               DZEB012 LIKE DZEB_T.DZEB012,
               DZEB021 LIKE DZEB_T.DZEB021,
               DZEB022 LIKE DZEB_T.DZEB022,
               DZEB023 LIKE DZEB_T.DZEB023,
               DZEB024 LIKE DZEB_T.DZEB024,
               DZEB027 LIKE DZEB_T.DZEB027,
               DZEB028 LIKE DZEB_T.DZEB028,
               DZEB029 LIKE DZEB_T.DZEB029,
               DZEB030 LIKE DZEB_T.DZEB030,
               DZEB031 LIKE DZEB_T.DZEB031
             END RECORD

PRIVATE TYPE T_INDEX_REV_DATA RECORD 
               DZEC001  LIKE DZEC_T.DZEC001,
               DZEC002  LIKE DZEC_T.DZEC002,
               DZEC003  LIKE DZEC_T.DZEC003,
               DZEC004  LIKE DZEC_T.DZEC004,
               DZEC005  LIKE DZEC_T.DZEC005,
               DZEC006  LIKE DZEC_T.DZEC006,
               DZEC007  LIKE DZEC_T.DZEC007,
               DZEC008  LIKE DZEC_T.DZEC008,
               DZECSTUS LIKE DZEC_T.DZECSTUS
             END RECORD

PRIVATE TYPE T_KEY_REV_DATA RECORD 
               DZED001 LIKE DZED_T.DZED001,
               DZED002 LIKE DZED_T.DZED002,
               DZED003 LIKE DZED_T.DZED003,
               DZED004 LIKE DZED_T.DZED004,
               DZED005 LIKE DZED_T.DZED005,
               DZED006 LIKE DZED_T.DZED006,
               DZED007 LIKE DZED_T.DZED007,
               DZED008 LIKE DZED_T.DZED008,
               DZED009 LIKE DZED_T.DZED009,
               DZED010 LIKE DZED_T.DZED010
             END RECORD

PRIVATE TYPE T_REVISION_LIST RECORD 
              rl_RECORDTYPE   VARCHAR(1),     --RECORDTYPE  
              rl_CHECKBOX     VARCHAR(4),     --CHECKBOX    
              rl_ALM_VERSION  VARCHAR(30),    --VERSION     
              rl_REQUEST_NO   VARCHAR(512),   --REV_DESC    
              rl_SEQ_PAD_ZERO VARCHAR(30),    --REVISION_PAD
              rl_SEQUENCE     VARCHAR(30),    --REVISION    
              rl_MDY_NUM_PAD  VARCHAR(30),    --MDY_NUM_PAD
              rl_MDY_NUM      VARCHAR(30)     --MDY_NUM
            END RECORD

PRIVATE TYPE T_LOG_MESSAGES RECORD
               lm_RECORD_NUM  INTEGER,
               lm_RECORD      VARCHAR(2048)
             END RECORD
            
&ifndef DEBUG
GLOBALS "../../cfg/top_global.inc"
&endif

CONSTANT cs_err_db_lead_code_ora STRING = "ORA"
             
DEFINE M_REVISION DYNAMIC ARRAY OF T_REVISION_LIST

DEFINE ms_owner_name STRING
DEFINE ms_table_name STRING
DEFINE ms_lang       STRING 
DEFINE mb_result     BOOLEAN
DEFINE ms_DGENV      STRING
DEFINE mb_checkout_active BOOLEAN 
DEFINE mb_debug      BOOLEAN

             
FUNCTION sadzi140_rev_run(p_table_name,p_owner_name,p_lang,p_checkout_active,p_debug)
DEFINE 
  p_table_name STRING,
  p_owner_name STRING,
  p_lang       STRING,
  p_checkout_active BOOLEAN,
  p_debug      BOOLEAN

  CALL sadzi140_rev_initialize(p_table_name,p_owner_name,p_lang,p_checkout_active,p_debug)  
  CALL sadzi140_rev_initial_form()
  CALL sadzi140_rev_start()
  CALL sadzi140_rev_finalize()

  RETURN mb_result
  
END FUNCTION

FUNCTION sadzi140_rev_initialize(p_table_name,p_owner_name,p_lang,p_checkout_active,p_debug)
DEFINE 
  p_table_name STRING,
  p_owner_name STRING,
  p_lang       STRING,
  p_checkout_active BOOLEAN,
  p_debug      BOOLEAN  

  LET ms_owner_name      = p_owner_name
  LET ms_table_name      = p_table_name
  LET ms_lang            = p_lang
  LET mb_checkout_active = p_checkout_active
  LET mb_debug           = p_debug
  
  LET mb_result = TRUE
  LET ms_DGENV  = FGL_GETENV("DGENV")
  
END FUNCTION

FUNCTION sadzi140_rev_initial_form()
DEFINE
  lo_window ui.Window, 
  lo_form   ui.Form

  #CLOSE WINDOW SCREEN
  
  &ifndef DEBUG
  OPEN WINDOW w_sadzi140_rev WITH FORM cl_ap_formpath("ADZ","sadzi140_rev")
  #CONNECT TO cs_master_db
  &else
  OPEN WINDOW w_sadzi140_rev WITH FORM sadzi140_util_get_form_path("sadzi140_rev")
  #CONNECT TO "local"
  &endif
  
  CURRENT WINDOW IS w_sadzi140_rev

  CALL sadzi140_util_set_form_title('sadzi140_rev',ms_lang)

  LET lo_window = ui.Window.getCurrent()
  LET lo_form   = lo_window.getForm()
  
  IF (ms_DGENV = cs_dgenv_customize) AND (NOT mb_debug) THEN #20170907
    CALL lo_form.setElementHidden("formonly.lbl_version", TRUE)
    CALL lo_form.setElementHidden("formonly.lb_card_no_pad", TRUE)
    CALL lo_form.setElementHidden("btn_recover", TRUE)
    CALL lo_form.setElementHidden("formonly.lbl_mdy_num_pad", FALSE)
  ELSE  
    CALL lo_form.setElementHidden("formonly.lbl_version", FALSE)
    CALL lo_form.setElementHidden("formonly.lb_card_no_pad", FALSE)
    CALL lo_form.setElementHidden("btn_recover", FALSE)
    CALL lo_form.setElementHidden("formonly.lbl_mdy_num_pad", TRUE)
  END IF
  
END FUNCTION

FUNCTION sadzi140_rev_start()
DEFINE
  lo_rev_list     DYNAMIC ARRAY OF T_REVISION_LIST,
  lo_revision     T_REVISION,
  lo_user_list    DYNAMIC ARRAY OF VARCHAR(20),
  lo_db_connstr   T_DB_CONNSTR,
  ls_sql_script   STRING,
  ls_sql_filename STRING,
  li_record_count INTEGER,
  li_loop         INTEGER,
  li_loop2        INTEGER,
  ls_message      STRING,
  ls_check        STRING,
  li_check        INTEGER,
  li_index        INTEGER,
  lb_success      BOOLEAN,
  ls_replace_arg  STRING,
  ls_question     STRING,
  ls_all_message  STRING,
  li_arr_curr     INTEGER,
  ls_compare_type STRING,
  ls_version      STRING,
  ls_reversion    STRING,
  ls_request_no   STRING,
  ls_sequence_no  STRING

  DIALOG ATTRIBUTES(UNBUFFERED)
    INPUT ARRAY lo_rev_list FROM sr_runversion.*
      BEFORE INPUT
        LET lo_rev_list.* = M_REVISION.*
        CALL DIALOG.setActionHidden("insert",TRUE)
        CALL DIALOG.setActionHidden("append",TRUE)
        CALL DIALOG.setActionHidden("delete",TRUE)
        
      ON CHANGE lbl_check
        LET li_check = ARR_CURR()
        LET ls_check = lo_rev_list[li_check].rl_CHECKBOX 
        FOR li_index = 1 TO lo_rev_list.getLength()
          LET lo_rev_list[li_index].rl_CHECKBOX = "N" 
        END FOR
        IF (NVL(lo_rev_list[li_check].rl_ALM_VERSION,cs_null_value) <> cs_null_value) AND 
           (NVL(lo_rev_list[li_check].rl_SEQUENCE,cs_null_value) <> cs_null_value) THEN
          LET lo_rev_list[li_check].rl_CHECKBOX = ls_check
        END IF   
      
    END INPUT

    INPUT ls_compare_type FROM sr_compare_base.rdg_compare_select 
    END INPUT    

    BEFORE DIALOG 
      CALL sadzi140_rev_fill_rev_list(ms_table_name,ms_dgenv)
      CALL DIALOG.setActionActive("btn_recover", mb_checkout_active)

    ON ACTION btn_diff
      LET li_arr_curr = 0
      FOR li_loop = 1 TO lo_rev_list.getLength()
        IF lo_rev_list[li_loop].rl_CHECKBOX = "Y" THEN
          LET li_arr_curr = li_loop
          EXIT FOR
        END IF
      END FOR
      IF li_arr_curr = 0 THEN 
        LET ls_replace_arg = "|"
        CALL sadzp000_msg_show_error(NULL, 'adz-00927', ls_replace_arg, 0)
        CONTINUE DIALOG 
      END IF  
      IF li_arr_curr <> 0 THEN 
        IF (NVL(lo_rev_list[li_arr_curr].rl_ALM_VERSION,cs_null_value) <> cs_null_value) AND 
           (NVL(lo_rev_list[li_arr_curr].rl_SEQUENCE,cs_null_value) <> cs_null_value) THEN
          CALL sadzi140_rev_refresh_revision_data(lo_rev_list[li_arr_curr].*) RETURNING lo_revision.*
          CALL sadzi140_rdif_run(ms_owner_name,ms_table_name,lo_revision.*,ls_compare_type)
        END IF
      END IF
      
    ON ACTION btn_recover
      LET lb_success = FALSE
      FOR li_index = 1 TO lo_rev_list.getLength()
        IF lo_rev_list[li_index].rl_CHECKBOX = "Y" THEN
          LET li_arr_curr = li_index
          EXIT FOR
        END IF
      END FOR
      IF li_arr_curr = 0 THEN 
        LET ls_replace_arg = "|"
        CALL sadzp000_msg_show_error(NULL, 'adz-00928', ls_replace_arg, 0)
        CONTINUE DIALOG 
      END IF  
      IF li_arr_curr <> 0 THEN
        LET ls_replace_arg = "|"
        CALL sadzp000_msg_question_box(NULL, "adz-00188", ls_replace_arg, 0) RETURNING ls_question                                       
        IF ls_question = cs_question_yes THEN
          CALL sadzi140_rev_refresh_revision_data(lo_rev_list[li_arr_curr].*) RETURNING lo_revision.*
          CALL sadzi140_rev_make_revision(ms_table_name,lo_revision.*) RETURNING lb_success
          IF lb_success THEN
            LET ls_replace_arg = "|"
            CALL sadzp000_msg_show_info(NULL, 'adz-00189', ls_replace_arg, 0)
          ELSE
            LET ls_replace_arg = "|"
            CALL sadzp000_msg_show_error(NULL, 'adz-00190', ls_replace_arg, 0)
          END IF
        END IF
      ELSE
        CONTINUE DIALOG   
      END IF
      LET mb_result = lb_success
      
    ON ACTION btn_export
      FOR li_index = 1 TO lo_rev_list.getLength()
        IF lo_rev_list[li_index].rl_CHECKBOX = "Y" THEN
          LET li_arr_curr = li_index
          IF (NVL(lo_rev_list[li_arr_curr].rl_ALM_VERSION,cs_null_value) <> cs_null_value) AND 
             (NVL(lo_rev_list[li_arr_curr].rl_SEQUENCE,cs_null_value) <> cs_null_value) THEN
            LET ls_version     = lo_rev_list[li_arr_curr].rl_ALM_VERSION 
            LET ls_reversion   = lo_rev_list[li_arr_curr].rl_ALM_VERSION
            LET ls_request_no  = lo_rev_list[li_arr_curr].rl_REQUEST_NO
            LET ls_sequence_no = lo_rev_list[li_arr_curr].rl_SEQUENCE
            CALL sadzi140_util_export_table_schema(ms_table_name,ms_lang,ls_version,ls_reversion,ls_request_no,ls_sequence_no)
          END IF    
        END IF
      END FOR
    
    ON ACTION btn_cancel
      LET mb_result = FALSE
      EXIT DIALOG
      
    ON ACTION CLOSE
      LET mb_result = FALSE
      EXIT DIALOG
      
  END DIALOG
  
END FUNCTION

FUNCTION sadzi140_rev_finalize()
  CLOSE WINDOW w_sadzi140_rev
END FUNCTION

FUNCTION sadzi140_rev_fill_rev_list(p_table_name,p_dgenv)
DEFINE 
  p_table_name STRING,
  p_dgenv      STRING
DEFINE 
  ls_table_name STRING,
  ls_dgenv      STRING,
  ls_sql        STRING,
  li_count      INTEGER,
  ls_dgenv_sql  STRING

  LET ls_table_name = p_table_name
  LET ls_dgenv      = p_dgenv

  { 
  LET ls_sql = "SELECT DISTINCT '','N', EV.DZEV018, EV.DZEV019, LPAD(EV.DZEV003,5,'0') DZEV003_PAD, EV.DZEV003 ",
               "  FROM DZEV_T EV                                                                               ",
               " WHERE EV.DZEV002 = '",ls_table_name.toUpperCase(),"'                                          ", 
               "   AND EV.DZEV004 = 'TableDesigner'                                                            ",
               " ORDER BY TO_NUMBER(EV.DZEV003) DESC                                                           "
  }

  IF (ls_dgenv = cs_dgenv_customize) AND (NOT mb_debug) THEN #20170907
    LET ls_dgenv_sql = " and ev.dzev022 = '",cs_dgenv_customize,"' ",
                       " and ev.dzev019 not in ('",cs_disable_alm_request_no,"','",cs_topstd_request_no,"') " 
  ELSE
    LET ls_dgenv_sql = ""
  END IF

  &ifdef is_msv
  LET ls_sql = "select evf.note,evf.checked,evf.dzev018,evf.dzev019,evf.dzev003_pad,evf.dzev003,evf.mdy_num_pad,evf.mdy_num                                         ",
               "  from (                                                                                                                                            ",
               "        select evx.*,ROW_NUMBER() over (order by evx.dzev003) mdy_num, replace(str(ROW_NUMBER() over (order by evx.dzev003),5),' ','0') mdy_num_pad ",                                                        
               "          from (                                                                                                                                    ",
               "                select distinct '' note,'N' checked, ev.dzev018, ev.dzev019,  replace(str(ev.dzev003,5),' ','0') dzev003_pad, ev.dzev003            ",
               "                  from dzev_t ev                                                                                                                    ",
               "                 where 1=1                                                                                                                          ",
               "                   and ev.dzev001 = '",cs_master_db.toLowerCase(),"'                                                                                ",
               "                   and ev.dzev002 = '",ls_table_name.toLowerCase(),"'                                                                               ",
               "                   and ev.dzev004 = 'TableDesigner'                                                                                                 ",
               ls_dgenv_sql,                                                                                                                                        
               "               ) evx                                                                                                                                ",
               "       ) evf                                                                                                                                        ",
               " order by evf.mdy_num desc                                                                                                                          "
  &else
  &ifdef is_pgs
  LET ls_sql = "select evf.note,evf.checked,evf.dzev018,evf.dzev019,evf.dzev003_pad,                       ", 
               "       evf.dzev003,evf.mdy_num_pad,evf.mdy_num                                             ", 
               "  from (                                                                                   ",                                                          
               "        select evx.*,                                                                      ", 
               "               ROW_NUMBER() over (order by evx.dzev003) mdy_num,                           ", 
               "               lpad((ROW_NUMBER() over (order by evx.dzev003))::varchar,5,'0') mdy_num_pad ", 
               "          from (                                                                           ",                                                          
               "                select distinct '' note,'N' checked, ev.dzev018, ev.dzev019,               ", 
               "                       lpad(ev.dzev003,5,'0') dzev003_pad, ev.dzev003                      ", 
               "                  from dzev_t ev                                                           ",                                                          
               "                 where 1=1                                                                 ",                                                          
               "                   and ev.dzev001 = '",cs_master_db.toLowerCase(),"'                       ",                                                          
               "                   and ev.dzev002 = '",ls_table_name.toLowerCase(),"'                      ",                                                          
               "                   and ev.dzev004 = 'TableDesigner'                                        ",                                                          
               ls_dgenv_sql,
               "               ) evx                                                                       ",                                                          
               "       ) evf                                                                               ",                                                          
               " order by evf.mdy_num desc                                                                 "                                                          
  &else
  LET ls_sql = "SELECT EVF.NOTE,EVF.CHECKED,EVF.DZEV018,EVF.DZEV019,EVF.DZEV003_PAD,EVF.DZEV003,EVF.MDY_NUM_PAD,EVF.MDY_NUM                 ",
               "  FROM (                                                                                                                    ",
               "        SELECT EVX.*,ROWNUM MDY_NUM, LPAD(ROWNUM,5,'0') MDY_NUM_PAD                                                         ",
               "          FROM (                                                                                                            ",
               "                SELECT DISTINCT '' NOTE,'N' CHECKED, EV.DZEV018, EV.DZEV019, LPAD(EV.DZEV003,5,'0') DZEV003_PAD, EV.DZEV003 ",
               "                  FROM DZEV_T EV                                                                                            ",
               "                 WHERE 1=1                                                                                                  ",
               "                   AND EV.DZEV001 = '",cs_master_db.toUpperCase(),"'                                                        ",
               "                   AND EV.DZEV002 = '",ls_table_name.toUpperCase(),"'                                                       ",
               "                   AND EV.DZEV004 = 'TableDesigner'                                                                         ",
               ls_dgenv_sql,
               "                 ORDER BY TO_NUMBER(EV.DZEV003) ASC                                                                         ",
               "               ) EVX                                                                                                        ",
               "       ) EVF                                                                                                                ",
               " ORDER BY EVF.MDY_NUM DESC                                                                                                  "
  &endif
  &endif
  
  PREPARE lpre_fill_rev_list FROM ls_sql
  DECLARE lcur_fill_rev_list CURSOR FOR lpre_fill_rev_list

  CALL M_REVISION.clear()
  
  LET li_count = 1
  
  #DISPLAY ls_TableName
  FOREACH lcur_fill_rev_list INTO M_REVISION[li_count].*    
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF
    LET li_count = li_count + 1
  END FOREACH
  CALL M_REVISION.deleteElement(li_count)

END FUNCTION

#預覽 SQL
FUNCTION sadzi140_rev_preview_sql(p_version)
DEFINE
  p_version STRING
DEFINE
  txt_sqlscript   STRING,
  ed_version      STRING
DEFINE
  ls_sql          STRING,
  ls_version      STRING,
  ls_sql_script   STRING,
  ls_sql_filename STRING,
  lo_user_list    DYNAMIC ARRAY OF VARCHAR(20),
  lo_db_connstr   T_DB_CONNSTR,
  li_loop         INTEGER,
  ls_message      STRING,
  ls_all_message  STRING
   
  LET ls_version = p_version
  LET ed_version = p_version

  LET ls_all_message = ""
  
  CALL sadzi140_rev_open_preview_form()

  MENU 
    ON ACTION btn_cancel
      EXIT MENU  

    ON ACTION btn_execute
      
    ON ACTION CLOSE
      EXIT MENU
  END MENU
  
  CALL sadzi140_rev_close_preview_form()
  
END FUNCTION

#開啟預覽的 Form
FUNCTION sadzi140_rev_open_preview_form()

  &ifndef DEBUG
  OPEN WINDOW w_sadzi140_revl WITH FORM cl_ap_formpath("ADZ","sadzi140_revl")
  #CONNECT TO cs_master_db
  &else
  OPEN WINDOW w_sadzi140_revl WITH FORM sadzi140_util_get_form_path("sadzi140_revl")
  #CONNECT TO "local"
  &endif
  
  CURRENT WINDOW IS w_sadzi140_revl
  
  CALL sadzi140_util_set_form_title('sadzi140_revl',ms_lang)
  
END FUNCTION

#關閉預覽的 Form
FUNCTION sadzi140_rev_close_preview_form()

  CLOSE WINDOW w_sadzi140_revl
  
END FUNCTION

#觀看 Log
FUNCTION sadzi140_rev_view_logresult(p_message)
DEFINE
  p_message     STRING
DEFINE  
  ls_message     STRING, #20180903
  lo_COMBO_BOX   DYNAMIC ARRAY OF T_COMBO_BOX, #20180903
  lo_LOG_MSGS    DYNAMIC ARRAY OF T_LOG_MESSAGES, #20180903
  lo_combobox    ui.ComboBox, #20180903
  ls_db_err_list STRING, #20180903
  txt_memo       STRING  #20180903

  LET ls_message = p_message #20180903
  LET txt_memo = ls_message #20180903
  
  CALL sadzi140_rev_parse_error_log(ls_message) RETURNING lo_COMBO_BOX,lo_LOG_MSGS
  CALL sadzi140_rev_open_logresult_form()
  
  LET lo_combobox = ui.ComboBox.forName("formonly.cb_db_err_list")
  CALL sadzi140_util_fill_combobox_by_array(lo_combobox,lo_COMBO_BOX)

  DISPLAY BY NAME txt_memo
  
  DIALOG ATTRIBUTES(UNBUFFERED)
    DISPLAY ARRAY lo_LOG_MSGS TO sr_messages.*
    END DISPLAY 

    {
    INPUT ls_db_err_list FROM sr_err_list.cb_db_err_list
      ON CHANGE sr_err_list.cb_db_err_list
        IF ls_db_err_list.trim() IS NULL THEN
          CALL DIALOG.setActionActive("btn_show_sol", FALSE)
        ELSE          
          CALL DIALOG.setActionActive("btn_show_sol", TRUE)
        END IF
        CALL DIALOG.setCurrentRow("sr_messages",ls_db_err_list)
    END INPUT

    BEFORE DIALOG 
      CALL DIALOG.setActionActive("btn_show_sol", FALSE)
    
    ON ACTION btn_show_sol
    }
      
    ON ACTION btn_ok
      EXIT DIALOG  

    ON ACTION CLOSE
      EXIT DIALOG
    
  END DIALOG

  CALL sadzi140_rev_close_logresult_form()
  
END FUNCTION

#開啟看 Log 的Form
FUNCTION sadzi140_rev_open_logresult_form()

  &ifndef DEBUG
  OPEN WINDOW w_sadzi140_log WITH FORM cl_ap_formpath("ADZ","sadzi140_log")
  #CONNECT TO cs_master_db
  &else
  OPEN WINDOW w_sadzi140_log WITH FORM sadzi140_util_get_form_path("sadzi140_log")
  #CONNECT TO "local"
  &endif
  
  CURRENT WINDOW IS w_sadzi140_log
  
  CALL sadzi140_util_set_form_title('sadzi140_log',ms_lang)
  
END FUNCTION

#關閉看 Log 的Form
FUNCTION sadzi140_rev_close_logresult_form()

  CLOSE WINDOW w_sadzi140_log
  
END FUNCTION

FUNCTION sadzi140_rev_gen_sql_script_file(p_sql,p_sql_type)
DEFINE
  p_sql      STRING,
  p_sql_type STRING 
DEFINE
  ls_sql             STRING,
  ls_sql_type        STRING, 
  ls_exit_sign       STRING,
  ls_random_name     STRING,
  ls_tempdir         STRING,
  lchannel_write     base.Channel,
  ls_sript_filename  STRING,
  ls_separator       STRING, 
  ls_return          STRING
  
  LET ls_sql      = p_sql
  LET ls_sql_type = p_sql_type

  
  LET ls_exit_sign = sadzi140_util_get_exit_sign() --cs_exit_sig --"exit;"

  LET ls_separator = os.Path.separator()

  LET ls_tempdir = FGL_GETENV("TEMPDIR")
  CALL sadzi140_db_get_GUID() RETURNING ls_random_name 
  LET ls_sript_filename = ls_tempdir,ls_separator,"sadzi140_rev_",ls_sql_type,"_",ls_random_name,".sql"
  
  LET lchannel_write = base.Channel.create()
  CALL lchannel_write.setDelimiter("")

  LET ls_sql = ls_sql,"\n",
               ls_exit_sign
  
  #開檔寫入
  TRY
    CALL lchannel_write.openFile(ls_sript_filename, "w" )
    CALL lchannel_write.write(ls_sql)
    #關檔
    CALL lchannel_write.close()
  CATCH
    DISPLAY "Generate Script Error !!"
  END TRY  

  LET ls_return = ls_sript_filename

  RETURN ls_return

END FUNCTION

FUNCTION sadzi140_rev_refresh_revision_data(p_revision_list)
DEFINE
  p_revision_list  T_REVISION_LIST
DEFINE
  lo_revision_list  T_REVISION_LIST,
  lo_revision       T_REVISION
DEFINE
  lo_return T_REVISION

  LET lo_revision_list.* = p_revision_list.*

  LET lo_revision.rv_RECORDTYPE    = lo_revision_list.rl_RECORDTYPE   
  LET lo_revision.rv_CHECKBOX      = lo_revision_list.rl_CHECKBOX
  LET lo_revision.rv_ALM_VERSION   = lo_revision_list.rl_ALM_VERSION
  LET lo_revision.rv_REQUEST_NO    = lo_revision_list.rl_REQUEST_NO
  LET lo_revision.rv_SEQ_PAD_ZERO  = lo_revision_list.rl_SEQ_PAD_ZERO
  LET lo_revision.rv_SEQUENCE      = lo_revision_list.rl_SEQUENCE
  
  LET lo_return.* = lo_revision.*

  RETURN lo_return.*
  
END FUNCTION

FUNCTION sadzi140_rev_make_revision(p_table_name,p_revision)
DEFINE
  p_table_name STRING,
  p_revision   T_REVISION
DEFINE  
  ls_table_name      STRING,
  lb_success         BOOLEAN,
  lo_revision        T_REVISION,
  lo_column_rev_data DYNAMIC ARRAY OF T_COLUMN_REV_DATA,
  lo_key_rev_data    DYNAMIC ARRAY OF T_KEY_REV_DATA,
  lo_index_rev_data  DYNAMIC ARRAY OF T_INDEX_REV_DATA
DEFINE
  lb_return BOOLEAN  

  LET ls_table_name = p_table_name
  LET lo_revision.* = p_revision.*

  LET lb_success = TRUE

  CALL sadzi140_rev_get_column_rev_data(ls_table_name,lo_revision.*) RETURNING lo_column_rev_data
  CALL sadzi140_rev_get_key_rev_data(ls_table_name,lo_revision.*) RETURNING lo_key_rev_data
  CALL sadzi140_rev_get_index_rev_data(ls_table_name,lo_revision.*) RETURNING lo_index_rev_data

  #刪除及新增資料到相關表格時需要一氣呵成
  BEGIN WORK 
  
  #刪除r.t資料
  CALL sadzi140_rev_kill_column_data(ls_table_name) RETURNING lb_success  
  IF NOT lb_success THEN GOTO _rev_error END IF 
  CALL sadzi140_rev_kill_key_data(ls_table_name) RETURNING lb_success  
  IF NOT lb_success THEN GOTO _rev_error END IF 
  CALL sadzi140_rev_kill_index_data(ls_table_name) RETURNING lb_success
  IF NOT lb_success THEN GOTO _rev_error END IF 
  
  #將選擇的版本資料Insert到r.t中 
  CALL sadzi140_rev_insert_column_data(lo_column_rev_data) RETURNING lb_success  
  IF NOT lb_success THEN GOTO _rev_error END IF 
  CALL sadzi140_rev_insert_key_data(lo_key_rev_data) RETURNING lb_success  
  IF NOT lb_success THEN GOTO _rev_error END IF 
  CALL sadzi140_rev_insert_index_data(lo_index_rev_data) RETURNING lb_success  
  IF NOT lb_success THEN GOTO _rev_error END IF 

  #更新DZEA_T的版本資訊
  CALL sadzi140_rev_update_dzea_version_info(ls_table_name,lo_revision.*) RETURNING lb_success  
  IF NOT lb_success THEN GOTO _rev_error END IF 

  LABEL _rev_error:
  
  IF NOT lb_success THEN
    ROLLBACK WORK
  ELSE
    COMMIT WORK
  END IF   

  LET lb_return = lb_success

  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_rev_get_column_rev_data(p_table_name,p_revision)
DEFINE
  p_table_name STRING,
  p_revision   T_REVISION
DEFINE  
  ls_table_name      STRING,
  lo_revision        T_REVISION,
  lo_column_rev_data DYNAMIC ARRAY OF T_COLUMN_REV_DATA
DEFINE  
  ls_SQL     STRING,
  li_RecCnt  INTEGER,
  lo_RETURN  DYNAMIC ARRAY OF T_COLUMN_REV_DATA

  LET ls_table_name = p_table_name
  LET lo_revision.* = p_revision.*

  INITIALIZE lo_column_rev_data TO NULL
  INITIALIZE lo_RETURN TO NULL
  &ifdef is_msv
  LET ls_SQL = "select ev.dzev002                                dzeb001, ",
               "       ev.dzev005                                dzeb002, ",
               "       ev.dzev010                                dzeb003, ",
               "       ev.dzev006                                dzeb004, ",
               "       iif(ev.dzev007='Y','N','Y')               dzeb005, ",
               "       ev.dzev013                                dzeb006, ",
               "       ev.dzev008                                dzeb007, ",
               "       case upper(ev.dzev008)                             ",
               "          when 'DATE' then null                           ",
               "          when 'INT' then null                            ",
               "          when 'SMALLINT' then null                       ",
               "       else                                               ",
               "         ev.dzev009                                       ",
               "       end                                       dzeb008, ",
               "       ev.dzev021                                dzeb012, ",
               "       ev.dzev014                                dzeb021, ",
               "       ev.dzev015                                dzeb022, ",
               "       ev.dzev016                                dzeb023, ",
               "       ev.dzev017                                dzeb024, ",
               "       null                                      dzeb027, ",
               "       'N'                                       dzeb028, ",
               "       ev.dzev020                                dzeb029, ",
               "       ev.dzev022                                dzeb030, ", 
               "       ev.dzev023                                dzeb031  ", 
               "  from dzev_t ev                                          ", 
               " where ev.dzev002 = '",ls_table_name.toLowerCase(),"'     ", 
               "   and ev.dzev004 = 'TableDesigner'                       ", 
               "   and ev.dzev003 = '",lo_revision.rv_SEQUENCE,"'         ", 
               "   and ev.dzev018 = '",lo_revision.rv_ALM_VERSION,"'      " 
  &else
  &ifdef is_pgs
  LET ls_SQL = "select ev.dzev002                                dzeb001, ",
               "       ev.dzev005                                dzeb002, ",
               "       ev.dzev010                                dzeb003, ",
               "       ev.dzev006                                dzeb004, ",
               "       case when ev.dzev007='Y'                           ", 
               "       then 'N' else 'Y' end                     dzeb005, ",
               "       ev.dzev013                                dzeb006, ",
               "       ev.dzev008                                dzeb007, ",
               "       case upper(ev.dzev008)                             ",
               "         when 'BYTEA' then null                           ",     
               "         when 'TEXT' then null                            ",     
               "         when 'DATE' then null                            ",
               "         when 'INT2' then null                            ",
               "         when 'INT4' then null                            ",
               "         when 'INT8' then null                            ",
               "       else                                               ",
               "         ev.dzev009                                       ",
               "       end                                       dzeb008, ",
               "       ev.dzev021                                dzeb012, ",
               "       ev.dzev014                                dzeb021, ",
               "       ev.dzev015                                dzeb022, ",
               "       ev.dzev016                                dzeb023, ",
               "       ev.dzev017                                dzeb024, ",
               "       null                                      dzeb027, ",
               "       'N'                                       dzeb028, ",
               "       ev.dzev020                                dzeb029, ",
               "       ev.dzev022                                dzeb030, ", 
               "       ev.dzev023                                dzeb031  ", 
               "  from dzev_t ev                                          ", 
               " where ev.dzev002 = '",ls_table_name.toLowerCase(),"'     ", 
               "   and ev.dzev004 = 'TableDesigner'                       ", 
               "   and ev.dzev003 = '",lo_revision.rv_SEQUENCE,"'         ", 
               "   and ev.dzev018 = '",lo_revision.rv_ALM_VERSION,"'      " 
  &else
  LET ls_SQL = "SELECT LOWER(EV.DZEV002)                         DZEB001, ",
               "       LOWER(EV.DZEV005)                         DZEB002, ",
               "       EV.DZEV010                                DZEB003, ",
               "       EV.DZEV006                                DZEB004, ",
               "       DECODE(EV.DZEV007,'Y','N','N','Y','Y')    DZEB005, ",
               "       EV.DZEV013                                DZEB006, ",
               "       LOWER(EV.DZEV008)                         DZEB007, ",
               "       DECODE(EV.DZEV008,'DATE',NULL,EV.DZEV009) DZEB008, ",
               "       EV.DZEV021                                DZEB012, ",
               "       EV.DZEV014                                DZEB021, ",
               "       EV.DZEV015                                DZEB022, ",
               "       EV.DZEV016                                DZEB023, ",
               "       EV.DZEV017                                DZEB024, ",
               "       NULL                                      DZEB027, ",
               "       'N'                                       DZEB028, ",
               "       EV.DZEV020                                DZEB029, ",
               "       EV.DZEV022                                DZEB030, ",
               "       EV.DZEV023                                DZEB031  ",  
               "  FROM DZEV_T EV                                          ",
               " WHERE EV.DZEV002 = '",ls_table_name.toUpperCase(),"'     ",
               "   AND EV.DZEV004 = 'TableDesigner'                       ",
               "   AND EV.DZEV003 = '",lo_revision.rv_SEQUENCE,"'         ",
               "   AND EV.DZEV018 = '",lo_revision.rv_ALM_VERSION,"'      "
  &endif
  &endif
  
  PREPARE lpre_get_column_rev_data FROM ls_SQL
  DECLARE lcur_get_column_rev_data CURSOR FOR lpre_get_column_rev_data

  LET li_RecCnt = 1 
  
  OPEN lcur_get_column_rev_data
  FOREACH lcur_get_column_rev_data INTO lo_column_rev_data[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_get_column_rev_data
  IF li_RecCnt > 1 THEN
    CALL lo_column_rev_data.deleteElement(li_RecCnt)
  END If  
  
  FREE lpre_get_column_rev_data
  FREE lcur_get_column_rev_data  

  LET lo_RETURN.* = lo_column_rev_data.* 
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_rev_get_key_rev_data(p_table_name,p_revision)
DEFINE
  p_table_name STRING,
  p_revision   T_REVISION
DEFINE  
  ls_table_name    STRING,
  lo_revision      T_REVISION,
  lo_key_rev_data  DYNAMIC ARRAY OF T_KEY_REV_DATA
DEFINE  
  ls_SQL     STRING,
  li_RecCnt  INTEGER,
  lo_RETURN  DYNAMIC ARRAY OF T_KEY_REV_DATA

  LET ls_table_name = p_table_name
  LET lo_revision.* = p_revision.*

  INITIALIZE lo_key_rev_data TO NULL
  INITIALIZE lo_RETURN TO NULL
  
  LET ls_SQL = "select ew.dzew001 dzed001,                            ",  
               "       ew.dzew004 dzed002,                            ",  
               "       ew.dzew005 dzed003,                            ",  
               "       ew.dzew006 dzed004,                            ",  
               "       ew.dzew007 dzed005,                            ",  
               "       ew.dzew008 dzed006,                            ",  
               "       'N'        dzed007,                            ",  
               "       ew.dzew012 dzed008,                            ",  
               "       ew.dzew013 dzed009,                            ",  
               "       ew.dzew014 dzed010                             ",  
               "  from dzew_t ew                                      ",  
               " where ew.dzew001 = '",ls_table_name.toLowerCase(),"' ",  
               "   and ew.dzew002 = '",lo_revision.rv_SEQUENCE,"'     ",  
               "   and ew.dzew003 = 'Constraint'                      ",  
               "   and ew.dzew010 = '",lo_revision.rv_ALM_VERSION,"'  "   
               
  PREPARE lpre_get_key_rev_data FROM ls_SQL
  DECLARE lcur_get_key_rev_data CURSOR FOR lpre_get_key_rev_data

  LET li_RecCnt = 1 
  
  OPEN lcur_get_key_rev_data
  FOREACH lcur_get_key_rev_data INTO lo_key_rev_data[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_get_key_rev_data
  IF li_RecCnt > 1 THEN
    CALL lo_key_rev_data.deleteElement(li_RecCnt)
  END If  
  
  FREE lpre_get_key_rev_data
  FREE lcur_get_key_rev_data  

  LET lo_RETURN.* = lo_key_rev_data.* 
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_rev_get_index_rev_data(p_table_name,p_revision)
DEFINE
  p_table_name STRING,
  p_revision   T_REVISION
DEFINE  
  ls_table_name      STRING,
  lo_revision        T_REVISION,
  lo_index_rev_data  DYNAMIC ARRAY OF T_INDEX_REV_DATA
DEFINE  
  ls_SQL     STRING,
  li_RecCnt  INTEGER,
  lo_RETURN  DYNAMIC ARRAY OF T_INDEX_REV_DATA

  LET ls_table_name = p_table_name
  LET lo_revision.* = p_revision.*

  INITIALIZE lo_index_rev_data TO NULL
  INITIALIZE lo_RETURN TO NULL
  
  LET ls_SQL = "select ew.dzew001  dzec001,                          ",               
               "       ew.dzew004  dzec002,                          ",               
               "       ew.dzew005  dzec003,                          ",               
               "       ew.dzew006  dzec004,                          ",               
               "       'N'         dzec005,                          ",               
               "       ew.dzew012  dzec006,                          ",               
               "       ew.dzew013  dzec007,                          ",               
               "       ew.dzew014  dzec008,                          ",               
               "       coalesce(ew.dzewstus,'Y') dzecstus            ", #20170912         
               "  from dzew_t ew                                     ",               
               "where ew.dzew001 = '",ls_table_name.toLowerCase(),"' ",               
               "  and ew.dzew002 = '",lo_revision.rv_SEQUENCE,"'     ",               
               "  and ew.dzew003 = 'Index'                           ",               
               "  and ew.dzew010 = '",lo_revision.rv_ALM_VERSION,"'  "                  
               
  PREPARE lpre_get_index_rev_data FROM ls_SQL
  DECLARE lcur_get_index_rev_data CURSOR FOR lpre_get_index_rev_data

  LET li_RecCnt = 1 
  
  OPEN lcur_get_index_rev_data
  FOREACH lcur_get_index_rev_data INTO lo_index_rev_data[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_get_index_rev_data
  IF li_RecCnt > 1 THEN
    CALL lo_index_rev_data.deleteElement(li_RecCnt)
  END If  
  
  FREE lpre_get_index_rev_data
  FREE lcur_get_index_rev_data  

  LET lo_RETURN.* = lo_index_rev_data.* 
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_rev_kill_column_data(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_table_name STRING,
  ls_SQL        STRING,
  lb_success    BOOLEAN  
DEFINE  
  lb_return BOOLEAN

  LET ls_table_name = p_table_name
  
  LET lb_success = TRUE

  LET ls_SQL = "delete from dzeb_t                   ", 
               " where dzeb001 = '",ls_table_name,"' "
               
  TRY
    PREPARE lpre_kill_column_data FROM ls_SQL
    EXECUTE lpre_kill_column_data
  CATCH
    LET lb_success = FALSE
  END TRY  

  LET lb_return = lb_success
  
  RETURN lb_return 
  
END FUNCTION 

FUNCTION sadzi140_rev_kill_key_data(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_table_name STRING,
  ls_SQL        STRING,
  lb_success    BOOLEAN  
DEFINE  
  lb_return BOOLEAN

  LET ls_table_name = p_table_name

  LET lb_success = TRUE

  LET ls_SQL = "delete from dzed_t                   ", 
               " where dzed001 = '",ls_table_name,"' ",
               "   and dzed003 <> 'F'                "  

  TRY
    PREPARE lpre_kill_key_data FROM ls_SQL
    EXECUTE lpre_kill_key_data
  CATCH
    LET lb_success = FALSE
  END TRY  

  LET lb_return = lb_success
  
  RETURN lb_return 
  
END FUNCTION 

FUNCTION sadzi140_rev_kill_index_data(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_table_name STRING,
  ls_SQL     STRING,
  lb_success BOOLEAN  
DEFINE  
  lb_return BOOLEAN

  LET ls_table_name = p_table_name

  LET lb_success = TRUE

  LET ls_SQL = "delete from dzec_t                   ", 
               " where dzec001 = '",ls_table_name,"' "

  TRY
    PREPARE lpre_kill_index_data FROM ls_SQL
    EXECUTE lpre_kill_index_data
  CATCH
    LET lb_success = FALSE
  END TRY  

  LET lb_return = lb_success
  
  RETURN lb_return 
  
END FUNCTION 

FUNCTION sadzi140_rev_insert_column_data(p_column_rev_data)
DEFINE
  p_column_rev_data   DYNAMIC ARRAY OF T_COLUMN_REV_DATA
DEFINE 
  lo_column_rev_data  DYNAMIC ARRAY OF T_COLUMN_REV_DATA,
  li_count     INTEGER,
  lb_success   BOOLEAN,
  ls_user      VARCHAR(100),
  ldt_datetime DATETIME YEAR TO SECOND
DEFINE
  lb_return BOOLEAN
  
  LET lo_column_rev_data.* = p_column_rev_data.*
  
  LET lb_success = TRUE

  &ifndef DEBUG
  LET ls_user = g_user
  LET ldt_datetime = cl_get_current()
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  LET ldt_datetime = CURRENT YEAR TO SECOND
  &endif

  FOR li_count = 1 TO lo_column_rev_data.getLength()

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
                          lo_column_rev_data[li_count].dzeb001,lo_column_rev_data[li_count].dzeb002,lo_column_rev_data[li_count].dzeb003,lo_column_rev_data[li_count].dzeb004,lo_column_rev_data[li_count].dzeb005,
                          lo_column_rev_data[li_count].dzeb006,lo_column_rev_data[li_count].dzeb007,lo_column_rev_data[li_count].dzeb008,lo_column_rev_data[li_count].dzeb012,lo_column_rev_data[li_count].dzeb021,
                          lo_column_rev_data[li_count].dzeb022,lo_column_rev_data[li_count].dzeb023,lo_column_rev_data[li_count].dzeb024,lo_column_rev_data[li_count].dzeb027,lo_column_rev_data[li_count].dzeb028,
                          lo_column_rev_data[li_count].dzeb029,lo_column_rev_data[li_count].dzeb030,lo_column_rev_data[li_count].dzeb031,
                          ls_user,ldt_datetime,ls_user,ldt_datetime
                        )
                         
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif
      TRY
        &ifdef is_pgs
        SAVEPOINT sp_tran
        &endif
        update dzeb_t                                 
           set dzeb003  = lo_column_rev_data[li_count].dzeb003,
               dzeb004  = lo_column_rev_data[li_count].dzeb004,
               dzeb005  = lo_column_rev_data[li_count].dzeb005,
               dzeb006  = lo_column_rev_data[li_count].dzeb006,
               dzeb007  = lo_column_rev_data[li_count].dzeb007,
               dzeb008  = lo_column_rev_data[li_count].dzeb008,
               dzeb012  = lo_column_rev_data[li_count].dzeb012,
               dzeb021  = lo_column_rev_data[li_count].dzeb021,
               dzeb022  = lo_column_rev_data[li_count].dzeb022,
               dzeb023  = lo_column_rev_data[li_count].dzeb023,
               dzeb024  = lo_column_rev_data[li_count].dzeb024,
               dzeb027  = lo_column_rev_data[li_count].dzeb027,
               dzeb028  = lo_column_rev_data[li_count].dzeb028,
               dzeb029  = lo_column_rev_data[li_count].dzeb029,
               dzeb030  = lo_column_rev_data[li_count].dzeb030,
               dzeb031  = lo_column_rev_data[li_count].dzeb031,
               dzebmodid = ls_user,
               dzebmoddt = ldt_datetime
         where dzeb001  = lo_column_rev_data[li_count].dzeb001
           and dzeb002  = lo_column_rev_data[li_count].dzeb002
      CATCH
        &ifdef is_pgs
        ROLLBACK WORK TO SAVEPOINT sp_tran
        &endif
        LET lb_success = FALSE
      END TRY
    END TRY
    IF NOT lb_success THEN EXIT FOR END IF;
  END FOR  
  
  LET lb_return = lb_success
  
  RETURN lb_return

END FUNCTION

FUNCTION sadzi140_rev_insert_key_data(p_key_rev_data)
DEFINE
  p_key_rev_data  DYNAMIC ARRAY OF T_KEY_REV_DATA
DEFINE 
  lb_success BOOLEAN,
  li_count   INTEGER,
  ls_user          VARCHAR(100),
  ldt_datetime     DATETIME YEAR TO SECOND,
  lo_key_rev_data  DYNAMIC ARRAY OF T_KEY_REV_DATA
DEFINE
  lb_return BOOLEAN

  LET lo_key_rev_data.* = p_key_rev_data.*

  LET lb_success = TRUE

  &ifndef DEBUG
  LET ls_user = g_user
  LET ldt_datetime = cl_get_current()
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  LET ldt_datetime = CURRENT YEAR TO SECOND
  &endif
  
  FOR li_count = 1 TO lo_key_rev_data.getLength()
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
                          lo_key_rev_data[li_count].dzed001,lo_key_rev_data[li_count].dzed002,lo_key_rev_data[li_count].dzed003,lo_key_rev_data[li_count].dzed004,lo_key_rev_data[li_count].dzed005,
                          lo_key_rev_data[li_count].dzed006,lo_key_rev_data[li_count].dzed007,lo_key_rev_data[li_count].DZED008,lo_key_rev_data[li_count].DZED009,lo_key_rev_data[li_count].DZED010,
                          ls_user,ldt_datetime,ls_user,ldt_datetime
                        )
                         
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif

      TRY
        &ifdef is_pgs
        SAVEPOINT sp_tran
        &endif
        update dzed_t                                 
           set dzed003  = lo_key_rev_data[li_count].dzed003,
               dzed004  = lo_key_rev_data[li_count].dzed004,
               dzed005  = lo_key_rev_data[li_count].dzed005,
               dzed006  = lo_key_rev_data[li_count].dzed006,
               dzed007  = lo_key_rev_data[li_count].dzed007,
               dzed008  = lo_key_rev_data[li_count].dzed008,
               dzed009  = lo_key_rev_data[li_count].dzed009,
               dzed010  = lo_key_rev_data[li_count].dzed010,
               dzedmodid = ls_user,
               dzedmoddt = ldt_datetime          
         where dzed001  = lo_key_rev_data[li_count].dzed001
           and dzed002  = lo_key_rev_data[li_count].dzed002
      CATCH
        &ifdef is_pgs
        ROLLBACK WORK TO SAVEPOINT sp_tran
        &endif
        LET lb_success = FALSE
      END TRY
    END TRY
    IF NOT lb_success THEN EXIT FOR END IF
  END FOR

  LET lb_return = lb_success
  
  RETURN lb_return

END FUNCTION

FUNCTION sadzi140_rev_insert_index_data(p_index_rev_data)
DEFINE
  p_index_rev_data  DYNAMIC ARRAY OF T_INDEX_REV_DATA
DEFINE 
  lb_success         BOOLEAN,
  li_count           INTEGER,
  ls_user            VARCHAR(100),
  ldt_datetime       DATETIME YEAR TO SECOND,
  lo_index_rev_data  DYNAMIC ARRAY OF T_INDEX_REV_DATA
DEFINE
  lb_return BOOLEAN

  LET lo_index_rev_data.* = p_index_rev_data.*

  &ifndef DEBUG
  LET ls_user = g_user
  LET ldt_datetime = cl_get_current()
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  LET ldt_datetime = CURRENT YEAR TO SECOND
  &endif
  
  LET lb_success = TRUE

  FOR li_count = 1 TO lo_index_rev_data.getLength()
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
                          lo_index_rev_data[li_count].dzec001,lo_index_rev_data[li_count].dzec002,lo_index_rev_data[li_count].dzec003,lo_index_rev_data[li_count].dzec004,lo_index_rev_data[li_count].dzec005,
                          lo_index_rev_data[li_count].dzec006,lo_index_rev_data[li_count].dzec007,lo_index_rev_data[li_count].dzec008,lo_index_rev_data[li_count].dzecstus, #20170912 add dzecstus
                          ls_user,ldt_datetime,ls_user,ldt_datetime
                        )
                         
    CATCH
      &ifdef is_pgs
      ROLLBACK WORK TO SAVEPOINT sp_tran
      &endif

      TRY
        &ifdef is_pgs
        SAVEPOINT sp_tran
        &endif
        update dzec_t                                 
           set dzec003   = lo_index_rev_data[li_count].dzec003,
               dzec004   = lo_index_rev_data[li_count].dzec004,
               dzec005   = lo_index_rev_data[li_count].dzec005,
               dzec006   = lo_index_rev_data[li_count].dzec006,
               dzec007   = lo_index_rev_data[li_count].dzec007,
               dzec008   = lo_index_rev_data[li_count].dzec008,
               dzecmodid = ls_user,
               dzecmoddt = ldt_datetime
         where dzec001  = lo_index_rev_data[li_count].dzec001
           and dzec002  = lo_index_rev_data[li_count].dzec002
      CATCH
        &ifdef is_pgs
        ROLLBACK WORK TO SAVEPOINT sp_tran
        &endif
        LET lb_success = FALSE
      END TRY
    END TRY
    IF NOT lb_success THEN EXIT FOR END IF
  END FOR  

  LET lb_return = lb_success
  
  RETURN lb_return

END FUNCTION

FUNCTION sadzi140_rev_update_dzea_version_info(p_table_name,p_revision)
DEFINE
  p_table_name STRING,
  p_revision   T_REVISION
DEFINE  
  ls_table_name STRING,
  lo_revision   T_REVISION,
  lb_success    BOOLEAN,
  ls_SQL        STRING 
DEFINE
  lb_return BOOLEAN  

  LET ls_table_name = p_table_name
  LET lo_revision.* = p_revision.*

  LET lb_success = TRUE

  LET ls_SQL = "update dzea_t                                      ",
               "   set dzea012 = '",lo_revision.rv_ALM_VERSION,"', ",
               "       dzea013 = '",lo_revision.rv_SEQUENCE,"'     ",
               " where dzea001 = '",ls_table_name.toLowerCase(),"' " 

  TRY
    PREPARE lpre_update_dzea_version_info FROM ls_SQL
    EXECUTE lpre_update_dzea_version_info
  CATCH
    LET lb_success = FALSE
  END TRY  

  LET lb_return = lb_success
  
  RETURN lb_return 
  
END FUNCTION 

FUNCTION sadzi140_rev_get_revision_data(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name  STRING,
  lo_REVISION    T_REVISION, 
  ls_sql         STRING
DEFINE
  lo_RETURN T_REVISION

  LET ls_table_name = p_table_name.toUpperCase()

  &ifdef is_msv
  LET ls_sql = "select distinct '','N', ev.dzev018, ev.dzev019, replace(str(ev.dzev003,5),' ','0') dzev003_pad, ev.dzev003      ", 
               "  from dzev_t ev                                                                                                ", 
               " where ev.dzev002 = '",ls_table_name,"'                                                                         ", 
               "   and ev.dzev004 = 'TableDesigner'                                                                             ", 
               "   and ev.dzev003 = (                                                                                           ", 
               "                      select coalesce(cast(max(cast(coalesce(evo.dzev003,'0') as int)) as varchar),'0') max_seq ",             
               "                        from dzev_t evo                                                                         ", 
               "                       where evo.dzev001 = ev.dzev001                                                           ", 
               "                         and evo.dzev002 = ev.dzev002                                                           ", 
               "                         and evo.dzev004 = ev.dzev004                                                           ", 
               "                    )                                                                                           "  
  &else
  &ifdef is_pgs
  LET ls_sql = "select distinct '','N', ev.dzev018, ev.dzev019, lpad(ev.dzev003,5,'0') dzev003_pad, ev.dzev003                  ",
               "  from dzev_t ev                                                                                                ",
               " where ev.dzev002 = '",ls_table_name,"'                                                                         ",
               "   and ev.dzev004 = 'TableDesigner'                                                                             ",
               "   and ev.dzev003 = (                                                                                           ",
               "                      select coalesce(cast(max(cast(coalesce(evo.dzev003,'0') as int)) as varchar),'0') max_seq ",
               "                        from dzev_t evo                                                                         ",
               "                       where evo.dzev001 = ev.dzev001                                                           ",
               "                         and evo.dzev002 = ev.dzev002                                                           ",
               "                         and evo.dzev004 = ev.dzev004                                                           ",
               "                    )                                                                                           " 
  &else
  LET ls_sql = "SELECT DISTINCT '','N', EV.DZEV018, EV.DZEV019, LPAD(EV.DZEV003,5,'0') DZEV003_PAD, EV.DZEV003 ",
               "  FROM DZEV_T EV                                                                               ",          
               " WHERE EV.DZEV002 = '",ls_table_name,"'                                                        ",
               "   AND EV.DZEV004 = 'TableDesigner'                                                            ",          
               "   AND EV.DZEV003 = (                                                                          ",
               "                      SELECT TO_CHAR(MAX(TO_NUMBER(NVL(EVO.DZEV003,'0')))) MAX_SEQ             ",
               "                        FROM DZEV_T EVO                                                        ",                                 
               "                       WHERE EVO.DZEV001 = EV.DZEV001                                          ",
               "                         AND EVO.DZEV002 = EV.DZEV002                                          ",          
               "                         AND EVO.DZEV004 = EV.DZEV004                                          ",                            
               "                    )                                                                          "
  &endif
  &endif
  
  PREPARE lpre_get_revision_data FROM ls_sql                                                        
  DECLARE lcur_get_revision_data CURSOR FOR lpre_get_revision_data
  OPEN lcur_get_revision_data
  FETCH lcur_get_revision_data INTO lo_REVISION.*
  CLOSE lcur_get_revision_data
  FREE lcur_get_revision_data
  FREE lpre_get_revision_data

  LET lo_RETURN.* = lo_REVISION.*
   
  RETURN lo_RETURN.*
  
END FUNCTION

FUNCTION sadzi140_rev_parse_error_log(p_message)
DEFINE
  p_message  STRING
DEFINE
  ls_message   STRING,
  lo_text      base.StringBuffer,
  lo_COMBO_BOX DYNAMIC ARRAY OF T_COMBO_BOX,
  lo_LOG_MSGS  DYNAMIC ARRAY OF T_LOG_MESSAGES,
  li_loop      INTEGER,
  li_line_cnts INTEGER,
  ls_line_cnts STRING,
  li_cbx_cnts  INTEGER,
  ls_text      STRING,
  ls_text_line STRING
DEFINE
  lo_rtn_cbx  DYNAMIC ARRAY OF T_COMBO_BOX,
  lo_rtn_msgs DYNAMIC ARRAY OF T_LOG_MESSAGES

  LET ls_message = p_message

  LET li_line_cnts = 0
  LET li_cbx_cnts  = 0

  LET lo_text = base.StringBuffer.create()
  CALL lo_text.append(ls_message)

  FOR li_loop = 1 TO lo_text.getLength()
    LET ls_text = lo_text.subString(li_loop,li_loop)
    IF ls_text = "\n" THEN
      LET li_line_cnts = li_line_cnts + 1
      LET ls_line_cnts = li_line_cnts
      LET lo_LOG_MSGS[li_line_cnts].lm_RECORD_NUM = li_line_cnts
      LET lo_LOG_MSGS[li_line_cnts].lm_RECORD     = ls_text_line 
      IF (ls_text_line.getIndexOf(cs_err_db_lead_code_ora,1) > 0) AND 
         (ls_text_line.getIndexOf(cs_warning_tag,1) = 0) THEN 
        LET li_cbx_cnts = li_cbx_cnts + 1
        LET lo_COMBO_BOX[li_cbx_cnts].cb_combobox_name = ls_line_cnts.trim()
        LET lo_COMBO_BOX[li_cbx_cnts].cb_combobox_desc = ls_line_cnts.trim(),".",ls_text_line
      END IF
      --DISPLAY ls_text_line
      LET ls_text_line = ""
    ELSE
      LET ls_text_line = ls_text_line,ls_text
    END IF     
  END FOR  

  LET lo_rtn_cbx  = lo_COMBO_BOX
  LET lo_rtn_msgs = lo_LOG_MSGS

  RETURN lo_rtn_cbx,lo_rtn_msgs
  
END FUNCTION