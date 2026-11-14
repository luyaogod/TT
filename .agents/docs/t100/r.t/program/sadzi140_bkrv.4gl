
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

FUNCTION sadzi140_bkrv_make_hard_copy(p_hard_copy_info)
DEFINE
  p_hard_copy_info T_HARD_COPY_INFO
DEFINE
  lo_hard_copy_info T_HARD_COPY_INFO
DEFINE
  lb_return BOOLEAN  

  LET lo_hard_copy_info.* = p_hard_copy_info.*

  RETURN lb_return
  
END FUNCTION 