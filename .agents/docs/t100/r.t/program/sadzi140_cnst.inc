{
  異動型態    異動單號       日 期       異動者      異動內容
  ========= ============= ========== ========== ===============================================
  Modify                  20180129   02160      1. 匯入時判斷, 若匯出檔的 ERPVER 和目標區的不一樣則報錯
  Modify                  20180307   02160      1. 增加Patch Mode, 以讓 adzi140_patch 順利執行
  Modify                  20181224   02160      1.修正比對欄位型態或長度不同時所給予的背景訊息
  Modify                  20190819   02160      1.當PK設定超過 10 個欄位的時候就加以控卡
  Modify                  20191205   02160      1.在客戶端, 無論是r.t或是後台新增的標準欄位, 一律不去刪除, 只做新增與修改
  ------------------------------------------------------------------------------
  Modify                  20191209   02160      1.修改支援 MSSQLServer(1.1.071)   
  Modify                  20200330   02160      1.修改支援 PosgreSQL(1.1.072)   
  Modify                  20211122   02160      1.修改欄位轉為PK時內容更新預設值
  Modify                  20230804   02160      1.修改支援 DM
  Modify                  20240521   02160      1.修改支援 KINGBASE
  Modify                  20241021   02160      1.修改 pk limit 為 11
}
  
#Primary Key Constant
CONSTANT cs_db_type           STRING = "ORA" 
CONSTANT cs_enterprise        STRING = "99"

CONSTANT cs_primary_key_type  STRING = "P"
CONSTANT cs_unique_index_type STRING = "U"

CONSTANT cs_pk_keyword        STRING = "_pk"
CONSTANT cs_fk_keyword        STRING = "_fk"
CONSTANT cs_fd_keyword        STRING = "_fd"
CONSTANT cs_rk_keyword        STRING = "_rk"
CONSTANT cs_table_tail_code   STRING = "_t"

CONSTANT cs_idx_normal_keyword STRING = "_n"
CONSTANT cs_idx_unique_keyword STRING = "_u"

CONSTANT cs_null_value        STRING = "!@#"      

CONSTANT cs_top_module_name   STRING = "TOP"
CONSTANT cs_table_designer    STRING = "TableDesigner"
CONSTANT cs_database          STRING = "Database"

CONSTANT cs_alter_column      STRING = "altcol"
CONSTANT cs_drop_column       STRING = "drpcol"
CONSTANT cs_create_table      STRING = "crttbl"
CONSTANT cs_create_synonym    STRING = "crtsym"
CONSTANT cs_drop_synonym      STRING = "drpsym" #20200330
CONSTANT cs_backup_view       STRING = "bakview" #20200330
CONSTANT cs_alter_table       STRING = "alttable"
CONSTANT cs_toggle_bookmark   STRING = "tglebmark" 
CONSTANT cs_grant_table       STRING = "granttbl" #20240606

CONSTANT cs_reversion         STRING = "rev"
CONSTANT cs_digits            STRING = "&&&"

CONSTANT cs_master_db         STRING = "ds"
CONSTANT cs_master_user       STRING = "ds"
CONSTANT cs_slave_user        STRING = "dsdemo"

CONSTANT cs_t10dit_db         STRING = "t10dit"
CONSTANT cs_local_db          STRING = "local"
CONSTANT cs_local_msv         STRING = "local_msv" #20191209

CONSTANT cs_test_msv          STRING = "test_msv" #20191209
CONSTANT cs_test_pgs          STRING = "pgs" #20200330
CONSTANT cs_test_kdb          STRING = "kdb" #20240521

CONSTANT cs_dummy_table_name  STRING = 'DUMMY'
CONSTANT cs_dummy_column_name STRING = 'DUMMY'

CONSTANT cs_default_privilege_role STRING = "TIPTOP"

CONSTANT cs_question_yes      STRING = "yes"    
CONSTANT cs_question_no       STRING = "no" 

CONSTANT cs_result_success    STRING = "Success"    
CONSTANT cs_result_fail       STRING = "Fail" 
CONSTANT cs_result_cancel     STRING = "Cancel"    

CONSTANT cs_init_version_code STRING = '0'
CONSTANT cs_init_alm_ver_code STRING = '0'
CONSTANT cs_first_version_code STRING = '1' #20180410

CONSTANT cs_no_define          STRING = 'NoDefine' 
CONSTANT cs_group_leading_code STRING = "grp" 

CONSTANT cs_cdf_serial_number STRING = "cdfSerialNumber"  
CONSTANT cs_cdf_user_define   STRING = "cdfUserDefine"
CONSTANT cs_cdf_enterprise    STRING = "cdfEnterprise" #20180903

CONSTANT cs_cdf_user_define_varchar  STRING = "cdfUserDefineVarchar"  
CONSTANT cs_cdf_user_define_number   STRING = "cdfUserDefineNumber"  
CONSTANT cs_cdf_user_define_datetime STRING = "cdfUserDefineDateTime"  

CONSTANT cs_default_lang      STRING = "zh_TW"
CONSTANT cs_dept              STRING = "None"

CONSTANT cs_lang_zh_cn        STRING = "zh_CN"
CONSTANT cs_lang_zh_tw        STRING = "zh_TW"

CONSTANT cs_order_by_asc      STRING = "ASC"
CONSTANT cs_order_by_desc     STRING = "DESC"

CONSTANT cs_message_hint      STRING = "Hint"
CONSTANT cs_message_warning   STRING = "Warning"
CONSTANT cs_message_error     STRING = "Error"

CONSTANT cs_rebuild_tail      STRING = "_rebuild"
CONSTANT cs_export_import_ext STRING = 'tdi'

CONSTANT cs_export_import_all_file STRING = '*.tdi'

CONSTANT cs_table_rev         STRING = "Rev"
CONSTANT cs_column_shipped    STRING = "Y"
CONSTANT cs_lang_modified     STRING = "Y"

CONSTANT cs_view_type_db      STRING = "DB"
CONSTANT cs_view_type_rt      STRING = "RT"

CONSTANT cs_alter_type_add    STRING = "add"
CONSTANT cs_alter_type_modify STRING = "modify"
CONSTANT cs_alter_type_alter  STRING = "alter" 
CONSTANT cs_alter_type_alter_column STRING = "alter column" #20191209

CONSTANT cs_alm_table         STRING = "DZLM_T"
CONSTANT cs_mdm_module_name   STRING = "MDM" 

CONSTANT cs_schema_type_all  STRING = "ALL"

CONSTANT cs_pass_through_status_code STRING = "P"

CONSTANT cs_default_db_type  STRING = "ORA"
CONSTANT cs_default_topind   STRING = "sd"
CONSTANT cs_standard_topind  STRING = "sd"

CONSTANT cs_status_finalizing STRING = "Finalizing..."

CONSTANT cs_privilege_receiver_dsaps STRING = "dsaps"

CONSTANT cs_keyword_DGENV         STRING = "DGENV"
CONSTANT cs_keyword_ERPID         STRING = "ERPID"
CONSTANT cs_keyword_CUST          STRING = "CUST"
CONSTANT cs_keyword_DB_TYPE       STRING = "DB_TYPE"
CONSTANT cs_keyword_TOPIND        STRING = "TOPIND"
CONSTANT cs_keyword_renamed       STRING = "RENAMED"
CONSTANT cs_keyword_PICKUP        STRING = "PICKUP" #20171120
CONSTANT cs_keyword_ERPVER        STRING = "ERPVER" #20180129
CONSTANT cs_keyword_CONSTRUCT_VER STRING = "CONSTRUCT_VER" #20180410

CONSTANT cs_ERP_default_ID     STRING = "T100ERP"
CONSTANT cs_ERP_default_ERPVER STRING = "1.0" #20180129

#Grant Privilege
CONSTANT cs_privilege_all        STRING = "ALL"
CONSTANT cs_privilege_select     STRING = "SELECT"
CONSTANT cs_privilege_insert     STRING = "INSERT"
CONSTANT cs_privilege_update     STRING = "UPDATE"
CONSTANT cs_privilege_delete     STRING = "DELETE"
CONSTANT cs_privilege_references STRING = "REFERENCES"
CONSTANT cs_privilege_alter      STRING = "ALTER"
CONSTANT cs_privilege_index      STRING = "INDEX"

CONSTANT cs_custom_module_leading_code STRING = "E"

#Parameter Level define
CONSTANT cs_param_level_system          STRING = "System"
CONSTANT cs_param_level_editor          STRING = "Editor"
CONSTANT cs_param_level_alter_table     STRING = "AlterTable"
CONSTANT cs_param_level_rebuild         STRING = "Rebuild"
CONSTANT cs_param_level_alm             STRING = "ALM"
CONSTANT cs_param_level_column_define   STRING = "ColumnDefine"
CONSTANT cs_param_level_lob             STRING = "LOB"
CONSTANT cs_param_level_language        STRING = "Language"
CONSTANT cs_param_level_default_value   STRING = "DefaultValue"
CONSTANT cs_param_level_patch_Key_value STRING = "PatchKeyValue"
CONSTANT cs_param_level_session         STRING = "Session"
CONSTANT cs_param_level_export_import   STRING = "ExportImport"

#System Level
CONSTANT cs_param_master_db         STRING = "MasterDB"
CONSTANT cs_param_master_user       STRING = "MasterUser"
CONSTANT cs_param_enable_customize  STRING = "EnableCustomize"
CONSTANT cs_param_enable_industry   STRING = "EnableIndustry"
CONSTANT cs_param_essential_schemas STRING = "EssentialSchemas"

#Editor Level
CONSTANT cs_param_enable_user_define            STRING = "EnableUserDefine"
CONSTANT cs_param_version                       STRING = "Version"
CONSTANT cs_param_show_detail_color             STRING = "ShowDetailColor"
CONSTANT cs_param_drag_and_drop_internal        STRING = "DragAndDropInternal"
CONSTANT cs_param_delete_table_control          STRING = "DeleteTableControl"
CONSTANT cs_param_synonym_define                STRING = "SynonymDefine" 
CONSTANT cs_param_check_table_tail_code         STRING = "CheckTableTailCode" 
CONSTANT cs_param_table_user_define_code        STRING = "TableUserDefineCode" 
CONSTANT cs_param_column_user_define_code       STRING = "ColumnUserDefineCode" 
CONSTANT cs_param_table_tail_code               STRING = "TableTailCode"
CONSTANT cs_param_enable_foreign_key            STRING = "EnableForeignKey"
CONSTANT cs_param_table_user_custom_code        STRING = "TableUserCustomCode" 
CONSTANT cs_param_column_user_custom_code       STRING = "ColumnUserCustomCode" 
CONSTANT cs_param_column_user_alter_code        STRING = "ColumnUserAlterCode" 
CONSTANT cs_param_key_user_alter_leading_code   STRING = "KeyUserAlterLeadingCode" 
CONSTANT cs_param_index_user_alter_leading_code STRING = "IndexUserAlterLeadingCode" 
CONSTANT cs_param_max_table_name_length         STRING = "MaxTableNameLength" 
CONSTANT cs_param_notify_custom_field_type      STRING = "NotifyCustomFieldType"            

#AlterTable Level
CONSTANT cs_param_drop_abnormal_column   STRING = "DropAbnormalColumn"

#ALM Level
CONSTANT cs_param_enable_alm             STRING = "EnableALM"
CONSTANT cs_param_check_pass_request     STRING = "CheckPassRequest"
CONSTANT cs_param_enable_backup_trunc    STRING = "EnableBackupTrunc"
CONSTANT cs_param_update_shipping_code   STRING = "UpdateShippingCode"

#LOB Level
CONSTANT cs_param_tablespace STRING = "Tablespace"

#Rebuild Level
CONSTANT cs_param_enable_data_moving     STRING = "EnableDataMoving"
CONSTANT cs_param_drop_backup_table      STRING = "DropBackupTable"
CONSTANT cs_param_check_data_exists      STRING = "CheckDataExists"
CONSTANT cs_param_max_rebuild_records    STRING = "MaxRebuildRecords"
CONSTANT cs_param_enable_force_rebuild   STRING = "EnableForceRebuild"

#Column Define Level
CONSTANT cs_param_number_of_varchar_fields  STRING = "NumberOfVarcharFields"
CONSTANT cs_param_number_of_number_fields   STRING = "NumberOfNumberFields"
CONSTANT cs_param_number_of_datetime_fields STRING = "NumberOfDateTimeFields"

#Language level
CONSTANT cs_param_local_language  STRING = "LocalLanguage"

#DefaultValue Level
CONSTANT cs_param_enable_blob       STRING = "EnableBLOB"
CONSTANT cs_param_enable_clob       STRING = "EnableCLOB"
CONSTANT cs_param_enable_date       STRING = "EnableDATE"
CONSTANT cs_param_enable_number     STRING = "EnableNUMBER"
CONSTANT cs_param_enable_numeric    STRING = "EnableNUMERIC" #20240529
CONSTANT cs_param_enable_timestamp  STRING = "EnableTIMESTAMP"
CONSTANT cs_param_enable_varchar2   STRING = "EnableVARCHAR2"

CONSTANT cs_param_enable_varbinary  STRING = "EnableVARBINARY"
CONSTANT cs_param_enable_decimal    STRING = "EnableDECIMAL"
CONSTANT cs_param_enable_datetime2  STRING = "EnableDATETIME2"
CONSTANT cs_param_enable_nvarchar   STRING = "EnableNVARCHAR"

#20200330 add begin
CONSTANT cs_param_enable_text       STRING = "EnableTEXT"
CONSTANT cs_param_enable_int2       STRING = "EnableINT2"
CONSTANT cs_param_enable_int4       STRING = "EnableINT4"
CONSTANT cs_param_enable_int8       STRING = "EnableINT8"
CONSTANT cs_param_enable_varchar    STRING = "EnableVARCHAR"
CONSTANT cs_param_enable_bytea      STRING = "EnableBYTEA"
#20200330 add end
#20230804 add begin
CONSTANT cs_param_enable_smallint      STRING = "EnableSMALLINT"
CONSTANT cs_param_enable_integer       STRING = "EnableINTEGER"
CONSTANT cs_param_enable_bigint        STRING = "EnableBIGINT"
#20230804 add end
#PatchKeyValue Level
CONSTANT cs_param_patch_key_value_date      STRING = "DATE"
CONSTANT cs_param_patch_key_value_number    STRING = "NUMBER"
CONSTANT cs_param_patch_key_value_numeric   STRING = "NUMERIC" #20240529
CONSTANT cs_param_patch_key_value_timestamp STRING = "TIMESTAMP"
CONSTANT cs_param_patch_key_value_varchar2  STRING = "VARCHAR2"

#ExportImport Level
CONSTANT cs_param_check_erpid  STRING = "CheckERPID"

#Optimizer Features Enable
CONSTANT cs_param_optimizer_features_enable  STRING = "OptimizerFeaturesEnable"

CONSTANT cs_side_left  STRING = "LEFT"
CONSTANT cs_side_right STRING = "RIGHT"

CONSTANT cs_editor_mode_browse STRING = "BROWSE"
CONSTANT cs_editor_mode_edit   STRING = "EDIT"
CONSTANT cs_editor_mode_design STRING = "DESIGN"

CONSTANT cs_exim_dump     STRING = "idmp" 
CONSTANT cs_exim_export   STRING = "iexp" 
CONSTANT cs_exim_import   STRING = "iimp" 
CONSTANT cs_exim_verify   STRING = "ivfy" 
CONSTANT cs_exim_assemble STRING = "iasm" 
CONSTANT cs_exim_delete   STRING = "idel" 
CONSTANT cs_exim_drop     STRING = "idrp"
CONSTANT cs_exim_alter    STRING = "ialt" 
CONSTANT cs_exim_rebuild  STRING = "ireb" 
CONSTANT cs_exim_sel_imp  STRING = "isimp" #20171120 for selection import

CONSTANT cs_working_dir_type_import STRING = "IMP"
CONSTANT cs_working_dir_type_export STRING = "EXP"

CONSTANT cs_rt_design_data_version_code STRING = "0"
CONSTANT cs_rt_design_data_code         STRING = "T"
CONSTANT cs_rt_design_data_number       STRING = "9999999999" #20180420
CONSTANT cs_rt_design_data_version      STRING = "99999"
CONSTANT cs_rt_design_sd_version        STRING = "0"

CONSTANT cs_rt_design_request_no  STRING = "999999-99999"
CONSTANT cs_rt_design_sequence_no STRING = "99999"

CONSTANT cs_null_username_default STRING = "Tester"
CONSTANT cs_max_table_name_length STRING = "4" 

CONSTANT cs_default_optimizer_features_enable STRING = "11.2.0.3"

#Stored procedure name
CONSTANT cs_prc_get_col_default          STRING = "GET_COL_DEFAULT"
CONSTANT cs_prc_get_col_default_by_owner STRING = "GET_COL_DEFAULT_BY_OWNER"

#Backup, Recover Type
CONSTANT cs_bkrv_type_qc_pass   STRING = "PASS"
CONSTANT cs_bkrv_type_patch     STRING = "PATCH"
CONSTANT cs_bkrv_type_check_out STRING = "CHECK_OUT"

CONSTANT cs_bkrv_backup    STRING = "backup"
CONSTANT cs_bkrv_recovery  STRING = "recovery"

CONSTANT cs_env_RT_DEBUG_MODE   STRING = "RT_DEBUG_MODE" 
CONSTANT cs_env_RT_PATCH_MODE   STRING = "RT_PATCH_MODE" #20180307
CONSTANT cs_env_T100PASSERPVER  STRING = "T100PASSERPVER" #20180420
CONSTANT cs_env_cust_digiwin    STRING = "DIGIWIN"
CONSTANT cs_env_RT_DROP_BACKDOOR_COLUMN STRING = "RT_DROP_BACKDOOR_COLUMN"
CONSTANT cs_env_RT_DEL_EXTRA_STD_COLS   STRING = "RT_DEL_EXTRA_STD_COLS" #20191205
CONSTANT cs_env_full_locale     STRING = "FULL_LOCALE"  #20230821

#Privileges
CONSTANT cs_priv_type_select     STRING = 'SELECT'
CONSTANT cs_priv_type_insert     STRING = 'INSERT'
CONSTANT cs_priv_type_update     STRING = 'UPDATE'
CONSTANT cs_priv_type_delete     STRING = 'DELETE'
CONSTANT cs_priv_type_alter      STRING = 'ALTER'
CONSTANT cs_priv_type_index      STRING = 'INDEX'
CONSTANT cs_priv_type_references STRING = 'REFERENCES'

#Logs
CONSTANT cs_logs_level_information STRING = "Information"
CONSTANT cs_logs_level_warning     STRING = "Warning"
CONSTANT cs_logs_level_error       STRING = "Error"

CONSTANT cs_logs_type_table    STRING = "Table"
CONSTANT cs_logs_type_column   STRING = "Column"
CONSTANT cs_logs_type_key      STRING = "Key"
CONSTANT cs_logs_type_index    STRING = "Index"
CONSTANT cs_logs_type_synonym  STRING = "Synonym"
CONSTANT cs_logs_type_view     STRING = "View"
CONSTANT cs_logs_type_trigger  STRING = "Trigger"

CONSTANT cs_logs_type_reserve_column STRING = "ReserveColumn"
CONSTANT cs_logs_type_drop_backdoor_column STRING = "DropBackdoorColumn"
CONSTANT cs_logs_type_set_drop_backdoor_column STRING = "SetDropBackDoorCols"

#20170912 begin
CONSTANT cs_resort_index_type_text  STRING = "TEXT"
CONSTANT cs_resort_index_type_table STRING = "TABLE"
#20170912 end

#20181224 begin
CONSTANT cs_valid_err_type_err   STRING = "TYPE_ERR"
CONSTANT cs_valid_err_length_err STRING = "LENGTH_ERR"
#20181224 end

#20190304 begin
CONSTANT cs_col_type_null STRING = "COL_TYPE:null"
#20190304 end

#20190819 begin
CONSTANT cs_disable_pk_cols_limits STRING = "RT_DISABLE_PK_COLS_LIMITS"
CONSTANT ci_pk_cols_limits         INTEGER = 11 -- T100 defined --20241021
CONSTANT ci_max_pk_cols_limits     INTEGER = 32 -- Oracle defined
#20190819 end

#20191209 begin
&ifdef is_msv
CONSTANT cs_exit_sign       STRING = "exit"
CONSTANT cs_get_date        STRING = "getdate()"
CONSTANT cs_get_time_stamp  STRING = "sysdatetime()"
CONSTANT cs_get_time_stamp0 STRING = "sysdatetime()" -- 20230804
CONSTANT cs_get_time_stamp5 STRING = "sysdatetime()" -- 20230804
&else
&ifdef is_pgs
CONSTANT cs_exit_sign      STRING = ""
&else
CONSTANT cs_exit_sign       STRING = "exit;"
&endif
CONSTANT cs_get_date        STRING = "current_date" -- 20211122
CONSTANT cs_get_time_stamp  STRING = "current_timestamp" -- 20211122
CONSTANT cs_get_time_stamp0 STRING = "current_timestamp(0)" -- 20230804
CONSTANT cs_get_time_stamp5 STRING = "current_timestamp(5)" -- 20230804
&endif

CONSTANT cs_os_type_win  STRING = "win"
CONSTANT cs_os_type_unix STRING = "unix"
#20191209 end

CONSTANT cs_pgs_alter_using STRING = "using"
CONSTANT cs_pgs_type_cast   STRING = "::"

CONSTANT cs_pgs_sub_pgs     STRING = "POSTGRESQL"
CONSTANT cs_pgs_sub_mogdb   STRING = "MOGDB"
CONSTANT cs_pgs_sub_polardb STRING = "POLARDB"
CONSTANT cs_pgs_sub_kdb     STRING = "KINGBASE" --20240521

CONSTANT cs_pgs_mog_home    STRING = "GAUSSHOME"
CONSTANT cs_pgs_kdb_home    STRING = "KINGBASE_HOME" --20240521

CONSTANT cs_db_driver_type_ifx STRING = "ifx"
CONSTANT cs_db_driver_type_ora STRING = "ora"
CONSTANT cs_db_driver_type_msv STRING = "msv"
CONSTANT cs_db_driver_type_snc STRING = "snc"
CONSTANT cs_db_driver_type_pgs STRING = "pgs" 
CONSTANT cs_db_driver_type_dmg STRING = "dmg" --20230804 
CONSTANT cs_db_driver_type_kdb STRING = "kdb" --20240521 

CONSTANT cs_indsd_ux STRING = "ux" #20201007
