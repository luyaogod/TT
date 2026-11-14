{
  異動型態    異動單號       日 期       異動者      異動內容
  ========= ============= ========== ========== ===============================================
  Modify                  20170912   Ernestliou 1.索引匯入時判斷有無已存在 index 欄位, 有則匯入方設失效
  Modify                  20200330   02160      1.修改支援 PosgreSQL(1.1.072)   
  Modify                  20230807   02160      1.修改支援 DM
}  
&include "../4gl/sadzi140_mcr.inc"

IMPORT os
IMPORT util

SCHEMA ds

&include "../4gl/sadzi140_cnst.inc"

&include "../4gl/sadzi140_type.inc"

PRIVATE TYPE T_DIFF_COL_LIST RECORD 
                               RECORD_TYPE            VARCHAR(1),
                               COLUMN_NAME            VARCHAR(30),
                               ISKEY_DB_TABLE         VARCHAR(1), 
                               NULLABLE_DB_TABLE      VARCHAR(1), 
                               DATA_TYPE_DB_TABLE     VARCHAR(80),
                               DATA_LENGTH_DB_TABLE   VARCHAR(40),
                               DEFAULT_VALUE_DB_TABLE VARCHAR(4000),
                               ISKEY_RT_LIST          VARCHAR(1), 
                               NULLABLE_RT_LIST       VARCHAR(1), 
                               DATA_TYPE_RT_LIST      VARCHAR(80),
                               DATA_LENGTH_RT_LIST    VARCHAR(40),
                               DEFAULT_VALUE_RT_LIST  VARCHAR(4000)
                             END RECORD

PRIVATE TYPE T_DIFF_COL_LIST_COLOR RECORD 
                                     RECORD_TYPE            STRING,
                                     COLUMN_NAME            STRING,
                                     ISKEY_DB_TABLE         STRING, 
                                     NULLABLE_DB_TABLE      STRING, 
                                     DATA_TYPE_DB_TABLE     STRING,
                                     DATA_LENGTH_DB_TABLE   STRING,
                                     DEFAULT_VALUE_DB_TABLE STRING,
                                     ISKEY_RT_LIST          STRING, 
                                     NULLABLE_RT_LIST       STRING, 
                                     DATA_TYPE_RT_LIST      STRING,
                                     DATA_LENGTH_RT_LIST    STRING,
                                     DEFAULT_VALUE_RT_LIST  STRING
                                   END RECORD

PRIVATE TYPE T_DIFF_KEY_LIST RECORD
                               RECORD_TYPE      VARCHAR(1),
                               KEY_NAME         VARCHAR(30),
                               KEY_TYPE_DB      VARCHAR(30),
                               COLUMN_NAME_DB   VARCHAR(120),
                               F_TABLE_NAME_DB  VARCHAR(120),
                               F_COLUMN_NAME_DB VARCHAR(500), 
                               KEY_TYPE_RT      VARCHAR(30),
                               COLUMN_NAME_RT   VARCHAR(120),
                               F_TABLE_NAME_RT  VARCHAR(120),
                               F_COLUMN_NAME_RT VARCHAR(500) 
                             END RECORD

PRIVATE TYPE T_DIFF_KEY_LIST_COLOR RECORD 
                                     RECORD_TYPE      STRING,
                                     KEY_NAME         STRING,
                                     KEY_TYPE_DB      STRING,
                                     COLUMN_NAME_DB   STRING,
                                     F_TABLE_NAME_DB  STRING,
                                     F_COLUMN_NAME_DB STRING,
                                     KEY_TYPE_RT      STRING,
                                     COLUMN_NAME_RT   STRING,
                                     F_TABLE_NAME_RT  STRING,
                                     F_COLUMN_NAME_RT STRING
                                   END RECORD

PRIVATE TYPE T_DIFF_IDX_LIST RECORD
                               RECORD_TYPE      VARCHAR(1),
                               INDEX_NAME       VARCHAR(30),
                               INDEX_TYPE_DB    VARCHAR(30),
                               COLUMN_NAME_DB   VARCHAR(120),
                               INDEX_TYPE_RT    VARCHAR(30),
                               COLUMN_NAME_RT   VARCHAR(120)
                             END RECORD

PRIVATE TYPE T_DIFF_IDX_LIST_COLOR RECORD 
                                     RECORD_TYPE      STRING,
                                     INDEX_NAME       STRING,
                                     INDEX_TYPE_DB    STRING,
                                     COLUMN_NAME_DB   STRING,
                                     INDEX_TYPE_RT    STRING,
                                     COLUMN_NAME_RT   STRING
                                   END RECORD

PRIVATE TYPE T_DIFF_LIST_BY_COLUMN_ID RECORD
                                        RECORD_TYPE     VARCHAR(1),
                                        COLUMN_ID       INTEGER, 
                                        DB_COLUMN_NAME  VARCHAR(30),
                                        RT_COLUMN_NAME  VARCHAR(30)
                                      END RECORD 
                                   
PRIVATE TYPE T_DIFF_LIST_BY_COLUMN_ID_COLOR RECORD
                                              RECORD_TYPE     STRING,
                                              COLUMN_ID       STRING, 
                                              DB_COLUMN_NAME  STRING,
                                              RT_COLUMN_NAME  STRING
                                            END RECORD
                                      
&ifndef DEBUG
GLOBALS "../../cfg/top_global.inc"
&endif

DEFINE m_diff_list DYNAMIC ARRAY OF T_DIFF_COL_LIST
DEFINE m_diff_list_color DYNAMIC ARRAY OF T_DIFF_COL_LIST_COLOR

DEFINE m_diff_idx_list DYNAMIC ARRAY OF T_DIFF_IDX_LIST
DEFINE m_diff_idx_list_color DYNAMIC ARRAY OF T_DIFF_IDX_LIST_COLOR

DEFINE m_diff_key_list DYNAMIC ARRAY OF T_DIFF_KEY_LIST
DEFINE m_diff_key_list_color DYNAMIC ARRAY OF T_DIFF_KEY_LIST_COLOR

DEFINE m_diff_list_by_column_id DYNAMIC ARRAY OF T_DIFF_LIST_BY_COLUMN_ID
DEFINE m_diff_list_by_column_id_color DYNAMIC ARRAY OF T_DIFF_LIST_BY_COLUMN_ID_COLOR

DEFINE ms_lang STRING

DEFINE ms_owner_name STRING
DEFINE ms_table_name STRING

DEFINE ms_revision    STRING
DEFINE ms_alm_version STRING

DEFINE ms_view_type STRING 

FUNCTION sadzi140_rdif_run(p_owner_name,p_table_name,p_revision,p_view_type)
DEFINE
  p_owner_name STRING,
  p_table_name STRING,
  p_revision   T_REVISION,
  p_view_type  STRING

  CALL sadzi140_rdif_initialize(p_owner_name,p_table_name,p_revision.*,p_view_type)
  CALL sadzi140_rdif_initial_form()
  CALL sadzi140_rdif_start()
  CALL sadzi140_rdif_finalize()
  
END FUNCTION

FUNCTION sadzi140_rdif_initialize(p_owner_name,p_table_name,p_revision,p_view_type)
DEFINE
  p_owner_name STRING,
  p_table_name STRING,
  p_revision   T_REVISION,
  p_view_type  STRING

  LET ms_owner_name  = p_owner_name
  LET ms_table_name  = p_table_name
  LET ms_alm_version = p_revision.rv_ALM_VERSION
  LET ms_revision    = p_revision.rv_SEQUENCE
  LET ms_view_type   = p_view_type
  
  &ifndef DEBUG
  LET ms_lang = g_lang
  &else
  LET ms_lang = cs_default_lang
  &endif
  
END FUNCTION

FUNCTION sadzi140_rdif_initial_form()
#Being:20150417 by Hiko
DEFINE lw_window   ui.Window,
       lf_form     ui.Form,
       ls_cfg_path STRING,
       ls_4st_path STRING,
       ls_img_path STRING
#End:20150417 by Hiko

  &ifndef DEBUG
    OPEN WINDOW w_sadzi140_rdif WITH FORM cl_ap_formpath("ADZ","sadzi140_rdif")
    #CONNECT TO cs_master_db
  &else
    OPEN WINDOW w_sadzi140_rdif WITH FORM sadzi140_util_get_form_path("sadzi140_rdif")
    #CONNECT TO "local"
  &endif
  
  CURRENT WINDOW IS w_sadzi140_rdif
  
  #Being:20150417 by Hiko
  &ifndef DEBUG
  CALL cl_ui_wintitle(1) #工具抬頭名稱
  CALL cl_load_4ad_interface(NULL)
  &endif

  CALL sadzi140_util_set_form_title('sadzi140_rdif',ms_lang)
  CALL sadzi140_util_set_form_style(ms_lang)
  
  #End:20150417 by Hiko
END FUNCTION

FUNCTION sadzi140_rdif_start()

  DIALOG ATTRIBUTES(UNBUFFERED)
    DISPLAY ARRAY m_diff_list TO sr_table_diff.*
    END DISPLAY

    DISPLAY ARRAY m_diff_key_list TO sr_key_diff.*
    END DISPLAY
    
    DISPLAY ARRAY m_diff_idx_list TO sr_index_diff.*
    END DISPLAY

    DISPLAY ARRAY m_diff_list_by_column_id TO sr_diff_by_column_id.*
    END DISPLAY

    BEFORE DIALOG 
      CALL sadzi140_rdif_fill_table_diff_list()
      CALL DIALOG.setArrayAttributes("sr_table_diff", m_diff_list_color)
      CALL sadzi140_rdif_fill_key_diff_list()
      CALL DIALOG.setArrayAttributes("sr_key_diff", m_diff_key_list_color)
      CALL sadzi140_rdif_fill_index_diff_list()
      CALL DIALOG.setArrayAttributes("sr_index_diff", m_diff_idx_list_color)
      CALL sadzi140_rdif_fill_diff_list_by_column_id()
      CALL DIALOG.setArrayAttributes("sr_diff_by_column_id", m_diff_list_by_column_id_color)

    ON ACTION btn_ok
      EXIT DIALOG
      
    ON ACTION CLOSE
      EXIT DIALOG
      
  END DIALOG

END FUNCTION

FUNCTION sadzi140_rdif_finalize()
  CLOSE WINDOW w_sadzi140_rdif
END FUNCTION

FUNCTION sadzi140_rdif_fill_table_diff_list()
DEFINE
  lo_column_forward_records  DYNAMIC ARRAY OF T_COLUMN_DATA,
  lo_column_backward_records DYNAMIC ARRAY OF T_COLUMN_DATA 
DEFINE 
  ls_sql         STRING,
  ls_ora_coldef  STRING,  #20240529
  ls_db_type     STRING,  #20240606 
  li_index       INTEGER,
  li_count       INTEGER

  IF ms_view_type = cs_view_type_db THEN
    &ifdef is_msv
    LET ls_sql = "select '' record_type,                                                       ",             
                 "       orig.column_name,                                                     ",             
                 "       src_data.iskey         iskey_src_data,                                ",             
                 "       src_data.nullable      nullable_src_data,                             ",             
                 "       src_data.data_type     data_type_src_data,                            ",             
                 "       src_data.data_length   data_length_src_data,                          ",             
                 "       src_data.data_default  default_value_src_data,                        ",             
                 "       tgt_data.iskey         iskey_tgt_data,                                ",             
                 "       tgt_data.nullable      nullable_tgt_data,                             ",             
                 "       tgt_data.data_type     data_type_tgt_data,                            ",             
                 "       tgt_data.data_length   data_length_tgt_data,                          ",             
                 "       tgt_data.data_default  default_value_tgt_data                         ",             
                 "  from (                                                                     ",               
                 "         select cols.column_name, max(column_id) column_id                   ",               
                 "           from (                                                            ",               
                 "                   select atc.column_name    column_name,                    ",               
                 "                          atc.column_id      column_id                       ",               
                 "                     from dba_tab_columns atc                                ",               
                 "                    where 1=1                                                ",               
                 "                      and atc.owner      = '",ms_owner_name.toLowerCase(),"' ",               
                 "                      and atc.table_name = '",ms_table_name.toLowerCase(),"' ",               
                 "                      and atc.column_name <> 'DUMMY'                         ",               
                 "                   union                                                     ",               
                 "                   select ev.dzev005   column_name,                          ",               
                 "                          ev.dzev011   column_id                             ",               
                 "                     from dzev_t ev                                          ",               
                 "                    where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'     ",               
                 "                      and ev.dzev002 = '",ms_table_name.toLowerCase(),"'     ",               
                 "                      and ev.dzev003 = '",ms_revision,"'                     ",               
                 "                      and ev.dzev004 = 'TableDesigner'                       ",               
                 "                      and ev.dzev018 = '",ms_alm_version,"'                  ",               
                 "                ) cols                                                       ",               
                 "          group by cols.column_name                                          ",               
                 "       ) orig                                                                ",    
                 "       left outer join                                                       ",    
                 "       (                                                                     ",               
                 "         select atc.column_name                   column_name,               ",            
                 "                iif(atc.nullable='N','Y','N')     iskey,                     ",    
                 "                atc.nullable                      nullable,                  ",            
                 "                atc.data_type                     data_type,                 ",    
                 "                atc.data_length                   data_length,               ",            
                 "                atc.data_default                  data_default               ",    
                 "           from dba_tab_columns atc                                          ",               
                 "          where 1=1                                                          ",               
                 "            and atc.owner      = '",ms_owner_name.toLowerCase(),"'           ",               
                 "            and atc.table_name = '",ms_table_name.toLowerCase(),"'           ",               
                 "       ) src_data on src_data.column_name = orig.column_name                 ",    
                 "       left outer join                                                       ",    
                 "       (                                                                     ",               
                 "         select ev.dzev005       column_name,                                ",               
                 "                ev.dzev006       iskey,                                      ",               
                 "                ev.dzev007       nullable,                                   ",               
                 "                ev.dzev008       data_type,                                  ",               
                 "                ev.dzev009       data_length,                                ",               
                 "                trim(ev.dzev021) data_default                                ",               
                 "           from dzev_t ev                                                    ",               
                 "          where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'               ",               
                 "            and ev.dzev002 = '",ms_table_name.toLowerCase(),"'               ",               
                 "            and ev.dzev003 = '",ms_revision,"'                               ",               
                 "            and ev.dzev004 = 'TableDesigner'                                 ",               
                 "            and ev.dzev018 = '",ms_alm_version,"'                            ",               
                 "       ) tgt_data on tgt_data.column_name = orig.column_name                 ",                                                         
                 " where 1=1                                                                   ",    
                 " order by orig.column_id                                                     "              
    &else
    &ifdef is_pgs
    LET ls_sql = "select '' record_type,                                                               ",
                 "       orig.column_name,                                                             ",
                 "       src_data.iskey         iskey_src_data,                                        ",
                 "       src_data.nullable      nullable_src_data,                                     ",
                 "       src_data.data_type     data_type_src_data,                                    ",
                 "       src_data.data_length   data_length_src_data,                                  ",
                 "       src_data.data_default  default_value_src_data,                                ",
                 "       tgt_data.iskey         iskey_tgt_data,                                        ",
                 "       tgt_data.nullable      nullable_tgt_data,                                     ",
                 "       tgt_data.data_type     data_type_tgt_data,                                    ",
                 "       tgt_data.data_length   data_length_tgt_data,                                  ",
                 "       tgt_data.data_default  default_value_tgt_data                                 ",
                 "  from (                                                                             ",
                 "         select cols.column_name, max(column_id) column_id                           ",
                 "           from (                                                                    ",
                 "                   select atc.column_name    column_name,                            ",
                 "                          atc.column_id      column_id                               ",
                 "                     from dba_tab_columns atc                                        ",
                 "                    where 1=1                                                        ",
                 "                      and atc.owner      = '",ms_owner_name.toLowerCase(),"'         ",
                 "                      and atc.table_name = '",ms_table_name.toLowerCase(),"'         ",
                 "                      and atc.column_name <> 'dummy'                                 ",
                 "                   union                                                             ",
                 "                   select ev.dzev005   column_name,                                  ",
                 "                          ev.dzev011   column_id                                     ",
                 "                     from dzev_t ev                                                  ",
                 "                    where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'             ",
                 "                      and ev.dzev002 = '",ms_table_name.toLowerCase(),"'             ",
                 "                      and ev.dzev003 = '",ms_revision,"'                             ",
                 "                      and ev.dzev004 = 'TableDesigner'                               ",
                 "                      and ev.dzev018 = '",ms_alm_version,"'                          ",
                 "                ) cols                                                               ",
                 "          group by cols.column_name                                                  ",
                 "       ) orig                                                                        ",
                 "       left outer join                                                               ",
                 "       (                                                                             ",
                 "         select atc.column_name                   column_name,                       ",
                 "                case                                                                 ",
                 "                  when atc.nullable='N' then 'Y'                                     ",
                 "                else 'N' end                      iskey,                             ",
                 "                atc.nullable                      nullable,                          ",
                 "                atc.data_type                     data_type,                         ",
                 "                case                                                                 ",
                 "                  when atc.data_type = 'text' then null                              ",
                 "                  when atc.data_type = 'bytea' then null                             ",
                 "                  when atc.data_type = 'date' then null                              ",
                 "                  when atc.data_type = 'int2' then null                              ",
                 "                  when atc.data_type = 'int4' then null                              ",
                 "                  when atc.data_type = 'int8' then null                              ",
                 "                  when atc.data_type = 'decimal' then                                ",
                 "                    atc.data_precision::varchar||','||                               ",
                 "                    atc.data_scale::varchar                                          ",
                 "                else                                                                 ",
                 "                  atc.data_length::varchar                                           ",
                 "                end::varchar                      data_length,                       ",
                 "                case                                                                 ",
                 "                  when atc.data_default is null then null                            ",
                 "                  when atc.data_default = 'NULL' then null                           ",
                 "                  when atc.data_default = 'null' then null                           ",
                 "                  when atc.data_default like 'CURRENT%' then lower(atc.data_default) ",
                 "                  when atc.data_default = '\"sysdate\"()' then 'sysdate'             ",
                 "                  when atc.data_default = 'SYSDATE' then 'sysdate'                   ",
                 "                else                                                                 ",
                 "                  atc.data_default                                                   ",
                 "                end as data_default                                                  ",
                 "           from dba_tab_columns atc                                                  ",
                 "          where 1=1                                                                  ",
                 "            and atc.owner      = '",ms_owner_name.toLowerCase(),"'                   ",
                 "            and atc.table_name = '",ms_table_name.toLowerCase(),"'                   ",
                 "       ) src_data on src_data.column_name = orig.column_name                         ",
                 "       left outer join                                                               ",
                 "       (                                                                             ",
                 "         select ev.dzev005       column_name,                                        ",
                 "                ev.dzev006       iskey,                                              ",
                 "                ev.dzev007       nullable,                                           ",
                 "                ev.dzev008       data_type,                                          ",
                 "                ev.dzev009       data_length,                                        ",
                 "                trim(ev.dzev021) data_default                                        ",
                 "           from dzev_t ev                                                            ",
                 "          where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'                       ",
                 "            and ev.dzev002 = '",ms_table_name.toLowerCase(),"'                       ",
                 "            and ev.dzev003 = '",ms_revision,"'                                       ",
                 "            and ev.dzev004 = 'TableDesigner'                                         ",
                 "            and ev.dzev018 = '",ms_alm_version,"'                                    ",
                 "       ) tgt_data on tgt_data.column_name = orig.column_name                         ",
                 " where 1=1                                                                           ",
                 " order by orig.column_id                                                             "
    &else
    LET ls_db_type = sadzi140_util_get_db_type()
    CASE ls_db_type
      WHEN cs_db_driver_type_kdb
        LET ls_ora_coldef = " CASE                                                                 ",
                            "   WHEN ATC.DATA_DEFAULT IS NULL THEN NULL                            ",
                            "   WHEN ATC.DATA_DEFAULT = 'NULL' THEN NULL                           ",
                            "   WHEN ATC.DATA_DEFAULT = 'null' THEN NULL                           ",
                            "   WHEN ATC.DATA_DEFAULT LIKE 'CURRENT%' THEN LOWER(ATC.DATA_DEFAULT) ",
                            "   WHEN ATC.DATA_DEFAULT = '\"sysdate\"()' then 'sysdate'             ",
                            "   WHEN ATC.DATA_DEFAULT = 'SYSDATE' then 'sysdate'                   ",
                            " ELSE                                                                 ",
                            "   ATC.DATA_DEFAULT                                                   ",
                            " END AS DATA_DEFAULT                                                  "
      WHEN cs_db_driver_type_dmg 
        LET ls_ora_coldef = " CASE                                                                 ",
                            "   WHEN ATC.DATA_DEFAULT IS NULL THEN NULL                            ",
                            "   WHEN ATC.DATA_DEFAULT = 'NULL' THEN NULL                           ",
                            "   WHEN ATC.DATA_DEFAULT = 'null' THEN NULL                           ",
                            "   WHEN ATC.DATA_DEFAULT LIKE 'CURRENT%' THEN LOWER(ATC.DATA_DEFAULT) ",
                            "   WHEN ATC.DATA_DEFAULT = '\"sysdate\"()' then 'sysdate'             ",
                            "   WHEN ATC.DATA_DEFAULT = 'SYSDATE' then 'sysdate'                   ",
                            " ELSE                                                                 ",
                            "   ATC.DATA_DEFAULT                                                   ",
                            " END AS DATA_DEFAULT                                                  "
    OTHERWISE  
      LET ls_ora_coldef = " DECODE(TRIM(DS.GET_COL_DEFAULT(ATC.TABLE_NAME,ATC.COLUMN_NAME)),'',NULL,          ",
                          "        TRIM(DS.GET_COL_DEFAULT(ATC.TABLE_NAME,ATC.COLUMN_NAME))) DATA_DEFAULT     "
    END CASE
    
    LET ls_sql = "SELECT '' RECORD_TYPE,                                                                         ",
                 "       ORIG.COLUMN_NAME,                                                                       ",
                 "       SRC_DATA.ISKEY         ISKEY_SRC_DATA,                                                  ",
                 "       SRC_DATA.NULLABLE      NULLABLE_SRC_DATA,                                               ",
                 "       SRC_DATA.DATA_TYPE     DATA_TYPE_SRC_DATA,                                              ",
                 "       SRC_DATA.DATA_LENGTH   DATA_LENGTH_SRC_DATA,                                            ",
                 "       SRC_DATA.DATA_DEFAULT  DEFAULT_VALUE_SRC_DATA,                                          ",
                 "       TGT_DATA.ISKEY         ISKEY_TGT_DATA,                                                  ",
                 "       TGT_DATA.NULLABLE      NULLABLE_TGT_DATA,                                               ",
                 "       TGT_DATA.DATA_TYPE     DATA_TYPE_TGT_DATA,                                              ",
                 "       TGT_DATA.DATA_LENGTH   DATA_LENGTH_TGT_DATA,                                            ",
                 "       TGT_DATA.DATA_DEFAULT  DEFAULT_VALUE_TGT_DATA                                           ",
                 "FROM (                                                                                         ",
                 "       SELECT COLS.COLUMN_NAME, MAX(COLUMN_ID) COLUMN_ID                                       ",
                 "         FROM (                                                                                ",
                 "                 SELECT ATC.COLUMN_NAME    COLUMN_NAME,                                        ",
                 "                        ATC.COLUMN_ID      COLUMN_ID                                           ",
                 "                   FROM DBA_TAB_COLUMNS ATC                                                    ",
                 "                  WHERE 1=1                                                                    ",
                 "                    AND ATC.OWNER      = '",ms_owner_name.toUpperCase(),"'                     ",
                 "                    AND ATC.TABLE_NAME = '",ms_table_name.toUpperCase(),"'                     ",
                 "                    AND ATC.COLUMN_NAME <> 'DUMMY'                                             ",
                 "                 UNION                                                                         ",
                 "                 SELECT EV.DZEV005   COLUMN_NAME,                                              ",
                 "                        EV.DZEV011   COLUMN_ID                                                 ",
                 "                   FROM DZEV_T EV                                                              ",        
                 "                  WHERE EV.DZEV001 = '",ms_owner_name.toUpperCase(),"'                         ",
                 "                    AND EV.DZEV002 = '",ms_table_name.toUpperCase(),"'                         ",
                 "                    AND EV.DZEV003 = '",ms_revision,"'                                         ",
                 "                    AND EV.DZEV004 = 'TableDesigner'                                           ",
                 "                    AND EV.DZEV018 = '",ms_alm_version,"'                                      ",
                 "              ) COLS                                                                           ",
                 "        GROUP BY COLS.COLUMN_NAME                                                              ",
                 "     ) ORIG,                                                                                   ",
                 "     (                                                                                         ",
                 "       SELECT ATC.COLUMN_NAME                      COLUMN_NAME,                                ",
                 "              DECODE(ACC.COLUMN_NAME,NULL,'N','Y') ISKEY,                                      ",
                 "              ATC.NULLABLE                         NULLABLE,                                   ",
                 "              DECODE(INSTRB(ATC.DATA_TYPE,'TIMESTAMP'),1,'TIMESTAMP',ATC.DATA_TYPE) DATA_TYPE, ",
                 "              REPLACE(TO_CHAR(DECODE(                                                          ",
                 "                      ATC.DATA_TYPE,                                                           ",
                 "                      'BLOB',4000,                                                             ", #20240529
                 "                      'CLOB',4000,                                                             ", #20240529
                 "                      'INT2',NULL,                                                             ", #20240529
                 "                      'INT4',NULL,                                                             ", #20240529
                 "                      'INT8',NULL,                                                             ", #20240529
                 "                      'NUMBER',DECODE(                                                         ",
                 "                                       NVL(ATC.DATA_SCALE,'0'),                                ",
                 "                                       '0',ATC.DATA_PRECISION,                                 ",
                 "                                       ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE                 ",
                 "                                     ),                                                        ",
                 "                      'NUMERIC',DECODE(                                                        ", #20240529 begin
                 "                                        NVL(ATC.DATA_SCALE,'0'),                               ",
                 "                                        '0',ATC.DATA_PRECISION,                                ",
                 "                                        ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE                ", 
                 "                                      ),                                                       ", #20240529 end
                 "                      'DATE',7,                                                                ", #20240529
                 "                      'TIMESTAMP(0)',7,                                                        ", #20240529
                 "                      'TIMESTAMP(5)',11,                                                       ", #20240529
                 "                      'TIMESTAMP(6)',11,                                                       ", #20240529
                 "                      'TIMESTAMP',DECODE(                                                      ",
                 "                                          ATC.DATA_SCALE,0,7,8,7,11                            ", #20240529
                 "                                        ),                                                     ",                
                 "                      'DATE',DECODE(                                                           ",
                 "                                     ATC.DATA_LENGTH,3,7,ATC.DATA_LENGTH                       ",
                 "                                   ),                                                          ",                
                 "                      ATC.DATA_LENGTH                                                          ",
                 "                    )),'.',',')                    DATA_LENGTH,                                ",
                 ls_ora_coldef,
                 "         FROM DBA_TAB_COLUMNS ATC                                                              ",
                 "         LEFT OUTER JOIN DBA_CONS_COLUMNS ACC ON ACC.OWNER       = ATC.OWNER                   ",
                 "                                             AND ACC.TABLE_NAME  = ATC.TABLE_NAME              ",
                 "                                             AND ACC.COLUMN_NAME = ATC.COLUMN_NAME             ",
                 "                                             AND ACC.CONSTRAINT_NAME NOT LIKE 'SYS%'           ",
                 "                                             AND ACC.CONSTRAINT_NAME LIKE '%PK'                ",
                 "        WHERE 1=1                                                                              ",
                 "          AND ATC.OWNER      = '",ms_owner_name.toUpperCase(),"'                               ",
                 "          AND ATC.TABLE_NAME = '",ms_table_name.toUpperCase(),"'                               ",
                 "     ) SRC_DATA,                                                                               ",
                 "     (                                                                                         ",
                 "       SELECT EV.DZEV005       COLUMN_NAME,                                                    ",
                 "              EV.DZEV006       ISKEY,                                                          ",
                 "              EV.DZEV007       NULLABLE,                                                       ",
                 "              EV.DZEV008       DATA_TYPE,                                                      ",
                 "              EV.DZEV009       DATA_LENGTH,                                                    ",
                 "              TRIM(DECODE(EV.DZEV021,'',NULL,EV.DZEV021)) DATA_DEFAULT                         ",
                 "         FROM DZEV_T EV                                                                        ",
                 "        WHERE EV.DZEV001 = '",ms_owner_name.toUpperCase(),"'                                   ",
                 "          AND EV.DZEV002 = '",ms_table_name.toUpperCase(),"'                                   ",
                 "          AND EV.DZEV003 = '",ms_revision,"'                                                   ",
                 "          AND EV.DZEV004 = 'TableDesigner'                                                     ",
                 "          AND EV.DZEV018 = '",ms_alm_version,"'                                                ",
                 "     ) TGT_DATA                                                                                ",
                 "WHERE TGT_DATA.COLUMN_NAME (+)= ORIG.COLUMN_NAME                                               ",
                 "  AND SRC_DATA.COLUMN_NAME (+)= ORIG.COLUMN_NAME                                               ",
                 "ORDER BY ORIG.COLUMN_ID                                                                        "
  &endif                   
  &endif                   
  ELSE 
    IF ms_view_type = cs_view_type_rt THEN 
      &ifdef is_msv
      LET ls_sql = "select '' record_type,                                                  ",                  
                   "       orig.column_name,                                                ",                  
                   "       src_data.iskey         iskey_src_data,                           ",                  
                   "       src_data.nullable      nullable_src_data,                        ",                  
                   "       src_data.data_type     data_type_src_data,                       ",                  
                   "       src_data.data_length   data_length_src_data,                     ",                  
                   "       src_data.data_default  default_value_src_data,                   ",                  
                   "       tgt_data.iskey         iskey_tgt_data,                           ",                  
                   "       tgt_data.nullable      nullable_tgt_data,                        ",                  
                   "       tgt_data.data_type     data_type_tgt_data,                       ",                  
                   "       tgt_data.data_length   data_length_tgt_data,                     ",                  
                   "       tgt_data.data_default  default_value_tgt_data                    ",                  
                   "  from (                                                                ",                    
                   "        select cols.column_name, max(column_id) column_id               ",                   
                   "          from (                                                        ",                   
                   "                  select eb.dzeb002              column_name,           ",                        
                   "                         coalesce(eb.dzeb021,0)  column_id              ",                        
                   "                    from dzeb_t eb                                      ",                   
                   "                   where eb.dzeb001 = '",ms_table_name.toLowerCase(),"' ",                   
                   "                  union                                                 ",                   
                   "                  select ev.dzev005   column_name,                      ",                   
                   "                         ev.dzev011   column_id                         ",                   
                   "                    from dzev_t ev                                      ",                   
                   "                   where ev.dzev001 = '",ms_owner_name.toLowerCase(),"' ",                   
                   "                    and ev.dzev002  = '",ms_table_name.toLowerCase(),"' ",                  
                   "                     and ev.dzev003 = '",ms_revision,"'                 ",                   
                   "                     and ev.dzev004 = 'TableDesigner'                   ",                   
                   "                     and ev.dzev018 = '",ms_alm_version,"'              ",                   
                   "               ) cols                                                   ",                   
                   "         group by cols.column_name                                      ",                   
                   "       ) orig                                                           ",    
                   "       left outer join                                                  ",    
                   "       (                                                                ",                    
                   "         select eb.dzeb002                             column_name,     ",                    
                   "                eb.dzeb004                             iskey,           ",    
                   "                iif(eb.dzeb004='Y','N','Y')            nullable,        ",    
                   "                td.gztd003                             data_type,       ",                    
                   "                case td.gztd003                                         ",    
                   "                  when 'varbinary'                                      ",    
                   "                    then 'max'                                          ",    
                   "                  when 'nvarchar'                                       ",    
                   "                    then iif(td.gztd009='TEXT','max',td.gztd008)        ",    
                   "                else                                                    ",    
                   "                  td.gztd008                                            ",    
                   "                end                                    data_length,     ",    
                   "                trim(eb.dzeb012)                       data_default     ",                    
                   "           from dzeb_t eb                                               ",                    
                   "                left outer join gztd_t td on td.gztd001 = eb.dzeb006    ",                    
                   "          where eb.dzeb001 = '",ms_table_name.toLowerCase(),"'          ",                    
                   "       ) src_data on src_data.column_name = orig.column_name            ",    
                   "       left outer join                                                  ",    
                   "       (                                                                ",                    
                   "         select ev.dzev005       column_name,                           ",                    
                   "                ev.dzev006       iskey,                                 ",                    
                   "                ev.dzev007       nullable,                              ",                    
                   "                ev.dzev008       data_type,                             ",                    
                   "                ev.dzev009       data_length,                           ",                    
                   "                trim(ev.dzev021) data_default                           ",                    
                   "           from dzev_t ev                                               ",                    
                   "          where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'          ",                    
                   "            and ev.dzev002 = '",ms_table_name.toLowerCase(),"'          ",                    
                   "            and ev.dzev003 = '",ms_revision,"'                          ",                    
                   "            and ev.dzev004 = 'TableDesigner'                            ",                    
                   "            and ev.dzev018 = '",ms_alm_version,"'                       ",                    
                   "       ) tgt_data on tgt_data.column_name = orig.column_name            ",                                                              
                   " where 1=1                                                              ",    
                   " order by orig.column_id                                                "               
      &else
      &ifdef is_pgs
      LET ls_sql = "select '' record_type,                                                  ",
                   "       orig.column_name,                                                ",
                   "       src_data.iskey         iskey_src_data,                           ",
                   "       src_data.nullable      nullable_src_data,                        ",
                   "       src_data.data_type     data_type_src_data,                       ",
                   "       src_data.data_length   data_length_src_data,                     ",
                   "       src_data.data_default  default_value_src_data,                   ",
                   "       tgt_data.iskey         iskey_tgt_data,                           ",
                   "       tgt_data.nullable      nullable_tgt_data,                        ",
                   "       tgt_data.data_type     data_type_tgt_data,                       ",
                   "       tgt_data.data_length   data_length_tgt_data,                     ",
                   "       tgt_data.data_default  default_value_tgt_data                    ",
                   "  from (                                                                ",
                   "        select cols.column_name, max(column_id) column_id               ",
                   "          from (                                                        ",
                   "                  select eb.dzeb002              column_name,           ",
                   "                         coalesce(eb.dzeb021,0)  column_id              ",
                   "                    from dzeb_t eb                                      ",
                   "                   where eb.dzeb001 = '",ms_table_name.toLowerCase(),"' ",
                   "                  union                                                 ",
                   "                  select ev.dzev005   column_name,                      ",
                   "                         ev.dzev011   column_id                         ",
                   "                    from dzev_t ev                                      ",
                   "                   where ev.dzev001 = '",ms_owner_name.toLowerCase(),"' ",
                   "                    and ev.dzev002  = '",ms_table_name.toLowerCase(),"' ",
                   "                     and ev.dzev003 = '",ms_revision,"'                 ",
                   "                     and ev.dzev004 = 'TableDesigner'                   ",
                   "                     and ev.dzev018 = '",ms_alm_version,"'              ",
                   "               ) cols                                                   ",
                   "         group by cols.column_name                                      ",
                   "       ) orig                                                           ",
                   "       left outer join                                                  ",
                   "       (                                                                ",
                   "         select eb.dzeb002                             column_name,     ",
                   "                eb.dzeb004                             iskey,           ",
                   "                case when eb.dzeb004='Y'                                ",
                   "                  then 'N'                                              ",
                   "                else 'Y' end                           nullable,        ",
                   "                td.gztd003                             data_type,       ",
                   "                td.gztd008                             data_length,     ",
                   "                trim(eb.dzeb012)                       data_default     ",
                   "           from dzeb_t eb                                               ",
                   "                left outer join gztd_t td on td.gztd001 = eb.dzeb006    ",
                   "          where eb.dzeb001 = '",ms_table_name.toLowerCase(),"'          ",
                   "       ) src_data on src_data.column_name = orig.column_name            ",
                   "       left outer join                                                  ",
                   "       (                                                                ",
                   "         select ev.dzev005       column_name,                           ",
                   "                ev.dzev006       iskey,                                 ",
                   "                ev.dzev007       nullable,                              ",
                   "                ev.dzev008       data_type,                             ",
                   "                ev.dzev009       data_length,                           ",
                   "                trim(ev.dzev021) data_default                           ",
                   "           from dzev_t ev                                               ",
                   "          where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'          ",
                   "            and ev.dzev002 = '",ms_table_name.toLowerCase(),"'          ",
                   "            and ev.dzev003 = '",ms_revision,"'                          ",
                   "            and ev.dzev004 = 'TableDesigner'                            ",
                   "            and ev.dzev018 = '",ms_alm_version,"'                       ",
                   "       ) tgt_data on tgt_data.column_name = orig.column_name            ",
                   " where 1=1                                                              ",
                   " order by orig.column_id                                                "
      &else
      LET ls_sql = "SELECT '' RECORD_TYPE,                                                                ",
                   "       ORIG.COLUMN_NAME,                                                              ",
                   "       SRC_DATA.ISKEY         ISKEY_SRC_DATA,                                         ",
                   "       SRC_DATA.NULLABLE      NULLABLE_SRC_DATA,                                      ",
                   "       SRC_DATA.DATA_TYPE     DATA_TYPE_SRC_DATA,                                     ",
                   "       SRC_DATA.DATA_LENGTH   DATA_LENGTH_SRC_DATA,                                   ",
                   "       SRC_DATA.DATA_DEFAULT  DEFAULT_VALUE_SRC_DATA,                                 ",
                   "       TGT_DATA.ISKEY         ISKEY_TGT_DATA,                                         ",
                   "       TGT_DATA.NULLABLE      NULLABLE_TGT_DATA,                                      ",
                   "       TGT_DATA.DATA_TYPE     DATA_TYPE_TGT_DATA,                                     ",
                   "       TGT_DATA.DATA_LENGTH   DATA_LENGTH_TGT_DATA,                                   ",
                   "       TGT_DATA.DATA_DEFAULT  DEFAULT_VALUE_TGT_DATA                                  ",
                   "FROM (                                                                                ",
                   "       SELECT COLS.COLUMN_NAME, MAX(COLUMN_ID) COLUMN_ID                              ",
                   "         FROM (                                                                       ",
                   "                 SELECT UPPER(EB.DZEB002)  COLUMN_NAME,                               ",           
                   "                        NVL(EB.DZEB021,0)  COLUMN_ID                                  ",           
                   "                   FROM DZEB_T EB                                                     ",           
                   "                  WHERE EB.DZEB001 = '",ms_table_name.toLowerCase(),"'                ",           
                   "                 UNION                                                                ",
                   "                 SELECT EV.DZEV005   COLUMN_NAME,                                     ",
                   "                        EV.DZEV011   COLUMN_ID                                        ",
                   "                   FROM DZEV_T EV                                                     ",        
                   "                  WHERE EV.DZEV001 = '",ms_owner_name.toUpperCase(),"'                ",
                   "                    AND EV.DZEV002 = '",ms_table_name.toUpperCase(),"'                ",
                   "                    AND EV.DZEV003 = '",ms_revision,"'                                ",
                   "                    AND EV.DZEV004 = 'TableDesigner'                                  ",
                   "                    AND EV.DZEV018 = '",ms_alm_version,"'                             ",
                   "              ) COLS                                                                  ",
                   "        GROUP BY COLS.COLUMN_NAME                                                     ",
                   "     ) ORIG,                                                                          ",
                   "     (                                                                                ",
                   "       SELECT UPPER(EB.DZEB002)                      COLUMN_NAME,                     ", 
                   "              EB.DZEB004                             ISKEY,                           ", 
                   "              DECODE(EB.DZEB004,'Y','N','N','Y','Y') NULLABLE,                        ", 
                   "              UPPER(TD.GZTD003)                      DATA_TYPE,                       ", 
                   "              REPLACE(TO_CHAR(TO_NUMBER(REPLACE(REPLACE(                              ", 
                   "              DECODE(                                                                 ", 
                   "                     UPPER(TD.GZTD003),                                               ", 
                   "                     'BLOB','4000',                                                   ", 
                   "                     'CLOB','4000',                                                   ", 
                   "                     'DATE','7',                                                      ", 
                   "                     'TIMESTAMP',DECODE(TD.GZTD008,'0','7','11'),                     ",
                   "                     TD.GZTD008                                                       ", 
                   "                    ),',','.'),' ',''))),'.',',')    DATA_LENGTH,                     ",
                   "              TRIM(DECODE(EB.DZEB012,'',NULL,EB.DZEB012))   DATA_DEFAULT              ", 
                   "         FROM DZEB_T EB                                                               ", 
                   "              LEFT OUTER JOIN GZTD_T TD ON TD.GZTD001 = EB.DZEB006                    ", 
                   "        WHERE EB.DZEB001 = '",ms_table_name.toLowerCase(),"'                          ",
                   "     ) SRC_DATA,                                                                      ",
                   "     (                                                                                ",
                   "       SELECT EV.DZEV005       COLUMN_NAME,                                           ",
                   "              EV.DZEV006       ISKEY,                                                 ",
                   "              EV.DZEV007       NULLABLE,                                              ",
                   "              EV.DZEV008       DATA_TYPE,                                             ",
                   "              EV.DZEV009       DATA_LENGTH,                                           ",
                   "              TRIM(DECODE(EV.DZEV021,'',NULL,EV.DZEV021)) DATA_DEFAULT                ", 
                   "         FROM DZEV_T EV                                                               ",
                   "        WHERE EV.DZEV001 = '",ms_owner_name.toUpperCase(),"'                          ",
                   "          AND EV.DZEV002 = '",ms_table_name.toUpperCase(),"'                          ",
                   "          AND EV.DZEV003 = '",ms_revision,"'                                          ",
                   "          AND EV.DZEV004 = 'TableDesigner'                                            ",
                   "          AND EV.DZEV018 = '",ms_alm_version,"'                                       ",
                   "     ) TGT_DATA                                                                       ",
                   "WHERE TGT_DATA.COLUMN_NAME (+)= ORIG.COLUMN_NAME                                      ",
                   "  AND SRC_DATA.COLUMN_NAME (+)= ORIG.COLUMN_NAME                                      ",
                   "ORDER BY ORIG.COLUMN_ID                                                               "
      &endif             
      &endif             
    END IF 
  END IF
  
  PREPARE lpre_fill_fill_table_diff_list FROM ls_sql
  DECLARE lcur_fill_fill_table_diff_list CURSOR FOR lpre_fill_fill_table_diff_list

  CALL m_diff_list.clear()
  CALL m_diff_list_color.clear()
  
  LET li_count = 1

  #取得表格差異資料
  CALL sadzi140_rdif_get_column_forward_records(ms_owner_name,ms_table_name) RETURNING lo_column_forward_records
  CALL sadzi140_rdif_get_column_backward_records(ms_owner_name,ms_table_name) RETURNING lo_column_backward_records
  
  #DISPLAY ls_TableName
  FOREACH lcur_fill_fill_table_diff_list INTO m_diff_list[li_count].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET m_diff_list[li_count].RECORD_TYPE = NULL
    
    LET m_diff_list_color[li_count].RECORD_TYPE         = NULL
    LET m_diff_list_color[li_count].COLUMN_NAME         = NULL

    LET m_diff_list_color[li_count].ISKEY_RT_LIST         = NULL
    LET m_diff_list_color[li_count].NULLABLE_RT_LIST      = NULL
    LET m_diff_list_color[li_count].DATA_TYPE_RT_LIST     = NULL
    LET m_diff_list_color[li_count].DATA_LENGTH_RT_LIST   = NULL
    LET m_diff_list_color[li_count].DEFAULT_VALUE_RT_LIST = NULL
    
    LET m_diff_list_color[li_count].ISKEY_DB_TABLE         = NULL
    LET m_diff_list_color[li_count].NULLABLE_DB_TABLE      = NULL
    LET m_diff_list_color[li_count].DATA_TYPE_DB_TABLE     = NULL
    LET m_diff_list_color[li_count].DATA_LENGTH_DB_TABLE   = NULL
    LET m_diff_list_color[li_count].DEFAULT_VALUE_DB_TABLE = NULL

    FOR li_index = 1 TO lo_column_forward_records.getLength()
      IF m_diff_list[li_count].column_name = lo_column_forward_records[li_index].column_name THEN
        LET m_diff_list[li_count].RECORD_TYPE = "!"
        LET m_diff_list_color[li_count].RECORD_TYPE           = "red"
        LET m_diff_list_color[li_count].COLUMN_NAME           = "blue"
        LET m_diff_list_color[li_count].ISKEY_RT_LIST         = "yellow reverse"
        LET m_diff_list_color[li_count].NULLABLE_RT_LIST      = "yellow reverse"
        LET m_diff_list_color[li_count].DATA_TYPE_RT_LIST     = "yellow reverse"
        LET m_diff_list_color[li_count].DATA_LENGTH_RT_LIST   = "yellow reverse"
        LET m_diff_list_color[li_count].DEFAULT_VALUE_RT_LIST = "yellow reverse"
        EXIT FOR;
      END IF    
    END FOR

    FOR li_index = 1 TO lo_column_backward_records.getLength()
      IF m_diff_list[li_count].column_name = lo_column_backward_records[li_index].column_name THEN
        LET m_diff_list[li_count].RECORD_TYPE = "!"
        LET m_diff_list_color[li_count].RECORD_TYPE            = "red"
        LET m_diff_list_color[li_count].COLUMN_NAME            = "blue"
        LET m_diff_list_color[li_count].ISKEY_DB_TABLE         = "red reverse"
        LET m_diff_list_color[li_count].NULLABLE_DB_TABLE      = "red reverse"
        LET m_diff_list_color[li_count].DATA_TYPE_DB_TABLE     = "red reverse"
        LET m_diff_list_color[li_count].DATA_LENGTH_DB_TABLE   = "red reverse"
        LET m_diff_list_color[li_count].DEFAULT_VALUE_DB_TABLE = "red reverse"
        EXIT FOR
      END IF    
    END FOR
    
    LET li_count = li_count + 1
    
  END FOREACH
  #CALL m_diff_list.deleteElement(li_count)

END FUNCTION

FUNCTION sadzi140_rdif_fill_key_diff_list()
DEFINE
  lo_key_forward_records  DYNAMIC ARRAY OF T_KEY_DATA,
  lo_key_backward_records DYNAMIC ARRAY OF T_KEY_DATA 
DEFINE 
  ls_sql   STRING,
  li_key   INTEGER,
  li_count INTEGER

  IF ms_view_type = cs_view_type_db THEN
    &ifdef is_msv
    LET ls_sql = "select ''                            record_type,                                       ",                                  
                 "       orig.constraint_name          constraint_name,                                   ",                                  
                 "       src_key_list.constraint_type  constraint_type_db,                                ",                                  
                 "       src_key_list.column_names     column_names_db,                                   ",                                  
                 "       src_key_list.r_table_name     r_table_name_db,                                   ",                                  
                 "       src_key_list.r_column_names   r_column_names_db,                                 ",                                  
                 "       dst_key_list.constraint_type  constraint_type_rt,                                ",                                  
                 "       dst_key_list.column_names     column_names_rt,                                   ",                                  
                 "       dst_key_list.r_table_name     r_table_name_rt,                                   ",                                  
                 "       dst_key_list.r_column_names   r_column_names_rt                                  ",                                  
                 "  from (                                                                                ",                                  
                 "         select acs.constraint_name                                                     ",                                  
                 "           from dba_constraints acs                                                     ",                                  
                 "          where 1 = 1                                                                   ",                                  
                 "            and acs.table_name = '",ms_table_name.toLowerCase(),"'                      ",                                  
                 "            and acs.owner      = '",ms_owner_name.toLowerCase(),"'                      ",                                  
                 "         union                                                                          ",                                  
                 "         select ew.dzew004 constraint_name                                              ",                           
                 "           from dzew_t ew                                                               ",                                  
                 "          where ew.dzew001 = '",ms_table_name.toLowerCase(),"'                          ",                                  
                 "            and ew.dzew002 = '",ms_revision,"'                                          ",                                  
                 "            and ew.dzew003 = 'Constraint'                                               ",                                  
                 "            and ew.dzew010 = '",ms_alm_version,"'                                       ",                                  
                 "       ) orig                                                                           ",      
                 "       left outer join                                                                  ",      
                 "       (                                                                                ",                                  
                 "         select acs.constraint_name                                    constraint_name, ",                              
                 "                iif(acs.constraint_type='R', 'R', acs.constraint_type) constraint_type, ",                              
                 "                acc.column_names                                       column_names,    ",                              
                 "                NULL                                                   r_table_name,    ",                              
                 "                NULL                                                   r_column_names   ",                              
                 "           from dba_constraints acs                                                     ",                                  
                 "           left outer join (                                                            ",                                  
                 "                            select accs.constraint_name,accs.owner,accs.table_name,     ",                                  
                 "                                   string_agg(accs.column_name, ',')                    ",      
                 "                                     within group(order by accs.position) column_names  ",      
                 "                              from dba_cons_columns accs                                ",                                  
                 "                             group by accs.constraint_name,accs.owner,accs.table_name   ",                                  
                 "                           ) acc on acc.constraint_name = acs.constraint_name           ",                                  
                 "                                and acc.table_name      = acs.table_name                ",                                                                             
                 "                                and acc.owner           = acs.owner                     ",                                  
                 "          where 1 = 1                                                                   ",                                  
                 "            and acs.table_name = '",ms_table_name.toLowerCase(),"'                      ",                                  
                 "            and acs.owner      = '",ms_owner_name.toLowerCase(),"'                      ",                                  
                 "       ) src_key_list on src_key_list.constraint_name = orig.constraint_name            ",      
                 "       left outer join                                                                  ",      
                 "       (                                                                                ",                                  
                 "         select ew.dzew004                 constraint_name,                             ",                           
                 "                ew.dzew005                 constraint_type,                             ",                           
                 "                replace(ew.dzew006,' ','') column_names,                                ",                           
                 "                ew.dzew007                 r_table_name,                                ",                           
                 "                ew.dzew008                 r_column_names                               ",                           
                 "           from dzew_t ew                                                               ",                                  
                 "          where ew.dzew001 = '",ms_table_name.toLowerCase(),"'                          ",                                  
                 "            and ew.dzew002 = '",ms_revision,"'                                          ",                                  
                 "            and ew.dzew003 = 'Constraint'                                               ",                                  
                 "            and ew.dzew010 = '",ms_alm_version,"'                                       ",                                  
                 "       ) dst_key_list on dst_key_list.constraint_name = orig.constraint_name            ",      
                 "     where 1=1                                                                          "      
    &else
    &ifdef is_pgs
    LET ls_sql = "select ''                            record_type,                                     ",  
                 "       orig.constraint_name          constraint_name,                                 ",  
                 "       src_key_list.constraint_type  constraint_type_db,                              ",  
                 "       src_key_list.column_names     column_names_db,                                 ",  
                 "       src_key_list.r_table_name     r_table_name_db,                                 ",  
                 "       src_key_list.r_column_names   r_column_names_db,                               ",  
                 "       dst_key_list.constraint_type  constraint_type_rt,                              ",  
                 "       dst_key_list.column_names     column_names_rt,                                 ",  
                 "       dst_key_list.r_table_name     r_table_name_rt,                                 ",  
                 "       dst_key_list.r_column_names   r_column_names_rt                                ",  
                 "  from (                                                                              ",  
                 "         select acs.constraint_name                                                   ",  
                 "           from dba_constraints acs                                                   ",  
                 "          where 1 = 1                                                                 ",  
                 "            and acs.table_name = '",ms_table_name.toLowerCase(),"'                    ",  
                 "            and acs.owner      = '",ms_owner_name.toLowerCase(),"'                    ",  
                 "         union                                                                        ",  
                 "         select ew.dzew004 constraint_name                                            ",  
                 "           from dzew_t ew                                                             ",  
                 "          where ew.dzew001 = '",ms_table_name.toLowerCase(),"'                        ",  
                 "            and ew.dzew002 = '",ms_revision,"'                                        ",  
                 "            and ew.dzew003 = 'Constraint'                                             ",  
                 "            and ew.dzew010 = '",ms_alm_version,"'                                     ",  
                 "       ) orig                                                                         ",  
                 "       left outer join                                                                ",  
                 "       (                                                                              ",  
                 "         select acs.constraint_name                        constraint_name,           ",
                 "                case                                                                  ",
                 "                  when acs.constraint_type='R' then 'R'                               ",
                 "                else acs.constraint_type end               constraint_type,           ",
                 "                acc.column_names                           column_names,              ",
                 "                NULL                                       r_table_name,              ",
                 "                NULL                                       r_column_names             ",
                 "           from dba_constraints acs                                                   ",  
                 "           left outer join (                                                          ",  
                 "                            select accs.constraint_name,accs.owner,accs.table_name,   ",  
                 "                                   string_agg(accs.column_name, ','                   ",
                 "                                     order by accs.position) column_names             ",
                 "                              from dba_cons_columns accs                              ",  
                 "                             group by accs.constraint_name,accs.owner,accs.table_name ",  
                 "                           ) acc on acc.constraint_name = acs.constraint_name         ",  
                 "                                and acc.table_name      = acs.table_name              ",                                                                             
                 "                                and acc.owner           = acs.owner                   ",  
                 "          where 1 = 1                                                                 ",  
                 "            and acs.table_name = '",ms_table_name.toLowerCase(),"'                    ",  
                 "            and acs.owner      = '",ms_owner_name.toLowerCase(),"'                    ",  
                 "       ) src_key_list on src_key_list.constraint_name = orig.constraint_name          ",  
                 "       left outer join                                                                ",  
                 "       (                                                                              ",  
                 "         select ew.dzew004                 constraint_name,                           ",  
                 "                ew.dzew005                 constraint_type,                           ",  
                 "                replace(ew.dzew006,' ','') column_names,                              ",  
                 "                ew.dzew007                 r_table_name,                              ",  
                 "                ew.dzew008                 r_column_names                             ",  
                 "           from dzew_t ew                                                             ",  
                 "          where ew.dzew001 = '",ms_table_name.toLowerCase(),"'                        ",  
                 "            and ew.dzew002 = '",ms_revision,"'                                        ",  
                 "            and ew.dzew003 = 'Constraint'                                             ",  
                 "            and ew.dzew010 = '",ms_alm_version,"'                                     ",  
                 "       ) dst_key_list on dst_key_list.constraint_name = orig.constraint_name          ",  
                 "     where 1=1                                                                        "
    &else
    LET ls_sql = "SELECT ''                            RECORD_TYPE,                                                                   ",
                 "       ORIG.CONSTRAINT_NAME          CONSTRAINT_NAME,                                                               ",
                 "       SRC_KEY_LIST.CONSTRAINT_TYPE  CONSTRAINT_TYPE_DB,                                                            ",
                 "       SRC_KEY_LIST.COLUMN_NAMES     COLUMN_NAMES_DB,                                                               ",
                 "       SRC_KEY_LIST.R_TABLE_NAME     R_TABLE_NAME_DB,                                                               ",
                 "       SRC_KEY_LIST.R_COLUMN_NAMES   R_COLUMN_NAMES_DB,                                                             ",
                 "       DST_KEY_LIST.CONSTRAINT_TYPE  CONSTRAINT_TYPE_RT,                                                            ",
                 "       DST_KEY_LIST.COLUMN_NAMES     COLUMN_NAMES_RT,                                                               ",
                 "       DST_KEY_LIST.R_TABLE_NAME     R_TABLE_NAME_RT,                                                               ",
                 "       DST_KEY_LIST.R_COLUMN_NAMES   R_COLUMN_NAMES_RT                                                              ",
                 "  FROM (                                                                                                            ",
                 "         SELECT ACS.CONSTRAINT_NAME                                                                                 ",                                                  
                 "           FROM DBA_CONSTRAINTS ACS                                                                                 ",                                                 
                 "          WHERE 1 = 1                                                                                               ",                                                 
                 "            AND ACS.TABLE_NAME = '",ms_table_name.toUpperCase(),"'                                                  ",                             
                 "            AND ACS.OWNER      = '",ms_owner_name.toUpperCase(),"'                                                  ",
                 "         UNION                                                                                                      ",                                                 
                 "         SELECT UPPER(EW.DZEW004) CONSTRAINT_NAME                                                                   ",
                 "           FROM DZEW_T EW                                                                                           ",
                 "          WHERE EW.DZEW001 = '",ms_table_name.toLowerCase(),"'                                                      ",
                 "            AND EW.DZEW002 = '",ms_revision,"'                                                                      ",
                 "            AND EW.DZEW003 = 'Constraint'                                                                           ",
                 "            AND EW.DZEW010 = '",ms_alm_version,"'                                                                   ",
                 "       ) ORIG,                                                                                                      ",     
                 "       (                                                                                                            ",
                 "         SELECT ACS.CONSTRAINT_NAME                                        CONSTRAINT_NAME,                         ",                                                                                                        
                 "                DECODE(ACS.CONSTRAINT_TYPE, 'R', 'R', ACS.CONSTRAINT_TYPE) CONSTRAINT_TYPE,                         ",                                                 
                 "                ACC.COLUMN_NAMES                                           COLUMN_NAMES,                            ",                                                                                                        
                 "                ACCR.TABLE_NAME                                            R_TABLE_NAME,                            ",                                                                                            
                 "                ACCR.COLUMN_NAMES                                          R_COLUMN_NAMES                           ",                                                                              
                 "           FROM DBA_CONSTRAINTS ACS                                                                                 ",                                                 
                 "           LEFT OUTER JOIN (                                                                                        ",                                                 
                 "                            SELECT ACCS.CONSTRAINT_NAME,ACCS.OWNER,ACCS.TABLE_NAME,                                 ",
                 "                                   LISTAGG(ACCS.COLUMN_NAME, ',') WITHIN GROUP(ORDER BY ACCS.POSITION) COLUMN_NAMES ", 
                 "                              FROM DBA_CONS_COLUMNS ACCS                                                            ",                                                 
                 "                             GROUP BY ACCS.CONSTRAINT_NAME,ACCS.OWNER,ACCS.TABLE_NAME                               ",                                                 
                 "                           ) ACC ON ACC.CONSTRAINT_NAME = ACS.CONSTRAINT_NAME                                       ",                                                 
                 "                                AND ACC.TABLE_NAME      = ACS.TABLE_NAME                                            ",                                                                             
                 "                                AND ACC.OWNER           = ACS.OWNER                                                 ",                                                 
                 "           LEFT OUTER JOIN (                                                                                        ",                                                 
                 "                            SELECT ACCS.CONSTRAINT_NAME,ACCS.OWNER,ACCS.TABLE_NAME,                                 ",
                 "                                   LISTAGG(ACCS.COLUMN_NAME, ',') WITHIN GROUP(ORDER BY ACCS.POSITION) COLUMN_NAMES ", 
                 "                              FROM DBA_CONS_COLUMNS ACCS                                                            ",                                                 
                 "                             GROUP BY ACCS.CONSTRAINT_NAME,ACCS.OWNER,ACCS.TABLE_NAME                               ",                                                 
                 "                           ) ACCR ON ACCR.CONSTRAINT_NAME = ACS.R_CONSTRAINT_NAME                                   ",                                                 
                 "                                 AND ACCR.OWNER           = ACS.R_OWNER                                             ",                                                 
                 "          WHERE 1 = 1                                                                                               ",                                                 
                 "            AND ACS.TABLE_NAME = '",ms_table_name.toUpperCase(),"'                                                  ",                                                 
                 "            AND ACS.OWNER      = '",ms_owner_name.toUpperCase(),"'                                                  ",                                                 
                 "       ) SRC_KEY_LIST,                                                                                              ",
                 "       (                                                                                                            ",
                 "         SELECT UPPER(EW.DZEW004)                 CONSTRAINT_NAME,                                                  ",
                 "                UPPER(EW.DZEW005)                 CONSTRAINT_TYPE,                                                  ",
                 "                UPPER(REPLACE(EW.DZEW006,' ','')) COLUMN_NAMES,                                                     ",
                 "                UPPER(EW.DZEW007)                 R_TABLE_NAME,                                                     ",                                                                                            
                 "                UPPER(EW.DZEW008)                 R_COLUMN_NAMES                                                    ",                                                                              
                 "           FROM DZEW_T EW                                                                                           ",
                 "          WHERE EW.DZEW001 = '",ms_table_name.toLowerCase(),"'                                                      ",
                 "            AND EW.DZEW002 = '",ms_revision,"'                                                                      ",
                 "            AND EW.DZEW003 = 'Constraint'                                                                           ",
                 "            AND EW.DZEW010 = '",ms_alm_version,"'                                                                   ",
                 "       ) DST_KEY_LIST                                                                                               ",
                 "     WHERE SRC_KEY_LIST.CONSTRAINT_NAME (+)= ORIG.CONSTRAINT_NAME                                                   ",
                 "       AND DST_KEY_LIST.CONSTRAINT_NAME (+)= ORIG.CONSTRAINT_NAME                                                   "
    &endif                  
    &endif                  
  ELSE 
    IF ms_view_type = cs_view_type_rt THEN
      &ifdef is_msv
      LET ls_sql = "select ''                            record_type,                             ",                                      
                   "       orig.constraint_name          constraint_name,                         ",                                      
                   "       src_key_list.constraint_type  constraint_type_db,                      ",                                      
                   "       src_key_list.column_names     column_names_db,                         ",                                      
                   "       src_key_list.r_table_name     r_table_name_db,                         ",                                      
                   "       src_key_list.r_column_names   r_column_names_db,                       ",                                      
                   "       dst_key_list.constraint_type  constraint_type_rt,                      ",                                      
                   "       dst_key_list.column_names     column_names_rt,                         ",                                      
                   "       dst_key_list.r_table_name     r_table_name_rt,                         ",                                      
                   "       dst_key_list.r_column_names   r_column_names_rt                        ",                                      
                   "  from (                                                                      ",                                      
                   "         select ed.dzed002 constraint_name                                    ",                               
                   "           from dzed_t ed                                                     ",                                      
                   "          where ed.dzed001 = '",ms_table_name.toLowerCase(),"'                ",                                      
                   "            and ed.dzed003 <> 'F'                                             ",                                      
                   "         union                                                                ",                                      
                   "         select ew.dzew004 constraint_name                                    ",                               
                   "           from dzew_t ew                                                     ",                                      
                   "          where ew.dzew001 = '",ms_table_name.toLowerCase(),"'                ",                                      
                   "            and ew.dzew002 = '",ms_revision,"'                                ",                                      
                   "            and ew.dzew003 = 'Constraint'                                     ",                                      
                   "            and ew.dzew010 = '",ms_alm_version,"'                             ",                                      
                   "       ) orig                                                                 ",
                   "       left outer join                                                        ",
                   "       (                                                                      ",                                      
                   "         select ed.dzed002       constraint_name,                             ",                               
                   "                ed.dzed003       constraint_type,                             ",                               
                   "                ed.dzed004       column_names,                                ",                               
                   "                trim(ed.dzed005) r_table_name,                                ",                               
                   "                trim(ed.dzed006) r_column_names                               ",                               
                   "           from dzed_t ed                                                     ",                                      
                   "          where ed.dzed001 = '",ms_table_name.toLowerCase(),"'                ",                                      
                   "            and ed.dzed003 <> 'F'                                             ",                                      
                   "       ) src_key_list on src_key_list.constraint_name = orig.constraint_name  ",        
                   "       left outer join                                                        ",
                   "       (                                                                      ",                                      
                   "         select ew.dzew004                 constraint_name,                   ",                               
                   "                ew.dzew005                 constraint_type,                   ",                               
                   "                replace(ew.dzew006,' ','') column_names,                      ",                               
                   "                ew.dzew007                 r_table_name,                      ",                               
                   "                ew.dzew008                 r_column_names                     ",                               
                   "           from dzew_t ew                                                     ",                                      
                   "          where ew.dzew001 = '",ms_table_name.toLowerCase(),"'                ",                                      
                   "            and ew.dzew002 = '",ms_revision,"'                                ",                                      
                   "            and ew.dzew003 = 'Constraint'                                     ",                                      
                   "            and ew.dzew010 = '",ms_alm_version,"'                             ",                                      
                   "       ) dst_key_list on dst_key_list.constraint_name = orig.constraint_name  ",                                                                                              
                   " where 1=1                                                                    "
      &else
      &ifdef is_pgs
      LET ls_sql = "select ''                            record_type,                             ",                                      
                   "       orig.constraint_name          constraint_name,                         ",                                      
                   "       src_key_list.constraint_type  constraint_type_db,                      ",                                      
                   "       src_key_list.column_names     column_names_db,                         ",                                      
                   "       src_key_list.r_table_name     r_table_name_db,                         ",                                      
                   "       src_key_list.r_column_names   r_column_names_db,                       ",                                      
                   "       dst_key_list.constraint_type  constraint_type_rt,                      ",                                      
                   "       dst_key_list.column_names     column_names_rt,                         ",                                      
                   "       dst_key_list.r_table_name     r_table_name_rt,                         ",                                      
                   "       dst_key_list.r_column_names   r_column_names_rt                        ",                                      
                   "  from (                                                                      ",                                      
                   "         select ed.dzed002 constraint_name                                    ",                               
                   "           from dzed_t ed                                                     ",                                      
                   "          where ed.dzed001 = '",ms_table_name.toLowerCase(),"'                ",                                      
                   "            and ed.dzed003 <> 'F'                                             ",                                      
                   "         union                                                                ",                                      
                   "         select ew.dzew004 constraint_name                                    ",                               
                   "           from dzew_t ew                                                     ",                                      
                   "          where ew.dzew001 = '",ms_table_name.toLowerCase(),"'                ",                                      
                   "            and ew.dzew002 = '",ms_revision,"'                                ",                                      
                   "            and ew.dzew003 = 'Constraint'                                     ",                                      
                   "            and ew.dzew010 = '",ms_alm_version,"'                             ",                                      
                   "       ) orig                                                                 ",
                   "       left outer join                                                        ",
                   "       (                                                                      ",                                      
                   "         select ed.dzed002       constraint_name,                             ",                               
                   "                ed.dzed003       constraint_type,                             ",                               
                   "                ed.dzed004       column_names,                                ",                               
                   "                trim(ed.dzed005) r_table_name,                                ",                               
                   "                trim(ed.dzed006) r_column_names                               ",                               
                   "           from dzed_t ed                                                     ",                                      
                   "          where ed.dzed001 = '",ms_table_name.toLowerCase(),"'                ",                                      
                   "            and ed.dzed003 <> 'F'                                             ",                                      
                   "       ) src_key_list on src_key_list.constraint_name = orig.constraint_name  ",        
                   "       left outer join                                                        ",
                   "       (                                                                      ",                                      
                   "         select ew.dzew004                 constraint_name,                   ",                               
                   "                ew.dzew005                 constraint_type,                   ",                               
                   "                replace(ew.dzew006,' ','') column_names,                      ",                               
                   "                ew.dzew007                 r_table_name,                      ",                               
                   "                ew.dzew008                 r_column_names                     ",                               
                   "           from dzew_t ew                                                     ",                                      
                   "          where ew.dzew001 = '",ms_table_name.toLowerCase(),"'                ",                                      
                   "            and ew.dzew002 = '",ms_revision,"'                                ",                                      
                   "            and ew.dzew003 = 'Constraint'                                     ",                                      
                   "            and ew.dzew010 = '",ms_alm_version,"'                             ",                                      
                   "       ) dst_key_list on dst_key_list.constraint_name = orig.constraint_name  ",                                                                                              
                   " where 1=1                                                                    "
      &else
      LET ls_sql = "SELECT ''                            RECORD_TYPE,                                                                   ",
                   "       ORIG.CONSTRAINT_NAME          CONSTRAINT_NAME,                                                               ",
                   "       SRC_KEY_LIST.CONSTRAINT_TYPE  CONSTRAINT_TYPE_DB,                                                            ",
                   "       SRC_KEY_LIST.COLUMN_NAMES     COLUMN_NAMES_DB,                                                               ",
                   "       SRC_KEY_LIST.R_TABLE_NAME     R_TABLE_NAME_DB,                                                               ",
                   "       SRC_KEY_LIST.R_COLUMN_NAMES   R_COLUMN_NAMES_DB,                                                             ",
                   "       DST_KEY_LIST.CONSTRAINT_TYPE  CONSTRAINT_TYPE_RT,                                                            ",
                   "       DST_KEY_LIST.COLUMN_NAMES     COLUMN_NAMES_RT,                                                               ",
                   "       DST_KEY_LIST.R_TABLE_NAME     R_TABLE_NAME_RT,                                                               ",
                   "       DST_KEY_LIST.R_COLUMN_NAMES   R_COLUMN_NAMES_RT                                                              ",
                   "  FROM (                                                                                                            ",
                   "         SELECT UPPER(ED.DZED002) CONSTRAINT_NAME                                                                   ",                                           
                   "           FROM DZED_T ED                                                                                           ",                                                 
                   "          WHERE ED.DZED001 = '",ms_table_name.toLowerCase(),"'                                                      ",
                   "            AND ED.DZED003 <> 'F'                                                                                   ",               
                   "         UNION                                                                                                      ",                                                 
                   "         SELECT UPPER(EW.DZEW004) CONSTRAINT_NAME                                                                   ",
                   "           FROM DZEW_T EW                                                                                           ",
                   "          WHERE EW.DZEW001 = '",ms_table_name.toLowerCase(),"'                                                      ",
                   "            AND EW.DZEW002 = '",ms_revision,"'                                                                      ",
                   "            AND EW.DZEW003 = 'Constraint'                                                                           ",
                   "            AND EW.DZEW010 = '",ms_alm_version,"'                                                                   ",
                   "       ) ORIG,                                                                                                      ",     
                   "       (                                                                                                            ",
                   "         SELECT UPPER(ED.DZED002)       CONSTRAINT_NAME,                                                            ",
                   "                UPPER(ED.DZED003)       CONSTRAINT_TYPE,                                                            ",
                   "                UPPER(ED.DZED004)       COLUMN_NAMES,                                                               ",
                   "                UPPER(TRIM(ED.DZED005)) R_TABLE_NAME,                                                               ",
                   "                UPPER(TRIM(ED.DZED006)) R_COLUMN_NAMES                                                              ",
                   "           FROM DZED_T ED                                                                                           ",
                   "          WHERE ED.DZED001 = '",ms_table_name.toLowerCase(),"'                                                      ",
                   "            AND ED.DZED003 <> 'F'                                                                                   ",               
                   "       ) SRC_KEY_LIST,                                                                                              ",
                   "       (                                                                                                            ",
                   "         SELECT UPPER(EW.DZEW004)                 CONSTRAINT_NAME,                                                  ",
                   "                UPPER(EW.DZEW005)                 CONSTRAINT_TYPE,                                                  ",
                   "                UPPER(REPLACE(EW.DZEW006,' ','')) COLUMN_NAMES,                                                     ",
                   "                UPPER(EW.DZEW007)                 R_TABLE_NAME,                                                     ",                                                                                            
                   "                UPPER(EW.DZEW008)                 R_COLUMN_NAMES                                                    ",                                                                              
                   "           FROM DZEW_T EW                                                                                           ",
                   "          WHERE EW.DZEW001 = '",ms_table_name.toLowerCase(),"'                                                      ",
                   "            AND EW.DZEW002 = '",ms_revision,"'                                                                      ",
                   "            AND EW.DZEW003 = 'Constraint'                                                                           ",
                   "            AND EW.DZEW010 = '",ms_alm_version,"'                                                                   ",
                   "       ) DST_KEY_LIST                                                                                               ",
                   "     WHERE SRC_KEY_LIST.CONSTRAINT_NAME (+)= ORIG.CONSTRAINT_NAME                                                   ",
                   "       AND DST_KEY_LIST.CONSTRAINT_NAME (+)= ORIG.CONSTRAINT_NAME                                                   "
      &endif                   
      &endif                   
    END IF 
  END IF 
  
  PREPARE lpre_fill_fill_key_diff_list FROM ls_sql
  DECLARE lcur_fill_fill_key_diff_list CURSOR FOR lpre_fill_fill_key_diff_list

  CALL m_diff_key_list.clear()
  CALL m_diff_key_list_color.clear()
  
  LET li_count = 1

  #取得KEY差異資料
  CALL sadzi140_rdif_get_key_forward_records(ms_owner_name,ms_table_name) RETURNING lo_key_forward_records
  CALL sadzi140_rdif_get_key_backward_records(ms_owner_name,ms_table_name) RETURNING lo_key_backward_records
  
  #DISPLAY ls_TableName
  FOREACH lcur_fill_fill_key_diff_list INTO m_diff_key_list[li_count].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET m_diff_key_list[li_count].RECORD_TYPE = NULL
    
    LET m_diff_key_list_color[li_count].RECORD_TYPE      = NULL
    LET m_diff_key_list_color[li_count].KEY_NAME         = NULL
    LET m_diff_key_list_color[li_count].KEY_TYPE_DB      = NULL
    LET m_diff_key_list_color[li_count].COLUMN_NAME_DB   = NULL
    LET m_diff_key_list_color[li_count].F_TABLE_NAME_DB  = NULL
    LET m_diff_key_list_color[li_count].F_COLUMN_NAME_DB = NULL 
    LET m_diff_key_list_color[li_count].KEY_TYPE_RT      = NULL
    LET m_diff_key_list_color[li_count].COLUMN_NAME_RT   = NULL
    LET m_diff_key_list_color[li_count].F_TABLE_NAME_RT  = NULL
    LET m_diff_key_list_color[li_count].F_COLUMN_NAME_RT = NULL 

    FOR li_key = 1 TO lo_key_forward_records.getLength()
      IF m_diff_key_list[li_count].key_name = lo_key_forward_records[li_key].key_name THEN
        LET m_diff_key_list[li_count].RECORD_TYPE = "!"
        LET m_diff_key_list_color[li_count].RECORD_TYPE      = "red"
        LET m_diff_key_list_color[li_count].KEY_NAME         = "blue"
        LET m_diff_key_list_color[li_count].KEY_TYPE_RT      = "yellow reverse"
        LET m_diff_key_list_color[li_count].COLUMN_NAME_RT   = "yellow reverse"
        LET m_diff_key_list_color[li_count].F_TABLE_NAME_RT  = "yellow reverse"
        LET m_diff_key_list_color[li_count].F_COLUMN_NAME_RT = "yellow reverse" 
        EXIT FOR;
      END IF    
    END FOR

    FOR li_key = 1 TO lo_key_backward_records.getLength()
      IF m_diff_key_list[li_count].key_name = lo_key_backward_records[li_key].key_name THEN
        LET m_diff_key_list[li_count].RECORD_TYPE = "!"
        LET m_diff_key_list_color[li_count].RECORD_TYPE      = "red"
        LET m_diff_key_list_color[li_count].KEY_NAME         = "blue"
        LET m_diff_key_list_color[li_count].KEY_TYPE_DB      = "red reverse"
        LET m_diff_key_list_color[li_count].COLUMN_NAME_DB   = "red reverse"
        LET m_diff_key_list_color[li_count].F_TABLE_NAME_DB  = "red reverse"
        LET m_diff_key_list_color[li_count].F_COLUMN_NAME_DB = "red reverse" 
        EXIT FOR
      END IF    
    END FOR
    
    LET li_count = li_count + 1
    
  END FOREACH
  #CALL m_diff_key_list.deleteElement(li_count)

END FUNCTION

FUNCTION sadzi140_rdif_get_key_forward_records(p_owner,p_table_name)
DEFINE
  p_owner      STRING,
  p_table_name STRING
DEFINE
  lo_key_forward_records DYNAMIC ARRAY OF T_KEY_DATA
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_key_name    STRING,
  ls_sql         STRING,
  ls_exec_sql    STRING,
  li_rec_cnt     INTEGER,
  ls_alter_table VARCHAR(30)
DEFINE
  lo_return DYNAMIC ARRAY OF T_KEY_DATA  

  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name
  
  LET li_rec_cnt = 1

  ------------------------------------------------------------------------------------------------
  -- key forward
  ------------------------------------------------------------------------------------------------
  IF ms_view_type = cs_view_type_db THEN
    &ifdef is_msv
    LET ls_sql = "select ew.dzew001                 table_name,                                 ",                                                                      
                 "       ew.dzew004                 constraint_name,                            ",                                                                      
                 "       ew.dzew005                 constraint_type,                            ",                                                                      
                 "       replace(ew.dzew006,' ','') column_names,                               ",                                                                      
                 "       trim(ew.dzew007)           r_table_name,                               ",                                                                      
                 "       trim(ew.dzew008)           r_column_names                              ",                                                                      
                 "  from dzew_t ew                                                              ",                                                                             
                 " where ew.dzew001 = '",ms_table_name.toLowerCase(),"'                         ",                                                                             
                 "   and ew.dzew002 = '",ms_revision,"'                                         ",                                                                             
                 "   and ew.dzew003 = 'Constraint'                                              ",                                                                             
                 "   and ew.dzew010 = '",ms_alm_version,"'                                      ",                                                                             
                 "except                                                                        ",                                                                              
                 "select acs.table_name,                                                        ",                                                                             
                 "       acs.constraint_name,                                                   ",                                                                             
                 "       iif(acs.constraint_type='R','R',acs.constraint_type) constraint_type,  ",                                                                       
                 "       acc.column_names,                                                      ",                                                                             
                 "       NULL  r_table_name,                                                    ",                                                                   
                 "       NULL  r_column_name                                                    ",                                                                              
                 "  from dba_constraints acs                                                    ",                                                                             
                 "  left outer join (                                                           ",                                                                             
                 "                   select accs.constraint_name,accs.owner,accs.table_name,    ",
                 "                          string_agg(accs.column_name, ',')                   ",
                 "                            within group(order by accs.position) column_names ",
                 "                     from dba_cons_columns accs                               ",                                                                             
                 "                    group by accs.constraint_name,accs.owner,accs.table_name  ",                                                                             
                 "                  ) acc on acc.constraint_name = acs.constraint_name          ",                                                                             
                 "                       and acc.table_name      = acs.table_name               ",                                                                             
                 "                       and acc.owner           = acs.owner                    ",                                                                             
                 " where 1 = 1                                                                  ",                                                                             
                 "   and acs.table_name = '",ls_table_name.toLowerCase(),"'                     ",                                                                             
                 "   and acs.owner      = '",ls_owner.toLowerCase(),"'                          "                                                                             
    &else
    &ifdef is_pgs
    LET ls_sql = "select ew.dzew001                 table_name,                                 ", 
                 "       ew.dzew004                 constraint_name,                            ", 
                 "       ew.dzew005                 constraint_type,                            ", 
                 "       replace(ew.dzew006,' ','') column_names,                               ", 
                 "       trim(ew.dzew007)           r_table_name,                               ", 
                 "       trim(ew.dzew008)           r_column_names                              ", 
                 "  from dzew_t ew                                                              ", 
                 " where ew.dzew001 = '",ms_table_name.toLowerCase(),"'                         ", 
                 "   and ew.dzew002 = '",ms_revision,"'                                         ", 
                 "   and ew.dzew003 = 'Constraint'                                              ", 
                 "   and ew.dzew010 = '",ms_alm_version,"'                                      ", 
                 "except                                                                        ", 
                 "select acs.table_name,                                                        ", 
                 "       acs.constraint_name,                                                   ", 
                 "       case                                                                   ", 
                 "         when acs.constraint_type='R'                                         ", 
                 "         then 'R' else acs.constraint_type                                    ", 
                 "       end constraint_type,                                                   ", 
                 "       acc.column_names,                                                      ", 
                 "       NULL  r_table_name,                                                    ", 
                 "       NULL  r_column_name                                                    ", 
                 "  from dba_constraints acs                                                    ", 
                 "  left outer join (                                                           ", 
                 "                   select accs.constraint_name,accs.owner,accs.table_name,    ", 
                 "                          string_agg(accs.column_name, ','                    ", 
                 "                            order by accs.position) column_names              ", 
                 "                     from dba_cons_columns accs                               ", 
                 "                    group by accs.constraint_name,accs.owner,accs.table_name  ", 
                 "                  ) acc on acc.constraint_name = acs.constraint_name          ", 
                 "                       and acc.table_name      = acs.table_name               ",                                                                             
                 "                       and acc.owner           = acs.owner                    ",
                 " where 1 = 1                                                                  ",
                 "   and acs.table_name = '",ls_table_name.toLowerCase(),"'                     ",
                 "   and acs.owner      = '",ls_owner.toLowerCase(),"'                          "
    &else
    LET ls_sql = "SELECT UPPER(EW.DZEW001)                 TABLE_NAME,                                                                                                        ",
                 "       UPPER(EW.DZEW004)                 CONSTRAINT_NAME,                                                                                                   ",
                 "       UPPER(EW.DZEW005)                 CONSTRAINT_TYPE,                                                                                                   ",
                 "       UPPER(REPLACE(EW.DZEW006,' ','')) COLUMN_NAMES,                                                                                                      ",
                 "       UPPER(TRIM(EW.DZEW007))           R_TABLE_NAME,                                                                                                      ",                                                                                            
                 "       UPPER(TRIM(EW.DZEW008))           R_COLUMN_NAMES                                                                                                     ",                                                                              
                 "  FROM DZEW_T EW                                                                                                                                            ",
                 " WHERE EW.DZEW001 = '",ms_table_name.toLowerCase(),"'                                                                                                       ",
                 "   AND EW.DZEW002 = '",ms_revision,"'                                                                                                                       ",
                 "   AND EW.DZEW003 = 'Constraint'                                                                                                                            ",
                 "   AND EW.DZEW010 = '",ms_alm_version,"'                                                                                                                    ",
                 "MINUS                                                                                                                                                       ",
                 "SELECT ACS.TABLE_NAME,                                                                                                                                      ",
                 "       ACS.CONSTRAINT_NAME,                                                                                                                                 ",
                 "       DECODE(ACS.CONSTRAINT_TYPE, 'R', 'R', ACS.CONSTRAINT_TYPE) CONSTRAINT_TYPE,                                                                          ",
                 "       ACC.COLUMN_NAMES,                                                                                                                                    ",
                 "       ACCR.TABLE_NAME R_TABLE_NAME,                                                                                                                        ",
                 "       ACCR.COLUMN_NAMES                                                                                                                                    ",
                 "  FROM DBA_CONSTRAINTS ACS                                                                                                                                  ",
                 "  LEFT OUTER JOIN (                                                                                                                                         ",
                 "                   SELECT ACCS.CONSTRAINT_NAME,ACCS.OWNER,ACCS.TABLE_NAME,LISTAGG(ACCS.COLUMN_NAME, ',') WITHIN GROUP(ORDER BY ACCS.POSITION) COLUMN_NAMES  ",
                 "                     FROM DBA_CONS_COLUMNS ACCS                                                                                                             ",
                 "                    GROUP BY ACCS.CONSTRAINT_NAME,ACCS.OWNER,ACCS.TABLE_NAME                                                                                ",
                 "                  ) ACC ON ACC.CONSTRAINT_NAME = ACS.CONSTRAINT_NAME                                                                                        ",
                 "                       AND ACC.TABLE_NAME      = ACS.TABLE_NAME                                                                                             ",                                                                             
                 "                       AND ACC.OWNER           = ACS.OWNER                                                                                                  ",
                 "  LEFT OUTER JOIN (                                                                                                                                         ",
                 "                   SELECT ACCS.CONSTRAINT_NAME,ACCS.OWNER,ACCS.TABLE_NAME,LISTAGG(ACCS.COLUMN_NAME, ',') WITHIN GROUP(ORDER BY ACCS.POSITION) COLUMN_NAMES  ",
                 "                     FROM DBA_CONS_COLUMNS ACCS                                                                                                             ",
                 "                    GROUP BY ACCS.CONSTRAINT_NAME,ACCS.OWNER,ACCS.TABLE_NAME                                                                                ",
                 "                  ) ACCR ON ACCR.CONSTRAINT_NAME = ACS.R_CONSTRAINT_NAME                                                                                    ",
                 "                        AND ACCR.OWNER           = ACS.R_OWNER                                                                                              ",
                 " WHERE 1 = 1                                                                                                                                                ",
                 "   AND ACS.TABLE_NAME = '",ls_table_name.toUpperCase(),"'                                                                                                   ",
                 "   AND ACS.OWNER      = '",ls_owner.toUpperCase(),"'                                                                                                        "
    &endif                 
    &endif                 
  ELSE 
    IF ms_view_type = cs_view_type_rt THEN
      &ifdef is_msv
      LET ls_sql = "select ew.dzew001                 table_name,         ",
                   "       ew.dzew004                 constraint_name,    ",
                   "       ew.dzew005                 constraint_type,    ",
                   "       replace(ew.dzew006,' ','') column_names,       ",
                   "       trim(ew.dzew007)           r_table_name,       ",
                   "       trim(ew.dzew008)           r_column_names      ",
                   "  from dzew_t ew                                      ",    
                   " where ew.dzew001 = '",ms_table_name.toLowerCase(),"' ",    
                   "   and ew.dzew002 = '",ms_revision,"'                 ",    
                   "   and ew.dzew003 = 'Constraint'                      ",    
                   "   and ew.dzew010 = '",ms_alm_version,"'              ",    
                   "except                                                ",
                   "select ed.dzed001       table_name,                   ",
                   "       ed.dzed002       constraint_name,              ",
                   "       ed.dzed003       constraint_type,              ",
                   "       ed.dzed004       column_names,                 ",
                   "       trim(ed.dzed005) r_table_name,                 ",
                   "       trim(ed.dzed006) r_column_names                ",
                   "  from dzed_t ed                                      ",    
                   " where ed.dzed001 = '",ls_table_name.toLowerCase(),"' ",    
                   "   and ed.dzed003 <> 'F'                              "     
      &else
      &ifdef is_pgs
      LET ls_sql = "select ew.dzew001                 table_name,         ",
                   "       ew.dzew004                 constraint_name,    ",
                   "       ew.dzew005                 constraint_type,    ",
                   "       replace(ew.dzew006,' ','') column_names,       ",
                   "       trim(ew.dzew007)           r_table_name,       ",
                   "       trim(ew.dzew008)           r_column_names      ",
                   "  from dzew_t ew                                      ",    
                   " where ew.dzew001 = '",ms_table_name.toLowerCase(),"' ",    
                   "   and ew.dzew002 = '",ms_revision,"'                 ",    
                   "   and ew.dzew003 = 'Constraint'                      ",    
                   "   and ew.dzew010 = '",ms_alm_version,"'              ",    
                   "except                                                ",
                   "select ed.dzed001       table_name,                   ",
                   "       ed.dzed002       constraint_name,              ",
                   "       ed.dzed003       constraint_type,              ",
                   "       ed.dzed004       column_names,                 ",
                   "       trim(ed.dzed005) r_table_name,                 ",
                   "       trim(ed.dzed006) r_column_names                ",
                   "  from dzed_t ed                                      ",    
                   " where ed.dzed001 = '",ls_table_name.toLowerCase(),"' ",    
                   "   and ed.dzed003 <> 'F'                              "     
      &else
      LET ls_sql = "SELECT UPPER(EW.DZEW001)                 TABLE_NAME,       ",
                   "       UPPER(EW.DZEW004)                 CONSTRAINT_NAME,  ",
                   "       UPPER(EW.DZEW005)                 CONSTRAINT_TYPE,  ",
                   "       UPPER(REPLACE(EW.DZEW006,' ','')) COLUMN_NAMES,     ",
                   "       UPPER(TRIM(EW.DZEW007))           R_TABLE_NAME,     ",                                                                                            
                   "       UPPER(TRIM(EW.DZEW008))           R_COLUMN_NAMES    ",                                                                              
                   "  FROM DZEW_T EW                                           ",
                   " WHERE EW.DZEW001 = '",ms_table_name.toLowerCase(),"'      ",
                   "   AND EW.DZEW002 = '",ms_revision,"'                      ",
                   "   AND EW.DZEW003 = 'Constraint'                           ",
                   "   AND EW.DZEW010 = '",ms_alm_version,"'                   ",
                   "MINUS                                                      ",
                   "SELECT UPPER(ED.DZED001)       TABLE_NAME,                 ",
                   "       UPPER(ED.DZED002)       CONSTRAINT_NAME,            ",
                   "       UPPER(ED.DZED003)       CONSTRAINT_TYPE,            ",
                   "       UPPER(ED.DZED004)       COLUMN_NAMES,               ",
                   "       UPPER(TRIM(ED.DZED005)) R_TABLE_NAME,               ",
                   "       UPPER(TRIM(ED.DZED006)) R_COLUMN_NAMES              ",
                   "  FROM DZED_T ED                                           ",
                   " WHERE ED.DZED001 = '",ls_table_name.toLowerCase(),"'      ",
                   "   AND ED.DZED003 <> 'F'                                   "
      &endif                 
      &endif                 
    END IF 
  END IF  
                
  PREPARE lpre_get_key_forward_records FROM ls_sql
  DECLARE lcur_get_key_forward_records CURSOR FOR lpre_get_key_forward_records

  OPEN lcur_get_key_forward_records
  FOREACH lcur_get_key_forward_records INTO lo_key_forward_records[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_key_forward_records
  
  FREE lpre_get_key_forward_records
  FREE lcur_get_key_forward_records  
  -------------------------------------------------------------------------------------------------
  
  LET lo_return.* = lo_key_forward_records.*
  
  RETURN lo_return
  
END FUNCTION 

FUNCTION sadzi140_rdif_get_key_backward_records(p_owner,p_table_name)
DEFINE
  p_owner      STRING,
  p_table_name STRING
DEFINE
  lo_key_backward_records DYNAMIC ARRAY OF T_KEY_DATA
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_key_name STRING,
  ls_sql         STRING,
  ls_exec_sql    STRING,
  li_rec_cnt     INTEGER,
  ls_alter_table VARCHAR(30)
DEFINE
  lo_return DYNAMIC ARRAY OF T_KEY_DATA  

  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name
  
  LET li_rec_cnt = 1

  ------------------------------------------------------------------------------------------------
  -- key backward
  ------------------------------------------------------------------------------------------------
  IF ms_view_type = cs_view_type_db THEN
    &ifdef is_msv
    LET ls_sql = "select acs.table_name,                                                        ",                                                                             
                 "       acs.constraint_name,                                                   ",                                                                             
                 "       iif(acs.constraint_type='R','R',acs.constraint_type) constraint_type,  ",                                                                       
                 "       acc.column_names,                                                      ",                                                                             
                 "       NULL  r_table_name,                                                    ",                                                                   
                 "       NULL  r_column_name                                                    ",                                                                              
                 "  from dba_constraints acs                                                    ",                                                                             
                 "  left outer join (                                                           ",                                                                             
                 "                   select accs.constraint_name,accs.owner,accs.table_name,    ",
                 "                          string_agg(accs.column_name, ',')                   ",
                 "                            within group(order by accs.position) column_names ",
                 "                     from dba_cons_columns accs                               ",                                                                             
                 "                    group by accs.constraint_name,accs.owner,accs.table_name  ",                                                                             
                 "                  ) acc on acc.constraint_name = acs.constraint_name          ",                                                                             
                 "                       and acc.table_name      = acs.table_name               ",                                                                             
                 "                       and acc.owner           = acs.owner                    ",                                                                             
                 " where 1 = 1                                                                  ",                                                                             
                 "   and acs.table_name = '",ls_table_name.toLowerCase(),"'                     ",                                                                             
                 "   and acs.owner      = '",ls_owner.toLowerCase(),"'                          ",                                                                             
                 "except                                                                        ",                                                                              
                 "select ew.dzew001                 table_name,                                 ",                                                                      
                 "       ew.dzew004                 constraint_name,                            ",                                                                      
                 "       ew.dzew005                 constraint_type,                            ",                                                                      
                 "       replace(ew.dzew006,' ','') column_names,                               ",                                                                      
                 "       trim(ew.dzew007)           r_table_name,                               ",                                                                      
                 "       trim(ew.dzew008)           r_column_names                              ",                                                                      
                 "  from dzew_t ew                                                              ",                                                                             
                 " where ew.dzew001 = '",ms_table_name.toLowerCase(),"'                         ",                                                                             
                 "   and ew.dzew002 = '",ms_revision,"'                                         ",                                                                             
                 "   and ew.dzew003 = 'Constraint'                                              ",                                                                             
                 "   and ew.dzew010 = '",ms_alm_version,"'                                      "                                                                             
    &else
    &ifdef is_pgs
    LET ls_sql = "select acs.table_name,                                                       ", 
                 "       acs.constraint_name,                                                  ", 
                 "       case                                                                  ",
                 "         when acs.constraint_type='R'                                        ",
                 "       then 'R'                                                              ",
                 "         else acs.constraint_type                                            ",
                 "       end constraint_type,                                                  ",
                 "       acc.column_names,                                                     ", 
                 "       NULL  r_table_name,                                                   ", 
                 "       NULL  r_column_name                                                   ", 
                 "  from dba_constraints acs                                                   ", 
                 "  left outer join (                                                          ", 
                 "                   select accs.constraint_name,accs.owner,accs.table_name,   ", 
                 "                          string_agg(accs.column_name, ','                   ",
                 "                            order by accs.position) column_names             ",
                 "                     from dba_cons_columns accs                              ", 
                 "                    group by accs.constraint_name,accs.owner,accs.table_name ", 
                 "                  ) acc on acc.constraint_name = acs.constraint_name         ", 
                 "                       and acc.table_name      = acs.table_name              ",                                                                             
                 "                       and acc.owner           = acs.owner                   ", 
                 " where 1 = 1                                                                 ", 
                 "   and acs.table_name = '",ls_table_name.toLowerCase(),"'                    ", 
                 "   and acs.owner      = '",ls_owner.toLowerCase(),"'                         ", 
                 "except                                                                       ", 
                 "select ew.dzew001                 table_name,                                ", 
                 "       ew.dzew004                 constraint_name,                           ", 
                 "       ew.dzew005                 constraint_type,                           ", 
                 "       replace(ew.dzew006,' ','') column_names,                              ", 
                 "       trim(ew.dzew007)           r_table_name,                              ", 
                 "       trim(ew.dzew008)           r_column_names                             ", 
                 "  from dzew_t ew                                                             ", 
                 " where ew.dzew001 = '",ms_table_name.toLowerCase(),"'                        ", 
                 "   and ew.dzew002 = '",ms_revision,"'                                        ", 
                 "   and ew.dzew003 = 'Constraint'                                             ", 
                 "   and ew.dzew010 = '",ms_alm_version,"'                                     " 
    &else
    LET ls_sql = "SELECT ACS.TABLE_NAME,                                                                                                                                      ",
                 "       ACS.CONSTRAINT_NAME,                                                                                                                                 ",
                 "       DECODE(ACS.CONSTRAINT_TYPE, 'R', 'R', ACS.CONSTRAINT_TYPE) CONSTRAINT_TYPE,                                                                          ",
                 "       ACC.COLUMN_NAMES,                                                                                                                                    ",
                 "       ACCR.TABLE_NAME R_TABLE_NAME,                                                                                                                        ",
                 "       ACCR.COLUMN_NAMES                                                                                                                                    ",
                 "  FROM DBA_CONSTRAINTS ACS                                                                                                                                  ",
                 "  LEFT OUTER JOIN (                                                                                                                                         ",
                 "                   SELECT ACCS.CONSTRAINT_NAME,ACCS.OWNER,ACCS.TABLE_NAME,LISTAGG(ACCS.COLUMN_NAME, ',') WITHIN GROUP(ORDER BY ACCS.POSITION) COLUMN_NAMES  ",
                 "                     FROM DBA_CONS_COLUMNS ACCS                                                                                                             ",
                 "                    GROUP BY ACCS.CONSTRAINT_NAME,ACCS.OWNER,ACCS.TABLE_NAME                                                                                ",
                 "                  ) ACC ON ACC.CONSTRAINT_NAME = ACS.CONSTRAINT_NAME                                                                                        ",
                 "                       AND ACC.TABLE_NAME      = ACS.TABLE_NAME                                                                                             ",                                                                             
                 "                       AND ACC.OWNER           = ACS.OWNER                                                                                                  ",
                 "  LEFT OUTER JOIN (                                                                                                                                         ",
                 "                   SELECT ACCS.CONSTRAINT_NAME,ACCS.OWNER,ACCS.TABLE_NAME,LISTAGG(ACCS.COLUMN_NAME, ',') WITHIN GROUP(ORDER BY ACCS.POSITION) COLUMN_NAMES  ",
                 "                     FROM DBA_CONS_COLUMNS ACCS                                                                                                             ",
                 "                    GROUP BY ACCS.CONSTRAINT_NAME,ACCS.OWNER,ACCS.TABLE_NAME                                                                                ",
                 "                  ) ACCR ON ACCR.CONSTRAINT_NAME = ACS.R_CONSTRAINT_NAME                                                                                    ",
                 "                        AND ACCR.OWNER           = ACS.R_OWNER                                                                                              ",
                 " WHERE 1 = 1                                                                                                                                                ",
                 "   AND ACS.TABLE_NAME = '",ls_table_name.toUpperCase(),"'                                                                                                   ",
                 "   AND ACS.OWNER      = '",ls_owner.toUpperCase(),"'                                                                                                        ",
                 "MINUS                                                                                                                                                       ",
                 "SELECT UPPER(EW.DZEW001)                 TABLE_NAME,                                                                                                        ",
                 "       UPPER(EW.DZEW004)                 CONSTRAINT_NAME,                                                                                                   ",
                 "       UPPER(EW.DZEW005)                 CONSTRAINT_TYPE,                                                                                                   ",
                 "       UPPER(REPLACE(EW.DZEW006,' ','')) COLUMN_NAMES,                                                                                                      ",
                 "       UPPER(TRIM(EW.DZEW007))           R_TABLE_NAME,                                                                                                      ",                                                                                            
                 "       UPPER(TRIM(EW.DZEW008))           R_COLUMN_NAMES                                                                                                     ",                                                                              
                 "  FROM DZEW_T EW                                                                                                                                            ",
                 " WHERE EW.DZEW001 = '",ms_table_name.toLowerCase(),"'                                                                                                       ",
                 "   AND EW.DZEW002 = '",ms_revision,"'                                                                                                                       ",
                 "   AND EW.DZEW003 = 'Constraint'                                                                                                                            ",
                 "   AND EW.DZEW010 = '",ms_alm_version,"'                                                                                                                    "
    &endif                  
    &endif                  
  ELSE 
    IF ms_view_type = cs_view_type_rt THEN
      &ifdef is_msv
      LET ls_sql = "select ed.dzed001       table_name,                   ",
                   "       ed.dzed002       constraint_name,              ",
                   "       ed.dzed003       constraint_type,              ",
                   "       ed.dzed004       column_names,                 ",
                   "       trim(ed.dzed005) r_table_name,                 ",
                   "       trim(ed.dzed006) r_column_names                ",
                   "  from dzed_t ed                                      ",    
                   " where ed.dzed001 = '",ls_table_name.toLowerCase(),"' ",    
                   "   and ed.dzed003 <> 'F'                              ",     
                   "except                                                ",
                   "select ew.dzew001                 table_name,         ",
                   "       ew.dzew004                 constraint_name,    ",
                   "       ew.dzew005                 constraint_type,    ",
                   "       replace(ew.dzew006,' ','') column_names,       ",
                   "       trim(ew.dzew007)           r_table_name,       ",
                   "       trim(ew.dzew008)           r_column_names      ",
                   "  from dzew_t ew                                      ",    
                   " where ew.dzew001 = '",ms_table_name.toLowerCase(),"' ",    
                   "   and ew.dzew002 = '",ms_revision,"'                 ",    
                   "   and ew.dzew003 = 'Constraint'                      ",    
                   "   and ew.dzew010 = '",ms_alm_version,"'              "    
      &else
      &ifdef is_pgs
      LET ls_sql = "select ed.dzed001       table_name,                   ",
                   "       ed.dzed002       constraint_name,              ",
                   "       ed.dzed003       constraint_type,              ",
                   "       ed.dzed004       column_names,                 ",
                   "       trim(ed.dzed005) r_table_name,                 ",
                   "       trim(ed.dzed006) r_column_names                ",
                   "  from dzed_t ed                                      ",    
                   " where ed.dzed001 = '",ls_table_name.toLowerCase(),"' ",    
                   "   and ed.dzed003 <> 'F'                              ",     
                   "except                                                ",
                   "select ew.dzew001                 table_name,         ",
                   "       ew.dzew004                 constraint_name,    ",
                   "       ew.dzew005                 constraint_type,    ",
                   "       replace(ew.dzew006,' ','') column_names,       ",
                   "       trim(ew.dzew007)           r_table_name,       ",
                   "       trim(ew.dzew008)           r_column_names      ",
                   "  from dzew_t ew                                      ",    
                   " where ew.dzew001 = '",ms_table_name.toLowerCase(),"' ",    
                   "   and ew.dzew002 = '",ms_revision,"'                 ",    
                   "   and ew.dzew003 = 'Constraint'                      ",    
                   "   and ew.dzew010 = '",ms_alm_version,"'              "    
      &else
      LET ls_sql = "SELECT UPPER(ED.DZED001)       TABLE_NAME,                ",
                   "       UPPER(ED.DZED002)       CONSTRAINT_NAME,           ",
                   "       UPPER(ED.DZED003)       CONSTRAINT_TYPE,           ",
                   "       UPPER(ED.DZED004)       COLUMN_NAMES,              ",
                   "       UPPER(TRIM(ED.DZED005)) R_TABLE_NAME,              ",
                   "       UPPER(TRIM(ED.DZED006)) R_COLUMN_NAMES             ",
                   "  FROM DZED_T ED                                          ",
                   " WHERE ED.DZED001 = '",ls_table_name.toLowerCase(),"'     ",
                   "   AND ED.DZED003 <> 'F'                                  ",               
                   "MINUS                                                     ",
                   "SELECT UPPER(EW.DZEW001)                 TABLE_NAME,      ",
                   "       UPPER(EW.DZEW004)                 CONSTRAINT_NAME, ",
                   "       UPPER(EW.DZEW005)                 CONSTRAINT_TYPE, ",
                   "       UPPER(REPLACE(EW.DZEW006,' ','')) COLUMN_NAMES,    ",
                   "       UPPER(TRIM(EW.DZEW007))           R_TABLE_NAME,    ",                                                                                            
                   "       UPPER(TRIM(EW.DZEW008))           R_COLUMN_NAMES   ",                                                                              
                   "  FROM DZEW_T EW                                          ",
                   " WHERE EW.DZEW001 = '",ms_table_name.toLowerCase(),"'     ",
                   "   AND EW.DZEW002 = '",ms_revision,"'                     ",
                   "   AND EW.DZEW003 = 'Constraint'                          ",
                   "   AND EW.DZEW010 = '",ms_alm_version,"'                  "
      &endif                   
      &endif                   
    END IF 
  END IF  
                
  PREPARE lpre_get_key_backward_records FROM ls_sql
  DECLARE lcur_get_key_backward_records CURSOR FOR lpre_get_key_backward_records

  OPEN lcur_get_key_backward_records
  FOREACH lcur_get_key_backward_records INTO lo_key_backward_records[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_key_backward_records
  
  FREE lpre_get_key_backward_records
  FREE lcur_get_key_backward_records  
  -------------------------------------------------------------------------------------------------
  
  LET lo_return.* = lo_key_backward_records.*
  
  RETURN lo_return
  
END FUNCTION 

FUNCTION sadzi140_rdif_fill_index_diff_list()
DEFINE
  lo_index_forward_records  DYNAMIC ARRAY OF T_INDEX_DATA,
  lo_index_backward_records DYNAMIC ARRAY OF T_INDEX_DATA 
DEFINE 
  ls_sql   STRING,
  li_index INTEGER,
  li_count INTEGER

  IF ms_view_type = cs_view_type_db THEN
    &ifdef is_msv
    LET ls_sql = "select ''                       record_type,                                            ",  
                 "       orig.index_name          index_name,                                             ",  
                 "       src_idx_list.index_type  index_type_src,                                         ",  
                 "       src_idx_list.column_name column_name_src,                                        ",  
                 "       dst_idx_list.index_type  index_type_dst,                                         ",  
                 "       dst_idx_list.column_name column_name_dst                                         ",  
                 "  from (                                                                                ",  
                 "         select ais.index_name index_name                                               ",  
                 "           from dba_indexes ais                                                         ",  
                 "          where ais.owner        = '",ms_owner_name.toLowerCase(),"'                    ",  
                 "            and ais.table_name   = '",ms_table_name.toLowerCase(),"'                    ",  
                 "         union                                                                          ",  
                 "         select ew.dzew004 index_name                                                   ",
                 "           from dzew_t ew                                                               ",  
                 "          where ew.dzew001 = '",ms_table_name.toLowerCase(),"'                          ",  
                 "            and ew.dzew002 = '",ms_revision,"'                                          ",  
                 "            and ew.dzew003 = 'Index'                                                    ",  
                 "            and ew.dzew010 = '",ms_alm_version,"'                                       ",  
                 "            and coalesce(ew.dzewstus,'Y') = 'Y'                                         ",       
                 "       ) orig                                                                           ",
                 "       left outer join                                                                  ",
                 "       (                                                                                ",  
                 "         select ais.index_name index_name,                                              ",
                 "                case upper(substring(ais.uniqueness,1,1))                               ",
                 "                  when 'N' then 'NORMAL'                                                ",                             
                 "                  when 'U' then 'UNIQUE'                                                ",                             
                 "                else                                                                    ", 
                 "                  ais.uniqueness                                                        ",
                 "                end index_type,                                                         ",
                 "                ais.column_names  column_name                                           ",
                 "           from dba_indexes ais                                                         ",
                 "          where ais.owner        = '",ms_owner_name.toLowerCase(),"'                    ",  
                 "            and ais.table_name   = '",ms_table_name.toLowerCase(),"'                    ",  
                 "       ) src_idx_list on src_idx_list.index_name = orig.index_name                      ",
                 "       left outer join                                                                  ",
                 "       (                                                                                ",  
                 "         select ew.dzew004 index_name,                                                  ",
                 "                case upper(ew.dzew005)                                                  ",
                 "                  when 'N' then 'NORMAL'                                                ",
                 "                  when 'U' then 'UNIQUE'                                                ",
                 "                else                                                                    ",
                 "                  upper(ew.dzew005)                                                     ",
                 "                end  index_type,                                                        ",
                 "                replace(ew.dzew006,' ','') column_name                                  ",
                 "           from dzew_t ew                                                               ",  
                 "          where ew.dzew001 = '",ms_table_name.toLowerCase(),"'                          ",  
                 "            and ew.dzew002 = '",ms_revision,"'                                          ",  
                 "            and ew.dzew003 = 'Index'                                                    ",  
                 "            and ew.dzew010 = '",ms_alm_version,"'                                       ",  
                 "            and coalesce(ew.dzewstus,'Y') = 'Y'                                         ",       
                 "       ) dst_idx_list on dst_idx_list.index_name = orig.index_name                      ",                                               
                 " where 1=1                                                                              "
    &else
    &ifdef is_pgs
    LET ls_sql = "select ''                       record_type,                                            ",  
                 "       orig.index_name          index_name,                                             ",  
                 "       src_idx_list.index_type  index_type_src,                                         ",  
                 "       src_idx_list.column_name column_name_src,                                        ",  
                 "       dst_idx_list.index_type  index_type_dst,                                         ",  
                 "       dst_idx_list.column_name column_name_dst                                         ",  
                 "  from (                                                                                ",  
                 "         select ais.index_name index_name                                               ",  
                 "           from dba_indexes ais                                                         ",  
                 "          where ais.owner        = '",ms_owner_name.toLowerCase(),"'                    ",  
                 "            and ais.table_name   = '",ms_table_name.toLowerCase(),"'                    ",  
                 "         union                                                                          ",  
                 "         select ew.dzew004 index_name                                                   ",
                 "           from dzew_t ew                                                               ",  
                 "          where ew.dzew001 = '",ms_table_name.toLowerCase(),"'                          ",  
                 "            and ew.dzew002 = '",ms_revision,"'                                          ",  
                 "            and ew.dzew003 = 'Index'                                                    ",  
                 "            and ew.dzew010 = '",ms_alm_version,"'                                       ",  
                 "            and coalesce(ew.dzewstus,'Y') = 'Y'                                         ",       
                 "       ) orig                                                                           ",
                 "       left outer join                                                                  ",
                 "       (                                                                                ",  
                 "         select ais.index_name index_name,                                              ",
                 "                case upper(substring(ais.uniqueness,1,1))                               ",
                 "                  when 'N' then 'NORMAL'                                                ",                             
                 "                  when 'U' then 'UNIQUE'                                                ",                             
                 "                else                                                                    ", 
                 "                  ais.uniqueness                                                        ",
                 "                end index_type,                                                         ",
                 "                string_agg(aic.column_name,','                                          ",
                 "                           order by aic.column_position)  column_name                   ",
                 "           from dba_indexes ais                                                         ",
                 "          inner join dba_ind_columns aic                                                ",
                 "                  on aic.table_owner = ais.table_owner                                  ",
                 "                 and aic.table_name  = ais.table_name                                   ",
                 "                 and aic.index_name  = ais.index_name                                   ",
                 "          where ais.owner        = '",ms_owner_name.toLowerCase(),"'                    ",  
                 "            and ais.table_name   = '",ms_table_name.toLowerCase(),"'                    ",  
                 "          group by ais.index_name,                                                      ",
                 "                   case upper(substring(ais.uniqueness,1,1))                            ",
                 "                     when 'N' then 'NORMAL'                                             ",                             
                 "                     when 'U' then 'UNIQUE'                                             ",                             
                 "                   else                                                                 ", 
                 "                     ais.uniqueness                                                     ",
                 "                   end                                                                  ",
                 "       ) src_idx_list on src_idx_list.index_name = orig.index_name                      ",
                 "       left outer join                                                                  ",
                 "       (                                                                                ",  
                 "         select ew.dzew004 index_name,                                                  ",
                 "                case upper(ew.dzew005)                                                  ",
                 "                  when 'N' then 'NORMAL'                                                ",
                 "                  when 'U' then 'UNIQUE'                                                ",
                 "                else                                                                    ",
                 "                  upper(ew.dzew005)                                                     ",
                 "                end  index_type,                                                        ",
                 "                replace(ew.dzew006,' ','') column_name                                  ",
                 "           from dzew_t ew                                                               ",  
                 "          where ew.dzew001 = '",ms_table_name.toLowerCase(),"'                          ",  
                 "            and ew.dzew002 = '",ms_revision,"'                                          ",  
                 "            and ew.dzew003 = 'Index'                                                    ",  
                 "            and ew.dzew010 = '",ms_alm_version,"'                                       ",  
                 "            and coalesce(ew.dzewstus,'Y') = 'Y'                                         ",       
                 "       ) dst_idx_list on dst_idx_list.index_name = orig.index_name                      ",                                               
                 " where 1=1                                                                              "
    &else
    LET ls_sql = "SELECT ''                       RECORD_TYPE,                                              ",    
                 "       ORIG.INDEX_NAME          INDEX_NAME,                                               ",
                 "       SRC_IDX_LIST.INDEX_TYPE  INDEX_TYPE_SRC,                                           ",
                 "       SRC_IDX_LIST.COLUMN_NAME COLUMN_NAME_SRC,                                          ",
                 "       DST_IDX_LIST.INDEX_TYPE  INDEX_TYPE_DST,                                           ",
                 "       DST_IDX_LIST.COLUMN_NAME COLUMN_NAME_DST                                           ",
                 "  FROM (                                                                                  ",
                 "         SELECT NVL(DC.CONSTRAINT_NAME,AIS.INDEX_NAME) INDEX_NAME                         ", --20230807
                 "           FROM DBA_INDEXES AIS,DBA_CONSTRAINTS DC                                        ", --20230807                 
                 "          WHERE AIS.OWNER        = '",ms_owner_name.toUpperCase(),"'                      ",              
                 "            AND AIS.TABLE_NAME   = '",ms_table_name.toUpperCase(),"'                      ",
                 "            AND AIS.INDEX_TYPE <> 'CLUSTER'                                               ", --20230807    
                 "            AND DC.OWNER      (+)= AIS.OWNER                                              ", --20230807
                 "            AND DC.INDEX_NAME (+)= AIS.INDEX_NAME                                         ", --20230807
                 "            AND DC.TABLE_NAME (+)= AIS.TABLE_NAME                                         ", --20230807
                 "         UNION                                                                            ",                                      
                 "         SELECT UPPER(EW.DZEW004) INDEX_NAME                                              ",
                 "           FROM DZEW_T EW                                                                 ",
                 "          WHERE EW.DZEW001 = '",ms_table_name.toLowerCase(),"'                            ",
                 "            AND EW.DZEW002 = '",ms_revision,"'                                            ",
                 "            AND EW.DZEW003 = 'Index'                                                      ",
                 "            AND EW.DZEW010 = '",ms_alm_version,"'                                         ",
                 "            AND NVL(EW.DZEWSTUS,'Y') = 'Y'                                                ", #20170912                      
                 "       ) ORIG,                                                                            ",
                 "       (                                                                                  ",
                 "         SELECT NVL(DC.CONSTRAINT_NAME,AIS.INDEX_NAME) INDEX_NAME,                        ", --20230807                                    
                 "                DECODE(AIS.UNIQUENESS,'UNIQUE','UNIQUE',AIS.INDEX_TYPE) INDEX_TYPE,       ",                                     
                 "                LISTAGG(AICS.COLUMN_NAME,',')                                             ",
                 "                 WITHIN GROUP                                                             ",
                 "                 (ORDER BY AIS.INDEX_NAME,AICS.COLUMN_POSITION) COLUMN_NAME               ",
                 "           FROM DBA_INDEXES AIS,DBA_IND_COLUMNS AICS,DBA_CONSTRAINTS DC                   ", --20230807                                     
                 "          WHERE AIS.OWNER        = '",ms_owner_name.toUpperCase(),"'                      ",             
                 "            AND AIS.TABLE_NAME   = '",ms_table_name.toUpperCase(),"'                      ",            
                 "            AND AIS.INDEX_TYPE <> 'CLUSTER'                                               ", --20230807    
                 "            AND AICS.TABLE_OWNER = AIS.TABLE_OWNER                                        ",                                     
                 "            AND AICS.TABLE_NAME  = AIS.TABLE_NAME                                         ",                                     
                 "            AND AICS.INDEX_NAME  = AIS.INDEX_NAME                                         ",
                 "            AND DC.OWNER      (+)= AIS.OWNER                                              ", --20230807
                 "            AND DC.INDEX_NAME (+)= AIS.INDEX_NAME                                         ", --20230807
                 "            AND DC.TABLE_NAME (+)= AIS.TABLE_NAME                                         ", --20230807
                 "          GROUP BY NVL(DC.CONSTRAINT_NAME,AIS.INDEX_NAME),                                ", --20230807
                 "                   DECODE(AIS.UNIQUENESS,'UNIQUE','UNIQUE',AIS.INDEX_TYPE)                ",
                 "       ) SRC_IDX_LIST,                                                                    ",
                 "       (                                                                                  ",
                 "         SELECT UPPER(EW.DZEW004) INDEX_NAME,                                             ",
                 "                DECODE(UPPER(EW.DZEW005),                                                 ",
                 "                       'N','NORMAL',                                                      ",
                 "                       'U','UNIQUE',                                                      ",
                 "                       UPPER(EW.DZEW005)) INDEX_TYPE,                                     ",
                 "                UPPER(REPLACE(EW.DZEW006,' ','')) COLUMN_NAME                             ",
                 "           FROM DZEW_T EW                                                                 ",
                 "          WHERE EW.DZEW001 = '",ms_table_name.toLowerCase(),"'                            ",
                 "            AND EW.DZEW002 = '",ms_revision,"'                                            ",
                 "            AND EW.DZEW003 = 'Index'                                                      ",
                 "            AND EW.DZEW010 = '",ms_alm_version,"'                                         ",
                 "            AND NVL(EW.DZEWSTUS,'Y') = 'Y'                                                ", #20170912  
                 "       ) DST_IDX_LIST                                                                     ",
                 " WHERE SRC_IDX_LIST.INDEX_NAME (+)= ORIG.INDEX_NAME                                       ",
                 "   AND DST_IDX_LIST.INDEX_NAME (+)= ORIG.INDEX_NAME                                       "
    &endif                 
    &endif                 
  ELSE 
    IF ms_view_type = cs_view_type_rt THEN
      &ifdef is_msv
      LET ls_sql = "select ''                       record_type,                       ",
                   "       orig.index_name          index_name,                        ",
                   "       src_idx_list.index_type  index_type_src,                    ",
                   "       src_idx_list.column_name column_name_src,                   ",
                   "       dst_idx_list.index_type  index_type_dst,                    ",
                   "       dst_idx_list.column_name column_name_dst                    ",
                   "  from (                                                           ",
                   "         select ec.dzec002 index_name                              ",
                   "           from dzec_t ec                                          ",
                   "          where ec.dzec001 = '",ms_table_name.toLowerCase(),"'     ",
                   "            and coalesce(ec.dzecstus,'Y') = 'Y'                    ", 
                   "         union                                                     ",
                   "         select ew.dzew004 index_name                              ",
                   "           from dzew_t ew                                          ",
                   "          where ew.dzew001 = '",ms_table_name.toLowerCase(),"'     ",
                   "            and ew.dzew002 = '",ms_revision,"'                     ",
                   "            and ew.dzew003 = 'Index'                               ",
                   "            and ew.dzew010 = '",ms_alm_version,"'                  ",
                   "            and coalesce(ew.dzewstus,'Y') = 'Y'                    ", 
                   "       ) orig                                                      ",
                   "       left outer join                                             ",
                   "       (                                                           ",
                   "         select ec.dzec002 index_name,                             ",
                   "                case upper(ec.dzec003)                             ",
                   "                  when 'N' then 'NORMAL'                           ",
                   "                  when 'U' then 'UNIQUE'                           ",
                   "                else                                               ",
                   "                  upper(ec.dzec003)                                ",
                   "                end index_type,                                    ",
                   "                replace(ec.dzec004,' ','') column_name             ",
                   "           from dzec_t ec                                          ",
                   "          where ec.dzec001 = '",ms_table_name.toLowerCase(),"'     ",
                   "            and coalesce(ec.dzecstus,'Y') = 'Y'                    ", 
                   "       ) src_idx_list on src_idx_list.index_name = orig.index_name ",
                   "       left outer join                                             ",
                   "       (                                                           ",
                   "         select ew.dzew004 index_name,                             ",
                   "                case upper(ew.dzew005)                             ",
                   "                  when 'N' then 'NORMAL'                           ",
                   "                  when 'U' then 'UNIQUE'                           ",
                   "                else                                               ",
                   "                  upper(ew.dzew005)                                ",
                   "                end index_type,                                    ",
                   "                replace(ew.dzew006,' ','') column_name             ",
                   "           from dzew_t ew                                          ",
                   "          where ew.dzew001 = '",ms_table_name.toLowerCase(),"'     ",
                   "            and ew.dzew002 = '",ms_revision,"'                     ", 
                   "            and ew.dzew003 = 'Index'                               ", 
                   "            and ew.dzew010 = '",ms_alm_version,"'                  ", 
                   "            and coalesce(ew.dzewstus,'Y') = 'Y'                    ",  
                   "       ) dst_idx_list on dst_idx_list.index_name = orig.index_name ",            
                   " where 1=1                                                         " 
      &else
      &ifdef is_pgs
      LET ls_sql = "select ''                       record_type,                       ",
                   "       orig.index_name          index_name,                        ",
                   "       src_idx_list.index_type  index_type_src,                    ",
                   "       src_idx_list.column_name column_name_src,                   ",
                   "       dst_idx_list.index_type  index_type_dst,                    ",
                   "       dst_idx_list.column_name column_name_dst                    ",
                   "  from (                                                           ",
                   "         select ec.dzec002 index_name                              ",
                   "           from dzec_t ec                                          ",
                   "          where ec.dzec001 = '",ms_table_name.toLowerCase(),"'     ",
                   "            and coalesce(ec.dzecstus,'Y') = 'Y'                    ", 
                   "         union                                                     ",
                   "         select ew.dzew004 index_name                              ",
                   "           from dzew_t ew                                          ",
                   "          where ew.dzew001 = '",ms_table_name.toLowerCase(),"'     ",
                   "            and ew.dzew002 = '",ms_revision,"'                     ",
                   "            and ew.dzew003 = 'Index'                               ",
                   "            and ew.dzew010 = '",ms_alm_version,"'                  ",
                   "            and coalesce(ew.dzewstus,'Y') = 'Y'                    ", 
                   "       ) orig                                                      ",
                   "       left outer join                                             ",
                   "       (                                                           ",
                   "         select ec.dzec002 index_name,                             ",
                   "                case upper(ec.dzec003)                             ",
                   "                  when 'N' then 'NORMAL'                           ",
                   "                  when 'U' then 'UNIQUE'                           ",
                   "                else                                               ",
                   "                  upper(ec.dzec003)                                ",
                   "                end index_type,                                    ",
                   "                replace(ec.dzec004,' ','') column_name             ",
                   "           from dzec_t ec                                          ",
                   "          where ec.dzec001 = '",ms_table_name.toLowerCase(),"'     ",
                   "            and coalesce(ec.dzecstus,'Y') = 'Y'                    ", 
                   "       ) src_idx_list on src_idx_list.index_name = orig.index_name ",
                   "       left outer join                                             ",
                   "       (                                                           ",
                   "         select ew.dzew004 index_name,                             ",
                   "                case upper(ew.dzew005)                             ",
                   "                  when 'N' then 'NORMAL'                           ",
                   "                  when 'U' then 'UNIQUE'                           ",
                   "                else                                               ",
                   "                  upper(ew.dzew005)                                ",
                   "                end index_type,                                    ",
                   "                replace(ew.dzew006,' ','') column_name             ",
                   "           from dzew_t ew                                          ",
                   "          where ew.dzew001 = '",ms_table_name.toLowerCase(),"'     ",
                   "            and ew.dzew002 = '",ms_revision,"'                     ", 
                   "            and ew.dzew003 = 'Index'                               ", 
                   "            and ew.dzew010 = '",ms_alm_version,"'                  ", 
                   "            and coalesce(ew.dzewstus,'Y') = 'Y'                    ",  
                   "       ) dst_idx_list on dst_idx_list.index_name = orig.index_name ",            
                   " where 1=1                                                         " 
      &else
      LET ls_sql = "SELECT ''                       RECORD_TYPE,                   ",    
                   "       ORIG.INDEX_NAME          INDEX_NAME,                    ",
                   "       SRC_IDX_LIST.INDEX_TYPE  INDEX_TYPE_SRC,                ",
                   "       SRC_IDX_LIST.COLUMN_NAME COLUMN_NAME_SRC,               ",
                   "       DST_IDX_LIST.INDEX_TYPE  INDEX_TYPE_DST,                ",
                   "       DST_IDX_LIST.COLUMN_NAME COLUMN_NAME_DST                ",
                   "  FROM (                                                       ",
                   "         SELECT UPPER(EC.DZEC002) INDEX_NAME                   ",                                     
                   "           FROM DZEC_T EC                                      ",                                      
                   "          WHERE EC.DZEC001 = '",ms_table_name.toLowerCase(),"' ",
                   "            AND NVL(EC.DZECSTUS,'Y') = 'Y'                     ", #20170912                      
                   "         UNION                                                 ",                                      
                   "         SELECT UPPER(EW.DZEW004) INDEX_NAME                   ",
                   "           FROM DZEW_T EW                                      ",
                   "          WHERE EW.DZEW001 = '",ms_table_name.toLowerCase(),"' ",
                   "            AND EW.DZEW002 = '",ms_revision,"'                 ",
                   "            AND EW.DZEW003 = 'Index'                           ",
                   "            AND EW.DZEW010 = '",ms_alm_version,"'              ",
                   "            AND NVL(EW.DZEWSTUS,'Y') = 'Y'                     ", #20170912                      
                   "       ) ORIG,                                                 ",
                   "       (                                                       ",
                   "         SELECT UPPER(EC.DZEC002) INDEX_NAME,                  ",                                     
                   "                DECODE(UPPER(EC.DZEC003),                      ",
                   "                       'N','NORMAL',                           ",
                   "                       'U','UNIQUE',                           ",
                   "                       UPPER(EC.DZEC003)) INDEX_TYPE,          ",                                             
                   "                UPPER(REPLACE(EC.DZEC004,' ','')) COLUMN_NAME  ",                                     
                   "           FROM DZEC_T EC                                      ",                                     
                   "          WHERE EC.DZEC001 = '",ms_table_name.toLowerCase(),"' ",            
                   "            AND NVL(EC.DZECSTUS,'Y') = 'Y'                     ", #20170912                      
                   "       ) SRC_IDX_LIST,                                         ",
                   "       (                                                       ",
                   "         SELECT UPPER(EW.DZEW004) INDEX_NAME,                  ",
                   "                DECODE(UPPER(EW.DZEW005),                      ",
                   "                       'N','NORMAL',                           ",
                   "                       'U','UNIQUE',                           ",
                   "                       UPPER(EW.DZEW005)) INDEX_TYPE,          ",
                   "                UPPER(REPLACE(EW.DZEW006,' ','')) COLUMN_NAME  ",
                   "           FROM DZEW_T EW                                      ",
                   "          WHERE EW.DZEW001 = '",ms_table_name.toLowerCase(),"' ",
                   "            AND EW.DZEW002 = '",ms_revision,"'                 ",
                   "            AND EW.DZEW003 = 'Index'                           ",
                   "            AND EW.DZEW010 = '",ms_alm_version,"'              ",
                   "            AND NVL(EW.DZEWSTUS,'Y') = 'Y'                     ", #20170912                      
                   "       ) DST_IDX_LIST                                          ",
                   " WHERE SRC_IDX_LIST.INDEX_NAME (+)= ORIG.INDEX_NAME            ",
                   "   AND DST_IDX_LIST.INDEX_NAME (+)= ORIG.INDEX_NAME            "
      &endif             
      &endif             
    END IF
  END IF   

  PREPARE lpre_fill_fill_index_diff_list FROM ls_sql
  DECLARE lcur_fill_fill_index_diff_list CURSOR FOR lpre_fill_fill_index_diff_list

  CALL m_diff_idx_list.clear()
  CALL m_diff_idx_list_color.clear()
  
  LET li_count = 1

  #取得索引差異資料
  CALL sadzi140_rdif_get_index_forward_records(ms_owner_name,ms_table_name) RETURNING lo_index_forward_records
  CALL sadzi140_rdif_get_index_backward_records(ms_owner_name,ms_table_name) RETURNING lo_index_backward_records
  
  #DISPLAY ls_TableName
  FOREACH lcur_fill_fill_index_diff_list INTO m_diff_idx_list[li_count].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET m_diff_idx_list[li_count].RECORD_TYPE = NULL
    
    LET m_diff_idx_list_color[li_count].RECORD_TYPE    = NULL
    LET m_diff_idx_list_color[li_count].INDEX_NAME     = NULL
    LET m_diff_idx_list_color[li_count].INDEX_TYPE_DB  = NULL
    LET m_diff_idx_list_color[li_count].COLUMN_NAME_DB = NULL
    LET m_diff_idx_list_color[li_count].INDEX_TYPE_RT  = NULL
    LET m_diff_idx_list_color[li_count].COLUMN_NAME_RT = NULL

    FOR li_index = 1 TO lo_index_forward_records.getLength()
      IF m_diff_idx_list[li_count].index_name = lo_index_forward_records[li_index].index_name THEN
        LET m_diff_idx_list[li_count].RECORD_TYPE = "!"
        LET m_diff_idx_list_color[li_count].RECORD_TYPE    = "red"
        LET m_diff_idx_list_color[li_count].INDEX_NAME     = "blue"
        LET m_diff_idx_list_color[li_count].INDEX_TYPE_RT  = "yellow reverse"
        LET m_diff_idx_list_color[li_count].COLUMN_NAME_RT = "yellow reverse"
        EXIT FOR;
      END IF    
    END FOR

    FOR li_index = 1 TO lo_index_backward_records.getLength()
      IF m_diff_idx_list[li_count].index_name = lo_index_backward_records[li_index].index_name THEN
        LET m_diff_idx_list[li_count].RECORD_TYPE = "!"
        LET m_diff_idx_list_color[li_count].RECORD_TYPE    = "red"
        LET m_diff_idx_list_color[li_count].INDEX_NAME     = "blue"
        LET m_diff_idx_list_color[li_count].INDEX_TYPE_DB  = "red reverse"
        LET m_diff_idx_list_color[li_count].COLUMN_NAME_DB = "red reverse"
        EXIT FOR
      END IF    
    END FOR
    
    LET li_count = li_count + 1
    
  END FOREACH
  #CALL m_diff_idx_list.deleteElement(li_count)

END FUNCTION

FUNCTION sadzi140_rdif_get_column_forward_records(p_owner,p_table_name)
DEFINE
  p_owner      STRING,
  p_table_name STRING
DEFINE
  lo_column_forward_records DYNAMIC ARRAY OF T_COLUMN_DATA
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_column_name STRING,
  ls_sql         STRING,
  ls_exec_sql    STRING,
  ls_ora_coldef  STRING, #20240529
  ls_db_type     STRING, #20240606
  li_rec_cnt     INTEGER,
  ls_alter_table VARCHAR(30)
DEFINE
  lo_return DYNAMIC ARRAY OF T_COLUMN_DATA

  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name
  
  LET li_rec_cnt = 1

  ------------------------------------------------------------------------------------------------
  -- Column forward
  ------------------------------------------------------------------------------------------------

  IF ms_view_type = cs_view_type_db THEN
    &ifdef is_msv
    LET ls_sql = "select ev.dzev005       column_name,                      ",                                
                 "       ev.dzev006       iskey,                            ",                                
                 "       ev.dzev007       nullable,                         ",                                
                 "       ev.dzev008       data_type,                        ",                                
                 "       ev.dzev009       data_length,                      ",                                
                 "       trim(ev.dzev021) data_default                      ",                                
                 "  from dzev_t ev                                          ",                                
                 " where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'     ",                                
                 "   and ev.dzev002 = '",ms_table_name.toLowerCase(),"'     ",                                
                 "   and ev.dzev003 = '",ms_revision,"'                     ",                                
                 "   and ev.dzev004 = 'TableDesigner'                       ",                                
                 "   and ev.dzev018 = '",ms_alm_version,"'                  ",                                
                 "except                                                    ",                                 
                 "select atc.column_name                 column_name,       ",                           
                 "       iif(atc.nullable='Y','N','Y')   iskey,             ",                           
                 "       atc.nullable                    nullable,          ",                           
                 "       atc.data_type                   data_type,         ",   
                 "       atc.data_length                 data_length,       ",                  
                 "       atc.data_default                data_default       ",      
                 "  from dba_tab_columns atc                                ",                                
                 "  left outer join dba_cons_columns acc                    ",   
                 "               on acc.owner       = atc.owner             ",        
                 "              and acc.table_name  = atc.table_name        ",        
                 "              and acc.column_name = atc.column_name       ",        
                 "              and acc.constraint_name like '%pk'          ",        
                 " where 1=1                                                ",                                
                 "   and atc.owner      = '",ls_owner.toLowerCase(),"'      ",                                
                 "   and atc.table_name = '",ls_table_name.toLowerCase(),"' ",                                
                 " order by 1                                               "                                
    &else 
    &ifdef is_pgs
    LET ls_sql = "select ev.dzev005           column_name,                                    ",                                
                 "       ev.dzev006           iskey,                                          ",                                
                 "       ev.dzev007           nullable,                                       ",                                
                 "       ev.dzev008           data_type,                                      ",                                
                 "       ev.dzev009::integer  data_length,                                    ",                                
                 "       trim(ev.dzev021)     data_default                                    ",                                
                 "  from dzev_t ev                                                            ",                                
                 " where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'                       ",                                
                 "   and ev.dzev002 = '",ms_table_name.toLowerCase(),"'                       ",                                
                 "   and ev.dzev003 = '",ms_revision,"'                                       ",                                
                 "   and ev.dzev004 = 'TableDesigner'                                         ",                                
                 "   and ev.dzev018 = '",ms_alm_version,"'                                    ",                                
                 "except                                                                      ",                                 
                 "select atc.column_name                 column_name,                         ",                           
                 "       case                                                                 ", 
                 "         when atc.nullable='Y' then 'N'                                     ",  
                 "       else 'Y' end                    iskey,                               ",                           
                 "       atc.nullable                    nullable,                            ",                           
                 "       atc.data_type                   data_type,                           ",   
                 "       case                                                                 ",
                 "         when atc.data_type = 'text' then null                              ",
                 "         when atc.data_type = 'bytea' then null                             ",
                 "         when atc.data_type = 'date' then null                              ",
                 "         when atc.data_type = 'int2' then null                              ",
                 "         when atc.data_type = 'int4' then null                              ",
                 "         when atc.data_type = 'int8' then null                              ",
                 "         when atc.data_type = 'decimal' then                                ",
                 "           atc.data_precision::integer                                      ",
                 "       else                                                                 ",
                 "         atc.data_length::integer                                           ",
                 "       end::integer                    data_length,                         ",
                 "       case                                                                 ",
                 "         when atc.data_default is null then null                            ",
                 "         when atc.data_default = 'NULL' then null                           ",
                 "         when atc.data_default = 'null' then null                           ",
                 "         when atc.data_default like 'CURRENT%' then lower(atc.data_default) ",
                 "         when atc.data_default = '\"sysdate\"()' then 'sysdate'             ",
                 "         when atc.data_default = 'SYSDATE' then 'sysdate'                   ",
                 "       else                                                                 ",
                 "         atc.data_default                                                   ",
                 "       end as data_default                                                  ",
                 "  from dba_tab_columns atc                                                  ",                                
                 "  left outer join dba_cons_columns acc                                      ",   
                 "               on acc.owner       = atc.owner                               ",        
                 "              and acc.table_name  = atc.table_name                          ",        
                 "              and acc.column_name = atc.column_name                         ",        
                 "              and acc.constraint_name like '%pk'                            ",        
                 " where 1=1                                                                  ",                                
                 "   and atc.owner      = '",ls_owner.toLowerCase(),"'                        ",                                
                 "   and atc.table_name = '",ls_table_name.toLowerCase(),"'                   ",                                
                 " order by 1                                                                 "                                
    &else 
    LET ls_db_type = sadzi140_util_get_db_type()
    CASE ls_db_type
      WHEN cs_db_driver_type_kdb
        LET ls_ora_coldef = " CASE                                                                 ",
                            "   WHEN ATC.DATA_DEFAULT IS NULL THEN NULL                            ",
                            "   WHEN ATC.DATA_DEFAULT = 'NULL' THEN NULL                           ",
                            "   WHEN ATC.DATA_DEFAULT = 'null' THEN NULL                           ",
                            "   WHEN ATC.DATA_DEFAULT LIKE 'CURRENT%' THEN LOWER(ATC.DATA_DEFAULT) ",
                            "   WHEN ATC.DATA_DEFAULT = '\"sysdate\"()' then 'sysdate'             ",
                            "   WHEN ATC.DATA_DEFAULT = 'SYSDATE' then 'sysdate'                   ",
                            " ELSE                                                                 ",
                            "   ATC.DATA_DEFAULT                                                   ",
                            " END AS DATA_DEFAULT                                                  "
      WHEN cs_db_driver_type_dmg 
        LET ls_ora_coldef = " CASE                                                                 ",
                            "   WHEN ATC.DATA_DEFAULT IS NULL THEN NULL                            ",
                            "   WHEN ATC.DATA_DEFAULT = 'NULL' THEN NULL                           ",
                            "   WHEN ATC.DATA_DEFAULT = 'null' THEN NULL                           ",
                            "   WHEN ATC.DATA_DEFAULT LIKE 'CURRENT%' THEN LOWER(ATC.DATA_DEFAULT) ",
                            "   WHEN ATC.DATA_DEFAULT = '\"sysdate\"()' then 'sysdate'             ",
                            "   WHEN ATC.DATA_DEFAULT = 'SYSDATE' then 'sysdate'                   ",
                            " ELSE                                                                 ",
                            "   ATC.DATA_DEFAULT                                                   ",
                            " END AS DATA_DEFAULT                                                  "
    OTHERWISE  
      LET ls_ora_coldef = " DECODE(TRIM(DS.GET_COL_DEFAULT(ATC.TABLE_NAME,ATC.COLUMN_NAME)),'',NULL,          ",
                          "        TRIM(DS.GET_COL_DEFAULT(ATC.TABLE_NAME,ATC.COLUMN_NAME))) DATA_DEFAULT     "
    END CASE
    
    LET ls_sql = "SELECT EV.DZEV005       COLUMN_NAME,                                                     ",
                 "       EV.DZEV006       ISKEY,                                                           ",
                 "       EV.DZEV007       NULLABLE,                                                        ",
                 "       EV.DZEV008       DATA_TYPE,                                                       ",
                 "       EV.DZEV009       DATA_LENGTH,                                                     ",
                 "       TRIM(DECODE(EV.DZEV021,'',NULL,EV.DZEV021)) DATA_DEFAULT                          ",
                 "  FROM DZEV_T EV                                                                         ",
                 " WHERE EV.DZEV001 = '",ms_owner_name.toUpperCase(),"'                                    ",
                 "   AND EV.DZEV002 = '",ms_table_name.toUpperCase(),"'                                    ",
                 "   AND EV.DZEV003 = '",ms_revision,"'                                                    ",
                 "   AND EV.DZEV004 = 'TableDesigner'                                                      ",
                 "   AND EV.DZEV018 = '",ms_alm_version,"'                                                 ",
                 "MINUS                                                                                    ",
                 "SELECT ATC.COLUMN_NAME                      COLUMN_NAME,                                 ",
                 "       DECODE(ACC.COLUMN_NAME,NULL,'N','Y') ISKEY,                                       ",
                 "       ATC.NULLABLE                         NULLABLE,                                    ",
                 "       DECODE(INSTRB(ATC.DATA_TYPE,'TIMESTAMP'),1,'TIMESTAMP',ATC.DATA_TYPE) DATA_TYPE,  ",
                 "       REPLACE(TO_CHAR(DECODE(                                                           ",
                 "               ATC.DATA_TYPE,                                                            ",
                 "               'BLOB',4000,                                                              ", #20240529
                 "               'CLOB',4000,                                                              ", #20240529
                 "               'INT2',NULL,                                                              ", #20240529
                 "               'INT4',NULL,                                                              ", #20240529
                 "               'INT8',NULL,                                                              ", #20240529
                 "               'NUMBER',DECODE(                                                          ",
                 "                                NVL(ATC.DATA_SCALE,'0'),                                 ",
                 "                                '0',ATC.DATA_PRECISION,                                  ",
                 "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE                  ",
                 "                              ),                                                         ",
                 "               'NUMERIC',DECODE(                                                         ", #20240529 begin
                 "                                NVL(ATC.DATA_SCALE,'0'),                                 ",
                 "                                '0',ATC.DATA_PRECISION,                                  ",
                 "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE                  ", 
                 "                              ),                                                         ", #20240529 end
                 "               'DATE',7,                                                                 ", #20240529
                 "               'TIMESTAMP(0)',7,                                                         ", #20240529
                 "               'TIMESTAMP(5)',11,                                                        ", #20240529
                 "               'TIMESTAMP(6)',11,                                                        ", #20240529
                 "               'TIMESTAMP',DECODE(                                                       ",
                 "                                   ATC.DATA_SCALE,0,7,8,7,11                             ", #20240529
                 "                                 ),                                                      ",                
                 "               'DATE',DECODE(                                                            ",
                 "                              ATC.DATA_LENGTH,3,7,ATC.DATA_LENGTH                        ",
                 "                            ),                                                           ",                
                 "               ATC.DATA_LENGTH                                                           ",
                 "             )),'.',',')                             DATA_LENGTH,                        ",
                 ls_ora_coldef,
                 "  FROM DBA_TAB_COLUMNS ATC                                                               ",
                 "  LEFT OUTER JOIN DBA_CONS_COLUMNS ACC ON ACC.OWNER       = ATC.OWNER                    ",
                 "                                      AND ACC.TABLE_NAME  = ATC.TABLE_NAME               ",
                 "                                      AND ACC.COLUMN_NAME = ATC.COLUMN_NAME              ",
                 "                                      AND ACC.CONSTRAINT_NAME LIKE '%PK'                 ",
                 " WHERE 1=1                                                                               ",
                 "   AND ATC.OWNER      = '",ls_owner.toUpperCase(),"'                                     ",
                 "   AND ATC.TABLE_NAME = '",ls_table_name.toUpperCase(),"'                                ",
                 " ORDER BY 1                                                                              "
    &endif             
    &endif             
  ELSE 
    IF ms_view_type = cs_view_type_rt THEN
      &ifdef is_msv
      LET ls_sql = "select ev.dzev005       column_name,                        ",    
                   "       ev.dzev006       iskey,                              ",    
                   "       ev.dzev007       nullable,                           ",    
                   "       ev.dzev008       data_type,                          ",    
                   "       ev.dzev009       data_length,                        ",    
                   "       trim(ev.dzev021) data_default                        ",    
                   "  from dzev_t ev                                            ",    
                   " where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'       ",    
                   "   and ev.dzev002 = '",ms_table_name.toLowerCase(),"'       ",    
                   "   and ev.dzev003 = '",ms_revision,"'                       ",    
                   "   and ev.dzev004 = 'TableDesigner'                         ",    
                   "   and ev.dzev018 = '",ms_alm_version,"'                    ",    
                   "except                                                      ",     
                   "select eb.dzeb002                      column_name,         ",  
                   "       eb.dzeb004                      iskey,               ",  
                   "       iif(eb.dzeb004='Y','N','Y')     nullable,            ",  
                   "       td.gztd003                      data_type,           ",  
                   "       case td.gztd003                                      ",  
                   "         when 'varbinary'                                   ",  
                   "           then 'max'                                       ",  
                   "         when 'nvarchar'                                    ",  
                   "           then iif(td.gztd009='TEXT','max',td.gztd008)     ",  
                   "       else                                                 ",  
                   "         td.gztd008                                         ",  
                   "       end                             data_length,         ",  
                   "       trim(eb.dzeb012)                data_default         ",  
                   "  from dzeb_t eb                                            ",    
                   "  left outer join gztd_t td                                 ",
                   "               on td.gztd001 = eb.dzeb006                   ",    
                   " where eb.dzeb001 = '",ms_table_name.toLowerCase(),"'       ",    
                   " order by 1                                                 "
      &else 
      &ifdef is_pgs
      LET ls_sql = "select ev.dzev005       column_name,                        ",    
                   "       ev.dzev006       iskey,                              ",    
                   "       ev.dzev007       nullable,                           ",    
                   "       ev.dzev008       data_type,                          ",    
                   "       ev.dzev009       data_length,                        ",    
                   "       trim(ev.dzev021) data_default                        ",    
                   "  from dzev_t ev                                            ",    
                   " where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'       ",    
                   "   and ev.dzev002 = '",ms_table_name.toLowerCase(),"'       ",    
                   "   and ev.dzev003 = '",ms_revision,"'                       ",    
                   "   and ev.dzev004 = 'TableDesigner'                         ",    
                   "   and ev.dzev018 = '",ms_alm_version,"'                    ",    
                   "except                                                      ",     
                   "select eb.dzeb002                      column_name,         ",  
                   "       eb.dzeb004                      iskey,               ",  
                   "       case                                                 ",
                   "         when eb.dzeb004='Y'                                ",
                   "           then 'N'                                         ",
                   "         else                                               ",
                   "           'Y'                                              ",
                   "         end                           nullable,            ",  
                   "       td.gztd003                      data_type,           ",  
                   "       td.gztd008                      data_length,         ",  
                   "       trim(eb.dzeb012)                data_default         ",  
                   "  from dzeb_t eb                                            ",    
                   "  left outer join gztd_t td                                 ",
                   "               on td.gztd001 = eb.dzeb006                   ",    
                   " where eb.dzeb001 = '",ms_table_name.toLowerCase(),"'       ",    
                   " order by 1                                                 "
      &else 
      LET ls_sql = "SELECT EV.DZEV005       COLUMN_NAME,                             ",
                   "       EV.DZEV006       ISKEY,                                   ",
                   "       EV.DZEV007       NULLABLE,                                ",
                   "       EV.DZEV008       DATA_TYPE,                               ",
                   "       EV.DZEV009       DATA_LENGTH,                             ",
                   "       TRIM(DECODE(EV.DZEV021,'',NULL,EV.DZEV021)) DATA_DEFAULT  ",
                   "  FROM DZEV_T EV                                                 ",
                   " WHERE EV.DZEV001 = '",ms_owner_name.toUpperCase(),"'            ",
                   "   AND EV.DZEV002 = '",ms_table_name.toUpperCase(),"'            ",
                   "   AND EV.DZEV003 = '",ms_revision,"'                            ",
                   "   AND EV.DZEV004 = 'TableDesigner'                              ",
                   "   AND EV.DZEV018 = '",ms_alm_version,"'                         ",
                   "MINUS                                                            ",
                   "SELECT UPPER(EB.DZEB002)                      COLUMN_NAME,       ", 
                   "       EB.DZEB004                             ISKEY,             ", 
                   "       DECODE(EB.DZEB004,'Y','N','N','Y','Y') NULLABLE,          ", 
                   "       UPPER(TD.GZTD003)                      DATA_TYPE,         ", 
                   "       REPLACE(TO_CHAR(TO_NUMBER(REPLACE(REPLACE(                ", 
                   "       DECODE(                                                   ", 
                   "              UPPER(TD.GZTD003),                                 ", 
                   "              'BLOB','4000',                                     ", 
                   "              'CLOB','4000',                                     ", 
                   "              'DATE','7',                                        ", 
                   "              'TIMESTAMP',DECODE(TD.GZTD008,'0','7','11'),       ",
                   "              TD.GZTD008                                         ", 
                   "             ),',','.'),' ',''))),'.',',')    DATA_LENGTH,       ",
                   "       TRIM(DECODE(EB.DZEB012,'',NULL,EB.DZEB012))  DATA_DEFAULT ",   
                   "  FROM DZEB_T EB                                                 ", 
                   "       LEFT OUTER JOIN GZTD_T TD ON TD.GZTD001 = EB.DZEB006      ", 
                   " WHERE EB.DZEB001 = '",ms_table_name.toLowerCase(),"'            ",
                   " ORDER BY 1                                                      "
      &endif             
      &endif             
    END IF
  END IF   
                
  PREPARE lpre_get_column_forward_records FROM ls_sql
  DECLARE lcur_get_column_forward_records CURSOR FOR lpre_get_column_forward_records

  OPEN lcur_get_column_forward_records
  FOREACH lcur_get_column_forward_records INTO lo_column_forward_records[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_column_forward_records

  FREE lpre_get_column_forward_records
  FREE lcur_get_column_forward_records  
  -------------------------------------------------------------------------------------------------
  
  LET lo_return.* = lo_column_forward_records.*
  
  RETURN lo_return
  
END FUNCTION 

FUNCTION sadzi140_rdif_get_column_backward_records(p_owner,p_table_name)
DEFINE
  p_owner      STRING,
  p_table_name STRING
DEFINE
  lo_column_backward_records DYNAMIC ARRAY OF T_COLUMN_DATA
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_column_name STRING,
  ls_sql         STRING,
  ls_exec_sql    STRING,
  ls_ora_coldef  STRING, #20240529
  li_rec_cnt     INTEGER,
  ls_db_type     STRING, #20240606
  ls_alter_table VARCHAR(30)
DEFINE
  lo_return DYNAMIC ARRAY OF T_COLUMN_DATA

  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name
  
  LET li_rec_cnt = 1

  ------------------------------------------------------------------------------------------------
  -- Column backward
  ------------------------------------------------------------------------------------------------

  IF ms_view_type = cs_view_type_db THEN
    &ifdef is_msv
    LET ls_sql = "select atc.column_name                 column_name,       ",                           
                 "       iif(atc.nullable='Y','N','Y')   iskey,             ",                           
                 "       atc.nullable                    nullable,          ",                           
                 "       atc.data_type                   data_type,         ",   
                 "       atc.data_length                 data_length,       ",                  
                 "       atc.data_default                data_default       ",      
                 "  from dba_tab_columns atc                                ",                                
                 "  left outer join dba_cons_columns acc                    ",   
                 "               on acc.owner       = atc.owner             ",        
                 "              and acc.table_name  = atc.table_name        ",        
                 "              and acc.column_name = atc.column_name       ",        
                 "              and acc.constraint_name like '%pk'          ",        
                 " where 1=1                                                ",                                
                 "   and atc.owner      = '",ls_owner.toLowerCase(),"'      ",                                
                 "   and atc.table_name = '",ls_table_name.toLowerCase(),"' ",                                
                 "except                                                    ",                                 
                 "select ev.dzev005       column_name,                      ",                                
                 "       ev.dzev006       iskey,                            ",                                
                 "       ev.dzev007       nullable,                         ",                                
                 "       ev.dzev008       data_type,                        ",                                
                 "       ev.dzev009       data_length,                      ",                                
                 "       trim(ev.dzev021) data_default                      ",                                
                 "  from dzev_t ev                                          ",                                
                 " where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'     ",                                
                 "   and ev.dzev002 = '",ms_table_name.toLowerCase(),"'     ",                                
                 "   and ev.dzev003 = '",ms_revision,"'                     ",                                
                 "   and ev.dzev004 = 'TableDesigner'                       ",                                
                 "   and ev.dzev018 = '",ms_alm_version,"'                  ",                                
                 " order by 1                                               "                                
    &else 
    &ifdef is_pgs
    LET ls_sql = "select atc.column_name                 column_name,                         ", 
                 "       case                                                                 ", 
                 "         when atc.nullable='Y' then 'N'                                     ", 
                 "         else 'Y'                                                           ", 
                 "       end                             iskey,                               ", 
                 "       atc.nullable                    nullable,                            ", 
                 "       atc.data_type                   data_type,                           ", 
                 "       case                                                                 ",
                 "         when atc.data_type = 'text' then null                              ",
                 "         when atc.data_type = 'bytea' then null                             ",
                 "         when atc.data_type = 'date' then null                              ",
                 "         when atc.data_type = 'int2' then null                              ",
                 "         when atc.data_type = 'int4' then null                              ",
                 "         when atc.data_type = 'int8' then null                              ",
                 "         when atc.data_type = 'decimal' then                                ",
                 "           atc.data_precision::integer                                      ",
                 "       else                                                                 ",
                 "         atc.data_length::integer                                           ",
                 "       end::integer                    data_length,                         ",
                 "       case                                                                 ",
                 "         when atc.data_default is null then null                            ",
                 "         when atc.data_default = 'NULL' then null                           ",
                 "         when atc.data_default = 'null' then null                           ",
                 "         when atc.data_default like 'CURRENT%' then lower(atc.data_default) ",
                 "         when atc.data_default = '\"sysdate\"()' then 'sysdate'             ",
                 "         when atc.data_default = 'SYSDATE' then 'sysdate'                   ",
                 "       else                                                                 ",
                 "         atc.data_default                                                   ",
                 "       end as data_default                                                  ",
                 "  from dba_tab_columns atc                                                  ", 
                 "  left outer join dba_cons_columns acc                                      ", 
                 "               on acc.owner       = atc.owner                               ", 
                 "              and acc.table_name  = atc.table_name                          ", 
                 "              and acc.column_name = atc.column_name                         ", 
                 "              and acc.constraint_name like '%pk'                            ", 
                 " where 1=1                                                                  ", 
                 "   and atc.owner      = '",ls_owner.toLowerCase(),"'                        ", 
                 "   and atc.table_name = '",ls_table_name.toLowerCase(),"'                   ", 
                 "except                                                                      ", 
                 "select ev.dzev005           column_name,                                    ",     
                 "       ev.dzev006           iskey,                                          ",     
                 "       ev.dzev007           nullable,                                       ",     
                 "       ev.dzev008           data_type,                                      ",     
                 "       ev.dzev009::integer  data_length,                                    ",     
                 "       trim(ev.dzev021)     data_default                                    ",     
                 "  from dzev_t ev                                                            ", 
                 " where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'                       ", 
                 "   and ev.dzev002 = '",ms_table_name.toLowerCase(),"'                       ", 
                 "   and ev.dzev003 = '",ms_revision,"'                                       ",
                 "   and ev.dzev004 = 'TableDesigner'                                         ",
                 "   and ev.dzev018 = '",ms_alm_version,"'                                    ",
                 " order by 1                                                                 "
    &else 
    LET ls_db_type = sadzi140_util_get_db_type()
    CASE ls_db_type
      WHEN cs_db_driver_type_kdb
        LET ls_ora_coldef = " CASE                                                                 ",
                            "   WHEN ATC.DATA_DEFAULT IS NULL THEN NULL                            ",
                            "   WHEN ATC.DATA_DEFAULT = 'NULL' THEN NULL                           ",
                            "   WHEN ATC.DATA_DEFAULT = 'null' THEN NULL                           ",
                            "   WHEN ATC.DATA_DEFAULT LIKE 'CURRENT%' THEN LOWER(ATC.DATA_DEFAULT) ",
                            "   WHEN ATC.DATA_DEFAULT = '\"sysdate\"()' then 'sysdate'             ",
                            "   WHEN ATC.DATA_DEFAULT = 'SYSDATE' then 'sysdate'                   ",
                            " ELSE                                                                 ",
                            "   ATC.DATA_DEFAULT                                                   ",
                            " END AS DATA_DEFAULT                                                  "
      WHEN cs_db_driver_type_dmg 
        LET ls_ora_coldef = " CASE                                                                 ",
                            "   WHEN ATC.DATA_DEFAULT IS NULL THEN NULL                            ",
                            "   WHEN ATC.DATA_DEFAULT = 'NULL' THEN NULL                           ",
                            "   WHEN ATC.DATA_DEFAULT = 'null' THEN NULL                           ",
                            "   WHEN ATC.DATA_DEFAULT LIKE 'CURRENT%' THEN LOWER(ATC.DATA_DEFAULT) ",
                            "   WHEN ATC.DATA_DEFAULT = '\"sysdate\"()' then 'sysdate'             ",
                            "   WHEN ATC.DATA_DEFAULT = 'SYSDATE' then 'sysdate'                   ",
                            " ELSE                                                                 ",
                            "   ATC.DATA_DEFAULT                                                   ",
                            " END AS DATA_DEFAULT                                                  "
    OTHERWISE  
      LET ls_ora_coldef = " DECODE(TRIM(DS.GET_COL_DEFAULT(ATC.TABLE_NAME,ATC.COLUMN_NAME)),'',NULL,          ",
                          "        TRIM(DS.GET_COL_DEFAULT(ATC.TABLE_NAME,ATC.COLUMN_NAME))) DATA_DEFAULT     "
    END CASE
    
    LET ls_sql = "SELECT ATC.COLUMN_NAME                      COLUMN_NAME,                                 ",
                 "       DECODE(ACC.COLUMN_NAME,NULL,'N','Y') ISKEY,                                       ",
                 "       ATC.NULLABLE                         NULLABLE,                                    ",
                 "       DECODE(INSTRB(ATC.DATA_TYPE,'TIMESTAMP'),1,'TIMESTAMP',ATC.DATA_TYPE) DATA_TYPE,  ",
                 "       REPLACE(TO_CHAR(DECODE(                                                           ",
                 "               ATC.DATA_TYPE,                                                            ",
                 "               'BLOB',4000,                                                              ", #20240529
                 "               'CLOB',4000,                                                              ", #20240529
                 "               'INT2',NULL,                                                              ", #20240529
                 "               'INT4',NULL,                                                              ", #20240529
                 "               'INT8',NULL,                                                              ", #20240529
                 "               'NUMBER',DECODE(                                                          ",
                 "                                NVL(ATC.DATA_SCALE,'0'),                                 ",
                 "                                '0',ATC.DATA_PRECISION,                                  ",
                 "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE                  ",
                 "                              ),                                                         ",
                 "               'NUMERIC',DECODE(                                                         ", #20240529 begin
                 "                                NVL(ATC.DATA_SCALE,'0'),                                 ",
                 "                                '0',ATC.DATA_PRECISION,                                  ",
                 "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE                  ", 
                 "                              ),                                                         ", #20240529 end
                 "               'DATE',7,                                                                 ", #20240529
                 "               'TIMESTAMP(0)',7,                                                         ", #20240529
                 "               'TIMESTAMP(5)',11,                                                        ", #20240529
                 "               'TIMESTAMP(6)',11,                                                        ", #20240529
                 "               'TIMESTAMP',DECODE(                                                       ",
                 "                                   ATC.DATA_SCALE,0,7,8,7,11                             ", #20240529
                 "                                 ),                                                      ",                
                 "               'DATE',DECODE(                                                            ",
                 "                              ATC.DATA_LENGTH,3,7,ATC.DATA_LENGTH                        ",
                 "                            ),                                                           ",                
                 "               ATC.DATA_LENGTH                                                           ",
                 "             )),'.',',')                    DATA_LENGTH,                                 ",
                 ls_ora_coldef,
                 "  FROM DBA_TAB_COLUMNS ATC                                                               ",
                 "  LEFT OUTER JOIN DBA_CONS_COLUMNS ACC ON ACC.OWNER       = ATC.OWNER                    ",
                 "                                      AND ACC.TABLE_NAME  = ATC.TABLE_NAME               ",
                 "                                      AND ACC.COLUMN_NAME = ATC.COLUMN_NAME              ",
                 "                                      AND ACC.CONSTRAINT_NAME LIKE '%PK'                 ",
                 " WHERE 1=1                                                                               ",
                 "   AND ATC.OWNER      = '",ls_owner.toUpperCase(),"'                                     ",
                 "   AND ATC.TABLE_NAME = '",ls_table_name.toUpperCase(),"'                                ",
                 "MINUS                                                                                    ",
                 "SELECT EV.DZEV005       COLUMN_NAME,                                                     ",
                 "       EV.DZEV006       ISKEY,                                                           ",
                 "       EV.DZEV007       NULLABLE,                                                        ",
                 "       EV.DZEV008       DATA_TYPE,                                                       ",
                 "       EV.DZEV009       DATA_LENGTH,                                                     ",
                 "       TRIM(DECODE(EV.DZEV021,'',NULL,EV.DZEV021)) DATA_DEFAULT                          ",
                 "  FROM DZEV_T EV                                                                         ",
                 " WHERE EV.DZEV001 = '",ms_owner_name.toUpperCase(),"'                                    ",
                 "   AND EV.DZEV002 = '",ms_table_name.toUpperCase(),"'                                    ",
                 "   AND EV.DZEV003 = '",ms_revision,"'                                                    ",
                 "   AND EV.DZEV004 = 'TableDesigner'                                                      ",
                 "   AND EV.DZEV018 = '",ms_alm_version,"'                                                 ",
                 " ORDER BY 1                                                                              "
    &endif             
    &endif             
  ELSE 
    IF ms_view_type = cs_view_type_rt THEN
      &ifdef is_msv
      LET ls_sql = 
                   "select eb.dzeb002                      column_name,         ",  
                   "       eb.dzeb004                      iskey,               ",  
                   "       iif(eb.dzeb004='Y','N','Y')     nullable,            ",  
                   "       td.gztd003                      data_type,           ",  
                   "       case td.gztd003                                      ",  
                   "         when 'varbinary'                                   ",  
                   "           then 'max'                                       ",  
                   "         when 'nvarchar'                                    ",  
                   "           then iif(td.gztd009='TEXT','max',td.gztd008)     ",  
                   "       else                                                 ",  
                   "         td.gztd008                                         ",  
                   "       end                             data_length,         ",  
                   "       trim(eb.dzeb012)                data_default         ",  
                   "  from dzeb_t eb                                            ",    
                   "       left outer join gztd_t td on td.gztd001 = eb.dzeb006 ",    
                   " where eb.dzeb001 = '",ms_table_name.toLowerCase(),"'       ",    
                   "except                                                      ",     
                   "select ev.dzev005       column_name,                        ",    
                   "       ev.dzev006       iskey,                              ",    
                   "       ev.dzev007       nullable,                           ",    
                   "       ev.dzev008       data_type,                          ",    
                   "       ev.dzev009       data_length,                        ",    
                   "       trim(ev.dzev021) data_default                        ",    
                   "  from dzev_t ev                                            ",    
                   " where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'       ",    
                   "   and ev.dzev002 = '",ms_table_name.toLowerCase(),"'       ",    
                   "   and ev.dzev003 = '",ms_revision,"'                       ",    
                   "   and ev.dzev004 = 'TableDesigner'                         ",    
                   "   and ev.dzev018 = '",ms_alm_version,"'                    ",    
                   " order by 1                                                 "
      &else 
      &ifdef is_pgs
      LET ls_sql = 
                   "select eb.dzeb002                      column_name,         ",  
                   "       eb.dzeb004                      iskey,               ",  
                   "       case                                                 ", 
                   "         when eb.dzeb004='Y'                                ",
                   "           then 'N'                                         ",
                   "       else                                                 ", 
                   "         'Y'                                                ",
                   "       end                             nullable,            ",  
                   "       td.gztd003                      data_type,           ",  
                   "       td.gztd008                      data_length,         ",  
                   "       trim(eb.dzeb012)                data_default         ",  
                   "  from dzeb_t eb                                            ",    
                   "       left outer join gztd_t td on td.gztd001 = eb.dzeb006 ",    
                   " where eb.dzeb001 = '",ms_table_name.toLowerCase(),"'       ",    
                   "except                                                      ",     
                   "select ev.dzev005       column_name,                        ",    
                   "       ev.dzev006       iskey,                              ",    
                   "       ev.dzev007       nullable,                           ",    
                   "       ev.dzev008       data_type,                          ",    
                   "       ev.dzev009       data_length,                        ",    
                   "       trim(ev.dzev021) data_default                        ",    
                   "  from dzev_t ev                                            ",    
                   " where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'       ",    
                   "   and ev.dzev002 = '",ms_table_name.toLowerCase(),"'       ",    
                   "   and ev.dzev003 = '",ms_revision,"'                       ",    
                   "   and ev.dzev004 = 'TableDesigner'                         ",    
                   "   and ev.dzev018 = '",ms_alm_version,"'                    ",    
                   " order by 1                                                 "
      &else 
      LET ls_sql = "SELECT UPPER(EB.DZEB002)                      COLUMN_NAME,       ", 
                   "       EB.DZEB004                             ISKEY,             ", 
                   "       DECODE(EB.DZEB004,'Y','N','N','Y','Y') NULLABLE,          ", 
                   "       UPPER(TD.GZTD003)                      DATA_TYPE,         ", 
                   "       REPLACE(TO_CHAR(TO_NUMBER(REPLACE(REPLACE(                ", 
                   "       DECODE(                                                   ", 
                   "              UPPER(TD.GZTD003),                                 ", 
                   "              'BLOB','4000',                                     ", 
                   "              'CLOB','4000',                                     ", 
                   "              'DATE','7',                                        ", 
                   "              'TIMESTAMP',DECODE(TD.GZTD008,'0','7','11'),       ",
                   "              TD.GZTD008                                         ", 
                   "             ),',','.'),' ',''))),'.',',')    DATA_LENGTH,       ",
                   "       TRIM(DECODE(EB.DZEB012,'',NULL,EB.DZEB012))  DATA_DEFAULT ",
                   "  FROM DZEB_T EB                                                 ", 
                   "       LEFT OUTER JOIN GZTD_T TD ON TD.GZTD001 = EB.DZEB006      ", 
                   " WHERE EB.DZEB001 = '",ms_table_name.toLowerCase(),"'            ",
                   "MINUS                                                            ",
                   "SELECT EV.DZEV005       COLUMN_NAME,                             ",
                   "       EV.DZEV006       ISKEY,                                   ",
                   "       EV.DZEV007       NULLABLE,                                ",
                   "       EV.DZEV008       DATA_TYPE,                               ",
                   "       EV.DZEV009       DATA_LENGTH,                             ",
                   "       TRIM(DECODE(EV.DZEV021,'',NULL,EV.DZEV021)) DATA_DEFAULT  ",
                   "  FROM DZEV_T EV                                                 ",
                   " WHERE EV.DZEV001 = '",ms_owner_name.toUpperCase(),"'            ",
                   "   AND EV.DZEV002 = '",ms_table_name.toUpperCase(),"'            ",
                   "   AND EV.DZEV003 = '",ms_revision,"'                            ",
                   "   AND EV.DZEV004 = 'TableDesigner'                              ",
                   "   AND EV.DZEV018 = '",ms_alm_version,"'                         ",
                   " ORDER BY 1                                                      "
      &endif             
      &endif             
    END IF
  END IF   
                
  PREPARE lpre_get_column_backward_records FROM ls_sql
  DECLARE lcur_get_column_backward_records CURSOR FOR lpre_get_column_backward_records

  OPEN lcur_get_column_backward_records
  FOREACH lcur_get_column_backward_records INTO lo_column_backward_records[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_column_backward_records
  
  FREE lpre_get_column_backward_records
  FREE lcur_get_column_backward_records  
  -------------------------------------------------------------------------------------------------
  
  LET lo_return.* = lo_column_backward_records.*
  
  RETURN lo_return
  
END FUNCTION 

FUNCTION sadzi140_rdif_get_index_forward_records(p_owner,p_table_name)
DEFINE
  p_owner      STRING,
  p_table_name STRING
DEFINE
  lo_index_forward_records DYNAMIC ARRAY OF T_INDEX_DATA
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_index_name STRING,
  ls_sql         STRING,
  ls_exec_sql    STRING,
  li_rec_cnt     INTEGER,
  ls_alter_table VARCHAR(30)
DEFINE
  lo_return DYNAMIC ARRAY OF T_INDEX_DATA  

  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name
  
  LET li_rec_cnt = 1

  ------------------------------------------------------------------------------------------------
  -- Index forward
  ------------------------------------------------------------------------------------------------
  IF ms_view_type = cs_view_type_db THEN
    &ifdef is_msv
    LET ls_sql = "select ew.dzew001 table_name,                             ",                                          
                 "       ew.dzew004 index_name,                             ",                                          
                 "       case upper(ew.dzew005)                             ",
                 "         when 'N' then 'NORMAL'                           ",
                 "         when 'U' then 'UNIQUE'                           ",
                 "       else                                               ",
                 "         upper(ew.dzew005)                                ",
                 "       end index_type,                                    ",                           
                 "       replace(ew.dzew006,' ','') column_name             ",                                          
                 "  from dzew_t ew                                          ",                                                 
                 " where ew.dzew001 = '",ms_table_name.toLowerCase(),"'     ",                                                 
                 "   and ew.dzew002 = '",ms_revision,"'                     ",                                                 
                 "   and ew.dzew003 = 'Index'                               ",                                                 
                 "   and ew.dzew010 = '",ms_alm_version,"'                  ",                                                 
                 "   and coalesce(ew.dzewstus,'Y') = 'Y'                    ",                                                      
                 "except                                                    ",
                 "select ais.table_name table_name,                         ",                                                 
                 "       ais.index_name index_name,                         ",                                                 
                 "       case upper(substring(ais.uniqueness,1,1))          ",
                 "         when 'N' then 'NORMAL'                           ",
                 "         when 'U' then 'UNIQUE'                           ",
                 "       else                                               ",
                 "         ais.index_type                                   ",
                 "       end index_type,                                    ",
                 "       ais.column_names column_name                       ",  
                 "  from dba_indexes ais                                    ",  
                 " where ais.owner      = '",ls_owner.toLowerCase(),"'      ",                                                   
                 "   and ais.table_name = '",ls_table_name.toLowerCase(),"' "                                                   
    &else  
    &ifdef is_pgs
    LET ls_sql = "select ew.dzew001 table_name,                             ",                                          
                 "       ew.dzew004 index_name,                             ",                                          
                 "       case upper(ew.dzew005)                             ",
                 "         when 'N' then 'NORMAL'                           ",
                 "         when 'U' then 'UNIQUE'                           ",
                 "       else                                               ",
                 "         upper(ew.dzew005)                                ",
                 "       end index_type,                                    ",                           
                 "       replace(ew.dzew006,' ','') column_name             ",                                          
                 "  from dzew_t ew                                          ",                                                 
                 " where ew.dzew001 = '",ms_table_name.toLowerCase(),"'     ",                                                 
                 "   and ew.dzew002 = '",ms_revision,"'                     ",                                                 
                 "   and ew.dzew003 = 'Index'                               ",                                                 
                 "   and ew.dzew010 = '",ms_alm_version,"'                  ",                                                 
                 "   and coalesce(ew.dzewstus,'Y') = 'Y'                    ",                                                      
                 "except                                                    ",
                 "select ais.table_name table_name,                         ",                                                 
                 "       ais.index_name index_name,                         ",                                                 
                 "       case upper(substring(ais.uniqueness,1,1))          ",
                 "         when 'N' then 'NORMAL'                           ",
                 "         when 'U' then 'UNIQUE'                           ",
                 --"       else                                               ",
                 --"         ais.index_type                                   ",
                 "       end index_type,                                    ",
                 "       string_agg(aic.column_name,','                     ",
                 "         order by aic.column_position) column_name        ",  
                 "  from dba_indexes ais                                    ",  
                 " inner join dba_ind_columns aic                           ",
                 "         on aic.table_owner = ais.table_owner             ",
                 "        and aic.table_name  = ais.table_name              ",
                 "        and aic.index_name  = ais.index_name              ",
                 " where ais.owner      = '",ls_owner.toLowerCase(),"'      ",                                                   
                 "   and ais.table_name = '",ls_table_name.toLowerCase(),"' ",                                                   
                 " group by ais.table_name,                                 ",                                                 
                 "          ais.index_name,                                 ",                                                 
                 "          case upper(substring(ais.uniqueness,1,1))       ",
                 "            when 'N' then 'NORMAL'                        ",
                 "            when 'U' then 'UNIQUE'                        ",
                 --"          else                                            ",
                 --"            ais.index_type                                ",
                 "          end                                             "                                                   
    &else  
    LET ls_sql = "SELECT UPPER(EW.DZEW001) TABLE_NAME,                                                                         ",
                 "       UPPER(EW.DZEW004) INDEX_NAME,                                                                         ",
                 "       DECODE(UPPER(EW.DZEW005),                                                                             ",
                 "              'N','NORMAL',                                                                                  ",
                 "              'U','UNIQUE',                                                                                  ",
                 "              UPPER(EW.DZEW005)) INDEX_TYPE,                                                                 ",
                 "       UPPER(REPLACE(EW.DZEW006,' ','')) COLUMN_NAME                                                         ",
                 "  FROM DZEW_T EW                                                                                             ",
                 " WHERE EW.DZEW001 = '",ms_table_name.toLowerCase(),"'                                                        ",
                 "   AND EW.DZEW002 = '",ms_revision,"'                                                                        ",
                 "   AND EW.DZEW003 = 'Index'                                                                                  ",
                 "   AND EW.DZEW010 = '",ms_alm_version,"'                                                                     ",
                 "   AND NVL(EW.DZEWSTUS,'Y') = 'Y'                                                                            ", #20170912
                 " MINUS                                                                                                       ",
                 "SELECT AIS.TABLE_NAME TABLE_NAME,                                                                            ",
                 "       NVL(DC.CONSTRAINT_NAME,AIS.INDEX_NAME) INDEX_NAME,                                                    ", --20230807
                 "       DECODE(AIS.UNIQUENESS,'UNIQUE','UNIQUE',AIS.INDEX_TYPE) INDEX_TYPE,                                   ",
                 "       LISTAGG(AICS.COLUMN_NAME,',') WITHIN GROUP (ORDER BY AIS.INDEX_NAME,AICS.COLUMN_POSITION) COLUMN_NAME ",
                 "  FROM DBA_INDEXES AIS,DBA_IND_COLUMNS AICS,DBA_CONSTRAINTS DC                                               ", --20230807    
                 " WHERE AIS.OWNER        = '",ls_owner.toUpperCase(),"'                                                       ",
                 "   AND AIS.TABLE_NAME   = '",ls_table_name.toUpperCase(),"'                                                  ",
                 "   AND AIS.INDEX_TYPE <> 'CLUSTER'                                                                           ", --20230807    
                 "   AND AICS.TABLE_OWNER = AIS.TABLE_OWNER                                                                    ",
                 "   AND AICS.TABLE_NAME  = AIS.TABLE_NAME                                                                     ",
                 "   AND AICS.INDEX_NAME  = AIS.INDEX_NAME                                                                     ",
                 "   AND DC.OWNER      (+)= AIS.OWNER                                                                          ", --20230807
                 "   AND DC.INDEX_NAME (+)= AIS.INDEX_NAME                                                                     ", --20230807
                 "   AND DC.TABLE_NAME (+)= AIS.TABLE_NAME                                                                     ", --20230807
                 " GROUP BY AIS.TABLE_NAME,NVL(DC.CONSTRAINT_NAME,AIS.INDEX_NAME),                                             ", --20230807
                 "          DECODE(AIS.UNIQUENESS,'UNIQUE','UNIQUE',AIS.INDEX_TYPE)                                            "
    &endif                 
    &endif                 
  ELSE 
    IF ms_view_type = cs_view_type_rt THEN
      &ifdef is_msv
      LET ls_sql = "select ew.dzew001                 table_name,         ",         
                   "       ew.dzew004                 index_name,         ",         
                   "       case upper(ew.dzew005)                         ", 
                   "         when 'N' then 'NORMAL'                       ", 
                   "         when 'U' then 'UNIQUE'                       ", 
                   "       else                                           ", 
                   "         upper(ew.dzew005)                            ", 
                   "       end                        index_type,         ", 
                   "       replace(ew.dzew006,' ','') column_name         ", 
                   "  from dzew_t ew                                      ", 
                   " where ew.dzew001 = '",ms_table_name.toLowerCase(),"' ", 
                   "   and ew.dzew002 = '",ms_revision,"'                 ", 
                   "   and ew.dzew003 = 'Index'                           ", 
                   "   and ew.dzew010 = '",ms_alm_version,"'              ", 
                   "   and coalesce(ew.dzewstus,'Y') = 'Y'                ",     
                   "except                                                ", 
                   "select ec.dzec001                  table_name,        ",          
                   "       ec.dzec002                  index_name,        ",          
                   "       case upper(ec.dzec003)                         ", 
                   "         when 'N' then 'NORMAL'                       ", 
                   "         when 'U' then 'UNIQUE'                       ", 
                   "       else                                           ", 
                   "         upper(ec.dzec003)                            ", 
                   "       end                         index_type,        ",   
                   "       replace(ec.dzec004,' ','')  column_name        ",  
                   "  from dzec_t ec                                      ",  
                   " where ec.dzec001 = '",ls_table_name.toLowerCase(),"' ",  
                   "   and coalesce(ec.dzecstus,'Y') = 'Y'                "      
      &else 
      &ifdef is_pgs
      LET ls_sql = "select ew.dzew001                 table_name,         ",         
                   "       ew.dzew004                 index_name,         ",         
                   "       case upper(ew.dzew005)                         ", 
                   "         when 'N' then 'NORMAL'                       ", 
                   "         when 'U' then 'UNIQUE'                       ", 
                   "       else                                           ", 
                   "         upper(ew.dzew005)                            ", 
                   "       end                        index_type,         ", 
                   "       replace(ew.dzew006,' ','') column_name         ", 
                   "  from dzew_t ew                                      ", 
                   " where ew.dzew001 = '",ms_table_name.toLowerCase(),"' ", 
                   "   and ew.dzew002 = '",ms_revision,"'                 ", 
                   "   and ew.dzew003 = 'Index'                           ", 
                   "   and ew.dzew010 = '",ms_alm_version,"'              ", 
                   "   and coalesce(ew.dzewstus,'Y') = 'Y'                ",     
                   "except                                                ", 
                   "select ec.dzec001                  table_name,        ",          
                   "       ec.dzec002                  index_name,        ",          
                   "       case upper(ec.dzec003)                         ", 
                   "         when 'N' then 'NORMAL'                       ", 
                   "         when 'U' then 'UNIQUE'                       ", 
                   "       else                                           ", 
                   "         upper(ec.dzec003)                            ", 
                   "       end                         index_type,        ",   
                   "       replace(ec.dzec004,' ','')  column_name        ",  
                   "  from dzec_t ec                                      ",  
                   " where ec.dzec001 = '",ls_table_name.toLowerCase(),"' ",  
                   "   and coalesce(ec.dzecstus,'Y') = 'Y'                "      
      &else 
      LET ls_sql = "SELECT UPPER(EW.DZEW001) TABLE_NAME,                  ",
                   "       UPPER(EW.DZEW004) INDEX_NAME,                  ",
                   "       DECODE(UPPER(EW.DZEW005),                      ",
                   "              'N','NORMAL',                           ",
                   "              'U','UNIQUE',                           ",
                   "              UPPER(EW.DZEW005)) INDEX_TYPE,          ",
                   "       UPPER(REPLACE(EW.DZEW006,' ','')) COLUMN_NAME  ",
                   "  FROM DZEW_T EW                                      ",
                   " WHERE EW.DZEW001 = '",ms_table_name.toLowerCase(),"' ",
                   "   AND EW.DZEW002 = '",ms_revision,"'                 ",
                   "   AND EW.DZEW003 = 'Index'                           ",
                   "   AND EW.DZEW010 = '",ms_alm_version,"'              ",
                   "   AND NVL(EW.DZEWSTUS,'Y') = 'Y'                     ", #20170912
                   " MINUS                                                ",
                   "SELECT UPPER(EC.DZEC001) TABLE_NAME,                  ",
                   "       UPPER(EC.DZEC002) INDEX_NAME,                  ",
                   "       DECODE(UPPER(EC.DZEC003),                      ",
                   "              'N','NORMAL',                           ",
                   "              'U','UNIQUE',                           ",
                   "              UPPER(EC.DZEC003)) INDEX_TYPE,          ",                                             
                   "       UPPER(REPLACE(EC.DZEC004,' ','')) COLUMN_NAME  ",
                   "  FROM DZEC_T EC                                      ",
                   " WHERE EC.DZEC001 = '",ls_table_name.toLowerCase(),"' ",
                   "   AND NVL(EC.DZECSTUS,'Y') = 'Y'                     "  #20170912
      &endif             
      &endif             
    END IF
  END IF   
                
  PREPARE lpre_get_index_forward_records FROM ls_sql
  DECLARE lcur_get_index_forward_records CURSOR FOR lpre_get_index_forward_records

  OPEN lcur_get_index_forward_records
  FOREACH lcur_get_index_forward_records INTO lo_index_forward_records[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_index_forward_records
  
  FREE lpre_get_index_forward_records
  FREE lcur_get_index_forward_records  
  -------------------------------------------------------------------------------------------------
  
  LET lo_return.* = lo_index_forward_records.*
  
  RETURN lo_return
  
END FUNCTION 

FUNCTION sadzi140_rdif_get_index_backward_records(p_owner,p_table_name)
DEFINE
  p_owner      STRING,
  p_table_name STRING
DEFINE
  lo_index_backward_records DYNAMIC ARRAY OF T_INDEX_DATA
DEFINE
  ls_owner       STRING,
  ls_table_name  STRING,
  ls_index_name STRING,
  ls_sql         STRING,
  ls_exec_sql    STRING,
  li_rec_cnt     INTEGER,
  ls_alter_table VARCHAR(30)
DEFINE
  lo_return DYNAMIC ARRAY OF T_INDEX_DATA  

  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name
  
  LET li_rec_cnt = 1

  ------------------------------------------------------------------------------------------------
  -- Index backward
  ------------------------------------------------------------------------------------------------
  IF ms_view_type = cs_view_type_db THEN
    &ifdef is_msv
    LET ls_sql = "select ais.table_name table_name,                         ",                                                 
                 "       ais.index_name index_name,                         ",                                                 
                 "       case upper(substring(ais.uniqueness,1,1))          ",
                 "         when 'N' then 'NORMAL'                           ",
                 "         when 'U' then 'UNIQUE'                           ",
                 "       else                                               ",
                 "         ais.index_type                                   ",
                 "       end index_type,                                    ",
                 "       ais.column_names column_name                       ",  
                 "  from dba_indexes ais                                    ",  
                 " where ais.owner      = '",ls_owner.toLowerCase(),"'      ",                                                   
                 "   and ais.table_name = '",ls_table_name.toLowerCase(),"' ",                                                   
                 "except                                                    ",
                 "select ew.dzew001 table_name,                             ",                                          
                 "       ew.dzew004 index_name,                             ",                                          
                 "       case upper(ew.dzew005)                             ",
                 "         when 'N' then 'NORMAL'                           ",
                 "         when 'U' then 'UNIQUE'                           ",
                 "       else                                               ",
                 "         upper(ew.dzew005)                                ",
                 "       end index_type,                                    ",                           
                 "       replace(ew.dzew006,' ','') column_name             ",                                          
                 "  from dzew_t ew                                          ",                                                 
                 " where ew.dzew001 = '",ms_table_name.toLowerCase(),"'     ",                                                 
                 "   and ew.dzew002 = '",ms_revision,"'                     ",                                                 
                 "   and ew.dzew003 = 'Index'                               ",                                                 
                 "   and ew.dzew010 = '",ms_alm_version,"'                  ",                                                 
                 "   and coalesce(ew.dzewstus,'Y') = 'Y'                    "                                                      
    &else  
    &ifdef is_pgs
    LET ls_sql = "select ais.table_name table_name,                           ",                                                 
                 "       ais.index_name index_name,                           ",                                                 
                 "       case upper(substring(ais.uniqueness,1,1))            ",
                 "         when 'N' then 'NORMAL'                             ",
                 "         when 'U' then 'UNIQUE'                             ",
                 --"       else                                                 ",
                 --"         ais.index_type                                     ",
                 "       end index_type,                                      ",
                 "       string_agg(aic.column_name,','                       ",
                 "                  order by aic.column_position) column_name ",  
                 "  from dba_indexes ais                                      ",  
                 " inner join dba_ind_columns aic                             ",
                 "         on aic.table_owner = ais.table_owner               ",
                 "        and aic.table_name  = ais.table_name                ",
                 "        and aic.index_name  = ais.index_name                ",
                 " where ais.owner      = '",ls_owner.toLowerCase(),"'        ",                                                   
                 "   and ais.table_name = '",ls_table_name.toLowerCase(),"'   ",                                                   
                 " group by ais.table_name,                                   ",                                                 
                 "          ais.index_name,                                   ",                                                 
                 "          case upper(substring(ais.uniqueness,1,1))         ",
                 "            when 'N' then 'NORMAL'                          ",
                 "            when 'U' then 'UNIQUE'                          ",
                 --"          else                                              ",
                 --"            ais.index_type                                  ",
                 "          end                                               ",
                 "except                                                      ",
                 "select ew.dzew001 table_name,                               ",                                          
                 "       ew.dzew004 index_name,                               ",                                          
                 "       case upper(ew.dzew005)                               ",
                 "         when 'N' then 'NORMAL'                             ",
                 "         when 'U' then 'UNIQUE'                             ",
                 "       else                                                 ",
                 "         upper(ew.dzew005)                                  ",
                 "       end index_type,                                      ",                           
                 "       replace(ew.dzew006,' ','') column_name               ",                                          
                 "  from dzew_t ew                                            ",                                                 
                 " where ew.dzew001 = '",ms_table_name.toLowerCase(),"'       ",                                                 
                 "   and ew.dzew002 = '",ms_revision,"'                       ",                                                 
                 "   and ew.dzew003 = 'Index'                                 ",                                                 
                 "   and ew.dzew010 = '",ms_alm_version,"'                    ",                                                 
                 "   and coalesce(ew.dzewstus,'Y') = 'Y'                      "                                                      
    &else  
    LET ls_sql = "SELECT AIS.TABLE_NAME TABLE_NAME,                                                                            ",
                 "       NVL(DC.CONSTRAINT_NAME,AIS.INDEX_NAME) INDEX_NAME,                                                    ", --20230807
                 "       DECODE(AIS.UNIQUENESS,'UNIQUE','UNIQUE',AIS.INDEX_TYPE) INDEX_TYPE,                                   ",
                 "       LISTAGG(AICS.COLUMN_NAME,',') WITHIN GROUP (ORDER BY AIS.INDEX_NAME,AICS.COLUMN_POSITION) COLUMN_NAME ",
                 "  FROM DBA_INDEXES AIS,DBA_IND_COLUMNS AICS,DBA_CONSTRAINTS DC                                               ", --20230807
                 " WHERE AIS.OWNER        = '",ls_owner.toUpperCase(),"'                                                       ",
                 "   AND AIS.TABLE_NAME   = '",ls_table_name.toUpperCase(),"'                                                  ",
                 "   AND AIS.INDEX_TYPE <> 'CLUSTER'                                                                           ", --20230807    
                 "   AND AICS.TABLE_OWNER = AIS.TABLE_OWNER                                                                    ",
                 "   AND AICS.TABLE_NAME  = AIS.TABLE_NAME                                                                     ",
                 "   AND AICS.INDEX_NAME  = AIS.INDEX_NAME                                                                     ",
                 "   AND DC.OWNER      (+)= AIS.OWNER                                                                          ", --20230807
                 "   AND DC.INDEX_NAME (+)= AIS.INDEX_NAME                                                                     ", --20230807
                 "   AND DC.TABLE_NAME (+)= AIS.TABLE_NAME                                                                     ", --20230807
                 " GROUP BY AIS.TABLE_NAME,NVL(DC.CONSTRAINT_NAME,AIS.INDEX_NAME),                                             ", --20230807
                 "          DECODE(AIS.UNIQUENESS,'UNIQUE','UNIQUE',AIS.INDEX_TYPE)                                            ", --20230807
                 "MINUS                                                                                                        ",
                 "SELECT UPPER(EW.DZEW001) TABLE_NAME,                                                                         ",
                 "       UPPER(EW.DZEW004) INDEX_NAME,                                                                         ",
                 "       DECODE(UPPER(EW.DZEW005),                                                                             ",
                 "              'N','NORMAL',                                                                                  ",
                 "              'U','UNIQUE',                                                                                  ",
                 "              UPPER(EW.DZEW005)) INDEX_TYPE,                                                                 ",
                 "       UPPER(REPLACE(EW.DZEW006,' ','')) COLUMN_NAME                                                         ",
                 "  FROM DZEW_T EW                                                                                             ",
                 " WHERE EW.DZEW001 = '",ms_table_name.toLowerCase(),"'                                                        ",
                 "   AND EW.DZEW002 = '",ms_revision,"'                                                                        ",
                 "   AND EW.DZEW003 = 'Index'                                                                                  ",
                 "   AND EW.DZEW010 = '",ms_alm_version,"'                                                                     ",
                 "   AND NVL(EW.DZEWSTUS,'Y') = 'Y'                                                                            "  #20170912
    &endif             
    &endif             
  ELSE 
    IF ms_view_type = cs_view_type_rt THEN
      &ifdef is_msv
      LET ls_sql = "select ec.dzec001                  table_name,        ",          
                   "       ec.dzec002                  index_name,        ",          
                   "       case upper(ec.dzec003)                         ", 
                   "         when 'N' then 'NORMAL'                       ", 
                   "         when 'U' then 'UNIQUE'                       ", 
                   "       else                                           ", 
                   "         upper(ec.dzec003)                            ", 
                   "       end                         index_type,        ",   
                   "       replace(ec.dzec004,' ','')  column_name        ",  
                   "  from dzec_t ec                                      ",  
                   " where ec.dzec001 = '",ls_table_name.toLowerCase(),"' ",  
                   "   and coalesce(ec.dzecstus,'Y') = 'Y'                ",      
                   "except                                                ", 
                   "select ew.dzew001                 table_name,         ",         
                   "       ew.dzew004                 index_name,         ",         
                   "       case upper(ew.dzew005)                         ", 
                   "         when 'N' then 'NORMAL'                       ", 
                   "         when 'U' then 'UNIQUE'                       ", 
                   "       else                                           ", 
                   "         upper(ew.dzew005)                            ", 
                   "       end                        index_type,         ", 
                   "       replace(ew.dzew006,' ','') column_name         ", 
                   "  from dzew_t ew                                      ", 
                   " where ew.dzew001 = '",ms_table_name.toLowerCase(),"' ", 
                   "   and ew.dzew002 = '",ms_revision,"'                 ", 
                   "   and ew.dzew003 = 'Index'                           ", 
                   "   and ew.dzew010 = '",ms_alm_version,"'              ", 
                   "   and coalesce(ew.dzewstus,'Y') = 'Y'                "     
      &else 
      &ifdef is_pgs
      LET ls_sql = "select ec.dzec001                  table_name,        ",          
                   "       ec.dzec002                  index_name,        ",          
                   "       case upper(ec.dzec003)                         ", 
                   "         when 'N' then 'NORMAL'                       ", 
                   "         when 'U' then 'UNIQUE'                       ", 
                   "       else                                           ", 
                   "         upper(ec.dzec003)                            ", 
                   "       end                         index_type,        ",   
                   "       replace(ec.dzec004,' ','')  column_name        ",  
                   "  from dzec_t ec                                      ",  
                   " where ec.dzec001 = '",ls_table_name.toLowerCase(),"' ",  
                   "   and coalesce(ec.dzecstus,'Y') = 'Y'                ",      
                   "except                                                ", 
                   "select ew.dzew001                 table_name,         ",         
                   "       ew.dzew004                 index_name,         ",         
                   "       case upper(ew.dzew005)                         ", 
                   "         when 'N' then 'NORMAL'                       ", 
                   "         when 'U' then 'UNIQUE'                       ", 
                   "       else                                           ", 
                   "         upper(ew.dzew005)                            ", 
                   "       end                        index_type,         ", 
                   "       replace(ew.dzew006,' ','') column_name         ", 
                   "  from dzew_t ew                                      ", 
                   " where ew.dzew001 = '",ms_table_name.toLowerCase(),"' ", 
                   "   and ew.dzew002 = '",ms_revision,"'                 ", 
                   "   and ew.dzew003 = 'Index'                           ", 
                   "   and ew.dzew010 = '",ms_alm_version,"'              ", 
                   "   and coalesce(ew.dzewstus,'Y') = 'Y'                "     
      &else 
      LET ls_sql = "SELECT UPPER(EC.DZEC001) TABLE_NAME,                   ",
                   "       UPPER(EC.DZEC002) INDEX_NAME,                   ",
                   "       DECODE(UPPER(EC.DZEC003),                       ",
                   "              'N','NORMAL',                            ",
                   "              'U','UNIQUE',                            ",
                   "              UPPER(EC.DZEC003)) INDEX_TYPE,           ",                                             
                   "       UPPER(REPLACE(EC.DZEC004,' ','')) COLUMN_NAME   ",
                   "  FROM DZEC_T EC                                       ",
                   " WHERE EC.DZEC001 = '",ls_table_name.toLowerCase(),"'  ",
                   "   AND NVL(EC.DZECSTUS,'Y') = 'Y'                      ",  #20170912
                   " MINUS                                                 ",
                   "SELECT UPPER(EW.DZEW001) TABLE_NAME,                   ",
                   "       UPPER(EW.DZEW004) INDEX_NAME,                   ",
                   "       DECODE(UPPER(EW.DZEW005),                       ",
                   "              'N','NORMAL',                            ",
                   "              'U','UNIQUE',                            ",
                   "              UPPER(EW.DZEW005)) INDEX_TYPE,           ",
                   "       UPPER(REPLACE(EW.DZEW006,' ','')) COLUMN_NAME   ",
                   "  FROM DZEW_T EW                                       ",
                   " WHERE EW.DZEW001 = '",ms_table_name.toLowerCase(),"'  ",
                   "   AND EW.DZEW002 = '",ms_revision,"'                  ",
                   "   AND EW.DZEW003 = 'Index'                            ",
                   "   AND EW.DZEW010 = '",ms_alm_version,"'               ",
                   "   AND NVL(EW.DZEWSTUS,'Y') = 'Y'                      "  #20170912
      &endif             
      &endif             
    END IF
  END IF   
                
  PREPARE lpre_get_index_backward_records FROM ls_sql
  DECLARE lcur_get_index_backward_records CURSOR FOR lpre_get_index_backward_records

  OPEN lcur_get_index_backward_records
  FOREACH lcur_get_index_backward_records INTO lo_index_backward_records[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_index_backward_records
  
  FREE lpre_get_index_backward_records
  FREE lcur_get_index_backward_records  
  -------------------------------------------------------------------------------------------------
  
  LET lo_return.* = lo_index_backward_records.*
  
  RETURN lo_return
  
END FUNCTION 

FUNCTION sadzi140_rdif_fill_diff_list_by_column_id()
DEFINE 
  ls_sql   STRING,
  li_index INTEGER,
  li_count INTEGER
  
  IF ms_view_type = cs_view_type_db THEN
    &ifdef is_msv
    LET ls_sql = "with get_num as (                                                                                 ",  
                 "                  select 1 as numbers                                                             ",  
                 "                   union all                                                                      ",  
                 "                  select numbers + 1 as numbers                                                   ",  
                 "                    from get_num                                                                  ",  
                 "                   where numbers < 1000                                                           ",  
                 "                )                                                                                 ",  
                 "select ''              record_type,                                                               ",  
                 "       ro.row_num      row_num,                                                                   ",  
                 "       src.column_name src_col_name,                                                              ",  
                 "       dst.column_name dst_col_name                                                               ",  
                 "  from (                                                                                          ",  
                 "          select numbers as row_num                                                               ",  
                 "            from get_num                                                                          ",  
                 "           where numbers <= (                                                                     ",  
                 "                               select max(mx.max_value) max_value                                 ",  
                 "                                 from (                                                           ",  
                 "                                        select max(coalesce(ev.dzev011,0)) max_value              ",      
                 "                                          from dzev_t ev                                          ",  
                 "                                         where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'     ",  
                 "                                           and ev.dzev002 = '",ms_table_name.toLowerCase(),"'     ",  
                 "                                           and ev.dzev003 = '",ms_revision,"'                     ",  
                 "                                           and ev.dzev004 = 'TableDesigner'                       ",  
                 "                                           and ev.dzev018 = '",ms_alm_version,"'                  ",  
                 "                                        union all                                                 ",  
                 "                                        select max(atc.column_id) max_value                       ",  
                 "                                          from dba_tab_columns atc                                ",  
                 "                                         where atc.owner      = '",ms_owner_name.toLowerCase(),"' ",  
                 "                                           and atc.table_name = '",ms_table_name.toLowerCase(),"' ",  
                 "                                      ) mx                                                        ",  
                 "                             )                                                                    ",  
                 "       ) ro                                                                                       ",  
                 "       left outer join                                                                            ",  
                 "       (                                                                                          ",  
                 "         select atc.column_id   row_num,                                                          ",  
                 "                atc.column_name column_name                                                       ",  
                 "           from dba_tab_columns atc                                                               ",  
                 "          where atc.owner      = '",ms_owner_name.toLowerCase(),"'                                ",  
                 "            and atc.table_name = '",ms_table_name.toLowerCase(),"'                                ",  
                 "            and atc.column_name <> 'dummy'                                                        ",  
                 "       ) src on src.row_num = ro.row_num                                                          ",  
                 "       left outer join                                                                            ",  
                 "       (                                                                                          ",  
                 "         select ev.dzev011        row_num,                                                        ",  
                 "                upper(ev.dzev005) column_name                                                     ",  
                 "           from dzev_t ev                                                                         ",  
                 "         where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'                                     ",  
                 "           and ev.dzev002 = '",ms_table_name.toLowerCase(),"'                                     ",  
                 "           and ev.dzev003 = '",ms_revision,"'                                                     ",  
                 "           and ev.dzev004 = 'TableDesigner'                                                       ",  
                 "           and ev.dzev018 = '",ms_alm_version,"'                                                  ",  
                 "       ) dst on dst.row_num = ro.row_num                                                          ",                            
                 " where 1=1                                                                                        ",  
                 " order by ro.row_num                                                                              ",  
                 " option(MaxRecursion 0)                                                                           "  
    &else
    &ifdef is_pgs
    LET ls_sql = "with recursive get_num as (                                                                       ",  
                 "                  select 1 as numbers                                                             ",  
                 "                   union all                                                                      ",  
                 "                  select numbers + 1 as numbers                                                   ",  
                 "                    from get_num                                                                  ",  
                 "                   where numbers < 1000                                                           ",  
                 "                )                                                                                 ",  
                 "select ''              record_type,                                                               ",  
                 "       ro.row_num      row_num,                                                                   ",  
                 "       src.column_name src_col_name,                                                              ",  
                 "       dst.column_name dst_col_name                                                               ",  
                 "  from (                                                                                          ",  
                 "          select numbers as row_num                                                               ",  
                 "            from get_num                                                                          ",  
                 "           where numbers <= (                                                                     ",  
                 "                               select max(mx.max_value) max_value                                 ",  
                 "                                 from (                                                           ",  
                 "                                        select max(coalesce(ev.dzev011,0)) max_value              ",      
                 "                                          from dzev_t ev                                          ",  
                 "                                         where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'     ",  
                 "                                           and ev.dzev002 = '",ms_table_name.toLowerCase(),"'     ",  
                 "                                           and ev.dzev003 = '",ms_revision,"'                     ",  
                 "                                           and ev.dzev004 = 'TableDesigner'                       ",  
                 "                                           and ev.dzev018 = '",ms_alm_version,"'                  ",  
                 "                                        union all                                                 ",  
                 "                                        select max(atc.column_id) max_value                       ",  
                 "                                          from dba_tab_columns atc                                ",  
                 "                                         where atc.owner      = '",ms_owner_name.toLowerCase(),"' ",  
                 "                                           and atc.table_name = '",ms_table_name.toLowerCase(),"' ",  
                 "                                      ) mx                                                        ",  
                 "                             )                                                                    ",  
                 "       ) ro                                                                                       ",  
                 "       left outer join                                                                            ",  
                 "       (                                                                                          ",  
                 "         select atc.column_id   row_num,                                                          ",  
                 "                atc.column_name column_name                                                       ",  
                 "           from dba_tab_columns atc                                                               ",  
                 "          where atc.owner      = '",ms_owner_name.toLowerCase(),"'                                ",  
                 "            and atc.table_name = '",ms_table_name.toLowerCase(),"'                                ",  
                 "            and atc.column_name <> 'dummy'                                                        ",  
                 "       ) src on src.row_num = ro.row_num                                                          ",  
                 "       left outer join                                                                            ",  
                 "       (                                                                                          ",  
                 "         select ev.dzev011  row_num,                                                              ",  
                 "                ev.dzev005  column_name                                                           ",  
                 "           from dzev_t ev                                                                         ",  
                 "         where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'                                     ",  
                 "           and ev.dzev002 = '",ms_table_name.toLowerCase(),"'                                     ",  
                 "           and ev.dzev003 = '",ms_revision,"'                                                     ",  
                 "           and ev.dzev004 = 'TableDesigner'                                                       ",  
                 "           and ev.dzev018 = '",ms_alm_version,"'                                                  ",  
                 "       ) dst on dst.row_num = ro.row_num                                                          ",                            
                 " where 1=1                                                                                        ",  
                 " order by ro.row_num                                                                              "  
    &else
    LET ls_sql = "SELECT ''              RECORD_TYPE,                                                              ",
                 "       RO.ROW_NUM      ROW_NUM,                                                                  ",
                 "       SRC.COLUMN_NAME SRC_COL_NAME,                                                             ",
                 "       DST.COLUMN_NAME DST_COL_NAME                                                              ",
                 "  FROM (                                                                                         ",
                 "          SELECT ROWNUM ROW_NUM                                                                  ",
                 "            FROM DUAL                                                                            ",
                 "         CONNECT BY LEVEL <= (                                                                   ",
                 "                               SELECT MAX(MX.MAX_VALUE) MAX_VALUE                                ",
                 "                                 FROM (                                                          ",
                 "                                        SELECT MAX(NVL(EV.DZEV011,0)) MAX_VALUE                  ",
                 "                                          FROM DZEV_T EV                                         ",
                 "                                         WHERE EV.DZEV001 = '",ms_owner_name.toUpperCase(),"'    ",
                 "                                           AND EV.DZEV002 = '",ms_table_name.toUpperCase(),"'    ",
                 "                                           AND EV.DZEV003 = '",ms_revision,"'                    ",
                 "                                           AND EV.DZEV004 = 'TableDesigner'                      ",
                 "                                           AND EV.DZEV018 = '",ms_alm_version,"'                 ",
                 "                                        UNION ALL                                                ",
                 "                                        SELECT MAX(ATC.COLUMN_ID) MAX_VALUE                      ",
                 "                                          FROM DBA_TAB_COLUMNS ATC                               ",
                 "                                         WHERE ATC.OWNER      = '",ms_owner_name.toUpperCase(),"'",    
                 "                                           AND ATC.TABLE_NAME = '",ms_table_name.toUpperCase(),"'",
                 "                                      ) MX                                                       ",
                 "                             )                                                                   ",
                 "       ) RO,                                                                                     ",
                 "       (                                                                                         ",
                 "         SELECT ATC.COLUMN_ID   ROW_NUM,                                                         ",
                 "                ATC.COLUMN_NAME COLUMN_NAME                                                      ",
                 "           FROM DBA_TAB_COLUMNS ATC                                                              ",
                 "          WHERE ATC.OWNER      = '",ms_owner_name.toUpperCase(),"'                               ",   
                 "            AND ATC.TABLE_NAME = '",ms_table_name.toUpperCase(),"'                               ",
                 "            AND ATC.COLUMN_NAME <> 'DUMMY'                                                       ",
                 "       ) SRC,                                                                                    ",
                 "       (                                                                                         ",
                 "         SELECT EV.DZEV011        ROW_NUM,                                                       ",
                 "                UPPER(EV.DZEV005) COLUMN_NAME                                                    ",
                 "           FROM DZEV_T EV                                                                        ",
                 "         WHERE EV.DZEV001 = '",ms_owner_name.toUpperCase(),"'                                    ",
                 "           AND EV.DZEV002 = '",ms_table_name.toUpperCase(),"'                                    ",
                 "           AND EV.DZEV003 = '",ms_revision,"'                                                    ",
                 "           AND EV.DZEV004 = 'TableDesigner'                                                      ",
                 "           AND EV.DZEV018 = '",ms_alm_version,"'                                                 ",
                 "       ) DST                                                                                     ",
                 " WHERE SRC.ROW_NUM (+)= RO.ROW_NUM                                                               ",
                 "   AND DST.ROW_NUM (+)= RO.ROW_NUM                                                               ",
                 " ORDER BY RO.ROW_NUM                                                                             "
    &endif                 
    &endif                 
  ELSE 
    IF ms_view_type = cs_view_type_rt THEN
      &ifdef is_msv
      LET ls_sql = "with get_num as (                                                                             ",
                   "                  select 1 as numbers                                                         ",
                   "                   union all                                                                  ",
                   "                  select numbers + 1 as numbers                                               ",
                   "                    from get_num                                                              ",
                   "                   where numbers < 1000                                                       ",
                   "                )                                                                             ",
                   "select ''              record_type,                                                           ",   
                   "       ro.row_num      row_num,                                                               ",   
                   "       src.column_name src_col_name,                                                          ",   
                   "       dst.column_name dst_col_name                                                           ",   
                   "  from (                                                                                      ",   
                   "          select numbers as row_num                                                           ",
                   "            from get_num                                                                      ",
                   "           where numbers <= (                                                                 ",  
                   "                               select max(mx.max_value) max_value                             ",   
                   "                                 from (                                                       ",   
                   "                                        select max(coalesce(ev.dzev011,0)) max_value          ",        
                   "                                          from dzev_t ev                                      ",   
                   "                                         where ev.dzev001 = '",ms_owner_name.toLowerCase(),"' ",   
                   "                                           and ev.dzev002 = '",ms_table_name.toLowerCase(),"' ",   
                   "                                           and ev.dzev003 = '",ms_revision,"'                 ",   
                   "                                           and ev.dzev004 = 'TableDesigner'                   ",   
                   "                                           and ev.dzev018 = '",ms_alm_version,"'              ",   
                   "                                        union all                                             ",   
                   "                                        select max(coalesce(eb.dzeb021,0)) max_value          ",        
                   "                                          from dzeb_t eb                                      ",   
                   "                                        where eb.dzeb001 = '",ms_table_name.toLowerCase(),"'  ",   
                   "                                      ) mx                                                    ",   
                   "                             )                                                                ",   
                   "       ) ro                                                                                   ",
                   "       left outer join                                                                        ",
                   "       (                                                                                      ",   
                   "         select eb.dzeb021 row_num,                                                           ",
                   "                eb.dzeb002 column_name                                                        ",
                   "           from dzeb_t eb                                                                     ",   
                   "         where eb.dzeb001 = '",ms_table_name.toLowerCase(),"'                                 ",   
                   "       ) src on src.row_num = ro.row_num                                                      ",
                   "       left outer join                                                                        ",
                   "       (                                                                                      ",   
                   "         select ev.dzev011 row_num,                                                           ",
                   "                ev.dzev005 column_name                                                        ",
                   "           from dzev_t ev                                                                     ",    
                   "         where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'                                 ",    
                   "           and ev.dzev002 = '",ms_table_name.toLowerCase(),"'                                 ",    
                   "           and ev.dzev003 = '",ms_revision,"'                                                 ",    
                   "           and ev.dzev004 = 'TableDesigner'                                                   ",    
                   "           and ev.dzev018 = '",ms_alm_version,"'                                              ",    
                   "       ) dst on dst.row_num = ro.row_num                                                      ",          
                   " where 1=1                                                                                    ", 
                   " order by ro.row_num                                                                          ", 
                   " option(MaxRecursion 0)                                                                       " 
      &else
      &ifdef is_pgs
      LET ls_sql = "with recursive get_num as (                                                                   ",
                   "                  select 1 as numbers                                                         ",
                   "                   union all                                                                  ",
                   "                  select numbers + 1 as numbers                                               ",
                   "                    from get_num                                                              ",
                   "                   where numbers < 1000                                                       ",
                   "                )                                                                             ",
                   "select ''              record_type,                                                           ",   
                   "       ro.row_num      row_num,                                                               ",   
                   "       src.column_name src_col_name,                                                          ",   
                   "       dst.column_name dst_col_name                                                           ",   
                   "  from (                                                                                      ",   
                   "          select numbers as row_num                                                           ",
                   "            from get_num                                                                      ",
                   "           where numbers <= (                                                                 ",  
                   "                               select max(mx.max_value) max_value                             ",   
                   "                                 from (                                                       ",   
                   "                                        select max(coalesce(ev.dzev011,0)) max_value          ",        
                   "                                          from dzev_t ev                                      ",   
                   "                                         where ev.dzev001 = '",ms_owner_name.toLowerCase(),"' ",   
                   "                                           and ev.dzev002 = '",ms_table_name.toLowerCase(),"' ",   
                   "                                           and ev.dzev003 = '",ms_revision,"'                 ",   
                   "                                           and ev.dzev004 = 'TableDesigner'                   ",   
                   "                                           and ev.dzev018 = '",ms_alm_version,"'              ",   
                   "                                        union all                                             ",   
                   "                                        select max(coalesce(eb.dzeb021,0)) max_value          ",        
                   "                                          from dzeb_t eb                                      ",   
                   "                                        where eb.dzeb001 = '",ms_table_name.toLowerCase(),"'  ",   
                   "                                      ) mx                                                    ",   
                   "                             )                                                                ",   
                   "       ) ro                                                                                   ",
                   "       left outer join                                                                        ",
                   "       (                                                                                      ",   
                   "         select eb.dzeb021 row_num,                                                           ",
                   "                eb.dzeb002 column_name                                                        ",
                   "           from dzeb_t eb                                                                     ",   
                   "         where eb.dzeb001 = '",ms_table_name.toLowerCase(),"'                                 ",   
                   "       ) src on src.row_num = ro.row_num                                                      ",
                   "       left outer join                                                                        ",
                   "       (                                                                                      ",   
                   "         select ev.dzev011 row_num,                                                           ",
                   "                ev.dzev005 column_name                                                        ",
                   "           from dzev_t ev                                                                     ",    
                   "         where ev.dzev001 = '",ms_owner_name.toLowerCase(),"'                                 ",    
                   "           and ev.dzev002 = '",ms_table_name.toLowerCase(),"'                                 ",    
                   "           and ev.dzev003 = '",ms_revision,"'                                                 ",    
                   "           and ev.dzev004 = 'TableDesigner'                                                   ",    
                   "           and ev.dzev018 = '",ms_alm_version,"'                                              ",    
                   "       ) dst on dst.row_num = ro.row_num                                                      ",          
                   " where 1=1                                                                                    ", 
                   " order by ro.row_num                                                                          " 
      &else
      LET ls_sql = "SELECT ''              RECORD_TYPE,                                                              ",
                   "       RO.ROW_NUM      ROW_NUM,                                                                  ",
                   "       SRC.COLUMN_NAME SRC_COL_NAME,                                                             ",
                   "       DST.COLUMN_NAME DST_COL_NAME                                                              ",
                   "  FROM (                                                                                         ",
                   "          SELECT ROWNUM ROW_NUM                                                                  ",
                   "            FROM DUAL                                                                            ",
                   "         CONNECT BY LEVEL <= (                                                                   ",
                   "                               SELECT MAX(MX.MAX_VALUE) MAX_VALUE                                ",
                   "                                 FROM (                                                          ",
                   "                                        SELECT MAX(NVL(EV.DZEV011,0)) MAX_VALUE                  ",
                   "                                          FROM DZEV_T EV                                         ",
                   "                                         WHERE EV.DZEV001 = '",ms_owner_name.toUpperCase(),"'    ",
                   "                                           AND EV.DZEV002 = '",ms_table_name.toUpperCase(),"'    ",
                   "                                           AND EV.DZEV003 = '",ms_revision,"'                    ",
                   "                                           AND EV.DZEV004 = 'TableDesigner'                      ",
                   "                                           AND EV.DZEV018 = '",ms_alm_version,"'                 ",
                   "                                        UNION ALL                                                ",
                   "                                        SELECT MAX(NVL(EB.DZEB021,0)) MAX_VALUE                  ",
                   "                                          FROM DZEB_T EB                                         ",
                   "                                        WHERE EB.DZEB001 = '",ms_table_name.toLowerCase(),"'     ",
                   "                                      ) MX                                                       ",
                   "                             )                                                                   ",
                   "       ) RO,                                                                                     ",
                   "       (                                                                                         ",
                   "         SELECT EB.DZEB021        ROW_NUM,                                                       ",
                   "                UPPER(EB.DZEB002) COLUMN_NAME                                                    ",
                   "           FROM DZEB_T EB                                                                        ",
                   "         WHERE EB.DZEB001 = '",ms_table_name.toLowerCase(),"'                                    ",
                   "       ) SRC,                                                                                    ",
                   "       (                                                                                         ",
                   "         SELECT EV.DZEV011        ROW_NUM,                                                       ",
                   "                UPPER(EV.DZEV005) COLUMN_NAME                                                    ",
                   "           FROM DZEV_T EV                                                                        ",
                   "         WHERE EV.DZEV001 = '",ms_owner_name.toUpperCase(),"'                                    ",
                   "           AND EV.DZEV002 = '",ms_table_name.toUpperCase(),"'                                    ",
                   "           AND EV.DZEV003 = '",ms_revision,"'                                                    ",
                   "           AND EV.DZEV004 = 'TableDesigner'                                                      ",
                   "           AND EV.DZEV018 = '",ms_alm_version,"'                                                 ",
                   "       ) DST                                                                                     ",
                   " WHERE SRC.ROW_NUM (+)= RO.ROW_NUM                                                               ",
                   "   AND DST.ROW_NUM (+)= RO.ROW_NUM                                                               ",
                   " ORDER BY RO.ROW_NUM                                                                             "
      &endif                   
      &endif                   
    END IF
  END IF     

  PREPARE lpre_fill_diff_list_by_column_id FROM ls_sql
  DECLARE lcur_fill_diff_list_by_column_id CURSOR FOR lpre_fill_diff_list_by_column_id

  CALL m_diff_list_by_column_id.clear()
  CALL m_diff_list_by_column_id_color.clear()
  
  LET li_count = 1

  #DISPLAY ls_TableName
  FOREACH lcur_fill_diff_list_by_column_id INTO m_diff_list_by_column_id[li_count].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET m_diff_list_by_column_id[li_count].RECORD_TYPE = NULL

    LET m_diff_list_by_column_id_color[li_count].RECORD_TYPE    = NULL
    LET m_diff_list_by_column_id_color[li_count].COLUMN_ID      = NULL
    LET m_diff_list_by_column_id_color[li_count].DB_COLUMN_NAME = NULL
    LET m_diff_list_by_column_id_color[li_count].RT_COLUMN_NAME = NULL

    IF NVL(m_diff_list_by_column_id[li_count].DB_COLUMN_NAME,cs_null_value) <> NVL(m_diff_list_by_column_id[li_count].RT_COLUMN_NAME,cs_null_value) THEN
      LET m_diff_list_by_column_id[li_count].RECORD_TYPE = "!"
      LET m_diff_list_by_column_id_color[li_count].RECORD_TYPE    = "red"
      LET m_diff_list_by_column_id_color[li_count].DB_COLUMN_NAME = "red reverse"
      LET m_diff_list_by_column_id_color[li_count].RT_COLUMN_NAME = "yellow reverse"
    END IF
    
    LET li_count = li_count + 1
    
  END FOREACH

END FUNCTION