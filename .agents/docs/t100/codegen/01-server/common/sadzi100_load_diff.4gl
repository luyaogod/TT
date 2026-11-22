IMPORT os

SCHEMA ds

GLOBALS "../../cfg/top_global.inc"
#單身 type 宣告
PRIVATE TYPE type_g_diff_d  RECORD
   line_no   STRING, #行號 
   left_col  STRING, #左邊差異收集
   right_col STRING  #右邊差異收集  
       END RECORD

DEFINE g_diff_d DYNAMIC ARRAY OF type_g_diff_d  

#+比對差異
PUBLIC FUNCTION sadzi100_load_diff(ps_file_new,ps_file_old)
   DEFINE ps_file_new  STRING  #新的樣板
   DEFINE ps_file_old  STRING  #前一版的樣板
   DEFINE lch_channel base.Channel
   DEFINE ls_cmd       STRING 
   DEFINE ls_temp      STRING  
   DEFINE ls_temp2     STRING
   DEFINE li_cnt       LIKE type_t.num5 
   DEFINE li_chk       LIKE type_t.num5
   DEFINE ls_left      STRING 
   DEFINE ls_right     STRING 

   #diff code_i01.template /u1/t10dap/erp/mdl/code_i01.template
   CALL g_diff_d.clear()
   
   DISPLAY " ps_file_new:",ps_file_new
   DISPLAY " ps_file_old:",ps_file_old

   LET lch_channel = base.Channel.create()
   LET ls_cmd = "diff ",ps_file_new ," " ,ps_file_old
   #CALL lch_channel.setDelimiter("")
   CALL lch_channel.openPipe(ls_cmd ,"r")
   
   WHILE lch_channel.read(ls_temp)
      IF ls_temp.trim() IS NULL THEN CONTINUE WHILE END IF
 
      #濾掉分隔線
      IF ls_temp.getIndexOf("---",1) THEN CONTINUE WHILE END IF
      LET ls_temp2 = ls_temp
      
      IF ls_temp2.subString(1,1) = "<" OR ls_temp2.subString(1,1) = ">" THEN 
         IF ls_temp2.subString(1,1) = "<" THEN 
            LET li_chk = FALSE 
            LET ls_left = ls_left,"\n",ls_temp2
         ELSE 
            LET li_chk = FALSE
            LET ls_right = ls_right,"\n",ls_temp2
         END IF 
      ELSE
         #表示數字
         FOR li_cnt = 1 TO ls_temp2.getLength()
             IF ls_temp2.subString(li_cnt,li_cnt) NOT MATCHES "[01-9]" THEN 
                
                #表示要換列收集 
                LET li_chk = TRUE 
                LET g_diff_d[g_diff_d.getLength()+1].line_no = ls_temp2.subString(1,li_cnt - 1)
                EXIT FOR 
             END IF  
         END FOR
      END IF

      #遇到數字，準備開始收集差異
      IF li_chk THEN 
         #第一筆先跳過 遇到數字
         IF g_diff_d.getLength() = 1 THEN 
            CONTINUE WHILE 
         ELSE
         #第二筆收集前一筆的資料，接下來以此類推
            LET g_diff_d[g_diff_d.getLength() - 1].left_col = ls_left 
            LET g_diff_d[g_diff_d.getLength() - 1].right_col = ls_right  
            LET ls_left = ""
            LET ls_right = "" 
         END IF 
      END IF
   END WHILE
   
   IF g_diff_d.getLength() = 0 THEN

   ELSE 
      #收集最後一筆 
      IF g_diff_d[g_diff_d.getLength()].left_col IS NULL  AND g_diff_d[g_diff_d.getLength()].right_col IS NULL  THEN 
         LET g_diff_d[g_diff_d.getLength()].left_col = ls_left 
         LET g_diff_d[g_diff_d.getLength()].right_col = ls_right 
      END IF 
   END IF 

   CALL lch_channel.close()
   IF g_t100debug = "9" THEN 
      DISPLAY "-------------------"
      FOR li_cnt = 1 TO g_diff_d.getLength()
          DISPLAY "cnt: ",li_cnt 
          DISPLAY " line_no: ",g_diff_d[li_cnt].line_no 
          DISPLAY " left_col: ",g_diff_d[li_cnt].left_col  
          DISPLAY " right_col: ",g_diff_d[li_cnt].right_col  
      END FOR  
      DISPLAY "-------------------"
   END IF 
   RETURN g_diff_d
END FUNCTION 



