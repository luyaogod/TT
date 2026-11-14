&include "../4gl/sadzi140_mcr.inc" 

IMPORT os
IMPORT util
IMPORT FGL fgldialog

SCHEMA ds

&include "../4gl/sadzp000_cnst.inc" 
&include "../4gl/sadzi140_cnst.inc"
&include "../4gl/sadzi140_type.inc"

&ifndef DEBUG
GLOBALS "../../cfg/top_global.inc"
&endif

FUNCTION sadzi140_msg_show_error(ps_title, ps_err_code, ps_replace_arg, pi_idle_sec)
DEFINE
  ps_title        STRING, 
  ps_err_code,    STRING,
  ps_replace_arg  STRING, 
  pi_idle_sec     INTEGER
DEFINE
  ls_title        STRING, 
  ls_err_code     STRING,
  ls_replace_arg  STRING, 
  li_idle_sec     INTEGER,
  li_length       INTEGER,
  li_arg_cnts     INTEGER,
  ls_arg_char     STRING,
  ls_arg_string   STRING,
  ls_err_message  STRING  

  LET ls_title       = NVL(ps_title,ps_err_code)
  LET ls_err_code    = ps_err_code
  LET ls_replace_arg = ps_replace_arg
  LET li_idle_sec    = pi_idle_sec

  &ifndef DEBUG
    LET ls_arg_string = ""
    LET li_arg_cnts = 1
    
    INITIALIZE g_errparam TO NULL
    
    LET g_errparam.code   = ls_err_code
    LET g_errparam.extend = NULL
    LET g_errparam.popup  = TRUE

    IF ls_replace_arg IS NOT NULL THEN 
      FOR li_length = 1 TO ls_replace_arg.getLength()
        LET ls_arg_char = ls_replace_arg.subString(li_length,li_length)
        IF ls_arg_char = cs_divide THEN 
          LET g_errparam.replace[li_arg_cnts] = ls_arg_string
          LET li_arg_cnts = li_arg_cnts + 1
          LET ls_arg_string = ""
        ELSE
          LET ls_arg_string = ls_arg_string,ls_arg_char
        END IF 
      END FOR
    END IF    
    
    LET g_errparam.sqlerr = 0 #SQLCA.SQLERRD[2] #或 SQLCA.SQLCODE (有需要再設定，0則不顯示)
    
    CALL cl_err()  
  &else
    LET ls_err_message = sadzi140_msg_get_message(ps_err_code,cs_default_lang)
    CALL sadzi140_msg_replace_message(ls_err_message,ps_replace_arg) RETURNING ls_err_message
    MENU ps_title ATTRIBUTE (STYLE="dialog", COMMENT=ls_err_message.trim() CLIPPED, IMAGE="stop")
      ON ACTION ok
        EXIT MENU
    END MENU
  &endif
  
END FUNCTION

FUNCTION sadzi140_msg_show_info(ps_title, ps_msg_code, ps_replace_arg, pi_idle_sec)
DEFINE
  ps_title        STRING, 
  ps_msg_code,    STRING,
  ps_replace_arg  STRING, 
  pi_idle_sec     INTEGER
DEFINE
  ls_info_message STRING,
  lb_result       BOOLEAN   

  LET ps_title = NVL(ps_title,ps_msg_code)

  LET ls_info_message = sadzi140_msg_get_message(ps_msg_code,cs_default_lang)
  CALL sadzi140_msg_replace_message(ls_info_message,ps_replace_arg) RETURNING ls_info_message
  
  MENU ps_title ATTRIBUTE (STYLE="dialog", COMMENT=ls_info_message.trim() CLIPPED, IMAGE="information")
    ON ACTION ok
      EXIT MENU
  END MENU
  
END FUNCTION

FUNCTION sadzi140_msg_message_box(ps_title, ps_msg_code, ps_replace_arg, pi_idle_sec)
DEFINE
  ps_title        STRING, 
  ps_msg_code,    STRING,
  ps_replace_arg  STRING, 
  pi_idle_sec     INTEGER
DEFINE
  ls_info_message STRING,
  ls_question     STRING,
  lb_result       BOOLEAN  
DEFINE
  ls_return  STRING  

  LET ps_title = NVL(ps_title,ps_msg_code)
  
  &ifndef DEBUG
  CALL cl_ask_confirm_parm(ps_msg_code, ps_replace_arg) RETURNING lb_result
  LET ls_question = IIF(lb_result,cs_question_yes,cs_question_no)
  &else
  LET ls_info_message = sadzi140_msg_get_message(ps_msg_code,cs_default_lang)
  CALL sadzi140_msg_replace_message(ls_info_message,ps_replace_arg) RETURNING ls_info_message
  LET ls_question = FGL_WINQUESTION(
                                     ps_title, 
                                     ls_info_message,
                                     cs_question_no, 
                                     "yes|no", 
                                     "question", 
                                     0
                                   )
  &endif  

  LET ls_return = ls_question

  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_msg_alert_box(ps_title, ps_msg_code, ps_replace_arg, pi_idle_sec)
DEFINE
  ps_title        STRING, 
  ps_msg_code,    STRING,
  ps_replace_arg  STRING, 
  pi_idle_sec     INTEGER
DEFINE
  ls_info_message STRING,
  ls_question     STRING,
  lb_result       BOOLEAN   
DEFINE
  ls_return  STRING  

  LET ps_title = NVL(ps_title,ps_msg_code)

  &ifndef DEBUG
  CALL cl_ask_confirm_parm(ps_msg_code, ps_replace_arg) RETURNING lb_result
  LET ls_question = IIF(lb_result,cs_question_yes,cs_question_no)
  &else
  LET ls_info_message = sadzi140_msg_get_message(ps_msg_code,cs_default_lang)
  CALL sadzi140_msg_replace_message(ls_info_message,ps_replace_arg) RETURNING ls_info_message
  LET ls_question = FGL_WINQUESTION(
                                     ps_title, 
                                     ls_info_message,
                                     cs_question_no, 
                                     "yes|no", 
                                     "exclamation", 
                                     0
                                   )
  &endif  

  LET ls_return = ls_question

  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_msg_get_message(p_code,p_lang)
DEFINE 
  p_code VARCHAR(30),
  p_lang VARCHAR(30)
DEFINE
  lc_msg VARCHAR(255)
DEFINE 
  ls_return STRING  
 
  LET lc_msg = ''

  SELECT gzze003 
    INTO lc_msg 
    FROM gzze_t 
   WHERE gzze001 = p_code 
     AND gzze002 = p_lang

  LET ls_return = lc_msg
  
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_msg_replace_message(ps_err_msg, ps_replace_arg)
DEFINE   
  ps_err_msg         STRING,
  ps_replace_arg     STRING
DEFINE   
  lst_replace_arg    base.StringTokenizer,
  ls_arg             STRING,
  li_index           INTEGER,
  li_replace_index   INTEGER
DEFINE
  ls_return STRING  
 
  LET ps_err_msg = ps_err_msg.trim()
  LET lst_replace_arg = base.StringTokenizer.create(ps_replace_arg, "|")
  WHILE lst_replace_arg.hasMoreTokens()
    LET ls_arg = lst_replace_arg.nextToken()
    LET li_replace_index = ps_err_msg.getIndexOf("%" || li_index+1, 1)
    IF (li_replace_index > 0) THEN
      LET ps_err_msg = sadzi140_msg_replace_string_by_index(ps_err_msg, li_replace_index, li_replace_index+1, ls_arg)
    END IF
    LET li_index = li_index + 1   
  END WHILE

  LET ls_return = ps_err_msg
  
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_msg_replace_string_by_index(ps_source, pi_from, pi_end, ps_new)
DEFINE 
  ps_source STRING, 
  pi_from   INTEGER,
  pi_end    INTEGER,  
  ps_new    STRING
DEFINE 
  li_source_length INTEGER, 
  lo_str_buf       base.StringBuffer
DEFINE
  ls_return STRING
  
  LET ps_source = ps_source.trimRight()
  LET li_source_length = ps_source.getLength()
 
  IF (pi_from < 1) THEN
    LET pi_from = 1
  ELSE
    IF (pi_from > li_source_length) THEN
      LET pi_from = li_source_length
    END IF
  END IF
 
  IF (pi_end < pi_from) THEN
    LET pi_end = pi_from
  ELSE
    IF (pi_end > li_source_length) THEN
      LET pi_end = li_source_length
    END IF
  END IF

  LET lo_str_buf = base.StringBuffer.create()
  CALL lo_str_buf.append(ps_source)
  LET pi_end = pi_end - pi_from + 1
  CALL lo_str_buf.replaceAt(pi_from, pi_end, ps_new)

  LET ls_return = lo_str_buf.toString()
  
  RETURN ls_return
  
END FUNCTION