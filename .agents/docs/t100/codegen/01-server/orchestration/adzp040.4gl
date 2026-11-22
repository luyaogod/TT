# Modify              : 2015/11/09 by Hiko    : 1.改善上傳效能:r.c3第四個參數預設為2:只做組合器以後的事:包含編譯和鏈結.
#                                               2.報表元件規格上傳不需要tab gen.
#                                               3.重產是包含tab gen, 所以程式的地方也得加上判斷條件.
#                                               4.執行r.c/r.f改成呼叫cl_cmdrun_openpipe().
#                                               5.由adzp080跳出錯誤訊息即可.
# Modify 170612-00043 : 2017/06/12 by Ernest  : 1.增加讓 adzp080 辨識的 tag  
#        170622-00041   20170622 by Hiko        : 1.為了簡化開發流程,下載規格不重新產生str,而由上傳一起處理.
#        171219-00036   20171219 by 02785       : 上傳程式檢查client與server端4gl行數
#        180517-00048   20180928 by 02632       : 上傳4fd解析之後,要將dzfk009的相關屬性回寫到原本控件有設定的dzci_t內.
#        190125-00026   20190125 by 02785       : 1.T100 v3.0 Auto Completer功能支持:上傳規格時，增加呼叫adzp158
#        200215-00009   20200220 by 10553       : MSSQL Server相容性調整
#        210819-00004#2 20200220 by 02431       : 上傳規格呼叫azzp188
# Modify 230109-00005#1 20230206 by 02160       : 1.新增支援 XR 報表
# Modify 230925-00019#3 20231030 by 02160       : 1.新增支援 SPECRC3


#&define DEBUG

IMPORT os
IMPORT util

SCHEMA ds

#匯入adzp080的常數設定   
&include "../4gl/sadzp000_cnst.inc" 
&include "../4gl/adzp080_cnst.inc"

CONSTANT cs_arguments_tag STRING = "[ARGUMENTS]"

PRIVATE TYPE T_PROGRAM_TYPE RECORD
               temp_ver    VARCHAR(10), 
               temp_type   VARCHAR(10),
               temp_name   VARCHAR(50)
             END RECORD

DEFINE ms_lang    STRING 

DEFINE
  mi_progress     INTEGER,
  ms_progress_all STRING,
  mb_back_job     BOOLEAN,
  mb_result       BOOLEAN 

DEFINE 
  mi_spend_time DATETIME HOUR TO SECOND

&ifndef DEBUG
GLOBALS "../../cfg/top_global.inc"
#Begin: 171219-00036
GLOBALS
   DEFINE gi_client_4gl_lines LIKE type_t.num10  #設計器看到的4gl總行數
END GLOBALS
#End: 171219-00036
&endif

MAIN
  CALL adzp040_initialize()
  CALL adzp040_start()
  CALL adzp040_finalize()
END MAIN

FUNCTION adzp040_initialize()
DEFINE
  ls_erp_path STRING,
  ls_arg_val1 STRING,
  ls_back_job STRING 

  DISPLAY cs_note_tag  #170612-00043
  
  LET ls_arg_val1 = ARG_VAL(2)          #Program Name
  LET ls_back_job = NVL(ARG_VAL(7),"N") #Background runtime

  IF ls_arg_val1 IS NULL THEN
    DISPLAY "\n",
            "Usage : r.r ",ui.Interface.getName()," [ProgramName.ext] [ModuleName] [SpecType] [Spec Version] [Code Version]","\n",
            "        r.r ",ui.Interface.getName()," [GENDATA]","\n",
            "        r.r ",ui.Interface.getName()," [GENXR] [TableName|ModuleName|TableList] [REGEN]","\n", -- 230109-00005#1
            "        r.r ",ui.Interface.getName()," [GENTSD] [ProgramName] [ModuleName] [SpecType] [Spec Version] [Code Version] [Std Program] [Std Spec Version] [Std Code Version]","\n"
    EXIT PROGRAM
  END IF
  
  LET mi_progress = 0
  LET mb_result = TRUE

  CALL FGL_GETENV(cs_erp_path) RETURNING ls_erp_path

  IF ls_back_job = "Y" THEN
    LET mb_back_job = TRUE
  ELSE
    LET mb_back_job = FALSE
  END IF

  &ifndef DEBUG
    #依模組進行系統初始化設定(系統設定)
    CALL cl_tool_init()
    LET g_bgjob = ls_back_job
    #DISPLAY "g_bgjob : ",g_bgjob
    LET ms_lang = g_lang
  &else
    CONNECT TO "local"
    LET ms_lang = cs_default_lang
  &endif

END FUNCTION

FUNCTION adzp040_start()
DEFINE
  ls_prog_name       STRING,
  ls_ext_name        STRING,
  ls_module_name     STRING,
  ls_arg_val1        STRING,
  ls_arg_val2        STRING,
  ls_arg_val3        STRING,
  ls_arg_val4        STRING,
  ls_arg_val5        STRING,
  ls_arg_val6        STRING,
  ls_arg_val7        STRING,
  ls_arg_val8        STRING,
  ls_arg_val9        STRING,
  ls_arg_val10       STRING,
  ls_arg_val11       STRING, #230109-00005#1
  ls_zip_type        STRING,
  ls_upload_doc_type STRING,
  ls_message         STRING,
  lb_result          BOOLEAN,
  ls_sd_version      STRING,
  ls_pr_version      STRING,
  ls_std_prog_name   STRING,
  ls_std_sd_version  STRING,
  ls_std_pr_version  STRING,
  ls_spec_type       STRING,
  ls_exec_type       STRING,  
  ls_DGENV           STRING,  
  ls_lang            STRING, #230109-00005#1   
  ls_identity        STRING,
  ls_err_msg         STRING,
  ls_err_no          STRING,
  ls_type            STRING,
  ls_tiptop          STRING,  
  lo_program_type    T_PROGRAM_TYPE,
  li_time_start      DATETIME YEAR TO SECOND,
  li_time_end        DATETIME YEAR TO SECOND,  
  ls_run_cmd         STRING

#Begin:210819-00004#2
DEFINE ls_command  STRING,
       lb_result9  BOOLEAN,
       ls_err_msg9 STRING
#End:210819-00004#2


  LET lb_result = FALSE
  LET ls_exec_type = ARG_VAL(2)

  IF ls_exec_type = "GENTSD" THEN
    LET ls_arg_val1  = ARG_VAL(3)  #Program Name
    LET ls_arg_val2  = ARG_VAL(4)  #Module Name
    LET ls_arg_val3  = ARG_VAL(5)  #Upload Document Type (SPEC or CODE)
    LET ls_arg_val4  = ARG_VAL(6)  #SDVersionCode/SpecType(GENTSD)
    LET ls_arg_val5  = ARG_VAL(7)  #PRVersionCode/Version(GENTSD)
    LET ls_arg_val6  = ARG_VAL(8)  #Standard Program
    LET ls_arg_val7  = ARG_VAL(9)  #Standard SDVersionCode/SpecType(GENTSD)
    LET ls_arg_val8  = ARG_VAL(10) #Standard PRVersionCode/Version(GENTSD)  
    LET ls_arg_val9  = ARG_VAL(11) #DGENV
    LET ls_arg_val10 = ARG_VAL(12) #Type
    LET ls_arg_val11 = ARG_VAL(13) #XR Lang #230109-00005#1
  ELSE
    LET ls_arg_val1  = ARG_VAL(2)  #Program Name or GENXR
    LET ls_arg_val2  = ARG_VAL(3)  #Module Name or For GENXR [TableName|ModuleName|TableList]
    LET ls_arg_val3  = ARG_VAL(4)  #Upload Document Type (SPEC or CODE) or For XR REGEN
    LET ls_arg_val4  = ARG_VAL(5)  #SDVersionCode/SpecType(GENTSD)
    LET ls_arg_val5  = ARG_VAL(6)  #PRVersionCode/Version(GENTSD)
    LET ls_arg_val6  = ARG_VAL(7)  #Standard Program
    LET ls_arg_val7  = ARG_VAL(8)  #Standard SDVersionCode/SpecType(GENTSD)
    LET ls_arg_val8  = ARG_VAL(9)  #Standard PRVersionCode/Version(GENTSD)  
    LET ls_arg_val9  = ARG_VAL(10) #DGENV
    LET ls_arg_val10 = ARG_VAL(11) #Type
    LET ls_arg_val11 = ARG_VAL(12) #XR Lang #230109-00005#1
  END IF 

  DISPLAY cs_arguments_tag,"Args01[",ls_arg_val1,"] ", 
                           "Args02[",ls_arg_val2,"] ",  
                           "Args03[",ls_arg_val3,"] ",  
                           "Args04[",ls_arg_val4,"] ",  
                           "Args05[",ls_arg_val5,"] ",  
                           "Args06[",ls_arg_val6,"] ",  
                           "Args07[",ls_arg_val7,"] ",  
                           "Args08[",ls_arg_val8,"] ",  
                           "Args09[",ls_arg_val9,"] ",  
                           "Args10[",ls_arg_val10,"] ",
                           "Args11[",ls_arg_val11,"]" #230109-00005#1
  
  LET ls_prog_name      = ls_arg_val1
  LET ls_module_name    = ls_arg_val2
  LET ls_spec_type      = ls_arg_val3
  LET ls_sd_version     = ls_arg_val4
  LET ls_pr_version     = ls_arg_val5
  LET ls_std_prog_name  = NVL(ls_arg_val6,ls_prog_name)
  LET ls_std_sd_version = ls_arg_val7
  LET ls_DGENV          = ls_arg_val9
  LET ls_lang           = ls_arg_val11 #230109-00005#1

  CASE
    WHEN (ls_exec_type.toUpperCase() = "GENXR") -- 230109-00005#1
      CALL sadzp040_util_generate_xr_files(ls_arg_val2,ls_arg_val3) RETURNING lb_result,ls_message -- 230109-00005#1
    WHEN ((ls_exec_type.toUpperCase() = "GENDATA") OR (ls_exec_type.toUpperCase() = "GENTSD"))
      LET ls_zip_type = ls_exec_type.toUpperCase()
      LET ls_upload_doc_type = ""
      CALL adzp040_get_adzp040_progress_all(ls_zip_type,ls_upload_doc_type) RETURNING ms_progress_all
      IF ls_exec_type.toUpperCase() = "GENDATA" THEN
        LET ls_ext_name  = cs_zip_name_tzp
        CALL sadzp040_util_generate_designer_files(ms_lang) RETURNING lb_result,ls_message
      ELSE
        #2014.06.10 加入 Parse 4fd
        CALL sadzp040_util_parse_4fd(ls_prog_name, ls_module_name, ls_sd_version, ls_DGENV,ms_lang) RETURNING lb_result 
        IF NOT lb_result THEN GOTO _ERROR END IF
        CALL sadzp040_util_generate_tsd(ls_prog_name,ls_module_name,ls_sd_version,ls_spec_type,ls_std_prog_name,ls_std_sd_version,ls_DGENV) RETURNING lb_result
        IF NOT lb_result THEN GOTO _ERROR END IF
      END IF
      LABEL _ERROR:
      CALL adzp040_set_adzp040_progress_all(ls_zip_type,ls_upload_doc_type,mi_progress)
  OTHERWISE  
    #取得程式名稱及副檔名
    IF ls_arg_val1.getIndexOf(".",1) > 1 THEN
      LET ls_prog_name = ls_arg_val1.subString(1,ls_arg_val1.getIndexOf(".",1)-1)
      LET ls_ext_name  = ls_arg_val1.subString(ls_arg_val1.getIndexOf(".",1)+1,ls_arg_val1.getLength())
    ELSE
      LET ls_prog_name = ls_arg_val1
      LET ls_ext_name  = cs_zip_name_tzp
    END IF

    LET ls_zip_type = ls_ext_name

    
    #取得模組名稱
    IF ls_arg_val2 IS NULL THEN
      LET ls_module_name = ls_prog_name.subString(1,3)
      LET ls_module_name = ls_module_name.toUpperCase()
    ELSE  
      LET ls_module_name = ls_arg_val2.trim()
    END IF  

    LET ls_upload_doc_type = ls_arg_val3

    #取得程式類型(M,B ...)
    LET ls_type = sadzp060_2_chk_spec_type(ls_prog_name)
    LET ls_type = ls_type.toUpperCase()
    
    #取得Progress的總筆數
    CALL adzp040_get_adzp040_progress_all(ls_zip_type,ls_upload_doc_type) RETURNING ms_progress_all

    #SPEC/SPECGEN/CSPEC/RSPEC 時執行
    IF (
        ls_upload_doc_type.toUpperCase() = cs_doc_type_spec OR 
        ls_upload_doc_type.toUpperCase() = cs_doc_type_specrc3 OR  #230925-00019#3 
        ls_upload_doc_type.toUpperCase() = cs_doc_type_specgen OR 
        ls_upload_doc_type.toUpperCase() = cs_doc_type_cspec OR
        ls_upload_doc_type.toUpperCase() = cs_doc_type_rspec 
       ) THEN
      #Update tsd
      LET lb_result = TRUE
      CALL adzp040_set_progress("")
      LET li_time_start = cl_get_current() 
      CALL adzp040_update_sd_data(ls_upload_doc_type,ls_zip_type,ls_prog_name,ls_module_name,ls_DGENV) RETURNING lb_result
      LET li_time_end = cl_get_current()
      DISPLAY cs_information_tag,"Update SPEC Data START Time : ",li_time_start  
      DISPLAY cs_information_tag,"Update SPEC Data  END Time : ",li_time_end  
      LET mi_spend_time = mi_spend_time + (li_time_end - li_time_start)
      IF NOT lb_result THEN GOTO _complete END IF
    END IF 

    #SPEC/SPECGEN 時執行
    IF (
        ls_upload_doc_type.toUpperCase() = cs_doc_type_spec OR 
        ls_upload_doc_type.toUpperCase() = cs_doc_type_specrc3 OR  #230925-00019#3 
        ls_upload_doc_type.toUpperCase() = cs_doc_type_specgen
       ) THEN

      IF (ls_type = "M") OR (ls_type = "S") OR (ls_type = "F") OR (ls_type = "Q") THEN 

        #20141204 區分類別
        LET ls_tiptop = "tiptop"
        IF ls_type = "M" THEN
          LET ls_tiptop = ""
        END IF

        #編譯 4fd
        LET lb_result = TRUE
        CALL adzp040_set_progress("")
        LET li_time_start = cl_get_current() 
        CALL adzp040_compile_4fd(ls_prog_name,ls_module_name,ls_tiptop) RETURNING lb_result
        LET li_time_end = cl_get_current()
        DISPLAY cs_information_tag,"Compile 4fd START Time : ",li_time_start  
        DISPLAY cs_information_tag,"Compile 4fd  END Time : ",li_time_end  
        LET mi_spend_time = mi_spend_time + (li_time_end - li_time_start)
        IF NOT lb_result THEN GOTO _complete END IF #2015/11/09 by Hiko

        #Begin:170622-00041:不是主程式的畫面呼叫r.f的時候沒有重產自已的多語言,所以要多做.
        IF ls_type<>"M" THEN
           LET ls_run_cmd = "r.r azzp191 ",ls_prog_name," ",ms_lang," '' '' 'Y'"," &"
           DISPLAY "[INFO]Non-M type to regen 42s:",ls_run_cmd
           RUN ls_run_cmd WITHOUT WAITING
        END IF
        #End:170622-00041

        #Parse 4fd
        LET lb_result = TRUE
        CALL adzp040_set_progress("")
        LET li_time_start = cl_get_current() 
        CALL adzp040_parse_4fd(ls_prog_name,ls_module_name,ls_sd_version,ls_DGENV) RETURNING lb_result
        LET li_time_end = cl_get_current()
        DISPLAY cs_information_tag,"Parse 4fd START Time : ",li_time_start  
        DISPLAY cs_information_tag,"Parse 4fd  END Time : ",li_time_end  
        LET mi_spend_time = mi_spend_time + (li_time_end - li_time_start)
        IF NOT lb_result THEN GOTO _complete END IF
        
        LET lb_result = TRUE
        #分析 4fd 和 4gl 資訊, 錯誤的話不管它
        LET li_time_start = cl_get_current() 
        CALL adzp040_analyze_4fd_4gl(ls_prog_name,ls_sd_version,ls_DGENV) RETURNING lb_result
        LET li_time_end = cl_get_current()
        DISPLAY cs_information_tag,"Analyze 4fd START Time : ",li_time_start  
        DISPLAY cs_information_tag,"Analyze 4fd  END Time : ",li_time_end  
        LET mi_spend_time = mi_spend_time + (li_time_end - li_time_start)
        --IF NOT lb_result THEN GOTO _complete END IF

        #Begin: 190125-00026
        #T100 v3.0 Auto Completer功能支持:呼叫adzp158
        IF ls_type<>"Q" THEN
           LET li_time_start = cl_get_current() 
           IF sadzp040_util_set_auto_completer_data(ls_prog_name,ls_module_name,ls_sd_version,ls_dgenv) THEN END IF #todo 
           LET li_time_end = cl_get_current()
           DISPLAY cs_information_tag,"Set auto completer START Time : ",li_time_start  
           DISPLAY cs_information_tag,"Set auto completer END Time : ",li_time_end  
           LET mi_spend_time = mi_spend_time + (li_time_end - li_time_start)
        END IF
        #End: 190125-00026
      END IF   

    END IF 

    #SPEC 時執行
    IF (
         ls_upload_doc_type.toUpperCase() = cs_doc_type_spec OR
         ls_upload_doc_type.toUpperCase() = cs_doc_type_specrc3  #230925-00019#3
       ) THEN

      #Gen tab
      #2015.10.19 改為SPEC呼叫
      LET lb_result = TRUE
      CALL adzp040_set_progress("")
      LET li_time_start = cl_get_current() 
      #Begin:2015/11/09 by Hiko
      IF (ls_arg_val10.toUpperCase() MATCHES "[MSQ]") THEN #只有M/S/Q類的規格才會影響程式,所以才需要特別在規格上傳的時候先檢查規格設定的正確性.
      #IF (ls_arg_val10.toUpperCase() = "G" OR ls_arg_val10.toUpperCase() = "X") THEN
      #  CALL adzp040_generate_tab(ls_prog_name,lo_program_type.*,ls_pr_version,ls_DGENV) RETURNING lb_result
      #ELSE
      #End:2015/11/09 by Hiko
        IF ls_sd_version IS NULL THEN
          CALL sadzp060_2_get_spec_curr_revision(ls_prog_name, ls_arg_val10.toUpperCase(), ls_module_name) RETURNING ls_sd_version,ls_identity,ls_err_msg
        END IF
        CALL adzp040_generate_tab(ls_prog_name,lo_program_type.*,ls_sd_version,ls_DGENV) RETURNING lb_result
      END IF 

      #Begin:210819-00004#2
      LET ls_command = "r.r azzp188 ",ls_prog_name
      CALL cl_cmdrun_openpipe("adzp040_execute_program", ls_command, FALSE) RETURNING lb_result9,ls_err_msg9 #忽略錯誤
      #End:210819-00004#2

      LET li_time_end = cl_get_current()
      DISPLAY cs_information_tag,"Generate Tab START Time : ",li_time_start  
      DISPLAY cs_information_tag,"Generate Tab  END Time : ",li_time_end  
      LET mi_spend_time = mi_spend_time + (li_time_end - li_time_start)
      IF NOT lb_result THEN GOTO _complete END IF
      
    END IF   

    #230925-00019#3 begin
    #SPECRC3 時執行
    IF (ls_upload_doc_type.toUpperCase() = cs_doc_type_specrc3) THEN
      #產生 tgl 及 4gl 及 42s 及 compile link
      LET lb_result = TRUE
      CALL adzp040_set_progress("")
      LET li_time_start = cl_get_current() 
      CALL adzp040_generate_tgl_4gl(ls_prog_name,ls_module_name,ls_pr_version,ls_DGENV,ls_arg_val10.toUpperCase(),ls_upload_doc_type) RETURNING lb_result #230109-00005#1
      LET li_time_end = cl_get_current()
      DISPLAY cs_information_tag,"Generate tgl 4gl START Time : ",li_time_start  
      DISPLAY cs_information_tag,"Generate tgl 4gl  END Time : ",li_time_end  
      LET mi_spend_time = mi_spend_time + (li_time_end - li_time_start)
      IF NOT lb_result THEN GOTO _complete END IF
    END IF
    #230925-00019#3 end   
    
    #CODE/XRP/SPECGEN/GCODE 時執行
    IF (
         ls_upload_doc_type.toUpperCase() = cs_doc_type_xrp OR  #230109-00005#1 
         ls_upload_doc_type.toUpperCase() = cs_doc_type_code OR 
         ls_upload_doc_type.toUpperCase() = cs_doc_type_specgen OR 
         ls_upload_doc_type.toUpperCase() = cs_doc_type_gcode
       ) THEN

       WHILE TRUE
       
        #XRP Update xmd contents
        #230109-00005#1 begin
        IF ls_upload_doc_type.toUpperCase() = cs_doc_type_xrp THEN 
          CALL adzp040_analyze_xmd(ls_prog_name,ls_module_name,ls_pr_version,ls_DGENV,ls_lang) RETURNING lb_result
          IF NOT lb_result THEN EXIT WHILE END IF
        ELSE  
          #Update tap
          LET lb_result = TRUE
          CALL adzp040_set_progress("")
          LET li_time_start = cl_get_current() 
          CALL adzp040_update_tap(ls_upload_doc_type,ls_zip_type,ls_prog_name,ls_module_name,ls_DGENV) RETURNING lb_result
          LET li_time_end = cl_get_current()
          DISPLAY cs_information_tag,"Update tap data START Time : ",li_time_start  
          DISPLAY cs_information_tag,"Update tap data  END Time : ",li_time_end  
          LET mi_spend_time = mi_spend_time + (li_time_end - li_time_start)
          #IF NOT lb_result THEN GOTO _complete END IF
          IF NOT lb_result THEN EXIT WHILE END IF
        END IF      
        #230109-00005#1 end
        
        #Gen tab
        #2014.06.10 改為不呼叫
        #2014.07.09 改回呼叫,且區分為若為G,X類,則給PR
        #           若SD Version為空, 則呼叫sadzp060_2_get_spec_curr_revision取得之
        #2014.09.24 改為不呼叫
        LET lb_result = TRUE
        CALL adzp040_set_progress("")
        IF sadzp060_2_is_regen(ls_prog_name) OR #2015/11/09 by Hiko
           sadzp030_tab_file_template_renewed(ls_prog_name) OR #2015/11/09 by Hiko
           ls_arg_val10.toUpperCase() = "K"  #230109-00005#1
           THEN 
           LET li_time_start = cl_get_current() 
           IF (
               ls_arg_val10.toUpperCase() = "G" OR 
               ls_arg_val10.toUpperCase() = "K" OR  #230109-00005#1
               ls_arg_val10.toUpperCase() = "X"
              ) THEN
             CALL adzp040_generate_tab(ls_prog_name,lo_program_type.*,ls_pr_version,ls_DGENV) RETURNING lb_result
             IF NOT lb_result THEN EXIT WHILE END IF
           ELSE
             IF ls_sd_version IS NULL THEN
               CALL sadzp060_2_get_spec_curr_revision(ls_prog_name, ls_arg_val10.toUpperCase(), ls_module_name) RETURNING ls_sd_version,ls_identity,ls_err_msg
             END IF
             CALL adzp040_generate_tab(ls_prog_name,lo_program_type.*,ls_sd_version,ls_DGENV) RETURNING lb_result
             IF NOT lb_result THEN EXIT WHILE END IF
           END IF 
           LET li_time_end = cl_get_current()
           DISPLAY cs_information_tag,"Generate Tab START Time : ",li_time_start  
           DISPLAY cs_information_tag,"Generate Tab  END Time : ",li_time_end  
           LET mi_spend_time = mi_spend_time + (li_time_end - li_time_start)
           #IF NOT lb_result THEN GOTO _complete END IF
           IF NOT lb_result THEN EXIT WHILE END IF
        END IF
     
        #產生 tgl 及 4gl 及 42s 及 compile link
        LET lb_result = TRUE
        CALL adzp040_set_progress("")
        LET li_time_start = cl_get_current() 
        CALL adzp040_generate_tgl_4gl(ls_prog_name,ls_module_name,ls_pr_version,ls_DGENV,ls_arg_val10.toUpperCase(),ls_upload_doc_type) RETURNING lb_result #230109-00005#1
        LET li_time_end = cl_get_current()
        DISPLAY cs_information_tag,"Generate tgl 4gl START Time : ",li_time_start  
        DISPLAY cs_information_tag,"Generate tgl 4gl  END Time : ",li_time_end  
        LET mi_spend_time = mi_spend_time + (li_time_end - li_time_start)
        IF NOT lb_result THEN EXIT WHILE END IF
        #IF NOT lb_result THEN GOTO _complete END IF

        #GCODE需產生rdd檔
        IF (ls_upload_doc_type.toUpperCase() = cs_doc_type_gcode) THEN
          LET lb_result = TRUE
          CALL adzp040_set_progress("")
          LET li_time_start = cl_get_current() 
          CALL adzp040_generate_rdd(ls_prog_name,ls_module_name,ls_pr_version) RETURNING lb_result      
          LET li_time_end = cl_get_current()
          DISPLAY cs_information_tag,"Generate rdd START Time : ",li_time_start  
          DISPLAY cs_information_tag,"Generate rdd  END Time : ",li_time_end  
          LET mi_spend_time = mi_spend_time + (li_time_end - li_time_start)
          IF NOT lb_result THEN EXIT WHILE END IF
          #IF NOT lb_result THEN GOTO _complete END IF
        END IF  

        #Begin: 171219-00036
        #放在這裡:是因為必須等到r.c3組合完4gl才能檢查，雖然若編譯有錯不會跑這段，但沒差，編譯有錯user就會去看內容了;此處針對沒報錯但4gl內容異常
        DISPLAY cs_message_tag,"Call adzp040_chk_4gl_lines [ ",ls_prog_name,",",ls_module_name," ]"
        CALL adzp040_chk_4gl_lines(ls_prog_name,ls_module_name) RETURNING lb_result
        IF NOT lb_result THEN
           #by pass (目前僅提示訊息，不視為錯誤)
           LET lb_result = TRUE
        END IF
        #End: 171219-00036

        EXIT WHILE
      END WHILE   
    END IF   

    LABEL _complete:

    IF lb_result THEN
      CALL adzp040_set_adzp040_progress_all(ls_zip_type,ls_upload_doc_type,mi_progress)
    END IF
  
  END CASE  
    
  IF NOT lb_result THEN
    IF NVL(ls_message.trim(),cs_null_default) <> cs_null_default THEN 
      DISPLAY cs_error_tag," ",ls_message
    END IF   
    DISPLAY "[Failed]",ui.Interface.getName()
  ELSE
    DISPLAY "[Complete]",ui.Interface.getName()
  END IF  

  #DISPLAY cs_information_tag,"Build program Total Time : ",mi_spend_time  
  
  LET mb_result = lb_result
  
END FUNCTION

FUNCTION adzp040_finalize()

  {  
  CLOSE WINDOW w_adzp041
  }
  
  IF mb_result THEN
    EXIT PROGRAM 
  ELSE
    EXIT PROGRAM -1
  END IF
  
END FUNCTION

{
FUNCTION adzp040_refresh_progressbar(pi_position,ps_message)
DEFINE 
  pi_position INTEGER,
  ps_message  STRING
DEFINE 
  pgr_status  INTEGER,
  txt_Message STRING   

  DISPLAY "***** Time : ",CURRENT YEAR TO SECOND," *****"

  IF pi_position <> 0 THEN 
    LET pgr_status = pi_position
    DISPLAY BY NAME pgr_status
  END IF 

  IF NVL(ps_message,"NULL") <> "NULL" THEN
    LET txt_Message = ps_message
    DISPLAY BY NAME txt_Message
  END IF
  
  CALL ui.Interface.refresh()
  
END FUNCTION
}

FUNCTION adzp040_generate_tab(p_prog_name,p_program_type,p_version,p_dgenv)
DEFINE
  p_prog_name     STRING,
  p_program_type  T_PROGRAM_TYPE,
  p_version       STRING,
  p_dgenv         STRING   
DEFINE
  ls_prog_name     STRING,
  lo_program_type  T_PROGRAM_TYPE,
  ls_version       STRING, 
  ls_dgenv         STRING,   
  ls_message       STRING,
  ls_form_style    STRING,
  ls_err_no        STRING,
  ls_debug_info    STRING,
  lb_result        BOOLEAN

  LET ls_prog_name      = p_prog_name
  LET lo_program_type.* = p_program_type.*
  LET ls_version        = p_version
  LET ls_dgenv          = p_dgenv

  CALL adzp040_show_message_generator()
  
  LET lb_result = TRUE
  LET ls_err_no = "adz-00038"
  
  LET ls_message = "Call sadzp030_tab_gen [ "||ls_prog_name||" ]"
  DISPLAY cs_message_tag,ls_message
  DISPLAY cs_information_tag,ls_message
  CALL adzp040_set_progress(ls_message)
  LET ls_form_style = "" #2014.03.12 改為不傳遞 Form Style
  #CALL adzp040_get_form_style(ls_prog_name) RETURNING ls_form_style
  DISPLAY cs_command_tag,"CALL sadzp030_tab_gen(",ls_prog_name,",",ls_version,",",ls_form_style,",",ls_dgenv,")"
  CALL sadzp030_tab_gen(ls_prog_name,ls_version,ls_form_style,ls_dgenv) RETURNING lb_result
  LET ls_debug_info = "Call sadzp030_tab_gen : Program Name -> ",ls_prog_name,"\n",
                      "                        Form Style   -> ",ls_form_style,"\n",
                      "                        Version      -> ",ls_version,"\n",
                      "                        DGENV        -> ",ls_dgenv
  
  IF NOT lb_result THEN
    IF mb_back_job THEN
      DISPLAY cs_error_tag,"[",ls_err_no,"]",cl_getmsg(ls_err_no, ms_lang),"\n",
               ls_debug_info
    ELSE
      DISPLAY cs_error_tag,"[",ls_err_no,"]",ls_debug_info,"\n"
      #Begin:2015/11/09 by Hiko:由adzp080跳出錯誤訊息即可.
      #INITIALIZE g_errparam TO NULL
      #LET g_errparam.code =  ls_err_no
      #LET g_errparam.extend = ls_message
      #LET g_errparam.popup = TRUE
      #CALL cl_err()
      #End:2015/11/09 by Hiko
    END IF  
  END IF

  RETURN lb_result
  
END FUNCTION

FUNCTION adzp040_analyze_xmd(p_prog_name,p_module,p_version,p_dgenv,p_lang)
DEFINE
  p_prog_name  STRING,
  p_module     STRING,
  p_version    STRING,
  p_dgenv      STRING,
  p_lang       STRING   
DEFINE
  ls_prog_name     STRING,
  ls_module        STRING,
  ls_version       STRING, 
  ls_dgenv         STRING,   
  ls_lang          STRING,   
  ls_message       STRING,
  ls_err_no        STRING,
  ls_debug_info    STRING,
  lb_result        BOOLEAN

  LET ls_prog_name = p_prog_name
  LET ls_module    = p_module
  LET ls_version   = p_version
  LET ls_dgenv     = p_dgenv
  LET ls_lang      = p_lang

  CALL adzp040_show_message_generator()
  
  LET lb_result = TRUE
  
  LET ls_message = "Call sadzp199_05_update_xr [ "||ls_prog_name||" ]"
  DISPLAY cs_message_tag,ls_message
  CALL sadzp199_05_update_xr(ls_prog_name,ls_module,ls_version,ls_dgenv,ls_lang) RETURNING lb_result
  LET ls_debug_info = "Call sadzp199_05_update_xr : Program Name -> ",ls_prog_name,"\n",
                      "                             Module       -> ",ls_module,"\n",
                      "                             Version      -> ",ls_version,"\n",
                      "                             DGENV        -> ",ls_dgenv,"\n",
                      "                             Lang         -> ",ls_lang
  
  IF NOT lb_result THEN
    DISPLAY cs_error_tag,"Call sadzp199_05_update_xr not successful !! ","\n",ls_debug_info
  END IF

  RETURN lb_result
  
END FUNCTION

FUNCTION adzp040_get_form_style(p_prog_name)
DEFINE
  p_prog_name  STRING 
DEFINE
  ls_prog_name  STRING,
  ls_sql        STRING,
  ls_form_style VARCHAR(50),
  ls_return     STRING

  LET ls_prog_name = p_prog_name

  #Begin:200215-00009
  LET ls_sql = "select fq.dzfq001 form_style           ",
               "  from dzfq_t fq                       ",
               " where fq.dzfq004 = '",ls_prog_name,"' ",
               "   and fq.dzfq003 = '1'                "  #'",ls_ver,"'"  #2013.12.03 修改為寫死 "1"
  #End:200215-00009
 
  PREPARE lpre_get_form_style FROM ls_sql
  DECLARE lcur_get_form_style CURSOR FOR lpre_get_form_style

  OPEN lcur_get_form_style
  FETCH lcur_get_form_style INTO ls_form_style
  CLOSE lcur_get_form_style
  
  FREE lpre_get_form_style
  FREE lcur_get_form_style  

  LET ls_return = ls_form_style
  
  RETURN ls_return
  
END FUNCTION

FUNCTION adzp040_update_sd_data(p_upload_doc_type,p_zip_type,p_prog_name,p_module_name,p_DGENV)
DEFINE
  p_upload_doc_type STRING,
  p_zip_type        STRING,
  p_prog_name       STRING,
  p_module_name     STRING,
  p_DGENV           STRING   
DEFINE
  ls_upload_doc_type STRING,
  ls_zip_type        STRING,
  ls_prog_name       STRING,
  ls_module_name     STRING, 
  ls_DGENV           STRING,   
  ls_message         STRING,
  ls_err_message     STRING,
  ls_err_no          STRING,
  lb_result          BOOLEAN  

  LET ls_upload_doc_type = p_upload_doc_type 
  LET ls_zip_type        = p_zip_type     
  LET ls_prog_name       = p_prog_name
  LET ls_module_name     = p_module_name
  LET ls_DGENV           = p_DGENV

  CALL adzp040_show_message_designer()
  
  LET lb_result = TRUE

  #SPEC, SPECGEN 時處理
  IF (ls_upload_doc_type IS NOT NULL) AND 
     (
       ls_upload_doc_type.toUpperCase() = cs_doc_type_spec OR 
       ls_upload_doc_type.toUpperCase() = cs_doc_type_specrc3 OR  #230925-00019#3 
       ls_upload_doc_type.toUpperCase() = cs_doc_type_specgen
     ) THEN
    LET ls_err_no      = "adz-00040"
    LET ls_message     = "Update tsd (SPEC,SPECGEN) : "||ls_prog_name
    LET ls_err_message = "Update tsd error !"
    DISPLAY cs_message_tag,ls_message
    DISPLAY cs_information_tag,ls_message
    CALL adzp040_set_progress(ls_message)
    #呼叫產出 tsd
    DISPLAY cs_command_tag,"CALL sadzp060_1_update_tsd(",ls_prog_name,",",ls_module_name,",",ls_DGENV,")"
    CALL sadzp060_1_update_tsd(ls_prog_name,ls_module_name,ls_DGENV) RETURNING lb_result
  END IF
  
  #RSPEC 時處理
  IF (ls_upload_doc_type IS NOT NULL) AND (ls_upload_doc_type.toUpperCase() = cs_doc_type_rspec) THEN
    LET ls_err_no      = "adz-00040"
    LET ls_message     = "Update rsd (RSPEC) : "||ls_prog_name
    LET ls_err_message = "Update rsd error !"
    DISPLAY cs_message_tag,ls_message
    DISPLAY cs_information_tag,ls_message
    CALL adzp040_set_progress(ls_message)
    #呼叫產出 rsd
    DISPLAY cs_command_tag,"CALL sadzp060_1_update_rsd(",ls_prog_name,",",ls_module_name,",",ls_DGENV,")"
    CALL sadzp060_1_update_rsd(ls_prog_name,ls_module_name,ls_DGENV) RETURNING lb_result
  END IF
  
  #CSPEC 時處理
  IF (ls_upload_doc_type IS NOT NULL) AND (ls_upload_doc_type.toUpperCase() = cs_doc_type_cspec) THEN
    LET ls_err_no      = "adz-00040"
    LET ls_message     = "Update tsd (CSPEC) : "||ls_prog_name
    LET ls_err_message = "Update tsd error !"
    DISPLAY cs_message_tag,ls_message
    DISPLAY cs_information_tag,ls_message
    CALL adzp040_set_progress(ls_message)
    #呼叫產出 tsd
    DISPLAY cs_command_tag,"CALL sadzp040_2_update_csd(",ls_prog_name,",",ls_module_name,",",ls_DGENV,")"
    CALL sadzp040_2_update_csd(ls_prog_name,ls_module_name,ls_DGENV) RETURNING lb_result
  END IF
  
  IF NOT lb_result THEN
    IF mb_back_job THEN
      DISPLAY cs_error_tag,"[",ls_err_no,"]",cl_getmsg(ls_err_no, ms_lang)
    ELSE
      DISPLAY cs_error_tag,"[",ls_err_no,"]",cl_getmsg(ls_err_no, ms_lang)
      #Begin:2015/11/09 by Hiko:由adzp080跳出錯誤訊息即可.
      #INITIALIZE g_errparam TO NULL
      #LET g_errparam.code =  ls_err_no
      #LET g_errparam.extend = ls_err_message
      #LET g_errparam.popup = TRUE
      #CALL cl_err()
      #End:2015/11/09 by Hiko
    END IF  
  END IF

  RETURN lb_result
  
END FUNCTION

FUNCTION adzp040_update_tap(p_upload_doc_type,p_zip_type,p_prog_name,p_module_name,p_DGENV)
DEFINE
  p_upload_doc_type STRING,
  p_zip_type        STRING,
  p_prog_name       STRING,
  p_module_name     STRING, 
  p_DGENV           STRING 
DEFINE
  ls_upload_doc_type STRING,
  ls_zip_type        STRING,
  ls_prog_name       STRING,
  ls_module_name     STRING,
  ls_DGENV           STRING, 
  ls_message         STRING,
  ls_err_message     STRING,
  ls_err_no          STRING,
  lb_result          BOOLEAN  

  LET ls_upload_doc_type = p_upload_doc_type 
  LET ls_zip_type        = p_zip_type     
  LET ls_prog_name       = p_prog_name
  LET ls_module_name     = p_module_name
  LET ls_DGENV           = p_DGENV

  CALL adzp040_show_message_designer()
  
  LET lb_result = TRUE

  IF (ls_upload_doc_type IS NOT NULL) AND 
     (
      ls_upload_doc_type.toUpperCase() = cs_doc_type_code OR
      ls_upload_doc_type.toUpperCase() = cs_doc_type_gcode
     ) THEN
    LET ls_err_no      = "adz-00039"
    LET ls_message     = "Update tap : "||ls_prog_name
    LET ls_err_message = "Update tap error !"
    DISPLAY cs_message_tag,ls_message
    DISPLAY cs_information_tag,ls_message
    CALL adzp040_set_progress(ls_message)
    
    #呼叫更新 tap
    DISPLAY cs_command_tag,"CALL sadzp060_1_update_tap(",ls_prog_name,",",ls_module_name,",",ls_DGENV,")"
    CALL sadzp060_1_update_tap(ls_prog_name,ls_module_name,ls_DGENV) RETURNING lb_result
  ELSE 
    LET ls_message = "Not 'CODE' type package, skip update TAP data."
    DISPLAY cs_information_tag,ls_message
  END IF

  IF NOT lb_result THEN
    IF mb_back_job THEN
      DISPLAY cs_error_tag,"[",ls_err_no,"]",cl_getmsg(ls_err_no, ms_lang)
    ELSE
      DISPLAY cs_error_tag,"[",ls_err_no,"]",cl_getmsg(ls_err_no, ms_lang)
      #Begin:2015/11/09 by Hiko:由adzp080跳出錯誤訊息即可.
      #INITIALIZE g_errparam TO NULL
      #LET g_errparam.code =  ls_err_no
      #LET g_errparam.extend = ls_err_message
      #LET g_errparam.popup = TRUE
      #CALL cl_err()
      #End:2015/11/09 by Hiko
    END IF  
  END IF

  RETURN lb_result
  
END FUNCTION

FUNCTION adzp040_generate_tgl_4gl(p_prog_name, p_module_name, p_version,p_dgenv,p_rep_type,p_doc_type)
DEFINE
  p_prog_name    STRING,
  p_module_name  STRING,
  p_version      STRING,
  p_dgenv        STRING,
  p_rep_type     STRING,  #230109-00005#1
  p_doc_type     STRING   #230925-00019#3
DEFINE
  ls_prog_name     STRING,
  ls_module_name   STRING,
  ls_version       STRING, 
  ls_dgenv         STRING,
  ls_rep_type      STRING,  #230109-00005#1
  ls_doc_type      STRING,  #230925-00019#3
  ls_module_env    STRING,
  ls_exec_str      STRING,
  ls_sub_root      STRING,
  li_chdir_result  STRING,
  lb_result        BOOLEAN,
  ls_message       STRING,
  ls_err_message   STRING, 
  ls_pid           STRING, 
  ls_err_file_name STRING,
  ls_temp_dir      STRING,
  ls_err_no        STRING,
  ls_parameter     STRING

  LET ls_prog_name   = p_prog_name
  LET ls_module_name = p_module_name.toLowerCase()
  LET ls_version     = p_version
  LET ls_dgenv       = p_dgenv
  LET ls_rep_type    = p_rep_type #230109-00005#1
  LET ls_doc_type    = p_doc_type.toUpperCase() #230925-00019#3
  
  CALL adzp040_show_message_generator()
  
  LET ls_err_no = "adz-00128"
  
  LET ls_sub_root   = "4gl"
  LET ls_module_env = os.Path.join(FGL_GETENV(p_module_name.toUpperCase()),ls_sub_root)

  CALL os.Path.chdir(ls_module_env) RETURNING li_chdir_result
  
  LET ls_message  = "Call r.c3 generate tgl and 4gl"
  DISPLAY cs_message_tag,ls_message
  DISPLAY cs_information_tag,ls_message
  CALL adzp040_set_progress(ls_message)
  #第四個參數為 0 表示全部做完編譯以下的事情.
  #第四個參數為 1 表示做完產生與組合,但不做編譯和鏈結.
  #第四個參數為 2 表示只做組合器以後的事:包含編譯和鏈結.
  #第四個參數為 3 表示只做組合器以後的事:不做編譯和鏈結. #2015/11/09 by Hiko
  #2014.07.03 第四個參數從 2 改為 1
  #2014.09.24 第四個參數改回 2
  #2014.12.05 第四個參數改回 0

  #230925-00019#3 begin
  IF ls_doc_type = cs_doc_type_specrc3 THEN
    LET ls_parameter = ls_prog_name," '' A"
  ELSE
    #Begin:2015/11/09 by Hiko
    #LET ls_parameter = ls_prog_name," ","''"," ",ls_version," ","0"," ",ls_dgenv
    IF sadzp060_2_is_regen(ls_prog_name) OR 
       sadzp030_tab_file_template_renewed(ls_prog_name) OR
       ls_rep_type = "K" THEN #230109-00005#1
       LET ls_parameter = ls_prog_name," ","''"," ",ls_version," ","0"," ",ls_dgenv
    ELSE
       LET ls_parameter = ls_prog_name," ","''"," ",ls_version," ","2"," ",ls_dgenv
    END IF
    #Begin:2015/11/09 by Hiko
  END IF
  #230925-00019#3 end
    
  DISPLAY cs_command_tag,"r.c3 ",ls_module_name," ",ls_parameter
  CALL adzp040_execute_program("r.c3 ",ls_module_name,ls_parameter) RETURNING lb_result,ls_err_message
  
  IF NOT lb_result THEN
    IF mb_back_job THEN
      DISPLAY cs_error_tag,"[",ls_err_no,"]",cl_getmsg(ls_err_no, ms_lang)
    ELSE
      DISPLAY cs_error_tag,"[",ls_err_no,"]",cl_getmsg(ls_err_no, ms_lang)
      #Begin:2015/11/09 by Hiko:由adzp080跳出錯誤訊息即可.
      #INITIALIZE g_errparam TO NULL
      #LET g_errparam.code =  ls_err_no
      #LET g_errparam.extend = ls_message
      #LET g_errparam.popup = TRUE
      #CALL cl_err()
      #End:2015/11/09 by Hiko
    END IF  
  END IF

  RETURN lb_result    
  
END FUNCTION

FUNCTION adzp040_generate_rdd(p_prog_name, p_module_name, p_version)
DEFINE
  p_prog_name    STRING,
  p_module_name  STRING,
  p_version      STRING
DEFINE
  ls_prog_name     STRING,
  ls_module_name   STRING,
  ls_version       STRING, 
  ls_module_env    STRING,
  ls_exec_str      STRING,
  ls_sub_root      STRING,
  li_chdir_result  STRING,
  lb_result        BOOLEAN,
  ls_message       STRING,
  ls_err_message   STRING, 
  ls_pid           STRING, 
  ls_err_file_name STRING,
  ls_temp_dir      STRING,
  ls_err_no        STRING,
  ls_parameter     STRING

  LET ls_prog_name   = p_prog_name
  LET ls_module_name = p_module_name.toLowerCase()
  LET ls_version     = p_version

  CALL adzp040_show_message_generator()
  
  LET ls_err_no = "adz-00128"
  
  LET ls_sub_root   = "4gl"
  LET ls_module_env = os.Path.join(FGL_GETENV(p_module_name.toUpperCase()),ls_sub_root)

  CALL os.Path.chdir(ls_module_env) RETURNING li_chdir_result
  
  LET ls_message  = "Call r.c generate rdd"
  DISPLAY cs_message_tag,ls_message
  DISPLAY cs_information_tag,ls_message
  CALL adzp040_set_progress(ls_message)

  LET ls_parameter = ls_prog_name," ","rdd"
  DISPLAY cs_command_tag,"r.c ",ls_module_name," ",ls_parameter
  CALL adzp040_execute_program("r.c ",ls_module_name,ls_parameter) RETURNING lb_result,ls_err_message
  
  IF NOT lb_result THEN
    IF mb_back_job THEN
      DISPLAY cs_error_tag,"[",ls_err_no,"]",cl_getmsg(ls_err_no, ms_lang)
    ELSE
      DISPLAY cs_error_tag,"[",ls_err_no,"]",cl_getmsg(ls_err_no, ms_lang)
      #Begin:2015/11/09 by Hiko:由adzp080跳出錯誤訊息即可.
      #INITIALIZE g_errparam TO NULL
      #LET g_errparam.code =  ls_err_no
      #LET g_errparam.extend = ls_message
      #LET g_errparam.popup = TRUE
      #CALL cl_err()
      #End:2015/11/09 by Hiko
    END IF  
  END IF

  RETURN lb_result    
  
END FUNCTION

FUNCTION adzp040_compile_4fd(p_prog_name,p_module_name,p_tiptop)
DEFINE
  p_prog_name    STRING,
  p_module_name  STRING,
  p_tiptop       STRING
DEFINE
  ls_prog_name    STRING,
  ls_module_name  STRING,
  ls_tiptop       STRING,
  ls_module_env   STRING,
  ls_exec_str     STRING,
  ls_sub_root     STRING,
  li_chdir_result STRING,
  lb_result       BOOLEAN,
  ls_parameter    STRING,
  ls_message      STRING,
  ls_err_message  STRING, 
  ls_err_no       STRING

  LET ls_prog_name   = p_prog_name
  LET ls_module_name = p_module_name.toLowerCase()
  LET ls_tiptop      = p_tiptop

  CALL adzp040_show_message_user_control()
  
  LET ls_err_no = "adz-00034"
  
  LET ls_sub_root   = "4fd"
  LET ls_module_env = os.Path.join(FGL_GETENV(p_module_name.toUpperCase()),ls_sub_root)

  CALL os.Path.chdir(ls_module_env) RETURNING li_chdir_result
  
  LET ls_message = "Compile form (4fd) "||ls_prog_name
  DISPLAY cs_message_tag,ls_message
  DISPLAY cs_information_tag,ls_message
  CALL adzp040_set_progress(ls_message)

  LET ls_parameter = ls_prog_name," ",ls_tiptop
  DISPLAY cs_command_tag,"r.f ",ls_parameter
  CALL adzp040_execute_program("r.f ",ls_module_name,ls_parameter) RETURNING lb_result,ls_err_message

  IF NOT lb_result THEN
    IF mb_back_job THEN
      DISPLAY cs_error_tag,"[",ls_err_no,"]",cl_getmsg(ls_err_no, ms_lang)
    ELSE
      DISPLAY cs_error_tag,"[",ls_err_no,"]",cl_getmsg(ls_err_no, ms_lang)
      #Begin:2015/11/09 by Hiko:由adzp080跳出錯誤訊息即可.
      #INITIALIZE g_errparam TO NULL
      #LET g_errparam.code =  ls_err_no
      #LET g_errparam.extend = ls_message
      #LET g_errparam.popup = TRUE
      #CALL cl_err()
      #End:2015/11/09 by Hiko
    END IF  
  END IF

  RETURN lb_result    
  
END FUNCTION

FUNCTION adzp040_parse_4fd(p_prog_name, p_module_name, p_version, p_dgenv)
DEFINE
  p_prog_name    STRING,
  p_module_name  STRING,
  p_version      STRING,
  p_dgenv        STRING
DEFINE
  ls_prog_name    LIKE dzfk_t.dzfk001, #180517-00048
  ls_module_name  STRING,
  ls_version      STRING,
  ls_dgenv        STRING,
  ls_module_env   STRING,
  ls_exec_str     STRING,
  ls_sub_root     STRING,
  li_chdir_result STRING,
  lb_result       BOOLEAN,
  ls_return       STRING,
  ls_message      STRING,
  ls_err_no       STRING
#Begin:180517-00048
DEFINE ls_upd_sql  STRING,
       ls_sql      STRING,
       la_dzfk     DYNAMIC ARRAY OF RECORD
                   dzfk003 LIKE dzfk_t.dzfk003,
                   dzfk009 LIKE dzfk_t.dzfk009
                   END RECORD,
       li_cnt,li_i INTEGER,
       ls_info     STRING,
       obj_dzfk009 util.JSONObject,
       l_dzci005   LIKE dzci_t.dzci005,
       l_dzci006   LIKE dzci_t.dzci006,
       l_dzci007   LIKE dzci_t.dzci007,
       l_dzci008   LIKE dzci_t.dzci008,
       l_dzci009   LIKE dzci_t.dzci009,
       l_date      DATETIME YEAR TO SECOND
#End:180517-00048

  LET ls_prog_name    = p_prog_name
  LET ls_module_name  = p_module_name.toLowerCase()
  LET ls_version      = p_version.Trim()
  LET ls_dgenv        = p_dgenv.trim()
  
  CALL adzp040_show_message_designer()
  
  LET ls_err_no = "adz-00137"
  
  LET ls_sub_root   = "4fd"
  LET ls_module_env = os.Path.join(FGL_GETENV(p_module_name.toUpperCase()),ls_sub_root)

  CALL os.Path.chdir(ls_module_env) RETURNING li_chdir_result
  
  LET ls_message = "Parse form (4fd) "||ls_prog_name
  DISPLAY cs_message_tag,ls_message
  DISPLAY cs_information_tag,ls_message
  CALL adzp040_set_progress(ls_message)

  LET lb_result = TRUE
  TRY
    #DISPLAY "sadzp168_3 version : ",ls_version 
    DISPLAY cs_command_tag,"CALL sadzp168_3(",ls_module_name,",",ls_prog_name,",",ls_version,",",ls_dgenv,")"
    CALL sadzp168_3(ls_module_name, ls_prog_name,ls_version,ls_dgenv) RETURNING lb_result,ls_return
    #Begin:180517-00048
    IF lb_result THEN
       IF ls_dgenv="c" OR g_user="topstd" THEN #客製環境下才支持azzi250(dzci_t)的設定.
          SELECT COUNT(1) INTO li_cnt FROM gzde_t
           WHERE gzde001 = 'sadzp252_1' AND gzdestus = 'Y'
          IF li_cnt>0 THEN #防呆.
             DISPLAY "[INFO]adzp040_parse_4fd:CALL sadzp252_1_parse_property(...):update dzci_t start..."
             LET l_date = cl_get_current()
             
             #因為是INNER JOIN找出來的dzfk003,所以只會UPDATE.
             #Begin:200215-00009
             LET ls_upd_sql = "update dzci_t set dzci005 = ?,",
                              "                  dzci006 = ?,",
                              "                  dzci007 = ?,",
                              "                  dzci008 = ?,",
                              "                  dzci009 = ?,",
                              "                  dzcimodid = ?,",
                              "                  dzcimoddt = ?",
                              " where dzci001 = ?",
                              "   and dzci003 = ?"
             #End:200215-00009
             PREPARE dzci_upd_prep FROM ls_upd_sql
             
             #Begin:200215-00009
             LET ls_sql = "select dzfk003, dzfk009",
                          "  from dzfk_t",
                          " inner join dzci_t on dzci001 = dzfk001 and dzci003 = dzfk003",
                          " where dzfk001 = '",ls_prog_name,"'",
                          "   and dzfk002 = ",ls_version,
                          "   and dzfk007 = '",ls_dgenv,"'",
                          " order by dzfk003"
             #End:200215-00009
             PREPARE dzfk_prep FROM ls_sql
             DECLARE dzfk_curs CURSOR FOR dzfk_prep
             LET li_i = 1
             LOCATE la_dzfk[li_i].dzfk009 IN FILE
             FOREACH dzfk_curs INTO la_dzfk[li_i].*
                LET ls_info = "[INFO]adzp040_parse_4fd:CALL sadzp252_1_parse_property(...) failure:program_id=",ls_prog_name,",widget_id=",la_dzfk[li_i].dzfk003
                CALL sadzp252_1_parse_property(la_dzfk[li_i].dzfk009) RETURNING obj_dzfk009
                LET l_dzci005 = obj_dzfk009.get("format")
                LET l_dzci006 = obj_dzfk009.get("case")
                LET l_dzci007 = obj_dzfk009.get("required")
                LET l_dzci008 = obj_dzfk009.get("hidden")
                LET l_dzci009 = obj_dzfk009.get("noEntry")
                LET ls_info = "[INFO]adzp040_parse_4fd:Update dzci_t failure:program_id=",ls_prog_name,",widget_id=",la_dzfk[li_i].dzfk003
                EXECUTE dzci_upd_prep USING l_dzci005,
                                            l_dzci006,
                                            l_dzci007,
                                            l_dzci008,
                                            l_dzci009,
                                            g_user,
                                            l_date,
                                            ls_prog_name,
                                            la_dzfk[li_i].dzfk003
                LET ls_info = ""
                FREE la_dzfk[li_i].dzfk009
                LET li_i = li_i + 1
                LOCATE la_dzfk[li_i].dzfk009 IN FILE
             END FOREACH
             CALL la_dzfk.deleteElement(li_i)
             FREE la_dzfk[li_i].dzfk009
             DISPLAY "[INFO]adzp040_parse_4fd:CALL sadzp252_1_parse_property(...):update dzci_t finish."
          END IF
       END IF
    END IF
    #End:180517-00048
  CATCH
    DISPLAY ls_info
    LET lb_result = FALSE
  END TRY  
  
  IF NOT lb_result THEN
    IF mb_back_job THEN
      DISPLAY cs_error_tag,"[",ls_err_no,"]",cl_getmsg(ls_err_no, ms_lang)
    ELSE
      DISPLAY cs_error_tag,"[",ls_err_no,"]",cl_getmsg(ls_err_no, ms_lang)
      #Begin:2015/11/09 by Hiko:由adzp080跳出錯誤訊息即可.
      #INITIALIZE g_errparam TO NULL
      #LET g_errparam.code =  ls_err_no
      #LET g_errparam.extend = ls_message
      #LET g_errparam.popup = TRUE
      #CALL cl_err()
      #End:2015/11/09 by Hiko
    END IF  
  END IF

  RETURN lb_result    
  
END FUNCTION

FUNCTION adzp040_analyze_4fd_4gl(p_prog_name,p_version,p_dgenv)
DEFINE
  p_prog_name    STRING,
  p_version      STRING,
  p_dgenv        STRING
DEFINE
  ls_prog_name    STRING,
  ls_version      STRING,
  ls_dgenv        STRING,
  ls_module_env   STRING,
  ls_exec_str     STRING,
  ls_sub_root     STRING,
  li_chdir_result STRING,
  lb_result       BOOLEAN,
  ls_return       STRING,
  ls_message      STRING,
  ls_err_no       STRING

  LET ls_prog_name    = p_prog_name
  LET ls_version      = p_version.Trim()
  LET ls_dgenv        = p_dgenv.Trim()

  CALL adzp040_show_message_designer()

  LET ls_err_no = "adz-00301"
  
  LET ls_message = "Analyze 4fd and 4gl information "||ls_prog_name
  DISPLAY cs_message_tag,ls_message
  DISPLAY cs_information_tag,ls_message
  CALL adzp040_set_progress(ls_message)

  LET lb_result = TRUE
  TRY
    DISPLAY cs_command_tag,"CALL sadzp168_4(",ls_prog_name,",",ls_version,",",ls_dgenv,")"
    DISPLAY cs_ignore_tag
    CALL sadzp168_4(ls_prog_name,ls_version,ls_dgenv) RETURNING lb_result,ls_return
    DISPLAY cs_note_tag
  CATCH
    LET lb_result = FALSE
  END TRY  
  LET lb_result = TRUE -- 錯誤都忽略
  
  IF NOT lb_result THEN
    IF mb_back_job THEN
      DISPLAY cs_error_tag,"[",ls_err_no,"]",cl_getmsg(ls_err_no, ms_lang)
    ELSE
      DISPLAY cs_error_tag,"[",ls_err_no,"]",cl_getmsg(ls_err_no, ms_lang)
      --CALL cl_err_msg(ls_message, ls_err_no, NULL, 0)
    END IF  
  END IF

  RETURN lb_result    
  
END FUNCTION

PUBLIC FUNCTION adzp040_execute_program(p_command,p_module_name,p_other_parm)
DEFINE 
  p_command      STRING, #要執行的指令
  p_module_name  STRING, #模組名稱
  p_other_parm   STRING  #其他參數
DEFINE  
  ls_command      STRING, #要執行的指令
  ls_module_name  STRING,
  ls_other_parm   STRING, #其他參數
  ls_path         STRING, #要編譯的路徑
  ls_prog_line    STRING,
  ls_msg          STRING, #編譯後訊息
  lb_result       BOOLEAN,#成功或失敗
  li_chdir_result INTEGER, 
  lch_read        base.Channel

  LET ls_command     = p_command
  LET ls_module_name = p_module_name   
  LET ls_other_parm  = p_other_parm
  
  LET lb_result = TRUE
  LET lch_read  = base.Channel.create()
  CALL lch_read.setDelimiter("")
  LET ls_msg    = ls_command,"\n" 
   
  #切換目錄
  LET ls_path = FGL_GETENV(ls_module_name.toUpperCase())

  #若是 r.f 就切換到 4fd 目錄下
  IF ls_command.getIndexOf("r.f",1) > 0 THEN
    LET ls_path = os.Path.join(ls_path, "4fd")
  ELSE  
    IF ls_module_name IS NOT NULL THEN
      LET ls_path = os.Path.join(ls_path, "4gl")
    ELSE
      LET ls_path = ""  
    END IF  
  END IF

  IF ls_path IS NOT NULL THEN
    CALL os.Path.chdir(ls_path) RETURNING li_chdir_result
    IF NOT li_chdir_result THEN
      LET lb_result = FALSE
      DISPLAY "Change directory to """,ls_path,""" fail."
    ELSE
      DISPLAY "Change directory to """,ls_path,""" success."  
    END IF
  END IF  

  IF lb_result THEN
    #Begin:2015/11/09 by Hiko
    ##執行指令
    #LET ls_command = ls_command,ls_other_parm," 2>&1"
    
    #DISPLAY cs_command_tag,ls_command
    #
    #CALL lch_read.openPipe(ls_command, "r")
    #WHILE TRUE
    #  LET ls_prog_line = lch_read.readLine()
    #  IF lch_read.isEof() THEN 
    #    EXIT WHILE 
    #  END IF
    #  
    #  #DISPLAY cs_debug_tag,ls_prog_line
    #  LET ls_msg = ls_msg, 
    #               ls_prog_line, "\n"
    #  LET ls_prog_line = ls_prog_line.toUpperCase()
    #   
    #  #遇到錯誤訊息, 回傳 False
    #  IF ls_prog_line.getIndexOf("ERROR", 1) > 0 THEN
    #    LET lb_result = FALSE
    #  END IF
    #END WHILE
    #CALL lch_read.close()
    
    #DISPLAY ls_msg
    
    LET ls_command = ls_command,ls_other_parm
    DISPLAY cs_command_tag,ls_command
    CALL cl_cmdrun_openpipe("adzp040_execute_program", ls_command, FALSE) RETURNING lb_result,ls_msg
    #End:2015/11/09 by Hiko
  END IF
  
  RETURN lb_result,ls_msg
  
END FUNCTION

FUNCTION adzp040_get_program_type(p_program_name)
DEFINE
  p_program_name STRING
DEFINE
  ls_sql           STRING,
  ls_program_name  STRING,
  lo_program_type  T_PROGRAM_TYPE
DEFINE
  lo_return  T_PROGRAM_TYPE

  LET ls_program_name = p_program_name

  #取得程式樣板型態
  #Begin:200215-00009
  LET ls_sql = "select fq.dzfq003 temp_ver,               ",
               "       fq.dzfq005 temp_type,              ",
               "       fq.dzfq001 temp_name               ",
               "  from dzfq_t fq                          ",
               " where fq.dzfq004  = '",ls_program_name,"'",
               "   and fq.dzfq003  = '1'                  ",
               "   and fq.dzfqstus = 'Y'                  "
  #End:200215-00009
               
  PREPARE lpre_get_program_type FROM ls_sql
  DECLARE lcur_get_program_type CURSOR FOR lpre_get_program_type
  OPEN lcur_get_program_type
  FETCH lcur_get_program_type INTO lo_program_type.*
  CLOSE lcur_get_program_type
  FREE lcur_get_program_type
  FREE lpre_get_program_type

  LET lo_return.* = lo_program_type.*
  
  RETURN lo_return.*
  
END FUNCTION

#以亂數產出副檔名
FUNCTION adzp040_gen_random_name()
DEFINE
  lr_random_name RECORD
    segment1 STRING,
    segment2 STRING,
    segment3 STRING,
    segment4 STRING
  END RECORD
DEFINE 
  li_random_value   INTEGER,
  li_max_random_num INTEGER,
  ls_final_name     STRING,
  ls_using_format   STRING
DEFINE  
  ls_return STRING

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

FUNCTION adzp040_write_file(p_file_name,p_line_string)
DEFINE
  p_file_name   STRING,
  p_line_string STRING
DEFINE   
  ls_file_name      STRING,
  ls_line_string    STRING,
  lo_channel_write  base.Channel

  LET ls_file_name   = p_file_name
  LET ls_line_string = p_line_string

  LET  lo_channel_write = base.Channel.create()
  CALL lo_channel_write.setDelimiter("")

  CALL lo_channel_write.openFile(ls_file_name, "w" )
  CALL lo_channel_write.write(ls_line_string)
  CALL lo_channel_write.close()
   
END FUNCTION

FUNCTION adzp040_if_first_version(p_program)
DEFINE
  p_program STRING
DEFINE
  lr_first_version RECORD
                     VER_COUNT INTEGER,
                     MIN_VER   VARCHAR(20)
                   END RECORD
DEFINE
  ls_program STRING,
  ls_sql     STRING
DEFINE
  lb_return BOOLEAN   

  LET ls_program = p_program.toLowerCase()

  #Begin:200215-00009
  LET ls_sql = "select count(distinct ba.dzba002) ver_count,min(ba.dzba002) min_ver ",
               "  from dzba_t ba                                                    ",
               " where ba.dzba001 = '",ls_program,"'                                "
  #End:200215-00009
               
  PREPARE lpre_if_first_version FROM ls_sql
  DECLARE lcur_if_first_version CURSOR FOR lpre_if_first_version

  OPEN lcur_if_first_version
  FETCH lcur_if_first_version INTO lr_first_version.*
  CLOSE lcur_if_first_version
  
  FREE lpre_if_first_version
  FREE lcur_if_first_version  

  DISPLAY "VER_COUNT : ",lr_first_version.VER_COUNT
  DISPLAY "MIN_VER : ",lr_first_version.MIN_VER

  IF (lr_first_version.VER_COUNT = 0) OR (lr_first_version.VER_COUNT >= 2) OR (lr_first_version.VER_COUNT = 1 AND lr_first_version.MIN_VER <> "1") THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF   
  
  RETURN lb_return
  
END FUNCTION

FUNCTION adzp040_get_adzp040_progress_all(p_pgrs_type,p_doc_type)
DEFINE
  p_pgrs_type STRING,
  p_doc_type  STRING
DEFINE
  ls_pgrs_type    STRING,
  ls_doc_type     STRING,
  ls_progress_all VARCHAR(5),
  ls_sql          STRING
DEFINE
  ls_return STRING  

  LET ls_pgrs_type = p_pgrs_type
  LET ls_doc_type  = p_doc_type.toUpperCase()
  
  #Begin:200215-00009
  LET ls_sql = "select ej.dzej006 progress_all                         ",
               "  from dzej_t ej                                       ",
               " where ej.dzej001 = 'adzp040_Progress'                 ",  
               "   and ej.dzej003 = 'ALL'                              ",
               "   and ej.dzej004 = '",ls_pgrs_type,"'                 ",
               "   and nvl(ej.dzej005,'X') = nvl('",ls_doc_type,"','X')"
  #End:200215-00009
 
  PREPARE lpre_get_adzp040_progress_all FROM ls_sql
  DECLARE lcur_get_adzp040_progress_all CURSOR FOR lpre_get_adzp040_progress_all

  OPEN lcur_get_adzp040_progress_all
  FETCH lcur_get_adzp040_progress_all INTO ls_progress_all
  CLOSE lcur_get_adzp040_progress_all
  
  FREE lpre_get_adzp040_progress_all
  FREE lcur_get_adzp040_progress_all  

  LET ls_return = ls_progress_all
  
  RETURN ls_return
  
END FUNCTION

FUNCTION adzp040_set_adzp040_progress_all(p_pgrs_type,p_doc_type,p_progress)
DEFINE
  p_pgrs_type STRING,
  p_doc_type  STRING,
  p_progress  STRING
DEFINE
  ls_pgrs_type STRING,
  ls_doc_type  STRING,
  ls_progress  STRING,
  ls_sql       STRING

  LET ls_pgrs_type = p_pgrs_type
  LET ls_doc_type  = p_doc_type.toUpperCase()
  LET ls_progress  = p_progress

  #Begin:200215-00009
  &ifdef is_msv
  LET ls_sql = "update dzej_t                                       ", 
               "   set dzej006 = '",ls_progress,"'                  ",
               " where dzej001 = 'adzp040_Progress'                 ",  
               "   and dzej003 = 'ALL'                              ",
               "   and dzej004 = '",ls_pgrs_type,"'                 ",
               "   and nvl(dzej005,'X') = nvl('",ls_doc_type,"','X')"
  &endif

  #PGS-10553 add -S-
  &ifdef is_pgs
  LET ls_sql = "update dzej_t ej                                                 ", 
               "   set    dzej006 = '",ls_progress,"'                            ",
               " where ej.dzej001 = 'adzp040_Progress'                           ",  
               "   and ej.dzej003 = 'ALL'                                        ",
               "   and ej.dzej004 = '",ls_pgrs_type,"'                           ",
               "   and coalesce(ej.dzej005,'X') = coalesce('",ls_doc_type,"','X')"
  &endif
  #PGS-10553 add -E-
  
  &ifdef is_ora
  LET ls_sql = "update dzej_t ej                                       ", 
               "   set ej.dzej006 = '",ls_progress,"'                  ",
               " where ej.dzej001 = 'adzp040_Progress'                 ",  
               "   and ej.dzej003 = 'ALL'                              ",
               "   and ej.dzej004 = '",ls_pgrs_type,"'                 ",
               "   and nvl(ej.dzej005,'X') = nvl('",ls_doc_type,"','X')"
  &endif
  #End:200215-00009

  #DISPLAY "Update SQL : ",ls_sql 
  BEGIN WORK

  TRY
    PREPARE lpre_set_adzp040_progress_all FROM ls_sql
    EXECUTE lpre_set_adzp040_progress_all
    COMMIT WORK
  CATCH
    DISPLAY "[ERROR] Update adzp040 progress. "
    ROLLBACK WORK
  END TRY  
  
END FUNCTION

FUNCTION adzp040_set_progress(p_message)
DEFINE
  p_message STRING
DEFINE
  ls_message      STRING,
  ls_progress     STRING,
  ls_env_progress STRING

  IF mb_back_job THEN RETURN END IF

  LET ls_message = p_message
  
  LET mi_progress = mi_progress + 1
  LET ls_progress = mi_progress
  
  DISPLAY cs_progress_tag,ls_progress
  
  IF NVL(ls_message,"X") <> "X" THEN
    DISPLAY cs_message_tag,ls_message
  END IF
  
END FUNCTION

FUNCTION adzp040_show_message_designer()
  DISPLAY cs_message_half_separator,cs_message_designer,cs_message_half_separator
END FUNCTION

FUNCTION adzp040_show_message_generator()
  DISPLAY cs_message_half_separator,cs_message_generator,cs_message_half_separator
END FUNCTION

FUNCTION adzp040_show_message_user_control()
  DISPLAY cs_message_half_separator,cs_message_user_control,cs_message_half_separator
END FUNCTION

FUNCTION adzp040_show_message_separator()
  DISPLAY cs_message_separator
END FUNCTION

#Begin: 171219-00036
#+ 檢核離線檔4gl行數與Sever端4gl行數是否一致
#  True: 正常   False:異常
FUNCTION adzp040_chk_4gl_lines(ps_prog_name,ps_module_name)
  DEFINE 
    ps_prog_name        STRING,
    ps_module_name      STRING,
    ls_file             STRING,
    li_server_4gl_lines LIKE type_t.num10,
    li_client_4gl_lines LIKE type_t.num10,
    li_result           LIKE type_t.num10, #筆數差
    ls_sql              STRING,
    li_cnt              LIKE type_t.num5,
    lc_gzoj001          LIKE gzoj_t.gzoj001,
    lb_result           BOOLEAN #檢查結果 
    
    
   
    LET lb_result = TRUE
    LET li_cnt = 0 

    TRY
       #檢查是否有設定azzi051(行業別插入標準程式),若有則暫不檢查,直到設計器支持能顯示這些azzi051資料才會準,否則會誤判
       LET lc_gzoj001 = ps_prog_name
       LET ls_sql = "select count(1) from gzoj_t where gzoj001=?"    #200215-00009
       PREPARE pre_get_gzoj_t FROM ls_sql
       EXECUTE pre_get_gzoj_t USING lc_gzoj001 INTO li_cnt
       IF li_cnt = 0 THEN
          LET li_client_4gl_lines = gi_client_4gl_lines
          IF li_client_4gl_lines > 0 THEN #向下相容:設計器舊版不會有這個屬性(0)
             LET ls_file = os.path.JOIN(os.path.JOIN(FGL_GETENV(ps_module_name),"4gl"),ps_prog_name||'.4gl')
             LET li_server_4gl_lines = sadzp007_util_get_file_lines(ls_file)
             LET li_result = li_server_4gl_lines - li_client_4gl_lines
             IF li_server_4gl_lines > 0 AND sadzp007_util_math_abs(li_result) > 10 THEN
                DISPLAY cs_warning_tag,"server_4gl_lines=",li_server_4gl_lines USING "<<<<<<<<<<", " ,client_4gl_lines=",li_client_4gl_lines USING "<<<<<<<<<<"
                LET lb_result = FALSE
                INITIALIZE g_errparam TO NULL
                LET g_errparam.code =  "adz-01297"
                LET g_errparam.popup = TRUE
                LET g_errparam.replace[1] = li_server_4gl_lines USING "<<<<<<<<<<"
                LET g_errparam.replace[2] = li_client_4gl_lines USING "<<<<<<<<<<"
                LET g_errparam.type = 0
                CALL cl_err()
             END IF
          END IF
       ELSE
          DISPLAY cs_message_tag,"Call adzp040_chk_4gl_lines [ ",ps_prog_name,",",ps_module_name," ]"," :azzi051",cl_getmsg("azz-01135",g_lang)," (by pass)"
       END IF

    CATCH
       DISPLAY "SQLERRMESSAGE=",SQLERRMESSAGE  
       DISPLAY "SQLCA.SQLCODE=",cl_getmsg(SQLCA.SQLCODE,g_lang)
       RETURN lb_result

    END TRY

    RETURN lb_result
 
END FUNCTION
#End: 171219-00036

