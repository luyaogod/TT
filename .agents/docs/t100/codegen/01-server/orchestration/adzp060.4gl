# Modify 180427-00055 : 20180601   by 02160   : 1.提供一個單純將報表備份檔解縮到報表主機的功能給 adzp063 呼叫

#&define DEBUG

IMPORT os
IMPORT util
IMPORT XML

SCHEMA ds

#匯入adzp080的常數設定  
&include "../4gl/sadzp000_cnst.inc" 
&include "../4gl/adzp080_cnst.inc" 
&include "../4gl/adzp080_type.inc"  

GLOBALS "../../cfg/top_global.inc"

CONSTANT cs_default_client_path STRING = "c:\\tt\\"

DEFINE
  ms_lang STRING

MAIN
  CALL adzp060_initialize()
  CALL adzp060_start()
  CALL adzp060_finalize()
END MAIN

FUNCTION adzp060_initialize()

  &ifndef DEBUG
  #CONNECT TO "ds"
  CALL cl_tool_init()
  &else
  CONNECT TO "local"
  &endif DEBUG

  LET ms_lang = NVL(g_lang,cs_default_lang)
  
  CALL ui.Dialog.setDefaultUnbuffered(TRUE)
  
END FUNCTION

FUNCTION adzp060_start()
DEFINE
  lo_parameters     T_UPLOAD_PARAM,
  lb_success        BOOLEAN,
  ls_message        STRING,  
  ls_err_code       STRING,
  ls_err_msg        STRING,
  ls_tempdir        STRING, #170626-00037
  lo_return_params  T_UPLOAD_PARAM
  
  &ifndef DEBUG
  LET lo_parameters.arg1_program_name    = ARG_VAL(2) #Program Name
  LET lo_parameters.arg2_module_name     = ARG_VAL(3) #Module Name
  LET lo_parameters.arg3_upload_doc_type = ARG_VAL(4) #Upload Document Type (SPEC or CODE)
  LET lo_parameters.arg4_client_path     = ARG_VAL(5) #Client Path
  &else
  LET lo_parameters.arg1_program_name    = "aiti333.tzp"
  LET lo_parameters.arg2_module_name     = "ait"
  LET lo_parameters.arg3_upload_doc_type = "PR"
  LET lo_parameters.arg4_client_path     = cs_default_client_path
  &endif

  IF (lo_parameters.arg1_program_name IS NULL) AND (NVL(lo_parameters.arg3_upload_doc_type,cs_null_default) <> cs_doc_type_4rp) THEN
    DISPLAY "\n",
            "Usage : ","\n",
            "  r.r ",ui.Interface.getName()," [ProgramName.ext] [ModuleName] [SPEC/CODE/4RP] [ClientPath]","\n",
            "  r.r ",ui.Interface.getName()," [ProgramName] [ModuleName] [4RPUPK] [ZipFilePath]","\n", #180427-00055
            "\n"
    EXIT PROGRAM
  END IF

  CALL sadzp060_run(lo_parameters.*,ms_lang) RETURNING lb_success,ls_message,ls_tempdir,lo_return_params.* #170626-00037
  
  LABEL _return:
  
  IF NOT lb_success THEN
    LET ls_err_code = "adz-00387"
    LET ls_err_msg  = ls_message,"|"
    CALL sadzp000_msg_show_error(ls_err_code, ls_err_code, ls_err_msg, 1)
    DISPLAY "[Failed] ",ui.Interface.getName()
  ELSE
    DISPLAY "[Success] ",ui.Interface.getName() 
  END IF
  
END FUNCTION

FUNCTION adzp060_finalize()
END FUNCTION

