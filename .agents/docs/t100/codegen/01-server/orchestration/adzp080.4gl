# Modify              : 2015/11/09 by Hiko    : 調整有錯誤的訊息顯現機制
# Modify 160310-00002 : 2016/03/10 by Ernest  : 上傳錯誤時回傳錯誤訊息
# Modify 160922-00039 : 2016/09/26 by Ernest  : 配合本版次調整更新呼叫sadzp060_8_modify_curr_ver
# Modify 170503-00048 : 2017/05/03 by Ernest  : 調整當progress bar被後端關閉後, 仍能顯示progress資訊在背景
# Modify 170515-00013 : 2017/05/13 by Ernest  : 新增上傳預編譯型態 C4GL,C4FD, (adz-01198,adz-01199)
# Modify 170612-00043 : 2017/06/12 by Ernest  : 1.辨識 adzp040 的執行狀況  
# Modify 170626-00037 : 2017/06/29 by Ernest  : 上傳解壓縮後, 將tsd2/tap2更名成tsd/tap後才開始跑原本更新設計資料的流程
# Modify 170720-00034 : 2017/07/20 by ernest: 1.下載的時候紀錄host name.
#                                             2.上傳檢查權限之後,先判斷上傳host name是否同上次的上傳紀錄:
#                                               Y:繼續上傳, 同時記錄上傳的host name.
#                                               N:判斷是否同下載紀錄:
#                                                 Y:繼續上傳, 同時記錄上傳的host name.
#                                                 N:提示訊息:系統偵測到...... 若現在上傳, 有可能會造成上一位修改者的程式被覆蓋..請問您是否要繼續上傳?
#                                             3.簽出時, 要將相關紀錄完全清空.
# Modify 170904-00003 : 2017/09/04 by ernest: 1.1.修改 adzp080 上傳時檢查 TOPCHKOUT , 若為 N, 就不可以上傳.
# Modify 180105-00019 : 2018/01/05 by ernest: 1.緊急修繕上傳permission判斷錯誤
# Modify 180327-00002   : 20180327 by 02160 : 1.異動的時機點寫入記錄表，並提供查詢功能
# Modify 180904-00038   : 20180904 by 02785 : 上傳預編譯型態C4GL增加呼叫檢查語法工具(src_standard)
# Modify 180920-00008   : 20180920 by 02160 : 1.adzp080在預編譯時不須呼叫 adzq058
# Modify 181127-00049   : 20181127 by 02160 : 1.上傳失敗時, 於最前段錯誤訊息中顯示 : "程式編譯及連結程序失敗, 請搜尋error/ERROR確認原因" 以供使用者確認問題點
# Modify 181130-00028   : 20181130 by 02160 : 1.上傳 type 為 4rp 時不需要呼叫 adzp058
# Modify 200215-00009   : 20200220 by 10553 : MSSQL Server相容性調整
# Modify 230109-00005#1 : 20230110 by 02160 : 新增支援 XR 報表
# Modify 230925-00019#3 : 20231030 by 02160 : 新增支援 SPECRC3
# Modify 241030-00015#1 : 20241030 by 02160 : 修正上傳規格/程式進度條不正確的問題
# Modify 250703-00023#1 : 20250703 by 02785 : 修正上傳規格/程式在adzq058沒有保存記錄導致查不到 (fix 230925-00019#3 bug)
#
#&define DEBUG

IMPORT util 
IMPORT os
IMPORT security #170515-00013 

SCHEMA ds

#181127-00049 begin
CONSTANT 
  cs_adzp080_success STRING = "[ADZP080_SUCCESS]",
  cs_adzp080_failed  STRING = "[ADZP080_FAILED]"
#181127-00049 end
  
&include "../4gl/sadzp000_cnst.inc"
&include "../4gl/adzp080_cnst.inc"
&include "../4gl/adzp080_type.inc"

#180327-00002 begin
&include "../4gl/sadzp200_type.inc" 
&include "../4gl/adzp058_type.inc"
#180327-00002 end

GLOBALS "../../cfg/top_global.inc"

#執行的環境變數
DEFINE 
  ms_lang         STRING,
  mi_log_line     INTEGER,
  ms_message      STRING,
  ms_all_progress STRING,
  ms_file_name    STRING,
  ms_ext_name     STRING,
  mo_parameters   T_UPLOAD_PARAM,
  ms_all_tasks    STRING,
  mb_run_normal   BOOLEAN,
  mb_success      BOOLEAN,
  ms_user         STRING, #170720-00034  
  mi_spend_time   DATETIME YEAR TO SECOND,
  ms_adjust_ver   STRING #160922-00039

            
MAIN
  CALL adzp080_initialize()
  IF mo_parameters.arg3_upload_doc_type <> cs_doc_type_4rp THEN #230109-00005#1 
    CALL adzp080_initial_form()
  END IF   
  CALL adzp080_start()
  CALL adzp080_finalize()
END MAIN

FUNCTION adzp080_initialize()
DEFINE 
  ls_program_name STRING,
  ls_message      STRING,
  ls_dzaf010      STRING,
  ls_err_code     STRING,
  ls_dgenv        STRING,
  lb_result       BOOLEAN 

  LET mb_run_normal = TRUE

  LET ls_dgenv = FGL_GETENV("DGENV")
  
  #Initialize parameters
  &ifndef DEBUG
    LET mo_parameters.arg1_program_name      = NVL(ARG_VAL(2),"NULL")  #Program Name                   
    LET mo_parameters.arg2_module_name       = NVL(ARG_VAL(3),"NULL")  #Module Name                    
    LET mo_parameters.arg3_upload_doc_type   = NVL(ARG_VAL(4),"''")      #Upload Document Type (SPEC or CODE)
    LET mo_parameters.arg4_client_path       = NVL(ARG_VAL(5),"''")       #Client Path
    LET ms_adjust_ver                        = NVL(ARG_VAL(6),"N")        #160922-00039            
    --LET mo_parameters.arg5_spec_version     = NVL(ARG_VAL(6),"''")      #SD Version Code                   
    --LET mo_parameters.arg6_code_version     = NVL(ARG_VAL(7),"''")      #PR Version Code  
    --LET mo_parameters.arg7_std_program      = NVL(ARG_VAL(8),"''")      #Standard Program
    --LET mo_parameters.arg8_std_spec_version = NVL(ARG_VAL(9),"''")      #StandardSD Version Code                   
    --LET mo_parameters.arg9_std_code_version = NVL(ARG_VAL(10),"''")     #StandardPR Version Code  
    --LET mo_parameters.arg10_dgenv           = NVL(ARG_VAL(11),ls_dgenv) #DGENV
  &else
    LET mo_parameters.arg1_program_name      = "aiti333.tzp"     #Program Name                   
    LET mo_parameters.arg2_module_name       = "ait"             #Module Name                    
    LET mo_parameters.arg3_upload_doc_type   = cs_doc_type_spec  #Upload Document Type (SPEC or CODE)
    LET mo_parameters.arg4_client_path       = "c:\\tt\\"        #Client Path                    
    LET ms_adjust_ver                        = "N"               #160922-00039     
    --LET mo_parameters.arg5_spec_version     = "1"               #SD Version Code                   
    --LET mo_parameters.arg6_code_version     = "1"               #PR Version Code
    --LET mo_parameters.arg7_std_program      = "''"              #Standard Program
    --LET mo_parameters.arg8_std_spec_version = "''"              #SD Version Code                   
    --LET mo_parameters.arg9_std_code_version = "''"              #PR Version Code  
    --LET mo_parameters.arg10_dgenv           = ls_dgenv          #DGENV
  &endif

  IF (
      mo_parameters.arg1_program_name    = "NULL" OR
      mo_parameters.arg2_module_name     = "NULL" OR
      mo_parameters.arg3_upload_doc_type = "''" OR
      mo_parameters.arg4_client_path     = "''" 
      --mo_parameters.arg5_spec_version    = "''" OR
      --mo_parameters.arg6_code_version    = "''" OR
      --mo_parameters.arg10_dgenv          = "''" 
     ) THEN 
    IF (mo_parameters.arg3_upload_doc_type <>  cs_doc_type_4rp) AND 
       (mo_parameters.arg3_upload_doc_type <>  cs_doc_type_xrp) THEN #230109-00005#1
      DISPLAY "\n",
              "Usage : ","\n",
              "  r.r ",ui.Interface.getName()," [ProgramName.ext] [ModuleName] [SPEC/CSPEC/SPECGEN/CODE/C4GL/C4FD] [ClientPath]","\n",
              "  r.r ",ui.Interface.getName()," '' '' [4RP/XRP] [ClientPath] [ClientLang]","\n"
      LET mb_run_normal = FALSE
      CALL adzp080_finalize()
    END IF   
  END IF

  IF (mo_parameters.arg3_upload_doc_type <> cs_doc_type_4rp) AND 
     (mo_parameters.arg3_upload_doc_type <> cs_doc_type_xrp) THEN #230109-00005#1 
    LET ls_program_name = mo_parameters.arg1_program_name
    IF ls_program_name.getIndexOf(".",1) > 1 THEN
      LET ms_file_name = ls_program_name.subString(1,ls_program_name.getIndexOf(".",1)-1)
      LET ms_ext_name  = ls_program_name.subString(ls_program_name.getIndexOf(".",1)+1,ls_program_name.getLength())
    ELSE
      LET mb_success = FALSE
      LET ls_message = cs_error_tag,"No define compressed file extension name(.tzs,.tzc,.tzr ...) !!"
      LET ms_message = ms_message,ls_message 
      CALL adzp080_finalize()
    END IF
  END IF    

  &ifndef DEBUG
    #依模組進行系統初始化設定(系統設定)
    CALL cl_tool_init()
    #CONNECT TO "ds"
    IF mo_parameters.arg3_upload_doc_type = cs_doc_type_xrp THEN #230109-00005#1 begin
      LET ms_lang = NVL(ARG_VAL(6), NVL(g_lang,cs_default_lang)) 
    ELSE  #230109-00005#1 end
      LET ms_lang = NVL(g_lang,cs_default_lang)
    END IF #230109-00005#1 
  &else
    CONNECT TO "local"
    LET ms_lang = cs_default_lang
  &endif

  #背景執行
  LET g_bgjob = "N" 

  #Log 行數重設
  LET mi_log_line = 0
  LET ms_message = ""

  #目前跑兩個程序
  #LET ms_all_tasks = "2"
  #160922-00039 begin 
  IF ms_adjust_ver = "Y" THEN
    LET ms_all_tasks = "3"
  ELSE
    LET ms_all_tasks = "2"
  END IF
  #160922-00039 end 

  LET ms_user = g_account #170720-00034  
  
END FUNCTION

FUNCTION adzp080_initial_form()
DEFINE
  ls_erp_path STRING,
  ls_cfg_path STRING,
  ls_4st_path STRING

  CALL FGL_GETENV(cs_erp_path) RETURNING ls_erp_path
  
  &ifndef DEBUG
    OPEN WINDOW w_adzp080 WITH FORM cl_ap_formpath("ADZ","adzp080") 
    ATTRIBUTE(STYLE="progress_bar", TEXT="PROGRESSING")
    CLOSE WINDOW SCREEN
    #CONNECT TO cs_master_db
  &else
    OPEN WINDOW w_adzp080 WITH FORM "adzp080"
    ATTRIBUTE(STYLE="popup")
    CLOSE WINDOW SCREEN
    #CONNECT TO "local"
  &endif

  CURRENT WINDOW IS w_adzp080

  #瀏覽頁簽資料初始化
  &ifndef DEBUG
  #CALL cl_ui_init()
  LET ls_cfg_path = os.Path.join(FGL_GETENV("ERP"), "cfg")
  LET ls_4st_path = os.Path.join(os.Path.join(os.Path.join(ls_cfg_path, "4st"), ms_lang), "designer.4st")
  #CALL ui.Interface.loadStyles(ls_erp_path||"/cfg/4st/adzp080")
  CALL ui.Interface.loadStyles(ls_4st_path)
  &endif

  CALL adzp080_set_window_title(ms_lang) #241030-00015

END FUNCTION

FUNCTION adzp080_start()
DEFINE
  lb_success        BOOLEAN,
  ls_message        STRING,
  ls_main_message   STRING,
  ls_err_code       STRING,
  ls_err_msg        STRING,
  li_tasks          INTEGER,
  ls_tasks          STRING,
  li_all_tasks      INTEGER,
  li_tasks_percent  INTEGER,
  ls_tasks_progress STRING,
  ls_temp_dir       STRING,  #170626-00037
  ls_separator      STRING,  #170626-00037
  lb_result         BOOLEAN, #170626-00037
  ls_topchkout      STRING,  #170904-00003
  li_time_start     DATETIME YEAR TO SECOND,
  li_time_end       DATETIME YEAR TO SECOND  

  LET ls_separator = os.Path.separator() #170626-00037
  LET lb_success = TRUE #170904-00003
  
  LET ls_message = "[Start] ","[",cl_get_current(),"] ",ui.Interface.getName(),"\n"
  DISPLAY ls_message
  #LET ls_message = ls_message, "mo_parameters.arg5_spec_version : ",mo_parameters.arg5_spec_version,"\n"
  #LET ls_message = ls_message, "mo_parameters.arg6_code_version : ",mo_parameters.arg6_code_version,"\n"
  LET ms_message = ms_message,ls_message 

  #170904-00003 begin
  LET ls_topchkout = NVL(FGL_GETENV("TOPCHKOUT"),"N")
  IF ls_topchkout = "N" THEN
    LET lb_success = FALSE
    LET ls_err_code = "adz-01238"
    LET ls_err_msg  = "|"
    CALL sadzp000_msg_show_error(ls_err_code, ls_err_code, ls_err_msg, 1)
    CALL adzp080_finalize()
  END IF
  #170904-00003 end
  
  LET li_tasks = 0

  #上傳壓縮包及解縮
  IF mo_parameters.arg3_upload_doc_type <> cs_doc_type_4rp AND
     mo_parameters.arg3_upload_doc_type <> cs_doc_type_xrp THEN #230109-00005#1
    LET li_tasks = li_tasks + 1
    LET ls_tasks = li_tasks
    DISPLAY ms_all_tasks TO formonly.ed_all_perccent
    DISPLAY ls_tasks TO formonly.ed_all_single
    LET li_all_tasks = ms_all_tasks
    LET li_tasks_percent = ((100*li_tasks) / li_all_tasks)
    LET ls_tasks_progress = li_tasks_percent
    LET ls_main_message = "Upload package and uncompress"
    DISPLAY ls_main_message TO formonly.lbl_main_message
    CALL adzp080_refresh_tasks_progressbar(ls_tasks_progress)
  END IF   
  LET li_time_start = cl_get_current()
  CALL adzp080_upload_package() RETURNING lb_success,ls_message,ls_temp_dir  #170626-00037
  LET li_time_end = cl_get_current()
  DISPLAY cs_information_tag,"Upload Package START Time : ",li_time_start  
  DISPLAY cs_information_tag,"Upload Package  END Time : ",li_time_end  
  LET mi_spend_time = mi_spend_time + (li_time_end - li_time_start)
  IF NOT lb_success THEN
    LET ls_err_code = "adz-00152"
    LET ls_err_msg  = ls_message,"|" -- 160310-00002
    DISPLAY cs_error_tag,"[",ls_err_code,"]",ls_message -- 160310-00002
    &ifndef DEBUG
    # 160310-00002 begin
      INITIALIZE g_errparam TO NULL
         LET g_errparam.code = "adz-00152"
         LET g_errparam.extend = ls_message
         LET g_errparam.popup = TRUE
         CALL cl_err()
    # 160310-00002 end
    -- CALL sadzp000_msg_show_error(ls_err_code, ls_err_code, ls_err_msg, 1) -- 160310-00002 Mark
    &else
    DISPLAY ls_err_code
    &endif
  END IF   
  LET ms_message = ms_message,ls_message 

  #本版次調整更新及編譯及連結(4RP,C4GL,C4FD 不做)
  IF lb_success AND 
     (mo_parameters.arg3_upload_doc_type <> cs_doc_type_4rp) THEN 
    
    #160922-00039 begin
    IF lb_success AND (ms_adjust_ver = "Y") THEN

        LET li_tasks = li_tasks + 1
        LET ls_tasks = li_tasks
        DISPLAY ms_all_tasks TO formonly.ed_all_perccent
        DISPLAY ls_tasks TO formonly.ed_all_single
        LET li_all_tasks = ms_all_tasks
        LET li_tasks_percent = ((100*li_tasks) / li_all_tasks)
        LET ls_tasks_progress = li_tasks_percent
        LET ls_main_message = "Adjust update for this version."
        DISPLAY ls_main_message TO formonly.lbl_main_message
        CALL adzp080_refresh_tasks_progressbar(ls_tasks_progress)
         
      IF (mo_parameters.arg3_upload_doc_type <> cs_doc_type_c4gl) AND 
         (mo_parameters.arg3_upload_doc_type <> cs_doc_type_c4fd) THEN #170515-00013 add
        LET li_time_start = cl_get_current()
        CALL adzp080_adjust_version_update() RETURNING lb_success,ls_message
        LET li_time_end = cl_get_current()
        DISPLAY cs_information_tag,"Adjust version START Time : ",li_time_start  
        DISPLAY cs_information_tag,"Adjust version  END Time : ",li_time_end  
        LET mi_spend_time = mi_spend_time + (li_time_end - li_time_start)
      END IF #170515-00013 add
      
    END IF
    #160922-00039 end
    
    IF lb_success THEN
      LET li_tasks = li_tasks + 1
      LET ls_tasks = li_tasks
      DISPLAY ms_all_tasks TO formonly.ed_all_perccent
      DISPLAY ls_tasks TO formonly.ed_all_single
      LET li_all_tasks = ms_all_tasks
      LET li_tasks_percent = ((100*li_tasks) / li_all_tasks)
      LET ls_tasks_progress = li_tasks_percent
      #170515-00013 begin
      IF (mo_parameters.arg3_upload_doc_type = cs_doc_type_c4gl) OR (mo_parameters.arg3_upload_doc_type = cs_doc_type_c4fd) THEN
        LET ls_main_message = "Precompile object ..."
      ELSE
      #170515-00013 end
        LET ls_main_message = "Build and link package ..."
      END IF #170515-00013 add
      DISPLAY ls_main_message TO formonly.lbl_main_message
      CALL adzp080_refresh_tasks_progressbar(ls_tasks_progress)
       
      LET li_time_start = cl_get_current()
      #重新 Build 程式
      CALL adzp080_build_package() RETURNING lb_success,ls_message
      LET li_time_end = cl_get_current()
      DISPLAY cs_information_tag,"Build Package or Precompile START Time : ",li_time_start  
      DISPLAY cs_information_tag,"Build Package or Precompile END Time : ",li_time_end  
      LET mi_spend_time = mi_spend_time + (li_time_end - li_time_start)

      #170626-00037 begin
      CASE
        WHEN (mo_parameters.arg3_upload_doc_type = cs_doc_type_code) OR 
             (mo_parameters.arg3_upload_doc_type = cs_doc_type_xrp) #230109-00005#1
          CALL sadzp060_move_tap_file(mo_parameters.arg2_module_name,mo_parameters.arg1_program_name,ls_temp_dir||ls_separator) RETURNING lb_result 
        WHEN mo_parameters.arg3_upload_doc_type = cs_doc_type_spec 
          CALL sadzp060_move_tsd_file(mo_parameters.arg2_module_name,mo_parameters.arg1_program_name,ls_temp_dir||ls_separator) RETURNING lb_result
        #230925-00019#3 begin   
        WHEN mo_parameters.arg3_upload_doc_type = cs_doc_type_specrc3  
          CALL sadzp060_move_tsd_file(mo_parameters.arg2_module_name,mo_parameters.arg1_program_name,ls_temp_dir||ls_separator) RETURNING lb_result 
        #230925-00019#3 end   
      END CASE
      #170626-00037 end      
      
    END IF
    
    IF NOT lb_success THEN
      #Begin:2015/11/09 by Hiko
      #LET ls_err_code = "adz-00153"
      #LET ls_err_msg  = "|"
      #&ifndef DEBUG
      #CALL sadzp000_msg_show_error(ls_err_code, ls_err_code, ls_err_msg, 1)
      #&else
      #DISPLAY ls_err_code
      #&endif
      #181127-00049 begin
      LET ls_err_code = "adz-00153"
      CALL sadzp000_msg_get_message(ls_err_code,ms_lang) RETURNING ls_err_msg
      LET ls_message = "\n",ls_err_msg,"\n\n",ls_message
      #181127-00049 end
      INITIALIZE g_errparam TO NULL
         LET g_errparam.code = "!"
         LET g_errparam.extend = ls_message
         LET g_errparam.popup = TRUE
         CALL cl_err()
      #End:2015/11/09 by Hiko
    END IF
    
    LET ms_message = ms_message,ls_message 
    
  END IF    

  IF (mo_parameters.arg3_upload_doc_type <> cs_doc_type_c4gl) AND (mo_parameters.arg3_upload_doc_type <> cs_doc_type_c4fd) THEN #170515-00013 add
    #顯示參數值供參考
    DISPLAY cs_parameter_tag,"arg1_program_name      : ",mo_parameters.arg1_program_name
    DISPLAY cs_parameter_tag,"arg2_module_name       : ",mo_parameters.arg2_module_name
    DISPLAY cs_parameter_tag,"arg3_upload_doc_type   : ",mo_parameters.arg3_upload_doc_type
    DISPLAY cs_parameter_tag,"arg4_client_path       : ",mo_parameters.arg4_client_path
    DISPLAY cs_parameter_tag,"arg5_spec_version      : ",mo_parameters.arg5_spec_version
    DISPLAY cs_parameter_tag,"arg6_code_version      : ",mo_parameters.arg6_code_version
    DISPLAY cs_parameter_tag,"arg7_std_program       : ",mo_parameters.arg7_std_program
    DISPLAY cs_parameter_tag,"arg8_std_spec_version  : ",mo_parameters.arg8_std_spec_version
    DISPLAY cs_parameter_tag,"arg9_std_code_version  : ",mo_parameters.arg9_std_code_version
    DISPLAY cs_parameter_tag,"arg10_dgenv            : ",mo_parameters.arg10_dgenv
    DISPLAY cs_parameter_tag,"arg11_type             : ",mo_parameters.arg11_type
  END IF #170515-00013 add   

  DISPLAY IIF(lb_success,cs_adzp080_success,cs_adzp080_failed),"[",cl_get_current(),"] ",ui.Interface.getName() #181127-00049
  
  IF lb_success THEN
    LET ls_message = "[SUCCESS] ","[",cl_get_current(),"] ",ui.Interface.getName()
    CALL adzp080_refresh_all_progressbar("100")
    SLEEP 1
  ELSE
    LET ls_message = "[FAILED] ","[",cl_get_current(),"] ",ui.Interface.getName()
  END IF   
  DISPLAY ls_message

  #DISPLAY cs_information_tag,"Upload package and build program Total Time : ",mi_spend_time #170503-00048 mark  
  
  LET ms_message = ms_message,ls_message 

  LET mb_success = lb_success
  
END FUNCTION

FUNCTION adzp080_finalize()
  
  TRY
    CLOSE WINDOW w_adzp080
  CATCH
  END TRY  
  
  TRY
    DISCONNECT CURRENT
  CATCH
  END TRY  
  
  EXIT PROGRAM
  
END FUNCTION

################################################################################

FUNCTION adzp080_upload_package()
DEFINE
  ls_command    STRING,
  lb_success    BOOLEAN,
  ls_message    STRING,
  ls_program    STRING,  
  lb_result       BOOLEAN,  #170720-00034
  ls_upload_type  STRING,   #170720-00034
  ls_ul_version   STRING,   #170720-00034 
  ls_err_msg      STRING,   #170720-00034
  ls_err_code     STRING,   #170720-00034
  ls_question     STRING,   #170720-00034
  ls_last_dzhi009 STRING,   #170720-00034
  lo_curr_dzhi_t  T_DZHI_T, #170720-00034
  lo_dl_dzhi_t    T_DZHI_T, #170720-00034
  lo_ul_dzhi_t    T_DZHI_T, #170720-00034
  lo_curr_DZAF_T   T_DZAF_T, #180327-00002
  ls_prog_name  STRING, #170515-00013
  ls_temp_dir   STRING,  #170626-00037
  lo_parameters T_UPLOAD_PARAM

  LET lb_success = TRUE
  LET ls_message = ""

  LET ls_program = "adzp060"

  #180327-00002 begin
  #取得現今DZAF的資料
  TRY 
    LET ls_prog_name = mo_parameters.arg1_program_name
    LET ls_prog_name = ls_prog_name.subString(1,ls_prog_name.getIndexOf("(",1)-1)
    LET lo_curr_DZAF_T.DZAF001 = ls_prog_name
    LET lo_curr_DZAF_T.DZAF005 = adzp080_get_dzaf_spec_type(ls_prog_name)
    LET lo_curr_DZAF_T.DZAF010 = FGL_GETENV("DGENV")
    CALL sadzp200_ver_get_curr_ver_info(lo_curr_DZAF_T.*) RETURNING lo_curr_DZAF_T.*
    #180920-00008 begin
    IF (mo_parameters.arg3_upload_doc_type <> cs_doc_type_c4gl) AND 
       (mo_parameters.arg3_upload_doc_type <> cs_doc_type_c4fd) AND
       (mo_parameters.arg3_upload_doc_type <> cs_doc_type_4rp) AND #181130-00028
       (mo_parameters.arg3_upload_doc_type <> cs_doc_type_xrp) THEN #230109-00005#1
      CALL adzp080_set_download_upload_log(lo_curr_DZAF_T.*,mo_parameters.arg3_upload_doc_type)
    END IF  
    #180920-00008 begin
  CATCH
    DISPLAY cs_warning_tag,"Set upload log not success."
  END TRY  
  #180327-00002 end
                  
  CALL adzp080_get_prog_progress_all(ls_program,ms_ext_name,mo_parameters.arg3_upload_doc_type) RETURNING ms_all_progress

  #檢核及取得剩下的參數
  CALL sadzp060_run(mo_parameters.*,ms_lang) RETURNING lb_success,ls_message,ls_temp_dir,lo_parameters.* #170626-00037

  #170515-00013 begin
  IF (mo_parameters.arg3_upload_doc_type = cs_doc_type_c4gl) OR (mo_parameters.arg3_upload_doc_type = cs_doc_type_c4fd) THEN
    LET ls_prog_name = mo_parameters.arg1_program_name
    LET ls_prog_name = ls_prog_name.subString(1,ls_prog_name.getIndexOf("(",1)-1)
    LET lo_parameters.arg1_program_name    = ls_prog_name
    LET lo_parameters.arg3_upload_doc_type = mo_parameters.arg3_upload_doc_type 
  END IF
  #170515-00013 end

  #170720-00034 begin
  IF (mo_parameters.arg3_upload_doc_type = cs_doc_type_spec) OR 
     (mo_parameters.arg3_upload_doc_type = cs_doc_type_specrc3) OR  #230925-00019#3
     (mo_parameters.arg3_upload_doc_type = cs_doc_type_code) THEN 
    CASE 
      WHEN mo_parameters.arg3_upload_doc_type = cs_doc_type_spec 
        LET ls_ul_version = lo_parameters.arg5_spec_version
      #230925-00019#3 begin  
      WHEN mo_parameters.arg3_upload_doc_type = cs_doc_type_specrc3 
        LET ls_ul_version = lo_parameters.arg5_spec_version
      ##230925-00019#3 end  
      WHEN mo_parameters.arg3_upload_doc_type = cs_doc_type_code 
        LET ls_ul_version = lo_parameters.arg6_code_version
    END CASE
    #取得當前要上傳的資訊及之前下載的資訊並做比對 
    CALL sadzp080_dlul_get_client_info(lo_parameters.arg1_program_name,ls_ul_version,"U",lo_parameters.arg3_upload_doc_type,ms_user) RETURNING lo_curr_dzhi_t.*
    CALL sadzp080_dlul_get_download_log_info(lo_curr_dzhi_t.DZHI001,lo_curr_dzhi_t.DZHI002,lo_curr_dzhi_t.DZHI007) RETURNING lo_dl_dzhi_t.*
    CALL sadzp080_dlul_get_upload_log_info(lo_curr_dzhi_t.DZHI001,lo_curr_dzhi_t.DZHI002,lo_curr_dzhi_t.DZHI007) RETURNING lo_ul_dzhi_t.*

    CASE
      #上傳資訊為空或是前一次上傳主機名稱等於當下的主機名稱
      WHEN (lo_ul_dzhi_t.DZHI001 IS NULL) OR (lo_curr_dzhi_t.DZHI009 = lo_ul_dzhi_t.DZHI009)
        CALL sadzp080_dlul_set_dzhi_t_data(lo_curr_dzhi_t.*) RETURNING lb_result
      #下傳資訊為空或是前一次下傳主機名稱等於當下的主機名稱
      WHEN (lo_dl_dzhi_t.DZHI001 IS NULL) OR (lo_curr_dzhi_t.DZHI009 = lo_dl_dzhi_t.DZHI009)  
        CALL sadzp080_dlul_set_dzhi_t_data(lo_curr_dzhi_t.*) RETURNING lb_result
    OTHERWISE    
      LET ls_last_dzhi009 = IIF(lo_curr_dzhi_t.DZHI009 <> lo_dl_dzhi_t.DZHI009,lo_dl_dzhi_t.DZHI009,IIF(lo_curr_dzhi_t.DZHI009 <> lo_ul_dzhi_t.DZHI009,lo_ul_dzhi_t.DZHI009, "N/A"))
      LET ls_err_msg  = lo_curr_dzhi_t.DZHI009,"|",ls_last_dzhi009,"|"
      LET ls_err_code = "adz-01226"
      CALL sadzp000_msg_alert_box(ls_err_code,ls_err_code,ls_err_msg,0) RETURNING ls_question
      IF ls_question = cs_response_yes THEN
        LET lb_result = TRUE
        CALL sadzp080_dlul_set_dzhi_t_data(lo_curr_dzhi_t.*) RETURNING lb_result
      ELSE
        LET lb_result = FALSE
      END IF 
    END CASE

    LET lb_success = IIF(lb_success,lb_result,lb_success) #180105-00019
    
  END IF
  #170720-00034 end
  
  #設定模組變數
  LET mo_parameters.* = lo_parameters.*

  { 
  LET ls_command = "r.r ",ls_program," ",
                   mo_parameters.arg1_program_name," ",   
                   mo_parameters.arg2_module_name," ",    
                   mo_parameters.arg3_upload_doc_type," ",
                   mo_parameters.arg4_client_path," "

  CALL adzp080_exec_prog(ls_program,ls_command) RETURNING lb_success,ls_message
  }
  
  RETURN lb_success,ls_message,ls_temp_dir  #170626-00037
  
END FUNCTION

FUNCTION adzp080_build_package()
DEFINE
  ls_command   STRING,
  lb_success   BOOLEAN,
  ls_message   STRING,
  ls_program   STRING,
  lb_result    BOOLEAN, #170515-00013
  ls_resmsg    STRING,  #170515-00013
  ls_msg_code  STRING,  #170515-00013 
  ls_msg_msg   STRING,  #170515-00013
  ls_result    STRING,  #170515-00013
  ls_prog_type STRING   #170515-00013
  
  LET lb_success = TRUE
  LET ls_message = ""

  IF (mo_parameters.arg3_upload_doc_type <> cs_doc_type_c4gl) AND (mo_parameters.arg3_upload_doc_type <> cs_doc_type_c4fd) THEN #170515-00013 add
    LET ls_program = "adzp040"

    CALL adzp080_get_prog_progress_all(ls_program,ms_ext_name,mo_parameters.arg3_upload_doc_type) RETURNING ms_all_progress
    
    LET ls_command = "r.r ",ls_program," ",
                     mo_parameters.arg1_program_name," ",   
                     mo_parameters.arg2_module_name," ",    
                     mo_parameters.arg3_upload_doc_type," ",
                     NVL(mo_parameters.arg5_spec_version,"''")," ",   
                     NVL(mo_parameters.arg6_code_version,"''")," ",
                     NVL(mo_parameters.arg7_std_program,"''")," ",
                     NVL(mo_parameters.arg8_std_spec_version,"''")," ",
                     NVL(mo_parameters.arg9_std_code_version,"''")," ",
                     NVL(mo_parameters.arg10_dgenv,"''")," ",
                     NVL(mo_parameters.arg11_type,"''")," ",
                     ms_lang  #230109-00005#1 

    CALL adzp080_exec_prog(ls_program,ls_command) RETURNING lb_success,ls_message
    
  #170515-00013 begin
  ELSE
    CASE
      WHEN mo_parameters.arg3_upload_doc_type = cs_doc_type_c4gl
        LET ls_prog_type = "4gl"
        CALL adzp800_pre_compile_program(mo_parameters.arg1_program_name,cs_doc_type_c4gl) RETURNING lb_result,ls_resmsg
      WHEN mo_parameters.arg3_upload_doc_type = cs_doc_type_c4fd
        LET ls_prog_type = "4fd"
        CALL adzp800_pre_compile_program(mo_parameters.arg1_program_name,cs_doc_type_c4fd) RETURNING lb_result,ls_resmsg
    END CASE
    
    IF NOT lb_result THEN 
      CALL adzp080_rev_view_logresult(ls_resmsg)
    END IF  
    
    LET lb_success = TRUE
    LET ls_message = ls_resmsg
  END IF  
  #170515-00013 end

  RETURN lb_success,ls_message

END FUNCTION

#160922-00039 begin 
FUNCTION adzp080_adjust_version_update()
DEFINE
  lb_success  BOOLEAN,
  ls_message  STRING,
  ls_version  STRING
  
  CASE
    WHEN mo_parameters.arg3_upload_doc_type = "SPEC"
      LET ls_version = mo_parameters.arg5_spec_version 
    WHEN mo_parameters.arg3_upload_doc_type = "CODE"
      LET ls_version = mo_parameters.arg6_code_version 
  END CASE

  #本版次調整更新
  CALL sadzp060_8_modify_curr_ver(mo_parameters.arg1_program_name, ls_version, mo_parameters.arg2_module_name, mo_parameters.arg3_upload_doc_type) RETURNING lb_success,ls_message
  
  RETURN lb_success,ls_message
  
END FUNCTION
#160922-00039 end 

FUNCTION adzp080_exec_prog(p_prog,p_command) #170612-00043
DEFINE 
  p_prog    STRING, #170612-00043
  p_command STRING  
DEFINE 
  ls_prog         STRING, #170612-00043
  ls_command      STRING,  
  lb_success      BOOLEAN, 
  ls_all_message  STRING,  
  lo_channel      base.Channel,
  ls_message      STRING,
  ls_err_message  STRING,
  ls_line_no      STRING,
  li_progress     INTEGER,
  ls_progress     STRING,
  ls_show_message STRING,
  ls_version_code STRING, 
  lb_record       BOOLEAN,
  li_all_progress INTEGER

  LET ls_prog = p_prog  #170612-00043
  LET ls_command = p_command
  
  LET ls_all_message = NULL
  LET lb_success = TRUE
  LET lb_record  = TRUE

  LET ls_show_message = "Waiting for next process (r.r ",ls_prog,") ..."  #170612-00043
  TRY 
    DISPLAY ls_show_message TO formonly.lbl_sub_message  #170612-00043
    CALL ui.Interface.refresh() #170612-00043
  CATCH
  END TRY   

  LET lo_channel = base.Channel.create()
  CALL lo_channel.setDelimiter("")

  DISPLAY cs_command_tag,ls_command
  LET ls_command = ls_command CLIPPED," 2>&1"

  CALL lo_channel.openPipe(ls_command,"r")   #執行指令
  IF STATUS THEN
    LET lb_success = FALSE
  ELSE
    WHILE TRUE
      CALL lo_channel.readLine() RETURNING ls_message
      IF lo_channel.isEof() THEN
        EXIT WHILE
      END IF

      #看到 IGNORE 就不紀錄
      IF ls_message.getIndexOf(cs_ignore_tag,1) THEN
        LET lb_record = FALSE
        CONTINUE WHILE
      END IF 

      #看到 NOTE 就紀錄
      IF ls_message.getIndexOf(cs_note_tag,1) THEN
        LET lb_record = TRUE
        CONTINUE WHILE
      END IF 

      IF lb_record THEN 
        IF ls_message.getIndexOf(cs_progress_tag,1) THEN
          LET li_progress = ls_message.subString(11,ls_message.getLength())
          LET li_all_progress = ms_all_progress
          IF li_all_progress < li_progress THEN
            LET li_all_progress = li_progress
          END IF
          TRY   
          DISPLAY li_progress TO formonly.ed_prog_single
          CATCH
          END TRY
          
          TRY   
          DISPLAY li_all_progress TO formonly.ed_prog_perccent
          CATCH
          END TRY  
          LET li_progress = ((100*li_progress) / li_all_progress)
          LET ls_progress = li_progress
          #DISPLAY "PROGRESS : ",ls_progress
          IF ls_progress = "100" THEN LET ls_progress = "96" END IF
          CALL adzp080_refresh_all_progressbar(ls_progress)
        ELSE   
          #取得訊息秀在訊息Bar上
          IF ls_message.getIndexOf(cs_message_tag,1) THEN
            LET ls_show_message = ls_message.subString(10,ls_message.getLength())
            TRY 
              DISPLAY ls_show_message TO formonly.lbl_sub_message
              CALL ui.Interface.refresh() #170612-00043
            CATCH
            END TRY   
          ELSE   
            LET mi_log_line = mi_log_line + 1
            LET ls_line_no = "[",(mi_log_line USING "&&&&"),"]",
                             "[",cl_get_current(),"] " 

            DISPLAY ls_line_no,ls_message   #顯示背景訊息
            LET ls_all_message = ls_all_message CLIPPED," ",
                                 ls_line_no,ls_message CLIPPED,"\n"
          END IF                       
        END IF
        
        #有錯誤訊息
        LET ls_err_message = ls_message.toUpperCase()
        IF ls_err_message.getIndexOf("ERROR" ,1) OR ls_err_message.getIndexOf("FAILED" ,1) THEN
          LET lb_success = FALSE
        END IF
      END IF   
    END WHILE
  END IF
  CALL lo_channel.close()

  RETURN lb_success, ls_all_message
  
END FUNCTION

FUNCTION adzp080_get_prog_progress_all(p_program,p_pgrs_type,p_doc_type)
DEFINE
  p_program   STRING,
  p_pgrs_type STRING,
  p_doc_type  STRING
DEFINE
  ls_program      STRING,
  ls_pgrs_type    STRING,
  ls_doc_type     STRING,
  ls_progress_all VARCHAR(5),
  li_progress_all INTEGER, --241030-00015
  ls_sql          STRING
DEFINE
  ls_return STRING  

  LET ls_program   = p_program
  LET ls_pgrs_type = p_pgrs_type
  LET ls_doc_type  = p_doc_type.toUpperCase()
  
  #Begin:200215-00009
  LET ls_sql = "select ej.dzej006 progress_all                                     ",
               "  from dzej_t ej                                                   ",
               " where ej.dzej001 = '",ls_program,"_Progress'                      ",  
               "   and ej.dzej003 = 'ALL'                                          ",
               "   and ej.dzej004 = '",ls_pgrs_type,"'                             ",
               "   and (                                                           ",
               "         coalesce(ej.dzej005,'X') = coalesce('",ls_doc_type,"','X')", --241030-00015
               "         or                                                        ", 
               "         ej.dzej005 is null                                        ",
               "       )                                                           "
  #End:200215-00009

  #DISPLAY "SQL : ",ls_sql               

  PREPARE lpre_get_prog_progress_all FROM ls_sql
  DECLARE lcur_get_prog_progress_all CURSOR FOR lpre_get_prog_progress_all

  OPEN lcur_get_prog_progress_all
  FETCH lcur_get_prog_progress_all INTO ls_progress_all
  CLOSE lcur_get_prog_progress_all
  
  FREE lpre_get_prog_progress_all
  FREE lcur_get_prog_progress_all  

  LET li_progress_all = NVL(ls_progress_all,"7") --241030-00015
  LET li_progress_all = li_progress_all + 1  --241030-00015

  LET ls_return = li_progress_all --241030-00015
  LET ls_return = ls_return.trim() --241030-00015
  
  RETURN ls_return
  
END FUNCTION

FUNCTION adzp080_refresh_all_progressbar(pi_position)
DEFINE 
  pi_position INTEGER
DEFINE 
  pbr_progress INTEGER  

  TRY #170503-00048
    IF pi_position <> 0 THEN 
      LET pbr_progress = pi_position
      DISPLAY BY NAME pbr_progress
    END IF 

    CALL ui.Interface.refresh()
  CATCH #170503-00048
    DISPLAY cs_warning_tag,"Could not get UI for display all progress, display on terminal." #170503-00048
  END TRY #170503-00048

  DISPLAY cs_message_tag,"Upload progress for all : ",pi_position #170503-00048

  #SLEEP 1
  
END FUNCTION

FUNCTION adzp080_refresh_tasks_progressbar(pi_position)
DEFINE 
  pi_position INTEGER
DEFINE 
  pbr_tasks INTEGER  

  TRY #170503-00048
    IF pi_position <> 0 THEN 
      LET pbr_tasks = pi_position
      DISPLAY BY NAME pbr_tasks
    END IF 

    CALL ui.Interface.refresh()
  CATCH #170503-00048
    DISPLAY cs_warning_tag,"Could not get UI for display tasks progress, display on terminal." #170503-00048
  END TRY  #170503-00048 

  DISPLAY cs_message_tag,"Upload progress for tasks : ",pi_position #170503-00048
  
END FUNCTION

#觀看 Log
FUNCTION adzp080_rev_view_logresult(p_message)
DEFINE
  p_message  STRING
DEFINE  
  ls_message  STRING
DEFINE  
  txt_LogResult STRING

  LET txt_LogResult = p_message

  CALL adzp080_rev_open_logresult_form()

  DISPLAY BY NAME txt_LogResult
  
  MENU
    ON ACTION btn_ok
      EXIT MENU  

    ON ACTION CLOSE
      EXIT MENU
  END MENU

  CALL adzp080_rev_close_logresult_form()
  
END FUNCTION

#開啟看 Log 的Form
FUNCTION adzp080_rev_open_logresult_form()
#170515-00013 begin
DEFINE
  ls_cfg_path  STRING,
  ls_4st_path  STRING
#170515-00013 end

  &ifndef DEBUG
    OPEN WINDOW w_adzp080_log WITH FORM cl_ap_formpath("ADZ","adzp080_log")
  &else
    OPEN WINDOW w_adzp080_log WITH FORM "adzp080_log"
  &endif
  
  CURRENT WINDOW IS w_adzp080_log

  #170515-00013 begin
  LET ls_cfg_path = os.Path.join(FGL_GETENV("ERP"), "cfg")
  LET ls_4st_path = os.Path.join(os.Path.join(os.Path.join(ls_cfg_path, "4st"), g_lang), "designer.4st")
  CALL ui.Interface.loadStyles(ls_4st_path)
  #170515-00013 end
  
END FUNCTION

#關閉看 Log 的Form
FUNCTION adzp080_rev_close_logresult_form()

  CLOSE WINDOW w_adzp080_log
  
END FUNCTION

#170515-00013 begin
FUNCTION adzp800_pre_compile_program(p_prog_name,p_prog_type)
DEFINE
  p_prog_name  STRING,
  p_prog_type  STRING
DEFINE
  ls_prog_name      STRING,
  ls_prog_type      STRING,
  ls_precompile_log STRING,
  ls_GUID           STRING,
  ls_command        STRING,
  lb_result         BOOLEAN,
  ls_resmsg         STRING,
  ls_pwd            STRING,
  ls_prog_full_name STRING,
  lb_compile_status BOOLEAN
DEFINE
  lb_return BOOLEAN,
  ls_return STRING  

  LET ls_prog_name = p_prog_name
  LET ls_prog_type = p_prog_type

  LET lb_compile_status = TRUE
  LET ls_pwd = os.Path.pwd()
  CALL os.Path.chdir(FGL_GETENV("TEMPDIR")) RETURNING lb_result

  CALL security.RandomGenerator.CreateUUIDString() RETURNING ls_GUID
  LET ls_precompile_log = os.Path.join(FGL_GETENV("TEMPDIR"),'adzp080_pre_compile'||ls_GUID||".log")
  DISPLAY cs_message_tag,"Precompile log on : ",ls_precompile_log 

  CASE
    WHEN ls_prog_type = cs_doc_type_c4gl
      LET ls_prog_full_name = ls_prog_name,".4gl"
    #180904-00038 begin
     #LET ls_command = "r.c '",ls_prog_name,"' TEMP | tee '",ls_precompile_log,"'"
      LET ls_command = "r.c '",ls_prog_name,"' TEMP 2>&1 | tee -a '",ls_precompile_log,"'"
      IF FGL_GETENV("DGENV") = "s" AND FGL_GETENV("LOGNAME") <> "topstd" THEN #標準環境下檢查
         LET ls_command = "$FGLRUN $FBIN/42m/chk_displayinfield $TEMPDIR/",ls_prog_full_name," 2>&1 | tee -a '",ls_precompile_log,"'",";",ls_command
         LET ls_command = "$FGLRUN $FBIN/42m/src_standard $TEMPDIR/",ls_prog_full_name," 2>&1 | tee -a '",ls_precompile_log,"'",";",ls_command
      END IF
    #180904-00038 end
    WHEN ls_prog_type = cs_doc_type_c4fd
      LET ls_prog_full_name = ls_prog_name,".4fd"
      LET ls_command = "r.f '",ls_prog_name,"' TEMP | tee '",ls_precompile_log,"'"
  END CASE 
  
  CALL os.Path.chrwx(ls_prog_full_name,511) RETURNING lb_result
  DISPLAY cs_message_tag,"Change file permission to full access: ",IIF(lb_result,"SUCCESS","FALSE")

  DISPLAY cs_command_tag,ls_command
  CALL cl_cmdrun_openpipe("r.r",ls_command,FALSE) RETURNING lb_result,ls_resmsg
  --RUN ls_command IN LINE MODE
  CALL adzp080_check_precompile_result(ls_precompile_log) RETURNING lb_result,ls_resmsg
  LET lb_compile_status = lb_result

  CALL os.Path.delete(ls_precompile_log) RETURNING lb_result
  DISPLAY cs_message_tag,"Delete log file : ",ls_precompile_log," => ",IIF(lb_result,"SUCCESS","FALSE")
  CALL os.Path.chdir(ls_pwd) RETURNING lb_result

  LET lb_return = lb_compile_status
  LET ls_return = ls_resmsg
  
  RETURN lb_return,ls_return
  
END FUNCTION

FUNCTION adzp080_check_precompile_result(p_precompile_log)
DEFINE
  p_precompile_log STRING
DEFINE
  ls_precompile_log STRING,
  lb_success        BOOLEAN,
  li_line_count     INTEGER,
  ls_text           STRING,
  lb_error          BOOLEAN,
  lo_channel_read   base.Channel
DEFINE
  lb_return  BOOLEAN,
  ls_return  STRING  
  
  LET ls_precompile_log  = p_precompile_log
  
  LET lo_channel_read = base.Channel.create()
  LET lb_success = TRUE
  LET lb_error   = FALSE
  
  TRY 
    CALL lo_channel_read.openFile(ls_precompile_log,"r")
  CATCH
    LET lb_success = FALSE
    DISPLAY cs_error_tag,"Open log file ",ls_precompile_log," fault !!"
  END TRY  

  IF lb_success THEN
  
    LET li_line_count = 1
    
    WHILE TRUE
      IF lo_channel_read.isEof() THEN 
        EXIT WHILE
      END IF

      LET ls_text = lo_channel_read.readLine()
      LET ls_return = ls_return,"\n",
                      ls_text
      LET ls_text = ls_text.toUpperCase()                
      IF ls_text.getIndexOf('ERROR',1) > 0 THEN
        LET lb_error = TRUE
      END IF 

      LET li_line_count = li_line_count + 1
      
    END WHILE
    
  END IF
  
  CALL lo_channel_read.close()
  
  IF lb_error THEN
    LET lb_return = FALSE 
  ELSE
    LET lb_return = TRUE 
  END IF 
  
  RETURN lb_return,ls_return
  
END FUNCTION
#170515-00013 end

#180327-00002 begin
FUNCTION adzp080_set_download_upload_log(p_DZAF_T,p_upload_doc_type)
DEFINE
  p_DZAF_T           T_DZAF_T,
  p_upload_doc_type  STRING
DEFINE  
  lo_DZAF_T  T_DZAF_T,
  ls_upload_doc_type STRING,
  ls_js      STRING,
  ls_action  STRING
#呼叫cl_cmdrun用
DEFINE la_param RECORD
         prog   STRING,
         param  DYNAMIC ARRAY OF STRING
       END RECORD
#呼叫adzp058用
DEFINE 
  obj_adzp058     util.JSONObject,
  lr_adzp058_info type_adzp058_info

  LET lo_DZAF_T.*        = p_DZAF_T.*
  LET ls_upload_doc_type = p_upload_doc_type
  
  #LET ls_action = IIF(ls_upload_doc_type==cs_doc_type_spec,cs_spec_up,cs_code_up)
#250703-00023 mark(s)
# #230925-00019#3 begin
# CASE
#   WHEN ls_upload_doc_type=cs_doc_type_spec
#     LET ls_upload_doc_type=cs_doc_type_spec
#   WHEN ls_upload_doc_type=cs_doc_type_specrc3
#     LET ls_upload_doc_type=cs_doc_type_spec
# OTHERWISE  
#   LET ls_upload_doc_type=cs_code_up
# END CASE
# #230925-00019#3 end  
#250703-00023 mark(e)

  #250703-00023 begin
  CASE
    WHEN ls_upload_doc_type=cs_doc_type_spec
      LET ls_upload_doc_type=cs_spec_up
    WHEN ls_upload_doc_type=cs_doc_type_specrc3
      LET ls_upload_doc_type=cs_spec_up
    OTHERWISE  
      LET ls_upload_doc_type=cs_code_up
  END CASE
  LET ls_action = ls_upload_doc_type
  #250703-00023 end

  LET lr_adzp058_info.action               = ls_action  #cs_spec_up/cs_code_up/cs_spec_checkout/cs_code_checkout
  LET lr_adzp058_info.key                  = NULL
  LET lr_adzp058_info.proglist[1].prog     = lo_DZAF_T.DZAF001 -- "aiti456"
  LET lr_adzp058_info.proglist[1].TYPE     = lo_DZAF_T.DZAF005 -- "M"
  LET lr_adzp058_info.proglist[1].cons_ver = lo_DZAF_T.DZAF002 -- 2
  LET lr_adzp058_info.proglist[1].spec_ver = lo_DZAF_T.DZAF003 
  LET lr_adzp058_info.proglist[1].code_ver = lo_DZAF_T.DZAF004
  LET lr_adzp058_info.proglist[1].identity = lo_DZAF_T.DZAF010 -- "s"

  LET obj_adzp058 = util.JSONObject.CREATE()
  LET obj_adzp058 = util.JSONObject.fromFGL(lr_adzp058_info)
 
  #呼叫cl_cmdrun
  LET la_param.prog = "adzp058"
  LET la_param.param[1] = obj_adzp058.toString()
  LET ls_js = util.JSON.stringify(la_param)

  CALL cl_cmdrun(ls_js)
   
END FUNCTION 

FUNCTION adzp080_get_dzaf_spec_type(p_program)
DEFINE
  p_program   STRING
DEFINE
  ls_program      STRING,
  ls_spec_type    VARCHAR(5),
  ls_sql          STRING
DEFINE
  ls_return STRING  

  LET ls_program   = p_program
  
  #Begin:200215-00009
  LET ls_sql = "select distinct af.dzaf005           ",
               "  from dzaf_t af                     ", 
               " where af.dzaf001 = '",ls_program,"' "
  #End:200215-00009

  PREPARE lpre_get_dzaf_spec_type FROM ls_sql
  DECLARE lcur_get_dzaf_spec_type CURSOR FOR lpre_get_dzaf_spec_type

  OPEN lcur_get_dzaf_spec_type
  FETCH lcur_get_dzaf_spec_type INTO ls_spec_type
  CLOSE lcur_get_dzaf_spec_type
  
  FREE lpre_get_dzaf_spec_type
  FREE lcur_get_dzaf_spec_type  

  LET ls_return = ls_spec_type
  
  RETURN ls_return
  
END FUNCTION
#180327-00002 end

#241030-00015 begin
FUNCTION adzp080_set_window_title(p_lang)
DEFINE
  p_lang  STRING
DEFINE
  ls_lang         STRING,
  ls_sql          STRING,
  ls_form_title   VARCHAR(256),
  ls_window_title STRING

  LET ls_lang = p_lang

  LET ls_sql = "select gzzd005                 ",
               "  from gzzd_t                  ",
               " where gzzd001 = 'adzp080'     ",
               "   and gzzd002 = '",ls_lang,"' ",
               "   and gzzd003 = 'f_adzp080'   "

  PREPARE lpre_get_form_title FROM ls_sql
  DECLARE lcur_get_form_title CURSOR FOR lpre_get_form_title

  OPEN lcur_get_form_title
  FETCH lcur_get_form_title INTO ls_form_title
  CLOSE lcur_get_form_title
  
  FREE lpre_get_form_title
  FREE lcur_get_form_title  
  
  CALL FGL_SETTITLE(ls_form_title)

END FUNCTION
#241030-00015 end
