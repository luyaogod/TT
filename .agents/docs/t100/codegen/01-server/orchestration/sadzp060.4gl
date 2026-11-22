#&define DEBUG

#2015/11/02              by Hiko : 將FGL_GETENV("MNT4RP")，改成cl_rpt_get_env_global(“MNT4RP”)
#2016/06/07 160607-00012 by Ernest : 在Light平台上傳時不做Report Server的檢核
#2016/12/23 161214-00010 by Ernest : 中文顯示訊息改為英文
# Modify    170111-00006   : 2017/01/11 by Ernest  : 下載時同時賦予壓縮包 everyone 的權限
# Modify    170515-00013   : 2017/05/13 by Ernest  : 新增上傳預編譯型態 C4GL,C4FD
# Modify    170626-00037   : 2017/06/29 by Ernest  : 上傳解壓縮後, 將tsd2/tap2更名成tsd/tap後才開始跑原本更新設計資料的流程
# Modify    180418-00019   : 20180418   by 02160   : 1.Mark掉tap2的相關修改
# Modify    180427-00055   : 20180601   by 02160   : 1.提供一個單純將報表備份檔解縮到報表主機的功能給 adzp063 呼叫
# Modify    180427-00055   : 20180504   by 02160   : 1.4RP 簽出時備份上一個版次的壓縮檔
#           xxxxxx-yyyyy   : 20210517   by 02160   : 檢核欲壓縮的檔案是否可寫入, 若不可寫入, 就有可能在上傳時就已經失敗
# Modify    230109-00005#1 : 20230110   by 02160   : 新增支援 XR 報表
# Modify    230925-00019#3 : 20231030   by 02160   : 新增支援 SPECRC3
# Modify                   : 20240802   by 02160   : 修正 mount 端給予權限的問題
# Modify    240923-00028#1 : 20241001   by 02160   : 1.上傳規格 : tzs解壓縮之後, 請將 bdx 檔放在dzx/tsd 目錄下


IMPORT os
IMPORT util
IMPORT XML

SCHEMA ds

#匯入adzp080的常數設定  
&include "../4gl/sadzp000_cnst.inc" 
&include "../4gl/adzp080_cnst.inc" 
&include "../4gl/sadzp000_type.inc" 
&include "../4gl/adzp080_type.inc"  

CONSTANT cs_default_client_path STRING = "c:\\tt\\"

PRIVATE TYPE T_FRP_LIST RECORD
               fl_NAME     LIKE GZGD_T.GZGD007,
               fl_EXIST    VARCHAR(1), 
               fl_TEMPLATE LIKE GZGD_T.GZGD007
             END RECORD

PUBLIC TYPE T_FILE_DIALOG RECORD
              PATH       VARCHAR(1024),
              TYPE_DESC  VARCHAR(1024),
              TYPE_LIST  VARCHAR(50),
              CAPTION    VARCHAR(1024)
            END RECORD             

DEFINE
  mo_parameters   T_UPLOAD_PARAM,
  mi_progress     INTEGER,
  ms_progress_all STRING,
  ms_lang         STRING

FUNCTION sadzp060_run(p_parameters,p_lang)
DEFINE
  p_parameters  T_UPLOAD_PARAM,
  p_lang        STRING
DEFINE 
  lb_success   BOOLEAN, 
  ls_message   STRING,
  ls_temp_dir  STRING, #170626-00037
  lo_return_params  T_UPLOAD_PARAM

  LET mo_parameters.* = p_parameters.*
  LET ms_lang         = NVL(p_lang,cs_default_lang)
  
  CALL sadzp060_initialize()
  CALL sadzp060_start() RETURNING lb_success,ls_message,ls_temp_dir,lo_return_params.* #170626-00037
  CALL sadzp060_finalize()

  RETURN lb_success,ls_message,ls_temp_dir,lo_return_params.* #170626-00037
  
END FUNCTION

FUNCTION sadzp060_initialize()

  LET mi_progress = 0
  
  CALL ui.Dialog.setDefaultUnbuffered(TRUE)
  
END FUNCTION

FUNCTION sadzp060_start()
DEFINE
  ls_temp_dir         STRING,
  ls_file_name        STRING,
  ls_module_name      STRING,
  ls_ext_name         STRING,
  ls_client_path      STRING,
  ls_arg_val1         STRING,
  ls_arg_val2         STRING,
  ls_arg_val3         STRING,
  ls_arg_val4         STRING,
  ls_upload_doc_type  STRING,
  ls_zip_type         STRING,
  ls_message          STRING,
  ls_progress         STRING,
  ls_version          STRING,
  ls_client_temp_path STRING,
  ls_separator        STRING,
  ls_ERPID            STRING, -- 160607-00012
  li_progress_count   INTEGER,
  lo_parameters       T_UPLOAD_PARAM,
  lo_return_params    T_UPLOAD_PARAM,
  lb_success          BOOLEAN
DEFINE 
  lb_result BOOLEAN,
  lo_result T_UPLOAD_PARAM

  LET lo_parameters.* = mo_parameters.*
  
  &ifndef DEBUG
  LET ls_arg_val1 = lo_parameters.arg1_program_name    #Program Name
  LET ls_arg_val2 = lo_parameters.arg2_module_name     #Module Name
  LET ls_arg_val3 = lo_parameters.arg3_upload_doc_type #Upload Document Type (SPEC or CODE)
  LET ls_arg_val4 = lo_parameters.arg4_client_path     #Client Path
  &else
  LET ls_arg_val1 = "aiti333.tzp"
  LET ls_arg_val2 = "ait"
  LET ls_arg_val3 = "PR"
  LET ls_arg_val4 = cs_default_client_path
  &endif
  
  LET ls_ERPID = NVL(FGL_GETENV("ERPID"),"T100ERP")

  LET lb_success = TRUE
  LET ls_separator = os.Path.separator()
  CALL util.Math.srand()
  
  IF ls_arg_val1.getIndexOf(".",1) > 1 THEN
    LET ls_file_name = ls_arg_val1.subString(1,ls_arg_val1.getIndexOf(".",1)-1)
    LET ls_ext_name  = ls_arg_val1.subString(ls_arg_val1.getIndexOf(".",1)+1,ls_arg_val1.getLength())
  ELSE
    LET ls_file_name = ls_arg_val1

    #確認上傳副檔名及對應形式
    CASE
      WHEN (
            ls_arg_val3 = cs_doc_type_spec OR 
            ls_arg_val3 = cs_doc_type_specrc3 OR #230925-00019#3 
            ls_arg_val3 = cs_doc_type_specgen OR
            ls_arg_val3 = cs_doc_type_c4fd #170515-00013
            )
        LET ls_ext_name  = cs_zip_name_tzs
      WHEN (
            ls_arg_val3 = cs_doc_type_code OR 
            ls_arg_val3 = cs_doc_type_c4gl #170515-00013
           )
        LET ls_ext_name  = cs_zip_name_tzc
      WHEN (ls_arg_val3 = cs_doc_type_cspec)
        LET ls_ext_name  = cs_zip_name_tzd
      WHEN (ls_arg_val3 = cs_doc_type_rspec)
        LET ls_ext_name  = cs_zip_name_tzr
      WHEN (ls_arg_val3 = cs_doc_type_gcode)
        LET ls_ext_name  = cs_zip_name_tzg
      WHEN (ls_arg_val3 = cs_doc_type_4rp)
        LET ls_ext_name  = cs_zip_name_tzt
      #180427-00055 begin
      WHEN (ls_arg_val3 = cs_doc_type_4rpupk)
        LET ls_ext_name  = cs_zip_name_tzt
      #180427-00055 end
      #230109-00005#1 begin
      WHEN (ls_arg_val3 = cs_doc_type_xrp)
        LET ls_ext_name  = cs_zip_name_tzt
      WHEN (ls_arg_val3 = cs_doc_type_xrpupk)
        LET ls_ext_name  = cs_zip_name_tzt
      #230109-00005#1 end
    OTHERWISE  
      LET ls_ext_name  = cs_zip_name_tzp
    END CASE
  END IF

  LET ls_zip_type = ls_ext_name.trim()
  LET ls_upload_doc_type = ls_arg_val3.trim()

  #取得Progress的總筆數
  CALL sadzp060_get_sadzp060_progress_all(ls_zip_type) RETURNING ms_progress_all

  {
  #180427-00055  mark begin
  IF NVL(ls_arg_val4,"NULL") = "NULL" THEN
    DISPLAY "************************************************************"
    DISPLAY "[Warning] Client path is not setting, using default : ",cs_default_client_path
    DISPLAY "************************************************************"
    LET ls_client_path = cs_default_client_path
  ELSE
    LET ls_client_path = ls_arg_val4.trim() #"C:\\TT\\"
  END IF
  #180427-00055 mark end
  }  
  LET ls_client_path = ls_arg_val4.trim()
  
  #取得模組名稱
  IF ls_arg_val2 IS NULL THEN
    LET ls_module_name = ls_file_name.subString(1,3)
    LET ls_module_name = ls_module_name.toUpperCase()
  ELSE  
    LET ls_module_name = ls_arg_val2.trim()
  END IF  

  CALL sadzp060_creat_temp_directory(ls_file_name.trim()) RETURNING ls_temp_dir
  
  DISPLAY "[TEMPORARY DIR] ",ls_temp_dir

  #處理報表區段
  #180427-00055 begin
  CASE 
    WHEN (ls_upload_doc_type = cs_doc_type_4rpupk) OR 
         (ls_upload_doc_type = cs_doc_type_xrpupk) #230109-00005#1
      CALL sadzp060_check_zip_file_if_exists(ls_client_path) RETURNING lb_success
      IF NOT lb_success THEN LET ls_message = cs_error_tag,"The 4RP/XRP package file '",ls_client_path," not exist !" DISPLAY ls_message EXIT CASE END IF
    
      CALL sadzp060_move_4rp_xrp_zip_to_worktemp(os.Path.basename(ls_client_path),os.Path.dirName(ls_client_path),ls_temp_dir) RETURNING lb_success,ls_file_name 
      IF NOT lb_success THEN LET ls_message = cs_error_tag,"Move source file '",ls_client_path,"' to temp dir '",ls_temp_dir,"' failed !!" DISPLAY ls_message EXIT CASE END IF
      
      CALL sadzp060_unzip_to_temp_dir(ls_file_name,ls_temp_dir||ls_separator,ls_zip_type,ls_upload_doc_type) RETURNING lb_success
      IF NOT lb_success THEN LET ls_message = cs_error_tag,"Unzip verify file failed !!" DISPLAY ls_message EXIT CASE END IF
      
      IF ls_ERPID = "T100ERP" THEN
        LET ls_message = "Checking report server."
        DISPLAY cs_message_tag,ls_message
        CALL sadzp060_check_report_server_valid(ls_module_name,ms_lang,FALSE,ls_upload_doc_type) RETURNING lb_success 
        IF NOT lb_success THEN LET ls_message = cs_error_tag,"Checking report server failed !!" DISPLAY ls_message EXIT CASE END IF
      ELSE
        LET lb_success = TRUE
        DISPLAY cs_warning_tag,"Not T100ERP skip checking report server."   
      END IF  
      
      LET ls_message = "Uncompressing and moving files to destination folder."  
      DISPLAY cs_message_tag,ls_message
      CALL sadzp060_unzip_4rp_xrp_file(ls_module_name,ls_file_name,ls_temp_dir||ls_separator,ls_zip_type,ms_lang,ls_upload_doc_type) RETURNING lb_success #230109-00005#1 
      IF NOT lb_success THEN LET ls_message = cs_error_tag,"Unpacking and moving files to destination folder failed !!" DISPLAY ls_message EXIT CASE END IF
      
      {
      LET ls_message = "Generating report multi language datas."  
      DISPLAY cs_message_tag,ls_message
      CALL sadzp060_4rp_xrp_gen_multi_lang(ls_module_name,ls_file_name,ls_upload_doc_type) RETURNING lb_success 
      IF NOT lb_success THEN LET ls_message = cs_error_tag,"Generating report multi language datas failed !!" DISPLAY ls_message EXIT CASE END IF
      }
      
    WHEN (ls_upload_doc_type = cs_doc_type_4rp) OR
         (ls_upload_doc_type = cs_doc_type_xrp) #230109-00005#1
  #180427-00055 end

      LET li_progress_count = 7
      #開啟Progressbar
      CALL cl_progress_bar(li_progress_count)
      #CLOSE WINDOW SCREEN

      LET ls_message = "Uploading Report template ",ls_zip_type," file." -- 161214-00010
      CALL cl_progress_ing(ls_message)
      --CALL sadzp060_set_progress(ls_message)
      
      LET ls_message = "Compressing client report files." -- 161214-00010 
      CALL cl_progress_ing(ls_message)
      --CALL sadzp060_set_progress(ls_message)
      CALL sadzp060_compress_client_files(ls_client_path,ls_upload_doc_type) RETURNING lb_success,ls_file_name
      IF NOT lb_success THEN DISPLAY cs_error_tag,"Compress client report files failed !!" GOTO _report_error END IF  -- 161214-00010 

      LET ls_message = "Uploading report compressed pacakge." -- 161214-00010  
      CALL cl_progress_ing(ls_message)
      --CALL sadzp060_set_progress(ls_message)
      LET ls_client_temp_path = cs_client_temp_dir,cs_dos_separator
      CALL sadzp060_upload_zip(ls_file_name,ls_client_temp_path,ls_temp_dir||ls_separator,ls_zip_type) RETURNING lb_success,ls_file_name
      IF NOT lb_success THEN DISPLAY cs_error_tag,"Upload report compressed package failed !!" GOTO _report_error END IF -- 161214-00010  

      LET ls_message = "Checking contents if valid." -- 161214-00010  
      CALL cl_progress_ing(ls_message)
      --CALL sadzp060_set_progress(ls_message)
      CALL sadzp060_check_zip_content_valid(lo_parameters.*,ls_file_name,ls_temp_dir||ls_separator,ls_zip_type) RETURNING lb_success, ls_version, lo_return_params.*
      IF NOT lb_success THEN DISPLAY cs_error_tag,"Checking contents failed !!" GOTO _report_error END IF -- 161214-00010 
      #重新給定由 vfy 中取得的模組名稱
      LET ls_module_name = lo_return_params.arg2_module_name
      --LET ls_module_name = "AIN" #For debug
      
      -- 160607-00012 begin
      IF ls_ERPID = "T100ERP" THEN
        LET ls_message = "Checking report server." -- 161214-00010
        CALL cl_progress_ing(ls_message)
        --CALL sadzp060_set_progress(ls_message)
        CALL sadzp060_check_report_server_valid(ls_module_name,ms_lang,TRUE,ls_upload_doc_type) RETURNING lb_success 
        IF NOT lb_success THEN DISPLAY cs_error_tag,"Checking report server failed !!" GOTO _report_error END IF -- 161214-00010  
      ELSE
        LET lb_success = TRUE
        DISPLAY cs_warning_tag,"Not T100ERP skip checking report server." -- 161214-00010   
      END IF  
      -- 160607-00012 end
      
      LET ls_message = "Uncompressing and moving files to destination folder." -- 161214-00010  
      CALL cl_progress_ing(ls_message)
      --CALL sadzp060_set_progress(ls_message)
      CALL sadzp060_unzip_4rp_xrp_file(ls_module_name,ls_file_name,ls_temp_dir||ls_separator,ls_zip_type,ms_lang,ls_upload_doc_type) RETURNING lb_success #230109-00005#1  
      IF NOT lb_success THEN DISPLAY cs_error_tag,"Unpacking and moving files to destination folder failed !!" GOTO _report_error END IF -- 161214-00010 

      LET ls_message = "Generating report multi language datas." -- 161214-00010  
      CALL cl_progress_ing(ls_message)
      --CALL sadzp060_set_progress(ls_message)
      CALL sadzp060_4rp_xrp_gen_multi_lang(ls_module_name,ls_file_name,ls_upload_doc_type) RETURNING lb_success 
      IF NOT lb_success THEN DISPLAY cs_error_tag,"Generating report multi language datas failed !!" GOTO _report_error END IF -- 161214-00010 
      
      #LET ls_message = "Uploading report template finished with file ",ls_zip_type," ." -- 161214-00010
      #CALL cl_progress_ing(ls_message)
      
      LABEL _report_error:

      IF NOT lb_success THEN 
        TRY 
          #關閉Progressbar
          CALL cl_progress_bar_close()
        CATCH
          DISPLAY cs_warning_tag,"Call for closing progress bar failed !!" -- 161214-00010 
        END TRY  
        GOTO _return
      END IF  
    
  --ELSE #180427-00055 mark
  OTHERWISE #180427-00055 add  

    CALL sadzp060_upload_zip(ls_file_name,ls_client_path,ls_temp_dir||ls_separator,ls_zip_type) RETURNING lb_success,ls_file_name

    CALL sadzp060_show_message_designer()
    
    IF (ls_upload_doc_type <> cs_doc_type_c4fd) AND (ls_upload_doc_type <> cs_doc_type_c4gl) THEN #170515-00013 add
      #檢查權限是否足夠, 並且取得版號, 用Display 的方式讓adzp080取得以傳遞給adzp040
      LET ls_message = "Check ",ls_zip_type," if valid."
      CALL sadzp060_set_progress(ls_message)
      CALL sadzp060_check_zip_content_valid(lo_parameters.*,ls_file_name,ls_temp_dir||ls_separator,ls_zip_type) RETURNING lb_success, ls_message, lo_return_params.*
    #170515-00013 begin
    ELSE 
      LET lb_success = TRUE
    END IF 
    #170515-00013 end
    
    #DISPLAY cs_version_tag,ls_version

    IF lb_success THEN
      LET ls_message = "Uncompress file : ",ls_file_name
      DISPLAY "[",CURRENT,"] ",ls_message
      CALL sadzp060_set_progress(ls_message)

      #SPEC or SPECGEN
      IF (
          ls_upload_doc_type = cs_doc_type_spec OR 
          ls_upload_doc_type = cs_doc_type_specrc3 OR  #230925-00019#3 
          ls_upload_doc_type = cs_doc_type_specgen
         )THEN
        CALL sadzp060_unzip_form_file(ls_module_name,ls_file_name,ls_temp_dir||ls_separator,ls_zip_type) RETURNING lb_success
      END IF 

      #CODE or GCODE
      IF (ls_upload_doc_type = cs_doc_type_code OR ls_upload_doc_type = cs_doc_type_gcode) THEN
        CALL sadzp060_unzip_code_file(ls_module_name,ls_file_name,ls_temp_dir||ls_separator,ls_zip_type) RETURNING lb_success
      END IF 

      #CSPEC
      IF ls_upload_doc_type = cs_doc_type_cspec THEN
        CALL sadzp060_unzip_csd_file(ls_module_name,ls_file_name,ls_temp_dir||ls_separator,ls_zip_type) RETURNING lb_success
      END IF

      #RSPEC
      IF ls_upload_doc_type = cs_doc_type_rspec THEN
        CALL sadzp060_unzip_rsd_file(ls_module_name,ls_file_name,ls_temp_dir||ls_separator,ls_zip_type) RETURNING lb_success
      END IF

      #170515-00013 begin
      #C4FD
      IF (ls_upload_doc_type = cs_doc_type_c4fd)THEN
        CALL sadzp060_unzip_c4fd_file(ls_module_name,ls_file_name,ls_temp_dir||ls_separator,ls_zip_type) RETURNING lb_success
      END IF 
      
      #C4GL
      IF (ls_upload_doc_type = cs_doc_type_c4gl)THEN
        CALL sadzp060_unzip_c4gl_file(ls_module_name,ls_file_name,ls_temp_dir||ls_separator,ls_zip_type) RETURNING lb_success
      END IF
      #170515-00013 end
      
      LET ls_progress = mi_progress
      IF ls_progress <> ms_progress_all THEN
        #若原始記錄的總Progress項目和現行的不符, 則要重新更新
        CALL sadzp060_set_sadzp060_progress_all(ls_zip_type,mi_progress)
      END IF
    END IF
  --END IF #180427-00055 mark
  END CASE #180427-00055 add  
  
  LABEL _return:
  
  LET lb_result = lb_success
  LET lo_result.* = lo_return_params.*

  RETURN lb_success,ls_message,ls_temp_dir,lo_result.* #170626-00037 add temp_dir
  
END FUNCTION

FUNCTION sadzp060_finalize()
END FUNCTION

FUNCTION sadzp060_upload_zip(p_file_name,p_source_path,p_destination_path,p_zip_type)
DEFINE
  p_file_name        STRING,
  p_source_path      STRING,
  p_destination_path STRING,
  p_zip_type         STRING 
DEFINE
  ls_file_name        STRING,
  ls_source_path      STRING,
  ls_destination_path STRING,
  ls_zip_type         STRING,
  ls_source_file      STRING,
  ls_destination_file STRING,
  ls_src_file_name    STRING,
  ls_dst_file_name    STRING,
  lb_success          BOOLEAN,
  lb_result           BOOLEAN #170111-00006
DEFINE 
  lb_return BOOLEAN  

  LET ls_file_name        = p_file_name
  LET ls_source_path      = p_source_path        
  LET ls_destination_path = p_destination_path
  LET ls_zip_type         = p_zip_type

  LET lb_success = TRUE 

  #當上傳為 tzt(報表樣板)檔時, 來源名稱會加上 "_up"
  IF ls_zip_type = cs_zip_name_tzt THEN
    LET ls_src_file_name = ls_file_name,cs_upload_string
  ELSE   
    LET ls_src_file_name = ls_file_name
  END IF

  #上傳到暫存檔的時候一律改回原名稱的壓縮檔
  IF ls_file_name.getIndexOf("(",1) > 0 THEN
    LET ls_dst_file_name = ls_file_name.subString(1,ls_file_name.getIndexOf(cs_open_paren,1)-1)
  ELSE
    LET ls_dst_file_name = ls_file_name
  END IF   
  
  LET ls_source_file      = ls_source_path,ls_src_file_name,".",ls_zip_type
  LET ls_destination_file = ls_destination_path,ls_dst_file_name,".",ls_zip_type

  --DISPLAY ls_source_file
  --DISPLAY ls_destination_file
  
  TRY
    CALL sadzp060_grant_clinet_permission(ls_source_file) RETURNING lb_result #170111-00006
    CALL FGL_GETFILE(ls_source_file,ls_destination_file)
    DISPLAY cs_message_tag,"Upload ",ls_zip_type.toUpperCase()," file "||ls_file_name||" success ! "
  CATCH
    LET lb_success = FALSE 
    DISPLAY cs_error_tag,"Upload ",ls_zip_type.toUpperCase()," "||ls_file_name||" File Error ! "
  END TRY  

  LET lb_return = lb_success

  RETURN lb_return,ls_dst_file_name

END FUNCTION

FUNCTION sadzp060_unzip_to_temp_dir(p_file_name, p_temp_dir, p_zip_type,p_upload_doc_type)
DEFINE
  p_file_name   STRING,
  p_temp_dir    STRING,
  p_zip_type    STRING,
  p_upload_doc_type STRING
DEFINE
  ls_file_name     STRING,
  ls_temp_dir      STRING,
  ls_zip_type      STRING,
  ls_upload_doc_type STRING,
  ls_zip_full_name STRING,
  ls_command       STRING,
  ls_err_command   STRING,
  lb_success       BOOLEAN,
  lb_complete      BOOLEAN,
  ls_step          STRING
DEFINE 
  lb_return  BOOLEAN  
  
  LET ls_file_name   = p_file_name
  LET ls_temp_dir    = p_temp_dir
  LET ls_zip_type    = p_zip_type
  LET ls_upload_doc_type = p_upload_doc_type

  LET lb_success  = TRUE
  LET lb_complete = TRUE
  LET lb_return   = TRUE
  LET ls_err_command = ""

  LET ls_zip_full_name = ls_temp_dir,ls_file_name
  CALL os.Path.chdir(ls_temp_dir) RETURNING lb_success

  CASE 
    WHEN (
          ls_upload_doc_type = cs_doc_type_spec OR 
          ls_upload_doc_type = cs_doc_type_specrc3 OR  #230925-00019#3  
          ls_upload_doc_type = cs_doc_type_specgen
         )
      LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".tsd -d '",ls_temp_dir,"'"
      CALL sadzp060_run_command(ls_command) RETURNING lb_success
      IF NOT lb_success THEN 
        LET lb_complete = FALSE
        LET ls_err_command = ls_command
        LET ls_step = "Unzip program file : STEP 1" 
      END IF
    WHEN (ls_upload_doc_type = cs_doc_type_code OR ls_upload_doc_type = cs_doc_type_gcode)
      LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".tap -d '",ls_temp_dir,"'"
      CALL sadzp060_run_command(ls_command) RETURNING lb_success
      IF NOT lb_success THEN 
        LET lb_complete = FALSE
        LET ls_err_command = ls_command
        LET ls_step = "Unzip program file : STEP 1" 
      END IF
    WHEN (ls_upload_doc_type = cs_doc_type_cspec)
      LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".csd -d '",ls_temp_dir,"'"
      CALL sadzp060_run_command(ls_command) RETURNING lb_success
      IF NOT lb_success THEN 
        LET lb_complete = FALSE
        LET ls_err_command = ls_command
        LET ls_step = "Unzip program file : STEP 1" 
      END IF
    WHEN (ls_upload_doc_type = cs_doc_type_rspec)
      LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".rsd -d '",ls_temp_dir,"'"
      CALL sadzp060_run_command(ls_command) RETURNING lb_success
      IF NOT lb_success THEN 
        LET lb_complete = FALSE
        LET ls_err_command = ls_command
        LET ls_step = "Unzip program file : STEP 1" 
      END IF
    WHEN (ls_upload_doc_type = cs_doc_type_4rp) OR (ls_upload_doc_type = cs_doc_type_4rpupk) #180427-00055
      LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".vfy -d '",ls_temp_dir,"'"
      --LET ls_command = "tar -xvf ",ls_zip_full_name,".",ls_zip_type --," ",ls_temp_dir
      CALL sadzp060_run_command(ls_command) RETURNING lb_success
      IF NOT lb_success THEN 
        LET lb_complete = FALSE
        LET ls_err_command = ls_command
        LET ls_step = "Unzip program file : STEP 1" 
      END IF
    #230109-00005#1 begin  
    WHEN (ls_upload_doc_type = cs_doc_type_xrp) OR (ls_upload_doc_type = cs_doc_type_xrpupk)
      LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".vfy -d '",ls_temp_dir,"'"
      CALL sadzp060_run_command(ls_command) RETURNING lb_success
      IF NOT lb_success THEN 
        LET lb_complete = FALSE
        LET ls_err_command = ls_command
        LET ls_step = "Unzip program file : STEP 1" 
      END IF
    #230109-00005#1 end
  OTHERWISE  
  END CASE   
  
  SLEEP 1

  IF NOT lb_complete THEN
    LET lb_return = FALSE
    DISPLAY "[WARNING] ",ls_step
    DISPLAY "[Command] ",ls_err_command
  ELSE
    LET lb_return = TRUE
  END IF

  #一律設為True
  LET lb_return = TRUE
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzp060_unzip_program_file(p_module_name, p_file_name, p_temp_dir, p_zip_type)
DEFINE
  p_module_name STRING, 
  p_file_name   STRING,
  p_temp_dir    STRING,
  p_zip_type    STRING
DEFINE
  ls_module_name   STRING, 
  ls_file_name     STRING,
  ls_temp_dir      STRING,
  ls_zip_type      STRING,
  ls_module_path   STRING,
  ls_zip_full_name STRING,
  ls_command       STRING,
  ls_err_command   STRING,
  lb_success       BOOLEAN,
  lb_complete      BOOLEAN,
  ls_step          STRING,
  ls_separator     STRING
DEFINE 
  lb_return  BOOLEAN  
  
  LET ls_module_name = p_module_name.toUpperCase()
  LET ls_file_name   = p_file_name
  LET ls_temp_dir    = p_temp_dir
  LET ls_zip_type    = p_zip_type

  LET lb_success  = TRUE
  LET lb_complete = TRUE
  LET lb_return   = TRUE
  LET ls_err_command = ""

  LET ls_separator = os.Path.separator()
  LET ls_zip_full_name = ls_temp_dir,ls_file_name
  
  LET ls_module_path = FGL_GETENV(ls_module_name)

  #備份
  LET ls_command = "cp ",ls_module_path,ls_separator,"4fd",ls_separator,ls_file_name,".4fd ",ls_module_path,ls_separator,"4fd",ls_separator,ls_file_name,".4fd.bak" 
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 1" 
  END IF

  LET ls_command = "cp ",ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".tsd ",ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".tsd.bak"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 2" 
  END IF

  #240923-00028#1 begin
  LET ls_command = "cp ",ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".bdx ",ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".bdx.bak"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 2" 
  END IF
  #240923-00028#1 end

  LET ls_command = "cp ",ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".csd ",ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".csd.bak"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 2" 
  END IF
  
  LET ls_command = "cp ",ls_module_path,ls_separator,"dzx",ls_separator,"tap",ls_separator,ls_file_name,".tap ",ls_module_path,ls_separator,"dzx",ls_separator,"tap",ls_separator,ls_file_name,".tap.bak" 
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 3" 
  END IF
  
  LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".4fd -d '",ls_module_path,ls_separator,"4fd","'"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 4" 
  END IF

  LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".tsd -d '",ls_module_path,ls_separator,"dzx",ls_separator,"tsd","'"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 5" 
  END IF

  #240923-00028#1 begin
  LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".bdx -d '",ls_module_path,ls_separator,"dzx",ls_separator,"tsd","'"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 5" 
  END IF
  #240923-00028#1 end

  LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".csd -d '",ls_module_path,ls_separator,"dzx",ls_separator,"tsd","'"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 5" 
  END IF

  LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".tap -d '",ls_module_path,ls_separator,"dzx",ls_separator,"tap","'"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 6" 
  END IF
  
  SLEEP 1

  IF NOT lb_complete THEN
    LET lb_return = FALSE
    DISPLAY "[WARNING] ",ls_step
    DISPLAY "[Command] ",ls_err_command
  ELSE
    LET lb_return = TRUE
  END IF

  #一律設為True
  LET lb_return = TRUE
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzp060_unzip_code_file(p_module_name, p_file_name, p_temp_dir, p_zip_type)
DEFINE
  p_module_name STRING, 
  p_file_name   STRING,
  p_temp_dir    STRING,
  p_zip_type    STRING
DEFINE
  ls_module_name   STRING, 
  ls_file_name     STRING,
  ls_temp_dir      STRING,
  ls_temp_root     STRING, #170626-00037
  ls_src_file      STRING, #170626-00037
  ls_dst_file      STRING, #170626-00037
  ls_zip_type      STRING,
  ls_module_path   STRING,
  ls_zip_full_name STRING,
  ls_command       STRING,
  ls_err_command   STRING,
  lb_success       BOOLEAN,
  lb_complete      BOOLEAN,
  lb_result        BOOLEAN, #170626-00037
  ls_separator     STRING,
  ls_step          STRING
DEFINE 
  lb_return  BOOLEAN  
  
  LET ls_module_name = p_module_name.toUpperCase()
  LET ls_file_name   = p_file_name
  LET ls_temp_dir    = p_temp_dir
  LET ls_zip_type    = p_zip_type

  LET lb_success  = TRUE
  LET lb_complete = TRUE
  LET lb_return   = TRUE
  LET ls_err_command = ""

  LET ls_separator = os.Path.separator()
  LET ls_zip_full_name = ls_temp_dir,ls_file_name

  #170626-00037 begin
  IF ls_temp_dir.subString(ls_temp_dir.getLength(),ls_temp_dir.getLength()) = ls_separator THEN 
    LET ls_temp_root = ls_temp_dir.subString(1,ls_temp_dir.getLength()-1)
  ELSE  
    LET ls_temp_root = ls_temp_dir
  END IF   
  #170626-00037 end
  
  LET ls_module_path = FGL_GETENV(ls_module_name)

  #備份
  LET ls_command = "cp ",ls_module_path,ls_separator,"dzx",ls_separator,"tap",ls_separator,ls_file_name,".tap ",ls_module_path,ls_separator,"dzx",ls_separator,"tap",ls_separator,ls_file_name,".tap.bak" 
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 1" 
  END IF
  
  CALL sadzp060_run_command(ls_command) RETURNING lb_success

  #170626-00037 bagin
  #LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".tap -d '",ls_module_path,ls_separator,"dzx",ls_separator,"tap","'" #170626-00037 mark
  LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".tap -d '",ls_temp_root,"'"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  --LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".tap2 -d '",ls_temp_root,"'" -- 180418-00019  mark
  --CALL sadzp060_run_command(ls_command) RETURNING lb_success -- 180418-00019 mark

  --LET ls_src_file = ls_temp_dir,ls_file_name,".tap2" -- 180418-00019 mark
  LET ls_src_file = ls_temp_dir,ls_file_name,".tap" -- 180418-00019 add
  LET ls_dst_file = ls_module_path,ls_separator,"dzx",ls_separator,"tap",ls_separator,ls_file_name,".tap"

  {
  -- 180418-00019 mark begin
  IF NOT os.Path.exists(ls_src_file) OR NOT (sadzp060_check_design_file_if_have_contents(ls_src_file)) THEN
    DISPLAY cs_warning_tag,"Not exist tap2 or contents is blank, clone from tap."
    LET ls_src_file = ls_temp_dir,ls_file_name,".tap"
  END IF
  -- 180418-00019 mark end  }
  
  CALL os.Path.copy(ls_src_file,ls_dst_file) RETURNING lb_success
  CALL os.Path.chrwx(ls_dst_file,511) RETURNING lb_result
  #170626-00037 end
  
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 2" 
  END IF
  
  SLEEP 1

  IF NOT lb_complete THEN
    LET lb_return = FALSE
    DISPLAY "[WARNING] ",ls_step
    DISPLAY "[Command] ",ls_err_command
  ELSE
    LET lb_return = TRUE
  END IF

  #一律設為True
  LET lb_return = TRUE
  
  RETURN lb_return
  
END FUNCTION

#170626-00037 begin
FUNCTION sadzp060_move_tap_file(p_module_name, p_file_name, p_temp_dir)
DEFINE
  p_module_name STRING, 
  p_file_name   STRING,
  p_temp_dir    STRING
DEFINE
  ls_module_name   STRING, 
  ls_file_name     STRING,
  ls_temp_dir      STRING,
  ls_temp_root     STRING, 
  ls_src_file      STRING, 
  ls_dst_file      STRING, 
  ls_module_path   STRING,
  lb_result        BOOLEAN, 
  ls_separator     STRING
DEFINE 
  lb_return  BOOLEAN  
  
  LET ls_module_name = p_module_name.toUpperCase()
  LET ls_file_name   = p_file_name
  LET ls_temp_dir    = p_temp_dir

  LET lb_return   = TRUE
  
  LET ls_separator = os.Path.separator()
  LET ls_module_path = FGL_GETENV(ls_module_name)

  IF ls_temp_dir.subString(ls_temp_dir.getLength(),ls_temp_dir.getLength()) = ls_separator THEN 
    LET ls_temp_root = ls_temp_dir.subString(1,ls_temp_dir.getLength()-1)
  ELSE  
    LET ls_temp_root = ls_temp_dir
  END IF   

  {
  #180418-00019 mark begin
  LET ls_src_file = ls_temp_dir,ls_file_name,".tap2"
  LET ls_dst_file = ls_module_path,ls_separator,"dzx",ls_separator,"tap",ls_separator,ls_file_name,".tap2"
  IF os.Path.exists(ls_src_file) THEN
    CALL os.Path.copy(ls_src_file,ls_dst_file) RETURNING lb_result
    CALL os.Path.chrwx(ls_dst_file,511) RETURNING lb_result
  ELSE
    DISPLAY cs_warning_tag,"Not exists file : ",ls_file_name,".tap2"," skip clone file"  
  END IF
  #180418-00019 mark end
  }
  
  LET ls_src_file = ls_temp_dir,ls_file_name,".tap"
  LET ls_dst_file = ls_module_path,ls_separator,"dzx",ls_separator,"tap",ls_separator,ls_file_name,".tap"
  IF os.Path.exists(ls_src_file) THEN
    CALL os.Path.copy(ls_src_file,ls_dst_file) RETURNING lb_result
    CALL os.Path.chrwx(ls_dst_file,511) RETURNING lb_result
  ELSE
    DISPLAY cs_warning_tag,"Not exists file : ",ls_file_name,".tap"," skip clone file"  
  END IF
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzp060_move_tsd_file(p_module_name, p_file_name, p_temp_dir)
DEFINE
  p_module_name STRING, 
  p_file_name   STRING,
  p_temp_dir    STRING
DEFINE
  ls_module_name   STRING, 
  ls_file_name     STRING,
  ls_temp_dir      STRING,
  ls_temp_root     STRING, 
  ls_src_file      STRING, 
  ls_dst_file      STRING, 
  ls_module_path   STRING,
  lb_result        BOOLEAN, 
  ls_separator     STRING
DEFINE 
  lb_return  BOOLEAN  
  
  LET ls_module_name = p_module_name.toUpperCase()
  LET ls_file_name   = p_file_name
  LET ls_temp_dir    = p_temp_dir

  LET lb_return   = TRUE
  
  LET ls_separator = os.Path.separator()
  LET ls_module_path = FGL_GETENV(ls_module_name)

  IF ls_temp_dir.subString(ls_temp_dir.getLength(),ls_temp_dir.getLength()) = ls_separator THEN 
    LET ls_temp_root = ls_temp_dir.subString(1,ls_temp_dir.getLength()-1)
  ELSE  
    LET ls_temp_root = ls_temp_dir
  END IF   
  
  {
  #180418-00019 mark begin
  LET ls_src_file = ls_temp_dir,ls_file_name,".tsd2"
  LET ls_dst_file = ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".tsd2"
  IF os.Path.exists(ls_src_file) THEN
    CALL os.Path.copy(ls_src_file,ls_dst_file) RETURNING lb_result
    CALL os.Path.chrwx(ls_dst_file,511) RETURNING lb_result
  ELSE
    DISPLAY cs_warning_tag,"Not exists file : ",ls_file_name,".tsd2"," skip clone file"  
  END IF
  #180418-00019 mark end
  }
  
  LET ls_src_file = ls_temp_dir,ls_file_name,".tsd"
  LET ls_dst_file = ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".tsd"
  IF os.Path.exists(ls_src_file) THEN
    CALL os.Path.copy(ls_src_file,ls_dst_file) RETURNING lb_result
    CALL os.Path.chrwx(ls_dst_file,511) RETURNING lb_result
  ELSE
    DISPLAY cs_warning_tag,"Not exists file : ",ls_file_name,".tsd"," skip clone file"  
  END IF

  #240923-00028#1 begin
  LET ls_src_file = ls_temp_dir,ls_file_name,".bdx"
  LET ls_dst_file = ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".bdx"
  IF os.Path.exists(ls_src_file) THEN
    CALL os.Path.copy(ls_src_file,ls_dst_file) RETURNING lb_result
    CALL os.Path.chrwx(ls_dst_file,511) RETURNING lb_result
  ELSE
    DISPLAY cs_warning_tag,"Not exists file : ",ls_file_name,".bdx"," skip clone file"  
  END IF
  #240923-00028#1 end
  
  RETURN lb_return
  
END FUNCTION
#170626-00037 end

FUNCTION sadzp060_unzip_form_file(p_module_name, p_file_name, p_temp_dir, p_zip_type)
DEFINE
  p_module_name STRING, 
  p_file_name   STRING,
  p_temp_dir    STRING,
  p_zip_type    STRING
DEFINE
  ls_module_name   STRING, 
  ls_file_name     STRING,
  ls_temp_dir      STRING,
  ls_temp_root     STRING, #170626-00037
  ls_src_file      STRING, #170626-00037
  ls_dst_file      STRING, #170626-00037
  ls_zip_type      STRING,
  ls_module_path   STRING,
  ls_zip_full_name STRING,
  ls_command       STRING,
  ls_err_command   STRING,
  lb_success       BOOLEAN,
  lb_complete      BOOLEAN,
  lb_result        BOOLEAN, #170626-00037
  ls_step          STRING,
  ls_separator     STRING
DEFINE 
  lb_return  BOOLEAN  
  
  LET ls_module_name = p_module_name.toUpperCase()
  LET ls_file_name   = p_file_name
  LET ls_temp_dir    = p_temp_dir
  LET ls_zip_type    = p_zip_type

  LET ls_separator = os.Path.separator()
  LET lb_success  = TRUE
  LET lb_complete = TRUE
  LET lb_return   = TRUE
  LET ls_err_command = ""

  LET ls_zip_full_name = ls_temp_dir,ls_file_name

  #170626-00037 begin
  IF ls_temp_dir.subString(ls_temp_dir.getLength(),ls_temp_dir.getLength()) = ls_separator THEN 
    LET ls_temp_root = ls_temp_dir.subString(1,ls_temp_dir.getLength()-1)
  ELSE  
    LET ls_temp_root = ls_temp_dir
  END IF   
  #170626-00037 end
  
  LET ls_module_path = FGL_GETENV(ls_module_name)

  #備份
  LET ls_command = "cp ",ls_module_path,ls_separator,"4fd",ls_separator,ls_file_name,".4fd ",ls_module_path,ls_separator,"4fd",ls_separator,ls_file_name,".4fd.bak" 
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 1" 
  END IF

  LET ls_command = "cp ",ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".tsd ",ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".tsd.bak"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 2" 
  END IF

  #240923-00028#1 begin
  LET ls_command = "cp ",ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".bdx ",ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".bdx.bak"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 2" 
  END IF
  #240923-00028#1 end

  LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".4fd -d '",ls_module_path,ls_separator,"4fd","'"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 3" 
  END IF

  #170626-00037 bagin
  #LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".tsd -d '",ls_module_path,ls_separator,"dzx",ls_separator,"tsd","'" #170626-00037 mark
  LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".tsd -d '",ls_temp_root,"'"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  --LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".tsd2 -d '",ls_temp_root,"'" #180418-00019 mark
  --CALL sadzp060_run_command(ls_command) RETURNING lb_success #180418-00019 mark

  --LET ls_src_file = ls_temp_dir,ls_file_name,".tsd2" #180418-00019 mark
  LET ls_src_file = ls_temp_dir,ls_file_name,".tsd"
  LET ls_dst_file = ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".tsd"

  {
  #180418-00019 mark begin 
  IF NOT os.Path.exists(ls_src_file) OR NOT (sadzp060_check_design_file_if_have_contents(ls_src_file)) THEN
    DISPLAY cs_warning_tag,"Not exist tsd2 or contents is blank, clone from tsd."
    LET ls_src_file = ls_temp_dir,ls_file_name,".tsd"
  END IF
  #180418-00019 mark end 
  }
  
  CALL os.Path.copy(ls_src_file,ls_dst_file) RETURNING lb_success
  CALL os.Path.chrwx(ls_dst_file,511) RETURNING lb_result
  #170626-00037 end
  
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 4" 
  END IF

  #240923-00028#1 bagin
  LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".bdx -d '",ls_temp_root,"'"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success

  LET ls_src_file = ls_temp_dir,ls_file_name,".bdx"
  LET ls_dst_file = ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".bdx"

  CALL os.Path.copy(ls_src_file,ls_dst_file) RETURNING lb_success
  CALL os.Path.chrwx(ls_dst_file,511) RETURNING lb_result
  #240923-00028#1 end

  SLEEP 1

  IF NOT lb_complete THEN
    LET lb_return = FALSE
    DISPLAY cs_warning_tag,ls_step
    DISPLAY cs_command_tag,ls_err_command
  ELSE
    LET lb_return = TRUE
  END IF

  #一律設為True
  LET lb_return = TRUE
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzp060_unzip_csd_file(p_module_name, p_file_name, p_temp_dir, p_zip_type)
DEFINE
  p_module_name STRING, 
  p_file_name   STRING,
  p_temp_dir    STRING,
  p_zip_type    STRING
DEFINE
  ls_module_name   STRING, 
  ls_file_name     STRING,
  ls_temp_dir      STRING,
  ls_zip_type      STRING,
  ls_module_path   STRING,
  ls_zip_full_name STRING,
  ls_command       STRING,
  ls_err_command   STRING,
  lb_success       BOOLEAN,
  lb_complete      BOOLEAN,
  ls_separator     STRING,
  ls_step          STRING
DEFINE 
  lb_return  BOOLEAN  
  
  LET ls_module_name = p_module_name.toUpperCase()
  LET ls_file_name   = p_file_name
  LET ls_temp_dir    = p_temp_dir
  LET ls_zip_type    = p_zip_type

  LET ls_separator = os.Path.separator()
  LET lb_success  = TRUE
  LET lb_complete = TRUE
  LET lb_return   = TRUE
  LET ls_err_command = ""

  LET ls_zip_full_name = ls_temp_dir,ls_file_name
  
  LET ls_module_path = FGL_GETENV(ls_module_name)

  #備份
  LET ls_command = "cp ",ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".csd ",ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".csd.bak"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 1" 
  END IF
  
  LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".csd -d '",ls_module_path,ls_separator,"dzx",ls_separator,"tsd","'"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 2" 
  END IF
  
  SLEEP 1

  IF NOT lb_complete THEN
    LET lb_return = FALSE
    DISPLAY "[WARNING] ",ls_step
    DISPLAY "[Command] ",ls_err_command
  ELSE
    LET lb_return = TRUE
  END IF

  #一律設為True
  LET lb_return = TRUE
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzp060_unzip_4rp_xrp_file(p_module_name, p_file_name, p_temp_dir, p_zip_type, p_lang,p_doc_type)  #230109-00005#1 
CONSTANT cs_client_xr_design_data = "design_data" #230109-00005#1
DEFINE
  p_module_name STRING, 
  p_file_name   STRING,
  p_temp_dir    STRING,
  p_zip_type    STRING,
  p_lang        STRING,
  p_doc_type    STRING  #230109-00005#1 
DEFINE
  ls_module_name   STRING, 
  ls_file_name     STRING,
  ls_temp_dir      STRING,
  ls_zip_type      STRING,
  ls_lang          STRING,
  ls_doc_type      STRING,  #230109-00005#1 
  ls_mnt_path      STRING,  #230109-00005#1 
  ls_rep_type      STRING,  #230109-00005#1 
  ls_module_path   STRING,
  ls_module_full_path STRING,
  ls_zip_full_name STRING,
  ls_command       STRING,
  lb_success       BOOLEAN,
  lb_chrwx         BOOLEAN, #230109-00005#1 
  ls_step          STRING,
  ls_mnt4rp        STRING,
  ls_source_dir    STRING,
  ls_dest_dir      STRING,
  ls_rep_name      STRING,
  ls_separator     STRING,
  li_count         INTEGER,
  li_lng_count     INTEGER,
  lo_frp_list      DYNAMIC ARRAY OF T_FRP_LIST,
  lo_lang_arr      DYNAMIC ARRAY OF T_LANGUAGE_TYPE
DEFINE 
  lb_return  BOOLEAN  
  
  LET ls_module_name = p_module_name.toUpperCase()
  LET ls_file_name   = p_file_name
  LET ls_temp_dir    = p_temp_dir
  LET ls_zip_type    = p_zip_type
  LET ls_lang        = p_lang
  LET ls_doc_type    = p_doc_type  #230109-00005#1

  LET lb_success   = TRUE
  LET lb_chrwx     = TRUE
  LET lb_return    = TRUE
  LET ls_separator = os.Path.separator()

  #230109-00005#1 begin
  CASE 
    WHEN (ls_doc_type = cs_doc_type_4rp) OR (ls_doc_type = cs_doc_type_4rpupk)
      LET ls_mnt_path = "MNT4RP"
      LET ls_rep_type = "4rp"
    WHEN (ls_doc_type = cs_doc_type_xrp) OR (ls_doc_type = cs_doc_type_xrpupk)
      LET ls_mnt_path = "MNTXRP"
      LET ls_rep_type = "xrp"
  END CASE
  #230109-00005#1 end
  
  #CALL FGL_GETENV("MNT4RP") RETURNING ls_mnt4rp
  CALL cl_rpt_get_env_global(ls_mnt_path) RETURNING ls_mnt4rp #2015/11/02 by Hiko  #230109-00005#1
  
  LET ls_zip_full_name = ls_temp_dir,ls_file_name

  LET ls_module_path = FGL_GETENV(ls_module_name)
  LET ls_module_full_path = ls_module_path,ls_separator,ls_rep_type,ls_separator,ls_lang
  CALL os.Path.chdir(ls_module_full_path) RETURNING lb_success

  CALL sadzp060_get_vfy_rep_list(ls_file_name,ls_temp_dir) RETURNING lo_frp_list
  CALL sadzp060_get_lang_type_list() RETURNING lo_lang_arr

  #解壓縮4rp/xrp檔到暫存目錄中
  LET ls_step = "Uncompress 4rp files to tempary folder."
  CALL sadzp060_set_progress(ls_step)
  DISPLAY cs_new_line
  DISPLAY cs_message_tag,ls_step
  FOR li_lng_count = 1 TO lo_lang_arr.getLength()
    LET ls_lang = lo_lang_arr[li_lng_count] 
    FOR li_count = 1 TO lo_frp_list.getLength()
      LET ls_rep_name = lo_frp_list[li_count].fl_NAME  
      LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_lang,ls_separator,ls_rep_name,".",ls_rep_type," -d '",ls_temp_dir,"'" --," 2> /dev/null"
      --LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' *.4rp -d '",ls_temp_dir,"'" --," 2> /dev/null"
      CALL sadzp060_run_command(ls_command) RETURNING lb_success
      IF NOT lb_success THEN 
        --DISPLAY cs_error_tag,ls_step," failed"
        --EXIT FOR
        DISPLAY cs_warning_tag,ls_step," failed"
        LET lb_success = TRUE
      END IF
      #解壓縮圖檔
      IF lb_success THEN 
        LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_lang,ls_separator,ls_rep_name,".png -d '",ls_temp_dir,"'" --," 2> /dev/null"
        --LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' *.4rp -d '",ls_temp_dir,"'" --," 2> /dev/null"
        CALL sadzp060_run_command(ls_command) RETURNING lb_success
        LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_lang,ls_separator,ls_rep_name,".jpg -d '",ls_temp_dir,"'" --," 2> /dev/null"
        --LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' *.4rp -d '",ls_temp_dir,"'" --," 2> /dev/null"
        CALL sadzp060_run_command(ls_command) RETURNING lb_success
        LET lb_success = TRUE
      END IF 
    END FOR
  END FOR    
  IF NOT lb_success THEN GOTO _RETURN END IF

  #230109-00005#1 begin
  #解壓縮xmd/xdd檔到暫存目錄中
  LET ls_step = "Uncompress xmd/xdd files to tempary folder."
  CALL sadzp060_set_progress(ls_step)
  DISPLAY cs_new_line
  DISPLAY cs_message_tag,ls_step
  FOR li_lng_count = 1 TO lo_lang_arr.getLength()
    LET ls_rep_name = ls_file_name  
    LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",lo_lang_arr[li_lng_count],ls_separator,"xdd",ls_separator,ls_rep_name,".xmd"," -d '",ls_temp_dir,"'" --," 2> /dev/null"
    CALL sadzp060_run_command(ls_command) RETURNING lb_success
    IF NOT lb_success THEN DISPLAY cs_warning_tag,ls_step," failed" LET lb_success = TRUE END IF
    LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",lo_lang_arr[li_lng_count],ls_separator,"xdd",ls_separator,ls_rep_name,".xdd"," -d '",ls_temp_dir,"'" --," 2> /dev/null"
    CALL sadzp060_run_command(ls_command) RETURNING lb_success
    IF NOT lb_success THEN DISPLAY cs_warning_tag,ls_step," failed" LET lb_success = TRUE END IF
    IF NOT lb_success THEN GOTO _RETURN END IF
  END FOR
  #230109-00005#1 end  

  SLEEP 1  
  
  #備份模組的4rp/xrp檔
  LET ls_step = "Backing up ",ls_rep_type," files of module" 
  CALL sadzp060_set_progress(ls_step)
  DISPLAY cs_new_line
  DISPLAY cs_message_tag,ls_step
  FOR li_lng_count = 1 TO lo_lang_arr.getLength()
    LET ls_lang = lo_lang_arr[li_lng_count] 
    FOR li_count = 1 TO lo_frp_list.getLength()
      LET ls_rep_name = lo_frp_list[li_count].fl_NAME  
      LET ls_source_dir = ls_module_path,ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".",ls_rep_type,""
      LET ls_dest_dir   = ls_module_path,ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".",ls_rep_type,".bak"
      DISPLAY cs_copy_from,ls_source_dir
      DISPLAY cs_copy_to,ls_dest_dir
      CALL os.Path.copy(ls_source_dir,ls_dest_dir) RETURNING lb_success
      IF NOT lb_success THEN 
        DISPLAY cs_warning_tag,ls_step," failed"
        --EXIT FOR
      ELSE
        LET ls_step = "Modifying file read/write attribute" 
        CALL sadzp060_set_progress(ls_step)
        DISPLAY cs_new_line
        DISPLAY cs_message_tag,ls_step
        CALL os.Path.chrwx(ls_dest_dir,511) RETURNING lb_chrwx
        IF NOT lb_chrwx THEN
          DISPLAY cs_warning_tag,ls_step," failed"
          LET lb_chrwx = TRUE
        END IF 
      END IF
      #備份圖檔
      IF lb_success THEN 
        LET ls_source_dir = ls_module_path,ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".png"
        LET ls_dest_dir   = ls_module_path,ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".png.bak"
        DISPLAY cs_copy_from,ls_source_dir
        DISPLAY cs_copy_to,ls_dest_dir
        CALL os.Path.copy(ls_source_dir,ls_dest_dir) RETURNING lb_success
        CALL os.Path.chrwx(ls_dest_dir,511) RETURNING lb_chrwx #20240802
        LET ls_source_dir = ls_module_path,ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".jpg"
        LET ls_dest_dir   = ls_module_path,ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".jpg.bak"
        DISPLAY cs_copy_from,ls_source_dir
        DISPLAY cs_copy_to,ls_dest_dir
        CALL os.Path.copy(ls_source_dir,ls_dest_dir) RETURNING lb_success
        CALL os.Path.chrwx(ls_dest_dir,511) RETURNING lb_chrwx #20240802
        LET lb_success = TRUE
      END IF   
      
    END FOR
  END FOR     
  --IF NOT lb_success THEN GOTO _RETURN END IF  

  SLEEP 1

  #複製4rp/xrp檔到該模組目錄中
  LET ls_step = "Cloning the ",ls_rep_type," file to module directory." 
  CALL sadzp060_set_progress(ls_step)
  DISPLAY cs_new_line
  DISPLAY cs_message_tag,ls_step
  FOR li_lng_count = 1 TO lo_lang_arr.getLength()
    LET ls_lang = lo_lang_arr[li_lng_count] 
    FOR li_count = 1 TO lo_frp_list.getLength()
      LET ls_rep_name = lo_frp_list[li_count].fl_NAME  
      LET ls_source_dir = ls_temp_dir,ls_lang,ls_separator,ls_rep_name,".",ls_rep_type,""
      LET ls_dest_dir   = ls_module_path,ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".",ls_rep_type,""
      DISPLAY cs_copy_from,ls_source_dir
      DISPLAY cs_copy_to,ls_dest_dir
      CALL os.Path.copy(ls_source_dir,ls_dest_dir) RETURNING lb_success
      IF NOT lb_success THEN 
        --DISPLAY cs_error_tag,ls_step," failed"
        --EXIT FOR
        DISPLAY cs_warning_tag,ls_step," failed"
        LET lb_success = TRUE
      ELSE
        LET ls_step = "Modifying file read/write attribute" 
        CALL sadzp060_set_progress(ls_step)
        DISPLAY cs_new_line
        DISPLAY cs_message_tag,ls_step
        CALL os.Path.chrwx(ls_dest_dir,511) RETURNING lb_chrwx
        IF NOT lb_chrwx THEN
          DISPLAY cs_warning_tag,ls_step," failed."
          LET lb_chrwx = TRUE
        END IF 
      END IF
      #複製圖檔
      IF lb_success THEN
        LET ls_source_dir = ls_temp_dir,ls_lang,ls_separator,ls_rep_name,".png"
        LET ls_dest_dir   = ls_module_path,ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".png"
        DISPLAY cs_copy_from,ls_source_dir
        DISPLAY cs_copy_to,ls_dest_dir
        CALL os.Path.copy(ls_source_dir,ls_dest_dir) RETURNING lb_success
        CALL os.Path.chrwx(ls_dest_dir,511) RETURNING lb_chrwx #20240802
        LET ls_source_dir = ls_temp_dir,ls_lang,ls_separator,ls_rep_name,".jpg"
        LET ls_dest_dir   = ls_module_path,ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".jpg"
        DISPLAY cs_copy_from,ls_source_dir
        DISPLAY cs_copy_to,ls_dest_dir
        CALL os.Path.copy(ls_source_dir,ls_dest_dir) RETURNING lb_success
        CALL os.Path.chrwx(ls_dest_dir,511) RETURNING lb_chrwx #20240802
        LET lb_success = TRUE
      END IF  
    END FOR   
  END FOR   
  IF NOT lb_success THEN GOTO _RETURN END IF  
  
  SLEEP 1

  #備份報表伺服器中的4rp/xrp檔
  LET ls_step = "Backing up ",ls_rep_type," files which in report server." 
  CALL sadzp060_set_progress(ls_step)
  DISPLAY cs_new_line
  DISPLAY cs_message_tag,ls_step
  FOR li_lng_count = 1 TO lo_lang_arr.getLength()
    LET ls_lang = lo_lang_arr[li_lng_count] 
    FOR li_count = 1 TO lo_frp_list.getLength()
      LET ls_rep_name = lo_frp_list[li_count].fl_NAME  
      LET ls_source_dir = ls_mnt4rp,ls_separator,ls_module_name.toLowerCase(),ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".",ls_rep_type,""
      LET ls_dest_dir   = ls_mnt4rp,ls_separator,ls_module_name.toLowerCase(),ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".",ls_rep_type,".bak"
      DISPLAY cs_copy_from,ls_source_dir
      DISPLAY cs_copy_to,ls_dest_dir
      CALL os.Path.copy(ls_source_dir,ls_dest_dir) RETURNING lb_success
      CALL os.Path.chrwx(ls_dest_dir,511) RETURNING lb_chrwx #20240802
      IF NOT lb_success THEN 
        DISPLAY cs_warning_tag,ls_step," failed"
        --EXIT FOR
      END IF
      #備份圖檔
      IF lb_success THEN
        LET ls_source_dir = ls_mnt4rp,ls_separator,ls_module_name.toLowerCase(),ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".png"
        LET ls_dest_dir   = ls_mnt4rp,ls_separator,ls_module_name.toLowerCase(),ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".png.bak"
        DISPLAY cs_copy_from,ls_source_dir
        DISPLAY cs_copy_to,ls_dest_dir
        CALL os.Path.copy(ls_source_dir,ls_dest_dir) RETURNING lb_success
        CALL os.Path.chrwx(ls_dest_dir,511) RETURNING lb_chrwx #20240802
        LET ls_source_dir = ls_mnt4rp,ls_separator,ls_module_name.toLowerCase(),ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".jpg"
        LET ls_dest_dir   = ls_mnt4rp,ls_separator,ls_module_name.toLowerCase(),ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".jpg.bak"
        DISPLAY cs_copy_from,ls_source_dir
        DISPLAY cs_copy_to,ls_dest_dir
        CALL os.Path.copy(ls_source_dir,ls_dest_dir) RETURNING lb_success
        CALL os.Path.chrwx(ls_dest_dir,511) RETURNING lb_chrwx #20240802
        LET lb_success = TRUE
      END IF 
    END FOR
  END FOR  
  --IF NOT lb_success THEN GOTO _RETURN END IF  

  SLEEP 1

  #複製4rp/xrp檔到報表伺服器各模組目錄中
  LET ls_step = "Cloning the ",ls_rep_type," file to report server module directory." 
  CALL sadzp060_set_progress(ls_step)
  DISPLAY cs_new_line
  DISPLAY cs_message_tag,ls_step
  FOR li_lng_count = 1 TO lo_lang_arr.getLength()
    LET ls_lang = lo_lang_arr[li_lng_count] 
    FOR li_count = 1 TO lo_frp_list.getLength()
      LET ls_rep_name = lo_frp_list[li_count].fl_NAME  
      LET ls_source_dir = ls_temp_dir,ls_lang,ls_separator,ls_rep_name,".",ls_rep_type,""
      LET ls_dest_dir   = ls_mnt4rp,ls_separator,ls_module_name.toLowerCase(),ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".",ls_rep_type,""
      DISPLAY cs_copy_from,ls_source_dir
      DISPLAY cs_copy_to,ls_dest_dir
      CALL os.Path.copy(ls_source_dir,ls_dest_dir) RETURNING lb_success
      CALL os.Path.chrwx(ls_dest_dir,511) RETURNING lb_chrwx #20240802
      IF NOT lb_success THEN 
        --DISPLAY cs_error_tag,ls_step," failed"
        --EXIT FOR
        DISPLAY cs_warning_tag,ls_step," failed"
        LET lb_success = TRUE
      END IF
      #複製圖檔
      IF lb_success THEN
        LET ls_source_dir = ls_temp_dir,ls_lang,ls_separator,ls_rep_name,".png"
        LET ls_dest_dir   = ls_mnt4rp,ls_separator,ls_module_name.toLowerCase(),ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".png"
        DISPLAY cs_copy_from,ls_source_dir
        DISPLAY cs_copy_to,ls_dest_dir
        CALL os.Path.copy(ls_source_dir,ls_dest_dir) RETURNING lb_success
        CALL os.Path.chrwx(ls_dest_dir,511) RETURNING lb_chrwx #20240802
        
        LET ls_source_dir = ls_temp_dir,ls_lang,ls_separator,ls_rep_name,".jpg"
        LET ls_dest_dir   = ls_mnt4rp,ls_separator,ls_module_name.toLowerCase(),ls_separator,ls_rep_type,ls_separator,ls_lang,ls_separator,ls_rep_name,".jpg"
        DISPLAY cs_copy_from,ls_source_dir
        DISPLAY cs_copy_to,ls_dest_dir
        CALL os.Path.copy(ls_source_dir,ls_dest_dir) RETURNING lb_success
        CALL os.Path.chrwx(ls_dest_dir,511) RETURNING lb_chrwx #20240802
        LET lb_success = TRUE
      END IF 
    END FOR   
  END FOR   
  IF NOT lb_success THEN GOTO _RETURN END IF  
  
  SLEEP 1

  #230109-00005#1 begin
  IF (ls_doc_type = cs_doc_type_xrp) OR (ls_doc_type = cs_doc_type_xrpupk) THEN 
    FOR li_lng_count = 1 TO lo_lang_arr.getLength()
      #備份模組的xmd/xdd檔
      LET ls_step = "Backing up xdd/xmd files of module" 
      CALL sadzp060_set_progress(ls_step)
      DISPLAY cs_new_line
      DISPLAY cs_message_tag,ls_step
      LET ls_rep_name = ls_file_name 
      LET ls_source_dir = ls_module_path,ls_separator,ls_rep_type,ls_separator,lo_lang_arr[li_lng_count],ls_separator,"xdd",ls_separator,ls_rep_name,".xdd",""
      LET ls_dest_dir   = ls_module_path,ls_separator,ls_rep_type,ls_separator,lo_lang_arr[li_lng_count],ls_separator,"xdd",ls_separator,ls_rep_name,".xdd",".bak"
      DISPLAY cs_copy_from,ls_source_dir
      DISPLAY cs_copy_to,ls_dest_dir
      CALL os.Path.copy(ls_source_dir,ls_dest_dir) RETURNING lb_success
      IF NOT lb_success THEN DISPLAY cs_warning_tag,ls_step," failed" LET lb_success = TRUE END IF
      CALL os.Path.chrwx(ls_dest_dir,511) RETURNING lb_chrwx
      LET ls_source_dir = ls_module_path,ls_separator,ls_rep_type,ls_separator,lo_lang_arr[li_lng_count],ls_separator,"xdd",ls_separator,ls_rep_name,".xmd",""
      LET ls_dest_dir   = ls_module_path,ls_separator,ls_rep_type,ls_separator,lo_lang_arr[li_lng_count],ls_separator,"xdd",ls_separator,ls_rep_name,".xmd",".bak"
      DISPLAY cs_copy_from,ls_source_dir
      DISPLAY cs_copy_to,ls_dest_dir
      CALL os.Path.copy(ls_source_dir,ls_dest_dir) RETURNING lb_success
      IF NOT lb_success THEN DISPLAY cs_warning_tag,ls_step," failed" LET lb_success = TRUE END IF
      CALL os.Path.chrwx(ls_dest_dir,511) RETURNING lb_chrwx    

      #複製 xmd/xdd 到模組中
      LET ls_step = "Clone xdd/xmd files to destination." 
      CALL sadzp060_set_progress(ls_step)
      DISPLAY cs_new_line
      DISPLAY cs_message_tag,ls_step
      LET ls_source_dir = ls_temp_dir,lo_lang_arr[li_lng_count],ls_separator,"xdd",ls_separator,ls_rep_name,".xdd"
      LET ls_dest_dir   = ls_module_path,ls_separator,ls_rep_type,ls_separator,lo_lang_arr[li_lng_count],ls_separator,"xdd",ls_separator,ls_rep_name,".xdd"
      DISPLAY cs_copy_from,ls_source_dir
      DISPLAY cs_copy_to,ls_dest_dir
      CALL os.Path.copy(ls_source_dir,ls_dest_dir) RETURNING lb_success
      IF NOT lb_success THEN DISPLAY cs_warning_tag,ls_step," failed" LET lb_success = TRUE END IF
      CALL os.Path.chrwx(ls_dest_dir,511) RETURNING lb_chrwx    
      LET ls_source_dir = ls_temp_dir,lo_lang_arr[li_lng_count],ls_separator,"xdd",ls_separator,ls_rep_name,".xmd"
      LET ls_dest_dir   = ls_module_path,ls_separator,ls_rep_type,ls_separator,lo_lang_arr[li_lng_count],ls_separator,"xdd",ls_separator,ls_rep_name,".xmd"
      DISPLAY cs_copy_from,ls_source_dir
      DISPLAY cs_copy_to,ls_dest_dir
      CALL os.Path.copy(ls_source_dir,ls_dest_dir) RETURNING lb_success
      IF NOT lb_success THEN DISPLAY cs_warning_tag,ls_step," failed" LET lb_success = TRUE END IF
      CALL os.Path.chrwx(ls_dest_dir,511) RETURNING lb_chrwx
    END FOR  
  END IF    
  #230109-00005#1 end
  
  LABEL _RETURN:
  
  IF NOT lb_success THEN
    LET lb_return = FALSE
    DISPLAY cs_error_tag,ls_step
  ELSE
    LET lb_return = TRUE
  END IF

  #一律設為True
  --LET lb_return = TRUE
  
  RETURN lb_return
  
END FUNCTION

#170515-00013 begin
FUNCTION sadzp060_unzip_c4fd_file(p_module_name, p_file_name, p_temp_dir, p_zip_type)
DEFINE
  p_module_name STRING, 
  p_file_name   STRING,
  p_temp_dir    STRING,
  p_zip_type    STRING
DEFINE
  ls_module_name   STRING, 
  ls_file_name     STRING,
  ls_temp_dir      STRING,
  ls_zip_type      STRING,
  ls_module_path   STRING,
  ls_zip_full_name STRING,
  ls_command       STRING,
  ls_err_command   STRING,
  lb_success       BOOLEAN,
  lb_complete      BOOLEAN,
  ls_step          STRING,
  ls_separator     STRING,
  ls_tempdir       STRING
DEFINE 
  lb_return  BOOLEAN  
  
  LET ls_module_name = p_module_name.toUpperCase()
  LET ls_file_name   = p_file_name
  LET ls_temp_dir    = p_temp_dir
  LET ls_zip_type    = p_zip_type

  LET ls_separator = os.Path.separator()
  LET lb_success  = TRUE
  LET lb_complete = TRUE
  LET lb_return   = TRUE
  LET ls_err_command = ""

  LET ls_zip_full_name = ls_temp_dir,ls_file_name
  
  LET ls_module_path = FGL_GETENV(ls_module_name)
  LET ls_tempdir     = FGL_GETENV("TEMPDIR")

  LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".4fd -d '",ls_tempdir,"'"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip 4fd file : STEP 1" 
  END IF

  SLEEP 1

  IF NOT lb_complete THEN
    LET lb_return = FALSE
    DISPLAY "[WARNING] ",ls_step
    DISPLAY "[Command] ",ls_err_command
  ELSE
    LET lb_return = TRUE
  END IF

  #一律設為True
  LET lb_return = TRUE
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzp060_unzip_c4gl_file(p_module_name, p_file_name, p_temp_dir, p_zip_type)
DEFINE
  p_module_name STRING, 
  p_file_name   STRING,
  p_temp_dir    STRING,
  p_zip_type    STRING
DEFINE
  ls_module_name   STRING, 
  ls_file_name     STRING,
  ls_temp_dir      STRING,
  ls_zip_type      STRING,
  ls_module_path   STRING,
  ls_zip_full_name STRING,
  ls_command       STRING,
  ls_err_command   STRING,
  lb_success       BOOLEAN,
  lb_complete      BOOLEAN,
  ls_step          STRING,
  ls_separator     STRING,
  ls_tempdir       STRING
DEFINE 
  lb_return  BOOLEAN  
  
  LET ls_module_name = p_module_name.toUpperCase()
  LET ls_file_name   = p_file_name
  LET ls_temp_dir    = p_temp_dir
  LET ls_zip_type    = p_zip_type

  LET ls_separator = os.Path.separator()
  LET lb_success  = TRUE
  LET lb_complete = TRUE
  LET lb_return   = TRUE
  LET ls_err_command = ""

  LET ls_zip_full_name = ls_temp_dir,ls_file_name
  
  LET ls_module_path = FGL_GETENV(ls_module_name)
  LET ls_tempdir     = FGL_GETENV("TEMPDIR")

  LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".4gl -d '",ls_tempdir,"'"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip 4gl file : STEP 1" 
  END IF

  SLEEP 1

  IF NOT lb_complete THEN
    LET lb_return = FALSE
    DISPLAY "[WARNING] ",ls_step
    DISPLAY "[Command] ",ls_err_command
  ELSE
    LET lb_return = TRUE
  END IF

  #一律設為True
  LET lb_return = TRUE
  
  RETURN lb_return
  
END FUNCTION
#170515-00013 end

FUNCTION sadzp060_4rp_xrp_gen_multi_lang(p_module_name,p_file_name,p_doc_type)
DEFINE
  p_module_name STRING,
  p_file_name   STRING,
  p_doc_type    STRING  #230109-00005#1
DEFINE
  ls_module_name STRING,
  ls_file_name   STRING,
  ls_doc_type    STRING,  #230109-00005#1
  lb_result      BOOLEAN
DEFINE
  lb_return BOOLEAN

  LET ls_module_name = p_module_name
  LET ls_file_name   = p_file_name
  LET ls_doc_type    = p_doc_type  #230109-00005#1

  LET lb_result = TRUE

  #230109-00005#1 begin
  CASE 
    WHEN (ls_doc_type = cs_doc_type_4rp) OR (ls_doc_type = cs_doc_type_4rpupk)
      TRY
        DISPLAY "[Check Point]Call sadzp188_multilang_4rp."
        CALL sadzp188_multilang_4rp("",ls_module_name,ls_file_name,"","zh_TW","zh_CN","Y") RETURNING lb_result
      CATCH
        LET lb_result = FALSE
      END TRY  
    WHEN (ls_doc_type = cs_doc_type_xrp) OR (ls_doc_type = cs_doc_type_xrpupk)
      {
      #先保留供 xrp 產出多語言用 
      TRY
        CALL xxxxxxx_multilang_xrp("",ls_module_name,ls_file_name,"","zh_TW","zh_CN","Y") RETURNING lb_result
      CATCH
        LET lb_result = FALSE
      END TRY
      }  
  END CASE
  #230109-00005#1 end  
  
  LET lb_return = lb_result
  
  RETURN lb_return
  
END FUNCTION 

FUNCTION sadzp060_check_report_server_valid(p_module_name, p_lang, p_show_dialog,p_doc_type) #180427-00055
DEFINE
  p_module_name STRING, 
  p_lang        STRING,
  p_show_dialog BOOLEAN, #180427-00055
  p_doc_type    STRING  #230109-00005#1
DEFINE
  ls_module_name  STRING, 
  ls_lang         STRING,
  lb_show_dialog  BOOLEAN, #180427-00055
  ls_doc_type     STRING,  #230109-00005#1
  ls_mnt_env      STRING,  #230109-00005#1
  ls_rep_split    STRING,  #230109-00005#1
  ls_separator    STRING,
  ls_module_path  STRING,
  ls_mnt_path     STRING,
  lb_success      BOOLEAN,
  ls_pwd          STRING,
  ls_err_code     STRING,
  ls_err_msg      STRING,
  ls_full_path    STRING
DEFINE
  lb_return BOOLEAN  
  
  LET ls_module_name = p_module_name.toUpperCase()
  LET ls_lang        = p_lang
  LET lb_show_dialog = p_show_dialog #180427-00055
  LET ls_doc_type    = p_doc_type    #230109-00005#1 
  
  LET ls_separator = os.Path.separator()

  #230109-00005#1 begin
  CASE
    WHEN (ls_doc_type = cs_doc_type_4rpupk) OR (ls_doc_type = cs_doc_type_4rp)
      LET ls_mnt_env = "MNT4RP"
      LET ls_rep_split = "4rp"
    WHEN (ls_doc_type = cs_doc_type_xrpupk) OR (ls_doc_type = cs_doc_type_xrp)
      LET ls_mnt_env = "MNTXRP"
      LET ls_rep_split = "xrp"
  END CASE
  #230109-00005#1 end 
  
  #CALL FGL_GETENV("MNT4RP") RETURNING ls_mnt4rp
  CALL cl_rpt_get_env_global(ls_mnt_env) RETURNING ls_mnt_path #2015/11/02 by Hiko #230109-00005#1
  DISPLAY cs_message_tag,"Checking report server mount ",ls_mnt_env," path : ",ls_mnt_path

  #先備份原先的目錄
  LET ls_pwd = os.Path.pwd()
  LET ls_module_path = FGL_GETENV(ls_module_name)

  #試著切換到Report Server的目錄看看是否成功
  LET lb_success = FALSE
  TRY 
    LET ls_full_path = ls_mnt_path,ls_separator,ls_module_name.toLowerCase(),ls_separator,ls_rep_split,ls_separator,ls_lang #230109-00005#1
    CALL os.Path.chdir(ls_full_path) RETURNING lb_success
  CATCH 
    LET lb_success = FALSE
  END TRY
  IF NOT lb_success THEN DISPLAY cs_error_tag,"Switch to path ",ls_full_path," not correct." END IF

  IF NOT lb_success THEN #180427-00055 add
    IF lb_show_dialog THEN 
      LET ls_err_code = "adz-00271"
      LET ls_err_msg  = "|"
      CALL sadzp000_msg_show_error(ls_err_code, ls_err_code, ls_err_msg, 1)
    ELSE   #180427-00055 begin
      DISPLAY cs_error_tag,"The template file unable to upload to the report host, please check the connection with the report host!"
    END IF #180427-00055 end  
  END IF
  
  LET lb_return = lb_success

  #再跳回原來的目錄
  CALL os.Path.chdir(ls_pwd) RETURNING lb_success
  
  RETURN lb_return
    
END FUNCTION 

FUNCTION sadzp060_get_vfy_rep_list(p_file_name,p_temp_dir)
DEFINE
  p_file_name STRING,
  p_temp_dir  STRING
DEFINE 
  ls_file_name STRING,
  ls_temp_dir  STRING,
  ls_vfy_file  STRING,
  ls_separator STRING,
  li_count     INTEGER,
  lo_xml_file  xml.domDocument,
  lo_dom_node  xml.DomNode,
  lo_frp_node  xml.DomNode
DEFINE
  lo_frp_list  DYNAMIC ARRAY OF T_FRP_LIST

  LET ls_file_name = p_file_name
  LET ls_temp_dir  = p_temp_dir

  LET ls_separator = os.Path.separator()

  LET ls_vfy_file = ls_temp_dir,ls_separator,ls_file_name,".vfy"

  LET lo_xml_file = xml.domDocument.create()
  DISPLAY ls_vfy_file
  CALL lo_xml_file.load(ls_vfy_file)
  LET lo_dom_node = lo_xml_file.getDocumentElement()
  LET lo_frp_node = lo_dom_node.getFirstChildElement()

  LET li_count = 1  
  WHILE (lo_frp_node IS NOT NULL)

    LET lo_frp_list[li_count].fl_NAME     = lo_frp_node.getAttribute("name") 
    LET lo_frp_list[li_count].fl_EXIST    = lo_frp_node.getAttribute("exist") 
    LET lo_frp_list[li_count].fl_TEMPLATE = lo_frp_node.getAttribute("template")

    LET li_count = li_count + 1 
    
    LET lo_frp_node = lo_frp_node.getNextSiblingElement()
  END WHILE 
  
  RETURN lo_frp_list
  
END FUNCTION

FUNCTION sadzp060_unzip_rsd_file(p_module_name, p_file_name, p_temp_dir, p_zip_type)
DEFINE
  p_module_name STRING, 
  p_file_name   STRING,
  p_temp_dir    STRING,
  p_zip_type    STRING
DEFINE
  ls_module_name   STRING, 
  ls_file_name     STRING,
  ls_temp_dir      STRING,
  ls_zip_type      STRING,
  ls_module_path   STRING,
  ls_zip_full_name STRING,
  ls_command       STRING,
  ls_err_command   STRING,
  lb_success       BOOLEAN,
  lb_complete      BOOLEAN,
  ls_separator     STRING,
  ls_step          STRING
DEFINE 
  lb_return  BOOLEAN  
  
  LET ls_module_name = p_module_name.toUpperCase()
  LET ls_file_name   = p_file_name
  LET ls_temp_dir    = p_temp_dir
  LET ls_zip_type    = p_zip_type

  LET lb_success  = TRUE
  LET lb_complete = TRUE
  LET lb_return   = TRUE
  LET ls_err_command = ""

  LET ls_zip_full_name = ls_temp_dir,ls_file_name
  LET ls_separator   = os.Path.separator()
  
  LET ls_module_path = FGL_GETENV(ls_module_name)

  #備份
  LET ls_command = "cp ",ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".rsd ",ls_module_path,ls_separator,"dzx",ls_separator,"tsd",ls_separator,ls_file_name,".rsd.bak"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 1" 
  END IF
  
  LET ls_command = "unzip -o '",ls_zip_full_name,".",ls_zip_type,"' ",ls_file_name,".rsd -d '",ls_module_path,ls_separator,"dzx",ls_separator,"tsd","'"
  CALL sadzp060_run_command(ls_command) RETURNING lb_success
  IF NOT lb_success THEN 
    LET lb_complete = FALSE
    LET ls_err_command = ls_command
    LET ls_step = "Unzip program file : STEP 2" 
  END IF
  
  SLEEP 1

  IF NOT lb_complete THEN
    LET lb_return = FALSE
    DISPLAY "[WARNING] ",ls_step
    DISPLAY "[Command] ",ls_err_command
  ELSE
    LET lb_return = TRUE
  END IF

  #一律設為True
  LET lb_return = TRUE
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzp060_creat_temp_directory(p_file_name)
DEFINE
  p_file_name  STRING
DEFINE
  ls_file_name   STRING,
  ls_zip_path    STRING,
  ls_dir_pid     STRING,
  li_mkdir       INTEGER,
  ls_separator   STRING,
  ls_random_name STRING
DEFINE
  ls_return STRING  

  LET ls_file_name   = p_file_name

  LET ls_separator   = os.Path.separator()
  CALL sadzp060_gen_random_name() RETURNING ls_random_name

  LET ls_zip_path = FGL_GETENV("TEMPDIR")
  LET ls_dir_pid = ls_zip_path,ls_separator,ls_file_name,"_",ui.Interface.getName(),"_",ls_random_name
  CALL os.Path.mkdir(ls_dir_pid) RETURNING li_mkdir
  
  LET ls_return = ls_dir_pid

  RETURN ls_return
  
END FUNCTION

#以亂數產出副檔名
FUNCTION sadzp060_gen_random_name()
DEFINE
  lr_random_name RECORD
    segment1 STRING,
    segment2 STRING,
    segment3 STRING,
    segment4 STRING
  END RECORD
DEFINE 
  li_random_value  INTEGER,
  li_max_random_num INTEGER,
  ls_final_name    STRING,
  ls_using_format  STRING
DEFINE  
  ls_return  STRING

  LET li_max_random_num = 9999
  LET ls_using_format  = "&&&&"
  
  CALL util.math.rand(li_max_random_num) RETURNING li_random_value
  LET lr_random_name.segment1 = li_random_value USING ls_using_format
  CALL util.math.rand(li_max_random_num) RETURNING li_random_value
  LET lr_random_name.segment2 = li_random_value USING ls_using_format
  CALL util.math.rand(li_max_random_num) RETURNING li_random_value
  LET lr_random_name.segment3 = li_random_value USING ls_using_format
  CALL util.math.rand(li_max_random_num) RETURNING li_random_value
  LET lr_random_name.segment4 = li_random_value USING ls_using_format
  
  LET ls_final_name = lr_random_name.segment1,".",
                      lr_random_name.segment2,".",
                      lr_random_name.segment3,".",
                      lr_random_name.segment4

  LET ls_return = ls_final_name
  
  RETURN ls_return                     
  
END FUNCTION

FUNCTION sadzp060_set_progress(p_message)
DEFINE
  p_message STRING
DEFINE
  ls_message      STRING,
  ls_progress     STRING

  LET ls_message = p_message
  
  LET mi_progress = mi_progress + 1
  LET ls_progress = mi_progress
  
  DISPLAY cs_progress_tag,ls_progress
  
  IF NVL(ls_message,"X") <> "X" THEN
    DISPLAY cs_message_tag,ls_message
  END IF
  
END FUNCTION

FUNCTION sadzp060_run_command(p_command)
DEFINE 
  p_command STRING  
DEFINE 
  li_success SMALLINT, 
  ls_command STRING
DEFINE
  lb_return BOOLEAN  

  LET ls_command = p_command

  DISPLAY cs_command_tag,ls_command  

  #每一個Run也是一個Progress
  CALL sadzp060_set_progress(ls_command)
  RUN ls_command RETURNING li_success
  
  #DISPLAY "OS : ",FGL_GETENV("ostype")," Success : ",li_success

  {
  IF (FGL_GETENV("ostype") MATCHES "Win*") THEN
    LET li_success = li_success 
  ELSE
    LET li_success = ( li_success / 256 )
  END IF  
  }
  
  LET li_success = ( li_success / 256 )

  #DISPLAY "OS : ",FGL_GETENV("ostype")," Success : ",li_success

  IF li_success = 0 THEN
    LET lb_return = TRUE
  ELSE
    LET lb_return = FALSE
  END IF

  LET lb_return = TRUE

  RETURN lb_return
  
END FUNCTION

FUNCTION sadzp060_get_sadzp060_progress_all(p_prog_type)
DEFINE
  p_prog_type STRING
DEFINE
  ls_prog_type    STRING,
  ls_progress_all VARCHAR(5),
  ls_sql          STRING
DEFINE
  ls_return STRING  

  LET ls_prog_type = p_prog_type
  
  LET ls_sql = "SELECT EJ.DZEJ006 PROGRESS_ALL        ",
               "  FROM DZEJ_T EJ                      ",
               " WHERE EJ.DZEJ001 = 'sadzp060_Progress'",  
               "   AND EJ.DZEJ003 = 'ALL'             ",
               "   AND EJ.DZEJ004 = '",ls_prog_type,"'"
 
  PREPARE lpre_get_sadzp060_progress_all FROM ls_sql
  DECLARE lcur_get_sadzp060_progress_all CURSOR FOR lpre_get_sadzp060_progress_all

  OPEN lcur_get_sadzp060_progress_all
  FETCH lcur_get_sadzp060_progress_all INTO ls_progress_all
  CLOSE lcur_get_sadzp060_progress_all
  
  FREE lpre_get_sadzp060_progress_all
  FREE lcur_get_sadzp060_progress_all  

  LET ls_return = ls_progress_all
  
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzp060_set_sadzp060_progress_all(p_prog_type,p_progress)
DEFINE
  p_prog_type STRING,
  p_progress  STRING
DEFINE
  ls_prog_type STRING,
  ls_progress  STRING,
  ls_sql       STRING

  LET ls_prog_type = p_prog_type
  LET ls_progress  = p_progress

  LET ls_sql = "UPDATE DZEJ_T EJ                      ", 
               "   set ej.dzej006 = '",ls_progress,"' ",
               " WHERE EJ.DZEJ001 = 'sadzp060_Progress'",  
               "   AND EJ.DZEJ003 = 'ALL'             ",
               "   AND EJ.DZEJ004 = '",ls_prog_type,"'"

  BEGIN WORK

  TRY
    PREPARE lpre_set_sadzp060_progress_all FROM ls_sql
    EXECUTE lpre_set_sadzp060_progress_all
    COMMIT WORK
  CATCH
    DISPLAY "[Error] Update sadzp060 progress. "
    ROLLBACK WORK
  END TRY  
  
END FUNCTION

FUNCTION sadzp060_check_zip_content_valid(p_parameters,p_file_name,p_temp_dir, p_zip_type)
DEFINE
  p_parameters      T_UPLOAD_PARAM,
  p_file_name       STRING,
  p_temp_dir        STRING,
  p_zip_type        STRING
DEFINE
  ls_file_name       STRING,
  ls_temp_dir        STRING,
  ls_zip_type        STRING,
  ls_upload_doc_type STRING,
  ls_message         STRING,
  lb_success         BOOLEAN,
  lo_parameters      T_UPLOAD_PARAM
DEFINE 
  lo_return_params   T_UPLOAD_PARAM,
  lb_return          BOOLEAN  
  
  LET ls_file_name       = p_file_name
  LET ls_temp_dir        = p_temp_dir
  LET ls_zip_type        = p_zip_type
  LET ls_upload_doc_type = p_parameters.arg3_upload_doc_type

  LET lo_parameters.* = p_parameters.*

  LET lb_success  = TRUE
  LET lb_return   = TRUE
  LET ls_message  = ""

  LET ls_message = "Uncompress to temp dir."
  CALL sadzp060_unzip_to_temp_dir(ls_file_name,ls_temp_dir,ls_zip_type,ls_upload_doc_type) RETURNING lb_success
  IF NOT lb_success THEN GOTO _check_error END IF

  LET ls_message = "Check compressed file valid."

  #2014.05.20 該為全部呼叫sadzp060_2_chk_permission
  {
  IF ls_upload_doc_type = cs_doc_type_cspec THEN
    #上傳格式為 CSPEC 時處理
    CALL sadzp040_2_chk_permission(ls_file_name, ls_temp_dir, ls_upload_doc_type) RETURNING lb_success,ls_version
    IF NOT lb_success THEN GOTO _check_error END IF
  ELSE 
    CALL sadzp060_2_chk_permission(ls_file_name, ls_temp_dir, ls_upload_doc_type) RETURNING lb_success,ls_version
    IF NOT lb_success THEN GOTO _check_error END IF
  END IF   
  }

  CALL sadzp060_2_chk_permission(ls_file_name, ls_temp_dir, ls_upload_doc_type) RETURNING lo_return_params.*,ls_message
  LET lb_success = IIF(ls_message.trim()IS NULL,TRUE,FALSE)
  IF NOT lb_success THEN GOTO _check_error END IF

  #將回傳後的 T_UPLOAD_PARAM 的前四項值回填
  LET lo_return_params.arg1_program_name    = ls_file_name --lo_parameters.arg1_program_name
  --LET lo_return_params.arg2_module_name    = lo_parameters.arg2_module_name
  LET lo_return_params.arg3_upload_doc_type = lo_parameters.arg3_upload_doc_type
  LET lo_return_params.arg4_client_path     = lo_parameters.arg4_client_path
  
  LABEL _check_error:
  
  IF NOT lb_success THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF

  RETURN lb_return,ls_message,lo_return_params.*
  
END FUNCTION

FUNCTION sadzp060_show_message_designer()
  DISPLAY cs_message_half_separator,cs_message_designer,cs_message_half_separator
END FUNCTION

FUNCTION sadzp060_show_message_generator()
  DISPLAY cs_message_half_separator,cs_message_generator,cs_message_half_separator
END FUNCTION

FUNCTION sadzp060_show_message_user_control()
  DISPLAY cs_message_half_separator,cs_message_user_control,cs_message_half_separator
END FUNCTION

FUNCTION sadzp060_show_message_separator()
  DISPLAY cs_message_separator
END FUNCTION

FUNCTION sadzp060_compress_client_files(p_work_dir,p_doc_type)
DEFINE
  p_work_dir  STRING,
  p_doc_type  STRING
DEFINE
  ls_work_dir    STRING,
  ls_doc_type    STRING,
  ls_path_name   STRING,    
  ls_file_name   STRING,
  ls_caption     STRING, #230109-00005#1
  lo_str_buf     base.StringBuffer, #230109-00005#1
  ls_xr_base     STRING, #230109-00005#1
  ls_xr_prog     STRING, #230109-00005#1
  ls_vfy_name    STRING, #230109-00005#1
  lb_vfy_exists  BOOLEAN, #230109-00005#1
  ls_doscmd      STRING, #230109-00005#1
  lb_result      BOOLEAN, #230109-00005#1
  li_integer     INTEGER,
  ls_temp_name   STRING,
  ls_temp_char   STRING,
  lo_open_dialog T_FILE_DIALOG
DEFINE
  lb_return BOOLEAN   

  LET ls_work_dir = p_work_dir
  LET ls_doc_type = p_doc_type

  LET lb_return = TRUE
  LET lb_vfy_exists = FALSE #230109-00005#1

  #230109-00005#1 begin
  CASE 
    WHEN ls_doc_type = cs_doc_type_4rp
      LET ls_caption = cs_open_4rp_dialog
    WHEN ls_doc_type = cs_doc_type_xrp
      LET ls_caption = cs_open_xrp_dialog
  END CASE

  #Check verify file exists.
  IF (ls_work_dir.getIndexOf("(s)",1) > 0) OR (ls_work_dir.getIndexOf("(c)",1) > 0) THEN
    LET lo_str_buf = base.StringBuffer.create()
    CALL lo_str_buf.append(ls_work_dir)
    CALL lo_str_buf.replace("\\","/",0)
    LET ls_path_name = lo_str_buf.toString()
    CALL lo_str_buf.clear()
    LET ls_xr_base = os.Path.basename(ls_path_name)
    LET ls_xr_prog = ls_xr_base.subString(1,ls_xr_base.getIndexOf("(",1)-1)
    LET ls_vfy_name = ls_work_dir,"\\",ls_xr_prog,".vfy"
    LET ls_doscmd = SFMT('cmd /C if exist "%1" (exit 0) else (exit 1)',ls_vfy_name)
    CALL ui.Interface.frontCall("standard","execute",[ls_doscmd,1],[lb_result])    
    LET lb_vfy_exists = lb_result
  END IF    
  #230109-00005#1 end  

  IF NOT lb_vfy_exists THEN #230109-00005#1
    LET lo_open_dialog.PATH = ls_work_dir
    LET lo_open_dialog.CAPTION = ls_caption 
    CALL sadzp060_open_dir_dialog(lo_open_dialog.*) RETURNING ls_path_name
  END IF #230109-00005#1
  
  --DISPLAY "ls_work_dir : ",ls_work_dir
  --DISPLAY "ls_path_name : ",ls_path_name

  IF NVL(ls_path_name.trim(),cs_null_default) <> cs_null_default THEN 
    LET ls_temp_name = ""
    FOR li_integer = ls_path_name.getLength() TO 1 STEP -1
      LET ls_temp_char = ls_path_name.getCharAt(li_integer)
      IF (ls_temp_char = cs_dos_separator) OR (ls_temp_char = cs_unix_separator) THEN 
        EXIT FOR 
      END IF   
      LET ls_temp_name = ls_temp_char,ls_temp_name
    END FOR 
    LET ls_file_name = ls_temp_name 
    
    DISPLAY cs_message_tag,"You selected the PATH : ",ls_path_name
    DISPLAY cs_message_tag,"Filename : ",ls_file_name
    CALL sadzp060_compress_client_4rp_xrp(ls_file_name,ls_path_name) RETURNING lb_result
    LET lb_return = lb_result
  ELSE 
    LET lb_return = FALSE
  END IF  

  RETURN lb_return,ls_file_name
  
END FUNCTION 

FUNCTION sadzp060_open_dir_dialog(p_save_dialog)
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

FUNCTION sadzp060_compress_client_4rp_xrp(p_file_name,p_work_path)
DEFINE
  p_file_name STRING, 
  p_work_path STRING
DEFINE
  ls_file_name STRING, 
  ls_work_path STRING,
  ls_command   STRING,
  ls_separator STRING,
  ls_7zip_path  STRING,
  ls_source    STRING,
  ls_destination STRING,
  lb_result    BOOLEAN 

  LET ls_file_name = p_file_name
  LET ls_work_path = p_work_path

  LET ls_separator = os.Path.separator()
  LET ls_7zip_path = FGL_GETENV("UTL"),ls_separator,"3rd",ls_separator,"7zip",ls_separator

  LET lb_result = TRUE
  
  ------------------------------------------------------------------------------
  #測試是否有解壓程式
  LET ls_command = cs_client_temp_dir,cs_dos_separator,"7za.exe"
  CALL ui.Interface.frontCall("standard", "execute", [ls_command,1], [lb_result])
  IF lb_result THEN 
    DISPLAY cs_message_tag,"Check compress/uncompress program OK !!" 
  ELSE 
    DISPLAY cs_warning_tag,"Check compress/uncompress program unsuccess !!" 
  
    #建立 Client 暫存目錄
    LET ls_command = "cmd.exe /C \"mkdir ",cs_client_temp_dir,""
    CALL ui.Interface.frontCall("standard", "execute", [ls_command,1], [lb_result])
    IF lb_result THEN DISPLAY cs_message_tag,"Create client temp directory OK !" ELSE DISPLAY cs_warning_tag,"Create client temp directory unsuccess : ",ls_command END IF

    #將暫存目錄設為隱藏
    {
    LET ls_command = "cmd.exe /C \"attrib +h ",cs_client_temp_dir,""
    CALL ui.Interface.frontCall("standard", "execute", [ls_command,1], [lb_result])
    IF lb_result THEN DISPLAY cs_message_tag,"Set client temp directory hidden OK !" ELSE DISPLAY cs_warning_tag,"Set client temp directory hidden unsuccess : ",ls_command END IF
    }
    
    #下載7za.exe到 c:\t100temp
    LET ls_source      = ls_7zip_path,"7za.exe"
    LET ls_destination = cs_client_temp_dir,cs_dos_separator,"7za.exe"
    TRY
      DISPLAY cs_message_tag,"Downloading compress/uncompress program ....." 
      LET lb_result = TRUE
      CALL FGL_PUTFILE(ls_source,ls_destination)
      CALL sadzp060_grant_clinet_permission(ls_destination) RETURNING lb_result #170111-00006
    CATCH
      LET lb_result = FALSE
    END TRY   
    IF lb_result THEN DISPLAY cs_message_tag,"Download compress/uncompress program OK !" ELSE DISPLAY cs_error_tag,"Download compress/uncompress program unsuccess : ",ls_command END IF
    
  END IF 
  ------------------------------------------------------------------------------
  
  ------------------------------------------------------------------------------

  #移除暫存目錄中舊的tzt檔
  LET ls_command = "cmd.exe /C \"del -F -Q ",cs_client_temp_dir,cs_dos_separator,ls_file_name,cs_upload_string,".tzt","" 
  CALL ui.Interface.frontCall("standard", "execute", [ls_command,1], [lb_result])
  IF lb_result THEN DISPLAY cs_message_tag,"Delete tempory tzt file OK !" ELSE DISPLAY cs_warning_tag,"Delete tempory tzt file unsuccess : ",ls_command END IF

  #壓縮4rp檔為.tzt
  LET ls_command = cs_client_temp_dir,cs_dos_separator,"7za.exe a -tzip ",cs_client_temp_dir,cs_dos_separator,ls_file_name,cs_upload_string,".tzt ",ls_work_path,cs_dos_separator,"*"
  CALL ui.Interface.frontCall("standard", "execute", [ls_command,1], [lb_result])
  IF lb_result THEN DISPLAY cs_message_tag,"Compress 4rp/xrp files OK" ELSE DISPLAY cs_error_tag,"Compress 4rp/xrp files unsuccess : ",ls_command GOTO _return  END IF

  #將vfy檔設為可見
  {
  LET ls_command = "cmd.exe /C \"attrib -h ",ls_work_path,cs_dos_separator,ls_file_name,".vfy"
  CALL ui.Interface.frontCall("standard", "execute", [ls_command,1], [lb_result])
  IF lb_result THEN DISPLAY cs_message_tag,"Set vfy file unhidden OK !" ELSE DISPLAY cs_warning_tag,"Set vfy file unhidden unsuccess : ",ls_command END IF
  }
  
  #壓縮vfy檔為.tzt
  {
  #170111-00006 mark begin
  LET ls_command = cs_client_temp_dir,cs_dos_separator,"7za.exe a -tzip ",cs_client_temp_dir,cs_dos_separator,ls_file_name,cs_upload_string,".tzt ",ls_work_path,cs_dos_separator,"*.vfy"
  CALL ui.Interface.frontCall("standard", "execute", [ls_command,1], [lb_result])
  IF lb_result THEN DISPLAY cs_message_tag,"Compress vfy files OK" ELSE DISPLAY cs_error_tag,"Compress vfy files unsuccess : ",ls_command GOTO _return END IF
  #170111-00006 mark end
  }

  #將vfy檔設為隱藏
  {
  LET ls_command = "cmd.exe /C \"attrib +h ",ls_work_path,cs_dos_separator,ls_file_name,".vfy"
  CALL ui.Interface.frontCall("standard", "execute", [ls_command,1], [lb_result])
  IF lb_result THEN DISPLAY cs_message_tag,"Set vfy file hidden OK !" ELSE DISPLAY cs_warning_tag,"Set vfy file hidden unsuccess : ",ls_command END IF
  }
  ------------------------------------------------------------------------------
  
  LABEL _return:

  RETURN lb_result
  
END FUNCTION 

FUNCTION sadzp060_get_lang_type_list()
DEFINE
  lo_lang_arr  DYNAMIC ARRAY OF T_LANGUAGE_TYPE,
  li_count     INTEGER,
  ls_sql       STRING 
DEFINE
  lo_return DYNAMIC ARRAY OF T_LANGUAGE_TYPE

  LET ls_sql = "SELECT GZZY001    ", 
               "  FROM GZZY_T     ", 
               " ORDER BY GZZY001 "
               
  PREPARE lpre_get_lang_type_list FROM ls_sql
  DECLARE lcur_get_lang_type_list CURSOR FOR lpre_get_lang_type_list

  LET li_count = 1
  
  OPEN lcur_get_lang_type_list
  FOREACH lcur_get_lang_type_list INTO lo_lang_arr[li_count]  
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_count = li_count + 1

  END FOREACH
  CLOSE lcur_get_lang_type_list

  FREE lcur_get_lang_type_list
  FREE lpre_get_lang_type_list
  
  CALL lo_lang_arr.deleteElement(li_count)
  
  LET lo_return = lo_lang_arr

  RETURN lo_return
  
END FUNCTION 

#170111-00006 begin
FUNCTION sadzp060_grant_clinet_permission(p_full_name)
DEFINE
  p_full_name STRING
DEFINE
  ls_full_name STRING,
  ls_command   STRING,
  lb_result    BOOLEAN
DEFINE
  lb_return  BOOLEAN  

  LET ls_full_name = p_full_name

  LET lb_result = TRUE
  
  LET ls_command = "cmd.exe /C \"icacls ",ls_full_name," /grant Everyone:F"
  CALL ui.Interface.frontCall("standard", "execute", [ls_command,1], [lb_result])
  IF lb_result THEN DISPLAY cs_message_tag,"Grant permission to ",ls_full_name," OK !" ELSE DISPLAY cs_warning_tag,"Grant permission to ",ls_full_name," unsuccess : ",ls_command END IF

  LET lb_return = lb_result

  RETURN lb_return
  
END FUNCTION
#170111-00006 end

#170626-00037 begin
FUNCTION sadzp060_check_design_file_if_have_contents(p_file_name)
DEFINE
  p_file_name  STRING
DEFINE
  ls_file_name     STRING,
  lo_channel_read  base.Channel,
  lo_str_buf       base.StringBuffer,
  lb_result        BOOLEAN,
  ls_text          VARCHAR(4096),
  li_rec_cnt       INTEGER
DEFINE
  lb_return  BOOLEAN  
  
  LET ls_file_name  = p_file_name
  
  LET lb_return = TRUE
  LET lb_result = TRUE

  LET lo_channel_read = base.Channel.create()
  TRY 
    CALL lo_channel_read.openFile(ls_file_name,"r")
  CATCH
    LET lb_result = FALSE
    DISPLAY cs_error_tag,"Open file ",ls_file_name," fault !!"
  END TRY  

  IF lb_result THEN
    
    DISPLAY cs_message_tag,"Read design file : '",ls_file_name,"'."
    LET li_rec_cnt = 1 
    LET lo_str_buf = base.StringBuffer.create()
    CALL lo_str_buf.clear()

    LET ls_text = lo_channel_read.readLine()
    #DISPLAY cs_message_tag,"The first line of file is : ",ls_text
    IF ls_text CLIPPED IS NULL THEN
      DISPLAY cs_message_tag,"The contents of file '",ls_file_name,"' is blank."
      LET lb_result = FALSE 
    END IF
    
  END IF
  
  CALL lo_channel_read.close()

  LET lb_return = lb_result

  RETURN lb_return
  
END FUNCTION
#170626-00037 end

#180427-00055 begin
FUNCTION sadzp060_check_zip_file_if_exists(p_file_path)
DEFINE
  p_file_path STRING 
DEFINE
  ls_file_path STRING,
  lb_result    BOOLEAN
DEFINE
  lb_return  BOOLEAN  

  LET ls_file_path = p_file_path

  LET lb_return = TRUE

  CALL os.Path.exists(ls_file_path) RETURNING lb_result

  LET lb_return = lb_result 

  RETURN lb_return 
  
END FUNCTION

FUNCTION sadzp060_move_4rp_xrp_zip_to_worktemp(p_file_name,p_source_path,p_destination_path)
DEFINE
  p_file_name        STRING,
  p_source_path      STRING,
  p_destination_path STRING
DEFINE
  ls_file_name        STRING,
  ls_source_path      STRING,
  ls_destination_path STRING,
  ls_base_name        STRING,
  lb_result           BOOLEAN, 
  ls_source_file      STRING,
  ls_destination_file STRING,
  ls_src_file_name    STRING,
  ls_dst_file_name    STRING,
  lb_success          BOOLEAN
DEFINE 
  lb_return BOOLEAN,
  ls_return STRING  

  LET ls_file_name        = p_file_name
  LET ls_source_path      = p_source_path        
  LET ls_destination_path = p_destination_path
  
  LET lb_success = TRUE 

  LET ls_src_file_name = ls_file_name
  #傳到暫存的時候一律改回原名稱的壓縮檔
  IF ls_file_name.getIndexOf(cs_open_paren,1) > 0 THEN
    LET ls_dst_file_name = ls_file_name.subString(1,ls_file_name.getIndexOf(cs_open_paren,1)-1),ls_file_name.subString(ls_file_name.getIndexOf(".",1),ls_file_name.getLength())
    LET ls_base_name     = ls_file_name.subString(1,ls_file_name.getIndexOf(cs_open_paren,1)-1)
  ELSE
    LET ls_dst_file_name = ls_file_name
  END IF   
  
  LET ls_source_file      = os.Path.join(ls_source_path, ls_src_file_name)
  LET ls_destination_file = os.Path.join(ls_destination_path, ls_dst_file_name)

  TRY
    CALL os.Path.copy(ls_source_file,ls_destination_file) RETURNING lb_result
    DISPLAY cs_message_tag,"Move file from ",ls_source_file," to ",ls_destination_file," success ! "
  CATCH
    LET lb_success = FALSE 
    DISPLAY cs_error_tag,"Move file from ",ls_source_file," to ",ls_destination_file," fail ! "
  END TRY  

  LET lb_return = lb_success
  LET ls_return = ls_base_name

  RETURN lb_return,ls_return

END FUNCTION
#180427-00055 end

#20210517 begin
FUNCTION sadzp060_check_file_if_writable(p_file_path)
DEFINE 
  p_file_path STRING
DEFINE 
  ls_file_path  STRING,
  lb_result     BOOLEAN
DEFINE
  lb_return     BOOLEAN

  LET ls_file_path = p_file_path

  LET lb_result = TRUE

  IF NOT os.Path.writable(ls_file_path) THEN
    LET lb_result = FALSE 
    DISPLAY cs_warning_tag,"File ",ls_file_path," without write permission, mayby generate fail."
  END IF 

  LET lb_return = lb_result

  RETURN lb_return  

END FUNCTION 
#20210517 end