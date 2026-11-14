PUBLIC TYPE T_PROGRAM_INFO RECORD
              pi_NAME    VARCHAR(30),   -- 代號
              pi_MODULE  VARCHAR(10),   -- 模組
              pi_DESC    VARCHAR(1024), -- 名稱
              pi_TYPE    VARCHAR(20)    -- 類型 : TAB:表格, PRG:程式, COM:元件, RPT:報表, SVC:服務
            END RECORD

PUBLIC TYPE T_USER_INFO RECORD
              ui_NUMBER   VARCHAR(30), -- 工號
              ui_NAME     VARCHAR(30), -- 姓名
              ui_ROLE     VARCHAR(10), -- 角色(SD, PR)
              ui_LANG     VARCHAR(10)  -- 語言別
            END RECORD                         

PUBLIC TYPE T_COLLECT_MESSAGE RECORD
              cm_MSG_NO    VARCHAR(20),
              cm_MSG_TEXT  VARCHAR(500)
            END RECORD                         

PRIVATE TYPE T_LANGUAGE_TYPE VARCHAR(10)                
