{
  異動型態    異動單號       日 期       異動者      異動內容
  ========= ============= ========== ========== ===============================================
  Modify                  20170912   Ernestliou 1.索引匯入時判斷有無已存在 index 欄位, 有則匯入方設失效
  Modify                  20180903   Ernestliou 1.修正比對Constratint時惕除系統Constraint的問題
  Modify                  20200330   02160      1.修改支援 PosgreSQL(1.1.072)   
  Modify                  20230807   02160      1.修改支援 DM
  Modify                  20230901   02160      1.修改支援 MogDB
}  
&include "../4gl/sadzi140_mcr.inc"

IMPORT os
IMPORT util

SCHEMA ds

&include "../4gl/sadzp000_cnst.inc"
&include "../4gl/sadzi140_type.inc"
&include "../4gl/sadzi140_cnst.inc"

#新增表格設計器表格資料的快照
FUNCTION sadzi140_shot_set_table_designer_snapshot(p_owner,p_table_name,p_version,p_curr_version,p_alm_version,p_alm_request_no,p_dgenv)
DEFINE
  p_owner          STRING,
  p_table_name     STRING,
  p_version        STRING,
  p_curr_version   STRING, 
  p_alm_version    STRING,
  p_alm_request_no STRING,
  p_dgenv          STRING  
DEFINE
  ls_owner        STRING,
  ls_table_name   STRING,
  ls_version      STRING,
  ls_curr_version STRING, 
  ls_alm_version  STRING, 
  ls_alm_request_no STRING, 
  ls_dgenv          STRING,  
  ls_sql          STRING,
  li_count        INTEGER,
  lo_snapshot     T_SNAPSHOT

  LET ls_owner          = p_owner
  LET ls_table_name     = p_table_name
  LET ls_version        = p_version
  LET ls_curr_version   = p_curr_version
  LET ls_alm_version    = p_alm_version
  LET ls_alm_request_no = p_alm_request_no
  LET ls_dgenv          = p_dgenv
  
  LET li_count = 1

  BEGIN WORK
  
  &ifdef is_msv
  LET ls_sql = "select '",ls_owner.toLowerCase(),"'           owner,           ",
               "       '",ls_table_name.toLowerCase(),"'      table_name,      ",
               "       ''                                     version,         ",
               "       ''                                     code_type,       ",
               "       eb.dzeb002                             column_name,     ",
               "       eb.dzeb004                             iskey,           ",
               "       iif(eb.dzeb004='Y','N','Y')            nullable,        ",
               "       td.gztd003                             data_type,       ",
               "       case td.gztd003                                         ",
               "         when 'varbinary'                                      ",
               "           then 'max'                                          ",
               "         when 'nvarchar'                                       ",
               "           then iif(td.gztd009='TEXT','max',td.gztd008)        ",
               "       else                                                    ",
               "         td.gztd008                                            ",
               "       end                                    data_length,     ",
               "       eb.dzeb003                             comments,        ",
               "       eb.dzeb021                             column_seq,      ",
               "       ''                                     prev_version,    ",
               "       eb.dzeb006                             column_type,     ",
               "       eb.dzeb021                             column_seq,      ",
               "       eb.dzeb022                             column_type_def, ",
               "       eb.dzeb023                             number_type_seq, ",
               "       eb.dzeb024                             column_memo,     ",
               "       ''                                     alm_version,     ",
               "       ''                                     alm_request_no,  ",
               "       eb.dzeb029                             dgenv,           ",
               "       eb.dzeb012                             defaultv,        ",
               "       eb.dzeb030                             orig_dgenv,      ",
               "       eb.dzeb031                             shipping_code    ",
               "  from dzeb_t eb                                               ",
               "       left outer join gztd_t td                               ",
               "                    on td.gztd001 = eb.dzeb006                 ",
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'          ",
               " order by eb.dzeb021                                           " 
  &else
  &ifdef is_pgs
  LET ls_sql = "select '",ls_owner.toLowerCase(),"'        as owner,           ",
               "       '",ls_table_name.toLowerCase(),"'      table_name,      ",
               "       ''                                  as version,         ",
               "       ''                                     code_type,       ",
               "       eb.dzeb002                             column_name,     ",
               "       eb.dzeb004                             iskey,           ",
               "       case when eb.dzeb004='Y'                                ",
               "       then 'N' else 'Y' end                  nullable,        ",
               "       td.gztd003                             data_type,       ",
               "       td.gztd008                             data_length,     ",
               "       eb.dzeb003                          as comments,        ",
               "       eb.dzeb021                             column_seq,      ",
               "       ''                                     prev_version,    ",
               "       eb.dzeb006                             column_type,     ",
               "       eb.dzeb021                             column_seq,      ",
               "       eb.dzeb022                             column_type_def, ",
               "       eb.dzeb023                             number_type_seq, ",
               "       eb.dzeb024                             column_memo,     ",
               "       ''                                     alm_version,     ",
               "       ''                                     alm_request_no,  ",
               "       eb.dzeb029                             dgenv,           ",
               "       eb.dzeb012                             defaultv,        ",
               "       eb.dzeb030                             orig_dgenv,      ",
               "       eb.dzeb031                             shipping_code    ",
               "  from dzeb_t eb                                               ",
               "       left outer join gztd_t td                               ",
               "                    on td.gztd001 = eb.dzeb006                 ",
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'          ",
               " order by eb.dzeb021                                           " 
  &else
  LET ls_sql = "SELECT '",ls_owner.toUpperCase(),"'           AS OWNER,           ",
               "       '",ls_table_name.toUpperCase(),"'      AS TABLE_NAME,      ",
               "       ''                                     AS VERSION,         ",
               "       ''                                     AS CODE_TYPE,       ",               
               "       UPPER(EB.DZEB002)                      AS COLUMN_NAME,     ",
               "       EB.DZEB004                             AS ISKEY,           ",
               "       DECODE(EB.DZEB004,'Y','N','N','Y','Y') AS NULLABLE,        ",
               "       UPPER(TD.GZTD003)                      AS DATA_TYPE,       ",
               "       REPLACE(TO_CHAR(TO_NUMBER(REPLACE(REPLACE(                 ", 
               "       DECODE(                                                    ",
               "              UPPER(TD.GZTD003),                                  ",
               "              'BLOB','4000',                                      ",
               "              'CLOB','4000',                                      ",
               "              'DATE','7',                                         ",
               "              'TIMESTAMP',DECODE(TD.GZTD008,'0','7','11'),        ",
               "              TD.GZTD008                                          ",
               "             ),',','.'),' ',''))),'.',',')    AS DATA_LENGTH,     ",
               "       EB.DZEB003                             AS COMMENTS,        ",      
               "       EB.DZEB021                             AS COLUMN_SEQ,      ",
               "       ''                                     AS PREV_VERSION,    ",
               "       EB.DZEB006                             AS COLUMN_TYPE,     ",
               "       EB.DZEB021                             AS COLUMN_SEQ,      ",
               "       EB.DZEB022                             AS COLUMN_TYPE_DEF, ",
               "       EB.DZEB023                             AS NUMBER_TYPE_SEQ, ",
               "       EB.DZEB024                             AS COLUMN_MEMO,     ",
               "       ''                                     AS ALM_VERSION,     ",
               "       ''                                     AS ALM_REQUEST_NO,  ",
               "       EB.DZEB029                             AS DGENV,           ",
               "       EB.DZEB012                             AS DEFAULTV,        ", 
               "       EB.DZEB030                             AS ORIG_DGENV,      ",
               "       EB.DZEB031                             AS SHIPPING_CODE    ", 
               "  FROM DZEB_T EB                                                  ",   
               "       LEFT OUTER JOIN GZTD_T TD ON TD.GZTD001 = EB.DZEB006       ",   
               " WHERE EB.DZEB001 = '",ls_table_name.toLowerCase(),"'             ",
               " ORDER BY EB.DZEB021                                              "
  &endif               
  &endif               

  PREPARE lpre_GetTableDesignerSnapshot FROM ls_sql
  DECLARE lcur_GetTableDesignerSnapshot CURSOR FOR lpre_GetTableDesignerSnapshot
  
  FOREACH lcur_GetTableDesignerSnapshot INTO lo_snapshot.*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    --LET lo_snapshot.DZEV001 = ls_owner
    --LET lo_snapshot.DZEV002 = ls_table_name
    LET lo_snapshot.DZEV003 = ls_version
    LET lo_snapshot.DZEV004 = cs_table_designer
    LET lo_snapshot.DZEV012 = ls_curr_version
    LET lo_snapshot.DZEV018 = ls_alm_version
    LET lo_snapshot.DZEV019 = ls_alm_request_no
    
    CALL sadzi140_shot_insert_DZEV_T(lo_snapshot.*)

    LET li_count = li_count + 1
    
  END FOREACH

  COMMIT WORK
  
END FUNCTION

#新增現行表格的快照
FUNCTION sadzi140_shot_set_database_snapshot(p_owner,p_table_name,p_version,p_curr_version,p_alm_version,p_alm_request_no,p_dgenv)
DEFINE
  p_owner          STRING, 
  p_table_name     STRING,
  p_version        STRING,
  p_curr_version   STRING, 
  p_alm_version    STRING,
  p_alm_request_no STRING,
  p_dgenv          STRING  
DEFINE
  lo_snapshot       T_SNAPSHOT,
  ls_owner          STRING, 
  ls_table_name     STRING,
  ls_version        STRING,
  ls_curr_version   STRING, 
  ls_alm_version    STRING,
  ls_alm_request_no STRING, 
  ls_dgenv          STRING,  
  ls_sql            STRING,
  li_count          INTEGER

  LET ls_owner          = p_owner
  LET ls_table_name     = p_table_name
  LET ls_version        = p_version
  LET ls_curr_version   = p_curr_version
  LET ls_alm_version    = p_alm_version
  LET ls_alm_request_no = p_alm_request_no
  LET ls_dgenv          = p_dgenv
  
  LET li_count = 1

  ##############################################################################
  BEGIN WORK
  &ifdef is_msv
  LET ls_sql = "select atc.owner                            owner,           ",                       
               "       atc.table_name                       table_name,      ",                       
               "       ''                                   version,         ",                       
               "       ''                                   code_type,       ",                       
               "       atc.column_name                      column_name,     ",                       
               "       iif(atc.nullable='Y','N','Y')        iskey,           ",                       
               "       atc.nullable                         nullable,        ",                       
               "       atc.data_type                        data_type,       ",                       
               "       atc.data_length                      data_length,     ",                       
               "       atc.comment                          comments,        ",                       
               "       ''                                   column_seq,      ",                       
               "       ''                                   prev_version,    ",                       
               "       ''                                   column_type,     ",                       
               "       ''                                   column_seq,      ",                       
               "       ''                                   column_type_def, ",                       
               "       ''                                   number_type_seq, ",                       
               "       ''                                   column_memo,     ",                       
               "       ''                                   alm_version,     ",                       
               "       ''                                   alm_request_no,  ",                       
               "       ''                                   dgenv,           ",                       
               "       ''                                   defaultv,        ",                       
               "       ''                                   org_dgenv,       ",                       
               "       ''                                   shipping_notice  ",                       
               "  from dba_tab_columns atc                                   ",                       
               "  left outer join dba_cons_columns acc                       ",       
               "               on acc.owner       = atc.owner                ",       
               "              and acc.table_name  = atc.table_name           ",       
               "              and acc.column_name = atc.column_name          ",       
               " where 1=1                                                   ",                       
               "   and atc.owner      = '",ls_owner.toLowerCase(),"'         ",                       
               "   and atc.table_name = '",ls_table_name.toLowerCase(),"'    ",                       
               " order by atc.column_name                                    "                       
  &else
  &ifdef is_pgs
  LET ls_sql = "select atc.owner                         as owner,                            ",                       
               "       atc.table_name                    as table_name,                       ",                       
               "       ''                                as version,                          ",                       
               "       ''                                as code_type,                        ",                       
               "       atc.column_name                   as column_name,                      ",                       
               "       case when atc.nullable='Y'                                             ",
               "       then 'N' else 'Y' end             as iskey,                            ",                       
               "       atc.nullable                      as nullable,                         ",                       
               "       atc.data_type                     as data_type,                        ",                       
               "       case                                                                   ",
               "         when atc.data_type = 'text' then null                                ",
               "         when atc.data_type = 'bytea' then null                               ",
               "         when atc.data_type = 'date' then null                                ",
               "         when atc.data_type = 'int2' then null                                ",
               "         when atc.data_type = 'int4' then null                                ",
               "         when atc.data_type = 'int8' then null                                ",
               "         when atc.data_type = 'decimal' then                                  ",
               "           atc.data_precision::varchar||','||                                 ",
               "           atc.data_scale::varchar                                            ",
               "       else                                                                   ",
               "         atc.data_length::varchar                                             ",
               "       end::varchar                      as data_length,                      ",
               "       act.comments                      as comments,                         ",                       
               "       ''                                as column_seq,                       ",                       
               "       ''                                as prev_version,                     ",                       
               "       ''                                as column_type,                      ",                       
               "       ''                                as column_seq,                       ",                       
               "       ''                                as column_type_def,                  ",                       
               "       ''                                as number_type_seq,                  ",                       
               "       ''                                as column_memo,                      ",                       
               "       ''                                as alm_version,                      ",                       
               "       ''                                as alm_request_no,                   ",                       
               "       ''                                as dgenv,                            ",                       
               "       ''                                as defaultv,                         ",                       
               "       ''                                as org_dgenv,                        ",                       
               "       ''                                as shipping_notice                   ",                       
               "  from dba_tab_columns atc                                                    ",                       
               "  left outer join (                                                           ",
               "                    select * from dba_cons_columns dcc                        ",
               "                     where dcc.table_name = '",ls_table_name.toLowerCase(),"' ",
               "                  ) acc                                                       ",       
               "               on acc.owner       = atc.owner                                 ",       
               "              and acc.table_name  = atc.table_name                            ",       
               "              and acc.column_name = atc.column_name                           ",       
               "  left outer join (                                                           ",
               "                    select * from dba_col_comments dcc                        ",
               "                     where dcc.table_name = '",ls_table_name.toLowerCase(),"' ",
               "                  ) act                                                       ",
               "               on act.owner       = atc.owner                                 ",        
               "              and act.table_name  = atc.table_name                            ",
               "              and act.column_name = atc.column_name                           ",
               " where 1=1                                                                    ",                       
               "   and atc.owner      = '",ls_owner.toLowerCase(),"'                          ",                       
               "   and atc.table_name = '",ls_table_name.toLowerCase(),"'                     ",                       
               " order by atc.column_name                                                     "                       
  &else
  LET ls_sql = "SELECT ATC.OWNER                            AS OWNER,                      ",
               "       ATC.TABLE_NAME                       AS TABLE_NAME,                 ",
               "       ''                                   AS VERSION,                    ",
               "       ''                                   AS CODE_TYPE,                  ",               
               "       ATC.COLUMN_NAME                      AS COLUMN_NAME,                ",
               "       DECODE(ACC.COLUMN_NAME,NULL,'N','Y') AS ISKEY,                      ",
               "       ATC.NULLABLE                         AS NULLABLE,                   ",
               "       ATC.DATA_TYPE                        AS DATA_TYPE,                  ",
               "       REPLACE(TO_CHAR(DECODE(                                             ",
               "               ATC.DATA_TYPE,                                              ",
               "               'BLOB',4000,                                                ", #20240529
               "               'CLOB',4000,                                                ", #20240529
               "               'INT2',NULL,                                                ", #20240529
               "               'INT4',NULL,                                                ", #20240529
               "               'INT8',NULL,                                                ", #20240529
               "               'NUMBER',DECODE(                                            ",
               "                                NVL(ATC.DATA_SCALE,'0'),                   ",
               "                                '0',ATC.DATA_PRECISION,                    ",
               "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE    ",
               "                              ),                                           ",
               "               'NUMERIC',DECODE(                                           ", #20240529 begin
               "                                NVL(ATC.DATA_SCALE,'0'),                   ",
               "                                '0',ATC.DATA_PRECISION,                    ",
               "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE    ", 
               "                              ),                                           ", #20240529 end
               "               'DATE',7,                                                   ", #20240529
               "               'TIMESTAMP(0)',7,                                           ", #20240529
               "               'TIMESTAMP(5)',11,                                          ", #20240529
               "               'TIMESTAMP(6)',11,                                          ", #20240529
               "               'TIMESTAMP',DECODE(                                         ",
               "                                   ATC.DATA_SCALE,0,7,8,7,11               ", #20240529
               "                                 ),                                        ",                
               "               'DATE',DECODE(                                              ",
               "                              ATC.DATA_LENGTH,3,7,ATC.DATA_LENGTH          ",
               "                            ),                                             ",                
               "               ATC.DATA_LENGTH                                             ",
               "             )),'.',',')                    AS DATA_LENGTH,                ",
               "       ACT.COMMENTS                         AS COMMENTS,                   ",   
               "       ''                                   AS COLUMN_SEQ,                 ",
               "       ''                                   AS PREV_VERSION,               ",
               "       ''                                   AS COLUMN_TYPE,                ",
               "       ''                                   AS COLUMN_SEQ,                 ",
               "       ''                                   AS COLUMN_TYPE_DEF,            ",
               "       ''                                   AS NUMBER_TYPE_SEQ,            ",
               "       ''                                   AS COLUMN_MEMO,                ",
               "       ''                                   AS ALM_VERSION,                ",
               "       ''                                   AS ALM_REQUEST_NO,             ",
               "       ''                                   AS DGENV,                      ",
               "       ''                                   AS DEFAULTV,                   ",  
               "       ''                                   AS ORG_DGENV,                  ",
               "       ''                                   AS SHIPPING_NOTICE             ",
               "  FROM DBA_TAB_COLUMNS ATC                                                 ",
               "  LEFT OUTER JOIN DBA_CONS_COLUMNS ACC                                     ", 
               "               ON ACC.OWNER = ATC.OWNER                                    ",
               "              AND ACC.TABLE_NAME  = ATC.TABLE_NAME                         ",
               "              AND ACC.COLUMN_NAME = ATC.COLUMN_NAME                        ",
               "  LEFT OUTER JOIN DBA_COL_COMMENTS ACT                                     ",
               "               ON ACT.OWNER       = ATC.OWNER                              ",        
               "              AND ACT.TABLE_NAME  = ATC.TABLE_NAME                         ",
               "              AND ACT.COLUMN_NAME = ATC.COLUMN_NAME                        ",
               " WHERE 1=1                                                                 ",
               "   AND ATC.OWNER      = '",ls_owner.toUpperCase(),"'                       ",
               "   AND ATC.TABLE_NAME = '",ls_table_name.toUpperCase(),"'                  ",
               "   AND ATC.COLUMN_NAME NOT LIKE 'SYS_%'                                    ",
               " ORDER BY ATC.COLUMN_NAME                                                  "
  &endif               
  &endif               

  PREPARE lpre_get_database_snapshot FROM ls_sql
  DECLARE lcur_get_database_snapshot CURSOR FOR lpre_get_database_snapshot
  
  FOREACH lcur_get_database_snapshot INTO lo_snapshot.*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    --LET lo_snapshot.DZEV001 = ls_owner
    --LET lo_snapshot.DZEV002 = ls_table_name
    LET lo_snapshot.DZEV003 = ls_version
    LET lo_snapshot.DZEV004 = cs_database
    LET lo_snapshot.DZEV012 = ls_curr_version
    LET lo_snapshot.DZEV018 = ls_alm_version
    LET lo_snapshot.DZEV019 = ls_alm_request_no
    LET lo_snapshot.DZEV020 = ls_dgenv
    LET lo_snapshot.DZEV022 = ls_dgenv
    LET lo_snapshot.DZEV023 = "N"

    CALL sadzi140_shot_insert_DZEV_T(lo_snapshot.*)
    
    LET li_count = li_count + 1
    
  END FOREACH

  COMMIT WORK
  ##############################################################################

  ## 更新序號, 有就更新, 沒有留空白 #################################################
  BEGIN WORK
  
  LET ls_sql = "update dzev_t                                                   ",
               "   set dzev011 = (                                              ",
               "                   select ev2.dzev011                           ",
               "                     from dzev_t ev2                            ",
               "                    where ev2.dzev001 = dzev001                 ",
               "                      and ev2.dzev002 = dzev002                 ",
               "                      and ev2.dzev003 = dzev003                 ",
               "                      and ev2.dzev005 = dzev005                 ",
               "                      and ev2.dzev004 = '",cs_table_designer,"' ",
               "                 )                                              ",
               &ifdef is_msv
               " where dzev001 = '",ls_owner.toLowerCase(),"'                   ",
               "   and dzev002 = '",ls_table_name.toLowerCase(),"'              ",
               &else
               &ifdef is_pgs
               " where dzev001 = '",ls_owner.toLowerCase(),"'                   ",
               "   and dzev002 = '",ls_table_name.toLowerCase(),"'              ",
               &else
               " where dzev001 = '",ls_owner.toUpperCase(),"'                   ",
               "   and dzev002 = '",ls_table_name.toUpperCase(),"'              ",
               &endif               
               &endif               
               "   and dzev003 = '",ls_version,"'                               ",
               "   and dzev004 = '",cs_database,"'                              "

  TRY
    PREPARE lpre_update_seq_no FROM ls_sql
    EXECUTE lpre_update_seq_no
  CATCH
  END TRY
  
  COMMIT WORK
  ##############################################################################
  
END FUNCTION

FUNCTION sadzi140_shot_insert_DZEV_T(p_snapshot)
DEFINE
  p_snapshot  T_SNAPSHOT
DEFINE
  lo_snapshot  T_SNAPSHOT

  LET lo_snapshot.* = p_snapshot.*

  TRY
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    insert into dzev_t(
                        dzev001,dzev002,dzev003,dzev004,dzev005,
                        dzev006,dzev007,dzev008,dzev009,dzev010,
                        dzev011,dzev012,dzev013,dzev014,dzev015,
                        dzev016,dzev017,dzev018,dzev019,dzev020,
                        dzev021,dzev022,dzev023
                      ) 
               values (
                        lo_snapshot.DZEV001,lo_snapshot.DZEV002,lo_snapshot.DZEV003,lo_snapshot.DZEV004,lo_snapshot.DZEV005,
                        lo_snapshot.DZEV006,lo_snapshot.DZEV007,lo_snapshot.DZEV008,lo_snapshot.DZEV009,lo_snapshot.DZEV010,
                        lo_snapshot.DZEV011,lo_snapshot.DZEV012,lo_snapshot.DZEV013,lo_snapshot.DZEV014,lo_snapshot.DZEV015,
                        lo_snapshot.DZEV016,lo_snapshot.DZEV017,lo_snapshot.DZEV018,lo_snapshot.DZEV019,lo_snapshot.DZEV020,
                        lo_snapshot.DZEV021,lo_snapshot.DZEV022,lo_snapshot.DZEV023
                      )
  CATCH 
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif
    DISPLAY cs_warning_tag,"Insert snapshot into DZEV_T fault."
    #ROLLBACK WORK
    #BEGIN WORK
  END TRY  
  
END FUNCTION

FUNCTION sadzi140_shot_create_snapshot(p_master_user,p_table_name,p_curr_version,p_new_version,p_alm_version,p_alm_request_no,p_dgenv)
DEFINE
  p_master_user    STRING,
  p_table_name     STRING,
  p_curr_version   STRING,
  p_new_version    STRING,
  p_alm_version    STRING,
  p_alm_request_no STRING,
  p_dgenv          STRING 
DEFINE
  ls_master_user      STRING,
  ls_table_name       STRING,
  ls_curr_version     STRING,
  ls_new_version      STRING,
  ls_alm_version      STRING,
  ls_alm_request_no   STRING,
  ls_dgenv            STRING,
  lo_constraint_index DYNAMIC ARRAY OF T_CONSTRAINT_INDEX

  LET ls_master_user    = p_master_user
  LET ls_table_name     = p_table_name
  LET ls_curr_version   = p_curr_version
  LET ls_new_version    = p_new_version
  LET ls_alm_version    = p_alm_version
  LET ls_alm_request_no = p_alm_request_no
  LET ls_dgenv          = p_dgenv
  
  CALL sadzi140_shot_set_table_designer_snapshot(ls_master_user,ls_table_name,ls_new_version,ls_curr_version,ls_alm_version,ls_alm_request_no,ls_dgenv)
  CALL sadzi140_shot_set_database_snapshot(ls_master_user,ls_table_name,ls_new_version,ls_curr_version,ls_alm_version,ls_alm_request_no,ls_dgenv)
  CALL sadzi140_shot_get_constraint_index_data(ls_table_name,ls_new_version,ls_curr_version,ls_alm_version,ls_alm_request_no,ls_dgenv) RETURNING lo_constraint_index
  IF lo_constraint_index.getLength() >= 1 THEN  
    CALL sadzi140_shot_set_constraint_index_data(lo_constraint_index)
  END IF  
  
END FUNCTION

FUNCTION sadzi140_shot_check_diff(p_owner,p_table_name,p_drop_abnormal_column)
DEFINE
  p_owner      STRING,
  p_table_name STRING,
  p_drop_abnormal_column BOOLEAN
DEFINE
  ls_owner            STRING,
  ls_table_name       STRING,
  lb_drop_abnormal_column BOOLEAN,
  lo_column_diff      DYNAMIC ARRAY OF T_COLUMN_DIFF,
  lo_index_diff       DYNAMIC ARRAY OF T_INDEX_DIFF,
  lo_constraint_diff  DYNAMIC ARRAY OF T_CONSTRAINT_DIFF,
  lb_different        BOOLEAN,
  lb_alter            BOOLEAN
DEFINE
  lb_return        BOOLEAN
  
  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name
  LET lb_drop_abnormal_column = p_drop_abnormal_column

  LET lb_different = FALSE
  LET lb_alter     = FALSE

  #與 r.t 的資料比對看有無差異(比對表格欄位)
  CALL sadzi140_shot_check_column_diff(ls_owner,ls_table_name,lb_drop_abnormal_column) RETURNING lb_different,lo_column_diff
  IF lb_different THEN 
    LET lb_alter = TRUE 
    #DISPLAY "比對表格欄位" 
    #DISPLAY "curr_version",ls_curr_version 
  END IF

  #與 r.t 的資料比對看有無差異(比對Constraint)
  CALL sadzi140_shot_check_constraint_diff(ls_owner,ls_table_name,lb_drop_abnormal_column) RETURNING lb_different,lo_constraint_diff
  IF lb_different THEN 
    LET lb_alter = TRUE 
    #DISPLAY "比對 Constraint diff" 
    #DISPLAY "curr_version",ls_curr_version 
  END IF
  
  #與 r.t 的資料比對看有無差異(比對Index)
  CALL sadzi140_shot_check_index_diff(ls_owner,ls_table_name,lb_drop_abnormal_column) RETURNING lb_different,lo_index_diff
  IF lb_different THEN 
    LET lb_alter = TRUE 
    #DISPLAY "比對 Index diff" 
    #DISPLAY "curr_version",ls_curr_version 
  END IF

  LET lb_return = lb_alter
  
  RETURN lb_return
  
END FUNCTION

#比對現行 Table Editor 的資料和實際表格的差異(正反向都要抓)
FUNCTION sadzi140_shot_check_column_diff(p_owner,p_table_name,p_drop_abnormal_column)
DEFINE
  p_owner       STRING, 
  p_table_name  STRING,
  p_drop_abnormal_column BOOLEAN
DEFINE
  ls_owner        STRING, 
  ls_tableName    STRING,
  lb_drop_abnormal_column BOOLEAN,
  lo_column_diff  DYNAMIC ARRAY OF T_COLUMN_DIFF,
  li_rec_cnt      INTEGER,
  ls_sql          STRING,
  ls_db_type      STRING, #20240606
  ls_ora_coldef   STRING, #20240529
  ls_dummy_sql    STRING
DEFINE
  lb_different  BOOLEAN,
  lo_different  DYNAMIC ARRAY OF T_COLUMN_DIFF 

  LET ls_owner     = p_owner
  LET ls_tableName = p_table_name
  LET lb_drop_abnormal_column = p_drop_abnormal_column

  LET lb_different = FALSE
  LET lo_different = NULL
  LET li_rec_cnt = 1 

  IF lb_drop_abnormal_column THEN
    LET ls_dummy_sql = ""
  ELSE
    &ifdef is_msv
    LET ls_dummy_sql = " and atc.column_name = 'dummy' "
    &else
    &ifdef is_pgs
    LET ls_dummy_sql = " and atc.column_name = 'dummy' "
    &else
    LET ls_dummy_sql = " AND ATC.COLUMN_NAME = 'DUMMY' "
    &endif
    &endif
  END IF  
  
  ##############################################################################
  #正向比對
  &ifdef is_msv
  LET ls_sql = "select eb.dzeb002                        column_name,      ",                            
               "       eb.dzeb004                        iskey,            ",                            
               "       case eb.dzeb004                                     ",    
               "         when 'Y' then 'N'                                 ",    
               "         when 'N' then 'Y'                                 ",    
               "       else                                                ",    
               "         'Y'                                               ",    
               "       end                               nullable,         ",                            
               "       td.gztd003                        data_type,        ",                            
               "       case td.gztd003                                     ",    
               "         when 'varbinary'                                  ",    
               "           then 'max'                                      ",    
               "         when 'nvarchar'                                   ",    
               "           then iif(td.gztd009='TEXT','max',td.gztd008)    ",    
               "       else                                                ",    
               "         td.gztd008                                        ",    
               "       end                               data_length,      ",    
               "       trim(eb.dzeb012)                  data_default      ",                            
               "  from dzeb_t eb                                           ",                                 
               "       left outer join gztd_t td                           ",    
               "                    on td.gztd001 = eb.dzeb006             ",                    
               " where eb.dzeb001 = '",ls_tableName.toLowerCase(),"'       ",                                 
               "except                                                     ",                   
               "select atc.column_name                      column_name,   ",                                 
               "       iif(atc.nullable='Y','N','Y')        iskey,         ",                                 
               "       atc.nullable                         nullable,      ",                                 
               "       atc.data_type                        data_type,     ",    
               "       atc.data_length                      data_length,   ",                                 
               "       atc.data_default                     data_default   ",            
               "  from dba_tab_columns atc                                 ",                                 
               "  left outer join dba_cons_columns acc                     ",    
               "               on acc.owner       = atc.owner              ",         
               "              and acc.table_name  = atc.table_name         ",         
               "              and acc.column_name = atc.column_name        ",         
               "              and acc.constraint_name like '%PK'           ",         
               " where 1=1                                                 ",                                 
               "   and atc.owner      = '",ls_owner.toLowerCase(),"'       ",                                 
               "   and atc.table_name = '",ls_tableName.toLowerCase(),"'   ",                                 
               " order by 1                                                "                                 
  &else
  &ifdef is_pgs
  LET ls_sql = "select eb.dzeb002                        column_name,                       ",                            
               "       eb.dzeb004                        iskey,                             ",                            
               "       case eb.dzeb004                                                      ",    
               "         when 'Y' then 'N'                                                  ",    
               "         when 'N' then 'Y'                                                  ",    
               "       else                                                                 ",    
               "         'Y'                                                                ",    
               "       end                               nullable,                          ",                            
               "       td.gztd003                        data_type,                         ",                            
               "       td.gztd008                        data_length,                       ",    
               "       trim(eb.dzeb012)                  data_default                       ",                            
               "  from dzeb_t eb                                                            ",                                 
               "       left outer join gztd_t td                                            ",    
               "                    on td.gztd001 = eb.dzeb006                              ",                    
               " where eb.dzeb001 = '",ls_tableName.toLowerCase(),"'                        ",                                 
               "except                                                                      ",                   
               "select atc.column_name                      column_name,                    ",                                 
               "       case when atc.nullable='Y'                                           ", 
               "       then 'N' else 'Y' end                iskey,                          ",
               "       atc.nullable                         nullable,                       ",                                 
               "       atc.data_type                        data_type,                      ",    
               "       case                                                                 ",
               "         when atc.data_type = 'text' then null                              ",
               "         when atc.data_type = 'bytea' then null                             ",
               "         when atc.data_type = 'date' then null                              ",
               "         when atc.data_type = 'int2' then null                              ",
               "         when atc.data_type = 'int4' then null                              ",
               "         when atc.data_type = 'int8' then null                              ",
               "         when atc.data_type = 'decimal' then                                ",
               "           atc.data_precision::varchar||','||                               ",
               "           atc.data_scale::varchar                                          ",
               "       else                                                                 ",
               "         atc.data_length::varchar                                           ",
               "       end::varchar                         data_length,                    ",
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
               "              and acc.constraint_name like '%PK'                            ",         
               " where 1=1                                                                  ",                                 
               "   and atc.owner      = '",ls_owner.toLowerCase(),"'                        ",                                 
               "   and atc.table_name = '",ls_tableName.toLowerCase(),"'                    ",                                 
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
    LET ls_ora_coldef = " DECODE(TRIM(DS.GET_COL_DEFAULT(ATC.TABLE_NAME,ATC.COLUMN_NAME)),'',NULL,       ",
                        "        TRIM(DS.GET_COL_DEFAULT(ATC.TABLE_NAME,ATC.COLUMN_NAME))) DATA_DEFAULT  "
  END CASE
  
  LET ls_sql = "SELECT UPPER(EB.DZEB002)                      AS COLUMN_NAME,                              ",
               "       EB.DZEB004                             AS ISKEY,                                    ",
               "       DECODE(EB.DZEB004,'Y','N','N','Y','Y') AS NULLABLE,                                 ",
               "       UPPER(TD.GZTD003)                      AS DATA_TYPE,                                ",
               "       REPLACE(TO_CHAR(TO_NUMBER(REPLACE(REPLACE(                                          ", 
               "       DECODE(                                                                             ",
               "              UPPER(TD.GZTD003),                                                           ",
               "              'BLOB','4000',                                                               ",
               "              'CLOB','4000',                                                               ",
               "              'DATE','7',                                                                  ",
               "              'TIMESTAMP',DECODE(TD.GZTD008,'0','7','11'),                                 ",
               "              TD.GZTD008                                                                   ",
               "             ),',','.'),' ',''))),'.',',')    AS DATA_LENGTH,                              ",
               "       TRIM(DECODE(EB.DZEB012,'',NULL,EB.DZEB012))   AS DATA_DEFAULT                       ", 
               "  FROM DZEB_T EB                                                                           ",
               "       LEFT OUTER JOIN GZTD_T TD ON TD.GZTD001 = EB.DZEB006                                ",
               " WHERE EB.DZEB001 = '",ls_tableName.toLowerCase(),"'                                       ",
               "MINUS                                                                                      ",
               "SELECT ATC.COLUMN_NAME                      AS COLUMN_NAME,                                ",
               "       DECODE(ACC.COLUMN_NAME,NULL,'N','Y') AS ISKEY,                                      ",
               "       ATC.NULLABLE                         AS NULLABLE,                                   ",
               "       DECODE(INSTRB(ATC.DATA_TYPE,'TIMESTAMP'),1,'TIMESTAMP',ATC.DATA_TYPE) AS DATA_TYPE, ",
               "       REPLACE(TO_CHAR(DECODE(                                                             ",
               "               ATC.DATA_TYPE,                                                              ",
               "               'BLOB',4000,                                                                ", #20240529
               "               'CLOB',4000,                                                                ", #20240529
               "               'INT2',NULL,                                                                ", #20240529
               "               'INT4',NULL,                                                                ", #20240529
               "               'INT8',NULL,                                                                ", #20240529
               "               'NUMBER',DECODE(                                                            ",
               "                                NVL(ATC.DATA_SCALE,'0'),                                   ",
               "                                '0',ATC.DATA_PRECISION,                                    ",
               "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE                    ",
               "                              ),                                                           ",
               "               'NUMERIC',DECODE(                                                           ", #20240529 begin
               "                                NVL(ATC.DATA_SCALE,'0'),                                   ",
               "                                '0',ATC.DATA_PRECISION,                                    ",
               "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE                    ", 
               "                              ),                                                           ", #20240529 end
               "               'DATE',7,                                                                   ", #20240529
               "               'TIMESTAMP(0)',7,                                                           ", #20240529
               "               'TIMESTAMP(5)',11,                                                          ", #20240529
               "               'TIMESTAMP(6)',11,                                                          ", #20240529
               "               'TIMESTAMP',DECODE(                                                         ",
               "                                   ATC.DATA_SCALE,0,7,8,7,11                               ", #20240529
               "                                 ),                                                        ",                
               "               'DATE',DECODE(                                                              ",
               "                              ATC.DATA_LENGTH,3,7,ATC.DATA_LENGTH                          ",
               "                            ),                                                             ",                
               "               ATC.DATA_LENGTH                                                             ",
               "             )),'.',',')                    AS DATA_LENGTH,                                ",
               ls_ora_coldef,
               "  FROM DBA_TAB_COLUMNS ATC                                                                 ",
               "  LEFT OUTER JOIN DBA_CONS_COLUMNS ACC ON ACC.OWNER       = ATC.OWNER                      ",
               "                                      AND ACC.TABLE_NAME  = ATC.TABLE_NAME                 ",
               "                                      AND ACC.COLUMN_NAME = ATC.COLUMN_NAME                ",
               "                                      AND ACC.CONSTRAINT_NAME LIKE '%PK'                   ",
               " WHERE 1=1                                                                                 ",
               "   AND ATC.OWNER      = '",ls_owner.toUpperCase(),"'                                       ",
               "   AND ATC.TABLE_NAME = '",ls_tableName.toUpperCase(),"'                                   ",
               "   AND ATC.COLUMN_NAME NOT LIKE 'SYS_%'                                                    ",
               " ORDER BY 1                                                                                "
  &endif
  &endif
  
  PREPARE lpre_check_column_diff_forward FROM ls_sql
  DECLARE lcur_check_column_diff_forward CURSOR FOR lpre_check_column_diff_forward

  OPEN lcur_check_column_diff_forward
  FOREACH lcur_check_column_diff_forward INTO lo_column_diff[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_check_column_diff_forward
  {
  IF li_rec_cnt > 1 THEN
    CALL lo_column_diff.deleteElement(li_rec_cnt)
  END IF  
  }
  
  FREE lpre_check_column_diff_forward
  FREE lcur_check_column_diff_forward
  ##############################################################################

  ##############################################################################
  #反向比對
  &ifdef is_msv
  LET ls_sql = "select atc.column_name                      column_name,   ",                                 
               "       iif(atc.nullable='Y','N','Y')        iskey,         ",                                 
               "       atc.nullable                         nullable,      ",                                 
               "       atc.data_type                        data_type,     ",    
               "       atc.data_length                      data_length,   ",                                 
               "       atc.data_default                     data_default   ",            
               "  from dba_tab_columns atc                                 ",                                 
               "  left outer join dba_cons_columns acc                     ",    
               "               on acc.owner       = atc.owner              ",         
               "              and acc.table_name  = atc.table_name         ",         
               "              and acc.column_name = atc.column_name        ",         
               "              and acc.constraint_name like '%PK'           ",         
               " where 1=1                                                 ",                                 
               "   and atc.owner      = '",ls_owner.toLowerCase(),"'       ",                                 
               "   and atc.table_name = '",ls_tableName.toLowerCase(),"'   ",                                 
               ls_dummy_sql,
               "except                                                     ",                   
               "select eb.dzeb002                        column_name,      ",                            
               "       eb.dzeb004                        iskey,            ",                            
               "       case eb.dzeb004                                     ",    
               "         when 'Y' then 'N'                                 ",    
               "         when 'N' then 'Y'                                 ",    
               "       else                                                ",    
               "         'Y'                                               ",    
               "       end                               nullable,         ",                            
               "       td.gztd003                        data_type,        ",                            
               "       case td.gztd003                                     ",    
               "         when 'varbinary'                                  ",    
               "           then 'max'                                      ",    
               "         when 'nvarchar'                                   ",    
               "           then iif(td.gztd009='TEXT','max',td.gztd008)    ",    
               "       else                                                ",    
               "         td.gztd008                                        ",    
               "       end                               data_length,      ",    
               "       trim(eb.dzeb012)                  data_default      ",                            
               "  from dzeb_t eb                                           ",                                 
               "       left outer join gztd_t td                           ",    
               "                    on td.gztd001 = eb.dzeb006             ",                    
               " where eb.dzeb001 = '",ls_tableName.toLowerCase(),"'       ",                                 
               " order by 1                                                "                                 
  &else
  &ifdef is_pgs
  LET ls_sql = "select atc.column_name                      column_name,                    ",                                 
               "       case when atc.nullable='Y'                                           ",
               "       then 'N' else 'Y' end                iskey,                          ",                                 
               "       atc.nullable                         nullable,                       ",                                 
               "       atc.data_type                        data_type,                      ",    
               "       case                                                                 ",
               "         when atc.data_type = 'text' then null                              ",
               "         when atc.data_type = 'bytea' then null                             ",
               "         when atc.data_type = 'date' then null                              ",
               "         when atc.data_type = 'int2' then null                              ",
               "         when atc.data_type = 'int4' then null                              ",
               "         when atc.data_type = 'int8' then null                              ",
               "         when atc.data_type = 'decimal' then                                ",
               "           atc.data_precision::varchar||','||                               ",
               "           atc.data_scale::varchar                                          ",
               "       else                                                                 ",
               "         atc.data_length::varchar                                           ",
               "       end::varchar                         data_length,                    ",
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
               "              and acc.constraint_name like '%PK'                            ",         
               " where 1=1                                                                  ",                                 
               "   and atc.owner      = '",ls_owner.toLowerCase(),"'                        ",                                 
               "   and atc.table_name = '",ls_tableName.toLowerCase(),"'                    ",                                 
               ls_dummy_sql,
               "except                                                                      ",                   
               "select eb.dzeb002                        column_name,                       ",                            
               "       eb.dzeb004                        iskey,                             ",                            
               "       case eb.dzeb004                                                      ",    
               "         when 'Y' then 'N'                                                  ",    
               "         when 'N' then 'Y'                                                  ",    
               "       else                                                                 ",    
               "         'Y'                                                                ",    
               "       end                               nullable,                          ",                            
               "       td.gztd003                        data_type,                         ",                            
               "       td.gztd008                        data_length,                       ",    
               "       trim(eb.dzeb012)                  data_default                       ",                            
               "  from dzeb_t eb                                                            ",                                 
               "       left outer join gztd_t td                                            ",    
               "                    on td.gztd001 = eb.dzeb006                              ",                    
               " where eb.dzeb001 = '",ls_tableName.toLowerCase(),"'                        ",                                 
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
    LET ls_ora_coldef = " DECODE(TRIM(DS.GET_COL_DEFAULT(ATC.TABLE_NAME,ATC.COLUMN_NAME)),'',NULL,       ",
                        "        TRIM(DS.GET_COL_DEFAULT(ATC.TABLE_NAME,ATC.COLUMN_NAME))) DATA_DEFAULT  "
  END CASE
  
  LET ls_sql = "SELECT ATC.COLUMN_NAME                      AS COLUMN_NAME,                             ",
               "       DECODE(ACC.COLUMN_NAME,NULL,'N','Y') AS ISKEY,                                   ",
               "       ATC.NULLABLE                         AS NULLABLE,                                ",
               "       DECODE(INSTRB(ATC.DATA_TYPE,'TIMESTAMP'),1,'TIMESTAMP',ATC.DATA_TYPE) AS DATA_TYPE, ",
               "       REPLACE(TO_CHAR(DECODE(                                                          ",
               "               ATC.DATA_TYPE,                                                           ",
               "               'BLOB',4000,                                                             ", #20240529
               "               'CLOB',4000,                                                             ", #20240529
               "               'INT2',NULL,                                                             ", #20240529
               "               'INT4',NULL,                                                             ", #20240529
               "               'INT8',NULL,                                                             ", #20240529
               "               'NUMBER',DECODE(                                                         ",
               "                                NVL(ATC.DATA_SCALE,'0'),                                ",
               "                                '0',ATC.DATA_PRECISION,                                 ",
               "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE                 ",
               "                              ),                                                        ",
               "               'NUMERIC',DECODE(                                                        ", #20240529 begin
               "                                NVL(ATC.DATA_SCALE,'0'),                                ",
               "                                '0',ATC.DATA_PRECISION,                                 ",
               "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE                 ", 
               "                              ),                                                        ", #20240529 end
               "               'DATE',7,                                                                ", #20240529
               "               'TIMESTAMP(0)',7,                                                        ", #20240529
               "               'TIMESTAMP(5)',11,                                                       ", #20240529
               "               'TIMESTAMP(6)',11,                                                       ", #20240529
               "               'TIMESTAMP',DECODE(                                                      ",
               "                                   ATC.DATA_SCALE,0,7,8,7,11                            ", #20240529
               "                                 ),                                                     ",                
               "               'DATE',DECODE(                                                           ",
               "                              ATC.DATA_LENGTH,3,7,ATC.DATA_LENGTH                       ",
               "                            ),                                                          ",                
               "               ATC.DATA_LENGTH                                                          ",
               "             )),'.',',')                    AS DATA_LENGTH,                             ",
               ls_ora_coldef,
               "  FROM DBA_TAB_COLUMNS ATC                                                              ",
               "  LEFT OUTER JOIN DBA_CONS_COLUMNS ACC ON ACC.OWNER       = ATC.OWNER                   ",
               "                                      AND ACC.TABLE_NAME  = ATC.TABLE_NAME              ",
               "                                      AND ACC.COLUMN_NAME = ATC.COLUMN_NAME             ",
               "                                      AND ACC.CONSTRAINT_NAME LIKE '%PK'                ",
               " WHERE 1=1                                                                              ",
               "   AND ATC.OWNER      = '",ls_owner.toUpperCase(),"'                                    ",
               "   AND ATC.TABLE_NAME = '",ls_tableName.toUpperCase(),"'                                ",
               "   AND ATC.COLUMN_NAME NOT LIKE 'SYS_%'                                                 ",
               ls_dummy_sql,
               "MINUS                                                                                   ",
               "SELECT UPPER(EB.DZEB002)                      AS COLUMN_NAME,                           ",
               "       EB.DZEB004                             AS ISKEY,                                 ",
               "       DECODE(EB.DZEB004,'Y','N','N','Y','Y') AS NULLABLE,                              ",
               "       UPPER(TD.GZTD003)                      AS DATA_TYPE,                             ",
               "       REPLACE(TO_CHAR(TO_NUMBER(REPLACE(REPLACE(                                       ",                                  
               "       DECODE(                                                                          ",
               "              UPPER(TD.GZTD003),                                                        ",
               "              'BLOB','4000',                                                            ",
               "              'CLOB','4000',                                                            ",
               "              'DATE','7',                                                               ",
               "              'TIMESTAMP',DECODE(TD.GZTD008,'0','7','11'),                              ",
               "              TD.GZTD008                                                                ",
               "             ),',','.'),' ',''))),'.',',')    AS DATA_LENGTH,                           ",
               "       TRIM(DECODE(EB.DZEB012,'',NULL,EB.DZEB012))    AS DATA_DEFAULT                   ",
               "  FROM DZEB_T EB                                                                        ",
               "       LEFT OUTER JOIN GZTD_T TD ON TD.GZTD001 = EB.DZEB006                             ",
               " WHERE EB.DZEB001 = '",ls_tableName.toLowerCase(),"'                                    ",
               " ORDER BY 1                                                                             "
  &endif
  &endif
  
  PREPARE lpre_check_column_diff_reverse FROM ls_sql
  DECLARE lcur_check_column_diff_reverse CURSOR FOR lpre_check_column_diff_reverse

  OPEN lcur_check_column_diff_reverse
  FOREACH lcur_check_column_diff_reverse INTO lo_column_diff[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_check_column_diff_reverse
  {
  IF li_rec_cnt > 1 THEN
    CALL lo_column_diff.deleteElement(li_rec_cnt)
  END IF  
  }
  FREE lpre_check_column_diff_reverse
  FREE lcur_check_column_diff_reverse
  ##############################################################################
  
  IF li_rec_cnt >= 2 THEN
    LET lb_different = TRUE
  END IF  

  LET lo_different.* = lo_column_diff.*
  
  RETURN lb_different, lo_different
  
END FUNCTION

#取得 Table 之間的比對(By TableEditor)
FUNCTION sadzi140_shot_get_column_diff_by_table_editor(p_owner,p_table_name)
DEFINE
  p_owner       STRING, 
  p_table_name  STRING
DEFINE
  lo_column_diff  DYNAMIC ARRAY OF T_COLUMN_DIFF,
  ls_owner        STRING, 
  ls_tablename    STRING,
  ls_ora_coldef   STRING, #20240529
  li_rec_cnt      INTEGER,
  ls_db_type      STRING, #20240606
  ls_sql          STRING
DEFINE
  lb_different    BOOLEAN,
  lo_different    DYNAMIC ARRAY OF T_COLUMN_DIFF 

  LET ls_owner     = p_owner
  LET ls_tablename = p_table_name

  LET lb_different    = FALSE
  LET lo_different    = NULL

  &ifdef is_msv
  LET ls_sql = "select eb.dzeb002                             dzev005,   ",                               
               "       eb.dzeb004                             dzev006,   ",                               
               "       iif(eb.dzeb004='Y','N','Y')            dzev007,   ",                               
               "       td.gztd003                             dzev008,   ",                               
               "       case td.gztd003                                   ",  
               "         when 'varbinary'                                ",  
               "           then 'max'                                    ",  
               "         when 'nvarchar'                                 ",  
               "           then iif(td.gztd009='TEXT','max',td.gztd008)  ",  
               "       else                                              ",  
               "         td.gztd008                                      ",  
               "       end                                    dzev009,   ",                               
               "       trim(eb.dzeb012)                       dzev021    ",                               
               "  from dzeb_t eb                                         ",                            
               "       left outer join gztd_t td                         ",  
               "                    on td.gztd001 = eb.dzeb006           ",               
               " where eb.dzeb001 = '",ls_tableName.toLowerCase(),"'     ",                               
               "except                                                   ",  
               "select atc.column_name                      dzev005,     ",                               
               "       iif(atc.nullable='Y','N','Y')        dzev006,     ",                               
               "       atc.nullable                         dzev007,     ",   
               "       atc.data_type                        dzev008,     ",  
               "       atc.data_length                      dzev009,     ",                               
               "       atc.data_default                     dzev021      ",          
               "  from dba_tab_columns atc                               ",                               
               " where 1=1                                               ",                               
               "   and atc.owner      = '",ls_owner.toLowerCase(),"'     ",                               
               "   and atc.table_name = '",ls_tableName.toLowerCase(),"' ",                               
               " order by 1                                              "                               
  &else
  &ifdef is_pgs
  LET ls_sql = "select eb.dzeb002                             dzev005,   ",                               
               "       eb.dzeb004                             dzev006,   ",                               
               "       case when eb.dzeb004='Y'                          ",
               "       then 'N' else 'Y' end                  dzev007,   ",                               
               "       td.gztd003                             dzev008,   ",                               
               "       td.gztd008                             dzev009,   ",                               
               "       trim(eb.dzeb012)                       dzev021    ",                               
               "  from dzeb_t eb                                         ",                            
               "       left outer join gztd_t td                         ",  
               "                    on td.gztd001 = eb.dzeb006           ",               
               " where eb.dzeb001 = '",ls_tableName.toLowerCase(),"'     ",                               
               "except                                                   ",  
               "select atc.column_name                      dzev005,     ",                               
               "       case when atc.nullable='Y'                                           ",
               "       then 'N' else 'Y' end                dzev006,                        ",                               
               "       atc.nullable                         dzev007,                        ",   
               "       atc.data_type                        dzev008,                        ",  
               "       case                                                                 ",
               "         when atc.data_type = 'text' then null                              ",
               "         when atc.data_type = 'bytea' then null                             ",
               "         when atc.data_type = 'date' then null                              ",
               "         when atc.data_type = 'int2' then null                              ",
               "         when atc.data_type = 'int4' then null                              ",
               "         when atc.data_type = 'int8' then null                              ",
               "         when atc.data_type = 'decimal' then                                ",
               "           atc.data_precision::varchar||','||                               ",
               "           atc.data_scale::varchar                                          ",
               "       else                                                                 ",
               "         atc.data_length::varchar                                           ",
               "       end::varchar                         dzev009,                        ",
               "       case                                                                 ",
               "         when atc.data_default is null then null                            ",
               "         when atc.data_default = 'NULL' then null                           ",
               "         when atc.data_default = 'null' then null                           ",
               "         when atc.data_default like 'CURRENT%' then lower(atc.data_default) ",
               "         when atc.data_default = '\"sysdate\"()' then 'sysdate'             ",
               "         when atc.data_default = 'SYSDATE' then 'sysdate'                   ",
               "       else                                                                 ",
               "         atc.data_default                                                   ",
               "       end as dzev021                                                       ",
               "  from dba_tab_columns atc                                                  ",                               
               " where 1=1                                                                  ",                               
               "   and atc.owner      = '",ls_owner.toLowerCase(),"'                        ",                               
               "   and atc.table_name = '",ls_tableName.toLowerCase(),"'                    ",                               
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
                          " END AS DZEV021                                                       "
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
                          " END AS DZEV021                                                       "
  OTHERWISE  
    LET ls_ora_coldef = " DECODE(TRIM(DS.GET_COL_DEFAULT(ATC.TABLE_NAME,ATC.COLUMN_NAME)),'',NULL,  ",
                        "        TRIM(DS.GET_COL_DEFAULT(ATC.TABLE_NAME,ATC.COLUMN_NAME))) DZEV021  "
  END CASE
  
  LET ls_sql = "SELECT UPPER(EB.DZEB002)                      DZEV005,                                ",
               "       EB.DZEB004                             DZEV006,                                ",
               "       DECODE(EB.DZEB004,'Y','N','N','Y','Y') DZEV007,                                ",
               "       UPPER(TD.GZTD003)                      DZEV008,                                ",
               "       REPLACE(TO_CHAR(TO_NUMBER(REPLACE(REPLACE(                                     ",                                  
               "       DECODE(                                                                        ",
               "              UPPER(TD.GZTD003),                                                      ",
               "              'BLOB','4000',                                                          ",
               "              'CLOB','4000',                                                          ",
               "              'DATE','7',                                                             ",
               "              'TIMESTAMP',DECODE(TD.GZTD008,'0','7','11'),                            ",
               "              TD.GZTD008                                                              ",
               "             ),',','.'),' ',''))),'.',',')    DZEV009,                                ",
               "       TRIM(DECODE(EB.DZEB012,'',NULL,EB.DZEB012))   DZEV021                          ",
               "  FROM DS.DZEB_T EB                                                                   ",
               "       LEFT OUTER JOIN DS.GZTD_T TD ON TD.GZTD001 = EB.DZEB006                        ",
               " WHERE EB.DZEB001 = '",ls_tableName.toLowerCase(),"'                                  ",
               "MINUS                                                                                 ",
               "SELECT ATC.COLUMN_NAME                      DZEV005,                                  ",
               "       DECODE(ACC.COLUMN_NAME,NULL,'N','Y') DZEV006,                                  ",
               "       ATC.NULLABLE                         DZEV007,                                  ",
               "       DECODE(INSTRB(ATC.DATA_TYPE,'TIMESTAMP'),1,'TIMESTAMP',ATC.DATA_TYPE) DZEV008, ",
               "       REPLACE(TO_CHAR(DECODE(                                                        ",
               "               ATC.DATA_TYPE,                                                         ",
               "               'BLOB',4000,                                                           ", #20240529
               "               'CLOB',4000,                                                           ", #20240529
               "               'INT2',NULL,                                                           ", #20240529
               "               'INT4',NULL,                                                           ", #20240529
               "               'INT8',NULL,                                                           ", #20240529
               "               'NUMBER',DECODE(                                                       ",
               "                                NVL(ATC.DATA_SCALE,'0'),                              ",
               "                                '0',ATC.DATA_PRECISION,                               ",
               "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE               ",
               "                              ),                                                      ",
               "               'NUMERIC',DECODE(                                                      ", #20240529 begin
               "                                NVL(ATC.DATA_SCALE,'0'),                              ",
               "                                '0',ATC.DATA_PRECISION,                               ",
               "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE               ", 
               "                              ),                                                      ", #20240529 end
               "               'DATE',7,                                                              ", #20240529
               "               'TIMESTAMP(0)',7,                                                      ", #20240529
               "               'TIMESTAMP(5)',11,                                                     ", #20240529
               "               'TIMESTAMP(6)',11,                                                     ", #20240529
               "               'TIMESTAMP',DECODE(                                                    ",
               "                                   ATC.DATA_SCALE,0,7,8,7,11                          ", #20240529
               "                                 ),                                                   ",                
               "               'DATE',DECODE(                                                         ",
               "                              ATC.DATA_LENGTH,3,7,ATC.DATA_LENGTH                     ",
               "                            ),                                                        ",                
               "               ATC.DATA_LENGTH                                                        ",
               "             )),'.',',')                    DZEV009,                                  ",
               ls_ora_coldef,
               "  FROM DBA_TAB_COLUMNS ATC                                                            ",
               "  LEFT OUTER JOIN DBA_CONS_COLUMNS ACC ON ACC.OWNER       = ATC.OWNER                 ",
               "                                      AND ACC.TABLE_NAME  = ATC.TABLE_NAME            ",
               "                                      AND ACC.COLUMN_NAME = ATC.COLUMN_NAME           ",
               "                                      AND ACC.CONSTRAINT_NAME LIKE '%PK'              ",
               " WHERE 1=1                                                                            ",
               "   AND ATC.OWNER      = '",ls_owner.toUpperCase(),"'                                  ",
               "   AND ATC.TABLE_NAME = '",ls_tableName.toUpperCase(),"'                              ",
               "   AND ATC.COLUMN_NAME NOT LIKE 'SYS_%'                                               ",
               " ORDER BY 1                                                                           "
  &endif             
  &endif             

  PREPARE lpre_get_diff_list_by_editor FROM ls_sql
  DECLARE lcur_get_diff_list_by_editor CURSOR FOR lpre_get_diff_list_by_editor

  LET li_rec_cnt = 1 
  
  OPEN lcur_get_diff_list_by_editor
  FOREACH lcur_get_diff_list_by_editor INTO lo_column_diff[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_diff_list_by_editor
  IF li_rec_cnt > 1 THEN
    CALL lo_column_diff.deleteElement(li_rec_cnt)
  END If  
  
  FREE lpre_get_diff_list_by_editor
  FREE lcur_get_diff_list_by_editor

  IF li_rec_cnt >= 2 THEN
    LET lb_different = TRUE
  END IF  

  LET lo_different.* = lo_column_diff.*
  
  RETURN lb_different, lo_different
  
END FUNCTION

#取得 Table 之間的比對(By Database)
FUNCTION sadzi140_shot_get_column_diff_by_database(p_owner,p_table_name,p_drop_abnormal_column)
DEFINE
  p_owner      STRING, 
  p_table_name STRING,
  p_drop_abnormal_column BOOLEAN
DEFINE
  ls_owner        STRING, 
  ls_tablename    STRING,
  lb_drop_abnormal_column BOOLEAN,
  lo_column_diff  DYNAMIC ARRAY OF T_COLUMN_DIFF,
  li_rec_cnt      INTEGER,
  ls_ora_coldef   STRING, #20240529
  ls_sql          STRING,
  ls_db_type      STRING, #20240606
  ls_dummy_sql    STRING
DEFINE
  lb_different    BOOLEAN,
  lo_different    DYNAMIC ARRAY OF T_COLUMN_DIFF 

  LET ls_owner     = p_owner
  LET ls_tablename = p_table_name
  LET lb_drop_abnormal_column = p_drop_abnormal_column

  LET lb_different    = FALSE
  LET lo_different    = NULL

  IF lb_drop_abnormal_column THEN
    LET ls_dummy_sql = ""
  ELSE
    &ifdef is_msv
    LET ls_dummy_sql = " and atc.column_name = 'dummy' "
    &else
    &ifdef is_pgs
    LET ls_dummy_sql = " and atc.column_name = 'dummy' "
    &else
    LET ls_dummy_sql = " AND ATC.COLUMN_NAME = 'DUMMY' "
    &endif
    &endif
  END IF  

  &ifdef is_msv
  LET ls_sql = "select atc.column_name                      dzev005,     ",                               
               "       iif(atc.nullable='Y','N','Y')        dzev006,     ",                               
               "       atc.nullable                         dzev007,     ",   
               "       atc.data_type                        dzev008,     ",  
               "       atc.data_length                      dzev009,     ",                               
               "       atc.data_default                     dzev021      ",          
               "  from dba_tab_columns atc                               ",                               
               " where 1=1                                               ",                               
               "   and atc.owner      = '",ls_owner.toLowerCase(),"'     ",                               
               "   and atc.table_name = '",ls_tableName.toLowerCase(),"' ",                               
               ls_dummy_sql,
               "except                                                   ",  
               "select eb.dzeb002                             dzev005,   ",                               
               "       eb.dzeb004                             dzev006,   ",                               
               "       iif(eb.dzeb004='Y','N','Y')            dzev007,   ",                               
               "       td.gztd003                             dzev008,   ",                               
               "       case td.gztd003                                   ",  
               "         when 'varbinary'                                ",  
               "           then 'max'                                    ",  
               "         when 'nvarchar'                                 ",  
               "           then iif(td.gztd009='TEXT','max',td.gztd008)  ",  
               "       else                                              ",  
               "         td.gztd008                                      ",  
               "       end                                    dzev009,   ",                               
               "       trim(eb.dzeb012)                       dzev021    ",                               
               "  from dzeb_t eb                                         ",                            
               "       left outer join gztd_t td                         ",  
               "                    on td.gztd001 = eb.dzeb006           ",               
               " where eb.dzeb001 = '",ls_tableName.toLowerCase(),"'     ",                               
               " order by 1                                              "                               
  &else
  &ifdef is_pgs
  LET ls_sql = "select atc.column_name                      dzev005,                        ",                               
               "       case when atc.nullable='Y'                                           ",
               "       then 'N' else 'Y' end                dzev006,                        ",                               
               "       atc.nullable                         dzev007,                        ",   
               "       atc.data_type                        dzev008,                        ",  
               "       case                                                                 ",
               "         when atc.data_type = 'text' then null                              ",
               "         when atc.data_type = 'bytea' then null                             ",
               "         when atc.data_type = 'date' then null                              ",
               "         when atc.data_type = 'int2' then null                              ",
               "         when atc.data_type = 'int4' then null                              ",
               "         when atc.data_type = 'int8' then null                              ",
               "         when atc.data_type = 'decimal' then                                ",
               "           atc.data_precision::varchar||','||                               ",
               "           atc.data_scale::varchar                                          ",
               "       else                                                                 ",
               "         atc.data_length::varchar                                           ",
               "       end::varchar                         dzev009,                        ",
               "       case                                                                 ",
               "         when atc.data_default is null then null                            ",
               "         when atc.data_default = 'NULL' then null                           ",
               "         when atc.data_default = 'null' then null                           ",
               "         when atc.data_default like 'CURRENT%' then lower(atc.data_default) ",
               "         when atc.data_default = '\"sysdate\"()' then 'sysdate'             ",
               "         when atc.data_default = 'SYSDATE' then 'sysdate'                   ",
               "       else                                                                 ",
               "         atc.data_default                                                   ",
               "       end as dzev021                                                       ",
               "  from dba_tab_columns atc                                                  ",                               
               " where 1=1                                                                  ",                               
               "   and atc.owner      = '",ls_owner.toLowerCase(),"'                        ",                               
               "   and atc.table_name = '",ls_tableName.toLowerCase(),"'                    ",                               
               ls_dummy_sql,
               "except                                                                      ",  
               "select eb.dzeb002                             dzev005,                      ",                               
               "       eb.dzeb004                             dzev006,                      ",                               
               "       case when eb.dzeb004='Y'                                             ",
               "       then 'N' else 'Y' end                  dzev007,                      ",                               
               "       td.gztd003                             dzev008,                      ",                               
               "       td.gztd008                             dzev009,                      ",                               
               "       trim(eb.dzeb012)                       dzev021                       ",                               
               "  from dzeb_t eb                                                            ",                            
               "       left outer join gztd_t td                                            ",  
               "                    on td.gztd001 = eb.dzeb006                              ",               
               " where eb.dzeb001 = '",ls_tableName.toLowerCase(),"'                        ",                               
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
                          " END AS DZEV021                                                       "
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
                          " END AS DZEV021                                                       "
  OTHERWISE  
    LET ls_ora_coldef = " DECODE(TRIM(DS.GET_COL_DEFAULT(ATC.TABLE_NAME,ATC.COLUMN_NAME)),'',NULL,  ",
                        "        TRIM(DS.GET_COL_DEFAULT(ATC.TABLE_NAME,ATC.COLUMN_NAME))) DZEV021  "
  END CASE
    
  LET ls_sql = "SELECT ATC.COLUMN_NAME                      DZEV005,                                  ",
               "       DECODE(ACC.COLUMN_NAME,NULL,'N','Y') DZEV006,                                  ",
               "       ATC.NULLABLE                         DZEV007,                                  ",
               "       DECODE(INSTRB(ATC.DATA_TYPE,'TIMESTAMP'),1,'TIMESTAMP',ATC.DATA_TYPE) DZEV008, ",
               "       REPLACE(TO_CHAR(DECODE(                                                        ",
               "               ATC.DATA_TYPE,                                                         ",
               "               'BLOB',4000,                                                           ", #20240529
               "               'CLOB',4000,                                                           ", #20240529
               "               'INT2',NULL,                                                           ", #20240529
               "               'INT4',NULL,                                                           ", #20240529
               "               'INT8',NULL,                                                           ", #20240529
               "               'NUMBER',DECODE(                                                       ",
               "                                NVL(ATC.DATA_SCALE,'0'),                              ",
               "                                '0',ATC.DATA_PRECISION,                               ",
               "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE               ",
               "                              ),                                                      ",
               "               'NUMERIC',DECODE(                                                      ", #20240529 begin
               "                                NVL(ATC.DATA_SCALE,'0'),                              ",
               "                                '0',ATC.DATA_PRECISION,                               ",
               "                                ATC.DATA_PRECISION||'.'||ATC.DATA_SCALE               ", 
               "                              ),                                                      ", #20240529 end
               "               'DATE',7,                                                              ", #20240529
               "               'TIMESTAMP(0)',7,                                                      ", #20240529
               "               'TIMESTAMP(5)',11,                                                     ", #20240529
               "               'TIMESTAMP(6)',11,                                                     ", #20240529
               "               'TIMESTAMP',DECODE(                                                    ",
               "                                   ATC.DATA_SCALE,0,7,8,7,11                          ", #20240529
               "                                 ),                                                   ",                
               "               'DATE',DECODE(                                                         ",
               "                              ATC.DATA_LENGTH,3,7,ATC.DATA_LENGTH                     ",
               "                            ),                                                        ",                
               "               ATC.DATA_LENGTH                                                        ",
               "             )),'.',',')                    DZEV009,                                  ",
               ls_ora_coldef,
               "  FROM DBA_TAB_COLUMNS ATC                                                            ",
               "  LEFT OUTER JOIN DBA_CONS_COLUMNS ACC ON ACC.OWNER       = ATC.OWNER                 ",
               "                                      AND ACC.TABLE_NAME  = ATC.TABLE_NAME            ",
               "                                      AND ACC.COLUMN_NAME = ATC.COLUMN_NAME           ",
               "                                      AND ACC.CONSTRAINT_NAME LIKE '%PK'              ",
               " WHERE 1=1                                                                            ",
               "   AND ATC.OWNER      = '",ls_owner.toUpperCase(),"'                                  ",
               "   AND ATC.TABLE_NAME = '",ls_tableName.toUpperCase(),"'                              ",
               "   AND ATC.COLUMN_NAME NOT LIKE 'SYS_%'                                               ",
               ls_dummy_sql,
               "MINUS                                                                                 ",
               "SELECT UPPER(EB.DZEB002)                      DZEV005,                                ",
               "       EB.DZEB004                             DZEV006,                                ",
               "       DECODE(EB.DZEB004,'Y','N','N','Y','Y') DZEV007,                                ",
               "       UPPER(TD.GZTD003)                      DZEV008,                                ",
               "       REPLACE(TO_CHAR(TO_NUMBER(REPLACE(REPLACE(                                     ",                                  
               "       DECODE(                                                                        ",
               "              UPPER(TD.GZTD003),                                                      ",
               "              'BLOB','4000',                                                          ",
               "              'CLOB','4000',                                                          ",
               "              'DATE','7',                                                             ",
               "              'TIMESTAMP',DECODE(TD.GZTD008,'0','7','11'),                            ",
               "              TD.GZTD008                                                              ",
               "             ),',','.'),' ',''))),'.',',')    DZEV009,                                ",
               "       TRIM(DECODE(EB.DZEB012,'',NULL,EB.DZEB012))   DZEV021                          ",
               "  FROM DS.DZEB_T EB                                                                   ",
               "       LEFT OUTER JOIN DS.GZTD_T TD ON TD.GZTD001 = EB.DZEB006                        ",
               " WHERE EB.DZEB001 = '",ls_tableName.toLowerCase(),"'                                  ",
               " ORDER BY 1                                                                           "
  &endif
  &endif
  
  PREPARE lpre_get_diff_list_by_database FROM ls_sql
  DECLARE lcur_get_diff_list_by_database CURSOR FOR lpre_get_diff_list_by_database

  LET li_rec_cnt = 1 
  
  OPEN lcur_get_diff_list_by_database
  FOREACH lcur_get_diff_list_by_database INTO lo_column_diff[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_diff_list_by_database
  IF li_rec_cnt > 1 THEN
    CALL lo_column_diff.deleteElement(li_rec_cnt)
  END IF  
  
  FREE lpre_get_diff_list_by_database
  FREE lcur_get_diff_list_by_database

  IF li_rec_cnt >= 2 THEN
    LET lb_different = TRUE
  END IF  

  LET lo_different.* = lo_column_diff.*
  
  RETURN lb_different, lo_different
  

END FUNCTION

#比對現行 Table Editor 的 Constraint 資料和最近一次的 Snapshot 的差異(正反向都要抓)
FUNCTION sadzi140_shot_check_constraint_diff(p_owner,p_table_name,p_drop_abnormal_column)
DEFINE
  p_owner       STRING, 
  p_table_name  STRING,
  p_drop_abnormal_column BOOLEAN
DEFINE
  ls_owner      STRING, 
  ls_table_name STRING,
  lb_drop_abnormal_column BOOLEAN,
  li_rec_cnt    INTEGER,
  ls_sql        STRING
DEFINE
  lo_constraint_diff  DYNAMIC ARRAY OF T_CONSTRAINT_DIFF,
  lb_different        BOOLEAN,
  lo_different        DYNAMIC ARRAY OF T_CONSTRAINT_DIFF 

  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name
  LET lb_drop_abnormal_column = p_drop_abnormal_column

  LET lb_different = FALSE
  LET lo_different = NULL
  LET li_rec_cnt = 1 
  
  ##############################################################################
  #正向比對
  &ifdef is_msv
  LET ls_sql = "select ed.dzed001       table_name,                                           ",                                                                     
               "       ed.dzed002       constraint_name,                                      ",                                                                      
               "       ed.dzed003       constraint_type,                                      ",                                                                      
               "       ed.dzed004       column_names,                                         ",                                                                      
               "       ed.dzed005       r_table_name,                                         ",                                                                      
               "       ed.dzed006       r_column_names                                        ",                                                                      
               "  from dzed_t ed                                                              ",                                                                             
               " where ed.dzed001 = '",ls_table_name.toLowerCase(),"'                         ",                                                    
               "   and ed.dzed003 <> 'F'                                                      ",                                                                             
               "except                                                                        ",                                                                              
               "select acs.table_name,                                                        ",                                                                             
               "       acs.constraint_name,                                                   ",  
               "       acs.constraint_type,                                                   ",  
               "       acc.column_names,                                                      ",                                                                             
               "       NULL r_table_name,                                                     ",                                                                  
               "       NULL r_column_names                                                    ",                                                                               
               "  from dba_constraints acs                                                    ",                                                                             
               "  left outer join (                                                           ",                                                                             
               "                   select accs.constraint_name,accs.owner,accs.table_name,    ",  
               "                          string_agg(accs.column_name, ',')                   ",  
               "                          within group(order by accs.position) column_names   ",   
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
  LET ls_sql = "select ed.dzed001       table_name,                                           ",                                                                     
               "       ed.dzed002       constraint_name,                                      ",                                                                      
               "       ed.dzed003       constraint_type,                                      ",                                                                      
               "       ed.dzed004       column_names,                                         ",                                                                      
               "       ed.dzed005       r_table_name,                                         ",                                                                      
               "       ed.dzed006       r_column_names                                        ",                                                                      
               "  from dzed_t ed                                                              ",                                                                             
               " where ed.dzed001 = '",ls_table_name.toLowerCase(),"'                         ",                                                    
               "   and ed.dzed003 <> 'F'                                                      ",                                                                             
               "except                                                                        ",                                                                              
               "select acs.table_name,                                                        ",                                                                             
               "       acs.constraint_name,                                                   ",  
               "       acs.constraint_type,                                                   ",  
               "       acc.column_names,                                                      ",                                                                             
               "       NULL r_table_name,                                                     ",                                                                  
               "       NULL r_column_names                                                    ",                                                                               
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
  LET ls_sql = "SELECT UPPER(ED.DZED001)       TABLE_NAME,                                                                                                                  ",
               "       UPPER(ED.DZED002)       CONSTRAINT_NAME,                                                                                                             ",
               "       UPPER(ED.DZED003)       CONSTRAINT_TYPE,                                                                                                             ",
               "       UPPER(ED.DZED004)       COLUMN_NAMES,                                                                                                                ",
               "       UPPER(TRIM(ED.DZED005)) R_TABLE_NAME,                                                                                                                ",
               "       UPPER(TRIM(ED.DZED006)) R_COLUMN_NAMES                                                                                                               ",
               "  FROM DZED_T ED                                                                                                                                            ",
               " WHERE ED.DZED001 = '",ls_table_name.toLowerCase(),"'                                                                                                       ",
               "   AND ED.DZED003 <> 'F'                                                                                                                                    ",         
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
               "   AND ACS.CONSTRAINT_NAME NOT LIKE 'SYS%'                                                                                                                  ", #20180903
               "   AND ACS.TABLE_NAME = '",ls_table_name.toUpperCase(),"'                                                                                                   ",
               "   AND ACS.OWNER      = '",ls_owner.toUpperCase(),"'                                                                                                        "
  &endif             
  &endif             
               
  PREPARE lpre_check_key_diff_forward FROM ls_sql
  DECLARE lcur_check_key_diff_forward CURSOR FOR lpre_check_key_diff_forward

  OPEN lcur_check_key_diff_forward
  FOREACH lcur_check_key_diff_forward INTO lo_constraint_diff[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_check_key_diff_forward
  
  FREE lpre_check_key_diff_forward
  FREE lcur_check_key_diff_forward
  ##############################################################################

  ##############################################################################
  #反向比對
  &ifdef is_msv
  LET ls_sql = 
               "select acs.table_name,                                                        ",                                                                             
               "       acs.constraint_name,                                                   ",  
               "       acs.constraint_type,                                                   ",  
               "       acc.column_names,                                                      ",                                                                             
               "       NULL r_table_name,                                                     ",                                                                  
               "       NULL r_column_names                                                    ",                                                                               
               "  from dba_constraints acs                                                    ",                                                                             
               "  left outer join (                                                           ",                                                                             
               "                   select accs.constraint_name,accs.owner,accs.table_name,    ",  
               "                          string_agg(accs.column_name, ',')                   ",  
               "                          within group(order by accs.position) column_names   ",   
               "                     from dba_cons_columns accs                               ",                                                                             
               "                    group by accs.constraint_name,accs.owner,accs.table_name  ",                                                                             
               "                  ) acc on acc.constraint_name = acs.constraint_name          ",                                                                             
               "                       and acc.table_name      = acs.table_name               ",                                                                             
               "                       and acc.owner           = acs.owner                    ",                                                                             
               " where 1 = 1                                                                  ",                                                                             
               "   and acs.table_name = '",ls_table_name.toLowerCase(),"'                     ",                                                    
               "   and acs.owner      = '",ls_owner.toLowerCase(),"'                          ",
               "except                                                                        ",                                                                              
               "select ed.dzed001       table_name,                                           ",                                                                     
               "       ed.dzed002       constraint_name,                                      ",                                                                      
               "       ed.dzed003       constraint_type,                                      ",                                                                      
               "       ed.dzed004       column_names,                                         ",                                                                      
               "       ed.dzed005       r_table_name,                                         ",                                                                      
               "       ed.dzed006       r_column_names                                        ",                                                                      
               "  from dzed_t ed                                                              ",                                                                             
               " where ed.dzed001 = '",ls_table_name.toLowerCase(),"'                         ",                                                    
               "   and ed.dzed003 <> 'F'                                                      "                                                                             
  &else               
  &ifdef is_pgs
  LET ls_sql = 
               "select acs.table_name,                                                        ",                                                                             
               "       acs.constraint_name,                                                   ",  
               "       acs.constraint_type,                                                   ",  
               "       acc.column_names,                                                      ",                                                                             
               "       NULL r_table_name,                                                     ",                                                                  
               "       NULL r_column_names                                                    ",                                                                               
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
               "   and acs.owner      = '",ls_owner.toLowerCase(),"'                          ",
               "except                                                                        ",                                                                              
               "select ed.dzed001       table_name,                                           ",                                                                     
               "       ed.dzed002       constraint_name,                                      ",                                                                      
               "       ed.dzed003       constraint_type,                                      ",                                                                      
               "       ed.dzed004       column_names,                                         ",                                                                      
               "       ed.dzed005       r_table_name,                                         ",                                                                      
               "       ed.dzed006       r_column_names                                        ",                                                                      
               "  from dzed_t ed                                                              ",                                                                             
               " where ed.dzed001 = '",ls_table_name.toLowerCase(),"'                         ",                                                    
               "   and ed.dzed003 <> 'F'                                                      "                                                                             
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
               "   AND ACS.CONSTRAINT_NAME NOT LIKE 'SYS%'                                                                                                                  ", #20180903
               "   AND ACS.TABLE_NAME = '",ls_table_name.toUpperCase(),"'                                                                                                   ",
               "   AND ACS.OWNER      = '",ls_owner.toUpperCase(),"'                                                                                                        ",
               "MINUS                                                                                                                                                       ",
               "SELECT UPPER(ED.DZED001)       TABLE_NAME,                                                                                                                  ",
               "       UPPER(ED.DZED002)       CONSTRAINT_NAME,                                                                                                             ",
               "       UPPER(ED.DZED003)       CONSTRAINT_TYPE,                                                                                                             ",
               "       UPPER(ED.DZED004)       COLUMN_NAMES,                                                                                                                ",
               "       UPPER(TRIM(ED.DZED005)) R_TABLE_NAME,                                                                                                                ",
               "       UPPER(TRIM(ED.DZED006)) R_COLUMN_NAMES                                                                                                               ",
               "  FROM DZED_T ED                                                                                                                                            ",
               " WHERE ED.DZED001 = '",ls_table_name.toLowerCase(),"'                                                                                                       ",
               "   AND ED.DZED003 <> 'F'                                                                                                                                    "
  &endif               
  &endif               

  PREPARE lpre_check_key_diff_reverse FROM ls_sql
  DECLARE lcur_check_key_diff_reverse CURSOR FOR lpre_check_key_diff_reverse

  OPEN lcur_check_key_diff_reverse
  FOREACH lcur_check_key_diff_reverse INTO lo_constraint_diff[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_check_key_diff_reverse

  FREE lpre_check_key_diff_reverse
  FREE lcur_check_key_diff_reverse
  ##############################################################################
  
  IF li_rec_cnt >= 2 THEN
    LET lb_different = TRUE
  END IF  

  LET lo_different.* = lo_constraint_diff.*
  
  RETURN lb_different, lo_different
  
END FUNCTION

#取得 Constraint 資料的比對(TableEditor)
FUNCTION sadzi140_shot_get_constraint_diff_by_table_editor(p_owner,p_table_name)
DEFINE
  p_owner        STRING,
  p_table_name   STRING
DEFINE
  ls_owner           STRING,
  ls_table_name      STRING,
  lo_constraint_diff DYNAMIC ARRAY OF T_CONSTRAINT_DIFF,
  li_rec_cnt         INTEGER,
  ls_sql             STRING
DEFINE
  lb_different BOOLEAN,
  lo_different DYNAMIC ARRAY OF T_CONSTRAINT_DIFF 

  LET ls_owner        = p_owner
  LET ls_table_name   = p_table_name

  LET lb_different = FALSE
  LET lo_different = NULL
  
  &ifdef is_msv
  LET ls_sql = "select ed.dzed001       table_name,                                           ",                                                                     
               "       ed.dzed002       constraint_name,                                      ",                                                                      
               "       ed.dzed003       constraint_type,                                      ",                                                                      
               "       ed.dzed004       column_names,                                         ",                                                                      
               "       ed.dzed005       r_table_name,                                         ",                                                                      
               "       ed.dzed006       r_column_names                                        ",                                                                      
               "  from dzed_t ed                                                              ",                                                                             
               " where ed.dzed001 = '",ls_table_name.toLowerCase(),"'                         ",                                                    
               "   and ed.dzed003 <> 'F'                                                      ",                                                                             
               "except                                                                        ",                                                                              
               "select acs.table_name,                                                        ",                                                                             
               "       acs.constraint_name,                                                   ",  
               "       acs.constraint_type,                                                   ",  
               "       acc.column_names,                                                      ",                                                                             
               "       NULL r_table_name,                                                     ",                                                                  
               "       NULL r_column_names                                                    ",                                                                               
               "  from dba_constraints acs                                                    ",                                                                             
               "  left outer join (                                                           ",                                                                             
               "                   select accs.constraint_name,accs.owner,accs.table_name,    ",  
               "                          string_agg(accs.column_name, ',')                   ",  
               "                          within group(order by accs.position) column_names   ",   
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
  LET ls_sql = "select ed.dzed001       table_name,                                           ",                                                                     
               "       ed.dzed002       constraint_name,                                      ",                                                                      
               "       ed.dzed003       constraint_type,                                      ",                                                                      
               "       ed.dzed004       column_names,                                         ",                                                                      
               "       ed.dzed005       r_table_name,                                         ",                                                                      
               "       ed.dzed006       r_column_names                                        ",                                                                      
               "  from dzed_t ed                                                              ",                                                                             
               " where ed.dzed001 = '",ls_table_name.toLowerCase(),"'                         ",                                                    
               "   and ed.dzed003 <> 'F'                                                      ",                                                                             
               "except                                                                        ",                                                                              
               "select acs.table_name,                                                        ",                                                                             
               "       acs.constraint_name,                                                   ",  
               "       acs.constraint_type,                                                   ",  
               "       acc.column_names,                                                      ",                                                                             
               "       NULL r_table_name,                                                     ",                                                                  
               "       NULL r_column_names                                                    ",                                                                               
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
  LET ls_sql = "SELECT UPPER(ED.DZED001)       TABLE_NAME,                                                                                                                  ",
               "       UPPER(ED.DZED002)       CONSTRAINT_NAME,                                                                                                             ",
               "       UPPER(ED.DZED003)       CONSTRAINT_TYPE,                                                                                                             ",
               "       UPPER(ED.DZED004)       COLUMN_NAMES,                                                                                                                ",
               "       UPPER(TRIM(ED.DZED005)) R_TABLE_NAME,                                                                                                                ",
               "       UPPER(TRIM(ED.DZED006)) R_COLUMN_NAMES                                                                                                               ",
               "  FROM DZED_T ED                                                                                                                                            ",
               " WHERE ED.DZED001 = '",ls_table_name.toLowerCase(),"'                                                                                                       ",
               "   AND ED.DZED003 <> 'F'                                                                                                                                    ",         
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
  
  PREPARE lpre_get_constraint_diff_by_editor FROM ls_sql
  DECLARE lcur_get_constraint_diff_by_editor CURSOR FOR lpre_get_constraint_diff_by_editor

  LET li_rec_cnt = 1 
  
  OPEN lcur_get_constraint_diff_by_editor
  FOREACH lcur_get_constraint_diff_by_editor INTO lo_constraint_diff[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_constraint_diff_by_editor
  IF li_rec_cnt > 1 THEN
    CALL lo_constraint_diff.deleteElement(li_rec_cnt)
  END If  
  
  FREE lpre_get_constraint_diff_by_editor
  FREE lcur_get_constraint_diff_by_editor

  IF li_rec_cnt >= 2 THEN
    LET lb_different = TRUE
  END IF  

  LET lo_different.* = lo_constraint_diff.*
  
  RETURN lb_different, lo_different  

END FUNCTION

#取得 Constraint 資料的比對(DataBase)
FUNCTION sadzi140_shot_get_constraint_diff_by_database(p_owner,p_table_name,p_drop_abnormal_column)
DEFINE
  p_owner        STRING,
  p_table_name   STRING,
  p_drop_abnormal_column BOOLEAN
DEFINE
  ls_owner            STRING,
  ls_table_name       STRING,
  lb_drop_abnormal_column BOOLEAN,
  lo_constraint_diff  DYNAMIC ARRAY OF T_CONSTRAINT_DIFF,
  li_rec_cnt          INTEGER,
  ls_sql              STRING
DEFINE
  lb_different   BOOLEAN,
  lo_different   DYNAMIC ARRAY OF T_CONSTRAINT_DIFF 

  LET ls_owner        = p_owner
  LET ls_table_name   = p_table_name
  LET lb_drop_abnormal_column = p_drop_abnormal_column

  LET lb_different = FALSE
  LET lo_different = NULL

  &ifdef is_msv
  LET ls_sql = "select acs.table_name,                                                        ",                                                                             
               "       acs.constraint_name,                                                   ",  
               "       acs.constraint_type,                                                   ",  
               "       acc.column_names,                                                      ",                                                                             
               "       NULL r_table_name,                                                     ",                                                                  
               "       NULL r_column_names                                                    ",                                                                               
               "  from dba_constraints acs                                                    ",                                                                             
               "  left outer join (                                                           ",                                                                             
               "                   select accs.constraint_name,accs.owner,accs.table_name,    ",  
               "                          string_agg(accs.column_name, ',')                   ",  
               "                          within group(order by accs.position) column_names   ",   
               "                     from dba_cons_columns accs                               ",                                                                             
               "                    group by accs.constraint_name,accs.owner,accs.table_name  ",                                                                             
               "                  ) acc on acc.constraint_name = acs.constraint_name          ",                                                                             
               "                       and acc.table_name      = acs.table_name               ",                                                                             
               "                       and acc.owner           = acs.owner                    ",                                                                             
               " where 1 = 1                                                                  ",                                                                             
               "   and acs.table_name = '",ls_table_name.toLowerCase(),"'                     ",                                                    
               "   and acs.owner      = '",ls_owner.toLowerCase(),"'                          ",
               "except                                                                        ",                                                                              
               "select ed.dzed001       table_name,                                           ",                                                                     
               "       ed.dzed002       constraint_name,                                      ",                                                                      
               "       ed.dzed003       constraint_type,                                      ",                                                                      
               "       ed.dzed004       column_names,                                         ",                                                                      
               "       ed.dzed005       r_table_name,                                         ",                                                                      
               "       ed.dzed006       r_column_names                                        ",                                                                      
               "  from dzed_t ed                                                              ",                                                                             
               " where ed.dzed001 = '",ls_table_name.toLowerCase(),"'                         ",                                                    
               "   and ed.dzed003 <> 'R'                                                      "                                                                             
  &else               
  &ifdef is_pgs
  LET ls_sql = "select acs.table_name,                                                        ",                                                                             
               "       acs.constraint_name,                                                   ",  
               "       acs.constraint_type,                                                   ",  
               "       acc.column_names,                                                      ",                                                                             
               "       NULL r_table_name,                                                     ",                                                                  
               "       NULL r_column_names                                                    ",                                                                               
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
               "   and acs.owner      = '",ls_owner.toLowerCase(),"'                          ",
               "except                                                                        ",                                                                              
               "select ed.dzed001       table_name,                                           ",                                                                     
               "       ed.dzed002       constraint_name,                                      ",                                                                      
               "       ed.dzed003       constraint_type,                                      ",                                                                      
               "       ed.dzed004       column_names,                                         ",                                                                      
               "       ed.dzed005       r_table_name,                                         ",                                                                      
               "       ed.dzed006       r_column_names                                        ",                                                                      
               "  from dzed_t ed                                                              ",                                                                             
               " where ed.dzed001 = '",ls_table_name.toLowerCase(),"'                         ",                                                    
               "   and ed.dzed003 <> 'R'                                                      "                                                                             
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
               "SELECT UPPER(ED.DZED001)       TABLE_NAME,                                                                                                                  ",
               "       UPPER(ED.DZED002)       CONSTRAINT_NAME,                                                                                                             ",
               "       UPPER(ED.DZED003)       CONSTRAINT_TYPE,                                                                                                             ",
               "       UPPER(ED.DZED004)       COLUMN_NAMES,                                                                                                                ",
               "       UPPER(TRIM(ED.DZED005)) R_TABLE_NAME,                                                                                                                ",
               "       UPPER(TRIM(ED.DZED006)) R_COLUMN_NAMES                                                                                                               ",
               "  FROM DZED_T ED                                                                                                                                            ",
               " WHERE ED.DZED001 = '",ls_table_name.toLowerCase(),"'                                                                                                       ",
               "   AND ED.DZED003 <> 'R'                                                                                                                                    "         
  &endif
  &endif
               
  PREPARE lpre_get_constraint_diff_by_database FROM ls_sql
  DECLARE lcur_get_constraint_diff_by_database CURSOR FOR lpre_get_constraint_diff_by_database

  LET li_rec_cnt = 1 
  
  OPEN lcur_get_constraint_diff_by_database
  FOREACH lcur_get_constraint_diff_by_database INTO lo_constraint_diff[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_constraint_diff_by_database
  IF li_rec_cnt > 1 THEN
    CALL lo_constraint_diff.deleteElement(li_rec_cnt)
  END IF  
  
  FREE lpre_get_constraint_diff_by_database
  FREE lcur_get_constraint_diff_by_database

  IF li_rec_cnt >= 2 THEN
    LET lb_different = TRUE
  END IF  

  LET lo_different.* = lo_constraint_diff.*
  
  RETURN lb_different, lo_different  

END FUNCTION

#比對現行 Table Editor 的 Index 資料和最近一次的 Snapshot 的差異(正反向都要抓)
FUNCTION sadzi140_shot_check_index_diff(p_owner,p_table_name,p_drop_abnormal_column)
DEFINE
  p_owner       STRING, 
  p_table_name  STRING,
  p_drop_abnormal_column BOOLEAN
DEFINE
  ls_owner      STRING, 
  ls_table_name STRING,
  lb_drop_abnormal_column BOOLEAN,
  li_rec_cnt    INTEGER,
  ls_sql        STRING
DEFINE
  lo_index_diff  DYNAMIC ARRAY OF T_INDEX_DIFF,
  lb_different   BOOLEAN,
  lo_different   DYNAMIC ARRAY OF T_INDEX_DIFF 

  LET ls_owner      = p_owner
  LET ls_table_name = p_table_name
  LET lb_drop_abnormal_column = p_drop_abnormal_column

  LET lb_different = FALSE
  LET lo_different = NULL
  LET li_rec_cnt = 1 
  
  ##############################################################################
  #正向比對
  &ifdef is_msv
  LET ls_sql = "select ec.dzec001 table_name,                                         ",                                
               "       ec.dzec002 index_name,                                         ",                                
               "       ec.dzec003 index_type,                                         ",                                
               "       replace(ec.dzec004,' ','') column_name                         ",                                
               "  from dzec_t ec                                                      ",                                       
               " where ec.dzec001 = '",ls_table_name.toLowerCase(),"'                 ",              
               "   and coalesce(ec.dzecstus,'Y') = 'Y'                                ",                                            
               "   and not exists (                                                   ",                                       
               "                    select 1                                          ",                                       
               "                      from dzed_t ed                                  ",                                       
               "                     where ed.dzed001 = ec.dzec001                    ",                                       
               "                       and ed.dzed002 = ec.dzec002                    ",                                       
               "                  )                                                   ",                                       
               "except                                                                ",
               "select ais.table_name table_name,                                     ",                                       
               "       ais.index_name index_name,                                     ",                                       
               "       substring(ais.uniqueness,1,1) index_type,                      ",
               "       ais.column_names column_name                                   ",
               "  from dba_indexes ais                                                ",
               " where ais.owner        = '",ls_owner.toLowerCase(),"'                ",                      
               "   and ais.table_name   = '",ls_table_name.toLowerCase(),"'           ",
               "   and ais.index_name not in (                                        ",                                       
               "                               select acs.constraint_name             ",                                       
               "                                 from dba_constraints acs             ",                                       
               "                                where acs.table_name = ais.table_name ",                                       
               "                                  and acs.owner      = ais.owner      ",                                       
               "                             )                                        "                                       
  &else
  &ifdef is_pgs
  LET ls_sql = "select ec.dzec001 table_name,                                         ",                                
               "       ec.dzec002 index_name,                                         ",                                
               "       ec.dzec003 index_type,                                         ",                                
               "       replace(ec.dzec004,' ','') column_name                         ",                                
               "  from dzec_t ec                                                      ",                                       
               " where ec.dzec001 = '",ls_table_name.toLowerCase(),"'                 ",              
               "   and coalesce(ec.dzecstus,'Y') = 'Y'                                ",                                            
               "   and not exists (                                                   ",                                       
               "                    select 1                                          ",                                       
               "                      from dzed_t ed                                  ",                                       
               "                     where ed.dzed001 = ec.dzec001                    ",                                       
               "                       and ed.dzed002 = ec.dzec002                    ",                                       
               "                  )                                                   ",                                       
               "except                                                                ",
               "select ais.table_name table_name,                                     ",                                       
               "       ais.index_name index_name,                                     ",                                       
               "       substring(ais.uniqueness,1,1) index_type,                      ",
               "       string_agg(aic.column_name,','                                 ",
               "                  order by aic.column_position) column_name           ",
               "  from dba_indexes ais                                                ",
               "  inner join dba_ind_columns aic                                      ",
               "          on aic.table_owner = ais.table_owner                        ",
               "         and aic.table_name  = ais.table_name                         ",
               "         and aic.index_name  = ais.index_name                         ",
               " where ais.owner        = '",ls_owner.toLowerCase(),"'                ",                      
               "   and ais.table_name   = '",ls_table_name.toLowerCase(),"'           ",
               "   and ais.index_name not in (                                        ",                                       
               "                               select acs.constraint_name             ",                                       
               "                                 from dba_constraints acs             ",                                       
               "                                where acs.table_name = ais.table_name ",                                       
               "                                  and acs.owner      = ais.owner      ",                                       
               "                             )                                        ",                                       
               " group by ais.table_name,ais.index_name,substring(ais.uniqueness,1,1) "
  &else
  LET ls_sql = "SELECT UPPER(EC.DZEC001) TABLE_NAME,                                                                         ",
               "       UPPER(EC.DZEC002) INDEX_NAME,                                                                         ",
               "       UPPER(EC.DZEC003) INDEX_TYPE,                                                                         ",
               "       UPPER(EC.DZEC004) COLUMN_NAME                                                                         ",
               "  FROM DZEC_T EC                                                                                             ",
               " WHERE EC.DZEC001 = '",ls_table_name.toLowerCase(),"'                                                        ",
               "   AND NVL(EC.DZECSTUS,'Y') = 'Y'                                                                            ", #20170912
               "   AND NOT EXISTS (                                                                                          ",
               "                    SELECT 1                                                                                 ",
               "                      FROM DZED_T ED                                                                         ",
               "                     WHERE ED.DZED001 = EC.DZEC001                                                           ",
               "                       AND ED.DZED002 = EC.DZEC002                                                           ",
               "                  )                                                                                          ",
               " MINUS                                                                                                       ",
               "SELECT AIS.TABLE_NAME                TABLE_NAME,                                                             ",
               "       AIS.INDEX_NAME                INDEX_NAME,                                                             ",
               "       SUBSTRB(AIS.UNIQUENESS,1,1)   INDEX_TYPE,                                                             ",
               "       LISTAGG(AICS.COLUMN_NAME,',') WITHIN GROUP (ORDER BY AIS.INDEX_NAME,AICS.COLUMN_POSITION) COLUMN_NAME ",
               "  FROM DBA_INDEXES AIS,DBA_IND_COLUMNS AICS                                                                  ",
               " WHERE AIS.OWNER        = '",ls_owner.toUpperCase(),"'                                                       ",
               "   AND AIS.TABLE_NAME   = '",ls_table_name.toUpperCase(),"'                                                  ",
               "   AND AIS.INDEX_TYPE <> 'CLUSTER'                                                                           ", --20230807    
               "   AND AICS.TABLE_OWNER = AIS.TABLE_OWNER                                                                    ",
               "   AND AICS.TABLE_NAME  = AIS.TABLE_NAME                                                                     ",
               "   AND AICS.INDEX_NAME  = AIS.INDEX_NAME                                                                     ",
               "   AND AIS.INDEX_NAME NOT IN (                                                                               ",
               "                               SELECT ACS.INDEX_NAME                                                         ", --20230807    
               "                                 FROM DBA_CONSTRAINTS ACS                                                    ",
               "                                WHERE ACS.TABLE_NAME = AIS.TABLE_NAME                                        ",
               "                                  AND ACS.OWNER      = AIS.OWNER                                             ",
               "                             )                                                                               ",
               "  GROUP BY AIS.TABLE_NAME,AIS.INDEX_NAME,SUBSTRB(AIS.UNIQUENESS,1,1)                                         "               
  &endif
  &endif
  
  PREPARE lpre_check_index_diff_forward FROM ls_sql
  DECLARE lcur_check_index_diff_forward CURSOR FOR lpre_check_index_diff_forward

  OPEN lcur_check_index_diff_forward
  FOREACH lcur_check_index_diff_forward INTO lo_index_diff[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_check_index_diff_forward
  {
  IF li_rec_cnt > 1 THEN
    CALL lo_index_diff.deleteElement(li_rec_cnt)
  END IF  
  }
  
  FREE lpre_check_index_diff_forward
  FREE lcur_check_index_diff_forward
  ##############################################################################

  ##############################################################################
  #反向比對
  &ifdef is_msv
  LET ls_sql = 
               "select ais.table_name table_name,                                     ",                                       
               "       ais.index_name index_name,                                     ",                                       
               "       substring(ais.uniqueness,1,1) index_type,                      ",
               "       ais.column_names column_name                                   ",
               "  from dba_indexes ais                                                ",
               " where ais.owner        = '",ls_owner.toLowerCase(),"'                ",                      
               "   and ais.table_name   = '",ls_table_name.toLowerCase(),"'           ",
               "   and ais.index_name not in (                                        ",                                       
               "                               select acs.constraint_name             ",                                       
               "                                 from dba_constraints acs             ",                                       
               "                                where acs.table_name = ais.table_name ",                                       
               "                                  and acs.owner      = ais.owner      ",                                       
               "                             )                                        ",
               "except                                                                ",
               "select ec.dzec001 table_name,                                         ",                                
               "       ec.dzec002 index_name,                                         ",                                
               "       ec.dzec003 index_type,                                         ",                                
               "       replace(ec.dzec004,' ','') column_name                         ",                                
               "  from dzec_t ec                                                      ",                                       
               " where ec.dzec001 = '",ls_table_name.toLowerCase(),"'                 ",              
               "   and coalesce(ec.dzecstus,'Y') = 'Y'                                ",                                            
               "   and not exists (                                                   ",                                       
               "                    select 1                                          ",                                       
               "                      from dzed_t ed                                  ",                                       
               "                     where ed.dzed001 = ec.dzec001                    ",                                       
               "                       and ed.dzed002 = ec.dzec002                    ",                                       
               "                  )                                                   "                                       
  &else               
  &ifdef is_pgs
  LET ls_sql = "select ais.table_name table_name,                                     ",                                       
               "       ais.index_name index_name,                                     ",                                       
               "       substring(ais.uniqueness,1,1) index_type,                      ",
               "       string_agg(aic.column_name,','                                 ",
               "                  order by aic.column_position) column_name           ",
               "  from dba_indexes ais                                                ",
               "  inner join dba_ind_columns aic                                      ",
               "          on aic.table_owner = ais.table_owner                        ",
               "         and aic.table_name  = ais.table_name                         ",
               "         and aic.index_name  = ais.index_name                         ",
               " where ais.owner        = '",ls_owner.toLowerCase(),"'                ",                      
               "   and ais.table_name   = '",ls_table_name.toLowerCase(),"'           ",
               "   and ais.index_name not in (                                        ",                                       
               "                               select acs.constraint_name             ",                                       
               "                                 from dba_constraints acs             ",                                       
               "                                where acs.table_name = ais.table_name ",                                       
               "                                  and acs.owner      = ais.owner      ",                                       
               "                             )                                        ",                                       
               " group by ais.table_name,ais.index_name,substring(ais.uniqueness,1,1) ",
               "except                                                                ",
               "select ec.dzec001 table_name,                                         ",                                
               "       ec.dzec002 index_name,                                         ",                                
               "       ec.dzec003 index_type,                                         ",                                
               "       replace(ec.dzec004,' ','') column_name                         ",                                
               "  from dzec_t ec                                                      ",                                       
               " where ec.dzec001 = '",ls_table_name.toLowerCase(),"'                 ",              
               "   and coalesce(ec.dzecstus,'Y') = 'Y'                                ",                                            
               "   and not exists (                                                   ",                                       
               "                    select 1                                          ",                                       
               "                      from dzed_t ed                                  ",                                       
               "                     where ed.dzed001 = ec.dzec001                    ",                                       
               "                       and ed.dzed002 = ec.dzec002                    ",                                       
               "                  )                                                   "                                       
  &else               
  LET ls_sql = "SELECT AIS.TABLE_NAME                TABLE_NAME,                                                             ",
               "       AIS.INDEX_NAME                INDEX_NAME,                                                             ",
               "       SUBSTRB(AIS.UNIQUENESS,1,1)   INDEX_TYPE,                                                             ",
               "       LISTAGG(AICS.COLUMN_NAME,',') WITHIN GROUP (ORDER BY AIS.INDEX_NAME,AICS.COLUMN_POSITION) COLUMN_NAME ",
               "  FROM DBA_INDEXES AIS,DBA_IND_COLUMNS AICS                                                                  ",
               " WHERE AIS.OWNER        = '",ls_owner.toUpperCase(),"'                                                       ",
               "   AND AIS.TABLE_NAME   = '",ls_table_name.toUpperCase(),"'                                                  ",
               "   AND AIS.INDEX_TYPE <> 'CLUSTER'                                                                           ", --20230807    
               "   AND AICS.TABLE_OWNER = AIS.TABLE_OWNER                                                                    ",
               "   AND AICS.TABLE_NAME  = AIS.TABLE_NAME                                                                     ",
               "   AND AICS.INDEX_NAME  = AIS.INDEX_NAME                                                                     ",
               "   AND AIS.INDEX_NAME NOT IN (                                                                               ",
               "                               SELECT ACS.INDEX_NAME                                                         ", --20230807
               "                                 FROM DBA_CONSTRAINTS ACS                                                    ",
               "                                WHERE ACS.TABLE_NAME = AIS.TABLE_NAME                                        ",
               "                                  AND ACS.OWNER      = AIS.OWNER                                             ",
               "                             )                                                                               ",
               "  GROUP BY AIS.TABLE_NAME,AIS.INDEX_NAME,SUBSTRB(AIS.UNIQUENESS,1,1)                                         ",
               "MINUS                                                                                                        ",
               "SELECT UPPER(EC.DZEC001) TABLE_NAME,                                                                         ",
               "       UPPER(EC.DZEC002) INDEX_NAME,                                                                         ",
               "       UPPER(EC.DZEC003) INDEX_TYPE,                                                                         ",
               "       UPPER(EC.DZEC004) COLUMN_NAME                                                                         ",
               "  FROM DZEC_T EC                                                                                             ",
               " WHERE EC.DZEC001 = '",ls_table_name.toLowerCase(),"'                                                        ",
               "   AND NVL(EC.DZECSTUS,'Y') = 'Y'                                                                            ", #20170912
               "   AND NOT EXISTS (                                                                                          ",
               "                    SELECT 1                                                                                 ",
               "                      FROM DZED_T ED                                                                         ",
               "                     WHERE ED.DZED001 = EC.DZEC001                                                           ",
               "                       AND ED.DZED002 = EC.DZEC002                                                           ",
               "                  )                                                                                          "
  &endif                
  &endif                
  
  PREPARE lpre_check_index_diff_reverse FROM ls_sql
  DECLARE lcur_check_index_diff_reverse CURSOR FOR lpre_check_index_diff_reverse

  OPEN lcur_check_index_diff_reverse
  FOREACH lcur_check_index_diff_reverse INTO lo_index_diff[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_check_index_diff_reverse
  {
  IF li_rec_cnt > 1 THEN
    CALL lo_index_diff.deleteElement(li_rec_cnt)
  END IF  
  }
  FREE lpre_check_index_diff_reverse
  FREE lcur_check_index_diff_reverse
  ##############################################################################
  
  IF li_rec_cnt >= 2 THEN
    LET lb_different = TRUE
  END IF  

  LET lo_different.* = lo_index_diff.*
  
  RETURN lb_different, lo_different
  
END FUNCTION

#取得Index資料的比對(TableEditor)
FUNCTION sadzi140_shot_get_index_diff_by_table_editor(p_owner,p_table_name)
DEFINE
  p_owner        STRING,
  p_table_name   STRING
DEFINE
  ls_owner        STRING,
  ls_table_name   STRING,
  ls_version      STRING,
  lo_index_diff   DYNAMIC ARRAY OF T_INDEX_DIFF,
  li_rec_cnt      INTEGER,
  ls_sql          STRING
DEFINE
  lb_different   BOOLEAN,
  lo_different   DYNAMIC ARRAY OF T_INDEX_DIFF 

  LET ls_owner        = p_owner
  LET ls_table_name   = p_table_name

  LET lb_different = FALSE
  LET lo_different = NULL
  
  &ifdef is_msv
  LET ls_sql = "select ec.dzec001 table_name,                                         ",                                
               "       ec.dzec002 index_name,                                         ",                                
               "       ec.dzec003 index_type,                                         ",                                
               "       replace(ec.dzec004,' ','') column_name                         ",                                
               "  from dzec_t ec                                                      ",                                       
               " where ec.dzec001 = '",ls_table_name.toLowerCase(),"'                 ",              
               "   and coalesce(ec.dzecstus,'Y') = 'Y'                                ",                                            
               "   and not exists (                                                   ",                                       
               "                    select 1                                          ",                                       
               "                      from dzed_t ed                                  ",                                       
               "                     where ed.dzed001 = ec.dzec001                    ",                                       
               "                       and ed.dzed002 = ec.dzec002                    ",                                       
               "                  )                                                   ",                                       
               "except                                                                ",
               "select ais.table_name table_name,                                     ",                                       
               "       ais.index_name index_name,                                     ",                                       
               "       substring(ais.uniqueness,1,1) index_type,                      ",
               "       ais.column_names column_name                                   ",
               "  from dba_indexes ais                                                ",
               " where ais.owner        = '",ls_owner.toLowerCase(),"'                ",                      
               "   and ais.table_name   = '",ls_table_name.toLowerCase(),"'           ",
               "   and ais.index_name not in (                                        ",                                       
               "                               select acs.constraint_name             ",                                       
               "                                 from dba_constraints acs             ",                                       
               "                                where acs.table_name = ais.table_name ",                                       
               "                                  and acs.owner      = ais.owner      ",                                       
               "                             )                                        "                                       
  &else
  &ifdef is_pgs
  LET ls_sql = "select ec.dzec001 table_name,                                         ",                                
               "       ec.dzec002 index_name,                                         ",                                
               "       ec.dzec003 index_type,                                         ",                                
               "       replace(ec.dzec004,' ','') column_name                         ",                                
               "  from dzec_t ec                                                      ",                                       
               " where ec.dzec001 = '",ls_table_name.toLowerCase(),"'                 ",              
               "   and coalesce(ec.dzecstus,'Y') = 'Y'                                ",                                            
               "   and not exists (                                                   ",                                       
               "                    select 1                                          ",                                       
               "                      from dzed_t ed                                  ",                                       
               "                     where ed.dzed001 = ec.dzec001                    ",                                       
               "                       and ed.dzed002 = ec.dzec002                    ",                                       
               "                  )                                                   ",                                       
               "except                                                                ",
               "select ais.table_name table_name,                                     ",                                       
               "       ais.index_name index_name,                                     ",                                       
               "       substring(ais.uniqueness,1,1) index_type,                      ",
               "       string_agg(aic.column_name,','                                 ",
               "                  order by aic.column_position) column_name           ",
               "  from dba_indexes ais                                                ",
               "  inner join dba_ind_columns aic                                      ",
               "          on aic.table_owner = ais.table_owner                        ",
               "         and aic.table_name  = ais.table_name                         ",
               "         and aic.index_name  = ais.index_name                         ",
               " where ais.owner        = '",ls_owner.toLowerCase(),"'                ",                      
               "   and ais.table_name   = '",ls_table_name.toLowerCase(),"'           ",
               "   and ais.index_name not in (                                        ",                                       
               "                               select acs.constraint_name             ",                                       
               "                                 from dba_constraints acs             ",                                       
               "                                where acs.table_name = ais.table_name ",                                       
               "                                  and acs.owner      = ais.owner      ",                                       
               "                             )                                        ",
               "         group by ais.table_name,ais.index_name,                      ",
               "                  substring(ais.uniqueness,1,1)                       "                                       
  &else
  LET ls_sql = "SELECT UPPER(EC.DZEC001) TABLE_NAME,                                                                         ",
               "       UPPER(EC.DZEC002) INDEX_NAME,                                                                         ",
               "       UPPER(EC.DZEC003) INDEX_TYPE,                                                                         ",
               "       UPPER(EC.DZEC004) COLUMN_NAME                                                                         ",
               "  FROM DZEC_T EC                                                                                             ",
               " WHERE EC.DZEC001 = '",ls_table_name.toLowerCase(),"'                                                        ",
               "   AND NVL(EC.DZECSTUS,'Y') = 'Y'                                                                            ", #20170912
               "   AND NOT EXISTS (                                                                                          ",
               "                    SELECT 1                                                                                 ",
               "                      FROM DZED_T ED                                                                         ",
               "                     WHERE ED.DZED001 = EC.DZEC001                                                           ",
               "                       AND ED.DZED002 = EC.DZEC002                                                           ",
               "                  )                                                                                          ",
               " MINUS                                                                                                       ",
               "SELECT AIS.TABLE_NAME,AIS.INDEX_NAME,SUBSTRB(AIS.UNIQUENESS,1,1) INDEX_TYPE,                                 ",
               "       LISTAGG(AICS.COLUMN_NAME,',') WITHIN GROUP (ORDER BY AIS.INDEX_NAME,AICS.COLUMN_POSITION) COLUMN_NAME ",
               "                                                                                                             ",
               "  FROM DBA_INDEXES AIS,DBA_IND_COLUMNS AICS                                                                  ",
               " WHERE AIS.OWNER        = '",ls_owner.toUpperCase(),"'                                                       ",
               "   AND AIS.TABLE_NAME   = '",ls_table_name.toUpperCase(),"'                                                  ",
               "   AND AIS.INDEX_TYPE <> 'CLUSTER'                                                                           ", --20230807    
               "   AND AICS.TABLE_OWNER = AIS.TABLE_OWNER                                                                    ",
               "   AND AICS.TABLE_NAME  = AIS.TABLE_NAME                                                                     ",
               "   AND AICS.INDEX_NAME  = AIS.INDEX_NAME                                                                     ",
               "   AND AIS.INDEX_NAME NOT IN (                                                                               ",
               "                               SELECT ACS.INDEX_NAME                                                         ", --20230807
               "                                 FROM DBA_CONSTRAINTS ACS                                                    ",
               "                                WHERE ACS.TABLE_NAME = AIS.TABLE_NAME                                        ",
               "                                  AND ACS.OWNER      = AIS.OWNER                                             ",
               "                             )                                                                               ",
               "  GROUP BY AIS.TABLE_NAME,AIS.INDEX_NAME,SUBSTRB(AIS.UNIQUENESS,1,1)                                         "               
  &endif 
  &endif 
  
  PREPARE lpre_get_index_diff_by_editor FROM ls_sql
  DECLARE lcur_get_index_diff_by_editor CURSOR FOR lpre_get_index_diff_by_editor

  LET li_rec_cnt = 1 
  
  OPEN lcur_get_index_diff_by_editor
  FOREACH lcur_get_index_diff_by_editor INTO lo_index_diff[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_index_diff_by_editor
  IF li_rec_cnt > 1 THEN
    CALL lo_index_diff.deleteElement(li_rec_cnt)
  END If  
  
  FREE lpre_get_index_diff_by_editor
  FREE lcur_get_index_diff_by_editor

  IF li_rec_cnt >= 2 THEN
    LET lb_different = TRUE
  END IF  

  LET lo_different.* = lo_index_diff.*
  
  RETURN lb_different, lo_different  

END FUNCTION

#取得Index 資料的比對(Snapshot)
FUNCTION sadzi140_shot_get_index_diff_by_database(p_owner,p_table_name,p_drop_abnormal_column)
DEFINE
  p_owner        STRING,
  p_table_name   STRING,
  p_drop_abnormal_column BOOLEAN
DEFINE
  ls_owner        STRING,
  ls_table_name   STRING,
  lb_drop_abnormal_column BOOLEAN,
  lo_index_diff   DYNAMIC ARRAY OF T_INDEX_DIFF,
  li_rec_cnt      INTEGER,
  ls_sql          STRING
DEFINE
  lb_different   BOOLEAN,
  lo_different   DYNAMIC ARRAY OF T_INDEX_DIFF 

  LET ls_owner        = p_owner
  LET ls_table_name   = p_table_name
  LET lb_drop_abnormal_column = p_drop_abnormal_column

  LET lb_different = FALSE
  LET lo_different = NULL

  &ifdef is_msv
  LET ls_sql = "select ais.table_name table_name,                                     ",                                       
               "       ais.index_name index_name,                                     ",                                       
               "       substring(ais.uniqueness,1,1) index_type,                      ",
               "       ais.column_names column_name                                   ",
               "  from dba_indexes ais                                                ",
               " where ais.owner        = '",ls_owner.toLowerCase(),"'                ",                      
               "   and ais.table_name   = '",ls_table_name.toLowerCase(),"'           ",
               "   and ais.index_name not in (                                        ",                                       
               "                               select acs.constraint_name             ",                                       
               "                                 from dba_constraints acs             ",                                       
               "                                where acs.table_name = ais.table_name ",                                       
               "                                  and acs.owner      = ais.owner      ",                                       
               "                             )                                        ",
               "except                                                                ",
               "select ec.dzec001 table_name,                                         ",                                
               "       ec.dzec002 index_name,                                         ",                                
               "       ec.dzec003 index_type,                                         ",                                
               "       replace(ec.dzec004,' ','') column_name                         ",                                
               "  from dzec_t ec                                                      ",                                       
               " where ec.dzec001 = '",ls_table_name.toLowerCase(),"'                 ",              
               "   and coalesce(ec.dzecstus,'Y') = 'Y'                                ",                                            
               "   and not exists (                                                   ",                                       
               "                    select 1                                          ",                                       
               "                      from dzed_t ed                                  ",                                       
               "                     where ed.dzed001 = ec.dzec001                    ",                                       
               "                       and ed.dzed002 = ec.dzec002                    ",                                       
               "                  )                                                   "                                       
  &else               
  &ifdef is_pgs
  LET ls_sql = "select ais.table_name table_name,                                     ",                                       
               "       ais.index_name index_name,                                     ",                                       
               "       substring(ais.uniqueness,1,1) index_type,                      ",
               "       string_agg(aic.column_name,','                                 ",
               "                  order by aic.column_position) column_name           ",
               "  from dba_indexes ais                                                ",
               "  inner join dba_ind_columns aic                                      ",
               "          on aic.table_owner = ais.table_owner                        ",
               "         and aic.table_name  = ais.table_name                         ",
               "         and aic.index_name  = ais.index_name                         ",
               " where ais.owner        = '",ls_owner.toLowerCase(),"'                ",                      
               "   and ais.table_name   = '",ls_table_name.toLowerCase(),"'           ",
               "   and ais.index_name not in (                                        ",                                       
               "                               select acs.constraint_name             ",                                       
               "                                 from dba_constraints acs             ",                                       
               "                                where acs.table_name = ais.table_name ",                                       
               "                                  and acs.owner      = ais.owner      ",                                       
               "                             )                                        ",
               "         group by ais.table_name,ais.index_name,                      ",
               "                  substring(ais.uniqueness,1,1)                       ",
               "except                                                                ",
               "select ec.dzec001 table_name,                                         ",                                
               "       ec.dzec002 index_name,                                         ",                                
               "       ec.dzec003 index_type,                                         ",                                
               "       replace(ec.dzec004,' ','') column_name                         ",                                
               "  from dzec_t ec                                                      ",                                       
               " where ec.dzec001 = '",ls_table_name.toLowerCase(),"'                 ",              
               "   and coalesce(ec.dzecstus,'Y') = 'Y'                                ",                                            
               "   and not exists (                                                   ",                                       
               "                    select 1                                          ",                                       
               "                      from dzed_t ed                                  ",                                       
               "                     where ed.dzed001 = ec.dzec001                    ",                                       
               "                       and ed.dzed002 = ec.dzec002                    ",                                       
               "                  )                                                   "                                       
  &else               
  LET ls_sql = "SELECT AIS.TABLE_NAME,AIS.INDEX_NAME,SUBSTRB(AIS.UNIQUENESS,1,1) INDEX_TYPE,                                 ",
               "       LISTAGG(AICS.COLUMN_NAME,',') WITHIN GROUP (ORDER BY AIS.INDEX_NAME,AICS.COLUMN_POSITION) COLUMN_NAME ",
               "  FROM DBA_INDEXES AIS,DBA_IND_COLUMNS AICS                                                                  ",
               " WHERE AIS.OWNER        = '",ls_owner.toUpperCase(),"'                                                       ",
               "   AND AIS.TABLE_NAME   = '",ls_table_name.toUpperCase(),"'                                                  ",
               "   AND AIS.INDEX_TYPE <> 'CLUSTER'                                                                           ", --20230807    
               "   AND AICS.TABLE_OWNER = AIS.TABLE_OWNER                                                                    ",
               "   AND AICS.TABLE_NAME  = AIS.TABLE_NAME                                                                     ",
               "   AND AICS.INDEX_NAME  = AIS.INDEX_NAME                                                                     ",
               "   AND AIS.INDEX_NAME NOT IN (                                                                               ",
               "                               SELECT ACS.INDEX_NAME                                                         ", --20230807
               "                                 FROM DBA_CONSTRAINTS ACS                                                    ",
               "                                WHERE ACS.TABLE_NAME = AIS.TABLE_NAME                                        ",
               "                                  AND ACS.OWNER      = AIS.OWNER                                             ",
               "                             )                                                                               ",
               "  GROUP BY AIS.TABLE_NAME,AIS.INDEX_NAME,SUBSTRB(AIS.UNIQUENESS,1,1)                                         ",
               "MINUS                                                                                                        ",
               "SELECT UPPER(EC.DZEC001) TABLE_NAME,                                                                         ",
               "       UPPER(EC.DZEC002) INDEX_NAME,                                                                         ",
               "       UPPER(EC.DZEC003) INDEX_TYPE,                                                                         ",
               "       UPPER(EC.DZEC004) COLUMN_NAME                                                                         ",
               "  FROM DZEC_T EC                                                                                             ",
               " WHERE EC.DZEC001 = '",ls_table_name.toLowerCase(),"'                                                        ",
               "   AND NVL(EC.DZECSTUS,'Y') = 'Y'                                                                            ", #20170912
               "   AND NOT EXISTS (                                                                                          ",
               "                    SELECT 1                                                                                 ",
               "                      FROM DZED_T ED                                                                         ",
               "                     WHERE ED.DZED001 = EC.DZEC001                                                           ",
               "                       AND ED.DZED002 = EC.DZEC002                                                           ",
               "                  )                                                                                          "
  &endif
  &endif
  
  PREPARE lpre_get_index_diff_by_database FROM ls_sql
  DECLARE lcur_get_index_diff_by_database CURSOR FOR lpre_get_index_diff_by_database

  LET li_rec_cnt = 1 
  
  OPEN lcur_get_index_diff_by_database
  FOREACH lcur_get_index_diff_by_database INTO lo_index_diff[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_index_diff_by_database
  IF li_rec_cnt > 1 THEN
    CALL lo_index_diff.deleteElement(li_rec_cnt)
  END IF  
  
  FREE lpre_get_index_diff_by_database
  FREE lcur_get_index_diff_by_database

  IF li_rec_cnt >= 2 THEN
    LET lb_different = TRUE
  END IF  

  LET lo_different.* = lo_index_diff.*
  
  RETURN lb_different, lo_different  

END FUNCTION

FUNCTION sadzi140_shot_get_key_list(p_master_user,p_table_name)
DEFINE
  p_master_user  STRING,
  p_table_name   STRING
DEFINE
  ls_master_user   STRING,
  ls_table_name    STRING,
  lo_key_field     DYNAMIC ARRAY OF T_COLUMN_DIFF
DEFINE  
  ls_sql         STRING,
  li_rec_cnt     INTEGER,
  ls_key_field   STRING

  LET ls_master_user  = p_master_user
  LET ls_table_name   = p_table_name

  LET ls_key_field = ""

  &ifdef is_msv
  LET ls_sql = "select eb.dzeb002                      dzeb002,             ", 
               "       eb.dzeb004                      dzeb004,             ", 
               "       iif(eb.dzeb004='Y','N','Y')     dzeb005,             ", 
               "       td.gztd003                      dzeb007,             ", 
               "       case td.gztd003                                      ", 
               "         when 'varbinary'                                   ", 
               "           then 'max'                                       ", 
               "         when 'nvarchar'                                    ", 
               "           then iif(td.gztd009='TEXT','max',td.gztd008)     ", 
               "       else                                                 ", 
               "         td.gztd008                                         ", 
               "       end                              dzeb008             ", 
               "  from dzeb_t eb                                            ",  
               "       left outer join gztd_t td on td.gztd001 = eb.dzeb006 ",  
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"'       ",  
               " order by eb.dzeb021                                        "  
  &else
  &ifdef is_pgs
  LET ls_sql = "select eb.dzeb002                dzeb002,             ", 
               "       eb.dzeb004                dzeb004,             ", 
               "       case when eb.dzeb004='Y'                       ",
               "       then 'N' else 'Y' end     dzeb005,             ", 
               "       td.gztd003                dzeb007,             ", 
               "       td.gztd008                dzeb008              ", 
               "  from dzeb_t eb                                      ",  
               "  left outer join gztd_t td                           ",
               "               on td.gztd001 = eb.dzeb006             ",  
               " where eb.dzeb001 = '",ls_table_name.toLowerCase(),"' ",  
               " order by eb.dzeb021                                  "  
  &else
  LET ls_sql = "SELECT UPPER(EB.DZEB002)                      DZEB002,       ",
               "       EB.DZEB004                             DZEB004,       ",  
               "       DECODE(EB.DZEB004,'Y','N','N','Y','Y') DZEB005,       ",
               "       UPPER(TD.GZTD003)                      DZEB007,       ",
               "       REPLACE(TO_CHAR(TO_NUMBER(REPLACE(REPLACE(            ",
               "       DECODE(                                               ",
               "               UPPER(TD.GZTD003),                            ",
               "               'BLOB','4000',                                ",
               "               'CLOB','4000',                                ",
               "               'DATE','7',                                   ",
               "               'TIMESTAMP',DECODE(TD.GZTD008,'0','7','11'),  ",
               "               TD.GZTD008                                    ",
               "             ),',','.'),' ',''))),'.',',')    DZEB008        ",
               "  FROM DZEB_T EB                                             ",
               "       LEFT OUTER JOIN GZTD_T TD ON TD.GZTD001 = EB.DZEB006  ",
               " WHERE EB.DZEB001 = '",ls_table_name.toLowerCase(),"'        ",
               " ORDER BY EB.DZEB021                                         "
  &endif               
  &endif               
               
  PREPARE lpre_get_key_list FROM ls_sql
  DECLARE lcur_get_key_list CURSOR FOR lpre_get_key_list

  LET li_rec_cnt = 1 
  
  OPEN lcur_get_key_list
  FOREACH lcur_get_key_list INTO lo_key_field[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    IF lo_key_field[li_rec_cnt].DZEV006 = "Y" THEN
      LET ls_key_field = ls_key_field,",",lo_key_field[li_rec_cnt].DZEV005
    END IF    
    
    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_key_list
  
  FREE lpre_get_key_list
  FREE lcur_get_key_list  

  RETURN ls_key_field
  
END FUNCTION

FUNCTION sadzi140_shot_delete_table_shot_data(p_master_db,p_table_name,p_version_info)
DEFINE
  p_master_db    STRING,
  p_table_name   STRING,
  p_version_info T_VERSION_INFO
DEFINE
  ls_master_db     STRING,
  ls_table_name    STRING,
  lo_version_info  T_VERSION_INFO,
  lb_success       BOOLEAN,
  ls_delete_sql    STRING
DEFINE
  lb_return BOOLEAN
  
  &ifdef is_msv
  LET ls_master_db  = p_master_db.toLowerCase() 
  LET ls_table_name = p_table_name.toLowerCase()
  &else
  &ifdef is_pgs
  LET ls_master_db  = p_master_db.toLowerCase() 
  LET ls_table_name = p_table_name.toLowerCase()
  &else
  LET ls_master_db  = p_master_db.toUpperCase() 
  LET ls_table_name = p_table_name.toUpperCase()
  &endif               
  &endif               
  
  LET lo_version_info.* = p_version_info.*

  LET lb_success = TRUE

  LET ls_delete_sql = "delete from dzev_t                                 ",
                      " where dzev001 = '",ls_master_db,"'                ",
                      "   and dzev002 = '",ls_table_name,"'               ",
                      --"   and dzev003 = '",lo_version_info.REVISION,"'    ",
                      "   and dzev018 = '",lo_version_info.ALM_VERSION,"' "

  TRY
    PREPARE lpre_delete_table_shot_data FROM ls_delete_sql
    EXECUTE lpre_delete_table_shot_data
  CATCH
    LET lb_success = FALSE
  END TRY

  LET lb_return = lb_success

  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_shot_delete_constraint_index_shot_data(p_table_name,p_version_info)
DEFINE
  p_table_name   STRING,
  p_version_info T_VERSION_INFO
DEFINE
  ls_table_name   STRING,
  lo_version_info T_VERSION_INFO,
  lb_success      BOOLEAN,
  ls_sql          STRING
DEFINE
  lb_return BOOLEAN  
  
  LET ls_table_name     = p_table_name
  LET lo_version_info.* = p_version_info.*

  LET lb_success = TRUE

  LET ls_sql = "delete from dzew_t                                 ",
               " where dzew001 = '",ls_table_name.toLowerCase(),"' ",
               "   and dzew002 = '",lo_version_info.REVISION,"'    ",
               "   and dzew010 = '",lo_version_info.ALM_VERSION,"' "

  TRY
    PREPARE lpre_delete_constraint_index_shot_data FROM ls_sql
    EXECUTE lpre_delete_constraint_index_shot_data
  CATCH
    LET lb_success = FALSE
  END TRY

  LET lb_return = lb_success

  RETURN lb_return
  
END FUNCTION

FUNCTION sadzi140_shot_get_constraint_index_data(p_table_name,p_new_version,p_curr_version,p_alm_version,p_alm_request_no,p_dgenv)
DEFINE
  p_table_name     STRING,
  p_new_version    STRING,
  p_curr_version   STRING,
  p_alm_version    STRING,
  p_alm_request_no STRING,
  p_dgenv          STRING 
DEFINE
  ls_table_name       STRING,
  ls_new_version      STRING,
  ls_curr_version     STRING,
  ls_alm_version      STRING,
  ls_alm_request_no   STRING, 
  ls_dgenv            STRING, 
  lo_constraint_index DYNAMIC ARRAY OF T_CONSTRAINT_INDEX
DEFINE  
  ls_sql     STRING,
  li_rec_cnt INTEGER,
  lo_return  DYNAMIC ARRAY OF T_CONSTRAINT_INDEX

  LET ls_table_name     = p_table_name
  LET ls_new_version    = p_new_version  
  LET ls_curr_version   = p_curr_version
  LET ls_alm_version    = p_alm_version  
  LET ls_alm_request_no = p_alm_request_no
  LET ls_dgenv          = p_dgenv

  INITIALIZE lo_constraint_index TO NULL
  INITIALIZE lo_return TO NULL
  
  LET ls_sql = "select ec.dzec001                          table_name,           ",             
               "       ''                                  seq_no,               ",             
               "       'Index'                             constraint_or_index,  ",             
               "       ec.dzec002                          data_name,            ",             
               "       ec.dzec003                          datatypes,            ",             
               "       translate(ec.dzec004,'+-._',',,,,') data_columns,         ",             
               "       ''                                  foreign_table,        ",             
               "       ''                                  foreign_columns,      ",             
               "       ''                                  prev_seq_no,          ",             
               "       ''                                  alm_version,          ",             
               "       ''                                  alm_request_no,       ",             
               "       ec.dzec006                          dgenv,                ",             
               "       ec.dzec007                          orig_dgenv,           ",             
               "       ec.dzec008                          ship_notice,          ",             
               "       ec.dzecstus                         status_code           ", #20170912   
               "  from dzec_t ec                                                 ",             
               " where ec.dzec001 = '",ls_table_name,"'                          ",             
               "union all                                                        ",             
               "select ed.dzed001                          table_name,           ",             
               "       ''                                  seq_no,               ",             
               "       'Constraint'                        constraint_or_index,  ",             
               "       ed.dzed002                          data_name,            ",             
               "       ed.dzed003                          datatypes,            ",             
               "       translate(ed.dzed004,'+-._',',,,,') data_columns,         ",             
               "       ed.dzed005                          foreign_table,        ",             
               "       ed.dzed006                          foreign_columns,      ",             
               "       ''                                  prev_seq_no,          ",             
               "       ''                                  alm_version,          ",             
               "       ''                                  alm_request_no,       ",             
               "       ed.dzed008                          dgenv,                ",             
               "       ed.dzed009                          orig_dgenv,           ",             
               "       ed.dzed010                          ship_notice,          ",             
               "       ed.dzedstus                         status_code           ", #20170912   
               "  from dzed_t ed                                                ",             
               " where ed.dzed001 = '",ls_table_name,"'                         ",             
               "   and ed.dzed003 <> 'F'                                        "              
               
  PREPARE lpre_GetConstraintIndexData FROM ls_sql
  DECLARE lcur_GetConstraintIndexData CURSOR FOR lpre_GetConstraintIndexData

  LET li_rec_cnt = 1 
  
  OPEN lcur_GetConstraintIndexData
  FOREACH lcur_GetConstraintIndexData INTO lo_constraint_index[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET lo_constraint_index[li_rec_cnt].DZEW002 = ls_new_version 
    LET lo_constraint_index[li_rec_cnt].DZEW009 = ls_curr_version
    LET lo_constraint_index[li_rec_cnt].DZEW010 = ls_alm_version
    LET lo_constraint_index[li_rec_cnt].DZEW011 = ls_alm_request_no
    
    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_GetConstraintIndexData
  IF li_rec_cnt > 1 THEN
    CALL lo_constraint_index.deleteElement(li_rec_cnt)
  END If  
  
  FREE lpre_GetConstraintIndexData
  FREE lcur_GetConstraintIndexData  

  LET lo_return = lo_constraint_index 
  
  RETURN lo_return
  
END FUNCTION

FUNCTION sadzi140_shot_set_constraint_index_data(p_constraint_index)
DEFINE
  p_constraint_index DYNAMIC ARRAY OF T_CONSTRAINT_INDEX
DEFINE
  li_loop INTEGER,
  lo_constraint_index DYNAMIC ARRAY OF T_CONSTRAINT_INDEX

  LET lo_constraint_index.* = p_constraint_index.*

  BEGIN WORK 
  FOR li_loop = 1 TO lo_constraint_index.getLength()
    CALL sadzi140_shot_insert_DZEW_T(lo_constraint_index[li_loop].*)
  END FOR  
  COMMIT WORK
  
END FUNCTION

FUNCTION sadzi140_shot_insert_DZEW_T(p_constraint_index)
DEFINE
  p_constraint_index T_CONSTRAINT_INDEX
DEFINE
  li_loop INTEGER,
  lo_constraint_index T_CONSTRAINT_INDEX

  LET lo_constraint_index.* = p_constraint_index.*

  TRY
    &ifdef is_pgs
    SAVEPOINT sp_tran
    &endif
    insert into dzew_t(
                        dzew001,dzew002,dzew003,dzew004,dzew005,
                        dzew006,dzew007,dzew008,dzew009,dzew010,
                        dzew011,dzew012,dzew013,dzew014,dzewstus #20170912 add dzewstus
                      ) 
               values (
                        lo_constraint_index.DZEW001,lo_constraint_index.DZEW002,lo_constraint_index.DZEW003,lo_constraint_index.DZEW004,lo_constraint_index.DZEW005,
                        lo_constraint_index.DZEW006,TRIM(lo_constraint_index.DZEW007),TRIM(lo_constraint_index.DZEW008),lo_constraint_index.DZEW009,lo_constraint_index.DZEW010,
                        lo_constraint_index.DZEW011,lo_constraint_index.DZEW012,lo_constraint_index.DZEW013,lo_constraint_index.DZEW014,lo_constraint_index.dzewstus #20170912 add dzewstus
                      )
  CATCH 
    &ifdef is_pgs
    ROLLBACK WORK TO SAVEPOINT sp_tran
    &endif
    #DISPLAY cs_warning_tag,"Insert into DZEW_T : ",SQLCA.SQLCODE
  END TRY
  
END FUNCTION

FUNCTION sadzi140_shot_get_max_index_prev_version(p_table_name,p_curr_version)
DEFINE 
  p_table_name   STRING,
  p_curr_version STRING
DEFINE
  ls_table_name   STRING,
  ls_curr_version STRING,
  ls_sql          STRING,
  li_rec_cnt      INTEGER,
  ls_prev_version VARCHAR(30)         
DEFINE  
  ls_return STRING

  LET ls_table_name   = p_table_name
  LET ls_curr_version = p_curr_version
  
  LET li_rec_cnt = 1
  
  LET ls_sql = "select max(ew.dzew002) max_prev_version   ",
               "  from dzew_t ew                          ",
               " where 1=1                                ",
               "   and ew.dzew001 = '",ls_table_name,"'   ",
               "   and ew.dzew003 = 'Index'               ",
               "   and ew.dzew002 <> '",ls_curr_version,"'"
               
  PREPARE lpre_get_max_index_prev_version FROM ls_sql
  DECLARE lcur_get_max_index_prev_version CURSOR FOR lpre_get_max_index_prev_version

  OPEN lcur_get_max_index_prev_version
  FETCH lcur_get_max_index_prev_version INTO ls_prev_version
  CLOSE lcur_get_max_index_prev_version
  
  FREE lpre_get_max_index_prev_version
  FREE lcur_get_max_index_prev_version  

  LET ls_return = NVL(ls_prev_version,cs_init_version_code)
  
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_shot_get_max_construct_revision(p_master_db,p_table_name,p_construct_version)
DEFINE 
  p_master_db         STRING,
  p_table_name        STRING,
  p_construct_version STRING
DEFINE
  ls_master_db         STRING,
  ls_table_name        STRING,
  ls_construct_version STRING,
  ls_sql               STRING,
  li_rec_cnt           INTEGER,
  ls_max_version       VARCHAR(30)         
DEFINE  
  ls_return STRING

  LET ls_master_db  = p_master_db
  LET ls_table_name = p_table_name
  LET ls_construct_version = p_construct_version
  
  LET li_rec_cnt = 1

  &ifdef is_msv
  LET ls_sql = "select coalesce(cast(max(cast(coalesce(ev.dzev003,'",cs_init_version_code,"') as int)) as varchar),'0') max_version ",
               "  from dzev_t ev                                                                                                    ",
               " where ev.dzev001 = '",ls_master_db.toLowerCase(),"'                                                                ",
               "   and ev.dzev002 = '",ls_table_name.toLowerCase(),"'                                                               ",
               "   and ev.dzev004 = 'TableDesigner'                                                                                 ",
               "   and ev.dzev018 = '",ls_construct_version,"'                                                                      " 
  &else
  &ifdef is_pgs
  LET ls_sql = "select coalesce(cast(max(cast(coalesce(ev.dzev003,'",cs_init_version_code,"') as int)) as varchar),'0') max_version ",
               "  from dzev_t ev                                                                                                    ",
               " where ev.dzev001 = '",ls_master_db.toLowerCase(),"'                                                                ",
               "   and ev.dzev002 = '",ls_table_name.toLowerCase(),"'                                                               ",
               "   and ev.dzev004 = 'TableDesigner'                                                                                 ",
               "   and ev.dzev018 = '",ls_construct_version,"'                                                                      " 
  &else
  LET ls_sql = "SELECT TO_CHAR(MAX(TO_NUMBER(NVL(EV.DZEV003,'",cs_init_version_code,"')))) MAX_VERSION ",
               "  FROM DZEV_T EV                                                                       ",
               " WHERE EV.DZEV001 = '",ls_master_db.toUpperCase(),"'                                   ",
               "   AND EV.DZEV002 = '",ls_table_name.toUpperCase(),"'                                  ",
               "   AND EV.DZEV004 = 'TableDesigner'                                                    ",
               "   AND EV.DZEV018 = '",ls_construct_version,"'                                         "
  &endif
  &endif
  
  PREPARE lpre_get_max_construct_revision FROM ls_sql
  DECLARE lcur_get_max_construct_revision CURSOR FOR lpre_get_max_construct_revision

  OPEN lcur_get_max_construct_revision
  FETCH lcur_get_max_construct_revision INTO ls_max_version
  CLOSE lcur_get_max_construct_revision
  
  FREE lpre_get_max_construct_revision
  FREE lcur_get_max_construct_revision  

  LET ls_return = ls_max_version
  
  RETURN ls_return
  
END FUNCTION

FUNCTION sadzi140_shot_get_privileges_diff_by_table_editor(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_table_name     STRING,
  ls_sql            STRING,
  li_rec_cnt        INTEGER,
  lo_privilege_diff DYNAMIC ARRAY OF T_PRIVILEGE_DIFF
DEFINE
  lb_return BOOLEAN,
  lo_return DYNAMIC ARRAY OF T_PRIVILEGE_DIFF 

  LET ls_table_name = p_table_name
  
  LET li_rec_cnt = 1

  &ifdef is_msv
  #SQL Server 無權限議題
  LET ls_sql = "select en.dzen002,en.dzen003,en.dzen004,en.dzen005,en.dzen006,  ",
               "       en.dzen007,en.dzen008,en.dzen009,en.dzen010              ", 
               "  from dzen_t en                                                ",
               " where 1=2                                                      "
  &else
  &ifdef is_pgs
  LET ls_sql = "select en.dzen002,en.dzen003,en.dzen004,en.dzen005,en.dzen006,                       ",
               "       en.dzen007,en.dzen008,en.dzen009,en.dzen010                                   ",
               "  from dzen_t en                                                                     ",
               " where en.dzen001 = '",ls_table_name.toLowerCase(),"'                                ",
               "except                                                                               ",
               "select lower(dtp.owner)                                                     dzen002, ", 
               "       lower(dtp.grantee)                                                   dzen003, ", 
               "       coalesce(max(case when dtp.privilege='SELECT' then 'Y' end),'N')     dzen004, ", 
               "       coalesce(max(case when dtp.privilege='INSERT' then 'Y' end),'N')     dzen005, ", 
               "       coalesce(max(case when dtp.privilege='UPDATE' then 'Y' end),'N')     dzen006, ", 
               "       coalesce(max(case when dtp.privilege='DELETE' then 'Y' end),'N')     dzen007, ", 
               "       coalesce(max(case when dtp.privilege='REFERENCES' then 'Y' end),'N') dzen008, ", 
               "       coalesce(max(case when dtp.privilege='ALTER' then 'Y' end),'N')      dzen009, ", 
               "       coalesce(max(case when dtp.privilege='INDEX' then 'Y' end),'N')      dzen010  ", 
               "  from dba_tab_privs dtp                                                             ",
               " where dtp.table_name = '",ls_table_name.toLowerCase(),"'                            ",
               "   and dtp.owner <> dtp.grantee                                                      ", #20230901
               "   and dtp.grantee <> '",cs_master_user,"'                                           ", #20230901 
               "   and exists (                                                                      ",
               "                select 1                                                             ",
               "                  from gzda_t da                                                     ",
               "                 where da.gzda001 = lower(dtp.owner)                                 ",
               "              )                                                                      ",
               "   and exists (                                                                      ",
               "                select 1                                                             ",
               "                  from gzda_t da                                                     ",
               "                 where da.gzda001 = lower(dtp.grantee)                               ",
               "              )                                                                      ",
               " group by dtp.owner,dtp.grantee                                                      " 
  &else
  LET ls_sql = "SELECT EN.DZEN002,EN.DZEN003,EN.DZEN004,EN.DZEN005,EN.DZEN006,        ",
               "       EN.DZEN007,EN.DZEN008,EN.DZEN009,EN.DZEN010                    ", 
               "  FROM DZEN_T EN                                                      ",
               " WHERE EN.DZEN001 = '",ls_table_name.toLowerCase(),"'                 ",
               "MINUS                                                                 ",
               "SELECT LOWER(DTP.OWNER)                                     DZEN002,  ",
               "       LOWER(DTP.GRANTEE)                                   DZEN003,  ",
               "       NVL(MAX(DECODE(DTP.PRIVILEGE,'SELECT','Y')),'N')     DZEN004,  ",
               "       NVL(MAX(DECODE(DTP.PRIVILEGE,'INSERT','Y')),'N')     DZEN005,  ",
               "       NVL(MAX(DECODE(DTP.PRIVILEGE,'UPDATE','Y')),'N')     DZEN006,  ",
               "       NVL(MAX(DECODE(DTP.PRIVILEGE,'DELETE','Y')),'N')     DZEN007,  ",
               "       NVL(MAX(DECODE(DTP.PRIVILEGE,'REFERENCES','Y')),'N') DZEN008,  ",
               "       NVL(MAX(DECODE(DTP.PRIVILEGE,'ALTER','Y')),'N')      DZEN009,  ",
               "       NVL(MAX(DECODE(DTP.PRIVILEGE,'INDEX','Y')),'N')      DZEN010   ",
               "  FROM DBA_TAB_PRIVS DTP                                              ",
               " WHERE DTP.TABLE_NAME = '",ls_table_name.toUpperCase(),"'             ",
               "   AND EXISTS (                                                       ", # 20160812 begin
               "                SELECT 1                                              ",
               "                  FROM GZDA_T DA                                      ",
               "                 WHERE DA.GZDA001 = LOWER(DTP.OWNER)                  ",
               "              )                                                       ",
               "   AND EXISTS (                                                       ",
               "                SELECT 1                                              ",
               "                  FROM GZDA_T DA                                      ",
               "                 WHERE DA.GZDA001 = LOWER(DTP.GRANTEE)                ",
               "              )                                                       ", # 20160812 end
               " GROUP BY DTP.OWNER,DTP.GRANTEE                                       "
  &endif
  &endif
  
  PREPARE lpre_get_privileges_diff_by_table_editor FROM ls_sql
  DECLARE lcur_get_privileges_diff_by_table_editor CURSOR FOR lpre_get_privileges_diff_by_table_editor

  OPEN lcur_get_privileges_diff_by_table_editor
  FOREACH lcur_get_privileges_diff_by_table_editor INTO lo_privilege_diff[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_privileges_diff_by_table_editor

  CALL lo_privilege_diff.deleteElement(li_rec_cnt)
  
  FREE lpre_get_privileges_diff_by_table_editor
  FREE lcur_get_privileges_diff_by_table_editor  
  
  LET lb_return = IIF(lo_privilege_diff.getLength() >= 1,TRUE,FALSE)
  LET lo_return = lo_privilege_diff 
  
  RETURN lb_return,lo_return
  
END FUNCTION 

FUNCTION sadzi140_shot_get_privileges_diff_by_database(p_table_name)
DEFINE
  p_table_name STRING
DEFINE
  ls_table_name     STRING,
  ls_sql            STRING,
  li_rec_cnt        INTEGER,
  lo_privilege_diff DYNAMIC ARRAY OF T_PRIVILEGE_DIFF
DEFINE
  lb_return BOOLEAN,
  lo_return DYNAMIC ARRAY OF T_PRIVILEGE_DIFF 

  LET ls_table_name = p_table_name
  
  LET li_rec_cnt = 1
  
  &ifdef is_msv
  #SQL Server 無權限議題
  LET ls_sql = "select en.dzen002,en.dzen003,en.dzen004,en.dzen005,en.dzen006,  ",
               "       en.dzen007,en.dzen008,en.dzen009,en.dzen010              ", 
               "  from dzen_t en                                                ",
               " where 1=2                                                      "
  &else
  &ifdef is_pgs
  LET ls_sql = "select lower(dtp.owner)                                                     dzen002, ", 
               "       lower(dtp.grantee)                                                   dzen003, ", 
               "       coalesce(max(case when dtp.privilege='SELECT' then 'Y' end),'N')     dzen004, ", 
               "       coalesce(max(case when dtp.privilege='INSERT' then 'Y' end),'N')     dzen005, ", 
               "       coalesce(max(case when dtp.privilege='UPDATE' then 'Y' end),'N')     dzen006, ", 
               "       coalesce(max(case when dtp.privilege='DELETE' then 'Y' end),'N')     dzen007, ", 
               "       coalesce(max(case when dtp.privilege='REFERENCES' then 'Y' end),'N') dzen008, ", 
               "       coalesce(max(case when dtp.privilege='ALTER' then 'Y' end),'N')      dzen009, ", 
               "       coalesce(max(case when dtp.privilege='INDEX' then 'Y' end),'N')      dzen010  ", 
               "  from dba_tab_privs dtp                                                             ",
               " where dtp.table_name = '",ls_table_name.toLowerCase(),"'                            ",
               "   and dtp.owner <> dtp.grantee                                                      ", #20230901
               "   and dtp.grantee <> '",cs_master_user,"'                                           ", #20230901 
               "   and exists (                                                                      ",
               "                select 1                                                             ",
               "                  from gzda_t da                                                     ",
               "                 where da.gzda001 = lower(dtp.owner)                                 ",
               "              )                                                                      ",
               "   and exists (                                                                      ",
               "                select 1                                                             ",
               "                  from gzda_t da                                                     ",
               "                 where da.gzda001 = lower(dtp.grantee)                               ",
               "              )                                                                      ",
               " group by dtp.owner,dtp.grantee                                                      ", 
               "except                                                                               ",
               "select en.dzen002,en.dzen003,en.dzen004,en.dzen005,en.dzen006,                       ",
               "       en.dzen007,en.dzen008,en.dzen009,en.dzen010                                   ",
               "  from dzen_t en                                                                     ",
               " where en.dzen001 = '",ls_table_name.toLowerCase(),"'                                "
  &else
  LET ls_sql = "SELECT LOWER(DTP.OWNER)                                     DZEN002,  ",
               "       LOWER(DTP.GRANTEE)                                   DZEN003,  ",
               "       NVL(MAX(DECODE(DTP.PRIVILEGE,'SELECT','Y')),'N')     DZEN004,  ",
               "       NVL(MAX(DECODE(DTP.PRIVILEGE,'INSERT','Y')),'N')     DZEN005,  ",
               "       NVL(MAX(DECODE(DTP.PRIVILEGE,'UPDATE','Y')),'N')     DZEN006,  ",
               "       NVL(MAX(DECODE(DTP.PRIVILEGE,'DELETE','Y')),'N')     DZEN007,  ",
               "       NVL(MAX(DECODE(DTP.PRIVILEGE,'REFERENCES','Y')),'N') DZEN008,  ",
               "       NVL(MAX(DECODE(DTP.PRIVILEGE,'ALTER','Y')),'N')      DZEN009,  ",
               "       NVL(MAX(DECODE(DTP.PRIVILEGE,'INDEX','Y')),'N')      DZEN010   ",
               "  FROM DBA_TAB_PRIVS DTP                                              ",
               " WHERE DTP.TABLE_NAME = '",ls_table_name.toUpperCase(),"'             ",
               "   AND EXISTS (                                                       ", # 20160812 begin
               "                SELECT 1                                              ",
               "                  FROM GZDA_T DA                                      ",
               "                 WHERE DA.GZDA001 = LOWER(DTP.OWNER)                  ",
               "              )                                                       ",
               "   AND EXISTS (                                                       ",
               "                SELECT 1                                              ",
               "                  FROM GZDA_T DA                                      ",
               "                 WHERE DA.GZDA001 = LOWER(DTP.GRANTEE)                ",
               "              )                                                       ", # 20160812 end
               " GROUP BY DTP.OWNER,DTP.GRANTEE                                       ",
               "MINUS                                                                 ",
               "SELECT EN.DZEN002,EN.DZEN003,EN.DZEN004,EN.DZEN005,EN.DZEN006,        ",
               "       EN.DZEN007,EN.DZEN008,EN.DZEN009,EN.DZEN010                    ", 
               "  FROM DZEN_T EN                                                      ",
               " WHERE EN.DZEN001 = '",ls_table_name.toLowerCase(),"'                 "
  &endif
  &endif
  
  PREPARE lpre_get_privileges_diff_by_database FROM ls_sql
  DECLARE lcur_get_privileges_diff_by_database CURSOR FOR lpre_get_privileges_diff_by_database

  OPEN lcur_get_privileges_diff_by_database
  FOREACH lcur_get_privileges_diff_by_database INTO lo_privilege_diff[li_rec_cnt].*
    IF SQLCA.sqlcode THEN
      EXIT FOREACH
    END IF

    LET li_rec_cnt = li_rec_cnt + 1

  END FOREACH
  CLOSE lcur_get_privileges_diff_by_database

  CALL lo_privilege_diff.deleteElement(li_rec_cnt)
  
  FREE lpre_get_privileges_diff_by_database
  FREE lcur_get_privileges_diff_by_database  
  
  LET lb_return = IIF(lo_privilege_diff.getLength() >= 1,TRUE,FALSE)
  LET lo_return = lo_privilege_diff 
  
  RETURN lb_return,lo_return
  
END FUNCTION 

FUNCTION sadzi140_shot_check_privileges_if_diff(p_table_name)
DEFINE 
  p_table_name STRING
DEFINE 
  ls_table_name            STRING,
  lo_privilege_design_diff DYNAMIC ARRAY OF T_PRIVILEGE_DIFF,
  lo_privilege_db_diff     DYNAMIC ARRAY OF T_PRIVILEGE_DIFF,
  lb_result                BOOLEAN,
  lb_diff                  BOOLEAN
DEFINE
  lb_return BOOLEAN  

  LET ls_table_name = p_table_name

  LET lb_diff = FALSE 

  CALL sadzi140_shot_get_privileges_diff_by_table_editor(p_table_name) RETURNING lb_result,lo_privilege_design_diff
  IF lb_result THEN LET lb_diff = TRUE END IF

  CALL sadzi140_shot_get_privileges_diff_by_database(p_table_name) RETURNING lb_result,lo_privilege_db_diff
  IF lb_result THEN LET lb_diff = TRUE END IF

  LET lb_return = lb_diff
  
  RETURN lb_return
  
END FUNCTION 
