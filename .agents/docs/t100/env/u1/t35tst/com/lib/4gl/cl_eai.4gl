#該程式未解開Section, 採用最新樣板產出!
#該程式為freestyle程式!
{<section id="cl_eai.description" type="s" >}
#應用 a00 樣板自動產生(Version:3)
#+ Standard Version.....: SD版次:0017(1900-01-01 00:00:00), PR版次:0017(2025-09-03 11:32:32)
#+ Customerized Version.: SD版次:0000(1900-01-01 00:00:00), PR版次:0000(1900-01-01 00:00:00)
#+ Build......: 000654
#+ Filename...: cl_eai
#+ Description: EAI 整合應用元件
#+ Creator....: 00544(2014-12-08 15:20:39)
#+ Modifier...: 00000 -SD/PR- 02013
 
{</section>}
 
{<section id="cl_eai.global" type="s" >}
#應用 p00 樣板自動產生(Version:6)
#add-point:填寫註解說明 name="main.memo"
#160527-00003#1 2016/5/27  By yc.chao       補入request xml內的param節點資料
#161228-00023  2016/12/28 By yc.chao  新增中台檔案傳輸處理
#170512-00048#2 2017/05/17 By nikoybeat     1. 新增取得中台產品資訊(getProdList)
#                                           2. 調整取得互聯中台設定，以awsi100設定為主，若不存在設定時才改取aoos010設定
#180625-00010#1 2018/06/25 By 10553         調整ERPDB與AWSDB對應關係改由參考azzi085設定作業
#180912-00001#1 2018/09/12 By 10553         調整取得中間庫編號方式,改由參考azzi085設定作業 
#200116-00011#1 2020/01/20 by circlelai, 07558  取得資料庫名稱行為因 MSSQL Server行為調整 (eq: dsaws change to topprd_dsaws.dbo);
#241025-00007#1 2024/10/25 By 01856         修正取呼叫中台取得已註冊服務後被截斷問題。ex:semc.10.100.42.11.nip.io ，取回服務被截斷semc.10.100.42.11.ni
#end add-point
#add-point:填寫註解說明(客製用) name="main.memo_customerization"

#end add-point
#(ver:6) ---start---
#add-point:填寫註解說明(行業用) name="global.memo_industry"

#end add-point
#(ver:6) --- end ---
 
IMPORT os
#add-point:增加匯入項目 name="main.import"
IMPORT xml
IMPORT com
IMPORT util
IMPORT security
#end add-point
 
SCHEMA ds
 
GLOBALS "../../cfg/top_global.inc"
#add-point:增加匯入變數檔 name="global.inc"
GLOBALS "../../cfg/top_ws.inc"
GLOBALS "../../../erp/aws/4gl/aws_eai.inc"
#end add-point
 
{</section>}
 
{<section id="cl_eai.free_style_variable" type="s" >}
#add-point:free_style模組變數(Module Variable) name="free_style.variable"

#end add-point
 
{</section>}
 
{<section id="cl_eai.global_variable" type="s" >}
#add-point:自定義模組變數(Module Variable) name="global.variable"
TYPE type_g_attribute DYNAMIC ARRAY OF RECORD
                name         STRING,      #屬性名稱
                value        STRING       #屬性值
             END RECORD
             
TYPE type_g_param DYNAMIC ARRAY OF RECORD
                name         STRING,      #屬性名稱
                value        STRING,      #屬性值
                text         STRING       #node value
             END RECORD 
             
DEFINE g_ws_Start       STRING           #啟動services時間  
DEFINE g_ws_End         STRING           #結束services時間
DEFINE g_logtime_start  DATETIME HOUR TO FRACTION(5)            

TYPE type_g_logv_t RECORD
          logv002   DATETIME YEAR TO SECOND, 
          logv003   LIKE logv_t.logv003, 
          logv004   LIKE logv_t.logv004, 
          logv005   LIKE logv_t.logv005, 
          logv006   LIKE logv_t.logv006,     
          logv007   LIKE logv_t.logv007,
          logv008   LIKE logv_t.logv008,
          logv009   LIKE logv_t.logv009,
          logv010   LIKE logv_t.logv010          
END RECORD
#end add-point
 
{</section>}
 
{<section id="cl_eai.other_dialog" type="s" >}

 
{</section>}
 
{<section id="cl_eai.other_function" readonly="Y" type="s" >}

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_eai_end_log(p_request_xml,p_response_xml)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_eai_end_log(p_request_xml,p_response_xml)
   DEFINE p_request_xml          STRING           #送出之request xml
   DEFINE p_response_xml         STRING           #取回之response xml   
   
   DEFINE l_logtime_end          DATETIME HOUR TO FRACTION(5) #結束service時間
   DEFINE l_interval             INTERVAL HOUR TO FRACTION(5) #service耗費時間   
   
   LET g_ws_End = CURRENT YEAR TO SECOND
   LET l_logtime_end = CURRENT HOUR TO FRACTION(5)
   LET l_interval = l_logtime_end - g_logtime_start
   
   #LET l_process_time = "Process Duration: ", l_interval   
   CALL cl_ws_write_log_file("eaicli", p_request_xml,p_response_xml,g_logtime_start,l_logtime_end,l_interval)
   
   
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_eai_start_log()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_eai_start_log()

   LET g_ws_Start = CURRENT YEAR TO SECOND   
   LET g_logtime_start = CURRENT HOUR TO FRACTION(5)
   
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_eai_create_job_request(p_sync,p_job_name,p_job_attr)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_eai_create_job_request(p_sync,p_job_name,p_job_attr)
   DEFINE p_job_name             STRING
   DEFINE p_sync                 STRING
   DEFINE p_job_attr             type_g_attribute
   DEFINE p_attr                 type_g_attribute   #request 屬性資料
   DEFINE l_req_doc              xml.DomDocument
   DEFINE l_request_root         xml.DomNode        #Request XML Node
   DEFINE l_tag                  STRING             #節點tag名稱
   DEFINE l_attr                 type_g_attribute   #節點屬性資料
   DEFINE l_node                 xml.DomNode
   DEFINE l_Request              STRING
   DEFINE l_xml_str              STRING
   DEFINE ls_sync                STRING
   DEFINE l_start                LIKE type_t.num10
   DEFINE l_end                  LIKE type_t.num10    
   
   INITIALIZE l_request_root TO NULL
   INITIALIZE l_node TO NULL
   
   #建立 Request XML Document
   LET l_req_doc = xml.DomDocument.createDocument("request")
   LET l_request_root = l_req_doc.getDocumentElement()       

   #IF p_attr.getLength() > 0 THEN     #設定 request 屬性
   #   CALL l_request_root.setAttribute(p_attr[1].name, p_attr[1].value)
   #END IF
   
   #建立TT主機資訊(<host>)
   LET l_tag = "host"
   CALL l_attr.clear()
   IF cl_null(p_job_name) THEN
      LET l_attr = cl_eai_get_base_info()
   ELSE 
      LET l_attr = cl_eai_get_host_info()
   END IF
  
   
   CALL cl_bpm_create_node(l_req_doc, l_request_root, l_tag, l_attr)
      RETURNING l_req_doc, l_node

   IF NOT cl_null(p_job_name) THEN
      #建立 job 資訊 <etljob>
      LET l_tag = "etljob"  
  
      CALL cl_bpm_create_node(l_req_doc, l_request_root, l_tag, p_job_attr)
         RETURNING l_req_doc, l_node
         
      #取得 host 及 service 字串 <host prod=.../>
      LET l_Request = l_request_root.toString()
      LET l_start = l_Request.getIndexOf("<host",1)
      LET l_end = l_Request.getIndexOf("/>", l_start)
      LET l_xml_str =  l_Request.subString(l_start,l_end+1)
      
      #建立屬性資料
      CALL l_request_root.setAttribute("type", p_sync)
      CALL l_request_root.setAttribute("key", cl_trust_crosskey(l_xml_str))      
   END IF
   
   RETURN l_req_doc, l_request_root
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_eai_get_base_info()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_eai_get_base_info()
   DEFINE r_attr                 type_g_attribute   #子節點的屬性資料
   DEFINE l_value                STRING
   DEFINE l_p                    LIKE type_t.num5
   DEFINE l_sb                   base.StringBuffer
   DEFINE l_timezone             STRING             #需求端時區
  
   LET r_attr[1].name = "prod"
   LET r_attr[1].value = "T100"

   LET r_attr[2].name = "ver"
   LET r_attr[2].value = FGL_GETENV("ERPVER") #cl_getenv_erp_version()

   LET r_attr[3].name = "ip"
   LET r_attr[3].value = cl_getenv_tpserver_ip()

   LET r_attr[4].name = "id"
   LET r_attr[4].value = FGL_GETENV("ZONE")          #"t10dap"

   LET r_attr[5].name = "lang"
   LET r_attr[5].value = g_lang
   
   RETURN r_attr
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
PUBLIC FUNCTION cl_eai_get_mailaddress(p_user)
   DEFINE l_oofc012    LIKE oofc_t.oofc012          #通訊內容
   DEFINE l_empid      LIKE gzxa_t.gzxa001
   DEFINE p_user       LIKE gzxa_t.gzxa003
   
   #SELECT gzxa001 INTO l_empid FROM gzxa_t
   # WHERE gzxa003 = p_user

   #參考aooi350 主要通訊
   SELECT oofc012 INTO l_oofc012  #找出 mail address
     FROM oofc_t,oofa_t
    WHERE oofa001 = oofc002
      AND oofc008 = '4'
      AND oofc010 = 'Y'
      AND oofa003 = p_user
      AND oofcent = oofaent  #r.c3 檢查規格變更需求

   RETURN l_oofc012
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
PUBLIC FUNCTION cl_eai_sendmail(p_user,p_title,p_content)
   DEFINE p_user      STRING
   DEFINE p_title     STRING 
   DEFINE p_content   STRING
   DEFINE l_mail_path STRING
   DEFINE ls_tmp      STRING
   DEFINE l_channel   base.Channel
   
   #寫檔案
   LET l_mail_path = os.Path.join(FGL_GETENV("TEMPDIR"), "cl_notify_mail.txt")
   LET l_channel = base.Channel.create()
   CALL l_channel.setDelimiter("")
   CALL l_channel.openFile(l_mail_path,"w")      
   CALL l_channel.write(p_content)
   CALL l_channel.close()
      
   LET g_xml.body = l_mail_path
   LET g_xml.subject = p_title   
   LET g_xml.recipient = p_user  #收件人

   #寄發mail
   CALL cl_jmail() RETURNING ls_tmp
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
PUBLIC FUNCTION cl_eai_format_time()
   DEFINE l_sb                    base.StringBuffer
   DEFINE l_timestamp             STRING

   LET l_timestamp = CURRENT YEAR TO FRACTION(3)
   LET l_sb = base.StringBuffer.create()
   CALL l_sb.append(l_timestamp)
   CALL l_sb.replace(":","",0)
   CALL l_sb.replace(" ","",0)
   CALL l_sb.replace("-","",0)
   CALL l_sb.replace(".","",0)
   LET l_timestamp = l_sb.toString()

   RETURN  l_timestamp   
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
PUBLIC FUNCTION cl_eai_download_file(ps_source,ps_target)
   DEFINE ps_source    STRING,
         ps_target    STRING

   CALL FGL_PUTFILE(ps_source, ps_target)
   IF STATUS THEN
      RETURN FALSE
   ELSE
      RETURN TRUE
   END IF
END FUNCTION

################################################################################
# Descriptions...: ETL 專用
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
PUBLIC FUNCTION cl_eai_get_host_info()
DEFINE r_attr                 type_g_attribute   #子節點的屬性資料
   DEFINE l_value                STRING
   DEFINE l_p                    LIKE type_t.num5
   DEFINE l_sb                   base.StringBuffer
   DEFINE l_timezone             STRING             #需求端時區
   
   #<host prod="TIPTOP" ver="1.0" ip="10.40.40.18" id="topprod" lang="zh_TW" timezone="+8" timestamp="20130925155232023" acct="tiptop" />
   LET r_attr[1].name = "prod"
   LET r_attr[1].value = "T100"

   LET r_attr[2].name = "ver"
   LET r_attr[2].value = FGL_GETENV("ERPVER") #cl_getenv_erp_version()

   LET r_attr[3].name = "ip"
   LET r_attr[3].value = cl_getenv_tpserver_ip()

   LET r_attr[4].name = "id"
   LET r_attr[4].value = FGL_GETENV("ZONE")          #"t10dap"

   LET r_attr[5].name = "lang"
   LET r_attr[5].value = g_lang

   LET l_timezone = g_timezone CLIPPED
   LET l_timezone = l_timezone.subString(4, l_timezone.getLength())
   IF cl_null(l_timezone) THEN
      LET l_timezone = "+8"
   END IF
   
   LET r_attr[6].name = "acct"
   LET r_attr[6].value = "etl"
   
   #LET r_attr[6].name = "timezone"
   #LET r_attr[6].value = l_timezone.trim()
   
   #LET l_value = CURRENT YEAR TO FRACTION(3)  
   LET l_sb = base.StringBuffer.create()
   CALL l_sb.append(l_value)
   CALL l_sb.replace(":","",0)
   CALL l_sb.replace(" ","",0)
   CALL l_sb.replace("-","",0)
   CALL l_sb.replace(".","",0)
   #LET l_value = l_sb.toString()
   #LET g_timestamp = l_value

   LET r_attr[7].name = "timestamp"
   LET r_attr[7].value = l_value.trim()          #"20131009140132023"
   
   RETURN r_attr
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
PUBLIC FUNCTION cl_eai_db_connect(ps_dbs)
   DEFINE ps_dbs      STRING
   DEFINE ps_middb    STRING
   #180625-00010#1 add by 10553  (S)
   DEFINE ls_dbs      LIKE gzda_t.gzda001
   DEFINE ls_middb    LIKE gzda_t.gzda010
   #180625-00010#1 add by 10553  (E)
   
   #180625-00010#1 mark by 10553  (S)
   #CASE ps_dbs
   #     WHEN 'dsdemo'
   #        LET ps_middb = 'dsawst'
   #     WHEN 'dsdata'
   #        LET ps_middb = 'dsaws'
   #     OTHERWISE
   #        LET ps_middb = 'dsawst'
   #   END CASE
   #180625-00010#1 mark by 10553  (E)
   
   #180625-00010#1 add by 10553  (S)
   LET ls_dbs = ps_dbs
   #從azzi085的設定資料取得ERPDB與AWSDB對應關係
   TRY
      SELECT gzda010 INTO ls_middb
        FROM gzda_t
       WHERE gzda001 = ls_dbs AND gzdastus = 'Y'
   CATCH
      INITIALIZE ls_middb TO NULL
      DISPLAY "Failed to get data of azzi085(gzda_t)."
   END TRY
   
   IF cl_null(ls_middb) THEN
      CASE ps_dbs
        WHEN 'dsdemo'
           LET ps_middb = 'dsawst'
        WHEN 'dsdata'
           LET ps_middb = 'dsaws'
        OTHERWISE
           LET ps_middb = 'dsawst'
      END CASE
   ELSE
      LET ps_middb = ls_middb
   END IF
   #180625-00010#1 add by 10553  (E)
   CALL cl_db_connect(ps_middb, FALSE)
   
   RETURN ps_middb
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_eai_invokeSrv(p_prod,p_srvname,p_type,p_Request)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_eai_invokeSrv(p_json_record)
   DEFINE p_json_record     util.JSONObject
   
   DEFINE l_Request         STRING
   DEFINE l_Result          STRING
   DEFINE l_host_att        STRING
   DEFINE l_request_att     STRING
   DEFINE l_service_attr    type_g_attribute
   
   DEFINE l_req_doc         xml.DomDocument
   DEFINE l_request_root    xml.DomNode           #Request XML Node
   DEFINE l_pnode           xml.DomNode
   DEFINE l_param_node      xml.DomNode           #add by 160527-00003
   DEFINE l_tnode           xml.DomNode
   DEFINE l_response        xml.DomDocument
   
   DEFINE l_xml_str         STRING
   DEFINE ls_prod           STRING
   DEFINE ls_srvname        STRING
   DEFINE ls_req_data       STRING
   DEFINE ls_sync_type      STRING
   DEFINE lb_sync           BOOLEAN
   
   DEFINE l_status          LIKE type_t.num10
   DEFINE l_start           LIKE type_t.num10
   DEFINE l_end             LIKE type_t.num10    
   
   DEFINE l_code            STRING
   DEFINE l_message         STRING
   DEFINE l_payload         STRING
   DEFINE lr_result         RECORD
                              status LIKE type_t.num5,
                              data   STRING
                            END RECORD
                            
   DEFINE ls_ip             STRING
   DEFINE ls_id             STRING
                              
   IF p_json_record.getLength() = 0 THEN 
      display "json length: 0"
      LET lr_result.status = -1
      RETURN lr_result.*
   END IF 
   
   IF NOT cl_eai_set_server() THEN
      LET lr_result.status = -1
      RETURN lr_result.*
   END IF  
   
   LET ls_prod     = p_json_record.get("prod")    
   LET ls_srvname  = p_json_record.get("srvname")    
   LET ls_ip       = p_json_record.get("ip")       
   LET ls_id       = p_json_record.get("id")       
   LET lb_sync     = p_json_record.get("async")
   LET ls_req_data = p_json_record.get("request")    
   
   #建立 Request XML Document
   LET l_service_attr[1].name = "prod"                                                                            
   LET l_service_attr[1].value = ls_prod
   
   LET l_service_attr[2].name = "name"                                                                            
   LET l_service_attr[2].value = ls_srvname
   
   #由 prod 產品別取回IP
   LET l_service_attr[3].name = "ip"
   IF cl_null(ls_ip) THEN
      LET l_service_attr[3].value = "10.40.41.216"
   ELSE
      LET l_service_attr[3].value = ls_ip
   END IF
   
   LET l_service_attr[4].name = "id"                                                                            
   LET l_service_attr[4].value = ls_id    #

   CALL cl_ws_create_request_head("sync",l_service_attr)
        RETURNING l_req_doc, l_request_root          
    
   #取得 host 及 service 字串 <host prod=.../>
   LET l_Request = l_request_root.toString()
   LET l_start = l_Request.getIndexOf("<host",1)
   LET l_end = l_Request.getIndexOf("/>",l_Request.getIndexOf("<service",l_start))   
   LET l_xml_str =  l_Request.subString(l_start,l_end+1)
   CALL l_request_root.setAttribute("key", cl_trust_crosskey(l_xml_str))
            
   # payload 資訊
   LET l_pnode = l_req_doc.createElement("payload")            #建立 <payload>   
   CALL l_request_root.appendChild(l_pnode)

   #add by 160527-00003 
   LET l_param_node =  l_req_doc.createElement("param")      #建立 <param>
   CALL l_param_node.setAttribute("key" ,"data")
   CALL l_param_node.setAttribute("type" ,"XML")
   #add by 160527-00003   

   #指定 payload 為 CDATA 型態
   LET l_tnode = l_req_doc.createCDATASection(ls_req_data.trim())
   #CALL l_pnode.appendChild(l_tnode)    #mark by 160527-00003
   
   CALL l_param_node.appendChild(l_tnode)      #add by 160527-00003
   CALL l_pnode.appendChild(l_param_node)      #add by 160527-00003
      
   #若 60 秒內無回應則放棄
   CALL fgl_ws_setOption("http_invoketimeout", 60)

   LET l_Request = l_request_root.toString()  
   
   #確認 CROSS 編碼
   LET l_Request = cl_trust_coding_en(l_Request) 
   
   #呼叫 CROSS 服務
   CALL cl_eai_callSrv(l_Request)  RETURNING l_status, l_Result
    
   IF l_status != 0 THEN
      LET lr_result.status = l_status
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = 'lib-00299'     #呼叫 CROSS 服務失敗
      LET g_errparam.extend = l_status
      LET g_errparam.popup = TRUE
      CALL cl_err()
      
      RETURN lr_result.*
   END IF
   
   #確認 CROSS 解碼
   LET l_Result = cl_trust_coding_de(l_Result)
      
   LET l_response = xml.DomDocument.Create()
   CALL l_response.loadFromString(l_Result)
  
   LET l_code = cl_bpm_get_node_value(l_response, "code")
   LET l_message = cl_bpm_get_node_value(l_response, "message")
   
   IF l_code = "019" THEN  #產品資訊
      #LET l_payload = cl_bpm_get_node_value(l_response, "payload")    
      LET l_payload = cl_ws_getNodeValue(l_response, "param","")
      LET lr_result.data = l_payload
   ELSE
      LET lr_result.status = 100
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = 'lib-00299'     #呼叫 CROSS 服務失敗
      LET g_errparam.extend = l_message
      LET g_errparam.popup = TRUE
      CALL cl_err()
   
   END IF  
   
   RETURN lr_result.*
END FUNCTION

################################################################################
# Descriptions...: 取出中間庫
# Memo...........:
# Usage..........: CALL cl_eai_get_middb(ps_dbs)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_eai_get_middb(ps_dbs)
   DEFINE ps_dbs      STRING
   DEFINE ls_middb    STRING
   #180912-00001#1 add by 10553  (S)
   DEFINE ls_dbs      LIKE gzda_t.gzda001
   DEFINE l_gzda010   LIKE gzda_t.gzda010
   #180912-00001#1 add by 10553  (E)
   
   #180912-00001#1 mark by 10553  (S)
   #CASE ps_dbs
   #     WHEN 'dsdemo'
   #        LET ls_middb = 'dsawst'
   #     WHEN 'dsdata'
   #        LET ls_middb = 'dsaws'
   #     OTHERWISE
   #        LET ls_middb = 'dsawst'
   #END CASE
   #180912-00001#1 mark by 10553  (E)
   
   #180912-00001#1 add by 10553  (S)
   LET ls_dbs = ps_dbs
   #從azzi085的設定資料取得ERPDB與AWSDB對應關係
   TRY
      SELECT gzda010 INTO l_gzda010
        FROM gzda_t
       WHERE gzda001 = ls_dbs AND gzdastus = 'Y'
   CATCH
      INITIALIZE l_gzda010 TO NULL
      DISPLAY "Failed to get data of azzi085(gzda_t)."
   END TRY
   
   IF cl_null(l_gzda010) THEN
      CASE ps_dbs
        WHEN 'dsdemo'
           LET ls_middb = 'dsawst'
        WHEN 'dsdata'
           LET ls_middb = 'dsaws'
        OTHERWISE
           LET ls_middb = 'dsawst'
      END CASE
   ELSE
      LET ls_middb = l_gzda010
   END IF
   #180912-00001#1 add by 10553  (E)
   #200116-00011#1 add by 07558 -s-
   #200116-00011#1 add by 07558 -s-   
   RETURN ls_middb
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
PUBLIC FUNCTION cl_eai_set_server()
   #170512-00048#2-S
   LET g_eai_server = cl_aws_prod_para(g_enterprise, "eai-0002")
   IF cl_null(g_eai_server) THEN
      LET g_eai_server = cl_get_para(g_enterprise,"","E-SYS-0704") 
   END IF
   #170512-00048#2-E
   IF cl_null(g_eai_server) THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code =  "aws-00024"
      LET g_errparam.extend = ""
      LET g_errparam.popup = TRUE
      CALL cl_err()
   
      RETURN FALSE
   END IF
   RETURN TRUE
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
PUBLIC FUNCTION cl_eai_callSrv(p_str)
  DEFINE        p_str           STRING
  DEFINE        soapStatus      INTEGER

  LET invokeSrv.in0 = p_str
  
  LET soapStatus = cl_eai_invokeSrv_g()

  RETURN soapStatus, invokeSrvResponse.invokeSrvReturn
  
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
PUBLIC FUNCTION cl_eai_invokeSrv_g()
  DEFINE wsstatus   INTEGER
  DEFINE retryAuth  INTEGER
  DEFINE retryProxy INTEGER
  DEFINE retry      INTEGER
  DEFINE nb         INTEGER
  DEFINE uri        STRING
  DEFINE setcookie  STRING
  DEFINE mustUnderstand INTEGER
  DEFINE request    com.HTTPRequest
  DEFINE response   com.HTTPResponse
  DEFINE writer     xml.StaxWriter
  DEFINE reader     xml.StaxReader
  #
  # INIT VARIABLES
  #
  LET wsstatus = -1
  LET retryAuth = FALSE
  LET retryProxy = FALSE
  LET retry = TRUE
  LET uri = com.WebServiceEngine.GetOption("SoapModuleURI")

  IF IntegrationEntry_IntegrationEntryEndpoint.Address.Uri IS NULL THEN
    LET IntegrationEntry_IntegrationEntryEndpoint.Address.Uri = g_eai_server 
  END IF

  #
  # CREATE REQUEST
  #
  TRY
    LET request = com.HTTPRequest.Create(IntegrationEntry_IntegrationEntryEndpoint.Address.Uri)
    CALL request.setMethod("POST")
    CALL request.setCharset("UTF-8")
    CALL request.setHeader("SOAPAction","\"\"")
    CALL WSHelper_SetRequestHeaders(request, IntegrationEntry_IntegrationEntryEndpoint.Binding.Request.Headers)
    IF IntegrationEntry_IntegrationEntryEndpoint.Binding.Version IS NOT NULL THEN
      CALL request.setVersion(IntegrationEntry_IntegrationEntryEndpoint.Binding.Version)
    END IF
    IF IntegrationEntry_IntegrationEntryEndpoint.Binding.Cookie IS NOT NULL THEN
      CALL request.setHeader("Cookie",IntegrationEntry_IntegrationEntryEndpoint.Binding.Cookie)
    END IF
    IF IntegrationEntry_IntegrationEntryEndpoint.Binding.ConnectionTimeout <> 0 THEN
      CALL request.setConnectionTimeout(IntegrationEntry_IntegrationEntryEndpoint.Binding.ConnectionTimeout)
    END IF
    IF IntegrationEntry_IntegrationEntryEndpoint.Binding.ReadWriteTimeout <> 0 THEN
      CALL request.setTimeout(IntegrationEntry_IntegrationEntryEndpoint.Binding.ReadWriteTimeout)
    END IF
    IF IntegrationEntry_IntegrationEntryEndpoint.Binding.CompressRequest IS NOT NULL THEN
      CALL request.setHeader("Content-Encoding",IntegrationEntry_IntegrationEntryEndpoint.Binding.CompressRequest)
    END IF
    CALL request.setHeader("Accept-Encoding","gzip, deflate")
  CATCH
    LET wsstatus = STATUS
    CALL WSHelper_FillSOAP11WSError("Client","Cannot create HTTPRequest")
    RETURN wsstatus    
  END TRY

  # START LOOP
  WHILE retry
    LET retry = FALSE

    #
    # Stax request
    #
    TRY
      LET writer = request.beginXmlRequest()
      CALL WSHelper_WriteStaxSOAP11StartEnvelope(writer)
      CALL WSHelper_WriteStaxSOAP11StartBody(writer)
      #
      # STAX SOAP REQUEST SERIALIZE
      #
      CALL xml.Serializer.VariableToStax(invokeSrv,writer)
      CALL WSHelper_WriteStaxSOAP11EndBody(writer)
      CALL WSHelper_WriteStaxSOAP11EndEnvelope(writer)
      CALL request.endXmlRequest(writer)
    CATCH
      LET wsstatus = STATUS
      CALL WSHelper_FillSOAP11WSError("Client",SQLCA.SQLERRM)
      RETURN wsstatus    
    END TRY

    #
    # PROCESS RESPONSE
    #
    TRY
      LET response = request.getResponse()

      #
      # RETRIEVE SERVICE SESSION COOKIE
      #
      LET setcookie = response.getHeader("Set-Cookie")
      IF setcookie IS NOT NULL THEN
        LET IntegrationEntry_IntegrationEntryEndpoint.Binding.Cookie = WSHelper_ExtractServerCookie(setcookie,IntegrationEntry_IntegrationEntryEndpoint.Address.Uri)
      END IF

      #
      # RETRIEVE HTTP RESPONSE Headers
      #
      CALL WSHelper_SetResponseHeaders(response, IntegrationEntry_IntegrationEntryEndpoint.Binding.Response.Headers)
      CASE response.getStatusCode()

        WHEN 500 # SOAP Fault
          #
          # STAX SOAP FAULT
          #
          LET reader = response.beginXmlResponse() # Begin Streaming Response
          IF NOT WSHelper_ReadStaxSOAP11StartEnvelope(reader) THEN
            EXIT CASE
          END IF
          # Process SOAP headers 
          IF WSHelper_CheckStaxSOAP11Header(reader) AND NOT reader.isEmptyElement() THEN
            CALL reader.nextTag()
            WHILE (reader.getEventType()=="START_ELEMENT")
              IF WSHelper_CheckStaxSOAP11HeaderActor(reader,uri) THEN
                LET mustUnderstand = WSHelper_GetStaxSOAP11HeaderMustUnderstand(reader)
                IF mustUnderstand = -1 THEN
                  CALL WSHelper_FillSOAP11WSError("Client","Invalid mustUnderstand value")
                  EXIT CASE
                END IF
                IF mustUnderstand THEN
                  CALL WSHelper_FillSOAP11WSError("MustUnderstand","Mandatory header block not understood")
                  EXIT CASE
                ELSE
                  CALL reader.nextSibling() # Skip header, not necessary
                END IF
              ELSE
                CALL reader.nextSibling() # Skip header, not intended to us
              END IF
            END WHILE
            IF NOT WSHelper_CheckStaxSOAP11Header(reader) THEN
              CALL WSHelper_FillSOAP11WSError("Client","No ending header tag found")
              EXIT CASE
            ELSE
              CALL reader.nextTag()
            END IF
          ELSE
            IF WSHelper_CheckStaxSOAP11Header(reader) THEN
              CALL reader.nextSibling() # Skip SOAP headers
            END IF
          END IF
          IF NOT WSHelper_ReadStaxSOAP11StartBody(reader) THEN
            EXIT CASE
          END IF
          IF NOT WSHelper_ReadStaxSOAP11StartFault(reader) THEN
            EXIT CASE
          END IF
          IF NOT WSHelper_ReadStaxSOAP11FaultUntilDetail(reader) THEN
            EXIT CASE
          END IF
          IF WSHelper_CheckStaxSOAP11FaultDetail(reader) THEN
            #
            # STAX SOAP FAULT DESERIALIZE
            #
            CALL reader.nextSibling() # Skip SOAP detail
          END IF
          IF NOT WSHelper_ReadStaxSOAP11EndFault(reader) THEN
            EXIT CASE
          END IF
          IF NOT WSHelper_ReadStaxSOAP11EndBody(reader) THEN
            EXIT CASE
          END IF
          IF NOT WSHelper_ReadStaxSOAP11EndEnvelope(reader) THEN
            EXIT CASE
          END IF
          # End Streaming Response
          CALL response.endXmlResponse(reader)

        WHEN 200 # SOAP Result
          #
          # STAX SOAP RESPONSE
          #
          LET reader = response.beginXmlResponse() # Begin Streaming Response
          IF NOT WSHelper_ReadStaxSOAP11StartEnvelope(reader) THEN
            EXIT CASE
          END IF
          # Process SOAP headers 
          IF WSHelper_CheckStaxSOAP11Header(reader) AND NOT reader.isEmptyElement() THEN
            LET nb = 0
            CALL reader.nextTag()
            WHILE (reader.getEventType()=="START_ELEMENT")
              IF WSHelper_CheckStaxSOAP11HeaderActor(reader,uri) THEN
                LET mustUnderstand = WSHelper_GetStaxSOAP11HeaderMustUnderstand(reader)
                IF mustUnderstand = -1 THEN
                  CALL WSHelper_FillSOAP11WSError("Client","Invalid mustUnderstand value")
                  EXIT CASE
                END IF
                #
                # STAX SOAP RESPONSE HEADER DESERIALIZE
                #
                IF mustUnderstand THEN
                  CALL WSHelper_FillSOAP11WSError("MustUnderstand","Mandatory header block not understood")
                  EXIT CASE
                ELSE
                  CALL reader.nextSibling() # Skip header, not necessary
                END IF
              ELSE
                CALL reader.nextSibling() # Skip header, not intended to us
              END IF
            END WHILE
            IF NOT WSHelper_CheckStaxSOAP11Header(reader) THEN
              CALL WSHelper_FillSOAP11WSError("Client","No ending header tag found")
              EXIT CASE
            ELSE
              IF nb != 0 THEN
                CALL WSHelper_FillSOAP11WSError("Client","One or more headers are missing")
                EXIT CASE
              ELSE
                CALL reader.nextTag()
              END IF
            END IF
          ELSE
           IF reader.isEmptyElement() THEN
             CALL reader.nextTag()
           END IF
          END IF
          IF NOT WSHelper_ReadStaxSOAP11StartBody(reader) THEN
            EXIT CASE
          END IF
          # Retrieve SOAP Message taking soap:root attribute into account
          IF NOT WSHelper_RetrieveStaxSOAP11Message(reader) THEN
            EXIT CASE
          END IF
          #
          # STAX SOAP RESPONSE DESERIALIZE
          #
          CALL xml.Serializer.StaxToVariable(reader,invokeSrvResponse)
          IF NOT WSHelper_ReadStaxSOAP11EndBody(reader) THEN
            EXIT CASE
          END IF
          IF NOT WSHelper_ReadStaxSOAP11EndEnvelope(reader) THEN
            EXIT CASE
          END IF
          # End Streaming Response
          CALL response.endXmlResponse(reader)
          LET wsstatus = 0

        WHEN 401 # HTTP Authentication
          IF retryAuth THEN
            CALL WSHelper_FillSOAP11WSError("Server","HTTP Error 401 ("||response.getStatusDescription()||")")
          ELSE
            LET retryAuth = TRUE
            LET retry = TRUE
          END IF

        WHEN 407 # Proxy Authentication
          IF retryProxy THEN
            CALL WSHelper_FillSOAP11WSError("Server","HTTP Error 407 ("||response.getStatusDescription()||")")
          ELSE
            LET retryProxy = TRUE
            LET retry = TRUE
          END IF

        OTHERWISE
          CALL WSHelper_FillSOAP11WSError("Server","HTTP Error "||response.getStatusCode()||" ("||response.getStatusDescription()||")")

      END CASE
    CATCH
      LET wsstatus = status
      CALL WSHelper_FillSOAP11WSError("Server",SQLCA.SQLERRM)
      RETURN wsstatus    
    END TRY

  # END LOOP
  END WHILE

  RETURN wsstatus 
  
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_eai_string_to_file(type,binary,fileName)
#                  RETURNING STRING
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2016/12/21 By Hank Wu
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_eai_forClient()
#  DEFINE serviceType LIKE type_t.num5
#  DEFINE binary      STRING
#  DEFINE fileName    STRING
#  DEFINE l_outputs   STRING
#  DEFINE l_timestamp     LIKE type_t.chr50
#  DEFINE l_xml_str       STRING              #產生requestKey用字串
#  DEFINE l_key           STRING              #用來存放requestKey回傳值
#  DEFINE l_req    STRING
#  
#  SELECT TO_CHAR(SYSDATE,'YYYYMMDDHH24MISSSSS')
#     INTO l_timestamp
#  FROM dual
#  
#  LET l_xml_str = '<host prod="T100" ver="1.0" ip="10.40.40.18" id="t10tst" timestamp="',l_timestamp,'" acct="" />',                     
#    '<service prod="MES" name="create_mo" srvver="1.0"/>'
#  LET l_key = cl_trust_crosskey(l_xml_str)
#  DISPLAY "l_key=",l_key
#  
#  LET fileName = "/u1/ociis/t10dev/s_transfer/",fileName ,".zip"             #壓縮檔完整路徑+檔名
#  DISPLAY "壓縮檔完整路徑+檔名 = ",fileName
#  LET l_outputs = security.Base64.LoadBinary(fileName)                     #進行轉base64動作
#  
#  CASE serviceType
#     WHEN 1  #XML
#        LET l_req =
#        '<request type="sync" key="',l_key,'">',                        
#    '<host prod="T100" ver="1.0" ip="10.40.40.18" id="t10tst" timestamp="',l_timestamp,'" acct="" />',                     
#    '<service prod="MES" name="create_mo" srvver="1.0"/>',                            
#   '<payload>',                                
#      '<param key="zip_data" type="string" >',l_outputs,'</param>',
#      '<param key="file_name" type="string" >',fileName,'</param>',      #待確認
#   '</payload>',                             
#'</request>'
#   
#     WHEN 2  #JSON
#     
#  END CASE
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_eai_serial_to_string(p_filetype,p_zipdata,p_filename)
#                  RETURNING 回传参数
# Input parameter: p_filetype     檔案類型
#                : p_zipdata      序列化字串
#                : p_filename     zip檔案內的檔案名稱
# Return code....: l_str          檔案字串
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 20170103 By yc.chao
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_eai_serial_to_string(p_filetype,p_zipdata,p_filename)
  DEFINE p_filetype    LIKE type_t.num5
  DEFINE p_zipdata     STRING
  DEFINE p_filename    STRING
  DEFINE l_str         STRING
  DEFINE l_cmd         STRING
  DEFINE l_zippath     STRING
  DEFINE l_filepath    STRING
  DEFINE l_time        STRING
  DEFINE ls_channel        base.Channel
  DEFINE lb_return     BOOLEAN
  
  LET l_time = CURRENT HOUR TO FRACTION(3)
  LET l_zippath = fgl_getenv("TEMPDIR"),'/', l_time , "_" , FGL_GETPID() USING '<<<<<<<<<<',".zip"
  
  CALL security.Base64.SaveBinary(l_zippath,p_zipdata)
  IF NOT os.Path.exists(l_zippath) THEN
     LET l_str = 'SaveBinary Failed!'
     DISPLAY l_str
     RETURN l_str
  END IF

  LET l_filepath = fgl_getenv("TEMPDIR"),'/',p_filename
  IF os.Path.exists(l_filepath) THEN
     LET lb_return = os.Path.delete(l_filepath)
  END IF  

  LET l_cmd = 'unzip -o ',l_zippath,' -d ',fgl_getenv("TEMPDIR"),'/'
  RUN l_cmd

  IF NOT os.Path.exists(l_filepath) THEN
     LET l_str = 'Unzip Failed!'
     DISPLAY l_str
     RETURN l_str
  END IF  
  
  LET ls_channel = base.Channel.create()
  CALL ls_channel.setDelimiter("")
  CALL ls_channel.openFile(l_filepath,"r")
  LET l_str = ls_channel.readLine()
  CALL ls_channel.close()  
  
  return l_str
END FUNCTION
################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_eai_ftpfile_to_string(p_filetype,p_zipdata,p_filename)
#                  RETURNING 回传参数
# Input parameter: p_filetype     檔案類型
#                : p_zipdata      序列化字串
#                : p_filename     zip檔案內的檔案名稱
# Return code....: l_str          檔案字串
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 20170103 By yc.chao
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_eai_ftpfile_to_string(p_filetype,p_zipdata,p_filename)
  DEFINE p_filetype    LIKE type_t.num5
  DEFINE p_zipdata     STRING
  DEFINE p_filename    STRING
  DEFINE l_str         STRING
  DEFINE l_cmd         STRING
  DEFINE l_zippath     STRING
  DEFINE l_filepath    STRING
  DEFINE l_time        STRING
  DEFINE ls_channel        base.Channel
  DEFINE l_temppath    STRING
  DEFINE lb_return     BOOLEAN

  LET l_time = CURRENT HOUR TO FRACTION(3)
  LET l_zippath = fgl_getenv("OCIISPATH"),'/', p_zipdata
  
  IF NOT os.Path.exists(l_zippath) THEN
     LET l_str = 'FTP FILE NOT exist!'
     DISPLAY l_str
     RETURN l_str
  END IF
  
  LET l_temppath = fgl_getenv("TEMPDIR"),'/'
  LET l_filepath = fgl_getenv("TEMPDIR"),'/',p_filename
  IF os.Path.exists(l_filepath) THEN
     LET lb_return = os.Path.delete(l_filepath)
  END IF  
  #LET l_cmd = 'cp ',l_zippath,' ',l_temppath
  LET l_cmd = 'unzip -o ',l_zippath,' -d ',l_temppath
  RUN l_cmd
  
  IF NOT os.Path.exists(l_filepath) THEN
     LET l_str = 'Unzip Failed!'
     DISPLAY l_str
     RETURN l_str
  END IF  
  
  LET ls_channel = base.Channel.create()
  CALL ls_channel.setDelimiter("")
  CALL ls_channel.openFile(l_filepath,"r")
  LET l_str = ls_channel.readLine()
  CALL ls_channel.close()  
  
  return l_str
END FUNCTION
################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_eai_string_to_serial(p_filetype,p_zipdata,p_filename)
#                  RETURNING 回传参数
# Input parameter: p_filetype     檔案類型
#                : p_zipdata      序列化字串
#                : p_filename     zip檔案內的檔案名稱
# Return code....: l_str          檔案字串
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 20170103 By yc.chao
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_eai_string_to_serial(p_filetype,p_zipdata,p_filename)
  DEFINE p_filetype    LIKE type_t.num5
  DEFINE p_zipdata     STRING
  DEFINE p_filename    STRING
  DEFINE l_str         STRING
  DEFINE l_cmd         STRING
  DEFINE l_zippath     STRING
  DEFINE l_filepath    STRING
  DEFINE l_temppath    STRING
  DEFINE l_time        STRING
  DEFINE ls_channel        base.Channel
  DEFINE lb_return     BOOLEAN

  CASE p_filetype
    WHEN 1
      LET l_filepath = fgl_getenv("TEMPDIR"),'/',p_filename,'.xml'
    WHEN 2
      LET l_filepath = fgl_getenv("TEMPDIR"),'/',p_filename,'.json'
  END CASE    
  
  LET ls_channel = base.Channel.create()
  CALL ls_channel.openFile(l_filepath, "w")
  CALL ls_channel.writeLine(p_zipdata.trim())
  CALL ls_channel.close()
      
  LET l_zippath = fgl_getenv("TEMPDIR"),'/',p_filename,'.zip' 
  IF os.Path.exists(l_zippath) THEN
     LET lb_return = os.Path.delete(l_zippath)
  END IF    
  
  LET l_cmd = 'zip ',l_zippath,' ',l_filepath
  RUN l_cmd
  
  IF NOT os.Path.exists(l_zippath) THEN
     LET l_str = 'zip Failed!'
     DISPLAY l_str
     RETURN l_str
  END IF  

  LET l_str = security.Base64.LoadBinary(l_zippath)
  RETURN l_str
END FUNCTION
################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_eai_string_to_ftpfile(p_filetype,p_zipdata,p_filename)
#                  RETURNING 回传参数
# Input parameter: p_filetype     檔案類型
#                : p_zipdata      序列化字串
#                : p_filename     zip檔案內的檔案名稱
# Return code....: l_str          檔案字串
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 20170103 By yc.chao
# Modify.........:
################################################################################
PUBLIC FUNCTION cl_eai_string_to_ftpfile(p_filetype,p_zipdata,p_filename)
  DEFINE p_filetype    LIKE type_t.num5
  DEFINE p_zipdata     STRING
  DEFINE p_filename    STRING
  DEFINE l_str         STRING
  DEFINE l_cmd         STRING
  DEFINE l_zippath     STRING
  DEFINE l_filepath    STRING
  DEFINE l_temppath    STRING
  DEFINE l_time        STRING
  DEFINE ls_channel        base.Channel
  DEFINE lb_return     BOOLEAN

  CASE p_filetype
    WHEN 1
      LET l_filepath = fgl_getenv("TEMPDIR"),'/',p_filename,'.xml'
    WHEN 2
      LET l_filepath = fgl_getenv("TEMPDIR"),'/',p_filename,'.json'
  END CASE    
  
  LET ls_channel = base.Channel.create()
  CALL ls_channel.openFile(l_filepath, "w")
  CALL ls_channel.writeLine(p_zipdata.trim())
  CALL ls_channel.close()
      
  LET l_temppath = fgl_getenv("TEMPDIR"),'/',p_filename,'.zip' 
  IF os.Path.exists(l_temppath) THEN
     LET lb_return = os.Path.delete(l_temppath)
  END IF    
  
  LET l_cmd = 'zip ',l_temppath,' ',l_filepath
  RUN l_cmd
  
  IF NOT os.Path.exists(l_temppath) THEN
     LET l_str = 'zip Failed!'
     DISPLAY l_str
     RETURN l_str
  END IF  

  LET l_zippath = fgl_getenv("OCIISPATH"),'/s_transfer/',p_filename,'.zip'
  LET l_str = 's_transfer/',p_filename,'.zip'
  
  IF os.Path.exists(l_zippath) THEN
     LET lb_return = os.Path.delete(l_zippath)
  END IF   
  LET l_cmd = 'cp ',l_temppath,' ',l_zippath
  RUN l_cmd
  IF NOT os.Path.exists(l_zippath) THEN
     LET l_str = 'copy to FTP Failed!'
     DISPLAY l_str
     RETURN l_str
  END IF    
  
  RETURN l_str
  
#{"zip_data":"UEsDBBQAAAAIANVMJEpWG/+GPAAAAFoAAAAOABwAanNvbl90ZXN0Lmpzb25VVAkAAyJSbFjGUmxYdXgLAAEEddQAAAT3AQAAU6hWUCouSYlPSSxJVLJSqFaAAqWCxKLE3NSS1CJkUbBMam5BTn5lamp8Xj5QTsnA3NjcVAkmWwsla7kAUEsBAh4DFAAAAAgA1UwkSlYb/4Y8AAAAWgAAAA4AGAAAAAAAAQAAALSBAAAAAGpzb25fdGVzdC5qc29uVVQFAAMiUmxYdXgLAAEEddQAAAT3AQAAUEsFBgAAAAABAAEAVAAAAIQAAAAAAA==","file_name":"json_test.json"}

#{ "std_data": {        "parameter": {            "employee_no": "07375"        }     } }


#{"zip_path":"r_pending/json_test.zip","file_name":"json_test.json"}
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_eai_cross_getProdList()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........: 170512-00048#2
################################################################################
PUBLIC FUNCTION cl_eai_cross_getProdList()
  DEFINE        soapStatus              INTEGER
  
  LET soapStatus = cl_eai_cross_getProdList_g()
  
  RETURN soapStatus, getProdListResponse.getProdListReturn
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_eai_cross_getProdList_g()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........: 170512-00048#2
################################################################################
PUBLIC FUNCTION cl_eai_cross_getProdList_g()
  DEFINE wsstatus   INTEGER
  DEFINE retryAuth  INTEGER
  DEFINE retryProxy INTEGER
  DEFINE retry      INTEGER
  DEFINE request    com.HTTPRequest
  DEFINE response   com.HTTPResponse
  DEFINE writer     xml.StaxWriter
  DEFINE reader     xml.StaxReader
  DEFINE uri        STRING

  IF NOT cl_eai_set_server() THEN
     DISPLAY "get eai server error"
  END IF
  # 
  # INIT VARIABLES
  #
  LET wsstatus = -1
  LET retryAuth = FALSE
  LET retryProxy = FALSE
  LET retry = TRUE
  LET uri = com.WebServiceEngine.GetOption("SoapModuleURI")

  IF IntegrationEntry_IntegrationEntryEndpoint.Address.Uri IS NULL THEN
    LET IntegrationEntry_IntegrationEntryEndpoint.Address.Uri = g_eai_server 
  END IF

  #
  # CREATE REQUEST
  #
  TRY
    LET request = com.HTTPRequest.Create(IntegrationEntry_IntegrationEntryEndpoint.Address.Uri)
    CALL request.setMethod("POST")
    CALL request.setCharset("UTF-8")
    CALL request.setHeader("SOAPAction","\"\"")
  CATCH
    LET wsstatus = STATUS
    CALL WSHelper_FillSOAP11WSError("Client","Cannot create HTTPRequest")
    RETURN wsstatus
  END TRY

  # START LOOP
  WHILE retry
    LET retry = FALSE

    #
    # Stax request
    #
    TRY
      LET writer = request.beginXmlRequest()
      CALL WSHelper_WriteStaxSOAP11StartEnvelope(writer)
      CALL WSHelper_WriteStaxSOAP11StartBody(writer)
      CALL writer.setPrefix("fjs","http://entry.serviceengine.cross.digiwin.com")
      CALL writer.emptyElementNS("getProdList","http://entry.serviceengine.cross.digiwin.com")
      CALL WSHelper_WriteStaxSOAP11EndBody(writer)
      CALL WSHelper_WriteStaxSOAP11EndEnvelope(writer)
      CALL request.endXmlRequest(writer)
    CATCH
      LET wsstatus = STATUS
      CALL WSHelper_FillSOAP11WSError("Client",SQLCA.SQLERRM)
      RETURN wsstatus
    END TRY

    #
    # PROCESS RESPONSE
    #
    TRY
      LET response = request.getResponse()

      CASE response.getStatusCode()

        WHEN 500 # SOAP Fault
          #
          # STAX SOAP FAULT
          #
          LET reader = response.beginXmlResponse() # Begin Streaming Response
          IF NOT WSHelper_ReadStaxSOAP11StartEnvelope(reader) THEN
            EXIT CASE
          END IF
          IF WSHelper_CheckStaxSOAP11Header(reader) THEN
            CALL reader.nextSibling() # Skip SOAP headers
          END IF
          IF NOT WSHelper_ReadStaxSOAP11StartBody(reader) THEN
            EXIT CASE
          END IF
          IF NOT WSHelper_ReadStaxSOAP11Fault(reader) THEN
            EXIT CASE
          END IF
          IF NOT WSHelper_ReadStaxSOAP11EndBody(reader) THEN
            EXIT CASE
          END IF
          IF NOT WSHelper_ReadStaxSOAP11EndEnvelope(reader) THEN
            EXIT CASE
          END IF
          # End Streaming Response
          CALL response.endXmlResponse(reader)

        WHEN 200 # SOAP Result
          #
          # STAX SOAP RESPONSE
          #
          LET reader = response.beginXmlResponse() # Begin Streaming Response
          IF NOT WSHelper_ReadStaxSOAP11StartEnvelope(reader) THEN
            EXIT CASE
          END IF
          IF WSHelper_CheckStaxSOAP11Header(reader) THEN
            IF NOT reader.isEmptyElement() THEN
              CALL WSHelper_FillSOAP11WSError("Client","No SOAP Header expected")
              EXIT CASE
            ELSE
              CALL reader.nextTag()
            END IF
          END IF
          IF NOT WSHelper_ReadStaxSOAP11StartBody(reader) THEN
            EXIT CASE
          END IF
          # Retrieve SOAP Message taking soap:root attribute into account
          IF NOT WSHelper_RetrieveStaxSOAP11Message(reader) THEN
            EXIT CASE
          END IF
          CALL xml.Serializer.StaxToVariable(reader,getProdListResponse)
          IF NOT WSHelper_ReadStaxSOAP11EndBody(reader) THEN
            EXIT CASE
          END IF
          IF NOT WSHelper_ReadStaxSOAP11EndEnvelope(reader) THEN
            EXIT CASE
          END IF
          # End Streaming Response
          CALL response.endXmlResponse(reader)
          LET wsstatus = 0

        WHEN 401 # HTTP Authentication
          IF retryAuth THEN
            CALL WSHelper_FillSOAP11WSError("Server","HTTP Error 401 ("||response.getStatusDescription()||")")
          ELSE
            LET retryAuth = TRUE
            LET retry = TRUE
          END IF

        WHEN 407 # Proxy Authentication
          IF retryProxy THEN
            CALL WSHelper_FillSOAP11WSError("Server","HTTP Error 407 ("||response.getStatusDescription()||")")
          ELSE
            LET retryProxy = TRUE
            LET retry = TRUE
          END IF

        OTHERWISE
          CALL WSHelper_FillSOAP11WSError("Server","HTTP Error "||response.getStatusCode()||" ("||response.getStatusDescription()||")")

      END CASE
    CATCH
      LET wsstatus = status
      CALL WSHelper_FillSOAP11WSError("Server",SQLCA.SQLERRM)
      RETURN wsstatus
    END TRY

  # END LOOP
  END WHILE

  RETURN wsstatus

END FUNCTION

################################################################################
# Descriptions...: 查詢已註冊產品清單
# Memo...........:
# Usage..........: CALL cl_eai_getProdList(p_prod_list)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2017/05/17 by nikoybeat
# Modify.........: 170512-00048#2
################################################################################
PUBLIC FUNCTION cl_eai_getProdList(p_prod_list)
DEFINE p_prod_list       DYNAMIC ARRAY OF RECORD
       #241025-00007#1 add(s)
       #name              LIKE type_t.chr30,       #產品名稱
       #ver               LIKE type_t.chr20,       #產品版本
       #ip                LIKE type_t.chr20,       #IP位置
       #id                LIKE type_t.chr20,       #識別碼
       name              LIKE wsef_t.wsef001,      #產品名稱
       ver               LIKE wsef_t.wsef002,      #產品版本
       ip                LIKE wsef_t.wsef003,      #IP位置
       id                LIKE wsef_t.wsef004,      #識別碼       
       #241025-00007#1 add(e)       
       wsdl              LIKE wsef_t.wsef005      #wsdl 
       END RECORD
DEFINE l_Request         STRING
DEFINE l_Result          STRING
DEFINE l_host_att        STRING
DEFINE l_status          LIKE type_t.num10
DEFINE l_i               LIKE type_t.num10
DEFINE l_node_list       om.NodeList
DEFINE l_node            om.DomNode   
DEFINE l_str             STRING
DEFINE l_code            STRING
DEFINE l_message         STRING
DEFINE l_response_root   om.DomNode


    #取得 CROSS 整合參數資料
    #CALL aws_cross_wap()   

    #若 60 秒內無回應則放棄
    CALL fgl_ws_setOption("http_invoketimeout", 60)                   

    #呼叫 CROSS, 查詢已註冊的產品資訊
    CALL cl_eai_cross_getProdList()  RETURNING l_status, l_Result

    #紀錄傳遞 CROSS XML 字串
    #CALL aws_crosscli_logfile("getProdList",l_status,l_Request,l_Result)

    IF l_status != 0 THEN
       IF fgl_getenv('FGLGUI') = '1' THEN
          LET l_str = "Connection failed:\n\n"
          #         "  [Code]: ", wserror.code, "\n",
          #         "  [Action]: ", wserror.action, "\n", 
          #         "  [Description]: ", wserror.description
       ELSE        
          #LET l_str = "Connection failed: ", wserror.description
       END IF
       DISPLAY 'connection failed'       
       #CALL cl_err(l_str, '!', 1)   #連接失敗
       RETURN 
    END IF

    #取得 XML 字串的 Root Node
    LET l_response_root = cl_eai_stringToXml(l_Result)
    IF l_response_root IS NULL THEN
       DISPLAY "Response isn't valid XML document"
       RETURN 
    END IF

    # 尋找對應名稱的 tag  資料                                       
    LET l_code = cl_eai_getTagValue(l_response_root,"code","@chars")
    LET l_message = cl_eai_getTagValue(l_response_root,"message","@chars")

    CALL p_prod_list.clear()
    IF l_code = "091" THEN  #已註冊產品資訊
       LET l_node_list = l_response_root.selectByPath("//payload/prod")
       FOR l_i = 1 TO l_node_list.getLength()
           LET l_node = l_node_list.item(l_i)
           LET p_prod_list[l_i].name = l_node.getAttribute("name")
           LET p_prod_list[l_i].ver = l_node.getAttribute("ver")
           LET p_prod_list[l_i].ip = l_node.getAttribute("ip")
           LET p_prod_list[l_i].id = l_node.getAttribute("id")
           LET p_prod_list[l_i].wsdl = l_node.getAttribute("wsdl")
       END FOR  
    ELSE
      LET l_str = "XML parser :\n\n",
                    l_code.trim() ," ", l_message.trim() 
       DISPLAY l_str
       RETURN 
    END IF


END FUNCTION

################################################################################
# Descriptions...: 取得 xml tag value值
# Memo...........:
# Usage..........: CALL cl_eai_getTagValue(p_xml,p_tag,p_att_name)
#                  RETURNING l_value
# Input parameter: p_xml       - DomNode    - XML Root Node 
#                : p_tag       - STRING     - Tag Name
#                : p_att_name  - STRING     - Attribute Name
# Return code....: l_value
# Date & Author..: 2017/05/17 by nikoybeat
# Modify.........: 170512-00048#2
################################################################################
PUBLIC FUNCTION cl_eai_getTagValue(p_xml,p_tag,p_att_name)
DEFINE p_xml             om.DomNode
DEFINE p_tag             STRING
DEFINE p_att_name        STRING
DEFINE l_value           STRING
DEFINE l_node_list       om.NodeList
DEFINE l_snode           om.DomNode
DEFINE ln_node           om.DomNode
DEFINE l_node            om.DomNode

    LET l_value = NULL

    LET p_tag = p_tag.trim()
    LET p_att_name = p_att_name.trim()

    LET l_node_list = p_xml.selectByTagName(p_tag)
    IF l_node_list.getLength() != 0 THEN
       LET l_snode = l_node_list.item(1)	
       IF p_att_name = "@chars" THEN
          LET ln_node = l_snode.getFirstChild()
          IF ln_node IS NOT NULL THEN
             LET l_value = ln_node.getAttribute(p_att_name)
          END IF
       ELSE
          LET l_value = l_snode.getAttribute(p_att_name)
       END IF
    END IF
 
    RETURN l_value 
END FUNCTION
################################################################################
# Descriptions...: 取得 XML 字串的 Root Node, 提供後續作業處理
# Memo...........:
# Usage..........: CALL cl_eai_stringToXml(p_xml)
#                  RETURNING 回传参数
# Input parameter: p_xml  - STRING  - XML 字串
# Return code....: l_root - DomNode - XML Root Node
# Date & Author..: 2017/05/17 by nikoybeat
# Modify.........: 170512-00048#2
################################################################################
PUBLIC FUNCTION cl_eai_stringToXml(p_xml)
DEFINE p_xml       STRING
DEFINE l_t_doc     om.DomDocument    #For String to Xml
DEFINE l_ch        base.Channel,
       l_xmlFile   STRING,
       l_root      om.DomNode
    
    
    IF cl_null(p_xml) THEN
       RETURN NULL
    END IF
 
    #--------------------------------------------------------------------------#
    # 建立 Dom Node                                                            #
    #--------------------------------------------------------------------------#
    LET l_t_doc = om.DomDocument.createFromString(p_xml.trim())  

    IF l_t_doc IS NOT NULL THEN
       INITIALIZE l_root TO NULL
       LET l_root = l_t_doc.getDocumentElement()
    END IF
    
    RETURN l_root
END FUNCTION

################################################################################
# Descriptions...: 查詢特定產品的服務清單
# Memo...........:
# Usage..........: CALL cl_eai_getSrv(p_prod,p_ver,p_ip,p_id,p_srv_list)
#                  RETURNING TRUE/FALSE
# Input parameter: p_prod      -    STRING    - 產品名稱
#                : p_ver       -    STRING    - 版本
#                : p_ip        -    STRING    - IP
#                : p_id        -    STRING    - 識別碼
#                : p_srv_list  -    ARRAY     - 服務清單
# Return code....: TRUE/FALSE
# Date & Author..: 2017/06/23 by nikoybeat
# Modify.........: 170512-00048#2
################################################################################
PUBLIC FUNCTION cl_eai_getSrv(p_prod,p_ver,p_ip,p_id,p_srv_list)
DEFINE p_prod            STRING
DEFINE p_ver             STRING
DEFINE p_ip              STRING
DEFINE p_id              STRING
DEFINE p_srv_list       DYNAMIC ARRAY OF RECORD
       name              LIKE type_t.chr50       #服務名稱
       END RECORD
DEFINE l_Request         STRING
DEFINE l_Result          STRING
DEFINE l_host_att        STRING
DEFINE l_status          LIKE type_t.num10
DEFINE l_i               LIKE type_t.num10
DEFINE l_node_list       om.NodeList
DEFINE l_node            om.DomNode   
DEFINE l_str             STRING
DEFINE l_req_doc         xml.DomDocument
DEFINE l_code            STRING
DEFINE l_message         STRING
DEFINE l_response_root   om.DomNode
DEFINE l_request_root    xml.DomNode        #Request XML Node
DEFINE l_tag             STRING             #節點tag名稱
DEFINE l_attr            type_g_attribute   #節點屬性資料
DEFINE l_tnode            xml.DomNode


    INITIALIZE l_request_root TO NULL
    INITIALIZE l_tnode TO NULL
   
    #建立 Request XML Document
    LET l_req_doc = xml.DomDocument.createDocument("request")
    LET l_request_root = l_req_doc.getDocumentElement()       

   
    #建立TT主機資訊(<host>)
    LET l_tag = "host"
    CALL l_attr.clear()
    LET l_attr = cl_eai_get_base_info()
  
    CALL cl_bpm_create_node(l_req_doc, l_request_root, l_tag, l_attr)
       RETURNING l_req_doc, l_tnode
      
    #將 XML 文件轉換成 String 字串
    LET l_Request = l_req_doc.saveToString()

    #若 60 秒內無回應則放棄
    CALL fgl_ws_setOption("http_invoketimeout", 60)                   

    #呼叫 CROSS, 查詢已註冊產品的服務清單
    CALL cl_eai_cross_getSrv(l_Request)  RETURNING l_status, l_Result

    #紀錄傳遞 CROSS XML 字串
    #CALL aws_crosscli_logfile("getSrv",l_status,l_Request,l_Result)

    IF l_status != 0 THEN
       IF fgl_getenv('FGLGUI') = '1' THEN
          LET l_str = "Connection failed:\n\n"
                   #"  [Code]: ", wserror.code, "\n",
                   #"  [Action]: ", wserror.action, "\n", 
                   #"  [Description]: ", wserror.description
       ELSE        
          LET l_str = "Connection failed: "#, wserror.description
       END IF 
       #CALL cl_err(l_str, '!', 1)   #連接失敗
       RETURN  FALSE
    END IF


    #取得 XML 字串的 Root Node
    LET l_response_root = cl_eai_stringToXml(l_Result)
    IF l_response_root IS NULL THEN
       DISPLAY "Response isn't valid XML document"
       RETURN FALSE
    END IF

    # 尋找對應名稱的 tag  資料                                       
    LET l_code = cl_eai_getTagValue(l_response_root,"code","@chars")
    LET l_message = cl_eai_getTagValue(l_response_root,"message","@chars")

    CALL p_srv_list.clear()
    IF l_code = "093" THEN  #產品服務清單
       LET l_node_list = l_response_root.selectByTagName("srv")
       FOR l_i = 1 TO l_node_list.getLength()
           LET l_node = l_node_list.item(l_i)
           LET p_srv_list[l_i].name = l_node.getAttribute("name")
       END FOR  
    ELSE
       RETURN FALSE
    END IF

    RETURN TRUE
    
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_eai_cross_getSrv(p_in0)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........: 170512-00048#2
################################################################################
PUBLIC FUNCTION cl_eai_cross_getSrv(p_in0)
  DEFINE	soapStatus		INTEGER
  DEFINE	p_in0		STRING

  LET getSrv.in0 = p_in0

  LET soapStatus = cl_eai_cross_getSrv_g()

  RETURN soapStatus, getSrvResponse.getSrvReturn
END FUNCTION

################################################################################
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL cl_eai_cross_getSrv_g()
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 2017/06/23 by nikoybeat
# Modify.........: 170512-00048#2
################################################################################
PUBLIC FUNCTION cl_eai_cross_getSrv_g()
  DEFINE wsstatus   INTEGER
  DEFINE retryAuth  INTEGER
  DEFINE retryProxy INTEGER
  DEFINE retry      INTEGER
  DEFINE request    com.HTTPRequest
  DEFINE response   com.HTTPResponse
  DEFINE writer     xml.StaxWriter
  DEFINE reader     xml.StaxReader
  DEFINE uri        STRING

  IF NOT cl_eai_set_server() THEN
     DISPLAY "get eai server error"
  END IF
  #
  # INIT VARIABLES
  #
  LET wsstatus = -1
  LET retryAuth = FALSE
  LET retryProxy = FALSE
  LET retry = TRUE
  LET uri = com.WebServiceEngine.GetOption("SoapModuleURI")

  IF IntegrationEntry_IntegrationEntryEndpoint.Address.Uri IS NULL THEN
    LET IntegrationEntry_IntegrationEntryEndpoint.Address.Uri = g_eai_server 
  END IF

  #
  # CREATE REQUEST
  #
  TRY
    LET request = com.HTTPRequest.Create(IntegrationEntry_IntegrationEntryEndpoint.Address.Uri)
    CALL request.setMethod("POST")
    CALL request.setCharset("UTF-8")
    CALL request.setHeader("SOAPAction","\"\"")
  CATCH
    LET wsstatus = STATUS
    CALL WSHelper_FillSOAP11WSError("Client","Cannot create HTTPRequest")
    RETURN wsstatus
  END TRY

  # START LOOP
  WHILE retry
    LET retry = FALSE

    #
    # Stax request
    #
    TRY
      LET writer = request.beginXmlRequest()
      CALL WSHelper_WriteStaxSOAP11StartEnvelope(writer)
      CALL WSHelper_WriteStaxSOAP11StartBody(writer)
      CALL xml.Serializer.VariableToStax(getSrv,writer)
      CALL WSHelper_WriteStaxSOAP11EndBody(writer)
      CALL WSHelper_WriteStaxSOAP11EndEnvelope(writer)
      CALL request.endXmlRequest(writer)
    CATCH
      LET wsstatus = STATUS
      CALL WSHelper_FillSOAP11WSError("Client",SQLCA.SQLERRM)
      RETURN wsstatus
    END TRY

    #
    # PROCESS RESPONSE
    #
    TRY
      LET response = request.getResponse()

      CASE response.getStatusCode()

        WHEN 500 # SOAP Fault
          #
          # STAX SOAP FAULT
          #
          LET reader = response.beginXmlResponse() # Begin Streaming Response
          IF NOT WSHelper_ReadStaxSOAP11StartEnvelope(reader) THEN
            EXIT CASE
          END IF
          IF WSHelper_CheckStaxSOAP11Header(reader) THEN
            CALL reader.nextSibling() # Skip SOAP headers
          END IF
          IF NOT WSHelper_ReadStaxSOAP11StartBody(reader) THEN
            EXIT CASE
          END IF
          IF NOT WSHelper_ReadStaxSOAP11Fault(reader) THEN
            EXIT CASE
          END IF
          IF NOT WSHelper_ReadStaxSOAP11EndBody(reader) THEN
            EXIT CASE
          END IF
          IF NOT WSHelper_ReadStaxSOAP11EndEnvelope(reader) THEN
            EXIT CASE
          END IF
          # End Streaming Response
          CALL response.endXmlResponse(reader)

        WHEN 200 # SOAP Result
          #
          # STAX SOAP RESPONSE
          #
          LET reader = response.beginXmlResponse() # Begin Streaming Response
          IF NOT WSHelper_ReadStaxSOAP11StartEnvelope(reader) THEN
            EXIT CASE
          END IF
          IF WSHelper_CheckStaxSOAP11Header(reader) THEN
            IF NOT reader.isEmptyElement() THEN
              CALL WSHelper_FillSOAP11WSError("Client","No SOAP Header expected")
              EXIT CASE
            ELSE
              CALL reader.nextTag()
            END IF
          END IF
          IF NOT WSHelper_ReadStaxSOAP11StartBody(reader) THEN
            EXIT CASE
          END IF
          # Retrieve SOAP Message taking soap:root attribute into account
          IF NOT WSHelper_RetrieveStaxSOAP11Message(reader) THEN
            EXIT CASE
          END IF
          CALL xml.Serializer.StaxToVariable(reader,getSrvResponse)
          IF NOT WSHelper_ReadStaxSOAP11EndBody(reader) THEN
            EXIT CASE
          END IF
          IF NOT WSHelper_ReadStaxSOAP11EndEnvelope(reader) THEN
            EXIT CASE
          END IF
          # End Streaming Response
          CALL response.endXmlResponse(reader)
          LET wsstatus = 0

        WHEN 401 # HTTP Authentication
          IF retryAuth THEN
            CALL WSHelper_FillSOAP11WSError("Server","HTTP Error 401 ("||response.getStatusDescription()||")")
          ELSE
            LET retryAuth = TRUE
            LET retry = TRUE
          END IF

        WHEN 407 # Proxy Authentication
          IF retryProxy THEN
            CALL WSHelper_FillSOAP11WSError("Server","HTTP Error 407 ("||response.getStatusDescription()||")")
          ELSE
            LET retryProxy = TRUE
            LET retry = TRUE
          END IF

        OTHERWISE
          CALL WSHelper_FillSOAP11WSError("Server","HTTP Error "||response.getStatusCode()||" ("||response.getStatusDescription()||")")

      END CASE
    CATCH
      LET wsstatus = status
      CALL WSHelper_FillSOAP11WSError("Server",SQLCA.SQLERRM)
      RETURN wsstatus
    END TRY

  # END LOOP
  END WHILE

  RETURN wsstatus

END FUNCTION

 
{</section>}
 
