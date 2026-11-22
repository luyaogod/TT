#+ Version..: T6-ERP-1.00.00 Build-00001
#+
#+ Filename.: sadzp030_tab_de
#+ Buildtype: p01樣板
#+ Memo.....: 由4fd分析資料搜尋出單身資訊
#+          : 190826-00045 02667  修正ref參照型態
#+          : 191118-00010 02667  修正Q類欄位參考
#+          : 201109-00009 00824  模糊查詢功能
#+          : 221011-00017 00824  將瀏覽頁籤的reference由chr80放寬至chr200
#+          : 240930-00022 00824  因應項標中心設計器優化需求，配合調整相關工具及樣板


SCHEMA ds

GLOBALS "../../cfg/top_global.inc"
GLOBALS "../4gl/adzp150_global.inc"   #180411-00062

GLOBALS
   DEFINE g_prog_id         LIKE dzfi_t.dzfi001
   DEFINE g_sd_ver          LIKE dzaa_t.dzaa002  #規格版次/程式版次
#  DEFINE g_erp_ver         LIKE dzaa_t.dzaa008  #產品版本
   DEFINE g_dgenv           LIKE type_t.chr1     #客製標示
#  DEFINE g_cust            LIKE dzax_t.dzax005  #客戶編號
   DEFINE g_prog_type       STRING
   DEFINE g_cross_4gl_stus  LIKE type_t.num5     #跨4gl判斷正確與否
   DEFINE g_table_main      STRING
   DEFINE g_detail_op       DYNAMIC ARRAY OF RECORD
            record_id       STRING,
            table_id        STRING,
            cols            STRING
            END RECORD
END GLOBALS

############################################################
#+ @code
#+ 函式目的 確認指定的s_detail數量
#+
#+ @return  BOOLEAN    TRUE/FALSE
#+ @return  ARRAY      將抓出的s_detail%組合為STRING dynamic輸出
#+ @return  NUMBER(5)  s_detail個數
############################################################
PUBLIC FUNCTION sadzp030_tab_de_cnt()

   DEFINE li_cnt     LIKE type_t.num10
   DEFINE li_cntchk  LIKE type_t.num10
   DEFINE li_deny    LIKE type_t.num5
   DEFINE l_success  LIKE type_t.num5
   DEFINE l_dzfi     DYNAMIC ARRAY OF RECORD
          detail_sn  LIKE type_t.num5,
          detail_id  LIKE dzfi_t.dzfi004,
          detail_tab LIKE dzfs_t.dzfs004,
          layer3_up  LIKE dzag_t.dzag004,  #上階table編號
          cluster    STRING,
          detail_ins LIKE dzfs_t.dzfs006,
          detail_del LIKE dzfs_t.dzfs007,
          detail_app LIKE dzfs_t.dzfs008,
          connect    LIKE dzfs_t.dzfs009   #是否連動
                     END RECORD
   DEFINE lc_dzfi005 LIKE dzfi_t.dzfi005   #順序
   DEFINE li_chk     LIKE type_t.num10  

   LET li_cnt = 1
   LET l_success = TRUE

   DECLARE dzfi_cs CURSOR FOR
    SELECT DISTINCT dzfi004,dzfi005 FROM dzfi_t
     WHERE dzfi001 = g_prog_id    #畫面代碼
       AND dzfi002 = g_sd_ver     #規格版次
       AND dzfi009 = g_dgenv      #客製標示
       AND dzfi004 LIKE 's_detail%'
#      AND dzfi017 = g_cust       #客戶編號
     ORDER BY dzfi005

   FOREACH dzfi_cs INTO l_dzfi[li_cnt].detail_id,lc_dzfi005

      #因應SCROLLGRID會有出現多筆,需濾除
      IF li_cnt > 1 THEN
#         FOR li_cntchk = 1 TO l_dzfi.getLength() -1
#            IF l_dzfi[li_cnt].detail_id = l_dzfi[li_cntchk].detail_id THEN
#display 'delete:',li_cnt
#               CALL l_dzfi.deleteElement(li_cnt)
#               CONTINUE FOREACH
#            END IF
#         END FOR
            IF l_dzfi[li_cnt].detail_id = l_dzfi[li_cnt-1].detail_id THEN
               CONTINUE FOREACH
            END IF
      END IF

      #查看是否此欄位已經被設置成排除
      SELECT COUNT(1) INTO li_deny FROM dzaa_t,dzam_t
       WHERE dzaa001 = g_prog_id   #規格編號
         AND dzaa002 = g_sd_ver    #規格版次
         AND dzaa003 = "EXCLUDE"   #識別代碼
         AND dzaa005 = "a"         #識別碼類型
#        AND dzaa008 = g_erp_ver   #產品版本
         AND dzaa009 = g_dgenv     #客製標示
#        AND dzaa010 = g_cust      #客戶編號
         AND dzaastus = "Y"        #有效碼
      #------------------------------------------------------#
         AND dzam001 = dzaa001     #規格編號
         AND dzam003 = l_dzfi[li_cnt].detail_id   #排除項目
         AND dzam004 = dzaa004     #識別碼版次
         AND dzam005 = dzaa006     #使用標示
#        AND dzam006 = dzaa010     #客戶編號
         AND dzamstus = "Y"        #有效碼
      IF li_deny > 0 THEN CONTINUE FOREACH END IF

      #填入序號,供給4gl內部使用
      LET l_dzfi[li_cnt].detail_sn = li_cnt

      #若抓不到資料，就回傳失敗
      SELECT dzfs004,dzfs006,dzfs007,dzfs008,dzfs009,dzag004
        INTO l_dzfi[li_cnt].detail_tab,l_dzfi[li_cnt].detail_ins,
             l_dzfi[li_cnt].detail_del,l_dzfi[li_cnt].detail_app,
             l_dzfi[li_cnt].connect,   l_dzfi[li_cnt].layer3_up
        FROM dzaa_t,dzag_t,dzfs_t
       #-----------------------dzaa-只有一筆---------------------#
       WHERE dzaa001 = g_prog_id         #規格編號
         AND dzaa002 = g_sd_ver          #規格版次
         AND dzaa003 = "TABLE"           #串dzag_t 的識別碼
         AND dzaa005 = "4"               #串dzag_t 的識別碼編號
#        AND dzaa008 = g_erp_ver         #產品版本
         AND dzaa009 = g_dgenv           #客製標示
#        AND dzaa010 = g_cust            #客戶編號
         AND dzaastus = "Y"              #只取有效的
       #-----------------------dzag------------------------------#
         AND dzag001 = dzaa001           #串識別碼版號
         AND dzag003 = dzaa004           #串識別碼版次
         AND dzag006 = dzaa006           #使用標示
#        AND dzag011 = dzaa010           #客戶編號
         AND dzagstus = "Y"
       #-----------------------dzfs------------------------------#
         AND dzfs001 = dzag003           #串識別碼版號
         AND dzfs002 = dzag001           #串規格編號
         AND dzfs003 = l_dzfi[li_cnt].detail_id
         AND dzfs004 = dzag002           #串table編號
         AND dzfs005 = dzaa006           #使用標示
#        AND dzfs011 = dzaa010           #客戶編號
         AND dzfsstus = "Y"

      IF SQLCA.sqlcode THEN
         DISPLAY "ERROR: 抓取單身資料(dzfs)錯誤, 技術代碼:",SQLCA.SQLCODE," 問題表格編號:",l_dzfi[li_cnt].detail_id
         DISPLAY "       請檢查相關設計資料是否存在或完成上傳!"
         LET l_success = FALSE
         IF g_t100debug >= 3 THEN
            DISPLAY "注意: SR項目(",l_dzfi[li_cnt].detail_id,")未設table關聯設定"
         END IF
         EXIT FOREACH
      ELSE
         IF g_t100debug >= 6 THEN
            DISPLAY "INFO: 表格",l_dzfi[li_cnt].detail_tab,"的上階是",l_dzfi[li_cnt].layer3_up
         END IF
         #檢查一下connect如果為NULL,自動補為Y
         IF l_dzfi[li_cnt].connect IS NULL THEN 
            LET l_dzfi[li_cnt].connect = "Y"
         END IF

         #檢查一下,如果上階table等於主表,則表示為layer2的,不要留下
         IF l_dzfi[li_cnt].layer3_up IS NOT NULL THEN
            IF l_dzfi[li_cnt].layer3_up CLIPPED = g_table_main.trim() OR
               l_dzfi[li_cnt].layer3_up = l_dzfi[li_cnt].detail_tab THEN
               LET l_dzfi[li_cnt].layer3_up = ""
            END IF
         END IF
      END IF

      LET li_cnt = li_cnt + 1
   END FOREACH
   CALL l_dzfi.deleteElement(li_cnt)
   LET li_cnt = li_cnt - 1
   IF li_cnt < 1 THEN
      IF g_t100debug >= 3 THEN
         DISPLAY "注意: 取出單身欄位筆數為:",li_cnt,"筆 (如果為0筆可能會無法產出dataset資料)"
      END IF
   END IF

   RETURN l_success,li_cnt,l_dzfi
END FUNCTION

############################################################
#+ @code
#+ 函式目的 依照需求回傳欄位名稱組合
#+          回傳型態=0 
#+          回傳型態=1  傳回 DEFINE 的需求   reference用欄位補上 (chr80)
#+          回傳型態=2  傳回 INPUT 的需求    去除非主表的欄位
#+          回傳型態=3  傳回 CONSTRUCT 的需求 去除非主表的欄位
#+          回傳型態=4  傳回 sql用欄位  非主表的全部用 ''代換
#+          回傳型態=5  傳回 widget整串欄位名稱  4fd處理處使用
#+          回傳型態=6  傳回 Q類需要的field_type
#+          回傳型態=7  傳回 DEFINE 的需求   reference用欄位補上,去除PK 
#+          回傳型態=8  傳回 CONSTRUCT 的需求 去除非主表的欄位，只取文字欄位 (模糊查詢功能用)   #201109-00009
#+          回傳型態=9  傳回 INPUT 的需求 目標控件(簡易帶值) 僅保留有設定簡易帶值或公式的欄位 (簡易帶值/公式功能用)   #240930-00022
#+          回傳型態=10 傳回 INPUT 的需求 使用控件(簡易帶值) 僅保留有設定簡易帶值或公式的欄位 (簡易帶值/公式功能用)   #240930-00022
#+          回傳型態=11 傳回 INPUT 的需求 目標控件(公式) 僅保留有設定簡易帶值或公式的欄位 (簡易帶值/公式功能用)   #240930-00022
#+          回傳型態=12 傳回 INPUT 的需求 使用控件(公式) 僅保留有設定簡易帶值或公式的欄位 (簡易帶值/公式功能用)   #240930-00022
#+ @param   pc_detailname  char(20)  容器名稱
#+ @param   pi_type        number(5) 回傳型態
#+ @param   pc_tablename   chr(20)   主表
#+
#+ @return  STRING  s_detail欄位敘述字串
############################################################
PUBLIC FUNCTION sadzp030_tab_de(pc_detailname,pi_type,pc_tablename)

   DEFINE pc_detailname  LIKE dzfi_t.dzfi004
   DEFINE pi_type        LIKE type_t.num5
   DEFINE pc_tablename   LIKE type_t.chr20  
   DEFINE lc_dzfi005     LIKE dzfi_t.dzfi005  #順序
   DEFINE lc_dzfi006     LIKE dzfi_t.dzfi006
   DEFINE lc_dzfj004     LIKE dzfj_t.dzfj004  #順序
   DEFINE lc_dzfj005     LIKE dzfj_t.dzfj005
   DEFINE lc_dzfj006     LIKE dzfj_t.dzfj006  #欄位使用的型態
   DEFINE la_dzfj DYNAMIC ARRAY OF RECORD
             dzfj005     LIKE dzfj_t.dzfj005,
             dzfj006     LIKE dzfj_t.dzfj006
                         END RECORD
   DEFINE ls_dzfj005     STRING
   DEFINE ls_tablename   STRING
   DEFINE ls_fldname     STRING
   DEFINE ls_fldname_bak STRING
   DEFINE ls_str         STRING
   DEFINE lc_temp        LIKE type_t.chr1
   DEFINE li_checked     LIKE type_t.num5
   DEFINE li_cnt         LIKE type_t.num5
   DEFINE li_pos         LIKE type_t.num5
   DEFINE lc_dzac005     LIKE dzac_t.dzac005  #Table編號(給q類欄位用的)
   DEFINE lc_dzac002     LIKE dzac_t.dzac002  #欄位編號
   DEFINE li_dzacsqlstus LIKE type_t.num5
   DEFINE lc_dzeb006     LIKE dzeb_t.dzeb006  #欄位屬性 (給設定型態用的)
   DEFINE lc_dzaa003     LIKE dzaa_t.dzaa003
   DEFINE ls_sql STRING
   DEFINE lc_dzai008     LIKE dzai_t.dzai008
   DEFINE lc_dzai011     LIKE dzai_t.dzai011
   DEFINE li_q_extra     LIKE type_t.num5
   DEFINE ls_cnt         LIKE type_t.num5
   DEFINE ls_field_name  LIKE gztz_t.gztz001
   DEFINE lc_tmp         LIKE type_t.chr50 #191118-00010
   DEFINE ls_tmp         STRING            #191118-00010
   DEFINE li_count       LIKE type_t.num5  #201109-00009
   DEFINE li_dzeb001     LIKE dzeb_t.dzeb001   #201109-00009
   DEFINE li_dzeb002     LIKE dzeb_t.dzeb002   #201109-00009


   LET ls_str = ""

   #若容器非s_detail開頭，則return
   IF pc_detailname[1,8] <> "s_detail" THEN
      RETURN ls_str
   END IF
  
   #以s_detail1 取出 dzfi006
   LET ls_sql = " SELECT dzfi005,dzfi006 FROM dzfi_t ",
                 " WHERE dzfi001 = '",g_prog_id CLIPPED,"' ",    #畫面代碼
                   " AND dzfi002 = '",g_sd_ver CLIPPED,"' ",     #規格版次
                   " AND dzfi009 = '",g_dgenv CLIPPED,"' ",      #客製標示
                   " AND dzfi004 = '",pc_detailname CLIPPED,"' ", 
#                  " AND dzfi017 = '",g_cust CLIPPED,"' ",       #客戶編號
                 " ORDER BY dzfi005 " 
   DECLARE dzfj2_cs CURSOR FROM ls_sql

   #以dzfi006 取出單身欄位  
   LET ls_sql = " SELECT dzfj004,dzfj005,dzfj006 FROM dzfj_t ",
                 " WHERE dzfj001 = '",g_prog_id CLIPPED,"' ",    #畫面代碼
                   " AND dzfj002 = '",g_sd_ver CLIPPED,"' ",     #5/23雄傑確認是ALM版次
                   " AND dzfj003 = ? ",
                   " AND dzfj017 = '",g_dgenv CLIPPED,"' ",      #客製標示
#                  " AND dzfj018 = '",g_cust CLIPPED,"' ",       #客戶編號
                 " ORDER BY dzfj004 " 
   DECLARE dzfj_cs CURSOR FROM ls_sql

   CALL la_dzfj.clear()
   FOREACH dzfj2_cs INTO lc_dzfi005,lc_dzfi006
      FOREACH dzfj_cs USING lc_dzfi006 INTO lc_dzfj004,lc_dzfj005,lc_dzfj006 
         IF lc_dzfj006 = "Label" THEN 
         ELSE
            LET li_cnt = la_dzfj.getLength()+1
            LET la_dzfj[li_cnt].dzfj005 = lc_dzfj005  
            LET la_dzfj[li_cnt].dzfj006 = lc_dzfj006
         END IF
      END FOREACH
   END FOREACH

   CLOSE dzfj2_cs FREE dzfj2_cs
   CLOSE dzfj_cs FREE dzfj_cs

   #特別的pi_type=5處理lc_dzfj005
   IF pi_type = 5 THEN
      FOR li_pos = 1 TO la_dzfj.getLength() 
         IF cl_null(ls_str) THEN 
            LET ls_str = la_dzfj[li_pos].dzfj005 CLIPPED
         ELSE
            LET ls_str = ls_str CLIPPED,",",la_dzfj[li_pos].dzfj005 CLIPPED
         END IF
      END FOR
      RETURN ls_str
   END IF

   #其他一般回傳項目
   FOR li_pos = 1 TO la_dzfj.getLength() 
      LET lc_dzfj005 = la_dzfj[li_pos].dzfj005
      LET lc_dzfj006 = la_dzfj[li_pos].dzfj006

      LET ls_dzfj005 = lc_dzfj005
      LET ls_fldname_bak = ""   #備份清空

      #201109-00009 ---start---
      #模糊查詢須排除非文字欄位 或是 不會加上*的文字欄位
      IF pi_type = 8 THEN
         IF lc_dzfj006 = "DateTimeEdit" OR lc_dzfj006 = "TimeEdit" OR lc_dzfj006 = "DateEdit" OR
            lc_dzfj006 = "CheckBox" OR lc_dzfj006 = "RadioGroup" OR lc_dzfj006 = "ComboBox" OR
            lc_dzfj006 = "FFLabel" OR lc_dzfj006 = "Slider" OR lc_dzfj006 = "SpinEdit" THEN

            CONTINUE FOR
         END IF
      END IF
      #201109-00009 --- end ---

      #抓取該欄位的table name
      LET ls_tablename = ls_dzfj005.subString(1,ls_dzfj005.getIndexOf(".",1)-1)

      #抓取欄位代號
      LET ls_fldname = ls_dzfj005.subString(ls_dzfj005.getIndexOf(".",1)+1,ls_dzfj005.getLength())
      IF cl_null(ls_fldname) THEN
         LET ls_fldname = ls_dzfj005
      END IF

      #如果是q類樣板
      IF g_prog_type.subString(1,1) = "q" THEN

         LET lc_dzaa003 = ls_fldname
         
         #ka0132
         #若為b_xxxxx 要去除b_
         #IF lc_dzaa003[1,2] = "b_" THEN
         #   LET lc_dzaa003 = cl_str_replace(lc_dzaa003,"b_","")
         #   DISPLAY "lc_dzaa003-d1:",lc_dzaa003
         #ELSE 
         #   DISPLAY "lc_dzaa003-d2:",lc_dzaa003
         #END IF
         #ka0132
         
         #因為q類設定s_detail內均以COLUMN_LIKE方式呈現,所以若b_開頭的,
         #要進行ls_tablename/ls_fldname的回歸抓取
         IF ls_fldname.subString(1,2) = "b_" AND ls_tablename IS NULL THEN
            SELECT dzac005 INTO lc_dzac005 FROM dzaa_t,dzac_t
             WHERE dzaa001 = g_prog_id   #程式編號/規格編號
               AND dzaa002 = g_sd_ver    #規格版號
               AND dzaa003 = lc_dzaa003  #識別碼
               AND dzaa005 = '1'         #識別碼類型
#              AND dzaa008 = g_erp_ver   #產品版本
               AND dzaa009 = g_dgenv     #客製標示
#              AND dzaa010 = g_cust      #客戶編號
               AND dzaastus = "Y"        #狀態取 Y 有效
             #-----------------------------------------------#
               AND dzac001 = dzaa001     #規格編號
               AND dzac003 = dzaa003     #識別碼
               AND dzac004 = dzaa004     #規格版次
               AND dzac012 = dzaa006     #使用標示
#              AND dzac013 = dzaa010     #客戶編號
            IF SQLCA.SQLCODE THEN
            ELSE
               LET ls_tablename = lc_dzac005

               #q類樣板若是column like的方式定義 (q類b_開頭的欄位代號是以column like的方式定義)
               #需檢查此欄位的欄位代號，在去掉b_後所得的欄位代號是否存在於資料庫中，
               #以避免程式中define或select失敗
               LET ls_cnt = 0
               LET ls_field_name = ls_fldname.subString(ls_fldname.getIndexOf("b_",1)+2,ls_fldname.getLength())
               SELECT COUNT(1) INTO ls_cnt FROM gztz_t
                WHERE gztz001 = lc_dzac005
                  AND gztz002 = ls_field_name

               IF ls_cnt <= 0 THEN
                  LET ls_tablename = ""
               END IF
            END IF
            IF NOT cl_null(ls_tablename) THEN
               LET ls_fldname = ls_fldname.subString(3,ls_fldname.getLength())
            END IF
         ELSE
            LET ls_tablename = ""
         END IF

        #IF (ls_fldname.subString(1,2) = "b_" OR ls_fldname.getIndexOf("_desc",3) THEN
         # b_開頭都要濾除
         IF ls_fldname.subString(1,2) = "b_" THEN
            LET ls_fldname_bak = ls_fldname   #備份代用
            LET ls_fldname = ls_fldname.subString(3,ls_fldname.getLength())
         END IF
      END IF

      #重複的PK會使用alias id,主要就是原表名後方 +1,2,3,4,...應濾除
      IF pi_type = 1 OR pi_type = 7 THEN
         IF ls_tablename.subString(ls_tablename.getIndexOf("_t",4)+2,ls_tablename.getLength()) MATCHES '[123456789]' THEN
            IF g_t100debug >= 3 THEN
               DISPLAY "注意: 偵測到PK欄位重複:",ls_tablename," 應用於欄位:",ls_fldname
            END IF
            LET ls_tablename = ls_tablename.subString(1,ls_tablename.getIndexOf("_t",4)+1)
         END IF
      END IF

      IF pi_type = 1 AND g_prog_type.subString(1,1) = "q" AND ls_fldname.getIndexOf("_",3) THEN 
         LET li_q_extra = TRUE
      ELSE
         LET li_q_extra = FALSE
      END IF

      #若無table name(reference)或table name不等同主表
      # ==>回傳型態= 1時，欄位後面接 (chr80)
      # ==>回傳型態= 2時，CONTINUE FOR
      # ==>回傳型態= 3時，CONTINUE FOR
      # ==>回傳型態= 4時，欄位改為""
      # ==>回傳型態= 5時，彙整lc_dzfj005回傳 (CASE外處理)
      IF cl_null(ls_tablename) OR ( ls_tablename <> pc_tablename ) OR li_q_extra THEN
         CASE 
            WHEN pi_type = 1 OR pi_type = 7  #做出DEFINE
               #檢查有無在單身出現的 reference
               IF lc_dzfj006 = "FFLabel" THEN
                  CALL sadzp030_tab_gen_reference_array(ls_fldname)
               END IF

               #補充規則:EDIT也會有_desc
               IF lc_dzfj006 = "Edit" THEN
                  IF ls_fldname.subString(ls_fldname.getLength()-4,ls_fldname.getLength()) = "_desc" THEN
                     CALL sadzp030_tab_gen_reference_array(ls_fldname)
                  END IF
               END IF

               CASE
                  #若出現欄位名稱為 abcd001_N 且 N = s_detailN 時,表示為重複出現的PK
                  WHEN sadzp030_tab_de_pk_again(ls_fldname,pc_detailname)
                     LET ls_fldname = ls_fldname.subString(1,ls_fldname.getIndexOf("_",1)-1)
                     IF pi_type = 7 THEN
                        CONTINUE FOR
                     END IF

                  OTHERWISE
                     #原來就不存在table name
                     IF cl_null(ls_tablename) THEN
                        LET lc_dzaa003 = ls_fldname

                        IF pi_type = 1 AND ls_fldname_bak IS NOT NULL THEN 
                           LET lc_dzaa003 = ls_fldname_bak   #備份回用
                        END IF

                        #一般欄位有dzaa+dzac/ 可是reference只有dzaa+dzai, 所以偵測到_desc要加做 dzaa_dzai
                        SELECT dzac002,dzac005 INTO lc_dzac002,lc_dzac005
                          FROM dzaa_t,dzac_t
                         WHERE dzaa001 = g_prog_id   #程式編號/規格編號
                           AND dzaa002 = g_sd_ver    #規格版號
                           AND dzaa003 = lc_dzaa003  #識別碼
                           AND dzaa005 = '1'         #識別碼類型
#                          AND dzaa008 = g_erp_ver   #產品版本
                           AND dzaa009 = g_dgenv     #客製標示
#                          AND dzaa010 = g_cust      #客戶編號
                           AND dzaastus = "Y"        #狀態取 Y 有效
                         #-----------------------------------------------#
                           AND dzac001 = dzaa001     #規格編號
                           AND dzac003 = dzaa003     #識別碼
                           AND dzac004 = dzaa004     #規格版次
                           AND dzac012 = dzaa006     #使用標示
#                          AND dzac013 = dzaa010     #客戶編號
                        IF SQLCA.SQLCODE THEN
                           LET li_dzacsqlstus = FALSE
                        ELSE
                           LET li_dzacsqlstus = TRUE
                        END IF

                        SELECT dzeb006 INTO lc_dzeb006 FROM dzeb_t
                         WHERE dzeb001 = lc_dzac005 
                           AND dzeb002 = lc_dzac002
                        IF SQLCA.SQLCODE OR NOT li_dzacsqlstus THEN
                           LET lc_dzeb006 = NULL
                           #此處排除 sel 的任何設定 (公版需求)
                           IF g_prog_type.subString(1,1) = "q" AND lc_dzaa003 = "sel" THEN
                              LET lc_dzeb006 = "C001"
                           ELSE
                              IF pi_type = 1 THEN
                                 #如果dzaa003最後五碼是 _desc,多跑dzaa+dzai
                                 IF LENGTH(lc_dzaa003) > 5 THEN
                                    IF lc_dzaa003[LENGTH(lc_dzaa003)-4,LENGTH(lc_dzaa003)] = "_desc" THEN
                           
                                       SELECT dzai008,dzai011 INTO lc_dzai008,lc_dzai011 FROM dzaa_t,dzai_t
                                        WHERE dzaa001 = g_prog_id   #程式編號/規格編號
                                          AND dzaa002 = g_sd_ver    #規格版號
                                          AND dzaa003 = lc_dzaa003  #識別碼
                                          AND dzaa005 = "6"         #識別碼類型
#                                         AND dzaa008 = g_erp_ver   #產品版本
                                          AND dzaa009 = g_dgenv     #客製標示
#                                         AND dzaa010 = g_cust      #客戶編號
                                          AND dzaastus = "Y"        #狀態取 Y 有效
                                        #-----------------------------------------------#
                                          AND dzai001 = dzaa001     #規格編號
                                          AND dzai002 = dzaa003     #識別碼
                                          AND dzai003 = dzaa004     #規格版次
                                          AND dzai004 = dzaa006     #使用標示
#                                         AND dzai012 = dzaa010     #客戶編號
                                       IF NOT SQLCA.SQLCODE THEN
                                          SELECT dzeb006 INTO lc_dzeb006 FROM dzeb_t
                                           WHERE dzeb001 = lc_dzai008
                                             AND dzeb002 = lc_dzai011
                                       END IF
                                    END IF
                                 END IF

                                 #191118-00010 -S- 
                                 # Q類額外加工
                                 IF cl_null(lc_dzeb006)              AND  #沒有抓到型態
                                    g_prog_type.subString(1,1) = "q" AND  #Q類
                                    lc_dzaa003[1,2] = "b_"           THEN #欄位有出現b_
                                    #若為b_xxxx -> 把b_截掉重抓型態看看
                                    LET ls_tmp = lc_dzaa003
                                    LET lc_tmp = ls_tmp.subString(3,ls_tmp.getLength())
                                    SELECT dzeb006 INTO lc_dzeb006 FROM dzeb_t
                                     WHERE dzeb002 = lc_tmp
                                 ELSE
                                    #其他的也比看看, 是否有對應的參考型態
                                    SELECT dzeb006 INTO lc_dzeb006 FROM dzeb_t
                                     WHERE dzeb002 = lc_dzaa003
                                 END IF
                                 #191118-00010 -E-
 
                                 IF cl_null(lc_dzeb006) THEN
                                    LET lc_dzeb006 = "C302"
                                    IF g_t100debug >= 3 THEN
                                       DISPLAY "注意: ",lc_dzaa003,"在第",g_sd_ver,"版次的資料不存在,改用chr500(",SQLCA.SQLCODE,")"
                                    END IF
                                 END IF
                              END IF
                           END IF
                        END IF

                        IF g_t100debug >= 6 THEN
                           DISPLAY "INFO: 欄位(",ls_fldname,")使用",lc_dzeb006,"做為資料型態"
                        END IF
                        #190826-00045#1 -S-
                        #如果有找到對應欄位就使用該欄位作為型態參考
                        #LET ls_fldname = ls_fldname CLIPPED,sadzp030_tab_ma_datatype2(lc_dzeb006)
                        IF NOT cl_null(lc_dzai008) AND NOT cl_null(lc_dzai011) THEN
                           LET ls_fldname = ls_fldname CLIPPED, "(", lc_dzai008,".",lc_dzai011,")"
                           LET lc_dzai008 = ""
                           LET lc_dzai011 = ""
                        ELSE 
                           LET ls_fldname = ls_fldname CLIPPED,sadzp030_tab_ma_datatype2(lc_dzeb006)
                        END IF
                        #190826-00045#1 -E-
                     ELSE
                        #存在table name的話
                        IF g_t100debug >= 6 THEN
                           DISPLAY "INFO: 欄位(",ls_fldname,")使用",ls_tablename,"表及",ls_fldname,"做為資料型態"
                        END IF
                        LET ls_fldname = ls_fldname CLIPPED,sadzp030_tab_ma_datatype(ls_tablename,ls_fldname)
                     END IF
               END CASE

            WHEN pi_type = 2  #做出INPUT ARRAY
               IF lc_dzfj006 = "FFLabel" THEN
                  CONTINUE FOR
               END IF

            WHEN pi_type = 3  #做出CONSTRUCT

            WHEN pi_type = 4  #做出SQL
               #主表回傳調整成'' (為了SQL),但資料必須保存在一個全域變數組備用
               IF ls_tablename IS NOT NULL THEN
                  LET li_checked = FALSE
                  IF g_detail_op.getLength() > 0 THEN
                     FOR li_cnt = 1 TO g_detail_op.getLength()
                        IF g_detail_op[li_cnt].record_id = pc_detailname AND
                           g_detail_op[li_cnt].table_id = ls_tablename THEN
                           LET li_checked = TRUE
                           EXIT FOR
                        END IF
                     END FOR
                     IF NOT li_checked THEN
                        LET li_cnt = g_detail_op.getLength() + 1
                     END IF
                  ELSE
                     LET li_cnt = 1
                  END IF
                  LET g_detail_op[li_cnt].record_id = pc_detailname CLIPPED
                  LET g_detail_op[li_cnt].table_id = ls_tablename CLIPPED
                  LET g_detail_op[li_cnt].cols = g_detail_op[li_cnt].cols,ls_fldname,","
                  #最後回傳前調整
               END IF
               LET ls_fldname = "''"

            WHEN pi_type = 6  #做出DEFINE
               #檢查有無在單身出現的 reference
               IF lc_dzfj006 = "FFLabel" THEN
                  CALL sadzp030_tab_gen_reference_array(ls_fldname)
               END IF
         END CASE
      END IF

      #特別的欄位處理
      CASE 
         WHEN pi_type = 2  #做出INPUT ARRAY
            LET lc_temp = sadzp030_tab_ma_2(ls_fldname,"dzac017")
            IF lc_temp IS NOT NULL THEN
               IF lc_temp = "Y" OR lc_temp = "y" THEN
               ELSE
                  #DISPLAY "注意: 欄位(",ls_fldname,")未設定為可編輯"
                  CONTINUE FOR
               END IF
            ELSE
               CONTINUE FOR
            END IF
            #table由於INPUT ARRAY是整組的概念,因此dzac017是否應該關閉
            #調整到 sadzp030_tab_check_comm_for_deny() 整批判斷

         WHEN pi_type = 3  #做出CONSTRUCT
            LET lc_temp = sadzp030_tab_ma_2(ls_fldname,"dzac018")
          # IF g_prog_type.subString(1,1) = "q" THEN
          #    LET lc_temp = "Y"
          # END IF
            IF lc_temp IS NOT NULL THEN
               IF lc_temp = "Y" OR lc_temp = "y" THEN
               ELSE
                  CONTINUE FOR
               END IF
            ELSE    
               CONTINUE FOR
            END IF

         #201109-00009 ---start---
         WHEN pi_type = 8   #模糊查詢
            IF cl_null(ls_tablename) OR ( ls_tablename <> pc_tablename ) OR li_q_extra THEN
               #在設計器上沒勾選,就不做後面的事
               LET lc_temp = sadzp030_tab_ma_2(ls_fldname,"dzac018")
               IF lc_temp IS NOT NULL THEN
                  IF lc_temp = "Y" OR lc_temp = "y" THEN
                     #確認是否為文字型態，若是才組入
                     #原來就不存在table name
                     IF cl_null(ls_tablename) THEN

                        LET lc_dzaa003 = ls_fldname

                        SELECT dzac002,dzac005 INTO lc_dzac002,lc_dzac005
                          FROM dzaa_t,dzac_t
                         WHERE dzaa001 = g_prog_id   #程式編號/規格編號
                           AND dzaa002 = g_sd_ver    #規格版號
                           AND dzaa003 = lc_dzaa003  #識別碼
                           AND dzaa005 = "1"         #識別碼類型
##                         AND dzaa008 = g_erp_ver   #產品版本
                           AND dzaa009 = g_dgenv     #客製標示
##                         AND dzaa010 = g_cust      #客戶編號
                           AND dzaastus = "Y"        #狀態取 Y 有效
                         #-----------------------------------------------#
                           AND dzac001 = dzaa001     #規格編號
                           AND dzac003 = dzaa003     #識別碼
                           AND dzac004 = dzaa004     #規格版次
                           AND dzac012 = dzaa006     #使用標示
##                         AND dzac013 = dzaa010     #客戶編號
                        IF SQLCA.SQLCODE THEN
                           LET lc_dzeb006 = "C302"
                           IF g_t100debug >= 3 THEN
                              DISPLAY "注意: ",lc_dzaa003,"在第",g_sd_ver,"版次的資料不存在,改用chr80"
                           END IF
                        END IF

                        #確認是否為文字型態欄位
                        IF cl_null(lc_dzac002) OR cl_null(lc_dzac005) THEN
                           SELECT COUNT(1) INTO li_count
                             FROM dzeb_t
                            WHERE dzeb002 = lc_dzaa003
                              AND dzeb007 IN ('varchar','varchar2','clob','blob')   #模糊查詢只針對文字型態欄位
                           IF li_count = 0 THEN
                              CONTINUE FOR
                           END IF
                        ELSE
                           SELECT COUNT(1) INTO li_count
                             FROM dzeb_t
                            WHERE dzeb001 = lc_dzac005
                              AND dzeb002 = lc_dzac002
                              AND dzeb007 IN ('varchar','varchar2','clob','blob')   #模糊查詢只針對文字型態欄位
                           IF li_count = 0 THEN
                              CONTINUE FOR
                           END IF
                        END IF
                     ELSE
                        #存在table name的話
                        LET li_dzeb001 = ls_tablename
                        LET li_dzeb002 = ls_fldname
                        LET li_count = 0
                        SELECT COUNT(1) INTO li_count FROM dzeb_t
                         WHERE dzeb001 = li_dzeb001
                           AND dzeb002 = li_dzeb002
                           AND dzeb007 IN ('varchar','varchar2','clob','blob')   #模糊查詢只針對文字型態欄位
                        IF li_count = 0 THEN
                           CONTINUE FOR
                        END IF
                     END IF
                  ELSE
                     CONTINUE FOR
                  END IF
               ELSE
                  CONTINUE FOR
               END IF
            ELSE
               #在設計器上沒勾選,就不做後面的事
               LET lc_temp = sadzp030_tab_ma_2(ls_fldname,"dzac018")
               IF lc_temp IS NOT NULL THEN
                  IF lc_temp = "Y" OR lc_temp = "y" THEN
                     LET li_dzeb001 = ls_tablename
                     LET li_dzeb002 = ls_fldname
                     LET li_count = 0
                     SELECT COUNT(1) INTO li_count FROM dzeb_t
                      WHERE dzeb001 = li_dzeb001
                        AND dzeb002 = li_dzeb002
                        AND dzeb007 IN ('varchar','varchar2','clob','blob')   #模糊查詢只針對文字型態欄位
                     IF li_count = 0 THEN
                        CONTINUE FOR
                     END IF
                  ELSE
                     CONTINUE FOR
                  END IF
               ELSE
                  CONTINUE FOR
               END IF
            END IF
         #201109-00009 --- end ---

         #240930-00022 ---start---
         WHEN pi_type = 9   #簡易帶值   #目標控件
            #怕舊客環境無此表格，先檢查表格是否存在
            LET li_count = 0
            SELECT COUNT(1) INTO li_count FROM gztz_t
             WHERE gztz001 = 'dzcp_t'
            IF li_count > 0 THEN
               LET li_count = 0
               SELECT COUNT(1) INTO li_count
                 FROM dzcp_t
                WHERE dzcp001 = g_prog_id
                  AND dzcp002 = lc_dzfj005
                  AND dzcpstus = 'Y'
             
               IF li_count <= 0 THEN
                  CONTINUE FOR
               END IF
            ELSE
               CONTINUE FOR
            END IF

         WHEN pi_type = 10   #簡易帶值   #使用控件
            #怕舊客環境無此表格，先檢查表格是否存在
            LET li_count = 0
            SELECT COUNT(1) INTO li_count FROM gztz_t
             WHERE gztz001 = 'dzcp_t'
            IF li_count > 0 THEN
               LET li_count = 0
               SELECT COUNT(1) INTO li_count
                 FROM dzcp_t,dzcq_t
                WHERE dzcp001 = dzcq001
                  AND dzcp002 = dzcq002
                  AND dzcpstus = 'Y'
                  AND dzcp001 = g_prog_id
                  AND dzcq003 = lc_dzfj005
             
               IF li_count <= 0 THEN
                  CONTINUE FOR
               END IF
            ELSE
               CONTINUE FOR
            END IF

         WHEN pi_type = 11   #公式   #目標控件
            #怕舊客環境無此表格，先檢查表格是否存在
            LET li_count = 0
            SELECT COUNT(1) INTO li_count FROM gztz_t
             WHERE gztz001 = 'dzct_t'
            IF li_count > 0 THEN
               LET li_count = 0
               SELECT COUNT(1) INTO li_count
                 FROM dzct_t
                WHERE dzct001 = g_prog_id
                  AND dzct002 = lc_dzfj005
                  AND dzctstus = 'Y'
             
               IF li_count <= 0 THEN
                  CONTINUE FOR
               END IF
            ELSE
               CONTINUE FOR
            END IF

         WHEN pi_type = 12   #公式   #使用控件
            #怕舊客環境無此表格，先檢查表格是否存在
            LET li_count = 0
            SELECT COUNT(1) INTO li_count FROM gztz_t
             WHERE gztz001 = 'dzct_t'
            IF li_count > 0 THEN
               LET li_count = 0
               SELECT COUNT(1) INTO li_count
                 FROM dzct_t,dzcw_t
                WHERE dzct001 = dzcw001
                  AND dzct002 = dzcw002
                  AND dzctstus = 'Y'
                  AND dzct001 = g_prog_id
                  AND dzcw003 = lc_dzfj005
             
               IF li_count <= 0 THEN
                  CONTINUE FOR
               END IF
            ELSE
               CONTINUE FOR
            END IF
         #240930-00022 --- end ---
      END CASE

      #Q的CONSTRUCT
      IF (g_prog_type.subString(1,1) = "q" AND pi_type = 3) OR
         (g_prog_type.subString(1,1) = "q" AND pi_type = 6) THEN

         #如果遇到控件與參考不同的情況時
         IF lc_dzaa003 <> ls_fldname THEN
            #先處理串查欄位 (串查欄位回傳TRUE)
            IF sadzp030_tab_de_is_dzal(ls_fldname) THEN
               #pi_type = 3 (construct的時候不要放串查)
               IF pi_type = 6 THEN
                  LET ls_str = ls_str CLIPPED,",",ls_fldname CLIPPED,"(",ls_fldname CLIPPED,")"
               END IF
            #在處理非串查欄位
            ELSE
               IF pi_type = 6 THEN
                  LET ls_str = ls_str CLIPPED,",",ls_fldname CLIPPED,"(",lc_dzaa003 CLIPPED,")"
               ELSE  #pi_type = 3 (construct的時候不要放reference)
                  IF ls_fldname.getIndexOf("_desc",1) THEN
                  ELSE
                     #沒有設定欄位歸屬table的就不放到CONSTRUCT
                     IF ls_tablename IS NULL THEN
                     ELSE
                        LET ls_str = ls_str CLIPPED,",",ls_fldname CLIPPED,"(",lc_dzaa003 CLIPPED,")"
                     END IF
                  END IF
               END IF
            END IF

         #控件等於參考欄位,就是一般欄位==>簡單處理
         ELSE
            IF ls_tablename IS NOT NULL THEN
               LET ls_str = ls_str CLIPPED,",",ls_fldname CLIPPED
            ELSE
               IF pi_type = 6 THEN
                  LET ls_str = ls_str CLIPPED,",",ls_fldname CLIPPED,"(",ls_fldname.trim(),")"
               END IF 
            END IF
         END IF

         IF ls_str.subString(1,1) = "," THEN
            LET ls_str = ls_str.subString(2,ls_str.getLength())
         END IF
      ELSE
         #240930-00022 ---modify start---
      #  #將欄位ls_fldname組進字串ls_str內
      #  IF cl_null(ls_str) THEN 
      #     LET ls_str = ls_fldname CLIPPED
      #  ELSE
      #     LET ls_str = ls_str CLIPPED,",",ls_fldname CLIPPED
      #  END IF

         #將欄位ls_fldname組進字串ls_str內
         IF pi_type = 9 OR pi_type = 10 OR pi_type = 11 OR pi_type = 12 THEN
            #把控件編號跟欄位編號都組進字串中
            IF cl_null(ls_str) THEN
               LET ls_str = lc_dzfj005 CLIPPED,"(",ls_fldname CLIPPED,")"
            ELSE
               LET ls_str = ls_str CLIPPED,",",lc_dzfj005 CLIPPED,"(",ls_fldname CLIPPED,")"
            END IF
         ELSE
            IF cl_null(ls_str) THEN
               LET ls_str = ls_fldname CLIPPED
            ELSE
               LET ls_str = ls_str CLIPPED,",",ls_fldname CLIPPED
            END IF
         END IF
         #240930-00022 --- modify end ---
      END IF

   END FOR

   RETURN ls_str

END FUNCTION

PRIVATE FUNCTION sadzp030_tab_de_is_dzal(ls_fldname) 
   DEFINE ls_fldname  STRING
   DEFINE lc_dzal002  LIKE dzal_t.dzal002
   DEFINE li_return   LIKE type_t.num5
   DEFINE li_cnt      LIKE type_t.num5

   LET lc_dzal002 = ls_fldname.trim()

   SELECT COUNT(1) INTO li_cnt FROM dzal_t
    WHERE dzal001 = g_prog_id
      AND dzal002 = lc_dzal002
#     AND dzal009 = g_cust     #客戶編號
      AND dzalstus = "Y"

   IF li_cnt = 1 THEN
      LET li_return = TRUE
   ELSE
      LET li_return = FALSE
   END IF
   RETURN li_return
END FUNCTION

############################################################
#+ @code
#+ 函式目的 檢查PK是否重複出現
#+
#+ @return  BOOLEN  TRUE/FALSE
############################################################
PRIVATE FUNCTION sadzp030_tab_de_pk_again(ls_fldname,ls_detailname)

   DEFINE ls_fldname    STRING  #欄位代碼
   DEFINE ls_detailname STRING
   DEFINE li_succ       LIKE type_t.num5

   #出現欄位名稱為 abcd001_N 且 N = s_detailN 時,表示為重複出現的PK
   IF ls_fldname.subString(ls_fldname.getIndexOf("_",1)+1,ls_fldname.getLength()) =
      ls_detailname.subString(ls_detailname.getIndexOf("s_detail",1)+8,ls_detailname.getLength()) THEN
      LET li_succ = TRUE
   ELSE
      LET li_succ = FALSE
   END IF

   RETURN li_succ
END FUNCTION

PUBLIC FUNCTION sadzp030_tab_brecover(ls_cols)

   DEFINE ls_cols   STRING
   DEFINE ls_return STRING
   DEFINE l_token  base.StringTokenizer
   DEFINE ls_next  STRING

   #若為空值則退出
   LET ls_cols = ls_cols.trim()
   LET ls_return = ""
   IF ls_cols.getLength() = 0 THEN
      RETURN ls_return
   END IF

   LET l_token = base.StringTokenizer.create(ls_cols, ",")

   WHILE l_token.hasMoreTokens()
      LET ls_next = l_token.nextToken()

      #欄位分拆後,各別都補回 b_
      IF ls_next = "b_statepic" OR ls_next = "name"    OR ls_next = "pid"    OR
         ls_next = "id"         OR ls_next = "exp"     OR ls_next = "isnode" OR
         ls_next = "isExp"      OR ls_next = "expcode" OR ls_next = "rank"   THEN
      ELSE
         LET ls_next = "b_",ls_next
      END IF

      LET ls_return = ls_return,",",ls_next
   END WHILE

   RETURN ls_return.subString(2,ls_return.getLength())

END FUNCTION

############################################################
#+ @code
#+ 函式目的 依照需求回傳欄位名稱組合
#+ @param   pc_browsename  char(20) 容器名稱
#+
#+ @return  STRING  s_browse欄位敘述字串
############################################################
PUBLIC FUNCTION sadzp030_tab_de_browse(pc_browsename)

   DEFINE pc_browsename  LIKE dzfi_t.dzfi004
   DEFINE lc_dzfi005     LIKE dzfi_t.dzfi005  #順序
   DEFINE lc_dzfi006     LIKE dzfi_t.dzfi006
   DEFINE lc_dzfj004     LIKE dzfj_t.dzfj004  #順序
   DEFINE lc_dzfj005     LIKE dzfj_t.dzfj005
   DEFINE lc_dzfj006     LIKE dzfj_t.dzfj006
   DEFINE li_len         LIKE type_t.num5
   DEFINE li_cnt         LIKE type_t.num5
   DEFINE ls_fldname     STRING
   DEFINE ls_str         STRING
   DEFINE lc_gztz001     LIKE gztz_t.gztz001  #表名稱
   DEFINE lc_gztz002     LIKE gztz_t.gztz002  #欄名稱
   DEFINE lc_dzac002     LIKE dzac_t.dzac002  #參照欄位
   DEFINE lc_dzac005     LIKE dzac_t.dzac005  #參照表格

   LET ls_str = ""
  
   #若容器非s_browse開頭，則return
   IF pc_browsename[1,8] <> "s_browse" THEN
      RETURN ls_str
   END IF

   SELECT dzfi005,dzfi006 INTO lc_dzfi005,lc_dzfi006 FROM dzfi_t
    WHERE dzfi001 = g_prog_id    #畫面代碼
      AND dzfi002 = g_sd_ver     #規格版次
      AND dzfi009 = g_dgenv      #客製標示
      AND dzfi004 = pc_browsename
#     AND dzfi017 = g_cust       #客戶編號
    ORDER BY dzfi005
  
   DECLARE dzfj_cs2 CURSOR FOR
    SELECT dzfj004,dzfj005,dzfj006 FROM dzfj_t
     WHERE dzfj001 = g_prog_id    #畫面代碼
       AND dzfj002 = g_sd_ver     #5/23雄傑確認是ALM版次
       AND dzfj003 = lc_dzfi006
       AND dzfj017 = g_dgenv      #客製標示
#      AND dzfj018 = g_cust       #客戶編號
     ORDER BY dzfj004

   FOREACH dzfj_cs2 INTO lc_dzfj004,lc_dzfj005,lc_dzfj006
      LET ls_fldname = ""
     
      #b_statepic/name/pid/id/exp/isnode/isExp/expcode/rank，不組進字串內
      IF lc_dzfj005 = "b_statepic" OR lc_dzfj005 = "name"    OR lc_dzfj005 = "pid"    OR
         lc_dzfj005 = "id"         OR lc_dzfj005 = "exp"     OR lc_dzfj005 = "isnode" OR
         lc_dzfj005 = "isExp"      OR lc_dzfj005 = "expcode" OR lc_dzfj005 = "rank"   THEN
         CONTINUE FOREACH
      END IF

      #確認長度
      LET li_len = LENGTH(lc_dzfj005)

      #注意順序性 有衝突的要判斷先後
      CASE
         #型態為FFImage                  ==> 欄位後面接 (chr50)
         WHEN lc_dzfj006 = "FFImage" 
            LET ls_fldname = lc_dzfj005 CLIPPED,"(chr50)"

         #樣板為i05且f_開頭的欄位        ==> 欄位後面接 (chr80)
         WHEN (g_prog_type = "i05" AND lc_dzfj005[1,2] = "f_")
            LET ls_fldname = lc_dzfj005 CLIPPED,"(chr80)"

         #型態為FFLabel且欄位結尾是_desc ==> 欄位後面接 (chr80)
         WHEN (lc_dzfj006 = "FFLabel" AND lc_dzfj005[li_len-4,li_len] = "_desc")
            #browser上的被參考欄位,避免再多設一次,直接引用主畫面欄位資料 (去b_)
            IF lc_dzfj005[1,2] = "b_" THEN
            #  LET ls_fldname = lc_dzfj005[3,li_len],"(chr80)"   #221011-00017 mark
               LET ls_fldname = lc_dzfj005[3,li_len],"(chr200)"  #221011-00017 modify
            ELSE
            #  LET ls_fldname = lc_dzfj005 CLIPPED,"(chr80)"     #221011-00017 mark
               LET ls_fldname = lc_dzfj005 CLIPPED,"(chr200)"    #221011-00017 modify
            END IF

         #型態為FFLabel且欄位結尾非_desc ==> 欄位後面接 (STRING)
         WHEN (lc_dzfj006 = "FFLabel" AND lc_dzfj005[li_len-4,li_len] <> "_desc")
            LET ls_fldname = lc_dzfj005 CLIPPED,"(STRING)"

         #未使用FFLabel但結尾是 _desc 的 ==> 比照
         WHEN lc_dzfj005[li_len-4,li_len] = "_desc"
            #browser上的被參考欄位,避免再多設一次,直接引用主畫面欄位資料 (去b_)
            IF lc_dzfj005[1,2] = "b_" THEN
            #  LET ls_fldname = lc_dzfj005[3,li_len],"(chr80)"   #221011-00017 mark
               LET ls_fldname = lc_dzfj005[3,li_len],"(chr200)"  #221011-00017 modify
            ELSE
            #  LET ls_fldname = lc_dzfj005 CLIPPED,"(chr80)"     #221011-00017 mark
               LET ls_fldname = lc_dzfj005 CLIPPED,"(chr200)"    #221011-00017 modify
            END IF

         #b_開頭的欄位 ==> 將b_去除
         WHEN lc_dzfj005[1,2] = "b_"
            LET lc_gztz001 = g_table_main.trim()
            LET lc_gztz002 = lc_dzfj005[3,LENGTH(lc_dzfj005)]

            #和主表屬於同一個 table 的話
            SELECT COUNT(1) INTO li_cnt FROM gztz_t
             WHERE gztz001 = lc_gztz001
               AND gztz002 = lc_gztz002
            IF li_cnt > 0 THEN
               LET ls_fldname = lc_dzfj005[3,li_len]
            ELSE
               #2016/07/21 ka0132
               SELECT dzac002,dzac005 
                 INTO lc_dzac002,lc_dzac005 
                 FROM dzaa_t,dzac_t
                WHERE dzaa001 = g_prog_id   #程式編號/規格編號
                  AND dzaa002 = g_sd_ver    #規格版號
                  AND dzaa003 = lc_dzfj005  #識別碼
                  AND dzaa005 = '1'         #識別碼類型
                  AND dzaa009 = g_dgenv     #客製標示
                  AND dzaastus = "Y"        #狀態取 Y 有效
                #-----------------------------------------------#
                  AND dzac001 = dzaa001     #規格編號
                  AND dzac003 = dzaa003     #識別碼
                  AND dzac004 = dzaa004     #規格版次
                  AND dzac012 = dzaa006     #使用標示
               IF cl_null(lc_dzac002) OR cl_null(lc_dzac005) THEN
                  LET ls_fldname = lc_dzfj005[3,li_len],"(chr80)"
               ELSE
                  LET ls_fldname = lc_dzfj005[3,li_len],"(",lc_dzac005,".",lc_dzac002,")"
               END IF
               #2016/07/21 ka0132            
            END IF

         OTHERWISE
            LET ls_fldname = lc_dzfj005 CLIPPED
      END CASE

      ##將欄位ls_fldname組進字串ls_str內
      IF cl_null(ls_str) THEN
         LET ls_str = ls_fldname CLIPPED
      ELSE
         LET ls_str = ls_str CLIPPED,",",ls_fldname CLIPPED
      END IF
   END FOREACH

   RETURN ls_str
END FUNCTION




