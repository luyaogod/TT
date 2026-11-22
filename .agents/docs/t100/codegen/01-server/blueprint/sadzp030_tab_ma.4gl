#+ Version..: T6-ERP-1.00.00 Build-00001
#+
#+ Filename.: sadzp030_tab_ma
#+ Buildtype: p01樣板
#+ Memo.....: 由4fd分析資料搜尋出單頭資訊
#+          : 180402-00053 ka0132 修正chr4000無法正常產出
#+          : 180418-00005 ka0132 修正chr80無法正常產出
#+          : 180411-00062 yscheng 調整多語言設定資料寫入的陣列
#+          : 180619-00034 ka0132 增加chr40對應,增加遇錯處理
#+          : 180820-00028 ka0132 修正num20對應
#+          : 190826-00045 02667  修正ref參照型態
#+          : 191211-00021 02667  增加DateTimeEdit控件判斷
#+          : 200210-00032 00824  因SQL Server的欄位型態與ORACLE不同，需調整sadzp030_tab_ma_datatype2中取得欄位定義的方法
#+          : 200220-00036 00824  tab產生時，若是date型態的欄位，會給char(500)的資料型態
#+          : 201015-00004 00824  調整產tab檔時referance參照欄位的取法，不可直接取單身檔的最大版次，需對應到單頭才可取到正確的版次
#+          : 201109-00009 00824  模糊查詢功能
#+          : 200803-00005 00824  (210511-00011 代追單) 因為postgre環境部分欄位屬性改為int4，但長度變空白了
#+          : 240930-00022 00824  因應項標中心設計器優化需求，配合調整相關工具及樣板
#+          : 250919-00017 00824  金倉版本資料庫針對int4的欄位屬性，也需比照postgre體系處理


SCHEMA ds

GLOBALS "../../cfg/top_global.inc"
GLOBALS "../4gl/adzp150_global.inc"   #180411-00062

GLOBALS
   DEFINE g_dzac_m          DYNAMIC ARRAY OF RECORD
            type            LIKE type_t.chr1,      #單頭與單身的標示 M/D
            field           LIKE dzaa_t.dzaa003,   #欄位名稱/識別碼
            dzaa003         LIKE dzaa_t.dzaa003,   #欄位名稱
            dzaa004         LIKE dzaa_t.dzaa004,   #欄位版本/識別碼版次
            dzaa005         LIKE dzaa_t.dzaa005,   #識別類型/欄位類型
            dzaa006         LIKE dzaa_t.dzaa006,   #使用標示/客製碼s/c
            dzaa007         LIKE dzaa_t.dzaa007,   #引用
         #---------------------------------------------------------------#
            dzac005         LIKE dzac_t.dzac005,   #table編號
            dzac002         LIKE dzac_t.dzac002,   #欄位編號
            dzeb006         LIKE dzeb_t.dzeb006,   #欄位屬性
            dzac009         LIKE dzac_t.dzac009,   #編輯時開查詢視窗
            dzac010         LIKE dzac_t.dzac010,   #查詢時開查詢視窗
            dzac011         LIKE dzac_t.dzac011,   #校驗帶值設定
            dzac014         LIKE dzac_t.dzac014,   #預設值
            dzac015         LIKE dzac_t.dzac015,   #最大值
            dzac020         LIKE dzac_t.dzac020,   #最大值符號
            dzac016         LIKE dzac_t.dzac016,   #最小值
            dzac021         LIKE dzac_t.dzac021,   #最小值符號
            dzac017         LIKE dzac_t.dzac017,   #是否可編輯
            dzac018         LIKE dzac_t.dzac018,   #是否可查詢
         #---------------------------------------------------------------#
            dzak005         LIKE dzak_t.dzak005,   #助記碼-其他條件
            dzak007         LIKE dzak_t.dzak007,   #助記碼-助記碼搜尋欄位
            dzak008         LIKE dzak_t.dzak008,   #助記碼-助記碼table
            dzak009         LIKE dzak_t.dzak009,   #助記碼-助記碼欄位
            dzak010         LIKE dzak_t.dzak010,   #助記碼-助記碼語系
            dzak011         LIKE dzak_t.dzak011,   #助記碼-回傳對應控件
         #---------------------------------------------------------------#
            dzep011         LIKE dzep_t.dzep011    #是否有設定default scc(下拉式選單)
                            END RECORD
   DEFINE g_prog_id         LIKE dzfi_t.dzfi001
   DEFINE g_sd_ver          LIKE dzaa_t.dzaa002  #規格版次/程式版次
#  DEFINE g_erp_ver         LIKE dzaa_t.dzaa008  #產品版本
   DEFINE g_dgenv           LIKE dzaa_t.dzaa009  #客製標示
#  DEFINE g_cust            LIKE dzaa_t.dzaa010  #客戶編號
   DEFINE g_prog_type       STRING
   DEFINE g_cross_4gl_stus  LIKE type_t.num5     #跨4gl判斷正確與否
   DEFINE g_datalang DYNAMIC ARRAY OF RECORD #單頭
            columnid   LIKE dzaa_t.dzaa003,  #使用欄位
            columns    STRING,
            relay_on   STRING,
            map_col    STRING,
            ref_table  STRING,
            ref_fk     STRING,
            ref_lang   STRING,   #180411-00062
            ref_rtn    STRING
          END RECORD
   DEFINE g_mdatalang DYNAMIC ARRAY OF RECORD #單身
            page       LIKE type_t.num5,     #使用頁次
            columnid   LIKE dzaa_t.dzaa003,  #使用欄位
            columns    STRING,
            relay_on   STRING,
            map_col    STRING,
            ref_table  STRING,
            ref_fk     STRING,
            ref_lang   STRING,   #180411-00062
            ref_rtn    STRING,
            used       LIKE type_t.chr1
          END RECORD
   #180411-00062 ---start---
   DEFINE g_multilang_m DYNAMIC ARRAY OF RECORD #單頭
            columnid   LIKE dzaa_t.dzaa003,  #使用欄位
            columns    STRING,
            relay_on   STRING,
            map_col    STRING,
            ref_table  STRING,
            ref_fk     STRING,
            ref_lang   STRING,
            ref_rtn    STRING
          END RECORD
   DEFINE g_multilang_d DYNAMIC ARRAY OF RECORD #單身
            page       LIKE type_t.num5,     #使用頁次
            columnid   LIKE dzaa_t.dzaa003,  #使用欄位
            columns    STRING,
            relay_on   STRING,
            map_col    STRING,
            ref_table  STRING,
            ref_fk     STRING,
            ref_lang   STRING,
            ref_rtn    STRING,
            used       LIKE type_t.chr1
          END RECORD
   #180411-00062 --- end ---
   DEFINE gc_dzal     DYNAMIC ARRAY OF RECORD
             dzal002  LIKE dzal_t.dzal002,  #ON ACTION的 action-id (按鈕ID)
             dzal005  LIKE dzal_t.dzal005,  #串查作業名稱
             dzal006  LIKE dzal_t.dzal006,  #串查作業名稱
             dzal007  LIKE dzal_t.dzal007   #串查型態
                      END RECORD
   DEFINE g_db_branch STRING   #250919-00017
END GLOBALS

############################################################
#+ @code
#+ 函式目的 初始化
############################################################
PUBLIC FUNCTION sadzp030_tab_ma_init()

   WHENEVER ERROR CALL cl_err_msg_log #180619-00034

   CALL g_dzac_m.clear()
   CALL gc_dzal.clear()

END FUNCTION

#+ 函式目的 從column id 找 table id
PUBLIC FUNCTION sadzp030_tab_ma_find_tableid(lc_gztz002)
   DEFINE lc_gztz001  LIKE gztz_t.gztz001
   DEFINE lc_gztz002  LIKE gztz_t.gztz002
   DEFINE li_p        LIKE type_t.num5

   SELECT COUNT(1) INTO li_p FROM gztz_t
    WHERE gztz002 = lc_gztz002
   IF li_p = 1 THEN 
      SELECT gztz001 INTO lc_gztz001 FROM gztz_t
       WHERE gztz002 = lc_gztz002
   ELSE
      LET lc_gztz001 = ""
   END IF
   RETURN lc_gztz001

END FUNCTION

############################################################
#+ @code
#+ 函式目的 依照需求回傳欄位名稱組合
#+          回傳型態=1  傳回 DEFINE 的需求   reference用欄位補上 (chr80)
#+          回傳型態=2  傳回 INPUT 的需求    去除非LABEL的欄位
#+          回傳型態=3  傳回 CONSTRUCT 的需求 去除非主表的欄位
#+          回傳型態=4  傳回 sql用欄位  非主表的全部用 ''代換
#+          回傳型態=5  傳回 widget整串欄位名稱  4fd處理處使用
#+          回傳型態=6  Q04類單頭傳回 DEFINE 的需求   reference用欄位補上 (chr80)
#+          回傳型態=7  傳回 CONSTRUCT 的需求 去除非主表的欄位，只取文字欄位 (模糊查詢功能用)   #201109-00009
#+          回傳型態=8  傳回 INPUT 的需求 目標控件(簡易帶值) 僅保留有設定簡易帶值或公式的欄位 (簡易帶值/公式功能用)   #240930-00022
#+          回傳型態=9  傳回 INPUT 的需求 使用控件(簡易帶值) 僅保留有設定簡易帶值或公式的欄位 (簡易帶值/公式功能用)   #240930-00022
#+          回傳型態=10 傳回 INPUT 的需求 目標控件(公式) 僅保留有設定簡易帶值或公式的欄位 (簡易帶值/公式功能用)   #240930-00022
#+          回傳型態=11 傳回 INPUT 的需求 使用控件(公式) 僅保留有設定簡易帶值或公式的欄位 (簡易帶值/公式功能用)   #240930-00022
#+ @param   pc_mastername  char(20)  容器名稱
#+ @param   pi_type        number(5) 回傳型態
#+ @param   pc_tablename   chr(20)   主表
#+ @param   pi_qbe         number(5) 是否在Q類的qbe容器內(有特殊處理)
#+
#+ @return  STRING  單頭欄位敘述字串
############################################################
PUBLIC FUNCTION sadzp030_tab_ma(pc_mastername,pi_type,pc_tablename,pi_qbe)

#  DEFINE pi_type        LIKE type_t.chr1   #240930-00022 mark
   DEFINE pi_type        LIKE type_t.num5   #240930-00022 modify
   DEFINE pi_qbe         LIKE type_t.num5
   DEFINE pc_tablename   LIKE type_t.chr20  
   DEFINE sr             DYNAMIC ARRAY OF RECORD
            dzfi004      LIKE dzfi_t.dzfi004,
            dzfi006      LIKE dzfi_t.dzfi006,
            dzfi007      LIKE dzfi_t.dzfi007
                         END RECORD
   DEFINE pc_mastername  LIKE dzfi_t.dzfi004  #群組代碼
   DEFINE lc_dzfi005     LIKE dzfi_t.dzfi005  #順序
   DEFINE lc_dzfi006     LIKE dzfi_t.dzfi006  #元件組代碼
   DEFINE lc_dzfj004     LIKE dzfj_t.dzfj004  #順序
   DEFINE lc_dzfj005     LIKE dzfj_t.dzfj005
   DEFINE lc_dzam003     LIKE dzam_t.dzam003  #排除設定
   DEFINE lc_dzfj006     LIKE dzfj_t.dzfj006
   DEFINE ls_dzfj005     STRING
   DEFINE ls_tablename   STRING
   DEFINE ls_fldname     STRING
   DEFINE ls_str         STRING
   DEFINE ls_str1        STRING
   DEFINE li_cnt,i       LIKE type_t.num5
   DEFINE li_deny        LIKE type_t.num5
   DEFINE lc_gztz001     LIKE gztz_t.gztz001 #table名稱
   DEFINE lc_gztz002     LIKE gztz_t.gztz002 #欄位名稱
   DEFINE li_gztzcnt     LIKE type_t.num5    #數量
   DEFINE lc_temp        LIKE type_t.chr1    #flag暫存
   DEFINE lc_dzac005     LIKE dzac_t.dzac005 #table編號
   DEFINE lc_dzac002     LIKE dzac_t.dzac002 #欄位編號
   DEFINE lc_dzeb006     LIKE dzeb_t.dzeb006 #欄位屬性
   DEFINE lc_dzaa003     LIKE dzaa_t.dzaa003
   DEFINE ls_tmpp01      STRING
   DEFINE ls_dzai002     LIKE dzai_t.dzai002 #控件編號 #190826-00045#1
   DEFINE ls_dzai011     LIKE dzai_t.dzai011 #參照欄位 #190826-00045#1 
   DEFINE ls_dzai008     LIKE dzai_t.dzai008 #參照表   #190826-00045#1 
   DEFINE li_count       LIKE type_t.num5    #201109-00009
   DEFINE li_dzeb001     LIKE dzeb_t.dzeb001 #201109-00009
   DEFINE li_dzeb002     LIKE dzeb_t.dzeb002 #201109-00009

   
   LET ls_str = ''
 
   #抓取所有下階資料放到array
   DECLARE tab_ma_1_cs CURSOR FOR
    SELECT dzfi004,dzfi005,dzfi006,dzfi007 FROM dzfi_t
     WHERE dzfi001 = g_prog_id    #畫面代碼
       AND dzfi002 = g_sd_ver     #規格版次
       AND dzfi009 = g_dgenv      #客製標示
       AND dzfi004 = pc_mastername
#      AND dzfi017 = g_cust       #客戶編號
     ORDER BY dzfi005
   LET li_cnt = 1
   FOREACH tab_ma_1_cs INTO sr[li_cnt].dzfi004,lc_dzfi005,sr[li_cnt].dzfi006,sr[li_cnt].dzfi007
      LET li_cnt = li_cnt + 1
   END FOREACH
   CALL sr.deleteElement(li_cnt)

   LET li_cnt = sr.getLength()
   IF li_cnt = 0 AND pc_mastername = "vb_qbe" THEN
      DISPLAY cl_getmsg("adz-01136",g_lang)
     #DISPLAY "ERROR: 程式使用p01/r01樣板時, 4fd必須存在 vb_qbe 的VBOX控件, 目前規格內不存在!"
      LET g_cross_4gl_stus = FALSE     #跨4gl判斷正確與否
      RETURN ls_str
   END IF

   FOR i = 1 TO li_cnt
      #遇到s_borwse(樣板的)/s_detail(樣板的)/s_tree(自定義的TREE)
      # s_relateapps(UI固化)/s_queryplan(UI固化)
      #開頭的資料，就不繼續往下
      IF sr[i].dzfi006 = "s_browse" OR sr[i].dzfi006[1,8] = "s_detail" OR
         sr[i].dzfi006 = "s_tree" OR sr[i].dzfi006 = "s_relateapps" OR
         sr[i].dzfi006 = "s_queryplan" OR
         sr[i].dzfi006 = "patch_grid" THEN  #這個是patch產生的,若有需要使用要額外union
         CONTINUE FOR
      END IF

      #Folder/Page/Grid/Label/Item/HBox/VBox/Table/Group/Button，不撈取dzfj_t的欄位資料
      IF sr[i].dzfi007 = "Folder" OR sr[i].dzfi007 = "Page"  OR sr[i].dzfi007 = "Grid"  OR
         sr[i].dzfi007 = "Label"  OR sr[i].dzfi007 = "Item"  OR sr[i].dzfi007 = "HBox"  OR
         sr[i].dzfi007 = "VBox"   OR sr[i].dzfi007 = "Button" OR sr[i].dzfi007 = "Group" THEN
       # sr[i].dzfi007 = "Table" THEN  #開放讓PR自行添加Table元件
      ELSE
         DECLARE tab_ma_1_cs1 CURSOR FOR
          SELECT dzfj004,dzfj005,dzfj006 FROM dzfj_t
           WHERE dzfj001 = g_prog_id    #畫面代碼
             AND dzfj002 = g_sd_ver     #規格版次
             AND dzfj003 = sr[i].dzfi006
             AND dzfj017 = g_dgenv      #客製標示
#            AND dzfj018 = g_cust       #客戶編號
           ORDER BY dzfj004

         FOREACH tab_ma_1_cs1 INTO lc_dzfj004,lc_dzfj005,lc_dzfj006
            IF lc_dzfj006 = "Edit" OR lc_dzfj006 = "ButtonEdit" OR lc_dzfj006 = "ComboBox" OR
               lc_dzfj006 = "CheckBox" OR lc_dzfj006 = "DateEdit" OR lc_dzfj006 = "RadioGroup" OR
               lc_dzfj006 = "DateTimeEdit" OR #191211-00021 add
               lc_dzfj006 = "Slider" OR lc_dzfj006 = "SpinEdit" OR lc_dzfj006 = "TimeEdit" OR
               lc_dzfj006 = "Text" OR lc_dzfj006 = "TextEdit" OR lc_dzfj006 = "FFLabel" THEN

               #201109-00009 ---start---
               #模糊查詢須排除非文字欄位 或是 不會加上*的文字欄位
               IF pi_type = 7 THEN
                  IF lc_dzfj006 = "DateTimeEdit" OR lc_dzfj006 = "TimeEdit" OR lc_dzfj006 = "DateEdit" OR
                     lc_dzfj006 = "CheckBox" OR lc_dzfj006 = "RadioGroup" OR lc_dzfj006 = "ComboBox" OR
                     lc_dzfj006 = "FFLabel" OR lc_dzfj006 = "Slider" OR lc_dzfj006 = "SpinEdit" THEN

                     CONTINUE FOREACH
                  END IF
               END IF
               #201109-00009 --- end ---
            ELSE
               CONTINUE FOREACH
            END IF

            #濾除一些畫面上的公用標準欄位(不須輸入,樣板都有)
            IF lc_dzfj005 = "h_index" OR lc_dzfj005 = "h_count" OR lc_dzfj005 = "idx" OR
               lc_dzfj005 = "cnt" THEN
               CONTINUE FOREACH
            END IF

            #處理pi_type = 5
            IF pi_type = 5 THEN
               IF cl_null(ls_str) THEN
                  LET ls_str = lc_dzfj005 CLIPPED
               ELSE
                  LET ls_str = ls_str CLIPPED,",",lc_dzfj005 CLIPPED
               END IF
               CONTINUE FOREACH   #組合完成就離開,不處理其他項目
            END IF

            LET ls_dzfj005 = lc_dzfj005

            #抓取該欄位的table name
            LET ls_tablename = ls_dzfj005.subString(1,ls_dzfj005.getIndexOf(".",1)-1)
           
            #抓取欄位代號
            LET ls_fldname = ls_dzfj005.subString(ls_dzfj005.getIndexOf(".",1)+1,ls_dzfj005.getLength())
            IF cl_null(ls_fldname) THEN
               LET ls_fldname = ls_dzfj005
            END IF

            #查看是否此欄位已經被設置成排除
           #LET lc_dzam003 = ls_fldname
            SELECT COUNT(1) INTO li_deny FROM dzaa_t,dzam_t 
             WHERE dzaa001 = g_prog_id   #規格編號
               AND dzaa002 = g_sd_ver    #規格版次
               AND dzaa003 = "EXCLUDE"   #識別代碼
               AND dzaa005 = "a"         #識別碼類型
#              AND dzaa008 = g_erp_ver   #產品版本
               AND dzaa009 = g_dgenv     #客製標示
#              AND dzaa010 = g_cust      #客戶編號
               AND dzaastus = "Y"        #有效碼
             #-----------------------------------------------
               AND dzam001 = dzaa001     #規格編號
               AND dzam003 = lc_dzfj005  #排除項目
               AND dzam004 = dzaa004     #識別碼版次
               AND dzam005 = dzaa006     #使用標示
#              AND dzam006 = dzaa010     #客戶編號
               AND dzamstus = "Y"        #有效碼
            IF li_deny > 0 THEN
               IF g_t100debug >= 3 THEN
                  DISPLAY "INFO: 擴充欄位:",lc_dzfj005
               END IF
               CONTINUE FOREACH
            END IF

            #如果是q類樣板
            IF g_prog_type.subString(1,1) = "q" THEN

               LET lc_dzaa003 = ls_fldname
               
               #ka0132
               #若為b_xxxxx 要去除b_
               #IF lc_dzaa003[1,2] = "b_" THEN
               #   LET lc_dzaa003 = cl_str_replace(lc_dzaa003,"b_","")
               #   DISPLAY "lc_dzaa003-m1:",lc_dzaa003
               #ELSE 
               #   DISPLAY "lc_dzaa003-m2:",lc_dzaa003
               #END IF
               #ka0132
               
               #因為q類設定s_detail內均以COLUMN_LIKE方式呈現,所以若b_開頭的,
               #要進行ls_tablename/ls_fldname的回歸抓取
               IF ls_fldname.subString(1,2) = "b_" AND ls_tablename IS NULL THEN
                  SELECT dzac005 INTO lc_dzac005 FROM dzaa_t,dzac_t
                   WHERE dzaa001 = g_prog_id   #程式編號/規格編號
                     AND dzaa002 = g_sd_ver    #規格版號
                     AND dzaa003 = lc_dzaa003  #識別碼
                     AND dzaa005 = "1"         #識別碼類型
#                    AND dzaa008 = g_erp_ver   #產品版本
                     AND dzaa009 = g_dgenv     #客製標示
#                    AND dzaa010 = g_cust      #客戶編號
                     AND dzaastus = "Y"        #狀態取 Y 有效
                   #-----------------------------------------------#
                     AND dzac001 = dzaa001     #規格編號
                     AND dzac003 = dzaa003     #識別碼
                     AND dzac004 = dzaa004     #規格版次
                     AND dzac012 = dzaa006     #使用標示
#                    AND dzac013 = dzaa010     #客戶編號
                  IF NOT SQLCA.SQLCODE THEN
                     LET ls_tablename = lc_dzac005
                     LET ls_fldname = ls_fldname.subString(3,ls_fldname.getLength())
                  END IF
               ELSE
                  LET ls_tablename = ""
               END IF

               IF ls_fldname.subString(1,2) = "b_" THEN
                  LET ls_fldname = ls_fldname.subString(3,ls_fldname.getLength())
               END IF
            END IF

            #因為p01/r01的規格並不會特別去設定table資料,因此若 pc_tablename 為NULL要特別處理
            IF g_prog_type = "p01" OR g_prog_type = "r01" THEN
               LET ls_tmpp01 = lc_dzfj005
               LET pc_tablename = ls_tmpp01.subString(1,ls_tmpp01.getIndexOf("_t",1)+1)
            END IF 

            #若無table name(reference)或table name不等同主表
            IF cl_null(ls_tablename) OR ls_tablename <> pc_tablename THEN
               CASE 
                  # ==>回傳型態= 1時傳回 DEFINE 的需求，欄位後面接 (chr80)
                  WHEN pi_type = 1 
                     IF lc_dzfj006 = "FFLabel" THEN
                        #190826-00045#1 -S-
                        #抓取對應的欄位型態,找不到則指定chr80
                        LET ls_dzai002 = ls_fldname
                        LET ls_dzai011 = "" #先清空

                        #201015-00004 ---modify start---
                     #  SELECT dzai008,dzai011 INTO ls_dzai008,ls_dzai011
                     #    FROM dzai_t 
                     #   WHERE dzai001 = g_prog_id  #版本
                     #     AND dzai002 = ls_dzai002 #控件編號
                     #     AND dzai003 =            #識別碼版次
                     #     (SELECT MAX(dzai003) FROM dzai_t 
                     #       WHERE dzai001 = g_prog_id
                     #         AND dzai002 = ls_dzai002
                     #         AND dzai004 = g_dgenv)
                     #     AND dzai004 = g_dgenv    #客製標示

                     #  #組合為 表格.欄位 的格式
                     #  LET ls_dzai011 = ls_dzai008||"."||ls_dzai011
                     #  #如果沒找到資料則用chr80
                     #  IF NOT cl_null(ls_dzai011) THEN
                     #     LET ls_fldname = ls_fldname CLIPPED,"(",ls_dzai011,")"
                     #  ELSE
                     #     LET ls_fldname = ls_fldname CLIPPED,"(chr80)"
                     #  END IF

                        #先找欄位參考的referance檔案
                        SELECT dzai008,dzai011 INTO ls_dzai008,ls_dzai011
                          FROM dzaa_t,dzai_t
                         WHERE dzaa001 = g_prog_id   #程式編號/規格編號
                           AND dzaa002 = g_sd_ver    #規格版次
                           AND dzaa005 = "6"         #識別碼類型
                           AND dzaa009 = g_dgenv     #客製標示
                           AND dzaastus = "Y"        #狀態碼
                           AND dzai002 = ls_dzai002  #控件編號
                        #----------------------------------------------#
                           AND dzaa001 = dzai001     #規格編號
                           AND dzaa003 = dzai002     #識別碼
                           AND dzaa004 = dzai003     #規格版次
                           AND dzaa006 = dzai004     #使用標示

                        #組合為 表格.欄位 的格式
                        LET ls_dzai011 = ls_dzai008||"."||ls_dzai011
                        #如果沒找到資料則用chr80
                        IF NOT cl_null(ls_dzai011) THEN
                           LET ls_fldname = ls_fldname CLIPPED,"(",ls_dzai011,")"
                        ELSE
                           #若找不到，可能該referance是獨立的FFLABEL
                           IF cl_null(ls_tablename) THEN
                              SELECT dzac002,dzac005 INTO lc_dzac002,lc_dzac005
                                FROM dzaa_t,dzac_t
                               WHERE dzaa001 = g_prog_id   #程式編號/規格編號
                                 AND dzaa002 = g_sd_ver    #規格版號
                                 AND dzaa003 = ls_dzai002  #識別碼
                                 AND dzaa005 = "1"         #識別碼類型
                              #  AND dzaa008 = g_erp_ver   #產品版本
                                 AND dzaa009 = g_dgenv     #客製標示
                              #  AND dzaa010 = g_cust      #客戶編號
                                 AND dzaastus = "Y"        #狀態取 Y 有效
                               #-----------------------------------------------#
                                 AND dzac001 = dzaa001     #規格編號
                                 AND dzac003 = dzaa003     #識別碼
                                 AND dzac004 = dzaa004     #規格版次
                                 AND dzac012 = dzaa006     #使用標示
                              #  AND dzac013 = dzaa010     #客戶編號

                              IF SQLCA.SQLCODE THEN
                                 IF g_t100debug >= 3 THEN
                                    DISPLAY "注意: ",ls_dzai002,"在第",g_sd_ver,"版次的資料不存在,改用chr80"
                                 END IF
                                 LET ls_fldname = ls_fldname CLIPPED,"(chr80)"
                              ELSE
                                 IF cl_null(lc_dzac002) OR cl_null(lc_dzac005) THEN
                                    IF g_t100debug >= 3 THEN
                                       DISPLAY "注意: ",ls_dzai002,"在第",g_sd_ver,"版次的資料為空,改用chr80"
                                    END IF
                                    LET ls_fldname = ls_fldname CLIPPED,"(chr80)"
                                 ELSE
                                    LET ls_dzai011 = lc_dzac005||"."||lc_dzac002
                                    LET ls_fldname = ls_fldname CLIPPED,"(",ls_dzai011,")"
                                 END IF
                              END IF
                           ELSE
                              #存在table name的話
                              IF g_t100debug >= 6 THEN
                                 DISPLAY "INFO: 欄位(",ls_fldname,")使用",ls_tablename,"表及",ls_fldname,"做為資料型態"
                              END IF
                              LET ls_fldname = ls_fldname CLIPPED,sadzp030_tab_ma_datatype(ls_tablename,ls_fldname)
                           END IF
                        END IF
                        #201015-00004 --- modify end ---
                        #190826-00045#1 -E-
                        CALL sadzp030_tab_gen_reference_array(ls_fldname)
                     ELSE
                        #原來就不存在table name
                        IF cl_null(ls_tablename) THEN

                           LET lc_dzaa003 = ls_fldname

                           #補充規則:EDIT也會有_desc
                           IF lc_dzfj006 = "Edit" THEN
                              IF ls_fldname.subString(ls_fldname.getLength()-4,ls_fldname.getLength()) = "_desc" THEN 
                                 CALL sadzp030_tab_gen_reference_array(ls_fldname)
                              END IF
                           END IF

                           SELECT dzac002,dzac005 INTO lc_dzac002,lc_dzac005
                             FROM dzaa_t,dzac_t
                            WHERE dzaa001 = g_prog_id   #程式編號/規格編號
                              AND dzaa002 = g_sd_ver    #規格版號
                              AND dzaa003 = lc_dzaa003  #識別碼
                              AND dzaa005 = "1"         #識別碼類型
#                             AND dzaa008 = g_erp_ver   #產品版本
                              AND dzaa009 = g_dgenv     #客製標示
#                             AND dzaa010 = g_cust      #客戶編號
                              AND dzaastus = "Y"        #狀態取 Y 有效
                            #-----------------------------------------------#
                              AND dzac001 = dzaa001     #規格編號
                              AND dzac003 = dzaa003     #識別碼
                              AND dzac004 = dzaa004     #規格版次
                              AND dzac012 = dzaa006     #使用標示
#                             AND dzac013 = dzaa010     #客戶編號
                           IF SQLCA.SQLCODE THEN
                              LET lc_dzeb006 = "C302"
                              IF g_t100debug >= 3 THEN
                                 DISPLAY "注意: ",lc_dzaa003,"在第",g_sd_ver,"版次的資料不存在,改用chr80"
                              END IF
                           END IF

                           #從table與column再轉欄位型態
                           SELECT dzeb006 INTO lc_dzeb006
                             FROM dzeb_t
                            WHERE dzeb001 = lc_dzac005 AND dzeb002 = lc_dzac002
                           IF SQLCA.SQLCODE THEN
                              LET lc_dzeb006 = "C302"
                              IF g_t100debug >= 3 THEN
                                 DISPLAY "注意: ",lc_dzaa003,"在第",g_sd_ver,"版次的資料不存在,改用chr80"
                              END IF
                           END IF

                           IF g_t100debug >= 6 THEN
                              DISPLAY "INFO: 欄位(",ls_fldname,")使用",lc_dzeb006,"做為資料型態"
                           END IF
                           IF NOT pi_qbe THEN
                              LET ls_fldname = ls_fldname CLIPPED,sadzp030_tab_ma_datatype2(lc_dzeb006)
                           END IF
                        ELSE
                           #存在table name的話
                           IF g_t100debug >= 6 THEN
                              DISPLAY "INFO: 欄位(",ls_fldname,")使用",ls_tablename,"表及",ls_fldname,"做為資料型態"
                           END IF
                           LET ls_fldname = ls_fldname CLIPPED,sadzp030_tab_ma_datatype(ls_tablename,ls_fldname)
                        END IF
                     END IF

                  # ==>回傳型態= 2時傳回 INPUT 的需求，CONTINUE FOREACH
                  WHEN pi_type = 2
                     IF lc_dzfj006 = "FFLabel" THEN
                        CONTINUE FOREACH
                     END IF
                     #在設計器上有勾選可以輸入的話,就組出來
                     LET lc_temp = sadzp030_tab_ma_2(ls_fldname,"dzac017")
                     IF lc_temp IS NOT NULL THEN
                        IF lc_temp = "Y" OR lc_temp = "y" THEN
                        ELSE
                           CONTINUE FOREACH
                        END IF
                     ELSE
                        CONTINUE FOREACH
                     END IF

                  # ==>回傳型態= 3時傳回 CONSTRUCT 的需求，CONTINUE FOREACH
                  WHEN pi_type = 3
                     #在設計器上有勾選可以查詢的話,就組出來
                     LET lc_temp = sadzp030_tab_ma_2(ls_fldname,"dzac018")
                     IF lc_temp IS NOT NULL THEN
                        IF lc_temp = "Y" OR lc_temp = "y" THEN
                        ELSE
                           CONTINUE FOREACH
                        END IF
                     ELSE
                        CONTINUE FOREACH
                     END IF

                  # ==>回傳型態= 4時傳回 sql用欄位，欄位改為''
                  WHEN pi_type = 4
                     LET ls_fldname = "''"

                  # ==>回傳型態= 6時傳回Q類的DEFINE 用欄位
                  WHEN pi_type = 6
                     LET ls_fldname = ls_fldname,"(",ls_dzfj005 CLIPPED,")"

                  #201109-00009 ---start---
                  # ==>回傳型態= 7時傳回 CONSTRUCT 的需求 (模糊查詢用)，CONTINUE FOREACH
                  WHEN pi_type = 7
                     #在設計器上有勾選可以查詢的話,就組出來
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
                              #  AND dzaa008 = g_erp_ver   #產品版本
                                 AND dzaa009 = g_dgenv     #客製標示
                              #  AND dzaa010 = g_cust      #客戶編號
                                 AND dzaastus = "Y"        #狀態取 Y 有效
                               #-----------------------------------------------#
                                 AND dzac001 = dzaa001     #規格編號
                                 AND dzac003 = dzaa003     #識別碼
                                 AND dzac004 = dzaa004     #規格版次
                                 AND dzac012 = dzaa006     #使用標示
                              #  AND dzac013 = dzaa010     #客戶編號
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
                                    CONTINUE FOREACH
                                 END IF
                              ELSE
                                 SELECT COUNT(1) INTO li_count
                                   FROM dzeb_t
                                  WHERE dzeb001 = lc_dzac005
                                    AND dzeb002 = lc_dzac002
                                    AND dzeb007 IN ('varchar','varchar2','clob','blob')   #模糊查詢只針對文字型態欄位
                                 IF li_count = 0 THEN
                                    CONTINUE FOREACH
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
                                 CONTINUE FOREACH
                              END IF
                           END IF
                        ELSE
                           CONTINUE FOREACH
                        END IF
                     ELSE
                        CONTINUE FOREACH
                     END IF
                  #201109-00009 --- end ---

                  #240930-00022 ---start---
                  # ==>回傳型態= 8時傳回 INPUT 的需求 (簡易帶值用)，CONTINUE FOREACH
                  WHEN pi_type = 8   #目標控件
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
                           CONTINUE FOREACH
                        END IF
                     ELSE
                        CONTINUE FOREACH
                     END IF

                  # ==>回傳型態= 9時傳回 INPUT 的需求 (簡易帶值用)，CONTINUE FOREACH
                  WHEN pi_type = 9   #使用控件
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
                           CONTINUE FOREACH
                        END IF
                     ELSE
                        CONTINUE FOREACH
                     END IF

                  # ==>回傳型態= 10時傳回 INPUT 的需求 (公式用)，CONTINUE FOREACH
                  WHEN pi_type = 10   #目標控件
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
                           CONTINUE FOREACH
                        END IF
                     ELSE
                        CONTINUE FOREACH
                     END IF

                  # ==>回傳型態= 11時傳回 INPUT 的需求 (公式用)，CONTINUE FOREACH
                  WHEN pi_type = 11   #使用控件
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
                           CONTINUE FOREACH
                        END IF
                     ELSE
                        CONTINUE FOREACH
                     END IF
                  #240930-00022 --- end ---

               END CASE
            ELSE
               CASE 
                  #有table name,做define的時候也檢查一下資料庫是否真有這個欄位
                  WHEN pi_type = 1 OR pi_type = 6
                     LET lc_gztz001 = pc_tablename
                     LET lc_gztz002 = ls_fldname
                     SELECT COUNT(1) INTO li_gztzcnt FROM gztz_t
                      WHERE gztz001 = lc_gztz001 AND gztz002 = lc_gztz002
                     IF li_gztzcnt < 1 THEN
                        #此處應該中止執行
                        DISPLAY cl_getmsg_parm("adz-01137",g_lang,lc_dzfj005)
                       #DISPLAY "ERROR: 畫面上TABLE_COLUMN欄位(",lc_dzfj005,")不存在資料表內"
                        LET g_cross_4gl_stus = FALSE     #跨4gl判斷正確與否
                     END IF
                     IF pi_type = 6 THEN
                        LET ls_fldname = ls_fldname,"(",ls_dzfj005 CLIPPED,")"
                     END IF

                  WHEN pi_type = 2   #取INPUT
                     #在設計器上沒勾選,就不做後面的事
                     LET lc_temp = sadzp030_tab_ma_2(ls_fldname,"dzac017")
                     IF lc_temp IS NOT NULL THEN
                        IF lc_temp = "N" THEN
                           CONTINUE FOREACH
                        END IF
                     END IF

                  WHEN pi_type = 3   #取CONSTRUCT
                     #在設計器上沒勾選,就不做後面的事
                     LET lc_temp = sadzp030_tab_ma_2(ls_fldname,"dzac018")
                     IF lc_temp IS NOT NULL THEN
                        IF lc_temp = "N" THEN
                           CONTINUE FOREACH
                        END IF
                     END IF

                  #201109-00009 ---start---
                  WHEN pi_type = 7   #模糊查詢
                     #在設計器上沒勾選,就不做後面的事
                     LET lc_temp = sadzp030_tab_ma_2(ls_fldname,"dzac018")
                     IF lc_temp IS NOT NULL THEN
                        IF lc_temp = "N" THEN
                           CONTINUE FOREACH
                        ELSE
                           LET li_dzeb001 = ls_tablename
                           LET li_dzeb002 = ls_fldname
                           LET li_count = 0
                           SELECT COUNT(1) INTO li_count FROM dzeb_t
                            WHERE dzeb001 = li_dzeb001
                              AND dzeb002 = li_dzeb002
                              AND dzeb007 IN ('varchar','varchar2','clob','blob')   #模糊查詢只針對文字型態欄位
                           IF li_count = 0 THEN
                              CONTINUE FOREACH
                           END IF
                        END IF
                     ELSE
                        CONTINUE FOREACH
                     END IF
                  #201109-00009 --- end ---

                  #240930-00022 ---start---
                  WHEN pi_type = 8   #簡易帶值   #目標控件
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
                           CONTINUE FOREACH
                        END IF
                     ELSE
                        CONTINUE FOREACH
                     END IF

                  WHEN pi_type = 9   #簡易帶值   #使用控件
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
                           CONTINUE FOREACH
                        END IF
                     ELSE
                        CONTINUE FOREACH
                     END IF

                  WHEN pi_type = 10   #公式   #目標控件
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
                           CONTINUE FOREACH
                        END IF
                     ELSE
                        CONTINUE FOREACH
                     END IF

                  WHEN pi_type = 11   #公式   #使用控件
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
                           CONTINUE FOREACH
                        END IF
                     ELSE
                        CONTINUE FOREACH
                     END IF
                  #240930-00022 --- end ---

                  OTHERWISE
               END CASE
            END IF

            #240930-00022 ---modify start---
         #  #將欄位ls_fldname組進字串ls_str內
         #  IF cl_null(ls_str) THEN
         #     LET ls_str = ls_fldname CLIPPED
         #  ELSE
         #     LET ls_str = ls_str CLIPPED,",",ls_fldname CLIPPED
         #  END IF

            #將欄位ls_fldname組進字串ls_str內
            IF pi_type = 8 OR pi_type = 9 OR pi_type = 10 OR pi_type = 11 THEN
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
         END FOREACH
      END IF

      CLOSE tab_ma_1_cs1
      FREE tab_ma_1_cs1

      #遞迴呼叫
      IF NOT cl_null(sr[i].dzfi006) THEN
         CALL sadzp030_tab_ma(sr[i].dzfi006,pi_type,pc_tablename,pi_qbe) RETURNING ls_str1
      END IF
  
      IF cl_null(ls_str) THEN
         LET ls_str = ls_str1
      ELSE
         IF NOT cl_null(ls_str1) THEN
            LET ls_str = ls_str,",",ls_str1
         END IF
      END IF
   END FOR

   RETURN ls_str
END FUNCTION

#+ 從欄位名稱抓定義回來 (給define用,傳回type_t內的欄位名稱)
PUBLIC FUNCTION sadzp030_tab_ma_datatype(lc_gztz001,lc_gztz002)
   DEFINE lc_gztz001   LIKE gztz_t.gztz001 #表格名稱
   DEFINE lc_gztz002   LIKE gztz_t.gztz002 #欄位名稱
   DEFINE lc_gztz003   LIKE gztz_t.gztz003 #欄位名稱
   DEFINE lc_gztz004   LIKE gztz_t.gztz004 #欄位名稱
   DEFINE ls_return    STRING

   IF cl_null(lc_gztz001) OR cl_null(lc_gztz002) THEN
      LET ls_return = "(chr80)"
      IF g_t100debug >= 3 THEN
         DISPLAY cl_getmsg("adz-01138",g_lang)
        #DISPLAY "注意: 抓取定義名稱時未取得傳入資料,datatype直接判為chr80"
      END IF
      RETURN ls_return
   END IF

   SELECT gztz003,gztz004 INTO lc_gztz003,lc_gztz004
     FROM gztz_t
    WHERE gztz001 = lc_gztz001
      AND gztz002 = lc_gztz002
   IF SQLCA.SQLCODE OR cl_null(lc_gztz003) OR cl_null(lc_gztz004) THEN
      IF g_t100debug >= 3 THEN
         DISPLAY "注意: 表",lc_gztz001,"中的欄位",lc_gztz002,"取用型態有問題,判為chr80"
      END IF
      LET ls_return = "(chr80)"
      RETURN ls_return
   END IF 

   #CASE
   #   WHEN lc_gztz003 = "0" OR lc_gztz003 = "13" OR lc_gztz003 = "15" OR
   #        lc_gztz003 = "16" OR lc_gztz003 = "201" OR lc_gztz003 = "202"
   #      CASE
   #         WHEN  1  = lc_gztz004                       LET ls_return = "(chr1)"
   #         WHEN  1  < lc_gztz004 AND lc_gztz004 <=   5 LET ls_return = "(chr5)"
   #         WHEN  5  < lc_gztz004 AND lc_gztz004 <=  10 LET ls_return = "(chr10)"
   #         WHEN  10 < lc_gztz004 AND lc_gztz004 <=  20 LET ls_return = "(chr20)"
   #         WHEN  20 < lc_gztz004 AND lc_gztz004 <= 100 LET ls_return = "(chr100)"
   #         WHEN 100 < lc_gztz004                       LET ls_return = "(chr500)"
   #      END CASE
   #   WHEN lc_gztz003 = "7" OR lc_gztz003 = "10"  LET ls_return = "(dat)"
   #   OTHERWISE LET ls_return = "(chr80)"
   #END CASE

   RETURN ls_return

END FUNCTION

#+ 從欄位型態抓定義回來 (給define用,傳回type_t內的欄位名稱)
PUBLIC FUNCTION sadzp030_tab_ma_datatype2(lc_dzeb006)

   DEFINE lc_dzeb006   LIKE dzeb_t.dzeb006   #欄位型態
   DEFINE lc_gztd003   LIKE gztd_t.gztd003
   DEFINE lc_gztd008   LIKE gztd_t.gztd008
   DEFINE ls_return    STRING
   DEFINE lm_dzeb002   LIKE dzeb_t.dzeb002   #200210-00032
   DEFINE ls_sql       STRING   #200220-00036
   DEFINE lc_gztd009   LIKE gztd_t.gztd009   #250919-00017


   #如果沒有傳入資料,則一律回傳chr80
   IF cl_null(lc_dzeb006) THEN
      LET ls_return = "(chr80)"
      IF g_t100debug >= 3 THEN
         DISPLAY cl_getmsg("adz-01139",g_lang)
        #DISPLAY "注意: DEFINE取型態未取得傳入,直接判為chr80"
      END IF
      RETURN ls_return
   END IF

   #從gztd型態檔對應
#  SELECT gztd003,gztd008 INTO lc_gztd003,lc_gztd008   #250919-00017 mark
   SELECT gztd003,gztd008,gztd009 INTO lc_gztd003,lc_gztd008,lc_gztd009   #250919-00017 modify
     FROM gztd_t 
    WHERE gztd001 = lc_dzeb006 AND gztdstus = "Y"

   IF SQLCA.SQLCODE OR lc_gztd003 IS NULL OR lc_gztd008 IS NULL THEN
      #下列的型態(Binary/Text/Date/Datetime)的長度都不會填寫,所以濾除
      IF lc_dzeb006 = "B001" OR lc_dzeb006 = "B002" OR
         lc_dzeb006 = "D001" OR lc_dzeb006 = "D002" OR lc_dzeb006 = "T091" THEN
      ELSE
         #250919-00017 ---mark start---
      #  #200803-00005 (210511-00011 代追單)---start---
      #  &ifdef is_pgs
      #  #因為postgre環境下num5的型態會改為int4，且gztd008欄位為空，
      #  #因此須排排除此狀況
      #  IF lc_gztd003 = "int4" AND lc_gztd008 IS NULL THEN
      #  ELSE
      #  &endif
      #  #200803-00005 (210511-00011 代追單)--- end ---
      #     LET ls_return = "(chr80)"
      #     IF g_t100debug >= 3 THEN
      #        DISPLAY "注意: 欄位取用到型態",lc_dzeb006 CLIPPED,"發生問題,用chr80取代"
      #     END IF
      #     RETURN ls_return
      #  &ifdef is_pgs
      #  END IF   #200803-00005 (210511-00011 代追單)
      #  &endif
         #250919-00017 --- mark end ---

         #250919-00017 ---modify start---
         #因為postgre體系(含MogDB)及人大金倉版下，num5的型態會改為int4，且gztd008欄位為空，
         #因此須排排除此狀況
         IF lc_gztd003 = "int4" AND lc_gztd008 IS NULL THEN
            CASE
               WHEN g_db_branch = "POSTGRESQL" OR g_db_branch = "MOGDB" OR g_db_branch = "KINGBASE"

               OTHERWISE   #(ex：ORACLE、DAMONG)
                  LET ls_return = "(chr80)"
                  IF g_t100debug >= 3 THEN
                     DISPLAY "注意: 欄位取用到型態",lc_dzeb006 CLIPPED,"發生問題,用chr80取代"
                  END IF
                  RETURN ls_return
            END CASE
         ELSE
            LET ls_return = "(chr80)"
            IF g_t100debug >= 3 THEN
               DISPLAY "注意: 欄位取用到型態",lc_dzeb006 CLIPPED,"發生問題,用chr80取代"
            END IF
            RETURN ls_return
         END IF
         #250919-00017 --- modify end ---
      END IF
   END IF

   #200210-00032 ---modify start---
   #依據傳入的欄位型態取得type_t的欄位對應
   LET lm_dzeb002 = NULL

   #200220-00036 ---modify start---
#  SELECT dzeb002 INTO lm_dzeb002 FROM dzeb_t
#   WHERE dzeb001 = 'type_t'
#   AND dzeb007 = lc_gztd003
#   AND dzeb008 = lc_gztd008

   LET ls_sql = "SELECT dzeb002 FROM dzeb_t,gztd_t",
                " WHERE dzeb001 = 'type_t'",
                  " AND dzeb002 NOT LIKE '%ud%'",   #須排除掉自訂義(ud)的欄位
                  " AND gztd003 = '",lc_gztd003,"'",
                  " AND gztd001 = dzeb006 "

   #250919-00017 ---modify start---
#  #200803-00005 (210511-00011 代追單)---start---
#  &ifdef is_pgs
#  #因為int4在type_t中有多個欄位對應，所以若是int4的型態，就需要再加上欄位屬性的條件，確保可抓到正確的資料
#  IF lc_gztd003 = "int4" AND lc_gztd008 IS NULL THEN
#     LET ls_sql = ls_sql, " AND gztd001 = '",lc_dzeb006,"'"   #200803-00005
#  END IF
#  &endif
#  #200803-00005 (210511-00011 代追單)--- end ---


   #因為int4在type_t中有多個欄位對應，所以若是int4的型態，就需要再加上欄位屬性的條件，確保可抓到正確的資料
   IF g_db_branch = "POSTGRESQL" OR g_db_branch = "MOGDB" OR g_db_branch = "KINGBASE" THEN
      IF lc_gztd003 = "int4" AND lc_gztd008 IS NULL THEN
      #  LET ls_sql = ls_sql, " AND gztd001 = '",lc_dzeb006,"'"   #200803-00005
         LET ls_sql = ls_sql, " AND gztd001 LIKE 'T%' AND gztd009 = '",lc_gztd009,"'"
      END IF
   END IF
   #250919-00017 --- modify end ---

   #因為date的型態，gztd008會是NULL，若是null會影響sql執行，所以須先做判斷排除
   IF NOT cl_null(lc_gztd008) THEN
      LET ls_sql = ls_sql, " AND gztd008 = '",lc_gztd008,"'"
   END IF
   
   PREPARE sadzp030_tab_ma_dzeb002_pre FROM ls_sql
   EXECUTE sadzp030_tab_ma_dzeb002_pre INTO lm_dzeb002
   #200220-00036 --- modify end ---
   
   IF NOT cl_null(lm_dzeb002) THEN
      LET ls_return = "(",lm_dzeb002,")"
   ELSE
      #若真的都對應不到，就給chr500
      LET ls_return = "(chr500)"
   END IF

#  #轉換處理
#  CASE
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "1" LET ls_return = "(chr1)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "2" LET ls_return = "(chr2)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "3" LET ls_return = "(chr3)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "4" LET ls_return = "(chr4)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "5" LET ls_return = "(chr5)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "6" LET ls_return = "(chr6)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "7" LET ls_return = "(chr7)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "8" LET ls_return = "(chr8)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "9" LET ls_return = "(chr9)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "10" LET ls_return = "(chr10)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "12" LET ls_return = "(chr12)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "14" LET ls_return = "(chr14)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "18" LET ls_return = "(chr18)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "20" LET ls_return = "(chr20)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "21" LET ls_return = "(chr21)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "30" LET ls_return = "(chr30)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "37" LET ls_return = "(chr37)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "50" LET ls_return = "(chr50)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "50" LET ls_return = "(chr80)" #180418-00005#1
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "100" LET ls_return = "(chr100)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "200" LET ls_return = "(chr200)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "300" LET ls_return = "(chr300)"
#     #180402-00053#1
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "2000" LET ls_return = "(chr2000)"
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "4000" LET ls_return = "(chr4000)"
#     #180402-00053#1
#     WHEN lc_gztd003 = "varchar2"  AND lc_gztd008 = "40" LET ls_return = "(chr50)" #180619-00034#1
#     WHEN lc_gztd003 = "varchar2" 
#        LET ls_return = "(chr500)"
#     WHEN lc_gztd003 = "number"  AND (lc_gztd008 = "5" OR lc_gztd008 = "5,0")
#        LET ls_return = "(num5)"
#     WHEN lc_gztd003 = "number"  AND (lc_gztd008 = "10" OR lc_gztd008 = "10,0")
#        LET ls_return = "(num10)"
#     WHEN lc_gztd003 = "number"  AND (lc_gztd008 = "20" OR lc_gztd008 = "20,0")
#        #LET ls_return = "(num10)" #180820-00028#1
#        LET ls_return = "(num20)" #180820-00028#1
#     WHEN lc_gztd003 = "number"  AND (lc_gztd008 = "15,3" OR lc_gztd008 = "5,3" OR
#                                      lc_gztd008 = "7,3" )
#        LET ls_return = "(num15_3)"
#     WHEN lc_gztd003 = "number"  AND (lc_gztd008 = "20,6" OR lc_gztd008 = "20, 6")
#        LET ls_return = "(num20_6)"
#     WHEN lc_gztd003 = "number" 
#        LET ls_return = "(num26_10)"

#     WHEN lc_gztd003 = "date"       LET ls_return = "(dat)"
#     WHEN lc_gztd003 = "timestamp"  LET ls_return = "(dat)"

#     OTHERWISE
#        LET ls_return = "(",lc_gztd008 CLIPPED,")"
#  END CASE
   #200210-00032 --- modify end ---

   RETURN ls_return
END FUNCTION

############################################################
#+ @code
#+ 函式目的 解析傳入的欄位清單，抓取欄位屬性，放入 array 內
#+ @param   ls_fields      STRING    欄位清單
#+ @param   lc_type        char(1)   使用標示
#+ @param   li_page        SMALLINT  頁次
#+
#+ @return  BOOLEAN        TRUE/FALSE
############################################################
PUBLIC FUNCTION sadzp030_tab_ma_1(ls_fields,lc_type,li_page)

   DEFINE ls_fields   STRING
   DEFINE lc_fld      LIKE type_t.chr50
   DEFINE lc_fld_bak  LIKE type_t.chr50
   DEFINE li_cnt      LIKE type_t.num10
   DEFINE li_p        LIKE type_t.num5
   DEFINE li_dzal     LIKE type_t.num5
   DEFINE li_page     LIKE type_t.num5
   DEFINE l_token     base.StringTokenizer
   DEFINE ls_next     STRING
   DEFINE lc_dzaj005  LIKE dzaj_t.dzaj005   #依附控件編號
   DEFINE lc_dzaj007  LIKE dzaj_t.dzaj007   #對應原始表欄位
   DEFINE lc_dzaj008  LIKE dzaj_t.dzaj008   #欄位多語言設定表
   DEFINE lc_dzaj009  LIKE dzaj_t.dzaj009   #欄位多語言設定表
   DEFINE lc_dzaj010  LIKE dzaj_t.dzaj010   #欄位多語言設定表
   DEFINE lc_dzaj011  LIKE dzaj_t.dzaj011   #欄位多語言設定表
   DEFINE li_succ     LIKE type_t.num5
   DEFINE lc_type     LIKE type_t.chr1  #M:傳入為單頭,不處理串查
                                        #D:傳入為單身,回傳串查 action id
   DEFINE ls_sql      STRING
   DEFINE ls_tmp      STRING
   DEFINE lc_dzep001  LIKE dzep_t.dzep001   #規格=表格編號
   DEFINE lc_dzep002  LIKE dzep_t.dzep002   #規格=欄位編號
   DEFINE ls_return   STRING
   DEFINE lc_gzza008  LIKE gzza_t.gzza008   #引用對象
#  DEFINE lc_cite_dzaa002 LIKE dzaa_t.dzaa002  #引用規格版次(取最大值)
#  DEFINE lc_cite_dzaa004 LIKE dzaa_t.dzaa004  #引用識別碼版次
#  DEFINE lc_cite_dzaa006 LIKE dzaa_t.dzaa006  #引用使用標示

   #若為空值則退出
   IF cl_null(ls_fields) THEN RETURN "" END IF

   LET li_cnt = g_dzac_m.getLength() + 1
   LET l_token = base.StringTokenizer.create(ls_fields, ",")

   #回圈內可能會用到的CURSOR都先進行PREPARED
   LET ls_sql = "SELECT dzal002,dzal005,dzal006,dzal007 ",
                 " FROM dzal_t ",
                " WHERE dzal001 = '",g_prog_id CLIPPED,"' ",  #規格編號
                  " AND dzal005 = ? ",                        #依附控件編號
                  " AND dzal003 = ? ",                        #識別碼版次
                  " AND dzal004 = ? ",                        #使用標示
#                 " AND dzal009 = '",g_cust CLIPPED,"' ",     #客戶編號
                  " AND dzalstus = 'Y' ",
                " ORDER BY dzal008 "                        #順序
   DECLARE sadzp030_tab_ma_dzal_cs CURSOR FROM ls_sql

   WHILE l_token.hasMoreTokens()
      #抓取每個欄位
      LET ls_next = l_token.nextToken()

      #若為''，表示沒有欄位
      IF ls_next = "''" THEN
         CONTINUE WHILE
      END IF

      #檢查是否存在 (括號, 半邊就可以, 有則去除
      IF ls_next.getIndexOf("(",1) THEN
         LET ls_next = ls_next.subString(1,ls_next.getIndexOf("(",1)-1)
      END IF

      LET lc_fld = ls_next
      #當欄位不為空時，抓取資料庫該欄位的屬性

#SQL> run
#  1* select dzaa003,dzaa005,dzaa004,dzaa006 from dzaa_t  where dzaa001='azzi110' and dzaa003 like '%gzoz000%'
#
#DZAA003 	     DZAA005	   DZAA004 D
#-------------------- ---------- ---------- -
#gzoz000_2	     1			 1 s
#gzoz_t.gzoz000	     1			 1 s
#gzoz000_2	     1			 1 s
#gzoz_t.gzoz000	     1			 1 s
#gzoz000_2	     1			 1 s
#gzoz_t.gzoz000	     1			 5 s
#gzoz000_2	     1			 1 s
#gzoz000_2	     1			 1 s
#
#8 rows selected.

      IF NOT cl_null(lc_fld) THEN

         #q類樣板,如果原來就帶入b_開頭的,就表示這應該是formonly欄位,不在增加b_開頭 
         IF g_prog_type.subString(1,1) = "q" THEN
            IF ls_next = "sel" THEN
               CONTINUE WHILE
            END IF
            IF NOT ls_next.subString(1,2) = "b_" AND NOT ls_next.subString(1,2) = "l_" THEN
               LET lc_fld = "b_",lc_fld
            END IF
         END IF

         LET g_dzac_m[li_cnt].field = lc_fld
         LET g_dzac_m[li_cnt].type  = lc_type

         #若傳進來的 filed 名稱已含有底線,這一定不是表的欄位(行業表也沒有)
         IF ls_next.getIndexOf("_",2) THEN

            #確認 dzaa003欄位名稱, dzaa004設計器版號,dzaa005欄位類型,dzaa006使用標示
            #     dzaa007引用
            SELECT dzaa003,dzaa004,dzaa005,dzaa006,dzaa007
              INTO g_dzac_m[li_cnt].dzaa003,g_dzac_m[li_cnt].dzaa004,
                   g_dzac_m[li_cnt].dzaa005,g_dzac_m[li_cnt].dzaa006,
                   g_dzac_m[li_cnt].dzaa007
              FROM dzaa_t
             WHERE dzaa001 = g_prog_id   #程式編號/規格編號
               AND dzaa002 = g_sd_ver    #規格版號
               AND dzaa003 = lc_fld      #識別碼
#              AND dzaa008 = g_erp_ver   #產品版本
               AND dzaa009 = g_dgenv     #客製標示
#              AND dzaa010 = g_cust      #客戶編號
               AND dzaastus = "Y"        #狀態取 Y 有效
         ELSE
            #調整lc_fld配合dzac設置, browser部分除外
            IF lc_fld[1,2] <> "b_" THEN
               LET lc_fld_bak = lc_fld
               LET lc_fld = "%.",lc_fld
            END IF
            #如果是q類,調整lc_fld_bak改為加上 b_
            IF g_prog_type.subString(1,1) = "q" THEN
               LET lc_fld_bak = "b_",lc_fld_bak
            END IF

            #確認 dzaa003欄位名稱, dzaa004設計器版號,dzaa005欄位類型,dzaa006使用標示
            #     dzaa007引用
            SELECT dzaa003,dzaa004,dzaa005,dzaa006,dzaa007
              INTO g_dzac_m[li_cnt].dzaa003,g_dzac_m[li_cnt].dzaa004,
                   g_dzac_m[li_cnt].dzaa005,g_dzac_m[li_cnt].dzaa006,
                   g_dzac_m[li_cnt].dzaa007
              FROM dzaa_t
             WHERE dzaa001 = g_prog_id   #程式編號/規格編號
               AND dzaa002 = g_sd_ver    #規格版號
               AND dzaa003 LIKE lc_fld   #識別碼
#              AND dzaa008 = g_erp_ver   #產品版本
               AND dzaa009 = g_dgenv     #客製標示
#              AND dzaa010 = g_cust      #客戶編號
               AND dzaastus = "Y"        #狀態取 Y 有效

            #再做一次看看
            IF sqlca.sqlcode THEN
               SELECT dzaa003,dzaa004,dzaa005,dzaa006,dzaa007
                 INTO g_dzac_m[li_cnt].dzaa003,g_dzac_m[li_cnt].dzaa004,
                      g_dzac_m[li_cnt].dzaa005,g_dzac_m[li_cnt].dzaa006,
                      g_dzac_m[li_cnt].dzaa007
                 FROM dzaa_t
                WHERE dzaa001 = g_prog_id   #程式編號/規格編號
                  AND dzaa002 = g_sd_ver    #規格版號
                  AND dzaa003 = lc_fld_bak  #識別碼
#                 AND dzaa008 = g_erp_ver   #產品版本
                  AND dzaa009 = g_dgenv     #客製標示
#                 AND dzaa010 = g_cust      #客戶編號
                  AND dzaastus = "Y"        #狀態取 Y 有效
            END IF
         END IF

         #如果找不到規格,顯示錯誤訊息後做下一筆
         IF sqlca.sqlcode THEN
            IF g_t100debug >= 3 THEN
               DISPLAY "注意: 規格",g_dzac_m[li_cnt].dzaa003,"(",lc_fld,")資料不存在dzaa_t表(",sqlca.sqlcode,")"
            END IF
            LET li_cnt = li_cnt + 1
            CONTINUE WHILE
         END IF

         #引用的處理:如果發現dzaa007="Y"要引用的話
         #1.從gzza008找到主要關注的對象
         #2.找原來的設計資料的最大規格版號
         #3.用引用對象dzaa001+最大規格版號002+原有識別碼003=找出識別碼版次004+標示006
#         IF g_dzac_m[li_cnt].dzaa006 = "Y" THEN
#
#            #從gzza008找到主要關注的對象
#            SELECT gzza008 INTO lc_gzza008 FROM gzza_t
#             WHERE gzza001 = g_prog_id
#             # AND gzaastus = "Y" #此處不可以加，因為可能最大版次是不要enable的
#
#            #找原來的設計資料的最大規格版號
#            SELECT MAX(unique dzaa002) INTO lc_cite_dzaa002 FROM dzaa_t
#             WHERE dzaa001 = lc_gzza008
#             # AND dzaastus = "Y" #此處不可以加，因為可能最大版次是不要enable的
#
#            #用引用對象dzaa001+最大規格版號002+原有識別碼003=找出識別碼版次004+標示006
#            SELECT dzaa004,dzaa006 INTO lc_cite_dzaa004,lc_cite_dzaa006 FROM dzaa_t
#             WHERE dzaa001 = lc_gzza008       #引用對象
#               AND dzaa002 = lc_cite_dzaa002  #最大規格版號
#               AND dzaa003 = lc_fld_bak       #識別碼
#               AND dzaastus = "Y"             #狀態取 Y 有效
#         END IF

         #抓取規格細節
         CASE 
            #一般欄位 dzaa005="1"
            WHEN g_dzac_m[li_cnt].dzaa005 = "1" 
               #查詢細部規格
               SELECT dzac005,dzac002,dzac009,dzac010,dzac011,dzac014,
                      dzac015,dzac020,dzac016,dzac021,dzac017,dzac018
                 INTO g_dzac_m[li_cnt].dzac005,g_dzac_m[li_cnt].dzac002,
                      g_dzac_m[li_cnt].dzac009,g_dzac_m[li_cnt].dzac010,g_dzac_m[li_cnt].dzac011,
                      g_dzac_m[li_cnt].dzac014,
                      g_dzac_m[li_cnt].dzac015,g_dzac_m[li_cnt].dzac020,
                      g_dzac_m[li_cnt].dzac016,g_dzac_m[li_cnt].dzac021,
                      g_dzac_m[li_cnt].dzac017,g_dzac_m[li_cnt].dzac018
                 FROM dzac_t
                WHERE dzac001 = g_prog_id
                  AND dzac003 = g_dzac_m[li_cnt].dzaa003  #欄位名稱
                  AND dzac004 = g_dzac_m[li_cnt].dzaa004  #設計器版號
                  AND dzac012 = g_dzac_m[li_cnt].dzaa006  #使用標示
#                 AND dzac013 = g_cust                    #客戶編號
               IF sqlca.sqlcode THEN 
                  IF g_t100debug >= 3 THEN
                     DISPLAY "注意: 一般欄位",g_dzac_m[li_cnt].dzaa003,"(",lc_fld,")資料不存在dzac_t表"
                  END IF
               ELSE
                  #查詢欄位屬性 (從dzac005,dzac002)
                  SELECT dzeb006 INTO g_dzac_m[li_cnt].dzeb006  #欄位屬性
                    FROM dzeb_t
                   WHERE dzeb001 = g_dzac_m[li_cnt].dzac005   #table編號
                     AND dzeb002 = g_dzac_m[li_cnt].dzac002   #欄位編號

                  #查詢欄位其他項目 系統分類碼
                  LET ls_tmp = g_dzac_m[li_cnt].dzaa003
                  IF ls_tmp.getIndexOf(".",1) THEN
                     LET lc_dzep001 = ls_tmp.subString(1,ls_tmp.getIndexOf(".",1)-1)
                     LET lc_dzep002 = ls_tmp.subString(ls_tmp.getIndexOf(".",1)+1,ls_tmp.getLength())
                     #應該檢核該系統分類碼是否存在SCC項目內
                     SELECT dzep011 INTO g_dzac_m[li_cnt].dzep011 #系統分類碼
                       FROM dzep_t
                      WHERE dzep001 = lc_dzep001
                        AND dzep002 = lc_dzep002
                  END IF

                  #查詢欄位其他項目 欄位註記碼
                  SELECT dzak005,dzak007,dzak008,dzak009,dzak010,dzak011
                    INTO g_dzac_m[li_cnt].dzak005, g_dzac_m[li_cnt].dzak007,
                         g_dzac_m[li_cnt].dzak008, g_dzac_m[li_cnt].dzak009,
                         g_dzac_m[li_cnt].dzak010, g_dzac_m[li_cnt].dzak011
                    FROM dzak_t
                   WHERE dzak001 = g_prog_id
                     AND dzak002 = g_dzac_m[li_cnt].dzaa003  #控件編號
                     AND dzak003 = g_dzac_m[li_cnt].dzaa004  #設計器版次
                     AND dzak004 = g_dzac_m[li_cnt].dzaa006  #使用標示
#                    AND dzak012 = g_cust                    #客戶編號

                  #查詢欄位其他項目 程式串查(多選)
                  LET li_dzal = gc_dzal.getLength() + 1
                  FOREACH sadzp030_tab_ma_dzal_cs USING g_dzac_m[li_cnt].dzaa003,
                                                        g_dzac_m[li_cnt].dzaa004,
                                                        g_dzac_m[li_cnt].dzaa006
                     INTO gc_dzal[li_dzal].dzal002,gc_dzal[li_dzal].dzal005,
                          gc_dzal[li_dzal].dzal006,gc_dzal[li_dzal].dzal007
                     IF SQLCA.SQLCODE THEN
                        EXIT FOREACH
                     END IF
                     LET gc_dzal[li_dzal].dzal006 = gc_dzal[li_dzal].dzal006 CLIPPED
                     IF gc_dzal[li_dzal].dzal006 IS NULL THEN
                        IF g_t100debug >= 6 THEN
                           DISPLAY "INFO: 欄位",gc_dzal[li_dzal].dzal002,"存在串查資料,但沒有程式參考設定,by pass"
                        END IF
                        CALL gc_dzal.deleteElement(li_dzal)
                        CONTINUE FOREACH
                     END IF
                     LET li_dzal = li_dzal + 1
                  END FOREACH
                  CALL gc_dzal.deleteElement(li_dzal)

                  #如果是Q類,就砍除b_看一下
                  IF g_prog_type.subString(1,1) = "q" THEN
                     IF ls_tmp.subString(1,2) = "b_" THEN
                        LET lc_dzep002 = ls_tmp.subString(3,ls_tmp.getLength())
                        SELECT COUNT(1) INTO li_p FROM gztz_t
                         WHERE gztz002 = lc_dzep002
                        IF li_p = 1 THEN 
                           SELECT gztz001 INTO lc_dzep001 FROM gztz_t
                            WHERE gztz002 = lc_dzep002

                           #應該檢核該系統分類碼是否存在SCC項目內
                           SELECT dzep011 INTO g_dzac_m[li_cnt].dzep011 #系統分類碼
                             FROM dzep_t
                            WHERE dzep001 = lc_dzep001
                              AND dzep002 = lc_dzep002
                        END IF
                     END IF
                  END IF 
               END IF

            #欄位多語言 dzaj005="7"
            WHEN g_dzac_m[li_cnt].dzaa005 = "7" 
               SELECT dzaj005,dzaj007,dzaj008,dzaj009,dzaj010,dzaj011
                 INTO lc_dzaj005,lc_dzaj007,lc_dzaj008,lc_dzaj009,lc_dzaj010,lc_dzaj011
                 FROM dzaj_t
                WHERE dzaj001 = g_prog_id
                  AND dzaj002 = g_dzac_m[li_cnt].dzaa003  #欄位名稱
                  AND dzaj003 = g_dzac_m[li_cnt].dzaa004  #設計器版號
                  AND dzaj004 = g_dzac_m[li_cnt].dzaa006  #使用標示
#                 AND dzaj012 = g_cust                    #客戶編號
               IF NOT SQLCA.SQLCODE THEN
                  LET li_succ = FALSE

                  #因為多語言欄位(非reference)一定是要輸入的,給予輸入選項
                  LET g_dzac_m[li_cnt].dzac017 = "Y"  #是否可輸入
                  LET g_dzac_m[li_cnt].dzac018 = "Y"  #是否可查詢

                  #單頭放入 g_datalang / 單身放入 g_mdatalang
                  IF lc_type = "M" THEN
                     #180411-00062 ---start---
                     #因為原先只要多語言表格及對應傳值一樣的話，就視為是同一個多語言項，
                     #以前是一起處理，現在必須要分開處理，因此不可再依多語言表格及對應傳值項判斷是否重複
                     LET li_p = g_multilang_m.getLength() + 1
                     LET g_multilang_m[li_p].columnid  = g_dzac_m[li_cnt].dzaa003
                     LET g_multilang_m[li_p].columns   = g_dzac_m[li_cnt].dzaa003
                     LET g_multilang_m[li_p].relay_on  = lc_dzaj005  #依附欄位
                     LET g_multilang_m[li_p].map_col   = sadzp030_tab_relation_ent(lc_dzaj007)  #對應欄位
                     LET g_multilang_m[li_p].ref_table = lc_dzaj008  #多語言table
                     LET g_multilang_m[li_p].ref_fk    = lc_dzaj009, #多語言FK
                                                     ",",lc_dzaj010  #多語言語言別欄位
                     LET g_multilang_m[li_p].ref_fk = sadzp030_tab_relation_ent(g_multilang_m[li_p].ref_fk)
                     LET g_multilang_m[li_p].ref_lang  = lc_dzaj010  #多語言欄位
                     LET g_multilang_m[li_p].ref_rtn   = lc_dzaj011  #多語言回傳
                     #180411-00062 --- end ---

                     FOR li_p = 1 TO g_datalang.getLength()
                        IF g_datalang[li_p].ref_table = lc_dzaj008 AND
                           g_datalang[li_p].map_col   = sadzp030_tab_relation_ent(lc_dzaj007) THEN #對應欄位
                           LET li_succ = TRUE
                           EXIT FOR
                        END IF
                     END FOR

                     IF NOT li_succ THEN
                        LET li_p = g_datalang.getLength() + 1
                        LET g_datalang[li_p].columnid  = g_dzac_m[li_cnt].dzaa003
                        LET g_datalang[li_p].columns   = g_dzac_m[li_cnt].dzaa003
                        LET g_datalang[li_p].relay_on  = lc_dzaj005  #依附欄位
                        LET g_datalang[li_p].map_col   = sadzp030_tab_relation_ent(lc_dzaj007)  #對應欄位
                        LET g_datalang[li_p].ref_table = lc_dzaj008  #多語言table
                        LET g_datalang[li_p].ref_fk    = lc_dzaj009, #多語言FK
                                                     ",",lc_dzaj010  #多語言語言別欄位
                        LET g_datalang[li_p].ref_fk = sadzp030_tab_relation_ent(g_datalang[li_p].ref_fk)
                        LET g_datalang[li_p].ref_lang  = lc_dzaj010  #多語言欄位   #180411-00062
                        LET g_datalang[li_p].ref_rtn   = lc_dzaj011  #多語言回傳
                     ELSE
                        #同表收集再一起
                        IF NOT g_datalang[li_p].ref_rtn.getIndexOf(lc_dzaj011,1) THEN
                           LET g_datalang[li_p].ref_rtn = g_datalang[li_p].ref_rtn,",",lc_dzaj011
                        END IF
                        IF NOT g_datalang[li_p].columns.getIndexOf(g_dzac_m[li_cnt].dzaa003,1) THEN
                           LET g_datalang[li_p].columns = g_datalang[li_p].columns,",",g_dzac_m[li_cnt].dzaa003
                        END IF
                     END IF
                  ELSE
                     #180411-00062 ---start---
                     #因為原先只要多語言表格及對應傳值一樣的話，就視為是同一個多語言項，
                     #以前是一起處理，現在必須要分開處理，因此不可再依多語言表格及對應傳值項判斷是否重複
                     LET li_p = g_multilang_d.getLength() + 1
                     LET g_multilang_d[li_p].page      = li_page
                     LET g_multilang_d[li_p].columnid  = g_dzac_m[li_cnt].dzaa003
                     LET g_multilang_d[li_p].columns   = g_dzac_m[li_cnt].dzaa003
                     LET g_multilang_d[li_p].relay_on  = lc_dzaj005  #依附欄位
                     LET g_multilang_d[li_p].map_col   = sadzp030_tab_relation_ent(lc_dzaj007)  #對應欄位
                     LET g_multilang_d[li_p].ref_table = lc_dzaj008  #多語言table
                     LET g_multilang_d[li_p].ref_fk    = lc_dzaj009, #多語言FK
                                                     ",",lc_dzaj010  #多語言語言別欄位
                     LET g_multilang_d[li_p].ref_fk = sadzp030_tab_relation_ent(g_multilang_d[li_p].ref_fk)
                     LET g_multilang_d[li_p].ref_lang  = lc_dzaj010  #多語言欄位
                     LET g_multilang_d[li_p].ref_rtn   = lc_dzaj011  #多語言回傳
                     #180411-00062 --- end ---

                     FOR li_p = 1 TO g_mdatalang.getLength()
                        IF g_mdatalang[li_p].ref_table = lc_dzaj008 AND
                           g_mdatalang[li_p].map_col   = sadzp030_tab_relation_ent(lc_dzaj007) THEN #對應欄位
                           LET li_succ = TRUE
                           EXIT FOR
                        END IF
                     END FOR

                     IF NOT li_succ THEN
                        LET li_p = g_mdatalang.getLength() + 1
                        LET g_mdatalang[li_p].page      = li_page
                        LET g_mdatalang[li_p].columnid  = g_dzac_m[li_cnt].dzaa003
                        LET g_mdatalang[li_p].columns   = g_dzac_m[li_cnt].dzaa003
                        LET g_mdatalang[li_p].relay_on  = lc_dzaj005  #依附欄位
                        LET g_mdatalang[li_p].map_col   = sadzp030_tab_relation_ent(lc_dzaj007)  #對應欄位
                        LET g_mdatalang[li_p].ref_table = lc_dzaj008  #多語言table
                        LET g_mdatalang[li_p].ref_fk    = lc_dzaj009, #多語言FK
                                                     ",",lc_dzaj010  #多語言語言別欄位
                        LET g_mdatalang[li_p].ref_fk = sadzp030_tab_relation_ent(g_mdatalang[li_p].ref_fk)
                        LET g_mdatalang[li_p].ref_lang  = lc_dzaj010  #多語言欄位   #180411-00062
                        LET g_mdatalang[li_p].ref_rtn   = lc_dzaj011  #多語言回傳
                     ELSE
                        #同表收集再一起
                        IF NOT g_mdatalang[li_p].ref_rtn.getIndexOf(lc_dzaj011,1) THEN
                           LET g_mdatalang[li_p].ref_rtn = g_mdatalang[li_p].ref_rtn,",",lc_dzaj011
                        END IF
                        IF NOT g_mdatalang[li_p].columns.getIndexOf(g_dzac_m[li_cnt].dzaa003,1) THEN
                           LET g_mdatalang[li_p].columns = g_mdatalang[li_p].columns,",",g_dzac_m[li_cnt].dzaa003
                        END IF
                     END IF
                  END IF
               END IF
         END CASE

         #單身串查檢查
         IF lc_type = "D" THEN
            FOR li_dzal = 1 TO gc_dzal.getLength()
               IF g_dzac_m[li_cnt].dzaa003 = gc_dzal[li_dzal].dzal005 THEN
                  #檢查一下,如果已經有存在了就不要再加進去
                  IF NOT ls_return.getIndexOf(gc_dzal[li_dzal].dzal006||"(",1) THEN
                     LET ls_return = ls_return,gc_dzal[li_dzal].dzal006 CLIPPED,"(",
                                               gc_dzal[li_dzal].dzal005 CLIPPED,"),"
                  END IF
               END IF
            END FOR
         END IF

         LET li_cnt = li_cnt + 1
      END IF
   END WHILE

   IF ls_return.getLength() > 0 THEN
      LET ls_return = ls_return.subString(1,ls_return.getLength()-1)
   END IF

   RETURN ls_return
 
END FUNCTION


############################################################
#+ @code
#+ 函式目的 更換指定欄位的屬性
############################################################
PUBLIC FUNCTION sadzp030_tab_ma_change_type(ls_columns,lc_type)
   DEFINE ls_columns STRING
   DEFINE ls_next    STRING
   DEFINE ls_temp    STRING
   DEFINE ls_return  STRING
   DEFINE li_cnt     LIKE type_t.num5
   DEFINE l_token    base.StringTokenizer
   DEFINE lc_type    LIKE type_t.chr1
   DEFINE lc_dzal002 LIKE dzal_t.dzal002

   LET l_token = base.StringTokenizer.create(ls_columns, ",")

   WHILE l_token.hasMoreTokens()
      LET ls_next = l_token.nextToken()

      #如果欄位沒有特別註記,轉去查一下這個欄位的型態是否為特殊形態
      IF NOT ls_next.getIndexOf("(",1) THEN
         FOR li_cnt = 1 TO g_dzac_m.getLength()
            IF g_prog_type.subString(1,1) = "q" THEN 
               LET ls_temp = "b_",ls_next
            ELSE
               LET ls_temp = ls_next
            END IF

            IF g_dzac_m[li_cnt].field = ls_TEMP THEN
               CASE
             #    WHEN g_dzac_m[li_cnt].dzeb006 = "D002"   #欄位屬性
             #       LET ls_next = ls_next,"(datetime)"
                  WHEN g_dzac_m[li_cnt].dzeb006 = "D003"   #欄位屬性
                     LET ls_next = ls_next,"(timestamp)"
                  OTHERWISE
               END CASE
            END IF
         END FOR
      ELSE
         #如果有特別的註記(表示不是table-column),確認一下是否為Q類(找串查)
         IF lc_type = "D" THEN
         IF g_prog_type.subString(1,1) = "q" THEN
            LET lc_dzal002 = ls_next.subString(1,ls_next.getIndexOf("(",1)-1)
            SELECT COUNT(1) INTO li_cnt FROM dzal_t
             WHERE dzal001 = g_prog_id 
               AND dzal002 = lc_dzal002
#              AND dzal009 = g_cust     #客戶編號
               AND dzalstus = 'Y' 
            IF li_cnt = 1 THEN
               LET ls_next = lc_dzal002 CLIPPED,"(string)"
            END IF
         END IF
         END IF
      END IF

      LET ls_return = ls_return,",",ls_next
   END WHILE

   RETURN ls_return.subString(2,ls_return.getLength())
END FUNCTION

############################################################
#+ @code
#+ 函式目的 查詢欄位的屬性值：從 array 內抓出屬性
#+ @param   lc_field       char(20)  欄位
#+ @param   lc_attribute   char(10)  需要的元素
#+
#+ @return  STRING         屬性值
############################################################
PUBLIC FUNCTION sadzp030_tab_ma_2(ls_field,lc_attribute)

   DEFINE ls_field      STRING
   DEFINE lc_attribute  LIKE type_t.chr10
   DEFINE ls_value      STRING
   DEFINE i             LIKE type_t.num5
   DEFINE li_chk        LIKE type_t.num5

   #如果是Q類,最後把dzac003再還原去除b_,以供程式內抓取使用
   IF g_prog_type.subString(1,1) = "q" THEN
      LET ls_field = "b_",ls_field
   END IF

   LET li_chk = TRUE
   FOR i = 1 TO g_dzac_m.getLength()
      IF ls_field.trim() = g_dzac_m[i].field THEN
         CASE lc_attribute
            WHEN "dzaa003"   #識別碼 / 欄位名稱
               LET ls_value = g_dzac_m[i].dzaa003
            WHEN "dzaa004"   #設計器版號
               LET ls_value = g_dzac_m[i].dzaa004
            WHEN "dzaa005"   #欄位型態
               LET ls_value = g_dzac_m[i].dzaa005
            WHEN "dzaa006"   #使用標示
               LET ls_value = g_dzac_m[i].dzaa006
            WHEN "dzaa007"   #是否引用
               LET ls_value = g_dzac_m[i].dzaa007
         #---------------------------------------------#
            WHEN "dzac005"   #table編號
               LET ls_value = g_dzac_m[i].dzac005
            WHEN "dzac002"   #欄位編號
               LET ls_value = g_dzac_m[i].dzac002
            WHEN "dzac006"   #欄位屬性
               LET ls_value = g_dzac_m[i].dzeb006
            WHEN "dzeb006"   #欄位屬性
               LET ls_value = g_dzac_m[i].dzeb006
            WHEN "dzac009"   #編輯時開窗
               LET ls_value = g_dzac_m[i].dzac009
            WHEN "dzac010"   #查詢時開窗 
               LET ls_value = g_dzac_m[i].dzac010
            WHEN "dzac011"   #校驗帶值設定
               LET ls_value = g_dzac_m[i].dzac011
            WHEN "dzac014"   #預設值
               LET ls_value = g_dzac_m[i].dzac014
               LET ls_value = ls_value.trim()
               IF ls_value.subString(1,1) = "'" AND
                  ls_value.subString(ls_value.getLength(),ls_value.getLength()) = "'" THEN
                  LET ls_value = ls_value.subString(2,ls_value.getLength()-1)
               END IF
            WHEN "dzac015"   #最大值
               LET ls_value = g_dzac_m[i].dzac015
            WHEN "dzac016"   #最小值
               LET ls_value = g_dzac_m[i].dzac016
            WHEN "dzac020"   #最大值符號
               LET ls_value = g_dzac_m[i].dzac020
            WHEN "dzac021"   #最小值符號
               LET ls_value = g_dzac_m[i].dzac021
            WHEN "dzac017"   #是否可編輯
               LET ls_value = g_dzac_m[i].dzac017
            WHEN "dzac018"   #是否可查詢
               LET ls_value = g_dzac_m[i].dzac018
         #---------------------------------------------#
            WHEN "dzak005"   #是否可查詢
               LET ls_value = g_dzac_m[i].dzak005
            WHEN "dzak007"   #是否可查詢
               LET ls_value = g_dzac_m[i].dzak007
            WHEN "dzak008"   #是否可查詢
               LET ls_value = g_dzac_m[i].dzak008
            WHEN "dzak009"   #是否可查詢
               LET ls_value = g_dzac_m[i].dzak009
            WHEN "dzak010"   #是否可查詢
               LET ls_value = g_dzac_m[i].dzak010
            WHEN "dzak011"   #是否可查詢
               LET ls_value = g_dzac_m[i].dzak011
         #---------------------------------------------#
            WHEN "dzep011"   #系統分類碼
               LET ls_value = g_dzac_m[i].dzep011
            OTHERWISE
               IF g_t100debug >=3 THEN
                  DISPLAY "注意: 調用欄位(",ls_field,")屬性:",lc_attribute,"不存在"
               END IF
         END CASE

         EXIT FOR
         LET li_chk = FALSE
      END IF
   END FOR

   IF NOT li_chk THEN
      IF g_t100debug >=3 THEN
         DISPLAY "注意: 調用欄位(",ls_field,")不存在屬性清單"
      END IF
      LET ls_value = ""
   END IF

   RETURN ls_value.trim()
END FUNCTION


#單頭串查部分
PUBLIC FUNCTION sadzp030_tab_ma_qry_dzal(ls_type)

   DEFINE li_cnt      LIKE type_t.num5
   DEFINE li_items    LIKE type_t.num5
   DEFINE li_status   LIKE type_t.num5
   DEFINE lc_dzal002  LIKE dzal_t.dzal002   #ON ACTION的 action-id (按鈕ID)
   DEFINE lc_dzal007  LIKE dzal_t.dzal007   #串查型態
   DEFINE lc_dzal     DYNAMIC ARRAY OF RECORD
             dzal002  LIKE dzal_t.dzal002,  #ON ACTION的 action-id (按鈕ID)
             dzal005  LIKE dzal_t.dzal005,  #依附控件編號
             dzal006  LIKE dzal_t.dzal006,  #程式參考設定,串查作業名稱
             dzal007  LIKE dzal_t.dzal007   #串查型態
                      END RECORD
   DEFINE ls_type     STRING
   DEFINE li_dzal     LIKE type_t.num5

   CALL lc_dzal.clear()

   FOR li_cnt = 1 TO g_dzac_m.getLength()
      IF g_dzac_m[li_cnt].type = ls_type THEN
         FOR li_dzal = 1 TO gc_dzal.getLength()
            IF g_dzac_m[li_cnt].dzaa003 = gc_dzal[li_dzal].dzal005 THEN
               #檢查如果有一樣的就不加入了
               LET li_status = TRUE
               FOR li_items = 1 TO lc_dzal.getLength()
                  IF li_status AND lc_dzal[li_items].dzal002 = gc_dzal[li_dzal].dzal002 AND
                                   lc_dzal[li_items].dzal005 = gc_dzal[li_dzal].dzal005 AND
                                   lc_dzal[li_items].dzal006 = gc_dzal[li_dzal].dzal006 AND
                                   lc_dzal[li_items].dzal007 = gc_dzal[li_dzal].dzal007 THEN
                     LET li_status = FALSE
                  END IF
               END FOR

               IF li_status THEN
                  LET li_items = lc_dzal.getLength() + 1
                  LET lc_dzal[li_items].dzal002 = gc_dzal[li_dzal].dzal002 CLIPPED
                  LET lc_dzal[li_items].dzal005 = gc_dzal[li_dzal].dzal005 CLIPPED
                  LET lc_dzal[li_items].dzal006 = gc_dzal[li_dzal].dzal006 CLIPPED
                  LET lc_dzal[li_items].dzal007 = gc_dzal[li_dzal].dzal007 CLIPPED
               END IF
            END IF
         END FOR
      END IF
   END FOR

   RETURN lc_dzal

END FUNCTION

