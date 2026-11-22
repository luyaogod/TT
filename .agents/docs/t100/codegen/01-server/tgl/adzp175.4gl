# code_version.: 5.25.02-2012.04.16(0000-0001)

IMPORT os
SCHEMA ds

#+ code_id......: adzp175.4gl
#+ descriptions.: tgl整理
#+ code_creator.: ka0132
#+ modify       : ka0132 2018/03/06 BUG修正
#+ modify       : 200804-00006#1 2020/08/05 by 00824  postgre功能回追

GLOBALS "../../cfg/top_global.inc"

DEFINE g_function RECORD
       ins     BOOLEAN,
       mod     BOOLEAN,
       del     BOOLEAN,
       qry     BOOLEAN,
       rep     BOOLEAN
       END RECORD
       
DEFINE gs_prog_name STRING
DEFINE gs_mod_name  STRING
MAIN

   #先讀取tgl判讀所需資訊(function使用與否)
   CALL adzp150_read_function()
   
   #折行整理
   CALL adzp150_read_and_replace()

END MAIN

#+ 讀取tgl判斷使用的FUNCTION
PRIVATE FUNCTION adzp150_read_function()
   DEFINE lchannel_read           base.Channel
   DEFINE ls_sample_filename      STRING
   DEFINE ls_readline             STRING
   DEFINE ls_text                 STRING

   LET g_function.ins = FALSE
   LET g_function.mod = FALSE
   LET g_function.del = FALSE
   LET g_function.qry = FALSE
   LET g_function.rep = FALSE
   
   LET gs_prog_name = ARG_VAL(3)
   LET gs_mod_name  = ARG_VAL(2)
   LET lchannel_read = base.Channel.create()

   CALL lchannel_read.setDelimiter("")
   
   #讀取程式路徑
   LET ls_sample_filename = os.Path.join(FGL_GETENV(UPSHIFT(gs_mod_name)),"dzx")
   LET ls_sample_filename = os.Path.join(ls_sample_filename,"tgl")
   LET ls_sample_filename = os.Path.join(ls_sample_filename,gs_prog_name CLIPPED||".tgx")
   DISPLAY "程式檔位置:",ls_sample_filename

   IF NOT os.Path.exists(ls_sample_filename) THEN
      DISPLAY "ERROR:檔案:",ls_sample_filename.trim()," 不存在!"
      EXIT PROGRAM
   END IF
   CALL lchannel_read.openFile( ls_sample_filename CLIPPED, "r" )

   #讀取及轉換
   WHILE TRUE
   
      LET ls_text = ""
      LET ls_readline = lchannel_read.readLine()
      
      #新增
      IF ls_readline.getIndexOf('ON ACTION insert',1) THEN
         LET g_function.ins = TRUE
      END IF
      
      #修改
      IF ls_readline.getIndexOf('ON ACTION modify',1) THEN
         LET g_function.mod = TRUE
      END IF
      
      #刪除
      IF ls_readline.getIndexOf('ON ACTION delete',1) THEN
         LET g_function.del = TRUE
      END IF
      
      #查詢
      IF ls_readline.getIndexOf('ON ACTION query',1) THEN
         LET g_function.qry = TRUE
      END IF
      
      #複製
      IF ls_readline.getIndexOf('ON ACTION reproduce',1) THEN
         LET g_function.rep = TRUE
      END IF
        
      IF lchannel_read.isEof() THEN
         EXIT WHILE
      END IF

   END WHILE 

   CALL lchannel_read.close()
   
END FUNCTION


#+ 讀取tgl產出tglx
PRIVATE FUNCTION adzp150_read_and_replace()
   DEFINE lchannel_read           base.Channel
   DEFINE lchannel_write          base.Channel
   DEFINE ls_readline             STRING
   DEFINE ls_text                 STRING
   DEFINE ls_text_tmp             STRING
   DEFINE ls_code_filename        STRING
   DEFINE ls_sample_filename      STRING
   DEFINE li_max_length           LIKE type_t.num10
   DEFINE li_min_length           LIKE type_t.num10
   DEFINE li_tmp_length           LIKE type_t.num10
   DEFINE li_idx                  LIKE type_t.num10
   DEFINE ls_space                STRING
   DEFINE lb_func                 BOOLEAN
    
   LET lchannel_read = base.Channel.create()
   LET lchannel_write = base.Channel.create()

   CALL lchannel_read.setDelimiter("")
   CALL lchannel_write.setDelimiter("")
   
   #讀取程式路徑
   LET ls_sample_filename = os.Path.join(FGL_GETENV(UPSHIFT(gs_mod_name)),"dzx")
   LET ls_sample_filename = os.Path.join(ls_sample_filename,"tgl")
   LET ls_sample_filename = os.Path.join(ls_sample_filename,gs_prog_name CLIPPED||".tgx")
   DISPLAY "程式檔位置:",ls_sample_filename

   IF NOT os.Path.exists(ls_sample_filename) THEN
      DISPLAY "ERROR:檔案:",ls_sample_filename.trim()," 不存在!"
      EXIT PROGRAM
   END IF
   CALL lchannel_read.openFile( ls_sample_filename CLIPPED, "r" )

   #產出程式路徑
   LET ls_code_filename = os.Path.join(FGL_GETENV(UPSHIFT(gs_mod_name)),"dzx")
   LET ls_code_filename = os.Path.join(ls_code_filename,"tgl")
   LET ls_code_filename = os.Path.join(ls_code_filename,gs_prog_name CLIPPED||".tgx2")
   
   DISPLAY "產生檔位置:",ls_code_filename
   CALL lchannel_write.openFile( ls_code_filename, "w" )

   #讀取及轉換
   WHILE TRUE
   
      LET ls_text = ""
      LET ls_readline = lchannel_read.readLine()
        
      IF lchannel_read.isEof() THEN
         EXIT WHILE
      END IF
      
      #遇到註解段落整理
      IF ls_readline.getIndexOf('#該樣板不需此段落',1) #OR ls_readline.getIndexOf('#此程式無',1) THEN
         THEN
         CONTINUE WHILE
      END IF
      
      LET lb_func = FALSE
      
      #遇到不需要的FUNCTION不產出
      #CALL adzp150_function_filter(lchannel_read,ls_readline) RETURNING lb_func,lchannel_read
      #
      #IF lb_func THEN
      #   CONTINUE WHILE
      #END IF
      
      #遇到過長段落進行整理
   #  LET li_tmp_length = 100   #200804-00006 mark
      LET li_tmp_length = 300   #200804-00006
      LET li_min_length = 1
      LET li_max_length = li_min_length + li_tmp_length - 1
      LET ls_text_tmp = ""
      LET li_idx = 1
      LET ls_space = ""
      IF ls_readline.getLength() > li_tmp_length THEN
         WHILE TRUE
            IF ls_readline.subString(li_idx,li_idx) = " " THEN
               LET ls_space = ls_space, " "
            ELSE
               IF ls_readline.subString(li_idx,li_idx) = "#" THEN
                  LET ls_space = ls_space, "#"
               END IF
               EXIT WHILE
            END IF
            LET li_idx = li_idx + 1
         END WHILE
         LET ls_space = ls_space, "    "
      END IF

      LET li_idx = 1
      #判斷關鍵字
      IF adzp175_newline_chk(ls_readline) AND 
         ls_readline.getLength() > li_tmp_length THEN
         
         #開始處理折行
         LET li_tmp_length = li_tmp_length - ls_space.getLength()
         WHILE TRUE 
         
            IF li_max_length > ls_readline.getLength() THEN
               LET li_max_length = ls_readline.getLength()
               IF li_idx > 1 THEN
                  LET ls_text_tmp = ls_text_tmp, ls_space
               END IF
               LET ls_text_tmp = ls_text_tmp, ls_readline.subString(li_min_length, li_max_length)
               EXIT WHILE
            END IF
            IF li_idx > 1 THEN
               LET ls_text_tmp = ls_text_tmp, ls_space
            END IF
            #針對空白/逗點進行折行
            IF ls_readline.getIndexOf(',',li_max_length) > 0 OR 
               ls_readline.getIndexOf(' ',li_max_length) > 0 THEN
               IF ls_readline.getIndexOf(',',li_max_length) > 0 THEN
                  LET li_max_length = ls_readline.getIndexOf(',',li_max_length)
               ELSE
                  LET li_max_length = ls_readline.getIndexOf(' ',li_max_length)
               END IF
            ELSE
               LET li_max_length = ls_readline.getLength()
            END IF
            LET ls_text_tmp = ls_text_tmp, ls_readline.subString(li_min_length, li_max_length), ' \n'
            LET li_min_length = li_max_length + 1
            LET li_max_length = li_max_length + li_tmp_length
            IF li_min_length > ls_readline.getLength() THEN
               EXIT WHILE
            END IF
            LET li_idx = li_idx + 1
            
         END WHILE
         LET ls_readline = ls_text_tmp
      ELSE
         #確認那些段落過長但無折行
         #IF NOT adzp175_newline_chk(ls_readline) AND 
         #ls_readline.getLength() > li_tmp_length THEN
         #   DISPLAY ls_readline
         #END IF
      END IF

      #一般未進行任何處理段落產出
      IF ls_text IS NULL THEN
         LET ls_text = ls_readline
      END IF
      CALL lchannel_write.write(ls_text)

   END WHILE 

   CALL lchannel_read.close() 
   CALL lchannel_write.close()
   
   #對產生器寫出的檔案權限在UNIX下均全部打開
   IF os.Path.separator() = "/" THEN
      IF os.Path.chrwx(ls_code_filename,511) THEN
      END IF
   END IF
   
END FUNCTION

#判斷該function是否濾除
PRIVATE FUNCTION adzp150_function_filter(pchannel_read,ps_line)
   DEFINE pchannel_read           base.Channel
   DEFINE ps_line                 STRING
   DEFINE ls_readline             STRING
   DEFINE lb_return               BOOLEAN
   
   LET lb_return = FALSE
   
   IF (ps_line.getIndexOf('PRIVATE FUNCTION ',1)> 0) THEN
      
      IF (ps_line.getIndexOf('_insert()'   ,1) > 0 AND g_function.ins = FALSE) OR
         (ps_line.getIndexOf('_modify()'   ,1) > 0 AND g_function.mod = FALSE) OR
         (ps_line.getIndexOf('_delete()'   ,1) > 0 AND g_function.del = FALSE) OR
         (ps_line.getIndexOf('_query()'    ,1) > 0 AND g_function.qry = FALSE) OR
         (ps_line.getIndexOf('_reproduce()',1) > 0 AND g_function.rep = FALSE) OR
         (ps_line.getIndexOf('_input()'    ,1) > 0 AND g_function.ins = FALSE AND g_function.mod = FALSE AND g_function.rep = FALSE) THEN
         
         LET lb_return = TRUE
         
         WHILE TRUE
            LET ls_readline = pchannel_read.readLine()
            
            IF ls_readline.getIndexOf('END FUNCTION',1) > 0 THEN
               EXIT WHILE
            END IF
          
         END WHILE
         
      END IF
   END IF

   RETURN lb_return, pchannel_read
   
END FUNCTION

#判斷是否需要換行
PRIVATE FUNCTION adzp175_newline_chk(ps_line)
   DEFINE ps_line  STRING
   
   IF ps_line.getIndexOf('DISPLAY',      1) > 0 OR
      ps_line.getIndexOf('INPUT',        1) > 0 OR
      ps_line.getIndexOf('CONSTRUCT',    1) > 0 OR
      ps_line.getIndexOf('FOREACH',      1) > 0 OR
      ps_line.getIndexOf('FETCH',        1) > 0 OR
      ps_line.getIndexOf('INSERT',       1) > 0 OR
      ps_line.getIndexOf('UPDATE',       1) > 0 OR
      ps_line.getIndexOf('SELECT',       1) > 0 OR
      ps_line.getIndexOf('FROM',         1) > 0 OR
      ps_line.getIndexOf('INTO',         1) > 0 OR
      ps_line.getIndexOf('VALUES',       1) > 0 OR
      ps_line.getIndexOf('[l_ac',        1) > 0 OR
      ps_line.getIndexOf('[g_detail_idx',1) > 0 THEN
      #2018/03/06
      IF ps_line.getIndexOf('#',1) > 0 THEN
         IF ps_line.getIndexOf(' ',ps_line.getIndexOf('#',1)) OR
            ps_line.getIndexOf(',',ps_line.getIndexOf('#',1)) THEN
            RETURN FALSE
         END IF  
      END IF
      #2018/03/06
      RETURN TRUE
   ELSE
      RETURN FALSE
   END IF
   
END FUNCTION 




