{
  異動型態    異動單號       日 期       異動者      異動內容
  ========= ============= ========== ========== ================================
  Initialize              20170714   Ernestliou 1.新增程式
}

&include "../4gl/sadzp090_mcr.inc"

IMPORT util 
IMPORT os 

SCHEMA ds

&include "../4gl/sadzp000_cnst.inc"
&include "../4gl/sadzp090_cnst.inc"

GLOBALS "../../cfg/top_global.inc"

&include "../4gl/sadzp000_type.inc"
&include "../4gl/sadzp090_type.inc"

#執行的環境變數
DEFINE 
  ms_lang         STRING,
  mb_fault        BOOLEAN,
  mo_arguments    T_ARGUMENTS,
  mb_args_exists  BOOLEAN
            
MAIN
DEFINE
  lb_result        BOOLEAN,
  ls_TAR_file_path STRING
     
  CALL adzp090_initialize()
  CALL adzp090_start()
  CALL adzp090_finalize()
    
END MAIN

FUNCTION adzp090_initialize()
DEFINE
  li_args        INTEGER,
  li_loop        INTEGER,
  ls_args        STRING,
  lb_result      BOOLEAN,  
  li_total_args  INTEGER,
  lb_args_exists BOOLEAN

  LET lb_args_exists = TRUE
  
  &ifndef DEBUG
    #依模組進行系統初始化設定(系統設定)
    CALL cl_tool_init()
    CALL cl_db_connect("ds", TRUE)
    #CONNECT TO "ds"
    LET ms_lang = NVL(g_lang,cs_default_lang)
  &else
    CONNECT TO "local"
    LET ms_lang = cs_default_lang
  &endif

  #設定參數
  LET li_args = NUM_ARGS()

  #先作參數的檢核
  LET li_total_args = 0
  FOR li_loop = 1 TO li_args
    LET ls_args = UPSHIFT(ARG_VAL(li_loop))
    IF ls_args LIKE "-%" THEN
      LET li_total_args = li_total_args + 1
    END IF
  END FOR 
  IF li_total_args = 0 THEN 
    LET lb_args_exists = FALSE 
  END IF

  #參數有存在或正確的時候, 才會進行實際參數的指定
  IF lb_args_exists THEN
    FOR li_loop = 1 TO li_args
      LET ls_args = UPSHIFT(ARG_VAL(li_loop))

      #驗證參數
      CALL adzp090_check_arguments(ls_args,ARG_VAL(li_loop + 1)) RETURNING lb_result
      IF NOT lb_result THEN LET mb_fault = TRUE EXIT FOR END IF
      
      CASE 
        WHEN ls_args = cs_args_table_name 
          LET mo_arguments.a_args1 = ARG_VAL(li_loop + 1)
      END CASE 
      
    END FOR
  END IF  

  #本隻程式先設定有參數存在
  LET lb_args_exists = TRUE 
  LET mb_args_exists = lb_args_exists

  #驗證參數失敗, 直接結束
  IF mb_fault THEN CALL adzp090_finalize() END IF
  
END FUNCTION

FUNCTION adzp090_start()
DEFINE
  lo_left_array  DYNAMIC ARRAY OF T_TABLE_LIST,
  lo_right_array DYNAMIC ARRAY OF T_TABLE_LIST,
  lb_result       BOOLEAN

  IF mb_args_exists THEN 
    CALL sadzp090_run(lo_left_array,lo_right_array,mo_arguments.*) RETURNING lb_result
  ELSE 
    CALL adzp090_help()
  END IF  
  
END FUNCTION

FUNCTION adzp090_finalize()

  IF mb_fault THEN
    EXIT PROGRAM -1 
  ELSE
    EXIT PROGRAM 
  END IF
  
END FUNCTION

FUNCTION adzp090_help()

  DISPLAY "\n",
          "Usage :","\n",
          "  r.r adzp090","\n"
          
END FUNCTION 

FUNCTION adzp090_check_arguments(p_argument_type,p_value)
DEFINE
  p_argument_type STRING,
  p_value         STRING
DEFINE
  ls_argument_type STRING,
  ls_value         STRING
DEFINE
  lb_return BOOLEAN

  LET ls_argument_type = p_argument_type
  LET ls_value = p_value

  LET lb_return = TRUE

  CASE 
    WHEN ls_argument_type = cs_args_table_name
  END CASE 

  RETURN lb_return

END FUNCTION