#+ Version....: T6-ERP-1.00.00 Build-000005
#+
#+ Filename...: adzi888全域變數
#+ Buildtype..:
#+ Memo.......: 2017/09/13 - 170921-00005
#               2017/12/12 - 多人打包
#               2018/03/29 - 180329-00037 by 02785 :設計器匯出入模式:支持產中環境下跨$ERPVER可匯入  
#               2018/04/19 - 增加info000
#               2018/11/29 - 增加g_skip_regen TREU-不做重產/編譯連結, FALSE-要做
#               2019/01/11 - 增加g_patch_info.info015存放匯出環境$DGNEV變數for跨$ERPVER卡控
#190702-00015   2019/04/16 by 02785 - 支持v3.0(產中開發環境為多版本共存區,產中打包時42m檔1.0之外還有3.0版本,過單支持,客戶解包支持)
#190702-00013   2019/06/17 by 02785 - 增加g_patch_info.info016存放此patch包是否包含awsi011資料-Y/N
#190621-00012   2019/06/20 by 02785 - 增加CONSTANT for azz source不出貨
#190909-00042   2019/09/17 by 02785 - 增加g_oracle_sid,g_booking_azzp660
#191108-00019   2019/11/08 by 02785 - 增加g_patch_info.ind_pack RECORD存放行業包打包當下整體標準平台的基準線相關資訊
#191108-00020   2019/11/11 by 02785 - g_patch_info.prog[].prog000 LIKE type_t.chr20 改成LIKE dzaf_t.dzaf001
#191217-00020   2019/12/18 by 02785 - 增加g_booking_azzp999
#200113-00023   2020/04/06 by 02785 - 1.增加g_syn_allent_list 需要複製ent資料的作業
#                                     2.增加g_sys_schema2 放dsdemo
#                                     3.增加g_exp_para2放dsdemo密碼
#200526-00022   2020/05/26 by 02785 - 增加g_run_adzi888_with_god_mode
#PGS-02785                 by 02785 - open db
#                                     1.patch.info新增 info017 (記錄匯出端$TOPDB 資料庫別(ORA/PGS...) 
#                                     2.新增g_topdb (from $TOPDB)
#                                     3.新增gc_schema_dspatch (常數 dspatch)
#                                     4.新增g_alm_jb (是否與ALM綁定)
#                                     5.新增g_hostname (環境變數$HOSTNAME)
#200812-00059   2020/08/12 by 02785 - 過單/adzi888/設計器匯入支持load/unload模式:
#                                     1.新增T100SYNCO2P
#                                     2.新增patch_info.info018記錄匯出資料的方式
#200420-00045   2020/12/10 by 02785 - 1.增加gb_not_routine_patch(小包patch註記)
#210203-00022   2021/02/03 by 02785 - 1.新增patch_info.info019記錄此包有無4rp檔案
#                                     2.新增patch_info.info020記錄此包有無fr檔案
#                                     3.新增patch_info.module記錄此包程式的模組清單
#210511-00010   2021/05/11 by 02785   1.新增ga_t100patch_multilang :patch打包資料語系清單
#210715-00006   2021/07/15 by 02785   1.新增g_t100skipunpackfr(for $T100SKIPUNPACKFR)
#211104-00016   2022/04/15 by 02785   1.新增g_t100skipunpackabi(for $T100SKIPUNPACKABI)
#                                     2.新增patch_info.info021記錄此包有無abi檔案
#220105-00033   2022/01/05 by 02785   1.指定追單範圍更新:新增修改筆記tips
#220915-00030   2022/09/15 by 02431 - 二開優化議題#2-5 adzi888工具一體化
#230309-00032   2023/03/09 by 02785   1.新增patch_info.info022記錄此包有無表格包
#230411-00019   2023/04/11 by 02785   1.新增patch_info.info023記錄此包有無XR報表元件
#                                     2.新增CONSTANT gc_xr_design_data_tables定義XR報表元件設計資料表格清單(json format)
#230811-00002   2023/10/01 by 02785 - open db DM
#                                     1.新增g_database_type (from 4js db_get_database_type():db driver type)
#                                     2.新增patch_info.info024 (記錄匯出端db driver type)
#                                     3.新增patch_info.info025 (記錄匯出端g_db_branch)
#240603-00034   2024/06/12 by 02785   1.新增環境變數$T100SYNCO2P_NOTRANS (used unload/load但不強制轉型)
#231206-00011   2024/06/28 by 02785 - open db:patch支持ora環境打包dm環境解包(sqLite with json format)
#                                     1.新增常數gc_sqlite_file,gc_sqlite_conn_str
#                                     2.新增g_t100sqlite:資料以sqlite處理(for $T100SQLITE)
#                                     3.新增patch_info.info026 (記錄此包有無sqlite檔)
#                                     4.新增lr_patch4json_para record 
#240813-00016   2024/08/13 by 02785   1.新增環境變數$T100SYNCO2P_REVERSE (used unload/load,date強制轉型,逆向補時分秒)
#241114-00004   2025/09/25 by 02785   1.新增patch_info.info027 (記錄此包fglrun版本)
 
SCHEMA ds

GLOBALS
   CONSTANT g_dgenv_s = "s"
   CONSTANT g_dgenv_c = "c"
   CONSTANT g_pack_dir_env = "TEMPDIR"          #匯出檔打包工作目錄
   CONSTANT g_gen42s_arg2 = "tiptop"            #第二參數傳入tiptop就不會做關聯42s檔
   CONSTANT g_table_list = "table_list.xml"     #有patch data的table代碼
   CONSTANT g_sys_schema = "ds"                 #系統參照的DB schema
   CONSTANT g_sys_schema2 = "dsdemo"            #系統參照的DB schema:dsdemo #2020/04/06
   #CONSTANT g_patch_schema = "dspatch"         #patch參照的DB schema #2017/12/12
   DEFINE g_patch_schema STRING                 #打包使用的DB         #2017/12/12
   CONSTANT g_gzzq_standard = "gzzq_standard.unl"         #gzzq_t內standard屬於程式代碼內的所有ACTION多語言對照表
   CONSTANT g_patch_file_path = "/u6/scm"       #patch時tgz預設產生路徑
   CONSTANT g_patch_standard_source_path ="standard_source" #patch時標準程式source目錄  #160223-00028 add

   #####patch task流程順序定義
   CONSTANT gc_empty_dspatch = 1
   CONSTANT gc_imp_file = 2
   CONSTANT gc_imp_base_data = 3
   CONSTANT gc_imp_dmp = 4
   CONSTANT gc_del_data = 5
   CONSTANT gc_merge_data = 6
   CONSTANT gc_gen_pattern = 7
   CONSTANT gc_gen_base_prog = 8
   CONSTANT gc_gen_06_prog_init = 9
   CONSTANT gc_gen_06_prog_merge = 10   #160223-00028 add
   CONSTANT gc_gen_06_prog_gen = 11     #160223-00028 modify
   CONSTANT gc_gen_06_sub_compile = 12  #160223-00028 modify
   CONSTANT gc_gen_06_sub_link = 13     #160223-00028 modify
   CONSTANT gc_gen_06_prog_compile = 14 #160223-00028 modify
   CONSTANT gc_gen_06_prog_link = 15    #160223-00028 modify
   CONSTANT gc_gen_final_prog = 16      #160223-00028 modify
   CONSTANT gc_gen_final_patch = 17     #160223-00028 modify

   CONSTANT g_all_main_tasks = 17       #160223-00028 modify
   #########################

   CONSTANT gc_del_azz_source_shell_1 = "update_dzbx_t.sh" #將刪除清單當下版次記錄下來的腳本 #2019/06/20
   CONSTANT gc_del_azz_source_shell_2 = "DelAzzProg.sh"    #刪除azz source的腳本 #2019/06/20

   CONSTANT gc_schema_dspatch = "dspatch"  #dspatch #PGS-02785

   CONSTANT cs_tips_begin   STRING = "#Begin: patch update\n" #2022/01/05
   CONSTANT cs_tips_end     STRING = "#End: patch update\n"   #2022/01/05

   #json format "表格代號","程式代號欄位"#2023/04/11 j
   CONSTANT gc_xr_design_data_tables = '{ "dzka_t":"dzka001", "dzkc_t":"dzkc001", "dzkd_t":"dzkd001", "dzkdl_t":"dzkdl001", "dzki_t":"dzki001", "dzkj_t":"dzkj001", "dzkjl_t":"dzkjl001", "dzkq_t":"dzkq001", "dzkr_t":"dzkr001", "dzks_t":"dzks001" }' #2023/04/11
   #Begin: 231206-00011
   CONSTANT gc_sqlite_file = "database.sqlite"
   CONSTANT gc_sqlite_conn_str = "sqlite_patch"
   CONSTANT cs_quoto_escape = "&apos;"
   CONSTANT cs_quoto = "'"
   #End: 231206-00011
   
   #單頭 type 宣告
   TYPE type_g_master_m       RECORD
          master001         LIKE dzld_t.dzld001,         #GUID
          master002         LIKE type_t.chr50,           #匯出主題 
          master003         LIKE type_t.chr10,           #匯出人員
          master003_desc    LIKE type_t.chr80, 
          master004         LIKE type_t.chr10,           #自動產生類型(依1.prog:程式代號; 2.date:異動日期)
          master005         LIKE type_t.chr1,            #程式類型(GZCB001=114)
          master006         LIKE type_t.chr50,           #作業代號
          master007         LIKE type_t.dat,             #資料起始時間  #DATETIME YEAR TO SECOND
          master008         LIKE type_t.dat,             #資料結束時間 
          master009         LIKE type_t.dat,             #匯出時間
          master010         LIKE type_t.dat,             #匯入時間
          master011         LIKE type_t.chr1,            #自動產生註冊資料清單否
          master012         LIKE type_t.chr20,           #註冊資訊資料清單建立區域
          master013         LIKE type_t.chr20,           #需求單號
          master014         LIKE type_t.num10             #作業項次
          END RECORD
    
   #單身 type 宣告
   TYPE type_g_ddata_d        RECORD
          #ddata001          LIKE type_t.chr50,          #GUID
          ddata002          LIKE type_t.num10,            #序號
          ddata003          LIKE dzld_t.dzld003,         #作業代號(註冊資訊相關維護作業dzyb_t)
          ddata004          LIKE dzld_t.dzld004,         #匯入動作(1:Merge; 2:Insert; 3:Delete)
          ddata005          LIKE dzld_t.dzld005,         #維護作業
          ddata015          LIKE dzld_t.dzld015,         #過單項目(m:單頭主項; d:單身子項)
          ddata006          LIKE dzld_t.dzld006,         #條件式1
          ddata007          LIKE dzld_t.dzld007,         #條件式2
          ddata008          LIKE dzld_t.dzld008,         #匯出狀態(0:建立; 1:成功; 2:失敗)
          ddata009          LIKE dzld_t.dzld009,         #匯入狀態(0:建立; 1:成功; 2:失敗)
          ddata010          LIKE dzld_t.dzld010          #清單資料產生方式(0:手動建立; 1:依程式代號; 2:依異動日期)
          #ddata011          LIKE type_t.chr20,          #需求單號
          #ddata012          LIKE type_t.num10            #作業項次
          END RECORD
   
   TYPE type_g_dfile_d        RECORD
          #dfile001          LIKE type_t.chr50,          #GUID
          dfile002          LIKE type_t.num10,            #序號
          dfile003          LIKE dzlf_t.dzlf003,         #路徑(exp:/u1/{ZONE}/erp/cfg/4st)
          dfile004          LIKE dzlf_t.dzlf004,         #類型(f:檔案; d:整個目錄結構)
          dfile005          LIKE dzlf_t.dzlf005,         #檔名
          dfile006          LIKE dzlf_t.dzlf006,         #匯出狀態(0:建立; 1:成功; 2:失敗)
          dfile007          LIKE dzlf_t.dzlf007          #匯入狀態(0:建立; 1:成功; 2:失敗)
          END RECORD
   
   TYPE type_g_dtool_d        RECORD
          #dtool001          LIKE type_t.chr50,          #GUID
          dtool002          LIKE type_t.num10,             #序號
          dtool003          LIKE dzlt_t.dzlt003,         #模組(adz; azz; lib; sub; oth)
          dtool004          LIKE dzlt_t.dzlt004,         #匯入類型(4gl:4gl檔案; 4fd:4fd檔案; 42m:42m檔案; cp:cp file; del:delete data; load:load data)
          dtool005          LIKE dzlt_t.dzlt005,         #主程式否
          dtool006          LIKE dzlt_t.dzlt006,         #檔名
          dtool007          LIKE dzlt_t.dzlt007,         #no use
          dtool008          LIKE dzlt_t.dzlt008,         #匯出狀態(0:建立; 1:成功; 2:失敗)
          dtool009          LIKE dzlt_t.dzlt009          #匯入狀態(0:建立; 1:成功; 2:失敗)
          END RECORD
   
  #Begin:220915-00030 
  TYPE type_g_ddzlm_d        RECORD
         ddzlm001            LIKE dzlm_t.dzlm001,  #建構類型
         ddzlm002            LIKE dzlm_t.dzlm002,  #建構代號
         ddzlm003            LIKE dzlm_t.dzlm003,  #建構名稱
         ddzlm004            LIKE dzlm_t.dzlm004,  #模組
         ddzlm005            LIKE dzlm_t.dzlm005,  #建構版號
         ddzlm006            LIKE dzlm_t.dzlm006,  #SD版號
         ddzlm009            LIKE dzlm_t.dzlm009   #PR版號
                                   END RECORD
  #End:220915-00030 
   TYPE type_g_dzyb           RECORD
          dzyb001           LIKE dzyb_t.dzyb001,         #作業名稱
          dzyb002           LIKE dzyb_t.dzyb002,         #資料表名稱
          dzyb003           LIKE dzyb_t.dzyb003,         #主表否
          dzyb004           LIKE dzyb_t.dzyb004,         #主要資料條件欄位
          dzyb005           LIKE dzyb_t.dzyb005,         #創建日欄位名稱
          dzyb006           LIKE dzyb_t.dzyb006,         #修改日欄位名稱
          dzyb007           LIKE dzyb_t.dzyb007,         #匯入動作(1.update/insert(merge), 2.delete/insert, 3.insert(update), 4.insert)
          dzyb008           LIKE dzyb_t.dzyb008,         #次要資料條件欄位
          dzyb009           LIKE dzyb_t.dzyb009,         #客製欄位代號
          dzyc002           LIKE dzyc_t.dzyc002,         #patch動作(1.update/insert(merge), 2.delete/insert, 3.insert(update), 4.insert)
          dzyc003           LIKE dzyc_t.dzyc003,         #客製欄位代號
          dzyc004           LIKE dzyc_t.dzyc004,         #patch條件
          dzyc005           LIKE dzyc_t.dzyc005          #資料多語言欄位代號
          END RECORD
   
   #ALM 註冊資訊清單
   TYPE type_g_dzld_d        RECORD
          dzld001          LIKE dzld_t.dzld001,          #GUID
          dzld002          LIKE type_t.num10,             #序號
          dzld003          LIKE dzld_t.dzld003,          #作業代號(註冊資訊相關維護作業dzyb_t)
          dzld004          LIKE dzld_t.dzld004,          #匯入動作(1:Merge; 2:Insert; 3:Delete)
          dzld005          LIKE dzld_t.dzld005,          #維護作業
          dzld006          LIKE dzld_t.dzld006,          #條件式1
          dzld007          LIKE dzld_t.dzld007,          #條件式2
          dzld008          LIKE dzld_t.dzld008,          #匯出狀態(0:建立; 1:成功; 2:失敗)
          dzld009          LIKE dzld_t.dzld009,          #匯入狀態(0:建立; 1:成功; 2:失敗)
          dzld010          LIKE dzld_t.dzld010,          #清單資料產生方式(0:手動建立; 1:依程式代號; 2:依異動日期)
          dzld011          LIKE dzld_t.dzld011,          #需求單號
          dzld012          LIKE dzld_t.dzld012,          #產品代號
          dzld013          LIKE dzld_t.dzld013,          #產品版本
          dzld014          LIKE dzld_t.dzld014,          #作業項次
          dzld015          LIKE type_t.chr1              #過單項目(M:單頭主項; D:單身子項)
          END RECORD

   #ALM 檔案匯出清單
   TYPE type_g_dzlf_d        RECORD
          dzlf001          LIKE dzlf_t.dzlf001,          #GUID
          dzlf002          LIKE type_t.num10,             #序號
          dzlf003          LIKE dzlf_t.dzlf003,          #路徑
          dzlf004          LIKE dzlf_t.dzlf004,          #類型(f:檔案; d:目錄)
          dzlf005          LIKE dzlf_t.dzlf005,          #檔名
          dzlf006          LIKE dzlf_t.dzlf006,          #匯出狀態(0:建立; 1:成功; 2:失敗)
          dzlf007          LIKE dzlf_t.dzlf007,          #匯入狀態(0:建立; 1:成功; 2:失敗)
          dzlf008          LIKE dzlf_t.dzlf008,          #需求單號
          dzlf009          LIKE dzlf_t.dzlf009,          #產品代號
          dzlf010          LIKE dzlf_t.dzlf010,          #產品版本
          dzlf011          LIKE dzlf_t.dzlf011           #作業項次
          END RECORD

   #IT工具類匯出清單
   TYPE type_g_dzlt_d        RECORD
          dzlt001          LIKE dzlt_t.dzlt001,          #GUID
          dzlt002          LIKE type_t.num10,             #序號
          dzlt003          LIKE dzlt_t.dzlt003,          #模組(adz; azz; lib; sub; oth)
          dzlt004          LIKE dzlt_t.dzlt004,          #匯入類型(4gl:4gl檔案; 4fd:4fd檔案; cp:cp file; del:delete data; load:load data)
          dzlt005          LIKE dzlt_t.dzlt005,          #主程式否
          dzlt006          LIKE dzlt_t.dzlt006,          #檔名
          dzlt007          LIKE dzlt_t.dzlt007,          #no use
          dzlt008          LIKE dzlt_t.dzlt008,          #匯出狀態(0:建立; 1:成功; 2:失敗)
          dzlt009          LIKE dzlt_t.dzlt009,          #匯入狀態(0:建立; 1:成功; 2:失敗)
          dzlt011          LIKE dzlt_t.dzlt011,          #需求單號
          dzlt012          LIKE dzlt_t.dzlt012,          #產品代號
          dzlt013          LIKE dzlt_t.dzlt013,          #產品版本
          dzlt014          LIKE dzlt_t.dzlt014           #作業項次
          END RECORD

  #Begin:220915-00030 
  TYPE type_g_dzlm_d        RECORD
         dzlm001                    LIKE dzlm_t.dzlm001,  #建構類型
         dzlm002                    LIKE dzlm_t.dzlm002,  #建構代號
         dzlm003                    LIKE dzlm_t.dzlm003,  #建構名稱
         dzlm004                    LIKE dzlm_t.dzlm004,  #模組
         dzlm005                    LIKE dzlm_t.dzlm005,  #建構版號
         dzlm006                    LIKE dzlm_t.dzlm006,  #SD版號
         dzlm009                    LIKE dzlm_t.dzlm009,  #PR版號
         dzlm008                    LIKE dzlm_t.dzlm008,  #SD狀態
         dzlm011                    LIKE dzlm_t.dzlm011,  #PR狀態
         dzlm012                    LIKE dzlm_t.dzlm012,  #需求單號
         dzlm013                    LIKE dzlm_t.dzlm013,  #產品代號
         dzlm014                    LIKE dzlm_t.dzlm014,  #產品版本
         dzlm015                    LIKE dzlm_t.dzlm015,  #作業項次
         dzlm022                    LIKE dzlm_t.dzlm022,  #dzlm過單否
         dzlm017                    LIKE dzlm_t.dzlm017   #簽入時間
                                   END RECORD
  #End:220915-00030 



   #ALM 註冊資訊清單
   TYPE type_g_dzld_browser  RECORD
          dzld001          LIKE dzld_t.dzld001,          #GUID
          dzld011          LIKE dzld_t.dzld011,          #需求單號
          dzld012          LIKE dzld_t.dzld012,          #產品代號
          dzld013          LIKE dzld_t.dzld013,          #產品版本
          dzld014          LIKE dzld_t.dzld014           #作業項次
          END RECORD

   #patch 單頭 type 宣告
   TYPE type_g_patch_m       RECORD
          patch001          LIKE dzld_t.dzld001,         #GUID
          patch002          LIKE type_t.chr50,           #patch號
          patch003          LIKE type_t.chr100,          #patch包來源路徑         ###patch包所在路徑
          patch004          LIKE type_t.chr20,           #dspatch user密碼(Oracle DB User)
          patch005          LIKE type_t.chr1,            #只重做程式patch(不做設計資料匯入及合併)
          patch006          LIKE type_t.chr100,          #patch解包檔路徑
          patch007          LIKE type_t.chr1,            #是否gen 42s
          patch008          LIKE type_t.chr1,            #是否merge標準程式
          patch009          LIKE type_t.chr1             #是否為快速更新 160816-00047
          END RECORD
   
   #patch 單頭 type 宣告
   TYPE type_g_load_data     RECORD
          dzea001           LIKE dzea_t.dzea001,         #資料表代碼
          wc                STRING                       #where條件
          END RECORD

   #設計資料執行結果記錄表
   TYPE type_g_dzye_d        RECORD
          dzye001          LIKE dzye_t.dzye001,          #匯入包編號
          dzye002          LIKE dzye_t.dzye002,          #程式代碼
          dzye003          LIKE dzye_t.dzye003,          #建構類型
          dzye004          LIKE dzye_t.dzye004,          #已執行項目
          dzye005          LIKE dzye_t.dzye005,          #模組
          dzye006          LIKE dzye_t.dzye006,          #完成patch否
          dzye007          LIKE dzye_t.dzye007,          #錯誤訊息(簡單)
          dzye008          LIKE dzye_t.dzye008,          #PID
          dzye009          LIKE dzye_t.dzye009,          #目的建構版次
          dzye010          LIKE dzye_t.dzye010,          #目的規格版次
          dzye011          LIKE dzye_t.dzye011,          #目的程式版次
          dzye012          LIKE dzye_t.dzye012,          #目的程式客製否
          dzye013          LIKE dzye_t.dzye013,          #順序
          dzye014          LIKE dzye_t.dzye014,          #no use
          dzye015          LIKE dzye_t.dzye015,          #執行merge否
          dzye016          LIKE dzye_t.dzye016,          #執行階段
          dzye017          LIKE dzye_t.dzye017,          #執行日期
          dzye018          LIKE dzye_t.dzye018,          #錯誤訊息(完整)
          dzye019          LIKE dzye_t.dzye019,          #來源程式客制否
          dzye020          LIKE dzye_t.dzye020,          #來源規格版次
          dzye021          LIKE dzye_t.dzye021,          #來源程式版次
          dzye022          LIKE dzye_t.dzye022,          #目的程式簽出否
          dzye023          LIKE dzye_t.dzye023,          #patch狀態
          dzye024          LIKE dzye_t.dzye024           #執行使用者
          END RECORD

   DEFINE g_alm_jb          LIKE type_t.chr1        #是否與ALM綁定 #PGS-02785備註:此變數一直被拿來存$DGENV,在此就將錯就錯一直延用了
   DEFINE g_show_msg        LIKE type_t.chr1        #是否顯示錯誤訊息視窗
   DEFINE g_dzld012         LIKE dzld_t.dzld012     #產品代號
   DEFINE g_dzld013         LIKE dzld_t.dzld013     #產品版本
   DEFINE g_error_message   STRING                  #錯誤訊息
   DEFINE g_gen42s_flag     LIKE type_t.chr1        #是否自動產生所有的關聯的42s檔
   DEFINE g_batch_flag      LIKE type_t.chr1        #是否整批import data
   DEFINE g_patch_mode      LIKE type_t.chr1        #是否為patch流程
   DEFINE g_patch_pwd       STRING
   DEFINE g_unload_file     LIKE type_t.chr1        #是否在export data過程中UNLOAD TO FILE
   DEFINE g_exp_para        STRING
   DEFINE g_exp_para2       STRING                  #資料庫密碼 20200113
   DEFINE g_dmp_mode        LIKE type_t.chr1        #是否為exp/imp dmp模式
   DEFINE g_run_mode        LIKE type_t.chr1        #adzp999運行模式(1:過單打包; 2:過單解包; 3:patch打包; 4:patch解包; 5:需求單打包; 6:需求單解包)
   DEFINE g_import_dapatch  LIKE type_t.chr1        #是否import dspatch
   DEFINE g_top_alm         LIKE type_t.chr1        #是否串接ALM
   DEFINE g_top_checkout    LIKE type_t.chr1        #是否允許簽出
   DEFINE g_curr_tasks      LIKE type_t.num10        #目前執行階段
   DEFINE g_adzp999_01_s01  LIKE type_t.chr1        #adzp999_01_s01是否有open window
   DEFINE g_patch005        LIKE type_t.chr1        #只重做程式重產(不做設計資料匯入及合併)
   DEFINE g_backfile        LIKE type_t.num10        #重新產出程式時，舊程式備份版數 (0表示不留備份) #160223-00028 add
   DEFINE g_patch_standard_source_path_full STRING  #patch時標準程式source目錄  #new 160223-00028 add
   DEFINE g_quick_patch     LIKE type_t.chr1        #是否為快速patch (Y/N) 預設為N  #160816-00047
  
   DEFINE g_patch_info RECORD
       info000 LIKE type_t.chr50,     #Patch號 #2018/04/19
       info001 LIKE type_t.chr10,     #T100版本(目前固定1.0)
       info002 LIKE type_t.chr10,     #Patch包版本(目前也是1.0)
       info003 LIKE type_t.chr20,     #Oracle版本(exp)
       info004 LIKE type_t.chr20,     #Oracle版本(imp)
       info005 LIKE type_t.chr50,     #匯出環境資訊(FGLASIP)
       info006 LIKE type_t.chr20,     #匯出環境資訊(行業別TOPIND)
       info007 LIKE type_t.chr20,     #區域名稱(ERPID)
       info008 LIKE type_t.chr500,    #patch目的說明
       info009 LIKE type_t.chr1,      #datafix1 - Y/N 170301
       info010 LIKE type_t.chr1,      #datafix2 - Y/N 170301
       info011 LIKE type_t.chr1,      #datafix3 - Y/N 170301
       info012 LIKE type_t.num10,     #ut所需空間(MB) 170301
       info013 LIKE type_t.num10,     #dsblob所需空間(MB) 170301
       info014 LIKE type_t.num10,     #dsdata所需空間(MB) 170301
       info015 LIKE type_t.chr1,      #匯出環境資訊(DGENV) 190111
       info016 LIKE type_t.chr1,      #dfile設定頁是否包含awsi011資料-Y/N #2019/06/17
       info017 LIKE type_t.chr10,     #Patch包資料庫別(ORA/PGS...) #PGS-02785
       info018 LIKE type_t.chr10,     #匯出資料的方式(imp/pg_dump/load) #2020/08/12
       info019 LIKE type_t.chr1,      #此包有無GR報表元件 2021/02/03
       info020 LIKE type_t.chr1,      #此包有無FR報表元件 2021/02/03
       info021 LIKE type_t.chr1,      #此包有無abi報表元件 2022/04/15
       info022 LIKE type_t.chr1,      #此包有無表格包 2023/03/09
       info023 LIKE type_t.chr1,      #此包有無XR報表元件 #2023/04/11
       info024 LIKE type_t.chr10,     #匯出端db driver type #2023/10/01
       info025 LIKE type_t.chr20,     #匯出端g_db_branch #2023/10/01
       info026 LIKE type_t.chr1,      #此包有無sqlite檔 #231206-00011
       info027 LIKE type_t.chr20,     #此包fglrun版本 #241114-00004
       prog DYNAMIC ARRAY OF RECORD            
         #prog000 LIKE type_t.chr20,  #程式代號        #180329-00037 #191108-00020 remark
          prog000 LIKE dzaf_t.dzaf001,#程式代號        #191108-00020
          prog001 LIKE type_t.chr1,   #是否解開section
          prog002 LIKE type_t.chr1,   #是否為freestyle
          prog003 LIKE type_t.chr5    #歸屬行業
       END RECORD,
       #Begin: 2019/11/08
       ind_pack RECORD                #行業平台打包資訊
          std_release        LIKE type_t.chr100,  #出貨資訊t100_release line1
          ind_release        LIKE type_t.chr100,  #出貨資訊t100_release line2
          last_std_patch_num LIKE type_t.chr100   #最後一次更新標準例行包號
       END RECORD,
       #End: 2019/11/08
       #Begin: 2021/02/03
       module DYNAMIC ARRAY OF RECORD            
          module001 LIKE dzaf_t.dzaf006           #模組代號
       END RECORD 
       #End: 2021/02/03
   END RECORD

   DEFINE g_stop_chk_sd_ver_freestyle BOOLEAN  #DEFAULT=FALSE
   DEFINE g_run_adzi888               BOOLEAN  #是否為adzi888帶起   
   DEFINE g_run_adzp777               BOOLEAN  #是否為adzp777帶起 #2017/06/21
   DEFINE g_run_18_or_146             BOOLEAN  #是否為18主機/146主機打包 #2017/09/13
   DEFINE g_run_18                    BOOLEAN  #是否為18主機主機打包     #2019/06/20
   DEFINE g_run_146                   BOOLEAN  #是否為146主機主機打包    #2019/06/20
   DEFINE g_run_147                   BOOLEAN  #是否為147主機主機打包    #2019/06/20
   DEFINE g_skip_regen                BOOLEAN  #因應adzp777整併, 增加變數判斷是否重產及編譯連結 #2018/11/29   

   #Begin: 2019/04/16
   DEFINE g_top                       STRING   #環境變數$TOP
   DEFINE g_zone                      STRING   #環境變數$ZONE
   DEFINE g_erpver                    STRING   #環境變數$ERPVER
   DEFINE gb_multi_dev_pack           BOOLEAN  #True/False: 此patch包內是否包含多版本新架構新目錄
   DEFINE gb_multi_dev_zone           BOOLEAN  #True/False: 此環境是否屬於多版本共存區
   DEFINE g_multi_dev_zone_list       STRING   #多版本共存區清單
   DEFINE ga_multi_dev                DYNAMIC ARRAY OF RECORD   #多版本共存區詳細資訊
             zone                     STRING,  #$ZONE
             top                      STRING,  #$TOP
             erpver                   STRING   #$ERPVER
          END RECORD
   DEFINE gb_erpver_diff              BOOLEAN  #True/False 打包檔案的ERPVER與客戶家的不同
   #End: 2019/04/16
   DEFINE g_oracle_sid                STRING   #環境變數$ORACLE_SID #2019/09/17
   DEFINE g_booking_azzp660           BOOLEAN  #解包用:ddata是否有booking azzp660 #2019/09/17
   DEFINE g_booking_azzp999           BOOLEAN  #解包用:ddata是否有booking azzp999 #2019/12/18
   DEFINE g_syn_allent_list           STRING   #解包用:ddata是否有booking 需要複製ent資料的作業 #2020/04/06
   DEFINE g_topind                    STRING   #環境變數$TOPIND #2019/11/08
   DEFINE g_run_adzi888_with_god_mode BOOLEAN  #啟動adzi888是否帶debug參數 #2020/05/26
   DEFINE g_topdb                     STRING   #環境變數$TOPDB #PGS-02785
   DEFINE g_hostname                  STRING   #環境變數$HOSTNAME #PGS-02785
   DEFINE g_t100synco2p               STRING   #環境變數$T100SYNCO2P #過單:來源是oracle,目的是postgre #2020/08/12
   DEFINE gb_not_routine_patch        BOOLEAN  #小包patch註記(非例行包) #2020/12/10
   DEFINE ga_t100patch_multilang      DYNAMIC ARRAY OF STRING  #環境變數$T100PATCH_MULTILANG #patch打包自訂資料語系清單 #2021/05/11
   DEFINE g_t100skipunpackfr          STRING   #環境變數$T100SKIPUNPACKFR #過單匯入時控制skip覆蓋fr cpt樣板 #2021/07/15
   DEFINE g_t100skipunpackabi         STRING   #環境變數$T100SKIPUNPACKABI #過單匯入時控制skip覆蓋abi 樣板 #2022/04/15
   DEFINE g_database_type             STRING   #資料庫類型 #2023/10/01
   #Begin: 231206-00011
   DEFINE g_t100sqlite                STRING   #資料以sqlite處理否floag Y/N 
   DEFINE gr_patch4json_para  RECORD
           mode        STRING, #資料庫匯出josn檔/json檔匯入資料庫 exp/imp
           table       STRING, #表格代號
           file        STRING, #json檔(含full path)
           sql         STRING, #表格匯出where條件, value若有單引號請先轉成跳脫符號  '  --> &apos;
           real_schema STRING  #實際處理的表格(匯入用)
          END RECORD
   #End: 231206-00011
   DEFINE g_t100synco2p_notrans       STRING   #環境變數$T100SYNCO2P_NOTRANS #(used unload/load但不強制轉型) #240603-00034
   DEFINE g_t100synco2p_reverse       STRING   #環境變數$T100SYNCO2P_REVERSE #(used unload/load,date強制轉型,逆向補時分秒) #240813-00016

END GLOBALS
