#+ Version..: T100-ERP-1.00.00 Build-000075

IMPORT os
SCHEMA ds

#+ code_id......: adzp161.4gl
#+ descriptions.: 報表程式資料建立
#+ code_creator.: kevin
#+ Modifier...: 230309-00010#1 00544 #新增XR報表元件：1。新增程式樣板 2。調整產生器
#+              240118-00026#1 00544 附屬表與主表相同採用別名
#+              240620-00031#1 00544 產出主表g_doc_key
#+              240823-00003#1 19425 报表元件产生备注子报表框架规则调整 XR数据源没有设定备注子报表关联资料则不产出相关框架
#+              250408-00016#1 05292 適配人大金倉增加numeric類型處理

GLOBALS "../../cfg/top_global.inc"
GLOBALS "../../cfg/top_report.inc"
 
GLOBALS
   DEFINE g_properties         om.SaxAttributes
   CONSTANT li_space = 3       #空白
   CONSTANT cs_prog_ver = 1.0  #151020-00008#1 add 
   DEFINE g_component_list     DYNAMIC ARRAY OF RECORD
            comp_id            STRING,
            ACTION             LIKE type_t.num5,
            standard           LIKE type_t.num5,
            including          LIKE type_t.num5
            END RECORD
   DEFINE gdnode_all           om.DomNode
   DEFINE g_select_define      STRING 
   DEFINE g_tmp                STRING 
   DEFINE g_report             LIKE type_t.chr20             ##170809-00051#2 add 報表資料庫
   DEFINE g_report_t    LIKE type_t.chr20 #MSSQL
   DEFINE g_subrep01_show      LIKE type_t.chr1              #240823-00003#1 add
   DEFINE g_subrep02_show      LIKE type_t.chr1              #240823-00003#1 add
   DEFINE g_subrep03_show      LIKE type_t.chr1              #240823-00003#1 add
   DEFINE g_subrep04_show      LIKE type_t.chr1              #240823-00003#1 add
END GLOBALS

DEFINE g_dzbb RECORD 
         prog_name             STRING,
         point_name            STRING,
         point_ver             STRING,
         addpoint              STRING
         END RECORD
        
DEFINE g_update                STRING
DEFINE g_master_chk            BOOLEAN
DEFINE g_detail_chk            BOOLEAN
DEFINE g_value                 STRING
DEFINE g_code_ide              LIKE type_t.chr1 #140617 add
DEFINE g_prog_id               LIKE dzfi_t.dzfi001 #140718 add
#DEFINE g_prog_ver              LIKE type_t.chr1 #140718 add  #150316 mark
DEFINE g_prog_ver              LIKE dzgc_t.dzgc002 #150316 add
DEFINE g_msg                   STRING               ##150429 add bpm未加上的欄位訊息
DEFINE g_sub_properties        om.SaxAttributes #230309-00010
DEFINE g_keep_value            STRING           #230309-00010
DEFINE g_keep_query            STRING           #230309-00010

MAIN
   DEFINE lddoc_all            om.DomDocument
   DEFINE ls_filename          STRING
   DEFINE ls_mod               STRING
   DEFINE ls_path              STRING
   DEFINE ls_cmd               STRING   # 140225 add
   DEFINE ls_name              STRING 
   DEFINE li_cnt               INTEGER                  #140804 add
   DEFINE lr_dzbc              RECORD LIKE dzbc_t.*     #140804 add
   DEFINE l_reptable           STRING       ##170809-00051#2 add 報表檔   
   DEFINE l_chk_rep            BOOLEAN      ##170809-00051#2 add 
   DEFINE ls_dbs               STRING #MSSQL


   DISPLAY "adzp161開始運作時間:",cl_get_current()  ##141201 add 
   LET ls_dbs=FGL_GETENV("ZONE"),"_ds" #MSSQL
   CALL cl_db_connect("ds",FALSE)  #140804 add

   #是否覆蓋原有add-point
   LET g_update = ARG_VAL(3)
   IF cl_null(g_update) THEN
      LET g_update = 'n'
   END IF
   
   #顯示外部參數訊息
   DISPLAY "程式模組:",ARG_VAL(1)
   DISPLAY "程式名稱:",ARG_VAL(2)
   DISPLAY "程式add-point是否覆蓋:",ARG_VAL(3)
   DISPLAY "程式版次:",ARG_VAL(4)
   DISPLAY "程式客製標示:",ARG_VAL(5)


   #先確認Section是否解開
   LET li_cnt = 0
   LET lr_dzbc.dzbc001 = ARG_VAL(2)
   LET lr_dzbc.dzbc002 = ARG_VAL(4)
   LET lr_dzbc.dzbc007 = ARG_VAL(5)
   
   SELECT COUNT(1) INTO li_cnt FROM dzbc_t
                  WHERE dzbc001 = lr_dzbc.dzbc001 AND #程式名稱
                        dzbc002 = lr_dzbc.dzbc002 AND #PR版次
                        dzbc005 <> 's'            AND #使用標示
                        dzbc007 = lr_dzbc.dzbc007

   #先判斷是否解開section, 進行解析與寫入   
   IF li_cnt > 0 THEN
      DISPLAY '該程式(',lr_dzbc.dzbc001,')已解開Section調整, 不再產生標準Section!'
   ELSE

       #判斷程式類型
       LET ls_mod = ARG_VAL(1)
       IF ls_mod = 'lib' OR
          ls_mod = 'sub' OR
          ls_mod = 'wss' THEN
          LET ls_path = "COM"
       ELSE
          LET ls_path = "ERP"
       END IF   

       LET g_prog_id = ARG_VAL(2)
       LET g_prog_ver = ARG_VAL(4)
       LET g_code_ide = ARG_VAL(5)
       
       #tabx檔更正為tab檔，組路徑
       LET ls_filename = os.Path.join(os.Path.join(FGL_GETENV(ls_path),ARG_VAL(1)),"dzx")
       LET ls_filename = os.Path.join(ls_filename,"tab")
       LET ls_filename = os.Path.join(ls_filename,ARG_VAL(2) CLIPPED||".tab")
     


       #定義轉換用SaxAttributes
       LET g_properties = om.SaxAttributes.create()
       CALL g_properties.addAttribute("component_name",ARG_VAL(2)) #app_id 程式代號
       CALL g_properties.addAttribute("app_id",ARG_VAL(2)) #app_id 程式代號  #140623
       CALL g_properties.addAttribute("module",ARG_VAL(1)) #module 

       #讀入tab檔
       LET lddoc_all = om.DomDocument.createFromXmlFile(ls_filename)  #om.DomDocument

       #若讀取不到資料代表tabx不存在
       TRY
          LET gdnode_all = lddoc_all.getDocumentElement()                #om.DomNode
          CALL g_component_list.clear()
       CATCH
          #DISPLAY "ERROR(1):讀取",ls_filename,"發生錯誤, 程式中止!"          #170123-00046#1 mark
          DISPLAY cl_getmsg_parm("adz-01166",g_lang,ls_filename.trim())   #170123-00046#1 add       
          EXIT PROGRAM
       END TRY

       LET gdnode_all = lddoc_all.getDocumentElement()                #om.DomNode
       CALL g_component_list.clear()
       
       #產生流程=> 讀取樣板檔案 / 同時間寫出(一進一出) / 產生器符號中斷,擺放切片或產生code
       #獲取程式基本資料 (含填寫待轉換的對照表)
       DISPLAY "報表程式產生器(adzp161):開始讀取tab並產生基礎資料!"
       CALL adzp161_read_basic_data()

      DISPLAY " before g_msg:",g_msg
      IF g_msg.getLength() > 0 THEN    
        DISPLAY "g_msg:",g_msg
      END IF 

       #開始進行產生/轉換動作
       DISPLAY "報表程式產生器(adzp161):開始讀取template並取代相關變數產生tgl!"
       CALL adzp161_read_and_replace()       

       #呼叫adzp175重新產生
       DISPLAY "--------------------------------------------------"
       LET ls_cmd = ' adzp175 ',ARG_VAL(1),' ',ARG_VAL(2)
       #CALL cl_cmdrun(ls_cmd)   #150722-00011 mark 
       CALL cl_cmdrun_wait(ls_cmd) #150722-00011 add

       #呼叫adzp176寫入section
       DISPLAY "--------------------------------------------------"
       DISPLAY "開始section解析, 執行adzp176"                                                              #程式版本         #使用標示
       LET ls_cmd = ' adzp176 ',ARG_VAL(1),' ',ARG_VAL(2),' ',ARG_VAL(4),' ',ARG_VAL(5)  #140616 add ,' ',ARG_VAL(4),' ',g_code_ide
       CALL cl_cmdrun_wait(ls_cmd)

   END IF  #140804 add
   #呼叫adzp177產出tgl
   DISPLAY "--------------------------------------------------"
   DISPLAY "開始產生tgl檔案, 執行adzp177"
   LET ls_cmd = ' adzp177 ',ARG_VAL(1),' ',ARG_VAL(2),' ',ARG_VAL(4),' ',ARG_VAL(5) #140616 add ,' ',ARG_VAL(4),' ',g_code_ide
   CALL cl_cmdrun_wait(ls_cmd)

   DISPLAY "--------------------------------------------------"
   IF g_tmp="k01" THEN #XR要檢查報表檔
      LET l_reptable = g_properties.getValue("component_name"),"_t"
      CALL adzp161_check_reptable(g_properties.getValue("g_temp_query")) RETURNING l_chk_rep
      IF l_chk_rep = FALSE THEN
         CALL adzp161_drop_tmptable(l_reptable)
         IF NOT adzp161_create_reptable(g_properties.getValue("g_temp_query"),1,l_reptable) THEN
            DISPLAY "create reptable err!"
         END IF
      END IF
   END IF
   
END MAIN

#+ 更新所有暫存版本的add-point
#PRIVATE FUNCTION adzp161_upd_addpoint()
   #DEFINE lr_dzbb              RECORD LIKE dzbb_t.*
   #DEFINE ls_dzbb003           LIKE dzbb_t.dzbb003
#
   #程式名稱
   #LET lr_dzbb.dzbb001 = g_properties.getValue("component_name")
   #
   #正式版本號
   #LET lr_dzbb.dzbb003 = g_properties.getValue("general_adp_ver")
   #
   #temp版本號
   #LET ls_dzbb003 = lr_dzbb.dzbb003, "t"
#
   #BEGIN WORK  
            #
   #UPDATE dzbb_t SET dzbb003 = lr_dzbb.dzbb003
      #WHERE dzbb001 = lr_dzbb.dzbb001 AND 
            #dzbb003 = ls_dzbb003
#
   #COMMIT WORK
            #
#END FUNCTION


#+ 取出程式基本資料 <assembly>
PRIVATE FUNCTION adzp161_read_basic_data()
   DEFINE ldnode_action        om.DomNode
   DEFINE ldnode_section       om.DomNode                 #段落
   DEFINE lnl_selted           om.NodeList
   DEFINE ls_err               STRING
   DEFINE ls_tmp               STRING
   DEFINE ls_type_list         STRING
   DEFINE ls_arg_str           STRING
   DEFINE ls_arg_cnt           LIKE type_t.num5 
   DEFINE i                    LIKE type_t.num5   
   DEFINE ldnode_action1       om.DomNode
   DEFINE lnl_selted1          om.NodeList
   DEFINE l_select_define      STRING 
   DEFINE ls_token             base.StringTokenizer
   DEFINE ls_token_str         STRING 
   DEFINE l_selct_define       STRING 
   DEFINE l_str_token_cnt      LIKE type_t.num5
   DEFINE ls_token_cnt         LIKE type_t.num5
   DEFINE l_sub_node           om.DomNode #230309-00010#1
   DEFINE l_tmp_str            STRING   #230309-00010#1
   DEFINE l_sub_seq            STRING   #230309-00010#1   
   DEFINE l_subrep_str         STRING   #230309-00010#1
   DEFINE l_subproc_str        STRING   #230309-00010#1
   DEFINE l_sub_remark_str     STRING   #230309-00010#1
   DEFINE l_sub_show           LIKE type_t.chr1   #240823-00003#1 add
   
   
   #取出om.NodeList 有關tag list
   LET lnl_selted = gdnode_all.selectByTagName("assembly")
   LET ldnode_action = lnl_selted.item(1) 

   #模組
   LET ls_tmp = ldnode_action.getAttribute("module") 
   CALL g_properties.addAttribute("general_module",ls_tmp)
   
   #TIPTOP版本
   LET ls_tmp = ldnode_action.getAttribute("tpver") 
   CALL g_properties.addAttribute("general_prog_ver",ls_tmp)

   #程式版本(PR) -140623 add-(S)
   CALL g_properties.addAttribute("general_pr_ver",ARG_VAL(4)) 
   ##140623 add-(e)
   
   #程式版本(SD)
   #LET ls_tmp = ldnode_action.getAttribute("sdver") 
   LET ls_tmp = ARG_VAL(4)
   IF cl_null(ls_tmp) THEN
      LET ls_tmp = "1"
   END IF   
   CALL g_properties.addAttribute("general_adp_ver",ls_tmp)

   #程式描述
   LET ls_tmp = ldnode_action.getAttribute("description") 
   CALL g_properties.addAttribute("general_title_desc",ls_tmp)
   
   #程式創造者
   LET ls_tmp = ldnode_action.getAttribute("crtid") 
   CALL g_properties.addAttribute("general_title_creator",ls_tmp)
   
   #程式創造時間
   LET ls_tmp = ldnode_action.getAttribute("crtdt") 
   CALL g_properties.addAttribute("general_title_create_time",ls_tmp)
   
   #程式修改者
   LET ls_tmp = ldnode_action.getAttribute("modid") 
   CALL g_properties.addAttribute("general_title_modifier",ls_tmp)
   
   #程式修改時間
   LET ls_tmp = ldnode_action.getAttribute("moddt") 
   CALL g_properties.addAttribute("general_title_modify_time",ls_tmp)
   
   #模組
   LET ls_tmp = ldnode_action.getAttribute("jobmode") 
   CALL g_properties.addAttribute("general_jobmode",ls_tmp)

   #行業別
   LET ls_tmp = ldnode_action.getAttribute("industry") 
   CALL g_properties.addAttribute("industry_id",ls_tmp)

   #form代碼
   LET ls_tmp = ldnode_action.getAttribute("form") 
   CALL g_properties.addAttribute("form_id",ls_tmp)

   #樣板代碼 
   #LET ls_type_list = "g01,x01"     #160921-00012#1 mark
   #LET ls_type_list = "g01,x01,x02"      #160921-00012#1 add
   #LET ls_type_list = "g01,x01,x02,f01,f02"   #170809-00051#2 add  #elena 220302 mark
   #LET ls_type_list = "g01,x01,x02,f01,f02,abi01"  #elena 220302 add
   LET ls_type_list = "g01,x01,x02,f01,f02,abi01,k01"  #230309-00010#1
   LET ls_tmp = ldnode_action.getAttribute("type")
   LET g_tmp = ls_tmp
   IF ls_type_list.getIndexOf(ls_tmp,1) = 0 THEN
      #DISPLAY "ERROR(2):目前尚未支持此種樣板檔(",ls_tmp,"), 請重新定義!"   #170123-00046#1 mark
      DISPLAY cl_getmsg_parm("adz-01167",g_lang,ls_tmp.trim())        #170123-00046#1 add
      EXIT PROGRAM
   END IF   
   #名義樣板名稱
   CALL g_properties.addAttribute("type_id_t",ls_tmp)
   
   #實際使用的樣板名稱
   CALL g_properties.addAttribute("type_id",ls_tmp)
   
   DISPLAY "目前使用的程式樣板為 : code_",ls_tmp,"樣板"

   #141201 add 程式樣板版次管控-(s) 
   #確認MD5是否正確(錯誤則中斷)
   #CALL sadzi100_md5_chk(g_properties.getValue("type_id"))                 #151020-00008 mark
   #CALL sadzi100_template_chk(g_properties.getValue("type_id"),cs_prog_ver) #151020-00008 add   #janet 161209 mark
   #141201 add 程式樣板版次管控-(e) 
      
   #取出om.NodeList 抓select define 內容,為了比對save資料那段
   LET lnl_selted1 = gdnode_all.selectByTagName("sql")
   LET ldnode_action1 = lnl_selted1.item(1) 
   LET l_select_define = ldnode_action1.getAttribute("query")
   LET ls_token = base.StringTokenizer.create(l_select_define.trim(), ',')
   LET ls_token_cnt = ls_token.countTokens()
   WHILE ls_token.hasMoreTokens()
       LET l_str_token_cnt = l_str_token_cnt + 1
       LET ls_token_str = ls_token.nextToken()
       IF ls_token_str.getIndexOf("(",1)>0 THEN
          LET ls_token_str = ls_token_str.subString(1, ls_token_str.getIndexOf("(",1)-1)
       END IF 
       LET g_select_define = g_select_define , ",", ls_token_str       
   END WHILE 

   #開始走訪XML
   #CONNECT TO "ds"   #140804 mark
  
   #230309-00010#1
   IF g_tmp = "k01" THEN           
      LET g_sub_properties = om.SaxAttributes.create()    
      #xr先合併子報表,指定sub_name
      LET lnl_selted1 = gdnode_all.selectByTagName("subrep")          
      IF lnl_selted1.getLength() > 0 THEN
         LET l_subrep_str = ""
         LET l_subproc_str = ""
         LET l_sub_remark_str = ""
         
         FOR i = 1 TO lnl_selted1.getLength()
             CALL g_sub_properties.clear()  #清空子報表屬性
             LET l_sub_node = lnl_selted1.item(i)
             
             #LET l_tmp_str = adzp161_xr_subtmp("subrep","d08",l_sub_node,"")   #240823-00003#1 mark
             #LET l_subrep_str = l_subrep_str , l_tmp_str   #240823-00003#1 mark
             LET l_sub_show = l_sub_node.getAttribute("show")   #240823-00003#1 add
             #DISPLAY l_subrep_str
            
             LET l_sub_seq = l_sub_node.getAttribute("subseq")          
             #備註subrep
             IF (l_sub_seq = "01" OR l_sub_seq = "02" OR l_sub_seq = "03" OR l_sub_seq = "04") THEN
                #LET l_tmp_str = adzp161_xr_subtmp("subrep_remark_proc","d10",l_sub_node,"")   #240823-00003#1 mark   
                #LET l_sub_remark_str = l_sub_remark_str , l_tmp_str   #240823-00003#1 mark 
                #240823-00003#1 add(s)
                IF l_sub_seq = "01" THEN 
                   LET g_subrep01_show = l_sub_show
                END IF
 
                IF l_sub_seq = "02" THEN 
                   LET g_subrep02_show = l_sub_show
                END IF
 
                IF l_sub_seq = "03" THEN 
                   LET g_subrep03_show = l_sub_show
                END IF
 
                IF l_sub_seq = "04" THEN 
                   LET g_subrep04_show = l_sub_show
                END IF

                IF l_sub_show != "N" THEN
                   LET l_tmp_str = adzp161_xr_subtmp("subrep","d08",l_sub_node,"")
                   LET l_subrep_str = l_subrep_str , l_tmp_str
                   LET l_tmp_str = adzp161_xr_subtmp("subrep_remark_proc","d10",l_sub_node,"")
                   LET l_sub_remark_str = l_sub_remark_str , l_tmp_str
                END IF
                #240823-00003#1 add(e)
             ELSE
                #240823-00003#1 add(s)
                LET l_tmp_str = adzp161_xr_subtmp("subrep","d08",l_sub_node,"")
                LET l_subrep_str = l_subrep_str , l_tmp_str
                #240823-00003#1 add(e)
                LET l_tmp_str = adzp161_xr_subtmp("subrep_proc","d07",l_sub_node,"")
                LET l_subproc_str = l_subproc_str , l_tmp_str
             END IF

         END FOR
      END IF   
      CALL g_properties.addAttribute("subrep_str", l_subrep_str)      
      CALL g_properties.addAttribute("subrep_proc", l_subproc_str)
      CALL g_properties.addAttribute("subrep_remark_proc", l_sub_remark_str)
       
      CALL adzp161_xr_xml_read(ldnode_action)
   ELSE 
      CALL adzp161_xml_search(ldnode_action)
   END IF #230309-00010#1

   #171011-00004#1 add-(s)
   #先行寫入  #ARG_VAL(5)客製碼
   LET g_dzbb.prog_name   = g_properties.getValue("app_id")
   LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
   LET g_dzbb.addpoint    = "#Memos"
   LET g_dzbb.point_name  = "global.memo"
   CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))
   #171011-00004#1 add-(e)
   
END FUNCTION


#+ 程式樣板讀取/程式寫出(一進一出) / 產生器符號中斷,擺放切片或產生code
PRIVATE FUNCTION adzp161_read_and_replace()
   DEFINE lchannel_read           base.Channel
   DEFINE lchannel_write          base.Channel
   DEFINE ls_readline             STRING 
   DEFINE ls_text                 STRING
   DEFINE ls_code_filename        STRING
   DEFINE ls_sample_filename      STRING
   DEFINE lchannel_check          base.Channel
   DEFINE ls_mdlpath              STRING
   DEFINE l_currentline           STRING            #240823-00003#1 add

   LET lchannel_read = base.Channel.create()
   LET lchannel_write = base.Channel.create()

   CALL lchannel_read.setDelimiter("")
   CALL lchannel_write.setDelimiter("")

    #先讀取T100TEMPLATEPATH, 若無則採用$ERP/mdl
    LET ls_mdlpath = FGL_GETENV("T100TEMPLATEPATH")
    IF cl_null(ls_mdlpath) THEN
       LET ls_mdlpath = FGL_GETENV("ERP")
       LET ls_mdlpath = os.Path.join(ls_mdlpath,"mdl")
    END IF
   
   #定義取用樣板檔案
   LET ls_sample_filename = "code_",DOWNSHIFT(g_properties.getValue("type_id")),".template"
   LET ls_sample_filename = os.Path.join(ls_mdlpath,ls_sample_filename)
   DISPLAY "樣板檔位置:",ls_sample_filename

   IF NOT os.Path.exists(ls_sample_filename) THEN
      #DISPLAY "ERROR(4):樣板檔案:",ls_sample_filename.trim()," 不存在!"      #170123-00046#1 mark
      DISPLAY cl_getmsg_parm("adz-01168",g_lang,ls_sample_filename.trim())  #170123-00046#1 add
      EXIT PROGRAM
   END IF
   CALL lchannel_read.openFile( ls_sample_filename CLIPPED, "r" )

   #產出程式路徑
   LET ls_code_filename = os.Path.join(FGL_GETENV(UPSHIFT(g_properties.getValue("module"))),"dzx")
   LET ls_code_filename = os.Path.join(ls_code_filename,"tgl")
   LET ls_code_filename = os.Path.join(ls_code_filename,ARG_VAL(2) CLIPPED||".tgx")  #副檔名改成tgx
   
   #先行移除tgl
   IF os.Path.delete(ls_code_filename) THEN
      #DISPLAY "刪除舊檔案:",ls_code_filename                    #170123-00046#1 mark
      DISPLAY cl_getmsg("adz-01164",g_lang),ls_code_filename   #170123-00046#1 add
   END IF
   
   #判斷是否砍除成功
   IF NOT os.Path.exists(ls_code_filename) THEN
      #DISPLAY "舊檔案刪除成功:",ls_code_filename                 #170123-00046#1 mark
      DISPLAY cl_getmsg("adz-01165",g_lang),ls_code_filename   #170123-00046#1 add
   ELSE
      #DISPLAY "Error:舊檔案刪除失敗:",ls_code_filename            #170123-00046#1 mark
      DISPLAY cl_getmsg("adz-01144",g_lang),ls_code_filename   #170123-00046#1 add
      EXIT PROGRAM
   END IF
   
   DISPLAY "產生檔位置:",ls_code_filename
   CALL lchannel_write.openFile( ls_code_filename, "w" )

   #產生程式版本及說明
   CALL lchannel_write.write(adzp161_prog_memo())

   #讀取及轉換
   WHILE TRUE
   
      LET ls_text = ""
      LET ls_readline = lchannel_read.readLine()

      ##141119 add-(s) 
      #讀取到版本資料時進行回寫gzzx
      IF ls_readline.getIndexOf('樣板自動產生',1) > 0 THEN
         CALL adzp161_update_gzzx(ls_readline)
      END IF
      ##141119 add-(e)
      
      #DISPLAY "ls_readline:",ls_readline
      IF lchannel_read.isEof() THEN
         EXIT WHILE
      END IF

      #define Record段落產生
      IF ls_readline.getIndexOf("#rep - Start -",1) > 0 THEN 
         CALL adzp161_make_records(lchannel_read,lchannel_write,0) RETURNING lchannel_read,lchannel_write
         LET ls_readline = ""
      ELSE 
         IF ls_readline.getIndexOf("#xgprep - Start -",1) > 0 THEN 
             CALL adzp161_make_records(lchannel_read,lchannel_write,0) RETURNING lchannel_read,lchannel_write
             LET ls_readline = "" 
         END IF 
      END IF 
      
      #行代換/ 對 ${} 置換
      IF ls_readline.getIndexOf("${",1) AND 
         ( ls_readline.getIndexOf("${",1) < ls_readline.getIndexOf("}",1) ) THEN
         LET ls_text = adzp161_line_replace(ls_readline)
      END IF

      #一般未進行任何處理段落產出
      IF ls_text IS NULL THEN
         LET ls_text = ls_readline
      END IF

      #240823-00003#1 add(s)
      LET l_currentline = ls_readline.trim()
      IF g_subrep01_show = "N" AND l_currentline = "TYPE sr_r RECORD ${remark_fields_define} END RECORD"  THEN         
         LET ls_text = ""
      END IF
      IF g_subrep01_show = "N" AND l_currentline = "CALL ${component_name}_subrep01_ins_data()"  THEN         
         LET ls_text = ""
      END IF
      IF g_subrep02_show = "N" AND l_currentline = "CALL ${component_name}_subrep02_ins_data()"  THEN 
         LET ls_text = ""
      END IF
      IF g_subrep03_show = "N" AND l_currentline = "CALL ${component_name}_subrep03_ins_data()"  THEN 
         LET ls_text = ""
      END IF
      IF g_subrep04_show = "N" AND l_currentline = "CALL ${component_name}_subrep04_ins_data()"  THEN 
         LET ls_text = ""
      END IF
      #240823-00003#1 add(e)

      CALL lchannel_write.write(ls_text)

   END WHILE 

   CALL lchannel_read.close()
   CALL lchannel_write.close()
   
   #對產生器寫出的檔案權限在UNIX下均全部打開
   IF os.Path.separator() = "/" THEN
      IF os.Path.chrwx(ls_code_filename,511) THEN
      END IF
   END IF
   ##更新add-point
   CALL cl_add_point_update(g_dzbb.prog_name,g_dzbb.point_ver)
   
END FUNCTION


#+ 產生程式版本及說明
PRIVATE FUNCTION adzp161_prog_memo()
   DEFINE ls_text              STRING
   DEFINE ls_buildtime         STRING
   DEFINE li_buildtime         INTEGER
   DEFINE ls_value             STRING
  
   
   LET li_buildtime = adzp161_prog_buildtime()
   LET ls_buildtime = li_buildtime USING "&&&&&&"
    
   #產生次數 
   CALL g_properties.addAttribute("mdl_buildtime", ls_buildtime)
   
   #程式代號
   CALL g_properties.addAttribute("mdl_filename",ARG_VAL(2))
   
   #程式描述
   LET ls_value = g_properties.getValue("general_title_desc")
   CALL g_properties.addAttribute("mdl_desc",ls_value)
   
   #程式創造者
   LET ls_value = g_properties.getValue("general_title_creator")
   CALL g_properties.addAttribute("mdl_creator",ls_value)
   
   #程式創造時間
   LET ls_value = g_properties.getValue("general_title_create_time")
   CALL g_properties.addAttribute("mdl_create_time",ls_value)
   
   #程式修改者
   LET ls_value = g_properties.getValue("general_title_modifier")
   CALL g_properties.addAttribute("mdl_modifier",ls_value)
   
   #程式修改時間
   LET ls_value = g_properties.getValue("general_title_modify_time")
   CALL g_properties.addAttribute("mdl_modify_time",ls_value)
   
   #樣板名稱
   CALL g_properties.addAttribute("mdl_id",g_properties.getValue("type_id"))

   LET ls_text = adzp161_make_slice("a00")
   
   RETURN ls_text      
  
END FUNCTION


#+ 取出版本建置次數
PRIVATE FUNCTION adzp161_prog_buildtime()
   DEFINE lc_dzaz001           LIKE dzaz_t.dzaz001
   DEFINE li_dzaz002           LIKE dzaz_t.dzaz002
   DEFINE lc_dzaz003           LIKE dzaz_t.dzaz003
   DEFINE lc_dzaz004           DATETIME YEAR TO SECOND

   LET lc_dzaz001 = ARG_VAL(2)
   LET li_dzaz002 = 0
   
   SELECT MAX(dzaz002) INTO li_dzaz002 FROM dzaz_t WHERE dzaz001 = lc_dzaz001
   
   IF STATUS OR li_dzaz002 IS NULL THEN
      LET li_dzaz002 = 0
   END IF
   
   LET lc_dzaz003 = FGL_GETENV("LOGNAME")
   LET lc_dzaz004 = cl_get_current()

   INSERT INTO dzaz_t(dzaz001,dzaz002,dzaz003,dzaz004)
      VALUES(lc_dzaz001,li_dzaz002+1,lc_dzaz003,lc_dzaz004)

   RETURN li_dzaz002 USING "&&&&"
   
END FUNCTION

#+ 走訪XR XML
PRIVATE FUNCTION adzp161_xr_xml_read(ln_n)
   DEFINE ln_n                 om.DomNode
   DEFINE ls_tmp               STRING
   DEFINE l_i                  LIKE type_t.num5
   DEFINE l_define_str   DYNAMIC ARRAY OF STRING
   DEFINE l_son_node           om.DomNode 
   DEFINE l_seq,l_tmp          STRING 
   DEFINE l_block              STRING
   DEFINE l_str                STRING 
   DEFINE l_rep_rec_cnt        INTEGER 
   DEFINE l_id                 STRING 
   DEFINE l_query              STRING 
   DEFINE ls_sql_name          STRING 
   DEFINE l_sql_start          STRING
   DEFINE ls_arg_token         base.StringTokenizer
   DEFINE ls_token_str         STRING 
   DEFINE l_arg                STRING 
   DEFINE l_arg_tmp            STRING   
   DEFINE ls_rep_define_str    STRING
   DEFINE l_maintable          STRING

   DEFINE l_parent             STRING
   DEFINE l_alias_name         STRING   #240112 add
   DEFINE l_alias              STRING   #240112 add
   DEFINE ls_bmp_key           STRING
   
   

   #取得第一個子節點
   LET ln_n = ln_n.getFirstChild()
   WHILE ln_n IS NOT NULL

       LET l_rep_rec_cnt = 0
       #傳至元件參數整理
       LET l_parent = ln_n.getTagName()
       IF ln_n.getTagName() = "define" THEN
          CALL adzp161_arg_fields_define(ln_n.getAttribute("arg"))        
       END IF 
       #DISPLAY "ln_n_node name:",ln_n.getAttributeName(1)
       FOR l_i = 1 TO ln_n.getChildCount()
           LET l_son_node = ln_n.getChildByIndex(l_i)
           CASE l_son_node.gettagname()
               WHEN "var" 
                   LET l_seq = l_son_node.getAttribute("seq")
                   LET l_tmp = l_son_node.getAttribute("value")
                   LET l_block = l_son_node.getAttribute("blk")
                   CASE 
                     WHEN l_seq = 1
                        LET l_block = 'h'
                     WHEN l_seq = 2
                        LET l_block = 'd'
                   END CASE
                   
                   #組record refer的值
                   CALL adzp161_xr_fields_define("var",l_seq,l_tmp,l_block)
                   
               WHEN "sql"
                   LET l_id = l_son_node.getAttribute("id")
                   LET l_query = l_son_node.getAttribute("query")
                   #SQL的值存入
                   #IF l_query IS NOT NULL AND NOT cl_null(l_query) THEN 
                       CASE l_id
                           WHEN "g_select"
                                #在此function組record與select子句
                                CALL adzp161_xr_fields_define("sql",l_id,l_query,l_block)                               
                                
                           WHEN "g_from"
                                LET l_sql_start = " FROM "                                      
                                LET l_query = " LET g_from = \"",l_sql_start,l_query ,"\""
                                LET ls_sql_name = l_id, "_query"
                                CALL g_properties.addAttribute(ls_sql_name, l_query)
                                
                           WHEN "g_where"  
                                LET l_maintable = l_son_node.getAttribute("maintable")
                                CALL g_properties.addAttribute("maintable", l_maintable)                                   
                                LET l_sql_start = " WHERE "                                
                                LET l_query = " LET g_where = \"",l_sql_start,l_query ,"\""
                                LET ls_sql_name = l_id, "_query"
                                CALL g_properties.addAttribute(ls_sql_name, l_query)    
                                #240118-00026 -S
                                LET l_alias = l_son_node.getAttribute("alias")
                                IF l_alias="Y" THEN
                                   LET l_alias_name = l_maintable , "." , l_maintable.subString(1,4)
                                   LET l_query = " LET g_where = cl_sql_replace_str(g_where,'",l_maintable.subString(1,4),"','", l_alias_name ,"')"
                                   DISPLAY "master_alias:" ,l_query
                                   CALL g_properties.addAttribute("g_where_filter", l_query)
                                ELSE
                                   CALL g_properties.addAttribute("g_where_filter", "")
                                END IF
                                #240118-00026 -E
								#240620-00031 -S
                                CALL adzp161_combi_bmpkey(l_maintable) RETURNING ls_bmp_key
                                LET ls_bmp_key = "LET g_doc_key = ",ls_bmp_key
                                CALL g_properties.addAttribute("g_bmp_key", ls_bmp_key)
								#240620-00031 -E
                                
                           WHEN "g_body_select"
                                CALL adzp161_xr_fields_define("sql",l_id,l_query,l_block)
                           
                           WHEN "g_body_from"
                                LET l_sql_start = " FROM "                                      
                                LET l_query = " LET g_from = \"",l_sql_start,l_query ,"\""
                                LET ls_sql_name = l_id, "_query"
                                CALL g_properties.addAttribute(ls_sql_name, l_query)
                                
                           WHEN "g_body_where"  
                                LET l_maintable = l_son_node.getAttribute("bodytable")
                                CALL g_properties.addAttribute("bodytable", l_maintable)
                                #IF cl_null(l_query) THEN                                
                                #   LET l_query = " 1=1"
                                #END IF
                                
                                LET l_sql_start = " WHERE "                                
                                LET l_query = " LET g_where = \"",l_sql_start,l_query ,"\""
                                LET ls_sql_name = l_id, "_query"
                                CALL g_properties.addAttribute(ls_sql_name, l_query)
                                #240118-00026 -S
                                LET l_alias = l_son_node.getAttribute("alias")
                                IF l_alias="Y" THEN
                                   LET l_alias_name = l_maintable , "." , l_maintable.subString(1,4)
                                   LET l_query = " LET g_where = cl_sql_replace_str(g_where,'",l_maintable.subString(1,4),"','", l_alias_name ,"')"
                                   DISPLAY "body_alias:" ,l_query
                                   CALL g_properties.addAttribute("g_body_where_filter", l_query)
                                ELSE
                                   CALL g_properties.addAttribute("g_body_where_filter", "")
                                END IF
                                
                           OTHERWISE
                                
                       END CASE 
                   #ELSE       
                        
                   #END IF
           END CASE 
       END FOR  
        
      LET ln_n = ln_n.getNext()  #同一層 next sibling 
      
   END WHILE 
   
END FUNCTION

#SUB KEY的分解
PRIVATE FUNCTION adzp161_xr_get_subkey(p_key,p_where,p_block)
   DEFINE p_key      STRING
   DEFINE p_where    STRING
   DEFINE p_block    STRING
   DEFINE ls_prefix  STRING
   DEFINE ls_wc      STRING
   DEFINE l_token1   base.StringTokenizer
   DEFINE l_token2   base.StringTokenizer
   DEFINE ls_key     STRING
   DEFINE ls_value   STRING
   DEFINE l_i        INTEGER
   
   IF p_block = "h" THEN
      LET ls_prefix = "sr_h."
   ELSE  
      LET ls_prefix = "sr_d."
   END IF

   LET l_token1 = base.StringTokenizer.create(p_key, ",")
   LET l_token2 = base.StringTokenizer.create(p_where, ",")
   LET ls_wc = ""
   
   LET l_i = 1   
   WHILE l_token1.hasMoreTokens()
      LET ls_key = l_token1.nextToken()
      LET ls_value = l_token2.nextToken()

      IF l_i = 1 THEN
         LET ls_wc = " AND " ,ls_key ,"='\",", ls_prefix, ls_value , ",\"'\"" 
      ELSE
         LET ls_wc = ls_wc,",\" AND " ,ls_key ,"='\",", ls_prefix, ls_value , ",\"'\""      
      END IF
      LET l_i = l_i + 1   
   END WHILE

   DISPLAY ls_wc
   RETURN ls_wc
END FUNCTION

#+ 走訪XR SUB XML
FUNCTION adzp161_xr_read_subxml(ln_n,p_id)
   DEFINE ln_n                 om.DomNode
   DEFINE p_id                 LIKE type_t.num5
   DEFINE l_sub_id             INTEGER
   DEFINE ls_tmp               STRING
   DEFINE l_i                  LIKE type_t.num5
   DEFINE l_define_str   DYNAMIC ARRAY OF STRING
   DEFINE l_son_node           om.DomNode 
   DEFINE l_seq,l_tmp          STRING 
   DEFINE l_str                STRING 
   DEFINE l_rep_rec_cnt        INTEGER 
   DEFINE l_id                 STRING 
   DEFINE l_query              STRING 
   DEFINE ls_sql_name          STRING 
   DEFINE l_sql_start          STRING
   DEFINE ls_arg_token         base.StringTokenizer
   DEFINE ls_token_str         STRING  

   
   DEFINE l_maintable          STRING
   DEFINE l_pk                 STRING
   DEFINE l_where              STRING
   DEFINE l_param              STRING
   DEFINE l_value              STRING
   DEFINE l_parent             STRING
   DEFINE l_block              STRING
   DEFINE l_alias_name         STRING   #240118-00026 add
   DEFINE l_alias              STRING   #240118-00026 add
   

   #取得第一個子節點 returns the first child node 
   #LET ln_n = ln_n.getFirstChild()
   
   WHILE ln_n IS NOT NULL

       LET l_rep_rec_cnt = 0       
       #LET l_parent = ln_n.getTagName()
       LET l_sub_id = ln_n.getAttribute("subseq")
       IF l_sub_id = p_id THEN     
          #DISPLAY "ln_n_node name:",ln_n.getAttributeName(1)
          FOR l_i = 1 TO ln_n.getChildCount()
            LET l_son_node = ln_n.getChildByIndex(l_i)
            CASE l_son_node.gettagname()
               WHEN "var" 
                   LET l_seq = l_son_node.getAttribute("seq")
                   LET l_tmp = l_son_node.getAttribute("value") 
                   LET l_block = l_son_node.getAttribute("blk")
                   
                   CASE
                      WHEN l_block="s1"
                         IF NOT cl_null(l_tmp) THEN                              
                            LET g_keep_value = l_tmp                        
                         END IF
                      WHEN l_block="s2" OR l_block="s3" OR l_block="s4"
                         LET l_tmp = g_keep_value
                   END CASE            
                   CALL adzp161_xr_fields_define("var",l_seq,l_tmp,l_block)                   

               WHEN "sql"
                   LET l_id = l_son_node.getAttribute("id")
                   LET l_query = l_son_node.getAttribute("query")
                   LET l_block = l_son_node.getAttribute("blk")
                   #SQL的值存入
                   #IF l_query IS NOT NULL AND NOT cl_null(l_query) THEN 
                   CASE l_id                                
                       WHEN "g_sub_select"
                            
                            CASE
                               WHEN l_block="s1"
                                  IF NOT cl_null(l_query) THEN                             
                                     LET g_keep_query = l_query                         
                                  END IF                                  
                               WHEN l_block="s2" OR l_block="s3" OR l_block="s4"
                                  LET l_query = g_keep_query
                            END CASE
                            
                            CALL adzp161_xr_fields_define("sql",l_id,l_query,l_block)
                       
                       WHEN "g_sub_from"
                            LET l_sql_start = " FROM "                          
                            LET l_query = " LET g_from = \"",l_sql_start,l_query ,"\""
                            LET ls_sql_name = "sub_from_query"
                            CALL g_sub_properties.addAttribute(ls_sql_name, l_query)
                            
                       WHEN "g_sub_where"  
                            LET l_maintable = l_son_node.getAttribute("subtable")
                            
                            IF l_sub_id<5 THEN
                               LET l_pk = l_son_node.getAttribute("pk")
                               LET l_param = l_son_node.getAttribute("param")
                               CALL g_sub_properties.addAttribute("sub_param" ,l_param)
                               
                               {IF l_sub_id=1 OR l_sub_id=4 THEN
                                  LET l_where = adzp161_xr_get_subkey(l_pk,l_param,"h")   
                               ELSE
                                  LET l_where = adzp161_xr_get_subkey(l_pk,l_param,"d")   
                               END IF}
                            END IF
                            
                            CALL g_sub_properties.addAttribute("subtable", l_maintable)                                   
                            #IF cl_null(l_query) THEN
                            #   LET l_query = " 1=1"
                            #END IF
                            
                            LET l_sql_start = " WHERE "                               
                            LET l_query = " LET g_where = \"",l_sql_start,l_query ,"\""

                            LET ls_sql_name = "sub_where_query"
                            CALL g_sub_properties.addAttribute(ls_sql_name, l_query)
                            #240118-00026 -S
                            LET l_alias = l_son_node.getAttribute("alias")
                            IF l_alias="Y" THEN
                               LET l_alias_name = l_maintable , "." , l_maintable.subString(1,4)
                               LET l_query = " LET g_where = cl_sql_replace_str(g_where,'",l_maintable,"','", l_alias_name ,"')"
                               DISPLAY "sub_alias:" ,l_query
                               CALL g_properties.addAttribute("g_sub_filter", l_query)
                            ELSE
                               CALL g_properties.addAttribute("g_sub_filter", "")
                            END IF
                            #240118-00026 -E                            
                       OTHERWISE
                            
                   END CASE
            END CASE 
          END FOR
       END IF        
      LET ln_n = ln_n.getNext()  #同一層 next sibling 
      
   END WHILE 
   
END FUNCTION

#+ 走訪XML
PRIVATE FUNCTION adzp161_xml_search(ln_n)
   DEFINE ln_n                 om.DomNode
   DEFINE ls_tmp               STRING
   DEFINE l_i                  LIKE type_t.num5
   DEFINE l_define_str   DYNAMIC ARRAY OF STRING
   DEFINE l_son_node           om.DomNode 
   DEFINE l_seq,l_tmp          STRING 
   DEFINE l_str                STRING 
   DEFINE l_rep_rec_cnt        INTEGER 
   DEFINE l_id                 STRING 
   DEFINE l_query              STRING 
   DEFINE ls_sql_name          STRING 
   DEFINE l_sql_start          STRING
   DEFINE ls_arg_token         base.StringTokenizer
   DEFINE ls_token_str         STRING 
   DEFINE l_arg                STRING 
   DEFINE l_arg_tmp            STRING   
   DEFINE ls_rep_define_str    STRING
   DEFINE l_maintable          STRING  #140707 add
   DEFINE ls_bmp_key           STRING                ##150429 add
   
   
   

   #取得第一個子節點 returns the first child node
 
   LET ln_n = ln_n.getFirstChild()
   WHILE ln_n IS NOT NULL

       LET l_rep_rec_cnt = 0
       #傳至元件參數整理    
       IF ln_n.getTagName() = "define" THEN
          CALL adzp161_arg_fields_define(ln_n.getAttribute("arg"))        
       END IF 
       #DISPLAY "ln_n_node name:",ln_n.getAttributeName(1)
       FOR l_i = 1 TO ln_n.getChildCount()
           LET l_son_node = ln_n.getChildByIndex(l_i)
           CASE l_son_node.gettagname()
               WHEN "var" 
                   LET l_seq = l_son_node.getAttribute("seq")
                   LET l_tmp = l_son_node.getAttribute("value") 
                   #組record refer的值
                   CALL adzp161_fields_define("var",l_seq,l_tmp)
                   #rep段定義
                   LET ls_rep_define_str = ls_rep_define_str , "DEFINE sr",l_seq USING "<<<" ,
                       " sr",l_seq USING "<<<","_r"                    
                   IF l_i <> ln_n.getChildCount() THEN
                      LET ls_rep_define_str = ls_rep_define_str ,"\n"
                   END IF 
                   LET l_rep_rec_cnt = l_rep_rec_cnt + 1                #計算type record數

               WHEN "sql"
                   LET l_id = l_son_node.getAttribute("id")
                   LET l_query = l_son_node.getAttribute("query")
                   #SQL的值存入
                   IF l_query IS NOT NULL AND NOT cl_null(l_query) THEN 
                       CASE l_id
                           WHEN "g_select"
                                #在此function組record與select子句
                                CALL adzp161_fields_define("sql",l_id,l_query)
             
                           OTHERWISE
                                IF l_id ="g_from" THEN LET l_sql_start = " FROM " END IF
                                IF l_id ="g_where" THEN 
                                    #140707 add主表 -(s)
                                    LET l_maintable = l_son_node.getAttribute("maintable")
                                    CALL g_properties.addAttribute("maintable", l_maintable)
                                    #140707 add主表 -(e)
                                    LET l_sql_start = " WHERE " 
                                END IF
                                IF l_id ="g_order" THEN LET l_sql_start = " ORDER BY " END IF    
                                LET l_query = " LET ",l_id, " = \"",l_sql_start,l_query ,"\""
                                LET ls_sql_name = l_id, "_query"
                                #160615-00007#1 add -(s)
                                IF l_id = "g_from" THEN
                                   LET l_query = sadzp188_handle_dzgb011_dlang(l_query)
                                   #DISPLAY "l_query:",l_query
                                END IF 
                                #160615-00007#1 add -(e)
                               CALL g_properties.addAttribute(ls_sql_name, l_query)
                                ##150429 增加bpm的key值 add-(s)
                               ##取出主table的pk值
                               CALL adzp161_combi_bmpkey(l_maintable) RETURNING ls_bmp_key
                               CALL g_properties.addAttribute("group_bmp_key", ls_bmp_key)
                               ##150429 增加bpm的key值 add-(e)
                       END CASE 
                   ELSE        
                        #140118 for demo測試加-(s)
                        IF l_id ="g_where" THEN 
                           LET l_sql_start = " WHERE " 
                           LET l_query = " LET ",l_id, " = \"",l_sql_start,l_query ,"\""
                           LET ls_sql_name = l_id, "_query"                        
                        #140118 for demo測試加-(e)
                        ELSE 
                           LET ls_sql_name = l_id, "_query"
                        END IF 
                        CALL g_properties.addAttribute(ls_sql_name, l_query) 
                        #140707 add主表 -(s)
                        IF l_id ="g_where" THEN 
                            LET l_maintable = l_son_node.getAttribute("maintable")
                            CALL g_properties.addAttribute("maintable", l_maintable)
                        END IF 
                        #140707 add主表 -(e)   
                       ##150429 增加bpm的key值 add-(s)
                       ##取出主table的pk值
                       CALL adzp161_combi_bmpkey(l_maintable) RETURNING ls_bmp_key
                       CALL g_properties.addAttribute("group_bmp_key", ls_bmp_key)
                       ##150429 增加bpm的key值 add-(e)                        
                   END IF 

               WHEN "section"    #mainrep             
                   CALL adzp161_mainrep_prep("section",l_son_node)

               WHEN "subreptag"  #sub-report             
                   CALL adzp161_subreport_prep("subsection",ln_n) 

           END CASE 
       END FOR  
       IF ln_n.getChildCount()>0 THEN 
           IF l_son_node.gettagname() = "var" THEN 
              CALL g_properties.addAttribute("rep_reocrd_cnt", l_rep_rec_cnt)
                   
              CALL g_properties.addAttribute("rep_record_define", ls_rep_define_str.trimRight())
           END IF 
       END IF 
      LET ln_n = ln_n.getNext()  #同一層 next sibling 
      
   END WHILE 
   
END FUNCTION


#+ 移除saxAttribute內的準備代換元素
PRIVATE FUNCTION adzp161_remove_attribute(ls_propertyid)
   DEFINE ls_propertyid        STRING
   DEFINE li_cnt               LIKE type_t.num10

   FOR li_cnt = 1 to g_properties.getLength()
      IF ls_propertyid.equals(g_properties.getName(li_cnt)) THEN
         CALL g_properties.removeAttribute(li_cnt)
         EXIT FOR
      END IF
   END FOR

END FUNCTION



#+ 判斷該欄位參照的table名稱
PUBLIC FUNCTION adzp161_define_table_name(ps_field)
   DEFINE ps_field             STRING
   #sqlserver mark(s)
#   DEFINE ls_field             CHAR(100)
#   DEFINE ls_table             CHAR(100)
   #sqlserver mark(e)
   DEFINE ls_return            STRING
#   DEFINE ls_prefix            CHAR(100)   #sqlserver mark
   #sqlserver add(s)
   DEFINE ls_field             VARCHAR(100)
   DEFINE ls_table             VARCHAR(100)
   DEFINE ls_prefix            VARCHAR(100)
   #sqlserver add(e)
   DEFINE ls_dbs               STRING
   DEFINE ls_sql STRING
   DEFINE l_cnt  INTEGER
   
   #取table前四碼
   LET ls_prefix = '%_t'
   LET ls_field  = ps_field
   #sqlserver add(s) 
   #DISPLAY "db:",g_dbs #Hiko mark
   #DISPLAY "ls_prefix:",ls_prefix
   #DISPLAY "ls_field:",ls_field
   #sqlserver add(e)
   
   #改由gztz進行比對
   LET ls_table = ""
   SELECT gztz001 INTO ls_table FROM gztz_t 
   WHERE gztz002 = ls_field AND gztz001 LIKE ls_prefix

   IF cl_null(ls_table) THEN
      #DISPLAY "ERROR(10):欄位",ls_field CLIPPED, "不存在於資料庫中, 請重新檢查!"     #170123-00046#1 mark
      DISPLAY cl_getmsg_parm("adz-01169",g_lang,ls_field)                        #170123-00046#1 add
   END IF 

   LET ls_return = ls_table CLIPPED, '.'
   
   RETURN ls_return
   
END FUNCTION



#組標準子報表樣內容
PRIVATE FUNCTION adzp161_subreport_prep(p_tag,p_node)
   DEFINE p_tag                STRING 
   DEFINE p_node               om.DomNode
   DEFINE ls_son_node          om.DomNode
   DEFINE i                    INTEGER 
   DEFINE l_subreport          STRING 
   DEFINE l_subreport_tmp      STRING
   DEFINE l_sub_define         STRING 

   
   LET l_sub_define = ""   
   FOR i = 1 TO p_node.getChildCount()
       LET ls_son_node = p_node.getChildByIndex(i)
       CALL adzp161_combi_subtmp(ls_son_node.getTagName(),"d02",ls_son_node,"") RETURNING l_subreport
       LET l_subreport_tmp = l_subreport_tmp ,l_subreport,"\n"

        ##140508 子報表顯示變數收集(s)       
       IF i = 1 THEN 
          LET l_sub_define = "DEFINE l_subrep",ls_son_node.getAttribute("id"),"_show  LIKE type_t.chr1"
       ELSE 
          LET l_sub_define = l_sub_define,",","\n", 7 SPACES,"l_subrep",ls_son_node.getAttribute("id"),"_show  LIKE type_t.chr1"
       END IF           
       ##140508 子報表顯示變數收集(e)      
   END FOR 
   CALL g_properties.addAttribute("subrep_str",l_subreport_tmp)
   ##140508子報表變數定義
   CALL g_properties.addAttribute("subrep_var_define", l_sub_define.trimRight()) 
END FUNCTION 

#主報表段準備
PRIVATE FUNCTION adzp161_mainrep_prep(p_tag,p_node)
   DEFINE p_tag                STRING 
   DEFINE p_node               om.DomNode
   DEFINE i                    INTEGER 
   DEFINE ls_node              om.DomNode
   DEFINE ls_name              STRING 
   DEFINE ls_type              STRING
   DEFINE ls_subtype           STRING 
   DEFINE ls_token             base.StringTokenizer
   DEFINE ls_token_str         STRING 
   DEFINE ls_rep_order_var1    STRING 
   DEFINE l_str_token_cnt      LIKE type_t.num5
   DEFINE ls_token_cnt         LIKE type_t.num5   

   
       CASE p_node.getAttribute("id")

           WHEN "repOrder"
               LET ls_type = p_node.getAttribute("type")
               LET ls_name = p_node.getAttribute("name")
               
               IF NOT cl_null(ls_type) AND ls_type IS NOT NULL THEN
                  CALL g_properties.addAttribute("rep_order_type", ls_type)
               END IF 
               IF NOT cl_null(ls_name) AND ls_name IS NOT NULL THEN
                  CALL g_properties.addAttribute("rep_order_var", ls_name) 
                  LET ls_token = base.StringTokenizer.create(ls_name.trim(), ',')
                  LET ls_token_cnt = ls_token.countTokens()
                  WHILE ls_token.hasMoreTokens()
                     LET l_str_token_cnt = l_str_token_cnt + 1
                     LET ls_token_str = ls_token.nextToken()
                     LET ls_rep_order_var1 = ls_rep_order_var1 ,"sr1.",ls_token_str  
                     IF l_str_token_cnt != ls_token_cnt THEN
                        LET ls_rep_order_var1 = ls_rep_order_var1, ","
                     END IF                                    
                  END WHILE  
                  CALL g_properties.addAttribute("rep_order_var1", ls_rep_order_var1)                  
               END IF  

           WHEN "b_group"   
               #CALL adzp161_rep_signstr_prep(p_node,p_node.getAttribute("id"))          #組GROUP段簽核內容   
               CALL adzp161_rep_block_sub_prep(p_node,p_node.getAttribute("id"))         #組GROUP段子報表內容       
               CALL adzp161_rep_block_prep(p_node,p_node.getAttribute("id"))  

           WHEN "everyrow"            
               CALL adzp161_rep_block_sub_prep(p_node,p_node.getAttribute("id"))    
              
           WHEN "a_group"            
               CALL adzp161_rep_block_sub_prep(p_node,p_node.getAttribute("id"))         #組GROUP段內容
               CALL adzp161_rep_block_prep(p_node,p_node.getAttribute("id")) 

           #WHEN "execute_str"
           #    CALL adzp161_xg_execute_str_prep(p_node)
       END CASE 

END FUNCTION  


#報表段每個block的處理準備(b_group' a_gruop讀入子樣板檔)
PRIVATE FUNCTION adzp161_rep_block_prep(p_node,p_id)
   DEFINE p_node                 om.DomNode
   DEFINE p_id                   STRING 
   DEFINE ls_order_var           STRING 
   DEFINE ls_token               base.StringTokenizer
   DEFINE ls_token_str           STRING 
   DEFINE ls_tmp                 STRING 
   DEFINE ls_var                 STRING
   DEFINE l_block_str            STRING
   DEFINE l_block_tmp            STRING 
   DEFINE ls_block_name          STRING  
   DEFINE ls_token_cnt           LIKE type_t.num5 
   DEFINE ls_cnt                 LIKE type_t.num5 
   
    
   LET ls_tmp = p_node.getAttribute("reptype")
   LET ls_var = g_properties.getValue("rep_order_var")   
   LET ls_token = base.StringTokenizer.create(ls_var.trim(), ',')
   LET ls_token_cnt = ls_token.countTokens()

   WHILE ls_token.hasMoreTokens()
      LET ls_token_str = ls_token.nextToken()
      #讀取子樣板組合 filed搭配block內容
      CALL adzp161_combi_subtmp(p_id,ls_tmp,p_node,ls_token_str) RETURNING l_block_str
      IF ls_token_cnt <> ls_cnt THEN 
         LET l_block_tmp =l_block_tmp,l_block_str,"\n"
      ELSE
         LET l_block_tmp =l_block_tmp,l_block_str
      END IF    
     
   END WHILE    
   
   LET ls_block_name = "rep_" ,p_id,"_str" 
   CALL g_properties.addAttribute(ls_block_name,l_block_tmp)  
    
END FUNCTION 




#報表段每個block的子報表處理準備
PRIVATE FUNCTION adzp161_rep_block_sub_prep(p_node,p_id)
   DEFINE p_node                om.DomNode
   DEFINE p_id                  STRING 
   DEFINE ls_child_node         om.DomNode
   DEFINE ls_type               STRING
   DEFINE ls_type_t             STRING        #舊值
   DEFINE ls_subtype            STRING 
   DEFINE ls_subseq             STRING 
   DEFINE ls_subtmp             STRING  
   DEFINE l_sub_str             STRING 
   DEFINE ls_tag                STRING 
   DEFINE ls_block_type         STRING 
   DEFINE ls_reptype            STRING 
   DEFINE i                     INTEGER 
   DEFINE l_dzgg006             LIKE dzgg_t.dzgg006   ##是否簽核   ##140422
   DEFINE l_prog_id             LIKE dzgg_t.dzgg001
   DEFINE l_sd_ver              LIKE dzgg_t.dzgg002
   

       LET ls_subtmp = ""                       #子報表暫存字串 
       LET ls_type_t = ""
       LET ls_reptype = p_node.getAttribute("reptype") 
       FOR i = 1 TO p_node.getChildCount()
           LET ls_child_node = p_node.getChildByIndex(i)
           #xml的tagname，要存入add-point變數
           LET ls_tag = ls_child_node.getTagName()
           CALL g_properties.addAttribute("main_tag",ls_tag)
           LET ls_type = ls_child_node.getAttribute("type")   ##群組欄位
           #子樣板
           LET ls_subtype = ls_child_node.getAttribute("subtype")  

           ##140416 -(S)
           #組簽核樣板內容
           IF p_id = "b_group" AND i = 1 THEN 
              LET l_prog_id = g_properties.getValue("component_name")
              LET l_sd_ver = g_properties.getValue("general_adp_ver")
              ##否要加簽核140422 -(s)              
              SELECT dzgg006 INTO l_dzgg006 FROM dzgg_t 
               WHERE dzgg001 = l_prog_id AND dzgg002 = l_sd_ver AND dzgg003 = l_prog_id   
                 AND dzgg017 =  g_code_ide  #140619 add          
              IF l_dzgg006 = "Y" THEN 
              ##否要加簽核140422 -(e)
                  CALL adzp161_combi_subtmp("signstr","d05",p_node,ls_type) RETURNING l_sub_str
                  LET ls_subtmp = ls_subtmp, l_sub_str,"\n"
                  #DISPLAY "l_sign_str:",l_sub_str
                  LET ls_block_type = p_id,"_",ls_type,"_signstr"
                  #DISPLAY "ls_block_type:",ls_block_type
                  CALL g_properties.addAttribute(ls_block_type,ls_subtmp)
              END IF   ##否要加簽核140422 
           END IF 
           ##140416 -(E)   
       
           IF ls_subtype IS NOT NULL AND NOT cl_null(ls_subtype) THEN
              #組子報表樣板內容
              IF ls_type_t = "" OR cl_null(ls_type_t) OR ls_type_t IS NULL THEN
                  LET ls_subtmp ="" 
                  #組子報表樣板內容
                  CALL adzp161_combi_subtmp("subrep",ls_subtype,ls_child_node,"") RETURNING l_sub_str
                  LET ls_subtmp = ls_subtmp, l_sub_str,"\n"
                  LET ls_type_t = ls_type #處理完的type內容存入舊值
              ELSE          
                  IF ls_type = ls_type_t THEN  #同一block內，ls_subtmp繼續存
                     CALL adzp161_combi_subtmp("subrep",ls_subtype,ls_child_node,"") RETURNING l_sub_str
                     LET ls_subtmp = ls_subtmp, l_sub_str,"\n" 
                  ELSE 
                     #將上一個block存入屬性內
                     LET ls_block_type = p_id,"_",ls_type_t,"_subrep"
                     CALL g_properties.addAttribute(ls_block_type,ls_subtmp)
                     LET ls_subtmp = ""
                     CALL adzp161_combi_subtmp("subrep",ls_subtype,ls_child_node,"") RETURNING l_sub_str
                     LET ls_subtmp = ls_subtmp, l_sub_str,"\n" 
                     LET ls_type_t =ls_type #處理完的type內容存入舊值
                  END IF 
              END IF 
           ELSE
              IF ls_subtmp IS NOT NULL AND NOT cl_null(ls_subtmp) THEN 
                 #將上一個block存入屬性內
                 LET ls_block_type = p_id,"_",ls_type_t,"_subrep"
                 CALL g_properties.addAttribute(ls_block_type,ls_subtmp)             
              END IF     
           END IF 
       END FOR 
       #判斷是否在同一個block內
       
       LET ls_block_type = p_id,"_",ls_type_t,"_subrep"
       CALL g_properties.addAttribute(ls_block_type,ls_subtmp)
       LET ls_subtmp = ""

END FUNCTION

#XR子樣板的組合(讀入子報表樣板並組合置換內容)
PRIVATE FUNCTION adzp161_xr_subtmp(p_id,p_tmptype,p_node,p_field)
   DEFINE p_node               om.DomNode
   DEFINE p_tmptype            STRING 
   DEFINE p_id                 STRING 
   DEFINE p_field              STRING 
   DEFINE ls_subseq            STRING 
   DEFINE ls_channel_read      base.Channel
   DEFINE l_write_tmp          base.StringBuffer
   DEFINE ls_sub_query         STRING 
   DEFINE ls_tag               STRING 
   DEFINE ls_substr            STRING                 #區段子報表名稱
   DEFINE ls_readline          STRING 
   DEFINE ls_grup_type         STRING 
   DEFINE ls_sample_filename   STRING 
   DEFINE ls_sub_recordname    STRING 
   DEFINE ls_space_cnt         LIKE type_t.num5
   DEFINE ls_recseq            STRING 
   DEFINE ls_bmp_key           STRING    ##150429 add
   DEFINE li_sub_num           LIKE type_t.num5   
   DEFINE ls_sub_name          STRING
   DEFINE l_value              STRING
   DEFINE l_sub_node           om.DomNode
      

   LET ls_subseq = p_node.getAttribute("subseq") #子報表流水號
   LET li_sub_num = ls_subseq + 2  #子報表TEMP TABLE編號由3開始
   LET ls_sub_name = "subrep" , ls_subseq   

   CASE p_id
      WHEN "subrep"
         CALL adzp161_xr_read_subxml(p_node,ls_subseq) #讀取tab子報表屬性
           
   END CASE

   LET ls_channel_read = base.Channel.create()
   LET l_write_tmp = base.StringBuffer.create()

   CALL ls_channel_read.setDelimiter("")
   CALL l_write_tmp.clear()

   #定義取用樣板檔案
   IF (ls_subseq = "01" OR ls_subseq = "02" OR ls_subseq = "03" OR ls_subseq = "04") THEN
      #備註子報表專用slice 檔
      CASE p_id
         WHEN "subrep"
            LET ls_sample_filename = "slice/code_d09.template"
            
         WHEN "subrep_remark_proc"
            LET ls_sample_filename = "slice/code_d10.template"
         
         OTHERWISE
            RETURN ""
      END CASE
   ELSE
      #其他子報表   
      LET ls_sample_filename = "slice/code_",p_tmptype.trim(),".template"
   END IF
   
   LET ls_sample_filename = os.Path.join(os.Path.join(FGL_GETENV("ERP"),"mdl"),ls_sample_filename)
   DISPLAY "切片樣板檔位置:",ls_sample_filename

   IF NOT os.Path.exists(ls_sample_filename) THEN
      DISPLAY "ERROR(4):切片樣板檔案:",ls_sample_filename.trim()," 不存在!"
      EXIT PROGRAM
   END IF
   CALL ls_channel_read.openFile( ls_sample_filename CLIPPED, "r" )

   #讀取及轉換
   WHILE TRUE
   
      LET ls_readline = ls_channel_read.readLine()

      CASE p_id
         WHEN "subrep"
              CALL adzp161_part_replace(ls_readline,"${sub_name}",ls_sub_name) RETURNING ls_readline
              CALL adzp161_part_replace(ls_readline,"${sub_seq}",li_sub_num) RETURNING ls_readline
             
              LET l_value = g_sub_properties.getValue("sub_fields_define")
              CALL adzp161_part_replace(ls_readline,"${sub_fields_define}",l_value) RETURNING ls_readline             
              
              LET l_value = g_sub_properties.getValue("sub_sr_rep_str")
              CALL adzp161_part_replace(ls_readline,"${sub_sr_rep_str}",l_value) RETURNING ls_readline             
              
              LET l_value = g_sub_properties.getValue("sub_select_query")
              CALL adzp161_part_replace(ls_readline,"${sub_select_query}",l_value) RETURNING ls_readline
              
              LET l_value = g_sub_properties.getValue("sub_from_query")
              CALL adzp161_part_replace(ls_readline,"${sub_from_query}",l_value) RETURNING ls_readline
              
              LET l_value = g_sub_properties.getValue("sub_where_query")
              CALL adzp161_part_replace(ls_readline,"${sub_where_query}",l_value) RETURNING ls_readline
              
              LET l_value = g_sub_properties.getValue("g_sub_filter")
              CALL adzp161_part_replace(ls_readline,"${g_sub_filter}",l_value) RETURNING ls_readline
              
              LET l_value = g_sub_properties.getValue("subtable")
              CALL adzp161_part_replace(ls_readline,"${subtable}",l_value) RETURNING ls_readline
              
              LET l_value = g_sub_properties.getValue("sub_temp_query")
              CALL adzp161_part_replace(ls_readline,"${sub_temp_query}",l_value) RETURNING ls_readline
              
              LET l_value = g_sub_properties.getValue("sub_temp_ins_prep")
              CALL adzp161_part_replace(ls_readline,"${sub_temp_ins_prep}",l_value) RETURNING ls_readline
              
              LET l_value = g_sub_properties.getValue("sub_param")
              CALL adzp161_part_replace(ls_readline,"${sub_param}",l_value) RETURNING ls_readline
              
         WHEN "subrep_remark_proc"
              CALL adzp161_part_replace(ls_readline,"${sub_name}",ls_sub_name) RETURNING ls_readline
              CALL adzp161_part_replace(ls_readline,"${sub_seq}",ls_subseq) RETURNING ls_readline     
              
         WHEN "subrep_proc"
              CALL adzp161_part_replace(ls_readline,"${sub_name}",ls_sub_name) RETURNING ls_readline
              CALL adzp161_part_replace(ls_readline,"${sub_seq}",ls_subseq) RETURNING ls_readline
              
      END CASE 

      LET ls_readline = (ls_space_cnt) SPACES,ls_readline ,"\n"
      
      CALL l_write_tmp.append(ls_readline)
      IF ls_channel_read.isEof() THEN
         EXIT WHILE
      END IF
    END WHILE
    
    RETURN l_write_tmp.toString()    
END FUNCTION

#子樣板的組合(讀入子報表樣板或其他區段樣板並組合置換內容)
PRIVATE FUNCTION adzp161_combi_subtmp(p_id,p_tmptype,p_node,p_field)
   DEFINE p_node               om.DomNode
   DEFINE p_tmptype            STRING 
   DEFINE p_id                 STRING 
   DEFINE p_field              STRING 
   DEFINE ls_subseq            STRING 
   DEFINE ls_channel_read      base.Channel
   DEFINE l_write_tmp          base.StringBuffer
   DEFINE ls_sub_query         STRING 
   DEFINE ls_tag               STRING 
   DEFINE ls_substr            STRING                 #區段子報表名稱
   DEFINE ls_readline          STRING 
   DEFINE ls_grup_type         STRING 
   DEFINE ls_sample_filename   STRING 
   DEFINE ls_sub_recordname    STRING 
   DEFINE ls_space_cnt         LIKE type_t.num5
   DEFINE ls_recseq            STRING 
   DEFINE ls_bmp_key           STRING    ##150429 add 
   
   CASE p_id
      WHEN "subrep"
           LET ls_subseq = p_node.getAttribute("sub_seq")        #子報表流水號
           LET ls_sub_recordname = p_node.getAttribute("recordseq")
           #組子報表sql
           IF p_tmptype="d03" THEN 
              CALL adzp161_combi_sample_query(p_node) RETURNING ls_sub_query
           ELSE  #複雜子報表格式:d04
              LET ls_subseq = p_node.getAttribute("sub_seq")        #子報表流水號
              CALL adzp161_combi_complex_query(p_node,ls_subseq)             
           END IF 
           
      WHEN "b_group"
           LET ls_grup_type="BEFORE"
           IF p_node.getTagName() = "section" THEN 
              LET ls_tag = g_properties.getValue("main_tag")
           ELSE 
              LET ls_tag = g_properties.getValue("sub_tag")
           END IF
           ##150429 bmp key add -(s)
           LET ls_bmp_key = g_properties.getValue("group_bmp_key")
           ##150429 bmp key add -(e)
      WHEN "a_group"
           LET ls_grup_type="AFTER"
           IF p_node.getTagName()="section" THEN 
              LET ls_tag = g_properties.getValue("main_tag")
           ELSE 
              LET ls_tag = g_properties.getValue("sub_tag")
           END IF           

      WHEN "everyrow"
           IF p_node.getAttribute("type")="before" THEN 
              LET ls_grup_type="before"
           ELSE 
              LET ls_grup_type="after"
           END IF 
           LET ls_subseq = p_node.getAttribute("sub_seq")        #子報表流水號
           #組子報表sql
           IF p_tmptype="d03" THEN 
              CALL adzp161_combi_sample_query(p_node) RETURNING ls_sub_query
           ELSE  #複雜子報表格式:d04
              CALL adzp161_combi_complex_query(p_node,ls_subseq)             
           END IF 
           
      WHEN "subreptag"
           LET ls_subseq = p_node.getAttribute("id")        #子報表流水號
           LET ls_recseq = p_node.getAttribute("recordseq") #子報表record流水號
           
   END CASE 
   
   
   LET ls_channel_read = base.Channel.create()
   LET l_write_tmp = base.StringBuffer.create()

   CALL ls_channel_read.setDelimiter("")
   CALL l_write_tmp.clear()

   #定義取用樣板檔案
   LET ls_sample_filename = "slice/code_",p_tmptype.trim(),".template"
   LET ls_sample_filename = os.Path.join(os.Path.join(FGL_GETENV("ERP"),"mdl"),ls_sample_filename)
   #DISPLAY "切片樣板檔位置:",ls_sample_filename

   IF NOT os.Path.exists(ls_sample_filename) THEN
      #DISPLAY "ERROR(4):切片樣板檔案:",ls_sample_filename.trim()," 不存在!"
      EXIT PROGRAM
   END IF
   CALL ls_channel_read.openFile( ls_sample_filename CLIPPED, "r" )

   #讀取及轉換
   WHILE TRUE
   
      LET ls_readline = ls_channel_read.readLine()

      CASE p_id
         WHEN "subrep"
              CALL adzp161_part_replace(ls_readline,"${subseq}",ls_subseq) RETURNING ls_readline
              CALL adzp161_part_replace(ls_readline,"${sub_query}",ls_sub_query) RETURNING ls_readline
              CALL adzp161_part_replace(ls_readline,"${sub_recordname}",ls_sub_recordname) RETURNING ls_readline

        WHEN "subreptag"
              CALL adzp161_part_replace(ls_readline,"${subseq}",ls_subseq) RETURNING ls_readline
              CALL adzp161_part_replace(ls_readline,"${sub_recordname}",ls_recseq) RETURNING ls_readline
              
         WHEN "b_group"
              CALL adzp161_part_replace(ls_readline,"${group_typestr}",ls_grup_type) RETURNING ls_readline
              CALL adzp161_part_replace(ls_readline,"${group_type}",p_id) RETURNING ls_readline
              CALL adzp161_part_replace(ls_readline,"${group_field}",p_field) RETURNING ls_readline
              CALL adzp161_part_replace(ls_readline,"${rep_type}",ls_tag) RETURNING ls_readline    
              CALL adzp161_part_replace(ls_readline,"${group_bmp_key}",ls_bmp_key) RETURNING ls_readline       ##150429 add 

         WHEN "a_group"
              CALL adzp161_part_replace(ls_readline,"${group_typestr}",ls_grup_type) RETURNING ls_readline
              CALL adzp161_part_replace(ls_readline,"${group_type}",p_id) RETURNING ls_readline
              CALL adzp161_part_replace(ls_readline,"${group_field}",p_field) RETURNING ls_readline         
              CALL adzp161_part_replace(ls_readline,"${rep_type}",ls_tag) RETURNING ls_readline         

         WHEN "everyrow"
              CALL adzp161_part_replace(ls_readline,"${group_type}",ls_grup_type) RETURNING ls_readline
              CALL adzp161_part_replace(ls_readline,"${group_field}",ls_sub_query) RETURNING ls_readline         
              CALL adzp161_part_replace(ls_readline,"${rep_type}",ls_tag) RETURNING ls_readline         
              LET ls_substr = p_id,"_",p_field  
              CALL adzp161_part_replace(ls_readline,"${subrep_add}",g_properties.getValue(ls_substr)) RETURNING ls_readline          
        WHEN "signstr"
              CALL adzp161_part_replace(ls_readline,"${group_typestr}",ls_grup_type) RETURNING ls_readline
              CALL adzp161_part_replace(ls_readline,"${group_type}",p_id) RETURNING ls_readline
              CALL adzp161_part_replace(ls_readline,"${group_field}",p_field) RETURNING ls_readline
     
              
      END CASE 

      LET ls_readline = (ls_space_cnt) SPACES,ls_readline ,"\n"
      
      CALL l_write_tmp.append(ls_readline)
      IF ls_channel_read.isEof() THEN
         EXIT WHILE
      END IF
    END WHILE 
    
    RETURN l_write_tmp.toString()
    
END FUNCTION 

#組合複雜sql
PRIVATE FUNCTION adzp161_combi_complex_query(p_node,p_subseq)
   DEFINE p_node               om.DomNode
   DEFINE p_subseq             STRING 
   DEFINE l_select             STRING 
   DEFINE l_from               STRING 
   DEFINE l_where              STRING 
   DEFINE l_order              STRING
   DEFINE ls_sql_name           STRING
   DEFINE ls_where,ls_pk       STRING
   DEFINE ls_token_pk          base.StringTokenizer
   DEFINE ls_token_where       base.StringTokenizer
   DEFINE ls_token_pk_str      STRING 
   DEFINE ls_token_where_str   STRING 
   DEFINE ls_pk_cnt            LIKE type_t.num5
   DEFINE ls_token_pk_cnt      LIKE type_t.num5 
      
   
   LET l_select = p_node.getAttribute("sub_query")
   IF l_select IS NOT NULL AND NOT cl_null(l_select) THEN
      LET l_select = " LET g_select = \"",l_select ,"\"" 
      LET ls_sql_name = "g_select_subrep",p_subseq,"_query"
      CALL g_properties.addAttribute(ls_sql_name, l_select)  
   END IF 
   
   LET l_from = p_node.getAttribute("table")
   IF l_from IS NOT NULL AND NOT cl_null(l_from) THEN
      LET l_from = " LET g_from = \" FROM ",l_from ,"\"" 
      LET ls_sql_name = "g_from_subrep",p_subseq,"_query"
      CALL g_properties.addAttribute(ls_sql_name, l_from)   
   END IF
   
   LET ls_pk = p_node.getAttribute("pk")
   LET ls_where = p_node.getAttribute("where")
   LET l_where = ""
   IF ls_pk IS NOT NULL AND NOT cl_null(ls_pk) THEN
        LET ls_token_pk = base.StringTokenizer.create(ls_pk.trim(), ',')
        LET ls_pk_cnt = ls_token_pk.countTokens()
        LET ls_token_pk_cnt = 0
        LET ls_token_where = base.StringTokenizer.create(ls_where.trim(), ',')
        IF ls_pk_cnt > 0 THEN 
            WHILE ls_token_pk.hasMoreTokens()
               LET ls_token_pk_str = ls_token_pk.nextToken()
               LET ls_token_where_str = ls_token_where.nextToken()
               IF ls_token_pk_cnt <> 0 THEN
                  LET l_where = l_where ,", \" AND " 
               END IF 
               LET l_where = l_where,
                                " ", ls_token_pk_str," = '\", sr1.",ls_token_where_str," CLIPPED ,\"'\""
               LET ls_token_pk_cnt = ls_token_pk_cnt + 1           
            END WHILE 
        ELSE 
            LET l_where = l_where,"' = \", sr1.",ls_token_where_str," CLIPPED ,\"'\""
        END IF    
   END IF 
   #SQL句前加"
  #140118 先拿掉null的判斷
  # IF l_where IS NOT NULL AND NOT cl_null(l_where) THEN
      LET l_where = "LET g_where = \" WHERE ",l_where 
      LET ls_sql_name = "g_where_subrep",p_subseq,"_query"
      CALL g_properties.addAttribute(ls_sql_name, l_where)  
   #END IF 

   LET l_order = p_node.getAttribute("suborder")
   IF l_order IS NOT NULL AND NOT cl_null(l_order) THEN 
         LET l_order = " LET g_order = ","\" ORDER BY ",l_order
      LET ls_sql_name = "g_order_subrep",p_subseq,"_query"
      CALL g_properties.addAttribute(ls_sql_name, l_order)
   END IF 
  
END FUNCTION 


PRIVATE FUNCTION adzp161_part_replace(p_str,p_old,p_new)
   DEFINE p_str                STRING 
   DEFINE p_old                STRING 
   DEFINE p_new                STRING 
   DEFINE l_tmp_str            base.StringBuffer 
   DEFINE l_tmp                STRING 

   LET l_tmp_str = base.StringBuffer.create()
   CALL l_tmp_str.clear()
   CALL l_tmp_str.append(p_str)
   WHILE l_tmp_str.getIndexOf(p_old,1) > 0
      CALL l_tmp_str.replace(p_old,p_new,0)  
   END WHILE 
   LET l_tmp = l_tmp_str.toString()
   
   RETURN l_tmp
END FUNCTION 


#+ 逐行代換
PUBLIC FUNCTION adzp161_line_replace(ls_read)
   DEFINE ls_read              STRING
   DEFINE ls_text              STRING
   DEFINE ls_tag               STRING
   DEFINE li_pos1              LIKE type_t.num10
   DEFINE li_pos2              LIKE type_t.num10
   DEFINE ls_temp              STRING                #暫存properties資料用

   LET li_pos1 = ls_read.getIndexOf("${",1)
   LET li_pos2 = ls_read.getIndexOf("}", li_pos1)
    
   IF li_pos1 > 0 AND li_pos2 > 0 AND li_pos1 < li_pos2 THEN
      LET ls_text = ""
      LET ls_tag = ls_read.subString(li_pos1 +2, li_pos2 -1 ) #取出要置換的tag

      #由SaxAttribute內取出值進行代換
      #不在行首
      IF li_pos1 > 1 THEN
         LET ls_text = ls_read.subString(1,li_pos1 -1)
      END IF

      #中間段
      LET ls_temp = g_properties.getValue(ls_tag) CLIPPED
 
      LET ls_text = ls_text,g_properties.getValue(ls_tag) CLIPPED,
                       ls_read.subString(li_pos2+1,ls_read.getLength())

      #遞迴處理同行其他組
      IF ls_text.getIndexOf("${",1) THEN
         LET ls_text = adzp161_line_replace(ls_text)
      END IF
   END IF

   RETURN ls_text

END FUNCTION


#組簡單子報表的sql
PRIVATE FUNCTION adzp161_combi_sample_query(p_node)
   DEFINE p_node               om.DomNode 
   DEFINE ls_sub_query         STRING 
   DEFINE ls_pk                STRING 
   DEFINE ls_where             STRING 
   DEFINE ls_qry_tmp           STRING
   DEFINE ls_token_pk          base.StringTokenizer
   DEFINE ls_token_where       base.StringTokenizer
   DEFINE ls_token_str         STRING 
   DEFINE ls_token_pk_str      STRING  
   DEFINE ls_token_where_str   STRING 
   DEFINE ls_token_pk_cnt      INTEGER
   DEFINE ls_pk_cnt            LIKE type_t.num5   

   LET ls_sub_query = p_node.getAttribute("sub_query")
   LET ls_pk = p_node.getAttribute("pk")
   LET ls_where = p_node.getAttribute("where")

   LET ls_qry_tmp = ""
   IF ls_sub_query IS NOT NULL AND NOT cl_null(ls_sub_query) THEN
      LET ls_qry_tmp = ls_qry_tmp , ls_sub_query CLIPPED 
      IF ls_pk IS NOT NULL AND NOT cl_null(ls_pk) THEN
           LET ls_token_pk = base.StringTokenizer.create(ls_pk.trim(), ',')
           LET ls_pk_cnt = ls_token_pk.countTokens()
           LET ls_token_pk_cnt = 0
           LET ls_token_where = base.StringTokenizer.create(ls_where.trim(), ',')
           IF ls_pk_cnt > 0 THEN 
               WHILE ls_token_pk.hasMoreTokens()
                  LET ls_token_pk_str = ls_token_pk.nextToken()
                  LET ls_token_where_str = ls_token_where.nextToken()
                  IF ls_token_pk_cnt <> 0 THEN
                     LET ls_qry_tmp = ls_qry_tmp ,", \" AND " 
                  END IF 
                  #170120-00037#1 add -(s)
                  IF ls_token_pk_str = "ooff004" THEN #是數值
                      LET ls_qry_tmp = ls_qry_tmp,
                                       #" ", ls_token_pk_str," = \", sr1.",ls_token_where_str," CLIPPED ,\"\"" #PSG-2667 mark
                                       " ", ls_token_pk_str," = '\", sr1.",ls_token_where_str," CLIPPED ,\"'\"" #PSG-2667 add
                  ELSE 
                  #170120-00037#1 add -(e)                  
                      LET ls_qry_tmp = ls_qry_tmp,
                                       " ", ls_token_pk_str," = '\", sr1.",ls_token_where_str," CLIPPED ,\"'\""
                      LET ls_token_pk_cnt = ls_token_pk_cnt + 1  
                  END IF #170120-00037#1 add         
               END WHILE 
           ELSE 
               LET ls_qry_tmp = ls_qry_tmp,"' = \", sr1.",ls_token_where_str," CLIPPED ,\"'\""
           END IF    
      ELSE 
          IF ls_where IS NOT NULL AND NOT cl_null(ls_where) THEN
             LET ls_qry_tmp = ls_qry_tmp, ls_where," CLIPPED ,\"'\""
          END IF 
      END IF 
      #SQL句前加"
      LET ls_qry_tmp = "\"",ls_qry_tmp
      LET ls_qry_tmp = ls_qry_tmp,"\n"
   END IF 

   RETURN ls_qry_tmp
END FUNCTION

#+ XR產生field 別名
FUNCTION adzp161_xr_field_alias(p_field,p_type,p_block)
  DEFINE p_field         LIKE dzkc_t.dzkc004
  DEFINE p_type          STRING                      
  DEFINE p_block         LIKE dzkc_t.dzkc100
  DEFINE ls_dzkc004      LIKE dzkc_t.dzkc004
  DEFINE ls_dzkc006      LIKE dzkc_t.dzkc006
  DEFINE ls_dzkc011      LIKE dzkc_t.dzkc011
  DEFINE ls_comb_field   STRING 
  DEFINE ls_dzkd010      LIKE dzkd_t.dzkd010  
  DEFINE ls_dzkd011      LIKE dzkd_t.dzkd011
  DEFINE ls_dzkc010      LIKE dzkc_t.dzkc010
  DEFINE l_tmp           LIKE type_t.chr1   
  DEFINE l_field_type    LIKE dzkc_t.dzkc006          
  DEFINE l_table_name    LIKE dzkc_t.dzkc007
  DEFINE l_table_str     STRING             
  DEFINE l_cnt           LIKE type_t.num5 
 
  LET l_field_type = ''
  
   SELECT COUNT(1) INTO l_cnt
    FROM dzkc_t 
    WHERE dzkc001 = g_prog_id AND dzkc002 = g_prog_ver 
     AND (dzkc005 ='1' OR dzkc005 = '3') 
     AND dzkc011 = p_field
     AND dzkc009 = g_code_ide
     AND dzkc100 = p_block
     AND dzkc006 = 'A'
     
    IF l_cnt > 0 THEN
       RETURN 'A'
    END IF
    
  SELECT COUNT(*) INTO l_cnt
    FROM dzkc_t WHERE dzkc001 = g_prog_id AND dzkc002 = g_prog_ver 
     AND (dzkc005 ='1' OR dzkc005 = '3') 
     AND dzkc011 = p_field
     AND dzkc009 = g_code_ide
     AND dzkc100 = p_block
     AND dzkc006 = 'R'

    IF l_cnt > 0 THEN
       LET l_field_type = 'R'
    ELSE
      SELECT dzkc004 INTO ls_dzkc004
        FROM dzkc_t WHERE dzkc001 = g_prog_id AND dzkc002 = g_prog_ver 
         AND (dzkc005 ='1' OR dzkc005 = '3') 
         AND dzkc004 = p_field
         AND dzkc009 = g_code_ide
         AND dzkc100 = p_block
         AND dzkc006 = 'S'
  
      IF NOT cl_null(ls_dzkc004) THEN
         LET l_field_type = 'S'
      END IF
    END IF
    RETURN l_field_type 
END FUNCTION 

#+ XR產生create temp table field
FUNCTION adzp161_xr_temp_field(p_field,p_type,p_block)
  DEFINE p_field         LIKE dzkc_t.dzkc004
  DEFINE p_type          STRING                      
  DEFINE p_block         LIKE dzkc_t.dzkc100
  DEFINE ps_type_field   LIKE dzkc_t.dzkc004
  DEFINE ls_dzkc006      LIKE dzkc_t.dzkc006
  DEFINE ls_gztz001      LIKE gztz_t.gztz001
  DEFINE ls_comb_field   STRING 
  DEFINE ls_dzkd010      LIKE dzkd_t.dzkd010  
  DEFINE ls_dzkd011      LIKE dzkd_t.dzkd011
  DEFINE ls_dzkc010      LIKE dzkc_t.dzkc010
  DEFINE l_tmp           LIKE type_t.chr1   
  DEFINE l_field_str     STRING             
  DEFINE l_table_name    LIKE dzkc_t.dzkc007
  DEFINE l_table_str     STRING             
  DEFINE ls_dzkc006_cnt  LIKE type_t.num5   
  DEFINE l_dzkd_cnt      LIKE type_t.num5 
  
            
      LET l_tmp =''
      LET ls_comb_field = ''      
      #DISPLAY "p_field:",p_field
      
      #IF p_block matches 's*' THEN 
      #判斷自訂欄位:S   
      SELECT COUNT(1) INTO l_dzkd_cnt
       FROM dzkd_t WHERE dzkd001 = g_prog_id AND dzkd002 = g_prog_ver
        AND dzkd003 = p_field      
        AND dzkd008 = g_code_ide
        
      IF l_dzkd_cnt > 0 THEN
         LET ls_dzkc006 = "S"
      ELSE
        #判斷dzkc006 ㄧ般欄位:N ,參考欄位:R 可能多筆加上DISTINCT
        #先找參考欄位:R 
        SELECT DISTINCT dzkc006 INTO ls_dzkc006
        FROM dzkc_t WHERE dzkc001 = g_prog_id AND dzkc002 = g_prog_ver 
         AND (dzkc005 ='1' OR dzkc005 = '3') 
         AND dzkc004 = p_field
         AND dzkc009 = g_code_ide
         AND dzkc100 = p_block
         AND dzkc006 = 'R'
        
         IF cl_null(ls_dzkc006) THEN
            SELECT DISTINCT dzkc006 INTO ls_dzkc006
            FROM dzkc_t WHERE dzkc001 = g_prog_id AND dzkc002 = g_prog_ver 
             AND (dzkc005 ='1' OR dzkc005 = '3') 
             AND dzkc004 = p_field
             AND dzkc009 = g_code_ide
             AND dzkc100 = p_block
             AND dzkc006 = 'N'       
         END IF          

         #附属表栏位类型A
         IF cl_null(ls_dzkc006) THEN            
            SELECT DISTINCT dzkc006 INTO ls_dzkc006
            FROM dzkc_t WHERE dzkc001 = g_prog_id AND dzkc002 = g_prog_ver 
             AND (dzkc005 ='1' OR dzkc005 = '3') 
             AND dzkc004 = p_field
             AND dzkc009 = g_code_ide
             AND dzkc100 = p_block
            AND dzkc006 = 'A'           
         END IF
         
       END IF    
         
      IF SQLCA.SQLCODE THEN
         
         LET l_field_str = p_field         
         LET l_table_name = l_field_str.subString(1, l_field_str.getIndexOf(p_type,1)-2)
         
         LET ps_type_field =  p_type

         LET ls_dzkc006_cnt = 0 
         
         SELECT COUNT(dzkc006) INTO ls_dzkc006_cnt FROM dzkc_t 
          WHERE dzkc001 = g_prog_id AND dzkc002 = g_prog_ver AND (dzkc005 ='1' OR dzkc005 = '3')
            AND dzkc004 = ps_type_field AND dzkc009 = g_code_ide 
            
         IF ls_dzkc006_cnt =1 THEN  
            SELECT dzkc006 INTO ls_dzkc006 FROM dzkc_t WHERE dzkc001 = g_prog_id 
              AND dzkc002 = g_prog_ver AND (dzkc005 ='1' OR dzkc005 = '3') 
              AND dzkc004 = ps_type_field AND dzkc009 = g_code_ide 
         ELSE           
            IF ls_dzkc006_cnt > 1 THEN
            　#找到多個參考欄位
              LET ls_dzkc006 = 'N'　　
        　　ELSE
              SELECT dzkc006 INTO ls_dzkc006 FROM dzkc_t 
              WHERE dzkc001 = g_prog_id AND dzkc002 = g_prog_ver 
              AND (dzkc005 ='1' OR dzkc005 = '3') AND dzkc004 = ps_type_field 
              AND dzkc009 = g_code_ide AND dzkc007 = l_table_name
            END IF
         END IF 
         LET l_tmp = '1' 
      END IF
      
      IF ls_dzkc006 = 'N' OR ls_dzkc006 = 'R' OR ls_dzkc006 = 'A' THEN  #從table挑選的欄位,非自訂  
         IF l_tmp = '1' THEN 
            SELECT gztz001 INTO ls_gztz001 FROM gztz_t WHERE gztz002 = ps_type_field  
               AND gztz001 NOT LIKE '%rebuil' #151208-00023#1 add               
               AND gztz001 NOT LIKE 'erp%'   
               AND gztz001 NOT LIKE 'all%'
               AND gztz001 NOT LIKE 'b2b%'
               AND gztz001 NOT LIKE 'pos%'
               AND gztz001 NOT LIKE 'dsm%'
               
            LET ls_comb_field = p_field,".",ls_gztz001,".", ps_type_field          
         ELSE        
            SELECT gztz001 INTO ls_gztz001 FROM gztz_t WHERE gztz002 = p_field  
               AND gztz001 NOT LIKE '%rebuil' #151208-00023#1 add               
               AND gztz001 NOT LIKE 'erp%'   
               AND gztz001 NOT LIKE 'all%'
               AND gztz001 NOT LIKE 'b2b%'
               AND gztz001 NOT LIKE 'pos%'
               AND gztz001 NOT LIKE 'dsm%'
               
            LET ls_comb_field = p_field,".",ls_gztz001,".", p_field   
         END IF
      ELSE  
          #自定義欄位
          IF ls_dzkc006 = 'S' THEN 
             SELECT dzkd010,dzkd011 INTO ls_dzkd010 , ls_dzkd011
             FROM dzkd_t
             WHERE dzkd001 = g_prog_id AND dzkd002 = g_prog_ver 
               AND dzkd003 = p_field  
               AND dzkd008 = g_code_ide      
             IF SQLCA.SQLCODE THEN
                #目前想不到會有找不到的狀態
             END IF             
             LET ls_comb_field = p_field,".", ls_dzkd010 , "." , ls_dzkd011
           END IF
      END IF    

      RETURN ls_comb_field
END FUNCTION #adzp161_xr_temp_field

#XR組定義欄位段，組成欄位的內容
PRIVATE FUNCTION adzp161_xr_fields_define(p_tag,p_seq,p_list,p_block)  
   DEFINE p_list                  STRING               #內容值
   DEFINE p_tag                   STRING               #tagname
   DEFINE p_seq                   STRING               #record name
   DEFINE p_block                 LIKE dzkc_t.dzkc100  #區塊
   DEFINE l_prefix                STRING
   DEFINE lst_token               base.StringTokenizer
   DEFINE ls_token                STRING
   DEFINE li_i                    LIKE type_t.num10    #空幾格
   DEFINE pi_page                 LIKE type_t.num5
   DEFINE l_token_count           LIKE type_t.num5
   DEFINE l_str_token_count       LIKE type_t.num5
   DEFINE ls_type_fields_define   STRING      #form_head_field代換變數
   DEFINE ls_sql_var              STRING      #var_h_fill代換變數
   DEFINE ls_define_new           STRING      #define填入值
   DEFINE ls_sql_new              STRING      #sql填入值
   DEFINE ls_tmp                  STRING
   DEFINE ls_tmp2                 STRING
   DEFINE ls_value                STRING
   DEFINE ls_type                 STRING       #變數型態
   DEFINE ls_name                 STRING       #變數名稱
   DEFINE ls_default              STRING       #預設值
   DEFINE ls_default_new          STRING       #select sql default值
   DEFINE ls_type_fields_file     STRING 
   DEFINE ls_chk_field            STRING 
   DEFINE ls_token1               STRING 
   DEFINE ls_ins_rep_arr          STRING 
   DEFINE ls_crattmp_sql          STRING       #主報表CREATE TEMP SQL
   DEFINE ls_crattmp_insprep      STRING       #主報表INSERT段組出的?
   DEFINE l_comb_field            STRING       #回傳組create欄位
   DEFINE ls_name_str             STRING
   DEFINE ls_field_alias          STRING
   DEFINE ls_select_alias         STRING
   DEFINE ls_field_type           STRING
   DEFINE l_point                 LIKE type_t.num5
   DEFINE ls_dzkc004              LIKE dzkc_t.dzkc004
   DEFINE ls_dzkc007              LIKE dzkc_t.dzkc007
   DEFINE ls_dzkc011              LIKE dzkc_t.dzkc011
   
   LET ls_define_new = ""
   LET ls_sql_new = ""
   LET ls_default_new = ""            #select的字串
   LET ls_ins_rep_arr = ""            #ins_data段rep的陣列
   LET ls_crattmp_sql = ""
   LET ls_crattmp_insprep =""
   
   LET li_i = 1
   LET lst_token = base.StringTokenizer.create(p_list.trim(), ',')
   LET l_str_token_count = lst_token.countTokens()

   LET l_token_count = 0
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET l_token_count = l_token_count + 1

      LET ls_default = "" #161229-00050#1 add
      #逐個串接DEFINE 
      LET ls_define_new = ls_define_new, (li_space*li_i) SPACES
        
      #如果標示欄位名稱的字串中含有 (),表示要使用 type_t, 並且是用()內指定的型態
      #如果()裡有|，表示有預設值
      CASE 
         WHEN ls_token.getIndexOf("(",1) 
            LET ls_name = ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
            
            
            #參考欄位
            LET ls_type = ls_token.subString(ls_token.getIndexOf("(",1)+1,ls_token.getIndexOf(")",1)-1)
            IF ls_type.getIndexOf("|",1) > 0 THEN
               LET ls_type = ls_token.subString(ls_token.getIndexOf("(",1)+1,ls_token.getIndexOf("|",1)-1)
               #預設值
               #LET ls_default = ls_token.subString(ls_token.getIndexOf("|",1)+1,ls_token.getIndexOf(")",1)-1)
               ##140416
               LET ls_default = ls_token.subString(ls_token.getIndexOf("|",1)+1,ls_token.getLength()-1)
               #若有計算元，加上括號
               IF ls_default.getIndexOf("+",1)>0 OR ls_default.getIndexOf("-",1)>0 OR
                  ls_default.getIndexOf("*",1)>0 OR ls_default.getIndexOf("/",1)>0 THEN 
                  LET ls_default = "(",ls_default ,")"
               END IF 
               IF ls_type IS NULL OR cl_null(ls_type) THEN
                  LET ls_type = ls_token.subString(1,ls_token.getIndexOf("(",1)-1)    
               END IF 
               IF ls_default IS NULL OR cl_null(ls_default) THEN
                  LET ls_default = ls_name  
               END IF 
            END IF  
            #先判斷gztz_t是否有欄位與table存在
            LET ls_chk_field = ls_type
            LET ls_chk_field = ls_chk_field.subString(1, ls_chk_field.getIndexOf("_",1)-1)
            IF ls_type <> "STRING" AND ls_type <> "DATETIME" AND ls_type <> "BLOB" AND ls_type <> "CLOB" THEN 
               LET l_prefix = adzp161_define_table_name(ls_type)
            END IF 
            IF l_prefix IS NULL THEN LET l_prefix = "type_t." END IF
            CASE ls_type.toUpperCase()
               WHEN "STRING"                  
                  LET ls_define_new = ls_define_new, ls_name, " LIKE type_t.chr500"
               WHEN "DATETIME"
                  LET ls_define_new = ls_define_new, ls_name, " DATETIME YEAR TO SECOND"
               WHEN "BLOB"
                  LET ls_define_new = ls_define_new, ls_name, " LIKE type_t.", ls_type
                  LET ls_tmp = g_properties.getValue("general_lob")
                  LET ls_tmp = ls_tmp, "   LOCATE ",g_properties.getValue("master_var_title"),".",ls_name," IN FILE ", "\n"
                  CALL g_properties.addAttribute("general_lob_start",ls_tmp)
                  LET ls_tmp = g_properties.getValue("general_lob_end")
                  LET ls_tmp = ls_tmp, "   FREE ",g_properties.getValue("master_var_title"),".",ls_name, "\n"
                  CALL g_properties.addAttribute("general_lob_end",ls_tmp)
               WHEN "CLOB"
                  LET ls_define_new = ls_define_new, ls_name, " LIKE type_t.", ls_type
                  LET ls_tmp = g_properties.getValue("general_lob")
                  LET ls_tmp = ls_tmp, "   LOCATE ",g_properties.getValue("master_var_title"),".",ls_name," IN FILE ", "\n"
                  CALL g_properties.addAttribute("general_lob_start",ls_tmp)
                  LET ls_tmp = g_properties.getValue("general_lob_end")
                  LET ls_tmp = ls_tmp, "   FREE ",g_properties.getValue("master_var_title"),".",ls_name, "\n"
                  CALL g_properties.addAttribute("general_lob_end",ls_tmp)
               OTHERWISE
                  LET ls_define_new = ls_define_new, ls_name, " LIKE " ,l_prefix , ls_type
            END CASE 
            LET ls_sql_new = ls_sql_new,ls_default

         WHEN ls_token.getIndexOf("date",1) OR ls_token.getIndexOf("buid",1)
            LET ls_define_new = ls_define_new, ls_token, " DATETIME YEAR TO SECOND"
            LET ls_sql_new = ls_sql_new,ls_token.trim()
            LET ls_name = ls_token
         WHEN ls_token.getIndexOf("crtdt",1) OR ls_token.getIndexOf("moddt",1)
            LET ls_define_new = ls_define_new, ls_token, " DATETIME YEAR TO SECOND"
            LET ls_sql_new = ls_sql_new,ls_token.trim()
            LET ls_name = ls_token
         OTHERWISE      
            IF ls_token.getIndexOf("_",1)>0 THEN    
               LET ls_token1 = ls_token.subString(1, ls_token.getIndexOf("_",1)-1)
               LET ls_name = ls_token1
            ELSE
               LET ls_token1 = ls_token
               LET ls_name = ls_token
            END IF 

            LET l_prefix = adzp161_define_table_name(ls_token1)
            if ls_token = 'xmdb001' then
               display "xxx:" , ls_token , "xxx:" ,l_prefix

            end if
            LET ls_define_new = ls_define_new, ls_token.trim(), " LIKE ", l_prefix, ls_token1.trim()
            LET ls_sql_new = ls_sql_new,ls_token.trim()
      END CASE

      #240118-00026 -S 別名
      LET l_point = ls_name.getIndexOf(".",1)
      IF l_point > 1 THEN
         LET ls_name = ls_name.subString(l_point+1,ls_name.getLength())
         display "new name:", ls_name
      END IF
      #240118-00026 -E
      #處理ins_data段，rep陣列字串  
      CASE
         WHEN p_block = 'h'
            LET ls_ins_rep_arr = ls_ins_rep_arr , "sr_h.", ls_name
         WHEN p_block = 'd'
            LET ls_ins_rep_arr = ls_ins_rep_arr , "sr_d.", ls_name
         OTHERWISE
            LET ls_ins_rep_arr = ls_ins_rep_arr , "sr.", ls_name
      END CASE
      IF ls_name MATCHES "ooff*" THEN
         LET l_comb_field = ls_name ,".ooff_t." , ls_name
      ELSE

         CALL adzp161_xr_temp_field(ls_name,ls_type,p_block) RETURNING l_comb_field
         
         LET ls_field_type = adzp161_xr_field_alias(ls_name,ls_type,p_block)       
         IF ls_name='t2_xmdc001' then
             display "同名欄位:" , p_tag , "--" , ls_name , "--" , ls_field_type
         END IF
         #同名欄位      
         IF ls_field_type='A' then
            LET ls_dzkc011 = ls_name
            SELECT dzkc004,dzkc007 INTO ls_dzkc004,ls_dzkc007
              FROM dzkc_t 
             WHERE dzkc001 = g_prog_id AND dzkc002 = g_prog_ver 
               AND (dzkc005 ='1' OR dzkc005 = '3') 
               AND dzkc011 = ls_dzkc011
               AND dzkc009 = g_code_ide
               AND dzkc100 = p_block
               AND dzkc006 = 'A'
         END IF                 
      END IF
                  
      #組create temp table sql與insert into的?
      LET ls_crattmp_sql = ls_crattmp_sql,l_comb_field
      LET ls_crattmp_insprep =ls_crattmp_insprep,"?"
       
      IF l_str_token_count != l_token_count THEN
         LET ls_define_new = ls_define_new, ", \n"
         #IF ls_field_type="S" OR ls_field_type="R" THEN            
         #   LET ls_select_alias = ls_select_alias,ls_default," ",ls_name, "," 
         #ELSE
         #   LET ls_select_alias = ls_select_alias,ls_token.trim(), "," 
         #END IF
         CASE
            WHEN ls_field_type="S" OR ls_field_type="R" 
               LET ls_select_alias = ls_select_alias,ls_default," ",ls_name, "," 
            WHEN ls_field_type="A" 
               LET ls_select_alias = ls_select_alias,ls_dzkc007,".",ls_dzkc004," ",ls_name, ","
            OTHERWISE
               LET ls_select_alias = ls_select_alias,ls_token.trim(), "," 
         END CASE
         LET ls_sql_new = ls_sql_new, "," 
         
         LET ls_ins_rep_arr = ls_ins_rep_arr, ","
         LET ls_crattmp_sql = ls_crattmp_sql, ","
         LET ls_crattmp_insprep = ls_crattmp_insprep, ","
      ELSE
         #IF ls_field_type="S" OR ls_field_type="R" THEN            
         #   LET ls_select_alias = ls_select_alias,ls_default," ",ls_name
         #ELSE
         #   LET ls_select_alias = ls_select_alias,ls_token.trim()
         #END IF
         CASE
            WHEN ls_field_type="S" OR ls_field_type="R" 
               LET ls_select_alias = ls_select_alias,ls_default," ",ls_name
            WHEN ls_field_type="A" 
               LET ls_select_alias = ls_select_alias,ls_dzkc007,".",ls_dzkc004," ",ls_name
            OTHERWISE
               LET ls_select_alias = ls_select_alias,ls_token.trim()
         END CASE
      END IF
    
   END WHILE
 
   IF p_tag ="var" THEN 
      LET ls_type_fields_file = "rep_fields_define",p_seq USING "<<<"        
      LET ls_define_new = "\n" , ls_define_new ,"\n"  
      
      IF p_seq = 1 THEN  
         CALL g_properties.addAttribute("set_sr_rep_str", ls_ins_rep_arr.trimRight())
         #240118-00026 -S
         CALL g_properties.addAttribute("master_fields_define", ls_define_new.trimRight())
         #240118-00026 -E        
      END IF
      IF p_seq = 2 THEN
         CALL g_properties.addAttribute("body_fields_define", ls_define_new.trimRight())
         CALL g_properties.addAttribute("body_sr_rep_str", ls_ins_rep_arr.trimRight())
      END IF
      
      CASE
         WHEN p_seq = 2
            CALL g_properties.addAttribute("body_temp_query", ls_crattmp_sql.trimRight())
            CALL g_properties.addAttribute("body_temp_ins_prep", ls_crattmp_insprep.trimRight())
            
         WHEN p_seq > 2
            CALL g_sub_properties.addAttribute("sub_temp_query", ls_crattmp_sql.trimRight())            
            CALL g_sub_properties.addAttribute("sub_temp_ins_prep", ls_crattmp_insprep.trimRight())
            CALL g_sub_properties.addAttribute("sub_sr_rep_str", ls_ins_rep_arr.trimRight())
            CALL g_sub_properties.addAttribute("sub_fields_define" , ls_define_new.trimRight())
            
         OTHERWISE
            CALL g_properties.addAttribute("g_temp_query", ls_crattmp_sql.trimRight())
            CALL g_properties.addAttribute("g_temp_ins_prep", ls_crattmp_insprep.trimRight())               
      END CASE
   ELSE    #g_select才需設預設值
      LET ls_define_new = "\n" , ls_define_new ,"\n"
      LET ls_type_fields_file = "master_fields_define",p_seq USING "<<<" 
      LET ls_sql_var = "g_select_query"
      #SELECT參考欄位加上別名
      #DISPLAY "alias:" ,ls_select_alias
      LET ls_sql_new = ls_select_alias     
      LET ls_sql_new = "LET g_select = \" SELECT ",ls_sql_new ,"\""
      
      #DISPLAY "before ls_sql_new:",ls_sql_new
      LET ls_sql_new = adzp161_replace_keyword(ls_sql_new)
      #LET ls_sql_new = sadzp188_handle_dzgb011_dlang(ls_sql_new)
      #DISPLAY "ls_sql_new:",ls_sql_new
      
      
      IF p_seq = "g_body_select" THEN
         LET ls_sql_var = "g_body_select_query"
      END IF
      IF p_seq = "g_sub_select" THEN
         CALL g_sub_properties.addAttribute("sub_select_query", ls_sql_new.trim())
      ELSE
         CALL g_properties.addAttribute(ls_sql_var, ls_sql_new.trim())
      END IF
             
   END IF   
   
   IF p_seq = "g_sub_select" THEN
      IF p_block="s1" AND cl_null(g_properties.getValue("remark_fields_define")) THEN
         #k01主樣版備註欄位          
         CALL g_properties.addAttribute("remark_fields_define", ls_define_new.trimRight())       
      END IF
      #CALL g_sub_properties.addAttribute("sub_fields_define", ls_define_new.trimRight())      
   END IF
   
   #IF p_seq = "g_body_select" THEN
   #   CALL g_properties.addAttribute("body_fields_define", ls_define_new.trimRight())
   #END IF
   
   #240118-00026 -S
   #IF p_seq = "g_select" THEN
   #   CALL g_properties.addAttribute("master_fields_define", ls_define_new.trimRight())
   #END IF
   #240118-00026 -E

END FUNCTION #adzp161_xr_fields_define

FUNCTION adzp161_replace_keyword(ps_sql)     
   DEFINE ps_sql        STRING              
   DEFINE l_str         base.StringBuffer
   DEFINE l_tmp         STRING 
   
   LET l_str = base.StringBuffer.create()
   CALL l_str.clear()
   CALL l_str.append(ps_sql)
   
   IF l_str.getIndexOf("g_dlang",1) > 0 THEN   
      IF l_str.getIndexOf(",g_dlang,",1) = 0  THEN  #排除已被處理過的
         LET l_tmp = "'\" ,", "g_dlang",",\"'\" ,\""  
         CALL l_str.replace("g_dlang",l_tmp,0)    
      END IF 
   END IF
  
  IF l_str.getIndexOf("g_site",1) > 0 THEN   
     LET l_tmp = "'\" ,", "g_site",",\"'\" ,\""    
     CALL l_str.replace("g_site",l_tmp,0)      
  END IF
  
  #IF l_str.getIndexOf(":TODAY",1) > 0 THEN   
  #   LET l_tmp = "'\" ,", "g_today",",\"'\" ,\""   
  #   CALL l_str.replace(":TODAY",l_tmp,0)   
  #END IF
  #
  #IF l_str.getIndexOf(":LEGAL",1) > 0 THEN   
  #   LET l_tmp = "'\" ,", "g_legal",",\"'\" ,\""   
  #   CALL l_str.replace(":LEGAL",l_tmp,0)     
  #END IF
  #
  #IF l_str.getIndexOf(":USER",1) > 0 THEN 
  #   LET l_tmp = "'\" ,", "g_user",",\"'\" ,\""
  #   CALL l_str.replace(":USER",l_tmp,0)     
  #END IF
   
   RETURN l_str.toString()
END FUNCTION 

#組定義欄位段，組成欄位的內容
PRIVATE FUNCTION adzp161_fields_define(p_tag,p_seq,p_list)  
   DEFINE p_list                  STRING               #內容值
   DEFINE p_tag                   STRING               #tagname
   DEFINE p_seq                   STRING               #record name
   DEFINE l_prefix                STRING
   DEFINE lst_token               base.StringTokenizer
   DEFINE ls_token                STRING
   DEFINE li_i                    LIKE type_t.num10    #空幾格
   DEFINE pi_page                 LIKE type_t.num5
   DEFINE l_token_count           LIKE type_t.num5
   DEFINE l_str_token_count       LIKE type_t.num5
   DEFINE ls_type_fields_define   STRING      #form_head_field代換變數
   DEFINE ls_sql_var              STRING      #var_h_fill代換變數
   DEFINE ls_define_new           STRING      #define填入值
   DEFINE ls_sql_new              STRING      #sql填入值
   DEFINE ls_tmp                  STRING
   DEFINE ls_tmp2                 STRING
   DEFINE ls_value                STRING
   DEFINE ls_type                 STRING       #變數型態
   DEFINE ls_name                 STRING       #變數名稱
   DEFINE ls_default              STRING       #預設值
   DEFINE ls_default_new          STRING       #select sql default值
   DEFINE ls_type_fields_file     STRING 
   DEFINE ls_chk_field            STRING 
   DEFINE ls_token1               STRING 
   DEFINE ls_ins_rep_arr          STRING 
   DEFINE ls_crattmp_sql          STRING       #主報表CREATE TEMP SQL
   DEFINE ls_crattmp_insprep      STRING       #主報表INSERT段組出的?
   DEFINE l_comb_field            STRING       #回傳組create欄位
   DEFINE ls_name_str             STRING       #141120 add   

   
   
   LET ls_define_new = ""
   LET ls_sql_new = ""
   LET ls_default_new = ""            #select的字串
   LET ls_ins_rep_arr = ""            #ins_data段rep的陣列
   LET ls_crattmp_sql = ""
   LET ls_crattmp_insprep =""
   
   LET li_i = 1
   LET lst_token = base.StringTokenizer.create(p_list.trim(), ',')
   LET l_str_token_count = lst_token.countTokens()

   LET l_token_count = 0
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET l_token_count = l_token_count + 1

      LET ls_default = "" #161229-00050#1 add
      #逐個串接DEFINE 
      LET ls_define_new = ls_define_new, (li_space*li_i) SPACES
        
      #如果標示欄位名稱的字串中含有 (),表示要使用 type_t, 並且是用()內指定的型態
      #如果()裡有|，表示有預設值
      CASE 
         WHEN ls_token.getIndexOf("(",1) 
            LET ls_name = ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
            #參考欄位
            LET ls_type = ls_token.subString(ls_token.getIndexOf("(",1)+1,ls_token.getIndexOf(")",1)-1)
            IF ls_type.getIndexOf("|",1) > 0 THEN
               LET ls_type = ls_token.subString(ls_token.getIndexOf("(",1)+1,ls_token.getIndexOf("|",1)-1)
               #預設值
               #LET ls_default = ls_token.subString(ls_token.getIndexOf("|",1)+1,ls_token.getIndexOf(")",1)-1)
               ##140416
               LET ls_default = ls_token.subString(ls_token.getIndexOf("|",1)+1,ls_token.getLength()-1)
               #若有計算元，加上括號
               IF ls_default.getIndexOf("+",1)>0 OR ls_default.getIndexOf("-",1)>0 OR
                  ls_default.getIndexOf("*",1)>0 OR ls_default.getIndexOf("/",1)>0 THEN 
                  LET ls_default = "(",ls_default ,")"
               END IF 
               IF ls_type IS NULL OR cl_null(ls_type) THEN
                  LET ls_type = ls_token.subString(1,ls_token.getIndexOf("(",1)-1)    
               END IF 
               IF ls_default IS NULL OR cl_null(ls_default) THEN
                  LET ls_default = ls_name  
               END IF 
            END IF  
            #先判斷gztz_t是否有欄位與table存在
            LET ls_chk_field = ls_type
            LET ls_chk_field = ls_chk_field.subString(1, ls_chk_field.getIndexOf("_",1)-1)
            IF ls_type <> "STRING" AND ls_type <> "DATETIME" AND ls_type <> "BLOB" AND ls_type <> "CLOB" THEN 
               LET l_prefix = adzp161_define_table_name(ls_type)
            END IF 
            IF l_prefix IS NULL THEN LET l_prefix = "type_t." END IF
            CASE ls_type.toUpperCase()
               WHEN "STRING"                  
                  LET ls_define_new = ls_define_new, ls_name, " LIKE type_t.chr500"
               WHEN "DATETIME"
                  LET ls_define_new = ls_define_new, ls_name, " DATETIME YEAR TO SECOND"
               WHEN "BLOB"
                  LET ls_define_new = ls_define_new, ls_name, " LIKE type_t.", ls_type
                  LET ls_tmp = g_properties.getValue("general_lob")
                  LET ls_tmp = ls_tmp, "   LOCATE ",g_properties.getValue("master_var_title"),".",ls_name," IN FILE ", "\n"
                  CALL g_properties.addAttribute("general_lob_start",ls_tmp)
                  LET ls_tmp = g_properties.getValue("general_lob_end")
                  LET ls_tmp = ls_tmp, "   FREE ",g_properties.getValue("master_var_title"),".",ls_name, "\n"
                  CALL g_properties.addAttribute("general_lob_end",ls_tmp)
               WHEN "CLOB"
                  LET ls_define_new = ls_define_new, ls_name, " LIKE type_t.", ls_type
                  LET ls_tmp = g_properties.getValue("general_lob")
                  LET ls_tmp = ls_tmp, "   LOCATE ",g_properties.getValue("master_var_title"),".",ls_name," IN FILE ", "\n"
                  CALL g_properties.addAttribute("general_lob_start",ls_tmp)
                  LET ls_tmp = g_properties.getValue("general_lob_end")
                  LET ls_tmp = ls_tmp, "   FREE ",g_properties.getValue("master_var_title"),".",ls_name, "\n"
                  CALL g_properties.addAttribute("general_lob_end",ls_tmp)
               OTHERWISE
                  LET ls_define_new = ls_define_new, ls_name, " LIKE " ,l_prefix , ls_type
            END CASE 
            LET ls_sql_new = ls_sql_new,ls_default

         WHEN ls_token.getIndexOf("date",1) OR ls_token.getIndexOf("buid",1)
            LET ls_define_new = ls_define_new, ls_token, " DATETIME YEAR TO SECOND"
            LET ls_sql_new = ls_sql_new,ls_token.trim()
            LET ls_name = ls_token
         WHEN ls_token.getIndexOf("crtdt",1) OR ls_token.getIndexOf("moddt",1)
            LET ls_define_new = ls_define_new, ls_token, " DATETIME YEAR TO SECOND"
            LET ls_sql_new = ls_sql_new,ls_token.trim()
            LET ls_name = ls_token
         OTHERWISE      
            IF ls_token.getIndexOf("_",1)>0 THEN    
               LET ls_token1 = ls_token.subString(1, ls_token.getIndexOf("_",1)-1)
               LET ls_name = ls_token1
            ELSE
               LET ls_token1 = ls_token
               LET ls_name = ls_token
            END IF 

            LET l_prefix = adzp161_define_table_name(ls_token1)
            LET ls_define_new = ls_define_new, ls_token.trim(), " LIKE ", l_prefix, ls_token1.trim()
            LET ls_sql_new = ls_sql_new,ls_token.trim()
      END CASE


      #處理ins_data段，rep陣列字串
      IF p_seq = 1 THEN
         IF g_tmp = "g01" THEN  #GR
             IF g_select_define.getIndexOf(ls_name,1) > 0 THEN       #比對SELECT的RECORD，若無就給預設值
                LET ls_ins_rep_arr = ls_ins_rep_arr , "LET sr[l_cnt].", ls_name, " = sr_s.",ls_name,"\n",(li_space*2+1) SPACES
             ELSE
                LET ls_ins_rep_arr = ls_ins_rep_arr , "LET sr[l_cnt].", ls_name, " = ",ls_default,"\n",(li_space*2+1) SPACES
             END IF
         ELSE   #XG
             ##141120 mark -(s)
             IF g_select_define.getIndexOf(ls_name,1) > 0 THEN       #比對SELECT的RECORD，若無就給預設值
                LET ls_ins_rep_arr = ls_ins_rep_arr , "sr.", ls_name
                CALL adzp161_comb_crattemp_field(ls_name,ls_type) RETURNING l_comb_field  #140717 add #140728 add ls_type
             ELSE
                LET ls_ins_rep_arr = ls_ins_rep_arr , ls_default
                CALL adzp161_comb_crattemp_field(ls_default,ls_type) RETURNING l_comb_field #140717 add #140728 add ls_type
             END IF 
             #141120 mark -(e)
             #141120 add -(s)
             #IF g_select_define.getIndexOf(ls_name,1) > 0 THEN       #比對SELECT的RECORD，若無就給預設值
                #LET ls_ins_rep_arr = ls_ins_rep_arr , "sr.", ls_name
                #LET ls_name_str = ls_name
                #IF ls_name_str.getIndexOf("x_",1)>0 THEN
                   #LET ls_name_str = ls_name_str.subString(ls_name_str.getIndexOf("x_",1)+2,ls_name_str.getLength())
                   #CALL adzp161_comb_crattemp_field(ls_name_str,ls_type) RETURNING l_comb_field
                #ELSE
                   #CALL adzp161_comb_crattemp_field(ls_name,ls_type) RETURNING l_comb_field  #140717 add #140728 add ls_type
                #END IF
             #ELSE     
                 #IF g_select_define.getIndexOf(ls_default,1) > 0 THEN       #比對SELECT的RECORD，若無就給預設值
                    #LET ls_ins_rep_arr = ls_ins_rep_arr , ls_default
                    #LET ls_name_str = ls_default
                    #IF ls_name_str.getIndexOf("x.",1)>0 THEN   
                       #LET ls_name_str = ls_name_str.subString(ls_name_str.getIndexOf("x_",1)+2,ls_name_str.getLength())
                       #CALL adzp161_comb_crattemp_field(ls_name_str,ls_type) RETURNING l_comb_field
                    #ELSE
                       #CALL adzp161_comb_crattemp_field(ls_default,ls_type) RETURNING l_comb_field #140717 add #140728 add ls_type
                    #END IF
                 #END IF
             #END IF
             #141120 add -(e)



 
             #140717 add 組create temp table sql與insert into的?-(s)             
             LET ls_crattmp_sql = ls_crattmp_sql,l_comb_field
             LET ls_crattmp_insprep =ls_crattmp_insprep,"?"
             #140717 add 組create temp table sql與insert into的?-(e)
         END IF     
      END IF  
      IF l_str_token_count != l_token_count THEN
         LET ls_define_new = ls_define_new, ", \n"
         LET ls_sql_new = ls_sql_new, ","
         #IF g_tmp = "x01" AND p_seq = 1 THEN    #XG的需加,                  #160921-00012#1 mark
         #IF (g_tmp = "x01" OR g_tmp="x02") AND p_seq = 1 THEN    #XG的需加,  #160921-00012#1 add     #170809-00051#2 mark
         #IF ((g_tmp = "x01" OR g_tmp="x02" ) AND p_seq = 1) OR (g_tmp="f01" OR g_tmp ="f02" AND p_seq = 1) THEN    #增加f01、f02,  ##170809-00051#2 add #220302 mark elena
         IF ((g_tmp = "x01" OR g_tmp="x02" ) AND p_seq = 1) OR (g_tmp="f01" OR g_tmp ="f02" AND p_seq = 1) OR (g_tmp ="abi01" AND p_seq = 1) THEN #220302 add elena
            LET ls_ins_rep_arr = ls_ins_rep_arr, ","   
            #140717 add 組create temp table sql與insert into的?-(s)
            LET ls_crattmp_sql = ls_crattmp_sql, ","
            LET ls_crattmp_insprep = ls_crattmp_insprep, ","
            #140717 add 組create temp table sql與insert into的?-(e)
         END IF 
      END IF
    
   END WHILE
 
   IF p_tag ="var" THEN 
      LET ls_type_fields_file = "rep_fields_define",p_seq USING "<<<"        
      IF p_seq = 1 THEN
         CALL g_properties.addAttribute("set_sr_rep_str", ls_ins_rep_arr.trimRight())
         #140717 add 組create temp table sql與insert into的?-(s) 
         LET ls_crattmp_insprep = ls_crattmp_insprep, ")"
         LET ls_crattmp_insprep = ls_crattmp_insprep  ##130731 janet add  #將insert prep併進來
         #170809-00051#2 add -(s)
         IF g_tmp = "f01" OR g_tmp = "f02" THEN  ##FR在create sql前加2個欄位sessionid，timestamp
            LET ls_crattmp_sql = "fr_sessionid.gzgh_t.gzgh013,fr_timestamp.gzgh_t.gzgh007,",ls_crattmp_sql
            LET ls_crattmp_insprep = "?,?,",ls_crattmp_insprep  
            DISPLAY "ls_crattmp_insprep:",ls_crattmp_insprep       
         END IF 
         #170809-00051#2 add -(e)
         #220302 add elena start
         IF g_tmp = "abi01"  THEN  ##ABI在create sql前加2個欄位sessionid，timestamp
            LET ls_crattmp_sql = "abi_sessionid.gzgh_t.gzgh013,abi_timestamp.gzgh_t.gzgh007,",ls_crattmp_sql
            LET ls_crattmp_insprep = "?,?,",ls_crattmp_insprep
            DISPLAY "ls_crattmp_insprep:",ls_crattmp_insprep
         END IF
 
         #220302 add elena end
         CALL g_properties.addAttribute("g_temp_query", ls_crattmp_sql.trimRight())
         CALL g_properties.addAttribute("g_temp_ins_prep", ls_crattmp_insprep.trimRight())
         #140717 add 組create temp table sql與insert into的?-(e)
      END IF
       

   ELSE    #g_select才需設預設值
      LET ls_define_new = "\n" , ls_define_new ,"\n"
      LET ls_type_fields_file = "master_fields_define",p_seq USING "<<<" 
      LET ls_sql_var = "g_select_query"
 
      #220302 elena add start
      IF g_tmp = "abi01"  THEN
          LET ls_sql_new = "LET g_select = \" SELECT '\",","g_fr_sessionid,","\"","',","to_timestamp('","\"",",g_fr_timestamp,","\"","',","'yyyy-mm-dd hh24:mi:ss'),","\",", "\"",ls_sql_new ,"\""

      ELSE
      #220302 elena add end
         #170809-00051#2 add -(s) #FR前面要加上sessionid與timesatmp
         IF g_tmp = "f01" OR  g_tmp = "f02"  THEN 
            LET ls_sql_new = "LET g_select = \" SELECT '\",","g_fr_sessionid,","\"","',","to_timestamp('","\"",",g_fr_timestamp,","\"","',","'yyyy-mm-dd hh24:mi:ss'),","\",", "\"",ls_sql_new ,"\""
         ELSE 
         #170809-00051#2 add -(e)
            LET ls_sql_new = "LET g_select = \" SELECT ",ls_sql_new ,"\""
         END IF  #170809-00051#2 add
      END IF   #220302 elena add
 
      #DISPLAY "before ls_sql_new:",ls_sql_new
      LET ls_sql_new = sadzp188_handle_dzgb011_dlang(ls_sql_new) ##160615-00007#1 add
      #DISPLAY "ls_sql_new:",ls_sql_new
      CALL g_properties.addAttribute(ls_sql_var, ls_sql_new.trim())      
   END IF    
   
   CALL g_properties.addAttribute(ls_type_fields_file, ls_define_new.trimRight())

END FUNCTION


#傳入參數的global define宣告、組參數、元件內define、及init值 tm.wc,tm.d...等
PRIVATE FUNCTION adzp161_arg_fields_define(p_list)  
   DEFINE p_list                  STRING               #內容值
   DEFINE l_prefix                STRING
   DEFINE lst_token               base.StringTokenizer
   DEFINE ls_token                STRING
   DEFINE li_i                    LIKE type_t.num10    #空幾格
   DEFINE pi_page                 LIKE type_t.num5
   DEFINE l_token_count           LIKE type_t.num5
   DEFINE l_str_token_count       LIKE type_t.num5
   DEFINE ls_type_fields_define   STRING      #form_head_field代換變數
   DEFINE ls_sql_var              STRING      #var_h_fill代換變數
   DEFINE ls_g_define_new         STRING    #global define填入值
   DEFINE ls_define_new           STRING      #define填入值
   DEFINE ls_arg_new              STRING      #arg填入值
   DEFINE ls_tmp                  STRING
   DEFINE ls_tmp2                 STRING
   DEFINE ls_value                STRING
   DEFINE ls_type                 STRING       #變數型態
   DEFINE ls_name                 STRING       #變數名稱
   DEFINE ls_default              STRING       #預設值
   DEFINE ls_field_sql_head       STRING       #所有必需insert進單頭table的欄位 
   DEFINE ls_default_new          STRING       #select sql default值
   DEFINE ls_type_fields_file     STRING 
   DEFINE ls_chk_field            STRING 
   DEFINE ls_token1               STRING 
   DEFINE ls_init_str             STRING       #參數init值
   DEFINE l_arg_name              STRING       #參數命名
   DEFINE ls_g_name               STRING       #global define field
   DEFINE ls_slice                LIKE type_t.chr1  #,
   
   
   LET ls_g_define_new = ""
   LET ls_define_new = ""
   LET ls_arg_new = ""
   LET ls_init_str = "" 
   LET ls_slice =","   
   
   LET li_i = 1
   LET lst_token = base.StringTokenizer.create(p_list.trim(), ',')
   LET l_str_token_count = lst_token.countTokens()

   LET l_token_count = 0
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET l_token_count = l_token_count + 1
      LET l_arg_name = "p_arg",l_token_count USING "<<<"
      LET ls_arg_new = ls_arg_new,l_arg_name
      #逐個串接元件內的DEFINE 
      LET ls_define_new = ls_define_new, "DEFINE  "
      LET ls_g_define_new = ls_g_define_new, (li_space*2+1) SPACES

      #()內指定的型態與備註，ex:tm.wc(chr1|condition)
      #取括號前的值，ex:tm.wc
      LET ls_name = ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
      #.後的值，ex:"tm".wc
      LET ls_g_name = ls_name.subString(ls_name.getIndexOf(".",1)+1,ls_name.getLength())
      #參考欄位or型態
      LET ls_type = ls_token.subString(ls_token.getIndexOf("(",1)+1,ls_token.getIndexOf("|",1)-1)
      #備註
      LET ls_default = ls_token.subString(ls_token.getIndexOf("|",1)+1,ls_token.getIndexOf(")",1)-1)

      #先判斷gztz_t是否有欄位與table存在
      LET ls_chk_field = ls_type
      IF ls_type <> "STRING" AND ls_type <> "DATETIME" AND ls_type <> "BLOB" AND ls_type <> "CLOB" THEN 
         LET l_prefix = adzp161_define_table_name(ls_type)
      END IF  
      IF l_prefix IS NULL THEN LET l_prefix = "type_t." END IF

      IF l_str_token_count = l_token_count THEN 
         LET ls_slice= " " 
      END IF 
      CASE ls_type.toUpperCase()
         WHEN "STRING"  
                LET ls_define_new = ls_define_new, l_arg_name, " STRING", 18 SPACES , "#" ,ls_name,"  ",ls_default
                LET ls_g_define_new = ls_g_define_new, ls_g_name, " STRING",ls_slice, 18 SPACES ,  "#" ,ls_default
         WHEN "DATETIME"
            LET ls_define_new = ls_define_new, l_arg_name, " DATETIME YEAR TO SECOND" , 6 SPACES , "#" ,ls_name,"  ", ls_default
            LET ls_g_define_new = ls_g_define_new, ls_g_name, " DATETIME YEAR TO SECOND" ,ls_slice, 6 SPACES , "#" , ls_default
         WHEN "BLOB"
            LET ls_define_new = ls_define_new, l_arg_name, " LIKE type_t.", ls_type, 8 SPACES ,  "#" ,ls_name,"  ",ls_default
            LET ls_g_define_new = ls_g_define_new, ls_g_name, " LIKE type_t.", ls_type,ls_slice, 8 SPACES , "#" , ls_default
         WHEN "CLOB"
            LET ls_define_new = ls_define_new, l_arg_name, " LIKE type_t.", ls_type, 8 SPACES , "#" ,ls_name,"  ", ls_default
            LET ls_g_define_new = ls_g_define_new, ls_g_name, " LIKE type_t.", ls_type, ls_slice,8 SPACES , "#" , ls_default 
         OTHERWISE
            LET ls_define_new = ls_define_new, l_arg_name, " LIKE " ,l_prefix , ls_type, 9 SPACES , "#" ,ls_name,"  ", ls_default
            LET ls_g_define_new = ls_g_define_new, ls_g_name, " LIKE " ,l_prefix , ls_type,ls_slice, 9 SPACES , "#" , ls_default
      END CASE 

      IF l_str_token_count != l_token_count THEN
         LET ls_define_new = ls_define_new, " \n"
         LET ls_g_define_new = ls_g_define_new, " \n"
         LET ls_arg_new = ls_arg_new, ","
      END IF

      #init值
      LET ls_init_str = ls_init_str , "LET ", ls_name, " = ",l_arg_name,"\n",(li_space) SPACES
        
   END WHILE

   LET ls_type_fields_file = "tm_fields_define_str"
   CALL g_properties.addAttribute(ls_type_fields_file, ls_g_define_new.trimRight())
   LET ls_type_fields_file = "componet_define_str"
   CALL g_properties.addAttribute(ls_type_fields_file, ls_define_new.trimRight()) 
   ##舊版:傳遞參數 -(s)  
   LET ls_type_fields_file = "arg_name"
   CALL g_properties.addAttribute(ls_type_fields_file, ls_arg_new.trimRight())  
   ##舊版:傳遞參數 -(s)
   ##140721 mark -(s)改回用硬框架
   ##新版:傳遞參數寫到db裡:add-point -(s)140306
   #LET g_dzbb.prog_name   = g_properties.getValue("component_name")
   #LET g_dzbb.point_name  = "component.get_var"
   #LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
   #LET g_dzbb.addpoint    = g_properties.getValue("arg_name")
                            #程式名稱          add-point版本    add-point名稱     add-point內容   是否覆蓋add-poin(y/n)
   #CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,g_code_ide)
   ##新版:傳遞參數寫到db裡:add-point -(s)
   ##140721 mark -(e)改回用硬框架
   LET ls_type_fields_file = "componet_init_str"
   CALL g_properties.addAttribute(ls_type_fields_file, ls_init_str.trimRight())    
END FUNCTION



#+ 片段程式樣板讀取/程式寫出(一進一出) 
PRIVATE FUNCTION adzp161_replace_silce(p_slice_template,p_tag,p_type_grup,p_field)
   DEFINE p_slice_template     STRING 
   DEFINE p_tag                STRING 
   DEFINE p_type_grup          STRING 
   DEFINE p_field              STRING 
   DEFINE lchannel_read        base.Channel
   DEFINE l_write_tmp          base.StringBuffer    #存的暫存string
   DEFINE ls_readline          STRING 
   DEFINE ls_text              STRING
   DEFINE ls_code_filename     STRING
   DEFINE ls_sample_filename   STRING
   DEFINE lchannel_check       base.Channel
   DEFINE ls_w_tmp             STRING 

   LET lchannel_read = base.Channel.create()
   LET l_write_tmp = base.StringBuffer.create()

   CALL lchannel_read.setDelimiter("")
   CALL l_write_tmp.clear()

   #定義取用樣板檔案
   
   LET ls_sample_filename = "slice/code_",p_slice_template.trim(),".template"
   LET ls_sample_filename = os.Path.join(os.Path.join(FGL_GETENV("ERP"),"mdl"),ls_sample_filename)
   #DISPLAY "切片樣板檔位置:",ls_sample_filename

   IF NOT os.Path.exists(ls_sample_filename) THEN
      #DISPLAY "ERROR(4):切片樣板檔案:",ls_sample_filename.trim()," 不存在!"
      EXIT PROGRAM
   END IF
   CALL lchannel_read.openFile( ls_sample_filename CLIPPED, "r" )

   #讀取及轉換
   WHILE TRUE
   
      LET ls_text = ""
      LET ls_readline = lchannel_read.readLine()
      CALL adzp161_part_replace(ls_readline,"${group_type}",p_type_grup) RETURNING ls_readline
      CALL adzp161_part_replace(ls_readline,"${group_field}",p_field) RETURNING ls_readline
      CALL adzp161_part_replace(ls_readline,"${rep_type}",p_tag) RETURNING ls_readline
      LET ls_readline = ls_readline ,"\n"
      CALL l_write_tmp.append(ls_readline)
      IF lchannel_read.isEof() THEN
         EXIT WHILE
      END IF

      LET ls_w_tmp = l_write_tmp.toString()
      RETURN l_write_tmp
   END WHILE 
END FUNCTION

#+ 根據tabx設定之record數量產生對應之record段落
PUBLIC FUNCTION adzp161_make_records(pc_read,pc_write,pi_type)
   DEFINE pc_read              base.Channel
   DEFINE pc_write             base.Channel
   DEFINE pi_type              LIKE type_t.num5 
   DEFINE l_tmp                STRING 
   DEFINE l_record             DYNAMIC ARRAY OF STRING 
   DEFINE l_record_num         LIKE type_t.num5 
   DEFINE li_cnt               LIKE type_t.num5 
   DEFINE l_is                 STRING 
   DEFINE ls_name              STRING 
   DEFINE l_record_tmp         STRING 
   
   #先取出rep段落之樣本
   WHILE TRUE 
      LET l_tmp = pc_read.readLine()
      CASE 
         WHEN l_tmp.getIndexOf("#rep - End -",1)
            EXIT WHILE 
         WHEN l_tmp.getIndexOf("#rep - Start -",1)            
            CALL adzp161_make_keys_in_record(pc_read,pc_write,'m') RETURNING l_tmp,pc_read
         WHEN l_tmp.getIndexOf("#xgprep - End -",1)
            EXIT WHILE 
      END CASE
      #讀出由#rep -Start -開頭的到#rep - End -段的資訊
      LET l_record_tmp = l_record_tmp,l_tmp , "\n"
   END WHILE 

   #取得record總數量
   LET l_record_num = g_properties.getValue("rep_reocrd_cnt")

   #根據key數量產生對應的key段落並進行資料取代
   FOR li_cnt = 1 TO l_record_num   

      #先將${key}取代掉
      LET l_is = li_cnt
      LET l_record[li_cnt] = adzp161_replace_key(l_record_tmp,li_cnt)
   
      #若資料未取代完成則重複返回
      WHILE l_record[li_cnt].getIndexOf("${",1) > 0 
         LET l_record[li_cnt] = adzp161_line_replace(l_record[li_cnt])
      END WHILE  

      #回寫到檔案中
      CALL pc_write.write(l_record[li_cnt])
      
   END FOR    
   RETURN pc_read,pc_write
   
END FUNCTION 

#+ 根據tabx設定之key數量產生對應之key段落(但不回寫tgl)
PUBLIC FUNCTION adzp161_make_keys_in_record(pc_read,pc_write,l_type)
   DEFINE pc_read              base.Channel
   DEFINE pc_write             base.Channel
   DEFINE l_type               STRING 
   DEFINE l_tmp                STRING 
   DEFINE l_key                DYNAMIC ARRAY OF STRING 
   DEFINE l_key_num            LIKE type_t.num5 
   DEFINE li_cnt               LIKE type_t.num5 
   DEFINE l_is                 STRING 
   DEFINE ls_return            STRING
   DEFINE l_key_tmp            STRING 
   
   #先取出key段落之樣本
   WHILE TRUE 
      LET l_tmp = pc_read.readLine()
      IF l_tmp.getIndexOf("#rep - End -",1) THEN
          EXIT WHILE 
      END IF
      #讀出由#rep -Start -開頭的到#rep - End -段的資訊
      LET l_key_tmp = l_key_tmp,l_tmp , "\n"
   END WHILE 

   #取得key總數量
   LET l_key_num = g_properties.getValue("rep_reocrd_cnt")

   #根據key數量產生對應的key段落並進行資料取代
   FOR li_cnt = 1 TO l_key_num
      
      #若資料未取代完成則重複返回
      WHILE l_key[li_cnt].getIndexOf("${",1) > 0 
         LET l_key[li_cnt] = adzp161_line_replace(l_key[li_cnt])
      END WHILE  

      #回寫到檔案中
      LET ls_return = ls_return, l_key[li_cnt], "\n"    
   END FOR 
   RETURN ls_return,pc_read
   
END FUNCTION 


#+ 取代${key}並帶入實際數字
PUBLIC FUNCTION adzp161_replace_key(p_key, p_i)
   DEFINE p_key                STRING 
   DEFINE p_i                  STRING 
   DEFINE l_s                  LIKE type_t.num5 
   DEFINE l_e                  LIKE type_t.num5 
   DEFINE l_key                STRING 
   DEFINE l_i                  INTEGER 
   DEFINE l_s1                 LIKE type_t.num5 


   LET l_i = p_i
   WHILE TRUE     
      LET l_s = p_key.getIndexOf("${key}",1) - 1
      LET l_e = p_key.getLength()
      #取代${key}
         
      LET l_key = p_key.subString(1,l_s), p_i USING "<<<"    
      LET l_key = l_key, p_key.subString((l_s+7),l_e)  
      LET p_key = l_key

      IF l_key.getIndexOf("${key}",1) < 1 THEN 
          EXIT WHILE
      END IF 
   END WHILE 

   #FOR XG prep專用  
   #IF g_tmp="x01" THEN                  #160921-00012#1 mark
   IF g_tmp="x01" OR g_tmp="x02" THEN    #160921-00012#1 add
       LET l_i = p_i -1  

       WHILE TRUE     
          LET l_s = p_key.getIndexOf("${key1}",1) - 1
          LET l_e = p_key.getLength()
          #取代${key}             
          LET l_key = p_key.subString(1,l_s), l_i USING "<<<"    
          LET l_key = l_key, p_key.subString((l_s+8),l_e)   
          LET p_key = l_key

          IF l_key.getIndexOf("${key1}",1) < 1 THEN 
              EXIT WHILE
          END IF 
       END WHILE 
   END IF 
   
   RETURN l_key
   
END FUNCTION 


#+ 讀取片段樣板並進行取代(欄位檢查段落)
PUBLIC FUNCTION adzp161_make_slice(ps_mdl)
   DEFINE ps_mdl          STRING
   DEFINE ls_mdl          STRING
   DEFINE ls_read         STRING
   DEFINE ls_return       STRING
   DEFINE ls_return_tmp   STRING
   DEFINE ls_text         STRING
   DEFINE lchannel_read   base.Channel
   DEFINE ls_read_buf     base.StringBuffer
   
   #定義取用樣板檔案
   LET ls_mdl = "slice/code_",ps_mdl,".template"
   LET ls_mdl = os.Path.join(os.Path.join(FGL_GETENV("ERP"),"mdl"),ls_mdl)

   #開啟樣板檔
   LET lchannel_read = base.Channel.create()
   CALL lchannel_read.setDelimiter("")

   CALL lchannel_read.openFile( ls_mdl CLIPPED, "r" )
   
   WHILE NOT lchannel_read.isEof()
   
      LET ls_text = ""
      LET ls_read = lchannel_read.readLine()
      ##置換掉general_prefix為component_name
      IF ls_read.getIndexOf("general_prefix",1)>0 THEN 
          LET ls_read_buf = base.StringBuffer.create()
          CALL ls_read_buf.clear()
          CALL ls_read_buf.append(ls_read)
          CALL ls_read_buf.replace("general_prefix","component_name",1)
          LET ls_read = ls_read_buf.toString()
      END IF 
      #行代換/ 對 ${} 置換
      IF ls_read.getIndexOf("${",1) AND 
         ( ls_read.getIndexOf("${",1) < ls_read.getIndexOf("}",1) ) THEN
         LET ls_text = adzp161_line_replace(ls_read)
      ELSE
         LET ls_text = ls_read
      END IF
      LET ls_return = ls_return, ls_text, '\n'
   END WHILE  
   
   CALL lchannel_read.close()

   RETURN ls_return

END FUNCTION


PUBLIC FUNCTION adzp161_xg_execute_str_prep(p_node)
   DEFINE p_node          om.DomNode
   DEFINE ls_token        base.StringTokenizer
   DEFINE ls_tmp          STRING
   DEFINE ls_token_str    STRING  
   DEFINE ls_execute_str  STRING 
   DEFINE ls_token_cnt    LIKE type_t.num5
   DEFINE l_str_token_cnt LIKE type_t.num5

   LET ls_tmp = p_node.getAttribute("var")
   LET ls_token = base.StringTokenizer.create(ls_tmp.trim(), ',')
   LET ls_token_cnt = ls_token.countTokens()
   WHILE ls_token.hasMoreTokens()
      LET ls_token_str = ls_token.nextToken()
      LET l_str_token_cnt = l_str_token_cnt + 1  
      #與SELECT RECORD比較，若有直接給值，若無接預設值
      IF g_select_define.getIndexOf(ls_token_str,1)>0 THEN 
         LET ls_execute_str = ls_execute_str ,"sr.",ls_token_str
      ELSE 
         
      END IF 
      IF ls_token_cnt != l_str_token_cnt THEN          
         LET ls_execute_str = ls_execute_str ,","
      END IF         
   END WHILE 
      
   CALL g_properties.addAttribute("set_sr_rep_str", ls_execute_str.trimRight())
END FUNCTION 

##140717 add 組create temp table sql -(s)
FUNCTION adzp161_comb_crattemp_field(p_field,p_type) #140728 add p_type
  DEFINE p_field         LIKE dzgc_t.dzgc004
  DEFINE p_type          STRING                      #140728 add
  DEFINE ps_type_field   LIKE dzgc_t.dzgc004         #140729 add
  DEFINE ls_dzgc006      LIKE dzgc_t.dzgc006
  DEFINE ls_gztz001      LIKE gztz_t.gztz001
  DEFINE ls_comb_field   STRING 
  DEFINE ls_dzgd004_tmp  LIKE dzgd_t.dzgd004  
  DEFINE ls_dzgd006      LIKE dzgd_t.dzgd006  #公式
  DEFINE l_tmp           LIKE type_t.chr1            #140729 add
  DEFINE l_field_str     STRING                      #140917 add
  DEFINE l_table_name    LIKE dzgc_t.dzgc007         #140917 add
  DEFINE l_table_str     STRING                      #141023 add
  DEFINE ls_dzgc006_cnt  LIKE type_t.num5            #140917 add
  
      LET l_tmp =''
      LET ls_comb_field = ''
      #判斷是否為自訂欄位
      #DISPLAY "p_field:",p_field
      SELECT dzgc006 INTO ls_dzgc006 FROM dzgc_t WHERE dzgc001 = g_prog_id AND dzgc002 = g_prog_ver AND (dzgc005 ='1' OR dzgc005 = '3') AND dzgc004 = p_field AND dzgc009 = g_code_ide #140612 add
      IF SQLCA.SQLCODE THEN
         #目前想不到會有找不到的狀態
         #140728 add-(s)
         #用參考欄位再找一次 
         #140917 mark -(s)
         #LET ps_type_field =  p_type
         #SELECT dzgc006 INTO ls_dzgc006 FROM dzgc_t WHERE dzgc001 = g_prog_id AND dzgc002 = g_prog_ver AND (dzgc005 ='1' OR dzgc005 = '3') AND dzgc004 = ps_type_field AND dzgc009 = g_code_ide #140612 add
         #140917 mark -(e)
         #140917 add -(s)
         LET l_field_str = p_field
         #DISPLAY "p_type:",p_type
         LET l_table_name = l_field_str.subString(1, l_field_str.getIndexOf(p_type,1)-2) 
         
         LET ps_type_field =  p_type

         LET ls_dzgc006_cnt = 0 
         SELECT COUNT(dzgc006) INTO ls_dzgc006_cnt FROM dzgc_t WHERE dzgc001 = g_prog_id AND dzgc002 = g_prog_ver AND (dzgc005 ='1' OR dzgc005 = '3') AND dzgc004 = ps_type_field AND dzgc009 = g_code_ide 
         IF ls_dzgc006_cnt =1 THEN  
            SELECT dzgc006 INTO ls_dzgc006 FROM dzgc_t WHERE dzgc001 = g_prog_id AND dzgc002 = g_prog_ver AND (dzgc005 ='1' OR dzgc005 = '3') AND dzgc004 = ps_type_field AND dzgc009 = g_code_ide 
         ELSE
            DISPLAY "l_table_name:",l_table_name
            LET l_table_str = l_table_name
            ##141023 add -(s) 若是單身組出來 table=x_xxxxx，先將x_刪除
            IF l_table_str.getIndexOf("x_",1) > 0 THEN
               LET l_table_name = l_table_str.subString(l_table_str.getIndexOf("x_",1)+2,l_table_str.getLength())
            END IF
            ##141023 add -(e) 
            SELECT dzgc006 INTO ls_dzgc006 FROM dzgc_t WHERE dzgc001 = g_prog_id AND dzgc002 = g_prog_ver AND (dzgc005 ='1' OR dzgc005 = '3') AND dzgc004 = ps_type_field AND dzgc009 = g_code_ide AND dzgc007 = l_table_name
         END IF
         #140917 add -(e)              
         LET l_tmp = '1'
         #140728 add-(e)
      END IF
      IF ls_dzgc006 = 'N' THEN  #從table挑選的欄位,非自訂
         #140728 add -(s)
         IF l_tmp = '1' THEN 
            SELECT gztz001 INTO ls_gztz001 FROM gztz_t WHERE gztz002 = ps_type_field  
               AND gztz001 NOT LIKE '%rebuil' #151208-00023#1 add
               ##161227-00056#1 add -(s)
               AND gztz001 NOT LIKE 'erp%'   
               AND gztz001 NOT LIKE 'all%'
               AND gztz001 NOT LIKE 'b2b%'
               AND gztz001 NOT LIKE 'pos%'
               AND gztz001 NOT LIKE 'dsm%'
               ##161227-00056#1 add -(e)               
            LET ls_comb_field = p_field,".",ls_gztz001,".", ps_type_field          
         ELSE 
         #140728 add -(e)
            SELECT gztz001 INTO ls_gztz001 FROM gztz_t WHERE gztz002 = p_field  
               AND gztz001 NOT LIKE '%rebuil' #151208-00023#1 add
               ##161227-00056#1 add -(s)
               AND gztz001 NOT LIKE 'erp%'   
               AND gztz001 NOT LIKE 'all%'
               AND gztz001 NOT LIKE 'b2b%'
               AND gztz001 NOT LIKE 'pos%'
               AND gztz001 NOT LIKE 'dsm%'
               ##161227-00056#1 add -(e)               
            LET ls_comb_field = p_field,".",ls_gztz001,".", p_field   
         END IF   #140728 add
      ELSE  
          #自定義欄位
          IF ls_dzgc006 = 'Y' THEN 
             SELECT dzgd004,dzgd006 INTO ls_dzgd004_tmp,ls_dzgd006
             FROM dzgd_t
             WHERE dzgd001 = g_prog_id AND dzgd002 = g_prog_ver AND dzgd003 = p_field  AND dzgd008 = g_code_ide      
             IF SQLCA.SQLCODE THEN
                #目前想不到會有找不到的狀態
             END IF
               LET ls_comb_field = p_field,".",ls_dzgd004_tmp

           END IF  #IF ls_dzgc006 = 'Y'
      END IF    

      RETURN ls_comb_field
END FUNCTION 
##140717 add 組create temp table sql -(e)
      
##141119 add -(s)
#寫入樣板相關資訊
PUBLIC FUNCTION adzp161_update_gzzx(ps_line)
   DEFINE ps_line  STRING      #資訊字串(未解析)
   DEFINE lr_gzzx  RECORD LIKE gzzx_t.*
   DEFINE ls_ver   STRING
   DEFINE li_b     INTEGER
   DEFINE li_e     INTEGER
   DEFINE li_cnt   INTEGER

   #取得樣板代碼
   LET lr_gzzx.gzzx011 = g_properties.getValue("type_id")

   #取得樣板版本
   LET li_b = ps_line.getIndexOf('(Version:',1) + 8
   LET li_e = ps_line.getIndexOf(')',li_b)
   LET ls_ver = ps_line.subString(li_b+1,li_e-1)
   LET ls_ver = ls_ver.trim()
   LET lr_gzzx.gzzx012 = ls_ver

   #指定程式名稱
   LET lr_gzzx.gzzx001 = g_properties.getValue("app_id")

   #先檢查程式有註冊否
   LET li_cnt = 0
   SELECT COUNT(1) INTO li_cnt FROM gzzx_t
    WHERE gzzx001 = lr_gzzx.gzzx001

   IF li_cnt = 0 THEN
      #DISPLAY 'ERROR(0): 該程式未註冊，請先於azzi900進行註冊！'   #170123-00046#1 mark
      EXIT PROGRAM
   ELSE
      UPDATE gzzx_t
         SET gzzx011 = lr_gzzx.gzzx011,
             gzzx012 = lr_gzzx.gzzx012
       WHERE gzzx001 = lr_gzzx.gzzx001
          IF SQLCA.sqlcode THEN
             DISPLAY 'WARNING: 樣板資訊回寫gzzx_t失敗!'
          END IF
   END IF

END FUNCTION         
##141119 add -(e)

##150429 組bmp key add  -(s)
PUBLIC FUNCTION adzp161_combi_bmpkey(ps_dzed001)
   DEFINE ps_dzed001           LIKE dzed_t.dzed001
   DEFINE l_dzed004            LIKE dzed_t.dzed004   
   DEFINE ls_bmp_key           STRING                
   DEFINE ls_token             base.StringTokenizer
   DEFINE ls_token_str         STRING 
   DEFINE l_tmp_str            STRING 
   DEFINE ls_token_cnt         LIKE type_t.num5
   DEFINE ls_bmp_str           STRING 
   DEFINE l_dzgc004            LIKE dzgc_t.dzgc004

   SELECT dzed004 INTO l_dzed004 
     FROM dzed_t 
    WHERE dzed001 = ps_dzed001
      AND dzed003 = 'P'

   #解決pk值與dzgc中是否有選，有則組上bmp_key，沒有提示訊息
    LET g_msg = ''
   LET ls_bmp_key = l_dzed004
   LET ls_token = base.StringTokenizer.create(ls_bmp_key.trim(), ',')
   LET ls_token_cnt = ls_token.countTokens()
   WHILE ls_token.hasMoreTokens()
       LET ls_token_str = ls_token.nextToken()
       LET l_dzgc004 = ls_token_str
       IF adzp161_chk_dzgc_exist(ps_dzed001,l_dzgc004) THEN 
          LET l_tmp_str =''
          LET l_tmp_str = ls_token_str, "=' ,sr_h.",ls_token_str 
          IF ls_bmp_str.getLength() = 0 THEN 
             LET ls_bmp_str = "'",l_tmp_str
          ELSE
             LET ls_bmp_str = ls_bmp_str , ",'{+}", l_tmp_str
          END IF 
       ELSE 
          LET g_msg = g_msg, l_dzgc004,"未在已選欄位(dzgc_t)中，請將該欄位新增至已選欄位中"   
       END IF 
   END WHILE  
   IF g_msg IS NOT NULL THEN   
      DISPLAY "g_msg:",g_msg
   END IF 
   RETURN ls_bmp_str
END FUNCTION 

PUBLIC FUNCTION adzp161_chk_dzgc_exist(ps_table,ps_field)
   DEFINE ps_table    LIKE dzkc_t.dzkc007
   DEFINE ps_field    LIKE dzkc_t.dzkc004
   DEFINE l_cnt       LIKE type_t.num5
   
   LET l_cnt = 0 
   SELECT COUNT(dzkc004) INTO l_cnt
     FROM dzkc_t 
   WHERE dzkc001 = g_prog_id AND dzkc002 = g_prog_ver 
     AND (dzkc005 ='1' OR dzkc005 = '3') AND dzkc004 = ps_field 
     AND dzkc009 = g_code_ide AND dzkc007 = ps_table
    IF l_cnt = 0 THEN 
      RETURN FALSE 
    ELSE 
      RETURN TRUE 
    END IF 
END FUNCTION 
##150429 組bmp key add -(e)

#170809-00051#2 add -(s)
################################################################################
# Descriptions...: 檢查目標暫存檔狀態
#                : 若不存在則創建
#                : 若存在格式不同的暫存檔則DROP後再創建
# Memo...........: 170809-00051#2
# Usage..........: CALL adzp161_check_reptable ()
#                  RETURNING TRUE/FALSE
# Input parameter: p_sql
#                : p_cnt
#                : p_table
# Return code....: 作業成功否(TRUE/FALSE)
# Date & Author..: Janet
# Modify.........:
################################################################################
PUBLIC FUNCTION adzp161_check_reptable(p_sql)
   DEFINE p_sql         STRING
   DEFINE l_table       LIKE type_t.chr100
   DEFINE l_cnt         LIKE type_t.num5
   DEFINE l_i           LIKE type_t.num5
   DEFINE l_tok_table   base.StringTokenizer
   DEFINE l_name        STRING
   DEFINE l_p           LIKE type_t.num5
   DEFINE l_p2          LIKE type_t.num5
   DEFINE l_alias_name  LIKE type_t.chr100
   DEFINE l_table_name  LIKE type_t.chr100
   DEFINE l_field_name  LIKE type_t.chr100
   DEFINE l_datatype    LIKE type_t.chr100
   DEFINE l_length      LIKE gztd_t.gztd008
   DEFINE l_str         STRING
   DEFINE l_precision   LIKE gztd_t.gztd008
   DEFINE l_scale       LIKE gztd_t.gztd008
   DEFINE l_ts_type     LIKE type_t.chr100      #timestamp要另外處理欄位名
   DEFINE l_chk_rep     BOOLEAN  #MSSQL
   
   #MSSQL
   &ifdef is_msv 
#   LET g_report_t = g_report
   LET l_chk_rep = TRUE
   CALL adzp161_check_reptable_mssql(p_sql) RETURNING l_chk_rep
#   LET g_report = g_report_t
   RETURN l_chk_rep
   &else
   
   IF cl_null(g_report) THEN LET g_report = "dsrept" END IF
   
   LET l_table = g_properties.getValue("component_name"),"_t"

   #判斷目標報表檔是否存在
   SELECT COUNT(1) INTO l_cnt FROM ALL_TAB_COLUMNS WHERE LOWER(TABLE_NAME) = l_table AND LOWER(OWNER) = g_report
   IF l_cnt = 0 THEN RETURN FALSE END IF
   
   #判斷存在報表檔欄位與設計資料是否相同  
   Let l_i = 0
   LET l_tok_table = base.StringTokenizer.create(p_sql,",")
   WHILE l_tok_table.hasMoreTokens()
      LET l_name = l_tok_table.nextToken()
      LET l_p = l_name.getIndexOf(".",1)
      LET l_p2 = l_name.getIndexOf(".",l_p+1)
      LET l_alias_name = l_name.subString(1,l_p-1)
      LET l_table_name = l_name.subString(l_p+1,l_p2-1)
      LET l_field_name = l_name.subString(l_p2+1,l_name.getLength())
      ##adz執行，直接傳ds
      CALL cl_rpt_get_column_info("ds",l_table_name,l_field_name) RETURNING l_datatype,l_length
      
      Let l_i = l_i + 1
      
      CASE
         WHEN l_datatype = "timestamp"
            LET l_ts_type= l_datatype,"%"
            SELECT COUNT(1) INTO l_cnt FROM ALL_TAB_COLUMNS
               WHERE LOWER(TABLE_NAME) = l_table AND LOWER(OWNER) = g_report
               AND LOWER(COLUMN_NAME) = l_alias_name AND LOWER(DATA_TYPE) LIKE l_ts_type
               AND COLUMN_ID = l_i
            IF l_cnt = 0 THEN RETURN FALSE END IF
          
         WHEN (l_datatype = "blob" OR l_datatype = "clob" OR l_datatype = "date" )
            SELECT COUNT(1) INTO l_cnt FROM ALL_TAB_COLUMNS
               WHERE LOWER(TABLE_NAME) = l_table AND LOWER(OWNER) = g_report
               AND LOWER(COLUMN_NAME) = l_alias_name AND LOWER(DATA_TYPE) = l_datatype
               AND COLUMN_ID = l_i
            IF l_cnt = 0 THEN RETURN FALSE END IF
            
         WHEN l_datatype = "number" OR l_datatype = "decimal" OR l_datatype = "numeric"  #250408-00016 add numeric 
            LET l_str = l_length
            LET l_p = l_str.getIndexOf(",",1)
            IF l_p > 0 THEN 
               LET l_precision = l_str.subString(1,l_p-1)
               LET l_scale = l_str.subString(l_p+1,l_str.getLength())
            ELSE #處里ent欄位的型態
               LET l_precision = l_str
               LET l_scale = "0"           
            END IF 
            SELECT COUNT(1) INTO l_cnt FROM ALL_TAB_COLUMNS
               WHERE LOWER(TABLE_NAME) = l_table AND LOWER(OWNER) = g_report
               AND LOWER(COLUMN_NAME) = l_alias_name AND LOWER(DATA_TYPE) = l_datatype
               AND COLUMN_ID = l_i AND DATA_PRECISION = l_precision AND DATA_SCALE = l_scale
            IF l_cnt = 0 THEN RETURN FALSE END IF
            
         OTHERWISE
            SELECT COUNT(1) INTO l_cnt FROM ALL_TAB_COLUMNS
               WHERE LOWER(TABLE_NAME) = l_table AND LOWER(OWNER) = g_report
               AND LOWER(COLUMN_NAME) = l_alias_name AND LOWER(DATA_TYPE) = l_datatype
               AND COLUMN_ID = l_i AND DATA_LENGTH = l_length
            IF l_cnt = 0 THEN RETURN FALSE END IF
      END CASE
   END WHILE

   SELECT COUNT(1) INTO l_cnt FROM ALL_TAB_COLUMNS WHERE LOWER(TABLE_NAME) = l_table AND LOWER(OWNER) = g_report
   IF l_cnt <> l_i THEN RETURN FALSE END IF
   
   RETURN TRUE

  RETURN TRUE
  &endif
END FUNCTION


################################################################################
# Descriptions...: 產生暫存檔
# Memo...........:
# Usage..........: CALL adzp161_create_reptable (p_sql,p_cnt,p_table)
#                  RETURNING 回传参数
# Input parameter: p_sql 
#                : p_cnt  
#                : p_table
# Return code....: 作業成功否(TRUE/FALSE)
#                : 
# Date & Author..: Janet
# Modify.........:
################################################################################
PUBLIC FUNCTION adzp161_create_reptable(p_sql,p_cnt,p_table)
   DEFINE p_sql,p_table,l_length STRING,
          p_cnt                  DEC(5,0),
          l_name                 STRING,
          l_tok_table            base.StringTokenizer,
          l_table_name           VARCHAR(100),
          l_field_name           VARCHAR(100),
          l_alias_name           VARCHAR(100),
          l_datatype             VARCHAR(100),
          l_p                    DEC(5,0),
          l_p2                   DEC(5,0),
          l_sql                  STRING,
          l_dbname               STRING

   IF cl_null(p_sql) THEN
      DISPLAY "gzgg,dzgd及dzkc中無欄位資料，無法產生temp table"
      RETURN FALSE
   END IF

   IF cl_null(g_report) THEN LET g_report = "dsrept" END IF
   CALL cl_db_connect(g_report,FALSE)
   SET CONNECTION g_report
   IF cl_null(l_dbname) THEN 
      LET l_dbname = g_dbs 
      IF cl_null(l_dbname) THEN 
         LET l_dbname = "ds"
      END IF
   END IF 

   LET g_rep_tmpname[p_cnt] = p_table

   LET l_tok_table = base.StringTokenizer.create(p_sql,",")
   WHILE l_tok_table.hasMoreTokens()
      LET l_name = l_tok_table.nextToken()
      LET l_p = l_name.getIndexOf(".",1)
      LET l_p2 = l_name.getIndexOf(".",l_p+1)
      LET l_alias_name = l_name.subString(1,l_p-1)
      LET l_table_name = l_name.subString(l_p+1,l_p2-1)
      LET l_field_name = l_name.subString(l_p2+1,l_name.getLength())

      CALL cl_rpt_get_column_info(l_dbname,l_table_name,l_field_name) RETURNING l_datatype,l_length

      IF NOT cl_null(l_sql) THEN
         LET l_sql = l_sql,","
      END IF
      IF l_datatype = "smallint" OR l_datatype = "integer" OR l_datatype = "date"
      OR l_datatype = "datetime" OR l_datatype = "blob"  OR l_datatype = "byte"       
      OR l_datatype = "int2" OR l_datatype = "int4"
      THEN
         LET l_sql = l_sql, l_alias_name CLIPPED," ",l_datatype
      ELSE
         LET l_sql = l_sql, l_alias_name CLIPPED," ",l_datatype,"(",l_length,")"
      END IF
   END WHILE
   LET l_sql = "(",l_sql,")"

   LET l_sql = "CREATE TABLE ",g_rep_tmpname[p_cnt] CLIPPED,l_sql CLIPPED
   DISPLAY "Create sql:",l_sql
   PREPARE create_prep FROM l_sql
   IF SQLCA.sqlcode THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = SQLCA.sqlcode
      LET g_errparam.extend = 'create_prep:'
      LET g_errparam.popup = TRUE
      CALL cl_err()
      DISPLAY "Create TABLE Failure!"
      DISPLAY "p_sql:",p_sql
      RETURN FALSE
   END IF

   EXECUTE create_prep
   IF SQLCA.sqlcode = '-6372' THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = 'lib-00420'
      LET g_errparam.extend = ''
      LET g_errparam.popup = TRUE
      CALL cl_err()
      DISPLAY "[ERROR] 報表暫存table使用的資料庫(dsrept)空間不足，請聯繫資料庫管理者(DBA)"
      RETURN FALSE
   END IF
   IF SQLCA.sqlcode THEN
      INITIALIZE g_errparam TO NULL
      LET g_errparam.code = SQLCA.sqlcode
      LET g_errparam.extend = 'create_table:'
      LET g_errparam.popup = TRUE
      CALL cl_err()
      DISPLAY "Create TABLE Failure!"
      DISPLAY "p_sql:",p_sql
      RETURN FALSE
   END IF

   LET l_sql = "GRANT ALL ON ",g_rep_tmpname[p_cnt] CLIPPED," TO PUBLIC"
   PREPARE grant_prep FROM l_sql
   EXECUTE grant_prep

   &ifdef is_pgs   
      DISCONNECT "dsrept"                
   &else
      CALL cl_db_disconnect_current()
   &endif   
   #SET CONNECTION g_dbs  #170903暫時先mark

   RETURN TRUE
END FUNCTION



################################################################################
# Descriptions...: drop暫存檔
# Memo...........:
# Usage..........: CALL adzp161_drop_tmptable (p_table)
#                  RETURNING 回传参数
# Input parameter: p_table
#                : 
# Return code....: 
#                : 
# Date & Author..: Janet
# Modify.........:
################################################################################
PUBLIC FUNCTION adzp161_drop_tmptable(p_table)
   DEFINE p_table LIKE type_t.chr100
   DEFINE l_sql   STRING
   DEFINE l_cnt   LIKE type_t.num10

   #MSSQL
   &ifdef is_msv 
   #切換資料庫
   IF cl_null(g_report) THEN LET g_report = FGL_GETENV("ZONE"),"_dsrept" END IF
   IF cl_null(g_rep_db) THEN LET g_rep_db = "dsrept" END IF
   #判斷目標報表檔是否存在
#   SELECT COUNT(1) INTO l_cnt FROM INFORMATION_SCHEMA.COLUMNS WHERE LOWER(TABLE_NAME) = l_table AND LOWER(TABLE_CATALOG) = g_report   #sqlserver mark
   #sqlserver add(s)
   SELECT COUNT(1) INTO l_cnt FROM user_tab_columns WHERE table_name = p_table AND owner = g_report
#   LET l_sql = "SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS ",
#               " WHERE TABLE_NAME = '",p_table,"' ",
#              "   AND TABLE_CATALOG = '",g_report,"' "
#   PREPARE adzp161_drop_tmptable_pre FROM l_sql
#   EXECUTE adzp161_drop_tmptable_pre INTO l_cnt
   #DISPLAY "Aus drop l_cnt: ",l_cnt   #Hiko mark
   #sqlserver add(e)
   IF l_cnt = 0 THEN RETURN END IF
   #切換資料庫
#   CALL cl_db_disconnect_current()
#   SET CONNECTION g_dbs  
   &else
   SELECT COUNT(1) INTO l_cnt FROM ALL_TAB_COLUMNS WHERE LOWER(TABLE_NAME) = p_table AND LOWER(OWNER) = g_report
   &endif
   IF l_cnt > 0 THEN    
       #MSSQL
       &ifdef is_msv 
      LET l_sql = "DROP TABLE ",p_table CLIPPED
&endif
       IF cl_null(g_rep_db) THEN LET g_rep_db = "dsrept." END IF
&ifdef is_ora
      CALL cl_db_connect("dsrept",FALSE)   
      SET CONNECTION "dsrept"     
      LET l_sql = "DROP TABLE ",g_rep_db CLIPPED,p_table CLIPPED
&endif
&ifdef is_pgs
      CALL cl_db_connect("dsrept",FALSE)   
      SET CONNECTION "dsrept"     
      LET l_sql = "DROP TABLE IF EXISTS ",g_rep_db CLIPPED,p_table CLIPPED    
&endif
       EXECUTE IMMEDIATE l_sql
#&ifdef is_pgs   
      DISCONNECT "dsrept"          
      SET CONNECTION g_dbs
#&endif    
   END IF 
END FUNCTION
#170809-00051#2 add -(e)

#MSSQL add -(s)
################################################################################
# Descriptions...: 檢查目標暫存檔狀態
#                : 若不存在則創建
#                : 若存在格式不同的暫存檔則DROP後再創建
# Memo...........: 170809-00051#2
# Usage..........: CALL adzp161_check_reptable_mssql ()
#                  RETURNING TRUE/FALSE
# Input parameter: p_sql
#                : p_cnt
#                : p_table
# Return code....: 作業成功否(TRUE/FALSE)
# Date & Author..: Janet
# Modify.........:
################################################################################
PUBLIC FUNCTION adzp161_check_reptable_mssql(p_sql)
   DEFINE p_sql         STRING
   DEFINE l_table       LIKE type_t.chr100
   DEFINE l_cnt         LIKE type_t.num5
   DEFINE l_i           LIKE type_t.num5
   DEFINE l_tok_table   base.StringTokenizer
   DEFINE l_name        STRING
   DEFINE l_p           LIKE type_t.num5
   DEFINE l_p2          LIKE type_t.num5
   DEFINE l_alias_name  LIKE type_t.chr100
   DEFINE l_table_name  LIKE type_t.chr100
   DEFINE l_field_name  LIKE type_t.chr100
   DEFINE l_datatype    LIKE type_t.chr100
   DEFINE l_length      LIKE gztd_t.gztd008
   DEFINE l_str         STRING
   DEFINE l_precision   LIKE gztd_t.gztd008
   DEFINE l_scale       LIKE gztd_t.gztd008
   DEFINE l_ts_type     LIKE type_t.chr100      #timestamp要另外處理欄位名
DEFINE ls_sql STRING   #sqlserver add   
   
   IF cl_null(g_report) THEN LET g_report = FGL_GETENV("ZONE"),"_dsrept" END IF
   IF cl_null(g_rep_db) THEN LET g_rep_db = "dsrept" END IF

   
   LET l_table = g_properties.getValue("component_name"),"_t"
   #DISPLAY "Aus l_table: ",l_table #Hiko mark
   #DISPLAY "Aus table_catalog: ",g_report   
   #切換資料庫
   #sqlserver mark(s)
#   CALL cl_db_connect(g_report,FALSE)
#   SET CONNECTION g_report
   #sqlserver mark(e)
   #sqlserver add(s)
   #DISPLAY "reptable_mssql connect before" #Hiko mark
   CALL cl_db_connect(g_rep_db,FALSE)
   #DISPLAY "reptable_mssql connect after"   
#   SET CONNECTION g_rep_db
   #sqlserver add(e)
   #判斷目標報表檔是否存在
#   SELECT COUNT(1) INTO l_cnt FROM INFORMATION_SCHEMA.COLUMNS WHERE LOWER(TABLE_NAME) = l_table AND LOWER(TABLE_CATALOG) = g_report   #sqlserver mark
   SELECT COUNT(1) INTO l_cnt FROM user_tab_columns WHERE table_name = l_table AND owner = g_report
#   LET ls_sql = "SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS ",
#                " WHERE LOWER(TABLE_NAME) = '",l_table,"' ",
#                "   AND LOWER(TABLE_CATALOG) = '",g_report,"' "
#   PREPARE adzp161_chk_reptbl_mssql_cnt_pre FROM ls_sql
#   EXECUTE adzp161_chk_reptbl_mssql_cnt_pre INTO l_cnt
   #DISPLAY "Aus l_cnt: ",l_cnt    #Hiko mark

   IF l_cnt = 0 THEN RETURN FALSE END IF
   
   #組判斷SQL段落
   LET l_ts_type= "datetime2%"
   
                
   #判斷存在報表檔欄位與設計資料是否相同
   LET l_i = 0
   LET l_tok_table = base.StringTokenizer.create(p_sql,",")
   WHILE l_tok_table.hasMoreTokens()
      LET l_name = l_tok_table.nextToken()
      LET l_p = l_name.getIndexOf(".",1)
      LET l_p2 = l_name.getIndexOf(".",l_p+1)
      LET l_alias_name = l_name.subString(1,l_p-1)
      LET l_table_name = l_name.subString(l_p+1,l_p2-1)
      LET l_field_name = l_name.subString(l_p2+1,l_name.getLength())
      ##adz執行，直接傳ds
      CALL cl_rpt_get_column_info("ds",l_table_name,l_field_name) RETURNING l_datatype,l_length
      
      Let l_i = l_i + 1
      
      CASE
         WHEN l_datatype = "datetime2"
            LET l_ts_type= l_datatype,"%"
#            SELECT COUNT(1) INTO l_cnt FROM INFORMATION_SCHEMA.COLUMNS
#               WHERE LOWER(TABLE_NAME) = l_table AND LOWER(TABLE_CATALOG) = g_report
#               AND LOWER(COLUMN_NAME) = l_alias_name AND LOWER(DATA_TYPE) LIKE l_ts_type
#               AND ORDINAL_POSITION = l_i
#            LET ls_sql = "SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS ",
#                " WHERE LOWER(TABLE_NAME) = '",l_table,"' ",
#               "   AND LOWER(TABLE_CATALOG) = '",g_report,"' ",
#               "   AND LOWER(COLUMN_NAME) = '",l_alias_name,"' ",
#               "   AND LOWER(DATA_TYPE) LIKE '",l_ts_type,"' ",
#               "   AND ORDINAL_POSITION = '",l_i,"' "
#           PREPARE adzp161_chk_reptbl_mssql_date_pre FROM ls_sql
#           EXECUTE adzp161_chk_reptbl_mssql_date_pre INTO l_cnt
            SELECT COUNT(1) INTO l_cnt FROM user_tab_columns
             WHERE table_name = l_table AND owner = g_report
               AND column_name = l_alias_name AND data_type LIKE l_ts_type
               AND column_id = l_i
            IF l_cnt = 0 THEN RETURN FALSE END IF
          
         WHEN (l_datatype = "int" OR l_datatype = "smallint" OR l_datatype = "decimal")
            LET l_str = l_length
            LET l_p = l_str.getIndexOf(",",1)
            IF l_p > 0 THEN 
               LET l_precision = l_str.subString(1,l_p-1)
               LET l_scale = l_str.subString(l_p+1,l_str.getLength())
            ELSE #處理ent欄位的型態
               LET l_precision = l_str
               LET l_scale = "0"           
            END IF 
            #sqlserver mark(s)
#            SELECT COUNT(1) INTO l_cnt FROM INFORMATION_SCHEMA.COLUMNS
#               WHERE LOWER(TABLE_NAME) = l_table AND LOWER(TABLE_CATALOG) = g_report
#               AND LOWER(COLUMN_NAME) = l_alias_name AND LOWER(DATA_TYPE) = l_datatype
#               AND ORDINAL_POSITION = l_i AND NUMERIC_PRECISION = l_precision AND NUMERIC_SCALE = l_scale
            #sqlserver mark(e)
#           LET ls_sql = "SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS ",
#                " WHERE LOWER(TABLE_NAME) = '",l_table,"' ",
#               "   AND LOWER(TABLE_CATALOG) = '",g_report,"' ",
#               "   AND LOWER(COLUMN_NAME) = '",l_alias_name,"' ",
#               "   AND LOWER(DATA_TYPE) LIKE '",l_datatype,"' ",
#               "   AND ORDINAL_POSITION = '",l_i,"' ",
#               "   AND NUMERIC_PRECISION = '",l_precision,"' ",
#               "   AND NUMERIC_SCALE = '",l_scale,"' "
#           PREPARE adzp161_chk_reptbl_mssql_int_pre FROM ls_sql
#           EXECUTE adzp161_chk_reptbl_mssql_int_pre INTO l_cnt 
            SELECT COUNT(1) INTO l_cnt FROM user_tab_columns
             WHERE table_name = l_table AND owner = g_report
               AND column_name = l_alias_name AND data_type = l_datatype
               AND column_id = l_i AND data_precision = l_precision AND data_scale = l_scale
            IF l_cnt = 0 THEN RETURN FALSE END IF
            
         OTHERWISE
#            SELECT COUNT(1) INTO l_cnt FROM INFORMATION_SCHEMA.COLUMNS
#               WHERE LOWER(TABLE_NAME) = l_table AND LOWER(TABLE_CATALOG) = g_report
#               AND LOWER(COLUMN_NAME) = l_alias_name AND LOWER(DATA_TYPE) = l_datatype
#               AND ORDINAL_POSITION = l_i AND CHARACTER_MAXIMUM_LENGTH = l_length
#            LET ls_sql = "SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS ",
#                " WHERE LOWER(TABLE_NAME) = '",l_table,"' ",
#               "   AND LOWER(TABLE_CATALOG) = '",g_report,"' ",
#               "   AND LOWER(COLUMN_NAME) = '",l_alias_name,"' ",
#               "   AND LOWER(DATA_TYPE) LIKE '",l_datatype,"' ",
#               "   AND ORDINAL_POSITION = '",l_i,"' ",
#               "   AND CHARACTER_MAXIMUM_LENGTH = '",l_length,"' "
#           PREPARE adzp161_chk_reptbl_mssql_other_pre FROM ls_sql
#           EXECUTE adzp161_chk_reptbl_mssql_other_pre INTO l_cnt
            SELECT COUNT(1) INTO l_cnt FROM user_tab_columns
             WHERE table_name = l_table AND owner = g_report
               AND column_name = l_alias_name AND data_type = l_datatype
               AND column_id = l_i AND data_length = l_length
            IF l_cnt = 0 THEN RETURN FALSE END IF
      END CASE
   END WHILE
   
#   SELECT COUNT(1) INTO l_cnt FROM INFORMATION_SCHEMA.COLUMNS WHERE LOWER(TABLE_NAME) = l_table AND LOWER(TABLE_CATALOG) = g_report
#   LET ls_sql = "SELECT COUNT(1) FROM INFORMATION_SCHEMA.COLUMNS ",
#                " WHERE LOWER(TABLE_NAME) = '",l_table,"' ",
#               "   AND LOWER(TABLE_CATALOG) = '",g_report,"' "
#           PREPARE adzp161_chk_reptbl_mssql_comp_pre FROM ls_sql
#           EXECUTE adzp161_chk_reptbl_mssql_comp_pre INTO l_cnt
   SELECT COUNT(1) INTO l_cnt FROM user_tab_columns WHERE table_name = l_table AND owner = g_report
   IF l_cnt <> l_i THEN RETURN FALSE END IF
   
   #切換資料庫
#   CALL cl_db_disconnect_current()
#   SET CONNECTION g_dbs  
DISPLAY "Aus return true"
  RETURN TRUE

END FUNCTION

