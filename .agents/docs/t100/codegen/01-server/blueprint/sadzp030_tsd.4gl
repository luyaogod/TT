#
#+ 程式代碼......: sadzp030_tsd
#+ 設計人員......: Hiko
# Prog. Version..: 'T6-12.01.21(00000)'     #
#
# Program name   : sadzp030_tsd.4gl
# Description    : 建立tsd檔
# Memo           :
# Modify         : 2013/07/02 by Hiko : 因應子程式調整
#                : 2013/07/06 by Hiko : 調整tsd內<table>結構
#                : 2013/07/08 by Hiko : 調整tsd內<tree>結構
#                : 2013/07/30 by Hiko : 因應normal_style調整
#                : 2013/08/16 by Hiko : 因應azzi900,azzi901而調整
#                : 2013/08/23 by Hiko : 1.忽略多語言產生(adzp191)的錯誤
#                                       2.調整dzae004(UI版型)的程式段
#                : 2013/08/29 by Hiko : 因應ScreenRecord允許插入,刪除,新增的屬性
#                : 2013/09/02 by Hiko : 增加sa規格的結構
#                : 2013/09/04 by Hiko : 新建tsd前先做備份
#                : 2013/09/13 by Hiko : 1.增加TREE type2~type6調整.
#                                     : 2.刪除dzae_t,改用dzaq_t.
#                : 2013/09/24 by Hiko : 資料多語言增加規格描述功能(dzaj099)
#                : 2013/09/26 by Hiko : 增加排除設定功能(dzam_t)
#                : 2013/10/08 by Hiko : 因應ALM管控,增加規格樹的建立
#                : 2013/10/11 by Hiko : sadzp030_tsd_chk_alm增加回傳參數
#                : 2013/10/17 by Hiko : sadzp030_tsd_chk_alm改為cl_chk_competence_check_alm
#                : 2013/10/24 by Hiko : 修正dzaa_t複製時跳版次的問題
#                : 2013/11/28 by Hiko : 移除dzah_t,並將dzah005的設定改在dzad006紀錄,並改成以逗號(,)隔開觸發時機.
#                : 2013/12/06 by Hiko : dzfq003暫時都是取得第一版次.
#                : 2013/12/13 by Hiko : <table><exclude>若沒有版次則還是給ver,版次就給當下要產生tsd的版次即可,若存在真實資料,則會再次覆蓋,這樣對於設計器的操作會比較簡單.
#                : 2014/01/07 by Hiko : 調整子程式/子畫面的toolbar.
#                : 2014/01/09 by Hiko : 因應行業別調整:1.dzaa_t增加dzaa007,dzaa008.
#                                                      2.資源池的Table移除'引用'的相關欄位.
#                                                      3.tsd內cite_std改成取得dzaa007.
#                                                      4.取得資源池的SQL與接收都改成列舉.
#                                                      5.排除設定(dzam_t)與Table設定(dzag_t)不能引用,行業別自己儲存一份,以免出現其他問題.
#                                                      6.dzff008改成PK,<tree>的src和status改在頭的tag,且dzaa003改成tree name(不為固定"TREE").
#                                                      7.dzam005改成PK<exclude>的src改在頭的tag.
#                : 2014/01/21 by Hiko : 因應UI關聯進階設定調整
#                : 2014/02/10 by Hiko : 增加程式串查型態(dzal007)
#                : 2014/02/12 by Hiko : 1.dzfu_t移除dzfu001,dzfu002,並將原本的dzfu003,dzfu004往前改為dzfu001,dzfu002
#                                       2.新增dzax_t
#                : 2014/02/26 by Hiko : 增加type屬性(M,T,S,P..)
#                : 2014/03/06 by Hiko : 修改助記碼(dzak_t)
#                : 2014/03/14 by Hiko : 修改程式串查(dzal_t)的格式
#                : 2014/03/19 by Hiko : dzam005不改成PK,且dzaa006不是條件
#                : 2014/04/02 by Hiko : 下載要剃除垃圾資料(dzac_t,dzak_t,dzal_t,dzai_t,dzaj_t,dzam_t),以免設計器判斷出現問題.多語言資料沒關係.
#                : 2014/04/03 by Hiko : 修改comment產生(因為和text相同)
#                : 2014/04/07 by Hiko : 增加建立報表元件規格的程式段(gen_rsd)
#                : 2014/04/14 by Hiko : 修正行業別資料取得的SQL錯誤,重點為RIGHT JOIN dzaa_t.
#                : 2014/04/16 by Hiko : TREE的結構補齊所有tag
#                : 2014/04/18 by Hiko : 1.修正程式串查的取得方式
#                                       2.剔除strings重複Key
#                : 2014/04/23 by Hiko : 1.行業別沒有dzfq_t(r.a)的資料,所以要以標準程式來找
#                                       2.行業別引用改採用類似複製的動作(產生行業別程式會處理),所以這裡就不需要再去標準程式撈取,比照正常開發流程即可.
#                : 2014/04/30 by Hiko : 增加dzad007(產生標準程式)
#                : 2014/05/06 by Hiko : 簡化行業別複雜度,所以利用'複製'的方式來實作'引用',這樣複雜邏輯只需要在上傳(sadzp060_1)處理即可.
#                : 2014/05/15 by Hiko : 1.移除cl_chk_competence相關程式段,因為權限已經在簽出的時候就預備好了,但需要再新增一個tsd.read檔,讓純下載的時候使用.
#                                       2.移除gzza007,gzde004
#                : 2014/05/23 by Hiko : 因應二次開發調整:dzaa_t.dzaa009(PK)
#                : 2014/05/27 by Hiko : 1.調整Toolbar的取得
#                                       2.移除template屬性,不需要取得dzfq_t.
#                                       3.<other>(dzax_t)改在tap更新
#                : 2014/05/29 by Hiko : 因應PK調整:dzag006,dzfs005,dzam005
#                : 2014/06/06 by Hiko : tsd也是需要<other>,但是為了規格設計器的4fd控管.
#                : 2014/06/10 by Hiko : dzax_t增加dzax006(PK)
#                : 2014/06/11 by madey: 若p_std_prog參數為null則給gs_prog
#                : 2014/08/05 by Hiko : 同一個版次的status要完整呈現u和d,因此取得SQL要變成生/失效兩句UNION
#                : 2014/08/05 by Hiko : dzac006(欄位屬性),dzac007(資料型態)不再更新,所以就不再讀取.
#                : 2014/10/03 by Hiko : 增加dzax007(接收參數起始順序)
#                : 2014/10/22 by Hiko : 客戶環境不能刪除標準控件(dzfi_t,dzfj_t,dzfk_t,dzfl_t)
#                : 2014/10/27 by Hiko : 產生不可刪除清單的時候要判斷標準規格版次是否>0
#                : 2014/12/04 by Hiko : 下載規格就不重產42s,因為r.a會自行重產,上傳會呼叫r.f,Patch最後也會重產,所以就以目前server上的42s當作開發基礎即可.
#                : 2014/12/05 by Hiko : 將dzac015(最大值)拆解成dzac020+dzac015,dzac016(最小值)拆解成dzac021+dzac016.
#                : 2014/12/10 by Hiko : 1.本版次異動只記錄版次>1的情況,這樣比較有意義:由於失效資料的關係會造成錯誤,暫時移除.#todo
#                                       2.將gs_revision改成g_revision LIKE dzaf_t.dzaf003
#                : 2014/12/17 by Hiko : 1.修改程式串查的bug
#                                       2.本版次異動調整:dzaa006(使用標示)=dzaa009(客製標示)時,版次>1才設定u,這可提升剛開發程式的上傳效能;但d就不論版次都需要設定.
#                : 2014/12/19 by Hiko : 客製沒有行業別,所以行業別程式下載的時候,除了std_prog設定和prog相同外,還要將cite_std變成N.
#                : 2014/12/29 by Hiko : dzag_t增加三個欄位:dzag013,dzag014,dzag015
#                : 2014/12/30 by Hiko : 1.不需要特別去擋客製不可以改標準設定(將10/22調整的註解掉)
#                                       2.修正dzal_t(程式串查)的問題
#                : 2015/04/23 by Hiko : 呼叫azzp191來產生當下的str
#                : 2015/04/27 by Hiko : 增加設計器版次資訊的讀取,讓sadzp060_2得以判斷.
#                : 2015/11/09 by Hiko : 1.為了改善上傳效能,增加一個屬性'ch'來當作本版次異動.
#                                       2.dzax_t取消本版次異動.
#                                       3.失效的資料就不取出來了.
#                : 20160125 by Hiko : 下載路徑改在$TEMPDIR, 以免因為權限不足而無法下載.
#                : 20160128 by Hiko : gs_prog_type改成小寫.
#                : 20170517 170515-00013 by Hiko : 新增加預編譯屬性:pre_compile
#                : 20170622 170622-00041 by Hiko : 為了簡化開發流程,下載規格不重新產生str,而由上傳一起處理.
#                : 20170725 170725-00043 by madey: 新增屬性topind 
#                : 20230925 230925-00019#1 by 02632 : 註解掉,避免tsd.xsd問題:完成需处理程式讯息的提示功能:提示程序信息顯示於規格設計器之Message bar上.
#                : 20240422 .40422-00015#1 by 02431 : 註解掉,避免tsd.xsd問題:遇到僅有規格沒程式的情況下，原寫法'版次'的方式，導致MOG會辨別報錯，調整成用?的方式

 
import os
import xml

SCHEMA ds

GLOBALS "../../cfg/top_global.inc"
&include "../4gl/sadzp200_type.inc"

DEFINE gs_env STRING #辨識目前所在的環境:s.標準,c.客製
DEFINE gd_date DATETIME YEAR TO SECOND
#DEFINE gs_revision STRING #規格版次 #2013/12/13 by Hiko
DEFINE g_revision LIKE dzaf_t.dzaf003 #規格版次 #2014/12/10 by Hiko
DEFINE gs_erpver STRING,  #ERP大版版號
       gs_zone STRING     #執行區域
DEFINE gs_module STRING,
       gs_prog STRING, #2014/01/09 by Hiko
       gs_prog_type STRING, #2014/01/07 by Hiko
       gs_identity STRING #2014/05/23 by Hiko
#Begin:2014/04/18 by Hiko
#Begin:2014/12/17 by Hiko
TYPE T_DZAL_T DYNAMIC ARRAY OF RECORD
              cite_std LIKE dzaa_t.dzaa007,
              dzal002 LIKE dzal_t.dzal002, 
              dzal003 LIKE dzal_t.dzal003, 
              dzal004 LIKE dzal_t.dzal004, 
              dzal005 LIKE dzal_t.dzal005, 
              dzal006 LIKE dzal_t.dzal006, 
              dzal007 LIKE dzal_t.dzal007, 
              dzal008 LIKE dzal_t.dzal008,
              dzal009 LIKE dzal_t.dzal009, #2014/12/17 by Hiko
              dzaa009 LIKE dzaa_t.dzaa009, #2014/12/17 by Hiko
              dzalstus LIKE dzal_t.dzalstus			   
              END RECORD
DEFINE ga_dzal T_DZAL_T
#DEFINE ga_dzal DYNAMIC ARRAY OF RECORD
#               cite_std LIKE dzaa_t.dzaa007,
#               dzal002 LIKE dzal_t.dzal002, 
#               dzal003 LIKE dzal_t.dzal003, 
#               dzal004 LIKE dzal_t.dzal004, 
#               dzal005 LIKE dzal_t.dzal005, 
#               dzal006 LIKE dzal_t.dzal006, 
#               dzal007 LIKE dzal_t.dzal007, 
#               dzal008 LIKE dzal_t.dzal008,
#               dzal009 LIKE dzal_t.dzal009, #2014/12/17 by Hiko
#               dzaa009 LIKE dzaa_t.dzaa009, #2014/12/17 by Hiko
#               dzalstus LIKE dzal_t.dzalstus			   
#               END RECORD 
#End:2014/12/17 by Hiko
#End:2014/04/18 by Hiko
DEFINE gb_industry BOOLEAN #2014/12/19 by Hiko
DEFINE gs_topind   STRING #170725-00043 #行業平台環境變數 TOPIND

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 初始化變數.
# Input parameter : none
# Return code     : void
# Date & Author   : 2013/01/11 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_init_var()
   LET gs_env = FGL_GETENV("DGENV") CLIPPED
   LET gs_erpver = FGL_GETENV("ERPVER") CLIPPED
   LET gs_zone = FGL_GETENV("ZONE") CLIPPED
   LET gd_date = cl_get_current()
   CALL ga_dzal.clear() #一定要清空,要不然重複呼叫會殘留.
   LET gs_topind = FGL_GETENV("TOPIND") CLIPPED         #170725-00043
   IF cl_null(gs_topind) THEN LET gs_topind="sd" END IF #170725-00043
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC 
# Descriptions    : 產生(SD規格設計檔)(程式名稱.tsd).
# Input parameter : p_prog 規格代號
#                 : p_module 模組
#                 : p_revision 規格版次
#                 : p_type 建構類型
#                 : p_std_prog 行業別對應的標準規格代號
#                 : p_std_revision 行業別對應的標準規格版次
#                 : p_identity 識別標示(s/c)
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2013/01/11 by Hiko
# Modify          : 2013/08/16 by Hiko 增加程式類型的參數
# Modify          : 2014/01/09 by Hiko 增加程式類型的參數
##########################################################################
PUBLIC FUNCTION sadzp030_tsd_gen_tsd(p_prog, p_module, p_revision, p_type, p_std_prog, p_std_revision, p_identity) #2014/05/23 by Hiko
   DEFINE p_prog STRING,
          p_module STRING,
          #p_revision STRING,
          p_revision LIKE dzaf_t.dzaf003, #2014/12/10 by Hiko
          p_type STRING, #M.主程式與畫面、S.子程式與畫面、F.子畫面
          p_std_prog STRING,
          #p_std_revision STRING,
          p_std_revision LIKE dzaf_t.dzaf003, #2014/12/10 by Hiko
          p_identity STRING #s/c
   DEFINE ls_sql STRING
   DEFINE ls_trigger STRING
   DEFINE ls_module_dir STRING #模組目錄
   #Begin:2013/10/08 by Hiko
   DEFINE li_cnt SMALLINT,
          li_revision SMALLINT,
          ls_old_revision STRING,
          ls_dept STRING,
          ls_booking STRING
   #End:2013/10/08 by Hiko
   DEFINE lr_gzza RECORD
                  gzza002 LIKE gzza_t.gzza002 #程式型態
                  #gzza008 LIKE gzza_t.gzza008  #主畫面配置方式 #2013/07/30 by Hiko
                  #gzza007 LIKE gzza_t.gzza007  #2013/08/16 by Hiko #2014/05/15 by Hiko
                  END RECORD
   #Begin:2013/09/13 by Hiko
   DEFINE lr_dzfq RECORD LIKE dzfq_t.*
   #End:2013/09/13 by Hiko
   DEFINE ls_tsd STRING
   DEFINE ldoc_tsd xml.DomDocument,
          lnode_tsd xml.DomNode
   DEFINE ls_str STRING, #多語言檔路徑:str,4ad
          lnode_str xml.DomNode,
          ls_run_cmd STRING,
          lb_run_result BOOLEAN,
          ls_run_msg STRING
   DEFINE lr_gzde RECORD #2013/08/16 by Hiko
                  #gzde004 LIKE gzde_t.gzde004, #2014/05/15 by Hiko
                  gzde005 LIKE gzde_t.gzde005
                  END RECORD
   #Begin:2013/10/11 by Hiko
   DEFINE lb_alm_result BOOLEAN,
          ls_err_msg STRING
   #End:2013/10/11 by Hiko
   DEFINE li_idx SMALLINT
   DEFINE lo_DZAF_T T_DZAF_T #2014/10/22 by Hiko
   ##230925-00019#1 -s
   #DEFINE li_code_ver9     SMALLINT,
   #       ls_identity9     STRING, #no_use
   #       ls_err_msg9      STRING, #no_use
   #       ls_section_flag9 STRING
   ##230925-00019#1 -e

   LET gs_prog = p_prog.trim()
   LET gs_module = p_module.toUpperCase()
   #LET g_revision = p_revision.trim()
   LET g_revision = p_revision #2014/12/10 by Hiko
   LET gs_identity = p_identity.trim()

DISPLAY "Hiko:p_prog=",p_prog
DISPLAY "Hiko:p_module=",p_module
DISPLAY "Hiko:p_revision=",p_revision
DISPLAY "Hiko:p_type=",p_type
DISPLAY "Hiko:p_std_prog=",p_std_prog
DISPLAY "Hiko:p_std_revision=",p_std_revision
DISPLAY "Hiko:p_identity=",p_identity

   CALL sadzp030_tsd_init_var()

   TRY
      #Begin:2014/10/31 by Hiko
      IF cl_null(p_std_revision) THEN
         LET p_std_revision = 0
      END IF
      IF cl_null(p_revision) THEN
         LET p_revision = 0
      END IF
      #End:2014/10/31 by Hiko
      LET ls_module_dir = FGL_GETENV(gs_module) CLIPPED
      #產生tsd檔根節點.
      LET ls_trigger = "sadzp030_tsd_gen_tsd : gen tsd"
      #LET ls_tsd = os.path.join(os.path.join(os.path.join(ls_module_dir, "dzx"), "tsd"), gs_prog||".tsd")
      LET ls_tsd = os.path.join(FGL_GETENV("TEMPDIR"), gs_prog||".tsd") #20160125 by Hiko
      LET ldoc_tsd = xml.domDocument.createDocument("spec")
      LET lnode_tsd = ldoc_tsd.getDocumentElement()
      LET p_std_prog = NVL(p_std_prog,gs_prog) #2014/06/11 by madey
      #Begin:2014/12/19 by Hiko:客製環境沒有行業開發行為,而topstd就是在客製環境開發標準程式,所以也當作是沒有引用行為.
      LET gb_industry = FALSE
      IF NOT (gs_env="c" OR g_account="topstd") THEN
         IF NOT p_std_prog.equals(gs_prog) THEN
            LET gb_industry = TRUE
         END IF
      END IF

      IF NOT gb_industry THEN
         LET p_std_prog = gs_prog #不允許行業引用的狀況下就強制覆蓋p_std_prog
      END IF
      #End:2014/12/19 by Hiko
      CALL lnode_tsd.setAttribute("prog", gs_prog)
      CALL lnode_tsd.setAttribute("std_prog", p_std_prog.trim())
      CALL lnode_tsd.setAttribute("erpver", gs_erpver) #ERP大版
      CALL lnode_tsd.setAttribute("ver", g_revision)      
      CALL lnode_tsd.setAttribute("module", gs_module)      
      CALL lnode_tsd.setAttribute("type", p_type) #2014/05/20 by Hiko:M/S/F/B/G/X/...  
      CALL lnode_tsd.setAttribute("booking", "Y") #2014/05/15 by Hiko:以流程來看,會呼叫sadzp030_tsd的時候就是有權限,所以不需要再檢查.
      CALL lnode_tsd.setAttribute("designer_ver", sadzp060_2_get_designer_ver()) #2015/04/27 by Hiko     

      LET ls_trigger = "sadzp030_tsd_gen_tsd : get gzza_t data"
      LET ls_sql = "SELECT LOWER(gzza002)",
                    " FROM gzza_t",
                   " WHERE gzza001='",gs_prog,"' AND gzzastus='Y'"
      PREPARE gzza_prep FROM ls_sql
      EXECUTE gzza_prep INTO lr_gzza.*
      FREE gzza_prep
      
      LET gs_prog_type = lr_gzza.gzza002 CLIPPED

      IF cl_null(gs_prog_type) THEN
         LET ls_trigger = "sadzp030_tsd_gen_tsd : get gzde_t data"
         LET ls_sql = "SELECT gzde005", #2014/01/07 by Hiko
                       " FROM gzde_t",
                      " WHERE gzde001='",gs_prog,"' AND gzdestus='Y'"
         PREPARE gzde_prep FROM ls_sql
         EXECUTE gzde_prep INTO lr_gzde.*
         FREE gzde_prep
         
         LET gs_prog_type = lr_gzde.gzde005 CLIPPED #2014/01/07 by Hiko
      END IF

      CALL lnode_tsd.setAttribute("class", gs_prog_type) #子畫面(F)沒有Toolbar,但因為資料也不在gzza_t和gzde_t內,所以也不用特別排除
      CALL lnode_tsd.setAttribute("env", gs_env)      
      CALL lnode_tsd.setAttribute("zone", gs_zone)      
      CALL lnode_tsd.setAttribute("identity", gs_identity)      
      CALL lnode_tsd.setAttribute("pre_compile", "Y") #170515-00013
      CALL lnode_tsd.setAttribute("topind", gs_topind)#170725-00043
      ##230925-00019#1 -s
      #CALL sadzp060_2_get_code_curr_revision(gs_prog, p_type, gs_module) RETURNING li_code_ver9,ls_identity9,ls_err_msg9
      #CALL sadzp030_tsd_get_section_if_modify(gs_prog, li_code_ver9, gs_env) RETURNING ls_section_flag9
      #CALL lnode_tsd.setAttribute("section_flag", ls_section_flag9)   
      ##230925-00019#1 -e
      
      #Begin:2014/06/06 by Hiko
      #建立<other>程式其他設定.
      LET ls_trigger = "sadzp030_tsd_gen_tsd : call sadzp030_tsd_gen_other"
      IF NOT sadzp030_tsd_gen_other(lnode_tsd) THEN
         RETURN FALSE
      END IF
      DISPLAY "call sadzp030_tsd_gen_other finish!"
      #End:2014/06/06 by Hiko

      #建立<toolbar>符合樣版的設定.
      LET ls_trigger = "sadzp030_tsd_gen_tsd : call sadzp030_tsd_gen_toolbar"
      IF NOT cl_null(gs_prog_type) THEN #2014/01/07 by Hiko
         IF NOT sadzp030_tsd_gen_toolbar(lnode_tsd) THEN
            RETURN FALSE
         END IF
      END IF
      DISPLAY "call sadzp030_tsd_gen_toolbar finish!"
      
      #建立<table>程式主Table設定.
      LET ls_trigger = "sadzp030_tsd_gen_tsd : call sadzp030_tsd_gen_table"
      IF NOT sadzp030_tsd_gen_table(lnode_tsd) THEN
         RETURN FALSE
      END IF
      DISPLAY "call sadzp030_tsd_gen_table finish!"
 
      #建立<tree>設定.
      #2014/04/23 by Hiko:因為行業別不是透過r.a產生,所以樣版資料是透過對應的標準程式來取得.一般程式和行業別都會有p_std_prog,所以這樣調整即可.
      #固定版次1來找尋UI樣板資料.
      LET ls_sql = "SELECT dzfq_t.*",
                    " FROM dzfq_t",
                   " WHERE dzfq004='",p_std_prog,"' AND dzfq003='1' AND dzfqstus='Y'" 
      PREPARE dzfq_prep FROM ls_sql
      EXECUTE dzfq_prep INTO lr_dzfq.*
      FREE dzfq_prep

      LET ls_trigger = "sadzp030_tsd_gen_tsd : call sadzp030_tsd_gen_tree"
      IF NOT sadzp030_tsd_gen_tree(lnode_tsd, lr_dzfq.dzfq016 CLIPPED) THEN
         RETURN FALSE
      END IF
      DISPLAY "call sadzp030_tsd_gen_tree finish!"

      #建立<all>規格設定與描述.
      LET ls_trigger = "sadzp030_tsd_gen_tsd : call sadzp030_tsd_gen_all"
      IF NOT sadzp030_tsd_gen_all(ldoc_tsd, lnode_tsd) THEN
         RETURN FALSE
      END IF
      DISPLAY "call sadzp030_tsd_gen_all finish!"

      #建立<field>規格設定與描述.
      LET ls_trigger = "sadzp030_tsd_gen_tsd : call sadzp030_tsd_gen_field"
      IF NOT sadzp030_tsd_gen_field(ldoc_tsd, lnode_tsd) THEN
         RETURN FALSE
      END IF
      DISPLAY "call sadzp030_tsd_gen_field finish!"

      #建立<prog_rel>規格設定.
      LET ls_trigger = "sadzp030_tsd_gen_tsd : call sadzp030_tsd_gen_prog_rel"
      IF NOT sadzp030_tsd_gen_prog_rel(lnode_tsd) THEN
         RETURN FALSE
      END IF
      DISPLAY "call sadzp030_tsd_gen_prog_rel finish!"

      #建立<ref_field>規格設定.
      LET ls_trigger = "sadzp030_tsd_gen_tsd : call sadzp030_tsd_gen_ref_field"
      IF NOT sadzp030_tsd_gen_ref_field(lnode_tsd) THEN
         RETURN FALSE
      END IF
      DISPLAY "call sadzp030_tsd_gen_ref_field finish!"

      #建立<multi_lang>規格設定.
      LET ls_trigger = "sadzp030_tsd_gen_tsd : call sadzp030_tsd_gen_multi_lang"
      #IF NOT sadzp030_tsd_gen_multi_lang(lnode_tsd) THEN #2013/09/24 by Hiko
      IF NOT sadzp030_tsd_gen_multi_lang(ldoc_tsd, lnode_tsd) THEN
         RETURN FALSE
      END IF
      DISPLAY "call sadzp030_tsd_gen_multi_lang finish!"

      #建立<help_code>規格設定.
      LET ls_trigger = "sadzp030_tsd_gen_tsd : call sadzp030_tsd_gen_help_code"
      IF NOT sadzp030_tsd_gen_help_code(lnode_tsd) THEN
         RETURN FALSE
      END IF
      DISPLAY "call sadzp030_tsd_gen_help_code finish!"

      #建立<act>規格設定與描述.
      LET ls_trigger = "sadzp030_tsd_gen_tsd : call sadzp030_tsd_gen_act"
      IF NOT sadzp030_tsd_gen_act(ldoc_tsd, lnode_tsd) THEN
         RETURN FALSE
      END IF
      DISPLAY "call sadzp030_tsd_gen_act finish!"

      #建立<field>多語言節點資料.
      LET lnode_str = lnode_tsd.appendChildElement("strings")
      #2013/06/17 by Hiko : 還原讀取str當作tsd多語言資料的來源.
      #Begin:170622-00041
      # #取得str檔當作來源:先呼叫產生str的azzp191.
      # #2014/12/04 by Hiko:直接以server上的42s當作開發基礎即可.
      # LET ls_run_cmd = "r.r azzp191 ",gs_prog," ",g_lang," '' '' 'Y'" #2015/04/23 by Hiko:增加第五個參數
      # LET ls_trigger = "sadzp030_tsd_gen_tsd : call sadzp030_tsd_gen_strings_field : ",ls_run_cmd
      # CALL cl_cmdrun_openpipe("r.r azzp191", ls_run_cmd, FALSE) RETURNING lb_run_result,ls_run_msg
      # IF lb_run_result THEN
        LET ls_str = os.path.join(os.path.join(os.path.join(ls_module_dir, "str"), g_lang), gs_prog||".str")
        IF sadzp030_tsd_check_file_exist(ls_str) THEN
           LET ls_trigger = "sadzp030_tsd_gen_tsd : call sadzp030_tsd_gen_strings_field"
           #Begin:2013/08/23 by Hiko : 忽略產生多語言的錯誤, 因為多語言的產生一定需要此程式可以執行(可以r.r),
           #                           而現在下載規格/程式, 已經不管是否可以編譯成功都要可以下載, 所以這邊就不控管了.
           CALL sadzp030_tsd_gen_strings_field(ls_str, lnode_str) RETURNING lb_run_result
           #End:2013/08/23 by Hiko
           DISPLAY "call sadzp030_tsd_gen_strings_field finish!"
        END IF
      # ELSE
      #    #DISPLAY "r.r azzp191 ",ls_run_msg
      # END IF
      #End:170622-00041

      LET ls_trigger = "sadzp030_tsd_gen_tsd : call sadzp030_tsd_gen_strings_act"
      #忽略產生多語言的錯誤
      CALL sadzp030_tsd_gen_strings_act(lnode_str) RETURNING lb_run_result
      DISPLAY "call sadzp030_tsd_gen_strings_act finish!"

      #忽略取得SA規格的錯誤
      #建立<sa_spec>規格描述.
      LET ls_trigger = "sadzp030_tsd_gen_tsd : call sadzp030_tsd_gen_sa_spec"
      CALL sadzp030_tsd_gen_sa_spec(ldoc_tsd, lnode_tsd) RETURNING lb_run_result
      DISPLAY "call sadzp030_tsd_gen_sa_spec finish!"

      #Begin:2013/09/26 by Hiko
      #建立<exclude>排除設定.
      LET ls_trigger = "sadzp030_tsd_gen_tsd : call sadzp030_tsd_gen_exclude"
      IF NOT sadzp030_tsd_gen_exclude(ldoc_tsd, lnode_tsd) THEN
         RETURN FALSE
      END IF
      DISPLAY "call sadzp030_tsd_gen_exclude finish!"
      #End:2013/09/26 by Hiko

      #Begin:2014/12/30 by Hiko:不需要特別控卡
      ##Begin:2014/10/22 by Hiko
      #IF gs_env<>"s" THEN
      #   #建立<cannot_del>不可刪除控件清單.
      #   #先取得最大的標準規格版次.
      #   LET lo_DZAF_T.DZAF001 = p_prog.trim()
      #   LET lo_DZAF_T.DZAF005 = p_type.trim()
      #   LET lo_DZAF_T.DZAF006 = p_module.trim()
      #   LET lo_DZAF_T.DZAF010 = "s"
      #   CALL sadzp200_ver_get_curr_ver_info(lo_DZAF_T.*) RETURNING lo_DZAF_T.*
      #   
      #   LET ls_trigger = "sadzp030_tsd_gen_tsd : call sadzp030_tsd_gen_cannot_del"
      #   IF NOT sadzp030_tsd_gen_cannot_del(ldoc_tsd, lnode_tsd, lo_DZAF_T.dzaf003) THEN
      #      RETURN FALSE
      #   END IF
      #END IF
      ##End:2014/10/22 by Hiko
      #End:2014/12/30 by Hiko

      LET ls_trigger = "sadzp030_tsd_gen_tsd : save tsd"
      #Begin:2014/04/07 by Hiko
      IF NOT sadzp030_tsd_save_tsd(ldoc_tsd, ls_tsd) THEN
         RETURN FALSE
      END IF
      #End:2014/04/07 by Hiko
      #Begin:2014/05/15 by Hiko:建立tsd.read
      LET ls_trigger = "sadzp030_tsd_gen_tsd : save tsd.read"
      LET ls_tsd = ls_tsd,".read"
      CALL lnode_tsd.setAttribute("booking", "N")
      IF NOT sadzp030_tsd_save_tsd(ldoc_tsd, ls_tsd) THEN
         RETURN FALSE
      END IF
      #End:2014/05/15 by Hiko
      DISPLAY "gen tsd success : ",ls_tsd

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 完成規格離線檔的建立.
# Input parameter : p_doc_tsd <tsd> doc
#                 : p_tsd_path tsd檔路徑
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2014/04/07 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_save_tsd(p_doc_tsd, p_tsd_path)
   DEFINE p_doc_tsd xml.DomDocument,
          p_tsd_path STRING
   DEFINE ls_sql STRING,
          ls_trigger STRING

   TRY
      #Begin:2013/09/04 by Hiko : 將tsd更名
      LET ls_trigger = "sadzp030_tsd_save_tsd : tsd rename"
      IF os.Path.exists(p_tsd_path) THEN
         IF NOT os.Path.rename(p_tsd_path, p_tsd_path||".bak") THEN
            INITIALIZE g_errparam TO NULL
            LET g_errparam.code =  "adz-00174"
            LET g_errparam.extend = ls_trigger
            LET g_errparam.popup = TRUE
            LET g_errparam.replace[1] =  "rename"
            LET g_errparam.replace[2] = gs_prog
            CALL cl_err()
            RETURN FALSE
         END IF
      END IF
      #End:2013/09/04 by Hiko
      
      CALL p_doc_tsd.setFeature("format-pretty-print", "TRUE")
      CALL p_doc_tsd.save(p_tsd_path)
      #Begin:2013/09/04 by Hiko : 最後變成777, 這樣在運行過程中會比較保險.
      IF NOT os.Path.chrwx(p_tsd_path, 511) THEN
         DISPLAY "CALL sadzp030_tsd_save_tsd : ",p_tsd_path," change rwx ERROR!"
         INITIALIZE g_errparam TO NULL
         LET g_errparam.code =  "adz-00174"
         LET g_errparam.extend = ls_trigger
         LET g_errparam.popup = TRUE
         LET g_errparam.replace[1] =  "chmod"
         LET g_errparam.replace[2] = gs_prog
         CALL cl_err()
         RETURN FALSE
      END IF
      #End:2013/09/04 by Hiko
      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 擷取錯誤訊息.
# Input parameter : p_trigger 執行的顯現資訊
#                   p_sql 執行的SQL
# Return code     : void
# Date & Author   : 2013/04/12 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_err_catch(p_trigger, p_sql)
   DEFINE p_trigger STRING,
          p_sql STRING

   DISPLAY "ERROR : ls_trigger=",p_trigger
   DISPLAY "STATUS=",STATUS
   DISPLAY "SQLCA.SQLCODE=",SQLCA.SQLCODE
   DISPLAY "ls_sql=",p_sql,"<<"
END FUNCTION

##230925-00019#1 -s #從adzp110_get_section_if_modify()完整複製過來.
##判斷程式是否解開Section.
#PRIVATE FUNCTION sadzp030_tsd_get_section_if_modify(p_file_name,p_version,p_dgenv)
#DEFINE
#  p_file_name STRING,
#  p_version   STRING,
#  p_dgenv     STRING
#DEFINE
#  ls_file_name STRING,
#  ls_version   STRING,
#  ls_dgenv     STRING,
#  ls_SQL       STRING,
#  li_counts    INTEGER,
#  ls_return    STRING
#
#DEFINE ls_version_num LIKE dzbc_t.dzbc002   #240422-00015#1
#
#
#  LET ls_file_name = p_file_name.trim()
#  LET ls_version   = p_version.trim()
#  LET ls_dgenv     = p_dgenv.trim()
#
#  LET li_counts = 0
#  LET ls_version_num = ls_version          #240422-00015#1
#   
#  #Begin:200215-00009
#  LET ls_sql = "select count(*)                                     ",
#               "  from dzbd_t bd                                    ",
#               "  inner join dzbc_t bc on bc.dzbc001  = bd.dzbd001  ",
#               "                      and bc.dzbc004  = bd.dzbd003  ",
#               "                      and bc.dzbc003  = bd.dzbd002  ",
#               "                      and bc.dzbc005  = bd.dzbd004  ",
#               " where bc.dzbcstus = 'Y'                            ",
#               "   and bc.dzbc001  = '",ls_file_name,"'             ",
#              #"   and bc.dzbc002  = '",ls_version,"'               ",    #240422-00015#1
#               "   and bc.dzbc002  = ?              ",                    #240422-00015#1
#               "   and bc.dzbc007  = '",ls_dgenv,"'                 ",
#               "   and bd.dzbd004  <> 's'                           "
#  #End:200215-00009
#  DISPLAY "[DEBUG]sadzp030_tsd_get_section_if_modify:ls_sql=",ls_sql #tode             
#  PREPARE lpre_get_section_if_modify FROM ls_SQL
#  DECLARE lcur_get_section_if_modify CURSOR FOR lpre_get_section_if_modify
#
# #OPEN lcur_get_section_if_modify                                        #240422-00015#1
#  OPEN lcur_get_section_if_modify USING ls_version_num                   #240422-00015#1
#  FETCH lcur_get_section_if_modify INTO li_counts
#  CLOSE lcur_get_section_if_modify
#  
#  FREE lpre_get_section_if_modify
#  FREE lcur_get_section_if_modify  
#
#  LET ls_return = IIF(li_counts > 0,"Y","N")
#  
#  RETURN ls_return
#  
#END FUNCTION
##230925-00019#1 -e

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立<other>其他設定
# Input parameter : p_node_tsd <tsd> node
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2014/02/12 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_gen_other(p_node_tsd)
   DEFINE p_node_tsd xml.DomNode
   DEFINE lnode_other xml.DomNode,
          lnode_child xml.DomNode,
          ls_trigger STRING,
          ls_sql STRING,
          lr_dzax RECORD LIKE dzax_t.*
   DEFINE ls_upd STRING

   TRY
      LET ls_trigger = "sadzp030_tsd_gen_other : get ",gs_prog," other setting"
      LET ls_sql = "SELECT *",
                    " FROM dzax_t",
                   " WHERE dzax001='",gs_prog,"' AND dzax006='",gs_identity,"' AND dzaxstus='Y'" 
      DECLARE dzax_curs SCROLL CURSOR FROM ls_sql
      OPEN dzax_curs
      FETCH FIRST dzax_curs INTO lr_dzax.*
      CLOSE dzax_curs

      LET ls_upd = ""
      #Begin:2015/11/09 by Hiko
      #IF lr_dzax.dzax006=gs_identity THEN #2014/12/17 by Hiko
      #   IF lr_dzax.dzax004 CLIPPED="Y" THEN
      #      LET ls_upd = "u"
      #   END IF
      #END IF
      #End:2015/11/09 by Hiko

      LET lnode_other = p_node_tsd.appendChildElement("other")

      LET lnode_child = lnode_other.appendChildElement("code_template")
      CALL lnode_child.setAttribute("value", lr_dzax.dzax002 CLIPPED)
      CALL lnode_child.setAttribute("status", ls_upd) #2014/08/08 by Hiko

      LET lnode_child = lnode_other.appendChildElement("free_style")
      CALL lnode_child.setAttribute("value", lr_dzax.dzax003 CLIPPED)
      CALL lnode_child.setAttribute("status", ls_upd) #2014/08/08 by Hiko

      #Begin:2014/10/03 by Hiko
      LET lnode_child = lnode_other.appendChildElement("start_arg")
      CALL lnode_child.setAttribute("value", lr_dzax.dzax007 CLIPPED)
      CALL lnode_child.setAttribute("status", ls_upd)
      #End:2014/10/03 by Hiko

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立<toolbar>符合樣版的設定
# Input parameter : p_node_tsd <tsd> node
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2013/05/20 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_gen_toolbar(p_node_tsd)
   DEFINE p_node_tsd xml.DomNode
   DEFINE lnode_toolbar xml.DomNode,
          ls_trigger STRING,
          ls_sql STRING,
          la_dzfu DYNAMIC ARRAY OF RECORD LIKE dzfu_t.*,
          li_idx SMALLINT,
          lsb_dzfu002 base.StringBuffer #2014/02/12 by Hiko

   TRY
      LET lnode_toolbar = p_node_tsd.appendChildElement("toolbar")

      LET ls_trigger = "sadzp030_tsd_gen_toolbar : get ",gs_prog," toolbar setting"
      LET ls_sql = "SELECT *",
                    " FROM dzfu_t",
                   " WHERE dzfu001='",DOWNSHIFT(gs_prog_type),"' AND dzfustus='Y'" #2014/02/12 by Hiko
      PREPARE dzfu_prep FROM ls_sql
      DECLARE dzfu_curs CURSOR FOR dzfu_prep

      LET lsb_dzfu002 = base.StringBuffer.create()
      LET li_idx = 1
      FOREACH dzfu_curs INTO la_dzfu[li_idx].*
         IF lsb_dzfu002.getLength()>0 THEN
            CALL lsb_dzfu002.append(",")
         END IF

         CALL lsb_dzfu002.append(la_dzfu[li_idx].dzfu002 CLIPPED)

         LET li_idx = li_idx + 1
      END FOREACH

      CALL lnode_toolbar.setAttribute("items", lsb_dzfu002.toString())

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立<table>程式主Table設定.
# Input parameter : p_node_tsd <tsd> node
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2013/04/09 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_gen_table(p_node_tsd)
   DEFINE p_node_tsd xml.DomNode 
   DEFINE ls_trigger STRING,
          ls_sql STRING,
          #la_dzag DYNAMIC ARRAY OF RECORD LIKE dzag_t.*,
          la_dzag DYNAMIC ARRAY OF RECORD #2014/12/17 by Hiko
                                   dzag001 LIKE dzag_t.dzag001,
                                   dzag002 LIKE dzag_t.dzag002,
                                   dzag003 LIKE dzag_t.dzag003,
                                   dzag004 LIKE dzag_t.dzag004,
                                   dzag005 LIKE dzag_t.dzag005,
                                   dzag006 LIKE dzag_t.dzag006,
                                   dzag007 LIKE dzag_t.dzag007,
                                   dzag008 LIKE dzag_t.dzag008,
                                   dzag009 LIKE dzag_t.dzag009,
                                   dzag010 LIKE dzag_t.dzag010,
                                   dzag011 LIKE dzag_t.dzag011,
                                   dzag012 LIKE dzag_t.dzag012,
                                   dzag013 LIKE dzag_t.dzag013, #2014/12/29 by Hiko:三層二顯之上階Table
                                   dzag014 LIKE dzag_t.dzag014, #2014/12/29 by Hiko:三層二顯之上階Key值設定
                                   dzag015 LIKE dzag_t.dzag015, #2014/12/29 by Hiko:三層二顯之本階對應Key值設定
                                   dzagstus LIKE dzag_t.dzagstus,
                                   dzaa006 LIKE dzaa_t.dzaa006,
                                   dzaa009 LIKE dzaa_t.dzaa009
                                   END RECORD,
          li_idx SMALLINT,
          ls_revision STRING,
          ls_table_name STRING,
          ls_parent STRING,
          lnode_table xml.DomNode,
          lnode_tbl xml.DomNode
   DEFINE lnode_sr xml.DomNode,
          li_idx2 SMALLINT,
          la_dzfs DYNAMIC ARRAY OF RECORD LIKE dzfs_t.*
   DEFINE ls_main_sql STRING, #2014/08/05 by Hiko
          ls_upd STRING       #2014/08/05 by Hiko
   DEFINE ls_ch STRING #2015/11/09 by Hiko
		  
   TRY
      #不論有沒有'TABLE'資料,都要建立<table>.
      LET lnode_table = p_node_tsd.appendChildElement("table")
      CALL lnode_table.setAttribute("ver", "") #2013/12/13 by Hiko:保險作法
      CALL lnode_table.setAttribute("cite_std", "N") #2013/12/13 by Hiko:目前規劃都不引用

      LET ls_trigger = "sadzp030_tsd_gen_table : get ",gs_prog," table setting"
      LET ls_main_sql = "SELECT dzag001,dzag002,dzag003,dzag004,dzag005,", #2014/12/17 by Hiko
                               "dzag006,dzag007,dzag008,dzag009,dzag010,", 
                               "dzag011,dzag012,dzag013,dzag014,dzag015,", #2014/12/29 by Hiko
                               "dzagstus,dzaa006,dzaa009", #2014/12/17 by Hiko
                         " FROM dzaa_t"
      LET ls_sql = ls_main_sql,
                   " INNER JOIN dzag_t ON dzag001=dzaa001 AND dzag003=dzaa004 AND dzag006=dzaa006 AND dzagstus='Y'",
                   " WHERE dzaa001='",gs_prog,"' AND dzaa002=",g_revision," AND dzaa003='TABLE' AND dzaa005='4' AND dzaastus='Y' AND dzaa009='",gs_identity,"'",
                   #Begin:2015/11/09 by Hiko
                   #" UNION ",
                   #ls_main_sql,
                   #" INNER JOIN dzag_t ON dzag001=dzaa001 AND dzag003=dzaa004 AND dzag006=dzaa006 AND dzagstus='N'",
                   #" WHERE dzaa001='",gs_prog,"' AND dzaa002=",g_revision," AND dzaa003='TABLE' AND dzaa005='4' AND dzaastus='Y' AND dzaa009='",gs_identity,"'"
                   " ORDER BY dzag002"
                   #End:2015/11/09 by Hiko
      PREPARE dzag_prep FROM ls_sql
      DECLARE dzag_curs CURSOR FOR dzag_prep
      LET li_idx = 1
      FOREACH dzag_curs INTO la_dzag[li_idx].*
         #版次只需要設定一次.
         IF li_idx=1 THEN
            LET ls_revision = la_dzag[li_idx].dzag003 CLIPPED
            CALL lnode_table.setAttribute("ver", ls_revision)
         END IF

         LET lnode_tbl = lnode_table.appendChildElement("tbl")
         LET ls_table_name = la_dzag[li_idx].dzag002 CLIPPED
         CALL lnode_tbl.setAttribute("name", ls_table_name)
         CALL lnode_tbl.setAttribute("main", la_dzag[li_idx].dzag005 CLIPPED)
         #Begin:2014/01/21 by Hiko
         CALL lnode_tbl.setAttribute("head", la_dzag[li_idx].dzag007 CLIPPED)
         CALL lnode_tbl.setAttribute("pk", la_dzag[li_idx].dzag008 CLIPPED)
         CALL lnode_tbl.setAttribute("fk_detail", la_dzag[li_idx].dzag009 CLIPPED)
         LET ls_parent = la_dzag[li_idx].dzag004 CLIPPED
         CALL lnode_tbl.setAttribute("parent", ls_parent)
         CALL lnode_tbl.setAttribute("fk_master", la_dzag[li_idx].dzag010 CLIPPED)
         #End:2014/01/21 by Hiko
         CALL lnode_tbl.setAttribute("src", la_dzag[li_idx].dzag006 CLIPPED)
         #Begin:2014/12/29 by Hiko
         CALL lnode_tbl.setAttribute("upper_table", la_dzag[li_idx].dzag013 CLIPPED)
         CALL lnode_tbl.setAttribute("upper_key", la_dzag[li_idx].dzag014 CLIPPED)
         CALL lnode_tbl.setAttribute("this_key", la_dzag[li_idx].dzag015 CLIPPED)
         #End:2014/12/29 by Hiko
		 
         #Begin:2015/11/09 by Hiko
         #LET ls_upd = ""
         ##Begin:2014/12/17 by Hiko:dzaa006(使用標示)=dzaa009(客製標示)
         #IF la_dzag[li_idx].dzaa006=la_dzag[li_idx].dzaa009 AND
         #   la_dzag[li_idx].dzag003=g_revision THEN
         #   IF la_dzag[li_idx].dzagstus CLIPPED="Y" THEN
         #      IF g_revision>1 THEN 
         #         LET ls_upd = "u"
         #      END IF
         #   ELSE
         #      IF la_dzag[li_idx].dzagstus CLIPPED="N" THEN
         #         LET ls_upd = "d"
         #      END IF
         #   END IF
         #END IF
         ##End:2014/12/17 by Hiko
		 
         #CALL lnode_tbl.setAttribute("status", ls_upd)

         LET ls_ch = ""
         IF la_dzag[li_idx].dzaa006=la_dzag[li_idx].dzaa009 AND
            la_dzag[li_idx].dzag003=g_revision THEN
            IF g_revision>1 THEN 
               LET ls_ch = "Y"
            END IF
         END IF

         CALL lnode_tbl.setAttribute("ch", ls_ch)
         CALL lnode_tbl.setAttribute("status", "")
         #End:2015/11/09 by Hiko

         LET ls_trigger = "sadzp030_tsd_gen_table : get dzfs_t data"
         #dzfs_t的SQL並沒有特別拉出來為PUBLIC FUNCTION,因為要傳遞的參數比較多.
         LET ls_sql = "SELECT * FROM dzfs_t",
                      " WHERE dzfs002='",gs_prog,"' AND dzfs001='",ls_revision,"' AND dzfs004='",ls_table_name,"' AND dzfs005='",la_dzag[li_idx].dzag006 CLIPPED,"' AND dzfsstus='Y' ORDER BY dzfs003" #2015/11/09 by Hiko:增加dzfsstus='Y'的條件.
         PREPARE dzfs_prep FROM ls_sql
         DECLARE dzfs_curs CURSOR FOR dzfs_prep

         LET li_idx2 = 1
         FOREACH dzfs_curs INTO la_dzfs[li_idx2].*
            LET lnode_sr = lnode_tbl.appendChildElement("sr")
            CALL lnode_sr.setAttribute("name", la_dzfs[li_idx2].dzfs003 CLIPPED)
            CALL lnode_sr.setAttribute("src", la_dzfs[li_idx2].dzfs005 CLIPPED)
            #Begin:2013/08/29 by Hiko
            CALL lnode_sr.setAttribute("insert", la_dzfs[li_idx2].dzfs006 CLIPPED)
            CALL lnode_sr.setAttribute("delete", la_dzfs[li_idx2].dzfs007 CLIPPED)
            CALL lnode_sr.setAttribute("append", la_dzfs[li_idx2].dzfs008 CLIPPED)
            #End:2013/08/29
            #Begin:2014/01/21 by Hiko
            CALL lnode_sr.setAttribute("cascade", la_dzfs[li_idx2].dzfs009 CLIPPED)
            CALL lnode_sr.setAttribute("kind", la_dzfs[li_idx2].dzfs010 CLIPPED)
            #End:2014/01/21 by Hiko
			
            #Begin:2015/11/09 by Hiko
            #LET ls_upd = ""
            ##Begin:2014/12/17 by Hiko:dzaa006(使用標示)=dzaa009(客製標示)
            #IF la_dzag[li_idx].dzaa006=la_dzag[li_idx].dzaa009 AND
            #   la_dzfs[li_idx2].dzfs001 CLIPPED=g_revision THEN
            #   IF la_dzfs[li_idx2].dzfsstus CLIPPED="Y" THEN
            #      IF g_revision>1 THEN 
            #         LET ls_upd = "u"
            #      END IF
            #   ELSE
            #      IF la_dzfs[li_idx2].dzfsstus CLIPPED="N" THEN
            #         LET ls_upd = "d"
            #      END IF
            #   END IF
            #END IF
            ##End:2014/12/17 by Hiko

            #CALL lnode_sr.setAttribute("status", ls_upd)

            LET ls_ch = ""
            IF la_dzag[li_idx].dzaa006=la_dzag[li_idx].dzaa009 AND
               la_dzfs[li_idx2].dzfs001 CLIPPED=g_revision THEN
               IF g_revision>1 THEN 
                  LET ls_ch = "Y"
               END IF
            END IF

            CALL lnode_sr.setAttribute("ch", ls_ch)
            CALL lnode_sr.setAttribute("status", "")
            #End:2015/11/09 by Hiko

            LET li_idx2 = li_idx2 + 1
         END FOREACH

         LET li_idx = li_idx + 1
      END FOREACH

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立<all>規格設定與描述.
# Input parameter : p_doc_tsd <tsd> doc
#                   p_node_tsd <tsd> node
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2013/01/11 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_gen_all(p_doc_tsd, p_node_tsd)
   DEFINE p_doc_tsd xml.DomDocument,
          p_node_tsd xml.DomNode 
   DEFINE ls_trigger STRING,
          ls_sql STRING,
          la_dzab DYNAMIC ARRAY OF RECORD
                  dzaa004 LIKE dzaa_t.dzaa004,
                  dzaa006 LIKE dzaa_t.dzaa006,
                  dzaa003 LIKE dzaa_t.dzaa003,
                  dzab099 LIKE dzab_t.dzab099,
                  dzaa007 LIKE dzaa_t.dzaa007,
                  dzaa009 LIKE dzaa_t.dzaa009, #2014/12/17 by Hiko
                  dzaastus LIKE dzaa_t.dzaastus 
                  END RECORD,
          li_idx SMALLINT,
          lnode_all xml.DomNode,
          lnode_child xml.DomNode
   DEFINE lb_has_all BOOLEAN #2014/04/07 by Hiko
   DEFINE ls_upd STRING #2014/08/05 by Hiko
   DEFINE ls_ch STRING #2015/11/09 by Hiko

   TRY
      #dzab_t不會有刪除狀況,所以不需要剔除dzaastus='Y'的條件.
      LET ls_trigger = "sadzp030_tsd_gen_all : get ",gs_prog," overall spec"
      LET ls_sql = "SELECT dzaa004,dzaa006,dzaa003,dzab099,dzaa007,dzaa009,dzaastus",#2014/01/09 by Hiko
                    " FROM dzab_t",
                   " RIGHT JOIN dzaa_t ON dzaa001=dzab001 AND dzaa003=dzab004 AND dzaa004=dzab002 AND dzaa006=dzab003",
                   " WHERE dzaa001='",gs_prog,"' AND dzaa002=",g_revision," AND dzaa005='3' AND dzaastus='Y' AND dzaa009='",gs_identity,"' ORDER BY dzaa003"
      PREPARE dzab_prep FROM ls_sql
      LOCATE la_dzab[1].dzab099 IN FILE #在真正取得資料前,先設定陣列第一筆clob資料的LOCATE.
      DECLARE dzab_curs CURSOR FOR dzab_prep
      LET li_idx = 1
      LET lb_has_all = FALSE #2014/04/07 by Hiko
      FOREACH dzab_curs INTO la_dzab[li_idx].*
         CASE la_dzab[li_idx].dzaa003 CLIPPED
            WHEN "ALL" #整體規格:固定為"ALL"
               LET lb_has_all = TRUE #2014/04/07 by Hiko
               LET lnode_all = p_node_tsd.appendChildElement("all")
            WHEN "MI_ALL" #單頭編輯規格:固定為"MI_ALL"
               LET lnode_all = p_node_tsd.appendChildElement("mi_all")
            WHEN "DB_ALL" #單身瀏覽規格:固定為"DB_ALL"
               LET lnode_all = p_node_tsd.appendChildElement("db_all")
            WHEN "DI_ALL" #單身編輯規格:固定為"DI_ALL"
               LET lnode_all = p_node_tsd.appendChildElement("di_all")
         END CASE

         #Begin:2014/01/09 by Hiko
         IF NOT gb_industry THEN
            CALL lnode_all.setAttribute("cite_std", "N") #2014/12/19 by Hiko
         ELSE
            CALL lnode_all.setAttribute("cite_std", la_dzab[li_idx].dzaa007 CLIPPED)
         END IF
         CALL lnode_all.setAttribute("ver", la_dzab[li_idx].dzaa004 CLIPPED) #不紀錄引用版次,所以這裡的版次就是單純的版次.
         CALL lnode_all.setAttribute("src", la_dzab[li_idx].dzaa006 CLIPPED)

         LET lnode_child = p_doc_tsd.createCDATASection(la_dzab[li_idx].dzab099)
         CALL lnode_all.appendChild(lnode_child)
         
         #Begin:2015/11/09 by Hiko
         #LET ls_upd = ""
         ##Begin:2014/12/17 by Hiko:dzaa006(使用標示)=dzaa009(客製標示)
         #IF la_dzab[li_idx].dzaa006=la_dzab[li_idx].dzaa009 AND
         #   la_dzab[li_idx].dzaa004=g_revision THEN
         #   IF la_dzab[li_idx].dzaastus CLIPPED="Y" THEN
         #      IF g_revision>1 THEN 
         #         LET ls_upd = "u"
         #      END IF
         #   ELSE
         #      IF la_dzab[li_idx].dzaastus CLIPPED="N" THEN
         #         LET ls_upd = "d"
         #      END IF
         #   END IF
         #END IF
         ##End:2014/12/17 by Hiko
         
         #CALL lnode_all.setAttribute("status", ls_upd)

         LET ls_ch = ""
         IF la_dzab[li_idx].dzaa006=la_dzab[li_idx].dzaa009 AND
            la_dzab[li_idx].dzaa004=g_revision THEN
            IF g_revision>1 THEN 
               LET ls_ch = "Y"
            END IF
         END IF

         CALL lnode_all.setAttribute("ch", ls_ch)
         CALL lnode_all.setAttribute("status", "")
         #End:2015/11/09 by Hiko

         FREE la_dzab[li_idx].dzab099 #釋放LOCATE.
         LET li_idx = li_idx + 1
         LOCATE la_dzab[li_idx].dzab099 IN FILE #設定陣列第二筆以後clob資料的LOCATE.
         #End:2014/01/09 by Hiko
      END FOREACH

      #Begin:2014/04/07 by Hiko:因為rsd一定要有此結構,所以此段是為了和gen_rsd共用而寫
      IF NOT lb_has_all THEN
         LET lnode_all = p_node_tsd.appendChildElement("all")
         CALL lnode_all.setAttribute("cite_std", "N")
         CALL lnode_all.setAttribute("ver", "")
         CALL lnode_all.setAttribute("src", "")
         CALL lnode_all.setAttribute("status", "")
      END IF
      #End:2014/04/07 by Hiko

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立<tree>設定.
# Input parameter : p_node_tsd <tsd> node
#                 : p_kind 樹狀結構類別 #2013/09/13 by Hiko
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2013/01/11 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_gen_tree(p_node_tsd, p_kind)
   DEFINE p_node_tsd xml.DomNode,
          p_kind STRING #2013/09/13 by Hiko : TREE屬於樣板的一環,所以在r.a決定後,就不會有其他選擇,且目前產生器只支援一個固定TREE:s_browse.
   DEFINE ls_trigger STRING,
          ls_sql STRING,
          ls_tree_name STRING,
          la_dzaa DYNAMIC ARRAY OF RECORD LIKE dzaa_t.*,
          la_dzff DYNAMIC ARRAY OF RECORD
                  dzff004 LIKE dzff_t.dzff004,
                  dzff005 LIKE dzff_t.dzff005,
                  dzff006 LIKE dzff_t.dzff006,
                  dzff007 LIKE dzff_t.dzff007,
                  dzaa007 LIKE dzaa_t.dzaa007
                  END RECORD,
          li_idx5 SMALLINT, 
          li_idx SMALLINT, 
          lnode_tree xml.DomNode,
          lnode_att xml.DomNode
   #Begin:2014/04/16 by Hiko
   DEFINE li_idx3 SMALLINT,
          lstok_all_att base.StringTokenizer,
          ls_att STRING,
          ls_arr_att DYNAMIC ARRAY OF STRING,
          ls_arr_tmp DYNAMIC ARRAY OF STRING,
          ls_dzff005 STRING
   #End:2014/04/16 by Hiko
   DEFINE ls_upd STRING #2014/08/05 by Hiko
   DEFINE ls_ch STRING #2015/11/09 by Hiko

   TRY
      #Begin:2014/04/16 by Hiko:取得TREE的所有屬性,可讓使用者在設計器換TREE樣式
      LET lstok_all_att = base.StringTokenizer.create(sadzp168_1_dzff_default_str(ls_tree_name, "ALL"), ",")
      LET li_idx3 = 1
      WHILE lstok_all_att.hasMoreTokens()
         LET ls_arr_att[li_idx3] = lstok_all_att.nextToken()
         LET li_idx3 = li_idx3 + 1
      END WHILE
      #End:2014/04/16 by Hiko
      #支援多棵Tree的作法,保留以後的彈性.
      LET ls_trigger = "sadzp030_tsd_gen_tree : get ",gs_prog," tree setting"
      LET ls_sql = "SELECT * FROM dzaa_t WHERE dzaa001='",gs_prog,"' AND dzaa002=",g_revision," AND dzaa005='5' AND dzaa009='",gs_identity,"' AND dzaastus='Y' ORDER BY dzaa003" #2015/11/09 by Hiko:增加dzaastus='Y'的條件
      PREPARE dzaa_prep5 FROM ls_sql
      DECLARE dzaa_curs5 CURSOR FOR dzaa_prep5
      LET li_idx5 = 1
      FOREACH dzaa_curs5 INTO la_dzaa[li_idx5].*
         LET lnode_tree = p_node_tsd.appendChildElement("tree")
         LET ls_tree_name = la_dzaa[li_idx5].dzaa003 CLIPPED 
         #Begin:2014/04/16 by Hiko:先設定暫存陣列,在後面程式段會刪除
         CALL ls_arr_tmp.clear()
         FOR li_idx3=1 TO ls_arr_att.getLength()
            LET ls_arr_tmp[li_idx3] = ls_arr_att[li_idx3]
         END FOR
         #End:2014/04/16 by Hiko
         CALL lnode_tree.setAttribute("name", ls_tree_name) 
         CALL lnode_tree.setAttribute("ver", la_dzaa[li_idx5].dzaa004 CLIPPED)
         CALL lnode_tree.setAttribute("src", la_dzaa[li_idx5].dzaa006 CLIPPED)
         CALL lnode_tree.setAttribute("kind", p_kind)
         CALL lnode_tree.setAttribute("att", sadzp168_1_dzff_default_str(ls_tree_name, p_kind))
         IF NOT gb_industry THEN
            CALL lnode_tree.setAttribute("cite_std", "N") #2014/12/19 by Hiko
         ELSE
            CALL lnode_tree.setAttribute("cite_std", la_dzaa[li_idx5].dzaa007 CLIPPED)
         END IF
		 
         #Begin:2015/11/09 by Hiko
         #LET ls_upd = ""
         ##Begin:2014/12/17 by Hiko:dzaa006(使用標示)=dzaa009(客製標示)
         #IF la_dzaa[li_idx5].dzaa006=la_dzaa[li_idx5].dzaa009 AND
         #   la_dzaa[li_idx5].dzaa004=g_revision THEN
         #   IF la_dzaa[li_idx5].dzaastus CLIPPED="Y" THEN
         #      IF g_revision>1 THEN 
         #         LET ls_upd = "u"
         #      END IF
         #   ELSE
         #      IF la_dzaa[li_idx5].dzaastus CLIPPED="N" THEN
         #         LET ls_upd = "d"
         #      END IF
         #   END IF
         #END IF
         ##End:2014/12/17 by Hiko
	 #        
         #CALL lnode_tree.setAttribute("status", ls_upd)

         LET ls_ch = ""
         IF la_dzaa[li_idx5].dzaa006=la_dzaa[li_idx5].dzaa009 AND
            la_dzaa[li_idx5].dzaa004=g_revision THEN
            IF g_revision>1 THEN 
               LET ls_ch = "Y"
            END IF
         END IF
         CALL lnode_tree.setAttribute("ch", ls_ch)
         CALL lnode_tree.setAttribute("status", "")
         #End:2015/11/09 by Hiko

         LET ls_sql = "SELECT dzff004,dzff005,dzff006,dzff007",
                       " FROM dzff_t",
                      " WHERE dzff001='",la_dzaa[li_idx5].dzaa001 CLIPPED,"'",
                        " AND dzff002='",la_dzaa[li_idx5].dzaa004 CLIPPED,"'",
                        " AND dzff003='",ls_tree_name,"'",
                        " AND dzff008='",la_dzaa[li_idx5].dzaa006 CLIPPED,"'",
                      " ORDER BY dzff004" 
         PREPARE dzff_prep FROM ls_sql
         DECLARE dzff_curs CURSOR FOR dzff_prep
         LET li_idx = 1
         FOREACH dzff_curs INTO la_dzff[li_idx].*
            #Begin:2014/04/16 by Hiko
            LET ls_dzff005 = la_dzff[li_idx].dzff005 CLIPPED
            FOR li_idx3=1 TO ls_arr_tmp.getLength()
               IF ls_arr_tmp[li_idx3].equals(ls_dzff005) THEN
                  CALL ls_arr_tmp.deleteElement(li_idx3)
                  EXIT FOR
               END IF
            END FOR
            LET lnode_att = lnode_tree.appendChildElement(ls_dzff005)
            #End:2014/04/16 by Hiko
            CALL lnode_att.setAttribute("no", la_dzff[li_idx].dzff004 CLIPPED)
            CALL lnode_att.setAttribute("table", la_dzff[li_idx].dzff006 CLIPPED)
            CALL lnode_att.setAttribute("col", la_dzff[li_idx].dzff007 CLIPPED)
         
            LET li_idx = li_idx + 1
         END FOREACH #li_idx

         #Begin:2014/04/16 by Hiko
         FOR li_idx3=1 TO ls_arr_tmp.getLength()
            LET lnode_att = lnode_tree.appendChildElement(ls_arr_tmp[li_idx3])
            CALL lnode_att.setAttribute("no", "")
            CALL lnode_att.setAttribute("table", "")
            CALL lnode_att.setAttribute("col", "")
         END FOR
         #End:2014/04/16 by Hiko

         LET li_idx5 = li_idx5 + 1
      END FOREACH #li_idx5

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, ls_sql) 
      RETURN FALSE
   END TRY
   
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立<field>規格設定與描述.
# Input parameter : p_doc_tsd <tsd> doc
#                   p_node_tsd <tsd> node
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2013/01/11 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_gen_field(p_doc_tsd, p_node_tsd)
   DEFINE p_doc_tsd xml.DomDocument,
          p_node_tsd xml.DomNode 
   DEFINE ls_trigger STRING,
          ls_sql STRING,
          la_dzac DYNAMIC ARRAY OF RECORD
                  dzac002 LIKE dzac_t.dzac002,
                  dzaa003 LIKE dzaa_t.dzaa003,
                  dzaa004 LIKE dzaa_t.dzaa004,
                  dzac005 LIKE dzac_t.dzac005,
                  dzac006 LIKE dzac_t.dzac006,
                  dzac007 LIKE dzac_t.dzac007,
                  dzac008 LIKE dzac_t.dzac008,
                  dzac009 LIKE dzac_t.dzac009,
                  dzac010 LIKE dzac_t.dzac010,
                  dzac011 LIKE dzac_t.dzac011,
                  dzaa006 LIKE dzaa_t.dzaa006, #2014/04/14 by Hiko
                  dzac014 LIKE dzac_t.dzac014,
                  dzac015 LIKE dzac_t.dzac015, #2014/12/05 by Hiko:最大值
                  dzac016 LIKE dzac_t.dzac016, #2014/12/05 by Hiko:最小值
                  dzac017 LIKE dzac_t.dzac017,
                  dzac018 LIKE dzac_t.dzac018,
                  dzac019 LIKE dzac_t.dzac019,
                  dzac020 LIKE dzac_t.dzac020, #2014/12/05 by Hiko:最大值運算子
                  dzac021 LIKE dzac_t.dzac021, #2014/12/05 by Hiko:最小值運算子
                  dzac099 LIKE dzac_t.dzac099,
                  dzaa007 LIKE dzaa_t.dzaa007,
                  dzaa009 LIKE dzaa_t.dzaa009, #2014/12/17 by Hiko
                  dzaastus LIKE dzaa_t.dzaastus #2014/08/05 by Hiko
                  END RECORD,
          l_dzep011 LIKE dzep_t.dzep011,
          li_idx SMALLINT,
          lnode_field xml.DomNode,
          lnode_child xml.DomNode
   DEFINE ls_main_sql STRING, #2014/08/05 by Hiko
          ls_upd STRING       #2014/08/05 by Hiko
   DEFINE ls_ch STRING #2015/11/09 by Hiko

   TRY
      #將生/失效的SQL分開然後UNION. #2014/08/05 by Hiko
      LET ls_trigger = "sadzp030_tsd_gen_field : get ",gs_prog," column spec and setting"
      LET ls_main_sql = "SELECT dzac002,dzaa003,dzaa004,dzac005,dzac006,", #2014/01/09 by Hiko
                               "dzac007,dzac008,dzac009,dzac010,dzac011,",
                               "dzaa006,dzac014,dzac015,dzac016,dzac017,",
                               "dzac018,dzac019,dzac020,dzac021,dzac099,dzaa007,dzaa009,dzaastus",
                         " FROM dzac_t",
                        " RIGHT JOIN dzaa_t ON dzaa001=dzac001 AND dzaa003=dzac003 AND dzaa004=dzac004 AND dzaa006=dzac012"
      LET ls_sql = ls_main_sql,
                   " INNER JOIN dzfj_t ON dzfj001=dzaa001 AND dzfj002=dzaa002 AND dzfj005=dzaa003 AND dzfj017=dzaa009", #2014/04/02 by Hiko:剔除垃圾資料
                   " WHERE dzaa001='",gs_prog,"' AND dzaa002=",g_revision," AND dzaa005='1' AND dzaastus='Y' AND dzaa009='",gs_identity,"'",
                   #Begin:2015/11/09 by Hiko
                   #" UNION ALL ", #因為有CLOB,所以改成UNION ALL
                   #ls_main_sql,
                   #" WHERE dzaa001='",gs_prog,"' AND dzaa002=",g_revision," AND dzaa005='1' AND dzaastus='N' AND dzaa009='",gs_identity,"'"
                   " ORDER BY dzaa003"
                   #End:2015/11/09 by Hiko
      PREPARE dzac_prep FROM ls_sql
      LOCATE la_dzac[1].dzac099 IN FILE #在真正取得資料前,先設定陣列第一筆clob資料的LOCATE.
      DECLARE dzac_curs CURSOR FOR dzac_prep
      LET li_idx = 1
      FOREACH dzac_curs INTO la_dzac[li_idx].*
         LET lnode_field = p_node_tsd.appendChildElement("field")
         IF NOT gb_industry THEN
            CALL lnode_field.setAttribute("cite_std", "N") #2014/12/19 by Hiko
         ELSE
            CALL lnode_field.setAttribute("cite_std", la_dzac[li_idx].dzaa007 CLIPPED)
         END IF
         CALL lnode_field.setAttribute("name", la_dzac[li_idx].dzaa003 CLIPPED)       #控件名稱
         CALL lnode_field.setAttribute("ver", la_dzac[li_idx].dzaa004 CLIPPED)        #版次
         CALL lnode_field.setAttribute("src", la_dzac[li_idx].dzaa006 CLIPPED)        #使用標示
         CALL lnode_field.setAttribute("column", la_dzac[li_idx].dzac002 CLIPPED)     #欄位名稱
         CALL lnode_field.setAttribute("table", la_dzac[li_idx].dzac005 CLIPPED)      #Table名稱
         #Begin:2014/08/25 by Hiko
         #CALL lnode_field.setAttribute("attribute", la_dzac[li_idx].dzac006 CLIPPED)  #欄位屬性
         #CALL lnode_field.setAttribute("type", la_dzac[li_idx].dzac007 CLIPPED)       #資料型態:例如varchar2(10)
         CALL lnode_field.setAttribute("attribute", "") 
         CALL lnode_field.setAttribute("type", "")
         #End:2014/08/25 by Hiko
         CALL lnode_field.setAttribute("req", la_dzac[li_idx].dzac008 CLIPPED)        #必要欄位
         CALL lnode_field.setAttribute("i_zoom", la_dzac[li_idx].dzac009 CLIPPED)     #編輯時開窗
         CALL lnode_field.setAttribute("c_zoom", la_dzac[li_idx].dzac010 CLIPPED)     #查詢時開窗
         CALL lnode_field.setAttribute("chk_ref", la_dzac[li_idx].dzac011 CLIPPED)    #校驗帶值設定
         #取得adzi150預設的系統分類碼.
         LET ls_trigger = "sadzp030_tsd_gen_field : get scc setting"
         LET ls_sql = "SELECT dzep011 FROM dzep_t",
                      " WHERE dzep001='",la_dzac[li_idx].dzac005 CLIPPED,"'",
                        " AND dzep002='",la_dzac[li_idx].dzac002 CLIPPED,"'"
         PREPARE dzep_prep2 FROM ls_sql
         EXECUTE dzep_prep2 INTO l_dzep011
         FREE dzep_prep2
         
         CALL lnode_field.setAttribute("items", l_dzep011)                             #adzi150預設的系統分類碼
         CALL lnode_field.setAttribute("default", la_dzac[li_idx].dzac014 CLIPPED)     #預設值
         CALL lnode_field.setAttribute("max_compare", la_dzac[li_idx].dzac020 CLIPPED) #最大值運算子
         CALL lnode_field.setAttribute("max", la_dzac[li_idx].dzac015 CLIPPED)         #最大值
         CALL lnode_field.setAttribute("min_compare", la_dzac[li_idx].dzac021 CLIPPED) #最小值運算子
         CALL lnode_field.setAttribute("min", la_dzac[li_idx].dzac016 CLIPPED)         #最小值
         CALL lnode_field.setAttribute("can_edit", la_dzac[li_idx].dzac017 CLIPPED)    #是否可編輯
         CALL lnode_field.setAttribute("can_query", la_dzac[li_idx].dzac018 CLIPPED)   #是否可查詢
         CALL lnode_field.setAttribute("widget", la_dzac[li_idx].dzac019 CLIPPED)      #顯示控件
         LET lnode_child = p_doc_tsd.createCDATASection(la_dzac[li_idx].dzac099)
         CALL lnode_field.appendChild(lnode_child)

         #Begin:2014/04/18 by Hiko
         IF NOT sadzp030_tsd_collect_prog_rel(la_dzac[li_idx].dzaa007 CLIPPED, gs_prog, la_dzac[li_idx].dzaa003 CLIPPED, la_dzac[li_idx].dzaa004 CLIPPED, la_dzac[li_idx].dzaa006 CLIPPED, la_dzac[li_idx].dzaa009 CLIPPED) THEN
            RETURN FALSE
         END IF
         #End:2014/04/18 by Hiko

         #Begin:2015/11/09 by Hiko
         #LET ls_upd = ""
         ##Begin:2014/12/17 by Hiko:dzaa006(使用標示)=dzaa009(客製標示)
         #IF la_dzac[li_idx].dzaa006=la_dzac[li_idx].dzaa009 AND
         #   la_dzac[li_idx].dzaa004=g_revision THEN
         #   IF la_dzac[li_idx].dzaastus CLIPPED="Y" THEN
         #      IF g_revision>1 THEN 
         #         LET ls_upd = "u"
         #      END IF
         #   ELSE
         #      IF la_dzac[li_idx].dzaastus CLIPPED="N" THEN
         #         LET ls_upd = "d"
         #      END IF
         #   END IF
         #END IF
         ##End:2014/12/17 by Hiko
         
         #CALL lnode_field.setAttribute("status", ls_upd)

         LET ls_ch = ""
         IF la_dzac[li_idx].dzaa006=la_dzac[li_idx].dzaa009 AND
            la_dzac[li_idx].dzaa004=g_revision THEN
            IF g_revision>1 THEN 
               LET ls_ch = "Y"
            END IF
         END IF

         CALL lnode_field.setAttribute("ch", ls_ch)
         CALL lnode_field.setAttribute("status", "")
         #End:2015/11/09 by Hiko

         FREE la_dzac[li_idx].dzac099 #釋放LOCATE.
         LET li_idx = li_idx + 1
         LOCATE la_dzac[li_idx].dzac099 IN FILE #設定陣列第二筆以後clob資料的LOCATE.
         #End:2014/01/09 by Hiko
      END FOREACH

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 收集<prog_rel>規格設定.
# Input parameter : p_cite_std 是否引用
#                   p_dzal001 程式代號
#                   p_dzal005 依附控件編號
#                   p_dzal003 識別碼版次
#                   p_dzal004 使用標示
#                   p_dzaa009 客製標示
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2014/04/18 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_collect_prog_rel(p_cite_std, p_dzal001, p_dzal005, p_dzal003, p_dzal004, p_dzaa009)
   DEFINE p_cite_std STRING,#dzaa007
          p_dzal001 STRING, #dzaa001
          p_dzal005 STRING, #dzaa003
          #p_dzal003 STRING, #dzaa004
          p_dzal003 LIKE dzal_t.dzal003, #dzaa004
          p_dzal004 STRING, #dzaa006,
          p_dzaa009 STRING  #2014/12/17 by Hiko
   DEFINE ls_trigger STRING,
          ls_where STRING,
          ls_sql STRING,
          li_cnt SMALLINT,
          li_idx SMALLINT
   #Begin:2014/12/17 by Hiko
   DEFINE la_dzal   T_DZAL_T,
          li_tmp_i  SMALLINT,
          li_tmp_j  SMALLINT,
          lb_continue BOOLEAN,
          li_length SMALLINT 
   #End:2014/12/17 by Hiko

   TRY
      #Begin:2014/10/31 by Hiko
      IF cl_null(p_dzal003) THEN
         LET p_dzal003 = 0
      END IF
      #End:2014/10/31 by Hiko
      #重點在於dzaa003(控件編號)=dzal005(依附控件編號)
      LET ls_trigger = "sadzp030_tsd_collect_prog_rel : get dzal_t setting"      
      LET ls_where = " WHERE dzal001='",p_dzal001,"' AND dzal005='",p_dzal005,"' AND dzal003=",p_dzal003," AND dzal004='",p_dzal004,"' AND dzalstus='Y'" #2014/12/30 by Hiko:增加dzalstus='Y'的條件
      LET ls_sql = "SELECT COUNT(*)", 
                    " FROM dzal_t",ls_where
      PREPARE dzal_prep0 FROM ls_sql
      EXECUTE dzal_prep0 INTO li_cnt
      FREE dzal_prep0
      IF li_cnt>0 THEN 
         LET ls_sql = "SELECT '',dzal002,dzal003,dzal004,dzal005,dzal006,dzal007,dzal008,dzal009,dzalstus", 
                       " FROM dzal_t",ls_where,
                      " ORDER BY dzal002,dzal008" #一個控件代號可以設定多組程式串查
         PREPARE dzal_prep1 FROM ls_sql
         DECLARE dzal_curs1 CURSOR FOR dzal_prep1
         #Begin:2014/12/17 by Hiko:防止相同控件編號與相同依附控件編號的問題出現
         LET li_tmp_i = 1
         FOREACH dzal_curs1 INTO la_dzal[li_tmp_i].*
            LET lb_continue = TRUE
            LET li_length = ga_dzal.getLength()
            FOR li_tmp_j=1 TO li_length
               IF la_dzal[li_tmp_i].dzal002=ga_dzal[li_tmp_j].dzal002 AND 
                  la_dzal[li_tmp_i].dzal005=ga_dzal[li_tmp_j].dzal005 AND 
                  la_dzal[li_tmp_i].dzal008=ga_dzal[li_tmp_j].dzal008 THEN
                  LET lb_continue = FALSE
                  EXIT FOR
               END IF
            END FOR

            IF lb_continue THEN
               LET li_length = li_length + 1
               LET ga_dzal[li_length].* = la_dzal[li_tmp_i].* 
               LET ga_dzal[li_length].cite_std = p_cite_std
               LET ga_dzal[li_length].dzaa009 = p_dzaa009
            END IF

            LET li_tmp_i = li_tmp_i + 1
         END FOREACH
         #End:2014/12/17 by Hiko
      END IF

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立<prog_rel>規格設定.
# Input parameter : p_node_tsd <tsd> node
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2013/04/09 by Hiko
# Modify          : 2014/03/14 by Hiko:程式串查格式改變,整段重寫
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_gen_prog_rel(p_node_tsd)
   DEFINE p_node_tsd xml.DomNode 
   DEFINE ls_dzal002 STRING, #2014/12/17 by Hiko:判斷方式從dzal005改成dzal002
          ls_dzal002_tmp STRING,
          li_idx2 SMALLINT,
          lnode_prog_rel xml.DomNode,
          lnode_child xml.DomNode,
          lnode_program xml.DomNode
   DEFINE ls_upd STRING       #2014/08/05 by Hiko
   DEFINE ls_ch STRING #2015/11/09 by Hiko
   
   TRY
      LET lnode_prog_rel = p_node_tsd.appendChildElement("prog_rel")
      FOR li_idx2=1 TO ga_dzal.getLength()
         LET ls_dzal002 = ga_dzal[li_idx2].dzal002 CLIPPED
         #Begin:2014/12/17 by Hiko
         IF cl_null(ls_dzal002) THEN
            CONTINUE FOR
         END IF
         #End:2014/12/17 by Hiko
         
         IF NOT ls_dzal002.equals(ls_dzal002_tmp) THEN
            LET lnode_child = lnode_prog_rel.appendChildElement("pfield")
            IF NOT gb_industry THEN
               CALL lnode_child.setAttribute("cite_std", "N") #2014/12/19 by Hiko
            ELSE
               CALL lnode_child.setAttribute("cite_std", ga_dzal[li_idx2].cite_std CLIPPED)
            END IF
            CALL lnode_child.setAttribute("ver", ga_dzal[li_idx2].dzal003 CLIPPED)  #版本
            CALL lnode_child.setAttribute("name", ls_dzal002) #控件名稱
            CALL lnode_child.setAttribute("depend_field", ga_dzal[li_idx2].dzal005 CLIPPED) #依附控件名稱
            CALL lnode_child.setAttribute("src", ga_dzal[li_idx2].dzal004 CLIPPED)  #使用標示
			
            #Begin:2015/11/09 by Hiko
            #LET ls_upd = ""
            ##Begin:2014/12/17 by Hiko:dzaa006(使用標示)=dzaa009(客製標示)
            #IF ga_dzal[li_idx2].dzal004=ga_dzal[li_idx2].dzaa009 AND
            #   ga_dzal[li_idx2].dzal003=g_revision THEN
            #   IF ga_dzal[li_idx2].dzalstus CLIPPED="Y" THEN
            #      IF g_revision>1 THEN 
            #         LET ls_upd = "u"
            #      END IF
            #   ELSE
            #      IF ga_dzal[li_idx2].dzalstus CLIPPED="N" THEN
            #         LET ls_upd = "d"
            #      END IF
            #   END IF
            #END IF
            ##End:2014/12/17 by Hiko
            
            #CALL lnode_child.setAttribute("status", ls_upd)

            LET ls_ch = ""
            IF ga_dzal[li_idx2].dzal004=ga_dzal[li_idx2].dzaa009 AND
               ga_dzal[li_idx2].dzal003=g_revision THEN
               IF g_revision>1 THEN 
                  LET ls_ch = "Y"
               END IF
            END IF

            CALL lnode_child.setAttribute("ch", ls_ch)
            CALL lnode_child.setAttribute("status", "")
            #End:2015/11/09 by Hiko
         
            LET ls_dzal002_tmp = ls_dzal002
         END IF

         LET lnode_program = lnode_child.appendChildElement("program")
         CALL lnode_program.setAttribute("order", ga_dzal[li_idx2].dzal008 CLIPPED) #項次
         CALL lnode_program.setAttribute("type", ga_dzal[li_idx2].dzal007 CLIPPED)  #串查型態
         CALL lnode_program.setAttribute("name", ga_dzal[li_idx2].dzal006 CLIPPED)  #程式編號
      END FOR 

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch("", "") 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立<ref_field>規格設定.
# Input parameter : p_node_tsd <tsd> node
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2013/04/09 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_gen_ref_field(p_node_tsd)
   DEFINE p_node_tsd xml.DomNode 
   DEFINE ls_trigger STRING,
          ls_sql STRING,
          #Begin:2014/01/09 by Hiko
          la_dzai DYNAMIC ARRAY OF RECORD
                  dzaa003 LIKE dzaa_t.dzaa003,
                  dzaa004 LIKE dzaa_t.dzaa004,
                  dzaa006 LIKE dzaa_t.dzaa006,
                  dzai005 LIKE dzai_t.dzai005,
                  dzai007 LIKE dzai_t.dzai007,
                  dzai008 LIKE dzai_t.dzai008,
                  dzai009 LIKE dzai_t.dzai009,
                  dzai010 LIKE dzai_t.dzai010,
                  dzai011 LIKE dzai_t.dzai011,
                  dzaa007 LIKE dzaa_t.dzaa007,
                  dzaa009 LIKE dzaa_t.dzaa009, #2014/12/17 by Hiko
                  dzaastus LIKE dzaa_t.dzaastus #2014/08/05 by Hiko
                  END RECORD,
          #End:2014/01/09 by Hiko
          li_idx SMALLINT,
          lnode_ref_field xml.DomNode,
          lnode_child xml.DomNode
   DEFINE ls_main_sql STRING, #2014/08/05 by Hiko
          ls_upd STRING       #2014/08/05 by Hiko
   DEFINE ls_ch STRING #2015/11/09 by Hiko

   TRY
      LET lnode_ref_field = p_node_tsd.appendChildElement("ref_field")

      LET ls_trigger = "sadzp030_tsd_gen_ref_field : get ",gs_prog," column reference setting"
      LET ls_main_sql = "SELECT dzaa003,dzaa004,dzaa006,dzai005,dzai007,", #2014/01/09 by Hiko
                               "dzai008,dzai009,dzai010,dzai011,dzaa007,dzaa009,dzaastus",
                         " FROM dzai_t",
                        " RIGHT JOIN dzaa_t ON dzaa001=dzai001 AND dzaa003=dzai002 AND dzaa004=dzai003 AND dzaa006=dzai004"
      LET ls_sql = ls_main_sql,
                   " INNER JOIN dzfj_t ON dzfj001=dzaa001 AND dzfj002=dzaa002 AND dzfj005=dzaa003 AND dzfj017=dzaa009", #2014/04/02 by Hiko:剔除垃圾資料
                   " WHERE dzaa001='",gs_prog,"' AND dzaa002=",g_revision," AND dzaa005='6' AND dzaastus='Y' AND dzaa009='",gs_identity,"'",
                   #Begin:2015/11/09 by Hiko
                   #" UNION ",
                   #ls_main_sql,
                   #" WHERE dzaa001='",gs_prog,"' AND dzaa002=",g_revision," AND dzaa005='6' AND dzaastus='N' AND dzaa009='",gs_identity,"'"
                   " ORDER BY dzaa003"
                   #End:2015/11/09 by Hiko
      #End:2014/04/14 by Hiko
      PREPARE dzai_prep FROM ls_sql
      DECLARE dzai_curs CURSOR FOR dzai_prep
      LET li_idx = 1
      FOREACH dzai_curs INTO la_dzai[li_idx].*
         LET lnode_child = lnode_ref_field.appendChildElement("rfield")
         IF NOT gb_industry THEN
            CALL lnode_child.setAttribute("cite_std", "N") #2014/12/19 by Hiko
         ELSE
            CALL lnode_child.setAttribute("cite_std", la_dzai[li_idx].dzaa007 CLIPPED)
         END IF
         CALL lnode_child.setAttribute("ver", la_dzai[li_idx].dzaa004 CLIPPED)           #版本
         CALL lnode_child.setAttribute("name", la_dzai[li_idx].dzaa003 CLIPPED)          #控件名稱
         CALL lnode_child.setAttribute("src", la_dzai[li_idx].dzaa006 CLIPPED)           #使用標示
         CALL lnode_child.setAttribute("depend_field", la_dzai[li_idx].dzai005 CLIPPED)  #依附控件名稱
         CALL lnode_child.setAttribute("correspon_key", la_dzai[li_idx].dzai007 CLIPPED) #對應傳值設定
         CALL lnode_child.setAttribute("ref_table", la_dzai[li_idx].dzai008 CLIPPED)     #資料參考Table
         CALL lnode_child.setAttribute("ref_fk", la_dzai[li_idx].dzai009 CLIPPED)        #資料參考FK
         CALL lnode_child.setAttribute("ref_dlang", la_dzai[li_idx].dzai010 CLIPPED)     #資料參考語系
         CALL lnode_child.setAttribute("ref_rtn", la_dzai[li_idx].dzai011 CLIPPED)       #資料參考回傳

         #Begin:2015/11/09 by Hiko
         #LET ls_upd = ""
         ##Begin:2014/12/17 by Hiko:dzaa006(使用標示)=dzaa009(客製標示)
         #IF la_dzai[li_idx].dzaa006=la_dzai[li_idx].dzaa009 AND
         #   la_dzai[li_idx].dzaa004=g_revision THEN
         #   IF la_dzai[li_idx].dzaastus CLIPPED="Y" THEN
         #      IF g_revision>1 THEN 
         #         LET ls_upd = "u"
         #      END IF
         #   ELSE
         #      IF la_dzai[li_idx].dzaastus CLIPPED="N" THEN
         #         LET ls_upd = "d"
         #      END IF
         #   END IF
         #END IF
         ##End:2014/12/17 by Hiko
	 #        
         #CALL lnode_child.setAttribute("status", ls_upd)

         LET ls_ch = ""
         IF la_dzai[li_idx].dzaa006=la_dzai[li_idx].dzaa009 AND
            la_dzai[li_idx].dzaa004=g_revision THEN
            IF g_revision>1 THEN 
               LET ls_ch = "Y"
            END IF
         END IF

         CALL lnode_child.setAttribute("ch", ls_ch)
         CALL lnode_child.setAttribute("status", "")
         #End:2015/11/09 by Hiko
         
         LET li_idx = li_idx + 1
      END FOREACH

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立<multi_lang>規格設定.
# Input parameter : p_doc_tsd <tsd> doc
#                 : p_node_tsd <tsd> node
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2013/04/09 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_gen_multi_lang(p_doc_tsd, p_node_tsd)
   DEFINE p_doc_tsd xml.DomDocument, #2013/09/24 by Hiko
          p_node_tsd xml.DomNode 
   DEFINE ls_trigger STRING,
          ls_sql STRING,
          #Begin:2014/01/09 by Hiko
          la_dzaj DYNAMIC ARRAY OF RECORD
                  dzaa003 LIKE dzaa_t.dzaa003,
                  dzaa004 LIKE dzaa_t.dzaa004,
                  dzaa006 LIKE dzaa_t.dzaa006,
                  dzaj005 LIKE dzaj_t.dzaj005,
                  dzaj007 LIKE dzaj_t.dzaj007,
                  dzaj008 LIKE dzaj_t.dzaj008,
                  dzaj009 LIKE dzaj_t.dzaj009,
                  dzaj010 LIKE dzaj_t.dzaj010,
                  dzaj011 LIKE dzaj_t.dzaj011,
                  dzaj099 LIKE dzaj_t.dzaj099,
                  dzaa007 LIKE dzaa_t.dzaa007,
                  dzaa009 LIKE dzaa_t.dzaa009, #2014/12/17 by Hiko
                  dzaastus LIKE dzaa_t.dzaastus #2014/08/05 by Hiko
                  END RECORD,
          #End:2014/01/09 by Hiko
          li_idx SMALLINT,
          lnode_multi_lang xml.DomNode,
          lnode_child xml.DomNode,
          lnode_spec xml.DomNode #2013/09/24 by Hiko
   DEFINE ls_main_sql STRING, #2014/08/05 by Hiko
          ls_upd STRING       #2014/08/05 by Hiko
   DEFINE ls_ch STRING #2015/11/09 by Hiko

   TRY
      LET lnode_multi_lang = p_node_tsd.appendChildElement("multi_lang")

      LET ls_trigger = "sadzp030_tsd_gen_multi_lang : get ",gs_prog," column multi-lang setting"
      LET ls_main_sql = "SELECT dzaa003,dzaa004,dzaa006,dzaj005,dzaj007,", #2014/01/09 by Hiko
                               "dzaj008,dzaj009,dzaj010,dzaj011,dzaj099,dzaa007,dzaa009,dzaastus",
                         " FROM dzaj_t",
                        " RIGHT JOIN dzaa_t ON dzaa001=dzaj001 AND dzaa003=dzaj002 AND dzaa004=dzaj003 AND dzaa006=dzaj004"
      LET ls_sql = ls_main_sql,
                   " INNER JOIN dzfj_t ON dzfj001=dzaa001 AND dzfj002=dzaa002 AND dzfj005=dzaa003 AND dzfj017=dzaa009", #2014/04/02 by Hiko:剔除垃圾資料
                   " WHERE dzaa001='",gs_prog,"' AND dzaa002=",g_revision," AND dzaa005='7' AND dzaastus='Y' AND dzaa009='",gs_identity,"'",
                   #Begin:2015/11/09 by Hiko
                   #" UNION ALL ",
                   #ls_main_sql,
                   #" WHERE dzaa001='",gs_prog,"' AND dzaa002=",g_revision," AND dzaa005='7' AND dzaastus='N' AND dzaa009='",gs_identity,"'"
                   " ORDER BY dzaa003"
                   #End:2015/11/09 by Hiko
      PREPARE dzaj_prep FROM ls_sql
      LOCATE la_dzaj[1].dzaj099 IN FILE #在真正取得資料前,先設定陣列第一筆clob資料的LOCATE.
      DECLARE dzaj_curs CURSOR FOR dzaj_prep
      LET li_idx = 1
      FOREACH dzaj_curs INTO la_dzaj[li_idx].*
         LET lnode_child = lnode_multi_lang.appendChildElement("mfield")
         IF NOT gb_industry THEN
            CALL lnode_child.setAttribute("cite_std", "N") #2014/12/19 by Hiko
         ELSE
            CALL lnode_child.setAttribute("cite_std", la_dzaj[li_idx].dzaa007 CLIPPED)
         END IF
         CALL lnode_child.setAttribute("ver", la_dzaj[li_idx].dzaa004 CLIPPED)  #版本
         CALL lnode_child.setAttribute("name", la_dzaj[li_idx].dzaa003 CLIPPED) #控件名稱
         CALL lnode_child.setAttribute("src", la_dzaj[li_idx].dzaa006 CLIPPED)  #使用標示
         CALL lnode_child.setAttribute("depend_field", la_dzaj[li_idx].dzaj005 CLIPPED)  #依附控件名稱
         CALL lnode_child.setAttribute("correspon_key", la_dzaj[li_idx].dzaj007 CLIPPED) #對應傳值設定
         CALL lnode_child.setAttribute("lang_table", la_dzaj[li_idx].dzaj008 CLIPPED)    #多語言Table
         CALL lnode_child.setAttribute("lang_fk", la_dzaj[li_idx].dzaj009 CLIPPED)       #多語言FK
         CALL lnode_child.setAttribute("lang_dlang", la_dzaj[li_idx].dzaj010 CLIPPED)    #多語言語系
         CALL lnode_child.setAttribute("lang_rtn", la_dzaj[li_idx].dzaj011 CLIPPED)      #多語言回傳

         LET lnode_spec = p_doc_tsd.createCDATASection(la_dzaj[li_idx].dzaj099)
         CALL lnode_child.appendChild(lnode_spec)
		 
         #Begin:2015/11/09 by Hiko
         #LET ls_upd = ""
         ##Begin:2014/12/17 by Hiko:dzaa006(使用標示)=dzaa009(客製標示)
         #IF la_dzaj[li_idx].dzaa006=la_dzaj[li_idx].dzaa009 AND
         #   la_dzaj[li_idx].dzaa004=g_revision THEN
         #   IF la_dzaj[li_idx].dzaastus CLIPPED="Y" THEN
         #      IF g_revision>1 THEN 
         #         LET ls_upd = "u"
         #      END IF
         #   ELSE
         #      IF la_dzaj[li_idx].dzaastus CLIPPED="N" THEN
         #         LET ls_upd = "d"
         #      END IF
         #   END IF
         #END IF
         ##End:2014/12/17 by Hiko
	 #        
         #CALL lnode_child.setAttribute("status", ls_upd)

         LET ls_ch = ""
         IF la_dzaj[li_idx].dzaa006=la_dzaj[li_idx].dzaa009 AND
            la_dzaj[li_idx].dzaa004=g_revision THEN
            IF g_revision>1 THEN 
               LET ls_ch = "Y"
            END IF
         END IF

         CALL lnode_child.setAttribute("ch", ls_ch)
         CALL lnode_child.setAttribute("status", "")
         #End:2015/11/09 by Hiko

         FREE la_dzaj[li_idx].dzaj099 #釋放LOCATE.
         LET li_idx = li_idx + 1
         LOCATE la_dzaj[li_idx].dzaj099 IN FILE #設定陣列第二筆以後clob資料的LOCATE.
      END FOREACH

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立<help_code>規格設定.
# Input parameter : p_node_tsd <tsd> node
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2014/03/06 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_gen_help_code(p_node_tsd)
   DEFINE p_node_tsd xml.DomNode 
   DEFINE ls_trigger STRING,
          ls_sql STRING,
          la_dzak DYNAMIC ARRAY OF RECORD
                  dzaa003 LIKE dzaa_t.dzaa003,
                  dzaa004 LIKE dzaa_t.dzaa004,
                  dzaa006 LIKE dzaa_t.dzaa006,
                  dzak005 LIKE dzak_t.dzak005,
                  dzak007 LIKE dzak_t.dzak007,
                  dzak008 LIKE dzak_t.dzak008,
                  dzak009 LIKE dzak_t.dzak009,
                  dzak010 LIKE dzak_t.dzak010,
                  dzak011 LIKE dzak_t.dzak011,
                  dzaa007 LIKE dzaa_t.dzaa007,
                  dzaa009 LIKE dzaa_t.dzaa009, #2014/12/17 by Hiko
                  dzaastus LIKE dzaa_t.dzaastus #2014/08/05 by Hiko
                  END RECORD,
          li_idx SMALLINT,
          lnode_help_code xml.DomNode,
          lnode_child xml.DomNode
   DEFINE ls_main_sql STRING, #2014/08/05 by Hiko
          ls_upd STRING       #2014/08/05 by Hiko
   DEFINE ls_ch STRING #2015/11/09 by Hiko

   TRY
      LET lnode_help_code = p_node_tsd.appendChildElement("help_code")
      #這裡的SQL出來會有多筆hfield,但這是為了設計器點選控件時,能夠顯現註記碼設定而補上的.
      LET ls_trigger = "sadzp030_tsd_gen_help_code : get ",gs_prog," column help code setting"
      LET ls_main_sql = "SELECT dzaa003,dzaa004,dzaa006,dzak005,dzak007,dzak008,dzak009,dzak010,dzak011,dzaa007,dzaa009,dzaastus",
                         " FROM dzak_t",
                        " RIGHT JOIN dzaa_t ON dzaa001=dzak001 AND dzaa003=dzak002 AND dzaa004=dzak003 AND dzaa006=dzak004" #註記碼直接和dzaa_t關聯即可.
      LET ls_sql = ls_main_sql,
                   " INNER JOIN dzfj_t ON dzfj001=dzaa001 AND dzfj002=dzaa002 AND dzfj005=dzaa003 AND dzfj017=dzaa009", #2014/04/02 by Hiko:剔除垃圾資料
                   " WHERE dzaa001='",gs_prog,"' AND dzaa002=",g_revision," AND dzaa005='1' AND dzaastus='Y' AND dzaa009='",gs_identity,"'",
                   #Begin:2015/11/09 by Hiko
                   #" UNION ",
                   #ls_main_sql,
                   #" WHERE dzaa001='",gs_prog,"' AND dzaa002=",g_revision," AND dzaa005='1' AND dzaastus='N' AND dzaa009='",gs_identity,"'"
                   " ORDER BY dzaa003"
                   #End:2015/11/09 by Hiko
      PREPARE dzak_prep FROM ls_sql
      DECLARE dzak_curs CURSOR FOR dzak_prep
      LET li_idx = 1
      FOREACH dzak_curs INTO la_dzak[li_idx].*
         LET lnode_child = lnode_help_code.appendChildElement("hfield")
         IF NOT gb_industry THEN
            CALL lnode_child.setAttribute("cite_std", "N") #2014/12/19 by Hiko
         ELSE
            CALL lnode_child.setAttribute("cite_std", la_dzak[li_idx].dzaa007 CLIPPED)
         END IF
         CALL lnode_child.setAttribute("ver", la_dzak[li_idx].dzaa004 CLIPPED)  #版本
         CALL lnode_child.setAttribute("name", la_dzak[li_idx].dzaa003 CLIPPED) #控件名稱
         CALL lnode_child.setAttribute("src", la_dzak[li_idx].dzaa006 CLIPPED)  #使用標示
         CALL lnode_child.setAttribute("help_table", la_dzak[li_idx].dzak008 CLIPPED)     #助記碼Table
         CALL lnode_child.setAttribute("help_find", la_dzak[li_idx].dzak007 CLIPPED)      #助記碼搜尋欄位
         CALL lnode_child.setAttribute("help_dlang", la_dzak[li_idx].dzak010 CLIPPED)     #助記碼語系
         CALL lnode_child.setAttribute("help_field", la_dzak[li_idx].dzak009 CLIPPED)     #助記碼欄位
         CALL lnode_child.setAttribute("help_wc", la_dzak[li_idx].dzak005 CLIPPED)        #其他條件
         CALL lnode_child.setAttribute("mapping_widget", la_dzak[li_idx].dzak011 CLIPPED) #回傳對應控件
		 
         #Begin:2015/11/09 by Hiko
         #LET ls_upd = ""
         ##Begin:2014/12/17 by Hiko:dzaa006(使用標示)=dzaa009(客製標示)
         #IF la_dzak[li_idx].dzaa006=la_dzak[li_idx].dzaa009 AND
         #   la_dzak[li_idx].dzaa004=g_revision THEN
         #   IF la_dzak[li_idx].dzaastus CLIPPED="Y" THEN
         #      IF g_revision>1 THEN 
         #         LET ls_upd = "u"
         #      END IF
         #   ELSE
         #      IF la_dzak[li_idx].dzaastus CLIPPED="N" THEN
         #         LET ls_upd = "d"
         #      END IF
         #   END IF
         #END IF
         ##End:2014/12/17 by Hiko
	 #        
         #CALL lnode_child.setAttribute("status", ls_upd)

         LET ls_ch = ""
         IF la_dzak[li_idx].dzaa006=la_dzak[li_idx].dzaa009 AND
            la_dzak[li_idx].dzaa004=g_revision THEN
            IF g_revision>1 THEN 
               LET ls_ch = "Y"
            END IF
         END IF

         CALL lnode_child.setAttribute("ch", ls_ch)
         CALL lnode_child.setAttribute("status", "")
         #Begin:2015/11/09 by Hiko
         
         LET li_idx = li_idx + 1
      END FOREACH

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立<act>規格設定與描述.
# Input parameter : p_doc_tsd <tsd> doc
#                   p_node_tsd <tsd> node
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2013/01/11 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_gen_act(p_doc_tsd, p_node_tsd)
   DEFINE p_doc_tsd xml.DomDocument,
          p_node_tsd xml.DomNode 
   DEFINE ls_trigger STRING,
          ls_sql STRING,
          la_dzad DYNAMIC ARRAY OF RECORD
                  dzaa003 LIKE dzaa_t.dzaa003,
                  dzaa004 LIKE dzaa_t.dzaa004,
                  dzaa006 LIKE dzaa_t.dzaa006,
                  dzad006 LIKE dzad_t.dzad006,
                  dzad007 LIKE dzad_t.dzad007, #2014/04/30 by Hiko
                  dzad099 LIKE dzad_t.dzad099,
                  dzaa007 LIKE dzaa_t.dzaa007,
                  dzaa009 LIKE dzaa_t.dzaa009, #2014/12/17 by Hiko
                  dzaastus LIKE dzaa_t.dzaastus #2014/08/05 by Hiko
                  END RECORD,
          li_idx SMALLINT,
          lnode_act xml.DomNode,
          lnode_child xml.DomNode
   DEFINE ls_upd STRING #2014/08/05 by Hiko
   DEFINE ls_ch STRING #2015/11/09 by Hiko

   TRY
      LET ls_trigger = "sadzp030_tsd_gen_act : get ",gs_prog," action spec and setting"
      LET ls_sql = "SELECT dzaa003,dzaa004,dzaa006,dzad006,dzad007,dzad099,dzaa007,dzaa009,dzaastus", 
                    " FROM dzad_t",
                   " RIGHT JOIN dzaa_t ON dzaa001=dzad001 AND dzaa003=dzad002 AND dzaa004=dzad003 AND dzaa006=dzad005",
                   " WHERE dzaa001='",gs_prog,"' AND dzaa002=",g_revision," AND dzaa005='2' AND dzaa009='",gs_identity,"' AND dzaastus='Y' ORDER BY dzaa003" #2015/11/09 by Hiko:增加dzaastus='Y'的條件
      PREPARE dzad_prep FROM ls_sql
      LOCATE la_dzad[1].dzad099 IN FILE #在真正取得資料前,先設定陣列第一筆clob資料的LOCATE.
      DECLARE dzad_curs CURSOR FOR dzad_prep
      LET li_idx = 1
      FOREACH dzad_curs INTO la_dzad[li_idx].*
         LET lnode_act = p_node_tsd.appendChildElement("act")
         IF NOT gb_industry THEN
            CALL lnode_act.setAttribute("cite_std", "N") #2014/12/19 by Hiko
         ELSE
            CALL lnode_act.setAttribute("cite_std", la_dzad[li_idx].dzaa007 CLIPPED)
         END IF
         CALL lnode_act.setAttribute("id", la_dzad[li_idx].dzaa003 CLIPPED)
         CALL lnode_act.setAttribute("ver", la_dzad[li_idx].dzaa004 CLIPPED)
         CALL lnode_act.setAttribute("src", la_dzad[li_idx].dzaa006 CLIPPED)
         CALL lnode_act.setAttribute("type", la_dzad[li_idx].dzad006 CLIPPED)  
         CALL lnode_act.setAttribute("gen_code", la_dzad[li_idx].dzad007 CLIPPED)

         LET lnode_child = p_doc_tsd.createCDATASection(la_dzad[li_idx].dzad099)
         CALL lnode_act.appendChild(lnode_child)
		 
         #Begin:2015/11/09 by Hiko
         #LET ls_upd = ""
         ##Begin:2014/12/17 by Hiko:dzaa006(使用標示)=dzaa009(客製標示)
         #IF la_dzad[li_idx].dzaa006=la_dzad[li_idx].dzaa009 AND
         #   la_dzad[li_idx].dzaa004=g_revision THEN
         #   IF la_dzad[li_idx].dzaastus CLIPPED="Y" THEN
         #      IF g_revision>1 THEN 
         #         LET ls_upd = "u"
         #      END IF
         #   ELSE
         #      IF la_dzad[li_idx].dzaastus CLIPPED="N" THEN
         #         LET ls_upd = "d"
         #      END IF
         #   END IF
         #END IF
         ##End:2014/12/17 by Hiko

         #CALL lnode_act.setAttribute("status", ls_upd)

         LET ls_ch = ""
         IF la_dzad[li_idx].dzaa006=la_dzad[li_idx].dzaa009 AND
            la_dzad[li_idx].dzaa004=g_revision THEN
            IF g_revision>1 THEN 
               LET ls_ch = "Y"
            END IF
         END IF

         CALL lnode_act.setAttribute("ch", ls_ch)
         CALL lnode_act.setAttribute("status", "")
         #End:2015/11/09 by Hiko
      
         FREE la_dzad[li_idx].dzad099 #釋放LOCATE.
         LET li_idx = li_idx + 1
         LOCATE la_dzad[li_idx].dzad099 IN FILE #設定陣列第二筆以後clob資料的LOCATE.
      END FOREACH

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立<field>多語言節點資料.
# Input parameter : p_str_path .str路徑
#                   p_node_str <str> node
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2013/01/11 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_gen_strings_field(p_str_path, p_node_str)
   DEFINE p_str_path STRING, #.str路徑
          p_node_str xml.DomNode #<strings>
   DEFINE ch_str base.Channel,
          ls_trigger STRING,
          ls_line STRING,
          li_equal_idx SMALLINT,
          ls_name,ls_text STRING
   #Begin:2014/04/18 by Hiko
   DEFINE la_sfield DYNAMIC ARRAY OF RECORD
                    name STRING,
                    text STRING
                    END RECORD,
          li_idx SMALLINT
   #End:2014/04/18 by Hiko

   TRY
      LET ch_str = base.Channel.create()
      LET ls_trigger = "sadzp030_tsd_gen_strings_field : get str file : ",p_str_path
      CALL ch_str.openFile(p_str_path, "r")
      WHILE TRUE
         LET ls_line = ch_str.readLine() #ex:格式為"lbl_dzza001"="簽核等級"   
         IF ch_str.isEof() THEN
            EXIT WHILE
         END IF
      
         LET li_equal_idx = ls_line.getIndexOf("=", 1)
         IF li_equal_idx>0 THEN
            LET ls_name = ls_line.subString(1, li_equal_idx-1) #ex:"lbl_dzza001"(可能雙引號兩邊有空白)
            LET ls_name = ls_name.trim() #ex:"lbl_dzza001"(去除雙引號兩邊的空白)
            LET ls_name = ls_name.subString(2, ls_name.getLength()-1) #ex:lbl_dzza001(可能名稱兩邊有空白)
            LET ls_name = ls_name.trim() #這樣最保險.
            LET ls_text = ls_line.subString(li_equal_idx+1, ls_line.getLength())
            LET ls_text = ls_text.trim() #ex:"簽核等級"(去除雙引號兩邊的空白)
            LET ls_text = ls_text.subString(2, ls_text.getLength()-1) #ex:簽核等級(可能名稱兩邊有空白)
            LET ls_text = ls_text.trim() #這樣最保險.
            #Begin:2014/04/18 by Hiko
            LET li_idx = li_idx + 1
            LET la_sfield[li_idx].name = ls_name
            LET la_sfield[li_idx].text = ls_text
            #End:2014/04/18 by Hiko
         END IF
      END WHILE
      CALL ch_str.close()

      CALL sadzp030_tsd_gen_strings(p_node_str, la_sfield, "sfield") #2014/04/18 by Hiko

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, NULL) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立多語言
# Input parameter : p_node_str <strings> node
#                 : p_arr_str 多語言陣列
#                 : p_str_kind 多語言種類(sfield/sact)
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2014/04/18 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_gen_strings(p_node_str, p_arr_str, p_str_kind)
   DEFINE p_node_str xml.DomNode, #<strings>
          p_arr_str DYNAMIC ARRAY OF RECORD
                    name STRING,
                    text STRING
                    END RECORD,
          p_str_kind STRING
   DEFINE li_idx SMALLINT,
          lsb_name base.StringBuffer,
          ls_name STRING,
          ls_name_tmp STRING,
          lnode_child xml.DomNode

   LET lsb_name = base.StringBuffer.create()
   FOR li_idx=1 TO p_arr_str.getLength() #以前面為主
       LET ls_name = p_arr_str[li_idx].name CLIPPED
       LET ls_name_tmp = "#",ls_name,"#"
       IF lsb_name.getIndexOf(ls_name_tmp, 1)=0 THEN
          CALL lsb_name.append(ls_name_tmp)
          
          LET lnode_child = p_node_str.appendChildElement(p_str_kind)
          CALL lnode_child.setAttribute("name", ls_name)
          CALL lnode_child.setAttribute("text", p_arr_str[li_idx].text CLIPPED) 
          CALL lnode_child.setAttribute("lstr", "") 
       END IF
   END FOR
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立多語言:act
# Input parameter : p_node_str <strings> node
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2013/06/10 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_gen_strings_act(p_node_str)
   DEFINE p_node_str xml.DomNode #<strings>
   DEFINE ls_trigger STRING,
          ls_sql STRING,
          li_idx SMALLINT,
          la_dzad DYNAMIC ARRAY OF RECORD
                  dzad002 LIKE dzad_t.dzad002
                  END RECORD,
          ls_name STRING,
          ls_text STRING,
          ls_comment STRING,
          ls_standard STRING #這是平台要用的,和設計器沒關係.
   #Begin:2014/04/18 by Hiko
   DEFINE la_sact DYNAMIC ARRAY OF RECORD
                  name STRING,
                  text STRING
                  END RECORD
   #End:2014/04/18 by Hiko

   TRY
      LET ls_trigger = "sadzp030_tsd_gen_strings : gen act strings"
      #dzad_t的ls_sql的理由同dzac_t.
      LET ls_sql = "SELECT dzad_t.dzad002",
                    " FROM dzad_t",
                   " RIGHT JOIN dzaa_t ON dzaa001=dzad001 AND dzaa003=dzad002 AND dzaa004=dzad003 AND dzaa006=dzad005", #2014/04/14 by Hiko
                   " WHERE dzaa001='",gs_prog,"' AND dzaa002=",g_revision," AND dzaa005='2' AND dzaastus='Y' AND dzaa009='",gs_identity,"' AND dzaa003 NOT IN ",
                          "(SELECT dzfu002 FROM dzfu_t ",
                           " WHERE dzfu001='",DOWNSHIFT(gs_prog_type),"' AND dzfustus='Y')" #dzfustus='Y'表示可以開規格. #2014/05/27 by Hiko
      PREPARE dzad_prep5 FROM ls_sql
      DECLARE dzad_curs5 CURSOR FOR dzad_prep5
      LET li_idx = 1
      FOREACH dzad_curs5 INTO la_dzad[li_idx].*
         LET ls_name = la_dzad[li_idx].dzad002 CLIPPED 
         CALL s_azzi903_get_gzzq(gs_prog, ls_name) RETURNING ls_text,ls_comment,ls_standard
         #Begin:2014/04/18 by Hiko
         LET la_sact[li_idx].name = ls_name
         LET la_sact[li_idx].text = ls_text
         #End:2014/04/18 by Hiko
         LET li_idx = li_idx + 1
      END FOREACH

      CALL sadzp030_tsd_gen_strings(p_node_str, la_sact, "sact") #2014/04/03 by Hiko

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, NULL) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立<sa_spec>規格描述.
# Input parameter : p_doc_tsd <tsd> doc
#                   p_node_tsd <tsd> node
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2013/09/02 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_gen_sa_spec(p_doc_tsd, p_node_tsd)
   DEFINE p_doc_tsd xml.DomDocument,
          p_node_tsd xml.DomNode 
   DEFINE ls_trigger STRING,
          ls_sql STRING,
          la_dzem DYNAMIC ARRAY OF RECORD LIKE dzem_t.*,
          li_idx SMALLINT,
          lnode_sa_spec xml.DomNode,
          lnode_sa xml.DomNode,
          lnode_child xml.DomNode

   TRY
      LET lnode_sa_spec = p_node_tsd.appendChildElement("sa_spec")

      LET ls_trigger = "sadzp030_tsd_gen_sa_spec : get ",gs_prog," sa spec"
      LET ls_sql = "SELECT dzem_t.*",
                    " FROM dzem_t",
                   " WHERE dzem001='",gs_prog,"'",
                   " ORDER BY dzem003,dzem002"
      PREPARE dzem_prep FROM ls_sql
      LOCATE la_dzem[1].dzem099 IN FILE #在真正取得資料前,先設定陣列第一筆clob資料的LOCATE.
      DECLARE dzem_curs CURSOR FOR dzem_prep
      LET li_idx = 1
      FOREACH dzem_curs INTO la_dzem[li_idx].*
         CASE la_dzem[li_idx].dzem003 CLIPPED
            WHEN "1" #欄位規格
               LET lnode_sa = lnode_sa_spec.appendChildElement("sa_field")
               CALL lnode_sa.setAttribute("name", la_dzem[li_idx].dzem002 CLIPPED)
            WHEN "2" #Action規格
               LET lnode_sa = lnode_sa_spec.appendChildElement("sa_act")
               CALL lnode_sa.setAttribute("name", la_dzem[li_idx].dzem002 CLIPPED)
            WHEN "3" #整體規格:固定為"ALL"
               LET lnode_sa = lnode_sa_spec.appendChildElement("sa_all")
               CALL lnode_sa.setAttribute("name", "ALL")
         END CASE

         LET lnode_child = p_doc_tsd.createCDATASection(la_dzem[li_idx].dzem099)
         CALL lnode_sa.appendChild(lnode_child)
         FREE la_dzem[li_idx].dzem099 #釋放LOCATE.
         LET li_idx = li_idx + 1
         LOCATE la_dzem[li_idx].dzem099 IN FILE #設定陣列第二筆以後clob資料的LOCATE.
      END FOREACH

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立<exclude>排除設定
# Input parameter : p_doc_tsd <tsd> doc
#                   p_node_tsd <tsd> node
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2013/09/26 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_gen_exclude(p_doc_tsd, p_node_tsd)
   DEFINE p_doc_tsd xml.DomDocument,
          p_node_tsd xml.DomNode 
   DEFINE ls_trigger STRING,
          ls_sql STRING,
          #la_dzam DYNAMIC ARRAY OF RECORD LIKE dzam_t.*,
          la_dzam DYNAMIC ARRAY OF RECORD #2014/12/17 by Hiko
                                   dzam001 LIKE dzam_t.dzam001,
                                   dzam002 LIKE dzam_t.dzam002,
                                   dzam003 LIKE dzam_t.dzam003,
                                   dzam004 LIKE dzam_t.dzam004,
                                   dzam005 LIKE dzam_t.dzam005,
                                   dzam006 LIKE dzam_t.dzam006,
                                   dzamstus LIKE dzam_t.dzamstus,
                                   dzaa006 LIKE dzaa_t.dzaa006,
                                   dzaa009 LIKE dzaa_t.dzaa009
                                   END RECORD,
          li_idx SMALLINT,
          lnode_exclude xml.DomNode,
          lnode_child xml.DomNode
   DEFINE ls_main_sql STRING, #2014/08/05 by Hiko
          ls_upd STRING       #2014/08/05 by Hiko
   DEFINE ls_ch STRING #2015/11/09 by Hiko

   TRY
      LET lnode_exclude = p_node_tsd.appendChildElement("exclude")
      #Begin:2013/12/13 by Hiko:保險作法
      #討論後,行業別規格的排除設定自己也會有一份會比較合理,所以就不會有引用.
      CALL lnode_exclude.setAttribute("ver", "")
      CALL lnode_exclude.setAttribute("cite_std", "N") #2014/01/09 by Hiko:目前規劃都不引用
      CALL lnode_exclude.setAttribute("src", "")
      #End:2013/12/13 by Hiko:保險作法

      LET ls_trigger = "sadzp030_tsd_gen_exclude : get ",gs_prog," exclude"
      #LET ls_main_sql = "SELECT dzam_t.*", #2014/01/09 by Hiko
      LET ls_main_sql = "SELECT dzam001,dzam002,dzam003,dzam004,dzam005,", #2014/12/17 by Hiko
                               "dzam006,dzamstus,dzaa006,dzaa009",         #2014/12/17 by Hiko
                         " FROM dzaa_t"
      LET ls_sql = ls_main_sql,						 
                   " INNER JOIN dzam_t ON dzam001=dzaa001 AND dzam004=dzaa004 AND dzam005=dzaa006 AND dzamstus='Y'", #2014/05/29 by Hiko:dzam005改成PK
                   " WHERE dzaa001='",gs_prog,"' AND dzaa002=",g_revision," AND dzaa003='EXCLUDE' AND dzaa005='a' AND dzaastus='Y' AND dzaa009='",gs_identity,"'",
                     #Begin:2014/04/02 by Hiko:剔除垃圾資料
                     " AND (dzam003 IN (SELECT dzfj005 FROM dzfj_t",
                                                    " WHERE dzfj001='",gs_prog,"' AND dzfj002=",g_revision," AND dzfj017='",gs_identity,"') OR ",
                           "dzam003 IN (SELECT dzfi006 FROM dzfi_t",
                                                    " WHERE dzfi001='",gs_prog,"' AND dzfi002=",g_revision," AND dzfi009='",gs_identity,"'))",
                     #End:2014/04/02 by Hiko
                   #Begin:2015/11/09 by Hiko
                   #" UNION ",
                   #ls_main_sql,
                   #" INNER JOIN dzam_t ON dzam001=dzaa001 AND dzam004=dzaa004 AND dzam005=dzaa006 AND dzamstus='N'", #2014/05/29 by Hiko:dzam005改成PK
                   #" WHERE dzaa001='",gs_prog,"' AND dzaa002=",g_revision," AND dzaa003='EXCLUDE' AND dzaa005='a' AND dzaastus='Y' AND dzaa009='",gs_identity,"'"
                   " ORDER BY dzam003"
                   #End:2015/11/09 by Hiko
      PREPARE dzam_prep FROM ls_sql
      DECLARE dzam_curs CURSOR FOR dzam_prep
      LET li_idx = 1
      FOREACH dzam_curs INTO la_dzam[li_idx].*
         #版次只需要設定一次.
         IF li_idx=1 THEN
            CALL lnode_exclude.setAttribute("ver", la_dzam[li_idx].dzam004 CLIPPED)
            CALL lnode_exclude.setAttribute("src", la_dzam[li_idx].dzam005 CLIPPED)
         END IF
         
         LET lnode_child = lnode_exclude.appendChildElement("widget")
         CALL lnode_child.setAttribute("name", la_dzam[li_idx].dzam003 CLIPPED)

         #Begin:2015/11/09 by Hiko
         #LET ls_upd = ""
         ##Begin:2014/12/17 by Hiko:dzaa006(使用標示)=dzaa009(客製標示)
         #IF la_dzam[li_idx].dzamstus CLIPPED="N" THEN #防呆
         #   LET ls_upd = "d"
         #END IF
         #IF la_dzam[li_idx].dzaa006=la_dzam[li_idx].dzaa009 AND
         #   la_dzam[li_idx].dzam004=g_revision THEN
         #   IF la_dzam[li_idx].dzamstus CLIPPED="Y" THEN
         #      IF g_revision>1 THEN 
         #         LET ls_upd = "u"
         #      END IF
         #   ELSE
         #      IF la_dzam[li_idx].dzamstus CLIPPED="N" THEN
         #         LET ls_upd = "d"
         #      END IF
         #   END IF
         #END IF
         ##End:2014/12/17 by Hiko
	 #        
         #CALL lnode_child.setAttribute("status", ls_upd)

         LET ls_ch = ""
         IF la_dzam[li_idx].dzaa006=la_dzam[li_idx].dzaa009 AND
            la_dzam[li_idx].dzam004=g_revision THEN
            IF g_revision>1 THEN 
               LET ls_ch = "Y"
            END IF
         END IF

         CALL lnode_child.setAttribute("ch", ls_ch)
         CALL lnode_child.setAttribute("status", "")
         #End:2015/11/09 by Hiko

         LET li_idx = li_idx + 1
      END FOREACH

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立<cannot_del>不可刪除控件清單
# Input parameter : p_doc_tsd <tsd> doc
#                 : p_node_tsd <tsd> node
#                 : p_spec_rev 標準規格版次
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2014/10/22 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_gen_cannot_del(p_doc_tsd, p_node_tsd, p_spec_rev)
   DEFINE p_doc_tsd xml.DomDocument,
          p_node_tsd xml.DomNode,
          p_spec_rev LIKE dzaf_t.dzaf003 
   DEFINE ls_trigger STRING,
          ls_sql STRING,
          la_widget DYNAMIC ARRAY OF RECORD
                    widget LIKE dzfi_t.dzfi004
                    END RECORD,
          li_idx SMALLINT,
          lnode_cannot_del xml.DomNode,
          lnode_child xml.DomNode,
          lsb_widget base.StringBuffer

   TRY
      #Begin:2014/10/27 by Hiko
      IF cl_null(p_spec_rev) OR p_spec_rev=0 THEN
         RETURN TRUE
      END IF
      #End:2014/10/27 by Hiko

      DISPLAY "sadzp030_tsd_gen_cannot_del start..."

      LET lnode_cannot_del = p_node_tsd.appendChildElement("cannot_del")

      LET ls_trigger = "sadzp030_tsd_gen_cannot_del start..."
      LET lsb_widget = base.StringBuffer.create()

      #UNION會自行剔除重複資料,並且會自行排序.
      LET ls_sql = "SELECT DISTINCT(dzfi006) FROM dzfi_t", 
                   " WHERE dzfi001='",gs_prog,"' AND dzfi002='",p_spec_rev,"' AND dzfi009='s'",
                   " UNION ",
                   "SELECT DISTINCT(dzfj005) FROM dzfj_t", 
                   " WHERE dzfj001='",gs_prog,"' AND dzfj002='",p_spec_rev,"' AND dzfj017='s'"

      PREPARE widget_prep FROM ls_sql
      DECLARE widget_curs CURSOR FOR widget_prep

      LET li_idx = 1
      FOREACH widget_curs INTO la_widget[li_idx].*
         IF lsb_widget.getLength()>0 THEN
            CALL lsb_widget.append(",")
         END IF

         CALL lsb_widget.append(la_widget[li_idx].widget CLIPPED)

         LET li_idx = li_idx + 1
      END FOREACH
      
      LET lnode_child = p_doc_tsd.createCDATASection(lsb_widget.toString())
      CALL lnode_cannot_del.appendChild(lnode_child)

      DISPLAY "sadzp030_tsd_gen_cannot_del finish!"

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 檢查檔案/目錄是否存在.
# Input parameter : ps_file 檔案名稱
# Return code     : BOOLEAN(檔案存在:TRUE,檔案不存在:FALSE)
# Date & Author   : 2013/01/11 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp030_tsd_check_file_exist(ps_file) 
   DEFINE ps_file STRING
   DEFINE s_msg STRING

   IF NOT os.Path.exists(ps_file) THEN
      LET s_msg = "The file : ",ps_file," does not exist."
      DISPLAY s_msg
      RETURN FALSE
   END IF

   RETURN TRUE
END FUNCTION 

##########################################################################
# Access Modifier : PUBLIC 
# Descriptions    : 產生報表元件規格設計檔(程式名稱.rsd)
# Input parameter : p_prog 規格代號
#                 : p_module 模組
#                 : p_revision 規格版次
#                 : p_type 建構類型
#                 : p_identity 識別標示
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2014/04/07 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp030_tsd_gen_rsd(p_prog, p_module, p_revision, p_type, p_identity) #2014/05/23 by Hiko
   DEFINE p_prog STRING,
          p_module STRING,
          #p_revision STRING,
          p_revision STRING, #2014/12/10 by Hiko
          p_type STRING,
          p_identity STRING
   DEFINE ls_sql STRING,
          ls_trigger STRING
   DEFINE ls_module_dir STRING, #模組目錄
          ls_rsd STRING
   DEFINE ldoc_rsd xml.DomDocument,
          lnode_rsd xml.DomNode
   DEFINE lb_alm_result BOOLEAN,
          ls_err_msg STRING,
          ls_booking STRING

   LET gs_prog = p_prog.trim()
   LET gs_module = p_module.toUpperCase()
   #LET g_revision = p_revision.trim()
   LET g_revision = p_revision
   LET gs_identity = p_identity.trim()

   CALL sadzp030_tsd_init_var()

   TRY
      #Begin:2014/10/31 by Hiko
      IF cl_null(p_revision) THEN
         LET p_revision = 0
      END IF
      #End:2014/10/31 by Hiko
      LET ls_module_dir = FGL_GETENV(gs_module) CLIPPED
      #產生rsd檔根節點.
      LET ls_trigger = "sadzp030_tsd_gen_rsd : gen rsd"
      #LET ls_rsd = os.path.join(os.path.join(os.path.join(ls_module_dir, "dzx"), "tsd"), gs_prog||".rsd")
      LET ls_rsd = os.path.join(FGL_GETENV("TEMPDIR"), gs_prog||".rsd") #20160125 by Hiko
      LET ldoc_rsd = xml.domDocument.createDocument("spec")
      LET lnode_rsd = ldoc_rsd.getDocumentElement()
      CALL lnode_rsd.setAttribute("prog", gs_prog)
      CALL lnode_rsd.setAttribute("std_prog", gs_prog) #這是為了設計器固定格式而增加,並非行業別.非M/S/F就不會有行業別開發行為.
      CALL lnode_rsd.setAttribute("erpver", gs_erpver) #ERP大版
      CALL lnode_rsd.setAttribute("ver", g_revision)      
      CALL lnode_rsd.setAttribute("module", gs_module)      
      CALL lnode_rsd.setAttribute("type", p_type)  
      CALL lnode_rsd.setAttribute("booking", "Y") #2014/05/15 by Hiko:以流程來看,會呼叫sadzp030_tsd的時候就是有權限,所以不需要再檢查.

      CALL lnode_rsd.setAttribute("class", "") #這只是為了格式一致 
      CALL lnode_rsd.setAttribute("env", gs_env)      
      CALL lnode_rsd.setAttribute("zone", gs_zone)      
      CALL lnode_rsd.setAttribute("identity", gs_identity)      
      CALL lnode_rsd.setAttribute("topind", gs_topind)#170725-00043

      #報表元件只有一個且固定為<all>,並且沒有引用.
      LET ls_trigger = "sadzp030_tsd_gen_rsd : call sadzp030_tsd_gen_all"
      IF NOT sadzp030_tsd_gen_all(ldoc_rsd, lnode_rsd) THEN
         RETURN FALSE
      END IF
      DISPLAY "call sadzp030_tsd_gen_all finish!"

      LET ls_trigger = "sadzp030_tsd_gen_rsd : save rsd"
      #Begin:2014/04/07 by Hiko
      IF NOT sadzp030_tsd_save_tsd(ldoc_rsd, ls_rsd) THEN
         RETURN FALSE
      END IF
      #End:2014/04/07 by Hiko
      #Begin:2014/05/15 by Hiko:建立tsd.read
      LET ls_trigger = "sadzp030_tsd_gen_tsd : save tsd.read"
      LET ls_rsd = ls_rsd,".read"
      CALL lnode_rsd.setAttribute("booking", "N")
      IF NOT sadzp030_tsd_save_tsd(ldoc_rsd, ls_rsd) THEN
         RETURN FALSE
      END IF
      #End:2014/05/15 by Hiko
      DISPLAY "gen rsd success : ",ls_rsd

      RETURN TRUE
   CATCH 
      CALL sadzp030_tsd_err_catch(ls_trigger, ls_sql) 
      RETURN FALSE
   END TRY
END FUNCTION
