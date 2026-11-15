#該程式已解開Section, 不再透過樣板產出!
#該程式為freestyle程式!
{<section id="azzi000.description" type="s" >}
#應用 a00 樣板自動產生(Version:3)
#+ Standard Version.....: SD版次:0039(2025-12-17 16:25:22), PR版次:0039(2026-03-27 09:37:55)
#+ Customerized Version.: SD版次:0000(1900-01-01 00:00:00), PR版次:0000(1900-01-01 00:00:00)
#+ Build......: 000841
#+ Filename...: azzi000
#+ Description: T100 系統首頁
#+ Creator....: ()
#+ Modifier...: 20992 -SD/PR- 20992
 
{</section>}
 
{<section id="azzi000.global" type="s" >}
#應用 i00 樣板自動產生(Version:12)
#add-point:填寫註解說明 name="global.memo"
#180627-00036#1 08146 補上 top20 程式段
#190110-00024#1 08146 top20程式轉移
#191106-00012#1 05598 autoupdate搜尋最新版本update file關鍵字更新
#200811-00049#1 2020/08/14 BY 10167 1.新增首頁版本查詢功能 2.更新GDC的警告語改為繁中
#230413-00013#1 2023/05/09 BY 00824 為了主程式patch後可做r.l，以link到新FUNCTION，補簽出
#230920-00006#1 2023/09/20 BY 20992 調整版次判定方式，給自助更新程式判定版次用
#230920-00001#1 2023/10/11 BY 00845 新增T100HOMEADDON判別
#230920-00006#1 2023/10/13 BY 00845 V3調整版次,合併 T100HOMEADD判斷方式
#231110-00002#1 2023/11/10 BY 20992 調整版次，針對近期首頁問題，重新打包自助更新版次用
#231220-00026#1 2023/12/21 BY 20992 調整版次，增加首頁包遺漏部分，azzi990、azzi993設定項目
#240202-00013#1 2024/02/02 By 00845 系統關閉在web執行adz作業
#240201-00035#1 2024/02/15 By 00824 V3首頁納入通知中心功能
#240130-00011#5 2024/02/19 By 20992 對應個資法功能調整登入系統前提供個資安全宣告
#240429-00001#1 2024/04/01 By 20992 首頁警語增加開關選項判斷是否顯示
#240716-00002#1 2024/07/22 By 20992 調整 LDAP 登入時log顯示資訊
#240815-00021#1 2024/08/15 By 20992 修正關閉首頁時，其他未關閉程式資料未清除問題
#241024-00011#1 2024/10/24 By 20992 首頁包重包進版用，不打包報表lib 
#241120-00021#1 2024/11/20 By 20992 調整首頁【Copyright】部分資訊改用多語言設定呈現
#250225-00008#1 2025/02/25 By 20992 調整首頁登入 GDC 語系設定非中文狀況，設定預設呈現英文語系
#250327-00001#1 2025/03/27 By 20992 首頁增加控卡登入時營運據點不能空白
#250526-00010#1 2025/05/26 By 20992 修正首頁預設登入選擇 AD 模式未生效問題
#250701-00016#1 2025/07/01 By 20992 1. 首頁登入時刪除其他同帳號登入部分增加判定，避免誤刪狀況。 2. 增加串接NaNa相關功能。 3. 首頁語系切換欄位對應切換修正
#251030-00012#1 2025/10/30 By 20992 安全機制調整對應進版用 
#251219-00008#1 2025/12/19 By 20992 調整首頁判定版次方式
#260129-00012#1 2026/01/29 by 20992 因應AI 串接調整需求，配合調整相關工具。
#end add-point
#add-point:填寫註解說明(客製用) name="global.memo_customerization"

#end add-point
#(ver:11) ---start---
#add-point:填寫註解說明(行業用) name="global.memo_industry"

#end add-point
#(ver:11) --- end ---
 
IMPORT os
IMPORT FGL lib_cl_dlg
#add-point:增加匯入項目 name="global.import"
IMPORT util   #for cl_ap_url()
IMPORT com
IMPORT security
IMPORT FGL azz_azzi000_01
IMPORT FGL azz_azzi000_03
IMPORT JAVA java.net.URLDecoder
IMPORT JAVA java.net.URLEncoder
#end add-point
 
SCHEMA ds
 
GLOBALS "../../cfg/top_global.inc"
 
#add-point:free_style模組變數(Module Variable) name="free_style.variable"
#CONSTANT ls_version = "1.00.01 build-20230920"   #調整版次判定方式，給自助更新程式判定版次用
#CONSTANT ls_version = "1.00.02 build-20231013"    #V3調整版次,合併 T100HOMEADD判斷方式
#CONSTANT ls_version = "1.00.03 build-20231110"    #231110-00002#1 針對近期首頁問題，重新打包自助更新版次用
#CONSTANT ls_version = "1.00.04 build-20231110"    #231220-00026#1 增加首頁包遺漏部分，azzi990、azzi993設定項目
#CONSTANT ls_version = "1.00.05 build-20240202"    #240202-00013#1 增加對使用T100ADZ_WEBSWITCH的說明
#CONSTANT ls_version = "1.1.01 build-20240219"    #240130-00011#5 調整登入系統前提供個資安全宣告、V3首頁納入通知中心功能
#CONSTANT ls_version = "1.1.02 build-20240502"    #240429-00001#1 首頁警語增加開關選項判斷是否顯示
#CONSTANT ls_version = "1.1.03 build-20240722"    #240716-00002#1 2024/07/22 By 20992 調整 LDAP 登入時log顯示資訊
#CONSTANT ls_version = "1.1.04 build-20240815"    #240815-00021#1 2024/08/15 By 20992 修正關閉首頁時，其他未關閉程式資料未清除問題
#CONSTANT ls_version = "1.2.01 build-20241024"    #241024-00011#1 2024/10/24 By 20992 首頁包重包進版用，不打包報表lib
#CONSTANT ls_version = "1.2.02 build-20241120"    #241120-00021#1 2024/11/20 By 20992 調整首頁【Copyright】部分資訊改用多語言設定呈現
#CONSTANT ls_version = "1.2.03 build-20250225"    #250225-00008#1 2025/02/25 By 20992 調整首頁登入 GDC 語系設定非中文狀況，設定預設呈現英文語系
#CONSTANT ls_version = "1.2.04 build-20250327"    #250327-00001#1 2025/03/27 By 20992 首頁增加控卡登入時營運據點不能空白
#CONSTANT ls_version = "1.2.05 build-20250526"    #250526-00010#1 2025/05/26 By 20992 修正首頁預設登入選擇 AD 模式未生效問題
#CONSTANT ls_version = "1.2.06 build-20250630"    #250701-00016#1 2025/07/01 By 20992 1. 首頁登入時刪除其他同帳號登入部分增加判定，避免誤刪狀況。 2. 增加串接NaNa相關功能。 3. 首頁語系切換欄位對應切換修正
#CONSTANT ls_version = "1.2.07 build-20251030"    #251030-00012#1 2025/10/30 By 20992 安全機制調整對應進版用
CONSTANT ls_version = "1.3.01 build-20260327"    #260129-00012#1  2026/01/29 by 20992 因應AI 串接調整需求，配合調整相關工具。
#end add-point
 
#add-point:自定義模組變數(Module Variable) name="global.variable"
 
#end add-point
 
#add-point:自定義客戶專用模組變數(Module Variable) name="global.variable_customerization"

#end add-point
 
{</section>}
 
{<section id="azzi000.main" type="m" >}
#+ 作業開始
MAIN
   #200811-00049#1 By 10167 -----S
   DEFINE ls_arg STRING   
   #251219-00008#1 add --- start ---
   DEFINE l_json     util.JSONObject
   DEFINE ls_param   STRING
   DEFINE l_arr      util.JSONArray
   DEFINE l_param_arr DYNAMIC ARRAY OF STRING
   #251219-00008#1 add --- end ---
   
   LET ls_arg = ARG_VAL(1)
   
   #251219-00008#1 add --- start ---
   LET l_json = util.JSONObject.create()
   LET l_json = util.JSONObject.parse(ls_arg)
   LET l_arr = l_json.get("param")
   CALL l_arr.toFGL(l_param_arr)
   #251219-00008#1 add --- end ---
   
   #IF ls_arg.getIndexOf('param":["-V"]',1) OR ls_arg.getIndexOf('param":["-v"]',1) THEN #251219-00008#1 mark
   IF ls_arg.getIndexOf('param":["-V"]',1) OR ls_arg.getIndexOf('param":["-v"]',1) OR l_param_arr[1] = "-v" OR l_param_arr[1] = "-V" THEN #251219-00008#1 add
      #CALL azzi000_chk_ver()  #230920-00006#1 mark
      DISPLAY "INFO:本區首頁版本為:",ls_version #230920-00006#1 
      EXIT PROGRAM
   END IF
   #200811-00049#1 By 10167 -----E

   #240202-00013 現行系統為求方便性, 開放在web執行adz作業
   #             若客戶欲關閉開發工具在 web的調用(只准許從設計器連動) 則請放開下面這個設定
   #CALL FGL_SETENV("T100ADZ_WEBSWITCH","off")
   #     註記: 若用戶所啟用的任何一個角色,具備 superuser(azzi850) 的屬性, 則視同系統管理人(不理會據點限制)
   #           將會開放該用戶不受管制

   #180627-00036#1 --start--
   &ifdef v3                            
   CALL azzi000_gdc_autoupdate()
   &endif 
   #180627-00036#1 -- end --
   IF TRUE THEN
      CALL ui.interface.loadActionDefaults(os.Path.join(FGL_GETENV("ERP"), "cfg/4ad/homepage.4ad"))
      #190110-00024#1 -s-
      #CALL azzi000_main()
      &ifdef v1
         CALL azzi000_main()
      &endif
      &ifndef v1
         IF FGL_GETENV("T100HOMEADDON") IS NOT NULL AND FGL_GETENV("T100HOMEADDON") = "V1" THEN
            CALL azzi000_main()     #230920-00001#1
         ELSE
            CALL azzi000_main_v3()
         END IF
      &endif
      #190110-00024#1 -e-
   ELSE
      #Fix for GBC add --- start ---
   #OPEN WINDOW w1 WITH FORM cl_ap_formpath("azz","azzi000") #Fix for GBC mark
   IF FGL_GETENV('T100NANAWEB') = 'Y' THEN
      OPEN WINDOW w1 WITH FORM cl_ap_formpath("azz","azzi000") ATTRIBUTES(STYLE='layoutwin')
   ELSE
      OPEN WINDOW w1 WITH FORM cl_ap_formpath("azz","azzi000")
   END IF
   #Fix for GBC add --- end --- CLOSE WINDOW w1
      #Fix for GBC add --- start ---
   #OPEN WINDOW w2 WITH FORM cl_ap_formpath("azz","azzi000_02") #Fix for GBC mark
   IF FGL_GETENV('T100NANAWEB') = 'Y' THEN
      OPEN WINDOW w2 WITH FORM cl_ap_formpath("azz","azzi000_02") ATTRIBUTES(STYLE='layoutwin')
   ELSE
      OPEN WINDOW w2 WITH FORM cl_ap_formpath("azz","azzi000_02")
   END IF
   #Fix for GBC add --- end --- CLOSE WINDOW w2
      #Fix for GBC add --- start ---
   #OPEN WINDOW w3 WITH FORM cl_ap_formpath("azz","azzi000_s02") #Fix for GBC mark
   IF FGL_GETENV('T100NANAWEB') = 'Y' THEN
      OPEN WINDOW w3 WITH FORM cl_ap_formpath("azz","azzi000_s02") ATTRIBUTES(STYLE='layoutwin')
   ELSE
      OPEN WINDOW w3 WITH FORM cl_ap_formpath("azz","azzi000_s02")
   END IF
   #Fix for GBC add --- end --- CLOSE WINDOW w3
      
   END IF
 
 
 
 
 
 
 

 
END MAIN
 
{</section>}
 
{<section id="azzi000.other_function" readonly="Y" type="s" >}
#add-point:自定義元件(Function) name="other.function"

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi000_gdc_autoupdate()
   DEFINE ls_path STRING
&ifdef v3
   LET ls_path = azzi000_check_for_update()
   IF ls_path IS NOT NULL THEN
      #MENU "AutoUpdate check" ATTRIBUTES(STYLE="dialog",COMMENT="There is a new GDC version available, would like to update your installation?")   #200811-00049#1 By 10167 mark
      MENU "AutoUpdate check" ATTRIBUTES(STYLE="dialog",COMMENT="現在有新的GDC版本，請問您是否要安裝更新？")   #200811-00049#1 By 10167 add
         ON ACTION accept
            CALL azzi000_do_update(ls_path)
            EXIT MENU
         ON ACTION cancel
            EXIT MENU
      END MENU
   END IF
&endif 
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi000_check_for_update()
   DEFINE ls_version STRING
   DEFINE ls_os      STRING
   DEFINE ls_target  STRING
   DEFINE ls_path    STRING
   DEFINE ls_str     STRING
&ifdef v3
   LET ls_path = NULL
   LET ls_str = FGL_GETENV("GENERODIR"),"/autoupdate"
   IF NOT os.Path.exists(ls_str) THEN
      DISPLAY "error: ",FGL_GETENV("GENERODIR"),"/autoupdate  not exists"
      RETURN ls_path
   END IF
   #Retrieve front end information
   LET ls_version = ui.interface.getFrontEndVersion()
   CALL ui.Interface.frontcall("standard", "feinfo", ["ostype"], [ls_os])
   CALL ui.Interface.frontcall("standard", "feinfo", ["target"], [ls_target])

   DISPLAY "SYSTEM TYPE:",ls_target
   #Compute the path to the update archive
   LET ls_path = azzi000_get_update_path(ls_version, ls_os, ls_target)
   RETURN ls_path
&endif
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi000_get_update_path(ls_version,ls_os,ls_target)
   DEFINE ls_version STRING
   DEFINE ls_os      STRING
   DEFINE ls_target  STRING
   DEFINE ls_path    STRING
   -- TODO --
   DEFINE clientVersion          INTEGER 
   DEFINE serverVersion          INTEGER
   DEFINE sb_tmp                 base.StringBuffer
   DEFINE ls_str                 STRING
   DEFINE index_num              INTEGER
   DEFINE result                 INTEGER
&ifdef v3
   DISPLAY "current OS Type: ",ls_os 
   LET sb_tmp = base.StringBuffer.create()
   CALL sb_tmp.clear()
   CALL sb_tmp.append(ls_version) 
   CALL sb_tmp.getIndexOf("3.20.",1) RETURNING result
   CALL sb_tmp.getIndexOf("-",result) RETURNING index_num
   
   LET serverVersion = azzi000_get_update_file_version()
   DISPLAY "Server Version:",serverVersion 
   
   LET ls_str = sb_tmp.subString(result+5,index_num-1)    
   LET clientVersion = ls_str 
   DISPLAY "Client Version:",clientVersion
   
   LET ls_path = NULL
   
   IF clientVersion IS NULL THEN
      DISPLAY "Client Version is wrong"
      RETURN ls_path 
   END IF
   IF serverVersion > clientVersion THEN
      LET ls_path = azzi000_get_update_file_name(serverVersion,ls_target.subString(1,3))
   END IF

   RETURN ls_path
&endif
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi000_do_update(ls_server_path)
   DEFINE ls_server_path STRING
   DEFINE ls_client_path STRING
   DEFINE res STRING
&ifdef v3 
   #Compute the path to the update archive on the client and prepare the client
   LET ls_client_path = azzi000_get_target_path()
   #Transfer the archive on the client
   CALL FGL_PUTFILE(ls_server_path, ls_client_path)
   #Run the update
   DISPLAY "Updating.....",ls_client_path
   CALL ui.Interface.frontCall("monitor", "update", [ls_client_path], [res])
&endif
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi000_get_target_path()
   DEFINE ls_os STRING
   DEFINE ls_client_path STRING
   DEFINE ret STRING
   
&ifdef v3
   CALL ui.Interface.frontcall("standard", "feinfo", ["ostype"], [ls_os])
   IF ls_os= "WINDOWS" THEN
       LET ls_client_path = "c:\\t100temp\\gdcupdate"
       CALL ui.Interface.frontCall("standard", "execute", ["cmd /C rd /S /Q " || ls_client_path, TRUE], [ret])
       CALL ui.Interface.frontCall("standard", "execute", ["cmd /C md " || ls_client_path, TRUE], [ret])
       RETURN "c:\\t100temp\\gdcupdate\\updateArchive.zip"
   ELSE
       LET ls_client_path = "/tmp/fourjs.tmp"
       CALL ui.Interface.frontCall("standard", "execute", ["rm -rf " || ls_client_path, TRUE], [ret])
       CALL ui.Interface.frontCall("standard", "execute", ["mkdir -p " || ls_client_path, TRUE], [ret])
       RETURN "/tmp/fourjs.tmp/updateArchive.zip"
   END IF
&endif
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL s_aooi150_ins (传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi000_get_update_file_version()
   DEFINE ls_server_ver  STRING 
   DEFINE l_cmd          STRING
   DEFINE l_channel      base.Channel
   
&ifdef v3 
   LET l_channel = base.Channel.create()
   LET l_cmd = "ls ",FGL_GETENV("GENERODIR"),"/autoupdate/*autoupdate*  | grep 3.20 | awk -F- '{print $3}' | awk -F. '{print $3}' | awk 'BEGIN {max = 0} {if ($1+0 > max+0) max=$1} END {print max}'"
   DISPLAY l_cmd
   CALL l_channel.openPipe(l_cmd,"r")
   LET ls_server_ver =l_channel.readLine()
   CALL l_channel.close()
    
   RETURN ls_server_ver
&endif
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL azzi000_get_update_file_name(ls_version,ls_type)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi000_get_update_file_name(ls_version,ls_type)
   DEFINE ls_version     STRING
   DEFINE ls_type        STRING
   DEFINE ls_file_name   STRING 
   DEFINE l_cmd          STRING
   DEFINE l_channel      base.Channel
   
&ifdef v3 
   LET l_channel = base.Channel.create()
   LET l_cmd = "ls ",FGL_GETENV("GENERODIR"),"/autoupdate/*autoupdate* | grep ",ls_version,"-build | grep ",ls_type   #191106-00012#1 modify
   DISPLAY l_cmd
   CALL l_channel.openPipe(l_cmd,"r")
   LET ls_file_name=l_channel.readLine()
   CALL l_channel.close()
    
   RETURN ls_file_name
&endif
END FUNCTION

################################################################################
# Descriptions...: 首頁版本查詢
# Memo...........:
# Usage..........: CALL azzi000_chk_ver()
# Input parameter: 無
# Return code....: 無
# Date & Author..: 2020/08/14 By 10167 200811-00049#1
# Modify.........:
################################################################################
PRIVATE FUNCTION azzi000_chk_ver()
   DEFINE ps_num  STRING #patch號
   DEFINE lh_req  com.HTTPRequest
   DEFINE lh_resp com.HTTPResponse
   DEFINE ls_body STRING
   DEFINE lr_body RECORD 
            progId STRING
      END RECORD
   DEFINE ls_resp STRING 
   DEFINE lr_resp RECORD 
          duration          STRING,
          statusDescription STRING,
          response          RECORD
             progVer        DYNAMIC ARRAY OF STRING
          END RECORD,
          uuid              STRING
          END RECORD
   DEFINE ls_local_ver      STRING #當下版本
   DEFINE ls_cloud_ver      STRING #服務雲上的最新版本
   DEFINE l_read            base.Channel
   DEFINE ls_url            STRING
   DEFINE lj_json           util.JSONObject #2019/11/26
   DEFINE ls_path         STRING
   

   #抓取服務雲上的版本
   TRY
      LET lh_req = com.HTTPRequest.Create("http://tscloud.digiwin.com/DWGateway/restful/TCPortal/IT100Service/getProgramVersion")
      CALL lh_req.setMethod("POST")
      CALL lh_req.setHeader("Content-Type","application/json")
      
     LET lr_body.progId = "MENU"

      #Begin: 2019/08/26
      #若 60 秒內無回應則放棄
      CALL fgl_ws_setOption("http_invoketimeout", 60)
      CALL lh_req.setTimeOut(60)
      CALL lh_req.setConnectionTimeOut(60)
      #End: 2019/08/26

      LET ls_body = util.JSON.stringify(lr_body)
      
      #DISPLAY 'INFO:呼叫服務雲取得工具版本資訊-',ls_body
      CALL lh_req.doTextRequest(ls_body)    
      
      LET lh_resp = lh_req.getResponse()
      IF lh_resp.getStatusCode() != 200 THEN
        #DISPLAY "HTTP Error ("||lh_resp.getStatusCode()||") ", lh_resp.getStatusDescription()
         DISPLAY "WARNING:",cl_getmsg("adz-01612", g_lang),cl_getmsg("adz-01613", g_lang),
                 "HTTP Fail ("||lh_resp.getStatusCode()||") ", lh_resp.getStatusDescription() #2019/11/26
         #Begin: 2019/11/26
         LET ls_resp = lh_resp.getTextResponse()
         LET lj_json = util.JSONObject.parse(ls_resp)
         DISPLAY 'WARNING:',lj_json.get("errorMessage")
         #End: 2019/11/26
      ELSE
         LET ls_resp = lh_resp.getTextResponse()
         #DISPLAY "HTTP Response is : ",ls_resp
         CALL util.JSON.parse(ls_resp,lr_resp)
         LET ls_cloud_ver = lr_resp.response.progVer[1]
         #DISPLAY "最新版本為:",ls_cloud_ver
      END IF
   CATCH
     #DISPLAY "ERROR :",STATUS," (",SQLCA.SQLERRM,")"
      DISPLAY "WARNING:",cl_getmsg("adz-01612", g_lang),cl_getmsg("adz-01613", g_lang),
               STATUS," (",SQLCA.SQLERRM,")" #2019/11/26
   END TRY
   
   #比對後若不同則提示
   DISPLAY ""
   RUN "cat $AZZ/4gl/menu.ver" #顯示當下版本
   DISPLAY "最新首頁版本為:",ls_cloud_ver
   #DISPLAY "最新版更新包連結:http://dw-tgroup.digiwin.biz/..."
   
   #建立連結
   #LET ls_url = "<a href = 'http://dw-tgroup.digiwin.biz/forum/forum.php?mod=viewthread&tid=494&highlight=%E6%A8%A3%E6%9D%BF'>更新連結</a>"
   #LET ls_url = "<a href = 'http://dw-tgroup.digiwin.biz/forum/forum.php?mod=viewthread&tid=626&page=1&extra='>更新連結</a>" 
   #DISPLAY ls_url TO lbl_link
   #IF ls_local_ver <> ls_cloud_ver THEN
   #   ERROR "有新版工具，建議進行更新！"
   #END IF

END FUNCTION

#end add-point
 
{</section>}
 
