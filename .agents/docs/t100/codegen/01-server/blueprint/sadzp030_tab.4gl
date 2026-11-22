#+ Version..: T100-ERP-1.00.00 Build-00001
#+
#+ Filename.: sadzp030_tab
#+ Buildtype: p01樣板自動產生
#+ Memo.....: 產生tab用
#             160622 調整非第一個頁籤的單身子表錯誤
#+            180411-00062 by yscheng : 將多語言設定寫入tab檔中
#+            181029-00039#1 by 04013 :b_fill_sql中的WHERE條件未產生，缺少c03a樣板的tab檔案
#+            201109-00009#1 by 00824 :模糊查詢功能
#+            230223-00034#1 2023/02/23 by 00824 修正browser頁籤reference欄位(_desc)無法顯示資料問題
#+..........: 230328-00032#1 2023/03/28 by 00824 新增XR報表類型(k類)
#+..........: 230710-00021#1 2023/07/11 by 00824 修正action清單未排序問題，減少程式段比對難度
#+..........: 240930-00022#1 2024/09/30 by 00824 因應項標中心設計器優化需求，配合調整相關工具及樣板
#+..........: 250212-00006#1 2025/02/12 by 20992  配合項標中心對於設計器優化需求，調整日期預設值、最大最小值可使用 :TODAY 設定

IMPORT os
SCHEMA ds

GLOBALS "../../cfg/top_global.inc"
GLOBALS "../4gl/adzp150_global.inc"   #180411-00062

GLOBALS
   DEFINE g_prog_id         LIKE dzfi_t.dzfi001
   DEFINE g_generate_target_path STRING
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
   DEFINE g_mdatalang DYNAMIC ARRAY OF RECORD  #單身
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
            ref_lang   STRING,   #180411-00062
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
            ref_lang   STRING,   #180411-00062
            ref_rtn    STRING,
            used       LIKE type_t.chr1
          END RECORD
   #180411-00062 --- end ---
   DEFINE g_table_main     STRING
   DEFINE g_detail_op      DYNAMIC ARRAY OF RECORD
            record_id      STRING,
            table_id       STRING,
            cols           STRING
            END RECORD
   DEFINE g_keep_pk        DYNAMIC ARRAY OF RECORD
            dzag002   LIKE dzag_t.dzag002, #table編號
            dzed004   STRING
                           END RECORD
   DEFINE g_keep_fk        DYNAMIC ARRAY OF RECORD
            dzag002   LIKE dzag_t.dzag002, #table編號
            dzag004   LIKE dzag_t.dzag004, #上層table編號
            dzed004   STRING,
            dzed006   STRING    #鍵值欄位 / 外來鍵值欄位
                            END RECORD
   DEFINE g_more_pk_then_head BOOLEAN    #偵測到比單頭PK更多的欄位設定
END GLOBALS

DEFINE g_domDoc      om.DomDocument
DEFINE g_domRoot     om.DomNode
DEFINE g_gzzacrtid   LIKE gzza_t.gzzacrtid  #創建人
DEFINE g_gzzacrtdt   LIKE gzza_t.gzzacrtdt  #創建日期

#每個單身SR的id
DEFINE g_detail_id DYNAMIC ARRAY OF RECORD
         detail_sn   LIKE type_t.num5,
         detail_id   LIKE dzfi_t.dzfi004,
         detail_tab  LIKE dzfs_t.dzfs004,
         layer3_up   LIKE dzag_t.dzag004,   #上階table編號
         cluster     STRING,   #串查記錄字串
         detail_ins  LIKE dzfs_t.dzfs006,
         detail_del  LIKE dzfs_t.dzfs007,
         detail_app  LIKE dzfs_t.dzfs008,
         connect     LIKE dzfs_t.dzfs009    #是否有連動
END RECORD

#欄位存放陣列
DEFINE g_col_set DYNAMIC ARRAY OF RECORD
         pos     STRING,
         fields  STRING
END RECORD

DEFINE g_reference_col_set  STRING     #reference欄位存放字串

DEFINE g_4fdNode RECORD
         masterNode  om.DomNode,
         detailNode  om.DomNode,
         browserNode om.DomNode,
         doctailNode om.DomNode
             END RECORD

DEFINE g_table_main_pk  STRING
DEFINE g_table DYNAMIC ARRAY OF RECORD
         table_type  LIKE type_t.num5,   #1:單頭表 2:單身表
         table_id    STRING,
         table_pk    STRING,
         table_fk    STRING
#        table_ent   LIKE type_t.num5
           END RECORD

DEFINE g_dzff   RECORD
         id     STRING,
         pid    STRING,
         type   STRING,
         desc   STRING,
         speed  STRING,
         slid   STRING,
         spid   STRING,
         stype  STRING,
         type2  STRING,
         type3  STRING,
         type4  STRING,
         type5  STRING,
         type6  STRING
            END RECORD

#########主程式部分###########
# i01: O 含查詢計畫的單檔
# i10: O 單檔 
# i02: O 單檔多欄
# i03: X 以樹狀型態表示的單檔多欄 (暫時取消)
# i04: O 樹狀雙檔 (主從表)
# i05: O 樹狀單檔
# i06: X 引導式 (暫時取消)
# i07: O 含查詢計畫的假雙檔
# i08: O 樹狀單檔+單身
# i09: O 雙檔 (單身凍結)
# i12: O 不含查詢計畫的假雙檔
# i13: O 六階樹狀
# t01: O 含查詢計畫的雙檔
# t02: O 單檔多欄 分單頭單身
# q01: O 列表查詢與瀏覽
# q02: O QBE查詢與列表瀏覽
# q03: O 列表查詢與樹狀瀏覽
# q04: O 列表查詢與單頭單身瀏覽
#########子程式部分###########
# c01a: O 單檔全功能
# c01b: O 單檔只做單一INPUT
# c01c: O 單檔只做單一CONSTRUCT
# c02b: O 單檔多欄只做單一INPUT
# c02b: O 單檔多欄只做單一CONSTRUCT
# c03a: O 雙檔全功能
# c03b: O 雙檔只做單一INPUT
# c03c: O 雙檔只做單一CONSTRUCT
# c04a: O 雙檔多欄全功能

PUBLIC FUNCTION sadzp030_tab_gen(p_prog_id,p_prog_ver,p_prog_type,p_dgenv)

   DEFINE p_prog_id      STRING    #程式代碼
   DEFINE p_prog_ver     STRING    #產生版本
   DEFINE p_prog_type    STRING    #樣板型態
   DEFINE p_dgenv        LIKE dzaa_t.dzaa009  #區域 (識別標示)
   DEFINE ls_sys         STRING
   DEFINE ls_4fd_file    STRING    #4FD畫面檔
   DEFINE ls_tab_file    STRING
   DEFINE li_cnt         LIKE type_t.num5
   DEFINE li_scuss       LIKE type_t.num5       #操作上是否成功(全過程)
   DEFINE lc_spec_type   LIKE type_t.chr1       #型態
   DEFINE li_dzax007     LIKE dzax_t.dzax007    #需要的固定參數個數
   DEFINE ls_cmd         STRING    #230328-00032


   LET g_prog_id = p_prog_id       #程式代碼
   LET g_sd_ver = p_prog_ver       #規格版本/產生版本

#  LET g_erp_ver = FGL_GETENV("ERPVER")  #產品版本
#  LET g_cust = FGL_GETENV("CUST")       #客戶編號
   LET g_dgenv = p_dgenv                 #客製標示

   IF g_sd_ver IS NULL THEN
      DISPLAY cl_getmsg("adz-01100",g_lang)
     #DISPLAY "注意: 未接到產生版次,或傳入為NULL! 預設為版次 1"
      LET g_sd_ver = 1
#    #14/06/26 hiko的sub會有問題所以同意給預設
   END IF

   LET g_t100debug = FGL_GETENV("T100DEBUG")

   #如果進來的規格名稱是子畫面的,特別處理,Return TRUE回去
   LET lc_spec_type = cl_chk_spec_type(g_prog_id)
   IF lc_spec_type = "F" THEN
      RETURN TRUE
   END IF

   #執行前清空dynamic array,預防有工具多次連續呼叫
   CALL g_datalang.clear()
   CALL g_mdatalang.clear()
   CALL g_multilang_m.clear()   #180411-00062
   CALL g_multilang_d.clear()   #180411-00062
   CALL g_detail_op.clear()
   CALL g_detail_id.clear()
   CALL g_col_set.clear()
   CALL g_table.clear()
   CALL g_keep_pk.clear()
   CALL g_keep_fk.clear()
   CALL sadzp030_tab_ma_init()
   LET g_cross_4gl_stus = TRUE      #跨4gl判斷正確與否
   LET g_more_pk_then_head = FALSE    #偵測到比單頭PK更多的欄位設定

   #開始判斷關係
   CALL sadzp030_tab_relation_prog_type(p_prog_type)
   RETURNING g_prog_type,g_gzzacrtid,g_gzzacrtdt,li_dzax007

   #230328-00032 ---start---
   #依據重遠提供的程式段落做調整
   #如果是k類，改執行adzp031，
   #用cmdrun的方式去執行，以避免更新舊客時還需要額外更新其它程式及表格，
   #但因為adzp031是用cmdrun的方式去執行，所以無法接收回傳值，且adzp031程式中會直接顯示是否成功，
   #與重遠討論後，此處就直接回傳TRUE就好了
   IF lc_spec_type = "K" THEN
   #  CALL sadzp199_tab_gen(g_prog_id,g_sd_ver,p_dgenv) RETURNING li_scuss
   #  IF NOT li_scuss THEN
   #     DISPLAY cl_getmsg("adz-01101",g_lang)
   #    #DISPLAY "注意: 報表元件規格產生器Return FALSE"
   #  END IF
   #  RETURN li_scuss

      LET ls_cmd = "$FGLRUN $ADZ/42r/adzp031.42r"," ",g_prog_id," ",g_sd_ver," ",p_dgenv
      DISPLAY "K類報表tab檔產生 ls_cmd:",ls_cmd
      RUN ls_cmd
      RETURN TRUE
   END IF
   #230328-00032 --- end ---

   #如果是x01樣板,改CALL 報表元件tab產生function
   IF g_prog_type = "x00" THEN
      CALL sadzp188_tab_gen(g_prog_id,g_sd_ver,p_dgenv) RETURNING li_scuss
      IF NOT li_scuss THEN
         DISPLAY cl_getmsg("adz-01101",g_lang)
        #DISPLAY "注意: 報表元件規格產生器Return FALSE"
      END IF
      RETURN li_scuss
   END IF

   #將回傳的資料區分檢視
   IF g_prog_type IS NULL THEN
      DISPLAY cl_getmsg("adz-01102",g_lang)
     #DISPLAY "ERROR: 使用的樣板無法被偵測出來(NULL),請檢查程式命名與基礎設定"
      LET li_scuss = FALSE
      RETURN li_scuss
   ELSE
      DISPLAY "INFO: 使用",g_prog_type,"樣板 (第",g_sd_ver USING "<<<","版次)"
   END IF

   SELECT gzzx006 INTO li_cnt FROM gzzx_t WHERE gzzx001 = g_prog_id
   IF SQLCA.SQLCODE THEN 
      LET li_cnt = 0
   END IF
   IF li_cnt IS NULL THEN LET li_cnt = 1 END IF
   DISPLAY "INFO: 程式",g_prog_id CLIPPED,"第",li_cnt +1 USING "<<<<<<","次產出藍圖檔(tab)"

   LET li_scuss = TRUE
   
   #從規格編號判斷path路徑
   LET ls_sys = sadzp030_tab_file_module(g_prog_id,"A",g_dgenv)

   #確認有沒有傳入的預設tab路徑
   LET g_generate_target_path = g_generate_target_path.trim()
   LET ls_tab_file = ""
   IF g_generate_target_path IS NOT NULL THEN
      DISPLAY "INFO: 偵測到有傳入預設路徑值:",g_generate_target_path
      IF os.Path.isDirectory(g_generate_target_path) THEN
         LET ls_tab_file = g_prog_id,".tab"
         LET ls_tab_file = os.Path.join(g_generate_target_path,ls_tab_file)
         DISPLAY "INFO: 偵測到傳入路徑,自動補入檔名:",ls_tab_file
      ELSE
         IF os.Path.isDirectory(os.Path.dirName(g_generate_target_path)) THEN
            IF DOWNSHIFT(os.Path.extension(g_generate_target_path)) = "tab" THEN
               LET ls_tab_file = g_generate_target_path
            ELSE
               LET ls_tab_file = os.Path.rootname(g_generate_target_path),".tab"
               DISPLAY "INFO: 偵測到傳入附檔名錯誤,自動更正:",ls_tab_file
            END IF
         END IF
      END IF
   END IF

   #產生TAB的來源檔案
   #tab  $MODULE/dzx/tab/xxxxx.tab
   IF ls_tab_file IS NULL THEN
      LET ls_tab_file = g_prog_id,".tab"
      LET ls_tab_file = os.Path.join(os.Path.join(os.Path.join(ls_sys,"dzx"),"tab"),ls_tab_file)
      IF os.Path.exists(os.Path.dirname(ls_tab_file)) THEN
      ELSE
         DISPLAY cl_getmsg_parm("adz-01103",g_lang,os.Path.dirname(ls_tab_file))
        #DISPLAY "ERROR: 產出tab的路徑(",os.Path.dirname(ls_tab_file),")不存在,請檢查設定後重新執行!"
         LET li_scuss = FALSE
         RETURN li_scuss
      END IF
   END IF
   DISPLAY cl_getmsg_parm("adz-01104",g_lang,ls_tab_file)
  #DISPLAY "INFO: 產出tab路徑:",ls_tab_file

   #p00/i00/m00/q00 不讀取4fd(因為沒有檔案可以讀)
   IF g_prog_type = "p00" OR g_prog_type = "i00" OR g_prog_type = "m00" OR g_prog_type = "q00" OR
      g_prog_type = "p02" THEN
   ELSE
      #只確認,實體不讀取  4fd  $MODULE/4fd/xxxxx.4fd
      LET ls_4fd_file = g_prog_id,".4fd"
      LET ls_4fd_file = os.Path.join(os.Path.join(ls_sys,"4fd"),ls_4fd_file)
      IF os.Path.exists(ls_4fd_file) THEN
         IF g_t100debug >= 6 THEN
            DISPLAY "INFO: 讀取4fd路徑:",ls_4fd_file
         END IF
      ELSE
         DISPLAY cl_getmsg("adz-01105",g_lang),ls_4fd_file
        #DISPLAY "ERROR: 以下4fd不存在:",ls_4fd_file
         LET li_scuss = FALSE
         RETURN li_scuss
      END IF

      IF g_prog_type = "p01" OR g_prog_type = "r01" THEN
      ELSE
         #讀取基本資料 主表(dzag)及個別欄位關係 
         CALL sadzp030_tab_relation() RETURNING g_table_main,g_table_main_pk,g_table

         IF g_t100debug >= 6 THEN
            DISPLAY "INFO: Main Table:",g_table_main," PK:",g_table_main_pk
         END IF

         #若主表/主表PK任意一個不存在,則中止執行, RETURN FALSE
         IF g_table_main IS NULL OR g_table_main_pk IS NULL THEN 
            DISPLAY cl_getmsg("adz-01106",g_lang)
           #DISPLAY "ERROR: 主表或PK不存在,請至設計器調整後執行"
            LET li_scuss = FALSE
            RETURN li_scuss
         END IF

         #顯示其他子表PK
         IF g_t100debug >= 3 THEN
            FOR li_cnt = 1 TO g_table.getLength()
               IF g_table[li_cnt].table_type = 1 THEN
                  IF g_t100debug >= 6 THEN
                     DISPLAY "INTO: Master Table:",g_table[li_cnt].table_id," PK:",g_table[li_cnt].table_pk," FK:",g_table[li_cnt].table_fk
                  END IF
               END IF
            END FOR
            FOR li_cnt = 1 TO g_table.getLength()
               IF g_table[li_cnt].table_type = 2 THEN
                  IF g_t100debug >= 6 THEN
                     DISPLAY "INFO: Detail Table:",g_table[li_cnt].table_id," PK:",g_table[li_cnt].table_pk," FK:",g_table[li_cnt].table_fk
                  END IF
               END IF
            END FOR
         END IF
      END IF

   END IF

   #產出 tab
   IF li_scuss THEN
      CALL sadzp030_tab_gen_assembly(ls_tab_file,li_dzax007) RETURNING li_scuss
      IF NOT li_scuss THEN
         DISPLAY cl_getmsg("adz-01107",g_lang)
        #DISPLAY "ERROR: 產生 TAB 過程中發生異常!     請回捲查看第一個ERROR出現時的錯誤!!"
      END IF
   END IF

   IF os.Path.exists(ls_tab_file) THEN
      DISPLAY cl_getmsg("adz-01108",g_lang)
     #DISPLAY "INFO: sadzp030_tab任務結束."
      CALL sadzp030_tab_log(p_prog_type)
   ELSE
      DISPLAY cl_getmsg_parm("adz-01109",g_lang,ls_tab_file)
     #DISPLAY "ERROR: tab檔產出失敗,檔案",ls_tab_file,"不存在!"
      LET li_scuss = FALSE
   END IF

   IF g_more_pk_then_head THEN      #偵測到比單頭PK更多的欄位設定
      LET li_scuss = FALSE
   END IF

   #查看若有跨4gl錯誤的,最終也回傳錯誤
   IF NOT g_cross_4gl_stus THEN     #跨4gl判斷正確與否
      LET li_scuss = FALSE
   END IF

   DISPLAY "[CheckPoint1]",os.path.pwd()
   RETURN li_scuss
END FUNCTION


#更新程式開發紀錄檔
PRIVATE FUNCTION sadzp030_tab_log(p_prog_type)

   DEFINE lc_gzzx003     LIKE gzzx_t.gzzx003  #UI樣板
   DEFINE lc_gzzx004     LIKE gzzx_t.gzzx004  #程式樣板
   DEFINE lc_gzzx005     DATETIME YEAR TO SECOND   #異動時間
   DEFINE li_gzzx006     LIKE gzzx_t.gzzx006  #異動次數
   DEFINE p_prog_type    STRING
   DEFINE li_cnt         LIKE type_t.num5

   LET lc_gzzx003 = p_prog_type.trim() #UI樣板
   LET lc_gzzx004 = g_prog_type.trim() #pattern樣板
   LET lc_gzzx005 = CURRENT            #異動時間

   SELECT COUNT(1) INTO li_cnt FROM gzzx_t 
    WHERE gzzx001 = g_prog_id
   IF li_cnt = 0 THEN
      LET li_gzzx006 = 1
      INSERT INTO gzzx_t(gzzx001,gzzx002,gzzx003,gzzx004,gzzx005,gzzx006)
       VALUES(g_prog_id, g_sd_ver,lc_gzzx003,lc_gzzx004,lc_gzzx005,li_gzzx006)
      IF SQLCA.SQLCODE THEN
         IF g_t100debug >= 3 THEN
            DISPLAY cl_getmsg("adz-01110",g_lang)
           #DISPLAY "注意: 更新記錄檔(INS)無法執行,自動跳過此程序"
         END IF
      END IF
   ELSE
      SELECT gzzx006 INTO li_gzzx006 FROM gzzx_t
       WHERE gzzx001 = g_prog_id
      IF li_gzzx006 IS NULL THEN LET li_gzzx006 = 1 END IF
      LET li_gzzx006 = li_gzzx006 + 1
      UPDATE gzzx_t
         SET gzzx002 = g_sd_ver,   gzzx003 = lc_gzzx003,
             gzzx004 = lc_gzzx004, gzzx005 = lc_gzzx005,
             gzzx006 = li_gzzx006
       WHERE gzzx001 = g_prog_id
      IF SQLCA.SQLCODE THEN
         IF g_t100debug >= 3 THEN
            DISPLAY cl_getmsg("adz-01111",g_lang)
           #DISPLAY "注意: 更新記錄檔(UPD)無法執行,自動跳過此程序"
         END IF
      END IF
   END IF

END FUNCTION

#產生tab主結構
PRIVATE FUNCTION sadzp030_tab_gen_assembly(ls_tab_file,li_dzax007)

   DEFINE ls_tab_file    STRING
   DEFINE l_structure    om.DomNode
   DEFINE l_form         om.DomNode
   DEFINE l_dataset      om.DomNode
   DEFINE li_detail_cnt  LIKE type_t.num5
   DEFINE li_scuss       LIKE type_t.num5      #操作上是否成功(全過程)
#  DEFINE lc_dzfq014     LIKE dzfq_t.dzfq014   #空框架處理
   DEFINE lc_gzzal003    LIKE gzzal_t.gzzal003 #程式繁體中文名稱
   DEFINE ls_sys         STRING
   DEFINE li_dzax007     LIKE dzax_t.dzax007   #固定參數個數
   DEFINE lc_sd_ver      LIKE dzaa_t.dzaa002   #規格版次/程式版次

   LET li_scuss = TRUE
   
   #範例:<assembly name="程式名稱aooi002" module="模組aoo"      jobmode="背景程式N"
   #               type="應用樣板i05"     industry="行業別std"  page="單身頁數1"
   #               fix_arg="固化參數數量"
   #               crtid="首版開發人id"   crtdt="首版開發日期"
   #               modid="首版開發人id"   moddt="首版開發日期"
   #               sdver="產出程式版次"
   #               description="程式的繁體中文名稱" />
   LET g_domDoc = om.DomDocument.create("assembly")
   LET g_domRoot = g_domDoc.getDocumentElement()
   
   #name:程式名稱
   CALL g_domRoot.setAttribute("name",g_prog_id)

   #module:模組
   LET ls_sys = DOWNSHIFT(sadzp030_tab_file_module(g_prog_id,"B",g_dgenv))
   CALL g_domRoot.setAttribute("module",ls_sys)

   #jobmode:執行模式；N-視窗執行 B-背景執行 B-皆可(由參數指定)
   CALL g_domRoot.setAttribute("jobmode","N")

   #type: 樣板
   CALL g_domRoot.setAttribute("type",g_prog_type) 

   #industry:行業別
   CALL g_domRoot.setAttribute("industry","std")

   #fix_arg:固化參數數量
   IF li_dzax007 IS NULL THEN LET li_dzax007 = 0 END IF
   CALL g_domRoot.setAttribute("fix_arg",li_dzax007)

   #crt:創建作業的user-id/日期
   CALL g_domRoot.setAttribute("crtid",g_gzzacrtid) 
   CALL g_domRoot.setAttribute("crtdt",g_gzzacrtdt) 

   #mod:修改作業的user-id/日期
   CALL sadzp030_tab_modify_info("standard") RETURNING g_gzzacrtid,g_gzzacrtdt
   IF g_gzzacrtid IS NULL THEN LET g_gzzacrtid = "00000" END IF
   IF g_gzzacrtdt IS NULL THEN LET g_gzzacrtdt = "1900-01-01 00:00:00" END IF
   CALL g_domRoot.setAttribute("modid",g_gzzacrtid) 
   CALL g_domRoot.setAttribute("moddt",g_gzzacrtdt) 
   IF g_dgenv = "c" THEN
      #查看最後一個標準版次
      SELECT MAX(dzaa002) INTO lc_sd_ver FROM dzaa_t
       WHERE dzaa001 = g_prog_id AND dzaa009 = "s"
   ELSE 
      LET lc_sd_ver = g_sd_ver
   END IF
   CALL g_domRoot.setAttribute("sdver",lc_sd_ver USING "&&&&" ) 

   CALL sadzp030_tab_modify_info("customer") RETURNING g_gzzacrtid,g_gzzacrtdt
   IF g_gzzacrtid IS NULL THEN LET g_gzzacrtid = "00000" END IF
   IF g_gzzacrtdt IS NULL THEN LET g_gzzacrtdt = "1900-01-01 00:00:00" END IF
   CALL g_domRoot.setAttribute("cusdt",g_gzzacrtdt) 
   IF g_dgenv = "c" THEN
      LET lc_sd_ver = g_sd_ver
   ELSE 
      LET lc_sd_ver = "0"
   END IF
   CALL g_domRoot.setAttribute("cusver",lc_sd_ver USING "&&&&" ) 

   #page:單身多頁簽頁數  i02/i03/i04/i07/i08/i09/i12/t01/t02/q01/q02/q03/q04
   IF g_prog_type = "i02" OR g_prog_type = "i03" OR g_prog_type = "i04" OR
      g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR
      g_prog_type = "i12" OR g_prog_type = "t01" OR g_prog_type = "t02" OR
      g_prog_type = "q01" OR g_prog_type = "q02" OR g_prog_type = "q03" OR g_prog_type = "q04" OR
      g_prog_type = "c02a" OR g_prog_type = "c02b" OR g_prog_type = "c02c" OR
      g_prog_type = "c03a" OR g_prog_type = "c03b" OR g_prog_type = "c03c" OR
      g_prog_type = "c04a" THEN
      CALL sadzp030_tab_de_cnt() RETURNING li_scuss,li_detail_cnt,g_detail_id
      IF NOT li_scuss THEN
         DISPLAY cl_getmsg("adz-01112",g_lang)
        #DISPLAY "ERROR: 單身相關頁面產生過程發生錯誤!請回捲查看第一個ERROR出現時的錯誤!!"
      END IF
      CALL g_domRoot.setAttribute("page",li_detail_cnt)
      IF g_t100debug >= 3 THEN
         DISPLAY cl_getmsg("adz-01113",g_lang),li_detail_cnt
        #DISPLAY "INFO: 抓取總頁面數:",li_detail_cnt
      END IF
   END IF

   #tabversion:版本
   CALL g_domRoot.setAttribute("tabver","10001")

   #description:程式的繁體中文名稱
   SELECT gzzal003 INTO lc_gzzal003 FROM gzzal_t 
    WHERE gzzal001 = g_prog_id AND gzzal002 = "zh_TW"
   IF SQLCA.SQLCODE THEN
      #沒有畫面的就取子程式或WEB
      IF g_prog_id[1,3] = "wss" OR g_prog_id[1,4] = "cwss" THEN
         SELECT gzjal003 INTO lc_gzzal003 FROM gzjal_t
          WHERE gzjal001 = g_prog_id AND gzjal002 = "zh_TW"
      ELSE
         SELECT gzdel003 INTO lc_gzzal003 FROM gzdel_t
          WHERE gzdel001 = g_prog_id AND gzdel002 = "zh_TW"
      END IF
   END IF
   CALL g_domRoot.setAttribute("description",lc_gzzal003 CLIPPED)

   #排除查看4fd檔的類型
   IF g_prog_type = "p00" OR g_prog_type = "i00" OR g_prog_type = "m00" OR g_prog_type = "q00" OR
      g_prog_type = "p02" THEN
   ELSE
      #以下定義各大 SECTION 區塊

      #定義 <STRUCTURE>
      LET l_structure = g_domRoot.createChild("structure")
      CALL sadzp030_tab_gen_struct(l_structure)

      #定義 <FORM>
      LET l_form = g_domRoot.createChild("form")
      CALL sadzp030_tab_gen_form(l_form)

      #定義 <DATASET>  #下方為排除dataset清單
      IF g_prog_type = "p01" OR g_prog_type = "r01" OR g_prog_type = "q03" OR
        ( g_prog_type = "q04" AND (FGL_GETENV("ZONE") = "t10dev" AND g_prog_id <> "aloq005") ) THEN
      ELSE
         LET l_dataset = g_domRoot.createChild("dataset")
         CALL sadzp030_tab_gen_dataset(l_dataset)
      END IF
   END IF

   #因為如果是i07的樣板,有可能是從t01降級的,遇到i07都重設一次
   IF g_prog_type = "i07" THEN
      CALL g_domRoot.setAttribute("type",g_prog_type) 
   END IF

   #子程式的空框架處理 dzfq014 #目前沒有空框架這種事情!! 2014/11/04 全部走過一次再轉freestyle或解開section
   #IF g_prog_type <> "i00" THEN   #先排除i00主程式Free Style
   #   SELECT dzfq014 INTO lc_dzfq014 FROM dzfq_t 
   #    WHERE dzfq003 = "1" #g_sd_ver  #識別碼版號,因為dzfq只有r.a會寫入,所以咬死1
   #      AND dzfq004 = g_prog_id #畫面代號
   #   IF NOT STATUS THEN 
   #      IF lc_dzfq014 = "Y" THEN
   #         DISPLAY "注意: 因為dzfq014被設定為Y，故更改為p00樣板"
   #         CALL g_domRoot.setAttribute("type","p00") 
   #         UPDATE gzzx_t SET gzzx004 = "p00"
   #          WHERE gzzx001 = g_prog_id
   #      END IF
   #   END IF
   #END IF

   #檔案寫入
   CALL sadzp030_writefile(g_domRoot.toString(),ls_tab_file)

   RETURN li_scuss

END FUNCTION


#存檔 改用base.Channel的方式存檔
PRIVATE FUNCTION sadzp030_writefile(p_xmlStr,p_file)

   DEFINE p_file         STRING
   DEFINE p_xmlStr       STRING
   DEFINE l_ch_out       base.Channel
   DEFINE l_file_existed BOOLEAN
   DEFINE l_file_deleted BOOLEAN

   LET l_file_existed = FALSE
   LET l_file_deleted = FALSE

   IF os.Path.exists(p_file) THEN
      IF NOT os.Path.delete(p_file) THEN
         DISPLAY cl_getmsg_parm("adz-01114",g_lang,p_file)
        #DISPLAY "注意: 刪除檔案",p_file,"時發生問題,可能會無法寫入"
      END IF
   END IF
   
   LET l_ch_out = base.Channel.create()
   CALL l_ch_out.setDelimiter("")
   CALL l_ch_out.openFile(p_file,"w")
   CALL l_ch_out.write(p_xmlStr)
   CALL l_ch_out.close()

   #對產生器寫出的檔案權限在UNIX下均全部打開
   IF os.Path.separator() = "/" THEN 
      IF os.Path.chrwx(p_file,511) THEN
      END IF
   END IF

END FUNCTION


#+ 找出最後標準版修改人的資料(user-id與日期)
PRIVATE FUNCTION sadzp030_tab_modify_info(ls_type)

   DEFINE lc_crtid      LIKE dzaa_t.dzaacrtid
   DEFINE lc_crtdt      LIKE dzaa_t.dzaacrtdt
   DEFINE lc_modid      LIKE dzaa_t.dzaamodid
   DEFINE lc_moddt      LIKE dzaa_t.dzaamoddt
   DEFINE lc_bakid      LIKE dzaa_t.dzaamodid
   DEFINE lc_bakdt      LIKE dzaa_t.dzaamoddt
   DEFINE lc_dzaa004    LIKE dzaa_t.dzaa004    #設計點版次
   DEFINE lc_dzba004    LIKE dzba_t.dzba004    #設計點版次
   DEFINE ls_sql        STRING
   DEFINE ls_type       STRING
   DEFINE lc_dgenv      LIKE type_t.chr1

   IF ls_type = "customer" THEN
      LET lc_dgenv = "c"
   ELSE
      LET lc_dgenv = "s"
   END IF

   #先把dzaa_t內的最高版次找出來 (但dzaa004不可以大於dzaa002)
   SELECT MAX(DISTINCT(dzaa004)) INTO lc_dzaa004 FROM dzaa_t
    WHERE dzaa001 = g_prog_id 
##    AND dzaa002 = g_sd_ver
#     AND dzaa002 >= dzaa004
#     AND dzaa008 = g_erp_ver   #產品版本
      AND dzaa009 = lc_dgenv    #客製標示
#     AND dzaa010 = g_cust      #客戶編號
      AND dzaastus = "Y"
   IF SQLCA.SQLCODE OR lc_dzaa004 IS NULL THEN
      LET lc_dzaa004 = "1"   #找不到就只好當作 1
   END IF

   LET ls_sql = "SELECT dzaacrtid,dzaacrtdt,dzaamodid,dzaamoddt FROM dzaa_t ",
                " WHERE dzaa001 = '",g_prog_id CLIPPED,"' ",
##                " AND dzaa002 = '",g_sd_ver CLIPPED,"' ",
                  " AND dzaa004 = '",lc_dzaa004 CLIPPED,"' ",
#                 " AND dzaa008 = '",g_erp_ver CLIPPED,"' ",   #產品版本
                  " AND dzaa009 = '",lc_dgenv CLIPPED,"' ",    #客製標示
#                 " AND dzaa010 = '",g_cust CLIPPED,"' ",      #客戶編號
                  " AND dzaastus = 'Y' "

   DECLARE sadzp030_tab_modify_info_cs CURSOR FROM ls_sql
   FOREACH sadzp030_tab_modify_info_cs INTO lc_crtid,lc_crtdt,lc_modid,lc_moddt
      #撈到mod一整組都有值就算是新的
      IF lc_modid IS NOT NULL AND lc_moddt IS NOT NULL THEN
         EXIT FOREACH
      END IF
      #沒撈到就把crt這一組記起來
      IF lc_crtid IS NOT NULL AND lc_crtdt IS NOT NULL THEN
         LET lc_bakid = lc_crtid CLIPPED
         LET lc_bakdt = lc_crtdt CLIPPED
      END IF
   END FOREACH

   #滾完了如果mod這一組還是null,就把crt拿出來用(有可能這個版次只有新增一個other func)
   IF lc_modid IS NULL OR lc_moddt IS NULL THEN
      LET lc_modid = lc_bakid 
      LET lc_moddt = lc_bakdt 
   END IF

   FREE sadzp030_tab_modify_info_cs 
   RETURN lc_modid,lc_moddt
END FUNCTION

#產生 tab 檔內的 structure結構
PRIVATE FUNCTION sadzp030_tab_gen_struct(lnode_structure)

   DEFINE lnode_structure   om.DomNode   #TAB STRUCTURE結構
   DEFINE lnode_section     om.DomNode   #TAB SECTION結構
   DEFINE lnode_var         om.DomNode   #TAB VAR結構
   DEFINE lnode_formfield   om.DomNode   #TAB SECTION結構-formfiled Q專用
   DEFINE lnode_fvar        om.DomNode   #TAB VAR結構-formfield Q專用
   DEFINE lnode_action      om.DomNode   #TAB ACTION結構
   DEFINE lnode_action_branch      om.DomNode   #TAB ACTION結構
   DEFINE lnode_cluster     om.DomNode   #TAB 單身串查結構
   DEFINE lnode_component   om.DomNode   #TAB COMPONENT結構

   DEFINE ls_temp           STRING
   DEFINE ls_temp_detailid  STRING
   DEFINE li_cnt            LIKE type_t.num5
   DEFINE li_dzal           LIKE type_t.num5
   DEFINE li_chk            LIKE type_t.num5
   DEFINE ls_browse_col     STRING
   DEFINE ls_master_col     STRING
   DEFINE ls_detail_col     STRING
   DEFINE ls_site_col       STRING
   DEFINE ls_reftab         STRING
   DEFINE ls_refwc          STRING
   DEFINE la_reference   DYNAMIC ARRAY OF RECORD
             field       STRING,
             ref_sql     STRING,
             ref_field   STRING
                         END RECORD
   DEFINE l_bsreference     om.DomNode
   DEFINE l_token           base.StringTokenizer
   DEFINE ls_detailid       STRING
   DEFINE ls_token          STRING
   DEFINE la_pageid      DYNAMIC ARRAY OF RECORD
             record         STRING,
             page           LIKE type_t.num5
                         END RECORD
   DEFINE lc_dzal     DYNAMIC ARRAY OF RECORD
             dzal002  LIKE dzal_t.dzal002,  #ON ACTION的 action-id (按鈕ID)
             dzal005  LIKE dzal_t.dzal005,  #依附控件編號
             dzal006  LIKE dzal_t.dzal006,  #串查作業名稱,程式參考設定
             dzal007  LIKE dzal_t.dzal007   #串查型態
                      END RECORD
   DEFINE li_quota LIKE type_t.num5
   DEFINE ls_contact_col  STRING
   DEFINE ls_query_temp   STRING

   #建置 structure -> section id="global_var"
   LET lnode_section = lnode_structure.createChild("section")
   CALL lnode_section.setAttribute("id","global_var")

   #Q類要多一個form field存放對照表
   IF g_prog_type.subString(1,1)= "q" THEN
      LET lnode_formfield = lnode_structure.createChild("section")
      CALL lnode_formfield.setAttribute("id","form_field")
   END IF

      #建置 var id="head"  i01/i04/i05/i07/i08/i09/i10/i12/i13/t01
      IF g_prog_type = "i01" OR g_prog_type = "i04" OR g_prog_type = "i05" OR
         g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR 
         g_prog_type = "i10" OR g_prog_type = "i12" OR g_prog_type = "i13" OR
         g_prog_type = "t01" OR g_prog_type = "p01" OR g_prog_type = "r01" OR
         g_prog_type = "q01" OR g_prog_type = "q03" OR g_prog_type = "q04" OR
         g_prog_type = "c01a" OR g_prog_type = "c01b" OR g_prog_type = "c01c" OR 
         g_prog_type = "c03a" OR g_prog_type = "c03b" OR g_prog_type = "c03c" THEN

         LET lnode_var = lnode_section.createChild("var")

         #取得單頭上的欄位資料(對q類來說是qbe處稱為head)
         CASE
            WHEN g_prog_type = "p01" OR g_prog_type = "r01"
               LET ls_master_col = sadzp030_tab_ma("vb_qbe",1,g_table_main,FALSE)
               CALL lnode_var.setAttribute("id","head")
            WHEN g_prog_type.subString(1,1) = "q" 
               LET ls_master_col = sadzp030_tab_ma("qbe",1,g_table_main,TRUE)
               CALL lnode_var.setAttribute("id","qbe")
            OTHERWISE
               LET ls_master_col = sadzp030_tab_ma("mainlayout",1,g_table_main,FALSE),",",
                                   sadzp030_tab_ma("patch_grid",1,g_table_main,FALSE)
               IF ls_master_col.subString(ls_master_col.getLength(),ls_master_col.getLength())="," THEN
                  LET ls_master_col = ls_master_col.subString(1,ls_master_col.getLength()-1)
               END IF
               CALL lnode_var.setAttribute("id","head")
         END CASE

         #填入欄位收集陣列,單頭固定於1
         LET g_col_set[1].pos = "master"
         LET g_col_set[1].fields = ls_master_col
         LET g_col_set[1].fields = s_adzp030_tab_remove_quoter(g_col_set[1].fields CLIPPED)

         #檢查畫面上是否出現site欄位,若出現時site=一般PK處理(g_table_main_pk不異動)
         IF g_table_main_pk.getIndexOf("site",3) THEN
            IF NOT s_adzp030_tab_chk_site_4fd(ls_master_col) THEN
               LET g_table_main_pk = sadzp030_tab_relation_site(g_table_main_pk)
            END IF
         END IF

         #將欄位清單送入取值陣列
         CALL sadzp030_tab_ma_1(ls_master_col,"M",0) RETURNING ls_temp

         #特別欄位形態處理
         LET ls_master_col = sadzp030_tab_ma_change_type(ls_master_col,"M")
         IF g_t100debug >= 3 THEN
            DISPLAY "INFO: 單頭上的欄位:",ls_master_col
         END IF

         CALL lnode_var.setAttribute("value",ls_master_col)

         #如果之前判別是t01,其實有可能想要做出來的是i07+單身,確認一下pk是否全到單頭
         IF g_prog_type = "t01" THEN
            IF NOT sadzp030_tab_check_real_t01(ls_master_col,g_table_main_pk) THEN
               IF g_t100debug >= 3 THEN
                  DISPLAY cl_getmsg("adz-01115",g_lang)
                 #DISPLAY "注意: 單頭欄位不包含全部PK,降級為i07樣板"
               END IF
               LET g_prog_type = "i07"
            END IF
         END IF

         #傳出PK欄位, i07整形
         #非 i07 樣板,直接取用
         #i07/i12 樣板,只能輸出master PK有出現在單頭上的部分
         IF g_prog_type = "i07" OR g_prog_type = "i12" THEN
                                                              #pk放前面,畫面欄位放後面
            LET g_table_main_pk = sadzp030_tab_analyze_pk_i07(g_table_main_pk,ls_master_col)
         END IF
      END IF

            #額外建置 var id="head" 給q04
            IF g_prog_type = "q04" THEN
               LET lnode_var = lnode_section.createChild("var")
               CALL lnode_var.setAttribute("id","head")
               LET ls_master_col = sadzp030_tab_ma("vb_master",1,g_table_main,FALSE)
               CALL lnode_var.setAttribute("value",ls_master_col)

               LET lnode_fvar = lnode_formfield.createChild("var") 
               CALL lnode_fvar.setAttribute("id","head")
               LET ls_master_col = sadzp030_tab_ma("vb_master",6,g_table_main,FALSE)
               CALL lnode_fvar.setAttribute("value",ls_master_col)
            END IF

      #建置 var id="body"  i02/i03/i04/i08/i09/i12/t01/t02/q01/q02/q03/q04
      IF g_prog_type = "i02" OR g_prog_type = "i03" OR g_prog_type = "i04" OR 
         g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR
         g_prog_type = "i12" OR g_prog_type = "t01" OR g_prog_type = "t02" OR
         g_prog_type = "q01" OR g_prog_type = "q02" OR g_prog_type = "q03" OR g_prog_type = "q04" OR
         g_prog_type = "c02a" OR g_prog_type = "c02b" OR g_prog_type = "c02c" OR
         g_prog_type = "c03a" OR g_prog_type = "c03b" OR g_prog_type = "c03c" OR
         g_prog_type = "c04a" THEN

         #如果是i02/t02樣板,先找出主表site欄位名稱
         IF g_prog_type = "i02" OR g_prog_type = "t02" OR
            g_prog_type = "q01" OR g_prog_type = "q02" OR g_prog_type = "q03" OR g_prog_type = "q04" OR 
            g_prog_type = "c02a" OR g_prog_type = "c02b" OR g_prog_type = "c04a" THEN 
            LET ls_site_col = g_table_main.subString(1,g_table_main.getIndexOf("_t",1)-1),"site"
            LET li_chk = FALSE
         END IF

         #為了避免畫面上臨時調整s_detail位置造成全域變數名稱亂跳,先做page定位
#        LET la_pageid = sadzp030_tab_gen_pageid()

         #因應單身多頁籤, 此筆資料不唯一
         FOR li_cnt = 1 TO g_detail_id.getLength()

            LET lnode_var = lnode_section.createChild("var") 
            CALL lnode_var.setAttribute("id","body")

            #取得該單身頁碼 (page)
            CALL lnode_var.setAttribute("page",li_cnt)

            #取得該單身上的 SCREEN RECORD 名稱-
            IF g_prog_type = "i03" THEN
               #只有i03的公版是特別改s_browse
               LET ls_temp_detailid = "s_browse"
            ELSE
               #由陣列中拉出detail_id
               LET ls_temp_detailid = g_detail_id[li_cnt].detail_id
            END IF

            CALL lnode_var.setAttribute("record",ls_temp_detailid)
            IF ls_temp_detailid.getIndexOf("s_detail",1) THEN
               CALL lnode_var.setAttribute("page_id",ls_temp_detailid.subString(9,ls_temp_detailid.getLength()))
            END IF

            #Q類增加 - 因為用到ls_temp_detailid - 拆分part1
            IF g_prog_type.subString(1,1) = "q" THEN
               LET lnode_fvar = lnode_formfield.createChild("var") 
               CALL lnode_fvar.setAttribute("id","body")
               #取得該單身頁碼 (page)
               CALL lnode_fvar.setAttribute("page",li_cnt)
               CALL lnode_fvar.setAttribute("record",ls_temp_detailid)
               IF ls_temp_detailid.getIndexOf("s_detail",1) THEN
                  CALL lnode_fvar.setAttribute("page_id",ls_temp_detailid.subString(9,ls_temp_detailid.getLength()))
               END IF
            END IF

            #取得單身上的欄位資料 (value)
                                                 #傳入陣列 取DEFINE
            LET ls_temp = sadzp030_tab_de(ls_temp_detailid,1,g_detail_id[li_cnt].detail_tab)

            #將欄位清單送入取值陣列
            CALL sadzp030_tab_ma_1(ls_temp,"D",li_cnt) RETURNING g_detail_id[li_cnt].cluster

            #特別欄位形態處理
            LET ls_temp = sadzp030_tab_ma_change_type(ls_temp,"D")
            IF g_t100debug >= 6 THEN
               DISPLAY "INFO: 單身上的欄位:",ls_temp
            END IF
            CALL lnode_var.setAttribute("value",ls_temp)

            #檢查是否有site欄位存在
            IF g_prog_type = "i02" OR g_prog_type = "t02" OR
               g_prog_type = "q01" OR g_prog_type = "q02" OR g_prog_type = "q03" OR g_prog_type = "q04" OR 
               g_prog_type = "c02a" OR g_prog_type = "c02b" OR g_prog_type = "c04a" THEN 
               IF ls_temp.getIndexOf(ls_site_col,1) THEN
                  LET li_chk = TRUE
               END IF
            END IF

            #Q01類增加 - 因為用到ls_temp - 拆分part2
            IF g_prog_type.subString(1,1) = "q" THEN
               #特別欄位形態處理
                                                 #傳入陣列 取DEFINE
               LET ls_temp = sadzp030_tab_de(ls_temp_detailid,6,g_detail_id[li_cnt].detail_tab)
               CALL lnode_fvar.setAttribute("value",ls_temp)
            END IF

         END FOR

         #檢查i02畫面上是否出現site欄位,若出現時site=一般PK處理(g_table_main_pk不異動)
         IF g_prog_type = "i02" OR g_prog_type = "t02" OR
            g_prog_type = "q02" OR
            g_prog_type = "c02a" OR g_prog_type = "c02b" OR g_prog_type = "c04a" THEN 
            IF NOT li_chk THEN
               LET g_table_main_pk = sadzp030_tab_relation_site(g_table_main_pk)
            END IF
         END IF
      END IF
   
      #建置 var id="bs_field"  i01/i04/i05/i07/i08/t01 
      #範例:<var id="bs_field" value="ooad001,ooae003"   pk="ooad001"  order="ooad001"
      #                          欄位(不需要的部分濾除)  WHERE用的問號對應   排序依據  
      #          ref_table="ooae_t"    ref_type="left" 
      #                  有參考的部分        左側參照
      #          ref_wc="ooae001 = ooad001 AND ooae002 = $varg_dlang$var" />
      #             參照時需要用到的WHERE條件,用$var包起來可以調用全域變數
      IF g_prog_type = "i01" OR g_prog_type = "i04" OR g_prog_type = "i05" OR
         g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i13" OR
         g_prog_type = "t01" THEN

         LET lnode_var = lnode_section.createChild("var")

         CALL lnode_var.setAttribute("id","bs_field") 

         #取得 s_browse 上的欄位資料
         LET ls_browse_col = sadzp030_tab_de_browse("s_browse")

         #將欄位清單送入取值陣列 (使用b-recover調整一下字串)
         CALL sadzp030_tab_ma_1(sadzp030_tab_brecover(ls_browse_col),"M",0) RETURNING ls_temp

#        LET ls_browse_col = ls_browse_col.subString(1,ls_browse_col.getLength()-1)
         CALL lnode_var.setAttribute("value",ls_browse_col)

         CALL lnode_var.setAttribute("pk",g_table_main_pk)

         #範例:<bs_reference field="xxxx123_desc" ref_sql="參照SQL指令" ref_field="參照主欄位" />
         CALL sadzp030_chk_browseref_col(ls_browse_col) RETURNING li_chk,la_reference
         IF li_chk THEN
            FOR li_cnt = 1 TO la_reference.getLength()
               LET l_bsreference = lnode_section.createChild("bs_reference")

               #輸出放reference的欄位
               CALL l_bsreference.setAttribute("field",la_reference[li_cnt].field)
               #取用畫面上做對照的欄位
               CALL l_bsreference.setAttribute("ref_field",la_reference[li_cnt].ref_field)
               #取這些欄位內, 是否有去做 reference 用的 where 條件
               CALL l_bsreference.setAttribute("ref_sql",la_reference[li_cnt].ref_sql)
            END FOR
         END IF

         #取這些欄位內, 預設用來排序的欄位是哪個
         CALL lnode_var.setAttribute("order","")
      END IF

   #建置 structure -> section id="prog_init"
   LET lnode_section = lnode_structure.createChild("section")
   CALL lnode_section.setAttribute("id","prog_init")

      #建置 屬性 property id="bgjob" 背景程式
      LET lnode_action = lnode_section.createChild("property")
      CALL lnode_action.setAttribute("id","bgjob")
      CALL lnode_action.setAttribute("type","N")

   #建置 structure -> section id="menu"
   LET lnode_section = lnode_structure.createChild("section")
   CALL lnode_section.setAttribute("id","menu")

   #選出DIALOG層的action
   LET ls_token = sadzp030_tab_action("db")
   LET l_token = base.StringTokenizer.create(ls_token,",")
   
   WHILE l_token.hasMoreTokens()
      LET ls_temp = l_token.nextToken()
      LET lnode_action = lnode_section.createChild("action")

      CALL lnode_action.setAttribute("id",ls_temp.subString(1,ls_temp.getIndexOf("(",1)-1))
      IF ls_temp.getIndexOf("(N)",1) THEN
         CALL lnode_action.setAttribute("chk","N")
      ELSE
         #增加output出現時同步出現 quickprint
         IF ls_temp = "output(Y)" THEN
            LET lnode_action_branch = lnode_section.createChild("action")
            CALL lnode_action_branch.setAttribute("id","quickprint")
         END IF
      END IF

      LET ls_temp = ls_temp.subString(1,ls_temp.getIndexOf("(",1)-1)
      #以下action 須加上 standard
      IF ls_temp = "insert" OR ls_temp = "query" OR ls_temp = "delete" OR
         ls_temp = "modify" OR ls_temp = "reproduce" THEN
         CALL lnode_action.setAttribute("type","standard")
      END IF

      #以下action 濾除
      IF ls_temp = "first" OR ls_temp = "next" OR ls_temp = "jump" OR
         ls_temp = "previous" OR ls_temp = "last" THEN
         CALL lnode_section.removeChild(lnode_action)
      END IF

      #預防在下列樣板中，又特別拉出exporttoexcel功能
      IF ls_temp = "exporttoexcel" THEN
         IF g_prog_type = "i01" OR g_prog_type = "i02" OR g_prog_type = "i04" OR g_prog_type = "i07" OR 
            g_prog_type = "i08" OR g_prog_type = "i09" OR g_prog_type = "i10" OR 
            g_prog_type = "i12" OR g_prog_type = "q01" OR g_prog_type = "q02" OR g_prog_type = "q04" OR 
            g_prog_type = "t01" OR g_prog_type = "t02" OR
            g_prog_type.subString(g_prog_type.getLength(),g_prog_type.getLength()) = "a" THEN
            CALL lnode_section.removeChild(lnode_action)
         END IF
      END IF

      #預防特別拉出quickprint功能
      IF ls_temp = "quickprint" THEN
         IF g_prog_type = "i01" OR g_prog_type = "i02" OR g_prog_type = "i04" OR g_prog_type = "i07" OR 
            g_prog_type = "i08" OR g_prog_type = "i09" OR g_prog_type = "i10" OR 
            g_prog_type = "i12" OR g_prog_type = "q01" OR g_prog_type = "q02" OR g_prog_type = "q04" OR 
            g_prog_type = "t01" OR g_prog_type = "t02" OR
            g_prog_type.subString(g_prog_type.getLength(),g_prog_type.getLength()) = "a" THEN
            CALL lnode_section.removeChild(lnode_action)
         END IF
      END IF

      #Q類的分頁要特別取消
      IF ls_temp = "detail_first" OR ls_temp = "detail_next" OR ls_temp = "detail_previous" OR
         ls_temp = "detail_last" OR ls_temp = "qbehidden" THEN
         IF g_prog_type = "q01" OR g_prog_type = "q02" OR g_prog_type = "q03" OR g_prog_type = "q04" THEN
            CALL lnode_section.removeChild(lnode_action)
         END IF
      END IF

      IF ls_temp = "datainfo" THEN
         IF g_prog_type = "q04" THEN
            CALL lnode_section.removeChild(lnode_action)
         END IF
      END IF

      IF g_t100debug >= 6 THEN
         DISPLAY "INFO: 取得功能action-ID:",ls_temp
      END IF
   END WHILE

   #串查action (不含Q類)
   IF NOT g_prog_type.subString(1,1) = "q" THEN
      CALL lc_dzal.clear()
      LET lc_dzal = sadzp030_tab_ma_qry_dzal("M")

      FOR li_cnt = 1 TO lc_dzal.getLength()
         LET lnode_action = lnode_section.createChild("action")
         CALL lnode_action.setAttribute("id",lc_dzal[li_cnt].dzal002 CLIPPED)
         IF g_t100debug >= 6 THEN
            DISPLAY "INFO: 取得串查action-ID:",lc_dzal[li_cnt].dzal002
         END IF

         LET ls_temp = lc_dzal[li_cnt].dzal006 CLIPPED    #程式參考設定
         LET ls_temp = ls_temp.trim()

         IF lc_dzal[li_cnt].dzal007 = "1" THEN

            #如果參考設定有引號,先解開
            IF ls_temp.subString(1,1) = '"' OR ls_temp.subString(1,1) = "'" THEN
               LET ls_temp = ls_temp.subString(2,ls_temp.getLength()-1)
            END IF

            #如果參數寫在控鍵欄位內,就不把依附控鍵再加上去了(若有需要user要自行加入dzal006)
            IF ls_temp.getIndexOf(" ",1) THEN
               CALL lnode_action.setAttribute("prog",ls_temp.subString(1,ls_temp.getIndexOf(" ",1)-1))
               CALL lnode_action.setAttribute("parameter",ls_temp.subString(ls_temp.getIndexOf(" ",1)+1,
                                                                            ls_temp.getLength()))
            ELSE
               LET ls_temp = lc_dzal[li_cnt].dzal006 CLIPPED    #程式參考設定
               IF ls_temp.subString(1,1) = '"' OR ls_temp.subString(1,1) = "'" THEN
                  LET ls_temp = ls_temp.subString(2,ls_temp.getLength()-1)
               END IF
               CALL lnode_action.setAttribute("prog",ls_temp)                               #程式參考
               LET ls_temp = lc_dzal[li_cnt].dzal005 CLIPPED
               IF ls_temp.getIndexOf(".",1) THEN
                  LET ls_temp = ls_temp.subString(ls_temp.getIndexOf(".",1)+1,ls_temp.getLength())
               END IF
               CALL lnode_action.setAttribute("parameter",ls_temp)  #依附欄位資料作為參數傳入
            END IF

         ELSE  #聯絡人串查
            LET ls_temp = lc_dzal[li_cnt].dzal006 CLIPPED
            LET ls_temp = ls_temp.subString(1,ls_temp.getIndexOf("(",1)-1)
            CALL lnode_action.setAttribute("lib",ls_temp)

            #處理關聯的欄位變數 prog_imaa001 取成 g_imaa_m.imaa001
            LET ls_temp = lc_dzal[li_cnt].dzal002 CLIPPED
            IF ls_temp.getIndexOf("prog_",1) THEN
               LET ls_contact_col = ls_temp.subString(ls_temp.getIndexOf("prog_",1)+5,ls_temp.getLength())
               LET ls_temp = ls_contact_col.trim()
               LET ls_contact_col = "g_",sadzp030_tab_ma_find_tableid(ls_contact_col) 
               LET ls_contact_col = ls_contact_col.subString(1,ls_contact_col.getLength()-2),"_m.",ls_temp
            ELSE
               LET ls_contact_col = "''"
            END IF
            
            LET ls_temp = lc_dzal[li_cnt].dzal006 CLIPPED
            LET ls_temp = ls_temp.subString(ls_temp.getIndexOf("(",1)+1,
                                            ls_temp.getIndexOf(")",1)-1),",",ls_contact_col.trim()
            CALL lnode_action.setAttribute("parameter",ls_temp)
         END IF
      END FOR
   END IF

   #產出各s_detail的 ACTION
   IF g_prog_type = "i02" OR g_prog_type = "i07" OR g_prog_type = "i08" OR
      g_prog_type = "i09" OR
      g_prog_type = "i12" OR g_prog_type = "t01" OR g_prog_type = "t02" OR
      g_prog_type = "q01" OR g_prog_type = "q02" OR g_prog_type = "q03" OR g_prog_type = "q04" OR
      g_prog_type = "c02a" OR g_prog_type = "c02b" OR
      g_prog_type = "c03a" OR g_prog_type = "c03b" OR
      g_prog_type = "c04a" THEN

      IF g_prog_type.subString(1,1) = "q" THEN
         CALL lc_dzal.clear()
         LET lc_dzal = sadzp030_tab_ma_qry_dzal("D")
      END IF

      #個別DISPLAY ARRAY段的 ON ACTION
      FOR li_cnt = 1 TO g_detail_id.getLength()
       # LET ls_detailid = "db",li_cnt USING "<<<" 
         LET ls_detailid = g_detail_id[li_cnt].detail_id
         LET ls_detailid = "db",ls_detailid.subString(9,ls_detailid.getLength())
         #選出個別子層的action
         LET ls_token = sadzp030_tab_action(ls_detailid)
         IF ls_token.getLength() > 0 OR g_detail_id[li_cnt].cluster.getLength() > 0 THEN
            #建置 structure -> section id="detail_show"
            LET lnode_section = lnode_structure.createChild("section")
            CALL lnode_section.setAttribute("id","detail_show")
            CALL lnode_section.setAttribute("page",li_cnt USING "<<<")

            #功能項目
            IF ls_token.getLength() > 0 THEN
               LET l_token = base.StringTokenizer.create(ls_token, ",")
               WHILE l_token.hasMoreTokens()
                  LET ls_temp = l_token.nextToken()
                  LET lnode_action = lnode_section.createChild("action")
                  CALL lnode_action.setAttribute("id",ls_temp.subString(1,ls_temp.getIndexOf("(",1)-1))
                  IF ls_temp.getIndexOf("(N)",1) THEN
                     CALL lnode_action.setAttribute("chk","N")
                  END IF
               END WHILE
            END IF

            #串查項目
            IF g_detail_id[li_cnt].cluster.getLength() > 0 THEN

               IF g_prog_type.subString(1,1) = "q" THEN
                  FOR li_dzal = 1 TO lc_dzal.getLength()

                     LET ls_query_temp = lc_dzal[li_dzal].dzal006
                     IF ls_query_temp.getIndexOf("cl_user_contact(",1) THEN
                       #IF g_t100debug >= 3 THEN
                           DISPLAY "注意: Q類串查在欄位:",lc_dzal[li_dzal].dzal002 CLIPPED,"偵測到使用",ls_query_temp,"元件,排除!!"
                       #END IF
                     ELSE
                        LET lnode_action = lnode_section.createChild("cluster")
                        CALL lnode_action.setAttribute("id",lc_dzal[li_dzal].dzal002 CLIPPED)  #控件編號

                        #程式限制:一定要找到b_開頭的,但是傳進去要濾除b_
                        IF lc_dzal[li_dzal].dzal005[1,2] = "b_" THEN
                           LET ls_temp = lc_dzal[li_dzal].dzal005 CLIPPED
                           LET lc_dzal[li_dzal].dzal005 = ls_temp.subString(3,ls_temp.getLength())
                        ELSE
                           IF g_t100debug >= 3 THEN
                              DISPLAY "注意: Q類串查被串欄位名稱有問題(非屬於b_開頭)",ls_temp
                           END IF
                        END IF
                        CALL lnode_action.setAttribute("qry_field",lc_dzal[li_dzal].dzal005 CLIPPED) #依附控件
                        CALL lnode_action.setAttribute("prog",lc_dzal[li_dzal].dzal006 CLIPPED)  #作業設定
                        CALL lnode_action.setAttribute("parameter",lc_dzal[li_dzal].dzal005 CLIPPED) #傳入參數欄位
                     END IF                                                                             #暫時依照依附控件
                  END FOR
               END IF

               #非Q類-2015/4/30修正為Q類也要做
        #      IF g_prog_type.subString(1,1) <> "q" OR FGL_GETENV("ZONE") = "t10dit" THEN

                  LET ls_query_temp = g_detail_id[li_cnt].cluster
                  IF ls_query_temp.getIndexOf("cl_user_contact(",1) THEN
                     #IF g_t100debug >= 3 THEN
                         DISPLAY "注意: 串查偵測到使用",ls_query_temp,"元件,排除!!"
                     #END IF
                  ELSE
                     LET lnode_action = lnode_section.createChild("action")
                     CALL lnode_action.setAttribute("id","detail_qrystr")
                     CALL lnode_action.setAttribute("mode","popup")

                     LET l_token = base.StringTokenizer.create(g_detail_id[li_cnt].cluster, ",")
                     WHILE l_token.hasMoreTokens()
                        LET ls_temp = l_token.nextToken()
                        LET lnode_cluster = lnode_action.createChild("action")
                        CALL lnode_cluster.setAttribute("id","prog_"||ls_temp.subString(1,ls_temp.getIndexOf("(",1)-1))
                        CALL lnode_cluster.setAttribute("prog",ls_temp.subString(1,ls_temp.getIndexOf("(",1)-1))
                        LET ls_temp = ls_temp.subString(ls_temp.getIndexOf("(",1)+1,ls_temp.getLength()-1)
                        IF ls_temp.getIndexOf(".",1) THEN
                           LET ls_temp = ls_temp.subString(ls_temp.getIndexOf(".",1)+1,ls_temp.getLength())
                        END IF
                        IF g_prog_type.subString(1,1) = "q" THEN
                           IF ls_temp.subString(1,2) = "b_" THEN
                              LET ls_temp = ls_temp.subString(3,ls_temp.getLength())
                           END IF
                        END IF
                        CALL lnode_cluster.setAttribute("parameter",ls_temp)
                     END WHILE
                  END IF
        #      END IF
            END IF

            #confirm items
            IF lnode_section.getChildCount() = 0 THEN
               CALL lnode_structure.removeChild(lnode_section)
            END IF
         END IF
      END FOR

      #個別INPUT ARRAY段的 ON ACTION
      FOR li_cnt = 1 TO g_detail_id.getLength()
        #LET ls_detailid = "di",li_cnt USING "<<<" 
         LET ls_detailid = g_detail_id[li_cnt].detail_id
         LET ls_detailid = "di",ls_detailid.subString(9,ls_detailid.getLength())
         #選出個別子層的action
         LET ls_token = sadzp030_tab_action(ls_detailid)

         #若存在子層的action, 或有該子層的多語言欄位update_item則產生
         IF ls_token.getLength() > 0 OR g_mdatalang.getLength() > 0 THEN
            #建置 structure -> section id="detail_input"
            LET lnode_section = lnode_structure.createChild("section")
            CALL lnode_section.setAttribute("id","detail_input")
            CALL lnode_section.setAttribute("page",li_cnt USING "<<<")

            IF ls_token.getLength() > 0 THEN
               LET l_token = base.StringTokenizer.create(ls_token, ",")
               WHILE l_token.hasMoreTokens()
                  LET ls_temp = l_token.nextToken()
                  LET lnode_action = lnode_section.createChild("action")
                  CALL lnode_action.setAttribute("id",ls_temp.subString(1,ls_temp.getIndexOf("(",1)-1))
                  IF ls_temp.getIndexOf("(N)",1) THEN
                     CALL lnode_action.setAttribute("chk","N")
                  END IF
               END WHILE
            END IF

            IF g_mdatalang.getLength() > 0 THEN
               LET lnode_action = lnode_section.createChild("action")
               CALL lnode_action.setAttribute("id","update_item")
            END IF
         END IF
      END FOR
   END IF

   #建置 structure -> section id="master-input"
   IF g_prog_type = "i01" OR g_prog_type = "i04" OR g_prog_type = "i05" OR
      g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR 
      g_prog_type = "i10" OR g_prog_type = "i12" OR g_prog_type = "i13" OR
      g_prog_type = "t01" OR g_prog_type = "p01" OR g_prog_type = "r01" OR
      g_prog_type = "c01a" OR g_prog_type = "c01b" OR
      g_prog_type = "c03a" OR g_prog_type = "c03b" THEN

      LET ls_token = sadzp030_tab_action("mi")
      IF g_datalang.getLength() > 0 OR ls_token.getLength() > 0 THEN
         LET lnode_section = lnode_structure.createChild("section")
         CALL lnode_section.setAttribute("id","master_input")

         IF g_datalang.getLength() > 0 THEN
            LET lnode_action = lnode_section.createChild("action")
            CALL lnode_action.setAttribute("id","update_item")
         END IF
         IF ls_token.getLength() > 0 THEN
            LET l_token = base.StringTokenizer.create(ls_token, ",")
            WHILE l_token.hasMoreTokens()
               LET ls_temp = l_token.nextToken()
               LET lnode_action = lnode_section.createChild("action")
               CALL lnode_action.setAttribute("id",ls_temp.subString(1,ls_temp.getIndexOf("(",1)-1))
               IF ls_temp.getIndexOf("(N)",1) THEN
                  CALL lnode_action.setAttribute("chk","N")
               END IF
            END WHILE
         END IF
      END IF
   END IF

END FUNCTION


#分析傳入s_browse的清單中是否有reference欄位, 有則輸出 array
PRIVATE FUNCTION sadzp030_chk_browseref_col(ls_browse_col)

   DEFINE ls_browse_col  STRING
   DEFINE l_token1       base.StringTokenizer
   DEFINE ls_next1       STRING
   DEFINE ls_orignal     STRING  #原有欄位名稱
   DEFINE la_refreturn   DYNAMIC ARRAY OF RECORD
             field       STRING,
             ref_sql     STRING,
             ref_field   STRING
                         END RECORD
   DEFINE li_chk         LIKE type_t.num5
   DEFINE li_pos         LIKE type_t.num5
   DEFINE ls_refcol      STRING
   DEFINE ls_reftable    STRING
   DEFINE ls_wc          STRING
   DEFINE ls_temp        STRING
   DEFINE ls_fixid,ls_fixdata STRING
   DEFINE ls_wcfix       STRING
   DEFINE lc_dzai        RECORD 
             dzai001     LIKE dzai_t.dzai001,  #規格編號         dzaj001
             dzai002     LIKE dzai_t.dzai002,  #控件編號         dzaj002
             dzai003     LIKE dzai_t.dzai003,  #識別碼版次       dzaj003
             dzai004     LIKE dzai_t.dzai004,  #使用標示         dzaj004
             dzai005     LIKE dzai_t.dzai005,  #依附控件編號     dzaj005
             dzai007     LIKE dzai_t.dzai007,  #對應傳值設定     dzaj007
             dzai008     LIKE dzai_t.dzai008,  #資料參考table    dzaj008
             dzai009     LIKE dzai_t.dzai009,  #資料參考fk       dzaj009
             dzai010     LIKE dzai_t.dzai010,  #資料參考語系     dzaj010
             dzai011     LIKE dzai_t.dzai011,  #資料參考回傳     dzaj011
             dzai012     LIKE dzai_t.dzai012   #客戶代號         dzaj012
                         END RECORD

   CALL la_refreturn.clear()

   #確認傳入的欄位清單,個別在TSD內是否已經有設定參考欄位資料
   LET l_token1 = base.StringTokenizer.create(ls_browse_col, ",")
   WHILE l_token1.hasMoreTokens()
      LET ls_next1 = l_token1.nextToken()

      #原來截出來會有 abcd001_desc(chr80),取除結尾的 (chr80)
      IF ls_next1.getIndexOf("(chr80)",3) THEN
         LET ls_next1 = ls_next1.subString(1,ls_next1.getIndexOf("(chr80)",3)-1)

      #230223-00034 ---start---
      ELSE
         #修正221011-00017調整而產生的問題
         #原先是chr80，現已改為chr200
         #因此若chr80找不當到的話，要再檢查一下是否有chr200的

         #原來截出來會有 abcd001_desc(chr200),取除結尾的 (chr200)
         IF ls_next1.getIndexOf("(chr200)",3) THEN
            LET ls_next1 = ls_next1.subString(1,ls_next1.getIndexOf("(chr200)",3)-1)
         END IF
      #230223-00034 --- end ---
      END IF

      # 範例:<bs_reference field="xxxx123_desc" ref_sql="參照SQL指令" ref_field="參照欄位" />
      IF ls_next1.getIndexOf("_desc",4) THEN
         LET li_pos = la_refreturn.getLength() + 1

         #field
         LET la_refreturn[li_pos].field = ls_next1

         #原始名稱查看版次
         LET ls_orignal = "b_",ls_next1
         LET lc_dzai.dzai003 = sadzp030_tab_ma_2(ls_orignal,"dzaa004") #版號
         LET lc_dzai.dzai004 = sadzp030_tab_ma_2(ls_orignal,"dzaa006") #使用標示
 
         #如果取出的板號或使用標示,有任何一個不存在,濾除b_查看版次
         IF lc_dzai.dzai003 IS NULL OR lc_dzai.dzai004 IS NULL THEN
            DISPLAY "注意: ",ls_orignal,"去除b_後才能抓到版次"
            LET lc_dzai.dzai003 = sadzp030_tab_ma_2(ls_next1,"dzaa004") #版號
            LET lc_dzai.dzai004 = sadzp030_tab_ma_2(ls_next1,"dzaa006") #使用標示
         END IF

         LET ls_next1 = ls_next1.subString(1,ls_next1.getIndexOf("_desc",4)-1)
         LET la_refreturn[li_pos].ref_field = ls_next1 #此處的和上方的 ls_next1不同,差 "_desc"

         #如果取出的板號或使用標示,有任何一個不存在,則用去除_desc在查一次
         IF lc_dzai.dzai003 IS NULL OR lc_dzai.dzai004 IS NULL THEN
            LET lc_dzai.dzai003 = sadzp030_tab_ma_2(ls_next1,"dzaa004") #版號
            LET lc_dzai.dzai004 = sadzp030_tab_ma_2(ls_next1,"dzaa006") #使用標示
         END IF

         LET lc_dzai.dzai002 = la_refreturn[li_pos].field

         #browser端有做參考欄位,但在主頁面上可能是參考(dzai),也可能是多語言(dzaj)
         LET ls_temp = lc_dzai.dzai002          #先備份起來
         LET lc_dzai.dzai002 = "b_",lc_dzai.dzai002  #調整為正確的 browser欄位

         SELECT dzai001,dzai002,dzai003,dzai004,dzai005,dzai007,dzai008,dzai009,dzai010,
                dzai011,dzai012
           INTO lc_dzai.* FROM dzai_t,dzaa_t
         #---------------------------dzai--------------#
          WHERE dzai001 = g_prog_id       #規格編號
            AND dzai002 = lc_dzai.dzai002 #控件編號
            AND dzai003 = lc_dzai.dzai003 #識別碼版號
            AND dzai004 = lc_dzai.dzai004 #使用標示
#           AND dzai012 = g_cust          #客戶編號
         #---------------------------dzaa--------------#
            AND dzaa001 = dzai001         #規格編號
            AND dzaa002 = g_sd_ver        #規格版次
            AND dzaa003 = dzai002         #控件編號
            AND dzaa004 = dzai003         #識別碼版次
            AND dzaa005 = '6'             #識別碼類型
            AND dzaa006 = dzai004         #使用標示
            AND dzaa009 = g_dgenv         #識別標示
            AND dzaastus = "Y"
         IF SQLCA.SQLCODE THEN
            #試著用同字尾找看看 (沒設browser看看可否從主頁面抓來)
            IF g_t100debug >= 3 THEN
               DISPLAY "注意: 欄位:",lc_dzai.dzai002,"不存在dzai/dzaj表, 查看主表是否有設定(",SQLCA.SQLCODE,")"
            END IF

            LET lc_dzai.dzai002 = "%.",ls_temp.subString(1,ls_temp.getIndexOf("_desc",1)-1)
            SELECT dzaj001,dzaj002,dzaj003,dzaj004,dzaj005,dzaj007,dzaj008,dzaj009,dzaj010,
                   dzaj011,dzaj012
              INTO lc_dzai.* FROM dzaj_t 
             WHERE dzaj001 = g_prog_id          #規格編號
               AND dzaj002 LIKE lc_dzai.dzai002 #控件編號
               AND dzaj003 = lc_dzai.dzai003    #識別碼版號
               AND dzaj004 = lc_dzai.dzai004    #使用標示
#              AND dzaj012 = g_cust             #客戶編號
            IF SQLCA.SQLCODE THEN
               IF g_t100debug >= 3 THEN
                  DISPLAY "注意: 欄位:",lc_dzai.dzai002,"不存在dzai/dzaj表(",SQLCA.SQLCODE,")"
               END IF
               #什麼都找不到,清除此筆資料
               CALL la_refreturn.deleteElement(li_pos)
               CONTINUE WHILE
            END IF
         END IF

         LET ls_refcol   = lc_dzai.dzai011    #參考取值欄位
         LET ls_reftable = lc_dzai.dzai008    #參考table
         LET ls_temp     = sadzp030_tab_relation_ent(lc_dzai.dzai009)  #參考關係欄位
         LET ls_wc       = sadzp030_tab_relation_ent(lc_dzai.dzai007)  #對照給值欄位
         IF g_t100debug >= 9 THEN
            DISPLAY "INFO: dzai002待設定欄位:",lc_dzai.dzai002
            DISPLAY "      dzai011參考取值欄位:",ls_refcol
            DISPLAY "      dzai008參考table:",ls_reftable
            DISPLAY "      dzai009參考關係欄位:",ls_temp
         END IF
 
         #檢查存在逗號隔開時, 表示就是多key
         IF ls_temp.getIndexOf(",",3) THEN
            #多KEY的狀況下,重取ref_field
            #多KEY的狀態下要檢查是否有固定值夾雜 (變化型) (傳回fix欄位id及資料,準備回寫wc)
            CALL sadzp030_tab_chk_fixcond_ref(ls_wc) RETURNING ls_wc,ls_fixid,ls_fixdata

            #SQL中需要的被參照欄位PK資料 (支持多KEY)
            LET la_refreturn[li_pos].ref_field = ls_wc

            #變化型的wc修正 產出固定值修正的wc 與 ls_temp(欄位清單)
            IF ls_fixid.getLength() > 0 THEN
               CALL sadzp030_tab_fixcond_genwc(ls_fixid,ls_fixdata,ls_temp) RETURNING ls_wcfix,ls_temp
            ELSE
               #ls_wcfix由此而起,因此若上面func無執行,此處需要清空
               LET ls_wcfix = ""
            END IF

            #多KEY的ls_temp修正 (傳入ref_fk)
            CALL sadzp030_tab_get_ref_mkey(ls_temp) RETURNING ls_wc,ls_temp

            #wc = 固定值修正 + 多KEY修正 (固定資料可以一次刪掉比較多東西,放前面效率比較好)
            LET ls_wc = ls_wcfix, ls_wc
         ELSE
            #單KEY的 ref_field不需要重取,使用原來的就可以
            LET ls_wc = ls_temp,"=? "
         END IF

         #browser段落回傳的欄位名稱都要把 b_ 濾除, 因此整理一下
         LET la_refreturn[li_pos].ref_field = sadzp030_tab_cut_b(la_refreturn[li_pos].ref_field)

         #語言別欄位指定 
         IF lc_dzai.dzai010 CLIPPED IS NOT NULL THEN
            IF lc_dzai.dzai010[1,2] = "gz" OR lc_dzai.dzai010[1,2] = "dz" THEN
               LET ls_wc = ls_wc,"AND ",lc_dzai.dzai010 CLIPPED,"=$varg_lang$var"
            ELSE
               LET ls_wc = ls_wc,"AND ",lc_dzai.dzai010 CLIPPED,"=$varg_dlang$var"
            END IF
         END IF

         #ls_wc的環境修正 (補上ent)
         LET ls_temp = sadzp030_tab_entfix(ls_reftable)
         #真的有的話,在修正一下
         IF ls_temp.getLength() > 0 THEN
            LET ls_temp = ls_temp.subString(1,ls_temp.getIndexOf("=",3)-1),"=$varg_enterprise$var AND "
         END IF
         LET ls_wc = ls_temp,ls_wc
         LET ls_temp = sadzp030_tab_entfix(ls_reftable)

         #組合查詢SQL段落
         LET ls_temp = "SELECT ",ls_refcol," FROM ",ls_reftable," WHERE ",ls_wc
         LET la_refreturn[li_pos].ref_sql = ls_temp
         IF g_t100debug >= 6 THEN
            DISPLAY "IFNO: 欄位(",ls_next1,")參考,產出SQL:",ls_temp
         END IF
      ELSE
         CONTINUE WHILE
      END IF

      #若發生錯誤則刪除
      IF ls_refcol.getLength() < 1 OR ls_reftable.getLength() < 1 OR ls_wc.getLength() < 1 THEN
         IF g_t100debug >= 3 THEN
            DISPLAY "注意: b欄位(",ls_next1,")參考,資料不全,略過此項.SQL=",ls_temp
         END IF
         CALL la_refreturn.deleteElement(li_pos)
      END IF
   END WHILE

   IF la_refreturn.getLength() > 0 THEN
      LET li_chk = TRUE
   ELSE
      LET li_chk = FALSE
   END IF
   
   #若不符合,則顯示WARNING訊息
   RETURN li_chk,la_refreturn
END FUNCTION


#將傳入的欄位清單,濾除 b_ 後傳出
PRIVATE FUNCTION sadzp030_tab_cut_b(ls_cols)

   DEFINE l_token   base.StringTokenizer
   DEFINE ls_cols   STRING
   DEFINE ls_next   STRING
   DEFINE ls_return STRING

   LET l_token = base.StringTokenizer.create(ls_cols CLIPPED,",")
   LET ls_return = ""

   WHILE l_token.hasMoreTokens()
      LET ls_next = l_token.nextToken()
      IF ls_next.subString(1,2) = "b_" AND ls_next.getLength() > 3 THEN
         LET ls_next = ls_next.subString(3,ls_next.getLength())
      END IF
      LET ls_return = ls_return,ls_next,","
   END WHILE

   RETURN ls_return.subString(1,ls_return.getLength()-1)

END FUNCTION


#多KEY的分解
PRIVATE FUNCTION sadzp030_tab_get_ref_mkey(ls_reffk)

   DEFINE ls_reffk   STRING
   DEFINE ls_lastcol STRING
   DEFINE ls_wc      STRING
   DEFINE l_token1   base.StringTokenizer
   DEFINE ls_next1   STRING

   LET l_token1 = base.StringTokenizer.create(ls_reffk, ",")
   LET ls_wc = ""
   
   WHILE l_token1.hasMoreTokens()
      LET ls_next1 = l_token1.nextToken()
      LET ls_lastcol = ls_next1
      LET ls_wc = ls_wc,ls_next1,"=? AND "
   END WHILE

   RETURN ls_wc.subString(1,ls_wc.getLength()-4),ls_lastcol
END FUNCTION


#抓取多語言欄位 
PRIVATE FUNCTION sadzp030_get_langual_col(ls_colany)

   DEFINE ls_colany  STRING
   DEFINE ls_collang STRING
   DEFINE li_pos     LIKE type_t.num5
   DEFINE li_cnt     LIKE type_t.num5
   DEFINE lc_chr     LIKE type_t.chr1

   #傳入值安全性檢查:未傳入任何值,則直接回傳空
   IF ls_colany.getLength() < 1 THEN
      RETURN ""
   END IF

   #直接取流水號+1 (假定語言別欄位都是在PK後方第一個)
   FOR li_pos = 1 TO ls_colany.getLength()
      LET lc_chr = ls_colany.subString(li_pos,li_pos)
      IF lc_chr MATCHES "[0-9]" THEN 
         EXIT FOR
      END IF
   END FOR
 
   LET li_cnt = ls_colany.subString(li_pos,ls_colany.getLength()) + 1
   LET ls_collang = ls_colany.subString(1,li_pos-1),li_cnt USING "&&&"

   #此處應該要回頭檢查該欄位是否為語言別欄位

   RETURN ls_collang

END FUNCTION

#+ 確認PK欄位是否完全存在 master清單中 (全到:TRUE/缺少:FALSE)
PRIVATE FUNCTION sadzp030_tab_check_real_t01(ls_master,ls_pk) 
   DEFINE ls_master  STRING
   DEFINE ls_pk      STRING
   DEFINE li_status  LIKE type_t.num5  #檢查結果
   DEFINE l_token    base.StringTokenizer
   DEFINE ls_next    STRING

   #設定初始值
   LET li_status = TRUE

   LET l_token = base.StringTokenizer.create(ls_pk, ",")
   
   WHILE l_token.hasMoreTokens()
      LET ls_next = l_token.nextToken()
      IF NOT ls_master.getIndexOf(ls_next,1) THEN
         LET li_status = FALSE
         EXIT WHILE
      END IF
   END WHILE

   RETURN li_status
END FUNCTION


#分析i07 TSD傳入的兩組字樣,挑出都有出現的字段 (pk放前面,畫面欄位放後面)
PRIVATE FUNCTION sadzp030_tab_analyze_pk_i07(ls_list1,ls_list2)

   DEFINE ls_next1   STRING
   DEFINE ls_list1   STRING
   DEFINE ls_list2   STRING
   DEFINE l_token1   base.StringTokenizer
   DEFINE ls_common  STRING

   LET l_token1 = base.StringTokenizer.create(ls_list1, ",")
   
   WHILE l_token1.hasMoreTokens()
      LET ls_next1 = l_token1.nextToken()
      IF ls_list2.getIndexOf(ls_next1,1) THEN
         LET ls_common = ls_common,",",ls_next1
         CONTINUE WHILE
      END IF
   END WHILE

   RETURN ls_common.subString(2,ls_common.getLength())

END FUNCTION


#產生form結構
PRIVATE FUNCTION sadzp030_tab_gen_form(p_form)

   DEFINE p_form        om.DomNode
   DEFINE l_section     om.DomNode
   DEFINE l_init        om.DomNode
   DEFINE l_input       om.DomNode
   DEFINE l_construct   om.DomNode
   DEFINE l_set         om.DomNode
   DEFINE l_column      om.DomNode
   DEFINE l_cnt         LIKE type_t.num10
   DEFINE l_cnt1        LIKE type_t.num10
   DEFINE ls_master_col STRING
   DEFINE ls_detail_col STRING
   DEFINE l_tok         base.StringTokenizer
   DEFINE ls_col        STRING
   DEFINE ls_master_pr01 STRING             #p/r01兩個樣板有INPUT就沒有CONSTRUCT的特殊功能
   DEFINE l_fuzzyquery  om.DomNode   #201109-00009
   DEFINE ls_browse_col  STRING      #201109-00009
   DEFINE ls_fuzzyquery_col_bs   STRING   #201109-00009
   DEFINE ls_field      STRING       #201109-00009
   #240930-00022 ---start---
   DEFINE l_value_formula  om.DomNode
   DEFINE ls_target_col STRING
   DEFINE ls_vf         LIKE type_t.num5
   #240930-00022 --- end ---


   #畫面 編號
   CALL p_form.setAttribute("id","0")
   #畫面 所屬模組
   CALL p_form.setAttribute("module",g_prog_id[1,3])
   #畫面 檔名
   CALL p_form.setAttribute("filename",g_prog_id)
   #畫面 使用方法
   CALL p_form.setAttribute("method","dialog")
#  #這個畫面是否存在site欄位的顯示
#  IF g_siteshow = 1 THEN
#     CALL p_form.setAttribute("site","Y")
#  END IF
 
   #建置 form -> section id="input"  輸入段落
   LET l_section = p_form.createChild("section")
   CALL l_section.setAttribute("id","input")

      #input區塊中的 head 單頭輸入 i01/i04/i05/i07/i08/i09/i10/i12/i13/t01
      IF g_prog_type = "i01" OR g_prog_type = "i04" OR g_prog_type = "i05" OR
         g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR
         g_prog_type = "i10" OR g_prog_type = "i12" OR g_prog_type = "i13" OR
         g_prog_type = "t01" OR g_prog_type = "p01" OR g_prog_type = "r01" OR
         g_prog_type = "c01a" OR g_prog_type = "c01b" OR
         g_prog_type = "c03a" OR g_prog_type = "c03b" THEN

         LET l_input = l_section.createChild("input")
         CALL l_input.setAttribute("id","head")

         #單頭的欄位
         IF g_prog_type = "p01" OR g_prog_type = "r01" THEN
            LET ls_master_col = sadzp030_tab_ma("vb_qbe",2,g_table_main,TRUE)
            LET ls_master_pr01 = ls_master_col   #p/r01兩個樣板有INPUT就沒有CONSTRUCT的特殊功能
         ELSE
            LET ls_master_col = sadzp030_tab_ma("mainlayout",2,g_table_main,FALSE),",",
                                sadzp030_tab_ma("patch_grid",2,g_table_main,FALSE)
            IF ls_master_col.subString(ls_master_col.getLength(),ls_master_col.getLength())="," THEN
               LET ls_master_col = ls_master_col.subString(1,ls_master_col.getLength()-1)
            END IF
         END IF

         CALL l_input.setAttribute("field",ls_master_col.trim())
      END IF

      #input區塊中的 body 單身輸入 i02/i03/i04/i07/i08/i09/i12/t01/t02/q02/q03
      IF g_prog_type = "i02" OR g_prog_type = "i03" OR g_prog_type = "i04" OR 
         g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR
         g_prog_type = "i12" OR g_prog_type = "t01" OR g_prog_type = "t02" OR 
         g_prog_type = "q01" OR g_prog_type = "q02" OR g_prog_type = "q03" OR g_prog_type = "q04" OR
         g_prog_type = "c02a" OR g_prog_type = "c02b" OR
         g_prog_type = "c03a" OR g_prog_type = "c03b" OR
         g_prog_type = "c04a" THEN

         #區分多頁簽
         FOR l_cnt = 1 TO g_detail_id.getLength()
            LET l_input = l_section.createChild("input")
            CALL l_input.setAttribute("id","body")

            #頁碼
            CALL l_input.setAttribute("page",l_cnt)

            #欄位
            IF g_prog_type = "i03" THEN
               #i03樣板,由於Genero還沒有開放TREE使用INPUT ARRAY,所以暫時另外製作一個輸入
               #區塊,提供做為資料輸入使用,因此可以輸入的欄位應該要看該區塊為主
              #LET ls_detail_col = sadzp030_tab_get_i03_input(g_4fdNode.doctailNode)
            ELSE
               LET ls_detail_col =              #傳入陣列                 取INPUT欄位
                   sadzp030_tab_de(g_detail_id[l_cnt].detail_id,2,g_detail_id[l_cnt].detail_tab)
               #填入欄位收集陣列,單身固定從2開始
               LET g_col_set[l_cnt+1].pos = g_detail_id[l_cnt].detail_id
               IF g_prog_type = "q01" OR g_prog_type = "q03" OR g_prog_type = "q04" THEN
                  LET g_col_set[l_cnt+1].fields = #ls_detail_col
                      sadzp030_tab_de(g_detail_id[l_cnt].detail_id,6,g_detail_id[l_cnt].detail_tab)
               ELSE
                  #推翻了....
                  LET g_col_set[l_cnt+1].fields = #ls_detail_col
                      sadzp030_tab_de(g_detail_id[l_cnt].detail_id,7,g_detail_id[l_cnt].detail_tab)
               END IF
               #濾除括號(括號內為定義)
               LET g_col_set[l_cnt+1].fields = s_adzp030_tab_remove_quoter(g_col_set[l_cnt+1].fields CLIPPED)
            END IF

            CALL l_input.setAttribute("field", ls_detail_col)
            CALL l_input.setAttribute("insert",g_detail_id[l_cnt].detail_ins) 
            CALL l_input.setAttribute("append",g_detail_id[l_cnt].detail_app)
            CALL l_input.setAttribute("delete",g_detail_id[l_cnt].detail_del)

            #檢討：如果傳出的欄位全部都只剩下共用欄位,；則刪除本條目
            IF sadzp030_tab_check_comm_for_deny(ls_detail_col) THEN
               CALL l_section.removeChild(l_input)
            END IF
         END FOR
      END IF

      #q04的樣板不要留下"input-body"段落
      IF g_prog_type = "q04" THEN
         CALL p_form.removeChild(l_section)
      END IF

   #建置 form -> section id="construct"  查詢段落
   LET l_section = p_form.createChild("section")
   CALL l_section.setAttribute("id","construct")

      #construct區塊中的 head 單頭輸入  i01/i04/i05/i07/i08/i09/i10/i12/i13/t01
      IF g_prog_type = "i01" OR g_prog_type = "i04" OR g_prog_type = "i05" OR
         g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR
         g_prog_type = "i10" OR g_prog_type = "i12" OR g_prog_type = "i13" OR
         g_prog_type = "t01" OR g_prog_type = "p01" OR g_prog_type = "r01" OR
         g_prog_type = "q01" OR g_prog_type = "q04" OR
         g_prog_type = "c01a" OR g_prog_type = "c01c" OR
         g_prog_type = "c03a" OR g_prog_type = "c03c" THEN

         LET l_construct = l_section.createChild("construct")
         CALL l_construct.setAttribute("id","head")

         #單頭的欄位 (從mainlayout往下抓)
         CASE
            WHEN g_prog_type = "p01" OR g_prog_type = "r01"
               LET ls_master_col = sadzp030_tab_ma("vb_qbe",3,g_table_main,FALSE)
               #p/r01兩個樣板有INPUT就沒有CONSTRUCT的特殊功能
      #        LET ls_master_col = sadzp030_tab_notexists(ls_master_col,ls_master_pr01)
            WHEN g_prog_type = "q01" OR g_prog_type = "q04"
               LET ls_master_col = sadzp030_tab_ma("vb_master",3,g_table_main,FALSE)
            OTHERWISE
               LET ls_master_col = sadzp030_tab_ma("mainlayout",3,g_table_main,FALSE),",",
                                   sadzp030_tab_ma("patch_grid",3,g_table_main,FALSE)
               IF ls_master_col.subString(ls_master_col.getLength(),ls_master_col.getLength())="," THEN
                  LET ls_master_col = ls_master_col.subString(1,ls_master_col.getLength()-1)
               END IF
         END CASE

      #  IF ls_master_col IS NULL THEN
      #     CALL p_form.removeChild(l_section)
      #  ELSE
            CALL l_construct.setAttribute("field",ls_master_col.trim())
      #  END IF
      END IF

      #construct區塊中的 body 單身輸入 i02/i03/i04/i07/i08/i09/i12/t01/t02/q01/q02/q03/q04
      IF g_prog_type = "i02" OR g_prog_type = "i03" OR g_prog_type = "i04" OR 
         g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR
         g_prog_type = "i12" OR g_prog_type = "t01" OR g_prog_type = "t02" OR
         g_prog_type = "q01" OR g_prog_type = "q02" OR g_prog_type = "q03" OR g_prog_type = "q04" OR
         g_prog_type = "c02a" OR g_prog_type = "c02c" OR 
         g_prog_type = "c03a" OR g_prog_type = "c03c" OR g_prog_type = "c04a" THEN

         #區分多頁簽
         FOR l_cnt = 1 TO g_detail_id.getLength()
            #欄位                                 #傳入陣列                      取CONSTRUCT欄位
            LET ls_detail_col = sadzp030_tab_de(g_detail_id[l_cnt].detail_id,3,g_detail_id[l_cnt].detail_tab)

            IF ls_detail_col IS NOT NULL THEN
               LET l_construct = l_section.createChild("construct")
               CALL l_construct.setAttribute("id","body")

               #頁碼
               CALL l_construct.setAttribute("page",l_cnt)

               #當樣板屬於q01,q03,q04的時候,table需要濾除qbe(bs_field)不存在的項目

               IF g_t100debug >= 6 THEN
                  DISPLAY "INFO: construct區塊中的 body 單身輸入:",ls_detail_col
               END IF
               CALL l_construct.setAttribute("field",ls_detail_col)
            END IF
         END FOR
      END IF

   #201109-00009 ---start---
   #建置 form -> init id="fuzzyquery"  #模糊查詢
   LET l_init = p_form.createChild("init")
   CALL l_init.setAttribute("id","fuzzyquery")

      #construct區塊中的 head 單頭模糊查詢輸入  i01/i04/i05/i07/i08/i09/i10/i12/i13/t01
      IF g_prog_type = "i01" OR g_prog_type = "i04" OR g_prog_type = "i05" OR
         g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR
         g_prog_type = "i10" OR g_prog_type = "i12" OR g_prog_type = "i13" OR
         g_prog_type = "t01" OR g_prog_type = "p01" OR g_prog_type = "r01" OR
         g_prog_type = "q01" OR g_prog_type = "q04" OR
         g_prog_type = "c01a" OR g_prog_type = "c01c" OR
         g_prog_type = "c03a" OR g_prog_type = "c03c" THEN

         LET l_fuzzyquery = l_init.createChild("fuzzyquery")
         CALL l_fuzzyquery.setAttribute("id","head")

         #單頭的欄位 (從mainlayout往下抓)
         CASE
            WHEN g_prog_type = "p01" OR g_prog_type = "r01"
               LET ls_master_col = sadzp030_tab_ma("vb_qbe",7,g_table_main,FALSE)
               #p/r01兩個樣板有INPUT就沒有CONSTRUCT的特殊功能
      #        LET ls_master_col = sadzp030_tab_notexists(ls_master_col,ls_master_pr01)
            WHEN g_prog_type = "q01" OR g_prog_type = "q04"
               LET ls_master_col = sadzp030_tab_ma("vb_master",7,g_table_main,FALSE)
            OTHERWISE
               LET ls_master_col = sadzp030_tab_ma("mainlayout",7,g_table_main,FALSE),",",
                                   sadzp030_tab_ma("patch_grid",7,g_table_main,FALSE)
               IF ls_master_col.subString(ls_master_col.getLength(),ls_master_col.getLength())="," THEN
                  LET ls_master_col = ls_master_col.subString(1,ls_master_col.getLength()-1)
               END IF
         END CASE

      #  IF ls_master_col IS NULL THEN
      #     CALL p_form.removeChild(l_section)
      #  ELSE
            CALL l_fuzzyquery.setAttribute("field",ls_master_col.trim())
      #  END IF
      END IF

      #construct區塊中的 body 單身模糊查詢輸入 i02/i03/i04/i07/i08/i09/i12/t01/t02/q01/q02/q03/q04
      IF g_prog_type = "i02" OR g_prog_type = "i03" OR g_prog_type = "i04" OR
         g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR
         g_prog_type = "i12" OR g_prog_type = "t01" OR g_prog_type = "t02" OR
         g_prog_type = "q01" OR g_prog_type = "q02" OR g_prog_type = "q03" OR g_prog_type = "q04" OR
         g_prog_type = "c02a" OR g_prog_type = "c02c" OR
         g_prog_type = "c03a" OR g_prog_type = "c03c" OR g_prog_type = "c04a" THEN

         #區分多頁簽
         FOR l_cnt = 1 TO g_detail_id.getLength()
            #欄位                                 #傳入陣列                      取CONSTRUCT欄位
            LET ls_detail_col = sadzp030_tab_de(g_detail_id[l_cnt].detail_id,8,g_detail_id[l_cnt].detail_tab)
            IF ls_detail_col IS NOT NULL THEN
               LET l_fuzzyquery = l_init.createChild("fuzzyquery")
               CALL l_fuzzyquery.setAttribute("id","body")

               #頁碼
               CALL l_fuzzyquery.setAttribute("page",l_cnt)

               #當樣板屬於q01,q03,q04的時候,table需要濾除qbe(bs_field)不存在的項目

               IF g_t100debug >= 6 THEN
                  DISPLAY "INFO: construct區塊中的 body 單身模糊查詢:",ls_detail_col
               END IF
               CALL l_fuzzyquery.setAttribute("field",ls_detail_col)
            END IF
         END FOR
      END IF

      #201217 HJ說filter的部分暫緩處理，先賦予基本功能可以運行即可
   #  #browser頁籤過濾(filter)功能
   #  IF g_prog_type = "i01" OR g_prog_type = "i04" OR g_prog_type = "i05" OR
   #     g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i13" OR
   #     g_prog_type = "t01" THEN

   #
   #     LET ls_browse_col = sadzp030_tab_de_browse("s_browse")
   #     LET l_tok = base.StringTokenizer.create(ls_browse_col, ',')
   #     WHILE l_tok.hasMoreTokens()
   #        LET ls_col = l_tok.nextToken()
   #        #非主要table欄位
   #        IF ls_col.getIndexOf('(',1) > 0 THEN
   #           LET ls_field = ls_col.subString(1,ls_col.getIndexOf('(',1)-1)
   #           IF NOT cl_getField(ls_master_col,ls_field) THEN
   #              CONTINUE WHILE
   #           ELSE
   #              #因為是browser頁籤，所以要把b_加回來
   #              LET ls_fuzzyquery_col_bs = ls_fuzzyquery_col_bs, "b_",ls_field,","
   #           END IF
   #        ELSE
   #           #因為是browser頁籤，所以要把b_加回來
   #           LET ls_fuzzyquery_col_bs = ls_fuzzyquery_col_bs, "b_",ls_col,","
   #        END IF
   #     END WHILE

   #     LET ls_fuzzyquery_col_bs = ls_fuzzyquery_col_bs.subString(1,ls_fuzzyquery_col_bs.getLength()-1)

   #     LET l_fuzzyquery = l_init.createChild("fuzzyquery")
   #     CALL l_fuzzyquery.setAttribute("id","bs_field")

   #     IF g_t100debug >= 6 THEN
   #        DISPLAY "INFO: fuzzyquery區塊中的 browser 頁籤模糊查詢:",ls_fuzzyquery_col_bs
   #     END IF
   #     CALL l_fuzzyquery.setAttribute("field",ls_fuzzyquery_col_bs)

   #  END IF
   #201109-00009 --- end ---

   #240930-00022 ---start---
   #建置 form -> init id="value_formula"  #簡易帶值及公式
   LET l_init = p_form.createChild("init")
   CALL l_init.setAttribute("id","value_formula")

   FOR ls_vf = 1 TO 2
      #input區塊中的 head 單頭輸入 i01/i04/i05/i07/i08/i09/i10/i12/i13/t01
      #簡易帶值及公式僅在input時才有作用 (新增或修改)，
      #且僅針對ITM類樣板，因為其他類別的樣板不會有input的動作
      IF g_prog_type = "i01" OR g_prog_type = "i04" OR g_prog_type = "i05" OR
         g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR
         g_prog_type = "i10" OR g_prog_type = "i12" OR g_prog_type = "i13" OR
         g_prog_type = "t01" OR
         g_prog_type = "c01a" OR g_prog_type = "c01b" OR
         g_prog_type = "c03a" OR g_prog_type = "c03b" THEN

         IF ls_vf = 1 THEN
            #簡易帶值
            LET l_value_formula = l_init.createChild("value_formula_v")
            CALL l_value_formula.setAttribute("id","head")
         ELSE
            #公式
            LET l_value_formula = l_init.createChild("value_formula_f")
            CALL l_value_formula.setAttribute("id","head")
         END IF

         LET ls_target_col = NULL
         LET ls_master_col = NULL

         IF ls_vf = 1 THEN   #簡易帶值
            #單頭的欄位
            LET ls_target_col = sadzp030_tab_ma("mainlayout",8,g_table_main,FALSE),",",
                                sadzp030_tab_ma("patch_grid",8,g_table_main,FALSE)
            IF ls_target_col.subString(ls_target_col.getLength(),ls_target_col.getLength())="," THEN
               LET ls_target_col = ls_target_col.subString(1,ls_target_col.getLength()-1)
            END IF

            LET ls_master_col = sadzp030_tab_ma("mainlayout",9,g_table_main,FALSE),",",
                                sadzp030_tab_ma("patch_grid",9,g_table_main,FALSE)
            IF ls_master_col.subString(ls_master_col.getLength(),ls_master_col.getLength())="," THEN
               LET ls_master_col = ls_master_col.subString(1,ls_master_col.getLength()-1)
            END IF

         ELSE   #公式
            #單頭的欄位
            LET ls_target_col = sadzp030_tab_ma("mainlayout",10,g_table_main,FALSE),",",
                                sadzp030_tab_ma("patch_grid",10,g_table_main,FALSE)

            IF ls_target_col.subString(ls_target_col.getLength(),ls_target_col.getLength())="," THEN
               LET ls_target_col = ls_target_col.subString(1,ls_target_col.getLength()-1)
            END IF

            LET ls_master_col = sadzp030_tab_ma("mainlayout",11,g_table_main,FALSE),",",
                                sadzp030_tab_ma("patch_grid",11,g_table_main,FALSE)
            IF ls_master_col.subString(ls_master_col.getLength(),ls_master_col.getLength())="," THEN
               LET ls_master_col = ls_master_col.subString(1,ls_master_col.getLength()-1)
            END IF
         END IF

         CALL l_value_formula.setAttribute("target_field",ls_target_col.trim())
         CALL l_value_formula.setAttribute("field",ls_master_col.trim())
      END IF

      #input區塊中的 body 單身輸入 i02/i03/i04/i07/i08/i09/i12/t01/t02
      #簡易帶值及公式僅在input時才有作用 (新增或修改)，
      #且僅針對ITM類樣板，因為其他類別的樣板不會有input的動作
      IF g_prog_type = "i02" OR g_prog_type = "i03" OR g_prog_type = "i04" OR
         g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR
         g_prog_type = "i12" OR g_prog_type = "t01" OR g_prog_type = "t02" OR
         g_prog_type = "c02a" OR g_prog_type = "c02b" OR
         g_prog_type = "c03a" OR g_prog_type = "c03b" OR
         g_prog_type = "c04a" THEN

         #區分多頁簽
         FOR l_cnt = 1 TO g_detail_id.getLength()
            LET ls_target_col = NULL
            LET ls_master_col = NULL

            IF ls_vf = 1 THEN   #簡易帶值
               LET ls_target_col =              #傳入陣列                 取INPUT欄位
                   sadzp030_tab_de(g_detail_id[l_cnt].detail_id,9,g_detail_id[l_cnt].detail_tab)

               LET ls_detail_col =              #傳入陣列                 取INPUT欄位
                   sadzp030_tab_de(g_detail_id[l_cnt].detail_id,10,g_detail_id[l_cnt].detail_tab)
            ELSE   #公式
               LET ls_target_col =              #傳入陣列                 取INPUT欄位
                   sadzp030_tab_de(g_detail_id[l_cnt].detail_id,11,g_detail_id[l_cnt].detail_tab)

               LET ls_detail_col =              #傳入陣列                 取INPUT欄位
                   sadzp030_tab_de(g_detail_id[l_cnt].detail_id,12,g_detail_id[l_cnt].detail_tab)
            END IF

         #  IF ls_target_col IS NOT NULL AND ls_detail_col IS NOT NULL THEN
               IF ls_vf = 1 THEN   #簡易帶值
                  LET l_value_formula = l_init.createChild("value_formula_v")
                  CALL l_value_formula.setAttribute("id","body")
               ELSE   #公式
                  LET l_value_formula = l_init.createChild("value_formula_f")
                  CALL l_value_formula.setAttribute("id","body")
               END IF

               #頁碼
               CALL l_value_formula.setAttribute("page",l_cnt)

               IF g_t100debug >= 6 THEN
                  DISPLAY "INFO: input區塊中的 body 單身簡易帶值或公式:",ls_detail_col
               END IF
               CALL l_value_formula.setAttribute("target_field",ls_target_col)
               CALL l_value_formula.setAttribute("field",ls_detail_col)
         #  END IF
         END FOR
      END IF
   END FOR
   #240930-00022 --- end ---

   #建置 form -> init id="field_set"  輸入管制
   LET l_init = p_form.createChild("init")
   CALL l_init.setAttribute("id","field_set")

      #head 單頭部分輸入管制  i01/i04/i05/i07/i08/i09/i10/i12/i13/t01
      IF g_prog_type = "i01" OR g_prog_type = "i04" OR g_prog_type = "i05" OR
         g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR
         g_prog_type = "i10" OR g_prog_type = "i12" OR g_prog_type = "i13" OR
         g_prog_type = "t01" OR g_prog_type = "p01" OR g_prog_type = "r01" OR
         g_prog_type = "c01a" OR g_prog_type = "c03a" THEN

         #准入
         LET l_set = l_init.createChild("set")
         CALL l_set.setAttribute("id","head")
         CALL l_set.setAttribute("type","entry")
         CALL l_set.setAttribute("field",g_table_main_pk)

         #禁入
         LET l_set = l_init.createChild("set")
         CALL l_set.setAttribute("id","head")
         CALL l_set.setAttribute("type","no_entry")
         CALL l_set.setAttribute("field",g_table_main_pk)
      END IF

      #head 單頭部分輸入管制   i02/i04/i07/i08/i09/i12/t01/t02/q02
      IF g_prog_type = "i02" OR g_prog_type = "i04" OR g_prog_type = "i07" OR
         g_prog_type = "i08" OR g_prog_type = "i09" OR g_prog_type = "i12" OR 
         g_prog_type = "t01" OR g_prog_type = "t02" OR
         g_prog_type = "q02" OR
         g_prog_type = "c02a" OR g_prog_type = "c02b" OR
         g_prog_type = "c03a" OR g_prog_type = "c03b" OR g_prog_type = "c04a" THEN

         #准入
         LET l_set = l_init.createChild("set")
         CALL l_set.setAttribute("id","body")
         CALL l_set.setAttribute("type","entry")
         CALL l_set.setAttribute("field","")

         #禁入
         LET l_set = l_init.createChild("set")
         CALL l_set.setAttribute("id","body")
         CALL l_set.setAttribute("type","no_entry")
         CALL l_set.setAttribute("field","")
      END IF

   #建置 form -> init id="head" 單頭欄位細項設定   i01/i04/i07/i08/i09/i10/i12/i13/t01
   IF g_prog_type = "i01" OR g_prog_type = "i04" OR g_prog_type = "i05" OR
      g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR
      g_prog_type = "i10" OR g_prog_type = "i12" OR g_prog_type = "i13" OR
      g_prog_type = "t01" OR g_prog_type = "p01" OR g_prog_type = "r01" OR
      g_prog_type = "c01a" OR g_prog_type = "c01b" OR g_prog_type = "c01c" OR
      g_prog_type = "c03a" OR g_prog_type = "c03b" OR g_prog_type = "c03c" THEN

      LET l_init = p_form.createChild("init")
      CALL l_init.setAttribute("id","head")

      #由單頭欄位清單 g_col_set[1].fields 輸出
      LET l_tok = base.StringTokenizer.create(g_col_set[1].fields.trim(), ",")
      WHILE l_tok.hasMoreTokens()
         LET ls_col = l_tok.nextToken()
         LET l_column = l_init.createChild("column")
         CALL sadzp030_tab_gen_column(l_column,ls_col,g_col_set[1].fields.trim())
      END WHILE
   END IF

   #建置 form -> init id="body" 單身欄位細項設定 i02/i03/i04/i07/i08/i09/i12/t01/t02
   IF g_prog_type = "i02" OR g_prog_type = "i03" OR g_prog_type = "i04" OR 
      g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR 
      g_prog_type = "i12" OR g_prog_type = "t01" OR g_prog_type = "t02" OR
      g_prog_type = "q01" OR g_prog_type = "q02" OR g_prog_type = "q03" OR g_prog_type = "q04" OR
      g_prog_type = "c02a" OR g_prog_type = "c02b" OR
      g_prog_type = "c03a" OR g_prog_type = "c03b" OR g_prog_type = "c04a" THEN

      #區分多頁簽
      FOR l_cnt = 1 TO g_detail_id.getLength()

         LET l_init = p_form.createChild("init")
         CALL l_init.setAttribute("id","body")

         #頁碼
         CALL l_init.setAttribute("page",l_cnt)
                               #col_set的第一筆是單頭,所以本筆單身真正的位置在 cnt+1
         LET l_tok = base.StringTokenizer.create(g_col_set[l_cnt+1].fields.trim(), ",")
         LET l_cnt1 = 1
         WHILE l_tok.hasMoreTokens()
            LET ls_col = l_tok.nextToken()
            LET l_column = l_init.createChild("column")
            CALL sadzp030_tab_gen_column(l_column,ls_col,g_col_set[l_cnt+1].fields.trim())
         END WHILE
      END FOR
   END IF
END FUNCTION

#檢查傳進來的欄位是否都是不可輸入的,例如共用欄位
PRIVATE FUNCTION sadzp030_tab_check_comm_for_deny(ls_detail_col)
   DEFINE ls_detail_col STRING
   DEFINE ls_cleaned    STRING
   DEFINE li_return     LIKE type_t.num5
   DEFINE l_token1      base.StringTokenizer
   DEFINE ls_next1      STRING

   LET ls_cleaned = ""
   LET l_token1 = base.StringTokenizer.create(ls_detail_col,",")
   
   WHILE l_token1.hasMoreTokens()
      LET ls_next1 = l_token1.nextToken()
      IF ls_next1.getIndexOf("modid",1) OR ls_next1.getIndexOf("moddt",1) OR
         ls_next1.getIndexOf("posid",1) OR ls_next1.getIndexOf("posdt",1) OR
         ls_next1.getIndexOf("cnfid",1) OR ls_next1.getIndexOf("cnfdt",1) OR
         ls_next1.getIndexOf("ownid",1) OR ls_next1.getIndexOf("owndp",1) OR
         ls_next1.getIndexOf("crtid",1) OR ls_next1.getIndexOf("crtdt",1) OR
         ls_next1.getIndexOf("crtdp",1) THEN
         CONTINUE WHILE
      ELSE
         LET ls_cleaned = ls_cleaned,ls_next1,","
      END IF
   END WHILE

   IF ls_cleaned.getLength() = 0 THEN
      LET li_return = TRUE
   ELSE
      LET li_return = FALSE
   END IF

   RETURN li_return
END FUNCTION

#產生column下的結構                       待產生的node,欄位名稱, 本組欄位總表(確認ref用)
PRIVATE FUNCTION sadzp030_tab_gen_column(lnode_column,ls_columnid,ls_setcols)

   DEFINE lnode_column    om.DomNode
   DEFINE lnode_after     om.DomNode
   DEFINE lnode_controlp  om.DomNode
   DEFINE ls_columnid     STRING
   DEFINE ls_default      STRING
   DEFINE ls_controlp     STRING
   DEFINE ls_sql          STRING
   DEFINE ls_wc           STRING
   DEFINE ls_refcol       STRING
   DEFINE ls_reftable     STRING
   DEFINE ls_temp         STRING
   DEFINE ls_fixdata      STRING
   DEFINE ls_fixid        STRING
   DEFINE ls_wcfix        STRING
   DEFINE ls_bt,ls_st     STRING
   DEFINE lc_dzai RECORD LIKE dzai_t.*
   DEFINE lc_gzcc001     LIKE gzcc_t.gzcc001  #表單編號
   DEFINE lc_gzcc002     LIKE gzcc_t.gzcc002  #欄位名稱
   DEFINE lc_gzcc003     LIKE gzcc_t.gzcc003  #狀態碼的類別 (串SCC表)
   DEFINE lc_gzcc004     LIKE gzcc_t.gzcc004  #要填入程式中的狀態碼資料值
   DEFINE lc_gzcb004     LIKE gzcb_t.gzcb004  #SCC表內的程式名稱
   DEFINE ls_check       STRING               #校驗帶值 dzac011
   DEFINE lc_dzcd001     LIKE dzcd_t.dzcd001  #校驗帶值組別編號
   DEFINE lc_dzcd005     LIKE dzcd_t.dzcd005  #校驗帶值型態
   DEFINE li_stt,li_btt  LIKE type_t.num5     #最大或最小是否含等於符號
   DEFINE lc_dzep011     LIKE dzep_t.dzep011  #系統分類碼
   DEFINE ls_setcols     STRING               #欄位清單
   DEFINE ls_dzak007     STRING               #取用助記碼
   DEFINE li_cnt         LIKE type_t.num5
   DEFINE li_v_chk       LIKE type_t.num5     #校驗帶值碼的檢查
   DEFINE ls_dzaa005     LIKE dzaa_t.dzaa005  #檢查是否有多語言的開窗設定   #180411-00062
   DEFINE ls_str         STRING               #180411-00062
   DEFINE ls_lang_set    LIKE type_t.chr1     #180411-00062


   #id 欄位代碼
   CALL lnode_column.setAttribute("id",ls_columnid)

   #狀態碼處理
   IF ls_columnid.getIndexOf("stus",5) AND NOT ls_columnid.getIndexOf("l_",1) THEN
      LET lc_gzcc001 = ls_columnid.subString(1,ls_columnid.getIndexOf("stus",5)-1),"_t"
      LET lc_gzcc002 = ls_columnid   #狀態碼欄位名稱

      #先進行數量檢查-如果有此欄位卻又沒有設定狀態碼,則會無法產出程式,顯示ERROR
      SELECT COUNT(gzcc003) INTO li_cnt FROM gzcc_t     #狀態SCC編號/狀態值
       WHERE gzcc001 = lc_gzcc001  #表格編號
         AND gzcc002 = lc_gzcc002  #欄位編號
         AND ( gzcc005 = "s"       #客製碼
          OR gzcc005 = "c" )
         AND gzccstus = "Y"        #只抓取有效的
      IF li_cnt = 0 THEN
         DISPLAY "ERROR: 畫面設定stus狀態碼欄位(",lc_gzcc002,"),表格",lc_gzcc001,"在adzi150的『狀態碼設定』頁簽沒有定義狀態資料,產生程序中止"
         LET g_cross_4gl_stus = FALSE      #跨4gl判斷正確與否
      ELSE
         DECLARE sadzp030_tab_get_gzcc_cs CURSOR FOR
          SELECT gzcc003,gzcc004 FROM gzcc_t     #狀態SCC編號/狀態值
           WHERE gzcc001 = lc_gzcc001  #表格編號
             AND gzcc002 = lc_gzcc002  #欄位編號
             AND ( gzcc005 = "s"       #客製碼
              OR gzcc005 = "c" )
             AND gzccstus = "Y"        #只抓取有效的
           ORDER BY gzcc006
         FOREACH sadzp030_tab_get_gzcc_cs INTO lc_gzcc003,lc_gzcc004
            LET lnode_after = lnode_column.createChild("stus")
            CALL lnode_after.setAttribute("id",lc_gzcc004)
            #抓圖片名稱
            SELECT gzcb004 INTO lc_gzcb004 FROM gzcb_t
             WHERE gzcb001 = lc_gzcc003
               AND gzcb002 = lc_gzcc004
            IF SQLCA.SQLCODE THEN
               DISPLAY "ERROR: 從azzi600第",lc_gzcc003,"組取不出狀態碼:",lc_gzcc004,"資料,請檢查r.t或azzi600"
            ELSE
               CALL lnode_after.setAttribute("pic",lc_gzcb004)
            END IF
         END FOREACH
         FREE sadzp030_tab_get_gzcc_cs
         CALL lnode_column.setAttribute("scc",lc_gzcc003)
      END IF
   END IF

   #default 新增資料時預設值
   LET ls_default = sadzp030_tab_ma_2(ls_columnid,"dzac014")     #預設值

   IF ls_default.getLength() > 0 THEN
      #type 該欄位預設值類型(S-固定內容,V-變數)
      IF ls_default.subString(1,1) = ":" THEN
         CALL lnode_column.setAttribute("type","V")

         LET ls_default = ls_default.toUpperCase()
         LET ls_default = ls_default.trim()
         CASE   #此處採用正面表列
            WHEN ls_default = ":TODAY"
               CALL lnode_column.setAttribute("default","g_today")
            WHEN ls_default = ":LANG"
               CALL lnode_column.setAttribute("default","g_lang")
            WHEN ls_default = ":DLANG"
               CALL lnode_column.setAttribute("default","g_dlang")
            WHEN ls_default = ":ENT"
               CALL lnode_column.setAttribute("default","g_enterprise")
            WHEN ls_default = ":SITE"
               CALL lnode_column.setAttribute("default","g_site")
            WHEN ls_default = ":DEPT"
               CALL lnode_column.setAttribute("default","g_dept")
            WHEN ls_default = ":USER"
               CALL lnode_column.setAttribute("default","g_user")
            WHEN ls_default = ":CURRENT"
               CALL lnode_column.setAttribute("default","cl_get_current()")
            OTHERWISE
               IF g_t100debug >= 3 THEN
                  DISPLAY "注意: 欄位",ls_columnid,"預設值使用不合法的保留字:",ls_default
               END IF
               CALL lnode_column.setAttribute("default",ls_default)
         END CASE
         #250212-00006#1 add --- start ---
         IF ls_default.getIndexOf(":TODAY", 1) > 0 THEN
            LET ls_default = cl_replace_str(ls_default, ":TODAY", "g_today")
            CALL lnode_column.setAttribute("default", ls_default)
         END IF
         #250212-00006#1 add --- end ---
      ELSE
         LET ls_temp = ls_default.toUpperCase()
         CASE
            WHEN ls_temp.getIndexOf("SYSDATE",1)
               CALL lnode_column.setAttribute("type","V")
               CALL lnode_column.setAttribute("default","cl_get_current()")
            &ifdef is_ora   #210511-00011 代追單
            WHEN ls_temp.getIndexOf("SYSTIMESTAMP",1)
               CALL lnode_column.setAttribute("type","V")
               CALL lnode_column.setAttribute("default","cl_get_timestamp()")
            &endif   #210511-00011 代追單
            #210511-00011 代追單 ---start---
            &ifdef is_pgs
            WHEN ls_temp.getIndexOf("CURRENT_TIMESTAMP",1)
               CALL lnode_column.setAttribute("type","V")
               CALL lnode_column.setAttribute("default","cl_get_timestamp()")
            &endif
            #210511-00011 代追單 --- end ---

            OTHERWISE
               CALL lnode_column.setAttribute("type","S")
               CALL lnode_column.setAttribute("default",ls_default)
         END CASE
      END IF
   END IF

   #校驗帶值
   LET ls_check = sadzp030_tab_ma_2(ls_columnid,"dzac011")

   #查看是否有設定該欄位的最大或最小值
   CALL sadzp030_tab_check_range(ls_columnid) RETURNING ls_st,li_stt,ls_bt,li_btt #最小,最大

   #檢查是否有預設的combobox選項  (濾除stus欄位)
   IF NOT ls_columnid.subString(ls_columnid.getLength()-3,ls_columnid.getLength()) = "stus" THEN
      LET lc_dzep011 = sadzp030_tab_ma_2(ls_columnid,"dzep011")  #系統分類碼
      IF lc_dzep011 > 0 THEN
         CALL lnode_column.setAttribute("def_scc",lc_dzep011)
      END IF
   END IF

   #檢查是否要有助記碼
   LET ls_dzak007 = sadzp030_tab_ma_2(ls_columnid,"dzak007")  #助記碼-搜尋欄位

   #180411-00062 ---start---
   #檢查是否有設定多語言欄位
   LET ls_dzaa005 = sadzp030_tab_ma_2(ls_columnid,"dzaa005")  #識別碼類型
   #180411-00062 --- end ---

   #設定reference,檢查若 abcd001 欄位且又存在 abcd001_desc 於畫面上時,則 enable此段落
   IF sadzp030_tab_check_reference(ls_columnid) OR ls_check.getLength() > 0 OR
      ls_st.getLength() > 0 OR ls_bt.getLength() > 0 OR ls_dzak007.getLength() > 0 OR
      ls_dzaa005 = "7" THEN   #180411-00062
      LET lnode_after = lnode_column.createChild("after")

      #取用助記碼 (要做在校驗之前)
      IF ls_dzak007 IS NOT NULL THEN
         LET lnode_controlp = lnode_after.createChild("mnemonic")
         #助記碼表格
         CALL lnode_controlp.setAttribute("table",sadzp030_tab_ma_2(ls_columnid,"dzak008"))
         #回傳欄位,組合額外查詢欄位(dzak007)+註記碼欄位(dzak009)
         LET ls_controlp = sadzp030_tab_ma_2(ls_columnid,"dzak007")
         IF ls_controlp IS NOT NULL THEN 
            LET ls_controlp = ls_controlp,",",sadzp030_tab_ma_2(ls_columnid,"dzak009")
         ELSE
            LET ls_controlp = sadzp030_tab_ma_2(ls_columnid,"dzak009")
         END IF
         CALL lnode_controlp.setAttribute("field",ls_controlp)
         #撈出回傳對應控件
         CALL lnode_controlp.setAttribute("rtn_field",sadzp030_tab_ma_2(ls_columnid,"dzak011"))
         #條件,語系查詢條件(dzak010)+額外條件(dzak005)
         LET ls_controlp = sadzp030_tab_ma_2(ls_columnid,"dzak010") CLIPPED
         IF ls_controlp IS NOT NULL THEN 
            LET ls_controlp = ls_controlp,"=$varg_dlang$var"
            IF sadzp030_tab_ma_2(ls_columnid,"dzak005") IS NOT NULL THEN
               LET ls_controlp = ls_controlp," AND "
            END IF
         END IF
         LET ls_controlp = ls_controlp,sadzp030_tab_ma_2(ls_columnid,"dzak005")
         CALL lnode_controlp.setAttribute("wc",ls_controlp)
      END IF

      #180411-00062 ---start---
      #多語言設定相關資訊
      IF ls_dzaa005 = "7" THEN
         LET lnode_controlp = lnode_after.createChild("multilang")

         #<multilang target="prcbl_t.prcbl002" source="prcb_t.prcbseq" lang_table="prcbl_t" pk="prcadocno,prcbseq" fk="prcbldocno,prcblseq" lang_field="prcbl001" rtn_field="prcbl002" />
         # tag name  控件代號                  依附控件代號            多語言表格           pk值(可多個欄位)       fk值(需與pk對應)         多語言欄位            回傳欄位(只有一個欄位)

         LET ls_lang_set = "N"
         #先找單頭的多語言設定
         FOR li_cnt = 1 TO g_multilang_m.getLength()
            IF g_multilang_m[li_cnt].columnid = ls_columnid THEN
               LET ls_lang_set = "Y"
               EXIT FOR
            ELSE
               LET ls_str = g_multilang_m[li_cnt].columnid
               IF ls_str.subString(ls_str.getIndexOf("_t",1)+3,ls_str.getLength()) = ls_columnid THEN
                  LET ls_lang_set = "Y"
                  EXIT FOR
               END IF
            END IF
         END FOR
         IF ls_lang_set = "Y" THEN
            CALL lnode_controlp.setAttribute("target",g_multilang_m[li_cnt].columnid)
            CALL lnode_controlp.setAttribute("source",g_multilang_m[li_cnt].relay_on)
            CALL lnode_controlp.setAttribute("lang_table",g_multilang_m[li_cnt].ref_table)
            CALL lnode_controlp.setAttribute("pk",g_multilang_m[li_cnt].map_col)
            LET ls_str = g_multilang_m[li_cnt].ref_fk
            LET ls_str = ls_str.subString(1,ls_str.getIndexOf(g_multilang_m[li_cnt].ref_lang,1)-2)
            CALL lnode_controlp.setAttribute("fk",ls_str)
            CALL lnode_controlp.setAttribute("lang_field",g_multilang_m[li_cnt].ref_lang)
            CALL lnode_controlp.setAttribute("rtn_field",g_multilang_m[li_cnt].ref_rtn)
         ELSE
            #若找不到，再找單身的多語言設定
            FOR li_cnt = 1 TO g_multilang_d.getLength()
               IF g_multilang_d[li_cnt].columnid = ls_columnid THEN
                  LET ls_lang_set = "Y"
                  EXIT FOR
               ELSE
                  LET ls_str = g_multilang_d[li_cnt].columnid
                  IF ls_str.subString(ls_str.getIndexOf("_t",1)+3,ls_str.getLength()) = ls_columnid THEN
                     LET ls_lang_set = "Y"
                     EXIT FOR
                  END IF
               END IF
            END FOR
            IF ls_lang_set = "Y" THEN
               CALL lnode_controlp.setAttribute("target",g_multilang_d[li_cnt].columnid)
               CALL lnode_controlp.setAttribute("source",g_multilang_d[li_cnt].relay_on)
               CALL lnode_controlp.setAttribute("lang_table",g_multilang_d[li_cnt].ref_table)
               CALL lnode_controlp.setAttribute("pk",g_multilang_d[li_cnt].map_col)
               LET ls_str = g_multilang_d[li_cnt].ref_fk
               LET ls_str = ls_str.subString(1,ls_str.getIndexOf(g_multilang_d[li_cnt].ref_lang,1)-2)
               CALL lnode_controlp.setAttribute("fk",ls_str)
               CALL lnode_controlp.setAttribute("lang_field",g_multilang_d[li_cnt].ref_lang)
               CALL lnode_controlp.setAttribute("rtn_field",g_multilang_d[li_cnt].ref_rtn)
            END IF
         END IF
      END IF
      #180411-00062 --- end ---

      #校驗帶值, 最大最小值
      IF ls_st.getLength() > 0 OR ls_bt.getLength() > 0 OR ls_check.getLength() > 0 THEN
         LET lnode_controlp = lnode_after.createChild("check")

         #<check id="range" bt="大於此值" st="小於此值" errno="錯誤代碼" ow="開窗否" />
         IF ls_st.getLength() > 0 OR ls_bt.getLength() > 0 THEN
            CALL lnode_controlp.setAttribute("id","range")
            IF ls_st.getLength() > 0 THEN  #最大值 Small Than/st
               #250212-00006#1 add --- start ---
               IF ls_st.getIndexOf(":TODAY", 1) > 0 THEN
                  LET ls_st = cl_replace_str(ls_st, ":TODAY", "g_today")
               END IF
               #250212-00006#1 add --- end ---
               CALL lnode_controlp.setAttribute("st",ls_st)
               CALL lnode_controlp.setAttribute("st_type",li_stt)
               CALL lnode_controlp.setAttribute("errno","azz-00080")
            END IF
            IF ls_bt.getLength() > 0 THEN  #最小值 Bigger Than/bt
               #250212-00006#1 add --- start ---
               IF ls_bt.getIndexOf(":TODAY", 1) > 0 THEN
                  LET ls_bt = cl_replace_str(ls_bt, ":TODAY", "g_today")
               END IF
               #250212-00006#1 add --- end ---
               CALL lnode_controlp.setAttribute("bt",ls_bt)
               CALL lnode_controlp.setAttribute("bt_type",li_btt)
               CALL lnode_controlp.setAttribute("errno","azz-00079")
            END IF
            #同時
            IF ls_bt.getLength() > 0 AND ls_st.getLength() > 0 THEN 
               CALL lnode_controlp.setAttribute("errno","azz-00087")
            END IF
            CALL lnode_controlp.setAttribute("ow","1")
         END IF

         #校驗帶值 <check id="chkandReturn" chkid="檢驗ID " />
         #純檢查   <check id="isExist" chkid="檢驗ID " />
         #純帶值   <check id="reference" chkid="檢驗ID " />
         IF ls_check.getLength() > 0 THEN
            LET li_v_chk = TRUE           #校驗帶值預設是存在的
            CALL lnode_controlp.setAttribute("chkid",ls_check)
            LET lc_dzcd001 = ls_check
            SELECT dzcd005 INTO lc_dzcd005 FROM dzcd_t
             WHERE dzcd001 = lc_dzcd001    #校驗帶值識別碼
               AND dzcd002 = "c"
         #     AND dzcdstus = "Y"          #狀態碼 dzcd_t表不須檢查狀態碼
            IF SQLCA.SQLCODE THEN
               SELECT dzcd005 INTO lc_dzcd005 FROM dzcd_t
                WHERE dzcd001 = lc_dzcd001 #校驗帶值識別碼
                  AND dzcd002 = "s"
         #        AND dzcdstus = "Y"       #狀態碼 dxcd_t表不需檢查狀態碼
               IF SQLCA.SQLCODE THEN
                  LET li_v_chk = FALSE     #確定這個校驗帶值是不存在的
               END IF
            END IF
            #顯示檢查狀態
            IF li_v_chk THEN
               IF g_t100debug >= 6 THEN
                  DISPLAY "INFO: ",lc_dzcd001,"設定校驗帶值碼:",lc_dzcd005," (1:校驗/2:帶值/其他:都有)"
               END IF
            ELSE
               DISPLAY cl_getmsg_parm("adz-01116",g_lang,lc_dzcd001)
              #DISPLAY "注意: 校驗帶值碼 ",lc_dzcd001," 並不存在系統內,請檢查 r.v 的設定"
            END IF 
            CASE
               WHEN lc_dzcd005 IS NULL
                  DISPLAY cl_getmsg_parm("adz-01117",g_lang,lc_dzcd001)
                 #DISPLAY "注意: ",lc_dzcd001,"校驗帶值碼不存在,自動修正為校驗"
                                     CALL lnode_controlp.setAttribute("id","isExist")
               WHEN lc_dzcd005 = "1" CALL lnode_controlp.setAttribute("id","isExist")
               WHEN lc_dzcd005 = "2" CALL lnode_controlp.setAttribute("id","reference")
               WHEN lc_dzcd005 = "3" CALL lnode_controlp.setAttribute("id","chkandReturn")
               OTHERWISE            
                   DISPLAY "注意: ",lc_dzcd001,"校驗帶值碼:",lc_dzcd005," 是不合法的,自動修正為校驗"
                                     CALL lnode_controlp.setAttribute("id","isExist")
            END CASE
         END IF
      END IF

      #設定reference,檢查若 abcd001 欄位且又存在 abcd001_desc 於畫面上時,則 enable此段落
      IF sadzp030_tab_check_reference(ls_columnid) THEN

         LET lnode_controlp = lnode_after.createChild("reference")

         #reference目標欄位 field
         LET ls_temp = sadzp030_tab_get_reference_col(ls_columnid.trim(),ls_setcols)
         CALL lnode_controlp.setAttribute("field",ls_temp)

         #ref SQL 取值用的SQL
         LET lc_dzai.dzai003 = sadzp030_tab_ma_2(ls_temp,"dzaa004") #版號
         LET lc_dzai.dzai004 = sadzp030_tab_ma_2(ls_temp,"dzaa006") #使用標示
         IF g_prog_type.substring(1,1) = "q" THEN                   #登錄欄位名稱 
            LET lc_dzai.dzai002 = "b_",ls_temp
         ELSE
            LET lc_dzai.dzai002 = ls_temp
         END IF

         SELECT dzai_t.* INTO lc_dzai.* FROM dzai_t,dzaa_t
          WHERE dzai001 = g_prog_id       #規格編號
            AND dzai002 = lc_dzai.dzai002 #控件編號/登錄欄位名稱
            AND dzai003 = lc_dzai.dzai003 #設計器版號
            AND dzai004 = lc_dzai.dzai004 #使用標示
#           AND dzai012 = g_cust          #客戶編號
         #---------------------------dzaa--------------#
            AND dzaa001 = dzai001         #規格編號
            AND dzaa002 = g_sd_ver        #規格版次
            AND dzaa003 = dzai002         #控件編號
            AND dzaa004 = dzai003         #識別碼版次
            AND dzaa005 = '6'             #識別碼類型
            AND dzaa006 = dzai004         #使用標示
            AND dzaa009 = g_dgenv         #識別標示
            AND dzaastus = "Y"

         IF SQLCA.SQLCODE THEN
            IF g_prog_type.substring(1,1) = "q" THEN                   #登錄欄位名稱 
               LET lc_dzai.dzai002 = ls_temp
               SELECT dzai_t.* INTO lc_dzai.* FROM dzai_t,dzaa_t
                WHERE dzai001 = g_prog_id       #規格編號
                  AND dzai002 = lc_dzai.dzai002 #控件編號/登錄欄位名稱
                  AND dzai003 = lc_dzai.dzai003 #設計器版號
                  AND dzai004 = lc_dzai.dzai004 #使用標示
#                 AND dzai012 = g_cust          #客戶編號
               #---------------------------dzaa--------------#
                  AND dzaa001 = dzai001         #規格編號
                  AND dzaa002 = g_sd_ver        #規格版次
                  AND dzaa003 = dzai002         #控件編號
                  AND dzaa004 = dzai003         #識別碼版次
                  AND dzaa005 = '6'             #識別碼類型
                  AND dzaa006 = dzai004         #使用標示
                  AND dzaa009 = g_dgenv         #識別標示
                  AND dzaastus = "Y"
#display lc_dzai.dzai002,'-->',sqlca.sqlcode
            END IF
         END IF

         IF SQLCA.SQLCODE THEN
            IF g_t100debug >= 3 THEN
               DISPLAY "注意: 欄位(",lc_dzai.dzai002,")參考,資料未設定在dzai_t表內(",sqlca.sqlcode,")"
            END IF
         ELSE
            LET lc_dzai.dzai010 = lc_dzai.dzai010 CLIPPED
            LET ls_refcol   = lc_dzai.dzai011    #參考取值欄位
            LET ls_reftable = lc_dzai.dzai008    #參考table
            LET ls_temp     = sadzp030_tab_relation_ent(lc_dzai.dzai009) #參考關係欄位
            LET ls_wc       = sadzp030_tab_relation_ent(lc_dzai.dzai007) #對照給值欄位

            #檢查存在逗號隔開時, 表示就是多key
            IF ls_temp.getIndexOf(",",3) THEN
               #多KEY的狀態下要檢查是否有固定值夾雜 (變化型) (傳回fix欄位id及資料,準備回寫wc)
               CALL sadzp030_tab_chk_fixcond_ref(ls_wc) RETURNING ls_wc,ls_fixid,ls_fixdata

               #SQL中需要的被參照欄位PK資料 (支持多KEY)
               CALL lnode_controlp.setAttribute("ref_field",ls_wc)
               CALL lnode_controlp.setAttribute("map_field",ls_wc)

               #變化型的wc修正 產出固定值修正的wc 與 ls_temp(欄位清單)
               IF ls_fixid.getLength() > 0 THEN
                  CALL sadzp030_tab_fixcond_genwc(ls_fixid,ls_fixdata,ls_temp) RETURNING ls_wcfix,ls_temp
               ELSE
                  #ls_wcfix由此而起,因此若上面func無執行,此處需要清空
                  LET ls_wcfix = ""
               END IF

               #多KEY的ls_temp修正 (傳入ref_fk)
               CALL sadzp030_tab_get_ref_mkey(ls_temp) RETURNING ls_wc,ls_temp

               #wc = 固定值修正 + 多KEY修正 + 語言別欄位指定 (固定資料可以一次刪掉比較多東西,放前面效率比較好)
               LET ls_wc = ls_wcfix, ls_wc
               IF lc_dzai.dzai010 IS NOT NULL THEN
                  IF lc_dzai.dzai010[1,2] = "gz" OR lc_dzai.dzai010[1,2] = "dz" THEN   #g_dlang/g_lang差異
                     LET ls_wc = ls_wc,"AND ", lc_dzai.dzai010 CLIPPED,"=$varg_lang$var"
                  ELSE
                     LET ls_wc = ls_wc,"AND ", lc_dzai.dzai010 CLIPPED,"=$varg_dlang$var"
                  END IF
               END IF
            ELSE
               #q類特別處理
               IF g_prog_type.subString(1,1) = "q" THEN
                  IF lc_dzai.dzai005[1,2] = "b_" THEN
                     LET lc_dzai.dzai005 = lc_dzai.dzai005[3,LENGTH(lc_dzai.dzai005)]
                  END IF
               END IF
               #SQL中需要的被參照欄位PK資料 (支持單KEY)
               CALL lnode_controlp.setAttribute("ref_field",lc_dzai.dzai005 CLIPPED)
               CALL lnode_controlp.setAttribute("map_field",lc_dzai.dzai007 CLIPPED)

               # wc="該欄位id=ref_fk AND ref_fk+1欄位=$varg_dlang$var"
               LET ls_wc = ls_temp,"=? "
               IF lc_dzai.dzai010 IS NOT NULL THEN
                  IF lc_dzai.dzai010[1,2] = "gz" OR lc_dzai.dzai010[1,2] = "dz" THEN   #g_dlang/g_lang差異
                     LET ls_wc = ls_wc ,"AND ",lc_dzai.dzai010 CLIPPED,"=$varg_lang$var"
                  ELSE
                     LET ls_wc = ls_wc ,"AND ",lc_dzai.dzai010 CLIPPED,"=$varg_dlang$var"
                  END IF
               END IF
            END IF

            #ls_wc的環境修正 (補上ent)
            LET ls_temp = sadzp030_tab_entfix(ls_reftable)
            #真的有的話,在修正一下
            IF ls_temp.getLength() > 0 THEN
               LET ls_temp = ls_temp.subString(1,ls_temp.getIndexOf("=",3)-1),"=$varg_enterprise$var AND "
            END IF
            LET ls_wc = ls_temp,ls_wc

            #組合查詢SQL段落
            LET ls_sql = "SELECT ",ls_refcol," FROM ",ls_reftable," WHERE ",ls_wc
            CALL lnode_controlp.setAttribute("ref_sql",ls_sql)
          # display "IFNO:欄位(",ls_columnid,")參考,產出SQL:",ls_sql
         END IF

         #若發生錯誤則刪除
         IF ls_refcol.getLength() < 1 OR ls_reftable.getLength() < 1 OR ls_wc.getLength() < 1 THEN
            IF g_t100debug >= 3 THEN
               DISPLAY "注意: 欄位(",ls_columnid,")參考,資料不全,略過此項.SQL=",ls_sql
            END IF
            CALL lnode_after.removeChild(lnode_controlp)
         END IF
      END IF
   END IF

   #INPUT 開窗 controlp
   LET ls_controlp = sadzp030_tab_ma_2(ls_columnid,"dzac009")   #編輯時開窗
   IF ls_controlp.getLength() > 0 THEN
      LET lnode_controlp = lnode_column.createChild("controlp")
      #form 動態開窗id
      CALL lnode_controlp.setAttribute("form",ls_controlp)
      #state 狀態
      CALL lnode_controlp.setAttribute("state","i")
   END IF

   #CONSTRUCT 開窗 controlp
   LET ls_controlp = sadzp030_tab_ma_2(ls_columnid,"dzac010")   #查詢時開窗
   IF ls_controlp.getLength() > 0 THEN
      LET lnode_controlp = lnode_column.createChild("controlp")
      #form 動態開窗id
      CALL lnode_controlp.setAttribute("form",ls_controlp)
      #state 狀態
      CALL lnode_controlp.setAttribute("state","c")
   END IF

END FUNCTION


#檢查該欄位是否有設定最大或最小值,有的話就提出
PRIVATE FUNCTION sadzp030_tab_check_range(ls_columnid)
   DEFINE ls_columnid   STRING
   DEFINE ls_bt,ls_st   STRING  #bt="大於此值" st="小於此值"
   DEFINE li_btt,li_stt LIKE type_t.num5
   DEFINE li_pos        LIKE type_t.num5
   DEFINE ls_tmp        STRING

   #若設定的資料是 數字 符號  數字, 則不濾除, 直接輸出 (明顯的設定問題)
  
   IF TRUE THEN #FGL_GETENV("ZONE") = "t10dit" THEN
      LET ls_st = sadzp030_tab_ma_2(ls_columnid,"dzac015")   #最大值
      LET ls_st = ls_st.trim()

      LET ls_tmp= sadzp030_tab_ma_2(ls_columnid,"dzac020")   #最大值符號
      LET ls_tmp= ls_tmp.trim()
      IF ls_tmp.getIndexOf("=",1) THEN LET li_stt = 1 ELSE LET li_stt = 0 END IF

      LET ls_bt = sadzp030_tab_ma_2(ls_columnid,"dzac016")   #最小值
      LET ls_bt = ls_bt.trim()

      LET ls_tmp= sadzp030_tab_ma_2(ls_columnid,"dzac021")   #最小值符號
      LET ls_tmp= ls_tmp.trim()
      IF ls_tmp.getIndexOf("=",1) THEN LET li_btt = 1 ELSE LET li_btt = 0 END IF

   ELSE
      LET ls_bt = sadzp030_tab_ma_2(ls_columnid,"dzac016")   #最小值
      LET ls_bt = ls_bt.trim()

      #若包含等號 li_btt = 1 / 若明確的只有 > , 則 li_btt = 0
      LET li_btt = 0
      FOR li_pos = 1 TO ls_bt.getLength()
        IF ls_bt.subString(li_pos,li_pos) = "=" THEN
           LET li_btt = 1
        END IF
        IF ls_bt.subString(li_pos,li_pos) MATCHES "[0-9]" OR 
           ls_bt.subString(li_pos,li_pos) = "-" THEN
           EXIT FOR
        END IF
      END FOR
      LET ls_bt = ls_bt.subString(li_pos,ls_bt.getLength())

      LET ls_st = sadzp030_tab_ma_2(ls_columnid,"dzac015")   #最大值
      LET ls_st = ls_st.trim()

      #若包含等號 li_stt = 1 / 若明確的只有 < , 則 li_stt = 0
      LET li_stt = 0
      FOR li_pos = 1 TO ls_st.getLength()
        IF ls_st.subString(li_pos,li_pos) = "=" THEN
           LET li_stt = 1
        END IF
        IF ls_st.subString(li_pos,li_pos) MATCHES "[0-9]" OR
           ls_st.subString(li_pos,li_pos) = "-" THEN
           EXIT FOR
        END IF
      END FOR
      LET ls_st = ls_st.subString(li_pos,ls_st.getLength())
   END IF

   RETURN ls_st,li_stt,ls_bt,li_btt
END FUNCTION


#傳入已知為固定欄位的位址,固定欄位值,全部的被參考欄位清單
# 如  "1,3,5"          "['1']|,['88']|,['10-10-2021']"   "aabb001,aabb002,aabboo3,aabb004,aabb005.."
#回傳 被指定值的WC 以及  濾除固定參考欄位後的欄位清單
PRIVATE FUNCTION sadzp030_tab_fixcond_genwc(ls_fixid,ls_fixdata,ls_cols)

   DEFINE ls_fixid   STRING
   DEFINE ls_fixdata STRING
   DEFINE ls_cols    STRING
   DEFINE la_fixarry DYNAMIC ARRAY OF RECORD
             chk     LIKE type_t.num5,
             colid   STRING
          END RECORD
   DEFINE l_token1      base.StringTokenizer
   DEFINE l_token2      base.StringTokenizer
   DEFINE li_pos        LIKE type_t.num5
   DEFINE ls_next1      STRING
   DEFINE ls_next2      STRING
   DEFINE ls_wcfix      STRING

   #將 ls_cols 解開到 array 內,再進行處理
   LET l_token1 = base.StringTokenizer.create(ls_cols,",")
   CALL la_fixarry.clear()
   
   WHILE l_token1.hasMoreTokens()
      LET ls_next1 = l_token1.nextToken()
      LET li_pos = la_fixarry.getLength() + 1
      LET la_fixarry[li_pos].chk = FALSE
      LET la_fixarry[li_pos].colid = ls_next1.trim()
   END WHILE

   #解開目標
   LET l_token1 = base.StringTokenizer.create(ls_fixid,",")
   LET l_token2 = base.StringTokenizer.create(ls_fixdata,"|,")
   WHILE l_token1.hasMoreTokens()
      LET ls_next1 = l_token1.nextToken()
      LET ls_next2 = l_token2.nextToken()
      #組出wc
      LET li_pos = ls_next1
      LET ls_next2 = ls_next2.subString(2,ls_next2.getLength()-1)
      LET ls_wcfix = ls_wcfix,la_fixarry[li_pos].colid.trim(),"=",ls_next2," AND "
      LET la_fixarry[li_pos].chk = TRUE
   END WHILE

   LET ls_cols = ""
   #清理用過的col並回組
   FOR li_pos = 1 TO la_fixarry.getLength()
      IF NOT la_fixarry[li_pos].chk THEN
         LET ls_cols = ls_cols,la_fixarry[li_pos].colid,","
      END IF
   END FOR

   #整理準備回傳
   LET ls_cols = ls_cols.subString(1,ls_cols.getLength()-1)

   RETURN ls_wcfix,ls_cols

END FUNCTION


#檢查傳入條件內是否含有非欄位的條件
PRIVATE FUNCTION sadzp030_tab_chk_fixcond_ref(ls_field)

   DEFINE ls_field      STRING
   DEFINE ls_retfield   STRING  #濾除固定值後的field
   DEFINE ls_columnids  STRING  #屬於固定傳入參數的位址(逗號分隔)
   DEFINE ls_datas      STRING
   DEFINE l_token1      base.StringTokenizer
   DEFINE li_pos        LIKE type_t.num5
   DEFINE ls_next1      STRING

   #逗號分隔一個個截開
   LET l_token1 = base.StringTokenizer.create(ls_field,",")
   LET li_pos = 1
   
   WHILE l_token1.hasMoreTokens()
      LET ls_next1 = l_token1.nextToken()
      IF ls_next1.getIndexOf('"',1) THEN
         DISPLAY cl_getmsg_parm("adz-01118",g_lang,ls_next1)
        #DISPLAY "ERROR: 參考欄位中的設定含有雙引號 [",ls_next1,"] 系統無法處理!請改為單引號!!"
      END IF
      #判斷應該不可以含有單引號
      IF ls_next1.getIndexOf("'",1) THEN
         #有的話就在欄位位置上記一筆 (然後就不寫入欄位清單)
         LET ls_columnids = ls_columnids,li_pos USING "<<<<<",","
         LET ls_datas = ls_datas,"[",ls_next1,"]|,"
      ELSE
         #如果沒有的話,就回傳回去欄位清單
         LET ls_retfield = ls_retfield, ls_next1,","
      END IF
   END WHILE

   #回傳值重新整理,去尾
   LET ls_retfield = ls_retfield.subString(1,ls_retfield.getLength()-1)
   LET ls_columnids = ls_columnids.subString(1,ls_columnids.getLength()-1)
   LET ls_datas = ls_datas.subString(1,ls_datas.getLength()-2)   #這裡用 |, 當分隔運算子

   RETURN ls_retfield,ls_columnids,ls_datas
END FUNCTION


#檢查現在所在欄位,是否有存在reference column在此區塊中
PRIVATE FUNCTION sadzp030_tab_check_reference(ls_columnid)
   DEFINE ls_columnid  STRING

   LET ls_columnid = ls_columnid.trim(),"_desc"  #規範就是fflabel要設定成前方欄位的id+"_desc"

   RETURN g_reference_col_set.getIndexOf(ls_columnid,1)
END FUNCTION


#產生 tab 檔內的 dataset結構
PRIVATE FUNCTION sadzp030_tab_gen_dataset(p_dataset)

   DEFINE p_dataset      om.DomNode
   DEFINE l_head         om.DomNode
   DEFINE l_body         om.DomNode
   DEFINE l_sql          om.DomNode
   DEFINE ls_sql         STRING
   DEFINE ls_master_col  STRING
   DEFINE ls_pk          STRING
   DEFINE ls_fk          STRING
   DEFINE li_cnt         LIKE type_t.num5
   DEFINE li_cnt1        LIKE type_t.num5
   DEFINE li_cnt2        LIKE type_t.num5
   DEFINE li_abc         LIKE type_t.num5
   DEFINE li_stop        LIKE type_t.num5
   DEFINE lb_stus        LIKE type_t.num5
   DEFINE lb_pass        LIKE type_t.num5
   DEFINE ls_temp        STRING
   DEFINE la_page        DYNAMIC ARRAY OF SMALLINT
   DEFINE ls_32wc        STRING
   DEFINE li_def_site    LIKE type_t.num5
   DEFINE ls_append_pk   STRING
   DEFINE ls_detail_pk   STRING
   DEFINE lst_token      base.StringTokenizer
   DEFINE ls_token       STRING 
   DEFINE ls_field       STRING  #160622
   
   #建置 dataset -> head 單頭用表 i01/i04/i05/i07/i08/i09/i10/i12/i13/t01
   IF g_prog_type = "i01" OR g_prog_type = "i04" OR g_prog_type = "i05" OR
      g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR
      g_prog_type = "i10" OR g_prog_type = "i12" OR g_prog_type = "i13" OR
      g_prog_type = "t01" OR 
      g_prog_type = "q03" OR 
    ( g_prog_type = "q04" AND (FGL_GETENV("ZONE") = "t10dev" AND g_prog_id <> "aloq005") ) OR
      g_prog_type = "c01a" OR g_prog_type = "c01b" OR g_prog_type = "c01c" OR
      g_prog_type = "c03a" OR g_prog_type = "c03b" OR g_prog_type = "c03c" THEN

      LET l_head = p_dataset.createChild("head")
      CALL l_head.setAttribute("id",g_table_main)

      #樹狀的架構需要在單頭部份加上樹的結構描述 i04/i05/i08
      IF g_prog_type = "i04" OR g_prog_type = "i05" OR g_prog_type = "i08" OR
         g_prog_type = "i13" THEN
         CALL sadzp030_tab_relation_tree_data(g_prog_type) RETURNING g_dzff.*,lb_stus
         IF lb_stus THEN 
            #common="Y" 共用欄位 預設已經是Y

            CASE 
               WHEN g_prog_type = "i04"
                  CALL l_head.setAttribute("pid",g_dzff.pid) 
                  CALL l_head.setAttribute("type",g_dzff.type)  
                  CALL l_head.setAttribute("desc",g_dzff.desc)
               WHEN g_prog_type = "i05"
                  CALL l_head.setAttribute("lid",g_dzff.id)  
                  CALL l_head.setAttribute("pid",g_dzff.pid) 
                  CALL l_head.setAttribute("type",g_dzff.type) 
                  CALL l_head.setAttribute("desc",g_dzff.desc)
                  CALL l_head.setAttribute("speed",g_dzff.speed)
                  CALL l_head.setAttribute("stype",g_dzff.stype)
                  CALL l_head.setAttribute("slid",g_dzff.slid)
                  CALL l_head.setAttribute("spid",g_dzff.spid)
               WHEN g_prog_type = "i08"
                  CALL l_head.setAttribute("lid",g_dzff.id)  
                  CALL l_head.setAttribute("pid",g_dzff.pid) 
                  CALL l_head.setAttribute("type",g_dzff.type) 
               WHEN g_prog_type = "i13"
                  CALL l_head.setAttribute("lid",g_dzff.id)  
                  CALL l_head.setAttribute("type",g_dzff.type) 
                  CALL l_head.setAttribute("type2",g_dzff.type2)  
                  CALL l_head.setAttribute("type3",g_dzff.type3)  
                  CALL l_head.setAttribute("type4",g_dzff.type4)  
                  CALL l_head.setAttribute("type5",g_dzff.type5)  
                  CALL l_head.setAttribute("type6",g_dzff.type6)  
                  CALL l_head.setAttribute("desc",g_dzff.desc)
            END CASE 
         ELSE
            DISPLAY cl_getmsg("adz-01119",g_lang)
           #DISPLAY "ERROR: 樹狀資料不正確:"
         END IF
      END IF

      #樹狀架構內若出現 pk,其功能只是用來放檢查的程式段落
      CALL l_head.setAttribute("pk",g_table_main_pk)

      #建置 sql -> forupd_sql
      LET l_sql = l_head.createChild("sql")
      CALL l_sql.setAttribute("id","forupd_sql")

         #類別
         CALL l_sql.setAttribute("type","sql")
         #指令
         IF g_prog_type = "i01" OR g_prog_type = "i04" OR g_prog_type = "i05" OR
            g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR
            g_prog_type = "i10" OR g_prog_type = "i12" OR g_prog_type = "i13" OR
            g_prog_type = "t01" OR
            g_prog_type = "c01a" OR g_prog_type = "c01b" OR g_prog_type = "c01c" OR
            g_prog_type = "c03a" OR g_prog_type = "c03b" OR g_prog_type = "c03c" THEN
            CALL l_sql.setAttribute("query",sadzp030_tab_gen_sql("forupd_sql",g_table_main))
         END IF

      #建置 sql -> cs_sql
      LET l_sql = l_head.createChild("sql")
      CALL l_sql.setAttribute("id","cs_sql")

         #類別
         CALL l_sql.setAttribute("type","construct")
         #指令
         CASE
            WHEN g_prog_type = "i01" OR g_prog_type = "i10" OR g_prog_type = "i13"
               LET ls_sql = "SELECT ", g_table_main_pk,
                             " FROM ", g_table_main
               CALL l_sql.setAttribute("query",ls_sql)

            WHEN g_prog_type = "i05" OR g_prog_type = "i07" OR g_prog_type = "i08" OR
                 g_prog_type = "i09" OR g_prog_type = "i12" OR g_prog_type = "t01" OR
                 g_prog_type = "c03a" OR g_prog_type = "c03b"
               CALL l_sql.setAttribute("query",sadzp030_tab_gen_sql("cs_sql",g_table_main))
         END CASE
         #排序
         CALL l_sql.setAttribute("order",g_table_main_pk)

      IF g_datalang.getLength() > 0 THEN
         FOR li_cnt = 1 TO g_datalang.getLength()
            #建置 sql -> append
            LET l_sql = l_head.createChild("sql")
            CALL l_sql.setAttribute("id","append")
            #型態
            CALL l_sql.setAttribute("type","lang")
            #目標
            CALL l_sql.setAttribute("target",g_datalang[li_cnt].columns.trim())
            #表
            CALL l_sql.setAttribute("table",g_datalang[li_cnt].ref_table.trim())
            IF g_datalang[li_cnt].ref_table.trim() IS NULL THEN
               DISPLAY cl_getmsg_parm("adz-01120",g_lang,g_datalang[li_cnt].ref_table.trim())
              #DISPLAY "ERROR: 多語言欄位",g_datalang[li_cnt].ref_table.trim(),"規格的 '多語言Table' 資料未設定!"
            END IF
            #對應欄位
            #塞之前做一些處理(by site)
            IF g_table_main_pk.getIndexOf('site',4) = 0 THEN
               LET ls_append_pk = g_datalang[li_cnt].map_col.trim()
               IF ls_append_pk.getIndexOf('site',4) > 0 THEN
                  #重組並且濾除site
                  LET lst_token = base.StringTokenizer.create(ls_append_pk, ',')
                  LET ls_append_pk = ""
                  WHILE lst_token.hasMoreTokens()
                     LET ls_token = lst_token.nextToken()
                     IF ls_token.getIndexOf('site',4) = 0 THEN
                        LET ls_append_pk = ls_append_pk,ls_token,','
                     END IF
                  END WHILE
                  LET ls_append_pk = ls_append_pk.subString(1,ls_append_pk.getLength()-1)
               END IF
               CALL l_sql.setAttribute("fk",ls_append_pk)
               DISPLAY "fk:",ls_append_pk
            ELSE
               CALL l_sql.setAttribute("fk",g_datalang[li_cnt].map_col.trim())
            END IF
            
            #CALL l_sql.setAttribute("fk",g_datalang[li_cnt].map_col.trim())
            IF g_datalang[li_cnt].map_col.trim() IS NULL THEN
               DISPLAY "ERROR: 多語言欄位",g_datalang[li_cnt].ref_table.trim(),"規格的 'FK' 資料解析為空!"
               DISPLAY "       請檢查規格，或表格(",g_datalang[li_cnt].ref_table.trim(),") 在r.t的設定"
            END IF
            
            
            #PK
            #塞之前做一些處理(by site)
            IF g_table_main_pk.getIndexOf('site',4) = 0 THEN
               LET ls_append_pk = g_datalang[li_cnt].ref_fk.trim()
               IF ls_append_pk.getIndexOf('site',4) > 0 THEN
                  #重組並且濾除site
                  LET lst_token = base.StringTokenizer.create(ls_append_pk, ',')
                  LET ls_append_pk = ""
                  WHILE lst_token.hasMoreTokens()
                     LET ls_token = lst_token.nextToken()
                     IF ls_token.getIndexOf('site',4) = 0 THEN
                        LET ls_append_pk = ls_append_pk,ls_token,','
                     END IF
                  END WHILE
                  LET ls_append_pk = ls_append_pk.subString(1,ls_append_pk.getLength()-1)
               END IF
               CALL l_sql.setAttribute("pk",ls_append_pk)
               DISPLAY "pk:",ls_append_pk
            ELSE
               CALL l_sql.setAttribute("pk",g_datalang[li_cnt].ref_fk.trim())
            END IF
            
            #CALL l_sql.setAttribute("pk",g_datalang[li_cnt].ref_fk.trim())
            IF g_datalang[li_cnt].ref_fk.trim() IS NULL THEN
               DISPLAY "ERROR: 多語言欄位",g_datalang[li_cnt].ref_fk.trim(),"規格的 'PK' 資料解析為空!"
               DISPLAY "       請檢查規格，或表格(",g_datalang[li_cnt].ref_table.trim(),") 在r.t的設定"
            END IF
            #欄位
            CALL l_sql.setAttribute("field",g_datalang[li_cnt].ref_rtn.trim())
            IF g_datalang[li_cnt].ref_rtn.trim() IS NULL THEN
               DISPLAY cl_getmsg_parm("adz-01121",g_lang,g_datalang[li_cnt].ref_table.trim())
              #DISPLAY "ERROR: 多語言欄位",g_datalang[li_cnt].ref_table.trim(),"規格的 '回傳多語言欄位' 資料未設定!"
            END IF
         END FOR
      END IF

      #單頭其他表
      FOR li_cnt = 1 to g_table.getLength()
         IF g_table[li_cnt].table_type = "1" AND g_table[li_cnt].table_id <> g_table_main THEN
            LET l_head = p_dataset.createChild("head")

            #針對SITE特別處理(ka0132,2016-05-18)
            LET g_table[li_cnt].table_pk = sadzp030_tab_gen_append_pk(g_table[li_cnt].table_pk)

            CALL l_head.setAttribute("id",g_table[li_cnt].table_id)
            CALL l_head.setAttribute("pk",g_table[li_cnt].table_pk)
            
            #建置 sql -> forupd_sql
            LET l_sql = l_head.createChild("sql")
            CALL l_sql.setAttribute("id","forupd_sql")

            #類別
            CALL l_sql.setAttribute("type","sql")
            #指令
            CALL l_sql.setAttribute("query",sadzp030_tab_gen_sql("forupd_sql",g_table[li_cnt].table_id))
         END IF 
      END FOR
   END IF #判斷是否需要建置dataset -> head


   #建置 dataset -> body 單身用表  i02/i03/i04/i07/i08/i09/i12/t01/t02/q01/q02/q03
   IF g_prog_type = "i02" OR g_prog_type = "i03" OR g_prog_type = "i04" OR
      g_prog_type = "i07" OR g_prog_type = "i08" OR g_prog_type = "i09" OR
      g_prog_type = "i12" OR g_prog_type = "t01" OR g_prog_type = "t02" OR
      g_prog_type = "q01" OR g_prog_type = "q02" OR g_prog_type = "q03" OR
    ( g_prog_type = "q04" AND (FGL_GETENV("ZONE") = "t10dev" AND g_prog_id <> "aloq005") ) OR
      g_prog_type = "c02a" OR g_prog_type = "c02b" OR
      g_prog_type = "c03a" OR g_prog_type = "c03b" OR
      g_prog_type = "c04a" THEN

      FOR li_cnt = 1 TO g_detail_id.getLength()
         #dataset的輸出以table為主, 所以, 若該detail_id所屬表格已經出現過,就略過
         IF li_cnt >= 2 THEN
            #檢查連動旗標,若這張是N,停止旗標取消,如果是Y,則檢查之前是否已經處理過Y
            IF g_detail_id[li_cnt].connect = "N" THEN
               LET li_stop = FALSE
            ELSE
               LET li_stop = FALSE
               FOR li_abc = 1 TO li_cnt -1
                  #此處先檢查detail_id
                  IF g_detail_id[li_abc].detail_tab = g_detail_id[li_cnt].detail_tab AND
                     g_detail_id[li_abc].connect = "Y" THEN
                     LET li_stop = TRUE
                  END IF       
               END FOR
            END IF

            IF li_stop THEN
               #detail_append處理
               FOR li_cnt1 = 1 TO g_detail_op.getLength()
                  IF g_detail_op[li_cnt1].record_id = g_detail_id[li_cnt].detail_id THEN
                     LET l_sql = l_body.createChild("sql")
                     CALL l_sql.setAttribute("id","detail_append")
                     CALL l_sql.setAttribute("type","single")
                     CALL l_sql.setAttribute("table",g_detail_op[li_cnt1].table_id CLIPPED)
                     #160622
                     #CALL l_sql.setAttribute("pk","")
                     LET ls_append_pk = sadzp030_tab_relation_pk(g_detail_op[li_cnt1].table_id CLIPPED)
                     IF ls_append_pk.getIndexOf('site',1) > 0 THEN
                        #檢查上層單身是否有用到site
                        LET ls_detail_pk = ls_pk,',',ls_fk,',',g_table_main_pk
                        IF ls_detail_pk.getIndexOf('site',1) > 0 THEN
                           CALL l_sql.setAttribute("pk",ls_append_pk) 
                        ELSE
                           #重組並且濾除site
                           LET lst_token = base.StringTokenizer.create(ls_append_pk, ',')
                           LET ls_append_pk = ""
                           WHILE lst_token.hasMoreTokens()
                              LET ls_token = lst_token.nextToken()
                              IF ls_token.getIndexOf('site',4) = 0 THEN
                                 LET ls_append_pk = ls_append_pk,ls_token,','
                              END IF
                           END WHILE
                           LET ls_append_pk = ls_append_pk.subString(1,ls_append_pk.getLength()-1)
                        END IF
                        CALL l_sql.setAttribute("pk",ls_append_pk) 
                     ELSE
                        CALL l_sql.setAttribute("pk",ls_append_pk) 
                     END IF
                     
                     #濾除重複
                     LET lst_token = base.StringTokenizer.create(g_detail_op[li_cnt1].cols, ',')
                     WHILE lst_token.hasMoreTokens()
                        LET ls_token = lst_token.nextToken()
                        IF ls_field.getIndexOf(ls_token,1) = 0 THEN
                           LET ls_field = ls_field, ls_token, ','
                        END IF
                     END WHILE
            
                     #CALL l_sql.setAttribute("field",g_detail_op[li_cnt1].cols.subString(1,g_detail_op[li_cnt1].cols.getLength()-1))
                     CALL l_sql.setAttribute("field",ls_field.subString(1,ls_field.getLength()-1))
                     #160622
                  END IF
               END FOR
               CONTINUE FOR
            END IF
         END IF

         #創建
         LET l_body = p_dataset.createChild("body")
         CALL l_body.setAttribute("id",g_detail_id[li_cnt].detail_tab)

         #以表格為主,查看有多少page跟上
         IF g_detail_id[li_cnt].connect = "N" THEN
            CALL l_body.setAttribute("page",li_cnt)
         ELSE
            CALL l_body.setAttribute("page",sadzp030_tab_page_tab(g_detail_id[li_cnt].detail_tab))
         END IF

         CALL l_body.setAttribute("linked","")

         #處理 pk,fk
         CASE
            WHEN g_prog_type = "i02" OR
                 g_prog_type = "c02a" OR g_prog_type = "c02b"  #單檔多欄
               CALL l_body.setAttribute("pk",g_table_main_pk)

           #WHEN g_prog_type = "i03"  #樹狀假雙檔
            WHEN g_prog_type = "i04" OR g_prog_type = "i05" OR g_prog_type = "i08"

               IF lb_stus THEN 
                  #common="Y" 共用欄位 預設已經是Y
                  IF g_prog_type = "i04"  THEN 
                     CALL l_body.setAttribute("lid",g_dzff.id)
                     CALL sadzp030_tab_relation_i04_body_pid_type(g_detail_id[li_cnt].detail_tab,g_dzff.*) RETURNING g_dzff.*
                     CALL l_body.setAttribute("pid",g_dzff.pid) 
                     CALL l_body.setAttribute("type",g_dzff.type)
                  END IF
                  CALL sadzp030_tab_gen_fk_i04_i05_i08(g_detail_id[li_cnt].detail_tab) RETURNING ls_pk, ls_fk      
                  CALL l_body.setAttribute("pk",ls_pk)
                  CALL l_body.setAttribute("fk",ls_fk)
               ELSE
                  DISPLAY cl_getmsg("adz-01119",g_lang)
                 #DISPLAY "ERROR: 樹狀資料不正確:"
               END IF

            #假雙檔 fk="gztb001"(單頭的欄位) pk="gztb002"(表PK-FK) 
            WHEN g_prog_type = "i07" OR g_prog_type = "i12"
               IF g_detail_id[li_cnt].detail_tab = g_table_main THEN
                  CALL sadzp030_tab_gen_fk_i07(g_detail_id[li_cnt].detail_tab) RETURNING ls_pk, ls_fk
               ELSE
                  DISPLAY cl_getmsg("adz-01122",g_lang)
                 #DISPLAY "INFO: 產生假雙檔加單身設定段落"
                  #假雙檔加單身時使用
                  CALL sadzp030_tab_gen_fk_t01(g_detail_id[li_cnt].detail_tab,g_table_main) RETURNING ls_pk, ls_fk, li_def_site
                  CALL l_body.setAttribute("detail","Y")
                  CALL l_body.setAttribute("master",g_table_main)
               END IF

               CALL l_body.setAttribute("pk",ls_pk)
               CALL l_body.setAttribute("fk",ls_fk)

            #雙檔樣板 fk="gztb001"(與單頭相關的欄位) pk="gztb002"(表PK-FK) 
            WHEN g_prog_type = "i09" OR g_prog_type = "t01" OR
                 g_prog_type = "c03a" OR g_prog_type = "c03b"

               #t01 符合多檔結構 時要在第三階表加上 detail="Y"
               IF g_prog_type = "t01" OR g_prog_type = "i09" OR g_prog_type = "c03a" THEN
                  IF g_detail_id[li_cnt].layer3_up IS NOT NULL THEN
                     CALL l_body.setAttribute("detail","Y")
                     CALL l_body.setAttribute("master",g_detail_id[li_cnt].layer3_up CLIPPED)
                     CALL sadzp030_tab_gen_fk_t01(g_detail_id[li_cnt].detail_tab,g_detail_id[li_cnt].layer3_up) RETURNING ls_pk, ls_fk, li_def_site
                     IF li_def_site THEN
                        CALL l_body.setAttribute("default_site","Y")
                     ELSE
                        CALL l_body.setAttribute("default_site","N")
                     END IF
                  ELSE
                     CALL sadzp030_tab_gen_fk_t01(g_detail_id[li_cnt].detail_tab,g_table_main) RETURNING ls_pk, ls_fk, li_def_site
                  END IF
               ELSE
                  CALL sadzp030_tab_gen_fk_t01(g_detail_id[li_cnt].detail_tab,g_table_main) RETURNING ls_pk, ls_fk, li_def_site
               END IF

               CALL l_body.setAttribute("pk",ls_pk)
               CALL l_body.setAttribute("fk",ls_fk)

               IF g_prog_type = "t01" THEN
                  CALL sadzp030_tab_gen_3layer2show(g_detail_id[li_cnt].detail_tab) RETURNING ls_32wc
                  IF ls_32wc IS NOT NULL THEN
                     CALL l_body.setAttribute("32wc",ls_32wc)
                  END IF
               END IF

            #t02像單檔多欄的雙檔
            WHEN g_prog_type = "t02" OR
                 g_prog_type = "q01" OR g_prog_type = "q02" OR g_prog_type = "q03" OR 
               ( g_prog_type = "q04" AND (FGL_GETENV("ZONE") = "t10dev" AND g_prog_id <> "aloq005") ) OR
                 g_prog_type = "c04a"
               CALL sadzp030_tab_gen_fk_t01(g_detail_id[li_cnt].detail_tab,g_table_main) RETURNING ls_pk, ls_fk, li_def_site
               IF g_t100debug >= 6 THEN
                  DISPLAY "INFO: ",g_detail_id[li_cnt].detail_tab,'---',ls_pk
               END IF

               IF g_detail_id[li_cnt].detail_tab = g_table_main THEN
                  CASE
                     WHEN ls_pk IS NULL
                        LET ls_pk = ls_fk
                     WHEN ls_fk IS NULL
                     OTHERWISE
                        LET ls_pk = ls_fk,",",ls_pk
                  END CASE
                  CALL l_body.setAttribute("pk",ls_pk)
               ELSE
                  CALL l_body.setAttribute("pk",ls_pk)
                  CALL l_body.setAttribute("fk",ls_fk)
               END IF
               #t02時要在單身表加上 detail="Y"
               IF g_detail_id[li_cnt].detail_tab = g_table_main THEN
               ELSE
                  CALL l_body.setAttribute("detail","Y")
               END IF

            OTHERWISE
         END CASE

         #下方進行forupd_sql_detail的填入,q類免填
         IF NOT g_prog_type.subString(1,1) = "q" THEN

            #建置 sql -> forupd_sql
            LET l_sql = l_body.createChild("sql")
            CALL l_sql.setAttribute("id","forupd_sql_detail")

            #指令
            CALL l_sql.setAttribute("query",sadzp030_tab_gen_sql("forupd_sql_detail",g_detail_id[li_cnt].detail_tab))

            #detail_append處理
            FOR li_cnt1 = 1 TO g_detail_op.getLength()
                IF g_detail_op[li_cnt1].record_id = g_detail_id[li_cnt].detail_id THEN
                  #也檢查一下是否是多語言表內的項目,是的話就不輸出
                  LET lb_pass = TRUE
                  FOR li_cnt2 = 1 TO g_mdatalang.getLength()
                     IF g_detail_op[li_cnt1].table_id = g_mdatalang[li_cnt2].ref_table THEN
                        LET lb_pass = FALSE
                     END IF
                  END FOR

                  #檢查通過的話就可以加這一行
                  IF lb_pass THEN
                     LET l_sql = l_body.createChild("sql")
                     CALL l_sql.setAttribute("id","detail_append")
                     CALL l_sql.setAttribute("type","single")
                     CALL l_sql.setAttribute("table",g_detail_op[li_cnt1].table_id CLIPPED)
                     LET ls_append_pk = sadzp030_tab_relation_pk(g_detail_op[li_cnt1].table_id CLIPPED)
                     #檢查pk中是否有site
                     IF ls_append_pk.getIndexOf('site',1) > 0 THEN
                        #檢查上層單身是否有用到site
                        LET ls_detail_pk = ls_pk,',',ls_fk,',',g_table_main_pk
                        IF ls_detail_pk.getIndexOf('site',1) > 0 THEN
                           CALL l_sql.setAttribute("pk",ls_append_pk) 
                        ELSE
                           #重組並且濾除site
                           LET lst_token = base.StringTokenizer.create(ls_append_pk, ',')
                           LET ls_append_pk = ""
                           WHILE lst_token.hasMoreTokens()
                              LET ls_token = lst_token.nextToken()
                              IF ls_token.getIndexOf('site',4) = 0 THEN
                                 LET ls_append_pk = ls_append_pk,ls_token,','
                              END IF
                           END WHILE
                           LET ls_append_pk = ls_append_pk.subString(1,ls_append_pk.getLength()-1)
                        END IF
                        CALL l_sql.setAttribute("pk",ls_append_pk) 
                     ELSE
                        CALL l_sql.setAttribute("pk",ls_append_pk) 
                     END IF
                     CALL l_sql.setAttribute("field",g_detail_op[li_cnt1].cols.subString(1,g_detail_op[li_cnt1].cols.getLength()-1))
                  END IF
               END IF
            END FOR
         END IF  #q類免做forupd_sql與detail_append,因為只有查詢

         #建置 sql -> b_fill_sql
         IF g_prog_type = "i03" THEN #樹狀假雙檔 不需要使用 b_fill_sql,使用其他的組合方式
         ELSE
            LET l_sql = l_body.createChild("sql")
            CALL l_sql.setAttribute("id","b_fill_sql")

            #指令
            CALL l_sql.setAttribute("query",sadzp030_tab_gen_sql("b_fill_sql",g_detail_id[li_cnt].detail_tab))

         END IF

         IF g_mdatalang.getLength() > 0 THEN
            #找出本page之後,而且table相同的
            CALL la_page.clear()

            FOR li_abc = li_cnt TO g_detail_id.getLength()
               IF g_detail_id[li_abc].detail_tab = g_detail_id[li_cnt].detail_tab THEN
                  LET la_page[la_page.getLength()+1] = li_abc
               END IF 
            END FOR

            IF la_page.getLength() >= 1 THEN
               FOR li_cnt1 = 1 TO g_mdatalang.getLength()
                  FOR li_abc = 1 TO la_page.getLength()
                     IF g_mdatalang[li_cnt1].page = la_page[li_abc] AND
                        g_mdatalang[li_cnt1].used IS NULL THEN
                        #建置 sql -> append
                        LET l_sql = l_body.createChild("sql")
                        CALL l_sql.setAttribute("id","detail_append")
                        #型態
                        CALL l_sql.setAttribute("type","lang")
                        #目標
                        CALL l_sql.setAttribute("target",g_mdatalang[li_cnt1].columns.trim())
                        #表
                        CALL l_sql.setAttribute("table",g_mdatalang[li_cnt1].ref_table.trim())
                        IF cl_null(g_mdatalang[li_cnt1].ref_table.trim()) THEN
                           DISPLAY "注意: 欄位(",g_mdatalang[li_cnt1].columns.trim(),")多語言設定,表資料未填"
                        END IF
                        #對應欄位
                        CALL l_sql.setAttribute("fk",g_mdatalang[li_cnt1].map_col.trim())
                        IF cl_null(g_mdatalang[li_cnt1].map_col.trim()) THEN
                           DISPLAY "注意: 欄位(",g_mdatalang[li_cnt1].columns.trim(),")多語言設定,對應欄位未填"
                        END IF
                        #PK
                        CALL l_sql.setAttribute("pk",g_mdatalang[li_cnt1].ref_fk.trim())
                        IF cl_null(g_mdatalang[li_cnt1].ref_fk.trim()) THEN
                           DISPLAY "注意: 欄位(",g_mdatalang[li_cnt1].columns.trim(),")多語言設定,PK欄位未填"
                        END IF
                        #欄位
                        CALL l_sql.setAttribute("field",g_mdatalang[li_cnt1].ref_rtn.trim())
                        IF cl_null(g_mdatalang[li_cnt1].ref_rtn.trim()) THEN
                           DISPLAY "注意: 欄位(",g_mdatalang[li_cnt1].columns.trim(),")多語言設定,回傳欄位未填"
                        END IF
                        LET g_mdatalang[li_cnt1].used = "Y"
                     END IF
                  END FOR
               END FOR
            END IF

         END IF
      END FOR
   END IF

END FUNCTION

PRIVATE FUNCTION sadzp030_tab_gen_3layer2show(lc_detail_tab)
   DEFINE lc_detail_tab STRING
   DEFINE lc_dzag013    LIKE dzag_t.dzag013
   DEFINE lc_dzag014    LIKE dzag_t.dzag014
   DEFINE lc_dzag015    LIKE dzag_t.dzag015
   DEFINE ls_sql        STRING
   DEFINE ls_wc         STRING
   DEFINE ls_uppfk      STRING
   DEFINE ls_tmp        STRING
   DEFINE l_token4    base.StringTokenizer
   DEFINE l_token5    base.StringTokenizer
   DEFINE l_tokenf    base.StringTokenizer
   DEFINE la_dzag     DYNAMIC ARRAY OF RECORD
             dzag014  LIKE dzag_t.dzag014,
             dzag015  LIKE dzag_t.dzag015,
             checked  LIKE type_t.chr1
                      END RECORD
   DEFINE li_cnt      LIKE type_t.num5

   LET ls_sql = "SELECT dzag013,dzag014,dzag015 FROM dzaa_t,dzag_t",
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
                  " AND dzag002 = '",lc_detail_tab.trim(),"' ",
                  " AND dzag003 = dzaa004 ",         #識別碼版次
                  " AND dzag006 = dzaa006 ",         #使用標示
#                 " AND dzag011 = dzaa010 ",         #客戶編號
                  " AND dzagstus = 'Y' "

   PREPARE sadzp030_tab_gen_32pre FROM ls_sql
   EXECUTE sadzp030_tab_gen_32pre INTO lc_dzag013,lc_dzag014,lc_dzag015
   FREE sadzp030_tab_gen_32pre 

   #2016/11/22 避免該欄位為空白時判斷錯誤
   LET lc_dzag013 = lc_dzag013 CLIPPED
   LET lc_dzag014 = lc_dzag014 CLIPPED
   LET lc_dzag015 = lc_dzag015 CLIPPED

   IF lc_dzag013 IS NOT NULL AND lc_dzag014 IS NOT NULL AND lc_dzag015 IS NOT NULL THEN
      CALL sadzp030_tab_relation_fk(lc_dzag013,g_table_main,2) RETURNING ls_uppfk,ls_tmp

      LET l_token4 = base.StringTokenizer.create(lc_dzag014 CLIPPED, ",")
      LET l_token5 = base.StringTokenizer.create(lc_dzag015 CLIPPED, ",")
      LET li_cnt = 1
      WHILE l_token4.hasMoreTokens() AND l_token5.hasMoreTokens()
         LET la_dzag[li_cnt].dzag014 = l_token4.nextToken()
         LET la_dzag[li_cnt].dzag015 = l_token5.nextToken()
         LET la_dzag[li_cnt].checked = "N"
         LET li_cnt = li_cnt + 1
      END WHILE

      LET l_tokenf = base.StringTokenizer.create(ls_uppfk.trim(), ",")
      WHILE l_tokenf.hasMoreTokens()
         LET ls_tmp = l_tokenf.nextToken()
         FOR li_cnt = 1 TO la_dzag.getLength()
            IF ls_tmp = la_dzag[li_cnt].dzag014 THEN
               LET la_dzag[li_cnt].checked = "Y"
            END IF
         END FOR 
      END WHILE

      FOR li_cnt = 1 TO la_dzag.getLength()
         IF la_dzag[li_cnt].checked = "N" THEN
         #  LET ls_wc = ls_wc,la_dzag[li_cnt].dzag014 CLIPPED,"=",la_dzag[li_cnt].dzag015 CLIPPED," AND"    #171023-00044 mark
            LET ls_wc = ls_wc,la_dzag[li_cnt].dzag014 CLIPPED,"=",la_dzag[li_cnt].dzag015 CLIPPED," AND "   #171023-00044 modify
         END IF
      END FOR
      IF ls_wc IS NOT NULL THEN
         LET ls_wc = ls_wc.subString(1,ls_wc.getLength()-4)
      END IF

#display lc_detail_tab," 對上隱藏上階:",lc_dzag013
#display lc_detail_tab," 本表原有FK:",ls_uppfk
#display lc_detail_tab," 設定上階FK:",lc_dzag014
#display lc_detail_tab," 設定本體FK:",lc_dzag015
#display "最終wc=",ls_wc
   END IF

   RETURN ls_wc

END FUNCTION

#產出i07 dataset內需要的 fk, 及濾除fk後剩餘的 pk
PRIVATE FUNCTION sadzp030_tab_gen_fk_i07(ps_tableid) 
   DEFINE ls_pk         STRING
   DEFINE ls_fk         STRING 
   DEFINE ls_temp       STRING
   DEFINE li_pos        LIKE type_t.num5
   DEFINE li_cnt        LIKE type_t.num5
   DEFINE li_cnt1       LIKE type_t.num5
   DEFINE li_check      LIKE type_t.num5
   DEFINE li_exist      LIKE type_t.num5  #site欄位存在
   DEFINE l_token1      base.StringTokenizer
   DEFINE l_token2      base.StringTokenizer
   DEFINE ls_next1      STRING
   DEFINE ls_next2      STRING
   DEFINE ps_tableid    STRING
   DEFINE la_col_set    DYNAMIC ARRAY OF RECORD
             col_id     STRING
                        END RECORD

   #先將畫面master部分拆入array
   LET l_token1 = base.StringTokenizer.create(g_col_set[1].fields, ",")
   CALL la_col_set.clear()
   LET li_pos = 1
   WHILE l_token1.hasMoreTokens()
      LET ls_next1 = l_token1.nextToken()
      LET la_col_set[li_pos].col_id = ls_next1
      LET li_pos = li_pos + 1
   END WHILE

   #將PK依照畫面上所在位置,切分FK(在單頭的) 與PK(在單身的)
   LET l_token2 = base.StringTokenizer.create(sadzp030_tab_get_relation(ps_tableid,"pk",""), ",")
   WHILE l_token2.hasMoreTokens()
      LET ls_next2 = l_token2.nextToken()
      LET li_check = FALSE
      FOR li_pos = 1 TO la_col_set.getLength()
         IF la_col_set[li_pos].col_id = ls_next2 THEN
            LET li_check = TRUE
            EXIT FOR
         END IF 
      END FOR
      IF li_check THEN  #發現此欄位, 算FK(與單頭相關), 沒發現 算PK
         LET ls_fk = ls_fk,",",ls_next2
      ELSE
         #如果是site欄位,檢查一下單身畫面是否有出現,沒有的話要濾除
         LET ls_temp = ps_tableid.subString(1,ps_tableid.getIndexOf("_t",1)-1),"site"
         IF ls_next2 = ls_temp THEN
            LET li_exist = FALSE  #假設欄位不存在

            #從g_detail_id繞 table id
            FOR li_cnt = 1 TO g_detail_id.getLength()
               #找到相等的 table id 後去找 g_col_set 欄位清單
               IF g_detail_id[li_cnt].detail_tab = ps_tableid THEN
                  #找欄位清單
                  FOR li_cnt1 = 1 TO g_col_set.getLength()
                     #如果s_detail名稱相同
                     IF g_col_set[li_cnt1].pos = g_detail_id[li_cnt].detail_id THEN
                        #總算可以看是否存在
                        IF g_col_set[li_cnt1].fields.getIndexOf(ls_temp,1) THEN
                           LET li_exist = TRUE #site欄位存在
                        END IF
                     END IF
                  END FOR
               END IF
            END FOR

            #site欄位存在畫面上時,當作一般欄位處理
            IF li_exist THEN 
               LET ls_pk = ls_pk,",",ls_next2
            END IF

         #不是site欄位不用檢查
         ELSE
            LET ls_pk = ls_pk,",",ls_next2
         END IF
      END IF
   END WHILE

   IF ls_fk.getLength() >= 2 THEN
      LET ls_fk = ls_fk.subString(2,ls_fk.getLength())
   END IF
   IF ls_pk.getLength() >= 2 THEN 
      LET ls_pk = ls_pk.subString(2,ls_pk.getLength())
   END IF

   RETURN ls_pk,ls_fk
END FUNCTION

#產出t01 dataset內需要的 fk, 及濾除fk後剩餘的 pk
PRIVATE FUNCTION sadzp030_tab_gen_fk_t01(ps_tableid,ps_uptable) 

   DEFINE ps_tableid    STRING  #關注的table
   DEFINE ps_uptable    STRING  #對應要找的關聯上階table
   DEFINE ls_next1      STRING
   DEFINE ls_next2      STRING
   DEFINE li_flag       LIKE type_t.num5
   DEFINE li_cnt        LIKE type_t.num5
   DEFINE li_cnt1       LIKE type_t.num5
   DEFINE l_token1      base.StringTokenizer
   DEFINE l_token2      base.StringTokenizer
   DEFINE ls_getpk      STRING
   DEFINE ls_pk         STRING
   DEFINE ls_fk         STRING 
   DEFINE ls_str        STRING
   DEFINE ls_sitecol    STRING  #組出此表應有的site欄位名稱
   DEFINE ls_tmp1,ls_tmp2 STRING
   DEFINE li_def_site   LIKE type_t.num5

   LET ls_sitecol = ps_tableid.subString(1,ps_tableid.getIndexOf("_t",1)-1),"site"

   CALL sadzp030_tab_relation_fk(ps_tableid,ps_uptable,3) RETURNING ls_tmp1,ls_tmp2
   #pk,逗號分隔一個個截開
   LET ls_getpk = sadzp030_tab_get_relation(ps_tableid,"pk","")

   #初始預設site有出現
   IF ls_getpk.getIndexOf(ls_sitecol,1) THEN
      LET li_def_site = TRUE
   ELSE
      LET li_def_site = FALSE
   END IF

   LET l_token1 = base.StringTokenizer.create(ls_getpk, ",")
   IF g_t100debug >= 6 THEN
      DISPLAY "INFO: PK=",sadzp030_tab_get_relation(ps_tableid,"pk","")
   END IF

   #找出與上階關係(FK)
   IF ps_tableid.trim() = ps_uptable.trim() THEN  #如果表名稱相等,即為自己
      LET ls_str = ""
   ELSE
#問題發生點
      LET ls_str = sadzp030_tab_get_relation(ps_tableid,"fk",ps_uptable)
   END IF

   WHILE l_token1.hasMoreTokens()
      LET ls_next1 = l_token1.nextToken()

      #site欄位要做特別處理-檢查是否為site
      IF ls_next1 = ls_sitecol THEN
         #如果抓到的是site欄位,而主表PK也有site,則當作fk
         IF g_table_main_pk.getIndexOf(g_table_main.subString(1,g_table_main.getIndexOf("_t",1)-1)||"site",1) THEN
            LET li_flag = TRUE
         ELSE
            LET li_flag = FALSE
         END IF

      #非site欄位的處理
      ELSE
         LET li_flag = FALSE

         #取出 body的 fk (A,C,E=B,D,F) 逗號分隔截開
         LET l_token2 = base.StringTokenizer.create(ls_str, ",")

         WHILE l_token2.hasMoreTokens()
            LET ls_next2 = l_token2.nextToken()

            #PK與fk截出來的項目若有相等, 則 等於ls_fk(TRUE) 否則為FALSE(ls_pk)
            IF ls_next1 = ls_next2 THEN
               LET li_flag = TRUE
            ELSE
               #如果不存在的話,做一個提醒
               IF ls_getpk.getIndexOf(ls_next2,1) = 0 AND NOT ls_next2.getIndexOf("ent",1) THEN
                  DISPLAY "注意: 檢查到表格",ps_tableid,"的PK沒有",ls_next2,"欄位,會造成FK欄位缺少,程式錯誤!!"
               END IF
            END IF
         END WHILE
      END IF

      #取出濾除 ent 企業編號的部分
      IF NOT ls_next1.subString(ls_next1.getLength()-2,ls_next1.getLength()) = "ent" THEN
         #相等的是fk, 不相等的是pk
         IF li_flag THEN 
            LET ls_fk = ls_fk,",",ls_next1
         ELSE
            #如果是site欄位,檢查一下單身畫面是否有出現,沒有的話要濾除
            IF ls_next1 = ls_sitecol THEN
               #從g_detail_id繞 table id
               FOR li_cnt = 1 TO g_detail_id.getLength()
                  #找到相等的 table id 後去找 g_col_set 欄位清單
                  IF g_detail_id[li_cnt].detail_tab = ps_tableid THEN

                     #找欄位清單
                     FOR li_cnt1 = 1 TO g_col_set.getLength()
                        #如果s_detail名稱相同
                        IF g_col_set[li_cnt1].pos = g_detail_id[li_cnt].detail_id THEN
                           #總算可以看是否存在-再就組進去, 當普通PK
                           IF g_col_set[li_cnt1].fields.getIndexOf(ls_sitecol,1) THEN
                              #確認site沒進去過
                              IF ls_pk.getIndexOf(ls_next1,1) = 0 THEN
                                 LET ls_pk = ls_pk,",",ls_next1
                              END IF
                           END IF
                        END IF
                     END FOR

                  END IF
               END FOR
            ELSE
               LET ls_pk = ls_pk,",",ls_next1
            END IF
         END IF
      END IF
   END WHILE

   IF ls_fk.getLength() >= 2 THEN
      LET ls_fk = ls_fk.subString(2,ls_fk.getLength())
   END IF
   IF ls_pk.getLength() >= 2 THEN 
      LET ls_pk = ls_pk.subString(2,ls_pk.getLength())
   END IF

   RETURN ls_pk,ls_fk,li_def_site
END FUNCTION


#產出i04 i05 i08 dataset內需要的 fk, 及濾除fk後剩餘的 pk
PRIVATE FUNCTION sadzp030_tab_gen_fk_i04_i05_i08(ps_tableid) 

   DEFINE ls_next1      STRING
   DEFINE ls_next2      STRING
   DEFINE li_flag       LIKE type_t.num5
   DEFINE l_token1      base.StringTokenizer
   DEFINE l_token2      base.StringTokenizer
   DEFINE ls_pk         STRING
   DEFINE ls_fk         STRING 
   DEFINE ls_str        STRING
   DEFINE ps_tableid    STRING
   
   #取出tsd內 body的 pk 當作底,逗號分隔一個個截開
   LET l_token1 = base.StringTokenizer.create(sadzp030_tab_get_relation(ps_tableid,"pk",""), ",")

   #找出與 main table關係(FK)
   LET ls_str = sadzp030_tab_get_relation(ps_tableid,"fk",g_table_main)

   WHILE l_token1.hasMoreTokens()
      LET ls_next1 = l_token1.nextToken()
      LET li_flag = FALSE

      #取出 body的 fk (A,C,E=B,D,F) 逗號分隔截開
      LET l_token2 = base.StringTokenizer.create(ls_str, ",")

      WHILE l_token2.hasMoreTokens()
         LET ls_next2 = l_token2.nextToken()

         #濾除等號不屬於body表的那一側, 留下body欄位的 column name
#        LET ls_next2 = sadzp030_tab_clearfyfk(ls_next2,g_table.detail)
         #跟PK比若有相等, 則
         
         IF ls_next1 = ls_next2 THEN
            LET li_flag = TRUE
         END IF
      END WHILE

      #取出濾除 ent 企業編號的部分
      IF NOT ls_next1.subString(ls_next1.getLength()-2,ls_next1.getLength()) = "ent" THEN
         #相等的是fk, 不相等的是pk
         IF li_flag THEN 
            #當i04樣板 有type要比對FK 濾除掉跟type 同一個欄位
            IF g_prog_type = "i04" THEN 
               IF sadzp030_tab_gen_chk_type(ls_next1)  THEN 
                  CONTINUE WHILE 
               ELSE 
                  LET ls_fk = ls_fk,",",ls_next1 
               END IF
            ELSE 
               LET ls_fk = ls_fk,",",ls_next1 
            END IF 
         ELSE
            LET ls_pk = ls_pk,",",ls_next1
         END IF
      END IF
   END WHILE

   IF ls_fk.getLength() >= 2 THEN
      LET ls_fk = ls_fk.subString(2,ls_fk.getLength())
   END IF
   IF ls_pk.getLength() >= 2 THEN 
      LET ls_pk = ls_pk.subString(2,ls_pk.getLength())
   END IF
   RETURN ls_pk,ls_fk
END FUNCTION


#濾除資料多語言的輸入欄位 (只留下單頭table相關欄位,其餘都變成 '' )
PRIVATE FUNCTION sadzp030_tab_master_fixdlang_col(ls_cols,ls_tableid)

   DEFINE ls_cols    STRING
   DEFINE ls_tableid STRING
   DEFINE l_token1       base.StringTokenizer
   DEFINE ls_next1       STRING

   #取tableid的字首
   LET ls_tableid = ls_tableid.subString(1,ls_tableid.getIndexOf("_t",1)-1)
   #確認傳入的欄位清單,個別在TSD內是否已經有設定參考欄位資料
   LET l_token1 = base.StringTokenizer.create(ls_cols, ",")
   LET ls_cols = ""   
   WHILE l_token1.hasMoreTokens()
      LET ls_next1 = l_token1.nextToken()
      IF ls_next1.getIndexOf(ls_tableid,1) THEN
         LET ls_cols = ls_cols,ls_next1,","
      ELSE
         LET ls_cols = ls_cols,"'',"
      END IF
   END WHILE
   RETURN ls_cols

END FUNCTION


#產生 tab 內各處要用到的 SQL
PRIVATE FUNCTION sadzp030_tab_gen_sql(ps_sql_type,ps_tableid)

   DEFINE ps_sql_type     STRING
   DEFINE ls_sql          STRING
   DEFINE ls_temp         STRING
   DEFINE ls_master_col   STRING   #單頭欄位字串
   DEFINE ls_detail_col   STRING   #單身欄位字串
   DEFINE ps_tableid      STRING
   DEFINE li_cnt          LIKE type_t.num5

   CASE ps_sql_type 
      WHEN "forupd_sql"
         LET ls_master_col = sadzp030_tab_ma("mainlayout",4,ps_tableid,FALSE),",",
                             sadzp030_tab_ma("patch_grid",4,ps_tableid,FALSE)
         IF ls_master_col.subString(ls_master_col.getLength(),ls_master_col.getLength())="," THEN
            LET ls_master_col = ls_master_col.subString(1,ls_master_col.getLength()-1)
         END IF
         IF g_t100debug >= 6 THEN
            DISPLAY "INFO: ls_master_col:",ls_master_col
         END IF

         #資料多語言也會使用相同的走法,所以要做一個調整
         LET ls_master_col = sadzp030_tab_master_fixdlang_col(ls_master_col,ps_tableid)
         LET ls_master_col = ls_master_col.subString(1,ls_master_col.getLength()-1)
         #單頭其他表(非主表)提供只留下自己的功能
         IF ps_tableid <> g_table_main THEN
            LET ls_master_col = sadzp030_tab_master_not_main_fix(ls_master_col)
         END IF
         #校准後組合
         LET ls_sql = "SELECT ",ls_master_col,
                       " FROM ",ps_tableid,
                      " WHERE ",sadzp030_tab_entfix(ps_tableid),sadzp030_tab_gen_pk("forupd_sql",ps_tableid),
                        " FOR UPDATE"
         IF g_t100debug >= 6 THEN
            DISPLAY "INFO: 生成forupd_sql=",ls_sql
         END IF

      WHEN "cs_sql" 
         LET ls_master_col = sadzp030_tab_ma("mainlayout",4,ps_tableid,FALSE),",",
                             sadzp030_tab_ma("patch_grid",4,ps_tableid,FALSE)
         IF ls_master_col.subString(ls_master_col.getLength(),ls_master_col.getLength())="," THEN
            LET ls_master_col = ls_master_col.subString(1,ls_master_col.getLength()-1)
         END IF
         LET ls_sql = "SELECT ",ls_master_col,
                       " FROM ",ps_tableid
         IF g_t100debug >=6 THEN
            DISPLAY "INFO: 生成cs_sql=",ls_sql
         END IF

      WHEN "forupd_sql_detail"    #多頁簽的組合
         LET ls_detail_col = ""
         FOR li_cnt = 1 TO g_detail_id.getLength()
            IF g_detail_id[li_cnt].detail_tab = ps_tableid THEN

               LET ls_temp = sadzp030_tab_de(g_detail_id[li_cnt].detail_id,4,ps_tableid)
                                                      #傳入陣列                        取SQL   
               LET ls_detail_col = ls_detail_col,ls_temp,","
            END IF
         END FOR
         LET ls_detail_col = ls_detail_col.subString(1,ls_detail_col.getLength()-1)
         LET ls_sql = "SELECT ",ls_detail_col,
                       " FROM ",ps_tableid,
                      " WHERE ",sadzp030_tab_gen_pk("forupd_sql_detail",ps_tableid)," FOR UPDATE"
         IF g_t100debug >=6 THEN
            DISPLAY "INFO: 生成forupd_sql_detail=",ls_sql
         END IF

      WHEN "b_fill_sql" 
         LET ls_detail_col = ""
         FOR li_cnt = 1 TO g_detail_id.getLength()
            IF g_detail_id[li_cnt].detail_tab = ps_tableid THEN
                                                         #傳入陣列        取SQL   
               LET ls_temp = sadzp030_tab_de(g_detail_id[li_cnt].detail_id,4,ps_tableid)
               #Q類的選擇欄位必須排除串查的部分(因為串查的型態定義為string,然後FOREACH元素會在產生器處理完)
               IF g_prog_type.subString(1,1) = "q" THEN
                  CALL sadzp030_tab_exclude_q(ls_temp,li_cnt) RETURNING ls_temp
               END IF
               LET ls_detail_col = ls_detail_col,ls_temp,","
            END IF
         END FOR
         LET ls_detail_col = ls_detail_col.subString(1,ls_detail_col.getLength()-1)
         #將WHERE條件下放到函式內處理
         LET ls_sql = "SELECT ",ls_detail_col,
                       " FROM ",ps_tableid," WHERE ",sadzp030_tab_gen_pk("b_fill_sql",ps_tableid)
         IF g_t100debug >=6 THEN
            DISPLAY "INFO: 生成b_fill_sql=",ls_sql
         END IF
   END CASE
   RETURN ls_sql

END FUNCTION


#檢查資料庫, 回補ent(企業代碼)/site(營業據點代碼)
PRIVATE FUNCTION sadzp030_tab_entfix(ls_tabid)

   DEFINE ls_entfix    STRING
   DEFINE ls_tabid     STRING
   DEFINE ls_pk        STRING
   DEFINE lc_gztz001   LIKE gztz_t.gztz001
   DEFINE lc_gztz002   LIKE gztz_t.gztz002
   DEFINE li_cnt       LIKE type_t.num5
   DEFINE li_pos       LIKE type_t.num5
   DEFINE li_idx       LIKE type_t.num5
   DEFINE li_i07_chk   LIKE type_t.num5

   LET ls_tabid = ls_tabid.trim()
   LET ls_entfix = ""

   IF ls_tabid.getLength() < 1 THEN
      RETURN ls_entfix
   ELSE
      LET lc_gztz001 = ls_tabid
   END IF

   FOR li_pos = 1 TO 2
      CASE 
         WHEN li_pos = 1  #組出應該有的site欄位名稱
            #看一下g_table_main_pk,若這裡含有site表示畫面上有site,就不補了
            IF g_table_main_pk.getIndexOf("site",3) THEN
               CONTINUE FOR
            END IF
            LET lc_gztz002 = ls_tabid.subString(1,ls_tabid.getIndexOf("_t",1)-1),"site"

            #確認一下site是否為該表的PK欄位之一,若不是就不補了
            LET ls_pk = sadzp030_tab_relation_pk(ls_tabid)
            IF NOT ls_pk.getIndexOf(lc_gztz002,1) THEN
               CONTINUE FOR
            END IF
            
            #20160614
            #確認一下site是否出現再單身(假雙檔only)
            LET li_i07_chk = 0
            FOR li_idx = 1 TO g_col_set.getLength()
               IF g_col_set[li_idx].fields.getIndexOf(lc_gztz002,1) AND 
                  (g_prog_type = 'i07' OR g_prog_type = 'i12') THEN
                  LET li_i07_chk = 1
               END IF
            END FOR
            IF li_i07_chk = 1 AND (g_prog_type = 'i07' OR g_prog_type = 'i12') THEN
               CONTINUE FOR
            END IF

         WHEN li_pos = 2  #組出應該有的ent欄位名稱
            LET lc_gztz002 = ls_tabid.subString(1,ls_tabid.getIndexOf("_t",1)-1),"ent"
      END CASE

      #確認是否存在這個欄位
      SELECT COUNT(1) INTO li_cnt FROM gztz_t
       WHERE gztz001 = lc_gztz001 AND gztz002 = lc_gztz002 
      IF li_cnt < 1 THEN
         LET ls_entfix = ls_entfix
      ELSE
         LET ls_entfix = lc_gztz002,"= ? AND ",ls_entfix
      END IF
   END FOR

   RETURN ls_entfix
END FUNCTION


#產生 tab 內各處要用到的 pk 字串
PRIVATE FUNCTION sadzp030_tab_gen_pk(ls_type,ps_tableid) 

   DEFINE ls_type       STRING
   DEFINE ls_pk         STRING
   DEFINE ls_fk         STRING
   DEFINE ls_next       STRING
   DEFINE ps_tableid    STRING
   DEFINE l_token1      base.StringTokenizer
   DEFINE li_i07_fk     LIKE type_t.num5
   DEFINE li_cnt        LIKE type_t.num5
   DEFINE ls_temp       STRING
   DEFINE li_def_site   LIKE type_t.num5

   CASE
      WHEN ls_type = "forupd_sql"
         IF ps_tableid = g_table_main THEN
            LET l_token1 = base.StringTokenizer.create(g_table_main_pk, ",")
         ELSE
            FOR li_cnt = 1 TO g_table.getLength()
               IF g_table[li_cnt].table_id = ps_tableid THEN
                  LET l_token1 = base.StringTokenizer.create(g_table[li_cnt].table_pk, ",")
                  EXIT FOR
               END IF
            END FOR
         END IF
         WHILE l_token1.hasMoreTokens()
            LET ls_next = l_token1.nextToken()
            LET ls_pk = ls_pk,ls_next,"=? AND "
         END WHILE
         LET ls_pk = ls_pk.subString(1,ls_pk.getLength()-5)

      WHEN ls_type = "forupd_sql_detail" 

         LET li_i07_fk = FALSE
         #此處應該針對i12/i07 進行特別處理,先放單頭,再放單身
         CASE
            WHEN g_prog_type = "i12" OR g_prog_type = "i07" 
               CALL sadzp030_tab_gen_fk_i07(ps_tableid) RETURNING ls_pk, ls_fk
               LET ls_next = ls_fk,",",ls_pk
               LET l_token1 = base.StringTokenizer.create(ls_next,",")
               #假雙檔fk裡出現了site欄位，肯定是畫面上有這個,等等不可濾除
               IF ls_fk.getIndexOf("site",1) THEN
                  LET li_i07_fk = TRUE
               END IF
               LET ls_fk = ""
               LET ls_pk = ""  #用完就清空

            OTHERWISE
               CALL sadzp030_tab_gen_fk_t01(ps_tableid,g_table_main) RETURNING ls_pk, ls_fk,li_def_site
               LET ls_next = ls_fk,",",ls_pk
               LET l_token1 = base.StringTokenizer.create(ls_next, ",")
               LET ls_fk = ""
               LET ls_pk = ""  #用完就清空
         END CASE

         #site欄位都會補上,此處應先濾除
         LET ls_temp = ps_tableid.subString(1,ps_tableid.getIndexOf("_t",1)-1)
         WHILE l_token1.hasMoreTokens()
            LET ls_next = l_token1.nextToken()
            IF ls_next = ls_temp||"site" AND NOT li_i07_fk THEN
               CONTINUE WHILE
            ELSE
               LET ls_pk = ls_pk,ls_next,"=? AND "
            END IF
         END WHILE
         LET ls_pk = ls_pk.subString(1,ls_pk.getLength()-5)

         #補回營運據點欄位 (site) 於第二個位置
         IF sadzp030_tab_chk_site(ps_tableid) THEN
            IF ls_pk IS NOT NULL THEN
               LET ls_pk = ls_temp,"site=? AND ",ls_pk
            ELSE
               LET ls_pk = ls_temp,"site=? "
            END IF
         END IF

         #補回企業編號欄位 (ent) 於第一個位置
         IF sadzp030_tab_chk_ent(ps_tableid) THEN
            IF ls_pk IS NOT NULL THEN
               LET ls_pk = ls_temp,"ent=? AND ",ls_pk
            ELSE
               LET ls_pk = ls_temp,"ent=? "
            END IF
         END IF

      WHEN ls_type = "b_fill_sql" 
         CASE
            #i02的單身,如果沒有企業編號,通常就是 "1=1"
            WHEN g_prog_type = "i02" OR
                 g_prog_type = "c02a" OR g_prog_type = "c02b" OR
               ( g_prog_type = "t02" AND ps_tableid = g_table_main ) OR
               ( g_prog_type = "q01" AND ps_tableid = g_table_main ) OR
               ( g_prog_type = "q02" AND ps_tableid = g_table_main ) OR
               ( g_prog_type = "q03" AND ps_tableid = g_table_main ) OR
               ( g_prog_type = "c04a" AND ps_tableid = g_table_main ) 
               LET ls_pk = sadzp030_tab_entfix(ps_tableid),"1=1 "

            #i04 有type要比對FK 濾除掉跟type 同一個欄位
            WHEN g_prog_type = "i04"
               #找出與main table間的關係(FK)
               LET ls_next = sadzp030_tab_get_relation(ps_tableid,"fk",g_table_main)
               LET l_token1 = base.StringTokenizer.create(ls_next, ",")
               WHILE l_token1.hasMoreTokens()
                  LET ls_next = l_token1.nextToken()
                  
                  LET ls_next = ls_next.trim()
                  IF sadzp030_tab_gen_chk_type(ls_next) THEN 
                     CONTINUE WHILE 
                  ELSE 
                     LET ls_pk = ls_pk,ls_next,"=? AND "
                  END IF 
                  
               END WHILE
               LET ls_pk = ls_pk.subString(1,ls_pk.getLength()-5)

            #i07的單身,要將pk濾除掉已經出現在單頭的部分
            WHEN ( g_prog_type = "i07" AND g_table_main = ps_tableid) OR
                 ( g_prog_type = "i12" AND g_table_main = ps_tableid)
               CALL sadzp030_tab_gen_fk_i07(ps_tableid) RETURNING ls_pk, ls_fk
               LET ls_pk = ""
               LET l_token1 = base.StringTokenizer.create(ls_fk, ",")
               WHILE l_token1.hasMoreTokens()
                  LET ls_next = l_token1.nextToken()
                  LET ls_pk = ls_pk,ls_next,"=? AND "
               END WHILE
               LET ls_pk = sadzp030_tab_entfix(ps_tableid),ls_pk.subString(1,ls_pk.getLength()-5)

            #t01的部分, 與剩餘t02的部分
            WHEN g_prog_type = "t01" OR g_prog_type = "t02" OR g_prog_type = "i08" OR
                 g_prog_type = "i09" OR g_prog_type = "q01" OR g_prog_type = "q02" OR g_prog_type = "q04" OR
                 g_prog_type = "q03" OR g_prog_type = "c04a" OR g_prog_type = "c03a" OR   #181029-00039#1
               ( g_prog_type = "i07" AND g_table_main <> ps_tableid) OR
               ( g_prog_type = "i12" AND g_table_main <> ps_tableid) 
               #找出與main table間的關係(FK)
               #驗證一下如果不是第三階table,才可以傳入g_table_main
               LET ls_next = sadzp030_tab_get_relation(ps_tableid,"fk",sadzp030_tab_get_uptable(ps_tableid))
               LET l_token1 = base.StringTokenizer.create(ls_next, ",")
               WHILE l_token1.hasMoreTokens()
                  LET ls_next = l_token1.nextToken()
                  LET ls_next = ls_next.trim()

                  LET ls_pk = ls_pk,ls_next,"=? AND "
               END WHILE
               LET ls_pk = ls_pk.subString(1,ls_pk.getLength()-5)

            OTHERWISE
         END CASE
   END CASE

   RETURN ls_pk
END FUNCTION


#確定上階table是哪一張表
PRIVATE FUNCTION sadzp030_tab_get_uptable(ps_table)
   DEFINE ps_table   STRING   #目標table
   DEFINE ps_uptable STRING   #上階table
   DEFINE li_cnt     LIKE type_t.num5

   FOR li_cnt = 1 TO g_detail_id.getLength()
      IF g_detail_id[li_cnt].detail_tab = ps_table THEN
         IF g_detail_id[li_cnt].layer3_up IS NULL THEN
            LET ps_uptable = g_table_main
         ELSE
            LET ps_uptable = g_detail_id[li_cnt].layer3_up
         END IF
         EXIT FOR
      END IF
   END FOR

   RETURN ps_uptable
END FUNCTION


#抓取最後一個欄位id
PRIVATE FUNCTION sadzp030_tab_get_last(ls_cols)
   DEFINE ls_cols    STRING
   DEFINE ls_token   STRING
   DEFINE ls_return  STRING
   DEFINE tok        base.StringTokenizer

   IF ls_cols.getLength() < 1 THEN
      RETURN ""
   END IF

   LET tok = base.StringTokenizer.create(ls_cols,",")
   WHILE tok.hasMoreTokens()
      LET ls_token = tok.nextToken()
      LET ls_token = ls_token.trim()
      IF ls_token.getLength() >= 1 THEN
         LET ls_return = ls_token
      END IF
   END WHILE
   RETURN ls_return 
END FUNCTION

PRIVATE FUNCTION sadzp030_tab_master_not_main_fix(ls_cols)
   DEFINE ls_cols    STRING
   DEFINE ls_token   STRING
   DEFINE ls_return  STRING
   DEFINE tok        base.StringTokenizer

   IF ls_cols.getLength() < 1 THEN
      RETURN ""
   END IF
   LET tok = base.StringTokenizer.create(ls_cols,",")
   WHILE tok.hasMoreTokens()
      LET ls_token = tok.nextToken()
      LET ls_token = ls_token.trim()
      IF ls_token = "''" OR ls_token = '""' THEN
      ELSE
         LET ls_return = ls_return,ls_token,","
      END IF
   END WHILE
   RETURN ls_return.subString(1,ls_return.getLength()-1) 
END FUNCTION

#比對傳進來的reference欄位,是否已經寫入全域陣列,若沒有則補入,已有則離開
PUBLIC FUNCTION sadzp030_tab_gen_reference_array(ls_colid)
   DEFINE ls_colid  STRING

   IF NOT g_reference_col_set.getIndexOf(ls_colid,1) THEN
      LET g_reference_col_set = g_reference_col_set,ls_colid,","
   END IF

END FUNCTION



#濾除 = 號不等於傳入 table 欄位的那一個部分,留下等於 target_table 的欄位名稱回傳
FUNCTION sadzp030_tab_clearfyfk(ls_fkstr,ls_target_table)

   DEFINE ls_fkstr        STRING
   DEFINE ls_target_table STRING
   DEFINE ls_purefk       STRING
   DEFINE ls_next         STRING
   DEFINE l_line          base.StringTokenizer

   LET l_line = base.StringTokenizer.create(ls_fkstr, "=")

   WHILE l_line.hasMoreTokens()
      LET ls_next = l_line.nextToken()
      #暫時先假設 table id 是 4 碼，還沒考慮 行業別表/客製表
      IF ls_next.getIndexOf(ls_target_table.subString(1,4),1) THEN
         LET ls_purefk = ls_purefk,",",ls_next
      END IF
   END WHILE
   
   RETURN ls_purefk.subString(2,ls_purefk.getLength())
END FUNCTION


PUBLIC FUNCTION sadzp030_tab_get_relation(ps_tableid,ps_type,ps_uptable)

   DEFINE ps_tableid  STRING    #要取的主要關注表編號(PK,FK)
   DEFINE ps_type     STRING    #型態 PK OR FK
   DEFINE ps_uptable  STRING    #FK關注的上階表
   DEFINE ls_str      STRING
   DEFINE ls_sql      STRING
   DEFINE lc_dzed004  LIKE dzed_t.dzed004   #FK欄位
   DEFINE lc_dzed006  LIKE dzed_t.dzed006   #對應主表的PK欄位
   DEFINE ls_site     STRING
   DEFINE ls_next     STRING
   DEFINE l_line      base.StringTokenizer
   DEFINE l_token     base.StringTokenizer
   DEFINE ls_temp     STRING

   LET ls_str = NULL
   CASE
      #取得各資料表的PK
      WHEN ps_type = "pk"

         LET lc_dzed004 = sadzp030_tab_relation_pk(ps_tableid)

         LET l_line = base.StringTokenizer.create(lc_dzed004,",")
         WHILE l_line.hasMoreTokens()
            LET ls_next = l_line.nextToken()
            #如果主表有 site 欄位存在,就刪除 
            IF NOT g_table_main_pk.getIndexOf(g_table_main.trim()||"site",3) THEN
               LET ls_site = g_table_main.trim(),"site"
               IF ls_site = ls_next THEN
                  CONTINUE WHILE
               END IF
            END IF
            LET ls_str = ls_str, ls_next.trim(),","
         END WHILE
         LET ls_str = ls_str.subString(1,ls_str.getLength()-1)

      #取得各資料表的FK
      WHEN ps_type = "fk"
         CALL sadzp030_tab_relation_fk(ps_tableid,ps_uptable,1)
              RETURNING lc_dzed004,lc_dzed006
         LET ls_str = lc_dzed004 CLIPPED 
   END CASE
   RETURN ls_str

END FUNCTION


PRIVATE FUNCTION sadzp030_tab_page_tab(lc_dzfs004)

   DEFINE ls_str      STRING
   DEFINE ls_sql      STRING
   DEFINE ls_dzfs003  STRING
   DEFINE lc_dzfs004  LIKE dzfs_t.dzfs004
   DEFINE li_cnt      LIKE type_t.num5

   LET ls_str = ""

   FOR li_cnt = 1 TO g_detail_id.getLength()
      IF g_detail_id[li_cnt].detail_tab = lc_dzfs004 AND
         g_detail_id[li_cnt].connect = "Y" THEN
         LET ls_str = ls_str, g_detail_id[li_cnt].detail_sn USING "<<<<<" ,","
      END IF
   END FOR

   RETURN ls_str.subString(1,ls_str.getLength()-1)

END FUNCTION


PRIVATE FUNCTION sadzp030_tab_chk_site(lc_dzeb001)

   DEFINE li_ent     LIKE type_t.num5
   DEFINE lc_dzeb001 LIKE dzeb_t.dzeb001
   DEFINE ls_temp    STRING
   DEFINE ls_pkstr   STRING
   DEFINE ls_relatetable STRING

   LET ls_relatetable = lc_dzeb001
   LET ls_temp = ls_relatetable.subString(1,ls_relatetable.getIndexOf("_t",1)-1),"site"

   #如果傳進來的table id(lc_dzeb001)是 g_table_main
   IF g_table_main = lc_dzeb001 AND lc_dzeb001 IS NOT NULL THEN
      #且 g_table_main_pk 也出現site
      IF g_table_main_pk.getIndexOf(ls_temp,1) AND g_prog_type <> "i02" THEN
         #表示這個欄位有出現在畫面上,就不加上了 (i02出現也還是要加)
         IF g_t100debug >= 6 THEN
            DISPLAY cl_getmsg("adz-01123",g_lang)
           #DISPLAY "INFO: 偵測到site欄位有出現在畫面上,而且樣板不是使用i02 (i02有用也是要加上),判別不加!"
         END IF
         RETURN FALSE
      END IF
   END IF

   #確認table內是否已經有 site
  #SELECT COUNT(1) INTO li_ent FROM dzeb_t
  # WHERE dzeb001 = lc_dzeb001 AND dzeb002 LIKE "%site" AND dzeb004="Y"

   LET ls_pkstr = sadzp030_tab_relation_pk(lc_dzeb001)
   IF ls_pkstr.getIndexOf(ls_temp,1) THEN
      IF g_t100debug >= 6 THEN
         DISPLAY "INFO: 偵測欄位",ls_temp,"出現在",lc_dzeb001,"的PK(",ls_pkstr,"),判別加!"
      END IF
      LET li_ent = TRUE
   ELSE
      IF g_t100debug >= 6 THEN
         DISPLAY "INFO: 偵測欄位",ls_temp,"未出現在",lc_dzeb001,"的PK(",ls_pkstr,"),判別不加!"
      END IF
      LET li_ent = FALSE
   END IF

   RETURN li_ent

END FUNCTION


PRIVATE FUNCTION sadzp030_tab_chk_ent(lc_dzeb001)

   DEFINE li_ent     LIKE type_t.num5
   DEFINE lc_dzeb001 LIKE dzeb_t.dzeb001

   SELECT COUNT(1) INTO li_ent FROM dzeb_t
    WHERE dzeb001 = lc_dzeb001 AND dzeb002 LIKE "%ent" AND dzeb004="Y"

   RETURN li_ent

END FUNCTION


#+當i04樣板 有type要比對FK 濾除掉跟type 同一個欄位
PRIVATE FUNCTION sadzp030_tab_gen_chk_type(ps_field)
   DEFINE ps_field STRING 
   IF g_dzff.type = ps_field THEN 
      RETURN TRUE
   ELSE 
      RETURN FALSE 
   END IF 
END FUNCTION 


#+ 檢查傳入的欄位清單中是否含有 XXXsite,若有則回傳TRUE/無則回傳FALSE
PRIVATE FUNCTION s_adzp030_tab_chk_site_4fd(ls_fields)
   DEFINE ls_fields   STRING
   DEFINE ls_colid    STRING
   DEFINE li_return   LIKE type_t.num5

   #決定site欄位名稱
   LET ls_colid = g_table_main.subString(1,g_table_main.getIndexOf("_t",1)-1),"site"

   IF ls_fields.getIndexOf(ls_colid,1) THEN
      LET li_return = TRUE
   ELSE
      LET li_return = FALSE
   END IF
   RETURN li_return
END FUNCTION

#+ 找出reference的欄位名稱 
PRIVATE FUNCTION sadzp030_tab_get_reference_col(ls_columnid,ls_set)

   DEFINE ls_columnid  STRING
   DEFINE ls_set       STRING
   DEFINE ls_refid     STRING

   LET ls_refid = ls_columnid,"_desc"

   RETURN ls_refid
END FUNCTION

############################################################
#+ @code
#+ 函式目的 回傳指定的action(dzad002) 清單
#+          若傳入 "db" 則抓取所有 "db_" 開頭的 action id 回傳,注意dzad005可能同時出現YN
#+          若傳入空值，則抓取第一碼非 'm|d' 且第四碼非 '_' 的 action id 回傳
#+ @param   pc_type    char(20) 回傳型態
#+ @return  STRING     action清單
############################################################
PUBLIC FUNCTION sadzp030_tab_action(pc_type)

   DEFINE pc_type    LIKE type_t.chr20
   DEFINE ls_sql     STRING
   DEFINE l_dzad002  LIKE dzad_t.dzad002   #action識別碼
   DEFINE l_dzad007  LIKE dzad_t.dzad007   #產生標準程式框架
   DEFINE ls_str     STRING

   LET ls_str = "" #初始化清空

   #如果是 "db" 表示為主 menu, 附加基本功能項目
   IF pc_type = "db" THEN
      LET ls_sql = "SELECT DISTINCT dzad002,dzad007 FROM dzaa_t,dzad_t ",
                 #---------------------------dzaa--------------#
                   " WHERE dzaa001 = '",g_prog_id CLIPPED,"' ",
                     " AND dzaa002 = '",g_sd_ver CLIPPED,"' ",
                     " AND dzaa005 = '2' ",                       #識別碼類型
#                    " AND dzaa008 = '",g_erp_ver CLIPPED,"' ",   #產品版本
                     " AND dzaa009 = '",g_dgenv CLIPPED,"' ",     #客製標示
#                    " AND dzaa010 = '",g_cust CLIPPED,"' ",      #客戶編號
                     " AND dzaastus = 'Y' ",
                 #---------------------------dzad--------------#
                     " AND dzad001 = dzaa001 ",      #規格編號
                     " AND dzad002 = dzaa003 ",      #Action識別碼
                     " AND dzad005 = dzaa006 ",      #使用標示
                     " AND dzad003 = dzaa004 ",      #識別碼版次
                     " AND (dzad006 = 'all' ",       #觸發時機
                       " OR dzad006 LIKE 'all,%' ",     #觸發時機
                       " OR dzad006 LIKE '%,all' ",     #觸發時機
                       " OR dzad006 LIKE '%,all,%') ",  #觸發時機
#                    " AND dzad008 = dzaa010 ",      #識別碼版次
                     " AND dzadstus = 'Y' "          #有效碼
   ELSE
      #細項先決定,再組合出主表 (因為開始要動 pc_type)
      LET pc_type = pc_type CLIPPED
      LET ls_sql = "SELECT DISTINCT dzad002,dzad007 FROM dzaa_t,dzad_t ",
                 #---------------------------dzaa--------------#
                   " WHERE dzaa001 = '",g_prog_id CLIPPED,"'",
                     " AND dzaa002 = '",g_sd_ver CLIPPED,"'",
                     " AND dzaa005 = '2'",                        #識別碼類型
#                    " AND dzaa008 = '",g_erp_ver CLIPPED,"' ",   #產品版本
                     " AND dzaa009 = '",g_dgenv CLIPPED,"' ",     #客製標示
#                    " AND dzaa010 = '",g_cust CLIPPED,"' ",      #客戶編號
                     " AND dzaastus = 'Y'",
                 #---------------------------dzad--------------#
                     " AND dzad001 = dzaa001 ",      #規格編號
                     " AND dzad002 = dzaa003 ",      #Action識別碼
                     " AND dzad005 = dzaa006 ",      #使用標示
                     " AND dzad003 = dzaa004 ",      #識別碼版次
                     " AND (dzad006 = '",pc_type CLIPPED,"' ",       #觸發時機
                       " OR dzad006 LIKE '",pc_type CLIPPED,",%' ",     #觸發時機
                       " OR dzad006 LIKE '%,",pc_type CLIPPED,"' ",     #觸發時機
                       " OR dzad006 LIKE '%,",pc_type CLIPPED,",%') ",  #觸發時機
#                    " AND dzad008 = dzaa010 ",      #識別碼版次
                     " AND dzadstus = 'Y' "          #有效碼
   END IF

   LET ls_sql = ls_sql, "ORDER BY dzad002"   #230710-00021

   PREPARE dzad_prep FROM ls_sql
   DECLARE dzad_cs CURSOR FOR dzad_prep

   FOREACH dzad_cs INTO l_dzad002,l_dzad007
      #161130-00002 ---start---
      #此欄位NULL時判定填入Y
      IF l_dzad007 IS NULL THEN 
         LET l_dzad007 = "Y"
      END IF
      #161130-00002 --- end ---

      IF cl_null(ls_str) THEN
         LET ls_str = l_dzad002,"(",l_dzad007,")"
      ELSE
         LET ls_str = ls_str CLIPPED,",",l_dzad002 CLIPPED,"(",l_dzad007,")"
      END IF
   END FOREACH

   FREE dzad_cs

   RETURN ls_str
END FUNCTION

#+ 定位s_detail的順序,以避免畫面上插入一個就亂生一個,造成全域變數名稱混亂
PRIVATE FUNCTION sadzp030_tab_gen_pageid()
   DEFINE la_pageid   DYNAMIC ARRAY OF RECORD
             record      STRING,
             page        LIKE type_t.num5
                      END RECORD
   DEFINE li_cnt      LIKE type_t.num5
   DEFINE li_pos      LIKE type_t.num5
   DEFINE ls_temp     STRING
   DEFINE li_finish   LIKE type_t.num5

   #變數清理
   CALL la_pageid.clear()

   #以樹狀型態表示的單檔多欄 (暫時取消) 特別處理
   IF g_prog_type = "i03" THEN
      LET la_pageid[1].record = "s_browse" 
      LET la_pageid[1].page = 1
      RETURN la_pageid
   END IF

   #處理一般頁簽
   FOR li_cnt = 1 TO g_detail_id.getLength()

      LET ls_temp = g_detail_id[li_cnt].detail_id

      #有筆數就開始進行比序
      IF la_pageid.getLength() > 0 THEN
         LET li_finish = FALSE
         FOR li_pos = 1 TO la_pageid.getLength()
            #比序, 前比後還小就用isert填入,比較大(FLASE)就跳過
            IF NOT li_finish AND sadzp030_tab_page_sort(ls_temp,la_pageid[li_pos].record) THEN
               CALL la_pageid.insertElement(li_pos)
               LET la_pageid[li_pos].record = ls_temp
               LET li_finish = TRUE
            END IF
         END FOR
         #比完了還沒有比他小的,就是最大
         IF NOT li_finish THEN
            LET la_pageid[la_pageid.getLength()+1].record = ls_temp
         END IF
      ELSE
         LET la_pageid[1].record = ls_temp
      END IF
   END FOR

   #最後填上數字
   FOR li_cnt = 1 TO la_pageid.getLength()
      LET la_pageid[li_cnt].page = li_cnt
   END FOR

   RETURN la_pageid
END FUNCTION

#+ 如果ls_temp1比ls_temp2小,回傳TRUE,其他回傳FALSE
PRIVATE FUNCTION sadzp030_tab_page_sort(ls_temp1,ls_temp2) 
   DEFINE ls_temp1  STRING
   DEFINE ls_temp2  STRING
   DEFINE li_length1, li_length2  LIKE type_t.num5
   DEFINE li_return LIKE type_t.num5
   DEFINE li_cnt    LIKE type_t.num5

   #先比字串長度,比較長的就是大
   LET li_length1 = ls_temp1.getLength()
   LET li_length2 = ls_temp1.getLength()
 
   IF li_length1 <> li_length2 THEN
      CASE
         WHEN li_length1 < li_length2  LET li_return = TRUE
            RETURN li_return
         WHEN li_length1 > li_length2  LET li_return = FALSE
            RETURN li_return
         OTHERWISE
      END CASE
   END IF 

   #長度一樣的時候比內容(轉ASCII逐字比較)
   FOR li_cnt = 1 TO ls_temp1.getLength()
      LET li_length1 = ASCII(ls_temp1.subString(li_cnt,li_cnt))
      LET li_length2 = ASCII(ls_temp2.subString(li_cnt,li_cnt))
      CASE
         WHEN li_length1 < li_length2  LET li_return = TRUE
            EXIT FOR
         WHEN li_length1 > li_length2  LET li_return = FALSE
            EXIT FOR
         OTHERWISE
      END CASE
   END FOR

   RETURN li_return

END FUNCTION


PRIVATE FUNCTION s_adzp030_tab_remove_quoter(ls_cols)
   DEFINE ls_cols   STRING
   DEFINE ls_return STRING
   DEFINE l_token   base.StringTokenizer
   DEFINE ls_next   STRING

   IF ls_cols.getLength() < 2 THEN
      RETURN ls_cols
   END IF

   LET l_token = base.StringTokenizer.create(ls_cols, ",")
   WHILE l_token.hasMoreTokens()
      #抓取每個欄位
      LET ls_next = l_token.nextToken()
      IF ls_next.getIndexOf("(",1) AND ls_next.getIndexOf(")",1) AND 
         ls_next.getIndexOf("(",1) < ls_next.getIndexOf(")",1) THEN
         LET ls_next = ls_next.subString(1,ls_next.getIndexOf("(",1)-1)
      END IF
      LET ls_return = ls_return,",",ls_next
   END WHILE

   RETURN ls_return.subString(2,ls_return.getLength())
END FUNCTION

#+ Q類專用:比對傳入的ls_detail_col欄位清單和原始定義清單,
#  如果原始定義清單內的欄位為被串查用的顯示欄位,就剔除在ls_detail_col內
PRIVATE FUNCTION sadzp030_tab_exclude_q(ls_detail_col,li_pos)
   DEFINE ls_detail_col  STRING
   DEFINE li_pos         LIKE type_t.num5
   DEFINE li_cnt         LIKE type_t.num5
   DEFINE l_token1       base.StringTokenizer
   DEFINE l_token2       base.StringTokenizer
   DEFINE ls_return      STRING
   DEFINE ls_next1,ls_next2 STRING
   DEFINE lc_dzal002     LIKE dzal_t.dzal002

   LET l_token1 = base.StringTokenizer.create(ls_detail_col, ",")
   LET l_token2 = base.StringTokenizer.create(g_col_set[li_pos+1].fields,",")
   WHILE l_token1.hasMoreTokens()
      #抓取每個欄位
      LET ls_next1 = l_token1.nextToken()
      LET ls_next2 = l_token2.nextToken()
      LET lc_dzal002 = ls_next2.trim()
      SELECT COUNT(1) INTO li_cnt FROM dzal_t
       WHERE dzal001 = g_prog_id
         AND dzal002 = lc_dzal002
#        AND dzal009 = g_cust     #客戶編號
         AND dzalstus = "Y"
      IF li_cnt = 1 THEN
      ELSE
         LET ls_return = ls_return,ls_next1,","
      END IF
   END WHILE
   RETURN ls_return.subString(1,ls_return.getLength()-1)
END FUNCTION

#+ p/r01兩個樣板有INPUT就沒有CONSTRUCT的特殊功能
PRIVATE FUNCTION sadzp030_tab_notexists(ls_source,ls_minus)
   DEFINE ls_source   STRING
   DEFINE ls_minus    STRING
   DEFINE ls_return   STRING
   DEFINE la_col      DYNAMIC ARRAY OF STRING
   DEFINE l_token1    base.StringTokenizer
   DEFINE l_token2    base.StringTokenizer
   DEFINE ls_next     STRING
   DEFINE li_cnt      LIKE type_t.num5

   #拆解minus
   LET l_token1 = base.StringTokenizer.create(ls_minus, ",")
   WHILE l_token1.hasMoreTokens()
      LET ls_next = l_token1.nextToken()
      LET la_col[la_col.getLength()+1] = ls_next
   END WHILE

   #比對
   LET l_token2 = base.StringTokenizer.create(ls_source, ",")
   WHILE l_token2.hasMoreTokens()
      #抓取每個欄位
      LET ls_next = l_token2.nextToken()
      FOR li_cnt = 1 TO la_col.getLength()
         IF ls_next = la_col[li_cnt] THEN
            CONTINUE WHILE
         END IF
      END FOR
      LET ls_return = ls_return,ls_next,","
   END WHILE
   RETURN ls_return.subString(1,ls_return.getLength()-1)

END FUNCTION

FUNCTION sadzp030_tab_gen_append_pk(ps_append_pk)
   DEFINE ps_append_pk    STRING #次單頭PK
   DEFINE ls_master_pk    STRING #主單頭PK
   DEFINE ls_token        STRING
   DEFINE lst_token       base.StringTokenizer
   DEFINE ls_token2       STRING
   DEFINE lst_token2      base.StringTokenizer
   DEFINE la_append_pk    DYNAMIC ARRAY OF STRING #次單頭PK list
   DEFINE la_master_pk    DYNAMIC ARRAY OF STRING #主單頭PK list
   DEFINE li_idx          INTEGER
   DEFINE ls_return       STRING
   
   LET ls_master_pk = g_table_main_pk

   #拆解兩表key
   LET lst_token  = base.StringTokenizer.create(ls_master_pk,",")
   LET lst_token2 = base.StringTokenizer.create(ps_append_pk,",")
   WHILE TRUE
      IF lst_token.hasMoreTokens()  THEN
         LET la_master_pk[la_master_pk.getLength()+1] = lst_token.nextToken()
      END IF
      IF lst_token2.hasMoreTokens()  THEN
         LET la_append_pk[la_append_pk.getLength()+1] = lst_token2.nextToken()
      END IF
      IF NOT lst_token.hasMoreTokens() AND NOT lst_token2.hasMoreTokens() THEN
         EXIT WHILE
      END IF
   END WHILE
   
   #兩邊一樣 不處理
   IF la_master_pk.getLength() = la_append_pk.getLength() THEN
      LET ls_return = ps_append_pk
   END IF
   
   #兩邊不一樣 需處理(site問題)
   IF la_master_pk.getLength() <> la_append_pk.getLength() THEN
      IF la_master_pk.getLength() = la_append_pk.getLength() - 1 THEN
         FOR li_idx = 1 TO la_append_pk.getLength()
            IF la_append_pk[li_idx].getIndexOf('site',5) = 0 THEN
               LET ls_return = ls_return,la_append_pk[li_idx],','
            END IF
         END FOR
         LET ls_return = ls_return.subString(1,ls_return.getLength()-1)
      ELSE
         DISPLAY cl_getmsg("adz-01124",g_lang)
        #DISPLAY "ERROR: 單頭主表與單頭子表PK數量不一致!"
      END IF
   END IF
   
   RETURN ls_return
  
END FUNCTION

