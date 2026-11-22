# code_version.: 5.25.02-2012.04.16(0000-0001)
#
#170906-00011#1 by ka0132 i05樣板產生CREATE TEMP TABLE時參照型態改採用該欄位的實體欄位而非VARCHAR(500)
#                         進入修改狀態時由原本的畫面上第一個單頭欄位調整為畫面的第一個"可輸入"的單頭欄位
#170915-00032#1 by ka0132 給予預設語系g_lang = zh_TW
#171204-00032#1 by ka0132 進入修改狀態時改為:
#                         1. 第一個pk, 當第一個pk為不可輸入時改走2
#                         2. 第一個可輸入的欄位
#180108-00012#1 by yscheng 新增紀錄ERP版本的變數
#180411-00062#1 by yscheng 多語言開窗段自動加入update_item的add-point中
#180611-00048#1 by yscheng ITM類程式產生器(adzp150)偵測到作業代碼第四碼為Q者，須加上datarefresh的功能
#181128-00048#1 by 00824 3.0功能回追1.0環境(1.BPM功能調整 2.多語言開窗不區分版本)
#181211-00007#1 by 02667 調整i02/t02的foreach段,取消open段
#190108-00038#1 by 00824 簡易畫面功能
#                        註：因為簡易畫面是透過設定而來，所以有可能標準化面已客製，但簡易畫面仍為標準的狀況
#                            所以簡易畫面開啟時，需以dzzc_t中紀錄的實際模組(dzzc004)為主
#190412-00022#1 by 00824 資料所有者與所有部門改為呼叫FUNCTION取得
#190826-00045#1 by 02667 修正ref參照型態
#190806-00019#1 by 00824 修改前須先確認該筆資料是否存在
#191101-00032#1 by 00824 判斷資料是否存在的動作只需要在主程式做即可
#200804-00006#1 by 00824 postgre功能回追
#201109-00009#1 by 00824 模糊查詢功能
#210120-00006#1 by 00824 修正i02與t02樣板在set_pk_array段，判斷是否為空的方式
#201109-00009#2 by 00824 模糊查詢-二次篩選功能
#211006-00002#1 by 05598 修正t02樣板在新增時會修改到無修改權限的資料 
#230107-00001#1 by 00845 處理CONSTRUCT於tab欄位資料下放到fieldchk
#240930-00022#1 by 00824 因應項標中心設計器優化需求，配合調整相關工具及樣板


IMPORT os
SCHEMA ds

#+ code_id......: adzp150.4gl
#+ descriptions.: 程式資料建立
#+ code_creator.: hjwang
 
GLOBALS "../../cfg/top_global.inc"
GLOBALS "../4gl/adzp150_global.inc"   #180108-00012
 
GLOBALS
   DEFINE g_properties   om.SaxAttributes
   CONSTANT li_space = 3
   DEFINE gdnode_all     om.DomNode
   DEFINE g_update       STRING
   DEFINE g_type         STRING   
   DEFINE g_type_t       STRING
   DEFINE g_simple_type  LIKE type_t.chr1   #190108-00038
   #240930-00022 ---start---
   DEFINE g_value_formula DYNAMIC ARRAY OF RECORD
                         type         LIKE type_t.chr20,    #value / formula
                         location     LIKE type_t.chr20,
                         page_num     LIKE type_t.num5,
                         target_col   LIKE type_t.chr100,   #目的欄位 #xmdc_t.xmdcud001
                         target_var_col LIKE type_t.chr100, #目標控件轉化為變數
                         ctrol_col    LIKE type_t.chr100,   #使用控件欄位(原始) ex:xmdc_t.xmdc011
                         org_col      LIKE type_t.chr100,   #欄位代號 ex:xmdc011
                         var_col      LIKE type_t.chr100    #轉化成變數型態的控件欄位 ex:g_xmdc_d[l_ac].xmdc011
                         END RECORD
   DEFINE g_target_col   DYNAMIC ARRAY OF RECORD
                         ctrol_col   LIKE type_t.chr100,
                         var_col     LIKE type_t.chr100
                         END RECORD
   #240930-00022 --- end ---
END GLOBALS

#抓取執行時間用
DEFINE g_time        DATETIME YEAR TO SECOND 
DEFINE g_statetime   INTEGER
DEFINE g_endtime     INTEGER
DEFINE g_alltime     INTEGER

DEFINE g_dzbb RECORD 
         prog_name    STRING,
         point_name   STRING,
         point_ver    STRING,
         addpoint     STRING
         END RECORD
         
DEFINE g_master_chk   BOOLEAN
DEFINE g_detail_chk   BOOLEAN

DEFINE g_value        STRING
CONSTANT cs_prog_ver = 1.0

MAIN
   DEFINE lddoc_all      om.DomDocument
   DEFINE ls_filename    STRING
   DEFINE ls_prog        STRING
   DEFINE ls_path        STRING
   DEFINE ls_cmd         STRING
   DEFINE ls_mod         STRING
   DEFINE li_ver         INTEGER
   DEFINE li_cnt         INTEGER
   DEFINE lr_dzbc        RECORD LIKE dzbc_t.*
   DEFINE lc_dzax003     LIKE dzax_t.dzax003 #是否為free style
   DEFINE ls_wmode       STRING

  
   LET g_lang = "zh_TW" #170915-00032#1
 
   DISPLAY "adzp150開始運作時間:",cl_get_current()
   
   #開始走訪XML
   CALL cl_db_connect("ds",FALSE)
   
   #是否覆蓋原有add-point
   LET g_update = ARG_VAL(3)
   IF cl_null(g_update) THEN
      LET g_update = 'n'
   END IF
   
   LET li_ver = ARG_VAL(4)

   #180108-00012 ---start---
   #取得ERP版本
   LET g_erp_ver = FGL_GETENV("ERPVER")
   #180108-00012 --- end ---
   
   #顯示外部參數訊息
   DISPLAY "開始執行程式產生器, 以下為產生參數:"
   DISPLAY " 1. 模組名稱:",ARG_VAL(1)
   DISPLAY " 2. 程式名稱:",ARG_VAL(2)
   DISPLAY " 3. 程式add-point是否覆蓋:",g_update
   DISPLAY " 4. 程式版次:",ARG_VAL(4)
   DISPLAY " 5. 程式客製標示:",ARG_VAL(5)
   DISPLAY " 6. 產出模式:",ARG_VAL(6)
   
   #檢查傳入參數是否正確或正確, 若不正確則中斷
   #檢查1 - 模組別
   IF cl_null(ARG_VAL(1)) THEN
      DISPLAY "ERROR(參數1):未傳入程式模組類別!"
      EXIT PROGRAM
   END IF 
   
   #檢查2 - 程式名稱
   IF cl_null(ARG_VAL(2)) THEN
      DISPLAY "ERROR(參數2):未傳入程式名稱!"
      EXIT PROGRAM
   END IF 
   
   #檢查4 - 程式版次
   IF cl_null(ARG_VAL(4)) THEN
      DISPLAY "ERROR(參數4):未傳入程式版次!"
      EXIT PROGRAM
   END IF 
   
   #檢查5 - 程式客製標示
   IF cl_null(ARG_VAL(5)) THEN
      DISPLAY "ERROR(參數4):未傳入程式客製標示或標示錯誤(標示僅為s,c)!"
      EXIT PROGRAM
   END IF 
   IF NOT (ARG_VAL(5) = 's' OR ARG_VAL(5) = 'S' OR 
      ARG_VAL(5) = 'c' OR ARG_VAL(5) = 'C' ) THEN
      DISPLAY "ERROR(參數5):傳入程式客製標示錯誤(標示僅為s或c)!"
      EXIT PROGRAM
   END IF 
   
   #檢查6 - section處理模式
   LET ls_wmode = ARG_VAL(6)
   IF cl_null(ls_wmode) THEN
      LET ls_wmode = '1'
   END IF
   IF NOT (ls_wmode = '1'  OR 
           ls_wmode = '2'  OR
           ls_wmode = '3') THEN
      DISPLAY "ERROR(參數6):傳入的section處理模式不符規定, 須為1或2或3!"
      EXIT PROGRAM
   END IF 
   IF ls_wmode = '2' THEN
      DISPLAY '進行預覽模式4gl2產出!'
   END IF

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
   
   #確定是否為free style   
   SELECT dzax003 INTO lc_dzax003 FROM dzax_t 
    WHERE dzax001 = lr_dzbc.dzbc001 AND
          dzax006 = lr_dzbc.dzbc007
     
   #先判斷是否解開section, 進行解析與寫入
   IF li_cnt > 0 AND lc_dzax003 = 'N' AND ls_wmode = '1' THEN
      DISPLAY '該程式(',lr_dzbc.dzbc001,')已解開Section調整, 不再產生標準Section!'
   ELSE
      
      #若未解開Section則強制將ls_wmode改為1
      IF li_cnt = 0 AND ls_wmode <> '1' THEN
         LET ls_wmode = '1'
         DISPLAY "WARNING:該程式未解開Section, section處理模式改正為1!"
      END IF
    
      #僅有第一版次能夠覆蓋add-point
      IF g_update = 'y' AND li_ver > 1 THEN
         LET g_update = 'n'
         DISPLAY "WARNING:該程式非第一版次, 無法覆蓋add-point!"
      END IF
   
      #150開始
      CALL adzp150_start()
      
      LET ls_mod = ARG_VAL(1)
      #c開頭為客製模組, 其餘為一般模組
      IF ls_mod = 'lib'  OR 
         ls_mod = 'sub'  OR 
         ls_mod = 'wss'  OR 
         ls_mod = 'qry'  OR 
         ls_mod = 'clib' OR 
         ls_mod = 'csub' OR 
         ls_mod = 'cwss' OR 
         ls_mod = 'cqry' THEN
         LET ls_path = "COM"
      ELSE      
         LET ls_path = "ERP"
      END IF
      
      #tabx檔更正為tab檔
      LET ls_filename = os.Path.join(os.Path.join(FGL_GETENV(ls_path),ARG_VAL(1)),"dzx")
      LET ls_filename = os.Path.join(ls_filename,"tab")
      LET ls_filename = os.Path.join(ls_filename,ARG_VAL(2) CLIPPED||".tab")
      
      #定義轉換用SaxAttributes 
      LET g_properties = om.SaxAttributes.create()
      CALL g_properties.addAttribute("app_id",ARG_VAL(2)) #app_id 程式代號
      CALL g_properties.addAttribute("module",ARG_VAL(1)) #module 
      
      LET lddoc_all = om.DomDocument.createFromXmlFile(ls_filename)  #om.DomDocument
      
      #若讀取不到資料代表tabx不存在
      TRY
         LET gdnode_all = lddoc_all.getDocumentElement()                #om.DomNode
      CATCH
         DISPLAY "ERROR(1):讀取",ls_filename,"發生錯誤, 程式中止!"
         EXIT PROGRAM
      END TRY
      
      LET gdnode_all = lddoc_all.getDocumentElement()                #om.DomNode
      
      #將add-point版本一次更新
      CALL cl_add_point_update(g_properties.getValue("app_id"),'1')
      
      #產生流程=> 讀取樣板檔案 / 同時間寫出(一進一出) / 產生器符號中斷,擺放切片或產生code
      #獲取程式基本資料 (含填寫待轉換的對照表)
      DISPLAY "程式產生器(adzp150):開始讀取tab並產生基礎資料!"
      CALL adzp150_read_basic_data()
      
      #開始進行產生/轉換動作
      DISPLAY "程式產生器(adzp150):開始讀取template並取代相關變數產生tgl!"
      CALL adzp150_read_and_replace()
      
      #進行必要欄位檢查
      CALL adzp150_parameter_chk()
      
      #將add-point版本一次更新
      CALL cl_add_point_update(g_properties.getValue("app_id"),g_properties.getValue("general_adp_ver"))
      
      #150結束
      CALL adzp150_end('adzp150')
      
      #顯示錯誤訊息
      DISPLAY g_properties.getValue("genereal_error_message")
      
      #175開始
      CALL adzp150_start()
      #呼叫adzp175重新產生
      DISPLAY "--------------------------------------------------"
      DISPLAY "開始產生tgx2檔案, adzp175開始運作時間:",cl_get_current()
      LET ls_cmd = ' adzp175 ',ARG_VAL(1),' ',ARG_VAL(2)
      CALL cl_cmdrun_wait(ls_cmd)
      #175結束
      CALL adzp150_end('adzp175')
      
      #176開始
      CALL adzp150_start()
      #呼叫adzp176寫入section
      DISPLAY "--------------------------------------------------"
      DISPLAY "開始section解析, adzp176開始運作時間:",cl_get_current()
      LET ls_cmd = ' adzp176 ',ARG_VAL(1),' ',ARG_VAL(2),' ',ARG_VAL(4),' ',ARG_VAL(5),' ',ls_wmode
      CALL cl_cmdrun_wait(ls_cmd)
      #176結束
      CALL adzp150_end('adzp176')
      
   END IF
   
   #177開始(組合tgl)
   CALL adzp150_start()
   #呼叫adzp177產出tgl
   DISPLAY "--------------------------------------------------"
   DISPLAY "開始產生tgl檔案, adzp177開始運作時間:",cl_get_current()
   LET ls_cmd = ' adzp177 ',ARG_VAL(1),' ',ARG_VAL(2),' ',ARG_VAL(4),' ',ARG_VAL(5),' ',ls_wmode
   CALL cl_cmdrun_wait(ls_cmd)
   DISPLAY "--------------------------------------------------"
   #177結束
   CALL adzp150_end('adzp177')
   
   #155開始(只針對特定樣板)
   DISPLAY "開始產生mask檔案, adzp155開始運作時間:",cl_get_current()
   LET ls_cmd = ' adzp155 ',ARG_VAL(1),' ',ARG_VAL(2), ' n '
   CALL cl_cmdrun_wait(ls_cmd)

END MAIN

#+ 取出程式基本資料 <assembly>
PRIVATE FUNCTION adzp150_read_basic_data()
   DEFINE ldnode_action om.DomNode
   DEFINE lnl_selted    om.NodeList
   DEFINE ls_err        STRING
   DEFINE ls_tmp        STRING
   DEFINE li_tmp        INTEGER
   DEFINE ls_type_list  STRING
   DEFINE ls_gzza001    LIKE gzza_t.gzza001
   DEFINE ls_app        LIKE gzza_t.gzza001
   DEFINE ls_app_s      STRING
   
   #取出om.NodeList 有關tag list
   LET lnl_selted = gdnode_all.selectByTagName("assembly")
   LET ldnode_action = lnl_selted.item(1) 

   #模組
   LET ls_tmp = ldnode_action.getAttribute("module") 
   CALL g_properties.addAttribute("general_module",ls_tmp)
   
   #tiptop版本
   LET ls_tmp = ldnode_action.getAttribute("tpver") 
   CALL g_properties.addAttribute("general_prog_ver",ls_tmp)
   
   #程式版本(PR)
   CALL g_properties.addAttribute("general_pr_ver",ARG_VAL(4))
   
   #程式版本(PR)
   CALL g_properties.addAttribute("general_sd_ver",ldnode_action.getAttribute("sdver") )
   
   #設計點版本
   #LET ls_tmp = ldnode_action.getAttribute("sdver") 
   LET ls_tmp = ARG_VAL(4)
   IF cl_null(ls_tmp) THEN
      LET ls_tmp = "1"
   END IF
   LET li_tmp = ls_tmp
   IF li_tmp = 0 OR li_tmp IS NULL THEN 
      LET li_tmp = 1 
   END IF
   CALL g_properties.addAttribute("general_adp_ver",li_tmp USING "&&&&" )

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
   
   #客製規格版次
   LET ls_tmp = ldnode_action.getAttribute("cusver") 
   CALL g_properties.addAttribute("general_adp_ver_cust",ls_tmp)
   
   #客製規格時間
   LET ls_tmp = ldnode_action.getAttribute("cusdt") 
   CALL g_properties.addAttribute("general_modify_time_cust",ls_tmp)
   
   #模組
   LET ls_tmp = ldnode_action.getAttribute("jobmode") 
   CALL g_properties.addAttribute("general_jobmode",ls_tmp)

   #行業別
   LET ls_tmp = ldnode_action.getAttribute("industry") 
   CALL g_properties.addAttribute("industry_id",ls_tmp)

   #form代碼
   LET ls_tmp = ldnode_action.getAttribute("form") 
   CALL g_properties.addAttribute("form_id",ls_tmp)
   
   #預設參數數量
   LET ls_tmp = ldnode_action.getAttribute("fix_arg") 
   CALL g_properties.addAttribute("fix_arg",ls_tmp)
   IF cl_null(ls_tmp) THEN
      LET ls_tmp = '0'
   END IF   
   CALL adzp150_arg_idx(ls_tmp)
   
   #樣板代碼
   LET ls_type_list = "i00,i01,i02,i03,i04,i05,i06,i07,i08,i09,i10,i12,i13,t01,t02,c01a,c01b,c01c,c02a,c02b,c02c,c03a,c03b,c03c,c04a,p00,p01,p02,n00,m00,r01,q00"
   LET ls_tmp = ldnode_action.getAttribute("type")
   IF ls_type_list.getIndexOf(ls_tmp,1) = 0 THEN
      DISPLAY "ERROR(2):目前尚未支持此種樣板檔(",ls_tmp,"), 請重新定義!"
      EXIT PROGRAM
   END IF
   
   #名義樣板名稱
   CALL g_properties.addAttribute("type_id_t",ls_tmp)
   LET g_type = ls_tmp
   
   CASE ls_tmp
      #c01a
      WHEN "c01a"
         LET ls_tmp = "i01"
      #c02a
      WHEN "c02a"
         LET ls_tmp = "i02"
      #c03a
      WHEN "c03a"
         LET ls_tmp = "t01"
      #c04a
      WHEN "c04a"
         LET ls_tmp = "t02"
      #共用樣板(t01有browser,i09無browser)
      WHEN "i09" 
         LET ls_tmp = "t01"
      #共用樣板(i01有browser,i10無browser)
      WHEN "i10" 
         LET ls_tmp = "i01"

      #共用樣板(i07有browser,i12無browser)
      WHEN "i12"
         LET ls_tmp = "i07" 
    
      #共用樣板(i01為一般browser,i13為六層樹狀) 
      WHEN "i13"
         LET ls_tmp = "i01" 
         
      OTHERWISE
      
   END CASE
   
   #實際使用的樣板名稱
   CALL g_properties.addAttribute("type_id",ls_tmp)
   
   DISPLAY "目前使用的程式樣板為 : code_",ls_tmp,"樣板"
   
   #目前只有雙檔(t01)/單檔多欄(i02)/假雙檔(i07)
   #樣板會使用到單身多頁籤,用CASE可避免忘記新增樣板須調整的程式段
   CASE
      WHEN ls_tmp.equals("i02")  OR ls_tmp.equals("i03")  OR ls_tmp.equals("i04") OR
           ls_tmp.equals("i07")  OR ls_tmp.equals("i11")  OR ls_tmp.equals("i12") OR
           ls_tmp.equals("t01")  OR ls_tmp.equals("t02")  OR ls_tmp.equals("i08") OR
           ls_tmp.equals("c02b") OR ls_tmp.equals("c03b") OR ls_tmp.equals("c02c") 
        
         #頁簽數 
         LET ls_tmp = ldnode_action.getAttribute("page") 
         CALL g_properties.addAttribute("page",ls_tmp)
         
      WHEN ls_tmp.equals("s01") OR ls_tmp.equals("i01") OR 
           ls_tmp.equals("i05") OR ls_tmp.equals("i06") 
 
   END CASE
   
   #確認MD5是否正確(錯誤則中斷)
   CALL sadzi100_template_chk(g_properties.getValue("type_id"),cs_prog_ver)
   
   #判斷是否為行業別程式, 再決定general_prefix
   #app_id為實際程式名稱(包含行業別), EX:azzi900_ic
   #general_prefix為排除行業別後的程式名稱, EX:azzi900
   #程式中除了section名稱外皆使用general_prefix,
   #目的是確保程式在引用時能夠取到正確的原程式段落
   LET ls_tmp = g_properties.getValue("type_id_t")
   IF ls_tmp.getIndexOf('c',1) = 0 AND ls_tmp <> 'p00' THEN
      LET ls_app_s = g_properties.getValue("app_id")
      IF ls_app_s.getIndexOf('_',1) > 0 THEN
         LET ls_app = ls_app_s.subString(1,ls_app_s.getIndexOf('_',1)-1)
         LET li_tmp = 0
         SELECT COUNT(1) INTO li_tmp FROM gzza_t 
          WHERE gzza001 = ls_app
         IF li_tmp > 0 THEN
            CALL g_properties.addAttribute("general_prefix",ls_app)
         END IF
      END IF 
   END IF
   
   IF cl_null(g_properties.getValue("general_prefix")) THEN
      CALL g_properties.addAttribute("general_prefix",g_properties.getValue("app_id"))
   END IF 

   #先取出master_tbl_name跟detail_tbl_name的值,供轉換時應用
   CALL adzp150_get_table_name(ldnode_action)
   
   CALL adzp150_xml_search(ldnode_action)

   #填寫待轉換的對照表
   CALL adzp150_read_tab()
   
END FUNCTION

#+ 填寫待轉換的對照表
PRIVATE FUNCTION adzp150_read_tab()
   DEFINE l_str_tab_master     STRING 
   DEFINE l_str_tab_detail     STRING 
   DEFINE li_cnt               LIKE type_t.num5 
   DEFINE li_s                 LIKE type_t.num5 
   DEFINE li_e                 LIKE type_t.num5 
   DEFINE li_page              LIKE type_t.num5 
   DEFINE ls_column_chk_body   STRING
   DEFINE ls_column_ctrlp_body STRING
   DEFINE ls_page              STRING
   DEFINE ls_tmp               STRING
   DEFINE ls_tmp2              STRING
   DEFINE li_tmp               LIKE type_t.num5 
   DEFINE ls_err               STRING
   DEFINE l_chklist            om.NodeList
   DEFINE l_chknode            om.DomNode
   DEFINE l_nlist              om.NodeList
   DEFINE l_node               om.DomNode
   DEFINE l_chkn               INTEGER
   DEFINE l_buf                base.StringBuffer
   DEFINE ls_name              STRING
   DEFINE ls_value             STRING
   DEFINE lst_token            base.StringTokenizer
   DEFINE ls_token             STRING 
   
   #先行寫入
   LET g_dzbb.prog_name   = g_properties.getValue("app_id")
   LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
   LET g_dzbb.addpoint    = "#Memos"
   LET g_dzbb.point_name  = "global.memo"
   CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))     
   
   #ent/site特殊欄位相關處理
   CALL adzp150_append_field()
   
   #取得所有可輸入欄位資料以便後續使用
   LET ls_tmp = ""
   LET l_chklist = gdnode_all.selectByTagName("input")
   FOR l_chkn = 1 TO l_chklist.getLength()
      LET l_chknode = l_chklist.item(l_chkn)
      LET ls_tmp = ls_tmp,l_chknode.getAttribute("field"),","
   END FOR
   CALL g_properties.addAttribute("general_input_field",ls_tmp)
   LET ls_tmp = ""
   
   #取得所有可查詢欄位資料以便後續使用
   LET ls_tmp = ""
   LET l_chklist = gdnode_all.selectByTagName("construct")
   FOR l_chkn = 1 TO l_chklist.getLength()
      LET l_chknode = l_chklist.item(l_chkn)
      LET ls_tmp = ls_tmp,l_chknode.getAttribute("field"),","
      #重新定義個別table欄位 #230107-00001#1
      IF l_chknode.getAttribute("id") = "head" THEN
         CALL g_properties.addAttribute("eachone_construct_field",l_chknode.getAttribute("field"))
      ELSE
         CALL g_properties.addAttribute("eachone_construct_field"||l_chknode.getAttribute("page"),
                                                                   l_chknode.getAttribute("field"))
        #DISPLAY 'fieldchk:',"eachone_construct_field"||l_chknode.getAttribute("page"),'->',
        #                                               l_chknode.getAttribute("field")
      END IF
   END FOR
   CALL g_properties.addAttribute("general_construct_field",ls_tmp)
   LET ls_tmp = ""

   IF NOT cl_null(ls_err) THEN 
      DISPLAY "ERROR(3):",ls_err LET ls_err="" 
   END IF

   #特定樣板的特定段落執行
   IF adzp150_type_decide("m") THEN

      #特殊操作欄位代換
      LET l_str_tab_master = g_properties.getValue("master_tbl_prefix") 
      
      #單頭欄位控制
      CALL g_properties.addAttribute("location","head" )
      
      #stus
      CALL g_properties.addAttribute("master_field_stus", l_str_tab_master||"stus" )
      CALL g_properties.addAttribute("master_var_stus", g_properties.getValue("master_var_title")||"."||l_str_tab_master||"stus")       #${var_stus}
      
      #欄位單頭檢查(含after,before,controlp)
      CALL g_properties.addAttribute("master_fields_check",  adzp150_create_chk_field("0",3,"i") )
      CALL g_properties.addAttribute("master_fields_ctrlp_i",adzp150_create_chk_field("0",3,"ic") )
      CALL g_properties.addAttribute("master_fields_ctrlp",  adzp150_create_chk_field("0",3,"c") )

      #create inner join for multi master table
      CALL adzp150_master_inner_join()
      
      #reference填充到add point
      LET g_dzbb.prog_name   = g_properties.getValue("app_id")
      LET g_dzbb.point_name  = "show.head.reference"
      IF g_properties.getValue("type_id") = "i07" THEN
         #i07使用的add-point名稱不同
         LET g_dzbb.point_name  = "ref_show.head.reference"
      END IF
      LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
      LET g_dzbb.addpoint    = g_properties.getValue("master_vars_reference")
      CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))     
   
      #210120-00006 ---start---
      #若是單檔多欄的型態，應改為判斷陣列長度是否>0
   #  IF g_properties.getValue("type_id") = "i02" OR
   #     g_properties.getValue("type_id") = "t02" THEN
   #     CALL g_properties.addAttribute("mdl_mark_m","#該樣板不需此段落")
   #     CALL g_properties.addAttribute("mdl_mark_d","")
   #     IF g_properties.getValue("type_id") = "t02" THEN
   #        CALL g_properties.addAttribute("mdl_detail_flag","g_detail_idx")
   #     ELSE
   #        CALL g_properties.addAttribute("mdl_detail_flag","l_ac")
   #     END IF
   #  ELSE
         CALL g_properties.addAttribute("mdl_mark_d","#該樣板不需此段落")
         CALL g_properties.addAttribute("mdl_mark_m","")
         CALL g_properties.addAttribute("mdl_detail_flag","")
   #  END IF
      #210120-00006 --- end ---

      #產生對應的相關文件action
      LET ls_name  = "master_set_pk_array"
      LET ls_value = adzp150_related_document('','m','a51')
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      LET ls_name  = "master_related_document"
      LET ls_value = adzp150_related_document('dialog','m','a46')
      CALL g_properties.addAttribute(ls_name,ls_value)

      IF g_properties.getValue("type_id") = "i01" OR 
         g_properties.getValue("type_id") = "i05" THEN
         LET ls_name  = "master_related_document_menu"
         LET ls_value = adzp150_related_document('menu','m','a46')
         CALL g_properties.addAttribute(ls_name,ls_value)
      END IF
      
      #刪除相關文件
      LET ls_name  = "master_related_document_remove"
      LET ls_value = adzp150_related_document('','m','a47')
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #顯示相關文件的action圖樣
      LET ls_name  = "master_related_document_show"
      LET ls_value = adzp150_related_document('','m','a48')
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #組refresh用SQL
      LET ls_name  = "master_refresh"
      CALL adzp150_trans_input_fields() #會用到所以先處理
      LET ls_value = adzp150_master_referesh()
      CALL g_properties.addAttribute(ls_name,ls_value)
      
      #190806-00019 ---start---
      #確認主表該筆資料是否存在
      IF g_properties.getValue("type_id") = "i01" OR
         g_properties.getValue("type_id") = "i04" OR
         g_properties.getValue("type_id") = "i05" OR
         g_properties.getValue("type_id") = "i07" OR
         g_properties.getValue("type_id") = "i08" OR
         g_properties.getValue("type_id") = "t01" THEN

         #191101-00032 ---start---
         #僅檢查主程式，因此須將子程式排除
         #因為有的子程式在tab中雖是紀錄用子程式樣板，但實際使用的樣板可能為t01(例如：azzi309_01,azzi800_03)
         #因此需再判斷tab紀錄的樣板
         LET ls_tmp = g_properties.getValue("type_id_t")
         IF ls_tmp.getIndexOf("c",1) > 0 THEN
         ELSE
         #191101-00032 --- end ---
            LET ls_name  = "general_call_chk_data_exist_h"
            LET ls_value = "CALL ${general_prefix}_chk_data_exist(g_forupd_sql)"
            CALL g_properties.addAttribute(ls_name,ls_value)
        
            LET ls_name  = "general_chk_data_exist_h"
            LET ls_value = adzp150_make_slice("a76")
            CALL g_properties.addAttribute(ls_name,ls_value)
         END IF   #191101-00032
      END IF
      #190806-00019 --- end ---
   END IF

   #單身
   IF adzp150_type_decide("d") THEN

      #設定upd_body_fields及cs_body_fields 組單身欄位頁簽
      CALL adzp150_set_upd_body_fields()

      #凍結欄位處理
      CALL g_properties.addAttribute("location","body2")
      LET ls_tmp = adzp150_create_chk_field(0,3,"c")
      CALL g_properties.addAttribute("detail_frozen_fields_ctrlp",ls_tmp)   
      
      #單身欄位控制
      CALL g_properties.addAttribute("location","body" )

      LET l_str_tab_detail = g_properties.getValue("detail_tbl_prefix") 
      
      #多table處理
      CALL adzp150_create_detail_entrance()
      #CALL adzp150_create_detail2_entrance()
      
      #針對i02,t02加入相關文件段落
      IF g_properties.getValue("type_id") = "i02" OR
         g_properties.getValue("type_id") = "t02" THEN
         
         #210120-00006 ---start---
         #若是單檔多欄的型態，應改為判斷陣列長度是否>0
         CALL g_properties.addAttribute("mdl_mark_m","#該樣板不需此段落")
         CALL g_properties.addAttribute("mdl_mark_d","")
         IF g_properties.getValue("type_id") = "t02" THEN
            CALL g_properties.addAttribute("mdl_detail_flag","g_detail_idx")
         ELSE
            CALL g_properties.addAttribute("mdl_detail_flag","l_ac")
         END IF
         #210120-00006 --- end ---

         #產生對應的相關文件action
         LET ls_name  = "detail_set_pk_array"
         LET ls_value = adzp150_related_document('','d','a51')
         CALL g_properties.addAttribute(ls_name,ls_value)
         
         LET ls_name  = "detail_related_document"
         LET ls_value = adzp150_related_document('dialog','d','a46')
         CALL g_properties.addAttribute(ls_name,ls_value)
         
         IF g_properties.getValue("type_id") = "i01" OR 
            g_properties.getValue("type_id") = "i05" THEN
            LET ls_name  = "detail_related_document_menu"
            LET ls_value = adzp150_related_document('menu','d','a46')
            CALL g_properties.addAttribute(ls_name,ls_value)
         END IF
         
         #刪除相關文件
         LET ls_name  = "detail_related_document_remove"
         LET ls_value = adzp150_related_document('','d','a47')
         CALL g_properties.addAttribute(ls_name,ls_value)
         
         #刪除相關文件(for delete func - i02,t02)
         LET ls_name  = "detail_related_document_remove_func"
         LET ls_value = adzp150_related_document('.func','d','a47')
         CALL g_properties.addAttribute(ls_name,ls_value)

         #顯示相關文件的action圖樣
         LET ls_name  = "detail_related_document_show"
         LET ls_value = adzp150_related_document('','d','a48')
         CALL g_properties.addAttribute(ls_name,ls_value)
         
      END IF
      
      #改以跑迴圈方式處理單身多頁簽方式
      LET li_page = g_properties.getValue("page")
      
      FOR li_cnt = 1 TO li_page
         LET ls_page = li_cnt USING "<<<"
         CALL g_properties.addAttribute("general_current_page",ls_page)   
         LET ls_column_chk_body = "detail_fields_check"
         LET ls_column_ctrlp_body = "detail_fields_ctrlp"
         LET ls_tmp = adzp150_create_chk_field(li_cnt,3,"i")
         CALL g_properties.addAttribute(ls_column_chk_body||ls_page,ls_tmp)   
         LET ls_tmp = adzp150_create_chk_field(li_cnt,3,"ic")
         CALL g_properties.addAttribute(ls_column_ctrlp_body||ls_page||"_i",ls_tmp)    
         LET ls_tmp = adzp150_create_chk_field(li_cnt,3,"c")
         CALL g_properties.addAttribute(ls_column_ctrlp_body||ls_page,ls_tmp)   
      END FOR 
      
      CALL adzp150_detail_ctrl_make()
      
      #reference填充到add point
      CASE g_properties.getValue("type_id")
         WHEN "i02"
            LET g_dzbb.point_name  = "detail_show.reference"
         WHEN "i07"
            LET g_dzbb.point_name  = "ref_show.body.reference"
         WHEN "t02"
            LET g_dzbb.point_name  = "detail_show.body.reference"
         OTHERWISE 
            LET g_dzbb.point_name  = "show.body.reference"
      END CASE
      LET g_dzbb.prog_name   = g_properties.getValue("app_id")
      LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
      LET g_dzbb.addpoint    = g_properties.getValue("detail_vars_reference")
      CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))     
      
      FOR li_tmp = 2 TO g_properties.getValue("page")
         LET ls_tmp = li_tmp USING "<<<"
         CASE g_properties.getValue("type_id")
            WHEN "i02"
               LET g_dzbb.point_name  = "detail_show.body",adzp150_page_trans(ls_tmp),".reference"
            WHEN "i07"
               LET g_dzbb.point_name  = "ref_show.body",adzp150_page_trans(ls_tmp),".reference"
            WHEN "t02"
               LET g_dzbb.point_name  = "detail_show.body",adzp150_page_trans(ls_tmp),".reference"
            OTHERWISE 
               LET g_dzbb.point_name  = "show.body",adzp150_page_trans(ls_tmp),".reference"
         END CASE
         LET g_dzbb.prog_name   = g_properties.getValue("app_id")
         LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
         LET g_dzbb.addpoint    = g_properties.getValue("detail_vars_reference"||ls_tmp)
         CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))     
      END FOR
      
      #產生對應的相關文件action
      #LET ls_name  = "detail_related_document"
      #LET ls_value = adzp150_related_document('dialog','d')
      #CALL g_properties.addAttribute(ls_name,ls_value)
      
   END IF 
   
   #TREE
   IF adzp150_type_decide("t") THEN
   
      #樹狀欄位控制
      CALL g_properties.addAttribute("location","tree" )
      
      #產出reference段落
      LET l_nlist = gdnode_all.selectByTagName("bs_reference")
      FOR li_cnt = 1 TO l_nlist.getLength()
         LET l_node = l_nlist.item(li_cnt)
         CALL adzp150_create_reference(l_node,2) RETURNING ls_tmp
      END FOR
      
      LET l_buf = base.StringBuffer.create()
      LET ls_tmp = g_properties.getValue("browser_fields_all")
      CALL l_buf.append(ls_tmp)
      CALL l_buf.replace(g_properties.getValue("master_field_lid"), g_properties.getValue("master_field_pid"), 0)
      LET ls_tmp = l_buf.toString()
      CALL g_properties.addAttribute("browser_fields_expand", ls_tmp )
     
   END IF
   
   #Browser
   IF adzp150_type_decide("b") THEN
   
      #browser欄位控制
      CALL g_properties.addAttribute("location","browser" )
      
      #產出reference段落
      LET l_nlist = gdnode_all.selectByTagName("bs_reference")
      FOR li_cnt = 1 TO l_nlist.getLength()
         LET l_node = l_nlist.item(li_cnt)
         CALL adzp150_create_reference(l_node,2) RETURNING ls_tmp
      END FOR
      
      #filter段落(真正有顯示browser的樣板才需要)
      IF g_type = 'i01' OR g_type = 'i07' OR g_type = 't01' THEN
         CALL g_properties.addAttribute("browser_filter_action",adzp150_make_slice("a49"))
         CALL g_properties.addAttribute("browser_filter_funcs",adzp150_make_slice("a50"))
      END IF
    
      #重新組合browser_reference_join_sql 
      LET ls_value = g_properties.getValue("browser_reference_join_sql")
      LET ls_tmp  = g_properties.getValue("browser_join_table")
      LET ls_tmp2 = g_properties.getValue("master_append_join")
      LET li_s = 1
      LET li_e = 1
      WHILE li_e <> ls_tmp2.getLength() AND NOT cl_null(ls_tmp2)
         LET li_s = ls_tmp2.getIndexOf('LEFT JOIN',li_e) + 10
         LET li_e = ls_tmp2.getIndexOf('LEFT JOIN',li_s) - 1
         IF li_e = -1 THEN
            LET li_e = ls_tmp2.getLength()
         END IF
         LET ls_token = ls_tmp2.subString(li_s,li_e)
         #抓出table名稱
         LET ls_tmp = ls_token.subString(1,ls_token.getIndexOf('ON',1)-1) 
         #確認是否出現在browser_reference_join_sql
         IF ls_value.getIndexOf(ls_tmp,1) = 0 THEN
            #如果lang join的table 不存在於browser_join_table, 則加入
            LET ls_value = ls_value, "\n",
                           "               \" LEFT JOIN ", ls_token,"\", "
         END IF
         
      END WHILE 
      CALL g_properties.addAttribute("browser_reference_join_sql",ls_value)
      
   END IF
   
   #browser_field_stus
   LET ls_tmp = g_properties.getValue("master_tbl_prefix"),"stus"
   LET ls_tmp2 = g_properties.getValue("master_tbl_name")
   IF cl_getField(ls_tmp2, ls_tmp) THEN
      CALL g_properties.addAttribute("browser_field_stus","t0."||ls_tmp||",")
   ELSE
      CALL g_properties.addAttribute("browser_field_stus","'',")
   END IF
   
   #特別註解處理
   CALL adzp150_mark()
   
   #多語言檔處理(單頭)
   LET ls_tmp = g_properties.getValue("master_fields_lang")
   IF NOT cl_null(ls_tmp) THEN
      CALL g_properties.addAttribute("master_multi_language", adzp150_master_multilang(ls_tmp))
   END IF

   #多語言檔處理(單身)
   LET ls_tmp = g_properties.getValue("detail_fields_lang")
   IF NOT cl_null(ls_tmp) THEN
      CALL g_properties.addAttribute("detail_multi_language", adzp150_detail_multilang(ls_tmp))
   END IF
   
   #多語言備份資料額外處理(單頭)
   LET ls_tmp = g_properties.getValue("master_multi_table_define")
   LET ls_tmp = ls_tmp.subString(1,ls_tmp.getLength()-2)
   IF NOT cl_null(ls_tmp) THEN
      LET ls_tmp = "DEFINE g_master_multi_table_t    RECORD\n",
                   ls_tmp,
                   "\n      END RECORD"
   END IF
   IF cl_null(ls_tmp) THEN
      CALL g_properties.addAttribute("master_multi_table_define", "#該樣板不需此段落")
   ELSE
      CALL g_properties.addAttribute("master_multi_table_define", ls_tmp)
   END IF
   
   #多語言備份資料額外處理(單身)
   LET ls_tmp = g_properties.getValue("detail_multi_table_define")
   LET ls_tmp = ls_tmp.subString(1,ls_tmp.getLength()-2)
   IF NOT cl_null(ls_tmp) THEN
      LET ls_tmp = "DEFINE g_detail_multi_table_t    RECORD\n",
                   ls_tmp, "\n      END RECORD"
   END IF
   
   IF cl_null(ls_tmp) THEN
      CALL g_properties.addAttribute("detail_multi_table_define", "#該樣板不需此段落")
   ELSE
      CALL g_properties.addAttribute("detail_multi_table_define", ls_tmp)
   END IF
   
   #state切換資料組成
   CALL adzp150_create_stus()

   #180611-00048 ---start---
   #產生datarefresh的ON ACTION段落
   CALL adzp150_create_action_datarefresh()
   #180611-00048 --- end ---
   
   #190108-00038 ---start---
   #簡易畫面相關處理
   CALL adzp150_simple_mode()
   #190108-00038 --- end ---

   #產生slice相關段落(特定樣板使用)
   CALL adzp150_create_slice()
  
   #公用欄位處理&組成
   CALL adzp150_create_common()
   
   #遮罩相關處理
   CALL adzp150_create_all_mask()
   
   #p01,r01相關處理
   CALL adzp150_make_dialog()
   
   #訊息中心相關
   CALL adzp150_msgcentre_notify()
   
END FUNCTION

#+ 程式樣板讀取/程式寫出(一進一出) / 產生器符號中斷,擺放切片或產生code
PRIVATE FUNCTION adzp150_read_and_replace()
   DEFINE lchannel_read           base.Channel
   DEFINE lchannel_write          base.Channel
   DEFINE ls_readline             STRING
   DEFINE ls_text                 STRING
   DEFINE ls_code_filename        STRING
   DEFINE ls_sample_filename      STRING
   DEFINE lchannel_check          base.Channel
   DEFINE ls_mdlpath              STRING

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
      DISPLAY "ERROR(4):樣板檔案:",ls_sample_filename.trim()," 不存在!"
      EXIT PROGRAM
   END IF
   CALL lchannel_read.openFile( ls_sample_filename CLIPPED, "r" )

   #產出程式路徑
   LET ls_code_filename = os.Path.join(FGL_GETENV(UPSHIFT(g_properties.getValue("module"))),"dzx")
   LET ls_code_filename = os.Path.join(ls_code_filename,"tgl")
   LET ls_code_filename = os.Path.join(ls_code_filename,ARG_VAL(2) CLIPPED||".tgx")
   
   #先行移除tgl
   IF os.Path.delete(ls_code_filename) THEN
      DISPLAY "刪除舊檔案:",ls_code_filename
   END IF
   
   #判斷是否砍除成功
   IF NOT os.Path.exists(ls_code_filename) THEN
      DISPLAY "舊檔案刪除成功:",ls_code_filename
   ELSE
      DISPLAY "Error:舊檔案刪除失敗:",ls_code_filename
      EXIT PROGRAM
   END IF
   
   DISPLAY "產生檔位置:",ls_code_filename
   CALL lchannel_write.openFile( ls_code_filename, "w" )

   #產生程式版本及說明
   CALL lchannel_write.write(adzp150_prog_memo())

   #讀取及轉換
   WHILE TRUE
   
      LET ls_text = ""
      LET ls_readline = lchannel_read.readLine()
      
      #讀取到版本資料時進行回寫gzzx
      IF ls_readline.getIndexOf('樣板自動產生',1) > 0 THEN
         CALL adzp150_update_gzzx(ls_readline)
      END IF
        
      IF lchannel_read.isEof() THEN
         EXIT WHILE
      END IF

      #產生code部分
      CASE
         #page段落產生
         WHEN ls_readline.getIndexOf("#pages - Start -",1) > 0
            CALL adzp150_make_pages(lchannel_read,lchannel_write,0) RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
         #page段落產生(input段落)
         WHEN ls_readline.getIndexOf("#pages_input - Start -",1) > 0
            CALL adzp150_make_pages(lchannel_read,lchannel_write,1) RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""         
            
         #page段落產生(input段落,單身的單身)
         WHEN ls_readline.getIndexOf("#pages_input_d - Start -",1) > 0
            CALL adzp150_make_pages(lchannel_read,lchannel_write,2) RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
         
         #page段落產生(display段落)
         WHEN ls_readline.getIndexOf("#pages_display - Start -",1) > 0
            CALL adzp150_make_pages(lchannel_read,lchannel_write,3) RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""         
            
         #page段落產生(display段落,單身的單身)
         WHEN ls_readline.getIndexOf("#pages_display_d - Start -",1) > 0
            CALL adzp150_make_pages(lchannel_read,lchannel_write,4) RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
         #page段落產生(單身的單頭)
         WHEN ls_readline.getIndexOf("#pages_m - Start -",1) > 0
            CALL adzp150_make_pages(lchannel_read,lchannel_write,5) RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
         #page段落產生(單身的單身)
         WHEN ls_readline.getIndexOf("#pages_d - Start -",1) > 0
            CALL adzp150_make_pages(lchannel_read,lchannel_write,6) RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
         #單頭key段落產生    
         WHEN ls_readline.getIndexOf("#master_keys - Start -",1) > 0
            CALL adzp150_make_keys(lchannel_read,lchannel_write,'m') RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
         #單身key段落產生    
         WHEN ls_readline.getIndexOf("#detail_keys - Start -",1) > 0
            CALL adzp150_make_keys(lchannel_read,lchannel_write,'d') RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
         #多table段落產生
         WHEN ls_readline.getIndexOf("#tables - Start -",1) > 0
            CALL adzp150_make_tables(lchannel_read,lchannel_write,0) RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
         #多table段落產生(第一層單身)
         WHEN ls_readline.getIndexOf("#tables_d - Start -",1) > 0
            CALL adzp150_make_tables(lchannel_read,lchannel_write,2) RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
         #多table段落產生(第二層單身)
         WHEN ls_readline.getIndexOf("#tables_d2 - Start -",1) > 0
            CALL adzp150_make_tables(lchannel_read,lchannel_write,3) RETURNING lchannel_read,lchannel_write
            LET ls_readline = ""
            
      END CASE
      
      #行代換/ 對 ${} 置換
      IF ls_readline.getIndexOf("${",1) AND 
         ( ls_readline.getIndexOf("${",1) < ls_readline.getIndexOf("}",1) ) THEN
         LET ls_text = adzp150_line_replace(ls_readline)
      ELSE
         LET ls_text = ls_readline
      END IF
      
      CALL lchannel_write.write(ls_text)

   END WHILE 

   CALL lchannel_write.write(ls_text)

   CALL lchannel_read.close()
   CALL lchannel_write.close()
   
   #對產生器寫出的檔案權限在UNIX下均全部打開
   IF os.Path.separator() = "/" THEN
      IF os.Path.chrwx(ls_code_filename,511) THEN
      END IF
   END IF
   
END FUNCTION

#+ 產生程式版本及說明
PRIVATE FUNCTION adzp150_prog_memo()
   DEFINE ls_text      STRING
   DEFINE ls_buildtime STRING
   DEFINE li_buildtime INTEGER
   DEFINE ls_value     STRING
   DEFINE ls_mod_time  DATE
   
   LET li_buildtime = adzp150_prog_buildtime()
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
   
   #標準PR最後修改時間(mdl_modify_time_pr)
   #LET ls_value = g_properties.getValue("general_title_modify_time")
   #CALL g_properties.addAttribute("mdl_modify_time",ls_value)
   
   #客製SD版次(general_adp_ver_cust)
   LET ls_value = g_properties.getValue("general_adp_ver_cust")
   CALL g_properties.addAttribute("general_adp_ver_cust",ls_value)
   
   #客製SD最後修改時間(mdl_modify_time_cust)
   LET ls_value = g_properties.getValue("general_modify_time_cust")
   CALL g_properties.addAttribute("mdl_modify_time_cust",ls_value)
   
   #客製PR版次(mdl_pr_ver_cust)
   #LET ls_value = g_properties.getValue("")
   #CALL g_properties.addAttribute("mdl_pr_ver_cust",ls_value)
   
   #客製PR最後修改時間(mdl_modify_time_pr_cust)
   #LET ls_value = g_properties.getValue("")
   #CALL g_properties.addAttribute("mdl_modify_time_pr_cust",ls_value)
   
   #樣板名稱
   CALL g_properties.addAttribute("mdl_id",g_properties.getValue("type_id"))
   
   LET ls_text = adzp150_make_slice("a00")
   
   RETURN ls_text
   
END FUNCTION

#+ 取出版本建置次數
PRIVATE FUNCTION adzp150_prog_buildtime()
   DEFINE lc_dzaz001   LIKE dzaz_t.dzaz001
   DEFINE li_dzaz002   LIKE dzaz_t.dzaz002
   DEFINE lc_dzaz003   LIKE dzaz_t.dzaz003
   DEFINE lc_dzaz004   DATETIME YEAR TO SECOND

   LET lc_dzaz001 = ARG_VAL(2)
   LET li_dzaz002 = 0
   
   SELECT MAX(dzaz002) INTO li_dzaz002 FROM dzaz_t WHERE dzaz001 = lc_dzaz001
   
   IF STATUS OR li_dzaz002 IS NULL THEN
      LET li_dzaz002 = 0
   END IF
   
   LET lc_dzaz003 = FGL_GETENV("LOGNAME")
   LET lc_dzaz004 = CURRENT

   INSERT INTO dzaz_t(dzaz001,dzaz002,dzaz003,dzaz004)
      VALUES(lc_dzaz001,li_dzaz002+1,lc_dzaz003,lc_dzaz004)

   RETURN li_dzaz002 USING "&&&&"
   
END FUNCTION



#+ 逐行代換
PUBLIC FUNCTION adzp150_line_replace(ls_read)
   DEFINE ls_read     STRING
   DEFINE ls_text     STRING
   DEFINE ls_tag      STRING
   DEFINE li_pos1     LIKE type_t.num10
   DEFINE li_pos2     LIKE type_t.num10
   DEFINE ls_temp     STRING                #暫存properties資料用

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
         LET ls_text = adzp150_line_replace(ls_text)
      END IF

   END IF

   RETURN ls_text

END FUNCTION
    
#+ 走訪XML
PRIVATE FUNCTION adzp150_xml_search(ln_n)
   DEFINE ln_n        om.DomNode
   DEFINE ls_tmp      STRING
   
   LET g_master_chk = FALSE
   LET g_detail_chk = FALSE
   
   #取得第一個子節點 returns the first child node
   LET ln_n = ln_n.getFirstChild()
   WHILE ln_n IS NOT NULL
      CASE ln_n.getTagName()                                 
         WHEN "section"   #取 structure->section
            CASE 
               WHEN ln_n.getAttribute("id") = "global_var"
                  CALL adzp150_xml_global_var_search(ln_n) 

               #產生action的程式段落
               WHEN ln_n.getAttribute("id") = "menu"
                  CALL adzp150_write_on_action(3,g_properties.getValue("type_id"),ln_n,"menu",'')
                  CALL adzp150_create_action_default()
                  
               WHEN ln_n.getAttribute("id") = "detail_show"
                  CALL adzp150_write_on_action(3,g_properties.getValue("type_id"),ln_n,"detail_show",ln_n.getAttribute("page"))
                  
               WHEN ln_n.getAttribute("id") = "master_input"
                  CALL adzp150_write_on_action(3,g_properties.getValue("type_id"),ln_n,"master_input",'')

               WHEN ln_n.getAttribute("id") = "detail_input"
                  CALL adzp150_write_on_action(3,g_properties.getValue("type_id"),ln_n,"detail_input",ln_n.getAttribute("page"))

               WHEN ln_n.getAttribute("id") = "construct"
                  CALL adzp150_xml_construct_search(ln_n)  

               WHEN ln_n.getAttribute("id") = "input" 
                  CALL adzp150_xml_input_search(ln_n)  

               #產生串查action的程式段落
               #WHEN ln_n.getAttribute("id") = "chain"
               
            END CASE 
            CALL adzp150_xml_search(ln_n)

         WHEN "form"     #取 form 
            IF ln_n.getAttribute("site") = "Y" THEN
               LET ls_tmp = "DISPLAY g_site TO formonly.site"
               CALL g_properties.addAttribute("general_display_site",ls_tmp)
            END IF
            CALL adzp150_xml_search(ln_n)

         WHEN "dataset"  #取 dataset
            CALL adzp150_xml_dataset_search(ln_n)

         WHEN "init"    #取 form->init
            CASE
               WHEN ln_n.getAttribute("id") = "field_set"
                  CALL  adzp150_xml_field_set_search(ln_n)

               #201109-00009 ---start---
               WHEN ln_n.getAttribute("id") = "fuzzyquery"
                  CALL adzp150_xml_fuzzyquery_search(ln_n)
               #201109-00009 --- end ---

               #240930-00022 ---start---
               WHEN ln_n.getAttribute("id") = "value_formula"
                  CALL adzp150_xml_value_formula_search(ln_n,"1")   #處理目標控件
                  CALL adzp150_xml_value_formula_search(ln_n,"2")   #處理使用控件
               #240930-00022 --- end ---
            END CASE
            
         OTHERWISE 
            CALL adzp150_xml_search(ln_n)
            
      END CASE  
      LET ln_n = ln_n.getNext()  #同一層 next sibling 
      
   END WHILE 
   
END FUNCTION


#+ 取 structure->section id="global_var" 段落
PRIVATE FUNCTION adzp150_xml_global_var_search(ln_n)
   DEFINE ln_n            om.DomNode
   DEFINE li_page         LIKE type_t.num5
   DEFINE ls_sr_name      STRING 
   DEFINE ls_err          STRING
   DEFINE ls_tmp          STRING
   DEFINE ls_tmp2         STRING
   DEFINE l_tmp           STRING
   DEFINE ls_name         STRING
   DEFINE lst_token       base.StringTokenizer
   DEFINE ls_token        STRING
   DEFINE ls_bsvars       STRING
   DEFINE ls_page         STRING
   DEFINE ls_id           STRING
   
   LET ln_n = ln_n.getFirstChild()
   WHILE ln_n IS NOT NULL 

      LET ls_tmp = ln_n.getAttribute("value"),',', g_properties.getValue("general_form_field")
      
      CALL g_properties.addAttribute("general_form_field", ls_tmp )
   
      CASE ln_n.getAttribute("id") 
         WHEN "bs_field"    #取global_var中的 bs_field,BROWSER欄位定義 
            LET ls_tmp = ln_n.getAttribute("order")
            IF cl_null(ls_tmp) THEN
               LET ls_tmp = ln_n.getAttribute("pk") 
            END IF
            CALL g_properties.addAttribute("bs_order", ls_tmp )
            
            LET ls_tmp = ln_n.getAttribute("value")
            
            IF cl_null(g_properties.getValue("browser_fields_all_o")) THEN
               CALL g_properties.addAttribute("browser_fields_all_o", ln_n.getAttribute("value") )
            ELSE
               CALL g_properties.addAttribute("browser_fields_all_o", g_properties.getValue("browser_fields_all_o")||","||ln_n.getAttribute("value") )
            END IF
            
            CALL adzp150_create_borwser_filter(ls_tmp)
            
            LET ls_page = ln_n.getAttribute("page")
            LET ls_sr_name = ln_n.getAttribute("record")
            LET ls_name = adzp150_create_name(ls_page, "browser_sr_name", "<<<") 
            CALL g_properties.addAttribute(ls_name, ls_sr_name )
            
            #因應Q類
            LET ls_name = adzp150_create_name(ls_page, "browser_all_field", "<<<") 
            LET ls_tmp = ln_n.getAttribute("value")
            CALL g_properties.addAttribute(ls_name, ls_tmp )
            
            #去除特別欄位參照內容
            LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
            LET ls_tmp = g_properties.getValue("browser_fields_all")
            IF NOT cl_null(ls_tmp) THEN
               LET ls_tmp = ls_tmp, ","
            END IF
            WHILE lst_token.hasMoreTokens()
               LET ls_token = lst_token.nextToken()
               IF ls_token.getIndexOf('(',1) > 0 OR
                  ls_token.getIndexOf(g_properties.getValue("master_tbl_prefix"),1) = 0 THEN 
                  LET ls_token = ls_token.subString(1,ls_token.getIndexOf('(',1)-1)
                  IF cl_getField(g_properties.getValue("master_tbl_name"),ls_token) THEN
                     LET ls_tmp = ls_tmp, ls_token, ","
                  ELSE
                     LET ls_tmp = ls_tmp, "'',"
                  END IF
                  LET ls_tmp2 = ls_tmp2, ls_token, ","
               ELSE
                  LET ls_tmp = ls_tmp, ls_token, ','
                  LET ls_tmp2 = ls_tmp2, ls_token, ','
               END IF
            END WHILE
            LET ls_tmp = ls_tmp.substring(1,ls_tmp.getLength()-1)
            LET ls_tmp2 = ls_tmp2.substring(1,ls_tmp2.getLength()-1)
            CALL g_properties.addAttribute("browser_fields_insert_tmp", ls_tmp2 )
            CALL g_properties.addAttribute("browser_fields_all_tree", ls_tmp )
            LET ls_tmp = adzp150_add_table_name(ls_tmp,'b')
            CALL g_properties.addAttribute("browser_fields_all", ls_tmp )
    
            LET ls_tmp = ln_n.getAttribute("value")
            LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
            LET ls_bsvars = ""
            WHILE lst_token.hasMoreTokens()
               LET ls_token = lst_token.nextToken()
               IF ls_token.getIndexOf('(',1) = 0 THEN
                  LET ls_bsvars = ls_bsvars,"g_browser[g_cnt].b_",ls_token,','
               END IF      
            END WHILE            
            LET ls_bsvars = ls_bsvars.subString(1,ls_bsvars.getLength()-1)
            CALL g_properties.addAttribute("browser_vars_all", ls_bsvars )
        
            #檢查如果樣板含 Browser 需求，則務必要此項"type_id"
            LET ls_err = g_properties.getValue("type_id")
            
            #若樣板不含browser則不進行以下處理
            IF ls_err.equals("i01") OR ls_err.equals("t01") OR ls_err.equals("t02") OR 
               ls_err.equals("i04") OR ls_err.equals("i05") OR ls_err.equals("i07") OR 
               ls_err.equals("i08") THEN
               #定義 Browser的變數
               IF NOT adzp150_browser_fields_define(g_properties.getValue("browser_fields_all_o"),ln_n.getAttribute("append"),ln_n.getAttribute("pk"),ln_n.getAttribute("ref_table"),ln_n.getAttribute("ref_wc"),ln_n.getAttribute("ref_type")) THEN
                  #DISPLAY "ERROR(5):設定g_browse變數發生問題"
               END IF
            END IF

         #取單頭設定資料
         WHEN "head"
            #單頭 RECORD 設定,調整命名來源
            LET ls_tmp = g_properties.getValue("master_tbl_prefix")
            LET l_tmp = ls_tmp
            LET ls_tmp = "g_",ls_tmp,"_m"
            IF ls_tmp = 'g__m' THEN
               LET ls_tmp = 'g_master'
            END IF

            CALL g_properties.addAttribute("master_var_title",ls_tmp)                                       
            CALL g_properties.addAttribute("title_master",l_tmp)    
 
            #master_fields_define及var_h_fill的設定
            CALL adzp150_master_fields_define(li_page,ln_n.getAttribute("value"))
            
            #master_all_field
            CALL g_properties.addAttribute("master_all_field",ln_n.getAttribute("value")) 
            
            #general_all_field
            LET ls_tmp = g_properties.getValue("general_all_field")
            LET ls_tmp = ls_tmp, ln_n.getAttribute("value"), ','
            CALL g_properties.addAttribute("general_all_field",ls_tmp) 
            
         #取單身頁面設定資料
         WHEN "body"
            LET li_page = ln_n.getAttribute("page")
            
            LET ls_id = ln_n.getAttribute("page_id")
            IF cl_null(ls_id) THEN
               LET ls_id = ln_n.getAttribute("page")
            END IF
            LET ls_name = adzp150_create_name(li_page, "detail_page_id", "<<<") 
            CALL g_properties.addAttribute(ls_name,ls_id)

            #單身DYNAMIC ARRAY設定
            LET ls_tmp = ln_n.getAttribute("record")
            LET ls_name = adzp150_create_name(li_page, "general_page_input", "<<<") 
            IF ls_tmp.getIndexOf("_info",1) > 0 OR
               ln_n.getAttribute("input") = "N" THEN
               CALL g_properties.addAttribute(ls_name,"N")
            ELSE
               CALL g_properties.addAttribute(ls_name,"Y")
            END IF
            
            #紀錄各個page所含之欄位
            LET l_tmp = li_page USING "<<<"
            CALL g_properties.addAttribute("detail_field_loc"||l_tmp,ln_n.getAttribute("value"))
            
            IF li_page = 1 THEN 
            
               CALL g_properties.addAttribute("sr_name",ls_tmp)  #${sr_name1}

               #調整命名來源
               LET ls_tmp = g_properties.getValue("detail_tbl_prefix")
               LET l_tmp = ls_tmp
               LET ls_tmp = "g_",ls_tmp,"_d"

               CALL g_properties.addAttribute("detail_var_title",ls_tmp)
               CALL g_properties.addAttribute("title_detail",l_tmp)
               
               #190412-00022 ---start---
               LET ls_tmp = g_properties.getValue("detail_tbl_prefix")
               LET l_tmp = ls_tmp
               LET ls_tmp = "g_",ls_tmp,"_d[l_ac]"
               CALL g_properties.addAttribute("detail_var_title_idx",ls_tmp)
               #190412-00022 --- end ---

            ELSE
               LET ls_sr_name = "sr_name",li_page USING "<<<"
               CALL g_properties.addAttribute(ls_sr_name,ls_tmp)  #${sr_name2}

               LET ls_sr_name = "detail_var_title",li_page USING "<<<"
               LET ls_tmp = g_properties.getValue("detail_tbl_prefix")
               LET ls_tmp = "g_",ls_tmp,adzp150_page_trans(li_page), "_d"
               CALL g_properties.addAttribute(ls_sr_name,ls_tmp)

               #190412-00022 ---start---
               LET ls_sr_name = "detail_var_title_idx",li_page USING "<<<"
               LET ls_tmp = g_properties.getValue("detail_tbl_prefix")
               LET ls_tmp = "g_",ls_tmp,adzp150_page_trans(li_page), "_d[l_ac]"
               CALL g_properties.addAttribute(ls_sr_name,ls_tmp)
               #190412-00022 --- end ---
            END IF                                             

            #detail_fields_define及detail_vars_all的設定
            CALL adzp150_detail_fields_define(li_page,adzp150_field_filter(ln_n.getAttribute("value"),"field"))
            
            #detail_all_field
            LET ls_tmp = g_properties.getValue("detail_all_field")
            LET ls_tmp = ls_tmp, ln_n.getAttribute("value")
            CALL g_properties.addAttribute("detail_all_field",ls_tmp)
            
            #general_all_field
            LET ls_tmp = g_properties.getValue("general_all_field")
            LET ls_tmp = ls_tmp, ln_n.getAttribute("value"), ','
            CALL g_properties.addAttribute("general_all_field",ls_tmp) 
            
         #額外所需的global變數
         WHEN "define"
            LET ls_tmp = ""
            LET ls_tmp = g_properties.getValue("addictional_global")
            
            #IF NOT cl_null(ls_tmp) THEN
            #   CALL adzp150_remove_attribute("addictional_global")
            #END IF
            
            LET ls_tmp = ls_tmp,"DEFINE ",ln_n.getAttribute("variable"),10 space,
                                          ln_n.getAttribute("type"),"\n"
            CALL g_properties.addAttribute("addictional_global",ls_tmp)
         
         OTHERWISE 
            IF ln_n.getTagName() <> "bs_reference" THEN 
               DISPLAY "WARNING: global_var區段存在不允許的屬性:",ln_n.getAttribute("id") 
            END IF
      END CASE 
      LET ln_n = ln_n.getNext()  #同一層 next sibling   
      
   END WHILE
   
   #將全域變數的資料存進DB中(add-point)
   IF NOT cl_null(g_properties.getValue("addictional_global")) THEN
      LET g_dzbb.prog_name   = g_properties.getValue("app_id")
      LET g_dzbb.point_name  = "global.variable"
      LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
      LET g_dzbb.addpoint    = g_properties.getValue("addictional_global")
      CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))     
   END IF

END FUNCTION 


#+ CONSTRUCT段落讀取資料
PRIVATE FUNCTION adzp150_xml_construct_search(pn_node)
   DEFINE pn_node     om.DomNode
   DEFINE li_page     LIKE type_t.num5
   DEFINE ls_str_page STRING 
   DEFINE ls_tmp      STRING 
   
   LET pn_node = pn_node.getFirstChild()
   
   #預設沒有frozen欄位
   CALL g_properties.addAttribute("general_mark_frozen","#")
   
   WHILE pn_node IS NOT NULL
      
      CASE 
         WHEN pn_node.getAttribute("id")  = "head" #${master_fields_qbe}
            LET ls_tmp = g_properties.getValue("type_id")
            IF ls_tmp.subString(1,1) = 'q' THEN
               CALL g_properties.addAttribute("master_fields_qbe",adzp150_fields_qbe_filter(g_properties.getValue("master_all_field")))
            ELSE
               CALL g_properties.addAttribute("master_fields_qbe",pn_node.getAttribute("field"))
            END IF
            CALL adzp150_browser_filter_define(pn_node.getAttribute("page"),pn_node.getAttribute("field"))
            
         WHEN pn_node.getAttribute("id")  = "body"
            #c類特別處理
            LET li_page = pn_node.getAttribute("page")
            IF g_type.subString(4,4) = 'c' THEN
               CALL adzp150_cs_qbe(li_page,pn_node.getAttribute("field"))
            END IF 
            #IF l_page = 1 THEN
            #   CALL g_properties.addAttribute("cs_body_fields",n.getAttribute("field"))
            #   CALL g_properties.addAttribute("cs_sr_name",adzp150_change_cs_sr_detail("s_detail1",n.getAttribute("field")))
            #ELSE 
               LET ls_str_page = li_page
               IF li_page = 1 THEN
                  LET ls_tmp = g_properties.getValue("sr_name")
               ELSE
                  LET ls_tmp = g_properties.getValue("sr_name"||ls_str_page)
               END IF
               CALL g_properties.addAttribute("detail_fields_qbe"||ls_str_page,pn_node.getAttribute("field")) 
               CALL g_properties.addAttribute("detail_srfield_all"||ls_str_page,adzp150_change_cs_sr_detail(ls_tmp,pn_node.getAttribute("field")))
            #END IF 
            #CALL adzp150_xml_construct_search(n) 
            
         WHEN pn_node.getAttribute("id")  = "body2"
            CALL g_properties.addAttribute("detail_frozen_fields_qbe",pn_node.getAttribute("field")) 
            CALL g_properties.addAttribute("general_mark_frozen","")
            
      END CASE 
      LET pn_node = pn_node.getNext()  #同一層 next sibling 
   END WHILE 
   
END FUNCTION 


#+ 取出欄位的總表
PRIVATE FUNCTION adzp150_xml_input_search(pn_node)
   DEFINE pn_node     om.DomNode
   DEFINE ls_value    STRING
   DEFINE ls_page     STRING
   DEFINE ls_name     STRING
   DEFINE ls_tmp      STRING
   DEFINE li_idx      INTEGER
 
   LET pn_node = pn_node.getFirstChild()
   
   WHILE pn_node IS NOT NULL
   
      CASE 
         WHEN pn_node.getAttribute("id") = "head"
            LET ls_name = "master_input_fields_list"
            LET ls_value = pn_node.getAttribute("field")

            CALL g_properties.addAttribute(ls_name,ls_value)
            #171204-00032#1 - Start -
            #170906-00011#1 - Start -
            #取得第一個可輸入的欄位名稱
            #IF ls_value.getIndexOf(',',1) > 1 THEN
            #   LET ls_tmp = ls_value.subString(1,ls_value.getIndexOf(',',1)-1)
            #ELSE
            #   LET ls_tmp = ls_value
            #END IF
            #LET ls_name = "master_first_input_field"
            #LET ls_value = ls_tmp
            #CALL g_properties.addAttribute(ls_name,ls_value)
            #170906-00011#1 -  End  -
            #171204-00032#1 -  End  -
            
        WHEN pn_node.getAttribute("id") = "body"
            LET ls_value = pn_node.getAttribute("field")
            LET ls_page = pn_node.getAttribute("page")
            LET ls_name = adzp150_create_name(ls_page, "detail_input_fields_list", "<<<") 
            CALL g_properties.addAttribute(ls_name,ls_value)
            
            #該page能否insert/append/delete進行調整
            LET ls_tmp = pn_node.getAttribute("insert")
            IF cl_null(ls_tmp) THEN
               LET ls_tmp = "Y"
            END IF
            IF ls_tmp = "Y" THEN
               LET ls_value = "l_allow_insert"
            ELSE
               LET ls_value = "FALSE" 
            END IF
            LET ls_name = adzp150_create_name(ls_page, "detail_allow_insert", "<<<") 
            CALL g_properties.addAttribute(ls_name,ls_value)
            
            LET ls_tmp = pn_node.getAttribute("append")
            IF cl_null(ls_tmp) THEN
               LET ls_tmp = "Y"
            END IF
            IF ls_tmp = "Y" THEN
               LET ls_value = "l_allow_insert"
            ELSE
               LET ls_value = "FALSE" 
            END IF
            LET ls_name = adzp150_create_name(ls_page, "detail_allow_append", "<<<") 
            CALL g_properties.addAttribute(ls_name,ls_value)
            
            LET ls_tmp = pn_node.getAttribute("delete")
            IF cl_null(ls_tmp) THEN
               LET ls_tmp = "Y"
            END IF
            IF ls_tmp = "Y" THEN
               LET ls_value = "l_allow_delete"
            ELSE
               LET ls_value = "FALSE" 
            END IF
            LET ls_name = adzp150_create_name(ls_page, "detail_allow_delete", "<<<") 
            CALL g_properties.addAttribute(ls_name,ls_value)
           
            #170906-00011#1 - Start -
            #取得第一個可輸入的欄位名稱
            LET ls_value = pn_node.getAttribute("field")
            IF ls_value.getIndexOf(',',1) > 1 THEN
               LET ls_tmp = ls_value.subString(1,ls_value.getIndexOf(',',1)-1)
            ELSE
               LET ls_tmp = ls_value
            END IF
            LET ls_name = adzp150_create_name(ls_page, "detail_first_input_field", "<<<")
            LET ls_value = ls_tmp
            CALL g_properties.addAttribute(ls_name,ls_value)
            #170906-00011#1 -  End  -
  
      END CASE 
      
      LET pn_node = pn_node.getNext()  #同一層 next sibling 
      
   END WHILE 
   
   #額外判斷哪些page可做輸入
   
   FOR li_idx = 1 TO g_properties.getValue("page")
      #單身DYNAMIC ARRAY設定
      LET ls_name = adzp150_create_name(li_idx, "detail_input_fields_list", "<<<")
      LET ls_value = ""
      LET ls_value = g_properties.getValue(ls_name)
      LET ls_name = adzp150_create_name(li_idx, "general_page_input", "<<<")
      IF NOT cl_null(ls_value) THEN
         CALL g_properties.addAttribute(ls_name,"Y")        
      ELSE    
         CALL g_properties.addAttribute(ls_name,"N")        
      END IF
   END FOR
   
END FUNCTION 


#+ 轉換成含record name list
PRIVATE FUNCTION adzp150_trans_input_fields()
   DEFINE l_tok      base.StringTokenizer
   DEFINE ls_nxt     STRING
   DEFINE ls_tmp     STRING
   DEFINE ls_name    STRING
   DEFINE ls_value   STRING
   DEFINE ls_field   STRING
   DEFINE ls_title   STRING
   DEFINE ls_common  DYNAMIC ARRAY OF STRING
   DEFINE li_idx     INTEGER
   
   #定義共用欄位名稱(單頭)
   LET ls_common[1] = g_properties.getValue("master_tbl_prefix"), "ownid"
   LET ls_common[2] = g_properties.getValue("master_tbl_prefix"), "owndp"
   LET ls_common[3] = g_properties.getValue("master_tbl_prefix"), "crtid"
   LET ls_common[4] = g_properties.getValue("master_tbl_prefix"), "crtdt"
   LET ls_common[5] = g_properties.getValue("master_tbl_prefix"), "crtdp"
   LET ls_common[6] = g_properties.getValue("master_tbl_prefix"), "modid"
   LET ls_common[7] = g_properties.getValue("master_tbl_prefix"), "moddt"
   LET ls_common[8] = g_properties.getValue("master_tbl_prefix"), "cnfid"
   LET ls_common[9] = g_properties.getValue("master_tbl_prefix"), "cnfdt"

   #單頭input
   LET l_tok = base.StringTokenizer.create(g_properties.getValue("master_input_fields_list"),",")
   WHILE l_tok.hasMoreTokens()
      LET ls_field = l_tok.nextToken()
      #濾除共用欄位
      IF ls_common[1] = ls_field OR
         ls_common[2] = ls_field OR
         ls_common[3] = ls_field OR
         ls_common[4] = ls_field OR
         ls_common[5] = ls_field OR
         ls_common[6] = ls_field OR
         ls_common[7] = ls_field OR
         ls_common[8] = ls_field OR
         ls_common[9] = ls_field THEN
         CONTINUE WHILE
      END IF
      LET ls_tmp = ls_tmp,g_properties.getValue("master_var_title"),".",ls_field,","
   END WHILE
   
   LET ls_tmp = ls_tmp.subString(1,ls_tmp.getLength()-1)
   CALL g_properties.addAttribute("master_input_fields",ls_tmp)
   
   FOR li_idx = 1 TO g_properties.getValue("page")
      
      LET ls_tmp = ""
      
      #title
      LET ls_name = adzp150_create_name(li_idx, "sr_name", "<<<") 
      LET ls_title = g_properties.getValue(ls_name)

      #定義共用欄位名稱(單身)
      LET ls_common[1] = g_properties.getValue("detail_tbl_prefix"), "ownid"
      LET ls_common[2] = g_properties.getValue("detail_tbl_prefix"), "owndp"
      LET ls_common[3] = g_properties.getValue("detail_tbl_prefix"), "crtid"
      LET ls_common[4] = g_properties.getValue("detail_tbl_prefix"), "crtdt"
      LET ls_common[5] = g_properties.getValue("detail_tbl_prefix"), "crtdp"
      LET ls_common[6] = g_properties.getValue("detail_tbl_prefix"), "modid"
      LET ls_common[7] = g_properties.getValue("detail_tbl_prefix"), "moddt"
      LET ls_common[8] = g_properties.getValue("detail_tbl_prefix"), "cnfid"
      LET ls_common[9] = g_properties.getValue("detail_tbl_prefix"), "cnfdt"
      
      #單身input
      LET ls_name = adzp150_create_name(li_idx, "detail_input_fields_list", "<<<")
      LET ls_value = g_properties.getValue(ls_name)
      LET l_tok = base.StringTokenizer.create(ls_value,",")
      
      WHILE l_tok.hasMoreTokens()
         LET ls_field = l_tok.nextToken()
         #濾除共用欄位
         IF ls_common[1] = ls_field OR
            ls_common[2] = ls_field OR
            ls_common[3] = ls_field OR
            ls_common[4] = ls_field OR
            ls_common[5] = ls_field OR
            ls_common[6] = ls_field OR
            ls_common[7] = ls_field OR
            ls_common[8] = ls_field OR
            ls_common[9] = ls_field THEN
            CONTINUE WHILE
         END IF

         LET ls_tmp = ls_tmp,ls_title,".",ls_field
         IF l_tok.hasMoreTokens() THEN
            LET ls_tmp = ls_tmp,","
         END IF
      END WHILE
      
      LET ls_name = adzp150_create_name(li_idx, "general_page_input", "<<<") 
      IF cl_null(ls_tmp) THEN
         #CALL g_properties.addAttribute(ls_name,"N")
      ELSE
         #CALL g_properties.addAttribute(ls_name,"Y")
         LET ls_tmp = ls_tmp.subString(1,ls_tmp.getLength())
      END IF

      LET ls_name = adzp150_create_name(li_idx, "detail_input_fields", "<<<") 
      CALL g_properties.addAttribute(ls_name,ls_tmp)
   END FOR
   
END FUNCTION

#201109-00009 ---start---
#列出所有可做模糊查詢的欄位list
PRIVATE FUNCTION adzp150_xml_fuzzyquery_search(pn_node)
   DEFINE pn_node         om.DomNode
   DEFINE li_page         LIKE type_t.num5
   DEFINE ls_str_page     STRING
   DEFINE ls_tmp          STRING
   DEFINE ls_name         STRING
   DEFINE ls_value        STRING


   LET pn_node = pn_node.getFirstChild()

   WHILE pn_node IS NOT NULL

      CASE
         WHEN pn_node.getAttribute("id") = "head"
            CALL g_properties.addAttribute("master_fuzzyquery_fields",pn_node.getAttribute("field"))

         WHEN pn_node.getAttribute("id") = "body"
            LET li_page = pn_node.getAttribute("page")
            LET ls_str_page = li_page
            IF li_page = 1 THEN
               CALL g_properties.addAttribute("detail_fuzzyquery_fields",pn_node.getAttribute("field"))
            ELSE
               LET ls_value = pn_node.getAttribute("field")
               LET ls_name = adzp150_create_name(li_page, "detail_fuzzyquery_fields", "<<<")
               CALL g_properties.addAttribute(ls_name,ls_value)
            END IF
      END CASE
      LET pn_node = pn_node.getNext()  #同一層 next sibling

   END WHILE

END FUNCTION
#201109-00009 --- end ---

#240930-00022 ---start---
#列出所有有設定簡易帶值或公式的欄位list
PRIVATE FUNCTION adzp150_xml_value_formula_search(pn_node,ls_type)
   DEFINE pn_node         om.DomNode
   DEFINE ls_type         LIKE type_t.chr20   #1:目標控件  2:使用控件
   DEFINE li_page         LIKE type_t.num5
   DEFINE ls_str_page     STRING
   DEFINE ls_tmp          STRING
   DEFINE ls_name         STRING
   DEFINE ls_value        STRING
   DEFINE ls_sql          STRING
   DEFINE l_token         base.StringTokenizer
   DEFINE ls_token        STRING
   DEFINE ls_target_field LIKE type_t.chr100   #目標控件編號 ex:xmdc_t.xmdcud001
   DEFINE ls_ctrol_field  LIKE type_t.chr100   #控件編號 ex:xmdc_t.xmdc011
   DEFINE ls_org_field    LIKE type_t.chr100   #欄位代號 ex:xmdc011
   DEFINE ls_var_field    LIKE type_t.chr100   #程式中用的變數代號 ex:g_xmdc_d[l_ac].xmdc011
   DEFINE ls_i            LIKE type_t.num5
   DEFINE ls_m            LIKE type_t.num5
   DEFINE li_page_old     LIKE type_t.num5
   DEFINE ls_prog_id      LIKE type_t.chr100
   DEFINE ls_tagname      LIKE type_t.chr100
   DEFINE ls_for_used     LIKE type_t.chr20
   DEFINE ls_old_value    STRING

   LET pn_node = pn_node.getFirstChild()

   IF ls_type = "1" THEN
      CALL g_target_col.clear()
   ELSE
      CALL g_value_formula.clear()
   END IF

   LET ls_i = 0
   LET ls_prog_id = g_properties.getValue("app_id")

   WHILE pn_node IS NOT NULL

      LET ls_tagname = pn_node.getTagName()

      #取得目標控件
      IF ls_tagname = "value_formula_v" THEN   #簡易帶值
         LET ls_for_used = "value"
         LET ls_sql = "SELECT dzcp002",
                       " FROM dzcp_t,dzcq_t",
                      " WHERE dzcp001 = dzcq001",
                        " AND dzcp002 = dzcq002",
                        " AND dzcpstus = 'Y'",
                        " AND dzcp001 = '",ls_prog_id,"'",
                        " AND dzcq003 = ?"
      ELSE
         LET ls_for_used = "formula"
         LET ls_sql = "SELECT dzct002",
                       " FROM dzct_t,dzcw_t",
                      " WHERE dzct001 = dzcw001",
                        " AND dzct002 = dzcw002",
                        " AND dzctstus = 'Y'",
                        " AND dzct001 = '",ls_prog_id,"'",
                        " AND dzcw003 = ?"
      END IF

      PREPARE adzp150_xml_value_formula_get_target_col_pre FROM ls_sql
      DECLARE adzp150_xml_value_formula_get_target_col_cur CURSOR FOR adzp150_xml_value_formula_get_target_col_pre

      CASE
         WHEN pn_node.getAttribute("id") = "head"
            #因為簡單帶值與公式在tab檔中是分開呈現的，
            #但在程式段中是要合併一起產出程式段，
            #因為不確定前面有沒有處理過了，所以每次要寫入時，先取一遍，
            #如果有值的話，就繼續承接下去，而不是直接蓋掉
            LET ls_value = NULL
            LET ls_old_value = g_properties.getValue("master_value_formula_fields")
            IF NOT cl_null(ls_old_value) THEN
               LET ls_value = ls_old_value
            END IF

            IF ls_type = "1" THEN
               #目標控件
               LET ls_tmp = pn_node.getAttribute("target_field")
               LET l_token = base.StringTokenizer.create(ls_tmp,",")
               WHILE l_token.hasMoreTokens()
                  LET ls_token = l_token.nextToken()
                  LET ls_ctrol_field = ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
                  LET ls_org_field = ls_token.subString(ls_token.getIndexOf("(",1)+1,ls_token.getLength()-1)
                  LET ls_var_field = g_properties.getValue("master_var_title"),".",ls_org_field

                  LET ls_m = g_target_col.getLength() + 1
                  LET g_target_col[ls_m].ctrol_col = ls_ctrol_field
                  LET g_target_col[ls_m].var_col = ls_var_field
               END WHILE
            END IF

            IF ls_type = "2" THEN
               #使用控件
               LET ls_tmp = pn_node.getAttribute("field")
               LET l_token = base.StringTokenizer.create(ls_tmp,",")

               #ex:xmdc_t.xmdc011(xmdc011),xmdc_t.xmdc015(xmdc015)
               WHILE l_token.hasMoreTokens()
                  LET ls_token = l_token.nextToken()
                  LET ls_ctrol_field = ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
                  LET ls_org_field = ls_token.subString(ls_token.getIndexOf("(",1)+1,ls_token.getLength()-1)
                  LET ls_var_field = g_properties.getValue("master_var_title"),".",ls_org_field

                  #將簡易帶值及公式的相關資訊寫入陣列中
                  FOREACH adzp150_xml_value_formula_get_target_col_cur USING ls_ctrol_field INTO ls_target_field
                     FOR ls_m = 1 TO g_target_col.getLength()
                        IF g_target_col[ls_m].ctrol_col = ls_target_field THEN
                           LET ls_i = g_value_formula.getLength() + 1
                           LET g_value_formula[ls_i].type = ls_for_used
                           LET g_value_formula[ls_i].location = "head"
                           LET g_value_formula[ls_i].page_num = 0
                           LET g_value_formula[ls_i].target_col = ls_target_field
                           LET g_value_formula[ls_i].target_var_col = g_target_col[ls_m].var_col
                           LET g_value_formula[ls_i].ctrol_col = ls_ctrol_field
                           LET g_value_formula[ls_i].org_col = ls_org_field
                           LET g_value_formula[ls_i].var_col = ls_var_field
                           EXIT FOR
                        END IF
                     END FOR
                  END FOREACH

                  IF ls_value IS NULL THEN
                     LET ls_value = ls_org_field
                  ELSE
                     LET ls_value = ls_value,",",ls_org_field
                  END IF
               END WHILE
               CALL g_properties.addAttribute("master_value_formula_fields",ls_value)
            END IF


         WHEN pn_node.getAttribute("id") = "body"
            LET li_page = pn_node.getAttribute("page")
            #因為簡單帶值與公式在tab檔中是分開呈現的，
            #但在程式段中是要合併一起產出程式段，
            #因為不確定前面有沒有處理過了，所以每次要寫入時，先取一遍，
            #如果有值的話，就繼續承接下去，而不是直接蓋掉
            LET ls_value = NULL
            IF li_page = 1 THEN
               LET ls_name = "detail_value_formula_fields"
            ELSE
               LET ls_name = adzp150_create_name(li_page, "detail_value_formula_fields", "<<<")
            END IF
            LET ls_old_value = g_properties.getValue(ls_name)
            IF NOT cl_null(ls_old_value) THEN
               LET ls_value = ls_old_value
            END IF


            IF ls_type = "1" THEN
               #目標控件
               LET ls_tmp = pn_node.getAttribute("target_field")
               LET l_token = base.StringTokenizer.create(ls_tmp,",")
               WHILE l_token.hasMoreTokens()
                  LET ls_token = l_token.nextToken()
                  LET ls_ctrol_field = ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
                  LET ls_org_field = ls_token.subString(ls_token.getIndexOf("(",1)+1,ls_token.getLength()-1)

                  IF li_page = 1 THEN
                     LET ls_var_field = g_properties.getValue("detail_var_title"),"[l_ac].",ls_org_field
                  ELSE
                     LET ls_var_field = adzp150_create_name(li_page, "detail_var_title", "<<<"),"[l_ac.]",ls_org_field
                  END IF

                  LET ls_m = g_target_col.getLength() + 1
                  LET g_target_col[ls_m].ctrol_col = ls_ctrol_field
                  LET g_target_col[ls_m].var_col = ls_var_field
               END WHILE
            END IF


            IF ls_type = "2" THEN
               #使用控件
               LET ls_tmp = pn_node.getAttribute("field")
               LET l_token = base.StringTokenizer.create(ls_tmp,",")

               #ex:xmdc_t.xmdc011(xmdc011),xmdc_t.xmdc015(xmdc015)
               WHILE l_token.hasMoreTokens()
                  LET ls_token = l_token.nextToken()
                  LET ls_ctrol_field = ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
                  LET ls_org_field = ls_token.subString(ls_token.getIndexOf("(",1)+1,ls_token.getLength()-1)

                  IF li_page = 1 THEN
                     LET ls_name = "detail_value_formula_fields"
                     LET ls_var_field = g_properties.getValue("detail_var_title"),"[l_ac].",ls_org_field
                  ELSE
                     LET ls_name = adzp150_create_name(li_page, "detail_value_formula_fields", "<<<")
                     LET ls_var_field = adzp150_create_name(li_page, "detail_var_title", "<<<"),"[l_ac.]",ls_org_field
                  END IF

                  #將簡易帶值及公式的相關資訊寫入陣列中
                  FOREACH adzp150_xml_value_formula_get_target_col_cur USING ls_ctrol_field INTO ls_target_field
                     FOR ls_m = 1 TO g_target_col.getLength()
                        IF g_target_col[ls_m].ctrol_col = ls_target_field THEN
                           LET ls_i = g_value_formula.getLength() + 1
                           LET g_value_formula[ls_i].type = ls_for_used
                           LET g_value_formula[ls_i].location = "body"
                           LET g_value_formula[ls_i].page_num = li_page
                           LET g_value_formula[ls_i].target_col = ls_target_field
                           LET g_value_formula[ls_i].target_var_col = g_target_col[ls_m].var_col
                           LET g_value_formula[ls_i].ctrol_col = ls_ctrol_field
                           LET g_value_formula[ls_i].org_col = ls_org_field
                           LET g_value_formula[ls_i].var_col = ls_var_field
                           EXIT FOR
                        END IF
                     END FOR
                  END FOREACH

                  IF ls_value IS NULL THEN
                     LET ls_value = ls_org_field
                  ELSE
                     LET ls_value = ls_value,",",ls_org_field
                  END IF
               END WHILE
               CALL g_properties.addAttribute(ls_name,ls_value)
            END IF
      END CASE
      LET pn_node = pn_node.getNext()  #同一層 next sibling

   END WHILE

END FUNCTION
#240930-00022 --- end ---

PRIVATE FUNCTION adzp150_xml_field_set_search(pn_node)
   DEFINE pn_node         om.DomNode
   DEFINE ls_tbl_pre      STRING
   DEFINE ls_field        STRING
   DEFINE ls_field_list   STRING
   DEFINE ls_master_list  STRING
   
   LET pn_node = pn_node.getFirstChild()
   
   WHILE pn_node IS NOT NULL
      
      CASE 
         WHEN pn_node.getAttribute("id") = "head" AND pn_node.getAttribute("type") = "entry"  #${entry}
            CALL g_properties.addAttribute("master_entry_fields",pn_node.getAttribute("field"))
         WHEN pn_node.getAttribute("id") = "head" AND pn_node.getAttribute("type") = "no_entry" #${no_entry}
            CALL g_properties.addAttribute("master_no_entry_fields",pn_node.getAttribute("field"))   
         WHEN pn_node.getAttribute("id") = "body" AND pn_node.getAttribute("type") = "entry"  #${entry}
            CALL g_properties.addAttribute("detail_entry_fields",pn_node.getAttribute("field"))
         WHEN pn_node.getAttribute("id") = "body" AND pn_node.getAttribute("type") = "no_entry" #${no_entry}
            CALL g_properties.addAttribute("detail_no_entry_fields",pn_node.getAttribute("field"))   
      END CASE 
      LET pn_node = pn_node.getNext()  #同一層 next sibling 
      
   END WHILE 
   
   #先取出單頭所有欄位
   LET ls_master_list = g_properties.getValue("master_all_field"),','
   
   #docno, docdt, ld特別處理(目前僅針對單頭)
   LET ls_tbl_pre = g_properties.getValue("master_tbl_name")
   LET ls_tbl_pre = ls_tbl_pre.subString(1,ls_tbl_pre.getIndexOf('_',1)-1)
   
   #docno
   LET ls_field = ls_tbl_pre,'docno,'
   IF ls_master_list.getIndexOf(ls_field,1) THEN
      LET ls_field_list = ls_field_list, ls_field
   END IF
   
   #docdt
   LET ls_field = ls_tbl_pre,'docdt' #<-不要加逗號,是自己判斷的
   IF ls_master_list.getIndexOf(ls_field,1) THEN
      CALL g_properties.addAttribute("master_no_entry_docdt",ls_field)   
     #LET ls_field_list = ls_field_list, ls_field
   END IF
   
   #ld
   LET ls_field = ls_tbl_pre,'ld,'
   IF ls_master_list.getIndexOf(ls_field,1) THEN
      LET ls_field_list = ls_field_list, ls_field
   END IF
   
   IF NOT cl_null(ls_field_list) THEN
      LET ls_field_list = ls_field_list.subString(1,ls_field_list.getLength()-1)
      CALL g_properties.addAttribute("master_no_entry_anyway",ls_field_list)   
   END IF

END FUNCTION 


PRIVATE FUNCTION adzp150_xml_dataset_search(n)
   DEFINE n                   om.DomNode
   DEFINE n2                  om.DomNode
   DEFINE l_tag_attribute     STRING
   DEFINE ls_fix              STRING
   DEFINE li_idx              INTEGER
   DEFINE ls_idx              STRING
   DEFINE ls_tmp              STRING
   DEFINE ls_tmp2             STRING
   DEFINE ls_name             STRING
   DEFINE ls_value            STRING
   DEFINE ls_field_allkeys    STRING
   DEFINE ls_var_allkeys      STRING
   DEFINE lst_token           base.StringTokenizer
   DEFINE ls_token            STRING
   DEFINE lc_str              STRING   #170203-00031
   
   #先行取得第三階單身資料
   LET n2 = n.getFirstChild()
   
   LET n = n.getFirstChild()  #取得第一個子節點 returns the first child node
   WHILE n IS NOT NULL
      CASE n.getTagName()
         WHEN "head"
            CALL g_properties.addAttribute("master_tbl_cnt",1)
            
            IF g_master_chk THEN
               CALL adzp150_master_multitable_info(n)
               LET n = n.getNext()
               CONTINUE WHILE
            END IF
            
            CALL g_properties.addAttribute("master_tbl_name_by_tbl",n.getAttribute("id") )
            
            LET g_master_chk = TRUE
         
            LET ls_field_allkeys = ""
            LET ls_var_allkeys = ""
            
            #TREE lid,pid,type定義判斷
            LET ls_tmp = n.getAttribute("lid")
            LET ls_tmp2 = g_properties.getValue("master_field_allkeys")
            IF ls_tmp2.getIndexOf(ls_tmp,1) THEN
               CALL g_properties.addAttribute("master_lid_define_mark","#")
            END IF
            LET ls_tmp = n.getAttribute("pid")
            LET ls_tmp2 = g_properties.getValue("master_field_allkeys")
            IF ls_tmp2.getIndexOf(ls_tmp,1) THEN
               CALL g_properties.addAttribute("master_pid_define_mark","#")
            END IF
            LET ls_tmp = n.getAttribute("type")
            LET ls_tmp2 = g_properties.getValue("master_field_allkeys")
            IF ls_tmp2.getIndexOf(ls_tmp,1) THEN
               CALL g_properties.addAttribute("master_type_define_mark","#")
            END IF
            
            #單頭tree用local id
            LET ls_tmp = n.getAttribute("lid")
            CALL g_properties.addAttribute("master_field_lid",ls_tmp)

            #IF NOT cl_null(ls_tmp) THEN
            #   LET ls_field_allkeys = ls_field_allkeys, ls_tmp, ','
            #   LET ls_var_allkeys = ls_var_allkeys, g_properties.getValue("master_var_title"), '.', ls_tmp, ','
            #END IF

            LET ls_tmp = g_properties.getValue("master_var_title"),".",n.getAttribute("lid")
            CALL g_properties.addAttribute("master_var_lid",ls_tmp)
            
            #單頭tree用parent pid
            IF g_properties.getValue("type_id") <> "i04" THEN
               LET ls_tmp = n.getAttribute("pid")
               CALL g_properties.addAttribute("master_field_pid",ls_tmp)
               LET ls_tmp = g_properties.getValue("master_var_title"),".",n.getAttribute("pid")
               CALL g_properties.addAttribute("master_var_pid",ls_tmp)
            ELSE
               LET ls_tmp = n.getAttribute("pid")
               CALL g_properties.addAttribute("master_field_lid",ls_tmp)
               LET ls_tmp = g_properties.getValue("master_var_title"),".",n.getAttribute("lid")
               CALL g_properties.addAttribute("master_var_lid",ls_tmp)
            END IF

            #單頭tree用type
            FOR li_idx = 1 TO 10
               LET ls_name = adzp150_create_name(li_idx, "master_type_mark", "<<<") 
               LET ls_value = "#該樣板不需此段落:"
               CALL g_properties.addAttribute(ls_name,ls_value)
            END FOR
            
            LET ls_tmp = ""
            FOR li_idx = 1 TO 7
            
               LET ls_name = adzp150_create_name(li_idx, "type", "<<<") 
               LET ls_value = n.getAttribute(ls_name)
               
               IF cl_null(ls_value) THEN
                  LET ls_name = "browser_max_lv"
                  LET ls_value = li_idx USING "<<<"
                  CALL g_properties.addAttribute(ls_name,ls_value)
                  LET ls_tmp = ls_tmp.subString(1,ls_tmp.getLength()-1)
                  CALL g_properties.addAttribute("browser_vars_type_list",ls_tmp)
                  EXIT FOR
               ELSE
                  LET ls_tmp = ls_tmp, "g_browser[li_idx].b_",ls_value,","
               END IF
               
               LET ls_name = adzp150_create_name(li_idx, "master_field_type", "<<<") 
               CALL g_properties.addAttribute(ls_name,ls_value)
    
               LET ls_value = g_properties.getValue("master_var_title"),".",ls_value
               LET ls_name = adzp150_create_name(li_idx, "master_var_type", "<<<") 
               CALL g_properties.addAttribute(ls_name,ls_value)
               
               LET ls_name = adzp150_create_name(li_idx, "master_type_mark", "<<<") 
               LET ls_value = ""
               CALL g_properties.addAttribute(ls_name,ls_value)
            END FOR
            
            #單頭pk
            LET ls_tmp = n.getAttribute("pk")
            IF cl_null(ls_tmp) THEN
               DISPLAY cl_getmsg("adz-01145",g_lang)
              #DISPLAY "ERROR:無設定單頭key欄位!"
               EXIT PROGRAM
            END IF
            CALL g_properties.addAttribute("master_field_pks",ls_tmp)   
            CALL adzp150_tab_master_primarykey(n.getAttribute("pk")) 
            IF NOT cl_null(ls_tmp) THEN
               LET ls_field_allkeys = ls_field_allkeys, ls_tmp, ','
               LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
               WHILE lst_token.hasMoreTokens()
                  LET ls_token = lst_token.nextToken()
                  LET ls_var_allkeys = ls_var_allkeys, g_properties.getValue("master_var_title"), '.', ls_token, ','
               END WHILE
            END IF
            
            #單頭tree用描述欄位
            LET ls_tmp = n.getAttribute("desc")
            IF g_properties.getValue("type_id") = "i05" OR
               g_properties.getValue("type_id") = "i08" THEN
               #g_properties.getValue("type_id_t") = "i13" THEN
               CALL adzp150_create_desc(ls_tmp,'m')
            END IF

            #去掉最後的逗點 
            LET ls_field_allkeys = ls_field_allkeys.subString(1,ls_field_allkeys.getLength()-1)
            LET ls_var_allkeys = ls_var_allkeys.subString(1,ls_var_allkeys.getLength()-1)
            CALL g_properties.addAttribute("master_field_allkeys",ls_field_allkeys)   
            CALL g_properties.addAttribute("master_var_allkeys",ls_var_allkeys)   
            
            #因應i09,i10 寫入bs資訊
            IF g_properties.getValue("type_id_t") = "c01a" OR
               g_properties.getValue("type_id_t") = "c03a" OR
               g_properties.getValue("type_id_t") = "i09"  OR
               g_properties.getValue("type_id_t") = "i10"  OR
               g_properties.getValue("type_id_t") = "i12"  THEN
               CALL adzp150_create_bs_info(n.getAttribute("pk"))
            END IF
            
            #speed(提速檔相關)
            LET ls_tmp = n.getAttribute("speed")            
            CALL g_properties.addAttribute("master_spped_tbl",ls_tmp)
            LET ls_tmp = adzp150_get_table_pre(ls_tmp)
            CALL g_properties.addAttribute("master_spped_tbl_pre",ls_tmp)
            CALL g_properties.addAttribute("master_spped_lid",n.getAttribute("slid"))
            CALL g_properties.addAttribute("master_spped_pid",n.getAttribute("spid"))
            CALL g_properties.addAttribute("master_spped_type",n.getAttribute("stype"))
             
            CALL adzp150_xml_dataset_search(n) 
            
         WHEN "body"
            CALL g_properties.addAttribute("detail_tbl_cnt",1)
            LET ls_field_allkeys = ""
            LET ls_var_allkeys = ""
            
            #LET ls_tmp = ""
            #LET ls_tmp = n.getAttribute("detail")
            #IF ls_tmp = "Y" AND NOT cl_null(ls_tmp) THEN
            #   CALL adzp150_detail2_multitable_info(n)
            #END IF
            
            IF g_detail_chk THEN
               #多table資料擷取
               CALL adzp150_detail_multitable_info(n,1)
               LET n = n.getNext()
               CONTINUE WHILE
            ELSE
               #主table資料擷取
               CALL adzp150_detail_multitable_info(n,1)
            END IF
            
            LET g_detail_chk = TRUE
           
            #單身tree用帳別(type)
            LET ls_tmp = n.getAttribute("type")
            CALL g_properties.addAttribute("detail_field_type",ls_tmp)
            #IF NOT cl_null(ls_tmp) THEN
            #   LET ls_field_allkeys = ls_field_allkeys, ls_tmp, ','
            #   LET ls_var_allkeys = ls_var_allkeys, g_properties.getValue("detail_var_title"), '[l_ac].', ls_tmp, ','
            #END IF
            LET ls_tmp = g_properties.getValue("detail_var_title"),'[l_ac].',ls_tmp
            CALL g_properties.addAttribute("detail_var_type",ls_tmp)
           
            #單身tree用local id
            LET ls_tmp = n.getAttribute("lid")
            CALL g_properties.addAttribute("detail_field_lid",ls_tmp)

            #IF NOT cl_null(ls_tmp) THEN
            #   LET ls_field_allkeys = ls_field_allkeys, ls_tmp, ','
            #   LET ls_var_allkeys = ls_var_allkeys, g_properties.getValue("detail_var_title"), '[l_ac].', ls_tmp, ','
            #END IF
            LET ls_tmp = g_properties.getValue("detail_var_title"),'[l_ac].',ls_tmp
            CALL g_properties.addAttribute("detail_var_lid",ls_tmp)
            
            #單身tree用parent pid
            LET ls_tmp = n.getAttribute("pid")
            CALL g_properties.addAttribute("detail_field_pid",ls_tmp)

            #IF NOT cl_null(ls_tmp) THEN
            #   LET ls_field_allkeys = ls_field_allkeys, ls_tmp, ','
            #   LET ls_var_allkeys = ls_var_allkeys, g_properties.getValue("detail_var_title"), '[l_ac].', ls_tmp, ','
            #END IF
            LET ls_tmp = g_properties.getValue("detail_var_title"),'[l_ac].',ls_tmp
            CALL g_properties.addAttribute("detail_var_pid",ls_tmp)

            #單身tree用描述欄位
            LET ls_tmp = n.getAttribute("desc")
            CALL g_properties.addAttribute("detail_field_desc",ls_tmp)
            
            #判斷desc是否為主table欄位
            IF ls_tmp.getIndexOf(g_properties.getValue("detail_tbl_prefix"),1) > 0 THEN
               CALL g_properties.addAttribute("detail_field_desc_chk",ls_tmp)
            ELSE
               CALL g_properties.addAttribute("detail_field_desc_chk","''")
            END IF
            
            LET ls_tmp = g_properties.getValue("detail_var_title"),'[l_ac].',ls_tmp
            CALL g_properties.addAttribute("detail_var_desc",ls_tmp)
            
            #額外產生add-point段落
            LET ls_tmp = n.getAttribute("desc")
            IF g_properties.getValue("type_id") = "i05" OR
               g_properties.getValue("type_id") = "i04" THEN
              CALL adzp150_create_desc(ls_tmp,'m')
            END IF
            
            #對應到單頭的key
            CALL g_properties.addAttribute("detail_field_fks",n.getAttribute("fk"))   
            LET ls_tmp = n.getAttribute("fk")
            CALL adzp150_tab_detail_primarykey(ls_tmp,"fk")
            IF NOT cl_null(ls_tmp) THEN
               LET ls_field_allkeys = ls_field_allkeys, ls_tmp, ','
               LET ls_var_allkeys = ls_var_allkeys, g_properties.getValue("master_var_allkeys"),","
            END IF
            
            #本身單身的key
            CALL g_properties.addAttribute("detail_field_pks",n.getAttribute("pk"))   
            LET ls_tmp = n.getAttribute("pk")
            IF cl_null(ls_tmp) THEN
               DISPLAY cl_getmsg("adz-01146",g_lang)
              #DISPLAY "ERROR:無設定單身key欄位!"
               EXIT PROGRAM
            END IF
            CALL adzp150_tab_detail_primarykey(ls_tmp,"pk")
            IF NOT cl_null(ls_tmp) THEN
               LET ls_field_allkeys = ls_field_allkeys, ls_tmp, ','
            END IF
            
            LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
            WHILE lst_token.hasMoreTokens()
               LET ls_token = lst_token.nextToken()
               LET ls_var_allkeys = ls_var_allkeys, g_properties.getValue("detail_var_title"), '[l_ac].', ls_token, ','
            END WHILE
            
            #去掉最後的逗點
            LET ls_field_allkeys = ls_field_allkeys.subString(1,ls_field_allkeys.getLength()-1)
            LET ls_var_allkeys = ls_var_allkeys.subString(1,ls_var_allkeys.getLength()-1)
            CALL g_properties.addAttribute("detail_field_allkeys",ls_field_allkeys)
            CALL g_properties.addAttribute("detail_var_allkeys",ls_var_allkeys)

            #170203-00031 ---start---
            #因為部分程式段無法通用l_ac的指標變數，所以須新增使用g_detail_idx變數的用法
            LET lc_str = cl_replace_str(ls_var_allkeys,'[l_ac]','[g_detail_idx]')
            CALL g_properties.addAttribute("detail_var_allkeys_for_2nd_detail",lc_str)
            #170203-00031 --- end ---

            CALL adzp150_xml_dataset_search(n)

            #page特殊處理(隱藏資料)
            CALL adzp150_page_mark()
            
            
         WHEN "sql"   #最底層,不會有子節點
            CASE n.getAttribute("id")
               WHEN "forupd_sql"
                  CALL g_properties.addAttribute("master_sql_forupd",n.getAttribute("query"))
                  #master_sql_forupd_pre
                  LET ls_tmp = n.getAttribute("query")
                  LET ls_tmp = ls_tmp.subString(1,ls_tmp.getIndexOf('FROM',1)-1)
                  CALL g_properties.addAttribute("master_sql_forupd_pre",ls_tmp)
                  
                  #master_sql_forupd_post
                  LET ls_tmp = n.getAttribute("query")
                  LET ls_tmp = ls_tmp.subString(ls_tmp.getIndexOf('WHERE',1),ls_tmp.getLength())
                  CALL g_properties.addAttribute("master_sql_forupd_post",ls_tmp)
               
               WHEN "cs_sql"
                  CALL g_properties.addAttribute("cs_sql",n.getAttribute("query"))
               
               WHEN "forupd_sql_detail"
                  CALL g_properties.addAttribute("detail_sql_forupd",adzp150_field_filter(n.getAttribute("query"),"sql"))
               
               WHEN "b_fill_sql"
                  LET ls_tmp = adzp150_field_filter(n.getAttribute("query"),"sql")
                  CALL g_properties.addAttribute("detail_fill_sql",ls_tmp)
                  LET ls_tmp2 = ls_tmp.subString(1,ls_tmp.getIndexOf("WHERE",1)-1)
                  CALL g_properties.addAttribute("detail_fill_sql_pre",ls_tmp2)
                  LET ls_tmp2 = ls_tmp.subString(ls_tmp.getIndexOf("WHERE",1)-1,ls_tmp.getLength())
                  CALL g_properties.addAttribute("detail_fill_sql_post",ls_tmp2)
               
               WHEN "append"
                  #判斷是第幾個多語言table
                  LET ls_idx = g_properties.getValue("master_multi_table_idx")
                  IF cl_null(ls_idx) THEN
                     LET li_idx = 1
                  ELSE
                     LET li_idx = ls_idx + 1
                  END IF
                  CALL g_properties.addAttribute("master_multi_table_idx",li_idx)
                  CALL adzp150_master_multi_table(n)
                  CALL adzp150_master_multi_table_upd(n)
                  CALL adzp150_master_multi_table_bak(n)
                  CALL adzp150_update_item_action('m',n)
               
               WHEN "detail_append"
               #   CALL adzp150_detail_multi_table(n)
               #   CALL adzp150_update_item_action('d',n)
               
               OTHERWISE 
                  DISPLAY "ERROR(6): dataset內的",n.getAttribute("id"),"不符合需求定義"
            
            END CASE
            
         OTHERWISE 
            DISPLAY "ERROR(7): dataset內的",n.getTagName(),"節點不符合需求定義"
      END CASE 

      LET n = n.getNext()  #同一層 next sibling    
   END WHILE
   
   #先行取得第二階單身資料
   WHILE n2 IS NOT NULL
      IF n2.getTagName() = "body" THEN
         CALL adzp150_detail_multitable_info(n2,2)
      END IF
      LET n2 = n2.getNext()  #同一層 next sibling    
   END WHILE
   
END FUNCTION 

#+ 自定sql
PRIVATE FUNCTION adzp150_master_referesh()
   DEFINE ls_tmp          STRING
   DEFINE lst_token       base.StringTokenizer
   DEFINE ls_token        STRING
   DEFINE lst_token2      base.StringTokenizer
   DEFINE ls_token2       STRING
   DEFINE ls_field_list   STRING
   DEFINE ls_var_list     STRING
   DEFINE ls_return       STRING
   DEFINE ls_key_list     STRING
   DEFINE ls_name         STRING
   DEFINE ls_value        STRING
   DEFINE ls_site_using   STRING
   
   #組合要撈出的欄位, 以及對應的單頭變數
   LET ls_tmp     = g_properties.getValue("master_fields_all")
   LET lst_token  = base.StringTokenizer.create(ls_tmp,",")
   LET ls_tmp     = g_properties.getValue("master_vars_all")
   LET lst_token2 = base.StringTokenizer.create(ls_tmp,",")
   
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_token2 = lst_token2.nextToken()
      #先去除( )的內容
      IF ls_token.getIndexOf("(",1) THEN
         LET ls_token = ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
      END IF
      
      #檢查該欄位是否存在於單頭主table中
      IF adzp150_chk_intbl(g_properties.getValue("master_tbl_name"),ls_token) THEN
         LET ls_field_list = ls_field_list, ls_token, ","
         LET ls_var_list   = ls_var_list,  ls_token2, ","
      END IF
   END WHILE
   LET ls_field_list = ls_field_list.subString(1,ls_field_list.getLength()-1)
   LET ls_var_list   = ls_var_list.subString(1,ls_var_list.getLength()-1)
   
   #mdl_field_list(master_field_list)
   LET ls_name = "master_field_list"
   LET ls_value = adzp150_add_table_name(ls_field_list,'m')
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #mdl_var_list(master_var_list)
   LET ls_name = "master_var_list"
   LET ls_value = ls_var_list
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #將所有的key併入搜尋條件
   LET ls_tmp     = g_properties.getValue("master_field_allkeys")
   LET lst_token  = base.StringTokenizer.create(ls_tmp,",")
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_key_list = ls_key_list, 't0.', ls_token, " = ? AND "
   END WHILE
   LET ls_key_list = ls_key_list.subString(1,ls_key_list.getLength()-5)
   
   #mdl_key_list(master_key_wc)
   LET ls_name = "master_key_wc"
   LET ls_value = ls_key_list
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #LET ls_return = adzp150_make_slice("a52")
   
   #判斷是否有site, 若有則加入using
   IF NOT cl_null(g_properties.getValue("master_using_site")) THEN
      LET ls_site_using = 'g_site,'
   END IF
   
   LET ls_return = "EXECUTE ",g_properties.getValue("general_prefix"),
                   "_master_referesh USING ",ls_site_using,g_properties.getValue("master_var_allkeys"),
                   " INTO ",g_properties.getValue("master_var_list"),
                   g_properties.getValue("master_reference_fetch_field")

   RETURN ls_return 
   
END FUNCTION 


#+ convert fields to screen record #單身 查詢條件 cs_sr_name
PRIVATE FUNCTION adzp150_change_cs_sr_detail(p_name,p_list)  
   DEFINE p_name     STRING
   DEFINE p_list     STRING
   DEFINE l_prefix   STRING
   DEFINE lst_token  base.StringTokenizer
   DEFINE ls_token   STRING
   DEFINE l_new      STRING

   LET l_prefix = p_name
   LET lst_token = base.StringTokenizer.create(p_list.trim(), ',')
 
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()      
      LET l_new = l_new, l_prefix, "[1].", ls_token.trim(), ","      
   END WHILE   

   RETURN l_new.subString(1, l_new.getLength()-1)
      
END FUNCTION

#+ 定義g_browser 變數
PRIVATE FUNCTION adzp150_browser_fields_define(p_list,p_append_list,pk_list,ps_ref_tbl,ps_ref_wc,ps_ref_type)

   DEFINE p_list            STRING
   DEFINE p_append_list     STRING
   DEFINE pk_list           STRING
   DEFINE ps_ref_wc         STRING
   DEFINE ps_ref_tbl        STRING
   DEFINE ps_ref_type       STRING
   DEFINE l_prefix          STRING
   DEFINE lst_token         base.StringTokenizer
   DEFINE ls_token          STRING
   DEFINE lst_token2        base.StringTokenizer
   DEFINE ls_token2         STRING
   DEFINE l_new             STRING
   DEFINE l_str_token_count LIKE type_t.num5 
   DEFINE l_token_count     LIKE type_t.num5 
   DEFINE li_diff_tbl       LIKE type_t.num5  #辨識是否有兩種以上table
   DEFINE ls_tabid          STRING
   DEFINE ls_tmp            STRING
   DEFINE ls_tmp2           STRING
   DEFINE ls_ref_type       STRING
   DEFINE ls_ref_wc         STRING
   DEFINE ls_replace        LIKE type_t.num5
   DEFINE ls_type           STRING
   DEFINE ls_tmp_tbl        STRING
   DEFINE ls_name           STRING

   #處理必要變動欄位
   LET lst_token = base.StringTokenizer.create(p_list.trim(), ',')
   LET l_new = ""
   LET ls_tmp_tbl = ""   
   LET li_diff_tbl = FALSE
   LET l_str_token_count = lst_token.countTokens()
      
   LET l_token_count = 0
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET l_token_count = l_token_count + 1

      #修正table代碼
      IF ls_token.getIndexOf('(',1) = 0 THEN
         LET l_prefix = adzp150_define_table_name(ls_token)
      END IF
      
      #i05 insert 用
      LET ls_tmp2 = ls_tmp2,"?"

      #比對是否存在不同table
      IF NOT li_diff_tbl AND ls_tabid IS NOT NULL THEN
         IF NOT ls_tabid.equals(l_prefix) THEN
            LET li_diff_tbl = TRUE
         END IF
      END IF
      LET ls_tabid = l_prefix
     
      #非一般參照欄位
      IF ls_token.getIndexOf('(',1) > 0 THEN
         LET ls_type = ls_token.subString(ls_token.getIndexOf('(',1)+1,ls_token.getIndexOf(')',1)-1)
         LET ls_name = "b_",ls_token.subString(1,ls_token.getIndexOf('(',1)-1)
         CASE ls_type.toUpperCase()
            WHEN "STRING"
               #LET ls_fb_new = ls_fb_new, ls_name, " STRING" #STRING過於複雜, 導致讀資料時發生錯誤, 暫時停用
               LET l_new = l_new, ls_name, " LIKE type_t.chr500"
            WHEN "DATETIME"
               LET l_new = l_new, ls_name, " DATETIME YEAR TO SECOND"
            WHEN "TIMESTAMP"
               LET l_new = l_new, ls_name, " DATETIME YEAR TO FRACTION(5)"
            WHEN "BLOB"
               LET l_new = l_new, ls_name, " LIKE type_t.", ls_type
               LET ls_tmp = g_properties.getValue("general_lob")
               LET ls_tmp = ls_tmp, "   LOCATE g_browser.",ls_name," IN FILE ", "\n"
               CALL g_properties.addAttribute("general_lob_start",ls_tmp)
               LET ls_tmp = g_properties.getValue("general_lob_end")
               LET ls_tmp = ls_tmp, "   FREE g_browser.",ls_name, "\n"
               CALL g_properties.addAttribute("general_lob_end",ls_tmp)
            WHEN "CLOB"
               LET l_new = l_new, ls_name, " LIKE type_t.", ls_type
               LET ls_tmp = g_properties.getValue("general_lob")
               LET ls_tmp = ls_tmp, "   LOCATE g_browser.",ls_name," IN FILE ", "\n"
               CALL g_properties.addAttribute("general_lob_start",ls_tmp)
               LET ls_tmp = g_properties.getValue("general_lob_end")
               LET ls_tmp = ls_tmp, "   FREE g_browser.",ls_name, "\n"
               CALL g_properties.addAttribute("general_lob_end",ls_tmp)
            OTHERWISE
               IF ls_type.getIndexOf('.',1) > 0 THEN
                  LET l_new = l_new, ls_name, " LIKE ", ls_type
               ELSE
                  LET l_new = l_new, ls_name, " LIKE type_t.", ls_type
               END IF
         END CASE
         #LET l_new = l_new, li_space SPACES, ls_name , " LIKE type_t.", ls_type
         #LET ls_tmp_tbl = ls_tmp_tbl,li_space SPACES,ls_token.subString(1,ls_token.getIndexOf('(',1)-1), " VARCHAR(500),\n" #170906-00011#1
         LET ls_type = ls_token.subString(ls_token.getIndexOf('(',1)+1,ls_token.getIndexOf(')',1)-1)
         IF ls_type.getIndexOf('.',1) > 0 THEN
            LET ls_type = ls_type
         ELSE
            LET ls_type = "type_t.", ls_type
         END IF
         LET ls_tmp_tbl = ls_tmp_tbl,li_space SPACES,ls_token.subString(1,ls_token.getIndexOf('(',1)-1), " LIKE ", ls_type,",\n" #20170816 
     ELSE
         LET l_new = l_new, li_space SPACES, "b_",ls_token.trim() , " LIKE ", l_prefix ,ls_token.trim()
         #LET ls_tmp_tbl = ls_tmp_tbl, li_space SPACES, ls_token.trim() , " VARCHAR(500),\n" #170906-00011#1
         LET ls_tmp_tbl = ls_tmp_tbl,li_space SPACES,ls_token, " LIKE ",ls_tabid,"",ls_token,",\n" #170906-00011#1
      END IF 
      
      IF l_str_token_count != l_token_count THEN
         LET l_new = l_new ,",\n",li_space SPACES
         LET ls_tmp2 = ls_tmp2,","
      END IF 
   END WHILE
   LET ls_tmp = l_new.subString( 1, l_new.getLength())
   LET ls_tmp_tbl = ls_tmp_tbl.subString( 1, ls_tmp_tbl.getLength()-2)
   CALL g_properties.addAttribute("browser_fields_define",ls_tmp)
   CALL g_properties.addAttribute("browser_tmp_tbl_define", ls_tmp_tbl)
   CALL g_properties.addAttribute("browser_fields_qm", ls_tmp2)
   
   #設定Browser內若有使用referance時的wc
   IF li_diff_tbl THEN
      IF NOT cl_null(ps_ref_wc) AND NOT cl_null(ps_ref_tbl) THEN
         CASE ps_ref_type
            WHEN "left"
               LET ls_ref_type = " LEFT JOIN "
            WHEN "right"
               LET ls_ref_type = " RIGHT JOIN " 
            WHEN "inner"
               LET ls_ref_type = " INNER JOIN "
            OTHERWISE
               LET ls_ref_type = " LEFT JOIN "
         END CASE   

         LET ls_tmp = ""

         LET lst_token = base.StringTokenizer.create(ps_ref_wc, ',')
         LET lst_token2 = base.StringTokenizer.create(ps_ref_tbl, ',')
         WHILE lst_token.hasMoreTokens()
            LET ls_token = lst_token.nextToken()
            LET ls_token2 = lst_token2.nextToken()
            
            LET ls_tmp = ls_tmp, 
                         ls_ref_type,ls_token2, 
                         " ON ", ls_token, " "
            
         END WHILE
         
         LET ls_tmp = adzp150_replace_var(ls_tmp)

         CALL g_properties.addAttribute("browser_addiction_sql",ls_tmp)
         
      #ELSE
      #   DISPLAY "ERROR(8):Browser內設定Referance表,但未設定關聯"
      END IF
      
   END IF
   
   #檢視是否有值
   IF NOT cl_null(ls_tmp) THEN
      RETURN TRUE
   ELSE
      RETURN FALSE
   END IF
   
END FUNCTION


#+ convert fields to var
PRIVATE FUNCTION adzp150_master_fields_define(pi_page,p_list)  
   DEFINE p_list                  STRING
   DEFINE l_prefix                STRING
   DEFINE lst_token               base.StringTokenizer
   DEFINE ls_token                STRING
   DEFINE li_i                    LIKE type_t.num10    #空幾格
   DEFINE pi_page                 LIKE type_t.num5
   DEFINE l_token_count           LIKE type_t.num5
   DEFINE l_str_token_count       LIKE type_t.num5
   DEFINE ls_master_fields_define STRING       #form_head_field代換變數
   DEFINE ls_master_key_define    STRING       #key欄位_t宣告
   DEFINE ls_var_h_fill           STRING       #var_h_fill代換變數
   DEFINE ls_master_fields_all    STRING       #var_h_fill代換變數
   DEFINE ls_fb_new               STRING       #form_head_field填入值
   DEFINE ls_vb_new               STRING       #var_h_fill填入值
   DEFINE ls_tmp                  STRING
   DEFINE ls_tmp2                 STRING
   DEFINE ls_value                STRING
   DEFINE ls_type                 STRING       #變數型態
   DEFINE ls_name                 STRING       #變數名稱
   DEFINE ls_master_all_var       STRING       #所有單頭變數
   DEFINE ls_master_upd_var       STRING       #主table的單頭變數
   DEFINE ls_field_sql_head       STRING       #所有必需insert進單頭table的欄位 
   DEFINE ls_master_allkeys       STRING
   DEFINE ls_master_modify        STRING       #modid,moddt用(update時更新)
   DEFINE ls_stus_field           STRING       #status欄位名稱   #181128-00048


   #定義處理變數名稱
   LET ls_master_fields_define = "master_fields_define"
   LET ls_var_h_fill = "master_vars_all"
   LET ls_master_fields_all = "master_fields_all"
   
   #master_fields_all過濾&整理 
   LET lst_token = base.StringTokenizer.create(p_list.trim(), ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      IF ls_token.getIndexOf("(",1) THEN
         LET ls_vb_new = ls_vb_new,ls_token.subString(1,ls_token.getIndexOf("(",1)-1),","
      ELSE
         LET ls_vb_new = ls_vb_new,ls_token,","
      END IF
   END WHILE
   
   #去除最後逗點
   LET ls_vb_new = ls_vb_new.subString(1,ls_vb_new.getLength())

   CALL g_properties.addAttribute(ls_master_fields_all, ls_vb_new)

   LET ls_vb_new = ""
   
   LET li_i = 1
   LET lst_token = base.StringTokenizer.create(p_list.trim(), ',')
   LET l_str_token_count = lst_token.countTokens()

   LET ls_master_allkeys = g_properties.getValue("master_field_allkeys")
   
   LET l_token_count = 0
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET l_token_count = l_token_count + 1
 
      #逐個串接DEFINE for form_head_master
      LET ls_fb_new = ls_fb_new, (li_space*li_i) SPACES

      #逐個串接 for var_h_fill
      IF pi_page = 1 THEN 
         LET ls_vb_new = ls_vb_new, g_properties.getValue("master_var_title")
      ELSE
         LET ls_tmp = "master_var_title",pi_page USING "<<<"
         LET ls_vb_new = ls_vb_new, g_properties.getValue(ls_tmp)
      END IF

      #如果標示欄位名稱的字串中含有 (),表示要使用 type_t, 並且是用()內指定的型態
      CASE 
         WHEN ls_token.getIndexOf("(",1) 
            LET l_prefix = "type_t."
            LET ls_name = ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
            LET ls_type = ls_token.subString(ls_token.getIndexOf("(",1)+1,ls_token.getIndexOf(")",1)-1)
            CASE ls_type.toUpperCase()
               WHEN "STRING"
                  #LET ls_fb_new = ls_fb_new, ls_name, " STRING" #STRING過於複雜, 導致讀資料時發生錯誤, 暫時停用
                  LET ls_fb_new = ls_fb_new, ls_name, " LIKE type_t.chr500"
               WHEN "DATETIME"
                  IF ls_master_allkeys.getIndexOf(ls_name,1) > 0 THEN
                     LET ls_master_key_define = ls_master_key_define, "DEFINE g_",ls_name,"_t DATETIME YEAR TO SECOND\n"
                  END IF
                  LET ls_fb_new = ls_fb_new, ls_name, " DATETIME YEAR TO SECOND"
               WHEN "TIMESTAMP"
                  IF ls_master_allkeys.getIndexOf(ls_name,1) > 0 THEN
                     LET ls_master_key_define = ls_master_key_define, "DEFINE g_",ls_name,"_t DATETIME YEAR TO FRACTION(5)\n"
                  END IF
                  LET ls_fb_new = ls_fb_new, ls_name, " DATETIME YEAR TO FRACTION(5)"
               WHEN "BLOB"
                  LET ls_fb_new = ls_fb_new, ls_name, " LIKE type_t.", ls_type
                  LET ls_tmp = g_properties.getValue("general_lob")
                  LET ls_tmp = ls_tmp, "   LOCATE ",g_properties.getValue("master_var_title"),".",ls_name," IN FILE ", "\n"
                  CALL g_properties.addAttribute("general_lob_start",ls_tmp)
                  LET ls_tmp = g_properties.getValue("general_lob_end")
                  LET ls_tmp = ls_tmp, "   FREE ",g_properties.getValue("master_var_title"),".",ls_name, "\n"
                  CALL g_properties.addAttribute("general_lob_end",ls_tmp)
               WHEN "CLOB"
                  LET ls_fb_new = ls_fb_new, ls_name, " LIKE type_t.", ls_type
                  LET ls_tmp = g_properties.getValue("general_lob")
                  LET ls_tmp = ls_tmp, "   LOCATE ",g_properties.getValue("master_var_title"),".",ls_name," IN FILE ", "\n"
                  CALL g_properties.addAttribute("general_lob_start",ls_tmp)
                  LET ls_tmp = g_properties.getValue("general_lob_end")
                  LET ls_tmp = ls_tmp, "   FREE ",g_properties.getValue("master_var_title"),".",ls_name, "\n"
                  CALL g_properties.addAttribute("general_lob_end",ls_tmp)
               OTHERWISE
                  #190826-00045#1 -S-
                  IF ls_token.getIndexOf('.',1) > 0 THEN
                     LET ls_fb_new = ls_fb_new, ls_name, " LIKE ", ls_type
                     #DISPLAY "ls_name:",ls_name, " LIKE ", ls_type
                  ELSE
                     LET ls_fb_new = ls_fb_new, ls_name, " LIKE type_t.", ls_type
                  END IF
                  #LET ls_fb_new = ls_fb_new, ls_name, " LIKE type_t.", ls_type #mark
                  #190826-00045#1 -E-
            END CASE 
            LET ls_vb_new = ls_vb_new, ".",ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
         #WHEN ls_token.getIndexOf("crtdt",1) OR ls_token.getIndexOf("moddt",1) OR ls_token.getIndexOf("cnfdt",1)
         #   LET ls_fb_new = ls_fb_new, ls_token, " DATETIME YEAR TO SECOND"
         #   LET ls_vb_new = ls_vb_new, ".",ls_token.trim()
         OTHERWISE
            IF ls_master_allkeys.getIndexOf(ls_token,1) > 0 THEN
               LET ls_master_key_define = ls_master_key_define, 
                                          "DEFINE g_",ls_token,"_t LIKE ",g_properties.getValue("master_tbl_name"),".",ls_token,'\n'
            END IF
            LET l_prefix = adzp150_define_table_name(ls_token)
            LET ls_fb_new = ls_fb_new, ls_token.trim(), " LIKE ", l_prefix, ls_token.trim()
            LET ls_vb_new = ls_vb_new, ".",ls_token.trim()
      END CASE
      IF l_str_token_count != l_token_count THEN
         LET ls_fb_new = ls_fb_new, ", \n"
         LET ls_vb_new = ls_vb_new, ","
      END IF
   END WHILE

   IF pi_page = 1 THEN 
   ELSE
      LET ls_master_fields_define = ls_master_fields_define,pi_page USING "<<<<"
      LET ls_var_h_fill = ls_var_h_fill,pi_page USING "<<<<"
   END IF 

   CALL g_properties.addAttribute("master_key_t_define", ls_master_key_define)
   CALL g_properties.addAttribute(ls_master_fields_define, ls_fb_new.trim())
   CALL g_properties.addAttribute(ls_var_h_fill, ls_vb_new.trim())
   
   #額外處理var_sql_head,field_sql_head
   LET ls_tmp = g_properties.getValue("master_vars_all")
   LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
   LET ls_tmp = g_properties.getValue("master_var_title"),'.',g_properties.getValue("master_tbl_prefix"), "modid,",
                g_properties.getValue("master_var_title"),'.',g_properties.getValue("master_tbl_prefix"), "moddt,",   #181128-00048 modify
                g_properties.getValue("master_var_title"),'.',g_properties.getValue("master_tbl_prefix"), "stus"   #181128-00048
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      IF adzp150_chk_intbl(g_properties.getValue("master_tbl_name"),ls_token) THEN
         IF ls_tmp.getIndexOf(ls_token,1) > 0 THEN
            #modid,moddt,stus欄位
            LET ls_master_modify = ls_master_modify, ls_token,","
         END IF
         LET ls_master_upd_var = ls_master_upd_var, ls_token,","
      END IF
      LET ls_master_all_var = ls_master_all_var, ls_token,","
   END WHILE
   LET ls_master_upd_var = ls_master_upd_var.subString(1,ls_master_upd_var.getLength()-1)
   LET ls_master_all_var = ls_master_all_var.subString(1,ls_master_all_var.getLength()-1)
   LET ls_master_modify = ls_master_modify.subString(1,ls_master_modify.getLength()-1)
   
   CALL g_properties.addAttribute("master_vars_all",ls_master_all_var)  
   CALL g_properties.addAttribute("master_vars_update",ls_master_upd_var)   
   CALL g_properties.addAttribute("master_modify_vars_update",ls_master_modify)
   LET ls_master_modify = ''
   
   LET ls_tmp = g_properties.getValue("master_fields_all")
   LET lst_token = base.StringTokenizer.create(ls_tmp, ',')
   LET ls_tmp = g_properties.getValue("master_tbl_prefix"), "modid,",
                g_properties.getValue("master_tbl_prefix"), "moddt,",   #181128-00048 modify
                g_properties.getValue("master_tbl_prefix"), "stus"   #181128-00048
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      IF adzp150_chk_intbl(g_properties.getValue("master_tbl_name"),ls_token) THEN
         IF ls_tmp.getIndexOf(ls_token,1) > 0 THEN
            #modid,moddt,stus欄位
            LET ls_master_modify = ls_master_modify, ls_token,","
         END IF
         LET ls_field_sql_head = ls_field_sql_head, ls_token,","
      END IF
   END WHILE
   LET ls_field_sql_head = ls_field_sql_head.subString(1,ls_field_sql_head.getLength()-1)
   LET ls_master_modify = ls_master_modify.subString(1,ls_master_modify.getLength()-1)
   #DISPLAY "upd_head_fields:",ls_field_sql_head
   CALL g_properties.addAttribute("master_fields_update",ls_field_sql_head)
   CALL g_properties.addAttribute("master_modify_fields_update",ls_master_modify)
   
   #額外判斷, 若無modid,moddt欄位則加註
   IF cl_null(g_properties.getValue("master_modify_fields_update")) OR
      cl_null(g_properties.getValue("master_modify_vars_update"))   THEN
      CALL g_properties.addAttribute("general_modify_mark","#該樣板不需此段落")
   END IF   

   #複製段清空key欄位desc
   LET ls_tmp = ""
   LET lst_token = base.StringTokenizer.create(p_list, ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      #先判斷是否為參照欄位
      IF ls_token.getIndexOf('_desc',1) > 0 THEN
         LET ls_tmp = ls_token.subString(1,ls_token.getIndexOf('_desc',1)-1)
         #再判斷是否為key欄位的參照欄位
         LET ls_tmp2 = g_properties.getValue("master_field_allkeys")
         IF ls_tmp2.getIndexOf(ls_tmp,1) > 0 THEN
            LET ls_value = ls_value, 
                           "   LET ", g_properties.getValue("master_var_title"),".",
                           ls_token.subString(1,ls_token.getIndexOf('(',1)-1), " = ''\n",
                           "   DISPLAY BY NAME ", g_properties.getValue("master_var_title"),".",
                           ls_token.subString(1,ls_token.getIndexOf('(',1)-1), "\n"
         END IF
      END IF
   END WHILE
    
   CALL g_properties.addAttribute("general_repro_reset",ls_value)   

END FUNCTION


#+ convert fields to var
PRIVATE FUNCTION adzp150_detail_fields_define(pi_page,p_list)  
   DEFINE p_list                   STRING
   DEFINE l_prefix                 STRING
   DEFINE lst_token                base.StringTokenizer
   DEFINE ls_token                 STRING
   DEFINE li_i                     LIKE type_t.num10    
   DEFINE pi_page                  LIKE type_t.num5
   DEFINE l_token_count            LIKE type_t.num5
   DEFINE l_str_token_count        LIKE type_t.num5
   DEFINE ls_detail_fields_define  STRING               #form_body_field代換變數
   DEFINE ls_detail_vars_all       STRING               #detail_vars_all代換變數
   DEFINE ls_detail_fields_all     STRING               #detail_fields_all代換變數            
   DEFINE ls_fb_new                STRING               #form_body_field填入值
   DEFINE ls_vb_new                STRING               #detail_vars_all填入值
   DEFINE ls_tmp                   STRING
   DEFINE ls_name                  STRING 
   DEFINE ls_type                  STRING    
   DEFINE ls_page                  STRING   
      
   #定義處理變數名稱
   LET ls_detail_fields_define = "detail_fields_define"
   LET ls_detail_vars_all = "detail_vars_all"
   LET ls_detail_fields_all = "detail_fields_all"
   
   LET ls_page = pi_page USING "<<<"
   
   CALL g_properties.addAttribute("detail_page_field"||ls_page,p_list)

   LET li_i = 1
   LET lst_token = base.StringTokenizer.create(p_list.trim(), ',')
   LET l_str_token_count = lst_token.countTokens()

   LET l_token_count = 0
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET l_token_count = l_token_count + 1

      #取得每個page的第一個欄位
      IF l_token_count = 1 THEN
         LET ls_tmp = g_properties.getValue("detail_var_title"||ls_page),"[l_ac].",ls_token
         CALL g_properties.addAttribute("detail_fisrt_var"||ls_page,ls_tmp)
         CALL g_properties.addAttribute("detail_fisrt_field"||ls_page,ls_token)
      END IF
 
      #逐個串接DEFINE for form_body_detail
      LET ls_fb_new = ls_fb_new, (li_space*li_i) SPACES

      #逐個串接 for detail_vars_all
      IF pi_page = 1 THEN 
         LET ls_vb_new = ls_vb_new, g_properties.getValue("detail_var_title")
      ELSE
         LET ls_tmp = "detail_var_title",pi_page USING "<<<"
         LET ls_vb_new = ls_vb_new, g_properties.getValue(ls_tmp)
      END IF
    
      #如果標示欄位名稱的字串中含有 (),表示要使用 type_t, 並且是用()內指定的型態
      CASE
         WHEN ls_token.getIndexOf("(",1)
            LET l_prefix = "type_t."
            LET ls_name = ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
            LET ls_type = ls_token.subString(ls_token.getIndexOf("(",1)+1,ls_token.getIndexOf(")",1)-1)
            
            IF ls_type.getIndexOf('.',1) > 0 THEN
               LET ls_fb_new = ls_fb_new, ls_name, " LIKE ", ls_type
               LET ls_vb_new = ls_vb_new, "[l_ac].",ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
            ELSE
               CASE ls_type.toUpperCase()
                  WHEN "STRING"
                     #LET ls_fb_new = ls_fb_new, ls_name, " STRING"
                     LET ls_fb_new = ls_fb_new, ls_name, " LIKE type_t.chr500"
                  WHEN "DATETIME"
                     LET ls_fb_new = ls_fb_new, ls_name, " DATETIME YEAR TO SECOND"
                  WHEN "TIMESTAMP"
                     LET ls_fb_new = ls_fb_new, ls_name, " DATETIME YEAR TO FRACTION(5)"
                  OTHERWISE
                     LET ls_fb_new = ls_fb_new, ls_name, " LIKE type_t.", ls_type
               END CASE 
               LET ls_vb_new = ls_vb_new, "[l_ac].",ls_token.subString(1,ls_token.getIndexOf("(",1)-1)
            END IF
            
         WHEN ls_token.getIndexOf("crtdt",1) OR ls_token.getIndexOf("moddt",1) OR ls_token.getIndexOf("cnfdt",1)
            LET ls_fb_new = ls_fb_new, ls_token, " DATETIME YEAR TO SECOND"
            LET ls_vb_new = ls_vb_new, "[l_ac].",ls_token.trim()
            LET ls_name = ls_token
            
         OTHERWISE
            #修正table代碼
            LET l_prefix = adzp150_define_table_name(ls_token)
            LET ls_fb_new = ls_fb_new, ls_token.trim(), " LIKE ", l_prefix ,ls_token.trim()
            LET ls_vb_new = ls_vb_new, "[l_ac].",ls_token.trim()
            LET ls_name = ls_token
            
      END CASE

      IF l_str_token_count != l_token_count THEN
         LET ls_fb_new = ls_fb_new, ", \n"
         LET ls_vb_new = ls_vb_new, ","
      END IF
      
   END WHILE

   IF pi_page = 1 THEN 
   ELSE
      LET ls_detail_fields_define = ls_detail_fields_define,pi_page USING "<<<<"
      LET ls_detail_vars_all = ls_detail_vars_all,pi_page USING "<<<<"
      #LET ls_vb_new = ",",ls_vb_new
      LET ls_detail_fields_all = ls_detail_fields_all,pi_page USING "<<<<"
   END IF 

   CALL g_properties.addAttribute(ls_detail_fields_define, ls_fb_new.trim())
   CALL g_properties.addAttribute(ls_detail_vars_all, ls_vb_new.trim())
   CALL g_properties.addAttribute(ls_detail_fields_all, p_list.trim())
   
END FUNCTION

#+ 設定單頭key值
PRIVATE FUNCTION adzp150_tab_master_primarykey(p_keys)
   DEFINE p_prefix   STRING
   DEFINE p_keys     STRING
   DEFINE lst_token  base.StringTokenizer
   DEFINE ls_token   STRING
   DEFINE ls_name    STRING
   DEFINE ls_value   STRING
   DEFINE li_cnt     LIKE type_t.num5 
   #171204-00032#1 - Start -
   DEFINE ls_tmp     STRING
   DEFINE ls_tmp2    STRING
   #171204-00032#1 - End -

   LET li_cnt = 1   
   LET lst_token = base.StringTokenizer.create(p_keys.trim(), ',')
   
   WHILE lst_token.hasMoreTokens()
       LET ls_token = lst_token.nextToken()
       LET ls_name = "master_field_pk", li_cnt USING "&&"
       CALL g_properties.addAttribute(ls_name,ls_token) 
       LET ls_value = g_properties.getValue("master_var_title"),".",ls_token
       LET ls_name = "master_var_pk", li_cnt USING "&&"
       CALL g_properties.addAttribute(ls_name,ls_value)

       LET li_cnt =  li_cnt + 1
       
   END WHILE 
   LET li_cnt =  li_cnt - 1

   CALL g_properties.addAttribute("master_pk_num",li_cnt )

   #171204-00032#1 - Start -
   LET ls_name = "master_first_input_field"
   
   #單頭可輸入欄位清單
   LET ls_tmp  = g_properties.getValue("master_input_fields_list"),","
   
   #第一個單頭key
   LET ls_tmp2 = g_properties.getValue("master_field_pk01"),","
   DISPLAY 'master_field_pk01:',ls_tmp2
   #檢查master_field_pk01是否可輸入 
   IF ls_tmp.getIndexOf(ls_tmp2,1) > 0 THEN
      LET ls_value = g_properties.getValue("master_field_pk01") 
   ELSE
      #取得第一個可輸入的欄位名稱
      IF ls_tmp.getIndexOf(',',1) > 0 THEN
         LET ls_value = ls_tmp.subString(1,ls_tmp.getIndexOf(',',1)-1)
      ELSE 
         LET ls_value = g_properties.getValue("master_input_fields_list")
      END IF 
   END IF
   CALL g_properties.addAttribute(ls_name,ls_value)
   #171204-00032#1 - End -
   
END FUNCTION



#+ 設定單身key值
PRIVATE FUNCTION adzp150_tab_detail_primarykey(p_keys,p_type)

   DEFINE p_prefix   STRING
   DEFINE p_type     STRING
   DEFINE p_keys     STRING
   DEFINE lst_token  base.StringTokenizer
   DEFINE ls_token   STRING
   DEFINE ls_name    STRING
   DEFINE ls_value   STRING
   DEFINE li_cnt     LIKE type_t.num5 

   LET li_cnt = 1   
   LET p_prefix = "detail_field_",p_type
   LET lst_token = base.StringTokenizer.create(p_keys.trim(), ',')
   
   WHILE lst_token.hasMoreTokens()
       LET ls_token = lst_token.nextToken()
       
       LET ls_name = p_prefix, li_cnt USING "&&"
       CALL g_properties.addAttribute(ls_name,ls_token) 
       IF p_type = "fk" THEN
          LET ls_name = "master_field_pk", li_cnt USING "&&"
          LET ls_value = g_properties.getValue("master_var_title"),".",g_properties.getValue(ls_name)
          LET ls_name = "detail_var_fk", li_cnt USING "&&"
          CALL g_properties.addAttribute(ls_name,ls_value)
       ELSE
          LET ls_value = g_properties.getValue("detail_var_title"),"[l_ac].",ls_token
          LET ls_name = "detail_var_pk", li_cnt USING "&&"
          CALL g_properties.addAttribute(ls_name,ls_value)
       END IF

       LET li_cnt =  li_cnt + 1
       
   END WHILE 

   LET li_cnt =  li_cnt - 1
   
   CALL g_properties.addAttribute("detail_"||p_type||"_num",li_cnt )

END FUNCTION



#+ 設定upd_body_fields及cs_body_fields 組單身欄位頁簽
PRIVATE FUNCTION adzp150_set_upd_body_fields()

   DEFINE l_k                    LIKE type_t.num5 
   DEFINE l_upd_body_vars        STRING
   DEFINE l_upd_body_fields      STRING
   DEFINE l_upd_body_vars_page   STRING 
   DEFINE l_upd_body_fields_page STRING
   DEFINE l_tok                  base.StringTokenizer
   DEFINE l_sr                   STRING 
   DEFINE l_detail               STRING
   DEFINE l_table                STRING 
   DEFINE l_table2               STRING 
   DEFINE l_str_token_count      LIKE type_t.num5 
   DEFINE l_token_count          LIKE type_t.num5  
   DEFINE l_tag_id               STRING 
   DEFINE l_page                 LIKE type_t.num5  
   DEFINE l_str_page             STRING
   DEFINE l_upd_defaults         STRING
   DEFINE l_tmp                  STRING
   DEFINE l_detail_page          STRING
   DEFINE l_upd_body_fields_tmp  STRING #temp table用
   
   LET l_detail = g_properties.getValue("detail_tbl_prefix")
   LET l_detail = l_detail CLIPPED
   LET l_detail = "g_",l_detail,"_d"

   LET l_page = g_properties.getValue("page")  

   FOR l_k = 1 TO l_page
      LET l_upd_body_vars = ""  
      LET l_upd_body_fields = ""

      #第一頁簽的處理是特別的
      IF l_k = 1 THEN
         LET l_str_page = ""
         LET l_detail_page = l_detail.subString(1,l_detail.getIndexOf("_",3)-1),l_str_page,"_d"
         LET l_tag_id = "detail_fields_all" 
         LET l_table = g_properties.getValue("detail_tbl_name") 
         LET l_table2 = l_table
         LET l_table = l_table.subString(1,l_table.getIndexOf("_",1)-1)
      ELSE 
         LET l_str_page = l_k USING "<<<"
         #取 g_xxx_d 的第一碼,到第二個底線前一碼的意思
         LET l_detail_page = l_detail.subString(1,l_detail.getIndexOf("_",3)-1),l_str_page,"_d"
         LET l_tag_id = "detail_fields_all",l_str_page 
         LET l_table = g_properties.getValue("detail_tbl_name"||l_str_page)
         LET l_table2 = l_table         
         LET l_table = l_table.subString(1,l_table.getIndexOf("_",1)-1)
      END IF   
      
     
      LET l_upd_body_vars_page = "detail_vars_update",l_str_page
      LET l_upd_body_fields_page = "detail_fields_update",l_str_page

      LET l_tok = base.StringTokenizer.create(g_properties.getValue(l_tag_id),",")
      LET l_str_token_count = l_tok.countTokens()
      LET l_token_count = 0

      WHILE l_tok.hasMoreTokens()
         LET l_sr = l_tok.nextToken()
         LET l_token_count = l_token_count + 1
         LET l_tmp = g_properties.getValue("detail_field_pks")
         IF adzp150_chk_intbl(l_table2,l_sr) THEN
            IF l_tmp.getIndexOf(l_sr,1) > 0 AND l_k > 1 THEN
            ELSE
               LET l_upd_body_vars = l_upd_body_vars,l_detail_page,"[l_ac].",l_sr,","
               LET l_upd_body_fields = l_upd_body_fields,l_sr,","
               LET l_upd_body_fields_tmp = l_upd_body_fields_tmp, l_sr, " VARCHAR(500),\n"
            END IF
         END IF
      END WHILE 
      
      IF l_upd_body_vars.subString(l_upd_body_vars.getLength(),l_upd_body_vars.getLength()) = "," THEN
         LET l_upd_body_vars = l_upd_body_vars.subString(1,l_upd_body_vars.getLength()-1)
      END IF
      
      IF l_upd_body_fields.subString(l_upd_body_fields.getLength(),l_upd_body_fields.getLength()) = "," THEN
         LET l_upd_body_fields = l_upd_body_fields.subString(1,l_upd_body_fields.getLength()-1)
      END IF

      LET l_upd_body_fields_tmp = l_upd_body_fields_tmp.subString(1,l_upd_body_fields_tmp.getLength()-2)
      CALL g_properties.addAttribute("detail_tmp_tbl_define",l_upd_body_fields_tmp)
      CALL g_properties.addAttribute(l_upd_body_vars_page,l_upd_body_vars)
      CALL g_properties.addAttribute(l_upd_body_fields_page,l_upd_body_fields)

   END FOR   
   
END FUNCTION

#+ 判定該節點以下的子節點屬於何種類型並進行處理
PUBLIC FUNCTION adzp150_parse_element(p_node,pi_lv)
   DEFINE p_node       om.DomNode
   DEFINE pi_lv        LIKE type_t.num10     #退後 pi_lv 個層級
   DEFINE l_node       om.DomNode
   DEFINE ls_type      STRING
   DEFINE ls_ref_type  STRING
   DEFINE l_n          LIKE type_t.num10 
   DEFINE l_return     STRING 
   DEFINE ls_tmp       STRING
   DEFINE li_mchk      INTEGER
   DEFINE li_mchk2     INTEGER
   
   LET li_mchk = 0
   LET li_mchk2 = 0

   #若該節點沒有子節點則不做任何處理
   IF p_node.getChildCount() > 0 THEN 
      FOR l_n = 1 TO p_node.getChildCount()
         LET l_node = p_node.getChildByIndex(l_n)
         LET ls_type = l_node.getTagName()
    
         #根據不同的元件呼叫不同的製造Function
         CASE ls_type
            WHEN "component"
               LET l_return = l_return,adzp150_create_component(l_node,pi_lv)
            
            WHEN "operator"
               LET l_return = l_return,adzp150_create_operator(l_node,pi_lv)
            
            WHEN "reference"
               LET ls_ref_type = l_node.getAttribute("type")
               IF ls_ref_type = "lang" THEN
                  LET ls_ref_type = adzp150_create_reference(l_node,pi_lv)
               ELSE
                  LET l_return = l_return,adzp150_create_reference(l_node,pi_lv)
               END IF
            
            WHEN "mnemonic"
               LET li_mchk = li_mchk + 1
               LET l_return = l_return,adzp150_create_mnemonic(l_node,pi_lv,li_mchk)
            
            #WHEN "mcode"
            #   LET li_mchk2 = li_mchk2 + 1
            #   LET l_return = l_return,adzp150_create_mcode(l_node,pi_lv,li_mchk2)
            
            WHEN "check"
               LET l_return = l_return,adzp150_create_check(l_node,pi_lv)
            
            #WHEN "chain"
            #   LET l_return = l_return,adzp150_create_chain(l_node,pi_lv)
         
            #180411-00062 ---start---
            WHEN "multilang"
            #  IF g_erp_ver <> "1.0" THEN   #181128-00048 mark
                  #將多語言開窗段加入update_item的add-point中
                  #此功能只有在3.0以上版本才有
                  #因為是直接產生在add-point，而非產生在硬框架，所以不需要有return值
                  CALL adzp150_create_multilang(l_node,pi_lv)
            #  END IF   #181128-00048 mark
            #180411-00062 --- end ---

         END CASE 
         
      END FOR 
      
   END IF 
   
   RETURN l_return
   
END FUNCTION 


#+ 生成controlp元件(common欄位自動產生)
PRIVATE FUNCTION adzp150_create_controlp_common(ps_field,pi_lv)
   DEFINE ps_field        STRING 
   DEFINE pi_lv           LIKE type_t.num10 
   DEFINE ls_tbl          STRING
   DEFINE ls_field        STRING
   DEFINE ls_var          STRING
   DEFINE ls_return       STRING 
   
   #取得table名稱
   IF g_properties.getValue("location") = "head" THEN
      LET ls_tbl = g_properties.getValue("master_tbl_name")
   ELSE
      LET ls_tbl = g_properties.getValue("detail_tbl_name")
   END IF
   
   #取得欄位名稱
   LET ls_field = ps_field 

   #先組合變數名稱
   IF g_properties.getValue("location") = "head" THEN
      LET ls_var = g_properties.getValue("master_var_title"),".",ls_field
   ELSE
      LET ls_var = g_properties.getValue("detail_var_title"),"[l_ac].",ls_field
   END IF
   
   #格式 CALL q_gzxa('table名稱','欄位名稱',TRUE, FALSE, 變數名稱) RETURNING 變數名稱
   LET ls_return = (pi_lv*li_space) SPACES,
                  "CALL q_common('", ls_tbl,"','", ls_field,"',TRUE,FALSE,",ls_var,") RETURNING ", ls_var,"\n",
                  (pi_lv*li_space) SPACES,
                  "DISPLAY BY NAME ", ls_var, "\n",
                  (pi_lv*li_space) SPACES,
                  "NEXT FIELD ", ls_field
                  
   RETURN ls_return
   
END FUNCTION 

#+ 判斷該欄位參照的table名稱
PUBLIC FUNCTION adzp150_define_table_name(ps_field)
   DEFINE ps_field     STRING
   DEFINE ls_field     CHAR(100)
   DEFINE ls_table     CHAR(100)
   DEFINE ls_return    STRING
   DEFINE ls_prefix    CHAR(100)
   DEFINE li_cnt       INTEGER
   
   #取table前四碼
   LET ls_prefix = '%_t'
   LET ls_field  = ps_field
   
   #先比對資料是否重複
   LET ls_table = ""
   SELECT COUNT(1) INTO li_cnt FROM gztz_t 
    WHERE gztz002 = ls_field AND gztz001 LIKE ls_prefix
    
   IF li_cnt > 1 THEN
      DISPLAY "ERROR(10-1):欄位",ls_field CLIPPED, "重複存在於資料庫(gztz_t)中, 請到gztz中檢查該欄位出現多次的原因!"
      EXIT PROGRAM
   END IF 
    
   #改由gztz進行比對
   LET ls_table = ""
   SELECT gztz001 INTO ls_table FROM gztz_t 
    WHERE gztz002 = ls_field AND gztz001 LIKE ls_prefix
   IF SQLCA.SQLCODE THEN
      DISPLAY "ERROE: gztz_t表格紀錄表選取內容時發生錯誤:",SQLCA.SQLCODE,":",SQLERRMESSAGE
   END IF
 
   IF cl_null(ls_table) THEN
      DISPLAY "ERROR(10-2):欄位",ls_field CLIPPED, "不存在於資料庫(gztz_t)中, 請重新檢查!"
   END IF 

   LET ls_return = ls_table CLIPPED, '.'

   RETURN ls_return
   
END FUNCTION


#+ 判斷該欄位參照的table名稱
PUBLIC FUNCTION adzp150_get_table_pre(ps_tbl_name)
   DEFINE ps_tbl_name STRING
   DEFINE ls_tbl_pre  STRING
   DEFINE li_idx      INTEGER
   
   LET li_idx = ps_tbl_name.getIndexOf("_",1) - 1
   
   LET ls_tbl_pre = ps_tbl_name.subString(1,li_idx)
   
   RETURN ls_tbl_pre

END FUNCTION


#+ 判斷該欄位參照的table名稱
PRIVATE FUNCTION adzp150_get_table_name(pn_root)
   DEFINE pn_root     om.DomNode
   DEFINE lnl_data    om.NodeList
   DEFINE ln_node     om.DomNode
   DEFINE ln_hnode    om.DomNode
   DEFINE ln_bnode    om.DomNode
   DEFINE ls_hfile    STRING
   DEFINE ls_bfile    STRING
   DEFINE ls_key_list STRING

   LET lnl_data = pn_root.selectByTagName("dataset")
   LET ln_node = lnl_data.item(1)

   #此處無法使用CASE WHEN,其原因是大多數的樣板兩種都要做,非唯一執行
   #有單頭的樣板
   IF adzp150_type_decide("m") THEN
      #取出單頭table
      LET lnl_data = ln_node.selectByTagName("head")
      LET ln_hnode = lnl_data.item(1)
      IF ln_hnode IS NULL THEN 
         DISPLAY "ERROR(16):",g_properties.getValue("type_id"),"樣板必須設定單頭相關資料, 程式終止!"
         EXIT PROGRAM
      END IF
      LET ls_hfile = ln_hnode.getAttribute("id")
      CALL g_properties.addAttribute("master_tbl_name",ls_hfile)
      LET ls_hfile = adzp150_get_table_pre(ls_hfile)
      CALL g_properties.addAttribute("master_tbl_prefix",ls_hfile)
      LET ls_key_list = adzp150_add_table_name(ln_hnode.getAttribute("pk"),'m')
      CALL g_properties.addAttribute("master_field_allkeys",ls_key_list)
   END IF

   #有單身的樣板
   IF adzp150_type_decide("d") THEN
      #取出單身table
      LET lnl_data = ln_node.selectByTagName("body")
      LET ln_bnode = lnl_data.item(1)
      IF ln_bnode IS NULL THEN
         DISPLAY "ERROR(17):",g_properties.getValue("type_id"),"樣板必須設定單身相關資料, 程式終止!"
         EXIT PROGRAM
      END IF
      LET ls_bfile = ln_bnode.getAttribute("id")
      CALL g_properties.addAttribute("detail_tbl_name",ls_bfile)
      
      LET ls_bfile = adzp150_get_table_pre(ls_bfile)

      CALL g_properties.addAttribute("detail_tbl_prefix",ls_bfile)
      
   END IF

END FUNCTION

#+ 判斷樣板是否需進行此項判斷
PRIVATE FUNCTION adzp150_type_decide(ps_type)
   DEFINE ps_type STRING            
   # m ( master  -   單頭   )
   # d ( detail  -   單身   )
   # t ( tree    -   樹狀   )
   # b ( brwoser - 瀏覽頁籤 )

   CASE ps_type  
      WHEN 'm' #有單頭
         IF g_properties.getValue("type_id") = "c01a" OR 
            g_properties.getValue("type_id") = "c01b" OR 
            #g_properties.getValue("type_id") = "c01c" OR 
            g_properties.getValue("type_id") = "c03a" OR 
            g_properties.getValue("type_id") = "c03b" OR 
            #g_properties.getValue("type_id") = "c03c" OR 
            g_properties.getValue("type_id") = "i01"  OR 
            g_properties.getValue("type_id") = "i07"  OR
            g_properties.getValue("type_id") = "t01"  OR
            g_properties.getValue("type_id") = "i08"  OR
            g_properties.getValue("type_id") = "i05"  OR
            g_properties.getValue("type_id") = "i06"  OR
            g_properties.getValue("type_id") = "i04"  OR
            g_properties.getValue("type_id") = "i11"  OR
            g_properties.getValue("type_id") = "i12"  THEN
            RETURN TRUE
         END IF
         
      WHEN 'd' #有單身
         IF g_properties.getValue("type_id") = "c02b" OR 
            #g_properties.getValue("type_id") = "c02c" OR 
            g_properties.getValue("type_id") = "c03b" OR 
            #g_properties.getValue("type_id") = "c03c" OR 
            g_properties.getValue("type_id") = "i02"  OR   #單檔多欄
            g_properties.getValue("type_id") = "i03"  OR
            g_properties.getValue("type_id") = "i07"  OR
            g_properties.getValue("type_id") = "t01"  OR
            g_properties.getValue("type_id") = "i08"  OR
            g_properties.getValue("type_id") = "t02"  OR
            g_properties.getValue("type_id") = "i04"  OR
            g_properties.getValue("type_id") = "i11"  OR
            g_properties.getValue("type_id") = "i12"  OR
            g_properties.getValue("type_id") = "c04a" THEN
            RETURN TRUE
         END IF
         
      WHEN 't' #有TREE
         IF g_properties.getValue("type_id") = "i03" OR
            g_properties.getValue("type_id") = "i04" OR
            g_properties.getValue("type_id") = "i05" OR
            g_properties.getValue("type_id") = "i08" THEN
            RETURN TRUE
         END IF
         
      WHEN 'b' #有BROWSER
         IF g_properties.getValue("type_id") = "i01" OR
            g_properties.getValue("type_id") = "i07" OR
            g_properties.getValue("type_id") = "t01" THEN
            RETURN TRUE
         END IF    
         
   END CASE

   RETURN FALSE
    
END FUNCTION


#+ 取代掉$var並還原成正確字串
PUBLIC FUNCTION adzp150_replace_var(ps_str)
DEFINE ps_str        STRING
DEFINE ls_str        STRING
DEFINE ls_replace    LIKE type_t.num10
DEFINE ls_apostrophe LIKE type_t.chr1

   LET ps_str = cl_replace_str(ps_str,"$varg_enterprise$var","\"||g_enterprise||\"")

   LET ls_str = ps_str
   #IF ps_str.getIndexOf("g_enterprise",1) > 0 THEN
   #   LET ls_apostrophe = "" 
   #ELSE
      LET ls_apostrophe = "'"
   #END IF
   
   #進行資料比對並取代
   WHILE TRUE
      LET ls_replace= ls_str.getIndexOf("$var",1)
      
      #若有$var則開始取代
      IF ls_replace > 0 THEN
         LET ls_str = ls_str.subString(1,ls_replace-1),ls_apostrophe,"\"||",ls_str.subString(ls_replace+4,ls_str.getLength())
         LET ls_replace= ls_str.getIndexOf("$var",1)
         #$var為成對存在, 若只有單一存在則顯示錯誤訊息
         IF ls_replace = 0 THEN
            DISPLAY "ERROR(13):",ps_str,"定義錯誤, 請檢查該參數內容!"
         ELSE
            #進行重組並取代
            LET ls_str = ls_str.subString(1,ls_replace-1),"||\"",ls_apostrophe,ls_str.subString(ls_replace+4,ls_str.getLength())
         END IF
      ELSE         
         EXIT WHILE
      END IF
   END WHILE
   
   RETURN ls_str

END FUNCTION

#enterprise code/site code各項處理
FUNCTION adzp150_append_field()
   DEFINE l_tmp                 STRING
   DEFINE l_tmp2                STRING
   DEFINE lc_tmp                CHAR(500)
   DEFINE l_master_table        STRING
   DEFINE l_master_field_ent    STRING
   DEFINE l_master_field_site   STRING
   DEFINE l_detail_table        STRING
   DEFINE l_detail_field_ent    STRING
   DEFINE l_detail_field_site   STRING
   DEFINE l_master_ent          BOOLEAN
   DEFINE l_detail_ent          BOOLEAN
   DEFINE l_master_site         BOOLEAN
   DEFINE l_detail_site         BOOLEAN
   DEFINE l_master_wc           STRING
   DEFINE l_detail_wc           STRING
   DEFINE l_master_cl           STRING
   DEFINE l_detail_cl           STRING
   DEFINE l_master_define       STRING
   DEFINE l_detail_define       STRING
   DEFINE l_master_wc_s         STRING
   DEFINE l_master_wc_s_t0      STRING
   DEFINE l_master_wc_s_refresh STRING
   DEFINE l_detail_wc_s         STRING
   DEFINE l_detail_wc_s_t0      STRING
   DEFINE l_detail_wc_s_refresh STRING
   DEFINE l_master_var          STRING
   DEFINE l_detail_var          STRING
   DEFINE l_master_field        STRING
   DEFINE l_detail_field        STRING
   DEFINE l_chk_site            BOOLEAN
   DEFINE l_chk_site_pk         STRING
   DEFINE lc_chk_site_pk        CHAR(500)

   #判斷單頭table是否有enterprise
   LET l_tmp = g_properties.getValue("master_tbl_name")
   LET l_master_table = l_tmp.toLowerCase() 
   LET l_tmp = g_properties.getValue("title_master"),"ent"
   LET l_master_field_ent = l_tmp.toLowerCase() 
   IF cl_getField(l_master_table CLIPPED, l_master_field_ent CLIPPED) THEN
      CALL g_properties.addAttribute("master_field_ent",l_master_field_ent)
      LET l_master_ent = TRUE
   ELSE
      LET l_master_ent = FALSE
   END IF
   
   #判斷單身table是否有enterprise
   LET l_tmp = g_properties.getValue("detail_tbl_name")
   LET l_detail_table = l_tmp.toLowerCase() 
   LET l_tmp = g_properties.getValue("title_detail"),"ent"
   LET l_detail_field_ent = l_tmp.toLowerCase() 
   IF cl_getField(l_detail_table CLIPPED, l_detail_field_ent CLIPPED) THEN
      CALL g_properties.addAttribute("detail_field_ent",l_detail_field_ent)
      LET l_detail_ent = TRUE
   ELSE
      LET l_detail_ent = FALSE
   END IF
   
   #判斷單頭table是否有site
   LET l_tmp = g_properties.getValue("master_tbl_name")
   LET l_master_table = l_tmp.toLowerCase() 
   LET l_tmp = g_properties.getValue("title_master"),"site"
   LET l_master_field_site = l_tmp.toLowerCase() 
   LET l_tmp  = g_properties.getValue("master_field_allkeys")
   LET l_tmp2 = g_properties.getValue("general_all_field")
    
   #判定單頭table是否有site
   IF l_tmp.getIndexOf(l_master_field_site,1) = 0 THEN
      IF l_tmp2.getIndexOf(l_master_field_site,1) = 0 THEN
         CALL g_properties.addAttribute("master_field_site",l_master_field_site)
         LET l_chk_site = TRUE
      ELSE
         LET l_chk_site = FALSE
         DISPLAY 'Site欄位出現於畫面上, 不做特別處理!'
      END IF
   ELSE
      LET l_chk_site = FALSE
   END IF
   
   #判定單頭site是否為key
   LET lc_tmp = g_properties.getValue("master_tbl_name")
   SELECT dzed004 INTO lc_chk_site_pk FROM dzed_t
    WHERE dzed001 = lc_tmp
      AND dzed003 = 'P'
   
   LET l_chk_site_pk = lc_chk_site_pk
   IF cl_getField(l_master_table CLIPPED, l_master_field_site CLIPPED ) AND 
      l_chk_site AND
      l_chk_site_pk.getIndexOf(l_master_field_site,1) > 0 THEN
      LET l_master_site = TRUE
   ELSE
      LET l_master_site = FALSE
   END IF
   
   #判斷單身table是否有site
   LET l_tmp = g_properties.getValue("detail_tbl_name")
   LET l_detail_table = l_tmp.toLowerCase() 
   LET l_tmp = g_properties.getValue("title_detail"),"site"
   LET l_detail_field_site = l_tmp.toLowerCase() 
    
   #判定單身site是否為key
   LET lc_tmp = g_properties.getValue("detail_tbl_name")
   SELECT dzed004 INTO lc_chk_site_pk FROM dzed_t
    WHERE dzed001 = lc_tmp
      AND dzed003 = 'P'
   
   LET l_chk_site_pk = lc_chk_site_pk
   IF cl_getField(l_detail_table CLIPPED, l_detail_field_site CLIPPED) AND 
      l_chk_site_pk.getIndexOf(l_detail_field_site,1) > 0 AND
      l_chk_site THEN
      #CALL g_properties.addAttribute("detail_field_site",l_detail_field_site)
      LET l_detail_site = TRUE
   ELSE
      LET l_detail_site = FALSE
   END IF

   #若單頭有enterprise欄位, 進行所需資料組合
   IF l_master_ent THEN
      LET l_master_wc_s          = l_master_wc_s   , l_master_field_ent CLIPPED, " = \" ||g_enterprise|| \" AND "
      LET l_master_wc_s_t0       = l_master_wc_s_t0, 't0.', l_master_field_ent CLIPPED, " = \" ||g_enterprise|| \" AND "
      LET l_master_wc_s_refresh  = l_master_wc_s_refresh, 't0.', l_master_field_ent CLIPPED, " = \" ||g_enterprise|| \" AND "
      LET l_master_wc            = l_master_wc     , l_master_field_ent CLIPPED, " = g_enterprise AND "
      LET l_master_field         = l_master_field  , l_master_field_ent CLIPPED, ", "
      LET l_master_var           = l_master_var    , "g_enterprise, "
      LET l_master_define        = 
          l_master_define, "LET ",g_properties.getValue("master_var_title"),".",l_master_field_ent, " = g_enterprise \n" 
      LET l_master_cl = l_master_cl, l_master_field_ent, " = ? AND "
   END IF
   
   #若單身有enterprise欄位, 進行所需資料組合
   IF l_detail_ent THEN
      LET l_detail_wc_s  = l_detail_wc_s  , l_detail_field_ent CLIPPED, " = \" ||g_enterprise|| \" AND "
      LET l_detail_wc_s_t0  = l_detail_wc_s_t0, 't0.', l_detail_field_ent CLIPPED, " = \" ||g_enterprise|| \" AND "
      LET l_detail_wc    = l_detail_wc    , l_detail_field_ent CLIPPED, " = g_enterprise AND "
      LET l_detail_field = l_detail_field , l_detail_field_ent CLIPPED, ", "
      LET l_detail_var   = l_detail_var   , "g_enterprise, "
      LET l_detail_define = 
          l_detail_define, "LET ",g_properties.getValue("detail_var_title"),".",l_detail_field_ent, " = g_enterprise \n" 
      LET l_detail_cl = l_detail_cl, l_detail_field_ent, " = ? AND "
   END IF
      
   #若單頭有site欄位, 進行所需資料組合
   IF l_master_site THEN
      CALL g_properties.addAttribute("master_using_site",l_master_field_site)
      LET l_master_wc_s          = l_master_wc_s  , l_master_field_site CLIPPED, " = '\" ||g_site|| \"' AND "
      LET l_master_wc_s_t0       = l_master_wc_s_t0, 't0.', l_master_field_site CLIPPED, " = '\" ||g_site|| \"' AND "
      LET l_master_wc_s_refresh  = l_master_wc_s_refresh, 't0.', l_master_field_site CLIPPED, " = ? AND "
      LET l_master_wc            = l_master_wc    , l_master_field_site CLIPPED, " = g_site AND "
      LET l_master_field         = l_master_field , l_master_field_site CLIPPED, ", "
      LET l_master_var           = l_master_var   , "g_site, "
      LET l_master_define = 
          l_master_define, "LET ",g_properties.getValue("master_var_title"),".",l_master_field_site, " = g_site \n" 
      LET l_master_cl = l_master_cl, l_master_field_site, " = ? AND "
   END IF        
     
   #若單身有site欄位, 進行所需資料組合
   IF l_detail_site THEN
      LET l_detail_wc_s  = l_detail_wc_s  , l_detail_field_site CLIPPED, " = '\" ||g_site|| \"' AND "
      LET l_detail_wc_s_t0  = l_detail_wc_s_t0, 't0.', l_detail_field_site CLIPPED, " = '\" ||g_site|| \"' AND "
      LET l_detail_wc    = l_detail_wc    , l_detail_field_site CLIPPED, " = g_site AND "
      LET l_detail_field = l_detail_field , l_detail_field_site CLIPPED, ", "
      LET l_detail_var   = l_detail_var   , "g_site, "
      LET l_detail_define = 
          l_detail_define, "LET ",g_properties.getValue("detail_var_title"),".",l_detail_field_site, " = g_site \n" 
      LET l_detail_cl = l_detail_cl, l_detail_field_site, " = ? AND "
   END IF

   #單頭段
   CALL g_properties.addAttribute("master_append_wc_s"         , l_master_wc_s        )
   CALL g_properties.addAttribute("master_append_wc_s_t0"      , l_master_wc_s_t0     )
   CALL g_properties.addAttribute("master_append_wc_s_refresh" , l_master_wc_s_refresh)
   CALL g_properties.addAttribute("master_append_wc"           , l_master_wc          )
   CALL g_properties.addAttribute("master_field_append"        , l_master_field       )
   CALL g_properties.addAttribute("master_var_append"          , l_master_var         )
   #CALL g_properties.addAttribute("master_define_append", l_master_define)
                           
   #單身段
   CALL g_properties.addAttribute("detail_append_wc_s"  , l_detail_wc_s  )
   CALL g_properties.addAttribute("detail_append_wc"    , l_detail_wc    )
   CALL g_properties.addAttribute("detail_field_append" , l_detail_field )
   CALL g_properties.addAttribute("detail_var_append"   , l_detail_var   )
   #CALL g_properties.addAttribute("detail_define_append", l_detail_define)
    
   #暫時特別處理 for ent
   #IF NOT cl_null(l_master_cl) THEN
   #   LET l_tmp = g_properties.getValue("master_sql_forupd")
   #   LET l_tmp = l_tmp.subString(1,l_tmp.getIndexOf('WHERE',1)+4),
   #               " ",l_master_cl,
   #               l_tmp.subString(l_tmp.getIndexOf('WHERE',1)+5,l_tmp.getLength())
   #   CALL g_properties.addAttribute("master_sql_forupd",l_tmp)
   #END IF
   #
   #IF NOT cl_null(l_detail_cl) THEN
   #   LET l_tmp = g_properties.getValue("detail_sql_forupd")
   #   LET l_tmp = l_tmp.subString(1,l_tmp.getIndexOf('WHERE',1)+4),
   #               " ",l_detail_cl,
   #               l_tmp.subString(l_tmp.getIndexOf('WHERE',1)+5,l_tmp.getLength())
   #   CALL g_properties.addAttribute("detail_sql_forupd",l_tmp)
   #END IF
   
END FUNCTION       

#必要參數檢查是否存在
FUNCTION adzp150_parameter_chk()
   DEFINE li_idx  INTEGER
   DEFINE ls_tmp  STRING
   DEFINE lr_parameter DYNAMIC ARRAY OF RECORD
          name STRING,
          desc STRING
          END RECORD
   
   #無論何種皆必要的參數
   #模組別
   LET lr_parameter[lr_parameter.getLength()+1].name = "general_module"
   LET lr_parameter[lr_parameter.getLength()].desc = "模組別"
   #執行模式
   LET lr_parameter[lr_parameter.getLength()+1].name = "general_jobmode"
   LET lr_parameter[lr_parameter.getLength()].desc = "執行模式"
   #樣板類型
   LET lr_parameter[lr_parameter.getLength()+1].name = "type_id"
   LET lr_parameter[lr_parameter.getLength()].desc = "樣板類型"
   #行業別
   LET lr_parameter[lr_parameter.getLength()+1].name = "industry_id"
   LET lr_parameter[lr_parameter.getLength()].desc = "行業別"  
   #程式版本
   LET lr_parameter[lr_parameter.getLength()+1].name = "general_adp_ver"
   LET lr_parameter[lr_parameter.getLength()].desc = "程式版本"

      
   #有單頭的程式必要參數
   IF adzp150_type_decide("m") THEN
      #單頭pk欄位
      LET lr_parameter[lr_parameter.getLength()+1].name = "master_field_pks"
      LET lr_parameter[lr_parameter.getLength()].desc = "單頭pk欄位"
      #單頭table名稱
      LET lr_parameter[lr_parameter.getLength()+1].name = "master_tbl_name"
      LET lr_parameter[lr_parameter.getLength()].desc = "單頭table名稱"
      #單頭所有欄位定義
      LET lr_parameter[lr_parameter.getLength()+1].name = "master_fields_all"
      LET lr_parameter[lr_parameter.getLength()].desc = "單頭所有欄位定義"
      #單頭可輸入欄位定義
      LET lr_parameter[lr_parameter.getLength()+1].name = "master_input_fields"
      LET lr_parameter[lr_parameter.getLength()].desc = "單頭可輸入欄位定義"
      #單頭可查詢欄位定義
      LET lr_parameter[lr_parameter.getLength()+1].name = "master_fields_qbe"
      LET lr_parameter[lr_parameter.getLength()].desc = "單頭可查詢欄位定義"
      #單頭for update SQL定義
      LET lr_parameter[lr_parameter.getLength()+1].name = "master_sql_forupd"
      LET lr_parameter[lr_parameter.getLength()].desc = "單頭for update SQL定義"
   END IF
 
   #有單身的程式必要參數
   IF adzp150_type_decide("d") THEN
      #樹狀不一定有這些欄位
      IF NOT adzp150_type_decide("t") THEN
         #單身primary key欄位
         LET lr_parameter[lr_parameter.getLength()+1].name = "detail_field_pks"
         LET lr_parameter[lr_parameter.getLength()].desc = "單身pk欄位"
         #單檔多欄不需要fk欄位
         IF NOT g_properties.getValue("type_id") = "i02" THEN
            #單身foreign key欄位
            LET lr_parameter[lr_parameter.getLength()+1].name = "detail_field_fks"
            LET lr_parameter[lr_parameter.getLength()].desc = "單身fk欄位"
         END IF
      END IF
      #page數量
      LET lr_parameter[lr_parameter.getLength()+1].name = "page"
      LET lr_parameter[lr_parameter.getLength()].desc = "頁籤數量"
      #單身table名稱
      LET lr_parameter[lr_parameter.getLength()+1].name = "detail_tbl_name"
      LET lr_parameter[lr_parameter.getLength()].desc = "單身table名稱"
      #單身所有欄位定義
      LET lr_parameter[lr_parameter.getLength()+1].name = "detail_fields_all"
      LET lr_parameter[lr_parameter.getLength()].desc = "單身所有欄位定義"
      #單身可輸入欄位定義
      LET lr_parameter[lr_parameter.getLength()+1].name = "detail_input_fields"
      LET lr_parameter[lr_parameter.getLength()].desc = "單身可輸入欄位定義"
      #單身for update SQL定義
      LET lr_parameter[lr_parameter.getLength()+1].name = "detail_sql_forupd"
      LET lr_parameter[lr_parameter.getLength()].desc = "單身for update SQL定義"
      #單身for update SQL定義
      #LET lr_parameter[lr_parameter.getLength()+1].name = "detail_fill_sql"
      #LET lr_parameter[lr_parameter.getLength()].desc = "單身填充SQL定義"
      #單身qbe欄位定義
      LET lr_parameter[lr_parameter.getLength()+1].name = "detail_fields_qbe1"
      LET lr_parameter[lr_parameter.getLength()].desc = "單身qbe欄位定義"
   END IF

   #有樹狀的程式必要參數
   IF adzp150_type_decide("t") THEN
   
      #Tree單檔多欄沒有browser定義(單身=Browser)
      IF g_properties.getValue("type_id") <> "i03" THEN
         #TREE所有欄位定義
         LET lr_parameter[lr_parameter.getLength()+1].name = "browser_fields_define"
         LET lr_parameter[lr_parameter.getLength()].desc = "TREE所有欄位定義"
      END IF
      
      #LET lr_parameter[lr_parameter.getLength()+1].name = "browser_field_desc"
      #LET lr_parameter[lr_parameter.getLength()].desc = "瀏覽頁面描述欄位"
      
      #3種樹狀架構略有不同, 需額外進行判斷
      CASE g_properties.getValue("type_id")
         WHEN "i03"
            #子節點欄位
            LET lr_parameter[lr_parameter.getLength()+1].name = "detail_field_lid"
            LET lr_parameter[lr_parameter.getLength()].desc = "子節點欄位"
            #父節點欄位
            LET lr_parameter[lr_parameter.getLength()+1].name = "detail_field_pid"
            LET lr_parameter[lr_parameter.getLength()].desc = "父節點欄位"
         WHEN "i05" OR "i08"
            #子節點欄位
            LET lr_parameter[lr_parameter.getLength()+1].name = "master_field_lid"
            LET lr_parameter[lr_parameter.getLength()].desc = "子節點欄位"
            #父節點欄位
            LET lr_parameter[lr_parameter.getLength()+1].name = "master_field_pid"
            LET lr_parameter[lr_parameter.getLength()].desc = "父節點欄位"
         WHEN "i04"    
            #子節點欄位(單頭)
            LET lr_parameter[lr_parameter.getLength()+1].name = "master_field_lid"
            LET lr_parameter[lr_parameter.getLength()].desc = "單頭子節點欄位"
            #子節點欄位(單身)
            LET lr_parameter[lr_parameter.getLength()+1].name = "detail_field_lid"
            LET lr_parameter[lr_parameter.getLength()].desc = "單身子節點欄位"
            #父節點欄位(單身)
            LET lr_parameter[lr_parameter.getLength()+1].name = "detail_field_pid"
            LET lr_parameter[lr_parameter.getLength()].desc = "單身父節點欄位"
      END CASE
      
   END IF

   #有BROWSER的程式必要參數
   IF adzp150_type_decide("b") THEN
      #BROWSER所有欄位定義
      LET lr_parameter[lr_parameter.getLength()+1].name = "browser_fields_define"
      LET lr_parameter[lr_parameter.getLength()].desc = "BROWSER所有欄位定義"
   END IF
   
   #判斷必要屬性是否存在, 若不存在則顯示錯誤訊息
   #FOR li_idx = 1 TO lr_parameter.getLength()
   #   LET ls_tmp = g_properties.getValue(lr_parameter[li_idx].name)
   #   IF cl_null(ls_tmp) THEN
   #      DISPLAY "WARNING(1): tab資料遺漏, 請檢查 '",lr_parameter[li_idx].desc,"'"
   #   END IF
   #END FOR
   
END FUNCTION 

#特定樣板進行註解特定段落處理
FUNCTION adzp150_mark()
   DEFINE l_mdl    CHAR(10)
   DEFINE l_tmp    STRING
   DEFINE li_idx   INTEGER

   LET l_mdl = g_properties.getValue("type_id_t")

   #暫時全加
   CALL g_properties.addAttribute("master_refresh_unique","DISTINCT")
   
   CASE l_mdl
      #C開頭為子程式樣板
      WHEN "c01a"  
         CALL g_properties.addAttribute("general_c01a_mark","#該樣板不需此段落")
         CALL g_properties.addAttribute("general_cnt_num","1")
         CALL g_properties.addAttribute("general_dialog_menu","MENU")
         LET l_tmp = "b_statepic     LIKE type_t.chr50,"
         CALL g_properties.addAttribute("general_browser_define",l_tmp)

      WHEN "c01b"
      
      WHEN "c01c"
      
      WHEN "c02a"
         #若有ent欄位
         IF NOT cl_null(g_properties.getValue("detail_field_ent")) AND
            NOT cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs USING g_enterprise, g_site"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
            #181211-00007#1
            LET l_tmp = "USING g_enterprise, g_site"
            CALL g_properties.addAttribute("detail_foreach_using_ent",l_tmp )
            #181211-00007#1
         END IF
         IF NOT cl_null(g_properties.getValue("detail_field_ent")) AND
                cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs USING g_enterprise"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
            #181211-00007#1
            LET l_tmp = "USING g_enterprise"
            CALL g_properties.addAttribute("detail_foreach_using_ent",l_tmp )
            #181211-00007#1
         END IF
         IF     cl_null(g_properties.getValue("detail_field_ent")) AND
            NOT cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs USING g_site"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
            #181211-00007#1
            LET l_tmp = "USING g_site"
            CALL g_properties.addAttribute("detail_foreach_using_ent",l_tmp )
            #181211-00007#1
         END IF
         IF     cl_null(g_properties.getValue("detail_field_ent")) AND
                cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
         END IF
         CALL g_properties.addAttribute("general_c02a_mark","#該樣板不需此段落")
         
      WHEN "c02b"
      
      WHEN "c02c"
      
      WHEN "c03a"
         CALL g_properties.addAttribute("general_i09_mark","#該樣板不需此段落")
         CALL g_properties.addAttribute("general_c03a_mark","#該樣板不需此段落")
         CALL g_properties.addAttribute("general_cnt_num","1")
         
      WHEN "c03b"
      
      WHEN "c03c"
    
      WHEN "c04a"      
         #若有ent欄位
         IF NOT cl_null(g_properties.getValue("detail_field_ent")) AND
            NOT cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs USING g_enterprise, g_site"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
            #181211-00007#1
            LET l_tmp = "USING g_enterprise, g_site"
            CALL g_properties.addAttribute("detail_foreach_using_ent",l_tmp )
            #181211-00007#1
         END IF
         IF NOT cl_null(g_properties.getValue("detail_field_ent")) AND
                cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs USING g_enterprise"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
            #181211-00007#1
            LET l_tmp = "USING g_enterprise"
            CALL g_properties.addAttribute("detail_foreach_using_ent",l_tmp )
            #181211-00007#1
         END IF
         IF     cl_null(g_properties.getValue("detail_field_ent")) AND
            NOT cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs USING g_site"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
            #181211-00007#1
            LET l_tmp = "USING g_site"
            CALL g_properties.addAttribute("detail_foreach_using_ent",l_tmp )
            #181211-00007#1
         END IF
         IF     cl_null(g_properties.getValue("detail_field_ent")) AND
                cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
         END IF
      
      WHEN "i01"
         CALL g_properties.addAttribute("general_cnt_num","g_max_browse")
         CALL g_properties.addAttribute("general_i01_mark","#該樣板不需此段落")
         CALL g_properties.addAttribute("general_dialog_menu","DIALOG")
         LET l_tmp = "b_statepic     LIKE type_t.chr50,"
         CALL g_properties.addAttribute("general_browser_define",l_tmp)
         
      WHEN "i02"
         #若有ent欄位
         IF NOT cl_null(g_properties.getValue("detail_field_ent")) AND
            NOT cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs USING g_enterprise, g_site"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
            #181211-00007#1
            LET l_tmp = "USING g_enterprise, g_site"
            CALL g_properties.addAttribute("detail_foreach_using_ent",l_tmp )
            #181211-00007#1
         END IF
         IF NOT cl_null(g_properties.getValue("detail_field_ent")) AND
                cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs USING g_enterprise"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
            #181211-00007#1
            LET l_tmp = "USING g_enterprise"
            CALL g_properties.addAttribute("detail_foreach_using_ent",l_tmp )
            #181211-00007#1
         END IF
         IF     cl_null(g_properties.getValue("detail_field_ent")) AND
            NOT cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs USING g_site"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
            #181211-00007#1
            LET l_tmp = "USING g_site"
            CALL g_properties.addAttribute("detail_foreach_using_ent",l_tmp )
            #181211-00007#1
         END IF
         IF     cl_null(g_properties.getValue("detail_field_ent")) AND
                cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
         END IF
         #170515-00054 ---start---
         LET l_tmp = "OR g_action_choice = 'insert' "
         LET l_tmp = "\#170515-00054 新增g_action_choice = 'insert'，\n",
                     "      \#為了怕使用者先按了新增卻直接去修改其他筆資料，因此將insert時也加入權限控管\n",
                     "      OR g_action_choice = 'insert'"
         CALL g_properties.addAttribute("detail_action_choice_insert",l_tmp )
         #170515-00054 --- end ---
      WHEN "i03"
         #若無root欄位
         IF cl_null(g_properties.getValue("detail_field_type")) THEN
            CALL g_properties.addAttribute("detail_type_mark","#該樣板不需此段落: " )
         END IF
      
      WHEN "i04"
         LET l_tmp = ""
         
         #若無pid欄位
         IF cl_null(g_properties.getValue("detail_field_pid")) THEN
            CALL g_properties.addAttribute("detail_pid_mark","#該樣板不需此段落: " )
         ELSE
            LET l_tmp = l_tmp, ",ps_pid "
         END IF
         
         #若無root欄位
         IF cl_null(g_properties.getValue("detail_field_type")) THEN
            CALL g_properties.addAttribute("detail_type_mark","#該樣板不需此段落: " )
         ELSE        
            LET l_tmp = l_tmp, ",ps_type "
            CALL g_properties.addAttribute("detail_func_type",",ps_type ")
         END IF
         
         #若無root欄位
         IF cl_null(g_properties.getValue("master_field_type")) THEN
            CALL g_properties.addAttribute("master_type_mark","#該樣板不需此段落: " )
         ELSE        
            LET l_tmp = l_tmp, ",ps_type "
            CALL g_properties.addAttribute("master_func_type",",ps_type ")
         END IF
         
         #特別用途
         LET l_tmp = g_properties.getValue("browser_fields_all")
         LET l_tmp = cl_replace_str(l_tmp,g_properties.getValue("master_field_lid"),
                                          g_properties.getValue("detail_field_lid"))
         CALL g_properties.addAttribute("browser_fields_all_detail",l_tmp) 
         
      WHEN "i05"
         LET l_tmp = ""
         
         #若無pid欄位
         IF cl_null(g_properties.getValue("master_field_pid")) THEN
            CALL g_properties.addAttribute("master_pid_mark","#該樣板不需此段落: " )
         ELSE
            LET l_tmp = l_tmp, ",ps_pid "
         END IF
         
         #若無root欄位
         IF cl_null(g_properties.getValue("master_field_type")) THEN
            CALL g_properties.addAttribute("master_type_mark","#該樣板不需此段落: " )
         ELSE        
            LET l_tmp = l_tmp, ",ps_type "
            CALL g_properties.addAttribute("master_func_type",",ps_type ")
         END IF
         
         #FUNCTION參數定義
         CALL g_properties.addAttribute("master_func_pid_and_type",l_tmp)
         
      WHEN "i06"
         
      WHEN "i07"
         CALL g_properties.addAttribute("general_cnt_num","g_max_browse")
         CALL g_properties.addAttribute("general_i07_mark","#該樣板不需此段落")
         CALL g_properties.addAttribute("general_dialog_menu","DIALOG")
         CALL g_properties.addAttribute("master_refresh_unique","DISTINCT")
      
      WHEN "i08" 
         #若無pid欄位
         IF cl_null(g_properties.getValue("master_field_pid")) THEN
            CALL g_properties.addAttribute("master_pid_mark","#該樣板不需此段落: " )
         ELSE
            LET l_tmp = l_tmp, ",ps_pid "
         END IF
         
         #若無root欄位
         IF cl_null(g_properties.getValue("master_field_type")) THEN
            CALL g_properties.addAttribute("master_type_mark","#該樣板不需此段落: " )
         ELSE        
            LET l_tmp = l_tmp, ",ps_type "
            CALL g_properties.addAttribute("master_func_type",",ps_type ")
         END IF
    
      WHEN "i09"
         CALL g_properties.addAttribute("general_i09_mark","#該樣板不需此段落")
         CALL g_properties.addAttribute("general_cnt_num","1")
         
      WHEN "i10"      
         CALL g_properties.addAttribute("general_i10_mark","#該樣板不需此段落")
         CALL g_properties.addAttribute("general_cnt_num","1")
         #CALL g_properties.addAttribute("general_action_call",g_properties.getValue("general_action_call_in_menu"))
         CALL g_properties.addAttribute("general_dialog_menu","MENU")
         LET l_tmp = "b_statepic     LIKE type_t.chr50,"
         CALL g_properties.addAttribute("general_browser_define",l_tmp)
         
      WHEN "i11"
      
      WHEN "i12"
         CALL g_properties.addAttribute("general_i12_mark","#該樣板不需此段落")
         CALL g_properties.addAttribute("general_cnt_num","1")
         CALL g_properties.addAttribute("master_refresh_unique","DISTINCT")
      WHEN "i13"
         LET l_tmp = "
         #外顯欄位\n       b_show          LIKE type_t.chr100,\n
         #父節點id\n       b_pid           LIKE type_t.chr100,\n
         #本身節點id\n       b_id            LIKE type_t.chr100,\n
         #是否展開\n       b_exp           LIKE type_t.chr100,\n
         #是否有子節點\n       b_hasC          LIKE type_t.num5,\n
         #是否已展開\n       b_isExp         LIKE type_t.num5,\n
         #展開值\n       b_expcode       LIKE type_t.num5,\n
         #tree自定義欄位" 
         CALL g_properties.addAttribute("general_browser_define",l_tmp)
         CALL g_properties.addAttribute("general_i13_mark","#該樣板不需此段落")
        
      WHEN "t01"
         CALL g_properties.addAttribute("general_t01_mark","#該樣板不需此段落")
         CALL g_properties.addAttribute("general_cnt_num","g_max_browse")
      
      WHEN "t02"      
         #若有ent欄位
         IF NOT cl_null(g_properties.getValue("detail_field_ent")) AND
            NOT cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs USING g_enterprise, g_site"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
            #181211-00007#1
            LET l_tmp = "USING g_enterprise, g_site"
            CALL g_properties.addAttribute("detail_foreach_using_ent",l_tmp )
            #181211-00007#1
         END IF
         IF NOT cl_null(g_properties.getValue("detail_field_ent")) AND
                cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs USING g_enterprise"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
            #181211-00007#1
            LET l_tmp = "USING g_enterprise"
            CALL g_properties.addAttribute("detail_foreach_using_ent",l_tmp )
            #181211-00007#1
         END IF
         IF     cl_null(g_properties.getValue("detail_field_ent")) AND
            NOT cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs USING g_site"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
            #181211-00007#1
            LET l_tmp = "USING g_site"
            CALL g_properties.addAttribute("detail_foreach_using_ent",l_tmp )
            #181211-00007#1
         END IF
         IF     cl_null(g_properties.getValue("detail_field_ent")) AND
                cl_null(g_properties.getValue("detail_field_site")) THEN
            LET l_tmp = "OPEN b_fill_curs"
            CALL g_properties.addAttribute("detail_using_ent",l_tmp )
         END IF
         #211006-00002 ---start---
         LET l_tmp = "OR g_action_choice = 'insert' "
         LET l_tmp = "\#211006-00002 新增g_action_choice = 'insert'，\n",
                     "      \#為了怕使用者先按了新增卻直接去修改其他筆資料，因此將insert時也加入權限控管\n",
                     "      OR g_action_choice = 'insert'"
         CALL g_properties.addAttribute("detail_action_choice_insert",l_tmp )
         #211006-00002 --- end --- 
      OTHERWISE
      
   END CASE
   
END FUNCTION

#+ Tree架構desc段落處理
FUNCTION adzp150_create_desc(ps_list,ps_type)
   DEFINE ps_list      STRING
   DEFINE ps_type      STRING
   DEFINE ls_name      STRING
   DEFINE lst_token    base.StringTokenizer
   DEFINE ls_addpoint  STRING
   DEFINE ls_tmp1      STRING
   DEFINE ls_tmp2      STRING
   DEFINE ls_token     STRING
   
   CALL g_properties.addAttribute("browser_field_desc",ps_list)
   
   #判斷所用之變數名稱
   IF ps_type = 'm' THEN
      IF cl_null(ps_list) THEN
         LET ps_list = g_properties.getValue("master_field_pk01")
      END IF
      LET lst_token = base.StringTokenizer.create(ps_list, ',')
      
      IF g_properties.getValue("type_id_t") = "i13" THEN
         CALL g_properties.addAttribute("master_field_lid",g_properties.getValue("master_field_pk01"))
      END IF
      
      #分析欄位清單並處理
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         #LET ls_tmp1 = ls_tmp1,"      LET g_browser[l_ac].b_",ls_token," = g_browser[l_ac].b_",ls_token," CLIPPED\n"
         LET ls_tmp2 = ls_tmp2,"           g_browser[l_ac].b_",ls_token,","
         IF lst_token.hasMoreTokens() THEN
            LET ls_tmp2 = ls_tmp2, "'-',\n" 
         END IF
      END WHILE
      
      #進行組合
      LET ls_addpoint = ls_tmp1,
                        "      LET g_browser[l_ac].b_show = \n",
                        ls_tmp2,"\n",
                        "             '(',g_browser[l_ac].b_",g_properties.getValue("master_field_lid"),",')'\n"
                        
      LET g_dzbb.prog_name   = g_properties.getValue("app_id")
      LET g_dzbb.point_name  = "desc_show.show"
      LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
      LET g_dzbb.addpoint    = ls_addpoint
      CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))     
      
   ELSE
      IF cl_null(ps_list) THEN
         LET ps_list = g_properties.getValue("master_field_pk01")
      END IF
      LET lst_token = base.StringTokenizer.create(ps_list, ',')
      
      #分析欄位清單並處理
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         #LET ls_tmp1 = ls_tmp1,"      LET g_browser[l_ac].b_",ls_token," = g_browser[l_ac].b_",ls_token," CLIPPED\n"
         LET ls_tmp2 = ls_tmp2,"           g_browser[l_ac].b_",ls_token,","
         IF lst_token.hasMoreTokens() THEN
            LET ls_tmp2 = ls_tmp2, "'-',\n" 
         END IF
      END WHILE
      
      #進行組合
      LET ls_addpoint = ls_tmp1,
                        "      LET g_browser[l_ac].b_show = \n",
                        ls_tmp2,"\n",
                        "             '(',g_browser[l_ac].b_",g_properties.getValue("detail_field_lid"),",')'\n"
                        
      LET g_dzbb.prog_name   = g_properties.getValue("app_id")
      LET g_dzbb.point_name  = "desc_show.show"
      LET g_dzbb.point_ver   = g_properties.getValue("general_adp_ver")
      LET g_dzbb.addpoint    = ls_addpoint
      CALL cl_add_point_insert(g_dzbb.prog_name,g_dzbb.point_name,g_dzbb.point_ver,g_dzbb.addpoint,g_update,ARG_VAL(5))     
      
   END IF

END FUNCTION

#+ 凍結欄位處理
FUNCTION adzp150_freeze_field(ps_page,ps_fields)
   DEFINE ps_page          STRING
   DEFINE ps_fields        STRING
   DEFINE ls_var_title     STRING
   DEFINE lst_token        base.StringTokenizer
   DEFINE ls_var           STRING
   DEFINE ls_token         STRING
   DEFINE ls_freeze        STRING
   DEFINE ls_freeze_pre    STRING
   DEFINE ls_freeze_post   STRING
   
   IF ps_page = '1' THEN
      LET ls_var_title = g_properties.getValue("detail_var_title")
   ELSE
      LET ls_var_title = g_properties.getValue("detail_var_title"||ps_page)
   END IF
   
   LET ls_freeze_pre = "   DISPLAY "
   
   LET lst_token = base.StringTokenizer.create(ps_fields, ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_var = ls_var_title,"[l_ac].",ls_token
      LET ls_freeze_pre = ls_freeze_pre, ls_var
      LET ls_freeze_post = ls_freeze_post, "f_",ls_token
      IF lst_token.hasMoreTokens() THEN
         LET ls_freeze_pre = ls_freeze_pre, ", " 
         LET ls_freeze_post = ls_freeze_post, ", " 
      END IF 
   END WHILE
   
   LET ls_freeze = ls_freeze_pre, " TO ", ls_freeze_post
   
   CALL g_properties.addAttribute("detail_freeze_show",ls_freeze)
   
END FUNCTION

#+ 產生虛擬的bs info
FUNCTION adzp150_create_bs_info(ps_key_fields)
   DEFINE ps_key_fields   STRING
   DEFINE lst_token       base.StringTokenizer
   DEFINE ls_token        STRING
   DEFINE ls_tmp          STRING
   DEFINE ls_bsvars       STRING
   DEFINE ls_key_fields   STRING
   
   CALL g_properties.addAttribute("bs_order", ps_key_fields )

   #去除特別欄位參照內容
   LET lst_token = base.StringTokenizer.create(ps_key_fields, ',')
   LET ls_tmp = '' 
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      IF ls_token.getIndexOf('(',1) > 0 OR 
         ls_token.getIndexOf(g_properties.getValue("master_tbl_prefix"),1) = 0 THEN
         LET ls_tmp = ls_tmp, "'',"
      ELSE
         LET ls_tmp = ls_tmp, ls_token, ','
      END IF
   END WHILE
   LET ls_tmp = ls_tmp.substring(1,ls_tmp.getLength()-1)
   CALL g_properties.addAttribute("browser_fields_all_tree", ls_tmp )
   LET ls_tmp = adzp150_add_table_name(ls_tmp,'b')
   CALL g_properties.addAttribute("browser_fields_all", ls_tmp )
   
   #組合所有browser變數
   LET lst_token = base.StringTokenizer.create(ps_key_fields, ',')
   LET ls_bsvars = ""
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      IF ls_token.getIndexOf('(',1) > 0 THEN
         LET ls_token = ls_token.subString(1, ls_token.getIndexOf('(',1)-1)
      END IF      
      LET ls_bsvars = ls_bsvars,"g_browser[g_cnt].b_",ls_token
      IF lst_token.hasMoreTokens() THEN
         LET ls_bsvars = ls_bsvars,","
      END IF
   END WHILE               
   CALL g_properties.addAttribute("browser_vars_all", ls_bsvars )
   
   #ps_key_fields加工 
   LET ls_tmp = ""
   LET lst_token = base.StringTokenizer.create(g_properties.getValue("master_fields_define"), ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      IF ls_token.getIndexOf('DATETIME YEAR TO SECOND',1) THEN
         LET ls_tmp = ls_tmp, ls_token
      END IF
   END WHILE
   
   LET lst_token = base.StringTokenizer.create(ps_key_fields, ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      IF ls_tmp.getIndexOf(ls_token,1) THEN
         LET ls_key_fields = ls_key_fields, ls_token, "(datetime)"
      ELSE
         LET ls_key_fields = ls_key_fields, ls_token
      END IF
      IF lst_token.hasMoreTokens() THEN
         LET ls_key_fields = ls_key_fields, ','
      END IF
   END WHILE
   
   #組合browser欄位定義
   IF NOT adzp150_browser_fields_define(ls_key_fields,"",ls_key_fields,"","","") THEN
      #DISPLAY "ERROR(5):設定g_browse變數發生問題"
   END IF
            
END FUNCTION 


#+ 確定該欄位是否於該table中
PUBLIC FUNCTION adzp150_chk_intbl(ps_table,ps_field)
   DEFINE ps_table      STRING
   DEFINE ps_field      STRING
   DEFINE ls_table      STRING
   DEFINE ls_field      STRING
   DEFINE ls_table_pre  STRING
   DEFINE ls_length     INTEGER
   
   LET ls_table = ps_table
   
   LET ls_field = ps_field

   #變數而非欄位
   IF ls_field.getIndexOf(".",1) > 0 THEN
      LET ls_field = ls_field.subString(ls_field.getIndexOf(".",1)+1,ls_field.getLength())
   END IF
   
   #若有含括號
   IF ls_field.getIndexOf("(",1) > 0 THEN
      LET ls_field = ls_field.subString(1,ls_field.getIndexOf("(",1)-1)
   END IF
   
   IF cl_getField(ls_table,ls_field) THEN
      RETURN TRUE
   ELSE
      RETURN FALSE
   END IF

END FUNCTION

#+ 查找欄位所在的page(名稱)
PUBLIC FUNCTION adzp150_find_page(ps_field,ps_pages)
   DEFINE ps_field        STRING
   DEFINE ps_pages        STRING
   DEFINE li_idx          INTEGER
   DEFINE ls_name         STRING
   DEFINE ls_tmp          STRING
   DEFINE ls_return       STRING
   DEFINE lb_chk          BOOLEAN
   
   LET lb_chk = TRUE
   FOR li_idx = 1 TO g_properties.getValue("page")
      IF NOT cl_null(ps_pages) THEN
         LET ls_tmp = li_idx USING "<<<"
         IF ps_pages.getIndexOf(ls_tmp,1) = 0 THEN
            CONTINUE FOR
         END IF
      END IF
      LET ls_name = adzp150_create_name(li_idx, "detail_fields_define", "<<<") 
      LET ls_tmp = g_properties.getValue(ls_name)
      
      #確定尋找的欄位在單身的哪個page
      IF ls_tmp.getIndexOf(ps_field,1) > 0 THEN
         LET ls_tmp = li_idx USING "<<<"
         LET lb_chk = FALSE
         EXIT FOR
      END IF
   END FOR
   LET ls_tmp = adzp150_create_name(ls_tmp, "detail_var_title", "<<<") 
   LET ls_return = g_properties.getValue(ls_tmp),"[l_ac]"
   
   IF lb_chk THEN
      LET ls_return = ""
   END IF
  
   RETURN ls_return

END FUNCTION 

#+ 查找欄位所在的page(數字)
PUBLIC FUNCTION adzp150_find_page_num(ps_field,ps_pages)
   DEFINE ps_field        STRING
   DEFINE ps_pages        STRING
   DEFINE li_idx          INTEGER
   DEFINE ls_name         STRING
   DEFINE ls_tmp          STRING
   DEFINE ls_return       STRING
   
   FOR li_idx = 1 TO g_properties.getValue("page")
      IF NOT cl_null(ps_pages) THEN
         LET ls_tmp = li_idx USING "<<<"
         IF ps_pages.getIndexOf(ls_tmp,1) = 0 THEN
            CONTINUE FOR
         END IF
      END IF 
      LET ls_name = adzp150_create_name(li_idx, "detail_fields_define", "<<<") 
      LET ls_tmp = g_properties.getValue(ls_name)

      #確定尋找的欄位在單身的哪個page
      IF ls_tmp.getIndexOf(ps_field,1) > 0 THEN
         LET ls_tmp = li_idx USING "<<<"
         EXIT FOR
      END IF
   END FOR
   LET ls_return = ls_tmp

   RETURN ls_return

END FUNCTION 

#+ 產生browser filter相關段落
PUBLIC FUNCTION adzp150_create_borwser_filter(ps_list)
   DEFINE ps_list                  STRING
   DEFINE lst_token                base.StringTokenizer
   DEFINE ls_token                 STRING
   DEFINE ls_filter_field          STRING
   DEFINE ls_field_filter_bs       STRING
   DEFINE ls_display_info          STRING
   DEFINE ls_display_condition     STRING
   DEFINE ls_field                 STRING
   DEFINE ls_fq_field              STRING   #201109-00009#2
   DEFINE ls_fq_value              STRING   #201109-00009#2
   DEFINE ls_fq_value_all          STRING   #201109-00009#2
   DEFINE ls_fq_sql                STRING   #201109-00009#2
   DEFINE ls_fq_cnt                SMALLINT #201109-00009#2
   DEFINE ls_fq_field_1            LIKE dzeb_t.dzeb002   #201109-00009#2
   DEFINE ls_fq_ad_name            STRING   #201109-00009#2

   IF NOT adzp150_type_decide("b") THEN
      RETURN
   END IF
   
   IF g_properties.getValue("type_id_t") = "i13" THEN
      RETURN
   END IF
   
   #201109-00009#2 ---start---
   LET ls_fq_field = NULL
   LET ls_fq_value = NULL
   LET ls_fq_value_all = NULL
   LET ls_fq_sql = "SELECT COUNT(1) FROM dzeb_t",
                   " WHERE dzeb002 = ?",
                     " AND dzeb007 IN ('varchar','varchar2','clob','blob')"   #模糊查詢只針對文字型態欄位
   PREPARE adzp150_borwser_filter_fq_pre FROM ls_fq_sql
   #201109-00009#2 --- end ---

   LET lst_token = base.StringTokenizer.create(ps_list, ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_field = ls_token.subString(1,ls_token.getIndexOf('(',1)-1)
      #非主要table欄位
      IF ls_token.getIndexOf('(',1) > 0 THEN
         LET ls_field = ls_token.subString(1,ls_token.getIndexOf('(',1)-1)
         IF NOT cl_getField(g_properties.getValue("master_tbl_name"),ls_field) THEN
            CONTINUE WHILE
         ELSE
            LET ls_token = ls_field
         END IF
      END IF
      
      LET ls_filter_field = ls_filter_field, ls_token, ','
      
      LET ls_field_filter_bs = ls_field_filter_bs, 's_browse[1].b_', ls_token, ','
      
      LET ls_display_info = ls_display_info, 
                            12 SPACES, "DISPLAY ${general_prefix}_filter_parser('",ls_token,"')",
                            " TO s_browse[1].b_", ls_token, "\n"
                            
      LET ls_display_condition = ls_display_condition, 
                                 3 SPACES, "CALL ${general_prefix}_filter_show('",ls_token,"')", "\n"

      #201109-00009#2 ---start---
      #確認文字欄位才需要處理
      LET ls_fq_cnt = 0
      LET ls_fq_field_1 = ls_token
      EXECUTE adzp150_borwser_filter_fq_pre INTO ls_fq_cnt USING ls_fq_field_1
      IF ls_fq_cnt = NULL OR ls_fq_cnt = 0 THEN
         CONTINUE WHILE
      END IF

      #產生filter段落AFTER FIELD的部分(for 模糊查詢功能)
      LET ls_fq_field = "b_",ls_token
      LET ls_fq_ad_name = "construct.filter.af.",ls_fq_field
      LET ls_fq_value = 9 SPACES,"AFTER FIELD ",ls_fq_field,"\n",
                        9 SPACES,"   #add-point:AFTER FIELD ",ls_fq_field," name=\"",ls_fq_ad_name,"\"","\n",
                        9 SPACES,"   {<point name=\"",ls_fq_ad_name,"\" />}","\n",
                        9 SPACES,"   #END add-point","\n",
                        9 SPACES,"   CALL FGL_DIALOG_GETBUFFER() RETURNING lp_str","\n",
                        9 SPACES,"   IF NOT cl_null(lp_str) THEN","\n",
                        9 SPACES,"      LET lp_str = cl_ap_code_fuzzyquery(lp_str)   #模糊查詢功能","\n",
                        9 SPACES,"      CALL FGL_DIALOG_SETBUFFER(lp_str)","\n",
                        9 SPACES,"   END IF","\n"

      LET ls_fq_value_all = ls_fq_value_all,"\n",ls_fq_value
      #201109-00009#2 --- end ---
   END WHILE

   LET ls_filter_field = ls_filter_field.subString(1,ls_filter_field.getLength()-1)
   LET ls_field_filter_bs = ls_field_filter_bs.subString(1,ls_field_filter_bs.getLength()-1)
   LET ls_display_info = ls_display_info.subString(1,ls_display_info.getLength()-1)
   LET ls_display_condition = ls_display_condition.subString(1,ls_display_condition.getLength()-1)

   CALL g_properties.addAttribute("browser_fields_filter",ls_filter_field)
   CALL g_properties.addAttribute("browser_fields_filter_bs",ls_field_filter_bs)
   CALL g_properties.addAttribute("browser_display_info",ls_display_info)
   CALL g_properties.addAttribute("browser_display_condition",ls_display_condition)
   CALL g_properties.addAttribute("browser_fq_filter",ls_fq_value_all)   #201109-00009#2

END FUNCTION 


#+ 產生slice相關段落(特定樣板使用)
PRIVATE FUNCTION adzp150_create_slice()
   DEFINE ls_tmp      STRING
   DEFINE li_tmp      INTEGER
   
   #產出main段落
   LET ls_tmp = g_properties.getValue("type_id_t")
   IF ls_tmp.getIndexOf("c",1) > 0 THEN
      #子作業
      IF g_properties.getValue("type_id_t") = "c02a" OR 
         g_properties.getValue("type_id_t") = "c02b" OR 
         g_properties.getValue("type_id_t") = "c02c" THEN
         CALL g_properties.addAttribute("mdl_mark", "#該樣板不需此段落")
      END IF
      IF g_properties.getValue("type_id_t") = "c01a" OR 
         g_properties.getValue("type_id_t") = "c03a" THEN
         CALL g_properties.addAttribute("mdl_mark2", "#該樣板不需此段落")
      END IF

      #200804-00006 ---start---
      IF g_properties.getValue("type_id") = "t02" THEN
         CALL g_properties.addAttribute("mdl_mark", "#該樣板不需此段落")
         CALL g_properties.addAttribute("mdl_mark2", "#該樣板不需此段落")
      END IF
      #200804-00006 --- end ---

      LET ls_tmp = adzp150_make_slice("a27")
      CALL g_properties.addAttribute("general_other_dialog", '{<point name="other.dialog"/>}')
   ELSE
      #一般作業
      IF g_properties.getValue("type_id") = "i02"  THEN
         CALL g_properties.addAttribute("mdl_mark", "#該樣板不需此段落")
      ELSE
         CALL g_properties.addAttribute("mdl_mark2", "#該樣板不需此段落")
      END IF
      
      IF g_properties.getValue("type_id") = "t02" THEN
         CALL g_properties.addAttribute("mdl_mark", "#該樣板不需此段落")
         CALL g_properties.addAttribute("mdl_mark2", "#該樣板不需此段落")
      END IF
 
      #200804-00006 ---start---
      IF g_properties.getValue("type_id") = "q01" OR
         g_properties.getValue("type_id") = "q02" OR
         g_properties.getValue("type_id") = "q03" OR
         g_properties.getValue("type_id") = "q04" OR
         g_properties.getValue("type_id") = "p00" OR
         g_properties.getValue("type_id") = "p01" OR
         g_properties.getValue("type_id") = "p02" OR
         g_properties.getValue("type_id") = "r01" THEN
         CALL g_properties.addAttribute("mdl_mark", "#該樣板不需此段落")
         CALL g_properties.addAttribute("mdl_mark2", "#該樣板不需此段落")
      END IF
      #200804-00006 --- end ---
   
      LET ls_tmp = adzp150_make_slice("a26")
   END IF
   CALL g_properties.addAttribute("general_main_function", ls_tmp)
   
   #產出browser_fill段落(i01)
   LET ls_tmp = g_properties.getValue("type_id_t")
   CASE ls_tmp
      WHEN "i01"
         #一般browser
         LET ls_tmp = adzp150_make_slice("a29")
      WHEN "i10"
         #一般browser
         LET ls_tmp = adzp150_make_slice("a29")
      WHEN "i13"
         #額外產出expand功能
         LET ls_tmp = adzp150_make_slice("a53")  
         CALL g_properties.addAttribute("general_tree_expand",ls_tmp)        
         LET ls_tmp = adzp150_make_slice("a54")  
         CALL g_properties.addAttribute("general_tree_expand_define",ls_tmp)        
         LET ls_tmp = adzp150_make_slice("a55")  
         CALL g_properties.addAttribute("general_tree_default",ls_tmp)        
         #樹狀browser(六層架構) 
         LET li_tmp = g_properties.getValue("browser_max_lv")-1
         CALL g_properties.addAttribute("mdl_reproduces",li_tmp)
         LET ls_tmp = adzp150_make_slice("a30")     
         CALL g_properties.addAttribute("mdl_reproduces","0")         
      WHEN "c01a"
         #一般browser
         LET ls_tmp = adzp150_make_slice("a29")
      OTHERWISE
         LET ls_tmp = "#該樣板不需此段落"
   END CASE
   CALL g_properties.addAttribute("general_browser_fill", ls_tmp)

END FUNCTION

#+ page轉換為page_id
PUBLIC FUNCTION adzp150_page_trans(ps_page)
   DEFINE ps_page       STRING
   DEFINE ls_value      STRING
   DEFINE ls_name       STRING
      
   LET ls_name = adzp150_create_name(ps_page, "detail_page_id", "<<<")
   LET ls_value = g_properties.getValue(ls_name)

   RETURN ls_value
   
END FUNCTION

#c類qbe處理(c01c,c02c,c03c)
PUBLIC FUNCTION adzp150_cs_qbe(ps_page,ps_fields)
   DEFINE ps_page         STRING
   DEFINE ps_fields       STRING
   DEFINE ls_name         STRING
   DEFINE ls_value        STRING
   DEFINE ls_record       STRING
   DEFINE lst_token       base.StringTokenizer
   DEFINE ls_token        STRING
   DEFINE ls_all_fields   STRING
   DEFINE ls_all_srfields STRING
   
   #取得record名稱
   LET ls_name = adzp150_create_name(ps_page, "sr_name", "<<<") 
   LET ls_record = g_properties.getValue(ls_name),'[1].'
   
   LET ls_all_fields   = g_properties.getValue("detail_all_fields_qbe")
   LET ls_all_srfields = g_properties.getValue("detail_all_srfields_qbe")
   IF NOT cl_null(ls_all_fields) THEN
      LET ls_all_fields   = ls_all_fields, ','
      LET ls_all_srfields = ls_all_srfields ,','
   END IF 
   
   LET lst_token = base.StringTokenizer.create(ps_fields, ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      
      #去除括號
      IF ls_token.getIndexOf('(',1) THEN
         LET ls_token = ls_token.subString(1,ls_token.getIndexOf('(',1)-1)
      END IF

      #detail_all_fields_qbe
      LET ls_all_fields = ls_all_fields, ls_token, ','
   
      #detail_all_srfields_qbe
      LET ls_all_srfields = ls_all_srfields, ls_record,ls_token, ','
   
   END WHILE
   
   LET ls_all_fields = ls_all_fields.subString(1,ls_all_fields.getLength()-1)
   LET ls_all_srfields = ls_all_srfields.subString(1,ls_all_srfields.getLength()-1)
   
   LET ls_name = "detail_all_fields_qbe"
   LET ls_value = ls_all_fields
   CALL g_properties.addAttribute(ls_name,ls_value)   
   
   LET ls_name = "detail_all_srfields_qbe"
   LET ls_value = ls_all_srfields
   CALL g_properties.addAttribute(ls_name,ls_value)   
   
END FUNCTION

#開始時間
PUBLIC FUNCTION adzp150_start()
   DEFINE li_h    INTEGER
   DEFINE li_m    INTEGER
   DEFINE li_s    INTEGER
   DEFINE ls_time STRING
   
#   LET g_time = CURRENT
#   LET ls_time = g_time
#   LET li_h = ls_time.subString(12,13)
#   LET li_m = ls_time.subString(15,16)
#   LET li_s = ls_time.subString(18,19) 
#   LET g_statetime = li_h*3600 + li_m*60 + li_s
#
END FUNCTION

#結束時間
PUBLIC FUNCTION adzp150_end(ps_prog)
   DEFINE ps_prog CHAR(20)
   DEFINE li_h    INTEGER
   DEFINE li_m    INTEGER
   DEFINE li_s    INTEGER
   DEFINE ls_time STRING
   DEFINE li_time INTEGER
   DEFINE li_chk  INTEGER
   
#   SELECT COUNT(1) INTO li_chk FROM itme_t 
#    WHERE itme001 = ps_prog AND itme002 = pi_ver
#    
#   IF li_chk = 0 THEN
#   INSERT INTO itme_t(itmeent,itme001,itme002,itme003)
#      VALUES('99',ps_prog,pi_ver,0)
#   END IF
#   
#   LET g_alltime = 0
#   SELECT itme003 INTO g_alltime FROM itme_t 
#    WHERE itme001 = ps_prog AND itme002 = pi_ver
#   
#   LET g_time = CURRENT
#   LET ls_time = g_time
#   LET li_h = ls_time.subString(12,13)
#   LET li_m = ls_time.subString(15,16)
#   LET li_s = ls_time.subString(18,19) 
#   LET g_endtime = li_h*3600 + li_m*60 + li_s
#   LET g_alltime = g_alltime + g_endtime - g_statetime
#   
#   UPDATE itme_t SET itme003 = g_alltime
#    WHERE itme001 = ps_prog AND itme002 = pi_ver
#   
END FUNCTION

#針對r01,p01 dialog內容進行處理
PUBLIC FUNCTION adzp150_make_dialog()
   DEFINE ls_name       STRING
   DEFINE ls_value      STRING
   DEFINE lst_token     base.StringTokenizer
   DEFINE ls_token      STRING
   
   IF NOT (g_properties.getValue("type_id") = "r01" OR
           g_properties.getValue("type_id") = "p01" ) THEN
      RETURN   
   END IF
   
   CALL g_properties.addAttribute("location","head" )
   
   #單頭input
   LET ls_name = "master_input_fields"
   LET lst_token = base.StringTokenizer.create(g_properties.getValue("master_input_fields_list"),",")
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_value = ls_value,"g_master.",ls_token,","
   END WHILE
   
   LET ls_value = ls_value.subString(1,ls_value.getLength()-1)
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   CALL g_properties.addAttribute("master_fields_check",  adzp150_create_chk_field("0",3,"i") )
   CALL g_properties.addAttribute("master_fields_ctrlp_i",adzp150_create_chk_field("0",3,"ic") )
   CALL g_properties.addAttribute("master_fields_ctrlp",  adzp150_create_chk_field("0",3,"c") )
   
   IF NOT cl_null(g_properties.getValue("master_input_fields_list")) THEN
      LET ls_name = "general_input_frame"
      LET ls_value = adzp150_make_slice("a57")
      CALL g_properties.addAttribute(ls_name,ls_value)
   END IF
   
   IF NOT cl_null(g_properties.getValue("master_fields_qbe")) THEN
      LET ls_name = "general_construct_frame"
      LET ls_value = adzp150_make_slice("a58")
      CALL g_properties.addAttribute(ls_name,ls_value)
   END IF
   
   #單頭input(show)
   LET ls_value = '\n'
   LET ls_name = "master_get_buffer"
   LET lst_token = base.StringTokenizer.create(g_properties.getValue("master_input_fields_list"),",")
   WHILE lst_token.hasMoreTokens()
      LET ls_token = lst_token.nextToken()
      LET ls_value = ls_value, 
                     "   LET g_master.",ls_token," = p_dialog.getFieldBuffer('",ls_token,"')\n"
   END WHILE
   CALL g_properties.addAttribute(ls_name,ls_value)

END FUNCTION

#外部參數排序
FUNCTION adzp150_arg_idx(pi_fix)
   DEFINE pi_fix        INTEGER
   DEFINE li_idx        INTEGER
   DEFINE li_idx2       INTEGER
   DEFINE li_cnt        INTEGER
   DEFINE ls_name       STRING
   DEFINE ls_value      STRING
   
   #排除固定參數外, 重新確定外部參數序號
   FOR li_idx = 1 TO 100
      LET li_idx2 = li_idx + pi_fix
      LET ls_name  = "general_para", li_idx USING "&&"
      LET ls_value = li_idx2 USING "&&"
      CALL g_properties.addAttribute(ls_name,ls_value)
   END FOR

END FUNCTION

#寫入樣板相關資訊
PUBLIC FUNCTION adzp150_update_gzzx(ps_line)
   DEFINE ps_line  STRING      #資訊字串(未解析)
   DEFINE lr_gzzx  RECORD LIKE gzzx_t.* 
   DEFINE ls_ver   STRING 
   DEFINE li_b     INTEGER 
   DEFINE li_e     INTEGER 
   DEFINE li_cnt   INTEGER 

   #取得樣板代碼
   LET lr_gzzx.gzzx011 = g_properties.getValue("type_id") 
    
   #排除特定樣板(i00,p00,q00)
   IF lr_gzzx.gzzx011 = 'i00' OR
      lr_gzzx.gzzx011 = 'p00' OR
      lr_gzzx.gzzx011 = 'q00' THEN
      RETURN
   END IF
    
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
      DISPLAY 'ERROR(0): 該程式未註冊，請先於azzi900進行註冊！'
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

#產生相關的遮罩段落
FUNCTION adzp150_create_all_mask()
   DEFINE ls_field STRING
   DEFINE ls_var   STRING
   DEFINE li_idx   INTEGER
   DEFINE ls_tmp   STRING
   DEFINE ls_title STRING
   
   LET ls_title = 'general_mask_funcs'
   LET ls_tmp   = '&include "erp/',g_properties.getValue("module")
                              ,'/',g_properties.getValue("app_id"),'_mask.4gl"'
   CALL g_properties.addAttribute(ls_title,ls_tmp)

END FUNCTION

FUNCTION adzp150_msgcentre_notify()
   DEFINE ls_name    STRING
   DEFINE ls_value   STRING

   #目前僅針對i01,i07,t01處理
   LET ls_name  = "mdl_var_title"
   LET ls_value = g_properties.getValue("master_var_title")
   CALL g_properties.addAttribute(ls_name,ls_value)
   
   #190114-00003#2 - S -
   #單身資料也需要
   IF NOT cl_null(g_properties.getValue("detail_var_title")) THEN
      #mdl_detail_param
      LET ls_name = "mdl_detail_param" 
      LET ls_value = "LET g_msgparam.data_d[1] = util.JSON.stringify(",g_properties.getValue("detail_var_title"),")"
      CALL g_properties.addAttribute(ls_name,ls_value)
   END IF
   #190114-00003#2 - E -

   #寫入
   LET ls_name  = "general_msgcentre_notify"
   LET ls_value = adzp150_make_slice("a66")
   CALL g_properties.addAttribute(ls_name,ls_value)

END FUNCTION

#190108-00038 ---start---
#處理簡易畫面相關段落
FUNCTION adzp150_simple_mode()
   DEFINE ls_gzza001     LIKE gzza_t.gzza001   #程式代號
   DEFINE ls_dzzc004     LIKE dzzc_t.dzzc004   #簡易畫面實際存放的模組
   DEFINE ls_cnt         LIKE type_t.num5
   DEFINE ls_name        STRING
   DEFINE ls_value       STRING
   DEFINE ls_code        STRING
   DEFINE ls_module      STRING


   #確定是否有設定簡易畫面，先預設是未設定簡易畫面的模式
   LET g_simple_type = 'N'
   CALL g_properties.addAttribute("general_simple_mode_module","") #simple mode module 

   LET ls_gzza001 = g_properties.getValue("app_id")   #程式代號
   LET ls_cnt = 0
   SELECT COUNT(1) INTO ls_cnt FROM dzzc_t
    WHERE dzzcstus = 'Y'
      AND dzzc001 = ls_gzza001
  
   #確定有設定簡易畫面的程式才需要產出以下相關的段落
   IF ls_cnt > 0 THEN
      #取出簡易畫面實際存放的模組
      #註：因為簡易畫面是透過設定而來，所以有可能標準化面已客製，但簡易畫面仍為標準
      #    或是標準畫面仍為標準，但簡易畫面改為客製的狀況
      #    所以簡易畫面需以dzzc_t中紀錄的實際模組(dzzc004)為主
      LET ls_dzzc004 = NULL
      SELECT dzzc004 INTO ls_dzzc004 FROM dzzc_t
       WHERE dzzcstus = 'Y'
         AND dzzc001 = ls_gzza001

      IF NOT cl_null(ls_dzzc004) THEN
         LET g_simple_type = 'Y'
         LET ls_dzzc004 = DOWNSHIFT(ls_dzzc004)
         CALL g_properties.addAttribute("general_simple_mode_module",ls_dzzc004)   #simple mode module 

         #產生簡易畫面切換的ON ACTION段落
         LET ls_name  = "general_on_action_ui_change"
         LET ls_value = adzp150_create_action_ui_change('dialog','a72')
         CALL g_properties.addAttribute(ls_name,ls_value)
   
         IF g_properties.getValue("type_id") = "i01" OR
            g_properties.getValue("type_id") = "i05" THEN
            LET ls_name  = "general_on_action_ui_change_menu"
            LET ls_value = adzp150_create_action_ui_change('menu','a72')
            CALL g_properties.addAttribute(ls_name,ls_value)
         END IF
   
         #產生確認畫面模式的FUNCTION段
         LET ls_name = "general_ui_chk"
         LET ls_value = adzp150_make_slice('a73')
         CALL g_properties.addAttribute(ls_name,ls_value)
   
         #產生init()段更換圖片及說明的程式段
         LET ls_name = "general_define_ui_change_icon"
         LET ls_value = adzp150_make_slice('a74')
         CALL g_properties.addAttribute(ls_name,ls_value)
         
         #產生轉換畫面時相關程式段
         LET ls_name = "general_ui_change"
         LET ls_value = adzp150_make_slice('a75')
         CALL g_properties.addAttribute(ls_name,ls_value)
      END IF
   END IF

   #若是沒有設定簡易畫面，需隱藏畫面切換的action
   IF g_simple_type = 'N' THEN
      LET ls_name = "general_define_ui_change_icon"
      LET ls_value = "CALL gfrm_curr.setElementHidden('ui_change',1)   #隱藏畫面切換action"
      CALL g_properties.addAttribute(ls_name,ls_value)
   END IF


   LET ls_code = g_properties.getValue("general_prefix")
   LET ls_module = g_properties.getValue("general_module")

   #產生main段開啟畫面的相關段落
   LET ls_name = 'general_main_open_window'
   IF g_simple_type = 'Y' THEN
      #因為1.0與3.0是共用同一份程式碼，
      #因此有可能在3.0環境下有設了3.0版才有的功能，
      #但是因為1.0與3.0是讀同一份設定檔 (因為資料庫是同一份，沒有區分版本)
      #所以切換到1.0環境後，若單純只讀設定檔的話，會判斷是有設定的，
      #為避免在1.0環境下執行到3.0才有的功能，
      #必須在產出相關的程式段時，一併產出判斷版本的部分
      LET ls_value =          "&ifdef v1   #區分版本\n",
                     6 SPACES,"   OPEN WINDOW w_",ls_code," WITH FORM cl_ap_formpath('",ls_module,"',g_code)\n",
                     6 SPACES,"&else\n",
                     6 SPACES,"   IF g_simple_mode THEN   #簡化模式\n",
                     6 SPACES,"      OPEN WINDOW w_",ls_code," WITH FORM cl_ap_formpath('",ls_dzzc004,"',g_simple_name)\n",
                     6 SPACES,"   ELSE\n",
                     6 SPACES,"      OPEN WINDOW w_",ls_code," WITH FORM cl_ap_formpath('",ls_module,"',g_code)\n",
                     6 SPACES,"   END IF\n",
                     6 SPACES,"&endif"
   ELSE
      LET ls_value = "OPEN WINDOW w_",ls_code," WITH FORM cl_ap_formpath('",ls_module,"',g_code)"
   END IF
   CALL g_properties.addAttribute(ls_name,ls_value)

   #產生main段確認畫面模式的段落
   IF g_simple_type = 'Y' THEN
      LET ls_name = "general_check_ui_mode"
      LET ls_value =          "&ifndef v1   #區分版本\n",
                     3 SPACES,"   #確認畫面模式\n",
                     3 SPACES,"   CALL ",ls_code,"_ui_chk()\n",
                     3 SPACES,"&endif"
      CALL g_properties.addAttribute(ls_name,ls_value)
   END IF

END FUNCTION
#190108-00038 --- end ---
