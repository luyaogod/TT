#+ Version..: T6-ERP-1.00.00 Build-00001
#+
#+ Filename.: sadzp030_tab_relation.4gl
#+ Buildtype: p01樣板自動產生
#+ Memo.....: 產生tab用
#190509-00014#1 2019/05/17 ka0132 修正重複判斷freestyle問題
#201123-00010#1 2020/11/23 10167  調整欄位空值寫法(dzag004 IS NULL改為dzag004 IS NULL OR dzag004 = '')

IMPORT os

SCHEMA ds

GLOBALS "../../cfg/top_global.inc"
GLOBALS "../4gl/adzp150_global.inc"   #180411-00062

GLOBALS
   DEFINE g_prog_id         LIKE dzfi_t.dzfi001
   DEFINE g_sd_ver          LIKE dzaa_t.dzaa002  #規格版次/程式版次
#  DEFINE g_erp_ver         LIKE dzaa_t.dzaa008  #產品版本
   DEFINE g_dgenv           LIKE dzaa_t.dzaa009  #客製標示
#  DEFINE g_cust            LIKE dzaa_t.dzaa010  #客戶編號
   DEFINE g_prog_type       STRING
   DEFINE g_cross_4gl_stus  LIKE type_t.num5     #跨4gl判斷正確與否
   DEFINE g_table_main      STRING
   DEFINE g_keep_pk         DYNAMIC ARRAY OF RECORD
            dzag002   LIKE dzag_t.dzag002, #table編號
            dzed004   STRING
                            END RECORD
   DEFINE g_keep_fk         DYNAMIC ARRAY OF RECORD
            dzag002   LIKE dzag_t.dzag002, #table編號
            dzag004   LIKE dzag_t.dzag004, #上層table編號
            dzed004   STRING,
            dzed006   STRING    #鍵值欄位 / 外來鍵值欄位
                            END RECORD
   DEFINE g_more_pk_then_head BOOLEAN    #偵測到比單頭PK更多的欄位設定
   DEFINE g_stop_chk_sd_ver_freestyle BOOLEAN  #DEFAULT=FALSE
END GLOBALS

DEFINE gi_cnt   LIKE type_t.num5
DEFINE ga_table DYNAMIC ARRAY OF RECORD
          table_type      LIKE type_t.num5,   #1:單頭表 2:單身表
          table_id        STRING,             #表編號
          table_pk        STRING,
          table_fk        STRING
#         table_ent       LIKE type_t.num5    #暫時先留白不填入 有 1 沒有 0
   END RECORD

############################################################
#+ @code
#+ 函式目的  宣告單頭關聯 SCROLL CURSOR
############################################################
PRIVATE FUNCTION sadzp030_tab_relation_query_m()

   DEFINE ls_sql       STRING 

   LET ls_sql = "SELECT dzag002,dzag004 ",
                 " FROM dzaa_t,dzag_t", 
           #--------------------------dzaa---------------------#
                " WHERE dzaa001 = '",g_prog_id CLIPPED,"'",  #規格編號
                  " AND dzaa002 = '",g_sd_ver CLIPPED,"'",   #規格版次
                  " AND dzaa003 = 'TABLE'",                  #識別碼
                  " AND dzaa005 = '4'",                      #識別碼編號
#                 " AND dzaa008 = '",g_erp_ver CLIPPED,"' ", #產品版本
                  " AND dzaa009 = '",g_dgenv CLIPPED,"' ",   #客製標示
#                 " AND dzaa010 = '",g_cust CLIPPED,"' ",    #客戶編號
                  " AND dzaastus = 'Y'",             #狀態碼取 Y 有效
           #--------------------------dzag---------------------#
                  " AND dzag001 = dzaa001 ",         #規格編號
                  " AND dzag003 = dzaa004 ",         #識別碼版次
                  " AND dzag006 = dzaa006 ",         #使用標示
                  #" AND dzag004 IS NULL",                       #201123-00010#1 by 10167 mark
                  " AND ( dzag004 IS NULL OR dzag004 = '' ) ",   #201123-00010#1 by 10167 add
                  " AND dzag005 <> 'Y'",
#                 " AND dzag011 = dzaa010 ",         #客戶編號
                  " AND dzagstus = 'Y'"
   PREPARE sadzp030_prepare_m FROM ls_sql
   DECLARE sadzp030_cs_m                             # SCROLL CURSOR

   SCROLL CURSOR WITH HOLD FOR sadzp030_prepare_m
        
   OPEN sadzp030_cs_m     
   CALL sadzp030_tab_relation_fetch_m() 
END FUNCTION 

############################################################
#+ @code
#+ 函式目的 宣告單身關聯 SCROLL CURSOR
#+ @param  p_dzag002  CHAR(15) 程式名稱
############################################################
PRIVATE FUNCTION sadzp030_tab_relation_query_d(p_dzag002)

   DEFINE p_dzag002    LIKE dzag_t.dzag002
   DEFINE ls_sql       STRING 
   
   LET ls_sql = "SELECT dzag002,dzag004 ",
                 " FROM dzaa_t,dzag_t ",
           #--------------------------dzaa---------------------#
                " WHERE dzaa001 = '",g_prog_id CLIPPED,"'",  #規格編號
                  " AND dzaa002 = '",g_sd_ver CLIPPED,"'",   #規格版次
                  " AND dzaa003 = 'TABLE'",                  #識別碼
                  " AND dzaa005 = '4'",                      #識別碼編號
#                 " AND dzaa008 = '",g_erp_ver CLIPPED,"' ", #產品版本
                  " AND dzaa009 = '",g_dgenv CLIPPED,"' ",   #客製標示
#                 " AND dzaa010 = '",g_cust CLIPPED,"' ",    #客戶編號
                  " AND dzaastus = 'Y'",             #狀態碼取 Y 有效
           #--------------------------dzag---------------------#
                  " AND dzag001 = dzaa001 ",         #規格編號
                  " AND dzag003 = dzaa004 ",         #識別碼版次
                  " AND dzag006 = dzaa006 ",         #使用標示
                  " AND dzag004 = '",p_dzag002 CLIPPED,"'", 
#                 " AND dzag011 = dzaa010 ",         #客戶編號
                  " AND dzagstus = 'Y'"              #狀態碼取 Y 有效
                 
   PREPARE sadzp030_prepare_d FROM ls_sql
   DECLARE sadzp030_cs_d 
   SCROLL CURSOR WITH HOLD FOR sadzp030_prepare_d

   OPEN sadzp030_cs_d
   CALL sadzp030_tab_relation_fetch_d() 
END FUNCTION 

############################################################
#+ @code
#+ 函式目的 取設定表資料(主表)
#+
#+ @return STRING  主表table
#+ @return STRING  主表table primary key 
#+ @return array   主表相關table array
############################################################
PUBLIC FUNCTION sadzp030_tab_relation()

   DEFINE l_dzag002         LIKE dzag_t.dzag002
   DEFINE l_dzag005         LIKE dzag_t.dzag005
   DEFINE l_dzag004         LIKE dzag_t.dzag004
   DEFINE ls_table_main_pk  STRING 
   DEFINE ls_main_table_id  STRING 
   DEFINE ls_sql            STRING 
   DEFINE ls_temp           STRING
   DEFINE li_cnt            LIKE type_t.num5
   DEFINE li_dzag002        LIKE type_t.num5

   #取dzag r.a 應用的資料表 
   LET ls_temp = " FROM dzaa_t,dzag_t", 
           #--------------------------dzaa---------------------#
                " WHERE dzaa001 = '",g_prog_id CLIPPED,"'",  #規格編號
                  " AND dzaa002 = '",g_sd_ver CLIPPED,"'",   #規格版次
                  " AND dzaa003 = 'TABLE'",                  #識別碼
                  " AND dzaa005 = '4'",                      #識別碼編號
#                 " AND dzaa008 = '",g_erp_ver CLIPPED,"' ", #產品版本
                  " AND dzaa009 = '",g_dgenv CLIPPED,"' ",   #客製標示
#                 " AND dzaa010 = '",g_cust CLIPPED,"' ",    #客戶編號
                  " AND dzaastus = 'Y'",             #狀態碼取 Y 有效
           #--------------------------dzag---------------------#
                  " AND dzag001 = dzaa001 ",         #規格編號
                  " AND dzag003 = dzaa004 ",         #識別碼版次
                  " AND dzag006 = dzaa006 ",         #使用標示
                  " AND dzag005 = 'Y'",              #抓取主表 (dzag005='Y')
#                 " AND dzag011 = dzaa010 ",         #客戶編號
                  " AND dzagstus = 'Y'"

   LET ls_sql = "SELECT COUNT(1) ",ls_temp
   PREPARE sadzp030_tab_count_pre FROM ls_sql
   EXECUTE sadzp030_tab_count_pre INTO li_dzag002
   IF li_dzag002 > 1 THEN
      DISPLAY cl_getmsg("adz-01125",g_lang)
     #DISPLAY "ERROR: 本作業查到的主表 (dzag005='Y') 大於1筆,請確認!"
      LET g_more_pk_then_head = TRUE
   END IF

   LET ls_sql = "SELECT dzag002 ",ls_temp
   PREPARE sadzp030_tab_pre FROM ls_sql
   EXECUTE sadzp030_tab_pre INTO l_dzag002
   IF g_t100debug >= 3 THEN
      IF g_t100debug >= 9 THEN
         DISPLAY "INFO: SQL:",ls_sql
      END IF
      IF g_t100debug >= 6 THEN
         DISPLAY "INFO: 從dzaa,dzag抓取主表編號,取出:",l_dzag002," 狀態:",SQLCA.SQLCODE
      END IF
      IF SQLCA.SQLCODE = 100 THEN
         SELECT COUNT(1) INTO li_cnt FROM dzaa_t,dzag_t
        #--------------------------dzaa---------------------#
          WHERE dzaa001 = g_prog_id  #規格編號
            AND dzaa002 = g_sd_ver   #規格版次
            AND dzaa003 = "TABLE"    #識別碼
            AND dzaa005 = "4"        #識別碼編號
#           AND dzaa008 = g_erp_ver  #產品版本
            AND dzaa009 = g_dgenv    #客製標示
#           AND dzaa010 = g_cust     #客戶編號
            AND dzaastus = "Y"       #狀態碼取 Y 有效
           #--------------------------dzag---------------------#
            AND dzag001 = dzaa001    #規格編號
            AND dzag003 = dzaa004    #識別碼版次
            AND dzag006 = dzaa006    #使用標示
#           AND dzag011 = dzaa010    #客戶編號
            AND dzagstus = "Y"
         IF g_t100debug >= 3 THEN
            DISPLAY "INFO: 從dzaa,dzag抓出",li_cnt,"筆表單資料,都不是主表"
         END IF
      END IF   
   END IF
   FREE sadzp030_tab_pre

   LET gi_cnt = 1
   CALL ga_table.clear()
   IF cl_null(l_dzag002) THEN 
      IF g_t100debug >= 3 THEN
         DISPLAY cl_getmsg("adz-01126",g_lang)
        #DISPLAY "注意: 抓取表編號時, 資料不存在, 不再取PK!!"
      END IF
   ELSE
      LET ls_main_table_id = l_dzag002

      #取主表pk
      LET ls_table_main_pk = sadzp030_tab_relation_pk(l_dzag002)

      #有單頭及單身 取單頭有關單身子節點
      CALL sadzp030_tab_relation_query_d(l_dzag002)

      #單檔多table或單身多table 取關聯資料
      CALL sadzp030_tab_relation_query_m()
   END IF

   CLOSE sadzp030_cs_m   
   CLOSE sadzp030_cs_d

   RETURN l_dzag002,ls_table_main_pk,ga_table
END FUNCTION 

############################################################
#+ @code
#+ 函式目的 取出primary key
#+ @param  p_dzag002 CHAR(15) 表格名稱
#+ @return STRING primary key 
############################################################
PUBLIC FUNCTION sadzp030_tab_relation_pk(p_dzag002)

   DEFINE p_dzag002   LIKE dzag_t.dzag002
   DEFINE l_dzed004   LIKE dzed_t.dzed004
   DEFINE ls_sql      STRING 
   DEFINE ls_return   STRING
   DEFINE li_cnt      LIKE type_t.num5
   DEFINE li_pos      LIKE type_t.num5
   DEFINE l_token     base.StringTokenizer
   DEFINE l_tokenp    base.StringTokenizer
   DEFINE la_dzed     DYNAMIC ARRAY OF RECORD
             dzed004  STRING,
             checked  LIKE type_t.chr1
                      END RECORD
   DEFINE ls_dzed004  STRING
   DEFINE ls_pkcol    STRING
   DEFINE li_status   LIKE type_t.num5

   #查看是否有處理過
   FOR li_cnt = 1 TO g_keep_pk.getLength()
      IF g_keep_pk[li_cnt].dzag002 = p_dzag002 THEN
         LET ls_return = g_keep_pk[li_cnt].dzed004
         RETURN ls_return
         EXIT FOR
      END IF
   END FOR

   LET ls_sql = "SELECT dzag008 FROM dzaa_t,dzag_t", 
           #--------------------------dzaa---------------------#
                " WHERE dzaa001 = '",g_prog_id CLIPPED,"'",  #規格編號
                  " AND dzaa002 = '",g_sd_ver CLIPPED,"'",   #規格版次
                  " AND dzaa003 = 'TABLE'",                  #識別碼
                  " AND dzaa005 = '4'",                      #識別碼編號
#                 " AND dzaa008 = '",g_erp_ver CLIPPED,"' ", #產品版本
                  " AND dzaa009 = '",g_dgenv CLIPPED,"' ",   #客製標示
#                 " AND dzaa010 = '",g_cust CLIPPED,"' ",    #客戶編號
                  " AND dzaastus = 'Y'",             #狀態碼取 Y 有效
           #--------------------------dzag---------------------#
                  " AND dzag001 = dzaa001 ",         #規格編號
                  " AND dzag002 = '",p_dzag002,"' ", 
                  " AND dzag003 = dzaa004 ",         #識別碼版次
                  " AND dzag006 = dzaa006 ",         #使用標示
#                 " AND dzag011 = dzaa010 ",         #客戶編號
                  " AND dzagstus = 'Y' "
          
   PREPARE sadzp030_tab_relation_pk1 FROM ls_sql
   EXECUTE sadzp030_tab_relation_pk1 INTO l_dzed004

   IF SQLCA.SQLCODE OR cl_null(l_dzed004) THEN
      FREE sadzp030_tab_relation_pk1
   ELSE
      IF g_t100debug >= 3 THEN
         DISPLAY "INFO: 從設計器取得",p_dzag002 CLIPPED," PK=",l_dzed004 CLIPPED
      END IF
      FREE sadzp030_tab_relation_pk1

      LET ls_return = l_dzed004
   END IF

   IF ls_return IS NULL THEN
      LET ls_sql = "SELECT dzed004 FROM dzed_t", 
                   " WHERE dzed001 = '",p_dzag002,"'", 
                     " AND dzed003 = 'P'"  #鍵值形式 = p 表示為pk
      PREPARE sadzp030_tab_relation_pk2 FROM ls_sql
      EXECUTE sadzp030_tab_relation_pk2 INTO l_dzed004
      FREE sadzp030_tab_relation_pk2
   END IF

   IF g_t100debug >= 3 THEN
      DISPLAY "INFO: 從r.t取得",p_dzag002 CLIPPED," PK=",l_dzed004 CLIPPED
   END IF

   LET ls_dzed004 = ""
   LET li_status = FALSE
   FOR li_pos = 1 TO g_keep_fk.getLength()
      IF g_keep_fk[li_pos].dzag002 = p_dzag002 THEN
         LET li_status = TRUE

         #有存在資料,就把PK更新一下欄位的前後順序
         LET l_token = base.StringTokenizer.create(l_dzed004 CLIPPED, ",")
         LET li_cnt = 1
         WHILE l_token.hasMoreTokens()
            LET la_dzed[li_cnt].dzed004 = l_token.nextToken()
            LET la_dzed[li_cnt].checked = "N"

            #遇到ent先拉出
            IF la_dzed[li_cnt].dzed004.subString(la_dzed[li_cnt].dzed004.getLength()-2,
                                                 la_dzed[li_cnt].dzed004.getLength()) = "ent" THEN
               LET ls_dzed004 = ls_dzed004,la_dzed[li_cnt].dzed004,","
               LET la_dzed[li_cnt].checked = "Y"
            END IF

            LET li_cnt = li_cnt + 1
         END WHILE

         #取出主表PK
         LET l_tokenp = base.StringTokenizer.create(g_keep_fk[li_pos].dzed004, ",")
         WHILE l_tokenp.hasMoreTokens()
            LET ls_pkcol = l_tokenp.nextToken()
            FOR li_cnt = 1 TO la_dzed.getLength()
               IF la_dzed[li_cnt].checked = "N" THEN
                  IF ls_pkcol = la_dzed[li_cnt].dzed004 THEN
                     LET ls_dzed004 = ls_dzed004,la_dzed[li_cnt].dzed004,","
                     LET la_dzed[li_cnt].checked = "Y"
                  END IF
               END IF
            END FOR
         END WHILE

         #配上剩餘的項目
         FOR li_cnt = 1 TO la_dzed.getLength()
            IF la_dzed[li_cnt].checked = "N" THEN
               LET ls_dzed004 = ls_dzed004,la_dzed[li_cnt].dzed004,","
            END IF
         END FOR

         EXIT FOR
      END IF
   END FOR
   IF NOT li_status THEN
      IF g_t100debug >= 3 THEN
         DISPLAY "INFO: ",p_dzag002,"沒有存在可供比對的上階FK相關轉換資料, 自動省略不做...."
      END IF
   END IF

   #保險
   IF ls_dzed004 IS NULL THEN LET ls_dzed004 = l_dzed004 END IF

   #排除 ent
   LET li_cnt = ls_dzed004.getLength()
   IF li_cnt > 0 AND ls_dzed004.subString(li_cnt,li_cnt) = "," THEN
      LET ls_dzed004 = ls_dzed004.subString(1,ls_dzed004.getLength()-1)
   END IF

   IF g_t100debug >= 3 THEN
      DISPLAY "INFO:   順序調整後",p_dzag002 CLIPPED," PK=",ls_dzed004 CLIPPED
   END IF

   LET ls_return = sadzp030_tab_relation_ent(ls_dzed004)

   #存到資料表內
   LET li_cnt = g_keep_pk.getLength() + 1
   LET g_keep_pk[li_cnt].dzag002 = p_dzag002
   LET g_keep_pk[li_cnt].dzed004 = ls_return

   RETURN ls_return
END FUNCTION 

############################################################
#+ @code
#+ 函式目的 取出單身設定表資料
############################################################
PRIVATE FUNCTION sadzp030_tab_relation_fetch_d()

   DEFINE l_dzag002 LIKE dzag_t.dzag002
   DEFINE l_dzag004 LIKE dzag_t.dzag004
   
   FETCH NEXT sadzp030_cs_d INTO l_dzag002,l_dzag004

   IF SQLCA.sqlcode THEN
      RETURN
   ELSE
      #取單身foreign key
      CALL sadzp030_tab_relation_tab(l_dzag002,l_dzag004)
      #遞迴呼叫
      CALL sadzp030_tab_relation_fetch_d()
   END IF 
END FUNCTION 

############################################################
#+ @code
#+ 函式目的 取出單頭設定表資料
############################################################
PRIVATE FUNCTION sadzp030_tab_relation_fetch_m()

   DEFINE l_dzag002 LIKE dzag_t.dzag002
   DEFINE l_dzag004 LIKE dzag_t.dzag004

   FETCH NEXT sadzp030_cs_m INTO l_dzag002,l_dzag004

   IF SQLCA.sqlcode THEN
      RETURN
   ELSE
      #表示有多個table 區分單頭多table 跟單頭的主表有關聯的FK
      CALL sadzp030_tab_relation_tab(l_dzag002,l_dzag004)
      #遞迴呼叫
      CALL sadzp030_tab_relation_fetch_m()
   END IF 
    
END FUNCTION 

############################################################
#+ @code
#+ 函式目的 取出foreign key
#+ @param  p_dzag002  CHAR(15) 表格名稱
#+ @param  p_dzag004  CHAR(15) 表格名稱
############################################################
PUBLIC FUNCTION sadzp030_tab_relation_fk(p_dzag002,p_dzag004,li_type) 

   DEFINE p_dzag002   LIKE dzag_t.dzag002  #table編號
   DEFINE p_dzag004   LIKE dzag_t.dzag004  #上層table編號
   DEFINE l_dzed004   LIKE dzed_t.dzed004
   DEFINE l_dzed006   LIKE dzed_t.dzed006
   DEFINE ls_sql      STRING 
   DEFINE ls_next     STRING
   DEFINE l_token4    base.StringTokenizer
   DEFINE l_token6    base.StringTokenizer
   DEFINE l_tokenp    base.StringTokenizer
   DEFINE li_cnt      LIKE type_t.num5
   DEFINE la_dzed     DYNAMIC ARRAY OF RECORD
             dzed004  STRING,
             dzed006  STRING,
             checked  LIKE type_t.chr1
                      END RECORD
   DEFINE ls_pkcol    STRING
   DEFINE ls_dzed004,ls_dzed006 STRING    #鍵值欄位 / 外來鍵值欄位
   DEFINE li_status   LIKE type_t.num5
   DEFINE li_type     LIKE type_t.num5

   #搜尋是否處理過
   FOR li_cnt = 1 TO g_keep_fk.getLength()
      IF g_keep_fk[li_cnt].dzag002 = p_dzag002 AND 
         g_keep_fk[li_cnt].dzag004 = p_dzag004 THEN

         LET ls_dzed004 = g_keep_fk[li_cnt].dzed004
         LET ls_dzed006 = g_keep_fk[li_cnt].dzed006
         RETURN ls_dzed004,ls_dzed006
         EXIT FOR
      END IF
   END FOR

   #先確認數據庫中有沒有這張表
   SELECT COUNT(1) INTO li_cnt FROM gztz_t WHERE gztz001 = p_dzag002
   IF li_cnt = 0 THEN
      IF li_type <> 4 THEN
         DISPLAY cl_getmsg_parm("adz-01127",g_lang,p_dzag002 CLIPPED)
        #DISPLAY "Error: 取FK頭的表:",p_dzag002 CLIPPED,",不在實體資料庫內!(gztz_t not exists)."
      END IF
      RETURN ls_dzed004,ls_dzed006
   END IF
   SELECT COUNT(1) INTO li_cnt FROM gztz_t WHERE gztz001 = p_dzag004
   IF li_cnt = 0 THEN
      IF li_type <> 4 THEN
         DISPLAY cl_getmsg_parm("adz-01128",g_lang,p_dzag004 CLIPPED)
        #DISPLAY "Error: 取FK身的表:",p_dzag004 CLIPPED,",不在實體資料庫內!(gztz_t not exists)."
      END IF
      RETURN ls_dzed004,ls_dzed006
   END IF

   #取出table編號
   LET ls_next = p_dzag002
   LET ls_next = ls_next.subString(1,ls_next.getIndexOf("_t",1)-1)

   LET ls_sql = "SELECT dzag009,dzag010 FROM dzaa_t,dzag_t", 
           #--------------------------dzaa---------------------#
                " WHERE dzaa001 = '",g_prog_id CLIPPED,"'",  #規格編號
                  " AND dzaa002 = '",g_sd_ver CLIPPED,"'",   #規格版次
                  " AND dzaa003 = 'TABLE'",                  #識別碼
                  " AND dzaa005 = '4'",                      #識別碼編號
#                 " AND dzaa008 = '",g_erp_ver CLIPPED,"' ", #產品版本
                  " AND dzaa009 = '",g_dgenv CLIPPED,"' ",   #客製標示
#                 " AND dzaa010 = '",g_cust CLIPPED,"' ",    #客戶編號
                  " AND dzaastus = 'Y'",             #狀態碼取 Y 有效
           #--------------------------dzag---------------------#
                  " AND dzag001 = dzaa001 ",         #規格編號
                  " AND dzag003 = dzaa004 ",         #識別碼版次
                  " AND dzag006 = dzaa006 ",         #使用標示
                  " AND dzag002 = '",p_dzag002,"'", 
                  " AND dzag004 = '",p_dzag004,"'",  #外來鍵關聯表格
#                 " AND dzag011 = dzaa010 ",         #客戶編號
                  " AND dzagstus = 'Y' "
          
   PREPARE sadzp030_tab_relation_fk1 FROM ls_sql
   EXECUTE sadzp030_tab_relation_fk1 INTO l_dzed004,l_dzed006

   IF SQLCA.SQLCODE OR cl_null(l_dzed004) THEN
      FREE sadzp030_tab_relation_fk1

      LET ls_sql = "SELECT dzed004,dzed006 FROM dzed_t", 
                   " WHERE dzed001 = '",p_dzag002,"'", 
                     " AND ( dzed002 like '",ls_next.trim(),"_fk%' ",   #設定取名稱含有_FK
                     "    OR dzed002 like '",ls_next.trim(),"_fd%' ", #或有_FD
                     "    OR dzed002 like 'tic_",ls_next.trim(),"_fk%' ",   #設定取名稱含有tic和_FK
                     "    OR dzed002 like 'tic_",ls_next.trim(),"_fd%' ) ", #或有tic和_FD
                     " AND dzed003 = 'F'",              #鍵值形式=F 表示為FK
                     " AND dzed005 = '",p_dzag004,"'"   #外來鍵關聯表格
              
      PREPARE sadzp030_tab_relation_fk2 FROM ls_sql
      EXECUTE sadzp030_tab_relation_fk2 INTO l_dzed004,l_dzed006

      IF g_t100debug >= 3 THEN
         DISPLAY "INFO: 從r.t取得",p_dzag002,"對上階",p_dzag004 CLIPPED," FK=",l_dzed004 CLIPPED,' vs=',l_dzed006
      END IF
      FREE sadzp030_tab_relation_fk2
   ELSE
      IF g_t100debug >= 3 THEN
         DISPLAY "INFO: 從設計器取得",p_dzag002,"對上階",p_dzag004 CLIPPED," FK=",l_dzed004 CLIPPED,' vs=',l_dzed006
      END IF
      FREE sadzp030_tab_relation_fk1
   END IF

   #把對應建置起來之後,在依照主表(p_uptable)的PK順序重新排序
   LET l_token4 = base.StringTokenizer.create(l_dzed004 CLIPPED, ",")
   LET l_token6 = base.StringTokenizer.create(l_dzed006 CLIPPED, ",")
   LET li_cnt = 1
   WHILE l_token4.hasMoreTokens() AND l_token6.hasMoreTokens()
      LET la_dzed[li_cnt].dzed004 = l_token4.nextToken()
      LET la_dzed[li_cnt].dzed006 = l_token6.nextToken()
      LET la_dzed[li_cnt].checked = "N"

      #遇到ent先拉出
      IF la_dzed[li_cnt].dzed004.subString(la_dzed[li_cnt].dzed004.getLength()-2,
                                           la_dzed[li_cnt].dzed004.getLength()) = "ent" THEN
         LET ls_dzed004 = ls_dzed004,la_dzed[li_cnt].dzed004,","
         LET ls_dzed006 = ls_dzed006,la_dzed[li_cnt].dzed006,","
         LET la_dzed[li_cnt].checked = "Y"
      END IF

      LET li_cnt = li_cnt + 1
   END WHILE

   #取出主表PK
   LET ls_sql = sadzp030_tab_relation_pk(p_dzag004)
   LET l_tokenp = base.StringTokenizer.create(ls_sql, ",")
   WHILE l_tokenp.hasMoreTokens()
      LET ls_pkcol = l_tokenp.nextToken()
      FOR li_cnt = 1 TO la_dzed.getLength()
         IF ls_pkcol = la_dzed[li_cnt].dzed006 AND la_dzed[li_cnt].checked = "N" THEN
            LET ls_dzed004 = ls_dzed004,la_dzed[li_cnt].dzed004,","
            LET ls_dzed006 = ls_dzed006,la_dzed[li_cnt].dzed006,","
            LET la_dzed[li_cnt].checked = "Y"
         END IF
      END FOR
   END WHILE

   #配上剩餘的項目
   FOR li_cnt = 1 TO la_dzed.getLength()
      IF la_dzed[li_cnt].checked = "N" THEN
         LET ls_dzed004 = ls_dzed004,la_dzed[li_cnt].dzed004,","
         LET ls_dzed006 = ls_dzed006,la_dzed[li_cnt].dzed006,","
         DISPLAY cl_getmsg_parm("adz-01129",g_lang,la_dzed[li_cnt].dzed006)
        #DISPLAY "ERROR: 偵測到比單頭PK更多的欄位設定:欄位:",la_dzed[li_cnt].dzed006,"在單頭不具備PK欄位身分!"
         LET g_more_pk_then_head = TRUE
      END IF
   END FOR

   LET ls_dzed004 = ls_dzed004.subString(1,ls_dzed004.getLength()-1)
   LET ls_dzed006 = ls_dzed006.subString(1,ls_dzed006.getLength()-1)

   IF g_t100debug >= 6 THEN
      DISPLAY "INFO:   異動後取得",p_dzag002,"對上階",p_dzag004 CLIPPED," FK=",ls_dzed004 CLIPPED,' vs=',ls_dzed006
   END IF

   #存到已處理過的資料表內
   IF p_dzag002 IS NOT NULL AND p_dzag004 IS NOT NULL THEN
      LET li_cnt = g_keep_fk.getLength() + 1
      LET g_keep_fk[li_cnt].dzag002 = p_dzag002
      LET g_keep_fk[li_cnt].dzag004 = p_dzag004
      LET g_keep_fk[li_cnt].dzed004 = ls_dzed004
      LET g_keep_fk[li_cnt].dzed006 = ls_dzed006
   END IF

   RETURN ls_dzed004,ls_dzed006
END FUNCTION

############################################################
#+ @code
#+ 函式目的 將取出foreign key填入table陣列
#+ @param  p_dzag002  CHAR(15) 表格名稱
#+ @param  p_dzag004  CHAR(15) 表格名稱
############################################################
PRIVATE FUNCTION sadzp030_tab_relation_tab(p_dzag002,p_dzag004)

   DEFINE p_dzag002   LIKE dzag_t.dzag002
   DEFINE p_dzag004   LIKE dzag_t.dzag004
   DEFINE l_dzed004   LIKE dzed_t.dzed004
   DEFINE l_dzed006   LIKE dzed_t.dzed006

   CALL sadzp030_tab_relation_fk(p_dzag002,p_dzag004,4) RETURNING l_dzed004,l_dzed006
   #跟單頭主表關聯  
   IF cl_null(p_dzag004) THEN 
      LET ga_table[gi_cnt].table_type = 1
   ELSE 
      LET ga_table[gi_cnt].table_type = 2 
   END IF   

   LET ga_table[gi_cnt].table_id = p_dzag002 
   LET ga_table[gi_cnt].table_fk = sadzp030_tab_relation_ent(l_dzed004)   
   LET ga_table[gi_cnt].table_pk = sadzp030_tab_relation_pk(p_dzag002)
   LET gi_cnt = gi_cnt + 1

END FUNCTION 

############################################################
#+ @code
#+ 函式目的 濾除ent
#+ @param  ps_temp STRING 欄位字串
#+
#+ @return STRING  去除ent欄位字串
############################################################
PUBLIC FUNCTION sadzp030_tab_relation_ent(ps_temp)

   DEFINE ps_temp    STRING
   DEFINE ls_token   STRING
   DEFINE ls_return  STRING
   DEFINE tok        base.StringTokenizer

   LET tok = base.StringTokenizer.create(ps_temp,",")
   WHILE tok.hasMoreTokens()
      LET ls_token = tok.nextToken()
      IF ls_token.subString(ls_token.getLength()-2,ls_token.getLength()) = "ent" THEN
         CONTINUE WHILE
      ELSE
         LET ls_return = ls_return,",",ls_token
      END IF
   END WHILE
   RETURN ls_return.subString(2,ls_return.getLength())
END FUNCTION

############################################################
#+ @code
#+ 函式目的 濾除site
#+ @param  ps_temp STRING 欄位字串
#+
#+ @return STRING  去除site欄位字串
############################################################
PUBLIC FUNCTION sadzp030_tab_relation_site(ps_temp)

   DEFINE ps_temp    STRING
   DEFINE ls_token   STRING
   DEFINE ls_return  STRING
   DEFINE tok        base.StringTokenizer

   LET tok = base.StringTokenizer.create(ps_temp,",")
   WHILE tok.hasMoreTokens()
      LET ls_token = tok.nextToken()
      IF ls_token.subString(ls_token.getLength()-3,ls_token.getLength()) = "site" THEN
         CONTINUE WHILE
      ELSE
         LET ls_return = ls_return,",",ls_token
      END IF
   END WHILE
   RETURN ls_return.subString(2,ls_return.getLength())
END FUNCTION

############################################################
#+ @code
#+ 函式目的 取得4gl程式樣板
#+ @param  ls_type STRING 作業型態
#+
#+ @return STRING  4gl程式樣板
############################################################
PUBLIC FUNCTION sadzp030_tab_relation_prog_type(ls_type)

   DEFINE ls_type    STRING 
   DEFINE lc_dzfqcrtid LIKE dzfq_t.dzfqcrtid  #創建人id
   DEFINE lc_dzfqcrtdt LIKE dzfq_t.dzfqcrtdt  #創建作業日期
   DEFINE lc_dzax007   LIKE dzax_t.dzax007    #預支固定參數個數

   CALL sadzp030_tab_relation_prog_type_new() 
   RETURNING ls_type,lc_dzfqcrtid,lc_dzfqcrtdt,lc_dzax007

   IF lc_dzax007 IS NULL THEN LET lc_dzax007 = "0" END IF

   RETURN ls_type,lc_dzfqcrtid,lc_dzfqcrtdt,lc_dzax007

END FUNCTION


#+ 程式採用何種4gl樣板檢查流程

PUBLIC FUNCTION sadzp030_tab_relation_prog_type_new() 
   DEFINE ls_type      STRING
   DEFINE li_cnt       LIKE type_t.num5
   DEFINE lc_dzfqcrtid LIKE dzfq_t.dzfqcrtid  #創建人id
   DEFINE lc_dzfqcrtdt LIKE dzfq_t.dzfqcrtdt  #創建作業日期
   DEFINE lc_ms_type   LIKE type_t.chr1
   DEFINE lc_op_type   LIKE type_t.chr1
   DEFINE lc_freestyle LIKE type_t.chr1
   DEFINE li_grid_cnt, li_sr_cnt LIKE type_t.num5
   DEFINE lc_gzde006   LIKE gzde_t.gzde006
   DEFINE li_dzaf      LIKE type_t.num5
   DEFINE lc_dzax007   LIKE dzax_t.dzax007    #預支固定參數個數

   #取創建人資料
   SELECT dzfqcrtid,dzfqcrtdt INTO lc_dzfqcrtid,lc_dzfqcrtdt
     FROM dzfq_t
    WHERE dzfq003 = "1"  #g_sd_ver  #識別碼版號,因為dzfq只有r.a會寫入,所以咬死1
      AND dzfq004 = g_prog_id #畫面代號

   SELECT COUNT(1) INTO li_dzaf FROM dzaf_t 
    WHERE dzaf001 = g_prog_id AND dzaf005 = "M"
   IF li_dzaf < 1 THEN
      CASE
         #沒有畫面的就取子程式或WEB
         WHEN g_prog_id[1,3] = "wss" OR g_prog_id[1,4] = "cwss" 
            SELECT gzjacrtid,gzjacrtdt INTO lc_dzfqcrtid,lc_dzfqcrtdt
              FROM gzja_t
             WHERE gzja001 = g_prog_id

         #開窗查詢的Q類
         WHEN g_prog_id[1,2] = "q_" OR g_prog_id[1,3] = "cq_" 
            SELECT dzcacrtid,dzcacrtdt INTO lc_dzfqcrtid,lc_dzfqcrtdt
              FROM dzca_t
             WHERE dzca001 = g_prog_id
               AND dzca006 = "Y"
               AND dzca002 = "c"     #標準轉客製後，s複製一份c,會有2筆資料
            IF SQLCA.SQLCODE THEN
               SELECT dzcacrtid,dzcacrtdt INTO lc_dzfqcrtid,lc_dzfqcrtdt
                 FROM dzca_t
                WHERE dzca001 = g_prog_id
                  AND dzca006 = "Y"
                  AND dzca002 = "s"
            END IF

         OTHERWISE
            SELECT gzdecrtid,gzdecrtdt INTO lc_dzfqcrtid,lc_dzfqcrtdt
              FROM gzde_t
             WHERE gzde001 = g_prog_id
      END CASE
      IF SQLCA.SQLCODE = NOTFOUND THEN
         DISPLAY cl_getmsg_parm("adz-01132",g_lang,SQLCA.SQLCODE)
        #DISPLAY "注意: 規格資料不存在,或發生其他錯誤 (技術代碼:",SQLCA.SQLCODE,")"
         RETURN ls_type,lc_dzfqcrtid,lc_dzfqcrtdt,lc_dzax007
      END IF
   END IF

   #確認是主程式還是子程式/FreeStyle/樣板類型/子程式類型
   CALL sadzp030_tab_relation_chk_mainsub()
      RETURNING lc_ms_type,lc_freestyle,lc_op_type,lc_gzde006,lc_dzax007
   IF g_t100debug >= 6 THEN
      DISPLAY "INFO: 主/子程式:",lc_ms_type," Freestyle:",lc_freestyle," 樣板:",lc_op_type," 子程式類型:",lc_gzde006
   END IF

   #先把web service請出去
   IF lc_ms_type = "Z" THEN
      LET ls_type = "m00"
      RETURN ls_type,lc_dzfqcrtid,lc_dzfqcrtdt,lc_dzax007
   END IF

   #web service元件
   IF lc_ms_type = "W" THEN
      LET ls_type = "p00"
      RETURN ls_type,lc_dzfqcrtid,lc_dzfqcrtdt,lc_dzax007
   END IF

   #把報表元件請出去
   IF lc_ms_type = "X" OR lc_ms_type = "G" THEN
      LET ls_type = "x00"
      RETURN ls_type,lc_dzfqcrtid,lc_dzfqcrtdt,lc_dzax007
   END IF

   #把free style請出去
   IF lc_freestyle = "Y" THEN  #FreeStyle時此指標為N
      IF lc_ms_type = "M" THEN
         LET ls_type = "i00"
      ELSE
         IF lc_ms_type = "Q" THEN    #開窗查詢元件
            LET ls_type = "p00"      #開窗查詢也改成使用 p00 樣板
         ELSE
            LET ls_type = "p00"
         END IF
      END IF
      #呼叫檢查是否已經轉為free-style的add-point
      CALL sadzp030_tab_relation_chk_freestyle(lc_ms_type)
      DISPLAY cl_getmsg("adz-01133",g_lang)
     #DISPLAY "INFO: 本程式已在dzax_t內被設定為FreeStyle"

      RETURN ls_type,lc_dzfqcrtid,lc_dzfqcrtdt,lc_dzax007
   END IF

   CASE
      WHEN lc_ms_type = "M"  #主程式類別
         #程式類別區分
         IF lc_op_type = "Q" THEN  #Q類
            LET ls_type = sadzp030_relation_has_qbe()   #抓出qbe這個Vbox/Hbox
            RETURN ls_type,lc_dzfqcrtid,lc_dzfqcrtdt,lc_dzax007
         END IF

         IF lc_op_type = "P" THEN  #P01類
            LET ls_type = "p01"
            RETURN ls_type,lc_dzfqcrtid,lc_dzfqcrtdt,lc_dzax007
         END IF

         IF lc_op_type = "R" THEN  #R類
            LET ls_type = "r01"
            RETURN ls_type,lc_dzfqcrtid,lc_dzfqcrtdt,lc_dzax007
         END IF

         IF lc_op_type = "W" THEN  #P02類  業務待選擇類
            LET ls_type = "p02"
            RETURN ls_type,lc_dzfqcrtid,lc_dzfqcrtdt,lc_dzax007
         END IF

         IF lc_op_type = "F" THEN  #F類
            #查看table形式,查看是否有Grid存在
            SELECT COUNT(1) INTO li_grid_cnt FROM dzaa_t,dzag_t
             WHERE dzaa001 = g_prog_id  #程式編號
               AND dzaa002 = g_sd_ver   #規格版次
               AND dzaa003 = "TABLE"    #識別碼
               AND dzaa005 = "4"        #識別碼編號
#              AND dzaa008 = g_erp_ver  #產品版本
               AND dzaa009 = g_dgenv    #客製標示
#              AND dzaa010 = g_cust     #客戶編號
               AND dzaastus = "Y"       #狀態碼 (取有效)
           #-----------------------dzag---------------------#
               AND dzag001 = dzaa001
               AND dzag003 = dzaa004
               AND dzag005 = "Y"        #是否為主表
               AND dzag006 = dzaa006    #使用標示
               AND dzag007 = "Y"        #是否被放在單頭-如果存在任何一筆,表示有單頭
#              AND dzag011 = dzaa010    #客戶編號
               AND dzagstus = "Y"

            IF li_grid_cnt > 0 THEN #可能是單檔/雙檔/假雙檔/(單身凍結)
               IF g_t100debug >= 6 THEN
                  DISPLAY "INFO: 從dzag資料查出有單頭存在 count>0"
               END IF

               SELECT COUNT(1) INTO li_sr_cnt FROM dzaa_t,dzfs_t
                WHERE dzaa001 = g_prog_id
                  AND dzaa002 = g_sd_ver          #規格版次
                  AND dzaa003 = "TABLE"           #串dzag_t 的識別碼
                  AND dzaa005 = "4"               #串dzag_t 的識別碼編號
#                 AND dzaa008 = g_erp_ver         #產品版本
                  AND dzaa009 = g_dgenv           #客製標示
#                 AND dzaa010 = g_cust            #客戶編號
                  AND dzaastus = "Y"              #只取有效的
             #-----------------------dzfs------------------------------#
                  AND dzfs001 = dzaa004           #串識別碼版號
                  AND dzfs002 = dzaa001           #串規格編號
                  AND dzfs003 <> "s_browse"       #不要s_browse的
                  AND dzfs005 = dzaa006           #使用標示
#                 AND dzfs011 = dzaa010           #客戶編號
                  AND dzfsstus = "Y"              #只取有效的
                  AND dzfs003 NOT IN (
                      SELECT dzam003 FROM dzaa_t,dzam_t
                       WHERE dzaa001 = g_prog_id   #規格編號
                         AND dzaa002 = g_sd_ver    #規格版次
                         AND dzaa003 = "EXCLUDE"   #識別代碼
#                        AND dzaa008 = g_erp_ver   #產品版本
                         AND dzaa009 = g_dgenv     #客製標示
#                        AND dzaa010 = g_cust      #客戶編號
                         AND dzaastus = "Y"
                      #------------------------------------------------------#
                         AND dzam001 = dzaa001     #規格編號
                         AND dzam004 = dzaa004     #識別碼版次
                         AND dzam005 = dzaa006     #使用標示
#                        AND dzam006 = dzaa010     #客戶編號
                         AND dzamstus = "Y" )      #有效碼

               IF li_sr_cnt = 0 THEN
                  IF g_t100debug >= 6 THEN
                     DISPLAY "INFO: 從表格與排除資料查出沒有單身 count=0"
                  END IF

                  CASE sadzp030_relation_has_browse()
                     WHEN "Table"
                        IF g_t100debug >= 6 THEN
                           DISPLAY "INFO: 取得browser且為table形式-i01"
                        END IF
                        LET ls_type = "i01"  #有查詢計畫的單檔
                     WHEN "Tree"
                        IF g_t100debug >= 6 THEN
                           DISPLAY "INFO: 取得browser且為tree形式,歸入i13或i05"
                        END IF
                        IF sadzp030_tab_relation_is_i13() THEN
                           LET ls_type = "i13"   #六階樹狀單檔
                        ELSE
                           LET ls_type = "i05"   #左右排/上下排的樹狀單檔
                        END IF
                     OTHERWISE
                        LET ls_type = "i10"  #純單檔
                  END CASE
               ELSE
                  IF g_t100debug >= 6 THEN
                     DISPLAY "INFO: 從表格與排除資料查到單身 count>0"
                  END IF
                  CASE sadzp030_relation_has_browse()
                     WHEN "Table"
                        IF g_t100debug >= 6 THEN
                           DISPLAY "INFO: Browser是Table形式,歸入i07或t01"
                        END IF
                        IF sadzp030_tab_relation_is_i07("a") THEN
                           LET ls_type = "i07"   #假雙檔
                        ELSE
                           LET ls_type = "t01"   #雙檔
                        END IF
                     WHEN "Tree"
                        IF g_t100debug >= 6 THEN
                           DISPLAY "INFO: Browser是Tree形式,歸入i04或i08"
                        END IF
                        IF sadzp030_tab_relation_is_i04() THEN
                           LET ls_type = "i04"   #i04 樹狀雙檔(主從表)
                        ELSE
                           LET ls_type = "i08"   #i08 樹狀單檔+單身
                        END IF
                     OTHERWISE
                        IF g_t100debug >= 6 THEN
                           DISPLAY "INFO: 沒有Browser,歸入i12(i07)或i09"
                        END IF
                        IF sadzp030_tab_relation_is_i07("a") THEN
                           LET ls_type = "i12"   #沒有查詢計畫的假雙檔
                        ELSE
                           LET ls_type = "i09"   #沒有查詢計畫的雙檔
                        END IF
                  END CASE
               END IF

            ELSE #可能是單檔多欄/雙檔多欄
               IF g_t100debug >= 6 THEN
                  DISPLAY "INFO: 從dzag資料查不到單頭 count=0,歸入i02(i07)或t02"
               END IF
               IF sadzp030_tab_relation_is_i07("b") THEN
                  LET ls_type = "i02"        #單檔多欄
               ELSE
                  LET ls_type = "t02"        #雙檔多欄
               END IF
            END IF

            RETURN ls_type,lc_dzfqcrtid,lc_dzfqcrtdt,lc_dzax007
         END IF
         DISPLAY cl_getmsg_parm("adz-01134",g_lang,g_prog_type)
        #DISPLAY "ERROR: 判斷4fd後,確認畫面型態有問題:",g_prog_type
         LET g_cross_4gl_stus = FALSE     #跨4gl判斷正確與否

      WHEN lc_ms_type = "S"  #子程式類別
         #查看table形式,查看是否有Grid存在
         SELECT COUNT(1) INTO li_grid_cnt FROM dzaa_t,dzag_t
          WHERE dzaa001 = g_prog_id  #程式編號
            AND dzaa002 = g_sd_ver   #規格版次
            AND dzaa003 = "TABLE"    #識別碼
            AND dzaa005 = "4"        #識別碼編號
#           AND dzaa008 = g_erp_ver  #產品版本
            AND dzaa009 = g_dgenv    #客製標示
#           AND dzaa010 = g_cust     #客戶編號
            AND dzaastus = "Y"       #狀態碼 (取有效)
         #-----------------------dzag---------------------#
            AND dzag001 = dzaa001
            AND dzag003 = dzaa004
            AND dzag005 = "Y"        #是否為主表
            AND dzag006 = dzaa006    #使用標示
            AND dzag007 = "Y"        #是否被放在單頭-如果存在任何一筆,表示有單頭
#           AND dzag011 = dzaa010    #客戶編號
            AND dzagstus = "Y"

         IF li_grid_cnt > 0 THEN #可能是單檔/雙檔/假雙檔/(單身凍結)
            IF g_t100debug >= 6 THEN
               DISPLAY "INFO: 從dzag資料查出有單頭存在 count>0"
            END IF

            SELECT COUNT(1) INTO li_sr_cnt FROM dzaa_t,dzfs_t
             WHERE dzaa001 = g_prog_id
               AND dzaa002 = g_sd_ver          #規格版次
               AND dzaa003 = "TABLE"           #串dzag_t 的識別碼
               AND dzaa005 = "4"               #串dzag_t 的識別碼編號
#              AND dzaa008 = g_erp_ver         #產品版本
               AND dzaa009 = g_dgenv           #客製標示
#              AND dzaa010 = g_cust            #客戶編號
               AND dzaastus = "Y"              #只取有效的
          #-----------------------dzfs------------------------------#
               AND dzfs001 = dzaa004           #串識別碼版號
               AND dzfs002 = dzaa001           #串規格編號
               AND dzfs003 <> "s_browse"       #不要s_browse的
               AND dzfs005 = dzaa006           #使用標示
#              AND dzfs011 = dzaa010           #客戶編號
               AND dzfsstus = "Y"              #只取有效的
               AND dzfs003 NOT IN (
                   SELECT dzam003 FROM dzaa_t,dzam_t
                    WHERE dzaa001 = g_prog_id   #規格編號
                      AND dzaa002 = g_sd_ver    #規格版次
                      AND dzaa003 = "EXCLUDE"   #識別代碼
#                     AND dzaa008 = g_erp_ver         #產品版本
                      AND dzaa009 = g_dgenv           #客製標示
#                     AND dzaa010 = g_cust            #客戶編號
                      AND dzaastus = "Y"
                   #------------------------------------------------------#
                      AND dzam001 = dzaa001     #規格編號
                      AND dzam004 = dzaa004     #識別碼版次
                      AND dzam005 = dzaa006     #使用標示
#                     AND dzam006 = dzaa010     #客戶編號
                      AND dzamstus = "Y" )      #有效碼 

            IF li_sr_cnt = 0 THEN
               LET ls_type = "c01"   #單檔子程式
            ELSE
               LET ls_type = "c03"   #雙檔子程式
            END IF
         ELSE
            IF g_t100debug >= 6 THEN
               DISPLAY "INFO: 從表格與排除資料查到單身 count>0"
            END IF
            IF sadzp030_tab_relation_is_i07("b") THEN
               LET ls_type = "c02"   #單檔多欄子程式
            ELSE
               IF lc_gzde006 = "a" THEN
                  LET ls_type = "c04"   #雙檔多欄子程式
               ELSE
                  LET ls_type = "c02"   #雙檔多欄子程式
               END IF
            END IF
         END IF
         LET ls_type = ls_type,lc_gzde006

      WHEN lc_ms_type = "G"  #GR報表程式
         LET ls_type = "g01"

      WHEN lc_ms_type = "X"  #Xgrid報表程式
         LET ls_type = "x01"
   END CASE

   RETURN ls_type,lc_dzfqcrtid,lc_dzfqcrtdt,lc_dzax007

END FUNCTION 


PRIVATE FUNCTION sadzp030_tab_relation_chk_freestyle(lc_ms_type)

   DEFINE ls_sys      STRING
   DEFINE ls_path     STRING
   DEFINE lc_ms_type  LIKE type_t.chr1
   DEFINE li_cnt      LIKE type_t.num5

   #排除LIB/SUB
   IF g_prog_id[1,3] = "cl_" OR g_prog_id[1,4] = "ccl_" OR
      g_prog_id[1,2] = "s_" OR g_prog_id[1,3] = "cs_" THEN
   ELSE
      #檢查目錄下是否有4gl檔
      LET ls_sys = sadzp030_tab_file_module(g_prog_id,"A",g_dgenv)
      LET ls_path = os.Path.join(os.Path.join(ls_sys,"4gl"),g_prog_id),".4gl"

      IF lc_ms_type = "M" THEN
         DISPLAY "INFO: 目錄下存在檔案,轉為i00樣板,Path=",ls_path
      ELSE
         DISPLAY "INFO: 目錄下存在檔案,轉為p00樣板,Path=",ls_path
      END IF

      #在此處檢查一下i00所需要的add-point是否已經存在
      IF g_stop_chk_sd_ver_freestyle IS NULL THEN 
         LET g_stop_chk_sd_ver_freestyle = FALSE  #沒寫當FALSE (要查)
      END IF

      IF g_stop_chk_sd_ver_freestyle THEN
         SELECT COUNT(1) INTO li_cnt FROM dzbb_t
          WHERE dzbb001 = g_prog_id         #規格/程式編號
            AND dzbb002 = "free_style.variable"   #程式設計點/add-point id
#           AND dzbb003 = g_sd_ver          #不查! 標示為TRUE
#           AND dzbb005 = g_cust            #客戶編號
      ELSE
         SELECT COUNT(1) INTO li_cnt FROM dzbb_t
          WHERE dzbb001 = g_prog_id         #規格/程式編號
            AND dzbb002 = "free_style.variable"   #程式設計點/add-point id
#            AND dzbb003 = g_sd_ver          #規格版次/程式版次 #190509-00014 add mark
#           AND dzbb005 = g_cust            #客戶編號
         IF g_t100debug >= 3 THEN
            DISPLAY "INFO: 系統走入檢查free_style.variable的檢查"
         END IF
      END IF
      #如果沒有就呼叫自動塞入的工具
      IF li_cnt = 0 THEN
         IF NOT sadzp180_free(g_sd_ver,g_prog_id,lc_ms_type) THEN
            DISPLAY "注意: 將程式",g_prog_id,"轉換為add-point程序失敗"
         END IF
      ELSE
         IF g_t100debug >= 3 THEN
            DISPLAY "INFO: 系統有存在",g_prog_id,"轉換後的元件,不進行覆蓋"
         END IF
      END IF 
   END IF 

END FUNCTION


#+ 檢查程式類別(M:主程式/子程式/lib/sub/qry)與是否為FreeStyle/樣板類型/子程式類型
PRIVATE FUNCTION sadzp030_tab_relation_chk_mainsub() 

   DEFINE li_cnt     LIKE type_t.num5
   DEFINE li_gzja    LIKE type_t.num5
   DEFINE lc_dzax002 LIKE dzax_t.dzax002  #程式類別
   DEFINE lc_dzax003 LIKE dzax_t.dzax003  #FreeStyle指標
   DEFINE lc_dzax007 LIKE dzax_t.dzax007  #預支固定參數數量
   DEFINE lc_gzde003 LIKE gzde_t.gzde003  #規格類別
   DEFINE lc_gzde006 LIKE gzde_t.gzde006  #程式產生類型

   LET lc_dzax003 = "N"

   #查主程式註冊資料
   SELECT COUNT(1) INTO li_cnt FROM gzza_t
    WHERE gzza001 = g_prog_id
   
   CASE
      WHEN li_cnt = 0     #找不到--子程式

         #查詢Q類的額外處理
         IF g_prog_id[1,2] = "q_" OR g_prog_id[1,3] = "cq_" THEN
            SELECT dzca006 INTO lc_dzax003 FROM dzca_t
             WHERE dzca001 = g_prog_id
               AND dzca002 = "c"
            IF SQLCA.SQLCODE THEN
               SELECT dzca006 INTO lc_dzax003 FROM dzca_t
                WHERE dzca001 = g_prog_id
                  AND dzca002 = "s"
            END IF
            LET lc_gzde003 = "Q"  #開窗程式
         ELSE
            #查子程式註冊資料
            SELECT gzde003,gzde006
              INTO lc_gzde003,lc_gzde006
              FROM gzde_t
             WHERE gzde001 = g_prog_id #規格/畫面編號
            IF SQLCA.SQLCODE = NOTFOUND THEN
               #選不到時,檢查一下是否為web service作業
               SELECT COUNT(1) INTO li_gzja FROM gzja_t
                WHERE gzja001 = g_prog_id #規格/畫面編號
                  AND gzjastus = "Y"
               IF li_gzja > 0 THEN
                  LET lc_gzde003 = "Z"  #web service程式
                  LET lc_dzax003 = "Y"  #FREE STYLE
               END IF
            END IF

            #規格類別=B(lib or sub)/W(網頁)/Q(開窗qry) 就應該是要FreeStyle
            IF lc_gzde003 = "B" OR lc_gzde003 = "W" OR lc_gzde003 = "Q" THEN
               LET lc_dzax003 = "Y"   #FREE STYLE
            END IF
         END IF

      WHEN li_cnt = 1   #未發生錯誤--主程式
         LET lc_gzde003 = "M"

      OTHERWISE         #發生其他錯誤
         LET lc_gzde003 = ""
   END CASE

   #FREE STYLE的和WEB的都拉掉
   IF lc_dzax003 = "N" THEN
      SELECT dzax002,dzax003,dzax007 INTO lc_dzax002,lc_dzax003,lc_dzax007
        FROM dzax_t
       WHERE dzax001 = g_prog_id
#        AND dzax005 = g_cust      #客戶編號
         AND dzax006 = g_dgenv     #客製標示
         AND dzaxstus = "Y"
      IF SQLCA.SQLCODE THEN
         IF g_t100debug >= 3 THEN
            DISPLAY "注意: dzax資料不存在! dzax001=",g_prog_id," dzax006=",g_dgenv," stus='Y' "
         END IF
      END IF
   END IF
          #程式類別/FreeStyle/樣板類型/子程式類型/預支固定參數數量
   RETURN lc_gzde003,lc_dzax003,lc_dzax002,lc_gzde006,lc_dzax007
END FUNCTION 

#+ 查詢是否為q01/q03/q04 或是 q02(不存在時為q02)
PRIVATE FUNCTION sadzp030_relation_has_qbe() 
   DEFINE li_cnt     LIKE type_t.num5
   DEFINE lc_type    LIKE type_t.chr3
   DEFINE lc_dzfi007 LIKE dzfi_t.dzfi007

   #區分q02 與其他(暫記為q01)
   SELECT COUNT(1) INTO li_cnt FROM dzfi_t
    WHERE dzfi001 = g_prog_id                     #程式編號
      AND dzfi002 = g_sd_ver                      #規格版次
      AND dzfi009 = g_dgenv                       #客製標示
      AND dzfi006 = "qbe"                         #元件(組)代碼
      AND (dzfi007 = "VBox" OR dzfi007 = "HBox")  #節點類型
#     AND dzfi017 = g_cust                        #客戶編號
   IF li_cnt > 0 THEN
      IF g_t100debug >= 6 THEN 
         DISPLAY "INFO: 抓到qbe的VBox or HBox"
      END IF
      LET lc_type = "q01"
   ELSE
      LET lc_type = "q02"
   END IF

   #區分為 q04與其他(暫記為q01)
   IF lc_type = "q01" THEN
      SELECT COUNT(1) INTO li_cnt FROM dzfi_t
       WHERE dzfi001 = g_prog_id                     #程式編號
         AND dzfi002 = g_sd_ver                      #規格版次
         AND dzfi009 = g_dgenv                       #客製標示
         AND dzfi006 = "vb_master"
#        AND dzfi017 = g_cust   #客戶編號
      IF li_cnt > 0 THEN
         IF g_t100debug >= 6 THEN 
            DISPLAY "INFO: 抓到Q類右側的vb_master"
         END IF
         LET lc_type = "q04"
      ELSE
         LET lc_type = "q01"
      END IF
   END IF

   #區分為 q03與q01
   IF lc_type = "q01" THEN
      SELECT dzfi007 INTO lc_dzfi007 FROM dzfi_t
       WHERE dzfi001 = g_prog_id                     #程式編號
         AND dzfi002 = g_sd_ver                      #規格版次
         AND dzfi009 = g_dgenv                       #客製標示
         AND dzfi006 = "s_detail1"
#        AND dzfi017 = g_cust   #客戶編號
      IF NOT SQLCA.SQLCODE AND lc_dzfi007 = "Tree" THEN
         IF g_t100debug >= 6 THEN 
            DISPLAY "INFO: 抓到Q類右側的 Tree"
         END IF
         LET lc_type = "q03"
      ELSE
         LET lc_type = "q01"
      END IF
   END IF

   RETURN lc_type
END FUNCTION 
#+
PRIVATE FUNCTION sadzp030_relation_has_browse() 
   DEFINE lc_dzfs010 LIKE dzfs_t.dzfs010
display 'g_sd_ver:',g_sd_ver
   SELECT dzfs010 INTO lc_dzfs010 FROM dzaa_t,dzfs_t
    WHERE dzaa001 = g_prog_id
      AND dzaa002 = g_sd_ver          #規格版次
      AND dzaa003 = "TABLE"           #串dzag_t 的識別碼
      AND dzaa005 = "4"               #串dzag_t 的識別碼編號
#     AND dzaa008 = g_erp_ver         #產品版本
      AND dzaa009 = g_dgenv           #客製標示
#     AND dzaa010 = g_cust            #客戶編號
      AND dzaastus = "Y"              #只取有效的
   #----------------------dzfs------------------------------#
      AND dzfs001 = dzaa004           #串識別碼版號
      AND dzfs002 = dzaa001           #串規格編號
      AND dzfs003 = "s_browse"        #s_browse
      AND dzfs005 = dzaa006           #使用標示
#     AND dzfs011 = dzaa010           #客戶編號
      AND dzfsstus = "Y"              #只取有效的
   IF SQLCA.SQLCODE = 100 THEN
      LET lc_dzfs010 = "N"
   ELSE
      IF g_t100debug >= 6 THEN 
         DISPLAY "INFO: 抓到s_browse,且型態為",lc_dzfs010
      END IF
   END IF

   RETURN lc_dzfs010
END FUNCTION 
############################################################
#+ @code
#+ 函式目的 F003_sc
#  i07假雙檔 所有的 dzag004 IS NULL, t01一定存在dzag004 IS NOT NULL
#  dzag004 單頭主表資料
#+ @param  
#+ @return BOOLEAN  TRUE/FALSE
############################################################
PRIVATE FUNCTION sadzp030_tab_relation_is_i13() 

   DEFINE ls_sql        STRING 
   DEFINE l_dzff006     LIKE dzff_t.dzff006
   DEFINE l_dzag004     LIKE dzag_t.dzag004
   DEFINE l_dzff005     LIKE dzff_t.dzff005
   DEFINE li_cnt        LIKE type_t.num5 
   DEFINE li_rtn        LIKE type_t.num5 

   LET ls_sql = "SELECT dzff005,dzff006 ",
                 " FROM dzaa_t,dzff_t", 
             #-----------------------------dzaa-------------#
                " WHERE dzaa001 = '",g_prog_id CLIPPED,"'",  #規格編號
                  " AND dzaa002 = '",g_sd_ver CLIPPED,"'",   #規格版次
                  " AND dzaa003 = 's_browse' ",              #設計點/識別碼
                  " AND dzaa005 = '5'",                      #識別碼類型
#                 " AND dzaa008 = '",g_erp_ver CLIPPED,"' ", #產品版本
                  " AND dzaa009 = '",g_dgenv CLIPPED,"' ",   #客製標示
#                 " AND dzaa010 = '",g_cust CLIPPED,"' ",    #客戶編號
                  " AND dzaastus = 'Y'",
             #-----------------------------dzff-------------#
                  " AND dzff001 = dzaa001 ",        #規格編號
                  " AND dzff002 = dzaa004 ",        #識別碼版次
                  " AND dzff003 = dzaa003 ",        #4fd tag name
                  " AND dzff008 = dzaa006 ",        #使用標示
#                 " AND dzff009 = dzaa010 ",        #客戶編號
                  " AND dzffstus = 'Y'"
   PREPARE sadzp030_i13_pre FROM ls_sql
   DECLARE sadzp030_i13_pre_curs CURSOR FOR sadzp030_i13_pre

   LET li_rtn = FALSE
   FOREACH sadzp030_i13_pre_curs INTO l_dzff005,l_dzff006
      IF l_dzff005 = "type2" THEN
         LET li_rtn = TRUE
      END IF
   END FOREACH 

   FREE sadzp030_i13_pre_curs
   RETURN li_rtn

END FUNCTION


############################################################
#+ @code
#+ 函式目的 F003_sc
#  i07假雙檔 所有的 dzag004 IS NULL, t01一定存在dzag004 IS NOT NULL
#  dzag004 單頭主表資料
#+ @param  
#+ @return BOOLEAN  TRUE/FALSE
############################################################
PRIVATE FUNCTION sadzp030_tab_relation_is_i07(lc_type)

   DEFINE li_cnt         LIKE type_t.num5
   DEFINE li_rtn         LIKE type_t.num5
   DEFINE ls_sql         STRING
   DEFINE lc_dzag004     LIKE dzag_t.dzag004
   DEFINE lc_type        LIKE type_t.chr1
   DEFINE ls_t           STRING             #用在假雙檔加單身特別處理
   DEFINE li_detailcnt   LIKE type_t.num5   #用在假雙檔加單身查詢筆數

   #i07假雙檔 所有的 dzag004 IS NULL, t01一定存在dzag004 IS NOT NULL(存在上層表)
   LET ls_sql = "SELECT COUNT(1) ",
                 " FROM dzaa_t,dzag_t ", 
           #--------------------------dzaa---------------------#
                " WHERE dzaa001 = '",g_prog_id CLIPPED,"'",  #規格編號
                  " AND dzaa002 = '",g_sd_ver CLIPPED,"'",   #規格版次
                  " AND dzaa003 = 'TABLE'",                  #識別碼
                  " AND dzaa005 = '4'",                      #識別碼編號
#                 " AND dzaa008 = '",g_erp_ver CLIPPED,"' ", #產品版本
                  " AND dzaa009 = '",g_dgenv CLIPPED,"' ",   #客製標示
#                 " AND dzaa010 = '",g_cust CLIPPED,"' ",    #客戶編號
                  " AND dzaastus = 'Y'",             #狀態碼取 Y 有效
           #--------------------------dzag---------------------#
                  " AND dzag001 = dzaa001 ",         #規格編號
                  " AND dzag003 = dzaa004 ",         #識別碼版次
                  " AND dzag004 IS NOT NULL ",
                  " AND dzag006 = dzaa006 ",         #使用標示
                  " AND dzag007 = 'N' ",             #是否為單頭 (N)
#                 " AND dzag011 = dzaa010 ",         #客戶編號
                  " AND dzagstus = 'Y' ",            #狀態碼取 Y 有效
                  " AND dzag002 IN ( ",
                       "SELECT dzfs004 FROM dzaa_t,dzfs_t ",
                       " WHERE dzaa001 = '",g_prog_id CLIPPED,"' ",
                         " AND dzaa002 = '",g_sd_ver CLIPPED,"' ",  #規格版次
                         " AND dzaa003 = 'TABLE' ",        #串dzag_t 的識別碼
                         " AND dzaa005 = '4' ",            #串dzag_t 的識別碼編號
#                        " AND dzaa008 = '",g_erp_ver CLIPPED,"' ", #產品版本
                         " AND dzaa009 = '",g_dgenv CLIPPED,"' ",   #客製標示
#                        " AND dzaa010 = '",g_cust CLIPPED,"' ",    #客戶編號
                         " AND dzaastus = 'Y' ",           #只取有效的
                      #-----------------------dzfs------------------------------#
                         " AND dzfs001 = dzaa004 ",        #串識別碼版號
                         " AND dzfs002 = dzaa001 ",        #串規格編號
                         " AND dzfs003 <> 's_browse' ",    #不要s_browse的
                         " AND dzfs005 = dzaa006 ",        #使用標示
#                        " AND dzfs011 = dzaa010 ",        #客戶編號
                         " AND dzfsstus = 'Y' ",           #只取有效的
                         " AND dzfs003 NOT IN ( ",
                             " SELECT dzam003 FROM dzaa_t,dzam_t ",
                              " WHERE dzaa001 = '",g_prog_id CLIPPED,"' ",  #規格編號
                                " AND dzaa002 = '",g_sd_ver CLIPPED,"' ",   #規格版次
                                " AND dzaa003 = 'EXCLUDE' ",                #識別代碼
                                " AND dzaa005 = 'a' ",                      #識別碼類型
#                               " AND dzaa008 = '",g_erp_ver CLIPPED,"' ",  #產品版本
                                " AND dzaa009 = '",g_dgenv CLIPPED,"' ",    #客製標示
#                               " AND dzaa010 = '",g_cust CLIPPED,"' ",     #客戶編號
                                " AND dzaastus = 'Y' ",
                               #------------------------------------------------------#
                                " AND dzam001 = dzaa001 ",  #規格編號
                                " AND dzam004 = dzaa004 ",  #識別碼版次
                                " AND dzam005 = dzaa006 ",  #使用標示
#                               " AND dzam006 = dzaa010 ",  #客戶編號
                                " AND dzamstus = 'Y' ) ) "  #有效碼

   PREPARE sadzp030_prog_type_pre FROM ls_sql
   EXECUTE sadzp030_prog_type_pre INTO li_cnt
   FREE sadzp030_prog_type_pre

   IF li_cnt > 0 THEN 
      LET li_rtn = FALSE  #找到有上階單頭資料,非純假雙檔-> t01

      IF lc_type = "a" THEN
         #此處要排除假雙檔加單身 (若上階單頭資料和單身中某一張表相等,則秀出關係)
         LET ls_sql = "SELECT UNIQUE dzag002 ",
                       " FROM dzaa_t,dzag_t ", 
                 #--------------------------dzaa---------------------#
                      " WHERE dzaa001 = '",g_prog_id CLIPPED,"'",  #規格編號
                        " AND dzaa002 = '",g_sd_ver CLIPPED,"'",   #規格版次
                        " AND dzaa003 = 'TABLE'",                  #識別碼
                        " AND dzaa005 = '4'",                      #識別碼編號
#                       " AND dzaa008 = '",g_erp_ver CLIPPED,"' ", #產品版本
                        " AND dzaa009 = '",g_dgenv CLIPPED,"' ",   #客製標示
#                       " AND dzaa010 = '",g_cust CLIPPED,"' ",    #客戶編號
                        " AND dzaastus = 'Y'",             #狀態碼取 Y 有效
                 #--------------------------dzag---------------------#
                        " AND dzag001 = dzaa001 ",         #規格編號
                        " AND dzag003 = dzaa004 ",         #識別碼版次
                        " AND dzag006 = dzaa006 ",         #使用標示
                        " AND dzag007 = 'Y' ",             #是否為單頭 (Y)
#                       " AND dzag011 = dzaa010 ",         #客戶編號
                        " AND dzagstus = 'Y' "             #狀態碼取 Y 有效
         DECLARE sadzp030_prog_type_data_pre CURSOR FROM ls_sql

         LET ls_t = "SELECT COUNT(1) FROM dzaa_t,dzfs_t ",
                    " WHERE dzaa001 = '",g_prog_id CLIPPED,"' ",
                      " AND dzaa002 = '",g_sd_ver CLIPPED,"' ",  #規格版次
                      " AND dzaa003 = 'TABLE' ",        #串dzag_t 的識別碼
                      " AND dzaa005 = '4' ",            #串dzag_t 的識別碼編號
#                     " AND dzaa008 = '",g_erp_ver CLIPPED,"' ", #產品版本
                      " AND dzaa009 = '",g_dgenv CLIPPED,"' ",   #客製標示
#                     " AND dzaa010 = '",g_cust CLIPPED,"' ",    #客戶編號
                      " AND dzaastus = 'Y' ",           #只取有效的
                   #-----------------------dzfs------------------------------#
                      " AND dzfs001 = dzaa004 ",        #串識別碼版號
                      " AND dzfs002 = dzaa001 ",        #串規格編號
                      " AND dzfs003 <> 's_browse' ",    #不要s_browse的
                      " AND dzfs005 = dzaa006 ",        #使用標示
#                     " AND dzfs011 = dzaa010 ",        #客戶編號
                      " AND dzfs004 = ? ",              #table id
                      " AND dzfsstus = 'Y' ",           #只取有效的
                      " AND dzfs003 NOT IN ( ",
                          " SELECT dzam003 FROM dzaa_t,dzam_t ",
                           " WHERE dzaa001 = '",g_prog_id CLIPPED,"' ",  #規格編號
                             " AND dzaa002 = '",g_sd_ver CLIPPED,"' ",   #規格版次
                             " AND dzaa003 = 'EXCLUDE' ",                #識別代碼
                             " AND dzaa005 = 'a' ",                      #識別碼類型
#                            " AND dzaa008 = '",g_erp_ver CLIPPED,"' ",  #產品版本
                             " AND dzaa009 = '",g_dgenv CLIPPED,"' ",    #客製標示
#                            " AND dzaa010 = '",g_cust CLIPPED,"' ",     #客戶編號
                             " AND dzaastus = 'Y' ",
                            #------------------------------------------------------#
                             " AND dzam001 = dzaa001 ",  #規格編號
                             " AND dzam004 = dzaa004 ",  #識別碼版次
                             " AND dzam005 = dzaa006 ",  #使用標示
#                            " AND dzam006 = dzaa010 ",  #客戶編號
                             " AND dzamstus = 'Y' ) "  #有效碼

         PREPARE sadzp030_prog_detail_pre FROM ls_t

         #開始搜尋主表table id,並且將主表table id放入 dzfs_t 查看看有無單身歸屬表
         #也等同於主表, 若有,則確認為假雙檔加單身的結構, 因為正常雙檔不可能主表又用在單身

         FOREACH sadzp030_prog_type_data_pre INTO lc_dzag004
            DISPLAY "INFO: 搜尋到單頭表:",lc_dzag004 CLIPPED,"進行判斷..."
            EXECUTE sadzp030_prog_detail_pre USING lc_dzag004 INTO li_detailcnt
            IF li_detailcnt > 0 THEN
               DISPLAY "INFO: 發現",lc_dzag004 CLIPPED," 表格被使用在單身,判為假雙檔加單身"
               LET li_rtn = TRUE
               EXIT FOREACH
            ELSE
               DISPLAY "INFO: 發現",lc_dzag004 CLIPPED," 表格未被單身使用,判為雙檔"
            END IF
         END FOREACH

         FREE sadzp030_prog_detail_pre
         FREE sadzp030_prog_type_data_pre
      END IF
   ELSE
      LET li_rtn = TRUE   #找不到上階單頭資料,假雙檔-> i07
   END IF 

   RETURN li_rtn 
END FUNCTION

############################################################
#+ @code
#+ 函式目的 判斷i04主從表
#+ @param   
#+
#+ @return BOOLEAN TRUE/FALSE
############################################################
PRIVATE FUNCTION sadzp030_tab_relation_is_i04()

   DEFINE ps_pid_table  STRING 
   DEFINE ps_id_table   STRING 
   DEFINE ls_sql        STRING 
   DEFINE l_dzff006     LIKE dzff_t.dzff006
   DEFINE l_dzag004     LIKE dzag_t.dzag004
   DEFINE l_dzff005     LIKE dzff_t.dzff005
   DEFINE li_cnt        LIKE type_t.num5 
   DEFINE li_rtn        LIKE type_t.num5 

   LET ls_sql = "SELECT dzff005,dzff006 ",
                 " FROM dzaa_t,dzff_t", 
             #-----------------------------dzaa-------------#
                " WHERE dzaa001 = '",g_prog_id CLIPPED,"'",  #規格編號
                  " AND dzaa002 = '",g_sd_ver CLIPPED,"'",   #規格版次
                  " AND dzaa003 = 's_browse' ",              #設計點/識別碼
                  " AND dzaa005 = '5'",                      #識別碼類型
#                 " AND dzaa008 = '",g_erp_ver CLIPPED,"' ", #產品版本
                  " AND dzaa009 = '",g_dgenv CLIPPED,"' ",   #客製標示
#                 " AND dzaa010 = '",g_cust CLIPPED,"' ",    #客戶編號
                  " AND dzaastus = 'Y'",
             #-----------------------------dzff-------------#
                  " AND dzff001 = dzaa001 ",        #規格編號
                  " AND dzff002 = dzaa004 ",        #識別碼版次
                  " AND dzff003 = dzaa003 ",        #4fd tag name
                  " AND dzff008 = dzaa006 ",        #使用標示
#                 " AND dzff009 = dzaa010 ",        #客戶編號
                  " AND dzffstus = 'Y'"  
   PREPARE sadzp030_i04_pre FROM ls_sql
   DECLARE sadzp030_i04_pre_curs CURSOR FOR sadzp030_i04_pre

   FOREACH sadzp030_i04_pre_curs INTO l_dzff005,l_dzff006
      CASE 
         WHEN l_dzff005 = "id"
            LET ps_id_table = l_dzff006

         WHEN l_dzff005 = "pid" 
            LET ps_pid_table = l_dzff006
      END CASE    
   END FOREACH 

   # id / pid 是否屬於同一個 table 不同table i04/同一個 table i08  
   IF ps_id_table <> ps_pid_table THEN 
      LET li_rtn = TRUE     
   ELSE 
      LET li_rtn = FALSE 
   END IF 

   FREE sadzp030_i04_pre_curs
   RETURN li_rtn
END FUNCTION 
 
############################################################
#+ @code
#+ 函式目的 取得樹狀資料
#+ @param  ps_type STRING i05/i04/i08樣板 
#+
#+ @return 
############################################################
PUBLIC FUNCTION sadzp030_tab_relation_tree_data(ps_type) 
   DEFINE ls_sql          STRING 
   DEFINE ps_type         STRING 
   DEFINE l_dzff005       LIKE dzff_t.dzff005 
   DEFINE l_dzff006       LIKE dzff_t.dzff006
   DEFINE l_dzff007       LIKE dzff_t.dzff007
   DEFINE lr_dzff         RECORD
             id           STRING,
             pid          STRING,
             type         STRING,
             desc         STRING,
             speed        STRING,
             slid         STRING,
             spid         STRING, 
             stype        STRING,
             type2        STRING,
             type3        STRING,
             type4        STRING,
             type5        STRING,
             type6        STRING
                      END RECORD 
   DEFINE lb_stus LIKE type_t.num5
   
   LET ls_sql = "SELECT dzff005,dzff006,dzff007 ",
                 " FROM dzaa_t,dzff_t", 
             #-----------------------------dzaa-------------#
                " WHERE dzaa001 = '",g_prog_id CLIPPED,"'",  #規格編號
                  " AND dzaa002 = '",g_sd_ver CLIPPED,"'",   #規格版次
                  " AND dzaa003 = 's_browse' ",              #設計點/識別碼
                  " AND dzaa005 = '5' ",                     #識別碼類型
#                 " AND dzaa008 = '",g_erp_ver CLIPPED,"' ", #產品版本
                  " AND dzaa009 = '",g_dgenv CLIPPED,"' ",   #客製標示
#                 " AND dzaa010 = '",g_cust CLIPPED,"' ",    #客戶編號
                  " AND dzaastus = 'Y' ",
             #-----------------------------dzff-------------#
                  " AND dzff001 = dzaa001 ",        #規格編號
                  " AND dzff002 = dzaa004 ",        #識別碼版次
                  " AND dzff003 = dzaa003 ",        #4fd tag name
                  " AND dzff008 = dzaa006 ",        #使用標示
#                 " AND dzff009 = dzaa010 ",        #客戶編號
                  " AND dzffstus = 'Y'"
   PREPARE sadzp030_tree_data_pre FROM ls_sql
   DECLARE sadzp030_tree_data_pre_curs CURSOR FOR sadzp030_tree_data_pre
   
   FOREACH sadzp030_tree_data_pre_curs INTO l_dzff005,l_dzff006,l_dzff007
      CASE 
         WHEN l_dzff005 = "type"   
            LET lr_dzff.type = l_dzff007  
            LET lb_stus = TRUE 
         WHEN l_dzff005 = "type2"  LET lr_dzff.type2 = l_dzff007  
         WHEN l_dzff005 = "type3"  LET lr_dzff.type3 = l_dzff007  
         WHEN l_dzff005 = "type4"  LET lr_dzff.type4 = l_dzff007  
         WHEN l_dzff005 = "type5"  LET lr_dzff.type5 = l_dzff007  
         WHEN l_dzff005 = "type6"  LET lr_dzff.type6 = l_dzff007  

         WHEN l_dzff005 = "desc"   
            LET lr_dzff.desc = l_dzff007
            LET lb_stus = TRUE 
         WHEN l_dzff005 = "pid"   
            LET lr_dzff.pid = l_dzff007
            LET lb_stus = TRUE 
         WHEN l_dzff005 = "id"     
            LET lr_dzff.id = l_dzff007  
            LET lb_stus = TRUE
               
         #i05 樣板 table id 及field id 不可以null    
         WHEN l_dzff005 = "speed" 
            LET lr_dzff.speed = l_dzff006
            
            IF ps_type = "i05" THEN
               IF cl_null(lr_dzff.speed) THEN
                  DISPLAY cl_getmsg("adz-01130",g_lang)
                 #DISPLAY "ERROR: i05樣板:提速檔設定不可以為null"
                  LET g_cross_4gl_stus = FALSE     #跨4gl判斷正確與否
                  INITIALIZE lr_dzff TO NULL 
                  LET lb_stus = FALSE 
                  RETURN lr_dzff.*,lb_stus
               ELSE 
                  LET lb_stus = TRUE   
               END IF 
            ELSE 
               LET lb_stus = TRUE  
            END IF 

         WHEN l_dzff005 = "spid"
            LET lr_dzff.spid = l_dzff007
            IF ps_type = "i05" THEN
               IF cl_null(lr_dzff.spid) THEN 
                  DISPLAY cl_getmsg("adz-01131",g_lang)
                 #DISPLAY "ERROR: i05樣板:欄位spid設定不可以為null"
                  LET g_cross_4gl_stus = FALSE     #跨4gl判斷正確與否
                  INITIALIZE lr_dzff TO NULL 
                  LET lb_stus = FALSE 
                  RETURN lr_dzff.*,lb_stus
               ELSE 
                  LET lb_stus = TRUE   
               END IF 
            ELSE 
               LET lb_stus = TRUE  
            END IF  
         WHEN l_dzff005 = "sid" LET 
            lr_dzff.slid = l_dzff007
            IF ps_type = "i05" THEN
               IF cl_null(lr_dzff.slid) THEN 
                  DISPLAY cl_getmsg("adz-01135",g_lang)
                 #DISPLAY "ERROR: i05樣板:欄位slid設定不可以為null"
                  LET g_cross_4gl_stus = FALSE     #跨4gl判斷正確與否
                  INITIALIZE lr_dzff TO  NULL 
                  LET lb_stus = FALSE 
                  RETURN lr_dzff.*,lb_stus
               ELSE 
                  LET lb_stus = TRUE   
               END IF 
            ELSE 
               LET lb_stus = TRUE  
            END IF   
         WHEN l_dzff005 = "stype"
            LET lr_dzff.stype = l_dzff007
            LET lb_stus = TRUE  
       END CASE   
   END FOREACH
   FREE sadzp030_tree_data_pre_curs

   RETURN lr_dzff.*,lb_stus
END FUNCTION 

############################################################
#+ @code
#+ 函式目的 取得pid 及type
#+ @param  ps_table STRING table id
#+ @param  pr_dzff RECORD 樹狀record
#+ @return RECORD
############################################################
PUBLIC FUNCTION sadzp030_tab_relation_i04_body_pid_type(ps_table,pr_dzff)
   DEFINE l_token      base.StringTokenizer
   DEFINE l_token2     base.StringTokenizer
   DEFINE l_dzed004    LIKE dzed_t.dzed004  #PK
   DEFINE l_dzed006    LIKE dzed_t.dzed006  #FK
   DEFINE lst_fk       STRING 
   DEFINE lst_fk2      STRING 
   DEFINE l_sql        STRING
   DEFINE ps_table     STRING 
   DEFINE pr_dzff      RECORD
            id         STRING,
            pid        STRING,
            type       STRING,
            desc       STRING,
            speed      STRING,
            slid       STRING,
            spid       STRING, 
            stype      STRING,       
            type2      STRING,
            type3      STRING,
            type4      STRING,
            type5      STRING,
            type6      STRING
                   END RECORD 
   DEFINE ls_pid       STRING
   DEFINE ls_type      STRING  

   #取得fk
   CALL sadzp030_tab_relation_fk(ps_table,g_table_main,5) RETURNING l_dzed004,l_dzed006

   #分析取得的左右兩側欄位
   LET l_token = base.StringTokenizer.create(l_dzed004 CLIPPED, ",")
   LET l_token2 = base.StringTokenizer.create(l_dzed006 CLIPPED, ",")

   WHILE l_token.hasMoreTokens() AND l_token2.hasMoreTokens()
      LET lst_fk = l_token.nextToken()
      LET lst_fk2 = l_token2.nextToken() 
      IF lst_fk.getIndexOf("ent",1) AND lst_fk2.getIndexOf("ent",1) THEN #排除 ent 
         CONTINUE WHILE   
      END IF
      IF lst_fk2 = pr_dzff.pid THEN 
         LET ls_pid = lst_fk
      END IF 
    
   END WHILE 
   LET pr_dzff.pid = ls_pid
   LET pr_dzff.type = l_token.nextToken()
   RETURN pr_dzff.*
END FUNCTION    


