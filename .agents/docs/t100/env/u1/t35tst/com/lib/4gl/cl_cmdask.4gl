#該程式未解開Section, 採用最新樣板產出!
#該程式為freestyle程式!
{<section id="cl_cmdask.description" type="s" >}
#應用 a00 樣板自動產生(Version:3)
#+ Standard Version.....: SD版次:0005(1900-01-01 00:00:00), PR版次:0005(2026-04-28 14:31:35)
#+ Customerized Version.: SD版次:0000(1900-01-01 00:00:00), PR版次:0000(1900-01-01 00:00:00)
#+ Build......: 000653
#+ Filename...: cl_cmdask
#+ Description: 開窗詢問準備新執行作業
#+ Creator....: 01258(2013-08-16 09:55:12)
#+ Modifier...: 00000 -SD/PR- 20992
 
{</section>}
 
{<section id="cl_cmdask.global" type="s" >}
#應用 p00 樣板自動產生(Version:6)
#add-point:填寫註解說明 name="main.memo"
#190122-00044 #1 2019/04/19 By 10167   調整視窗全螢幕問題
#260129-00012#1  2026/01/29 by 20992   因應AI 串接調整需求，配合調整相關工具。

#end add-point
#add-point:填寫註解說明(客製用) name="main.memo_customerization"

#end add-point
#(ver:6) ---start---
#add-point:填寫註解說明(行業用) name="global.memo_industry"

#end add-point
#(ver:6) --- end ---
 
IMPORT os
#add-point:增加匯入項目 name="main.import"
IMPORT util #260129-00012#1 add 

#end add-point
 
SCHEMA ds
 
GLOBALS "../../cfg/top_global.inc"
#add-point:增加匯入變數檔 name="global.inc"

#end add-point
 
{</section>}
 
{<section id="cl_cmdask.free_style_variable" type="s" >}
#add-point:free_style模組變數(Module Variable) name="free_style.variable"

#end add-point
 
{</section>}
 
{<section id="cl_cmdask.global_variable" type="s" >}
#add-point:自定義模組變數(Module Variable) name="global.variable"

#end add-point
 
{</section>}
 
{<section id="cl_cmdask.other_dialog" type="s" >}

 
{</section>}
 
{<section id="cl_cmdask.other_function" readonly="Y" type="s" >}
############################################################
#+ @code
#+ 函式目的  指定時間延後執行 UNIX 指令
#+ @param    p_code  CHAR(20)  程式編號
#+ @param    p_time  CHAR(8)   指定時間
#+ @param    p_cmd   STRING    UNIX 指令
############################################################
PUBLIC FUNCTION cl_cmdask_at(p_code,p_time,p_cmd)
   DEFINE p_code    LIKE gzza_t.gzza001
   DEFINE p_time    LIKE type_t.chr8
   DEFINE p_cmd     STRING
   DEFINE l_n       LIKE type_t.num5 
   DEFINE l_gzza004 LIKE gzza_t.gzza004
   DEFINE li_flag   LIKE type_t.num10
   DEFINE li_pos    LIKE type_t.num5 
   
 
   WHENEVER ERROR CALL cl_err_msg_log
   
   # Recognize what's the program code
   LET li_flag = TRUE

   SELECT gzza004 INTO l_gzza004 FROM gzza_t WHERE gzza001 = p_code
   IF SQLCA.SQLCODE THEN
      LET li_flag = FALSE
   ELSE
      LET li_pos = p_cmd.getIndexOf(l_gzza004 CLIPPED, 1)
      IF li_pos = 0 THEN
         LET li_flag = FALSE
      ELSE
         LET li_pos = li_pos + LENGTH(l_gzza004 CLIPPED)
      END IF
   END IF

   IF NOT li_flag THEN
      LET li_pos = p_cmd.getIndexOf("'", 1)
   END IF
   LET p_cmd = "p_cron ", p_code CLIPPED, " ", p_cmd.subString(li_pos, p_cmd.getLength())
 
   CALL cl_cmdrun_wait(p_cmd)

END FUNCTION
############################################################
#+ @code
#+ 函式目的  開窗詢問使用者新增執行作業代碼,並執行
############################################################
PUBLIC FUNCTION cl_cmdask()
   DEFINE p_cmd STRING
   #260129-00012#1 add --- start ---
   DEFINE la_param      RECORD
         prog              STRING,
         param             DYNAMIC ARRAY OF STRING
                        END RECORD 
   DEFINE l_str         STRING
   DEFINE l_prog        STRING
   DEFINE l_td          STRING
   DEFINE li_ck         LIKE type_t.num5
   #260129-00012#1 add  --- end ---

   WHENEVER ERROR CALL cl_err_msg_log

   #視窗準備
   OPEN WINDOW w_cl_cmdask WITH FORM cl_ap_formpath("lib","cl_cmdask_s01")
   CALL cl_ui_init()

   #190122-00044 By 10167-----(S)
   &ifndef v1
      # style setting
      CALL ui.Interface.loadStyles(FGL_GETENV("ERP")||os.Path.separator()||"cfg"||os.Path.separator()||
                                   "4st"||os.Path.separator()|| g_lang || os.Path.separator() || "controlg.4st")
   &endif
   #190122-00044 By 10167-----(E)
   
   INPUT p_cmd WITHOUT DEFAULTS FROM cmd
      ATTRIBUTE(UNBUFFERED)

      ON IDLE g_idle_seconds
         CALL cl_on_idle()
         CONTINUE INPUT

      ON ACTION controlp
         INITIALIZE g_qryparam TO NULL
         LET g_qryparam.state = "i"
         LET g_qryparam.reqry = TRUE
         CALL q_gzzz001_1()                          #呼叫開窗
         DISPLAY g_qryparam.return1 TO formonly.cmd  #顯示到畫面上
         LET p_cmd = g_qryparam.return1
         NEXT FIELD cmd
   END INPUT

   CLOSE WINDOW w_cl_cmdask
   CALL cl_load_style_list(NULL)   #190122-00044 By 10167

   #放棄執行
   IF INT_FLAG THEN
      LET INT_FLAG = FALSE
      RETURN
   END IF

   #若輸入的項目為空
   IF cl_null(p_cmd) THEN
      RETURN
   END IF

   #查看是否有調整項目

   #調整後的項目不可執行
   IF p_cmd.getIndexOf("azzi000",1) OR p_cmd.getIndexOf("azzp000",1) THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = "lib-510"
      LET g_errparam.extend = NULL
      LET g_errparam.popup = TRUE
      CALL cl_err()
   ELSE
      #260129-00012#1 add --- start ---
      IF FGL_GETENV("T100NANAWEB") = 'Y' THEN
         CALL cl_cmdask_check(p_cmd) RETURNING li_ck, l_prog, l_td
         IF li_ck > 0 AND NOT cl_null(l_prog) THEN
            IF l_prog = l_td THEN
               LET l_td = ""
            END IF
            LET la_param.prog = "adzp902"
            LET la_param.param[1] = "open_erp_prog"
            LET la_param.param[2] = l_prog
            LET la_param.param[3] = g_site
            IF li_ck = 2 THEN
               LET la_param.param[4] = l_td      
            END IF
            LET l_str = util.JSON.stringify( la_param )   
            CALL cl_cmdrun(l_str)
         END IF
         
      ELSE
      #260129-00012#1 add --- end --- 
        #不提供進入shell的指令
        CALL cl_cmdrun(p_cmd)
      END IF #260129-00012#1 add
   END IF

END FUNCTION

################################################################################
# Descriptions...: 判定傳入的字串是否為程式編號或單號，做初步拆解
# Memo...........:
# Usage..........: CALL cl_cmdask_check (传入参数)
#                  RETURNING 回传参数
# Input parameter: p_cmd   呼叫字串
# Return code....: li_ck   類型判斷用 2 表示回傳單號 + 程式代號
#                : l_prog  程式編號
#                : p_cmd   單號
# Date & Author..: 20260312 by 20992
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_cmdask_check(p_cmd)
   DEFINE p_cmd               STRING
   DEFINE lc_prog             LIKE type_t.chr1000
   DEFINE li_industry_status  LIKE type_t.num5
   DEFINE lc_lower            LIKE gzza_t.gzza001  #gzzz_t.gzzz001  #作業名稱
   DEFINE li_flag             LIKE type_t.num5
   DEFINE li_status           LIKE type_t.num10
   DEFINE ls_msg              LIKE type_t.chr1000
   DEFINE lc_stus             LIKE type_t.chr1
   DEFINE lc_code             LIKE gzza_t.gzza001  #gzza_t.gzza001  #程式名稱
   DEFINE li_ck               LIKE type_t.num5
   DEFINE ls_topind2          STRING 
   DEFINE li_space_index  LIKE type_t.num5
   
   LET ls_topind2 = FGL_GETENV("TOPIND2")
   LET lc_prog = ""
   LET li_ck = 1

   LET p_cmd = p_cmd.trim()
   LET li_space_index = p_cmd.getIndexOf(" ", 1)
   IF li_space_index > 1 THEN
      LET lc_prog = p_cmd.subString(1, li_space_index-1)
      #確認是否是作業
      IF NOT cl_cmdask_chk_progcode(lc_prog) THEN
         LET lc_prog = p_cmd
      END IF
   ELSE
      LET lc_prog = p_cmd
   END IF
   

   #確認行業作業判定
   CALL cl_cmdask_replace_industry(lc_prog,"P", ls_topind2) RETURNING lc_prog, li_industry_status
   LET lc_lower = DOWNSHIFT(lc_prog)
   #memo:先针对大写的程式判断是否有行业版，若没有，则往下继续判断
   SELECT COUNT(1) INTO li_flag FROM gzzz_t
    WHERE gzzz001 = lc_prog               #檢驗作業編號
   IF li_flag = 1 THEN
      SELECT gzzz002,gzzzstus
             #程式編號
        INTO lc_code,lc_stus
        FROM gzzz_t
       WHERE gzzz001 = lc_prog               #檢驗作業編號
   ELSE
      SELECT COUNT(1) INTO li_flag FROM gzzz_t
       WHERE gzzz001 = lc_lower              #檢驗作業編號
      IF li_flag = 1 THEN
         SELECT gzzz002,gzzzstus
                #程式編號
           INTO lc_code,lc_stus
           FROM gzzz_t
          WHERE gzzz001 = lc_lower           #檢驗作業編號
          LET lc_prog = lc_lower
      ELSE
          LET SQLCA.SQLCODE = TRUE
      END IF

   END IF
   INITIALIZE g_errparam TO NULL 
   #無此作業編號
   IF SQLCA.SQLCODE THEN
      #無此作業編號 = -1
      CALL cl_cmdask_choose(lc_prog) RETURNING lc_prog,lc_code,lc_stus,li_status, li_ck

      IF li_status = -1 THEN
         LET ls_msg = cl_getmsg("lib-048",g_lang)
         LET ls_msg = cl_replace_err_msg(ls_msg, lc_prog)
      END IF
   ELSE
      IF NOT cl_null(lc_stus) AND lc_stus = "N" THEN
         LET li_status = -2  #作業編號已失效
         LET ls_msg = cl_getmsg("lib-049",g_lang)
         LET ls_msg = cl_replace_err_msg(ls_msg, lc_prog) 
      END IF
      IF cl_null(lc_code) THEN  #gzzz002程式編號不可為NULL
         LET li_status = -3
         LET ls_msg = cl_getmsg("lib-051",g_lang)
      END IF
   END IF

   #總結
   IF li_status < 0 THEN #<0 表示有錯誤
      MENU "Exclamation"
           ATTRIBUTE (STYLE="dialog", COMMENT=ls_msg CLIPPED, IMAGE="exclamation")
         ON ACTION ok
            EXIT MENU
         ON IDLE g_idle_seconds
            CALL cl_on_idle()
            CONTINUE MENU
      END MENU
      LET lc_prog = ""
      LET li_ck = 0
   END IF

   RETURN li_ck,lc_prog, p_cmd
END FUNCTION

################################################################################
# Descriptions...: 確認是否是作業編號、編碼是否為全英文使用
# Memo...........: cl_cmdask_check 的搭配功能
# Usage..........: cl_cmdask_chk_progcode(lc_prog)
#                  RETURNING 回传参数
# Input parameter: lc_prog   判斷字串
# Return code....: 是/否 
# Date & Author..: 20260312 By 20992
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_cmdask_chk_progcode(lc_prog)
   DEFINE buf base.StringBuffer
   DEFINE lc_prog  VARCHAR(40)
   DEFINE li_cnt   INT

   LET buf = base.StringBuffer.create()
   CALL buf.append(lc_prog)
   IF NOT buf.getCharAt(buf.getLength()) MATCHES "[0-9a-zA-Z]" THEN
      RETURN FALSE
   END IF

   SELECT COUNT(1) INTO li_cnt FROM gzzz_t WHERE gzzz001 = lc_prog
   IF SQLCA.SQLCODE OR li_cnt = 0 THEN
      RETURN FALSE
   END IF
   RETURN TRUE
END FUNCTION

################################################################################
# Descriptions...: 抽換行業別的程式名稱
# Memo...........: cl_cmdask_check 的搭配功能
# Usage..........: CALL cl_cmdask_replace_industry(lc_code,lc_mode,lc_topind2)
#                  RETURNING lc_code, li_industry
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 20260312 By 20992
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_cmdask_replace_industry(lc_code,lc_mode,lc_topind2)
   DEFINE lc_code     LIKE gzza_t.gzza001                              
   DEFINE lc_backup   LIKE gzza_t.gzza001                              
   DEFINE lc_gzza001  LIKE gzza_t.gzza001  #gzza_t.gzza001  #程式名稱   
   DEFINE ls_sql      STRING
   DEFINE lc_gzoi001  LIKE type_t.chr3
   DEFINE li_gzzacnt  LIKE type_t.num5
   DEFINE li_industry LIKE type_t.num5
   DEFINE lc_mode     LIKE type_t.chr1
   DEFINE l_gzou003   LIKE gzou_t.gzou003
   DEFINE lc_topind2         STRING 

   LET lc_topind2 = FGL_GETENV("TOPIND2")
   LET li_industry = 0     #FALSE

   IF LENGTH(lc_code) < 4 THEN
      RETURN lc_code, li_industry
   END IF
   LET lc_backup = lc_code   #備份原始傳入訊息

   LET ls_sql = "SELECT count(1) FROM dzeb_t WHERE dzeb002 = 'ooef027' "
   PREPARE cl_cmdask_cnt_ind_pre FROM ls_sql
   EXECUTE cl_cmdask_cnt_ind_pre INTO li_gzzacnt
   #表示舊客沒有這個欄位必免 runtime 錯誤
   IF li_gzzacnt = 0 THEN
      IF g_t100debug >= 6 THEN
         DISPLAY "INFO:舊客沒有這個欄位 ooef027，請使用r.t更新"
      END IF
      RETURN lc_code,li_industry
   END IF

   SELECT gzou003 INTO l_gzou003 FROM gzou_t
    WHERE gzou001 = g_enterprise AND gzoustus = 'Y'

   #抓取據點對應行業
   IF l_gzou003 IS NOT NULL THEN
      LET ls_sql = "SELECT ooef027 FROM ",l_gzou003 CLIPPED,".ooef_t "
   ELSE
      LET ls_sql = "SELECT ooef027 FROM ooef_t "
   END IF
   LET ls_sql = ls_sql," WHERE ooefstus = 'Y' AND ooef001 = ?",
                         " AND ooefent = ? AND ooef201 = 'Y' "
   IF cl_null(g_site) THEN
      #查 g_lang,g_site
      SELECT gzxa007 INTO g_site FROM gzxa_t
       WHERE gzxa001 = g_account
         AND gzxaent = g_enterprise
   END IF
   IF g_t100debug >= 6 THEN
      DISPLAY "ls_sql:",ls_sql
      DISPLAY "g_site:",g_site
   END IF
   PREPARE cl_cmdask_replace_ind_pre FROM ls_sql
   EXECUTE cl_cmdask_replace_ind_pre USING g_site,g_enterprise INTO lc_gzoi001
   IF g_t100debug >= 6 THEN
       DISPLAY "gzoi001:",lc_gzoi001
   END IF
   IF SQLCA.SQLCODE OR cl_null(lc_gzoi001) OR lc_gzoi001 = "sd" THEN
      IF NOT cl_null(lc_topind2) THEN
         LET lc_topind2 ='sd'
      END IF
      #預設標準程式
      RETURN lc_code,li_industry
   END IF
   LET lc_topind2 = lc_gzoi001 CLIPPED
   LET lc_gzza001 = lc_code CLIPPED,"_",lc_gzoi001

   IF lc_mode = "P" THEN
      SELECT COUNT(1) INTO li_gzzacnt FROM gzzz_t WHERE gzzz001 = lc_gzza001 AND gzzzstus = 'Y'
   ELSE
      SELECT COUNT(1) INTO li_gzzacnt FROM gzza_t WHERE gzza001 = lc_gzza001 AND gzzastus = 'Y'
   END IF

   IF li_gzzacnt > 0 THEN
      LET lc_code = lc_gzza001
      IF lc_mode = "P" THEN
         LET li_industry = 10  #作業
      ELSE
         LET li_industry = 20  #程式
      END IF
   ELSE
      DISPLAY "INFO:確認目標是否存在",li_gzzacnt," 運行原執行作業名稱:",lc_code CLIPPED ," lc_mode:",lc_mode
   END IF

   RETURN lc_code, li_industry
END FUNCTION

################################################################################
# Descriptions...: 判定傳入是作業編號或是單號
# Memo...........: cl_cmdask_check 的搭配功能
# Usage..........: cl_cmdask_choose(lc_prog)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 20260312 By 20992
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_cmdask_choose(lc_prog)
   DEFINE lc_param        LIKE type_t.num5
   DEFINE lc_code         LIKE gzza_t.gzza001  #程式名稱  
   DEFINE lc_prog         LIKE gzza_t.gzza001  #作業名稱   
   DEFINE lc_extra        LIKE type_t.chr500   #額外參數
   DEFINE lc_stus         LIKE type_t.chr1
   DEFINE lc_mod_prog     LIKE type_t.chr4     #作業的模組
   DEFINE li_status       LIKE type_t.num10
   DEFINE li_cnt          LIKE type_t.num10    #筆數
   DEFINE li_cnt1         LIKE type_t.num10    #筆數
   DEFINE li_cnt2         LIKE type_t.num10    #筆數
   DEFINE lc_like         LIKE type_t.chr300   
   DEFINE lc_gzzal003     LIKE type_t.chr500   #gzzal_t.gzzal003 #程式名稱
   DEFINE ls_sql          STRING
   DEFINE li_oobzcnt      LIKE type_t.num5
   DEFINE lc_gzou003      LIKE gzou_t.gzou003  #資料所在db  
   DEFINE ls_tmp          STRING               
   DEFINE lc_gzzc002      LIKE gzzc_t.gzzc002  
   DEFINE tok_param       base.StringTokenizer                                
   DEFINE ls_tok          STRING                                              
   DEFINE lc_param_data   LIKE type_t.chr300   #gzzk_t.gzzk004   #應用參數組資料
   DEFINE ls_str          LIKE gzza_t.gzza001              
   DEFINE ls_str_1        STRING               
   DEFINE lc_oobx001      LIKE oobx_t.oobx001  #單據別                        
   DEFINE lc_prog_tmp     LIKE gzza_t.gzza001                                 
   DEFINE lc_site         LIKE type_t.chr20    #單據運用據點 
   DEFINE lc_gzoi001      LIKE type_t.chr3     
   DEFINE ls_topind2         STRING 
   DEFINE li_ck               LIKE type_t.num5

   LET lc_like = ""
   LET ls_topind2 = FGL_GETENV("TOPIND2")
   LET li_ck = 1
   
   #先找作業名稱 或 助記碼完全相符的
   LET lc_gzzal003 = lc_prog
   SELECT gzzal001 INTO lc_prog FROM gzzal_t,gzza_t,gzzz_t
    WHERE gzzal003 = lc_gzzal003
      AND gzzal002 = g_lang
      AND gzzz001 = gzzal001 AND gzzzstus = 'Y'
      AND gzza001 = gzzz002 AND gzzastus = 'Y'

   IF SQLCA.SQLCODE THEN
      SELECT gzzal001 INTO lc_prog FROM gzzal_t,gzza_t,gzzz_t
       WHERE gzzal004 = lc_gzzal003
         AND gzzal002 = g_lang
         AND gzzz001 = gzzal001 AND gzzzstus = 'Y'
         AND gzza001 = gzzz002 AND gzzastus = 'Y'

      IF SQLCA.SQLCODE THEN
         LET lc_like = "%",lc_prog CLIPPED,"%"
      ELSE
         LET li_status = 1
      END IF
   ELSE
      LET li_status = 1
   END IF

   #查不到的作業id,先加上%查看看有沒有筆數
   IF lc_like IS NOT NULL THEN
      SELECT COUNT(1) INTO li_cnt FROM gzzz_t,gzza_t
       WHERE gzzz001 LIKE lc_like
         AND gzzzstus = "Y"
         AND gzza001 = gzzz002 AND gzzastus = 'Y'
      CASE
         WHEN li_cnt > 1
            LET li_status = 1
            CALL cl_cmdask_q_cmd(lc_like,"gzzz_t", ls_topind2) RETURNING lc_prog

         WHEN li_cnt = 1
            SELECT gzzz001 INTO lc_prog FROM gzzz_t,gzza_t
             WHERE gzzz001 LIKE lc_like
               AND gzzzstus = "Y"
               AND gzza001 = gzzz002 AND gzzastus = 'Y'
            LET li_status = 1

         OTHERWISE
            #再查看看作業名稱
            SELECT COUNT(1) INTO li_cnt1 FROM gzzal_t,gzzz_t
             WHERE gzzal003 LIKE lc_like
               AND gzzal002 = g_lang
               AND gzzal001 = gzzz001
               AND gzzzstus = 'Y'
            #220420-00007#1 -e-
            CASE
               WHEN li_cnt1 > 1
                  LET li_status = 1
                  CALL cl_cmdask_q_cmd(lc_like,"gzzal003", ls_topind2) RETURNING lc_prog

               WHEN li_cnt1 = 1
                  SELECT gzzal001 INTO lc_prog FROM gzzal_t,gzzz_t
                   WHERE gzzal003 LIKE lc_like
                     AND gzzal002 = g_lang
                     AND gzzal001 = gzzz001
                     AND gzzzstus = 'Y'
                  LET li_status = 1

               OTHERWISE
                  #再查看看助記碼
                  SELECT COUNT(1) INTO li_cnt2 FROM gzzal_t,gzzz_t,gzza_t
                   WHERE gzzal004 LIKE lc_like
                     AND gzzal002 = g_lang
                     AND gzzz001 = gzzal001 AND gzzzstus = 'Y'
                     AND gzza001 = gzzz002 AND gzzastus = 'Y'

                  CASE
                     WHEN li_cnt2 > 1
                        LET li_status = 1
                        CALL cl_cmdask_q_cmd(lc_like,"gzzal004", ls_topind2) RETURNING lc_prog
                     WHEN li_cnt2 = 1
                        SELECT gzzal001 INTO lc_prog FROM gzzal_t,gzza_t
                         WHERE gzzal004 LIKE lc_like
                           AND gzzal002 = g_lang
                           AND gzzal001 = gzza001
                           AND gzzastus = 'Y'
                        LET li_status = 1

                     OTHERWISE
                        #前方都不符合,開始查看輸入的是否為單據編號  
                        LET li_ck = 2
                        DISPLAY 'INFO:輸入的資料不是作業代號, 檢測是否屬單號:',lc_prog

                        LET ls_tmp = lc_prog                         #  備份單據編號
                        SELECT gzou003 INTO lc_gzou003 FROM gzou_t WHERE gzou001 = g_enterprise

                        #查oobz_t (開單後就寫一筆進入oobz_t) oobz004:流水號完整單號
                        IF lc_gzou003 IS NOT NULL THEN
                           LET ls_sql = "SELECT COUNT(1) FROM ",lc_gzou003 CLIPPED,".oobz_t "
                        ELSE
                           LET ls_sql = "SELECT COUNT(1) FROM oobz_t "
                        END IF
                        LET ls_sql = ls_sql," WHERE oobzent =? AND oobz004 = ? "
                        PREPARE cl_cmdask_docno_oobz FROM ls_sql
                        EXECUTE cl_cmdask_docno_oobz USING g_enterprise,lc_gzzal003 INTO li_oobzcnt

                        IF li_oobzcnt > 0 THEN
                           #有查到 oobz 對應的作業 oobz005:建議開啟作業 oobz002:單據別
                           IF lc_gzou003 IS NOT NULL THEN
                              LET ls_sql = "SELECT oobz002,oobz005,oobzsite FROM ",lc_gzou003 CLIPPED,".oobz_t "
                           ELSE
                              LET ls_sql = "SELECT oobz002,oobz005,oobzsite FROM oobz_t "
                           END IF
                           LET ls_sql = ls_sql," WHERE oobzent =? AND oobz004 = ? "

                           DECLARE cl_cmdask_docno_oobz005 CURSOR FROM ls_sql
                           FOREACH cl_cmdask_docno_oobz005 USING g_enterprise,lc_gzzal003
                              INTO lc_oobx001,lc_prog,lc_site  
                              IF lc_prog IS NOT NULL THEN  
                                 EXIT FOREACH
                              END IF
                           END FOREACH
                           FREE cl_cmdask_docno_oobz005

                           #P類作業規則調整，假如單據編號對應作業是P類改利用單據別到aooi199用單據別當作where 條件抓取對應作業
                              #查 aooi199 (搜尋沒在清單內的單據)
                              IF lc_gzou003 IS NOT NULL THEN
                                 LET ls_sql = "SELECT oobx004 FROM ",lc_gzou003 CLIPPED,".oobx_t "
                              ELSE
                                 LET ls_sql = "SELECT oobx004 FROM oobx_t "
                              END IF
                              LET ls_sql = ls_sql," WHERE oobxent = ? AND oobx001 = ?"
                              LET lc_gzzal003 = cl_cmdrun_get_oobx001(lc_gzzal003)    #取得單別
                              PREPARE cl_cmdask_docno_cs2 FROM ls_sql
                              EXECUTE cl_cmdask_docno_cs2 USING g_enterprise,lc_oobx001 INTO lc_prog_tmp   
                              IF NOT cl_null(lc_prog_tmp) THEN  
                                 LET lc_prog = lc_prog_tmp      
                              END IF 
                           IF g_site <> lc_site THEN
                              #据点切换后，需要重新获取TOPIND2的值
                              IF lc_gzou003 IS NULL THEN
                                 SELECT gzou003 INTO lc_gzou003 FROM gzou_t
                                  WHERE gzou001 = g_enterprise AND gzoustus = 'Y'
                              END IF
                              LET lc_gzoi001 = NULL              
                              #抓取據點對應行業
                              IF lc_gzou003 IS NOT NULL THEN
                                  LET ls_sql = "SELECT ooef027 FROM ",lc_gzou003 CLIPPED,".ooef_t "
                              ELSE
                                  LET ls_sql = "SELECT ooef027 FROM ooef_t "
                              END IF
                              LET ls_sql = ls_sql," WHERE ooefstus = 'Y' AND ooef001 = ?",
                                                      " AND ooefent = ? AND ooef201 = 'Y' "
                              IF g_t100debug >= 6 THEN
                                  DISPLAY "ls_sql:",ls_sql
                                  DISPLAY "lc_site:",lc_site
                              END IF
                              PREPARE cl_cmdrun_reget_ind_pre FROM ls_sql  
                              EXECUTE cl_cmdrun_reget_ind_pre USING lc_site,g_enterprise INTO lc_gzoi001
                              IF g_t100debug >= 6 THEN
                                  DISPLAY "gzoi001:",lc_gzoi001
                              END IF
                              #如果是标准行业的话，将TOPIND2设定为sd
                              IF SQLCA.SQLCODE OR cl_null(lc_gzoi001) OR lc_gzoi001 = "sd" THEN
                                 LET ls_topind2 = FGL_GETENV("TOPIND2")
                                 IF NOT cl_null(ls_topind2) THEN
                                    LET ls_topind2 = 'sd'
                                 END IF
                              END IF
                              LET ls_topind2 = lc_gzoi001 CLIPPED
                           ELSE
                              DISPLAY 'INFO:單號:',ls_tmp,' 作業:',lc_prog CLIPPED," 單別:",lc_gzzal003 CLIPPED
                           END IF
                        ELSE
                           #再查 aooi199 (搜尋沒在清單內的單據)
                           IF lc_gzou003 IS NOT NULL THEN
                              LET ls_sql = "SELECT oobx004 FROM ",lc_gzou003 CLIPPED,".oobx_t "
                           ELSE
                              LET ls_sql = "SELECT oobx004 FROM oobx_t "
                           END IF
                           LET ls_sql = ls_sql," WHERE oobxent = ? AND oobx001 = ?"
                           LET lc_gzzal003 = cl_cmdrun_get_oobx001(lc_gzzal003)    #取得單別

                           PREPARE cl_cmdask_docno_cs FROM ls_sql
                           EXECUTE cl_cmdask_docno_cs USING g_enterprise,lc_gzzal003 INTO lc_prog

                           DISPLAY 'INFO:作業:',lc_prog CLIPPED," 單別:",lc_gzzal003 CLIPPED
                        END IF

                        IF SQLCA.SQLCODE THEN
                           LET li_status = -1
                        ELSE
                           #假如從aooi199 找到的對應作業為 MULTI，跳其對應的作業列表
                           IF lc_prog = 'MULTI' THEN
                              IF lc_gzou003 IS NOT NULL THEN
                                 LET ls_sql = "SELECT oobl002 FROM ",lc_gzou003 CLIPPED,".oobl_t "
                              ELSE
                                 LET ls_sql = "SELECT oobl002 FROM oobl_t "
                              END IF
                              LET ls_sql = ls_sql," WHERE ooblent = ? AND oobl001 = ?"

                              PREPARE cl_cmdask_multi_pre FROM ls_sql
                              DECLARE cl_cmdask_multi_cur CURSOR FOR cl_cmdask_multi_pre
                              #确认当前的TOPIND2的值，site+ind
                              DISPLAY 'CURRENT ls_sql IS : ',ls_sql
                              LET ls_str_1 = ""
                              FOREACH cl_cmdask_multi_cur USING g_enterprise, lc_gzzal003 INTO ls_str
                                 IF cl_null(ls_str_1) THEN
                                    LET ls_str_1 = "'", ls_str, "'"
                                 ELSE
                                    LET ls_str_1 = ls_str_1, ",'", ls_str, "'"
                                 END IF

                                 IF SQLCA.SQLCODE THEN
                                    LET li_status = -1
                                    RETURN lc_prog,lc_code,lc_stus,li_status,li_ck
                                 END IF
                              END FOREACH
                              FREE cl_cmdask_multi_cur
                              FREE cl_cmdask_multi_pre

                              IF ls_str_1.getIndexOf(",",1) THEN
                                 IF NOT cl_null(ls_topind2) AND ls_topind2<> "sd" THEN 
                                    LET ls_str_1 = cl_cmdask_replace_industry3(ls_str_1)
                                 END IF
                              END IF

                              CALL cl_cmdask_q_cmd(ls_str_1,"oobl_t", ls_topind2) RETURNING lc_prog
                              IF cl_null(lc_prog) THEN
                                 LET li_status = -1
                                 RETURN lc_prog,lc_code,lc_stus,li_status,li_ck
                              ELSE
                                 LET li_status = 1
                              END IF
                           ELSE
                              LET li_status = 1
                           END IF

                           RETURN lc_prog,lc_code,lc_stus,li_status,li_ck
                        END IF
                  END CASE
            END CASE
      END CASE
   END IF


   IF lc_prog IS NULL THEN
      LET li_status = -1
   END IF

   #資料都正確的話再執行
   IF li_status = 1 THEN
      SELECT gzzz002,gzzzstus
             #程式編號, 
        INTO lc_code,lc_stus
        FROM gzzz_t
       WHERE gzzz001 = lc_prog    #檢驗作業編號
      IF SQLCA.SQLCODE THEN
         LET li_status = -1
      END IF
   END IF
   RETURN lc_prog,lc_code,lc_stus,li_status,li_ck
END FUNCTION

################################################################################
# Descriptions...: 行業作業檢查
# Memo...........: cl_cmdask_check 的搭配功能
# Usage..........: CALL cl_cmdask_replace_industry3(ls_str)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 20260312 By 20992
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_cmdask_replace_industry3(ls_str)
   DEFINE ls_str      STRING
   DEFINE ls_tmp      STRING
   DEFINE ls_return   STRING
   DEFINE l_cmd_token base.StringTokenizer
   DEFINE l_tok       DYNAMIC ARRAY OF STRING
   DEFINE li_cnt      LIKE type_t.num5
   DEFINE li_pos      LIKE type_t.num5

   CALL l_tok.clear()
   LET li_cnt = 1
   LET l_cmd_token = base.StringTokenizer.create(ls_str,",")
   WHILE l_cmd_token.hasMoreTokens()
      LET ls_tmp = l_cmd_token.nextToken()
      LET l_tok[li_cnt] = ls_tmp.subString(2,ls_tmp.getLength()-1)
      LET li_cnt = li_cnt + 1
   END WHILE

   #檢查
   FOR li_cnt = l_tok.getLength() TO 2 STEP -1
      #傳入值為行業作業->檢查是否有同名非行業作業存在,有就剔除那個非行業的
      IF l_tok[li_cnt].getIndexOf("_",1) THEN
         IF l_tok[li_cnt].subString(l_tok[li_cnt].getLength()-2,l_tok[li_cnt].getLength()-2) = "_" THEN
            FOR li_pos = 1 TO li_cnt-1
               IF l_tok[li_pos] = l_tok[li_cnt].subString(1,l_tok[li_cnt].getLength()-3) THEN
                  LET l_tok[li_pos] = ""
               END IF
            END FOR
         END IF
      END IF
   END FOR

   LET ls_return = ""
   #重組字串
   FOR li_cnt = 1 TO l_tok.getLength()
      IF l_tok[li_cnt].getLength() > 1 THEN
         LET ls_return = ls_return,",'",l_tok[li_cnt].trim(),"'"
      END IF
   END FOR
   LET ls_return = ls_return.subString(2,ls_return.getLength())
   RETURN ls_return
END FUNCTION

################################################################################
# Descriptions...: 開窗選擇作業編號
# Memo...........: cl_cmdask_check 的搭配功能
# Usage..........: CALL cl_cmdask_q_cmd(lc_like,ls_table,lc_topind2)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 20260312 By 20992
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_cmdask_q_cmd(lc_like,ls_table,lc_topind2)
   DEFINE lc_like     LIKE type_t.chr300       
   DEFINE ls_table    STRING
   DEFINE lc_prog     LIKE gzza_t.gzza001      #作業名稱    
   DEFINE la_array    DYNAMIC ARRAY OF RECORD
             gzzz001  LIKE gzza_t.gzza001,     #作業編號    
             gzzal003 LIKE gzzal_t.gzzal003    #程式名稱    
                  END RECORD
   DEFINE li_cnt      LIKE type_t.num10
   DEFINE li_pos      LIKE type_t.num5
   DEFINE ls_tmp      STRING
   DEFINE ls_sql      STRING
   DEFINE lc_keepwindow VARCHAR(1)  
   DEFINE lc_ch       base.channel
   DEFINE mi_q_cmd_acti   LIKE type_t.chr1
   DEFINE lc_topind2         STRING 
   DEFINE la_param      RECORD
         prog              STRING,
         param             DYNAMIC ARRAY OF STRING
                        END RECORD 
   DEFINE l_str         STRING

   LET mi_q_cmd_acti = "N" #先還原屬性:"開窗選的最大,不換!"
   CASE
      WHEN ls_table = "gzzz_t"
         LET ls_sql = "SELECT t1.gzzz001,'' FROM gzzz_t t1,gzza_t ",
                      " WHERE t1.gzzz001 LIKE '",lc_like CLIPPED,"' ",
                      " AND t1.gzzzstus = 'Y' ",
                      " AND t1.gzzz002=gzza001 AND gzzastus='Y' "
         IF NOT cl_null(lc_topind2) AND lc_topind2<> "sd" THEN
            LET ls_sql = ls_sql,
                         " AND (t1.gzzz010 = '",lc_topind2,"' ",
                         "  OR NOT EXISTS(SELECT 1 FROM gzzz_t t2 WHERE t2.gzzz010 = '",lc_topind2,"' AND t2.gzzz001 LIKE t1.gzzz001||'%'))"
         ELSE
            LET ls_sql = ls_sql,
                         " AND t1.gzzz010 = 'sd' "
         END IF
         LET ls_sql = ls_sql,
                      " ORDER BY t1.gzzz001 "

      WHEN ls_table = "gzzal003"
         LET ls_sql = "SELECT gzzal001,gzzal003 FROM gzzal_t,gzzz_t t1,gzza_t ",
                      " WHERE gzzal003 LIKE '",lc_like CLIPPED,"' ",
                        " AND gzzal002 = '",g_lang CLIPPED,"' ",
                        " AND t1.gzzz001 = gzzal001 AND t1.gzzzstus='Y' ",
                        " AND gzza001 = t1.gzzz002 AND gzzastus='Y' "
         IF NOT cl_null(lc_topind2) AND lc_topind2<> "sd" THEN
            LET ls_sql = ls_sql,
                         " AND (t1.gzzz010 = '",lc_topind2,"' ",
                         "  OR NOT EXISTS(SELECT 1 FROM gzzz_t t2 WHERE t2.gzzz010 = '",lc_topind2,"' AND t2.gzzz001 LIKE t1.gzzz001||'%'))"
         ELSE
            LET ls_sql = ls_sql,
                         " AND t1.gzzz010 = 'sd' "
         END IF
         LET ls_sql = ls_sql,
                      " ORDER BY gzzal001 "

      WHEN ls_table = "gzzal004"
         LET ls_sql = "SELECT gzzal001,gzzal003 FROM gzzal_t,gzzz_t t1,gzza_t ",
                      " WHERE gzzal004 LIKE '",lc_like CLIPPED,"' ",
                        " AND gzzal002 = '",g_lang CLIPPED,"' ",
                        " AND t1.gzzz001 = gzzal001 AND t1.gzzzstus='Y' ",
                        " AND gzza001 = t1.gzzz002 AND gzzastus='Y' "
         IF NOT cl_null(lc_topind2) AND lc_topind2<> "sd" THEN
            LET ls_sql = ls_sql,
                         " AND (t1.gzzz010 = '",lc_topind2,"' ",
                         "  OR NOT EXISTS(SELECT 1 FROM gzzz_t t2 WHERE t2.gzzz010 = '",lc_topind2,"' AND t2.gzzz001 LIKE t1.gzzz001||'%'))"
         ELSE
            LET ls_sql = ls_sql,
                         " AND t1.gzzz010 = 'sd' "
         END IF
         LET ls_sql = ls_sql,
                      " ORDER BY gzzal001 "

      WHEN ls_table = "oobl_t"  
         LET ls_sql = "SELECT gzzz001,gzzal003 FROM gzzz_t t1,gzza_t,gzzal_t ",
                      " WHERE t1.gzzz001 IN (",lc_like CLIPPED,")",
                        " AND t1.gzzzstus = 'Y' ",
                        " AND t1.gzzz002 = gzza001 AND gzzastus='Y' ",
                        " AND gzzal001 = gzzz001",
                        " AND gzzal002 = '",g_lang CLIPPED,"' "
         IF NOT cl_null(lc_topind2) AND lc_topind2<> "sd" THEN
            LET ls_sql = ls_sql,
                         " AND (t1.gzzz010 = '",lc_topind2,"' ",
                         "  OR NOT EXISTS(SELECT 1 FROM gzzz_t t2 WHERE t2.gzzz010 = '",lc_topind2,"' AND t2.gzzz001 LIKE t1.gzzz001||'%'))"
         ELSE
            LET ls_sql = ls_sql,
                         " AND t1.gzzz010 = 'sd' "
         END IF
         LET ls_sql = ls_sql,
                      " ORDER BY t1.gzzz001 "
      OTHERWISE
         DISPLAY "ERROR: 傳入表格",ls_table,"有錯!"
         RETURN ""
   END CASE

   DECLARE cl_cmdask_q_cmd_cs CURSOR FROM ls_sql
   IF g_t100debug >= 9 THEN
      DISPLAY "q_cmd ls_table:",ls_table
      DISPLAY "q_cmd ls_sql:",ls_sql
   END IF

   LET li_cnt = 1
   FOREACH cl_cmdask_q_cmd_cs INTO la_array[li_cnt].*
      IF ls_table = "gzzz_t" THEN
         SELECT gzzal003 INTO la_array[li_cnt].gzzal003
           FROM gzzal_t
          WHERE gzzal001 = la_array[li_cnt].gzzz001
            AND gzzal002 = g_lang
      END IF
      LET li_cnt = li_cnt + 1
   END FOREACH
   FREE cl_cmdask_q_cmd_cs
   CALL la_array.deleteElement(li_cnt)

   IF ls_table.subString(1,2) = "gz" AND la_array.getLength() > 1 THEN
      IF g_t100debug >= 6 THEN
         DISPLAY "获取清单时：TOPIND2=",lc_topind2
      END IF
      IF NOT cl_null(lc_topind2) AND lc_topind2<> "sd" THEN   
         #檢查 (基於上面有排序,行業的一定在標準的後面)
         FOR li_cnt = la_array.getLength() TO 2 STEP -1
            #傳入值為行業作業->檢查是否有同名非行業作業存在,有就剔除那個非行業的
            LET ls_tmp = la_array[li_cnt].gzzz001 CLIPPED
            IF ls_tmp.getIndexOf("_",1) THEN
               IF ls_tmp.subString(ls_tmp.getLength()-2,ls_tmp.getLength()-2) = "_" THEN
                  FOR li_pos = 1 TO li_cnt-1
                     IF la_array[li_pos].gzzz001 = ls_tmp.subString(1,ls_tmp.getLength()-3) THEN
                        LET la_array[li_pos].gzzz001 = ""
                     END IF
                  END FOR
               END IF
            END IF
         END FOR
         FOR li_cnt = la_array.getLength() TO 1 STEP -1
            IF la_array[li_cnt].gzzz001 IS NULL THEN
               CALL la_array.deleteElement(li_cnt)
            END IF
         END FOR
      END IF
   END IF

   #如果只有一筆就不開窗了
   CASE
      WHEN la_array.getLength() = 0
         DISPLAY "ERROR: 無法獲取相關作業條件: 表格",ls_table," 條件:",lc_like
         RETURN ""
      WHEN la_array.getLength() = 1
         LET lc_prog = la_array[1].gzzz001
         RETURN lc_prog
   END CASE

   #開窗選擇
   LET lc_keepwindow = "N"
   OPEN WINDOW w_cl_cmdrun_s01 WITH FORM cl_ap_formpath("lib","cl_cmdrun_s01")
      ATTRIBUTE(STYLE="layoutwin")
   CALL cl_ui_locale("cl_cmdrun_s01")

   DIALOG
   DISPLAY ARRAY la_array TO s_table1.*
      ON ACTION accept
         LET lc_prog = la_array[ARR_CURR()].gzzz001
         LET mi_q_cmd_acti = "Y"            #屬性:"開窗選的最大,不換!"
         EXIT DIALOG
   END DISPLAY
   INPUT lc_keepwindow FROM formonly.chk_keepwindow
   END INPUT
      ON ACTION close
         LET INT_FLAG = TRUE
         EXIT DIALOG
      ON ACTION exit
         LET INT_FLAG = TRUE
         EXIT DIALOG

         #交談指令共用ACTION
         &include "common_action.4gl"
            CONTINUE DIALOG
   END DIALOG

   #放棄執行  #150710-00011#1 -start
   IF INT_FLAG THEN
      LET lc_prog = ""
      LET lc_keepwindow = "N"
      LET INT_FLAG = FALSE
   END IF
   #150710-00011#1 -end

   #若選擇續用時,cl_cmdrun本體已經無法支應保留此視窗,改成呼叫azzp880存續
   IF lc_keepwindow = "Y" THEN
      LET ls_sql = os.Path.join(FGL_GETENV("TEMPDIR"),"qcmd_"||FGL_GETPID() CLIPPED||".txt")
      IF os.Path.exists(ls_sql) THEN
         IF os.Path.delete(ls_sql) THEN
         END IF
      END IF
      LET lc_ch = base.Channel.create()
      CALL lc_ch.setDelimiter("")
      CALL lc_ch.OpenFile(ls_sql,"w")
      LET ls_sql = util.JSON.stringify(la_array)
      CALL lc_ch.write(ls_sql)   #util.JSON.format(ls_sql))
      CALL lc_ch.close()
      LET ls_sql = FGL_GETPID()
      #LET ls_sql = "r.r azzp880 '",ls_sql.trim(),"' "
      
      LET la_param.prog = "adzp902"
      LET la_param.param[1] = "open_erp_prog"
      LET la_param.param[2] = "azzp880"
      LET la_param.param[3] = g_site
      LET la_param.param[4] = ls_sql.trim()      
      LET l_str = util.JSON.stringify( la_param )   
      CALL cl_cmdrun(l_str)
      
      RUN ls_sql WITHOUT WAITING
   END IF

   CLOSE WINDOW w_cl_cmdrun_s01
   IF g_t100debug>=3 THEN
      DISPLAY "INFO: 開窗(_s01)選擇作業為:",lc_prog
   END IF

   RETURN lc_prog
END FUNCTION

 
{</section>}
 
