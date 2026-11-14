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

CONSTANT cs_string_file_ext STRING = "str"

&include "../4gl/sadzi140_type.inc"

&ifndef DEBUG
GLOBALS "../../cfg/top_global.inc"
&endif

DEFINE ms_module  STRING 
DEFINE ms_program STRING
DEFINE ms_lang    STRING

FUNCTION sadzi140_lbl_run(p_module,p_program,p_lang)
DEFINE
  p_module  STRING,
  p_program STRING,
  p_lang    STRING  

  CALL sadzi140_lbl_initialize(p_module,p_program,p_lang)
  CALL sadzi140_lbl_start()
  CALL sadzi140_lbl_finalize()
  
END FUNCTION

FUNCTION sadzi140_lbl_initialize(p_module,p_program,p_lang)
DEFINE
  p_module  STRING,
  p_program STRING,
  p_lang    STRING 
  
  CALL sadzi140_lbl_set_module(p_module)
  CALL sadzi140_lbl_set_program(p_program)
  CALL sadzi140_lbl_set_language(p_lang)
  
END FUNCTION

FUNCTION sadzi140_lbl_start()
DEFINE
  ls_program_name STRING,
  ls_module_name  STRING,
  ls_language     STRING,
  li_index        INTEGER,
  ls_text         STRING,
  lb_success      BOOLEAN,
  lo_lang_data    DYNAMIC ARRAY OF VARCHAR(1024)

  CALL sadzi140_lbl_get_program()  RETURNING ls_program_name
  CALL sadzi140_lbl_get_module()   RETURNING ls_module_name
  CALL sadzi140_lbl_get_language() RETURNING ls_language
  
  CALL sadzi140_lbl_get_string_lang_type(ls_program_name) RETURNING lo_lang_data
  
  FOR li_index = 1 TO lo_lang_data.getLength()
    CALL sadzi140_lbl_get_string_data(ls_program_name,lo_lang_data[li_index]) RETURNING ls_text
    CALL sadzi140_lbl_gen_string_file(ls_module_name,ls_program_name,lo_lang_data[li_index],ls_text) RETURNING lb_success
    CALL sadzi140_lbl_gen_42s_file(ls_module_name,ls_program_name,lo_lang_data[li_index]) RETURNING lb_success
  END FOR 
  
END FUNCTION

FUNCTION sadzi140_lbl_gen_42s_file(p_module_name,p_file_name,p_lang)
DEFINE
  p_module_name  STRING,
  p_file_name    STRING,
  p_lang         STRING
DEFINE
  ls_module_name  STRING,
  ls_file_name    STRING,
  ls_lang         STRING,
  ls_42s_path     STRING,
  ls_str_path     STRING,
  ls_42s_file     STRING,
  ls_str_file     STRING,
  ls_src_file     STRING,
  ls_dst_file     STRING,
  ls_module_path  STRING,
  ls_os_separator STRING,
  ls_command      STRING,
  li_return       INTEGER,
  lb_success      BOOLEAN   
DEFINE
  lb_return       STRING

  LET ls_module_name = p_module_name
  LET ls_file_name   = p_file_name
  LET ls_lang        = p_lang

  LET lb_return = TRUE
  
  #取得作業系統的分隔符號
  CALL os.Path.separator() RETURNING ls_os_separator
  
  CALL FGL_GETENV(ls_module_name.toUpperCase()) RETURNING ls_module_path 
  
  LET ls_42s_path = ls_module_path,ls_os_separator,
                    "42s",ls_os_separator,
                    ls_lang

  LET ls_str_path = ls_module_path,ls_os_separator,
                    "str",ls_os_separator,
                    ls_lang

  LET ls_src_file = ls_str_path,ls_os_separator,
                    ls_file_name,".42s"  

  LET ls_dst_file = ls_42s_path,ls_os_separator,
                    ls_file_name,".42s"   

  CALL os.Path.chdir(ls_str_path) RETURNING lb_success
  IF NOT lb_success THEN GOTO _error END IF 
  
  LET ls_command = "fglmkstr ",ls_file_name,".str"
  RUN ls_command RETURNING li_return
  LET lb_success = IIF(li_return==0,TRUE,FALSE)
  IF NOT lb_success THEN GOTO _error END IF
  
  CALL os.path.copy(ls_src_file,ls_dst_file) RETURNING lb_success
  IF NOT lb_success THEN GOTO _error END IF

  CALL os.path.delete(ls_src_file) RETURNING lb_success
  IF NOT lb_success THEN GOTO _error END IF

  LABEL _error:
  
  LET lb_return = lb_success
  
  RETURN lb_return

END FUNCTION

FUNCTION sadzi140_lbl_gen_string_file(p_module_name,p_file_name,p_lang,p_text)
DEFINE
  p_module_name  STRING,
  p_file_name    STRING,
  p_lang         STRING,
  p_text         STRING
DEFINE
  ls_module_name  STRING,
  ls_file_name    STRING,
  ls_lang         STRING,
  ls_text         STRING,
  ls_string_name  STRING,
  ls_module_path  STRING,
  ls_os_separator STRING,
  lchannel_write  base.Channel
DEFINE
  lb_return       STRING

  LET ls_module_name = p_module_name
  LET ls_file_name   = p_file_name
  LET ls_lang        = p_lang
  LET ls_text        = p_text

  LET lb_return = TRUE
  
  #取得作業系統的分隔符號
  CALL os.Path.separator() RETURNING ls_os_separator
  
  CALL FGL_GETENV(ls_module_name.toUpperCase()) RETURNING ls_module_path 
  
  LET ls_string_name = ls_module_path,ls_os_separator,
                       "str",ls_os_separator,
                       ls_lang,ls_os_separator,
                       ls_file_name,".",cs_string_file_ext
  
  LET lchannel_write = base.Channel.create()
  CALL lchannel_write.setDelimiter("")

  #開檔寫入
  TRY
    CALL lchannel_write.openFile(ls_string_name,"w")
    CALL lchannel_write.write(ls_text)
    #關檔
    CALL lchannel_write.close()
  CATCH
    DISPLAY cs_error_tag,"Generate String file Error !!"
    LET lb_return = FALSE
  END TRY  

  RETURN lb_return

END FUNCTION

FUNCTION sadzi140_lbl_finalize()
END FUNCTION

FUNCTION sadzi140_lbl_get_string_lang_type(p_prog_name)
DEFINE
  p_prog_name STRING
DEFINE
  ls_prog_name    STRING,
  lo_string_data  DYNAMIC ARRAY OF VARCHAR(1024),
  ls_sql          STRING,
  li_rec_cnt      INTEGER 
DEFINE
  ls_return DYNAMIC ARRAY OF VARCHAR(1024)  

  LET ls_prog_name = p_prog_name

  LET li_rec_cnt = 1

  LET ls_sql = "select distinct zd.gzzd002             ",
               "  from gzzd_t zd                       ",
               " where zd.gzzd001 = '",ls_prog_name,"' "
                                                     
  PREPARE lpre_get_string_lang_type FROM ls_sql
  DECLARE lcur_get_string_lang_type CURSOR FOR lpre_get_string_lang_type

  OPEN lcur_get_string_lang_type
  FOREACH lcur_get_string_lang_type INTO lo_string_data[li_rec_cnt]
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF
    
    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_string_lang_type
  CALL lo_string_data.deleteElement(li_rec_cnt)
  
  FREE lpre_get_string_lang_type
  FREE lcur_get_string_lang_type  

  LET ls_return.* = lo_string_data.*
  
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_lbl_get_string_data(p_prog_name,p_lang)
DEFINE
  p_prog_name STRING,
  p_lang      STRING 
DEFINE
  ls_prog_name    STRING,
  ls_lang         STRING,
  ls_string_data  VARCHAR(1024),
  lo_line_str_buf base.StringBuffer,
  ls_sql          STRING,
  li_rec_cnt      INTEGER 
DEFINE
  ls_return STRING  

  LET ls_prog_name = p_prog_name
  LET ls_lang      = p_lang  

  LET li_rec_cnt = 1
  LET lo_line_str_buf = base.StringBuffer.create()

  &ifdef is_msv
  LET ls_sql = "select '""'+zd.gzzd003+'"" = ""'+zd.gzzd005+'""' ",
               "  from gzzd_t zd                                 ",
               " where zd.gzzd001 = '",ls_prog_name,"'           ",
               "   and zd.gzzd002 = '",ls_lang,"'                "
  &else
  #Oracle & PGS 共用
  LET ls_sql = "select '""'||zd.gzzd003||'"" = ""'||zd.gzzd005||'""' ",
               "  from gzzd_t zd                                     ",
               " where zd.gzzd001 = '",ls_prog_name,"'               ",
               "   and zd.gzzd002 = '",ls_lang,"'                    "
  &endif             
                                                     
  PREPARE lpre_get_string_data FROM ls_sql
  DECLARE lcur_get_string_data CURSOR FOR lpre_get_string_data

  OPEN lcur_get_string_data
  FOREACH lcur_get_string_data INTO ls_string_data
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET ls_string_data = ls_string_data,"\n"
    CALL lo_line_str_buf.append(ls_string_data)
    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_string_data
  
  FREE lpre_get_string_data
  FREE lcur_get_string_data  

  LET ls_return = lo_line_str_buf.toString()
  
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_lbl_set_language(p_lang)
DEFINE
  p_lang STRING

  LET ms_lang = p_lang
  
END FUNCTION
 
FUNCTION sadzi140_lbl_get_language()
DEFINE
  ls_return STRING

  LET ls_return = ms_lang

  RETURN ls_return
  
END FUNCTION 

FUNCTION sadzi140_lbl_set_program(p_program)
DEFINE
  p_program STRING
  
  LET ms_program = p_program
  
END FUNCTION
 
FUNCTION sadzi140_lbl_get_program()
DEFINE
  ls_return STRING 

  LET ls_return = ms_program

  RETURN ls_return
  
END FUNCTION 

FUNCTION sadzi140_lbl_set_module(p_module)
DEFINE
  p_module STRING
  
  LET ms_module = p_module
  
END FUNCTION
 
FUNCTION sadzi140_lbl_get_module()
DEFINE
  ls_return STRING 

  LET ls_return = ms_module

  RETURN ls_return
  
END FUNCTION 
