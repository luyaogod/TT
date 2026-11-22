#+ 程式版本......: T6 Version 1.00.00 Build-0001 at 14/02/27
#
#+ 程式代碼......: sadzp060_4
#+ 設計人員......: Hiko
# Prog. Version..: 'T6-12.01.21(00000)'     #
#
# Program name   : sadzp060_4.4gl
# Description    : 標準程式同步行業功能
# Modify         : 2015/01/15 by Hiko : 新建程式
#                : 2015/03/10 by Hiko : 將同步清單寫入dzar_t內
#                : 2015/03/13 by Hiko : 執行同步前,讓開發者可以看清楚要更新的前後內容
#                : 2015/04/29 by Hiko : 取得本次標準異動的清單,並且dzaq007改成PK(Key Word:STD_LIST)
#                : 2015/05/11 by Hiko : 精準更新本次標準異動的清單(Key Word:TARGET_LIST)
#                : 2015/05/19 by Hiko : 支持整批同步(Key Word:BATCH_SYNC)
#                : 2015/05/26 by Hiko : 同步後要重產tsd/tap(因為tab/4gl/4fd都已經是新的,若有人沒有簽出就下載,就會出現不一致的情況)
#                : 2015/06/03 by Hiko : 同步後由呼叫端來重產tsd/tap
#                : 2015/06/12 by Hiko : 同步改成未簽出的時候執行, 所以得自行長dzaf_t. 也就是說, 同步也是只能執行一次.
#                : 2015/06/30 by Hiko : 若沒有版次資料,則不寫入dzaq_t.
#                : 2015/08/19 by Hiko : 修改引用過程的版次問題,重點是改adzp640.
#                : 2015/09/18 by Hiko : 新增dzaq_t前先判斷是否有版次
#                : 20160118 by Hiko : 1.簡化同步機制, 不特別判斷dzaq_t的版次, 一律以最新的標準版次為同步來源.
#                                     2.若已經同步過, 還是可以透過adzp640來處理.
#                                     3.移除dzar_t的預覽機制, 改成呼叫adzq991來查看.
#                                     4.增加dzau_t/dzbu_t來記錄行業對應之標準規格/apt引用清單, 讓引用過程更有效率.
#                : 20160226 160226-00009 by Hiko : 1.調整sadzp060_4_batch_ins_dzaq_t參數.
#                                                  2.補上未簽出的檢查.
#                                                  3.補上dzaq013(模組別)的賦值.
#                : 20160727 160726-00007 by Hiko : 移除dzau_t,dzbu_t沒有資料的警告.
#                : 20160929 160929-00037 by Hiko : 1.只要標準與行業都不是Free Style或都是Free Style就可以同步.
#                                                  2.只要標準或行業有解開Section, 就在同步後提示訊息.
#                : 20170313 170313-00025 by Hiko : 取消dzaf_t和dzaq_t標準版次的判斷,最終的合併紀錄是完整的就可以了.
#                : 20170424 170424-00013 by madey: 1.修正bug:避免free-style程式同步時又重覆轉了一次free style
#                                                  2.修正bug:將部份程式段落往上搬(建立規格樹及代碼樹)
#                                        by Hiko : 1.修正bug:恢復copy_dzag_t,dzfs_t (Hiko)
#                                                  2.同步有錯時不rollback方便查看問題(Hiko):移除gb_trans_flag相關程式段.
#                                                  3.增加dzau002,dzbu002來記錄同步的行業規格與程式版次,可在退版的時候使用
#                                                  4.增加dzau006,dzbu005來記錄同步的標準規格與程式版次,可在同步的時候判斷哪些範圍的設計點要變成失效.
#                                                  5.修正INSERT dzaa_t與dzba_t資料重複的問題.
#                                                  6.修正dzba006(函式順序)的衝突.
#                                                  7.修正畫面,Action多語言同步問題.
#                                                  8.移除大片註解,讓程式乾淨一些.
#                                                  9.增加dzaq014(過單設定明細)來記錄有同步多語言與ON ACTION.
#                : 20170522 170519-00055 by Hiko : 1.支持S,F類的行業同步:
#                                                  2.增加dzaq015(建構類型).
#                                                  3.調整和dzaq_t相關的程式段落.
#                                                  4.調整非M類的程式控制.
#                                                  5.增加cl_progress的設定,以及g_bgjob的開關.
#                                                  6.計算各步驟的執行時間.
#                                                  7.變數整理:上一版的dzaf:prev_dzaf;本版的沒有加上前贅詞.
#                                                  8.因為Patch的程式一定是過正後的,所以僅同步標準版次與dzau_t/dzbu_t紀錄版次不同的資料,以加快同步速度.
#                                                    備註:小包Patch有可能是從第五區打包,但版次也一定是大於第六區,因此不會出現相同版次但內容不同的問題.
#                                                  9.改寫dzag_t,dzfs_t,dzam_t的同步規則:改成先刪除原本版次的資料,再新增進去,這樣比較完整.
#                                                 10.因為此程式一定是在標準行業環境執行,所有dzaf010都一定是's';所有使用標示都改成's'.
#                                                 11.為了因應子程式,子畫面的同步,所以規格池設計資料的取得改成逐一欄位.
#                                                    所有程式內容/規格描述,以及和控件編號有關的資料,在異動時都要將標準程式編號取代成行業程式編號.
#                                                 12.非一對多的關係的程式/規格資源池設計資料剔除xxxstus的取得與異動.
#                : 20170703 170703-00069 by Hiko : 以dzagownid,dzfsownid,dzamownid來當作行業改過的標示:'i',同步時就避開這些資料,其他就覆蓋.
#                : 20170726 170726-00035 by Hiko : 1.過濾建構類型為M,S,F,Q才有資格寫入dzaq_t
#                                                  2.sadzp060_4_batch_ins_dzaq_t要刪除adzi051的設定(dzbk_t),以提供給Patch流程使用.
#                : 20170802 170801-00004 by Hiko : 因應tab gen也會關注規格資源池, 所以sadzp060_4在更新資源池的時候, 也要變成生效狀態.
#                : 20170921 170921-00007 by 02632 : 1.修正INSERT INTO dzak_t失敗的問題.
#                                                   2.同步時,預設"引用否"為'Y'.
#                : 20190711 190711-00035 by 02632 : 修正主程式畫面資源池資料沒有同步的問題.
#                : 20200528 200529-00048 by 02632 : 解決dzfk009更新問題.
#                : 20211110 211109-00032 by 02632 : 修正dzfk009更新問題.

import os
import xml

SCHEMA ds

GLOBALS "../../cfg/top_global.inc"
GLOBALS
   DEFINE g_stop_chk_sd_ver_freestyle BOOLEAN
END GLOBALS

&include "../4gl/sadzp200_type.inc"
&include "../4gl/sadzp000_type.inc"

DEFINE g_ind_prog LIKE dzaf_t.dzaf001, #行業程式編號
       g_ind_sver LIKE dzaf_t.dzaf003, #行業規格版次(dzaf003)
       g_ind_cver LIKE dzaf_t.dzaf004, #行業程式版次(dzaf004)
       g_prev_ind_sver LIKE dzaf_t.dzaf003, #舊版行業規格版次(dzaf003)
       g_prev_ind_cver LIKE dzaf_t.dzaf004, #舊版行業程式版次(dzaf004)
       g_std_prog LIKE dzaf_t.dzaf001, #標準程式編號
       g_std_sver LIKE dzaf_t.dzaf003, #標準規格版次(dzaf003)
       g_std_cver LIKE dzaf_t.dzaf003, #標準程式版次(dzaf004)
       g_type LIKE dzaf_t.dzaf005,     #170519-00055
       gs_cust STRING,  
       gs_dept STRING, 
       g_date DATETIME YEAR TO SECOND,
       gs_erpid STRING,          #產品代號
       gs_erpver STRING,         #ERP大版版號
       g_env LIKE dzaa_t.dzaa009 #辨識目前所在的環境:s.產中環境,c.行業環境

#標準規格變數宣告:
PRIVATE TYPE TYPE_DZAA RECORD
                       dzaa001 LIKE dzaa_t.dzaa001,   #程式編號
                       dzaa002 LIKE dzaa_t.dzaa002,   #程式版次
                       dzaa003 LIKE dzaa_t.dzaa003,   #識別碼
                       dzaa004 LIKE dzaa_t.dzaa004,   #識別碼版次
                       dzaa005 LIKE dzaa_t.dzaa005,   #識別碼類型
                       dzaa006 LIKE dzaa_t.dzaa006,   #使用標示
                       dzaa007 LIKE dzaa_t.dzaa007,   #規格引用否
                       dzaa008 LIKE dzaa_t.dzaa008,   #產品版本
                       dzaa009 LIKE dzaa_t.dzaa009,   #客製
                       dzaa010 LIKE dzaa_t.dzaa010,   #客戶編號
                       dzaastus LIKE dzaa_t.dzaastus, #狀態碼
                       targetid LIKE dzaa_t.dzaa003,  #目標識別碼
                       upd_flag LIKE type_t.chr1,     #更新標示:a:新增;u:更新
                       upd_ver  LIKE dzaa_t.dzaa004   #更新版次
                       END RECORD
#標準程式變數宣告:
PRIVATE TYPE TYPE_DZBA RECORD
                       dzba001 LIKE dzba_t.dzba001,   #程式編號
                       dzba002 LIKE dzba_t.dzba002,   #程式版次
                       dzba003 LIKE dzba_t.dzba003,   #識別碼
                       dzba004 LIKE dzba_t.dzba004,   #識別碼版次
                       dzba005 LIKE dzba_t.dzba005,   #使用標示
                       dzba006 LIKE dzba_t.dzba006,   #函式順序
                       dzba007 LIKE dzba_t.dzba007,   #程式引用否
                       dzba008 LIKE dzba_t.dzba008,   #產品版本
                       dzba009 LIKE dzba_t.dzba009,   #下方的硬結構編號整段註解
                       dzba010 LIKE dzba_t.dzba010,   #客製
                       dzba011 LIKE dzba_t.dzba011,   #客戶編號
                       dzbastus LIKE dzba_t.dzbastus, #狀態碼
                       targetid LIKE dzba_t.dzba003,  #更新目標識別碼
                       upd_flag LIKE type_t.chr1,     #更新標示:a:新增;u:更新
                       upd_ver  LIKE dzba_t.dzba004,  #更新版次
                       upd_apt  LIKE dzbb_t.dzbb098   #更新標準內容
                       END RECORD
#紀錄執行時間.
DEFINE m_time_s DATETIME YEAR TO SECOND,
       m_time_e DATETIME YEAR TO SECOND,
       m_time_start DATETIME YEAR TO SECOND,
       m_time_end DATETIME YEAR TO SECOND

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 初始化變數.
# Input parameter : none
# Return code     : BOOLEAN(成功:TRUE,失敗:FALSE)
# Date & Author   : 2015/01/15 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_4_init_var()
   LET g_date = cl_get_current()
   LET gs_dept = g_dept CLIPPED
   LET gs_erpid = FGL_GETENV("ERPID") CLIPPED
   LET gs_erpver = FGL_GETENV("ERPVER") CLIPPED
   LET gs_cust = FGL_GETENV("CUST") CLIPPED
   LET g_env = FGL_GETENV("DGENV") CLIPPED
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 擷取錯誤訊息.
# Input parameter : p_tracing 執行的顯現資訊
#                   p_sql 執行的SQL
# Return code     : void
# Date & Author   : 2015/01/15 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_4_err_catch(p_tracing, p_sql)
   DEFINE p_tracing STRING,
          p_sql     STRING

   DISPLAY "[sadzp060_4]ERROR:",p_tracing
   DISPLAY "            STATUS=",STATUS
   DISPLAY "            SQLCA.SQLCODE=",SQLCA.SQLCODE
   DISPLAY "            ls_sql=",p_sql
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC
# Descriptions    : 批次新增sync資訊(dzaq_t)
# Input parameter : p_patch_no  匯入包編號
# Return code     : BOOLEAN  TRUE:成功;FALSE:失敗
#                 : STRING   錯誤訊息
# Date & Author   : 2015/05/22 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_4_batch_ins_dzaq_t(p_patch_no) #160226-00009
   DEFINE p_patch_no  LIKE dzye_t.dzye001
   DEFINE ls_tracing  STRING,
          li_i        SMALLINT,
          ls_sql      STRING,
          la_dzye     DYNAMIC ARRAY OF RECORD
                      dzye002 LIKE dzye_t.dzye002, #程式編號 #Begin:170726-00035
                      dzye003 LIKE dzye_t.dzye003, #建構類型
                      dzye005 LIKE dzye_t.dzye005, #模組
                      dzye014 LIKE dzye_t.dzye014, #來源建構版次
                      dzye019 LIKE dzye_t.dzye019, #來源程式客製否
                      dzye020 LIKE dzye_t.dzye020, #來源規格版次
                      dzye021 LIKE dzye_t.dzye021  #來源程式版次 #End:170726-00035
                      END RECORD,
          lo_std_dzaf T_DZAF_T,
          lb_result   BOOLEAN,
          ls_err_msg  STRING,
          lsb_err     base.StringBuffer,
          l_patch_no  LIKE dzye_t.dzye001,
          l_prog      LIKE dzye_t.dzye002,
          l_date      DATETIME YEAR TO SECOND, 
          l_fail_prog LIKE dzye_t.dzye018
   
   LET ls_tracing = "sadzp060_4_batch_ins_dzaq_t start !"
   display ls_tracing

   CALL sadzp060_4_init_var()

   LET lsb_err = base.StringBuffer.create()

   #160226-00009:曾經有大片註解.

   #Begin:170726-00035
   #LET ls_sql = "SELECT * FROM dzye_t WHERE dzye001=?"
   LET ls_sql = "SELECT dzye002,dzye003,dzye005,dzye014,dzye019,dzye020,dzye021",
                "  FROM dzye_t",
                " WHERE dzye001=?",
                "   AND dzye003='M'",
                "   AND EXISTS (SELECT 1 FROM gzzb_t",
                "                WHERE gzzb003=dzye002",
                "                  AND gzzb001<>gzzb003",
                "              )",
                " UNION ",
                "SELECT dzye002,dzye003,dzye005,dzye014,dzye019,dzye020,dzye021",
                "  FROM dzye_t",
                " WHERE dzye001=?",
                "   AND dzye003 IN ('S','F','Q')",
                "   AND EXISTS (SELECT 1 FROM dzas_t",
                "                WHERE dzas003=dzye002",
                "                  AND dzas001<>dzas003",
                "              )",
                " order by dzye002"
   #End:170726-00035
   
   PREPARE prep_dzye FROM ls_sql
   DECLARE curs_dzye CURSOR FOR prep_dzye
   OPEN curs_dzye USING p_patch_no,p_patch_no
   LET ls_sql = NULL
   LET li_i = 1
   FOREACH curs_dzye INTO la_dzye[li_i].*
      INITIALIZE lo_std_dzaf TO NULL
      LET lo_std_dzaf.dzaf001 = la_dzye[li_i].dzye002 #程式代碼
      LET lo_std_dzaf.dzaf002 = la_dzye[li_i].dzye014 #來源建構版次
      LET lo_std_dzaf.dzaf003 = la_dzye[li_i].dzye020 #來源規格版次
      LET lo_std_dzaf.dzaf004 = la_dzye[li_i].dzye021 #來源程式版次
      LET lo_std_dzaf.dzaf005 = la_dzye[li_i].dzye003 #建構類型
      LET lo_std_dzaf.dzaf006 = la_dzye[li_i].dzye005 #模組
      LET lo_std_dzaf.dzaf010 = la_dzye[li_i].dzye019 #來源程式客制否

    # #Begin:170519-00055
    # #目前只有M類才可以做行業同步.
    # IF lo_std_dzaf.dzaf005<>"M" THEN
    #    LET li_i = li_i + 1 #160226-00009
    #    CONTINUE FOREACH
    # END IF 
      
      #CALL sadzp060_4_ins_dzaq_t(lo_std_dzaf.*, p_patch_no, "") RETURNING lb_result,ls_err_msg

      IF lo_std_dzaf.dzaf005="M" OR
         lo_std_dzaf.dzaf005="S" OR
         lo_std_dzaf.dzaf005="F" OR
         lo_std_dzaf.dzaf005="Q" THEN
         CALL sadzp060_4_ins_dzaq_t(lo_std_dzaf.*, p_patch_no, "") RETURNING lb_result,ls_err_msg
      END IF

      LET li_i = li_i + 1
      #End:170519-00055

      #20160118 by Hiko:曾經有大片註解.
   END FOREACH

   #20160118 by Hiko:曾經有大片註解.
 
   DELETE FROM dzbk_t #170726-00035
   
   LET ls_tracing = "sadzp060_4_batch_ins_dzaq_t finish !"
   display ls_tracing

   RETURN TRUE,NULL
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC
# Descriptions    : 新增sync資訊(dzaq_t)
# Input parameter : po_std_dzaf 簽出當下的標準DZAF_T資料
#                 : p_req_no     需求單號
#                 : p_req_order  需求單項次 #170424-00013:dzaq002已經不是需求單項次,改成no use.
# Return code     : BOOLEAN  TRUE:成功;FALSE:失敗
#                 : STRING   錯誤訊息
# Date & Author   : 2015/01/15 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_4_ins_dzaq_t(po_std_dzaf, p_req_no, p_req_order)
   DEFINE po_std_dzaf T_DZAF_T,
          p_req_no    LIKE dzaq_t.dzaq003,
          p_req_order LIKE dzaq_t.dzaq002
   DEFINE la_gzzb    DYNAMIC ARRAY OF RECORD
                     gzzb001 LIKE gzzb_t.gzzb001, #行業程式代號
                     gzzb002 LIKE gzzb_t.gzzb002  #行業編號
                     END RECORD,
          li_idx     SMALLINT,
          l_prog     LIKE gzzb_t.gzzb001,
          ls_tracing STRING,
          ls_sql     STRING,
          ls_info    STRING,
          ls_err_msg STRING,
          li_cnt     SMALLINT,
          lo_dzaf    T_DZAF_T

   TRY
      #IF po_std_dzaf.dzaf005<>"M" THEN
      IF NOT (po_std_dzaf.dzaf005="M" OR 
              po_std_dzaf.dzaf005="S" OR
              po_std_dzaf.dzaf005="F" OR
              po_std_dzaf.dzaf005="Q") THEN #170519-00055
         LET ls_info = "[INFO]Only M,S,F,Q type of industry program can do synchronous."
         DISPLAY ls_info
         RETURN TRUE,ls_info
      END IF 

      LET ls_tracing = "sadzp060_4_ins_dzaq_t start !"
      display ls_tracing

      CALL sadzp060_4_init_var()

      #依據標準程式代號找到對應的各行業程式代號.
      IF po_std_dzaf.dzaf005="M" THEN
         LET ls_sql = "SELECT gzzb001,gzzb002 FROM gzzb_t",
                      " WHERE gzzb003=?",
                      "   AND gzzb003<>gzzb001",
                      " ORDER BY 1"
      ELSE
         #Begin:170519-00055:S,F,Q
         LET ls_sql = "SELECT dzas001,dzas002 FROM dzas_t",
                      " WHERE dzas003=?",
                      "   AND dzas003<>dzas001",
                      " ORDER BY 1"
         #End:170519-00055
      END IF

      PREPARE prep_gzzb FROM ls_sql
      DECLARE curs_gzzb CURSOR FOR prep_gzzb
      OPEN curs_gzzb USING po_std_dzaf.dzaf001

      LET ls_sql = NULL #20160118 by Hiko
      LET li_idx = 1
      FOREACH curs_gzzb INTO la_gzzb[li_idx].*
         LET l_prog = la_gzzb[li_idx].gzzb001
         
         #取得行業版次資訊:lo_dzaf
         CALL sadzp060_2_get_curr_ver_info(l_prog, po_std_dzaf.dzaf005, po_std_dzaf.dzaf006) RETURNING lo_dzaf.*,ls_err_msg
         IF cl_null(ls_err_msg) THEN #2015/06/30 by Hiko
            #IF (NOT cl_null(lo_dzaf.dzaf003) AND lo_dzaf.dzaf003<>0) AND
            #   (NOT cl_null(lo_dzaf.dzaf004) AND lo_dzaf.dzaf004<>0) THEN #170726-00035 #2015/09/18 by Hiko
               #這裡可以不用管行業程式是否已經同步過,總之就是要換新資料就是了.
               #DELETE FROM dzaq_t WHERE dzaq001=l_prog AND dzaq007=po_std_dzaf.dzaf002 
               DELETE FROM dzaq_t WHERE dzaq001=l_prog #170726-00035
               
               LET ls_sql = "INSERT INTO dzaq_t(dzaq001,dzaq002,dzaq003,dzaq004,dzaq005,",
                            "                   dzaq006,dzaq007,dzaq008,dzaq009,dzaq010,",
                            "                   dzaq011,dzaq012,dzaq013,dzaq014,dzaq015,", #160226-00009:增加dzaq013 #170424-00013:增加dzaq014 #170519-00055:增加dzaq015
                            "                   dzaqownid,dzaqowndp,dzaqcrtid,dzaqcrtdp,dzaqcrtdt)",
                            "            VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
               PREPARE ins_dzaq_prep FROM ls_sql
               #211109-00032:PGS:將number型態的欄位(dzaq004、005、006)預設值從''改成NULL(g_null).
               EXECUTE ins_dzaq_prep USING l_prog,p_req_order,p_req_no,g_null,g_null,
                                           g_null,po_std_dzaf.dzaf002,po_std_dzaf.dzaf003,po_std_dzaf.dzaf004,'N',
                                           la_gzzb[li_idx].gzzb002,po_std_dzaf.dzaf001,po_std_dzaf.dzaf006,'',po_std_dzaf.dzaf005, #160226-00009 #170424-00013
                                           g_user,g_dept,g_user,g_dept,g_date
               FREE ins_dzaq_prep
               LET ls_sql = NULL #20160118 by Hiko
            #END IF
         END IF

         LET li_idx = li_idx + 1
      END FOREACH
 
      LET ls_tracing = "sadzp060_4_ins_dzaq_t finish !"
      display ls_tracing

      RETURN TRUE,NULL
   CATCH
      LET ls_tracing = "sadzp060_4_ins_dzaq_t failure !"
      CALL sadzp060_4_err_catch(ls_tracing, ls_sql)
      RETURN FALSE,"ERROR:"||ls_tracing
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC
# Descriptions    : 更新sync資訊(dzaq_t)
# Input parameter : po_std_dzaf 標準程式DZAF_T資料
#                 : po_ind_dzaf 行業程式DZAF_T資料
#                 : p_dzaq014 過單提示設定
# Return code     : BOOLEAN TRUE:成功;FALSE:失敗
#                 : STRING  錯誤訊息
# Date & Author   : 2015/01/15 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_4_upd_dzaq_t(po_std_dzaf, po_ind_dzaf, p_dzaq014) #170424-00013
   DEFINE po_std_dzaf T_DZAF_T, 
          po_ind_dzaf T_DZAF_T,
          p_dzaq014   LIKE dzaq_t.dzaq014
   DEFINE l_prog     LIKE dzaf_t.dzaf001,
          ls_tracing STRING,
          ls_sql     STRING,
          ls_err_msg STRING,
          li_cnt     SMALLINT,
          lb_result  BOOLEAN

   TRY
      LET ls_tracing = "sadzp060_4_upd_dzaq_t start !"
      display ls_tracing

      LET l_prog = po_ind_dzaf.dzaf001

      #Begin:170519-00055:簡化做法.
    # CALL sadzp060_4_get_max_dzaq_t(l_prog) RETURNING lb_result
    # IF lb_result THEN
      SELECT COUNT(1) INTO li_cnt FROM dzaq_t WHERE dzaq001=l_prog
      IF li_cnt>0 THEN
      #End:170519-00055
         LET ls_sql = "UPDATE dzaq_t SET dzaq004=?,", #行業建構版次
                                        "dzaq005=?,", #行業規格版次
                                        "dzaq006=?,", #行業程式版次
                                        "dzaq007=?,", #170313-00025:標準建構版次
                                        "dzaq008=?,", #170313-00025:標準規格版次
                                        "dzaq009=?,", #170313-00025:標準程式版次
                                        "dzaq010='Y',",
                                        "dzaq014=?,", #170424-00013
                                        "dzaqmoddt=?,",
                                        "dzaqmodid=?",
                      " WHERE dzaq001=?"
         PREPARE upd_dzaq_prep2 FROM ls_sql
         EXECUTE upd_dzaq_prep2 USING po_ind_dzaf.dzaf002,po_ind_dzaf.dzaf003,po_ind_dzaf.dzaf004,
                                      po_std_dzaf.dzaf002,po_std_dzaf.dzaf003,po_std_dzaf.dzaf004, #170313-00025
                                      p_dzaq014,g_date,g_user,l_prog
         FREE upd_dzaq_prep2
      ELSE
         LET ls_err_msg = "Cannot find the ",l_prog," data within dzaq_t."
         GOTO _RTN_ERR
      END IF
  
      LET ls_tracing = "sadzp060_4_upd_dzaq_t finish !"
      display ls_tracing
 
      RETURN TRUE,NULL

      LABEL _RTN_ERR :
         LET ls_tracing = "sadzp060_4_upd_dzaq_t break !"
         display ls_tracing
 
         RETURN FALSE,ls_err_msg
   CATCH
      LET ls_tracing = "sadzp060_4_upd_dzaq_t failure !","\n",ls_err_msg
      CALL sadzp060_4_err_catch(ls_tracing, ls_sql)
      RETURN FALSE,"ERROR:"||ls_tracing
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 取得行業對應的最大標準建構版次的dzaq_t
# Input parameter : p_dzaq001 行業程式代號
# Return code     : BOOLEAN TRUE:成功;FALSE:失敗
# Date & Author   : 2015/04/29 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_4_get_max_dzaq_t(p_dzaq001)
   DEFINE p_dzaq001 LIKE dzaq_t.dzaq001
   DEFINE ls_tracing  STRING,
          li_cnt      SMALLINT,
          ls_err_msg  STRING,
          l_std_prog  LIKE dzaq_t.dzaq012,
          l_module    LIKE dzaq_t.dzaq013,
          l_construct LIKE dzaq_t.dzaq015,
          lo_std_dzaf T_DZAF_T

   TRY
      LET ls_tracing = "sadzp060_4_get_max_dzaq_t start !"
      display ls_tracing
 
      #20160118 by Hiko:曾經有大片註解.

      #Begin:170519-00055:改變作法.
    # INITIALIZE gr_dzaq.* TO NULL
    # LET li_cnt = 0
    # SELECT COUNT(1) INTO li_cnt FROM dzaq_t WHERE dzaq001=p_dzaq001
    # IF li_cnt>0 THEN
    #    LET ls_sql = "SELECT * FROM dzaq_t WHERE dzaq001='",p_dzaq001,"'"
    #    PREPARE dzaq_prep5 FROM ls_sql
    #    EXECUTE dzaq_prep5 INTO gr_dzaq.*
    #    FREE dzaq_prep5
    # ELSE
    #    RETURN FALSE
    # END IF

    # RETURN TRUE

      SELECT COUNT(1),dzaq012,dzaq013,dzaq015
        INTO li_cnt,l_std_prog,l_module,l_construct
        FROM dzaq_t WHERE dzaq001=p_dzaq001
       GROUP BY dzaq012,dzaq013,dzaq015
      IF li_cnt>0 THEN
         CALL sadzp060_2_get_curr_ver_info(l_std_prog, l_construct, l_module) RETURNING lo_std_dzaf.*,ls_err_msg
      ELSE
         LET ls_err_msg = cl_getmsg("adz-00961", g_lang) #查無例行性/累積包的Patch更新記錄.
      END IF

      LET ls_tracing = "sadzp060_4_get_max_dzaq_t finish !"
      display ls_tracing
 
      RETURN ls_err_msg,lo_std_dzaf.*
      #End:170519-00055
   CATCH
      LET ls_tracing = "sadzp060_4_get_max_dzaq_t failure !"
      CALL sadzp060_4_err_catch(ls_tracing, "")
      RETURN ls_err_msg,lo_std_dzaf.*
   END TRY
END FUNCTION

#20160118 by Hiko:曾經有大片註解:移除整個FUNCTION:sadzp060_4_show_sync_data

##########################################################################
# Access Modifier : PUBLIC
# Descriptions    : 判斷是否可以執行行業同步
# Input parameter : po_prev_ind_dzaf 舊版行業版次資料
# Return code     : BOOLEAN  TRUE:成功;FALSE:失敗
#                 : STRING 錯誤訊息
#                 : DZAF_T 標準程式DZAF_T資料
# Date & Author   : 2015/03/13 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_4_chk_data(po_prev_ind_dzaf)
   DEFINE po_prev_ind_dzaf T_DZAF_T
   DEFINE ls_tracing      STRING,
          ls_sql          STRING,
          lb_chkout       BOOLEAN,
          lb_std_not_free BOOLEAN,
          lb_ind_not_free BOOLEAN,
          lb_result       BOOLEAN,
          lb_result2      BOOLEAN,
          ls_err_code     STRING,
          ls_err_msg      STRING,
          lo_std_dzaf     T_DZAF_T,
          li_sync_cnt     SMALLINT,
          lb_ver_same     BOOLEAN,
          ls_diff_kind    STRING, #CONS/SPEC/CODE
          li_dzaq_ver     SMALLINT,
          li_dzaf_ver     SMALLINT,
          ls_replace_arg  STRING,
          lo_prog_info    DYNAMIC ARRAY OF T_PROGRAM_INFO,
          ls_owner        STRING

   TRY
      LET ls_tracing = "sadzp060_4_chk_data start !"
      display ls_tracing
 
      CALL sadzp060_4_init_var()

      #160726-00007:曾經有大片註解

      #Begin:170519-00055
    # IF po_prev_ind_dzaf.dzaf005<>"M" THEN
    #    LET ls_err_code = "adz-00522" #只有主程式才可以執行行業同步
    #    GOTO _RTN_ERR
    # END IF 
      #End:170519-00055

      IF g_account="topstd" OR g_env='c' THEN
         LET ls_err_code = "adz-00525" #非topstd登入以及標準環境才可以執行行業同步
         GOTO _RTN_ERR
      END IF

      LET lo_prog_info[1].pi_NAME = po_prev_ind_dzaf.dzaf001 #程式代碼
      LET lo_prog_info[1].pi_TYPE = po_prev_ind_dzaf.dzaf005 #類別
      CALL sadzp200_alm_get_alm_owner_by_role(lo_prog_info[1].*, "SD") RETURNING lb_chkout,ls_owner
      IF lb_chkout THEN
         LET ls_err_code = "adz-01002" #注意:此程式狀態為簽出中。(詳細簽出記錄可透過設計器的下載工具adzp050查看ALM資訊)
         GOTO _RTN_ERR
      ELSE
         CALL sadzp200_alm_get_alm_owner_by_role(lo_prog_info[1].*, "PR") RETURNING lb_chkout,ls_owner
         IF lb_chkout THEN
            LET ls_err_code = "adz-01002" #注意:此程式狀態為簽出中。(詳細簽出記錄可透過設計器的下載工具adzp050查看ALM資訊)
            GOTO _RTN_ERR
         END IF
      END IF

      LET g_ind_prog = po_prev_ind_dzaf.dzaf001 CLIPPED

      #Begin:170519-00055:改變標準dzaf_t的取得方式.
    # CALL sadzp060_4_get_max_dzaq_t(g_ind_prog) RETURNING lb_result
    # IF NOT lb_result THEN
    #    LET ls_err_msg = "Cannot find the ",g_ind_prog," data within dzaq_t."
    #    RETURN FALSE,ls_err_msg,lo_std_dzaf.*
    # END IF

    # #20160118 by Hiko:曾經有大片註解
   
    # #CALL sadzp060_2_get_curr_ver_info(gr_dzaq.dzaq012, "M", gr_dzaq.dzaq013) RETURNING lo_std_dzaf.*,ls_err_msg
    # CALL sadzp060_2_get_curr_ver_info(gr_dzaq.dzaq012, gr_dzaq.dzaq015, gr_dzaq.dzaq013) RETURNING lo_std_dzaf.*,ls_err_msg #170519-00055
    # IF NOT cl_null(ls_err_msg) THEN
    #    RETURN FALSE,ls_err_msg,lo_std_dzaf.*
    # END IF

      CALL sadzp060_4_get_max_dzaq_t(g_ind_prog) RETURNING ls_err_msg,lo_std_dzaf.*
      IF NOT cl_null(ls_err_msg) THEN
         RETURN FALSE,ls_err_msg,lo_std_dzaf.*
      END IF
      #End:170519-00055

      LET g_std_prog = lo_std_dzaf.dzaf001

      LET lb_ver_same = TRUE

      #判斷標準程式和行業程式是否為Free Style.
      CALL sadzp169_02_chk_not_free_style(g_std_prog, 's') RETURNING lb_std_not_free,ls_err_msg
      CALL sadzp169_02_chk_not_free_style(g_ind_prog, 's') RETURNING lb_ind_not_free,ls_err_msg
      IF lb_std_not_free<>lb_ind_not_free THEN #160929-00037
         LET ls_err_code = "adz-00523" #只有標準與行業都不是Free Style或都是Free Style才可以執行行業同步.
         GOTO _RTN_ERR
      END IF

      LET ls_tracing = "sadzp060_4_chk_data finish !"
      display ls_tracing
 
      RETURN TRUE,NULL,lo_std_dzaf.*

      LABEL _RTN_ERR :
         LET ls_tracing = "sadzp060_4_chk_data break !"
         display ls_tracing
 
         LET ls_err_msg = "ERROR:",cl_getmsg(ls_err_code, g_lang)
         RETURN FALSE,ls_err_msg,lo_std_dzaf.*
   CATCH
      LET ls_tracing = "sadzp060_4_chk_data failure !"
      CALL sadzp060_4_err_catch(ls_tracing, ls_sql)
      RETURN FALSE,ls_err_msg,lo_std_dzaf.*
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC
# Descriptions    : 執行規格/程式的sync
# Input parameter : po_prev_ind_dzaf 舊版行業版次資料
# Return code     : BOOLEAN  TRUE:成功;FALSE:失敗
#                 : STRING   錯誤訊息
# Date & Author   : 2015/01/15 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_4_start_sync(po_prev_ind_dzaf)
   DEFINE po_prev_ind_dzaf T_DZAF_T
   DEFINE ls_sql           STRING,
          li_cnt           SMALLINT,
          lb_result        BOOLEAN,
          ls_info          STRING,
          ls_err_msg       STRING,
          ls_section_info  STRING,
          ls_msg_replace   STRING,
          lo_std_dzaf      T_DZAF_T,
          ls_run_cmd       STRING, 
          lo_ind_dzaf      T_DZAF_T, 
          l_new_dzaf002    LIKE dzaf_t.dzaf002, 
          ls_dzaq014       LIKE dzaq_t.dzaq014,
          ls_message       STRING,
          lb_open_porgress BOOLEAN

   TRY
      DISPLAY "[sadzp060_4_start_sync]INFO:Industry synchronization start !"

      LET m_time_start = cl_get_current()

      CALL sadzp060_4_chk_data(po_prev_ind_dzaf.*) RETURNING lb_result,ls_err_msg,lo_std_dzaf.*
      IF NOT lb_result THEN
         RETURN FALSE,ls_err_msg
      END IF

      DISPLAY "[sadzp060_4]INFO:Industry prog=",g_ind_prog
      DISPLAY "                 Standard prog=",g_std_prog

      #STEP 1:處理版次資料.
      #Begin:2015/06/12 by Hiko:因為變成未簽出執行同步, 所以要自己長版次與設計資料樹.
      LET g_prev_ind_sver = po_prev_ind_dzaf.dzaf003 #舊版行業規格版次
      LET g_prev_ind_cver = po_prev_ind_dzaf.dzaf004 #舊版行業程式版次

      LET lo_ind_dzaf.* = po_prev_ind_dzaf.*
      LET l_new_dzaf002 = po_prev_ind_dzaf.dzaf002 + 1
      LET lo_ind_dzaf.dzaf002 = l_new_dzaf002
      LET lo_ind_dzaf.dzaf003 = po_prev_ind_dzaf.dzaf003 + 1

      #Begin:170519-00055
      LET g_bgjob = 'N'
      LET g_type = lo_ind_dzaf.dzaf005

      IF g_type="M" OR
         g_type="S" OR
         g_type="Q" THEN
         LET lo_ind_dzaf.dzaf004 = po_prev_ind_dzaf.dzaf004 + 1

         CALL cl_progress_bar(13) #規格與程式主要有13個動作.
      ELSE #F類沒有程式.
         LET lo_ind_dzaf.dzaf004 = po_prev_ind_dzaf.dzaf004

         CALL cl_progress_bar(10) #單純規格只有10個動作.
      END IF
      #End:170519-00055

      LET lb_open_porgress = TRUE

      LET ls_message = "SPEC 1.Processing version data(dzaf_t)"
      DISPLAY "[sadzp060_4]INFO:",ls_message
      CALL cl_progress_ing(ls_message)

      DISPLAY "[sadzp060_4]INFO:Industry SPEC now revision=",lo_ind_dzaf.dzaf003
      DISPLAY "                 Industry CODE now revision=",lo_ind_dzaf.dzaf004

      LET li_cnt = 0
      SELECT COUNT(1) INTO li_cnt FROM dzaf_t
       WHERE dzaf001=g_ind_prog 
         AND dzaf002=l_new_dzaf002
         AND dzaf005=lo_ind_dzaf.dzaf005
         AND dzaf010='s'

      IF li_cnt=0 THEN
         LET ls_sql = "INSERT INTO dzaf_t(dzaf001,dzaf002,dzaf003,dzaf004,dzaf005,",
                      "                   dzaf006,dzaf007,dzaf008,dzaf009,dzaf010)",
                      "            VALUES(?,?,?,?,?,?,?,?,?,?)"
         PREPARE ins_dzaf_prep FROM ls_sql
         EXECUTE ins_dzaf_prep USING g_ind_prog,lo_ind_dzaf.dzaf002,lo_ind_dzaf.dzaf003,lo_ind_dzaf.dzaf004,lo_ind_dzaf.dzaf005,
                                     lo_ind_dzaf.dzaf006,lo_ind_dzaf.dzaf007,lo_ind_dzaf.dzaf008,lo_ind_dzaf.dzaf009,'s' 
         FREE ins_dzaf_prep
      END IF

      #STEP 2:處理標準多餘行業的多語言資料.
      LET ls_message = "SPEC 2.Processing multiple languages(gzzd_t)"
      DISPLAY "[sadzp060_4]INFO:",ls_message
      CALL cl_progress_ing(ls_message)
      #Begin:170424-00013:修正畫面,Action多語言同步問題.
      #寫入標準多餘行業的gzzd_t資料(azzi902). PK=gzzd001,gzzd002(語系),gzzd003(Key),gzzd004(s/c)
      LET ls_sql = "SELECT COUNT(1) FROM gzzd_t zd1",
                   " WHERE zd1.gzzd001='",g_std_prog,"'", #標準程式
                   "   AND zd1.gzzd003 NOT IN (",
                   "                           SELECT DISTINCT(zd2.gzzd003) FROM gzzd_t zd2",
                   "                            WHERE zd2.gzzd001='",g_ind_prog,"'", #行業程式
                   "                          )"
      PREPARE prep_gzzd0 FROM ls_sql
      EXECUTE prep_gzzd0 INTO li_cnt
      FREE prep_gzzd0

      IF li_cnt>0 THEN
         LET ls_dzaq014 = "azzi902" #有同步的時候, dzaq014寫入'azzi902'.

         LET ls_sql = "INSERT INTO gzzd_t (gzzd001,gzzd002,gzzd003,gzzd004,gzzd005,gzzd006,gzzdstus)",
                      "            SELECT '",g_ind_prog,"',zd1.gzzd002,zd1.gzzd003,zd1.gzzd004,zd1.gzzd005,zd1.gzzd006,zd1.gzzdstus FROM gzzd_t zd1",
                      "             WHERE zd1.gzzd001='",g_std_prog,"'", #標準程式
                      "               AND zd1.gzzd003 NOT IN (",
                      "                                       SELECT DISTINCT(zd2.gzzd003) FROM gzzd_t zd2",
                      "                                        WHERE zd2.gzzd001='",g_ind_prog,"'", #行業程式
                      "                                      )"
         PREPARE prep_gzzd FROM ls_sql
         EXECUTE prep_gzzd
         FREE prep_gzzd
      END IF

      #STEP 3:判斷標準多餘行業的ON ACTION資料.
      LET ls_message = "SPEC 3.Processing ON ACTION(gzzp_t,gzzq_t)"
      DISPLAY "[sadzp060_4]INFO:",ls_message
      CALL cl_progress_ing(ls_message)
      #寫入標準多餘行業的gzzp_t資料(azzi903). PK=gzzp001,gzzd002(Key)
      LET ls_sql = "SELECT COUNT(1) FROM gzzp_t zp1",
                   " WHERE zp1.gzzp001='",g_std_prog,"'", #標準程式
                   "   AND zp1.gzzp002 NOT IN (",
                   "                           SELECT DISTINCT(zp2.gzzp002) FROM gzzp_t zp2",
                   "                            WHERE zp2.gzzp001='",g_ind_prog,"'", #行業程式
                   "                          )"
      PREPARE prep_gzzp0 FROM ls_sql
      EXECUTE prep_gzzp0 INTO li_cnt
      FREE prep_gzzp0

      IF li_cnt>0 THEN
         IF cl_null(ls_dzaq014) THEN
            LET ls_dzaq014 = "azzi903" #有同步的時候, dzaq014寫入'azzi903'.
         ELSE
            LET ls_dzaq014 = ls_dzaq014,",azzi903"
         END IF

         LET ls_sql = "INSERT INTO gzzp_t (gzzp001,gzzp002,gzzp003,gzzp004,gzzp005,gzzp006,gzzp007,gzzpstus)",
                      "            SELECT '",g_ind_prog,"',zp1.gzzp002,zp1.gzzp003,zp1.gzzp004,zp1.gzzp005,zp1.gzzp006,zp1.gzzp007,zp1.gzzpstus FROM gzzp_t zp1",
                      "             WHERE zp1.gzzp001='",g_std_prog,"'", #標準程式
                      "               AND zp1.gzzp002 NOT IN (",
                      "                                       SELECT DISTINCT(zp2.gzzp002) FROM gzzp_t zp2",
                      "                                        WHERE zp2.gzzp001='",g_ind_prog,"'", #行業程式
                      "                                      )"
         PREPARE prep_gzzp FROM ls_sql
         EXECUTE prep_gzzp
         FREE prep_gzzp
         
         #寫入標準多餘行業的gzzq_t資料(azzi903). PK=gzzq001,gzzq002(Key),gzzq003(語系)
         LET ls_sql = "INSERT INTO gzzq_t (gzzq001,gzzq002,gzzq003,gzzq004,gzzq005,gzzq006)",
                      "            SELECT '",g_ind_prog,"',zq1.gzzq002,zq1.gzzq003,zq1.gzzq004,zq1.gzzq005,zq1.gzzq006 FROM gzzq_t zq1",
                      "             WHERE zq1.gzzq001='",g_std_prog,"'", #標準程式
                      "               AND zq1.gzzq002 NOT IN (",
                      "                                       SELECT DISTINCT(zq2.gzzq002) FROM gzzq_t zq2",
                      "                                        WHERE zq2.gzzq001='",g_ind_prog,"'", #行業程式
                      "                                      )"
         PREPARE prep_gzzq FROM ls_sql
         EXECUTE prep_gzzq
         FREE prep_gzzq
      END IF
      #End:170424-00013

      #2015/08/19 by Hiko::曾經有大片註解
 
      LET g_stop_chk_sd_ver_freestyle = TRUE #170424-00013:強制不進入轉free-style那段邏輯

      LET g_ind_sver = lo_ind_dzaf.dzaf003 #行業規格版次
      LET g_std_sver = lo_std_dzaf.dzaf003 #標準規格版次

      #STEP 4:建立規格樹.
      LET ls_message = "SPEC 4.Create SPEC tree(dzaa_t)"
      DISPLAY "[sadzp060_4]INFO:",ls_message
      CALL cl_progress_ing(ls_message)
      #建立規格樹 #170424-00013:由下面搬上來.
      LET g_bgjob = 'Y'
      CALL sadzp060_2_create_spec_structure(g_ind_prog, lo_ind_dzaf.*, po_prev_ind_dzaf.*) RETURNING ls_err_msg
      LET g_bgjob = 'N'
      IF NOT cl_null(ls_err_msg) THEN GOTO _RTN_ERR END IF
      #若有需要更新才處理,所以移到sync_dzag才做.
    # #STEP 5:複製資料表關聯設定. #刻意放在這邊就處理,以免過程出錯.
    # LET ls_message = "SPEC 5.Copy the table association settings(dzag_t)"
    # DISPLAY "[sadzp060_4]INFO:",ls_message
    # CALL cl_progress_ing(ls_message)
    # #要先複製舊版資料.
    # LET g_bgjob = 'Y'
    # CALL sadzp060_4_copy_dzag_t(po_prev_ind_dzaf.*) RETURNING lb_result
    # LET g_bgjob = 'N'
    # IF NOT lb_result THEN
    #    CALL cl_progress_bar_close()
    #    RETURN FALSE,"[sadzp060_4]ERROR:Copy dzag_t failed!"
    # END IF
       
      #STEP 5:同步規格設計資料.
      LET ls_message = "SPEC 6.Synchronous SPEC design data(dzaa_t,dzac_t,... etc)"
      DISPLAY "[sadzp060_4]INFO:",ls_message
      CALL cl_progress_ing(ls_message)
      #170519-00055:和"建立程式樹"順序交換,這樣在sadzp168_4裡面呼叫sadzp030_tab等函式才會是正確的.
      LET g_bgjob = 'Y'
      LET m_time_s = cl_get_current()
      CALL sadzp060_4_sync_spec("SYNC", lo_std_dzaf.*, lo_ind_dzaf.*, po_prev_ind_dzaf.*) RETURNING lb_result,ls_err_msg
      LET m_time_e = cl_get_current()
      DISPLAY "-----[sadzp060_4]SPEC 6.COST TIME:",(m_time_e-m_time_s)
      LET g_bgjob = 'N'
      IF NOT lb_result THEN GOTO _RTN_ERR END IF
   
      #STEP 6:解析4fd畫面成設計資料.
      LET ls_message = "SPEC 7.Analysis 4fd into design data(sadzp168_3)"
      DISPLAY "[sadzp060_4]INFO:",ls_message
      CALL cl_progress_ing(ls_message)
      LET g_bgjob = 'Y'
      LET m_time_s = cl_get_current()
      CALL sadzp168_3(lo_ind_dzaf.dzaf006, g_ind_prog, g_ind_sver, 's') RETURNING lb_result,ls_err_msg
      LET m_time_e = cl_get_current()
      DISPLAY "-----[sadzp060_4]SPEC 7.COST TIME:",(m_time_e-m_time_s)
      LET g_bgjob = 'N'
      IF NOT lb_result THEN GOTO _RTN_ERR END IF

      #STEP 7:產生畫面欄位資訊. #失敗沒關係.
      LET ls_message = "SPEC 8.Generate 4fd field data(sadzp168_4)"
      DISPLAY "[sadzp060_4]INFO:",ls_message
      CALL cl_progress_ing(ls_message)
      LET g_bgjob = 'Y'
      LET m_time_s = cl_get_current()
      CALL sadzp168_4(g_ind_prog, g_ind_sver, 's') RETURNING lb_result,ls_info
      LET m_time_e = cl_get_current()
      DISPLAY "-----[sadzp060_4]SPEC 8.COST TIME:",(m_time_e-m_time_s)
      LET g_bgjob = 'N'
      IF NOT lb_result THEN
         DISPLAY "[sadzp060_4]WARNING:",ls_info
      END IF
 
      #STEP 8:行業別畫面設計資料同步.
      LET ls_message = "SPEC 9.Synchronous 4fd design data(sadzp168_8)"
      DISPLAY "[sadzp060_4]INFO:",ls_message
      CALL cl_progress_ing(ls_message)
      LET g_bgjob = 'Y'
      LET m_time_s = cl_get_current()
      CALL sadzp168_8(g_ind_prog, g_ind_sver, g_std_prog, g_std_sver) RETURNING lb_result,ls_err_msg
      LET m_time_e = cl_get_current()
      DISPLAY "-----[sadzp060_4]SPEC 9.COST TIME:",(m_time_e-m_time_s)
      LET g_bgjob = 'N'
      IF NOT lb_result THEN GOTO _RTN_ERR END IF
 
      #STEP 9:重新組合4fd.
      LET ls_message = "SPEC 10.Reassemble 4fd(sadzp168_5)"
      DISPLAY "[sadzp060_4]INFO:",ls_message
      CALL cl_progress_ing(ls_message)
      LET g_bgjob = 'Y'
      LET m_time_s = cl_get_current()
      CALL sadzp168_5(g_ind_prog, g_ind_sver, 's', FALSE) RETURNING lb_result,ls_info
      LET m_time_e = cl_get_current()
      DISPLAY "-----[sadzp060_4]SPEC 10.COST TIME:",(m_time_e-m_time_s)
      LET g_bgjob = 'N'
      LET lb_result = TRUE #170519-00055:忽略4fd編譯錯誤,所以重設flag.
  
      LET ls_section_info = NULL #170519-00055

      IF g_type="M" OR
         g_type="S" OR
         g_type="Q" THEN #170519-00055

         LET g_ind_cver = lo_ind_dzaf.dzaf004 #行業程式版次
         LET g_std_cver = lo_std_dzaf.dzaf004 #標準程式版次

         #STEP 10:建立程式樹. #170424-00013:由下面搬上來.
         LET ls_message = "CODE 1.Create CODE tree(dzba_t,dzbc_t)"
         DISPLAY "[sadzp060_4]INFO:",ls_message
         CALL cl_progress_ing(ls_message)
         LET g_bgjob = 'Y'
         CALL sadzp060_2_create_code_structure(g_ind_prog, lo_ind_dzaf.*, po_prev_ind_dzaf.*) RETURNING ls_err_msg
         LET g_bgjob = 'N'
         IF NOT cl_null(ls_err_msg) THEN GOTO _RTN_ERR END IF
 
         #STEP 11:同步程式設計資料.
         LET ls_message = "CODE 2.Synchronous CODE design data(dzba_t,dzbb_t,... etc)"
         DISPLAY "[sadzp060_4]INFO:",ls_message
         CALL cl_progress_ing(ls_message)
         LET g_bgjob = 'Y'
         LET m_time_s = cl_get_current()
         CALL sadzp060_4_sync_addpoint("SYNC", lo_std_dzaf.*, lo_ind_dzaf.*) RETURNING lb_result,ls_err_msg
         LET m_time_e = cl_get_current()
         DISPLAY "-----[sadzp060_4]CODE 2.COST TIME:",(m_time_e-m_time_s)
         LET g_bgjob = 'N'
         IF NOT lb_result THEN GOTO _RTN_ERR END IF
       
         #STEP 12:程式重新產生與組合.
         LET ls_message = "CODE 3.Program re-generated(r.c3)"
         DISPLAY "[sadzp060_4]INFO:",ls_message
         CALL cl_progress_ing(ls_message)
         #r.c3會自己印出訊息
         LET g_bgjob = 'Y'
         LET m_time_s = cl_get_current()
         CALL sadzp060_2_rc3(g_ind_prog, lo_ind_dzaf.*, '0') RETURNING lb_result,ls_info
         LET m_time_e = cl_get_current()
         DISPLAY "-----[sadzp060_4]CODE 3.COST TIME:",(m_time_e-m_time_s)
         LET g_bgjob = 'N'
         LET lb_result = TRUE #170519-00055:忽略編譯錯誤,所以重設flag.
      END IF
  
      #STEP 13:更新同步紀錄表.
      LET ls_message = "Update the synchronization log table(dzaq_t)"
      DISPLAY "[sadzp060_4]INFO:",ls_message
      CALL cl_progress_ing(ls_message)
      #將sync紀錄表的欄位[同步否]更新為'Y'.這邊就算失敗,也不影響sync結果,但還是要顯現有錯誤.
      CALL sadzp060_4_upd_dzaq_t(lo_std_dzaf.*, lo_ind_dzaf.*, ls_dzaq014) RETURNING lb_result,ls_info #170424-00013
      LET lb_result = TRUE #170519-00055:忽略錯誤,所以重設flag.

      DISPLAY "[sadzp060_4]INFO:Industry synchronization completed!"

      LET m_time_end = cl_get_current()
      DISPLAY "-----[sadzp060_4]TOTAL COST TIME:",(m_time_end-m_time_start)

      #RETURN TRUE,NULL
      RETURN TRUE,ls_section_info #170519-00055
   CATCH
      GOTO _RTN_ERR
   END TRY

   LABEL _RTN_ERR : #和CATCH相同
      #170424-00013:曾經有大片註解:移除RollBack的機制,這樣才比較容易除錯.
      LET g_bgjob = 'N'
      IF lb_open_porgress THEN #已經開啟cl_progress_bar才要主動關閉畫面.
         CALL cl_progress_bar_close()
      END IF

      DISPLAY "[sadzp060_4_sync_spec]ERROR:Industry synchronization failure !"
      RETURN FALSE,ls_err_msg
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE 
# Descriptions    : 同步行業規格:spec
# Input parameter : ps_act 執行動作(QUERY:查詢/SYNC:同步)
#                 : po_std_dzaf 標準dzaf_t資料
#                 : po_ind_dzaf 行業dzaf_t資料
#                 : po_prev_ind_dzaf 舊版行業版次資料
# Return code     : STRING 錯誤訊息(NULL表示成功)
# Date & Author   : 2015/01/15 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_4_sync_spec(ps_act, po_std_dzaf, po_ind_dzaf, po_prev_ind_dzaf)
   DEFINE ps_act      STRING, 
          po_std_dzaf T_DZAF_T,
          po_ind_dzaf T_DZAF_T,
          po_prev_ind_dzaf T_DZAF_T
   DEFINE ls_tracing   STRING,
          ls_sel_cols  STRING,
          ls_prev_sver STRING,
          ls_targetid_1 STRING,
          ls_targetid_2 STRING,
          ls_ind_dzaa003_1 STRING,
          ls_ind_dzaa003_2 STRING,
          ls_inner_sql STRING,
          ls_notin_sql STRING,
          ls_other_sql STRING,
          ls_sql       STRING,
          ls_ind_aa_wc STRING, 
          li_i         SMALLINT,
          dzaa_std_arr DYNAMIC ARRAY OF TYPE_DZAA,
          li_cnt       SMALLINT,
          lb_result    BOOLEAN,
          ls_err_msg   STRING,
          li_idx       SMALLINT,
          l_revision   LIKE dzaa_t.dzaa004,
          l_cite       LIKE dzaa_t.dzaa007,
          li_ind_ver   SMALLINT,
          li_ver_flag  SMALLINT, 
          ls_targetid  STRING 

   #需要同步的規格內容有:dzaa005=
   #1.欄位規格
   #2.Action規格
   #3.整體規格
   #5.樹狀結構
   #6.欄位參考
   #7.資料多語言

   TRY
      LET ls_tracing = "sadzp060_4_sync_spec start !"
      display ls_tracing

      LET lb_result = TRUE
      
      #20160118 by Hiko:SQL中間曾經有大片註解, 且SQL大改版:只更新使用版次不同的資料,以加快速度.
      #Begin:170519-00055
      IF g_type="M" THEN
         LET ls_targetid_1 = "stda1.dzaa003"
         LET ls_targetid_2 = "stda2.dzaa003"

         LET ls_ind_dzaa003_1 = "inda1.dzaa003"
         LET ls_ind_dzaa003_2 = "inda2.dzaa003"
      ELSE #S,Q
         LET ls_targetid_1 = "REPLACE(stda1.dzaa003,'",g_std_prog,"','",g_ind_prog,"')" #標準變行業.
         LET ls_targetid_2 = "REPLACE(stda2.dzaa003,'",g_std_prog,"','",g_ind_prog,"')" #標準變行業.

         LET ls_ind_dzaa003_1 = "REPLACE(inda1.dzaa003,'",g_ind_prog,"','",g_std_prog,"')" #行業變標準.
         LET ls_ind_dzaa003_2 = "REPLACE(inda2.dzaa003,'",g_ind_prog,"','",g_std_prog,"')" #行業變標準.
      END IF

      #取得上次的標準同步版次(或是還沒同步前的時候,行業程式產生當時所對應的標準版次).
      #備註:若要DEBUG此段SQL,則要特別注意同步的紀錄是否已經存在最新版(行業版次=g_ind_sver).
      LET ls_prev_sver = "(SELECT DISTINCT(au1.dzau006) FROM dzau_t au1",
                         "  WHERE au1.dzau001='",g_ind_prog,"'", #行業程式編號:取得行業程式的同步紀錄.在adzp600與adzp270產生行業程式的時候就會寫入初始引用紀錄.
                         "    AND au1.dzau002=(SELECT MAX(au2.dzau002) FROM dzau_t au2",
                         "                      WHERE au2.dzau001=au1.dzau001",
                         "                    )",
                         ")" 

      #取得新版存在於行業引用的資料(剔除TABLE和EXCLUDE),但新版使用版次不同於舊版使用版次.
      LET ls_inner_sql = "SELECT stda1.dzaa001,stda1.dzaa002,stda1.dzaa003,stda1.dzaa004,stda1.dzaa005,",
                         "       stda1.dzaa006,'Y',stda1.dzaa008,stda1.dzaa009,stda1.dzaa010,", #170921-00007:stda1.dzaa007-->'Y'
                         "       stda1.dzaastus,",
                         "       ",ls_targetid_1," targetid,", #直接給行業識別碼即可.
                         "       'u' upd_flag,",           #標示存在行業設計資料:UPDATE.
                                 g_ind_sver," upd_ver",    #直接給定要更新的版次為行業程式版次.    
                         "  FROM dzaa_t stda1",
                         " WHERE stda1.dzaa001='",g_std_prog,"'", #標準程式代號
                         "   AND stda1.dzaa002=",g_std_sver,      #標準程式版次
                         "   AND stda1.dzaa009='s'",              #客製識別碼:s
                         "   AND stda1.dzaa003 NOT IN ('TABLE','EXCLUDE')", #dzag_t,dzfs_t,dzam_t
                         "   AND stda1.dzaa003 IN (", #新版多餘行業的資料.
                         "                         SELECT ",ls_ind_dzaa003_1,
                         "                           FROM dzaa_t inda1",
                         "                          WHERE inda1.dzaa001='",g_ind_prog,"'",
                         "                            AND inda1.dzaa002=",g_prev_ind_sver,
                         "                            AND inda1.dzaa007='Y'",
                         "                            AND inda1.dzaa009='s'",        
                         "                        )",
                         "   AND stda1.dzaa004<>NVL((", #新版使用版次不同於舊版使用版次.
                         "                           SELECT prevaa.dzaa004",
                         "                             FROM dzaa_t prevaa",
                         "                            WHERE prevaa.dzaa001=stda1.dzaa001",
                         "                              AND prevaa.dzaa002=",ls_prev_sver,
                         "                              AND prevaa.dzaa003=stda1.dzaa003", 
                         "                              AND prevaa.dzaa009=stda1.dzaa009",
                         "                          ),-1)"

      #取得新版多餘行業的資料,可能會找到TABLE或EXCLUDE.
      LET ls_notin_sql = "SELECT stda2.dzaa001,stda2.dzaa002,stda2.dzaa003,stda2.dzaa004,stda2.dzaa005,",
                         "       NVL(stda2.dzaa006,0),'Y',stda2.dzaa008,stda2.dzaa009,stda2.dzaa010,", #170921-00007:stda2.dzaa007-->'Y'
                         "       stda2.dzaastus,",
                         "       ",ls_targetid_2," targetid,", #取得要更新的行業識別碼(將包含標準程式編號的識別碼換成行業程式編號).
                         "       'a' upd_flag,",              #因為是NOT IN,所以記錄行業資料為不存在:INSERT.
                                 g_ind_sver," upd_ver",       #直接給定要更新的版次為行業程式版次.    
                         "  FROM dzaa_t stda2",
                         " WHERE stda2.dzaa001='",g_std_prog,"'", #標準程式代號
                         "   AND stda2.dzaa002=",g_std_sver,      #標準程式版次
                         "   AND stda2.dzaa009='s'",              #客製識別碼:s
                         "   AND stda2.dzaa003 NOT IN ('TABLE','EXCLUDE')",
                         "   AND stda2.dzaa003 NOT IN (", #新版多餘行業的資料.
                         "                              SELECT ",ls_ind_dzaa003_2,
                         "                                FROM dzaa_t inda2",
                         "                               WHERE inda2.dzaa001='",g_ind_prog,"'",
                         "                                 AND inda2.dzaa002=",g_prev_ind_sver,
                         "                                 AND inda2.dzaa009='s'",        
                         "                             )" 

      #額外取得'TABLE','EXCLUDE'的資料.
      LET ls_other_sql = "SELECT stda3.dzaa001,stda3.dzaa002,stda3.dzaa003,stda3.dzaa004,stda3.dzaa005,",
                         "       stda3.dzaa006,'N',stda3.dzaa008,stda3.dzaa009,stda3.dzaa010,", #170921-00007:stda3.dzaa007-->'N'
                         "       stda3.dzaastus,",
                         "       stda3.dzaa003 targetid,", #'TABLE','EXCLUDE'是固定識別碼.
                         "       'u' upd_flag,",           #標示存在行業設計資料:UPDATE.
                                 g_ind_sver," upd_ver",    #直接給定要更新的版次為行業程式版次.    
                         "  FROM dzaa_t stda3",
                         " WHERE stda3.dzaa001='",g_std_prog,"'",       #標準程式代號
                         "   AND stda3.dzaa002=",g_std_sver,            #標準程式版次
                         "   AND stda3.dzaa009='s'",                    #客製識別碼:s
                         "   AND stda3.dzaa003 IN ('TABLE','EXCLUDE')",
                         "   AND stda3.dzaa004<>NVL((", #新版使用版次不同於舊版使用版次.
                         "                           SELECT prevaa.dzaa004",
                         "                             FROM dzaa_t prevaa",
                         "                            WHERE prevaa.dzaa001=stda3.dzaa001",
                         "                              AND prevaa.dzaa002=",ls_prev_sver,
                         "                              AND prevaa.dzaa003=stda3.dzaa003", 
                         "                              AND prevaa.dzaa009=stda3.dzaa009",
                         "                          ),-1)"

      LET ls_sql = ls_inner_sql,
                   " UNION ALL ",
                   ls_notin_sql,
                   " UNION ALL ",
                   ls_other_sql,
                   " ORDER BY 3"
      #End:170519-00055

      DISPLAY "[sadzp060_4_sync_spec]DEBUG:select std dzaa_t sql=",ls_sql
      PREPARE dzaa_std_prep FROM ls_sql
      DECLARE dzaa_std_curs CURSOR FOR dzaa_std_prep
      OPEN dzaa_std_curs 
      LET ls_sql = NULL #20160118 by Hiko

      LET li_i = 1
      FOREACH dzaa_std_curs INTO dzaa_std_arr[li_i].*
         LET l_cite = "Y"

         LET ls_targetid = dzaa_std_arr[li_i].targetid
         LET l_revision = dzaa_std_arr[li_i].upd_ver

         LET ls_ind_aa_wc = "     dzaa001='",g_ind_prog,"'",
                            " AND dzaa002=",g_ind_sver,
                            " AND dzaa003='",ls_targetid,"'",
                            " AND dzaa009='s'"

         #20160118:曾經有大片註解:這是為了更新時可以精準知道哪些資料是要真正新增行業資源池.

         IF dzaa_std_arr[li_i].dzaa005='4' OR 
            dzaa_std_arr[li_i].dzaa005='a' THEN #170519-00055
            LET l_cite = "N"
         END IF

         IF ps_act.equals("SYNC") THEN
            IF dzaa_std_arr[li_i].upd_flag='u' THEN #170519-00055 
               #修改一定是標準識別碼更新到行業樹內,所以皆以標準資料為主.
               LET ls_sql = "UPDATE dzaa_t SET dzaa004=",l_revision,",",                     #識別碼版次
                            "                  dzaa007='",l_cite,"',",                       #引用
                            "                  dzaastus='",dzaa_std_arr[li_i].dzaastus,"',", #生失效 
                            "                  dzaamoddt=?,",
                            "                  dzaamodid='",g_user,"'",
                            " WHERE ",ls_ind_aa_wc
               #display "sadzp060_4_sync_spec:update ind dzaa_t sql=",ls_sql
               PREPARE dzaa_cus_prep2 FROM ls_sql
               EXECUTE dzaa_cus_prep2 USING g_date
               FREE dzaa_cus_prep2
               LET ls_sql = NULL #20160118 by Hiko
            ELSE #upd_flag='a'
               #IF dzaa_std_arr[li_i].dzaastus<>"N" THEN #sync的時候,找不到行業樹內容,則標準失效的資料不需要新增進來,以免多垃圾. #170519-00055:失效的資料也得更新.
                  #新增一定是標準多餘行業的識別碼,所以皆以標準資料為主.
                  LET ls_sql = "INSERT INTO dzaa_t(dzaa001,dzaa002,dzaa003,dzaa004,dzaa005,",
                               "                   dzaa006,dzaa007,dzaa008,dzaa009,dzaa010,",
                               "                   dzaacrtdt,dzaacrtdp,dzaaowndp,dzaaownid,dzaastus,dzaacrtid)",
                               "        VALUES( '",g_ind_prog,"',",
                                                   g_ind_sver,",",                  #dzaa002:行業版次 #170424-00013
                               "                '",ls_targetid,"',",                #170519-00055
                                                   l_revision,",",                  #dzaa004 #170424-00013
                               "                '",dzaa_std_arr[li_i].dzaa005,"',",
                               "                '",dzaa_std_arr[li_i].dzaa006,"',",
                               "                   '",l_cite,"',",                  #dzaa007:引用
                               "                '",dzaa_std_arr[li_i].dzaa008,"',",
                               "                's',",                              #s:客製標示
                               "                '",dzaa_std_arr[li_i].dzaa010,"',",
                               "                   ?,",                             #dzaacrtdt
                               "                '",gs_dept,"',",                    #dzaacrtdp
                               "                '",gs_dept,"',",                    #dzaaowndp
                               "                '",g_user,"',",                     #dzaaownid
                               "                '",dzaa_std_arr[li_i].dzaastus,"',",
                               "                '",g_user,"')"                      #dzaaownid
                  #display "sadzp060_4_sync_spec:insert ind dzaa_t sql=",ls_sql
                  PREPARE dzaa_cus_prep3 FROM ls_sql
                  EXECUTE dzaa_cus_prep3 USING g_date
                  FREE dzaa_cus_prep3
                  LET ls_sql = NULL #20160118 by Hiko
               #END IF
            END IF

            #Begin:170801-00004:因為所有類型都要處理,因此由下方搬上來.        
            IF dzaa_std_arr[li_i].dzaa005='4' THEN
               CALL sadzp060_4_sync_dzag(dzaa_std_arr[li_i].*, po_prev_ind_dzaf.*) RETURNING lb_result,ls_err_msg
            END IF

            IF dzaa_std_arr[li_i].dzaa005='a' THEN
               CALL sadzp060_4_sync_dzam(dzaa_std_arr[li_i].*, po_prev_ind_dzaf.*) RETURNING lb_result,ls_err_msg
            END IF
            #End:170801-00004

            #190711-00035:主程式也是需要資源池.
            #IF g_type="S" OR g_type="F" OR g_type="Q" THEN #190711-00035 #170801-00004:行業引用主程式只需要樹頭即可,其他類型就得同時更新資源池.
               #除了'TABLE'和'EXCLUDE',其餘資源池都是在生效狀態才需要更新.
               CASE dzaa_std_arr[li_i].dzaa005
                  WHEN '1' IF dzaa_std_arr[li_i].dzaastus='Y' THEN
                              CALL sadzp060_4_sync_dzac(dzaa_std_arr[li_i].*) RETURNING lb_result,ls_err_msg
                              IF lb_result THEN #170519-00055
                                 CALL sadzp060_4_sync_dzak(dzaa_std_arr[li_i].*) RETURNING lb_result,ls_err_msg
                              END IF
                              
                              IF lb_result THEN #170519-00055
                                 CALL sadzp060_4_sync_dzal(dzaa_std_arr[li_i].*) RETURNING lb_result,ls_err_msg
                              END IF
                           END IF
                  WHEN '2' IF dzaa_std_arr[li_i].dzaastus='Y' THEN
                              CALL sadzp060_4_sync_dzad(dzaa_std_arr[li_i].*) RETURNING lb_result,ls_err_msg
                           END IF
                  WHEN '3' IF dzaa_std_arr[li_i].dzaastus='Y' THEN
                              CALL sadzp060_4_sync_dzab(dzaa_std_arr[li_i].*) RETURNING lb_result,ls_err_msg
                           END IF
                  WHEN '5' IF dzaa_std_arr[li_i].dzaastus='Y' THEN
                              CALL sadzp060_4_sync_dzff(dzaa_std_arr[li_i].*) RETURNING lb_result,ls_err_msg
                           END IF
                  WHEN '6' IF dzaa_std_arr[li_i].dzaastus='Y' THEN
                              CALL sadzp060_4_sync_dzai(dzaa_std_arr[li_i].*) RETURNING lb_result,ls_err_msg
                           END IF
                  WHEN '7' IF dzaa_std_arr[li_i].dzaastus='Y' THEN
                              CALL sadzp060_4_sync_dzaj(dzaa_std_arr[li_i].*) RETURNING lb_result,ls_err_msg
                           END IF
               END CASE
               
               IF NOT lb_result THEN
                  EXIT FOREACH
               END IF
            #END IF #190711-00035
            CALL sadzp060_4_sync_dzfk009(g_ind_prog, dzaa_std_arr[li_i].dzaa003 ) RETURNING lb_result
            IF NOT lb_result THEN
               EXIT FOREACH
            END IF
         END IF #ps_act.equals("SYNC")

         LET li_i = li_i + 1
      END FOREACH
      
      DISPLAY "-----[sadzp060_4]總共同步 ",(li_i-1)," 筆規格資料"

      #Begin:20161118 by Hiko : 最後再將標準dzaa_t的資料寫入dzau_t內.
      LET ls_tracing = "sadzp060_4_sync_spec : delete dzau_t"
      DELETE FROM dzau_t WHERE dzau001=g_ind_prog AND dzau002=g_ind_sver
      
      LET ls_tracing = "sadzp060_4_sync_spec : insert dzau_t"
      LET ls_sql = "INSERT INTO dzau_t(dzau001,dzau002,dzau003,dzau004,dzau005,dzau006)",
                   " SELECT '",g_ind_prog,"',",g_ind_sver,",dzaa003,dzaa004,dzaa005,dzaa002 FROM dzaa_t", #170424-00013:dzau002,dzau006
                   "  WHERE dzaa001='",g_std_prog,"'",
                   "    AND dzaa002=",g_std_sver
      #display "sadzp060_4_sync_spec:insert dzau_t sql=",ls_sql
      PREPARE dzau_prep2 FROM ls_sql
      EXECUTE dzau_prep2
      FREE dzau_prep2
      #End:20161118 by Hiko

      LET ls_tracing = "sadzp060_4_sync_spec finish !"
      display ls_tracing

      RETURN lb_result,ls_err_msg
   CATCH
      LET ls_tracing = "sadzp060_4_sync_spec failure !"
      CALL sadzp060_4_err_catch(ls_tracing, ls_sql)
      RETURN FALSE,"ERROR:"||ls_tracing
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE 
# Descriptions    : 同步行業程式:add point
# Input parameter : ps_act 執行動作(QUERY:查詢/SYNC:同步)
#                 : po_std_dzaf 標準dzaf_t資料
#                 : po_ind_dzaf 行業dzaf_t資料
# Return code     : BOOLEAN TRUE:成功;FALSE:失敗
#                 : STRING 錯誤訊息
# Date & Author   : 2015/01/15 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_4_sync_addpoint(ps_act, po_std_dzaf, po_ind_dzaf)
   DEFINE ps_act      STRING,
          po_std_dzaf T_DZAF_T,
          po_ind_dzaf T_DZAF_T
   DEFINE ls_tracing    STRING,
          ls_prev_cver  STRING,
          ls_targetid_1 STRING,
          ls_targetid_2 STRING,
          ls_dzbb098_1  STRING,
          ls_dzbb098_2  STRING,
          ls_ind_dzba003_1 STRING,
          ls_ind_dzba003_2 STRING,
          ls_inner_sql  STRING,
          ls_notin_sql  STRING,
          ls_sql        STRING,
          ls_ind_ba_wc  STRING, 
          li_i          SMALLINT,
          dzba_std_arr  DYNAMIC ARRAY OF TYPE_DZBA,
          li_cnt        SMALLINT,
          l_revision    LIKE dzba_t.dzba004,
          li_ind_ver    SMALLINT,
          li_ver_flag   SMALLINT,
          ls_apt        STRING,
          lb_func       BOOLEAN,
          ls_ind_bb_wc  STRING, 
          ls_targetid   STRING  

   TRY
      LET ls_tracing = "sadzp060_4_sync_addpoint start !"
      display ls_tracing

      #底下和sadzi888_07雷同

      #Begin:170519-00055
      IF g_type="M" THEN
         LET ls_targetid_1 = "stda1.dzba003"
         LET ls_targetid_2 = "stda2.dzba003"

         LET ls_dzbb098_1 = "stdb1.dzbb098"
         LET ls_dzbb098_2 = "stdb2.dzbb098"

         LET ls_ind_dzba003_1 = "inda1.dzba003"
         LET ls_ind_dzba003_2 = "inda2.dzba003"
      ELSE #S,Q
         LET ls_targetid_1 = "REPLACE(stda1.dzba003,'",g_std_prog,"','",g_ind_prog,"')" #標準變行業.
         LET ls_targetid_2 = "REPLACE(stda2.dzba003,'",g_std_prog,"','",g_ind_prog,"')" #標準變行業.

         LET ls_dzbb098_1 = "REPLACE(stdb1.dzbb098,'",g_std_prog,"','",g_ind_prog,"')"  #標準add point內容換行業.
         LET ls_dzbb098_2 = "REPLACE(stdb2.dzbb098,'",g_std_prog,"','",g_ind_prog,"')"  #標準add point內容換行業.

         LET ls_ind_dzba003_1 = "REPLACE(inda1.dzba003,'",g_ind_prog,"','",g_std_prog,"')" #行業變標準.
         LET ls_ind_dzba003_2 = "REPLACE(inda2.dzba003,'",g_ind_prog,"','",g_std_prog,"')" #行業變標準.
      END IF

      #取得要更新的標準程式版次資料

      #20160118 by Hiko:SQL中間曾經有大片註解, 且SQL大改版:只更新使用版次不同的資料,以加快速度.

      #取得上次的標準同步版次(或是還沒同步前的時候,行業程式產生當時所對應的標準版次).
      #備註:若要DEBUG此段SQL,則要特別注意同步的紀錄是否已經存在最新版(行業版次=g_ind_cver).
      LET ls_prev_cver = "(SELECT DISTINCT(bu1.dzbu005) FROM dzbu_t bu1",
                         "  WHERE bu1.dzbu001='",g_ind_prog,"'", #行業程式編號:取得行業程式的同步紀錄.在adzp600與adzp270產生行業程式的時候就會寫入初始引用紀錄.
                         "    AND bu1.dzbu002=(SELECT MAX(bu2.dzbu002) FROM dzbu_t bu2",
                         "                      WHERE bu2.dzbu001=bu1.dzbu001",
                         "                    )",
                         ")"

      #取得新版存在於行業引用的資料,但新版使用版次不同於舊版使用版次.
      LET ls_inner_sql = "SELECT stda1.dzba001,stda1.dzba002,stda1.dzba003,stda1.dzba004,stda1.dzba005,",
                         "       NVL(stda1.dzba006,0),'Y',stda1.dzba008,stda1.dzba009,stda1.dzba010,", #170921-00007:stda1.dzba007-->'Y'
                         "       stda1.dzba011,stda1.dzbastus,",
                         "       ",ls_targetid_1," targetid,", #取得要更新的行業識別碼(將包含標準程式編號的識別碼換成行業程式編號).
                         "       'u' upd_flag,",           #標示存在行業設計資料:UPDATE.
                                 g_ind_cver," upd_ver,",   #直接給定要更新的版次為行業程式版次.    
                                 ls_dzbb098_1," upd_apt",  #標準add point內容.
                         "  FROM dzba_t stda1",
                         "  LEFT JOIN dzbb_t stdb1 ON stdb1.dzbb001=stda1.dzba001",
                         "                        AND stdb1.dzbb002=stda1.dzba003",
                         "                        AND stdb1.dzbb003=stda1.dzba004",
                         "                        AND stdb1.dzbb004=stda1.dzba005",
                         " WHERE stda1.dzba001='",g_std_prog,"'", #標準程式代號
                         "   AND stda1.dzba002=",g_std_cver,      #標準程式版次
                         "   AND stda1.dzba010='s'",              #客製識別碼:s
                         "   AND stda1.dzba003 IN (", #新版存在行業引用的資料.
                         "                         SELECT ",ls_ind_dzba003_1,
                         "                           FROM dzba_t inda1",
                         "                          WHERE inda1.dzba001='",g_ind_prog,"'",
                         "                            AND inda1.dzba002=",g_prev_ind_cver,
                         "                            AND inda1.dzba007='Y'",
                         "                            AND inda1.dzba010='s'",        
                         "                        )", 
                         "   AND stda1.dzba004<>NVL((", #新版使用版次不同於舊版使用版次.
                         "                           SELECT prevba.dzba004",
                         "                             FROM dzba_t prevba",
                         "                            WHERE prevba.dzba001=stda1.dzba001",
                         "                              AND prevba.dzba002=",ls_prev_cver,
                         "                              AND prevba.dzba003=stda1.dzba003", 
                         "                              AND prevba.dzba010=stda1.dzba010",
                         "                          ),-1)"

      #取得新版多餘行業的資料.
      LET ls_notin_sql = "SELECT stda2.dzba001,stda2.dzba002,stda2.dzba003,stda2.dzba004,stda2.dzba005,",
                         "       NVL(stda2.dzba006,0),'Y',stda2.dzba008,stda2.dzba009,stda2.dzba010,", #170921-00007:stda2.dzaa007-->'Y'
                         "       stda2.dzba011,stda2.dzbastus,",
                         "       ",ls_targetid_2," targetid,", #取得要更新的行業識別碼(將包含標準程式編號的識別碼換成行業程式編號).
                         "       'a' upd_flag,",          #因為是NOT IN,所以記錄行業資料為不存在:INSERT.
                                 g_ind_cver," upd_ver,",  #直接給定要更新的版次為行業程式版次.    
                                 ls_dzbb098_2," upd_apt", #標準add point內容.
                         "  FROM dzba_t stda2",
                         "  LEFT JOIN dzbb_t stdb2 ON stdb2.dzbb001=stda2.dzba001",
                         "                        AND stdb2.dzbb002=stda2.dzba003",
                         "                        AND stdb2.dzbb003=stda2.dzba004",
                         "                        AND stdb2.dzbb004=stda2.dzba005",
                         " WHERE stda2.dzba001='",g_std_prog,"'", #標準程式代號
                         "   AND stda2.dzba002=",g_std_cver,      #標準程式版次
                         "   AND stda2.dzba010='s'",              #客製識別碼:s
                         "   AND stda2.dzba003 NOT IN (", #新版多餘行業的資料.
                         "                              SELECT ",ls_ind_dzba003_2,
                         "                                FROM dzba_t inda2",
                         "                               WHERE inda2.dzba001='",g_ind_prog,"'",
                         "                                 AND inda2.dzba002=",g_prev_ind_cver,
                         "                                 AND inda2.dzba010='s'",        
                         "                             )" 
      LET ls_sql = ls_inner_sql,
                   " UNION ALL ",
                   ls_notin_sql,
                   " ORDER BY 3"
      #End:170519-00055

      DISPLAY "[sadzp060_4_sync_addpoint]DEBUG:select std dzba_t sql=",ls_sql
      PREPARE dzba_std_prep FROM ls_sql
      DECLARE dzba_std_curs CURSOR FOR dzba_std_prep
      OPEN dzba_std_curs 
      LET ls_sql = NULL #20160118 by Hiko

      LET li_i = 1
      LOCATE dzba_std_arr[li_i].upd_apt IN FILE

      FOREACH dzba_std_curs INTO dzba_std_arr[li_i].*
         LET ls_targetid = dzba_std_arr[li_i].targetid
         LET l_revision = dzba_std_arr[li_i].upd_ver

         LET ls_ind_ba_wc = "     dzba001='",g_ind_prog,"'",
                            " AND dzba002=",g_ind_cver,
                            " AND dzba003='",ls_targetid,"'",
                            " AND dzba010='s'"

         #170519-00055:大改版:主SQL一起取得相關資訊,曾經有大片註解.

         IF ps_act.equals("SYNC") THEN #2015/03/13 by Hiko
            IF dzba_std_arr[li_i].upd_flag='u' THEN #170519-00055
               #修改一定是標準識別碼更新到行業樹內,所以皆以標準資料為主.
               LET ls_sql = "UPDATE dzba_t SET dzba004=",l_revision,",",     #設計點版次
                            "                  dzba007='Y',",                                #引用
                            "                  dzbastus='",dzba_std_arr[li_i].dzbastus,"',", #生失效 
                            "                  dzbamoddt=?,",
                            "                  dzbamodid='",g_user,"'",
                            " WHERE ",ls_ind_ba_wc
               #display "sadzp060_4_sync_addpoint:update ind dzba_t sql=",ls_sql
               PREPARE dzba_cus_prep2 FROM ls_sql
               EXECUTE dzba_cus_prep2 USING g_date
               FREE dzba_cus_prep2
               LET ls_sql = NULL #20160118 by Hiko
            ELSE #upd_flag='a'
               #IF dzba_std_arr[li_i].dzbastus<>"N" THEN #sync的時候,找不到行業樹內容,則標準失效的資料不需要新增進來,以免多垃圾. #170519-00055:失效的資料也得更新.
                  #新增一定是標準多餘行業的識別碼,所以皆以標準資料為主.
                  LET ls_sql = "INSERT INTO dzba_t(dzba001,dzba002,dzba003,dzba004,dzba005,",
                               "                   dzba006,dzba007,dzba008,dzba009,dzba010,dzba011,",
                               "                   dzbacrtdt,dzbacrtdp,dzbaowndp,dzbaownid,dzbastus,dzbacrtid)",
                               "        VALUES( '",g_ind_prog,"',",
                                                   g_ind_cver,",",                  #dzba002:行業版次 #170424-00013
                               "                '",ls_targetid,"',",                #170519-00055
                                                   l_revision,",",                  #dzba004 #170424-0001.
                               "                '",dzba_std_arr[li_i].dzba005,"',",
                               "                   ?,",                             #dzba006
                               "                   'Y',",                           #dzba007:引用
                               "                '",dzba_std_arr[li_i].dzba008,"',",
                               "                '",dzba_std_arr[li_i].dzba009,"',",
                               "                's',",                              #s:客製標示
                               "                '",dzba_std_arr[li_i].dzba011,"',",
                               "                   ?,",                             #dzbacrtdt
                               "                '",gs_dept,"',",                    #dzbacrtdp
                               "                '",gs_dept,"',",                    #dzbaowndp
                               "                '",g_user,"',",                     #dzbaownid
                               "                '",dzba_std_arr[li_i].dzbastus,"',",
                               "                '",g_user,"')"                      #dzbaownid
                  #display "sadzp060_4_sync_addpoint:insert ind dzba_t sql=",ls_sql
                  PREPARE dzba_cus_prep3 FROM ls_sql
                  EXECUTE dzba_cus_prep3 USING dzba_std_arr[li_i].dzba006,g_date
                  FREE dzba_cus_prep3
                  LET ls_sql = NULL #20160118 by Hiko

                  #Begin:170424-00013:若是函式,則要將行業程式內>=此標準函式的順序全部往後+1.
                  LET lb_func = FALSE
                  IF ls_targetid.getIndexOf("function.", 1)>0 THEN
                     LET lb_func = TRUE
                  ELSE
                     IF ls_targetid.getIndexOf("report.", 1)>0 THEN
                        LET lb_func = TRUE
                     END IF
                  END IF
                  
                  IF lb_func THEN
                     LET ls_sql = "UPDATE dzba_t SET dzba006=(dzba006+1)",
                                  " WHERE dzba001='",g_ind_prog,"'",
                                  "   AND dzba002=",g_ind_cver,
                                  "   AND (  dzba003 LIKE ('%function.%')",
                                  "       OR dzba003 LIKE ('%report.%'))",
                                  "   AND dzba003<>'",ls_targetid,"'", #170519-00055
                                  "   AND dzba006>=",dzba_std_arr[li_i].dzba006,
                                  "   AND dzba010='s'"
                     #display "sadzp060_4_sync_addpoint:offset ind dzba006 sql=",ls_sql
                     PREPARE prep1_dzba006 FROM ls_sql
                     EXECUTE prep1_dzba006
                     FREE prep1_dzba006
                  END IF
                  #End:170424-00013
               #END IF
            END IF
         END IF #ps_act.equals("SYNC")

         #170519-00055:大改版:主SQL一起取得相關資訊,曾經有大片註解.

         #為了和規格相關資源池的更新方式相同,所以ls_targetid和l_revision都改成targetid和upd_ver.

         IF ps_act.equals("SYNC") THEN #2015/03/13 by Hiko
            #更新dzbb_t:此段和sadzp060_1雷同.      
            LET li_cnt = 0
            LET ls_ind_bb_wc = " WHERE dzbb001='",g_ind_prog,"'", 
                               "   AND dzbb003=",dzba_std_arr[li_i].upd_ver,
                               "   AND dzbb002='",dzba_std_arr[li_i].targetid,"'",
                               "   AND dzbb004='s'"               
            LET ls_sql = "SELECT COUNT(1) FROM dzbb_t",ls_ind_bb_wc
            #display "sadzp060_4_sync_addpoint:select ind dzbb_t sql=",ls_sql
            PREPARE dzbb_prep1 FROM ls_sql
            EXECUTE dzbb_prep1 INTO li_cnt
            FREE dzbb_prep1
            
            IF li_cnt>0 THEN #資料已存在.
               #更新dzbb005
               LET ls_sql = "UPDATE dzbb_t",
                            "   SET dzbbmoddt=?,", 
                            "       dzbbmodid='",g_user,"' ",ls_ind_bb_wc
            ELSE
               LET ls_sql = "INSERT INTO dzbb_t(dzbb001,dzbb002,dzbb003,dzbb004,dzbb005,dzbb006,dzbb007,",
                            "                   dzbbcrtdt,dzbbcrtdp,dzbbowndp,dzbbownid,dzbbcrtid)",
                            "            VALUES('",g_ind_prog,"','",dzba_std_arr[li_i].targetid,"',",dzba_std_arr[li_i].upd_ver,",'s','",gs_cust,"','','',", #170519-00055:ls_targetid
                            "                   ?,'",gs_dept,"','",gs_dept,"','",g_user,"','",g_user,"')"
            END IF
            #display "sadzp060_4_sync_addpoint:modify ind dzbb_t sql=",ls_sql
            PREPARE dzbb_prep2 FROM ls_sql
            EXECUTE dzbb_prep2 USING g_date
            FREE dzbb_prep2
            LET ls_sql = NULL #20160118 by Hiko
             
            #更新程式代碼資料:上面就算是INSERT的動作,執行到此也是UPDATE.
            LET ls_tracing = "sadzp060_4_sync_addpoint : update dzbb_t data : dzbb098"

            UPDATE dzbb_t 
               SET dzbb098=dzba_std_arr[li_i].upd_apt
             WHERE dzbb001=g_ind_prog
               AND dzbb002=dzba_std_arr[li_i].targetid
               AND dzbb003=dzba_std_arr[li_i].upd_ver
               AND dzbb004='s' 
         END IF #ps_act.equals("SYNC")

         FREE dzba_std_arr[li_i].upd_apt #170519-00055

         LET li_i = li_i + 1
         LOCATE dzba_std_arr[li_i].upd_apt IN FILE #170519-00055
      END FOREACH
      
      DISPLAY "[sadzp060_4]總共同步 ",(li_i-1)," 筆程式資料"

      #Begin:20161118 by Hiko : 最後再將標準dzba_t的資料寫入dzbu_t內.
      LET ls_tracing = "sadzp060_4_sync_addpoint : delete dzbu_t"
      DELETE FROM dzbu_t WHERE dzbu001=g_ind_prog AND dzbu002=g_ind_cver
      
      LET ls_tracing = "sadzp060_4_sync_addpoint : insert dzbu_t"
      LET ls_sql = "INSERT INTO dzbu_t(dzbu001,dzbu002,dzbu003,dzbu004,dzbu005)",
                   " SELECT '",g_ind_prog,"',",g_ind_cver,",ba1.dzba003,ba1.dzba004,ba1.dzba002 FROM dzba_t ba1", #170424-00013:dzbu002,dzbu005
                   "  WHERE ba1.dzba001='",g_std_prog,"'",
                   "    AND ba1.dzba002=",g_std_cver 
      #display "sadzp060_4_sync_addpoint:insert dzbu_t sql=",ls_sql
      PREPARE dzbu_prep2 FROM ls_sql
      EXECUTE dzbu_prep2
      FREE dzbu_prep2
      #End:20161118 by Hiko

      LET ls_tracing = "sadzp060_4_sync_addpoint finish !"
      display ls_tracing

      RETURN TRUE,NULL
   CATCH
      LET ls_tracing = "sadzp060_4_sync_addpoint failure !"
      CALL sadzp060_4_err_catch(ls_tracing, ls_sql)
      RETURN FALSE,"ERROR"||ls_tracing
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE 
# Descriptions    : 同步行業規格:dzac_t
# Input parameter : pr_std_dzaa 標準dzaa_t資料
# Return code     : BOOLEAN TRUE:成功;FALSE:失敗
#                 : STRING 錯誤訊息
# Date & Author   : 2015/01/15 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_4_sync_dzac(pr_std_dzaa)
   DEFINE pr_std_dzaa TYPE_DZAA 
   DEFINE ls_tracing  STRING,
          ls_dzac099  STRING,
          ls_sql      STRING,
          ls_ind_wc   STRING,
          li_cnt      SMALLINT,
          lr_std_dzac RECORD
                      dzac001 LIKE dzac_t.dzac001, 
                      dzac002 LIKE dzac_t.dzac002, 
                      dzac003 LIKE dzac_t.dzac003, 
                      dzac004 LIKE dzac_t.dzac004, 
                      dzac005 LIKE dzac_t.dzac005, 
                      dzac008 LIKE dzac_t.dzac008, 
                      dzac009 LIKE dzac_t.dzac009, 
                      dzac010 LIKE dzac_t.dzac010, 
                      dzac011 LIKE dzac_t.dzac011, 
                      dzac012 LIKE dzac_t.dzac012, 
                      dzac013 LIKE dzac_t.dzac013, 
                      dzac014 LIKE dzac_t.dzac014, 
                      dzac015 LIKE dzac_t.dzac015, 
                      dzac016 LIKE dzac_t.dzac016, 
                      dzac017 LIKE dzac_t.dzac017, 
                      dzac018 LIKE dzac_t.dzac018, 
                      dzac019 LIKE dzac_t.dzac019, 
                      dzac020 LIKE dzac_t.dzac020, 
                      dzac021 LIKE dzac_t.dzac021, 
                      dzac099 LIKE dzac_t.dzac099 
                      END RECORD
   TRY
      LET ls_tracing = "sadzp060_4_sync_dzac start !"
      display ls_tracing

      #170519-00055:曾經有大片註解.

      IF g_type="M" THEN
         LET ls_dzac099 = "dzac099"
      ELSE #S,Q
         LET ls_dzac099 = "REPLACE(dzac099,'",g_std_prog,"','",g_ind_prog,"')"
      END IF

      #取得標準的dzac_t.
      LOCATE lr_std_dzac.dzac099 IN FILE

      LET ls_sql = "SELECT dzac001,dzac002,dzac003,dzac004,dzac005,",
                   "                       dzac008,dzac009,dzac010,",
                   "       dzac011,dzac012,dzac013,dzac014,dzac015,",
                   "       dzac016,dzac017,dzac018,dzac019,dzac020,",
                   "       dzac021,",ls_dzac099,
                   "  FROM dzac_t",
                   " WHERE dzac001='",g_std_prog,"'",
                   "   AND dzac003='",pr_std_dzaa.dzaa003,"'",
                   "   AND dzac004=",pr_std_dzaa.dzaa004,
                   "   AND dzac012='s'"
      #display "sadzp060_4_sync_dzac:select std dzac_t sql=",ls_sql
      PREPARE std_dzac_prep1 FROM ls_sql
      EXECUTE std_dzac_prep1 INTO lr_std_dzac.*
      FREE std_dzac_prep1

      #判斷要更新的行業資料是否存在.      
      LET ls_ind_wc = " WHERE dzac001='",g_ind_prog,"'",
                      "   AND dzac003='",pr_std_dzaa.targetid,"'",
                      "   AND dzac004=",pr_std_dzaa.upd_ver,
                      "   AND dzac012='s'"
      LET ls_sql = "SELECT COUNT(1) FROM dzac_t",ls_ind_wc
      #display "sadzp060_4_sync_dzac:select ind dzac_t sql=",ls_sql
      PREPARE dzac_prep1 FROM ls_sql
      EXECUTE dzac_prep1 INTO li_cnt
      FREE dzac_prep1

      IF li_cnt>0 THEN #資料已存在.
         LET ls_sql = "UPDATE dzac_t",
                      "   SET dzac002='",lr_std_dzac.dzac002 CLIPPED,"',",
                      "       dzac005='",lr_std_dzac.dzac005 CLIPPED,"',",
                      "       dzac008='",lr_std_dzac.dzac008 CLIPPED,"',",
                      "       dzac009='",lr_std_dzac.dzac009 CLIPPED,"',",
                      "       dzac010='",lr_std_dzac.dzac010 CLIPPED,"',",
                      "       dzac011='",lr_std_dzac.dzac011 CLIPPED,"',",
                      "       dzac013='",lr_std_dzac.dzac013 CLIPPED,"',",
                      "       dzac014=?,", #預設值的部分就已資料庫的資料為主
                      "       dzac015='",lr_std_dzac.dzac015 CLIPPED,"',",
                      "       dzac016='",lr_std_dzac.dzac016 CLIPPED,"',",
                      "       dzac017='",lr_std_dzac.dzac017 CLIPPED,"',",
                      "       dzac018='",lr_std_dzac.dzac018 CLIPPED,"',",
                      "       dzac019='",lr_std_dzac.dzac019 CLIPPED,"',",
                      "       dzac020='",lr_std_dzac.dzac020 CLIPPED,"',",
                      "       dzac021='",lr_std_dzac.dzac021 CLIPPED,"',",
                      "       dzacmoddt=?,",
                      "       dzacmodid='",g_user,"' ",ls_ind_wc
      ELSE
         LET ls_sql = "INSERT INTO dzac_t(dzac001,dzac002,dzac003,dzac004,dzac005,",
                      "                                   dzac008,dzac009,dzac010,",
                      "                   dzac011,dzac012,dzac013,dzac014,dzac015,",
                      "                   dzac016,dzac017,dzac018,dzac019,dzac020,",
                      "                   dzac021,dzacstus,", #170801-00004:dzacstus
                      "                   dzaccrtdt,dzaccrtdp,dzacowndp,dzacownid,dzaccrtid)",
                      "            VALUES('",g_ind_prog,"','",lr_std_dzac.dzac002 CLIPPED,"','",pr_std_dzaa.targetid,"',",pr_std_dzaa.upd_ver,",'",lr_std_dzac.dzac005 CLIPPED,"',",
                      "                   '",lr_std_dzac.dzac008 CLIPPED,"','",lr_std_dzac.dzac009 CLIPPED,"','",lr_std_dzac.dzac010 CLIPPED,"',", 
                      "                   '",lr_std_dzac.dzac011 CLIPPED,"','s','",lr_std_dzac.dzac013 CLIPPED,"',?,'",lr_std_dzac.dzac015 CLIPPED,"',", 
                      "                   '",lr_std_dzac.dzac016 CLIPPED,"','",lr_std_dzac.dzac017 CLIPPED,"','",lr_std_dzac.dzac018 CLIPPED,"','",lr_std_dzac.dzac019 CLIPPED,"','",lr_std_dzac.dzac020 CLIPPED,"',", 
                      "                   '",lr_std_dzac.dzac021 CLIPPED,"','Y',", 
                      "                   ?,'",gs_dept,"','",gs_dept,"','",g_user,"','",g_user,"')"
      END IF
      #display "sadzp060_4_sync_dzac:modify ind dzac_t sql=",ls_sql
      PREPARE dzac_prep2 FROM ls_sql
      EXECUTE dzac_prep2 USING lr_std_dzac.dzac014,g_date
      FREE dzac_prep2
      LET ls_sql = NULL #20160118 by Hiko
      
      #更新規格描述資料:上面就算是INSERT的動作,執行到此也是UPDATE.
      LET ls_tracing = "sadzp060_4_sync_dzac : update dzac099"

      UPDATE dzac_t 
         SET dzac099=lr_std_dzac.dzac099
       WHERE dzac001=g_ind_prog
         AND dzac003=pr_std_dzaa.targetid
         AND dzac004=pr_std_dzaa.upd_ver
         AND dzac012='s' 

      FREE lr_std_dzac.dzac099

      LET ls_tracing = "sadzp060_4_sync_dzac finish !"
      display ls_tracing

      RETURN TRUE,NULL
   CATCH
      LET ls_tracing = "sadzp060_4_sync_dzac failure !"
      CALL sadzp060_4_err_catch(ls_tracing, ls_sql)
      RETURN FALSE,"ERROR:"||ls_tracing
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE 
# Descriptions    : 同步行業規格:dzab_t
# Input parameter : pr_std_dzaa 標準dzaa_t資料
# Return code     : BOOLEAN TRUE:成功;FALSE:失敗
#                 : STRING 錯誤訊息
# Date & Author   : 2015/01/15 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_4_sync_dzab(pr_std_dzaa)
   DEFINE pr_std_dzaa TYPE_DZAA 
   DEFINE ls_tracing  STRING,
          ls_dzab099  STRING,
          ls_sql      STRING,
          ls_ind_wc   STRING,
          li_cnt      SMALLINT,
          lr_std_dzab RECORD
                      dzab001 LIKE dzab_t.dzab001, 
                      dzab002 LIKE dzab_t.dzab002, 
                      dzab003 LIKE dzab_t.dzab003, 
                      dzab004 LIKE dzab_t.dzab004, 
                      dzab005 LIKE dzab_t.dzab005, 
                      dzab099 LIKE dzab_t.dzab099 
                      END RECORD

   TRY
      LET ls_tracing = "sadzp060_4_sync_dzab start !"
      display ls_tracing

      #170519-00055:曾經有大片註解.

      IF g_type="M" THEN
         LET ls_dzab099 = "dzab099"
      ELSE #S,Q
         LET ls_dzab099 = "REPLACE(dzab099,'",g_std_prog,"','",g_ind_prog,"')"
      END IF

      #取得標準的dzab_t.
      LOCATE lr_std_dzab.dzab099 IN FILE

      LET ls_sql = "SELECT dzab001,dzab002,dzab003,dzab004,dzab005,",
                   "      ",ls_dzab099,
                   "  FROM dzab_t",
                   " WHERE dzab001='",g_std_prog,"'",
                   "   AND dzab004='",pr_std_dzaa.dzaa003,"'",
                   "   AND dzab002=",pr_std_dzaa.dzaa004,
                   "   AND dzab003='s'" 
      #display "sadzp060_4_sync_dzab:select std dzab_t sql=",ls_sql
      PREPARE std_dzab_prep1 FROM ls_sql
      EXECUTE std_dzab_prep1 INTO lr_std_dzab.*
      FREE std_dzab_prep1

      #判斷要更新的行業資料是否存在.
      LET ls_ind_wc = " WHERE dzab001='",g_ind_prog,"'",
                      "   AND dzab004='",pr_std_dzaa.targetid,"'",
                      "   AND dzab002=",pr_std_dzaa.upd_ver,
                      "   AND dzab003='s'" 
      LET ls_sql = "SELECT COUNT(1) FROM dzab_t",ls_ind_wc
      #display "sadzp060_4_sync_dzab:select ind dzab_t sql=",ls_sql
      PREPARE dzab_prep1 FROM ls_sql
      EXECUTE dzab_prep1 INTO li_cnt
      FREE dzab_prep1

      IF li_cnt>0 THEN #資料已存在.
         LET ls_sql = "UPDATE dzab_t",
                        " SET dzabmoddt=?,",
                             "dzabmodid='",g_user,"' ",ls_ind_wc
      ELSE
         LET ls_sql = "INSERT INTO dzab_t(dzab001,dzab002,dzab003,dzab004,dzab005,dzabstus,", #170801-00004:dzabstus
                      "                   dzabcrtdt,dzabcrtdp,dzabowndp,dzabownid,dzabcrtid)",
                      "            VALUES('",g_ind_prog,"',",pr_std_dzaa.upd_ver,",'s','",pr_std_dzaa.targetid,"','",lr_std_dzab.dzab005 CLIPPED,"','Y',",
                      "                   ?,'",gs_dept,"','",gs_dept,"','",g_user,"','",g_user,"')"
      END IF
      #display "sadzp060_4_sync_dzab:modify ind dzab_t sql=",ls_sql
      PREPARE dzab_prep2 FROM ls_sql
      EXECUTE dzab_prep2 USING g_date
      FREE dzab_prep2
      LET ls_sql = NULL #20160118 by Hiko
      
      #更新規格描述資料:上面就算是INSERT的動作,執行到此也是UPDATE.
      LET ls_tracing = "sadzp060_4_sync_dzab : update dzab099"

      UPDATE dzab_t 
         SET dzab099=lr_std_dzab.dzab099
       WHERE dzab001=g_ind_prog
         AND dzab002=pr_std_dzaa.upd_ver
         AND dzab003='s' 
         AND dzab004=pr_std_dzaa.targetid

      FREE lr_std_dzab.dzab099

      LET ls_tracing = "sadzp060_4_sync_dzab finish !"
      display ls_tracing

      RETURN TRUE,NULL
   CATCH
      LET ls_tracing = "sadzp060_4_sync_dzab failure !"
      CALL sadzp060_4_err_catch(ls_tracing, ls_sql)
      RETURN FALSE,"ERROR:"||ls_tracing
   END TRY
END FUNCTION
 
##########################################################################
# Access Modifier : PRIVATE 
# Descriptions    : 同步行業規格:dzad_t
# Input parameter : pr_std_dzaa 標準dzaa_t資料
# Return code     : BOOLEAN TRUE:成功;FALSE:失敗
#                 : STRING 錯誤訊息
# Date & Author   : 2015/01/15 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_4_sync_dzad(pr_std_dzaa)
   DEFINE pr_std_dzaa TYPE_DZAA 
   DEFINE ls_tracing  STRING,
          #ls_dzad002  STRING, #targetid
          ls_dzad099  STRING,
          ls_sql      STRING,
          ls_ind_wc   STRING,
          li_cnt      SMALLINT,
          lr_std_dzad RECORD
                      dzad001 LIKE dzad_t.dzad001,
                      dzad002 LIKE dzad_t.dzad002,
                      dzad003 LIKE dzad_t.dzad003,
                      dzad005 LIKE dzad_t.dzad005,
                      dzad006 LIKE dzad_t.dzad006,
                      dzad007 LIKE dzad_t.dzad007,
                      dzad008 LIKE dzad_t.dzad008,
                      dzad099 LIKE dzad_t.dzad099
                      END RECORD

   TRY
      LET ls_tracing = "sadzp060_4_sync_dzad start !"
      display ls_tracing

      #170519-00055:曾經有大片註解.

      IF g_type="M" THEN
         LET ls_dzad099 = "dzad099"
      ELSE #S,Q
         LET ls_dzad099 = "REPLACE(dzad099,'",g_std_prog,"','",g_ind_prog,"')"
      END IF

      #取得標準的dzad_t.
      LOCATE lr_std_dzad.dzad099 IN FILE

      LET ls_sql = "SELECT dzad001,dzad002,dzad003,dzad005,",
                   "       dzad006,dzad007,dzad008,",ls_dzad099,
                   "  FROM dzad_t",
                   " WHERE dzad001='",g_std_prog,"'",
                   "   AND dzad002='",pr_std_dzaa.dzaa003,"'",
                   "   AND dzad003=",pr_std_dzaa.dzaa004,
                   "   AND dzad005='s'"
      #display "sadzp060_4_sync_dzad:select std dzad_t sql=",ls_sql
      PREPARE std_dzad_prep1 FROM ls_sql
      EXECUTE std_dzad_prep1 INTO lr_std_dzad.*
      FREE std_dzad_prep1

      #判斷要更新的行業資料是否存在.      
      LET ls_ind_wc = " WHERE dzad001='",g_ind_prog,"'",
                      "   AND dzad002='",pr_std_dzaa.targetid,"'",
                      "   AND dzad003=",pr_std_dzaa.upd_ver,
                      "   AND dzad005='s'"
      LET ls_sql = "SELECT COUNT(1) FROM dzad_t",ls_ind_wc
      #display "sadzp060_4_sync_dzad:select ind dzad_t sql=",ls_sql
      PREPARE dzad_prep1 FROM ls_sql
      EXECUTE dzad_prep1 INTO li_cnt
      FREE dzad_prep1

      IF li_cnt>0 THEN #資料已存在.
         LET ls_sql = "UPDATE dzad_t",
                      "   SET dzad006='",lr_std_dzad.dzad006 CLIPPED,"',",
                      "       dzad007='",lr_std_dzad.dzad007 CLIPPED,"',",
                      "       dzadmoddt=?,",
                      "       dzadmodid='",g_user,"' ",ls_ind_wc
      ELSE
         LET ls_sql = "INSERT INTO dzad_t(dzad001,dzad002,dzad003,dzad005,",
                      "                   dzad006,dzad007,dzad008,dzadstus,", #170801-00004:dzadstus
                      "                   dzadcrtdt,dzadcrtdp,dzadowndp,dzadownid,dzadcrtid)",
                      "            VALUES('",g_ind_prog,"','",pr_std_dzaa.targetid,"',",pr_std_dzaa.upd_ver,",'s',",
                      "                   '",lr_std_dzad.dzad006 CLIPPED,"','",lr_std_dzad.dzad007 CLIPPED,"','",lr_std_dzad.dzad008 CLIPPED,"','Y',", 
                      "                   ?,'",gs_dept,"','",gs_dept,"','",g_user,"','",g_user,"')"
      END IF
      #display "sadzp060_4_sync_dzad:modify ind dzad_t sql=",ls_sql
      PREPARE dzad_prep2 FROM ls_sql
      EXECUTE dzad_prep2 USING g_date
      FREE dzad_prep2
      LET ls_sql = NULL #20160118 by Hiko
      
      #更新規格描述資料:上面就算是INSERT的動作,執行到此也是UPDATE.
      LET ls_tracing = "sadzp060_4_sync_dzad : update dzad099"

      UPDATE dzad_t 
         SET dzad099=lr_std_dzad.dzad099
       WHERE dzad001=g_ind_prog
         AND dzad002=pr_std_dzaa.targetid
         AND dzad003=pr_std_dzaa.upd_ver 
         AND dzad005='s'

      FREE lr_std_dzad.dzad099

      LET ls_tracing = "sadzp060_4_sync_dzad finish !"
      display ls_tracing

      RETURN TRUE,NULL
   CATCH
      LET ls_tracing = "sadzp060_4_sync_dzad failure !"
      CALL sadzp060_4_err_catch(ls_tracing, ls_sql)
      RETURN FALSE,"ERROR:"||ls_tracing
   END TRY
END FUNCTION
 
##########################################################################
# Access Modifier : PRIVATE 
# Descriptions    : 同步行業規格:dzak_t
# Input parameter : pr_std_dzaa 標準dzaa_t資料
# Return code     : BOOLEAN TRUE:成功;FALSE:失敗
#                 : STRING 錯誤訊息
# Date & Author   : 2015/01/15 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_4_sync_dzak(pr_std_dzaa)
   DEFINE pr_std_dzaa TYPE_DZAA 
   DEFINE ls_tracing  STRING,
          ls_dzak011  STRING,
          ls_sql      STRING,
          ls_ind_wc   STRING,
          li_cnt      SMALLINT,
          lr_std_dzak RECORD
                      dzak001 LIKE dzak_t.dzak001,
                      dzak002 LIKE dzak_t.dzak002,
                      dzak003 LIKE dzak_t.dzak003,
                      dzak004 LIKE dzak_t.dzak004,
                      dzak005 LIKE dzak_t.dzak005,
                      dzak007 LIKE dzak_t.dzak007,
                      dzak008 LIKE dzak_t.dzak008,
                      dzak009 LIKE dzak_t.dzak009,
                      dzak010 LIKE dzak_t.dzak010,
                      dzak011 LIKE dzak_t.dzak011,
                      dzak012 LIKE dzak_t.dzak012 
                      END RECORD 

   TRY
      LET ls_tracing = "sadzp060_4_sync_dzak start !"
      display ls_tracing

      #170519-00055:曾經有大片註解.

      IF g_type="M" THEN
         LET ls_dzak011 = "dzak011"
      ELSE #S,Q
         LET ls_dzak011 = "REPLACE(dzak011,'",g_std_prog,"','",g_ind_prog,"')"
      END IF

      #取得標準的dzak_t.
      LET ls_sql = "SELECT dzak001,dzak002,dzak003,dzak004,dzak005,",
                   "               dzak007,dzak008,dzak009,dzak010,",
                   "      ",ls_dzak011,",dzak012",
                   "  FROM dzak_t",
                   " WHERE dzak001='",g_std_prog,"'",
                   "   AND dzak002='",pr_std_dzaa.dzaa003,"'",
                   "   AND dzak003=",pr_std_dzaa.dzaa004,     
                   "   AND dzak004='s'" 
      #display "sadzp060_4_sync_dzak:select std dzak_t sql=",ls_sql
      PREPARE std_dzak_prep1 FROM ls_sql
      EXECUTE std_dzak_prep1 INTO lr_std_dzak.*
      FREE std_dzak_prep1

      IF NOT cl_null(lr_std_dzak.dzak002) THEN
         #判斷要更新的行業資料是否存在.
         LET ls_ind_wc = " WHERE dzak001='",g_ind_prog,"'",
                         "   AND dzak002='",pr_std_dzaa.targetid,"'", 
                         "   AND dzak003=",pr_std_dzaa.upd_ver,      
                         "   AND dzak004='s'" 
         LET ls_sql = "SELECT COUNT(1) FROM dzak_t",ls_ind_wc
         #display "sadzp060_4_sync_dzak:select ind dzak_t sql=",ls_sql
         PREPARE dzak_prep1 FROM ls_sql
         EXECUTE dzak_prep1 INTO li_cnt
         FREE dzak_prep1
         
         IF li_cnt>0 THEN #資料已存在.
            LET ls_sql = "UPDATE dzak_t",
                         "   SET dzak005=?,",
                         "       dzak007='",lr_std_dzak.dzak007 CLIPPED,"',",
                         "       dzak008='",lr_std_dzak.dzak008 CLIPPED,"',",
                         "       dzak009='",lr_std_dzak.dzak009 CLIPPED,"',",
                         "       dzak010='",lr_std_dzak.dzak010 CLIPPED,"',",
                         "       dzak011='",lr_std_dzak.dzak011 CLIPPED,"',",
                         "       dzakmoddt=?,",
                         "       dzakmodid='",g_user,"' ",ls_ind_wc
         ELSE
            LET ls_sql = "INSERT INTO dzak_t(dzak001,dzak002,dzak003,dzak004,dzak005,",
                         "                           dzak007,dzak008,dzak009,dzak010,",
                         "                   dzak011,dzak012,dzakstus,", #170801-00004:dzakstus
                         "                   dzakcrtdt,dzakcrtdp,dzakowndp,dzakownid,dzakcrtid)",
                         "            VALUES('",g_ind_prog,"','",pr_std_dzaa.targetid,"',",pr_std_dzaa.upd_ver,",'s',?,",
                         "                   '",lr_std_dzak.dzak007 CLIPPED,"','",lr_std_dzak.dzak008 CLIPPED,"','",lr_std_dzak.dzak009 CLIPPED,"','",lr_std_dzak.dzak010 CLIPPED,"',", 
                         "                   '",lr_std_dzak.dzak011 CLIPPED,"','",lr_std_dzak.dzak012 CLIPPED,"','Y',", #170921-00007:最後補上逗號.
                         "                   ?,'",gs_dept,"','",gs_dept,"','",g_user,"','",g_user,"')"
         END IF
         #display "sadzp060_4_sync_dzak:modify ind dzak_t sql=",ls_sql
         PREPARE dzak_prep2 FROM ls_sql
         EXECUTE dzak_prep2 USING lr_std_dzak.dzak005,g_date
         FREE dzak_prep2
      END IF

      LET ls_tracing = "sadzp060_4_sync_dzak finish !"
      display ls_tracing

      RETURN TRUE,NULL
   CATCH
      LET ls_tracing = "sadzp060_4_sync_dzak failure !"
      CALL sadzp060_4_err_catch(ls_tracing, ls_sql)
      RETURN FALSE,"ERROR:"||ls_tracing
   END TRY
END FUNCTION
 
##########################################################################
# Access Modifier : PRIVATE 
# Descriptions    : 同步行業規格:dzal_t
# Input parameter : pr_std_dzaa 標準dzaa_t資料
# Return code     : BOOLEAN TRUE:成功;FALSE:失敗
#                 : STRING 錯誤訊息
# Date & Author   : 2015/01/15 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_4_sync_dzal(pr_std_dzaa)
   DEFINE pr_std_dzaa TYPE_DZAA 
   DEFINE ls_tracing STRING,
          ls_dzal002 STRING,
          ls_sql     STRING,
          li_cnt     SMALLINT

   TRY
      LET ls_tracing = "sadzp060_4_sync_dzal start !"
      display ls_tracing

      #170519-00055:曾經有大片註解.

      #判斷標準的dzal_t是否存在:有資料才更新.
      LET ls_sql = "SELECT COUNT(1)",
                   "  FROM dzal_t",
                   " WHERE dzal001='",g_std_prog,"'",
                   "   AND dzal005='",pr_std_dzaa.dzaa003,"'",
                   "   AND dzal003=",pr_std_dzaa.dzaa004,     
                   "   AND dzal004='s'" 
      #display "sadzp060_4_sync_dzal:select std dzal_t sql=",ls_sql
      PREPARE std_dzal_prep1 FROM ls_sql
      EXECUTE std_dzal_prep1 INTO li_cnt
      FREE std_dzal_prep1

      IF li_cnt>0 THEN
         #一對多的關係採用先刪除後新增會比較快.
         LET ls_sql = "DELETE FROM dzal_t",
                      " WHERE dzal001='",g_ind_prog,"'",         
                      "   AND dzal005='",pr_std_dzaa.targetid,"'", #並非使用dzal002(PK).
                      "   AND dzal003=",pr_std_dzaa.upd_ver,     
                      "   AND dzal004='s'"                      
         #display "sadzp060_4_sync_dzal:delete ind dzal_t sql=",ls_sql
         PREPARE del_dzal_prep0 FROM ls_sql
         EXECUTE del_dzal_prep0
         FREE del_dzal_prep0
         
         IF g_type="M" THEN
            LET ls_dzal002 = "dzal002"
         ELSE #S,Q
            LET ls_dzal002 = "REPLACE(dzal002,'",g_std_prog,"','",g_ind_prog,"')"
         END IF

         LET ls_sql = "INSERT INTO dzal_t(dzal001,dzal002,dzal003,dzal004,dzal005,",
                      "                   dzal006,dzal007,dzal008,",
                      "                   dzalcrtdt,dzalcrtdp,dzalowndp,dzalownid,dzalstus,dzalcrtid)",
                      "            SELECT '",g_ind_prog,"',",ls_dzal002,",", #dzal002(控件編號)
                                         pr_std_dzaa.upd_ver,",'s','",pr_std_dzaa.targetid,"',", #dzal005(依附控件編號)
                      "                   dzal006,dzal007,dzal008,",                             #dzal006(程式參考設定):不需要換程式編號.
                      "                   ?,'",gs_dept,"','",gs_dept,"','",g_user,"',dzalstus,'",g_user,"'",
                      "              FROM dzal_t",
                      "             WHERE dzal001='",g_std_prog,"'",          
                      "               AND dzal003=",pr_std_dzaa.dzaa004,      
                      "               AND dzal004='s'",                       
                      "               AND dzal005='",pr_std_dzaa.dzaa003,"'" 
         #display "sadzp060_4_sync_dzal:insert ind dzal_t sql=",ls_sql
         PREPARE dzal_prep8 FROM ls_sql
         EXECUTE dzal_prep8 USING g_date
         FREE dzal_prep8
      END IF

      LET ls_tracing = "sadzp060_4_sync_dzal finish !"
      display ls_tracing

      RETURN TRUE,NULL
   CATCH
      LET ls_tracing = "sadzp060_4_sync_dzal failure !"
      CALL sadzp060_4_err_catch(ls_tracing, ls_sql)
      RETURN FALSE,"ERROR:"||ls_tracing
   END TRY
END FUNCTION
 
##########################################################################
# Access Modifier : PRIVATE 
# Descriptions    : 同步行業規格:dzff_t
# Input parameter : pr_std_dzaa 標準dzaa_t資料
# Return code     : BOOLEAN TRUE:成功;FALSE:失敗
#                 : STRING 錯誤訊息
# Date & Author   : 2015/01/15 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_4_sync_dzff(pr_std_dzaa)
   DEFINE pr_std_dzaa TYPE_DZAA 
   DEFINE ls_tracing  STRING,
          ls_sql      STRING 

   TRY
      LET ls_tracing = "sadzp060_4_sync_dzff start !"
      display ls_tracing

      #執行到這邊就表示有標準資料,則先將目前版次的設計資料清除乾淨,這樣同步結果才會正確.
      LET ls_sql = "DELETE FROM dzff_t",
                   " WHERE dzff001='",g_ind_prog,"'",
                   "   AND dzff002=",pr_std_dzaa.upd_ver,
                   "   AND dzff003='",pr_std_dzaa.targetid,"'",
                   "   AND dzff008='s'"
      #display "sadzp060_4_sync_dzff:delete ind dzff_t sql=",ls_sql
      PREPARE del_dzff_prep0 FROM ls_sql
      EXECUTE del_dzff_prep0
      FREE del_dzff_prep0

      LET ls_sql = "INSERT INTO dzff_t(dzff001,dzff002,dzff003,dzff004,dzff005,",
                   "                   dzff006,dzff007,dzff008,dzff009,",
                   "                   dzffcrtdt,dzffcrtdp,dzffowndp,dzffownid,dzffstus,dzffcrtid)",
                   "            SELECT '",g_ind_prog,"',",pr_std_dzaa.upd_ver,",'",pr_std_dzaa.targetid,"',dzff004,dzff005,",
                   "                   dzff006,dzff007,dzff008,dzff009,",
                   "                   ?,'",gs_dept,"','",gs_dept,"','",g_user,"',dzffstus,'",g_user,"'",
                   "              FROM dzff_t",
                   "             WHERE dzff001='",g_std_prog,"'",
                   "               AND dzff002=",pr_std_dzaa.dzaa004,
                   "               AND dzff003='",pr_std_dzaa.dzaa003,"'",
                   "               AND dzff008='s'" 
      #display "sadzp060_4_sync_dzff:insert ind dzff_t sql=",ls_sql 
      PREPARE dzff_prep8 FROM ls_sql
      EXECUTE dzff_prep8 USING g_date
      FREE dzff_prep8

      LET ls_tracing = "sadzp060_4_sync_dzff finish !"
      display ls_tracing

      RETURN TRUE,NULL
   CATCH
      LET ls_tracing = "sadzp060_4_sync_dzff failure !"
      CALL sadzp060_4_err_catch(ls_tracing, ls_sql)
      RETURN FALSE,"ERROR:"||ls_tracing
   END TRY
END FUNCTION
 
##########################################################################
# Access Modifier : PRIVATE 
# Descriptions    : 同步行業規格:dzai_t
# Input parameter : pr_std_dzaa 標準dzaa_t資料
# Return code     : BOOLEAN TRUE:成功;FALSE:失敗
#                 : STRING 錯誤訊息
# Date & Author   : 2015/01/15 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_4_sync_dzai(pr_std_dzaa)
   DEFINE pr_std_dzaa TYPE_DZAA 
   DEFINE ls_tracing  STRING,
          ls_dzai005  STRING,
          ls_sql      STRING,
          ls_ind_wc   STRING,
          li_cnt      SMALLINT,
          lr_std_dzai RECORD
                      dzai001 LIKE dzai_t.dzai001,
                      dzai002 LIKE dzai_t.dzai002,
                      dzai003 LIKE dzai_t.dzai003,
                      dzai004 LIKE dzai_t.dzai004,
                      dzai005 LIKE dzai_t.dzai005,
                      dzai007 LIKE dzai_t.dzai007,
                      dzai008 LIKE dzai_t.dzai008,
                      dzai009 LIKE dzai_t.dzai009,
                      dzai010 LIKE dzai_t.dzai010,
                      dzai011 LIKE dzai_t.dzai011,
                      dzai012 LIKE dzai_t.dzai012  
                      END RECORD

   TRY
      LET ls_tracing = "sadzp060_4_sync_dzai start !"
      display ls_tracing

      #170519-00055:曾經有大片註解.

      IF g_type="M" THEN
         LET ls_dzai005 = "dzai005"
      ELSE #S,Q
         LET ls_dzai005 = "REPLACE(dzai005,'",g_std_prog,"','",g_ind_prog,"')"
      END IF

      #取得標準的dzai_t.
      LET ls_sql = "SELECT dzai001,dzai002,dzai003,dzai004,",ls_dzai005,",",
                   "               dzai007,dzai008,dzai009,dzai010,",
                   "       dzai011,dzai012",
                   "  FROM dzai_t",
                   " WHERE dzai001='",g_std_prog,"'",
                   "   AND dzai002='",pr_std_dzaa.dzaa003,"'",
                   "   AND dzai003=",pr_std_dzaa.dzaa004,
                   "   AND dzai004='s'"
      #display "sadzp060_4_sync_dzai:select std dzai_t sql=",ls_sql 
      PREPARE std_dzai_prep1 FROM ls_sql
      EXECUTE std_dzai_prep1 INTO lr_std_dzai.*
      FREE std_dzai_prep1

      #判斷要更新的行業資料是否存在.      
      LET ls_ind_wc = " WHERE dzai001='",g_ind_prog,"'",
                      "   AND dzai002='",pr_std_dzaa.targetid,"'",
                      "   AND dzai003=",pr_std_dzaa.upd_ver,
                      "   AND dzai004='s'"
      LET ls_sql = "SELECT COUNT(1) FROM dzai_t",ls_ind_wc
      #display "sadzp060_4_sync_dzai:select ind dzai_t sql=",ls_sql
      PREPARE dzai_prep1 FROM ls_sql
      EXECUTE dzai_prep1 INTO li_cnt
      FREE dzai_prep1

      IF li_cnt>0 THEN #資料已存在.
         LET ls_sql = "UPDATE dzai_t",
                      "   SET dzai005='",lr_std_dzai.dzai005 CLIPPED,"',",
                      "       dzai007=?,",
                      "       dzai008='",lr_std_dzai.dzai008 CLIPPED,"',",
                      "       dzai009='",lr_std_dzai.dzai009 CLIPPED,"',",
                      "       dzai010='",lr_std_dzai.dzai010 CLIPPED,"',",
                      "       dzai011='",lr_std_dzai.dzai011 CLIPPED,"',",
                      "       dzaimoddt=?,",
                      "       dzaimodid='",g_user,"' ",ls_ind_wc
      ELSE
         LET ls_sql = "INSERT INTO dzai_t(dzai001,dzai002,dzai003,dzai004,dzai005,",
                      "                           dzai007,dzai008,dzai009,dzai010,",
                      "                   dzai011,dzai012,dzaistus,", #170801-00004:dzaistus
                      "                   dzaicrtdt,dzaicrtdp,dzaiowndp,dzaiownid,dzaicrtid)",
                      "            VALUES('",g_ind_prog,"','",pr_std_dzaa.targetid,"',",pr_std_dzaa.upd_ver,",'s','",lr_std_dzai.dzai005 CLIPPED,"',",
                      "                                 ?,'",lr_std_dzai.dzai008 CLIPPED,"','",lr_std_dzai.dzai009 CLIPPED,"','",lr_std_dzai.dzai010 CLIPPED,"',", 
                      "                   '",lr_std_dzai.dzai011 CLIPPED,"','",lr_std_dzai.dzai012 CLIPPED,"','Y',", 
                      "                   ?,'",gs_dept,"','",gs_dept,"','",g_user,"','",g_user,"')"
      END IF
      #display "sadzp060_4_sync_dzai:modify ind dzai_t sql=",ls_sql 
      PREPARE dzai_prep2 FROM ls_sql
      EXECUTE dzai_prep2 USING lr_std_dzai.dzai007,g_date
      FREE dzai_prep2

      LET ls_tracing = "sadzp060_4_sync_dzai finish !"
      display ls_tracing

      RETURN TRUE,NULL
   CATCH
      LET ls_tracing = "sadzp060_4_sync_dzai failure !"
      CALL sadzp060_4_err_catch(ls_tracing, ls_sql)
      RETURN FALSE,"ERROR:"||ls_tracing
   END TRY
END FUNCTION
 
##########################################################################
# Access Modifier : PRIVATE 
# Descriptions    : 同步行業規格:dzaj_t
# Input parameter : pr_std_dzaa 標準dzaa_t資料
# Return code     : BOOLEAN TRUE:成功;FALSE:失敗
#                 : STRING 錯誤訊息
# Date & Author   : 2015/01/15 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_4_sync_dzaj(pr_std_dzaa)
   DEFINE pr_std_dzaa TYPE_DZAA 
   DEFINE ls_tracing  STRING,
          ls_dzaj005  STRING,
          ls_dzaj099  STRING,
          ls_sql      STRING,
          ls_ind_wc   STRING,
          li_cnt      SMALLINT,
          lr_std_dzaj RECORD
                      dzaj001 LIKE dzaj_t.dzaj001,
                      dzaj002 LIKE dzaj_t.dzaj002,
                      dzaj003 LIKE dzaj_t.dzaj003,
                      dzaj004 LIKE dzaj_t.dzaj004,
                      dzaj005 LIKE dzaj_t.dzaj005,
                      dzaj007 LIKE dzaj_t.dzaj007,
                      dzaj008 LIKE dzaj_t.dzaj008,
                      dzaj009 LIKE dzaj_t.dzaj009,
                      dzaj010 LIKE dzaj_t.dzaj010,
                      dzaj011 LIKE dzaj_t.dzaj011,
                      dzaj012 LIKE dzaj_t.dzaj012, 
                      dzaj099 LIKE dzaj_t.dzaj099 
                      END RECORD

   TRY
      LET ls_tracing = "sadzp060_4_sync_dzaj start !"
      display ls_tracing

      #170519-00055:曾經有大片註解.

      IF g_type="M" THEN
         LET ls_dzaj005 = "dzaj005"
         LET ls_dzaj099 = "dzaj099"
      ELSE #S,Q
         LET ls_dzaj005 = "REPLACE(dzaj005,'",g_std_prog,"','",g_ind_prog,"')"
         LET ls_dzaj099 = "REPLACE(dzaj099,'",g_std_prog,"','",g_ind_prog,"')"
      END IF

      #取得標準的dzaj_t.
      LOCATE lr_std_dzaj.dzaj099 IN FILE

      LET ls_sql = "SELECT dzaj001,dzaj002,dzaj003,dzaj004,",ls_dzaj005,",",
                   "               dzaj007,dzaj008,dzaj009,dzaj010,",
                   "       dzaj011,dzaj012,",ls_dzaj099,
                   "  FROM dzaj_t",
                   " WHERE dzaj001='",g_std_prog,"'",
                   "   AND dzaj002='",pr_std_dzaa.dzaa003,"'",
                   "   AND dzaj003=",pr_std_dzaa.dzaa004,
                   "   AND dzaj004='s'"
      #display "sadzp060_4_sync_dzaj:select std dzaj_t sql=",ls_sql
      PREPARE std_dzaj_prep1 FROM ls_sql
      EXECUTE std_dzaj_prep1 INTO lr_std_dzaj.*
      FREE std_dzaj_prep1

      #判斷要更新的行業資料是否存在.      
      LET ls_ind_wc = " WHERE dzaj001='",g_ind_prog,"'",
                      "   AND dzaj002='",pr_std_dzaa.targetid,"'",
                      "   AND dzaj003=",pr_std_dzaa.upd_ver,
                      "   AND dzaj004='s'"
      LET ls_sql = "SELECT COUNT(1) FROM dzaj_t",ls_ind_wc
      #display "sadzp060_4_sync_dzaj:select ind dzaj_t sql=",ls_sql
      PREPARE dzaj_prep1 FROM ls_sql
      EXECUTE dzaj_prep1 INTO li_cnt
      FREE dzaj_prep1

      IF li_cnt>0 THEN #資料已存在.
         LET ls_sql = "UPDATE dzaj_t",
                      "   SET dzaj005='",lr_std_dzaj.dzaj005 CLIPPED,"',",
                      "       dzaj007=?,",
                      "       dzaj008='",lr_std_dzaj.dzaj008 CLIPPED,"',",
                      "       dzaj009='",lr_std_dzaj.dzaj009 CLIPPED,"',",
                      "       dzaj010='",lr_std_dzaj.dzaj010 CLIPPED,"',",
                      "       dzaj011='",lr_std_dzaj.dzaj011 CLIPPED,"',",
                      "       dzajmoddt=?,",
                      "       dzajmodid='",g_user,"' ",ls_ind_wc
      ELSE
         LET ls_sql = "INSERT INTO dzaj_t(dzaj001,dzaj002,dzaj003,dzaj004,dzaj005,",
                      "                           dzaj007,dzaj008,dzaj009,dzaj010,",
                      "                   dzaj011,dzaj012,dzajstus,", #170801-00004:dzajstus
                      "                   dzajcrtdt,dzajcrtdp,dzajowndp,dzajownid,dzajcrtid)",
                      "            VALUES('",g_ind_prog,"','",pr_std_dzaa.targetid,"',",pr_std_dzaa.upd_ver,",'s','",lr_std_dzaj.dzaj005 CLIPPED,"',",
                      "                                 ?,'",lr_std_dzaj.dzaj008 CLIPPED,"','",lr_std_dzaj.dzaj009 CLIPPED,"','",lr_std_dzaj.dzaj010 CLIPPED,"',", 
                      "                   '",lr_std_dzaj.dzaj011 CLIPPED,"','",lr_std_dzaj.dzaj012 CLIPPED,"','Y',", 
                      "                   ?,'",gs_dept,"','",gs_dept,"','",g_user,"','",g_user,"')"
      END IF
      #display "sadzp060_4_sync_dzaj:modify ind dzaj_t sql=",ls_sql
      PREPARE dzaj_prep2 FROM ls_sql
      EXECUTE dzaj_prep2 USING lr_std_dzaj.dzaj007,g_date
      FREE dzaj_prep2
      LET ls_sql = NULL #20160118 by Hiko
      
      #更新規格描述資料:上面就算是INSERT的動作,執行到此也是UPDATE.
      LET ls_tracing = "sadzp060_4_sync_dzaj : update dzaj099"

      UPDATE dzaj_t 
         SET dzaj099=lr_std_dzaj.dzaj099
       WHERE dzaj001=g_ind_prog
         AND dzaj002=pr_std_dzaa.targetid
         AND dzaj003=pr_std_dzaa.upd_ver 
         AND dzaj004='s'

      FREE lr_std_dzaj.dzaj099

      LET ls_tracing = "sadzp060_4_sync_dzaj finish !"
      display ls_tracing

      RETURN TRUE,NULL
   CATCH
      LET ls_tracing = "sadzp060_4_sync_dzaj failure !"
      CALL sadzp060_4_err_catch(ls_tracing, ls_sql)
      RETURN FALSE,"ERROR:"||ls_tracing
   END TRY
END FUNCTION

#20160118 by Hiko:曾經有大片註解:移除FUNCTION sadzp060_4_spec_dzar_t,sadzp060_4_multi_dzar_t

#170519-00055:移除sadzp060_4_prepare_dzag_t,將程式內容移到sadzp060_4_sync_dzag內.

#170519-00055:移除sadzp060_4_prepare_dzam_t,將程式內容移到sadzp060_4_sync_dzam內.

##########################################################################
# Access Modifier : PRIVATE 
# Descriptions    : 同步行業規格:dzag_t/dzfs_t
# Input parameter : pr_std_dzaa 標準dzaa_t資料
#                 : po_prev_ind_dzaf 舊版行業版次資料
# Return code     : BOOLEAN TRUE:成功;FALSE:失敗
#                 : STRING 錯誤訊息
# Date & Author   : 2015/05/13 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_4_sync_dzag(pr_std_dzaa, po_prev_ind_dzaf)
   DEFINE pr_std_dzaa TYPE_DZAA,
          po_prev_ind_dzaf T_DZAF_T
   DEFINE ls_tracing STRING,
          ls_dzfs003 STRING,
          ls_sql     STRING 

   TRY
      LET ls_tracing = "sadzp060_4_sync_dzag start !"
      display ls_tracing

      #要先複製舊版資料.
      IF NOT sadzp060_4_copy_dzag_t(po_prev_ind_dzaf.*) THEN
         RETURN FALSE,"ERROR:call sadzp060_4_copy_dzag_t error!"
      END IF

      #Begin:170703-00069:剔除dzagownid='i'的資料要避開,其他資料要覆蓋. #todo
      #刪除dzagownid標示為'i'的資料.
      LET ls_sql = "DELETE FROM dzag_t",
                   " WHERE dzag001='",g_ind_prog,"'",
                   "   AND dzag003=",g_ind_sver,
                   "   AND NVL(dzagownid,'n')<>'i'"
      display "sadzp060_4_sync_dzag_t:delete ind dzag_t sql=",ls_sql
      PREPARE dzag_prep4 FROM ls_sql
      EXECUTE dzag_prep4
      FREE dzag_prep4
      #End:170703-00069

      #新增標準規格多餘行業規格的TABLE設計資料:dzag_t
      LET ls_sql = "INSERT INTO dzag_t(dzag001,dzag002,dzag003,dzag004,dzag005,",
                   "                   dzag006,dzag007,dzag008,dzag009,dzag010,",
                   "                   dzag011,dzag012,dzag013,dzag014,dzag015,",
                   "                   dzagcrtdt,dzagcrtdp,dzagowndp,dzagownid,dzagstus,dzagcrtid) ",
                   "SELECT '",g_ind_prog,"',ag1.dzag002,",g_ind_sver,",ag1.dzag004,ag1.dzag005,",
                   "       's',ag1.dzag007,ag1.dzag008,ag1.dzag009,ag1.dzag010,",     
                   "       ag1.dzag011,ag1.dzag012,ag1.dzag013,ag1.dzag014,ag1.dzag015,",     
                   "       ?,'",gs_dept,"','",gs_dept,"','",g_user,"',ag1.dzagstus,'",g_user,"'",
                   "  FROM dzag_t ag1",
                   " WHERE ag1.dzag001='",g_std_prog,"'",
                   "   AND ag1.dzag002 NOT IN(SELECT ag2.dzag002 FROM dzag_t ag2",
                   "                           WHERE ag2.dzag001='",g_ind_prog,"'",
                   "                             AND ag2.dzag003=(SELECT dzaa004 FROM dzaa_t",
                   "                                               WHERE dzaa001='",g_ind_prog,"'",
                   "                                                 AND dzaa002=",g_ind_sver,
                   "                                                 AND dzaa003='TABLE'",
                   "                                                 AND dzaa009='s')", 
                   "                             AND ag2.dzag006='s')",
                   "   AND ag1.dzag003=",pr_std_dzaa.dzaa004,
                   "   AND ag1.dzag006='s'"
      #display "sadzp060_4_sync_dzag_t:insert ind dzag_t sql=",ls_sql
      PREPARE dzag_prep2 FROM ls_sql
      EXECUTE dzag_prep2 USING g_date
      FREE dzag_prep2

      #Begin:170703-00069:剔除dzfsownid='i'的資料要避開,其他資料要覆蓋. #todo
      #刪除dzfsownid標示為'i'的資料.
      LET ls_sql = "DELETE FROM dzfs_t",
                   " WHERE dzfs002='",g_ind_prog,"'",
                   "   AND dzfs001=",g_ind_sver,
                   "   AND NVL(dzfsownid,'n')<>'i'"
      display "sadzp060_4_sync_dzfs_t:delete ind dzfs_t sql=",ls_sql
      PREPARE dzfs_prep4 FROM ls_sql
      EXECUTE dzfs_prep4
      FREE dzfs_prep4
      #End:170703-00069

      #新增標準規格多餘行業規格的TABLE設計資料:dzfs_t
      IF g_type="M" THEN
         LET ls_dzfs003 = "fs1.dzfs003"
      ELSE #S,Q
         LET ls_dzfs003 = "REPLACE(fs1.dzfs003,'",g_std_prog,"','",g_ind_prog,"')"
      END IF

      #備註:目前識別不出來那些設定被行業改過,暫時無法更新已經存在的設定,補上有缺的即可.
      LET ls_sql = "INSERT INTO dzfs_t(dzfs001,dzfs002,dzfs003,dzfs004,dzfs005,",
                   "                   dzfs006,dzfs007,dzfs008,dzfs009,dzfs010,",
                   "                   dzfs011,dzfs012,",
                   "                   dzfscrtdt,dzfscrtdp,dzfsowndp,dzfsownid,dzfsstus,dzfscrtid) ",
                   "SELECT ",g_ind_sver,",'",g_ind_prog,"',",ls_dzfs003,",fs1.dzfs004,'s',",
                   "       fs1.dzfs006,fs1.dzfs007,fs1.dzfs008,fs1.dzfs009,fs1.dzfs010,",     
                   "       fs1.dzfs011,fs1.dzfs012,",     
                   "       ?,'",gs_dept,"','",gs_dept,"','",g_user,"',fs1.dzfsstus,'",g_user,"'",
                   "  FROM dzfs_t fs1",
                   " WHERE fs1.dzfs002='",g_std_prog,"'",
                   "   AND fs1.dzfs003 NOT IN(SELECT fs2.dzfs003 FROM dzfs_t fs2",
                   "                           WHERE fs2.dzfs002='",g_ind_prog,"'",
                   "                             AND fs2.dzfs001=(SELECT dzaa004 FROM dzaa_t",
                   "                                               WHERE dzaa001='",g_ind_prog,"'",
                   "                                                 AND dzaa002=",g_ind_sver,
                   "                                                 AND dzaa003='TABLE'",
                   "                                                 AND dzaa009='s')", 
                   "                             AND fs2.dzfs005='s')",
                   "   AND fs1.dzfs001=",pr_std_dzaa.dzaa004,
                   "   AND fs1.dzfs005='s'"
      #display "sadzp060_4_sync_dzag_t:insert ind dzfs_t sql=",ls_sql
      PREPARE dzfs_prep2 FROM ls_sql
      EXECUTE dzfs_prep2 USING g_date
      FREE dzfs_prep2

      LET ls_tracing = "sadzp060_4_sync_dzag finish !"
      display ls_tracing

      RETURN TRUE,NULL
   CATCH
      LET ls_tracing = "sadzp060_4_sync_dzag failure !"
      CALL sadzp060_4_err_catch(ls_tracing, ls_sql)
      RETURN FALSE,"ERROR:"||ls_tracing
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 複製dzag_t(包含dzfs_t).
# Input parameter : p_prev_ind_dzaf 舊版行業版次資料
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2015/05/13 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_4_copy_dzag_t(p_prev_ind_dzaf)
   DEFINE p_prev_ind_dzaf T_DZAF_T 
   DEFINE ls_sql     STRING,
          ls_tracing STRING,
          li_cnt     SMALLINT,
          l_dzaa004  LIKE dzaa_t.dzaa004,
          l_dzaa006  LIKE dzaa_t.dzaa006

   TRY
      LET ls_tracing = "sadzp060_4_copy_dzag start !"
      display ls_tracing

      #先找行業規格舊版的TABLE識別碼版次和使用標示.
      LET ls_sql = "SELECT dzaa004,dzaa006 FROM dzaa_t",
                   " WHERE dzaa001='",g_ind_prog,"'",
                   "   AND dzaa002=",p_prev_ind_dzaf.dzaf003,
                   "   AND dzaa003='TABLE'",
                   "   AND dzaa005='4'",
                   "   AND dzaa009='s'"
      #display "sadzp060_4_copy_dzag_t:select ind old 'TABLE' sql=",ls_sql
      PREPARE dzag_prep11 FROM ls_sql
      EXECUTE dzag_prep11 INTO l_dzaa004,l_dzaa006 #170424-00013:增加dzaa006
      FREE dzag_prep11

      IF cl_null(l_dzaa004) OR l_dzaa004=0 THEN
         RETURN TRUE
      END IF

      #找看看行業規格有沒有新版的TABLE設計資料.
      LET li_cnt = 0
      LET ls_sql = "SELECT COUNT(1) FROM dzag_t",
                   " WHERE dzag001='",g_ind_prog,"'",
                   "   AND dzag003=",g_ind_sver,
                   "   AND dzag006='",l_dzaa006,"'"
      #display "sadzp060_4_copy_dzag_t:select ind dzag_t sql=",ls_sql
      PREPARE dzag_prep10 FROM ls_sql
      EXECUTE dzag_prep10 INTO li_cnt
      FREE dzag_prep10

      IF li_cnt=0 THEN
         #不存在就將舊版複製為新版,狀態碼不變.
         LET ls_sql = "INSERT INTO dzag_t(dzag001,dzag002,dzag003,dzag004,dzag005,",
                      "                   dzag006,dzag007,dzag008,dzag009,dzag010,",
                      "                   dzag011,dzag012,dzag013,dzag014,dzag015,",
                      "                   dzagcrtdt,dzagcrtdp,dzagowndp,dzagownid,dzagstus,dzagcrtid)",
                      "            SELECT dzag001,dzag002,",g_ind_sver,",dzag004,dzag005,",
                      "                   dzag006,dzag007,dzag008,dzag009,dzag010,",
                      "                   dzag011,dzag012,dzag013,dzag014,dzag015,",
                    # "                   ?,'",gs_dept,"','",gs_dept,"','",g_user,"',dzagstus,'",g_user,"'",
                      "                   ?,'",gs_dept,"','",gs_dept,"',dzagownid,dzagstus,'",g_user,"'", #170519-00055
                      "              FROM dzag_t",
                      "             WHERE dzag001='",g_ind_prog,"'",
                      "               AND dzag003=",l_dzaa004,
                      "               AND dzag006='",l_dzaa006,"'"
         #display "sadzp060_4_copy_dzag_t:insert ind dzag_t sql=",ls_sql 
         PREPARE dzag_prep12 FROM ls_sql
         EXECUTE dzag_prep12 USING g_date
         FREE dzag_prep12
      END IF
	  
      #dzfs_t一起處理.
      #找看看行業規格有沒有本版次的SCREEN RECORD設計資料
      LET li_cnt = 0
      LET ls_sql = "SELECT COUNT(1) FROM dzfs_t",
                   " WHERE dzfs002='",g_ind_prog,"'",
                   "   AND dzfs001=",g_ind_sver,
                   "   AND dzfs005='",l_dzaa006,"'"
      #display "sadzp060_4_copy_dzag_t:select ind dzfs_t sql=",ls_sql
      PREPARE dzfs_prep10 FROM ls_sql
      EXECUTE dzfs_prep10 INTO li_cnt
      FREE dzfs_prep10

      IF li_cnt=0 THEN
         #不存在就將舊版複製為新版,狀態碼不變.
         LET ls_sql = "INSERT INTO dzfs_t(dzfs001,dzfs002,dzfs003,dzfs004,dzfs005,",
                      "                   dzfs006,dzfs007,dzfs008,dzfs009,dzfs010,",
                      "                   dzfs011,dzfs012,",
                      "                   dzfscrtdt,dzfscrtdp,dzfsowndp,dzfsownid,dzfsstus,dzfscrtid)",
                      "            SELECT ",g_ind_sver,",dzfs002,dzfs003,dzfs004,dzfs005,",
                      "                   dzfs006,dzfs007,dzfs008,dzfs009,dzfs010,",
                      "                   dzfs011,dzfs012,",
                    # "                   ?,'",gs_dept,"','",gs_dept,"','",g_user,"',dzfsstus,'",g_user,"'",
                      "                   ?,'",gs_dept,"','",gs_dept,"',dzfsownid,dzfsstus,'",g_user,"'", #170519-00055
                      "              FROM dzfs_t",
                      "             WHERE dzfs002='",g_ind_prog,"'",
                      "               AND dzfs001=",l_dzaa004,
                      "               AND dzfs005='",l_dzaa006,"'"
         #display "sadzp060_4_copy_dzag_t:insert ind dzfs_t sql=",ls_sql
         PREPARE dzfs_prep12 FROM ls_sql
         EXECUTE dzfs_prep12 USING g_date
         FREE dzfs_prep12
      END IF

      LET ls_tracing = "sadzp060_4_copy_dzag finish !"
      display ls_tracing

      RETURN TRUE
   CATCH 
      LET ls_tracing = "sadzp060_4_copy_dzag failure !"
      CALL sadzp060_4_err_catch(ls_tracing, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE 
# Descriptions    : 同步行業規格:dzam_t
# Input parameter : pr_std_dzaa 標準dzaa_t資料
#                 : po_prev_ind_dzaf 舊版行業版次資料
# Return code     : BOOLEAN TRUE:成功;FALSE:失敗
#                 : STRING 錯誤訊息
# Date & Author   : 2015/05/13 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_4_sync_dzam(pr_std_dzaa, po_prev_ind_dzaf)
   DEFINE pr_std_dzaa TYPE_DZAA,
          po_prev_ind_dzaf T_DZAF_T
   DEFINE ls_tracing   STRING,
          ls_dzam003_1 STRING,
          ls_dzam003_2 STRING,
          ls_sql       STRING 

   TRY
      LET ls_tracing = "sadzp060_4_sync_dzam start !"
      display ls_tracing

      #要先複製舊版資料.
      IF NOT sadzp060_4_copy_dzam_t(po_prev_ind_dzaf.*) THEN
         RETURN FALSE,"ERROR:call sadzp060_4_copy_dzam_t error!"
      END IF

      #Begin:170703-00069:剔除dzamownid='i'的資料要避開,其他資料要覆蓋. #todo
      #刪除dzamownid標示為'i'的資料.
      LET ls_sql = "DELETE FROM dzam_t",
                   " WHERE dzam001='",g_ind_prog,"'",
                   "   AND dzam004=",g_ind_sver,
                   "   AND NVL(dzamownid,'n')<>'i'"
      display "sadzp060_4_sync_dzam_t:delete ind dzam_t sql=",ls_sql
      PREPARE dzam_prep4 FROM ls_sql
      EXECUTE dzam_prep4
      FREE dzam_prep4
      #End:170703-00069

      #排除控件(EXCLUDE)和表格設計資料(TABLE)一樣,都是將標準多餘行業的設計資料複製到行業規格內.
      IF g_type="M" THEN
         LET ls_dzam003_1 = "am1.dzam003"
         LET ls_dzam003_2 = "am2.dzam003"
      ELSE #S,Q
         LET ls_dzam003_1 = "REPLACE(am1.dzam003,'",g_std_prog,"','",g_ind_prog,"')"
         LET ls_dzam003_2 = "REPLACE(am2.dzam003,'",g_std_prog,"','",g_ind_prog,"')"
      END IF

      #備註:目前識別不出來那些設定被行業改過,暫時無法更新已經存在的設定,補上有缺的即可.
      LET ls_sql = "INSERT INTO dzam_t(dzam001,dzam002,dzam003,",
                   "                   dzam004,dzam005,dzam006,",
                   "                   dzamcrtdt,dzamcrtdp,dzamowndp,dzamownid,dzamstus,dzamcrtid) ",
                   "SELECT '",g_ind_prog,"',am1.dzam002,",ls_dzam003_1,",",
                           g_ind_sver,",'s',am1.dzam006,",     
                   "       ?,'",gs_dept,"','",gs_dept,"','",g_user,"',am1.dzamstus,'",g_user,"'",
                   "  FROM dzam_t am1",
                   " WHERE am1.dzam001='",g_std_prog,"'",
                   "   AND am1.dzam003 NOT IN(SELECT ",ls_dzam003_2," FROM dzam_t am2",
                   "                           WHERE am2.dzam001='",g_ind_prog,"'",
                   "                             AND am2.dzam004=(SELECT dzaa004 FROM dzaa_t",
                   "                                               WHERE dzaa001='",g_ind_prog,"'",
                   "                                                 AND dzaa002=",g_ind_sver,
                   "                                                 AND dzaa003='EXCLUDE'",
                   "                                                 AND dzaa009='s')", 
                   "                             AND am2.dzam005='s')",
                   "   AND am1.dzam004=",pr_std_dzaa.dzaa004,
                   "   AND am1.dzam005='s'"
      #display "sadzp060_4_sync_dzam:insert ind dzam_t sql=",ls_sql
      PREPARE dzam_prep2 FROM ls_sql
      EXECUTE dzam_prep2 USING g_date
      FREE dzam_prep2

      LET ls_tracing = "sadzp060_4_sync_dzam finish !"
      display ls_tracing

      RETURN TRUE,NULL
   CATCH
      LET ls_tracing = "sadzp060_4_sync_dzam failure !"
      CALL sadzp060_4_err_catch(ls_tracing, ls_sql)
      RETURN FALSE,"ERROR:"||ls_tracing
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 複製dzam_t
# Input parameter : p_prev_ind_dzaf 舊版行業版次資料
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2015/05/13 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_4_copy_dzam_t(p_prev_ind_dzaf)
   DEFINE p_prev_ind_dzaf T_DZAF_T 
   DEFINE ls_sql     STRING,
          ls_tracing STRING,
          li_cnt     SMALLINT,
          l_dzaa004  LIKE dzaa_t.dzaa004,
          l_dzaa006  LIKE dzaa_t.dzaa006

   TRY
      LET ls_tracing = "sadzp060_4_copy_dzam start !"
      display ls_tracing

      #先找行業規格舊版的EXCLUDE識別碼版次和使用標示.
      LET ls_sql = "SELECT dzaa004,dzaa006 FROM dzaa_t",
                   " WHERE dzaa001='",g_ind_prog,"'",
                   "   AND dzaa002=",p_prev_ind_dzaf.dzaf003,
                   "   AND dzaa003='EXCLUDE'",
                   "   AND dzaa005='a'",
                   "   AND dzaa009='s'"
      #display "sadzp060_4_copy_dzam:select ind old 'EXCLUDE' sql=",ls_sql 
      PREPARE dzam_prep11 FROM ls_sql
      EXECUTE dzam_prep11 INTO l_dzaa004,l_dzaa006
      FREE dzam_prep11

      IF cl_null(l_dzaa004) OR l_dzaa004=0 THEN
         RETURN TRUE
      END IF

      #找看看行業規格有沒有新版的EXCLUDE設計資料.
      LET li_cnt = 0
      LET ls_sql = "SELECT COUNT(1) FROM dzam_t",
                   " WHERE dzam001='",g_ind_prog,"'",
                   "   AND dzam004=",g_ind_sver,
                   "   AND dzam005='",l_dzaa006,"'"
      #display "sadzp060_4_copy_dzam_t:select ind dzam_t sql=",ls_sql
      PREPARE dzam_prep10 FROM ls_sql
      EXECUTE dzam_prep10 INTO li_cnt
      FREE dzam_prep10

      IF li_cnt=0 THEN
         #不存在就將舊版複製為新版,狀態碼不變.
         LET ls_sql = "INSERT INTO dzam_t(dzam001,dzam002,dzam003,dzam004,dzam005,dzam006,",
                      "                   dzamcrtdt,dzamcrtdp,dzamowndp,dzamownid,dzamstus,dzamcrtid)",
                      "            SELECT dzam001,dzam002,dzam003,",g_ind_sver,",dzam005,dzam006,",
                    # "                   ?,'",gs_dept,"','",gs_dept,"','",g_user,"',dzamstus,'",g_user,"'",
                      "                   ?,'",gs_dept,"','",gs_dept,"',dzamownid,dzamstus,'",g_user,"'", #170519-00055
                      "              FROM dzam_t",
                      "             WHERE dzam001='",g_ind_prog,"'",
                      "               AND dzam004=",l_dzaa004,
                      "               AND dzam005='",l_dzaa006,"'"
         #display "sadzp060_4_copy_dzam_t:insert ind dzam_t sql=",ls_sql
         PREPARE dzam_prep12 FROM ls_sql
         EXECUTE dzam_prep12 USING g_date
         FREE dzam_prep12
      END IF

      LET ls_tracing = "sadzp060_4_copy_dzam finish !"
      display ls_tracing

      RETURN TRUE
   CATCH 
      LET ls_tracing = "sadzp060_4_copy_dzam failure !"
      CALL sadzp060_4_err_catch(ls_tracing, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION

#20160118 by Hiko:曾經有大片註解:移除FUNCTION sadzp060_4_code_dzar_t

##########################################################################
# Access Modifier : PRIVATE 
# Descriptions    : 同步行業引用:dzfk009
# Input parameter : ps_prog 程式編號
#                 : ps_component 元件代號
# Return code     : BOOLEAN TRUE:成功;FALSE:失敗
#                 : STRING 錯誤訊息
# Date & Author   : 2019/09/17 by Elena
##########################################################################
PRIVATE FUNCTION sadzp060_4_sync_dzfk009(ps_prog, ps_component )

   DEFINE ps_prog    STRING
   DEFINE ps_component STRING 
   DEFINE ls_dzfk009 LIKE dzfk_t.dzfk009
   DEFINE ls_sql     STRING
   DEFINE ln_cnt     LIKE type_t.num5
   DEFINE l_component LIKE dzfk_t.dzfk003 #211109-00032
#g_std_sver  標準規格版次
#g_ind_prog
   #LOCATE ls_dzfk009 IN FILE #211109-00032:移到真正要開始處理的段落

   LET ls_sql = "SELECT COUNT(1) FROM dzaa_t WHERE dzaa001 = '", g_ind_prog ,"' AND dzaa003='", ps_component ,"'  AND dzaa007='Y' "
   PREPARE select_dzaa_t FROM ls_sql
   EXECUTE select_dzaa_t INTO ln_cnt

   IF ln_cnt > 0 THEN
 
      #LET ls_sql = "select gzzb003 from gzzb_t where gzzb001='", g_ind_prog ,"'"
      #PREPARE select_gzzb_t FROM ls_sql
      #DECLARE gzzb_curs CURSOR FOR select_gzzb_t

      #g_std_prog

      LET ls_sql = "SELECT COUNT(1) FROM dzfk_t WHERE dzfk001= '", g_ind_prog ,"' AND dzfk002= '", g_ind_sver ,"' AND dzfk003= '", ps_component ,"' AND dzfk007='s'" #211109-00032:補上dzfk007客製標示條件
      PREPARE select_dzfk_t FROM ls_sql
      EXECUTE select_dzfk_t INTO ln_cnt
  
      IF ln_cnt >0 THEN
         LOCATE ls_dzfk009 IN FILE #211109-00032:從上面移下來

         LET ls_sql = "SELECT dzfk009 FROM dzfk_t WHERE dzfk001= '", g_std_prog ,"' AND dzfk002 = ", g_std_sver , " AND dzfk003='", ps_component ,"' AND dzfk007='s'" #211109-00032:補上dzfk007客製標示條件 
         PREPARE select_dzfk_t1 FROM ls_sql
         #DECLARE dzfk_curs CURSOR FOR select_dzfk_t1 
         EXECUTE select_dzfk_t1 INTO ls_dzfk009 #211109-00032:這裡的條件只會有一筆，不需要foreach.

         #Begin:211109-00032 mark
        #FOREACH dzfk_curs INTO ls_dzfk009
        #   LET ls_sql = "UPDATE dzfk_t SET dzfk009 = '", ls_dzfk009 ,"' WHERE dzfk001='", g_ind_prog ,"' AND dzfk002= ", g_ind_sver ," AND dzfk003 ='", ps_component ,"' "

        #   PREPARE dzfk_update_prep FROM ls_sql
        #   EXECUTE dzfk_update_prep USING ls_dzfk009
        #END FOREACH
         #End:211109-00032
      
         #Begin:211109-00032:改寫更新方式
         LET l_component = ps_component

         UPDATE dzfk_t 
            SET dzfk009=ls_dzfk009
          WHERE dzfk001=g_ind_prog AND
                dzfk002=g_ind_sver AND
                dzfk003=l_component AND
                dzfk007='s' 

         FREE ls_dzfk009
         #End:211109-00032
      END IF
   #Begin:20200528:200529-00048
   #   RETURN TRUE
   #ELSE
   #   RETURN FALSE 
   #End:20200528
   END IF

   RETURN TRUE #20200528:不是所有規格都有ui控件,所以沒有找到也沒關係.
END FUNCTION
