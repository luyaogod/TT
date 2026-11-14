{
  異動型態    異動單號       日 期       異動者      異動內容
  ========= ============= ========== ========== ================================
  Suspend                 20160919   Ernestliou 1.匯入資料比對功能暫停
  modify                  20170405   Ernestliou 1.取消行業別匯入比對
  modify                  20170419   Ernestliou 1.匯出 gztd_t 時, 只匯出該表格有使用的部份
                                                2.匯入時, 先匯入 gztd_t 再匯入其他表格
  Modify                  20170523   Ernestliou 1.當 DGENV 為 c 時, 不刪除任何客制欄位
  Modify                  20170524   Ernestliou 1.取消上述修改
  Modify                  20170912   Ernestliou 1.索引匯入時判斷有無已存在 index 欄位, 有則匯入方設失效
                                                2.匯入前判斷是否有比匯入檔較多的欄位, 有則警告
  Modify                  20171120   02160      1. 新增欄位挑選式匯入功能
                                                2. 若新的型態與舊的型態驗證不通過, 則判斷為 error
  Modify                  20180129   02160      1. 匯入時判斷, 若匯出檔的 ERPVER 和目標區的不一樣則報錯
  Modify                  20180213   02160      1. 針對 GST 3.0 dynamic array Compile 失敗的問題進行修正
  Modify                  20180302   02160      1. 挑選式匯入功能可以匯入已經刪除的標準欄位, 並在匯入時給予提示將匯入已經刪除的欄位, 並將已經刪除的紀錄移除
  Modify                  20180307   02160      1. 增加Patch Mode, 以讓 adzi140_patch 順利執行
  Modify                  20180410   02160      1. 鼎新內部開發環境支持全新程式支持跨版本匯入, 包含表格
  Modify                  20180420   02160      1. 讓'全新開發程式'可以在v1.1與v3.0雙向匯出入
  Modify                  20181224   02160      1.修正比對欄位型態或長度不同時所給予的背景訊息
  Modify                  20190108   02160      1.修正ERPVER 1.0 及 3.0 的卡控規則
  Modify                  20190304   02160      1.標準表可以在行業環境增加專屬欄位
  Modify                  20191205   02160      1.在客戶端, 無論是r.t或是後台新增的標準欄位, 一律不去刪除, 只做新增與修改
  ------------------------------------------------------------------------------
  Modify                  20191209   02160      1.修改支援 MSSQLServer(1.1.071)   
  Modify                  20200330   02160      1.修改支援 PosgreSQL(1.1.072)   
  Modify                  20210705   02160      1.修改 tar 解壓縮時加上 m 參數以修改原檔案時間日期
  Modify    230825-00006  20230825   02160      1.修正行業環境匯入時無判斷 INDSD 導致誤刪行業欄位的問題
}  

&include "../4gl/sadzi140_mcr.inc" 

IMPORT os
IMPORT util
IMPORT security

SCHEMA ds

&include "../4gl/sadzp000_cnst.inc"
&include "../4gl/sadzi140_cnst.inc"
&include "../4gl/sadzp200_cnst.inc"

&include "../4gl/sadzi140_type.inc"
&include "../4gl/sadzp200_type.inc"

PUBLIC TYPE T_COL_DATA_LENGTH VARCHAR(500) #20171120

&ifndef DEBUG
GLOBALS "../../cfg/top_global.inc"
&endif

CONSTANT cs_UsingFormat STRING = "&&&&&"
CONSTANT cs_NoVersionCharacter STRING = "99999"

#執行的環境變數
DEFINE 
  ms_dlang             STRING,
  ms_user              STRING, #20180410
  mb_fault             BOOLEAN,
  mo_local_sys_info    T_PKG_SYS_INFO,
  mo_src_pkg_sys_info  T_PKG_SYS_INFO,
  ms_entry_path        STRING,
  ms_MasterDB          STRING,
  ms_MasterUser        STRING,
  mb_backend_mode      BOOLEAN,
  mb_make_work_dir     BOOLEAN,
  ms_max_sequence      STRING,
  mb_same_dgenv        BOOLEAN, 
  mb_new_table         BOOLEAN,  
  mb_check_erpid       BOOLEAN,
  mb_same_db           BOOLEAN,
  mb_same_topind       BOOLEAN,
  mb_renamed           BOOLEAN,
  ms_construct_version STRING,
  ms_topind            STRING, #20190304
  ms_DGENV             STRING, #20170523
  mb_dgenv_is_cust     BOOLEAN, #20170523
  mb_pickup_exist      BOOLEAN, #20171120 
  ms_deleted_std_col_list  STRING, #20180302
  mo_deleted_std_col_list  DYNAMIC ARRAY OF T_COLUMN_LIST, #20180302
  mo_pickup_json_arr   util.JSONArray, #20171120
  mo_pick_col_list     DYNAMIC ARRAY OF T_COLUMN_LIST, #20171120
  mo_PARAMETERS        T_DZLM_T_SCR,
  mo_export_info       T_EXPORT_INFO,
  mo_exec_result       T_EXEC_RESULT,
  mo_json_obj          util.JSONObject,
  mo_column_info       DYNAMIC ARRAY OF T_COLUMN_INFO,
  mo_local_lang        DYNAMIC ARRAY OF T_LOCAL_LANGUAGE
            
FUNCTION sadzi140_exim_run(p_backend_mode,p_parameters)
DEFINE
  p_backend_mode  BOOLEAN,
  p_parameters    T_DZLM_T_SCR

  CALL sadzi140_exim_initialize(p_backend_mode,p_parameters.*)
  CALL sadzi140_exim_start()
  CALL sadzi140_exim_finalize()

  RETURN mo_export_info.*,mo_exec_result.*
  
END FUNCTION 

FUNCTION sadzi140_exim_initialize(p_backend_mode,p_parameters)
DEFINE
  p_backend_mode     BOOLEAN,
  p_parameters       T_DZLM_T_SCR
DEFINE
  lo_parameters  T_DZLM_T_SCR,
  ls_check_erpid STRING
DEFINE
  lb_return BOOLEAN  

  LET mb_backend_mode = p_backend_mode
  LET lo_parameters.* = p_parameters.*

  &ifndef DEBUG
    LET ms_dlang = NVL(g_dlang,cs_default_lang)
    LET ms_user = g_account --g_user -- #20180410
  &else
    LET ms_dlang = cs_default_lang
    LET ms_user  = FGL_GETENV("USER") #20180410
    --CALL FGL_SETENV("DGENV","s")
    --CALL FGL_SETENV("ERPVER","3.0")
    --CALL FGL_SETENV("RT_DEBUG_MODE","N")
    --CALL FGL_SETENV("RT_PATCH_MODE","N")
    --CALL FGL_SETENV("CUST","DIGIWIN")
  &endif

  #20170523 begin
  LET ms_DGENV  = NVL(FGL_GETENV(cs_keyword_DGENV),cs_dgenv_standard)
  LET ms_topind = NVL(FGL_GETENV(cs_keyword_TOPIND),cs_default_topind) #20190304
  IF ms_DGENV = cs_dgenv_customize THEN
    LET mb_dgenv_is_cust = TRUE
  ELSE 
    LET mb_dgenv_is_cust = FALSE
  END IF
  #20170523 end

  CALL sadzi140_exim_get_sys_info() RETURNING mo_local_sys_info.*

  LET mo_PARAMETERS.TYPES   = lo_parameters.TYPES   #iexp, iimp, idel, iasm
  LET mo_PARAMETERS.DZLM002 = lo_parameters.DZLM002 #TableName
  LET mo_PARAMETERS.DZLM005 = lo_parameters.DZLM005 #ConstructVersion
  LET mo_PARAMETERS.DZLM006 = lo_parameters.DZLM006 #SD_Version
  LET mo_PARAMETERS.DZLM012 = lo_parameters.DZLM012 #需求單號
  LET mo_PARAMETERS.DZLM015 = lo_parameters.DZLM015 #序號

  #若ConstructVersion設為"T", 表示要匯出 r.t 的設計資料, 而非歷史版本
  IF (mo_PARAMETERS.DZLM005 = cs_rt_design_data_code) OR 
     (mo_PARAMETERS.DZLM005 = cs_rt_design_data_number) THEN #20180420
    LET mo_PARAMETERS.DZLM005 = cs_NoVersionCharacter
    LET mo_PARAMETERS.DZLM012 = cs_rt_design_request_no  #需求單號
    LET mo_PARAMETERS.DZLM015 = cs_rt_design_sequence_no #序號
    LET mb_same_dgenv = TRUE
  ELSE   
    LET mb_same_dgenv = FALSE
  END IF 
  
  #呼叫亂數產生器
  CALL util.Math.srand()

  #取得Master DB 及 User
  CALL sadzi140_db_get_parameter(cs_param_level_system,cs_param_master_db) RETURNING ms_MasterDB
  CALL sadzi140_db_get_parameter(cs_param_level_system,cs_param_master_user) RETURNING ms_MasterUser

  #是否檢核 ERPID
  CALL sadzi140_db_get_parameter(cs_param_level_export_import,cs_param_check_erpid) RETURNING ls_check_erpid
  LET mb_check_erpid = IIF(NVL(ls_check_erpid,"Y") == "Y",TRUE,FALSE) 
  
  #取得當地多語言清單 #20160824
  CALL sadzi140_util_get_local_lang() RETURNING mo_local_lang
  
  #建立取得Table export notice物件
  LET mo_json_obj = util.JSONObject.Create()  
  LET mb_pickup_exist = FALSE #20171120

  LET mb_fault = FALSE

END FUNCTION

FUNCTION sadzi140_exim_start()
DEFINE
  lo_DZLM_T           DYNAMIC ARRAY OF T_DZLM_T_SCR,
  ls_TableName        STRING,
  ls_TARName          STRING,
  ls_ConstructVersion STRING,
  li_ConstructVersion INTEGER,
  li_SD_Version       INTEGER,
  ls_SD_Version       STRING
DEFINE  
  ls_TYPES             STRING,
  li_loop              INTEGER,
  lb_fault             BOOLEAN,
  lb_success           BOOLEAN,
  li_CHDIR             INTEGER,
  ls_rtn_code          STRING,
  lb_Copyfile          BOOLEAN,
  lb_DropSuccess       BOOLEAN,
  lb_DelSuccess        BOOLEAN,
  lb_TableExists       BOOLEAN,
  lb_RebuildSuccess    BOOLEAN,
  lb_ASMSuccess        BOOLEAN,
  lb_diff              BOOLEAN,
  lb_identified        BOOLEAN,
  ls_Revision          STRING,
  ls_TARString         VARCHAR(1024),
  ls_prev_working_path STRING,
  lo_VersionInfo       T_VERSION_INFO,
  lo_revision          T_REVISION,
  lo_curr_DZAF_T       T_DZAF_T,
  lo_extract_DATA      DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH, #20171120
  lo_left_array        DYNAMIC ARRAY OF T_COLUMN_LIST, #20171120
  lo_right_array       DYNAMIC ARRAY OF T_COLUMN_LIST, #20171120
  lo_pick_col_arr      DYNAMIC ARRAY OF T_COLUMN_LIST, #20171120
  ls_columns           STRING, #20171120 
  ls_result            STRING, #20171120 
  ls_json_name         STRING, #20171120
  ls_alm_request_no    STRING,
  ls_alm_request_seq   STRING,
  ls_dgenv             STRING,
  ls_message           STRING,
  ls_tar_name          STRING,
  ls_error_code        STRING,
  ls_err_message       STRING,
  ls_GUID              STRING, #20160519
  lb_result            BOOLEAN 

  &ifndef DEBUG
  IF NVL(mo_PARAMETERS.TYPES,cs_null_value) = cs_null_value THEN
    DISPLAY cs_error_tag,"No any parameters.","\n"
  END IF
  &endif

  LET lb_fault = FALSE
  
  LET ls_TYPES            = mo_PARAMETERS.TYPES
  LET ls_TableName        = mo_PARAMETERS.DZLM002
  LET ls_ConstructVersion = mo_PARAMETERS.DZLM005
  LET ls_SD_Version       = mo_PARAMETERS.DZLM006
  LET ls_alm_request_no   = mo_PARAMETERS.DZLM012
  LET ls_alm_request_seq  = mo_PARAMETERS.DZLM015
  
  LET ls_dgenv = mo_local_sys_info.psi_DGENV

  #取得及切換目前工作目錄
  CALL os.Path.pwd() RETURNING ms_entry_path

  #驗證資料庫資料是否完整(export 不檢查)
  #若驗證資料庫資料失敗, 則回報錯誤直接離開
  IF (ls_TYPES.ToLowerCase() <> cs_exim_export) THEN 
    CALL sadzi140_util_identify_db_data(mb_backend_mode,ms_dlang) RETURNING lb_identified
    IF NOT lb_identified THEN
      LET lb_fault = NOT lb_identified
      GOTO _exit_start
    END IF
  END IF    
 
  CASE NVL(ls_TYPES.ToLowerCase(),cs_null_value)
    #Export r.t setting data  
    WHEN cs_exim_export  
      LET lb_success = TRUE

      #給定匯出註記中的系統資訊
      CALL sadzi140_exim_set_export_notice_sys_info(mo_local_sys_info.*,mo_json_obj,FALSE)
      #20180410 begin
      CALL sadzi140_exim_get_current_construct_version(ls_TableName,ls_DGENV) RETURNING lo_curr_DZAF_T.*
      CALL sadzi140_exim_set_verify_construct_version(mo_json_obj,lo_curr_DZAF_T.DZAF002)
      #20180410 end
      
      #取得DZLM的表格名稱  
      IF ls_TableName IS NOT NULL THEN
        IF NOT mb_make_work_dir THEN 
          CALL sadzi140_exim_making_work_directory(ls_TableName,cs_working_dir_type_export) RETURNING lb_success,mo_export_info.WORKING_PATH
          IF NOT lb_success THEN GOTO _exp_error END IF 
        END IF            
        CALL sadzi140_exim_export_DZEA_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF 
        CALL sadzi140_exim_export_DZEAL_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF 
        CALL sadzi140_exim_export_DZEB_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF 
        CALL sadzi140_exim_export_DZEBL_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF 
        CALL sadzi140_exim_export_DZEC_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF 
        CALL sadzi140_exim_export_DZED_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF 
        CALL sadzi140_exim_export_DZEN_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF 
        -- adzi150 begin        
        CALL sadzi140_exim_export_DZEF_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF 
        CALL sadzi140_exim_export_DZEP_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF 
        CALL sadzi140_exim_export_DZEQ_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF 
        CALL sadzi140_exim_export_DZER_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF 
        CALL sadzi140_exim_export_DZET_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF 
        -- adzi150 end        
        #CALL sadzi140_exim_export_DZEO_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        #IF NOT lb_success THEN GOTO _exp_error END IF
        CALL sadzi140_exim_export_DZEU_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF 
        #CALL sadzi140_exim_export_DZEV_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        #IF NOT lb_success THEN GOTO _exp_error END IF 
        #CALL sadzi140_exim_export_DZEW_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        #IF NOT lb_success THEN GOTO _exp_error END IF 
        CALL sadzi140_exim_export_GZTD_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF 
        CALL sadzi140_exim_export_GZTDL_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF
        CALL sadzi140_exim_export_DZIG_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF 
        -- adzi150        
        CALL sadzi140_exim_export_GZCC_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF 

        #匯出表格匯出註記
        CALL sadzi140_exim_set_table_export_notice(ls_TableName,util.JSON.format(mo_json_obj.toString())) RETURNING lb_success
        IF NOT lb_success THEN GOTO _exp_error END IF
        #TAR File
        CALL sadzi140_exim_tar_file(mo_PARAMETERS.*) RETURNING lb_success,ls_tar_name
        IF NOT lb_success THEN GOTO _exp_error END IF
        
        LABEL _exp_error:
        
        #返回原路徑  
        CALL os.Path.chdir(ms_entry_path) RETURNING li_CHDIR
      END IF
      LET lb_fault = NOT lb_success

    #20160519 begin  
    #Verify import data
    WHEN cs_exim_verify

      #刪除過驗證資料(保留3天)
      CALL sadzi140_exim_delete_expired_verify_data() RETURNING lb_result
      DISPLAY "Delete expired verify data : ",IIF(lb_result,cs_result_success,cs_result_fail)
      
      LET lb_success = TRUE
      CALL sadzi140_db_get_column_with_sequence(ls_TableName) RETURNING mo_column_info 
      CALL sadzi140_exim_check_if_new_table(mo_column_info) RETURNING mb_new_table
      
      BEGIN WORK
      IF ls_TableName IS NOT NULL THEN
        IF NOT mb_make_work_dir THEN 
          CALL sadzi140_exim_making_work_directory(ls_TableName,cs_working_dir_type_import) RETURNING lb_success,mo_export_info.WORKING_PATH
          IF NOT lb_success THEN GOTO _vfy_error END IF 
        END IF   

        CALL sadzi140_exim_untar_file(mo_PARAMETERS.*) RETURNING lb_success
        DISPLAY "Untar : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _vfy_error END IF

        #取得檢核包資訊 
        CALL sadzi140_exim_get_export_pkg_sys_info(ls_TableName) RETURNING mo_json_obj,mo_src_pkg_sys_info.*

        #20170405 begin
        {
        #檢核是否符合行業包匯入原則
        CALL sadzi140_exim_check_industry_if_effective(mo_local_sys_info.*,mo_src_pkg_sys_info.*) RETURNING lb_success,ls_error_code
        DISPLAY "Check industry package import rule : ",IIF(lb_success,cs_result_success,cs_result_fail)
        LET ls_err_message = ls_error_code,"-",sadzp000_msg_get_message(ls_error_code,ms_dlang)
        IF NOT lb_success THEN DISPLAY cs_error_tag,ls_err_message GOTO _vfy_error END IF
        }
        #20170405 end
        
        #檢核檢核包必要條件
        CALL sadzi140_exim_table_check_notice(ls_TableName,mo_json_obj,mo_local_sys_info.*,mo_src_pkg_sys_info.*) RETURNING lb_success
        DISPLAY "Check export notice : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _vfy_error END IF 

        #抓取判斷匯出檔是不是匯出端和匯入端相同
        IF mo_src_pkg_sys_info.psi_DGENV = mo_local_sys_info.psi_DGENV THEN LET mb_same_dgenv = TRUE ELSE LET mb_same_dgenv = FALSE END IF
        #抓取判斷來源資料庫類型是不是匯出端和匯入端相同
        IF mo_src_pkg_sys_info.psi_DB_TYPE = mo_local_sys_info.psi_DB_TYPE THEN LET mb_same_db = TRUE ELSE LET mb_same_db = FALSE END IF
        #抓取判斷來源行業別是不是匯出端和匯入端相同
        IF mo_src_pkg_sys_info.psi_TOPIND = mo_local_sys_info.psi_TOPIND THEN LET mb_same_topind = TRUE ELSE LET mb_same_topind = FALSE END IF
        #抓取判斷來源包是不是經由更名式匯入的包
        IF mo_src_pkg_sys_info.psi_RENAMED = "Y" THEN LET mb_renamed = TRUE ELSE LET mb_renamed = FALSE END IF
        
        #取得比對 GUID
        LET ls_GUID = security.RandomGenerator.CreateUUIDString()
        DISPLAY cs_information_tag,"Verify GUID : ",ls_GUID

        {
        #20160919 Suspend begin
        CALL sadzi140_exim_fetch_dzea_t_to_dzha_t(ls_GUID,ls_TableName) RETURNING lb_success 
        DISPLAY "Fetch data from DZEA_T to DZHA_T : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _vfy_error END IF
        CALL sadzi140_exim_fetch_dzeb_t_to_dzhb_t(ls_GUID,ls_TableName) RETURNING lb_success  
        DISPLAY "Fetch data from DZEB_T to DZHB_T : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _vfy_error END IF

        COMMIT WORK 
        BEGIN WORK
        #20160919 Suspend end
        }

        CALL sadzi140_exim_import_DZHA_data(ls_GUID,ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        DISPLAY "Import DZHA : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _vfy_error END IF 
        --CALL sadzi140_exim_import_DZHAL_data(ls_GUID,ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        --DISPLAY "Import DZHAL : ",IIF(lb_success,cs_result_success,cs_result_fail)
        --IF NOT lb_success THEN GOTO _vfy_error END IF
        CALL sadzi140_exim_import_DZHB_data(ls_GUID,ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        DISPLAY "Import DZHB : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _vfy_error END IF 
        --CALL sadzi140_exim_import_DZHBL_data(ls_GUID,ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        --DISPLAY "Import DZHBL : ",IIF(lb_success,cs_result_success,cs_result_fail)
        --IF NOT lb_success THEN GOTO _vfy_error END IF 
        CALL sadzi140_exim_import_DZHC_data(ls_GUID,ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        DISPLAY "Import DZHC : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _vfy_error END IF 
        CALL sadzi140_exim_import_DZHD_data(ls_GUID,ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        DISPLAY "Import DZHD : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _vfy_error END IF
        CALL sadzi140_exim_import_DZHN_data(ls_GUID,ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        DISPLAY "Import DZHN : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _vfy_error END IF
                
        LABEL _vfy_error:
        
        IF NOT lb_success THEN
          TRY  
            ROLLBACK WORK
          CATCH 
          END TRY  
        ELSE
          TRY  
            COMMIT WORK
          CATCH 
          END TRY  
        END IF

        #匯入成功再進行比對
        IF lb_success THEN
          LET mo_exec_result.er_GUID = ls_GUID #20170912 
          CALL sadzi140_exim_verify_import_data_with_design_data(ls_GUID,ls_TableName) RETURNING lb_result
          LET lb_success = lb_result
        END IF  
        
        LABEL _exit_vfy:
        
        #返回原路徑  
        CALL os.Path.chdir(ms_entry_path) RETURNING li_CHDIR
        
      END IF          
      
      LET lb_fault = lb_success
      
    #20160519 end

    #20171120 begin  
    #Selection extract and pack data
    WHEN cs_exim_sel_imp

      LET lb_success = TRUE
      CALL sadzi140_db_get_column_with_sequence(ls_TableName) RETURNING mo_column_info 
      CALL sadzi140_exim_check_if_new_table(mo_column_info) RETURNING mb_new_table
      
      IF ls_TableName IS NOT NULL THEN
        IF NOT mb_make_work_dir THEN 
          CALL sadzi140_exim_making_work_directory(ls_TableName,cs_working_dir_type_import) RETURNING lb_success,mo_export_info.WORKING_PATH
          IF NOT lb_success THEN GOTO _simp_error END IF 
        END IF   

        CALL sadzi140_exim_untar_file(mo_PARAMETERS.*) RETURNING lb_success
        DISPLAY "Untar : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _simp_error END IF

        #取得包資訊 
        CALL sadzi140_exim_get_export_pkg_sys_info(ls_TableName) RETURNING mo_json_obj,mo_src_pkg_sys_info.*
        CALL sadzi140_exim_extract_DZEB_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success,lo_extract_DATA
        #過濾如果已經存在的欄位且型態相同就不讓挑選
        CALL sadzi140_exim_refresh_extract_data(lo_extract_DATA) RETURNING lo_extract_DATA
        #拆分及呼叫欄位選擇
        LET lo_left_array  = NULL
        LET lo_right_array = NULL
        CALL sadzi140_exim_get_col_array_from_extract_data(lo_extract_DATA) RETURNING lo_left_array   
        CALL sadzi140_cols_run(lo_left_array,lo_right_array) RETURNING ls_result,lo_pick_col_arr 
        IF NVL(ls_result,cs_null_value) = cs_result_cancel OR lo_pick_col_arr.getLength() = 0 THEN
          LET lb_success = FALSE
        ELSE
          CALL sadzi140_exim_export_to_pickup_column_json(ls_TableName,lo_pick_col_arr) RETURNING lb_result,ls_json_name
          IF lb_result THEN
            #設定Pickup標記
            CALL sadzi140_exim_set_export_notice_pick_up_info(mo_json_obj)
            #重新匯出表格匯出註記
            CALL sadzi140_exim_set_table_export_notice(ls_TableName,util.JSON.format(mo_json_obj.toString())) RETURNING lb_success
            IF NOT lb_success THEN GOTO _exp_error END IF
            #TAR File
            CALL sadzi140_exim_tar_file(mo_PARAMETERS.*) RETURNING lb_success,ls_tar_name
            IF NOT lb_success THEN GOTO _exp_error END IF
          END IF
        END IF

        LABEL _simp_error:
        
        #返回原路徑  
        CALL os.Path.chdir(ms_entry_path) RETURNING li_CHDIR
        
      END IF          
      
      LET lb_fault = lb_success
      
    #20171120 end
    
    #Import r.t setting data  
    WHEN cs_exim_import
      LET lb_success = TRUE
      CALL sadzi140_db_get_column_with_sequence(ls_TableName) RETURNING mo_column_info 
      CALL sadzi140_exim_check_if_new_table(mo_column_info) RETURNING mb_new_table

      IF ls_TableName IS NOT NULL THEN
        IF NOT mb_make_work_dir THEN 
          CALL sadzi140_exim_making_work_directory(ls_TableName,cs_working_dir_type_import) RETURNING lb_success,mo_export_info.WORKING_PATH
          IF NOT lb_success THEN GOTO _imp_error END IF 
        END IF   
        CALL sadzi140_vcs_get_new_seqence(ls_TableName,FALSE,FALSE,FALSE,TRUE,4) RETURNING ms_max_sequence

        #20180410 mark begin
        {
        #取得現今DZAF的資料
        LET lo_curr_DZAF_T.DZAF001 = ls_TableName
        LET lo_curr_DZAF_T.DZAF005 = cs_spec_type_table
        LET lo_curr_DZAF_T.DZAF010 = ls_DGENV
        CALL sadzp200_ver_get_curr_ver_info(lo_curr_DZAF_T.*) RETURNING lo_curr_DZAF_T.*
        }
        #20180410 mark end

        CALL sadzi140_exim_get_current_construct_version(ls_TableName,ls_DGENV) RETURNING lo_curr_DZAF_T.* #20180410
        LET mo_local_sys_info.psi_CONSTRUCT_VER = lo_curr_DZAF_T.DZAF002 #20180410
        IF lo_curr_DZAF_T.DZAF002 IS NULL THEN
          BEGIN WORK 
          CALL sadzp200_ver_get_new_ver_info(lo_curr_DZAF_T.*,cs_user_role_sd,FALSE) RETURNING lb_success,lo_curr_DZAF_T.*
          DISPLAY "Get New Version Number : ",IIF(lb_success,cs_result_success,cs_result_fail)
          IF NOT lb_success THEN ROLLBACK WORK GOTO _exit_imp ELSE COMMIT WORK END IF
        END IF
        LET ms_construct_version = lo_curr_DZAF_T.DZAF002  
        
        CALL sadzi140_exim_untar_file(mo_PARAMETERS.*) RETURNING lb_success
        DISPLAY "Untar : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF

        #取得匯入包資訊 
        CALL sadzi140_exim_get_export_pkg_sys_info(ls_TableName) RETURNING mo_json_obj,mo_src_pkg_sys_info.*
        #取得挑選欄位清單
        CALL sadzi140_exim_check_and_get_pickup_cols_json(ls_TableName,mo_json_obj) RETURNING mb_pickup_exist,mo_pickup_json_arr #20171120
        IF mb_pickup_exist THEN 
          CALL sadzi140_exim_transfer_json_arr_to_FGL(mo_pickup_json_arr) RETURNING mo_pick_col_list
          CALL sadzi140_util_recover_standard_column_info(ls_TableName,mo_pick_col_list) RETURNING lb_result #20180302 #回收在已刪除標準欄位清單中的資料
        END IF  

        #20180302 begin
        #取得已經刪除的標準欄位
        CALL sadzi140_util_get_deleted_std_column_list(ls_TableName) RETURNING mo_deleted_std_col_list,ms_deleted_std_col_list 
        #20180302 end
        
        #20170405 begin
        {
        #檢核是否符合行業包匯入原則
        CALL sadzi140_exim_check_industry_if_effective(mo_local_sys_info.*,mo_src_pkg_sys_info.*) RETURNING lb_success,ls_error_code
        DISPLAY "Check industry package import rule : ",IIF(lb_success,cs_result_success,cs_result_fail)
        LET ls_err_message = ls_error_code,"-",sadzp000_msg_get_message(ls_error_code,ms_dlang)
        IF NOT lb_success THEN DISPLAY cs_error_tag,ls_err_message GOTO _imp_error END IF
        }
        #20170405 end

        #檢核匯入包必要條件
        CALL sadzi140_exim_table_check_notice(ls_TableName,mo_json_obj,mo_local_sys_info.*,mo_src_pkg_sys_info.*) RETURNING lb_success
        DISPLAY "Check export notice : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 

        #抓取判斷匯出檔是不是匯出端和匯入端相同
        IF mo_src_pkg_sys_info.psi_DGENV = mo_local_sys_info.psi_DGENV THEN LET mb_same_dgenv = TRUE ELSE LET mb_same_dgenv = FALSE END IF
        #抓取判斷來源資料庫類型是不是匯出端和匯入端相同
        IF mo_src_pkg_sys_info.psi_DB_TYPE = mo_local_sys_info.psi_DB_TYPE THEN LET mb_same_db = TRUE ELSE LET mb_same_db = FALSE END IF
        #抓取判斷來源行業別是不是匯出端和匯入端相同
        IF mo_src_pkg_sys_info.psi_TOPIND = mo_local_sys_info.psi_TOPIND THEN LET mb_same_topind = TRUE ELSE LET mb_same_topind = FALSE END IF
        #抓取判斷來源包是不是經由更名式匯入的包
        IF mo_src_pkg_sys_info.psi_RENAMED = "Y" THEN LET mb_renamed = TRUE ELSE LET mb_renamed = FALSE END IF

        BEGIN WORK
        
        #20170419 begin
        #先做 gztd_t 的匯入
        CALL sadzi140_exim_import_GZTD_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        DISPLAY "Import GZTD : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        CALL sadzi140_exim_import_GZTDL_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        DISPLAY "Import GZTDL : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF
        TRY 
          COMMIT WORK 
          BEGIN WORK
        CATCH
        END TRY  
        #20170419 end
        
        #非為更名式匯入, 才匯DZEA的資料
        IF NOT mb_renamed THEN 
          CALL sadzi140_exim_import_DZEA_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
          DISPLAY "Import DZEA : ",IIF(lb_success,cs_result_success,cs_result_fail)
          IF NOT lb_success THEN GOTO _imp_error END IF 
          CALL sadzi140_exim_import_DZEAL_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
          DISPLAY "Import DZEAL : ",IIF(lb_success,cs_result_success,cs_result_fail)
          IF NOT lb_success THEN GOTO _imp_error END IF
        END IF  
        
        CALL sadzi140_exim_delete_DZEB_data(ls_TableName,mb_same_dgenv,mb_same_topind) RETURNING lb_success #20190304 
        DISPLAY "Delete DZEB : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        CALL sadzi140_exim_import_DZEB_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        DISPLAY "Import DZEB : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        CALL sadzi140_exim_import_DZEBL_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        DISPLAY "Import DZEBL : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        CALL sadzi140_exim_delete_DZEC_data(ls_TableName,mb_same_dgenv,mb_same_topind) RETURNING lb_success #20190304 
        DISPLAY "Delete DZEC : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        CALL sadzi140_exim_import_DZEC_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        DISPLAY "Import DZEC : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        CALL sadzi140_exim_delete_DZED_data(ls_TableName,mb_same_dgenv,mb_same_topind) RETURNING lb_success #20190304 
        DISPLAY "Delete DZED : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        CALL sadzi140_exim_import_DZED_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        DISPLAY "Import DZED : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF
        CALL sadzi140_exim_delete_DZEN_data(ls_TableName,mb_same_dgenv,mb_same_topind) RETURNING lb_success #20190304 
        DISPLAY "Delete DZEN : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        CALL sadzi140_exim_import_DZEN_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        DISPLAY "Import DZEN : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        
        -- adzi150 begin        
        CALL sadzi140_exim_import_DZEF_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        DISPLAY "Import DZEF : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        CALL sadzi140_exim_import_DZEP_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        DISPLAY "Import DZEP : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        CALL sadzi140_exim_import_DZEQ_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        DISPLAY "Import DZEQ : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        CALL sadzi140_exim_import_DZER_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        DISPLAY "Import DZER : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        CALL sadzi140_exim_import_DZET_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        DISPLAY "Import DZET : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        -- adzi150 end        
        
        #CALL sadzi140_exim_import_DZEO_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        #DISPLAY "Import DZEO : ",IIF(lb_success,cs_result_success,cs_result_fail)
        #IF NOT lb_success THEN GOTO _imp_error END IF 
        CALL sadzi140_exim_import_DZEU_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        DISPLAY "Import DZEU : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF
        #CALL sadzi140_exim_import_DZEV_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success 
        #DISPLAY "Import DZEV : ",IIF(lb_success,cs_result_success,cs_result_fail)
        #IF NOT lb_success THEN GOTO _imp_error END IF 
        #CALL sadzi140_exim_import_DZEW_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        #DISPLAY "Import DZEW : ",IIF(lb_success,cs_result_success,cs_result_fail)
        #IF NOT lb_success THEN GOTO _imp_error END IF
        #20170419 mark begin
        { 
        CALL sadzi140_exim_import_GZTD_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        DISPLAY "Import GZTD : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        CALL sadzi140_exim_import_GZTDL_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        DISPLAY "Import GZTDL : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF
        } 
        #20170419 mark end
        CALL sadzi140_exim_import_DZIG_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        DISPLAY "Import DZIG : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF
        -- adzi150 begin
        CALL sadzi140_exim_import_GZCC_data(ls_TableName,mo_json_obj,mo_PARAMETERS.*) RETURNING lb_success
        DISPLAY "Import GZCC : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        -- adzi150 end
        CALL sadzi140_exim_delete_invalid_dzeu_data(ls_TableName) RETURNING lb_success
        DISPLAY "Delete invalid DZEU_T data : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        CALL sadzi140_exim_add_missing_dzeu_data(ls_TableName) RETURNING lb_success
        DISPLAY "Add missing DZEU_T data : ",IIF(lb_success,cs_result_success,cs_result_fail)
        IF NOT lb_success THEN GOTO _imp_error END IF 
        
        LABEL _imp_error:
        
        IF NOT lb_success THEN
          TRY  
            ROLLBACK WORK
          CATCH 
          END TRY  
        ELSE
          TRY  
            COMMIT WORK
          CATCH 
          END TRY  
          #20171120 begin
          #重Load一次表格欄位
          CALL sadzi140_db_get_column_with_sequence(ls_TableName) RETURNING mo_column_info 
          CALL sadzi140_db_check_and_remove_invalid_key_index(mo_column_info,ls_TableName,ms_dlang) RETURNING lb_result
          CALL sadzi140_db_set_constraint_data(ls_TableName) RETURNING lb_result
          #20171120 end
          CALL sadzi140_db_update_alter_code(ms_MasterUser,ls_TableName)
        END IF 
        
        LABEL _exit_imp:
        
        #返回原路徑  
        CALL os.Path.chdir(ms_entry_path) RETURNING li_CHDIR
        
      END IF          
      
      LET lb_fault = NOT lb_success
      
    #Create table or modify table
    WHEN cs_exim_assemble
      LET lb_success = TRUE
      CALL sadzi140_db_check_table_exist(cs_master_user,ls_TableName) RETURNING lb_TableExists
      IF lb_TableExists THEN
        CALL sadzi140_exim_modify_table(ls_TableName,ls_ConstructVersion) RETURNING lb_success           
      ELSE
        CALL sadzi140_exim_create_table(ls_TableName,ls_ConstructVersion,ls_alm_request_no,ls_dgenv) RETURNING lb_success               
        #CALL sadzi140_vcs_reset_vcs_data(ls_TableName)
      END IF
      CALL sadzi140_db_update_alter_code(ms_MasterUser,ls_TableName)
      DISPLAY cs_information_tag," Check and generate language file."

      IF NOT mb_backend_mode THEN
        CALL sadzi140_db_gen_db_schema(ms_MasterUser.toLowerCase(),ls_TableName,sadzi140_exim_get_table_type(ls_TableName)) RETURNING lb_result
        IF NOT lb_result THEN DISPLAY cs_error_tag," Generate ds.sch error." END IF
      END IF   

      IF (sadzi140_exim_get_table_type(ls_TableName) = "L") AND NOT (mb_backend_mode) THEN
        DISPLAY cs_information_tag," Generating language file."
        #產生多語言 4gl 檔
        CALL sadzi140_util_gen_multi_lang(ls_TableName)
      END IF
      
      LET lb_fault = NOT lb_success
    #Kill snapshot  
    WHEN cs_exim_delete 
      LET lb_DelSuccess = TRUE
      IF ls_TableName IS NOT NULL THEN
        CALL sadzi140_exim_get_table_version_info(ls_TableName) RETURNING lo_VersionInfo.*
        CALL sadzi140_util_kill_snapshot(ms_MasterUser,ls_TableName,lo_VersionInfo.*) RETURNING lb_success
        IF lb_success THEN
          {
          IF ls_SD_Version = cs_rt_design_sd_version THEN 
            CALL sadzi140_exim_get_max_rev_version(ls_TableName,ls_ConstructVersion) RETURNING ls_Revision
          ELSE
            LET ls_Revision = ls_SD_Version
          END IF
          }    
          CALL sadzi140_exim_get_max_rev_version(ls_TableName,ls_ConstructVersion) RETURNING ls_Revision
          IF NVL(ls_Revision,cs_null_value) <> cs_null_value THEN
            LET lo_revision.rv_ALM_VERSION  = ls_ConstructVersion
            LET lo_revision.rv_SEQUENCE     = ls_Revision
            CALL sadzi140_rev_make_revision(ls_TableName,lo_revision.*) RETURNING lb_DelSuccess
            IF lb_DelSuccess THEN
              CALL sadzi140_db_update_alter_code(ms_MasterUser,ls_TableName)
            END IF
          ELSE
            DISPLAY "[adz-00192] There are no snapshot data for fetch."
            LET lb_DelSuccess = FALSE  
          END IF   
        END IF    
      END IF  
      LET lb_fault = NOT lb_DelSuccess
    #Drop table and related files.  
    WHEN cs_exim_drop 
      LET lb_DropSuccess = TRUE
      IF ls_TableName IS NOT NULL THEN
        LET lb_DropSuccess = FALSE      
        CALL sadzi140_util_delete_master_table(ms_MasterUser,"D",ls_TableName,TRUE) RETURNING lb_DropSuccess 
        IF NOT mb_backend_mode THEN
          CALL sadzi140_db_gen_db_schema(ms_MasterUser.toLowerCase(),ls_TableName,sadzi140_exim_get_table_type(ls_TableName)) RETURNING lb_result
          IF NOT lb_result THEN DISPLAY cs_error_tag," Generate ds.sch error." END IF
        END IF   
      END IF  
      LET lb_fault = NOT lb_DropSuccess
    #Alter Table  
    WHEN cs_exim_alter
      LET lb_ASMSuccess = TRUE
      CALL sadzi140_db_check_table_exist(cs_master_user,ls_TableName) RETURNING lb_TableExists
      IF lb_TableExists THEN
        CALL sadzi140_exim_modify_table(ls_TableName,ls_ConstructVersion) RETURNING lb_ASMSuccess           
      ELSE
        CALL sadzi140_exim_create_table(ls_TableName,ls_ConstructVersion,ls_alm_request_no,ls_dgenv) RETURNING lb_ASMSuccess               
        #CALL sadzi140_vcs_reset_vcs_data(ls_TableName)
      END IF
      IF (sadzi140_exim_get_table_type(ls_TableName) = "L") AND NOT (mb_backend_mode) THEN
        #產生多語言 4gl 檔
        CALL sadzi140_util_gen_multi_lang(ls_TableName)
      END IF
      LET lb_fault = NOT lb_ASMSuccess
    #Table rebuild.  
    WHEN cs_exim_rebuild
      LET lb_RebuildSuccess = TRUE      
      IF ls_TableName IS NOT NULL THEN

        #異動前檢核欄位和索引與實際表格有沒有差異
        CALL sadzi140_shot_check_diff(ms_MasterUser,ls_TableName,TRUE) RETURNING lb_diff
        #呼叫表格重建函式
        CALL sadzi140_util_table_rebuild(ms_MasterUser,ls_TableName,TRUE,ls_ConstructVersion) RETURNING lb_RebuildSuccess,ls_rtn_code,ls_message
        IF lb_RebuildSuccess AND ls_rtn_code = "1" THEN
          CALL sadzi140_db_update_alter_code(ms_MasterUser,ls_TableName)
          IF NOT mb_backend_mode THEN
            CALL sadzi140_db_gen_db_schema(ms_MasterUser.toLowerCase(),ls_TableName,sadzi140_exim_get_table_type(ls_TableName)) RETURNING lb_result
            IF NOT lb_result THEN DISPLAY cs_error_tag," Generate ds.sch error." END IF
          END IF   
          IF (sadzi140_exim_get_table_type(ls_TableName) = "L") AND NOT (mb_backend_mode) THEN
            #產生多語言 4gl 檔
            CALL sadzi140_util_gen_multi_lang(ls_TableName)
          END IF             
        END IF  
        
      END IF  
      LET lb_fault = NOT lb_RebuildSuccess
    WHEN cs_null_value   
      DISPLAY cs_error_tag,"No any parameters.","\n"
  OTHERWISE
    DISPLAY cs_error_tag,"Input parameters error !!"
  END CASE

  LABEL _exit_start: 

  #返回原路徑  
  CALL os.Path.chdir(ms_entry_path) RETURNING li_CHDIR  

  LET mb_fault = lb_fault
  LET mo_export_info.RESULT = mb_fault

  LET mo_exec_result.er_RESULT  = mb_fault 
  LET mo_exec_result.er_CODE    = ls_error_code
  LET mo_exec_result.er_MESSAGE = ls_err_message

END FUNCTION

FUNCTION sadzi140_exim_finalize()
END FUNCTION

################################################################################

FUNCTION sadzi140_exim_set_export_notice_sys_info(p_pkg_sys_info,p_json_obj,p_if_renamed)
DEFINE
  p_pkg_sys_info T_PKG_SYS_INFO, 
  p_json_obj     util.JSONObject,
  p_if_renamed   BOOLEAN
DEFINE
  ls_RENAMED  STRING

  LET ls_RENAMED = IIF(p_if_renamed,"Y","N")

  #給定DGENV的值
  CALL p_json_obj.put(cs_keyword_DGENV,p_pkg_sys_info.psi_DGENV)
  CALL p_json_obj.put(cs_keyword_ERPID,p_pkg_sys_info.psi_ERPID)
  CALL p_json_obj.put(cs_keyword_CUST,p_pkg_sys_info.psi_CUST)
  CALL p_json_obj.put(cs_keyword_DB_TYPE,p_pkg_sys_info.psi_DB_TYPE)
  CALL p_json_obj.put(cs_keyword_TOPIND,p_pkg_sys_info.psi_TOPIND)
  CALL p_json_obj.put(cs_keyword_renamed,ls_RENAMED)  
  CALL p_json_obj.put(cs_keyword_ERPVER,p_pkg_sys_info.psi_ERPVER) #20180129

END FUNCTION 

FUNCTION sadzi140_exim_making_work_directory(p_TableName,p_MakeType)
DEFINE
  p_TableName STRING,
  p_MakeType  STRING
DEFINE
  ls_TableName STRING,
  ls_MakeType  STRING
DEFINE
  li_ConstructVersion INTEGER,
  li_SD_Version       INTEGER,
  ls_PathName         STRING,
  ls_TEMPDIR          STRING,
  ls_GUID             STRING,
  li_MKDIR            INTEGER,
  li_CHDIR            INTEGER,
  ls_os_separator     STRING,
  ls_command          STRING
DEFINE 
  lb_return BOOLEAN, 
  ls_return STRING   

  LET ls_TableName = p_TableName
  LET ls_MakeType  = p_MakeType

  LET lb_return = TRUE

  #取得作業系統的分隔符號
  CALL os.Path.separator() RETURNING ls_os_separator
  
  LET ls_GUID = security.RandomGenerator.CreateUUIDString()
  
  LET ls_TEMPDIR          = FGL_GETENV("TEMPDIR")
  LET li_ConstructVersion = mo_PARAMETERS.DZLM005
  LET li_SD_Version       = mo_PARAMETERS.DZLM006
  LET ls_PathName = ls_TEMPDIR,ls_os_separator,NVL(ls_TableName,cs_no_define),"+TABLE+",
                    NVL(mo_PARAMETERS.DZLM012,cs_no_define),"+",
                    (NVL(mo_PARAMETERS.DZLM015,0) USING cs_UsingFormat),".",ls_GUID,".",ls_MakeType

  {                  
  LET ls_command = 'mkdir ',ls_PathName                    
  RUN ls_command
  LET ls_command = 'cd ',ls_PathName
  RUN ls_command
  }
  
  CALL os.Path.mkdir(ls_PathName) RETURNING li_MKDIR

  IF NOT os.Path.exists(ls_PathName) THEN
    LET lb_return = FALSE
    DISPLAY cs_error_tag," The temporary path '",ls_PathName,"' doesn't exists !!"
  ELSE  
    CALL os.Path.chdir(ls_PathName) RETURNING li_CHDIR  
  END IF 
  
  LET ls_return = ls_PathName
  
  RETURN lb_return,ls_return  
  
END FUNCTION

FUNCTION sadzi140_exim_get_DZLM_data(p_ConstructVersion,p_SD_Version)
DEFINE
  p_ConstructVersion  STRING,
  p_SD_Version        STRING
DEFINE
  ls_ConstructVersion  STRING,
  ls_SD_Version        STRING,
  lo_DZLM_T            DYNAMIC ARRAY OF T_DZLM_T_SCR
DEFINE  
  ls_SQL     STRING,
  li_RecCnt  INTEGER,
  lo_RETURN  DYNAMIC ARRAY OF T_DZLM_T_SCR

  LET ls_ConstructVersion = p_ConstructVersion
  LET ls_SD_Version       = p_SD_Version   

  INITIALIZE lo_DZLM_T TO NULL #20180213
  INITIALIZE lo_RETURN TO NULL #20180213
  
  LET ls_SQL = "select '',                                                     ",
               "       lm.dzlm001,lm.dzlm002,lm.dzlm003,lm.dzlm004,lm.dzlm005, ",
               "       lm.dzlm006,lm.dzlm007,lm.dzlm008,lm.dzlm009,lm.dzlm010, ",
               "       lm.dzlm011,lm.dzlm012,lm.dzlm013,lm.dzlm014,lm.dzlm015, ",
               "       lm.dzlm016,lm.dzlm017,lm.dzlm018,lm.dzlm019,lm.dzlm020, ",
               "       lm.dzlm021                                              ",  
               "  from dzlm_t lm                                               ",
               " where lm.dzlm001 = 'T'                                        ",
               "   and lm.dzlm005 = '",ls_ConstructVersion,"'                  ",
               "   and lm.dzlm006 = '",ls_SD_Version,"'                        "

  PREPARE lpre_GetDZLMData FROM ls_SQL
  DECLARE lcur_GetDZLMData CURSOR FOR lpre_GetDZLMData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZLMData
  FOREACH lcur_GetDZLMData INTO lo_DZLM_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZLMData
  CALL lo_DZLM_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZLMData
  FREE lcur_GetDZLMData  

  LET lo_RETURN = lo_DZLM_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_get_table_type(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_return     STRING,
  ls_sql        STRING,
  ls_table_type VARCHAR(10),
  ls_table_name STRING

  LET ls_table_name = p_table_name

  LET ls_sql = "select ea.dzea004                      ",
               "  from dzea_t ea                       ",
               " where ea.dzea001 = '",ls_table_name,"'"
 
  PREPARE lpre_get_table_type FROM ls_sql
  DECLARE lcur_get_table_type CURSOR FOR lpre_get_table_type
  OPEN lcur_get_table_type
  FETCH lcur_get_table_type INTO ls_table_type
  CLOSE lcur_get_table_type
  FREE lcur_get_table_type
  FREE lpre_get_table_type

  LET ls_return = ls_table_type
  
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_exim_get_table_version_info(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name STRING,
  ls_sql        STRING,
  lo_table      T_VERSION_INFO   
DEFINE
  lo_return     T_VERSION_INFO

  LET ls_table_name = p_table_name

  LET ls_sql = "select ea.dzea012,ea.dzea013           ",
               "  from dzea_t ea                       ",
               " where ea.dzea001 = '",ls_table_name,"'"
 
  PREPARE lpre_get_table_version_info FROM ls_sql
  DECLARE lcur_get_table_version_info CURSOR FOR lpre_get_table_version_info
  OPEN lcur_get_table_version_info
  FETCH lcur_get_table_version_info INTO lo_table.*
  CLOSE lcur_get_table_version_info
  FREE lcur_get_table_version_info
  FREE lpre_get_table_version_info

  LET lo_return.* = lo_table.*
  
  RETURN lo_return.*
  
END FUNCTION

FUNCTION sadzi140_exim_get_max_rev_version(p_table_name,p_alm_version)
DEFINE
  p_table_name  STRING,
  p_alm_version STRING 
DEFINE
  ls_table_name      STRING,
  ls_alm_version     STRING,
  ls_sql             STRING,
  ls_max_rev_version VARCHAR(40)
DEFINE
  ls_return STRING

  LET ls_table_name  = p_table_name
  LET ls_alm_version = p_alm_version

  #呼叫取得最大的Revision Number
  CALL sadzi140_shot_get_max_construct_revision(ms_MasterUser,ls_table_name,ls_alm_version) RETURNING ls_max_rev_version 

  LET ls_return = ls_max_rev_version
  
  RETURN ls_return
  
END FUNCTION

## DZEA_T ######################################################################

FUNCTION sadzi140_exim_export_DZEA_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEA_T"
  
  CALL sadzi140_exim_get_DZEA_data(ls_TableName,p_json_obj,lo_PARAMETERS.*) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
                    
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_DZEA_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName   STRING,
  p_json_obj    util.JSONObject,
  po_PARAMETERS T_DZLM_T_SCR  
DEFINE
  ls_TableName    STRING,
  lo_PARAMETERS   T_DZLM_T_SCR,
  ls_MaxVersion   STRING,
  ls_ALMInfo      STRING,
  lo_DZEA_T       DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*
  
  INITIALIZE lo_DZEA_T TO NULL #20180213
  INITIALIZE lo_RETURN TO NULL #20180213
  
  LET ls_ALMInfo = ""

  #DZEA019,020,021 為 GUID, 匯出空白
  LET ls_SQL = "select ea.dzea001,ea.dzea002,ea.dzea003,ea.dzea004,ea.dzea005,",
               "       ea.dzea006,ea.dzea007,ea.dzea008,ea.dzea009,ea.dzea010,",
               "       ea.dzea011,ea.dzea012,ea.dzea013,ea.dzea014,ea.dzea015,",
               "       ea.dzea016,ea.dzea017,ea.dzea018,'' dzea019,'' dzea020,",
               "       '' dzea021                                             ",  
               "  from dzea_t ea                                              ",
               " where ea.dzea001 = '",ls_TableName,"'                        "

  #取得Table export notice             
  LET ls_tsn_SQL = "select 'DZEA_T' table_name,count(1) records  ",
                   "  from dzea_t ea                             ",
                   " where ea.dzea001 = '",ls_TableName,"'       "

  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)    
              
  PREPARE lpre_GetDZEAData FROM ls_SQL
  DECLARE lcur_GetDZEAData CURSOR FOR lpre_GetDZEAData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZEAData
  FOREACH lcur_GetDZEAData INTO lo_DZEA_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZEAData
  CALL lo_DZEA_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZEAData
  FREE lcur_GetDZEAData  

  LET lo_RETURN = lo_DZEA_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_import_DZEA_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEA_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZEA_T_data(lo_DATAS) RETURNING lb_success
    END IF  
  END IF
  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZEA_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  ls_user         VARCHAR(100),
  ldt_datetime    DATETIME YEAR TO SECOND,
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  &ifndef DEBUG
  LET ls_user = g_user
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  &endif
  LET ldt_datetime = CURRENT YEAR TO SECOND
  
  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL THEN
      LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "'",lo_DATAS[li_loop,3],"',",
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",lo_DATAS[li_loop,5],"',",
                           "'",lo_DATAS[li_loop,6],"',",
                           "'",lo_DATAS[li_loop,7],"',",
                           "'",lo_DATAS[li_loop,8],"',",
                           "'",lo_DATAS[li_loop,9],"',",
                           "'",lo_DATAS[li_loop,10],"',",
                           "'",lo_DATAS[li_loop,11],"',",
                           "'",lo_DATAS[li_loop,12],"',",
                           #"'",IIF(mb_same_dgenv,ms_max_sequence,lo_DATAS[li_loop,13]),"',", #異動序號
                           "'",ms_max_sequence,"',", #異動序號
                           "'",lo_DATAS[li_loop,14],"',",
                           "'",lo_DATAS[li_loop,15],"',",
                           "'",lo_DATAS[li_loop,16],"',",
                           "'",lo_DATAS[li_loop,17],"',",
                           "'",lo_DATAS[li_loop,18],"',",
                           "'",lo_DATAS[li_loop,19],"',",
                           "'",lo_DATAS[li_loop,20],"',",
                           "'",lo_DATAS[li_loop,21],"',",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,",", #20191209
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,"" #20191209

      #DZEA019,020,021 為 GUID, 只Insert不Update                     
      LET ls_InsertSQL = "insert into dzea_t (                                         ",
                         "                    dzea001,dzea002,dzea003,dzea004,dzea005, ",
                         "                    dzea006,dzea007,dzea008,dzea009,dzea010, ",
                         "                    dzea011,dzea012,dzea013,dzea014,dzea015, ",
                         "                    dzea016,dzea017,dzea018,dzea019,dzea020, ", 
                         "                    dzea021,                                 ",
                         "                    dzeacrtid,dzeacrtdt,dzeamodid,dzeamoddt  ", 
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "

      #DZEA015,DZEA018 為客制標示, 不更新                   
      #DZEA019,020,021 為 GUID, 只Insert不Update                   
      LET ls_UpdateValue = "dzea002 = '",lo_DATAS[li_loop,2],"',",
                           "dzea003 = '",lo_DATAS[li_loop,3],"',",
                           "dzea004 = '",lo_DATAS[li_loop,4],"',",
                           "dzea005 = '",lo_DATAS[li_loop,5],"',",
                           "dzea006 = '",lo_DATAS[li_loop,6],"',",
                           "dzea007 = '",lo_DATAS[li_loop,7],"',",
                           "dzea008 = '",lo_DATAS[li_loop,8],"',",
                           "dzea009 = '",lo_DATAS[li_loop,9],"',",
                           "dzea010 = '",lo_DATAS[li_loop,10],"',",
                           "dzea011 = '",lo_DATAS[li_loop,11],"',",
                           "dzea012 = '",lo_DATAS[li_loop,12],"',",
                           #"dzea013 = '",IIF(mb_same_dgenv,ms_max_sequence,lo_DATAS[li_loop,13]),"',", #異動序號
                           "dzea013 = '",ms_max_sequence,"',", #異動序號
                           "dzea014 = '",lo_DATAS[li_loop,14],"',",
                           #"dzea015 = '",lo_DATAS[li_loop,15],"',",
                           "dzea016 = '",lo_DATAS[li_loop,16],"',",
                           "dzea017 = '",lo_DATAS[li_loop,17],"',",
                           #"dzea018 = '",lo_DATAS[li_loop,18],"',",
                           #"dzea019 = '",lo_DATAS[li_loop,19],"',",
                           #"dzea020 = '",lo_DATAS[li_loop,20],"',",
                           #"dzea021 = '",lo_DATAS[li_loop,21],"',",
                           "dzeamodid = trim('",ls_user,"'),",
                           "dzeamoddt = ",cs_get_time_stamp," " #20191209
                           
      LET ls_UpdateSQL = "update dzea_t set ",
                         ls_UpdateValue,
                         " where dzea001 = '",lo_DATAS[li_loop,1],"'",
                         "   and coalesce(dzea015,'",cs_dgenv_standard,"') = '",cs_dgenv_standard,"'" 
                         
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
    END IF
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZEBL_T #####################################################################

FUNCTION sadzi140_exim_export_DZEBL_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEBL_T"
  
  CALL sadzi140_exim_get_DZEBL_data(ls_TableName,p_json_obj) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
                    
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_DZEBL_data(p_TableName,p_json_obj)
DEFINE
  p_TableName  STRING,
  p_json_obj    util.JSONObject  
DEFINE
  ls_TableName  STRING,
  lo_DZEBL_T     DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName = p_TableName

  INITIALIZE lo_DZEBL_T TO NULL  #20180213
  INITIALIZE lo_RETURN TO NULL  #20180213
  
  LET ls_SQL = "select ebl.dzebl001,ebl.dzebl002,ebl.dzebl003,ebl.dzebl004,ebl.dzebl000 ",
               "  from dzebl_t ebl                                                      ",
               " where exists (                                                         ",
               "                select 1                                                ",
               "                  from dzeb_t eb                                        ",
               "                 where eb.dzeb002 = ebl.dzebl001                        ",
               "                   and eb.dzeb001 = '",ls_TableName,"'                  ",
               "              )                                                         "
               
  #取得Table export notice             
  LET ls_tsn_SQL = "select 'DZEBL_T' table_name,count(1) records           ",
                   "  from dzebl_t ebl                                     ",
                   " where exists (                                        ",
                   "                select 1                               ",
                   "                  from dzeb_t eb                       ",
                   "                 where eb.dzeb002 = ebl.dzebl001       ",
                   "                   and eb.dzeb001 = '",ls_TableName,"' ",
                   "              )                                        "

  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)    

  PREPARE lpre_GetDZEBLData FROM ls_SQL
  DECLARE lcur_GetDZEBLData CURSOR FOR lpre_GetDZEBLData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZEBLData
  FOREACH lcur_GetDZEBLData INTO lo_DZEBL_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZEBLData
  CALL lo_DZEBL_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZEBLData
  FREE lcur_GetDZEBLData  

  LET lo_RETURN = lo_DZEBL_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_import_DZEBL_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEBL_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZEBL_T_data(lo_DATAS) RETURNING lb_success
    END IF  
  END IF  

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZEBL_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  lb_is_lang      BOOLEAN,
  li_lng          INTEGER,
  li_loop         INTEGER,
  ls_IMPORT_TABLE STRING, #20180302 
  lb_result       BOOLEAN, #20171120
  lb_is_udfc      BOOLEAN
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  LET ls_IMPORT_TABLE = "DZEBL_T" #20180302
  
  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN
    
      #20171120 begin  
      IF mb_pickup_exist THEN 
        CALL sadzi140_exim_check_if_exist_in_pick_list(mo_pick_col_list,lo_DATAS[li_loop,1]) RETURNING lb_result
        IF NOT lb_result THEN CONTINUE FOR END IF
      END IF
      #20171120 end

      #20180302 begin  
      IF mo_local_sys_info.psi_RT_PATCH_MODE <> "Y" THEN 
        CALL sadzi140_exim_check_if_exist_in_deleted_std_col_list(mo_deleted_std_col_list,lo_DATAS[li_loop,1]) RETURNING lb_result
        IF lb_result THEN 
          DISPLAY cs_warning_tag,"Column ",lo_DATAS[li_loop,1]," exists in deleted standard column list, skip import to ",ls_IMPORT_TABLE,".","\n",
                                 "You can using 'Tools'->'Field selection import' to import it !" 
          CONTINUE FOR 
        END IF
      END IF
      #20180302 end
      
      #從清單中取得可匯入的語言別
      LET lb_is_lang = FALSE
      FOR li_lng = 1 TO mo_local_lang.getLength()
        IF lo_DATAS[li_loop,2] = mo_local_lang[li_lng] THEN
          LET lb_is_lang = TRUE
          EXIT FOR
        END IF 
      END FOR 

      #符合當地多語言清單的資料才匯入
      IF lb_is_lang THEN
      
        LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                             "'",lo_DATAS[li_loop,2],"',",
                             "'",lo_DATAS[li_loop,3],"',",
                             "'",lo_DATAS[li_loop,4],"',",
                             "'",lo_DATAS[li_loop,5],"'"
                             
        LET ls_InsertSQL = "insert into dzebl_t (                                              ",
                           "                      dzebl001,dzebl002,dzebl003,dzebl004,dzebl000 ",
                           "                    )                                              ",  
                           " values (                                                          ",
                           ls_InsertValue,
                           "        )                                                          "
                           
        LET ls_UpdateValue = "dzebl003 = '",lo_DATAS[li_loop,3],"',",
                             "dzebl004 = '",lo_DATAS[li_loop,4],"',",
                             "dzebl000 = '",lo_DATAS[li_loop,5],"'"

        #只更新非自定義欄位資料
        LET ls_UpdateSQL = "update dzebl_t set                          ",
                           ls_UpdateValue,
                           " where dzebl001 = '",lo_DATAS[li_loop,1],"' ",
                           "   and dzebl002 = '",lo_DATAS[li_loop,2],"' "
                           
        CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
        IF NOT lb_success THEN
          CALL sadzi140_exim_check_if_user_define_column(lo_DATAS[li_loop,5],lo_DATAS[li_loop,1]) RETURNING lb_is_udfc
          -- 非使用者定義欄位或者是使用者定義欄位而且同樣的區域才執行更新
          IF (NOT lb_is_udfc) OR (lb_is_udfc AND mb_same_dgenv) THEN 
            CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
          ELSE
            LET lb_success = TRUE  
          END IF  
          IF NOT lb_success THEN
            DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
            DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
            EXIT FOR 
          END IF
        END IF
      END IF
    END IF   
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZEB_T #####################################################################

FUNCTION sadzi140_exim_export_DZEB_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEB_T"
  
  CALL sadzi140_exim_get_DZEB_data(ls_TableName,p_json_obj,lo_PARAMETERS.*) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_DZEB_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName   STRING,
  p_json_obj    util.JSONObject,
  po_PARAMETERS T_DZLM_T_SCR  
DEFINE
  ls_TableName    STRING,
  lo_PARAMETERS   T_DZLM_T_SCR,
  ls_MaxVersion   STRING,
  lo_DZEB_T       DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  INITIALIZE lo_DZEB_T TO NULL  #20180213
  INITIALIZE lo_RETURN TO NULL  #20180213

  IF (lo_parameters.DZLM005 = cs_NoVersionCharacter) OR 
     (NOT sadzi140_exim_check_version_data_exist(ls_TableName)) OR 
     (NOT sadzi140_exim_check_alm_request_data_exist(ls_TableName,lo_PARAMETERS.*))THEN

    LET ls_SQL = "select eb.dzeb001,eb.dzeb002,eb.dzeb003,eb.dzeb004,eb.dzeb005,",
                 "       eb.dzeb006,eb.dzeb007,eb.dzeb008,eb.dzeb012,eb.dzeb021,",
                 "       eb.dzeb022,eb.dzeb023,eb.dzeb024,eb.dzeb027,eb.dzeb028,",
                 "       eb.dzeb029,eb.dzeb030,eb.dzeb031                       ",
                 "  from dzeb_t eb                                              ",
                 " where eb.dzeb001 = '",ls_TableName,"'                        ",
                 " order by eb.dzeb021                                          "

    #取得Table export notice             
    LET ls_tsn_SQL = "select 'DZEB_T' table_name,count(1) records  ",
                     "  from dzeb_t eb                             ",
                     " where eb.dzeb001 = '",ls_TableName,"'       "

  ELSE
    &ifdef is_msv
    LET ls_SQL = "select lower(ev.dzev002)                         dzeb001,          ",
                 "       lower(ev.dzev005)                         dzeb002,          ",
                 "       ev.dzev010                                dzeb003,          ",
                 "       ev.dzev006                                dzeb004,          ",
                 "       case ev.dzev007                                             ",
                 "         when 'Y' then 'N'                                         ",
                 "         when 'N' then 'Y'                                         ",
                 "       else                                                        ",
                 "         'Y'                                                       ",
                 "       end  dzeb005,                                               ",
                 "       ev.dzev013                                dzeb006,          ",
                 "       lower(ev.dzev008)                         dzeb007,          ",
                 "       case upper(ev.dzev008)                                      ",
                 "         when 'DATE' then null                                     ",
                 "         when 'INT' then null                                      ",
                 "         when 'SMALLINT' then null                                 ",
                 "       else                                                        ",
                 "         ev.dzev009                                                ",
                 "       end                                       dzeb008,          ",
                 "       ev.dzev021                                dzeb012,          ",
                 "       ev.dzev014                                dzeb021,          ",
                 "       ev.dzev015                                dzeb022,          ",
                 "       ev.dzev016                                dzeb023,          ",
                 "       ev.dzev017                                dzeb024,          ",
                 "       iif(                                                        ",
                 "            charindex('",cs_keyword_TOPIND,"',trim(eb.dzeb027))=0, ",
                 "            null,trim(eb.dzeb027)                                  ",
                 "          )                                      dzeb027,          ",
                 "       'N'                                       dzeb028,          ",
                 "       ev.dzev020                                dzeb029,          ",
                 "       ev.dzev022                                dzeb030,          ",
                 "       eb.dzeb031                                dzeb031           ",
                 "  from dzev_t ev                                                   ",
                 "  left outer join dzeb_t eb on eb.dzeb001 = lower(ev.dzev002)      ",
                 "                           and eb.dzeb002 = lower(ev.dzev005)      ",
                 " where ev.dzev002 = '",ls_TableName.toUpperCase(),"'               ",
                 "   and ev.dzev003 = trim('",lo_parameters.DZLM015,"')              ",
                 "   and ev.dzev004 = 'TableDesigner'                                ",
                 "   and ev.dzev018 = '",lo_parameters.DZLM005,"'                    ",
                 "   and ev.dzev019 = '",lo_parameters.DZLM012,"'                    ",
                 " order by ev.dzev011                                               "
                 
    #取得Table export notice             
    LET ls_tsn_SQL = "select 'DZEB_T' table_name,count(1) records          ",
                     "  from dzev_t ev                                     ",
                     " where ev.dzev002 = '",ls_TableName.toUpperCase(),"' ",
                     "   and ev.dzev003 = trim('",lo_parameters.DZLM015,"')", 
                     "   and ev.dzev004 = 'TableDesigner'                  ",
                     "   and ev.dzev018 = '",lo_parameters.DZLM005,"'      ",
                     "   and ev.dzev019 = '",lo_parameters.DZLM012,"'      "
                     
    &else
    &ifdef is_pgs
    LET ls_SQL = "select lower(ev.dzev002)                         dzeb001,     ",    
                 "       lower(ev.dzev005)                         dzeb002,     ",    
                 "       ev.dzev010                                dzeb003,     ",    
                 "       ev.dzev006                                dzeb004,     ",    
                 "       case ev.dzev007                                        ",    
                 "         when 'Y' then 'N'                                    ",    
                 "         when 'N' then 'Y'                                    ",    
                 "       else                                                   ",    
                 "         'Y'                                                  ",    
                 "       end  dzeb005,                                          ",    
                 "       ev.dzev013                                dzeb006,     ",    
                 "       lower(ev.dzev008)                         dzeb007,     ",    
                 "       case upper(ev.dzev008)                                 ",    
                 "         when 'BYTEA' then null                               ",     
                 "         when 'TEXT' then null                                ",     
                 "         when 'DATE' then null                                ",
                 "         when 'INT2' then null                                ",
                 "         when 'INT4' then null                                ",
                 "         when 'INT8' then null                                ",
                 "       else                                                   ",    
                 "         ev.dzev009                                           ",    
                 "       end                                       dzeb008,     ",    
                 "       ev.dzev021                                dzeb012,     ",    
                 "       ev.dzev014                                dzeb021,     ",    
                 "       ev.dzev015                                dzeb022,     ",    
                 "       ev.dzev016                                dzeb023,     ",    
                 "       ev.dzev017                                dzeb024,     ",    
                 "       case                                                   ",     
                 "         when strpos(trim(eb.dzeb027),                        ",
                 "                     '",cs_keyword_TOPIND,"') = 0 then null   ",
                 "       else trim(eb.dzeb027) end                 dzeb027,     ",    
                 "       'N'                                       dzeb028,     ",    
                 "       ev.dzev020                                dzeb029,     ",    
                 "       ev.dzev022                                dzeb030,     ",    
                 "       eb.dzeb031                                dzeb031      ",    
                 "  from dzev_t ev                                              ",    
                 "  left outer join dzeb_t eb on eb.dzeb001 = lower(ev.dzev002) ",    
                 "                           and eb.dzeb002 = lower(ev.dzev005) ",    
                 " where ev.dzev002 = '",ls_TableName.toUpperCase(),"'          ",    
                 "   and ev.dzev003 = trim('",lo_parameters.DZLM015,"')         ",    
                 "   and ev.dzev004 = 'TableDesigner'                           ",    
                 "   and ev.dzev018 = '",lo_parameters.DZLM005,"'               ",    
                 "   and ev.dzev019 = '",lo_parameters.DZLM012,"'               ",    
                 " order by ev.dzev011                                          "
                 
    #取得Table export notice             
    LET ls_tsn_SQL = "select 'DZEB_T' table_name,count(1) records          ",
                     "  from dzev_t ev                                     ",
                     " where ev.dzev002 = '",ls_TableName.toUpperCase(),"' ",
                     "   and ev.dzev003 = trim('",lo_parameters.DZLM015,"')", 
                     "   and ev.dzev004 = 'TableDesigner'                  ",
                     "   and ev.dzev018 = '",lo_parameters.DZLM005,"'      ",
                     "   and ev.dzev019 = '",lo_parameters.DZLM012,"'      "
                     
    &else
    LET ls_SQL = "SELECT LOWER(EV.DZEV002)                         DZEB001,        ",
                 "       LOWER(EV.DZEV005)                         DZEB002,        ",
                 "       EV.DZEV010                                DZEB003,        ",
                 "       EV.DZEV006                                DZEB004,        ",
                 "       DECODE(EV.DZEV007,'Y','N','N','Y','Y')    DZEB005,        ",
                 "       EV.DZEV013                                DZEB006,        ",
                 "       LOWER(EV.DZEV008)                         DZEB007,        ",
                 "       DECODE(EV.DZEV008,'DATE',NULL,EV.DZEV009) DZEB008,        ",
                 "       EV.DZEV021                                DZEB012,        ", 
                 "       EV.DZEV014                                DZEB021,        ",
                 "       EV.DZEV015                                DZEB022,        ",
                 "       EV.DZEV016                                DZEB023,        ",
                 "       EV.DZEV017                                DZEB024,        ",
                 "       DECODE(                                                   ", #20190304 begin
                 "               INSTRB(TRIM(EB.DZEB027),'",cs_keyword_TOPIND,"'), ",
                 "               0,NULL,                                           ",
                 "               TRIM(EB.DZEB027)                                  ",  
                 "              )                                  DZEB027,        ", #20190304 end
                 "       'N'                                       DZEB028,        ",                         
                 "       EV.DZEV020                                DZEB029,        ",                         
                 "       EV.DZEV022                                DZEB030,        ",                         
                 "       EB.DZEB031                                DZEB031         ",                         
                 "  FROM DZEV_T EV                                                 ",
                 "  LEFT OUTER JOIN DZEB_T EB ON EB.DZEB001 = LOWER(EV.DZEV002)    ",
                 "                           AND EB.DZEB002 = LOWER(EV.DZEV005)    ",
                 " WHERE EV.DZEV002 = '",ls_TableName.toUpperCase(),"'             ",
                 "   AND EV.DZEV003 = TRIM('",lo_parameters.DZLM015,"')            ", 
                 "   AND EV.DZEV004 = 'TableDesigner'                              ",
                 "   AND EV.DZEV018 = '",lo_parameters.DZLM005,"'                  ",
                 "   AND EV.DZEV019 = '",lo_parameters.DZLM012,"'                  ",
                 " ORDER BY EV.DZEV011                                             "

    #取得Table export notice             
    LET ls_tsn_SQL = "SELECT 'DZEB_T' TABLE_NAME,COUNT(1) RECORDS          ",
                     "  FROM DZEV_T EV                                     ",
                     " WHERE EV.DZEV002 = '",ls_TableName.toUpperCase(),"' ",
                     "   AND EV.DZEV003 = TRIM('",lo_parameters.DZLM015,"')", 
                     "   AND EV.DZEV004 = 'TableDesigner'                  ",
                     "   AND EV.DZEV018 = '",lo_parameters.DZLM005,"'      ",
                     "   AND EV.DZEV019 = '",lo_parameters.DZLM012,"'      "
    &endif
    &endif
  END IF

  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)    
  
  PREPARE lpre_GetDZEBData FROM ls_SQL
  DECLARE lcur_GetDZEBData CURSOR FOR lpre_GetDZEBData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZEBData
  FOREACH lcur_GetDZEBData INTO lo_DZEB_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZEBData
  CALL lo_DZEB_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZEBData
  FREE lcur_GetDZEBData  

  LET lo_RETURN = lo_DZEB_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_delete_DZEB_data(p_TableName,p_same_dgenv,p_same_topind) #20190304
DEFINE
  p_TableName    STRING,
  p_same_dgenv   BOOLEAN,
  p_same_topind  BOOLEAN #20190304
DEFINE
  ls_TableName    STRING,
  lb_same_dgenv   BOOLEAN, 
  lb_same_topind  BOOLEAN, #20190304
  ls_DeleteSQL    STRING, 
  ls_sql_cond     STRING,
  lb_RT_DEL_EXTRA_STD_COLS BOOLEAN, #20191205
  lb_success      BOOLEAN 

  LET ls_TableName   = p_TableName.toLowerCase()
  LET lb_same_dgenv  = p_same_dgenv
  LET lb_same_topind = p_same_topind #20190304

  LET lb_RT_DEL_EXTRA_STD_COLS = IIF(NVL(FGL_GETENV(cs_env_RT_DEL_EXTRA_STD_COLS),"N")=="N",FALSE,TRUE) #20191205

  #20171120 begin
  IF mb_pickup_exist THEN 
    LET lb_success = TRUE
  ELSE 
  #20171120 end
    -- 如果是相同 DGENV (s to s or c to c)
    IF lb_same_dgenv THEN --AND NOT mb_dgenv_is_cust THEN #20170524 #20170523
      -- 20190304 begin
      -- 如果是相同 TOPIND
      IF lb_same_topind THEN 
        LET ls_sql_cond = ""
      ELSE
        --LET ls_sql_cond = " and replace(coalesce(trim(dzeb027),'",cs_col_type_null,"'),'\"','') not like '%COL_TYPE:",ms_topind,"%' " 
      LET ls_sql_cond = " and replace(coalesce(trim(dzeb027),'",cs_col_type_null,"'),'\"','') like '%",cs_col_type_null,"%' " 
      END IF
      -- 20190304 end
      CALL sadzi140_exim_exec_delete_dzeb_data(ls_TableName,ls_sql_cond) RETURNING lb_success #20191205
    ELSE 
    -- 如果是不同 DGENV (s to c)
      #20191205 begin
      IF mb_dgenv_is_cust AND (NOT lb_RT_DEL_EXTRA_STD_COLS) THEN #標準到客制一律不刪欄位,除非有下 RT_DEL_EXTRA_STD_COLS=Y
        LET lb_success = TRUE 
      ELSE
      #20191205 end
        -- 20190304 begin
        IF mo_src_pkg_sys_info.psi_TOPIND <> cs_industry_type_standard THEN
          -- 從行業別主機過來只刪行業別欄位
          LET ls_sql_cond = " and coalesce(trim(dzeb029),'",cs_dgenv_customize,"') = '",cs_dgenv_standard,"'", #從嚴控管 dzeb029 若為 null, 則當成是客制
                            " and replace(coalesce(trim(dzeb027),'",cs_col_type_null,"'),'\"','') like '%COL_TYPE:",mo_src_pkg_sys_info.psi_TOPIND,"%' " 
        ELSE
          -- 從標準主機過來只刪標準欄位
          LET ls_sql_cond = " and coalesce(trim(dzeb029),'",cs_dgenv_customize,"') = '",cs_dgenv_standard,"'", #從嚴控管 dzeb029 若為 null, 則當成是客制
                            " and replace(coalesce(trim(dzeb027),'",cs_col_type_null,"'),'\"','') like '%",cs_col_type_null,"%' " 
        END IF  
        -- 20190304 end
        CALL sadzi140_exim_exec_delete_dzeb_data(ls_TableName,ls_sql_cond) RETURNING lb_success #20191205
      END IF #20191205  
    END IF 
    
  END IF #20171120 

  RETURN lb_success
  
END FUNCTION

#20191205 begin
FUNCTION sadzi140_exim_exec_delete_dzeb_data(p_table_name,p_sql_cond)
DEFINE
  p_table_name  STRING,
  p_sql_cond    STRING
DEFINE
  ls_table_name  STRING,
  ls_sql_cond    STRING,
  ls_DeleteSQL   STRING,
  lb_success     BOOLEAN
DEFINE
  lb_return  BOOLEAN  

  LET ls_table_name = p_table_name
  LET ls_sql_cond   = p_sql_cond

  LET lb_return = FALSE

  LET ls_DeleteSQL = "delete from dzeb_t                 ",
                     " where dzeb001 = '",ls_table_name,"'",
                     ls_sql_cond

  CALL sadzi140_exim_exec_SQL(ls_DeleteSQL) RETURNING lb_success
  LET lb_return = lb_success
  
  RETURN lb_return
  
END FUNCTION
#20191205 end

FUNCTION sadzi140_exim_extract_DZEB_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 
DEFINE
  lb_return  BOOLEAN,
  lo_return  DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH   

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEB_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success

  LET lb_return = lb_success
  LET lo_return = lo_DATAS
  
  RETURN lb_return,lo_return
  
END FUNCTION

FUNCTION sadzi140_exim_import_DZEB_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_result           BOOLEAN, #20200330
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEB_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZEB_T_data(lo_DATAS) RETURNING lb_success
      CALL sadzi140_exim_set_DZEB_T_datatype(ls_TableName) RETURNING lb_result #20200330
    END IF   
  END IF  

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZEB_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  lb_result       BOOLEAN, #20171120
  ls_user         VARCHAR(100),
  ls_orig_coltype STRING,
  ls_column_type  STRING,
  li_sequence     INTEGER,
  ls_sequence     STRING,
  ls_ship_notice  STRING,
  ls_IMPORT_TABLE STRING, #20180302
  lo_property     T_PROPERTY,
  ldt_datetime    DATETIME YEAR TO SECOND,
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  LET ls_IMPORT_TABLE = "DZEB_T" #20180302
  
  &ifndef DEBUG
  LET ls_user = g_user
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  &endif
  LET ldt_datetime = CURRENT YEAR TO SECOND
  
  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN
    
      #20171120 begin  
      IF mb_pickup_exist THEN 
        CALL sadzi140_exim_check_if_exist_in_pick_list(mo_pick_col_list,lo_DATAS[li_loop,2]) RETURNING lb_result
        IF NOT lb_result THEN CONTINUE FOR END IF
      END IF
      #20171120 end

      #20180302 begin  
      IF mo_local_sys_info.psi_RT_PATCH_MODE <> "Y" THEN 
        CALL sadzi140_exim_check_if_exist_in_deleted_std_col_list(mo_deleted_std_col_list,lo_DATAS[li_loop,2]) RETURNING lb_result
        IF lb_result THEN 
          DISPLAY cs_warning_tag,"Column ",lo_DATAS[li_loop,2]," exists in deleted standard column list, skip import to ",ls_IMPORT_TABLE,".","\n",
                                 "You can using 'Tools'->'Field selection import' to import it !" 
          CONTINUE FOR 
        END IF
      END IF  
      #20180302 end
      
      #相同的ZONE就用相同的序號 
      IF (mb_same_dgenv OR mb_new_table) AND mb_same_topind AND (NOT mb_pickup_exist) THEN #20171120 #20190304
        LET ls_sequence = lo_DATAS[li_loop,10]
      ELSE
        #取得Sequnece ID  
        CALL sadzi140_exim_get_column_sequence(lo_DATAS[li_loop,1],lo_DATAS[li_loop,2]) RETURNING li_sequence
        LET ls_sequence = li_sequence
        LET ls_sequence = ls_sequence.trim()
      END IF 

      #取得欄位型態, 當舊型態和新型態不同時才重取對應的資料型態值(已出貨欄位才做判斷)
      LET ls_ship_notice = lo_DATAS[li_loop,18]
      IF ls_ship_notice = cs_column_shipped THEN
        LET ls_orig_coltype = lo_DATAS[li_loop,6]
        CALL sadzi140_exim_get_column_type(lo_DATAS[li_loop,1],lo_DATAS[li_loop,2],ls_orig_coltype) RETURNING ls_column_type
        IF ls_column_type <> ls_orig_coltype THEN
          CALL sadzi140_util_get_property(ls_column_type,ms_dlang) RETURNING lo_property.*
          LET lo_DATAS[li_loop,6] = ls_column_type  
          LET lo_DATAS[li_loop,7] = lo_property.d_type
          LET lo_DATAS[li_loop,8] = lo_property.d_length
        END IF
      END IF  
      
      LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "'",lo_DATAS[li_loop,3],"',",
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",lo_DATAS[li_loop,5],"',",
                           "'",lo_DATAS[li_loop,6],"',",
                           "'",lo_DATAS[li_loop,7],"',",
                           "'",lo_DATAS[li_loop,8],"',",
                           --"'",sadzi140_exim_trim(lo_DATAS[li_loop,9]),"',",
                           "",IIF(lo_DATAS[li_loop,9] IS NULL,"NULL","'"||lo_DATAS[li_loop,9]||"'"),",",
                           "",NVL(ls_sequence,"NULL"),",",
                           "'",lo_DATAS[li_loop,11],"',",
                           "",NVL(lo_DATAS[li_loop,12],"NULL"),",",
                           "'",lo_DATAS[li_loop,13],"',",
                           "'",lo_DATAS[li_loop,14],"',",
                           "'",lo_DATAS[li_loop,15],"',",
                           "'",lo_DATAS[li_loop,16],"',",
                           "'",lo_DATAS[li_loop,17],"',",
                           "'",lo_DATAS[li_loop,18],"',",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,",",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,""
                                                    
      LET ls_InsertSQL = "Insert into dzeb_t (                                         ",
                         "                    dzeb001,dzeb002,dzeb003,dzeb004,dzeb005, ",
                         "                    dzeb006,dzeb007,dzeb008,dzeb012,dzeb021, ",
                         "                    dzeb022,dzeb023,dzeb024,dzeb027,dzeb028, ",
                         "                    dzeb029,dzeb030,dzeb031,                 ",
                         "                    dzebcrtid,dzebcrtdt,dzebmodid,dzebmoddt  ",
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "

      #DZEB029 若已標示為為客制標示, 則不更新                   
      LET ls_UpdateValue = "dzeb003 = '",lo_DATAS[li_loop,3],"',",
                           "dzeb004 = '",lo_DATAS[li_loop,4],"',",
                           "dzeb005 = '",lo_DATAS[li_loop,5],"',",
                           "dzeb006 = '",lo_DATAS[li_loop,6],"',",
                           "dzeb007 = '",lo_DATAS[li_loop,7],"',",
                           "dzeb008 = '",lo_DATAS[li_loop,8],"',",
                           --"dzeb012 = '",lo_DATAS[li_loop,9],"',",
                           "dzeb012 = ",IIF(lo_DATAS[li_loop,9] IS NULL,"NULL","'"||lo_DATAS[li_loop,9]||"'"),",",
                           "dzeb021 = ",NVL(ls_sequence,"NULL"),",",  #"DZEB021 = ",NVL(lo_DATAS[li_loop,10],"NULL"),",",
                           "dzeb022 = '",lo_DATAS[li_loop,11],"',",
                           "dzeb023 = ",NVL(lo_DATAS[li_loop,12],"NULL"),",",
                           "dzeb024 = '",lo_DATAS[li_loop,13],"',",
                           "dzeb027 = '",lo_DATAS[li_loop,14],"',",
                           "dzeb028 = '",lo_DATAS[li_loop,15],"',",
                           "dzeb029 = case dzeb029 when '",cs_dgenv_customize,"' then dzeb029 else '",lo_DATAS[li_loop,16],"' end,",
                           "dzeb030 = '",lo_DATAS[li_loop,17],"',",
                           "dzeb031 = '",lo_DATAS[li_loop,18],"',",
                           "dzebmodid = trim('",ls_user,"'),",
                           "dzebmoddt = ",cs_get_time_stamp,""
                                                    
      LET ls_UpdateSQL = "update dzeb_t set ",
                         ls_UpdateValue,
                         " where dzeb001 = '",lo_DATAS[li_loop,1],"'",
                         "   and dzeb002 = '",lo_DATAS[li_loop,2],"'"
                         --"   and coalesce(dzeb029,'",cs_dgenv_standard,"') = '",cs_dgenv_standard,"'"
                         
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
    END IF
  END FOR 

  RETURN lb_success
  
END FUNCTION

#20200330 begin
FUNCTION sadzi140_exim_set_DZEB_T_datatype(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name   STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN
  
  LET ls_table_name = p_table_name
  
  LET lb_success = TRUE

  &ifdef is_msv
  LET ls_UpdateSQL = "update dzeb_t                                      ",
                     "   set dzeb007 = t2.gztd003,                       ",
                     "       dzeb008 = t2.gztd008                        ",
                     "  from (                                           ",
                     "         select gztd001,gztd003,gztd008            ",
                     "           from gztd_t                             ", 
                     "       ) t2                                        ", 
                     " where dzeb006 = t2.gztd001                        ",
                     "   and dzeb001 = '",ls_table_name.toLowerCase(),"' "
  &else
  &ifdef is_pgs
  LET ls_UpdateSQL = "update dzeb_t                                      ",
                     "   set dzeb007 = t2.gztd003,                       ",
                     "       dzeb008 = t2.gztd008                        ",
                     "  from (                                           ",
                     "         select gztd001,gztd003,gztd008            ",
                     "           from gztd_t                             ", 
                     "       ) t2                                        ", 
                     " where dzeb006 = t2.gztd001                        ",
                     "   and dzeb001 = '",ls_table_name.toLowerCase(),"' "
  &else
  LET ls_UpdateSQL = "update dzeb_t                                               ", 
                     "    set (dzeb007,dzeb008) = (                               ",
                     "                              select td.gztd003,td.gztd008  ",
                     "                                from gztd_t td              ",
                     "                               where dzeb006 = td.gztd001   ",
                     "                            )                               ",
                     "  where dzeb001 = '",ls_table_name.toLowerCase(),"'         "
  &endif
  &endif
  
  CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
  IF NOT lb_success THEN
    DISPLAY "[ UPDATE SQL ] ",ls_UpdateSQL,"\n" 
  END IF

  RETURN lb_success
  
END FUNCTION
#20200330 end
## DZEC_T #####################################################################

FUNCTION sadzi140_exim_export_DZEC_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEC_T"
  
  CALL sadzi140_exim_get_DZEC_data(ls_TableName,p_json_obj,lo_PARAMETERS.*) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_DZEC_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName   STRING,
  p_json_obj    util.JSONObject,
  po_PARAMETERS T_DZLM_T_SCR  
DEFINE
  ls_TableName   STRING,
  lo_PARAMETERS  T_DZLM_T_SCR,
  ls_MaxVersion  STRING,
  lo_DZEC_T      DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*
  
  INITIALIZE lo_DZEC_T TO NULL  #20180213
  INITIALIZE lo_RETURN TO NULL  #20180213

  IF (lo_parameters.DZLM005 = cs_NoVersionCharacter) OR 
     (NOT sadzi140_exim_check_version_data_exist(ls_TableName)) OR 
     (NOT sadzi140_exim_check_alm_request_data_exist(ls_TableName,lo_PARAMETERS.*))THEN
     
    LET ls_SQL = "select ec.dzec001,ec.dzec002,ec.dzec003,ec.dzec004,ec.dzec005,",
                 "       ec.dzec006,ec.dzec007,ec.dzec008,ec.dzecstus           ", #20170912 add dzecstus  
                 "  from dzec_t ec                                              ",
                 " where ec.dzec001 = '",ls_TableName,"'                        "
    #取得Table export notice             
    LET ls_tsn_SQL = "select 'DZEC_T' table_name,count(1) records  ",
                     "  from dzec_t ec                              ",
                     " where ec.dzec001 = '",ls_TableName,"'        "

  ELSE
  
    LET ls_SQL = "select ew.dzew001  dzec001,                                   ",
                 "       ew.dzew004  dzec002,                                   ",
                 "       ew.dzew005  dzec003,                                   ",
                 "       ew.dzew006  dzec004,                                   ",
                 "       'n'         dzec005,                                   ",
                 "       ew.dzew012  dzec006,                                   ", 
                 "       ew.dzew013  dzec007,                                   ", 
                 #"       ew.dzew014  dzec008                                   ", 
                 "       ec.dzec008  dzec008,                                   ", 
                 "       ew.dzewstus dzecstus                                   ", #20170912 
                 "  from dzew_t ew                                              ",
                 "  left outer join dzec_t ec on ec.dzec001 = lower(ew.dzew001) ", 
                 "                           and ec.dzec002 = lower(ew.dzew004) ",
                 " where ew.dzew001 = '",ls_TableName.toLowerCase(),"'          ",
                 "   and ew.dzew002 = trim('",lo_PARAMETERS.DZLM015,"')         ",
                 "   and ew.dzew003 = 'Index'                                   ",
                 "   and ew.dzew010 = '",lo_PARAMETERS.DZLM005,"'               ",
                 "   and ew.dzew011 = '",lo_PARAMETERS.DZLM012,"'               "

    LET ls_tsn_SQL = "select 'DZEC_T' table_name,count(1) records          ",
                     "  from dzew_t ew                                     ",
                     " where ew.dzew001 = '",ls_TableName.toLowerCase(),"' ",
                     "   and ew.dzew002 = trim('",lo_PARAMETERS.DZLM015,"')",
                     "   and ew.dzew003 = 'Index'                          ",
                     "   and ew.dzew010 = '",lo_PARAMETERS.DZLM005,"'      ",
                     "   and ew.dzew011 = '",lo_PARAMETERS.DZLM012,"'      "
  END IF

  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)    
  
  PREPARE lpre_GetDZECData FROM ls_SQL
  DECLARE lcur_GetDZECData CURSOR FOR lpre_GetDZECData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZECData
  FOREACH lcur_GetDZECData INTO lo_DZEC_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZECData
  CALL lo_DZEC_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZECData
  FREE lcur_GetDZECData  

  LET lo_RETURN = lo_DZEC_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_delete_DZEC_data(p_TableName,p_same_dgenv,p_same_topind) #20190304
DEFINE
  p_TableName    STRING,
  p_same_dgenv   BOOLEAN,
  p_same_topind  BOOLEAN #20190304
DEFINE
  ls_TableName   STRING,
  lb_same_dgenv  BOOLEAN,
  lb_same_topind BOOLEAN, #20190304 
  ls_DeleteSQL   STRING, 
  ls_sql_cond    STRING,
  lb_success     BOOLEAN 

  LET ls_TableName   = p_TableName.toLowerCase()
  LET lb_same_dgenv  = p_same_dgenv
  LET lb_same_topind = p_same_topind #20190304

  #20171120 begin
  IF mb_pickup_exist THEN 
    LET lb_success = TRUE
  ELSE 
  #20171120 end
    IF lb_same_dgenv THEN --AND NOT mb_dgenv_is_cust THEN #20170524 #20170523
      LET ls_sql_cond = ""
    ELSE
      LET ls_sql_cond = " and coalesce(dzec006,'",cs_dgenv_standard,"') = '",cs_dgenv_standard,"'"
    END IF
    
    LET ls_DeleteSQL = "delete from dzec_t                               ",
                       " where dzec001 = '",ls_TableName.toLowerCase(),"'",
                       ls_sql_cond
                       
    #DISPLAY ls_DeleteSQL
    
    CALL sadzi140_exim_exec_SQL(ls_DeleteSQL) RETURNING lb_success
  END IF #20171120  

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_import_DZEC_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEC_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZEC_T_data(lo_DATAS) RETURNING lb_success
    END IF   
  END IF  

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZEC_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  ls_user         VARCHAR(100),
  li_idx_cnt      INTEGER, #20170912
  ls_txt_idx_list STRING, #20170912
  ls_arr_idx_list STRING, #20170912
  lo_resort_cols  DYNAMIC ARRAY OF T_RESORT_COLUMN, #20170912
  lo_arr_cols     T_RESORT_COLUMN, #20170912
  lo_txt_cols     T_RESORT_COLUMN, #20170912
  ldt_datetime    DATETIME YEAR TO SECOND,
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  &ifndef DEBUG
  LET ls_user = g_user
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  &endif
  LET ldt_datetime = CURRENT YEAR TO SECOND

  #20170912 begin
  CALL sadzi140_db_get_table_index_list(lo_DATAS[1,1],TRUE) RETURNING lo_resort_cols
  #20170912 end
  
  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN

      #20170912 begin
      #比對匯入的索引的欄位清單是否有重複, 若有, 則匯入的資料設為失效
      IF lo_resort_cols.getLength() > 0 THEN 
        LET lo_txt_cols.rc_TABLE_NAME  = lo_DATAS[li_loop,1]
        LET lo_txt_cols.rc_INDEX_NAME  = lo_DATAS[li_loop,2]
        LET lo_txt_cols.rc_COLUMN_LIST = lo_DATAS[li_loop,4]
        CALL sadzi140_db_get_resort_index_columns_list(lo_txt_cols.*,cs_resort_index_type_text) RETURNING ls_txt_idx_list
        FOR li_idx_cnt = 1 TO lo_resort_cols.getLength()
          CALL sadzi140_db_get_resort_index_columns_list(lo_resort_cols[li_idx_cnt].*,cs_resort_index_type_text) RETURNING ls_arr_idx_list
          IF ls_arr_idx_list = ls_txt_idx_list THEN
            LET lo_DATAS[li_loop,9] = "N"
            DISPLAY cs_message_tag,"Import index ",lo_DATAS[li_loop,1],".",lo_DATAS[li_loop,2]," column list [",lo_DATAS[li_loop,4],"] duplicate with customized index, set to disabled."
            EXIT FOR 
          END IF
        END FOR
      END IF
      #20170912 end
      
      LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "'",lo_DATAS[li_loop,3],"',",
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",lo_DATAS[li_loop,5],"',",
                           "'",lo_DATAS[li_loop,6],"',",
                           "'",lo_DATAS[li_loop,7],"',",
                           "'",lo_DATAS[li_loop,8],"',",
                           "'",NVL(lo_DATAS[li_loop,9],"Y"),"',", #20170912
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,",",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,""
                                                    
      LET ls_InsertSQL = "insert into dzec_t (                                        ",
                         "                    dzec001,dzec002,dzec003,dzec004,dzec005,",
                         "                    dzec006,dzec007,dzec008,dzecstus,       ", #20170912 add dzecstus
                         "                    dzeccrtid,dzeccrtdt,dzecmodid,dzecmoddt ",   
                         "                   )                                        ",  
                         " values (                                                   ",
                         ls_InsertValue,
                         "        )                                                   "
                         
      #DZEC006,DZEC007 為客制標示, 不更新                   
      LET ls_UpdateValue = "dzec003 = '",lo_DATAS[li_loop,3],"',",
                           "dzec004 = '",lo_DATAS[li_loop,4],"',",
                           "dzec005 = '",lo_DATAS[li_loop,5],"',",
                           #"dzec006 = '",lo_DATAS[li_loop,6],"',",
                           #"dzec007 = '",lo_DATAS[li_loop,7],"',",
                           "dzec008 = '",lo_DATAS[li_loop,8],"',",
                           "dzecstus = '",NVL(lo_DATAS[li_loop,9],"Y"),"',", #20170912
                           "dzecmodid = trim('",ls_user,"'),",
                           "dzecmoddt = ",cs_get_time_stamp,""
                                                    
      LET ls_UpdateSQL = "update dzec_t set ",
                         ls_UpdateValue,
                         " where dzec001 = '",lo_DATAS[li_loop,1],"'",
                         "   and dzec002 = '",lo_DATAS[li_loop,2],"'"
                         --"   AND NVL(DZEC006,'",cs_dgenv_standard,"') = '",cs_dgenv_standard,"'"
                         
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
    END IF    
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZED_T #####################################################################

FUNCTION sadzi140_exim_export_DZED_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZED_T"
  
  CALL sadzi140_exim_get_DZED_data(ls_TableName,p_json_obj,lo_PARAMETERS.*) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_DZED_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName   STRING,
  p_json_obj    util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName   STRING,
  lo_PARAMETERS  T_DZLM_T_SCR,
  lo_REVISION    T_REVISION,
  ls_MaxVersion  STRING,
  lo_DZED_T      DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  INITIALIZE lo_DZED_T TO NULL #20180213
  INITIALIZE lo_RETURN TO NULL #20180213

  IF (lo_parameters.DZLM005 = cs_NoVersionCharacter) OR 
     (NOT sadzi140_exim_check_version_data_exist(ls_TableName)) OR 
     (NOT sadzi140_exim_check_alm_request_data_exist(ls_TableName,lo_PARAMETERS.*))THEN

    -- F: Footing data, R:Referencing Key, P:Primary Key 
    -- Footint Data 沒有版次的架構, 所以要另外匯出
    LET ls_SQL = "select ed.dzed001,ed.dzed002,ed.dzed003,ed.dzed004,ed.dzed005,",
                 "       ed.dzed006,ed.dzed007,ed.dzed008,ed.dzed009,ed.dzed010 ",
                 "  from dzed_t ed                                              ",
                 " where ed.dzed001 = '",ls_TableName,"'                        ",
                 "   and ed.dzed003 <> 'F'                                      ",
                 "union                                                         ",
                 "select ed.dzed001,ed.dzed002,ed.dzed003,ed.dzed004,ed.dzed005,",
                 "       ed.dzed006,ed.dzed007,ed.dzed008,ed.dzed009,ed.dzed010 ",
                 "  from dzed_t ed                                              ",
                 " where ed.dzed001 = '",ls_TableName,"'                        ",
                 "   and ed.dzed003 = 'F'                                       "
                 
    -- 取得Table export notice             
    LET ls_tsn_SQL = "select 'DZED_T' table_name,count(1) records     ",
                     "  from (                                        ",
                     "        select *                                ",
                     "          from dzed_t ed                        ",
                     "         where ed.dzed001 = '",ls_TableName,"'  ",
                     "           and ed.dzed003 <> 'F'                ",
                     "        union                                   ",
                     "        select *                                ",
                     "          from dzed_t ed                        ",
                     "         where ed.dzed001 = '",ls_TableName,"'  ",
                     "           and ed.dzed003 = 'F'                 ",
                     "       ) ed                                     "
  ELSE
    -- Footing data 要另外 Union
    LET ls_SQL = "select ew.dzew001 dzed001,                                    ",
                 "       ew.dzew004 dzed002,                                    ",
                 "       ew.dzew005 dzed003,                                    ",
                 "       ew.dzew006 dzed004,                                    ",
                 "       ew.dzew007 dzed005,                                    ",
                 "       ew.dzew008 dzed006,                                    ",
                 "       'N'        dzed007,                                    ",
                 "       ew.dzew012 dzed008,                                    ",                 
                 "       ew.dzew013 dzed009,                                    ",                 
                 #"       ew.dzew014 dzed010                                     ",                 
                 "       ed.dzed010 dzed010                                     ",                 
                 "  from dzew_t ew                                              ",
                 "  left outer join dzed_t ed on ed.dzed001 = lower(ew.dzew001) ", 
                 "                           and ed.dzed002 = lower(ew.dzew004) ",
                 "where ew.dzew001 = '",ls_TableName.toLowerCase(),"'           ",
                 "  and ew.dzew002 = trim('",lo_PARAMETERS.DZLM015,"')          ",
                 "  and ew.dzew003 = 'Constraint'                               ",
                 "  and ew.dzew010 = '",lo_PARAMETERS.DZLM005,"'                ",
                 "  and ew.dzew011 = '",lo_PARAMETERS.DZLM012,"'                ",
                 "union                                                         ",
                 "select ed.dzed001,ed.dzed002,ed.dzed003,ed.dzed004,ed.dzed005,",
                 "       ed.dzed006,ed.dzed007,ed.dzed008,ed.dzed009,ed.dzed010 ",
                 "  from dzed_t ed                                              ",
                 " where ed.dzed001 = '",ls_TableName,"'                        ",
                 "   and ed.dzed003 = 'F'                                       "
                 
    -- 取得Table export notice     
    -- Footing data count 要另外 Union    
    LET ls_tsn_SQL = "select 'DZED_T' table_name,count(1) records                           ",
                     "  from (                                                              ",
                     "        select ew.dzew001 dzed001,                                    ",
                     "               ew.dzew004 dzed002,                                    ",
                     "               ew.dzew005 dzed003,                                    ",
                     "               ew.dzew006 dzed004,                                    ",
                     "               ew.dzew007 dzed005,                                    ",
                     "               ew.dzew008 dzed006,                                    ",
                     "               'N'        dzed007,                                    ",
                     "               ew.dzew012 dzed008,                                    ",                 
                     "               ew.dzew013 dzed009,                                    ",                 
                     "               ew.dzew014 dzed010                                     ",
                     "          from dzew_t ew                                              ",
                     "         where ew.dzew001 = '",ls_TableName.toLowerCase(),"'          ",
                     "           and ew.dzew002 = trim('",lo_PARAMETERS.DZLM015,"')         ",
                     "           and ew.dzew003 = 'Constraint'                              ",
                     "           and ew.dzew010 = '",lo_PARAMETERS.DZLM005,"'               ",
                     "           and ew.dzew011 = '",lo_PARAMETERS.DZLM012,"'               ",
                     "        union                                                         ",
                     "        select ed.dzed001,ed.dzed002,ed.dzed003,ed.dzed004,ed.dzed005,",
                     "               ed.dzed006,ed.dzed007,ed.dzed008,ed.dzed009,ed.dzed010 ",
                     "          from dzed_t ed                                              ",
                     "         where ed.dzed001 = '",ls_TableName.toLowerCase(),"'          ",
                     "           and ed.dzed003 = 'F'                                       ",
                     "       ) ed                                                           "
  END IF

  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)    
  
  PREPARE lpre_GetDZEDData FROM ls_SQL
  DECLARE lcur_GetDZEDData CURSOR FOR lpre_GetDZEDData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZEDData
  FOREACH lcur_GetDZEDData INTO lo_DZED_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZEDData
  CALL lo_DZED_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZEDData
  FREE lcur_GetDZEDData  

  LET lo_RETURN = lo_DZED_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_delete_DZED_data(p_TableName,p_same_dgenv,p_same_topind) #20190304
DEFINE
  p_TableName    STRING,
  p_same_dgenv   BOOLEAN,
  p_same_topind  BOOLEAN #20190304
DEFINE
  ls_TableName   STRING,
  lb_same_dgenv  BOOLEAN,
  lb_same_topind BOOLEAN, #20190304 
  ls_DeleteSQL   STRING, 
  ls_sql_cond    STRING,
  lb_success     BOOLEAN 

  LET ls_TableName   = p_TableName.toLowerCase()
  LET lb_same_dgenv  = p_same_dgenv
  LET lb_same_topind = p_same_topind #20190304
  
  #20171120 begin
  IF mb_pickup_exist THEN 
    LET lb_success = TRUE
  ELSE 
  #20171120 end
    IF lb_same_dgenv THEN --AND NOT mb_dgenv_is_cust THEN #20170524 #20170523
      LET ls_sql_cond = ""
    ELSE
      LET ls_sql_cond = " and nvl(dzed008,'",cs_dgenv_standard,"') = '",cs_dgenv_standard,"'"
    END IF
    
    LET ls_DeleteSQL = "delete from dzed_t                               ",
                       " where dzed001 = '",ls_TableName.toLowerCase(),"'",
                       ls_sql_cond
                       
    #DISPLAY ls_DeleteSQL
    
    CALL sadzi140_exim_exec_SQL(ls_DeleteSQL) RETURNING lb_success
  END IF #20171120  

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_import_DZED_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZED_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZED_T_data(lo_DATAS) RETURNING lb_success
    END IF   
  END IF  

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZED_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  ls_user         VARCHAR(100),
  ldt_datetime    DATETIME YEAR TO SECOND,
  lb_result       BOOLEAN, #20171120
  lb_success      BOOLEAN,
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  &ifndef DEBUG
  LET ls_user = g_user
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  &endif
  LET ldt_datetime = CURRENT YEAR TO SECOND
  
  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN
      
      LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "'",lo_DATAS[li_loop,3],"',",
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",sadzi140_exim_trim(lo_DATAS[li_loop,5]),"',",
                           "'",sadzi140_exim_trim(lo_DATAS[li_loop,6]),"',",
                           "'",sadzi140_exim_trim(lo_DATAS[li_loop,7]),"',",
                           "'",sadzi140_exim_trim(lo_DATAS[li_loop,8]),"',",
                           "'",sadzi140_exim_trim(lo_DATAS[li_loop,9]),"',",
                           "'",sadzi140_exim_trim(lo_DATAS[li_loop,10]),"',",
                           "'",sadzi140_exim_trim(ls_user),"',",
                           "",cs_get_time_stamp,",",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,""
                           
                                                    
      LET ls_InsertSQL = "insert into dzed_t (                                         ",
                         "                    dzed001,dzed002,dzed003,dzed004,dzed005, ",
                         "                    dzed006,dzed007,dzed008,dzed009,dzed010, ",
                         "                    dzedcrtid,dzedcrtdt,dzedmodid,dzedmoddt  ",
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "
                         
      #DZED008,DZED009 為客制標示, 不更新                   
      LET ls_UpdateValue = "dzed003 = '",lo_DATAS[li_loop,3],"',      ",
                           "dzed004 = '",lo_DATAS[li_loop,4],"',      ",
                           "dzed005 = '",sadzi140_exim_trim(lo_DATAS[li_loop,5]),"',",
                           "dzed006 = '",sadzi140_exim_trim(lo_DATAS[li_loop,6]),"',",
                           "dzed007 = '",sadzi140_exim_trim(lo_DATAS[li_loop,7]),"',",
                           #"dzed008 = '",sadzi140_exim_trim(lo_DATAS[li_loop,8]),"',",
                           #"dzed009 = '",sadzi140_exim_trim(lo_DATAS[li_loop,9]),"',",
                           "dzed010 = '",sadzi140_exim_trim(lo_DATAS[li_loop,10]),"',",
                           "dzedmodid = '",sadzi140_exim_trim(ls_user),"',",
                           "dzedmoddt = ",cs_get_time_stamp,""
                                                    
      LET ls_UpdateSQL = "update dzed_t set ",
                         ls_UpdateValue,
                         " where dzed001 = '",lo_DATAS[li_loop,1],"'",
                         "   and dzed002 = '",lo_DATAS[li_loop,2],"'"
                         --"   and coalesce(dzed008,'",cs_dgenv_standard,"') = '",cs_dgenv_standard,"'"
                         
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
    END IF    
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZEF_T #####################################################################

FUNCTION sadzi140_exim_export_DZEF_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEF_T"
  
  CALL sadzi140_exim_get_DZEF_data(ls_TableName,p_json_obj) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_DZEF_data(p_TableName,p_json_obj)
DEFINE
  p_TableName  STRING,
  p_json_obj   util.JSONObject
DEFINE
  ls_TableName  STRING,
  lo_DZEF_T     DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName = p_TableName

  INITIALIZE lo_DZEF_T TO NULL  #20180213
  INITIALIZE lo_RETURN TO NULL  #20180213
  
  LET ls_SQL = "select ef.dzef001,ef.dzef002,ef.dzef003,ef.dzef004,ef.dzef005, ",
               "       ef.dzef006,ef.dzef007,ef.dzef008,ef.dzef009,ef.dzef010, ",
               "       ef.dzefstus                                             ",
               "  from dzef_t ef                                               ",
               " where ef.dzef001 = '",ls_TableName,"'                         "
               
  #取得Table export notice             
  LET ls_tsn_SQL = "select 'DZEF_T' table_name,count(1) records ",
                   "  from dzef_t ef                            ",
                   " where ef.dzef001 = '",ls_TableName,"'      "
                 
  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)
  
  PREPARE lpre_GetDZEFData FROM ls_SQL
  DECLARE lcur_GetDZEFData CURSOR FOR lpre_GetDZEFData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZEFData
  FOREACH lcur_GetDZEFData INTO lo_DZEF_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZEFData
  CALL lo_DZEF_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZEFData
  FREE lcur_GetDZEFData  

  LET lo_RETURN = lo_DZEF_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_import_DZEF_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEF_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZEF_T_data(lo_DATAS) RETURNING lb_success
    END IF   
  END IF  

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZEF_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_result       BOOLEAN, #20171120
  lb_success      BOOLEAN,
  ls_IMPORT_TABLE STRING, #20180302
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  LET ls_IMPORT_TABLE = "DZEF_T" #20180302
  
  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN
      #20171120 begin  
      IF mb_pickup_exist THEN 
        CALL sadzi140_exim_check_if_exist_in_pick_list(mo_pick_col_list,lo_DATAS[li_loop,2]) RETURNING lb_result
        IF NOT lb_result THEN CONTINUE FOR END IF
      END IF
      #20171120 end

      #20180302 begin  
      IF mo_local_sys_info.psi_RT_PATCH_MODE <> "Y" THEN 
        CALL sadzi140_exim_check_if_exist_in_deleted_std_col_list(mo_deleted_std_col_list,lo_DATAS[li_loop,2]) RETURNING lb_result
        IF lb_result THEN 
          DISPLAY cs_warning_tag,"Column ",lo_DATAS[li_loop,2]," exists in deleted standard column list, skip import to ",ls_IMPORT_TABLE,".","\n",
                                 "You can using 'Tools'->'Field selection import' to import it !" 
          CONTINUE FOR 
        END IF
      END IF  
      #20180302 end
      
      LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "'",lo_DATAS[li_loop,3],"',",
                           "",NVL(lo_DATAS[li_loop,4],"NULL"),",",
                           "'",lo_DATAS[li_loop,5],"',",
                           "'",lo_DATAS[li_loop,6],"',",
                           "'",lo_DATAS[li_loop,7],"',",
                           "'",lo_DATAS[li_loop,8],"',",
                           "'",lo_DATAS[li_loop,9],"',",
                           "",NVL(lo_DATAS[li_loop,10],"NULL"),",",
                           "'",lo_DATAS[li_loop,11],"'"
                                                    
      LET ls_InsertSQL = "insert into dzef_t (                                         ",
                         "                    dzef001,dzef002,dzef003,dzef004,dzef005, ",
                         "                    dzef006,dzef007,dzef008,dzef009,dzef010, ",
                         "                    dzefstus                                 ", 
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "
                         
      LET ls_UpdateValue = "dzef003 = '",lo_DATAS[li_loop,3],"',",
                           "dzef004 = ",NVL(lo_DATAS[li_loop,4],"NULL"),",",
                           "dzef005 = '",lo_DATAS[li_loop,5],"',",
                           "dzef006 = '",lo_DATAS[li_loop,6],"',",
                           "dzef007 = '",lo_DATAS[li_loop,7],"',",
                           "dzef008 = '",lo_DATAS[li_loop,8],"',",
                           "dzef009 = '",lo_DATAS[li_loop,9],"',",
                           "dzef010 = ",NVL(lo_DATAS[li_loop,10],"NULL"),""
                           --"dzefstus = '",lo_DATAS[li_loop,11],"'"
                                                    
      LET ls_UpdateSQL = "update dzef_t set ",
                         ls_UpdateValue,
                         " where dzef001 = '",lo_DATAS[li_loop,1],"'",
                         "   and dzef002 = '",lo_DATAS[li_loop,2],"'"
                           
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
    END IF    
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZEP_T #####################################################################

FUNCTION sadzi140_exim_export_DZEP_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEP_T"
  
  CALL sadzi140_exim_get_DZEP_data(ls_TableName,p_json_obj) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_DZEP_data(p_TableName,p_json_obj)
DEFINE
  p_TableName  STRING,
  p_json_obj   util.JSONObject
DEFINE
  ls_TableName  STRING,
  lo_DZEP_T     DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName = p_TableName

  INITIALIZE lo_DZEP_T TO NULL  #20180213 
  INITIALIZE lo_RETURN TO NULL  #20180213
  
  LET ls_SQL = "select ep.dzep001,ep.dzep002,ep.dzep003,ep.dzep004,ep.dzep005,",
               "       ep.dzep006,ep.dzep007,ep.dzep008,ep.dzep009,ep.dzep010,",
               "       ep.dzep011,ep.dzep012,ep.dzep013,ep.dzep014,ep.dzep015,",
               "       ep.dzep016,ep.dzep017,ep.dzep018,ep.dzep019,ep.dzep020,",
               "       ep.dzep021,ep.dzep022,ep.dzep023,ep.dzep024,ep.dzep025,",
               "       ep.dzep026,ep.dzep027,ep.dzep028,                      ",
               "       ep.dzepstus                                            ", 
               "  from dzep_t ep                                              ",
               " where ep.dzep001 = '",ls_TableName,"'                        "
               
  #取得Table export notice             
  LET ls_tsn_SQL = "select 'DZEP_T' table_name,count(1) records ",
                   "  from dzep_t ep                            ",
                   " where ep.dzep001 = '",ls_TableName,"'      "
                 
  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)
  
  PREPARE lpre_GetDZEPData FROM ls_SQL
  DECLARE lcur_GetDZEPData CURSOR FOR lpre_GetDZEPData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZEPData
  FOREACH lcur_GetDZEPData INTO lo_DZEP_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZEPData
  CALL lo_DZEP_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZEPData
  FREE lcur_GetDZEPData  

  LET lo_RETURN = lo_DZEP_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_import_DZEP_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEP_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZEP_T_data(lo_DATAS) RETURNING lb_success
    END IF   
  END IF  

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZEP_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_result       BOOLEAN, #20171120
  lb_success      BOOLEAN,
  ls_IMPORT_TABLE STRING, #20180302
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  LET ls_IMPORT_TABLE = "DZEP_T" #20180302
  
  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN
      #20171120 begin  
      IF mb_pickup_exist THEN 
        CALL sadzi140_exim_check_if_exist_in_pick_list(mo_pick_col_list,lo_DATAS[li_loop,2]) RETURNING lb_result
        IF NOT lb_result THEN CONTINUE FOR END IF
      END IF
      #20171120 end

      #20180302 begin  
      IF mo_local_sys_info.psi_RT_PATCH_MODE <> "Y" THEN 
        CALL sadzi140_exim_check_if_exist_in_deleted_std_col_list(mo_deleted_std_col_list,lo_DATAS[li_loop,2]) RETURNING lb_result
        IF lb_result THEN 
          DISPLAY cs_warning_tag,"Column ",lo_DATAS[li_loop,2]," exists in deleted standard column list, skip import to ",ls_IMPORT_TABLE,".","\n",
                                 "You can using 'Tools'->'Field selection import' to import it !" 
          CONTINUE FOR 
        END IF
      END IF  
      #20180302 end
      
      LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "'",lo_DATAS[li_loop,3],"',",
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",lo_DATAS[li_loop,5],"',",
                           "'",lo_DATAS[li_loop,6],"',",
                           "'",lo_DATAS[li_loop,7],"',",
                           "'",lo_DATAS[li_loop,8],"',",
                           "'",lo_DATAS[li_loop,9],"',",
                           "'",lo_DATAS[li_loop,10],"',",
                           "'",lo_DATAS[li_loop,11],"',",
                           "'",lo_DATAS[li_loop,12],"',",
                           "",NVL(lo_DATAS[li_loop,13],"NULL"),",",
                           "",NVL(lo_DATAS[li_loop,14],"NULL"),",",
                           "'",lo_DATAS[li_loop,15],"',",
                           "'",lo_DATAS[li_loop,16],"',",
                           "'",lo_DATAS[li_loop,17],"',",
                           "'",lo_DATAS[li_loop,18],"',",
                           "'",lo_DATAS[li_loop,19],"',",
                           "'",lo_DATAS[li_loop,20],"',",
                           "'",lo_DATAS[li_loop,21],"',",
                           "'",lo_DATAS[li_loop,22],"',",
                           "'",lo_DATAS[li_loop,23],"',",
                           "'",lo_DATAS[li_loop,24],"',",
                           "'",lo_DATAS[li_loop,25],"',",
                           "'",lo_DATAS[li_loop,26],"',",
                           "",NVL(lo_DATAS[li_loop,27],"NULL"),",",
                           "'",lo_DATAS[li_loop,28],"',",
                           "'",lo_DATAS[li_loop,29],"'"
                                                    
      LET ls_InsertSQL = "insert into dzep_t (                                         ",
                         "                    dzep001,dzep002,dzep003,dzep004,dzep005, ",
                         "                    dzep006,dzep007,dzep008,dzep009,dzep010, ",
                         "                    dzep011,dzep012,dzep013,dzep014,dzep015, ",
                         "                    dzep016,dzep017,dzep018,dzep019,dzep020, ",
                         "                    dzep021,dzep022,dzep023,dzep024,dzep025, ",
                         "                    dzep026,dzep027,dzep028,                 ",
                         "                    dzepstus                                 ",
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "
                         
      LET ls_UpdateValue = {
                           "dzep003 = '",lo_DATAS[li_loop,3],"',",
                           "dzep004 = '",lo_DATAS[li_loop,4],"',",
                           "dzep005 = '",lo_DATAS[li_loop,5],"',",
                           "dzep006 = '",lo_DATAS[li_loop,6],"',",
                           "dzep007 = '",lo_DATAS[li_loop,7],"',",
                           "dzep008 = '",lo_DATAS[li_loop,8],"',",
                           "dzep009 = '",lo_DATAS[li_loop,9],"',",
                           "dzep010 = '",lo_DATAS[li_loop,10],"',",
                           "dzep011 = '",lo_DATAS[li_loop,11],"',",
                           }
                           "dzep012 = '",lo_DATAS[li_loop,12],"',",
                           {
                           "dzep013 = ",NVL(lo_DATAS[li_loop,13],"NULL"),",",
                           "dzep014 = ",NVL(lo_DATAS[li_loop,14],"NULL"),",",
                           "dzep015 = '",lo_DATAS[li_loop,15],"',",
                           "dzep016 = '",lo_DATAS[li_loop,16],"',",
                           "dzep017 = '",lo_DATAS[li_loop,17],"',",
                           "dzep018 = '",lo_DATAS[li_loop,18],"',",
                           "dzep019 = '",lo_DATAS[li_loop,19],"',",
                           "dzep020 = '",lo_DATAS[li_loop,20],"',",
                           }
                           "dzep021 = '",lo_DATAS[li_loop,21],"'"
                           {
                           "dzep022 = '",lo_DATAS[li_loop,22],"',",
                           "dzep023 = '",lo_DATAS[li_loop,23],"',",
                           "dzep024 = '",lo_DATAS[li_loop,24],"',",
                           "dzep025 = '",lo_DATAS[li_loop,25],"',",
                           "dzep026 = '",lo_DATAS[li_loop,26],"',",
                           "dzep027 = ",NVL(lo_DATAS[li_loop,27],"NULL"),",",
                           "dzep028 = '",lo_DATAS[li_loop,28],"'"
                           }
                           --"dzepstus = '",lo_DATAS[li_loop,29],"'"
                                                    
      LET ls_UpdateSQL = "update dzep_t set ",
                         ls_UpdateValue,
                         " where dzep001 = '",lo_DATAS[li_loop,1],"'",
                         "   and dzep002 = '",lo_DATAS[li_loop,2],"'"
                           
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      --LET lb_success = TRUE
      
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
      
    END IF    
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZEQ_T #####################################################################

FUNCTION sadzi140_exim_export_DZEQ_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEQ_T"
  
  CALL sadzi140_exim_get_DZEQ_data(ls_TableName,p_json_obj) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_DZEQ_data(p_TableName,p_json_obj)
DEFINE
  p_TableName  STRING,
  p_json_obj   util.JSONObject
DEFINE
  ls_TableName  STRING,
  lo_DZEQ_T     DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName = p_TableName

  INITIALIZE lo_DZEQ_T TO NULL  #20180213
  INITIALIZE lo_RETURN TO NULL  #20180213
  
  LET ls_SQL = "select eq.dzeq001,eq.dzeq002,eq.dzeq003,eq.dzeq004,eq.dzeq005,",
               "       eq.dzeq006,eq.dzeq007,eq.dzeq008,eq.dzeq009,           ",
               "       eq.dzeqstus                                            ",  
               "  from dzeq_t eq                                              ",
               " where eq.dzeq001 = '",ls_TableName,"'                        "
               
  #取得Table export notice             
  LET ls_tsn_SQL = "select 'DZEQ_T' table_name,count(1) records ",
                   "  from dzeq_t eq                            ",
                   " where eq.dzeq001 = '",ls_TableName,"'      "
                 
  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)
  
  PREPARE lpre_GetDZEQData FROM ls_SQL
  DECLARE lcur_GetDZEQData CURSOR FOR lpre_GetDZEQData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZEQData
  FOREACH lcur_GetDZEQData INTO lo_DZEQ_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZEQData
  CALL lo_DZEQ_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZEQData
  FREE lcur_GetDZEQData  

  LET lo_RETURN = lo_DZEQ_T #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_import_DZEQ_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEQ_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZEQ_T_data(lo_DATAS) RETURNING lb_success
    END IF   
  END IF  

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZEQ_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,6] IS NOT NULL THEN
      LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "'",lo_DATAS[li_loop,3],"',",
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",lo_DATAS[li_loop,5],"',",
                           "",NVL(lo_DATAS[li_loop,6],"NULL"),",",
                           "'",lo_DATAS[li_loop,7],"',",
                           "'",lo_DATAS[li_loop,8],"',",
                           "'",lo_DATAS[li_loop,9],"',",
                           "'",lo_DATAS[li_loop,10],"'"
                                                    
      LET ls_InsertSQL = "insert into dzeq_t (                                         ",
                         "                    dzeq001,dzeq002,dzeq003,dzeq004,dzeq005, ",
                         "                    dzeq006,dzeq007,dzeq008,dzeq009,         ",
                         "                    dzeqstus                                 ",
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "
                         
      LET ls_UpdateValue = "dzeq002 = '",lo_DATAS[li_loop,2],"',",
                           "dzeq003 = '",lo_DATAS[li_loop,3],"',",
                           "dzeq004 = '",lo_DATAS[li_loop,4],"',",
                           "dzeq005 = '",lo_DATAS[li_loop,5],"',",
                           "dzeq007 = '",lo_DATAS[li_loop,7],"',",
                           "dzeq008 = '",lo_DATAS[li_loop,8],"',",
                           "dzeq009 = '",lo_DATAS[li_loop,9],"'"
                           --"dzeqstus = '",lo_DATAS[li_loop,10],"'"
                                                    
      LET ls_UpdateSQL = "update dzeq_t set ",
                         ls_UpdateValue,
                         " where dzeq001 = '",lo_DATAS[li_loop,1],"'",
                         "   and dzeq006 = ",NVL(lo_DATAS[li_loop,6],"NULL"),""

      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      LET lb_success = TRUE
      {
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
      }
    END IF                         
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZER_T #####################################################################

FUNCTION sadzi140_exim_export_DZER_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZER_T"
  
  CALL sadzi140_exim_get_DZER_data(ls_TableName,p_json_obj) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_DZER_data(p_TableName,p_json_obj)
DEFINE
  p_TableName  STRING,
  p_json_obj   util.JSONObject
DEFINE
  ls_TableName  STRING,
  lo_DZER_T     DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName = p_TableName

  INITIALIZE lo_DZER_T TO NULL  #20180213
  INITIALIZE lo_RETURN TO NULL  #20180213
  
  LET ls_SQL = "select er.dzer001,er.dzer002,er.dzer003,er.dzer004,er.dzer005,",
               "       er.dzer006,er.dzer007,er.dzer008,                      ",
               "       er.dzerstus                                            ",
               "  from dzer_t er                                              ",
               " where er.dzer001 = '",ls_TableName,"'                        "
               
  #取得Table export notice             
  LET ls_tsn_SQL = "select 'DZER_T' table_name,count(1) records ",
                   "  from dzer_t er                            ",
                   " where er.dzer001 = '",ls_TableName,"'      "
                 
  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)
  
  PREPARE lpre_GetDZERData FROM ls_SQL
  DECLARE lcur_GetDZERData CURSOR FOR lpre_GetDZERData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZERData
  FOREACH lcur_GetDZERData INTO lo_DZER_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZERData
  CALL lo_DZER_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZERData
  FREE lcur_GetDZERData  

  LET lo_RETURN = lo_DZER_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_import_DZER_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZER_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZER_T_data(lo_DATAS) RETURNING lb_success
    END IF   
  END IF   

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZER_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  lb_result       BOOLEAN,
  ls_IMPORT_TABLE STRING, #20180302
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  LET ls_IMPORT_TABLE = "DZER_T" #20180302
  
  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL 
      AND lo_DATAS[li_loop,3] IS NOT NULL THEN
      #20171120 begin  
      IF mb_pickup_exist THEN 
        CALL sadzi140_exim_check_if_exist_in_pick_list(mo_pick_col_list,lo_DATAS[li_loop,2]) RETURNING lb_result
        IF NOT lb_result THEN CONTINUE FOR END IF
      END IF
      #20171120 end

      #20180302 begin  
      IF mo_local_sys_info.psi_RT_PATCH_MODE <> "Y" THEN 
        CALL sadzi140_exim_check_if_exist_in_deleted_std_col_list(mo_deleted_std_col_list,lo_DATAS[li_loop,2]) RETURNING lb_result
        IF lb_result THEN 
          DISPLAY cs_warning_tag,"Column ",lo_DATAS[li_loop,2]," exists in deleted standard column list, skip import to ",ls_IMPORT_TABLE,".","\n",
                                 "You can using 'Tools'->'Field selection import' to import it !" 
          CONTINUE FOR 
        END IF
      END IF  
      #20180302 end
      
      LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "",NVL(lo_DATAS[li_loop,3],"NULL"),",",
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",lo_DATAS[li_loop,5],"',",
                           "'",lo_DATAS[li_loop,6],"',",
                           "'",lo_DATAS[li_loop,7],"',",
                           "'",lo_DATAS[li_loop,8],"',",
                           "'",lo_DATAS[li_loop,9],"'"
                                                    
      LET ls_InsertSQL = "insert into dzer_t (                                         ",
                         "                    dzer001,dzer002,dzer003,dzer004,dzer005, ",
                         "                    dzer006,dzer007,dzer008,                 ",
                         "                    dzerstus                                 ",
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "
                         
      LET ls_UpdateValue = "dzer004 = '",lo_DATAS[li_loop,4],"',",
                           "dzer005 = '",lo_DATAS[li_loop,5],"',",
                           "dzer006 = '",lo_DATAS[li_loop,6],"',",
                           "dzer007 = '",lo_DATAS[li_loop,7],"',",
                           "dzer008 = '",lo_DATAS[li_loop,8],"'"
                           --"dzerstus = '",lo_DATAS[li_loop,9],"'"
                                                    
      LET ls_UpdateSQL = "update dzer_t set ",
                         ls_UpdateValue,
                         " where dzer001 = '",lo_DATAS[li_loop,1],"'",
                         "   and dzer002 = '",lo_DATAS[li_loop,2],"'",
                         "   and dzer003 = ",NVL(lo_DATAS[li_loop,3],"NULL"),""
                           
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      LET lb_success = TRUE
      {
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
      }
    END IF                         
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZET_T #####################################################################

FUNCTION sadzi140_exim_export_DZET_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZET_T"
  
  CALL sadzi140_exim_get_DZET_data(ls_TableName,p_json_obj) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_DZET_data(p_TableName,p_json_obj)
DEFINE
  p_TableName  STRING,
  p_json_obj   util.JSONObject
DEFINE
  ls_TableName  STRING,
  lo_DZET_T     DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName = p_TableName

  INITIALIZE lo_DZET_T TO NULL  #20180213
  INITIALIZE lo_RETURN TO NULL  #20180213
  
  LET ls_SQL = "select et.dzet001,et.dzet002,et.dzet003,et.dzet004,et.dzet005,",
               "       et.dzet006,et.dzet007,                                 ",
               "       et.dzetstus                                            ", 
               "  from dzet_t et                                              ",
               " where et.dzet001 = '",ls_TableName,"'                        "
               
  #取得Table export notice             
  LET ls_tsn_SQL = "select 'DZET_T' table_name,count(1) records ",
                   "  from dzet_t et                            ",
                   " where et.dzet001 = '",ls_TableName,"'      "
                 
  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)
  
  PREPARE lpre_GetDZETData FROM ls_SQL
  DECLARE lcur_GetDZETData CURSOR FOR lpre_GetDZETData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZETData
  FOREACH lcur_GetDZETData INTO lo_DZET_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZETData
  CALL lo_DZET_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZETData
  FREE lcur_GetDZETData  

  LET lo_RETURN = lo_DZET_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_import_DZET_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZET_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZET_T_data(lo_DATAS) RETURNING lb_success
    END IF   
  END IF  

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZET_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_result       BOOLEAN, #20171120
  lb_success      BOOLEAN,
  ls_IMPORT_TABLE STRING, #20180302
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  LET ls_IMPORT_TABLE = "DZET_T" #20180302
  
  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN
      #20171120 begin  
      IF mb_pickup_exist THEN 
        CALL sadzi140_exim_check_if_exist_in_pick_list(mo_pick_col_list,lo_DATAS[li_loop,2]) RETURNING lb_result
        IF NOT lb_result THEN CONTINUE FOR END IF
      END IF
      #20171120 end
      
      #20180302 begin  
      IF mo_local_sys_info.psi_RT_PATCH_MODE <> "Y" THEN 
        CALL sadzi140_exim_check_if_exist_in_deleted_std_col_list(mo_deleted_std_col_list,lo_DATAS[li_loop,2]) RETURNING lb_result
        IF lb_result THEN 
          DISPLAY cs_warning_tag,"Column ",lo_DATAS[li_loop,2]," exists in deleted standard column list, skip import to ",ls_IMPORT_TABLE,".","\n",
                                 "You can using 'Tools'->'Field selection import' to import it !" 
          CONTINUE FOR 
        END IF
      END IF  
      #20180302 end
      
      LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "'",lo_DATAS[li_loop,3],"',",
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",lo_DATAS[li_loop,5],"',",
                           "'",lo_DATAS[li_loop,6],"',",
                           "'",lo_DATAS[li_loop,7],"',",
                           "'",lo_DATAS[li_loop,8],"'"
                                                    
      LET ls_InsertSQL = "insert into dzet_t (                                         ",
                         "                    dzet001,dzet002,dzet003,dzet004,dzet005, ",
                         "                    dzet006,dzet007,                         ",
                         "                    dzetstus                                 ",
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "
                         
      LET ls_UpdateValue = "dzet003 = '",lo_DATAS[li_loop,3],"',",
                           "dzet004 = '",lo_DATAS[li_loop,4],"',",
                           "dzet005 = '",lo_DATAS[li_loop,5],"',",
                           "dzet006 = '",lo_DATAS[li_loop,6],"',",
                           "dzet007 = '",lo_DATAS[li_loop,7],"'"
                           --"dzetstus = '",lo_DATAS[li_loop,8],"'"
                                                    
      LET ls_UpdateSQL = "update dzet_t set ",
                         ls_UpdateValue,
                         " where dzet001 = '",lo_DATAS[li_loop,1],"'",
                         "   and dzet002 = '",lo_DATAS[li_loop,2],"'"
                           
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      LET lb_success = TRUE
      {
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
      }
    END IF    
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZEAL_T #####################################################################

FUNCTION sadzi140_exim_export_DZEAL_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEAL_T"
  
  CALL sadzi140_exim_get_DZEAL_data(ls_TableName,p_json_obj) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_DZEAL_data(p_TableName,p_json_obj)
DEFINE
  p_TableName  STRING,
  p_json_obj   util.JSONObject
DEFINE
  ls_TableName  STRING,
  lo_DZEAL_T     DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName = p_TableName

  INITIALIZE lo_DZEAL_T TO NULL  #20180213
  INITIALIZE lo_RETURN TO NULL  #20180213
  
  LET ls_SQL = "select eal.dzeal001,eal.dzeal002,eal.dzeal003,eal.dzeal004,eal.dzeal005 ",
               "  from dzeal_t eal                                                      ",
               " where eal.dzeal001 = '",ls_TableName,"'                                "
               
  #取得Table export notice             
  LET ls_tsn_SQL = "select 'DZEAL_T' table_name,count(1) records ",
                   "  from dzeal_t eal                           ",
                   " where eal.dzeal001 = '",ls_TableName,"'     "
                 
  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)
  
  PREPARE lpre_GetDZEALData FROM ls_SQL
  DECLARE lcur_GetDZEALData CURSOR FOR lpre_GetDZEALData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZEALData
  FOREACH lcur_GetDZEALData INTO lo_DZEAL_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZEALData
  CALL lo_DZEAL_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZEALData
  FREE lcur_GetDZEALData  

  LET lo_RETURN = lo_DZEAL_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_import_DZEAL_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEAL_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZEAL_T_data(lo_DATAS) RETURNING lb_success
    END IF   
  END IF   

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZEAL_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  lb_is_lang      BOOLEAN,
  li_lng          INTEGER,   
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN

      #從清單中取得可匯入的語言別
      LET lb_is_lang = FALSE
      FOR li_lng = 1 TO mo_local_lang.getLength()
        IF lo_DATAS[li_loop,2] = mo_local_lang[li_lng] THEN
          LET lb_is_lang = TRUE
          EXIT FOR
        END IF 
      END FOR 

      #符合當地多語言清單的資料才匯入
      IF lb_is_lang THEN
      
        LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                             "'",NVL(lo_DATAS[li_loop,2],"NULL"),"',",
                             "'",lo_DATAS[li_loop,3],"',",
                             "'",lo_DATAS[li_loop,4],"',",
                             "'",lo_DATAS[li_loop,5],"'"
                                                      
        LET ls_InsertSQL = "insert into dzeal_t (                                           ",
                           "                    dzeal001,dzeal002,dzeal003,dzeal004,dzeal005",
                           "                   )                                            ",  
                           " values (                                                       ",
                           ls_InsertValue,
                           "        )                                                       "  
                           
        LET ls_UpdateValue = "dzeal003 = '",lo_DATAS[li_loop,3],"',",
                             "dzeal004 = '",lo_DATAS[li_loop,4],"',",
                             "dzeal005 = '",lo_DATAS[li_loop,5],"'"
                                                      
        LET ls_UpdateSQL = "update dzeal_t set ",
                           ls_UpdateValue,
                           " where dzeal001 = '",lo_DATAS[li_loop,1],"'",
                           "   and dzeal002 = '",NVL(lo_DATAS[li_loop,2],"NULL"),"'"

        CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
        IF NOT lb_success THEN
          CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
          IF NOT lb_success THEN
            DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
            DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
            EXIT FOR 
          END IF
        END IF
        
      END IF      
    END IF                       
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZEO_T #####################################################################

FUNCTION sadzi140_exim_export_DZEO_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEO_T"
  
  CALL sadzi140_exim_get_DZEO_data(ls_TableName,p_json_obj) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_DZEO_data(p_TableName,p_json_obj)
DEFINE
  p_TableName  STRING,
  p_json_obj   util.JSONObject
DEFINE
  ls_TableName  STRING,
  lo_DZEO_T     DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName = p_TableName

  INITIALIZE lo_DZEO_T TO NULL  #20180213
  INITIALIZE lo_RETURN TO NULL  #20180213
  
  LET ls_SQL = "select eo.dzeo001,eo.dzeo002,eo.dzeo003,eo.dzeo004,eo.dzeo005,",
               "       eo.dzeo006,eo.dzeo007,eo.dzeo008,eo.dzeo009            ",
               "  from dzeo_t eo                                              ",
               " where eo.dzeo001 = '",ls_TableName.toUpperCase(),"'          "
               
  #取得Table export notice             
  LET ls_tsn_SQL = "select 'DZEO_T' table_name,count(1) records          ",
                   "  from dzeo_t eo                                     ",
                   " where eo.dzeo001 = '",ls_TableName.toUpperCase(),"' "
                 
  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)
  
  PREPARE lpre_GetDZEOData FROM ls_SQL
  DECLARE lcur_GetDZEOData CURSOR FOR lpre_GetDZEOData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZEOData
  FOREACH lcur_GetDZEOData INTO lo_DZEO_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZEOData
  CALL lo_DZEO_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZEOData
  FREE lcur_GetDZEOData  

  LET lo_RETURN = lo_DZEO_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_import_DZEO_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEO_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZEO_T_data(lo_DATAS) RETURNING lb_success
    END IF   
  END IF   

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZEO_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL THEN
      LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                           "",NVL(lo_DATAS[li_loop,2],"NULL"),",",
                           "",NVL(lo_DATAS[li_loop,3],"NULL"),",",
                           "",NVL(lo_DATAS[li_loop,4],"NULL"),",",
                           "",NVL(lo_DATAS[li_loop,5],"NULL"),",",
                           "'",lo_DATAS[li_loop,6],"',",
                           "'",lo_DATAS[li_loop,7],"',",
                           "'",lo_DATAS[li_loop,8],"',",
                           "'",lo_DATAS[li_loop,9],"'"
                                                    
      LET ls_InsertSQL = "insert into dzeo_t (                                         ",
                         "                    dzeo001,dzeo002,dzeo003,dzeo004,dzeo005, ",
                         "                    dzeo006,dzeo007,dzeo008,dzeo009          ",
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "
                         
      LET ls_UpdateValue = "dzeo002 = ",NVL(lo_DATAS[li_loop,2],"NULL"),",",
                           "dzeo003 = ",NVL(lo_DATAS[li_loop,3],"NULL"),",",
                           "dzeo004 = ",NVL(lo_DATAS[li_loop,4],"NULL"),",",
                           "dzeo005 = ",NVL(lo_DATAS[li_loop,5],"NULL"),",",
                           "dzeo006 = '",lo_DATAS[li_loop,6],"',",
                           "dzeo007 = '",lo_DATAS[li_loop,7],"',",
                           "dzeo008 = '",lo_DATAS[li_loop,8],"',",
                           "dzeo009 = '",lo_DATAS[li_loop,9],"'"
                                                    
      LET ls_UpdateSQL = "update dzeo_t set ",
                         ls_UpdateValue,
                         " where dzeo001 = '",lo_DATAS[li_loop,1],"'"
                         
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      LET lb_success = TRUE
      {
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
      }
    END IF
  END FOR 

  
  RETURN lb_success
  
END FUNCTION

## DZEU_T #####################################################################

FUNCTION sadzi140_exim_export_DZEU_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEU_T"
  
  CALL sadzi140_exim_get_DZEU_data(ls_TableName,p_json_obj) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_DZEU_data(p_TableName,p_json_obj)
DEFINE
  p_TableName  STRING,
  p_json_obj   util.JSONObject
DEFINE
  ls_TableName  STRING,
  lo_DZEU_T     DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName = p_TableName

  INITIALIZE lo_DZEU_T TO NULL  #20180213
  INITIALIZE lo_RETURN TO NULL  #20180213
  
  LET ls_SQL = "select eu.dzeu001,eu.dzeu002,eu.dzeu003,eu.dzeu004,eu.dzeu005 ",
               "  from dzeu_t eu                                              ",
               " where eu.dzeu001 = '",ls_TableName,"'                        "
               
  #取得Table export notice             
  LET ls_tsn_SQL = "select 'DZEU_T' table_name,count(1) records  ",
                   "  from dzeu_t eu                             ",
                   " where eu.dzeu001 = '",ls_TableName,"'       "
                 
  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)
  
  PREPARE lpre_GetDZEUData FROM ls_SQL
  DECLARE lcur_GetDZEUData CURSOR FOR lpre_GetDZEUData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZEUData
  FOREACH lcur_GetDZEUData INTO lo_DZEU_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZEUData
  CALL lo_DZEU_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZEUData
  FREE lcur_GetDZEUData  

  LET lo_RETURN = lo_DZEU_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_import_DZEU_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEU_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZEU_T_data(lo_DATAS) RETURNING lb_success
    END IF   
  END IF   

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZEU_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN
      LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "'",lo_DATAS[li_loop,3],"',",
                           "'",lo_DATAS[li_loop,4],"',",
                           "",NVL(lo_DATAS[li_loop,5],"NULL"),""
                                                    
      LET ls_InsertSQL = "insert into dzeu_t (                                         ",
                         "                    dzeu001,dzeu002,dzeu003,dzeu004,dzeu005  ",
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "
                         
      LET ls_UpdateValue = "dzeu003 = '",lo_DATAS[li_loop,3],"',",
                           "dzeu004 = '",lo_DATAS[li_loop,4],"',",
                           "dzeu005 = ",NVL(lo_DATAS[li_loop,5],"NULL"),""
                                                    
      LET ls_UpdateSQL = "update dzeu_t set ",
                         ls_UpdateValue,
                         " where dzeu001 = '",lo_DATAS[li_loop,1],"'",
                         "   and dzeu002 = '",lo_DATAS[li_loop,2],"'"
                           
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
    END IF    
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZEV_T #####################################################################

FUNCTION sadzi140_exim_export_DZEV_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEV_T"
  
  CALL sadzi140_exim_get_DZEV_data(ls_TableName,p_json_obj,lo_PARAMETERS.*) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_DZEV_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName   STRING,
  lo_REVISION    T_REVISION,
  lo_PARAMETERS  T_DZLM_T_SCR,  
  ls_MaxVersion  STRING,
  lo_DZEV_T      DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  ls_SQL_ALM STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*
  
  INITIALIZE lo_DZEV_T TO NULL  #20180213
  INITIALIZE lo_RETURN TO NULL  #20180213
  
  LET ls_SQL_ALM = ""

  LET ls_SQL = "select ev.dzev001,ev.dzev002,ev.dzev003,ev.dzev004,ev.dzev005,",
               "       ev.dzev006,ev.dzev007,ev.dzev008,ev.dzev009,ev.dzev010,",
               "       ev.dzev011,ev.dzev012,ev.dzev013,ev.dzev014,ev.dzev015,",
               "       ev.dzev016,ev.dzev017,ev.dzev018,ev.dzev019,ev.dzev020,",
               "       ev.dzev021,ev.dzev022,ev.dzev023                       ",  
               "  from dzev_t ev                                              ",
               " where ev.dzev002 = '",ls_TableName.toUpperCase(),"'          ",
               " order by ev.dzev011                                          " 

  #取得Table export notice             
  LET ls_tsn_SQL = "select 'DZEV_T' table_name,count(1) records          ",
                   "  from dzev_t ev                                     ",
                   " where ev.dzev002 = '",ls_TableName.toUpperCase(),"' "
                 
  IF lo_parameters.DZLM005 = cs_NoVersionCharacter THEN
    #取得該表格最大的異動版次
    CALL sadzi140_rev_get_revision_data(ls_TableName) RETURNING lo_REVISION.*
    LET ls_SQL_ALM = "   and ev.dzev003 = trim('",lo_REVISION.rv_SEQUENCE,"')", 
                     "   and ev.dzev018 = '",lo_REVISION.rv_ALM_VERSION,"'   ",
                     "   and ev.dzev019 = '",lo_REVISION.rv_REQUEST_NO,"'    "

  ELSE                    
    LET ls_SQL_ALM = "   and ev.dzev003 = trim('",lo_PARAMETERS.DZLM015,"') ", 
                     "   and ev.dzev018 = '",lo_PARAMETERS.DZLM005,"' ",
                     "   and ev.dzev019 = '",lo_PARAMETERS.DZLM012,"' "
  END IF
  
  
  LET ls_SQL = ls_SQL,ls_SQL_ALM   
  LET ls_tsn_SQL = ls_tsn_SQL,ls_SQL_ALM

  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)
  
  PREPARE lpre_GetDZEVData FROM ls_SQL
  DECLARE lcur_GetDZEVData CURSOR FOR lpre_GetDZEVData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZEVData
  FOREACH lcur_GetDZEVData INTO lo_DZEV_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZEVData
  CALL lo_DZEV_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZEVData
  FREE lcur_GetDZEVData  

  LET lo_RETURN = lo_DZEV_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_import_DZEV_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEV_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZEV_T_data(lo_DATAS) RETURNING lb_success
    END IF   
  END IF   

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZEV_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL AND
       lo_DATAS[li_loop,3] IS NOT NULL AND lo_DATAS[li_loop,4] IS NOT NULL AND
       lo_DATAS[li_loop,5] IS NOT NULL THEN
      LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           #"'",IIF(mb_same_dgenv,ms_max_sequence,lo_DATAS[li_loop,3]),"',", #異動序號
                           "'",ms_max_sequence,"',", #異動序號
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",lo_DATAS[li_loop,5],"',",
                           "'",lo_DATAS[li_loop,6],"',",
                           "'",lo_DATAS[li_loop,7],"',",
                           "'",lo_DATAS[li_loop,8],"',",
                           "'",lo_DATAS[li_loop,9],"',",
                           "'",lo_DATAS[li_loop,10],"',",
                           "",NVL(lo_DATAS[li_loop,11],"NULL"),",",
                           "'",lo_DATAS[li_loop,12],"',",
                           "'",lo_DATAS[li_loop,13],"',",
                           "",NVL(lo_DATAS[li_loop,14],"NULL"),",",
                           "'",lo_DATAS[li_loop,15],"',",
                           "",NVL(lo_DATAS[li_loop,16],"NULL"),",",
                           "'",lo_DATAS[li_loop,17],"',",
                           #"'",lo_DATAS[li_loop,18],"',", #建構版次
                           "'",ms_construct_version,"',",
                           "'",lo_DATAS[li_loop,19],"',",
                           "'",lo_DATAS[li_loop,20],"',",
                           "'",lo_DATAS[li_loop,21],"',",
                           "'",lo_DATAS[li_loop,22],"',",
                           "'",lo_DATAS[li_loop,23],"'"
                                                    
      LET ls_InsertSQL = "insert into dzev_t (                                         ",
                         "                    dzev001,dzev002,dzev003,dzev004,dzev005, ",
                         "                    dzev006,dzev007,dzev008,dzev009,dzev010, ",
                         "                    dzev011,dzev012,dzev013,dzev014,dzev015, ",
                         "                    dzev016,dzev017,dzev018,dzev019,dzev020, ",
                         "                    dzev021,dzev022,dzev023                  ",
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "
                         
      LET ls_UpdateValue = "dzev006 = '",lo_DATAS[li_loop,6],"',",
                           "dzev007 = '",lo_DATAS[li_loop,7],"',",
                           "dzev008 = '",lo_DATAS[li_loop,8],"',",
                           "dzev009 = '",lo_DATAS[li_loop,9],"',",
                           "dzev010 = '",lo_DATAS[li_loop,10],"',",
                           "dzev011 = ",NVL(lo_DATAS[li_loop,11],"NULL"),",",
                           "dzev012 = '",lo_DATAS[li_loop,12],"',",
                           "dzev013 = '",lo_DATAS[li_loop,13],"',",
                           "dzev014 = ",NVL(lo_DATAS[li_loop,14],"NULL"),",",
                           "dzev015 = '",lo_DATAS[li_loop,15],"',",
                           "dzev016 = ",NVL(lo_DATAS[li_loop,16],"NULL"),",",
                           "dzev017 = '",lo_DATAS[li_loop,17],"',",
                           #"dzev018 = '",lo_DATAS[li_loop,18],"',", #建構版次
                           "dzev018 = '",ms_construct_version,"',",
                           "dzev019 = '",lo_DATAS[li_loop,19],"',",
                           "dzev020 = '",lo_DATAS[li_loop,20],"',",
                           "dzev021 = '",lo_DATAS[li_loop,21],"',",
                           "dzev022 = '",lo_DATAS[li_loop,22],"',",
                           "dzev023 = '",lo_DATAS[li_loop,23],"'"
                                                    
      LET ls_UpdateSQL = "update dzev_t set ",
                         ls_UpdateValue,
                         " where dzev001 = '",lo_DATAS[li_loop,1],"'",
                         "   and dzev002 = '",lo_DATAS[li_loop,2],"'",
                         #"   and dzev003 = '",IIF(mb_same_dgenv,ms_max_sequence,lo_DATAS[li_loop,3]),"'", #異動序號
                         "   and dzev003 = '",ms_max_sequence,"'", #異動序號
                         "   and dzev004 = '",lo_DATAS[li_loop,4],"'",
                         "   and dzev005 = '",lo_DATAS[li_loop,5],"'"
                           
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
    END IF    
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZEW_T #####################################################################

FUNCTION sadzi140_exim_export_DZEW_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEW_T"
  
  CALL sadzi140_exim_get_DZEW_data(ls_TableName,p_json_obj,lo_PARAMETERS.*) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_DZEW_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName   STRING,
  p_json_obj    util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName   STRING,
  lo_PARAMETERS  T_DZLM_T_SCR,
  lo_REVISION    T_REVISION,
  ls_MaxVersion  STRING,
  lo_DZEW_T      DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  ls_SQL_ALM STRING,
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  INITIALIZE lo_DZEW_T TO NULL  #20180213
  INITIALIZE lo_RETURN TO NULL  #20180213
  LET ls_SQL_ALM = ""
  
  LET ls_SQL = "select ew.dzew001,ew.dzew002,ew.dzew003,ew.dzew004,ew.dzew005,",
               "       ew.dzew006,ew.dzew007,ew.dzew008,ew.dzew009,ew.dzew010,",
               "       ew.dzew011,ew.dzew012,ew.dzew013,ew.dzew014            ",
               "  from dzew_t ew                                              ",
               " where ew.dzew001 = '",ls_TableName,"'                        "

  #取得Table export notice             
  LET ls_tsn_SQL = "select 'DZEW_T' table_name,count(1) records  ",
                   "  from dzew_t ew                             ",
                   " where ew.dzew001 = '",ls_TableName,"'       "

                   
  IF lo_parameters.DZLM005 = cs_NoVersionCharacter THEN
    #取得該表格最大的異動版次
    CALL sadzi140_rev_get_revision_data(ls_TableName) RETURNING lo_REVISION.*
    LET ls_SQL_ALM = " and ew.dzew002 = trim('",lo_REVISION.rv_SEQUENCE,"')",
                     " and ew.dzew010 = '",lo_REVISION.rv_ALM_VERSION,"'   ",
                     " and ew.dzew011 = '",lo_REVISION.rv_REQUEST_NO,"'    "  

  ELSE                    
    LET ls_SQL_ALM = " and ew.dzew002 = trim('",lo_PARAMETERS.DZLM015,"')",
                     " and ew.dzew010 = '",lo_PARAMETERS.DZLM005,"'      ",
                     " and ew.dzew011 = '",lo_PARAMETERS.DZLM012,"'      "
  END IF 
  
  LET ls_SQL = ls_SQL,ls_SQL_ALM   
  LET ls_tsn_SQL = ls_tsn_SQL,ls_SQL_ALM   

  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)
               
  PREPARE lpre_GetDZEWData FROM ls_SQL
  DECLARE lcur_GetDZEWData CURSOR FOR lpre_GetDZEWData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZEWData
  FOREACH lcur_GetDZEWData INTO lo_DZEW_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZEWData
  CALL lo_DZEW_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZEWData
  FREE lcur_GetDZEWData  

  LET lo_RETURN = lo_DZEW_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_import_DZEW_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEW_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZEW_T_data(lo_DATAS) RETURNING lb_success
    END IF   
  END IF   

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZEW_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL AND
       lo_DATAS[li_loop,3] IS NOT NULL THEN
      LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                           #"'",IIF(mb_same_dgenv,ms_max_sequence,lo_DATAS[li_loop,2]),"',", #異動序號
                           "'",ms_max_sequence,"',", #異動序號
                           "'",lo_DATAS[li_loop,3],"',",
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",lo_DATAS[li_loop,5],"',",
                           "'",lo_DATAS[li_loop,6],"',",
                           "trim('",lo_DATAS[li_loop,7],"'),",
                           "trim('",lo_DATAS[li_loop,8],"'),",
                           "'",lo_DATAS[li_loop,9],"',",
                           #"'",lo_DATAS[li_loop,10],"',", #建構版次
                           "'",ms_construct_version,"',",
                           "'",lo_DATAS[li_loop,11],"',",
                           "'",lo_DATAS[li_loop,12],"',",
                           "'",lo_DATAS[li_loop,13],"',",
                           "'",lo_DATAS[li_loop,14],"'"
                                                    
      LET ls_InsertSQL = "insert into dzew_t (                                         ",
                         "                    dzew001,dzew002,dzew003,dzew004,dzew005, ",
                         "                    dzew006,dzew007,dzew008,dzew009,dzew010, ",
                         "                    dzew011,dzew012,dzew013,dzew014          ",
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "
                         
      LET ls_UpdateValue = "dzew005 = '",lo_DATAS[li_loop,5],"',",
                           "dzew006 = '",lo_DATAS[li_loop,6],"',",
                           "dzew007 = '",lo_DATAS[li_loop,7],"',",
                           "dzew008 = trim('",lo_DATAS[li_loop,8],"'),",
                           "dzew009 = trim('",lo_DATAS[li_loop,9],"'),",
                           #"dzew010 = '",lo_DATAS[li_loop,10],"',", #建構版次
                           "dzew010 = '",ms_construct_version,"',",
                           "dzew011 = '",lo_DATAS[li_loop,11],"',",
                           "dzew012 = '",lo_DATAS[li_loop,12],"',",
                           "dzew013 = '",lo_DATAS[li_loop,13],"',",
                           "dzew014 = '",lo_DATAS[li_loop,14],"'"
                                                    
      LET ls_UpdateSQL = "update dzew_t set ",
                         ls_UpdateValue,
                         " where dzew001 = '",lo_DATAS[li_loop,1],"'",
                         #"   and dzew002 = '",IIF(mb_same_dgenv,ms_max_sequence,lo_DATAS[li_loop,2]),"'", #異動序號
                         "   and dzew002 = '",ms_max_sequence,"'", #異動序號
                         "   and dzew003 = '",lo_DATAS[li_loop,3],"'",
                         "   and dzew004 = '",lo_DATAS[li_loop,4],"'"
                           
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
    END IF    
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZIG_T #####################################################################

FUNCTION sadzi140_exim_export_DZIG_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZIG_T"
  
  CALL sadzi140_exim_get_DZIG_data(ls_TableName,p_json_obj) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_DZIG_data(p_TableName,p_json_obj)
DEFINE
  p_TableName  STRING,
  p_json_obj   util.JSONObject
DEFINE
  ls_TableName  STRING,
  lo_DZIG_T     DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName = p_TableName

  INITIALIZE lo_DZIG_T TO NULL  #20180213
  INITIALIZE lo_RETURN TO NULL  #20180213
  
  LET ls_SQL = "select ig.dzig001,ig.dzig002,ig.dzig003 ",
               "  from dzig_t ig                        ",
               " where ig.dzig001 = '",ls_TableName,"'  "
               
  #取得Table export notice             
  LET ls_tsn_SQL = "select 'DZIG_T' table_name,count(1) records ",
                   "  from dzig_t ig                            ",
                   " where ig.dzig001 = '",ls_TableName,"'      "
                 
  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)
  
  PREPARE lpre_GetDZIGData FROM ls_SQL
  DECLARE lcur_GetDZIGData CURSOR FOR lpre_GetDZIGData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZIGData
  FOREACH lcur_GetDZIGData INTO lo_DZIG_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZIGData
  CALL lo_DZIG_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZIGData
  FREE lcur_GetDZIGData  

  LET lo_RETURN = lo_DZIG_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_import_DZIG_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZIG_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_warning_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZIG_T_data(lo_DATAS) RETURNING lb_success
    END IF  
  ELSE  
    #為了相容舊版匯出檔, 開檔不成功一律設為 TRUE
    LET lb_success = TRUE
  END IF  
  
  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZIG_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_result       BOOLEAN, #20171120
  lb_success      BOOLEAN,
  ls_IMPORT_TABLE STRING, #20180302
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  LET ls_IMPORT_TABLE = "DZIG_T" #20180302

  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN
      #20171120 begin  
      IF mb_pickup_exist THEN 
        CALL sadzi140_exim_check_if_exist_in_pick_list(mo_pick_col_list,lo_DATAS[li_loop,2]) RETURNING lb_result
        IF NOT lb_result THEN CONTINUE FOR END IF
      END IF
      #20171120 end
      
      #20180302 begin  
      IF mo_local_sys_info.psi_RT_PATCH_MODE <> "Y" THEN 
        CALL sadzi140_exim_check_if_exist_in_deleted_std_col_list(mo_deleted_std_col_list,lo_DATAS[li_loop,2]) RETURNING lb_result
        IF lb_result THEN 
          DISPLAY cs_warning_tag,"Column ",lo_DATAS[li_loop,2]," exists in deleted standard column list, skip import to ",ls_IMPORT_TABLE,".","\n",
                                 "You can using 'Tools'->'Field selection import' to import it !" 
          CONTINUE FOR 
        END IF
      END IF  
      #20180302 end
      
      LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "'",lo_DATAS[li_loop,3],"' "
                                                    
      LET ls_InsertSQL = "insert into dzig_t (                         ",
                         "                    dzig001,dzig002,dzig003  ",
                         "                   )                         ",  
                         " values (                                    ",
                         ls_InsertValue,                              
                         "        )                                    "
                         
      LET ls_UpdateValue = "dzig003 = '",lo_DATAS[li_loop,3],"'"
                                                    
      LET ls_UpdateSQL = "update dzig_t set ",
                         ls_UpdateValue,
                         " where dzig001 = '",lo_DATAS[li_loop,1],"'",
                         "   and dzig002 = '",lo_DATAS[li_loop,2],"'"
                           
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
    END IF    
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZEN_T #####################################################################

FUNCTION sadzi140_exim_export_DZEN_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEN_T"
  
  CALL sadzi140_exim_get_DZEN_data(ls_TableName,p_json_obj) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_DZEN_data(p_TableName,p_json_obj)
DEFINE
  p_TableName  STRING,
  p_json_obj   util.JSONObject
DEFINE
  ls_TableName  STRING,
  lo_DZEN_T     DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName = p_TableName

  INITIALIZE lo_DZEN_T TO NULL  #20180213
  INITIALIZE lo_RETURN TO NULL  #20180213
  
  LET ls_SQL = "select en.dzen001,en.dzen002,en.dzen003,en.dzen004,en.dzen005, ",
               "       en.dzen006,en.dzen007,en.dzen008,en.dzen009,en.dzen010, ",
               "       en.dzen011,en.dzen012,en.dzen013                        ",
               "  from dzen_t en                                               ",
               " where en.dzen001 = '",ls_TableName,"'                         "
               
  #取得Table export notice             
  LET ls_tsn_SQL = "select 'DZEN_T' table_name,count(1) records ",
                   "  from dzen_t en                            ",
                   " where en.dzen001 = '",ls_TableName,"'      "
                 
  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)
  
  PREPARE lpre_GetDZENData FROM ls_SQL
  DECLARE lcur_GetDZENData CURSOR FOR lpre_GetDZENData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetDZENData
  FOREACH lcur_GetDZENData INTO lo_DZEN_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetDZENData
  CALL lo_DZEN_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetDZENData
  FREE lcur_GetDZENData  

  LET lo_RETURN = lo_DZEN_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_delete_DZEN_data(p_TableName,p_same_dgenv,p_same_topind) #20190304
DEFINE
  p_TableName    STRING,
  p_same_dgenv   BOOLEAN,
  p_same_topind  BOOLEAN  #20190304
DEFINE
  ls_TableName   STRING,
  lb_same_dgenv  BOOLEAN, 
  lb_same_topind BOOLEAN, #20190304
  ls_DeleteSQL   STRING, 
  ls_sql_cond    STRING,
  lb_success     BOOLEAN 

  LET ls_TableName   = p_TableName.toLowerCase()
  LET lb_same_dgenv  = p_same_dgenv
  LET lb_same_topind = p_same_topind #20190304

  IF lb_same_dgenv THEN --AND NOT mb_dgenv_is_cust THEN #20170524 #20170523
    LET ls_sql_cond = ""
  ELSE
    LET ls_sql_cond = " and coalesce(dzen011,'",cs_dgenv_standard,"') = '",cs_dgenv_standard,"'"
  END IF
  
  LET ls_DeleteSQL = "delete from dzen_t                               ",
                     " where dzen001 = '",ls_TableName.toLowerCase(),"'",
                     ls_sql_cond
                     
  #DISPLAY ls_DeleteSQL
  
  CALL sadzi140_exim_exec_SQL(ls_DeleteSQL) RETURNING lb_success

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_import_DZEN_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEN_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_warning_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZEN_T_data(lo_DATAS) RETURNING lb_success
    END IF  
  ELSE  
    #為了相容舊版匯出檔, 開檔不成功一律設為 TRUE
    LET lb_success = TRUE
  END IF  
  
  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZEN_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  ls_user         VARCHAR(100),
  ldt_datetime    DATETIME YEAR TO SECOND,
  lb_success      BOOLEAN,
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  &ifndef DEBUG
  LET ls_user = g_user
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  &endif
  LET ldt_datetime = CURRENT YEAR TO SECOND
  
  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN
    
      LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "'",lo_DATAS[li_loop,3],"',",
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",lo_DATAS[li_loop,5],"',",
                           "'",lo_DATAS[li_loop,6],"',",
                           "'",lo_DATAS[li_loop,7],"',",
                           "'",lo_DATAS[li_loop,8],"',",
                           "'",lo_DATAS[li_loop,9],"',",
                           "'",lo_DATAS[li_loop,10],"',",
                           "'",lo_DATAS[li_loop,11],"',",
                           "'",lo_DATAS[li_loop,12],"',",
                           "'",lo_DATAS[li_loop,13],"',",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,",",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,""
                                                    
      LET ls_InsertSQL = "insert into dzen_t (                                         ",
                         "                    dzen001,dzen002,dzen003,dzen004,dzen005, ",
                         "                    dzen006,dzen007,dzen008,dzen009,dzen010, ",
                         "                    dzen011,dzen012,dzen013,                 ",
                         "                    dzencrtid,dzencrtdt,dzenmodid,dzenmoddt  ",
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,                              
                         "        )                                                    "
                         
      LET ls_UpdateValue = "dzen004 = case dzen004 when 'Y' then dzen004 else '",lo_DATAS[li_loop,4],"' end, ",
                           "dzen005 = case dzen005 when 'Y' then dzen005 else '",lo_DATAS[li_loop,5],"' end, ",
                           "dzen006 = case dzen006 when 'Y' then dzen006 else '",lo_DATAS[li_loop,6],"' end, ",
                           "dzen007 = case dzen007 when 'Y' then dzen007 else '",lo_DATAS[li_loop,7],"' end, ",
                           "dzen008 = case dzen008 when 'Y' then dzen008 else '",lo_DATAS[li_loop,8],"' end, ",
                           "dzen009 = case dzen009 when 'Y' then dzen009 else '",lo_DATAS[li_loop,9],"' end, ",
                           "dzen010 = case dzen010 when 'Y' then dzen010 else '",lo_DATAS[li_loop,10],"' end,",
                           "dzen011 = '",lo_DATAS[li_loop,11],"',                            ",
                           "dzen012 = '",lo_DATAS[li_loop,12],"',                            ",
                           "dzen013 = '",lo_DATAS[li_loop,13],"',                            ",
                           "dzenmodid = trim('",ls_user,"'),                                 ",
                           "dzenmoddt = ",cs_get_time_stamp,"                                "
                           
      LET ls_UpdateSQL = "update dzen_t set ",
                         ls_UpdateValue,
                         " where dzen001 = '",lo_DATAS[li_loop,1],"'",
                         "   and dzen002 = '",lo_DATAS[li_loop,2],"'",
                         "   and dzen003 = '",lo_DATAS[li_loop,3],"'"
                           
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
    END IF    
  END FOR 

  RETURN lb_success
  
END FUNCTION

################################################################################

FUNCTION sadzi140_exim_get_data_from_file(p_FileName,p_fault_type)
DEFINE
  p_FileName   STRING,
  p_fault_type STRING
DEFINE
  ls_FileName   STRING,
  ls_fault_type STRING,
  lo_DATAS      DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE  
  lo_channel_read  base.Channel,
  ls_Text          STRING,
  ls_TextLine      VARCHAR(1024),
  li_RecCnt        INTEGER,
  li_TokenCnt      INTEGER,
  ls_TokenStr      STRING,
  lb_success       BOOLEAN,
  lo_RETURN        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  lo_token         base.StringTokenizer
  
  LET ls_FileName   = p_FileName
  LET ls_fault_type = p_fault_type

  LET lo_RETURN = NULL
  LET lb_success = TRUE

  LET lo_channel_read = base.Channel.create()
  TRY 
    CALL lo_channel_read.openFile(ls_FileName,"r")
  CATCH
    LET lb_success = FALSE
    DISPLAY ls_fault_type,"Open file ",ls_FileName," fail !!"
  END TRY  

  IF lb_success THEN
    LET li_RecCnt = 1 
    WHILE TRUE
      IF lo_channel_read.isEof() THEN 
        EXIT WHILE
      END IF

      LET ls_Text = lo_channel_read.readLine()
      LET ls_TextLine = ls_Text.Trim()
      IF NVL(ls_TextLine,cs_null_value) <> cs_null_value THEN
        LET lo_token = base.StringTokenizer.createExt(ls_TextLine,"|","",TRUE)

        LET li_TokenCnt = 0
        WHILE lo_token.hasMoreTokens()

          LET li_TokenCnt = li_TokenCnt + 1
          LET ls_TokenStr = lo_token.nextToken()
          LET lo_DATAS[li_RecCnt,li_TokenCnt] = ls_TokenStr CLIPPED

        END WHILE
        
        LET li_RecCnt = li_RecCnt + 1
      END IF   
        
    END WHILE
  END IF
  
  CALL lo_channel_read.close()

  LET lo_RETURN.* = lo_DATAS.* 
  
  RETURN lo_RETURN,lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_export_data(p_DATA_LINE)
DEFINE
  p_DATA_LINE DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  lo_line_string_buf  base.StringBuffer
DEFINE 
  li_loop   INTEGER,
  ls_Text   STRING,
  ls_Return STRING  

  LET lo_DATA_LINE.* = p_DATA_LINE.*
  LET ls_Text = ""

  LET lo_line_string_buf = base.StringBuffer.create()
  CALL lo_line_string_buf.clear()

  FOR li_loop = 1 TO lo_DATA_LINE.getLength()
    LET ls_Text = ""

    LET ls_Text = sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA001) CLIPPED,"|",
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA002) CLIPPED,"|",    
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA003) CLIPPED,"|",   
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA004) CLIPPED,"|",    
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA005) CLIPPED,"|",    
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA006) CLIPPED,"|",    
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA007) CLIPPED,"|",    
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA008) CLIPPED,"|",    
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA009) CLIPPED,"|",    
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA010) CLIPPED,"|",    
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA011) CLIPPED,"|",    
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA012) CLIPPED,"|",    
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA013) CLIPPED,"|",    
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA014) CLIPPED,"|",    
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA015) CLIPPED,"|",    
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA016) CLIPPED,"|",    
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA017) CLIPPED,"|",    
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA018) CLIPPED,"|",    
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA019) CLIPPED,"|",    
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA020) CLIPPED,"|",
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA021) CLIPPED,"|",
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA022) CLIPPED,"|",
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA023) CLIPPED,"|",
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA024) CLIPPED,"|",
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA025) CLIPPED,"|",
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA026) CLIPPED,"|",
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA027) CLIPPED,"|",
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA028) CLIPPED,"|",
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA029) CLIPPED,"|",
                  sadzi140_exim_replace_escape_character(lo_DATA_LINE[li_loop].DATA030) CLIPPED,"|",
                  "\n"
                  
    CALL lo_line_string_buf.append(ls_Text)
    
  END FOR

  LET ls_Return = lo_line_string_buf.toString()
  CALL lo_line_string_buf.clear()
  
  RETURN ls_Return
  
END FUNCTION

#移除或標示特殊字元
FUNCTION sadzi140_exim_replace_escape_character(p_value)
DEFINE
  p_value STRING
DEFINE
  lo_str_buf base.StringBuffer
DEFINE
  ls_return STRING  

  LET lo_str_buf = base.StringBuffer.create()
  CALL lo_str_buf.Clear()

  CALL lo_str_buf.append(p_value)

  #有單引號的匯出時再加上一層, 主要在匯入時使用
  CALL lo_str_buf.replace("'","''",0)
  #移除換行符號
  CALL lo_str_buf.replace("\n","",0)

  CALL lo_str_buf.toString() RETURNING ls_return  

  RETURN ls_return
  
END FUNCTION 

FUNCTION sadzi140_exim_write_file(p_FileName,p_LineString)
DEFINE
  p_FileName   STRING,
  p_LineString STRING
DEFINE   
  ls_file_name      STRING,
  ls_LineString     STRING,
  lo_channel_write  base.Channel,
  lb_success        BOOLEAN

  LET ls_file_name  = p_FileName
  LET ls_LineString = p_LineString

  LET lb_success = TRUE
  
  LET  lo_channel_write = base.Channel.create()
  CALL lo_channel_write.setDelimiter("")

  TRY
    CALL lo_channel_write.openFile(ls_file_name, "w" )
    CALL lo_channel_write.write(ls_LineString)
    CALL lo_channel_write.close()
    DISPLAY cs_success_tag,"Data write to '",ls_file_name,"' successed !!"
  CATCH
    LET lb_success = FALSE
    DISPLAY cs_error_tag,"Data write to '",ls_file_name,"' failed !!"
  END TRY  

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_exec_SQL(p_SQL)
DEFINE
  p_SQL STRING
DEFINE
  ls_SQL     STRING,
  lb_Success BOOLEAN

  LET ls_SQL = p_SQL
  LET lb_Success = TRUE

  #BEGIN WORK

  TRY
    &ifdef is_pgs
    SET LOCK MODE TO WAIT
    SAVEPOINT sp_tran
    &endif
    PREPARE lpre_ExecSQL FROM ls_SQL
    EXECUTE lpre_ExecSQL
    #COMMIT WORK
    LET lb_Success = TRUE
  CATCH
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif
    #DISPLAY cs_error_tag,ls_SQL
    #ROLLBACK WORK
    LET lb_Success = FALSE
  END TRY  

  RETURN lb_Success
  
END FUNCTION

FUNCTION sadzi140_exim_show_help()

  DISPLAY "\n",
          "Usage : r.r ",ui.Interface.getName()," [iexp/iimp/iasm/idel/idrp/ialt/ireb] [Table Name] [Construct Version] [SD Version] [Request NO.]","\n",
          "\n",
          "Ex. ","\n",
          "    -- Export/Import version 2","\n",
          "    r.r ",ui.Interface.getName()," iexp test_t 2 0","\n",
          "    r.r ",ui.Interface.getName()," iimp test_t 2 0","\n",
          "    -- Export/Import r.t data only","\n",
          "    r.r ",ui.Interface.getName()," iexp test_t T 0","\n",
          "    r.r ",ui.Interface.getName()," iimp test_t T 0","\n",
          "    -- Create or modify table","\n",
          "    r.r ",ui.Interface.getName()," iasm test_t 1 0","\n",
          "    -- Delete current r.t data and corresponding snapshot and restore to version 1","\n",
          "    r.r ",ui.Interface.getName()," idel test_t 1 0","\n",
          "    -- Drop table and related datas","\n",
          "    r.r ",ui.Interface.getName()," idrp test_t","\n",
          "    -- Alter table and columns","\n",
          "    r.r ",ui.Interface.getName()," ialt test_t 2 3","\n",
          "    -- Rebuild table","\n",
          "    r.r ",ui.Interface.getName()," ireb test_t 2 1","\n"
          
  LET mb_fault = TRUE

END FUNCTION

FUNCTION sadzi140_exim_create_table(p_TableName,p_alm_version,p_alm_request_no,p_dgenv)
DEFINE
  p_TableName      STRING,
  p_alm_version    STRING,
  p_alm_request_no STRING,
  p_dgenv          STRING 
DEFINE
  ls_TableName      STRING,
  ls_alm_version    STRING, 
  ls_alm_request_no STRING,
  ls_dgenv          STRING 
DEFINE
  ls_MasterDB    STRING,
  ls_MasterUser  STRING,
  ls_all_message STRING,
  ls_version     STRING,
  ls_message     STRING,
  lb_error       BOOLEAN,
  lb_diff        BOOLEAN,
  lo_CreateTable T_DZEA_CREATE_TABLE
  #lo_SchemaType  DYNAMIC ARRAY OF T_TABLE_SYNONYM
DEFINE 
  lb_return BOOLEAN
  
  LET ls_TableName      = p_TableName
  LET ls_alm_version    = p_alm_version
  LET ls_alm_request_no = p_alm_request_no
  LET ls_dgenv          = p_dgenv

  LET ls_MasterDB    = ms_MasterDB
  LET ls_MasterUser  = ms_MasterUser
  
  ## 建立表格 ####################################################################
  CALL sadzi140_db_get_create_table_info(ls_TableName) RETURNING lo_CreateTable.*
  #CALL sadzi140_db_get_main_table_synonym_data(ls_TableName) RETURNING lo_SchemaType  
  LET lo_CreateTable.dct_master_db   = NVL(ls_MasterDB,cs_master_db)
  LET lo_CreateTable.dct_master_user = NVL(ls_MasterUser,cs_master_user)
  CALL sadzi140_util_create_real_table(ls_MasterUser,lo_CreateTable.*,ls_version,ls_alm_version,ls_alm_request_no,ls_dgenv,FALSE) RETURNING ls_all_message
  CALL sadzi140_util_make_alter_table(ls_MasterUser,ls_TableName,TRUE) RETURNING lb_error,ls_message
  ##############################################################################

  LET lb_return = NOT lb_error
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_exim_modify_table(p_TableName,p_ALMVersion)
DEFINE
  p_TableName  STRING,
  p_ALMVersion STRING 
DEFINE
  ls_TableName   STRING,
  ls_ALMVersion  STRING, 
  ls_MasterDB    STRING,
  ls_MasterUser  STRING,
  ls_all_message STRING,
  ls_version     STRING,
  ls_message     STRING,
  lb_diff        BOOLEAN,
  lb_success     BOOLEAN,
  lb_fault       BOOLEAN,
  lo_CreateTable T_DZEA_CREATE_TABLE
  #lo_SchemaType  DYNAMIC ARRAY OF T_TABLE_SYNONYM

  LET ls_TableName  = p_TableName
  LET ls_ALMVersion = p_ALMVersion

  LET ls_MasterDB    = ms_MasterDB
  LET ls_MasterUser  = ms_MasterUser
  
  ## 建立表格 ####################################################################
  CALL sadzi140_db_get_create_table_info(ls_TableName) RETURNING lo_CreateTable.*
  #CALL sadzi140_db_get_main_table_synonym_data(ls_TableName) RETURNING lo_SchemaType  
  LET lo_CreateTable.dct_master_db   = NVL(ls_MasterDB,cs_master_db)
  LET lo_CreateTable.dct_master_user = NVL(ls_MasterUser,cs_master_user)
  CALL sadzi140_util_make_alter_table(ls_MasterUser,ls_TableName,TRUE) RETURNING lb_fault,ls_message
  ##############################################################################

  LET lb_success = NOT lb_fault
  
  RETURN lb_success
  
END FUNCTION

#TAR資料
FUNCTION sadzi140_exim_tar_file(po_PARAMETERS)
DEFINE
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  lo_PARAMETERS       T_DZLM_T_SCR,  
  ls_TARName          STRING,
  ls_TARString        STRING,
  ls_os_separator     STRING,
  lb_Copyfile         BOOLEAN,
  ls_redirect         STRING, #20200330 
  ls_result           STRING, #20200330  
  ls_source_full_name STRING,
  ls_dest_full_name   STRING,
  ls_curr_path        STRING
DEFINE
  lb_return BOOLEAN,
  ls_return STRING    

  LET lo_PARAMETERS.* = po_PARAMETERS.*

  #取得作業系統的分隔符號
  CALL os.Path.separator() RETURNING ls_os_separator 
  
  LET lb_Copyfile = TRUE
  
  LET ls_TARName = lo_PARAMETERS.DZLM002,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING cs_UsingFormat),".tdi"
                    
  LET ls_redirect = sadzi140_util_get_direct_to_null_sign() #20200330 
  CALL os.Path.pwd() RETURNING ls_curr_path #20200330
  LET ls_TARString = "tar cvf ",ls_TARName," *.unl *.txn *.json" #20171120 add json
  LET ls_TARString = ls_TARString,ls_redirect #20200330 
  CALL sadzi140_util_run_command(ls_TARString,TRUE) RETURNING ls_result #20200330  
  --RUN ls_TARString    

  LET mo_export_info.TAR_NAME = ls_TARName
  
  #若為lmtb模式, 才搬移匯出檔到工作目錄
  IF mb_backend_mode THEN 
    LET ls_curr_path = os.Path.pwd()
    LET ls_source_full_name = ls_curr_path,ls_os_separator,ls_TARName
    LET ls_dest_full_name   = ms_entry_path,ls_os_separator,ls_TARName
    #來源和目的路徑不同才搬
    IF ls_curr_path <> ms_entry_path THEN 
      CALL os.Path.copy(ls_source_full_name,ls_dest_full_name) RETURNING lb_Copyfile
    END IF  
  END IF 

  LET lb_return = lb_Copyfile
  LET ls_return = ls_TARName
  
  RETURN lb_return,ls_return 
  
END FUNCTION

#UNTAR資料
FUNCTION sadzi140_exim_untar_file(po_PARAMETERS)
DEFINE
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  lo_PARAMETERS       T_DZLM_T_SCR,  
  ls_TARName          STRING,
  ls_TARString        STRING,
  lb_Copyfile         BOOLEAN,
  ls_os_separator     STRING,
  lb_success          BOOLEAN,
  li_run_result       INTEGER,
  ls_redirect         STRING, #20200330
  ls_result           STRING, #20200330 
  ls_source_full_name STRING,
  ls_dest_full_name   STRING,
  ls_curr_path        STRING

  LET lo_PARAMETERS.* = po_PARAMETERS.*

  #取得作業系統的分隔符號
  CALL os.Path.separator() RETURNING ls_os_separator 
  LET lb_Copyfile = TRUE
  LET lb_success = TRUE
  
  #UNTAR資料
  LET ls_TARName = lo_PARAMETERS.DZLM002,"+TABLE+",
                   lo_PARAMETERS.DZLM012,"+",
                   (lo_PARAMETERS.DZLM015 USING cs_UsingFormat),".tdi"
                   
  #若為lmtb模式, 才搬移匯入檔到工作目錄
  IF mb_backend_mode THEN 
    LET ls_curr_path = os.Path.pwd()
    LET ls_source_full_name = ms_entry_path,ls_os_separator,ls_TARName
    LET ls_dest_full_name   = ls_curr_path,ls_os_separator,ls_TARName
    #來源和目的路徑不同才搬
    IF ls_curr_path <> ms_entry_path THEN 
      CALL os.Path.copy(ms_entry_path||ls_os_separator||ls_TARName,ls_TARName) RETURNING lb_success
    END IF  
  END IF   
  IF NOT lb_success THEN GOTO _exit_untar END IF 
  
  LET ls_redirect = sadzi140_util_get_direct_to_null_sign() #20200330 
  LET ls_TARString = "tar xvfm ",ls_TARName #20210705
  LET ls_TARString = ls_TARString,ls_redirect #20200330 
  CALL sadzi140_util_run_command(ls_TARString,TRUE) RETURNING ls_result #20200330  
  --RUN ls_TARString RETURNING li_run_result
  LET lb_success = IIF(li_run_result==0,TRUE,FALSE)
  IF NOT lb_success THEN GOTO _exit_untar END IF 
  
  LABEL _exit_untar:
  
  RETURN lb_success 
  
END FUNCTION

## GZTD_T ######################################################################

FUNCTION sadzi140_exim_export_GZTD_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "GZTD_T"
  
  CALL sadzi140_exim_get_GZTD_data(ls_TableName,p_json_obj) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_GZTD_data(p_TableName,p_json_obj)
DEFINE
  p_TableName  STRING,
  p_json_obj   util.JSONObject
DEFINE
  ls_TableName  STRING,
  lo_GZTD_T     DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName = p_TableName

  INITIALIZE lo_GZTD_T TO NULL  #20180213
  INITIALIZE lo_RETURN TO NULL  #20180213

  #尚未匯出共用欄位, 因未來會再作修改
  LET ls_SQL = "select td.gztd001,td.gztd002,td.gztd003,td.gztd004,td.gztd005,",
               "       td.gztd006,td.gztd007,td.gztd008,td.gztd009,td.gztd010,",
               "       td.gztd011,td.gztd012,td.gztd013                       ", 
               "  from gztd_t td                                              ",
               " where 1=1                                                    ",
               "   and exists (                                               ", -- 20170419 begin
               "                select 1                                      ",
               "                  from dzeb_t eb                              ",
               "                 where eb.dzeb006 = td.gztd001                ",
               "                   and eb.dzeb001 = '",ls_TableName,"'        ",
               "              )                                               ",
               " order by td.gztd001                                          "  -- 20170419 end
               
  #取得Table export notice             
  LET ls_tsn_SQL = "select 'GZTD_T' table_name,count(1) records            ",
                   "  from gztd_t td                                       ",
                   " where 1=1                                             ",
                   "   and exists (                                        ", -- 20170419 begin
                   "                select 1                               ",
                   "                  from dzeb_t eb                       ",
                   "                 where eb.dzeb006 = td.gztd001         ",
                   "                   and eb.dzeb001 = '",ls_TableName,"' ",
                   "              )                                        "  -- 20170419 end
                 
  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)
  
  PREPARE lpre_GetGZTDData FROM ls_SQL
  DECLARE lcur_GetGZTDData CURSOR FOR lpre_GetGZTDData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetGZTDData
  FOREACH lcur_GetGZTDData INTO lo_GZTD_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetGZTDData
  CALL lo_GZTD_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetGZTDData
  FREE lcur_GetGZTDData  

  LET lo_RETURN = lo_GZTD_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_import_GZTD_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "GZTD_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_GZTD_T_data(lo_DATAS) RETURNING lb_success
    END IF   
  END IF
  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_GZTD_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL THEN
      LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "'",lo_DATAS[li_loop,3],"',",
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",lo_DATAS[li_loop,5],"',",
                           "'",lo_DATAS[li_loop,6],"',",
                           "'",lo_DATAS[li_loop,7],"',",
                           "'",lo_DATAS[li_loop,8],"',",
                           "'",lo_DATAS[li_loop,9],"',",
                           "",NVL(lo_DATAS[li_loop,10],"NULL"),",",
                           "",NVL(lo_DATAS[li_loop,11],"NULL"),",",
                           "'",lo_DATAS[li_loop,12],"',",
                           "'",lo_DATAS[li_loop,13],"'"
                           
      LET ls_InsertSQL = "insert into gztd_t (                                         ",
                         "                    gztd001,gztd002,gztd003,gztd004,gztd005, ",
                         "                    gztd006,gztd007,gztd008,gztd009,gztd010, ",
                         "                    gztd011,gztd012,gztd013                  ", 
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "
                         
      LET ls_UpdateValue = "gztd002 = '",lo_DATAS[li_loop,2],"',",
                           "gztd003 = '",lo_DATAS[li_loop,3],"',",
                           "gztd004 = '",lo_DATAS[li_loop,4],"',",
                           "gztd005 = '",lo_DATAS[li_loop,5],"',",
                           "gztd006 = '",lo_DATAS[li_loop,6],"',",
                           "gztd007 = '",lo_DATAS[li_loop,7],"',",
                           "gztd008 = '",lo_DATAS[li_loop,8],"',",
                           "gztd009 = '",lo_DATAS[li_loop,9],"',",
                           "gztd010 = ",NVL(lo_DATAS[li_loop,10],"NULL"),",",
                           "gztd011 = ",NVL(lo_DATAS[li_loop,11],"NULL"),",",
                           "gztd012 = '",lo_DATAS[li_loop,12],"',",
                           "gztd013 = '",lo_DATAS[li_loop,13],"'"
                           
      LET ls_UpdateSQL = "update gztd_t set ",
                         ls_UpdateValue,
                         " where gztd001 = '",lo_DATAS[li_loop,1],"'"
                         
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      
      IF NOT lb_success THEN
        LET lb_success = TRUE
        #gztd_t的update暫時不做 
        #CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
      
    END IF
  END FOR 

  RETURN lb_success
  
END FUNCTION

## GZTDL_T ######################################################################

FUNCTION sadzi140_exim_export_GZTDL_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "GZTDL_T"
  
  CALL sadzi140_exim_get_GZTDL_data(ls_TableName,p_json_obj) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_GZTDL_data(p_TableName,p_json_obj)
DEFINE
  p_TableName  STRING,
  p_json_obj   util.JSONObject
DEFINE
  ls_TableName  STRING,
  lo_GZTDL_T     DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName = p_TableName

  INITIALIZE lo_GZTDL_T TO NULL  #20180213
  INITIALIZE lo_RETURN TO NULL   #20180213

  #尚未匯出共用欄位, 因未來會再作修改
  LET ls_SQL = "select tdl.gztdl001,tdl.gztdl002,tdl.gztdl003,tdl.gztdl004,tdl.gztdl005 ",
               "  from gztdl_t tdl                                                      ",
               " where 1=1                                                              ",
               "   and exists (                                                         ", -- 20170419 begin
               "                select 1                                                ",
               "                  from dzeb_t eb                                        ",
               "                 where eb.dzeb006 = tdl.gztdl001                        ",
               "                   and eb.dzeb001 = '",ls_TableName,"'                  ",
               "              )                                                         ",
               " order by tdl.gztdl001,tdl.gztdl002                                     "  -- 20170419 end
               
  #取得Table export notice             
  LET ls_tsn_SQL = "select 'GZTDL_T' table_name,count(1) records           ",
                   "  from gztdl_t tdl                                     ",
                   " where 1=1                                             ",
                   "   and exists (                                        ", -- 20170419 begin
                   "                select 1                               ",
                   "                  from dzeb_t eb                       ",
                   "                 where eb.dzeb006 = tdl.gztdl001       ",
                   "                   and eb.dzeb001 = '",ls_TableName,"' ",
                   "              )                                        "  -- 20170419 end
                 
  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)
  
  PREPARE lpre_GetGZTDLData FROM ls_SQL
  DECLARE lcur_GetGZTDLData CURSOR FOR lpre_GetGZTDLData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetGZTDLData
  FOREACH lcur_GetGZTDLData INTO lo_GZTDL_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetGZTDLData
  CALL lo_GZTDL_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetGZTDLData
  FREE lcur_GetGZTDLData  

  LET lo_RETURN = lo_GZTDL_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_import_GZTDL_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "GZTDL_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_GZTDL_T_data(lo_DATAS) RETURNING lb_success
    END IF   
  END IF
  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_GZTDL_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  lb_is_lang      BOOLEAN,
  li_lng          INTEGER, 
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN
    
      #從清單中取得可匯入的語言別
      LET lb_is_lang = FALSE
      FOR li_lng = 1 TO mo_local_lang.getLength()
        IF lo_DATAS[li_loop,2] = mo_local_lang[li_lng] THEN
          LET lb_is_lang = TRUE
          EXIT FOR
        END IF 
      END FOR 

      #符合當地多語言清單的資料才匯入
      IF lb_is_lang THEN
        LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                             "'",lo_DATAS[li_loop,2],"',",
                             "'",lo_DATAS[li_loop,3],"',",
                             "'",lo_DATAS[li_loop,4],"',",
                             "'",lo_DATAS[li_loop,5],"'"
                             
        LET ls_InsertSQL = "insert into gztdl_t (                                              ",
                           "                      gztdl001,gztdl002,gztdl003,gztdl004,gztdl005 ",
                           "                    )                                              ",  
                           " values (                                                          ",
                           ls_InsertValue,
                           "        )                                                          "
                           
        LET ls_UpdateValue = "gztdl003 = '",lo_DATAS[li_loop,3],"',",
                             "gztdl004 = '",lo_DATAS[li_loop,4],"',",
                             "gztdl005 = '",lo_DATAS[li_loop,5],"'"
                             
        LET ls_UpdateSQL = "update gztdl_t set ",
                           ls_UpdateValue,
                           " where gztdl001 = '",lo_DATAS[li_loop,1],"'",
                           "   and gztdl002 = '",lo_DATAS[li_loop,2],"'"
                           
        CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
        IF NOT lb_success THEN
          CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
          IF NOT lb_success THEN
            DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
            DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
            EXIT FOR 
          END IF
        END IF
      END IF  
    END IF
  END FOR 

  RETURN lb_success
  
END FUNCTION

## GZCC_T ######################################################################

FUNCTION sadzi140_exim_export_GZCC_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject,
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATA_LINE        DYNAMIC ARRAY OF T_DATA_LINE,
  ls_UsingFormat      STRING,
  ls_DATA_LINE        STRING,
  ls_EXPORT_TABLE     STRING,
  lb_Success          BOOLEAN

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "GZCC_T"
  
  CALL sadzi140_exim_get_GZCC_data(ls_TableName,p_json_obj) RETURNING lo_DATA_LINE
  CALL sadzi140_exim_export_data(lo_DATA_LINE) RETURNING ls_DATA_LINE 
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  CALL sadzi140_exim_write_file(ls_FileName,ls_DATA_LINE) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_GZCC_data(p_TableName,p_json_obj)
DEFINE
  p_TableName  STRING,
  p_json_obj   util.JSONObject
DEFINE
  ls_TableName  STRING,
  lo_GZCC_T     DYNAMIC ARRAY OF T_DATA_LINE
DEFINE  
  ls_SQL     STRING,
  ls_tsn_SQL STRING, 
  li_RecCnt  INTEGER,
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  lo_RETURN  DYNAMIC ARRAY OF T_DATA_LINE

  LET ls_TableName = p_TableName

  INITIALIZE lo_GZCC_T TO NULL  #20180213
  INITIALIZE lo_RETURN TO NULL  #20180213

  #尚未匯出共用欄位, 因未來會再作修改
  LET ls_SQL = "select cc.gzcc001,cc.gzcc002,cc.gzcc003,cc.gzcc004,cc.gzcc005,",
               "       cc.gzcc006,cc.gzccstus                                 ",
               "  from gzcc_t cc                                              ",
               " where 1=1                                                    ",
               "   and cc.gzcc001 = '",ls_TableName,"'                        "
               
  #取得Table export notice             
  LET ls_tsn_SQL = "select 'GZCC_T' table_name,count(1) records  ",
                   "  from gzcc_t cc                             ",
                   " where 1=1                                   ",
                   "   and cc.gzcc001 = '",ls_TableName,"'       "
                 
  CALL sadzi140_exim_get_table_ship_notice(ls_tsn_SQL) RETURNING lo_table_ship_notice.*       
  CALL p_json_obj.put(lo_table_ship_notice.TABLE_NAME,lo_table_ship_notice.RECORDS)
  
  PREPARE lpre_GetGZCCData FROM ls_SQL
  DECLARE lcur_GetGZCCData CURSOR FOR lpre_GetGZCCData

  LET li_RecCnt = 1 
  
  OPEN lcur_GetGZCCData
  FOREACH lcur_GetGZCCData INTO lo_GZCC_T[li_RecCnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_RecCnt = li_RecCnt + 1

  END FOREACH
  CLOSE lcur_GetGZCCData
  CALL lo_GZCC_T.deleteElement(li_RecCnt)
  
  FREE lpre_GetGZCCData
  FREE lcur_GetGZCCData  

  LET lo_RETURN = lo_GZCC_T  #20180213
  
  RETURN lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_import_GZCC_data(p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "GZCC_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_GZCC_T_data(lo_DATAS) RETURNING lb_success
    END IF   
  END IF
  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_GZCC_T_data(p_DATAS)
DEFINE
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  lb_result       BOOLEAN, #20171120
  ls_IMPORT_TABLE STRING, #20180302 
  li_loop         INTEGER
  
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  LET ls_IMPORT_TABLE = "GZCC_T" #20180302

  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL THEN
      #20171120 begin  
      IF mb_pickup_exist THEN 
        CALL sadzi140_exim_check_if_exist_in_pick_list(mo_pick_col_list,lo_DATAS[li_loop,2]) RETURNING lb_result
        IF NOT lb_result THEN CONTINUE FOR END IF
      END IF
      #20171120 end
      
      #20180302 begin  
      IF mo_local_sys_info.psi_RT_PATCH_MODE <> "Y" THEN 
        CALL sadzi140_exim_check_if_exist_in_deleted_std_col_list(mo_deleted_std_col_list,lo_DATAS[li_loop,2]) RETURNING lb_result
        IF lb_result THEN 
          DISPLAY cs_warning_tag,"Column ",lo_DATAS[li_loop,2]," exists in deleted standard column list, skip import to ",ls_IMPORT_TABLE,".","\n",
                                 "You can using 'Tools'->'Field selection import' to import it !" 
          CONTINUE FOR 
        END IF
      END IF  
      #20180302 end
      
      LET ls_InsertValue = "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           " ",NVL(lo_DATAS[li_loop,3],"NULL"),", ",
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",lo_DATAS[li_loop,5],"',",
                           " ",NVL(lo_DATAS[li_loop,6],"NULL"),", ",
                           "'",lo_DATAS[li_loop,7],"' "
                           
      LET ls_InsertSQL = "insert into gzcc_t (                                         ",
                         "                    gzcc001,gzcc002,gzcc003,gzcc004,gzcc005, ",
                         "                    gzcc006,gzccstus                         ",
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "
                         
      LET ls_UpdateValue = "gzcc002  = '",lo_DATAS[li_loop,2],"',",
                           "gzcc003  =  ",NVL(lo_DATAS[li_loop,3],"NULL"),", ",
                           "gzcc005  = '",lo_DATAS[li_loop,5],"',",
                           "gzcc006  =  ",NVL(lo_DATAS[li_loop,6],"NULL"),", ",
                           "gzccstus = '",lo_DATAS[li_loop,7],"' "
                           
      LET ls_UpdateSQL = "update gzcc_t set ",
                         ls_UpdateValue,
                         " where gzcc001 = '",lo_DATAS[li_loop,1],"'",
                         "   and gzcc004 = '",lo_DATAS[li_loop,4],"'"
                         
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      LET lb_success = TRUE

      {
      #因應Merge因此不更新 2014.10.21
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
      }
      
    END IF
  END FOR 

  RETURN lb_success
  
END FUNCTION

#20160519 begin  
## DZHA_T for DZEA_T ###########################################################

FUNCTION sadzi140_exim_import_DZHA_data(p_GUID,p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_GUID         STRING,
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_GUID             STRING,
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER

  LET ls_GUID         = p_GUID
  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEA_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZHA_T_data(ls_GUID,lo_DATAS) RETURNING lb_success
    END IF  
  END IF
  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZHA_T_data(p_GUID,p_DATAS)
DEFINE
  p_GUID  STRING,
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  ls_GUID         STRING,
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  ls_user         VARCHAR(100),
  ldt_datetime    DATETIME YEAR TO SECOND,
  li_loop         INTEGER
  
  LET ls_GUID    = p_GUID 
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  &ifndef DEBUG
  LET ls_user = g_user
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  &endif
  LET ldt_datetime = CURRENT YEAR TO SECOND
  
  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL THEN
      LET ls_InsertValue = "'",ls_GUID,"',",
                           "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "'",lo_DATAS[li_loop,3],"',",
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",lo_DATAS[li_loop,5],"',",
                           "'",lo_DATAS[li_loop,6],"',",
                           "'",lo_DATAS[li_loop,7],"',",
                           "'",lo_DATAS[li_loop,8],"',",
                           "'",lo_DATAS[li_loop,9],"',",
                           "'",lo_DATAS[li_loop,10],"',",
                           "'",lo_DATAS[li_loop,11],"',",
                           "'",lo_DATAS[li_loop,12],"',",
                           "'",lo_DATAS[li_loop,13],"',", #異動序號
                           "'",lo_DATAS[li_loop,14],"',",
                           "'",lo_DATAS[li_loop,15],"',",
                           "'",lo_DATAS[li_loop,16],"',",
                           "'",lo_DATAS[li_loop,17],"',",
                           "'",lo_DATAS[li_loop,18],"',",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,",",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,""

      LET ls_InsertSQL = "insert into dzha_t (                                         ",
                         "                    dzha000,                                 ", 
                         "                    dzha001,dzha002,dzha003,dzha004,dzha005, ",
                         "                    dzha006,dzha007,dzha008,dzha009,dzha010, ",
                         "                    dzha011,dzha012,dzha013,dzha014,dzha015, ",
                         "                    dzha016,dzha017,dzha018,                 ",
                         "                    dzhacrtid,dzhacrtdt,dzhamodid,dzhamoddt  ", 
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "

      LET ls_UpdateValue = "dzha002 = '",lo_DATAS[li_loop,2],"',",
                           "dzha003 = '",lo_DATAS[li_loop,3],"',",
                           "dzha004 = '",lo_DATAS[li_loop,4],"',",
                           "dzha005 = '",lo_DATAS[li_loop,5],"',",
                           "dzha006 = '",lo_DATAS[li_loop,6],"',",
                           "dzha007 = '",lo_DATAS[li_loop,7],"',",
                           "dzha008 = '",lo_DATAS[li_loop,8],"',",
                           "dzha009 = '",lo_DATAS[li_loop,9],"',",
                           "dzha010 = '",lo_DATAS[li_loop,10],"',",
                           "dzha011 = '",lo_DATAS[li_loop,11],"',",
                           "dzha012 = '",lo_DATAS[li_loop,12],"',",
                           "dzha013 = '",lo_DATAS[li_loop,13],"',",
                           "dzha014 = '",lo_DATAS[li_loop,14],"',",
                           "dzha015 = '",lo_DATAS[li_loop,15],"',",
                           "dzha016 = '",lo_DATAS[li_loop,16],"',",
                           "dzha017 = '",lo_DATAS[li_loop,17],"',",
                           "dzha018 = '",lo_DATAS[li_loop,18],"'"
                           
      LET ls_UpdateSQL = "update dzha_t set ",
                         ls_UpdateValue,
                         " where dzha000 = '",ls_GUID,"'",
                         "   and dzha001 = '",lo_DATAS[li_loop,1],"'"
                         
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
                         
    END IF
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZHAL_T for DZEAL_T #########################################################

FUNCTION sadzi140_exim_import_DZHAL_data(p_GUID,p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_GUID         STRING,
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_GUID             STRING,
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_GUID         = p_GUID
  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEAL_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZHAL_T_data(ls_GUID,lo_DATAS) RETURNING lb_success
    END IF   
  END IF   

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZHAL_T_data(p_GUID,p_DATAS)
DEFINE
  p_GUID  STRING,
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  ls_GUID         STRING,
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  lb_is_lang      BOOLEAN,
  li_lng          INTEGER,   
  li_loop         INTEGER

  LET ls_GUID    = p_GUID
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN

      #從清單中取得可匯入的語言別
      LET lb_is_lang = FALSE
      FOR li_lng = 1 TO mo_local_lang.getLength()
        IF lo_DATAS[li_loop,2] = mo_local_lang[li_lng] THEN
          LET lb_is_lang = TRUE
          EXIT FOR
        END IF 
      END FOR 

      #符合當地多語言清單的資料才匯入
      IF lb_is_lang THEN
      
        LET ls_InsertValue = "'",ls_GUID,"',",
                             "'",lo_DATAS[li_loop,1],"',",
                             "'",NVL(lo_DATAS[li_loop,2],"NULL"),"',",
                             "'",lo_DATAS[li_loop,3],"',",
                             "'",lo_DATAS[li_loop,4],"',",
                             "'",lo_DATAS[li_loop,5],"'"
                                                      
        LET ls_InsertSQL = "insert into dzhal_t (                                              ",
                           "                      dzhal000,                                    ",
                           "                      dzhal001,dzhal002,dzhal003,dzhal004,dzhal005 ",
                           "                    )                                              ",  
                           " values (                                                          ",
                           ls_InsertValue,
                           "        )                                                          "  
                           
        CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          EXIT FOR 
        END IF
      END IF      
    END IF                       
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZHB_T for DZEB_T ###########################################################

FUNCTION sadzi140_exim_import_DZHB_data(p_GUID,p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_GUID         STRING,
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_GUID             STRING,
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_result           BOOLEAN, #20200330
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_GUID         = p_GUID
  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEB_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZHB_T_data(ls_GUID,lo_DATAS) RETURNING lb_success
      CALL sadzi140_exim_set_DZHB_T_datatype(ls_GUID,ls_TableName) RETURNING lb_result #20200330
    END IF   
  END IF  

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZHB_T_data(p_GUID,p_DATAS)
DEFINE
  p_GUID  STRING,
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  ls_GUID         STRING,
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  ls_user         VARCHAR(100),
  ls_orig_coltype STRING,
  ls_column_type  STRING,
  li_sequence     INTEGER,
  ls_sequence     STRING,
  ls_ship_notice  STRING,
  lo_property     T_PROPERTY,
  ldt_datetime    DATETIME YEAR TO SECOND,
  li_loop         INTEGER
  
  LET ls_GUID    = p_GUID
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  &ifndef DEBUG
  LET ls_user = g_user
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  &endif
  LET ldt_datetime = CURRENT YEAR TO SECOND
  
  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN
    
      #相同的ZONE就用相同的序號 
      IF mb_same_dgenv OR mb_new_table THEN
        LET ls_sequence = lo_DATAS[li_loop,10]
      ELSE
        #取得Sequnece ID  
        CALL sadzi140_exim_get_column_sequence(lo_DATAS[li_loop,1],lo_DATAS[li_loop,2]) RETURNING li_sequence
        LET ls_sequence = li_sequence
        LET ls_sequence = ls_sequence.trim()
      END IF 

      #取得欄位型態, 當舊型態和新型態不同時才重取對應的資料型態值(已出貨欄位才做判斷)
      LET ls_ship_notice = lo_DATAS[li_loop,18]
      IF ls_ship_notice = cs_column_shipped THEN
        LET ls_orig_coltype = lo_DATAS[li_loop,6]
        CALL sadzi140_exim_get_column_type(lo_DATAS[li_loop,1],lo_DATAS[li_loop,2],ls_orig_coltype) RETURNING ls_column_type
        IF ls_column_type <> ls_orig_coltype THEN
          CALL sadzi140_util_get_property(ls_column_type,ms_dlang) RETURNING lo_property.*
          LET lo_DATAS[li_loop,6] = ls_column_type  
          LET lo_DATAS[li_loop,7] = lo_property.d_type
          LET lo_DATAS[li_loop,8] = lo_property.d_length
        END IF
      END IF  

      LET ls_InsertValue = "'",ls_GUID,"',",
                           "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "'",lo_DATAS[li_loop,3],"',",
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",lo_DATAS[li_loop,5],"',",
                           "'",lo_DATAS[li_loop,6],"',",
                           "'",lo_DATAS[li_loop,7],"',",
                           "'",lo_DATAS[li_loop,8],"',",
                           "'",sadzi140_exim_trim(lo_DATAS[li_loop,9]),"',",
                           "",NVL(ls_sequence,"NULL"),",",
                           "'",lo_DATAS[li_loop,11],"',",
                           "",NVL(lo_DATAS[li_loop,12],"NULL"),",",
                           "'",lo_DATAS[li_loop,13],"',",
                           "'",lo_DATAS[li_loop,14],"',",
                           "'",lo_DATAS[li_loop,15],"',",
                           "'",lo_DATAS[li_loop,16],"',",
                           "'",lo_DATAS[li_loop,17],"',",
                           "'",lo_DATAS[li_loop,18],"',",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,",",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,""
                                                    
      LET ls_InsertSQL = "insert into dzhb_t (                                         ",
                         "                    dzhb000,                                 ", 
                         "                    dzhb001,dzhb002,dzhb003,dzhb004,dzhb005, ",
                         "                    dzhb006,dzhb007,dzhb008,dzhb012,dzhb021, ",
                         "                    dzhb022,dzhb023,dzhb024,dzhb027,dzhb028, ",
                         "                    dzhb029,dzhb030,dzhb031,                 ",
                         "                    dzhbcrtid,dzhbcrtdt,dzhbmodid,dzhbmoddt  ",
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "

      LET ls_UpdateValue = "dzhb003 = '",lo_DATAS[li_loop,3],"',",
                           "dzhb004 = '",lo_DATAS[li_loop,4],"',",
                           "dzhb005 = '",lo_DATAS[li_loop,5],"',",
                           "dzhb006 = '",lo_DATAS[li_loop,6],"',",
                           "dzhb007 = '",lo_DATAS[li_loop,7],"',",
                           "dzhb008 = '",lo_DATAS[li_loop,8],"',",
                           "dzhb012 = '",sadzi140_exim_trim(lo_DATAS[li_loop,9]),"',",
                           "dzhb021 = ",NVL(ls_sequence,"NULL"),",",
                           "dzhb022 = '",lo_DATAS[li_loop,11],"',",
                           "dzhb023 = ",NVL(lo_DATAS[li_loop,12],"NULL"),",",
                           "dzhb024 = '",lo_DATAS[li_loop,13],"',",
                           "dzhb027 = '",lo_DATAS[li_loop,14],"',",
                           "dzhb028 = '",lo_DATAS[li_loop,15],"',",
                           "dzhb029 = case dzhb029 when '",cs_dgenv_customize,"' then dzhb029 else '",lo_DATAS[li_loop,16],"' end,",
                           "dzhb030 = '",lo_DATAS[li_loop,17],"',",
                           "dzhb031 = '",lo_DATAS[li_loop,18],"'"
                           
      LET ls_UpdateSQL = "update dzhb_t set ",
                         ls_UpdateValue,
                         " where dzhb000 = '",ls_GUID,"'",
                         "   and dzhb001 = '",lo_DATAS[li_loop,1],"'",
                         "   and dzhb002 = '",lo_DATAS[li_loop,2],"'"
                           
                         
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
                         
    END IF
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZHBL_T for DZEBL_T #########################################################

FUNCTION sadzi140_exim_import_DZHBL_data(p_GUID,p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_GUID         STRING, 
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_GUID             STRING, 
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_GUID         = p_GUID
  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEBL_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZHBL_T_data(ls_GUID,lo_DATAS) RETURNING lb_success
    END IF  
  END IF  

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZHBL_T_data(p_GUID,p_DATAS)
DEFINE
  p_GUID  STRING,
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  ls_GUID         STRING,
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  lb_is_lang      BOOLEAN,
  li_lng          INTEGER,
  li_loop         INTEGER,
  lb_is_udfc      BOOLEAN
  
  LET ls_GUID    = p_GUID
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN
    
      #從清單中取得可匯入的語言別
      LET lb_is_lang = FALSE
      FOR li_lng = 1 TO mo_local_lang.getLength()
        IF lo_DATAS[li_loop,2] = mo_local_lang[li_lng] THEN
          LET lb_is_lang = TRUE
          EXIT FOR
        END IF 
      END FOR 

      #符合當地多語言清單的資料才匯入
      IF lb_is_lang THEN
      
        LET ls_InsertValue = "'",ls_GUID,"',",
                             "'",lo_DATAS[li_loop,1],"',",
                             "'",lo_DATAS[li_loop,2],"',",
                             "'",lo_DATAS[li_loop,3],"',",
                             "'",lo_DATAS[li_loop,4],"',",
                             "'",lo_DATAS[li_loop,5],"'"
                             
        LET ls_InsertSQL = "insert into dzhbl_t (                                              ",
                           "                      dzhbl000,                                    ",
                           "                      dzhbl002,dzhbl003,dzhbl004,dzhbl005,dzhbl001 ",
                           "                    )                                              ",  
                           " values (                                                          ",
                           ls_InsertValue,
                           "        )                                                          "
                           
        CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
        END IF
      END IF
    END IF   
  END FOR 

  RETURN lb_success
  
END FUNCTION

#20200330 begin
FUNCTION sadzi140_exim_set_DZHB_T_datatype(p_guid,p_table_name)
DEFINE
  p_guid        STRING, 
  p_table_name  STRING
DEFINE
  ls_guid         STRING, 
  ls_table_name   STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN
  
  LET ls_guid       = p_guid
  LET ls_table_name = p_table_name
  
  LET lb_success = TRUE

  &ifdef is_msv
  LET ls_UpdateSQL = "update dzhb_t                                      ",
                     "   set dzhb007 = t2.gztd003,                       ",
                     "       dzhb008 = t2.gztd008                        ",
                     "  from (                                           ",
                     "         select gztd001,gztd003,gztd008            ",
                     "           from gztd_t                             ", 
                     "       ) t2                                        ", 
                     " where dzhb006 = t2.gztd001                        ",
                     "   and dzhb000 = '",ls_guid,"'                     ",
                     "   and dzhb001 = '",ls_table_name.toLowerCase(),"' "
  &else
  &ifdef is_pgs
  LET ls_UpdateSQL = "update dzhb_t                                      ",
                     "   set dzhb007 = t2.gztd003,                       ",
                     "       dzhb008 = t2.gztd008                        ",
                     "  from (                                           ",
                     "         select gztd001,gztd003,gztd008            ",
                     "           from gztd_t                             ", 
                     "       ) t2                                        ", 
                     " where dzhb006 = t2.gztd001                        ",
                     "   and dzhb000 = '",ls_guid,"'                     ",
                     "   and dzhb001 = '",ls_table_name.toLowerCase(),"' "
  &else
  LET ls_UpdateSQL = "update dzhb_t                                               ", 
                     "    set (dzhb007,dzhb008) = (                               ",
                     "                              select td.gztd003,td.gztd008  ",
                     "                                from gztd_t td              ",
                     "                               where dzhb006 = td.gztd001   ",
                     "                            )                               ",
                     "  where 1=1                                                 ",
                     "   and dzhb000 = '",ls_guid,"'                              ",
                     "   and dzhb001 = '",ls_table_name.toLowerCase(),"'          "
  &endif
  &endif
  
  CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
  IF NOT lb_success THEN
    DISPLAY "[ UPDATE SQL ] ",ls_UpdateSQL,"\n" 
  END IF

  RETURN lb_success
  
END FUNCTION
#20200330 end

## DZHC_T for DZEC_T #########################################################

FUNCTION sadzi140_exim_import_DZHC_data(p_GUID,p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_GUID         STRING,
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_GUID             STRING,
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_GUID         = p_GUID
  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEC_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZHC_T_data(ls_GUID,lo_DATAS) RETURNING lb_success
    END IF   
  END IF  

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZHC_T_data(p_GUID,p_DATAS)
DEFINE
  p_GUID  STRING,
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  ls_GUID         STRING,
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  lb_success      BOOLEAN,
  ls_user         VARCHAR(100),
  ldt_datetime    DATETIME YEAR TO SECOND,
  li_loop         INTEGER
  
  LET ls_GUID    = p_GUID
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  &ifndef DEBUG
  LET ls_user = g_user
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  &endif
  LET ldt_datetime = CURRENT YEAR TO SECOND
  
  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN
    
      LET ls_InsertValue = "'",ls_GUID,"',",
                           "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "'",lo_DATAS[li_loop,3],"',",
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",lo_DATAS[li_loop,5],"',",
                           "'",lo_DATAS[li_loop,6],"',",
                           "'",lo_DATAS[li_loop,7],"',",
                           "'",lo_DATAS[li_loop,8],"',",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,",",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,""
                                                    
      LET ls_InsertSQL = "insert into dzhc_t (                                        ",
                         "                    dzhc000,                                ",
                         "                    dzhc001,dzhc002,dzhc003,dzhc004,dzhc005,",
                         "                    dzhc006,dzhc007,dzhc008,                ",
                         "                    dzhccrtid,dzhccrtdt,dzhcmodid,dzhcmoddt ",   
                         "                   )                                        ",  
                         " values (                                                   ",
                         ls_InsertValue,
                         "        )                                                   "
                         
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      IF NOT lb_success THEN
        DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
      END IF
    END IF
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZHD_T for DZED_T #########################################################

FUNCTION sadzi140_exim_import_DZHD_data(p_GUID,p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_GUID         STRING,
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_GUID             STRING,
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_GUID         = p_GUID
  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZED_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_error_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZHD_T_data(ls_GUID,lo_DATAS) RETURNING lb_success
    END IF   
  END IF  

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZHD_T_data(p_GUID,p_DATAS)
DEFINE
  p_GUID  STRING,
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  ls_GUID         STRING,
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  ls_user         VARCHAR(100),
  ldt_datetime    DATETIME YEAR TO SECOND,
  lb_success      BOOLEAN,
  li_loop         INTEGER
  
  LET ls_GUID    = p_GUID 
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  &ifndef DEBUG
  LET ls_user = g_user
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  &endif
  LET ldt_datetime = CURRENT YEAR TO SECOND
  
  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN
    
      LET ls_InsertValue = "'",ls_GUID,"',",
                           "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "'",lo_DATAS[li_loop,3],"',",
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",sadzi140_exim_trim(lo_DATAS[li_loop,5]),"',",
                           "'",sadzi140_exim_trim(lo_DATAS[li_loop,6]),"',",
                           "'",sadzi140_exim_trim(lo_DATAS[li_loop,7]),"',",
                           "'",sadzi140_exim_trim(lo_DATAS[li_loop,8]),"',",
                           "'",sadzi140_exim_trim(lo_DATAS[li_loop,9]),"',",
                           "'",sadzi140_exim_trim(lo_DATAS[li_loop,10]),"',",
                           "'",sadzi140_exim_trim(ls_user),"',",
                           "",cs_get_time_stamp,",",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,""
                           
                                                    
      LET ls_InsertSQL = "insert into dzhd_t (                                         ",
                         "                    dzhd000,                                 ",
                         "                    dzhd001,dzhd002,dzhd003,dzhd004,dzhd005, ",
                         "                    dzhd006,dzhd007,dzhd008,dzhd009,dzhd010, ",
                         "                    dzhdcrtid,dzhdcrtdt,dzhdmodid,dzhdmoddt  ",
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "
                         
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      IF NOT lb_success THEN
        DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
      END IF
    END IF    
  END FOR 

  RETURN lb_success
  
END FUNCTION

## DZHN_T for DZEN_T #########################################################

FUNCTION sadzi140_exim_import_DZHN_data(p_GUID,p_TableName,p_json_obj,po_PARAMETERS)
DEFINE
  p_GUID         STRING,
  p_TableName    STRING,
  p_json_obj     util.JSONObject, 
  po_PARAMETERS  T_DZLM_T_SCR  
DEFINE
  ls_GUID             STRING,
  ls_TableName        STRING,
  lo_PARAMETERS       T_DZLM_T_SCR,
  ls_FileName         STRING,
  lo_DATAS            DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_UsingFormat      STRING,
  ls_EXPORT_TABLE     STRING,
  lb_success          BOOLEAN,
  li_RecCount         INTEGER 

  LET ls_GUID         = p_GUID
  LET ls_TableName    = p_TableName
  LET lo_PARAMETERS.* = po_PARAMETERS.*

  LET ls_UsingFormat  = cs_UsingFormat
  LET ls_EXPORT_TABLE = "DZEN_T"
  
  LET ls_FileName = ls_TableName,"+TABLE+",
                    lo_PARAMETERS.DZLM012,"+",
                    (lo_PARAMETERS.DZLM015 USING ls_UsingFormat),".",ls_EXPORT_TABLE,".unl"
  
  CALL sadzi140_exim_get_data_from_file(ls_FileName,cs_warning_tag) RETURNING lo_DATAS,lb_success
  IF lb_success THEN
    LET li_RecCount = p_json_obj.Get(ls_EXPORT_TABLE)
    IF lo_DATAS.getLength() <> li_RecCount THEN
      LET lb_success = FALSE
      DISPLAY cs_error_tag,"Record and notice counts differ !!"
    END IF 
    IF lb_success THEN
      CALL sadzi140_exim_set_DZHN_T_data(ls_GUID,lo_DATAS) RETURNING lb_success
    END IF  
  ELSE  
    #為了相容舊版匯出檔, 開檔不成功一律設為 TRUE
    LET lb_success = TRUE
  END IF  
  
  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_set_DZHN_T_data(p_GUID,p_DATAS)
DEFINE
  p_GUID  STRING, 
  p_DATAS DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  ls_GUID         STRING, 
  lo_DATAS        DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH,
  ls_InsertValue  STRING,
  ls_InsertSQL    STRING,
  ls_UpdateValue  STRING,
  ls_UpdateSQL    STRING,
  ls_user         VARCHAR(100),
  ldt_datetime    DATETIME YEAR TO SECOND,
  lb_success      BOOLEAN,
  li_loop         INTEGER
  
  LET ls_GUID    = p_GUID
  LET lo_DATAS.* = p_DATAS.*
  LET lb_success = TRUE

  &ifndef DEBUG
  LET ls_user = g_user
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  &endif
  LET ldt_datetime = CURRENT YEAR TO SECOND
  
  FOR li_loop = 1 TO lo_DATAS.getLength()
    IF lo_DATAS[li_loop,1] IS NOT NULL AND lo_DATAS[li_loop,2] IS NOT NULL THEN
    
      LET ls_InsertValue = "'",ls_GUID,"',",
                           "'",lo_DATAS[li_loop,1],"',",
                           "'",lo_DATAS[li_loop,2],"',",
                           "'",lo_DATAS[li_loop,3],"',",
                           "'",lo_DATAS[li_loop,4],"',",
                           "'",lo_DATAS[li_loop,5],"',",
                           "'",lo_DATAS[li_loop,6],"',",
                           "'",lo_DATAS[li_loop,7],"',",
                           "'",lo_DATAS[li_loop,8],"',",
                           "'",lo_DATAS[li_loop,9],"',",
                           "'",lo_DATAS[li_loop,10],"',",
                           "'",lo_DATAS[li_loop,11],"',",
                           "'",lo_DATAS[li_loop,12],"',",
                           "'",lo_DATAS[li_loop,13],"',",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,",",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,""
                                                    
      LET ls_InsertSQL = "insert into dzhn_t (                                         ",
                         "                    dzhn000,                                 ",
                         "                    dzhn001,dzhn002,dzhn003,dzhn004,dzhn005, ",
                         "                    dzhn006,dzhn007,dzhn008,dzhn009,dzhn010, ",
                         "                    dzhn011,dzhn012,dzhn013,                 ",
                         "                    dzhncrtid,dzhncrtdt,dzhnmodid,dzhnmoddt  ",
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,                              
                         "        )                                                    "
                         
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      IF NOT lb_success THEN
        DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
      END IF
    END IF    
  END FOR 

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_delete_expired_verify_data()
CONSTANT 
  ci_reserve_day INTEGER = 3
DEFINE
  ls_delete_sql STRING, 
  lb_result     BOOLEAN 
DEFINE
  lb_return  BOOLEAN 
  
  BEGIN WORK 
  
  LET ls_delete_sql = "delete from dzhal_t                                                        ",
                      " where exists (                                                            ",
                      "                select 1                                                   ",
                      "                  from dzha_t ha                                           ",
                      "                 where ha.dzha000 = dzhal000                               ",
                      "                   and ha.dzhacrtdt < ",cs_get_date," - ",ci_reserve_day," ",
                      "              )                                                            "

  CALL sadzi140_exim_exec_SQL(ls_delete_sql) RETURNING lb_result
  IF NOT lb_result THEN GOTO _RETURN END IF
              
  LET ls_delete_sql = "delete from dzhb_t                                                         ",
                      " where exists (                                                            ",
                      "                select 1                                                   ",
                      "                  from dzha_t ha                                           ",
                      "                 where ha.dzha000 = dzhb000                                ",
                      "                   and ha.dzhacrtdt < ",cs_get_date," - ",ci_reserve_day," ",
                      "              )                                                            " 
              
  CALL sadzi140_exim_exec_SQL(ls_delete_sql) RETURNING lb_result
  IF NOT lb_result THEN GOTO _RETURN END IF
              
  LET ls_delete_sql = "delete from dzhbl_t                                                        ",
                      " where exists (                                                            ",
                      "                select 1                                                   ", 
                      "                  from dzha_t ha                                           ",
                      "                 where ha.dzha000 = dzhbl000                               ",
                      "                   and ha.dzhacrtdt < ",cs_get_date," - ",ci_reserve_day," ",
                      "              )                                                            " 

  CALL sadzi140_exim_exec_SQL(ls_delete_sql) RETURNING lb_result
  IF NOT lb_result THEN GOTO _RETURN END IF
              
  LET ls_delete_sql = "delete from dzhc_t                                                         ",
                      " where exists (                                                            ",
                      "                select 1                                                   ",
                      "                  from dzha_t ha                                           ",
                      "                 where ha.dzha000 = dzhc000                                ",
                      "                   and ha.dzhacrtdt < ",cs_get_date," - ",ci_reserve_day," ",
                      "              )                                                            " 

  CALL sadzi140_exim_exec_SQL(ls_delete_sql) RETURNING lb_result
  IF NOT lb_result THEN GOTO _RETURN END IF
              
  LET ls_delete_sql = "delete from dzhd_t                                                         ",
                      " where exists (                                                            ",
                      "                select 1                                                   ",
                      "                  from dzha_t ha                                           ",
                      "                 where ha.dzha000 = dzhd000                                ",
                      "                   and ha.dzhacrtdt < ",cs_get_date," - ",ci_reserve_day," ",
                      "              )                                                            " 

  CALL sadzi140_exim_exec_SQL(ls_delete_sql) RETURNING lb_result
  IF NOT lb_result THEN GOTO _RETURN END IF
              
  LET ls_delete_sql = "delete from dzhn_t                                                         ",
                      " where exists (                                                            ",
                      "                select 1                                                   ",
                      "                  from dzha_t ha                                           ",
                      "                 where ha.dzha000 = dzhn000                                ",
                      "                   and ha.dzhacrtdt < ",cs_get_date," - ",ci_reserve_day," ",
                      "              )                                                            " 

  CALL sadzi140_exim_exec_SQL(ls_delete_sql) RETURNING lb_result
  IF NOT lb_result THEN GOTO _RETURN END IF
              
  LET ls_delete_sql = "delete from dzha_t                                      ",
                      " where dzhacrtdt < ",cs_get_date," - ",ci_reserve_day," "

  CALL sadzi140_exim_exec_SQL(ls_delete_sql) RETURNING lb_result
  IF NOT lb_result THEN GOTO _RETURN END IF

  LABEL _RETURN:
  
  IF NOT lb_result THEN
    TRY 
      ROLLBACK WORK
    CATCH
    END TRY  
  ELSE
    TRY 
      COMMIT WORK
    CATCH
    END TRY  
  END IF 
  
  LET lb_return = lb_result
  
  RETURN lb_return
  
END FUNCTION
#20160519 end

#20160905 begin
## Fetch dzea_t to dzha_t begin
FUNCTION sadzi140_exim_fetch_dzea_t_to_dzha_t(p_GUID,p_table_name)
DEFINE
  p_GUID         STRING,
  p_table_name   STRING
DEFINE
  ls_GUID       STRING,
  ls_table_name STRING,
  lo_dzea_t     DYNAMIC ARRAY OF T_DZEA_ORIG_DATA,
  lb_result     BOOLEAN
DEFINE
  lb_return  BOOLEAN
  
  LET ls_GUID       = p_GUID
  LET ls_table_name = p_table_name

  CALL sadzi140_exim_get_dzea_t_raw_data(ls_table_name) RETURNING lo_dzea_t
  CALL sadzi140_exim_set_dzea_t_raw_data_to_dzha_t(ls_GUID,ls_table_name,lo_dzea_t) RETURNING lb_result

  LET lb_return = lb_result 
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_exim_get_dzea_t_raw_data(p_table_name)
DEFINE
  p_table_name   STRING
DEFINE
  ls_table_name STRING,
  lo_dzea_t     DYNAMIC ARRAY OF T_DZEA_ORIG_DATA,
  ls_SQL        STRING,
  li_rec_cnt    INTEGER
DEFINE
  lo_return  DYNAMIC ARRAY OF T_DZEA_ORIG_DATA

  LET ls_table_name = p_table_name

  LET ls_sql = "select ea.dzeastus,                                                      ",
               "       ea.dzea001,  ea.dzea002,  ea.dzea003,  ea.dzea004,  ea.dzea005,   ",
               "       ea.dzea006,  ea.dzea007,  ea.dzea008,  ea.dzea009,  ea.dzea010,   ",
               "       ea.dzeaownid,ea.dzeaowndp,ea.dzeacrtid,ea.dzeacrtdp,ea.dzeacrtdt, ",
               "       ea.dzeamodid,ea.dzeamoddt,ea.dzeacnfid,ea.dzeacnfdt,              ",
               "       ea.dzea011,  ea.dzea012,  ea.dzea013,  ea.dzea014,  ea.dzea015,   ",
               "       ea.dzea016,  ea.dzea017,  ea.dzea018,  ea.dzea019,  ea.dzea020,   ",
               "       ea.dzea021                                                        ",
               "  from dzea_t ea                                                         ",
               " where ea.dzea001 = '",ls_table_name,"'                                  ",
               " order by ea.dzea001                                                     "
                
  PREPARE lpre_get_dzea_t_raw_data FROM ls_sql
  DECLARE lcur_get_dzea_t_raw_data CURSOR FOR lpre_get_dzea_t_raw_data

  LET li_rec_cnt = 1
  
  OPEN lcur_get_dzea_t_raw_data
  FOREACH lcur_get_dzea_t_raw_data INTO lo_dzea_t[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_dzea_t_raw_data
  
  FREE lpre_get_dzea_t_raw_data
  FREE lcur_get_dzea_t_raw_data 

  CALL lo_dzea_t.deleteElement(li_rec_cnt)

  LET lo_return = lo_dzea_t

  RETURN lo_return
  
END FUNCTION

FUNCTION sadzi140_exim_set_dzea_t_raw_data_to_dzha_t(p_GUID,p_table_name,p_dzea_t)
DEFINE
  p_GUID       STRING,
  p_table_name STRING,
  p_dzea_t     DYNAMIC ARRAY OF T_DZEA_ORIG_DATA
DEFINE
  ls_GUID        STRING,
  ls_table_name  STRING,
  lo_dzea_t      DYNAMIC ARRAY OF T_DZEA_ORIG_DATA,
  li_loop        INTEGER,
  ls_InsertSQL   STRING,
  ls_UpdateSQL   STRING,
  ls_InsertValue STRING,
  ls_UpdateValue STRING,
  ldt_datetime   DATETIME YEAR TO SECOND,
  ls_user        STRING,
  lb_success     BOOLEAN
DEFINE
  lb_return BOOLEAN  

  LET ls_GUID       = p_GUID
  LET ls_table_name = p_table_name
  LET lo_dzea_t     = p_dzea_t

  &ifndef DEBUG
  LET ls_user = g_user
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  &endif
  LET ldt_datetime = CURRENT YEAR TO SECOND
  
  FOR li_loop = 1 TO lo_dzea_t.getLength()
      LET ls_InsertValue = "'",ls_GUID,"',",
                           "'",lo_dzea_t[li_loop].dzea001,"',",
                           "'",lo_dzea_t[li_loop].dzea002,"',",
                           "'",lo_dzea_t[li_loop].dzea003,"',",
                           "'",lo_dzea_t[li_loop].dzea004,"',",
                           "'",lo_dzea_t[li_loop].dzea005,"',",
                           "'",lo_dzea_t[li_loop].dzea006,"',",
                           "'",lo_dzea_t[li_loop].dzea007,"',",
                           "'",lo_dzea_t[li_loop].dzea008,"',",
                           "'",lo_dzea_t[li_loop].dzea009,"',",
                           "'",lo_dzea_t[li_loop].dzea010,"',",
                           "'",lo_dzea_t[li_loop].dzea011,"',",
                           "'",lo_dzea_t[li_loop].dzea012,"',",
                           "'",lo_dzea_t[li_loop].dzea013,"',",
                           "'",lo_dzea_t[li_loop].dzea014,"',",
                           "'",lo_dzea_t[li_loop].dzea015,"',",
                           "'",lo_dzea_t[li_loop].dzea016,"',",
                           "'",lo_dzea_t[li_loop].dzea017,"',",
                           "'",lo_dzea_t[li_loop].dzea018,"',",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,",",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,""

      LET ls_InsertSQL = "insert into dzha_t (                                         ",
                         "                    dzha000,                                 ", 
                         "                    dzha001,dzha002,dzha003,dzha004,dzha005, ",
                         "                    dzha006,dzha007,dzha008,dzha009,dzha010, ",
                         "                    dzha011,dzha012,dzha013,dzha014,dzha015, ",
                         "                    dzha016,dzha017,dzha018,                 ",
                         "                    dzhacrtid,dzhacrtdt,dzhamodid,dzhamoddt  ", 
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "

      LET ls_UpdateValue = "dzha002 = '",lo_dzea_t[li_loop].dzea002,"',",
                           "dzha003 = '",lo_dzea_t[li_loop].dzea003,"',",
                           "dzha004 = '",lo_dzea_t[li_loop].dzea004,"',",
                           "dzha005 = '",lo_dzea_t[li_loop].dzea005,"',",
                           "dzha006 = '",lo_dzea_t[li_loop].dzea006,"',",
                           "dzha007 = '",lo_dzea_t[li_loop].dzea007,"',",
                           "dzha008 = '",lo_dzea_t[li_loop].dzea008,"',",
                           "dzha009 = '",lo_dzea_t[li_loop].dzea009,"',",
                           "dzha010 = '",lo_dzea_t[li_loop].dzea010,"',",
                           "dzha011 = '",lo_dzea_t[li_loop].dzea011,"',",
                           "dzha012 = '",lo_dzea_t[li_loop].dzea012,"',",
                           "dzha013 = '",lo_dzea_t[li_loop].dzea013,"',",
                           "dzha014 = '",lo_dzea_t[li_loop].dzea014,"',",
                           "dzha015 = '",lo_dzea_t[li_loop].dzea015,"',",
                           "dzha016 = '",lo_dzea_t[li_loop].dzea016,"',",
                           "dzha017 = '",lo_dzea_t[li_loop].dzea017,"',",
                           "dzha018 = '",lo_dzea_t[li_loop].dzea018,"'"
                           
      LET ls_UpdateSQL = "update dzha_t set ",
                         ls_UpdateValue,
                         " where dzha000 = '",ls_GUID,"'",
                         "   and dzha001 = '",lo_dzea_t[li_loop].dzea001,"'"
                         
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
                         
  END FOR 

  LET lb_return = lb_success
  
  RETURN lb_return
  
END FUNCTION
## Fetch dzea_t to dzha_t end

## Fetch dzeb_t to dzhb_t begin
FUNCTION sadzi140_exim_fetch_dzeb_t_to_dzhb_t(p_GUID,p_table_name)
DEFINE
  p_GUID         STRING,
  p_table_name   STRING
DEFINE
  ls_GUID       STRING,
  ls_table_name STRING,
  lo_dzeb_t     DYNAMIC ARRAY OF T_DZEB_ORIG_DATA,
  lb_result     BOOLEAN
DEFINE
  lb_return  BOOLEAN
  
  LET ls_GUID       = p_GUID
  LET ls_table_name = p_table_name

  CALL sadzi140_exim_get_dzeb_t_raw_data(ls_table_name) RETURNING lo_dzeb_t
  CALL sadzi140_exim_set_dzeb_t_raw_data_to_dzhb_t(ls_GUID,ls_table_name,lo_dzeb_t) RETURNING lb_result

  LET lb_return = lb_result 
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_exim_get_dzeb_t_raw_data(p_table_name)
DEFINE
  p_table_name   STRING
DEFINE
  ls_table_name STRING,
  lo_dzeb_t     DYNAMIC ARRAY OF T_DZEB_ORIG_DATA,
  ls_SQL        STRING,
  li_rec_cnt    INTEGER
DEFINE
  lo_return  DYNAMIC ARRAY OF T_DZEB_ORIG_DATA

  LET ls_table_name = p_table_name

  LET ls_sql = "select eb.dzebstus,                                                      ",
               "       eb.dzeb001,  eb.dzeb002,  eb.dzeb003,  eb.dzeb004,  eb.dzeb005,   ",
               "       eb.dzeb006,  eb.dzeb007,  eb.dzeb008,  eb.dzeb021,  eb.dzeb022,   ",
               "       eb.dzeb023,  eb.dzeb024,  eb.dzeb027,                             ",
               "       eb.dzebownid,eb.dzebowndp,eb.dzebcrtid,eb.dzebcrtdp,eb.dzebcrtdt, ",
               "       eb.dzebmodid,eb.dzebmoddt,eb.dzebcnfid,eb.dzebcnfdt,              ",
               "       eb.dzeb028,  eb.dzeb029,  eb.dzeb012,  eb.dzeb030,  eb.dzeb031    ",
               "  from dzeb_t eb                                                         ",
               " where eb.dzeb001 = '",ls_table_name,"'                                  ",
               " order by eb.dzeb021                                                     "
                
  PREPARE lpre_get_dzeb_t_raw_data FROM ls_sql
  DECLARE lcur_get_dzeb_t_raw_data CURSOR FOR lpre_get_dzeb_t_raw_data

  LET li_rec_cnt = 1
  
  OPEN lcur_get_dzeb_t_raw_data
  FOREACH lcur_get_dzeb_t_raw_data INTO lo_dzeb_t[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_dzeb_t_raw_data
  
  FREE lpre_get_dzeb_t_raw_data
  FREE lcur_get_dzeb_t_raw_data 

  CALL lo_dzeb_t.deleteElement(li_rec_cnt)

  LET lo_return = lo_dzeb_t

  RETURN lo_return
  
END FUNCTION

FUNCTION sadzi140_exim_set_dzeb_t_raw_data_to_dzhb_t(p_GUID,p_table_name,p_dzeb_t)
DEFINE
  p_GUID       STRING,
  p_table_name STRING,
  p_dzeb_t     DYNAMIC ARRAY OF T_DZEB_ORIG_DATA
DEFINE
  ls_GUID        STRING,
  ls_table_name  STRING,
  lo_dzeb_t      DYNAMIC ARRAY OF T_DZEB_ORIG_DATA,
  li_loop        INTEGER,
  ls_InsertSQL   STRING,
  ls_UpdateSQL   STRING,
  ls_InsertValue STRING,
  ls_UpdateValue STRING,
  ldt_datetime   DATETIME YEAR TO SECOND,
  ls_user        STRING,
  lb_success     BOOLEAN
DEFINE
  lb_return BOOLEAN  

  LET ls_GUID       = p_GUID
  LET ls_table_name = p_table_name
  LET lo_dzeb_t     = p_dzeb_t

  &ifndef DEBUG
  LET ls_user = g_user
  &else
  LET ls_user = FGL_GETENV("USERNUMBER")
  &endif
  LET ldt_datetime = CURRENT YEAR TO SECOND
  
  FOR li_loop = 1 TO lo_dzeb_t.getLength()
      LET ls_InsertValue = "'",ls_GUID,"',",
                           "'",lo_dzeb_t[li_loop].dzeb001,"',",
                           "'",lo_dzeb_t[li_loop].dzeb002,"',",
                           "'",lo_dzeb_t[li_loop].dzeb003,"',",
                           "'",lo_dzeb_t[li_loop].dzeb004,"',",
                           "'",lo_dzeb_t[li_loop].dzeb005,"',",
                           "'",lo_dzeb_t[li_loop].dzeb006,"',",
                           "'",lo_dzeb_t[li_loop].dzeb007,"',",
                           "'",lo_dzeb_t[li_loop].dzeb008,"',",
                           "'",lo_dzeb_t[li_loop].dzeb012,"',",
                           "",NVL(lo_dzeb_t[li_loop].dzeb021,"NULL"),",",
                           "'",lo_dzeb_t[li_loop].dzeb022,"',",
                           "",NVL(lo_dzeb_t[li_loop].dzeb023,"NULL"),",",
                           "'",lo_dzeb_t[li_loop].dzeb024,"',",
                           "'",lo_dzeb_t[li_loop].dzeb027,"',",
                           "'",lo_dzeb_t[li_loop].dzeb028,"',",
                           "'",lo_dzeb_t[li_loop].dzeb029,"',",
                           "'",lo_dzeb_t[li_loop].dzeb030,"',",
                           "'",lo_dzeb_t[li_loop].dzeb031,"',",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,",",
                           "'",ls_user,"',",
                           "",cs_get_time_stamp,""

      LET ls_InsertSQL = "insert into dzhb_t (                                         ",
                         "                    dzhb000,                                 ", 
                         "                    dzhb001,dzhb002,dzhb003,dzhb004,dzhb005, ",
                         "                    dzhb006,dzhb007,dzhb008,dzhb012,dzhb021, ",
                         "                    dzhb022,dzhb023,dzhb024,dzhb027,dzhb028, ",
                         "                    dzhb029,dzhb030,dzhb031,                 ",
                         "                    dzhbcrtid,dzhbcrtdt,dzhbmodid,dzhbmoddt  ",
                         "                   )                                         ",  
                         " values (                                                    ",
                         ls_InsertValue,
                         "        )                                                    "

      LET ls_UpdateValue = "dzhb002 = '",lo_dzeb_t[li_loop].dzeb002,"',",
                           "dzhb003 = '",lo_dzeb_t[li_loop].dzeb003,"',",
                           "dzhb004 = '",lo_dzeb_t[li_loop].dzeb004,"',",
                           "dzhb005 = '",lo_dzeb_t[li_loop].dzeb005,"',",
                           "dzhb006 = '",lo_dzeb_t[li_loop].dzeb006,"',",
                           "dzhb007 = '",lo_dzeb_t[li_loop].dzeb007,"',",
                           "dzhb008 = '",lo_dzeb_t[li_loop].dzeb008,"',",
                           "dzhb012 = '",lo_dzeb_t[li_loop].dzeb012,"',",
                           "dzhb021 = ",NVL(lo_dzeb_t[li_loop].dzeb021,"NULL"),",",
                           "dzhb022 = '",lo_dzeb_t[li_loop].dzeb022,"',",
                           "dzhb023 = ",NVL(lo_dzeb_t[li_loop].dzeb023,"NULL"),",",
                           "dzhb024 = '",lo_dzeb_t[li_loop].dzeb024,"',",
                           "dzhb027 = '",lo_dzeb_t[li_loop].dzeb027,"',",
                           "dzhb028 = '",lo_dzeb_t[li_loop].dzeb028,"',",
                           "dzhb029 = '",lo_dzeb_t[li_loop].dzeb029,"',",
                           "dzhb030 = '",lo_dzeb_t[li_loop].dzeb030,"',",
                           "dzhb031 = '",lo_dzeb_t[li_loop].dzeb031,"'"
                           
      LET ls_UpdateSQL = "update dzhb_t set ",
                         ls_UpdateValue,
                         " where dzhb000 = '",ls_GUID,"'",
                         "   and dzhb001 = '",lo_dzeb_t[li_loop].dzeb001,"'"
                         
      CALL sadzi140_exim_exec_SQL(ls_InsertSQL) RETURNING lb_success
      IF NOT lb_success THEN
        CALL sadzi140_exim_exec_SQL(ls_UpdateSQL) RETURNING lb_success
        IF NOT lb_success THEN
          DISPLAY "[ INSERT SQL ] ",ls_InsertSQL,"\n" 
          DISPLAY "[ Update SQL ] ",ls_UpdateSQL,"\n" 
          EXIT FOR 
        END IF
      END IF
                         
  END FOR 

  LET lb_return = lb_success
  
  RETURN lb_return
  
END FUNCTION
## Fetch dzeb_t to dzhb_t end

#20160905 end

################################################################################

FUNCTION sadzi140_exim_delete_invalid_dzeu_data(p_TableName)
DEFINE
  p_TableName  STRING
DEFINE
  ls_TableName        STRING,
  ls_DeleteSQL        STRING, 
  lb_success          BOOLEAN 

  LET ls_TableName = p_TableName.toLowerCase()

  LET ls_DeleteSQL = "delete from dzeu_t                                ",
                     " where not exists (                               ",
                     "                    select 1                      ",
                     "                      from gzda_t da              ",
                     "                     where da.gzda001 = dzeu002   ",
                     "                  )                               ",
                     "   and dzeu001 = '",ls_TableName.toLowerCase(),"' "
  
  CALL sadzi140_exim_exec_SQL(ls_DeleteSQL) RETURNING lb_success

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_add_missing_dzeu_data(p_TableName)
DEFINE
  p_TableName  STRING
DEFINE
  ls_TableName   STRING,
  ls_DeleteSQL   STRING, 
  lo_schema_list DYNAMIC ARRAY OF T_SCHEMA_LIST, 
  lo_table_type  T_TABLE_TYPE, 
  li_loop        INTEGER,
  lb_success     BOOLEAN 

  LET ls_TableName = p_TableName.toLowerCase()

  LET lb_success = TRUE

  CALL sadzi140_db_get_valid_schema_list() RETURNING lo_schema_list
  FOR li_loop = 1 TO lo_schema_list.getLength()
    CALL sadzi140_db_get_table_type_by_dsdemo(ls_TableName,lo_schema_list[li_loop].sl_SCHEMA_NAME) RETURNING lo_table_type.*
    IF lo_table_type.DZEU001 IS NOT NULL THEN  
      CALL sadzi140_crud_insert_update_dzeu_t(lo_table_type.*)
    END IF  
  END FOR 

  RETURN lb_success
  
END FUNCTION

FUNCTION sadzi140_exim_check_version_data_exist(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_table_name  STRING,
  ls_sql         STRING,
  li_rec_count   INTEGER,
  lb_return      BOOLEAN

  LET ls_table_name = p_table_name.toUpperCase()

  #取得資料筆數
  LET ls_sql = "select count(*)                        ",
               "  from dzev_t ev                       ",
               " where ev.dzev002 = '",ls_table_name,"'",
               "   and ev.dzev004 = 'TableDesigner'    "
   
  PREPARE lpre_check_version_data_exist FROM ls_sql
  DECLARE lcur_check_version_data_exist CURSOR FOR lpre_check_version_data_exist
  OPEN lcur_check_version_data_exist
  FETCH lcur_check_version_data_exist INTO li_rec_count
  CLOSE lcur_check_version_data_exist
  FREE lcur_check_version_data_exist
  FREE lpre_check_version_data_exist

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_exim_check_alm_request_data_exist(p_table_name,p_parameters)
DEFINE
  p_table_name  STRING,
  p_parameters  T_DZLM_T_SCR  
DEFINE
  ls_table_name  STRING,
  lo_parameters  T_DZLM_T_SCR,  
  ls_sql         STRING,
  li_rec_count   INTEGER,
  lb_return      BOOLEAN

  LET ls_table_name   = p_table_name.toUpperCase()
  LET lo_parameters.* = p_parameters.*

  #取得資料筆數
  LET ls_SQL = "Select count(*)                                       ",                         
               "  from dzev_t ev                                      ",
               " where ev.dzev002 = '",ls_table_name.toUpperCase(),"' ",
               "   and ev.dzev003 = trim('",lo_parameters.DZLM015,"') ", 
               "   and ev.dzev004 = 'TableDesigner'                   ",
               "   and ev.dzev018 = '",lo_parameters.DZLM005,"'       ",
               "   and ev.dzev019 = '",lo_parameters.DZLM012,"'       "
   
  PREPARE lpre_check_alm_request_data_exist FROM ls_sql
  DECLARE lcur_check_alm_request_data_exist CURSOR FOR lpre_check_alm_request_data_exist
  OPEN lcur_check_alm_request_data_exist
  FETCH lcur_check_alm_request_data_exist INTO li_rec_count
  CLOSE lcur_check_alm_request_data_exist
  FREE lcur_check_alm_request_data_exist
  FREE lpre_check_alm_request_data_exist

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_exim_get_table_ship_notice(p_SQL)
DEFINE
  p_SQL STRING
DEFINE  
  lo_table_ship_notice T_TABLE_SHIP_NOTICE,
  ls_SQL     STRING,
  li_RecCnt  INTEGER
DEFINE
  lo_RETURN T_TABLE_SHIP_NOTICE

  LET ls_SQL = p_SQL

  INITIALIZE lo_table_ship_notice TO NULL
  INITIALIZE lo_RETURN TO NULL
  
  PREPARE lpre_get_table_ship_notice FROM ls_SQL
  DECLARE lcur_get_table_ship_notice CURSOR FOR lpre_get_table_ship_notice

  LET li_RecCnt = 1 
  
  OPEN lcur_get_table_ship_notice
  FETCH lcur_get_table_ship_notice INTO lo_table_ship_notice.*
  CLOSE lcur_get_table_ship_notice

  FREE lpre_get_table_ship_notice
  FREE lcur_get_table_ship_notice  

  LET lo_RETURN.* = lo_table_ship_notice.*
  
  RETURN lo_RETURN.*
  
END FUNCTION

FUNCTION sadzi140_exim_set_table_export_notice(p_TableName,p_data)
DEFINE
  p_TableName STRING,
  p_data      STRING
DEFINE
  ls_data       STRING,
  ls_TableName  STRING,
  ls_FileName   STRING,
  lb_Success    BOOLEAN

  LET ls_data      = p_data
  LET ls_TableName = p_TableName

  LET ls_FileName = ls_TableName,"+EXPORT+NOTICE.txn"
  CALL sadzi140_exim_write_file(ls_FileName,ls_data) RETURNING lb_Success

  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_get_export_pkg_sys_info(p_TableName)
DEFINE
  p_TableName STRING
DEFINE 
  ls_FileName     STRING,
  lb_Success      BOOLEAN,
  lo_json_obj     util.JSONObject,
  lo_pkg_sys_info T_PKG_SYS_INFO  
DEFINE
  lo_return T_PKG_SYS_INFO  

  LET ls_FileName = p_TableName

  CALL sadzi140_exim_get_table_export_notice(ls_FileName) RETURNING lb_Success,lo_json_obj
  CALL sadzi140_exim_get_pkg_sys_info(lo_json_obj) RETURNING lo_pkg_sys_info.*

  LET lo_return.* = lo_pkg_sys_info.*

  RETURN lo_json_obj,lo_return.*
  
END FUNCTION 

FUNCTION sadzi140_exim_get_table_export_notice(p_TableName)
DEFINE
  p_TableName STRING
DEFINE
  ls_TableName  STRING,
  ls_FileName   STRING,
  lb_Success    BOOLEAN,
  lo_json_obj   util.JSONObject

  LET ls_TableName = p_TableName

  LET ls_FileName = ls_TableName,"+EXPORT+NOTICE.txn"
  CALL sadzi140_exim_get_json_objects(ls_FileName) RETURNING lb_Success,lo_json_obj

  RETURN lb_Success,lo_json_obj

END FUNCTION

FUNCTION sadzi140_exim_get_pkg_sys_info(p_json_obj)
DEFINE
  p_json_obj util.JSONObject
DEFINE 
  lo_pkg_sys_info T_PKG_SYS_INFO  
DEFINE
  lo_return T_PKG_SYS_INFO

  LET lo_pkg_sys_info.psi_DGENV          = p_json_obj.get(cs_keyword_DGENV)  
  LET lo_pkg_sys_info.psi_ERPID          = NVL(p_json_obj.Get(cs_keyword_ERPID),cs_ERP_default_ID)
  LET lo_pkg_sys_info.psi_CUST           = p_json_obj.get(cs_keyword_CUST)  
  LET lo_pkg_sys_info.psi_DB_TYPE        = UPSHIFT(p_json_obj.get(cs_keyword_DB_TYPE)) #20200330  
  LET lo_pkg_sys_info.psi_TOPIND         = NVL(p_json_obj.get(cs_keyword_TOPIND),cs_default_topind)  
  LET lo_pkg_sys_info.psi_RENAMED        = p_json_obj.get(cs_keyword_RENAMED)  
  LET lo_pkg_sys_info.psi_RT_DEBUG_MODE  = "N"
  LET lo_pkg_sys_info.psi_ERPVER         = NVL(p_json_obj.get(cs_keyword_ERPVER),cs_ERP_default_ERPVER) #20180129
  LET lo_pkg_sys_info.psi_CONSTRUCT_VER  = NVL(p_json_obj.get(cs_keyword_CONSTRUCT_VER),cs_init_version_code) #20180410
  LET lo_pkg_sys_info.psi_RT_PATCH_MODE  = NULL #20180307
  LET lo_pkg_sys_info.psi_T100PASSERPVER = NULL #20180420
  
  LET lo_return.* = lo_pkg_sys_info.*

  RETURN lo_return.*    
  
END FUNCTION

FUNCTION sadzi140_exim_get_sys_info()
DEFINE 
  lo_local_sys_info T_PKG_SYS_INFO  
DEFINE
  lo_return T_PKG_SYS_INFO

  LET lo_local_sys_info.psi_DGENV          = FGL_GETENV(cs_keyword_DGENV)
  LET lo_local_sys_info.psi_ERPID          = NVL(FGL_GETENV(cs_keyword_ERPID),cs_ERP_default_ID)
  LET lo_local_sys_info.psi_CUST           = FGL_GETENV(cs_keyword_CUST) 
  LET lo_local_sys_info.psi_DB_TYPE        = UPSHIFT(sadzi140_util_get_db_type()) #20200330 #20240521
  LET lo_local_sys_info.psi_TOPIND         = NVL(NVL(FGL_GETENV("INDSD"),FGL_GETENV("TOPIND")),cs_default_topind) --20230825(230825-00006)
  LET lo_local_sys_info.psi_RENAMED        = "N"
  LET lo_local_sys_info.psi_RT_DEBUG_MODE  = NVL(FGL_GETENV(cs_env_RT_DEBUG_MODE),"N")
  LET lo_local_sys_info.psi_ERPVER         = NVL(FGL_GETENV("ERPVER"),cs_ERP_default_ERPVER) #20180129
  LET lo_local_sys_info.psi_CONSTRUCT_VER  = NULL --cs_init_version_code #20180410
  LET lo_local_sys_info.psi_RT_PATCH_MODE  = NVL(FGL_GETENV(cs_env_RT_PATCH_MODE),"N") #20180307
  LET lo_local_sys_info.psi_T100PASSERPVER = NVL(FGL_GETENV(cs_env_T100PASSERPVER),"N") #20180420

  LET lo_return.* = lo_local_sys_info.*

  RETURN lo_return.*    
  
END FUNCTION
  
FUNCTION sadzi140_exim_table_check_notice(p_TableName,p_json_obj,p_pkg_sys_info,p_src_pkg_sys_info)
DEFINE
  p_TableName        STRING,
  p_json_obj         util.JSONObject, 
  p_pkg_sys_info     T_PKG_SYS_INFO,
  p_src_pkg_sys_info T_PKG_SYS_INFO
DEFINE
  ls_TableName        STRING,
  lo_json_obj         util.JSONObject,
  lo_pkg_sys_info     T_PKG_SYS_INFO,
  lo_src_pkg_sys_info T_PKG_SYS_INFO,
  ls_FileName         STRING,
  ls_err_message      STRING, --20190108
  ls_error_code       STRING, --20190108
  lb_Success          BOOLEAN,
  li_dzea_count       INTEGER, 
  li_dzeb_count       INTEGER, 
  lb_result           BOOLEAN

  LET ls_TableName          = p_TableName
  LET lo_json_obj           = p_json_obj
  LET lo_pkg_sys_info.*     = p_pkg_sys_info.*
  LET lo_src_pkg_sys_info.* = p_src_pkg_sys_info.* 

  LET lb_success = TRUE

  #DZEA_T 或 DZEB_T 數量為 0 , 則不處理(失敗)  
  LET li_dzea_count = lo_json_obj.Get("DZEA_T")
  LET li_dzeb_count = lo_json_obj.Get("DZEB_T")

  IF (li_dzea_count = 0) OR (li_dzeb_count = 0) THEN
    DISPLAY cs_error_tag,"DZEA_T or DZEB_T record counts equal to 0."
    LET lb_success = FALSE
  END IF 

  #確認來源與目的端的 ERPID 是否一致
  IF mb_check_erpid THEN 
    IF (lo_src_pkg_sys_info.psi_ERPID <> lo_pkg_sys_info.psi_ERPID) AND (lo_pkg_sys_info.psi_RT_DEBUG_MODE <> "Y") THEN 
      DISPLAY cs_error_tag,"Source ERPID : '",lo_src_pkg_sys_info.psi_ERPID,"' is not equal to destination ERPID : '",lo_pkg_sys_info.psi_ERPID,"'."
      LET lb_success = FALSE
    END IF
  END IF  

  #20180129 begin
  IF (lo_src_pkg_sys_info.psi_ERPVER <> lo_pkg_sys_info.psi_ERPVER) THEN
    WHILE TRUE
      IF (NVL(lo_pkg_sys_info.psi_RT_DEBUG_MODE,"N") = "Y") OR 
         (NVL(lo_pkg_sys_info.psi_RT_PATCH_MODE,"N") = "Y") OR 
         (NVL(lo_pkg_sys_info.psi_T100PASSERPVER,"N") = "Y") THEN #20190108
        DISPLAY cs_message_tag,"RT_DEBUG_MODE : '",lo_pkg_sys_info.psi_RT_DEBUG_MODE,"',",
                               "RT_PATCH_MODE : '",lo_pkg_sys_info.psi_RT_PATCH_MODE,"',",
                               "T100PASSERPVER : '",lo_pkg_sys_info.psi_T100PASSERPVER,"'" #20190108
        EXIT WHILE 
      END IF
      
      #20180410 begin
      IF (
           (lo_src_pkg_sys_info.psi_CONSTRUCT_VER = cs_first_version_code) AND 
           (lo_src_pkg_sys_info.psi_CONSTRUCT_VER = lo_pkg_sys_info.psi_CONSTRUCT_VER)
         OR
           (lo_src_pkg_sys_info.psi_CONSTRUCT_VER = NVL(lo_pkg_sys_info.psi_CONSTRUCT_VER,lo_src_pkg_sys_info.psi_CONSTRUCT_VER))  
         ) AND
         (lo_src_pkg_sys_info.psi_DGENV = cs_dgenv_standard) AND
         (lo_src_pkg_sys_info.psi_DGENV = lo_pkg_sys_info.psi_DGENV) AND
         (lo_pkg_sys_info.psi_CUST      = cs_env_cust_digiwin) AND
         (ms_user <> cs_topstd_user_name) THEN
        EXIT WHILE 
      ELSE  
        #非產中環境則擋
        DISPLAY cs_message_tag,"Source ERPVER : '",lo_src_pkg_sys_info.psi_ERPVER,"', ",
                               "Target ERPVER : '",lo_pkg_sys_info.psi_ERPVER,"', ",
                               "Source version : '",lo_src_pkg_sys_info.psi_CONSTRUCT_VER,"', ",
                               "Target vsesion : '",lo_pkg_sys_info.psi_CONSTRUCT_VER,"', ",
                               "Source DGENV : '",lo_src_pkg_sys_info.psi_DGENV,"', ",
                               "Target DGENV : '",lo_pkg_sys_info.psi_DGENV,"', ",
                               "CUST : ",lo_pkg_sys_info.psi_CUST,"', ",
                               "User name : '",ms_user,"'"
      END IF 
      #20180410 end

      #20190108 mark begin
      {
      #20180420 begin
      IF (lo_pkg_sys_info.psi_T100PASSERPVER = "Y") AND (lo_pkg_sys_info.psi_DGENV = cs_dgenv_standard) THEN
        EXIT WHILE 
      ELSE  
        #若為客制環境則擋
        DISPLAY cs_message_tag,"T100PASSERPVER : '",lo_pkg_sys_info.psi_T100PASSERPVER,"', ",
                               "Target DGENV : '",lo_pkg_sys_info.psi_DGENV,"'"
      END IF
      #20180420 end
      }
      #20190108 mark end

      LET lb_success = FALSE

      LET ls_error_code = "adz-01301" -- 20190108 add
      LET ls_err_message = ls_error_code,"-",sadzp000_msg_get_message(ls_error_code,ms_dlang) -- 20190108 add
      DISPLAY cs_error_tag,ls_err_message --"Source ERPVER : '",lo_src_pkg_sys_info.psi_ERPVER,"' is not equal to destination ERPVER : '",lo_pkg_sys_info.psi_ERPVER,"'." -- 20190108 mod
      
      EXIT WHILE
      
    END WHILE 
  END IF
  #20180129 end
  
  RETURN lb_Success

END FUNCTION

FUNCTION sadzi140_exim_check_industry_if_effective(p_pkg_sys_info,p_src_pkg_sys_info)
DEFINE
  p_pkg_sys_info     T_PKG_SYS_INFO,
  p_src_pkg_sys_info T_PKG_SYS_INFO
DEFINE
  lo_pkg_sys_info     T_PKG_SYS_INFO,
  lo_src_pkg_sys_info T_PKG_SYS_INFO,
  lb_result           BOOLEAN
DEFINE
  lb_return  BOOLEAN,
  ls_return  STRING
  
  LET lo_pkg_sys_info.*     = p_pkg_sys_info.*
  LET lo_src_pkg_sys_info.* = p_src_pkg_sys_info.*

  LET lb_return = TRUE
  LET ls_return = "" 

  CASE
    #Debug 模式不做任何判斷 
    WHEN NVL(lo_pkg_sys_info.psi_RT_DEBUG_MODE,"N") = "Y"
      LET lb_return = TRUE
      LET ls_return = "" 
    #客制環境判斷 
    WHEN lo_pkg_sys_info.psi_DGENV = cs_dgenv_customize
      CASE 
        #來源是標準
        WHEN lo_src_pkg_sys_info.psi_TOPIND = cs_standard_topind AND
             (lo_pkg_sys_info.psi_TOPIND <> lo_src_pkg_sys_info.psi_TOPIND) 
          LET lb_return = TRUE
          LET ls_return = "" 
        #目的是標準
        WHEN lo_pkg_sys_info.psi_TOPIND = cs_standard_topind AND
             (lo_pkg_sys_info.psi_TOPIND <> lo_src_pkg_sys_info.psi_TOPIND)
          LET lb_return = TRUE
          LET ls_return = "" 
        #來源和目的都不是標準, 而且兩者不同
        WHEN lo_src_pkg_sys_info.psi_TOPIND <> cs_standard_topind AND
             lo_pkg_sys_info.psi_TOPIND     <> cs_standard_topind AND
             (lo_pkg_sys_info.psi_TOPIND <> lo_src_pkg_sys_info.psi_TOPIND)
          LET lb_return = FALSE
          LET ls_return = "adz-00806" 
      END CASE 
    #標準環境判斷 
    WHEN lo_pkg_sys_info.psi_DGENV = cs_dgenv_standard
      CASE 
        #來源是標準
        WHEN lo_src_pkg_sys_info.psi_TOPIND = cs_standard_topind AND
             (lo_pkg_sys_info.psi_TOPIND <> lo_src_pkg_sys_info.psi_TOPIND) 
          LET lb_return = TRUE
          LET ls_return = "" 
        #目的是標準
        WHEN lo_pkg_sys_info.psi_TOPIND = cs_standard_topind AND
             (lo_pkg_sys_info.psi_TOPIND <> lo_src_pkg_sys_info.psi_TOPIND)
          LET lb_return = FALSE
          LET ls_return = "adz-00807" 
        #來源和目的都不是標準, 而且兩者不同
        WHEN lo_src_pkg_sys_info.psi_TOPIND <> cs_standard_topind AND
             lo_pkg_sys_info.psi_TOPIND     <> cs_standard_topind AND
             (lo_pkg_sys_info.psi_TOPIND <> lo_src_pkg_sys_info.psi_TOPIND)
          LET lb_return = FALSE
          LET ls_return = "adz-00806" 
      END CASE 
  END CASE

  RETURN lb_return,ls_return 
  
END FUNCTION 

FUNCTION sadzi140_exim_get_json_objects(p_FileName)
DEFINE
  p_FileName  STRING
DEFINE
  ls_FileName  STRING
DEFINE  
  lo_channel_read  base.Channel,
  ls_TextLine      STRING,
  li_RecCnt        INTEGER,
  lb_success       BOOLEAN,
  lo_json_obj      util.JSONObject
DEFINE
  lo_RETURN  util.JSONObject  

  LET ls_FileName = p_FileName

  LET lb_success = TRUE

  LET lo_channel_read = base.Channel.create()
  TRY 
    CALL lo_channel_read.openFile(ls_FileName,"r")
  CATCH
    LET lb_success = FALSE
    DISPLAY cs_error_tag,"Open file ",ls_FileName," failed !!"
  END TRY  

  IF lb_success THEN
    LET li_RecCnt = 1 
    WHILE TRUE
      IF lo_channel_read.isEof() THEN 
        EXIT WHILE
      END IF

      LET ls_TextLine = ls_TextLine,lo_channel_read.readLine()
      
      LET li_RecCnt = li_RecCnt + 1 
        
    END WHILE
  END IF
  
  CALL lo_channel_read.close()
  TRY 
    LET lo_json_obj = util.JSONObject.parse( ls_TextLine )
  CATCH
    LET lo_json_obj = NULL
    DISPLAY cs_error_tag,"JSON File ",ls_FileName," not valid."
  END TRY  

  LET lo_RETURN = lo_json_obj
  
  RETURN lb_success,lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_get_json_array(p_FileName)
DEFINE
  p_FileName  STRING
DEFINE
  ls_FileName  STRING
DEFINE  
  lo_channel_read  base.Channel,
  ls_TextLine      STRING,
  li_RecCnt        INTEGER,
  lb_success       BOOLEAN,
  lo_json_arr      util.JSONArray
DEFINE
  lo_RETURN  util.JSONArray  

  LET ls_FileName = p_FileName

  LET lb_success = TRUE

  LET lo_channel_read = base.Channel.create()
  TRY 
    CALL lo_channel_read.openFile(ls_FileName,"r")
  CATCH
    LET lb_success = FALSE
    DISPLAY cs_error_tag,"Open file ",ls_FileName," failed !!"
  END TRY  

  IF lb_success THEN
    LET li_RecCnt = 1 
    WHILE TRUE
      IF lo_channel_read.isEof() THEN 
        EXIT WHILE
      END IF

      LET ls_TextLine = ls_TextLine,lo_channel_read.readLine()
      
      LET li_RecCnt = li_RecCnt + 1 
        
    END WHILE
  END IF
  
  CALL lo_channel_read.close()
  TRY 
    LET lo_json_arr = util.JSONArray.parse( ls_TextLine )
  CATCH
    LET lo_json_arr = NULL
    DISPLAY cs_error_tag,"JSON File ",ls_FileName," not valid."
  END TRY  

  LET lo_RETURN = lo_json_arr
  
  RETURN lb_success,lo_RETURN
  
END FUNCTION

FUNCTION sadzi140_exim_get_parameter(p_param1,p_param2,p_param3,p_param4,p_param5,p_param6)
DEFINE 
  p_param1,p_param2,p_param3,p_param4,p_param5,p_param6 STRING 
DEFINE
  lo_parameter T_DZLM_T_SCR
DEFINE
  lo_return  T_DZLM_T_SCR

  LET lo_parameter.TYPES   = p_param1 #iexp, iimp, idel, iasm
  LET lo_parameter.DZLM002 = p_param2 #TableName
  LET lo_parameter.DZLM005 = p_param3 #ConstructVersion
  LET lo_parameter.DZLM006 = p_param4 #SD_Version
  LET lo_parameter.DZLM012 = p_param5 #需求單號 
  LET lo_parameter.DZLM015 = p_param6 #序號

  LET lo_return.* = lo_parameter.*

  RETURN lo_return.*  

END FUNCTION 

FUNCTION sadzi140_exim_set_module_parameters(p_param1,p_param2,p_param3,p_param4,p_param5,p_param6)
DEFINE 
  p_param1,p_param2,p_param3,p_param4,p_param5,p_param6 STRING 
DEFINE
  lo_parameter T_DZLM_T_SCR

  LET mo_PARAMETERS.TYPES   = p_param1 #iimp
  LET mo_PARAMETERS.DZLM002 = p_param2 #TableName
  LET mo_PARAMETERS.DZLM005 = IIF(p_param3==cs_rt_design_data_code,cs_rt_design_data_version,p_param3) #ConstructVersion
  LET mo_PARAMETERS.DZLM006 = p_param4 #SD_Version
  LET mo_PARAMETERS.DZLM012 = IIF(p_param3==cs_rt_design_data_code,cs_disable_alm_request_no,p_param5) #需求單號 
  LET mo_PARAMETERS.DZLM015 = IIF(p_param3==cs_rt_design_data_code,cs_disable_alm_sequence_no,p_param6) #序號
  
END FUNCTION 

FUNCTION sadzi140_exim_save_to_client(mo_parameter)
DEFINE
  mo_parameter T_PUT_GET_FILE_PARA
DEFINE
  lo_parameter    T_PUT_GET_FILE_PARA,
  ls_source       STRING,
  ls_destination  STRING,
  ls_os_separator STRING  

  LET lo_parameter.* = mo_parameter.*

  #取得作業系統的分隔符號
  CALL os.Path.separator() RETURNING ls_os_separator 

  LET ls_source      = lo_parameter.SERVER_FILE_PATH,ls_os_separator,lo_parameter.SERVER_FILE_NAME
  LET ls_destination = lo_parameter.CLIENT_FILE_PATH,ls_os_separator,lo_parameter.CLIENT_FILE_NAME

  DISPLAY "Source File : ",ls_source
  DISPLAY "Destination File : ",ls_destination 

  TRY 
    CALL FGL_PUTFILE(ls_source,ls_destination)
  CATCH
    DISPLAY cs_error_tag,"The file ",lo_parameter.SERVER_FILE_NAME," can not put to client side."
  END TRY  
  
END FUNCTION 

FUNCTION sadzi140_exim_catch_to_server(mo_parameter)
DEFINE
  mo_parameter T_PUT_GET_FILE_PARA
DEFINE
  lo_parameter    T_PUT_GET_FILE_PARA,
  ls_source       STRING,
  ls_destination  STRING,
  ls_os_separator STRING  
DEFINE
  lb_return  BOOLEAN  

  LET lo_parameter.* = mo_parameter.*

  LET lb_return = TRUE

  #取得作業系統的分隔符號
  CALL os.Path.separator() RETURNING ls_os_separator 

  LET ls_source      = lo_parameter.CLIENT_FILE_PATH,ls_os_separator,lo_parameter.CLIENT_FILE_NAME
  LET ls_destination = lo_parameter.SERVER_FILE_PATH,ls_os_separator,lo_parameter.SERVER_FILE_NAME

  TRY
    CALL FGL_GETFILE(ls_source,ls_destination)
  CATCH
    LET lb_return = FALSE  
    DISPLAY cs_error_tag,"The file ",lo_parameter.CLIENT_FILE_NAME," can not get to server side."
  END TRY

  RETURN lb_return   
  
END FUNCTION 

FUNCTION sadzi140_exim_set_putfile_parameter(p_param1,p_param2,p_param3,p_param4)
DEFINE 
  p_param1,p_param2,p_param3,p_param4 STRING 
DEFINE
  lo_parameter T_PUT_GET_FILE_PARA
DEFINE
  lo_return  T_PUT_GET_FILE_PARA

  #Source
  LET lo_parameter.SERVER_FILE_PATH = p_param1   
  LET lo_parameter.SERVER_FILE_NAME = p_param2
  #Destination  
  LET lo_parameter.CLIENT_FILE_PATH = p_param3   
  LET lo_parameter.CLIENT_FILE_NAME = p_param4   

  LET lo_return.* = lo_parameter.*

  RETURN lo_return.*  

END FUNCTION 

FUNCTION sadzi140_exim_set_getfile_parameter(p_param1,p_param2,p_param3,p_param4)
DEFINE 
  p_param1,p_param2,p_param3,p_param4 STRING 
DEFINE
  lo_parameter T_PUT_GET_FILE_PARA
DEFINE
  lo_return  T_PUT_GET_FILE_PARA

  #Source
  LET lo_parameter.CLIENT_FILE_PATH = p_param1   
  LET lo_parameter.CLIENT_FILE_NAME = p_param2   
  #Destination  
  LET lo_parameter.SERVER_FILE_PATH = p_param3   
  LET lo_parameter.SERVER_FILE_NAME = p_param4   

  LET lo_return.* = lo_parameter.*

  RETURN lo_return.*  

END FUNCTION 

FUNCTION sadzi140_exim_get_server_working_path(p_table_name)
DEFINE
  p_table_name STRING 
DEFINE 
  ls_working_name STRING   
DEFINE 
  lb_return BOOLEAN,
  ls_return STRING   

  LET mb_make_work_dir = TRUE
  CALL sadzi140_exim_making_work_directory(p_table_name,cs_working_dir_type_import) RETURNING lb_return,ls_working_name

  LET ls_return = ls_working_name

  RETURN lb_return,ls_return
  
END FUNCTION 

FUNCTION sadzi140_exim_check_if_new_table(p_column_info)
DEFINE
  p_column_info DYNAMIC ARRAY OF T_COLUMN_INFO
DEFINE
  lo_column_info DYNAMIC ARRAY OF T_COLUMN_INFO
DEFINE
  lb_return BOOLEAN

  LET lo_column_info = p_column_info

  IF lo_column_info.getLength() = 0 THEN
    LET lb_return = TRUE  
  ELSE
    LET lb_return = FALSE  
  END IF 
    
  RETURN lb_return
  
END FUNCTION 

FUNCTION sadzi140_exim_get_column_sequence(p_table_name,p_column_name)
DEFINE 
  p_table_name  STRING,
  p_column_name STRING
DEFINE 
  ls_table_name  STRING,
  ls_column_name STRING,
  li_sequence    INTEGER,
  li_col_count   INTEGER,
  li_loop        INTEGER
DEFINE
  li_return INTEGER

  LET ls_table_name  = p_table_name.trim()
  LET ls_column_name = p_column_name.trim()

  LET li_sequence  = -1
  LET li_col_count = mo_column_info.getLength()

  #若在原先的欄位清單找得到該欄位, 就回傳欄位Sequence
  FOR li_loop = 1 TO mo_column_info.getLength()
    IF (mo_column_info[li_loop].cs_TABLE_NAME = ls_table_name) AND
       (mo_column_info[li_loop].cs_COLUMN_NAME = ls_column_name) THEN 
      LET li_sequence = mo_column_info[li_loop].cs_SEQUENCE 
      EXIT FOR
    END IF    
  END FOR  

  #否則,就是新欄位,要將欄位的Sequence重新指定並回傳
  IF li_sequence = -1 THEN
    LET li_col_count = li_col_count + 1
    LET li_sequence = li_col_count
    LET mo_column_info[li_col_count].cs_TABLE_NAME  = ls_table_name
    LET mo_column_info[li_col_count].cs_COLUMN_NAME = ls_column_name
    LET mo_column_info[li_col_count].cs_SEQUENCE    = li_sequence
  END IF 

  LET li_return = li_sequence
  
  RETURN li_return  
  
END FUNCTION 

FUNCTION sadzi140_exim_get_column_type(p_table_name,p_column_name,p_column_type)
DEFINE 
  p_table_name  STRING,
  p_column_name STRING,
  p_column_type STRING 
DEFINE 
  ls_table_name   STRING,
  ls_column_name  STRING,
  ls_column_type  STRING, 
  ls_new_col_type STRING,
  ls_old_col_type STRING,
  ls_ret_col_type STRING, 
  ls_err_type     STRING, #2181224 
  lb_result       BOOLEAN,
  li_loop         INTEGER
DEFINE
  ls_return  STRING

  LET ls_table_name  = p_table_name.trim()
  LET ls_column_name = p_column_name.trim()
  LET ls_column_type = p_column_type.trim()

  LET lb_result = FALSE
  LET ls_return = NULL
  
  #若在原先的欄位清單找得到該欄位, 就回傳欄位Column Type
  FOR li_loop = 1 TO mo_column_info.getLength()
    IF (mo_column_info[li_loop].cs_TABLE_NAME = ls_table_name) AND
       (mo_column_info[li_loop].cs_COLUMN_NAME = ls_column_name) THEN 
      LET ls_ret_col_type = mo_column_info[li_loop].cs_COLUMN_TYPE 
      EXIT FOR
    END IF    
  END FOR  

  #否則,就是新欄位,要將傳入的型態回傳
  IF ls_ret_col_type IS NULL THEN
    LET ls_ret_col_type = ls_column_type 
  END IF 

  #如果原來的型態和傳入的型態不一致時進行檢驗
  IF ls_ret_col_type <> ls_column_type THEN
    LET ls_new_col_type = ls_column_type
    LET ls_old_col_type = ls_ret_col_type
    CALL sadzi140_db_column_type_validation(ls_new_col_type,ls_old_col_type) RETURNING lb_result,ls_err_type #20181224
    #若新的型態與舊的型態驗證通過, 則用新的, 否則用舊的
    IF lb_result THEN
      LET ls_ret_col_type = ls_new_col_type
    ELSE
      DISPLAY IIF(NVL(ls_err_type,cs_null_value)==cs_valid_err_type_err,cs_error_tag,cs_warning_tag),"Modify data type for column '",ls_column_name,"' from '",ls_old_col_type,"' to '",ls_new_col_type,"' validation failed." --20171120 #20181224
      LET ls_ret_col_type = ls_old_col_type
    END IF
  END IF

  LET ls_return = ls_ret_col_type
  
  RETURN ls_return  
  
END FUNCTION 

FUNCTION sadzi140_exim_parse_client_full_name(p_full_name,p_export_info)
DEFINE
  p_full_name    STRING, 
  p_export_info  T_EXPORT_INFO
DEFINE
  ls_full_name    STRING, 
  lo_export_info  T_EXPORT_INFO,
  ls_client_path  STRING,
  ls_client_name  STRING 
DEFINE
  lo_return T_EXPORT_INFO

  LET ls_full_name     = p_full_name
  LET lo_export_info.* = p_export_info.*

  CALL os.Path.dirname(ls_full_name) RETURNING ls_client_path
  CALL os.Path.basename(ls_full_name) RETURNING ls_client_name

  LET lo_export_info.TAR_NAME    = ls_client_name
  LET lo_export_info.CLIENT_PATH = ls_client_path

  LET lo_return.* = lo_export_info.*

  RETURN lo_return.*
  
END FUNCTION 

FUNCTION sadzi140_exim_parse_tar_file_info(p_tar_file_name)
DEFINE 
  p_tar_file_name STRING
DEFINE 
  ls_tar_file_name STRING,
  li_loop          INTEGER,
  li_word_index    INTEGER,
  lo_tar_file_info T_TAR_FILE_INFO,
  ls_tar_char      STRING,
  ls_tar_word      STRING,
  li_tar_word      INTEGER 
DEFINE 
  lo_return T_TAR_FILE_INFO

  LET ls_tar_file_name = p_tar_file_name

  LET ls_tar_word = ""
  LET li_word_index = 1

  FOR li_loop = 1 TO ls_tar_file_name.getLength()
    LET ls_tar_char = NVL(ls_tar_file_name.subString(li_loop,li_loop),"@") 
    IF (ls_tar_char = "+") OR (ls_tar_char = ".") THEN
      CASE 
        #Table Name
        WHEN li_word_index = 1
          LET lo_tar_file_info.TABLE_NAME = ls_tar_word
        #需求單號
        WHEN li_word_index = 3
          IF ls_tar_word = cs_rt_design_request_no THEN
            LET lo_tar_file_info.CONSTRUCT_VERSION = cs_rt_design_data_code
          END IF 
          #LET mo_PARAMETERS.DZLM005 = ls_tar_word
          LET mo_PARAMETERS.DZLM012 = ls_tar_word
          LET lo_tar_file_info.REQUEST_NO = ls_tar_word
        #序號
        WHEN li_word_index = 4
          LET li_tar_word = ls_tar_word
          LET ls_tar_word = li_tar_word
          #LET mo_PARAMETERS.DZLM006 = ls_tar_word
          LET mo_PARAMETERS.DZLM015 = ls_tar_word
          LET lo_tar_file_info.SEQUENCE_NO = ls_tar_word
      END CASE   
      LET li_word_index = li_word_index + 1
      LET ls_tar_word = ""
    ELSE 
      LET ls_tar_word = ls_tar_word,ls_tar_char
    END IF    
  END FOR

  LET lo_return.* = lo_tar_file_info.*

  RETURN lo_return.* 
  
END FUNCTION 

FUNCTION sadzi140_exim_reset_parameters()
  INITIALIZE mo_PARAMETERS TO NULL
  LET mb_make_work_dir = FALSE
END FUNCTION 

FUNCTION sadzi140_exim_check_if_user_define_column(p_table_name,p_column_name)
DEFINE
  p_table_name  STRING, 
  p_column_name STRING
DEFINE
  ls_table_name  STRING, 
  ls_column_name STRING,
  ls_SQL         STRING,
  li_rec_count   INTEGER,
  ls_table_cond  STRING
DEFINE
  lb_return  BOOLEAN
  
  LET ls_table_name  = p_table_name.toLowerCase()
  LET ls_column_name = p_column_name.toLowerCase()

  IF (NVL(ls_table_name,cs_null_value) = cs_null_value) THEN 
    LET ls_table_cond = ""
  ELSE 
    LET ls_table_cond = " and eb.dzeb001 = '",ls_table_name,"' "
  END IF 

  #取得資料筆數
  LET ls_SQL = " select count(*)                                  ",
               "   from dzeb_t eb                                 ",            
               "  where 1=1                                       ",
               ls_table_cond,
               "    and eb.dzeb002 = '",ls_column_name,"'         ",
               "    and eb.dzeb022 like '",cs_cdf_user_define,"%' "

  PREPARE lpre_check_if_user_define_column FROM ls_sql
  DECLARE lcur_check_if_user_define_column CURSOR FOR lpre_check_if_user_define_column
  OPEN lcur_check_if_user_define_column
  FETCH lcur_check_if_user_define_column INTO li_rec_count
  CLOSE lcur_check_if_user_define_column
  FREE lcur_check_if_user_define_column
  FREE lpre_check_if_user_define_column

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_exim_trim(p_string)
DEFINE
  p_string STRING
DEFINE
  ls_string STRING
DEFINE
  ls_return STRING

  LET ls_string = p_string

  LET ls_string = ls_string.trim()
  
  LET ls_return = ls_string
  
  RETURN ls_return

END FUNCTION

#20160519 begin  
## Import data diff with design data ###########################################

FUNCTION sadzi140_exim_verify_import_data_with_design_data(p_GUID,p_table_name)
DEFINE
  p_GUID       STRING,
  p_table_name STRING
DEFINE
  ls_GUID       STRING,
  ls_table_name STRING,
  li_fwd_result INTEGER,
  li_bwd_result INTEGER,
  lb_diff       BOOLEAN,
  lb_different  BOOLEAN
DEFINE
  lb_return  BOOLEAN
  
  LET ls_GUID = p_GUID
  LET ls_table_name = p_table_name

  LET lb_diff = FALSE
  LET lb_different = FALSE

  LET li_fwd_result = 0
  CALL sadzi140_exim_get_dzea_diff_forward(ls_GUID,ls_table_name) RETURNING li_fwd_result
  LET li_bwd_result = 0
  CALL sadzi140_exim_get_dzea_diff_backward(ls_GUID,ls_table_name) RETURNING li_bwd_result
  LET lb_diff = IIF((li_fwd_result > 0 OR li_bwd_result > 0),TRUE,FALSE)
  DISPLAY cs_message_tag,"Comparing dzea_t with import data if different : ",IIF(lb_diff,"YES","NO")
  IF lb_diff THEN LET lb_different = TRUE END IF 

  {
  LET li_fwd_result = 0
  CALL sadzi140_exim_get_dzeal_diff_forward(ls_GUID,ls_table_name) RETURNING li_fwd_result
  LET li_bwd_result = 0
  CALL sadzi140_exim_get_dzeal_diff_backward(ls_GUID,ls_table_name) RETURNING li_bwd_result
  LET lb_diff = IIF((li_fwd_result > 0 OR li_bwd_result > 0),TRUE,FALSE)
  DISPLAY cs_message_tag,"Comparing dzeal_t with import data if different : ",IIF(lb_diff,"YES","NO")
  IF lb_diff THEN LET lb_different = TRUE END IF 
  }
  
  LET li_fwd_result = 0
  CALL sadzi140_exim_get_dzeb_diff_forward(ls_GUID,ls_table_name) RETURNING li_fwd_result
  LET li_bwd_result = 0
  CALL sadzi140_exim_get_dzeb_diff_backward(ls_GUID,ls_table_name) RETURNING li_bwd_result
  LET lb_diff = IIF((li_fwd_result > 0 OR li_bwd_result > 0),TRUE,FALSE)
  DISPLAY cs_message_tag,"Comparing dzeb_t with import data if different : ",IIF(lb_diff,"YES","NO")
  IF lb_diff THEN LET lb_different = TRUE END IF 

  {
  LET li_fwd_result = 0
  CALL sadzi140_exim_get_dzebl_diff_forward(ls_GUID,ls_table_name) RETURNING li_fwd_result
  LET li_bwd_result = 0
  CALL sadzi140_exim_get_dzebl_diff_backward(ls_GUID,ls_table_name) RETURNING li_bwd_result
  LET lb_diff = IIF((li_fwd_result > 0 OR li_bwd_result > 0),TRUE,FALSE)
  DISPLAY cs_message_tag,"Comparing dzebl_t with import data if different : ",IIF(lb_diff,"YES","NO")
  IF lb_diff THEN LET lb_different = TRUE END IF 
  }
  
  LET li_fwd_result = 0
  CALL sadzi140_exim_get_dzec_diff_forward(ls_GUID,ls_table_name) RETURNING li_fwd_result
  LET li_bwd_result = 0
  CALL sadzi140_exim_get_dzec_diff_backward(ls_GUID,ls_table_name) RETURNING li_bwd_result
  LET lb_diff = IIF((li_fwd_result > 0 OR li_bwd_result > 0),TRUE,FALSE)
  DISPLAY cs_message_tag,"Comparing dzec_t with import data if different : ",IIF(lb_diff,"YES","NO")
  IF lb_diff THEN LET lb_different = TRUE END IF 

  LET li_fwd_result = 0
  CALL sadzi140_exim_get_dzed_diff_forward(ls_GUID,ls_table_name) RETURNING li_fwd_result
  LET li_bwd_result = 0
  CALL sadzi140_exim_get_dzed_diff_backward(ls_GUID,ls_table_name) RETURNING li_bwd_result
  LET lb_diff = IIF((li_fwd_result > 0 OR li_bwd_result > 0),TRUE,FALSE)
  DISPLAY cs_message_tag,"Comparing dzed_t with import data if different : ",IIF(lb_diff,"YES","NO")
  IF lb_diff THEN LET lb_different = TRUE END IF 

  LET li_fwd_result = 0
  CALL sadzi140_exim_get_dzen_diff_forward(ls_GUID,ls_table_name) RETURNING li_fwd_result
  LET li_bwd_result = 0
  CALL sadzi140_exim_get_dzen_diff_backward(ls_GUID,ls_table_name) RETURNING li_bwd_result
  LET lb_diff = IIF((li_fwd_result > 0 OR li_bwd_result > 0),TRUE,FALSE)
  DISPLAY cs_message_tag,"Comparing dzen_t with import data if different : ",IIF(lb_diff,"YES","NO")
  IF lb_diff THEN LET lb_different = TRUE END IF 

  LABEL _RETURN:

  LET lb_return = lb_different
  
  RETURN lb_return
  
END FUNCTION 

FUNCTION sadzi140_exim_get_dzea_diff_forward(p_GUID,p_table_name)
DEFINE
  p_GUID       STRING,
  p_table_name STRING
DEFINE
  ls_GUID       STRING,
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_cnt    INTEGER,
  lo_dzea_diff_forward DYNAMIC ARRAY OF T_DZEA_DIFF_DATA  
DEFINE 
  li_return  INTEGER

  LET ls_GUID = p_GUID
  LET ls_table_name = p_table_name

  LET li_rec_cnt = 1

  LET ls_sql = "select ha.dzha001,ha.dzha002,ha.dzha003,ha.dzha004,ha.dzha005, ",
               "       ha.dzha006,ha.dzha007,ha.dzha008,ha.dzha009,ha.dzha010, ",
               "       ha.dzha014,ha.dzha015,ha.dzha016,ha.dzha017,ha.dzha018  ",
               "  from dzha_t ha                                               ",
               " where 1=1                                                     ",
               "   and ha.dzha000 = '",ls_GUID,"'                              ",
               "   and ha.dzha001 = '",ls_table_name,"'                        ",
               &ifdef is_msv 
               "except ",
               &else
               &ifdef is_pgs 
               "except ",
               &else
               "minus ",
               &endif
               &endif
               "select ea.dzea001,ea.dzea002,ea.dzea003,ea.dzea004,ea.dzea005, ",
               "       ea.dzea006,ea.dzea007,ea.dzea008,ea.dzea009,ea.dzea010, ",
               "       ea.dzea014,ea.dzea015,ea.dzea016,ea.dzea017,ea.dzea018  ",
               "  from dzea_t ea                                               ",
               " where ea.dzea001 = '",ls_table_name,"'                        "
               
  PREPARE lpre_get_dzea_diff_forward FROM ls_sql
  DECLARE lcur_get_dzea_diff_forward CURSOR FOR lpre_get_dzea_diff_forward

  OPEN lcur_get_dzea_diff_forward
  FOREACH lcur_get_dzea_diff_forward INTO lo_dzea_diff_forward[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_dzea_diff_forward

  FREE lpre_get_dzea_diff_forward
  FREE lcur_get_dzea_diff_forward    

  CALL lo_dzea_diff_forward.deleteElement(li_rec_cnt)
  
  LET li_return = lo_dzea_diff_forward.getLength()
  
  RETURN li_return 
  
END FUNCTION

FUNCTION sadzi140_exim_get_dzea_diff_backward(p_GUID,p_table_name)
DEFINE
  p_GUID       STRING,
  p_table_name STRING
DEFINE
  ls_GUID       STRING,
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_cnt    INTEGER,
  lo_dzea_diff_backward DYNAMIC ARRAY OF T_DZEA_DIFF_DATA  
DEFINE 
  li_return  INTEGER

  LET ls_GUID = p_GUID
  LET ls_table_name = p_table_name

  LET li_rec_cnt = 1

  LET ls_sql = "select ea.dzea001,ea.dzea002,ea.dzea003,ea.dzea004,ea.dzea005, ",
               "       ea.dzea006,ea.dzea007,ea.dzea008,ea.dzea009,ea.dzea010, ",
               "       ea.dzea014,ea.dzea015,ea.dzea016,ea.dzea017,ea.dzea018  ",
               "  from dzea_t ea                                               ",
               " where ea.dzea001 = '",ls_table_name,"'                        ",
               &ifdef is_msv 
               "except ",
               &else
               &ifdef is_pgs 
               "except ",
               &else
               "minus ",
               &endif
               &endif
               "select ha.dzha001,ha.dzha002,ha.dzha003,ha.dzha004,ha.dzha005, ",
               "       ha.dzha006,ha.dzha007,ha.dzha008,ha.dzha009,ha.dzha010, ",
               "       ha.dzha014,ha.dzha015,ha.dzha016,ha.dzha017,ha.dzha018  ",
               "  from dzha_t ha                                               ",
               " where 1=1                                                     ",
               "   and ha.dzha000 = '",ls_GUID,"'                              ",
               "   and ha.dzha001 = '",ls_table_name,"'                        "
               
  PREPARE lpre_get_dzea_diff_backward FROM ls_sql
  DECLARE lcur_get_dzea_diff_backward CURSOR FOR lpre_get_dzea_diff_backward

  OPEN lcur_get_dzea_diff_backward
  FOREACH lcur_get_dzea_diff_backward INTO lo_dzea_diff_backward[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_dzea_diff_backward

  FREE lpre_get_dzea_diff_backward
  FREE lcur_get_dzea_diff_backward    
  
  CALL lo_dzea_diff_backward.deleteElement(li_rec_cnt)
  
  LET li_return = lo_dzea_diff_backward.getLength()
  
  RETURN li_return 
  
END FUNCTION

FUNCTION sadzi140_exim_get_dzeal_diff_forward(p_GUID,p_table_name)
DEFINE
  p_GUID       STRING,
  p_table_name STRING
DEFINE
  ls_GUID       STRING,
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_cnt    INTEGER,
  lo_dzeal_diff_forward DYNAMIC ARRAY OF T_DZEAL_DIFF_DATA  
DEFINE 
  li_return  INTEGER

  LET ls_GUID = p_GUID
  LET ls_table_name = p_table_name

  LET li_rec_cnt = 1

  LET ls_sql = "select hal.dzhal001,hal.dzhal002,hal.dzhal003,hal.dzhal004,hal.dzhal005  ",
               "  from dzhal_t hal                                                       ",
               " where 1=1                                                               ",
               "   and hal.dzhal000 = '",ls_GUID,"'                                      ",
               "   and hal.dzhal001 = '",ls_table_name,"'                                ",
               &ifdef is_msv 
               "except ",
               &else
               &ifdef is_pgs 
               "except ",
               &else
               "minus ",
               &endif
               &endif
               "select eal.dzeal001,eal.dzeal002,eal.dzeal003,eal.dzeal004,eal.dzeal005  ",
               "  from dzeal_t eal                                                       ",
               " where eal.dzeal001 = '",ls_table_name,"'                                "
               
  PREPARE lpre_get_dzeal_diff_forward FROM ls_sql
  DECLARE lcur_get_dzeal_diff_forward CURSOR FOR lpre_get_dzeal_diff_forward

  OPEN lcur_get_dzeal_diff_forward
  FOREACH lcur_get_dzeal_diff_forward INTO lo_dzeal_diff_forward[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_dzeal_diff_forward

  FREE lpre_get_dzeal_diff_forward
  FREE lcur_get_dzeal_diff_forward    
  
  CALL lo_dzeal_diff_forward.deleteElement(li_rec_cnt)
  
  LET li_return = lo_dzeal_diff_forward.getLength()
  
  RETURN li_return 
  
END FUNCTION

FUNCTION sadzi140_exim_get_dzeal_diff_backward(p_GUID,p_table_name)
DEFINE
  p_GUID       STRING,
  p_table_name STRING
DEFINE
  ls_GUID       STRING,
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_cnt    INTEGER,
  lo_dzeal_diff_backward DYNAMIC ARRAY OF T_DZEAL_DIFF_DATA  
DEFINE 
  li_return  INTEGER

  LET ls_GUID = p_GUID
  LET ls_table_name = p_table_name

  LET li_rec_cnt = 1

  LET ls_sql = "select eal.dzeal001,eal.dzeal002,eal.dzeal003,eal.dzeal004,eal.dzeal005  ",
               "  from dzeal_t eal                                                       ",
               " where eal.dzeal001 = '",ls_table_name,"'                                ",
               &ifdef is_msv 
               "except ",
               &else
               &ifdef is_pgs 
               "except ",
               &else
               "minus ",
               &endif
               &endif
               "select hal.dzhal001,hal.dzhal002,hal.dzhal003,hal.dzhal004,hal.dzhal005  ",
               "  from dzhal_t hal                                                       ",
               " where 1=1                                                               ",
               "   and hal.dzhal000 = '",ls_GUID,"'                                      ",
               "   and hal.dzhal001 = '",ls_table_name,"'                                "
               
  PREPARE lpre_get_dzeal_diff_backward FROM ls_sql
  DECLARE lcur_get_dzeal_diff_backward CURSOR FOR lpre_get_dzeal_diff_backward

  OPEN lcur_get_dzeal_diff_backward
  FOREACH lcur_get_dzeal_diff_backward INTO lo_dzeal_diff_backward[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_dzeal_diff_backward

  FREE lpre_get_dzeal_diff_backward
  FREE lcur_get_dzeal_diff_backward    
  
  CALL lo_dzeal_diff_backward.deleteElement(li_rec_cnt)
  
  LET li_return = lo_dzeal_diff_backward.getLength()
  
  RETURN li_return 
  
END FUNCTION

FUNCTION sadzi140_exim_get_dzeb_diff_forward(p_GUID,p_table_name)
DEFINE
  p_GUID       STRING,
  p_table_name STRING
DEFINE
  ls_GUID       STRING,
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_cnt    INTEGER,
  lo_dzeb_diff_forward DYNAMIC ARRAY OF T_DZEB_DIFF_DATA  
DEFINE 
  li_return  INTEGER

  LET ls_GUID = p_GUID
  LET ls_table_name = p_table_name

  LET li_rec_cnt = 1

  LET ls_sql = "select hb.dzhb001,hb.dzhb002,hb.dzhb003,hb.dzhb004,hb.dzhb005, ", 
               "       hb.dzhb006,hb.dzhb007,hb.dzhb008,hb.dzhb021,hb.dzhb022, ",
               "       hb.dzhb023,hb.dzhb029,hb.dzhb012,hb.dzhb030             ",
               "  from dzhb_t hb                                               ",
               " where 1=1                                                     ",
               "   and hb.dzhb000 = '",ls_GUID,"'                              ",
               "   and hb.dzhb001 = '",ls_table_name,"'                        ",
               &ifdef is_msv 
               "except ",
               &else
               &ifdef is_pgs 
               "except ",
               &else
               "minus ",
               &endif
               &endif
               "select eb.dzeb001,eb.dzeb002,eb.dzeb003,eb.dzeb004,eb.dzeb005, ", 
               "       eb.dzeb006,eb.dzeb007,eb.dzeb008,eb.dzeb021,eb.dzeb022, ",
               "       eb.dzeb023,eb.dzeb029,eb.dzeb012,eb.dzeb030             ",
               "  from dzeb_t eb                                               ",
               " where eb.dzeb001 = '",ls_table_name,"'                        "
               
  PREPARE lpre_get_dzeb_diff_forward FROM ls_sql
  DECLARE lcur_get_dzeb_diff_forward CURSOR FOR lpre_get_dzeb_diff_forward

  OPEN lcur_get_dzeb_diff_forward
  FOREACH lcur_get_dzeb_diff_forward INTO lo_dzeb_diff_forward[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_dzeb_diff_forward

  FREE lpre_get_dzeb_diff_forward
  FREE lcur_get_dzeb_diff_forward    
  
  CALL lo_dzeb_diff_forward.deleteElement(li_rec_cnt)
  
  LET li_return = lo_dzeb_diff_forward.getLength()
  
  RETURN li_return 
  
END FUNCTION

FUNCTION sadzi140_exim_get_dzeb_diff_backward(p_GUID,p_table_name)
DEFINE
  p_GUID       STRING,
  p_table_name STRING
DEFINE
  ls_GUID       STRING,
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_cnt    INTEGER,
  lo_dzeb_diff_backward DYNAMIC ARRAY OF T_DZEB_DIFF_DATA  
DEFINE 
  li_return  INTEGER

  LET ls_GUID = p_GUID
  LET ls_table_name = p_table_name

  LET li_rec_cnt = 1

  LET ls_sql = "select eb.dzeb001,eb.dzeb002,eb.dzeb003,eb.dzeb004,eb.dzeb005, ", 
               "       eb.dzeb006,eb.dzeb007,eb.dzeb008,eb.dzeb021,eb.dzeb022, ",
               "       eb.dzeb023,eb.dzeb029,eb.dzeb012,eb.dzeb030             ",
               "  from dzeb_t eb                                               ",
               " where eb.dzeb001 = '",ls_table_name,"'                        ",
               &ifdef is_msv 
               "except ",
               &else
               &ifdef is_pgs 
               "except ",
               &else
               "minus ",
               &endif
               &endif
               "select hb.dzhb001,hb.dzhb002,hb.dzhb003,hb.dzhb004,hb.dzhb005, ", 
               "       hb.dzhb006,hb.dzhb007,hb.dzhb008,hb.dzhb021,hb.dzhb022, ",
               "       hb.dzhb023,hb.dzhb029,hb.dzhb012,hb.dzhb030             ",
               "  from dzhb_t hb                                               ",
               " where 1=1                                                     ",
               "   and hb.dzhb000 = '",ls_GUID,"'                              ",
               "   and hb.dzhb001 = '",ls_table_name,"'                        "
               
  PREPARE lpre_get_dzeb_diff_backward FROM ls_sql
  DECLARE lcur_get_dzeb_diff_backward CURSOR FOR lpre_get_dzeb_diff_backward

  OPEN lcur_get_dzeb_diff_backward
  FOREACH lcur_get_dzeb_diff_backward INTO lo_dzeb_diff_backward[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_dzeb_diff_backward

  FREE lpre_get_dzeb_diff_backward
  FREE lcur_get_dzeb_diff_backward    

  CALL lo_dzeb_diff_backward.deleteElement(li_rec_cnt)
  
  LET li_return = lo_dzeb_diff_backward.getLength()
  
  RETURN li_return 
  
END FUNCTION

FUNCTION sadzi140_exim_get_dzebl_diff_forward(p_GUID,p_table_name)
DEFINE
  p_GUID       STRING,
  p_table_name STRING
DEFINE
  ls_GUID       STRING,
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_cnt    INTEGER,
  lo_dzebl_diff_forward DYNAMIC ARRAY OF T_DZEBL_DIFF_DATA  
DEFINE 
  li_return  INTEGER

  LET ls_GUID = p_GUID
  LET ls_table_name = p_table_name

  LET li_rec_cnt = 1

  LET ls_sql = "select hbl.dzhbl001,hbl.dzhbl002,hbl.dzhbl003,hbl.dzhbl004,hbl.dzhbl005  ",
               "  from dzhbl_t hbl                                                       ",
               " where 1=1                                                               ",
               "   and hbl.dzhbl000 = '",ls_guid,"'                                      ",
               "   and hbl.dzhbl001 = '",ls_table_name,"'                                ",
               &ifdef is_msv 
               "except ",
               &else
               &ifdef is_pgs 
               "except ",
               &else
               "minus ",
               &endif
               &endif
               "select ebl.dzebl000,ebl.dzebl001,ebl.dzebl002,ebl.dzebl003,ebl.dzebl004  ",
               "  from dzebl_t ebl                                                       ",
               " where ebl.dzebl000 = '",ls_table_name,"'                                "
               
  PREPARE lpre_get_dzebl_diff_forward FROM ls_sql
  DECLARE lcur_get_dzebl_diff_forward CURSOR FOR lpre_get_dzebl_diff_forward

  OPEN lcur_get_dzebl_diff_forward
  FOREACH lcur_get_dzebl_diff_forward INTO lo_dzebl_diff_forward[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_dzebl_diff_forward

  FREE lpre_get_dzebl_diff_forward
  FREE lcur_get_dzebl_diff_forward    

  CALL lo_dzebl_diff_forward.deleteElement(li_rec_cnt)
  
  LET li_return = lo_dzebl_diff_forward.getLength()
  
  RETURN li_return 
  
END FUNCTION

FUNCTION sadzi140_exim_get_dzebl_diff_backward(p_GUID,p_table_name)
DEFINE
  p_GUID       STRING,
  p_table_name STRING
DEFINE
  ls_GUID       STRING,
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_cnt    INTEGER,
  lo_dzebl_diff_backward DYNAMIC ARRAY OF T_DZEBL_DIFF_DATA  
DEFINE 
  li_return  INTEGER

  LET ls_GUID = p_GUID
  LET ls_table_name = p_table_name

  LET li_rec_cnt = 1

  LET ls_sql = "select ebl.dzebl000,ebl.dzebl001,ebl.dzebl002,ebl.dzebl003,ebl.dzebl004  ",
               "  from dzebl_t ebl                                                       ",
               " where ebl.dzebl000 = '",ls_table_name,"'                                ",
               &ifdef is_msv 
               "except ",
               &else
               &ifdef is_pgs 
               "except ",
               &else
               "minus ",
               &endif
               &endif
               "select hbl.dzhbl001,hbl.dzhbl002,hbl.dzhbl003,hbl.dzhbl004,hbl.dzhbl005  ",
               "  from dzhbl_t hbl                                                       ",
               " where 1=1                                                               ",
               "   and hbl.dzhbl000 = '",ls_GUID,"'                                      ",
               "   and hbl.dzhbl001 = '",ls_table_name,"'                                "
               
  PREPARE lpre_get_dzebl_diff_backward FROM ls_sql
  DECLARE lcur_get_dzebl_diff_backward CURSOR FOR lpre_get_dzebl_diff_backward

  OPEN lcur_get_dzebl_diff_backward
  FOREACH lcur_get_dzebl_diff_backward INTO lo_dzebl_diff_backward[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_dzebl_diff_backward

  FREE lpre_get_dzebl_diff_backward
  FREE lcur_get_dzebl_diff_backward    

  CALL lo_dzebl_diff_backward.deleteElement(li_rec_cnt)
  
  LET li_return = lo_dzebl_diff_backward.getLength()
  
  RETURN li_return 
  
END FUNCTION

FUNCTION sadzi140_exim_get_dzec_diff_forward(p_GUID,p_table_name)
DEFINE
  p_GUID       STRING,
  p_table_name STRING
DEFINE
  ls_GUID       STRING,
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_cnt    INTEGER,
  lo_dzec_diff_forward DYNAMIC ARRAY OF T_DZEC_DIFF_DATA  
DEFINE 
  li_return  INTEGER

  LET ls_GUID = p_GUID
  LET ls_table_name = p_table_name

  LET li_rec_cnt = 1

  LET ls_sql = "select hc.dzhc001,hc.dzhc002,hc.dzhc003,hc.dzhc004             ",
               "  from dzhc_t hc                                               ",
               " where 1=1                                                     ",
               "   and hc.dzhc000 = '",ls_guid,"'                              ",
               "   and hc.dzhc001 = '",ls_table_name,"'                        ",
               &ifdef is_msv 
               "except ",
               &else
               &ifdef is_pgs 
               "except ",
               &else
               "minus ",
               &endif
               &endif
               "select ec.dzec001,ec.dzec002,ec.dzec003,ec.dzec004             ",
               "  from dzec_t ec                                               ",
               " where ec.dzec001 = '",ls_table_name,"'                        "
               
  PREPARE lpre_get_dzec_diff_forward FROM ls_sql
  DECLARE lcur_get_dzec_diff_forward CURSOR FOR lpre_get_dzec_diff_forward

  OPEN lcur_get_dzec_diff_forward
  FOREACH lcur_get_dzec_diff_forward INTO lo_dzec_diff_forward[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_dzec_diff_forward

  FREE lpre_get_dzec_diff_forward
  FREE lcur_get_dzec_diff_forward    

  CALL lo_dzec_diff_forward.deleteElement(li_rec_cnt)
  
  LET li_return = lo_dzec_diff_forward.getLength()
  
  RETURN li_return 
  
END FUNCTION

FUNCTION sadzi140_exim_get_dzec_diff_backward(p_GUID,p_table_name)
DEFINE
  p_GUID       STRING,
  p_table_name STRING
DEFINE
  ls_GUID       STRING,
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_cnt    INTEGER,
  lo_dzec_diff_backward DYNAMIC ARRAY OF T_DZEC_DIFF_DATA  
DEFINE 
  li_return  INTEGER

  LET ls_GUID = p_GUID
  LET ls_table_name = p_table_name

  LET li_rec_cnt = 1

  LET ls_sql = "select ec.dzec001,ec.dzec002,ec.dzec003,ec.dzec004             ",
               "  from dzec_t ec                                               ",
               " where ec.dzec001 = '",ls_table_name,"'                        ",
               &ifdef is_msv 
               "except ",
               &else
               &ifdef is_pgs 
               "except ",
               &else
               "minus ",
               &endif
               &endif
               "select hc.dzhc001,hc.dzhc002,hc.dzhc003,hc.dzhc004             ",
               "  from dzhc_t hc                                               ",
               " where 1=1                                                     ",
               "   and hc.dzhc000 = '",ls_GUID,"'                              ",
               "   and hc.dzhc001 = '",ls_table_name,"'                        "
               
  PREPARE lpre_get_dzec_diff_backward FROM ls_sql
  DECLARE lcur_get_dzec_diff_backward CURSOR FOR lpre_get_dzec_diff_backward

  OPEN lcur_get_dzec_diff_backward
  FOREACH lcur_get_dzec_diff_backward INTO lo_dzec_diff_backward[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_dzec_diff_backward

  FREE lpre_get_dzec_diff_backward
  FREE lcur_get_dzec_diff_backward    

  CALL lo_dzec_diff_backward.deleteElement(li_rec_cnt)
  
  LET li_return = lo_dzec_diff_backward.getLength()
  
  RETURN li_return 
  
END FUNCTION

FUNCTION sadzi140_exim_get_dzed_diff_forward(p_GUID,p_table_name)
DEFINE
  p_GUID       STRING,
  p_table_name STRING
DEFINE
  ls_GUID       STRING,
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_cnt    INTEGER,
  lo_dzed_diff_forward DYNAMIC ARRAY OF T_DZED_DIFF_DATA  
DEFINE 
  li_return  INTEGER

  LET ls_GUID = p_GUID
  LET ls_table_name = p_table_name

  LET li_rec_cnt = 1

  LET ls_sql = "select hd.dzhd001,hd.dzhd002,hd.dzhd003,hd.dzhd004,hd.dzhd005, ",
               "       hd.dzhd006                                              ",  
               "  from dzhd_t hd                                               ",
               " where 1=1                                                     ",
               "   and hd.dzhd000 = '",ls_GUID,"'                              ",
               "   and hd.dzhd001 = '",ls_table_name,"'                        ",
               &ifdef is_msv 
               "except ",
               &else
               &ifdef is_pgs 
               "except ",
               &else
               "minus ",
               &endif
               &endif
               "select ed.dzed001,ed.dzed002,ed.dzed003,ed.dzed004,ed.dzed005, ",
               "       ed.dzed006                                              ",
               "  from dzed_t ed                                               ",
               " where ed.dzed001 = '",ls_table_name,"'                        "
               
  PREPARE lpre_get_dzed_diff_forward FROM ls_sql
  DECLARE lcur_get_dzed_diff_forward CURSOR FOR lpre_get_dzed_diff_forward

  OPEN lcur_get_dzed_diff_forward
  FOREACH lcur_get_dzed_diff_forward INTO lo_dzed_diff_forward[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_dzed_diff_forward

  FREE lpre_get_dzed_diff_forward
  FREE lcur_get_dzed_diff_forward    

  CALL lo_dzed_diff_forward.deleteElement(li_rec_cnt)
  
  LET li_return = lo_dzed_diff_forward.getLength()
  
  RETURN li_return 
  
END FUNCTION

FUNCTION sadzi140_exim_get_dzed_diff_backward(p_GUID,p_table_name)
DEFINE
  p_GUID       STRING,
  p_table_name STRING
DEFINE
  ls_GUID       STRING,
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_cnt    INTEGER,
  lo_dzed_diff_backward DYNAMIC ARRAY OF T_DZED_DIFF_DATA  
DEFINE 
  li_return  INTEGER

  LET ls_GUID = p_GUID
  LET ls_table_name = p_table_name

  LET li_rec_cnt = 1

  LET ls_sql = "select ed.dzed001,ed.dzed002,ed.dzed003,ed.dzed004,ed.dzed005, ",
               "       ed.dzed006                                              ",
               "  from dzed_t ed                                               ",
               " where ed.dzed001 = '",ls_table_name,"'                        ",
               &ifdef is_msv 
               "except ",
               &else
               &ifdef is_pgs 
               "except ",
               &else
               "minus ",
               &endif
               &endif
               "select hd.dzhd001,hd.dzhd002,hd.dzhd003,hd.dzhd004,hd.dzhd005, ",
               "       hd.dzhd006                                              ",
               "  from dzhd_t hd                                               ",
               " where 1=1                                                     ",
               "   and hd.dzhd000 = '",ls_GUID,"'                              ",
               "   and hd.dzhd001 = '",ls_table_name,"'                        "
               
  PREPARE lpre_get_dzed_diff_backward FROM ls_sql
  DECLARE lcur_get_dzed_diff_backward CURSOR FOR lpre_get_dzed_diff_backward

  OPEN lcur_get_dzed_diff_backward
  FOREACH lcur_get_dzed_diff_backward INTO lo_dzed_diff_backward[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_dzed_diff_backward

  FREE lpre_get_dzed_diff_backward
  FREE lcur_get_dzed_diff_backward    

  CALL lo_dzed_diff_backward.deleteElement(li_rec_cnt)
  
  LET li_return = lo_dzed_diff_backward.getLength()
  
  RETURN li_return 
  
END FUNCTION

FUNCTION sadzi140_exim_get_dzen_diff_forward(p_GUID,p_table_name)
DEFINE
  p_GUID       STRING,
  p_table_name STRING
DEFINE
  ls_GUID       STRING,
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_cnt    INTEGER,
  lo_dzen_diff_forward DYNAMIC ARRAY OF T_DZEN_DIFF_DATA  
DEFINE 
  li_return  INTEGER

  LET ls_GUID = p_GUID
  LET ls_table_name = p_table_name

  LET li_rec_cnt = 1

  LET ls_sql = "select hn.dzhn001,hn.dzhn002,hn.dzhn003,hn.dzhn004,hn.dzhn005, ",
               "       hn.dzhn006,hn.dzhn007,hn.dzhn008,hn.dzhn009,hn.dzhn010  ",  
               "  from dzhn_t hn                                               ",
               " where 1=1                                                     ",
               "   and hn.dzhn000 = '",ls_GUID,"'                              ",
               "   and hn.dzhn001 = '",ls_table_name,"'                        ",
               &ifdef is_msv 
               "except ",
               &else
               &ifdef is_pgs 
               "except ",
               &else
               "minus ",
               &endif
               &endif
               "select en.dzen001,en.dzen002,en.dzen003,en.dzen004,en.dzen005, ",
               "       en.dzen006,en.dzen007,en.dzen008,en.dzen009,en.dzen010  ",
               "  from dzen_t en                                               ",
               " where en.dzen001 = '",ls_table_name,"'                        "
               
  PREPARE lpre_get_dzen_diff_forward FROM ls_sql
  DECLARE lcur_get_dzen_diff_forward CURSOR FOR lpre_get_dzen_diff_forward

  OPEN lcur_get_dzen_diff_forward
  FOREACH lcur_get_dzen_diff_forward INTO lo_dzen_diff_forward[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_dzen_diff_forward

  FREE lpre_get_dzen_diff_forward
  FREE lcur_get_dzen_diff_forward    

  CALL lo_dzen_diff_forward.deleteElement(li_rec_cnt)
  
  LET li_return = lo_dzen_diff_forward.getLength()
  
  RETURN li_return 
  
END FUNCTION

FUNCTION sadzi140_exim_get_dzen_diff_backward(p_GUID,p_table_name)
DEFINE
  p_GUID       STRING,
  p_table_name STRING
DEFINE
  ls_GUID       STRING,
  ls_table_name STRING,
  ls_sql        STRING,
  li_rec_cnt    INTEGER,
  lo_dzen_diff_backward DYNAMIC ARRAY OF T_DZEN_DIFF_DATA  
DEFINE 
  li_return  INTEGER

  LET ls_GUID = p_GUID
  LET ls_table_name = p_table_name

  LET li_rec_cnt = 1

  LET ls_sql = "select en.dzen001,en.dzen002,en.dzen003,en.dzen004,en.dzen005, ",
               "       en.dzen006,en.dzen007,en.dzen008,en.dzen009,en.dzen010  ",
               "  from dzen_t en                                               ",
               " where en.dzen001 = '",ls_table_name,"'                        ",
               &ifdef is_msv 
               "except ",
               &else
               &ifdef is_pgs 
               "except ",
               &else
               "minus ",
               &endif
               &endif
               "select hn.dzhn001,hn.dzhn002,hn.dzhn003,hn.dzhn004,hn.dzhn005, ",
               "       hn.dzhn006,hn.dzhn007,hn.dzhn008,hn.dzhn009,hn.dzhn010  ",  
               "  from dzhn_t hn                                               ",
               " where 1=1                                                     ",
               "   and hn.dzhn000 = '",ls_GUID,"'                              ",
               "   and hn.dzhn001 = '",ls_table_name,"'                        "
               
  PREPARE lpre_get_dzen_diff_backward FROM ls_sql
  DECLARE lcur_get_dzen_diff_backward CURSOR FOR lpre_get_dzen_diff_backward

  OPEN lcur_get_dzen_diff_backward
  FOREACH lcur_get_dzen_diff_backward INTO lo_dzen_diff_backward[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_dzen_diff_backward

  FREE lpre_get_dzen_diff_backward
  FREE lcur_get_dzen_diff_backward    

  CALL lo_dzen_diff_backward.deleteElement(li_rec_cnt)
  
  LET li_return = lo_dzen_diff_backward.getLength()
  
  RETURN li_return 
  
END FUNCTION
#20160519 end

#20170208 begin
FUNCTION sadzi140_exim_get_table_module(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_return       STRING,
  ls_sql          STRING,
  ls_table_module VARCHAR(10),
  ls_table_name   STRING

  LET ls_table_name = p_table_name

  LET ls_sql = "select ea.dzea003                      ",
               "  from dzea_t ea                       ",
               " where ea.dzea001 = '",ls_table_name,"'"
 
  PREPARE lpre_get_table_module FROM ls_sql
  DECLARE lcur_get_table_module CURSOR FOR lpre_get_table_module
  OPEN lcur_get_table_module
  FETCH lcur_get_table_module INTO ls_table_module
  CLOSE lcur_get_table_module
  FREE lcur_get_table_module
  FREE lpre_get_table_module

  LET ls_return = ls_table_module
  
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_exim_get_table_desc(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_return       STRING,
  ls_sql          STRING,
  ls_table_desc VARCHAR(10),
  ls_table_name   STRING

  LET ls_table_name = p_table_name

  LET ls_sql = "select ea.dzea002                      ",
               "  from dzea_t ea                       ",
               " where ea.dzea001 = '",ls_table_name,"'"
 
  PREPARE lpre_get_table_desc FROM ls_sql
  DECLARE lcur_get_table_desc CURSOR FOR lpre_get_table_desc
  OPEN lcur_get_table_desc
  FETCH lcur_get_table_desc INTO ls_table_desc
  CLOSE lcur_get_table_desc
  FREE lcur_get_table_desc
  FREE lpre_get_table_desc

  LET ls_return = ls_table_desc
  
  RETURN ls_return
  
END FUNCTION
#20170208 end

#20170912 begin
FUNCTION sadzi140_exim_get_if_dest_cols_much_than_source(p_guid,p_table_name)
DEFINE
  p_guid        STRING,
  p_table_name  STRING
DEFINE
  ls_guid         STRING,
  ls_table_name   STRING,
  li_dzeb_counts  INTEGER,
  li_dzhb_counts  INTEGER
DEFINE
  lb_return BOOLEAN

  LET ls_guid       = p_guid
  LET ls_table_name = p_table_name

  CALL sadzi140_exim_get_column_count_by_dzeb(ls_table_name) RETURNING li_dzeb_counts
  CALL sadzi140_exim_get_column_count_by_dzhb(ls_guid,ls_table_name) RETURNING li_dzhb_counts

  LET lb_return = IIF(li_dzeb_counts > li_dzhb_counts,TRUE,FALSE)

  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_exim_get_column_count_by_dzeb(p_table_name)
DEFINE
  p_table_name  STRING
DEFINE
  ls_table_name   STRING,
  ls_sql          STRING,
  li_column_count INTEGER
DEFINE
  li_return INTEGER  

  LET ls_table_name = p_table_name

  LET ls_sql = "select count(1)                         ",
               "  from dzeb_t eb                        ",
               " where eb.dzeb001 = '",ls_table_name,"' "
 
  PREPARE lpre_get_column_count_by_dzeb FROM ls_sql
  DECLARE lcur_get_column_count_by_dzeb CURSOR FOR lpre_get_column_count_by_dzeb
  OPEN lcur_get_column_count_by_dzeb
  FETCH lcur_get_column_count_by_dzeb INTO li_column_count
  CLOSE lcur_get_column_count_by_dzeb
  FREE lcur_get_column_count_by_dzeb
  FREE lpre_get_column_count_by_dzeb

  LET li_return = li_column_count
  
  RETURN li_return
  
END FUNCTION

FUNCTION sadzi140_exim_get_column_count_by_dzhb(p_guid,p_table_name)
DEFINE
  p_guid        STRING,
  p_table_name  STRING
DEFINE
  ls_guid         STRING,
  ls_table_name   STRING,
  ls_sql          STRING,
  li_column_count INTEGER
DEFINE
  li_return INTEGER  

  LET ls_guid       = p_guid
  LET ls_table_name = p_table_name

  LET ls_sql = "select count(1)                         ",
               "  from dzhb_t hb                        ",
               " where hb.dzhb000 = '",ls_guid,"'       ",
               "   and hb.dzhb001 = '",ls_table_name,"' "
 
  PREPARE lpre_get_column_count_by_dzhb FROM ls_sql
  DECLARE lcur_get_column_count_by_dzhb CURSOR FOR lpre_get_column_count_by_dzhb
  OPEN lcur_get_column_count_by_dzhb
  FETCH lcur_get_column_count_by_dzhb INTO li_column_count
  CLOSE lcur_get_column_count_by_dzhb
  FREE lcur_get_column_count_by_dzhb
  FREE lpre_get_column_count_by_dzhb

  LET li_return = li_column_count
  
  RETURN li_return
  
END FUNCTION

#20170912 end

#20171120 begin
FUNCTION sadzi140_exim_refresh_extract_data(p_extract_data)
DEFINE
  p_extract_data DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_extract_data  DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH, 
  lo_refresh_data  DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH, 
  lo_column_array  DYNAMIC ARRAY OF T_COLUMN_LIST,
  lb_col_exist     BOOLEAN,
  li_loop          INTEGER,
  li_col_info      INTEGER,
  li_count         INTEGER
DEFINE
  lo_return  DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH

  LET lo_extract_data = p_extract_data
  
  LET li_count = 0
  FOR li_loop = 1 TO lo_extract_data.getLength()
    IF lo_extract_data[li_loop,2] IS NOT NULL THEN
      LET lb_col_exist = FALSE
      FOR li_col_info = 1 TO mo_column_info.getLength()
        IF mo_column_info[li_col_info].cs_COLUMN_NAME = lo_extract_data[li_loop,2] AND
           mo_column_info[li_col_info].cs_COLUMN_TYPE = lo_extract_data[li_loop,6] THEN 
          LET lb_col_exist = TRUE
          EXIT FOR
        END IF
      END FOR
      IF NOT lb_col_exist THEN 
        LET lo_refresh_data[lo_refresh_data.getLength()+1] = lo_extract_data[li_loop]
      END IF  
    END IF
  END FOR

  LET lo_return = lo_refresh_data
  
  RETURN  lo_return
 
END FUNCTION

FUNCTION sadzi140_exim_get_col_array_from_extract_data(p_extract_data)
DEFINE
  p_extract_data DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH
DEFINE
  lo_extract_data  DYNAMIC ARRAY WITH DIMENSION 2 OF T_COL_DATA_LENGTH, 
  lo_column_array  DYNAMIC ARRAY OF T_COLUMN_LIST,
  li_loop          INTEGER,
  li_count         INTEGER
DEFINE
  lo_return  DYNAMIC ARRAY OF T_COLUMN_LIST

  LET lo_extract_data = p_extract_data

  LET li_count = 0
  FOR li_loop = 1 TO lo_extract_data.getLength()
    IF lo_extract_data[li_loop,2] IS NOT NULL THEN
      LET li_count = li_count + 1
      LET lo_column_array[li_count].COLUMN_NAME = lo_extract_data[li_loop,2] 
      LET lo_column_array[li_count].COLUMN_DESC = lo_extract_data[li_loop,3] 
    END IF
  END FOR

  LET lo_return = lo_column_array
  
  RETURN  lo_return
 
END FUNCTION

FUNCTION sadzi140_exim_export_to_pickup_column_json(p_table_name,p_pickup_columns)
DEFINE
  p_table_name      STRING, 
  p_pickup_columns  DYNAMIC ARRAY OF T_COLUMN_LIST
DEFINE
  ls_table_name       STRING, 
  lo_pickup_columns   DYNAMIC ARRAY OF T_COLUMN_LIST,
  lo_exp_header       T_PICKUP_COLUMN_HEADER,
  ls_json_data        STRING,
  ls_GUID             STRING,
  lb_result           BOOLEAN,
  ls_pickup_json_name STRING
DEFINE
  lb_return  BOOLEAN,
  ls_return  STRING  

  LET ls_table_name     = p_table_name
  LET lo_pickup_columns = p_pickup_columns

  LET ls_GUID = security.RandomGenerator.CreateUUIDString()
  #LET ls_pickup_json_name = ls_table_name,"+pickup_cols+",ls_GUID,".json"
  LET ls_pickup_json_name = ls_table_name,"+PICKUP_COLS.json"

  CALL sadzi140_exim_get_pickup_column_header() RETURNING lo_exp_header.*
  CALL sadzi140_exim_get_pickup_column_body(lo_exp_header.*,lo_pickup_columns) RETURNING ls_json_data
  CALL sadzi140_exim_write_file(ls_pickup_json_name,ls_json_data) RETURNING lb_result

  LET lb_return = lb_result
  LET ls_return = ls_pickup_json_name

  RETURN lb_return,ls_return
  
END FUNCTION

FUNCTION sadzi140_exim_get_pickup_column_header()
DEFINE
  lo_exp_header  T_PICKUP_COLUMN_HEADER
DEFINE
  lo_return T_PICKUP_COLUMN_HEADER

  LET lo_exp_header.pch_PICK_OWNER    = FGL_GETENV("USER")
  LET lo_exp_header.pch_PICK_DATETIME = CURRENT YEAR TO SECOND
  LET lo_exp_header.pch_CUST          = FGL_GETENV("CUST")
  LET lo_exp_header.pch_DGENV         = FGL_GETENV("DGENV")
  LET lo_exp_header.pch_HOST_NAME     = FGL_GETENV("HOSTNAME")
  LET lo_exp_header.pch_ZONE          = FGL_GETENV("ZONE")
  
  LET lo_return.* = lo_exp_header.*

  RETURN lo_return.*
  
END FUNCTION 

FUNCTION sadzi140_exim_get_pickup_column_body(p_exp_header,p_pickup_columns)
DEFINE
  p_exp_header      T_PICKUP_COLUMN_HEADER,
  p_pickup_columns  DYNAMIC ARRAY OF T_COLUMN_LIST
DEFINE
  lo_exp_header      T_PICKUP_COLUMN_HEADER,
  lo_pickup_columns  DYNAMIC ARRAY OF T_COLUMN_LIST,
  lo_export_columns  DYNAMIC ARRAY OF T_COLUMN_LIST,
  lo_json_arr        util.JSONArray,
  li_loop            INTEGER
DEFINE
  ls_return  STRING

  LET lo_exp_header.*   = p_exp_header.*
  LET lo_pickup_columns = p_pickup_columns

  LET lo_json_arr = util.JSONArray.create()

  CALL lo_json_arr.put(1,lo_exp_header)
  
  FOR li_loop = 1 TO lo_pickup_columns.getLength()

    LET lo_export_columns[li_loop].COLUMN_NAME = lo_pickup_columns[li_loop].COLUMN_NAME
    LET lo_export_columns[li_loop].COLUMN_DESC = lo_pickup_columns[li_loop].COLUMN_DESC

    #由於已經先放Header, 所以loop都要往後加一
    CALL lo_json_arr.put(li_loop + 1,lo_export_columns[li_loop])
    
  END FOR

  #排版一下JSON格式文件
  LET ls_return = util.JSON.format(lo_json_arr.toString())

  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_exim_set_export_notice_pick_up_info(p_json_obj)
DEFINE
  p_json_obj util.JSONObject

  #給定PICKUP的值
  CALL p_json_obj.put(cs_keyword_PICKUP,"Y")

END FUNCTION 

FUNCTION sadzi140_exim_get_pickup_column_contents(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_table_name STRING,
  ls_file_name  STRING,
  lb_result     BOOLEAN,
  lo_json_arr   util.JSONArray

  LET ls_table_name = p_table_name

  LET ls_file_name = ls_table_name,"+PICKUP_COLS.json"
  CALL sadzi140_exim_get_json_array(ls_file_name) RETURNING lb_result,lo_json_arr

  RETURN lb_result,lo_json_arr

END FUNCTION

FUNCTION sadzi140_exim_check_if_exist_in_pick_list(p_col_arr,p_check_column)
DEFINE
  p_col_arr       DYNAMIC ARRAY OF T_COLUMN_LIST,
  p_check_column  STRING
DEFINE
  lo_col_arr       DYNAMIC ARRAY OF T_COLUMN_LIST,
  ls_check_column  STRING,
  li_loop          INTEGER,
  lb_result        BOOLEAN
DEFINE 
  lb_return  BOOLEAN  

  LET lo_col_arr      = p_col_arr
  LET ls_check_column = p_check_column

  LET lb_result = FALSE 
  FOR li_loop = 1 TO lo_col_arr.getLength()
    IF lo_col_arr[li_loop].COLUMN_NAME = ls_check_column THEN
      LET lb_result = TRUE
      EXIT FOR 
    END IF 
  END FOR

  LET lb_return = lb_result

  RETURN lb_return

END FUNCTION

FUNCTION sadzi140_exim_transfer_json_arr_to_FGL(p_json_arr)
DEFINE
  p_json_arr   util.JSONArray
DEFINE
  lo_json_arr  util.JSONArray,
  lo_json_obj  util.JSONObject,
  lo_col_arr   DYNAMIC ARRAY OF T_COLUMN_LIST,
  lo_real_arr  DYNAMIC ARRAY OF T_COLUMN_LIST,
  li_loop      INTEGER
DEFINE
  lo_return  DYNAMIC ARRAY OF T_COLUMN_LIST
  
  LET lo_json_arr = p_json_arr

  CALL lo_col_arr.clear()

  FOR li_loop = 1 TO lo_json_arr.getLength()
    CALL lo_json_arr.get(li_loop) RETURNING lo_json_obj
    CALL lo_json_obj.toFGL(lo_col_arr[lo_col_arr.getLength()+1])
  END FOR

  FOR li_loop = 1 TO lo_col_arr.getLength()
    IF lo_col_arr[li_loop].COLUMN_NAME IS NOT NULL THEN
      LET lo_real_arr[lo_real_arr.getLength()+1].* = lo_col_arr[li_loop].*
    END IF
  END FOR

  LET lo_return = lo_real_arr

  RETURN lo_return
  
END FUNCTION

FUNCTION sadzi140_exim_check_and_get_pickup_cols_json(p_table_name,p_json_obj)
DEFINE 
  p_table_name  STRING,
  p_json_obj    util.JSONObject
DEFINE
  ls_table_name STRING,
  ls_pickup     STRING,
  lb_result     BOOLEAN,
  lo_json_arr   util.JSONArray
DEFINE
  lb_return  BOOLEAN,
  lo_return  util.JSONArray
  
  LET ls_table_name = p_table_name

  LET lb_return = FALSE
  LET lo_return = NULL
  
  LET ls_pickup = p_json_obj.Get(cs_keyword_PICKUP)
  IF NVL(ls_pickup,"N") = "Y" THEN
    CALL sadzi140_exim_get_pickup_column_contents(ls_table_name) RETURNING lb_result,lo_json_arr
    #CALL lo_json_obj.remove(1)
    LET lb_return = lb_result
    LET lo_return = lo_json_arr 
  END IF  

  RETURN lb_return,lo_return

END FUNCTION

#20171120 end

#20180302 begin
FUNCTION sadzi140_exim_check_if_exist_in_deleted_std_col_list(p_col_arr,p_check_column)
DEFINE
  p_col_arr       DYNAMIC ARRAY OF T_COLUMN_LIST,
  p_check_column  STRING
DEFINE
  lo_col_arr       DYNAMIC ARRAY OF T_COLUMN_LIST,
  ls_check_column  STRING,
  li_loop          INTEGER,
  lb_result        BOOLEAN
DEFINE 
  lb_return  BOOLEAN  

  LET lo_col_arr      = p_col_arr
  LET ls_check_column = p_check_column

  LET lb_result = FALSE 
  FOR li_loop = 1 TO lo_col_arr.getLength()
    IF lo_col_arr[li_loop].COLUMN_NAME = ls_check_column THEN
      LET lb_result = TRUE
      EXIT FOR 
    END IF 
  END FOR

  LET lb_return = lb_result

  RETURN lb_return

END FUNCTION
#20180302 end

#20180410 begin
FUNCTION sadzi140_exim_get_current_construct_version(p_table_name,p_dgenv)
DEFINE
  p_table_name  STRING,
  p_dgenv       STRING  
DEFINE
  ls_table_name   STRING,
  ls_dgenv        STRING,  
  lo_curr_DZAF_T  T_DZAF_T
DEFINE
  lo_return   T_DZAF_T 
  
  LET ls_table_name = p_table_name
  LET ls_dgenv      = p_dgenv
  
  #取得現今DZAF的資料
  LET lo_curr_DZAF_T.DZAF001 = ls_table_name
  LET lo_curr_DZAF_T.DZAF005 = cs_spec_type_table
  LET lo_curr_DZAF_T.DZAF010 = ls_dgenv
  CALL sadzp200_ver_get_curr_ver_info(lo_curr_DZAF_T.*) RETURNING lo_curr_DZAF_T.*

  LET lo_return.* = lo_curr_DZAF_T.*

  RETURN lo_return.*
  
END FUNCTION

FUNCTION sadzi140_exim_set_verify_construct_version(p_json_obj,p_version)
DEFINE
  p_json_obj  util.JSONObject,
  p_version   STRING

  CALL p_json_obj.put(cs_keyword_CONSTRUCT_VER,p_version)
  
END FUNCTION
#20180410 end
