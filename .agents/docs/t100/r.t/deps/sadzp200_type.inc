
PUBLIC TYPE T_DZAF_T RECORD
              DZAF001	VARCHAR(30), --建構代號
              DZAF002	INTEGER,     --建構版號
              DZAF003	INTEGER,     --規格版號
              DZAF004	INTEGER,     --代碼版號
              DZAF005	VARCHAR(20), --建構類型
              DZAF006	VARCHAR(20), --模組    
              DZAF007	VARCHAR(20), --產品代號
              DZAF008	VARCHAR(20), --產品版本
              DZAF009	VARCHAR(50), --客戶代碼              
              DZAF010	VARCHAR(5)   --識別標示
            END RECORD            

PUBLIC TYPE T_DZLU_T RECORD
               DZLU001   VARCHAR(10),  -- 角色    
               DZLU002   VARCHAR(20),  -- 工號    
               DZLU003   VARCHAR(30),  -- 需求單號
               DZLU004   VARCHAR(20),  -- 產品代號
               DZLU005   VARCHAR(20),  -- 產品版本
               DZLU006   INTEGER,      -- 作業項次
               DZLU007   VARCHAR(10),  -- 客戶代碼   
               DZLU008   VARCHAR(20),  -- 作業類型
               DZLU009   VARCHAR(300)  -- 作業代號
            END RECORD            
            
PUBLIC TYPE T_DZLM_T RECORD
              DZLM001	VARCHAR(10),  -- 建構類型
              DZLM002	VARCHAR(200),  -- 建構代號
              DZLM003	VARCHAR(500), -- 建構名稱
              DZLM004	VARCHAR(10),  -- 模組    
              DZLM005	INTEGER,       -- 建構版號
              DZLM006	INTEGER,       -- SD版號  
              DZLM007	VARCHAR(20),  -- SD工號  
              DZLM008	VARCHAR(1),   -- SD狀態  
              DZLM009	INTEGER,       -- PR版號  
              DZLM010	VARCHAR(20),  -- PR工號  
              DZLM011	VARCHAR(1),   -- PR狀態  
              DZLM012	VARCHAR(20),  -- 需求單號
              DZLM013	VARCHAR(10),  -- 產品代號
              DZLM014	VARCHAR(10),  -- 產品版本
              DZLM015	INTEGER,       -- 作業項次
              DZLM016	VARCHAR(40),  -- 客戶代碼    
              DZLM017	DATETIME YEAR TO SECOND, -- 簽入時間
              DZLM018 VARCHAR(40),  -- SD GUID
              DZLM019 VARCHAR(40),  -- PR GUID
              DZLM020 VARCHAR(1),   -- SD 已下載
              DZLM021 VARCHAR(1)    -- PR 已下載
            END RECORD            

PRIVATE TYPE T_DZAN_T RECORD
               DZAN001	VARCHAR(128), --建構代號
               DZAN002	VARCHAR(10), --建構版號
               DZAN003	VARCHAR(15), --規格版號
               DZAN004	VARCHAR(15), --代碼版號
               DZAN005	VARCHAR(10), --建構類型
               DZAN006	VARCHAR(10), --模組    
               DZAN007	VARCHAR(10), --產品代號
               DZAN008	VARCHAR(10), --產品版本
               DZAN009	VARCHAR(40), --客戶代碼
               DZAN010	VARCHAR(1),  --識別標示
               DZAN011	VARCHAR(40), --GUID
               DZAN012	VARCHAR(10), --角色
               DZAN013	VARCHAR(10), --儲存類別
               DZAN014	DATETIME YEAR TO SECOND, --建立日期
               DZAN015  VARCHAR(1)   -- 處理碼
             END RECORD             

PRIVATE TYPE T_STATIC_ROLE_LIST DYNAMIC ARRAY OF VARCHAR(2)      

PRIVATE TYPE T_CHECK_OUT_OWNER_LIST RECORD
               cool_ROLE        VARCHAR(10), 
               cool_ID          VARCHAR(128),
               cool_NAME        VARCHAR(50),
               cool_REQUEST_NO  VARCHAR(50),
               cool_SEQUENCE_NO VARCHAR(50)  
             END RECORD

PRIVATE TYPE T_REQUEST_INFORMATION RECORD
               ri_REQUEST_NO  VARCHAR(50),
               ri_SEQUENCE_NO VARCHAR(50)
             END RECORD                             

PRIVATE TYPE T_MARK_INFORMATION RECORD
               mi_result    BOOLEAN,
               mi_code      STRING,
               mi_line_no   STRING,
               mi_line_code STRING 
             END RECORD             

PRIVATE TYPE T_CHECK_OUT_INFO RECORD
               coi_VERSION  INTEGER,
               coi_OWNER    VARCHAR(50),
               coi_IO       VARCHAR(2) 
             END RECORD             