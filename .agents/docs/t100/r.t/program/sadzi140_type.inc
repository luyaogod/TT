{
  異動型態    異動單號       日 期       異動者      異動內容
  ========= ============= ========== ========== ===============================================
  Modify                  20170912   Ernestliou 1.索引匯入時判斷有無已存在 index 欄位, 有則匯入方設失效
  Modify                  20171120   02160      1.新增欄位挑選式匯入功能
  Modify                  20180129   02160      1.匯入時判斷, 若匯出檔的 ERPVER 和目標區的不一樣則報錯
  Modify                  20180302   02160      1.刪除標準欄位時給予警告, 並將刪除的欄位紀錄在表格(dzhj_t)中
  Modify                  20180307   02160      1.增加Patch Mode, 以讓 adzi140_patch 順利執行
  Modify                  20190304   02160      1.標準表可以在行業環境增加專屬欄位
  Modify                  20190424   02160      1.支援在建表時同步建立實體紀錄 Trigger 及 adzi190 上的 Trigger Scripts
  Modify                  20200330   02160      1.修改支援 PosgreSQL(1.1.072)   
}  

SCHEMA ds

PRIVATE TYPE T_DZEA_T RECORD
                recordtype LIKE type_t.chr1,
                dzeaseq    VARCHAR(4),
                dzea003    LIKE dzea_t.dzea003,
                dzea001    LIKE dzea_t.dzea001,
                dzea002    LIKE dzea_t.dzea002,
                dzea004    LIKE dzea_t.dzea004,
                dzea005    LIKE dzea_t.dzea005,
                dzea006    LIKE dzea_t.dzea006,
                dzea007    LIKE dzea_t.dzea007,
                dzea011    LIKE dzea_t.dzea011,
                dzlm005    VARCHAR(10),
                dzlm006    VARCHAR(10),
                dzlm007    VARCHAR(20),
                dzlm008    VARCHAR(2),
                dzlm012    VARCHAR(20),
                dzlm015    VARCHAR(20),
                ooag011    VARCHAR(320),
                dzea012    LIKE dzea_t.dzea012,
                dzea013    LIKE dzea_t.dzea013,
                dzea014    LIKE dzea_t.dzea014,
                dzea015    LIKE dzea_t.dzea015,
                dzea016    LIKE dzea_t.dzea016,
                dzea017    LIKE dzea_t.dzea017,
                dzea018    LIKE dzea_t.dzea018,
                dzea019    LIKE dzea_t.dzea019,
                dzea020    LIKE dzea_t.dzea020,
                dzea021    LIKE dzea_t.dzea021,
                dzeastus   LIKE dzea_t.dzeastus                
             END RECORD

PRIVATE TYPE T_DZEA_T_COLOR RECORD 
                recordtype STRING,
                dzeaseq    STRING,
                dzea003    STRING,
                dzea001    STRING,
                dzea002    STRING,
                dzea004    STRING,
                dzea005    STRING,
                dzea006    STRING,
                dzea007    STRING,
                dzea011    STRING,
                dzlm008    STRING,
                dzea012    STRING,
                dzea013    STRING,
                dzea014    STRING,
                dzea015    STRING,
                dzea016    STRING,
                dzea017    STRING,
                dzea018    STRING,
                dzea019    STRING,
                dzea020    STRING,
                dzea021    STRING,
                dzeastus   STRING
             END RECORD

PRIVATE TYPE T_DZEA_ORIG_DATA RECORD 
               dzeastus   like dzea_t.dzeastus,
               dzea001    like dzea_t.dzea001,
               dzea002    like dzea_t.dzea002,
               dzea003    like dzea_t.dzea003,
               dzea004    like dzea_t.dzea004,
               dzea005    like dzea_t.dzea005,
               dzea006    like dzea_t.dzea006,
               dzea007    like dzea_t.dzea007,
               dzea008    like dzea_t.dzea008,
               dzea009    like dzea_t.dzea009,
               dzea010    like dzea_t.dzea010,
               dzeaownid  like dzea_t.dzeaownid,
               dzeaowndp  like dzea_t.dzeaowndp,
               dzeacrtid  like dzea_t.dzeacrtid,
               dzeacrtdp  like dzea_t.dzeacrtdp,
               dzeacrtdt  like dzea_t.dzeacrtdt,
               dzeamodid  like dzea_t.dzeamodid,
               dzeamoddt  like dzea_t.dzeamoddt,
               dzeacnfid  like dzea_t.dzeacnfid,
               dzeacnfdt  like dzea_t.dzeacnfdt,
               dzea011    like dzea_t.dzea011,
               dzea012    like dzea_t.dzea012,
               dzea013    like dzea_t.dzea013,
               dzea014    like dzea_t.dzea014,
               dzea015    like dzea_t.dzea015,
               dzea016    like dzea_t.dzea016,
               dzea017    like dzea_t.dzea017,
               dzea018    like dzea_t.dzea018,
               dzea019    like dzea_t.dzea019,
               dzea020    like dzea_t.dzea020,
               dzea021    like dzea_t.dzea021
             END RECORD

PRIVATE TYPE T_DZEAL_ORIG_DATA RECORD 
               dzeal001  like dzeal_t.dzeal001,
               dzeal002  like dzeal_t.dzeal002,
               dzeal003  like dzeal_t.dzeal003,
               dzeal004  like dzeal_t.dzeal004,
               dzeal005  like dzeal_t.dzeal005
             END RECORD
             
PRIVATE TYPE T_DZEB_T RECORD
                recordtype    LIKE type_t.chr1,
                groupid       VARCHAR(256),
                dzebseq       VARCHAR(4),
                dzeb001       LIKE dzeb_t.dzeb001,
                dzeb002       LIKE dzeb_t.dzeb002,
                dzeb003       LIKE dzeb_t.dzeb003,
                dzeb004       LIKE dzeb_t.dzeb004,
                dzeb005       LIKE dzeb_t.dzeb005, 
                dzeb006       LIKE dzeb_t.dzeb006,
                dzeb006_desc  VARCHAR(512),
                dzeb007       LIKE dzeb_t.dzeb007,
                dzeb008       LIKE dzeb_t.dzeb008,
                dzeb012       LIKE dzeb_t.dzeb012,
                dzeb021       LIKE dzeb_t.dzeb021,
                dzeb022       LIKE dzeb_t.dzeb022,
                dzeb023       LIKE dzeb_t.dzeb023,
                dzeb024       LIKE dzeb_t.dzeb024,
                dzeb027       LIKE dzeb_t.dzeb027,
                dzeb028       LIKE dzeb_t.dzeb028, 
                dzeb029       LIKE dzeb_t.dzeb029, 
                dzeb030       LIKE dzeb_t.dzeb030, 
                dzeb031       LIKE dzeb_t.dzeb031, 
                dzeb002_orig  LIKE dzeb_t.dzeb002,
                dzep012       LIKE dzep_t.dzep012,
                category      VARCHAR(2048),
                lang_modified VARCHAR(1)
             END RECORD

PRIVATE TYPE T_DZEB_T_COLOR RECORD
                recordtype    STRING,
                groupid       STRING,
                dzebseq       STRING,
                dzeb001       STRING,
                dzeb002       STRING,
                dzeb003       STRING,
                dzeb004       STRING,
                dzeb005       STRING,
                dzeb006       STRING,
                dzeb006_desc  STRING,
                dzeb007       STRING,
                dzeb008       STRING,
                dzeb012       STRING,
                dzeb021       STRING,
                dzeb022       STRING,
                dzeb023       STRING,
                dzeb024       STRING,
                dzeb027       STRING,
                dzeb028       STRING,
                dzeb029       STRING,
                dzeb030       STRING,
                dzeb031       STRING,
                dzeb002_orig  STRING,
                dzep012       STRING,
                category      STRING
             END RECORD
             
PRIVATE TYPE T_DZEB_ORIG_DATA RECORD 
               dzebstus  like dzeb_t.dzebstus,
               dzeb001   like dzeb_t.dzeb001,
               dzeb002   like dzeb_t.dzeb002,
               dzeb003   like dzeb_t.dzeb003,
               dzeb004   like dzeb_t.dzeb004,
               dzeb005   like dzeb_t.dzeb005,
               dzeb006   like dzeb_t.dzeb006,
               dzeb007   like dzeb_t.dzeb007,
               dzeb008   like dzeb_t.dzeb008,
               dzeb021   like dzeb_t.dzeb021,
               dzeb022   like dzeb_t.dzeb022,
               dzeb023   like dzeb_t.dzeb023,
               dzeb024   like dzeb_t.dzeb024,
               dzeb027   like dzeb_t.dzeb027,
               dzebownid like dzeb_t.dzebownid,
               dzebowndp like dzeb_t.dzebowndp,
               dzebcrtid like dzeb_t.dzebcrtid,
               dzebcrtdp like dzeb_t.dzebcrtdp,
               dzebcrtdt like dzeb_t.dzebcrtdt,
               dzebmodid like dzeb_t.dzebmodid,
               dzebmoddt like dzeb_t.dzebmoddt,
               dzebcnfid like dzeb_t.dzebcnfid,
               dzebcnfdt like dzeb_t.dzebcnfdt,
               dzeb028   like dzeb_t.dzeb028,
               dzeb029   like dzeb_t.dzeb029,
               dzeb012   like dzeb_t.dzeb012,
               dzeb030   like dzeb_t.dzeb030,
               dzeb031   like dzeb_t.dzeb031  
             END RECORD
             
PRIVATE TYPE T_DZEBL_ORIG_DATA RECORD 
               dzebl000  like dzebl_t.dzebl000,
               dzebl001  like dzebl_t.dzebl001,
               dzebl002  like dzebl_t.dzebl002,
               dzebl003  like dzebl_t.dzebl003,
               dzebl004  like dzebl_t.dzebl004
             END RECORD
             
PRIVATE TYPE T_DZEC_T RECORD
                recordtype   LIKE type_t.chr1,
                dzec001      LIKE dzec_t.dzec001,
                dzec002      LIKE dzec_t.dzec002,
                dzec003      LIKE dzec_t.dzec003,
                dzec004      LIKE dzec_t.dzec004,
                dzec005      LIKE dzec_t.dzec005,
                dzec006      LIKE dzec_t.dzec006,
                dzec007      LIKE dzec_t.dzec007,
                dzec008      LIKE dzec_t.dzec008,
                dzecstus     LIKE dzec_t.dzecstus, #20170912
                dzec002_orig LIKE dzec_t.dzec002
             END RECORD

PRIVATE TYPE T_DZEC_T_COLOR RECORD
                recordtype   STRING,
                dzec001      STRING,
                dzec002      STRING,
                dzec003      STRING,
                dzec004      STRING,
                dzec005      STRING,
                dzec006      STRING,
                dzec007      STRING,
                dzec008      STRING,
                dzecstus     STRING, #20170912
                dzec002_orig STRING
             END RECORD
             
PRIVATE TYPE T_DZEC_ORIG_DATA RECORD 
               dzecstus  like dzec_t.dzecstus,
               dzec001   like dzec_t.dzec001,
               dzec002   like dzec_t.dzec002,
               dzec003   like dzec_t.dzec003,
               dzec004   like dzec_t.dzec004,
               dzecownid like dzec_t.dzecownid,
               dzecowndp like dzec_t.dzecowndp,
               dzeccrtid like dzec_t.dzeccrtid,
               dzeccrtdp like dzec_t.dzeccrtdp,
               dzeccrtdt like dzec_t.dzeccrtdt,
               dzecmodid like dzec_t.dzecmodid,
               dzecmoddt like dzec_t.dzecmoddt,
               dzeccnfid like dzec_t.dzeccnfid,
               dzeccnfdt like dzec_t.dzeccnfdt,
               dzec005   like dzec_t.dzec005,
               dzec006   like dzec_t.dzec006,
               dzec007   like dzec_t.dzec007,
               dzec008   like dzec_t.dzec008
             END RECORD
             
PRIVATE TYPE T_DZED_T RECORD
                recordtype   LIKE type_t.chr1,
                dzed001      LIKE dzed_t.dzed001,
                dzed002      LIKE dzed_t.dzed002,
                dzed003      LIKE dzed_t.dzed003,
                dzed004      LIKE dzed_t.dzed004,
                dzed005      LIKE dzed_t.dzed005,
                dzed006      LIKE dzed_t.dzed006,
                dzed007      LIKE dzed_t.dzed007,
                dzed008      LIKE dzed_t.dzed008,
                dzed009      LIKE dzed_t.dzed009,
                dzed010      LIKE dzed_t.dzed010,
                dzed002_orig LIKE dzed_t.dzed002
             END RECORD

PRIVATE TYPE T_DZED_ORIG_DATA RECORD 
               dzedstus  like dzed_t.dzedstus,
               dzed001   like dzed_t.dzed001,
               dzed002   like dzed_t.dzed002,
               dzed003   like dzed_t.dzed003,
               dzed004   like dzed_t.dzed004,
               dzed005   like dzed_t.dzed005,
               dzed006   like dzed_t.dzed006,
               dzedownid like dzed_t.dzedownid,
               dzedowndp like dzed_t.dzedowndp,
               dzedcrtid like dzed_t.dzedcrtid,
               dzedcrtdp like dzed_t.dzedcrtdp,
               dzedcrtdt like dzed_t.dzedcrtdt,
               dzedmodid like dzed_t.dzedmodid,
               dzedmoddt like dzed_t.dzedmoddt,
               dzedcnfid like dzed_t.dzedcnfid,
               dzedcnfdt like dzed_t.dzedcnfdt,
               dzed007   like dzed_t.dzed007,
               dzed008   like dzed_t.dzed008,
               dzed009   like dzed_t.dzed009,
               dzed010   like dzed_t.dzed010
             END RECORD

PRIVATE TYPE T_DZEN_T RECORD
               recordtype   LIKE type_t.chr1,
               dzen001      LIKE dzen_t.dzen001,
               dzen002      LIKE dzen_t.dzen002,
               dzen003      LIKE dzen_t.dzen003,
               dzen004      LIKE dzen_t.dzen004,
               dzen005      LIKE dzen_t.dzen005,
               dzen006      LIKE dzen_t.dzen006,
               dzen007      LIKE dzen_t.dzen007,
               dzen008      LIKE dzen_t.dzen008,
               dzen009      LIKE dzen_t.dzen009,
               dzen010      LIKE dzen_t.dzen010,
               dzen011      LIKE dzen_t.dzen011,
               dzen012      LIKE dzen_t.dzen012,
               dzen013      LIKE dzen_t.dzen013,
               dzen002_orig LIKE dzen_t.dzen002
             END RECORD

PRIVATE TYPE T_DZEN_ORIG_DATA RECORD
               dzen001      LIKE dzen_t.dzen001,
               dzen002      LIKE dzen_t.dzen002,
               dzen003      LIKE dzen_t.dzen003,
               dzen004      LIKE dzen_t.dzen004,
               dzen005      LIKE dzen_t.dzen005,
               dzen006      LIKE dzen_t.dzen006,
               dzen007      LIKE dzen_t.dzen007,
               dzen008      LIKE dzen_t.dzen008,
               dzen009      LIKE dzen_t.dzen009,
               dzen010      LIKE dzen_t.dzen010,
               dzen011      LIKE dzen_t.dzen011,
               dzen012      LIKE dzen_t.dzen012,
               dzen013      LIKE dzen_t.dzen013,
               dzencrtid    LIKE dzen_t.dzencrtid,
               dzencrtdt    LIKE dzen_t.dzencrtdt,
               dzenmodid    LIKE dzen_t.dzenmodid,
               dzenmoddt    LIKE dzen_t.dzenmoddt
             END RECORD
             
PRIVATE TYPE T_DZEF_T RECORD
                recordtype   LIKE type_t.chr1,
                DZEF001      LIKE DZEF_t.DZEF001,
                DZEF002      LIKE DZEF_t.DZEF002,
                DZEF003      LIKE DZEF_t.DZEF003,
                DZEF004      LIKE DZEF_t.DZEF004,
                DZEF005      LIKE DZEF_t.DZEF005,
                DZEF006      LIKE DZEF_t.DZEF006,
                DZEF007      LIKE DZEF_t.DZEF007,
                DZEF008      LIKE DZEF_t.DZEF008,
                DZEF009      LIKE DZEF_t.DZEF009,
                DZEF010      LIKE DZEF_t.DZEF010
             END RECORD
             
PRIVATE TYPE T_DZER_T RECORD
                recordtype   LIKE type_t.chr1,
                dzer001      LIKE dzer_t.dzer001,
                dzer002      LIKE dzer_t.dzer002,
                dzer003      LIKE dzer_t.dzer003,
                dzer004      LIKE dzer_t.dzer004,
                dzer005      LIKE dzer_t.dzer005,
                dzer006      LIKE dzer_t.dzer006,
                dzer007      LIKE dzer_t.dzer007,
                dzer008      LIKE dzer_t.dzer008
             END RECORD

PRIVATE TYPE T_DZEP_T RECORD
                recordtype   LIKE type_t.chr1,
                DZEP001      LIKE DZEP_T.DZEP001,
                DZEP002      LIKE DZEP_T.DZEP002,
                DZEP003      LIKE DZEP_T.DZEP003,
                DZEP004      LIKE DZEP_T.DZEP004,
                DZEP005      LIKE DZEP_T.DZEP005,
                DZEP006      LIKE DZEP_T.DZEP006,
                DZEP007      LIKE DZEP_T.DZEP007,
                DZEP008      LIKE DZEP_T.DZEP008,
                DZEP009      LIKE DZEP_T.DZEP009,
                DZEP010      LIKE DZEP_T.DZEP010,
                DZEP011      LIKE DZEP_T.DZEP011,
                DZEP012      LIKE DZEP_T.DZEP012,
                DZEP013      LIKE DZEP_T.DZEP013,
                DZEP014      LIKE DZEP_T.DZEP014,
                DZEP015      LIKE DZEP_T.DZEP015,
                DZEP016      LIKE DZEP_T.DZEP016,
                DZEP017      LIKE DZEP_T.DZEP017,
                DZEP018      LIKE DZEP_T.DZEP018,
                DZEP019      LIKE DZEP_T.DZEP019,
                DZEP020      LIKE DZEP_T.DZEP020,
                DZEP021      LIKE DZEP_T.DZEP021,
                DZEP022      LIKE DZEP_T.DZEP022,
                DZEP023      LIKE DZEP_T.DZEP023,
                DZEP024      LIKE DZEP_T.DZEP024,
                DZEP025      LIKE DZEP_T.DZEP025,
                DZEP026      LIKE DZEP_T.DZEP026,
                DZEP027      LIKE DZEP_T.DZEP027,
                DZEP028      LIKE DZEP_T.DZEP028 
             END RECORD

PRIVATE TYPE T_DZEQ_T RECORD
                recordtype   LIKE type_t.chr1,
                dzeq001      LIKE dzeq_t.dzeq001,
                dzeq002      LIKE dzeq_t.dzeq002,
                dzeq003      LIKE dzeq_t.dzeq003,
                dzeq004      LIKE dzeq_t.dzeq004,
                dzeq005      LIKE dzeq_t.dzeq005,
                dzeq006      LIKE dzeq_t.dzeq006,
                dzeq007      LIKE dzeq_t.dzeq007,
                dzeq008      LIKE dzeq_t.dzeq008,
                dzeq009      LIKE dzeq_t.dzeq009
             END RECORD
             
PUBLIC TYPE T_SNAPSHOT RECORD
              dzev001 LIKE dzev_t.dzev001,
              dzev002 LIKE dzev_t.dzev002,
              dzev003 LIKE dzev_t.dzev003,
              dzev004 LIKE dzev_t.dzev004,
              dzev005 LIKE dzev_t.dzev005,
              dzev006 LIKE dzev_t.dzev006,
              dzev007 LIKE dzev_t.dzev007,
              dzev008 LIKE dzev_t.dzev008,
              dzev009 LIKE dzev_t.dzev009,
              dzev010 LIKE dzev_t.dzev010,
              dzev011 LIKE dzev_t.dzev011,
              dzev012 LIKE dzev_t.dzev012,
              dzev013 LIKE dzev_t.dzev013,
              dzev014 LIKE dzev_t.dzev014,
              dzev015 LIKE dzev_t.dzev015,
              dzev016 LIKE dzev_t.dzev016,
              dzev017 LIKE dzev_t.dzev017,
              dzev018 LIKE dzev_t.dzev018,
              dzev019 LIKE dzev_t.dzev019,
              dzev020 LIKE dzev_t.dzev020,
              dzev021 LIKE dzev_t.dzev021,
              dzev022 LIKE dzev_t.dzev022,
              dzev023 LIKE dzev_t.dzev023
            END RECORD
            
PUBLIC TYPE T_CONSTRAINT_INDEX RECORD
              dzew001  LIKE dzew_t.dzew001,
              dzew002  LIKE dzew_t.dzew002,
              dzew003  LIKE dzew_t.dzew003,
              dzew004  LIKE dzew_t.dzew004,
              dzew005  LIKE dzew_t.dzew005,
              dzew006  LIKE dzew_t.dzew006,
              dzew007  LIKE dzew_t.dzew007,
              dzew008  LIKE dzew_t.dzew008,
              dzew009  LIKE dzew_t.dzew009,
              dzew010  LIKE dzew_t.dzew010,
              dzew011  LIKE dzew_t.dzew011,
              dzew012  LIKE dzew_t.dzew012,
              dzew013  LIKE dzew_t.dzew013,
              dzew014  LIKE dzew_t.dzew014,
              dzewstus LIKE dzew_t.dzewstus #20170912
            END RECORD

PUBLIC TYPE T_COLUMN_DIFF RECORD
              dzev005 LIKE dzev_t.dzev005,
              dzev006 LIKE dzev_t.dzev006,
              dzev007 LIKE dzev_t.dzev007,
              dzev008 LIKE dzev_t.dzev008,
              dzev009 LIKE dzev_t.dzev009,
              dzev021 LIKE dzev_t.dzev021
            END RECORD

PUBLIC TYPE T_INDEX_DIFF RECORD
              dzec001 LIKE dzec_t.dzec001,
              dzec002 LIKE dzec_t.dzec002,
              dzec003 LIKE dzec_t.dzec003,
              dzec004 LIKE dzec_t.dzec004
            END RECORD

PRIVATE TYPE T_DZIG_T RECORD
                dzig001 LIKE dzig_t.dzig001,
                dzig002 LIKE dzig_t.dzig002,
                dzig003 LIKE dzig_t.dzig003
             END RECORD
            
PUBLIC TYPE T_COLUMN_ALTER RECORD
              dzeb002 LIKE dzeb_t.dzeb002, --dzev005 
              dzeb004 LIKE dzeb_t.dzeb004, --dzev006
              dzeb005 LIKE dzeb_t.dzeb005, --dzev007
              dzeb007 LIKE dzeb_t.dzeb007, --dzev008
              dzeb008 LIKE dzeb_t.dzeb008, --dzev009
              dzeb012 LIKE dzeb_t.dzeb012  --dzev021  
            END RECORD
            
PUBLIC TYPE T_CONSTRAINT_DIFF RECORD
              dzed001 like dzed_t.dzed001,
              dzed002 like dzed_t.dzed002,
              dzed003 like dzed_t.dzed003,
              dzed004 like dzed_t.dzed004,
              dzed005 like dzed_t.dzed005,
              dzed006 like dzed_t.dzed006
            END RECORD

PUBLIC TYPE T_PRIVILEGE_DIFF RECORD
              DZEN002 LIKE DZEN_T.DZEN002, 
              DZEN003 LIKE DZEN_T.DZEN003,
              DZEN004 LIKE DZEN_T.DZEN004,
              DZEN005 LIKE DZEN_T.DZEN005,
              DZEN006 LIKE DZEN_T.DZEN006,
              DZEN007 LIKE DZEN_T.DZEN007,
              DZEN008 LIKE DZEN_T.DZEN008,
              DZEN009 LIKE DZEN_T.DZEN009, 
              DZEN010 LIKE DZEN_T.DZEN010 
            END RECORD

PUBLIC TYPE T_DZHL_T RECORD
              DZHL001	  LIKE DZHL_T.DZHL001,	
              DZHL002	  LIKE DZHL_T.DZHL002,	
              DZHL003	  LIKE DZHL_T.DZHL003,	
              DZHL004	  LIKE DZHL_T.DZHL004,	
              DZHL005	  LIKE DZHL_T.DZHL005,	
              DZHLCRTID	LIKE DZHL_T.DZHLCRTID,
              DZHLCRTDT	LIKE DZHL_T.DZHLCRTDT,
              DZHLMODID	LIKE DZHL_T.DZHLMODID,
              DZHLMODDT	LIKE DZHL_T.DZHLMODDT
            END RECORD 
            
PUBLIC TYPE T_DZEA_CREATE_TABLE RECORD
              dct_module_name           LIKE dzea_t.dzea003,
              dct_table_type            LIKE dzea_t.dzea004,
              dct_table_name            LIKE dzea_t.dzea001,
              dct_tail_code             VARCHAR(10),
              dct_table_description     LIKE dzea_t.dzea002,
              dct_is_multi_lang_table   LIKE dzea_t.dzea005,
              dct_is_system_table       LIKE dzea_t.dzea006,
              dct_create_notice_trigger LIKE dzea_t.dzea007, #20190424 #Chenge dct_common_columns to dct_create_notice_trigger
              dct_define_group          LIKE dzea_t.dzea008, 
              dct_up_level_table        LIKE dzea_t.dzea009,
              dct_table_space           VARCHAR(20),
              dct_master_db             VARCHAR(20),
              dct_master_user           VARCHAR(20),  
              dct_is_altered            LIKE dzea_t.dzea011,
              dct_industry_type         VARCHAR(20),
              dct_alm_construct_version VARCHAR(20),  
              dct_alm_sd_version        VARCHAR(20),
              dct_alm_request_no        VARCHAR(50),
              dct_dgenv                 VARCHAR(02),   
              dct_orig_module_name      LIKE dzea_t.dzea016,
              dct_shipping_notice       LIKE dzea_t.dzea017,
              dct_orig_dgenv            VARCHAR(02)   
            END RECORD

PUBLIC TYPE T_BASE_TABLE_INFO RECORD
              bti_if_rename         BOOLEAN,
              bti_table_name        LIKE dzea_t.dzea001,
              bti_table_type        LIKE dzea_t.dzea004,
              bti_table_description LIKE dzea_t.dzea002,
              bti_module_name       LIKE dzea_t.dzea003
            END RECORD

PUBLIC TYPE T_DIRECTORY_LIST RECORD
              dl_type  VARCHAR(01),
              dl_name  varchar(255)
            END RECORD
            
PUBLIC TYPE T_PKG_SYS_INFO RECORD
              psi_DGENV          VARCHAR(10),
              psi_ERPID          VARCHAR(50),
              psi_CUST           VARCHAR(50),
              psi_DB_TYPE        VARCHAR(10),
              psi_TOPIND         VARCHAR(10),
              psi_RENAMED        VARCHAR(10),
              psi_RT_DEBUG_MODE  VARCHAR(1),
              psi_RT_PATCH_MODE  VARCHAR(1),  #20180307
              psi_ERPVER         VARCHAR(20), #20180129
              psi_CONSTRUCT_VER  VARCHAR(10),  #20180410
              psi_T100PASSERPVER VARCHAR(10)  #20180420
            END RECORD 
            
PUBLIC TYPE T_DB_CONNSTR RECORD
              db_source        VARCHAR(20),
              db_username      VARCHAR(20),
              db_password      VARCHAR(20),
              db_schema        VARCHAR(20),
              db_sid           VARCHAR(20),
              db_sql_filename  VARCHAR(150),
              db_version       VARCHAR(50),
              db_host_ip       VARCHAR(200), #20191209
              db_host_port     VARCHAR(10)   #20200330
            END RECORD

PUBLIC TYPE T_TABLE_IN_DB_TYPE RECORD
              tidt_db          VARCHAR(20),
              tidt_object_type VARCHAR(20)
            END RECORD             

PUBLIC TYPE T_TABLE_SYNONYM RECORD
              p_row_number   INTEGER, --VARCHAR(10), #DZEU005 
              p_master_db    VARCHAR(15),
              p_table_name   VARCHAR(15), #DZEU001
              p_type         VARCHAR(2),
              p_checkbox     VARCHAR(1),  #DZEU004
              p_ent_db       VARCHAR(20), #DZEU002
              p_sch_type     VARCHAR(2)   #DZEU003 
            END RECORD

PUBLIC TYPE T_GZCC_T RECORD
              gzcc001  LIKE gzcc_t.gzcc001,
              gzcc002  LIKE gzcc_t.gzcc002,
              gzcc003  LIKE gzcc_t.gzcc003,
              gzcc004  LIKE gzcc_t.gzcc004,
              gzcc005  LIKE gzcc_t.gzcc005,
              gzcc006  LIKE gzcc_t.gzcc006,
              gzccstus LIKE gzcc_t.gzccstus
            END RECORD
            
PUBLIC TYPE T_DZLM_INFO RECORD 
              MODULE_NAME VARCHAR(20),
              TABLE_NAME  VARCHAR(200),
              TABLE_DESC  VARCHAR(130),
              VERSION     VARCHAR(20) 
            END RECORD

PUBLIC TYPE  T_DZLM_T_SCR RECORD 
                TYPES   VARCHAR(10),
                DZLM001	VARCHAR(10),  -- 建構類型
                DZLM002	VARCHAR(200), -- 建構代號
                DZLM003	VARCHAR(500), -- 建構名稱
                DZLM004	VARCHAR(10),  -- 模組    
                DZLM005	VARCHAR(10),  -- 建構版號
                DZLM006	VARCHAR(10),  -- SD版號  
                DZLM007	VARCHAR(20),  -- SD工號  
                DZLM008	VARCHAR(1),   -- SD狀態  
                DZLM009	VARCHAR(10),  -- PR版號  
                DZLM010	VARCHAR(20),  -- PR工號  
                DZLM011	VARCHAR(1),   -- PR狀態  
                DZLM012	VARCHAR(20),  -- 需求單號
                DZLM013	VARCHAR(10),  -- 產品代號
                DZLM014	VARCHAR(10),  -- 產品版本
                DZLM015	INTEGER,      -- 作業項次
                DZLM016	VARCHAR(40),  -- 客戶代碼    
                DZLM017	DATE,         -- 簽入時間
                DZLM018 VARCHAR(40),  -- SD GUID
                DZLM019 VARCHAR(40),  -- PR GUID
                DZLM020 VARCHAR(1),   -- SD 已下載
                DZLM021 VARCHAR(1)    -- PR 已下載
               END RECORD
           
PUBLIC TYPE T_DATA_LINE RECORD
                DATA001 VARCHAR(500),
                DATA002 VARCHAR(500), 
                DATA003 VARCHAR(500),
                DATA004 VARCHAR(500),
                DATA005 VARCHAR(500),
                DATA006 VARCHAR(500),
                DATA007 VARCHAR(500),
                DATA008 VARCHAR(500),
                DATA009 VARCHAR(500),
                DATA010 VARCHAR(500),
                DATA011 VARCHAR(500),
                DATA012 VARCHAR(500), 
                DATA013 VARCHAR(500),
                DATA014 VARCHAR(500),
                DATA015 VARCHAR(500),
                DATA016 VARCHAR(500),
                DATA017 VARCHAR(500),
                DATA018 VARCHAR(500),
                DATA019 VARCHAR(500),
                DATA020 VARCHAR(500),
                DATA021 VARCHAR(500),
                DATA022 VARCHAR(500), 
                DATA023 VARCHAR(500),
                DATA024 VARCHAR(500),
                DATA025 VARCHAR(500),
                DATA026 VARCHAR(500),
                DATA027 VARCHAR(500),
                DATA028 VARCHAR(500),
                DATA029 VARCHAR(500),
                DATA030 VARCHAR(500)
              END RECORD

PUBLIC TYPE T_REVISION RECORD 
              rv_RECORDTYPE   VARCHAR(1),      --RECORDTYPE  
              rv_CHECKBOX     VARCHAR(4),      --CHECKBOX    
              rv_ALM_VERSION  VARCHAR(30),     --VERSION     
              rv_REQUEST_NO   VARCHAR(512),    --REV_DESC    
              rv_SEQ_PAD_ZERO VARCHAR(30),     --REVISION_PAD
              rv_SEQUENCE     VARCHAR(30)      --REVISION    
            END RECORD

PUBLIC TYPE T_VERSION_INFO RECORD
              ALM_VERSION VARCHAR(30),
              REVISION    VARCHAR(30)
            END RECORD  
             
PUBLIC TYPE T_COLUMN_DATA RECORD
              COLUMN_NAME   VARCHAR(200),
              ISKEY         VARCHAR(200),
              NULLABLE      VARCHAR(200),
              DATA_TYPE     VARCHAR(200),
              DATA_LENGTH   VARCHAR(200),
              DEFAULT_VALUE VARCHAR(4000)
            END RECORD

PUBLIC TYPE T_COLUMNS_DATA_IN_MEM RECORD
              OWNER         VARCHAR(200),
              TABLE_NAME    VARCHAR(200),
              COLUMN_NAME   VARCHAR(200),
              ISKEY         VARCHAR(200),
              NULLABLE      VARCHAR(200),
              DATA_TYPE     VARCHAR(200),
              DATA_LENGTH   VARCHAR(200),
              DEFAULT_VALUE VARCHAR(4000)
            END RECORD
                          
PUBLIC TYPE T_DATA_TYPE RECORD
              DT_TYPE     VARCHAR(200), 
              DT_LENGTH   INTEGER,
              DT_INTEGER  INTEGER,
              DT_DECIMAL  INTEGER,
              DT_PINTEGER INTEGER
            END RECORD
                          
PUBLIC TYPE T_INDEX_DATA RECORD
              TABLE_NAME   VARCHAR(200),
              INDEX_NAME   VARCHAR(200),
              INDEX_TYPE   VARCHAR(200),
              COLUMN_NAME  VARCHAR(200)
            END RECORD

PUBLIC TYPE T_KEY_DATA RECORD
               TABLE_NAME    VARCHAR(200),
               KEY_NAME      VARCHAR(200),
               KEY_TYPE      VARCHAR(200),
               COLUMN_NAME   VARCHAR(200),
               F_TABLE_NAME  VARCHAR(200),
               F_COLUMN_NAME VARCHAR(200)
             END RECORD

PUBLIC TYPE T_DB_INFO RECORD
              DB_NAME   VARCHAR(200),
              USER_NAME VARCHAR(200)
            END RECORD

PUBLIC TYPE T_SYS_INFO RECORD
              SYS_LANG   VARCHAR(200),
              SYS_USER   VARCHAR(200),
              SYS_DEPT   VARCHAR(200)
            END RECORD

PUBLIC TYPE T_TABLE_SHIP_NOTICE RECORD
              TABLE_NAME VARCHAR(30),
              RECORDS    INTEGER
            END RECORD

PUBLIC TYPE T_COLUMN_LIST RECORD 
              COLUMN_NAME   VARCHAR(100),
              COLUMN_DESC   VARCHAR(100) 
            END RECORD

PUBLIC TYPE T_COMMON_FIELD RECORD
              common_define  VARCHAR(50),
              common_field   VARCHAR(50),
              dzep018        VARCHAR(50),
              dzef003        VARCHAR(50),
              dzef006        VARCHAR(50),
              dzef007        VARCHAR(50),
              dzef009        VARCHAR(50),
              dzef008        VARCHAR(50)
            END RECORD

PUBLIC TYPE T_FILE_DIALOG RECORD
              PATH       VARCHAR(1024),
              TYPE_DESC  VARCHAR(1024),
              TYPE_LIST  VARCHAR(50),
              CAPTION    VARCHAR(1024)
            END RECORD             

PUBLIC TYPE T_EXPORT_INFO RECORD
              WORKING_PATH  VARCHAR(1024),
              TAR_NAME      VARCHAR(1024),
              CLIENT_PATH   VARCHAR(1024),
              RESULT        BOOLEAN   
            END RECORD       

PUBLIC TYPE T_EXEC_RESULT RECORD
              er_RESULT   BOOLEAN,
              er_CODE     VARCHAR(50), 
              er_MESSAGE  VARCHAR(1024),
              er_GUID     VARCHAR(100) 
            END RECORD       

PUBLIC TYPE T_PUT_GET_FILE_PARA RECORD
              SERVER_FILE_PATH  VARCHAR(1024),
              SERVER_FILE_NAME  VARCHAR(1024),
              CLIENT_FILE_PATH  VARCHAR(1024),
              CLIENT_FILE_NAME  VARCHAR(1024)
            END RECORD                         

PUBLIC TYPE T_TAR_FILE_INFO RECORD
              TABLE_NAME         VARCHAR(100),
              CONSTRUCT_VERSION  VARCHAR(100),
              SD_VERSION         VARCHAR(100),
              REQUEST_NO         VARCHAR(100),
              SEQUENCE_NO        VARCHAR(100)
            END RECORD                         

PUBLIC TYPE T_PROPERTY RECORD
              d_property     VARCHAR(50),
              d_desc         VARCHAR(100),
              d_type         VARCHAR(50),
              d_length       VARCHAR(100),
              d_widgets      VARCHAR(50),
              d_percent      VARCHAR(02),
              d_format       VARCHAR(50),
              d_nullable     VARCHAR(02),
              d_default      VARCHAR(100),
              d_widget_width VARCHAR(10),
              d_report_width VARCHAR(10),
              d_report_digit VARCHAR(10),
              d_word_case    VARCHAR(15)              
            END RECORD

PUBLIC TYPE T_INDUSTRY_TYPE RECORD
              INDUSTRY_TYPE VARCHAR(20),
              TABLE_NAME    VARCHAR(30), 
              TAIL_CODE     VARCHAR(30) 
            END RECORD            
                        
PUBLIC TYPE T_DZEA_INFO RECORD
              DZEA001 LIKE DZEA_T.DZEA001,
              DZEA002 LIKE DZEA_T.DZEA002,
              DZEA003 LIKE DZEA_T.DZEA003,
              DZEA004 LIKE DZEA_T.DZEA004,
              DZEA005 LIKE DZEA_T.DZEA005,
              DZEA006 LIKE DZEA_T.DZEA006,
              DZEA007 LIKE DZEA_T.DZEA007,
              DZEA008 LIKE DZEA_T.DZEA008,
              DZEA009 LIKE DZEA_T.DZEA009,
              DZEA010 LIKE DZEA_T.DZEA010,
              DZEA011 LIKE DZEA_T.DZEA011,
              DZEA012 LIKE DZEA_T.DZEA012,
              DZEA013 LIKE DZEA_T.DZEA013,
              DZEA014 LIKE DZEA_T.DZEA014,
              DZEA015 LIKE DZEA_T.DZEA015
            END RECORD            

PUBLIC TYPE T_TABLE_TYPE RECORD
              DZEU001 LIKE DZEU_T.DZEU001,
              DZEU002 LIKE DZEU_T.DZEU002,
              DZEU003 LIKE DZEU_T.DZEU003,
              DZEU004 LIKE DZEU_T.DZEU004,
              DZEU005 VARCHAR(20)
            END RECORD

PUBLIC TYPE T_DZEY_T RECORD
              DZEY001 LIKE DZEY_T.DZEY001,
              DZEY002 LIKE DZEY_T.DZEY002
            END RECORD
            
PUBLIC TYPE T_SCHEMA_LIST RECORD
              sl_SCHEMA_NAME VARCHAR(50)
            END RECORD

PUBLIC TYPE T_LOCAL_LANGUAGE VARCHAR(10)

PUBLIC TYPE T_COLUMN_DEFINE RECORD
              cd_DEFINE_TYPE VARCHAR(100),
              cd_CHARACTER   VARCHAR(100),
              cd_COLUMN_TYPE VARCHAR(100),
              cd_DESC        VARCHAR(1024) 
            END RECORD            

PUBLIC TYPE T_MAPPING_DATA RECORD
              md_GLOBAL_VARIABLE  VARCHAR(100),
              md_DB_FUNCTION      VARCHAR(100),
              md_COLUMN_TYPE      VARCHAR(100)
            END RECORD

PUBLIC TYPE T_HARD_COPY_INFO RECORD
              hci_GUID       VARCHAR(40),
              hci_TABLE_NAME VARCHAR(30)
            END RECORD

PUBLIC TYPE T_PRIVILEGE_GRANTER VARCHAR(30)

PUBLIC TYPE T_APS_GRANT_SCRIPTS VARCHAR(1024)
            
PUBLIC TYPE T_PRIVILEGE_INFO RECORD
              pi_GRANTER     VARCHAR(30),
              pi_ACCEPTER    VARCHAR(30),
              pi_SELECT      VARCHAR(1),
              pi_INSERT      VARCHAR(1),
              pi_UPDATE      VARCHAR(1),
              pi_DELETE      VARCHAR(1),
              pi_REFERENCES  VARCHAR(1),
              pi_ALTER       VARCHAR(1),
              pi_INDEX       VARCHAR(1)
            END RECORD
            
PUBLIC TYPE T_STATIC_DB_SCHEMA RECORD
              sds_DB_NAME  VARCHAR(30)
            END RECORD            

PUBLIC TYPE T_DZEA_DIFF_DATA RECORD
              DZEA001 LIKE DZEA_T.DZEA001,
              DZEA002 LIKE DZEA_T.DZEA002,
              DZEA003 LIKE DZEA_T.DZEA003,
              DZEA004 LIKE DZEA_T.DZEA004,
              DZEA005 LIKE DZEA_T.DZEA005, 
              DZEA006 LIKE DZEA_T.DZEA006,
              DZEA007 LIKE DZEA_T.DZEA007,
              DZEA008 LIKE DZEA_T.DZEA008,
              DZEA009 LIKE DZEA_T.DZEA009,
              DZEA010 LIKE DZEA_T.DZEA010,
              DZEA014 LIKE DZEA_T.DZEA014,
              DZEA015 LIKE DZEA_T.DZEA015,
              DZEA016 LIKE DZEA_T.DZEA016,
              DZEA017 LIKE DZEA_T.DZEA017,
              DZEA018 LIKE DZEA_T.DZEA018
            END RECORD

PUBLIC TYPE T_DZEAL_DIFF_DATA RECORD
              DZEAL001 LIKE DZEAL_T.DZEAL001,
              DZEAL002 LIKE DZEAL_T.DZEAL002,
              DZEAL003 LIKE DZEAL_T.DZEAL003,
              DZEAL004 LIKE DZEAL_T.DZEAL004,
              DZEAL005 LIKE DZEAL_T.DZEAL005 
            END RECORD

PUBLIC TYPE T_DZEB_DIFF_DATA RECORD
              DZEB001 LIKE DZEB_T.DZEB001, 
              DZEB002 LIKE DZEB_T.DZEB002, 
              DZEB003 LIKE DZEB_T.DZEB003, 
              DZEB004 LIKE DZEB_T.DZEB004, 
              DZEB005 LIKE DZEB_T.DZEB005, 
              DZEB006 LIKE DZEB_T.DZEB006, 
              DZEB007 LIKE DZEB_T.DZEB007, 
              DZEB008 LIKE DZEB_T.DZEB008, 
              DZEB021 LIKE DZEB_T.DZEB021, 
              DZEB022 LIKE DZEB_T.DZEB022, 
              DZEB023 LIKE DZEB_T.DZEB023, 
              DZEB029 LIKE DZEB_T.DZEB029, 
              DZEB012 LIKE DZEB_T.DZEB012, 
              DZEB030 LIKE DZEB_T.DZEB030 
            END RECORD

PUBLIC TYPE T_DZEBL_DIFF_DATA RECORD
              DZHBL001 LIKE DZHBL_T.DZHBL001,
              DZHBL002 LIKE DZHBL_T.DZHBL002,
              DZHBL003 LIKE DZHBL_T.DZHBL003,
              DZHBL004 LIKE DZHBL_T.DZHBL004,
              DZHBL005 LIKE DZHBL_T.DZHBL005
            END RECORD

PUBLIC TYPE T_DZEC_DIFF_DATA RECORD
              DZEC001 LIKE DZEC_T.DZEC001, 
              DZEC002 LIKE DZEC_T.DZEC002, 
              DZEC003 LIKE DZEC_T.DZEC003, 
              DZEC004 LIKE DZEC_T.DZEC004 
            END RECORD

PUBLIC TYPE T_DZED_DIFF_DATA RECORD
              DZED001 LIKE DZED_T.DZED001, 
              DZED002 LIKE DZED_T.DZED002, 
              DZED003 LIKE DZED_T.DZED003, 
              DZED004 LIKE DZED_T.DZED004, 
              DZED005 LIKE DZED_T.DZED005, 
              DZED006 LIKE DZED_T.DZED006 
            END RECORD

PUBLIC TYPE T_DZEN_DIFF_DATA RECORD
              DZEN001 LIKE DZEN_T.DZEN001, 
              DZEN002 LIKE DZEN_T.DZEN002, 
              DZEN003 LIKE DZEN_T.DZEN003, 
              DZEN004 LIKE DZEN_T.DZEN004, 
              DZEN005 LIKE DZEN_T.DZEN005, 
              DZEN006 LIKE DZEN_T.DZEN006,
              DZEN007 LIKE DZEN_T.DZEN007, 
              DZEN008 LIKE DZEN_T.DZEN008, 
              DZEN009 LIKE DZEN_T.DZEN009, 
              DZEN010 LIKE DZEN_T.DZEN010 
            END RECORD

PUBLIC TYPE T_DZHH_T RECORD
              DZHH001 LIKE DZHH_T.DZHH001, 
              DZHH002 LIKE DZHH_T.DZHH002, 
              DZHH003 LIKE DZHH_T.DZHH003, 
              DZHH004 LIKE DZHH_T.DZHH004, 
              DZHH005 LIKE DZHH_T.DZHH005 
            END RECORD
            
PUBLIC TYPE T_KEY_IF_EXCEED RECORD
              kie_result   BOOLEAN,
              kie_sys_def  INTEGER,
              kie_user_def INTEGER
            END RECORD
            
#20161024 begin
TYPE T_SEARCH_CONTENTS RECORD
       sc_contents  VARCHAR(1024),
       sc_sql       VARCHAR(4096)
     END RECORD
#20161024 end

#20161220 begin
TYPE T_COMMENT_COLUMNS RECORD
       CC_OWNER        VARCHAR(100), 
       CC_TABLE_NAME   VARCHAR(100), 
       CC_COLUMN_NAME  VARCHAR(100),
       CC_COMMENTS     VARCHAR(4000)
     END RECORD
#20161220 end

#20170317 begin
PUBLIC TYPE T_TRIGGER_SPOOL_FILE RECORD
              tsf_db          VARCHAR(20),
              tsf_spool_name  VARCHAR(1024)
            END RECORD
#20170317 end
            
#20170912 begin
PUBLIC TYPE T_RESORT_COLUMN RECORD
              rc_TABLE_NAME  VARCHAR(60),
              rc_INDEX_NAME  VARCHAR(60),
              rc_COLUMN_LIST VARCHAR(4096),
              rc_STATUS      VARCHAR(2)
            END RECORD

PUBLIC TYPE T_SESSION_DATA RECORD
              sd_SID           BIGINT,     
              sd_SERIAL_NO     INTEGER,     
              sd_USERNAME      VARCHAR(60), 
              sd_OSUSER        VARCHAR(60), 
              sd_MACHINE       VARCHAR(200),
              sd_PROGRAM       VARCHAR(100),
              sd_PROCESS       VARCHAR(40), 
              sd_SERVICE_NAME  VARCHAR(100) 
            END RECORD            
#20170912 end

#20171120 begin
PUBLIC TYPE T_PICKUP_COLUMN_HEADER RECORD
              pch_PICK_OWNER     VARCHAR(20),
              pch_PICK_DATETIME  VARCHAR(20),
              pch_CUST           VARCHAR(50),
              pch_DGENV          VARCHAR(20),
              pch_HOST_NAME      VARCHAR(256),
              pch_ZONE           VARCHAR(50)              
            END RECORD

PUBLIC TYPE T_KEY_NAME_LIST RECORD 
              knl_TABLE_NAME  VARCHAR(50),
              knl_KEY_NAME    VARCHAR(50),
              knl_KEY_COLS    VARCHAR(1024)  
            END RECORD

PUBLIC TYPE T_INDEX_NAME_LIST RECORD 
              inl_TABLE_NAME  VARCHAR(50),
              inl_INDEX_NAME  VARCHAR(50),
              inl_INDEX_COLS  VARCHAR(1024)  
            END RECORD
                              
PUBLIC TYPE T_COLUMN_INFO RECORD
              cs_TABLE_NAME   VARCHAR(30),
              cs_COLUMN_NAME  VARCHAR(30),
              cs_COLUMN_TYPE  VARCHAR(30),
              cs_SEQUENCE     INTEGER
            END RECORD
                              
#20171120 begin
PUBLIC TYPE T_DZHJ_T RECORD
              DZHJ001 LIKE DZHJ_T.DZHJ001, 
              DZHJ002 LIKE DZHJ_T.DZHJ002, 
              DZHJ003 LIKE DZHJ_T.DZHJ003, 
              DZHJ004 LIKE DZHJ_T.DZHJ004, 
              DZHJ005 LIKE DZHJ_T.DZHJ005 
            END RECORD

PUBLIC TYPE T_CLIENT_PC_INFO RECORD
              cpi_COMPUTERNAME  VARCHAR(100),
              cpi_CLIENTNAME    VARCHAR(100),
              cpi_USERNAME      VARCHAR(100),
              cpi_IP            VARCHAR(100)
            END RECORD
#20180302 begin            

#20180903 begin
PUBLIC TYPE T_COMBO_BOX RECORD
              cb_combobox_name VARCHAR(100),
              cb_combobox_desc VARCHAR(512)
            END RECORD
#20180903 end

#20190304 begin
PUBLIC TYPE T_DZEB027_JSON RECORD
              TOPIND    VARCHAR(10),
              SER_NO    VARCHAR(10),
              COL_TYPE  VARCHAR(10)
            END RECORD
#20190304 end

#20200330 begin
PUBLIC TYPE T_CREATE_SYNONYM_VIEW RECORD
              csv_master_user   VARCHAR(20),  
              csv_table_name    VARCHAR(128)  
            END RECORD
            
PUBLIC TYPE T_SOURCE_OBJECT_INFO RECORD
              soi_schema       VARCHAR(128),
              soi_object_name  VARCHAR(128),
              soi_prc_guid     VARCHAR(128) 
            END RECORD
            
PUBLIC TYPE T_VIEW_DEPEND_TARGET RECORD
              vdt_dependent_schema VARCHAR(128)
            END RECORD        
#20200330 end

#20220831 begin
PUBLIC TYPE T_EXIM_PARA RECORD
              ep_backend_mode    BOOLEAN,
              ep_clone_imp_mode  BOOLEAN  
            END RECORD  
#20220831 end
