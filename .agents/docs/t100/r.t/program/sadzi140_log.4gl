
PUBLIC TYPE T_DBOBJ_DEFINE RECORD
              p_version             STRING,
              p_exec_type           STRING,
              p_obj_type            STRING,
              p_obj_name            STRING,
              p_subexec_type        STRING,
              p_subobj_type         STRING,
              p_subobj_name         STRING,
              p_data_type           STRING,
              p_data_length_or_list STRING,
              p_notnull             STRING, 
              p_comment             STRING,
              p_req_no              STRING
            END RECORD

FUNCTION sadzi140_log_set_log(p_dbobj_define)
DEFINE 
  p_dbobj_define T_DBOBJ_DEFINE
  
  CASE p_dbobj_define.p_exec_type.toUpperCase()
    WHEN "CREATE"
      CASE p_dbobj_define.p_obj_type.toUpperCase()
        WHEN "TABLE"
          CALL sadzi140_log_set_create_table_log(p_dbobj_define.*)
        WHEN "INDEX"
          CALL sadzi140_log_set_create_index_log(p_dbobj_define.*)
      OTHERWISE  
      END CASE
    WHEN "ALTER"
      CASE p_dbobj_define.p_subexec_type.toUpperCase()
        WHEN "ADD"
          CASE p_dbobj_define.p_subobj_type.toUpperCase()
            WHEN "COLUMN"
              CALL sadzi140_log_set_alter_add_column_log(p_dbobj_define.*)
            WHEN "CONSTRAINT"
              CALL sadzi140_log_set_alter_add_constraint_log(p_dbobj_define.*)
            WHEN "USER"
            WHEN "ROLE"
          OTHERWISE
          END CASE
        WHEN "MODIFY"
          CASE p_dbobj_define.p_subobj_type.toUpperCase()
            WHEN "COLUMN"
              CALL sadzi140_log_set_alter_modify_column_log(p_dbobj_define.*)
            WHEN "INDEX"
            WHEN "USER"
            WHEN "ROLE"
          OTHERWISE
          END CASE
        WHEN "DROP"
          CASE p_dbobj_define.p_subobj_type.toUpperCase()
            WHEN "COLUMN"
              CALL sadzi140_log_set_alter_drop_column_log(p_dbobj_define.*)
            WHEN "CONSTRAINT"
              CALL sadzi140_log_set_alter_drop_constraint_log(p_dbobj_define.*)
            WHEN "INDEX"
            WHEN "USER"
            WHEN "ROLE"
          OTHERWISE
          END CASE
      OTHERWISE
      END CASE
    WHEN "DROP"
      CASE p_dbobj_define.p_obj_type.toUpperCase()
        WHEN "TABLE"
          CALL sadzi140_log_set_drop_table_log(p_dbobj_define.*)
        WHEN "INDEX"
          CALL sadzi140_log_set_drop_index_log(p_dbobj_define.*)
      OTHERWISE    
      END CASE    
    WHEN "GRANT"
    WHEN "REVOKE"
    WHEN "COMMENT"
      CASE p_dbobj_define.p_obj_type.toUpperCase()
        WHEN "TABLE"
          CALL sadzi140_log_set_comment_table_log(p_dbobj_define.*)
        WHEN "COLUMN"
          CALL sadzi140_log_set_comment_column_log(p_dbobj_define.*)
      OTHERWISE
      END CASE
  OTHERWISE
  END CASE
  
END FUNCTION

FUNCTION sadzi140_log_set_comment_table_log(p_dbobj_define)
DEFINE 
  p_dbobj_define T_DBOBJ_DEFINE
DEFINE
  ls_insert_sql STRING,
  ls_sequence   STRING
  
  LET ls_insert_sql = ""
  LET ls_sequence  = "-1"

  CALL sadzi140_log_get_sequence() RETURNING ls_sequence
  
  LET ls_insert_sql = "INSERT INTO DZES_T (dzes001,dzes002,dzes003,dzes004,dzes011,dzes012,dzes013,dzes014)                         ", 
                      "            VALUES ('",p_dbobj_define.p_version,"','COMMENT','TABLE','",p_dbobj_define.p_obj_name,"','",p_dbobj_define.p_comment,"','",p_dbobj_define.p_req_no,"',",ls_sequence,",'0')" 
 
  TRY
    PREPARE lpre_set_comment_table_log FROM ls_insert_sql
    EXECUTE lpre_set_comment_table_log
  CATCH
    DISPLAY "[Error] Comment on table log insert into DZES not correct."
  END TRY
  
END FUNCTION

FUNCTION sadzi140_log_set_comment_column_log(p_dbobj_define)
DEFINE 
  p_dbobj_define T_DBOBJ_DEFINE
DEFINE
  ls_insert_sql STRING,
  ls_sequence  STRING
  
  LET ls_insert_sql = ""
  LET ls_sequence  = "-1"

  CALL sadzi140_log_get_sequence() RETURNING ls_sequence
  
  LET ls_insert_sql = "INSERT INTO DZES_T (dzes001,dzes002,dzes003,dzes004,dzes011,dzes012,dzes013,dzes014)                         ", 
                     "            VALUES ('",p_dbobj_define.p_version,"','COMMENT','COLUMN','",p_dbobj_define.p_obj_name,"','",p_dbobj_define.p_comment,"','",p_dbobj_define.p_req_no,"',",ls_sequence,",'0')" 
 
  TRY
    PREPARE lpre_set_comment_column_log FROM ls_insert_sql
    EXECUTE lpre_set_comment_column_log
  CATCH
    DISPLAY "[Error] Comment on table log insert into DZES not correct."
  END TRY
  
END FUNCTION

FUNCTION sadzi140_log_set_create_table_log(p_dbobj_define)
DEFINE 
  p_dbobj_define T_DBOBJ_DEFINE
DEFINE
  ls_insert_sql STRING,
  ls_sequence  STRING
  
  LET ls_insert_sql = ""
  LET ls_sequence  = "-1"

  CALL sadzi140_log_get_sequence() RETURNING ls_sequence
  
  LET ls_insert_sql = "INSERT INTO DZES_T (dzes001,dzes002,dzes003,dzes004,dzes011,dzes012,dzes013,dzes014)                         ", 
                     "            VALUES ('",p_dbobj_define.p_version,"','CREATE','TABLE','",p_dbobj_define.p_obj_name,"','",p_dbobj_define.p_comment,"','",p_dbobj_define.p_req_no,"',",ls_sequence,",'0')" 
 
  TRY
    PREPARE lpre_set_create_table_log FROM ls_insert_sql
    EXECUTE lpre_set_create_table_log
  CATCH
    DISPLAY "[Error] Create table log insert into DZES not correct."
  END TRY

END FUNCTION

FUNCTION sadzi140_log_set_create_index_log(p_dbobj_define)
DEFINE 
  p_dbobj_define T_DBOBJ_DEFINE
DEFINE
  ls_insert_sql STRING,
  ls_sequence  STRING
  
  LET ls_insert_sql = ""
  LET ls_sequence  = "-1"

  CALL sadzi140_log_get_sequence() RETURNING ls_sequence
  
  LET ls_insert_sql = "INSERT INTO DZES_T (dzes001,dzes002,dzes003,dzes004,dzes007,dzes008,dzes009,dzes012,dzes013,dzes014) ", 
                     "            VALUES ('",p_dbobj_define.p_version,"','CREATE','INDEX','",p_dbobj_define.p_obj_name,"','",p_dbobj_define.p_subobj_name,"','",p_dbobj_define.p_data_type,"','",p_dbobj_define.p_data_length_or_list,"','",p_dbobj_define.p_req_no,"',",ls_sequence,",'0')" 
 
  TRY
    PREPARE lpre_set_create_index_log FROM ls_insert_sql
    EXECUTE lpre_set_create_index_log
  CATCH
    DISPLAY "[Error] Create index log insert into DZES not correct."
  END TRY
  
END FUNCTION

FUNCTION sadzi140_log_set_drop_table_log(p_dbobj_define)
DEFINE 
  p_dbobj_define T_DBOBJ_DEFINE
DEFINE
  ls_insert_sql STRING,
  ls_sequence  STRING
  
  LET ls_insert_sql = ""
  LET ls_sequence  = "-1"

  CALL sadzi140_log_get_sequence() RETURNING ls_sequence
  
  LET ls_insert_sql = "INSERT INTO DZES_T (dzes001,dzes002,dzes003,dzes004,dzes012,dzes013,dzes014)                       ", 
                     "            VALUES ('",p_dbobj_define.p_version,"','DROP','TABLE','",p_dbobj_define.p_obj_name,"','",p_dbobj_define.p_req_no,"',",ls_sequence,",'0')" 
 
  TRY
    PREPARE lpre_set_drop_table_log FROM ls_insert_sql
    EXECUTE lpre_set_drop_table_log
  CATCH
    DISPLAY "[Error] Drop table log insert into DZES not correct."
  END TRY

END FUNCTION

FUNCTION sadzi140_log_set_drop_index_log(p_dbobj_define)
DEFINE 
  p_dbobj_define T_DBOBJ_DEFINE
DEFINE
  ls_insert_sql STRING,
  ls_sequence  STRING
  
  LET ls_insert_sql = ""
  LET ls_sequence  = "-1"

  CALL sadzi140_log_get_sequence() RETURNING ls_sequence
  
  LET ls_insert_sql = "INSERT INTO DZES_T (dzes001,dzes002,dzes003,dzes004,dzes012,dzes013,dzes014)                       ", 
                     "            VALUES ('",p_dbobj_define.p_version,"','DROP','INDEX','",p_dbobj_define.p_obj_name,"','",p_dbobj_define.p_req_no,"',",ls_sequence,",'0')" 
 
  TRY
    PREPARE lpre_set_drop_index_log FROM ls_insert_sql
    EXECUTE lpre_set_drop_index_log
  CATCH
    DISPLAY "[Error] Drop table log insert into DZES not correct."
  END TRY

END FUNCTION

FUNCTION sadzi140_log_set_alter_add_column_log(p_dbobj_define)
DEFINE 
  p_dbobj_define T_DBOBJ_DEFINE
DEFINE
  ls_insert_sql STRING,
  ls_sequence  STRING,
  ls_notnull   STRING
  
  LET ls_insert_sql = ""
  LET ls_sequence  = "-1"

  CALL sadzi140_log_get_sequence() RETURNING ls_sequence

  IF NVL(p_dbobj_define.p_notnull,"N") = "Y" THEN
    LET ls_notnull = "NOT NULL"
  ELSE
    LET ls_notnull = "NULL"
  END IF 
  
  LET ls_insert_sql = "INSERT INTO DZES_T (dzes001,dzes002,dzes003,dzes004,dzes005,dzes006,dzes007,dzes008,dzes009,dzes010,dzes011,dzes012,dzes013,dzes014) ", 
                     "            VALUES ('",p_dbobj_define.p_version,"','ALTER','TABLE','",p_dbobj_define.p_obj_name,"','ADD','COLUMN','",p_dbobj_define.p_subobj_name,"','",p_dbobj_define.p_data_type,"','",p_dbobj_define.p_data_length_or_list,"','",ls_notnull,"','",p_dbobj_define.p_comment,"','",p_dbobj_define.p_req_no,"',",ls_sequence,",'0')" 
 
  TRY
    PREPARE lpre_set_alter_add_column_log FROM ls_insert_sql
    EXECUTE lpre_set_alter_add_column_log
  CATCH
    DISPLAY "[Error] Alter table add column log insert into DZES not correct."
  END TRY
  
END FUNCTION

FUNCTION sadzi140_log_set_alter_add_constraint_log(p_dbobj_define)
DEFINE 
  p_dbobj_define T_DBOBJ_DEFINE
DEFINE
  ls_insert_sql STRING,
  ls_sequence  STRING
  
  LET ls_insert_sql = ""
  LET ls_sequence  = "-1"

  CALL sadzi140_log_get_sequence() RETURNING ls_sequence
  
  LET ls_insert_sql = "INSERT INTO DZES_T (dzes001,dzes002,dzes003,dzes004,dzes005,dzes006,dzes007,dzes008,dzes009,dzes012,dzes013,dzes014) ", 
                     "            VALUES ('",p_dbobj_define.p_version,"','ALTER','TABLE','",p_dbobj_define.p_obj_name,"','ADD','CONSTRAINT','",p_dbobj_define.p_subobj_name,"','",p_dbobj_define.p_data_type,"','",p_dbobj_define.p_data_length_or_list,"','",p_dbobj_define.p_req_no,"',",ls_sequence,",'0')" 
 
  TRY
    PREPARE lpre_set_alter_add_constraint_log FROM ls_insert_sql
    EXECUTE lpre_set_alter_add_constraint_log
  CATCH
    DISPLAY "[Error] Alter table add constraint log insert into DZES not correct."
  END TRY
  
END FUNCTION

FUNCTION sadzi140_log_set_alter_drop_constraint_log(p_dbobj_define)
DEFINE 
  p_dbobj_define T_DBOBJ_DEFINE
DEFINE
  ls_insert_sql STRING,
  ls_sequence  STRING
  
  LET ls_insert_sql = ""
  LET ls_sequence  = "-1"

  CALL sadzi140_log_get_sequence() RETURNING ls_sequence
  
  LET ls_insert_sql = "INSERT INTO DZES_T (dzes001,dzes002,dzes003,dzes004,dzes005,dzes006,dzes007,dzes008,dzes009,dzes012,dzes013,dzes014) ", 
                     "            VALUES ('",p_dbobj_define.p_version,"','ALTER','TABLE','",p_dbobj_define.p_obj_name,"','DROP','CONSTRAINT','",p_dbobj_define.p_subobj_name,"','",p_dbobj_define.p_data_type,"','",p_dbobj_define.p_data_length_or_list,"','",p_dbobj_define.p_req_no,"',",ls_sequence,",'0')" 
 
  TRY
    PREPARE lpre_set_alter_drop_constraint_log FROM ls_insert_sql
    EXECUTE lpre_set_alter_drop_constraint_log
  CATCH
    DISPLAY "[Error] Alter table drop constraint log insert into DZES not correct."
  END TRY
  
END FUNCTION

FUNCTION sadzi140_log_set_alter_modify_column_log(p_dbobj_define)
DEFINE 
  p_dbobj_define T_DBOBJ_DEFINE
DEFINE
  ls_insert_sql STRING,
  ls_sequence  STRING,
  ls_notnull   STRING
  
  LET ls_insert_sql = ""
  LET ls_sequence  = "-1"

  IF NVL(p_dbobj_define.p_notnull,"N") = "Y" THEN
    LET ls_notnull = "NOT NULL"
  ELSE
    LET ls_notnull = "NULL"
  END IF 

  CALL sadzi140_log_get_sequence() RETURNING ls_sequence
  
  LET ls_insert_sql = "INSERT INTO DZES_T (dzes001,dzes002,dzes003,dzes004,dzes005,dzes006,dzes007,dzes008,dzes009,dzes010,dzes011,dzes012,dzes013,dzes014) ", 
                     "            VALUES ('",p_dbobj_define.p_version,"','ALTER','TABLE','",p_dbobj_define.p_obj_name,"','MODIFY','COLUMN','",p_dbobj_define.p_subobj_name,"','",p_dbobj_define.p_data_type,"','",p_dbobj_define.p_data_length_or_list,"','",ls_notnull,"','",p_dbobj_define.p_comment,"','",p_dbobj_define.p_req_no,"',",ls_sequence,",'0')" 
 
  TRY
    PREPARE lpre_set_alter_modify_column_log FROM ls_insert_sql
    EXECUTE lpre_set_alter_modify_column_log
  CATCH
    DISPLAY "[Error] Alter table modify column log insert into DZES not correct."
  END TRY
  
END FUNCTION

FUNCTION sadzi140_log_set_alter_drop_column_log(p_dbobj_define)
DEFINE 
  p_dbobj_define T_DBOBJ_DEFINE
DEFINE
  ls_insert_sql STRING,
  ls_sequence  STRING
  
  LET ls_insert_sql = ""
  LET ls_sequence  = "-1"

  CALL sadzi140_log_get_sequence() RETURNING ls_sequence
  
  LET ls_insert_sql = "INSERT INTO DZES_T (dzes001,dzes002,dzes003,dzes004,dzes005,dzes006,dzes007,dzes012,dzes013,dzes014) ", 
                     "            VALUES ('",p_dbobj_define.p_version,"','ALTER','TABLE','",p_dbobj_define.p_obj_name,"','DROP','COLUMN','",p_dbobj_define.p_subobj_name,"','",p_dbobj_define.p_req_no,"',",ls_sequence,",'0')" 
 
  TRY
    PREPARE lpre_set_alter_drop_column_log FROM ls_insert_sql
    EXECUTE lpre_set_alter_drop_column_log
  CATCH
    DISPLAY "[Error] Alter table drop column log insert into DZES not correct."
  END TRY
  
END FUNCTION

FUNCTION sadzi140_log_get_sequence()
DEFINE
  ls_log_seq STRING
DEFINE
  ls_return STRING
  
  CALL sadzi140_vcs_get_new_seqence("LOG_SEQ",FALSE,FALSE,FALSE,TRUE,4) RETURNING ls_log_seq

  LET ls_return = ls_log_seq

  RETURN ls_return
 
END FUNCTION




