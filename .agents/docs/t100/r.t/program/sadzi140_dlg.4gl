
&include "../4gl/sadzi140_mcr.inc"

IMPORT os
IMPORT util

SCHEMA ds

&include "../4gl/sadzi140_cnst.inc"

&include "../4gl/sadzi140_type.inc"

&ifndef DEBUG
GLOBALS "../../cfg/top_global.inc"
&endif

FUNCTION sadzi140_dlg_set_dialog_parameter(p_param1,p_param2,p_param3,p_param4)
DEFINE
  p_param1,p_param2,p_param3,p_param4 STRING
DEFINE
  lo_open_dialog T_FILE_DIALOG
DEFINE
  lo_return T_FILE_DIALOG

  LET lo_open_dialog.PATH       = p_param1
  LET lo_open_dialog.TYPE_DESC  = p_param2
  LET lo_open_dialog.TYPE_LIST  = p_param3
  LET lo_open_dialog.CAPTION    = p_param4
  
  LET lo_return.* = lo_open_dialog.*

  RETURN lo_return.*
  
END FUNCTION 

FUNCTION sadzi140_dlg_open_file_dialog(p_open_dialog)
DEFINE
  p_open_dialog T_FILE_DIALOG
DEFINE
  ls_rtn_name  STRING  
DEFINE
  ls_return STRING   

  CALL ui.interface.frontCall("standard","openfile",[p_open_dialog.PATH,p_open_dialog.TYPE_DESC,p_open_dialog.TYPE_LIST,p_open_dialog.CAPTION],[ls_rtn_name])

  LET ls_return = ls_rtn_name

  RETURN ls_return
  
END FUNCTION 

FUNCTION sadzi140_dlg_save_file_dialog(p_save_dialog)
DEFINE
  p_save_dialog T_FILE_DIALOG
DEFINE
  ls_rtn_name STRING   
DEFINE
  ls_return STRING   

  CALL ui.interface.frontCall("standard","savefile",[p_save_dialog.PATH,p_save_dialog.TYPE_DESC,p_save_dialog.TYPE_LIST,p_save_dialog.CAPTION],[ls_rtn_name])

  LET ls_return = ls_rtn_name

  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_dlg_save_dir_dialog(p_save_dialog)
DEFINE
  p_save_dialog T_FILE_DIALOG
DEFINE
  ls_rtn_name STRING   
DEFINE
  ls_return STRING   

  CALL ui.interface.frontCall("standard","opendir",[p_save_dialog.PATH,p_save_dialog.CAPTION],[ls_rtn_name])

  LET ls_return = ls_rtn_name

  RETURN ls_return
  
END FUNCTION  