
&include "../4gl/sadzi140_mcr.inc" 

IMPORT os
IMPORT util
IMPORT security

SCHEMA ds

&include "../4gl/sadzp000_cnst.inc"
&include "../4gl/sadzi140_cnst.inc"

&include "../4gl/sadzi140_type.inc"

&ifndef DEBUG
GLOBALS "../../cfg/top_global.inc"
&endif

#Oracle 專用 begin
FUNCTION sadzi140_prc_ora_create_get_col_default()
DEFINE
  ls_exit_sign      STRING,
  ls_GUID           STRING,
  ls_tempdir        STRING,
  lchannel_write    base.Channel,
  ls_sript_filename STRING,
  ls_create_script  STRING,
  ls_separator      STRING
DEFINE
  ls_return STRING  
  
  LET ls_separator = os.Path.separator()
  
  LET ls_tempdir = FGL_GETENV("TEMPDIR")
  CALL security.RandomGenerator.CreateUUIDString() RETURNING ls_GUID
  LET ls_sript_filename = ls_tempdir,ls_separator,cs_prc_get_col_default,"_",ls_GUID,".prc"
  
  LET ls_exit_sign = sadzi140_util_get_exit_sign() --cs_exit_sign --"exit;"
  
  LET lchannel_write = base.Channel.create()
  CALL lchannel_write.setDelimiter("")

  LET ls_create_script = "-- Create stored function to get column default value                    ","\n",
                         "SET SERVEROUTPUT ON                                                      ","\n",
                         "CREATE OR REPLACE FUNCTION GET_COL_DEFAULT                               ","\n",
                         "(                                                                        ","\n",
                         "  p_table_name  VARCHAR2,                                                ","\n",
                         "  p_column_name VARCHAR2                                                 ","\n",
                         ")                                                                        ","\n",
                         "RETURN VARCHAR2                                                          ","\n",
                         "AS                                                                       ","\n",
                         "  ls_text VARCHAR2(32767);                                               ","\n",
                         "  ls_SQL  VARCHAR2(2000);                                                ","\n",
                         "BEGIN                                                                    ","\n",
                         "                                                                         ","\n",
                         "  ls_SQL := 'SELECT DATA_DEFAULT                                 '||     ","\n",
                         "            '  FROM USER_TAB_COLUMNS                             '||     ","\n",
                         "            ' WHERE TABLE_NAME  = '''||UPPER(p_table_name)||'''  '||     ","\n",
                         "            '   AND COLUMN_NAME = '''||UPPER(p_column_name)||''' ';      ","\n",
                         "                                                                         ","\n",
                         "  EXECUTE IMMEDIATE ls_SQL INTO ls_text;                                 ","\n",
                         "                                                                         ","\n",
                         "  ls_text := TRIM(REPLACE(REPLACE(ls_text,CHR(13),''),CHR(10),''));      ","\n",
                         "                                                                         ","\n",
                         "  IF UPPER(NVL(ls_text,'",cs_null_default,"')) = 'NULL' THEN             ","\n",
                         "    ls_text := '';                                                       ","\n",
                         "  END IF;                                                                ","\n",
                         "                                                                         ","\n",
                         "  RETURN ls_text;                                                        ","\n",
                         "                                                                         ","\n",
                         "EXCEPTION                                                                ","\n",
                         "  WHEN OTHERS THEN RETURN '';                                            ","\n",
                         "END;                                                                     ","\n",
                         "/                                                                        ","\n",
                         "                                                                         ","\n",
                         "GRANT EXECUTE ON GET_COL_DEFAULT TO PUBLIC;                              ","\n",
                         "GRANT EXECUTE ON GET_COL_DEFAULT TO TIPTOP;                              ","\n",
                         "                                                                         "
                                                                                                                           
  LET ls_create_script = ls_create_script,"\n",
                         ls_exit_sign
  
  #開檔寫入
  TRY
    CALL lchannel_write.openFile(ls_sript_filename, "w" )
    CALL lchannel_write.write(ls_create_script)
    #關檔
    CALL lchannel_write.close()
  CATCH
    DISPLAY cs_error_tag,"Generate Script Error !!"
  END TRY  

  LET ls_return = ls_sript_filename
  
  RETURN ls_return

END FUNCTION

FUNCTION sadzi140_prc_ora_create_get_col_default_by_owner()
DEFINE
  ls_exit_sign      STRING,
  ls_GUID           STRING,
  ls_tempdir        STRING,
  lchannel_write    base.Channel,
  ls_sript_filename STRING,
  ls_create_script  STRING,
  ls_separator      STRING
DEFINE
  ls_return STRING  
  
  LET ls_separator = os.Path.separator()
  
  LET ls_tempdir = FGL_GETENV("TEMPDIR")
  CALL security.RandomGenerator.CreateUUIDString() RETURNING ls_GUID
  LET ls_sript_filename = ls_tempdir,ls_separator,cs_prc_get_col_default_by_owner,"_",ls_GUID,".prc"
  
  LET ls_exit_sign = sadzi140_util_get_exit_sign() --cs_exit_sign --"exit;"
  
  LET lchannel_write = base.Channel.create()
  CALL lchannel_write.setDelimiter("")

  LET ls_create_script = "-- Create stored function to get column default value                    ","\n",
                         "SET SERVEROUTPUT ON                                                      ","\n",
                         "CREATE OR REPLACE FUNCTION GET_COL_DEFAULT_BY_OWNER                      ","\n",
                         "(                                                                        ","\n",
                         "  p_owner       VARCHAR2,                                                ","\n",
                         "  p_table_name  VARCHAR2,                                                ","\n",
                         "  p_column_name VARCHAR2                                                 ","\n",
                         ")                                                                        ","\n",
                         "RETURN VARCHAR2                                                          ","\n",
                         "AS                                                                       ","\n",
                         "  ls_text VARCHAR2(32767);                                               ","\n",
                         "  ls_SQL  VARCHAR2(2000);                                                ","\n",
                         "BEGIN                                                                    ","\n",
                         "                                                                         ","\n",
                         "  ls_SQL := 'SELECT DATA_DEFAULT                                 '||     ","\n",
                         "            '  FROM ALL_TAB_COLUMNS                              '||     ","\n",
                         "            ' WHERE 1=1                                          '||     ","\n",
                         "            '   AND OWNER       = '''||UPPER(p_owner)||'''       '||     ","\n",
                         "            '   AND TABLE_NAME  = '''||UPPER(p_table_name)||'''  '||     ","\n",
                         "            '   AND COLUMN_NAME = '''||UPPER(p_column_name)||''' ';      ","\n",
                         "                                                                         ","\n",
                         "  EXECUTE IMMEDIATE ls_SQL INTO ls_text;                                 ","\n",
                         "                                                                         ","\n",
                         "                                                                         ","\n",
                         "  ls_text := TRIM(REPLACE(REPLACE(ls_text,CHR(13),''),CHR(10),''));      ","\n",
                         "                                                                         ","\n",
                         "  IF UPPER(NVL(ls_text,'",cs_null_default,"')) = 'NULL' THEN             ","\n",
                         "    ls_text := '';                                                       ","\n",
                         "  END IF;                                                                ","\n",
                         "                                                                         ","\n",
                         "  RETURN ls_text;                                                        ","\n",
                         "                                                                         ","\n",
                         "EXCEPTION                                                                ","\n",
                         "  WHEN OTHERS THEN RETURN '';                                            ","\n",
                         "END;                                                                     ","\n",
                         "/                                                                        ","\n",
                         "                                                                         ","\n",
                         "GRANT EXECUTE ON GET_COL_DEFAULT_BY_OWNER TO PUBLIC;                     ","\n",
                         "GRANT EXECUTE ON GET_COL_DEFAULT_BY_OWNER TO TIPTOP;                     ","\n",
                         "                                                                         "
                                                                                                                           
  LET ls_create_script = ls_create_script,"\n",
                         ls_exit_sign
  
  #開檔寫入
  TRY
    CALL lchannel_write.openFile(ls_sript_filename, "w" )
    CALL lchannel_write.write(ls_create_script)
    #關檔
    CALL lchannel_write.close()
  CATCH
    DISPLAY cs_error_tag,"Generate Script Error !!"
  END TRY  

  LET ls_return = ls_sript_filename
  
  RETURN ls_return

END FUNCTION

FUNCTION sadzi140_prc_check_procedure_if_exists(p_owner,p_object_name)
DEFINE
  p_owner        STRING,
  p_object_name  STRING
DEFINE
  ls_owner       STRING,
  ls_object_name STRING,
  ls_sql         STRING,
  li_rec_count   INTEGER,
  lb_return      BOOLEAN

  LET ls_owner       = p_owner.toUpperCase()
  LET ls_object_name = p_object_name.toUpperCase()

  #取得資料筆數
  LET ls_sql = "select count(*)                               ",
               "  from all_objects aos                        ",
               " where aos.owner = '",ls_owner,"'             ",
               "   and aos.object_name = '",ls_object_name,"' " 
                 
  PREPARE lpre_check_procedure_if_exists FROM ls_sql
  DECLARE lcur_check_procedure_if_exists CURSOR FOR lpre_check_procedure_if_exists
  OPEN lcur_check_procedure_if_exists
  FETCH lcur_check_procedure_if_exists INTO li_rec_count
  CLOSE lcur_check_procedure_if_exists
  FREE lcur_check_procedure_if_exists
  FREE lpre_check_procedure_if_exists

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION
#Oracle 專用 end