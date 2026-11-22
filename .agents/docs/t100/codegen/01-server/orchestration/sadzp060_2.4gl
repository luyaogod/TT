#+ 程式版本......: T6 Version 1.00.00 Build-0001 at 14/02/27
#
#+ 程式代碼......: sadzp060_2
#+ 設計人員......: Hiko
# Prog. Version..: 'T6-12.01.21(00000)'     #
#
# Program name   : sadzp060_2.4gl
# Description    : 關於規格/程式設計資料的相關功能
# Modify         : 20160121 by Hiko : 將標準轉客製r.l ALL的行為改在adzp050簽出的時候處理, 這樣可避免標準轉客製不完全(c111)然後持續做r.l ALL(會很久).
#                : 20160302 160302-00026 by Hiko : 1.檢查簽出版次是否和dzaf_t相同.
#                                                  2.修正錯誤訊息adz-00351的取替代參數.
#                : 20160516 160516-00005 by Hiko : 配合過單/Patch流程的判斷, 將sadzp060_2_get_curr_ver_info的顯現訊息從ERROR改成WARNING.
#                : 20160706 160706-00001 by madey: 調整sadzp060_2_chk_spec_type()的F類規則
#                : 20160923 160919-00031 by Ernest: 上傳時檢核是否符合 P65T 的規則, topstd 不檢核
#                : 20160926 160922-00039 by Hiko: 修改sadzp060_2_chk_permission:1.CODE上傳要多檢查與'FreeStyle'的情境.
#                                                                               2.ls_revision的判斷都改成li_revision
#                : 20170523 170523-00012 by madey: TRY CATCH時多拋出SQLERRMESSAGE
#                : 20170919 170919-00003 by 02632 : 解決產生dzba_t/dzbc_t的SQL出現額外的'm'字元,導致產生樹頭錯誤的問題.
#                : 20171120 171120-00029 by 02632 : 提升標準轉客製的效能:
#                                                   1.不要重新r.f,直接複製42f.
#                                                   2.G類在編譯的同時就產生rdd. #r.c $程式編號 rdd
#                                                     備註:4gl只能維持現況做重產,因為會特別處理OPEN WINDOW的畫面路徑.
#                                                   3.移除2013,2014,2015一堆註解,讓程式更清爽.  
#                                                   4.減少r.l XXX.42x的動作.
#                : 20180104 180104-00017 by 02785 : 修改sadzp060_2_chk_permission:增加上傳時檢查ERPVER(T100切版)
#                : 20180418 180418-00066 by 02785 : 配合Patch與過單工具包部署,原呼叫s_azzi900_upd_gzza011()改為sadzi888_01_upd_gzza011()
#                : 20181004 181004-00016 by 02785 : 調整邏輯:標準轉客製時,針對g/x類版次長樹的處理段改為給1c,避免退版時誤刪資料
#                : 20190111 190111-00026 by 02785 : 調整ERPVER卡控原則,統一由sadzp007_util_check_erpver()來判斷
#                : 20200207 200207-00010 by 02632 : 減少行業環境規格與程式開發過程當中, 頻繁取得型管資料, 避免網路頻寬的因素影響開發效能.
#                : 20200224 200215-00009 by 10553 : MSSQL Server相容性調整
#                : 20230106 230216-00010#1 by 02632 : 1.加入XR('K'類)的相關功能.IMPORTMENT:adzp050已經沒有取消簽出功能,改用adzp063退版,所以sadzp060_2_after_recall()都不用改.
#                :                                    2.順手調整dzax_t的新增規則:包含標準轉客製的情況,無任何資料就增加一筆.所以並不需要刪除的動作.關建字:DZAX_T.
#                : 20230720                           3.呼叫sadzp188_copy_rep_table前要防呆.
#                : 20230626 230627-00008#1 by 02632 : 解決B類dzax_t沒有資料,導致標準轉客製4gl產生路徑發生問題.關鍵字:20230626.
#                : 20231030 230925-00019#3 by 02160 : 新增支援 SPECRC3
#                : 20241030 241030-00027#1 by 02785 : 標準轉客製時,要將bdx檔一起複製到客製目錄下.
#                : 20241118 230925-00019#3 by 02632 : (補)SPECRC3防呆,避免還未佈署造成上傳誤會.

IMPORT os
IMPORT xml
IMPORT util

SCHEMA ds

GLOBALS "../../cfg/top_global.inc"
&include "../4gl/sadzp000_type.inc"
&include "../4gl/sadzp200_type.inc"
#for sadzp060_2_chk_permission
&include "../4gl/adzp080_type.inc"
#for sadzp050_zip_copy_4rp_list
&include "../4gl/sadzp050_type.inc"

DEFINE g_date     DATETIME YEAR TO SECOND,
       gs_dept    STRING, 
       gs_erpver  STRING, #ERP大版版號
       g_customer LIKE dzaa_t.dzaa010, #170919-00003:'gs_'改成'g_',宣告STRING改成LIKE. 
       g_env      LIKE dzaa_t.dzaa009, #辨識目前所在的環境:s.產中環境,c.客製環境
       g_dzaf002  LIKE dzaf_t.dzaf002
DEFINE gs_topind  STRING #200207-00010 #行業平台環境變數 TOPIND

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 初始化變數.
# Input parameter : none
# Return code     : BOOLEAN(成功:TRUE,失敗:FALSE)
# Date & Author   : 2013/01/11 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_2_init_var()
   TRY
      LET g_date = cl_get_current()
      LET gs_dept = g_dept CLIPPED
      LET gs_erpver = FGL_GETENV("ERPVER") CLIPPED
      LET g_customer = FGL_GETENV("CUST") CLIPPED
      LET g_env = FGL_GETENV("DGENV") CLIPPED
      LET gs_topind = FGL_GETENV("TOPIND") CLIPPED         #200207-00010
      IF cl_null(gs_topind) THEN LET gs_topind="sd" END IF #200207-00010
      RETURN TRUE
   CATCH
      CALL sadzp060_2_err_catch("", NULL)
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC 
# Descriptions    : 上傳時檢查是否為合法來源
# Input parameter : p_prog 程式代號
#                 : p_path tsd/tap檔路徑
#                 : p_kind 檔案種類(SPEC/CODE)
# Return code     : T_UPLOAD_PARAM 參數物件
#                 : STRING 錯誤訊息
# Date & Author   : 2013/10/07 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_2_chk_permission(p_prog, p_path, p_kind)
   DEFINE p_prog STRING,
          p_path STRING,
          p_kind STRING
   DEFINE lb_result            BOOLEAN,
          ls_err_msg           STRING, 
          ls_trigger           STRING,
          ls_kind              STRING, 
          ls_file              STRING,
          ldoc_file            xml.domDocument,
          lnode_file           xml.DomNode,
          ls_sql               STRING, 
          ls_field             STRING,
          li_revision          SMALLINT,
          ls_revision          STRING,
          ls_identity          STRING,
          ls_file_revision     STRING,
          ls_role              STRING,
          ls_type              STRING,
          ls_module            STRING,
          ls_free_style        STRING,
          li_cnt               SMALLINT,  
          lst_other            XML.DomNodeList,
          lnode_other          XML.DomNode,
          lnode_child          XML.DomNode,
          ls_tag_name          STRING,
          ls_err_replace       STRING,
          lo_upload_param      T_UPLOAD_PARAM,
          ls_std_revision      STRING,
          ls_std_identity      STRING,
          ls_file_zone         STRING,
          ls_server_zone       STRING,
          ls_designer_ver      STRING,
          ls_file_designer_ver STRING, 
          lo_program_info      DYNAMIC ARRAY OF T_PROGRAM_INFO,  #程式的動態陣列物件
          lo_return            T_DZLM_T,
          lo_dzlm              T_DZLM_T,         
          ls_DGENV             STRING    
   DEFINE ls_file_erpver       STRING #180104-00017 設計器離線檔內的erpver屬性
   DEFINE ls_erpver            STRING #180104-00017 $ERPVER
   DEFINE ls_file_env          STRING #190111-00026 設計器離線檔內的env屬性
   DEFINE ls_booking           STRING #200207-00010 設計器離線檔內的booking屬性
   
   LET ls_DGENV = FGL_GETENV("DGENV")   

   #Begin:200207-00010:ls_DGENV和g_env相同,但不再改程式了.
   IF NOT sadzp060_2_init_var() THEN
      LET ls_err_msg = "ERROR : sadzp060_2_chk_permission.sadzp060_2_init_var"
      RETURN ls_err_msg
   END IF
   #End:200207-00010

   TRY
      DISPLAY "[INFO]sadzp060_2_chk_permission(",p_prog,",",p_path,",",p_kind,") begin ..."
    
      LET p_prog = p_prog.trim() #160922-00039
      LET p_path = p_path.trim() #160922-00039
      LET p_kind = p_kind.trim()
      LET ls_kind = p_kind
      IF ls_kind.equals("SPECGEN") THEN
         LET ls_kind = "SPEC"
      END IF
      
      #230925-00019#3 begin
      IF ls_kind.equals("SPECRC3") THEN
         LET ls_kind = "SPEC"
      END IF
      #230925-00019#3 end

      IF ls_kind.equals("GCODE") THEN
         LET ls_kind = "CODE"
      END IF

      CASE ls_kind
         WHEN "SPEC"
            LET p_path = os.path.join(p_path, p_prog||".tsd")
            LET ls_role = "SD"
         WHEN "CODE"
            LET p_path = os.path.join(p_path, p_prog||".tap")
            LET ls_role = "PR"
         WHEN "CSPEC" 
            LET p_path = os.path.join(p_path, p_prog||".csd")
            LET ls_role = "SD"
         WHEN "RSPEC" 
            LET p_path = os.path.join(p_path, p_prog||".rsd")
            LET ls_role = "SD"
         WHEN "4RP" 
            LET p_path = os.path.join(p_path, p_prog||".vfy")
            LET ls_role = "PR"
         #230216-00010#1 begin   
         WHEN "XRP" 
            LET p_path = os.path.join(p_path, p_prog||".vfy")
            LET ls_role = "PR"
         #230216-00010#1 end   
         #SPECRC3 -s  230925-00019#3(補)
         OTHERWISE
            LET ls_err_msg = "ERROR : ",cl_replace_err_msg(cl_getmsg("adz-01856", g_lang), ls_kind) #找不到上傳類別%1,請確認此開發區域是否有佈署,謝謝.
            GOTO _RTN_ERR
         #SPECRC3 -e  230925-00019#3(補)
      END CASE

      IF NOT os.Path.exists(p_path) THEN
         LET ls_err_msg = "ERROR : ",cl_replace_err_msg(cl_getmsg("adz-00175", g_lang), p_path)
         GOTO _RTN_ERR
      END IF

      LET ls_trigger = "sadzp060_2_chk_permission : load tmp file=",p_path
      LET ldoc_file = xml.domDocument.create()
      CALL ldoc_file.load(p_path)
      LET lnode_file = ldoc_file.getDocumentElement()

      LET ls_file_zone = lnode_file.getAttribute("zone") CLIPPED
      LET ls_server_zone = FGL_GETENV("ZONE") CLIPPED

      #判斷區域是否相同:沒有丟包的行為,只能透過匯出/匯入設計資料來達到這樣的效果.
      IF NOT ls_file_zone.equals(ls_server_zone) THEN
         LET ls_err_msg = "ERROR : ",cl_replace_err_msg(cl_getmsg("adz-00122", g_lang), ls_server_zone||"|"||ls_file_zone)
         GOTO _RTN_ERR
      END IF

      #Begin: 190111-00026 
      #判斷T100 ERP版本是否符合:以設計器離線檔的屬性erpver與該SERVER環境變數ERPVER比較
      LET ls_file_env = lnode_file.getAttribute("env") CLIPPED 
      IF cl_null(ls_file_env) THEN LET ls_file_env = "s" END IF #default
      LET ls_file_erpver = lnode_file.getAttribute("erpver") CLIPPED 
      IF cl_null(ls_file_erpver) THEN LET ls_file_erpver = "1.0" END IF #default
      LET ls_erpver = FGL_GETENV("ERPVER")
      IF cl_null(ls_erpver) THEN LET ls_erpver = "1.0" END IF #default
      CALL sadzp007_util_check_erpver(ls_file_env,ls_DGENV,ls_file_erpver,ls_erpver) RETURNING lb_result ,ls_err_msg
      IF NOT lb_result THEN 
         LET ls_err_msg =  "ERROR : ", cl_replace_err_msg(cl_getmsg("adz-01302", g_lang),ls_file_erpver||'|'||ls_erpver)
         GOTO _RTN_ERR
      END IF
      #End: 190111-00026 
      
      #原本可以直接使用tsd/tap內的booking屬性來判斷即可, 但為了避免出問題, 所以還是重查一次.
      LET ls_file_revision = lnode_file.getAttribute("ver") CLIPPED 
      LET ls_module = lnode_file.getAttribute("module") CLIPPED 
 
      LET ls_type = lnode_file.getAttribute("type") CLIPPED
      LET lo_upload_param.arg11_type = ls_type
      
      LET lo_upload_param.arg7_std_program = lnode_file.getAttribute("std_prog") CLIPPED
      LET ls_trigger = "sadzp060_2_chk_permission : sadzp060_2_get_spec(or code)_max_revision"
      IF ls_role.equals("SD") THEN
         CALL sadzp060_2_get_spec_curr_revision(p_prog, ls_type, ls_module) RETURNING li_revision,ls_identity,ls_err_msg
         IF NOT cl_null(ls_err_msg) THEN GOTO _RTN_ERR END IF
         LET lo_upload_param.arg5_spec_version = li_revision

         CALL sadzp060_2_get_spec_curr_revision(lo_upload_param.arg7_std_program, ls_type, ls_module) RETURNING ls_std_revision,ls_std_identity,ls_err_msg
         IF NOT cl_null(ls_err_msg) THEN GOTO _RTN_ERR END IF
         LET lo_upload_param.arg8_std_spec_version = ls_std_revision 
      ELSE #PR
         CALL sadzp060_2_get_code_curr_revision(p_prog, ls_type, ls_module) RETURNING li_revision,ls_identity,ls_err_msg
         IF NOT cl_null(ls_err_msg) THEN GOTO _RTN_ERR END IF
         LET lo_upload_param.arg6_code_version = li_revision 

         CALL sadzp060_2_get_code_curr_revision(lo_upload_param.arg7_std_program, ls_type, ls_module) RETURNING ls_std_revision,ls_std_identity,ls_err_msg
         IF NOT cl_null(ls_err_msg) THEN GOTO _RTN_ERR END IF
         LET lo_upload_param.arg9_std_code_version = ls_std_revision 
      END IF

      IF li_revision<>0 THEN
         #檢查要更新的版次是否相同
         IF ls_file_revision<>li_revision THEN
            LET ls_err_replace = p_prog,"|",ls_kind,"|",ls_file_revision,"|",li_revision
            LET ls_err_msg = "ERROR : ",cl_replace_err_msg(cl_getmsg("adz-00222", g_lang), ls_err_replace)
            GOTO _RTN_ERR  
         ELSE #ls_file_revision=li_revision
            LET ls_trigger = "sadzp060_2_chk_permission : sadzp060_2_have_checked_out"
            #Begin:200207-00010
            IF ls_DGENV="s" AND NOT gs_topind.equals("sd") THEN
               LET ls_booking = lnode_file.getAttribute("booking") CLIPPED
               IF NOT ls_booking.equals("Y") THEN
                  LET ls_err_msg = "ERROR : ",cl_getmsg("adz-00827", g_lang) #您還沒有簽出程式!
                  GOTO _RTN_ERR  
               END IF
               DISPLAY "[INFO]行業環境簽出修改"
            ELSE #End:200207-00010
               CALL sadzp060_2_have_checked_out(p_prog, ls_type, ls_role, 0) RETURNING ls_err_msg
               IF NOT cl_null(ls_err_msg) THEN GOTO _RTN_ERR END IF
               #Begin:160302-00026
               LET lo_program_info[1].pi_NAME = p_prog
               LET lo_program_info[1].pi_TYPE = ls_type
               CALL sadzp200_alm_get_dzlm(lo_program_info[1].*, ls_role) RETURNING lo_return.*
               IF ls_role.equals("SD") THEN
                  IF li_revision<>lo_return.dzlm006 THEN
                     LET ls_err_replace = "SPEC|",li_revision,"|",lo_return.dzlm006
                     LET ls_err_msg = "ERROR : ",cl_replace_err_msg(cl_getmsg("adz-00785", g_lang), ls_err_replace)
                     GOTO _RTN_ERR  
                  END IF
               ELSE #PR
                  IF li_revision<>lo_return.dzlm009 THEN
                     LET ls_err_replace = "CODE|",li_revision,"|",lo_return.dzlm009
                     LET ls_err_msg = "ERROR : ",cl_replace_err_msg(cl_getmsg("adz-00785", g_lang), ls_err_replace)
                     GOTO _RTN_ERR  
                  END IF
               END IF
               DISPLAY "[INFO]非行業環境簽出修改"
               #End:160302-00026
            END IF #200207-00010

            #Begin:160922-00039
            IF ls_role.equals("PR") AND ls_kind.equals("CODE") THEN
               LET lst_other = lnode_file.getElementsByTagName("other")
               LET lnode_other = lst_other.getItem(1) #目前只有一個other
               LET lnode_child = lnode_other.getFirstChildElement()
               WHILE lnode_child IS NOT NULL
                  LET ls_tag_name = lnode_child.getLocalName() CLIPPED
                  IF ls_tag_name.equals("free_style") THEN
                     LET ls_free_style = lnode_child.getAttribute("value") CLIPPED
                     IF cl_null(ls_free_style) THEN
                        #Begin:200215-00009
                        LET ls_sql = "select count(*) from dzax_t",
                                     " where dzax001='",p_prog,"'",
                                     "   and dzax003='Y'"
                        #End:200215-00009
                        PREPARE dzax_prep FROM ls_sql
                        EXECUTE dzax_prep INTO li_cnt
                        FREE dzax_prep
                        IF li_cnt > 0 THEN
                           LET ls_err_msg = "ERROR : ",cl_getmsg("adz-00904", g_lang) #我們偵測到Server端已經轉為FreeStyle, 但Client端還是標準, 麻煩請重新下載再繼續開發.
                           GOTO _RTN_ERR  
                        END IF
                     END IF
                  END IF

                  LET lnode_child = lnode_child.getNextSiblingElement()
               END WHILE
            END IF
            #End:160922-00039
         END IF #ls_file_revision<>li_revision
      ELSE
         LET ls_err_msg = "ERROR : the file revision is 0!"
         GOTO _RTN_ERR
      END IF

      #判斷設計器版本是否符合.
      LET ls_file_designer_ver = lnode_file.getAttribute("designer_ver") CLIPPED
      CALL sadzp060_2_get_designer_ver() RETURNING ls_designer_ver
      IF ls_file_designer_ver <> ls_designer_ver THEN
         LET ls_err_msg =  "ERROR : ",cl_getmsg("adz-00588", g_lang)
         GOTO _RTN_ERR
      END IF

      LET lo_upload_param.arg10_dgenv = lnode_file.getAttribute("identity") CLIPPED 
      LET lo_upload_param.arg2_module_name = ls_module 

     ##Begin: 190111-00026 remark
     ##Begin: 180104-00017
     ##判斷T100 ERP版本是否符合
     #LET ls_file_erpver = lnode_file.getAttribute("erpver") CLIPPED 
     #LET ls_erpver = FGL_GETENV("ERPVER")
     #IF cl_null(ls_file_erpver) THEN LET ls_file_erpver = "1.0" END IF #default
     #IF cl_null(ls_erpver) THEN LET ls_erpver = "1.0" END IF #default
     #IF ls_erpver <> ls_file_erpver THEN 
     #   LET ls_err_msg =  "ERROR : ", cl_replace_err_msg(cl_getmsg("adz-01302", g_lang),ls_file_erpver||'|'||ls_erpver)
     #   GOTO _RTN_ERR
     #END IF
     ##End: 180104-00017
     ##End: 190111-00026 remark

      #160919-00031 begin
      IF g_user <> "topstd" AND ls_DGENV = "s" THEN 
        LET ls_trigger = "sadzp060_2_chk_permission : Check if P65T stage."
        LET lo_program_info[1].pi_NAME = p_prog
        LET lo_program_info[1].pi_TYPE = ls_type
        CALL sadzp200_alm_get_dzlm(lo_program_info[1].*, ls_role) RETURNING lo_dzlm.*
        CALL sadzp060_2_check_if_stage_P65T(lo_dzlm.*) RETURNING lb_result
        IF lb_result THEN
           LET ls_err_msg =  "ERROR : ",cl_getmsg("adz-00902", g_lang) #該程式或規格已經處於P65T階段, 無法執行上傳 !
           GOTO _RTN_ERR
        END IF
      END IF  
      #160919-00031 end

      DISPLAY "[INFO]sadzp060_2_chk_permission(",p_prog,",",p_path,",",p_kind,") finish !"

      RETURN lo_upload_param.*,ls_err_msg

      LABEL _RTN_ERR:
         DISPLAY ls_err_msg
         RETURN lo_upload_param.*,ls_err_msg
   CATCH 
      CALL sadzp060_2_err_catch(ls_trigger, ls_sql)
      LET ls_err_msg = "ERROR : sadzp060_2_chk_permission(",p_prog,",",p_path,",",p_kind,")"
      RETURN lo_upload_param.*,ls_err_msg
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PBULIC
# Descriptions    : 取得設計器版本資訊.
# Input parameter : none
# Return code     : STRING 設計器版本
# Date & Author   : 2014/05/15 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_2_get_designer_ver()
   DEFINE ls_ver_path      STRING,
          lch_designer_ver base.Channel,
          ls_line          STRING

   LET ls_ver_path = os.path.join(os.path.join(FGL_GETENV("COM"), "mta"), "ver")
   LET lch_designer_ver = base.Channel.create()
   CALL lch_designer_ver.openFile(ls_ver_path, "r")
   WHILE TRUE
      IF lch_designer_ver.isEof() OR NOT cl_null(ls_line)THEN
         EXIT WHILE
      END IF

      LET ls_line = lch_designer_ver.readLine() #目前只有一行
   END WHILE
   CALL lch_designer_ver.close()

   RETURN ls_line
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 擷取錯誤訊息.
# Input parameter : p_trigger 執行的顯現資訊
#                   p_sql 執行的SQL
# Return code     : void
# Date & Author   : 2014/05/15 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_2_err_catch(p_trigger, p_sql)
   DEFINE p_trigger STRING,
          p_sql     STRING

   DISPLAY "ERROR : ls_trigger=",p_trigger
   DISPLAY "STATUS=",STATUS
   DISPLAY "SQLCA.SQLCODE=",SQLCA.SQLCODE
   DISPLAY "SQLERRMESSAGE=",SQLERRMESSAGE #170523-00012
   DISPLAY "ls_sql=",p_sql,"<<"
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC
# Descriptions    : 判斷是否已經簽出.
# Input parameter : p_prog 程式代號
#                   p_type 程式類型(M/S/F/G/X...)
#                   p_role 簽出角色(SD/PR)
#                   p_popup_err 彈出錯誤訊息窗(1:彈窗,0:不彈窗)
# Return code     : STRING (成功:NULL,失敗:錯誤訊息)
# Date & Author   : 2014/06/12 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_2_have_checked_out(p_prog, p_type, p_role, p_popup_err)
   DEFINE p_prog      STRING,
          p_type      STRING,
          p_role      STRING,
          p_popup_err SMALLINT
   DEFINE lo_program_info DYNAMIC ARRAY OF T_PROGRAM_INFO,
          lb_result       BOOLEAN,                        
          ls_owner        STRING, 
          ls_err_msg      STRING,
          ls_err_code     STRING,
          ls_replace_arg  STRING 

   LET p_prog = p_prog.trim()
   LET p_type = p_type.trim()
   LET p_role = p_role.trim()

   #設定程式資訊(有多支程式時請用For迴圈給定)
   LET lo_program_info[1].pi_NAME = p_prog #程式代碼
   LET lo_program_info[1].pi_TYPE = p_type #類別

   #檢核該程式是否已經被簽出
   CASE p_role
      WHEN "ALL" #先檢查規格,再檢查程式
        CALL sadzp200_alm_get_alm_owner_by_role(lo_program_info[1].*, "SD") RETURNING lb_result,ls_owner
        IF NOT lb_result THEN
           LET ls_err_code = "adz-00313"
        ELSE
           IF ls_owner.equals(g_account) THEN
              CALL sadzp200_alm_get_alm_owner_by_role(lo_program_info[1].*, "PR") RETURNING lb_result,ls_owner
              IF NOT lb_result THEN
                 LET ls_err_code = "adz-00313"
              ELSE
                 IF NOT ls_owner.equals(g_account) THEN 
                    LET lb_result = FALSE
                    LET ls_err_code = "adz-00351"
                    LET ls_replace_arg = "CODE|",ls_owner
                 END IF
              END IF
           ELSE
              LET lb_result = FALSE
              LET ls_err_code = "adz-00351"
              LET ls_replace_arg = "SPEC|",ls_owner
           END IF
        END IF
      WHEN "SD"
         CALL sadzp200_alm_get_alm_owner_by_role(lo_program_info[1].*, "SD") RETURNING lb_result,ls_owner
         IF NOT lb_result THEN
            LET ls_err_code = "adz-00314"
         ELSE #有簽出還得判斷是誰簽出
            IF NOT ls_owner.equals(g_account) THEN 
               LET lb_result = FALSE
               LET ls_err_code = "adz-00351"
               LET ls_replace_arg = "SPEC|",ls_owner
            END IF
         END IF
      WHEN "PR"
         CALL sadzp200_alm_get_alm_owner_by_role(lo_program_info[1].*, "PR") RETURNING lb_result,ls_owner
         IF NOT lb_result THEN
            LET ls_err_code = "adz-00315"
         ELSE
            IF NOT ls_owner.equals(g_account) THEN 
               LET lb_result = FALSE
               LET ls_err_code = "adz-00351"
               LET ls_replace_arg = "CODE|",ls_owner
            END IF
         END IF
   END CASE

   LET ls_err_msg = NULL

   IF NOT lb_result THEN
      IF p_popup_err=1 THEN
         INITIALIZE g_errparam TO NULL
         LET g_errparam.code =  ls_err_code
         LET g_errparam.replace[1] = ls_replace_arg
         LET g_errparam.replace[2] = ls_owner
         LET g_errparam.extend = NULL
         LET g_errparam.popup = TRUE
         CALL cl_err()
      END IF

      LET ls_err_msg = cl_replace_err_msg(cl_getmsg(ls_err_code, g_lang), ls_replace_arg)
   END IF

   RETURN ls_err_msg
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC
# Descriptions    : 在下載畫面選擇簽出清單之後要做的事情.
# Input parameter : p_prog 程式代號
#                 : p_role 角色(SD/PR)
#                 : p_curr_dzaf 目前版次資訊
#                 : p_prev_dzaf 前版版次資訊
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2014/05/23 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_2_after_check_out_for_download(p_prog, p_role, p_curr_dzaf, p_prev_dzaf)
   DEFINE p_prog      STRING,
          p_role      STRING,
          p_curr_dzaf T_DZAF_T, #範例:版次2
          p_prev_dzaf T_DZAF_T  #範例:版次1
   DEFINE ls_type               STRING,
          l_curr_spec_revision  LIKE dzaf_t.dzaf003,
          l_curr_code_revision  LIKE dzaf_t.dzaf004,
          ls_curr_identity      STRING,
          ls_curr_module        STRING,
          ls_curr_module_path   STRING,
          l_prev_spec_revision  LIKE dzaf_t.dzaf003,
          l_prev_code_revision  LIKE dzaf_t.dzaf004,
          ls_prev_identity      STRING,
          ls_prev_module        STRING,
          ls_prev_module_path   STRING,
          lb_result             BOOLEAN,
          ls_err_msg            STRING,
          ls_err_msg2           STRING,
          ls_err_msg3           STRING, #失敗沒關係,但要印出訊息
          lo_program_info       T_PROGRAM_INFO,
          lo_template_list      DYNAMIC ARRAY OF T_TEMPLATE_LIST_R,
          ls_cmd                STRING,
          ls_from_4fd           STRING,
          ls_to_4fd_dir         STRING,
          ls_to_4fd             STRING,
          ls_from_4rp_dir       STRING,
          ls_to_4rp_dir         STRING,
          ls_to_4gl_dir         STRING,
          ls_trigger            STRING,
          ls_sql                STRING,
          lb_do_flag            BOOLEAN,
          ls_mnt4rp_module_path STRING,
          ls_mnt4rp_4rp_dir     STRING,
          l_time_s DATETIME YEAR TO SECOND,
          l_time_e DATETIME YEAR TO SECOND,
          l_time_r INTERVAL DAY  TO SECOND
   #230216-00010#1 -s
   DEFINE ls_from_xrp_dir         STRING,
          ls_to_xrp_dir           STRING,
          ls_mntxrp_module_path   STRING,
          ls_mntxrp_xrp_dir       STRING,
          ls_to_xrp_back_dir      STRING,
          ls_prog_id_first_char   STRING,
          ls_from_xdd_lang_dir    STRING, #20230330
          ls_from_xdd_dir         STRING,
          ls_to_xdd_lang_dir      STRING, #20230330
          ls_to_xdd_dir           STRING,
          ls_to_xdd_back_lang_dir STRING, #20230330
          ls_to_xdd_back_dir      STRING
   DEFINE lb_has_list      BOOLEAN,
          lo_xrp_back_list DYNAMIC ARRAY OF T_TEMPLATE_LIST_R
   DEFINE lb_result9 BOOLEAN,
          ls_errmsg9 STRING
   #230216-00010#1 -e

   #簽出流程說明:
   #情境一.s4-->s5 or c2-->c3
   #1-1.長樹
   #1-2.複製4rp到back
   #情境二.s5-->c1
   #2-0.更新客製(c-->s)
   #2-1.長樹
   #2-2.複製4rp到客製back以及客製4rp下
   #2-3.重產4fd
   #2-4.重產4gl

   #規格簽出並修改會影響程式;程式就只是影響自己.

   LET l_time_s = cl_get_current()
   DISPLAY "### sadzp060_2_after_check_out_for_download : start time=",l_time_s

   TRY 
      IF NOT sadzp060_2_init_var() THEN
         LET ls_err_msg = "ERROR : sadzp060_2_after_check_out_for_download.sadzp060_2_init_var"
         RETURN ls_err_msg
      END IF
      
      LET p_prog = p_prog.trim()
      LET ls_type = p_curr_dzaf.dzaf005 CLIPPED
      LET l_curr_spec_revision = p_curr_dzaf.dzaf003 CLIPPED
      LET l_curr_code_revision = p_curr_dzaf.dzaf004 CLIPPED
      LET ls_curr_identity = p_curr_dzaf.dzaf010 CLIPPED
      LET ls_curr_module = p_curr_dzaf.dzaf006 CLIPPED
      LET ls_curr_module_path = FGL_GETENV(UPSHIFT(ls_curr_module))
      LET l_prev_spec_revision = p_prev_dzaf.dzaf003 CLIPPED
      LET l_prev_code_revision = p_prev_dzaf.dzaf004 CLIPPED
      LET ls_prev_identity = p_prev_dzaf.dzaf010 CLIPPED
      LET ls_prev_module = p_prev_dzaf.dzaf006 CLIPPED
      LET ls_prev_module_path = FGL_GETENV(UPSHIFT(ls_prev_module))

      DISPLAY "[INFO]sadzp060_2_after_check_out_for_download parameter :"
      DISPLAY "      ","p_role=",p_role   
      DISPLAY "      ","p_prog=",p_prog   
      DISPLAY "      ","ls_type=",ls_type   
      DISPLAY "      ","l_curr_spec_revision=",l_curr_spec_revision   
      DISPLAY "      ","l_curr_code_revision=",l_curr_code_revision   
      DISPLAY "      ","ls_curr_identity=",ls_curr_identity   
      DISPLAY "      ","ls_curr_module=",ls_curr_module   
      DISPLAY "      ","ls_curr_module_path=",ls_curr_module_path   
      DISPLAY "      ","l_prev_spec_revision=",l_prev_spec_revision   
      DISPLAY "      ","l_prev_code_revision=",l_prev_code_revision   
      DISPLAY "      ","ls_prev_identity=",ls_prev_identity   
      DISPLAY "      ","ls_prev_module=",ls_prev_module   
      DISPLAY "      ","ls_prev_module_path=",ls_prev_module_path   
      
      DISPLAY "      ",p_role," GO !!!"

      CASE p_role
         WHEN "SD"
            IF cl_null(l_prev_spec_revision) THEN
               RETURN NULL
            END IF
         WHEN "PR"
            #20230626 #DZAX_T -s
           #IF cl_null(l_prev_code_revision) THEN
           #   CALL sadzp060_2_ins_dzax(p_prog, ls_type, ls_curr_identity, ls_prev_identity) RETURNING lb_result,ls_err_msg  
           #   IF NOT lb_result THEN GOTO _ROLLBACK_FLAG END IF
           #   RETURN NULL
           #END IF
            CALL sadzp060_2_ins_dzax(p_prog, ls_type, ls_curr_identity, ls_prev_identity) RETURNING lb_result,ls_err_msg  
            IF NOT lb_result THEN GOTO _ROLLBACK_FLAG END IF
            #20230626 #DZAX_T -e
      END CASE

      IF ls_prev_identity <> ls_curr_identity THEN
         #Begin:171120-00029:標準轉客製分類別通常會跑兩次,所以這邊要判斷類別只做一次即可.
         LET lb_do_flag = FALSE
         CASE
            WHEN ls_type="M" OR ls_type="S" OR ls_type="F" OR ls_type="Q"
               IF p_role="SD" THEN
                  LET lb_do_flag = TRUE
               END IF
            OTHERWISE
               IF p_role="PR" THEN
                  LET lb_do_flag = TRUE
               END IF
         END CASE

         IF lb_do_flag THEN
            DISPLAY "      ","sadzp060_2_ins_dzax()..."   
            CALL sadzp060_2_ins_dzax(p_prog, ls_type, ls_curr_identity, ls_prev_identity) RETURNING lb_result,ls_err_msg
            IF NOT lb_result THEN GOTO _ROLLBACK_FLAG END IF
            
            #簽出時(s --> c)要呼叫更新註冊資料的[是否客製]:Y(c)
            IF ls_type<>"Q" THEN
               #Begin: 180418-00066 modify
              #DISPLAY "      ","s_azzi900_upd_gzza011()..."   
              #CALL s_azzi900_upd_gzza011(p_prog, ls_type, ls_curr_identity) RETURNING lb_result
               DISPLAY "      ","sadzi888_01_upd_gzza011()..."   
               CALL sadzi888_01_upd_gzza011(p_prog, ls_type, ls_curr_identity) RETURNING lb_result
               #End: 180418-00066 modify
            ELSE
               DISPLAY "      ","sadzp210_copy_qry_from_std_to_cust()..."   
               CALL sadzp210_copy_qry_from_std_to_cust(p_prog) RETURNING lb_result,ls_err_msg
            END IF
            
            IF NOT lb_result THEN 
               LET ls_err_msg = "ERROR : ",p_prog," updated registration information fail when check out!"
               GOTO _ROLLBACK_FLAG #有錯誤訊息表示失敗
            END IF
         END IF
         #End:171120-00029
      END IF
      
      CASE p_role
         WHEN "SD"
            #建立規格樹
            DISPLAY "      ","sadzp060_2_create_spec_structure()..."   
            CALL sadzp060_2_create_spec_structure(p_prog, p_curr_dzaf.*, p_prev_dzaf.*) RETURNING ls_err_msg
            IF NOT cl_null(ls_err_msg) THEN GOTO _ROLLBACK_FLAG END IF  
            
            IF ls_type="M" OR ls_type="S" OR ls_type="F" OR ls_type="Q" THEN #只有M/S/F才有4fd,Q類hard code也有
               #只有在標準(s)變客製(c)時才要做4fd複製的動作
               IF ls_prev_identity <> ls_curr_identity THEN
                  DISPLAY "[INFO]sadzp060_2_after_check_out_for_download : spec s --> c"
                  LET ls_from_4fd = os.path.JOIN(os.path.JOIN(ls_prev_module_path, "4fd"), p_prog||".4fd")
                  LET ls_to_4fd_dir = os.path.JOIN(ls_curr_module_path, "4fd")
                  LET ls_to_4fd = os.path.JOIN(ls_to_4fd_dir, p_prog||".4fd")
                  DISPLAY "      ","cl_copy_file(",ls_from_4fd,", ",ls_to_4fd,")"
                  IF NOT cl_copy_file(ls_from_4fd, ls_to_4fd) THEN
                     LET ls_err_msg = "ERROR : Fail to copy file : ",ls_to_4fd
                     GOTO _ROLLBACK_FLAG  
                  END IF

                  #Begin:171120-00029:不要編譯,直接搬移42f.
                  #編譯4fd:編譯過程失敗都不影響簽出
                  # IF cl_change_dir(ls_to_4fd_dir) THEN
                  #    LET ls_cmd = "r.f ",p_prog," tiptop" 
                  #    CALL cl_cmdrun_openpipe("r.f", ls_cmd, FALSE) RETURNING lb_result,ls_err_msg3
                  #    IF NOT lb_result THEN
                  #       DISPLAY "WARNING : ",ls_err_msg3
                  #    END IF
                  # ELSE
                  #    DISPLAY "WARNING : change dir to ",ls_to_4fd_dir," fail!"
                  # END IF
                  LET ls_from_4fd = os.path.JOIN(os.path.JOIN(ls_prev_module_path, "42f"), p_prog||".42f")
                  LET ls_to_4fd_dir = os.path.JOIN(ls_curr_module_path, "42f")
                  LET ls_to_4fd = os.path.JOIN(ls_to_4fd_dir, p_prog||".42f")
                  DISPLAY "      ","cl_copy_file(",ls_from_4fd,", ",ls_to_4fd,")"
                  IF NOT cl_copy_file(ls_from_4fd, ls_to_4fd) THEN
                     DISPLAY "WARNING : Fail to copy file : ",ls_to_4fd
                  END IF
                  #End:171120-00029

                  LET ls_from_4fd = os.path.JOIN(os.path.JOIN(os.path.JOIN(ls_prev_module_path, "str"), g_lang), p_prog||".str")
                  LET ls_to_4fd_dir = os.path.JOIN(os.path.JOIN(ls_curr_module_path, "str"), g_lang)
                  LET ls_to_4fd = os.path.JOIN(ls_to_4fd_dir, p_prog||".str")
                  DISPLAY "      ","cl_copy_file(",ls_from_4fd,", ",ls_to_4fd,")"
                  IF NOT cl_copy_file(ls_from_4fd, ls_to_4fd) THEN
                     #Begin:171120-00029:多語言不影響開發,所以改成WARNING即可.
                     # LET ls_err_msg = "ERROR : Fail to copy file : ",ls_to_4fd
                     # GOTO _ROLLBACK_FLAG  
                     DISPLAY "WARNING : Fail to copy file : ",ls_to_4fd
                     #End:171120-00029
                  END IF
                  LET ls_from_4fd = os.path.JOIN(os.path.JOIN(os.path.JOIN(ls_prev_module_path, "str"), g_lang), p_prog||".42s")
                  LET ls_to_4fd_dir = os.path.JOIN(os.path.JOIN(ls_curr_module_path, "str"), g_lang)
                  LET ls_to_4fd = os.path.JOIN(ls_to_4fd_dir, p_prog||".42s")
                  DISPLAY "      ","cl_copy_file(",ls_from_4fd,", ",ls_to_4fd,")"
                  IF NOT cl_copy_file(ls_from_4fd, ls_to_4fd) THEN
                     #Begin:171120-00029:多語言不影響開發,所以改成WARNING即可.
                     # LET ls_err_msg = "ERROR : Fail to copy file : ",ls_to_4fd
                     # GOTO _ROLLBACK_FLAG  
                     DISPLAY "WARNING : Fail to copy file : ",ls_to_4fd
                     #End:171120-00029
                  END IF

                  #Begin: 241030-00027#1
                  #標準轉客製時,要將bdx檔一起複製到客製目錄下(若沒有檔案就忽略)
                  LET ls_from_4fd   = os.path.JOIN(os.path.JOIN(os.path.JOIN(ls_prev_module_path, "dzx"), "tsd"), p_prog||".bdx")
                  LET ls_to_4fd_dir = os.path.JOIN(os.path.JOIN(ls_curr_module_path, "dzx"), "tsd")
                  LET ls_to_4fd = os.path.JOIN(ls_to_4fd_dir, p_prog||".bdx")
                  IF os.path.EXISTS(ls_from_4fd ) THEN
                     DISPLAY "      ","cl_copy_file(",ls_from_4fd,", ",ls_to_4fd,")"
                     IF NOT cl_copy_file(ls_from_4fd, ls_to_4fd) THEN
                        DISPLAY "WARNING : Fail to copy file : ",ls_to_4fd
                     END IF
                  END IF
                  #End: 241030-00027#1
               END IF

               #建立UI樹
               DISPLAY "      ","sadzp168_3(",ls_curr_module,",",p_prog,",",l_curr_spec_revision,",",ls_curr_identity,")"
               CALL sadzp168_3(ls_curr_module, p_prog, l_curr_spec_revision, ls_curr_identity) RETURNING lb_result,ls_err_msg #解析4fd畫面成設計資料
               IF NOT lb_result THEN GOTO _ROLLBACK_FLAG END IF
               #產生畫面欄位資訊
               DISPLAY "      ","sadzp168_4(",p_prog,",",l_curr_spec_revision,",",ls_curr_identity,")"
               CALL sadzp168_4(p_prog, l_curr_spec_revision, ls_curr_identity) RETURNING lb_result,ls_err_msg3 #失敗沒關係
               IF NOT lb_result THEN
                  DISPLAY "WARNING : ",ls_err_msg3
               END IF
            END IF
         WHEN "PR"
            IF ls_type="G" OR ls_type="X" THEN 
               #簽出報表元件設計資料(adzp188):最後一個參數:N代表New,D代表Delete
               DISPLAY "      ","sadzp188_copy_rep_table(",p_prog,",",l_prev_code_revision,",",ls_prev_identity,",",l_curr_code_revision,",",ls_curr_identity,",N)"
               IF NOT cl_null(l_prev_code_revision) THEN #20230720 -add
                  CALL sadzp188_copy_rep_table(p_prog, l_prev_code_revision, ls_prev_identity, l_curr_code_revision, ls_curr_identity, "N") RETURNING lb_result
                  IF NOT lb_result THEN
                     LET ls_err_msg = "ERROR : sadzp060_2_after_check_out_for_download : call sadzp188_copy_rep_table fail!"
                     GOTO _ROLLBACK_FLAG
                  END IF
               END IF
            END IF #ls_type="G" OR ls_type="X"

            #230216-00010#1 -s
            IF ls_type="K" THEN #XR
               DISPLAY "      ","sadzp199_01_copy_xr_table(",p_prog,",",l_prev_code_revision,",",ls_prev_identity,",",l_curr_code_revision,",",ls_curr_identity,")"
               CALL sadzp199_01_copy_xr_table(p_prog, l_prev_code_revision, ls_prev_identity, l_curr_code_revision, ls_curr_identity) RETURNING lb_result
               IF NOT lb_result THEN
                  LET ls_err_msg = "ERROR : sadzp060_2_after_check_out_for_download : call sadzp199_01_copy_xr_table fail!"
                  GOTO _ROLLBACK_FLAG
               END IF
            END IF
            #230216-00010#1 -e
      
            #建立代碼樹,Section樹
            DISPLAY "      ","sadzp060_2_create_code_structure()..."   
            CALL sadzp060_2_create_code_structure(p_prog, p_curr_dzaf.*, p_prev_dzaf.*) RETURNING ls_err_msg
            IF NOT cl_null(ls_err_msg) THEN
              GOTO _ROLLBACK_FLAG
            END IF

            IF ls_type="G" THEN #簽出樣版(4rp)
               #搬移4rp
               #G類報表幾乎都是透過程式簽出做標準轉客製,因此1c-->2c就不需要搬移4rp,可以省下一次的搬移的時間.
               #NOTE:標準轉客製的時候,adzp050會呼叫2次的sadzp060_2_after_check_out_for_download(),所以底下是排除第2次的判斷.
               IF NOT ((ls_prev_identity='c' AND l_prev_code_revision=1) AND
                       (ls_curr_identity='c' AND l_curr_code_revision=2)) THEN
                  CALL lo_template_list.clear()
                  LET lo_program_info.pi_MODULE = ls_prev_module
                  LET lo_program_info.pi_NAME = p_prog
                  #取得樣板陣列  
                  DISPLAY "      ","sadzp050_tple_get_template_records()..."   
                  CALL sadzp050_tple_get_template_records(g_lang, lo_program_info.*, lo_template_list) RETURNING lb_result,lo_template_list
                  IF NOT lb_result THEN GOTO _ROLLBACK_FLAG END IF
                  
                  #LET ls_from_4rp_dir = os.path.JOIN(FGL_GETENV(UPSHIFT(ls_prev_module)), "4rp")
                  LET ls_from_4rp_dir = os.path.JOIN(ls_prev_module_path, "4rp") #171120-00029
                  IF ls_prev_identity <> ls_curr_identity THEN
                     #要搬移標準目錄的4rp到客製目錄的4rp下,例如AXM到CXM之類的4rp搬移
                     #LET ls_to_4rp_dir = os.path.JOIN(FGL_GETENV(UPSHIFT(ls_curr_module)), "4rp") #語系之前的路徑
                     LET ls_to_4rp_dir = os.path.JOIN(ls_curr_module_path, "4rp") #語系之前的路徑 #171120-00029
                     DISPLAY "      ","sadzp050_zip_moving_4rp_files(..., ",ls_from_4rp_dir,", ",ls_to_4rp_dir,", ...)"
                     CALL sadzp050_zip_moving_4rp_files(p_prog, ls_from_4rp_dir, ls_to_4rp_dir, lo_template_list) RETURNING lb_result
                     IF NOT lb_result THEN 
                        LET ls_err_msg = "ERROR : sadzp060_2_after_check_out_for_download : call sadzp050_zip_moving_4rp_files() to ",ls_to_4rp_dir," fail!"
                        GOTO _ROLLBACK_FLAG
                     END IF
                  
                     #做報表主機的4rp/語系目錄/4rp的複製動作
                     #ex : D:/tiptop_gr/t10dev/aap/4rp/zh_TW/aapr300_g01.4rp
                     LET ls_mnt4rp_module_path = os.path.JOIN(cl_rpt_get_env_global("MNT4RP") CLIPPED, DOWNSHIFT(ls_curr_module))
                     LET ls_mnt4rp_4rp_dir = os.path.JOIN(ls_mnt4rp_module_path, "4rp")
                     DISPLAY "      ","sadzp050_zip_moving_4rp_files(..., ",ls_from_4rp_dir,", ",ls_mnt4rp_4rp_dir,", ...)"
                     CALL sadzp050_zip_moving_4rp_files(p_prog, ls_from_4rp_dir, ls_mnt4rp_4rp_dir, lo_template_list) RETURNING lb_result
                     IF NOT lb_result THEN 
                        LET ls_err_msg = "ERROR : sadzp060_2_after_check_out_for_download : call sadzp050_zip_moving_4rp_files() to ",ls_mnt4rp_4rp_dir," fail!"
                        GOTO _ROLLBACK_FLAG
                     END IF
                  END IF
                  
                  #4rp/語系目錄複製到4rp/back/語系目錄的動作
                  #LET ls_to_4rp_dir = os.path.JOIN(os.path.JOIN(FGL_GETENV(UPSHIFT(ls_curr_module)), "4rp"), "back") #語系之前的路徑
                  LET ls_to_4rp_dir = os.path.JOIN(os.path.JOIN(ls_curr_module_path, "4rp"), "back") #語系之前的路徑 #171120-00029
                  DISPLAY "      ","sadzp050_zip_moving_4rp_files(..., ",ls_from_4rp_dir,", ",ls_to_4rp_dir,", ...)"
                  CALL sadzp050_zip_moving_4rp_files(p_prog, ls_from_4rp_dir, ls_to_4rp_dir, lo_template_list) RETURNING lb_result
                  IF NOT lb_result THEN 
                     #230216-00010#1 -s #GR複製備份檔失敗就算了,不影響整個簽出.
                    #LET ls_err_msg = "ERROR : sadzp060_2_after_check_out_for_download : call sadzp050_zip_copy_4rp_list ",ls_to_4rp_dir," fail!"
                    #GOTO _ROLLBACK_FLAG
                     LET ls_err_msg = "WARNING : sadzp060_2_after_check_out_for_download : call sadzp050_zip_copy_4rp_list ",ls_to_4rp_dir," fail!"
                     DISPLAY ls_err_msg
                     #230216-00010#1 -e
                  END IF
               END IF
            END IF #ls_type="G"
            
            #230216-00010#1 -s
            IF ls_type="K" THEN #XR
               LET ls_prog_id_first_char = p_prog.getCharAt(1)
               #搬移xrp
               IF NOT ((ls_prev_identity='c' AND l_prev_code_revision=1) AND
                       (ls_curr_identity='c' AND l_curr_code_revision=2) AND
                       ls_prog_id_first_char.equals("a")) THEN #todo9 #正常下載:不是 標準轉客製 連續簽出(c1-->c2)的情況.
                  DISPLAY "[DEBUG]sadzp060_2_after_check_out_for_download():XR 正常下載."
                  CALL lo_template_list.clear()
                  LET lo_program_info.pi_MODULE = ls_prev_module
                  LET lo_program_info.pi_NAME = p_prog
                  #取得樣板陣列  
                  DISPLAY "      ","sadzp050_tple_get_template_records()..."   
                  CALL sadzp050_tple_get_template_records(g_lang, lo_program_info.*, lo_template_list) RETURNING lb_result,lo_template_list
                  IF NOT lb_result THEN
                     GOTO _ROLLBACK_FLAG
                  END IF
                  
                  LET ls_from_xrp_dir = os.path.JOIN(ls_prev_module_path, "xrp") #語系之前的路徑
                  LET ls_to_xrp_dir = os.path.JOIN(ls_curr_module_path, "xrp") #語系之前的路徑

                  #20230330 -s
                  #LET ls_from_xdd_dir = os.path.JOIN(ls_from_xrp_dir, "xdd")
                  LET ls_from_xdd_lang_dir = os.path.JOIN(ls_from_xrp_dir, g_lang)
                  LET ls_from_xdd_dir = os.path.JOIN(ls_from_xdd_lang_dir, "xdd")
                  #LET ls_to_xdd_dir = os.path.JOIN(ls_to_xrp_dir, "xdd")
                  LET ls_to_xdd_lang_dir = os.path.JOIN(ls_to_xrp_dir, g_lang)
                  LET ls_to_xdd_dir = os.path.JOIN(ls_to_xdd_lang_dir, "xdd")
                  #20230330 -e

                  IF ls_prev_identity <> ls_curr_identity THEN #標準轉客製
                     DISPLAY "[DEBUG]sadzp060_2_after_check_out_for_download():XR 標準轉客製下載."
                     #xdd也要搬.失敗就算了. #todo9
                     CALL sadzp060_2_xr_copy_xdd(p_prog, ls_from_xdd_dir, ls_to_xdd_dir) RETURNING lb_result9,ls_errmsg9

                     #要搬移標準目錄的xrp到客製目錄的4rp下,例如AXM到CXM之類的xrp搬移
                     DISPLAY "      ","sadzp050_zip_moving_4rp_files(..., ",ls_from_xrp_dir,", ",ls_to_xrp_dir,", ...)"
                     CALL sadzp050_zip_moving_4rp_files(p_prog, ls_from_xrp_dir, ls_to_xrp_dir, lo_template_list) RETURNING lb_result
                     IF NOT lb_result THEN 
                        LET ls_err_msg = "ERROR : sadzp060_2_after_check_out_for_download : call sadzp050_zip_moving_4rp_files() to ",ls_to_xrp_dir," fail!"
                        GOTO _ROLLBACK_FLAG
                     END IF
                     
                     #做報表主機的xrp/語系目錄/xrp的複製動作.
                     #ex : D:/tiptop_xr/t10dev/aap/xrp/zh_TW/aapr300_k01.xrp
                     LET ls_mntxrp_module_path = os.path.JOIN(cl_rpt_get_env_global("MNTXRP") CLIPPED, DOWNSHIFT(ls_curr_module))
                     LET ls_mntxrp_xrp_dir = os.path.JOIN(ls_mntxrp_module_path, "xrp")
                     DISPLAY "      ","sadzp050_zip_moving_4rp_files(..., ",ls_from_xrp_dir,", ",ls_mntxrp_xrp_dir,", ...)"
                     CALL sadzp050_zip_moving_4rp_files(p_prog, ls_from_xrp_dir, ls_mntxrp_xrp_dir, lo_template_list) RETURNING lb_result
                     IF NOT lb_result THEN 
                        LET ls_err_msg = "ERROR : sadzp060_2_after_check_out_for_download : call sadzp050_zip_moving_4rp_files() to ",ls_mntxrp_xrp_dir," fail!"
                        GOTO _ROLLBACK_FLAG
                     END IF
                     
                     #NOTE:報表主機不需要xdd與xmd.
                  END IF
                  
                  #檔案備份:xrp/語系目錄 複製到 xrp/back/語系目錄 的動作.失敗就算了.
                  #NOTE:1.剛簽出的時候,xrp還是上一版.但此FUNCTION在每次[簽出之下載]都會呼叫,所以只能做一次,避免真正的備份檔覆蓋掉.
                  #     2.sadzp050_zip_moving_4rp_files()會自己找語系,並且是做[copy],不是[move].
                  #     3.若是標準轉客製當下的c1-->c2可以不用搬;但若純客製的c1-->c2還是得搬.
                  #     4.xrp既然都有做了,那xdd肯定也會一起處理.
                  LET ls_to_xrp_back_dir = os.path.JOIN(ls_to_xrp_dir, "back")

                  #xdd備份失敗就算了. #todo9
                  #20230330 -s
                  #LET ls_to_xdd_back_dir = os.path.JOIN(ls_to_xrp_back_dir, "xdd")
                  LET ls_to_xdd_back_lang_dir = os.path.JOIN(ls_to_xrp_back_dir, g_lang)
                  LET ls_to_xdd_back_dir = os.path.JOIN(ls_to_xdd_back_lang_dir, "xdd")
                  #20230330 -e

                  CALL sadzp060_2_xr_backup_xdd(p_prog, l_prev_code_revision, ls_prev_identity, ls_from_xdd_dir, ls_to_xdd_back_dir) RETURNING lb_result9,ls_errmsg9

                  CALL sadzp060_2_xr_get_xrp_backup_list(p_prog, l_prev_code_revision, ls_prev_identity, ls_to_xrp_back_dir, lo_template_list) RETURNING lb_has_list,lo_xrp_back_list
                  IF lb_has_list THEN #todo9
                     DISPLAY "      ","sadzp050_zip_moving_4rp_files(..., ",ls_from_xrp_dir,", ",ls_to_xrp_back_dir,", ...)"
                     CALL sadzp050_zip_moving_4rp_files(p_prog, ls_from_xrp_dir, ls_to_xrp_back_dir, lo_xrp_back_list) RETURNING lb_result
                     IF lb_result THEN 
                        #依據gzgw002(XR報表樣板編號)取得gzgw007(XR樣板名稱),然後將這些樣板名稱對應的xrp/xdd的副檔名之後加上'.$客製標示+版次'.
                        #例如xrp/axmr500_k01.xrp.s8
                        #    xrp/xdd/axmr500_k01.xdd.s8
                        CALL sadzp060_2_xr_rename_backup_xrp("O", p_prog, l_prev_code_revision, ls_prev_identity, lo_xrp_back_list, ls_to_xrp_back_dir) RETURNING lb_result9,ls_errmsg9
                        IF NOT lb_result9 THEN
                           DISPLAY ls_errmsg9
                           #GOTO _ROLLBACK_FLAG #失敗就算了.
                        END IF
                     ELSE
                        #XR複製備份檔失敗就比較嚴重,會影響後續還原的正確性.
                        LET ls_err_msg = "WARNING : sadzp060_2_after_check_out_for_download : call sadzp050_zip_copy_xrp_list ",ls_to_xrp_back_dir," fail!"
                        DISPLAY ls_err_msg
                        #GOTO _ROLLBACK_FLAG #失敗就算了.
                     END IF
                  END IF
               ELSE
                  DISPLAY "[DEBUG]sadzp060_2_after_check_out_for_download():XR 標準轉客製之連續簽出下載."
               END IF
            END IF
            #230216-00010#1 -e

            #簽出下載程式就會做標準重產流程,所以只需要在標準變客製才會需要重產4gl,要不然會多做而影響效能.
            IF ls_prev_identity <> ls_curr_identity THEN 
               #要做4gl重新產生的動作,且要做編譯,避免測試時發生找不到執行檔的問題
               #Begin:20160121 by Hiko
               #Begin:171120-00029:移除一堆註解.
               #1.前面呼叫s_azzi900_upd_gzza011()時,非M與非Z類程式有特別產生標準模組.42x與客製模組.42x.
               #  但此時因為客製程式並還沒有產生(還沒有r.c3),所以此時是多做了,可能需要移除這兩個動作.
               #2.adzp050呼叫sadzp060_2_after_check_out_for_download()之後會執行r.l ALL,裡面包含兩個動作:
               #  (2.1)產生客製模組.42x.
               #  (2.2)將所有用到此非M與非Z類程式的主程式重新link.
               #3.由於上述2個原因,所以這邊呼叫r.c3就只做到重產,編譯與link自己做.
               #CALL sadzp060_2_rc3(p_prog, p_curr_dzaf.*, '0') RETURNING lb_result,ls_err_msg #20160118 by Hiko:要做編譯與link, 但有錯誤也沒關係.
               DISPLAY "      ","sadzp060_2_rc3()..."   
               CALL sadzp060_2_rc3(p_prog, p_curr_dzaf.*, '1') RETURNING lb_result,ls_err_msg #171120-00029:只做到組合.
               #LET ls_to_4gl_dir = os.path.JOIN(FGL_GETENV(UPSHIFT(ls_curr_module)), "4gl")
               LET ls_to_4gl_dir = os.path.JOIN(ls_curr_module_path, "4gl") #171120-00029
               DISPLAY "      ","cl_change_dir(",ls_to_4gl_dir,")"
               IF cl_change_dir(ls_to_4gl_dir) THEN
                  LET ls_cmd = "r.c ",p_prog
                  IF ls_type="G" THEN #編譯同時要產生rdd.
                     LET ls_cmd = ls_cmd," rdd"
                  END IF
                  DISPLAY "      ",ls_cmd
                  CALL cl_cmdrun_openpipe("r.c", ls_cmd, FALSE) RETURNING lb_result,ls_err_msg3
                  IF lb_result THEN
                     IF ls_type="M" OR ls_type="Z" THEN
                        LET ls_cmd = "r.l ",p_prog
                        DISPLAY "      ",ls_cmd
                        RUN ls_cmd #只需要做完link才往下走,並不需要再收集訊息.
                     ELSE
                        #非M與非Z程式,要在編譯之後重產原模組的42x.
                        #此處不採用丟到背景+WITHOUT WAITING原因:後面流程還會補一個r.l 新模組 ALL,要讓後面能吃到做完的原模組42x
                        LET ls_cmd = "r.l ",ls_prev_module
                        DISPLAY "      ",ls_cmd
                        RUN ls_cmd
                     END IF
                  ELSE
                     DISPLAY "WARNING : ",ls_err_msg3
                  END IF
               END IF
               #End:171120-00029
               #End:20160121
            END IF
      END CASE

      LET l_time_e = cl_get_current()
      LET l_time_r = l_time_e - l_time_s
      DISPLAY "### sadzp060_2_after_check_out_for_download : total time=",l_time_r

      RETURN NULL #成功就沒有回傳錯誤訊息
      
      LABEL _ROLLBACK_FLAG:
         IF ls_prev_identity <> ls_curr_identity THEN
            #Begin:171120-00029:adzp050已經有錯誤訊息了,所以這邊就不再額外給訊息,以免使用者看不懂.
            # LET ls_err_msg = "ERROR : ",ls_err_msg,' ',cl_getmsg("adz-00390", g_lang) 
            # DISPLAY "ERROR : ",ls_err_msg
            LET ls_err_msg = "ERROR" #只要讓adzp050知道有錯誤就可以了.
            #End:171120-00029

            #230216-00010#1 DZAX_T -s
           #CALL sadzp060_2_del_dzax(p_prog, ls_curr_identity) RETURNING ls_err_msg2
           #IF NOT cl_null(ls_err_msg2) THEN #有錯誤訊息表示失敗
           #   #Begin:171120-00029:不要再合併訊息了,直接印出來.
           #   #LET ls_err_msg = ls_err_msg,' ',ls_err_msg2
           #   DISPLAY ls_err_msg2
           #   #End:171120-00029
           #END IF
            #230216-00010#1 DZAX_T -e

            #Begin:171120-00029:標準轉客製失敗時,adzp050會呼叫sadzp063_1的退回標準功能,所以這邊就不用重複做了.
            # #還原簽出(s --> c)時更新註冊的[是否客製]:N(s)
            # IF ls_type<>"Q" THEN
            #    CALL s_azzi900_upd_gzza011(p_prog, ls_type, ls_prev_identity) RETURNING lb_result
            #    IF NOT lb_result THEN 
            #       LET ls_err_msg = ls_err_msg," ERROR : Updated registration information fail when download!"
            #    END IF
            # END IF  
            #End:171120-00029
         END IF
      
         RETURN ls_err_msg
   CATCH
      CALL sadzp060_2_err_catch(ls_trigger, ls_sql)
      LET ls_err_msg = "ERROR : call sadzp060_2_after_check_out_for_download fail!"
      RETURN ls_err_msg
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 異動dzax_t.
# Input parameter : p_prog 程式代號
#                 : p_type 建構類型
#                 : p_new_identity 新的客製標示
#                 : p_old_identity 舊的客製標示
# Return code     : BOOLEAN TRUE:成功;FALSE:失敗
#                 : STRING 錯誤訊息
# Date & Author   : 2014/07/29 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_2_ins_dzax(p_prog, p_type, p_new_identity, p_old_identity)
   DEFINE p_prog         LIKE dzaf_t.dzaf001,
          p_type         LIKE dzaf_t.dzaf005, 
          p_new_identity LIKE dzaf_t.dzaf010, 
          p_old_identity LIKE dzaf_t.dzaf010
   DEFINE ls_sql       STRING,
          ls_where     STRING,
          li_cnt       SMALLINT,
          ls_err_msg   STRING,
          lb_rtn       BOOLEAN,
          l_free_style LIKE dzax_t.dzax003  
   DEFINE ls_sql9 STRING,  #20230626:DZAX_T -s
          li_cnt9 SMALLINT #20230626:DZAX_T -e

   TRY
      #230216-00010#1 DZAX_T -s
     ##Begin:200215-00009
     #LET ls_where = " where dzax001='",p_prog CLIPPED,"' and dzax006='",p_new_identity,"'" 
     #LET ls_sql = "select count(*)",
     #              " from dzax_t",ls_where
     ##End:200215-00009
     #PREPARE dzax_prep1 FROM ls_sql
     #EXECUTE dzax_prep1 INTO li_cnt
     #FREE dzax_prep1
     #
     #IF li_cnt=0 THEN #找不到新的客製標示就要新增了.
     #   #Begin:200215-00009
     #   LET ls_where = " where dzax001='",p_prog CLIPPED,"' and dzax006='",p_old_identity,"'" 
     #   LET ls_sql = "select count(*)",
     #                 " from dzax_t",ls_where
     #   #End:200215-00009
     #   PREPARE dzax_prep2 FROM ls_sql
     #   EXECUTE dzax_prep2 INTO li_cnt
     #   FREE dzax_prep2

     #   IF li_cnt=0 THEN #連舊的客製標示都找不到(這通常是非r.a的程式)
     #      LET l_free_style = "N" 
     #      IF p_type="B" OR p_type="W" OR p_type="Z" OR p_type="Q" THEN #B,W,Z,Q的程式預設為Free Style
     #         LET l_free_style = "Y"
     #      END IF
     #      #Begin:200215-00009
     #      LET ls_sql = "insert into dzax_t(dzax001,dzax002,dzax003,dzax004,dzax005,dzax006,dzax007,",
     #                                      "dzaxcrtdt,dzaxcrtdp,dzaxowndp,dzaxownid,dzaxstus,dzaxcrtid)",
     #      #Begin:170919-00003
     #                              # " VALUES('",p_prog,"','",ps_type,"','",ls_free_style,"','N','",g_customer,"','",p_new_identity,"','',", #dzax003預設為"Y"是為了避開沒有設計資料的時候tab_gen的問題.
     #                              #         "?,'",gs_dept,"','",gs_dept,"','",g_user,"','Y','",g_user,"')"
     #                              " values(?,?,?,'N',?,?,NULL,",
     #                                      "?,?,?,?,'Y',?)"
     #      #End:200215-00009
     #      PREPARE dzax_prep91 FROM ls_sql
     #      EXECUTE dzax_prep91 USING p_prog,p_type,l_free_style,g_customer,p_new_identity,
     #                                g_date,g_dept,g_dept,g_user,g_user
     #      FREE dzax_prep91
     #      #End:170919-00003
     #   ELSE #找到舊的dzax_t就當作新增基礎(這通常是標準變客製)
     #      #Begin:200215-00009
     #      LET ls_sql = "insert into dzax_t(dzax001,dzax002,dzax003,dzax004,dzax005,dzax006,dzax007,",
     #                                      "dzaxcrtdt,dzaxcrtdp,dzaxowndp,dzaxownid,dzaxstus,dzaxcrtid)",
     #      #Begin:170919-00003
     #                              # " SELECT dzax001,dzax002,dzax003,'N','",g_customer,"','",p_new_identity,"',dzax007,",
     #                              #         "?,'",gs_dept CLIPPED,"','",gs_dept CLIPPED,"','",g_user,"',dzaxstus,'",g_user,"'",
     #                              #   " FROM dzax_t",ls_where
     #                              " select dzax001,dzax002,dzax003,'N',?,?,dzax007,",
     #                                      "?,?,?,?,dzaxstus,?",
     #                                " from dzax_t",ls_where
     #      #End:200215-00009
     #      PREPARE dzax_prep92 FROM ls_sql
     #      EXECUTE dzax_prep92 USING g_customer,p_new_identity,
     #                                g_date,g_dept,g_dept,g_user,g_user
     #      FREE dzax_prep92
     #      #End:170919-00003
     #   END IF
     #END IF
      #230216-00010#1 DZAX_T -e

      #230216-00010#1 DZAX_T -s
      LET ls_sql = "SELECT count(*)",
                   "  FROM  dzax_t",
                   " WHERE dzax001 = '",p_prog CLIPPED,"'",
                   "   AND dzax006 = '",p_new_identity,"'"
      PREPARE prep_dzax_cnt FROM ls_sql
      EXECUTE prep_dzax_cnt INTO li_cnt
      FREE prep_dzax_cnt
      
      IF li_cnt=0 THEN
         IF p_new_identity<>p_old_identity THEN
            #20230626 -s #若連標準資料都沒有,那就要產生新的dzax_t.
            LET ls_sql9 = "SELECT count(*)",
                         "  FROM  dzax_t",
                         " WHERE dzax001 = '",p_prog CLIPPED,"'",
                         "   AND dzax006 = '",p_old_identity,"'"
            PREPARE prep_dzax_cnt9 FROM ls_sql9
            EXECUTE prep_dzax_cnt9 INTO li_cnt9
            FREE prep_dzax_cnt9
            
            IF li_cnt9>0 THEN #20230626 -e
               #標準轉客製就從標準複製一筆為客製.
               LET ls_sql = "INSERT INTO dzax_t(dzax001,dzax002,dzax003,dzax004,dzax005,dzax006,dzax007,",
                            "                   dzaxcrtdt,dzaxcrtdp,dzaxowndp,dzaxownid,dzaxstus,dzaxcrtid)",
                            "            SELECT dzax001,dzax002,dzax003,'N',?,?,dzax007,",
                            "                   ?,?,?,?,dzaxstus,?",
                            "              FROM dzax_t",
                            "             WHERE dzax001 = '",p_prog CLIPPED,"'",
                            "               AND dzax006 = '",p_old_identity,"'" #s
               PREPARE prep_ins_dzax1 FROM ls_sql
               EXECUTE prep_ins_dzax1 USING g_customer,p_new_identity, #c
                                            g_date,g_dept,g_dept,g_user,g_user
               FREE prep_ins_dzax1
            ELSE #20230626 -s
               #若連標準資料都沒有就直接新增一筆標準,以及一筆客製. #備註:此段和下方完全相同.
               IF p_type="B" OR p_type="W" OR p_type="Z" OR p_type="Q" THEN #B,W,Z,Q的程式預設為Free Style
                  LET l_free_style = "Y"
               END IF
               
               LET ls_sql9 = "INSERT INTO dzax_t(dzax001,dzax002,dzax003,dzax004,dzax005,dzax006,dzax007,",
                             "                   dzaxcrtdt,dzaxcrtdp,dzaxowndp,dzaxownid,dzaxstus,dzaxcrtid)",
                             "            VALUES(?,?,?,'N',?,?,NULL,",
                             "                   ?,?,?,?,'Y',?)"
               PREPARE prep_ins_dzax91 FROM ls_sql9
               EXECUTE prep_ins_dzax91 USING p_prog,p_type,l_free_style,g_customer,p_old_identity,
                                             g_date,g_dept,g_dept,g_user,g_user
               FREE prep_ins_dzax91

               LET ls_sql9 = "INSERT INTO dzax_t(dzax001,dzax002,dzax003,dzax004,dzax005,dzax006,dzax007,",
                             "                   dzaxcrtdt,dzaxcrtdp,dzaxowndp,dzaxownid,dzaxstus,dzaxcrtid)",
                             "            VALUES(?,?,?,'N',?,?,NULL,",
                             "                   ?,?,?,?,'Y',?)"
               PREPARE prep_ins_dzax92 FROM ls_sql9
               EXECUTE prep_ins_dzax92 USING p_prog,p_type,l_free_style,g_customer,p_new_identity,
                                             g_date,g_dept,g_dept,g_user,g_user
               FREE prep_ins_dzax92
            END IF #20230626 -e
         ELSE
            #非標準轉客製就直接新增一筆.
            IF p_type="B" OR p_type="W" OR p_type="Z" OR p_type="Q" THEN #B,W,Z,Q的程式預設為Free Style
               LET l_free_style = "Y"
            END IF

            LET ls_sql = "INSERT INTO dzax_t(dzax001,dzax002,dzax003,dzax004,dzax005,dzax006,dzax007,",
                         "                   dzaxcrtdt,dzaxcrtdp,dzaxowndp,dzaxownid,dzaxstus,dzaxcrtid)",
                         "            VALUES(?,?,?,'N',?,?,NULL,",
                         "                   ?,?,?,?,'Y',?)"
            PREPARE prep_ins_dzax2 FROM ls_sql
            EXECUTE prep_ins_dzax2 USING p_prog,p_type,l_free_style,g_customer,p_new_identity,
                                         g_date,g_dept,g_dept,g_user,g_user
            FREE prep_ins_dzax2
         END IF
      END IF 
      #230216-00010#1 DZAX_T -e

      RETURN TRUE,NULL
   CATCH
      CALL sadzp060_2_err_catch('', ls_sql)
      LET ls_err_msg = "ERROR : call sadzp060_2_ins_dzax fail!"
      RETURN FALSE,ls_err_msg
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 刪除dzax_t.
# Input parameter : p_prog 程式代號
#                 : p_new_identity 新的客製標示
# Return code     : STRING
# Date & Author   : 2014/10/07 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_2_del_dzax(p_prog, p_new_identity)
   DEFINE p_prog         STRING,
          p_new_identity STRING
   DEFINE ls_sql     STRING,
          ls_where   STRING,
          li_cnt     SMALLINT,
          ls_err_msg STRING

   TRY
      #Begin:200215-00009
      LET ls_where = " where dzax001='",p_prog,"' and dzax006='",p_new_identity,"'" 
      LET ls_sql = "select count(*)",
                    " from dzax_t",ls_where
      #End:200215-00009
      PREPARE dzax_prep3 FROM ls_sql
      EXECUTE dzax_prep3 INTO li_cnt
      FREE dzax_prep3
      
      IF li_cnt>0 THEN
         LET ls_sql = "delete from dzax_t",ls_where    #200215-00009
         EXECUTE IMMEDIATE ls_sql
      END IF
      
      RETURN NULL
   CATCH
      CALL sadzp060_2_err_catch('', ls_sql)
      LET ls_err_msg = "ERROR : call sadzp060_2_del_dzax fail!"
      RETURN ls_err_msg
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC
# Descriptions    : 程式重產與重組
# Input parameter : p_prog 程式代號
#                 : p_dzaf 版次資訊
#                 : p_rc3_flag r.c3的選項(0:全做;1:做到產生和組合,不做編譯和鏈結;2:不做產生,只做組合,包含編譯和鏈結)
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
#                 : STRING 錯誤訊息
# Date & Author   : 2014/10/08 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_2_rc3(p_prog, p_dzaf, p_rc3_flag)
   DEFINE p_prog     STRING,
          p_dzaf     T_DZAF_T,
          p_rc3_flag STRING 
   DEFINE l_type          LIKE dzaf_t.dzaf005,
          l_spec_revision LIKE dzaf_t.dzaf003,
          l_code_revision LIKE dzaf_t.dzaf004,
          l_identity      LIKE dzaf_t.dzaf010,
          ls_cmd          STRING,
          lb_result       BOOLEAN,
          ls_err_msg      STRING,
          ls_err_msg3     STRING

   LET l_type = p_dzaf.dzaf005 CLIPPED
   LET l_spec_revision = p_dzaf.dzaf003
   LET l_code_revision = p_dzaf.dzaf004
   LET l_identity = p_dzaf.dzaf010 CLIPPED

   #要做4gl重新產生的動作(不做編譯..)
   IF l_type="G" OR l_type="X" OR l_type="K" THEN #報表元件設計資料的版次是屬於"程式" #230216-00010#1:增加K類.
      DISPLAY "[INFO]sadzp060_2_rc3 : call sadzp030_tab_gen(",p_prog,",",l_code_revision,",'',",l_identity,")"
      CALL sadzp030_tab_gen(p_prog, l_code_revision, "", l_identity) RETURNING lb_result
   ELSE
      DISPLAY "[INFO]sadzp060_2_rc3 : call sadzp030_tab_gen(",p_prog,",",l_spec_revision,",'',",l_identity,")"
      CALL sadzp030_tab_gen(p_prog, l_spec_revision, "", l_identity) RETURNING lb_result
   END IF
   
   IF NOT lb_result THEN
      LET ls_err_msg = "ERROR : call sadzp030_tab_gen fail!"
      RETURN FALSE,ls_err_msg
   ELSE
      LET ls_cmd = "r.c3 ",p_prog," ","''"," ",l_code_revision," ",p_rc3_flag," ",l_identity
      DISPLAY "      ","cl_cmdrun_openpipe(r.c, ",ls_cmd,", FALSE)"
      CALL cl_cmdrun_openpipe("r.c3", ls_cmd, FALSE) RETURNING lb_result,ls_err_msg3
      IF NOT lb_result THEN
         RETURN FALSE,ls_err_msg3 #真實反應對錯,由呼叫端控制流程
      END IF
   END IF

   RETURN TRUE,NULL
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC
# Descriptions    : 在取消簽出之後要做的事情.
# Input parameter : p_prog 程式代號
#                 : p_role 角色(SD/PR)
#                 : p_curr_dzaf 目前版次資訊
#                 : p_prev_dzaf 前版版次資訊
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2014/05/23 by Hiko
#                 : 20171122 by 02632 : INPORTMENT:adzp050已經沒有取消簽出功能,所以此段其實可以不用,但要和adzp050一起處理.
##########################################################################
PUBLIC FUNCTION sadzp060_2_after_recall(p_prog, p_role, p_curr_dzaf, p_prev_dzaf)
   DEFINE p_prog      STRING,
          p_role      STRING,
          p_curr_dzaf T_DZAF_T, #範例:版次2
          p_prev_dzaf T_DZAF_T  #範例:版次1
   DEFINE ls_type               STRING,
          l_curr_spec_revision  LIKE dzaf_t.dzaf003, #recall:5-->4:curr=5;download:4-->5:curr=4
          l_curr_code_revision  LIKE dzaf_t.dzaf004,
          ls_curr_identity      STRING,
          ls_curr_module        STRING,
          ls_curr_module_path   STRING,
          l_prev_spec_revision  LIKE dzaf_t.dzaf003,
          l_prev_code_revision  LIKE dzaf_t.dzaf004,
          ls_prev_identity      STRING,
          ls_prev_module        STRING,
          ls_prev_module_path   STRING,
          ls_trigger            STRING,
          ls_sql                STRING,
          lb_result             BOOLEAN,
          ls_err_msg            STRING,
          ls_err_msg2           STRING,
          ls_err_msg3           STRING,
          lo_program_info       T_PROGRAM_INFO,
          lo_template_list      DYNAMIC ARRAY OF T_TEMPLATE_LIST_R,
          ls_from_4rp_dir       STRING,
          ls_to_4rp_dir         STRING,
          ls_to_4fd_dir         STRING,
          ls_4fd                STRING,
          ls_4fd_dir            STRING,
          ls_cmd                STRING,
          ls_temp_file          STRING,
          ls_err_code           STRING,
          ls_owner              STRING,
          ls_replace_arg        STRING,
          li_idx                SMALLINT,
          la_gzzb               DYNAMIC ARRAY OF RECORD
                                gzzb001 LIKE gzzb_t.gzzb001
                                END RECORD,
          lb_continue           BOOLEAN,
          ls_mnt4rp_module_path STRING,
          ls_mnt4rp_4rp_dir     STRING 

   #取消簽出流程說明:
   #情境三.s4<--s5 or c2<--c3
   #3-1.刪除樹
   #3-2.還原4rp
   #3-3.4fd重產
   #3-4.4gl重產
   #情四.s5<--c1
   #4-0.更新客製(c-->s)
   #4-1~4-4同3-1~3-4

   TRY
      IF NOT sadzp060_2_init_var() THEN
         LET ls_err_msg = "ERROR : sadzp060_2_after_recall.sadzp060_2_init_var"
         GOTO _ROLLBACK_FLAG
      END IF
      
      LET p_prog = p_prog.trim()
      LET ls_type = p_curr_dzaf.dzaf005 CLIPPED
      LET l_curr_spec_revision = p_curr_dzaf.dzaf003 CLIPPED
      LET l_curr_code_revision = p_curr_dzaf.dzaf004 CLIPPED
      LET ls_curr_identity = p_curr_dzaf.dzaf010 CLIPPED
      LET ls_curr_module = p_curr_dzaf.dzaf006 CLIPPED
      LET ls_curr_module_path = FGL_GETENV(UPSHIFT(ls_curr_module))
      LET l_prev_spec_revision = p_prev_dzaf.dzaf003 CLIPPED
      LET l_prev_code_revision = p_prev_dzaf.dzaf004 CLIPPED
      LET ls_prev_identity = p_prev_dzaf.dzaf010 CLIPPED
      LET ls_prev_module = p_prev_dzaf.dzaf006 CLIPPED
      LET ls_prev_module_path = FGL_GETENV(UPSHIFT(ls_prev_module))

      DISPLAY "[INFO]sadzp060_2_after_recall parameter :"
      DISPLAY "      ","p_role=",p_role   
      DISPLAY "      ","p_prog=",p_prog   
      DISPLAY "      ","ls_type=",ls_type   
      DISPLAY "      ","l_curr_spec_revision=",l_curr_spec_revision   
      DISPLAY "      ","l_curr_code_revision=",l_curr_code_revision   
      DISPLAY "      ","ls_curr_identity=",ls_curr_identity   
      DISPLAY "      ","ls_curr_module=",ls_curr_module   
      DISPLAY "      ","ls_curr_module_path=",ls_curr_module_path   
      DISPLAY "      ","l_prev_spec_revision=",l_prev_spec_revision   
      DISPLAY "      ","l_prev_code_revision=",l_prev_code_revision   
      DISPLAY "      ","ls_prev_identity=",ls_prev_identity   
      DISPLAY "      ","ls_prev_module=",ls_prev_module   
      DISPLAY "      ","ls_prev_module_path=",ls_prev_module_path   

      #不分簽出的角色, M類程式要注意對應的行業程式是否已經簽出.
      IF ls_type.equals("M") THEN
         LET lb_continue = TRUE
         LET ls_sql = "select gzzb001 from gzzb_t where gzzb002<>'sd' and gzzb003='",p_prog,"' order by gzzb001"    #200215-00009
         PREPARE gzzb_prep FROM ls_sql
         DECLARE gzzb_curs CURSOR FOR gzzb_prep
         LET li_idx = 1
         FOREACH gzzb_curs INTO la_gzzb[li_idx].*
            LET lo_program_info.pi_NAME = la_gzzb[li_idx].gzzb001  #程式代碼
            LET lo_program_info.pi_TYPE = ls_type #類別
            CALL sadzp200_alm_get_alm_owner_by_role(lo_program_info.*, "SD") RETURNING lb_result,ls_owner
            IF NOT lb_result THEN #行業規格沒簽出繼續檢查行業程式是否簽出.
               CALL sadzp200_alm_get_alm_owner_by_role(lo_program_info.*, "PR") RETURNING lb_result,ls_owner
               IF lb_result THEN #行業程式已經被簽出.
                  LET lb_continue = FALSE
                  LET ls_err_code = "adz-00679" #對應的行業程式%1正被%2簽出中, 所以%3無法取消簽出.
                  LET ls_replace_arg = la_gzzb[li_idx].gzzb001,"|",ls_owner,"|",p_prog
                  EXIT FOREACH
               END IF
            ELSE
               LET lb_continue = FALSE
               LET ls_err_code = "adz-00678" #對應的行業規格%1正被%2簽出中, 所以%3無法取消簽出.
               LET ls_replace_arg = la_gzzb[li_idx].gzzb001,"|",ls_owner,"|",p_prog
               EXIT FOREACH
            END IF

            LET li_idx = li_idx + 1
         END FOREACH
         
         IF NOT lb_continue THEN
            CALL cl_replace_err_msg(cl_getmsg(ls_err_code, g_lang), ls_replace_arg) RETURNING ls_err_msg
            RETURN ls_err_msg
         END IF
      END IF

      CASE p_role
         WHEN "SD"
            IF cl_null(l_prev_spec_revision) THEN
               RETURN NULL
            END IF
         WHEN "PR"
            IF cl_null(l_prev_code_revision) THEN
               RETURN NULL
            END IF
      END CASE

      #以流程來說,不可能會有c --> s的情況,因為s --> c的當下就已經產生了1c並簽入,所以客戶看到的是2c的簽出,所以取消簽出只是2c --> 1c.
      
      CASE p_role
         WHEN "SD"
            CALL sadzp063_1_del_design_data(p_prev_dzaf.*, '1') RETURNING lb_result,ls_err_msg 
            IF NOT lb_result THEN GOTO _ROLLBACK_FLAG END IF
         WHEN "PR"
            IF ls_type="G" OR ls_type="X" THEN
               #刪除報表元件設計資料(adzp188)              
               DISPLAY "[INFO]sadzp060_2_after_recall : call sadzp188_copy_rep_table(",p_prog,",",l_prev_code_revision,",",ls_prev_identity,",",l_curr_code_revision,",",ls_curr_identity,",D)"
               IF NOT cl_null(l_curr_code_revision) THEN #20230720 -add
                  CALL sadzp188_copy_rep_table(p_prog, l_prev_code_revision, ls_prev_identity, l_curr_code_revision, ls_curr_identity, "D") RETURNING lb_result
                  IF NOT lb_result THEN
                     LET ls_err_msg = "ERROR : sadzp060_2_after_recall : call sadzp188_copy_rep_table fail! "
                     GOTO _ROLLBACK_FLAG
                  END IF
               END IF
            END IF

            IF ls_type="G" THEN 
               CALL lo_template_list.clear()
               LET lo_program_info.pi_MODULE = ls_prev_module
               LET lo_program_info.pi_NAME = p_prog
               #取得樣板陣列  
               CALL sadzp050_tple_get_template_records(g_lang, lo_program_info.*, lo_template_list) RETURNING lb_result,lo_template_list
               IF NOT lb_result THEN GOTO _ROLLBACK_FLAG END IF

               #將4rp/back/語系目錄下的4rp全部還原到對應的4rp/語系目錄
               LET ls_from_4rp_dir = os.path.JOIN(os.path.JOIN(ls_curr_module_path, "4rp"), "back")
               LET ls_to_4rp_dir = os.path.JOIN(FGL_GETENV(UPSHIFT(ls_prev_module)), "4rp") #語系之前的路徑
               DISPLAY "      ","sadzp050_zip_moving_4rp_files(..., ",ls_from_4rp_dir,", ",ls_to_4rp_dir,", ...)"
               CALL sadzp050_zip_moving_4rp_files(p_prog, ls_from_4rp_dir, ls_to_4rp_dir, lo_template_list) RETURNING lb_result
               IF NOT lb_result THEN 
                  LET ls_err_msg = "ERROR : sadzp060_2_after_recall : call sadzp050_zip_moving_4rp_files() to ",ls_from_4rp_dir," fail!"
                  GOTO _ROLLBACK_FLAG
               END IF

               #將4rp/語系目錄下的4rp全部還原到報表主機對應的4rp/語系目錄:同簽出的動作,但一定是在back搬到module後才可以這樣做
               LET ls_mnt4rp_module_path = os.path.JOIN(cl_rpt_get_env_global("MNT4RP") CLIPPED, DOWNSHIFT(ls_prev_module))
               LET ls_mnt4rp_4rp_dir = os.path.JOIN(ls_mnt4rp_module_path, "4rp")
               DISPLAY "      ","sadzp050_zip_moving_4rp_files(..., ",ls_from_4rp_dir,", ",ls_to_4rp_dir,", ...)"
               CALL sadzp050_zip_moving_4rp_files(p_prog, ls_from_4rp_dir, ls_mnt4rp_4rp_dir, lo_template_list) RETURNING lb_result
               IF NOT lb_result THEN 
                  LET ls_err_msg = "ERROR : sadzp060_2_after_recall : call sadzp050_zip_moving_4rp_files() to ",ls_mnt4rp_4rp_dir," fail!"
                  GOTO _ROLLBACK_FLAG
               END IF
            END IF
      END CASE
      
      RETURN NULL #成功就沒有回傳錯誤訊息
      
      LABEL _ROLLBACK_FLAG:
         #以流程來說,不可能會有c --> s的情況,因為s --> c的當下就已經產生了1c並簽入,所以客戶看到的是2c的簽出,所以取消簽出只是2c --> 1c.
         RETURN ls_err_msg
   CATCH
      CALL sadzp060_2_err_catch(ls_trigger, ls_sql)
      LET ls_err_msg = "ERROR : call sadzp060_2_after_recall fail!"
      RETURN ls_err_msg
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立規格結構樹.
# Input parameter : p_prog 程式代號
#                 : p_curr_dzaf 目前版次資訊
#                 : p_prev_dzaf 前版版次資訊
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2014/05/23 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_2_create_spec_structure(p_prog, p_curr_dzaf, p_prev_dzaf)
   DEFINE p_prog      STRING,
          p_curr_dzaf T_DZAF_T,
          p_prev_dzaf T_DZAF_T
   DEFINE ls_sql               STRING,
          ls_trigger           STRING,
          li_cnt               SMALLINT,
          ls_type              STRING,
          l_curr_spec_revision LIKE dzaf_t.dzaf003,
          l_prev_spec_revision LIKE dzaf_t.dzaf003,
          l_curr_identity      LIKE dzaf_t.dzaf010,
          l_prev_identity      LIKE dzaf_t.dzaf010, 
          ls_replace_arg       STRING, 
          lb_result            BOOLEAN,
          ls_err_msg           STRING

   TRY
      LET p_prog = p_prog.trim()
      LET ls_type = p_curr_dzaf.dzaf005 CLIPPED
      LET l_curr_spec_revision = p_curr_dzaf.dzaf003 CLIPPED
      LET l_prev_spec_revision = p_prev_dzaf.dzaf003 CLIPPED
      LET l_curr_identity = p_curr_dzaf.dzaf010 CLIPPED
      LET l_prev_identity = p_prev_dzaf.dzaf010 CLIPPED 
   
      IF cl_null(l_prev_spec_revision) OR l_prev_spec_revision=0 THEN
         RETURN NULL
      END IF

      LET ls_trigger = "sadzp060_2_create_spec_structure : check dzaa_t current ver data count"
      #標準轉客製的時候,要先將客製資料刪除,以免有殘留資料而造成資料不齊全.
      IF p_prev_dzaf.dzaf010='s' AND p_curr_dzaf.dzaf010='c' THEN
         #Begin:200215-00009
         LET ls_sql = "delete from dzaa_t",
                      " where dzaa001='",p_prog,"'",
                      "   and dzaa009='c'"
         #End:200215-00009
         EXECUTE IMMEDIATE ls_sql
      END IF

      #判斷規格樹(dzaa_t)是否存在目前正要下載的版本 : 沒有就要建立.
      #Begin:200215-00009
      LET ls_sql = "select count(*)",
                    " from dzaa_t",
                   " where dzaa001='",p_prog,"' and dzaa002='",l_curr_spec_revision,"' and dzaa009='",l_curr_identity,"'" #這裡不能加上dzaastus='Y'的條件,要不然底下就有可能出現重複插入資料的問題.
      #End:200215-00009
      PREPARE dzaa_prep1 FROM ls_sql
      EXECUTE dzaa_prep1 INTO li_cnt
      FREE dzaa_prep1

      IF li_cnt=0 THEN 
         IF ls_type MATCHES "[MSFQ]" THEN
            IF l_prev_identity=l_curr_identity THEN #相同客製標示才多這個判斷.
               IF l_prev_spec_revision>0 THEN
                  #Begin:200215-00009
                  LET ls_sql = "select count(*)",
                                " from dzaa_t",
                               " where dzaa001='",p_prog,"' and dzaa002='",l_prev_spec_revision,"' and dzaa009='",l_prev_identity,"'" #這裡不能加上dzaastus='Y'的條件,要不然底下就有可能出現重複插入資料的問題.
                  #End:200215-00009
                  PREPARE dzaa_prep9 FROM ls_sql
                  EXECUTE dzaa_prep9 INTO li_cnt
                  FREE dzaa_prep9
                  
                  IF li_cnt=0 THEN
                     LET ls_replace_arg = l_prev_spec_revision
                     LET ls_replace_arg = ls_replace_arg.trim(),"|SPEC"
                     LET ls_err_msg = cl_replace_err_msg(cl_getmsg("adz-00700", g_lang), ls_replace_arg) #找不到舊版(第%1版)的%2設計資料, 請查明問題後再繼續執行. 
                     DISPLAY "ERROR : ",ls_err_msg
                     RETURN ls_err_msg
                  END IF
               END IF
            END IF
            
            LET li_cnt = 0
         END IF

         #要以舊的版本當作基底建立規格樹.
         LET ls_trigger = "sadzp060_2_create_spec_structure : insert new ver dzaa_t"
         #Begin:200215-00009
         LET ls_sql = "insert into dzaa_t(dzaa001,dzaa002,dzaa003,dzaa004,dzaa005,",
                                         "dzaa006,dzaa007,dzaa008,dzaa009,dzaa010,",
                                         "dzaacrtdt,dzaacrtdp,dzaaowndp,dzaaownid,dzaastus,dzaacrtid)",
         #Begin:170919-00003
                                 # " SELECT dzaa001,'",l_curr_spec_revision,"',dzaa003,dzaa004,dzaa005,",
                                 #         "dzaa006,dzaa007,dzaa008,'",l_curr_identity,"','",g_customer,"',", 
                                 #         "?,'",gs_dept CLIPPED,"','",gs_dept CLIPPED,"','",g_user,"',dzaastus,'",g_user,"'",
                                 " select dzaa001,?,dzaa003,dzaa004,dzaa005,",
                                         "dzaa006,dzaa007,dzaa008,?,?,", 
                                         "?,?,?,?,dzaastus,?",
         #End:170919-00003
                                   " from dzaa_t",
                                  " where dzaa001='",p_prog,"'",
                                    " and dzaa002='",p_prev_dzaf.dzaf003 CLIPPED,"'",
                                    " and dzaa009='",p_prev_dzaf.dzaf010 CLIPPED,"'",
                                    " and dzaastus='Y'"
         #End:200215-00009
         PREPARE dzaa_prep3 FROM ls_sql
         #Begin:170919-00003
         EXECUTE dzaa_prep3 USING l_curr_spec_revision,
                                  l_curr_identity,g_customer,
                                  g_date,g_dept,g_dept,g_user,g_user
         FREE dzaa_prep3
         #End:170919-00003
      END IF

      RETURN NULL
   CATCH 
      CALL sadzp060_2_err_catch(ls_trigger, ls_sql) 
      RETURN "ERROR : sadzp060_2_create_spec_structure faliure!"
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立程式結構樹(包含add point和section)
# Input parameter : p_prog 程式代號
#                 : p_curr_dzaf 目前版次資訊
#                 : p_prev_dzaf 前版版次資訊
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2014/05/23 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_2_create_code_structure(p_prog, p_curr_dzaf, p_prev_dzaf)
   DEFINE p_prog      STRING,
          p_curr_dzaf T_DZAF_T,
          p_prev_dzaf T_DZAF_T
   DEFINE ls_sql               STRING,
          ls_trigger           STRING,
          li_cnt               SMALLINT,
          ls_type              STRING,
          l_curr_code_revision LIKE dzaf_t.dzaf004,
          l_prev_code_revision LIKE dzaf_t.dzaf004,
          l_curr_identity      LIKE dzaf_t.dzaf010,
          l_prev_identity      LIKE dzaf_t.dzaf010, 
          ls_replace_arg       STRING, 
          lb_result            BOOLEAN,
          ls_err_msg           STRING

   TRY
      LET p_prog = p_prog.trim()
      LET ls_type = p_curr_dzaf.dzaf005 CLIPPED
      LET l_curr_code_revision = p_curr_dzaf.dzaf004 CLIPPED
      LET l_prev_code_revision = p_prev_dzaf.dzaf004 CLIPPED
      LET l_curr_identity = p_curr_dzaf.dzaf010 CLIPPED
      LET l_prev_identity = p_prev_dzaf.dzaf010 CLIPPED 

      IF cl_null(l_prev_code_revision) OR l_prev_code_revision=0 THEN
         RETURN NULL
      END IF

      LET ls_trigger = "sadzp060_2_create_code_structure : check dzba_t current ver data count"
      #標準轉客製的時候,要先將客製資料刪除,以免有殘留資料而造成資料不齊全.
      IF p_prev_dzaf.dzaf010='s' AND p_curr_dzaf.dzaf010='c' THEN
         #Begin:200215-00009
         LET ls_sql = "delete from dzba_t",
                      " where dzba001='",p_prog,"'",
                      "   and dzba010='c'"
         #End:200215-00009
         EXECUTE IMMEDIATE ls_sql
      END IF

      #判斷add point樹(dzba_t)是否存在目前正要下載的版本 : 沒有就要建立.
      #Begin:200215-00009
      LET ls_sql = "select count(*)",
                    " from dzba_t",
                   " where dzba001='",p_prog,"' and dzba002='",l_curr_code_revision,"' and dzba010='",l_curr_identity,"'" #這裡不能加上dzbastus='Y'的條件,要不然底下就有可能出現重複插入資料的問題.
      #End:200215-00009
      PREPARE dzba_prep1 FROM ls_sql
      EXECUTE dzba_prep1 INTO li_cnt
      FREE dzba_prep1

      IF li_cnt=0 THEN 
         IF ls_type MATCHES "[MSQ]" THEN
            IF l_prev_identity=l_curr_identity THEN #相同客製標示才多這個判斷.
               IF l_prev_code_revision>0 THEN
                  #Begin:200215-00009
                  LET ls_sql = "select count(*)",
                                " from dzba_t",
                               " where dzba001='",p_prog,"' and dzba002='",l_prev_code_revision,"' and dzba010='",l_prev_identity,"'" #這裡不能加上dzbastus='Y'的條件,要不然底下就有可能出現重複插入資料的問題.
                  #End:200215-00009
                  PREPARE dzba_prep9 FROM ls_sql
                  EXECUTE dzba_prep9 INTO li_cnt
                  FREE dzba_prep9
                  
                  IF li_cnt=0 THEN
                     LET ls_replace_arg = l_prev_code_revision
                     LET ls_replace_arg = ls_replace_arg.trim(),"|ADD POINT"
                     LET ls_err_msg = cl_replace_err_msg(cl_getmsg("adz-00700", g_lang), ls_replace_arg) #找不到舊版(第%1版)的%2設計資料, 請查明問題後再繼續執行. 
                     DISPLAY "ERROR : ",ls_err_msg
                     RETURN ls_err_msg
                  END IF
               END IF
            END IF
            
            LET li_cnt = 0
         END IF

         #要以舊的版本當作基底建立程式樹.
         LET ls_trigger = "sadzp060_2_create_code_structure : insert new ver dzba_t"
         #170919-00003:因為簽出流程比較久,執行過程當中,有可能因為相關聯的程式有異動,導致此SQL出現'm'字元.
         #             為了避免單引號/雙引號的字串組合出現錯誤,所以將SELECT改用USING.
         #Begin:200215-00009
         LET ls_sql = "insert into dzba_t(dzba001,dzba002,dzba003,dzba004,dzba005,",
                                         "dzba006,dzba007,dzba008,dzba009,dzba010,dzba011,",
                                         "dzbacrtdt,dzbacrtdp,dzbaowndp,dzbaownid,dzbastus,dzbacrtid)",
         #Begin:170919-00003
                                 # " SELECT dzba001,'",l_curr_code_revision,"',dzba003,dzba004,dzba005,",
                                 #         "dzba006,dzba007,dzba008,dzba009,'",l_curr_identity,"','",g_customer,"',", 
                                 #         "?,'",gs_dept CLIPPED,"','",gs_dept CLIPPED,"','",g_user,"',dzbastus,'",g_user,"'",
                                 " select dzba001,?,dzba003,dzba004,dzba005,",
                                         "dzba006,dzba007,dzba008,dzba009,?,?,", 
                                         "?,?,?,?,dzbastus,?",
         #End:170919-00003
                                   " from dzba_t",
                                  " where dzba001='",p_prog,"'",
                                    " and dzba002='",p_prev_dzaf.dzaf004 CLIPPED,"'",
                                    " and dzba010='",p_prev_dzaf.dzaf010 CLIPPED,"'",
                                    " and dzbastus='Y'"
         #End:200215-00009
         PREPARE dzba_prep3 FROM ls_sql
         #Begin:170919-00003
         EXECUTE dzba_prep3 USING l_curr_code_revision,
                                  l_curr_identity,g_customer,
                                  g_date,g_dept,g_dept,g_user,g_user
         FREE dzba_prep3
         #End:170919-00003
      END IF

      LET ls_trigger = "sadzp060_2_create_code_structure : check dzbc_t current ver data count"
      IF p_prev_dzaf.dzaf010='s' AND p_curr_dzaf.dzaf010='c' THEN
         #G/X類在標準轉客製的時候,要將客製flag全部變成c.
         IF ls_type="G" OR ls_type="X" OR ls_type="K" THEN #230216-00010#1:增加'K'類.
            #Begin:200215-00009
            LET ls_sql = "insert into dzbb_t(dzbb001,dzbb002,dzbb003,dzbb004,dzbb005,dzbb006,dzbb007,dzbb098,dzbbstus) ",
                        #"SELECT dzbb001,dzbb002,dzbb003,'c',dzbb005,dzbb006,dzbb007,dzbb098,dzbbstus FROM dzbb_t", #181004-00016 remark
                         "select dzbb001,dzbb002,1,'c',dzbb005,dzbb006,dzbb007,dzbb098,dzbbstus from dzbb_t",       #181004-00016 使用版次(dzbb003)改固定為1
                         " inner join dzba_t on dzba001=dzbb001 and dzba003=dzbb002 and dzba004=dzbb003 and dzba005=dzbb004",
                         " where dzba001='",p_prog,"' and dzba002=",l_prev_code_revision," and dzba010='s' order by dzbb002"
            #End:200215-00009
            PREPARE dzbb_prep4 FROM ls_sql
            EXECUTE dzbb_prep4
            FREE dzbb_prep4
            
           #LET ls_sql = "UPDATE dzba_t SET dzba005='c'",           #181004-00016 remark
            #Begin:200215-00009
            LET ls_sql = "update dzba_t set dzba005='c',dzba004=1", #181004-00016 add:使用版次(dzba004)固定為1
                         " where dzba001='",p_prog,"'",
                         "   and dzba002=",l_curr_code_revision,
                         "   and dzba010='c'"
            #End:200215-00009
            PREPARE dzba_prep4 FROM ls_sql
            EXECUTE dzba_prep4
            FREE dzba_prep4
         END IF

         #標準轉客製的時候,要先將客製資料刪除,以免有殘留資料而造成資料不齊全.
         #Begin:200215-00009
         LET ls_sql = "delete from dzbc_t",
                      " where dzbc001='",p_prog,"'",
                      "   and dzbc007='c'"
         #End:200215-00009
         EXECUTE IMMEDIATE ls_sql
      END IF

      #判斷section樹(dzbc_t)是否存在目前正要下載的版本 : 沒有就要建立.
      #Begin:200215-00009
      LET ls_sql = "select count(*)",
                    " from dzbc_t",
                   " where dzbc001='",p_prog,"' and dzbc002='",l_curr_code_revision,"' and dzbc007='",l_curr_identity,"'" #這裡不能加上dzbcstus='Y'的條件,要不然底下就有可能出現重複插入資料的問題.
      #End:200215-00009
      PREPARE dzbc_prep1 FROM ls_sql
      EXECUTE dzbc_prep1 INTO li_cnt
      FREE dzbc_prep1

      IF li_cnt=0 THEN 
         IF ls_type MATCHES "[MSQ]" THEN
            IF l_prev_identity=l_curr_identity THEN #相同客製標示才多這個判斷.
               IF l_prev_code_revision>0 THEN
                  #Begin:200215-00009
                  LET ls_sql = "select count(*)",
                                " from dzbc_t",
                               " where dzbc001='",p_prog,"' and dzbc002='",l_prev_code_revision,"' and dzbc007='",l_prev_identity,"'" #這裡不能加上dzbcstus='Y'的條件,要不然底下就有可能出現重複插入資料的問題.
                  #End:200215-00009
                  PREPARE dzbc_prep9 FROM ls_sql
                  EXECUTE dzbc_prep9 INTO li_cnt
                  FREE dzbc_prep9
                  
                  IF li_cnt=0 THEN
                     LET ls_replace_arg = l_prev_code_revision
                     LET ls_replace_arg = ls_replace_arg.trim(),"|SECTION"
                     LET ls_err_msg = cl_replace_err_msg(cl_getmsg("adz-00700", g_lang), ls_replace_arg) #找不到舊版(第%1版)的%2設計資料, 請查明問題後再繼續執行. 
                     DISPLAY "ERROR : ",ls_err_msg
                     RETURN ls_err_msg
                  END IF
               END IF
            END IF
            
            LET li_cnt = 0
         END IF

         #要以舊的版本當作基底建立規格樹.
         LET ls_trigger = "sadzp060_2_create_code_structure : insert new ver dzbc_t"
         #Begin:200215-00009
         LET ls_sql = "insert into dzbc_t(dzbc001,dzbc002,dzbc003,dzbc004,dzbc005,",
                                         "dzbc006,dzbc007,dzbc008,dzbc009,dzbc010,dzbc011,", 
                                         "dzbccrtdt,dzbccrtdp,dzbcowndp,dzbcownid,dzbcstus,dzbccrtid)",
         #Begin:170919-00003
                                 # " SELECT dzbc001,'",l_curr_code_revision,"',dzbc003,dzbc004,dzbc005,",
                                 #         "dzbc006,'",l_curr_identity,"','",g_customer,"','',dzbc010,dzbc011,", 
                                 #         "?,'",gs_dept CLIPPED,"','",gs_dept CLIPPED,"','",g_user,"',dzbcstus,'",g_user,"'",
                                 " select dzbc001,?,dzbc003,dzbc004,dzbc005,",
                                         "dzbc006,?,?,'',dzbc010,dzbc011,", 
                                         "?,?,?,?,dzbcstus,?",
         #End:170919-00003
                                   " from dzbc_t",
                                  " where dzbc001='",p_prog,"'",
                                    " and dzbc002='",p_prev_dzaf.dzaf004 CLIPPED,"'",
                                    " and dzbc007='",p_prev_dzaf.dzaf010 CLIPPED,"'",
                                    " and dzbcstus='Y'"
         #End:200215-00009

         PREPARE dzbc_prep3 FROM ls_sql
         #Begin:170919-00003
         EXECUTE dzbc_prep3 USING l_curr_code_revision,
                                  l_curr_identity,g_customer,
                                  g_date,g_dept,g_dept,g_user,g_user
         FREE dzbc_prep3
         #End:170919-00003
      END IF 

      RETURN NULL
   CATCH 
      CALL sadzp060_2_err_catch(ls_trigger, ls_sql) 
      RETURN "ERROR : sadzp060_2_create_code_structure faliure!"
   END TRY
END FUNCTION

{ #作廢.
##########################################################################
# Access Modifier : PRIVATE
# Descriptions    : 建立元件規格結構樹.
# Input parameter : p_prog 程式代號
#                 : p_curr_dzaf 目前版次資訊
#                 : p_prev_dzaf 前版版次資訊
# Return code     : BOOLEAN(執行成功:TRUE,執行失敗:FALSE)
# Date & Author   : 2014/05/23 by Hiko
##########################################################################
PRIVATE FUNCTION sadzp060_2_create_comp_spec_structure(p_prog, p_curr_dzaf, p_prev_dzaf)
   DEFINE p_prog      STRING,
          p_curr_dzaf T_DZAF_T,
          p_prev_dzaf T_DZAF_T
   DEFINE ls_sql               STRING,
          ls_trigger           STRING,
          li_cnt               SMALLINT,
          ls_type              STRING,
          l_curr_spec_revision LIKE dzaf_t.dzaf003,
          l_prev_spec_revision LIKE dzaf_t.dzaf003,
          ls_curr_identity     STRING,
          lb_result            BOOLEAN,
          ls_err_msg           STRING

   TRY
      LET p_prog = p_prog.trim()
      LET ls_type = p_curr_dzaf.dzaf005 CLIPPED
      LET l_curr_spec_revision = p_curr_dzaf.dzaf003 CLIPPED
      LET l_prev_spec_revision = p_prev_dzaf.dzaf003 CLIPPED
      LET ls_curr_identity = p_curr_dzaf.dzaf010 CLIPPED

      IF cl_null(l_prev_spec_revision) OR l_prev_spec_revision=0 THEN
         RETURN NULL
      END IF

      LET ls_trigger = "sadzp060_2_create_comp_spec_structure : check dzdp_t current ver data count"
      #判斷元件規格樹(dzdp_t)是否存在目前正要下載的版本 : 沒有就要建立.
      #Begin:200215-00009
      LET ls_sql = "select count(*)",
                    " from dzdp_t",
                   " where dzdp001='",p_prog,"' and dzdp002='",l_curr_spec_revision,"' and dzdp006='",ls_curr_identity,"'" #這裡不能加上dzdpstus='Y'的條件,要不然底下就有可能出現重複插入資料的問題.
      #End:200215-00009
      PREPARE dzdp_prep1 FROM ls_sql
      EXECUTE dzdp_prep1 INTO li_cnt
      FREE dzdp_prep1
      #資料存在先砍舊資料,以免有殘留資料而造成資料不齊全.
      IF li_cnt=0 THEN 
         #要以舊的版本當作基底建立規格樹.
         LET ls_trigger = "sadzp060_2_create_spec_structure : insert new ver dzdp_t"
         #Begin:200215-00009
         LET ls_sql = "insert into dzdp_t(dzdp001,dzdp002,dzdp003,dzdp004,dzdp005,",
                                         "dzdp006,dzdp007,dzdp008,",
                                         "dzdpcrtdt,dzdpcrtdp,dzdpowndp,dzdpownid,dzdpstus,dzdpcrtid)",
         #Begin:170919-00003
                                 # " SELECT dzdp001,'",l_curr_spec_revision,"',dzdp003,dzdp004,dzdp005,",
                                 #         "'",ls_curr_identity,"','",g_customer,"',dzdp008,", 
                                 #         "?,'",gs_dept CLIPPED,"','",gs_dept CLIPPED,"','",g_user,"',dzdpstus,'",g_user,"'",
                                 " select dzdp001,?,dzdp003,dzdp004,dzdp005,",
                                         "?,?,dzdp008,", 
                                         "?,?,?,?,dzdpstus,?",
         #End:170919-00003
                                   " from dzdp_t",
                                  " where dzdp001='",p_prog,"'",
                                    " and dzdp002='",p_prev_dzaf.dzaf003 CLIPPED,"'",
                                    " and dzdp006='",p_prev_dzaf.dzaf010 CLIPPED,"'",
                                    " and dzdpstus='Y'"
         #End:200215-00009
         PREPARE dzdp_prep3 FROM ls_sql
         #Begin:170919-00003
         EXECUTE dzdp_prep3 USING l_curr_spec_revision,
                                  ls_curr_identity,g_customer,
                                  g_date,g_dept,g_dept,g_user,g_user
         #End:170919-00003
         FREE dzdp_prep3
      END IF

      RETURN NULL
   CATCH 
      CALL sadzp060_2_err_catch(ls_trigger, ls_sql) 
      RETURN "ERROR : sadzp060_2_create_comp_spec_structure faliure!"
   END TRY
END FUNCTION
} #FUNCTION作廢.

##########################################################################
# Access Modifier : PUBLIC 
# Descriptions    : 取得當下規格的版次和識別標示
# Input parameter : p_prog 程式代號
#                 : p_type 建構類型(M/S/F/B/G/X/...)
#                 : p_module 模組別
# Return code     : STRING 版次
#                 : STRING 識別標示
#                 : STRING 錯誤訊息(NULL表示成功)
# Date & Author   : 2014/05/23 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_2_get_spec_curr_revision(p_prog, p_type, p_module)
   DEFINE p_prog   STRING,
          p_type   STRING,
          p_module STRING
   DEFINE lo_DZAF_T  T_DZAF_T,
          l_revision LIKE dzaf_t.dzaf003,
          l_identity LIKE dzaf_t.dzaf010,
          ls_err_msg STRING

   CALL sadzp060_2_get_curr_ver_info(p_prog, p_type, p_module) RETURNING lo_DZAF_T.*,ls_err_msg
   IF cl_null(ls_err_msg) THEN
      LET l_revision = lo_DZAF_T.dzaf003 CLIPPED
      IF cl_null(l_revision) OR l_revision=0 THEN 
         LET ls_err_msg = "ERROR : ",p_prog," spec revision is NULL or 0"
         RETURN l_revision,lo_DZAF_T.dzaf010 CLIPPED,ls_err_msg
      END IF
   END IF

   RETURN l_revision,lo_DZAF_T.dzaf010 CLIPPED,ls_err_msg
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC 
# Descriptions    : 取得當下代碼的版次和識別標示
# Input parameter : p_prog 程式代號
#                 : p_type 建構類型(M/S/F/B/G/X/...)
#                 : p_module 模組別
# Return code     : STRING 版次
#                 : STRING 識別標示
#                 : STRING 錯誤訊息(NULL表示成功)
# Date & Author   : 2014/05/23 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_2_get_code_curr_revision(p_prog, p_type, p_module)
   DEFINE p_prog   STRING,
          p_type   STRING,
          p_module STRING
   DEFINE lo_DZAF_T  T_DZAF_T,
          l_revision LIKE dzaf_t.dzaf004,
          l_identity LIKE dzaf_t.dzaf010,
          ls_err_msg STRING

   CALL sadzp060_2_get_curr_ver_info(p_prog, p_type, p_module) RETURNING lo_DZAF_T.*,ls_err_msg
   IF cl_null(ls_err_msg) THEN
      LET l_revision = lo_DZAF_T.dzaf004 CLIPPED
      IF cl_null(l_revision) OR l_revision=0 THEN 
         LET ls_err_msg = "ERROR : ",p_prog," code revision is NULL or 0"
         RETURN l_revision,lo_DZAF_T.dzaf010 CLIPPED,ls_err_msg
      END IF
   END IF

   RETURN l_revision,lo_DZAF_T.dzaf010 CLIPPED,ls_err_msg
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC 
# Descriptions    : 取得當下程式(包含規格和代碼)的版次相關資訊
# Input parameter : p_prog 程式代號
#                 : p_type 建構類型(M/S/F/B/G/X/...)
#                 : p_module 模組別
# Return code     : T_DZAF_T 版次資訊物件
#                 : STRING 錯誤訊息(NULL表示成功)
# Date & Author   : 2014/05/22 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_2_get_curr_ver_info(p_prog, p_type, p_module)
   DEFINE p_prog   STRING,
          p_type   STRING,
          p_module STRING
   DEFINE lo_DZAF_T     T_DZAF_T,
          lo_DZAF_T_new T_DZAF_T,
          ls_err_msg    STRING

   LET lo_DZAF_T.DZAF001 = p_prog.trim()
   LET lo_DZAF_T.DZAF005 = p_type.trim()
   LET lo_DZAF_T.DZAF006 = p_module.trim()
   LET lo_DZAF_T.DZAF002 = g_dzaf002 
   LET lo_DZAF_T.DZAF010 = "c" #先找是否存在客製版次資料
   CALL sadzp200_ver_get_curr_ver_info(lo_DZAF_T.*) RETURNING lo_DZAF_T_new.*
   
   IF cl_null(lo_DZAF_T_new.DZAF002) OR lo_DZAF_T_new.DZAF002=0 THEN
      LET lo_DZAF_T.DZAF010 = "s" #沒有客製再找標準版次資料
      CALL sadzp200_ver_get_curr_ver_info(lo_DZAF_T.*) RETURNING lo_DZAF_T_new.*
      IF cl_null(lo_DZAF_T_new.DZAF002) OR lo_DZAF_T_new.DZAF002=0 THEN
         LET ls_err_msg = "WARNING : ",p_prog," revision info is NULL" #160516-00005
         RETURN lo_DZAF_T_new.*,ls_err_msg
      END IF
   END IF

   RETURN lo_DZAF_T_new.*,NULL
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC 
# Descriptions    : 取得當下程式(包含規格和代碼)的版次相關資訊
# Input parameter : p_prog 程式代號
#                 : p_type 建構類型(M/S/F/B/G/X/...)
#                 : p_module 模組別
#                 : p_dzaf002 建構版次
# Return code     : T_DZAF_T 版次資訊物件
#                 : STRING 錯誤訊息(NULL表示成功)
# Date & Author   : 2014/06/23 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_2_get_curr_ver_info_by_dzaf002(p_prog, p_type, p_module, p_dzaf002)
   DEFINE p_prog    STRING,
          p_dzaf002 LIKE dzaf_t.dzaf002,
          p_type    STRING,
          p_module  STRING
   DEFINE lo_DZAF_T  T_DZAF_T,
          ls_err_msg STRING

   LET g_dzaf002 = p_dzaf002 #初始建構版次

   CALL sadzp060_2_get_curr_ver_info(p_prog, p_type, p_module) RETURNING lo_DZAF_T.*,ls_err_msg

   LET g_dzaf002 = NULL #初始還原
   
   RETURN lo_DZAF_T.*,ls_err_msg
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC
# Descriptions    : 取得程式類型
# Input parameter : p_prog 程式代號
# Return code     : STRING type
# Date & Author   : 2014/08/26 by madey
##########################################################################
#+ 取得程式類型
PUBLIC FUNCTION sadzp060_2_chk_spec_type(p_specid)
   DEFINE p_specid STRING 
   DEFINE ls_spectype LIKE type_t.chr1,
          lc_gzde001  LIKE gzde_t.gzde001,
          lc_gzde003  LIKE gzde_t.gzde003,
          li_cnt      LIKE type_t.num5,
          ls_tmp      STRING,
          lc_dzca001  LIKE dzca_t.dzca001,
          lc_gzdf002  LIKE gzdf_t.gzdf002

   #注意:修改時請確認以下function規則是否需要一併調整:sadzp060_2_chk_spec_type()與cl_adz_chk_spec_type()與cl_chk_spec_type() #160706-00001

   LET p_specid = p_specid.trim()

   #確認傳入是否為子畫面
   #Begin: 160706-00001 modify
   # IF p_specid.subString(p_specid.getLength()-3,p_specid.getLength()-2) = "_s" THEN
   #    #注意最後兩碼一定是數字
   #    LET ls_tmp = p_specid.subString(p_specid.getLength()-1,p_specid.getLength())
   #    IF cl_chk_num(ls_tmp,"N") THEN
   #       #子畫面
   #       LET ls_spectype = "F"
   #       RETURN ls_spectype
   #    END IF
   # END IF
   LET lc_gzdf002 = p_specid
   SELECT COUNT(*) INTO li_cnt
     FROM gzdf_t
    WHERE gzdf002 = lc_gzdf002
   IF li_cnt > 0 THEN
      LET ls_spectype = "F"
      RETURN ls_spectype
   END IF
   #End: 160706-00001 modify

   LET lc_gzde001 = p_specid

   #確認傳入是否為子程式,或應用元件
   SELECT COUNT(*) INTO li_cnt
     FROM gzde_t
    WHERE gzde001 = lc_gzde001
   CASE
      WHEN li_cnt = 1
         #子程式/應用元件/LIB
         SELECT gzde003 INTO lc_gzde003 FROM gzde_t
          WHERE gzde001 = lc_gzde001
         RETURN lc_gzde003
      WHEN li_cnt > 1
         CASE
            WHEN p_specid.subString(1,2) = "s_"
               LET ls_spectype = "S"
            WHEN p_specid.subString(1,3) = "cl_"
               LET ls_spectype = "L"
            OTHERWISE
               LET ls_spectype = "B"  #應用元件
         END CASE
   END CASE

   #確認傳入是否為web service
   IF p_specid.subString(1,3) = "wss" OR p_specid.subString(1,4) = "cwss" THEN 
      LET lc_gzde001 = p_specid
      SELECT gzja001 INTO lc_gzde003 FROM gzja_t
       WHERE gzja001 = lc_gzde001
      IF NOT SQLCA.SQLCODE THEN
         LET ls_spectype = "Z"  #此處與cl_chk_spec_type不同!!!!!
         RETURN ls_spectype
      END IF
   END IF

   #確認傳入是否為開窗
   IF p_specid MATCHES "q_*" OR p_specid MATCHES "cq_*"  THEN
      LET li_cnt =0
      LET lc_dzca001 = p_specid
      SELECT COUNT(1) INTO li_cnt FROM dzca_t 
         WHERE dzca001=lc_dzca001
      IF li_cnt >0 THEN 
         LET ls_spectype = "Q"
         RETURN ls_spectype
      END IF
   END IF

   #最後查看是否為主程式
   SELECT COUNT(*) INTO li_cnt
     FROM gzza_t
    WHERE gzza001 = lc_gzde001
   IF li_cnt > 0 THEN
      LET ls_spectype = "M"   #主程式
   ELSE
      LET ls_spectype = "N"
   END IF

   RETURN ls_spectype
END FUNCTION

##########################################################################
# Access Modifier : PRIVATE 
# Descriptions    : 依據簽出/取消簽出處理dzay_t
# Input parameter : p_prog 程式代號
#                 : p_identity 客製標示
#                 : p_action 執行時機(1:簽出;2:取消簽出)
# Return code     : void
# Date & Author   : 2014/12/03 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_2_dzay_t(p_prog, p_identity, p_action)
   DEFINE p_prog     LIKE dzax_t.dzax001,
          p_identity LIKE dzax_t.dzax006,
          p_action   STRING 
   DEFINE ls_trigger  STRING,
          li_cnt      SMALLINT,
          lr_dzay     RECORD LIKE dzay_t.*,
          li_dzax_cnt SMALLINT 

   TRY
      SELECT COUNT(*) INTO li_cnt FROM dzay_t WHERE dzay001=p_prog
      CASE p_action
         WHEN "1" #簽出:將dzax_t的資料複製到dzay_t
            IF li_cnt>0 THEN
               #先刪除dzay_t.
               LET ls_trigger = "DELETE dzay_t:",p_prog
               DISPLAY ls_trigger
               DELETE FROM dzay_t WHERE dzay001=p_prog
            END IF

            #再複製.
            LET ls_trigger = "INSERT dzay_t:",p_prog
            DISPLAY ls_trigger
            SELECT COUNT(*) INTO li_dzax_cnt FROM dzax_t
             WHERE dzax001=p_prog AND dzax006=p_identity 
            IF li_dzax_cnt>0 THEN
               INSERT INTO dzay_t(dzay001,dzay002,dzay003,dzay004,dzay005)
                           SELECT dzax001,dzax002,dzax003,dzax007,'Y' FROM dzax_t #dzay005預設為'Y'
                            WHERE dzax001=p_prog AND dzax006=p_identity
            ELSE #找不到dzax_t就自己新增.
               CALL sadzp060_2_set_regen(p_prog, "Y") 
            END IF
         WHEN "2" #取消簽出:將dzay_t的資料複製到dzax_t
            IF li_cnt>0 THEN
               SELECT * INTO lr_dzay.* FROM dzay_t WHERE dzay001=p_prog
               LET ls_trigger = "UPDATE dzax_t:",p_prog
               DISPLAY ls_trigger
               UPDATE dzax_t SET dzax002=lr_dzay.dzay002,dzax003=lr_dzay.dzay003,dzax007=lr_dzay.dzay004
                WHERE dzax001=p_prog AND dzax006=p_identity
            END IF
      END CASE
   CATCH
      CALL sadzp060_2_err_catch(ls_trigger, NULL)
      #就算失敗也不影響簽出/取消簽出流程
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC 
# Descriptions    : 設定程式是否需要重產
# Input parameter : p_table 表格代號
# Return code     : void
# Date & Author   : 2015/11/09 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_2_need_regen(p_table)
   DEFINE p_table LIKE dzea_t.dzea001
   DEFINE ls_sql  STRING,
          li_idx  SMALLINT,
          la_prog DYNAMIC ARRAY OF RECORD
                  progid LIKE dzag_t.dzag001
                  END RECORD,
          ls_prog STRING

   DISPLAY "[INFO]sadzp060_2_need_regen : ",p_table CLIPPED
   LET ls_sql = "select distinct(dzag001) from dzag_t where dzag002='",p_table CLIPPED,"' order by dzag001"    #200215-00009
   PREPARE dzag_prep FROM ls_sql
   DECLARE dzag_curs CURSOR FOR dzag_prep
   LET li_idx = 1
   FOREACH dzag_curs INTO la_prog[li_idx].*
      CALL sadzp060_2_set_regen(la_prog[li_idx].progid, 'Y')
      IF li_idx=1 THEN
         LET ls_prog = la_prog[li_idx].progid
      ELSE
         LET ls_prog = ls_prog,", ",la_prog[li_idx].progid #逗號後面加上一個空白
      END IF
      
      LET li_idx = li_idx + 1
   END FOREACH

   IF ls_prog.getLength()>0 THEN
      DISPLAY "NEEDREGEN:",ls_prog
   END IF
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC 
# Descriptions    : 設定程式是否需要重產
# Input parameter : p_prog 程式代號
#                 : p_regen 程式是否需要重產(Y:要重產;N:不重產)
# Return code     : void
# Date & Author   : 2015/11/09 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_2_set_regen(p_prog, p_regen)
   DEFINE p_prog  LIKE dzay_t.dzay001,
          p_regen LIKE dzay_t.dzay005
   DEFINE ls_trigger STRING,
          li_cnt     SMALLINT

   #需要重產時機:
   #1.規格上傳(M/S/Q) : 有tab
   #2.r.t : 表格PK/FK改變
   #3.adzi150 : 狀態碼改變
   
   #當程式(M/S/Q)有權限下載後, 要將重產時機的標示變成N.

   TRY
      DISPLAY "[INFO]sadzp060_2_set_regen : ",p_prog," ",p_regen
      SELECT COUNT(*) INTO li_cnt FROM dzay_t WHERE dzay001=p_prog
      IF li_cnt>0 THEN #有可能未簽出過, dzay_t會沒有資料, 所以要先防呆.
         UPDATE dzay_t SET dzay005=p_regen
          WHERE dzay001=p_prog
      ELSE
         INSERT INTO dzay_t(dzay001,dzay002,dzay003,dzay004,dzay005)
                     VALUES(p_prog,'F','N',1,p_regen) #只有dzay005才重要, 其他都是預設值. 重新簽出規格後, 這些資料都會是正確的.
      END IF
   CATCH
      CALL sadzp060_2_err_catch(ls_trigger, NULL)
      #就算失敗也不影響流程
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC 
# Descriptions    : 判斷程式是否需要重產
# Input parameter : p_prog 程式代號
# Return code     : BOOLEAN Y:重產/N:不重產
# Date & Author   : 2015/11/09 by Hiko
##########################################################################
PUBLIC FUNCTION sadzp060_2_is_regen(p_prog)
   DEFINE p_prog LIKE dzay_t.dzay001
   DEFINE ls_trigger STRING,
          l_dzay005  LIKE dzay_t.dzay005,
          lb_return  BOOLEAN

   TRY
      LET ls_trigger = "[INFO]sadzp060_2_is_regen prog=",p_prog
      DISPLAY ls_trigger
      LET lb_return = FALSE
      SELECT dzay005 INTO l_dzay005 FROM dzay_t WHERE dzay001=p_prog
      IF NOT cl_null(l_dzay005) THEN
         IF l_dzay005="Y" THEN
            LET lb_return = TRUE
         END IF
      END IF

      LET ls_trigger = "[INFO]sadzp060_2_is_regen return=",lb_return
      DISPLAY ls_trigger
      RETURN lb_return
   CATCH
      CALL sadzp060_2_err_catch(ls_trigger, NULL)
      #就算失敗也不影響流程
      RETURN FALSE
   END TRY
END FUNCTION

##########################################################################
# Access Modifier : PUBLIC 
# Descriptions    : 判斷程式/規格是否處於 P65T 階段
# Input parameter : p_dzlm 簽出資訊
# Return code     : BOOLEAN Y:是/N:否
# Date & Author   : 2016/09/23 by Ernest
##########################################################################
PUBLIC FUNCTION sadzp060_2_check_if_stage_P65T(p_dzlm_t)
   DEFINE p_dzlm_t T_DZLM_T 
   DEFINE lo_dzlm_t    T_DZLM_T,
          ls_sql       STRING,
          li_rec_count INTEGER,
          lb_return    BOOLEAN  
  
  LET lo_dzlm_t.* = p_dzlm_t.*
  
  TRY 
    #取得資料筆數
    #Begin:200215-00009
    LET ls_sql = "select count(1)                          ",
                 "  from dzla_t                            ",
                 " where dzla001 = 'P65T'                  ",
                 "   and dzla003 = '",lo_dzlm_t.DZLM002,"' ",  
                 "   and dzla007 = '",lo_dzlm_t.DZLM012,"' ", 
                 "   and dzla010 =  ",lo_dzlm_t.DZLM015,"  "  
    #End:200215-00009

    PREPARE lpre_check_if_stage_P65T FROM ls_sql
    DECLARE lcur_check_if_stage_P65T CURSOR FOR lpre_check_if_stage_P65T
    OPEN lcur_check_if_stage_P65T
    FETCH lcur_check_if_stage_P65T INTO li_rec_count
    CLOSE lcur_check_if_stage_P65T
    FREE lcur_check_if_stage_P65T
    FREE lpre_check_if_stage_P65T
  CATCH 
    LET li_rec_count = 0
  END TRY   

  IF li_rec_count = 0 THEN
    LET lb_return = FALSE
  ELSE
    LET lb_return = TRUE
  END IF
  
  RETURN lb_return
END FUNCTION

################################################################################
# Descriptions...: XR報表複製xdd.
# Memo...........: XR需要額外複製xdd.而xmd只需要在還原的時候重產即可.
# Usage..........: CALL sadzp060_2_xr_backup_xdd(p_prog, p_from_xdd_dir, p_to_xdd_dir)
#                  RETURNING lb_result,ls_errmsg
# Input parameter: p_prog         XR報表元件編號
#                : p_from_xdd_dir xdd來源目錄
#                : p_to_xdd_dir   xdd目標目錄
# Return code....: BOOLEAN TRUE:執行成功/FALSE:執行失敗
#                : STRING  錯誤訊息
# Date & Author..: 2023/03/27 by 02632
# Modify.........: 230216-00010#1
################################################################################
PUBLIC FUNCTION sadzp060_2_xr_copy_xdd(p_prog, p_from_xdd_dir, p_to_xdd_dir)
   DEFINE p_prog         STRING,
          p_from_xdd_dir STRING,
          p_to_xdd_dir   STRING
   DEFINE lb_result BOOLEAN,
          ls_errmsg STRING 
   DEFINE ls_xdd_name      STRING,
          ls_from_xdd_file STRING,
          ls_to_xdd_file   STRING

   LET lb_result = TRUE
   LET ls_errmsg = NULL

   LET p_prog = p_prog.trim()

   LET ls_xdd_name = p_prog,".xdd"
   LET ls_from_xdd_file = os.path.JOIN(p_from_xdd_dir, ls_xdd_name)
   LET ls_to_xdd_file = os.path.JOIN(p_to_xdd_dir, ls_xdd_name)
   IF os.Path.exists(ls_from_xdd_file) THEN
      CALL os.Path.copy(ls_from_xdd_file, ls_to_xdd_file) RETURNING lb_result
      IF lb_result THEN 
         #最後變成777, 這樣在運行過程中會比較保險.
         IF NOT os.Path.chrwx(ls_to_xdd_file, 511) THEN                                                                                                               
            #檔案權限轉換失敗也沒關係.
            LET ls_errmsg = "[DEBUG]sadzp060_2_xr_move_xdd:change ",ls_to_xdd_file," rwx failure.It doesn't matter."
            DISPLAY ls_errmsg
         END IF
      ELSE
         #失敗就算了.
         LET ls_errmsg = "[WARNING]sadzp060_2_xr_move_xdd:copy xdd to ",ls_to_xdd_file," failure."
         DISPLAY ls_errmsg
      END IF
   END IF

   RETURN lb_result,ls_errmsg
END FUNCTION

################################################################################
# Descriptions...: XR報表備份xdd.
# Memo...........: XR需要額外備份xdd.xmd只需要在還原的時候重產即可.
# Usage..........: CALL sadzp060_2_xr_backup_xdd(p_prog, p_prev_code_ver, p_prev_prog_flag, p_from_xdd_dir, p_to_xdd_back_dir)
#                  RETURNING lb_result,ls_errmsg
# Input parameter: p_prog            XR報表元件編號
#                : p_prev_code_ver   前一版程式版次
#                : p_prev_prog_flag  前一版客製標示
#                : p_from_xdd_dir    xdd來源目錄
#                : p_to_xdd_back_dir back/xdd目錄
# Return code....: BOOLEAN TRUE:執行成功/FALSE:執行失敗
#                : STRING  錯誤訊息
# Date & Author..: 2023/03/24 by 02632
# Modify.........: 230216-00010#1
################################################################################
#PUBLIC FUNCTION sadzp060_2_xr_backup_xdd(p_prog, p_prev_code_ver, p_prev_prog_flag, p_prev_xrp_dir, p_curr_xrp_back_dir)
#   DEFINE p_prog              STRING,
#          p_prev_code_ver     STRING, #原本是SMALLINT,為了簡化判斷,則改宣告為STRING.
#          p_prev_prog_flag    STRING,
#          p_prev_xrp_dir      STRING,
#          p_curr_xrp_back_dir STRING
PUBLIC FUNCTION sadzp060_2_xr_backup_xdd(p_prog, p_prev_code_ver, p_prev_prog_flag, p_from_xdd_dir, p_to_xdd_back_dir)
   DEFINE p_prog            STRING,
          p_prev_code_ver   STRING, #原本是SMALLINT,為了簡化判斷,則改宣告為STRING.
          p_prev_prog_flag  STRING,
          p_from_xdd_dir    STRING,
          p_to_xdd_back_dir STRING
   DEFINE lb_result BOOLEAN,
          ls_errmsg STRING 
  #DEFINE ls_from_xdd_dir    STRING,
  #       ls_to_xdd_back_dir STRING,
   DEFINE ls_xdd_name      STRING,
          ls_from_xdd_file STRING,
          ls_to_xdd_file   STRING,
          ls_src_file      STRING,
          ls_dst_file      STRING

   LET lb_result = TRUE
   LET ls_errmsg = NULL

   LET p_prog = p_prog.trim()
   LET p_prev_code_ver = p_prev_code_ver.trim()
   LET p_prev_prog_flag = p_prev_prog_flag.trim()

   #NOTE:xmd只要重產即可.

   #1.先備份xdd.
  #LET ls_from_xdd_dir = os.path.JOIN(p_prev_xrp_dir, "xdd")
  #LET ls_to_xdd_back_dir = os.path.JOIN(p_curr_xrp_back_dir, "xdd")

  #LET ls_xdd_name = p_prog,".xdd"
  #LET ls_from_xdd_file = os.path.JOIN(ls_from_xdd_dir, ls_xdd_name)
  #LET ls_to_xdd_file = os.path.JOIN(ls_to_xdd_back_dir, ls_xdd_name)
  #IF os.Path.exists(ls_from_xdd_file) THEN #這邊持續覆蓋無所謂,sadzp060_2_xr_rename_backup_xrp()有保留正確的備份檔.
  #   CALL os.Path.copy(ls_from_xdd_file, ls_to_xdd_file) RETURNING lb_result
  #   IF lb_result THEN 
  #      #最後變成777, 這樣在運行過程中會比較保險.
  #      IF NOT os.Path.chrwx(ls_to_xdd_file, 511) THEN                                                                                                               
  #         #檔案權限轉換失敗也沒關係.
  #         LET ls_errmsg = "[DEBUG]sadzp060_2_xr_backup_xdd:change ",ls_to_xdd_file," rwx failure.It doesn't matter."
  #         DISPLAY ls_errmsg
  #      END IF
  #   ELSE
  #      #失敗就算了.
  #      LET ls_errmsg = "[WARNING]sadzp060_2_xr_backup_xdd:copy xdd to ",ls_to_xdd_file," failure."
  #      DISPLAY ls_errmsg
  #   END IF
  #END IF

   CALL sadzp060_2_xr_copy_xdd(p_prog, p_from_xdd_dir, p_to_xdd_back_dir) RETURNING lb_result,ls_errmsg

   #2.再將xdd備份檔複製出來並重新命名.
   LET ls_xdd_name = p_prog,".xdd"
   LET ls_src_file = os.path.JOIN(p_to_xdd_back_dir, ls_xdd_name)
   LET ls_dst_file = os.path.JOIN(p_to_xdd_back_dir, p_prog||".xdd_"||p_prev_prog_flag||p_prev_code_ver)
   
   IF os.Path.exists(ls_src_file) AND NOT os.Path.exists(ls_dst_file) THEN
      CALL os.Path.copy(ls_src_file, ls_dst_file) RETURNING lb_result
      IF lb_result THEN
         #最後變成777, 這樣在運行過程中會比較保險.
         IF NOT os.Path.chrwx(ls_dst_file, 511) THEN                                                                                                               
            #檔案權限轉換失敗也沒關係.
            DISPLAY "[WARNING]sadzp060_2_xr_backup_xdd:change ",ls_dst_file," rwx failure.It doesn't matter."
         END IF
      ELSE
         #失敗就算了.
         LET ls_errmsg = "[WARNING]sadzp060_2_xr_backup_xdd:copy xdd failure.src_file=",ls_src_file,
                         "                                                   dst_file=",ls_dst_file
         DISPLAY ls_errmsg
      END IF
   END IF

   RETURN lb_result,ls_errmsg
END FUNCTION

################################################################################
# Descriptions...: 取得需要進行備份的xrp清單.
# Memo...........:
# Usage..........: CALL sadzp060_2_xr_get_xrp_backup_list(p_prog, p_prev_code_ver, p_prev_prog_flag, p_curr_module, p_xrp_list)
#                  RETURNING lo_xrp_back_list
# Input parameter: p_prog              XR報表元件編號
#                : p_prev_code_ver     進版之前的程式版次
#                : p_prev_prog_flag    進版之前的客製標示
#                : p_curr_xrp_back_dir 進版之後的xrp/back目錄路徑
#                : p_xrp_list          XR樣板清單(gzgw_t)
# Return code....: BOOLEAN TRUE:有清單/FALSE:無清單
#                : DYNAMIC ARRAY OF T_TEMPLATE_LIST_R 需要進行備份的xrp清單
# Date & Author..: 2023/03/23 by 02632
# Modify.........: 230216-00010#1
################################################################################
PUBLIC FUNCTION sadzp060_2_xr_get_xrp_backup_list(p_prog, p_prev_code_ver, p_prev_prog_flag, p_curr_xrp_back_dir, p_xrp_list)
   DEFINE p_prog              STRING,
          p_prev_code_ver     STRING, #原本是SMALLINT,為了簡化判斷,則改宣告為STRING.
          p_prev_prog_flag    STRING,
          p_curr_xrp_back_dir STRING,
          p_xrp_list          DYNAMIC ARRAY OF T_TEMPLATE_LIST_R #同樣報表元件不同的樣板紀錄:雙層結構(gzgw_t)
   DEFINE lb_has_list      BOOLEAN,
          lo_xrp_back_list DYNAMIC ARRAY OF T_TEMPLATE_LIST_R
   DEFINE lo_lang_arr DYNAMIC ARRAY OF T_LANGUAGE_TYPE,
          li_i        SMALLINT, #樣板編號(gzgw002)變數
          ls_lang_dir STRING,   #back/$LANG目錄
          li_k        SMALLINT, #樣版名稱(gzgw007)變數
          ls_xrp_name STRING,   #gzgw007
          ls_dst_file STRING,   #複製目標檔名
          lo_gzgw002  util.JSONObject,
          ls_gzgw002  STRING,
          li_w2       SMALLINT,
          li_w7       SMALLINT

   LET lb_has_list = FALSE

   #備份一定是至少存在第1版的情況下才會有的動作,所以簽出第1版時,p_code_ver=0,直接跳過.
   LET p_prev_code_ver = p_prev_code_ver.trim()
   IF cl_null(p_prev_code_ver) OR p_prev_code_ver="0" THEN
      RETURN lb_has_list,lo_xrp_back_list
   END IF

   #LET ls_lang_dir = os.path.JOIN(p_curr_xrp_back_dir, "zh_TW") #檢查繁體模組下的備份檔即可當作代表.
   LET ls_lang_dir = os.path.JOIN(p_curr_xrp_back_dir, g_lang) #檢查登入語系的備份檔即可當作代表.

   #TYPE T_FRP_LIST_R RECORD
   #       fl_EXIST  VARCHAR(1),
   #       fl_NAME   VARCHAR(60),                                                                                                                        
   #       fl_DESC   VARCHAR(1024)
   #     END RECORD

   #TYPE T_TEMPLATE_LIST_R RECORD
   #       tl_COMPONENT VARCHAR(40),
   #       tl_TYPE      VARCHAR(02), -- 230216-00010#1
   #       tl_NAME      VARCHAR(40), #gzgw002
   #       tl_DESC      VARCHAR(1024),
   #       tl_CUSTOMIZE VARCHAR(02),
   #       tl_ROLE      VARCHAR(20),
   #       tl_USER      VARCHAR(20),
   #       FRP_LIST     DYNAMIC ARRAY OF T_FRP_LIST_R #gzgw007
   #     END RECORD

   LET lo_gzgw002 = util.JSONObject.create() 

   FOR li_i = 1 TO p_xrp_list.getLength() #樣板編號(gzgw002)逐個處理.
      IF p_xrp_list[li_i].tl_COMPONENT<>p_prog THEN #防呆.
         CONTINUE FOR
      END IF

      LET ls_gzgw002 = p_xrp_list[li_i].tl_NAME

      FOR li_k = 1 TO p_xrp_list[li_i].FRP_LIST.getLength() #樣版名稱(gzgw007)逐個判斷.
         LET ls_xrp_name = p_xrp_list[li_i].FRP_LIST[li_k].fl_NAME
         LET ls_dst_file = os.path.JOIN(ls_lang_dir, ls_xrp_name||".xrp_"||p_prev_prog_flag||p_prev_code_ver)
         #備份檔不存在就寫入清單內.
         IF NOT os.Path.exists(ls_dst_file) THEN
            IF NOT lb_has_list THEN
               LET lb_has_list = TRUE #只做一次即可.
            END IF

            IF NOT lo_gzgw002.has(ls_gzgw002) THEN
               LET li_w2 = li_w2 + 1
               CALL lo_gzgw002.put(ls_gzgw002, ls_gzgw002)
 
               LET li_w7 = 0 #每換一個樣板編號(gzgw002)就做初始化.
            END IF

            LET li_w7 = li_w7 + 1

            #這三個設定就可以正常執行sadzp050_zip_moving_4rp_files().
            LET lo_xrp_back_list[li_w2].tl_COMPONENT = p_xrp_list[li_i].tl_COMPONENT
            LET lo_xrp_back_list[li_w2].tl_TYPE = "K"
            LET lo_xrp_back_list[li_w2].FRP_LIST[li_w7].fl_NAME = p_xrp_list[li_i].FRP_LIST[li_k].fl_NAME
         END IF
      END FOR #li_k
   END FOR #li_i

   RETURN lb_has_list,lo_xrp_back_list
END FUNCTION

################################################################################
# Descriptions...: 重新命名xrp備份檔案的名稱.
# Memo...........:
# Usage..........: CALL sadzp060_2_xr_rename_backup_xrp(p_action, p_prog, p_prev_code_ver, p_prev_prog_flag, p_xrp_list, p_curr_back_dir)
#                  RETURNING lb_result,ls_errmsg
# Input parameter: p_action         執行的動作:O.簽出、R.還原
#                : p_prog           XR報表元件編號
#                : p_prev_code_ver  前一版程式版次
#                : p_prev_prog_flag 前一版客製標示
#                : p_xrp_list       XR樣板清單(gzgw_t)
#                : p_curr_back_dir  現在back目錄
# Return code....: BOOLEAN TRUE:執行成功/FALSE:執行失敗
#                : STRING  錯誤訊息
# Date & Author..: 2023/03/15 by 02632
# Modify.........: 230216-00010#1
################################################################################
PUBLIC FUNCTION sadzp060_2_xr_rename_backup_xrp(p_action, p_prog, p_prev_code_ver, p_prev_prog_flag, p_xrp_list, p_curr_back_dir)
   DEFINE p_action         STRING,
          p_prog           STRING,
          p_prev_code_ver  SMALLINT,
          p_prev_prog_flag STRING,
          p_xrp_list       DYNAMIC ARRAY OF T_TEMPLATE_LIST_R, #同樣報表元件不同的樣板紀錄:雙層結構(gzgw_t)
          p_curr_back_dir  STRING
   DEFINE lb_result BOOLEAN,
          ls_errmsg STRING
   DEFINE ls_code_ver STRING,
          lo_lang_arr DYNAMIC ARRAY OF T_LANGUAGE_TYPE,
          li_i        SMALLINT, #樣板編號(gzgw002)變數
          li_j        SMALLINT, #語系變數
          ls_lang_dir STRING,   #back/$LANG目錄
          li_k        SMALLINT, #樣版名稱(gzgw007)變數
          ls_xrp_name STRING,   #gzgw007
          ls_src_file STRING,   #複製來源檔名
          ls_dst_file STRING    #複製目標檔名

   #失敗就算了.

   LET lb_result = TRUE
   LET ls_errmsg = NULL

   #備份一定是至少存在第1版的情況下才會有的動作,所以簽出第1版時,p_prev_code_ver=0,直接跳過.
   IF p_prev_code_ver=0 THEN
      RETURN lb_result,ls_errmsg
   END IF

   LET ls_code_ver = p_prev_code_ver
   LET ls_code_ver = ls_code_ver.trim()

   CALL sadzp050_zip_get_lang_type_list() RETURNING lo_lang_arr
   #樣板編號(gzgw002)逐個處理.
   FOR li_i = 1 TO p_xrp_list.getLength()
      IF p_xrp_list[li_i].tl_COMPONENT<>p_prog THEN #防呆.
         CONTINUE FOR
      END IF

      #各語系的xrp逐個處理.
      FOR li_j = 1 TO lo_lang_arr.getLength()
         DISPLAY "[DEBUG]lang=",lo_lang_arr[li_j]
         LET ls_lang_dir = os.path.JOIN(p_curr_back_dir, lo_lang_arr[li_j])
         #各樣版名稱(gzgw007)逐個處理.
         FOR li_k = 1 TO p_xrp_list[li_i].FRP_LIST.getLength() 
            LET ls_xrp_name = p_xrp_list[li_i].FRP_LIST[li_k].fl_NAME
            
            #簽出與還原的動作剛好顛倒.
            IF p_action="O" THEN
               LET ls_src_file = os.path.JOIN(ls_lang_dir, ls_xrp_name||".xrp")
               LET ls_dst_file = os.path.JOIN(ls_lang_dir, ls_xrp_name||".xrp_"||p_prev_prog_flag||ls_code_ver)
            ELSE
               LET ls_src_file = os.path.JOIN(ls_lang_dir, ls_xrp_name||".xrp_"||p_prev_prog_flag||ls_code_ver)
               LET ls_dst_file = os.path.JOIN(ls_lang_dir, ls_xrp_name||".xrp")
            END IF
            
            IF os.Path.exists(ls_src_file) AND NOT os.Path.exists(ls_dst_file) THEN
               IF p_action="O" THEN
                  CALL os.Path.copy(ls_src_file, ls_dst_file) RETURNING lb_result #備份用copy.
               ELSE #todo9
                  CALL os.Path.rename(ls_src_file, ls_dst_file) RETURNING lb_result #退版用rename.
               END IF

               IF lb_result THEN
                  #最後變成777, 這樣在運行過程中會比較保險.
                  IF NOT os.Path.chrwx(ls_dst_file, 511) THEN                                                                                                               
                     #檔案權限轉換失敗也沒關係.
                     DISPLAY "[DEBUG]sadzp060_2_xr_rename_backup_xrp:change rwx failure.It doesn't matter."
                  END IF
               ELSE
                  LET ls_errmsg = "[ERROR]sadzp060_2_xr_rename_backup_xrp failure. xrp src_file=",ls_src_file,
                                  "                                             xrp dst_file=",ls_dst_file
                  DISPLAY ls_errmsg
               END IF
            END IF
         END FOR #li_k
      END FOR #li_j
   END FOR #li_i

   RETURN lb_result,ls_errmsg   
END FUNCTION
