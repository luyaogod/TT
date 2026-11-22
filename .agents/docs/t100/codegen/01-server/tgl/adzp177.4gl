# code_version.: 5.25.02-2014.01.20(0000-0001)

IMPORT os
SCHEMA ds

#+ code_id......: adzp177.4gl
#+ descriptions.: section combiner
#+ code_creator.: ka0132
#+ modify       : ka0132 2017/12/26 修正解開Section判斷
#                                   增加freestyle解開資訊
#                                   增加單個Section解開資訊
 
GLOBALS "../../cfg/top_global.inc"

DEFINE gs_prog_name LIKE type_t.chr50
DEFINE gs_mod_name  LIKE type_t.chr50
DEFINE gs_prog_ver  LIKE type_t.chr50
DEFINE gi_prog_ver  LIKE type_t.num10   #210511-00011 代追單
DEFINE gs_dgenv     LIKE type_t.chr50
DEFINE gs_wmode     LIKE type_t.chr50
DEFINE gr_section   DYNAMIC ARRAY OF RECORD
       tag          STRING,
       id           LIKE dzbc_t.dzbc003,
       ver          LIKE dzbc_t.dzbc004,
       type         LIKE dzbc_t.dzbc005,
       readonly     LIKE dzbc_t.dzbc011,
       content      LIKE dzbd_t.dzbd098
       END RECORD
DEFINE g_tgl_name   STRING 
       
MAIN
   DEFINE li_idx INTEGER

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
   
   IF cl_null(gs_prog_ver) THEN
      DISPLAY "ERROR:無輸入section版次, 程式中斷!"
     EXIT PROGRAM
   END IF
   LET gi_prog_ver = gs_prog_ver   #210511-00011 代追單

   #連結資料庫
   CALL cl_db_connect("ds",FALSE)
   
   #進行解析與寫入
   CALL adzp177_get_order()
   
   #取得section相關內容
   CALL adzp177_get_section()
   
   #組合並寫入tgl
   CALL adzp177_combine()
   
   #FREE CLOB
   FOR li_idx = 1 TO gr_section.getLength()
      FREE gr_section[li_idx].content
   END FOR
   
   IF gs_wmode = '2' THEN
      #刪除暫存資料
      DELETE FROM dzbd_t 
            WHERE dzbd001 = gs_prog_name AND 
                  dzbd004 = 'p'  
   END IF

END MAIN

#+ 確認section順序
PRIVATE FUNCTION adzp177_get_order()
   DEFINE ls_filename    STRING  #tgx2檔案名稱及路徑
   DEFINE ls_cmd         STRING  #grep指令
   DEFINE ls_sql         STRING  #sql
   DEFINE ls_sections    STRING  #讀取所有section
   DEFINE lc_ch          base.Channel 
   DEFINE lst_token      base.StringTokenizer 
   DEFINE ls_token       STRING
   DEFINE li_idx         INTEGER
   DEFINE li_left        INTEGER #section左側起點
   DEFINE li_right       INTEGER #section右側終點
   DEFINE li_cnt         INTEGER
   DEFINE li_dzbc002     INTEGER #210511-00011 代追單
   
   #開啟channel
   LET lc_ch = base.Channel.create()
   
   #確認tgx2位置(讀取sction順序)
   LET ls_filename = os.Path.join(FGL_GETENV(UPSHIFT(gs_mod_name)),"dzx")
   LET ls_filename = os.Path.join(ls_filename,"tgl")
   LET ls_filename = os.Path.join(ls_filename,gs_prog_name CLIPPED||".tgx2")
   DISPLAY "程式檔位置:",ls_filename
   
   #確定一下資料庫的Section是否有順序
   LET li_cnt = 0
   SELECT COUNT(1) INTO li_cnt FROM dzbc_t
                              WHERE dzbc001  = gs_prog_name
                              # AND dzbc002  = gs_prog_ver   #210511-00011 代追單
                                AND dzbc002  = gi_prog_ver   #210511-00011 代追單
                                AND dzbc007  = gs_dgenv
                                AND dzbcstus = 'Y'
                                AND dzbc010  IS NOT NULL

   #先確定tgx2是否存在
   #IF NOT os.Path.exists(ls_filename) THEN
   #   DISPLAY '因tgx2不存在, 轉由資料庫Section判定順序!'
   
   #先確定Section是否含順序
   IF li_cnt > 0 THEN
      DISPLAY '由資料庫順序判定Section組合!'
      LET ls_sql = #" SELECT dzbc003,dzbc011 FROM dzbc_t ", #2017/12/26
                   " SELECT dzbc003,dzbc011,dzbc005 FROM dzbc_t ", #2017/12/26
                   "  WHERE dzbc001  =  ? ",
                   "    AND dzbc002  =  ? ",
                   "    AND dzbc007  =  ? ",
                   "    AND dzbcstus = 'Y'",
                   "  ORDER BY dzbc010 "
      PREPARE section_pre FROM ls_sql
      DECLARE section_cur CURSOR FOR section_pre
      
   #  OPEN section_cur USING gs_prog_name, gs_prog_ver, gs_dgenv   #210511-00011 代追單
      OPEN section_cur USING gs_prog_name, gi_prog_ver, gs_dgenv   #210511-00011 代追單
      LET li_idx = 1
      FOREACH section_cur INTO gr_section[li_idx].id,gr_section[li_idx].readonly
                              ,gr_section[li_idx].type #2017/12/26
         LET gr_section[li_idx].tag = "{<section id=\"",gr_section[li_idx].id,"\""
         IF gr_section[li_idx].readonly = 'Y' THEN
            LET gr_section[li_idx].tag = gr_section[li_idx].tag, " readonly=\"Y\""
         END IF
         #2017/12/26
         LET gr_section[li_idx].tag = gr_section[li_idx].tag, " type=\"",gr_section[li_idx].type,"\""
         #2017/12/26
         LET gr_section[li_idx].tag = gr_section[li_idx].tag, " >}"
         LOCATE gr_section[li_idx].content IN FILE
         LET li_idx = li_idx + 1
      END FOREACH
      CALL gr_section.deleteElement(gr_section.getLength())
      
   ELSE
      DISPLAY '由tgx2順序判定Section組合!'
      #定義grep指令
      LET ls_cmd = "grep 'section id=' ", ls_filename
      
      CALL lc_ch.openPipe( ls_cmd, "r" )
      
      #讀取section順序
      WHILE TRUE
         LET ls_sections = ls_sections, ",", lc_ch.readLine()
         IF lc_ch.isEof() THEN 
            EXIT WHILE  
         END IF
      END WHILE
      CALL lc_ch.close()
      
      #解析section順序並儲存到陣列中
      LET li_idx = 1
      LET lst_token = base.StringTokenizer.create(ls_sections, ",")
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET gr_section[li_idx].tag = ls_token
         LET li_left  = ls_token.getIndexOf('"',1)+1
         LET li_right = ls_token.getIndexOf('"',li_left)-1
         LET gr_section[li_idx].id = ls_token.subString(li_left,li_right)
         #開啟CLOB
         LOCATE gr_section[li_idx].content IN FILE
         LET li_idx = li_idx + 1
      END WHILE
   END IF

END FUNCTION 

#+ 取得section內容
PRIVATE FUNCTION adzp177_get_section()
   DEFINE li_idx INTEGER
   DEFINE lc_env LIKE type_t.chr50

   LET lc_env = gs_dgenv
   
   #先讀取section版次以及使用標示
   FOR li_idx = 1 TO gr_section.getLength()
   
      #先從dzbc確定要用的版次&使用標示
      SELECT dzbc004, dzbc005 
        INTO gr_section[li_idx].ver, gr_section[li_idx].type
        FROM dzbc_t 
       WHERE dzbc001 = gs_prog_name 
       # AND dzbc002 = gs_prog_ver   #210511-00011 代追單
         AND dzbc002 = gi_prog_ver   #210511-00011 代追單
         AND dzbc003 = gr_section[li_idx].id      
         AND dzbc007 = lc_env
    
      #轉為預覽模式, 原本讀取S改為讀取P
      IF gr_section[li_idx].type = 's' AND gs_wmode = '2' THEN
         LET gr_section[li_idx].type = 'p'
      END IF

      #根據使用標示(s,c)決定使用的section版次與內容
      SELECT dzbd098
        INTO gr_section[li_idx].content
        FROM dzbd_t 
       WHERE dzbd001 = gs_prog_name
         AND dzbd002 = gr_section[li_idx].id
         AND dzbd003 = gr_section[li_idx].ver
         AND dzbd004 = gr_section[li_idx].type

   END FOR

END FUNCTION 

#+ 組合tgl
PRIVATE FUNCTION adzp177_combine()
   DEFINE ls_code_filename   STRING
   DEFINE lchannel_write     base.Channel
   DEFINE li_idx             INTEGER
   DEFINE ls_prog_name       STRING
   DEFINE li_dzbc005         INTEGER
   DEFINE lc_dzax003         LIKE dzax_t.dzax003 #2017/12/26
   
   LET lchannel_write = base.Channel.create()
   CALL lchannel_write.setDelimiter("")
   
   LET ls_prog_name = gs_prog_name CLIPPED, ".tgl"

   IF FGL_GETENV("GEN_MODE") = "diff" THEN
   #  LET ls_prog_name = gs_prog_name,".tgl.ver",gs_prog_ver USING "<<<<<", ".", ARG_VAL(5)   #210511-00011 代追單
      LET ls_prog_name = gs_prog_name,".tgl.ver",gi_prog_ver USING "<<<<<", ".", ARG_VAL(5)   #210511-00011 代追單
   ELSE    
      LET ls_prog_name = gs_prog_name,".tgl"
      IF gs_wmode = '2' THEN
         DISPLAY 'Info:產出預覽檔tgl2'
         LET ls_prog_name = ls_prog_name,"2"
      END IF
   END IF
   
   #產出程式路徑
   LET ls_code_filename = os.Path.join(FGL_GETENV(UPSHIFT(gs_mod_name)),"dzx")
   LET ls_code_filename = os.Path.join(ls_code_filename,"tgl")
   LET ls_code_filename = os.Path.join(ls_code_filename,ls_prog_name)
   
   DISPLAY "產生檔位置:",ls_code_filename
   CALL lchannel_write.openFile( ls_code_filename, "w" )
   
   LET li_dzbc005 = 0
   SELECT COUNT(1) INTO li_dzbc005 FROM dzbc_t 
    WHERE dzbc001= gs_prog_name AND dzbc005 <> 's'
   #2017/12/26
    # AND dzbc002 = gs_prog_ver   #210511-00011 代追單
      AND dzbc002 = gi_prog_ver   #210511-00011 代追單
   #2017/12/26
   IF li_dzbc005 > 0 THEN
      CALL lchannel_write.write("#該程式已解開Section, 不再透過樣板產出!")
   ELSE  
      CALL lchannel_write.write("#該程式未解開Section, 採用最新樣板產出!")
   END IF
   
   #2017/12/26
   LET lc_dzax003 = 'N'
   SELECT dzax003 INTO lc_dzax003 
     FROM dzax_t 
    WHERE dzax001 = gs_prog_name
      AND dzax006 = gs_dgenv
   IF lc_dzax003 = 'Y' THEN
      CALL lchannel_write.write("#該程式為freestyle程式!")
   ELSE  
      CALL lchannel_write.write("#該程式非freestyle程式!")
   END IF
   #2017/12/26
   
   #將section回寫tgl(根據array順序)
   FOR li_idx = 1 TO gr_section.getLength()
      CALL lchannel_write.write(gr_section[li_idx].tag)
      CALL lchannel_write.write(gr_section[li_idx].content)
      CALL lchannel_write.write("{</section>}\n")
   END FOR
   
   CALL lchannel_write.close()

END FUNCTION 
