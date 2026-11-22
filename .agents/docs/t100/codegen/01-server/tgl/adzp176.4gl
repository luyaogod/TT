# code_version.: 5.25.02-2014.01.20(0000-0001)

#20191209 01856  追加 cl_ap_exitprogram() #191209-00016#2

IMPORT os
SCHEMA ds
    
#+ code_id......: adzp176.4gl
#+ descriptions.: section parser
#+ code_creator.: ka0132
 
GLOBALS "../../cfg/top_global.inc"

DEFINE gs_prog_name STRING
DEFINE gs_mod_name  STRING
DEFINE gs_prog_ver  STRING
DEFINE gi_prog_ver  LIKE type_t.num10   #210511-00011 代追單
DEFINE gs_dgenv     STRING
DEFINE gs_wmode     STRING
DEFINE gc_dzax003   LIKE dzax_t.dzax003 #是否為free style
DEFINE gb_section   BOOLEAN

MAIN
   DEFINE li_cnt       INTEGER
   DEFINE lr_dzbc      RECORD LIKE dzbc_t.*
   
   CALL cl_tool_init()
   
   #顯示外部參數訊息
   #DISPLAY "ID:",ARG_VAL(1)
   DISPLAY "程式模組:",ARG_VAL(2)
   DISPLAY "程式名稱:",ARG_VAL(3)
   DISPLAY "程式版次:",ARG_VAL(4)
   DISPLAY "程式客製標示:",ARG_VAL(5)
   DISPLAY "程式寫入模式:",ARG_VAL(6)
   
   #讀取外部參數
   LET gs_mod_name  = ARG_VAL(2)
   LET gs_prog_name = ARG_VAL(3)
   LET gs_prog_ver  = ARG_VAL(4)
   LET gs_dgenv     = ARG_VAL(5)
   LET gs_wmode     = ARG_VAL(6)
   
   #預設null為1(gs_wmode)
   IF cl_null(gs_wmode) THEN
      LET gs_wmode = '1'
   END IF

   IF gs_dgenv = 's' OR gs_dgenv = 'c' THEN
   ELSE
      DISPLAY "ERROR:DGENV-不為s,c!(",gs_dgenv,")"
	  EXIT PROGRAM
   END IF   
   
   IF cl_null(gs_prog_ver) THEN
      DISPLAY "ERROR:無輸入section版次, 程式中斷!"
      EXIT PROGRAM
   END IF
   LET gi_prog_ver = gs_prog_ver   #210511-00011 代追單
   
   #連結資料庫
   CALL cl_db_connect("ds",FALSE)
   
   #先確定是否有進行客製
   LET li_cnt = 0
   LET lr_dzbc.dzbc001 = gs_prog_name
   LET lr_dzbc.dzbc002 = gs_prog_ver
   LET lr_dzbc.dzbc007 = gs_dgenv
   SELECT COUNT(1) INTO li_cnt FROM dzbc_t 
                  WHERE dzbc001 = lr_dzbc.dzbc001 AND #程式名稱
                        dzbc002 = lr_dzbc.dzbc002 AND #PR版次
                        dzbc005 <> 's'            AND #使用標示
                        dzbc007 = lr_dzbc.dzbc007      
   
   SELECT dzax003 INTO gc_dzax003 FROM dzax_t 
    WHERE dzax001 = lr_dzbc.dzbc001 AND
          dzax006 = lr_dzbc.dzbc007 
    
   IF li_cnt > 0 THEN
      LET gb_section = TRUE
   ELSE
      LET gb_section = FALSE
   END IF
   
   #進行解析與寫入
   IF li_cnt = 0 OR gc_dzax003 = 'Y' OR gs_wmode = '2' OR gs_wmode = '3' THEN
      CALL adzp176_parse_file()
   ELSE
      DISPLAY '該程式(',ARG_VAL(3),')已解開Section調整, 不再產生標準Section!'
   END IF

   #離開作業
   CALL cl_ap_exitprogram("0")   #2019/12/09 #191209-00016#2   add
END MAIN
    
#解析tgx2
PRIVATE FUNCTION adzp176_parse_file()
   DEFINE lt_file            TEXT
   DEFINE ls_tgx2_filename   STRING
   DEFINE ls_file            STRING
   DEFINE li_start           INTEGER
   DEFINE li_end             INTEGER
   DEFINE li_name_start      INTEGER
   DEFINE li_name_end        INTEGER
   DEFINE li_ver             INTEGER
   DEFINE li_idx             INTEGER
   DEFINE ls_section_name    STRING
   DEFINE ls_section         STRING
   DEFINE ls_readonly        STRING #read only
   DEFINE lr_dzbc            RECORD LIKE dzbc_t.*
   
   #DISPLAY 'adzp176_parse_file'
   
   LOCATE lt_file IN FILE 
   
   #讀取程式路徑
   LET ls_tgx2_filename = os.Path.join(FGL_GETENV(UPSHIFT(gs_mod_name)),"dzx")
   LET ls_tgx2_filename = os.Path.join(ls_tgx2_filename,"tgl")
   LET ls_tgx2_filename = os.Path.join(ls_tgx2_filename,gs_prog_name CLIPPED||".tgx2")
   DISPLAY "程式檔位置:",ls_tgx2_filename

   IF NOT os.Path.exists(ls_tgx2_filename) THEN
      DISPLAY "ERROR:檔案:",ls_tgx2_filename.trim()," 不存在!"
      EXIT PROGRAM
   END IF
   
   CALL lt_file.readFile(ls_tgx2_filename)
   
   LET ls_file = lt_file
   LET li_start = 1

   LET li_idx = 1   
   
   #先將狀態碼改為N
   LET lr_dzbc.dzbc001 = gs_prog_name
   LET lr_dzbc.dzbc002 = gs_prog_ver
   LET lr_dzbc.dzbc007 = gs_dgenv
   UPDATE dzbc_t 
      SET dzbcstus  = 'N'
      WHERE dzbc001 = lr_dzbc.dzbc001 AND
            dzbc002 = lr_dzbc.dzbc002 AND
            dzbc005 = 's'             AND
            dzbc007 = lr_dzbc.dzbc007 

   WHILE ls_file.getIndexOf('{<section id',li_start) <> 0 

      #先取section名稱
      LET li_name_start   = ls_file.getIndexOf('{<section id="',li_start ) + 14
      LET li_name_end     = ls_file.getIndexOf('"',li_name_start         ) - 1
      LET ls_section_name = ls_file.subString(li_name_start,li_name_end)
      
      #取出read only
      LET ls_readonly = ls_file.subString(li_name_start,ls_file.getIndexOf('\n',li_name_end))
      IF ls_readonly.getIndexOf('readonly="Y"',1) > 0 THEN
         LET ls_readonly = 'Y'
      ELSE
         LET ls_readonly = 'N'
      END IF 
      
      #先抓section內容起頭(tag結束後)
      LET li_start = ls_file.getIndexOf('\n',li_name_end) + 1
     
      #再抓section內容結尾(tag開始前)
      LET li_end = ls_file.getIndexOf('{</section>}',li_start) - 1
      
      #實際取出section內容
      LET ls_section = ls_file.subString(li_start,li_end)
      
      #寫入DB相關動作(description只進行更新, 其餘則新增)
      IF ls_section_name.getIndexOf('.description',1) > 0 THEN
         LET li_ver = adzp176_update_dzbd(gs_prog_name,ls_section_name,gs_prog_ver,ls_section)
         CALL adzp176_insert_dzbc(gs_prog_name,ls_section_name,gs_prog_ver,li_ver,li_idx,ls_readonly)
      ELSE
         IF NOT (ls_section_name.getIndexOf('.main',1) > 0 AND gc_dzax003 = "Y" AND gb_section) THEN
            LET li_ver = adzp176_insert_dzbd(gs_prog_name,ls_section_name,gs_prog_ver,ls_section)
            CALL adzp176_insert_dzbc(gs_prog_name,ls_section_name,gs_prog_ver,li_ver,li_idx,ls_readonly)
         END IF
      END IF
      
      LET li_idx = li_idx + 1
      
      #結束前將起頭換到下個位置上
      LET li_start = li_end + 1
      
   END WHILE   
   
   FREE lt_file
   
END FUNCTION

#判斷寫入dzbc
FUNCTION adzp176_insert_dzbc(ps_prog,ps_name,ps_ver,ps_sc_ver,ps_idx,ps_readonly)
   DEFINE ps_prog      LIKE type_t.chr200    #程式名稱
   DEFINE ps_ver       LIKE type_t.chr200    #section版本
   DEFINE ps_sc_ver    LIKE type_t.chr200    #section版本
   DEFINE ps_name      LIKE type_t.chr200    #section名稱
   DEFINE ps_section   STRING                #section實際內容
   DEFINE ps_idx       LIKE type_t.chr10     #section順序
   DEFINE ps_readonly  STRING                #read only
   DEFINE lr_dzbc      RECORD LIKE dzbc_t.*
   DEFINE li_cnt       INTEGER
   DEFINE lc_dzbc005   LIKE type_t.chr10
    
   #共用欄位
   LET lr_dzbc.dzbcownid = g_user
   LET lr_dzbc.dzbcowndp = g_dept
   LET lr_dzbc.dzbccrtid = g_user
   LET lr_dzbc.dzbccrtdp = g_dept
   LET lr_dzbc.dzbccrtdt = CURRENT
   LET lr_dzbc.dzbcmodid = g_user
   LET lr_dzbc.dzbcmoddt = CURRENT
    
   #section狀態
   LET lr_dzbc.dzbcstus = "Y"
   
   #程式名稱
   LET lr_dzbc.dzbc001 = ps_prog
   
   #程式版本
   LET lr_dzbc.dzbc002 = ps_ver
   
   #section名稱
   LET lr_dzbc.dzbc003 = ps_name
   
   #sction版次
   LET lr_dzbc.dzbc004 = ps_sc_ver
   
   #使用標示(固定為s)
   LET lr_dzbc.dzbc005 = 's'
   
   #產品版本
   LET lr_dzbc.dzbc006 = FGL_GETENV("ERPVER")
   
   #識別標示(KEY)
   LET lr_dzbc.dzbc007 = gs_dgenv
   
   #客戶代號
   LET lr_dzbc.dzbc008 = FGL_GETENV("CUST")
   
   #section序號
   LET lr_dzbc.dzbc010 = ps_idx USING "&&&"
   
   #section Read Only
   LET lr_dzbc.dzbc011 = ps_readonly
   
   SELECT COUNT(1) INTO li_cnt
                   FROM dzbc_t 
                  WHERE dzbc001 = lr_dzbc.dzbc001 AND
                        dzbc002 = lr_dzbc.dzbc002 AND
                        dzbc003 = lr_dzbc.dzbc003 AND
                        dzbc007 = lr_dzbc.dzbc007 

   IF li_cnt = 0 THEN
      #第一次產生
      DISPLAY "Section:",lr_dzbc.dzbc003,",第一次產生, 寫入dzbc!"
      INSERT INTO dzbc_t (dzbcstus,
                          dzbcownid,
                          dzbcowndp,
                          dzbccrtid,
                          dzbccrtdp,
                          dzbccrtdt,
                          dzbc001,
                          dzbc002,
                          dzbc003,
                          dzbc004,
                          dzbc005,
                          dzbc006,
                          dzbc007,
                          dzbc008,
                          dzbc010,
                          dzbc011) 
                  VALUES (lr_dzbc.dzbcstus,
                          lr_dzbc.dzbcownid,
                          lr_dzbc.dzbcowndp,
                          lr_dzbc.dzbccrtid,
                          lr_dzbc.dzbccrtdp,
                          lr_dzbc.dzbccrtdt,
                          lr_dzbc.dzbc001,
                          lr_dzbc.dzbc002,
                          lr_dzbc.dzbc003,
                          lr_dzbc.dzbc004,
                          lr_dzbc.dzbc005,
                          lr_dzbc.dzbc006,
                          lr_dzbc.dzbc007,
                          lr_dzbc.dzbc008,
                          lr_dzbc.dzbc010,
                          lr_dzbc.dzbc011)
      IF SQLCA.sqlcode THEN
         DISPLAY "ERROR:",lr_dzbc.dzbc003," 資料儲存發生錯誤!"
		 EXIT PROGRAM
      END IF
   ELSE
      #非第一次產生且為標準section
      IF g_t100debug >= 3 THEN 
         DISPLAY '讀取Section:',lr_dzbc.dzbc003,'(非第一次產生)'
      END IF
      SELECT dzbc005 INTO lc_dzbc005 
                     FROM dzbc_t 
                    WHERE dzbc001 = lr_dzbc.dzbc001 AND
                          dzbc002 = lr_dzbc.dzbc002 AND
                          dzbc003 = lr_dzbc.dzbc003 AND
                          dzbc007 = lr_dzbc.dzbc007 
      IF lc_dzbc005 = 's' THEN
         UPDATE dzbc_t SET dzbc004   = lr_dzbc.dzbc004,
                           dzbc010   = lr_dzbc.dzbc010,
                           dzbc011   = lr_dzbc.dzbc011,
                           dzbcmodid = lr_dzbc.dzbcmodid,
                           dzbcmoddt = lr_dzbc.dzbcmoddt,
                           dzbcstus  = lr_dzbc.dzbcstus
                     WHERE dzbc001 = lr_dzbc.dzbc001 AND
                           dzbc002 = lr_dzbc.dzbc002 AND
                           dzbc003 = lr_dzbc.dzbc003 AND
                           dzbc007 = lr_dzbc.dzbc007 
         IF SQLCA.sqlcode THEN
            DISPLAY "ERROR:",lr_dzbc.dzbc003," 資料更新發生錯誤!"
			EXIT PROGRAM
         END IF
         IF g_t100debug >= 3 THEN 
            DISPLAY '更新Section:',lr_dzbc.dzbc003,'(',SQLCA.sqlcode,')'
         END IF
      END IF
   END IF
                         
END FUNCTION 

#判斷寫入dzbd, 回傳當下的section版次
FUNCTION adzp176_insert_dzbd(ps_prog,ps_name,ps_ver,ps_section)
   DEFINE ps_prog      LIKE type_t.chr200     #程式名稱
   DEFINE ps_ver       LIKE type_t.chr200     #section版本
   DEFINE ps_name      LIKE type_t.chr200     #section名稱
   DEFINE ps_section   STRING                 #section實際內容
   DEFINE ls_section   STRING                 #section實際內容(比較用)
   DEFINE lr_dzbd      RECORD LIKE dzbd_t.*
   DEFINE li_cnt       INTEGER
   DEFINE li_ver       INTEGER
   
   LOCATE lr_dzbd.dzbd098 IN FILE
   
   #共用欄位
   LET lr_dzbd.dzbdownid = g_user
   LET lr_dzbd.dzbdowndp = g_dept
   LET lr_dzbd.dzbdcrtid = g_user
   LET lr_dzbd.dzbdcrtdp = g_dept
   LET lr_dzbd.dzbdcrtdt = CURRENT
   LET lr_dzbd.dzbdmodid = g_user
   LET lr_dzbd.dzbdmoddt = CURRENT
   
   #section狀態
   LET lr_dzbd.dzbdstus = "Y"
   
   #程式名稱
   LET lr_dzbd.dzbd001 = ps_prog

   #section名稱
   LET lr_dzbd.dzbd002 = ps_name
   
   #sction版次
   LET lr_dzbd.dzbd003 = 1
   
   #使用標示(固定為s)
   LET lr_dzbd.dzbd004 = 's'
   IF gs_wmode = '2' THEN
      LET lr_dzbd.dzbd004 = 'p' 
   END IF
   
   #客戶代號
   LET lr_dzbd.dzbd005 = FGL_GETENV("CUST")
   
   #section內容
   LET lr_dzbd.dzbd098 = ps_section
   
   #先確定是否有資料
   SELECT COUNT(1) INTO li_cnt FROM dzbd_t 
      WHERE dzbd001 = lr_dzbd.dzbd001 AND 
            dzbd002 = lr_dzbd.dzbd002 AND 
            #dzbd003 = lr_dzbd.dzbd003 AND 
            dzbd004 = lr_dzbd.dzbd004

   IF li_cnt = 0 THEN
      #如果還沒有資料
      DISPLAY "Section:",lr_dzbd.dzbd002,",第一次產生, 寫入dzbd!"
      INSERT INTO dzbd_t (dzbdstus,
                          dzbdownid,
                          dzbdowndp,
                          dzbdcrtid,
                          dzbdcrtdp,
                          dzbdcrtdt,
                          dzbd001,
                          dzbd002,
                          dzbd003,
                          dzbd004,
                          dzbd005,
                          dzbd098) 
                  VALUES (lr_dzbd.dzbdstus,
                          lr_dzbd.dzbdownid,
                          lr_dzbd.dzbdowndp,
                          lr_dzbd.dzbdcrtid,
                          lr_dzbd.dzbdcrtdp,
                          lr_dzbd.dzbdcrtdt,
                          lr_dzbd.dzbd001,
                          lr_dzbd.dzbd002,
                          lr_dzbd.dzbd003,
                          lr_dzbd.dzbd004,
                          lr_dzbd.dzbd005,
                          lr_dzbd.dzbd098)
      IF SQLCA.sqlcode THEN
         DISPLAY "ERROR:",lr_dzbd.dzbd003," 資料儲存發生錯誤!"
		 EXIT PROGRAM
      END IF 
   ELSE
      #讀出最新版次
      &ifdef is_ora   #210511-00011 代追單
      SELECT MAX(to_NUMBER(dzbd003)) INTO li_ver FROM dzbd_t 
         WHERE dzbd001 = lr_dzbd.dzbd001 AND 
               dzbd002 = lr_dzbd.dzbd002 AND 
               dzbd004 = lr_dzbd.dzbd004
      &endif   #210511-00011 代追單
      #210511-00011 ---start---
      &ifdef is_pgs
      SELECT MAX(dzbd003) INTO li_ver FROM dzbd_t
         WHERE dzbd001 = lr_dzbd.dzbd001 AND
               dzbd002 = lr_dzbd.dzbd002 AND
               dzbd004 = lr_dzbd.dzbd004
      &endif
      #210511-00011 --- end ---
      LET lr_dzbd.dzbd003 = li_ver USING "<<<<<<<<<<<<<<<"
   
      #讀出最新section
      SELECT dzbd098 INTO lr_dzbd.dzbd098 FROM dzbd_t 
         WHERE dzbd001 = lr_dzbd.dzbd001 AND 
               dzbd002 = lr_dzbd.dzbd002 AND 
               dzbd003 = lr_dzbd.dzbd003 AND 
               dzbd004 = lr_dzbd.dzbd004

      LET ls_section = lr_dzbd.dzbd098
      #有異動則寫入
      IF ls_section <> ps_section THEN
         DISPLAY "Section:",lr_dzbd.dzbd002,",異動, 更新dzbd!"
         LET lr_dzbd.dzbd098 = ps_section
         LET li_cnt = lr_dzbd.dzbd003 + 1
         LET lr_dzbd.dzbd003 = li_cnt USING "<<<<<<<<<<<<<<<"
         INSERT INTO dzbd_t (dzbdstus,
                             dzbdownid,
                             dzbdowndp,
                             dzbdcrtid,
                             dzbdcrtdp,
                             dzbdcrtdt,
                             dzbd001,
                             dzbd002,
                             dzbd003,
                             dzbd004,
                             dzbd005,
                             dzbd098) 
                     VALUES (lr_dzbd.dzbdstus,
                             lr_dzbd.dzbdownid,
                             lr_dzbd.dzbdowndp,
                             lr_dzbd.dzbdcrtid,
                             lr_dzbd.dzbdcrtdp,
                             lr_dzbd.dzbdcrtdt,
                             lr_dzbd.dzbd001,
                             lr_dzbd.dzbd002,
                             lr_dzbd.dzbd003,
                             lr_dzbd.dzbd004,
                             lr_dzbd.dzbd005,
                             lr_dzbd.dzbd098)
         IF SQLCA.sqlcode THEN
            DISPLAY "ERROR:",lr_dzbd.dzbd003," 資料儲存發生錯誤!"
			EXIT PROGRAM
         END IF 
      END IF
   
   END IF

   FREE lr_dzbd.dzbd098
   
   RETURN lr_dzbd.dzbd003
   
END FUNCTION 

#判斷更新dzbd
FUNCTION adzp176_update_dzbd(ps_prog,ps_name,ps_ver,ps_section)
   DEFINE ps_prog      LIKE type_t.chr200     #程式名稱
   DEFINE ps_ver       LIKE type_t.chr200     #section版本
   DEFINE ps_name      LIKE type_t.chr200     #section名稱
   DEFINE ps_section   STRING                 #section實際內容
   DEFINE ls_section   STRING                 #section實際內容(比較用)
   DEFINE lr_dzbd      RECORD LIKE dzbd_t.*
   DEFINE li_cnt       INTEGER
   DEFINE li_ver       INTEGER
   
   LOCATE lr_dzbd.dzbd098 IN FILE
   
   #共用欄位
   LET lr_dzbd.dzbdownid = g_user
   LET lr_dzbd.dzbdowndp = g_dept
   LET lr_dzbd.dzbdcrtid = g_user
   LET lr_dzbd.dzbdcrtdp = g_dept
   LET lr_dzbd.dzbdcrtdt = CURRENT
   LET lr_dzbd.dzbdmodid = g_user
   LET lr_dzbd.dzbdmoddt = CURRENT
   
   #section狀態
   LET lr_dzbd.dzbdstus = "Y"
   
   #程式名稱
   LET lr_dzbd.dzbd001 = ps_prog

   #section名稱
   LET lr_dzbd.dzbd002 = ps_name
   
   #sction版次
   LET lr_dzbd.dzbd003 = 1
   
   #使用標示(固定為s)
   LET lr_dzbd.dzbd004 = 's'
   IF gs_wmode = '2' THEN
      LET lr_dzbd.dzbd004 = 'p' 
   END IF
   
   #客戶代號
   LET lr_dzbd.dzbd005 = FGL_GETENV("CUST")
   
   #section內容
   LET lr_dzbd.dzbd098 = ps_section
   
   #先確定是否有資料
   SELECT COUNT(1) INTO li_cnt FROM dzbd_t 
      WHERE dzbd001 = lr_dzbd.dzbd001 AND 
            dzbd002 = lr_dzbd.dzbd002 AND 
            #dzbd003 = lr_dzbd.dzbd003 AND 
            dzbd004 = lr_dzbd.dzbd004

   IF li_cnt = 0 THEN
      #如果還沒有資料
      DISPLAY "Section:",lr_dzbd.dzbd002,",第一次產生, 寫入dzbd!"
      INSERT INTO dzbd_t (dzbdstus,
                          dzbdownid,
                          dzbdowndp,
                          dzbdcrtid,
                          dzbdcrtdp,
                          dzbdcrtdt,
                          dzbd001,
                          dzbd002,
                          dzbd003,
                          dzbd004,
                          dzbd005,
                          dzbd098) 
                  VALUES (lr_dzbd.dzbdstus,
                          lr_dzbd.dzbdownid,
                          lr_dzbd.dzbdowndp,
                          lr_dzbd.dzbdcrtid,
                          lr_dzbd.dzbdcrtdp,
                          lr_dzbd.dzbdcrtdt,
                          lr_dzbd.dzbd001,
                          lr_dzbd.dzbd002,
                          lr_dzbd.dzbd003,
                          lr_dzbd.dzbd004,
                          lr_dzbd.dzbd005,
                          lr_dzbd.dzbd098)
      IF SQLCA.sqlcode THEN
         DISPLAY "ERROR:",lr_dzbd.dzbd003," 資料儲存發生錯誤!"
		 EXIT PROGRAM
      END IF 
   ELSE
      #讀出最新版次
      &ifdef is_ora   #210511-00011 代追單
      SELECT MAX(to_NUMBER(dzbd003)) INTO li_ver FROM dzbd_t 
         WHERE dzbd001 = lr_dzbd.dzbd001 AND 
               dzbd002 = lr_dzbd.dzbd002 AND 
               dzbd004 = lr_dzbd.dzbd004
      &endif   #210511-00011 代追單
      #210511-00011 ---start---
      &ifdef is_pgs
      SELECT MAX(dzbd003) INTO li_ver FROM dzbd_t
         WHERE dzbd001 = lr_dzbd.dzbd001 AND
               dzbd002 = lr_dzbd.dzbd002 AND
               dzbd004 = lr_dzbd.dzbd004
      &endif
      #210511-00011 --- end ---
      LET lr_dzbd.dzbd003 = li_ver USING "<<<<<<<<<<<<<<<"
   
      #讀出最新section
      SELECT dzbd098 INTO lr_dzbd.dzbd098 FROM dzbd_t 
         WHERE dzbd001 = lr_dzbd.dzbd001 AND 
               dzbd002 = lr_dzbd.dzbd002 AND 
               dzbd003 = lr_dzbd.dzbd003 AND 
               dzbd004 = lr_dzbd.dzbd004

      LET ls_section = lr_dzbd.dzbd098
      #有異動則寫入
      IF ls_section <> ps_section THEN
         DISPLAY "Section:",lr_dzbd.dzbd002,",異動, 更新dzbd!"
         LET lr_dzbd.dzbd098 = ps_section
         LET li_cnt = lr_dzbd.dzbd003
         LET lr_dzbd.dzbd003 = li_cnt USING "<<<<<<<<<<<<<<<"
         UPDATE dzbd_t SET dzbd098   = lr_dzbd.dzbd098,
                           dzbdmodid = lr_dzbd.dzbdmodid,
                           dzbdmoddt = lr_dzbd.dzbdmoddt
          WHERE dzbd001 = lr_dzbd.dzbd001 AND 
                dzbd002 = lr_dzbd.dzbd002 AND 
                dzbd003 = lr_dzbd.dzbd003 AND 
                dzbd004 = lr_dzbd.dzbd004
         IF SQLCA.sqlcode THEN
            DISPLAY "ERROR:",lr_dzbd.dzbd003," 資料儲存發生錯誤!"
			EXIT PROGRAM
         END IF 
      END IF
   
   END IF

   FREE lr_dzbd.dzbd098
   
   RETURN lr_dzbd.dzbd003
   
END FUNCTION 




