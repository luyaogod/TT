#+ Version..: T6-ERP-1.00.00 Build-00001
#+
#+ Filename.: sadzp030_tab
#+ Buildtype: p01樣板自動產生
#+ Memo.....: 產生tab用

IMPORT os
SCHEMA ds

GLOBALS "../../cfg/top_global.inc"
GLOBALS "../4gl/adzp150_global.inc"   #180411-00062

CONSTANT li_regen = FALSE  #TRUE:一定要重產/FALSE:不一定重產,看板本設定

#判別樣板在指定程式從上次修改到現在的這個期間中，有沒有被改變過
#有TRUE / 沒有FALSE
PUBLIC FUNCTION sadzp030_tab_file_template_renewed(ls_prog_id)
   DEFINE ls_prog_id  STRING             #規格編號
   DEFINE li_return   LIKE type_t.num5
   DEFINE lc_gzzx001  LIKE gzzx_t.gzzx001
   DEFINE lc_gzzx012  LIKE gzzx_t.gzzx012  #應用樣板版本(離手版本)
   DEFINE lc_dzdz001  LIKE dzdz_t.dzdz001  #應用樣板(離手版本)
   DEFINE lc_dzdz002  LIKE dzdz_t.dzdz002  #版次

   #先看總體設定開關
   IF li_regen THEN
      LET li_return = TRUE
      RETURN li_return
   END IF

   LET lc_gzzx001 = ls_prog_id.trim()
   SELECT gzzx011,gzzx012 INTO lc_dzdz001,lc_gzzx012 FROM gzzx_t
    WHERE gzzx001 = lc_gzzx001

   #查不到的開發項目
   IF SQLCA.SQLCODE THEN
      LET li_return = TRUE
      RETURN li_return
   END IF

   #查不到東西
   IF lc_dzdz001 IS NULL OR lc_dzdz001 = " " THEN
      LET li_return = TRUE
      RETURN li_return
   ELSE
      LET lc_dzdz001 = "code_",lc_dzdz001 CLIPPED
   END IF

   #查版本最大值
   SELECT MAX(dzdz002) INTO lc_dzdz002 FROM dzdz_t 
    WHERE dzdz001 = lc_dzdz001

   #判斷版次
   IF lc_dzdz002 = lc_gzzx012 THEN
      IF g_t100debug >= 3 THEN
         DISPLAY "INFO: 偵測到使用樣板:",lc_dzdz001 CLIPPED," 現存最新版本:",lc_gzzx012," 上次使用版本:",lc_dzdz002
         DISPLAY "      判斷為相同版次,不做重新產出"
      END IF
      LET li_return = FALSE
   ELSE
      IF g_t100debug >= 3 THEN
         DISPLAY "INFO: 偵測到使用樣板:",lc_dzdz001 CLIPPED," 現存最新版本:",lc_gzzx012," 上次使用版本:",lc_dzdz002
         DISPLAY "      判斷為不同版次,此處預計將進行重新產出"
      END IF
      LET li_return = TRUE
   END IF

   RETURN li_return
END FUNCTION


PUBLIC FUNCTION sadzp030_tab_file_module(ls_prog_id,lc_type,lc_dgenv)
   DEFINE ls_prog_id  STRING             #規格編號
   DEFINE lc_dgenv    LIKE type_t.chr1   #DGENV資料
   DEFINE lc_type     LIKE type_t.chr1   #回傳型態 A:給我路徑 B:給我模組編號(大寫)
   DEFINE ls_sys      STRING
   DEFINE ls_module   STRING
   DEFINE ls_path     STRING
   DEFINE lc_module   LIKE type_t.chr4
   DEFINE lc_prog     LIKE gzza_t.gzza001
   DEFINE li_cnt      LIKE type_t.num5

   IF lc_type IS NULL OR lc_type = " " THEN
      LET lc_type = "A"
   END IF

   #ls_sys模組代碼
   LET ls_sys = UPSHIFT(ls_prog_id.subString(1,4))

   #ls_sys模組代碼修正
   CASE
      #A/B/D/E開頭均為模組程式, C開頭需要進一步分析
      WHEN ls_sys.subString(1,1) = "A" 
         IF lc_dgenv = "c" THEN   #sadzp030_tab_file_chk_cust(ls_prog_id) THEN
            LET ls_module = "C",ls_sys.subString(2,3)
         ELSE
            LET ls_module = ls_sys.subString(1,3)
         END IF

      WHEN ls_sys.subString(1,1) = "B" 
         LET lc_module = ls_sys.subString(1,3)
         #檢查若模組存在 gzzo_t ->則為行業專用模組下的程式, 不存在則需更換為 A開頭模組
         SELECT COUNT(1) INTO li_cnt FROM gzzo_t
          WHERE gzzo001 = lc_module
         IF li_cnt > 0 THEN
            IF lc_dgenv = "c" THEN  #sadzp030_tab_file_chk_cust(ls_prog_id) THEN
               LET ls_module = "D",ls_sys.subString(2,3)
            ELSE
               LET ls_module = ls_sys.subString(1,3)
            END IF
         ELSE
            IF lc_dgenv = "c" THEN  #sadzp030_tab_file_chk_cust(ls_prog_id) THEN
               LET ls_module = "C",ls_sys.subString(2,3)
            ELSE
               LET ls_module = "A",ls_sys.subString(2,3)
            END IF
         END IF

      WHEN ls_sys.subString(1,1) = "E" 
         LET ls_module = ls_sys.subString(1,3)

      WHEN ls_sys.subString(1,1) = "W" 
         IF lc_dgenv = "c" THEN
            LET ls_module = "CWSS"
         ELSE
            LET ls_module = "WSS"
         END IF

      WHEN ls_sys.substring(1,3) = "CL_"
         IF lc_dgenv = "c" THEN 
            LET ls_module = "CLIB"
         ELSE
            LET ls_module = "LIB"
         END IF

      WHEN ls_sys.subString(1,2) = "S_"
         IF lc_dgenv = "c" THEN 
            LET ls_module = "CSUB"
         ELSE
            LET ls_module = "SUB"
         END IF

      WHEN ls_sys.subString(1,2) = "Q_"
         IF lc_dgenv = "c" THEN 
            LET ls_module = "CQRY"
         ELSE
            LET ls_module = "QRY"
         END IF

      WHEN ls_sys.subString(1,3) = "CS_" LET ls_module = "CSUB"
      WHEN ls_sys.subString(1,3) = "CQ_" LET ls_module = "CQRY"
      WHEN ls_sys = "CCL_" LET ls_module = "CLIB"
      WHEN ls_sys = "CWSS" LET ls_module = "CWSS"

      #剩下的就是 C模組
      OTHERWISE
         LET ls_module = ls_sys.subString(1,3)
   END CASE

   #傳入B回傳 module id/ A回傳 module path
   IF lc_type = "B" THEN
      LET ls_path = ls_module
   ELSE
      #path路徑
      LET ls_path = FGL_GETENV(ls_module)
      IF g_t100debug >= 6 THEN
         DISPLAY "INFO: 取得的tab產生路徑為:",ls_path," <-不可為空"
      END IF
      IF NOT os.Path.exists(ls_path) THEN
         DISPLAY cl_getmsg_parm("adz-01140",g_lang,ls_path)
        #DISPLAY "ERROR: 依據命名原則取出的歸屬模組(",ls_path,")並不存在,請檢查是否符合各項4gl命名原則!"
      END IF
   END IF

   RETURN ls_path
END FUNCTION


PRIVATE FUNCTION sadzp030_tab_file_chk_cust(ls_prog_id)
   DEFINE lc_prog    LIKE gzza_t.gzza001
   DEFINE li_stus    LIKE type_t.num5
   DEFINE lc_type    LIKE type_t.chr1   #類型
   DEFINE ls_prog_id STRING             #規格名稱
   DEFINE l_token    base.StringTokenizer
   DEFINE ls_extension STRING
   DEFINE lc_gzza011 LIKE gzza_t.gzza011

   LET lc_prog = ls_prog_id

   #若程式編號不含 underline,則一定是主程式編號
   IF NOT ls_prog_id.getIndexOf("_",1) THEN
      LET lc_type = "M"  #主程式
   ELSE
      #取出最後一段 underline 的字
      LET l_token = base.StringTokenizer.create(ls_prog_id, "_")
      WHILE l_token.hasMoreTokens()
         LET ls_extension = l_token.nextToken()
      END WHILE

      CASE
         WHEN ls_extension.getLength() = 3 AND ls_extension.subString(1,1) = "s"
            LET lc_type = "F"  #子畫面
         WHEN ( ls_extension.getLength() = 3 AND ls_extension.subString(1,1) = "x" ) OR
              ( ls_extension.getLength() = 3 AND ls_extension.subString(1,1) = "g" ) 
            LET lc_type = "C"  #子程式
         WHEN ls_extension.subString(1,1) MATCHES '[01-9]' AND
              ls_extension.subString(2,2) MATCHES '[01-9]' 
            LET lc_type = "C"  #子程式
         OTHERWISE
            LET lc_type = "M"  #行業別主程式
      END CASE
   END IF

   #判斷是否為客製
   CASE
      WHEN lc_type = "M"
         SELECT gzza011 INTO lc_gzza011 FROM gzza_t
          WHERE gzza001 = lc_prog
      WHEN lc_type = "C"
         SELECT gzde008 INTO lc_gzza011 FROM gzde_t
          WHERE gzde001 = lc_prog
      WHEN lc_type = "F"
         SELECT gzdf003 INTO lc_gzza011 FROM gzdf_t
          WHERE gzdf001 = lc_prog
   END CASE

   IF NOT SQLCA.SQLCODE AND lc_gzza011 = "Y" THEN
      LET li_stus = TRUE
   ELSE
      LET li_stus = FALSE
   END IF

   RETURN li_stus
END FUNCTION




