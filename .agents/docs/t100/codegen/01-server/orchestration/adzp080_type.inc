
SCHEMA ds

PUBLIC TYPE T_UPLOAD_PARAM RECORD
              arg1_program_name      VARCHAR(50),                    
              arg2_module_name       VARCHAR(10),               
              arg3_upload_doc_type   VARCHAR(10),
              arg4_client_path       VARCHAR(1024),                
              arg5_spec_version      VARCHAR(20),                
              arg6_code_version      VARCHAR(20),
              arg7_std_program       VARCHAR(50),                    
              arg8_std_spec_version  VARCHAR(20),                
              arg9_std_code_version  VARCHAR(20),
              arg10_dgenv            VARCHAR(5),
              arg11_type             VARCHAR(5)
            END RECORD

PUBLIC TYPE T_DZHI_T RECORD
              DZHI001   LIKE DZHI_T.DZHI001,
              DZHI002   LIKE DZHI_T.DZHI002,
              DZHI003   LIKE DZHI_T.DZHI003,
              DZHI004   LIKE DZHI_T.DZHI004,
              DZHI005   LIKE DZHI_T.DZHI005,
              DZHI006   LIKE DZHI_T.DZHI006,
              DZHI007   LIKE DZHI_T.DZHI007,
              DZHI008   LIKE DZHI_T.DZHI008,
              DZHI009   LIKE DZHI_T.DZHI009,
              DZHI010   LIKE DZHI_T.DZHI010,
              DZHISTUS  LIKE DZHI_T.DZHISTUS,
              DZHICRTID LIKE DZHI_T.DZHICRTID,
              DZHICRTDT LIKE DZHI_T.DZHICRTDT,
              DZHIMODID LIKE DZHI_T.DZHIMODID,
              DZHIMODDT LIKE DZHI_T.DZHIMODDT
            END RECORD
